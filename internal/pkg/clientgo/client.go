package clientgo

import (
	"context"
	"fmt"

	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/adapter"
	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/informer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// 파사드 패턴
type client struct{ cs kubernetes.Interface }

func newClient(cs kubernetes.Interface) Client { return &client{cs: cs} }

func (c *client) ListPodInfos(ctx context.Context) ([]api.PodInfo, error) {
	pl, err := c.cs.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}
	return adapter.ToPodInfoList(pl)
}

func (c *client) ListNodeInfos(ctx context.Context) ([]api.NodeInfo, error) {

	nl, err := c.cs.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list nodes: %w", err)
	}

	allPodInfos, err := c.ListPodInfos(ctx)
	if err != nil {
		return nil, fmt.Errorf("list podInfos: %w", err)
	}

	return adapter.ToNodeInfoList(nl, allPodInfos)
}

func (c *client) NewPodController(ctx context.Context, find api.FinderFunc) (Controller, error) {

	factory := informer.NewInformerFactory(c.cs)

	podInformer := factory.Core().V1().Pods()
	nodeInformer := factory.Core().V1().Nodes()

	ctrl := informer.NewPodController(
		ctx,
		c.cs,
		podInformer,
		nodeInformer.Lister(),
		find,
	)

	return &controllerWrapper{ctrl}, nil
}
