package config

import (
	"os"
	"path/filepath"
	"testing"
)

func setupLegacyDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	ResetDefaultStore()
	t.Cleanup(func() { ResetDefaultStore() })
	legacyDir := filepath.Join(dir, LegacyDir)
	os.MkdirAll(legacyDir, 0755)
	return dir
}

func writeLegacyEnv(t *testing.T, home, name, content string) {
	t.Helper()
	path := filepath.Join(home, LegacyDir, name+".env")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func writeLegacyConf(t *testing.T, home, filename, content string) {
	t.Helper()
	path := filepath.Join(home, LegacyDir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestMigrateFromLegacyFull(t *testing.T) {
	home := setupLegacyDir(t)

	writeLegacyEnv(t, home, "work", "ANTHROPIC_BASE_URL=https://work.com\nANTHROPIC_AUTH_TOKEN=tok1\nANTHROPIC_MODEL=opus\n")
	writeLegacyEnv(t, home, "backup", "ANTHROPIC_BASE_URL=https://backup.com\nANTHROPIC_AUTH_TOKEN=tok2\n")
	writeLegacyConf(t, home, "fallback.conf", "work\nbackup\n")
	writeLegacyConf(t, home, "fallback.staging.conf", "backup\n")

	cfg, err := MigrateFromLegacy()
	if err != nil {
		t.Fatalf("MigrateFromLegacy() error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}

	// Check providers
	if len(cfg.Providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(cfg.Providers))
	}

	work := cfg.Providers["work"]
	if work == nil {
		t.Fatal("missing provider 'work'")
	}
	if work.BaseURL != "https://work.com" {
		t.Errorf("work.BaseURL = %q", work.BaseURL)
	}
	if work.AuthToken != "tok1" {
		t.Errorf("work.AuthToken = %q", work.AuthToken)
	}
	if work.Model != "opus" {
		t.Errorf("work.Model = %q", work.Model)
	}

	backup := cfg.Providers["backup"]
	if backup == nil {
		t.Fatal("missing provider 'backup'")
	}
	if backup.BaseURL != "https://backup.com" {
		t.Errorf("backup.BaseURL = %q", backup.BaseURL)
	}

	// Check profiles
	if len(cfg.Profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(cfg.Profiles))
	}

	defaultProfile := cfg.Profiles["default"]
	if len(defaultProfile) != 2 || defaultProfile[0] != "work" || defaultProfile[1] != "backup" {
		t.Errorf("default profile = %v", defaultProfile)
	}

	stagingProfile := cfg.Profiles["staging"]
	if len(stagingProfile) != 1 || stagingProfile[0] != "backup" {
		t.Errorf("staging profile = %v", stagingProfile)
	}
}

func TestMigrateFromLegacyNoLegacy(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	ResetDefaultStore()
	t.Cleanup(func() { ResetDefaultStore() })

	// No .cc_envs directory at all
	cfg, err := MigrateFromLegacy()
	if err != nil {
		t.Fatalf("MigrateFromLegacy() error: %v", err)
	}
	if cfg != nil {
		t.Error("expected nil config when no legacy dir")
	}
}

func TestMigrateFromLegacyPartial(t *testing.T) {
	home := setupLegacyDir(t)

	writeLegacyEnv(t, home, "only", "ANTHROPIC_BASE_URL=https://only.com\nANTHROPIC_AUTH_TOKEN=tok\n")
	// No fallback.conf

	cfg, err := MigrateFromLegacy()
	if err != nil {
		t.Fatalf("MigrateFromLegacy() error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if len(cfg.Providers) != 1 {
		t.Errorf("expected 1 provider, got %d", len(cfg.Providers))
	}
	if len(cfg.Profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(cfg.Profiles))
	}
}

func TestMigrateIdempotent(t *testing.T) {
	home := setupLegacyDir(t)

	writeLegacyEnv(t, home, "x", "ANTHROPIC_BASE_URL=https://x.com\nANTHROPIC_AUTH_TOKEN=tok\n")
	writeLegacyConf(t, home, "fallback.conf", "x\n")

	// Load via Store (triggers migration)
	s := &Store{path: filepath.Join(home, ConfigDir, ConfigFile)}
	if err := s.Load(); err != nil {
		t.Fatalf("first Load() error: %v", err)
	}

	// Verify JSON was written
	jsonPath := filepath.Join(home, ConfigDir, ConfigFile)
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		t.Fatal("JSON config should exist after migration")
	}

	// Load again — should read JSON, not re-migrate
	s2 := &Store{path: jsonPath}
	if err := s2.Load(); err != nil {
		t.Fatalf("second Load() error: %v", err)
	}

	p := s2.GetProvider("x")
	if p == nil {
		t.Fatal("provider 'x' should exist after reload")
	}
	if p.BaseURL != "https://x.com" {
		t.Errorf("BaseURL = %q", p.BaseURL)
	}
}

func TestMigrateFromLegacySkipsComments(t *testing.T) {
	home := setupLegacyDir(t)

	writeLegacyEnv(t, home, "test", "# comment\nANTHROPIC_BASE_URL=https://test.com\n\nANTHROPIC_AUTH_TOKEN=tok\nANTHROPIC_MODEL=opus # inline\n")
	writeLegacyConf(t, home, "fallback.conf", "# comment\ntest\n\n# another\n")

	cfg, err := MigrateFromLegacy()
	if err != nil {
		t.Fatalf("MigrateFromLegacy() error: %v", err)
	}

	p := cfg.Providers["test"]
	if p == nil {
		t.Fatal("expected provider 'test'")
	}
	if p.BaseURL != "https://test.com" {
		t.Errorf("BaseURL = %q", p.BaseURL)
	}
	if p.Model != "opus" {
		t.Errorf("Model = %q (inline comment should be stripped)", p.Model)
	}

	defaultProfile := cfg.Profiles["default"]
	if len(defaultProfile) != 1 || defaultProfile[0] != "test" {
		t.Errorf("default profile = %v", defaultProfile)
	}
}

func TestMigrateFromLegacyEmptyDir(t *testing.T) {
	setupLegacyDir(t)
	// Empty .cc_envs dir — no .env files

	cfg, err := MigrateFromLegacy()
	if err != nil {
		t.Fatalf("MigrateFromLegacy() error: %v", err)
	}
	if cfg != nil {
		t.Error("expected nil config for empty legacy dir")
	}
}

func TestMigrateFromLegacyAllModelFields(t *testing.T) {
	home := setupLegacyDir(t)

	content := "ANTHROPIC_BASE_URL=https://x.com\n" +
		"ANTHROPIC_AUTH_TOKEN=tok\n" +
		"ANTHROPIC_MODEL=m1\n" +
		"ANTHROPIC_REASONING_MODEL=m2\n" +
		"ANTHROPIC_DEFAULT_HAIKU_MODEL=m3\n" +
		"ANTHROPIC_DEFAULT_OPUS_MODEL=m4\n" +
		"ANTHROPIC_DEFAULT_SONNET_MODEL=m5\n"
	writeLegacyEnv(t, home, "full", content)

	cfg, err := MigrateFromLegacy()
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	p := cfg.Providers["full"]
	if p.Model != "m1" || p.ReasoningModel != "m2" || p.HaikuModel != "m3" || p.OpusModel != "m4" || p.SonnetModel != "m5" {
		t.Errorf("models not migrated correctly: %+v", p)
	}
}
