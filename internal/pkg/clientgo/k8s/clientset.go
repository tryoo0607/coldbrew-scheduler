package k8s

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

// 인클러스터용 Clientset
func NewClientsetInCluster() (kubernetes.Interface, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("in-cluster config: %w", err)
	}
	return kubernetes.NewForConfig(cfg)
}

// kubeconfig 경로 기반 Clientset (빈 문자열이면 기본 경로/환경변수 적용)
func NewClientsetFromKubeconfig(path string) (kubernetes.Interface, error) {
	if path == "" {
		path = resolveKubeconfigPath("")
	}
	cfg, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, fmt.Errorf("kubeconfig %q: %w", path, err)
	}
	return kubernetes.NewForConfig(cfg)
}

// 테스트/로컬용 Fake Clientset
func NewFakeClientset() kubernetes.Interface {

	cs := fake.NewClientset()

	// TODO. [TR-YOO] fmt.Println()을 logging 라이브러리로 교체하기
	// ---- Reactors: binder 경로 추적용 --------------------------------------
	cs.Fake.PrependReactor("create", "bindings",
		func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			fmt.Println("[reactor] create bindings called")
			// 기본 동작 계속 타게 false 반환 (fake는 보통 미구현 -> 실패 가능)
			return false, nil, nil
		})
	cs.Fake.PrependReactor("update", "pods",
		func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			fmt.Println("[reactor] update pods called (fallback path likely)")
			return false, nil, nil
		})

	return cs
}
