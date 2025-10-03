package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/easymvp-ai/llm"
)

// Callback defines lifecycle hooks for agent execution
// All methods return (result, error):
// - If result is non-nil, it overrides the normal execution
// - If error is non-nil, execution continues with error handling
type Callback interface {
	// BeforeModel is called before sending a request to the LLM
	BeforeModel(ctx context.Context, provider string, model string, req *llm.CompletionRequest) (*llm.CompletionResponse, error)

	// AfterModel is called after receiving a response from the LLM
	AfterModel(ctx context.Context, provider string, model string, req *llm.CompletionRequest, resp *llm.CompletionResponse) (*llm.CompletionResponse, error)

	// BeforeToolCall is called before executing a tool
	BeforeToolCall(ctx context.Context, toolName string, input any) (any, error)

	// AfterToolCall is called after a tool execution completes
	AfterToolCall(ctx context.Context, toolName string, input any, output interface{}) (any, error)
}

// DefaultCallback implements the Callback interface with logging support
type DefaultCallback struct {
	Logger Logger
}

// NewDefaultCallback creates a new DefaultCallback with the given logger
func NewDefaultCallback(logger Logger) *DefaultCallback {
	if logger == nil {
		logger = &NoOpLogger{}
	}
	return &DefaultCallback{Logger: logger}
}

// BeforeModel logs the model call
func (c *DefaultCallback) BeforeModel(ctx context.Context, provider string, model string, req *llm.CompletionRequest) (*llm.CompletionResponse, error) {
	if c.Logger != nil {
		c.Logger.Debug("Calling model", "provider", provider, "model", model)
	}
	return nil, nil
}

// AfterModel logs the model response
func (c *DefaultCallback) AfterModel(ctx context.Context, provider string, model string, req *llm.CompletionRequest, resp *llm.CompletionResponse) (*llm.CompletionResponse, error) {
	if c.Logger != nil {
		c.Logger.Debug("Model response received", "provider", provider, "model", model)
	}
	return nil, nil
}

// sensitiveKeys are keys that should be redacted in logs
var sensitiveKeys = []string{
	"password", "passwd", "pwd",
	"secret", "api_key", "apikey", "api-key",
	"token", "auth", "authorization",
	"private_key", "privatekey", "private-key",
	"access_token", "refresh_token",
	"session", "cookie",
	"credential", "credentials",
}

// sanitizeForLogging removes sensitive data from logs
func sanitizeForLogging(data any, maxLen int) any {
	if data == nil {
		return nil
	}

	// If maxLen is 0, use a reasonable default
	if maxLen == 0 {
		maxLen = 500
	}

	// Handle different types
	switch v := data.(type) {
	case string:
		if len(v) > maxLen {
			return v[:maxLen] + "... (truncated)"
		}
		return v
	case map[string]any:
		sanitized := make(map[string]any)
		for key, value := range v {
			lowerKey := strings.ToLower(key)
			isSensitive := false
			for _, sensitiveKey := range sensitiveKeys {
				if strings.Contains(lowerKey, sensitiveKey) {
					isSensitive = true
					break
				}
			}
			if isSensitive {
				sanitized[key] = "***REDACTED***"
			} else {
				sanitized[key] = sanitizeForLogging(value, maxLen)
			}
		}
		return sanitized
	case []any:
		sanitized := make([]any, len(v))
		for i, item := range v {
			sanitized[i] = sanitizeForLogging(item, maxLen)
		}
		return sanitized
	default:
		// For other types, try to convert to JSON and truncate if needed
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		jsonStr := string(jsonBytes)
		if len(jsonStr) > maxLen {
			return jsonStr[:maxLen] + "... (truncated)"
		}
		// Try to unmarshal back to map to apply sanitization
		var m map[string]any
		if err := json.Unmarshal(jsonBytes, &m); err == nil {
			return sanitizeForLogging(m, maxLen)
		}
		return jsonStr
	}
}

// BeforeToolCall logs the tool call with sanitized input
func (c *DefaultCallback) BeforeToolCall(ctx context.Context, toolName string, input any) (any, error) {
	if c.Logger != nil {
		sanitizedInput := sanitizeForLogging(input, 500)
		c.Logger.Info("Calling tool", "tool", toolName, "input", sanitizedInput)
	}
	return nil, nil
}

// AfterToolCall logs the tool result with sanitized output
func (c *DefaultCallback) AfterToolCall(ctx context.Context, toolName string, input any, output any) (any, error) {
	if c.Logger != nil {
		sanitizedOutput := sanitizeForLogging(output, 500)
		c.Logger.Info("Tool completed", "tool", toolName, "output", sanitizedOutput)
	}
	return nil, nil
}
