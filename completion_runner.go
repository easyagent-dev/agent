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
				Content: err.Error(),
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
				Content: err.Error(),
			})
			continue
		}

		toolCall := &llm.ToolCall{}
		err = json.Unmarshal([]byte(output.Output), toolCall)
		if err != nil {
			messages = append(messages, &llm.ModelMessage{
				Role:    llm.RoleUser,
				Content: err.Error(),
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
			messages = append(messages, &llm.ModelMessage{
				Role:    llm.RoleUser,
				Content: err.Error(),
			})
			continue
		}

		if tool.Name() != CompleteTaskToolName {
			beforeToolCallOutput, err := callback.BeforeToolCall(ctx, toolCall.Name, toolCall.Input)
			if err != nil {
				messages = append(messages, &llm.ModelMessage{
					Role:    llm.RoleUser,
					Content: err.Error(),
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
			messages = append(messages, &llm.ModelMessage{
				Role:    llm.RoleUser,
				Content: err.Error(),
			})
			continue
		}

		if tool.Name() != CompleteTaskToolName {
			afterToolCallOutput, err := callback.AfterToolCall(ctx, toolCall.Name, toolCall.Input, toolCallOutput)
			if err != nil {
				messages = append(messages, &llm.ModelMessage{
					Role:    llm.RoleUser,
					Content: err.Error(),
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
