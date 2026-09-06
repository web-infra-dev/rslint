# Rule Porting Quick Reference

Use this file to look up commands and locations. Read the relevant [porting contract](./PORT_RULE.md) for the behavior they verify. Branch selection, test layout and verification scope follow [AGENTS.md](../../../../AGENTS.md); formatting selections and the commit gate follow [CONTRIBUTING.md — Verify a change](../../../../CONTRIBUTING.md#verify-a-change).

## Commands

Run from the repository root unless the command selects a workspace with `--dir`. Replace placeholders with the affected paths, identified from the diff and callers. Reuse passing results while their relevant inputs remain unchanged.

| When needed                                                           | Command                                                                      |
| --------------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| Core rule implementation or tests changed                             | `go test ./internal/rules/<rule_name>`                                       |
| Plugin rule implementation or tests changed                           | `go test ./internal/plugins/<go-plugin>/rules/<rule_name>`                   |
| Shared behavior changed                                               | `go test <changed-package-dir> <affected-consumer-package-dirs>`             |
| Focus a reproduction before package verification                      | `go test <rule-package-dir> -run '<test-or-subtest-pattern>'`                |
| Catalog registration changed                                          | `go test ./internal/rules -run '^TestAllContainsEveryGoRuleExactlyOnce$'`    |
| Lint affected Go packages                                             | `golangci-lint run --new-from-merge-base=<base-ref> <affected-package-dirs>` |
| Fix formatting in changed Go files                                    | `gofmt -w <changed-go-files>`                                                |
| Fix formatting in supported, non-ignored changed JS/TS/Markdown files | `pnpm exec rs fmt <changed-js-ts-md-files>`                                  |

Keep the catalog test's `-run` selection for registration-only changes: unfiltered `go test ./internal/rules` also runs all-rule compiler compatibility and heritage suites. Shared code changes require the affected consumers, not an automatic whole-plugin run. The Go lint base is the task's selected base ref, normally `origin/main`.

Formatting commands still obey repository exclusions, including explicitly passed rule Markdown paths. Skip empty or ignored-only selections. Required delivery and commit checks are defined in the linked repository instructions; this command table is not a checklist to run in full.

## Build and JS tests

Prepare only missing or stale artifacts needed by the selected test. The build commands produce different outputs:

| Artifact and invalidation condition                                                                                         | Command                                                                |
| --------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------- |
| Native binary: relevant Go code changed, or the binary is missing/stale                                                     | `pnpm --filter @rslint/core build:bin`                                 |
| Schema dump: catalog/schema changed, or the dump is missing/stale                                                           | `go run ./tools/dump_rule_schemas > packages/rslint/rule-schemas.json` |
| Core JS and public option types: core JS changed, required `dist` output is missing/stale, or the schema dump was refreshed | `pnpm --filter @rslint/core build:js`                                  |
| TypeScript rule tester: the selected test imports `@typescript-eslint/rule-tester` and its `dist` is missing/stale          | `pnpm --filter @typescript-eslint/rule-tester build`                   |

`build:bin` builds the Go binary only. Refresh the schema dump **before** core `build:js` after catalog/schema changes: the JS build reads that dump to generate public option types and does not create it automatically. Native Go-rule IPC tests do not by themselves require a Rust parser rebuild.

Verify the exact registered JS file after its required artifacts are ready:

```bash
CI=true pnpm --dir packages/rslint-test-tools exec rs test run tests/<suite>/rules/<rule-name>.test.ts
```

Use `-u` only when adding or intentionally updating snapshots in a wrapper that actually uses snapshots:

```bash
pnpm --dir packages/rslint-test-tools exec rs test run tests/<suite>/rules/<rule-name>.test.ts -u
```

Review generated snapshots against upstream expectations, then verify with the CI command. Local Rstest runs may add missing snapshots without `-u`; CI verification must fail on missing snapshots. Wrapper capabilities differ: inspect the selected wrapper's assertions and skip handling rather than assuming all plugin suites use snapshots. See [Coverage and assertions](./PORT_RULE.md#coverage-and-assertions).

## Workspace names and directories

Use package names with `--filter` and directories with `--dir`; they are not interchangeable.

| Package name                     | Source directory              | Relevant output or configuration                                  |
| -------------------------------- | ----------------------------- | ----------------------------------------------------------------- |
| `@rslint/core`                   | `packages/rslint/`            | `dist/`, `rule-schemas.json`, `rstack.config.ts`                  |
| `@rslint/test-tools`             | `packages/rslint-test-tools/` | `rstack.config.mts` selects integration files                     |
| `@typescript-eslint/rule-tester` | `packages/rule-tester/`       | `src/index.ts` implements the wrapper; imports resolve to `dist/` |

## Rule files and registration

Go rule directories and filenames use snake_case (`<rule_name>`); exported rule variables use PascalCase with a `Rule` suffix. Rule keys and JS test filenames use kebab-case (`<rule-name>`); preserve upstream message IDs.

| Item                                           | Location                                                                                         |
| ---------------------------------------------- | ------------------------------------------------------------------------------------------------ |
| Core rule source, documentation and Go tests   | `internal/rules/<rule_name>/`                                                                    |
| Plugin rule source, documentation and Go tests | `internal/plugins/<go-plugin>/rules/<rule_name>/`                                                |
| Rule documentation                             | `<rule_name>.md` beside the Go implementation                                                    |
| Options schema, when options exist             | `<rule_name>.schema.json` beside the Go implementation                                           |
| Core catalog entry                             | Import and entry in `internal/rules/all.go` → `coreRules()`                                      |
| Plugin catalog entry                           | Import and entry in `internal/plugins/<go-plugin>/all.go` → `GetAllRules()`                      |
| JS rule mirror                                 | `packages/rslint-test-tools/tests/<suite>/rules/<rule-name>.test.ts`                             |
| JS test registration                           | `packages/rslint-test-tools/rstack.config.mts` → `include`                                       |
| Suite configuration and local wrapper          | `packages/rslint-test-tools/tests/<suite>/rslint.config.mjs` and `rule-tester.ts`, where present |

Resolve `<go-plugin>` and `<suite>` from the actual family; directory names do not always match public prefixes:

| Family            | Go plugin directory | JS suite directory          | Catalog key                      |
| ----------------- | ------------------- | --------------------------- | -------------------------------- |
| ESLint core       | Core path above     | `eslint`                    | `<rule-name>`                    |
| typescript-eslint | `typescript`        | `typescript-eslint`         | `@typescript-eslint/<rule-name>` |
| import            | `import`            | `eslint-plugin-import`      | `import/<rule-name>`             |
| jest              | `jest`              | `eslint-plugin-jest`        | `jest/<rule-name>`               |
| jsx-a11y          | `jsx_a11y`          | `eslint-plugin-jsx-a11y`    | `jsx-a11y/<rule-name>`           |
| promise           | `promise`           | `eslint-plugin-promise`     | `promise/<rule-name>`            |
| react             | `react`             | `eslint-plugin-react`       | `react/<rule-name>`              |
| react-hooks       | `react_hooks`       | `eslint-plugin-react-hooks` | `react-hooks/<rule-name>`        |
| rstest            | `rstest`            | `rstest`                    | `rstest/<rule-name>`             |
| unicorn           | `unicorn`           | `eslint-plugin-unicorn`     | `unicorn/<rule-name>`            |

The catalog key is `rule.Name`. `rule.CreateRule` automatically prefixes `@typescript-eslint/`, so use it only for that family. Core rules use a bare `rule.Rule` name; other plugins put their public prefix directly in `Name`. Resolve deprecated/extended rule origin using [Upstream contract](./PORT_RULE.md#upstream-contract), rather than registering both core and TypeScript keys.

`rules.All()` combines explicit sources in `internal/rules/all.go`; a new plugin directory is not discovered automatically. Its first native rule also requires adding the plugin's `GetAllRules()` to that aggregation and checking the plugin-enablement boundaries in `architecture.md`. New rules do not belong in `internal/config`. See [Integration and documentation](./PORT_RULE.md#integration-and-documentation).

## API and contract lookup

Use the entry for the upstream operation, then locate the named declaration in its owning file with `rg -n`. Read that declaration and relevant callers when its contract is uncertain. These are lookup locations, not a prerequisite reading list.

| Question                                                                        | Source or reference                                                                                                                       |
| ------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| Rule declaration, listeners and prefix factory                                  | `internal/rule/rule.go`; [Listener Types](./AST_PATTERNS.md#listener-types)                                                               |
| Options array, compilation and defaults                                         | `internal/rule/schema.go`; [Options and schema](./PORT_RULE.md#options-and-schema)                                                        |
| Go case fields and diagnostic/edit assertions                                   | `internal/rule_tester/rule_tester.go`; [Coverage and assertions](./PORT_RULE.md#coverage-and-assertions)                                  |
| Member/call expressions, private/computed keys, JSX/heritage, parentheses/JSDoc | [Member and Call Expressions](./AST_PATTERNS.md#member-and-call-expressions); `internal/utils/ast_helpers.go` and `internal/utils/jsx.go` |
| Literal values and raw text                                                     | [Literal Kinds](./AST_PATTERNS.md#literal-kinds), [Node Text and Positions](./AST_PATTERNS.md#node-text-and-positions)                    |
| Diagnostics, ranges and deferred edits                                          | `internal/rule/context.go`; [Reporting Functions](./AST_PATTERNS.md#reporting-functions)                                                  |
| References, globals and source/module services                                  | `internal/rule/ref_store.go`, `internal/rule/globals.go`; [Framework boundaries](./PORT_RULE.md#framework-boundaries)                     |
| Scope-sensitive upstream behavior                                               | `internal/utils/scope/`; [ESLint Scope Model](./UTILS_REFERENCE.md#internalutilsscope---eslint-scope-model)                               |
| TypeChecker availability and access                                             | [Using TypeChecker](./AST_PATTERNS.md#using-typechecker)                                                                                  |
| JS strings, numbers, Unicode, regexps and globs                                 | [JavaScript Semantics](./UTILS_REFERENCE.md#javascript-semantics-ecmascript-minimatch3-isglob)                                            |
| Equivalent comparison inputs and output limitations                             | [Differential validation](./PORT_RULE.md#differential-validation)                                                                         |
| Completion scope or a failing integration check                                 | [Delivery and troubleshooting](./PORT_RULE.md#delivery-and-troubleshooting)                                                               |
