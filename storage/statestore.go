package storage

import (
	"context"
	"encoding/json"
	"errors"
	"home_automation_server/storage/models"
	"home_automation_server/types"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// StateStore defines methods to manage entity states in the database
type StateStore interface {
	SaveState(ctx context.Context, state types.State) error
	SaveAllStates(ctx context.Context, states []types.State) error
	LoadAllStates(ctx context.Context) ([]types.State, error)
	GetState(ctx context.Context, entityID string) (types.State, error)
}

// GormStateStore implements StateStore using GORM
type GormStateStore struct {
	db *gorm.DB
}

// NewGormStateStore creates a new GormStateStore
func NewGormStateStore(db *gorm.DB) *GormStateStore {
	return &GormStateStore{db: db}
}

// SaveState saves or updates a single state
func (s *GormStateStore) SaveState(ctx context.Context, state types.State) error {
	stateModel, err := StateToStorage(state)
	if err != nil {
		return err
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Upsert state - update if exists, insert if not
		return tx.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&stateModel).Error
	})
}

// SaveAllStates saves or updates all states in batch
func (s *GormStateStore) SaveAllStates(ctx context.Context, states []types.State) error {
	if len(states) == 0 {
		return nil
	}

	stateModels := make([]models.State, len(states))
	for i, state := range states {
		model, err := StateToStorage(state)
		if err != nil {
			return err
		}
		stateModels[i] = model
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Batch upsert all states
		return tx.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&stateModels).Error
	})
}

// LoadAllStates loads all states from the database
func (s *GormStateStore) LoadAllStates(ctx context.Context) ([]types.State, error) {
	var stateModels []models.State
	if err := s.db.WithContext(ctx).Find(&stateModels).Error; err != nil {
		return nil, err
	}

	states := make([]types.State, len(stateModels))
	for i, model := range stateModels {
		state, err := StateFromStorage(model)
		if err != nil {
			return nil, err
		}
		states[i] = state
	}

	return states, nil
}

// GetState retrieves a single state by entityID
func (s *GormStateStore) GetState(ctx context.Context, entityID string) (types.State, error) {
	var stateModel models.State
	err := s.db.WithContext(ctx).Where("entity_id = ?", entityID).First(&stateModel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.State{}, err
		}
		return types.State{}, err
	}

	return StateFromStorage(stateModel)
}

// StateToStorage converts types.State to models.State
func StateToStorage(state types.State) (models.State, error) {
	var stateBytes []byte
	var err error

	if state.State != nil {
		stateBytes, err = json.Marshal(state.State)
		if err != nil {
			return models.State{}, err
		}
	}

	var attributesBytes []byte
	if len(state.Attributes) > 0 {
		attributesBytes, err = json.Marshal(state.Attributes)
		if err != nil {
			return models.State{}, err
		}
	}

	return models.State{
		EntityID:    state.EntityID,
		State:       stateBytes,
		Attributes:  attributesBytes,
		LastChanged: state.LastChanged,
		LastUpdated: state.LastUpdated,
	}, nil
}

// StateFromStorage converts models.State to types.State
func StateFromStorage(model models.State) (types.State, error) {
	var stateValue any
	if len(model.State) > 0 {
		if err := json.Unmarshal(model.State, &stateValue); err != nil {
			return types.State{}, err
		}
	}

	var attributes map[string]any
	if len(model.Attributes) > 0 {
		attributes = make(map[string]any)
		if err := json.Unmarshal(model.Attributes, &attributes); err != nil {
			return types.State{}, err
		}
	}

	return types.State{
		EntityID:    model.EntityID,
		State:       stateValue,
		Attributes:  attributes,
		LastChanged: model.LastChanged,
		LastUpdated: model.LastUpdated,
		Context:     &types.Context{ID: uuid.NewString(), Origin: "state_store"}, // Context is not persisted for states
	}, nil
}
