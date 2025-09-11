package events

var Registry = map[string]func() Event{
	"wheel": func() Event { return &WheelEvent{} },
}
