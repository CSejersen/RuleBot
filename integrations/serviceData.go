package integrations

import (
	"context"
	"home_automation_server/automation"
	"home_automation_server/types"
)

type TargetType string

const (
	TargetTypeEntity TargetType = "entity"
)

type ServiceHandler func(ctx context.Context, action *automation.Action) error

type ServiceSpec struct {
	Handler        ServiceHandler
	RequiredParams map[string]ParamMetadata
	AllowedTargets TargetSpec
}

type TargetSpec struct {
	Type        []TargetType       // could be TargetTypeEntity, TargetTypeDevice ... only entity supported at the moment.
	EntityTypes []types.EntityType // valid if TargetTypeEntity is in TargetType array
}

type ParamMetadata struct {
	DataType    string
	Description string
}
