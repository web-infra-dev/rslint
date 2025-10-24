package rule_tester_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Example_basicUsage demonstrates basic rule testing
func Example_basicUsage() {
	// In a real test, you would pass *testing.T
	var t *testing.T

	// Create mock rule for demonstration (in real tests, use an actual rule like &dotNotationRule)
	var testRule *rule.Rule // Replace with actual rule in real tests

	rule_tester.RunRuleTester(
		"/path/to/root",
		"tsconfig.json",
		t,
		testRule,
		// Valid test cases
		[]rule_tester.ValidTestCase{
			{Code: "const x = 1;"},
			{Code: "let y = 2;"},
		},
		// Invalid test cases
		[]rule_tester.InvalidTestCase{
			{
				Code: "var x = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 1},
				},
				Output: []string{"const x = 1;"},
			},
		},
	)
}

// Example_batchTestBuilder demonstrates using the BatchTestBuilder
func Example_batchTestBuilder() {
	builder := rule_tester.NewBatchTestBuilder()

	// Build test suite programmatically
	builder.
		AddValid("const x = 1;").
		AddValid("let y = 2;").
		AddInvalid("var x = 1;", "useConst", 1, 1, "const x = 1;").
		AddInvalid("var y = 2;", "useConst", 1, 1, "const y = 2;")

	valid, invalid := builder.Build()

	// Use in tests
	_ = valid
	_ = invalid
}

// Example_commonFixtures demonstrates using CommonFixtures
func Example_commonFixtures() {
	fixtures := rule_tester.NewCommonFixtures()

	// Generate common TypeScript patterns
	classCode := fixtures.Class("MyClass",
		fixtures.Method("myMethod", "x: number", "string", "return x.toString();"))

	interfaceCode := fixtures.Interface("MyInterface",
		fixtures.Property("prop", "string", ""))

	functionCode := fixtures.Function("myFunc", "x: number", "number", "return x * 2;")

	// Use in tests
	_ = classCode
	_ = interfaceCode
	_ = functionCode
}

// Example_loadFromJSON demonstrates loading tests from JSON
func Example_loadFromJSON() {
	var t *testing.T
	var testRule *rule.Rule // Replace with actual rule in real tests

	// Load tests from JSON file
	err := rule_tester.RunRuleTesterFromJSON(
		"/path/to/root",
		"tsconfig.json",
		"testdata/my_rule_tests.json",
		t,
		testRule,
	)

	_ = err // Handle error in real code
}

// Example_convertESLint demonstrates converting ESLint tests
func Example_convertESLint() {
	// Load ESLint test suite
	eslintSuite, err := rule_tester.LoadESLintTestSuiteFromJSON("eslint_tests.json")
	if err != nil {
		panic(err)
	}

	// Convert to RSLint format
	suite := rule_tester.ConvertESLintTestSuite(eslintSuite)

	// Use converted tests
	_ = suite.Valid
	_ = suite.Invalid
}

// Example_withOptions demonstrates testing with rule options
func Example_withOptions() {
	var t *testing.T
	var testRule *rule.Rule // Replace with actual rule in real tests

	rule_tester.RunRuleTester(
		"/path/to/root",
		"tsconfig.json",
		t,
		testRule,
		[]rule_tester.ValidTestCase{
			{
				Code: "let x = 1;",
				Options: map[string]interface{}{
					"allowLet": true,
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "var x = 1;",
				Options: map[string]interface{}{
					"allowVar": false,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noVar"},
				},
			},
		},
	)
}

// Example_withSuggestions demonstrates testing suggestions
func Example_withSuggestions() {
	var t *testing.T
	var testRule *rule.Rule // Replace with actual rule in real tests

	rule_tester.RunRuleTester(
		"/path/to/root",
		"tsconfig.json",
		t,
		testRule,
		nil,
		[]rule_tester.InvalidTestCase{
			{
				Code: "const x: any = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedAny",
						Line:      1,
						Column:    10,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestUnknown",
								Output:    "const x: unknown = 1;",
							},
							{
								MessageId: "suggestNever",
								Output:    "const x: never = 1;",
							},
						},
					},
				},
			},
		},
	)
}

// Example_focusMode demonstrates using focus mode during development
func Example_focusMode() {
	var t *testing.T
	var testRule *rule.Rule // Replace with actual rule in real tests

	rule_tester.RunRuleTester(
		"/path/to/root",
		"tsconfig.json",
		t,
		testRule,
		[]rule_tester.ValidTestCase{
			{Code: "const x = 1;"},
			{Code: "const y = 2;", Only: true}, // Only this test runs
			{Code: "const z = 3;"},
		},
		nil,
	)
}

// Example_skipTest demonstrates skipping tests
func Example_skipTest() {
	var t *testing.T
	var testRule *rule.Rule // Replace with actual rule in real tests

	rule_tester.RunRuleTester(
		"/path/to/root",
		"tsconfig.json",
		t,
		testRule,
		[]rule_tester.ValidTestCase{
			{Code: "const x = 1;"},
			{Code: "// TODO: Fix this", Skip: true}, // This test is skipped
			{Code: "const z = 3;"},
		},
		nil,
	)
}

// Example_iterativeFixes demonstrates testing iterative fixes
func Example_iterativeFixes() {
	var t *testing.T
	var testRule *rule.Rule // Replace with actual rule in real tests

	rule_tester.RunRuleTester(
		"/path/to/root",
		"tsconfig.json",
		t,
		testRule,
		nil,
		[]rule_tester.InvalidTestCase{
			{
				Code: "a['b']['c'];",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useDot"},
				},
				Output: []string{
					"a.b['c'];", // After first fix
					"a.b.c;",    // After second fix (stable)
				},
			},
		},
	)
}

// Example_customFileName demonstrates using custom filenames
func Example_customFileName() {
	var t *testing.T
	var testRule *rule.Rule // Replace with actual rule in real tests

	rule_tester.RunRuleTester(
		"/path/to/root",
		"tsconfig.json",
		t,
		testRule,
		[]rule_tester.ValidTestCase{
			{
				Code:     "import { foo } from './bar';",
				FileName: "index.ts",
			},
			{
				Code: "export default function Component() { return <div />; }",
				Tsx:  true, // Uses .tsx extension
			},
		},
		nil,
	)
}

// Example_programHelper demonstrates creating TypeScript programs
func Example_programHelper() {
	helper := rule_tester.NewProgramHelper("/path/to/root")

	program, sourceFile, err := helper.CreateTestProgram(
		"const x: string = 'hello';",
		"test.ts",
		"tsconfig.json",
	)

	_ = program
	_ = sourceFile
	_ = err
}

// Example_diagnosticAssertion demonstrates creating diagnostic assertions
func Example_diagnosticAssertion() {
	assertion := rule_tester.NewDiagnosticAssertion()

	// Create a simple error
	error1 := assertion.FormatDiagnosticError("messageId", 1, 5, 1, 10)

	// Create an error with suggestions
	error2 := assertion.FormatDiagnosticErrorWithSuggestions(
		"messageId",
		1,
		5,
		[]rule_tester.InvalidTestCaseSuggestion{
			{MessageId: "suggestion1", Output: "fixed code 1"},
			{MessageId: "suggestion2", Output: "fixed code 2"},
		},
	)

	_ = error1
	_ = error2
}
