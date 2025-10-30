package utils

import (
	"encoding/json"
	"home_automation_server/types"
	"reflect"
	"regexp"
	"strings"
)

func NormalizeString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	re := regexp.MustCompile(`[^\p{L}\p{N}]+`)

	s = re.ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")

	return s
}

func DeepCopyState(s *types.State) types.State {
	if s == nil {
		return types.State{}
	}

	deepCopy := types.State{
		EntityID:    s.EntityID,
		State:       s.State,
		LastChanged: s.LastChanged,
		LastUpdated: s.LastUpdated,
		Context:     s.Context,
	}

	// Deep deepCopy Attributes map
	if s.Attributes != nil {
		deepCopy.Attributes = make(map[string]any, len(s.Attributes))
		for k, v := range s.Attributes {
			deepCopy.Attributes[k] = v
		}
	}

	return deepCopy
}

func AnyEqual(a, b any) bool {
	switch aVal := a.(type) {
	case string:
		if bVal, ok := b.(string); ok {
			return aVal == bVal
		}
	case bool:
		if bVal, ok := b.(bool); ok {
			return aVal == bVal
		}
	case int, int32, int64, float32, float64, json.Number:
		aF, ok1 := ToFloat64(a)
		bF, ok2 := ToFloat64(b)
		if ok1 && ok2 {
			return aF == bF
		}
	}
	return reflect.DeepEqual(a, b)
}

func ToFloat64(val any) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case json.Number:
		f, err := v.Float64()
		if err != nil {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
}
