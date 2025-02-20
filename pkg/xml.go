package pkg

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// XML node representation
type XMLNode struct {
	Name           string
	Namespace      string
	Prefix         string
	Attributes     map[string]string
	Content        string
	Children       []*XMLNode
	NamespaceDecls map[string]string
}

// ParseXML parses XML document and returns XMLNode
func ParseXML(r io.Reader) (*XMLNode, error) {
	decoder := xml.NewDecoder(r)
	var stack []*XMLNode
	var root *XMLNode

	// Track namespaces at each level
	nsStack := []map[string]string{{}}

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			// Create new namespace context for this element
			currentNS := make(map[string]string)
			for prefix, uri := range nsStack[len(nsStack)-1] {
				currentNS[prefix] = uri
			}

			// Process namespace declarations
			for _, attr := range t.Attr {
				if attr.Name.Space == "xmlns" {
					currentNS[attr.Name.Local] = attr.Value
				} else if attr.Name.Local == "xmlns" {
					currentNS[""] = attr.Value
				}
			}
			nsStack = append(nsStack, currentNS)

			// Resolve element namespace
			namespace := t.Name.Space
			if namespace == "" {
				// Check for default namespace
				if defaultNS, ok := currentNS[""]; ok {
					namespace = defaultNS
				}
			} else {
				// Resolve prefixed namespace
				if uri, ok := currentNS[namespace]; ok {
					namespace = uri
				}
			}

			node := &XMLNode{
				Name:           t.Name.Local,
				Namespace:      namespace,
				Prefix:         t.Name.Space,
				Attributes:     make(map[string]string),
				NamespaceDecls: make(map[string]string),
			}

			// Process attributes
			for _, attr := range t.Attr {
				if attr.Name.Space == "xmlns" || attr.Name.Local == "xmlns" {
					node.NamespaceDecls[attr.Name.Local] = attr.Value
				} else {
					attrNS := attr.Name.Space
					if attrNS != "" {
						if uri, ok := currentNS[attrNS]; ok {
							attrNS = uri
						}
						node.Attributes[fmt.Sprintf("{%s}%s", attrNS, attr.Name.Local)] = attr.Value
					} else {
						node.Attributes[attr.Name.Local] = attr.Value
					}
				}
			}

			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, node)
			} else {
				root = node
			}
			stack = append(stack, node)

		case xml.EndElement:
			stack = stack[:len(stack)-1]
			nsStack = nsStack[:len(nsStack)-1]

		case xml.CharData:
			if len(stack) > 0 {
				current := stack[len(stack)-1]
				current.Content += strings.TrimSpace(string(t))
			}
		}
	}

	return root, nil
}
