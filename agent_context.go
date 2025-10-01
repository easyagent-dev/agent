package easyagent

import (
	"context"
	"github.com/easymvp/easyllm"
)

const ContextKey = "agent"

func AgentContextOf(ctx context.Context) *AgentContext {
	return ctx.Value(ContextKey).(*AgentContext)
}

func WithAgentContext(ctx context.Context, ac *AgentContext) context.Context {
	return context.WithValue(ctx, ContextKey, ac)
}

type AgentContext struct {
	Agent     *Agent
	Messages  []*easyllm.ModelMessage
	ToolCalls []*easyllm.ToolCall
	Callback  Callback
	Session   map[string]any
}

func (ac *AgentContext) IsToolCalled(name string) bool {
	for _, toolCall := range ac.ToolCalls {
		if toolCall.Name == name {
			return true
		}
	}
	return false
}
