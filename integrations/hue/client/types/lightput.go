package types

type LightPut struct {
	Type                  string                         `json:"type"` // always "light"
	Metadata              *LightMetadataPut              `json:"metadata,omitempty"`
	Identify              *LightIdentifyPut              `json:"identify,omitempty"`
	On                    *LightOnPut                    `json:"on,omitempty"`
	Dimming               *LightDimmingPut               `json:"dimming,omitempty"`
	DimmingDelta          *LightDimmingDeltaPut          `json:"dimming_delta,omitempty"`
	ColorTemperature      *LightColorTemperaturePut      `json:"color_temperature,omitempty"`
	ColorTemperatureDelta *LightColorTemperatureDeltaPut `json:"color_temperature_delta,omitempty"`
	Color                 *LightColorPut                 `json:"color,omitempty"`
	Dynamics              *LightDynamicsPut              `json:"dynamics,omitempty"`
	Alert                 *LightAlertPut                 `json:"alert,omitempty"`
	Signaling             *LightSignalingPut             `json:"signaling,omitempty"`
	Gradient              *LightGradientPut              `json:"gradient,omitempty"`
	EffectsV2             *LightEffectV2Put              `json:"effects_v2,omitempty"`
	TimedEffects          *LightTimedEffectsPut          `json:"timed_effects,omitempty"`
	PowerUp               *LightPowerUpPut               `json:"powerup,omitempty"`
	ContentConfiguration  *LightContentConfigurationPut  `json:"content_configuration,omitempty"`
}

// ---------- Metadata ----------

type LightMetadataPut struct {
	Name      *string `json:"name,omitempty"`
	Archetype *string `json:"archetype,omitempty"`
	Function  *string `json:"function,omitempty"`
}

type LightIdentifyPut struct {
	Action   string `json:"action"`   // always "identify"
	Duration int    `json:"duration"` // ms
}

type LightOnPut struct {
	On bool `json:"on"`
}

type LightDimmingPut struct {
	Brightness float64 `json:"brightness"`
}

type LightDimmingDeltaPut struct {
	Action          string  `json:"action"`           // up, down, stop
	BrightnessDelta float64 `json:"brightness_delta"` // percentage delta
}

type LightColorTemperaturePut struct {
	Mirek int `json:"mirek"`
}

type LightColorTemperatureDeltaPut struct {
	Action     string `json:"action"`      // up, down, stop
	MirekDelta int    `json:"mirek_delta"` // delta in mirek
}

type LightColorPut struct {
	XY XY `json:"xy"`
}

type LightDynamicsPut struct {
	Duration *int     `json:"duration,omitempty"`
	Speed    *float64 `json:"speed,omitempty"`
}

type LightAlertPut struct {
	Action string `json:"action"` // always "breathe"
}

type LightSignalingPut struct {
	Signal   string                 `json:"signal"`   // no_signal, on_off, on_off_color, alternating
	Duration int                    `json:"duration"` // ms
	Colors   []LightColorFeaturePut `json:"colors,omitempty"`
}

type LightColorFeaturePut struct {
	XY XY `json:"xy"`
}

type LightEffectV2Put struct {
	Action LightEffectV2ActionPut `json:"action"`
}

type LightEffectV2ActionPut struct {
	Effect     string                      `json:"effect"` // prism, opal, glisten, ...
	Parameters *LightEffectV2ParametersPut `json:"parameters,omitempty"`
}

type LightEffectV2ParametersPut struct {
	Color            *LightColorPut            `json:"color,omitempty"`
	ColorTemperature *LightColorTemperaturePut `json:"color_temperature,omitempty"`
	Speed            *float64                  `json:"speed,omitempty"`
}

type LightTimedEffectsPut struct {
	Effect   string `json:"effect"`   // sunrise, sunset, no_effect
	Duration int    `json:"duration"` // ms
}

// ---------- PowerUp ----------

type LightPowerUpPut struct {
	Preset  string                  `json:"preset"` // safety, powerfail, last_on_state, custom
	On      *LightPowerUpOnPut      `json:"on,omitempty"`
	Dimming *LightPowerUpDimmingPut `json:"dimming,omitempty"`
	Color   *LightPowerUpColorPut   `json:"color,omitempty"`
}

type LightPowerUpOnPut struct {
	Mode string      `json:"mode"` // on, toggle, previous
	On   *LightOnPut `json:"on,omitempty"`
}

type LightPowerUpDimmingPut struct {
	Mode    string           `json:"mode"` // dimming, previous
	Dimming *LightDimmingPut `json:"dimming,omitempty"`
}

type LightPowerUpColorPut struct {
	Mode             string                    `json:"mode"` // color_temperature, color, previous
	ColorTemperature *LightColorTemperaturePut `json:"color_temperature,omitempty"`
	Color            *LightColorPut            `json:"color,omitempty"`
}

// ---------- Gradient ----------

type LightGradientPut struct {
	Points []LightGradientPointPut `json:"points"`
	Mode   *string                 `json:"mode,omitempty"` // interpolated_palette, interpolated_palette_mirrored, random_pixelated, segmented_palette
}

type LightGradientPointPut struct {
	Color LightColorPut `json:"color"`
}

// ---------- Content Configuration ----------

type LightContentConfigurationPut struct {
	Orientation *LightOrientationPut `json:"orientation,omitempty"`
	Order       *LightOrderPut       `json:"order,omitempty"`
}

type LightOrientationPut struct {
	Orientation string `json:"orientation"` // horizontal, vertical
}

type LightOrderPut struct {
	Order string `json:"order"` // forward, reversed
}
