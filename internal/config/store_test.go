package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func newTestStore(t *testing.T) (*Store, string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	ResetDefaultStore()
	t.Cleanup(func() { ResetDefaultStore() })
	path := filepath.Join(dir, ConfigDir, ConfigFile)
	s := &Store{path: path}
	return s, dir
}

func TestStoreLoadEmpty(t *testing.T) {
	s, _ := newTestStore(t)

	if err := s.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if s.config == nil {
		t.Fatal("config should not be nil")
	}
	if len(s.config.Providers) != 0 {
		t.Errorf("expected 0 providers, got %d", len(s.config.Providers))
	}
	if len(s.config.Profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(s.config.Profiles))
	}
}

func TestStoreLoadExisting(t *testing.T) {
	s, home := newTestStore(t)

	cfg := &OpenCCConfig{
		Providers: map[string]*ProviderConfig{
			"work": {BaseURL: "https://work.example.com", AuthToken: "tok1", Model: "opus"},
		},
		Profiles: map[string][]string{
			"default": {"work"},
		},
	}
	dir := filepath.Join(home, ConfigDir)
	os.MkdirAll(dir, 0755)
	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(filepath.Join(dir, ConfigFile), data, 0600)

	if err := s.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	p := s.GetProvider("work")
	if p == nil {
		t.Fatal("expected provider 'work'")
	}
	if p.BaseURL != "https://work.example.com" {
		t.Errorf("BaseURL = %q", p.BaseURL)
	}
	if p.Model != "opus" {
		t.Errorf("Model = %q", p.Model)
	}

	order := s.GetProfileOrder("default")
	if len(order) != 1 || order[0] != "work" {
		t.Errorf("default profile = %v", order)
	}
}

func TestStoreSaveAndReload(t *testing.T) {
	s, _ := newTestStore(t)
	s.Load()

	s.SetProvider("test", &ProviderConfig{
		BaseURL:   "https://test.com",
		AuthToken: "tok",
		Model:     "sonnet",
	})
	s.SetProfileOrder("default", []string{"test"})

	// Create a new store pointing to same path
	s2 := &Store{path: s.path}
	if err := s2.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	p := s2.GetProvider("test")
	if p == nil {
		t.Fatal("expected provider 'test' after reload")
	}
	if p.BaseURL != "https://test.com" {
		t.Errorf("BaseURL = %q", p.BaseURL)
	}
	if p.Model != "sonnet" {
		t.Errorf("Model = %q", p.Model)
	}

	order := s2.GetProfileOrder("default")
	if len(order) != 1 || order[0] != "test" {
		t.Errorf("default profile = %v", order)
	}
}

func TestStoreProviderCRUD(t *testing.T) {
	s, _ := newTestStore(t)
	s.Load()

	// Create
	s.SetProvider("a", &ProviderConfig{BaseURL: "https://a.com", AuthToken: "tok-a"})
	s.SetProvider("b", &ProviderConfig{BaseURL: "https://b.com", AuthToken: "tok-b"})

	names := s.ProviderNames()
	if len(names) != 2 || names[0] != "a" || names[1] != "b" {
		t.Errorf("ProviderNames() = %v", names)
	}

	// Read
	p := s.GetProvider("a")
	if p == nil || p.BaseURL != "https://a.com" {
		t.Errorf("GetProvider('a') = %+v", p)
	}

	// Update
	s.SetProvider("a", &ProviderConfig{BaseURL: "https://a2.com", AuthToken: "tok-a2"})
	p = s.GetProvider("a")
	if p == nil || p.BaseURL != "https://a2.com" {
		t.Errorf("after update, GetProvider('a') = %+v", p)
	}

	// Delete
	s.DeleteProvider("b")
	if s.GetProvider("b") != nil {
		t.Error("provider 'b' should be deleted")
	}
	names = s.ProviderNames()
	if len(names) != 1 {
		t.Errorf("expected 1 provider after delete, got %d", len(names))
	}
}

func TestStoreProfileCRUD(t *testing.T) {
	s, _ := newTestStore(t)
	s.Load()

	// Create profile
	s.SetProfileOrder("work", []string{"a", "b"})
	order := s.GetProfileOrder("work")
	if len(order) != 2 || order[0] != "a" || order[1] != "b" {
		t.Errorf("GetProfileOrder('work') = %v", order)
	}

	profiles := s.ListProfiles()
	if len(profiles) != 1 || profiles[0] != "work" {
		t.Errorf("ListProfiles() = %v", profiles)
	}

	// Update
	s.SetProfileOrder("work", []string{"b", "a", "c"})
	order = s.GetProfileOrder("work")
	if len(order) != 3 || order[0] != "b" {
		t.Errorf("after update, GetProfileOrder('work') = %v", order)
	}

	// Remove from profile
	s.RemoveFromProfile("work", "a")
	order = s.GetProfileOrder("work")
	if len(order) != 2 || order[0] != "b" || order[1] != "c" {
		t.Errorf("after remove, GetProfileOrder('work') = %v", order)
	}

	// Delete profile
	s.DeleteProfile("work")
	order = s.GetProfileOrder("work")
	if order != nil {
		t.Errorf("expected nil after delete, got %v", order)
	}
}

func TestStoreDeleteProfileDefault(t *testing.T) {
	s, _ := newTestStore(t)
	s.Load()

	s.SetProfileOrder("default", []string{"a"})

	err := s.DeleteProfile("default")
	if err == nil {
		t.Error("expected error when deleting default profile")
	}

	err = s.DeleteProfile("")
	if err == nil {
		t.Error("expected error when deleting empty profile name")
	}

	// Should still exist
	order := s.GetProfileOrder("default")
	if len(order) != 1 || order[0] != "a" {
		t.Errorf("default profile should still exist, got %v", order)
	}
}

func TestStoreDeleteProviderCascade(t *testing.T) {
	s, _ := newTestStore(t)
	s.Load()

	s.SetProvider("x", &ProviderConfig{BaseURL: "https://x.com", AuthToken: "tok"})
	s.SetProvider("y", &ProviderConfig{BaseURL: "https://y.com", AuthToken: "tok"})

	s.SetProfileOrder("default", []string{"x", "y"})
	s.SetProfileOrder("work", []string{"y", "x"})

	// Delete provider x — should be removed from all profiles
	s.DeleteProvider("x")

	defaultOrder := s.GetProfileOrder("default")
	if len(defaultOrder) != 1 || defaultOrder[0] != "y" {
		t.Errorf("default profile after cascade = %v", defaultOrder)
	}

	workOrder := s.GetProfileOrder("work")
	if len(workOrder) != 1 || workOrder[0] != "y" {
		t.Errorf("work profile after cascade = %v", workOrder)
	}
}

func TestStoreExportProviderToEnv(t *testing.T) {
	s, _ := newTestStore(t)
	s.Load()

	s.SetProvider("test", &ProviderConfig{
		BaseURL:   "https://test.com",
		AuthToken: "tok-test",
		Model:     "test-model",
	})

	if err := s.ExportProviderToEnv("test"); err != nil {
		t.Fatalf("ExportProviderToEnv() error: %v", err)
	}

	if v := os.Getenv("ANTHROPIC_BASE_URL"); v != "https://test.com" {
		t.Errorf("ANTHROPIC_BASE_URL = %q", v)
	}
	if v := os.Getenv("ANTHROPIC_AUTH_TOKEN"); v != "tok-test" {
		t.Errorf("ANTHROPIC_AUTH_TOKEN = %q", v)
	}
	if v := os.Getenv("ANTHROPIC_MODEL"); v != "test-model" {
		t.Errorf("ANTHROPIC_MODEL = %q", v)
	}

	// Cleanup
	os.Unsetenv("ANTHROPIC_BASE_URL")
	os.Unsetenv("ANTHROPIC_AUTH_TOKEN")
	os.Unsetenv("ANTHROPIC_MODEL")
}

func TestStoreExportProviderToEnvNotFound(t *testing.T) {
	s, _ := newTestStore(t)
	s.Load()

	err := s.ExportProviderToEnv("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent provider")
	}
}

func TestStoreSavePermissions(t *testing.T) {
	s, home := newTestStore(t)
	s.Load()

	s.SetProvider("x", &ProviderConfig{BaseURL: "https://x.com", AuthToken: "tok"})

	path := filepath.Join(home, ConfigDir, ConfigFile)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat error: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("file permissions = %o, want 0600", info.Mode().Perm())
	}
}

func TestDefaultStore(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	ResetDefaultStore()
	t.Cleanup(func() { ResetDefaultStore() })

	s := DefaultStore()
	if s == nil {
		t.Fatal("DefaultStore() returned nil")
	}

	// Calling again should return the same instance
	s2 := DefaultStore()
	if s != s2 {
		t.Error("DefaultStore() returned different instances")
	}
}

func TestStoreProviderMap(t *testing.T) {
	s, _ := newTestStore(t)
	s.Load()

	s.SetProvider("a", &ProviderConfig{BaseURL: "https://a.com", AuthToken: "tok"})
	s.SetProvider("b", &ProviderConfig{BaseURL: "https://b.com", AuthToken: "tok"})

	m := s.ProviderMap()
	if len(m) != 2 {
		t.Errorf("ProviderMap() has %d entries, want 2", len(m))
	}
	if m["a"] == nil || m["b"] == nil {
		t.Error("ProviderMap() missing entries")
	}
}

func TestStoreSetProfileOrderNil(t *testing.T) {
	s, _ := newTestStore(t)
	s.Load()

	s.SetProfileOrder("test", nil)
	order := s.GetProfileOrder("test")
	if order == nil || len(order) != 0 {
		t.Errorf("expected empty slice, got %v", order)
	}
}
