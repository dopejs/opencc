package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/dopejs/opencc/internal/config"
)

func discardLogger() *log.Logger {
	return log.New(io.Discard, "", 0)
}

// generateLongTextForTest creates varied text to get realistic token counts.
// Approximately 5.5 characters per token for English text.
func generateLongTextForTest(chars int) string {
	var sb strings.Builder
	words := []string{"hello", "world", "this", "is", "a", "test", "message", "with", "varied", "content"}
	wordIndex := 0
	for sb.Len() < chars {
		if wordIndex > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(words[wordIndex%len(words)])
		wordIndex++
	}
	return sb.String()
}

func TestSingleJoiningSlash(t *testing.T) {
	tests := []struct {
		a, b, want string
	}{
		{"http://host", "/path", "http://host/path"},
		{"http://host/", "/path", "http://host/path"},
		{"http://host/", "path", "http://host/path"},
		{"http://host", "path", "http://host/path"},
		{"http://host/api", "/v1/messages", "http://host/api/v1/messages"},
		{"http://host/api/", "/v1/messages", "http://host/api/v1/messages"},
	}
	for _, tt := range tests {
		got := singleJoiningSlash(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("singleJoiningSlash(%q, %q) = %q, want %q", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestModelMappingSonnet(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if data["model"] != "my-sonnet" {
			t.Errorf("model = %v, want %q", data["model"], "my-sonnet")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name: "test", BaseURL: u, Token: "t", Model: "default-model",
		SonnetModel: "my-sonnet", HaikuModel: "my-haiku", OpusModel: "my-opus",
		ReasoningModel: "my-reasoning", Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude-sonnet-4-5-20250929","prompt":"hi"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestModelMappingHaiku(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if data["model"] != "my-haiku" {
			t.Errorf("model = %v, want %q", data["model"], "my-haiku")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name: "test", BaseURL: u, Token: "t", Model: "default-model",
		SonnetModel: "my-sonnet", HaikuModel: "my-haiku", OpusModel: "my-opus",
		Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude-haiku-4-5","prompt":"hi"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestModelMappingOpus(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if data["model"] != "my-opus" {
			t.Errorf("model = %v, want %q", data["model"], "my-opus")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name: "test", BaseURL: u, Token: "t", Model: "default-model",
		SonnetModel: "my-sonnet", HaikuModel: "my-haiku", OpusModel: "my-opus",
		Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude-opus-4-5","prompt":"hi"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestModelMappingThinkingMode(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if data["model"] != "my-reasoning" {
			t.Errorf("model = %v, want %q", data["model"], "my-reasoning")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name: "test", BaseURL: u, Token: "t", Model: "default-model",
		SonnetModel: "my-sonnet", ReasoningModel: "my-reasoning", Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude-sonnet-4-5","thinking":{"type":"enabled"}}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestModelMappingThinkingDisabledUsesSonnet(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if data["model"] != "my-sonnet" {
			t.Errorf("model = %v, want %q", data["model"], "my-sonnet")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name: "test", BaseURL: u, Token: "t", Model: "default-model",
		SonnetModel: "my-sonnet", ReasoningModel: "my-reasoning", Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude-sonnet-4-5","thinking":{"type":"disabled"}}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestModelMappingUnknownModelUsesDefault(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if data["model"] != "default-model" {
			t.Errorf("model = %v, want %q", data["model"], "default-model")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name: "test", BaseURL: u, Token: "t", Model: "default-model",
		SonnetModel: "my-sonnet", Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"some-unknown-model","prompt":"hi"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestModelMappingNoMappingKeepsOriginal(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if data["model"] != "claude-sonnet-4-5" {
			t.Errorf("model = %v, want %q", data["model"], "claude-sonnet-4-5")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name: "test", BaseURL: u, Token: "t", Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude-sonnet-4-5","prompt":"hi"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestModelMappingCaseInsensitive(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if data["model"] != "my-sonnet" {
			t.Errorf("model = %v, want %q", data["model"], "my-sonnet")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name: "test", BaseURL: u, Token: "t",
		SonnetModel: "my-sonnet", Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"Claude-SONNET-4-5","prompt":"hi"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestModelMappingInvalidJSON(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		// Invalid JSON should be passed through unchanged
		if string(body) != "not json" {
			t.Errorf("body = %q, want %q", string(body), "not json")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name: "test", BaseURL: u, Token: "t", Model: "test-model", Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader("not json"))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
}

func TestModelMappingFailoverUsesSecondProviderMapping(t *testing.T) {
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		// Second provider should use its own sonnet mapping
		if data["model"] != "provider2-sonnet" {
			t.Errorf("model = %v, want %q", data["model"], "provider2-sonnet")
		}
		w.WriteHeader(200)
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u1, Token: "t1", SonnetModel: "provider1-sonnet", Healthy: true},
		{Name: "p2", BaseURL: u2, Token: "t2", SonnetModel: "provider2-sonnet", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude-sonnet-4-5"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// TestFailoverAppliesAllProviderConfig verifies that when failing over to the
// second provider, auth token, base URL, and all model type mappings are
// correctly applied from the second provider's configuration.
func TestFailoverAppliesAllProviderConfig(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantModel string
	}{
		{"sonnet", `{"model":"claude-sonnet-4-5"}`, "p2-sonnet"},
		{"haiku", `{"model":"claude-haiku-4-5"}`, "p2-haiku"},
		{"opus", `{"model":"claude-opus-4-5"}`, "p2-opus"},
		{"thinking", `{"model":"claude-sonnet-4-5","thinking":{"type":"enabled"}}`, "p2-reasoning"},
		{"unknown fallback", `{"model":"some-custom-model"}`, "p2-default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(500)
			}))
			defer backend1.Close()

			backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify auth token from second provider
				if r.Header.Get("x-api-key") != "token-p2" {
					t.Errorf("x-api-key = %q, want %q", r.Header.Get("x-api-key"), "token-p2")
				}
				if r.Header.Get("Authorization") != "Bearer token-p2" {
					t.Errorf("Authorization = %q, want %q", r.Header.Get("Authorization"), "Bearer token-p2")
				}

				// Verify model mapping from second provider
				body, _ := io.ReadAll(r.Body)
				var data map[string]interface{}
				json.Unmarshal(body, &data)
				if data["model"] != tt.wantModel {
					t.Errorf("model = %v, want %q", data["model"], tt.wantModel)
				}

				w.WriteHeader(200)
				w.Write([]byte(`{"ok":true}`))
			}))
			defer backend2.Close()

			u1, _ := url.Parse(backend1.URL)
			u2, _ := url.Parse(backend2.URL)
			providers := []*Provider{
				{
					Name: "p1", BaseURL: u1, Token: "token-p1",
					Model: "p1-default", SonnetModel: "p1-sonnet", HaikuModel: "p1-haiku",
					OpusModel: "p1-opus", ReasoningModel: "p1-reasoning", Healthy: true,
				},
				{
					Name: "p2", BaseURL: u2, Token: "token-p2",
					Model: "p2-default", SonnetModel: "p2-sonnet", HaikuModel: "p2-haiku",
					OpusModel: "p2-opus", ReasoningModel: "p2-reasoning", Healthy: true,
				},
			}

			srv := NewProxyServer(providers, discardLogger())
			req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Errorf("status = %d, want 200", w.Code)
			}
		})
	}
}

// TestFailoverThreeProviders verifies correct mapping when first two providers
// fail and the third succeeds.
func TestFailoverThreeProviders(t *testing.T) {
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer backend2.Close()

	backend3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "token-p3" {
			t.Errorf("x-api-key = %q, want %q", r.Header.Get("x-api-key"), "token-p3")
		}
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if data["model"] != "p3-haiku" {
			t.Errorf("model = %v, want %q", data["model"], "p3-haiku")
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer backend3.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)
	u3, _ := url.Parse(backend3.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u1, Token: "token-p1", HaikuModel: "p1-haiku", Healthy: true},
		{Name: "p2", BaseURL: u2, Token: "token-p2", HaikuModel: "p2-haiku", Healthy: true},
		{Name: "p3", BaseURL: u3, Token: "token-p3", HaikuModel: "p3-haiku", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude-haiku-4-5"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestHasThinkingEnabled(t *testing.T) {
	tests := []struct {
		name string
		body map[string]interface{}
		want bool
	}{
		{"enabled", map[string]interface{}{"thinking": map[string]interface{}{"type": "enabled"}}, true},
		{"disabled", map[string]interface{}{"thinking": map[string]interface{}{"type": "disabled"}}, false},
		{"no thinking", map[string]interface{}{}, false},
		{"thinking not object", map[string]interface{}{"thinking": "enabled"}, false},
		{"thinking no type", map[string]interface{}{"thinking": map[string]interface{}{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasThinkingEnabled(tt.body)
			if got != tt.want {
				t.Errorf("hasThinkingEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestServeHTTPSuccess tests a successful proxy request.
func TestServeHTTPSuccess(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth headers
		if r.Header.Get("x-api-key") != "test-token" {
			t.Errorf("x-api-key = %q", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}

		// Verify model mapping (sonnet → test-model via default)
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if data["model"] != "test-model" {
			t.Errorf("model = %v, want %q", data["model"], "test-model")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name: "test", BaseURL: u, Token: "test-token", Model: "test-model", Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger())

	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"some-model","prompt":"hi"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"ok":true`) {
		t.Errorf("body = %q", w.Body.String())
	}
}

// TestServeHTTPFailoverOn500 tests that 500 triggers failover to next provider.
func TestServeHTTPFailoverOn500(t *testing.T) {
	callCount := 0
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(500)
		w.Write([]byte("error"))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u1, Token: "t1", Model: "m", Healthy: true},
		{Name: "p2", BaseURL: u2, Token: "t2", Model: "m", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger())

	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200 (failover)", w.Code)
	}
	if callCount != 2 {
		t.Errorf("callCount = %d, want 2", callCount)
	}
}

// TestServeHTTPFailoverOn429 tests that 429 triggers failover.
func TestServeHTTPFailoverOn429(t *testing.T) {
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u1, Token: "t1", Healthy: true},
		{Name: "p2", BaseURL: u2, Token: "t2", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// TestServeHTTPAllProvidersFail tests 502 when all providers fail.
func TestServeHTTPAllProvidersFail(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadGateway)
	}
}

// TestServeHTTPSkipsUnhealthyProvider tests that unhealthy providers are skipped.
func TestServeHTTPSkipsUnhealthyProvider(t *testing.T) {
	called := make(map[string]bool)

	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called["p1"] = true
		w.WriteHeader(200)
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called["p2"] = true
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)
	p1 := &Provider{Name: "p1", BaseURL: u1, Token: "t1", Healthy: true}
	p2 := &Provider{Name: "p2", BaseURL: u2, Token: "t2", Healthy: true}

	// Mark p1 as unhealthy
	p1.MarkFailed()

	srv := NewProxyServer([]*Provider{p1, p2}, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if called["p1"] {
		t.Error("p1 should have been skipped (unhealthy)")
	}
	if !called["p2"] {
		t.Error("p2 should have been called")
	}
	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// TestServeHTTPNoModelInjectionWhenEmpty tests that empty model skips injection.
func TestServeHTTPNoModelInjectionWhenEmpty(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if _, ok := data["model"]; ok {
			t.Error("model should not be injected when provider model is empty")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Model: "", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"prompt":"hi"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
}

// TestServeHTTPPreservesQueryString tests that query params are forwarded.
func TestServeHTTPPreservesQueryString(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "beta=true" {
			t.Errorf("query = %q, want %q", r.URL.RawQuery, "beta=true")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages?beta=true", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
}

// TestServeHTTPSSEStreaming tests SSE response streaming.
func TestServeHTTPSSEStreaming(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
		w.Write([]byte("data: hello\n\n"))
		w.Write([]byte("data: world\n\n"))
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "data: hello") || !strings.Contains(body, "data: world") {
		t.Errorf("SSE body = %q", body)
	}
}

// TestStartProxy tests that StartProxy returns a valid port.
func TestStartProxy(t *testing.T) {
	u, _ := url.Parse("https://api.example.com")
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}

	port, err := StartProxy(providers, "anthropic", "127.0.0.1:0", discardLogger())
	if err != nil {
		t.Fatalf("StartProxy() error: %v", err)
	}
	if port <= 0 {
		t.Errorf("port = %d, want > 0", port)
	}

	// Verify the server is actually listening
	resp, err := http.Post(
		fmt.Sprintf("http://127.0.0.1:%d/v1/messages", port),
		"application/json",
		strings.NewReader(`{}`),
	)
	if err != nil {
		t.Fatalf("request to proxy error: %v", err)
	}
	resp.Body.Close()
	// Should get 502 since the backend URL is fake
	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadGateway)
	}
}

func TestNewProxyServer(t *testing.T) {
	u, _ := url.Parse("https://api.example.com")
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}
	srv := NewProxyServer(providers, discardLogger())
	if srv == nil {
		t.Fatal("NewProxyServer returned nil")
	}
	if len(srv.Providers) != 1 {
		t.Errorf("providers count = %d, want 1", len(srv.Providers))
	}
	if srv.Client == nil {
		t.Error("Client should not be nil")
	}
}

// TestServeHTTPCopiesResponseHeaders tests that response headers are forwarded.
func TestServeHTTPCopiesResponseHeaders(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "custom-value")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Header().Get("X-Custom-Header") != "custom-value" {
		t.Errorf("X-Custom-Header = %q, want %q", w.Header().Get("X-Custom-Header"), "custom-value")
	}
}

// TestStartProxyListenError tests that StartProxy returns error for invalid address.
func TestStartProxyListenError(t *testing.T) {
	u, _ := url.Parse("https://api.example.com")
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}

	// Use an invalid listen address
	_, err := StartProxy(providers, "anthropic", "999.999.999.999:0", discardLogger())
	if err == nil {
		t.Error("expected error for invalid listen address")
	}
}

// TestServeHTTPConnectionError tests failover when backend is unreachable.
func TestServeHTTPConnectionError(t *testing.T) {
	// Use a URL that will refuse connections
	u1, _ := url.Parse("http://127.0.0.1:1") // port 1 should refuse
	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	defer backend2.Close()
	u2, _ := url.Parse(backend2.URL)

	providers := []*Provider{
		{Name: "p1", BaseURL: u1, Token: "t1", Healthy: true},
		{Name: "p2", BaseURL: u2, Token: "t2", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200 (failover from connection error)", w.Code)
	}
}

// TestServeHTTPBadBodyRead tests handling of body read error.
func TestServeHTTPBadBodyRead(t *testing.T) {
	u, _ := url.Parse("https://api.example.com")
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", &errorReader{})
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadGateway)
	}
}

// errorReader always returns an error on Read.
type errorReader struct{}

func (e *errorReader) Read(p []byte) (int, error) {
	return 0, fmt.Errorf("read error")
}

// TestServeHTTP4xxNoFailover tests that non-auth 4xx (e.g. 400) don't trigger failover.
// Auth errors (401, 403) are tested separately and DO trigger failover.
func TestServeHTTP4xxNoFailover(t *testing.T) {
	callCount := 0
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(400)
		w.Write([]byte("bad request"))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(200)
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u1, Token: "t1", Healthy: true},
		{Name: "p2", BaseURL: u2, Token: "t2", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	// 400 should NOT trigger failover — only 429 and 5xx do
	if w.Code != 400 {
		t.Errorf("status = %d, want 400", w.Code)
	}
	if callCount != 1 {
		t.Errorf("callCount = %d, want 1 (no failover for 400)", callCount)
	}
}

// TestServeHTTPFailoverOn401 tests that 401 triggers failover to next provider.
func TestServeHTTPFailoverOn401(t *testing.T) {
	callCount := 0
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(401)
		w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u1, Token: "bad-token", Model: "m", Healthy: true},
		{Name: "p2", BaseURL: u2, Token: "good-token", Model: "m", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200 (failover from 401)", w.Code)
	}
	if callCount != 2 {
		t.Errorf("callCount = %d, want 2", callCount)
	}
}

// TestServeHTTPFailoverOn403 tests that 403 triggers failover to next provider.
func TestServeHTTPFailoverOn403(t *testing.T) {
	callCount := 0
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(403)
		w.Write([]byte(`{"error":"forbidden"}`))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u1, Token: "t1", Model: "m", Healthy: true},
		{Name: "p2", BaseURL: u2, Token: "t2", Model: "m", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200 (failover from 403)", w.Code)
	}
	if callCount != 2 {
		t.Errorf("callCount = %d, want 2", callCount)
	}
}

// TestServeHTTPFailoverOn402 tests that 402 (payment required) triggers failover.
func TestServeHTTPFailoverOn402(t *testing.T) {
	callCount := 0
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(402)
		w.Write([]byte(`{"error":"payment required"}`))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u1, Token: "t1", Model: "m", Healthy: true},
		{Name: "p2", BaseURL: u2, Token: "t2", Model: "m", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200 (failover from 402)", w.Code)
	}
	if callCount != 2 {
		t.Errorf("callCount = %d, want 2", callCount)
	}
}

// TestAuthFailedLongBackoff tests that auth failure (401/403) uses long backoff.
func TestAuthFailedLongBackoff(t *testing.T) {
	u, _ := url.Parse("https://api.example.com")
	p := &Provider{Name: "p1", BaseURL: u, Token: "t", Healthy: true}

	p.MarkAuthFailed()

	if p.Healthy {
		t.Error("expected Healthy = false after MarkAuthFailed")
	}
	if !p.AuthFailed {
		t.Error("expected AuthFailed = true after MarkAuthFailed")
	}
	if p.Backoff != AuthInitialBackoff {
		t.Errorf("Backoff = %v, want %v", p.Backoff, AuthInitialBackoff)
	}

	// Second auth failure should double the backoff
	p.MarkAuthFailed()
	want := AuthInitialBackoff * 2
	if p.Backoff != want {
		t.Errorf("Backoff after 2nd failure = %v, want %v", p.Backoff, want)
	}

	// Verify it's much larger than transient backoff
	if p.Backoff < MaxBackoff {
		t.Errorf("auth backoff %v should be larger than transient max %v", p.Backoff, MaxBackoff)
	}
}

// TestAuthFailedRecovery tests that a provider recovers after auth backoff expires.
func TestAuthFailedRecovery(t *testing.T) {
	u, _ := url.Parse("https://api.example.com")
	p := &Provider{Name: "p1", BaseURL: u, Token: "t", Healthy: true}

	p.MarkAuthFailed()

	// Immediately after failure, should be unhealthy
	if p.IsHealthy() {
		t.Error("expected unhealthy immediately after MarkAuthFailed")
	}

	// Simulate time passing beyond the backoff
	p.mu.Lock()
	p.FailedAt = time.Now().Add(-AuthInitialBackoff - time.Second)
	p.mu.Unlock()

	// Should now be considered healthy again
	if !p.IsHealthy() {
		t.Error("expected healthy after backoff period expires")
	}
}

// TestMarkHealthyClearsAuthFailed tests that MarkHealthy resets AuthFailed flag.
func TestMarkHealthyClearsAuthFailed(t *testing.T) {
	u, _ := url.Parse("https://api.example.com")
	p := &Provider{Name: "p1", BaseURL: u, Token: "t", Healthy: true}

	p.MarkAuthFailed()
	if !p.AuthFailed {
		t.Error("expected AuthFailed = true")
	}

	p.MarkHealthy()
	if p.AuthFailed {
		t.Error("expected AuthFailed = false after MarkHealthy")
	}
	if p.Backoff != 0 {
		t.Errorf("Backoff = %v, want 0 after MarkHealthy", p.Backoff)
	}
}

// --- Scenario routing tests ---

func TestRoutingThinkScenarioUsesThinkProviders(t *testing.T) {
	defaultCalled := false
	thinkCalled := false

	defaultBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defaultCalled = true
		w.WriteHeader(200)
	}))
	defer defaultBackend.Close()

	thinkBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		thinkCalled = true
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		// Model override should be applied
		if data["model"] != "think-model" {
			t.Errorf("model = %v, want %q", data["model"], "think-model")
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer thinkBackend.Close()

	u1, _ := url.Parse(defaultBackend.URL)
	u2, _ := url.Parse(thinkBackend.URL)

	defaultProvider := &Provider{Name: "default-p", BaseURL: u1, Token: "t1", Model: "m1", Healthy: true}
	thinkProvider := &Provider{Name: "think-p", BaseURL: u2, Token: "t2", Model: "m2", Healthy: true}

	routing := &RoutingConfig{
		DefaultProviders: []*Provider{defaultProvider},
		ScenarioRoutes: map[config.Scenario]*ScenarioProviders{
			config.ScenarioThink: {
				Providers: []*Provider{thinkProvider},
				Models:    map[string]string{"think-p": "think-model"},
			},
		},
	}

	srv := NewProxyServerWithRouting(routing, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
		`{"model":"claude-sonnet-4-5","thinking":{"type":"enabled"},"messages":[{"role":"user","content":"hi"}]}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if defaultCalled {
		t.Error("default provider should not have been called for think scenario")
	}
	if !thinkCalled {
		t.Error("think provider should have been called")
	}
}

func TestRoutingDefaultScenarioUsesDefaultProviders(t *testing.T) {
	defaultCalled := false

	defaultBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defaultCalled = true
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer defaultBackend.Close()

	u1, _ := url.Parse(defaultBackend.URL)
	defaultProvider := &Provider{Name: "default-p", BaseURL: u1, Token: "t1", Model: "m1", Healthy: true}

	routing := &RoutingConfig{
		DefaultProviders: []*Provider{defaultProvider},
		ScenarioRoutes: map[config.Scenario]*ScenarioProviders{
			config.ScenarioThink: {
				Providers: []*Provider{{Name: "think-p", BaseURL: u1, Token: "t2", Healthy: true}},
				Models:    map[string]string{"think-p": "think-model"},
			},
		},
	}

	srv := NewProxyServerWithRouting(routing, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
		`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"hello"}]}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if !defaultCalled {
		t.Error("default provider should have been called for non-matching scenario")
	}
}

func TestRoutingModelOverrideSkipsMapping(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		// Should use the override model, not the provider's sonnet mapping
		if data["model"] != "override-model" {
			t.Errorf("model = %v, want %q", data["model"], "override-model")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	provider := &Provider{
		Name: "p1", BaseURL: u, Token: "t",
		Model: "default-model", SonnetModel: "my-sonnet",
		Healthy: true,
	}

	routing := &RoutingConfig{
		DefaultProviders: []*Provider{provider},
		ScenarioRoutes: map[config.Scenario]*ScenarioProviders{
			config.ScenarioThink: {
				Providers: []*Provider{provider},
				Models:    map[string]string{"p1": "override-model"},
			},
		},
	}

	srv := NewProxyServerWithRouting(routing, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
		`{"model":"claude-sonnet-4-5","thinking":{"type":"enabled"}}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestRoutingNoRoutingBackwardCompat(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		// Should use normal model mapping (sonnet)
		if data["model"] != "my-sonnet" {
			t.Errorf("model = %v, want %q", data["model"], "my-sonnet")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name: "p1", BaseURL: u, Token: "t",
		SonnetModel: "my-sonnet", Healthy: true,
	}}

	// No routing — plain old proxy
	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
		`{"model":"claude-sonnet-4-5","prompt":"hi"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestRoutingSharedProviderHealth(t *testing.T) {
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)

	// Same provider instance shared across default and think scenarios
	sharedProvider := &Provider{Name: "shared", BaseURL: u1, Token: "t1", Model: "m", Healthy: true}
	backupProvider := &Provider{Name: "backup", BaseURL: u2, Token: "t2", Model: "m", Healthy: true}

	routing := &RoutingConfig{
		DefaultProviders: []*Provider{sharedProvider, backupProvider},
		ScenarioRoutes: map[config.Scenario]*ScenarioProviders{
			config.ScenarioThink: {
				Providers: []*Provider{sharedProvider},
			},
		},
	}

	srv := NewProxyServerWithRouting(routing, discardLogger())

	// First request — default scenario. Provider "shared" will fail (500) and get marked unhealthy.
	req1 := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
		`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"hi"}]}`))
	w1 := httptest.NewRecorder()
	srv.ServeHTTP(w1, req1)

	if w1.Code != 200 {
		t.Errorf("first request status = %d, want 200 (failover to backup)", w1.Code)
	}

	// Now "shared" is unhealthy. A think scenario request should skip it too,
	// but will fallback to default providers where backup is healthy.
	req2 := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
		`{"model":"claude-sonnet-4-5","thinking":{"type":"enabled"},"messages":[{"role":"user","content":"think"}]}`))
	w2 := httptest.NewRecorder()
	srv.ServeHTTP(w2, req2)

	// Think scenario providers are unhealthy, but fallback to default providers succeeds
	if w2.Code != 200 {
		t.Errorf("second request status = %d, want 200 (fallback to default providers)", w2.Code)
	}
}

func TestRoutingScenarioFallbackAllFail(t *testing.T) {
	// Test that when both scenario and default providers fail, we get 502
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"server error"}`))
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)

	scenarioProvider := &Provider{Name: "scenario-p", BaseURL: u, Token: "t1", Model: "m", Healthy: true}
	defaultProvider := &Provider{Name: "default-p", BaseURL: u, Token: "t2", Model: "m", Healthy: true}

	routing := &RoutingConfig{
		DefaultProviders: []*Provider{defaultProvider},
		ScenarioRoutes: map[config.Scenario]*ScenarioProviders{
			config.ScenarioThink: {
				Providers: []*Provider{scenarioProvider},
			},
		},
	}

	srv := NewProxyServerWithRouting(routing, discardLogger())

	// Think scenario request - both scenario and default providers will fail
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
		`{"model":"claude-sonnet-4-5","thinking":{"type":"enabled"},"messages":[{"role":"user","content":"think"}]}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	// Both scenario and default providers failed → 502
	if w.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want 502 (all providers failed)", w.Code)
	}
}

func TestRoutingImageScenario(t *testing.T) {
	imageCalled := false

	imageBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		imageCalled = true
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer imageBackend.Close()

	u, _ := url.Parse(imageBackend.URL)
	imageProvider := &Provider{Name: "image-p", BaseURL: u, Token: "t", Healthy: true}

	routing := &RoutingConfig{
		DefaultProviders: []*Provider{},
		ScenarioRoutes: map[config.Scenario]*ScenarioProviders{
			config.ScenarioImage: {Providers: []*Provider{imageProvider}},
		},
	}

	srv := NewProxyServerWithRouting(routing, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
		`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":[{"type":"image","source":{"type":"base64","data":"abc"}}]}]}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if !imageCalled {
		t.Error("image provider should have been called")
	}
}

func TestRoutingLongContextScenario(t *testing.T) {
	defaultCalled := false
	longCtxCalled := false

	defaultBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defaultCalled = true
		w.WriteHeader(200)
	}))
	defer defaultBackend.Close()

	longCtxBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		longCtxCalled = true
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if data["model"] != "cheap-model" {
			t.Errorf("model = %v, want %q", data["model"], "cheap-model")
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer longCtxBackend.Close()

	u1, _ := url.Parse(defaultBackend.URL)
	u2, _ := url.Parse(longCtxBackend.URL)

	defaultProvider := &Provider{Name: "default-p", BaseURL: u1, Token: "t1", Model: "m1", Healthy: true}
	longCtxProvider := &Provider{Name: "cheap-p", BaseURL: u2, Token: "t2", Model: "m2", Healthy: true}

	routing := &RoutingConfig{
		DefaultProviders: []*Provider{defaultProvider},
		ScenarioRoutes: map[config.Scenario]*ScenarioProviders{
			config.ScenarioLongContext: {
				Providers: []*Provider{longCtxProvider},
				Models:    map[string]string{"cheap-p": "cheap-model"},
			},
		},
	}

	// Build a request with >32k tokens
	// Generate varied text to get realistic token count (~5.5 chars per token)
	longText := generateLongTextForTest(32000 * 6)
	reqBody := fmt.Sprintf(`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"%s"}]}`, longText)

	srv := NewProxyServerWithRouting(routing, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(reqBody))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if defaultCalled {
		t.Error("default provider should not have been called for longContext scenario")
	}
	if !longCtxCalled {
		t.Error("longContext provider should have been called")
	}
}

func TestRoutingScenarioFailover(t *testing.T) {
	// Scenario chain has two providers; first fails 500 → should failover to second
	p1Called := false
	p2Called := false

	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p1Called = true
		w.WriteHeader(500)
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p2Called = true
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		// Model override should persist through failover
		if data["model"] != "think-override" {
			t.Errorf("model = %v, want %q", data["model"], "think-override")
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)

	provider1 := &Provider{Name: "think-p1", BaseURL: u1, Token: "t1", Model: "m1", SonnetModel: "my-sonnet", Healthy: true}
	provider2 := &Provider{Name: "think-p2", BaseURL: u2, Token: "t2", Model: "m2", SonnetModel: "other-sonnet", Healthy: true}

	routing := &RoutingConfig{
		DefaultProviders: []*Provider{},
		ScenarioRoutes: map[config.Scenario]*ScenarioProviders{
			config.ScenarioThink: {
				Providers: []*Provider{provider1, provider2},
				Models:    map[string]string{"think-p1": "think-override", "think-p2": "think-override"},
			},
		},
	}

	srv := NewProxyServerWithRouting(routing, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
		`{"model":"claude-sonnet-4-5","thinking":{"type":"enabled"},"messages":[{"role":"user","content":"hi"}]}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if !p1Called {
		t.Error("first think provider should have been called (then failed)")
	}
	if !p2Called {
		t.Error("second think provider should have been called (failover)")
	}
}

func TestRoutingScenarioFailoverWithoutModelOverride(t *testing.T) {
	// Scenario chain with failover, no model override → each provider uses its own mapping
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		// No model override → should use provider2's sonnet mapping
		if data["model"] != "p2-sonnet" {
			t.Errorf("model = %v, want %q", data["model"], "p2-sonnet")
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)

	provider1 := &Provider{Name: "img-p1", BaseURL: u1, Token: "t1", SonnetModel: "p1-sonnet", Healthy: true}
	provider2 := &Provider{Name: "img-p2", BaseURL: u2, Token: "t2", SonnetModel: "p2-sonnet", Healthy: true}

	routing := &RoutingConfig{
		DefaultProviders: []*Provider{},
		ScenarioRoutes: map[config.Scenario]*ScenarioProviders{
			config.ScenarioImage: {
				Providers: []*Provider{provider1, provider2},
				// No Model → normal mapping per provider
			},
		},
	}

	srv := NewProxyServerWithRouting(routing, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
		`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":[{"type":"image","source":{"type":"base64","data":"abc"}}]}]}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestRoutingScenarioWithoutModelOverrideUsesNormalMapping(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		// No model override → should use provider's normal model mapping
		if data["model"] != "my-sonnet" {
			t.Errorf("model = %v, want %q (normal mapping)", data["model"], "my-sonnet")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	provider := &Provider{
		Name: "p1", BaseURL: u, Token: "t",
		SonnetModel: "my-sonnet", Healthy: true,
	}

	routing := &RoutingConfig{
		DefaultProviders: []*Provider{provider},
		ScenarioRoutes: map[config.Scenario]*ScenarioProviders{
			config.ScenarioImage: {
				Providers: []*Provider{provider},
				// No Model override → normal mapping should apply
			},
		},
	}

	srv := NewProxyServerWithRouting(routing, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
		`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":[{"type":"image","source":{}}]}]}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// TestEnvVarsAppliedAsHeaders tests that env vars are converted to HTTP headers.
func TestEnvVarsAppliedAsHeaders(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify env var headers are present
		if r.Header.Get("x-env-claude-code-max-output-tokens") != "64000" {
			t.Errorf("x-env-claude-code-max-output-tokens = %q, want 64000",
				r.Header.Get("x-env-claude-code-max-output-tokens"))
		}
		if r.Header.Get("x-env-max-thinking-tokens") != "50000" {
			t.Errorf("x-env-max-thinking-tokens = %q, want 50000",
				r.Header.Get("x-env-max-thinking-tokens"))
		}
		if r.Header.Get("x-env-claude-code-effort-level") != "high" {
			t.Errorf("x-env-claude-code-effort-level = %q, want high",
				r.Header.Get("x-env-claude-code-effort-level"))
		}
		if r.Header.Get("x-env-my-custom-var") != "custom_value" {
			t.Errorf("x-env-my-custom-var = %q, want custom_value",
				r.Header.Get("x-env-my-custom-var"))
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name:    "test",
		BaseURL: u,
		Token:   "test-token",
		EnvVars: map[string]string{
			"CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000",
			"MAX_THINKING_TOKENS":            "50000",
			"CLAUDE_CODE_EFFORT_LEVEL":       "high",
			"MY_CUSTOM_VAR":                  "custom_value",
		},
		Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude-sonnet-4-5"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// TestEnvVarsFailoverSwitchesEnvVars tests that failover switches to the second provider's env vars.
func TestEnvVarsFailoverSwitchesEnvVars(t *testing.T) {
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// First provider fails
		w.WriteHeader(500)
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify second provider's env vars are used
		if r.Header.Get("x-env-claude-code-max-output-tokens") != "32000" {
			t.Errorf("x-env-claude-code-max-output-tokens = %q, want 32000 (from provider2)",
				r.Header.Get("x-env-claude-code-max-output-tokens"))
		}
		if r.Header.Get("x-env-claude-code-effort-level") != "medium" {
			t.Errorf("x-env-claude-code-effort-level = %q, want medium (from provider2)",
				r.Header.Get("x-env-claude-code-effort-level"))
		}
		// Provider1's custom var should NOT be present
		if r.Header.Get("x-env-provider1-var") != "" {
			t.Errorf("x-env-provider1-var should not be present, got %q",
				r.Header.Get("x-env-provider1-var"))
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)
	providers := []*Provider{
		{
			Name:    "p1",
			BaseURL: u1,
			Token:   "token1",
			EnvVars: map[string]string{
				"CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000",
				"CLAUDE_CODE_EFFORT_LEVEL":       "high",
				"PROVIDER1_VAR":                  "p1_value",
			},
			Healthy: true,
		},
		{
			Name:    "p2",
			BaseURL: u2,
			Token:   "token2",
			EnvVars: map[string]string{
				"CLAUDE_CODE_MAX_OUTPUT_TOKENS": "32000",
				"CLAUDE_CODE_EFFORT_LEVEL":       "medium",
			},
			Healthy: true,
		},
	}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude-sonnet-4-5"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200 (failover)", w.Code)
	}
}

// TestEnvVarsEmptyMapNoHeaders tests that empty env vars map doesn't add headers.
func TestEnvVarsEmptyMapNoHeaders(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify no x-env- headers are present
		for k := range r.Header {
			if strings.HasPrefix(strings.ToLower(k), "x-env-") {
				t.Errorf("unexpected header %q", k)
			}
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name:    "test",
		BaseURL: u,
		Token:   "test-token",
		EnvVars: map[string]string{}, // Empty map
		Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// TestEnvVarsNilMapNoHeaders tests that nil env vars map doesn't add headers.
func TestEnvVarsNilMapNoHeaders(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify no x-env- headers are present
		for k := range r.Header {
			if strings.HasPrefix(strings.ToLower(k), "x-env-") {
				t.Errorf("unexpected header %q", k)
			}
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name:    "test",
		BaseURL: u,
		Token:   "test-token",
		EnvVars: nil, // Nil map
		Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// TestNewProxyServerWithClientFormat tests creating a proxy with specific client format.
func TestNewProxyServerWithClientFormat(t *testing.T) {
	u, _ := url.Parse("https://api.example.com")
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}

	tests := []struct {
		name         string
		clientFormat string
		wantFormat   string
	}{
		{"anthropic", "anthropic", "anthropic"},
		{"openai", "openai", "openai"},
		{"empty defaults to anthropic", "", "anthropic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := NewProxyServerWithClientFormat(providers, tt.clientFormat, discardLogger())
			if srv.ClientFormat != tt.wantFormat {
				t.Errorf("ClientFormat = %q, want %q", srv.ClientFormat, tt.wantFormat)
			}
		})
	}
}

// TestStartProxyWithClientFormat tests that StartProxy respects client format.
func TestStartProxyWithClientFormat(t *testing.T) {
	u, _ := url.Parse("https://api.example.com")
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}

	// Test with openai client format
	port, err := StartProxy(providers, "openai", "127.0.0.1:0", discardLogger())
	if err != nil {
		t.Fatalf("StartProxy() error: %v", err)
	}
	if port <= 0 {
		t.Errorf("port = %d, want > 0", port)
	}
}

// TestStartProxyWithRoutingClientFormat tests that StartProxyWithRouting respects client format.
func TestStartProxyWithRoutingClientFormat(t *testing.T) {
	u, _ := url.Parse("https://api.example.com")
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}

	routing := &RoutingConfig{
		DefaultProviders: providers,
	}

	port, err := StartProxyWithRouting(routing, "openai", "127.0.0.1:0", discardLogger())
	if err != nil {
		t.Fatalf("StartProxyWithRouting() error: %v", err)
	}
	if port <= 0 {
		t.Errorf("port = %d, want > 0", port)
	}
}
