package agent

import (
	"context"
	"fmt"
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

// BeforeModel is called before sending a request to the LLM
func (c *DefaultCallback) BeforeModel(ctx context.Context, provider string, model string, prompts string, messages []*llm.ModelMessage) error {
	if c.trace {
		println(fmt.Sprintf("BeforeModel: %s/%s | Prompts: %d chars | Messages: %d", provider, model, len(prompts), len(messages)))
	}
	return nil
}

// AfterModel is called after receiving a response from the LLM
func (c *DefaultCallback) AfterModel(ctx context.Context, provider string, model string, prompts string, messages []*llm.ModelMessage, output string, usage *llm.TokenUsage) error {
	if c.trace {
		usageStr := "no usage"
		if usage != nil {
			usageStr = fmt.Sprintf("in:%d out:%d", usage.TotalInputTokens, usage.TotalOutputTokens)
		}
		println(fmt.Sprintf("AfterModel: %s/%s | Output: %d chars | Usage: %s", provider, model, len(output), usageStr))
	}
	return nil
}

// BeforeToolCall is called before executing a tool
func (c *DefaultCallback) BeforeToolCall(ctx context.Context, toolName string, input any) error {
	if c.trace {
		println(fmt.Sprintf("BeforeToolCall: %s", toolName))
	}
	return nil
}

// AfterToolCall is called after a tool execution completes
func (c *DefaultCallback) AfterToolCall(ctx context.Context, toolName string, input any, output interface{}) error {
	if c.trace {
		println(fmt.Sprintf("AfterToolCall: %s | Output: %T", toolName, output))
	}
	return nil
}
