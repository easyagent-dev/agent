package agent

import (
	"context"
	_ "embed"
)

type Runner interface {
	Run(ctx context.Context, req *AgentRequest) (*AgentResponse, error)
	StreamRun(ctx context.Context, req *AgentRequest) (*AgentStreamResponse, error)
}
