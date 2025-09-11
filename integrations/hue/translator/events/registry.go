package events

var Registry = map[string]func() Event{
	"light": func() Event { return &LightUpdate{} },
}
