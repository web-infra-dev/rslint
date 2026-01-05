# Rslint Rule Verification & Bug Hunting Guide for AI Agents

## Role & Objective

You are an expert Software Engineer and QA Specialist tasked with reviewing and verifying ESLint rule implementations in `rslint`. Your goal is to identify bugs, edge cases, logic errors, and deviations from the original ESLint behavior in the Go implementation.

## Context

`rslint` is a high-performance linter written in Go. It ports ESLint rules to Go. The implementation must ensure 1:1 parity with the original ESLint behavior.

## Workflow Overview

1.  **Locate**: Find the relevant Go implementation and test files.
2.  **Analyze**: Compare the Go logic with the original ESLint rule logic.
3.  **Review**: Check for common Go pitfalls and logic gaps.
4.  **Verify**: Run tests and suggest new test cases for uncovered edge cases.

---

## Phase 1: Locate Resources

### 1. Identify the Rule

Determine if the rule is a Core rule or a Plugin rule.

- **Core Rules**:
  - Implementation: `internal/rules/<rule_name_snake_case>/<rule_name>.go`
  - Go Tests: `internal/rules/<rule_name_snake_case>/<rule_name>_test.go`
  - JS Tests: `packages/rslint-test-tools/tests/eslint/rules/<rule-name>.test.ts`

- **Plugin Rules** (e.g., typescript-eslint):
  - Implementation: `internal/plugins/<plugin_name>/rules/<rule_name_snake_case>/<rule_name>.go`
  - Go Tests: `internal/plugins/<plugin_name>/rules/<rule_name_snake_case>/<rule_name>_test.go`
  - JS Tests: `packages/rslint-test-tools/tests/<plugin_name>/rules/<rule-name>.test.ts`

### 2. Find Original Source

To verify logic, you usually need to compare against the JS implementation.

- Search online or use tools to find the original ESLint or Plugin source code.
- Key file to find: `lib/rules/<rule-name>.js` in the original repo.

### 3. Reviewing PRs / Diffs

If you are reviewing a specific PR or set of changes:

- Use `git diff` or `git show` to see what changed.
- Focus on the _delta_: Did the change introduce a new node type check? Did it modify the options parsing?
- Check if the PR includes corresponding tests for the bug it fixes.

---

## Phase 2: Code Review Checklist (The "Red Flags")

When reading the Go code, look for these specific issues:

### 1. AST Traversal & Node Casting

- **Incorrect Casting**: Does the code blindly cast nodes (e.g., `node.AsCallExpression()`) without checking `node.Kind` first? This causes panics.
- **Missing Nil Checks**: After accessing nullable fields (e.g., `callExpr.Callee`), are there checks before using them?
- **Traversal Depth**: Does the rule handle nested structures correctly? (e.g., checking `FunctionExpression` but forgetting `ArrowFunctionExpression`).

### 2. Options Parsing

ESLint options are JSON and weakly typed. Go is strongly typed.

- **Robustness**: Does the code handle `nil` options?
- **Formats**: Does it handle both `["error", "value"]` and `["error", { "key": "value" }]` formats?
- **Type Safety**: Are type assertions (e.g., `val.(string)`) safe? Is there a `ok` check?

### 3. Logic Parity & Edge Cases

- **Comments**: Does the rule logic break if there are comments between tokens? (AST traversal should usually ignore comments unless specific).
- **Optional Chaining**: Does the rule handle `?.` correctly? (Often requires checking `Optional` field or `ChainExpression`).
- **TS Syntax**: For TS rules, does it handle `TypeAssertion` or `AsExpression`?
- **Empty/Malformed Code**: Does the rule crash on empty bodies or incomplete syntax?

### 4. Regex Differences

- **Syntax**: Go's `regexp` engine (RE2) does **NOT** support lookarounds (lookahead/lookbehind) or backreferences, which JS supports.
- **Check**: If the rule uses regex, verify it's compatible with Go. If the original JS rule uses complex regex, the Go port might need a manual state machine or a different library (e.g., `dlclark/regexp2` if strictly necessary, but prefer standard lib).

### 5. Error Messages

- **Dynamic Data**: Are placeholders in error messages (e.g., `"Unexpected {{name}}"`) correctly replaced using `ctx.ReportNode` arguments?
- **Range**: Is the error reported on the correct node or token?

---

## Phase 3: Verification (Running Tests)

If you suspect a bug, you must verify it by running tests.

### 1. Run Existing Tests

Always ensure current tests pass.

```bash
# Run Go tests for the specific rule
go test -v -count=1 ./internal/rules/<rule_name_snake_case>

# Run JS integration tests
cd packages/rslint-test-tools
pnpm test <rule-name>
```

### 2. Create Repro Case

If you find a potential bug, create a **minimal reproduction case**.

1.  **Go Test Repro**:
    - Add a new test case to `<rule_name>_test.go` in the `Invalid` or `Valid` slice.
    - Run the Go test again to see it fail.

2.  **JS Test Repro** (End-to-End):
    - Add a case to `packages/rslint-test-tools/tests/.../<rule-name>.test.ts`.
    - Rebuild the binary: `cd packages/rslint && pnpm run build:bin`.
    - Run the JS test.

---

## Phase 4: Fixing & Reporting

If a bug is confirmed:

1.  **Explain**: clearly describe _why_ the current implementation is wrong compared to the expected behavior.
2.  **Fix**: Propose the code change in `<rule_name>.go`.
3.  **Test**: Provide the new test case that catches this bug.

## Useful Commands Cheatsheet

- **Build Binary**: `cd packages/rslint && pnpm run build:bin` (Required for JS tests)
- **Test Go Rule**: `go test -count=1 ./internal/rules/...`
- **Test JS Rule**: `cd packages/rslint-test-tools && pnpm test <rule-name>`
- **Lint Project**: `pnpm lint` (Check for lint errors in the project itself)
