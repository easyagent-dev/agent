package agent

import "errors"

// Agent represents an AI agent with specific capabilities and behaviors.
// It encapsulates the agent's identity, instructions, available tools,
// callback handlers, and logging configuration.
type Agent struct {
	// Name is the identifier for this agent
	Name string

	// Description provides a brief explanation of the agent's purpose
	Description string

	// Instructions contain the system prompt or guidelines for the agent
	Instructions string

	// Tools are the available tools this agent can use
	Tools []ModelTool

	// Callback handles lifecycle events during agent execution
	Callback Callback

	// Logger handles logging for this agent (optional, defaults to NoOpLogger)
	Logger Logger
}

// Validate validates the agent configuration
func (a *Agent) Validate() error {
	if a.Name == "" {
		return errors.New("agent name is required")
	}
	if a.Description == "" {
		return errors.New("agent description is required")
	}
	if a.Instructions == "" {
		return errors.New("agent instructions are required")
	}
	if a.Callback == nil {
		return errors.New("agent callback is required")
	}
	// Logger is optional, will default to NoOpLogger if not set
	return nil
}

// GetLogger returns the agent's logger or a NoOpLogger if none is set
func (a *Agent) GetLogger() Logger {
	if a.Logger == nil {
		return &NoOpLogger{}
	}
	return a.Logger
}
