package scheduler

import (
	"fmt"

	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
)

// TODO. [TR-YOO] requset / limit 적용하기
func FilterNodes(podInfo api.PodInfo, listNodeInfos []api.NodeInfo) ([]*api.NodeInfo, error) {

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
	selectedNodes, err := findNodesByNodeSelector(podInfo.NodeSelector, schedulableNodes)
	if err != nil {
		return nil, fmt.Errorf("nodeSelector check failed: %w", err)
	}

	// 4. Node Affinity 맞는 Node 필터링
	// TODO. [TR-YOO] 구현하기

	// 5. Pod Affinity 맞는 Node 필터링
	// TODO. [TR-YOO] 구현하기

	// 6. Pod AntiAffinity 맞는 Node 필터링
	// TODO. [TR-YOO] 구현하기

	// 7. 리소스 충분한지 검사
	// TODO. [TR-YOO] 구현하기

	return selectedNodes, nil
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

func hasKeyValue(m map[string]string, key, value string) bool {
	v, ok := m[key]
	return ok && v == value
}
