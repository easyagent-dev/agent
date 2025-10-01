package easyagent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
)

// AST node
type Node struct {
	Key        string
	Value      interface{}
	ChildMap   map[string]*Node
	ChildArray []*Node
}

// JSON parser struct
type StreamJsonParser struct {
	root     *Node
	stack    []*Node
	dec      *json.Decoder
	buffer   *bytes.Buffer
	mu       sync.Mutex
	keyStack []string // track nested keys
}

// NewStreamJsonParser creates a parser with streaming capability
func NewStreamJsonParser() *StreamJsonParser {
	buf := &bytes.Buffer{}
	return &StreamJsonParser{
		root:   &Node{Key: "root", ChildMap: make(map[string]*Node)},
		buffer: buf,
		dec:    json.NewDecoder(buf),
		stack:  []*Node{},
	}
}

// Append feeds new JSON bytes into the parser and updates AST
func (p *StreamJsonParser) Append(data []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, err := p.buffer.Write(data)
	if err != nil {
		return err
	}

	var currentKey string

	for {
		tok, err := p.dec.Token()
		if err == io.EOF {
			break // wait for more data
		}
		if err != nil {
			return err
		}

		switch v := tok.(type) {
		case json.Delim:
			switch v {
			case '{':
				node := &Node{Key: currentKey, ChildMap: make(map[string]*Node)}
				if len(p.stack) == 0 {
					// Root level object
					if currentKey == "" {
						// This is the root object itself
						p.stack = append(p.stack, p.root)
					} else {
						p.root.ChildMap[currentKey] = node
						p.stack = append(p.stack, node)
					}
				} else {
					parent := p.stack[len(p.stack)-1]
					if currentKey != "" {
						if parent.ChildMap == nil {
							parent.ChildMap = make(map[string]*Node)
						}
						parent.ChildMap[currentKey] = node
					} else {
						parent.ChildArray = append(parent.ChildArray, node)
					}
					p.stack = append(p.stack, node)
				}
				currentKey = ""

			case '[':
				node := &Node{Key: currentKey, ChildArray: []*Node{}}
				if len(p.stack) == 0 {
					p.root.ChildMap[currentKey] = node
				} else {
					parent := p.stack[len(p.stack)-1]
					if currentKey != "" {
						if parent.ChildMap == nil {
							parent.ChildMap = make(map[string]*Node)
						}
						parent.ChildMap[currentKey] = node
					} else {
						parent.ChildArray = append(parent.ChildArray, node)
					}
				}
				p.stack = append(p.stack, node)
				currentKey = ""

			case '}', ']':
				if len(p.stack) > 0 {
					p.stack = p.stack[:len(p.stack)-1]
				}
			}

		case string:
			if currentKey == "" {
				currentKey = v
			} else {
				node := &Node{Key: currentKey, Value: v}
				parent := p.currentParent()
				if parent.ChildMap == nil {
					parent.ChildMap = make(map[string]*Node)
				}
				parent.ChildMap[currentKey] = node
				currentKey = ""
			}

		case float64, bool, nil:
			parent := p.currentParent()
			if currentKey != "" {
				node := &Node{Key: currentKey, Value: v}
				if parent.ChildMap == nil {
					parent.ChildMap = make(map[string]*Node)
				}
				parent.ChildMap[currentKey] = node
			} else {
				node := &Node{Value: v}
				parent.ChildArray = append(parent.ChildArray, node)
			}
			currentKey = ""
		}
	}

	return nil
}

func (p *StreamJsonParser) currentParent() *Node {
	if len(p.stack) == 0 {
		return p.root
	}
	return p.stack[len(p.stack)-1]
}

// Get retrieves a value by key path
func (p *StreamJsonParser) Get(keys ...string) interface{} {
	node := p.root
	for _, k := range keys {
		if node.ChildMap == nil {
			return nil
		}
		next, ok := node.ChildMap[k]
		if !ok {
			return nil
		}
		node = next
	}
	return node.Value
}

// ToString converts the AST back to JSON string
func (p *StreamJsonParser) ToString() string {
	return p.nodeToJSON(p.root)
}

// nodeToJSON recursively converts a node to JSON string
func (p *StreamJsonParser) nodeToJSON(node *Node) string {
	if node.Value != nil {
		// Leaf node with a value
		return p.valueToJSON(node.Value)
	}

	if len(node.ChildArray) > 0 {
		// Array node
		var elements []string
		for _, child := range node.ChildArray {
			elements = append(elements, p.nodeToJSON(child))
		}
		return "[" + strings.Join(elements, ",") + "]"
	}

	if len(node.ChildMap) > 0 {
		// Object node
		var pairs []string
		for key, child := range node.ChildMap {
			jsonValue := p.nodeToJSON(child)
			pairs = append(pairs, fmt.Sprintf(`"%s":%s`, key, jsonValue))
		}
		return "{" + strings.Join(pairs, ",") + "}"
	}

	// Empty node
	return "null"
}

// valueToJSON converts a Go value to JSON string representation
func (p *StreamJsonParser) valueToJSON(value interface{}) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf(`"%s"`, v)
	case float64:
		// Check if it's an integer
		if v == float64(int64(v)) {
			return fmt.Sprintf("%.0f", v)
		}
		return fmt.Sprintf("%g", v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case nil:
		return "null"
	default:
		return fmt.Sprintf(`"%v"`, v)
	}
}
