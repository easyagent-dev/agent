package agent

import (
	"context"
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

// BeforeToolCall logs the tool call
func (c *DefaultCallback) BeforeToolCall(ctx context.Context, toolName string, input any) (any, error) {
	if c.Logger != nil {
		c.Logger.Info("Calling tool", "tool", toolName, "input", input)
	}
	return nil, nil
}

// AfterToolCall logs the tool result
func (c *DefaultCallback) AfterToolCall(ctx context.Context, toolName string, input any, output any) (any, error) {
	if c.Logger != nil {
		c.Logger.Info("Tool completed", "tool", toolName, "output", output)
	}
	return nil, nil
}
