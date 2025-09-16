package types

// LightGet represents the root Hue Light object (response model)
type LightGet struct {
	ID                    string                    `json:"id"`
	IDV1                  string                    `json:"id_v1,omitempty"`
	Owner                 ResourceIdentifier        `json:"owner"`
	Type                  string                    `json:"type"` // always "light"
	Metadata              MetadataGet               `json:"metadata"`
	ProductData           *ProductDataGet           `json:"product_data,omitempty"`
	ServiceID             int                       `json:"service_id"`
	On                    OnGet                     `json:"on"`
	Dimming               *DimmingGet               `json:"dimming,omitempty"`
	DimmingDelta          *DimmingDeltaGet          `json:"dimming_delta,omitempty"`
	ColorTemperature      *ColorTemperatureGet      `json:"color_temperature,omitempty"`
	ColorTemperatureDelta *ColorTemperatureDeltaGet `json:"color_temperature_delta,omitempty"`
	Color                 *ColorGet                 `json:"color,omitempty"`
	Dynamics              *DynamicsGet              `json:"dynamics,omitempty"`
	Alert                 *AlertGet                 `json:"alert,omitempty"`
	Signaling             *SignalingGet             `json:"signaling,omitempty"`
	Mode                  string                    `json:"mode"` // "normal" | "streaming"
	Gradient              *GradientGet              `json:"gradient,omitempty"`
	Effects               *EffectsGet               `json:"effects,omitempty"`
	EffectsV2             *EffectsV2Get             `json:"effects_v2,omitempty"`
	TimedEffects          *TimedEffectsGet          `json:"timed_effects,omitempty"`
	PowerUp               *PowerUpGet               `json:"powerup,omitempty"`
	ContentConfiguration  *ContentConfigurationGet  `json:"content_configuration,omitempty"`
}

func (l *LightGet) GetType() string {
	return l.Type
}

func (l *LightGet) GetID() string {
	return l.ID
}
