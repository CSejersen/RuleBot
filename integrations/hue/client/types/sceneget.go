package types

// SceneGet represents the root Hue Scene object (response model)
type SceneGet struct {
	ID          string             `json:"id"`
	IDV1        string             `json:"id_v1,omitempty"`
	Actions     []SceneActionGet   `json:"actions"`
	Palette     *ScenePaletteGet   `json:"palette,omitempty"`
	Recall      SceneRecallGet     `json:"recall"`
	Type        string             `json:"type"`
	Metadata    SceneMetadataGet   `json:"metadata"`
	Group       ResourceIdentifier `json:"group"`
	Speed       float64            `json:"speed"`
	AutoDynamic bool               `json:"auto_dynamic"`
	Status      SceneStatusGet     `json:"status"`
}

// ---------- Actions ----------

type SceneActionGet struct {
	Target ResourceIdentifier `json:"target"`
	Action SceneActionDetails `json:"action"`
}

type SceneActionDetails struct {
	On               *SceneOnGet               `json:"on,omitempty"`
	Dimming          *SceneDimmingGet          `json:"dimming,omitempty"`
	Color            *SceneColorGet            `json:"color,omitempty"`
	ColorTemperature *SceneColorTemperatureGet `json:"color_temperature,omitempty"`
	Gradient         *SceneGradientGet         `json:"gradient,omitempty"`
	Effects          *SceneEffectsGet          `json:"effects,omitempty"`
	EffectsV2        *SceneEffectsV2Get        `json:"effects_v2,omitempty"`
	Dynamics         *SceneDynamicsGet         `json:"dynamics,omitempty"`
}

// ---------- Features ----------

type SceneOnGet struct {
	On bool `json:"on"`
}

type SceneDimmingGet struct {
	Brightness float64 `json:"brightness"`
}

type SceneColorGet struct {
	XY XY `json:"xy"`
}

type SceneColorTemperatureGet struct {
	Mirek int `json:"mirek"`
}

// ---------- Gradient ----------

type SceneGradientGet struct {
	Points []SceneGradientPointGet `json:"points"`
	Mode   string                  `json:"mode"`
}

type SceneGradientPointGet struct {
	Color SceneColorGet `json:"color"`
}

// ---------- Effects ----------

type SceneEffectsGet struct {
	Effect string `json:"effect"`
}

type SceneEffectsV2Get struct {
	Action SceneEffectV2ActionGet `json:"action"`
}

type SceneEffectV2ActionGet struct {
	Effect     string                      `json:"effect"`
	Parameters *SceneEffectV2ParametersGet `json:"parameters,omitempty"`
}

type SceneEffectV2ParametersGet struct {
	Color            *SceneColorGet            `json:"color,omitempty"`
	ColorTemperature *SceneColorTemperatureGet `json:"color_temperature,omitempty"`
	Speed            *float64                  `json:"speed,omitempty"`
}

// ---------- Dynamics ----------

type SceneDynamicsGet struct {
	Duration int `json:"duration"`
}

// ---------- Palette ----------

type ScenePaletteGet struct {
	Color            []SceneColorPaletteGet            `json:"color"`
	Dimming          []SceneDimmingFeatureBasicGet     `json:"dimming"`
	ColorTemperature []SceneColorTemperaturePaletteGet `json:"color_temperature"`
	Effects          []SceneEffectFeatureBasicGet      `json:"effects"`
	EffectsV2        []SceneEffectV2FeatureBasicGet    `json:"effects_v2"`
}

type SceneColorPaletteGet struct {
	Color   SceneColorGet   `json:"color"`
	Dimming SceneDimmingGet `json:"dimming"`
}

type SceneDimmingFeatureBasicGet struct {
	Brightness float64 `json:"brightness"`
}

type SceneColorTemperaturePaletteGet struct {
	ColorTemperature SceneColorTemperatureGet `json:"color_temperature"`
	Dimming          SceneDimmingGet          `json:"dimming"`
}

type SceneEffectFeatureBasicGet struct {
	Effect string `json:"effect"`
}

type SceneEffectV2FeatureBasicGet struct {
	Action SceneEffectV2ActionGet `json:"action"`
}

// ---------- Metadata ----------

type SceneMetadataGet struct {
	Name    string              `json:"name"`
	Image   *ResourceIdentifier `json:"image,omitempty"`
	AppData string              `json:"appdata,omitempty"`
}

// ---------- Recall ----------

type SceneRecallGet struct {
	Action   string           `json:"action,omitempty"`   // "active", "dynamic_palette", "static"
	Duration *int             `json:"duration,omitempty"` // transition duration in ms
	Dimming  *SceneDimmingGet `json:"dimming,omitempty"`  // override brightness
}

// ---------- Scene Status ----------

type SceneStatusGet struct {
	Active     string  `json:"active"`
	LastRecall *string `json:"last_recall,omitempty"`
}
