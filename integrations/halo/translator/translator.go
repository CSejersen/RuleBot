package translator

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"home_automation_server/integrations/halo/client"
	"home_automation_server/integrations/halo/translator/events"
	"home_automation_server/types"
	"home_automation_server/utils"
	"time"
)

const (
	wheelBufferDuration       = 300 * time.Millisecond
	ButtonAttributePressState = "press_state"
)

type Translator struct {
	Client      *client.Client
	EventParser EventParser
	Logger      *zap.Logger

	StateStore     types.StateStore
	EntityRegistry types.EntityRegistry

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
		Logger:         logger,
		StateStore:     stateStore,
		EntityRegistry: entityRegistry,
	}
	t.init()
	return t
}

func (t *Translator) init() {
	t.LoadEvents() // load all events registered in events/registry
}

func (t *Translator) Translate(raw []byte) ([]types.Event, error) {
	event, err := t.EventParser.ParseEvent(raw)
	if err != nil {
		t.Logger.Error("failed to parse event", zap.Error(err))
		return nil, err
	}

	// If number of switch cases grows beyond what is reasonable we could implement a map of event-type to translator-func
	switch event.GetType() {
	case "wheel":
		translated, err := t.translateWheelEvent(event)
		if err != nil {
			t.Logger.Error("failed to translate wheel types", zap.Error(err))
			return []types.Event{}, err
		}
		return []types.Event{translated}, nil

	case "button":
		translated, err := t.translateButtonEvent(event)
		if err != nil {
			t.Logger.Error("failed to translate button types", zap.Error(err))
			return []types.Event{}, err
		}
		return []types.Event{translated}, nil

	case "system":
		translated, err := t.translateSystemEvent(event)
		if err != nil {
			t.Logger.Error("failed to translate system types", zap.Error(err))
			return []types.Event{}, err
		}
		return []types.Event{translated}, nil

	default:
		t.Logger.Info("translator not implemented, skipping types", zap.String("type", event.GetType()))
	}
	return []types.Event{}, nil
}

func (t *Translator) translateSystemEvent(e types.ExternalEvent) (types.Event, error) {
	sysEvent, ok := e.(*events.SystemEvent)
	if !ok {
		return types.Event{}, nil
	}

	entityID, ok := t.EntityRegistry.Resolve(t.Client.Config.ID)
	if !ok {
		t.Logger.Info("skipping translation of event with no registered underlying entity")
	}

	context := &types.Context{ID: uuid.NewString()}

	oldState, ok := t.StateStore.Get(entityID)
	if !ok {
		t.Logger.Warn("failed to find old system_state in state store", zap.String("system_id", t.Client.Config.ID))
	}
	newState := oldState
	newState.Context = context
	newState.State = sysEvent.State

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

func (t *Translator) translateWheelEvent(e types.ExternalEvent) (types.Event, error) {
	wheelEvent, ok := e.(*events.WheelEvent)
	if !ok {
		return types.Event{}, errors.New("expected a *WheelEvent for types type 'wheel'")
	}

	entityID, ok := t.EntityRegistry.Resolve(wheelEvent.ID)
	if !ok {
		return types.Event{}, fmt.Errorf("failed to resolve wheel event with no registered underlying entity")
	}

	var newValue float64
	oldButtonState, ok := t.StateStore.Get(entityID)
	if !ok {
		t.Logger.Info("failed to get state for button")
		oldButtonState = types.State{
			EntityID: entityID,
			State:    50.0,
			Context:  nil,
		}
	}

	oldValue, ok := utils.ToFloat64(oldButtonState.State)
	if !ok {
		return types.Event{}, fmt.Errorf("failed to cast button state to float64")
	}
	newValue = clamp(oldValue+float64(wheelEvent.Counts), 0, 100)

	context := &types.Context{
		ID: uuid.NewString(),
	}

	return types.Event{
		Type: types.EventTypeStateChanged,
		Data: types.StateChangedData{
			EntityID: entityID,
			OldState: &oldButtonState,
			NewState: &types.State{
				EntityID:    entityID,
				State:       newValue,
				Attributes:  oldButtonState.Attributes,
				LastChanged: time.Now(),
				LastUpdated: time.Now(),
				Context:     context,
			},
		},
		Context:   context,
		TimeFired: time.Now(),
	}, nil
}

func (t *Translator) translateButtonEvent(e types.ExternalEvent) (types.Event, error) {
	buttonEvent, ok := e.(*events.ButtonEvent)
	if !ok {
		return types.Event{}, errors.New("expected a *ButtonEvent for types type 'button'")
	}

	oldState, ok := t.StateStore.Get(buttonEvent.ID)
	if !ok {
		return types.Event{}, errors.New("failed to find old button state in state store")
	}

	newAttributes := oldState.Attributes
	newAttributes[ButtonAttributePressState] = buttonEvent.State

	return types.Event{
		Type: types.EventTypeStateChanged,
		Data: types.StateChangedData{
			EntityID: buttonEvent.ID,
			OldState: &oldState,
			NewState: &types.State{
				EntityID:    buttonEvent.ID,
				State:       oldState.State, // button.value (0-100)
				Attributes:  newAttributes,
				LastChanged: time.Time{},
				LastUpdated: time.Time{},
				Context:     nil,
			},
		},
		Context:   nil,
		TimeFired: time.Time{},
	}, nil

}

func ResolveWheelStateChange(count int) string {
	if count > 0 {
		return "clockwise"
	}
	return "counter_clockwise"
}

func (t *Translator) LoadEvents() {
	for typ, data := range events.Registry {
		t.EventParser.RegisterEvent(typ, data)
	}
}

func (t *Translator) EventTypes() []string {
	eventTypes := []string{}
	for k, _ := range t.EventParser.EventRegistry {
		eventTypes = append(eventTypes, k)
	}
	return eventTypes
}

func (t *Translator) EntitiesForType(typ string) []string {
	if typ == "system" {
		return nil
	}
	return t.Client.ButtonNames()
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
