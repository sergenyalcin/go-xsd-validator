package pkg

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// XML node representation
type XMLNode struct {
	Name       string
	Attributes map[string]string
	Content    string
	Children   []*XMLNode
}

// ParseXML parses XML document and returns XMLNode
func ParseXML(r io.Reader) (*XMLNode, error) {
	decoder := xml.NewDecoder(r)
	var root *XMLNode
	var stack []*XMLNode

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading XML: %v", err)
		}
		switch t := token.(type) {
		case xml.StartElement:
			node := &XMLNode{
				Name:       t.Name.Local,
				Attributes: make(map[string]string),
			}
			for _, attr := range t.Attr {
				node.Attributes[attr.Name.Local] = attr.Value
			}
			if root == nil {
				root = node
			} else {
				// Append the new node as a child of the last node in the stack.
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, node)
			}
			stack = append(stack, node)
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			content := strings.TrimSpace(string(t))
			if content != "" && len(stack) > 0 {
				current := stack[len(stack)-1]
				current.Content += content
			}
		}
	}

	if root == nil {
		return nil, fmt.Errorf("empty or invalid XML document")
	}
	return root, nil
}
