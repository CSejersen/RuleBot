package rules

import (
	"fmt"
	"strings"
)

type Action struct {
	Service  string         `yaml:"service"`
	Target   Target         `yaml:"target,omitempty"`
	Params   map[string]any `yaml:"params,omitempty"`
	Blocking bool           `yaml:"blocking,omitempty"`
}

type Target struct {
	Typ string `yaml:"type,omitempty"`
	ID  string `yaml:"id"`
}

type TemplateRef struct {
	Source  string // "payload", "state" ..
	Path    string
	Default any
}

func ParseTemplateParam(val string) (*TemplateRef, bool) {
	val = strings.TrimSpace(val)

	if !strings.HasPrefix(val, "${") || !strings.HasSuffix(val, "}") {
		return nil, false
	}

	inner := strings.TrimSuffix(strings.TrimPrefix(val, "${"), "}")

	// optional defaultVal, split on "|"
	var path string
	var defaultVal any
	parts := strings.Split(inner, "|")
	path = strings.TrimSpace(parts[0])
	if len(parts) == 2 {
		defaultVal = strings.TrimSpace(parts[1])
	}

	switch {
	case strings.HasPrefix(path, "payload."):
		return &TemplateRef{
			Source:  "payload",
			Path:    strings.TrimPrefix(path, "payload."),
			Default: defaultVal,
		}, true
	case strings.HasPrefix(path, "state."):
		return &TemplateRef{
			Source:  "state",
			Path:    strings.TrimPrefix(path, "state."),
			Default: defaultVal,
		}, true
	default:
		return nil, false
	}
}

func (a *Action) FloatParam(key string) (float64, error) {
	raw, ok := a.Params[key]
	if !ok {
		return 0, fmt.Errorf("missing param: %s", key)
	}

	switch v := raw.(type) {
	case float64:
		return v, nil
	case *float64:
		if v == nil {
			return 0, fmt.Errorf("param %s is nil pointer", key)
		}
		return *v, nil
	default:
		return 0, fmt.Errorf("param %s must be float, got %T", key, raw)
	}
}

func (a *Action) IntParam(key string) (int, error) {
	raw, ok := a.Params[key]
	if !ok {
		return 0, fmt.Errorf("missing param: %s", key)
	}
	val, ok := raw.(int)
	if !ok {
		return 0, fmt.Errorf("param %s must be int, got %T", key, raw)
	}
	return val, nil
}

func (a *Action) StringParam(key string) (string, error) {
	raw, ok := a.Params[key]
	if !ok {
		return "", fmt.Errorf("missing param: %s", key)
	}
	val, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("param %s must be string, got %T", key, raw)
	}
	return val, nil
}

func (a *Action) BooleanParam(key string) (bool, error) {
	raw, ok := a.Params[key]
	if !ok {
		return false, fmt.Errorf("missing param: %s", key)
	}

	switch v := raw.(type) {
	case bool:
		return v, nil
	case *bool:
		if v == nil {
			return false, fmt.Errorf("param %s is nil pointer", key)
		}
		return *v, nil
	default:
		return false, fmt.Errorf("param %s must be float, got %T", key, raw)
	}
}
