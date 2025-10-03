# EasyAgent - Go Agent Framework

A powerful and flexible Go framework for building AI agents with tool-calling capabilities.

[![Go Version](https://img.shields.io/badge/Go-1.24%2B-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](LICENSE)

## Features

- 🤖 **Simple Agent Definition** - Define agents with clear instructions and tools
- 🔧 **Tool System** - Easy-to-implement tool interface with automatic registration
- 🔄 **Iterative Execution** - Agents iterate through tool calls to complete tasks
- 🪝 **Lifecycle Callbacks** - Hook into model calls and tool executions
- 📝 **Flexible Logging** - Built-in logging interface with customizable implementations
- ⚡ **Performance Optimized** - Thread-safe, memory-efficient, with bounded message history
- 🎯 **Type-Safe** - Leverages Go's type system for safety and clarity

## Installation

```bash
go get github.com/easymvp-ai/agent
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/easymvp-ai/agent"
    "github.com/easymvp-ai/llm"
    "github.com/easymvp-ai/llm/openai"
)

func main() {
    // Create your tools
    weatherTool := NewWeatherTool()
    
    // Create an agent
    agentInstance := &agent.Agent{
        Name:         "Weather Assistant",
        Description:  "An AI assistant that provides weather information",
        Instructions: "You are a helpful assistant that provides weather information.",
        Tools:        []agent.ModelTool{weatherTool},
        Callback:     agent.NewDefaultCallback(&agent.DefaultLogger{}),
        Logger:       &agent.DefaultLogger{},
    }
    
    // Create LLM model
    model, err := openai.NewOpenAIModel(llm.WithAPIKey("your-api-key"))
    if err != nil {
        log.Fatal(err)
    }
    
    // Create runner
    runner, err := agent.NewCompletionRunner(agentInstance, model)
    if err != nil {
        log.Fatal(err)
    }
    
    // Execute agent
    req := &agent.AgentRequest{
        Model: "gpt-4",
        Messages: []*llm.ModelMessage{
            {Role: llm.RoleUser, Content: "What's the weather in Tokyo?"},
        },
        MaxIterations: 10,
    }
    
    resp, err := runner.Run(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Result: %+v\n", resp.Output)
}
```

## Creating Tools

Tools are the primary way agents interact with external systems. Implement the `ModelTool` interface:

```go
type WeatherTool struct{}

func (t *WeatherTool) Name() string {
    return "get_weather"
}

func (t *WeatherTool) Description() string {
    return "Get current weather for a location"
}

func (t *WeatherTool) InputSchema() any {
    return WeatherInput{} // Your input struct
}

func (t *WeatherTool) OutputSchema() any {
    return WeatherOutput{} // Your output struct
}

func (t *WeatherTool) Usage() string {
    return `{"location": "Tokyo, Japan"}`
}

func (t *WeatherTool) Run(ctx context.Context, input any) (any, error) {
    // Convert input and execute your logic
    // Return structured output
}
```

## Architecture

### Agent Components

- **Agent**: Defines the agent's identity, instructions, tools, and configuration
- **Runner**: Executes the agent with iterative tool calling
- **Tools**: Extensible tool system for external integrations
- **Callbacks**: Lifecycle hooks for customization
- **Logger**: Flexible logging interface

### Execution Flow

1. User sends a message to the agent
2. Agent analyzes the request and instructions
3. Agent selects and calls appropriate tools
4. Results are processed and fed back to the agent
5. Process repeats until task is complete or max iterations reached

## Configuration

### Agent Configuration

```go
agent := &agent.Agent{
    Name:         "Assistant Name",
    Description:  "What this agent does",
    Instructions: "Detailed instructions for the agent",
    Tools:        []agent.ModelTool{tool1, tool2},
    Callback:     agent.NewDefaultCallback(logger),
    Logger:       logger, // Optional
}
```

### Request Configuration

```go
req := &agent.AgentRequest{
    Model:         "gpt-4",           // Model to use
    Messages:      messages,          // Conversation history
    MaxIterations: 10,                // Max tool call iterations
    OutputSchema:  schema,            // Expected output format
}
```

## Advanced Features

### Custom Callbacks

Implement the `Callback` interface to customize behavior:

```go
type MyCallback struct {
    logger agent.Logger
}

func (c *MyCallback) BeforeModel(ctx context.Context, provider, model string, req *llm.CompletionRequest) (*llm.CompletionResponse, error) {
    // Log or modify request
    return nil, nil // Return nil to continue normally
}

func (c *MyCallback) BeforeToolCall(ctx context.Context, toolName string, input any) (any, error) {
    // Validate or modify tool input
    return nil, nil // Return value to override tool execution
}
```

### Custom Logger

Implement the `Logger` interface for your logging system:

```go
type MyLogger struct {
    // Your logger implementation
}

func (l *MyLogger) Info(msg string, fields ...interface{}) {
    // Your logging logic
}

// Implement other methods: Debug, Warn, Error
```

### Message History Management

The framework automatically manages message history to prevent memory issues:

```go
// Default limit is 100 messages
// Automatically trims to keep recent history
// Configurable via CompletionRunner
```

## Error Handling

The framework provides structured error types:

```go
import "errors"

if err != nil {
    if errors.Is(err, agent.ErrToolNotFound) {
        // Handle tool not found
    } else if errors.Is(err, agent.ErrMaxIterations) {
        // Handle max iterations reached
    }
}

// Or use AgentError for detailed context
if agentErr, ok := err.(*agent.AgentError); ok {
    fmt.Printf("Type: %s, Message: %s, Context: %v\n", 
        agentErr.Type, agentErr.Message, agentErr.Context)
}
```

## Performance Considerations

- **Thread-Safe**: ToolRegistry is safe for concurrent use
- **Memory Efficient**: Automatic message history trimming (default: 100 messages)
- **Optimized String Building**: Uses `strings.Builder` for efficient concatenation
- **Context Cancellation**: Supports context cancellation for graceful shutdown

## Best Practices

1. **Input Validation**: Always validate agent and request configurations
2. **Error Handling**: Use structured error types for better debugging
3. **Logging**: Implement appropriate logging for production deployments
4. **Tool Design**: Keep tools focused and single-purpose
5. **Message History**: Monitor and adjust `maxMessageHistory` based on your use case
6. **Context Usage**: Always respect context cancellation

## Examples

See the [examples/](examples/) directory for complete examples:

- **Weather Agent**: Basic agent with tool calling
- More examples coming soon!

## API Reference

### Core Types

- `Agent` - Agent configuration and definition
- `AgentRequest` - Request parameters for agent execution
- `AgentResponse` - Response from agent execution
- `ModelTool` - Interface for implementing tools
- `Callback` - Lifecycle hooks interface
- `Logger` - Logging interface

### Key Functions

- `NewCompletionRunner(agent, model)` - Create an agent runner
- `runner.Run(ctx, req)` - Execute the agent
- `agent.Validate()` - Validate agent configuration
- `req.Validate()` - Validate request parameters

## Contributing

Contributions are welcome! Please read our contributing guidelines and submit pull requests.

## License

Apache License 2.0 - See [LICENSE](LICENSE) file for details.

## Support

- GitHub Issues: [Report bugs or request features](https://github.com/easymvp/easyagent/issues)
- Documentation: [Full documentation](https://docs.easymvp.ai/agent)

## Changelog

### Recent Improvements (2025-01-04)

- ✅ Fixed critical bug in ToolsPrompts string building
- ✅ Added thread-safe ToolRegistry with sync.RWMutex
- ✅ Implemented input validation for Agent and AgentRequest
- ✅ Added message history limiting to prevent memory leaks
- ✅ Introduced structured error types (AgentError)
- ✅ Added flexible logging interface
- ✅ Improved context key type safety
- ✅ Added comprehensive documentation

See [CODE_REVIEW.md](CODE_REVIEW.md) for detailed optimization recommendations.
