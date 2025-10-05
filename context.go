package agent

import (
	"context"
	"sync"

	"github.com/easyagent-dev/llm"
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
// This type is safe for concurrent use.
type AgentContext struct {
	// Agent is the agent being executed
	Agent *CompletionAgent

	// Messages is the current conversation history
	Messages []*llm.ModelMessage

	// Session is a key-value store for session-specific data
	Session map[string]any

	// mu protects ExecutionHistory from concurrent access
	mu sync.RWMutex

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
// This method is safe for concurrent use.
func (ac *AgentContext) IsToolCalled(name string) bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	for _, toolCall := range ac.ExecutionHistory {
		if toolCall.ToolName == name {
			return true
		}
	}
	return false
}

// AddExecution records a tool execution in the execution history.
// This method is safe for concurrent use.
func (ac *AgentContext) AddExecution(execution ToolExecution) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if ac.ExecutionHistory == nil {
		ac.ExecutionHistory = make([]ToolExecution, 0, 10) // Pre-allocate with capacity
	}
	ac.ExecutionHistory = append(ac.ExecutionHistory, execution)
}

// GetExecutionsByTool returns all executions for a specific tool.
// This method is safe for concurrent use.
func (ac *AgentContext) GetExecutionsByTool(toolName string) []ToolExecution {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	executions := make([]ToolExecution, 0, len(ac.ExecutionHistory)/2) // Reasonable pre-allocation
	for _, exec := range ac.ExecutionHistory {
		if exec.ToolName == toolName {
			executions = append(executions, exec)
		}
	}
	return executions
}
