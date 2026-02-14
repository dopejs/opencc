package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/dopejs/opencc/internal/config"
	"github.com/dopejs/opencc/internal/proxy"
)

// Server is the web configuration management server.
type Server struct {
	httpServer *http.Server
	logger     *log.Logger
	version    string
}

// NewServer creates a new web server bound to 127.0.0.1 on the configured port.
func NewServer(version string, logger *log.Logger) *Server {
	s := &Server{
		logger:  logger,
		version: version,
	}

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/v1/health", s.handleHealth)
	mux.HandleFunc("/api/v1/reload", s.handleReload)
	mux.HandleFunc("/api/v1/providers", s.handleProviders)
	mux.HandleFunc("/api/v1/providers/", s.handleProvider)
	mux.HandleFunc("/api/v1/profiles", s.handleProfiles)
	mux.HandleFunc("/api/v1/profiles/", s.handleProfile)
	mux.HandleFunc("/api/v1/logs", s.handleLogs)

	// Static files
	staticSub, _ := fs.Sub(staticFS, "static")
	fileServer := http.FileServer(http.FS(staticSub))
	mux.Handle("/", fileServer)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", config.WebPort),
		Handler: s.securityHeaders(mux),
	}

	return s
}

// Start begins listening. Returns an error if the port is already in use.
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("port %d is already in use: %w", config.WebPort, err)
	}
	s.logger.Printf("Web server listening on %s", s.httpServer.Addr)
	return s.httpServer.Serve(ln)
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// securityHeaders adds security response headers.
func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

// --- health & reload ---

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"version": s.version,
	})
}

func (s *Server) handleReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	store := config.DefaultStore()
	if err := store.Reload(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "reloaded"})
}

// --- helpers ---

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func readJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

// maskToken masks an auth token for display: "sk-abc...xyz" style.
func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:5] + "..." + token[len(token)-4:]
}

// WaitForReady polls the health endpoint until the server is ready or ctx is cancelled.
func WaitForReady(ctx context.Context) error {
	url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/health", config.WebPort)
	client := &http.Client{Timeout: 500 * time.Millisecond}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// --- logs ---

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	filter := proxy.LogFilter{
		Provider: query.Get("provider"),
	}

	if query.Get("errors_only") == "true" {
		filter.ErrorsOnly = true
	}

	if level := query.Get("level"); level != "" {
		filter.Level = proxy.LogLevel(level)
	}

	if statusCode := query.Get("status_code"); statusCode != "" {
		if code, err := strconv.Atoi(statusCode); err == nil {
			filter.StatusCode = code
		}
	}

	if statusMin := query.Get("status_min"); statusMin != "" {
		if code, err := strconv.Atoi(statusMin); err == nil {
			filter.StatusMin = code
		}
	}

	if statusMax := query.Get("status_max"); statusMax != "" {
		if code, err := strconv.Atoi(statusMax); err == nil {
			filter.StatusMax = code
		}
	}

	if limit := query.Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = l
		}
	}

	if filter.Limit <= 0 {
		filter.Limit = 100 // default limit
	}

	// Try in-memory logger first (same process as proxy), fall back to file
	var entries []proxy.LogEntry
	var providers []string

	logger := proxy.GetGlobalLogger()
	if logger != nil && logger.HasEntries() {
		entries = logger.GetEntries(filter)
		providers = logger.GetProviders()
	} else {
		// Read from JSON log file (web server running in separate process)
		logDir := config.ConfigDirPath()
		var err error
		entries, providers, err = proxy.ReadEntriesFromFile(logDir, filter)
		if err != nil {
			s.logger.Printf("Failed to read log file: %v", err)
			entries = []proxy.LogEntry{}
			providers = []string{}
		}
	}

	writeJSON(w, http.StatusOK, proxy.LogsResponse{
		Entries:   entries,
		Total:     len(entries),
		Providers: providers,
	})
}
