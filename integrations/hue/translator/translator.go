package translator

import (
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"home_automation_server/integrations/hue/client"
	"home_automation_server/integrations/hue/translator/events"
	"home_automation_server/types"
	"home_automation_server/utils"
	"time"
)

// Translator translates a parsed hue types into a pubsub.Event
type Translator struct {
	Client         *client.ApiClient
	EventParser    EventParser
	logger         *zap.Logger
	stateStore     types.StateStore
	entityRegistry types.EntityRegistry
}

// New is the constructor for Translator
func New(client *client.ApiClient, stateStore types.StateStore, entityRegistry types.EntityRegistry, logger *zap.Logger) (*Translator, error) {
	t := &Translator{
		Client:         client,
		EventParser:    newEventParser(logger),
		logger:         logger,
		stateStore:     stateStore,
		entityRegistry: entityRegistry,
	}

	t.init()
	return t, nil
}

func (t *Translator) init() {
	t.LoadEvents()
}

func (t *Translator) Translate(raw []byte) ([]types.Event, error) {
	eventBatch, err := t.EventParser.parse(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse events: %w", err)
	}

	// TODO: if switch cases grow beyond what is reasonable we could implement a map of types type to translator-func
	translatedEvents := []types.Event{}
	for _, e := range eventBatch.Events {
		var translated *types.Event

		// The event context
		context := &types.Context{
			ID: uuid.NewString(),
		}

		switch e.GetType() {
		case "light":
			translated, err = t.translateLightUpdate(e, eventBatch.TimeStamp, context)
			if err != nil {
				t.logger.Error("failed to translate light update", zap.Error(err))
				continue
			}
		case "grouped_light":
			translated, err = t.translateGroupedLightUpdate(e, eventBatch.TimeStamp, context)
			if err != nil {
				t.logger.Error("failed to translate grouped_light update", zap.Error(err))
				continue
			}
		case "scene":
			translated, err = t.translateSceneUpdate(e, eventBatch.TimeStamp, context)
			if err != nil {
				t.logger.Error("failed to translate scene update", zap.Error(err))
				continue
			}
		default:
			t.logger.Info("unknown event_type", zap.String("type", e.GetType()))
		}

		// An event might be a no-op in that case the specialized translator func will return nil
		if translated != nil {
			translatedEvents = append(translatedEvents, *translated)
		}
	}

	return translatedEvents, nil
}

func (t *Translator) translateSceneUpdate(e types.ExternalEvent, ts time.Time, context *types.Context) (*types.Event, error) {
	sceneUpdate, ok := e.(*events.SceneUpdate)
	if !ok {
		return nil, fmt.Errorf("expected a *SceneUpdate for event type 'scene'")
	}

	entityID, ok := t.entityRegistry.Resolve(sceneUpdate.ID)
	if !ok {
		return nil, fmt.Errorf("failed to resolve entityID for scene with id: %s", sceneUpdate.ID)
	}
	oldState, err := t.getOldState(entityID)
	if err != nil {
		return nil, err
	}
	newState := utils.DeepCopyState(&oldState)
	newState.Context = context

	sceneStatus := sceneUpdate.SafeStatus()
	if sceneStatus == nil {
		t.logger.Debug("Ignoring SceneUpdate with no status field", zap.String("id", sceneUpdate.ID))
		return nil, nil
	}

	newState.State = sceneStatus.Active
	newState.Attributes["last_recall"] = sceneStatus.LastRecall

	return &types.Event{
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

func (t *Translator) translateLightUpdate(e types.ExternalEvent, ts time.Time, context *types.Context) (*types.Event, error) {
	lightUpdate, ok := e.(*events.LightUpdate)
	if !ok {
		return nil, fmt.Errorf("expected a *LightUpdate for types type 'light'")
	}

	on := lightUpdate.SafeOn()
	brightness := lightUpdate.SafeBrightness()
	colorXY := lightUpdate.SafeColorXY()
	mirek := lightUpdate.SafeMirek()

	// If no tracked fields changed, skip emitting an event
	if on == nil && brightness == nil && colorXY == nil && mirek == nil {
		t.logger.Debug("Ignoring LightUpdate with no tracked changes", zap.String("light_name", *lightUpdate.Metadata.Name))
		return nil, nil
	}

	entityID, ok := t.entityRegistry.Resolve(lightUpdate.ID)
	if !ok {
		return nil, fmt.Errorf("failed to resolve entityID for %s", lightUpdate.Metadata.Name)
	}

	oldState, err := t.getOldState(entityID)
	if err != nil {
		return nil, err
	}
	newState := utils.DeepCopyState(&oldState)
	newState.Context = context

	if on != nil {
		newState.State = *on
	}
	if newState.Attributes == nil {
		newState.Attributes = make(map[string]any)
	}
	if brightness != nil {
		newState.Attributes["brightness"] = *brightness
	}
	if colorXY != nil {
		newState.Attributes["color_xy"] = *colorXY
	}
	if mirek != nil {
		newState.Attributes["mirek"] = *mirek
	}

	return &types.Event{
		Type: types.EventTypeStateChanged,
		Data: types.StateChangedData{
			EntityID: entityID,
			OldState: &oldState,
			NewState: &newState,
		},
		Context:   context,
		TimeFired: ts,
	}, nil
}

func (t *Translator) translateGroupedLightUpdate(e types.ExternalEvent, ts time.Time, context *types.Context) (*types.Event, error) {
	groupedLightEvent, ok := e.(*events.GroupedLightUpdate)
	if !ok {
		return nil, fmt.Errorf("expected a *GroupedLightUpdate for types type 'grouped_light'")
	}

	on := groupedLightEvent.SafeOn()
	brightness := groupedLightEvent.SafeBrightness()
	colorXY := groupedLightEvent.SafeColorXY()
	mirek := groupedLightEvent.SafeMirek()

	// If no tracked fields changed, skip emitting an event
	if on == nil && brightness == nil && colorXY == nil && mirek == nil {
		t.logger.Debug("Ignoring GroupedLightUpdate with no tracked changes", zap.String("id", groupedLightEvent.ID))
		return nil, nil
	}

	entityID, ok := t.entityRegistry.Resolve(groupedLightEvent.ID)
	if !ok {
		t.logger.Info("skipping grouped_light update with no underlying entity in the registry")
		return nil, nil
	}

	oldState, err := t.getOldState(entityID)
	if err != nil {
		return nil, err
	}

	newState := utils.DeepCopyState(&oldState)
	newState.Context = context

	if on != nil {
		newState.State = *on
	}
	if newState.Attributes == nil {
		newState.Attributes = make(map[string]any)
	}
	if brightness != nil {
		newState.Attributes["brightness"] = *brightness
	}
	if colorXY != nil {
		newState.Attributes["color_xy"] = *colorXY
	}
	if mirek != nil {
		newState.Attributes["mirek"] = *mirek
	}

	return &types.Event{
		Type: types.EventTypeStateChanged,
		Data: types.StateChangedData{
			EntityID: entityID,
			OldState: &oldState,
			NewState: &newState,
		},
		Context:   context,
		TimeFired: ts,
	}, nil
}

func (t *Translator) getOldState(entityID string) (types.State, error) {
	oldState, ok := t.stateStore.Get(entityID)
	if !ok {
		t.logger.Info("failed to fetch old state", zap.String("entity_id", entityID))
		return types.State{
			EntityID: entityID,
		}, nil
	}
	return oldState, nil
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
	return t.Client.ResourceRegistry.EntityNamesForType(typ)
}
