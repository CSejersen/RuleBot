package models

import (
	"time"
)

type Entity struct {
	ExternalID string `gorm:"primaryKey;size:191"`
	DeviceID   string `gorm:"size:191;not null;index"` // the ID of the device that exposes the entity
	EntityID   string `gorm:"size:255;not null;index"` // human readable unique_id e.g. "light.living_room"
	Type       string `gorm:"size:50;index;not null"`  // e.g. "light", "sensor", "switch"
	Name       string `gorm:"size:100;not null"`
	Enabled    bool   `gorm:"default:false"`
	Available  bool   `gorm:"default:true"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
