package agent

import (
	"errors"
	"fmt"
)

// Common error types for agent operations
var (
	// ErrToolNotFound is returned when a requested tool is not in the registry
	ErrToolNotFound = errors.New("tool not found")

	// ErrToolExecution is returned when a tool execution fails
	ErrToolExecution = errors.New("tool execution failed")

	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")

	// ErrContextCancelled is returned when the context is cancelled
	ErrContextCancelled = errors.New("context cancelled")

	// ErrMaxIterations is returned when max iterations is reached without completion
	ErrMaxIterations = errors.New("max iterations reached")

	// ErrInvalidConfiguration is returned when agent or request configuration is invalid
	ErrInvalidConfiguration = errors.New("invalid configuration")

	// ErrToolAlreadyRegistered is returned when attempting to register a duplicate tool
	ErrToolAlreadyRegistered = errors.New("tool already registered")
)

// AgentError represents a detailed error from agent operations
type AgentError struct {
	Type    string                 // Error type (e.g., "ToolNotFound", "ValidationError")
	Message string                 // Human-readable error message
	Err     error                  // Underlying error, if any
	Context map[string]interface{} // Additional context
}

// Error implements the error interface
func (e *AgentError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *AgentError) Unwrap() error {
	return e.Err
}

// NewAgentError creates a new AgentError
func NewAgentError(errType string, message string, err error) *AgentError {
	return &AgentError{
		Type:    errType,
		Message: message,
		Err:     err,
		Context: make(map[string]interface{}),
	}
}

// WithContext adds context to the error
func (e *AgentError) WithContext(key string, value interface{}) *AgentError {
	e.Context[key] = value
	return e
}

// ValidationError creates a validation error
func ValidationError(message string) *AgentError {
	return NewAgentError("ValidationError", message, ErrInvalidInput)
}

// ToolNotFoundError creates a tool not found error
func ToolNotFoundError(toolName string, availableTools []string) *AgentError {
	err := NewAgentError("ToolNotFound", fmt.Sprintf("tool '%s' not found", toolName), ErrToolNotFound)
	err.WithContext("toolName", toolName)
	err.WithContext("availableTools", availableTools)
	return err
}

// ToolExecutionError creates a tool execution error
func ToolExecutionError(toolName string, err error) *AgentError {
	return NewAgentError("ToolExecutionError", fmt.Sprintf("tool '%s' execution failed", toolName), err).
		WithContext("toolName", toolName)
}
