package storage

import (
	"context"
	"errors"
	"home_automation_server/storage/models"
	"time"

	"gorm.io/gorm"
)

// EntityStore defines methods to manage entities
type EntityStore interface {
	AddEntity(ctx context.Context, e *models.Entity) error
	UpdateEntity(ctx context.Context, e *models.Entity) error
	GetEntityByID(ctx context.Context, id string) (*models.Entity, error)
	GetAllEntities(ctx context.Context) ([]models.Entity, error)
	GetEntitiesByDevice(ctx context.Context, deviceID string) ([]models.Entity, error)
	GetEntitiesByDeviceIDs(ctx context.Context, deviceIDs []string) ([]models.Entity, error)
	DeleteEntity(ctx context.Context, id uint) error
}

// GormEntityStore implements EntityStore using GORM
type GormEntityStore struct {
	db *gorm.DB
}

// NewGormEntityStore creates a new GormEntityStore
func NewGormEntityStore(db *gorm.DB) *GormEntityStore {
	return &GormEntityStore{db: db}
}

// AddEntity inserts a new entity
func (s *GormEntityStore) AddEntity(ctx context.Context, e *models.Entity) error {
	e.CreatedAt = time.Now()
	e.UpdatedAt = time.Now()
	return s.db.WithContext(ctx).Create(e).Error
}

// UpdateEntity updates an existing entity
func (s *GormEntityStore) UpdateEntity(ctx context.Context, e *models.Entity) error {
	e.UpdatedAt = time.Now()
	return s.db.WithContext(ctx).Save(e).Error
}

// GetEntityByID fetches an entity by primary key
func (s *GormEntityStore) GetEntityByID(ctx context.Context, id string) (*models.Entity, error) {
	var e models.Entity
	err := s.db.WithContext(ctx).Where("external_id = ?", id).First(&e).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &e, nil
}

// GetEntitiesByDeviceIDs fetches entities for a list of device IDs
func (s *GormEntityStore) GetEntitiesByDeviceIDs(ctx context.Context, deviceIDs []string) ([]models.Entity, error) {
	if len(deviceIDs) == 0 {
		return nil, nil // nothing to fetch
	}

	var entities []models.Entity
	if err := s.db.WithContext(ctx).Where("device_id IN ?", deviceIDs).Find(&entities).Error; err != nil {
		return nil, err
	}

	return entities, nil
}

// GetEntityByEntityID fetches an entity by its unique EntityID
func (s *GormEntityStore) GetEntityByEntityID(ctx context.Context, entityID string) (*models.Entity, error) {
	var e models.Entity
	if err := s.db.WithContext(ctx).Where("entity_id = ?", entityID).First(&e).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &e, nil
}

// GetEntitiesByDevice fetches all entities for a specific device
func (s *GormEntityStore) GetEntitiesByDevice(ctx context.Context, deviceID string) ([]models.Entity, error) {
	var entities []models.Entity
	if err := s.db.WithContext(ctx).Where("device_id = ?", deviceID).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

// DeleteEntity removes an entity by ExternalID
func (s *GormEntityStore) DeleteEntity(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&models.Entity{}, id).Error
}

// GetAllEntities fetches all entities in the database
func (s *GormEntityStore) GetAllEntities(ctx context.Context) ([]models.Entity, error) {
	var entities []models.Entity
	if err := s.db.WithContext(ctx).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}
