package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// --- Path helpers ---

// ConfigDirPath returns ~/.opencc
func ConfigDirPath() string {
	return filepath.Join(os.Getenv("HOME"), ConfigDir)
}

// ConfigFilePath returns ~/.opencc/opencc.json
func ConfigFilePath() string {
	return filepath.Join(ConfigDirPath(), ConfigFile)
}

// LogPath returns ~/.opencc/proxy.log
func LogPath() string {
	return filepath.Join(ConfigDirPath(), "proxy.log")
}

// legacyDirPath returns ~/.cc_envs
func legacyDirPath() string {
	return filepath.Join(os.Getenv("HOME"), LegacyDir)
}

// --- Store ---

// Store manages reading and writing the unified JSON config.
type Store struct {
	mu       sync.Mutex
	path     string
	config   *OpenCCConfig
	modTime  time.Time // last known modification time of config file
}

var (
	defaultStore *Store
	defaultOnce  sync.Once
	defaultMu    sync.Mutex
)

// DefaultStore returns the global Store singleton.
// On first call it loads from disk (with legacy migration if needed).
// On subsequent calls, it checks if the config file has been modified
// and reloads if necessary.
func DefaultStore() *Store {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	if defaultStore == nil {
		defaultStore = &Store{path: ConfigFilePath()}
		defaultStore.Load()
	} else {
		// Check if config file has been modified since last load
		if info, err := os.Stat(defaultStore.path); err == nil {
			if info.ModTime().After(defaultStore.modTime) {
				// File has been modified, reload
				defaultStore.Load()
			}
		}
	}
	return defaultStore
}

// ResetDefaultStore clears the singleton so the next DefaultStore() call
// re-initializes. Intended for tests.
func ResetDefaultStore() {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	defaultStore = nil
}

// --- Provider operations ---

// GetProvider returns the config for a named provider, or nil.
func (s *Store) GetProvider(name string) *ProviderConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.config == nil {
		return nil
	}
	return s.config.Providers[name]
}

// SetProvider creates or updates a provider and saves.
func (s *Store) SetProvider(name string, p *ProviderConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureConfig()
	s.config.Providers[name] = p
	return s.saveLocked()
}

// DeleteProvider removes a provider and removes it from all profiles
// (including routing scenarios), then saves.
func (s *Store) DeleteProvider(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureConfig()
	delete(s.config.Providers, name)
	for _, pc := range s.config.Profiles {
		pc.Providers = removeString(pc.Providers, name)
		for scenario, route := range pc.Routing {
			route.Providers = filterProviderRoutes(route.Providers, name)
			if len(route.Providers) == 0 {
				delete(pc.Routing, scenario)
			}
		}
	}
	return s.saveLocked()
}

// ProviderNames returns sorted provider names.
func (s *Store) ProviderNames() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.config == nil {
		return nil
	}
	names := make([]string, 0, len(s.config.Providers))
	for n := range s.config.Providers {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// ProviderMap returns all providers.
func (s *Store) ProviderMap() map[string]*ProviderConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.config == nil {
		return nil
	}
	return s.config.Providers
}

// ExportProviderToEnv sets ANTHROPIC_* env vars for the named provider.
func (s *Store) ExportProviderToEnv(name string) error {
	p := s.GetProvider(name)
	if p == nil {
		return fmt.Errorf("provider %q not found", name)
	}
	p.ExportToEnv()
	return nil
}

// --- Profile operations ---

// GetProfileOrder returns the provider list for a profile.
func (s *Store) GetProfileOrder(profile string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.config == nil {
		return nil
	}
	pc := s.config.Profiles[profile]
	if pc == nil {
		return nil
	}
	return pc.Providers
}

// SetProfileOrder sets the provider list for a profile and saves.
// Preserves existing routing configuration if the profile already exists.
func (s *Store) SetProfileOrder(profile string, names []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureConfig()
	if names == nil {
		names = []string{}
	}
	pc := s.config.Profiles[profile]
	if pc == nil {
		pc = &ProfileConfig{}
		s.config.Profiles[profile] = pc
	}
	pc.Providers = names
	return s.saveLocked()
}

// GetProfileConfig returns the full profile configuration.
func (s *Store) GetProfileConfig(profile string) *ProfileConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.config == nil {
		return nil
	}
	return s.config.Profiles[profile]
}

// SetProfileConfig sets the full profile configuration and saves.
func (s *Store) SetProfileConfig(profile string, pc *ProfileConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureConfig()
	if pc == nil {
		pc = &ProfileConfig{Providers: []string{}}
	}
	s.config.Profiles[profile] = pc
	return s.saveLocked()
}

// RemoveFromProfile removes a provider name from a specific profile
// (both from the main providers list and from all routing scenarios) and saves.
func (s *Store) RemoveFromProfile(profile, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureConfig()
	pc := s.config.Profiles[profile]
	if pc == nil {
		return nil
	}
	pc.Providers = removeString(pc.Providers, name)
	for scenario, route := range pc.Routing {
		route.Providers = filterProviderRoutes(route.Providers, name)
		if len(route.Providers) == 0 {
			delete(pc.Routing, scenario)
		}
	}
	return s.saveLocked()
}

// DeleteProfile deletes a profile. Cannot delete "default".
func (s *Store) DeleteProfile(profile string) error {
	if profile == "" || profile == "default" {
		return fmt.Errorf("cannot delete the default profile")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureConfig()
	delete(s.config.Profiles, profile)
	return s.saveLocked()
}

// ListProfiles returns sorted profile names.
func (s *Store) ListProfiles() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.config == nil {
		return nil
	}
	names := make([]string, 0, len(s.config.Profiles))
	for n := range s.config.Profiles {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// --- I/O ---

// Load reads the JSON config from disk. If the file doesn't exist, it tries
// to migrate from the legacy .cc_envs format. If neither exists, it creates
// an empty config.
func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err == nil {
		var cfg OpenCCConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("failed to parse %s: %w", s.path, err)
		}
		if cfg.Providers == nil {
			cfg.Providers = make(map[string]*ProviderConfig)
		}
		if cfg.Profiles == nil {
			cfg.Profiles = make(map[string]*ProfileConfig)
		}
		s.config = &cfg
		// Update modification time
		if info, statErr := os.Stat(s.path); statErr == nil {
			s.modTime = info.ModTime()
		}
		return nil
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("failed to read %s: %w", s.path, err)
	}

	// JSON doesn't exist — try legacy migration
	legacyDir := legacyDirPath()
	if info, statErr := os.Stat(legacyDir); statErr == nil && info.IsDir() {
		cfg, migrateErr := MigrateFromLegacy()
		if migrateErr != nil {
			return fmt.Errorf("migration failed: %w", migrateErr)
		}
		if cfg != nil {
			s.config = cfg
			return s.saveLocked()
		}
	}

	// Nothing exists — create empty config
	s.config = &OpenCCConfig{
		Providers: make(map[string]*ProviderConfig),
		Profiles:  make(map[string]*ProfileConfig),
	}
	s.modTime = time.Time{} // zero time for non-existent file
	return nil
}

// Save writes the config to disk atomically (temp + rename), with 0600 permissions.
func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked()
}

func (s *Store) saveLocked() error {
	s.ensureConfig()
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	data = append(data, '\n')

	tmp, err := os.CreateTemp(dir, "opencc-*.json")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("failed to chmod temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Rename(tmpName, s.path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("failed to rename config file: %w", err)
	}
	// Update modification time after successful save
	if info, statErr := os.Stat(s.path); statErr == nil {
		s.modTime = info.ModTime()
	}
	return nil
}

// Reload re-reads the config from disk.
func (s *Store) Reload() error {
	return s.Load()
}

// ensureConfig makes sure s.config is non-nil with initialized maps.
func (s *Store) ensureConfig() {
	if s.config == nil {
		s.config = &OpenCCConfig{
			Providers: make(map[string]*ProviderConfig),
			Profiles:  make(map[string]*ProfileConfig),
		}
	}
	if s.config.Providers == nil {
		s.config.Providers = make(map[string]*ProviderConfig)
	}
	if s.config.Profiles == nil {
		s.config.Profiles = make(map[string]*ProfileConfig)
	}
}

// --- helpers ---

func removeString(ss []string, s string) []string {
	var out []string
	for _, v := range ss {
		if v != s {
			out = append(out, v)
		}
	}
	return out
}

func filterProviderRoutes(routes []*ProviderRoute, name string) []*ProviderRoute {
	var out []*ProviderRoute
	for _, pr := range routes {
		if pr.Name != name {
			out = append(out, pr)
		}
	}
	return out
}
