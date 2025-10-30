package types

type StateStore interface {
	Get(entityID string) (State, bool)
	Set(entityID string, newState State)
	GetAll() []State
}
