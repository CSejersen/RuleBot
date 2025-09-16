package events

func (l *LightUpdate) GetType() string {
	return l.Type
}

type LightUpdate struct {
	ID                    string                       `json:"id"`
	IDV1                  string                       `json:"id_v1,omitempty"`
	Type                  string                       `json:"type"` // always "light"
	Metadata              *MetadataUpdate              `json:"metadata,omitempty"`
	Identify              *IdentifyUpdate              `json:"identify,omitempty"`
	On                    *OnUpdate                    `json:"on,omitempty"`
	Dimming               *DimmingUpdate               `json:"dimming,omitempty"`
	DimmingDelta          *DimmingDeltaUpdate          `json:"dimming_delta,omitempty"`
	ColorTemperature      *ColorTemperatureUpdate      `json:"color_temperature,omitempty"`
	ColorTemperatureDelta *ColorTemperatureDeltaUpdate `json:"color_temperature_delta,omitempty"`
	Color                 *ColorUpdate                 `json:"color,omitempty"`
	Dynamics              *DynamicsUpdate              `json:"dynamics,omitempty"`
	Alert                 *AlertUpdate                 `json:"alert,omitempty"`
	Signaling             *SignalingUpdate             `json:"signaling,omitempty"`
	Gradient              *GradientUpdate              `json:"gradient,omitempty"`
	EffectsV2             *EffectV2Update              `json:"effects_v2,omitempty"`
	TimedEffects          *TimedEffectsUpdate          `json:"timed_effects,omitempty"`
	PowerUp               *PowerUpUpdate               `json:"powerup,omitempty"`
	ContentConfiguration  *ContentConfigurationUpdate  `json:"content_configuration,omitempty"`
}

type IdentifyUpdate struct {
	Action   string `json:"action"`   // always "identify"
	Duration int    `json:"duration"` // ms
}

type PowerUpUpdate struct {
	Preset  string                `json:"preset"` // safety, powerfail, last_on_state, custom
	On      *PowerUpOnUpdate      `json:"on,omitempty"`
	Dimming *PowerUpDimmingUpdate `json:"dimming,omitempty"`
	Color   *PowerUpColorUpdate   `json:"color,omitempty"`
}

type PowerUpOnUpdate struct {
	Mode string    `json:"mode"` // on, toggle, previous
	On   *OnUpdate `json:"on,omitempty"`
}

type PowerUpDimmingUpdate struct {
	Mode    string         `json:"mode"` // dimming, previous
	Dimming *DimmingUpdate `json:"dimming,omitempty"`
}

type PowerUpColorUpdate struct {
	Mode             string                  `json:"mode"` // color_temperature, color, previous
	ColorTemperature *ColorTemperatureUpdate `json:"color_temperature,omitempty"`
	Color            *ColorUpdate            `json:"color,omitempty"`
}

type ContentConfigurationUpdate struct {
	Orientation *OrientationUpdate `json:"orientation,omitempty"`
	Order       *OrderUpdate       `json:"order,omitempty"`
}

type OrientationUpdate struct {
	Orientation string `json:"orientation"` // horizontal, vertical
}

type OrderUpdate struct {
	Order string `json:"order"` // forward, reversed
}

func (l *LightUpdate) ResolveStateChanges() []string {
	var changes []string

	if l.SafeBrightness() != nil {
		changes = append(changes, "brightness")
	}
	if l.SafeOn() != nil {
		changes = append(changes, "power_mode")
	}
	if l.SafeMirek() != nil {
		changes = append(changes, "mirek")
	}
	if l.SafeColorXY() != nil {
		changes = append(changes, "color_xy")
	}
	if l.SafeEffect() != nil {
		changes = append(changes, "effect")
	}
	if l.SafeAlert() != nil {
		changes = append(changes, "alert")
	}
	if l.SafeDynamicsSpeed() != nil {
		changes = append(changes, "dynamics_speed")
	}
	if l.SafeGradientMode() != nil {
		changes = append(changes, "gradient_mode")
	}

	return changes
}

// SafeOn returns On/off state if it exists
func (l *LightUpdate) SafeOn() *bool {
	if l.On != nil {
		return &l.On.On
	}
	return nil
}

// SafeBrightness returns Brightness (0–100) if it exists
func (l *LightUpdate) SafeBrightness() *float64 {
	if l.Dimming != nil {
		return &l.Dimming.Brightness
	}
	return nil
}

// SafeColorXY returns XY color coordinates if exists
func (l *LightUpdate) SafeColorXY() *XY {
	if l.Color != nil {
		return &l.Color.XY
	}
	return nil
}

// SafeMirek returns Color temperature in Mirek (153–500 is usual range) if exists
func (l *LightUpdate) SafeMirek() *int {
	if l.ColorTemperature != nil {
		return &l.ColorTemperature.Mirek
	}
	return nil
}

// SafeEffect returns Current active effect (prism, glisten, no_effect, etc.) if exists
func (l *LightUpdate) SafeEffect() *string {
	if l.EffectsV2 != nil {
		return &l.EffectsV2.Action.Effect
	}
	return nil
}

// SafeGradientMode returns Gradient mode (interpolated_palette, random_pixelated, etc.) if exists
func (l *LightUpdate) SafeGradientMode() *string {
	if l.Gradient != nil && l.Gradient.Mode != nil {
		return l.Gradient.Mode
	}
	return nil
}

// SafeDynamicsSpeed returns Dynamics speed (0.0–1.0) is it exists
func (l *LightUpdate) SafeDynamicsSpeed() *float64 {
	if l.Dynamics != nil && l.Dynamics.Speed != nil {
		return l.Dynamics.Speed
	}
	return nil
}

// SafeAlert return Alert action (breathe, etc.) is it exists
func (l *LightUpdate) SafeAlert() *string {
	if l.Alert != nil {
		return &l.Alert.Action
	}
	return nil
}
