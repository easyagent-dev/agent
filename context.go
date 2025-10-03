package agent

import (
	"context"
	"github.com/easymvp-ai/llm"
)

// contextKey is a private type for context keys to prevent collisions
type contextKey string

// agentContextKey is the key for storing AgentContext in context.Context
const agentContextKey contextKey = "agent"

// AgentContextOf retrieves the AgentContext from a context.Context
// It returns the context and a boolean indicating if it was found
func AgentContextOf(ctx context.Context) (*AgentContext, bool) {
	ac, ok := ctx.Value(agentContextKey).(*AgentContext)
	return ac, ok
}

// WithAgentContext returns a new context with the given AgentContext
func WithAgentContext(ctx context.Context, ac *AgentContext) context.Context {
	return context.WithValue(ctx, agentContextKey, ac)
}

// AgentContext holds the execution context for an agent execution.
// It tracks the agent state, conversation history, and execution history.
type AgentContext struct {
	// Agent is the agent being executed
	Agent *Agent

	// Messages is the current conversation history
	Messages []*llm.ModelMessage

	// ToolsCalled is the list of all tool calls made during execution
	ToolsCalled []*llm.ToolCall

	// Session is a key-value store for session-specific data
	Session map[string]any

	// ExecutionHistory tracks detailed tool execution information
	ExecutionHistory []ToolExecution
}

// ToolExecution represents a single tool execution with timing and result information.
// This is useful for debugging, monitoring, and analytics.
type ToolExecution struct {
	// ToolName is the name of the tool that was executed
	ToolName string

	// Input is the input provided to the tool
	Input any

	// Output is the result returned by the tool
	Output any

	// Error contains any error that occurred during execution
	Error error

	// Duration is how long the tool execution took
	Duration int64 // nanoseconds

	// Timestamp is when the execution started (Unix timestamp)
	Timestamp int64
}

// IsToolCalled checks if a tool with the given name has been called during this execution.
func (ac *AgentContext) IsToolCalled(name string) bool {
	for _, toolCall := range ac.ToolsCalled {
		if toolCall.Name == name {
			return true
		}
	}
	return false
}

// AddExecution records a tool execution in the execution history.
func (ac *AgentContext) AddExecution(execution ToolExecution) {
	if ac.ExecutionHistory == nil {
		ac.ExecutionHistory = make([]ToolExecution, 0)
	}
	ac.ExecutionHistory = append(ac.ExecutionHistory, execution)
}

// GetExecutionsByTool returns all executions for a specific tool.
func (ac *AgentContext) GetExecutionsByTool(toolName string) []ToolExecution {
	var executions []ToolExecution
	for _, exec := range ac.ExecutionHistory {
		if exec.ToolName == toolName {
			executions = append(executions, exec)
		}
	}
	return executions
}
