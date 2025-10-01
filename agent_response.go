package easyagent

import (
	"github.com/easymvp/easyllm"
)

type AgentResponse struct {
	Output any                 `json:"output"`
	Usage  *easyllm.TokenUsage `json:"usage"`
	Cost   *float64
}
