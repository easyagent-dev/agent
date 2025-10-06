package agent

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/easyagent-dev/llm"
)

type XMLCompletionStreamRunner struct {
	BaseRunner
	agent        *Agent
	model        llm.CompletionModel
	toolRegistry *ToolRegistry
}

var _ StreamRunner = (*XMLCompletionStreamRunner)(nil)

func NewXMLCompletionStreamRunner(agent *Agent, model llm.CompletionModel, opts ...RunnerOption) (StreamRunner, error) {
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

	return &XMLCompletionStreamRunner{
		BaseRunner: BaseRunner{
			systemPrompts:     systemPrompt,
			maxMessageHistory: config.maxMessageHistory,
		},
		agent:        agent,
		model:        model,
		toolRegistry: toolRegistry,
	}, nil
}

// Run executes the agent with streaming support, returning a channel of events
func (r *XMLCompletionStreamRunner) Run(ctx context.Context, req *AgentRequest, callback Callback) (*AgentStreamResponse, error) {
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

			prompts, err := r.GetSystemPrompt(r.agent, userMessage, r.toolRegistry.GetTools())
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

			// Call BeforeModel callback
			if callback != nil {
				if err := callback.BeforeModel(ctx, r.agent.ModelProvider, r.agent.Model, prompts, messages); err != nil {
					errMsg := fmt.Sprintf("callback BeforeModel failed: %v", err)
					eventChan <- AgentEvent{
						Type:         AgentEventTypeError,
						ErrorMessage: &errMsg,
					}
					return
				}
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

			// Create parser for streaming XML tool calls
			parser := NewToolCallXMLParser()
			streamClosed := false
			var toolCall *llm.ToolCall
			var fullOutput string
			reasoningSent := false

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

						// Accumulate full output for AfterModel callback
						fullOutput += content

						// Append to parser
						parser.Append(content)

						// Parse events
						currentToolCall, toolCompleted, reasoning, err := parser.Parse()
						if err != nil {
							errMsg := fmt.Sprintf("failed to parse stream, content:%s, %v", content, err)
							eventChan <- AgentEvent{
								Type:         AgentEventTypeError,
								ErrorMessage: &errMsg,
							}
							return
						}

						// Send reasoning event if available and not sent yet
						if reasoning != nil && !reasoningSent {
							eventChan <- AgentEvent{
								Type:      AgentEventTypeReasoning,
								Reasoning: reasoning,
							}
							reasoningSent = true
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

			// Call AfterModel callback
			if callback != nil && toolCall != nil {
				if cbErr := callback.AfterModel(ctx, r.agent.ModelProvider, r.agent.Model, prompts, messages, fullOutput, &usage); cbErr != nil {
					errMsg := fmt.Sprintf("callback AfterModel failed: %v", cbErr)
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
					Content: fmt.Sprintf("ERROR [Iteration %d]: No valid tool call was generated. You MUST call a tool.\n\nPlease ensure your response contains a valid <use-tool> tag.", i+1),
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

			// Call BeforeToolCall callback
			if callback != nil {
				if cbErr := callback.BeforeToolCall(ctx, toolCall.Name, toolCall.Input); cbErr != nil {
					errMsg := fmt.Sprintf("callback BeforeToolCall failed: %v", cbErr)
					eventChan <- AgentEvent{
						Type:         AgentEventTypeError,
						ErrorMessage: &errMsg,
					}
					return
				}
			}

			// Track tool execution with timing
			toolCall.StartAt = time.Now()
			toolCallOutput, err := tool.Run(ctx, toolCall.Input)
			toolCall.EndAt = time.Now()

			// Call AfterToolCall callback
			if callback != nil && err == nil {
				if cbErr := callback.AfterToolCall(ctx, toolCall.Name, toolCall.Input, toolCallOutput); cbErr != nil {
					errMsg := fmt.Sprintf("callback AfterToolCall failed: %v", cbErr)
					eventChan <- AgentEvent{
						Type:         AgentEventTypeError,
						ErrorMessage: &errMsg,
					}
					return
				}
			}

			agentContext.AppendToolCall(toolCall)

			if err != nil {
				messages = append(messages, &llm.ModelMessage{
					Role:    llm.RoleUser,
					Content: fmt.Sprintf("ERROR [Iteration %d]: %s", i+1, err.Error()),
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
