package agent

import (
	"fmt"
)

// ToolRegistry manages a collection of tools available to an agent
type ToolRegistry struct {
	tools map[string]ModelTool
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]ModelTool),
	}
}

// RegisterTool adds a tool to the registry
func (tr *ToolRegistry) RegisterTool(tool ModelTool) error {
	name := tool.Name()
	if _, exists := tr.tools[name]; exists {
		return fmt.Errorf("tool with name '%s' already registered", name)
	}

	tr.tools[name] = tool
	return nil
}

// UnregisterTool removes a tool from the registry
func (tr *ToolRegistry) UnregisterTool(name string) error {
	if _, exists := tr.tools[name]; !exists {
		return fmt.Errorf("tool with name '%s' not found", name)
	}

	delete(tr.tools, name)
	return nil
}

// GetTool retrieves a tool by name
func (tr *ToolRegistry) GetTool(name string) (ModelTool, error) {
	tool, exists := tr.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool with name '%s' not found", name)
	}

	return tool, nil
}

// GetTools returns all registered tools
func (tr *ToolRegistry) GetTools() []ModelTool {
	tools := make([]ModelTool, 0, len(tr.tools))
	for _, tool := range tr.tools {
		tools = append(tools, tool)
	}
	return tools
}
