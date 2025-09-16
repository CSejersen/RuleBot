package types

// Get types
type PutResponse struct {
	Errors []ApiError           `json:"errors"`
	Data   []ResourceIdentifier `json:"data"`
}

type ApiError struct {
	Description string `json:"description"`
}

type ResourceIdentifier struct {
	RID   string `json:"rid"`   // UUID of referenced resource
	RType string `json:"rtype"` // type of resource, e.g., device, grouped_light, etc.
}

type XY struct {
	X float64 `json:"x"` // 0-1
	Y float64 `json:"y"` // 0-1
}

type MetadataGet struct {
	Name       string `json:"name"`
	Archetype  string `json:"archetype"`
	FixedMired *int   `json:"fixed_mired,omitempty"`
	Function   string `json:"function"`
}

type ProductDataGet struct {
	Name      string `json:"name,omitempty"`
	Archetype string `json:"archetype,omitempty"`
	Function  string `json:"function"`
}

type OnGet struct {
	On bool `json:"on"`
}

type DimmingGet struct {
	Brightness  float64  `json:"brightness"`
	MinDimLevel *float64 `json:"min_dim_level,omitempty"`
}

type DimmingDeltaGet struct {
	Action          string  `json:"action"`
	BrightnessDelta float64 `json:"brightness_delta"`
}

type ColorTemperatureGet struct {
	Mirek       int                       `json:"mirek"`
	MirekValid  bool                      `json:"mirek_valid"`
	MirekSchema ColorTemperatureSchemaGet `json:"mirek_schema"`
}

type ColorTemperatureSchemaGet struct {
	MirekMinimum int `json:"mirek_minimum"`
	MirekMaximum int `json:"mirek_maximum"`
}

type ColorTemperatureDeltaGet struct {
	Action     string `json:"action"`
	MirekDelta int    `json:"mirek_delta"`
}

type ColorGet struct {
	XY        XY        `json:"xy"`
	Gamut     *GamutGet `json:"gamut,omitempty"`
	GamutType string    `json:"gamut_type"`
}

type GamutGet struct {
	Red   XY `json:"red"`
	Green XY `json:"green"`
	Blue  XY `json:"blue"`
}

type DynamicsGet struct {
	Status       string   `json:"status"`
	StatusValues []string `json:"status_values"`
	Speed        float64  `json:"speed"`
	SpeedValid   bool     `json:"speed_valid"`
}

type AlertGet struct {
	ActionValues []string `json:"action_values"`
}

type SignalingGet struct {
	SignalValues []string            `json:"signal_values"`
	Status       *SignalingStatusGet `json:"status,omitempty"`
}

type SignalingStatusGet struct {
	Signal       string          `json:"signal"`
	EstimatedEnd string          `json:"estimated_end"`
	Colors       []ColorBasicGet `json:"colors"`
}

type ColorBasicGet struct {
	XY XY `json:"xy"`
}

type GradientGet struct {
	Points        []GradientPointGet `json:"points"`
	Mode          string             `json:"mode"`
	PointsCapable int                `json:"points_capable"`
	ModeValues    []string           `json:"mode_values"`
	PixelCount    *int               `json:"pixel_count,omitempty"`
}

type GradientPointGet struct {
	Color ColorBasicGet `json:"color"`
}

type EffectsGet struct {
	StatusValues []string `json:"status_values"`
	Status       string   `json:"status"`
	EffectValues []string `json:"effect_values"`
}

type EffectsV2Get struct {
	Action EffectsV2ActionGet `json:"action"`
	Status EffectsV2StatusGet `json:"status"`
}

type EffectsV2ActionGet struct {
	EffectValues []string `json:"effect_values"`
}

type EffectsV2StatusGet struct {
	Effect       string               `json:"effect"`
	EffectValues []string             `json:"effect_values"`
	Parameters   *EffectParametersGet `json:"parameters,omitempty"`
}

type EffectParametersGet struct {
	Color            *ColorBasicGet             `json:"color,omitempty"`
	ColorTemperature *ColorTemperatureSimpleGet `json:"color_temperature,omitempty"`
	Speed            float64                    `json:"speed"`
}

type ColorTemperatureSimpleGet struct {
	Mirek      int  `json:"mirek"`
	MirekValid bool `json:"mirek_valid"`
}

type TimedEffectsGet struct {
	StatusValues []string `json:"status_values"`
	Status       string   `json:"status"`
	EffectValues []string `json:"effect_values"`
}

type PowerUpGet struct {
	Preset     string             `json:"preset"`
	Configured bool               `json:"configured"`
	On         PowerUpOnGet       `json:"on"`
	Dimming    *PowerUpDimmingGet `json:"dimming,omitempty"`
	Color      *PowerUpColorGet   `json:"color,omitempty"`
}

type PowerUpOnGet struct {
	Mode string `json:"mode"`
	On   *OnGet `json:"on,omitempty"`
}

type PowerUpDimmingGet struct {
	Mode    string      `json:"mode"`
	Dimming *DimmingGet `json:"dimming,omitempty"`
}

type PowerUpColorGet struct {
	Mode             string                     `json:"mode"`
	ColorTemperature *ColorTemperatureSimpleGet `json:"color_temperature,omitempty"`
	Color            *ColorBasicGet             `json:"color,omitempty"`
}

type ContentConfigurationGet struct {
	Orientation *ContentOrientationGet `json:"orientation,omitempty"`
	Order       *ContentOrderGet       `json:"order,omitempty"`
}

type ContentOrientationGet struct {
	Status       string `json:"status"`
	Configurable bool   `json:"configurable"`
	Orientation  string `json:"orientation"`
}

type ContentOrderGet struct {
	Status       string `json:"status"`
	Configurable bool   `json:"configurable"`
	Order        string `json:"order"`
}

// Put types
type MetadataPut struct {
	Name      *string `json:"name,omitempty"`
	Archetype *string `json:"archetype,omitempty"`
	Function  *string `json:"function,omitempty"`
}

type IdentifyPut struct {
	Action   string `json:"action"`   // always "identify"
	Duration int    `json:"duration"` // ms
}

type OnPut struct {
	On bool `json:"on"`
}

type DimmingPut struct {
	Brightness float64 `json:"brightness"`
}

type DimmingDeltaPut struct {
	Action          string  `json:"action"`           // up, down, stop
	BrightnessDelta float64 `json:"brightness_delta"` // percentage delta
}

type ColorTemperaturePut struct {
	Mirek int `json:"mirek"`
}

type ColorTemperatureDeltaPut struct {
	Action     string `json:"action"`      // up, down, stop
	MirekDelta int    `json:"mirek_delta"` // delta in mirek
}

type ColorPut struct {
	XY XY `json:"xy"`
}

type DynamicsPut struct {
	Duration *int     `json:"duration,omitempty"`
	Speed    *float64 `json:"speed,omitempty"`
}

type AlertPut struct {
	Action string `json:"action"` // always "breathe"
}

type SignalingPut struct {
	Signal   string            `json:"signal"`   // no_signal, on_off, on_off_color, alternating
	Duration int               `json:"duration"` // ms
	Colors   []ColorFeaturePut `json:"colors,omitempty"`
}

type ColorFeaturePut struct {
	XY XY `json:"xy"`
}

type EffectV2Put struct {
	Action EffectV2ActionPut `json:"action"`
}

type EffectV2ActionPut struct {
	Effect     string                 `json:"effect"` // prism, opal, glisten, ...
	Parameters *EffectV2ParametersPut `json:"parameters,omitempty"`
}

type EffectV2ParametersPut struct {
	Color            *ColorPut            `json:"color,omitempty"`
	ColorTemperature *ColorTemperaturePut `json:"color_temperature,omitempty"`
	Speed            *float64             `json:"speed,omitempty"`
}

type TimedEffectsPut struct {
	Effect   string `json:"effect"`   // sunrise, sunset, no_effect
	Duration int    `json:"duration"` // ms
}

type PowerUpPut struct {
	Preset  string             `json:"preset"` // safety, powerfail, last_on_state, custom
	On      *PowerUpOnPut      `json:"on,omitempty"`
	Dimming *PowerUpDimmingPut `json:"dimming,omitempty"`
	Color   *PowerUpColorPut   `json:"color,omitempty"`
}

type PowerUpOnPut struct {
	Mode string `json:"mode"` // on, toggle, previous
	On   *OnPut `json:"on,omitempty"`
}

type PowerUpDimmingPut struct {
	Mode    string      `json:"mode"` // dimming, previous
	Dimming *DimmingPut `json:"dimming,omitempty"`
}

type PowerUpColorPut struct {
	Mode             string               `json:"mode"` // color_temperature, color, previous
	ColorTemperature *ColorTemperaturePut `json:"color_temperature,omitempty"`
	Color            *ColorPut            `json:"color,omitempty"`
}

type GradientPut struct {
	Points []GradientPointPut `json:"points"`
	Mode   *string            `json:"mode,omitempty"` // interpolated_palette, interpolated_palette_mirrored, random_pixelated, segmented_palette
}

type GradientPointPut struct {
	Color ColorPut `json:"color"`
}

type ContentConfigurationPut struct {
	Orientation *OrientationPut `json:"orientation,omitempty"`
	Order       *OrderPut       `json:"order,omitempty"`
}

type OrientationPut struct {
	Orientation string `json:"orientation"` // horizontal, vertical
}

type OrderPut struct {
	Order string `json:"order"` // forward, reversed
}
