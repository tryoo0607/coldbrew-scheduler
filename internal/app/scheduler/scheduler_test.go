// internal/app/scheduler/scheduler_test.go
package scheduler_test

import (
	"context"
	"testing"
	"time"

	"github.com/tryoo0607/coldbrew-scheduler/internal/app/scheduler"
	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo"
	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
	clientk8s "github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/k8s"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func TestSchedulerBindsPod(t *testing.T) {
	// 1) 테스트용 파사드 + fake clientset 준비
	cli, cs := newTestFacadeAndCS()

	// 2) 컨텍스트 & 파인더
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	find := func(ctx context.Context, _ api.PodInfo, _ []api.NodeInfo) (string, error) {
		return "node-a", nil
	}

	// 3) 스케줄러 실행
	t.Log("→ start scheduler")
	errCh := startScheduler(ctx, cli, find)

	// 4) 테스트 리소스 준비
	t.Log("→ ensure node")
	ensureNode(t, ctx, cs, "node-a")

	t.Log("→ create schedulable pod")
	createSchedulablePod(t, ctx, cs, "default", "p1")

	// 5) 바인딩 검증
	t.Log("→ wait for binding")
	waitForBinding(t, ctx, cs, "default", "p1", "node-a")

	// 6) 종료
	t.Log("→ cancel & wait")
	cancel()
	<-errCh
	t.Log("✓ done")
}

/* ----------------------- Helpers ----------------------- */

// 테스트용 파사드 + fake clientset을 함께 반환
func newTestFacadeAndCS() (clientgo.Client, kubernetes.Interface) {
	cs := clientk8s.NewFakeClientset()
	cli := clientgo.NewWithClientset(cs)
	return cli, cs
}

// 스케줄러 실행 헬퍼
func startScheduler(
	ctx context.Context,
	cli clientgo.Client,
	find api.FinderFunc,
) <-chan error {
	ch := make(chan error, 1)
	go func() {
		ch <- scheduler.Run(ctx, cli, find)
	}()
	return ch
}

// 리소스 생성/검증 유틸은 그대로 사용
func ensureNode(t *testing.T, ctx context.Context, cs kubernetes.Interface, name string) {
	t.Helper()
	_, err := cs.CoreV1().Nodes().Create(ctx, &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}, metav1.CreateOptions{})
	if err != nil && !isAlreadyExists(err) {
		t.Fatalf("create node %q: %v", name, err)
	}
}

func createSchedulablePod(t *testing.T, ctx context.Context, cs kubernetes.Interface, ns, name string) {
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

func waitForBinding(t *testing.T, ctx context.Context, cs kubernetes.Interface, ns, name, expectNode string) {
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
