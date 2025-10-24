# RSLint Scripts

This directory contains automation scripts and tools for RSLint development.

## Available Scripts

### Rule Generation Tools

#### `generate-rule.go`

Automated tool to scaffold new ESLint and TypeScript-ESLint rule implementations.

**Quick Start:**

```bash
# Generate a TypeScript-ESLint rule
go run scripts/generate-rule.go -rule no-explicit-any -plugin typescript-eslint

# Generate with metadata fetching
go run scripts/generate-rule.go -rule no-unused-vars -plugin typescript-eslint -fetch

# Batch generate multiple rules
go run scripts/generate-rule.go -batch scripts/examples/typescript-eslint-rules.txt -plugin typescript-eslint
```

**Features:**

- Generates complete rule implementation files with proper structure
- Creates test file templates with placeholder test cases
- Supports metadata fetching from ESLint/TypeScript-ESLint repos
- Batch processing for multiple rules
- Dry-run mode for previewing changes
- Automatic code formatting

**Documentation:** See [Rule Scaffolding Guide](../docs/RULE_SCAFFOLDING_GUIDE.md)

#### `register-rule.go`

Automated tool to register rules in the global rule registry.

**Quick Start:**

```bash
# Register a specific rule
go run scripts/register-rule.go -rule no-explicit-any -plugin typescript-eslint

# Auto-register all unregistered rules
go run scripts/register-rule.go -auto
```

**Features:**

- Adds import statements to config.go
- Adds registration calls in alphabetical order
- Auto-detection of unregistered rules
- Dry-run mode for previewing changes
- Maintains proper code formatting

**Documentation:** See [Rule Scaffolding Guide](../docs/RULE_SCAFFOLDING_GUIDE.md)

## Example Batch Files

The `scripts/examples/` directory contains example batch files for common rule generation tasks:

### `typescript-eslint-rules.txt`

List of common TypeScript-ESLint rules to implement:

```bash
go run scripts/generate-rule.go -batch scripts/examples/typescript-eslint-rules.txt -plugin typescript-eslint -fetch
```

### `eslint-core-rules.txt`

List of core ESLint rules to implement:

```bash
go run scripts/generate-rule.go -batch scripts/examples/eslint-core-rules.txt -plugin "" -fetch
```

### `import-plugin-rules.txt`

List of eslint-plugin-import rules to implement:

```bash
go run scripts/generate-rule.go -batch scripts/examples/import-plugin-rules.txt -plugin import -fetch
```

## Common Workflows

### Add a Single Rule

```bash
# 1. Generate the rule
go run scripts/generate-rule.go \
  -rule no-explicit-any \
  -plugin typescript-eslint \
  -description "Disallow the any type" \
  -ast-nodes "TypeReference" \
  -has-autofix \
  -fetch

# 2. Implement the rule logic
# Edit: internal/plugins/typescript/rules/no_explicit_any/no_explicit_any.go

# 3. Add test cases
# Edit: internal/plugins/typescript/rules/no_explicit_any/no_explicit_any_test.go

# 4. Register the rule
go run scripts/register-rule.go -rule no-explicit-any -plugin typescript-eslint

# 5. Test and verify
go test ./internal/plugins/typescript/rules/no_explicit_any/
go build ./...
```

### Batch Add Multiple Rules

```bash
# 1. Create a rule list file
cat > my-rules.txt <<EOF
no-explicit-any
no-unsafe-assignment
no-unsafe-call
EOF

# 2. Generate all rules
go run scripts/generate-rule.go -batch my-rules.txt -plugin typescript-eslint -fetch

# 3. Auto-register all rules
go run scripts/register-rule.go -auto

# 4. Verify compilation
go build ./...

# 5. Implement each rule individually
```

### Preview Before Generating

```bash
# Dry-run to preview
go run scripts/generate-rule.go -rule no-explicit-any -plugin typescript-eslint -dry-run

# If it looks good, generate for real
go run scripts/generate-rule.go -rule no-explicit-any -plugin typescript-eslint
```

## Development Tools

### Utility Scripts

Additional utility scripts for development tasks:

- **Planned:** `scripts/list-unimplemented-rules.go` - List ESLint rules not yet implemented
- **Planned:** `scripts/validate-rules.go` - Validate all rule implementations
- **Planned:** `scripts/benchmark-rules.go` - Benchmark rule performance

## Tips and Best Practices

### 1. Use Metadata Fetching

Always use `-fetch` when generating rules to automatically retrieve descriptions and metadata:

```bash
go run scripts/generate-rule.go -rule no-unused-vars -plugin typescript-eslint -fetch
```

### 2. Specify AST Nodes

Explicitly specify which AST nodes your rule listens to:

```bash
go run scripts/generate-rule.go \
  -rule no-empty-interface \
  -plugin typescript-eslint \
  -ast-nodes "InterfaceDeclaration"
```

### 3. Auto-Register After Batch Generation

After generating multiple rules, use auto-registration:

```bash
go run scripts/generate-rule.go -batch rules.txt -plugin typescript-eslint
go run scripts/register-rule.go -auto
```

### 4. Check Compilation Frequently

Verify everything compiles after generating/registering rules:

```bash
go build ./...
```

### 5. Reference Existing Rules

Look at existing implementations for patterns:

```bash
# Simple rule
cat internal/plugins/typescript/rules/no_empty_interface/no_empty_interface.go

# Complex rule with autofix
cat internal/plugins/typescript/rules/dot_notation/dot_notation.go
```

## Integration with Testing Tools

These scaffolding tools work seamlessly with the testing infrastructure from PR #11:

```go
// In your generated test file, you can use:

// 1. Batch test builder
builder := rule_tester.NewBatchTestBuilder()
builder.AddValid("const x = 1;")
       .AddInvalid("var x = 1;", "useConst", 1, 1, "const x = 1;")

// 2. Load tests from JSON
rule_tester.RunRuleTesterFromJSON(
    fixtures.GetRootDir(),
    "tsconfig.json",
    "testdata/my_rule_tests.json",
    t,
    &MyRule,
)

// 3. Convert ESLint tests
go run tools/typescript_eslint_test_converter.go \
  -input testdata/eslint/no-var.json \
  -output internal/plugins/typescript/rules/no_var/tests.json
```

See [docs/RULE_TESTING_GUIDE.md](../docs/RULE_TESTING_GUIDE.md) for comprehensive testing documentation.

## Troubleshooting

### Generated code doesn't compile

Run `go fmt` on the generated files:

```bash
go fmt ./internal/plugins/typescript/rules/my_rule/
```

### Registration fails

Verify the rule variable name matches the pattern `{PascalCaseRuleName}Rule`.

### Auto-registration doesn't find rules

Ensure:

1. Rule directory contains a non-test `.go` file
2. Directory name is in snake_case
3. Rule is in the expected plugin path

### Metadata fetch fails

Check:

1. Internet connection is available
2. Rule name matches the upstream repository
3. Rule exists in the specified plugin

## Related Documentation

- [Rule Scaffolding Guide](../docs/RULE_SCAFFOLDING_GUIDE.md) - Comprehensive guide for using these tools
- [Rule Testing Guide](../docs/RULE_TESTING_GUIDE.md) - Testing infrastructure and best practices
- [Architecture Guide](../architecture.md) - System architecture and rule framework
- [Migration Tools](../tools/README.md) - ESLint test conversion tools

## Contributing

When adding new scripts:

1. Follow Go naming conventions
2. Add comprehensive flag documentation
3. Include error handling and validation
4. Update this README
5. Add examples to the examples/ directory
6. Test with various scenarios

For questions or suggestions, please [open an issue](https://github.com/web-infra-dev/rslint/issues).
