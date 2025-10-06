package agent

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/easyagent-dev/llm"
)

type Runner interface {
	Run(ctx context.Context, req *AgentRequest, callback Callback) (*AgentResponse, error)
}

type StreamRunner interface {
	Run(ctx context.Context, req *AgentRequest, callback Callback) (*AgentStreamResponse, error)
}

type BaseRunner struct {
	systemPrompts     string
	maxMessageHistory int
}

// RunnerOption is a functional option for configuring runners
type RunnerOption func(*runnerConfig)

// runnerConfig holds configuration options for runners
type runnerConfig struct {
	systemPrompts     string
	maxMessageHistory int
}

// WithSystemPrompt sets a custom system prompt for the runner
func WithSystemPrompt(prompt string) RunnerOption {
	return func(c *runnerConfig) {
		c.systemPrompts = prompt
	}
}

// WithMaxMessageHistory sets the maximum message history for the runner
func WithMaxMessageHistory(max int) RunnerOption {
	return func(c *runnerConfig) {
		c.maxMessageHistory = max
	}
}

// newRunnerConfig creates a new runner configuration with default values
func newRunnerConfig(opts ...RunnerOption) *runnerConfig {
	config := &runnerConfig{
		maxMessageHistory: DefaultMaxMessageHistory,
	}
	for _, opt := range opts {
		opt(config)
	}
	return config
}

//go:embed prompts/json_system.md
var jsonSystemPrompt string //nolint:gochecknoglobals

func (r *BaseRunner) GetSystemPrompt(agent *Agent, message *llm.ModelMessage, tools []ModelTool) (string, error) {
	toolsPrompt, err := r.ToolsPrompts(tools)
	if err != nil {
		return "", fmt.Errorf("failed to create tools prompt: %w", err)
	}

	// Use custom prompts if set, otherwise use default jsonSystemPrompt
	systemPrompt := jsonSystemPrompt
	if r.systemPrompts != "" {
		systemPrompt = r.systemPrompts
	}

	prompts, err := llm.GetPrompts(systemPrompt, map[string]interface{}{
		"agent":     agent,
		"tools":     toolsPrompt,
		"userQuery": message.Content,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get prompts: %w", err)
	}
	return prompts, nil
}

func (r *BaseRunner) ToolsPrompts(tools []ModelTool) (string, error) {
	if len(tools) == 0 {
		return "No tools available", nil
	}

	// Use strings.Builder for efficient string concatenation
	var builder strings.Builder
	builder.Grow(len(tools) * 256) // Pre-allocate reasonable size

	for i, tool := range tools {
		if i > 0 {
			builder.WriteString("\n")
		}
		inputSchema, _ := json.Marshal(tool.InputSchema())
		builder.WriteString("<tool name=\"")
		builder.WriteString(tool.Name())
		builder.WriteString("\">\n<description>")
		builder.WriteString(tool.Description())
		builder.WriteString("</description>\n<input_schema>\n")
		builder.Write(inputSchema)
		builder.WriteString("\n</input_schema>")

		usage := tool.Usage()
		if usage != "" {
			builder.WriteString("\n<usage>\n")
			builder.WriteString(usage)
			builder.WriteString("\n</usage>")
		}
		builder.WriteString("\n</tool>")
	}
	return builder.String(), nil
}
