package informer

import (
	"context"

	informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"
)

type NodeController struct {
	ctx          context.Context
	nodeInformer informers.NodeInformer
}

func NewNodeController(ctx context.Context, nodeInformer informers.NodeInformer) *NodeController {
	c := &NodeController{
		ctx:          ctx,
		nodeInformer: nodeInformer,
	}

	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{})

	return c
}

func (c *NodeController) Run(stopCh <-chan struct{}) {
	c.nodeInformer.Informer().Run(stopCh)
}
