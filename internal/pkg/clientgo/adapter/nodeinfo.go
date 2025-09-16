package adapter

import (
	"errors"
	"fmt"

	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
	corev1 "k8s.io/api/core/v1"
)

func ToNodeInfoList(nodeList *corev1.NodeList, allPodInfos []api.PodInfo) ([]api.NodeInfo, error) {
	if nodeList == nil {
		return nil, fmt.Errorf("nodeList is nil")
	}

	out := make([]api.NodeInfo, 0, len(nodeList.Items))
	var errs []error

	for i := range nodeList.Items {
		node := &nodeList.Items[i]

		ni, err := ToNodeInfo(node, allPodInfos)
		if err != nil {
			errs = append(errs, fmt.Errorf("convert node %q: %w", node.Name, err))
			continue
		}
		out = append(out, ni)
	}

	if len(errs) > 0 {
		return out, errors.Join(errs...)
	}
	return out, nil
}

func ToNodeInfo(n *corev1.Node, allPodInfos []api.PodInfo) (api.NodeInfo, error) {
	if n == nil {
		return api.NodeInfo{}, fmt.Errorf("node is nil")
	}

	cpuMilli, memBytes := getAllocatableResources(n)
	ready := isNodeReady(n)
	usedCPU, usedMem := calcNodeUsedResources(n.Name, allPodInfos)

	return api.NodeInfo{
		Name:                n.Name,
		Labels:              n.Labels,
		Annotations:         n.Annotations,
		Taints:              n.Spec.Taints,
		AllocatableCPUMilli: cpuMilli,
		AllocatableMemBytes: memBytes,
		UsedCPUMilli:        usedCPU,
		UsedMemBytes:        usedMem,
		Ready:               ready,
		Unschedulable:       n.Spec.Unschedulable,
		Score:               0,
	}, nil
}
