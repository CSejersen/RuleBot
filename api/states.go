package api

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

func (s *Server) handleStatesSubresources(w http.ResponseWriter, r *http.Request) {
	// Get entity_id from query param
	entityID := r.URL.Query().Get("entity_id")
	if entityID == "" {
		http.Error(w, "missing entity_id query parameter", http.StatusBadRequest)
		return
	}

	state, ok := s.Engine.StateCache.Get(entityID)
	if !ok {
		s.Logger.Error("failed to fetch state for entity", zap.String("entityID", entityID))
		http.NotFound(w, r)
		return
	}

	// Return JSON with "state" key
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{"state": state}); err != nil {
		s.Logger.Error("failed to encode state response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
