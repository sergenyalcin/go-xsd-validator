package pkg

import (
	"fmt"
	"strings"
)

// findSchemaElementNS locates an XSD element definition by its name and namespace.
func (v *Validator) findSchemaElementNS(name, namespace string, elements []XSDElement) *XSDElement {
	for i, elem := range elements {
		elemNS := elem.Namespace
		if elemNS == "" {
			elemNS = v.schema.TargetNS
		}

		if elem.Name == name {
			// Match if:
			// 1. Namespaces are exactly equal, or
			// 2. Element is unqualified and we're matching against target namespace
			if elemNS == namespace ||
				(v.schema.ElementFormDefault != "qualified" && namespace == v.schema.TargetNS) {
				return &elements[i]
			}
		}
	}
	return nil
}

// validateElementNameAndNS ensures the XML element matches the expected name and namespace.
func (v *Validator) validateElementNameAndNS(xmlNode *XMLNode, xsdElem XSDElement) bool {
	if xmlNode.Name != xsdElem.Name {
		return false
	}

	// Determine the expected namespace.
	schemaNamespace := xsdElem.Namespace
	if schemaNamespace == "" {
		schemaNamespace = v.schema.TargetNS
	}

	// Handle namespace validation based on schema settings.
	if v.schema.ElementFormDefault == "qualified" {
		if schemaNamespace == "" {
			return xmlNode.Namespace == ""
		}
		return xmlNode.Namespace == schemaNamespace
	}

	// For unqualified elements
	return true
}

func (v *Validator) resolveElementRef(ref string) (*XSDElement, error) {
	// Handle namespace prefix in ref
	parts := strings.Split(ref, ":")
	var localName string
	if len(parts) > 1 {
		localName = parts[1]
	} else {
		localName = parts[0]
	}

	// Search in schema elements
	for _, elem := range v.schema.Elements {
		if elem.Name == localName {
			return &elem, nil
		}
	}
	return nil, fmt.Errorf("referenced element not found: %s", ref)
}
