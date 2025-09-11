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
	NodeAffinity    *NodeAffinity
	PodAffinity     *PodAffinity
	PodAntiAffinity *PodAntiAffinity
	Tolerations     []Toleration
	CPUmilliRequest int64
	MemoryBytes     int64
}

type Requirement struct {
	Key      string   // 라벨 키
	Operator string   // In, NotIn, Exists, DoesNotExist, Gt, Lt
	Values   []string // 연산자가 In, NotIn일 때만 사용
}

type AffinityTerm struct {
	Requirements []Requirement // AND 조건
}

type NodeAffinityTerm = AffinityTerm

type PodAffinityTerm struct {
	AffinityTerm
	TopologyKey string
}

type NodeAffinity struct {
	Required  []NodeAffinityTerm
	Preferred []NodeAffinityTerm
}

type PodAffinity struct {
	Required  []PodAffinityTerm
	Preferred []PodAffinityTerm
}

type PodAntiAffinity struct {
	Required  []PodAffinityTerm
	Preferred []PodAffinityTerm
}

type Toleration struct {
	Key      string
	Operator string
	Value    string
	Effect   string
}

type FinderFunc func(context.Context, PodInfo, []NodeInfo) (string, error)
