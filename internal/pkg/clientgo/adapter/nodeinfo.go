package adapter

import (
	"errors"
	"fmt"

	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
	corev1 "k8s.io/api/core/v1"
)

func ToNodeInfo(n *corev1.Node) (api.NodeInfo, error) {
	if n == nil {
		return api.NodeInfo{}, fmt.Errorf("node is nil")
	}

	var cpuMilli int64
	if q, ok := n.Status.Allocatable[corev1.ResourceCPU]; ok {
		cpuMilli = q.MilliValue()
	}
	var memBytes int64
	if q, ok := n.Status.Allocatable[corev1.ResourceMemory]; ok {
		memBytes = q.Value()
	}

	ready := false
	for _, c := range n.Status.Conditions {
		if c.Type == corev1.NodeReady && c.Status == corev1.ConditionTrue {
			ready = true
			break
		}
	}

	return api.NodeInfo{
		Name:                n.Name,
		Labels:              n.Labels,
		Taints:              n.Spec.Taints,
		AllocatableCPUMilli: cpuMilli,
		AllocatableMemBytes: memBytes,
		Ready:               ready,
	}, nil
}

func ToNodeInfoList(nodeList *corev1.NodeList) ([]api.NodeInfo, error) {
	if nodeList == nil {
		return nil, fmt.Errorf("nodeList is nil")
	}

	out := make([]api.NodeInfo, 0, len(nodeList.Items))
	var errs []error

	for i := range nodeList.Items {
		node := &nodeList.Items[i]

		ni, err := ToNodeInfo(node)
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
