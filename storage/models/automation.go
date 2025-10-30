package models

import (
	"gorm.io/datatypes"
	"time"
)

type Automation struct {
	ID            uint           `gorm:"primaryKey;autoIncrement"`
	Alias         string         `gorm:"size:255;uniqueIndex;not null"`
	Description   string         `gorm:"size:255;not null"`
	Triggers      datatypes.JSON `gorm:"type:json;not null"`
	Conditions    datatypes.JSON `gorm:"type:json"`
	Actions       datatypes.JSON `gorm:"type:json"`
	Enabled       bool           `gorm:"default:true"`
	LastTriggered *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
