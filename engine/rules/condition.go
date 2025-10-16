package rules

import (
	"home_automation_server/engine/types"
	"strings"
)

type Condition struct {
	Entity string `yaml:"entity" json:"entity"` // Integration.Typ.Entity_name
	Field  string `yaml:"field" json:"field"`   // "brightness"

	// Comparison operators
	Equals      interface{} `yaml:"equals,omitempty" json:"equals,omitempty"`
	NotEquals   interface{} `yaml:"not_equals,omitempty" json:"not_equals,omitempty"`
	GreaterThan *float64    `yaml:"gt,omitempty" json:"gt,omitempty"`
	LessThan    *float64    `yaml:"lt,omitempty" json:"lt,omitempty"`
}

func (c *Condition) Matches(s types.StateGetter) bool {
	splitEntity := strings.Split(c.Entity, ".")
	if len(splitEntity) != 3 {
		return false
	}
	source := splitEntity[0]
	typ := splitEntity[1]
	entity := splitEntity[2]

	// TODO: add support for system-wide pseudo-entities like system.time.now

	state, ok := s.GetState(source, typ, entity)
	if !ok {
		return false
	}

	val, exists := state.Fields[c.Field]
	if !exists {
		return false
	}

	if c.Equals != nil && val != c.Equals {
		return false
	}
	if c.NotEquals != nil && val == c.NotEquals {
		return false
	}

	if c.GreaterThan != nil {
		num, ok := val.(float64)
		if !ok || num <= *c.GreaterThan {
			return false
		}
	}
	if c.LessThan != nil {
		num, ok := val.(float64)
		if !ok || num >= *c.LessThan {
			return false
		}
	}

	// Template evaluation

	return true
}
