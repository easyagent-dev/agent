package easyagent

import (
	"github.com/easymvp/easyllm"
)

type AgentRequest struct {
	ModelProvider string
	Model         string
	Config        *easyllm.ModelConfig
	Messages      []*easyllm.ModelMessage
	MaxIterations int
}
