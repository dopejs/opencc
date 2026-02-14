package proxy

import (
	"net/url"
	"sync"
	"time"

	"github.com/dopejs/opencc/internal/config"
)

const (
	InitialBackoff     = 60 * time.Second
	MaxBackoff         = 5 * time.Minute
	AuthInitialBackoff = 30 * time.Minute
	AuthMaxBackoff     = 2 * time.Hour
)

type Provider struct {
	Name            string
	Type            string // "anthropic" or "openai"
	BaseURL         *url.URL
	Token           string
	Model           string
	ReasoningModel  string
	HaikuModel      string
	OpusModel       string
	SonnetModel     string
	EnvVars         map[string]string // Legacy env vars (for backward compat)
	ClaudeEnvVars   map[string]string // Claude Code specific
	CodexEnvVars    map[string]string // Codex specific
	OpenCodeEnvVars map[string]string // OpenCode specific
	Healthy         bool
	AuthFailed      bool
	FailedAt        time.Time
	Backoff         time.Duration
	mu              sync.Mutex
}

// GetType returns the provider type, defaulting to "anthropic".
func (p *Provider) GetType() string {
	if p.Type == "" {
		return config.ProviderTypeAnthropic
	}
	return p.Type
}

// GetEnvVarsForCLI returns the environment variables for a specific CLI.
func (p *Provider) GetEnvVarsForCLI(cli string) map[string]string {
	switch cli {
	case "codex":
		if len(p.CodexEnvVars) > 0 {
			return p.CodexEnvVars
		}
	case "opencode":
		if len(p.OpenCodeEnvVars) > 0 {
			return p.OpenCodeEnvVars
		}
	default: // claude
		if len(p.ClaudeEnvVars) > 0 {
			return p.ClaudeEnvVars
		}
	}
	return p.EnvVars
}

func (p *Provider) IsHealthy() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Healthy {
		return true
	}
	if time.Since(p.FailedAt) >= p.Backoff {
		p.Healthy = true
		return true
	}
	return false
}

func (p *Provider) MarkFailed() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Healthy = false
	p.FailedAt = time.Now()
	if p.Backoff == 0 {
		p.Backoff = InitialBackoff
	} else {
		p.Backoff *= 2
		if p.Backoff > MaxBackoff {
			p.Backoff = MaxBackoff
		}
	}
}

func (p *Provider) MarkAuthFailed() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Healthy = false
	p.AuthFailed = true
	p.FailedAt = time.Now()
	if p.Backoff < AuthInitialBackoff {
		p.Backoff = AuthInitialBackoff
	} else {
		p.Backoff *= 2
		if p.Backoff > AuthMaxBackoff {
			p.Backoff = AuthMaxBackoff
		}
	}
}

func (p *Provider) MarkHealthy() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Healthy = true
	p.AuthFailed = false
	p.Backoff = 0
}
