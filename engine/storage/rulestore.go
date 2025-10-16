package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"home_automation_server/engine/rules"
)

type RuleStore interface {
	SaveRule(ctx context.Context, r *rules.Rule) error
	LoadRules(ctx context.Context) (*rules.RuleSet, error)
	UpdateLastTriggered(ctx context.Context, r rules.Rule) error
	EnsureTableExists() error
}

type MSqlRuleStore struct {
	db *sql.DB
}

func NewMSqlRuleStore(db *sql.DB) *MSqlRuleStore {
	return &MSqlRuleStore{db: db}
}

// TODO: move to using migrations.. doing this for now
func (s *MSqlRuleStore) EnsureTableExists() error {
	query := `
	CREATE TABLE IF NOT EXISTS rules (
    	alias VARCHAR(255) PRIMARY KEY,
    	trigger_json JSON NOT NULL,
    	condition_json JSON NOT NULL,
    	action_json JSON NOT NULL,
    	active BOOLEAN NOT NULL,
		last_triggered DATETIME DEFAULT NULL,
    	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	)`
	_, err := s.db.Exec(query)
	return err
}

func (s *MSqlRuleStore) UpdateLastTriggered(ctx context.Context, r rules.Rule) error {
	_, err := s.db.ExecContext(ctx, "UPDATE rules SET last_triggered = NOW() WHERE alias = ?", r.Alias)
	if err != nil {
		return fmt.Errorf("failed to update last_triggered for rule: %w", err)
	}
	return nil
}

func (s *MSqlRuleStore) SaveRule(ctx context.Context, r *rules.Rule) error {
	triggerJSON, err := json.Marshal(r.Trigger)
	if err != nil {
		return fmt.Errorf("marshal trigger: %w", err)
	}
	conditionJSON, err := json.Marshal(r.Condition)
	if err != nil {
		return fmt.Errorf("marshal condition: %w", err)
	}
	actionJSON, err := json.Marshal(r.Action)
	if err != nil {
		return fmt.Errorf("marshal action: %w", err)
	}

	query := `
	INSERT INTO rules (alias, trigger_json, condition_json, action_json)
	VALUES (?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
		trigger_json = VALUES(trigger_json),
		condition_json = VALUES(condition_json),
		action_json = VALUES(action_json),
		updated_at = CURRENT_TIMESTAMP
	`

	_, err = s.db.ExecContext(ctx, query, r.Alias, triggerJSON, conditionJSON, actionJSON)
	if err != nil {
		return fmt.Errorf("save rule '%s': %w", r.Alias, err)
	}
	return nil
}

func (s *MSqlRuleStore) LoadRules(ctx context.Context) (*rules.RuleSet, error) {
	const query = `
		SELECT alias, trigger_json, condition_json, action_json, active
		FROM rules
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("LoadRules: failed to query rules: %w", err)
	}
	defer rows.Close()

	var ruleSet rules.RuleSet

	for rows.Next() {
		var (
			alias         string
			triggerJSON   []byte
			conditionJSON []byte
			actionJSON    []byte
			active        bool
			rule          rules.Rule
		)

		if err := rows.Scan(&alias, &triggerJSON, &conditionJSON, &actionJSON, &active); err != nil {
			return nil, fmt.Errorf("LoadRules: failed to scan rule row: %w", err)
		}

		rule.Alias = alias
		rule.Active = active

		if err := json.Unmarshal(triggerJSON, &rule.Trigger); err != nil {
			return nil, fmt.Errorf("LoadRules: failed to unmarshal trigger for rule %q: %w", alias, err)
		}
		if err := json.Unmarshal(conditionJSON, &rule.Condition); err != nil {
			return nil, fmt.Errorf("LoadRules: failed to unmarshal condition for rule %q: %w", alias, err)
		}
		if err := json.Unmarshal(actionJSON, &rule.Action); err != nil {
			return nil, fmt.Errorf("LoadRules: failed to unmarshal action for rule %q: %w", alias, err)
		}

		ruleSet.Rules = append(ruleSet.Rules, rule)
	}

	// Always check for errors after iterating rows
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("LoadRules: row iteration error: %w", err)
	}

	return &ruleSet, nil
}
