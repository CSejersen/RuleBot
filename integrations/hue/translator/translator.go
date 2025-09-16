package translator

import (
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/engine/pubsub"
	"home_automation_server/integrations/hue/client"
	"home_automation_server/integrations/hue/translator/events"
	"time"
)

// Translator translates a parsed hue event into a pubsub.Event
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

func (t *Translator) Translate(raw []byte) ([]pubsub.Event, error) {
	eventBatch, err := t.EventParser.parse(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse events: %w", err)
	}

	// TODO: if switch cases grow beyond what is reasonable we could implement a map of event type to translator-func
	psEvents := []pubsub.Event{}
	for _, e := range eventBatch.Events {
		switch e.GetType() {
		case "light":
			// a light event might contain multiple changes, a separate event is emitted for each change.
			psEvent, err := t.translateLightUpdate(e, eventBatch.TimeStamp)
			if err != nil {
				t.Logger.Error("failed to translate light update", zap.Error(err))
				continue
			}
			psEvents = append(psEvents, psEvent...)

		case "grouped_light":
			// a grouped_light event might contain multiple changes, a separate event is emitted for each change.
			psEvent, err := t.translateGroupedLightUpdate(e, eventBatch.TimeStamp)
			if err != nil {
				continue
			}
			psEvents = append(psEvents, psEvent...)
		}
	}
	return psEvents, nil
}

func (t *Translator) translateLightUpdate(e events.Event, ts time.Time) ([]pubsub.Event, error) {
	lightEvent, ok := e.(*events.LightUpdate)
	if !ok {
		return []pubsub.Event{}, fmt.Errorf("expected a *LightUpdate for event type 'light'")
	}

	humanID, ok := t.Client.ResourceRegistry.ResolveName(lightEvent.Type, lightEvent.ID)
	if !ok {
		t.Logger.Warn("failed to lookup name", zap.String("type", lightEvent.Type), zap.String("id", lightEvent.ID))
		return []pubsub.Event{}, fmt.Errorf("failed to lookup name")
	}

	psEvents := []pubsub.Event{}
	// a light event might contain multiple changes, a separate event is emitted for each change.
	for _, change := range lightEvent.ResolveStateChanges() {
		event := pubsub.Event{
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

func (t *Translator) translateGroupedLightUpdate(e events.Event, ts time.Time) ([]pubsub.Event, error) {
	groupedLightEvent, ok := e.(*events.GroupedLightUpdate)
	if !ok {
		return []pubsub.Event{}, fmt.Errorf("expected a *GroupedLightUpdate for event type 'grouped_light'")
	}

	humanID, ok := t.Client.ResourceRegistry.ResolveName(groupedLightEvent.Type, groupedLightEvent.ID)
	if !ok {
		return []pubsub.Event{}, fmt.Errorf("failed to lookup name")
	}

	psEvents := []pubsub.Event{}
	// a grouped_light event might contain multiple changes, a separate event is emitted for each change.
	for _, change := range groupedLightEvent.ResolveStateChanges() {
		event := pubsub.Event{
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
		psEvents = append(psEvents, event)
	}
	return psEvents, nil
}

func (t *Translator) LoadEvents() {
	for typ, constructor := range events.Registry {
		t.EventParser.RegisterEvent(typ, constructor)
	}
}
