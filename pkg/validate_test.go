package pkg

import (
	"bytes"
	"testing"
)

func TestXMLValidator(t *testing.T) {
	tests := []struct {
		name     string
		xmlInput string
		xsdInput string
		valid    bool
		errors   []string
	}{
		{
			name:     "Basic Validation - Valid XML",
			xmlInput: `<?xml version="1.0" encoding="UTF-8"?><name>John Doe</name>`,
			xsdInput: `<?xml version="1.0" encoding="UTF-8"?><xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:element name="name" type="xs:string"/></xs:schema>`,
			valid:    true,
		},
		{
			name:     "Missing Required Attribute",
			xmlInput: `<?xml version="1.0" encoding="UTF-8"?><employee></employee>`,
			xsdInput: `<?xml version="1.0" encoding="UTF-8"?><xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:element name="employee"><xs:complexType><xs:attribute name="id" type="xs:positiveInteger" use="required"/></xs:complexType></xs:element></xs:schema>`,
			valid:    false,
			errors:   []string{"missing required attribute 'id'"},
		},
		{
			name:     "Complex Type Validation",
			xmlInput: `<?xml version="1.0" encoding="UTF-8"?><person><name>John</name><age>30</age></person>`,
			xsdInput: `<?xml version="1.0" encoding="UTF-8"?><xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:element name="person"><xs:complexType><xs:sequence><xs:element name="name" type="xs:string"/><xs:element name="age" type="xs:integer"/></xs:sequence></xs:complexType></xs:element></xs:schema>`,
			valid:    true,
		},
		{
			name:     "Enumeration Validation - Invalid Value",
			xmlInput: `<?xml version="1.0" encoding="UTF-8"?><priority>urgent</priority>`,
			xsdInput: `<?xml version="1.0" encoding="UTF-8"?><xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:element name="priority"><xs:simpleType><xs:restriction base="xs:string"><xs:enumeration value="high"/><xs:enumeration value="medium"/><xs:enumeration value="low"/></xs:restriction></xs:simpleType></xs:element></xs:schema>`,
			valid:    false,
			errors:   []string{"invalid content in element 'priority': value must be one of the enumerated values"},
		},
		{
			name:     "Pattern Restriction - Valid",
			xmlInput: `<?xml version="1.0" encoding="UTF-8"?><code>AB-123</code>`,
			xsdInput: `<?xml version="1.0" encoding="UTF-8"?><xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:element name="code"><xs:simpleType><xs:restriction base="xs:string"><xs:pattern value="[A-Z]{2}-[0-9]{3}"/></xs:restriction></xs:simpleType></xs:element></xs:schema>`,
			valid:    true,
		},
		{
			name:     "Pattern Restriction - Invalid",
			xmlInput: `<?xml version="1.0" encoding="UTF-8"?><code>123-AB</code>`,
			xsdInput: `<?xml version="1.0" encoding="UTF-8"?><xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:element name="code"><xs:simpleType><xs:restriction base="xs:string"><xs:pattern value="[A-Z]{2}-[0-9]{3}"/></xs:restriction></xs:simpleType></xs:element></xs:schema>`,
			valid:    false,
			errors:   []string{"invalid content in element 'code': value does not match pattern: [A-Z]{2}-[0-9]{3}"},
		},
		{
			name:     "Min and Max Inclusive - Invalid",
			xmlInput: `<?xml version="1.0" encoding="UTF-8"?><price>0</price>`,
			xsdInput: `<?xml version="1.0" encoding="UTF-8"?><xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:element name="price"><xs:simpleType><xs:restriction base="xs:decimal"><xs:minInclusive value="1"/><xs:maxInclusive value="100"/></xs:restriction></xs:simpleType></xs:element></xs:schema>`,
			valid:    false,
			errors:   []string{"invalid content in element 'price': value must be >= 1, got 0"},
		},
		{
			name:     "Missing Child Element",
			xmlInput: `<?xml version="1.0" encoding="UTF-8"?><book><title>Go in Action</title></book>`,
			xsdInput: `<?xml version="1.0" encoding="UTF-8"?><xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:element name="book"><xs:complexType><xs:sequence><xs:element name="title" type="xs:string"/><xs:element name="author" type="xs:string"/></xs:sequence></xs:complexType></xs:element></xs:schema>`,
			valid:    false,
			errors:   []string{"element 'author' occurs 0 times, minimum required is 1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			xsdReader := bytes.NewReader([]byte(tc.xsdInput))
			xmlReader := bytes.NewReader([]byte(tc.xmlInput))

			validator, err := NewValidator(xsdReader)
			if err != nil {
				t.Fatalf("Failed to create validator: %v", err)
			}

			result, err := validator.Validate(xmlReader)
			if err != nil {
				t.Fatalf("Validation error: %v", err)
			}

			if result.Valid != tc.valid {
				t.Errorf("Expected valid=%v, got valid=%v", tc.valid, result.Valid)
			}

			if !tc.valid && len(tc.errors) > 0 {
				for _, expectedErr := range tc.errors {
					found := false
					for _, actualErr := range result.Errors {
						if actualErr == expectedErr {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error: %q, but not found in actual errors: %v", expectedErr, result.Errors)
					}
				}
			}
		})
	}
}
