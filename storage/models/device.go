package models

import (
	"gorm.io/datatypes"
	"time"
)

type Device struct {
	ID            string         `gorm:"primaryKey;size:191" json:"id"`
	IntegrationID uint           `gorm:"not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"integration_id"`
	Type          string         `gorm:"size:255;not null" json:"type"` // e.g., "light", "sensor"
	Name          string         `gorm:"size:255" json:"name"`          // human-readable name
	Metadata      datatypes.JSON `gorm:"type:json" json:"metadata"`     // extra device info
	Enabled       bool           `gorm:"default:false" json:"enabled"`
	Available     bool           `gorm:"default:true" json:"available"`

	Entities  []Entity  `gorm:"foreignKey:DeviceID" json:"entities,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
