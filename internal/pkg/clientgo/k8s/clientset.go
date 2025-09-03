package k8s

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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
	return fake.NewClientset()
}
