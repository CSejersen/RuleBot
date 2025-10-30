package integration

import (
	"context"
	"go.uber.org/zap"
	"home_automation_server/types"
)

type ConfigFieldType string

const (
	ConfigFieldTypeText ConfigFieldType = "text"
)

// IntegrationDescriptor defines metadata about an integration without initializing it.
type IntegrationDescriptor struct {
	Name         string                 `json:"name" yaml:"name"`
	DisplayName  string                 `json:"display_name" yaml:"display_name"`
	Description  string                 `json:"description" yaml:"description"`
	Version      string                 `json:"version" yaml:"version"`
	Capabilities []string               `json:"capabilities" yaml:"capabilities"`
	ConfigSchema map[string]ConfigField `json:"config_schema" yaml:"config_schema"`
	CreateFunc   IntegrationFactoryFunc `json:"-" yaml:"-"`
}

// IntegrationFactoryFunc creates an initialized IntegrationInstance from a config.
type IntegrationFactoryFunc func(context.Context, map[string]any, types.StateStore, types.EntityRegistry, *zap.Logger) (Instance, error)

type ConfigField struct {
	Label       string          `json:"label" yaml:"label"`
	Description string          `json:"description,omitempty" yaml:"description,omitempty"`
	Type        ConfigFieldType `json:"type" yaml:"type"`
	Required    bool            `json:"required" yaml:"required"`
	Placeholder string          `json:"placeholder,omitempty" yaml:"placeholder,omitempty"`
	Default     any             `json:"default,omitempty" yaml:"default,omitempty"`
	Options     []string        `json:"options,omitempty" yaml:"options,omitempty"`
}
