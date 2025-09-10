package executor

import (
	"errors"
	"go.uber.org/zap"
	"home_automation_server/engine/rules"
	"home_automation_server/integrations/halo/client"
)

type Executor struct {
	Client *client.Client
	Logger *zap.Logger
}

func NewExecutor(client *client.Client, logger *zap.Logger) *Executor {
	return &Executor{
		Client: client,
		Logger: logger,
	}
}

func (e *Executor) ExecuteAction(action *rules.Action) error {
	switch action.Action {
	case "update_button_value":
		return e.updateButtonValue(action)
	case "set_scene":
		return e.setScene(action)
	default:
		return errors.New("invalid action")
	}
}
