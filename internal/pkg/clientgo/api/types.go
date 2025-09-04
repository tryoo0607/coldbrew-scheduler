package api

import (
	"context"

	corev1 "k8s.io/api/core/v1"
)

type NodeInfo struct {
	Name                string
	Labels              map[string]string
	Annotations         map[string]string
	Taints              []corev1.Taint
	AllocatableCPUMilli int64
	AllocatableMemBytes int64
	Ready               bool
}

type PodInfo struct {
	Namespace       string
	Name            string
	Labels          map[string]string
	Annotations     map[string]string
	NodeSelector    map[string]string
	CPUmilliRequest int64
	MemoryBytes     int64
}

type FinderFunc func(context.Context, PodInfo, []NodeInfo) (string, error)