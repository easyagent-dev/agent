package agent

import (
	"fmt"
	"sync"
)

// ToolRegistry manages a collection of tools available to an agent
// It is safe for concurrent use by multiple goroutines
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]ModelTool
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]ModelTool),
	}
}

// RegisterTool adds a tool to the registry
// It returns an error if a tool with the same name already exists
func (tr *ToolRegistry) RegisterTool(tool ModelTool) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	name := tool.Name()
	if _, exists := tr.tools[name]; exists {
		return fmt.Errorf("tool with name '%s' already registered", name)
	}

	tr.tools[name] = tool
	return nil
}

// UnregisterTool removes a tool from the registry
// It returns an error if the tool is not found
func (tr *ToolRegistry) UnregisterTool(name string) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	if _, exists := tr.tools[name]; !exists {
		return fmt.Errorf("tool with name '%s' not found", name)
	}

	delete(tr.tools, name)
	return nil
}

// GetTool retrieves a tool by name
// It returns an error if the tool is not found
func (tr *ToolRegistry) GetTool(name string) (ModelTool, error) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	tool, exists := tr.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool with name '%s' not found", name)
	}

	return tool, nil
}

// GetTools returns all registered tools
// The returned slice is a copy and safe to modify
func (tr *ToolRegistry) GetTools() []ModelTool {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	tools := make([]ModelTool, 0, len(tr.tools))
	for _, tool := range tr.tools {
		tools = append(tools, tool)
	}
	return tools
}
