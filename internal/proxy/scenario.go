package proxy

import (
	"encoding/json"

	"github.com/dopejs/opencc/internal/config"
)

const longContextThreshold = 32000

// DetectScenario examines a parsed request body and returns the matching scenario.
// Priority: think > image > longContext > default.
func DetectScenario(body map[string]interface{}) config.Scenario {
	if hasThinkingEnabled(body) {
		return config.ScenarioThink
	}
	if hasImageContent(body) {
		return config.ScenarioImage
	}
	if isLongContext(body) {
		return config.ScenarioLongContext
	}
	return config.ScenarioDefault
}

// DetectScenarioFromJSON parses raw JSON and detects the scenario.
func DetectScenarioFromJSON(data []byte) (config.Scenario, map[string]interface{}) {
	var body map[string]interface{}
	if err := json.Unmarshal(data, &body); err != nil {
		return config.ScenarioDefault, nil
	}
	return DetectScenario(body), body
}

// hasImageContent checks if any message contains an image content block.
func hasImageContent(body map[string]interface{}) bool {
	messages, ok := body["messages"].([]interface{})
	if !ok {
		return false
	}
	for _, msg := range messages {
		m, ok := msg.(map[string]interface{})
		if !ok {
			continue
		}
		content, ok := m["content"].([]interface{})
		if !ok {
			continue
		}
		for _, block := range content {
			b, ok := block.(map[string]interface{})
			if !ok {
				continue
			}
			if t, ok := b["type"].(string); ok && t == "image" {
				return true
			}
		}
	}
	return false
}

// isLongContext checks if the total text content in messages exceeds the threshold.
func isLongContext(body map[string]interface{}) bool {
	messages, ok := body["messages"].([]interface{})
	if !ok {
		return false
	}

	totalLen := 0
	for _, msg := range messages {
		m, ok := msg.(map[string]interface{})
		if !ok {
			continue
		}

		// Content can be a string or an array of blocks
		switch content := m["content"].(type) {
		case string:
			totalLen += len(content)
		case []interface{}:
			for _, block := range content {
				b, ok := block.(map[string]interface{})
				if !ok {
					continue
				}
				if t, ok := b["type"].(string); ok && t == "text" {
					if text, ok := b["text"].(string); ok {
						totalLen += len(text)
					}
				}
			}
		}

		if totalLen >= longContextThreshold {
			return true
		}
	}

	// Also check system prompt
	if system, ok := body["system"].(string); ok {
		totalLen += len(system)
	} else if systemBlocks, ok := body["system"].([]interface{}); ok {
		for _, block := range systemBlocks {
			b, ok := block.(map[string]interface{})
			if !ok {
				continue
			}
			if t, ok := b["type"].(string); ok && t == "text" {
				if text, ok := b["text"].(string); ok {
					totalLen += len(text)
				}
			}
		}
	}

	return totalLen >= longContextThreshold
}
