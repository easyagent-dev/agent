package agent

import (
	"context"
	_ "embed"
)

type Runner interface {
	Run(ctx context.Context, req *AgentRequest) (*AgentResponse, error)
}

type StreamRunner interface {
	RunStream(ctx context.Context, req *AgentRequest) (*AgentStreamResponse, error)
}
