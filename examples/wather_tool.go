package examples

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/easyagent-dev/agent"
	"math/rand"
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

func (t *WeatherTool) Run(ctx context.Context, input map[string]any) (any, error) {
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
