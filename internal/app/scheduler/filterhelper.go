package scheduler

import (
	"strconv"

	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
)

func hasKeyValue(m map[string]string, key, value string) bool {
	v, ok := m[key]
	return ok && v == value
}

// NodeAffinity.Required 검사
func matchRequiredNodeAffinity(labels map[string]string, required []api.NodeAffinityTerm) bool {
	if len(required) == 0 {
		return true // Required 조건 없으면 통과
	}

	for _, term := range required {
		if matchNodeAffinityTerm(labels, term) {
			return true // 하나라도 만족하면 OK
		}
	}
	return false
}

// NodeAffinity.Preferred 점수 계산
func scorePreferredNodeAffinity(labels map[string]string, prefs []api.WeightedNodeAffinityTerm) int {
	score := 0
	for _, pref := range prefs {
		if matchNodeAffinityTerm(labels, pref.AffinityTerm) {
			score += pref.Weight
		}
	}
	return score
}

func matchNodeAffinityTerm(labels map[string]string, term api.NodeAffinityTerm) bool {
	for _, req := range term.Requirements {
		if !matchRequirement(labels, req) {
			return false
		}
	}
	return true
}

// In
func matchIn(val string, values []string) bool {
	for _, v := range values {
		if v == val {
			return true
		}
	}
	return false
}

// NotIn
func matchNotIn(val string, values []string) bool {
	for _, v := range values {
		if v == val {
			return false
		}
	}
	return true
}

// Exists
func matchExists(exists bool) bool {
	return exists
}

// DoesNotExist
func matchDoesNotExist(exists bool) bool {
	return !exists
}

// Gt
func matchGt(val string, exists bool, values []string) bool {
	if !exists || len(values) == 0 {
		return false
	}
	iv, err1 := strconv.Atoi(val)
	target, err2 := strconv.Atoi(values[0])
	if err1 != nil || err2 != nil {
		return false
	}
	return iv > target
}

// Lt
func matchLt(val string, exists bool, values []string) bool {
	if !exists || len(values) == 0 {
		return false
	}
	iv, err1 := strconv.Atoi(val)
	target, err2 := strconv.Atoi(values[0])
	if err1 != nil || err2 != nil {
		return false
	}
	return iv < target
}

func matchRequirement(labels map[string]string, req api.Requirement) bool {
	val, exists := labels[req.Key]

	switch req.Operator {
	case api.OpIn:
		return matchIn(val, req.Values)
	case api.OpNotIn:
		return matchNotIn(val, req.Values)
	case api.OpExists:
		return matchExists(exists)
	case api.OpDoesNotExist:
		return matchDoesNotExist(exists)
	case api.OpGt:
		return matchGt(val, exists, req.Values)
	case api.OpLt:
		return matchLt(val, exists, req.Values)
	}
	return false
}
