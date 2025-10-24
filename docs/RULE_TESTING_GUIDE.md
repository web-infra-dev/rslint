# Rule Testing Guide

This guide explains how to write comprehensive tests for RSLint rules to support porting 150+ ESLint and TypeScript-ESLint rules.

## Table of Contents

1. [Overview](#overview)
2. [Quick Start](#quick-start)
3. [Test Structure](#test-structure)
4. [Writing Test Cases](#writing-test-cases)
5. [Using Test Utilities](#using-test-utilities)
6. [Loading Tests from JSON](#loading-tests-from-json)
7. [Converting ESLint Tests](#converting-eslint-tests)
8. [Best Practices](#best-practices)
9. [Advanced Features](#advanced-features)

## Overview

RSLint uses a centralized testing framework located in `internal/rule_tester/` that provides:

- **Parallel test execution** for fast test runs
- **ESLint-compatible test format** for easy porting
- **Automatic fix validation** with iterative application
- **Suggestion testing** for alternative fixes
- **JSON batch test loading** for large test suites
- **Focus and skip modes** for development

## Quick Start

### Basic Test File

```go
package my_rule

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestMyRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&MyRule,
		[]rule_tester.ValidTestCase{
			{Code: "const x = 1;"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "var x = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useConst",
						Line:      1,
						Column:    1,
					},
				},
				Output: []string{"const x = 1;"},
			},
		},
	)
}
```

## Test Structure

### ValidTestCase

Valid test cases should NOT produce any diagnostics.

```go
type ValidTestCase struct {
	Code     string  // Code to test (required)
	FileName string  // Custom filename (default: "file.ts")
	Only     bool    // Run only this test (focus mode)
	Skip     bool    // Skip this test
	Options  any     // Rule configuration options
	TSConfig string  // Custom tsconfig.json path
	Tsx      bool    // Use .tsx extension
}
```

### InvalidTestCase

Invalid test cases MUST produce diagnostics and optionally fixes.

```go
type InvalidTestCase struct {
	Code     string                    // Code to test (required)
	FileName string                    // Custom filename
	Only     bool                      // Focus mode
	Skip     bool                      // Skip test
	Output   []string                  // Expected code after fixes
	Errors   []InvalidTestCaseError    // Expected diagnostics (required)
	TSConfig string                    // Custom tsconfig.json
	Options  any                       // Rule options
	Tsx      bool                      // Use .tsx extension
}

type InvalidTestCaseError struct {
	MessageId   string                         // Error message ID (required)
	Line        int                            // 1-indexed line number
	Column      int                            // 1-indexed column number
	EndLine     int                            // End line (optional)
	EndColumn   int                            // End column (optional)
	Suggestions []InvalidTestCaseSuggestion    // Alternative suggestions
}
```

## Writing Test Cases

### Valid Cases

```go
[]rule_tester.ValidTestCase{
	// Simple valid code
	{Code: "const x = 1;"},

	// With options
	{
		Code: "let x = 1;",
		Options: map[string]interface{}{"allowLet": true},
	},

	// Custom filename
	{
		Code: "export default function() {}",
		FileName: "index.tsx",
		Tsx: true,
	},

	// Custom TypeScript config
	{
		Code: "const x: any = 1;",
		TSConfig: "tsconfig.strict.json",
	},
}
```

### Invalid Cases

```go
[]rule_tester.InvalidTestCase{
	// Simple error
	{
		Code: "var x = 1;",
		Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "useConst"},
		},
	},

	// With fix
	{
		Code: "var x = 1;",
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "useConst",
				Line:      1,
				Column:    1,
			},
		},
		Output: []string{"const x = 1;"},
	},

	// Multiple errors
	{
		Code: "var x = 1; var y = 2;",
		Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "useConst", Line: 1, Column: 1},
			{MessageId: "useConst", Line: 1, Column: 12},
		},
	},

	// With suggestions
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

	// Iterative fixes
	{
		Code: "a['b']['c'];",
		Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "useDot"},
		},
		Output: []string{
			"a.b['c'];",  // After first fix
			"a.b.c;",     // After second fix (stable)
		},
	},
}
```

## Using Test Utilities

### Common Fixtures

The `CommonFixtures` helper generates common TypeScript patterns:

```go
fixtures := rule_tester.NewCommonFixtures()

// Generate a class
classCode := fixtures.Class("MyClass", fixtures.Method("method", "", "void", "return;"))

// Generate an interface
interfaceCode := fixtures.Interface("MyInterface", fixtures.Property("prop", "string", ""))

// Generate a function
funcCode := fixtures.Function("myFunc", "x: number", "string", "return x.toString();")

// Use in tests
rule_tester.RunRuleTester(/* ... */, []rule_tester.ValidTestCase{
	{Code: classCode},
	{Code: interfaceCode},
	{Code: funcCode},
}, /* ... */)
```

### Batch Test Builder

Build large test suites programmatically:

```go
builder := rule_tester.NewBatchTestBuilder()

builder.
	AddValid("const x = 1;").
	AddValid("const y = 2;").
	AddInvalid("var x = 1;", "useConst", 1, 1, "const x = 1;").
	AddInvalid("var y = 2;", "useConst", 1, 1, "const y = 2;")

valid, invalid := builder.Build()

rule_tester.RunRuleTester(/* ... */, valid, invalid)
```

### Program Helper

Create TypeScript programs for advanced testing:

```go
helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())

program, sourceFile, err := helper.CreateTestProgram(
	"const x: string = 'hello';",
	"test.ts",
	"tsconfig.json",
)
// Use program and sourceFile for manual testing
```

## Loading Tests from JSON

### Native Format

Create a JSON file with test cases:

```json
{
  "valid": [
    { "code": "const x = 1;" },
    { "code": "let y = 2;", "options": { "allowLet": true } }
  ],
  "invalid": [
    {
      "code": "var x = 1;",
      "errors": [{ "messageId": "useConst", "line": 1, "column": 1 }],
      "output": ["const x = 1;"]
    }
  ]
}
```

Load and run:

```go
func TestMyRuleFromJSON(t *testing.T) {
	err := rule_tester.RunRuleTesterFromJSON(
		fixtures.GetRootDir(),
		"tsconfig.json",
		"testdata/my_rule_tests.json",
		t,
		&MyRule,
	)
	assert.NilError(t, err)
}
```

### ESLint Format

Load ESLint-compatible JSON:

```go
func TestMyRuleFromESLint(t *testing.T) {
	err := rule_tester.RunRuleTesterFromESLintJSON(
		fixtures.GetRootDir(),
		"tsconfig.json",
		"testdata/eslint_tests.json",
		t,
		&MyRule,
	)
	assert.NilError(t, err)
}
```

## Converting ESLint Tests

### Using the CLI Tool

Convert ESLint test JSON to RSLint format:

```bash
# Convert to JSON
go run tools/eslint_test_converter.go \
  -input testdata/eslint/no-var.json \
  -output testdata/rslint/no-var.json \
  -verbose

# Convert TypeScript-ESLint tests to JSON
go run tools/typescript_eslint_test_converter.go \
  -input testdata/ts-eslint/no-explicit-any.json \
  -output testdata/rslint/no-explicit-any.json

# Generate Go test file directly
go run tools/typescript_eslint_test_converter.go \
  -input testdata/ts-eslint/no-explicit-any.json \
  -output internal/plugins/typescript/rules/no_explicit_any/no_explicit_any_test.go \
  -go \
  -rule no_explicit_any
```

### Programmatic Conversion

```go
// Load ESLint test suite
eslintSuite, err := rule_tester.LoadESLintTestSuiteFromJSON("eslint_tests.json")
if err != nil {
	// handle error
}

// Convert to RSLint format
suite := rule_tester.ConvertESLintTestSuite(eslintSuite)

// Use converted tests
rule_tester.RunRuleTester(/* ... */, suite.Valid, suite.Invalid)
```

## Best Practices

### 1. Test Organization

```go
func TestMyRule(t *testing.T) {
	// Group related test cases
	validCases := []rule_tester.ValidTestCase{
		// Basic valid cases
		{Code: "const x = 1;"},
		{Code: "let y = 2;"},

		// Edge cases
		{Code: "for (let i = 0; i < 10; i++) {}"},
		{Code: "const { x, y } = obj;"},
	}

	invalidCases := []rule_tester.InvalidTestCase{
		// Basic invalid cases
		{Code: "var x = 1;", /* ... */},

		// Complex scenarios
		{Code: "var { x, y } = obj;", /* ... */},
	}

	rule_tester.RunRuleTester(/* ... */, validCases, invalidCases)
}
```

### 2. Use Focus Mode During Development

```go
{
	Code: "var x = 1;",
	Only: true,  // Run only this test
	Errors: []rule_tester.InvalidTestCaseError{
		{MessageId: "useConst"},
	},
}
```

### 3. Test All Message IDs

Ensure every message ID in your rule is tested:

```go
// In your rule
messages := map[string]string{
	"useConst": "Use 'const' instead of 'var'",
	"useLet":   "Use 'let' instead of 'var'",
}

// In your tests - test both message IDs
invalidCases := []rule_tester.InvalidTestCase{
	{
		Code: "var x = 1;",  // Never reassigned
		Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "useConst"},
		},
	},
	{
		Code: "var y = 1; y = 2;",  // Reassigned
		Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "useLet"},
		},
	},
}
```

### 4. Test Fixes Thoroughly

```go
// Test that fixes are correct
{
	Code: "var x = 1;",
	Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useConst"}},
	Output: []string{"const x = 1;"},  // Verify exact output
}

// Test iterative fixes
{
	Code: "var x = 1; var y = 2;",
	Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useConst"}},
	Output: []string{
		"const x = 1; var y = 2;",      // After first fix
		"const x = 1; const y = 2;",    // After second fix
	},
}
```

### 5. Test Edge Cases

```go
validCases := []rule_tester.ValidTestCase{
	// Empty file
	{Code: ""},

	// Comments only
	{Code: "// comment"},

	// TypeScript-specific syntax
	{Code: "const x: number = 1;"},
	{Code: "interface Foo { bar: string; }"},

	// JSX/TSX
	{Code: "const x = <div />;", Tsx: true},
}
```

### 6. Use Descriptive Test Names

The test runner automatically generates names like `valid-0`, `invalid-1`, but you can use subtests for better organization:

```go
func TestMyRule(t *testing.T) {
	t.Run("valid cases", func(t *testing.T) {
		rule_tester.RunRuleTester(/* ... */, validCases, nil)
	})

	t.Run("invalid cases", func(t *testing.T) {
		rule_tester.RunRuleTester(/* ... */, nil, invalidCases)
	})
}
```

## Advanced Features

### Custom TypeScript Configuration

```go
{
	Code: "const x: any = 1;",
	TSConfig: "tsconfig.strict.json",  // Use different tsconfig
}
```

### Multiple Suggestions

```go
{
	Code: "const x: any = 1;",
	Errors: []rule_tester.InvalidTestCaseError{
		{
			MessageId: "unexpectedAny",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestUnknown", Output: "const x: unknown = 1;"},
				{MessageId: "suggestNever", Output: "const x: never = 1;"},
			},
		},
	},
}
```

### Testing Options

```go
{
	Code: "var x = 1;",
	Options: map[string]interface{}{
		"allowVar": false,
		"severity": "error",
	},
	Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noVar"}},
}

// Or as array (ESLint compatibility)
{
	Code: "var x = 1;",
	Options: []interface{}{"never", map[string]interface{}{"strict": true}},
	Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noVar"}},
}
```

### Skip Tests

```go
{
	Code: "// TODO: Fix this test",
	Skip: true,  // Skip this test
}
```

## Testing Checklist

When writing tests for a new rule:

- [ ] Test all valid code patterns the rule should allow
- [ ] Test all invalid code patterns the rule should catch
- [ ] Test all message IDs defined in the rule
- [ ] Test all auto-fixes produce correct code
- [ ] Test all suggestions are correct
- [ ] Test with and without options (if rule supports options)
- [ ] Test TypeScript-specific syntax (types, interfaces, etc.)
- [ ] Test JSX/TSX syntax (if applicable)
- [ ] Test edge cases (empty files, comments, complex nesting)
- [ ] Test iterative fixes reach stable state
- [ ] Verify position information (line, column) is correct
- [ ] Test with different TypeScript compiler options (if relevant)

## Example: Complete Test File

```go
package no_var

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoVarRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoVarRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			{Code: "const x = 1;"},
			{Code: "let y = 2;"},
			{Code: "const { a, b } = obj;"},
			{Code: "for (let i = 0; i < 10; i++) {}"},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			{
				Code: "var x = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useConst",
						Line:      1,
						Column:    1,
					},
				},
				Output: []string{"const x = 1;"},
			},
			{
				Code: "var y = 1; y = 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useLet",
						Line:      1,
						Column:    1,
					},
				},
				Output: []string{"let y = 1; y = 2;"},
			},
			{
				Code: "for (var i = 0; i < 10; i++) {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useLet",
						Line:      1,
						Column:    6,
					},
				},
				Output: []string{"for (let i = 0; i < 10; i++) {}"},
			},
		},
	)
}
```

## Further Reading

- [RSLint Architecture](https://github.com/web-infra-dev/rslint/blob/main/architecture.md)
- [ESLint Rule Testing](https://eslint.org/docs/latest/integrate/nodejs-api#ruletester)
- [TypeScript-ESLint Testing](https://typescript-eslint.io/developers/custom-rules#testing)
- [Go Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)
