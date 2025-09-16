package api

import (
	"context"

	corev1 "k8s.io/api/core/v1"
)

/*
	Node, Pod 기본 정보
*/

// NodeInfo: 노드 상태 요약
type NodeInfo struct {
	Name                string
	Labels              map[string]string
	Annotations         map[string]string
	Taints              []corev1.Taint
	AllocatableCPUMilli int64
	AllocatableMemBytes int64
	Ready               bool
	Unschedulable       bool
	Score               int
}

// PodInfo: 파드 스펙 요약
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

/*
	Affinity / Toleration 표현
*/

type Operator string

const (
	OpIn           Operator = "In"
	OpNotIn        Operator = "NotIn"
	OpExists       Operator = "Exists"
	OpDoesNotExist Operator = "DoesNotExist"
	OpGt           Operator = "Gt"
	OpLt           Operator = "Lt"
)

// Requirement: 라벨 매칭 조건
type Requirement struct {
	Key      string // 라벨 키
	Operator Operator
	Values   []string // In, NotIn일 때만 사용
}

// AffinityTerm: 공통 조건 (AND 연결)
type AffinityTerm struct {
	Requirements []Requirement
}

// NodeAffinityTerm는 단순히 Requirement 집합
type NodeAffinityTerm = AffinityTerm

// PodAffinityTerm: 파드 매칭 조건 + topologyKey
type PodAffinityTerm struct {
	AffinityTerm
	TopologyKey string
}

/*
	NodeAffinity
*/

// WeightedNodeAffinityTerm: NodeAffinityTerm + Weight
type WeightedNodeAffinityTerm struct {
	AffinityTerm
	Weight int // 1~100, Kubernetes 스펙 반영
}

type NodeAffinity struct {
	Required  []NodeAffinityTerm
	Preferred []WeightedNodeAffinityTerm
}

/*
	PodAffinity / PodAntiAffinity
*/

// WeightedPodAffinityTerm: PodAffinityTerm + Weight
type WeightedPodAffinityTerm struct {
	PodAffinityTerm
	Weight int
}

type PodAffinity struct {
	Required  []PodAffinityTerm
	Preferred []WeightedPodAffinityTerm
}

type PodAntiAffinity struct {
	Required  []PodAffinityTerm
	Preferred []WeightedPodAffinityTerm
}

/*
	Toleration
*/

type Toleration struct {
	Key      string
	Operator string
	Value    string
	Effect   string
}

/*
	Scheduler Finder 함수 시그니처
*/

// FinderFunc: Pod와 Node 리스트를 받아 스케줄링 대상 노드명 결정
type FinderFunc func(context.Context, PodInfo, []NodeInfo, []PodInfo) (string, error)
