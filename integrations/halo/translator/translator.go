package translator

import (
	"errors"
	"go.uber.org/zap"
	"home_automation_server/engine/pubsub"
	"home_automation_server/integrations/halo/client"
	"home_automation_server/integrations/halo/translator/events"
	"time"
)

const wheelBufferDuration = 300 * time.Millisecond

type Translator struct {
	Client      *client.Client
	EventParser EventParser
	Logger      *zap.Logger

	// Registry: ID -> Human-readable name
	Registry map[string]string

	wheelBuf wheelBuffer
}

type wheelBuffer struct {
	count     int
	lastFlush time.Time
}

func New(client *client.Client, logger *zap.Logger) *Translator {
	t := &Translator{
		Client:      client,
		EventParser: newEventParser(logger),
		Logger:      logger,
	}
	t.init()
	return t
}

func (t *Translator) init() {
	t.LoadEvents() // load all events registered in events/registry

	// init registry of human-readable id's
	t.Registry = make(map[string]string)
	buttons := t.Client.Buttons()
	for _, b := range buttons {
		t.Registry[b.ID] = b.Title
	}
}

func (t *Translator) Translate(raw []byte) ([]pubsub.Event, error) {
	event, err := t.EventParser.ParseEvent(raw)
	if err != nil {
		return nil, err
	}

	// If number of switch cases grows beyond what is reasonable we could implement a map of event type to translator-func
	switch event.GetType() {
	case "wheel":
		psEvent, err := t.translateWheelEvent(event)
		if err != nil {
			t.Logger.Error("failed to translate wheel event", zap.Error(err))
			return []pubsub.Event{}, err
		}
		return []pubsub.Event{psEvent}, nil

	case "button":
		psEvent, err := t.translateButtonEvent(event)
		if err != nil {
			t.Logger.Error("failed to translate button event", zap.Error(err))
			return []pubsub.Event{}, err
		}
		return []pubsub.Event{psEvent}, nil

	case "system":
		psEvent, err := t.translateSystemEvent(event)
		if err != nil {
			t.Logger.Error("failed to translate system event", zap.Error(err))
			return []pubsub.Event{}, err
		}
		return []pubsub.Event{psEvent}, nil

	default:
		t.Logger.Info("translator not implemented, skipping event", zap.String("type", event.GetType()))
	}
	return []pubsub.Event{}, nil
}

func (t *Translator) translateSystemEvent(e events.Event) (pubsub.Event, error) {
	sysEvent, ok := e.(*events.SystemEvent)
	if !ok {
		return pubsub.Event{}, nil
	}

	return pubsub.Event{
		Source:      "halo",
		Type:        "system",
		StateChange: sysEvent.State,
		Payload:     map[string]any{},
		Time:        time.Now(),
	}, nil
}

func (t *Translator) translateWheelEvent(e events.Event) (pubsub.Event, error) {
	wheelEvent, ok := e.(*events.WheelEvent)
	if !ok {
		return pubsub.Event{}, errors.New("expected a *WheelEvent for event type 'wheel'")
	}

	humanID, ok := t.LookupName(wheelEvent.ID)
	if !ok {
		t.Logger.Warn("failed to lookup human id", zap.String("id", wheelEvent.ID))
	}

	return pubsub.Event{
		Source:      "halo",
		Type:        "wheel",
		Entity:      humanID,
		StateChange: ResolveWheelStateChange(wheelEvent.Counts),
		Payload: map[string]any{
			"step": wheelEvent.Counts,
		},
		Time: time.Now(),
	}, nil
}

func (t *Translator) translateButtonEvent(e events.Event) (pubsub.Event, error) {
	buttonEvent, ok := e.(*events.ButtonEvent)
	if !ok {
		return pubsub.Event{}, errors.New("expected a *ButtonEvent for event type 'button'")
	}

	humanID, ok := t.LookupName(buttonEvent.ID)
	if !ok {
		t.Logger.Warn("failed to lookup human id", zap.String("id", buttonEvent.ID))
	}

	return pubsub.Event{
		Source:      "halo",
		Type:        "button",
		Entity:      humanID,
		StateChange: buttonEvent.State,
		Payload:     nil,
		Time:        time.Now(),
	}, nil
}

func ResolveWheelStateChange(count int) string {
	if count > 0 {
		return "clockwise"
	}
	return "counter_clockwise"
}

func (t *Translator) LookupName(id string) (string, bool) {
	if name, ok := t.Registry[id]; ok {
		return name, true
	}
	return "", false
}

func (t *Translator) LoadEvents() {
	for typ, constructor := range events.Registry {
		t.EventParser.RegisterEvent(typ, constructor)
	}
}
