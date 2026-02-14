package web

import (
	"net/http"
	"strings"

	"github.com/dopejs/opencc/internal/config"
)

// providerResponse is the JSON shape returned for a single provider.
type providerResponse struct {
	Name            string            `json:"name"`
	Type            string            `json:"type,omitempty"`
	BaseURL         string            `json:"base_url"`
	AuthToken       string            `json:"auth_token"`
	Model           string            `json:"model,omitempty"`
	ReasoningModel  string            `json:"reasoning_model,omitempty"`
	HaikuModel      string            `json:"haiku_model,omitempty"`
	OpusModel       string            `json:"opus_model,omitempty"`
	SonnetModel     string            `json:"sonnet_model,omitempty"`
	EnvVars         map[string]string `json:"env_vars,omitempty"`
	ClaudeEnvVars   map[string]string `json:"claude_env_vars,omitempty"`
	CodexEnvVars    map[string]string `json:"codex_env_vars,omitempty"`
	OpenCodeEnvVars map[string]string `json:"opencode_env_vars,omitempty"`
}

type createProviderRequest struct {
	Name          string                `json:"name"`
	Config        config.ProviderConfig `json:"config"`
	AddToProfiles []string              `json:"add_to_profiles,omitempty"`
}

func toProviderResponse(name string, p *config.ProviderConfig, mask bool) providerResponse {
	token := p.AuthToken
	if mask {
		token = maskToken(token)
	}
	return providerResponse{
		Name:            name,
		Type:            p.Type,
		BaseURL:         p.BaseURL,
		AuthToken:       token,
		Model:           p.Model,
		ReasoningModel:  p.ReasoningModel,
		HaikuModel:      p.HaikuModel,
		OpusModel:       p.OpusModel,
		SonnetModel:     p.SonnetModel,
		EnvVars:         p.EnvVars,
		ClaudeEnvVars:   p.ClaudeEnvVars,
		CodexEnvVars:    p.CodexEnvVars,
		OpenCodeEnvVars: p.OpenCodeEnvVars,
	}
}

// handleProviders handles GET /api/v1/providers and POST /api/v1/providers.
func (s *Server) handleProviders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listProviders(w, r)
	case http.MethodPost:
		s.createProvider(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleProvider handles GET/PUT/DELETE /api/v1/providers/{name}.
func (s *Server) handleProvider(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/v1/providers/")
	if name == "" {
		writeError(w, http.StatusBadRequest, "provider name required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getProvider(w, r, name)
	case http.MethodPut:
		s.updateProvider(w, r, name)
	case http.MethodDelete:
		s.deleteProvider(w, r, name)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) listProviders(w http.ResponseWriter, r *http.Request) {
	store := config.DefaultStore()
	names := store.ProviderNames()
	providers := make([]providerResponse, 0, len(names))
	for _, name := range names {
		p := store.GetProvider(name)
		if p != nil {
			providers = append(providers, toProviderResponse(name, p, false))
		}
	}
	writeJSON(w, http.StatusOK, providers)
}

func (s *Server) getProvider(w http.ResponseWriter, r *http.Request, name string) {
	store := config.DefaultStore()
	p := store.GetProvider(name)
	if p == nil {
		writeError(w, http.StatusNotFound, "provider not found")
		return
	}
	writeJSON(w, http.StatusOK, toProviderResponse(name, p, false))
}

func (s *Server) createProvider(w http.ResponseWriter, r *http.Request) {
	var req createProviderRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	store := config.DefaultStore()
	if store.GetProvider(req.Name) != nil {
		writeError(w, http.StatusConflict, "provider already exists")
		return
	}

	if err := store.SetProvider(req.Name, &req.Config); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Add provider to requested profiles
	for _, profile := range req.AddToProfiles {
		order := store.GetProfileOrder(profile)
		if order != nil {
			order = append(order, req.Name)
			store.SetProfileOrder(profile, order)
		}
	}

	writeJSON(w, http.StatusCreated, toProviderResponse(req.Name, &req.Config, false))
}

func (s *Server) updateProvider(w http.ResponseWriter, r *http.Request, name string) {
	store := config.DefaultStore()
	existing := store.GetProvider(name)
	if existing == nil {
		writeError(w, http.StatusNotFound, "provider not found")
		return
	}

	var update config.ProviderConfig
	if err := readJSON(r, &update); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// If token is empty, keep the original.
	if update.AuthToken == "" {
		update.AuthToken = existing.AuthToken
	}

	if update.BaseURL != "" {
		existing.BaseURL = update.BaseURL
	}
	if update.AuthToken != "" {
		existing.AuthToken = update.AuthToken
	}
	existing.Type = update.Type
	existing.Model = update.Model
	existing.ReasoningModel = update.ReasoningModel
	existing.HaikuModel = update.HaikuModel
	existing.OpusModel = update.OpusModel
	existing.SonnetModel = update.SonnetModel
	existing.EnvVars = update.EnvVars
	existing.ClaudeEnvVars = update.ClaudeEnvVars
	existing.CodexEnvVars = update.CodexEnvVars
	existing.OpenCodeEnvVars = update.OpenCodeEnvVars

	if err := store.SetProvider(name, existing); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toProviderResponse(name, existing, false))
}

func (s *Server) deleteProvider(w http.ResponseWriter, r *http.Request, name string) {
	store := config.DefaultStore()
	if store.GetProvider(name) == nil {
		writeError(w, http.StatusNotFound, "provider not found")
		return
	}

	if err := store.DeleteProvider(name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
