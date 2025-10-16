package types

import (
	"context"
	"home_automation_server/engine/rules"
)

type ServiceHandler func(ctx context.Context, action *rules.Action) error

type ServiceData struct {
	FullName           string
	Handler            ServiceHandler
	RequiredParams     map[string]ParamMetadata
	RequiresTargetType bool
	RequiresTargetID   bool
}

type ParamMetadata struct {
	DataType    string
	Description string
}
