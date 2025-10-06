package agent

import (
	"context"
	_ "embed"
)

type Runner interface {
	Run(ctx context.Context, req *AgentRequest, callback Callback) (*AgentResponse, error)
}

type StreamRunner interface {
	Run(ctx context.Context, req *AgentRequest, callback Callback) (*AgentStreamResponse, error)
}
