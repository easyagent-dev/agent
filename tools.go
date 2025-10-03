package agent

import "context"

// ModelTool defines the interface that all agent tools must implement.
// Tools are the primary way agents interact with external systems and perform actions.
type ModelTool interface {
	// Name returns the unique identifier for this tool
	Name() string

	// Description returns a human-readable description of what the tool does
	Description() string

	// InputSchema returns the Go type for the tool's output
	// This should be a struct type that can be used for JSON marshaling
	InputSchema() any

	// OutputSchema generates a JSON schema from the InputType
	// This method supports jsonschema struct tags for annotations
	// It can be overridden by tools that need custom schema generation
	OutputSchema() any

	// Run executes the tool with the given input and returns the result
	// The input will be unmarshaled according to InputSchema
	// The context can be used for cancellation and timeouts
	Run(ctx context.Context, input map[string]any) (any, error)

	// Usage returns an example of how to use the tool in JSON format
	Usage() string
}
