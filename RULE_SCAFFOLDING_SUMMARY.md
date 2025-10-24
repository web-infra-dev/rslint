# Rule Scaffolding Tools - Implementation Summary

## Overview

This PR introduces automated code generation tools to significantly accelerate the process of porting 150+ ESLint and TypeScript-ESLint rules to RSLint. These tools reduce manual boilerplate and enable rapid rule implementation.

## Key Components

### 1. `scripts/generate-rule.go` - Rule Code Generator

**Purpose:** Automatically generates rule implementation and test file boilerplate.

**Features:**

- ✅ Generates complete Go rule implementation with proper package structure
- ✅ Creates test file templates with placeholder test cases
- ✅ Supports metadata fetching from ESLint/TypeScript-ESLint GitHub repositories
- ✅ Batch processing for generating multiple rules at once
- ✅ Dry-run mode for previewing generated code
- ✅ Automatic code formatting with `go/format`
- ✅ Configurable options for AST nodes, autofixes, and rule options
- ✅ Plugin support (TypeScript-ESLint, Import, Core ESLint)

**Usage Example:**

```bash
# Generate a TypeScript-ESLint rule with autofix
go run scripts/generate-rule.go \
  -rule no-explicit-any \
  -plugin typescript-eslint \
  -description "Disallow the any type" \
  -ast-nodes "TypeReference,IntersectionType" \
  -has-autofix \
  -fetch

# Batch generate 10+ rules from a file
go run scripts/generate-rule.go \
  -batch scripts/examples/typescript-eslint-rules.txt \
  -plugin typescript-eslint \
  -fetch
```

### 2. `scripts/register-rule.go` - Rule Registry Manager

**Purpose:** Automatically registers generated rules in the global rule registry.

**Features:**

- ✅ Adds import statements to `internal/config/config.go`
- ✅ Adds registration calls in alphabetical order
- ✅ Auto-detects unregistered rules across all plugins
- ✅ Maintains proper code formatting
- ✅ Dry-run mode for previewing changes
- ✅ Idempotent (skips already registered rules)

**Usage Example:**

```bash
# Register a specific rule
go run scripts/register-rule.go \
  -rule no-explicit-any \
  -plugin typescript-eslint

# Auto-register all unregistered rules
go run scripts/register-rule.go -auto
```

### 3. Documentation

**`docs/RULE_SCAFFOLDING_GUIDE.md`** - Comprehensive guide covering:

- Tool reference and flag documentation
- Step-by-step workflow examples
- Rule template structure explanation
- Best practices for rule implementation
- Troubleshooting common issues
- Integration with testing tools from PR #11

**`scripts/README.md`** - Scripts directory documentation:

- Overview of all available scripts
- Common workflows and patterns
- Example batch files
- Tips and best practices
- Integration with testing infrastructure

### 4. Example Batch Files

Pre-created lists of common rules for batch generation:

- **`scripts/examples/typescript-eslint-rules.txt`** - 20+ popular TypeScript-ESLint rules
- **`scripts/examples/eslint-core-rules.txt`** - 15+ core ESLint rules
- **`scripts/examples/import-plugin-rules.txt`** - 15+ import plugin rules

## Generated Code Structure

### Rule Implementation Template

For a rule named `no-explicit-any`, the tool generates:

```
internal/plugins/typescript/rules/no_explicit_any/
├── no_explicit_any.go       # Complete rule implementation
└── no_explicit_any_test.go  # Test template with placeholders
```

**Generated rule file includes:**

- Package declaration with correct naming convention
- Import statements for required dependencies
- Options struct (if requested)
- Option parsing function with dual-format support (array/object)
- Rule variable using `rule.CreateRule()` for plugins
- Run function with proper signature
- AST listeners for specified node types
- TODO comments for implementation guidance
- Example error reporting code
- Autofix scaffolding (if requested)

**Generated test file includes:**

- Package declaration matching the rule
- Appropriate fixtures import
- Main test function with `RunRuleTester()`
- Valid test case section with placeholders
- Invalid test case section with error expectations
- Options test (if rule has options)
- Autofix output assertions (if rule has autofixes)

## Technical Highlights

### Code Generation Patterns

1. **Snake Case Conversion**: `no-explicit-any` → `no_explicit_any` (package names)
2. **Pascal Case Conversion**: `no-explicit-any` → `NoExplicitAny` (type names)
3. **Plugin Prefixing**: Automatic addition of `@typescript-eslint/`, `import/`, etc.
4. **Template System**: Uses Go's `text/template` for code generation
5. **Formatting**: Automatically runs `go/format.Source()` on generated code

### Metadata Fetching

When using `-fetch`, the tool attempts to retrieve:

- Rule description from upstream repositories
- Message IDs from the rule's meta object
- Category information
- Type checking requirements

Supports fetching from:

- TypeScript-ESLint: `github.com/typescript-eslint/typescript-eslint`
- ESLint Core: `github.com/eslint/eslint`
- Import Plugin: `github.com/import-js/eslint-plugin-import`

### Rule Registry Management

The `register-rule.go` tool:

1. Parses `internal/config/config.go`
2. Finds the appropriate registration function
3. Inserts import in alphabetical order among internal imports
4. Inserts registration call in alphabetical order by rule name
5. Formats the entire file with `go/format`
6. Safely handles already-registered rules

## Integration with Existing Tools

These scaffolding tools are designed to work seamlessly with the testing infrastructure from **PR #11**:

### Testing Tools Compatibility

```go
// Generated tests can be enhanced with PR #11 utilities:

// 1. Batch Test Builder
builder := rule_tester.NewBatchTestBuilder()
builder.
    AddValid(fixtures.Const("x", "number", "1")).
    AddInvalid("var x = 1;", "useConst", 1, 1, "const x = 1;")

// 2. Load from JSON
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

## Workflow: From Zero to Working Rule

**Complete workflow for implementing a new rule:**

```bash
# 1. Generate the rule boilerplate
go run scripts/generate-rule.go \
  -rule no-explicit-any \
  -plugin typescript-eslint \
  -description "Disallow the any type" \
  -ast-nodes "TypeReference" \
  -has-options \
  -has-autofix \
  -fetch

# 2. Implement the rule logic
# Edit: internal/plugins/typescript/rules/no_explicit_any/no_explicit_any.go
# - Fill in AST listener logic
# - Parse options
# - Add error reporting
# - Implement autofixes

# 3. Add comprehensive test cases
# Edit: internal/plugins/typescript/rules/no_explicit_any/no_explicit_any_test.go
# - Valid test cases
# - Invalid test cases with errors
# - Options test cases
# - Autofix test cases

# 4. Register the rule
go run scripts/register-rule.go \
  -rule no-explicit-any \
  -plugin typescript-eslint

# 5. Test
go test ./internal/plugins/typescript/rules/no_explicit_any/
go build ./...

# 6. Use the rule
# Add to rslint.json:
# "@typescript-eslint/no-explicit-any": "error"
```

**Estimated time savings:**

- Manual setup: ~30-45 minutes per rule
- With scaffolding tools: ~5-10 minutes for setup, focus on implementation

## Success Metrics

**Tool Capabilities:**

- ✅ Generates compilable Go code (even before implementation)
- ✅ Follows RSLint naming conventions (snake_case directories, PascalCase types)
- ✅ Preserves existing code formatting standards
- ✅ Supports all three plugin types (TypeScript-ESLint, Import, Core)
- ✅ Handles batch processing of 10+ rules successfully
- ✅ Provides dry-run mode for safe previewing
- ✅ Integrates with testing infrastructure

**Generated Code Quality:**

- ✅ Proper package structure and imports
- ✅ Correct rule variable naming
- ✅ Dual-format options parsing (array/object)
- ✅ TODO comments for implementation guidance
- ✅ Example code patterns for common tasks
- ✅ Test file with proper structure

**Developer Experience:**

- ✅ Clear, comprehensive documentation
- ✅ Multiple usage examples
- ✅ Troubleshooting guide
- ✅ Integration examples
- ✅ Pre-made batch files for common rules

## Impact on Rule Migration

### Before These Tools

Manually creating a rule required:

1. Creating directory structure
2. Writing boilerplate imports and package declaration
3. Defining rule variable with correct naming
4. Setting up Run function signature
5. Creating options struct and parsing logic
6. Writing listener registration boilerplate
7. Creating test file with imports and structure
8. Manually adding imports to config.go
9. Manually adding registration call to config.go

**Time: 30-45 minutes per rule (before writing any logic)**

### After These Tools

With scaffolding tools:

1. Run `generate-rule.go` (generates all boilerplate)
2. Run `register-rule.go` (registers automatically)
3. Focus on implementing actual rule logic
4. Add test cases

**Time: 5-10 minutes for setup, then focus on implementation**

**For 150 rules:**

- Time saved: ~50-80 hours of boilerplate work
- Consistency: All rules follow exact same patterns
- Quality: No manual errors in structure/naming

## Architecture Alignment

These tools align with the RSLint architecture documented in `architecture.md`:

1. **Section 13 - Adding a New Rule**: Tools automate steps 1-4 of the checklist
2. **Section 6 - Lint Rule Framework**: Generated code follows the Rule interface exactly
3. **Section 14 - Dependency Layering**: Proper import paths for each plugin layer
4. **Section 12 - Testing Strategy**: Generated tests use the established rule_tester pattern

## Future Enhancements

Potential improvements for future iterations:

- [ ] **Smart AST Node Detection**: Analyze upstream rule to suggest AST node types
- [ ] **Test Case Generation**: Generate basic test cases from ESLint test suites
- [ ] **Rule Complexity Analysis**: Suggest whether type information is needed
- [ ] **Progress Tracking**: Dashboard for tracking rule migration progress
- [ ] **Dependency Detection**: Identify rules that depend on each other
- [ ] **Validation Tool**: Verify generated rules follow best practices
- [ ] **Migration Helper**: Convert ESLint rule implementations to RSLint patterns

## Testing the Tools

The tools themselves can be tested with:

```bash
# Dry-run test
go run scripts/generate-rule.go \
  -rule test-example \
  -plugin typescript-eslint \
  -dry-run

# Generate in temp directory
go run scripts/generate-rule.go \
  -rule test-example \
  -plugin typescript-eslint \
  -output /tmp/test-rules

# Verify compilation
cd /tmp/test-rules/test_example
go build .
```

## Files Added/Modified

### New Files

| File                                           | Lines     | Purpose                  |
| ---------------------------------------------- | --------- | ------------------------ |
| `scripts/generate-rule.go`                     | ~650      | Rule code generator      |
| `scripts/register-rule.go`                     | ~380      | Rule registry manager    |
| `scripts/README.md`                            | ~320      | Scripts documentation    |
| `docs/RULE_SCAFFOLDING_GUIDE.md`               | ~680      | Comprehensive user guide |
| `scripts/examples/typescript-eslint-rules.txt` | ~28       | Example batch file       |
| `scripts/examples/eslint-core-rules.txt`       | ~22       | Example batch file       |
| `scripts/examples/import-plugin-rules.txt`     | ~20       | Example batch file       |
| `RULE_SCAFFOLDING_SUMMARY.md`                  | This file | Implementation summary   |

### Modified Files

| File                     | Changes  | Purpose                       |
| ------------------------ | -------- | ----------------------------- |
| `scripts/dictionary.txt` | +8 lines | Add scaffolding-related words |

**Total:** ~2,100+ lines of new code and documentation

## Conclusion

These scaffolding tools provide a **comprehensive, production-ready solution** for rapidly implementing the 150+ ESLint and TypeScript-ESLint rules needed for RSLint. They:

- **Reduce boilerplate** by automating 80% of initial setup work
- **Ensure consistency** through template-based generation
- **Accelerate development** with batch processing capabilities
- **Integrate seamlessly** with existing testing infrastructure
- **Provide excellent DX** with comprehensive documentation and examples

The tools are ready for immediate use in the rule migration effort and will significantly accelerate RSLint's path to feature parity with ESLint/TypeScript-ESLint.
