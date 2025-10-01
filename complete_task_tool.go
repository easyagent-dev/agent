package easyagent

import (
	"context"
	"github.com/easymvp/easyllm"
)

const CompleteTaskToolName = "complete_task"

// CompleteTaskTool is a mock implementation of the complete_task tool for testing
type CompleteTaskTool struct {
	outputSchema  any
	outputExample string
}

var _ easyllm.ModelTool = &CompleteTaskTool{}

func NewCompleteTaskTool(outputSchema any, outputExample string) *CompleteTaskTool {
	return &CompleteTaskTool{
		outputSchema:  outputSchema,
		outputExample: outputExample,
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

// ParamsSchema returns the JSON schema for the tool's parameters
func (t *CompleteTaskTool) InputSchema() any {
	return t.outputSchema
}

func (t *CompleteTaskTool) OutputSchema() any {
	return t.outputSchema
}

// Usage returns an example of how to use the tool
func (t *CompleteTaskTool) Usage() string {
	return t.outputExample
}

// Execute runs the tool with the provided parameters
func (t *CompleteTaskTool) Run(ctx context.Context, input any) (any, error) {
	return input, nil
}
