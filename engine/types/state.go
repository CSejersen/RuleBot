package types

import "time"

type StateGetter interface {
	GetState(source string, typ string, entity string) (*State, bool)
}

type State struct {
	Integration string
	Type        string
	Entity      string
	Fields      map[string]any
	LastSeen    time.Time
}
