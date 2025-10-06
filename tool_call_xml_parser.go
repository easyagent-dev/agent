package agent

import (
	"encoding/json"
	"strings"

	"github.com/easyagent-dev/llm"
	"github.com/easyagent-dev/streamjson"
	"github.com/easyagent-dev/streamxml"
)

// ToolCallXMLParser parses streaming XML for ToolCall
type ToolCallXMLParser struct {
	xmlParser  *streamxml.StreamXmlParser
	jsonParser *streamjson.StreamJSONParser
	buffer     string
	reasoning  string
	toolName   string
	jsonBuffer string
	foundTag   bool
}

// NewToolCallXMLParser creates a new XML parser for ToolCall
func NewToolCallXMLParser() *ToolCallXMLParser {
	parser := streamxml.NewStreamXmlParser()
	parser.SetAllowedElements([]string{"use-tool"})
	return &ToolCallXMLParser{
		xmlParser:  parser,
		jsonParser: streamjson.NewStreamJSONParser(),
		foundTag:   false,
	}
}

// Append adds new content to the buffer
func (p *ToolCallXMLParser) Append(content string) {
	p.buffer += content
	_ = p.xmlParser.Append(content)
}

// Parse parses the next events from the stream
// Returns (toolCall, completed, reasoning, error)
func (p *ToolCallXMLParser) Parse() (*llm.ToolCall, bool, *string, error) {
	// Get XML node
	node, err := p.xmlParser.GetXmlNode()
	if err != nil || node == nil {
		return nil, false, nil, err
	}

	// Check if this is the use-tool tag
	if node.Name == "use-tool" {
		if !p.foundTag {
			p.foundTag = true

			// Extract tool name from attribute
			if name, ok := node.Attributes["name"]; ok {
				p.toolName = name
			}

			// Extract reasoning from text before the tag
			text, _ := p.xmlParser.GetText()
			if text != "" {
				p.reasoning = strings.TrimSpace(text)
			}
		}

		// Get the JSON content
		jsonContent := strings.TrimSpace(node.Content)

		// If content changed, append to JSON parser
		if jsonContent != p.jsonBuffer {
			newContent := jsonContent[len(p.jsonBuffer):]
			p.jsonBuffer = jsonContent
			p.jsonParser.Append(newContent)
		}

		// Check if the tag is complete (not partial)
		if !node.Partial {
			// Parse complete JSON
			var input map[string]any
			err := json.Unmarshal([]byte(jsonContent), &input)
			if err != nil {
				return nil, false, nil, err
			}

			toolCall := &llm.ToolCall{
				Name:  p.toolName,
				Input: input,
			}

			var reasoningPtr *string
			if p.reasoning != "" {
				reasoningPtr = &p.reasoning
			}

			return toolCall, true, reasoningPtr, nil
		}

		// Return partial tool call if we have enough data
		if p.toolName != "" && p.jsonParser.Get("") != nil {
			input := p.jsonParser.Get("")
			if inputMap, ok := input.(map[string]any); ok {
				toolCall := &llm.ToolCall{
					Name:  p.toolName,
					Input: inputMap,
				}

				var reasoningPtr *string
				if p.reasoning != "" {
					reasoningPtr = &p.reasoning
				}

				return toolCall, false, reasoningPtr, nil
			}
		}
	}

	return nil, false, nil, nil
}
