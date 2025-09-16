package informer

import (
	"context"
	"fmt"

	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/adapter"
	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/binder"
	corev1 "k8s.io/api/core/v1"
	informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

type PodController struct {
	ctx         context.Context
	find        api.FinderFunc
	podInformer informers.PodInformer
	nodeLister  v1.NodeLister
	clientset   kubernetes.Interface
}

func NewPodController(
	ctx context.Context,
	clientset kubernetes.Interface,
	podInformer informers.PodInformer,
	nodeLister v1.NodeLister,
	find api.FinderFunc,
) *PodController {
	c := &PodController{
		ctx:         ctx,
		clientset:   clientset,
		find:        find,
		podInformer: podInformer,
		nodeLister:  nodeLister,
	}

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.onPodAdded,
	})

	return c
}

func (c *PodController) Run(stopCh <-chan struct{}) {

	c.podInformer.Informer().Run(stopCh)
}

func (c *PodController) onPodAdded(obj interface{}) {

	pod, ok := obj.(*corev1.Pod)

	if !ok {
		return
	}

	fmt.Printf("NEW POD: %s/%s\n", pod.Namespace, pod.Name)

	c.schedulePod(pod)
}

func (c *PodController) schedulePod(pod *corev1.Pod) {

	candidates, err := listNodeInfos(c)

	if err != nil {
		fmt.Println(err)
		return
	}

	pi, err := adapter.ToPodInfo(pod)

	if err != nil {
		fmt.Printf("Convert Pod to PodInfo error for %s/%s: %v\n", pod.Namespace, pod.Name, err)
		return
	}

	allPodInfos, err := listPodInfos(c)

	if err != nil {
		fmt.Printf("listPodInfos error: %v\n", err)
		return
	}

	nodeName, err := c.find(c.ctx, pi, candidates, allPodInfos)
	if err != nil {
		fmt.Printf("findBestNode error for %s/%s: %v\n", pod.Namespace, pod.Name, err)
		return
	}
	if nodeName == "" {
		return
	}

	if err := binder.BindPodToNode(binder.BindOptions{
		ClientSet: c.clientset,
		Ctx:       c.ctx,
		Pod:       pod,
		NodeName:  nodeName,
	}); err != nil {
		fmt.Printf("bind error %s/%s → %s: %v\n", pod.Namespace, pod.Name, nodeName, err)
	}
}
