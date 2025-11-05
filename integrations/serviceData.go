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

// EventEmitter allows services to emit optimistic state_changed events
// immediately after successful service execution
type EventEmitter interface {
	// EmitStateChanged emits a state_changed event with the given entity ID and new state
	// oldState will be retrieved from the state cache if not provided
	EmitOptimisticStateChanged(entityID string, newState *types.State, oldState *types.State) error
}

type ServiceHandler func(ctx context.Context, action *automation.Action, emitter EventEmitter) error

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
