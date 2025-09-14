package types

// LightGet represents the root Hue Light object (response model)
type LightGet struct {
	ID                    string                         `json:"id"`
	IDV1                  string                         `json:"id_v1,omitempty"`
	Owner                 ResourceIdentifier             `json:"owner"`
	Type                  string                         `json:"type"` // always "light"
	Metadata              LightMetadataGet               `json:"metadata"`
	ProductData           *LightProductDataGet           `json:"product_data,omitempty"`
	Identify              LightIdentifyGet               `json:"identify"`
	ServiceID             int                            `json:"service_id"`
	On                    LightOnGet                     `json:"on"`
	Dimming               *LightDimmingGet               `json:"dimming,omitempty"`
	DimmingDelta          *LightDimmingDeltaGet          `json:"dimming_delta,omitempty"`
	ColorTemperature      *LightColorTemperatureGet      `json:"color_temperature,omitempty"`
	ColorTemperatureDelta *LightColorTemperatureDeltaGet `json:"color_temperature_delta,omitempty"`
	Color                 *LightColorGet                 `json:"color,omitempty"`
	Dynamics              *LightDynamicsGet              `json:"dynamics,omitempty"`
	Alert                 *LightAlertGet                 `json:"alert,omitempty"`
	Signaling             *LightSignalingGet             `json:"signaling,omitempty"`
	Mode                  string                         `json:"mode"` // "normal" | "streaming"
	Gradient              *LightGradientGet              `json:"gradient,omitempty"`
	Effects               *LightEffectsGet               `json:"effects,omitempty"`
	EffectsV2             *LightEffectsV2Get             `json:"effects_v2,omitempty"`
	TimedEffects          *LightTimedEffectsGet          `json:"timed_effects,omitempty"`
	PowerUp               *LightPowerUpGet               `json:"powerup,omitempty"`
	ContentConfiguration  *LightContentConfigurationGet  `json:"content_configuration,omitempty"`
}

// ---------- Metadata ----------

type LightMetadataGet struct {
	Name       string `json:"name"`
	Archetype  string `json:"archetype"`
	FixedMired *int   `json:"fixed_mired,omitempty"`
	Function   string `json:"function"`
}

type LightProductDataGet struct {
	Name      string `json:"name,omitempty"`
	Archetype string `json:"archetype,omitempty"`
	Function  string `json:"function"`
}

type LightIdentifyGet struct{}

// ---------- Basic Features ----------

type LightOnGet struct {
	On bool `json:"on"`
}

type LightDimmingGet struct {
	Brightness  float64  `json:"brightness"`
	MinDimLevel *float64 `json:"min_dim_level,omitempty"`
}

type LightDimmingDeltaGet struct {
	Action          string  `json:"action"`
	BrightnessDelta float64 `json:"brightness_delta"`
}

type LightColorTemperatureGet struct {
	Mirek       int                            `json:"mirek"`
	MirekValid  bool                           `json:"mirek_valid"`
	MirekSchema LightColorTemperatureSchemaGet `json:"mirek_schema"`
}

type LightColorTemperatureSchemaGet struct {
	MirekMinimum int `json:"mirek_minimum"`
	MirekMaximum int `json:"mirek_maximum"`
}

type LightColorTemperatureDeltaGet struct {
	Action     string `json:"action"`
	MirekDelta int    `json:"mirek_delta"`
}

// ---------- Color ----------

type LightColorGet struct {
	XY        XY             `json:"xy"`
	Gamut     *LightGamutGet `json:"gamut,omitempty"`
	GamutType string         `json:"gamut_type"`
}

type LightGamutGet struct {
	Red   XY `json:"red"`
	Green XY `json:"green"`
	Blue  XY `json:"blue"`
}

// ---------- Dynamics ----------

type LightDynamicsGet struct {
	Status       string   `json:"status"`
	StatusValues []string `json:"status_values"`
	Speed        float64  `json:"speed"`
	SpeedValid   bool     `json:"speed_valid"`
}

// ---------- Alert ----------

type LightAlertGet struct {
	ActionValues []string `json:"action_values"`
}

// ---------- Signaling ----------

type LightSignalingGet struct {
	SignalValues []string                 `json:"signal_values"`
	Status       *LightSignalingStatusGet `json:"status,omitempty"`
}

type LightSignalingStatusGet struct {
	Signal       string               `json:"signal"`
	EstimatedEnd string               `json:"estimated_end"`
	Colors       []LightColorBasicGet `json:"colors"`
}

type LightColorBasicGet struct {
	XY XY `json:"xy"`
}

// ---------- Gradient ----------

type LightGradientGet struct {
	Points        []LightGradientPointGet `json:"points"`
	Mode          string                  `json:"mode"`
	PointsCapable int                     `json:"points_capable"`
	ModeValues    []string                `json:"mode_values"`
	PixelCount    *int                    `json:"pixel_count,omitempty"`
}

type LightGradientPointGet struct {
	Color LightColorBasicGet `json:"color"`
}

// ---------- Effects ----------

type LightEffectsGet struct {
	StatusValues []string `json:"status_values"`
	Status       string   `json:"status"`
	EffectValues []string `json:"effect_values"`
}

type LightEffectsV2Get struct {
	Action LightEffectsV2ActionGet `json:"action"`
	Status LightEffectsV2StatusGet `json:"status"`
}

type LightEffectsV2ActionGet struct {
	EffectValues []string `json:"effect_values"`
}

type LightEffectsV2StatusGet struct {
	Effect       string                    `json:"effect"`
	EffectValues []string                  `json:"effect_values"`
	Parameters   *LightEffectParametersGet `json:"parameters,omitempty"`
}

type LightEffectParametersGet struct {
	Color            *LightColorBasicGet             `json:"color,omitempty"`
	ColorTemperature *LightColorTemperatureSimpleGet `json:"color_temperature,omitempty"`
	Speed            float64                         `json:"speed"`
}

type LightColorTemperatureSimpleGet struct {
	Mirek      int  `json:"mirek"`
	MirekValid bool `json:"mirek_valid"`
}

// ---------- Timed Effects ----------

type LightTimedEffectsGet struct {
	StatusValues []string `json:"status_values"`
	Status       string   `json:"status"`
	EffectValues []string `json:"effect_values"`
}

// ---------- Power Up ----------

type LightPowerUpGet struct {
	Preset     string                  `json:"preset"`
	Configured bool                    `json:"configured"`
	On         LightPowerUpOnGet       `json:"on"`
	Dimming    *LightPowerUpDimmingGet `json:"dimming,omitempty"`
	Color      *LightPowerUpColorGet   `json:"color,omitempty"`
}

type LightPowerUpOnGet struct {
	Mode string      `json:"mode"`
	On   *LightOnGet `json:"on,omitempty"`
}

type LightPowerUpDimmingGet struct {
	Mode    string           `json:"mode"`
	Dimming *LightDimmingGet `json:"dimming,omitempty"`
}

type LightPowerUpColorGet struct {
	Mode             string                          `json:"mode"`
	ColorTemperature *LightColorTemperatureSimpleGet `json:"color_temperature,omitempty"`
	Color            *LightColorBasicGet             `json:"color,omitempty"`
}

// ---------- Content Configuration ----------

type LightContentConfigurationGet struct {
	Orientation *LightContentOrientationGet `json:"orientation,omitempty"`
	Order       *LightContentOrderGet       `json:"order,omitempty"`
}

type LightContentOrientationGet struct {
	Status       string `json:"status"`
	Configurable bool   `json:"configurable"`
	Orientation  string `json:"orientation"`
}

type LightContentOrderGet struct {
	Status       string `json:"status"`
	Configurable bool   `json:"configurable"`
	Order        string `json:"order"`
}
