package rules

import (
	"home_automation_server/engine/pubsub"
	"home_automation_server/engine/statestore"
	"regexp"
	"strings"
)

type RuleSet struct {
	Rules []Rule `yaml:"rules"`
}

type Rule struct {
	Alias     string      `yaml:"alias"`
	Trigger   Trigger     `yaml:"trigger"`
	Condition []Condition `yaml:"condition"`
	Action    []Action    `yaml:"action"`
}

type Trigger struct {
	Event       string `yaml:"event"`
	Entity      string `yaml:"entity_name"`
	StateChange string `yaml:"state_change"`
}

// TODO: Validate Rule config on start up

func (r *Rule) ConditionsMatch(s *statestore.StateStore) bool {
	for _, condition := range r.Condition {
		if !condition.Matches(s) {
			return false
		}
	}
	return true
}

func (t *Trigger) Matches(event pubsub.Event) bool {
	split := strings.Split(t.Event, ".")
	if len(split) != 2 {
		return false
	}
	triggerSource := split[0]
	triggerType := split[1]

	if normalizeString(event.Source) != normalizeString(triggerSource) {
		return false
	}
	if normalizeString(event.Type) != normalizeString(triggerType) {
		return false
	}
	if normalizeString(event.Entity) != normalizeString(t.Entity) {
		return false
	}
	if normalizeString(event.StateChange) != normalizeString(t.StateChange) {
		return false
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
