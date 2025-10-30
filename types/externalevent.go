package types

type ExternalEvent interface {
	GetType() string
}

type ExternalEventDescriptor struct {
	Constructor func() ExternalEvent
}
