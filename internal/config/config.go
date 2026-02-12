package config

import "os"

const (
	ConfigDir  = ".opencc"
	ConfigFile = "opencc.json"
	LegacyDir  = ".cc_envs"
)

// ProviderConfig holds connection and model settings for a single API provider.
type ProviderConfig struct {
	BaseURL        string `json:"base_url"`
	AuthToken      string `json:"auth_token"`
	Model          string `json:"model,omitempty"`
	ReasoningModel string `json:"reasoning_model,omitempty"`
	HaikuModel     string `json:"haiku_model,omitempty"`
	OpusModel      string `json:"opus_model,omitempty"`
	SonnetModel    string `json:"sonnet_model,omitempty"`
}

// ExportToEnv sets all ANTHROPIC_* environment variables from this provider config.
func (p *ProviderConfig) ExportToEnv() {
	os.Setenv("ANTHROPIC_BASE_URL", p.BaseURL)
	os.Setenv("ANTHROPIC_AUTH_TOKEN", p.AuthToken)
	if p.Model != "" {
		os.Setenv("ANTHROPIC_MODEL", p.Model)
	}
	if p.ReasoningModel != "" {
		os.Setenv("ANTHROPIC_REASONING_MODEL", p.ReasoningModel)
	}
	if p.HaikuModel != "" {
		os.Setenv("ANTHROPIC_DEFAULT_HAIKU_MODEL", p.HaikuModel)
	}
	if p.OpusModel != "" {
		os.Setenv("ANTHROPIC_DEFAULT_OPUS_MODEL", p.OpusModel)
	}
	if p.SonnetModel != "" {
		os.Setenv("ANTHROPIC_DEFAULT_SONNET_MODEL", p.SonnetModel)
	}
}

// OpenCCConfig is the top-level configuration structure stored in opencc.json.
type OpenCCConfig struct {
	Providers map[string]*ProviderConfig `json:"providers"`
	Profiles  map[string][]string        `json:"profiles"`
}
