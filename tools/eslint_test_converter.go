package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// This tool converts ESLint test cases to RSLint format
func main() {
	inputFile := flag.String("input", "", "Input ESLint test file (JSON)")
	outputFile := flag.String("output", "", "Output RSLint test file (JSON)")
	verbose := flag.Bool("verbose", false, "Verbose output")

	flag.Parse()

	if *inputFile == "" {
		fmt.Fprintf(os.Stderr, "Error: -input flag is required\n")
		flag.Usage()
		os.Exit(1)
	}

	if *outputFile == "" {
		// Generate output filename from input
		ext := filepath.Ext(*inputFile)
		*outputFile = (*inputFile)[0:len(*inputFile)-len(ext)] + "_rslint" + ext
	}

	if *verbose {
		fmt.Printf("Converting ESLint tests from %s to %s\n", *inputFile, *outputFile)
	}

	// Load ESLint test suite
	eslintSuite, err := rule_tester.LoadESLintTestSuiteFromJSON(*inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading ESLint test suite: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Printf("Loaded %d valid and %d invalid test cases\n", len(eslintSuite.Valid), len(eslintSuite.Invalid))
	}

	// Convert to RSLint format
	suite := rule_tester.ConvertESLintTestSuite(eslintSuite)

	// Write output
	data, err := json.MarshalIndent(suite, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling test suite: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*outputFile, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully converted tests to %s\n", *outputFile)
	fmt.Printf("  Valid test cases: %d\n", len(suite.Valid))
	fmt.Printf("  Invalid test cases: %d\n", len(suite.Invalid))
}
