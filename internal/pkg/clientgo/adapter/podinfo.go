package adapter

import (
	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
	corev1 "k8s.io/api/core/v1"
)

func ToPodInfo(pod *corev1.Pod) api.PodInfo {
	// 리소스 요청 합계 계산 (간단 버전)
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
		NodeSelector:    pod.Spec.NodeSelector,
		CPUmilliRequest: cpuMilli,
		MemoryBytes:     memBytes,
	}

	return podInfo
}
