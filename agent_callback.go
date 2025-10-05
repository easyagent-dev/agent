package agent

import (
	"context"
	"github.com/easyagent-dev/llm"
)

// Callback defines lifecycle hooks for agent execution
// All methods return (result, error):
// - If result is non-nil, it overrides the normal execution
// - If error is non-nil, execution continues with error handling
type Callback interface {
	// BeforeModel is called before sending a request to the LLM
	BeforeModel(ctx context.Context, provider string, model string, prompts string, messages []*llm.ModelMessage) error

	// AfterModel is called after receiving a response from the LLM
	AfterModel(ctx context.Context, provider string, model string, prompts string, messages []*llm.ModelMessage, output string, usage *llm.TokenUsage) error

	// BeforeToolCall is called before executing a tool
	BeforeToolCall(ctx context.Context, toolName string, input any) error

	// AfterToolCall is called after a tool execution completes
	AfterToolCall(ctx context.Context, toolName string, input any, output interface{}) error
}

// DefaultCallback implements the Callback interface with logging support
type DefaultCallback struct {
	trace bool
}

// NewDefaultCallback creates a new DefaultCallback with the given logger
func NewDefaultCallback(trace bool) *DefaultCallback {
	return &DefaultCallback{trace: trace}
}
