package api

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

func (s *Server) handleIntegrationConfigSubresources(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/discover") && r.Method == http.MethodPost {
		s.handleDiscoverForIntegration(w, r)
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	return
}

func (s *Server) handleDiscoverForIntegration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract integration_name from URL: /api/integrations/configs/:name/discover
	nameStr := strings.TrimPrefix(r.URL.Path, "/api/integrations/configs/")
	nameStr = strings.TrimSuffix(nameStr, "/discover")

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	if err := s.Engine.DiscoverDevicesForIntegration(ctx, nameStr); err != nil {
		s.Logger.Error("Failed to discover devices", zap.Error(err), zap.String("integration", nameStr))
		http.Error(w, "failed to discover devices", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"discovery complete"}`))
}

func (s *Server) handleIntegrationDescriptors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	descriptors := s.Engine.IntegrationDescRegistry.List()
	json.NewEncoder(w).Encode(map[string]any{"descriptors": descriptors})
}

func (s *Server) handleIntegrations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var names []string
	for name, _ := range s.Engine.Integrations {
		names = append(names, name)
	}

	json.NewEncoder(w).Encode(map[string]any{"integration": names})
}

func (s *Server) handleIntegrationSubresources(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	// Expected: /api/integration/{integration}/event-types
	if len(pathParts) != 4 || pathParts[3] != "event-types" {
		http.NotFound(w, r)
		return
	}

	//integration := pathParts[2]

	type EventInfo struct {
		Type         string   `json:"type"`
		Entities     []string `json:"entities,omitempty"`
		StateChanges []string `json:"state_changes,omitempty"`
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"events": []string{"not supported"},
	})
}
