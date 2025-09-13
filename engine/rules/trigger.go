package rules

import (
	"home_automation_server/engine/pubsub"
	"home_automation_server/utils"
	"strings"
)

type Trigger struct {
	Event       string  `yaml:"event"`
	Entity      *string `yaml:"entity_name,omitempty"`
	StateChange string  `yaml:"state_change"`
}

func (t *Trigger) Matches(event pubsub.Event) bool {
	split := strings.Split(t.Event, ".")
	if len(split) != 2 {
		return false
	}
	triggerSource := split[0]
	triggerType := split[1]

	if utils.NormalizeString(event.Source) != utils.NormalizeString(triggerSource) {
		return false
	}
	if utils.NormalizeString(event.Type) != utils.NormalizeString(triggerType) {
		return false
	}
	if t.Entity != nil && utils.NormalizeString(event.Entity) != utils.NormalizeString(*t.Entity) {
		return false
	}
	if utils.NormalizeString(event.StateChange) != utils.NormalizeString(t.StateChange) {
		return false
	}
	return true
}
