package agent

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/easyagent-dev/llm"
	"github.com/google/uuid"
)

// Helper functions for timing
func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

func getCurrentNanos() int64 {
	return time.Now().UnixNano()
}

const (
	// DefaultMaxMessageHistory is the default maximum number of messages to keep in history
	DefaultMaxMessageHistory = 100
	// InputSummaryMaxLen is the maximum length for input summary in error messages
	InputSummaryMaxLen = 200
	// InputSummaryEllipsis is the ellipsis string for truncated input summaries
	InputSummaryEllipsis = "..."
)

type CompletionRunner struct {
	agent             *Agent
	model             llm.CompletionModel
	toolRegistry      *ToolRegistry
	maxMessageHistory int
}

func NewCompletionRunner(agent *Agent, model llm.CompletionModel) (*CompletionRunner, error) {
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
	return &CompletionRunner{
		agent:             agent,
		model:             model,
		toolRegistry:      toolRegistry,
		maxMessageHistory: DefaultMaxMessageHistory,
	}, nil
}

// StreamRun executes the agent with streaming support, returning a channel of events
func (r *CompletionRunner) StreamRun(ctx context.Context, req *AgentRequest, options ...llm.CompletionOption) (*AgentStreamResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	eventChan := make(chan AgentEvent, 100)
	streamResp := AgentStreamResponse(eventChan)

	go func() {
		defer close(eventChan)

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

		completed := false
		usage := llm.TokenUsage{}
		totalCost := 0.0

		for i := 0; i < maxIterations && !completed; i++ {
			// Check context cancellation
			select {
			case <-ctx.Done():
				eventChan <- AgentEvent{
					Type:         AgentEventTypeError,
					ErrorMessage: &[]string{ctx.Err().Error()}[0],
				}
				return
			default:
			}

			prompts, err := GetJsonAgentSystemPrompt(r.agent, options, userMessage, r.toolRegistry.GetTools())
			if err != nil {
				errMsg := err.Error()
				eventChan <- AgentEvent{
					Type:         AgentEventTypeError,
					ErrorMessage: &errMsg,
				}
				return
			}

			completionReq := &llm.CompletionRequest{
				Instructions: prompts,
				Messages:     messages,
			}

			// Use StreamComplete for streaming
			stream, err := r.model.StreamComplete(ctx, completionReq)
			if err != nil {
				messages = append(messages, &llm.ModelMessage{
					Role:    llm.RoleUser,
					Content: fmt.Sprintf("ERROR [Iteration %d]: Model streaming failed: %s\n\nPlease try a different approach or tool.", i+1, err.Error()),
				})
				continue
			}

			// Create parser for streaming JSON tool calls
			parser := NewToolCallJsonParser()
			streamClosed := false
			var toolCall *llm.ToolCall

			// Process stream chunks
			for {
				if streamClosed || completed || toolCall != nil {
					break
				}

				select {
				case chunk, ok := <-stream:
					if !ok || chunk == nil {
						streamClosed = true
						break
					}

					chunkType := chunk.Type()
					if chunkType == llm.ReasoningChunkType {
						reasoningChunk := chunk.(llm.StreamReasoningChunk)
						eventChan <- AgentEvent{
							Type:      AgentEventTypeReasoning,
							Reasoning: &reasoningChunk.Reasoning,
						}
					} else if chunkType == llm.TextChunkType {
						textChunk := chunk.(llm.StreamTextChunk)
						content := textChunk.Text

						// Append to parser
						parser.Append(content)

						// Parse events
						currentToolCall, toolCompleted, err := parser.Parse()
						if err != nil {
							errMsg := fmt.Sprintf("failed to parse stream: %v", err)
							eventChan <- AgentEvent{
								Type:         AgentEventTypeError,
								ErrorMessage: &errMsg,
							}
							return
						}

						if currentToolCall != nil {
							if toolCompleted {
								toolCall = currentToolCall
								streamClosed = true
							} else {
								eventChan <- AgentEvent{
									Type:     AgentEventTypeUseTool,
									ToolCall: currentToolCall,
									Partial:  true,
								}
							}
						}
					} else if chunkType == llm.UsageChunkType {
						usageChunk := chunk.(llm.StreamUsageChunk)
						usage.Append(usageChunk.Usage)
						if usageChunk.Cost != nil {
							totalCost += *usageChunk.Cost
						}
					}
				case <-ctx.Done():
					errMsg := ctx.Err().Error()
					eventChan <- AgentEvent{
						Type:         AgentEventTypeError,
						ErrorMessage: &errMsg,
					}
					return
				}
			}

			// If no tool call was parsed, handle the error
			if toolCall == nil {
				messages = append(messages, &llm.ModelMessage{
					Role:    llm.RoleUser,
					Content: fmt.Sprintf("ERROR [Iteration %d]: No valid tool call was generated. You MUST call a tool.\n\nPlease ensure your response contains a valid tool call.", i+1),
				})
				continue
			}

			// Add assistant message with tool call
			messages = append(messages, &llm.ModelMessage{
				Role:     llm.RoleAssistant,
				Content:  "",
				ToolCall: toolCall,
			})

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

			// Track tool execution with timing
			toolCall.StartAt = time.Now()
			toolCallOutput, err := tool.Run(ctx, toolCall.Input)
			toolCall.EndAt = time.Now()

			agentContext.AppendToolCall(toolCall)

			if err != nil {
				inputSummary := fmt.Sprintf("%v", toolCall.Input)
				if len(inputSummary) > InputSummaryMaxLen {
					inputSummary = inputSummary[:InputSummaryMaxLen] + InputSummaryEllipsis
				}
				messages = append(messages, &llm.ModelMessage{
					Role:    llm.RoleUser,
					Content: fmt.Sprintf("ERROR [Iteration %d]: Tool '%s' execution failed.\n\nTool Input: %s\n\nError: %s\n\nPlease review the error and adjust your tool parameters or try a different approach.", i+1, toolCall.Name, inputSummary, err.Error()),
				})
				continue
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
						errMsg := fmt.Sprintf("failed to marshal tool call output: %v", err)
						eventChan <- AgentEvent{
							Type:         AgentEventTypeError,
							ErrorMessage: &errMsg,
						}
						return
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

			// Trim message history to prevent unbounded growth
			if len(messages) > r.maxMessageHistory {
				// Keep initial messages and recent history
				keepInitial := 1 // Keep at least the first user message
				if len(messages)-r.maxMessageHistory+keepInitial > 0 {
					messages = append(messages[:keepInitial], messages[len(messages)-r.maxMessageHistory+keepInitial:]...)
				}
			}
		}

		if !completed {
			errMsg := fmt.Sprintf("agent exceeded max iterations: %d", maxIterations)
			eventChan <- AgentEvent{
				Type:         AgentEventTypeError,
				ErrorMessage: &errMsg,
			}
			return
		}

		_ = results // results would be sent through events if needed
	}()

	return &streamResp, nil
}

// Run executes the agent with the given content
func (r *CompletionRunner) Run(ctx context.Context, req *AgentRequest, options ...llm.CompletionOption) (*AgentResponse, error) {
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

		prompts, err := GetJsonAgentSystemPrompt(r.agent, options, userMessage, r.toolRegistry.GetTools())
		if err != nil {
			return nil, fmt.Errorf("failed to create prompts: %w", err)
		}
		completionReq := &llm.CompletionRequest{
			Instructions: prompts,
			Messages:     messages,
		}

		output, err := r.model.Complete(ctx, completionReq)
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

		toolCall := &llm.ToolCall{}
		err = json.Unmarshal([]byte(output.Output), toolCall)
		if err != nil {
			consecutiveErrors++
			if req.MaxRetries > 0 && consecutiveErrors > req.MaxRetries {
				return nil, fmt.Errorf("exceeded max retries (%d) due to consecutive errors", req.MaxRetries)
			}
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

		// Track tool execution with timing
		toolCall.StartAt = time.Now()
		toolCallOutput, err := tool.Run(ctx, toolCall.Input)
		toolCall.EndAt = time.Now()

		agentContext.AppendToolCall(toolCall)

		if err != nil {
			consecutiveErrors++
			if req.MaxRetries > 0 && consecutiveErrors > req.MaxRetries {
				return nil, fmt.Errorf("exceeded max retries (%d) due to consecutive errors", req.MaxRetries)
			}
			inputSummary := fmt.Sprintf("%v", toolCall.Input)
			if len(inputSummary) > InputSummaryMaxLen {
				inputSummary = inputSummary[:InputSummaryMaxLen] + InputSummaryEllipsis
			}
			messages = append(messages, &llm.ModelMessage{
				Role:    llm.RoleUser,
				Content: fmt.Sprintf("ERROR [Iteration %d]: Tool '%s' execution failed.\n\nTool Input: %s\n\nError: %s\n\nPlease review the error and adjust your tool parameters or try a different approach.", i+1, toolCall.Name, inputSummary, err.Error()),
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
