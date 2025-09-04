package clientgo

import (
	"context"
	"fmt"

	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/adapter"
	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/informer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// 파사드 패턴
type client struct{ cs kubernetes.Interface }

func newClient(cs kubernetes.Interface) Client { return &client{cs: cs} }

func (c *client) ListNodeInfos(ctx context.Context) ([]api.NodeInfo, error) {
	nl, err := c.cs.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list nodes: %w", err)
	}
	return adapter.ToNodeInfoList(nl)
}

func (c *client) NewPodInformer(ctx context.Context, find api.FinderFunc) (cache.Controller, error) {
   
    ctrl := informer.NewPodInformer(ctx, c.cs, find)
    return ctrl, nil
}