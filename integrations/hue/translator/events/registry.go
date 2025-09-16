package events

var Registry = map[string]func() Event{
	"light":         func() Event { return &LightUpdate{} },
	"grouped_light": func() Event { return &LightUpdate{} },
}
