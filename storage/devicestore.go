package storage

import (
	"context"
	"gorm.io/gorm"
	"home_automation_server/storage/models"
	"time"
)

type DeviceStore interface {
	AddDevice(ctx context.Context, d *models.Device) error
	UpdateDevice(ctx context.Context, d *models.Device) error
	GetDeviceByID(ctx context.Context, id string) (*models.Device, error)
	GetAllDevices(ctx context.Context) ([]*models.Device, error)
	GetDevicesByIntegration(ctx context.Context, integrationID uint) ([]*models.Device, error)
	DeleteDevice(ctx context.Context, id uint) error
}

// GormDeviceStore implements DeviceStore using GORM
type GormDeviceStore struct {
	db *gorm.DB
}

// NewGormDeviceStore creates a new GormDeviceStore
func NewGormDeviceStore(db *gorm.DB) *GormDeviceStore {
	return &GormDeviceStore{db: db}
}

func (s *GormDeviceStore) GetAllDevices(ctx context.Context) ([]*models.Device, error) {
	var devices []*models.Device
	err := s.db.Find(&devices).Error
	if err != nil {
		return nil, err
	}
	return devices, nil
}

// AddDevice inserts a new device
func (s *GormDeviceStore) AddDevice(ctx context.Context, d *models.Device) error {
	d.CreatedAt = time.Now()
	d.UpdatedAt = time.Now()
	return s.db.WithContext(ctx).Create(d).Error
}

// UpdateDevice updates an existing device
func (s *GormDeviceStore) UpdateDevice(ctx context.Context, d *models.Device) error {
	d.UpdatedAt = time.Now()
	return s.db.WithContext(ctx).Save(d).Error
}

func (s *GormDeviceStore) GetDeviceByID(ctx context.Context, id string) (*models.Device, error) {
	var d models.Device
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&d).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// GetDevicesByIntegration fetches all devices for a specific integration
func (s *GormDeviceStore) GetDevicesByIntegration(ctx context.Context, integrationID uint) ([]*models.Device, error) {
	var devices []*models.Device
	if err := s.db.WithContext(ctx).
		Where("integration_id = ?", integrationID).
		Find(&devices).Error; err != nil {
		return nil, err
	}
	return devices, nil
}

// DeleteDevice removes a device by ExternalID
func (s *GormDeviceStore) DeleteDevice(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&models.Device{}, id).Error
}
