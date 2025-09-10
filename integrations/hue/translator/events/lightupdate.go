package events

type LightUpdate struct {
	ID           string        `json:"id"`
	Owner        Owner         `json:"owner"`
	On           *On           `json:"on,omitempty"`
	Dimming      *Dimming      `json:"dimming,omitempty"`
	DimmingDelta *DimmingDelta `json:"dimming_delta,omitempty"`
	Metadata     *Metadata     `json:"metadata,omitempty"`
	Type         string        `json:"type"`
}

func (l LightUpdate) GetType() string {
	return l.Type
}

func (l LightUpdate) ResolveStateChange() string {
	switch {
	case l.SafeBrightness() != nil:
		return "brightness"
	case l.SafeOn() != nil:
		if l.On.On {
			return "on"
		}
		return "off"

	default:
		return ""
	}
}

func (l LightUpdate) SafeBrightness() *float64 {
	if l.Dimming != nil {
		return &l.Dimming.Brightness
	}
	return nil
}

func (l LightUpdate) SafeOn() *bool {
	if l.On != nil {
		return &l.On.On
	}
	return nil
}

func (l LightUpdate) SafeDimmingDelta() *DimmingDelta {
	if l.DimmingDelta != nil {
		return l.DimmingDelta
	}
	return nil
}
