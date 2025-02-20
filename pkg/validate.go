package pkg

import (
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Validator type
type Validator struct {
	schema     *XSDSchema
	patterns   *PatternCache
	namespaces map[string]string
	defaultNS  string
}

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

// Main validation logic
func (v *Validator) Validate(xmlFile io.Reader) (*ValidationResult, error) {
	xmlNode, err := ParseXML(xmlFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML: %v", err)
	}

	rootXsd := v.findSchemaElementNS(xmlNode.Name, xmlNode.Namespace, v.schema.Elements)
	if rootXsd == nil {
		return nil, fmt.Errorf("root element '{%s}%s' not defined in schema", xmlNode.Namespace, xmlNode.Name)
	}

	result := &ValidationResult{
		Valid:    true,
		Filename: xmlNode.Name,
		Errors:   v.validateElement(xmlNode, *rootXsd),
	}

	if len(result.Errors) > 0 {
		result.Valid = false
	}

	return result, nil
}

func (v *Validator) validateElement(xmlNode *XMLNode, xsdElem XSDElement) []string {
	var errors []string

	if xsdElem.Ref != "" {
		refElement, err := v.resolveElementRef(xsdElem.Ref)
		if err != nil {
			return append(errors, err.Error())
		}
		return v.validateElement(xmlNode, *refElement)
	}

	// Validate element name and namespace
	if !v.validateElementNameAndNS(xmlNode, xsdElem) {
		errors = append(errors, fmt.Sprintf("element name or namespace mismatch: expected '{%s}%s', got '{%s}%s'",
			xsdElem.Namespace, xsdElem.Name, xmlNode.Namespace, xmlNode.Name))
		return errors
	}

	// Resolve complex type if referenced by name
	if xsdElem.ComplexType == nil && xsdElem.Type != "" {
		if ct := v.findComplexType(xsdElem.Type); ct != nil {
			xsdElem.ComplexType = ct
		}
	}

	// Validate attributes
	if xsdElem.ComplexType != nil {
		errors = append(errors, v.validateAttributes(xmlNode.Attributes, xsdElem.ComplexType.Attributes)...)
	}

	// Validate content
	if xmlNode.Content != "" {
		if err := v.validateElementContent(xmlNode.Content, &xsdElem); err != nil {
			errors = append(errors, fmt.Sprintf("invalid content in element '%s': %v", xmlNode.Name, err))
		}
	}

	// Validate children based on complex type
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

func (v *Validator) validateChoice(children []*XMLNode, choice *XSDChoice) []string {
	var errors []string

	// Get occurrence constraints
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

	if choice.Choice != nil {
		errors = append(errors, v.validateChoice(children, choice.Choice)...)
	}

	// Track which choice elements were found
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

	// Validate occurrence constraints
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

func (v *Validator) validateElementNameAndNS(xmlNode *XMLNode, xsdElem XSDElement) bool {
	if xmlNode.Name != xsdElem.Name {
		return false
	}

	// Handle namespace validation
	schemaNamespace := xsdElem.Namespace
	if schemaNamespace == "" {
		schemaNamespace = v.schema.TargetNS
	}

	// For qualified elements
	if v.schema.ElementFormDefault == "qualified" {
		if schemaNamespace == "" {
			return xmlNode.Namespace == ""
		}
		return xmlNode.Namespace == schemaNamespace
	}

	// For unqualified elements
	return true
}

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

// Extended type validation functions
func (v *Validator) validateType(value string, typeName string, restrictions *XSDRestriction) error {
	// First validate base type
	if err := v.validateBaseType(value, typeName); err != nil {
		return err
	}

	// Then apply any restrictions
	if restrictions != nil {
		if err := v.validateRestrictions(value, typeName, restrictions); err != nil {
			return err
		}
	}

	return nil
}

func (v *Validator) validateBaseType(value string, typeName string) error {
	switch typeName {
	case "xs:string", "string":
		return nil // Always valid

	case "xs:integer", "integer":
		if _, err := strconv.ParseInt(value, 10, 64); err != nil {
			return fmt.Errorf("invalid integer value: %s", value)
		}
	case "xs:positiveInteger":
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid positive integer value: %s", value)
		}
		if val <= 0 {
			return fmt.Errorf("value must be positive, got %d", val)
		}
	case "xs:decimal", "decimal":
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return fmt.Errorf("invalid decimal value: %s", value)
		}
	case "xs:float", "float":
		if _, err := strconv.ParseFloat(value, 32); err != nil {
			return fmt.Errorf("invalid float value: %s", value)
		}
	case "xs:double", "double":
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return fmt.Errorf("invalid double value: %s", value)
		}
	case "xs:long", "long":
		if _, err := strconv.ParseInt(value, 10, 64); err != nil {
			return fmt.Errorf("invalid long value: %s", value)
		}
	case "xs:boolean", "boolean":
		if value != "true" && value != "false" && value != "1" && value != "0" {
			return fmt.Errorf("invalid boolean value: %s", value)
		}
	case "xs:date", "date":
		if _, err := time.Parse("2006-01-02", value); err != nil {
			return fmt.Errorf("invalid date value: %s", value)
		}
	case "xs:time", "time":
		if _, err := time.Parse("15:04:05", value); err != nil {
			return fmt.Errorf("invalid time value: %s", value)
		}
	case "xs:dateTime", "dateTime":
		if _, err := time.Parse("2006-01-02T15:04:05", value); err != nil {
			return fmt.Errorf("invalid dateTime value: %s", value)
		}
	case "xs:duration", "duration":
		if err := validateDuration(value); err != nil {
			return fmt.Errorf("invalid duration value: %s", value)
		}
	case "xs:gYear", "gYear":
		if matched, _ := regexp.MatchString(`^-?\d{4}$`, value); !matched {
			return fmt.Errorf("invalid gYear value: %s", value)
		}
	case "xs:gYearMonth", "gYearMonth":
		if matched, _ := regexp.MatchString(`^-?\d{4}-\d{2}$`, value); !matched {
			return fmt.Errorf("invalid gYearMonth value: %s", value)
		}
	case "xs:hexBinary", "hexBinary":
		if matched, _ := regexp.MatchString(`^[0-9a-fA-F]*$`, value); !matched {
			return fmt.Errorf("invalid hexBinary value: %s", value)
		}
	case "xs:base64Binary", "base64Binary":
		if matched, _ := regexp.MatchString(`^[A-Za-z0-9+/]*={0,2}$`, value); !matched {
			return fmt.Errorf("invalid base64Binary value: %s", value)
		}
	case "xs:anyURI", "anyURI":
		if matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9+.-]*:`, value); !matched {
			return fmt.Errorf("invalid anyURI value: %s", value)
		}
	default:
		// Not a built-in type. Try to resolve it as a user-defined simple type.
		simpleType := v.findSimpleType(typeName)
		if simpleType == nil {
			return fmt.Errorf("unsupported type: %s", typeName)
		}
		// Validate using the simple type's restriction.
		if simpleType.Restriction != nil {
			return v.validateType(value, simpleType.Restriction.Base, simpleType.Restriction)
		}
		return fmt.Errorf("simple type %s has no restriction", typeName)
	}
	return nil
}

func (v *Validator) validateRestrictions(value string, baseType string, restrictions *XSDRestriction) error {
	// Length restrictions
	if restrictions.Length.Value != "" {
		length, _ := strconv.Atoi(restrictions.Length.Value)
		actualLen := utf8.RuneCountInString(value)
		if actualLen != length {
			return fmt.Errorf("length must be exactly %d, got %d", length, actualLen)
		}
	}

	if restrictions.MinLength.Value != "" {
		minLength, _ := strconv.Atoi(restrictions.MinLength.Value)
		actualLen := utf8.RuneCountInString(value)
		if actualLen < minLength {
			return fmt.Errorf("length must be at least %d, got %d", minLength, actualLen)
		}
	}

	if restrictions.MaxLength.Value != "" {
		maxLength, _ := strconv.Atoi(restrictions.MaxLength.Value)
		actualLen := utf8.RuneCountInString(value)
		if actualLen > maxLength {
			return fmt.Errorf("length must be at most %d, got %d", maxLength, actualLen)
		}
	}

	// Numeric restrictions
	switch baseType {
	case "xs:decimal", "xs:integer", "xs:float", "xs:double":
		num, _ := strconv.ParseFloat(value, 64)

		if restrictions.MinInclusive.Value != "" {
			minInclusive, err := strconv.ParseFloat(restrictions.MinInclusive.Value, 64)
			if err == nil && num < minInclusive {
				return fmt.Errorf("value must be >= %v, got %v", minInclusive, num)
			}
		}

		if restrictions.MaxInclusive.Value != "" {
			maxInclusive, err := strconv.ParseFloat(restrictions.MaxInclusive.Value, 64)
			if err == nil && num > maxInclusive {
				return fmt.Errorf("value must be <= %v, got %v", maxInclusive, num)
			}
		}

		if restrictions.MinExclusive.Value != "" {
			minExclusive, err := strconv.ParseFloat(restrictions.MinExclusive.Value, 64)
			if err == nil && num <= minExclusive {
				return fmt.Errorf("value must be > %v, got %v", minExclusive, num)
			}
		}

		if restrictions.MaxExclusive.Value != "" {
			maxExclusive, err := strconv.ParseFloat(restrictions.MaxExclusive.Value, 64)
			if err == nil && num >= maxExclusive {
				return fmt.Errorf("value must be < %v, got %v", maxExclusive, num)
			}
		}
	}

	// Pattern restrictions
	for _, pattern := range restrictions.Pattern {
		matched, err := regexp.MatchString(pattern.Value, value)
		if err != nil {
			return fmt.Errorf("invalid pattern: %s", pattern.Value)
		}
		if !matched {
			return fmt.Errorf("value does not match pattern: %s", pattern.Value)
		}
	}

	// Enumeration restrictions
	if len(restrictions.Enumeration) > 0 {
		valid := false
		for _, enum := range restrictions.Enumeration {
			if value == enum.Value {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("value must be one of the enumerated values")
		}
	}

	return nil
}

// Helper function for duration validation
func validateDuration(value string) error {
	// Duration format: -?P([0-9]+Y)?([0-9]+M)?([0-9]+D)?(T([0-9]+H)?([0-9]+M)?([0-9]+(\.[0-9]+)?S)?)?
	pattern := `^-?P(([0-9]+Y)?([0-9]+M)?([0-9]+D)?)?(T([0-9]+H)?([0-9]+M)?([0-9]+(\.[0-9]+)?S)?)?$`
	if matched, _ := regexp.MatchString(pattern, value); !matched {
		return fmt.Errorf("invalid duration format")
	}
	return nil
}

func (v *Validator) findSchemaElement(name string, elements []XSDElement) *XSDElement {
	for i, elem := range elements {
		if elem.Name == name {
			return &elements[i]
		}
	}
	return nil
}

func (v *Validator) validateAttributes(nodeAttrs map[string]string, schemaAttrs []XSDAttribute) []string {
	var errors []string

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

// Update the existing validateAttributeValue method to use the new type validation
func (v *Validator) validateAttributeValue(value string, attr XSDAttribute) error {
	if attr.Type != "" {
		return v.validateType(value, attr.Type, nil)
	} else if attr.SimpleType != nil && attr.SimpleType.Restriction != nil {
		return v.validateType(value, attr.SimpleType.Restriction.Base, attr.SimpleType.Restriction)
	}
	return nil
}

// Add this method to validate element content
func (v *Validator) validateElementContent(content string, element *XSDElement) error {
	if element.SimpleType != nil {
		return v.validateType(content, element.SimpleType.Restriction.Base, element.SimpleType.Restriction)
	} else if element.Type != "" {
		return v.validateType(content, element.Type, nil)
	}
	return nil
}

func (v *Validator) findSimpleType(name string) *XSDSimpleType {
	for i, st := range v.schema.SimpleTypes {
		if st.Name == name {
			return &v.schema.SimpleTypes[i]
		}
	}
	return nil
}

func (v *Validator) findComplexType(name string) *XSDComplexType {
	for i, ct := range v.schema.ComplexTypes {
		if ct.Name == name {
			return &v.schema.ComplexTypes[i]
		}
	}
	return nil
}

// Helper functions for XSD pattern conversion
func convertXSDPatternToGoRegex(pattern string) string {
	// Convert XSD pattern syntax to Go regex syntax
	// This is a simplified version - extend as needed
	pattern = strings.ReplaceAll(pattern, "\\i\\c*", "[a-zA-Z][a-zA-Z0-9_]*")
	pattern = strings.ReplaceAll(pattern, "\\c", "[a-zA-Z0-9_]")
	pattern = strings.ReplaceAll(pattern, "\\d", "[0-9]")
	pattern = strings.ReplaceAll(pattern, "\\w", "[a-zA-Z0-9_]")
	return "^" + pattern + "$"
}
