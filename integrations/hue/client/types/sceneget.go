package types

func (s *SceneGet) GetType() string {
	return s.Type
}

func (s *SceneGet) GetID() string {
	return s.ID
}

// SceneGet is the top-level representation of a scene
type SceneGet struct {
	ID          string             `json:"id"`
	IDV1        *string            `json:"id_v1,omitempty"`
	Actions     []ActionGet        `json:"actions"`
	Palette     *ScenePaletteGet   `json:"palette,omitempty"`
	Recall      SceneRecallGet     `json:"recall"`
	Type        string             `json:"type"` // always "scene"
	Metadata    SceneMetadataGet   `json:"metadata"`
	Group       ResourceIdentifier `json:"group"`
	Speed       float64            `json:"speed"`
	AutoDynamic bool               `json:"auto_dynamic"`
	Status      SceneStatusGet     `json:"status"`
}

// ActionGet represents an action in the scene
type ActionGet struct {
	Target ResourceIdentifier `json:"target"`
	Action ActionDetailGet    `json:"action"`
}

type ActionDetailGet struct {
	On               *OnGet                     `json:"on,omitempty"`
	Dimming          *DimmingGet                `json:"dimming,omitempty"`
	Color            *ColorBasicGet             `json:"color,omitempty"`
	ColorTemperature *ColorTemperatureSimpleGet `json:"color_temperature,omitempty"`
	Gradient         *GradientGet               `json:"gradient,omitempty"`
	EffectsV2        *EffectsV2Get              `json:"effects_v2,omitempty"`
	Dynamics         *ActionDynamicsGet         `json:"dynamics,omitempty"`
}

type ActionDynamicsGet struct {
	Duration *int `json:"duration,omitempty"` // transition duration in ms
}

type ScenePaletteGet struct {
	Color            []ColorPaletteGet            `json:"color"`
	Dimming          []DimmingFeatureBasicGet     `json:"dimming"`
	ColorTemperature []ColorTemperaturePaletteGet `json:"color_temperature"`
	Effects          []EffectFeatureBasicGet      `json:"effects"`
	EffectsV2        []EffectV2FeatureBasicGet    `json:"effects_v2"`
}

type ColorPaletteGet struct {
	Color   ColorBasicGet `json:"color"`
	Dimming DimmingGet    `json:"dimming"`
}

type DimmingFeatureBasicGet struct {
	Brightness float64 `json:"brightness"`
}

type ColorTemperaturePaletteGet struct {
	ColorTemperature ColorTemperatureSimpleGet `json:"color_temperature"`
	Dimming          DimmingGet                `json:"dimming"`
}

type EffectFeatureBasicGet struct {
	Effect string `json:"effect"`
}

type EffectV2FeatureBasicGet struct {
	Action EffectV2ActionFullGet `json:"action"`
}

type EffectV2ActionFullGet struct {
	Effect     string               `json:"effect"`
	Parameters *EffectParametersGet `json:"parameters,omitempty"`
}

type SceneMetadataGet struct {
	Name    string              `json:"name"`
	Image   *ResourceIdentifier `json:"image,omitempty"`
	AppData *string             `json:"appdata,omitempty"`
}

type SceneRecallGet struct {
}

type SceneStatusGet struct {
	Active     string  `json:"active"`                // inactive | static | dynamic_palette
	LastRecall *string `json:"last_recall,omitempty"` // datetime string
}
