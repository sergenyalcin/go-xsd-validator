package pkg

import (
	"encoding/xml"
	"regexp"
)

// XSDSchema ore types for XML Schema representation
type XSDSchema struct {
	XMLName            xml.Name `xml:"schema"`
	TargetNS           string   `xml:"targetNamespace,attr"`
	ElementFormDefault string   `xml:"elementFormDefault,attr"`
	Xmlns              map[string]string
	Elements           []XSDElement     `xml:"element"`
	ComplexTypes       []XSDComplexType `xml:"complexType"`
	SimpleTypes        []XSDSimpleType  `xml:"simpleType"`
}

type XSDElement struct {
	Name        string          `xml:"name,attr"`
	Namespace   string          `xml:"namespace,attr"`
	Type        string          `xml:"type,attr"`
	Ref         string          `xml:"ref,attr"`
	MinOccurs   string          `xml:"minOccurs,attr"`
	MaxOccurs   string          `xml:"maxOccurs,attr"`
	ComplexType *XSDComplexType `xml:"complexType"`
	SimpleType  *XSDSimpleType  `xml:"simpleType"`
}

type XSDElementRef struct {
	Ref       string `xml:"ref,attr"`
	MinOccurs string `xml:"minOccurs,attr"`
	MaxOccurs string `xml:"maxOccurs,attr"`
}

type XSDComplexType struct {
	Name       string         `xml:"name,attr"`
	Sequence   *XSDSequence   `xml:"sequence"`
	Choice     *XSDChoice     `xml:"choice"`
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

type XSDChoice struct {
	MinOccurs string       `xml:"minOccurs,attr"`
	MaxOccurs string       `xml:"maxOccurs,attr"`
	Choice    *XSDChoice   `xml:"choice"`
	Elements  []XSDElement `xml:"element"`
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
	Base           string     `xml:"base,attr"`
	Pattern        []XSDValue `xml:"pattern"`
	Enumeration    []XSDValue `xml:"enumeration"`
	Length         XSDValue   `xml:"length"`
	MinLength      XSDValue   `xml:"minLength"`
	MaxLength      XSDValue   `xml:"maxLength"`
	MinInclusive   XSDValue   `xml:"minInclusive"`
	MaxInclusive   XSDValue   `xml:"maxInclusive"`
	MinExclusive   XSDValue   `xml:"minExclusive"`
	MaxExclusive   XSDValue   `xml:"maxExclusive"`
	WhiteSpace     string     `xml:"whiteSpace,attr"`
	TotalDigits    string     `xml:"totalDigits,attr"`
	FractionDigits string     `xml:"fractionDigits,attr"`
}

type XSDValue struct {
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
