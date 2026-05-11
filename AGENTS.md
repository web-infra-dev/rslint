# Repository Guidelines

This document summarizes how to work on rslint effectively and consistently.

## Project Structure & Module Organization

- `architecture.md`: Current high-level architecture, major runtime flows, and subsystem relationships.
- `cmd/rslint/`: CLI entry (default), IPC API (`--api`), LSP (`--lsp`).
- `internal/config/`: Config types/loader, rule registry and registration.
- `internal/linter/`: Linter engine, traversal, and fix application.
- `internal/rule/`: Rule framework, diagnostics, disable manager, listeners.
- `internal/plugins/typescript/`: `@typescript-eslint` rules under `rules/<rule>/`.
- `internal/plugins/import/`: `eslint-plugin-import` registration.
- `internal/utils/`: JSONC, overlay VFS, TS program creation, helpers.
- `internal/lsp/`: Language Server integration. Also see `website/` and `packages/` for UI/tooling.

## Build, Test, and Development Commands

- Setup submodule: `git submodule update --init --recursive`
- Install Deps: `pnpm install`
- Build JS/TS: `pnpm build`
- Run Go tests: `pnpm run test:go`
- Run JS tests: `pnpm run test`
- Run Check Spell: `pnpm run check-spell`
- Lint Go: `pnpm run lint:go`
- Lint JS: `pnpm run lint`
- Format JS/TS/MD: `pnpm run format`
- CLI: invoke the Node wrapper, not the Go binary directly — the Go
  binary requires an IPC handshake from a Node parent and will time
  out otherwise. Use `pnpm --filter @rslint/core dev -- --help` or run
  `node packages/rslint/bin/rslint.js --help` after `pnpm build`.
  - Examples: append `--config rslint.config.ts`, `--fix`, `--format default|jsonline|github`, `--quiet`, `--max-warnings 0`.
- LSP: started by editors (e.g. VS Code extension) via
  `node packages/vscode-extension/dist/main.js` after `pnpm build`;
  the extension spawns the Go binary with `--lsp` for you.
- IPC API mode: `go run ./cmd/rslint --api` is fine — `--api` and
  `--lsp` are the only flags the Go binary self-services without the
  Node wrapper. The default (no-flag) path always expects an IPC
  parent and is not a useful target for direct `go run` invocations.

## Coding Style & Naming Conventions

- Go uses gofmt/goimports; keep functions focused and small.
- TS/JS/MD/CSS use Prettier via `pnpm run format`.
- Rules: `internal/plugins/typescript/rules/<rule>/`; tests: `<rule>_test.go`.
- Prefer table-driven tests and existing helpers in `internal/utils`.

## Testing Guidelines

- Co-locate Go tests with implementation; name files `*_test.go` and functions `TestXxx`.
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
- Use `--format github` in CI to emit GitHub workflow annotations.

## Website UI Guidelines (shadcn/ui)

- Prefer shadcn/ui components from `@components/ui/*` (e.g., `button`, `toggle-group`, `alert`, `card`, `table`) over custom elements.
- Minimize custom CSS. Use component variants, utility classes, and existing styles instead of adding new selectors when possible.
- Icons: use `lucide-react` for consistent iconography (e.g., import `{ Share2Icon, CheckIcon } from 'lucide-react'`).
- Keep layout simple: compose shadcn primitives and flex utilities for alignment instead of bespoke CSS blocks.
- Only add custom CSS for domain‑specific visuals that primitives can’t express (e.g., AST tree expanders), and keep it scoped.
