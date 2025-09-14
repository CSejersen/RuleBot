package events

type GroupedLightUpdate struct {
	ID           string        `json:"id"`
	Owner        Owner         `json:"owner"`
	On           *On           `json:"on,omitempty"`
	Dimming      *Dimming      `json:"dimming,omitempty"`
	DimmingDelta *DimmingDelta `json:"dimming_delta,omitempty"`
	Metadata     *Metadata     `json:"metadata,omitempty"`
	Type         string        `json:"type"`
}
