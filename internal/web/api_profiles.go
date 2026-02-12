package web

import (
	"net/http"
	"strings"

	"github.com/dopejs/opencc/internal/config"
)

// profileResponse is the JSON shape returned for a single profile.
type profileResponse struct {
	Name      string   `json:"name"`
	Providers []string `json:"providers"`
}

type createProfileRequest struct {
	Name      string   `json:"name"`
	Providers []string `json:"providers"`
}

type updateProfileRequest struct {
	Providers []string `json:"providers"`
}

// handleProfiles handles GET /api/v1/profiles and POST /api/v1/profiles.
func (s *Server) handleProfiles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listProfiles(w, r)
	case http.MethodPost:
		s.createProfile(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleProfile handles GET/PUT/DELETE /api/v1/profiles/{name}.
func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/v1/profiles/")
	if name == "" {
		writeError(w, http.StatusBadRequest, "profile name required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getProfile(w, r, name)
	case http.MethodPut:
		s.updateProfile(w, r, name)
	case http.MethodDelete:
		s.deleteProfile(w, r, name)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) listProfiles(w http.ResponseWriter, r *http.Request) {
	store := config.DefaultStore()
	names := store.ListProfiles()
	profiles := make([]profileResponse, 0, len(names))
	for _, name := range names {
		order := store.GetProfileOrder(name)
		if order == nil {
			order = []string{}
		}
		profiles = append(profiles, profileResponse{
			Name:      name,
			Providers: order,
		})
	}
	writeJSON(w, http.StatusOK, profiles)
}

func (s *Server) getProfile(w http.ResponseWriter, r *http.Request, name string) {
	store := config.DefaultStore()
	order := store.GetProfileOrder(name)
	if order == nil {
		writeError(w, http.StatusNotFound, "profile not found")
		return
	}
	writeJSON(w, http.StatusOK, profileResponse{
		Name:      name,
		Providers: order,
	})
}

func (s *Server) createProfile(w http.ResponseWriter, r *http.Request) {
	var req createProfileRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	store := config.DefaultStore()
	existing := store.GetProfileOrder(req.Name)
	if existing != nil {
		writeError(w, http.StatusConflict, "profile already exists")
		return
	}

	providers := req.Providers
	if providers == nil {
		providers = []string{}
	}

	if err := store.SetProfileOrder(req.Name, providers); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, profileResponse{
		Name:      req.Name,
		Providers: providers,
	})
}

func (s *Server) updateProfile(w http.ResponseWriter, r *http.Request, name string) {
	store := config.DefaultStore()
	existing := store.GetProfileOrder(name)
	if existing == nil {
		writeError(w, http.StatusNotFound, "profile not found")
		return
	}

	var req updateProfileRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	providers := req.Providers
	if providers == nil {
		providers = []string{}
	}

	if err := store.SetProfileOrder(name, providers); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, profileResponse{
		Name:      name,
		Providers: providers,
	})
}

func (s *Server) deleteProfile(w http.ResponseWriter, r *http.Request, name string) {
	if name == "default" {
		writeError(w, http.StatusForbidden, "cannot delete the default profile")
		return
	}

	store := config.DefaultStore()
	existing := store.GetProfileOrder(name)
	if existing == nil {
		writeError(w, http.StatusNotFound, "profile not found")
		return
	}

	if err := store.DeleteProfile(name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
