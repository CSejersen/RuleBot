package models

import (
	"time"

	"gorm.io/datatypes"
)

// State represents a persisted entity state in the database
type State struct {
	EntityID    string         `gorm:"primaryKey;size:255;not null;index"` // entity_id e.g. "light.living_room"
	State       datatypes.JSON `gorm:"type:json"`                          // main state value
	Attributes  datatypes.JSON `gorm:"type:json"`                          // state attributes map
	LastChanged time.Time      `gorm:"not null;index"`
	LastUpdated time.Time      `gorm:"not null;index"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
