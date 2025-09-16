package adapter

import (
	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func toRequirementsHelper(exprs []corev1.NodeSelectorRequirement) []api.Requirement {
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

func toPodAffinityTermsHelper(terms []corev1.PodAffinityTerm) []api.PodAffinityTerm {
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

// NodeAffinity Preferred 변환
func toWeightedNodeAffinityTerms(prefs []corev1.PreferredSchedulingTerm) []api.WeightedNodeAffinityTerm {
	result := make([]api.WeightedNodeAffinityTerm, 0, len(prefs))
	for _, pref := range prefs {
		reqs := toRequirementsHelper(pref.Preference.MatchExpressions)
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

// PodAffinity / PodAntiAffinity Preferred 변환
func toWeightedPodAffinityTerms(weighted []corev1.WeightedPodAffinityTerm) []api.WeightedPodAffinityTerm {
	result := make([]api.WeightedPodAffinityTerm, 0, len(weighted))
	for _, w := range weighted {
		terms := toPodAffinityTermsHelper([]corev1.PodAffinityTerm{w.PodAffinityTerm})
		for _, t := range terms {
			result = append(result, api.WeightedPodAffinityTerm{
				PodAffinityTerm: t,
				Weight:          checkWeight(int(w.Weight)),
			})
		}
	}
	return result
}
