package agent

import "fmt"

// Logger defines the interface for logging within the agent framework
type Logger interface {
	// Debug logs a debug message with optional fields
	Debug(msg string, fields ...interface{})

	// Info logs an informational message with optional fields
	Info(msg string, fields ...interface{})

	// Warn logs a warning message with optional fields
	Warn(msg string, fields ...interface{})

	// Error logs an error message with optional fields
	Error(msg string, fields ...interface{})
}

// NoOpLogger is a logger that does nothing
// Use this when you don't want any logging
type NoOpLogger struct{}

// Debug does nothing
func (l *NoOpLogger) Debug(msg string, fields ...interface{}) {}

// Info does nothing
func (l *NoOpLogger) Info(msg string, fields ...interface{}) {}

// Warn does nothing
func (l *NoOpLogger) Warn(msg string, fields ...interface{}) {}

// Error does nothing
func (l *NoOpLogger) Error(msg string, fields ...interface{}) {}

// DefaultLogger is a simple logger that writes to stdout
type DefaultLogger struct{}

// Debug logs a debug message
func (l *DefaultLogger) Debug(msg string, fields ...interface{}) {
	fmt.Printf("[DEBUG] %s", msg)
	if len(fields) > 0 {
		fmt.Printf(" %v", fields)
	}
	fmt.Println()
}

// Info logs an informational message
func (l *DefaultLogger) Info(msg string, fields ...interface{}) {
	fmt.Printf("[INFO] %s", msg)
	if len(fields) > 0 {
		fmt.Printf(" %v", fields)
	}
	fmt.Println()
}

// Warn logs a warning message
func (l *DefaultLogger) Warn(msg string, fields ...interface{}) {
	fmt.Printf("[WARN] %s", msg)
	if len(fields) > 0 {
		fmt.Printf(" %v", fields)
	}
	fmt.Println()
}

// Error logs an error message
func (l *DefaultLogger) Error(msg string, fields ...interface{}) {
	fmt.Printf("[ERROR] %s", msg)
	if len(fields) > 0 {
		fmt.Printf(" %v", fields)
	}
	fmt.Println()
}
