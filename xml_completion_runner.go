package agent

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/easyagent-dev/llm"
	"github.com/google/uuid"
)

//go:embed prompts/xml_system.md
var xmlSystemPrompt string

type XMLCompletionRunner struct {
	BaseRunner
	agent        *Agent
	model        llm.CompletionModel
	toolRegistry *ToolRegistry
}

var _ Runner = (*XMLCompletionRunner)(nil)

func NewXMLCompletionRunner(agent *Agent, model llm.CompletionModel, opts ...RunnerOption) (Runner, error) {
	// Validate agent configuration
	if err := agent.Validate(); err != nil {
		return nil, fmt.Errorf("invalid agent: %w", err)
	}

	toolRegistry := NewToolRegistry()
	for _, tool := range agent.Tools {
		if err := toolRegistry.RegisterTool(tool); err != nil {
			return nil, fmt.Errorf("failed to register tool %s: %w", tool.Name(), err)
		}
	}

	config := newRunnerConfig(opts...)

	// Use XML system prompt if no custom prompt is set
	systemPrompt := xmlSystemPrompt
	if config.systemPrompts != "" {
		systemPrompt = config.systemPrompts
	}

	return &XMLCompletionRunner{
		BaseRunner: BaseRunner{
			systemPrompts:     systemPrompt,
			maxMessageHistory: config.maxMessageHistory,
		},
		agent:        agent,
		model:        model,
		toolRegistry: toolRegistry,
	}, nil
}

// parseXMLToolCall parses a tool call from XML format
func parseXMLToolCall(output string) (*llm.ToolCall, error) {
	// Pattern to match: <use-tool name="tool_name">{"param":"value"}</use-tool>
	// Parse the JSON input using the XML parser which internally uses JSON parser
	parser := NewToolCallXMLParser()
	parser.Append(output)
	toolCall, completed, _, err := parser.Parse()

	if err != nil {
		return nil, fmt.Errorf("failed to parse tool call: %w", err)
	}

	if !completed || toolCall == nil {
		return nil, fmt.Errorf("incomplete tool call in output")
	}

	return toolCall, nil
}

// Run executes the agent with the given content
func (r *XMLCompletionRunner) Run(ctx context.Context, req *AgentRequest, callback Callback) (*AgentResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var results any = nil
	_ = r.toolRegistry.RegisterTool(NewCompleteTaskTool(req.OutputSchema, req.OutputUsage))

	messages := req.Messages
	maxIterations := req.MaxIterations

	userMessage := messages[len(messages)-1]
	agentContext := &AgentContext{
		Agent:    r.agent,
		Messages: messages,
	}
	ctx = WithAgentContext(ctx, agentContext)

	usage := &llm.TokenUsage{}
	totalCost := 0.0

	completed := false
	consecutiveErrors := 0
	for i := 0; i < maxIterations && !completed; i++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
		default:
		}

		prompts, err := r.GetSystemPrompt(r.agent, userMessage, r.toolRegistry.GetTools())
		if err != nil {
			return nil, fmt.Errorf("failed to create prompts: %w", err)
		}
		completionReq := &llm.CompletionRequest{
			Instructions: prompts,
			Messages:     messages,
		}

		// Call BeforeModel callback
		if callback != nil {
			if err := callback.BeforeModel(ctx, r.agent.ModelProvider, r.agent.Model, prompts, messages); err != nil {
				return nil, fmt.Errorf("callback BeforeModel failed: %w", err)
			}
		}

		output, err := r.model.Complete(ctx, completionReq)

		// Call AfterModel callback
		if callback != nil && err == nil {
			if cbErr := callback.AfterModel(ctx, r.agent.ModelProvider, r.agent.Model, prompts, messages, output.Output, output.Usage); cbErr != nil {
				return nil, fmt.Errorf("callback AfterModel failed: %w", cbErr)
			}
		}

		if err != nil {
			consecutiveErrors++
			if req.MaxRetries > 0 && consecutiveErrors > req.MaxRetries {
				return nil, fmt.Errorf("exceeded max retries (%d) due to consecutive errors", req.MaxRetries)
			}
			messages = append(messages, &llm.ModelMessage{
				Role:    llm.RoleUser,
				Content: fmt.Sprintf("ERROR [Iteration %d]: Model completion failed: %s\n\nPlease try a different approach or tool.", i+1, err.Error()),
			})
			continue
		}

		toolCall, err := parseXMLToolCall(output.Output)
		if err != nil {
			consecutiveErrors++
			if req.MaxRetries > 0 && consecutiveErrors > req.MaxRetries {
				return nil, fmt.Errorf("exceeded max retries (%d) due to consecutive errors", req.MaxRetries)
			}
			messages = append(messages, &llm.ModelMessage{
				Role:    llm.RoleUser,
				Content: fmt.Sprintf("ERROR [Iteration %d]: Failed to parse tool call from your response.\n\nInvalid XML: %s\n\nError: %s\n\nPlease ensure your response contains a valid <use-tool> tag with proper JSON input.", i+1, output.Output, err.Error()),
			})
			continue
		}

		toolCall.ID = uuid.New().String()
		messages = append(messages, &llm.ModelMessage{
			Role:     llm.RoleAssistant,
			Content:  "",
			ToolCall: toolCall,
		})

		if output.Usage != nil {
			usage.Append(output.Usage)
		}

		if output.Cost != nil {
			totalCost += *output.Cost
		}

		// Handle tool call
		tool, err := r.toolRegistry.GetTool(toolCall.Name)
		if err != nil {
			availableTools := []string{}
			for _, t := range r.toolRegistry.GetTools() {
				availableTools = append(availableTools, t.Name())
			}
			messages = append(messages, &llm.ModelMessage{
				Role:    llm.RoleUser,
				Content: fmt.Sprintf("ERROR [Iteration %d]: Tool '%s' not found.\n\nAvailable tools: %v\n\nPlease use one of the available tools.", i+1, toolCall.Name, availableTools),
			})
			continue
		}

		// Call BeforeToolCall callback
		if callback != nil {
			if cbErr := callback.BeforeToolCall(ctx, toolCall.Name, toolCall.Input); cbErr != nil {
				return nil, fmt.Errorf("callback BeforeToolCall failed: %w", cbErr)
			}
		}

		// Track tool execution with timing
		toolCall.StartAt = time.Now()
		toolCallOutput, err := tool.Run(ctx, toolCall.Input)
		toolCall.EndAt = time.Now()

		// Call AfterToolCall callback
		if callback != nil && err == nil {
			if cbErr := callback.AfterToolCall(ctx, toolCall.Name, toolCall.Input, toolCallOutput); cbErr != nil {
				return nil, fmt.Errorf("callback AfterToolCall failed: %w", cbErr)
			}
		}

		agentContext.AppendToolCall(toolCall)

		if err != nil {
			consecutiveErrors++
			if req.MaxRetries > 0 && consecutiveErrors > req.MaxRetries {
				return nil, fmt.Errorf("exceeded max retries (%d) due to consecutive errors", req.MaxRetries)
			}
			messages = append(messages, &llm.ModelMessage{
				Role:    llm.RoleUser,
				Content: fmt.Sprintf("ERROR [Iteration %d]: %s", i+1, err.Error()),
			})
			continue
		}

		consecutiveErrors = 0

		if tool.Name() == CompleteTaskToolName {
			completed = true
			results = toolCallOutput
		} else {
			if toolCallOutput == nil {
				messages = append(messages, &llm.ModelMessage{
					Role:    llm.RoleTool,
					Content: "Tool call success, no results",
				})
			} else {
				// For XML format, we need to serialize the output
				content := fmt.Sprintf("%v", toolCallOutput)
				messages = append(messages, &llm.ModelMessage{
					Role: llm.RoleTool,
					ToolCall: &llm.ToolCall{
						ID:     toolCall.ID,
						Name:   toolCall.Name,
						Input:  toolCall.Input,
						Output: content,
					},
				})
			}
		}

		// Trim message history to prevent unbounded growth
		if len(messages) > r.maxMessageHistory {
			// Keep initial messages and recent history
			keepInitial := 1 // Keep at least the first user message
			if len(messages)-r.maxMessageHistory+keepInitial > 0 {
				messages = append(messages[:keepInitial], messages[len(messages)-r.maxMessageHistory+keepInitial:]...)
			}
		}
	}

	resp := &AgentResponse{
		Output: results,
		Usage:  usage,
		Cost:   &totalCost,
	}
	return resp, nil
}
