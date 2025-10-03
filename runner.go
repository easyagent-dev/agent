package agent

import (
	"context"
	_ "embed"
)

type Runner interface {
	Run(ctx context.Context, req *AgentRequest) (*AgentResponse, error)
	//RunStream(ctx context.Context, req *AgentRequest, callback Callback) (*AgentResponse, error)
}
