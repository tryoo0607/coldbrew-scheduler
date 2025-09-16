package scheduler

import (
	"strconv"

	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
)

/* --- 공통 유틸 --- */

// Map에 key=value 존재 여부 확인
func hasKeyValue(m map[string]string, key, value string) bool {
	v, ok := m[key]
	return ok && v == value
}

/* --- NodeAffinity 처리 --- */

// NodeAffinity.Required 검사
func matchRequiredNodeAffinity(labels map[string]string, required []api.NodeAffinityTerm) bool {
	if len(required) == 0 {
		return true // Required 조건 없으면 통과
	}

	for _, term := range required {
		if matchNodeAffinityTerm(labels, term) {
			return true // 하나라도 만족하면 OK
		}
	}
	return false
}

// NodeAffinity.Preferred 점수 계산
func scorePreferredNodeAffinity(labels map[string]string, prefs []api.WeightedNodeAffinityTerm) int {
	score := 0
	for _, pref := range prefs {
		if matchNodeAffinityTerm(labels, pref.AffinityTerm) {
			score += pref.Weight
		}
	}
	return score
}

// NodeAffinityTerm 검사 (AND 조건)
func matchNodeAffinityTerm(labels map[string]string, term api.NodeAffinityTerm) bool {
	for _, req := range term.Requirements {
		if !matchRequirement(labels, req) {
			return false
		}
	}
	return true
}

/* --- PodAffinity 처리 --- */

// PodAffinity.Required 검사
func matchRequiredPodAffinity(pod api.PodInfo, node api.NodeInfo, required []api.PodAffinityTerm, allPodInfos []api.PodInfo) bool {
	for _, term := range required {
		if !matchPodAffinityTerm(pod, node, term, allPodInfos) {
			return false
		}
	}
	return true
}

// PodAffinity.Preferred 점수 계산
func scorePreferredPodAffinity(pod api.PodInfo, node api.NodeInfo, prefs []api.WeightedPodAffinityTerm, allPodInfos []api.PodInfo) int {
	score := 0
	for _, pref := range prefs {
		if matchPodAffinityTerm(pod, node, pref.PodAffinityTerm, allPodInfos) {
			score += pref.Weight
		}
	}
	return score
}

/* --- PodAntiAffinity 처리 --- */

// PodAntiAffinity.Required 검사
func matchRequiredPodAntiAffinity(pod api.PodInfo, node api.NodeInfo, required []api.PodAffinityTerm, allPodInfos []api.PodInfo) bool {
	for _, term := range required {
		if matchPodAffinityTerm(pod, node, term, allPodInfos) {
			// anti-affinity인데 충족 → 배치 불가
			return false
		}
	}
	return true
}

// PodAntiAffinity.Preferred 점수 계산
func scorePreferredPodAntiAffinity(pod api.PodInfo, node api.NodeInfo, prefs []api.WeightedPodAffinityTerm, allPodInfos []api.PodInfo) int {
	score := 0
	for _, pref := range prefs {
		if matchPodAffinityTerm(pod, node, pref.PodAffinityTerm, allPodInfos) {
			// anti-affinity 조건 충족 → 감점
			score -= pref.Weight
		}
	}
	return score
}

/* --- PodAffinity / PodAntiAffinity 공통 헬퍼 --- */

// PodAffinityTerm 검사
func matchPodAffinityTerm(pod api.PodInfo, node api.NodeInfo, term api.PodAffinityTerm, allPodInfos []api.PodInfo) bool {
	for _, existingPod := range allPodInfos {
		if existingPod.NodeName != node.Name {
			continue
		}

		// requirements 불일치 → continue
		if !matchPodRequirements(existingPod.Labels, term.Requirements) {
			continue
		}

		// topologyKey 불일치 → continue
		if !matchTopologyKey(pod, existingPod, term.TopologyKey) {
			continue
		}

		// 둘 다 만족하면 term 충족
		return true
	}
	return false
}

// Pod label requirements 검사
func matchPodRequirements(labels map[string]string, reqs []api.Requirement) bool {
	for _, req := range reqs {
		if !matchRequirement(labels, req) {
			return false
		}
	}
	return true
}

// topologyKey 일치 여부 검사
func matchTopologyKey(pod api.PodInfo, existingPod api.PodInfo, key string) bool {
	if key == "" {
		return true // topologyKey 없으면 무시
	}
	val1, ok1 := pod.Labels[key]
	val2, ok2 := existingPod.Labels[key]
	return ok1 && ok2 && val1 == val2
}

/* --- Resource 계산 유틸 --- */

// 특정 노드에 올라간 Pod들의 리소스 요청 합계
func calcNodeUsedResources(nodeName string, allPods []api.PodInfo) (cpuMilli int64, memBytes int64) {
	nodePods := make([]api.PodInfo, 0)
	for _, p := range allPods {
		if p.NodeName == nodeName {
			nodePods = append(nodePods, p)
		}
	}
	return sumPodRequests(nodePods)
}

// PodInfo 리스트의 리소스 요청 합계
func sumPodRequests(pods []api.PodInfo) (cpuMilli int64, memBytes int64) {
	for _, p := range pods {
		cpuMilli += p.CPUmilliRequest
		memBytes += p.MemoryBytes
	}
	return
}

/* --- Requirement 매칭 처리 (In, NotIn, Exists, Gt, Lt 등) --- */

// Requirement 매칭
func matchRequirement(labels map[string]string, req api.Requirement) bool {
	val, exists := labels[req.Key]

	switch req.Operator {
	case api.OpIn:
		return matchIn(val, req.Values)
	case api.OpNotIn:
		return matchNotIn(val, req.Values)
	case api.OpExists:
		return matchExists(exists)
	case api.OpDoesNotExist:
		return matchDoesNotExist(exists)
	case api.OpGt:
		return matchGt(val, exists, req.Values)
	case api.OpLt:
		return matchLt(val, exists, req.Values)
	}
	return false
}

// In
func matchIn(val string, values []string) bool {
	for _, v := range values {
		if v == val {
			return true
		}
	}
	return false
}

// NotIn
func matchNotIn(val string, values []string) bool {
	for _, v := range values {
		if v == val {
			return false
		}
	}
	return true
}

// Exists
func matchExists(exists bool) bool {
	return exists
}

// DoesNotExist
func matchDoesNotExist(exists bool) bool {
	return !exists
}

// Gt
func matchGt(val string, exists bool, values []string) bool {
	if !exists || len(values) == 0 {
		return false
	}
	iv, err1 := strconv.Atoi(val)
	target, err2 := strconv.Atoi(values[0])
	if err1 != nil || err2 != nil {
		return false
	}
	return iv > target
}

// Lt
func matchLt(val string, exists bool, values []string) bool {
	if !exists || len(values) == 0 {
		return false
	}
	iv, err1 := strconv.Atoi(val)
	target, err2 := strconv.Atoi(values[0])
	if err1 != nil || err2 != nil {
		return false
	}
	return iv < target
}
