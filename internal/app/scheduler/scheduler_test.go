// internal/app/scheduler/scheduler_test.go
package scheduler_test

import (
	"context"
	"testing"
	"time"

	"github.com/tryoo0607/coldbrew-scheduler/internal/app/scheduler"
	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestSchedulerBindsPod(t *testing.T) {
	cs := fake.NewClientset()

	// ---- Reactors: binder 경로 추적용 --------------------------------------
	cs.Fake.PrependReactor("create", "bindings",
		func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			t.Log("[reactor] create bindings called")
			// 기본 동작 계속 타게 false 반환 (fake는 보통 미구현 -> 실패 가능)
			return false, nil, nil
		})
	cs.Fake.PrependReactor("update", "pods",
		func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			t.Log("[reactor] update pods called (fallback path likely)")
			return false, nil, nil
		})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Log("→ start scheduler")
	done := startScheduler(t, ctx, cs, fixedFinder("node-a"))

	t.Log("→ ensure node")
	ensureNode(t, ctx, cs, "node-a")

	t.Log("→ create schedulable pod")
	createSchedulablePod(t, ctx, cs, "default", "p1")

	t.Log("→ wait for binding")
	waitForBinding(t, ctx, cs, "default", "p1", "node-a")

	// 종료 유도 및 그레이스풀 종료 확인
	t.Log("→ cancel & wait")
	cancel()
	<-done
	t.Log("✓ done")
}

// --- Helpers -----------------------------------------------------------------

func fixedFinder(node string) func(api.PodInfo) (string, error) {
	return func(api.PodInfo) (string, error) { return node, nil }
}

func startScheduler(
	t *testing.T,
	ctx context.Context,
	cs *fake.Clientset,
	find func(api.PodInfo) (string, error),
) <-chan error {
	t.Helper()
	done := make(chan error, 1)
	go func() {
		err := scheduler.Run(ctx, cs, find)
		done <- err
	}()
	return done
}

func ensureNode(t *testing.T, ctx context.Context, cs *fake.Clientset, name string) {
	t.Helper()
	_, err := cs.CoreV1().Nodes().Create(ctx, &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}, metav1.CreateOptions{})
	if err != nil && !isAlreadyExists(err) {
		t.Fatalf("create node %q: %v", name, err)
	}
}

func createSchedulablePod(t *testing.T, ctx context.Context, cs *fake.Clientset, ns, name string) {
	t.Helper()
	_, err := cs.CoreV1().Pods(ns).Create(ctx, &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Spec: corev1.PodSpec{
			SchedulerName: api.ColdBrewScheduler,
			Containers:    []corev1.Container{{Name: "c", Image: "busybox"}},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("create pod %s/%s: %v", ns, name, err)
	}
}

func waitForBinding(t *testing.T, ctx context.Context, cs *fake.Clientset, ns, name, expectNode string) {
	t.Helper()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	tries := 0

	for {
		select {
		case <-ctx.Done():
			// 타임아웃 시 현재 상태를 덤프해서 디버깅에 도움
			p, _ := cs.CoreV1().Pods(ns).Get(context.Background(), name, metav1.GetOptions{})
			t.Fatalf("timeout: pod=%s/%s current=%s, status=%s",
				ns, name, safeNode(p), dumpPhase(p))
		case <-ticker.C:
			tries++
			p, err := cs.CoreV1().Pods(ns).Get(ctx, name, metav1.GetOptions{})
			if err == nil && p.Spec.NodeName == expectNode {
				t.Logf("bound after %d tries: %s/%s → %s", tries, ns, name, expectNode)
				return
			}
			// 1초마다 한 번 현재 상태 로그
			if tries%10 == 0 {
				t.Logf("[poll] nodeName=%q phase=%s", safeNode(p), dumpPhase(p))
			}
		}
	}
}

func safeNode(p *corev1.Pod) string {
	if p == nil {
		return "<nil>"
	}
	return p.Spec.NodeName
}

func dumpPhase(p *corev1.Pod) corev1.PodPhase {
	if p == nil {
		return ""
	}
	return p.Status.Phase
}

func isAlreadyExists(err error) bool {
	// strings.Contains로 바꿔도 OK
	if err == nil {
		return false
	}
	es := err.Error()
	return contains(es, "already exists") || contains(es, "AlreadyExists")
}

// --- 미니 문자열 유틸(의존 줄이기용) -----------------------------------------

func contains(s, sub string) bool {
	return indexOf(s, sub) >= 0
}

func indexOf(s, sub string) int {
outer:
	for i := 0; i+len(sub) <= len(s); i++ {
		for j := 0; j < len(sub); j++ {
			if s[i+j] != sub[j] {
				continue outer
			}
		}
		return i
	}
	return -1
}
