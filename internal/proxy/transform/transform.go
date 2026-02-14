// Package transform provides request/response transformation between different API formats.
package transform

import (
	"encoding/json"
)

// Transformer defines the interface for API format transformation.
type Transformer interface {
	// Name returns the transformer name (e.g., "anthropic", "openai")
	Name() string

	// TransformRequest transforms an incoming request body to the target format.
	// clientFormat is the format the client is using (e.g., "anthropic" for Claude Code)
	// Returns the transformed body and any error.
	TransformRequest(body []byte, clientFormat string) ([]byte, error)

	// TransformResponse transforms an outgoing response body from the target format.
	// clientFormat is the format the client expects.
	// Returns the transformed body and any error.
	TransformResponse(body []byte, clientFormat string) ([]byte, error)
}

// GetTransformer returns the appropriate transformer for the given provider type.
func GetTransformer(providerType string) Transformer {
	switch providerType {
	case "openai":
		return &OpenAITransformer{}
	default:
		return &AnthropicTransformer{}
	}
}

// NeedsTransform returns true if transformation is needed between client and provider formats.
func NeedsTransform(clientFormat, providerFormat string) bool {
	// Normalize empty to anthropic (default)
	if clientFormat == "" {
		clientFormat = "anthropic"
	}
	if providerFormat == "" {
		providerFormat = "anthropic"
	}
	return clientFormat != providerFormat
}

// parseJSON is a helper to parse JSON body into a map.
func parseJSON(body []byte) (map[string]interface{}, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	return data, nil
}

// toJSON is a helper to convert a map back to JSON.
func toJSON(data map[string]interface{}) ([]byte, error) {
	return json.Marshal(data)
}
