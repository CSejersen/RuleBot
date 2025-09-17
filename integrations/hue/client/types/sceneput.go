package types

// ScenePut is the payload for updating a scene
type ScenePut struct {
	Actions     []ActionPut      `json:"actions,omitempty"`
	Palette     *ScenePalettePut `json:"palette,omitempty"`
	Recall      *SceneRecallPut  `json:"recall,omitempty"`
	Type        *string          `json:"type,omitempty"`
	Metadata    *MetadataPut     `json:"metadata,omitempty"`
	Speed       *float64         `json:"speed,omitempty"`
	AutoDynamic *bool            `json:"auto_dynamic,omitempty"`
}

// ActionPut updates/defines an action for a target inside a scene
type ActionPut struct {
	Target ResourceIdentifier `json:"target"`
	Action ActionDetailPut    `json:"action"`
}

type ActionDetailPut struct {
	On               *OnPut               `json:"on,omitempty"`
	Dimming          *DimmingPut          `json:"dimming,omitempty"`
	Color            *ColorPut            `json:"color,omitempty"`
	ColorTemperature *ColorTemperaturePut `json:"color_temperature,omitempty"`
	Gradient         *GradientPut         `json:"gradient,omitempty"`
	EffectsV2        *EffectV2Put         `json:"effects_v2,omitempty"`
	Dynamics         *DynamicsPut         `json:"dynamics,omitempty"`
}

type ScenePalettePut struct {
	Color            []ColorPalettePut            `json:"color,omitempty"`
	Dimming          []DimmingFeatureBasicPut     `json:"dimming,omitempty"`
	ColorTemperature []ColorTemperaturePalettePut `json:"color_temperature,omitempty"`
	Effects          []EffectFeatureBasicPut      `json:"effects,omitempty"` // deprecated
	EffectsV2        []EffectV2FeatureBasicPut    `json:"effects_v2,omitempty"`
}

type ColorPalettePut struct {
	Color   ColorPut   `json:"color"`
	Dimming DimmingPut `json:"dimming"`
}

type DimmingFeatureBasicPut struct {
	Brightness float64 `json:"brightness"`
}

type ColorTemperaturePalettePut struct {
	ColorTemperature ColorTemperaturePut `json:"color_temperature"`
	Dimming          DimmingPut          `json:"dimming"`
}

type EffectFeatureBasicPut struct {
	Effect string `json:"effect"`
}

type EffectV2FeatureBasicPut struct {
	Action EffectV2ActionPut `json:"action"`
}

type SceneRecallPut struct {
	Action   *string     `json:"action,omitempty"`   // active | dynamic_palette | static
	Duration *int        `json:"duration,omitempty"` // ms
	Dimming  *DimmingPut `json:"dimming,omitempty"`
}
