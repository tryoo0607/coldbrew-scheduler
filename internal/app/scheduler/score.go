package scheduler

import (
	"fmt"
	"math"

	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
)

func ScoringNodes(nodes []*api.NodeInfo) ([]*api.NodeInfo, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes to score")
	}

	for _, node := range nodes {
		score := 0

		// 1. 리소스 여유도 점수
		cpuFraction := float64(node.UsedCPUMilli) / float64(node.AllocatableCPUMilli)
		memFraction := float64(node.UsedMemBytes) / float64(node.AllocatableMemBytes)

		// 여유가 많을수록 점수 ↑
		score += int((1.0 - cpuFraction) * 50)
		score += int((1.0 - memFraction) * 50)

		// 2. BalancedResourceAllocation (균형일수록 점수 ↑)
		diff := math.Abs(cpuFraction - memFraction)
		score += int((1.0 - diff) * 20)

		// 3. 이미 NodeAffinity/PodAffinity 단계에서 더해둔 score 반영
		node.Score += score
	}

	return nodes, nil
}
