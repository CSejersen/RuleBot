package rules

import (
	"fmt"
	"home_automation_server/engine/pubsub"
	"strings"
)

type Action struct {
	Service string         `yaml:"service"`
	Target  Target         `yaml:"target"`
	Params  map[string]any `yaml:"params,omitempty"`
}

type Target struct {
	Type string `yaml:"type"`
	ID   string `yaml:"id"`
}

func (a *Action) ResolveTemplatedParams(event pubsub.Event) map[string]interface{} {
	resolved := make(map[string]interface{})
	for k, v := range a.Params {
		switch val := v.(type) {
		case string:
			// template syntax ${payload.field}
			if strings.HasPrefix(val, "${payload.") && strings.HasSuffix(val, "}") {
				field := val[len("${payload.") : len(val)-1]
				if value, ok := event.Payload[field]; ok {
					resolved[k] = value
				} else {
					resolved[k] = nil
				}
			} else {
				resolved[k] = val
			}
		default:
			resolved[k] = val
		}
	}
	return resolved
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
