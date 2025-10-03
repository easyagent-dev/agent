package agent

import "github.com/easymvp-ai/llm"

type AgentRequest struct {
	Model         string
	OutputSchema  any
	OutputUsage   string
	Messages      []*llm.ModelMessage
	Options       []llm.CompletionOption
	MaxIterations int
}
