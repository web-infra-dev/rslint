# Rule Scaffolding Guide

This guide explains how to use the automated rule generation tools to quickly scaffold new ESLint and TypeScript-ESLint rule implementations in RSLint.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Tool Reference](#tool-reference)
  - [generate-rule.go](#generate-rulego)
  - [register-rule.go](#register-rulego)
- [Workflow Examples](#workflow-examples)
- [Rule Template Structure](#rule-template-structure)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Overview

RSLint provides automated tools to scaffold new rule implementations, reducing the manual boilerplate required to port 150+ rules from ESLint and TypeScript-ESLint. These tools:

- **Generate** complete rule implementation files with proper structure
- **Create** test file templates with placeholder test cases
- **Register** rules automatically in the rule registry
- **Support** batch processing for multiple rules
- **Fetch** metadata from ESLint/TypeScript-ESLint repositories (optional)

## Quick Start

### Generate a Single Rule

```bash
# Generate a TypeScript-ESLint rule
go run scripts/generate-rule.go \
  -rule no-explicit-any \
  -plugin typescript-eslint \
  -description "Disallow the any type" \
  -ast-nodes "TypeReference,IntersectionType" \
  -has-autofix

# Generate an import plugin rule
go run scripts/generate-rule.go \
  -rule no-unresolved \
  -plugin import \
  -description "Ensure imports point to files/modules that can be resolved" \
  -requires-types

# Generate a core ESLint rule
go run scripts/generate-rule.go \
  -rule prefer-const \
  -plugin "" \
  -description "Require const for variables never reassigned"
```

### Register the Rule

```bash
# Manually register a specific rule
go run scripts/register-rule.go \
  -rule no-explicit-any \
  -plugin typescript-eslint

# Auto-register all unregistered rules
go run scripts/register-rule.go -auto
```

### Test the Generated Rule

```bash
# Run tests for the new rule
go test ./internal/plugins/typescript/rules/no_explicit_any/

# Build to verify compilation
go build ./...
```

## Tool Reference

### generate-rule.go

Generates rule implementation and test files.

#### Flags

| Flag              | Type   | Default             | Description                                                            |
| ----------------- | ------ | ------------------- | ---------------------------------------------------------------------- |
| `-rule`           | string | _required_          | Rule name in kebab-case (e.g., `no-explicit-any`)                      |
| `-plugin`         | string | `typescript-eslint` | Plugin name: `typescript-eslint`, `import`, or empty for core rules    |
| `-description`    | string | ""                  | Brief description of the rule                                          |
| `-ast-nodes`      | string | ""                  | Comma-separated list of AST node types to listen to                    |
| `-requires-types` | bool   | `false`             | Whether the rule requires TypeScript type information                  |
| `-has-options`    | bool   | `false`             | Whether the rule has configuration options                             |
| `-has-autofix`    | bool   | `false`             | Whether the rule provides automatic fixes                              |
| `-batch`          | string | ""                  | Path to file containing rule names (one per line) for batch generation |
| `-dry-run`        | bool   | `false`             | Preview what would be generated without creating files                 |
| `-fetch`          | bool   | `false`             | Attempt to fetch rule metadata from ESLint/TypeScript-ESLint repos     |
| `-output`         | string | ""                  | Custom output directory (defaults to appropriate plugin directory)     |

#### Examples

**Basic Rule Generation:**

```bash
go run scripts/generate-rule.go -rule no-empty-interface -plugin typescript-eslint
```

**Rule with All Features:**

```bash
go run scripts/generate-rule.go \
  -rule no-explicit-any \
  -plugin typescript-eslint \
  -description "Disallow the any type" \
  -ast-nodes "TypeReference,IntersectionType,UnionType" \
  -requires-types \
  -has-options \
  -has-autofix
```

**Fetch Metadata from Source:**

```bash
go run scripts/generate-rule.go \
  -rule no-unused-vars \
  -plugin typescript-eslint \
  -fetch
```

**Dry Run (Preview):**

```bash
go run scripts/generate-rule.go \
  -rule no-explicit-any \
  -plugin typescript-eslint \
  -dry-run
```

**Batch Generation:**

Create a file `rules.txt`:

```
no-explicit-any
no-unsafe-assignment
no-unsafe-call
no-unsafe-member-access
no-unsafe-return
```

Then run:

```bash
go run scripts/generate-rule.go \
  -batch rules.txt \
  -plugin typescript-eslint \
  -fetch
```

#### Output Structure

For a TypeScript-ESLint rule named `no-explicit-any`:

```
internal/plugins/typescript/rules/no_explicit_any/
├── no_explicit_any.go       # Rule implementation
└── no_explicit_any_test.go  # Test file
```

### register-rule.go

Registers rules in the global rule registry.

#### Flags

| Flag       | Type   | Default                     | Description                                          |
| ---------- | ------ | --------------------------- | ---------------------------------------------------- |
| `-rule`    | string | ""                          | Rule name to register (e.g., `no-explicit-any`)      |
| `-plugin`  | string | `typescript-eslint`         | Plugin name: `typescript-eslint`, `import`, or empty |
| `-config`  | string | `internal/config/config.go` | Path to config.go file                               |
| `-dry-run` | bool   | `false`                     | Preview changes without modifying files              |
| `-auto`    | bool   | `false`                     | Auto-detect and register all unregistered rules      |

#### Examples

**Register a Single Rule:**

```bash
go run scripts/register-rule.go \
  -rule no-explicit-any \
  -plugin typescript-eslint
```

**Auto-Register All Rules:**

```bash
go run scripts/register-rule.go -auto
```

**Dry Run:**

```bash
go run scripts/register-rule.go \
  -rule no-explicit-any \
  -plugin typescript-eslint \
  -dry-run
```

## Workflow Examples

### Example 1: Adding a New TypeScript-ESLint Rule

```bash
# Step 1: Generate the rule
go run scripts/generate-rule.go \
  -rule no-explicit-any \
  -plugin typescript-eslint \
  -description "Disallow the any type" \
  -ast-nodes "TypeReference" \
  -has-options \
  -has-autofix \
  -fetch

# Step 2: Implement the rule logic
# Edit: internal/plugins/typescript/rules/no_explicit_any/no_explicit_any.go
# - Fill in the listener logic
# - Implement option parsing
# - Add autofix logic

# Step 3: Add test cases
# Edit: internal/plugins/typescript/rules/no_explicit_any/no_explicit_any_test.go
# - Add valid test cases
# - Add invalid test cases with expected errors
# - Add test cases for options
# - Add test cases for autofixes

# Step 4: Register the rule
go run scripts/register-rule.go \
  -rule no-explicit-any \
  -plugin typescript-eslint

# Step 5: Test and verify
go test ./internal/plugins/typescript/rules/no_explicit_any/
go build ./...

# Step 6: Run the linter
go run cmd/rslint/main.go --config rslint.json
```

### Example 2: Batch Migration of ESLint Rules

```bash
# Step 1: Create a list of rules to migrate
cat > eslint-rules.txt <<EOF
no-console
no-debugger
no-eval
no-implied-eval
no-new-func
prefer-const
prefer-template
EOF

# Step 2: Generate all rules
go run scripts/generate-rule.go \
  -batch eslint-rules.txt \
  -plugin "" \
  -fetch

# Step 3: Auto-register all generated rules
go run scripts/register-rule.go -auto

# Step 4: Verify compilation
go build ./...

# Step 5: Implement rules one by one
# - Edit each generated file
# - Add test cases
# - Run tests
```

### Example 3: Import Plugin Rule

```bash
# Generate import plugin rule
go run scripts/generate-rule.go \
  -rule no-unresolved \
  -plugin import \
  -description "Ensure imports point to resolvable files/modules" \
  -ast-nodes "ImportDeclaration,ExportDeclaration"

# Register the rule
go run scripts/register-rule.go \
  -rule no-unresolved \
  -plugin import

# Implement and test
# Edit: internal/plugins/import/rules/no_unresolved/no_unresolved.go
go test ./internal/plugins/import/rules/no_unresolved/
```

## Rule Template Structure

### Generated Rule Implementation

The tool generates a complete rule implementation with:

1. **Package declaration** using snake_case naming
2. **Options struct** (if `-has-options` is specified)
3. **Option parsing function** with dual-format support (array/object)
4. **Rule variable** using `rule.CreateRule()` for plugin rules
5. **Run function** with listener registration
6. **AST listeners** for specified node types
7. **TODO comments** for implementation guidance

### Generated Test File

The tool generates a test file with:

1. **Package declaration** matching the rule package
2. **Test imports** including appropriate fixtures
3. **Main test function** with `RunRuleTester()`
4. **Valid test cases** section (with placeholders)
5. **Invalid test cases** section (with placeholders)
6. **Optional options test** (if `-has-options` is specified)

## Best Practices

### 1. Start with Metadata Fetching

Use the `-fetch` flag to automatically retrieve rule descriptions and message IDs from the source repositories:

```bash
go run scripts/generate-rule.go -rule no-unused-vars -plugin typescript-eslint -fetch
```

### 2. Specify AST Nodes Explicitly

Provide the AST node types your rule will listen to:

```bash
go run scripts/generate-rule.go \
  -rule no-empty-interface \
  -plugin typescript-eslint \
  -ast-nodes "InterfaceDeclaration"
```

Common AST node types:

- `FunctionDeclaration`
- `ClassDeclaration`
- `InterfaceDeclaration`
- `TypeReference`
- `VariableDeclaration`
- `ImportDeclaration`
- `CallExpression`
- `PropertyAccessExpression`

See the [TypeScript AST documentation](https://github.com/microsoft/TypeScript/blob/main/src/compiler/types.ts) for a complete list.

### 3. Use Dry Run for Verification

Always preview the generated code before creating files:

```bash
go run scripts/generate-rule.go -rule my-rule -plugin typescript-eslint -dry-run
```

### 4. Batch Process Similar Rules

When porting multiple related rules, use batch processing:

```bash
# Create a list file
echo "no-explicit-any
no-unsafe-assignment
no-unsafe-call" > rules.txt

# Generate all at once
go run scripts/generate-rule.go -batch rules.txt -plugin typescript-eslint
```

### 5. Follow the Implementation Checklist

For each generated rule:

- [ ] **Implement** the core rule logic in listeners
- [ ] **Parse** options correctly (if applicable)
- [ ] **Add** comprehensive test cases
  - [ ] Valid cases
  - [ ] Invalid cases with error expectations
  - [ ] Edge cases
  - [ ] Options variations
- [ ] **Implement** autofixes (if applicable)
- [ ] **Test** autofixes with Output assertions
- [ ] **Register** the rule in config.go
- [ ] **Verify** compilation with `go build ./...`
- [ ] **Run** tests with `go test ./...`
- [ ] **Update** documentation if needed

### 6. Reference Existing Rules

Look at existing rule implementations for patterns:

```bash
# View a simple rule
cat internal/plugins/typescript/rules/no_empty_interface/no_empty_interface.go

# View a complex rule with options and autofixes
cat internal/plugins/typescript/rules/dot_notation/dot_notation.go
```

### 7. Use the Rule Tester Utilities

Leverage the testing infrastructure from PR #11:

```go
// Use batch test builder
builder := rule_tester.NewBatchTestBuilder()
builder.
    AddValid("const x = 1;").
    AddInvalid("var x = 1;", "useConst", 1, 1, "const x = 1;")
valid, invalid := builder.Build()

// Load tests from JSON
err := rule_tester.RunRuleTesterFromJSON(
    fixtures.GetRootDir(),
    "tsconfig.json",
    "testdata/my_rule_tests.json",
    t,
    &MyRule,
)
```

## Troubleshooting

### Issue: Generated code doesn't compile

**Solution:** Run `go fmt` on the generated files:

```bash
go fmt ./internal/plugins/typescript/rules/my_rule/
```

If errors persist, check:

- Import paths are correct
- Variable names match expectations
- AST node type names are valid

### Issue: Rule registration fails

**Solution:** Verify the rule variable name matches the expected pattern:

```go
// Correct pattern
var NoExplicitAnyRule = rule.CreateRule(...)

// Variable name must be: {PascalCaseRuleName}Rule
```

### Issue: Tests fail with "fixtures not found"

**Solution:** Ensure the fixtures import path is correct:

```go
// For TypeScript-ESLint rules
import "github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"

// For import plugin rules
import "github.com/web-infra-dev/rslint/internal/plugins/import/rules/fixtures"

// For core rules
import "github.com/web-infra-dev/rslint/internal/rules/fixtures"
```

### Issue: Auto-registration doesn't find rules

**Solution:** Ensure:

1. Rule directory contains a non-test `.go` file
2. Directory name matches snake_case pattern
3. Rule is in the expected plugin path

### Issue: Metadata fetch fails

**Solution:** Check:

1. Internet connection is available
2. Rule name matches the upstream repository
3. Rule exists in the specified plugin

For debugging, try:

```bash
# Manual URL check
curl https://raw.githubusercontent.com/typescript-eslint/typescript-eslint/main/packages/eslint-plugin/src/rules/no-explicit-any.ts
```

### Issue: Want to regenerate a rule

**Solution:** Delete the existing directory and regenerate:

```bash
rm -rf internal/plugins/typescript/rules/no_explicit_any
go run scripts/generate-rule.go -rule no-explicit-any -plugin typescript-eslint -fetch
```

## Advanced Usage

### Custom Output Directory

Generate rules in a custom location:

```bash
go run scripts/generate-rule.go \
  -rule my-custom-rule \
  -plugin typescript-eslint \
  -output ./custom-rules
```

### Generating Stub Rules for Planning

Use dry-run to generate a list of rules to implement:

```bash
# Create a comprehensive list
for rule in no-explicit-any no-unused-vars no-unsafe-assignment; do
    go run scripts/generate-rule.go \
        -rule $rule \
        -plugin typescript-eslint \
        -fetch \
        -dry-run
done > planned-rules.txt
```

### Integration with CI/CD

Add rule generation to your workflow:

```bash
#!/bin/bash
# generate-and-test.sh

RULE_NAME=$1
PLUGIN=${2:-typescript-eslint}

# Generate
go run scripts/generate-rule.go \
    -rule "$RULE_NAME" \
    -plugin "$PLUGIN" \
    -fetch

# Register
go run scripts/register-rule.go \
    -rule "$RULE_NAME" \
    -plugin "$PLUGIN"

# Verify
go build ./... && go test ./...
```

## Related Documentation

- [Architecture Guide](../architecture.md) - System architecture and design
- [Rule Testing Guide](./RULE_TESTING_GUIDE.md) - Comprehensive testing guide
- [Migration Tools Guide](../tools/README.md) - ESLint test conversion tools
- [ESLint Rule Documentation](https://eslint.org/docs/latest/extend/custom-rules)
- [TypeScript-ESLint Rules](https://typescript-eslint.io/rules/)

## Contributing

When contributing new rules:

1. Use the scaffolding tools to generate boilerplate
2. Follow the existing code patterns
3. Write comprehensive tests
4. Update documentation as needed
5. Ensure all tests pass before submitting PR

For questions or issues with the scaffolding tools, please [open an issue](https://github.com/web-infra-dev/rslint/issues).
