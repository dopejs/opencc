package config

import (
	"encoding/json"
	"os"
)

const (
	ConfigDir  = ".opencc"
	ConfigFile = "opencc.json"
	LegacyDir  = ".cc_envs"

	DefaultWebPort = 19840
	WebPidFile     = "web.pid"
	WebLogFile     = "web.log"

	DefaultProfileName = "default"
	DefaultCLIName     = "claude"

	// Provider API types
	ProviderTypeAnthropic = "anthropic"
	ProviderTypeOpenAI    = "openai"
)

// ProviderConfig holds connection and model settings for a single API provider.
type ProviderConfig struct {
	Type           string            `json:"type,omitempty"` // "anthropic" (default) or "openai"
	BaseURL        string            `json:"base_url"`
	AuthToken      string            `json:"auth_token"`
	Model          string            `json:"model,omitempty"`
	ReasoningModel string            `json:"reasoning_model,omitempty"`
	HaikuModel     string            `json:"haiku_model,omitempty"`
	OpusModel      string            `json:"opus_model,omitempty"`
	SonnetModel    string            `json:"sonnet_model,omitempty"`
	EnvVars        map[string]string `json:"env_vars,omitempty"`          // Claude Code env vars (legacy, for backward compat)
	ClaudeEnvVars  map[string]string `json:"claude_env_vars,omitempty"`   // Claude Code specific env vars
	CodexEnvVars   map[string]string `json:"codex_env_vars,omitempty"`    // Codex specific env vars
	OpenCodeEnvVars map[string]string `json:"opencode_env_vars,omitempty"` // OpenCode specific env vars
}

// GetType returns the provider type, defaulting to "anthropic".
func (p *ProviderConfig) GetType() string {
	if p.Type == "" {
		return ProviderTypeAnthropic
	}
	return p.Type
}

// GetEnvVarsForCLI returns the environment variables for a specific CLI.
// Falls back to legacy EnvVars if CLI-specific vars are not set.
func (p *ProviderConfig) GetEnvVarsForCLI(cli string) map[string]string {
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
	// Fallback to legacy EnvVars
	return p.EnvVars
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

	// Export custom environment variables
	for k, v := range p.EnvVars {
		if k != "" && v != "" {
			os.Setenv(k, v)
		}
	}
}

// Scenario represents a request scenario for routing decisions.
type Scenario string

const (
	ScenarioThink       Scenario = "think"
	ScenarioImage       Scenario = "image"
	ScenarioLongContext Scenario = "longContext"
	ScenarioWebSearch   Scenario = "webSearch"
	ScenarioBackground  Scenario = "background"
	ScenarioDefault     Scenario = "default"
)

// ProviderRoute defines a provider and its optional model override in a scenario.
type ProviderRoute struct {
	Name  string `json:"name"`
	Model string `json:"model,omitempty"`
}

// ScenarioRoute defines providers and their model overrides for a scenario.
type ScenarioRoute struct {
	Providers []*ProviderRoute `json:"providers"`
}

// UnmarshalJSON supports both old format (providers: ["p1"], model: "m") and new format (providers: [{name, model}]).
func (sr *ScenarioRoute) UnmarshalJSON(data []byte) error {
	// Try new format first
	type scenarioRouteAlias struct {
		Providers []*ProviderRoute `json:"providers"`
	}
	var alias scenarioRouteAlias
	if err := json.Unmarshal(data, &alias); err == nil && len(alias.Providers) > 0 {
		// Check if first provider is actually a ProviderRoute (has Name field)
		if alias.Providers[0].Name != "" {
			sr.Providers = alias.Providers
			return nil
		}
	}

	// Try old format: {providers: ["p1", "p2"], model: "m"}
	var oldFormat struct {
		Providers []string `json:"providers"`
		Model     string   `json:"model,omitempty"`
	}
	if err := json.Unmarshal(data, &oldFormat); err != nil {
		return err
	}

	// Convert old format to new
	sr.Providers = make([]*ProviderRoute, len(oldFormat.Providers))
	for i, name := range oldFormat.Providers {
		sr.Providers[i] = &ProviderRoute{
			Name:  name,
			Model: oldFormat.Model, // All providers share the same model in old format
		}
	}
	return nil
}

// ProviderNames returns the list of provider names in order.
func (sr *ScenarioRoute) ProviderNames() []string {
	names := make([]string, len(sr.Providers))
	for i, pr := range sr.Providers {
		names[i] = pr.Name
	}
	return names
}

// ModelForProvider returns the model override for a specific provider, or empty string.
func (sr *ScenarioRoute) ModelForProvider(name string) string {
	for _, pr := range sr.Providers {
		if pr.Name == name {
			return pr.Model
		}
	}
	return ""
}

// ProfileConfig holds a profile's provider list and optional scenario routing.
type ProfileConfig struct {
	Providers            []string                    `json:"providers"`
	Routing              map[Scenario]*ScenarioRoute `json:"routing,omitempty"`
	LongContextThreshold int                         `json:"long_context_threshold,omitempty"` // defaults to 32000 if not set
}

// UnmarshalJSON supports both old format (["p1","p2"]) and new format ({providers: [...], routing: {...}}).
func (pc *ProfileConfig) UnmarshalJSON(data []byte) error {
	// Trim whitespace to check first character
	for _, b := range data {
		if b == ' ' || b == '\t' || b == '\n' || b == '\r' {
			continue
		}
		if b == '[' {
			// Old format: plain string array
			var providers []string
			if err := json.Unmarshal(data, &providers); err != nil {
				return err
			}
			pc.Providers = providers
			pc.Routing = nil
			return nil
		}
		break
	}

	// New format: object with providers and optional routing
	type profileConfigAlias ProfileConfig
	var alias profileConfigAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*pc = ProfileConfig(alias)
	return nil
}

// Config version history:
// - Version 1 (implicit, no version field): profiles as string arrays
// - Version 2 (v1.3.2+): profiles as objects with routing support
// - Version 3 (v1.4.0+): project bindings support
// - Version 4 (v1.5.0+): default profile and web port settings
const CurrentConfigVersion = 4

// OpenCCConfig is the top-level configuration structure stored in opencc.json.
type OpenCCConfig struct {
	Version         int                        `json:"version,omitempty"`          // config file version
	DefaultProfile  string                     `json:"default_profile,omitempty"`  // default profile name (defaults to "default")
	DefaultCLI      string                     `json:"default_cli,omitempty"`      // default CLI (claude, codex, opencode)
	WebPort         int                        `json:"web_port,omitempty"`         // web UI port (defaults to 19841)
	Providers       map[string]*ProviderConfig `json:"providers"`                  // provider configurations
	Profiles        map[string]*ProfileConfig  `json:"profiles"`                   // profile configurations
	ProjectBindings map[string]string          `json:"project_bindings,omitempty"` // directory path -> profile name
}
