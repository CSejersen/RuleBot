package types

import (
	"time"
)

type DeviceType string

const (
	DeviceTypeLight        DeviceType = "light"
	DeviceTypeGroupedLight DeviceType = "grouped_light"
	DeviceTypeRemote       DeviceType = "remote"
)

type Device struct {
	ID            string
	IntegrationID uint
	Type          DeviceType
	Name          string
	Metadata      map[string]any
	Enabled       bool
	Available     bool
	CreatedAt     time.Time
}
