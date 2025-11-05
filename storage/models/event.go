package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// EventType mirrors the domain enum
type EventType string

const (
	EventTypeStateChanged EventType = "state_changed"
	EventTypeCallService  EventType = "call_service"
	EventTypeTimeChanged  EventType = "time_changed"
)

// Event represents a persisted event in the database
type Event struct {
	ID        string `gorm:"primaryKey;type:char(36)"` // UUID from Context.ID
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Type      EventType      `gorm:"type:varchar(32);index;not null"`
	Data      datatypes.JSON `gorm:"type:json;not null"` // any event payload
	ContextID string         `gorm:"type:char(36);index"`
	TimeFired time.Time      `gorm:"index;not null"`
	Context   *Context       `gorm:"foreignKey:ContextID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

// Context stores metadata about the event's origin
type Context struct {
	ID        string `gorm:"primaryKey;type:char(36)"` // UUID
	ParentID  string `gorm:"type:char(36);index"`
	Origin    string `gorm:"type:varchar(64)"` // e.g., "service", "external", "automation", "hue", "halo"
	CreatedAt time.Time
}
