package informer

import (
	"fmt"

	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/adapter"
	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func listNodeInfos(c *PodController) ([]api.NodeInfo, error) {
	// 캐시에서 node 목록 가져오기
	nodes, err := c.nodeLister.List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("list nodes error : %v", err)
	}

	// []*corev1.Node → []corev1.Node 로 변환
	var nodeList corev1.NodeList
	for _, n := range nodes {
		nodeList.Items = append(nodeList.Items, *n)
	}

	// []corev1.Node → []api.NodeInfo 로 변환
	candidates, err := adapter.ToNodeInfoList(&nodeList)
	if err != nil {
		return nil, fmt.Errorf("convert to NodeInfoList error: %v", err)
	}

	return candidates, nil
}
