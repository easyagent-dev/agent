package easyagent

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/easymvp/easyllm"
	"github.com/google/uuid"
)

type CompletionRunner struct {
	agent         *Agent
	modelRegistry *ModelRegistry
	toolRegistry  *ToolRegistry
}

var _ Runner = (*CompletionRunner)(nil)

func NewCompletionRunner(agent *Agent, modelRegistry *ModelRegistry, toolRegistry *ToolRegistry) (Runner, error) {
	agentRegistry := NewToolRegistry()
	for _, toolName := range agent.Tools {
		tool, err := toolRegistry.GetTool(toolName)
		if err != nil {
			return nil, fmt.Errorf("failed to register tool: %w", err)
		}
		_ = agentRegistry.RegisterTool(tool)
	}
	return &CompletionRunner{
		agent:         agent,
		modelRegistry: modelRegistry,
		toolRegistry:  agentRegistry,
	}, nil
}

// Run executes the agent with the given content
func (r *CompletionRunner) Run(ctx context.Context, req *AgentRequest, callback Callback) (*AgentResponse, error) {
	var results any = nil
	_ = r.toolRegistry.RegisterTool(NewCompleteTaskTool(req.Config.JSONSchema, ""))

	messages := req.Messages
	maxIterations := req.MaxIterations

	modelInstance, err := r.modelRegistry.GetModel(req.ModelProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to get model provider: %w", err)
	}

	userMessage := messages[len(messages)-1]
	if userMessage.Role != easyllm.MessageRoleUser {
		return nil, errors.New("last message is not user message")
	}

	agentContext := &AgentContext{
		Agent:     r.agent,
		Messages:  messages,
		Callback:  callback,
		ToolCalls: []*easyllm.ToolCall{},
	}
	ctx = WithAgentContext(ctx, agentContext)

	usage := &easyllm.TokenUsage{}
	totalCost := 0.0

	completed := false

	for i := 0; i < maxIterations && !completed; i++ {
		prompts, err := GetJsonAgentSystemPrompt(r.agent.Instructions, req.Config.JSONSchema, userMessage, r.toolRegistry.GetTools())
		if err != nil {
			return nil, fmt.Errorf("failed to create prompts: %w", err)
		}
		modelReq := &easyllm.ModelRequest{
			Instructions: prompts,
			Messages:     messages,
			Config:       req.Config,
		}
		callback.OnModel(ctx, req.ModelProvider, req.Model, prompts, messages)
		output, err := modelInstance.GenerateContent(ctx, modelReq, r.toolRegistry.GetTools())
		if err != nil {
			return nil, fmt.Errorf("failed to call model completion: %w", err)
		}

		agentStep := &AgentStep{}
		err = json.Unmarshal([]byte(output.Output), agentStep)
		if err != nil {
			return nil, fmt.Errorf("failed to convert agent step: %w", err)
		}

		if output.Usage != nil {
			usage.Append(output.Usage)
		}
		if output.Cost != nil {
			totalCost += *output.Cost
		}

		// Handle reasoning
		if agentStep.Reasoning != "" {
			callback.OnReasoning(ctx, agentStep.Reasoning)
		}

		// Handle tool call
		toolName := agentStep.ToolCall.Name
		tool, err := r.toolRegistry.GetTool(toolName)
		toolCallId := uuid.New().String()
		if err != nil {
			return nil, fmt.Errorf("failed to get tool, %s: %w", toolName, err)
		}

		toolInput := ""
		if agentStep.ToolCall.Input != nil {
			toolInput = fmt.Sprintf("%v", agentStep.ToolCall.Input)
		}

		if strings.Contains(toolInput, "google") {
			usage.TotalWebSearches += 1
		}

		messages = append(messages, &easyllm.ModelMessage{
			Role:    easyllm.MessageRoleAssistant,
			Content: "",
			ToolCall: &easyllm.ToolCall{
				ID:    toolCallId,
				Name:  toolName,
				Input: agentStep.ToolCall.Input,
			},
		})
		callback.OnToolCallStart(ctx, toolName, toolInput)
		callback.OnReasoning(ctx, "\n\n[tool] call "+toolName)

		timestamp := time.Now().UnixMilli()
		toolResults, err := tool.Run(ctx, toolInput)
		agentContext.ToolCalls = append(agentContext.ToolCalls, &easyllm.ToolCall{
			ID:     toolCallId,
			Name:   toolName,
			Input:  toolInput,
			Output: &toolResults,
		})
		if err != nil {
			output := "tool call failed: " + err.Error()
			messages = append(messages, &easyllm.ModelMessage{
				Role:    easyllm.MessageRoleTool,
				Content: "",
				ToolCall: &easyllm.ToolCall{
					ID:     toolCallId,
					Name:   toolName,
					Output: &output,
				},
			})
			callback.OnReasoning(ctx, " failed, "+err.Error()+", in "+strconv.FormatInt(time.Now().UnixMilli()-timestamp, 10)+"ms\n\n")
		} else {
			if toolResults == "" {
				output := "tool call successfully with no results"
				messages = append(messages, &easyllm.ModelMessage{
					Role:    easyllm.MessageRoleTool,
					Content: "",
					ToolCall: &easyllm.ToolCall{
						ID:     toolCallId,
						Name:   toolName,
						Output: &output,
					},
				})
				callback.OnReasoning(ctx, " completed, returns no results, in "+strconv.FormatInt(time.Now().UnixMilli()-timestamp, 10)+"ms\n\n")
			} else {
				messages = append(messages, &easyllm.ModelMessage{
					Role:    easyllm.MessageRoleTool,
					Content: "",
					ToolCall: &easyllm.ToolCall{
						ID:     toolCallId,
						Name:   toolName,
						Output: &toolResults,
					},
				})
				callback.OnReasoning(ctx, " completed, in "+strconv.FormatInt(time.Now().UnixMilli()-timestamp, 10)+"ms\n\n")
			}
		}
		callback.OnToolCallEnd(ctx, toolName, toolInput, toolResults)
		if toolName == CompleteTaskToolName {
			completed = true
			results = &toolResults
		}
	}

	if results == nil {
		return nil, errors.New("agent exceeded max iterations")
	}

	runnerOutput := &AgentResponse{
		Output: results,
		Usage:  usage,
		Cost:   &totalCost,
	}
	callback.OnUsage(ctx, req.ModelProvider, req.Model, runnerOutput.Usage)
	return runnerOutput, nil
}
