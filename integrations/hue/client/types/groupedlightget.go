package types

// GroupedLightGet represents a grouped_light resource GET response
type GroupedLightGet struct {
	ID                    string                    `json:"id"` // required
	IDV1                  *string                   `json:"id_v1,omitempty"`
	Owner                 ResourceIdentifier        `json:"owner"` // required
	Type                  string                    `json:"type"`  // always "grouped_light"
	On                    *OnGet                    `json:"on,omitempty"`
	Dimming               *DimmingGet               `json:"dimming,omitempty"`
	DimmingDelta          *DimmingDeltaGet          `json:"dimming_delta,omitempty"`
	ColorTemperature      *ColorTemperatureGet      `json:"color_temperature,omitempty"`
	ColorTemperatureDelta *ColorTemperatureDeltaGet `json:"color_temperature_delta,omitempty"`
	Color                 *ColorGet                 `json:"color,omitempty"`
	Alert                 *AlertGet                 `json:"alert,omitempty"`
	Signaling             *SignalingGet             `json:"signaling,omitempty"`
	Dynamics              *DynamicsGet              `json:"dynamics,omitempty"`
}

func (g *GroupedLightGet) GetType() string {
	return g.Type
}

func (g *GroupedLightGet) GetID() string {
	return g.ID
}
