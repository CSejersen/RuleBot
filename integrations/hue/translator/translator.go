package translator

import (
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/integrations/hue/apiclient"
	"home_automation_server/integrations/hue/translator/events"
	"home_automation_server/pubsub"
	"time"
)

// Translator translates a parsed hue event into a pubsub.Event
type Translator struct {
	Client      *apiclient.ApiClient
	EventParser EventParser
	Logger      *zap.Logger

	// Registry: Type + ID -> Human-readable name
	Registry map[string]map[string]string
}

// New is the constructor for Translator
func New(client *apiclient.ApiClient, logger *zap.Logger) (*Translator, error) {
	t := &Translator{
		Client:      client,
		EventParser: NewEventParser(logger),
		Logger:      logger,
	}

	if err := t.init(); err != nil {
		return nil, fmt.Errorf("failed to init translator: %w", err)
	}

	return t, nil
}

// TODO: implement continouos refreshing of the registry to keep it aligned with the bridge state.
// maybe events are published on state changes otherwise we refresh on some interval
func (t *Translator) init() error {
	t.LoadEvents()

	// init registry of human-readable id's
	t.Registry = make(map[string]map[string]string)

	// Lights
	lights, err := t.Client.Lights()
	if err != nil {
		return err
	}

	t.Registry["light"] = make(map[string]string)
	for _, l := range lights {
		t.Registry["light"][l.ID] = l.Metadata.Name
	}

	// Rooms
	// Scenes
	return nil
}

func (t *Translator) Translate(raw []byte) ([]pubsub.Event, error) {
	eventBatch, err := t.EventParser.parse(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse events: %w", err)
	}

	// TODO: if switch case grows beyond what is reasonable we could implement a map of event type to translator-func
	psEvents := []pubsub.Event{}
	for _, e := range eventBatch.Events {
		switch e.GetType() {
		case "light":
			psEvent, err := t.translateLightUpdate(e, eventBatch.TimeStamp)
			if err != nil {
				t.Logger.Error("failed to translate light update", zap.Error(err))
				continue
			}
			psEvents = append(psEvents, psEvent)
		default:
			t.Logger.Info("translator not implemented, skipping event", zap.String("type", e.GetType()))
		}
	}
	return psEvents, nil
}

func (t *Translator) translateLightUpdate(e events.Event, ts time.Time) (pubsub.Event, error) {
	lightEvent, ok := e.(*events.LightUpdate)
	if !ok {
		return pubsub.Event{}, fmt.Errorf("expected a *LightUpdate for event type 'light'")
	}

	humanID, ok := t.LookupName(lightEvent.Type, lightEvent.ID)
	if !ok {
		t.Logger.Warn("failed to lookup name", zap.String("type", lightEvent.Type), zap.String("id", lightEvent.ID))
		humanID = lightEvent.ID // fallback
	}

	return pubsub.Event{
		Source:      "hue",
		Type:        "light",
		Entity:      humanID,
		StateChange: lightEvent.ResolveStateChange(),
		Payload: map[string]interface{}{
			"brightness": lightEvent.SafeBrightness(),
			"on":         lightEvent.SafeOn(),
			"metadata":   lightEvent.Metadata,
		},
		Time: ts,
	}, nil
}

func (t *Translator) LookupName(eventType, id string) (string, bool) {
	if typeMap, ok := t.Registry[eventType]; ok {
		if name, ok := typeMap[id]; ok {
			return name, true
		}
	}
	return "", false
}

func (t *Translator) LoadEvents() {
	for typ, constructor := range events.Registry {
		t.RegisterEvent(typ, constructor)
	}
}
