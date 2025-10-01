package easyagent

import "github.com/easymvp/easyllm"

type AgentStep struct {
	Reasoning string            `json:"reasoning" jsonschema:"title=Reasoning,description=The reasoning process of the agent before making a tool call,required"`
	ToolCall  *easyllm.ToolCall `json:"toolCall" jsonschema:"title=Tool Call,description=The tool call to be executed by the agent,required"`
}
