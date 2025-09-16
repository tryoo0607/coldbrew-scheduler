package scheduler

import (
	"fmt"

	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
)

func FilterNodes(targetPodInfo api.PodInfo, listNodeInfos []api.NodeInfo, allPodInfos []api.PodInfo) ([]*api.NodeInfo, error) {

	// 1. Node Ready 상태 검사
	readyNodes, err := findNodesByReady(listNodeInfos)
	if err != nil {
		return nil, fmt.Errorf("ready check failed: %w", err)
	}

	// 2. Node Unschedulable 검사
	schedulableNodes, err := findNodesByUnschedulable(readyNodes)
	if err != nil {
		return nil, fmt.Errorf("unschedulable check failed: %w", err)
	}

	// 3. NodeSelector에 맞는 Node 필터링
	selectedNodes, err := findNodesByNodeSelector(targetPodInfo.NodeSelector, schedulableNodes)
	if err != nil {
		return nil, fmt.Errorf("nodeSelector check failed: %w", err)
	}

	// 4. Node Affinity 맞는 Node 필터링
	nodeAffinityNodes, err := findNodesByNodeAffinity(targetPodInfo, selectedNodes)
	if err != nil {
		return nil, fmt.Errorf("nodeAffinity check failed: %w", err)
	}

	// 5. Pod Affinity 맞는 Node 필터링
	podAffinityNodes, err := findNodesByPodAffinity(targetPodInfo, nodeAffinityNodes, allPodInfos)
	if err != nil {
		return nil, fmt.Errorf("podAffinity check failed: %w", err)
	}

	// 6. Pod AntiAffinity 맞는 Node 필터링
	podAntiAffinityNodes, err := findNodesByPodAntiAffinity(targetPodInfo, podAffinityNodes, allPodInfos)
	if err != nil {
		return nil, fmt.Errorf("podAntiAffinity check failed: %w", err)
	}

	// 7. 리소스 충분한지 검사
	resourceFilteredNodes, err := findNodesByResources(targetPodInfo, podAntiAffinityNodes, allPodInfos)
	if err != nil {
		return nil, fmt.Errorf("resource check failed: %w", err)
	}

	return resourceFilteredNodes, nil
}

func findNodesByReady(listNodeInfos []api.NodeInfo) ([]*api.NodeInfo, error) {
	filteredListNodeInfos := make([]*api.NodeInfo, 0, len(listNodeInfos))

	for i := range listNodeInfos {

		ready := listNodeInfos[i].Ready

		if ready {
			filteredListNodeInfos = append(filteredListNodeInfos, &listNodeInfos[i])
		}
	}

	if len(filteredListNodeInfos) == 0 {
		return nil, fmt.Errorf("no Ready nodes found (all nodes are NotReady)")
	}

	return filteredListNodeInfos, nil
}

func findNodesByUnschedulable(listNodeInfos []*api.NodeInfo) ([]*api.NodeInfo, error) {

	filteredListNodeInfos := make([]*api.NodeInfo, 0, len(listNodeInfos))

	for i := range listNodeInfos {

		unschedulable := listNodeInfos[i].Unschedulable

		if !unschedulable {
			filteredListNodeInfos = append(filteredListNodeInfos, listNodeInfos[i])
		}
	}

	if len(filteredListNodeInfos) == 0 {
		return nil, fmt.Errorf("no schedulable nodes found (all nodes are Unschedulable)")
	}

	return filteredListNodeInfos, nil
}

func findNodesByNodeSelector(nodeSelector map[string]string, listNodeInfos []*api.NodeInfo) ([]*api.NodeInfo, error) {

	filteredListNodeInfos := make([]*api.NodeInfo, 0, len(listNodeInfos))

	for i := range listNodeInfos {

		labels := listNodeInfos[i].Labels

		isSelected := true
		for key, value := range nodeSelector {

			if !hasKeyValue(labels, key, value) {
				isSelected = false
				break
			}
		}

		if isSelected {
			filteredListNodeInfos = append(filteredListNodeInfos, listNodeInfos[i])
		}
	}

	if len(filteredListNodeInfos) == 0 {
		return nil, fmt.Errorf("no nodes matched the given NodeSelector: %+v", nodeSelector)
	}

	return filteredListNodeInfos, nil
}

func findNodesByNodeAffinity(pod api.PodInfo, nodes []*api.NodeInfo) ([]*api.NodeInfo, error) {
	if pod.NodeAffinity == nil {
		return nodes, nil
	}

	results := make([]*api.NodeInfo, 0, len(nodes))

	for _, node := range nodes {
		labels := node.Labels

		// 1) Required 조건 검사
		if !matchRequiredNodeAffinity(labels, pod.NodeAffinity.Required) {
			continue
		}

		// 2) Preferred 점수 계산
		score := scorePreferredNodeAffinity(labels, pod.NodeAffinity.Preferred)

		// 점수 반영
		node.Score += score

		results = append(results, node)
	}

	return results, nil
}

func findNodesByPodAffinity(pod api.PodInfo, nodes []*api.NodeInfo, allPodInfos []api.PodInfo) ([]*api.NodeInfo, error) {
	// affinity 없는 경우 그대로 리턴
	if pod.PodAffinity == nil {
		return nodes, nil
	}

	results := make([]*api.NodeInfo, 0, len(nodes))

	for _, node := range nodes {
		// 1) Required 조건 검사
		if !matchRequiredPodAffinity(pod, *node, pod.PodAffinity.Required, allPodInfos) {
			continue
		}

		// 2) Preferred 점수 계산
		score := scorePreferredPodAffinity(pod, *node, pod.PodAffinity.Preferred, allPodInfos)

		// 점수 반영
		node.Score += score

		results = append(results, node)
	}

	return results, nil
}

func findNodesByPodAntiAffinity(pod api.PodInfo, nodes []*api.NodeInfo, allPodInfos []api.PodInfo) ([]*api.NodeInfo, error) {
	// anti-affinity 없는 경우 그대로 리턴
	if pod.PodAntiAffinity == nil {
		return nodes, nil
	}

	results := make([]*api.NodeInfo, 0, len(nodes))

	for _, node := range nodes {
		// 1) Required 조건 검사
		if !matchRequiredPodAntiAffinity(pod, *node, pod.PodAntiAffinity.Required, allPodInfos) {
			continue
		}

		// 2) Preferred 점수 계산
		score := scorePreferredPodAntiAffinity(pod, *node, pod.PodAntiAffinity.Preferred, allPodInfos)

		// 점수 반영 (AntiAffinity는 weight가 높을수록 불리하게 적용할 수도 있음 → 정책에 따라 조정)
		node.Score += score

		results = append(results, node)
	}

	return results, nil
}

func findNodesByResources(pod api.PodInfo, nodes []*api.NodeInfo, allPods []api.PodInfo) ([]*api.NodeInfo, error) {
	results := make([]*api.NodeInfo, 0, len(nodes))

	for _, node := range nodes {
		usedCPU, usedMem := calcNodeUsedResources(node.Name, allPods)

		totalCPU := usedCPU + pod.CPUmilliRequest
		totalMem := usedMem + pod.MemoryBytes

		if totalCPU <= node.AllocatableCPUMilli && totalMem <= node.AllocatableMemBytes {
			results = append(results, node)
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no nodes have enough resources for pod %s/%s", pod.Namespace, pod.Name)
	}

	return results, nil
}
