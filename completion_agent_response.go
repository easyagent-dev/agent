package agent

import (
	"github.com/easyagent-dev/llm"
)

// AgentResponse represents the result of an agent execution.
// It contains the final output, token usage statistics, and cost information.
type AgentResponse struct {
	// Output is the final result from the agent's execution
	// The structure matches the OutputSchema specified in AgentRequest
	Output any `json:"output"`

	// Usage contains token usage statistics for the entire execution
	// Includes prompt tokens, completion tokens, and total tokens
	Usage *llm.TokenUsage `json:"usage"`

	// Cost is the estimated cost of the execution in USD
	// May be nil if cost tracking is not enabled
	Cost *float64
}

// AgentStreamResponse is a channel that streams agent events during execution.
// This enables real-time monitoring of agent progress.
type AgentStreamResponse <-chan AgentEvent

// AgentEventType represents the type of event in a streaming response
type AgentEventType string

const (
	// AgentEventTypeText indicates a text output event
	AgentEventTypeText AgentEventType = "text"

	// AgentEventTypeUseTool indicates a tool usage event
	AgentEventTypeUseTool AgentEventType = "use-tool"

	// AgentEventTypeReasoning indicates an internal reasoning event
	AgentEventTypeReasoning AgentEventType = "reasoning"

	// AgentEventTypeError indicates an error event
	AgentEventTypeError AgentEventType = "error"
)

// AgentEvent represents a single event in a streaming agent response.
// Different event types will populate different fields.
type AgentEvent struct {
	// Type identifies what kind of event this is
	Type AgentEventType

	// Text contains text output (for Text events)
	Text *string

	// Reasoning contains internal reasoning (for Reasoning events)
	Reasoning *string

	// ErrorMessage contains error details (for Error events)
	ErrorMessage *string

	ToolCall *llm.ToolCall

	// Partial indicates if this is a partial event (more data coming)
	Partial bool
}
