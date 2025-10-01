package easyagent

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/easymvp/easyllm"
)

//go:embed prompts/json_system.md
var jsonSystemPrompt string //nolint:gochecknoglobals

func GetJsonAgentSystemPrompt(instructions string, outputSchema any, message *easyllm.ModelMessage, tools []easyllm.ModelTool) (string, error) {
	toolsPrompt, err := ToolsPrompts(tools)
	if err != nil {
		return "", fmt.Errorf("failed to create tools prompt: %w", err)
	}

	outputSchemaJSON, _ := json.Marshal(outputSchema)
	prompts, err := easyllm.GetPrompts(jsonSystemPrompt, map[string]interface{}{
		"instructions": instructions,
		"tools":        toolsPrompt,
		"userQuery":    message.Content,
		"outputSchema": string(outputSchemaJSON),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get prompts: %w", err)
	}
	return prompts, nil
}
