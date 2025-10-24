package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// This tool converts TypeScript-ESLint test cases to RSLint format
// TypeScript-ESLint uses the same format as ESLint but with TypeScript-specific options
func main() {
	inputFile := flag.String("input", "", "Input TypeScript-ESLint test file (JSON)")
	outputFile := flag.String("output", "", "Output RSLint test file (JSON)")
	verbose := flag.Bool("verbose", false, "Verbose output")
	generateGo := flag.Bool("go", false, "Generate Go test file instead of JSON")
	ruleName := flag.String("rule", "", "Rule name for Go test generation")

	flag.Parse()

	if *inputFile == "" {
		fmt.Fprintf(os.Stderr, "Error: -input flag is required\n")
		flag.Usage()
		os.Exit(1)
	}

	if *outputFile == "" {
		// Generate output filename from input
		ext := filepath.Ext(*inputFile)
		if *generateGo {
			*outputFile = (*inputFile)[0:len(*inputFile)-len(ext)] + "_test.go"
		} else {
			*outputFile = (*inputFile)[0:len(*inputFile)-len(ext)] + "_rslint" + ext
		}
	}

	if *generateGo && *ruleName == "" {
		fmt.Fprintf(os.Stderr, "Error: -rule flag is required when -go is specified\n")
		flag.Usage()
		os.Exit(1)
	}

	if *verbose {
		fmt.Printf("Converting TypeScript-ESLint tests from %s to %s\n", *inputFile, *outputFile)
	}

	// Load TypeScript-ESLint test suite (same format as ESLint)
	eslintSuite, err := rule_tester.LoadESLintTestSuiteFromJSON(*inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading TypeScript-ESLint test suite: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Printf("Loaded %d valid and %d invalid test cases\n", len(eslintSuite.Valid), len(eslintSuite.Invalid))
	}

	// Convert to RSLint format
	suite := rule_tester.ConvertESLintTestSuite(eslintSuite)

	if *generateGo {
		// Generate Go test file
		goCode := generateGoTestFile(*ruleName, suite)
		if err := os.WriteFile(*outputFile, []byte(goCode), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing Go test file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully generated Go test file: %s\n", *outputFile)
	} else {
		// Write JSON output
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
	}

	fmt.Printf("  Valid test cases: %d\n", len(suite.Valid))
	fmt.Printf("  Invalid test cases: %d\n", len(suite.Invalid))
}

func generateGoTestFile(ruleName string, suite *rule_tester.TestSuite) string {
	code := fmt.Sprintf(`package %s

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func Test%sRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&%sRule,
		[]rule_tester.ValidTestCase{
`, ruleName, capitalize(ruleName), capitalize(ruleName))

	// Add valid test cases
	for _, tc := range suite.Valid {
		code += fmt.Sprintf("\t\t\t{Code: %q},\n", tc.Code)
	}

	code += "\t\t},\n\t\t[]rule_tester.InvalidTestCase{\n"

	// Add invalid test cases
	for _, tc := range suite.Invalid {
		code += "\t\t\t{\n"
		code += fmt.Sprintf("\t\t\t\tCode: %q,\n", tc.Code)
		code += "\t\t\t\tErrors: []rule_tester.InvalidTestCaseError{\n"
		for _, err := range tc.Errors {
			code += "\t\t\t\t\t{\n"
			code += fmt.Sprintf("\t\t\t\t\t\tMessageId: %q,\n", err.MessageId)
			if err.Line != 0 {
				code += fmt.Sprintf("\t\t\t\t\t\tLine: %d,\n", err.Line)
			}
			if err.Column != 0 {
				code += fmt.Sprintf("\t\t\t\t\t\tColumn: %d,\n", err.Column)
			}
			code += "\t\t\t\t\t},\n"
		}
		code += "\t\t\t\t},\n"
		if len(tc.Output) > 0 {
			code += "\t\t\t\tOutput: []string{\n"
			for _, out := range tc.Output {
				code += fmt.Sprintf("\t\t\t\t\t%q,\n", out)
			}
			code += "\t\t\t\t},\n"
		}
		code += "\t\t\t},\n"
	}

	code += "\t\t},\n\t)\n}\n"
	return code
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:]
}
