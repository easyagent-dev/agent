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
	Agent *Agent

	// Messages is the current conversation history
	Messages []*llm.ModelMessage

	// Session is a key-value store for session-specific data
	Session map[string]any

	// mu protects ExecutionHistory from concurrent access
	mu sync.RWMutex

	// ToolExecutions tracks detailed tool execution information
	ToolCalls []*llm.ToolCall
}

// IsToolCalled checks if a tool with the given name has been called during this execution.
// This method is safe for concurrent use.
func (ac *AgentContext) IsToolCalled(name string) bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	for _, toolCall := range ac.ToolCalls {
		if toolCall.Name == name {
			return true
		}
	}
	return false
}

// AppendToolCall records a tool execution in the execution history.
// This method is safe for concurrent use.
func (ac *AgentContext) AppendToolCall(toolCall *llm.ToolCall) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	if ac.ToolCalls == nil {
		ac.ToolCalls = make([]*llm.ToolCall, 0, 10) // Pre-allocate with capacity
	}
	ac.ToolCalls = append(ac.ToolCalls, toolCall)
}

// FindToolCalls returns all executions for a specific tool.
// This method is safe for concurrent use.
func (ac *AgentContext) FindToolCalls(toolName string) []*llm.ToolCall {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	toolCalls := make([]*llm.ToolCall, 0, len(ac.ToolCalls)/2) // Reasonable pre-allocation
	for _, toolCall := range ac.ToolCalls {
		if toolCall.Name == toolName {
			toolCalls = append(toolCalls, toolCall)
		}
	}
	return toolCalls
}
