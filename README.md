# EasyAgent

A simple and powerful Go framework for building AI agents with tool-calling capabilities.

[![Go Version](https://img.shields.io/badge/Go-1.24%2B-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](LICENSE)

## Quick Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/easymvp-ai/agent"
    "github.com/easymvp-ai/llm"
    "github.com/easymvp-ai/llm/openai"
)

func main() {
    // Create a weather tool
    weatherTool := NewWeatherTool()
    
    // Create an agent
    agentInstance := &agent.CompletionAgent{
        Name:         "Weather Assistant",
        Description:  "An AI assistant that can provide weather information",
        Instructions: "You are a helpful assistant that provides weather information.",
        Tools:        []agent.ModelTool{weatherTool},
        Callback:     agent.NewDefaultCallback(&agent.DefaultLogger{}),
        Logger:       &agent.DefaultLogger{},
    }
    
    // Create OpenAI model
    model, err := openai.NewOpenAIModel(llm.WithAPIKey(os.Getenv("OPENAI_API_KEY")))
    if err != nil {
        log.Fatal(err)
    }
    
    // Create runner and execute
    runner, _ := agent.NewCompletionRunner(agentInstance, model)
    
    resp, err := runner.Run(context.Background(), &agent.AgentRequest{
        Model: "gpt-4o-mini",
        Messages: []*llm.ModelMessage{
            {Role: llm.RoleUser, Content: "What's the weather in Tokyo?"},
        },
        OutputSchema:  llm.GenerateSchema[agent.Reply](),
        MaxIterations: 10,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Result: %+v\n", resp.Output)
}
```

## Features

- 🤖 **Simple Agent Definition** - Define agents with clear instructions and tools
- 🔧 **Tool System** - Easy-to-implement tool interface with automatic registration
- 🔄 **Iterative Execution** - Agents iterate through tool calls to complete tasks
- 🪝 **Lifecycle Callbacks** - Hook into model calls and tool executions
- 📝 **Flexible Logging** - Built-in logging interface with customizable implementations
- ⚡ **Performance Optimized** - Thread-safe, memory-efficient execution

## Installation

```bash
go get github.com/easymvp-ai/agent
```

## Creating Tools

Implement the `ModelTool` interface to create custom tools:

```go
type WeatherTool struct{}

func (t *WeatherTool) Name() string {
    return "get_weather"
}

func (t *WeatherTool) Description() string {
    return "Get current weather for a location"
}

func (t *WeatherTool) InputSchema() any {
    return WeatherInput{}
}

func (t *WeatherTool) OutputSchema() any {
    return WeatherOutput{}
}

func (t *WeatherTool) Usage() string {
    return `{"location": "Tokyo, Japan"}`
}

func (t *WeatherTool) Run(ctx context.Context, input any) (any, error) {
    // Your tool implementation
    return result, nil
}
```

## Architecture

### Core Components

- **CompletionAgent** - Defines the agent's identity, instructions, and tools
- **CompletionRunner** - Executes the agent with iterative tool calling
- **ModelTool** - Interface for implementing custom tools
- **Callback** - Lifecycle hooks for customization
- **Logger** - Flexible logging interface

### Execution Flow

1. User sends a message to the agent
2. Agent analyzes the request and selects appropriate tools
3. Tools are executed and results are processed
4. Process repeats until task is complete or max iterations reached

## Configuration

### Agent Configuration

```go
agent := &agent.CompletionAgent{
    Name:         "Assistant Name",
    Description:  "What this agent does",
    Instructions: "Detailed instructions for the agent",
    Tools:        []agent.ModelTool{tool1, tool2},
    Callback:     agent.NewDefaultCallback(logger),
    Logger:       logger,
}
```

### Request Configuration

```go
req := &agent.AgentRequest{
    Model:         "gpt-4o-mini",     // Model to use
    Messages:      messages,           // Conversation history
    MaxIterations: 10,                 // Max tool call iterations
    OutputSchema:  schema,             // Expected output format
    Options:       []llm.CompletionOption{llm.WithUsage(true)},
}
```

## Advanced Features

### Custom Callbacks

```go
type MyCallback struct {
    logger agent.Logger
}

func (c *MyCallback) BeforeModel(ctx context.Context, provider, model string, req *llm.CompletionRequest) (*llm.CompletionResponse, error) {
    // Log or modify request
    return nil, nil
}

func (c *MyCallback) BeforeToolCall(ctx context.Context, toolName string, input any) (any, error) {
    // Validate or modify tool input
    return nil, nil
}
```

### Custom Logger

```go
type MyLogger struct{}

func (l *MyLogger) Info(msg string, fields ...interface{}) {
    // Your logging logic
}

func (l *MyLogger) Debug(msg string, fields ...interface{}) {}
func (l *MyLogger) Warn(msg string, fields ...interface{}) {}
func (l *MyLogger) Error(msg string, fields ...interface{}) {}
```

## Error Handling

```go
if err != nil {
    if errors.Is(err, agent.ErrToolNotFound) {
        // Handle tool not found
    } else if errors.Is(err, agent.ErrMaxIterations) {
        // Handle max iterations reached
    }
}
```

## Examples

See the [examples/](examples/) directory for complete examples:
- **weather_agent** - Basic agent with tool calling
- **stream_weather_agent** - Streaming agent example
- **deepseek_stream_weather_agent** - DeepSeek integration

## Best Practices

1. **Input Validation** - Always validate agent and request configurations
2. **Error Handling** - Use structured error types for better debugging
3. **Logging** - Implement appropriate logging for production deployments
4. **Tool Design** - Keep tools focused and single-purpose
5. **Context Usage** - Always respect context cancellation

## License

Apache License 2.0 - See [LICENSE](LICENSE) file for details.

## Support

- GitHub Issues: [Report bugs or request features](https://github.com/easymvp/easyagent/issues)
- Documentation: [Full documentation](https://docs.easymvp.ai/agent)
