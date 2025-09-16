package adapter

import (
	"fmt"

	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
	corev1 "k8s.io/api/core/v1"
)

/* --- Pod 변환 유틸 --- */

// PodList → []PodInfo
func ToPodInfoList(pl *corev1.PodList) ([]api.PodInfo, error) {
	if pl == nil {
		return nil, fmt.Errorf("podList is nil")
	}

	out := make([]api.PodInfo, 0, len(pl.Items))
	for i := range pl.Items {
		pi, err := ToPodInfo(&pl.Items[i])
		if err != nil {
			return nil, fmt.Errorf("convert pod %s/%s: %w", pl.Items[i].Namespace, pl.Items[i].Name, err)
		}
		out = append(out, pi)
	}
	return out, nil
}

// Pod → PodInfo
func ToPodInfo(pod *corev1.Pod) (api.PodInfo, error) {
	if pod == nil {
		return api.PodInfo{}, fmt.Errorf("pod is nil")
	}

	// 리소스 요청 합계 계산
	var cpuMilli int64
	var memBytes int64
	for _, c := range pod.Spec.Containers {
		if q, ok := c.Resources.Requests[corev1.ResourceCPU]; ok {
			cpuMilli += q.MilliValue()
		}
		if q, ok := c.Resources.Requests[corev1.ResourceMemory]; ok {
			memBytes += q.Value()
		}
	}

	// --- Affinity nil-safe 변환 ---
	var nodeAffinity *api.NodeAffinity
	var podAffinity *api.PodAffinity
	var podAntiAffinity *api.PodAntiAffinity
	if pod.Spec.Affinity != nil {
		if pod.Spec.Affinity.NodeAffinity != nil {
			nodeAffinity = toNodeAffinity(pod.Spec.Affinity.NodeAffinity)
		}
		if pod.Spec.Affinity.PodAffinity != nil {
			podAffinity = toPodAffinity(pod.Spec.Affinity.PodAffinity)
		}
		if pod.Spec.Affinity.PodAntiAffinity != nil {
			podAntiAffinity = toPodAntiAffinity(pod.Spec.Affinity.PodAntiAffinity)
		}
	}

	podInfo := api.PodInfo{
		Namespace:       pod.Namespace,
		Name:            pod.Name,
		Labels:          pod.Labels,
		Annotations:     pod.Annotations,
		NodeName:        pod.Spec.NodeName,
		NodeSelector:    pod.Spec.NodeSelector,
		NodeAffinity:    nodeAffinity,
		PodAffinity:     podAffinity,
		PodAntiAffinity: podAntiAffinity,
		Tolerations:     toTolerations(pod.Spec.Tolerations),
		CPUmilliRequest: cpuMilli,
		MemoryBytes:     memBytes,
	}

	return podInfo, nil
}

/* --- NodeAffinity 변환 --- */

func toNodeAffinity(na *corev1.NodeAffinity) *api.NodeAffinity {
	if na == nil {
		return nil
	}

	result := &api.NodeAffinity{}

	// Required
	if na.RequiredDuringSchedulingIgnoredDuringExecution != nil {
		for _, term := range na.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
			reqs := toRequirements(term.MatchExpressions)
			if len(reqs) > 0 {
				result.Required = append(result.Required, api.NodeAffinityTerm{
					Requirements: reqs,
				})
			}
		}
	}

	// Preferred (weight 반영 → 헬퍼 사용)
	result.Preferred = toWeightedNodeAffinity(na.PreferredDuringSchedulingIgnoredDuringExecution)

	return result
}

/* --- PodAffinity 변환 --- */

func toPodAffinity(pa *corev1.PodAffinity) *api.PodAffinity {
	if pa == nil {
		return nil
	}

	required := toPodAffinityTerms(pa.RequiredDuringSchedulingIgnoredDuringExecution)
	preferred := toWeightedPodAffinity(pa.PreferredDuringSchedulingIgnoredDuringExecution)

	return &api.PodAffinity{
		Required:  required,
		Preferred: preferred,
	}
}

/* --- PodAntiAffinity 변환 --- */

func toPodAntiAffinity(pa *corev1.PodAntiAffinity) *api.PodAntiAffinity {
	if pa == nil {
		return nil
	}

	required := toPodAffinityTerms(pa.RequiredDuringSchedulingIgnoredDuringExecution)
	preferred := toWeightedPodAffinity(pa.PreferredDuringSchedulingIgnoredDuringExecution)

	return &api.PodAntiAffinity{
		Required:  required,
		Preferred: preferred,
	}
}

/* --- Toleration 변환 --- */

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
