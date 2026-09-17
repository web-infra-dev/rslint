# Rule utility integration

The work is split into three independent pull requests. Reuse rslint and tsgo
capabilities, preserve each rule's upstream semantics, and keep plugin policy in
the owning plugin. Each implementation must trace its consumers and run their
targeted tests. Correct existing bugs exposed by the extraction against upstream
behavior and add regressions instead of preserving accidental compatibility.
This plan does not request whole-repository or whole-plugin suites.

## 1. Reference tracking and computed property names

Status: implemented in this PR on `refactor/shared-reference-tracker`.

- Extract the shared alias, assignment, destructuring, and global-reference
  traversal from `nodeutil`, using `RuleContext.Refs`, the existing reference
  index, file cache, static evaluator, and tsgo binding helpers.
- Share computed property-name evaluation without changing the contracts of
  existing literal-only property-name helpers.
- Migrate Node consumers and `unicorn/no-document-cookie`; preserve diagnostic
  ranges, counts, global shadowing, cycle handling, and source-only behavior.
- Keep Node module loading, strict versus legacy ESM behavior, configuration,
  and diagnostics in `nodeutil` and the rules.
- Document the shared APIs and verify the affected consumers. Further core-rule
  migrations require their own semantic comparison.

The shared tracker also corrects global shorthand assignment aliases such as
`({require} = globalThis); require('fs')` and
`({Buffer} = globalThis); new Buffer()`. A shorthand property's binder symbol
does not declare a variable and must not hide the global alias. The fix uses
the existing declaration predicate and reference store; no compatibility mode
or alternate symbol resolver is needed.

Validation:

- `go test ./internal/utils -run '^TestStaticStringEvaluatorPropertyNames$'`
- `go test ./internal/utils/referencetracker ./internal/plugins/node/rules/no_deprecated_api ./internal/plugins/node/rules/no_restricted_require ./internal/plugins/node/rules/no_extraneous_require ./internal/plugins/node/rules/no_path_concat ./internal/plugins/unicorn/rules/no_document_cookie`
- `golangci-lint run --new-from-merge-base=origin/main ./internal/utils ./internal/utils/referencetracker ./internal/plugins/node/nodeutil ./internal/plugins/node/rules/no_deprecated_api ./internal/plugins/node/rules/no_restricted_require ./internal/plugins/node/rules/no_extraneous_require ./internal/plugins/node/rules/no_path_concat ./internal/plugins/unicorn/rules/no_document_cookie`
- A native CLI comparison over 320 JavaScript/TypeScript inputs preserved all
  369 existing diagnostics, including their messages, ranges and multiplicity.
  Only the two confirmed shorthand-alias diagnostics were added. The corpus
  enabled `no-document-cookie`, `no-deprecated-api`, `no-restricted-require`
  and `no-path-concat`; `no-extraneous-require` has a package regression.
- Probed the shorthand aliases with eslint-utils 4.10.1, and the affected
  TypeScript binding shapes with Unicorn 72.0.0 and the TypeScript ESLint parser.
  Permanent tests cover all five rule consumers and shared traversal contracts.

## 2. Call matching and fix ranges

Status: pending.

- Extract reusable call/member matching from `unicornutil`, retaining explicit
  choices for computed properties, optional chains, TypeScript wrappers, spread
  arguments, and plugin defaults.
- Promote framework-independent argument and type-argument range helpers from
  `test_framework`; reuse existing token, comment, parentheses, and `RuleFix`
  primitives for compound edits.
- Migrate existing consumers with equivalent contracts and add regressions for
  comments, grouping, and automatic semicolon insertion where affected.

## 3. Side-effect analysis and static-value queries

Status: pending.

- Extract a shared side-effect predicate from the existing `prefer-ternary`
  implementation after comparing the upstream operations and options needed by
  consumers.
- Keep getter and implicit-conversion policies explicit. Preserve the separate
  contracts of control-flow safety and `no-unused-expressions` checks.
- Extend the existing static evaluator with the read-only queries required by
  concrete rules, preserving unknown values and JavaScript coercion semantics.
- Retain plugin-specific evaluation and receiver-classification policies.

Each pull request updates `architecture.md` where ownership changes and the
rule utility reference with entry points, evaluation scope, and limitations.
