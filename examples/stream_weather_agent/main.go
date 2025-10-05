package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/easyagent-dev/agent/examples"
	"github.com/easyagent-dev/llm/providers"
	"log"
	"os"

	"github.com/easyagent-dev/agent"
	"github.com/easyagent-dev/llm"
)

func main() {
	// Get OpenAI API key from environment variable
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is not set")
	}

	// Create a weather tool
	weatherTool := examples.NewWeatherTool()

	// Create an agent with the weather tool
	agentInstance := &agent.Agent{
		Name:         "Weather Assistant",
		Description:  "An AI assistant that can provide weather information",
		Instructions: "You are a helpful assistant that provides weather information for any location requested by the user.",
		Tools:        []agent.ModelTool{weatherTool},
	}

	// Create OpenAI model
	provider, err := providers.NewOpenAIModelProvider(llm.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	model, err := provider.NewCompletionModel("o4-mini")
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}
	// Create a completion runner
	runner, err := agent.NewCompletionRunner(agentInstance, model)
	if err != nil {
		log.Fatalf("Failed to create runner: %v", err)
	}

	// Create an agent request
	req := &agent.AgentRequest{
		Messages: []*llm.ModelMessage{
			{
				Role:    llm.RoleUser,
				Content: "What's the weather like in Tokyo?",
			},
		},
		OutputSchema:  llm.GenerateSchema[examples.Reply](),
		OutputUsage:   "",
		MaxIterations: 10,
	}

	// Run the agent with streaming
	ctx := context.Background()
	streamResp, err := runner.StreamRun(ctx, req, llm.WithUsage(true), llm.WithCost(true), llm.WithMaxTokens(1000))
	if err != nil {
		log.Fatalf("Failed to start streaming: %v", err)
	}

	// Process streaming events
	fmt.Printf("\n=== Streaming Agent Events ===\n")
	for event := range *streamResp {
		switch event.Type {
		case agent.AgentEventTypeUseTool:
			if event.Partial {
				// Partial tool call - show progress
				fmt.Printf("  [Partial Tool Call]\n")
				if event.ToolCall != nil {
					inputJSON, _ := json.MarshalIndent(event.ToolCall, "    ", "  ")
					fmt.Printf("    Partial Input: %s\n", string(inputJSON))
				}
			} else {
				// Complete tool call
				fmt.Printf("  [Complete Tool Call]\n")
				if event.ToolCall != nil {
					inputJSON, _ := json.MarshalIndent(event.ToolCall, "    ", "  ")
					fmt.Printf("    Input: %s\n", string(inputJSON))
				}
			}
		case agent.AgentEventTypeError:
			if event.ErrorMessage != nil {
				fmt.Printf("  [Error] %s\n", *event.ErrorMessage)
			}
		}
	}

	fmt.Printf("\n=== Streaming Complete ===\n")
}
