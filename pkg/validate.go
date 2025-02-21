package pkg

import (
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"strconv"
)

// Validator is responsible for validating XML files against an XSD schema.
// It holds the schema, precompiled regex patterns, and namespace mappings.
type Validator struct {
	schema     *XSDSchema
	patterns   *PatternCache
	namespaces map[string]string
	defaultNS  string
}

// NewValidator initializes a Validator instance by parsing an XSD file.
// It returns an error if the XSD cannot be parsed.
func NewValidator(xsdFile io.Reader) (*Validator, error) {
	schema := &XSDSchema{}
	decoder := xml.NewDecoder(xsdFile)
	if err := decoder.Decode(schema); err != nil {
		return nil, fmt.Errorf("failed to parse XSD: %v", err)
	}

	return &Validator{
		schema:     schema,
		patterns:   NewPatternCache(),
		namespaces: make(map[string]string),
		defaultNS:  schema.TargetNS,
	}, nil
}

// Validate checks an XML file against the XSD schema and returns a ValidationResult.
// If the XML file does not conform to the schema, errors are collected.
func (v *Validator) Validate(xmlFile io.Reader) (*ValidationResult, error) {
	xmlNode, err := ParseXML(xmlFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML: %v", err)
	}

	// Find the corresponding schema definition for the root element.
	rootXsd := v.findSchemaElementNS(xmlNode.Name, xmlNode.Namespace, v.schema.Elements)
	if rootXsd == nil {
		return nil, fmt.Errorf("root element '{%s}%s' not defined in schema", xmlNode.Namespace, xmlNode.Name)
	}

	// Validate the XML file recursively.
	result := &ValidationResult{
		Valid:    true,
		Filename: xmlNode.Name,
		Errors:   v.validateElement(xmlNode, *rootXsd),
	}

	// If any validation errors are found, mark the XML as invalid.
	if len(result.Errors) > 0 {
		result.Valid = false
	}

	return result, nil
}

// validateElement performs recursive validation of an XML element against the
// schema definition.
func (v *Validator) validateElement(xmlNode *XMLNode, xsdElem XSDElement) []string {
	var errors []string

	// If the element references another definition, resolve it first.
	if xsdElem.Ref != "" {
		refElement, err := v.resolveElementRef(xsdElem.Ref)
		if err != nil {
			return append(errors, err.Error())
		}
		return v.validateElement(xmlNode, *refElement)
	}

	// Validate the element name and namespace.
	if !v.validateElementNameAndNS(xmlNode, xsdElem) {
		errors = append(errors, fmt.Sprintf("element name or namespace mismatch: expected '{%s}%s', got '{%s}%s'",
			xsdElem.Namespace, xsdElem.Name, xmlNode.Namespace, xmlNode.Name))
		return errors
	}

	// If the element has a referenced complex type, retrieve it.
	if xsdElem.ComplexType == nil && xsdElem.Type != "" {
		if ct := v.findComplexType(xsdElem.Type); ct != nil {
			xsdElem.ComplexType = ct
		}
	}

	// Validate attributes of the element.
	if xsdElem.ComplexType != nil {
		errors = append(errors, v.validateAttributes(xmlNode.Attributes, xsdElem.ComplexType.Attributes)...)
	}

	// Validate text content inside the element.
	if xmlNode.Content != "" {
		if err := v.validateElementContent(xmlNode.Content, &xsdElem); err != nil {
			errors = append(errors, fmt.Sprintf("invalid content in element '%s': %v", xmlNode.Name, err))
		}
	}

	// Validate child elements based on complex type constraints.
	if xsdElem.ComplexType != nil {
		if xsdElem.ComplexType.Sequence != nil {
			errors = append(errors, v.validateSequence(xmlNode.Children, xsdElem.ComplexType.Sequence)...)
		}
		if xsdElem.ComplexType.Choice != nil {
			errors = append(errors, v.validateChoice(xmlNode.Children, xsdElem.ComplexType.Choice)...)
		}
	}

	return errors
}

// validateChoice checks whether the child elements satisfy an XSD <choice> constraint.
func (v *Validator) validateChoice(children []*XMLNode, choice *XSDChoice) []string {
	var errors []string

	// Parse MinOccurs and MaxOccurs values for constraints.
	minOccurs := 1
	if choice.MinOccurs != "" {
		if val, err := strconv.Atoi(choice.MinOccurs); err == nil {
			minOccurs = val
		}
	}
	maxOccurs := 1
	if choice.MaxOccurs != "" {
		if choice.MaxOccurs == "unbounded" {
			maxOccurs = math.MaxInt32
		} else if val, err := strconv.Atoi(choice.MaxOccurs); err == nil {
			maxOccurs = val
		}
	}

	// Recursively validate nested choice elements.
	if choice.Choice != nil {
		errors = append(errors, v.validateChoice(children, choice.Choice)...)
	}

	// Check which elements from the choice group are present.
	validChoices := 0
	if choice.Elements != nil {
		for _, child := range children {
			found := false
			for _, choiceElem := range choice.Elements {
				var elemToValidate XSDElement
				if choiceElem.Ref != "" {
					refElement, err := v.resolveElementRef(choiceElem.Ref)
					if err != nil {
						errors = append(errors, err.Error())
						continue
					}
					elemToValidate = *refElement
				} else {
					elemToValidate = choiceElem
				}

				if v.validateElementNameAndNS(child, elemToValidate) {
					found = true
					validChoices++
					errors = append(errors, v.validateElement(child, elemToValidate)...)
					break
				}
			}
			if !found {
				errors = append(errors, fmt.Sprintf("element '%s' is not a valid choice", child.Name))
			}
		}
	}

	// Validate occurrence constraints for the choice group.
	if validChoices < minOccurs {
		errors = append(errors, fmt.Sprintf("choice group occurs %d times, minimum required is %d",
			validChoices, minOccurs))
	}
	if validChoices > maxOccurs {
		errors = append(errors, fmt.Sprintf("choice group occurs %d times, maximum allowed is %d",
			validChoices, maxOccurs))
	}

	return errors
}

// validateSequence checks whether the child elements satisfy an XSD <sequence> constraint.
func (v *Validator) validateSequence(children []*XMLNode, sequence *XSDSequence) []string {
	var errors []string
	expectedChildren := make(map[string]XSDElement)
	counts := make(map[string]int)

	for _, childDef := range sequence.Elements {
		expectedChildren[childDef.Name] = childDef
	}

	for _, child := range children {
		if childDef, ok := expectedChildren[child.Name]; ok {
			counts[child.Name]++
			errors = append(errors, v.validateElement(child, childDef)...)
		} else {
			errors = append(errors, fmt.Sprintf("unexpected element '%s'", child.Name))
		}
	}

	// Validate sequence occurrence constraints
	for _, childDef := range sequence.Elements {
		minOccurs := 1
		if childDef.MinOccurs != "" {
			if val, err := strconv.Atoi(childDef.MinOccurs); err == nil {
				minOccurs = val
			}
		}

		maxOccurs := 1
		if childDef.MaxOccurs != "" {
			if childDef.MaxOccurs == "unbounded" {
				maxOccurs = math.MaxInt32
			} else if val, err := strconv.Atoi(childDef.MaxOccurs); err == nil {
				maxOccurs = val
			}
		}

		count := counts[childDef.Name]
		if count < minOccurs {
			errors = append(errors, fmt.Sprintf("element '%s' occurs %d times, minimum required is %d",
				childDef.Name, count, minOccurs))
		}
		if count > maxOccurs {
			errors = append(errors, fmt.Sprintf("element '%s' occurs %d times, maximum allowed is %d",
				childDef.Name, count, maxOccurs))
		}
	}

	return errors
}

func (v *Validator) validateAttributes(nodeAttrs map[string]string, schemaAttrs []XSDAttribute) []string {
	errors := make([]string, 0, len(nodeAttrs)+len(schemaAttrs))

	// Create a map of required attributes from schema
	requiredAttrs := make(map[string]XSDAttribute)
	for _, attr := range schemaAttrs {
		if attr.Use == "required" {
			requiredAttrs[attr.Name] = attr
		}
	}

	// Check all attributes in node
	for name, value := range nodeAttrs {
		// Find corresponding schema attribute
		var found bool
		for _, schemaAttr := range schemaAttrs {
			if schemaAttr.Name == name {
				found = true
				delete(requiredAttrs, name) // Remove from required map if found

				// Validate attribute value (basic type checking)
				if err := v.validateAttributeValue(value, schemaAttr); err != nil {
					errors = append(errors, fmt.Sprintf("attribute '%s': %s", name, err))
				}
				break
			}
		}

		if !found {
			errors = append(errors, fmt.Sprintf("unexpected attribute '%s'", name))
		}
	}

	// Check if any required attributes are missing
	for name := range requiredAttrs {
		errors = append(errors, fmt.Sprintf("missing required attribute '%s'", name))
	}

	return errors
}

// validateAttributeValue calls the appropriate type validation
func (v *Validator) validateAttributeValue(value string, attr XSDAttribute) error {
	if attr.Type != "" {
		return v.validateType(value, attr.Type, nil)
	} else if attr.SimpleType != nil && attr.SimpleType.Restriction != nil {
		return v.validateType(value, attr.SimpleType.Restriction.Base, attr.SimpleType.Restriction)
	}
	return nil
}

// validateElementContent validates element content
func (v *Validator) validateElementContent(content string, element *XSDElement) error {
	if element.SimpleType != nil {
		return v.validateType(content, element.SimpleType.Restriction.Base, element.SimpleType.Restriction)
	} else if element.Type != "" {
		return v.validateType(content, element.Type, nil)
	}
	return nil
}
