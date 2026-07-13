# Rslint Rule Porting Guide

## Role & Objective

You are an expert Software Engineer tasked with porting ESLint rules to `rslint`, a high-performance linter written in Go. Your goal is to implement the rule logic in Go, ensuring 1:1 parity with the original ESLint behavior, including all edge cases and error messages.

---

## Scope: rule semantics, not framework parity

Your job is porting the **rule's semantics** — given equivalent input, produce equivalent diagnostics. You are **not** responsible for re-implementing ESLint framework concepts that rslint deliberately does not expose. Examples:

- `/*eslint ...*/` directive comments
- `parserOptions.sourceType` override / `parserOptions.ecmaFeatures.*`
- `env: 'browser' | 'node' | ...`

Note: `languageOptions.globals` and `/*global ...*/` comments are automatically parsed by rslint and exposed through `ctx.Globals`. When porting rules that reference global variables, do not skip these test cases; instead, check `ctx.Globals` (e.g., `ctx.Globals[name]`).

When an upstream test case depends on other unsupported concepts (like `env` or `/*eslint*/` configurations):

- **Don't** reimplement the concept inside your rule.
- **Don't** list the gap under the rule's "Differences from ESLint" section — framework gaps apply to every rule, not yours.
- **Do** mark the upstream case `skip: true` with an inline reason such as `// SKIP: rslint does not support ESLint's <concept>`.

The rule doc's "Differences from ESLint" section is reserved for semantic differences of this specific rule — either intentional choices (Phase 1 Step 6.A) or tsgo/Go-semantic side effects (Phase 1 Step 6.B).

---

## Testing Philosophy

Porting is **re-implementation on a different substrate**, not translation. tsgo's AST diverges from ESTree in many small ways (parenthesized nodes are explicit, optional chain is a flag, literals split into multiple `Kind*Literal` kinds, numeric/string text is normalized at parse time, `AssignmentExpression` / `SequenceExpression` collapse into `BinaryExpression`); the type checker and scope manager are independent codebases with their own quirks. **Behavioral divergence between Go and ESLint is the default outcome; tests are the only mechanism that turns it into convergence.**

Three principles follow — internalize them before writing a single test case:

1. **Upstream's test suite is a floor, not a goal.** Migrating every `valid` / `invalid` case from the upstream test file proves only that you didn't miss a _documented_ behavior. It is **not** evidence the port is aligned. Upstream tests exercise the paths _upstream's authors_ found important on _their AST_; by construction they cannot cover the divergence your Go implementation introduces, because that divergence does not exist in their world.

2. **The augmentation IS the alignment work.** Every rule's tests are composed of three layers, all required, physically split across two files so the upstream-mirror and the rslint-added cases stay visually separated:

   | Layer                                      | What it covers                                                                                                                                                                                                       | Planned in     | Lives in                  |
   | ------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------- | ------------------------- |
   | 1. **Upstream migration**                  | Every upstream `valid` / `invalid` case (or `Skip: true` with reason)                                                                                                                                                | Phase 1 Step 3 | `<rule>_upstream_test.go` |
   | 2. **Edge-shape & real-user augmentation** | tsgo↔ESTree shape divergence on every child-node access (paren / optional-chain / literal-kind / type-wrapper / computed-key forms), plus real-user code shapes pulled from the upstream rule's GitHub issue tracker | Phase 1 Step 4 | `<rule>_extras_test.go`   |
   | 3. **Branch lock-ins**                     | A minimum-input test for every reachable branch in the upstream source — including branches upstream itself never tests                                                                                              | Phase 1 Step 5 | `<rule>_extras_test.go`   |

   The `_upstream_*` / `_extras_*` filename split is a contract — never mix migrated and rslint-added cases in the same file. See Phase 2 Step 4 for layout details and Phase 2 Step 1 for the split-when-too-large threshold.

3. **Green tests are necessary, not sufficient.** Before claiming alignment, the rule must additionally pass the Contract Alignment Checklist (Phase 4 Step 6) and — for any rule with non-trivial branching — a differential validation against the reference implementation on a real codebase (Phase 4 Step 8). Any divergence the differential run surfaces feeds back into layer 2 or 3 as a new locked-in test. A green Go suite alone proves only that the rule handles the inputs _you thought of_.

**Coverage bar.** The point of layers 2 + 3 is to prove the Go/tsgo port stays aligned where it structurally diverges from upstream's ESTree implementation — so the bar is _what they cover_, not how many cases they add up to. There is no case-count target. Concretely: every applicable Dimension 4 edge shape and ≥2 real-user shapes from the issue tracker (Phase 1 Step 4), plus every reachable branch locked in (Phase 1 Step 5). A near-empty `_extras_test.go` — or worse, none at all — is a reliable smell that Phase 1 Steps 4 and 5 were skipped: re-walk them before submitting. Phase 4 Step 6's per-layer checkboxes are what enforce this.

**JS tests are not a coverage layer — do not split them.** The three-layer model and the `_upstream_*` / `_extras_*` file split apply to **Go tests only**. The JS file `packages/rslint-test-tools/tests/.../<rule>.test.ts` exists for a different purpose: it spawns the compiled binary over IPC and verifies registration + wire protocol + ESLint-compatible diagnostic shape end-to-end. That contract is input-independent — running it against more cases doesn't verify it any better. So:

- **JS mirrors Layer 1 only** (upstream `valid` / `invalid` cases). Layers 2 and 3 stay in Go.
- A JS file far smaller than the Go suite — sometimes by 10× or more, depending on whether upstream uses fixture files — is the **expected** state, not "JS is under-tested." The semantic check is "every JS-asserted behavior also has a Go-upstream case"; literal case-count parity is **not** required (see Phase 4 Step 5).
- See Phase 3 Step 2 for what goes in the JS file and Phase 4 Step 5 for the alignment-direction check (JS ⊆ Go upstream, semantic).

---

## Related Documents

| Document                                 | Description                                                         |
| ---------------------------------------- | ------------------------------------------------------------------- |
| [AST_PATTERNS.md](AST_PATTERNS.md)       | AST traversal patterns, listeners, TypeChecker, reporting functions |
| [UTILS_REFERENCE.md](UTILS_REFERENCE.md) | Utility functions in `internal/utils/`                              |
| [QUICK_REFERENCE.md](QUICK_REFERENCE.md) | Commands cheatsheet, file locations, naming conventions, checklist  |

---

## Source Code Reference

Before starting, familiarize yourself with these key source locations:

### Core Infrastructure

| File/Directory                        | Description                                                                                                                                                |
| ------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `internal/rule/rule.go`               | **Core rule interface** - `Rule`, `RuleContext`, `RuleListeners`, `RuleMessage`, `RuleFix`, `RuleSuggestion` definitions                                   |
| `internal/rule/disable_manager.go`    | Logic for handling `// rslint-disable` and `// eslint-disable` comments                                                                                    |
| `internal/config/config.go`           | Registration orchestration and config loading. Per-rule registration data lives in each group's `all.go` — see Phase 3 Step 4 for where to add a new rule. |
| `internal/rule_tester/rule_tester.go` | Go test framework - `RunRuleTester`, `ValidTestCase`, `InvalidTestCase`                                                                                    |

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

3. **Collect Test Cases — Layer 1 (baseline migration)**:

   This step is the planning input for `<rule>_upstream_test.go` (Layer 1) — the _floor_ of the rule's overall test coverage. The augmentation in Phase 1 Steps 4 and 5 (which is the planning input for `<rule>_extras_test.go`, Layers 2 + 3) is what actually verifies alignment. See the [Testing Philosophy](#testing-philosophy) for why migration alone is insufficient.

   > **Phase 1 is planning, not writing.** Test files are physically created in Phase 2 Step 1 and populated in Phase 2 Step 4. In this phase you collect, organize, and annotate the cases — don't start a `_test.go` file yet.
   - Extract **ALL** `valid` and `invalid` cases from the official documentation.
   - Migrate **ALL** `valid` and `invalid` cases from the official unit test file (`tests/lib/rules/<rule>.js` for ESLint core; plugin equivalents otherwise) — not a representative subset.
   - **Skip with explanation**: If a case exercises an option or syntax we intentionally don't support, keep it in the file as a `Skip: true` test with a `// SKIP: <reason>` comment — don't drop it silently.
   - **Ensure Coverage**: Ensure Line and Column numbers are tested in invalid cases.

   **Do NOT stop here.** The migrated suite is a baseline; proceed to Phase 1 Step 4 (edge-shape augmentation) and Phase 1 Step 5 (branch lock-ins) — without them the port can pass every upstream test and silently diverge on real-user inputs.

4. **Identify Edge Cases — Layer 2 (edge-shape & real-user augmentation)**:

   This layer covers the divergence Go-on-tsgo introduces that upstream's tests cannot see — see [Testing Philosophy](#testing-philosophy). Without it, the port can pass every upstream test and silently drift on inputs upstream's contributors didn't write but real users do.

   Walk all four dimensions below. Dimension 4 (Universal Edge Shapes) is **non-skippable** — every applicable row needs ≥1 dedicated Go test (written in `<rule>_extras_test.go` during Phase 2 Step 4) marked `// ---- Dimension 4: <what> ----`, and rows that genuinely don't apply need an explicit `// N/A: <reason>` marker so future audits can verify the walk happened.

   Systematically enumerate edge cases across four dimensions.

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

   **Dimension 4: Universal edge shapes** — walk this checklist for EVERY port, regardless of what the rule does. Mark rows as N/A when genuinely irrelevant (and briefly note why), and add ≥1 dedicated test for each applicable row. Upstream's own test suite rarely covers all of these; they are the most common source of "looks aligned but silently drifts" regressions:
   - **Receiver / expression wrappers on inputs the rule inspects**:
     - `(X).y`, `((X)).y` — single and multi-level parenthesized receiver (tsgo preserves; ESTree flattens)
     - `X!.y` — TS non-null assertion
     - `(X as any).y`, `X satisfies T` — TS type-expression wrappers
     - `X?.y`, `X?.()` — optional chain (tsgo: flag on `PropertyAccessExpression`/`CallExpression`; no `ChainExpression` wrapper)
   - **Access / key forms**:
     - Identifier key vs string-literal key (`"x": ...`) vs numeric-literal key (`0: ...`) vs `PrivateIdentifier` (`#x`) vs `ComputedPropertyName` (`[expr]: ...`) — state explicitly which forms the rule accepts and lock every other form as an un-matched case
     - Element access `X['y']`, `X[`y`]`, `X[0]`, `X[Symbol.iterator]` when the rule handles dotted member access
   - **Declaration / container forms** (when the rule targets functions or classes):
     - Class declaration vs class expression (`class X extends Y` vs `const X = class extends Y`)
     - Function declaration vs function expression vs arrow vs method vs class-field arrow (`componentDidMount = () => {}`)
     - `async` / `generator` / `async generator` variants
   - **Nesting / traversal boundaries**:
     - Same-kind nesting where only the outer (or only the inner) should match — e.g. class-in-class, function-in-function. Verify the listener doesn't "bleed" past the boundary
     - Rule-specific ancestor walks (`getThisContainer`, `FindEnclosingScope`, etc.) crossed with arrow bodies, method bodies, and class-static-block bodies
   - **Graceful degradation**:
     - `SpreadAssignment` inside an object literal, `RestElement` inside a binding pattern — must not crash and must not mask sibling-property checks
     - Empty class body, empty function body, empty destructuring pattern, empty arguments list
     - Overload signatures / `abstract` / `declare` members — body-absent forms

   **Real-user shapes** (after walking Dimensions 1–4) — scan the upstream rule's GitHub issue tracker for closed regressions, false-positive reports, and false-negative reports. Convert ≥2 representative real-user code shapes into Go tests, marked `// ---- Real-user: <issue# or scenario> ----`. These are inputs production codebases produce that upstream's contrived test suite typically misses — and they're the inputs your rule will most likely face in real use. Do not skip this step on the grounds that "upstream's tests pass"; that is precisely the failure mode this layer prevents.

5. **Upstream Semantic Walk — Layer 3 (branch lock-ins)**:

   Migrating upstream's `valid`/`invalid` tests covers the main path, but nearly every ESLint rule has branches / OR conditions that are reachable but not tested upstream. Missing these is the #1 source of "passes all upstream tests, silently drifts in semantics" regressions — exactly the failure mode the [Testing Philosophy](#testing-philosophy) calls out.

   Do this walk BEFORE moving to Phase 2:
   1. Read the upstream rule source file end-to-end.
   2. For each listener / visitor, enumerate every branch — in particular:
      - Every `||` / `&&` in a gating `if`, including the ones whose second arm is reachable only by a specific input shape.
      - Every `.some()` / `.find()` / `entries().some()` predicate — each `return moduleName;`-style early-exit is a distinct branch.
      - Every fallback value (`X || defaultY`, `X ?? Y`) where `X` can realistically be undefined.
   3. For each branch, write down a MINIMAL input code snippet that exercises it.
   4. Add a Go test for every snippet, even if upstream never tests it. Typical examples that slip past upstream tests:
      - Destructuring from a non-`require` call whose first arg happens to match a watched module (e.g. `var {X} = myFunc('react')`).
      - Fallback `reactModuleName || pragma` when `reactModuleName` is falsy.
      - A condition that becomes true only for a TS-only syntax form (non-null, `as`, `satisfies`).

   These tests protect against future refactors silently flipping semantics. They live in `<rule>_extras_test.go` (Phase 2 Step 4); each case carries an inline comment referencing the upstream branch it locks in: `// Locks in upstream <fn>() arm <N>: <what>`.

6. **Document Divergence from ESLint**:

   Two classes of divergence may arise when porting. Both must be documented; they differ in _how_ and _where_.

   **A. Intentional divergence** — a choice we make (e.g. more precise error locations, different reporting granularity). Do all three:
   1. **Source code comment**: Add a `// NOTE: Unlike ESLint...` explaining the difference and rationale.
   2. **Rule documentation**: Add a "Differences from ESLint" section in the rule's `.md` file.
   3. **Test cases**: Ensure the differing behavior is covered by a dedicated test — a green-path `ValidTestCase` or a case with an exact `Message` / position assertion — so that future refactors can't silently flip it.

   **B. Language-natural divergence** — a side effect of tsgo's AST or Go semantics that we don't actively choose (e.g. tsgo decimal-normalizes `NumericLiteral` at parse time, so a dynamic computed key `[0x1]` compares equal to `[1]` where ESLint's token-level comparison sees them as distinct). Usually more permissive than ESLint.
   1. **Rule documentation** (or [AST_PATTERNS.md](AST_PATTERNS.md) if the quirk is general, not rule-specific): note the divergence under "Differences from ESLint" / the relevant AST-shape section.
   2. **Test cases**: Lock the current behavior in with a test — typically the ESLint-fails-but-we-pass case stays on the `valid` side with a comment pointing at the underlying quirk, so the behavior can't flip silently.

---

## Phase 2: Implementation (Go)

> **AST note**: rslint is built on the tsgo AST, which is structurally different from ESLint's ESTree. Child-access patterns (`node.left`, `node.argument`, `node.callee`, …) do **not** correspond 1:1: parentheses are explicit nodes, optional chains are flag-based (no `ChainExpression` wrapper), `Literal` is split across several `Kind*Literal` kinds, and `AssignmentExpression` / `SequenceExpression` collapse into `BinaryExpression`. Review [AST_PATTERNS.md § AST Shape Essentials](AST_PATTERNS.md#ast-shape-essentials) before implementing, and run the Alignment Audit (end of Step 2) before tests.
>
> **If you discover a new tsgo↔ESTree shape difference during porting** (e.g. a kind that has no ESTree analog, an `.Text` field that's normalized at parse time when ESLint sees raw source, an access pattern that requires an extra unwrap), **append it to [AST_PATTERNS.md § AST Shape Essentials](AST_PATTERNS.md#ast-shape-essentials) as part of your PR**. That file is the living knowledge base; every new rule is a chance to grow it.

### Step 1: Directory Setup

- **Core Rules**: `internal/rules/<rule_name_snake_case>/`
- **Plugin Rules**: `internal/plugins/<plugin_name>/rules/<rule_name_snake_case>/`

**Action**: Create the directory and the standard file set:

1. `<rule_name>.go` — Implementation
2. `<rule_name>.md` — Documentation
3. `<rule_name>_upstream_test.go` — Layer 1 tests (upstream 1:1 migration; see [Testing Philosophy](#testing-philosophy))
4. `<rule_name>_extras_test.go` — Layers 2 + 3 tests (edge-shape augmentation, real-user shapes, branch lock-ins)

The `_upstream_*` / `_extras_*` split is a hard contract: a reviewer can `ls` the directory and immediately see (a) that the rule has rslint-added augmentation at all and (b) which side of the fence each case lives on. **Never** mix migrated and rslint-added cases in the same file, and **never** put augmentation cases in `_upstream_*`.

**When to split further** — if `_extras_test.go` grows past roughly **80 cases or 600 lines**, partition by functional area and create one file per area:

- `<rule_name>_extras_dim4_test.go` — Dimension 4 universal-edge-shape rows
- `<rule_name>_extras_branches_test.go` — upstream-branch lock-ins
- `<rule_name>_extras_realuser_test.go` — issue-tracker shapes
- `<rule_name>_extras_<feature>_test.go` — option / mode / receiver type, etc.

Same threshold for `_upstream_test.go` if upstream itself partitions cleanly into feature subsets (e.g. one file per option mode). When upstream is also split, each subfile's header docstring should describe its own subset, not copy the whole-suite template (e.g. "TestRuleUpstreamCallbackArg migrates upstream's callback-arg test cases ...").

**Test function naming for area splits** — each split file gets one Test function whose name mirrors the area suffix in PascalCase: `<rule>_extras_dim4_test.go` → `TestRuleExtrasDim4`, `<rule>_extras_branches_test.go` → `TestRuleExtrasBranches`, `<rule>_upstream_callback_arg_test.go` → `TestRuleUpstreamCallbackArg`. This keeps a 1:1 file ↔ function mapping that `grep` can exploit.

For a worked example of large-rule splitting, see `internal/plugins/react_hooks/rules/exhaustive_deps/` (12 `upstream_*_test.go` + 5 extras files). **Important**: `exhaustive_deps` predates this convention and uses a hybrid naming pattern — some files keep the `<rule>_` prefix, others drop it; Test function names use `_`-separated snake (`TestExhaustiveDeps_Upstream_CallbackArg`) instead of the documented `Test<Rule><Suffix>` PascalCase. **New rules should follow the documented patterns above; reference `exhaustive_deps` only for _how_ to partition by feature, not for naming.**

### Step 2: Write Rule Logic

**File**: `<rule_name>.go`

**Prerequisites**:

- Read `internal/rule/rule.go` to understand core definitions
- Reference existing rules for the standard implementation pattern
- Review AST node types in `shim/ast/shim.go`
- See [AST_PATTERNS.md](AST_PATTERNS.md) for traversal patterns and examples

**Check plugin-local helpers FIRST** (before touching `internal/utils/`): grep the same plugin's neighbor rules for near-duplicates of the helper you're about to write:

```bash
# For plugin rules:
grep -rn "^func [a-z]" internal/plugins/<plugin>/rules/
# For core rules:
grep -rn "^func [a-z]" internal/rules/
```

If ≥1 rule in the same plugin already defines a near-equivalent helper, you MUST extract the shared helper to `internal/plugins/<plugin>/<plugin>util/` (or an existing shared package) BEFORE adding your new rule. No second copy. This is a hard rule — see _Helper Extraction_ below for the override criterion.

**Check for reusable `internal/utils/` helpers** (SECOND): Before writing any helper function, grep `internal/utils/` for an existing one. Helpful prefixes to search:

- `IsSpecific*`, `IsArgument*` — well-known API-call recognition (`Object.defineProperty`-style, member-access patterns, nth-argument-of)
- `GetStatic*`, `Normalize*` — property-name / literal-value normalization (e.g. `GetStaticPropertyName`, `NormalizeNumericLiteral`, `NormalizeBigIntLiteral`)
- `AreNodes*`, `IsSame*` — structural / reference AST comparison
- `GetFunction*`, `TrimmedNodeText*`, `TrimNodeTextRange` — function head / trimmed source text
- `IsShadowed`, `FindEnclosingScope`, `CollectBindingNames` — scope / binding queries
- `GetOptionsMap` — options parsing (handles both array and map inputs)
- **Type-aware queries** (for `@typescript-eslint` rules that use `ctx.TypeChecker`): `Is*Type*` / `Get*Type*` — type-flag tests and classifications (`IsTypeAnyType`, `IsUnionType`, `GetTypeName`, `GetContextualType`, `GetConstraintInfo`); `IsPromise*` / `IsError*` / `IsReadonly*` — builtin-type detection; `NeedsToBeAwaited`, `GetCallSignatures`, `CollectAllCallSignatures` — signature / awaitability helpers; `IsUnsafeAssignment`, `DiscriminateAnyType` — any-type safety. See the `ts_api_utils.go` / `ts_eslint.go` / `builtin_symbol_likes.go` sections of [UTILS_REFERENCE.md](UTILS_REFERENCE.md) for the complete inventory — **do not re-implement type analysis inline**.

See [UTILS_REFERENCE.md](UTILS_REFERENCE.md) for the full inventory. **If you find a near-match that's missing some behavior, extend it in place** rather than writing a parallel implementation inline. Extraction is explicitly preferred over duplication (see _Helper Extraction_ below for criteria).

**Check for reusable shim utilities** (THIRD): If `internal/utils/` has nothing, check if the `shim/` packages already provide what you need:

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
    Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
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
    Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
        // ...
    },
})

// For ESLint Core rules:
var MyCoreRule = rule.Rule{
    Name: "my-core-rule",
    Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
        // ...
    },
}
```

**Key Points**:

- `RuleListeners` is a map from `ast.Kind` to a callback function
- Each callback receives a `*ast.Node` and reports diagnostics via `ctx.ReportNode()`
- Options parsing happens inside the `Run` function before returning listeners
- Use `rule.CreateRule` **ONLY** for `@typescript-eslint` rules (it adds the prefix)
- **`RequiresTypeInfo`**: If a `@typescript-eslint` rule uses `ctx.TypeChecker`, you **MUST** set `RequiresTypeInfo: true`. This tells the linter to skip the rule on files without a type checker, preventing nil-pointer panics. Core ESLint rules should NOT set this flag — use `ctx.TypeChecker == nil` guards instead (see [AST_PATTERNS.md — Using TypeChecker](AST_PATTERNS.md#using-typechecker)).
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

`Run` receives `options []any` — ESLint's `context.options` array (the configured options after the severity level; empty when none were configured). Write `parseOptions` to take that slice directly and extract the first element's map with `utils.GetOptionsMap()`:

```go
Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
    opts := parseOptions(options)
    // ...
}

func parseOptions(options []any) Options {
    opts := Options{/* defaults */}
    optsMap := utils.GetOptionsMap(options)
    if optsMap != nil {
        // Parse options from optsMap...
    }
    return opts
}
```

`GetOptionsMap` is the only safe extractor — do not reimplement it with a hand-rolled `options[0].(map[string]interface{})` type assertion.

For a rule with multiple positional options (e.g. `["error", "both", {...}]`), index `options` directly (`options[0]`, `options[1]`, ...) instead of using `GetOptionsMap`.

#### Options schema

Every new rule declares a JSON Schema for its options on the `Schema` field. The linter validates each configured rule's options against it before linting starts (in the CLI, as a separate fail-fast step right after configuration is resolved), so a misconfigured rule fails with a clear error instead of being silently misread.

- **No options**: set `Schema: rule.EmptyArraySchema`. Always reference the shared value — never author your own copy of the empty-array schema.
- **With options**: copy ESLint's `meta.schema` into a `<rule-name>.schema.json` file beside the rule source and embed it. When upstream's `meta.schema` is a plain **array** of item schemas, wrap it the way ESLint itself does: `{"type": "array", "items": <the upstream array, used directly as the tuple items>, "minItems": 0, "maxItems": <len>}`. When it's already a full schema **object** (e.g. eqeqeq's top-level `anyOf`), copy it as-is.

```go
import _ "embed"

//go:embed my-rule.schema.json
var schemaJSON []byte

var MyRule = rule.Rule{
    Name:   "my-rule",
    Schema: rule.NewSchema(schemaJSON),
    Run:    /* ... */,
}
```

Schemas are JSON Schema Draft 4 (the draft ESLint itself uses) and compile lazily on first use; the CI sweep `TestAllRules_DeclaredSchemasCompile` (internal/config) catches a schema that fails to compile. Validation also fills schema `default` values into the options in place, matching ajv's `useDefaults` as ESLint configures it (cross-checked against ajv@6 by `TestValidateMatchesAjvFixtures`), so an option object a user partially fills in arrives at the rule with its schema defaults present. Keep `parseOptions` handling defaults anyway: like in ESLint, a default only lands when the user supplied the enclosing options object at all (a grown tuple slot never reaches the rule), and API/LSP entry points don't run the validation step.

### Alignment Audit

Before moving on, walk through each check. Each one targets a class of AST-shape bug that is not caught by compilation and may slip past narrowly-written unit tests. Skip a row when it doesn't apply to your rule.

| If the rule …                                                                   | Audit                                                                                                                                                             | Reference                                                                            |
| ------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------ |
| Reads a child node (every rule)                                                 | Any `.Kind ==` / `.Kind !=` / `.As<Type>()` / `.Text` access on a child must go through `ast.SkipParentheses` first (directly or via a helper).                   | [AST_PATTERNS.md § ParenthesizedExpression](AST_PATTERNS.md#parenthesizedexpression) |
| Handles `foo?.bar` / `foo?.()`                                                  | Use `ast.IsOptionalChain(node)`; don't hand-check node flags.                                                                                                     | [AST_PATTERNS.md § Optional Chain](AST_PATTERNS.md#optional-chain)                   |
| Compares literal values                                                         | Match the precise `Kind*Literal`; normalize numeric text via `utils.NormalizeNumericLiteral` before value comparison.                                             | [AST_PATTERNS.md § Literal Kinds](AST_PATTERNS.md#literal-kinds)                     |
| Has separate ESLint listeners for `AssignmentExpression` / `SequenceExpression` | Collapse into one `BinaryExpression` listener and branch on `OperatorToken.Kind`.                                                                                 | [AST_PATTERNS.md § Binary Operator Kinds](AST_PATTERNS.md#binary-operator-kinds)     |
| Emits fix/suggestion text starting with an identifier                           | Guard against token fusion with the preceding character before emitting (otherwise e.g. `typeof` + `Number(foo)` becomes `typeofNumber(foo)`).                    | —                                                                                    |
| Checks whether a name resolves to a global                                      | Use `utils.IsShadowed(node, name)`. Note: stricter than ESLint's scope manager on TS type-only bindings — document in the rule's `.md` if the difference matters. | —                                                                                    |
| Reads source text for recommendation / fix                                      | Prefer `utils.TrimmedNodeText(sf, node)` (skips leading trivia) over raw `node.Pos()/End()`.                                                                      | [AST_PATTERNS.md § Node Text and Positions](AST_PATTERNS.md#node-text-and-positions) |

### Helper Extraction

After Step 2 is done, review each helper for extraction to `internal/utils/`:

**Extract if all hold:**

- Input/output is AST- or source-oriented (not encoding the rule's own semantics)
- The name reads sensibly without context of the current rule
- Another rule would plausibly need the same thing

**Keep local otherwise.** Predicates that encode a specific rule's definition (e.g. a `isDoubleLogicalNegating`-style helper that codifies "what counts as a double-negation coercion for THIS rule") stay with the rule — extracting would mislead future readers.

**Hard override — duplicate-across-rules rule**: if the same helper (or a near-duplicate) already lives in ≥1 other rule within the same plugin, it MUST be extracted to `<plugin>util/`, even if the "plausibly needed by another rule" criterion above feels borderline. The fact that you're about to write the second copy is itself proof of reusability. Don't let the first duplicate bend your judgement.

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

Examples of **incorrect** code for this rule with `{ "someOption": true }`:

```json
{ "<rule-name>": ["error", { "someOption": true }] }
```

```javascript
// Example
```

## Original Documentation

[Link to ESLint documentation]
````

**Options in examples**: when a code block demonstrates a specific option combination, precede the `javascript` block with a standalone `json` block containing the rule's config entry — shape: `{ "<rule-name>": ["error", { ...options... }] }`. Let prettier format it (single-line when short, multi-line when the options list grows). Keep the `javascript` block pure source code (no annotations). Do **not** wrap the config entry in a `"rules": { ... }` object (redundant here) and do **not** copy upstream linter directives such as `/* eslint <rule>: [...] */` into the examples.

**Writing a "Differences from ESLint" section** (when the rule has one):

- The audience is the **rule user**, not the porter. Describe what they will observe, not why.
- Each bullet states a concrete input pattern and the observable difference ("rslint reports X on this code; ESLint does not", "positions differ by N columns", "message text differs", etc.). Keep each bullet to ≤2 lines.
- **Do NOT** mention implementation details: `getText`, `SkipParentheses`, `AST shape`, `ESTree vs tsgo`, "we chose to…" — those belong in source-code comments, not the rule doc.
- If you can't explain the divergence in terms of observable input-vs-output behavior without reaching for mechanism, the divergence is probably a bug, not a documented difference. Reconsider.

### Step 4: Write Go Tests

**Files** (per Phase 2 Step 1): `<rule_name>_upstream_test.go` and `<rule_name>_extras_test.go`. The two-file split is the physical embodiment of the [Testing Philosophy](#testing-philosophy) — a reviewer should be able to `ls` the rule directory and immediately tell whether the augmentation work was done.

**Layer-to-file mapping:**

| Layer                                  | File                      | Test function                                                                    | In-file group markers (on the case directly)                                                                                                                                                                                                                                                                                                                                      |
| -------------------------------------- | ------------------------- | -------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1. Upstream migration                  | `<rule>_upstream_test.go` | `Test<Rule>Upstream`                                                             | `// ---- <upstream group name> ----` — preserve upstream's grouping verbatim so a top-to-bottom read matches the upstream test file                                                                                                                                                                                                                                               |
| 2. Edge-shape & real-user augmentation | `<rule>_extras_test.go`   | `Test<Rule>Extras`                                                               | `// ---- <description> ----` on each case (free-form descriptive text, as used by existing jsx-a11y rules). For new rules, prefer the prefix-tagged forms `// ---- Dimension 4: <what> ----` and `// ---- Real-user: <issue# or scenario> ----` because they let `grep` find a category quickly; both styles are accepted. `// N/A: <reason>` for rows that genuinely don't apply |
| 3. Branch lock-ins                     | `<rule>_extras_test.go`   | `Test<Rule>Extras` (or a separate `Test<Rule>ExtrasBranches` if extras is split) | `// Locks in upstream <fn>() arm <N>: <what>` on each case                                                                                                                                                                                                                                                                                                                        |

Layers 2 + 3 — not case count — are the real alignment work; there is no numeric target. A near-empty `_extras` file is a smell that Phase 1 Steps 4 and 5 were skipped (Phase 4 Step 6's per-layer checkboxes enforce this).

**File-header docstring** — open each test file with a top-of-file comment that names what the file is for and points at its sibling:

- `_upstream_test.go`: `// Test<Rule>Upstream migrates the full valid/invalid suite from upstream <upstream test path> 1:1. Position assertions cover line/column for every invalid case. rslint-specific lock-in cases live in the <rule>_extras_*_test.go file(s).`
- `_extras_test.go`: `// Test<Rule>Extras locks in branches and edge shapes that the upstream test suite doesn't exercise. Each case carries an inline comment pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers, so future refactors can't silently regress them without breaking a named lock-in.`

These docstrings are how a reader (or `grep`) confirms a file is doing its assigned job.

**Reference examples** in `internal/plugins/jsx_a11y/rules/`:

- Standard two-file rule: `anchor_ambiguous_text/`, `lang/`, `aria_role/`
- Large-rule split (further partitioned by area): `internal/plugins/react_hooks/rules/exhaustive_deps/`

**Conventions:**

- Use `rule_tester.RunRuleTester` in each test file (one `Test<Name>` function per file is typical; multiple are fine when it improves grouping).
- Shared fixtures (option-map literals, expected message strings) can live as package-level vars; both files share the same Go package so they compose freely.
- Invalid cases **MUST** include `Line` and `Column` assertions.
- Use `map[string]interface{}` to pass options in Go tests.
- Ensure `tsconfig.json` path uses `fixtures.GetRootDir()`.

**Options coverage — MUST exercise the JSON path.** Passing a typed struct directly (e.g. `Options: MyRuleOptions{CheckX: utils.Ref(true)}`) short-circuits the `options.(MyRuleOptions)` type assertion and never exercises `utils.GetOptionsMap` or JSON round-trip. CLI and JS configs always take the JSON path, so a struct-only suite leaves the CLI-facing wiring untested.

For every option your rule accepts, include **at least one** Valid case and **at least one** Invalid case whose `Options` field is `map[string]interface{}{...}` (bare object — matches the single-option CLI shape) or `[]interface{}{map[string]interface{}{...}}` (array-wrapped — matches the multi-element / rule_tester shape). This catches bugs like missing `GetOptionsMap` integration, wrong JSON tag casing, and option-name typos that typed structs silently hide. See `no_floating_promises_test.go → TestNoFloatingPromisesOptionParsing` for a reference suite covering both shapes, nil options, empty arrays, malformed values, and nested specifier arrays.

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

### Step 1: Check & Setup Test Environment

**Goal**: Ensure the test directory and necessary configuration files exist.

1. **Check Directory**: Verify if `packages/rslint-test-tools/tests/<plugin-name>` exists.

2. **Check Configuration**:
   - **Reference**: Use `packages/rslint-test-tools/tests/eslint-plugin-jsx-a11y` as the template (ESM flat-config format).
   - **Required Files**:
     - `rslint.config.mjs` (Configuration for the linter — JSON `rslint.json` is legacy and the loader auto-migrates it to JS/TS; do NOT create new `rslint.json` files)
     - `tsconfig.files.json` (TS Config for file-based tests)
     - `tsconfig.virtual.json` (TS Config for virtual/code-based tests)
   - **Plugin Configuration**: In `rslint.config.mjs`, set the `plugins` field (use the short plugin name, matching how rules are addressed in tests):
     - **Core Rules**: `plugins: []`
     - **Plugin Rules**: `plugins: ['<short-name>']` (e.g. `'jsx-a11y'`, `'jest'`, `'react'`, `'promise'`)
   - **Warning**: When copying `rule-tester.ts`, remove any hardcoded rule prefixes (e.g., `ruleName = 'jsx-a11y/' + ruleName;`).

### Step 2: Add JS Tests

**Purpose & scope.** The JS suite is **not** a duplicate of the Go suite. It spawns the compiled binary over IPC and verifies registration + wire protocol + ESLint-compatible diagnostic shape — a contract that is input-independent. So the JS file **mirrors Layer 1 only** (the upstream `valid` / `invalid` cases). Layer 2 (edge-shape & real-user augmentation) and Layer 3 (branch lock-ins) live exclusively in `<rule>_extras_test.go` on the Go side and **must not** be copied into the JS file. See [Testing Philosophy](#testing-philosophy) for the rationale.

**Practical rule:** the JS file should assert exactly the upstream `valid` / `invalid` semantic set — nothing less, nothing more. Case **counts** between JS and `<rule>_upstream_test.go` may legitimately differ (one side may inline what the other folds into a fixture file); the contract is semantic-subset equivalence, not numeric parity. If you find yourself reaching for a tsgo-specific edge shape, a Dimension 4 row, a branch lock-in, or a GitHub-issue real-user shape while writing the JS file — stop. Those belong in Go extras.

**File Locations** (determined by Phase 1 Step 2):

- **Core ESLint Rules** (including deprecated typescript-eslint rules): `packages/rslint-test-tools/tests/eslint/rules/<rule-name>.test.ts`
- **typescript-eslint Rules**: `packages/rslint-test-tools/tests/typescript-eslint/rules/<rule-name>.test.ts`
- **Other Plugin Rules**: `packages/rslint-test-tools/tests/<plugin-name>/rules/<rule-name>.test.ts`

**Setup RuleTester**:

- **Core ESLint Rules**: Import `RuleTester` from `../rule-tester` (in `tests/eslint/rule-tester.ts`, no prefix)
- **typescript-eslint Rules**: Import `RuleTester` from `@typescript-eslint/rule-tester` (auto-prefixes with `@typescript-eslint/`)
- **Other Plugin Rules**: Refer to `packages/rslint-test-tools/tests/eslint-plugin-jsx-a11y/rule-tester.ts`

**Options Format**: JS tests use array format: `options: [{ allow: ['warn'] }]`

### Step 3: Register Test File

**File**: `packages/rslint-test-tools/rstest.config.mts`

Add the new test file path to the `include` array.

### Step 4: Register Rule

**Where to add depends on rule type** (determined by Phase 1 Step 2):

| Rule type                                              | File to edit                         | What to add                                                                |
| ------------------------------------------------------ | ------------------------------------ | -------------------------------------------------------------------------- |
| Core ESLint (incl. deprecated typescript-eslint rules) | `internal/rules/all.go`              | Import the rule package; append `package.RuleNameRule` to `GetAllRules()`. |
| typescript-eslint (active)                             | `internal/plugins/typescript/all.go` | Same — append to that plugin's `GetAllRules()`.                            |
| Other plugins (react, jest, import, jsx-a11y, …)       | `internal/plugins/<plugin>/all.go`   | Same.                                                                      |

Each `all.go` exports a `GetAllRules() []rule.Rule` slice. `RegisterAllRules()` in `internal/config/config.go` iterates each slice and calls `GlobalRuleRegistry.Register(rule.Name, rule)` — **do not edit `config.go` for a new rule**.

**Registration key vs `rule.Name` must match** — the registrar uses `rule.Name` as the key. How that key is produced depends on the rule wrapper:

- **Core rule** — `rule.Rule{Name: "no-debugger", ...}` registers as `"no-debugger"`.
- **typescript-eslint rule** — `rule.CreateRule(rule.Rule{Name: "no-shadow", ...})` registers as `"@typescript-eslint/no-shadow"`. The factory auto-prefixes; **only** use it for `@typescript-eslint/` rules — using it on a core or other-plugin rule will silently mis-register the rule key.
- **Other plugins** — `rule.Rule{Name: "react/jsx-key", ...}` — the prefix is part of the literal `Name`, no factory.

**Do NOT register a rule under both `"rule-name"` and `"@typescript-eslint/rule-name"`** — pick the canonical one based on deprecation status.

---

## Phase 4: Verification & Build

**Goal**: Ensure the compiled binary runs the rule correctly.

Follow this **strict order** — each step depends on the previous one:

1. **Go formatting** (catches indentation issues early):

   ```bash
   gofmt -l internal/rules/<rule_name>/
   ```

   If files are listed, run `gofmt -w` on them to fix.

2. **Go tests** (the package-level invocation runs every `*_test.go` in the rule directory — both `_upstream_test.go` and `_extras_test.go`, plus any further `_extras_<area>_test.go` splits):

   ```bash
   go test -count=1 ./internal/rules/<rule_name>
   # or, for plugin rules:
   go test -count=1 ./internal/plugins/<plugin>/rules/<rule_name>
   ```

   **Related-rule regression**: if this port introduced or modified any exported symbol in a shared package (e.g. `internal/plugins/<plugin>/<plugin>util/`, or `internal/utils/`), you MUST also run tests for the changed package and the direct consumer packages that import or call the changed API. Keep the scope related to the changed Go code; do not run whole-plugin or whole-tree Go tests as part of the port-rule workflow.

   ```bash
   go test -count=1 <changed-package-dir> <direct-consumer-package-dir>
   ```

   Extracting / renaming a helper is a silent-regression hotspot; running only the new rule package is not enough when another package consumes the helper. Identify direct consumers with `rg` / `git grep`, run their package tests, and do not fall back to `go test ./internal/...`, `go test ./internal/plugins/<plugin>/...`, or `pnpm run test:go`.

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

5. **Verify Go ↔ JS Alignment** (asymmetric — JS is a Layer-1 semantic subset of Go):

   The two suites have asymmetric roles (see [Testing Philosophy](#testing-philosophy) and Phase 3 Step 2):
   - **JS suite** = Layer 1 mirror only. It exists to verify the binary, registration, and wire protocol — not rule logic.
   - **Go suite** = Layer 1 + Layer 2 + Layer 3 (full coverage). It is the source of truth for rule behavior.

   Two checks:
   - [ ] **JS ⊆ Go upstream (semantic)**: every behavior asserted by a JS case is also asserted somewhere in `<rule>_upstream_test.go`. The match is **semantic**, not literal — Go may legitimately split one fixture-driven upstream case into many inline cases, or the reverse. If JS asserts a behavior that has no corresponding Go-upstream case, the upstream migration is incomplete — fix Go.
   - [ ] **JS contains no Layer 2 / 3 cases**: review the JS file's contents (not its case count) for tsgo-specific edge shapes (Dimension 4 rows), branch lock-ins, or GitHub-issue real-user shapes. If any are present they leaked from Go extras — move them out.

   > **Do not use literal case-count equality as the alignment check.** It only happens to match when both sides are written from the same inline-case template (e.g. `lang` is 19=19, `anchor-ambiguous-text` 39=39, `aria-role` 38=38). For the majority of jsx-a11y rules the counts legitimately differ — `no_static_element_interactions` is 644 (Go upstream) vs 135 (JS), `aria_props` is 12 vs 99 — because upstream uses fixture files that one side expands and the other folds. Both are correct as long as the semantic-subset check above holds.

   Layer 2 and Layer 3 cases stay in Go only. Do **not** add them to the JS file even if "for completeness" feels tempting — see Phase 3 Step 2 Purpose & scope for why.

   **Go vs JS test differences**:

   | Aspect           | Go tests                                | JS tests                                    |
   | ---------------- | --------------------------------------- | ------------------------------------------- |
   | Autofix          | `Output: []string{...}` field           | Not verified (snapshot filters out `fixes`) |
   | Position         | `Line`/`Column` fields on each error    | Implicitly covered by snapshot              |
   | Multiple errors  | `Errors: []...{{...}, {...}}`           | `errors: [{...}, {...}]`                    |
   | MessageId format | camelCase (e.g., `"noLossOfPrecision"`) | camelCase (e.g., `"noLossOfPrecision"`)     |

6. **Contract Alignment Checklist (Go ↔ ESLint)**:

   Phase 4 Step 5 verifies our two test suites agree with each other. This step verifies the **public contract** of the rule agrees with ESLint. The oracle is ESLint's diagnostic output (`messageId` + message text + report position) and its options schema — **not** ESLint's internal implementation. Language-level implementation differences are acceptable (see Phase 1 Step 6.B); contract differences are not.

   Before claiming the port is aligned, confirm every row. Missing any row means the claim is premature.

   **File split** (each layer has a designated file — see [Testing Philosophy](#testing-philosophy) and Phase 2 Step 4):
   - [ ] **Two files exist**: `<rule>_upstream_test.go` and `<rule>_extras_test.go` (or area-split variants `<rule>_extras_<area>_test.go` if the rule is large).
   - [ ] **Header docstrings present**: each file's top-of-file comment names what the file is for and points at its sibling.
   - [ ] **Split contract honored**: `_upstream_*` files contain only migrated upstream cases; `_extras_*` files contain only rslint-added cases. No mixing.

   **Coverage layers**:
   - [ ] **Layer 1 — Upstream migration complete** (in `_upstream_test.go`): every `valid` / `invalid` case from the upstream unit-test file has a corresponding Go case (or `Skip: true` + `// SKIP: <reason>`).
   - [ ] **Layer 2 — Edge-shape augmentation present** (in `_extras_test.go`): Phase 1 Step 4 Dimension 4 walked row-by-row; every applicable row has ≥1 dedicated Go test marked `// ---- Dimension 4: <what> ----`; N/A rows carry an explicit `// N/A: <reason>` marker so the walk is auditable.
   - [ ] **Layer 2 — Real-user shapes present** (in `_extras_test.go`): ≥2 cases pulled from the upstream rule's GitHub issue tracker (closed regressions / FP / FN reports), marked `// ---- Real-user: <issue# or scenario> ----`.
   - [ ] **Layer 3 — Branch lock-ins present** (in `_extras_test.go`): every reachable branch in the upstream source has a minimum-input Go test marked `// Locks in upstream <fn>() arm <N>: <what>`, including branches upstream itself never tests.
   - [ ] **Extras aren't a token gesture**: with the layers 2 + 3 boxes above checked, step back and confirm `_extras_*` substantively exercises the rule's divergence surface — not one perfunctory case per layer. There is no case-count target; a near-empty `_extras_*` is a smell to re-walk Phase 1 Steps 4 and 5, not a number to hit.

   **Diagnostic contract** (each invalid output is exactly what ESLint emits):
   - [ ] **Message text assertions**: each `messageId` has ≥1 test using the `InvalidTestCaseError.Message` field (exact string match), covering every modifier combination the rule can emit (`static`, `private`, `async`, computed-no-name, etc.).
   - [ ] **Position assertions per container**: for each container the rule emits into (object literal / class / type / descriptor / …), ≥2 cases assert `Line` + `Column` + `EndLine` + `EndColumn`, including one multi-line case.

   **Options contract**:
   - [ ] **Schema match**: option names, types, and **defaults** match ESLint's schema exactly. Assert every default by running a case with no options vs. `[{}]` options and confirming identical output.
   - [ ] **Combination matrix**: for every boolean option, include ≥1 test where it is `true` and ≥1 where it is `false`. Triggering combinations (rule behaves differently when two options are both on) get dedicated cases.

   **Equivalence classes** (when applicable):
   - [ ] **Three-way equivalence classes** (if the rule compares names / keys): static / private / dynamic keys form separate equivalence classes; test ≥1 cross-class negative (e.g. `'#a'` string vs `#a` private identifier should NOT pair up).

7. **Pre-commit gate** (BLOCKING — must all pass before Phase 5):

   ```bash
   # Type check and lint (JS/TS side)
   pnpm typecheck && pnpm lint

   # Spell check (catches typos in comments and strings)
   pnpm -w run check-spell

   # Format check
   pnpm format:check

   # Go lint (packages containing changed Go files only)
   changed_go_dirs="$(
     {
       git diff --name-only --diff-filter=ACMR origin/main...HEAD -- '*.go'
       git diff --name-only --diff-filter=ACMR --cached -- '*.go'
       git diff --name-only --diff-filter=ACMR -- '*.go'
       git ls-files --others --exclude-standard -- '*.go'
     } | sort -u | grep -E '^(cmd|internal)/' | while IFS= read -r file; do dirname "$file"; done | sort -u
   )"
   if [ -n "$changed_go_dirs" ]; then
     printf '%s\n' "$changed_go_dirs" | xargs golangci-lint run --new-from-rev=origin/main --timeout=10m
   fi
   ```

   These are BLOCKING. If any fails, fix before moving on — **do not** commit, push, or open a PR with any of them red.
   - **Go lint scope**: lint only packages containing changed `.go` files under `cmd/` and `internal/` during port-rule pre-commit verification, with `--new-from-rev=origin/main` so only issues introduced by the branch are reported. Do not run `pnpm lint:go` here; it lints the full `cmd/` and `internal/` trees and is reserved for explicit full-tree checks / CI. Do not pass changed files from multiple directories to one `golangci-lint run` invocation; named file arguments must all be in one directory and can also produce typecheck false positives when a file depends on sibling files.
   - **Unknown-word failures from `check-spell`**: inspect each reported word in context before changing anything. Fix misspellings, invented words, or other accidental text in the source. Add a word to `scripts/dictionary.txt` only when it is intentional: a valid standard word, ESLint ecosystem identifier, Go module/package name, API name, or similar technical token. Use the original case. Do not add `cspell` ignore comments in Markdown files; they can cause documentation compilation failures.
   - **Format failures**: auto-fix (`pnpm format && pnpm format:go`); never silence.
   - **Lint failures**: fix the code. Don't bypass with `//nolint`, `// eslint-disable`, or equivalent, unless the exception is already justified by an in-file comment pattern this repo uses.

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

   | Category                            | What it means                                                                                                                                                                                   | Action                                                                                                                  |
   | ----------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- |
   | **(a) Language-natural divergence** | tsgo AST or Go-semantic effect we don't actively choose (see Phase 1 Step 6.B — e.g. `NumericLiteral` parse-time normalization, normalized string cooked values).                               | Document under the rule's `.md` "Differences from ESLint" (or in [AST_PATTERNS.md](AST_PATTERNS.md) if general). Leave. |
   | **(b) Scan-scope divergence**       | The two tools see different file sets (e.g., rslint respects `.gitignore` by default; ESLint does not; tsconfig `include` excludes a dir). Not a rule issue.                                    | No action. Optionally note in the PR description if a reviewer might be confused.                                       |
   | **(c) Genuine bug**                 | Neither (a) nor (b). Rule logic, message text, or position is actually wrong on our side (or, rarely, ESLint's — but we align to ESLint unless we have a standing Phase 1 Step 6.A divergence). | **Must fix** before merging. Re-run the diff until it clears or reduces to (a)/(b) only.                                |

---

## Phase 5: Submission & PR

### Phase 5A: Per-Rule Commit

**Goal**: Commit each rule independently after it passes verification (Phase 4).

This step is executed **after each rule's Phase 4 completes** (both in single-rule and batch mode).

1. **Configure Project Settings (Conditional)**:
   - If the rule's plugin is already in the repo-root `rslint.config.ts` `plugins`, add the rule with `'warn'` severity
   - Otherwise, do NOT modify `rslint.config.ts`

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

   **Note**: The `--body` templates below follow the repo's PR template (`.github/PULL_REQUEST_TEMPLATE.md`). `gh` only auto-fills that template when `--body` is omitted, so the explicit bodies here take its place.

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
2. **Is the rule registered?** Check the appropriate `all.go` (`internal/rules/all.go` for core, `internal/plugins/<plugin>/all.go` otherwise) — confirm both the package import and the entry in `GetAllRules()` are present.
3. **Are test files included?** Check `rstest.config.mts`
4. **Is the test-dir `rslint.config.mjs` configured?** Ensure the plugin is listed and the rule is enabled
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

- [AST_PATTERNS.md](AST_PATTERNS.md) - AST traversal, listeners, reporting functions, fix helpers
- [UTILS_REFERENCE.md](UTILS_REFERENCE.md) - Utility functions in `internal/utils/`
- [QUICK_REFERENCE.md](QUICK_REFERENCE.md) - Commands, file locations, naming conventions, checklist
