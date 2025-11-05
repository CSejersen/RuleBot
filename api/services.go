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

type ValidateParamRequest struct {
	Service   string `json:"service"`
	ParamName string `json:"param_name"`
	Value     any    `json:"value"`
}

type ValidateParamResponse struct {
	ResolvedValue any `json:"resolved_value"`
}

func (s *Server) handleValidateParam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ValidateParamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	resolved, err := s.Engine.ResolveActionParam(req.Value)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ValidateParamResponse{ResolvedValue: resolved})
}
