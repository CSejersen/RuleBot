package types

type LightPut struct {
	Type                  string                    `json:"type"` // always "light"
	Metadata              *MetadataPut              `json:"metadata,omitempty"`
	Identify              *IdentifyPut              `json:"identify,omitempty"`
	On                    *OnPut                    `json:"on,omitempty"`
	Dimming               *DimmingPut               `json:"dimming,omitempty"`
	DimmingDelta          *DimmingDeltaPut          `json:"dimming_delta,omitempty"`
	ColorTemperature      *ColorTemperaturePut      `json:"color_temperature,omitempty"`
	ColorTemperatureDelta *ColorTemperatureDeltaPut `json:"color_temperature_delta,omitempty"`
	Color                 *ColorPut                 `json:"color,omitempty"`
	Dynamics              *DynamicsPut              `json:"dynamics,omitempty"`
	Alert                 *AlertPut                 `json:"alert,omitempty"`
	Signaling             *SignalingPut             `json:"signaling,omitempty"`
	Gradient              *GradientPut              `json:"gradient,omitempty"`
	EffectsV2             *EffectV2Put              `json:"effects_v2,omitempty"`
	TimedEffects          *TimedEffectsPut          `json:"timed_effects,omitempty"`
	PowerUp               *PowerUpPut               `json:"powerup,omitempty"`
	ContentConfiguration  *ContentConfigurationPut  `json:"content_configuration,omitempty"`
}
