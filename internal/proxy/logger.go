package proxy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogLevel represents the severity of a log entry.
type LogLevel string

const (
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// LogEntry represents a structured log entry.
type LogEntry struct {
	Timestamp    time.Time `json:"timestamp"`
	Level        LogLevel  `json:"level"`
	Provider     string    `json:"provider,omitempty"`
	Message      string    `json:"message"`
	StatusCode   int       `json:"status_code,omitempty"`
	Method       string    `json:"method,omitempty"`
	Path         string    `json:"path,omitempty"`
	Error        string    `json:"error,omitempty"`
	ResponseBody string    `json:"response_body,omitempty"`
}

// StructuredLogger provides structured logging with separate error log file.
type StructuredLogger struct {
	mu          sync.Mutex
	logFile     *os.File
	errLogFile  *os.File
	jsonLogFile *os.File // JSON format log for web API
	logDir      string
	entries     []LogEntry
	maxEntries  int
}

// NewStructuredLogger creates a new structured logger.
func NewStructuredLogger(logDir string, maxEntries int) (*StructuredLogger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	logPath := filepath.Join(logDir, "proxy.log")
	errLogPath := filepath.Join(logDir, "err.log")
	jsonLogPath := filepath.Join(logDir, "proxy.jsonl")

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	errLogFile, err := os.OpenFile(errLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logFile.Close()
		return nil, fmt.Errorf("open error log file: %w", err)
	}

	jsonLogFile, err := os.OpenFile(jsonLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logFile.Close()
		errLogFile.Close()
		return nil, fmt.Errorf("open json log file: %w", err)
	}

	if maxEntries <= 0 {
		maxEntries = 1000
	}

	return &StructuredLogger{
		logFile:     logFile,
		errLogFile:  errLogFile,
		jsonLogFile: jsonLogFile,
		logDir:      logDir,
		entries:     make([]LogEntry, 0, maxEntries),
		maxEntries:  maxEntries,
	}, nil
}

// Close closes the log files.
func (l *StructuredLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var errs []error
	if l.logFile != nil {
		if err := l.logFile.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if l.errLogFile != nil {
		if err := l.errLogFile.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if l.jsonLogFile != nil {
		if err := l.jsonLogFile.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// Log writes a log entry.
func (l *StructuredLogger) Log(entry LogEntry) {
	entry.Timestamp = time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Add to in-memory buffer
	if len(l.entries) >= l.maxEntries {
		// Remove oldest entries (keep last 80%)
		keep := l.maxEntries * 8 / 10
		copy(l.entries, l.entries[len(l.entries)-keep:])
		l.entries = l.entries[:keep]
	}
	l.entries = append(l.entries, entry)

	// Write to log file (human-readable format)
	line := l.formatEntry(entry)
	if l.logFile != nil {
		l.logFile.WriteString(line + "\n")
	}

	// Write to JSON log file (for web API)
	if l.jsonLogFile != nil {
		if jsonLine, err := json.Marshal(entry); err == nil {
			l.jsonLogFile.WriteString(string(jsonLine) + "\n")
		}
	}

	// Write errors to err.log
	if entry.Level == LogLevelError || entry.Level == LogLevelWarn {
		if l.errLogFile != nil {
			l.errLogFile.WriteString(line + "\n")
		}
	}
}

// formatEntry formats a log entry as a string.
func (l *StructuredLogger) formatEntry(entry LogEntry) string {
	ts := entry.Timestamp.Format("2006/01/02 15:04:05")
	level := string(entry.Level)

	var msg string
	if entry.Provider != "" {
		msg = fmt.Sprintf("[%s] [%s] [%s]", ts, level, entry.Provider)
	} else {
		msg = fmt.Sprintf("[%s] [%s]", ts, level)
	}

	if entry.Method != "" && entry.Path != "" {
		msg += fmt.Sprintf(" %s %s", entry.Method, entry.Path)
	}

	if entry.StatusCode > 0 {
		msg += fmt.Sprintf(" status=%d", entry.StatusCode)
	}

	msg += " " + entry.Message

	if entry.Error != "" {
		msg += " error=" + entry.Error
	}

	if entry.ResponseBody != "" {
		msg += " response=" + entry.ResponseBody
	}

	return msg
}

// Info logs an info message.
func (l *StructuredLogger) Info(provider, message string) {
	l.Log(LogEntry{
		Level:    LogLevelInfo,
		Provider: provider,
		Message:  message,
	})
}

// Warn logs a warning message.
func (l *StructuredLogger) Warn(provider, message string) {
	l.Log(LogEntry{
		Level:    LogLevelWarn,
		Provider: provider,
		Message:  message,
	})
}

// Error logs an error message.
func (l *StructuredLogger) Error(provider, message string) {
	l.Log(LogEntry{
		Level:    LogLevelError,
		Provider: provider,
		Message:  message,
	})
}

// RequestLog logs a request with status code.
func (l *StructuredLogger) RequestLog(provider, method, path string, statusCode int, message string) {
	level := LogLevelInfo
	if statusCode >= 400 && statusCode < 500 {
		level = LogLevelWarn
	} else if statusCode >= 500 {
		level = LogLevelError
	}

	l.Log(LogEntry{
		Level:      level,
		Provider:   provider,
		Method:     method,
		Path:       path,
		StatusCode: statusCode,
		Message:    message,
	})
}

// RequestError logs a request error.
func (l *StructuredLogger) RequestError(provider, method, path string, err error) {
	l.Log(LogEntry{
		Level:    LogLevelError,
		Provider: provider,
		Method:   method,
		Path:     path,
		Message:  "request failed",
		Error:    err.Error(),
	})
}

// RequestErrorWithResponse logs a request error with response details.
func (l *StructuredLogger) RequestErrorWithResponse(provider, method, path string, statusCode int, message string, responseBody []byte) {
	// Truncate response body if too long
	bodyStr := string(responseBody)
	if len(bodyStr) > 500 {
		bodyStr = bodyStr[:500] + "..."
	}

	l.Log(LogEntry{
		Level:        LogLevelError,
		Provider:     provider,
		Method:       method,
		Path:         path,
		StatusCode:   statusCode,
		Message:      message,
		ResponseBody: bodyStr,
	})
}

// HasEntries returns true if the in-memory log buffer has entries.
func (l *StructuredLogger) HasEntries() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.entries) > 0
}

// GetEntries returns log entries matching the filter criteria.
func (l *StructuredLogger) GetEntries(filter LogFilter) []LogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	var result []LogEntry
	for _, entry := range l.entries {
		if filter.Match(entry) {
			result = append(result, entry)
		}
	}

	// Return in reverse order (newest first)
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	// Apply limit
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}

	return result
}

// GetProviders returns a list of unique provider names from the logs.
func (l *StructuredLogger) GetProviders() []string {
	l.mu.Lock()
	defer l.mu.Unlock()

	seen := make(map[string]bool)
	var providers []string
	for _, entry := range l.entries {
		if entry.Provider != "" && !seen[entry.Provider] {
			seen[entry.Provider] = true
			providers = append(providers, entry.Provider)
		}
	}
	return providers
}

// LogFilter defines criteria for filtering log entries.
type LogFilter struct {
	Provider   string   `json:"provider,omitempty"`
	Level      LogLevel `json:"level,omitempty"`      // empty means all levels
	ErrorsOnly bool     `json:"errors_only,omitempty"` // only error and warn levels
	StatusCode int      `json:"status_code,omitempty"` // filter by specific status code
	StatusMin  int      `json:"status_min,omitempty"`  // filter by status code range (min)
	StatusMax  int      `json:"status_max,omitempty"`  // filter by status code range (max)
	Limit      int      `json:"limit,omitempty"`       // max entries to return
}

// Match checks if a log entry matches the filter criteria.
func (f LogFilter) Match(entry LogEntry) bool {
	// Provider filter
	if f.Provider != "" && entry.Provider != f.Provider {
		return false
	}

	// Level filter
	if f.Level != "" && entry.Level != f.Level {
		return false
	}

	// Errors only filter
	if f.ErrorsOnly && entry.Level != LogLevelError && entry.Level != LogLevelWarn {
		return false
	}

	// Status code filter
	if f.StatusCode > 0 && entry.StatusCode != f.StatusCode {
		return false
	}

	// Status code range filter
	if f.StatusMin > 0 && entry.StatusCode < f.StatusMin {
		return false
	}
	if f.StatusMax > 0 && entry.StatusCode > f.StatusMax {
		return false
	}

	return true
}

// LogsResponse is the API response for log queries.
type LogsResponse struct {
	Entries   []LogEntry `json:"entries"`
	Total     int        `json:"total"`
	Providers []string   `json:"providers"`
}

// ToJSON serializes a log entry to JSON.
func (e LogEntry) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// ReadEntriesFromFile reads log entries from the JSON log file.
// This is used by the web server to read logs from a different process.
func ReadEntriesFromFile(logDir string, filter LogFilter) ([]LogEntry, []string, error) {
	jsonLogPath := filepath.Join(logDir, "proxy.jsonl")

	file, err := os.Open(jsonLogPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []LogEntry{}, []string{}, nil
		}
		return nil, nil, err
	}
	defer file.Close()

	var entries []LogEntry
	providerSet := make(map[string]bool)

	scanner := bufio.NewScanner(file)
	// Increase buffer size for long lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		var entry LogEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue // Skip malformed lines
		}

		if entry.Provider != "" {
			providerSet[entry.Provider] = true
		}

		if filter.Match(entry) {
			entries = append(entries, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	// Collect providers
	var providers []string
	for p := range providerSet {
		providers = append(providers, p)
	}

	// Return in reverse order (newest first)
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	// Apply limit
	if filter.Limit > 0 && len(entries) > filter.Limit {
		entries = entries[:filter.Limit]
	}

	return entries, providers, nil
}
