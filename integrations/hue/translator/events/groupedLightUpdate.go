package events

func (g *GroupedLightUpdate) GetType() string {
	return g.Type
}

type GroupedLightUpdate struct {
	ID                    string                       `json:"id"`
	Type                  string                       `json:"type"` // always "grouped_light"
	Metadata              *MetadataUpdate              `json:"metadata,omitempty"`
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
}

func (l *GroupedLightUpdate) ResolveStateChanges() []string {
	changes := []string{}

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

// On/off state
func (g *GroupedLightUpdate) SafeOn() *bool {
	if g.On != nil {
		return &g.On.On
	}
	return nil
}

// Brightness (0–100)
func (g *GroupedLightUpdate) SafeBrightness() *float64 {
	if g.Dimming != nil {
		return &g.Dimming.Brightness
	}
	return nil
}

// XY color coordinates
func (g *GroupedLightUpdate) SafeColorXY() *XY {
	if g.Color != nil {
		return &g.Color.XY
	}
	return nil
}

// Color temperature in Mirek
func (g *GroupedLightUpdate) SafeMirek() *int {
	if g.ColorTemperature != nil {
		return &g.ColorTemperature.Mirek
	}
	return nil
}

// Current active effect
func (g *GroupedLightUpdate) SafeEffect() *string {
	if g.EffectsV2 != nil {
		return &g.EffectsV2.Action.Effect
	}
	return nil
}

// Gradient mode
func (g *GroupedLightUpdate) SafeGradientMode() *string {
	if g.Gradient != nil && g.Gradient.Mode != nil {
		return g.Gradient.Mode
	}
	return nil
}

// Dynamics speed
func (g *GroupedLightUpdate) SafeDynamicsSpeed() *float64 {
	if g.Dynamics != nil && g.Dynamics.Speed != nil {
		return g.Dynamics.Speed
	}
	return nil
}

// Alert action
func (g *GroupedLightUpdate) SafeAlert() *string {
	if g.Alert != nil {
		return &g.Alert.Action
	}
	return nil
}
