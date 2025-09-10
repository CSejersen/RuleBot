package handlers

type UpdateCommand[T any] struct {
	Update T `json:"update"`
}
