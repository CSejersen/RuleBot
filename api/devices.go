package api

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"home_automation_server/storage/models"
	"home_automation_server/types"
	"net/http"
	"strings"
	"time"
)

func (s *Server) handleDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var devices []*models.Device
	devices, err := s.Engine.DeviceStore.GetAllDevices(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("unable to fetch devices: %w", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(devices); err != nil {
		s.Logger.Error("Failed to encode devices response", zap.Error(err))
	}
}

// handleDevicesSubResources forwards requests for /api/devices/{id}/...
func (s *Server) handleDevicesSubResources(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	s.Logger.Debug("Handling device sub resources", zap.Strings("path-parts", pathParts))

	if len(pathParts) < 4 {
		http.Error(w, "Invalid device subresource path", http.StatusBadRequest)
		return
	}

	deviceID := pathParts[2]
	resource := pathParts[3]

	switch resource {
	case "entities":
		s.handleDeviceEntities(w, r, deviceID)
	case "states":
		s.handleDeviceEntityStates(w, r, deviceID)
	default:
		http.Error(w, "Unknown device subresource", http.StatusNotFound)
	}
}

// Example stubs for the actual handlers
func (s *Server) handleDeviceEntities(w http.ResponseWriter, r *http.Request, deviceID string) {
	http.Error(w, "not supported", http.StatusBadRequest)
}

func (s *Server) handleDeviceEntityStates(w http.ResponseWriter, r *http.Request, deviceID string) {
	s.Logger.Debug("Handling device state query", zap.String("device_id", deviceID))
	entities, err := s.Engine.EntityStore.GetEntitiesByDevice(r.Context(), deviceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("unable to fetch entities: %v", err), http.StatusInternalServerError)
		return
	}

	states := []types.State{}
	for _, entity := range entities {
		state, ok := s.Engine.StateCache.Get(entity.EntityID)
		if !ok {
			s.Logger.Info("no state for entity", zap.String("entity_id", entity.EntityID))
			continue
		}
		states = append(states, state)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{"states": states}); err != nil {
		s.Logger.Error("Failed to encode states response", zap.Error(err))
		http.Error(w, "Failed to encode states response", http.StatusInternalServerError)
	}
}
