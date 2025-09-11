package adapter

import (
	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func toRequirements(exprs []corev1.NodeSelectorRequirement) []api.Requirement {
	reqs := make([]api.Requirement, 0, len(exprs))
	for _, expr := range exprs {
		reqs = append(reqs, api.Requirement{
			Key:      expr.Key,
			Operator: string(expr.Operator),
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
			Operator: string(expr.Operator),
			Values:   expr.Values,
		})
	}
	return reqs
}

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

func extractPreferredTerms(weighted []corev1.WeightedPodAffinityTerm) []corev1.PodAffinityTerm {
	result := make([]corev1.PodAffinityTerm, len(weighted))
	for i, w := range weighted {
		result[i] = w.PodAffinityTerm
	}
	return result
}
