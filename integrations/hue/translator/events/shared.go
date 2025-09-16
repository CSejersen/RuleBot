package events

type Event interface {
	GetType() string
}

type XY struct {
	X float64 `json:"x"` // 0-1
	Y float64 `json:"y"` // 0-1
}

type OnUpdate struct {
	On bool `json:"on"`
}

type DimmingUpdate struct {
	Brightness float64 `json:"brightness"`
}

type DimmingDeltaUpdate struct {
	Action          string  `json:"action"`           // up, down, stop
	BrightnessDelta float64 `json:"brightness_delta"` // percentage delta
}

type ColorUpdate struct {
	XY XY `json:"xy"`
}

type ColorTemperatureUpdate struct {
	Mirek int `json:"mirek"`
}

type ColorTemperatureDeltaUpdate struct {
	Action     string `json:"action"`      // up, down, stop
	MirekDelta int    `json:"mirek_delta"` // delta in mirek
}

type GradientUpdate struct {
	Points []GradientPointUpdate `json:"points"`
	Mode   *string               `json:"mode,omitempty"` // interpolated_palette, random_pixelated, etc.
}

type GradientPointUpdate struct {
	Color ColorUpdate `json:"color"`
}

type DynamicsUpdate struct {
	Duration *int     `json:"duration,omitempty"`
	Speed    *float64 `json:"speed,omitempty"`
}

type AlertUpdate struct {
	Action string `json:"action"` // e.g. breathe
}

type SignalingUpdate struct {
	Signal   string               `json:"signal"`   // no_signal, on_off, on_off_color, alternating
	Duration int                  `json:"duration"` // ms
	Colors   []ColorFeatureUpdate `json:"colors,omitempty"`
}

type ColorFeatureUpdate struct {
	XY XY `json:"xy"`
}

type EffectV2Update struct {
	Action EffectV2ActionUpdate `json:"action"`
}

type EffectV2ActionUpdate struct {
	Effect     string                    `json:"effect"`
	Parameters *EffectV2ParametersUpdate `json:"parameters,omitempty"`
}

type EffectV2ParametersUpdate struct {
	Color            *ColorUpdate            `json:"color,omitempty"`
	ColorTemperature *ColorTemperatureUpdate `json:"color_temperature,omitempty"`
	Speed            *float64                `json:"speed,omitempty"`
}

type TimedEffectsUpdate struct {
	Effect   string `json:"effect"`   // sunrise, sunset, no_effect
	Duration int    `json:"duration"` // ms
}

type MetadataUpdate struct {
	Name      *string `json:"name,omitempty"`
	Archetype *string `json:"archetype,omitempty"`
	Function  *string `json:"function,omitempty"`
}
