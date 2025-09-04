package clientgo

import (
	"context"
	"fmt"

	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
	clientk8s "github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/k8s"
	"k8s.io/client-go/kubernetes"
)

// 외부에 노출되는 파사드 인터페이스 (도메인 타입만 주고받음)

type Client interface {
	ListNodeInfos(ctx context.Context) ([]api.NodeInfo, error)
	NewPodController(ctx context.Context, find api.FinderFunc) (Controller, error)
}
type Options struct {
	UseFake    bool
	InCluster  bool
	Kubeconfig string
}

// 통합 생성자: 옵션에 따라 적절한 clientset을 생성해 파사드 반환
func New(opt Options) (Client, error) {
	var (
		cs  kubernetes.Interface
		err error
	)
	switch {
	case opt.UseFake:
		cs = clientk8s.NewFakeClientset()
	case opt.InCluster:
		cs, err = clientk8s.NewClientsetInCluster()
	default:
		cs, err = clientk8s.NewClientsetFromKubeconfig(opt.Kubeconfig)
	}
	if err != nil {
		return nil, fmt.Errorf("new clientset: %w", err)
	}
	return newClient(cs), nil
}
