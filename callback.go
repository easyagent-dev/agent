package agent

import (
	"context"
	"fmt"
	"github.com/easymvp-ai/llm"
)

type Callback interface {
	BeforeModel(ctx context.Context, provider string, model string, req *llm.CompletionRequest) (*llm.CompletionResponse, error)
	AfterModel(ctx context.Context, provider string, model string, req *llm.CompletionRequest, resp *llm.CompletionResponse) (*llm.CompletionResponse, error)
	BeforeToolCall(ctx context.Context, toolName string, input any) (any, error)
	AfterToolCall(ctx context.Context, toolName string, input any, output interface{}) (any, error)
}

// SimpleCallback implements the agent.Callback interface
type DefaultCallback struct{}

func (c *DefaultCallback) BeforeModel(ctx context.Context, provider string, model string, req *llm.CompletionRequest) (*llm.CompletionResponse, error) {
	return nil, nil
}

func (c *DefaultCallback) AfterModel(ctx context.Context, provider string, model string, req *llm.CompletionRequest, resp *llm.CompletionResponse) (*llm.CompletionResponse, error) {
	return nil, nil
}

func (c *DefaultCallback) BeforeToolCall(ctx context.Context, toolName string, input any) (any, error) {
	fmt.Printf("Calling tool: %s with input: %+v\n", toolName, input)
	return nil, nil
}

func (c *DefaultCallback) AfterToolCall(ctx context.Context, toolName string, input any, output any) (any, error) {
	fmt.Printf("Tool %s returned: %+v\n", toolName, output)
	return nil, nil
}
