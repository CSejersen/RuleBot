package events

import (
	"time"
)

func (s *SceneUpdate) GetType() string {
	return s.Type
}

type SceneStatus struct {
	Active     string    `json:"active"` // "inactive", "static", "dynamic_palette"
	LastRecall time.Time `json:"last_recall"`
}

type SceneUpdate struct {
	ID     string       `json:"id"`
	Type   string       `json:"type"` // always "scene"
	Status *SceneStatus `json:"status,omitempty"`
}

func (s *SceneUpdate) SafeActive() *string {
	if s.Status != nil {
		return &s.Status.Active
	}
	return nil
}
