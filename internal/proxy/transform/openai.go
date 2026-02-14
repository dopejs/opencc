package transform

// OpenAITransformer handles OpenAI Chat Completions API format.
// This is used by Codex and other OpenAI-compatible clients.
type OpenAITransformer struct{}

func (t *OpenAITransformer) Name() string {
	return "openai"
}

// TransformRequest transforms a request to OpenAI format.
// If the client is already using OpenAI format, no transformation is needed.
// If the client is using Anthropic format, convert to OpenAI format.
func (t *OpenAITransformer) TransformRequest(body []byte, clientFormat string) ([]byte, error) {
	if clientFormat == "openai" {
		// No transformation needed
		return body, nil
	}

	// Anthropic → OpenAI transformation
	data, err := parseJSON(body)
	if err != nil {
		return body, nil
	}

	// Transform max_tokens → max_completion_tokens
	if maxTokens, ok := data["max_tokens"]; ok {
		data["max_completion_tokens"] = maxTokens
		delete(data, "max_tokens")
	}

	// Transform stop_sequences → stop
	if stopSeq, ok := data["stop_sequences"]; ok {
		data["stop"] = stopSeq
		delete(data, "stop_sequences")
	}

	// Remove Anthropic-specific fields
	delete(data, "metadata")
	delete(data, "thinking")

	// Transform tools format if present
	// Anthropic tools format is different from OpenAI
	if tools, ok := data["tools"].([]interface{}); ok {
		openAITools := make([]interface{}, 0, len(tools))
		for _, tool := range tools {
			if toolMap, ok := tool.(map[string]interface{}); ok {
				openAITool := map[string]interface{}{
					"type": "function",
					"function": map[string]interface{}{
						"name":        toolMap["name"],
						"description": toolMap["description"],
					},
				}
				if inputSchema, ok := toolMap["input_schema"]; ok {
					openAITool["function"].(map[string]interface{})["parameters"] = inputSchema
				}
				openAITools = append(openAITools, openAITool)
			}
		}
		if len(openAITools) > 0 {
			data["tools"] = openAITools
		}
	}

	// Transform system message format
	// Anthropic uses "system" field, OpenAI uses system role in messages
	if system, ok := data["system"].(string); ok && system != "" {
		messages, _ := data["messages"].([]interface{})
		// Prepend system message
		systemMsg := map[string]interface{}{
			"role":    "system",
			"content": system,
		}
		data["messages"] = append([]interface{}{systemMsg}, messages...)
		delete(data, "system")
	}

	return toJSON(data)
}

// TransformResponse transforms a response from OpenAI format.
// If the client expects OpenAI format, no transformation is needed.
// If the client expects Anthropic format, convert from OpenAI format.
func (t *OpenAITransformer) TransformResponse(body []byte, clientFormat string) ([]byte, error) {
	if clientFormat == "openai" {
		// No transformation needed
		return body, nil
	}

	// OpenAI → Anthropic response transformation
	data, err := parseJSON(body)
	if err != nil {
		return body, nil
	}

	// Transform OpenAI response to Anthropic format
	// OpenAI: { id, object, created, model, choices: [{index, message, finish_reason}], usage }
	// Anthropic: { id, type, role, content: [{type, text}], model, stop_reason, usage }

	anthropicResponse := map[string]interface{}{
		"id":    data["id"],
		"type":  "message",
		"role":  "assistant",
		"model": data["model"],
	}

	// Transform choices to content
	if choices, ok := data["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					anthropicResponse["content"] = []interface{}{
						map[string]interface{}{
							"type": "text",
							"text": content,
						},
					}
				}
			}

			// Map finish_reason to stop_reason
			if finishReason, ok := choice["finish_reason"].(string); ok {
				switch finishReason {
				case "stop":
					anthropicResponse["stop_reason"] = "end_turn"
				case "length":
					anthropicResponse["stop_reason"] = "max_tokens"
				case "tool_calls":
					anthropicResponse["stop_reason"] = "tool_use"
				default:
					anthropicResponse["stop_reason"] = finishReason
				}
			}
		}
	}

	// Transform usage
	if usage, ok := data["usage"].(map[string]interface{}); ok {
		anthropicResponse["usage"] = map[string]interface{}{
			"input_tokens":  usage["prompt_tokens"],
			"output_tokens": usage["completion_tokens"],
		}
	}

	return toJSON(anthropicResponse)
}
