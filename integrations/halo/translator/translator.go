package translator

import (
	"errors"
	"fmt"
	"home_automation_server/engine/integration"
	"home_automation_server/integrations/halo/client"
	"home_automation_server/integrations/halo/translator/events"
	"home_automation_server/types"
	"home_automation_server/utils"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	wheelBufferDuration       = 300 * time.Millisecond
	ButtonAttributePressState = "press_state"
)

type Translator struct {
	Client      *client.Client
	EventParser EventParser
	logger      *zap.Logger

	stateStore         types.StateStore
	EntityRegistry     types.EntityRegistry
	serviceCallTracker integration.ServiceCallTracker

	wheelBuf wheelBuffer
}

type wheelBuffer struct {
	count     int
	lastFlush time.Time
}

func New(client *client.Client, stateStore types.StateStore, entityRegistry types.EntityRegistry, logger *zap.Logger) *Translator {
	t := &Translator{
		Client:         client,
		EventParser:    newEventParser(logger),
		logger:         logger,
		stateStore:     stateStore,
		EntityRegistry: entityRegistry,
	}
	t.init()
	return t
}

func (t *Translator) init() {
	t.LoadEvents() // load all events registered in events/registry
}

// SetServiceCallTracker sets the service call tracker for causality linking
func (t *Translator) SetServiceCallTracker(tracker integration.ServiceCallTracker) {
	t.serviceCallTracker = tracker
}

func (t *Translator) Translate(raw []byte) ([]types.Event, error) {
	event, err := t.EventParser.ParseEvent(raw)
	if err != nil {
		t.logger.Error("failed to parse event", zap.Error(err))
		return nil, err
	}

	context := &types.Context{
		ID:     uuid.NewString(),
		Origin: "external",
	}

	// If number of switch cases grows beyond what is reasonable we could implement a map of event-type to translator-func
	switch event.GetType() {
	case "wheel":
		translated, err := t.translateWheelEvent(event, context)
		if err != nil {
			t.logger.Error("failed to translate wheel event", zap.Error(err))
			return []types.Event{}, err
		}
		return []types.Event{translated}, nil

	case "button":
		translated, err := t.translateButtonEvent(event, context)
		if err != nil {
			t.logger.Error("failed to translate button event", zap.Error(err))
			return []types.Event{}, err
		}
		return []types.Event{translated}, nil

	case "system":
		translated, err := t.translateSystemEvent(event, context)
		if err != nil {
			t.logger.Error("failed to translate system event", zap.Error(err))
			return []types.Event{}, err
		}
		return []types.Event{translated}, nil

	case "status":
		translated, err := t.translateStatusEvent(event, context)
		if err != nil {
			t.logger.Error("failed to translate status event", zap.Error(err))
			return []types.Event{}, err
		}
		return []types.Event{translated}, nil

	default:
		t.logger.Info("translator not implemented, skipping event", zap.String("type", event.GetType()))
	}
	return []types.Event{}, nil
}

func (t *Translator) translateStatusEvent(e types.ExternalEvent, context *types.Context) (types.Event, error) {
	statusEvent, ok := e.(*events.StatusEvent)
	if !ok {
		return types.Event{}, errors.New("failed to type assert status event")
	}

	// Try to match to service call
	if serviceCallContextID, exists := t.serviceCallTracker.PopOldestForDomain("beoremote_halo"); exists {
		context.ParentID = serviceCallContextID
		context.Origin = "halo"
	}

	return types.Event{
		Type: types.EventTypeDomainSpecific,
		Data: types.DomainSpecificEventData{
			Type: "halo_status",
			Data: map[string]any{
				"state":   statusEvent.State,
				"message": statusEvent.Message,
			},
		},
		Context:   context,
		TimeFired: time.Now(),
	}, nil
}

func (t *Translator) translateSystemEvent(e types.ExternalEvent, context *types.Context) (types.Event, error) {
	sysEvent, ok := e.(*events.SystemEvent)
	if !ok {
		return types.Event{}, errors.New("failed to type assert system event")
	}

	entityID, ok := t.EntityRegistry.Resolve(t.Client.Config.ID)
	if !ok {
		t.logger.Info("skipping translation of event with no registered underlying entity")
		return types.Event{}, nil
	}

	oldState := t.getOldState(entityID)
	newState := utils.DeepCopyState(&oldState)
	newState.Context = context
	newState.State = sysEvent.State

	// Check if the new state is actually different from the old state
	if utils.StatesEqual(oldState, newState) {
		t.logger.Debug("Ignoring SystemEvent with no state changes", zap.String("entity_id", entityID))
		return types.Event{}, nil
	}

	return types.Event{
		Type: types.EventTypeStateChanged,
		Data: types.StateChangedData{
			EntityID: entityID,
			OldState: &oldState,
			NewState: &newState,
		},
		Context:   context,
		TimeFired: time.Now(),
	}, nil
}

func (t *Translator) translateWheelEvent(e types.ExternalEvent, context *types.Context) (types.Event, error) {
	wheelEvent, ok := e.(*events.WheelEvent)
	if !ok {
		return types.Event{}, errors.New("expected a *WheelEvent for types type 'wheel'")
	}

	entityID, ok := t.EntityRegistry.Resolve(wheelEvent.ID)
	if !ok {
		return types.Event{}, fmt.Errorf("failed to resolve wheel event with no registered underlying entity")
	}

	oldState := t.getOldState(entityID)
	oldStateVal, ok := ToInt(oldState.State)
	if !ok {
		return types.Event{}, fmt.Errorf("failed to cast wheel state to int: state is %T (value: %v)", oldState.State, oldState.State)
	}

	newState := utils.DeepCopyState(&oldState)
	newState.State = clampInt(oldStateVal+wheelEvent.Counts, 0, 100)

	// Check if the new state is actually different from the old state (e.g. we want to ignore clockwise rotations when state is capped at 100)
	if utils.StatesEqual(oldState, newState) {
		t.logger.Debug("Ignoring WheelEvent with no state changes", zap.String("entity_id", entityID))
		return types.Event{}, nil
	}

	return types.Event{
		Type: types.EventTypeStateChanged,
		Data: types.StateChangedData{
			EntityID: entityID,
			OldState: &oldState,
			NewState: &newState,
		},
		Context:   context,
		TimeFired: time.Now(),
	}, nil
}

func (t *Translator) translateButtonEvent(e types.ExternalEvent, context *types.Context) (types.Event, error) {
	return types.Event{}, errors.New("no translator for button pressed events")
}

func (t *Translator) LoadEvents() {
	for typ, data := range events.Registry {
		t.EventParser.RegisterEvent(typ, data)
	}
}

func (t *Translator) getOldState(entityID string) types.State {
	oldState, ok := t.stateStore.Get(entityID)
	if !ok {
		t.logger.Info("failed to fetch old state", zap.String("entity_id", entityID))
		return types.State{
			EntityID:   entityID,
			Attributes: make(map[string]any),
		}
	}
	return oldState
}

// ToInt converts various numeric types to int
func ToInt(val any) (int, bool) {
	if val == nil {
		return 0, true
	}

	// Use utils.ToFloat64 which already handles json.Number and all numeric types
	f, ok := utils.ToFloat64(val)
	if !ok {
		return 0, false
	}
	return int(f), true
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
