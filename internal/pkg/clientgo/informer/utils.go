package informer

import (
	"fmt"

	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/adapter"
	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func listNodeInfos(c *PodController, allPodInfos []api.PodInfo) ([]api.NodeInfo, error) {
	// 캐시에서 node 목록 가져오기
	nodes, err := c.nodeInformer.Lister().List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("list nodes error : %v", err)
	}

	// []*corev1.Node → []corev1.Node 로 변환
	var nodeList corev1.NodeList
	for _, n := range nodes {
		nodeList.Items = append(nodeList.Items, *n)
	}

	// []corev1.Node → []api.NodeInfo 로 변환
	candidates, err := adapter.ToNodeInfoList(&nodeList, allPodInfos)
	if err != nil {
		return nil, fmt.Errorf("convert to NodeInfoList error: %v", err)
	}

	return candidates, nil
}

func listPodInfos(c *PodController) ([]api.PodInfo, error) {
	podList, err := c.podInformer.Lister().List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}

	out := make([]api.PodInfo, 0, len(podList))
	for _, pod := range podList {
		pi, err := adapter.ToPodInfo(pod)
		if err != nil {
			fmt.Printf("convert pod %s/%s error: %v\n", pod.Namespace, pod.Name, err)
			continue
		}
		out = append(out, pi)
	}
	return out, nil
}
