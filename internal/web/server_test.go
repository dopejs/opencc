package web

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/dopejs/opencc/internal/config"
)

func setupTestServer(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	config.ResetDefaultStore()
	t.Cleanup(func() { config.ResetDefaultStore() })

	// Create config dir and file
	configDir := filepath.Join(dir, config.ConfigDir)
	os.MkdirAll(configDir, 0755)
	cfg := &config.OpenCCConfig{
		Providers: map[string]*config.ProviderConfig{
			"test-provider": {
				BaseURL:   "https://api.test.com",
				AuthToken: "sk-test-secret-token-1234",
				Model:     "claude-sonnet-4-5",
			},
			"backup": {
				BaseURL:   "https://api.backup.com",
				AuthToken: "sk-backup-token-5678",
			},
		},
		Profiles: map[string][]string{
			"default": {"test-provider", "backup"},
			"work":    {"test-provider"},
		},
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(filepath.Join(configDir, config.ConfigFile), data, 0600)

	// Force reload
	config.DefaultStore()

	logger := log.New(io.Discard, "", 0)
	return NewServer("1.0.0-test", logger)
}

func doRequest(s *Server, method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reqBody = bytes.NewReader(data)
	}
	req := httptest.NewRequest(method, path, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	s.httpServer.Handler.ServeHTTP(w, req)
	return w
}

func decodeJSON(t *testing.T, r *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}
}

// --- Health ---

func TestHealthEndpoint(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/health", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	decodeJSON(t, w, &resp)
	if resp["status"] != "ok" {
		t.Errorf("status = %q", resp["status"])
	}
	if resp["version"] != "1.0.0-test" {
		t.Errorf("version = %q", resp["version"])
	}
}

func TestHealthMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/health", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// --- Security Headers ---

func TestSecurityHeaders(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/health", nil)

	if v := w.Header().Get("X-Content-Type-Options"); v != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q", v)
	}
	if v := w.Header().Get("X-Frame-Options"); v != "DENY" {
		t.Errorf("X-Frame-Options = %q", v)
	}
}

// --- Providers ---

func TestListProviders(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/providers", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var providers []providerResponse
	decodeJSON(t, w, &providers)
	if len(providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(providers))
	}

	// Tokens should be returned unmasked
	expectedTokens := map[string]string{
		"test-provider": "sk-test-secret-token-1234",
		"backup":        "sk-backup-token-5678",
	}
	for _, p := range providers {
		if expected, ok := expectedTokens[p.Name]; ok {
			if p.AuthToken != expected {
				t.Errorf("token for %s should be %q, got %q", p.Name, expected, p.AuthToken)
			}
		}
	}
}

func TestGetProvider(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/providers/test-provider", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var p providerResponse
	decodeJSON(t, w, &p)
	if p.Name != "test-provider" {
		t.Errorf("name = %q", p.Name)
	}
	if p.BaseURL != "https://api.test.com" {
		t.Errorf("base_url = %q", p.BaseURL)
	}
}

func TestGetProviderNotFound(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/providers/nonexistent", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestCreateProvider(t *testing.T) {
	s := setupTestServer(t)

	body := createProviderRequest{
		Name: "new-provider",
		Config: config.ProviderConfig{
			BaseURL:   "https://api.new.com",
			AuthToken: "sk-new-token",
			Model:     "claude-opus-4-5",
		},
	}
	w := doRequest(s, "POST", "/api/v1/providers", body)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// Verify it's persisted
	w2 := doRequest(s, "GET", "/api/v1/providers/new-provider", nil)
	if w2.Code != http.StatusOK {
		t.Fatalf("created provider not found")
	}
}

func TestCreateProviderConflict(t *testing.T) {
	s := setupTestServer(t)

	body := createProviderRequest{
		Name: "test-provider",
		Config: config.ProviderConfig{
			BaseURL:   "https://dup.com",
			AuthToken: "tok",
		},
	}
	w := doRequest(s, "POST", "/api/v1/providers", body)
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestCreateProviderNoName(t *testing.T) {
	s := setupTestServer(t)
	body := createProviderRequest{
		Config: config.ProviderConfig{BaseURL: "https://x.com", AuthToken: "tok"},
	}
	w := doRequest(s, "POST", "/api/v1/providers", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateProvider(t *testing.T) {
	s := setupTestServer(t)

	update := config.ProviderConfig{
		BaseURL: "https://api.updated.com",
		Model:   "claude-opus-4-5",
	}
	w := doRequest(s, "PUT", "/api/v1/providers/test-provider", update)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp providerResponse
	decodeJSON(t, w, &resp)
	if resp.BaseURL != "https://api.updated.com" {
		t.Errorf("base_url = %q", resp.BaseURL)
	}
	if resp.Model != "claude-opus-4-5" {
		t.Errorf("model = %q", resp.Model)
	}
}

func TestUpdateProviderKeepsToken(t *testing.T) {
	s := setupTestServer(t)

	// Send empty token - should keep original
	update := config.ProviderConfig{
		BaseURL: "https://api.updated.com",
	}
	w := doRequest(s, "PUT", "/api/v1/providers/test-provider", update)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Verify token is still there by checking the store directly
	p := config.DefaultStore().GetProvider("test-provider")
	if p.AuthToken != "sk-test-secret-token-1234" {
		t.Errorf("token was changed, got %q", p.AuthToken)
	}
}

func TestUpdateProviderNotFound(t *testing.T) {
	s := setupTestServer(t)
	update := config.ProviderConfig{BaseURL: "https://x.com"}
	w := doRequest(s, "PUT", "/api/v1/providers/nonexistent", update)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestDeleteProvider(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "DELETE", "/api/v1/providers/backup", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Verify deleted
	w2 := doRequest(s, "GET", "/api/v1/providers/backup", nil)
	if w2.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", w2.Code)
	}

	// Verify cascade: backup should be removed from default profile
	order := config.DefaultStore().GetProfileOrder("default")
	for _, n := range order {
		if n == "backup" {
			t.Error("backup should have been removed from default profile")
		}
	}
}

func TestDeleteProviderNotFound(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "DELETE", "/api/v1/providers/nonexistent", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// --- Profiles ---

func TestListProfiles(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/profiles", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var profiles []profileResponse
	decodeJSON(t, w, &profiles)
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}
}

func TestGetProfile(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/profiles/default", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var p profileResponse
	decodeJSON(t, w, &p)
	if p.Name != "default" {
		t.Errorf("name = %q", p.Name)
	}
	if len(p.Providers) != 2 {
		t.Errorf("expected 2 providers in default profile, got %d", len(p.Providers))
	}
}

func TestGetProfileNotFound(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/profiles/nonexistent", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestCreateProfile(t *testing.T) {
	s := setupTestServer(t)

	body := createProfileRequest{
		Name:      "staging",
		Providers: []string{"backup", "test-provider"},
	}
	w := doRequest(s, "POST", "/api/v1/profiles", body)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// Verify
	w2 := doRequest(s, "GET", "/api/v1/profiles/staging", nil)
	if w2.Code != http.StatusOK {
		t.Fatalf("created profile not found")
	}
	var p profileResponse
	decodeJSON(t, w2, &p)
	if len(p.Providers) != 2 || p.Providers[0] != "backup" {
		t.Errorf("providers = %v", p.Providers)
	}
}

func TestCreateProfileConflict(t *testing.T) {
	s := setupTestServer(t)
	body := createProfileRequest{
		Name:      "default",
		Providers: []string{"test-provider"},
	}
	w := doRequest(s, "POST", "/api/v1/profiles", body)
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestCreateProfileNoName(t *testing.T) {
	s := setupTestServer(t)
	body := createProfileRequest{Providers: []string{"test-provider"}}
	w := doRequest(s, "POST", "/api/v1/profiles", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateProfile(t *testing.T) {
	s := setupTestServer(t)

	body := updateProfileRequest{Providers: []string{"backup"}}
	w := doRequest(s, "PUT", "/api/v1/profiles/work", body)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var p profileResponse
	decodeJSON(t, w, &p)
	if len(p.Providers) != 1 || p.Providers[0] != "backup" {
		t.Errorf("providers = %v", p.Providers)
	}
}

func TestUpdateProfileNotFound(t *testing.T) {
	s := setupTestServer(t)
	body := updateProfileRequest{Providers: []string{"test-provider"}}
	w := doRequest(s, "PUT", "/api/v1/profiles/nonexistent", body)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestDeleteProfile(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "DELETE", "/api/v1/profiles/work", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Verify deleted
	w2 := doRequest(s, "GET", "/api/v1/profiles/work", nil)
	if w2.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", w2.Code)
	}
}

func TestDeleteProfileDefault(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "DELETE", "/api/v1/profiles/default", nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestDeleteProfileNotFound(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "DELETE", "/api/v1/profiles/nonexistent", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// --- Reload ---

func TestReload(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/reload", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	decodeJSON(t, w, &resp)
	if resp["status"] != "reloaded" {
		t.Errorf("status = %q", resp["status"])
	}
}

func TestReloadMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/reload", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// --- Token masking ---

func TestMaskToken(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"sk-test-secret-token-1234", "sk-te...1234"},
		{"short", "****"},
		{"12345678", "****"},
		{"123456789", "12345...6789"},
	}
	for _, tt := range tests {
		got := maskToken(tt.input)
		if got != tt.want {
			t.Errorf("maskToken(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
