package adapter

import (
	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/* --- Requirement 변환 --- */

// NodeSelectorRequirement → []Requirement
func toRequirements(exprs []corev1.NodeSelectorRequirement) []api.Requirement {
	reqs := make([]api.Requirement, 0, len(exprs))
	for _, expr := range exprs {
		reqs = append(reqs, api.Requirement{
			Key:      expr.Key,
			Operator: api.Operator(expr.Operator),
			Values:   expr.Values,
		})
	}
	return reqs
}

// LabelSelectorRequirement → []Requirement
func toLabelRequirements(exprs []metav1.LabelSelectorRequirement) []api.Requirement {
	reqs := make([]api.Requirement, 0, len(exprs))
	for _, expr := range exprs {
		reqs = append(reqs, api.Requirement{
			Key:      expr.Key,
			Operator: api.Operator(expr.Operator),
			Values:   expr.Values,
		})
	}
	return reqs
}

/* --- PodAffinityTerm 변환 --- */

// PodAffinityTerm → []PodAffinityTerm
func toPodAffinityTerms(terms []corev1.PodAffinityTerm) []api.PodAffinityTerm {
	result := make([]api.PodAffinityTerm, 0, len(terms))
	for _, term := range terms {
		if term.LabelSelector != nil {
			reqs := toLabelRequirements(term.LabelSelector.MatchExpressions)
			if len(reqs) > 0 {
				result = append(result, api.PodAffinityTerm{
					AffinityTerm: api.AffinityTerm{
						Requirements: reqs,
					},
					TopologyKey: term.TopologyKey,
				})
			}
		}
	}
	return result
}

/* --- NodeAffinity Preferred 변환 --- */

// PreferredSchedulingTerm → []WeightedNodeAffinityTerm
func toWeightedNodeAffinity(prefs []corev1.PreferredSchedulingTerm) []api.WeightedNodeAffinityTerm {
	result := make([]api.WeightedNodeAffinityTerm, 0, len(prefs))
	for _, pref := range prefs {
		reqs := toRequirements(pref.Preference.MatchExpressions)
		if len(reqs) > 0 {
			result = append(result, api.WeightedNodeAffinityTerm{
				AffinityTerm: api.AffinityTerm{
					Requirements: reqs,
				},
				Weight: int(pref.Weight),
			})
		}
	}
	return result
}

/* --- Weight 보정 --- */

// Weight 값(1~100 범위 보정)
func checkWeight(weight int) int {
	w := weight
	if w < 1 {
		w = 1
	}
	if w > 100 {
		w = 100
	}
	return w
}

/* --- PodAffinity / PodAntiAffinity Preferred 변환 --- */

// WeightedPodAffinityTerm → []WeightedPodAffinityTerm
func toWeightedPodAffinity(weighted []corev1.WeightedPodAffinityTerm) []api.WeightedPodAffinityTerm {
	result := make([]api.WeightedPodAffinityTerm, 0, len(weighted))
	for _, w := range weighted {
		terms := toPodAffinityTerms([]corev1.PodAffinityTerm{w.PodAffinityTerm})
		for _, t := range terms {
			result = append(result, api.WeightedPodAffinityTerm{
				PodAffinityTerm: t,
				Weight:          checkWeight(int(w.Weight)),
			})
		}
	}
	return result
}

/* --- Node 리소스/상태 유틸 --- */

// Node에서 할당 가능한 리소스 추출
func getAllocatableResources(n *corev1.Node) (cpuMilli int64, memBytes int64) {
	if q, ok := n.Status.Allocatable[corev1.ResourceCPU]; ok {
		cpuMilli = q.MilliValue()
	}
	if q, ok := n.Status.Allocatable[corev1.ResourceMemory]; ok {
		memBytes = q.Value()
	}
	return
}

// NodeReady 상태 체크
func isNodeReady(n *corev1.Node) bool {
	for _, c := range n.Status.Conditions {
		if c.Type == corev1.NodeReady && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

/* --- Pod 리소스 사용량 계산 --- */

// 특정 노드에 올라간 Pod들의 리소스 요청 합계 계산
func calcNodeUsedResources(nodeName string, allPods []api.PodInfo) (cpuMilli int64, memBytes int64) {
	nodePods := make([]api.PodInfo, 0)
	for _, p := range allPods {
		if p.NodeName == nodeName {
			nodePods = append(nodePods, p)
		}
	}
	return sumPodRequests(nodePods)
}

// PodInfo 리스트 전체의 리소스 요청 합계
func sumPodRequests(pods []api.PodInfo) (cpuMilli int64, memBytes int64) {
	for _, p := range pods {
		cpuMilli += p.CPUmilliRequest
		memBytes += p.MemoryBytes
	}
	return
}
