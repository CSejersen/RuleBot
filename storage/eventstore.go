package storage

import (
	"context"
	"home_automation_server/storage/models"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type EventStore interface {
	SaveEvent(ctx context.Context, event models.Event) (models.Event, error)
}

type GormEventStore struct {
	db *gorm.DB
}

func NewGormEventStore(db *gorm.DB) *GormEventStore {
	return &GormEventStore{db: db}
}

// SaveEvent persists a single event (and its context if provided).
func (s *GormEventStore) SaveEvent(ctx context.Context, event models.Event) (models.Event, error) {
	var savedEvent models.Event
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// If the event includes a Context, make sure it's stored or already exists
		if event.Context != nil {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(event.Context).Error; err != nil {
				return err
			}

			event.ContextID = event.Context.ID
			// Set ID to Context.ID if not already set
			if event.ID == "" {
				event.ID = event.Context.ID
			}
		}

		if err := tx.Create(&event).Error; err != nil {
			return err
		}

		savedEvent = event
		return nil
	})

	return savedEvent, err
}
