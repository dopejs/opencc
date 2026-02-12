package proxy

import (
	"strings"
	"testing"

	"github.com/dopejs/opencc/internal/config"
)

func TestDetectScenarioThink(t *testing.T) {
	body := map[string]interface{}{
		"model":    "claude-sonnet-4-5",
		"thinking": map[string]interface{}{"type": "enabled"},
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "hi"},
		},
	}
	got := DetectScenario(body)
	if got != config.ScenarioThink {
		t.Errorf("DetectScenario() = %q, want %q", got, config.ScenarioThink)
	}
}

func TestDetectScenarioThinkDisabled(t *testing.T) {
	body := map[string]interface{}{
		"model":    "claude-sonnet-4-5",
		"thinking": map[string]interface{}{"type": "disabled"},
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "hi"},
		},
	}
	got := DetectScenario(body)
	if got != config.ScenarioDefault {
		t.Errorf("DetectScenario() = %q, want %q", got, config.ScenarioDefault)
	}
}

func TestDetectScenarioImage(t *testing.T) {
	body := map[string]interface{}{
		"model": "claude-sonnet-4-5",
		"messages": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{"type": "text", "text": "What is this?"},
					map[string]interface{}{
						"type": "image",
						"source": map[string]interface{}{
							"type":       "base64",
							"media_type": "image/png",
							"data":       "iVBOR...",
						},
					},
				},
			},
		},
	}
	got := DetectScenario(body)
	if got != config.ScenarioImage {
		t.Errorf("DetectScenario() = %q, want %q", got, config.ScenarioImage)
	}
}

func TestDetectScenarioLongContext(t *testing.T) {
	longText := strings.Repeat("x", longContextThreshold+1)
	body := map[string]interface{}{
		"model": "claude-sonnet-4-5",
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": longText},
		},
	}
	got := DetectScenario(body)
	if got != config.ScenarioLongContext {
		t.Errorf("DetectScenario() = %q, want %q", got, config.ScenarioLongContext)
	}
}

func TestDetectScenarioLongContextFromBlocks(t *testing.T) {
	longText := strings.Repeat("y", longContextThreshold+1)
	body := map[string]interface{}{
		"model": "claude-sonnet-4-5",
		"messages": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{"type": "text", "text": longText},
				},
			},
		},
	}
	got := DetectScenario(body)
	if got != config.ScenarioLongContext {
		t.Errorf("DetectScenario() = %q, want %q", got, config.ScenarioLongContext)
	}
}

func TestDetectScenarioLongContextFromSystem(t *testing.T) {
	system := strings.Repeat("s", longContextThreshold+1)
	body := map[string]interface{}{
		"model":  "claude-sonnet-4-5",
		"system": system,
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "hi"},
		},
	}
	got := DetectScenario(body)
	if got != config.ScenarioLongContext {
		t.Errorf("DetectScenario() = %q, want %q", got, config.ScenarioLongContext)
	}
}

func TestDetectScenarioDefault(t *testing.T) {
	body := map[string]interface{}{
		"model": "claude-sonnet-4-5",
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "hello"},
		},
	}
	got := DetectScenario(body)
	if got != config.ScenarioDefault {
		t.Errorf("DetectScenario() = %q, want %q", got, config.ScenarioDefault)
	}
}

func TestDetectScenarioPriority_ThinkOverImage(t *testing.T) {
	body := map[string]interface{}{
		"model":    "claude-sonnet-4-5",
		"thinking": map[string]interface{}{"type": "enabled"},
		"messages": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{"type": "image", "source": map[string]interface{}{}},
				},
			},
		},
	}
	got := DetectScenario(body)
	if got != config.ScenarioThink {
		t.Errorf("DetectScenario() = %q, want %q (think takes priority over image)", got, config.ScenarioThink)
	}
}

func TestDetectScenarioPriority_ImageOverLongContext(t *testing.T) {
	longText := strings.Repeat("x", longContextThreshold+1)
	body := map[string]interface{}{
		"model": "claude-sonnet-4-5",
		"messages": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{"type": "text", "text": longText},
					map[string]interface{}{"type": "image", "source": map[string]interface{}{}},
				},
			},
		},
	}
	got := DetectScenario(body)
	if got != config.ScenarioImage {
		t.Errorf("DetectScenario() = %q, want %q (image takes priority over longContext)", got, config.ScenarioImage)
	}
}

func TestDetectScenarioFromJSON(t *testing.T) {
	data := []byte(`{"model":"claude-sonnet-4-5","thinking":{"type":"enabled"},"messages":[{"role":"user","content":"hi"}]}`)
	scenario, body := DetectScenarioFromJSON(data)
	if scenario != config.ScenarioThink {
		t.Errorf("scenario = %q, want %q", scenario, config.ScenarioThink)
	}
	if body == nil {
		t.Error("body should not be nil")
	}
}

func TestDetectScenarioFromJSONInvalid(t *testing.T) {
	scenario, body := DetectScenarioFromJSON([]byte("not json"))
	if scenario != config.ScenarioDefault {
		t.Errorf("scenario = %q, want %q for invalid JSON", scenario, config.ScenarioDefault)
	}
	if body != nil {
		t.Error("body should be nil for invalid JSON")
	}
}

func TestHasImageContentNoMessages(t *testing.T) {
	body := map[string]interface{}{}
	if hasImageContent(body) {
		t.Error("expected false for empty body")
	}
}

func TestIsLongContextShort(t *testing.T) {
	body := map[string]interface{}{
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "short"},
		},
	}
	if isLongContext(body) {
		t.Error("expected false for short content")
	}
}

func TestIsLongContextMultipleMessages(t *testing.T) {
	halfText := strings.Repeat("a", longContextThreshold/2+1)
	body := map[string]interface{}{
		"model": "claude-sonnet-4-5",
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": halfText},
			map[string]interface{}{"role": "assistant", "content": halfText},
		},
	}
	got := DetectScenario(body)
	if got != config.ScenarioLongContext {
		t.Errorf("DetectScenario() = %q, want %q for multiple messages totaling > threshold", got, config.ScenarioLongContext)
	}
}
