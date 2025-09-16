package finder

import (
	"context"
	"fmt"

	"github.com/tryoo0607/coldbrew-scheduler/internal/app/scheduler"
	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
)

func FindBestNode(ctx context.Context, targetPodInfo api.PodInfo, listNodeInfos []api.NodeInfo, allPodInfos []api.PodInfo) (string, error) {

	// 1. nodeName 우선 적용
	if targetPodInfo.NodeName != "" {
		nodeName, err := applyNodeName(targetPodInfo, listNodeInfos)

		if err != nil {
			return "", err
		}

		return nodeName, nil
	}

	// 2. 필터 단계 (Ready, Unschedulable, NodeSelector 등)
	filteredNodes, err := scheduler.FilterNodes(targetPodInfo, listNodeInfos, allPodInfos)

	if err != nil {
		return "", fmt.Errorf("failed to filter nodes for pod %q: %w", targetPodInfo.Name, err)
	}

	// 3. 스코어링 단계 (노드 점수 계산)
	scoredNodes, err := scheduler.ScoringNodes(filteredNodes)

	if err != nil {

		return "", fmt.Errorf("failed to score nodes for pod %q: %w", targetPodInfo.Name, err)
	}

	// 5. 최고 점수 노드 선택 단계
	targetNode, err := findBestScoreNode(scoredNodes)

	if err != nil {
		return "", fmt.Errorf("failed to select best node for pod %q: %w", targetPodInfo.Name, err)
	}

	return targetNode, nil
}

func applyNodeName(podInfo api.PodInfo, listNodeInfos []api.NodeInfo) (string, error) {
	// NodeName이 지정되지 않은 경우 → 스킵
	if podInfo.NodeName == "" {
		return "", nil
	}

	// 노드 존재 여부 확인
	node, err := findNodeByName(podInfo.NodeName, listNodeInfos)
	if err != nil {
		return "", fmt.Errorf("specified node %q not found", podInfo.NodeName)
	}

	// Ready 상태 검사
	if !node.Ready {
		return "", fmt.Errorf("specified node %q is not Ready", podInfo.NodeName)
	}

	// 스케줄링 불가 여부 검사
	if node.Unschedulable {
		return "", fmt.Errorf("specified node %q is marked Unschedulable", podInfo.NodeName)
	}

	// 리소스 충분한지 검사
	if node.AllocatableCPUMilli < podInfo.CPUmilliRequest ||
		node.AllocatableMemBytes < podInfo.MemoryBytes {
		return "", fmt.Errorf("specified node %q doesn't have enough resources", podInfo.NodeName)
	}

	// 문제 없으면 해당 노드로 스케줄링
	return node.Name, nil
}

func findNodeByName(nodeName string, listNodeInfos []api.NodeInfo) (*api.NodeInfo, error) {

	for i := range listNodeInfos {

		if listNodeInfos[i].Name == nodeName {

			return &listNodeInfos[i], nil
		}

	}

	return nil, fmt.Errorf("node %s not found", nodeName)
}

func findBestScoreNode(listNodeInfos []*api.NodeInfo) (string, error) {

	return "", nil
}
