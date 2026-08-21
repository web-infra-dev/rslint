# Repository Guidelines

This document summarizes how to work on rslint effectively and consistently.

## Project Structure & Module Organization

- `architecture.md`: Current high-level architecture, major runtime flows, and subsystem relationships.
- `cmd/rslint/`: CLI entry (default), IPC API (`--api`), LSP (`--lsp`).
- `internal/config/`: Config types/loader, rule registry and registration.
- `internal/program/`: Unified source Program, module resolution/graph, and generation-scoped derived caches.
- `internal/linter/`: Linter engine, traversal, and fix application.
- `internal/rule/`: Rule descriptors/environment, context, diagnostics, disable manager, listeners.
- `internal/plugins/typescript/`: `@typescript-eslint` rules under `rules/<rule>/`.
- `internal/plugins/import/`: `eslint-plugin-import` registration.
- `internal/testutil/`: Cross-package test infrastructure, including safe txtar fixture materialization.
- `internal/utils/`: JSONC, overlay VFS, compiler construction, AST/type helpers.
- `internal/utils/ecmascript/`: JavaScript's own semantics (trim, blank, upper/lower case, number-to-string, the general categories) plus `ecmascript/regexp` for a JavaScript RegExp and its `/i` comparison. `internal/utils/minimatch3/` and `internal/utils/isglob/` port the glob packages ESLint plugins depend on.
- `internal/lsp/`: Language Server integration. Also see `website/` and `packages/` for UI/tooling.

## Build, Test, and Development Commands

- Setup submodule: `git submodule update --init --depth 1`
- Install Deps: `pnpm install`
- Build JS/TS: `pnpm build`
- Run Go tests: `pnpm run test:go`
- Run JS tests: `pnpm run test`
- Run Check Spell: `pnpm run check-spell`
- Lint Go: `pnpm run lint:go`
- Lint JS: `pnpm run lint`
- Format JS/TS/MD: `pnpm run format`
- CLI: `go run ./cmd/rslint --help`
  - Examples: `go run ./cmd/rslint --config rslint.jsonc`, `--fix`, `--format default|jsonline|github|gitlab`, `--quiet`, `--max-warnings 0`
- LSP: `go run ./cmd/rslint --lsp` | IPC API: `go run ./cmd/rslint --api`

## Coding Style & Naming Conventions

- Go uses gofmt/goimports; keep functions focused and small.
- TS/JS/MD/CSS use Prettier via `pnpm run format`.
- Rules: `internal/plugins/typescript/rules/<rule>/`; tests: `<rule>_test.go`.
- Prefer table-driven tests. Keep package-specific helpers beside their tests; put reusable test infrastructure in `internal/testutil`, not production utility packages.
- A value that came from JavaScript — a string to trim or case, a character to classify, a number to print, a regexp or glob out of a rule option — is read through `internal/utils/ecmascript`, `ecmascript/regexp`, `minimatch3`, or `isglob`, never through `strings.TrimSpace`, `strings.ToLower`, the `unicode` package, the stdlib `regexp`, or `doublestar`. An identifier question goes to tsgo's `scanner`, so a rule and the parser never disagree. `depguard` and `forbidigo` enforce this under `internal/rules/**` and `internal/plugins/**`.
- The stdlib `regexp` is for a pattern written in this repository that RE2 and JavaScript read the same way and that no user input reaches. A pattern out of a rule option, a config file or the source under lint takes `esregexp`, however plain it looks.
- **Only minimatch 3 and is-glob are ported.** A rule that needs another glob package — minimatch 10 included — is reported to the user, naming the package and version, rather than being pointed at `minimatch3` or `doublestar` or given a fresh port.

## Testing Guidelines

- Co-locate Go tests with implementation; name files `*_test.go` and functions `TestXxx`.
- Keep small inputs inline. Put multi-file textual filesystem fixtures under the nearest package's `testdata/` directory, and group related layouts in `.txtar` when that makes the scenario easier to review.
- Use `.txtar` only for portable regular text files. Construct symlinks, permissions, concurrency, and other OS behavior directly in Go tests.
- Fixture helpers must reject missing or empty selections instead of allowing a test to pass without exercising a case.
- Keep tests minimal and behavior-focused; avoid unrelated scenarios.
- Run `pnpm run test:go` (Go) and `pnpm run test` (JS) before submitting.

## Commit & Pull Request Guidelines

- Use Conventional Commits: `feat:`, `fix:`, `chore:`, `docs:`, `ci:`, etc.
- PRs should be small, with clear description, repro steps, and linked issues.
- Include examples (commands or code) and update docs when behavior changes.
- Preserve existing CLI behavior unless a change is explicitly requested.

## Architecture & Configuration Tips

- Read `architecture.md` before making broad changes that touch module boundaries, entrypoints, or cross-package flows.
- If a change affects the high-level architecture, runtime data flow, or major integration paths, update `architecture.md` in the same change.
- rslint loads `rslint.json`/`rslint.jsonc`; rules accept ESLint-style levels/options.
- The linter walks each file once and dispatches to registered listeners; `--singleThreaded` disables parallelism.
- Use `--format github` in CI to emit GitHub workflow annotations, or `--format gitlab` to emit a Code Quality report (`codequality` artifact) for GitLab CI merge requests.

## Website UI Guidelines (shadcn/ui)

- Prefer shadcn/ui components from `@components/ui/*` (e.g., `button`, `toggle-group`, `alert`, `card`, `table`) over custom elements.
- Minimize custom CSS. Use component variants, utility classes, and existing styles instead of adding new selectors when possible.
- Icons: use `lucide-react` for consistent iconography (e.g., import `{ Share2Icon, CheckIcon } from 'lucide-react'`).
- Keep layout simple: compose shadcn primitives and flex utilities for alignment instead of bespoke CSS blocks.
- Only add custom CSS for domain‑specific visuals that primitives can’t express (e.g., AST tree expanders), and keep it scoped.
