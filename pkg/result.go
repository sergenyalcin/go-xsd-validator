package pkg

import (
	"encoding/json"
	"fmt"
	"os"
)

// ValidationResult includes the results of the XML | XSD validation
type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Filename string   `json:"filename"`
	Errors   []string `json:"errors,omitempty"`
}

// OutputResult is responsible on output formatting
func (r *ValidationResult) OutputResult(format string) {
	switch format {
	case "json":
		output, err := json.MarshalIndent(r, "", "  ")
		if err != nil {
			if _, err := fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err); err != nil {
				panic(err)
			}
		}
		fmt.Println(string(output))
	default:
		if r.Valid {
			fmt.Printf("✓ XML file '%s' is valid\n", r.Filename)
		} else {
			fmt.Printf("✗ XML file '%s' is invalid:\n", r.Filename)
			for _, err := range r.Errors {
				fmt.Printf("  - %s\n", err)
			}
		}
	}
}
