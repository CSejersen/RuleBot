package events

import (
	"home_automation_server/integrations/hue/client/types"
)

func (s *SceneUpdate) GetType() string {
	return s.Type
}

type SceneUpdate struct {
	ID       string                `json:"id"`
	Type     string                `json:"type"` // always "scene"
	Metadata *types.MetadataPut    `json:"metadata,omitempty"`
	Recall   *types.SceneRecallPut `json:"recall,omitempty"`
	Dimming  *types.DimmingPut     `json:"dimming,omitempty"`
	Speed    *float64              `json:"speed,omitempty"`
	// add more fields later
}

// SafeAction returns the recall action if it exists ("active", "dynamic_palette", "static").
func (s *SceneUpdate) SafeAction() *string {
	if s.Recall != nil && s.Recall.Action != nil {
		return s.Recall.Action
	}
	return nil
}

// SafeRecallDuration returns the transition duration in ms if it exists.
func (s *SceneUpdate) SafeRecallDuration() *int {
	if s.Recall != nil && s.Recall.Duration != nil {
		return s.Recall.Duration
	}
	return nil
}

// SafeRecallBrightness returns the recall dimming brightness (0–100) if it exists.
func (s *SceneUpdate) SafeRecallBrightness() *float64 {
	if s.Recall != nil && s.Recall.Dimming != nil {
		return &s.Recall.Dimming.Brightness
	}
	return nil
}

// SafeBrightness returns the scene-level brightness (0–100) if it exists.
// Note: this is distinct from recall overrides.
func (s *SceneUpdate) SafeBrightness() *float64 {
	if s.Dimming != nil {
		return &s.Dimming.Brightness
	}
	return nil
}

// SafeSpeed returns the dynamic palette speed (0–1) if it exists.
func (s *SceneUpdate) SafeSpeed() *float64 {
	if s.Speed != nil {
		return s.Speed
	}
	return nil
}

// SafeName returns the scene name if it exists.
func (s *SceneUpdate) SafeName() *string {
	if s.Metadata != nil && s.Metadata.Name != nil {
		return s.Metadata.Name
	}
	return nil
}
