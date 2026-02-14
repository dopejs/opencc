package web

import (
	"encoding/json"
	"net/http"

	"github.com/dopejs/opencc/internal/config"
)

// settingsResponse is the JSON shape for global settings.
type settingsResponse struct {
	DefaultProfile string   `json:"default_profile"`
	DefaultCLI     string   `json:"default_cli"`
	WebPort        int      `json:"web_port"`
	Profiles       []string `json:"profiles"`       // available profiles for selection
	CLIs           []string `json:"clis"`           // available CLIs
}

// settingsRequest is the JSON shape for updating settings.
type settingsRequest struct {
	DefaultProfile string `json:"default_profile,omitempty"`
	DefaultCLI     string `json:"default_cli,omitempty"`
	WebPort        int    `json:"web_port,omitempty"`
}

var availableCLIs = []string{"claude", "codex", "opencode"}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getSettings(w, r)
	case http.MethodPut:
		s.updateSettings(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) getSettings(w http.ResponseWriter, r *http.Request) {
	store := config.DefaultStore()
	profiles := store.ListProfiles()

	resp := settingsResponse{
		DefaultProfile: store.GetDefaultProfile(),
		DefaultCLI:     store.GetDefaultCLI(),
		WebPort:        store.GetWebPort(),
		Profiles:       profiles,
		CLIs:           availableCLIs,
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) updateSettings(w http.ResponseWriter, r *http.Request) {
	var req settingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	store := config.DefaultStore()

	// Update default profile if provided
	if req.DefaultProfile != "" {
		// Verify profile exists
		if store.GetProfileOrder(req.DefaultProfile) == nil {
			writeError(w, http.StatusBadRequest, "profile not found")
			return
		}
		if err := store.SetDefaultProfile(req.DefaultProfile); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// Update default CLI if provided
	if req.DefaultCLI != "" {
		valid := false
		for _, cli := range availableCLIs {
			if cli == req.DefaultCLI {
				valid = true
				break
			}
		}
		if !valid {
			writeError(w, http.StatusBadRequest, "invalid CLI")
			return
		}
		if err := store.SetDefaultCLI(req.DefaultCLI); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// Update web port if provided
	if req.WebPort > 0 {
		if req.WebPort < 1024 || req.WebPort > 65535 {
			writeError(w, http.StatusBadRequest, "port must be between 1024 and 65535")
			return
		}
		if err := store.SetWebPort(req.WebPort); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// Return updated settings
	s.getSettings(w, r)
}
