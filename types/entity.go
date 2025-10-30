package types

import "time"

type EntityType string

const (
	EntityTypeLight   EntityType = "light"
	EntityTypeScene   EntityType = "scene"
	EntityTypeSpeaker EntityType = "speaker"
	EntityTypeButton  EntityType = "button"
	EntityTypeUnknown EntityType = "unknown"
)

type Entity struct {
	ExternalID string // ExternalID used internally in integration. eg. hue bridge supplied UUID.
	DeviceID   string // ExternalID of the device that exposes the entity
	EntityID   string // human readable unique id (eg. light.living_room)
	Type       EntityType
	Name       string
	Enabled    bool
	Available  bool
	CreatedAt  time.Time
}
