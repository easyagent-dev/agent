package agent

import (
	"github.com/easymvp-ai/llm"
)

type AgentResponse struct {
	Output any             `json:"output"`
	Usage  *llm.TokenUsage `json:"usage"`
	Cost   *float64
}

type AgentStreamResponse <-chan AgentEvent

type AgentEventType string

const (
	AgentEventTypeText      AgentEventType = "text"
	AgentEventTypeUseTool   AgentEventType = "use-tool"
	AgentEventTypeReasoning AgentEventType = "reasoning"
	AgentEventTypeError     AgentEventType = "error"
)

type AgentEvent struct {
	Type         AgentEventType
	Text         *string
	Reasoning    *string
	ErrorMessage *string
	Input        any
	Output       any
	Partial      bool
}
