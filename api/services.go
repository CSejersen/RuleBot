package api

import (
	"encoding/json"
	"home_automation_server/integrations/types"
	"net/http"
)

type ServiceResponse struct {
	Name               string                         `json:"name"`
	RequiredParams     map[string]types.ParamMetadata `json:"required_params"`
	RequiresTargetType bool                           `json:"requires_target_type"`
	RequiresTargetID   bool                           `json:"requires_target_id"`
}

func (s *Server) handleServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := []ServiceResponse{}
	services := s.Engine.ServiceRegistry.GetAll()
	for _, s := range services {
		resp = append(resp, ServiceResponse{
			Name:               s.FullName,
			RequiredParams:     s.RequiredParams,
			RequiresTargetType: s.RequiresTargetType,
			RequiresTargetID:   s.RequiresTargetID,
		})
	}
	json.NewEncoder(w).Encode(map[string]any{"services": resp})
}
