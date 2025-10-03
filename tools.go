package agent

import "context"

// ModelTool defines the interface that all agent tools must implement.
// Tools are the primary way agents interact with external systems and perform actions.
type ModelTool interface {
	// Name returns the unique identifier for this tool
	Name() string

	// Description returns a human-readable description of what the tool does
	Description() string

	// InputSchema returns the JSON schema for the tool's input parameters
	// This should be a struct that can be marshaled to JSON schema
	InputSchema() any

	// OutputSchema returns the JSON schema for the tool's output
	// This should be a struct that can be marshaled to JSON schema
	OutputSchema() any

	// Run executes the tool with the given input and returns the result
	// The input will be unmarshaled according to InputSchema
	// The context can be used for cancellation and timeouts
	Run(ctx context.Context, input any) (any, error)

	// Usage returns an example of how to use the tool in JSON format
	Usage() string
}
