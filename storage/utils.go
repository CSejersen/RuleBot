package storage

import (
	"encoding/json"
	"fmt"
	"home_automation_server/automation"
	"home_automation_server/engine/integration"
	"home_automation_server/storage/models"
	"home_automation_server/types"
)

func IntegrationCfgToStorage(icfg integration.IntegrationConfig) (models.IntegrationConfig, error) {
	userConfigJSON, err := json.Marshal(icfg.UserConfig)
	if err != nil {
		return models.IntegrationConfig{}, fmt.Errorf("Error marshalling user config: %v", err)
	}

	return models.IntegrationConfig{
		ID:              icfg.ID,
		IntegrationName: icfg.IntegrationName,
		DisplayName:     icfg.DisplayName,
		UserConfig:      userConfigJSON,
		Enabled:         icfg.Enabled,
		CreatedAt:       icfg.CreatedAt,
	}, nil
}

func IntegrationCfgFromStorage(icfg models.IntegrationConfig) (integration.IntegrationConfig, error) {
	userConfig := map[string]any{}
	err := json.Unmarshal(icfg.UserConfig, &userConfig)
	if err != nil {
		return integration.IntegrationConfig{}, fmt.Errorf("failed unmarshalling config: %w", err)
	}

	return integration.IntegrationConfig{
		ID:              icfg.ID,
		IntegrationName: icfg.IntegrationName,
		DisplayName:     icfg.DisplayName,
		UserConfig:      userConfig,
		Enabled:         icfg.Enabled,
		CreatedAt:       icfg.CreatedAt,
	}, nil
}

func EntityFromStorage(e models.Entity) (types.Entity, error) {
	return types.Entity{
		ExternalID: e.ExternalID,
		DeviceID:   e.DeviceID,
		EntityID:   e.EntityID,
		Type:       types.EntityType(e.Type),
		Name:       e.Name,
		Enabled:    e.Enabled,
		Available:  e.Available,
		CreatedAt:  e.CreatedAt,
	}, nil
}

func DeviceFromStorage(d models.Device) (types.Device, error) {
	metadata := map[string]any{}
	err := json.Unmarshal(d.Metadata, &metadata)
	if err != nil {
		return types.Device{}, fmt.Errorf("failed unmarshalling metadata: %w", err)
	}

	return types.Device{
		ID:            d.ID,
		IntegrationID: d.IntegrationID,
		Type:          types.DeviceType(d.Type),
		Name:          d.Name,
		Metadata:      metadata,
		Enabled:       d.Enabled,
		CreatedAt:     d.CreatedAt,
	}, nil
}

func EntityToStorage(e types.Entity) (models.Entity, error) {
	return models.Entity{
		ExternalID: e.ExternalID,
		DeviceID:   e.DeviceID,
		EntityID:   e.EntityID,
		Type:       string(e.Type),
		Name:       e.Name,
		Enabled:    e.Enabled,
		Available:  e.Available,
		CreatedAt:  e.CreatedAt,
	}, nil
}

func DeviceToStorage(d types.Device) (models.Device, error) {
	metadataJSON, err := json.Marshal(d.Metadata)
	if err != nil {
		return models.Device{}, fmt.Errorf("failed marshalling device metadata: %v", err)
	}
	return models.Device{
		ID:            d.ID,
		IntegrationID: d.IntegrationID,
		Type:          string(d.Type),
		Name:          d.Name,
		Metadata:      metadataJSON,
		Enabled:       d.Enabled,
		Available:     d.Available,
		CreatedAt:     d.CreatedAt,
	}, nil
}

func AutomationFromStorage(m models.Automation) (automation.Automation, error) {
	var triggers []automation.BaseTrigger
	var conditions []automation.Condition
	var actions []automation.Action

	if err := json.Unmarshal(m.Triggers, &triggers); err != nil {
		return automation.Automation{}, fmt.Errorf("failed unmarshalling triggers: %w", err)
	}
	if len(m.Conditions) > 0 {
		if err := json.Unmarshal(m.Conditions, &conditions); err != nil {
			return automation.Automation{}, fmt.Errorf("failed unmarshalling conditions: %w", err)
		}
	}
	if len(m.Actions) > 0 {
		if err := json.Unmarshal(m.Actions, &actions); err != nil {
			return automation.Automation{}, fmt.Errorf("failed unmarshalling actions: %w", err)
		}
	}

	return automation.Automation{
		Id:          m.ID,
		Alias:       m.Alias,
		Description: m.Description,
		Trigger:     triggers,
		Condition:   conditions,
		Actions:     actions,
		Enabled:     m.Enabled,
	}, nil
}

func AutomationToStorage(a automation.Automation) (models.Automation, error) {
	triggersJSON, err := json.Marshal(a.Trigger)
	if err != nil {
		return models.Automation{}, fmt.Errorf("failed marshalling triggers: %w", err)
	}
	conditionsJSON, err := json.Marshal(a.Condition)
	if err != nil {
		return models.Automation{}, fmt.Errorf("failed marshalling conditions: %w", err)
	}
	actionsJSON, err := json.Marshal(a.Actions)
	if err != nil {
		return models.Automation{}, fmt.Errorf("failed marshalling actions: %w", err)
	}

	return models.Automation{
		ID:            a.Id,
		Alias:         a.Alias,
		Description:   a.Description,
		Triggers:      triggersJSON,
		Conditions:    conditionsJSON,
		Actions:       actionsJSON,
		Enabled:       a.Enabled,
		LastTriggered: nil,
	}, nil
}

func EventToStorage(e types.Event) (models.Event, error) {
	var dataBytes []byte
	var err error

	if e.Data != nil {
		dataBytes, err = json.Marshal(e.Data)
		if err != nil {
			return models.Event{}, err
		}
	}

	var ctxModel *models.Context
	if e.Context != nil {
		ctxModel = &models.Context{
			ID:       e.Context.ID,
			ParentID: e.Context.ParentID,
			// CreatedAt is autofilled by GORM on save
		}
	}

	eventModel := models.Event{
		Type:      models.EventType(e.Type),
		Data:      dataBytes,
		ContextID: "", // set below if context is present
		TimeFired: e.TimeFired,
		Context:   ctxModel,
	}

	if ctxModel != nil {
		eventModel.ContextID = ctxModel.ID
	}

	return eventModel, nil
}
