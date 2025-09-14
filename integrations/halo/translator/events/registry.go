package events

var Registry = map[string]func() Event{
	"wheel":  func() Event { return &WheelEvent{} },
	"button": func() Event { return &ButtonEvent{} },
	"system": func() Event { return &SystemEvent{} },
}
