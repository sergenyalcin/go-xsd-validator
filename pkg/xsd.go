package pkg

import (
	"encoding/xml"
	"regexp"
)

// XSDSchema ore types for XML Schema representation
type XSDSchema struct {
	XMLName      xml.Name         `xml:"schema"`
	Elements     []XSDElement     `xml:"element"`
	ComplexTypes []XSDComplexType `xml:"complexType"`
	SimpleTypes  []XSDSimpleType  `xml:"simpleType"`
}

type XSDElement struct {
	Name        string          `xml:"name,attr"`
	Type        string          `xml:"type,attr"`
	MinOccurs   string          `xml:"minOccurs,attr"`
	MaxOccurs   string          `xml:"maxOccurs,attr"`
	ComplexType *XSDComplexType `xml:"complexType"`
	SimpleType  *XSDSimpleType  `xml:"simpleType"`
}

type XSDComplexType struct {
	Name       string         `xml:"name,attr"`
	Sequence   *XSDSequence   `xml:"sequence"`
	Attributes []XSDAttribute `xml:"attribute"`
}

type XSDSimpleType struct {
	Name        string          `xml:"name,attr"`
	Restriction *XSDRestriction `xml:"restriction"`
	Union       *XSDUnion       `xml:"union"`
	List        *XSDList        `xml:"list"`
}

type XSDSequence struct {
	Elements []XSDElement `xml:"element"`
}

type XSDAttribute struct {
	Name       string         `xml:"name,attr"`
	Type       string         `xml:"type,attr"`
	Use        string         `xml:"use,attr"`
	Default    string         `xml:"default,attr"`
	Fixed      string         `xml:"fixed,attr"`
	SimpleType *XSDSimpleType `xml:"simpleType"`
}

type XSDRestriction struct {
	Base           string           `xml:"base,attr"`
	Pattern        []XSDPattern     `xml:"pattern"`
	Enumeration    []XSDEnumeration `xml:"enumeration"`
	Length         string           `xml:"length,attr"`
	MinLength      string           `xml:"minLength,attr"`
	MaxLength      string           `xml:"maxLength,attr"`
	MinInclusive   string           `xml:"minInclusive,attr"`
	MaxInclusive   string           `xml:"maxInclusive,attr"`
	MinExclusive   string           `xml:"minExclusive,attr"`
	MaxExclusive   string           `xml:"maxExclusive,attr"`
	WhiteSpace     string           `xml:"whiteSpace,attr"`
	TotalDigits    string           `xml:"totalDigits,attr"`
	FractionDigits string           `xml:"fractionDigits,attr"`
}

type XSDPattern struct {
	Value string `xml:"value,attr"`
}

type XSDEnumeration struct {
	Value string `xml:"value,attr"`
}

type XSDUnion struct {
	MemberTypes []string        `xml:"memberTypes,attr"`
	SimpleTypes []XSDSimpleType `xml:"simpleType"`
}

type XSDList struct {
	ItemType string `xml:"itemType,attr"`
}

// Pattern cache to improve performance
type PatternCache struct {
	patterns map[string]*regexp.Regexp
}

func NewPatternCache() *PatternCache {
	return &PatternCache{
		patterns: make(map[string]*regexp.Regexp),
	}
}

func (pc *PatternCache) GetPattern(pattern string) (*regexp.Regexp, error) {
	if compiled, ok := pc.patterns[pattern]; ok {
		return compiled, nil
	}

	// Convert XSD pattern to Go regex
	goPattern := convertXSDPatternToGoRegex(pattern)
	compiled, err := regexp.Compile(goPattern)
	if err != nil {
		return nil, err
	}

	pc.patterns[pattern] = compiled
	return compiled, nil
}
