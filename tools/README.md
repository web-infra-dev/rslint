# RSLint Testing Tools

This directory contains tools for converting and managing test cases when porting ESLint and TypeScript-ESLint rules to RSLint.

## Available Tools

### 1. ESLint Test Converter

Converts ESLint test cases from JSON format to RSLint format.

**Usage:**

```bash
# Convert ESLint tests to RSLint JSON format
go run tools/eslint_test_converter.go \
  -input testdata/eslint/no-var.json \
  -output testdata/rslint/no-var.json \
  -verbose

# Auto-generate output filename
go run tools/eslint_test_converter.go \
  -input testdata/eslint/no-var.json \
  -verbose
# Output will be: testdata/eslint/no-var_rslint.json
```

**Flags:**

- `-input`: Input ESLint test file (JSON) - **required**
- `-output`: Output RSLint test file (JSON) - optional, auto-generated if not specified
- `-verbose`: Enable verbose output

**Example:**

Input ESLint format (`eslint_tests.json`):

```json
{
  "valid": [
    { "code": "const x = 1;" },
    { "code": "let y = 2;", "options": [{ "allowLet": true }] }
  ],
  "invalid": [
    {
      "code": "var x = 1;",
      "errors": [{ "messageId": "useConst", "line": 1, "column": 1 }],
      "output": "const x = 1;"
    }
  ]
}
```

Output RSLint format (`rslint_tests.json`):

```json
{
  "valid": [
    { "Code": "const x = 1;" },
    { "Code": "let y = 2;", "Options": { "allowLet": true } }
  ],
  "invalid": [
    {
      "Code": "var x = 1;",
      "Errors": [{ "MessageId": "useConst", "Line": 1, "Column": 1 }],
      "Output": ["const x = 1;"]
    }
  ]
}
```

### 2. TypeScript-ESLint Test Converter

Converts TypeScript-ESLint test cases to RSLint format. Can generate either JSON or Go test files.

**Usage:**

```bash
# Convert to JSON format
go run tools/typescript_eslint_test_converter.go \
  -input testdata/ts-eslint/no-explicit-any.json \
  -output testdata/rslint/no-explicit-any.json \
  -verbose

# Generate Go test file directly
go run tools/typescript_eslint_test_converter.go \
  -input testdata/ts-eslint/no-explicit-any.json \
  -output internal/plugins/typescript/rules/no_explicit_any/no_explicit_any_test.go \
  -go \
  -rule no_explicit_any \
  -verbose
```

**Flags:**

- `-input`: Input TypeScript-ESLint test file (JSON) - **required**
- `-output`: Output file (JSON or Go) - optional, auto-generated if not specified
- `-go`: Generate Go test file instead of JSON
- `-rule`: Rule name for Go test generation - **required when -go is used**
- `-verbose`: Enable verbose output

**Example Go Output:**

```go
package no_explicit_any

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoExplicitAnyRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoExplicitAnyRule,
		[]rule_tester.ValidTestCase{
			{Code: "const x: number = 1;"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "const x: any = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedAny",
						Line: 1,
						Column: 10,
					},
				},
				Output: []string{"const x: unknown = 1;"},
			},
		},
	)
}
```

## Building the Tools

To build the tools as standalone binaries:

```bash
# Build ESLint converter
go build -o bin/eslint-test-converter tools/eslint_test_converter.go

# Build TypeScript-ESLint converter
go build -o bin/ts-eslint-test-converter tools/typescript_eslint_test_converter.go

# Use the binaries
./bin/eslint-test-converter -input tests.json -output rslint_tests.json
./bin/ts-eslint-test-converter -input tests.json -go -rule my_rule
```

## Workflow for Porting Rules

### From ESLint

1. **Extract tests from ESLint repository**

   - Locate the rule in the ESLint repository
   - Export the test cases to JSON format

2. **Convert tests**

   ```bash
   go run tools/eslint_test_converter.go \
     -input eslint_tests.json \
     -output rslint_tests.json
   ```

3. **Load tests in your rule test file**
   ```go
   func TestMyRule(t *testing.T) {
       err := rule_tester.RunRuleTesterFromJSON(
           fixtures.GetRootDir(),
           "tsconfig.json",
           "testdata/rslint_tests.json",
           t,
           &MyRule,
       )
       assert.NilError(t, err)
   }
   ```

### From TypeScript-ESLint

1. **Extract tests from TypeScript-ESLint repository**

   - Locate the rule in the TypeScript-ESLint repository
   - Export the test cases to JSON format

2. **Generate Go test file directly**

   ```bash
   go run tools/typescript_eslint_test_converter.go \
     -input ts_eslint_tests.json \
     -output internal/plugins/typescript/rules/my_rule/my_rule_test.go \
     -go \
     -rule my_rule
   ```

3. **Review and adjust the generated test file**
   - Add any missing test cases
   - Adjust options if needed
   - Run tests: `go test ./internal/plugins/typescript/rules/my_rule/`

## Test Format Reference

### ESLint/TypeScript-ESLint Format

```json
{
  "valid": [
    {
      "code": "const x = 1;",
      "filename": "test.ts",
      "options": [{ "strict": true }],
      "parser": "@typescript-eslint/parser"
    }
  ],
  "invalid": [
    {
      "code": "var x = 1;",
      "filename": "test.ts",
      "options": [{ "strict": true }],
      "errors": [
        {
          "messageId": "useConst",
          "line": 1,
          "column": 1,
          "endLine": 1,
          "endColumn": 4,
          "suggestions": [
            {
              "messageId": "replaceWithConst",
              "output": "const x = 1;"
            }
          ]
        }
      ],
      "output": "const x = 1;"
    }
  ]
}
```

### RSLint Format

```json
{
  "Valid": [
    {
      "Code": "const x = 1;",
      "FileName": "test.ts",
      "Options": { "strict": true }
    }
  ],
  "Invalid": [
    {
      "Code": "var x = 1;",
      "FileName": "test.ts",
      "Options": { "strict": true },
      "Errors": [
        {
          "MessageId": "useConst",
          "Line": 1,
          "Column": 1,
          "EndLine": 1,
          "EndColumn": 4,
          "Suggestions": [
            {
              "MessageId": "replaceWithConst",
              "Output": "const x = 1;"
            }
          ]
        }
      ],
      "Output": ["const x = 1;"]
    }
  ]
}
```

## Tips and Best Practices

### 1. Batch Conversion

Convert multiple test files at once using a shell script:

```bash
#!/bin/bash
for file in testdata/eslint/*.json; do
  basename=$(basename "$file" .json)
  go run tools/eslint_test_converter.go \
    -input "$file" \
    -output "testdata/rslint/${basename}_rslint.json"
done
```

### 2. Validation

After conversion, always verify:

- Test cases are complete
- Options are correctly converted
- File paths are correct
- Message IDs match your rule implementation

### 3. Manual Adjustments

Some test cases may require manual adjustment after conversion:

- Complex option structures
- Custom parsers or plugins
- Environment-specific configurations
- TypeScript-specific compiler options

### 4. Incremental Migration

When porting a large rule:

1. Convert a small subset of tests first
2. Implement the rule to pass those tests
3. Convert and add more tests incrementally
4. This helps catch issues early

## Contributing

When adding new conversion tools:

1. Follow the existing pattern in `eslint_test_converter.go`
2. Add comprehensive error handling
3. Include `-verbose` flag for debugging
4. Update this README with usage examples
5. Add tests for the conversion logic

## Troubleshooting

### Issue: Conversion fails with JSON parse error

**Solution:** Ensure the input JSON is valid. Use a JSON validator or:

```bash
cat input.json | jq .
```

### Issue: Generated Go code doesn't compile

**Solution:**

- Check that the rule name is correct
- Verify the import paths are correct
- Ensure the rule variable name follows Go naming conventions

### Issue: Test cases have incorrect positions

**Solution:** ESLint uses 1-indexed line/column numbers. The converter handles this automatically, but verify that your rule implementation also uses 1-indexed positions.

### Issue: Options aren't converted correctly

**Solution:** ESLint options can be arrays or objects. Check the original test format and manually adjust if needed:

```go
// If options were an array in ESLint
Options: []interface{}{"option1", "option2"}

// If options were an object
Options: map[string]interface{}{"key": "value"}
```

## Further Reading

- [Rule Testing Guide](../docs/RULE_TESTING_GUIDE.md)
- [ESLint Rule Testing Documentation](https://eslint.org/docs/latest/integrate/nodejs-api#ruletester)
- [TypeScript-ESLint Custom Rules](https://typescript-eslint.io/developers/custom-rules)
