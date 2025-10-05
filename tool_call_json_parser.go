package agent

import (
	"encoding/json"
	"github.com/easyagent-dev/llm"
	"github.com/easyagent-dev/streamjson"
)

// ToolCallJsonParser parses streaming JSON for ToolCall
type ToolCallJsonParser struct {
	parser *streamjson.StreamJSONParser
	buffer string
}

// NewToolCallJsonParser creates a new JSON parser for ToolCall
func NewToolCallJsonParser() *ToolCallJsonParser {
	return &ToolCallJsonParser{
		parser: streamjson.NewStreamJSONParser(),
	}
}

// Append adds new content to the buffer
func (p *ToolCallJsonParser) Append(content string) {
	p.buffer += content
	p.parser.Append(content)
}

// ParseNext parses the next events from the stream
func (p *ToolCallJsonParser) Parse() (*llm.ToolCall, bool, error) {
	// Check if parsing is completed
	completed := p.parser.IsCompleted()

	if completed {
		var currentToolCall llm.ToolCall
		err := json.Unmarshal([]byte(p.buffer), &currentToolCall)
		if err != nil {
			return nil, false, err
		}
		return &currentToolCall, true, nil
	} else {
		toolName := p.parser.Get("name")
		if toolName != nil {
			input := p.parser.Get("input")
			if input != nil {
				return &llm.ToolCall{
					Name:  toolName.(string),
					Input: input.(map[string]any),
				}, false, nil
			}
		}
	}

	return nil, false, nil
}
