package models

import (
	"time"

	"gorm.io/datatypes"
)

// IntegrationConfig represents a user-added integration in the database
type IntegrationConfig struct {
	ID              uint           `gorm:"primaryKey;autoIncrement"`
	IntegrationName string         `gorm:"size:255;not null"` // e.g., "hue"
	DisplayName     string         `gorm:"size:255"`          // optional user-friendly name
	UserConfig      datatypes.JSON `gorm:"type:json"`         // user-provided configuration
	Enabled         bool           `gorm:"default:true"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
