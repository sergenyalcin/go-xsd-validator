package pkg

import (
	"fmt"
	"regexp"
	"strconv"
	"unicode/utf8"
)

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
