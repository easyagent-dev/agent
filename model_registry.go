package easyagent

import (
	"fmt"
	"os"
	"sync"

	"github.com/easymvp/easyllm"
)

type ModelRegistry struct {
	models map[string]easyllm.Model
	mu     sync.RWMutex
}

func NewModelRegistry() *ModelRegistry {
	registry := &ModelRegistry{
		models: make(map[string]easyllm.Model),
	}

	// Auto-register models based on environment variables
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		config := easyllm.OpenAIModelConfig{
			APIKey: apiKey,
		}
		if model, err := easyllm.NewOpenAIModel(config); err == nil && model != nil {
			registry.RegisterModel("openai", model)
		}
	}

	if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
		config := easyllm.OpenRouterModelConfig{
			APIKey: apiKey,
		}
		if model, err := easyllm.NewOpenRouterModel(config); err == nil && model != nil {
			registry.RegisterModel("openrouter", model)
		}
	}

	if apiKey := os.Getenv("DEEPSEEK_API_KEY"); apiKey != "" {
		config := easyllm.DeepSeekModelConfig{
			APIKey: apiKey,
		}
		if model, err := easyllm.NewDeepSeekModel(config); err == nil && model != nil {
			registry.RegisterModel("deepseek", model)
		}
	}

	if apiKey := os.Getenv("CLAUDE_API_KEY"); apiKey != "" {
		config := easyllm.ClaudeModelConfig{
			APIKey: apiKey,
		}
		if model, err := easyllm.NewClaudeModel(config); err == nil && model != nil {
			registry.RegisterModel("claude", model)
		}
	}

	return registry
}

func (r *ModelRegistry) GetModel(modelName string) (easyllm.Model, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	model, ok := r.models[modelName]
	if !ok {
		return nil, fmt.Errorf("model %s not found in registry", modelName)
	}

	return model, nil
}

func (r *ModelRegistry) RegisterModel(modelName string, model easyllm.Model) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.models[modelName] = model
}

func (r *ModelRegistry) UnregisterModel(modelName string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.models, modelName)
}

func (r *ModelRegistry) HasModel(modelName string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.models[modelName]
	return ok
}
