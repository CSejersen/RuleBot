package storage

import (
	"context"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"home_automation_server/storage/models"
)

type EventStore interface {
	SaveEvent(ctx context.Context, event models.Event) error
}

type GormEventStore struct {
	db *gorm.DB
}

func NewGormEventStore(db *gorm.DB) *GormEventStore {
	return &GormEventStore{db: db}
}

// SaveEvent persists a single event (and its context if provided).
func (s *GormEventStore) SaveEvent(ctx context.Context, event models.Event) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// If the event includes a Context, make sure it's stored or already exists
		if event.Context != nil {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(event.Context).Error; err != nil {
				return err
			}

			event.ContextID = event.Context.ID
		}

		if err := tx.Create(&event).Error; err != nil {
			return err
		}

		return nil
	})
}
