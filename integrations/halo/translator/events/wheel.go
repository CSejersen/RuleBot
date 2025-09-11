package events

type WheelEvent struct {
	ID     string `json:"id"`
	Counts int    `json:"counts"`
	Type   string `json:"type"`
}

func (w *WheelEvent) GetType() string { return w.Type }

// TODO: stop using magic strings all over, typing would be nice for these scenarios.
