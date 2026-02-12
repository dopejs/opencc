package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// MigrateFromLegacy reads the legacy ~/.cc_envs directory and converts
// *.env files and fallback*.conf files into an OpenCCConfig.
// Returns nil if the legacy directory has no .env files.
func MigrateFromLegacy() (*OpenCCConfig, error) {
	dir := legacyDirPath()

	// 1. Read all *.env files → providers
	envMatches, _ := filepath.Glob(filepath.Join(dir, "*.env"))
	if len(envMatches) == 0 {
		return nil, nil
	}

	providers := make(map[string]*ProviderConfig)
	for _, path := range envMatches {
		name := strings.TrimSuffix(filepath.Base(path), ".env")
		p, err := parseLegacyEnvFile(path)
		if err != nil {
			continue
		}
		providers[name] = p
	}

	if len(providers) == 0 {
		return nil, nil
	}

	// 2. Read fallback*.conf files → profiles
	profiles := make(map[string][]string)
	confMatches, _ := filepath.Glob(filepath.Join(dir, "fallback*.conf"))
	for _, path := range confMatches {
		base := filepath.Base(path)
		var profileName string
		switch {
		case base == "fallback.conf":
			profileName = "default"
		case strings.HasPrefix(base, "fallback.") && strings.HasSuffix(base, ".conf"):
			profileName = strings.TrimPrefix(base, "fallback.")
			profileName = strings.TrimSuffix(profileName, ".conf")
			if profileName == "" {
				continue
			}
		default:
			continue
		}

		names, err := parseLegacyConfFile(path)
		if err != nil {
			continue
		}
		if len(names) > 0 {
			profiles[profileName] = names
		}
	}

	return &OpenCCConfig{
		Providers: providers,
		Profiles:  profiles,
	}, nil
}

// parseLegacyEnvFile parses a key=value .env file into a ProviderConfig.
func parseLegacyEnvFile(path string) (*ProviderConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	kv := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Strip inline comments
		if idx := strings.Index(line, " #"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		kv[k] = v
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &ProviderConfig{
		BaseURL:        kv["ANTHROPIC_BASE_URL"],
		AuthToken:      kv["ANTHROPIC_AUTH_TOKEN"],
		Model:          kv["ANTHROPIC_MODEL"],
		ReasoningModel: kv["ANTHROPIC_REASONING_MODEL"],
		HaikuModel:     kv["ANTHROPIC_DEFAULT_HAIKU_MODEL"],
		OpusModel:      kv["ANTHROPIC_DEFAULT_OPUS_MODEL"],
		SonnetModel:    kv["ANTHROPIC_DEFAULT_SONNET_MODEL"],
	}, nil
}

// parseLegacyConfFile reads a fallback*.conf file (one provider name per line).
func parseLegacyConfFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var names []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		names = append(names, line)
	}
	return names, scanner.Err()
}
