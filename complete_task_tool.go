package agent

import (
	"context"
)

const CompleteTaskToolName = "complete_task"

// CompleteTaskTool is a mock implementation of the complete_task tool for testing
type CompleteTaskTool struct {
	outputSchema any
	usage        string
}

var _ ModelTool = &CompleteTaskTool{}

func NewCompleteTaskTool(outputSchema any, usage string) *CompleteTaskTool {
	return &CompleteTaskTool{
		outputSchema: outputSchema,
		usage:        usage,
	}
}

// Name returns the name of the tool
func (t *CompleteTaskTool) Name() string {
	return CompleteTaskToolName
}

// Description returns a description of what the tool does
func (t *CompleteTaskTool) Description() string {
	return "Completes the user query and output the final results"
}

// GenerateSchema generates a JSON schema from the InputType
func (t *CompleteTaskTool) InputSchema() any {
	return t.outputSchema
}

func (t *CompleteTaskTool) OutputSchema() any {
	return nil
}

// Usage returns an example of how to use the tool
func (t *CompleteTaskTool) Usage() string {
	return t.usage
}

// Execute runs the tool with the provided parameters
func (t *CompleteTaskTool) Run(ctx context.Context, input map[string]any) (any, error) {
	return input, nil
}
