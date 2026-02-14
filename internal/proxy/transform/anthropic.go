package transform

// AnthropicTransformer handles Anthropic Messages API format.
// This is the default format used by Claude Code.
type AnthropicTransformer struct{}

func (t *AnthropicTransformer) Name() string {
	return "anthropic"
}

// TransformRequest transforms a request to Anthropic format.
// If the client is already using Anthropic format, no transformation is needed.
// If the client is using OpenAI format, convert to Anthropic format.
func (t *AnthropicTransformer) TransformRequest(body []byte, clientFormat string) ([]byte, error) {
	if clientFormat == "" || clientFormat == "anthropic" {
		// No transformation needed
		return body, nil
	}

	// OpenAI → Anthropic transformation
	data, err := parseJSON(body)
	if err != nil {
		return body, nil // Return original on parse error
	}

	// Transform max_completion_tokens → max_tokens
	if maxCompletionTokens, ok := data["max_completion_tokens"]; ok {
		data["max_tokens"] = maxCompletionTokens
		delete(data, "max_completion_tokens")
	}

	// Transform n parameter (OpenAI uses n for number of completions)
	delete(data, "n")

	// Transform temperature (both use same format, no change needed)

	// Transform stop sequences (OpenAI: stop, Anthropic: stop_sequences)
	if stop, ok := data["stop"]; ok {
		data["stop_sequences"] = stop
		delete(data, "stop")
	}

	// Transform stream_options (OpenAI specific)
	delete(data, "stream_options")

	// Transform logprobs (OpenAI specific)
	delete(data, "logprobs")
	delete(data, "top_logprobs")

	// Transform presence_penalty/frequency_penalty (OpenAI specific, not supported in Anthropic)
	delete(data, "presence_penalty")
	delete(data, "frequency_penalty")

	// Transform seed (OpenAI specific)
	delete(data, "seed")

	// Transform response_format (OpenAI specific)
	delete(data, "response_format")

	// Transform tools format if present
	// OpenAI tools format is different from Anthropic
	// This is a simplified transformation - full implementation would need more work
	if tools, ok := data["tools"].([]interface{}); ok {
		anthropicTools := make([]interface{}, 0, len(tools))
		for _, tool := range tools {
			if toolMap, ok := tool.(map[string]interface{}); ok {
				if toolMap["type"] == "function" {
					if fn, ok := toolMap["function"].(map[string]interface{}); ok {
						anthropicTool := map[string]interface{}{
							"name":        fn["name"],
							"description": fn["description"],
						}
						if params, ok := fn["parameters"]; ok {
							anthropicTool["input_schema"] = params
						}
						anthropicTools = append(anthropicTools, anthropicTool)
					}
				}
			}
		}
		if len(anthropicTools) > 0 {
			data["tools"] = anthropicTools
		}
	}

	return toJSON(data)
}

// TransformResponse transforms a response from Anthropic format.
// If the client expects Anthropic format, no transformation is needed.
// If the client expects OpenAI format, convert from Anthropic format.
func (t *AnthropicTransformer) TransformResponse(body []byte, clientFormat string) ([]byte, error) {
	if clientFormat == "" || clientFormat == "anthropic" {
		// No transformation needed
		return body, nil
	}

	// Anthropic → OpenAI response transformation
	data, err := parseJSON(body)
	if err != nil {
		return body, nil
	}

	// Transform Anthropic response to OpenAI format
	// Anthropic: { id, type, role, content: [{type, text}], model, stop_reason, usage }
	// OpenAI: { id, object, created, model, choices: [{index, message, finish_reason}], usage }

	openAIResponse := map[string]interface{}{
		"id":      data["id"],
		"object":  "chat.completion",
		"created": 0, // Anthropic doesn't provide this
		"model":   data["model"],
	}

	// Transform content to choices
	var messageContent string
	if content, ok := data["content"].([]interface{}); ok {
		for _, c := range content {
			if cMap, ok := c.(map[string]interface{}); ok {
				if cMap["type"] == "text" {
					if text, ok := cMap["text"].(string); ok {
						messageContent = text
						break
					}
				}
			}
		}
	}

	// Map stop_reason to finish_reason
	finishReason := "stop"
	if stopReason, ok := data["stop_reason"].(string); ok {
		switch stopReason {
		case "end_turn":
			finishReason = "stop"
		case "max_tokens":
			finishReason = "length"
		case "tool_use":
			finishReason = "tool_calls"
		default:
			finishReason = stopReason
		}
	}

	openAIResponse["choices"] = []interface{}{
		map[string]interface{}{
			"index": 0,
			"message": map[string]interface{}{
				"role":    "assistant",
				"content": messageContent,
			},
			"finish_reason": finishReason,
		},
	}

	// Transform usage
	if usage, ok := data["usage"].(map[string]interface{}); ok {
		openAIResponse["usage"] = map[string]interface{}{
			"prompt_tokens":     usage["input_tokens"],
			"completion_tokens": usage["output_tokens"],
			"total_tokens": func() interface{} {
				input, _ := usage["input_tokens"].(float64)
				output, _ := usage["output_tokens"].(float64)
				return input + output
			}(),
		}
	}

	return toJSON(openAIResponse)
}
