package integration

import "time"

type IntegrationConfig struct {
	ID              uint
	IntegrationName string
	DisplayName     string
	UserConfig      map[string]any
	Enabled         bool
	CreatedAt       time.Time
}
