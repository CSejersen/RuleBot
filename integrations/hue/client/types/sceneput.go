package types

type ScenePut struct {
	Actions []SceneActionPut `json:"actions"`
}

// ---------- SceneActionPut ----------

type SceneActionPut struct {
	Target      *ResourceIdentifier   `json:"target"`
	Action      *SceneActionActionPut `json:"action,omitempty"`
	Palette     *ScenePalettePut      `json:"palette,omitempty"`
	Recall      *SceneRecallPut       `json:"recall,omitempty"`
	Type        *string               `json:"type,omitempty"` // always "scene"
	Metadata    *SceneMetadataPut     `json:"metadata,omitempty"`
	Speed       *float64              `json:"speed,omitempty"` // 0-1
	AutoDynamic *bool                 `json:"auto_dynamic,omitempty"`
}

// ---------- SceneActionActionPut ----------

type SceneActionActionPut struct {
	On               *SceneOnPut               `json:"on,omitempty"`
	Dimming          *SceneDimmingPut          `json:"dimming,omitempty"`
	Color            *SceneColorPut            `json:"color,omitempty"`
	ColorTemperature *SceneColorTemperaturePut `json:"color_temperature,omitempty"`
	Gradient         *SceneGradientPut         `json:"gradient,omitempty"`
	Effects          *SceneEffectPut           `json:"effects,omitempty"` // deprecated
	EffectsV2        *SceneEffectV2Put         `json:"effects_v2,omitempty"`
	Dynamics         *SceneDynamicsPut         `json:"dynamics,omitempty"`
}

// ---------- SceneOnPut ----------

type SceneOnPut struct {
	On bool `json:"on"`
}

// ---------- SceneDimmingPut ----------

type SceneDimmingPut struct {
	Brightness float64 `json:"brightness"`
}

// ---------- SceneColorPut ----------

type SceneColorPut struct {
	XY XY `json:"xy"`
}

// ---------- SceneColorTemperaturePut ----------

type SceneColorTemperaturePut struct {
	Mirek int `json:"mirek"`
}

// ---------- SceneGradientPut ----------

type SceneGradientPut struct {
	Points []SceneGradientPointPut `json:"points"`
	Mode   *string                 `json:"mode,omitempty"` // interpolated_palette, interpolated_palette_mirrored, etc.
}

type SceneGradientPointPut struct {
	Color SceneColorPut `json:"color"`
}

// ---------- SceneEffectPut ----------

type SceneEffectPut struct {
	Effect string `json:"effect"` // prism, opal, glisten, candle, etc.
}

// ---------- SceneEffectV2Put ----------

type SceneEffectV2Put struct {
	Action SceneEffectV2ActionPut `json:"action"`
}

type SceneEffectV2ActionPut struct {
	Effect     string                      `json:"effect"` // prism, opal, glisten, candle, etc.
	Parameters *SceneEffectV2ParametersPut `json:"parameters,omitempty"`
}

type SceneEffectV2ParametersPut struct {
	Color            *SceneColorPut            `json:"color,omitempty"`
	ColorTemperature *SceneColorTemperaturePut `json:"color_temperature,omitempty"`
	Speed            *float64                  `json:"speed,omitempty"` // 0-1
}

// ---------- SceneDynamicsPut ----------

type SceneDynamicsPut struct {
	Duration *int     `json:"duration,omitempty"` // ms
	Speed    *float64 `json:"speed,omitempty"`    // 0-1
}

// ---------- ScenePalettePut ----------

type ScenePalettePut struct {
	Color            []SceneColorPalettePut            `json:"color,omitempty"`             // 0-9
	Dimming          []SceneDimmingFeatureBasicPut     `json:"dimming,omitempty"`           // 0-1
	ColorTemperature []SceneColorTemperaturePalettePut `json:"color_temperature,omitempty"` // 0-1
	Effects          []SceneEffectFeatureBasicPut      `json:"effects,omitempty"`           // deprecated
	EffectsV2        []SceneEffectV2FeatureBasicPut    `json:"effects_v2,omitempty"`        // 0-3
}

// Palette items

type SceneColorPalettePut struct {
	Color   SceneColorPut    `json:"color"`
	Dimming *SceneDimmingPut `json:"dimming,omitempty"`
}

type SceneDimmingFeatureBasicPut struct {
	Brightness float64 `json:"brightness"`
}

type SceneColorTemperaturePalettePut struct {
	ColorTemperature SceneColorTemperaturePut `json:"color_temperature"`
	Dimming          SceneDimmingPut          `json:"dimming"`
}

type SceneEffectFeatureBasicPut struct {
	Effect string `json:"effect"`
}

type SceneEffectV2FeatureBasicPut struct {
	Action SceneEffectV2ActionPut `json:"action"`
}

// ---------- SceneRecallPut ----------

type SceneRecallPut struct {
	Action   string           `json:"action"`            // active, dynamic_palette, static
	Duration int              `json:"duration"`          // ms
	Dimming  *SceneDimmingPut `json:"dimming,omitempty"` // overrides scene dimming
}

// ---------- SceneMetadataPut ----------

type SceneMetadataPut struct {
	Name    string  `json:"name"`
	AppData *string `json:"appdata,omitempty"` // optional 1-16 chars
}
