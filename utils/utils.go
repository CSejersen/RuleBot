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
	} else {
		// if attributes are nil we initialize them to avoid nil pointer dereferences.
		deepCopy.Attributes = make(map[string]any, len(s.Attributes))
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

// StatesEqual compares two State objects and returns true if they have the same EntityID, State, and Attributes.
// It ignores LastChanged, LastUpdated, and Context fields as these are metadata.
func StatesEqual(a, b types.State) bool {
	if a.EntityID != b.EntityID {
		return false
	}

	// Compare main state using AnyEqual to handle numeric type conversions
	if !AnyEqual(a.State, b.State) {
		return false
	}

	// Compare attributes
	aAttrs := a.Attributes
	bAttrs := b.Attributes

	// Both nil or both empty maps
	if len(aAttrs) == 0 && len(bAttrs) == 0 {
		return true
	}

	// One is nil/empty and the other isn't
	if len(aAttrs) != len(bAttrs) {
		return false
	}

	// Compare each attribute value
	for k, aVal := range aAttrs {
		bVal, ok := bAttrs[k]
		if !ok || !AnyEqual(aVal, bVal) {
			return false
		}
	}

	return true
}
