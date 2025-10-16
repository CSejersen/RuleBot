package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"home_automation_server/engine/types"
)

type EventStore interface {
	SaveEvent(ctx context.Context, e *types.ProcessedEvent) error
	EnsureTableExists() error
}

type MSqlEventStore struct {
	db *sql.DB
}

func NewMSqlEventStore(db *sql.DB) *MSqlEventStore {
	return &MSqlEventStore{db: db}
}

// TODO: move to using migrations.. doing this for now
func (s *MSqlEventStore) EnsureTableExists() error {
	query := `
	CREATE TABLE IF NOT EXISTS events (
		id VARCHAR(255) PRIMARY KEY,
		source VARCHAR(255),
		type VARCHAR(255),
		entity VARCHAR(255),
		state_change VARCHAR(255),
		timestamp DATETIME,
		payload JSON,
		triggered_rules JSON
	)`
	_, err := s.db.Exec(query)
	return err
}

func (s *MSqlEventStore) SaveEvent(ctx context.Context, e *types.ProcessedEvent) error {

	payloadJSON, err := json.Marshal(e.Event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	triggeredRulesJSON, err := json.Marshal(e.TriggeredRules)
	if err != nil {
		return fmt.Errorf("failed to marshal triggered rules: %w", err)
	}

	query := `
		INSERT INTO events (id, source, type, entity, state_change, timestamp, payload, triggered_rules)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.ExecContext(ctx, query, e.Event.Id, e.Event.Source, e.Event.Type, e.Event.Entity, e.Event.StateChange, e.Event.Time, payloadJSON, triggeredRulesJSON)
	if err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	return nil
}
