package easyagent

import (
	"context"
	_ "embed"
	"encoding/json"
	"github.com/easymvp/easyllm"
	"strings"
)

type AgentEventType string

const (
	AgentEventTypeText      AgentEventType = "text"
	AgentEventTypeUseTool   AgentEventType = "use-tool"
	AgentEventTypeReasoning AgentEventType = "reasoning"
)

type AgentEvent struct {
	// Type indicates whether this is a text or tool_use block
	Type AgentEventType
	// Content contains the text content or tool name for tool_use blocks
	Content string
	// Attributes contains the parameters for tool_use blocks
	Attributes map[string]string
	Partial    bool
}

// String returns a string representation of the MessageBlock
func (b AgentEvent) String() string {
	s := ""
	if b.Type != "" {
		s += string(b.Type) + ": "
	}
	if b.Content != "" {
		s += b.Content + " "
	}
	if len(b.Attributes) > 0 {
		s += "("
		for k, v := range b.Attributes {
			s += k + "=" + v + " "
		}
		s += ")"
	}
	return s
}

type Runner interface {
	Run(ctx context.Context, req *AgentRequest, callback Callback) (*AgentResponse, error)
}

func ToolsPrompts(tools []easyllm.ModelTool) (string, error) {
	if len(tools) == 0 {
		return "No tools", nil
	}
	lines := make([]string, len(tools))
	jsonSchema, _ := json.Marshal(tools)
	for _, tool := range tools {
		line := "## " + tool.Name()
		line += "\nDescription:" + tool.Description()
		line += "\nParams: ```jsonschema\n" + string(jsonSchema) + "\n```"
		line += "\nUsage: ```" + tool.Usage() + "\n```"
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n"), nil
}
