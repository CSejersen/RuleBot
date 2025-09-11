package actionexecutor

import (
	"home_automation_server/engine/rules"
)

func (e *Executor) updateButtonValue(action *rules.Action) error {
	value, err := action.FloatParam("value")
	if err != nil {
		return err
	}

	return e.Client.UpdateButtonValue(action.Target.ID, int(value))
}
