package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/easyagent-dev/agent/examples/tools"
	"log"
	"os"

	"github.com/easyagent-dev/agent"
	"github.com/easyagent-dev/llm"
	"github.com/easyagent-dev/llm/openai"
)

func main() {
	// Get OpenAI API key from environment variable
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is not set")
	}

	// Create a weather tool
	weatherTool := tools.NewWeatherTool()

	// Create an agent with the weather tool
	agentInstance := &agent.CompletionAgent{
		Name:         "Weather Assistant",
		Description:  "An AI assistant that can provide weather information",
		Instructions: "You are a helpful assistant that provides weather information for any location requested by the user.",
		Tools:        []agent.ModelTool{weatherTool},
		Callback:     agent.NewDefaultCallback(&agent.DefaultLogger{}),
		Logger:       &agent.DefaultLogger{},
	}

	// Create OpenAI model
	model, err := openai.NewOpenAIModel(llm.WithAPIKey(apiKey))
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
		Model: "o4-mini",
		Messages: []*llm.ModelMessage{
			{
				Role:    llm.RoleUser,
				Content: "What's the weather like in Tokyo?",
			},
		},
		OutputSchema:  llm.GenerateSchema[agent.Reply](),
		OutputUsage:   "",
		MaxIterations: 10,
		Options:       []llm.CompletionOption{llm.WithUsage(true), llm.WithCost(true), llm.WithMaxTokens(1000)},
	}

	// Run the agent
	ctx := context.Background()
	resp, err := runner.Run(ctx, req)
	if err != nil {
		log.Fatalf("Failed to run agent: %v", err)
	}

	// Print the response
	fmt.Printf("\n=== Agent Response ===\n")
	output, _ := json.MarshalIndent(resp.Output, "", "  ")
	fmt.Printf("Output: %s\n", string(output))
	fmt.Printf("Token Usage: %+v\n", resp.Usage)
	if resp.Cost != nil {
		fmt.Printf("Cost: $%.4f\n", *resp.Cost)
	}
}
