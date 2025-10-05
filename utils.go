package agent

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/easyagent-dev/llm"
	"strings"
)

//go:embed prompts/json_system.md
var jsonSystemPrompt string //nolint:gochecknoglobals

func GetJsonAgentSystemPrompt(agent *Agent, message *llm.ModelMessage, tools []ModelTool) (string, error) {
	toolsPrompt, err := ToolsPrompts(tools)
	if err != nil {
		return "", fmt.Errorf("failed to create tools prompt: %w", err)
	}

	prompts, err := llm.GetPrompts(jsonSystemPrompt, map[string]interface{}{
		"agent":     agent,
		"tools":     toolsPrompt,
		"userQuery": message.Content,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get prompts: %w", err)
	}
	return prompts, nil
}

func ToolsPrompts(tools []ModelTool) (string, error) {
	if len(tools) == 0 {
		return "No tools available", nil
	}

	// Use strings.Builder for efficient string concatenation
	var builder strings.Builder
	builder.Grow(len(tools) * 256) // Pre-allocate reasonable size

	for i, tool := range tools {
		if i > 0 {
			builder.WriteString("\n")
		}
		inputSchema, _ := json.Marshal(tool.InputSchema())
		builder.WriteString("<tool name=\"")
		builder.WriteString(tool.Name())
		builder.WriteString("\">\n<description>")
		builder.WriteString(tool.Description())
		builder.WriteString("</description>\n<input_schema>\n")
		builder.Write(inputSchema)
		builder.WriteString("\n</input_schema>")

		usage := tool.Usage()
		if usage != "" {
			builder.WriteString("\n<usage>\n")
			builder.WriteString(usage)
			builder.WriteString("\n</usage>")
		}
		builder.WriteString("\n</tool>")
	}
	return builder.String(), nil
}
