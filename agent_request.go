package agent

import (
	"errors"
	"github.com/easyagent-dev/llm"
)

// AgentRequest represents a request to execute an agent with specific parameters.
// It contains the model configuration, conversation history, and execution constraints.
type AgentRequest struct {
	// OutputSchema defines the expected structure of the final output
	// This should be a struct that can be marshaled to JSON schema
	OutputSchema any

	// OutputUsage provides an example or description of how to use the output
	OutputUsage string

	// Messages is the conversation history to provide context to the agent
	// Must contain at least one message, with the last message from the user
	Messages []*llm.ModelMessage

	// MaxIterations is the maximum number of tool-calling iterations allowed
	// Must be positive. Prevents infinite loops in agent execution.
	MaxIterations int

	// MaxRetries is the maximum number of consecutive retries allowed when errors occur
	// If 0 or negative, no retry limit is enforced
	MaxRetries int
}

// Validate validates the agent request parameters and returns an error if invalid.
// It checks that all required fields are set and have valid values.
func (r *AgentRequest) Validate() error {
	if len(r.Messages) == 0 {
		return errors.New("at least one message is required")
	}
	if r.MaxIterations <= 0 {
		return errors.New("max iterations must be positive")
	}
	// Validate last message is from user
	if r.Messages[len(r.Messages)-1].Role != llm.RoleUser {
		return errors.New("last message must be from user")
	}
	return nil
}
