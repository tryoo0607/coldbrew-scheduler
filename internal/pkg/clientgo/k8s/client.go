package clientgo

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	fake "k8s.io/client-go/kubernetes/fake"
)

func NewFakeClient() kubernetes.Interface {

	return fake.NewClientset()
}

func ResolveKubeconfigPath(kubeConfig string) string {
	// 1. CLI 플래그 우선
	if kubeConfig != "" {
        return kubeConfig
    }

    // 2. 환경 변수 KUBECONFIG 사용 가능
    if env := os.Getenv("KUBECONFIG"); env != "" {
        return env
    }

    // 3. 기본 경로 사용 (~/.kube/config)
    home, err := os.UserHomeDir()
    if err != nil {
        // 홈 디렉토리 조회 실패 시, kubeconfig 비워서 client-go 기본 로직에 맡김
        return ""
    }
    return filepath.Join(home, ".kube", "config")
}
