package easyagent

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/easymvp/easyllm"
)

type Callback interface {
	OnModel(ctx context.Context, provider string, modelName string, prompts string, messages []*easyllm.ModelMessage)
	OnToolCallPartial(ctx context.Context, toolName string, params string)
	OnToolCallStart(ctx context.Context, toolName string, params string)
	OnToolCallEnd(ctx context.Context, toolName string, params string, results interface{})
	OnReasoning(ctx context.Context, reasoning string)
	OnUsage(ctx context.Context, provider string, modelName string, usage *easyllm.TokenUsage)
	OnAttachments(ctx context.Context, attachments []*easyllm.ModelArtifact)
	End(ctx context.Context, content string)
}

type DefaultCallback struct {
	name             string
	TokenUsage       *easyllm.TokenUsage
	TotalToolCalls   map[string]int
	TotalAttachments int
}

var _ Callback = (*DefaultCallback)(nil)

func NewDefaultCallback() *DefaultCallback {
	return &DefaultCallback{
		name:             "DefaultCallback",
		TokenUsage:       &easyllm.TokenUsage{},
		TotalToolCalls:   map[string]int{},
		TotalAttachments: 0,
	}
}

func (c *DefaultCallback) OnModel(ctx context.Context, provider string, modelName string, prompts string, messages []*easyllm.ModelMessage) {
	messagesJSON, _ := json.Marshal(messages)
	fmt.Printf("%s::on model - provider: %s, model: %s, prompts: %s, messages: %s\n",
		c.name, provider, modelName, prompts, string(messagesJSON))
}

func (c *DefaultCallback) OnToolCallStart(ctx context.Context, toolName string, toolInput string) {
	fmt.Printf("%s::tool call start - toolName: %s, toolInput: %s\n", c.name, toolName, toolInput)
	if _, ok := c.TotalToolCalls[toolName]; ok {
		c.TotalToolCalls[toolName]++
	} else {
		c.TotalToolCalls[toolName] = 1
	}
}

func (c *DefaultCallback) OnToolCallPartial(ctx context.Context, toolName string, toolInput string) {
	fmt.Printf("%s::tool call partial - toolName: %s, toolInput: %s\n", c.name, toolName, toolInput)
	if _, ok := c.TotalToolCalls[toolName]; ok {
		c.TotalToolCalls[toolName]++
	} else {
		c.TotalToolCalls[toolName] = 1
	}
}

func (c *DefaultCallback) OnToolCallEnd(ctx context.Context, toolName string, toolInput string, results interface{}) {
	resultsJSON, _ := json.Marshal(results)
	fmt.Printf("%s::tool call end - toolName: %s, toolInput: %s, results: %s\n",
		c.name, toolName, toolInput, string(resultsJSON))
}
func (c *DefaultCallback) OnReasoning(ctx context.Context, reasoning string) {
	fmt.Printf("%s::reasoning - content: %s\n", c.name, reasoning)
}
func (c *DefaultCallback) OnUsage(ctx context.Context, provider string, modelName string, usage *easyllm.TokenUsage) {
	usageJSON, _ := json.Marshal(usage)
	fmt.Printf("%s::usage - provider: %s, model: %s, usage: %s\n",
		c.name, provider, modelName, string(usageJSON))
	c.TokenUsage.Append(usage)
}
func (c *DefaultCallback) OnAttachments(ctx context.Context, attachments []*easyllm.ModelArtifact) {
	attachmentsJSON, _ := json.Marshal(attachments)
	fmt.Printf("%s::attachments - attachments: %s\n", c.name, string(attachmentsJSON))
	c.TotalAttachments += len(attachments)
}
func (c *DefaultCallback) End(ctx context.Context, content string) {
	fmt.Printf("%s::end - content: %s\n", c.name, content)
}
