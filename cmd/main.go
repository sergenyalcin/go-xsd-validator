package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sergenyalcin/go-xsd-validator/pkg"
)

func main() {
	xmlPath := flag.String("xml", "", "Path to XML file (required)")
	xsdPath := flag.String("xsd", "", "Path to XSD schema file (required)")
	outputFormat := flag.String("format", "text", "Output format (text, json)")
	flag.Parse()

	if *xmlPath == "" || *xsdPath == "" {
		flag.Usage()
		os.Exit(1)
	}
	run(xsdPath, xmlPath, outputFormat)
}

func run(xsdPath, xmlPath, outputFormat *string) {
	// Read XSD file
	xsdFile, err := os.Open(*xsdPath)
	if err != nil {
		if _, err := fmt.Fprintf(os.Stderr, "Error opening XSD file: %v\n", err); err != nil {
			panic(err)
		}
		os.Exit(1)
	}
	defer func(xsdFile *os.File) {
		if err := xsdFile.Close(); err != nil {
			os.Exit(1)
		}
	}(xsdFile)

	// Create validator
	validator, err := pkg.NewValidator(xsdFile)
	if err != nil {
		if _, err := fmt.Fprintf(os.Stderr, "Error creating validator: %v\n", err); err != nil {
			panic(err)
		}
		os.Exit(1)
	}

	// Read XML file
	xmlFile, err := os.Open(*xmlPath)
	if err != nil {
		if _, err := fmt.Fprintf(os.Stderr, "Error opening XML file: %v\n", err); err != nil {
			panic(err)
		}
		os.Exit(1)
	}
	defer func(xmlFile *os.File) {
		if err := xmlFile.Close(); err != nil {
			panic(err)
		}
	}(xmlFile)

	// Validate XML
	result, err := validator.Validate(xmlFile)
	if err != nil {
		if _, err := fmt.Fprintf(os.Stderr, "Error during validation: %v\n", err); err != nil {
			panic(err)
		}
		os.Exit(1)
	}

	// Output results
	result.OutputResult(*outputFormat)
}
