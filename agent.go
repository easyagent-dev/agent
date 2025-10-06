package agent

import "errors"

type AgentType string

const (
	AgentTypeJSON AgentType = "json"
	AgentTypeXml  AgentType = "xml"
)

// Agent represents an AI agent with specific capabilities and behaviors.
// It encapsulates the agent's identity, instructions, available tools,
// callback handlers, and logging configuration.
type Agent struct {
	// Name is the identifier for this agent
	Name string

	// ModelProvider is the model provider
	ModelProvider string

	// Model is the model provider
	Model string

	// Description provides a brief explanation of the agent's purpose
	Description string

	// Instructions contain the system prompt or guidelines for the agent
	Instructions string

	//Type is the type of agent this is
	Type AgentType

	// Tools are the available tools this agent can use
	Tools []ModelTool
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
	// Logger is optional, will default to NoOpLogger if not set
	return nil
}
