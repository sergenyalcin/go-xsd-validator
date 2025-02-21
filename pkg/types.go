package pkg

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

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

// validateType verifies that a value conforms to the given XSD type.
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

func (v *Validator) validateBaseType(value string, typeName string) error { //nolint:gocyclo
	switch typeName {
	case "xs:string", "string":
		return nil // Always valid

	case "xs:integer", "integer":
		if _, err := strconv.ParseInt(value, 10, 64); err != nil {
			return fmt.Errorf("invalid integer value: %s", value)
		}
	case "xs:int", "int":
		if _, err := strconv.ParseInt(value, 10, 32); err != nil {
			return fmt.Errorf("invalid int value: %s", value)
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
