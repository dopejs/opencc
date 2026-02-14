package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dopejs/opencc/internal/config"
	"github.com/dopejs/opencc/internal/proxy/transform"
)

var (
	globalLogger     *StructuredLogger
	globalLoggerOnce sync.Once
	globalLoggerMu   sync.RWMutex
)

// InitGlobalLogger initializes the global structured logger.
func InitGlobalLogger(logDir string) error {
	var initErr error
	globalLoggerOnce.Do(func() {
		logger, err := NewStructuredLogger(logDir, 2000)
		if err != nil {
			initErr = err
			return
		}
		globalLoggerMu.Lock()
		globalLogger = logger
		globalLoggerMu.Unlock()
	})
	return initErr
}

// GetGlobalLogger returns the global structured logger.
func GetGlobalLogger() *StructuredLogger {
	globalLoggerMu.RLock()
	defer globalLoggerMu.RUnlock()
	return globalLogger
}

// RoutingConfig holds the default provider chain and optional scenario routes.
type RoutingConfig struct {
	DefaultProviders     []*Provider
	ScenarioRoutes       map[config.Scenario]*ScenarioProviders
	LongContextThreshold int // threshold for longContext scenario detection
}

// ScenarioProviders defines the providers and per-provider model overrides for a scenario.
type ScenarioProviders struct {
	Providers []*Provider
	Models    map[string]string // provider name → model override
}

type ProxyServer struct {
	Providers        []*Provider
	Routing          *RoutingConfig // optional; nil means use Providers as-is
	ClientFormat     string         // API format the client uses ("anthropic" or "openai")
	Logger           *log.Logger
	StructuredLogger *StructuredLogger
	Client           *http.Client
}

func NewProxyServer(providers []*Provider, logger *log.Logger) *ProxyServer {
	return &ProxyServer{
		Providers:        providers,
		ClientFormat:     config.ProviderTypeAnthropic, // Default: Claude Code uses Anthropic format
		Logger:           logger,
		StructuredLogger: GetGlobalLogger(),
		Client: &http.Client{
			Timeout: 10 * time.Minute,
		},
	}
}

// NewProxyServerWithRouting creates a proxy server with scenario-based routing.
func NewProxyServerWithRouting(routing *RoutingConfig, logger *log.Logger) *ProxyServer {
	return &ProxyServer{
		Providers:        routing.DefaultProviders,
		Routing:          routing,
		ClientFormat:     config.ProviderTypeAnthropic, // Default: Claude Code uses Anthropic format
		Logger:           logger,
		StructuredLogger: GetGlobalLogger(),
		Client: &http.Client{
			Timeout: 10 * time.Minute,
		},
	}
}

// NewProxyServerWithClientFormat creates a proxy server with a specific client format.
func NewProxyServerWithClientFormat(providers []*Provider, clientFormat string, logger *log.Logger) *ProxyServer {
	if clientFormat == "" {
		clientFormat = config.ProviderTypeAnthropic
	}
	return &ProxyServer{
		Providers:        providers,
		ClientFormat:     clientFormat,
		Logger:           logger,
		StructuredLogger: GetGlobalLogger(),
		Client: &http.Client{
			Timeout: 10 * time.Minute,
		},
	}
}

func (s *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadGateway)
		return
	}
	r.Body.Close()

	// Parse request body to extract session ID
	var bodyMap map[string]interface{}
	sessionID := ""
	if err := json.Unmarshal(bodyBytes, &bodyMap); err == nil {
		sessionID = extractSessionID(bodyMap)
	}

	// Determine provider chain and per-provider model overrides from routing
	providers := s.Providers
	var modelOverrides map[string]string

	if s.Routing != nil && len(s.Routing.ScenarioRoutes) > 0 {
		threshold := s.Routing.LongContextThreshold
		if threshold <= 0 {
			threshold = defaultLongContextThreshold
		}
		scenario, _ := DetectScenarioFromJSON(bodyBytes, threshold, sessionID)
		if sp, ok := s.Routing.ScenarioRoutes[scenario]; ok {
			providers = sp.Providers
			modelOverrides = sp.Models
			s.Logger.Printf("[routing] scenario=%s, providers=%d, model_overrides=%d",
				scenario, len(providers), len(modelOverrides))
		} else if scenario != config.ScenarioDefault {
			s.Logger.Printf("[routing] scenario=%s (no route configured, using default)", scenario)
		}
	}

	// Track provider failure details for error reporting
	type providerFailure struct {
		Name       string
		StatusCode int
		Body       string
	}
	var failures []providerFailure

	for i, p := range providers {
		isLast := i == len(providers)-1

		if !p.IsHealthy() && !isLast {
			msg := fmt.Sprintf("skipping (unhealthy, backoff %v)", p.Backoff)
			s.Logger.Printf("[%s] %s", p.Name, msg)
			s.logStructured(p.Name, r.Method, r.URL.Path, 0, LogLevelInfo, msg)
			continue
		}

		if !p.IsHealthy() && isLast {
			s.Logger.Printf("[%s] last provider, forcing request despite unhealthy (backoff %v)", p.Name, p.Backoff)
		}

		// Get model override for this specific provider
		var modelOverride string
		if modelOverrides != nil {
			modelOverride = modelOverrides[p.Name]
		}

		s.Logger.Printf("[%s] trying %s %s", p.Name, r.Method, r.URL.Path)
		resp, err := s.forwardRequest(r, p, bodyBytes, modelOverride)
		if err != nil {
			// Check if client canceled the request - don't mark provider unhealthy
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				msg := fmt.Sprintf("request canceled by client: %v", err)
				s.Logger.Printf("[%s] %s", p.Name, msg)
				s.logStructured(p.Name, r.Method, r.URL.Path, 0, LogLevelInfo, msg)
				// Return immediately - client is gone, no point in failover
				return
			}
			msg := fmt.Sprintf("request error: %v", err)
			s.Logger.Printf("[%s] %s", p.Name, msg)
			s.logStructuredError(p.Name, r.Method, r.URL.Path, err)
			failures = append(failures, providerFailure{Name: p.Name, StatusCode: 0, Body: err.Error()})
			p.MarkFailed()
			continue
		}

		// Auth/account errors → failover with long backoff
		if resp.StatusCode == 401 || resp.StatusCode == 402 || resp.StatusCode == 403 {
			errBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			msg := fmt.Sprintf("got %d (auth/account error), failing over", resp.StatusCode)
			s.Logger.Printf("[%s] %s response=%s", p.Name, msg, string(errBody))
			s.logStructuredWithResponse(p.Name, r.Method, r.URL.Path, resp.StatusCode, msg, errBody)
			failures = append(failures, providerFailure{Name: p.Name, StatusCode: resp.StatusCode, Body: string(errBody)})
			p.MarkAuthFailed()
			continue
		}

		// Rate limit → failover with short backoff
		if resp.StatusCode == 429 {
			errBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			msg := fmt.Sprintf("got %d (rate limited), failing over", resp.StatusCode)
			s.Logger.Printf("[%s] %s response=%s", p.Name, msg, string(errBody))
			s.logStructuredWithResponse(p.Name, r.Method, r.URL.Path, resp.StatusCode, msg, errBody)
			failures = append(failures, providerFailure{Name: p.Name, StatusCode: resp.StatusCode, Body: string(errBody)})
			p.MarkFailed()
			continue
		}

		// Server errors → check if request-related or server-side issue
		if resp.StatusCode >= 500 {
			// Read body to check error type
			errBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			if isRequestRelatedError(errBody) {
				// Request-related error (e.g., context too long) - failover without marking unhealthy
				msg := fmt.Sprintf("got %d (request-related error), failing over without backoff, request_body_size=%d", resp.StatusCode, len(bodyBytes))
				s.Logger.Printf("[%s] %s response=%s", p.Name, msg, string(errBody))
				s.logStructuredWithResponse(p.Name, r.Method, r.URL.Path, resp.StatusCode, msg, errBody)
				failures = append(failures, providerFailure{Name: p.Name, StatusCode: resp.StatusCode, Body: string(errBody)})
				continue
			}

			// Server-side issue - mark as failed with backoff
			msg := fmt.Sprintf("got %d (server error), failing over", resp.StatusCode)
			s.Logger.Printf("[%s] %s response=%s", p.Name, msg, string(errBody))
			s.logStructuredWithResponse(p.Name, r.Method, r.URL.Path, resp.StatusCode, msg, errBody)
			failures = append(failures, providerFailure{Name: p.Name, StatusCode: resp.StatusCode, Body: string(errBody)})
			p.MarkFailed()
			continue
		}

		p.MarkHealthy()
		msg := fmt.Sprintf("success %d", resp.StatusCode)
		s.Logger.Printf("[%s] %s", p.Name, msg)
		s.logStructured(p.Name, r.Method, r.URL.Path, resp.StatusCode, LogLevelInfo, msg)

		// Update session cache with token usage from response
		s.updateSessionCache(sessionID, resp)

		s.copyResponse(w, resp, p)
		return
	}

	// Build detailed error message with all provider failures
	var errMsg strings.Builder
	errMsg.WriteString("all providers failed\n")
	for _, f := range failures {
		if f.StatusCode > 0 {
			errMsg.WriteString(fmt.Sprintf("[%s] %d %s\n", f.Name, f.StatusCode, f.Body))
		} else {
			errMsg.WriteString(fmt.Sprintf("[%s] error: %s\n", f.Name, f.Body))
		}
	}

	errStr := errMsg.String()
	s.Logger.Printf("%s", errStr)
	if s.StructuredLogger != nil {
		s.StructuredLogger.Error("", errStr)
	}
	http.Error(w, errStr, http.StatusBadGateway)
}

// logStructured logs to the structured logger if available.
func (s *ProxyServer) logStructured(provider, method, path string, statusCode int, level LogLevel, message string) {
	if s.StructuredLogger == nil {
		return
	}
	s.StructuredLogger.Log(LogEntry{
		Level:      level,
		Provider:   provider,
		Method:     method,
		Path:       path,
		StatusCode: statusCode,
		Message:    message,
	})
}

// logStructuredError logs an error to the structured logger.
func (s *ProxyServer) logStructuredError(provider, method, path string, err error) {
	if s.StructuredLogger == nil {
		return
	}
	s.StructuredLogger.RequestError(provider, method, path, err)
}

// logStructuredWithResponse logs an error with response body to the structured logger.
func (s *ProxyServer) logStructuredWithResponse(provider, method, path string, statusCode int, message string, responseBody []byte) {
	if s.StructuredLogger == nil {
		return
	}
	s.StructuredLogger.RequestErrorWithResponse(provider, method, path, statusCode, message, responseBody)
}

func (s *ProxyServer) forwardRequest(r *http.Request, p *Provider, body []byte, modelOverride string) (*http.Response, error) {
	var modifiedBody []byte
	if modelOverride != "" {
		// Scenario routing: skip model mapping, use the override model directly
		modifiedBody = s.applyModelOverride(body, modelOverride, p.Name)
	} else {
		// Normal: apply per-provider model mapping
		modifiedBody = s.applyModelMapping(body, p)
	}

	// Apply request transformation if needed
	providerFormat := p.GetType()
	if transform.NeedsTransform(s.ClientFormat, providerFormat) {
		transformer := transform.GetTransformer(providerFormat)
		transformed, err := transformer.TransformRequest(modifiedBody, s.ClientFormat)
		if err != nil {
			s.Logger.Printf("[%s] transform request error: %v", p.Name, err)
		} else {
			s.Logger.Printf("[%s] transformed request: %s → %s", p.Name, s.ClientFormat, providerFormat)
			modifiedBody = transformed
		}
	}

	targetURL := singleJoiningSlash(p.BaseURL.String(), r.URL.Path)
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	req, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, bytes.NewReader(modifiedBody))
	if err != nil {
		return nil, err
	}

	// Copy headers
	for k, vv := range r.Header {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}

	// Override auth
	req.Header.Set("x-api-key", p.Token)
	req.Header.Set("Authorization", "Bearer "+p.Token)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(modifiedBody)))

	// Apply environment variable headers
	s.applyEnvVarsHeaders(req, p.EnvVars)

	return s.Client.Do(req)
}

func (s *ProxyServer) copyResponse(w http.ResponseWriter, resp *http.Response, p *Provider) {
	defer resp.Body.Close()

	// Check if response transformation is needed
	providerFormat := p.GetType()
	needsTransform := transform.NeedsTransform(s.ClientFormat, providerFormat)

	// Stream SSE responses (no transformation for streaming)
	if strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream") {
		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)

		flusher, ok := w.(http.Flusher)
		buf := make([]byte, 4096)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				w.Write(buf[:n])
				if ok {
					flusher.Flush()
				}
			}
			if err != nil {
				break
			}
		}
		return
	}

	// Non-streaming response - can apply transformation
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "failed to read response", http.StatusBadGateway)
		return
	}

	// Apply response transformation if needed
	if needsTransform && len(body) > 0 {
		transformer := transform.GetTransformer(providerFormat)
		transformed, err := transformer.TransformResponse(body, s.ClientFormat)
		if err != nil {
			s.Logger.Printf("[%s] transform response error: %v", p.Name, err)
		} else {
			s.Logger.Printf("[%s] transformed response: %s → %s", p.Name, providerFormat, s.ClientFormat)
			body = transformed
		}
	}

	// Copy headers (except Content-Length which may have changed)
	for k, vv := range resp.Header {
		if strings.ToLower(k) == "content-length" {
			continue
		}
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

// applyModelOverride replaces the model in the request body with the given override.
func (s *ProxyServer) applyModelOverride(body []byte, override string, providerName string) []byte {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return body
	}

	originalModel, _ := data["model"].(string)
	if originalModel == override {
		return body
	}

	s.Logger.Printf("[%s] model override: %s → %s", providerName, originalModel, override)
	data["model"] = override
	modified, err := json.Marshal(data)
	if err != nil {
		return body
	}
	return modified
}

// applyModelMapping detects the model type in the request and maps it to
// the provider's corresponding model. This ensures each provider gets the
// correct model name during failover.
//
// Mapping priority:
//  1. Thinking mode enabled → ReasoningModel
//  2. Model name contains "haiku" → HaikuModel
//  3. Model name contains "opus" → OpusModel
//  4. Model name contains "sonnet" → SonnetModel
//  5. Fallback → Model (default model)
func (s *ProxyServer) applyModelMapping(body []byte, p *Provider) []byte {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return body
	}

	originalModel, ok := data["model"].(string)
	if !ok || originalModel == "" {
		return body
	}

	mapped := s.mapModel(originalModel, data, p)
	if mapped == originalModel {
		return body
	}

	s.Logger.Printf("[%s] model mapping: %s → %s", p.Name, originalModel, mapped)
	data["model"] = mapped
	modified, err := json.Marshal(data)
	if err != nil {
		return body
	}
	return modified
}

// mapModel determines which provider model to use based on the request.
func (s *ProxyServer) mapModel(original string, body map[string]interface{}, p *Provider) string {
	// 1. Thinking mode → reasoning model
	if hasThinkingEnabled(body) && p.ReasoningModel != "" {
		return p.ReasoningModel
	}

	// 2. Match by model type (case-insensitive)
	lower := strings.ToLower(original)
	if strings.Contains(lower, "haiku") && p.HaikuModel != "" {
		return p.HaikuModel
	}
	if strings.Contains(lower, "opus") && p.OpusModel != "" {
		return p.OpusModel
	}
	if strings.Contains(lower, "sonnet") && p.SonnetModel != "" {
		return p.SonnetModel
	}

	// 3. Default model
	if p.Model != "" {
		return p.Model
	}

	// 4. No mapping — keep original
	return original
}

// updateSessionCache extracts token usage from the response and updates the session cache.
func (s *ProxyServer) updateSessionCache(sessionID string, resp *http.Response) {
	if sessionID == "" {
		return
	}

	// Read response body to extract usage information
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	// Restore body for copyResponse
	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	var respData map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &respData); err != nil {
		return
	}

	// Extract usage from response
	usage, ok := respData["usage"].(map[string]interface{})
	if !ok {
		return
	}

	inputTokens, _ := usage["input_tokens"].(float64)
	outputTokens, _ := usage["output_tokens"].(float64)

	if inputTokens > 0 || outputTokens > 0 {
		UpdateSessionUsage(sessionID, &SessionUsage{
			InputTokens:  int(inputTokens),
			OutputTokens: int(outputTokens),
		})
		s.Logger.Printf("[session] updated cache for %s: input=%d, output=%d",
			sessionID, int(inputTokens), int(outputTokens))
	}
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

// isRequestRelatedError checks if a 5xx error is caused by the request itself
// (e.g., context too long) rather than a server-side issue.
// These errors should trigger failover but not mark the provider as unhealthy.
func isRequestRelatedError(body []byte) bool {
	var errResp struct {
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &errResp); err != nil {
		return false
	}

	// Check for known request-related error types
	errType := strings.ToLower(errResp.Error.Type)
	errMsg := strings.ToLower(errResp.Error.Message)

	// invalid_request_error with context/token related messages
	if errType == "invalid_request_error" {
		return true
	}

	// Check message for context/token length issues
	contextKeywords := []string{
		"context", "token", "too long", "too large", "exceeds", "maximum",
		"limit", "length", "size", "prompt",
	}
	for _, kw := range contextKeywords {
		if strings.Contains(errMsg, kw) {
			return true
		}
	}

	return false
}

// applyEnvVarsHeaders converts environment variables to HTTP headers.
// Environment variable names are converted to lowercase and prefixed with "x-env-".
// For example: CLAUDE_CODE_MAX_OUTPUT_TOKENS -> x-env-claude-code-max-output-tokens
func (s *ProxyServer) applyEnvVarsHeaders(req *http.Request, envVars map[string]string) {
	if envVars == nil {
		return
	}

	for k, v := range envVars {
		if k == "" || v == "" {
			continue
		}
		// Convert env var name to HTTP header format
		// CLAUDE_CODE_MAX_OUTPUT_TOKENS -> x-env-claude-code-max-output-tokens
		headerName := "x-env-" + strings.ToLower(strings.ReplaceAll(k, "_", "-"))
		req.Header.Set(headerName, v)
	}
}

// StartProxy starts the proxy server and returns the port.
func StartProxy(providers []*Provider, listenAddr string, logger *log.Logger) (int, error) {
	srv := NewProxyServer(providers, logger)

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return 0, fmt.Errorf("listen: %w", err)
	}

	port := ln.Addr().(*net.TCPAddr).Port

	go http.Serve(ln, srv)

	return port, nil
}

// StartProxyWithRouting starts the proxy server with scenario-based routing.
func StartProxyWithRouting(routing *RoutingConfig, listenAddr string, logger *log.Logger) (int, error) {
	srv := NewProxyServerWithRouting(routing, logger)

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return 0, fmt.Errorf("listen: %w", err)
	}

	port := ln.Addr().(*net.TCPAddr).Port

	go http.Serve(ln, srv)

	return port, nil
}
