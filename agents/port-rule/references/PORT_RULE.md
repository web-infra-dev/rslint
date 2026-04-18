# Rslint Rule Porting Guide

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
| `internal/rule/disable_manager.go`    | Logic for handling `// rslint-disable` and `// eslint-disable` comments                                                  |
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

2. **Create Branch**: Create a new feature branch.

   **Single rule**:
   - **Naming Convention**: `feat/port-rule-<rule_name_snake_case>-<YYYYMMDD>`

   ```bash
   git checkout -b feat/port-rule-<rule_name_snake_case>-$(date +%Y%m%d)
   ```

   **Batch mode**:
   - **Naming Convention**: `feat/port-rules-batch-<YYYYMMDD>`

   ```bash
   git checkout -b feat/port-rules-batch-$(date +%Y%m%d)
   ```

---

## Phase 1: Preparation (CRITICAL)

**Goal**: Understand _exactly_ what the rule does before writing code.

1. **Locate Official Source**:
   - **Priority**: If the user provides an official link, **FIRST** read and analyze that link's content.
   - **Fallback**: If no link is provided, search for the rule documentation (ESLint website or Plugin repo) and source code (GitHub).
   - Find the rule test file (usually `tests/lib/rules/<rule>.js`).

2. **Determine Rule Origin & Deprecation Status**:

   Some rules exist in both core ESLint and typescript-eslint. Before implementing, determine the canonical source:

   | Scenario                                                                         | Registration                                | Test Location                    | Rule Wrapper        |
   | -------------------------------------------------------------------------------- | ------------------------------------------- | -------------------------------- | ------------------- |
   | **Core ESLint only** (e.g., `no-debugger`)                                       | `"no-debugger"`                             | `tests/eslint/rules/`            | `rule.Rule{}`       |
   | **typescript-eslint only** (e.g., `await-thenable`)                              | `"@typescript-eslint/await-thenable"`       | `tests/typescript-eslint/rules/` | `rule.CreateRule()` |
   | **typescript-eslint extends core** (active, e.g., `no-array-constructor`)        | `"@typescript-eslint/no-array-constructor"` | `tests/typescript-eslint/rules/` | `rule.CreateRule()` |
   | **typescript-eslint deprecated in favor of core** (e.g., `no-loss-of-precision`) | `"no-loss-of-precision"`                    | `tests/eslint/rules/`            | `rule.Rule{}`       |

   **How to check**: Visit the typescript-eslint rule page. If it shows a deprecation notice like _"use the base ESLint rule instead"_, treat it as a **core ESLint rule** — do NOT register with `@typescript-eslint/` prefix.

3. **Collect Test Cases**:
   - Extract **ALL** `valid` and `invalid` cases from the official documentation.
   - Migrate **ALL** `valid` and `invalid` cases from the official unit test file (`tests/lib/rules/<rule>.js` for ESLint core; plugin equivalents otherwise) — not a representative subset. The ESLint suite is the **lower bound** for coverage, not the upper bound.
   - **Skip with explanation**: If a case exercises an option or syntax we intentionally don't support, keep it in the file as a `Skip: true` test with a `// SKIP: <reason>` comment — don't drop it silently.
   - **Add extra edge cases on top**: Beyond the ESLint suite, add cases that exercise tsgo-specific AST quirks (see [AST_PATTERNS.md § AST Shape Essentials](../../AST_PATTERNS.md#ast-shape-essentials)) — nested expressions, paren / bracket forms, reserved words in various positions, declaration merging, computed keys the upstream parser may not distinguish, etc.
   - **Ensure Coverage**: Ensure Line and Column numbers are tested in invalid cases.

4. **Identify Edge Cases**:

   Systematically enumerate edge cases across three dimensions:

   **Dimension 1: AST node types** — List every syntax construct the rule should handle:
   - All access patterns (e.g., `.prop`, `['prop']`, ``[`prop`]``)
   - Optional Chaining (`?.`)
   - TypeScript-specific syntax (type annotations, generics, enums, etc.)
   - Async functions, generators, arrow functions
   - Empty bodies, malformed code

   **Dimension 2: Scoping & nesting** — Enumerate nested combinations:
   - Function / arrow / method / constructor / getter / setter crossed with each other
   - Class bodies, computed property names, extends clauses, static blocks
   - `this` / `super` binding semantics across scope boundaries
   - Deeply nested patterns (3+ levels)

   **Dimension 3: Autofix boundaries** (if the rule has autofix):
   - Comments between tokens that must be preserved
   - Arguments with side effects (should suppress autofix)
   - Parenthesized expressions (multiple levels)
   - Multi-line code with varying whitespace

5. **Document Divergence from ESLint**:

   Two classes of divergence may arise when porting. Both must be documented; they differ in _how_ and _where_.

   **A. Intentional divergence** — a choice we make (e.g. more precise error locations, different reporting granularity). Do all three:
   1. **Source code comment**: Add a `// NOTE: Unlike ESLint...` explaining the difference and rationale.
   2. **Rule documentation**: Add a "Differences from ESLint" section in the rule's `.md` file.
   3. **Test cases**: Ensure the differing behavior is covered by a dedicated test — a green-path `ValidTestCase` or a case with an exact `Message` / position assertion — so that future refactors can't silently flip it.

   **B. Language-natural divergence** — a side effect of tsgo's AST or Go semantics that we don't actively choose (e.g. tsgo decimal-normalizes `NumericLiteral` at parse time, so a dynamic computed key `[0x1]` compares equal to `[1]` where ESLint's token-level comparison sees them as distinct). Usually more permissive than ESLint.
   1. **Rule documentation** (or [AST_PATTERNS.md](../../AST_PATTERNS.md) if the quirk is general, not rule-specific): note the divergence under "Differences from ESLint" / the relevant AST-shape section.
   2. **Test cases**: Lock the current behavior in with a test — typically the ESLint-fails-but-we-pass case stays on the `valid` side with a comment pointing at the underlying quirk, so the behavior can't flip silently.

---

## Phase 2: Implementation (Go)

> **AST note**: rslint is built on the tsgo AST, which is structurally different from ESLint's ESTree. Child-access patterns (`node.left`, `node.argument`, `node.callee`, …) do **not** correspond 1:1: parentheses are explicit nodes, optional chains are flag-based (no `ChainExpression` wrapper), `Literal` is split across several `Kind*Literal` kinds, and `AssignmentExpression` / `SequenceExpression` collapse into `BinaryExpression`. Review [AST_PATTERNS.md § AST Shape Essentials](../../AST_PATTERNS.md#ast-shape-essentials) before implementing, and run the Alignment Audit (end of Step 2) before tests.
>
> **If you discover a new tsgo↔ESTree shape difference during porting** (e.g. a kind that has no ESTree analog, an `.Text` field that's normalized at parse time when ESLint sees raw source, an access pattern that requires an extra unwrap), **append it to [AST_PATTERNS.md § AST Shape Essentials](../../AST_PATTERNS.md#ast-shape-essentials) as part of your PR**. That file is the living knowledge base; every new rule is a chance to grow it.

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

**Check for reusable `internal/utils/` helpers** (FIRST): Before writing any helper function, grep `internal/utils/` for an existing one. Helpful prefixes to search:

- `IsSpecific*`, `IsArgument*` — well-known API-call recognition (`Object.defineProperty`-style, member-access patterns, nth-argument-of)
- `GetStatic*`, `Normalize*` — property-name / literal-value normalization (e.g. `GetStaticPropertyName`, `NormalizeNumericLiteral`, `NormalizeBigIntLiteral`)
- `AreNodes*`, `IsSame*` — structural / reference AST comparison
- `GetFunction*`, `TrimmedNodeText*`, `TrimNodeTextRange` — function head / trimmed source text
- `IsShadowed`, `FindEnclosingScope`, `CollectBindingNames` — scope / binding queries
- `GetOptionsMap` — options parsing (handles both array and map inputs)
- **Type-aware queries** (for `@typescript-eslint` rules that use `ctx.TypeChecker`): `Is*Type*` / `Get*Type*` — type-flag tests and classifications (`IsTypeAnyType`, `IsUnionType`, `GetTypeName`, `GetContextualType`, `GetConstraintInfo`); `IsPromise*` / `IsError*` / `IsReadonly*` — builtin-type detection; `NeedsToBeAwaited`, `GetCallSignatures`, `CollectAllCallSignatures` — signature / awaitability helpers; `IsUnsafeAssignment`, `DiscriminateAnyType` — any-type safety. See the `ts_api_utils.go` / `ts_eslint.go` / `builtin_symbol_likes.go` sections of [UTILS_REFERENCE.md](../../UTILS_REFERENCE.md) for the complete inventory — **do not re-implement type analysis inline**.

See [UTILS_REFERENCE.md](../../UTILS_REFERENCE.md) for the full inventory. **If you find a near-match that's missing some behavior, extend it in place** rather than writing a parallel implementation inline. Extraction is explicitly preferred over duplication (see _Helper Extraction_ below for criteria).

**Check for reusable shim utilities** (SECOND): If `internal/utils/` has nothing, check if the `shim/` packages already provide what you need:

- `shim/scanner/` — `SkipTrivia` (skip whitespace/comments to find next token position), `GetScannerForSourceFile`, `GetSourceTextOfNodeFromSourceFile` (raw source text — useful when an AST node's `.Text` field has been normalized at parse time)
- `shim/ast/` — `GetThisContainer`, `IsFunctionLike`, `IsFunctionLikeDeclaration`, `SkipParentheses`, `IsOptionalChain`, and other AST utilities
- `shim/checker/` — native tsgo TypeChecker methods exposed as `Checker_*` functions (`GetReturnTypeOfSignature`, `GetApparentType`, `GetWidenedType`, `GetTypeArguments`, `GetPropertyOfType`, `GetIndexInfosOfType`, …). Reach here **only when** `internal/utils/` doesn't already wrap what you need; the wrappers encode invariants you'd otherwise have to re-derive. See `shim/checker/shim.go` for the full surface.
- `shim/core/` — `NewTextRange` and other core utilities

> **Warning**: Some shim functions have different semantics from ESLint's model. For example, `ast.GetThisContainer` treats `PropertyDeclaration`, `ClassStaticBlockDeclaration`, `ModuleDeclaration`, etc. as `this` containers, which does not match ESLint's scope model. Always compare the shim function's behavior against ESLint before reusing.

**Rule Interface**:

```go
// For typescript-eslint rules that use TypeChecker (auto-prefixes with @typescript-eslint/):
var MyRuleRule = rule.CreateRule(rule.Rule{
    Name:             "my-rule",
    RequiresTypeInfo: true,
    Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
        return rule.RuleListeners{
            ast.KindSomeNode: func(node *ast.Node) {
                // ctx.TypeChecker is guaranteed non-nil when RequiresTypeInfo is true
            },
        }
    },
})

// For typescript-eslint rules that do NOT use TypeChecker:
var MyOtherRule = rule.CreateRule(rule.Rule{
    Name: "my-other-rule",
    Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
        // ...
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
- **`RequiresTypeInfo`**: If a `@typescript-eslint` rule uses `ctx.TypeChecker`, you **MUST** set `RequiresTypeInfo: true`. This tells the linter to skip the rule on files without a type checker, preventing nil-pointer panics. Core ESLint rules should NOT set this flag — use `ctx.TypeChecker == nil` guards instead (see [AST_PATTERNS.md — Using TypeChecker](../../AST_PATTERNS.md#using-typechecker)).
- **MessageId convention**: Use camelCase for `RuleMessage.Id` (e.g., `"unexpectedAny"`, `"missingSuper"`). Match the original ESLint rule's messageId names. The JS rule-tester has a `toCamelCase` compatibility layer, but new rules should use camelCase directly.

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

ESLint options are weakly typed (JSON). Use `utils.GetOptionsMap()` to extract the options map — it handles both array format (`[]interface{}` from JS tests) and direct object format (`map[string]interface{}` from Go tests):

```go
func parseOptions(options any) Options {
    opts := Options{/* defaults */}
    optsMap := utils.GetOptionsMap(options)
    if optsMap != nil {
        // Parse options from optsMap...
    }
    return opts
}
```

### Alignment Audit

Before moving on, walk through each check. Each one targets a class of AST-shape bug that is not caught by compilation and may slip past narrowly-written unit tests. Skip a row when it doesn't apply to your rule.

| If the rule …                                                                   | Audit                                                                                                                                                             | Reference                                                                                  |
| ------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------ |
| Reads a child node (every rule)                                                 | Any `.Kind ==` / `.Kind !=` / `.As<Type>()` / `.Text` access on a child must go through `ast.SkipParentheses` first (directly or via a helper).                   | [AST_PATTERNS.md § ParenthesizedExpression](../../AST_PATTERNS.md#parenthesizedexpression) |
| Handles `foo?.bar` / `foo?.()`                                                  | Use `ast.IsOptionalChain(node)`; don't hand-check node flags.                                                                                                     | [AST_PATTERNS.md § Optional Chain](../../AST_PATTERNS.md#optional-chain)                   |
| Compares literal values                                                         | Match the precise `Kind*Literal`; normalize numeric text via `utils.NormalizeNumericLiteral` before value comparison.                                             | [AST_PATTERNS.md § Literal Kinds](../../AST_PATTERNS.md#literal-kinds)                     |
| Has separate ESLint listeners for `AssignmentExpression` / `SequenceExpression` | Collapse into one `BinaryExpression` listener and branch on `OperatorToken.Kind`.                                                                                 | [AST_PATTERNS.md § Binary Operator Kinds](../../AST_PATTERNS.md#binary-operator-kinds)     |
| Emits fix/suggestion text starting with an identifier                           | Guard against token fusion with the preceding character before emitting (otherwise e.g. `typeof` + `Number(foo)` becomes `typeofNumber(foo)`).                    | —                                                                                          |
| Checks whether a name resolves to a global                                      | Use `utils.IsShadowed(node, name)`. Note: stricter than ESLint's scope manager on TS type-only bindings — document in the rule's `.md` if the difference matters. | —                                                                                          |
| Reads source text for recommendation / fix                                      | Prefer `utils.TrimmedNodeText(sf, node)` (skips leading trivia) over raw `node.Pos()/End()`.                                                                      | [AST_PATTERNS.md § Node Text and Positions](../../AST_PATTERNS.md#node-text-and-positions) |

### Helper Extraction

After Step 2 is done, review each helper for extraction to `internal/utils/`:

**Extract if all hold:**

- Input/output is AST- or source-oriented (not encoding the rule's own semantics)
- The name reads sensibly without context of the current rule
- Another rule would plausibly need the same thing

**Keep local otherwise.** Predicates that encode a specific rule's definition (e.g. a `isDoubleLogicalNegating`-style helper that codifies "what counts as a double-negation coercion for THIS rule") stay with the rule — extracting would mislead future readers.

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
- **All Go test cases go into one `<rule_name>_test.go` file** — do not split by feature, option, or container type. Single-file organization keeps grep / diff against the upstream ESLint test file trivial.
- **Preserve ESLint's original grouping as inline comments** (e.g. `// ---- Various getter keys ----`, `// ---- Property descriptors ----`). Future audits should be able to read the file top-to-bottom and match the upstream layout.
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

**Autofix Testing**: If the rule provides autofix, use the `Output` field to verify the fixed code:

```go
// With autofix: provide Output field with the expected fixed code
rule_tester.InvalidTestCase{
    Code:   `var a = function() { return 1; }.bind(b)`,
    Output: []string{`var a = function() { return 1; }`},
    Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
}

// Without autofix (e.g., side-effect argument): omit Output field
rule_tester.InvalidTestCase{
    Code:   `var a = function() {}.bind(b++)`,
    Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
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

**File Locations** (determined by Phase 1 Step 2):

- **Core ESLint Rules** (including deprecated typescript-eslint rules): `packages/rslint-test-tools/tests/eslint/rules/<rule-name>.test.ts`
- **typescript-eslint Rules**: `packages/rslint-test-tools/tests/typescript-eslint/rules/<rule-name>.test.ts`
- **Other Plugin Rules**: `packages/rslint-test-tools/tests/<plugin-name>/rules/<rule-name>.test.ts`

**Setup RuleTester**:

- **Core ESLint Rules**: Import `RuleTester` from `../rule-tester` (in `tests/eslint/rule-tester.ts`, no prefix)
- **typescript-eslint Rules**: Import `RuleTester` from `@typescript-eslint/rule-tester` (auto-prefixes with `@typescript-eslint/`)
- **Other Plugin Rules**: Refer to `packages/rslint-test-tools/tests/eslint-plugin-import/rule-tester.ts`

**Options Format**: JS tests use array format: `options: [{ allow: ['warn'] }]`

### Step 7: Register Test File

**File**: `packages/rslint-test-tools/rstest.config.mts`

Add the new test file path to the `include` array.

### Step 8: Register Rule in Config

**File**: `internal/config/config.go`

1. Import your new package
2. Register in the appropriate function (determined by Phase 1 Step 2):
   - **Core ESLint rules** (including deprecated typescript-eslint rules): `registerAllCoreEslintRules()` with `rule.Rule{}`
   - **typescript-eslint rules** (active): `registerAllTypeScriptEslintPluginRules()` with `rule.CreateRule()`
   - **Import plugin rules**: `registerAllEslintImportPluginRules()`
3. Format: `GlobalRuleRegistry.Register("rule-name", package.RuleNameRule)`
4. **Do NOT register a rule under both `"rule-name"` and `"@typescript-eslint/rule-name"`** — pick the canonical one based on deprecation status

---

## Phase 4: Verification & Build

**Goal**: Ensure the compiled binary runs the rule correctly.

Follow this **strict order** — each step depends on the previous one:

1. **Go formatting** (catches indentation issues early):

   ```bash
   gofmt -l internal/rules/<rule_name>/
   ```

   If files are listed, run `gofmt -w` on them to fix.

2. **Go tests**:

   ```bash
   go test -count=1 ./internal/rules/<rule_name>
   ```

3. **Build binary** (REQUIRED before JS tests — they spawn the binary via IPC):

   ```bash
   cd packages/rslint && pnpm run build:bin
   ```

4. **JS tests** (note: this changes cwd, use absolute paths for subsequent commands):

   ```bash
   # First run for new test cases: generate snapshots with -u flag
   cd packages/rslint-test-tools && npx rstest run --testTimeout=10000 <rule-name> -u

   # Subsequent runs: verify against existing snapshots
   cd packages/rslint-test-tools && npx rstest run --testTimeout=10000 <rule-name>
   ```

5. **Verify Test Coverage Alignment (Go ↔ JS)**:

   Ensure Go tests cover the same cases as JS tests:
   - Check the JS test snapshot file for the number of invalid cases
   - Go tests should include equivalent test cases
   - Pay special attention to edge cases:
     - Expressions with comments (e.g., `/* a */ foo /* b */ ()`)
     - Multi-line expressions
     - Nested structures (e.g., `foo((x), y)`, `foo(bar(), baz())`)

   **Go vs JS test differences**:

   | Aspect           | Go tests                                | JS tests                                    |
   | ---------------- | --------------------------------------- | ------------------------------------------- |
   | Autofix          | `Output: []string{...}` field           | Not verified (snapshot filters out `fixes`) |
   | Position         | `Line`/`Column` fields on each error    | Implicitly covered by snapshot              |
   | Multiple errors  | `Errors: []...{{...}, {...}}`           | `errors: [{...}, {...}]`                    |
   | MessageId format | camelCase (e.g., `"noLossOfPrecision"`) | camelCase (e.g., `"noLossOfPrecision"`)     |

6. **Contract Alignment Checklist (Go ↔ ESLint)**:

   Step 5 verifies our two test suites agree with each other. This step verifies the **public contract** of the rule agrees with ESLint. The oracle is ESLint's diagnostic output (`messageId` + message text + report position) and its options schema — **not** ESLint's internal implementation. Language-level implementation differences are acceptable (see Phase 1 Step 5.B); contract differences are not.

   Before claiming the port is aligned, confirm every row. Missing any row means the claim is premature:
   - [ ] **Full ESLint test migration** — every `valid` / `invalid` case from the upstream unit-test file has a corresponding Go case (or a `Skip: true` with a `// SKIP: <reason>` comment).
   - [ ] **Message text assertions** — each `messageId` has **at least one** test using the `InvalidTestCaseError.Message` field (exact string match), covering every modifier combination the rule can emit (`static`, `private`, `async`, computed-no-name, etc.).
   - [ ] **Position assertions per container** — for each container the rule emits into (object literal / class / type / descriptor / …), at least 2 cases assert `Line` + `Column` + `EndLine` + `EndColumn`, including one multi-line case.
   - [ ] **Options schema match** — option names, types, and **defaults** match ESLint's schema byte-for-byte. Assert every default by running an invalid/valid case with no options vs. `[{}]` options and confirming identical output.
   - [ ] **Options combination matrix** — for every boolean option, include at least one test where it is `true` and one where it is `false`. Triggering combinations (e.g. rule behaves differently when two options are both on) get dedicated cases.
   - [ ] **Three-way equivalence classes** (if the rule compares names / keys) — static / private / dynamic keys form separate equivalence classes; test at least one cross-class negative (e.g. `'#a'` string vs `#a` private identifier should NOT pair up).

7. **Project-wide Checks**:

   ```bash
   # Type check and lint
   pnpm typecheck && pnpm lint

   # Spell check (catches typos in comments and strings)
   pnpm -w run check-spell

   # Format and Go lint checks (REQUIRED before commit)
   pnpm format:check && pnpm lint:go
   ```

   **If checks fail**, run these to auto-fix:

   ```bash
   pnpm format      # Fix JS/TS formatting
   pnpm format:go   # Fix Go formatting (e.g., import order)
   ```

8. **Differential Validation** (recommended for rules with non-trivial branching):

   Unit tests verify cases you thought of; diffing against the reference implementation on a real codebase catches the rest. Skip when the rule has ≤ 2 branches and trivial messages, or when the rule is a new rslint invention with no reference.

   **Procedure**:

   ```bash
   # 1. Scratch-install the reference tool.
   mkdir -p /tmp/ref-cmp && cd /tmp/ref-cmp
   npm init -y >/dev/null
   npm i --silent eslint @typescript-eslint/parser  # + plugin if non-core
   cat > eslint.config.mjs <<'EOF'
   import parser from '@typescript-eslint/parser';
   export default [{
     files: ['**/*.ts', '**/*.tsx'],
     languageOptions: { parser },
     rules: { '<rule-name>': 'warn' },
   }];
   EOF

   # 2. Pick a target codebase that exercises typical patterns.
   # 3. Run both; normalize to sorted JSON of {file, line, col, messageId, message}; diff.
   ```

   **Prerequisite for type-info rules**: the reference tool must run with the same `parserOptions.project` / `tsconfig.json` as rslint, otherwise the comparison is meaningless. Pick a target codebase where the tsconfig loads cleanly under both tools.

   **Interpreting a non-empty diff**:

   | Diff kind                       | Likely cause                                           |
   | ------------------------------- | ------------------------------------------------------ |
   | rslint misses a report          | AST-shape mismatch (often a missing `SkipParentheses`) |
   | rslint over-reports             | Same as above, inverted                                |
   | Different message text          | paren / text-range handling in the recommendation      |
   | Same count, different positions | column offset (0- vs 1-based, multibyte)               |

   **Disposition standard**: a non-empty diff is **not** automatically a failure. Every differing line must fall into exactly one of the three categories below — anything that cannot be confidently classified is treated as (c).

   | Category                            | What it means                                                                                                                                                                                   | Action                                                                                                                        |
   | ----------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
   | **(a) Language-natural divergence** | tsgo AST or Go-semantic effect we don't actively choose (see Phase 1 Step 5.B — e.g. `NumericLiteral` parse-time normalization, normalized string cooked values).                               | Document under the rule's `.md` "Differences from ESLint" (or in [AST_PATTERNS.md](../../AST_PATTERNS.md) if general). Leave. |
   | **(b) Scan-scope divergence**       | The two tools see different file sets (e.g., rslint respects `.gitignore` by default; ESLint does not; tsconfig `include` excludes a dir). Not a rule issue.                                    | No action. Optionally note in the PR description if a reviewer might be confused.                                             |
   | **(c) Genuine bug**                 | Neither (a) nor (b). Rule logic, message text, or position is actually wrong on our side (or, rarely, ESLint's — but we align to ESLint unless we have a standing Phase 1 Step 5.A divergence). | **Must fix** before merging. Re-run the diff until it clears or reduces to (a)/(b) only.                                      |

---

## Phase 5: Submission & PR

### Phase 5A: Per-Rule Commit

**Goal**: Commit each rule independently after it passes verification (Phase 4).

This step is executed **after each rule's Phase 4 completes** (both in single-rule and batch mode).

1. **Configure Project Settings (Conditional)**:
   - If the rule's plugin is already in `rslint.json` `plugins`, add the rule with `"warn"` severity
   - Otherwise, do NOT modify `rslint.json`

2. **Commit Changes**:

   ```bash
   git add <specific_files_related_to_this_rule>
   git commit -m "feat: port rule <rule-name>"
   ```

   - Use specific file paths with `git add` (NOT `git add .`)
   - Only include files related to **this specific rule** in the commit
   - Ensure all tests pass before committing
   - **Do NOT include AI-related information** in commit messages (e.g., no `Co-Authored-By: Claude` or similar)

**In batch mode**: After committing, briefly report the result, update the checklist, then proceed to the next rule.

### Phase 5B: Push & Create PR

**Goal**: Push all commits and create a PR.

This step is executed **once**, after all rules are committed (or after the single rule is committed).

1. **Push**:

   ```bash
   git push origin <branch-name>
   ```

2. **Create PR**:

   **Important**: Use the repository's PR template at `.github/PULL_REQUEST_TEMPLATE.md`.

   **Single rule**:

   ```bash
   gh pr create --base main --title "feat: port rule <rule-name>" --body "## Summary

   Port the \`<rule-name>\` rule from ESLint to rslint.

   [Brief description of what the rule does]

   ## Related Links

   - ESLint rule: <link_to_eslint_doc>
   - Source code: <link_to_source_code>

   ## Checklist

   - [x] Tests updated (or not required).
   - [x] Documentation updated (or not required)."
   ```

   **Batch mode (single plugin)**:

   PR title format: `feat: port N <plugin-name> rules`

   ```bash
   gh pr create --base main --title "feat: port N <plugin-name> rules" --body "## Summary

   Port N <plugin-name> rules to rslint.

   ### Rules ported
   | Rule | Description | Doc |
   |------|-------------|-----|
   | \`<rule-1>\` | [brief description] | [link](<url>) |
   | \`<rule-2>\` | [brief description] | [link](<url>) |
   | ... | ... | ... |

   ## Checklist

   - [x] Tests updated (or not required).
   - [x] Documentation updated (or not required)."
   ```

   Examples:
   - `feat: port 4 @typescript-eslint non-null assertion rules`
   - `feat: port 3 eslint-plugin-import rules`

   **Batch mode (multiple plugins)**:

   PR title format: `feat: port N rules from <plugin-1>, <plugin-2>`

   ```bash
   gh pr create --base main --title "feat: port N rules from <plugin-1>, <plugin-2>" --body "## Summary

   Port N rules from <plugin-1> and <plugin-2> to rslint.

   ### Rules ported
   | Rule | Plugin | Description | Doc |
   |------|--------|-------------|-----|
   | \`<rule-1>\` | <plugin-1> | [brief description] | [link](<url>) |
   | \`<rule-2>\` | <plugin-2> | [brief description] | [link](<url>) |
   | ... | ... | ... | ... |

   ## Checklist

   - [x] Tests updated (or not required).
   - [x] Documentation updated (or not required)."
   ```

   Examples:
   - `feat: port 5 rules from @typescript-eslint, eslint-plugin-import`
   - `feat: port 3 rules from ESLint core, @typescript-eslint`

   - **Do NOT include AI-related information** in PR title or body
   - If any rules were skipped during batch execution, note them in the PR body

---

## Post-Porting Validation (Optional)

For complex rules (rules involving scope tracking, autofix, or many configuration options), consider running a deeper alignment check after the initial port:

- Use the `validate-rule-alignment` skill to exhaustively verify edge cases
- Run the rule on real-world projects and compare output with the original ESLint rule
- This step is especially valuable for rules that track state across nested scopes (e.g., `this` binding, variable declarations)

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
