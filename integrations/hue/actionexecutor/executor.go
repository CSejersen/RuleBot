package actionexecutor

import (
	"errors"
	"go.uber.org/zap"
	"home_automation_server/engine/rules"
	hue_client "home_automation_server/integrations/hue/apiclient"
)

type Executor struct {
	Client *hue_client.ApiClient
	Logger *zap.Logger
}

func New(client *hue_client.ApiClient, logger *zap.Logger) *Executor {
	return &Executor{
		Client: client,
		Logger: logger,
	}
}

func (e *Executor) ExecuteAction(action *rules.Action) error {
	switch action.Action {
	case "adjust_brightness":
		return e.adjustBrightness(action)
	case "set_scene":
		return e.setScene(action)
	default:
		return errors.New("invalid action")
	}
}
