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
	Name           string
	Type           string // "anthropic" or "openai"
	BaseURL        *url.URL
	Token          string
	Model          string
	ReasoningModel string
	HaikuModel     string
	OpusModel      string
	SonnetModel    string
	EnvVars        map[string]string
	Healthy        bool
	AuthFailed     bool
	FailedAt       time.Time
	Backoff        time.Duration
	mu             sync.Mutex
}

// GetType returns the provider type, defaulting to "anthropic".
func (p *Provider) GetType() string {
	if p.Type == "" {
		return config.ProviderTypeAnthropic
	}
	return p.Type
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
