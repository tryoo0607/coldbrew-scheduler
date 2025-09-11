package adapter

import (
	"fmt"

	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
	corev1 "k8s.io/api/core/v1"
)

func ToPodInfo(pod *corev1.Pod) (api.PodInfo, error) {
	if pod == nil {
		return api.PodInfo{}, fmt.Errorf("pod is nil")
	}

	// 리소스 요청 합계 계산
	var cpuMilli int64
	var memBytes int64
	for _, c := range pod.Spec.Containers {
		if q, ok := c.Resources.Requests[corev1.ResourceCPU]; ok {
			// CPU는 milli로 변환
			cpuMilli += q.MilliValue()
		}
		if q, ok := c.Resources.Requests[corev1.ResourceMemory]; ok {
			memBytes += q.Value()
		}
	}

	podInfo := api.PodInfo{
		Namespace:       pod.Namespace,
		Name:            pod.Name,
		Labels:          pod.Labels,
		Annotations:     pod.Annotations,
		NodeName:        pod.Spec.NodeName,
		NodeSelector:    pod.Spec.NodeSelector,
		NodeAffinity:    toNodeAffinity(pod.Spec.Affinity.NodeAffinity),
		PodAffinity:     toPodAffinity(pod.Spec.Affinity.PodAffinity),
		PodAntiAffinity: toPodAntiAffinity(pod.Spec.Affinity.PodAntiAffinity),
		Tolerations:     toTolerations(pod.Spec.Tolerations),
		CPUmilliRequest: cpuMilli,
		MemoryBytes:     memBytes,
	}

	return podInfo, nil
}

func toNodeAffinity(na *corev1.NodeAffinity) *api.NodeAffinity {
	if na == nil {
		return nil
	}

	result := &api.NodeAffinity{}

	if na.RequiredDuringSchedulingIgnoredDuringExecution != nil {
		for _, term := range na.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
			reqs := toRequirements(term.MatchExpressions)
			if len(reqs) > 0 {
				result.Required = append(result.Required, api.AffinityTerm{
					Requirements: reqs,
				})
			}
		}
	}

	for _, pref := range na.PreferredDuringSchedulingIgnoredDuringExecution {
		reqs := toRequirements(pref.Preference.MatchExpressions)
		if len(reqs) > 0 {
			result.Preferred = append(result.Preferred, api.AffinityTerm{
				Requirements: reqs,
			})
		}
	}

	return result
}

func toPodAffinity(pa *corev1.PodAffinity) *api.PodAffinity {
	if pa == nil {
		return nil
	}
	return &api.PodAffinity{
		Required: toPodAffinityTerms(pa.RequiredDuringSchedulingIgnoredDuringExecution),
		Preferred: toPodAffinityTerms(
			extractPreferredTerms(pa.PreferredDuringSchedulingIgnoredDuringExecution),
		),
	}
}

func toPodAntiAffinity(pa *corev1.PodAntiAffinity) *api.PodAntiAffinity {
	if pa == nil {
		return nil
	}
	return &api.PodAntiAffinity{
		Required: toPodAffinityTerms(pa.RequiredDuringSchedulingIgnoredDuringExecution),
		Preferred: toPodAffinityTerms(
			extractPreferredTerms(pa.PreferredDuringSchedulingIgnoredDuringExecution),
		),
	}
}

func toTolerations(k8sTolerations []corev1.Toleration) []api.Toleration {
	tolerations := make([]api.Toleration, 0, len(k8sTolerations))
	for _, t := range k8sTolerations {
		tolerations = append(tolerations, api.Toleration{
			Key:      t.Key,
			Operator: string(t.Operator),
			Value:    t.Value,
			Effect:   string(t.Effect),
		})
	}
	return tolerations
}
