package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (s *Server) handleIntegrations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var names []string
	for name, _ := range s.Engine.Integrations {
		names = append(names, name)
	}

	json.NewEncoder(w).Encode(map[string]any{"integrations": names})
}

func (s *Server) handleIntegrationSubresources(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	// Expected: /api/integrations/{integration}/event-types
	if len(pathParts) != 4 || pathParts[3] != "event-types" {
		http.NotFound(w, r)
		return
	}

	integration := pathParts[2]

	intg, ok := s.Engine.Integrations[integration]
	if !ok {
		http.Error(w, fmt.Sprintf("unknown integration: %s", integration), http.StatusNotFound)
		return
	}

	eventTypes := intg.Translator.EventTypes()

	type EventInfo struct {
		Type         string   `json:"type"`
		Entities     []string `json:"entities,omitempty"`
		StateChanges []string `json:"state_changes,omitempty"`
	}

	events := []EventInfo{}
	for _, eventType := range eventTypes {
		entities := intg.Translator.EntitiesForType(eventType)
		stateChanges := intg.Translator.StateChangesForType(eventType)
		events = append(events, EventInfo{
			Type:         eventType,
			Entities:     entities,
			StateChanges: stateChanges,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"events": events,
	})
}
