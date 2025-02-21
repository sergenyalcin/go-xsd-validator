package main

import (
	"testing"
)

func Test(t *testing.T) {
	tests := []struct {
		xsdPath      string
		xmlPath      string
		outputFormat string
	}{
		{
			xsdPath:      "../testdata/xsd/book.xsd",
			xmlPath:      "../testdata/xml/book.xml",
			outputFormat: "text",
		},
		{
			xsdPath:      "../testdata/xsd/complex_required.xsd",
			xmlPath:      "../testdata/xml/complex_required.xml",
			outputFormat: "text",
		},
		{
			xsdPath:      "../testdata/xsd/deeply_nested.xsd",
			xmlPath:      "../testdata/xml/deeply_nested.xml",
			outputFormat: "text",
		},
		{
			xsdPath:      "../testdata/xsd/employee_directory.xsd",
			xmlPath:      "../testdata/xml/employee_directory.xml",
			outputFormat: "text",
		},
		{
			xsdPath:      "../testdata/xsd/minmax.xsd",
			xmlPath:      "../testdata/xml/minmax.xml",
			outputFormat: "text",
		},
		{
			xsdPath:      "../testdata/xsd/ns.xsd",
			xmlPath:      "../testdata/xml/ns.xml",
			outputFormat: "text",
		},
		{
			xsdPath:      "../testdata/xsd/purchase_order.xsd",
			xmlPath:      "../testdata/xml/purchase_order.xml",
			outputFormat: "text",
		},
		{
			xsdPath:      "../testdata/xsd/rec_structs.xsd",
			xmlPath:      "../testdata/xml/rec_structs.xml",
			outputFormat: "text",
		},
		{
			xsdPath:      "../testdata/xsd/regex.xsd",
			xmlPath:      "../testdata/xml/regex.xml",
			outputFormat: "text",
		},
	}

	for _, tc := range tests {
		run(&tc.xsdPath, &tc.xmlPath, &tc.outputFormat)
	}
}
