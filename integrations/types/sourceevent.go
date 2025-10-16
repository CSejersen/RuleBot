package types

type SourceEvent interface {
	GetType() string
}

type EventData struct {
	Constructor  func() SourceEvent
	StateChanges []string
}
