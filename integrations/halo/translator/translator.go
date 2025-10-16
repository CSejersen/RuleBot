package translator

import (
	"errors"
	"go.uber.org/zap"
	enginetypes "home_automation_server/engine/types"
	"home_automation_server/integrations/halo/client"
	"home_automation_server/integrations/halo/translator/events"
	"home_automation_server/integrations/types"
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

func (t *Translator) Translate(raw []byte) ([]enginetypes.Event, error) {
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
			return []enginetypes.Event{}, err
		}
		return []enginetypes.Event{translated}, nil

	case "button":
		translated, err := t.translateButtonEvent(event)
		if err != nil {
			t.Logger.Error("failed to translate button types", zap.Error(err))
			return []enginetypes.Event{}, err
		}
		return []enginetypes.Event{translated}, nil

	case "system":
		translated, err := t.translateSystemEvent(event)
		if err != nil {
			t.Logger.Error("failed to translate system types", zap.Error(err))
			return []enginetypes.Event{}, err
		}
		return []enginetypes.Event{translated}, nil

	default:
		t.Logger.Info("translator not implemented, skipping types", zap.String("type", event.GetType()))
	}
	return []enginetypes.Event{}, nil
}

func (t *Translator) translateSystemEvent(e types.SourceEvent) (enginetypes.Event, error) {
	sysEvent, ok := e.(*events.SystemEvent)
	if !ok {
		return enginetypes.Event{}, nil
	}

	return enginetypes.Event{
		Source:      "halo",
		Type:        "system",
		StateChange: sysEvent.State,
		Payload:     map[string]any{},
		Time:        time.Now(),
	}, nil
}

func (t *Translator) translateWheelEvent(e types.SourceEvent) (enginetypes.Event, error) {
	wheelEvent, ok := e.(*events.WheelEvent)
	if !ok {
		return enginetypes.Event{}, errors.New("expected a *WheelEvent for types type 'wheel'")
	}

	humanID, ok := t.LookupName(wheelEvent.ID)
	if !ok {
		t.Logger.Warn("failed to lookup human id", zap.String("id", wheelEvent.ID))
	}

	return enginetypes.Event{
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

func (t *Translator) translateButtonEvent(e types.SourceEvent) (enginetypes.Event, error) {
	buttonEvent, ok := e.(*events.ButtonEvent)
	if !ok {
		return enginetypes.Event{}, errors.New("expected a *ButtonEvent for types type 'button'")
	}

	humanID, ok := t.LookupName(buttonEvent.ID)
	if !ok {
		t.Logger.Warn("failed to lookup human id", zap.String("id", buttonEvent.ID))
	}

	return enginetypes.Event{
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

func (t *Translator) StateChangesForType(typ string) []string {
	return t.EventParser.EventRegistry[typ].StateChanges
}
