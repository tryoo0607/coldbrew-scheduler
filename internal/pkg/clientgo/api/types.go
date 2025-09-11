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
	Unschedulable       bool
}

type PodInfo struct {
	Name            string
	Namespace       string
	Labels          map[string]string
	Annotations     map[string]string
	NodeName        string
	NodeSelector    map[string]string
	Tolerations     []Toleration
	CPUmilliRequest int64
	MemoryBytes     int64
}

type Toleration struct {
	Key      string
	Operator string
	Value    string
	Effect   string
}

type FinderFunc func(context.Context, PodInfo, []NodeInfo) (string, error)
