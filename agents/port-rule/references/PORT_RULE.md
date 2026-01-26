# Rslint Rule Porting Guide for AI Agents

## Role & Objective

You are an expert Software Engineer tasked with porting ESLint rules to `rslint`, a high-performance linter written in Go. Your goal is to implement the rule logic in Go, ensuring 1:1 parity with the original ESLint behavior, including all edge cases and error messages.

---

## Related Documents

| Document                                       | Description                                                         |
| ---------------------------------------------- | ------------------------------------------------------------------- |
| [AST_PATTERNS.md](../../AST_PATTERNS.md)       | AST traversal patterns, listeners, TypeChecker, reporting functions |
| [UTILS_REFERENCE.md](../../UTILS_REFERENCE.md) | Utility functions in `internal/utils/`                              |
| [QUICK_REFERENCE.md](../../QUICK_REFERENCE.md) | Commands cheatsheet, file locations, naming conventions, checklist  |

---

## Source Code Reference

Before starting, familiarize yourself with these key source locations:

### Core Infrastructure

| File/Directory                        | Description                                                                                                              |
| ------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| `internal/rule/rule.go`               | **Core rule interface** - `Rule`, `RuleContext`, `RuleListeners`, `RuleMessage`, `RuleFix`, `RuleSuggestion` definitions |
| `internal/rule/disable_manager.go`    | Logic for handling `// eslint-disable` comments                                                                          |
| `internal/config/config.go`           | Rule registration and config loading                                                                                     |
| `internal/rule_tester/rule_tester.go` | Go test framework - `RunRuleTester`, `ValidTestCase`, `InvalidTestCase`                                                  |

### AST & Type System

| File/Directory                                        | Description                                                          |
| ----------------------------------------------------- | -------------------------------------------------------------------- |
| `typescript-go/_packages/ast/src/nodes.ts`            | **AST node type definitions** - All JS/TS syntax nodes               |
| `typescript-go/_packages/ast/src/syntaxKind.enum.ts`  | **SyntaxKind enum** - Node type constants (maps to Go's `ast.Kind*`) |
| `typescript-go/_packages/api/src/typeFlags.enum.ts`   | **TypeFlags enum** - Type checking flags                             |
| `typescript-go/_packages/api/src/symbolFlags.enum.ts` | **SymbolFlags enum** - Symbol flags                                  |
| `shim/ast/shim.go`                                    | Go-side AST shim implementation (auto-generated)                     |

### Example Rules (Recommended Reading)

| Rule                    | Path                                                 | Highlights                              |
| ----------------------- | ---------------------------------------------------- | --------------------------------------- |
| `no-debugger`           | `internal/rules/no_debugger/`                        | Simplest rule example                   |
| `constructor-super`     | `internal/rules/constructor_super/`                  | Complex control flow analysis           |
| `array-callback-return` | `internal/rules/array_callback_return/`              | Options parsing, function body analysis |
| `no-explicit-any`       | `internal/plugins/typescript/rules/no_explicit_any/` | TypeScript rule, Fix suggestions        |

---

## Workflow Overview

1. **Setup**: Create and switch to a new branch from main.
2. **Preparation**: Gather requirements and test cases.
3. **Implementation**: Write Go code and unit tests.
4. **Integration**: Add JS tests and register the rule.
5. **Verification**: Build and verify everything works.

---

## Phase 0: Branch Setup

**Goal**: Start from a clean state.

1. **Checkout Main**: Ensure your workspace is on the latest `main` branch code.

   ```bash
   git checkout main && git pull origin main
   ```

2. **Create Branch**: Create a new feature branch for the rule.
   - **Naming Convention**: `feat/port-rule-<rule_name_snake_case>-<YYYYMMDD>`
   ```bash
   git checkout -b feat/port-rule-<rule_name_snake_case>-$(date +%Y%m%d)
   ```

---

## Phase 1: Preparation (CRITICAL)

**Goal**: Understand _exactly_ what the rule does before writing code.

1. **Locate Official Source**:
   - **Priority**: If the user provides an official link, **FIRST** read and analyze that link's content.
   - **Fallback**: If no link is provided, search for the rule documentation (ESLint website or Plugin repo) and source code (GitHub).
   - Find the rule test file (usually `tests/lib/rules/<rule>.js`).

2. **Collect Test Cases**:
   - Extract **ALL** `valid` and `invalid` cases from the official documentation.
   - Extract representative cases from the official unit tests, covering all branches of logic.
   - **Add Boundary Cases**: Add sufficient boundary cases (e.g., empty files, nested structures, edge cases in syntax).
   - **Ensure Coverage**: Ensure Line and Column numbers are tested in invalid cases.

3. **Identify Edge Cases**:
   - Does the rule handle comments?
   - Does it handle Optional Chaining (`?.`)?
   - Does it handle TypeScript-specific syntax (if applicable)?
   - Does it handle empty bodies or malformed code?

---

## Phase 2: Implementation (Go)

### Step 1: Directory Setup

- **Core Rules**: `internal/rules/<rule_name_snake_case>/`
- **Plugin Rules**: `internal/plugins/<plugin_name>/rules/<rule_name_snake_case>/`

**Action**: Create the directory and three files:

1. `<rule_name>.go` (Implementation)
2. `<rule_name>_test.go` (Tests)
3. `<rule_name>.md` (Documentation)

### Step 2: Write Rule Logic

**File**: `<rule_name>.go`

**Prerequisites**:

- Read `internal/rule/rule.go` to understand core definitions
- Reference existing rules for the standard implementation pattern
- Review AST node types in `shim/ast/shim.go`
- See [AST_PATTERNS.md](../../AST_PATTERNS.md) for traversal patterns and examples

**Rule Interface**:

```go
// For typescript-eslint rules (auto-prefixes with @typescript-eslint/):
var MyRuleRule = rule.CreateRule(rule.Rule{
    Name: "my-rule",
    Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
        return rule.RuleListeners{
            ast.KindSomeNode: func(node *ast.Node) {
                // Check conditions and report
            },
        }
    },
})

// For ESLint Core rules:
var MyCoreRule = rule.Rule{
    Name: "my-core-rule",
    Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
        // ...
    },
}
```

**Key Points**:

- `RuleListeners` is a map from `ast.Kind` to a callback function
- Each callback receives a `*ast.Node` and reports diagnostics via `ctx.ReportNode()`
- Options parsing happens inside the `Run` function before returning listeners
- Use `rule.CreateRule` **ONLY** for `@typescript-eslint` rules (it adds the prefix)

**AST Shim API Warning**: In `github.com/microsoft/typescript-go/shim/ast`:

- **General Nodes** (`*ast.Node`): Use methods (e.g., `node.Kind()`, `node.Text()`)
- **Concrete Nodes** (e.g., `*ast.Identifier`): Use fields (e.g., `id.Text`)
- Do not assume; check the shim source code to confirm.

```go
// Example: Checking if callee is "Array"
if callee.Kind == ast.KindIdentifier {
    identifier := callee.AsIdentifier()

    // ✓ Correct - Text is a FIELD on concrete type
    if identifier.Text == "Array" { ... }

    // ✗ Wrong - Text is not a method!
    if identifier.Text() == "Array" { ... }  // Compile error
}
```

### Handling Options

ESLint options are weakly typed (JSON). You **MUST** handle both Go and JS test formats:

```go
func parseOptions(options any) Options {
    opts := Options{/* defaults */}
    if options == nil {
        return opts
    }

    var optsMap map[string]interface{}
    // Handle array format (JS tests pass []interface{})
    if arr, ok := options.([]interface{}); ok && len(arr) > 0 {
        optsMap, _ = arr[0].(map[string]interface{})
    } else {
        // Handle object format (Go tests pass map[string]interface{})
        optsMap, _ = options.(map[string]interface{})
    }

    if optsMap != nil {
        // Parse options from optsMap...
    }
    return opts
}
```

### Step 3: Write Documentation

**File**: `<rule_name>.md`

**Template**:

````markdown
# <rule-name>

## Rule Details

[Description of the rule]

Examples of **incorrect** code for this rule:

```javascript
// Example
var x = { a: 1, a: 2 };
```

Examples of **correct** code for this rule:

```javascript
// Example
var x = { a: 1, b: 2 };
```

## Original Documentation

[Link to ESLint documentation]
````

### Step 4: Write Go Tests

**File**: `<rule_name>_test.go`

- Use `rule_tester.RunRuleTester`
- Invalid cases **MUST** include `Line` and `Column` assertions
- Use `map[string]interface{}` to pass options in Go tests
- Ensure `tsconfig.json` path uses `fixtures.GetRootDir()`

**Debug Flags**:

```go
rule_tester.ValidTestCase{
    Code: `some code`,
    Only: true,  // Run only this test
}

rule_tester.InvalidTestCase{
    Code: `some code`,
    Skip: true,  // Skip this test
    Errors: []rule_tester.InvalidTestCaseError{...},
}
```

**Test Case Structs**: See `internal/rule_tester/rule_tester.go` for `ValidTestCase`, `InvalidTestCase`, and `InvalidTestCaseError` definitions.

---

## Phase 3: Integration (JS)

### Step 5: Check & Setup Test Environment

**Goal**: Ensure the test directory and necessary configuration files exist.

1. **Check Directory**: Verify if `packages/rslint-test-tools/tests/<plugin-name>` exists.

2. **Check Configuration**:
   - **Reference**: Use `packages/rslint-test-tools/tests/eslint-plugin-import` as the template.
   - **Required Files**:
     - `rslint.json` (Configuration for the linter)
     - `tsconfig.files.json` (TS Config for file-based tests)
     - `tsconfig.virtual.json` (TS Config for virtual/code-based tests)
   - **Plugin Configuration**: In `rslint.json`, set `plugins` field:
     - **Core Rules**: `"plugins": []`
     - **Plugin Rules**: `"plugins": ["eslint-plugin-import"]`
   - **Warning**: When copying `rule-tester.ts`, remove any hardcoded rule prefixes (e.g., `ruleName = 'import/' + ruleName;`).

### Step 6: Add JS Tests

**File Locations**:

- **Core Rules**: `packages/rslint-test-tools/tests/eslint/rules/<rule-name>.test.ts`
- **Plugin Rules**: `packages/rslint-test-tools/tests/<plugin-name>/rules/<rule-name>.test.ts`

**Setup RuleTester**:

- **typescript-eslint Rules**: Import `RuleTester` from `@typescript-eslint/rule-tester`
- **Other Rules**: Refer to `packages/rslint-test-tools/tests/eslint-plugin-import/rule-tester.ts`

**Options Format**: JS tests use array format: `options: [{ allow: ['warn'] }]`

### Step 7: Register Test File

**File**: `packages/rslint-test-tools/rstest.config.mts`

Add the new test file path to the `include` array.

### Step 8: Register Rule in Config

**File**: `internal/config/config.go`

1. Import your new package
2. Register in the appropriate function:
   - Core rules: `registerAllCoreEslintRules()`
   - TypeScript rules: `registerAllTypeScriptEslintPluginRules()`
   - Import rules: `registerAllEslintImportPluginRules()`
3. Format: `GlobalRuleRegistry.Register("rule-name", package.RuleNameRule)`

---

## Phase 4: Verification & Build

**Goal**: Ensure the compiled binary runs the rule correctly.

1. **Build Binary (REQUIRED)**:

   ```bash
   cd packages/rslint && pnpm run build:bin
   ```

2. **Run Tests**:

   ```bash
   # Go tests
   go test -count=1 ./internal/rules/<rule_name>

   # JS tests
   cd packages/rslint-test-tools && npx rstest run --testTimeout=10000 <rule-name>
   ```

3. **Verify Test Coverage Alignment**:

   Ensure Go tests cover the same cases as JS tests:
   - Check the JS test snapshot file for the number of invalid cases
   - Go tests should include equivalent test cases
   - Pay special attention to edge cases:
     - Expressions with comments (e.g., `/* a */ foo /* b */ ()`)
     - Multi-line expressions
     - Nested structures (e.g., `foo((x), y)`, `foo(bar(), baz())`)

4. **Project-wide Checks**:

   ```bash
   # Type check and lint
   pnpm typecheck && pnpm lint

   # Format and Go lint checks (REQUIRED before commit)
   pnpm format:check && pnpm lint:go
   ```

   **If checks fail**, run these to auto-fix:

   ```bash
   pnpm format      # Fix JS/TS formatting
   pnpm format:go   # Fix Go formatting (e.g., import order)
   ```

---

## Phase 5: Submission & PR

**Goal**: Submit the new rule.

1. **Configure Project Settings (Conditional)**:
   - If the rule's plugin is already in `rslint.json` `plugins`, add the rule with `"warn"` severity
   - Otherwise, do NOT modify `rslint.json`

2. **Commit Changes**:

   ```bash
   git add <specific_files_related_to_port>
   git commit -m "feat: port rule <rule-name>"
   ```

   - Use specific file paths with `git add` (NOT `git add .`)
   - Ensure all tests pass before committing

3. **Push & Create PR**:

   ```bash
   git push origin feat/port-rule-<rule_name_snake_case>-$(date +%Y%m%d)

   gh pr create --base main --title "feat: port rule <rule-name>" --body "## Description
   Ported the \`<rule-name>\` rule from ESLint to rslint.

   ## References
   - **Original Rule**: <link_to_eslint_doc>

   ## Agent Info
   - **Model**: <ai_model_name>"
   ```

---

## Troubleshooting

### "Expected diagnostics... but received false"

If JS tests fail with 0 diagnostics found:

1. **Did you rebuild the binary?** Run `cd packages/rslint && pnpm run build:bin`
2. **Is the rule registered?** Check `internal/config/config.go`
3. **Are test files included?** Check `rstest.config.mts`
4. **Is rslint.json configured?** Ensure the rule is enabled
5. **Debug Mode**: Use `fmt.Fprintf(os.Stderr, "DEBUG: ...")` in Go code

### Line/Column Mismatch

1. Check 0-based vs 1-based column expectations
2. Multi-byte characters may affect column calculation
3. Debug with:
   ```go
   pos := ctx.SourceFile.LineMap().LineAndColumn(node.Pos())
   fmt.Fprintf(os.Stderr, "DEBUG: Line=%d, Column=%d\n", pos.Line, pos.Column)
   ```

### TypeChecker is nil

1. Ensure test fixtures have correct `tsconfig.json`
2. Use `fixtures.GetRootDir()` for correct path
3. Check `parserOptions.project` is configured correctly

---

## See Also

- [AST_PATTERNS.md](../../AST_PATTERNS.md) - AST traversal, listeners, reporting functions, fix helpers
- [UTILS_REFERENCE.md](../../UTILS_REFERENCE.md) - Utility functions in `internal/utils/`
- [QUICK_REFERENCE.md](../../QUICK_REFERENCE.md) - Commands, file locations, naming conventions, checklist
