package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"home_automation_server/engine/integration"
	"home_automation_server/storage"
	"home_automation_server/storage/models"
	"home_automation_server/types"
	"time"
)

// AddIntegration enables an integration with the supplied userConfig. The integration is not loaded into the engine, for that call LoadIntegration.
func (e *Engine) AddIntegration(ctx context.Context, integrationName string, userConfig map[string]any) error {
	desc, ok := e.IntegrationDescRegistry.Available[integrationName]
	if !ok {
		return fmt.Errorf("integration %s not available", integrationName)
	}

	cfgJSON, _ := json.Marshal(userConfig)
	cfg := &models.IntegrationConfig{
		IntegrationName: integrationName,
		DisplayName:     desc.DisplayName,
		UserConfig:      cfgJSON,
		Enabled:         true,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	if err := e.IntegrationCfgStore.Save(ctx, cfg); err != nil {
		return fmt.Errorf("failed to save integration config: %w", err)
	}
	return nil
}

func (e *Engine) LoadIntegration(ctx context.Context, integrationName string) error {
	e.Logger.Debug("LOADING INTEGRATION", zap.String("integration_name", integrationName))
	storageCfg, err := e.IntegrationCfgStore.LoadByIntegrationName(ctx, integrationName)
	if err != nil {
		return fmt.Errorf("failed to load integration config: %w", err)
	}

	cfg, err := storage.IntegrationCfgFromStorage(*storageCfg)
	if err != nil {
		return fmt.Errorf("failed to convert integration config: %w", err)
	}

	desc, ok := e.IntegrationDescRegistry.Get(integrationName)
	if !ok {
		return fmt.Errorf("integration %s not available", integrationName)
	}

	integrationInstance, err := desc.CreateFunc(ctx, cfg.UserConfig, e.StateCache, e.EntityRegistry, e.Logger.Named(integrationName))
	if err != nil {
		return fmt.Errorf("failed to create integration instance: %w", err)
	}
	integrationInstance.ConfigID = cfg.ID
	integrationInstance.Descriptor = desc

	// Load existing devices & entities from DB
	devices, err := e.DeviceStore.GetDevicesByIntegration(ctx, storageCfg.ID)
	if err != nil {
		return fmt.Errorf("failed to load devices for integration %s: %w", integrationName, err)
	}

	deviceIDs := make([]string, 0, len(devices))
	for _, d := range devices {
		deviceIDs = append(deviceIDs, d.ID)
	}

	allEntities, err := e.EntityStore.GetEntitiesByDeviceIDs(ctx, deviceIDs)
	if err != nil {
		return fmt.Errorf("failed to load entities for integration %s: %w", integrationName, err)
	}

	entityMap := make(map[string][]models.Entity)
	for _, ent := range allEntities {
		entityMap[ent.DeviceID] = append(entityMap[ent.DeviceID], ent)
	}

	// todo: add devices and integration to stateStore

	// Register services
	for serviceName, data := range integrationInstance.Services {
		e.RegisterService(integrationInstance.Descriptor.Name, serviceName, data)
	}

	// Start event pipeline
	go func(name string, i *integration.Instance) {
		p := e.constructEventPipeline(name, e.StateCache, i)
		if err := p.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			e.Logger.Error("event pipeline exited with error", zap.Error(err))
		}
	}(integrationName, &integrationInstance)

	e.Integrations[integrationName] = integrationInstance
	e.Logger.Info("integration loaded", zap.String("display_name", integrationInstance.Descriptor.DisplayName))
	return nil
}

func (e *Engine) DiscoverDevicesForIntegration(ctx context.Context, integrationName string) error {
	integration, ok := e.Integrations[integrationName]
	if !ok {
		return fmt.Errorf("integration %s not active", integrationName)
	}

	// Add a timeout for discovery
	discoveryCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	devices, entities, err := integration.Discovery.Discover(discoveryCtx)
	if err != nil {
		return fmt.Errorf("discovery failed: %w", err)
	}

	discoveredDeviceIDs := make(map[string]struct{})
	discoveredEntityIDs := make(map[string]struct{})

	// Add/update devices
	for _, d := range devices {
		d.Available = true
		d.IntegrationID = integration.ConfigID
		discoveredDeviceIDs[d.ID] = struct{}{}
		storageDevice, err := storage.DeviceToStorage(d)
		if err != nil {
			e.Logger.Error("failed to convert device", zap.Error(err))
			continue
		}
		if err := e.addOrUpdateDevicePreserveEnabled(ctx, &storageDevice); err != nil {
			return fmt.Errorf("failed to save device %s: %w", d.Name, err)
		}
	}

	// Add/update entities
	for _, ent := range entities {
		ent.Available = true
		discoveredEntityIDs[ent.ExternalID] = struct{}{}
		if err := e.addOrUpdateEntityPreserveEnabled(ctx, &ent); err != nil {
			return fmt.Errorf("failed to save entity %s: %w", ent.Name, err)
		}
	}

	// Mark unavailable devices/entities
	if err := e.markUnavailable(ctx, integrationName, discoveredDeviceIDs, discoveredEntityIDs); err != nil {
		e.Logger.Warn("failed to mark unavailable devices/entities", zap.Error(err))
	}

	if err := e.RefreshEntityRegistry(ctx); err != nil {
		e.Logger.Error("failed to refresh entity registry", zap.Error(err))
	}
	e.Logger.Info("successfully ran discovery for integration", zap.String("integration_name", integrationName))

	return nil
}

// addOrUpdateDevicePreserveEnabled preserves enabled value
func (e *Engine) addOrUpdateDevicePreserveEnabled(ctx context.Context, d *models.Device) error {
	existing, err := e.DeviceStore.GetDeviceByID(ctx, d.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// New device
			if err := e.DeviceStore.AddDevice(ctx, d); err != nil {
				e.Logger.Error("failed to add new device", zap.Error(err))
			}
		} else {
			e.Logger.Error("failed to query device", zap.Error(err))
		}
	} else {
		// Existing device: preserve enacbled flag
		d.Enabled = existing.Enabled
		d.CreatedAt = existing.CreatedAt
		if err := e.DeviceStore.UpdateDevice(ctx, d); err != nil {
			e.Logger.Error("failed to update device", zap.Error(err))
		}
	}
	return nil
}

// addOrUpdateEntityPreserveEnabled preserves enabled value
func (e *Engine) addOrUpdateEntityPreserveEnabled(ctx context.Context, entity *types.Entity) error {
	storageEntity, err := storage.EntityToStorage(*entity)
	if err != nil {
		return fmt.Errorf("failed to convert entity to storage: %w", err)
	}
	existing, err := e.EntityStore.GetEntityByID(ctx, storageEntity.ExternalID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			e.Logger.Info("adding new entity", zap.String("entity", entity.Name))
			// New device
			if err := e.EntityStore.AddEntity(ctx, &storageEntity); err != nil {
				e.Logger.Error("failed to add new entity", zap.Error(err))
			}
		} else {
			e.Logger.Error("failed to query entity", zap.Error(err))
		}
	} else {
		// Existing entity: preserve enabled flag
		entity.Enabled = existing.Enabled
		entity.CreatedAt = existing.CreatedAt
		if err := e.EntityStore.UpdateEntity(ctx, &storageEntity); err != nil {
			e.Logger.Error("failed to update device", zap.Error(err))
		}
	}

	return nil
}

func (e *Engine) markUnavailable(ctx context.Context, integrationName string, discoveredDevices map[string]struct{}, discoveredEntities map[string]struct{}) error {
	integration, ok := e.Integrations[integrationName]
	if !ok {
		return fmt.Errorf("integration %s not active", integrationName)
	}

	// Load all devices for the integration
	allDevices, err := e.DeviceStore.GetDevicesByIntegration(ctx, integration.ConfigID)
	if err != nil {
		return fmt.Errorf("failed to load devices for unavailable check: %w", err)
	}

	if len(allDevices) == 0 {
		return nil
	}

	// Load all entities
	deviceIDs := make([]string, 0, len(allDevices))
	for _, d := range allDevices {
		deviceIDs = append(deviceIDs, d.ID)
	}

	allEntities, err := e.EntityStore.GetEntitiesByDeviceIDs(ctx, deviceIDs)
	if err != nil {
		return fmt.Errorf("failed to load entities for unavailable check: %w", err)
	}

	// Map entities to their devices
	entityMap := make(map[string][]models.Entity)
	for _, ent := range allEntities {
		entityMap[ent.DeviceID] = append(entityMap[ent.DeviceID], ent)
	}

	// Mark devices and entities unavailable if they weren't discovered
	for _, d := range allDevices {
		if _, found := discoveredDevices[d.ID]; !found && d.Available {
			d.Available = false
			if err := e.DeviceStore.UpdateDevice(ctx, d); err != nil {
				e.Logger.Warn("failed to mark device unavailable", zap.Error(err), zap.String("device", d.Name))
			}
		}

		for _, entity := range entityMap[d.ID] {
			if _, found := discoveredEntities[entity.ExternalID]; !found && entity.Available {
				entity.Available = false
				if err := e.EntityStore.UpdateEntity(ctx, &entity); err != nil {
					e.Logger.Warn("failed to mark entity unavailable", zap.Error(err), zap.String("device", d.Name), zap.String("entity", entity.Name))
				}
			}
		}
	}

	return nil
}

func (e *Engine) markUnavailableEntities(ctx context.Context, integrationName string, discovered map[string]struct{}) {
	allDevices, err := e.DeviceStore.GetDevicesByIntegration(ctx, e.Integrations[integrationName].ConfigID)
	if err != nil {
		e.Logger.Warn("failed to load devices for unavailable check", zap.Error(err))
	}

	for _, d := range allDevices {
		if _, ok := discovered[d.ID]; !ok && d.Available {
			d.Available = false
			if err := e.DeviceStore.UpdateDevice(ctx, d); err != nil {
				e.Logger.Warn("failed to mark device unavailable", zap.Error(err), zap.String("device", d.Name))
			}
		}
	}
}
