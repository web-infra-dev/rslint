# Rslint Rule Porting Guide for AI Agents

## Role & Objective

You are an expert Software Engineer tasked with porting ESLint rules to `rslint`, a high-performance linter written in Go. Your goal is to implement the rule logic in Go, ensuring 1:1 parity with the original ESLint behavior, including all edge cases and error messages.

## Workflow Overview

1.  **Preparation**: Gather requirements and test cases.
2.  **Implementation**: Write Go code and unit tests.
3.  **Integration**: Add JS tests and register the rule.
4.  **Verification**: Build and verify everything works.

---

## Phase 1: Preparation (CRITICAL)

**Goal**: Understand _exactly_ what the rule does before writing code.

1.  **Locate Official Source**:
    - **Priority**: If the user provides an official link, **FIRST** read and analyze that link's content.
    - **Fallback**: If no link is provided, search for the rule documentation (ESLint website or Plugin repo) and source code (GitHub).
    - Find the rule test file (usually `tests/lib/rules/<rule>.js`).

2.  **Collect Test Cases**:
    - Extract **ALL** `valid` and `invalid` cases from the official documentation.
    - Extract representative cases from the official unit tests, covering all branches of logic.
    - **Add Boundary Cases**: Add sufficient boundary cases (e.g., empty files, nested structures, edge cases in syntax) to ensure robust testing.
    - **Ensure Coverage**: Ensure Line and Column numbers are tested in invalid cases.
    - **Action**: Create a temporary scratchpad or keep these accessible. You will need them for both Go and JS tests.

3.  **Identify Edge Cases**:
    - Does the rule handle comments?
    - Does it handle Optional Chaining (`?.`)?
    - Does it handle TypeScript-specific syntax (if applicable)?
    - Does it handle empty bodies or malformed code?

---

## Phase 2: Implementation (Go)

### Step 1: Directory Setup

- **Core Rules**: `internal/rules/<rule_name_snake_case>/`
- **Plugin Rules**: `internal/plugins/<plugin_name>/rules/<rule_name_snake_case>/`

**Action**: Create the directory and two files:

1.  `<rule_name>.go` (Implementation)
2.  `<rule_name>_test.go` (Tests)

### Step 2: Write Rule Logic

**File**: `<rule_name>.go`

- **Prerequisite**:
  - **Understand the Rule Interface**: Read `internal/rule/rule.go` to understand core definitions like `rule.CreateRule`, `rule.RuleListeners`, `rule.RuleMessage`, and `rule.RuleContext`.
  - **Follow Patterns**: Reference existing rules (e.g., `internal/rules/constructor_super/constructor_super.go`) for the standard implementation pattern.
  - **Understand AST**: Review AST node types in `shim/ast/shim.go`.
  - **AST Shim API Inconsistency Warning**: In `github.com/microsoft/typescript-go/shim/ast`, some properties are accessed via **Methods** (e.g., `node.Name()`) while others are **Fields** (e.g., `node.AsIdentifier().Text`).
    - **General Nodes** (`*ast.Node`): Typically use methods (e.g., `node.Kind()`, `node.Text()`).
    - **Concrete Nodes** (e.g., `*ast.Identifier`): Typically use fields (e.g., `id.Text`).
    - **Action**: Do not assume; check the shim source code or use your IDE to confirm.

- **Implementation Checklist**:
  - Implement the `rule.Rule` interface.
    - **Note**: Use `rule.CreateRule` **ONLY** for `typescript-eslint` rules (it adds a prefix). For ESLint core rules or others, use the `rule.Rule{}` struct literal directly.
  - Define `rule.RuleListeners` to subscribe to specific AST node kinds (e.g., `ast.KindCallExpression`).
  - Define `rule.RuleMessage` for each error type.
  - Use `github.com/microsoft/typescript-go/shim/ast` for AST traversal.

**Rule Interface Reference (Complete Template)**:

```go
// For typescript-eslint rules:
var MyRuleRule = rule.CreateRule(rule.Rule{
    Name: "my-rule",
    // ...
})

// For ESLint Core rules:
var MyCoreRule = rule.Rule{
    Name: "my-core-rule",
    // ...
}
```

**Key Points**:

- The `Run` function signature is: `func(ctx rule.RuleContext, options any) rule.RuleListeners`
- `RuleListeners` is a map from `ast.Kind` to a callback function
- Each callback receives a `*ast.Node` and should check conditions and report diagnostics via `ctx.ReportNode()`
- Options parsing happens inside the `Run` function before returning the listeners

**Handling Options (Important!)**:

- **Structure**: Define a Go struct for your options (e.g., `type Options struct { Allow []string }`).
- **Parsing**: ESLint options are weakly typed (JSON). You **MUST** implement robust parsing logic to handle:
  - `nil` input (use defaults).
  - `[]interface{}` (array format, e.g., `["error", { "allow": [...] }]` -> take the object from index 0).
  - `map[string]interface{}` (object format).
  - Type assertions (safely cast `interface{}` to specific types).
- **Usage**: Parse options inside the `Run` function.

**Common Pitfall (Options Parsing)**:

- **Go Tests vs JS Tests**: Go tests typically pass `map[string]interface{}`, while JS integration tests (via `rslint.json` or `RuleTester`) pass `[]interface{}` (the arguments array from ESLint config, e.g., `["error", { ... }]`).
- **Action**: Your `parseOptions` function **MUST** handle both formats. If the input is `[]interface{}`, check if the first element is a map and use it.

**AST Cheatsheet (Avoid these common mistakes!)**:

- **Methods vs Fields**:
  - For generic `*ast.Node`: Use methods (e.g., `node.Kind()`, `node.Text()`).
  - For concrete types: Check `internal/rule/rule.go` or existing rules to see if fields are fields or methods.

- **Node Lists**: `node.Arguments` is usually `*ast.NodeList`. Iterate over `node.Arguments.Nodes`.

- **Casting**: Use helper methods like `node.AsIdentifier()`, `node.AsCallExpression()` to cast generic nodes.

- **Regex Literals**:
  - `RegularExpressionLiteral` text includes slashes and flags (e.g., `/abc/i`)
  - Extract pattern by finding the last `/` to separate pattern from flags
  - **IMPORTANT**: JS regex escape syntax (like `\u{...}`) differs from Go regex syntax. Parse manually if needed.

### Step 3: Write Go Tests

**File**: `<rule_name>_test.go`

- Use `rule_tester.RunRuleTester`.
- **Mandatory**: Port the "Valid" and "Invalid" cases you collected in Phase 1.
- **Requirements**: Invalid cases **MUST** include `Line` and `Column` assertions (e.g., `{MessageId: "foo", Line: 1, Column: 5}`). Tests will fail if these are missing.
- **Options**:
  - Use `map[string]interface{}` to pass options in Go tests (e.g., `Options: map[string]interface{}{"allow": []string{"warn"}}`).
  - Ensure you test both **default** (no options) and **custom** configurations.
- Ensure `tsconfig.json` path uses `fixtures.GetRootDir()`.

**Command**: `go test -count=1 ./internal/rules/<rule_name>`

---

## Phase 3: Integration (JS)

### Step 4: Check & Setup Test Environment

**Goal**: Ensure the test directory and necessary configuration files exist before adding tests.

1.  **Check Directory**:
    - Verify if `packages/rslint-test-tools/tests/<plugin-name>` exists.
    - If not, you **MUST** create it.

2.  **Check Configuration (Mandatory)**:
    - **Reference**: Use `packages/rslint-test-tools/tests/eslint-plugin-import` as the **primary reference** for directory structure and configuration files.
    - **Requirement**: The plugin test directory **MUST** contain the following files to run tests correctly:
      - `rslint.json` (Configuration for the linter)
      - `tsconfig.files.json` (TS Config for file-based tests)
      - `tsconfig.virtual.json` (TS Config for virtual/code-based tests)
    - **Action**:
      - If these files are missing (e.g., new directory), copy them **directly** from `packages/rslint-test-tools/tests/eslint-plugin-import/`.
      - **Critical**: Ensure `rslint.json` in the new directory contains `"project": ["./tsconfig.files.json", "./tsconfig.virtual.json"]` in `parserOptions`.
      - **Plugin Configuration**: In `rslint.json`, the `plugins` field must match your rule type.
        - **Core Rules**: Set to empty array `[]` (e.g., `"plugins": []`).
        - **Plugin Rules**: Set to the plugin name (e.g., `"plugins": ["eslint-plugin-import"]`).
      - Copy `rule-tester.ts` if missing.
      - **Warning (Prefix Issue)**: When copying `rule-tester.ts`, you **MUST** check for and remove any hardcoded rule prefixes (e.g., `ruleName = 'import/' + ruleName;`) inside the `run` method. This is a common cause of "Expected X errors but had 0" failures for Core Rules.
      - **Critical (Assertion Logic)**: The default `rule-tester.ts` copied from `eslint-plugin-import` has incomplete assertion logic for object-style errors (it only checks `message`). You **MUST** enhance the assertion logic to check `line`, `column`, `messageId`, `data`, etc., if your test cases use them.

### Step 5: Add JS Tests

**Important**: The location of the test file depends on whether it is a Core rule or a Plugin rule.

1.  **Core Rules**:
    - Path: `packages/rslint-test-tools/tests/eslint/rules/<rule-name>.test.ts`

2.  **Plugin Rules**:
    - Path: `packages/rslint-test-tools/tests/<plugin-name>/rules/<rule-name>.test.ts`
    - Create the `<plugin-name>` directory if it doesn't exist.
    - **Distinction of References**:
      - **Directory & Config**: Refer to `packages/rslint-test-tools/tests/eslint-plugin-import` (as per Step 4). You **MUST** ensure `rslint.json` and `tsconfig` files are present.
      - **Test Code Style**: Refer to `packages/rslint-test-tools/tests/typescript-eslint` for how to write the test cases.

**Action**:

1.  Create the file at the correct path.
2.  **Setup RuleTester**:
    - **typescript-eslint Rules**: Import `RuleTester` from `@typescript-eslint/rule-tester`.
    - **Other Rules**: Do **NOT** use `@typescript-eslint/rule-tester`. Refer to `packages/rslint-test-tools/tests/eslint-plugin-import/rule-tester.ts` to implement a local `RuleTester` wrapper.
3.  Paste the **exact same** test cases from Phase 1 (JS syntax).
    - **Options**: JS tests must use the ESLint array format for options (e.g., `options: [{ allow: ['warn'] }]`). Ensure this matches what your Go parser expects.

### Step 6: Register Test File

**File**: `packages/rslint-test-tools/rstest.config.mts`

- Add the new test file path to the `include` array.

### Step 7: Register Rule in Config

**File**: `internal/config/config.go`

- Import your new package.
- Register it in `registerAllCoreEslintRules()` (for core) or `registerAll<Plugin>Rules()` (for plugins).
- Format: `GlobalRuleRegistry.Register("rule-name", package.RuleName)`
- This registration is **required** for the rule to be discoverable by the linter.

### Step 8: Configure Test Environment (Conditional)

**File**: `packages/rslint-test-tools/rslint.json`

- **Only if this file is used by your test setup**: Ensure your rule is enabled:
  ```json
  "rules": {
    "rule-name": "error"
  }
  ```
- Verify the `files` pattern covers your test cases (e.g., `"**/*.ts"`, `"**/*.js"`).
- **If this file is NOT used**: The rule registration in Step 7 (`internal/config/config.go`) is sufficient; ensure the test file is included in `rstest.config.mts`.

---

## Phase 4: Verification & Build

**Goal**: Ensure the compiled binary runs the rule correctly against JS tests.

1.  **Build Binary (REQUIRED)**:
    - The JS tests run the _compiled binary_. You **MUST** rebuild after any Go code change.
    - Command: `cd packages/rslint && pnpm run build:bin`

2.  **Run JS Tests**:
    - Command: `cd packages/rslint-test-tools && npx rstest run --testTimeout=10000 <rule-name>`

3.  **Run Go Tests**:
    - Command: `go test -count=1 ./internal/rules/...`

---

## Troubleshooting & Workflow Lessons

**Common Failure: "Expected diagnostics... but received false"**
If your JS tests fail with 0 diagnostics found, check the following (in order):

1.  **Did you rebuild the binary after code changes?** (See Phase 4: Build Binary step)
2.  **Is the rule registered?** Check `internal/config/config.go` - rule must be in `GlobalRuleRegistry.Register()`.
3.  **Are test files included?** Check `rstest.config.mts` includes your test file path.
4.  **Is rslint.json configured?** (If used) Ensure the rule is enabled and file patterns are correct.
5.  **Debug Mode**: Use `fmt.Fprintf(os.Stderr, "DEBUG: ...")` in your Go code to verify execution.

**Appendix: Example AST Traversal**

```go
func checkNode(ctx rule.RuleContext, node *ast.Node) {
    if node.Kind != ast.KindCallExpression {
        return
    }
    callExpr := node.AsCallExpression()

    // Check Expression (e.g., "foo" in "foo()")
    if callExpr.Expression.Kind == ast.KindIdentifier {
        id := callExpr.Expression.AsIdentifier()
        if id.Text() == "eval" {
             ctx.AddDiagnostic(ctx.CreateDiagnostic(rule.RuleDiagnostic{
                Node:    node,
                Message: buildMessage("noEval"),
            }))
        }
    }
}
```

---

## Phase 5: Submission & PR (Automated)

**Goal**: Automate the submission of the new rule.

**Instructions for AI Agent**: Execute the following workflow to submit your changes.

1.  **Branch Management**:
    - **Naming Convention**: `feat/port-rule-<rule_name_snake_case>-<YYYYMMDD>`
    - **Action**: Create and switch to the new branch.
      ```bash
      git checkout -b feat/port-rule-<rule_name_snake_case>-$(date +%Y%m%d)
      ```

2.  **Commit Changes**:
    - **Constraint**: Commit **ONLY** changes related to the porting of this rule.
      - **Selectivity**: You **MUST** use specific file paths with `git add`. Do **NOT** use `git add .`. Only add files related to this port (implementation, tests, config).
      - **Verification**: You **MUST** ensure all tests pass before committing. Do **NOT** use `--no-verify` flag in `git commit`.
    - **Commit Message**: `feat: port rule <rule-name>`
    - **Action**:
      ```bash
      git add <specific_files_related_to_port>
      git commit -m "feat: port rule <rule-name>"
      ```

3.  **Push & Create PR**:
    - **Push**:
      ```bash
      git push origin feat/port-rule-<rule_name_snake_case>-$(date +%Y%m%d)
      ```
    - **Create PR**: Use the `gh` CLI to create the Pull Request.
      - **Title**: `feat: port rule <rule-name>`
      - **Body Requirement**: Must include the rule name, original documentation link, and the AI model used.
      - **Command Template**:

        ```bash
        gh pr create --base main --title "feat: port rule <rule-name>" --body "## Description
        Ported the \`<rule-name>\` rule from ESLint to rslint.

        ## References
        - **Original Rule**: <link_to_eslint_doc>

        ## Agent Info
        - **Model**: <ai_model_name>"
        ```
