package informer

import (
	"context"
	"fmt"

	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/adapter"
	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/binder"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type Controller struct {
	ctx              context.Context
	clientset        kubernetes.Interface
	find             api.FinderFunc
	toPodInfo        func(*corev1.Pod) (api.PodInfo, error)
	bind             func(binder.BindOptions) error
	newListerWatcher func(kubernetes.Interface) cache.ListerWatcher
}

func NewPodInformer(ctx context.Context, clientset kubernetes.Interface, find api.FinderFunc) cache.Controller {
	c := &Controller{
		ctx:              ctx,
		clientset:        clientset,
		find:             find,
		toPodInfo:        adapter.ToPodInfo,
		bind:             binder.BindPodToNode,
		newListerWatcher: newListWatcher,
	}

	options := c.buildInformerOptions()

	_, controller := cache.NewInformerWithOptions(options)

	return controller
}

func (c *Controller) buildInformerOptions() cache.InformerOptions {
	lw := c.newListerWatcher(c.clientset)
	return cache.InformerOptions{
		ListerWatcher: lw,
		ObjectType:    &corev1.Pod{},
		ResyncPeriod:  0,
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: c.onAdd,
		},
	}
}

func (c *Controller) onAdd(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return
	}
	c.schedulePod(pod)
}

func (c *Controller) schedulePod(pod *corev1.Pod) {

	pi, err := c.toPodInfo(pod)
	if err != nil {
		fmt.Printf("Convert Pod to PodInfo  error for %s/%s: %v\n", pod.Namespace, pod.Name, err)
		return
	}

	node, err := c.find(pi)
	if err != nil {
		fmt.Printf("findBestNode error for %s/%s: %v\n", pod.Namespace, pod.Name, err)
		return
	}
	if node == "" {
		return
	}

	if err := c.bind(binder.BindOptions{
		ClientSet: c.clientset,
		Ctx:       c.ctx,
		Pod:       pod,
		NodeName:  node,
	}); err != nil {
		fmt.Printf("bind error %s/%s → %s: %v\n", pod.Namespace, pod.Name, node, err)
	}
}
