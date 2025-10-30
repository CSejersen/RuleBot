package storage

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	models2 "home_automation_server/storage/models"
)

type IntegrationCfgStore interface {
	Save(ctx context.Context, cfg *models2.IntegrationConfig) error
	LoadAll(ctx context.Context) ([]*models2.IntegrationConfig, error)
	LoadByID(ctx context.Context, id uint) (*models2.IntegrationConfig, error)
	LoadByIntegrationName(ctx context.Context, name string) (*models2.IntegrationConfig, error)
	Delete(ctx context.Context, id uint) error
	AutoMigrate() error
}

type GormIntegrationCfgStore struct {
	db *gorm.DB
}

func NewGormIntegrationCfgStore(db *gorm.DB) *GormIntegrationCfgStore {
	return &GormIntegrationCfgStore{db: db}
}

// AutoMigrate ensures the tables exist
func (s *GormIntegrationCfgStore) AutoMigrate() error {
	return s.db.AutoMigrate(&models2.IntegrationConfig{}, &models2.Device{})
}

// Save inserts or updates an integration
func (s *GormIntegrationCfgStore) Save(ctx context.Context, cfg *models2.IntegrationConfig) error {
	return s.db.WithContext(ctx).Save(cfg).Error
}

// LoadAll fetches all integration with their devices preloaded
func (s *GormIntegrationCfgStore) LoadAll(ctx context.Context) ([]*models2.IntegrationConfig, error) {
	var configs []*models2.IntegrationConfig
	if err := s.db.WithContext(ctx).
		Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("failed to load integration: %w", err)
	}
	return configs, nil
}

// LoadByID fetches a single integration by ExternalID
func (s *GormIntegrationCfgStore) LoadByID(ctx context.Context, id uint) (*models2.IntegrationConfig, error) {
	var cfg models2.IntegrationConfig
	if err := s.db.WithContext(ctx).
		First(&cfg, id).Error; err != nil {
		return nil, fmt.Errorf("integration not found: %w", err)
	}
	return &cfg, nil
}

// LoadByIntegrationName fetches a single integration by name
func (s *GormIntegrationCfgStore) LoadByIntegrationName(ctx context.Context, name string) (*models2.IntegrationConfig, error) {
	var cfg models2.IntegrationConfig
	if err := s.db.WithContext(ctx).
		Where("integration_name = ?", name).
		First(&cfg).Error; err != nil {
		return nil, fmt.Errorf("integration not found: %w", err)
	}
	return &cfg, nil
}

// Delete removes an integration (devices are deleted via OnDelete:CASCADE)
func (s *GormIntegrationCfgStore) Delete(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&models2.IntegrationConfig{}, id).Error
}
