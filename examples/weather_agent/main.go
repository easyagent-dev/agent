package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"

	"github.com/easymvp-ai/agent"
	"github.com/easymvp-ai/llm"
	"github.com/easymvp-ai/llm/openai"
)

// WeatherTool provides weather information for a given location
type WeatherTool struct{}

var _ agent.ModelTool = &WeatherTool{}

// WeatherInput defines the input schema for the weather tool
type WeatherInput struct {
	Location string `json:"location" jsonschema:"required,description=The city and country (e.g. 'London, UK' or 'New York, USA')"`
}

// WeatherOutput defines the output schema for the weather tool
type WeatherOutput struct {
	Location    string  `json:"location"`
	Temperature float64 `json:"temperature"`
	Condition   string  `json:"condition"`
	Humidity    int     `json:"humidity"`
	WindSpeed   float64 `json:"wind_speed"`
}

func NewWeatherTool() *WeatherTool {
	return &WeatherTool{}
}

func (t *WeatherTool) Name() string {
	return "get_weather"
}

func (t *WeatherTool) Description() string {
	return "Get current weather information for a specified location"
}

func (t *WeatherTool) InputSchema() any {
	return WeatherInput{}
}

func (t *WeatherTool) OutputSchema() any {
	return WeatherOutput{}
}

func (t *WeatherTool) Usage() string {
	return `Example usage:
{
  "location": "London, UK"
}`
}

func (t *WeatherTool) Run(ctx context.Context, input any) (any, error) {
	// Convert input to WeatherInput
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	var weatherInput WeatherInput
	if err := json.Unmarshal(inputBytes, &weatherInput); err != nil {
		return nil, fmt.Errorf("failed to unmarshal input: %w", err)
	}

	if weatherInput.Location == "" {
		return nil, errors.New("location is required")
	}

	// Generate mock weather data
	conditions := []string{"Sunny", "Cloudy", "Rainy", "Partly Cloudy", "Stormy", "Snowy"}
	temperature := 10.0 + rand.Float64()*25.0 // Random temp between 10-35°C
	humidity := 30 + rand.Intn(60)            // Random humidity between 30-90%
	windSpeed := rand.Float64() * 30.0        // Random wind speed 0-30 km/h

	output := WeatherOutput{
		Location:    weatherInput.Location,
		Temperature: float64(int(temperature*10)) / 10, // Round to 1 decimal
		Condition:   conditions[rand.Intn(len(conditions))],
		Humidity:    humidity,
		WindSpeed:   float64(int(windSpeed*10)) / 10, // Round to 1 decimal
	}

	return output, nil
}

func main() {
	// Get OpenAI API key from environment variable
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is not set")
	}

	// Create a weather tool
	weatherTool := NewWeatherTool()

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
