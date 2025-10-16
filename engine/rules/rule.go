package rules

import (
	"home_automation_server/engine/types"
)

type RuleSet struct {
	Rules []Rule `yaml:"rules"`
}

type Rule struct {
	Alias     string      `yaml:"alias" json:"alias"`
	Trigger   Trigger     `yaml:"trigger" json:"trigger"`
	Condition []Condition `yaml:"condition" json:"condition"`
	Action    []Action    `yaml:"action" json:"action"`
	Active    bool        `yaml:"active" json:"active"`
}

func (r *Rule) ConditionsMatch(s types.StateGetter) bool {
	for _, condition := range r.Condition {
		if !condition.Matches(s) {
			return false
		}
	}
	return true
}
