package events

type WheelEvent struct {
	Id     string `json:"id"`
	Counts int    `json:"counts"`
	Type   string `json:"type"`
}

func (w *WheelEvent) GetType() string { return w.Type }
