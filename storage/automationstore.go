package storage

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"home_automation_server/storage/models"
	"time"
)

type AutomationStore interface {
	LoadAutomations(ctx context.Context) ([]models.Automation, error)
	UpdateLastTriggered(ctx context.Context, id uint) error
}

type GormRuleStore struct {
	db *gorm.DB
}

func NewGormRuleStore(db *gorm.DB) *GormRuleStore {
	return &GormRuleStore{db: db}
}

// LoadAutomations fetches all automations with their conditions and actions
func (s *GormRuleStore) LoadAutomations(ctx context.Context) ([]models.Automation, error) {
	var rules []models.Automation
	if err := s.db.WithContext(ctx).
		Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("failed to load rules: %w", err)
	}
	return rules, nil
}

// UpdateLastTriggered sets the last_triggered timestamp for an automation
func (s *GormRuleStore) UpdateLastTriggered(ctx context.Context, id uint) error {
	now := time.Now().UTC()
	return s.db.WithContext(ctx).Model(&models.Automation{}).Where("id = ?", id).Update("last_triggered", &now).Error
}
