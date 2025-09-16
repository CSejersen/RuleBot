package types

// GroupedLightPut represents a PUT request for a grouped_light resource
type GroupedLightPut struct {
	On                    *OnPut                    `json:"on,omitempty"`
	Dimming               *DimmingPut               `json:"dimming,omitempty"`
	DimmingDelta          *DimmingDeltaPut          `json:"dimming_delta,omitempty"`
	ColorTemperature      *ColorTemperaturePut      `json:"color_temperature,omitempty"`
	ColorTemperatureDelta *ColorTemperatureDeltaPut `json:"color_temperature_delta,omitempty"`
	Color                 *ColorPut                 `json:"color,omitempty"`
	Alert                 *AlertPut                 `json:"alert,omitempty"`
	Signaling             *SignalingPut             `json:"signaling,omitempty"`
	Dynamics              *DynamicsPut              `json:"dynamics,omitempty"`
}
