package rules

import (
	"home_automation_server/pubsub"
	"regexp"
	"strings"
)

type RuleSet struct {
	Rules []Rule `yaml:"rules"`
}

type Rule struct {
	When Condition `yaml:"when"`
	Then []Action  `yaml:"then"`
}

type Condition struct {
	Source      string         `yaml:"source"`
	Type        string         `yaml:"type"`
	Entity      string         `yaml:"entity"`
	StateChange string         `yaml:"state_change"`
	Conditions  map[string]any `yaml:"conditions,omitempty"`
}

// TODO: Validate Rule config on start up

func (c *Condition) Matches(event pubsub.Event) bool {
	if normalizeString(event.Source) != normalizeString(c.Source) {
		return false
	}
	if normalizeString(event.Type) != normalizeString(c.Type) {
		return false
	}
	if normalizeString(event.Entity) != normalizeString(c.Entity) {
		return false
	}
	if normalizeString(event.StateChange) != normalizeString(c.StateChange) {
		return false
	}

	for key, expected := range c.Conditions {
		if _, ok := event.Payload[key]; !ok {
			return false
		}

		// TODO: Extend conditions to allow for "in: ["movie", "reading"]" or ">=: 10"
		// to achieve that we can default to strict equality for primitive types
		// but if the parsed yaml gives a map we can switch on the operator ("in", "between", "<" ...) to decide if it's a match.
		if expected != event.Payload[key] {
			return false
		}
	}
	return true
}

func normalizeString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")

	return s
}
