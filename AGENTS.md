 # AGENTS.md — Agent Instructions for rslint

 These instructions apply to the entire repository unless a more deeply nested AGENTS.md overrides them.

 ## Scope and Priorities

 - Do the minimum necessary, high-quality change to satisfy the task.
 - Preserve existing public/CLI behavior unless explicitly asked to change it.
 - Do not introduce unrelated changes or refactors.
 - Match existing code style and structure; prefer local patterns and utilities.
 - Avoid adding new external dependencies without explicit request.

 ## How to Run

 - CLI: `go run ./cmd/rslint [--config rslint.jsonc] [--fix] [--format default|jsonline|github] [--quiet] [--max-warnings N]`
 - LSP: `go run ./cmd/rslint --lsp`
 - IPC API: `go run ./cmd/rslint --api` (length‑prefixed JSON over stdio)
 - Go tests: `pnpm run test:go`
 - JS build: `pnpm build`
 - JS tests: `pnpm run test`
 - Lint Go: `pnpm run lint:go`
 - Format JS/TS/MD/CSS: `pnpm run format`

 ## Repository Layout (essentials)

 - `cmd/rslint/` — CLI, IPC (`--api`), and LSP (`--lsp`) entrypoints
 - `internal/config/` — Config types/loader, rule registry, registration
 - `internal/linter/` — Linter core and fix application
 - `internal/rule/` — Rule framework, diagnostics, disable manager, listener kinds
 - `internal/plugins/typescript/` — `@typescript-eslint` rules (`rules/` subfolders)
 - `internal/plugins/import/` — `eslint-plugin-import` registration and recommended set
 - `internal/utils/` — JSONC, overlay VFS, TS program creation, helpers
 - `internal/lsp/` — Language Server integration

 See `architecture.md` for a broader overview.

 ## Configuration

 - Config files: `rslint.json` or `rslint.jsonc` (JSONC supported)
 - Rules accept ESLint-compatible formats: string level, `{ level, options }` object, or `[level, options?]` array
 - Plugins: `"@typescript-eslint"`, `"eslint-plugin-import"`
 - Ignores: glob patterns matched with doublestar against normalized paths

 ## Implementing/Modifying Rules (Go)

 - Place TypeScript plugin rules under `internal/plugins/typescript/rules/<rule>/`.
 - Implement a `rule.Rule` with `Run(ctx, options) RuleListeners`.
 - Listener kinds from `internal/rule/rule.go`:
   - On-enter: `ast.KindX`
   - On-exit: `rule.ListenerOnExit(ast.KindX)`
   - Allow‑pattern aware: `rule.ListenerOnAllowPattern/ListenerOnNotAllowPattern`
 - Use `RuleContext` (SourceFile, Program, TypeChecker, DisableManager) and report via provided helpers.
 - Register new rules in `internal/config/config.go` within `RegisterAllRules()`.

 ## Diagnostics and Fixes

 - Use `RuleMessage{ Id, Description }`.
 - Provide fixes as `RuleFix{ Text, Range }` and/or suggestions (`RuleSuggestion`).
 - The CLI `--fix` applies non‑overlapping edits with `internal/linter/ApplyRuleFixes`.

 ## Concurrency & Traversal

 - The linter walks each file once and dispatches to all registered listeners.
 - Allow‑pattern contexts are propagated for assignment LHS, array/object patterns, spreads, and property values.
 - CLI uses a work group to parallelize across programs/files; `--singleThreaded` disables parallelism.

 ## Testing

 - Prefer small, table‑driven tests colocated with rules: `<rule>_test.go`.
 - Keep tests focused on changed behavior; do not add unrelated tests.
 - Run `pnpm run test:go` locally before completing.

 ## Do / Don’t

 Do:
 - Use existing helpers in `internal/utils` and typescript‑go shims.
 - Keep changes narrow and consistent with surrounding code.
 - Update `architecture.md` and this file only if the behavior or guidance changes.

 Don’t:
 - Rename or move files/packages unless requested.
 - Introduce new dependencies or large refactors without approval.
 - Change CLI flags or output formats unless explicitly asked.

 ---
 
 For more context, see `architecture.md` and `AGENTS.md` at the repo root.

