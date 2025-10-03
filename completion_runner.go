package agent

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/easymvp-ai/llm"
	"github.com/google/uuid"
)

type CompletionRunner struct {
	agent        *Agent
	model        llm.CompletionModel
	toolRegistry *ToolRegistry
}

var _ Runner = (*CompletionRunner)(nil)

func NewCompletionRunner(agent *Agent, model llm.CompletionModel) (Runner, error) {
	toolRegistry := NewToolRegistry()
	for _, tool := range agent.Tools {
		_ = toolRegistry.RegisterTool(tool)
	}
	return &CompletionRunner{
		agent:        agent,
		model:        model,
		toolRegistry: toolRegistry,
	}, nil
}

// Run executes the agent with the given content
func (r *CompletionRunner) Run(ctx context.Context, req *AgentRequest) (*AgentResponse, error) {
	var results any = nil
	_ = r.toolRegistry.RegisterTool(NewCompleteTaskTool(req.OutputSchema, req.OutputUsage))

	messages := req.Messages
	maxIterations := req.MaxIterations

	userMessage := messages[len(messages)-1]
	if userMessage.Role != llm.RoleUser {
		return nil, errors.New("last message is not user message")
	}
	agentContext := &AgentContext{
		Agent:       r.agent,
		Messages:    messages,
		ToolsCalled: []*llm.ToolCall{},
	}
	ctx = WithAgentContext(ctx, agentContext)

	usage := &llm.TokenUsage{}
	totalCost := 0.0

	completed := false
	callback := r.agent.Callback
	for i := 0; i < maxIterations && !completed; i++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
		default:
		}

		prompts, err := GetJsonAgentSystemPrompt(r.agent.Instructions, req.Options, userMessage, r.toolRegistry.GetTools())
		if err != nil {
			return nil, fmt.Errorf("failed to create prompts: %w", err)
		}
		completionReq := &llm.CompletionRequest{
			Model:        req.Model,
			Instructions: prompts,
			Messages:     messages,
		}
		completionResp, err := callback.BeforeModel(ctx, r.model.Name(), req.Model, completionReq)
		if err != nil {
			messages = append(messages, &llm.ModelMessage{
				Role:    llm.RoleUser,
				Content: fmt.Sprintf("ERROR [Iteration %d]: Failed to execute BeforeModel callback: %s\n\nPlease adjust your approach and try again.", i+1, err.Error()),
			})
			continue
		}

		if completionResp != nil {
			messages = append(messages, &llm.ModelMessage{
				Role:    llm.RoleAssistant,
				Content: completionResp.Output,
			})
			continue
		}

		output, err := r.model.Complete(ctx, completionReq)
		if err != nil {
			messages = append(messages, &llm.ModelMessage{
				Role:    llm.RoleUser,
				Content: fmt.Sprintf("ERROR [Iteration %d]: Model completion failed: %s\n\nPlease try a different approach or tool.", i+1, err.Error()),
			})
			continue
		}

		toolCall := &llm.ToolCall{}
		err = json.Unmarshal([]byte(output.Output), toolCall)
		if err != nil {
			messages = append(messages, &llm.ModelMessage{
				Role:    llm.RoleUser,
				Content: fmt.Sprintf("ERROR [Iteration %d]: Failed to parse tool call from your response.\n\nInvalid JSON: %s\n\nError: %s\n\nPlease ensure your response is valid JSON matching the tool call schema.", i+1, output.Output, err.Error()),
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

		if tool.Name() != CompleteTaskToolName {
			beforeToolCallOutput, err := callback.BeforeToolCall(ctx, toolCall.Name, toolCall.Input)
			if err != nil {
				messages = append(messages, &llm.ModelMessage{
					Role:    llm.RoleUser,
					Content: fmt.Sprintf("ERROR [Iteration %d]: BeforeToolCall callback failed for tool '%s'.\n\nError: %s\n\nPlease try a different tool or approach.", i+1, toolCall.Name, err.Error()),
				})
				continue
			}

			if beforeToolCallOutput != nil {
				content, err := json.Marshal(beforeToolCallOutput)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal tool call output: %w", err)
				}
				messages = append(messages, &llm.ModelMessage{
					Role:    llm.RoleTool,
					Content: string(content),
				})
				continue
			}
		}

		toolCallOutput, err := tool.Run(ctx, toolCall.Input)
		if err != nil {
			inputSummary := fmt.Sprintf("%v", toolCall.Input)
			if len(inputSummary) > 200 {
				inputSummary = inputSummary[:200] + "..."
			}
			messages = append(messages, &llm.ModelMessage{
				Role:    llm.RoleUser,
				Content: fmt.Sprintf("ERROR [Iteration %d]: Tool '%s' execution failed.\n\nTool Input: %s\n\nError: %s\n\nPlease review the error and adjust your tool parameters or try a different approach.", i+1, toolCall.Name, inputSummary, err.Error()),
			})
			continue
		}

		if tool.Name() != CompleteTaskToolName {
			afterToolCallOutput, err := callback.AfterToolCall(ctx, toolCall.Name, toolCall.Input, toolCallOutput)
			if err != nil {
				messages = append(messages, &llm.ModelMessage{
					Role:    llm.RoleUser,
					Content: fmt.Sprintf("ERROR [Iteration %d]: AfterToolCall callback failed for tool '%s'.\n\nError: %s\n\nThe tool executed successfully, but post-processing failed. Please proceed with the next step.", i+1, toolCall.Name, err.Error()),
				})
				continue
			}

			if afterToolCallOutput != nil {
				content, err := json.Marshal(afterToolCallOutput)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal tool call output: %w", err)
				}
				messages = append(messages, &llm.ModelMessage{
					Role:    llm.RoleTool,
					Content: string(content),
				})
				continue
			}
		}

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
				content, err := json.Marshal(toolCallOutput)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal tool call output: %w", err)
				}
				messages = append(messages, &llm.ModelMessage{
					Role: llm.RoleTool,
					ToolCall: &llm.ToolCall{
						ID:     toolCall.ID,
						Name:   toolCall.Name,
						Input:  toolCall.Input,
						Output: string(content),
					},
				})
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
