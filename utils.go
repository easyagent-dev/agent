package agent

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/easymvp-ai/llm"
	"strings"
	"sync"
	"text/template"
)

//go:embed prompts/json_system.md
var jsonSystemPrompt string //nolint:gochecknoglobals

func GetJsonAgentSystemPrompt(instructions string, outputSchema any, message *llm.ModelMessage, tools []ModelTool) (string, error) {
	toolsPrompt, err := ToolsPrompts(tools)
	if err != nil {
		return "", fmt.Errorf("failed to create tools prompt: %w", err)
	}

	outputSchemaJSON, _ := json.Marshal(outputSchema)
	prompts, err := GetPrompts(jsonSystemPrompt, map[string]interface{}{
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

func ToolsPrompts(tools []ModelTool) (string, error) {
	if len(tools) == 0 {
		return "No tools", nil
	}

	// Use strings.Builder for efficient string concatenation
	var builder strings.Builder
	builder.Grow(len(tools) * 256) // Pre-allocate reasonable size

	for i, tool := range tools {
		if i > 0 {
			builder.WriteString("\n")
		}
		inputSchema, _ := json.Marshal(tool.InputSchema())
		builder.WriteString("## ")
		builder.WriteString(tool.Name())
		builder.WriteString("\nDescription:")
		builder.WriteString(tool.Description())
		builder.WriteString("\nInput: ```jsonschema\n")
		builder.Write(inputSchema)
		builder.WriteString("\n```\nUsage: ```")
		builder.WriteString(tool.Usage())
		builder.WriteString("\n```")
	}
	return builder.String(), nil
}

// Template cache for better performance
var (
	templateCache = make(map[string]*template.Template)
	templateMutex sync.RWMutex
)

// GetPrompts executes a template with caching for better performance
func GetPrompts(prompt string, params map[string]interface{}) (string, error) {
	// Try to get cached template first (read lock)
	templateMutex.RLock()
	tmpl, exists := templateCache[prompt]
	templateMutex.RUnlock()

	if !exists {
		// Parse and cache the template (write lock)
		templateMutex.Lock()
		// Double-check in case another goroutine added it
		if tmpl, exists = templateCache[prompt]; !exists {
			var err error
			tmpl, err = template.New("prompt").Parse(prompt)
			if err != nil {
				templateMutex.Unlock()
				return "", fmt.Errorf("failed to parse template: %w", err)
			}
			templateCache[prompt] = tmpl
		}
		templateMutex.Unlock()
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
