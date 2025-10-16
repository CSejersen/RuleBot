package translator

import (
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	enginetypes "home_automation_server/engine/types"
	"home_automation_server/integrations/hue/client"
	"home_automation_server/integrations/hue/translator/events"
	"home_automation_server/integrations/types"
	"time"
)

// Translator translates a parsed hue types into a pubsub.Event
type Translator struct {
	Client      *client.Client
	EventParser EventParser
	Logger      *zap.Logger
}

// New is the constructor for Translator
func New(client *client.Client, logger *zap.Logger) (*Translator, error) {
	t := &Translator{
		Client:      client,
		EventParser: newEventParser(logger),
		Logger:      logger,
	}

	if err := t.init(); err != nil {
		return nil, fmt.Errorf("failed to init translator: %w", err)
	}

	return t, nil
}

// TODO: implement continuous refreshing of the registry to keep it aligned with the bridge state.
// maybe events are published on state changes otherwise we refresh on some interval
func (t *Translator) init() error {
	t.LoadEvents()

	return nil
}

func (t *Translator) Translate(raw []byte) ([]enginetypes.Event, error) {
	eventBatch, err := t.EventParser.parse(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse events: %w", err)
	}

	// TODO: if switch cases grow beyond what is reasonable we could implement a map of types type to translator-func
	translatedEvents := []enginetypes.Event{}
	for _, e := range eventBatch.Events {
		switch e.GetType() {
		case "light":
			// a light event might contain multiple changes, a separate event is emitted for each change.
			translated, err := t.translateLightUpdate(e, eventBatch.TimeStamp)
			if err != nil {
				t.Logger.Error("failed to translate light update", zap.Error(err))
				continue
			}
			translatedEvents = append(translatedEvents, translated...)

		case "grouped_light":
			// a grouped_light event might contain multiple changes, a separate event is emitted for each change.
			translated, err := t.translateGroupedLightUpdate(e, eventBatch.TimeStamp)
			if err != nil {
				continue
			}
			translatedEvents = append(translatedEvents, translated...)

		case "scene":
			translated, err := t.translateSceneUpdate(e, eventBatch.TimeStamp)
			if err != nil {
				continue
			}
			translatedEvents = append(translatedEvents, translated)

		default:
			t.Logger.Debug("unknown types", zap.String("type", e.GetType()))
		}

	}
	return translatedEvents, nil
}

func (t *Translator) translateSceneUpdate(e types.SourceEvent, ts time.Time) (enginetypes.Event, error) {
	sceneUpdate, ok := e.(*events.SceneUpdate)
	if !ok {
		t.Logger.Error("failed to type assert on sceneUpdate", zap.Any("types", e))
		return enginetypes.Event{}, fmt.Errorf("expected a *SceneUpdate for types type 'scene'")
	}

	active := sceneUpdate.SafeActive()
	if active == nil {
		t.Logger.Error("no status.active for scene update", zap.Any("scene_update", sceneUpdate))
		return enginetypes.Event{}, fmt.Errorf("no status.active for scene update")
	}

	humanID, ok := t.Client.ResourceRegistry.ResolveName(sceneUpdate.Type, sceneUpdate.ID)
	if !ok {
		t.Logger.Error("failed to lookup name", zap.String("type", sceneUpdate.Type), zap.String("id", sceneUpdate.ID))
	}

	group, ok := t.Client.ResourceRegistry.ResolveGroupForScene(sceneUpdate.ID)
	if !ok {
		t.Logger.Error("failed to lookup group", zap.String("type", sceneUpdate.Type), zap.String("id", sceneUpdate.ID))
	}

	return enginetypes.Event{
		Id:          uuid.NewString(),
		Source:      "hue",
		Type:        "scene",
		Entity:      humanID,
		StateChange: "active_status",
		Payload: map[string]any{
			"active": active,
			"group":  group,
		},
		Time: ts,
	}, nil
}

func (t *Translator) translateLightUpdate(e types.SourceEvent, ts time.Time) ([]enginetypes.Event, error) {
	lightEvent, ok := e.(*events.LightUpdate)
	if !ok {
		return []enginetypes.Event{}, fmt.Errorf("expected a *LightUpdate for types type 'light'")
	}

	humanID, ok := t.Client.ResourceRegistry.ResolveName(lightEvent.Type, lightEvent.ID)
	if !ok {
		t.Logger.Warn("failed to lookup name", zap.String("type", lightEvent.Type), zap.String("id", lightEvent.ID))
		return []enginetypes.Event{}, fmt.Errorf("failed to lookup name")
	}

	psEvents := []enginetypes.Event{}
	// a light event might contain multiple changes, a separate event is emitted for each change.
	for _, change := range lightEvent.ResolveStateChanges() {
		event := enginetypes.Event{
			Id:          uuid.NewString(),
			Source:      "hue",
			Type:        "light",
			Entity:      humanID,
			StateChange: change,
			Payload: map[string]interface{}{
				"brightness":     lightEvent.SafeBrightness(),
				"on":             lightEvent.SafeOn(),
				"mirek":          lightEvent.SafeMirek(),
				"color_xy":       lightEvent.SafeColorXY(),
				"effect":         lightEvent.SafeEffect(),
				"alert":          lightEvent.SafeAlert(),
				"dynamics_speed": lightEvent.SafeDynamicsSpeed(),
				"gradient_mode":  lightEvent.SafeGradientMode(),
			},
			Time: ts,
		}
		psEvents = append(psEvents, event)
	}
	return psEvents, nil
}

func (t *Translator) translateGroupedLightUpdate(e types.SourceEvent, ts time.Time) ([]enginetypes.Event, error) {
	groupedLightEvent, ok := e.(*events.GroupedLightUpdate)
	if !ok {
		return []enginetypes.Event{}, fmt.Errorf("expected a *GroupedLightUpdate for types type 'grouped_light'")
	}

	humanID, ok := t.Client.ResourceRegistry.ResolveName(groupedLightEvent.Type, groupedLightEvent.ID)
	if !ok {
		return []enginetypes.Event{}, fmt.Errorf("failed to lookup name")
	}

	translated := []enginetypes.Event{}
	// a grouped_light types might contain multiple changes, a separate types is emitted for each change.
	for _, change := range groupedLightEvent.ResolveStateChanges() {
		event := enginetypes.Event{
			Id:          uuid.NewString(),
			Source:      "hue",
			Type:        "grouped_light",
			Entity:      humanID,
			StateChange: change,
			Payload: map[string]interface{}{
				"brightness":     groupedLightEvent.SafeBrightness(),
				"on":             groupedLightEvent.SafeOn(),
				"mirek":          groupedLightEvent.SafeMirek(),
				"color_xy":       groupedLightEvent.SafeColorXY(),
				"effect":         groupedLightEvent.SafeEffect(),
				"alert":          groupedLightEvent.SafeAlert(),
				"dynamics_speed": groupedLightEvent.SafeDynamicsSpeed(),
				"gradient_mode":  groupedLightEvent.SafeGradientMode(),
			},
			Time: ts,
		}
		translated = append(translated, event)
	}
	return translated, nil
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

func (t *Translator) StateChangesForType(typ string) []string {
	return t.EventParser.EventRegistry[typ].StateChanges
}
