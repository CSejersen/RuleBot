package rules

import (
	"home_automation_server/engine/statestore"
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

// TODO: Validate Rule config on start up

func (r *Rule) ConditionsMatch(s *statestore.StateStore) bool {
	for _, condition := range r.Condition {
		if !condition.Matches(s) {
			return false
		}
	}
	return true
}
