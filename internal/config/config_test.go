package config

import (
	"encoding/json"
	"os"
	"testing"
)

func setTestHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	ResetDefaultStore()
	t.Cleanup(func() { ResetDefaultStore() })
	return dir
}

func TestReadWriteFallbackOrder(t *testing.T) {
	setTestHome(t)

	names := []string{"yunyi", "cctq", "minimax"}
	if err := WriteFallbackOrder(names); err != nil {
		t.Fatalf("WriteFallbackOrder() error: %v", err)
	}

	got, err := ReadFallbackOrder()
	if err != nil {
		t.Fatalf("ReadFallbackOrder() error: %v", err)
	}

	if len(got) != len(names) {
		t.Fatalf("got %d names, want %d", len(got), len(names))
	}
	for i, n := range names {
		if got[i] != n {
			t.Errorf("got[%d] = %q, want %q", i, got[i], n)
		}
	}
}

func TestReadFallbackOrderMissing(t *testing.T) {
	setTestHome(t)
	// Don't create default profile

	_, err := ReadFallbackOrder()
	if err == nil {
		t.Error("expected error for missing default profile")
	}
}

func TestWriteFallbackOrderEmpty(t *testing.T) {
	setTestHome(t)

	if err := WriteFallbackOrder(nil); err != nil {
		t.Fatalf("WriteFallbackOrder(nil) error: %v", err)
	}

	got, err := ReadFallbackOrder()
	if err != nil {
		t.Fatalf("ReadFallbackOrder() error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 names, got %d", len(got))
	}
}

func TestWriteFallbackOrderCreatesDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	ResetDefaultStore()
	t.Cleanup(func() { ResetDefaultStore() })

	if err := WriteFallbackOrder([]string{"a"}); err != nil {
		t.Fatalf("WriteFallbackOrder() error: %v", err)
	}

	got, err := ReadFallbackOrder()
	if err != nil {
		t.Fatalf("ReadFallbackOrder() error: %v", err)
	}
	if len(got) != 1 || got[0] != "a" {
		t.Errorf("got %v, want [a]", got)
	}
}

func TestWriteFallbackOrderErrorBadDir(t *testing.T) {
	t.Setenv("HOME", "/dev/null/impossible")
	ResetDefaultStore()
	t.Cleanup(func() { ResetDefaultStore() })

	err := WriteFallbackOrder([]string{"a"})
	if err == nil {
		t.Error("expected error when dir can't be created")
	}
}

func TestRemoveFromFallbackOrder(t *testing.T) {
	setTestHome(t)
	WriteFallbackOrder([]string{"a", "b", "c"})

	if err := RemoveFromFallbackOrder("b"); err != nil {
		t.Fatalf("RemoveFromFallbackOrder() error: %v", err)
	}

	got, _ := ReadFallbackOrder()
	if len(got) != 2 || got[0] != "a" || got[1] != "c" {
		t.Errorf("got %v, want [a c]", got)
	}
}

func TestRemoveFromFallbackOrderMissingProfile(t *testing.T) {
	setTestHome(t)
	// No default profile — should be a no-op
	if err := RemoveFromFallbackOrder("x"); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestRemoveFromFallbackOrderNotPresent(t *testing.T) {
	setTestHome(t)
	WriteFallbackOrder([]string{"a", "b"})

	if err := RemoveFromFallbackOrder("z"); err != nil {
		t.Fatalf("error: %v", err)
	}

	got, _ := ReadFallbackOrder()
	if len(got) != 2 {
		t.Errorf("expected 2 names unchanged, got %v", got)
	}
}

func TestRemoveFromFallbackOrderFirst(t *testing.T) {
	setTestHome(t)
	WriteFallbackOrder([]string{"a", "b", "c"})

	if err := RemoveFromFallbackOrder("a"); err != nil {
		t.Fatalf("error: %v", err)
	}

	got, _ := ReadFallbackOrder()
	if len(got) != 2 || got[0] != "b" || got[1] != "c" {
		t.Errorf("got %v, want [b c]", got)
	}
}

func TestRemoveFromFallbackOrderLast(t *testing.T) {
	setTestHome(t)
	WriteFallbackOrder([]string{"a", "b", "c"})

	if err := RemoveFromFallbackOrder("c"); err != nil {
		t.Fatalf("error: %v", err)
	}

	got, _ := ReadFallbackOrder()
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("got %v, want [a b]", got)
	}
}

func TestRemoveFromFallbackOrderOnlyItem(t *testing.T) {
	setTestHome(t)
	WriteFallbackOrder([]string{"solo"})

	if err := RemoveFromFallbackOrder("solo"); err != nil {
		t.Fatalf("error: %v", err)
	}

	got, _ := ReadFallbackOrder()
	if len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}

func TestRemoveFromFallbackOrderDuplicates(t *testing.T) {
	setTestHome(t)
	WriteFallbackOrder([]string{"a", "b", "a", "c"})

	if err := RemoveFromFallbackOrder("a"); err != nil {
		t.Fatalf("error: %v", err)
	}

	got, _ := ReadFallbackOrder()
	if len(got) != 2 || got[0] != "b" || got[1] != "c" {
		t.Errorf("got %v, want [b c]", got)
	}
}

// --- Profile tests ---

func TestReadWriteProfileOrder(t *testing.T) {
	setTestHome(t)

	names := []string{"p1", "p2"}
	if err := WriteProfileOrder("work", names); err != nil {
		t.Fatalf("WriteProfileOrder() error: %v", err)
	}

	got, err := ReadProfileOrder("work")
	if err != nil {
		t.Fatalf("ReadProfileOrder() error: %v", err)
	}
	if len(got) != 2 || got[0] != "p1" || got[1] != "p2" {
		t.Errorf("got %v, want [p1 p2]", got)
	}

	// Default profile should be unaffected
	_, err = ReadProfileOrder("default")
	if err == nil {
		t.Error("expected error for missing default profile")
	}
}

func TestListProfiles(t *testing.T) {
	setTestHome(t)

	WriteProfileOrder("default", []string{"a"})
	WriteProfileOrder("work", []string{"b"})
	WriteProfileOrder("staging", []string{"c"})

	profiles := ListProfiles()
	if len(profiles) != 3 {
		t.Fatalf("expected 3 profiles, got %d: %v", len(profiles), profiles)
	}
	// Should be sorted
	if profiles[0] != "default" || profiles[1] != "staging" || profiles[2] != "work" {
		t.Errorf("got %v, want [default staging work]", profiles)
	}
}

func TestListProfilesEmpty(t *testing.T) {
	setTestHome(t)
	profiles := ListProfiles()
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles, got %v", profiles)
	}
}

func TestDeleteProfile(t *testing.T) {
	setTestHome(t)
	WriteProfileOrder("work", []string{"a"})

	if err := DeleteProfile("work"); err != nil {
		t.Fatalf("DeleteProfile() error: %v", err)
	}

	_, err := ReadProfileOrder("work")
	if err == nil {
		t.Error("expected error after deleting profile")
	}
}

func TestDeleteProfileDefault(t *testing.T) {
	setTestHome(t)
	WriteProfileOrder("default", []string{"a"})

	err := DeleteProfile("default")
	if err == nil {
		t.Error("expected error when deleting default profile")
	}

	// Default should still exist
	got, err := ReadProfileOrder("default")
	if err != nil {
		t.Fatalf("default profile should still exist: %v", err)
	}
	if len(got) != 1 || got[0] != "a" {
		t.Errorf("got %v, want [a]", got)
	}
}

func TestDeleteProfileEmpty(t *testing.T) {
	setTestHome(t)
	err := DeleteProfile("")
	if err == nil {
		t.Error("expected error when deleting empty profile name")
	}
}

func TestRemoveFromProfileOrder(t *testing.T) {
	setTestHome(t)
	WriteProfileOrder("work", []string{"a", "b", "c"})

	if err := RemoveFromProfileOrder("work", "b"); err != nil {
		t.Fatalf("error: %v", err)
	}

	got, _ := ReadProfileOrder("work")
	if len(got) != 2 || got[0] != "a" || got[1] != "c" {
		t.Errorf("got %v, want [a c]", got)
	}
}

func TestProviderConfigExportToEnv(t *testing.T) {
	p := &ProviderConfig{
		BaseURL:        "https://test.com",
		AuthToken:      "tok-test",
		Model:          "m1",
		ReasoningModel: "m2",
		HaikuModel:     "m3",
		OpusModel:      "m4",
		SonnetModel:    "m5",
	}

	p.ExportToEnv()

	tests := map[string]string{
		"ANTHROPIC_BASE_URL":              "https://test.com",
		"ANTHROPIC_AUTH_TOKEN":            "tok-test",
		"ANTHROPIC_MODEL":                 "m1",
		"ANTHROPIC_REASONING_MODEL":       "m2",
		"ANTHROPIC_DEFAULT_HAIKU_MODEL":   "m3",
		"ANTHROPIC_DEFAULT_OPUS_MODEL":    "m4",
		"ANTHROPIC_DEFAULT_SONNET_MODEL":  "m5",
	}

	for k, want := range tests {
		if got := os.Getenv(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}

	// Cleanup
	for k := range tests {
		os.Unsetenv(k)
	}
}

func TestConfigDirPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	got := ConfigDirPath()
	if got != dir+"/.opencc" {
		t.Errorf("ConfigDirPath() = %q", got)
	}
}

func TestConfigFilePath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	got := ConfigFilePath()
	if got != dir+"/.opencc/opencc.json" {
		t.Errorf("ConfigFilePath() = %q", got)
	}
}

func TestLogPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	got := LogPath()
	if got != dir+"/.opencc/proxy.log" {
		t.Errorf("LogPath() = %q", got)
	}
}

// --- ProfileConfig JSON tests ---

func TestProfileConfigUnmarshalOldFormat(t *testing.T) {
	// Old format: ["p1", "p2"]
	data := []byte(`["p1", "p2"]`)
	var pc ProfileConfig
	if err := json.Unmarshal(data, &pc); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}
	if len(pc.Providers) != 2 || pc.Providers[0] != "p1" || pc.Providers[1] != "p2" {
		t.Errorf("got providers %v, want [p1 p2]", pc.Providers)
	}
	if pc.Routing != nil {
		t.Errorf("routing should be nil for old format, got %v", pc.Routing)
	}
}

func TestProfileConfigUnmarshalNewFormat(t *testing.T) {
	data := []byte(`{
		"providers": ["a", "b"],
		"routing": {
			"think": {"providers": ["b", "a"], "model": "claude-opus-4-5"},
			"image": {"providers": ["a"]}
		}
	}`)
	var pc ProfileConfig
	if err := json.Unmarshal(data, &pc); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}
	if len(pc.Providers) != 2 || pc.Providers[0] != "a" || pc.Providers[1] != "b" {
		t.Errorf("got providers %v, want [a b]", pc.Providers)
	}
	if pc.Routing == nil {
		t.Fatal("routing should not be nil")
	}
	if len(pc.Routing) != 2 {
		t.Fatalf("expected 2 routing entries, got %d", len(pc.Routing))
	}

	thinkRoute := pc.Routing[ScenarioThink]
	if thinkRoute == nil {
		t.Fatal("think route should exist")
	}
	if len(thinkRoute.Providers) != 2 || thinkRoute.Providers[0] != "b" {
		t.Errorf("think providers = %v", thinkRoute.Providers)
	}
	if thinkRoute.Model != "claude-opus-4-5" {
		t.Errorf("think model = %q", thinkRoute.Model)
	}

	imageRoute := pc.Routing[ScenarioImage]
	if imageRoute == nil {
		t.Fatal("image route should exist")
	}
	if len(imageRoute.Providers) != 1 || imageRoute.Providers[0] != "a" {
		t.Errorf("image providers = %v", imageRoute.Providers)
	}
	if imageRoute.Model != "" {
		t.Errorf("image model should be empty, got %q", imageRoute.Model)
	}
}

func TestProfileConfigUnmarshalNewFormatNoRouting(t *testing.T) {
	data := []byte(`{"providers": ["x", "y"]}`)
	var pc ProfileConfig
	if err := json.Unmarshal(data, &pc); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}
	if len(pc.Providers) != 2 {
		t.Errorf("expected 2 providers, got %d", len(pc.Providers))
	}
	if pc.Routing != nil {
		t.Errorf("routing should be nil, got %v", pc.Routing)
	}
}

func TestProfileConfigRoundTrip(t *testing.T) {
	original := ProfileConfig{
		Providers: []string{"a", "b", "c"},
		Routing: map[Scenario]*ScenarioRoute{
			ScenarioThink: {
				Providers: []string{"c", "a"},
				Model:     "claude-opus-4-5",
			},
			ScenarioLongContext: {
				Providers: []string{"b"},
			},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var restored ProfileConfig
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if len(restored.Providers) != 3 {
		t.Errorf("providers count: got %d, want 3", len(restored.Providers))
	}
	for i, want := range original.Providers {
		if restored.Providers[i] != want {
			t.Errorf("providers[%d] = %q, want %q", i, restored.Providers[i], want)
		}
	}

	if len(restored.Routing) != 2 {
		t.Fatalf("routing count: got %d, want 2", len(restored.Routing))
	}

	thinkRoute := restored.Routing[ScenarioThink]
	if thinkRoute == nil || thinkRoute.Model != "claude-opus-4-5" {
		t.Errorf("think route not properly round-tripped")
	}
	if len(thinkRoute.Providers) != 2 || thinkRoute.Providers[0] != "c" {
		t.Errorf("think providers = %v", thinkRoute.Providers)
	}

	lcRoute := restored.Routing[ScenarioLongContext]
	if lcRoute == nil || len(lcRoute.Providers) != 1 || lcRoute.Providers[0] != "b" {
		t.Errorf("longContext route not properly round-tripped")
	}
}

func TestProfileConfigRoundTripOldFormat(t *testing.T) {
	// Start with old format, marshal, unmarshal — should produce equivalent result
	oldData := []byte(`["x", "y"]`)
	var pc ProfileConfig
	if err := json.Unmarshal(oldData, &pc); err != nil {
		t.Fatalf("Unmarshal old format error: %v", err)
	}

	newData, err := json.Marshal(pc)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var restored ProfileConfig
	if err := json.Unmarshal(newData, &restored); err != nil {
		t.Fatalf("Unmarshal new format error: %v", err)
	}

	if len(restored.Providers) != 2 || restored.Providers[0] != "x" || restored.Providers[1] != "y" {
		t.Errorf("got providers %v, want [x y]", restored.Providers)
	}
	if restored.Routing != nil {
		t.Errorf("routing should be nil after round-trip from old format")
	}
}

func TestFullConfigRoundTrip(t *testing.T) {
	setTestHome(t)

	// Write config with routing
	pc := &ProfileConfig{
		Providers: []string{"p1", "p2"},
		Routing: map[Scenario]*ScenarioRoute{
			ScenarioThink: {Providers: []string{"p2"}, Model: "model-x"},
		},
	}
	if err := SetProfileConfig("myprofile", pc); err != nil {
		t.Fatalf("SetProfileConfig error: %v", err)
	}

	// Read it back
	got := GetProfileConfig("myprofile")
	if got == nil {
		t.Fatal("GetProfileConfig returned nil")
	}
	if len(got.Providers) != 2 {
		t.Errorf("providers count = %d", len(got.Providers))
	}
	if got.Routing == nil || got.Routing[ScenarioThink] == nil {
		t.Fatal("routing not preserved")
	}
	if got.Routing[ScenarioThink].Model != "model-x" {
		t.Errorf("model = %q", got.Routing[ScenarioThink].Model)
	}
}

func TestDeleteProviderCascadeRouting(t *testing.T) {
	setTestHome(t)

	// Setup: provider "a" and "b", profile with routing referencing both
	store := DefaultStore()
	store.SetProvider("a", &ProviderConfig{BaseURL: "https://a.com", AuthToken: "t"})
	store.SetProvider("b", &ProviderConfig{BaseURL: "https://b.com", AuthToken: "t"})

	pc := &ProfileConfig{
		Providers: []string{"a", "b"},
		Routing: map[Scenario]*ScenarioRoute{
			ScenarioThink: {Providers: []string{"a", "b"}, Model: "m1"},
			ScenarioImage: {Providers: []string{"a"}},
		},
	}
	SetProfileConfig("default", pc)

	// Delete provider "a"
	DeleteProviderByName("a")

	// Check routing was updated
	got := GetProfileConfig("default")
	if got == nil {
		t.Fatal("profile should still exist")
	}

	// "a" should be removed from providers
	for _, p := range got.Providers {
		if p == "a" {
			t.Error("provider 'a' should have been removed from providers")
		}
	}

	// Check routing
	if got.Routing != nil {
		if think := got.Routing[ScenarioThink]; think != nil {
			for _, p := range think.Providers {
				if p == "a" {
					t.Error("provider 'a' should have been removed from think route")
				}
			}
			if len(think.Providers) != 1 || think.Providers[0] != "b" {
				t.Errorf("think providers = %v, want [b]", think.Providers)
			}
		}
		// image route had only "a" — should be removed entirely
		if image := got.Routing[ScenarioImage]; image != nil {
			t.Error("image route should have been removed (no providers left)")
		}
	}
}
