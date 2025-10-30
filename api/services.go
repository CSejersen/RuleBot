package api

import (
	"encoding/json"
	"home_automation_server/integrations"
	"net/http"
)

type ServiceResponse struct {
	Name           string                                `json:"name"`
	RequiredParams map[string]integrations.ParamMetadata `json:"required_params"`
	AllowedTargets integrations.TargetSpec               `json:"allowed_targets"`
}

func (s *Server) handleServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := []ServiceResponse{}
	serviceSpecs := s.Engine.ServiceRegistry.GetAll()
	for name, spec := range serviceSpecs {
		serviceResponse := ServiceResponse{
			Name:           name,
			RequiredParams: spec.RequiredParams,
			AllowedTargets: spec.AllowedTargets,
		}
		resp = append(resp, serviceResponse)
	}
	json.NewEncoder(w).Encode(map[string]any{"services": resp})
}
