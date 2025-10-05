package agent

import (
	"errors"
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
