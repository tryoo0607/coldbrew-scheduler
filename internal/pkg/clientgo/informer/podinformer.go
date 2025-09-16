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
	"k8s.io/client-go/tools/cache"
)

type PodController struct {
	ctx          context.Context
	find         api.FinderFunc
	podInformer  informers.PodInformer
	nodeInformer informers.NodeInformer
	clientset    kubernetes.Interface
}

func NewPodController(
	ctx context.Context,
	clientset kubernetes.Interface,
	podInformer informers.PodInformer,
	nodeInformer informers.NodeInformer,
	find api.FinderFunc,
) *PodController {
	c := &PodController{
		ctx:          ctx,
		clientset:    clientset,
		find:         find,
		podInformer:  podInformer,
		nodeInformer: nodeInformer,
	}

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.onPodAdded,
	})

	return c
}

func (c *PodController) Run(stopCh <-chan struct{}) {
	go c.podInformer.Informer().Run(stopCh)
	go c.nodeInformer.Informer().Run(stopCh)

	// 캐시 동기화 대기
	if !cache.WaitForCacheSync(stopCh,
		c.podInformer.Informer().HasSynced,
		c.nodeInformer.Informer().HasSynced,
	) {
		// 실패 시 로그 출력이나 에러 핸들링 가능
		fmt.Println("✗ Informer cache sync failed")
		return
	}

	fmt.Println("✓ Informer cache synced")

	// block (테스트 환경이라면 <-stopCh만 해도 됨)
	<-stopCh
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
	// 1. 전체 PodInfos 수집
	allPodInfos, err := listPodInfos(c)
	if err != nil {
		fmt.Printf("listPodInfos error: %v\n", err)
		return
	}

	// 2. NodeInfos 변환 (allPods 포함)
	candidates, err := listNodeInfos(c, allPodInfos)
	if err != nil {
		fmt.Printf("listNodeInfos error: %v\n", err)
		return
	}

	// 3. 현재 스케줄링 대상 Pod 변환
	pi, err := adapter.ToPodInfo(pod)
	if err != nil {
		fmt.Printf("Convert Pod to PodInfo error for %s/%s: %v\n", pod.Namespace, pod.Name, err)
		return
	}

	// 4. Finder로 최적 노드 선택
	nodeName, err := c.find(c.ctx, pi, candidates, allPodInfos)
	if err != nil {
		fmt.Printf("findBestNode error for %s/%s: %v\n", pod.Namespace, pod.Name, err)
		return
	}
	if nodeName == "" {
		return
	}

	// 5. 바인딩
	if err := binder.BindPodToNode(binder.BindOptions{
		ClientSet: c.clientset,
		Ctx:       c.ctx,
		Pod:       pod,
		NodeName:  nodeName,
	}); err != nil {
		fmt.Printf("bind error %s/%s → %s: %v\n", pod.Namespace, pod.Name, nodeName, err)
	}
}
