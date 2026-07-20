# Architecture Overview

Rslint is a high-performance JavaScript and TypeScript linter, designed as a drop-in replacement for ESLint and TypeScript-ESLint. It leverages [typescript-go](https://github.com/microsoft/typescript-go) to achieve 20-40x speedup over traditional ESLint setups through native parsing, direct TypeScript AST usage, and parallel processing.

## Table of Contents

- [1. Goals & Non-Goals](#1-goals--non-goals)
- [2. High-Level System Diagram](#2-high-level-system-diagram)
- [3. Directory / Crate Structure](#3-directory--crate-structure)
- [4. Parsing Pipeline](#4-parsing-pipeline)
- [5. Abstract Syntax Tree (AST)](#5-abstract-syntax-tree-ast)
- [6. Lint Rule Framework](#6-lint-rule-framework)
- [7. Diagnostics & Autofixes](#7-diagnostics--autofixes)
- [8. Configuration & Directives](#8-configuration--directives)
- [9. CLI Flow](#9-cli-flow)
- [10. Performance & Memory Considerations](#10-performance--memory-considerations)
- [11. Extensibility & Future Directions](#11-extensibility--future-directions)
- [12. Testing Strategy](#12-testing-strategy)
- [13. Adding a New Rule (Checklist)](#13-adding-a-new-rule-checklist)
- [14. Dependency Layering & Boundaries](#14-dependency-layering--boundaries)
- [15. Data Flow (Textual Diagram)](#15-data-flow-textual-diagram)
- [16. Glossary](#16-glossary)
- [17. TODO / Open Questions](#17-todo--open-questions)

## 1. Goals & Non-Goals

### Goals

- **Lightning Fast Performance**: 20-40x faster than ESLint through Go implementation and typescript-go integration
- **ESLint Compatibility**: Best effort compatibility with ESLint and TypeScript-ESLint configurations and rules
- **TypeScript First**: Uses TypeScript Compiler semantics as single source of truth for 100% consistency
- **Project-Level Analysis**: Cross-module analysis by default for powerful semantic linting
- **Monorepo Ready**: First-class support for large-scale monorepos with TypeScript project references
- **Batteries Included**: Ships with all existing TypeScript-ESLint rules and widely-used ESLint rules

### Non-Goals

- **Complete Third-Party Plugin Compatibility**: The Node worker supports third-party ESLint plugins on a best-effort API surface, not every parser, processor, or ESLint runtime API
- **Runtime Performance Optimization**: Optimized for build-time linting, not runtime performance
- **Custom Parser Support**: Standardized on TypeScript parser through typescript-go

## 2. High-Level System Diagram

```
┌────────────────────────────────────────────────────────────────────────────────┐
│                              RSLINT SYSTEM                                     │
├────────────────────────────────────────────────────────────────────────────────┤
│                                                                                │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────┐     │
│  │ CLI       │  │ IPC API   │  │ LSP Server│  │ Website   │  │ tsgo Tool │     │
│  │ cmd/rslint│  │ cmd/rslint│  │ cmd/rslint│  │ Playground│  │ cmd/tsgo  │     │
│  │           │  │ --api     │  │ --lsp     │  │           │  │           │     │
│  └────┬──────┘  └────┬──────┘  └────┬──────┘  └────┬──────┘  └────┬──────┘     │
│       │              │              │              │              │            │
├───────┴──────────────┴──────────────┴──────────────┴──────────────┴────────────┤
│                                     │                                          │
│  ┌──────────────────────────────────▼───────────────────────────────────────┐  │
│  │                              GO BACKEND                                  │  │
│  │                                                                          │  │
│  │  ┌──────────────────────────────────────┐  ┌──────────────────────────┐  │  │
│  │  │ LINT CORE                            │  │ ADAPTERS / AUXILIARY     │  │  │
│  │  │                                      │  │                          │  │  │
│  │  │  internal/config                     │  │  internal/api            │  │  │
│  │  │  cmd/rslint/programs.go              │  │  internal/lsp            │  │  │
│  │  │  internal/linter                     │  │  internal/inspector      │  │  │
│  │  │  internal/rule                       │  │                          │  │  │
│  │  │  internal/utils                      │  │                          │  │  │
│  │  └──────────────────────────────────────┘  └──────────────────────────┘  │  │
│  └─────────────────────────┬────────────────────────────────────────────────┘  │
│                            │                                                   │
│  ┌─────────────────────────▼────────────────────────────────────────────────┐  │
│  │                    TYPESCRIPT-GO FOUNDATION / BRIDGE                     │  │
│  │                                                                          │  │
│  │  ┌──────────────────┐  ┌──────────────────┐  ┌────────────────────────┐  │  │
│  │  │ typescript-go    │  │ shim/            │  │ tools/                 │  │  │
│  │  │ parser / AST     │  │ generated bridge │  │ shim generator         │  │  │
│  │  │ checker / Program│  │ import surface   │  │ ts-go updater          │  │  │
│  │  │ Session / VFS    │  │                  │  │                        │  │  │
│  │  └──────────────────┘  └──────────────────┘  └────────────────────────┘  │  │
│  └─────────────────────────┬────────────────────────────────────────────────┘  │
│                            │                                                   │
│  ┌─────────────────────────▼────────────────────────────────────────────────┐  │
│  │                         PACKAGES / CLIENTS                               │  │
│  │                                                                          │  │
│  │  ┌────────────────────────────────┐  ┌─────────────────────────────────┐ │  │
│  │  │ WEBSITE / PLAYGROUND           │  │ OTHER PACKAGES / CLIENTS        │ │  │
│  │  │ website Playground             │  │ packages/rslint                 │ │  │
│  │  │ packages/rslint-wasm           │  │ packages/vscode-extension       │ │  │
│  │  │ packages/rslint-api            │  │ packages/rslint-test-tools      │ │  │
│  │  │ browser worker / wasm runtime  │  │ packages/rule-tester            │ │  │
│  │  │ lint path    -> internal/linter│  │ crates/tsgo-client              │ │  │
│  │  │ inspect path -> internal/      │  │                                 │ │  │
│  │  │                 inspector      │  │                                 │ │  │
│  │  └────────────────────────────────┘  └─────────────────────────────────┘ │  │
│  └──────────────────────────────────────────────────────────────────────────┘  │
│                                                                                │
└────────────────────────────────────────────────────────────────────────────────┘
```

## 3. Directory / Crate Structure

The directory map below folds the high-level module relationships into the package list, so each row shows both role and main dependencies.

| Path                           | Purpose                                                                                                                                     | Key Relationships                                                                                                                                                                                                                                                                      |
| ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `website/`                     | Documentation site and Playground UI                                                                                                        | Uses `packages/rslint-wasm` to run browser linting and `packages/rslint-api` to decode encoded source files; Playground lint requests ultimately reach `internal/linter`, and inspect requests reach `internal/inspector` through `internal/api`                                       |
| `cmd/rslint/`                  | Main Go binary entry point with CLI, API, and LSP modes                                                                                     | JS/TS config CLI path is `internal/config/discovery -> internal/config -> cmd/rslint/programs.go -> internal/linter`; JSON config starts at `internal/config`. `--api` is consumed by `packages/rslint` and `packages/rslint-wasm`; `--lsp` is consumed by `packages/vscode-extension` |
| `internal/output/`             | Report model, summary, colors, and stdout formatters                                                                                        | Consumes final sorted `internal/rule` diagnostics and renders `default`, `jsonline`, `github`, or `gitlab`; the CLI is its current consumer, while the package remains available to other repository integrations that need the same output behavior                                   |
| `cmd/tsgo/`                    | ts-go semantic inspection/export tool                                                                                                       | Talks directly to `typescript-go` and bypasses the lint framework; consumed by `packages/tsgo` and `crates/tsgo-client`                                                                                                                                                                |
| `internal/api/`                | stdio IPC protocol and service types for JS/WASM integration                                                                                | Shared protocol layer for `cmd/rslint --api`; used by `packages/rslint`, `packages/rslint-wasm`, `internal/linter`, and `internal/inspector`                                                                                                                                           |
| `internal/config/`             | Configuration models, JSON loading, matching/merging, runtime ownership resolution, lint-target planning, and centralized rule registration | Owns the shared authored-global-ignore matcher consumed by both discovery phases. `RegisterAllRules()` orchestrates rule registration; `rule_registry.go` implements registry/query logic used by `cmd/rslint/programs.go` and `internal/linter`                                       |
| `internal/config/discovery/`   | Go-owned JS/TS config candidate discovery and immutable catalog construction                                                                | Imports the parent config model/matching policy, batches exact candidates to a host-supplied Node loader, and returns configs/scopes/failures/effective IDs. CLI, API, and LSP call `DiscoverAutomatic` or `LoadExplicitConfig`; the parent package never imports this child package   |
| `internal/config/gitignore/`   | Config-scoped `.gitignore` parsing, directory reachability, and pattern projection                                                          | Automatic catalog discovery carries a filesystem-independent cursor through its existing walk, pruning Git-inaccessible config subtrees and freezing observed patterns for lint-target admission; explicit/JSON fallback paths reuse the standalone collector                          |
| `internal/inspector/`          | AST/type/symbol/signature/flow inspection for Playground                                                                                    | Auxiliary backend used mainly by website Playground inspect panels; builds rich semantic data from `typescript-go` programs                                                                                                                                                            |
| `internal/linter/`             | Core lint engine, traversal, and fix application                                                                                            | Consumes rules from `internal/rule`, file config from `internal/config`, and `Program` / `TypeChecker` data from `typescript-go`; also serves `internal/api` and `internal/lsp`                                                                                                        |
| `internal/lsp/`                | Language Server Protocol implementation                                                                                                     | Wraps `typescript-go project.Session`, owns transactional config discovery and last-good commit state with `packages/vscode-extension` as the module/plugin host, and invokes `internal/linter` on session-backed programs                                                             |
| `internal/rule/`               | Rule framework, context, diagnostics, fixes, and disable manager                                                                            | Shared foundation for core rules and plugin rules; called by `internal/linter` through listeners and reporting APIs                                                                                                                                                                    |
| `internal/rule_tester/`        | Go-side rule testing helpers                                                                                                                | Supports rule development and complements JS-side testers in `packages/rule-tester` and `packages/rslint-test-tools`                                                                                                                                                                   |
| `internal/rules/`              | Core lint rule implementations without plugin namespace; `all.go` aggregates them into the `GetAllRules()` slice                            | `internal/config/config.go`'s `RegisterAllRules()` consumes the slice and registers each rule; then executed by `internal/linter` like plugin rules                                                                                                                                    |
| `internal/plugins/typescript/` | `@typescript-eslint`-style rules                                                                                                            | Registered into the shared rule registry by `RegisterAllRules()` and often rely on `TypeChecker` from `typescript-go`                                                                                                                                                                  |
| `internal/plugins/react/`      | React rule implementations                                                                                                                  | Registered into the shared rule registry by `RegisterAllRules()` and executed through the same listener pipeline in `internal/linter`                                                                                                                                                  |
| `internal/plugins/jest/`       | Jest rule implementations                                                                                                                   | Registered into the shared rule registry by `RegisterAllRules()` and executed through the same listener pipeline in `internal/linter`                                                                                                                                                  |
| `internal/plugins/import/`     | Import plugin registration and rules                                                                                                        | Contributes plugin rules through `RegisterAllRules()` and participates in normal config-driven linting                                                                                                                                                                                 |
| `internal/utils/`              | Shared utilities for JSONC, compiler hosts, overlay VFS, and helpers                                                                        | Supports `cmd/rslint/programs.go`, config loading, and various linter entry points                                                                                                                                                                                                     |
| `packages/rslint/`             | Main npm package with JavaScript API and CLI wrapper                                                                                        | Spawns `cmd/rslint --api` in JavaScript runtime environments and uses `internal/api` message shapes                                                                                                                                                                                    |
| `packages/rslint-api/`         | Frontend-facing encoded source file / AST decoding helpers                                                                                  | Used mainly by website Playground to decode AST/source data returned from the Go API                                                                                                                                                                                                   |
| `packages/rslint-test-tools/`  | Testing utilities and cross-ecosystem rule tests                                                                                            | Supports package-side and integration-style tests around the linter and rule ecosystem                                                                                                                                                                                                 |
| `packages/rslint-wasm/`        | Browser/WASM package for running `rslint --api` in a worker                                                                                 | Starts the browser worker, hosts the wasm runtime, and bridges website Playground requests to `internal/api`, `internal/linter`, and `internal/inspector`                                                                                                                              |
| `packages/rule-tester/`        | Forked `@typescript-eslint/rule-tester` package used in tests                                                                               | JS-side rule testing support that complements Go-side helpers                                                                                                                                                                                                                          |
| `packages/utils/`              | Shared JavaScript utilities                                                                                                                 | Shared support package for the JS/website tooling layer                                                                                                                                                                                                                                |
| `packages/vscode-extension/`   | VS Code extension for IDE integration                                                                                                       | Launches `cmd/rslint --lsp`, starts `rslint/configRefresh`, serves reverse load/activate/commit/abort and plugin-lint requests, and consumes diagnostics/code actions from `internal/lsp`                                                                                              |
| `packages/tsgo/`               | JS wrapper package for the `tsgo` tool                                                                                                      | JavaScript-facing wrapper around `cmd/tsgo` output                                                                                                                                                                                                                                     |
| `typescript-go/`               | Git submodule containing TypeScript compiler Go port                                                                                        | Provides parser, AST, checker, `Program`, `project.Session`, diagnostics, scanner, and VFS primitives used throughout the backend                                                                                                                                                      |
| `shim/`                        | Generated bridge packages exposing ts-go internals                                                                                          | Bridge layer between repository Go code and `typescript-go` internals; generated and updated by `tools/`                                                                                                                                                                               |
| `tools/`                       | Shim generator and ts-go update scripts                                                                                                     | Generates `shim/` code and maintains the pinned `typescript-go` integration                                                                                                                                                                                                            |
| `crates/tsgo-client/`          | Rust client for communicating with `cmd/tsgo`                                                                                               | Spawns `cmd/tsgo` and consumes its semantic/project output from Rust                                                                                                                                                                                                                   |

## 4. Parsing Pipeline

The current parsing and linting pipeline is built on top of ts-go `Program` and `project.Session` rather than a standalone parser wrapper written by rslint itself.

### Pipeline Overview

```
┌───────────────────────┐
│ Source Text           │
│ - disk files          │
│ - overlay VFS         │
│ - LSP document state  │
└───────────┬───────────┘
            │
            ▼
┌───────────────────────┐
│ ts-go Program /       │
│ project.Session       │
└───────────┬───────────┘
            │
            ▼
┌───────────────────────┐
│ ts-go Parser / AST    │
│ + optional Checker    │
└───────────┬───────────┘
            │
            ▼
┌───────────────────────┐
│ Rule Initialization   │
│ -> listener registry  │
└───────────┬───────────┘
            │
            ▼
┌───────────────────────┐
│ Single AST Traversal  │
│ + listener dispatch   │
└───────────┬───────────┘
            │
            ▼
┌───────────────────────┐
│ Diagnostics / Fixes / │
│ Suggestions / Output  │
└───────────────────────┘
```

### Detailed Pipeline Steps

1. **Source Text Loading**: Files come from the real filesystem, an overlay VFS, or LSP document overlays. CLI runs and individual API requests keep a compiler-host source snapshot keyed by the exact normalized source path. The snapshot stores the first successful text read and its xxh3 hash for the current generation, allowing later Programs to skip both work items. CLI fix writes replace the entire source generation before rebuilding Programs; API snapshots end with the request. LSP does not use this one-shot layer because its `project.Session` already follows document versions and `didChange` updates.
2. **Program Construction**: Plain CLI/API lint resolves its effective targets first and builds `Program` objects only for governing configs that own a selected target. Program-wide type-check modes build every project declared by the effective loaded config catalog. LSP reuses matching projects from ts-go `project.Session`; a declared custom tsconfig that project service has not loaded is built on demand against an isolated editor overlay.
3. **Lexical + Syntax Parsing**: ts-go tokenizes and parses source files into TypeScript-native AST nodes.
4. **Semantic Analysis**: When needed, the linter acquires a `TypeChecker` from the `Program`.
5. **Rule Registration**: Enabled rules register listeners keyed by AST kind.
6. **AST Traversal**: The linter traverses each file once using a DFS walk.
7. **Rule Execution**: Listener callbacks inspect nodes and may use syntax only or syntax plus type information.
8. **Diagnostic Collection**: Findings are reported as `RuleDiagnostic` values, with optional fixes or suggestions.
9. **Output Generation**: CLI builds one report from the final post-fix diagnostics and passes it to `internal/output`; API returns structured data, and LSP converts diagnostics to LSP diagnostics/code actions.

### Error Recovery Strategy

The parser and program builder are tolerant enough to support editor and fallback scenarios:

- ts-go can continue producing ASTs after syntax errors
- LSP delays lint on rapid edits to avoid repeated work on broken intermediate text
- lint rules and third-party plugin dispatch are suppressed for malformed lint targets; syntax diagnostics remain authoritative
- CLI/API create tsconfig-backed Programs leniently so plain lint can decide
  syntax diagnostics from the final lint target set instead of failing during
  broad tsconfig construction
- `--type-check` and `--type-check-only` still surface TypeScript syntactic and
  semantic diagnostics from the tsconfig-backed Program boundary
- fallback Programs for selected files outside tsconfig coverage are not
  project-backed, do not enable type-aware rules, and are intentionally skipped
  by the type-check phase

## 5. Abstract Syntax Tree (AST)

### AST Representation

The AST comes directly from ts-go. Rslint does not build a second custom AST layer for linting.

Important characteristics:

- **Node Types**: ts-go `ast.Kind` values
- **Node Objects**: `*ast.Node` and `*ast.SourceFile`
- **Traversal Style**: `ForEachChild(...)` with depth-first recursion
- **Source Locations**: node ranges and source-file-aware line/column conversion via scanner helpers
- **Comments**: exposed through one lazy per-file store for directives and comment-based rules

### Key AST Properties

In practice, rules usually interact with:

- `node.Kind`
- `node.Pos()` / `node.End()`
- `node.Loc`
- `file.Node`
- `node.ForEachChild(...)`
- ts-go helper predicates and casts, for example assignment-expression checks

Rslint also trims leading trivia when reporting node-based diagnostics so that disable comments do not shift reported positions upward.

### Span and Location Handling

- **Positions**: ts-go source positions are stored as offsets and later converted to editor-friendly line/column values
- **Ranges**: lint diagnostics use `core.TextRange`
- **Line/Column**: computed through `scanner.GetECMALineAndUTF16CharacterOfPosition(...)`
- **Editor Encoding**: LSP diagnostics and edits are emitted using LSP position encoding rules

## 6. Lint Rule Framework

### Rule Interface

Rules are defined in `internal/rule/rule.go`:

```go
type Rule struct {
    Name             string
    RequiresTypeInfo bool
    Run              func(ctx RuleContext, options []any) RuleListeners
}

type RuleListeners map[ast.Kind]func(node *ast.Node)
```

`RequiresTypeInfo` is important because gap-file fallback Programs intentionally do not run type-aware rules.

### Rule Context

`RuleContext` is the runtime environment passed to each rule. It includes:

```go
type RuleContext struct {
    SourceFile                 *ast.SourceFile
    Settings                   map[string]interface{}
    ConfigGlobals              map[string]bool
    InlineGlobals              []InlineGlobal
    Globals                    map[string]bool
    Comments                   *CommentStore
    Program                    *compiler.Program
    TypeChecker                *checker.Checker
    DisableManager             *DisableManager
    ReportRange                func(textRange core.TextRange, msg RuleMessage)
    ReportRangeWithFixes       func(textRange core.TextRange, msg RuleMessage, fixes ...RuleFix)
    ReportRangeWithSuggestions func(textRange core.TextRange, msg RuleMessage, suggestions ...RuleSuggestion)
    ReportNode                 func(node *ast.Node, msg RuleMessage)
    ReportNodeWithFixes        func(node *ast.Node, msg RuleMessage, fixes ...RuleFix)
    ReportNodeWithSuggestions  func(node *ast.Node, msg RuleMessage, suggestions ...RuleSuggestion)
}
```

The linter creates one short-lived `CommentStore` per file. `Comments.All()`
materializes the scanner-backed, source-ordered, deduplicated comment list only
for the first consumer; later consumers share that list. A source without `//`
or `/*` takes a cheap byte-scan fast path. Inline-global parsing first checks
for an exact raw-text directive candidate, so ordinary files do not force
comment collection.
`ConfigGlobals` preserves the effective `languageOptions.globals` source,
`InlineGlobals` preserves ordered comment name ranges, and `Globals` is the
resolved map after inline settings override configuration. Rules consume this
context data instead of scanning comments independently. Configured access
aliases are normalized consistently: writable and read-only aliases declare a
name, `null` is read-only, and only `"off"` disables it. The Node plugin scope separately
preserves writable versus read-only access for ESLint-compatible scope APIs.

### Listener Registration

Rules do not walk the AST themselves. Instead:

1. `RegisterAllRules()` in `internal/config/config.go` populates the global registry once
2. config merge selects enabled rules for a file
3. each enabled rule runs `Run(ctx)`
4. `Run(ctx)` returns listeners keyed by `ast.Kind`
5. the linter aggregates them into a per-file dispatch table

This allows one AST traversal to serve many rules.

### Listener Types

- **OnEnter**: the default listener keyed by a real `ast.Kind`
- **OnExit**: synthetic listener kind created by `ListenerOnExit(kind)`
- **OnAllowPattern**: synthetic listener kind used for pattern/destructuring contexts
- **OnNotAllowPattern**: synthetic listener kind used for non-pattern contexts of the same AST shape

Those synthetic kinds are defined by offsetting real `ast.Kind` values. They are a rule-framework dispatch mechanism, not native ts-go node kinds.

## 7. Diagnostics & Autofixes

### Diagnostic Structure

The actual diagnostic model is text-range based and fix-aware:

```go
type RuleDiagnostic struct {
    Range       core.TextRange
    RuleName    string
    Message     RuleMessage
    FixesPtr    *[]RuleFix
    Suggestions *[]RuleSuggestion
    SourceFile  *ast.SourceFile
    Severity    DiagnosticSeverity
}

type RuleFix struct {
    Text  string
    Range core.TextRange
}
```

### Severity Levels

- `SeverityError`: lint error
- `SeverityWarning`: lint warning
- `SeverityOff`: rule disabled

### Autofix System

Autofix is implemented as text edits:

- insert = replace an empty range with text
- replace = replace a non-empty range with text
- remove = replace a range with the empty string

Rules attach fixes through `ReportRangeWithFixes(...)` or `ReportNodeWithFixes(...)`.

Fix application happens in `internal/linter/source_code_fixer.go`:

1. sort fixes within each diagnostic
2. sort fixable diagnostics by position
3. skip overlapping or conflicting adjacent edits
4. rebuild the source text

Important behavior differences by integration:

- **CLI**: can rerun lint and fix for multiple passes
- **LSP quick fix**: returns direct text edits for one diagnostic
- **LSP fix-all**: runs repeated lint-fix cycles, then returns one whole-document replacement edit
- **API**: `lint({ fix: true })` applies fixes in a single pass and returns the fixed source per file in `output` (the JS side persists it via `Rslint.outputFixes`). There is no separate `applyFixes`, and—unlike the CLI—it does not re-lint across passes.

## 8. Configuration & Directives

### Configuration Formats

Rslint supports two configuration formats following ESLint flat config semantics (array of config entries):

#### JS/TS Configuration (Recommended)

Rslint automatically discovers `rslint.config.js`, `rslint.config.mjs`, `rslint.config.ts`, and `rslint.config.mts`. Explicit configuration paths also support `.cjs` and `.cts` files through CLI `--config` and API `overrideConfigFile`. JS/TS config files support preset composition via `defineConfig()`:

```typescript
import { defineConfig, ts } from '@rslint/core';

export default defineConfig([
  {
    ignores: ['**/dist/**', '**/fixtures/**'],
  },
  ts.configs.recommended,
  {
    rules: {
      '@typescript-eslint/no-unused-vars': 'error',
      '@typescript-eslint/array-type': ['warn', { default: 'array' }],
    },
  },
]);
```

Available presets currently include:

- `ts.configs.recommended`
- `js.configs.recommended`
- `reactPlugin.configs.recommended`
- `importPlugin.configs.recommended`

#### JSON Configuration (Deprecated)

JSON config files (`rslint.json`, `rslint.jsonc`) are deprecated and will be removed in a future version. A deprecation warning is printed to stderr when used. Run `rslint --init` to generate a recommended JS/TS config.

```json
[
  {
    "ignores": ["./files-not-want-lint.ts", "./tests/**/fixtures/**.ts"],
    "languageOptions": {
      "parserOptions": {
        "project": ["./tsconfig.json", "packages/app1/tsconfig.json"]
      }
    },
    "plugins": ["@typescript-eslint"],
    "rules": {
      "@typescript-eslint/no-unused-vars": "error",
      "@typescript-eslint/array-type": ["warn", { "default": "array" }]
    }
  }
]
```

**Key difference**: JSON configs are normalized by `normalizeJSONConfig()`, which auto-enables core rules and rules from declared plugins unless explicitly overridden. JS/TS configs only enable what the normalized config entries specify, usually via presets.

### Config Entry Structure

Each entry in the config array supports:

| Field             | Type                                       | Description                                                                          |
| ----------------- | ------------------------------------------ | ------------------------------------------------------------------------------------ |
| `files`           | `(string \| string[])[]`                   | Non-empty selector list; top-level selectors are ORed and nested selectors are ANDed |
| `ignores`         | `string[]`                                 | Glob patterns excluded by this entry                                                 |
| `languageOptions` | `object`                                   | Parser options such as `project` and `projectService`                                |
| `rules`           | `Record<string, …>`                        | Rule level or `[level, options]`                                                     |
| `plugins`         | `string[] \| Record<string, ESLintPlugin>` | Native plugin declarations or third-party plugin instances                           |
| `settings`        | `Record<string, …>`                        | Shared settings available in `RuleContext`                                           |

### Configuration Loading

The loading flow differs by config type:

**JS/TS staged catalog discovery**:

CLI, the native JavaScript API path, and transactional LSP refreshes reuse
the one-shot `internal/config/discovery.DiscoverAutomatic` operation (or
`LoadExplicitConfig` for an exact path). Automatic discovery builds an immutable
config/ownership catalog and observes `.gitignore` sources during the same
directory walk; it does not collect lint targets. Go owns candidate discovery,
default exclusions, config hierarchy, authored and Git directory reachability,
the frozen Git projection for each owner, and final effective IDs. Node only
executes the exact JS/TS modules requested by Go, normalizes their entries,
retains live third-party plugin objects, and returns serializable entries. Source
fingerprints stay in the Node transaction session; after Go selects the final
effective IDs, activation revalidates those fingerprints and returns only the
effective plugin metadata.

The package boundary is deliberate: `internal/config/discovery` imports the
parent `internal/config` model and its narrow authored-global-ignore matcher;
the parent never imports discovery. Runtime file routing stays in the parent
`ConfigOwnerResolver`, while CLI/API/LSP adapters own transport, commit/abort,
and last-good lifecycle. Discovery has no cross-transaction session,
synchronization, or generation state because every production request is one
transaction; request-local coordination only freezes concurrent observations.
A process-random nonce plus atomic sequence allocates IDs that cannot collide
with a stale host session after a native-process restart. The returned catalog
publishes final configs, scopes, failures, effective IDs, plugin metadata, and
whether the invocation used an explicit config. Candidate fingerprints and
plugin-aggregation scratch remain private to the Node transaction session;
source-selection scratch remains private to the Go builder.

Automatic discovery uses these rules:

1. `.git` and `node_modules` are default discovery boundaries. Within one
   directory, automatic filename priority is `.js` → `.mjs` → `.ts` → `.mts`.
   After the first successful JS/TS owner, Git-inaccessible directory nodes are
   config-discovery boundaries. Git never filters a candidate filename, so a
   local or ancestor pattern naming `rslint.config.js` does not change priority
   or fall through to `.mjs`. A directly supplied directory/static-glob root
   reopens that root's inherited Git gate, but not hidden intermediate configs;
   overlapping supplied roots retain independent reachability. When a requested
   root is default-excluded, Go skips downward traversal but still resolves
   reachable ancestors outside the boundary.
2. A directory target's config ancestry is evaluated outer-to-inner. The
   current successful ancestor's standalone global ignores can therefore stop
   a nested config before it executes. An absolute directory cover such as
   `dir/**` prunes the frontier. A file cover such as `dir/**/*` keeps the
   directory traversable for later authored negations, but an automatic config
   candidate that still matches that authored cover is not executed; filename
   priority falls through to the first non-ignored candidate. Ordered authored
   patterns may reopen a Git-blocked node when they match that exact directory:
   `!dir`, `!dir/`, and `!dir/**` reopen `dir`, while `!dir/**/*` and
   `!dir/file.ts` do not. Descendant patterns such as `!dir/*` can reopen a
   child node they directly match, and a later matching positive pattern closes
   the node again.
   CLI/LSP directory roots (including mixed
   CLI file-and-directory input) recursively scan reachable sibling frontiers
   with a bounded worker pool. When the native API supplies an already-expanded
   exact file set plus static glob roots, Go builds a lexical target-ancestor
   trie and visits only branches that can govern those files. Candidate nodes in
   either frontier are suspended, their paths are sent in one stable batch, and
   the next frontier begins only after that result is merged in lexical order.
3. A literal file is the sole authored-global-ignore ownership exception: it searches
   nearest-first and falls back to a loadable ancestor when a nearer candidate
   fails. Lexical ancestry is authoritative; canonical ancestry is consulted
   only when the complete lexical ancestry has no candidate. A default-excluded
   literal file is not lintable and cannot escape through canonical fallback.
   Git is not used to choose that literal's owner; `.gitignore` is applied after
   ownership is known. A config reached only
   through the literal exception is marked explicit-only: it owns its
   discovery-scoped literal files and uses its directory as the `.gitignore`
   source boundary for that scope, but is excluded from automatic lint-target
   ownership and handoff. Ancestor-owned automatic siblings therefore continue
   reading nested `.gitignore` sources through that directory. Files produced by a
   glob/directory expansion do not independently reopen authored-global-ignore
   discovery boundaries. If literal and automatic routes select different
   candidate filenames in the same directory, the automatic candidate defines
   that directory's single config boundary and the literal file remains scoped
   to it.
4. Each `loadConfigs` batch carries a protocol version, transaction ID, load
   mode, the `--singleThreaded` hint, and opaque candidate IDs. Go validates the
   matching ordered results. `configDirectory` is an opaque Go-owned routing
   identity and must round-trip byte-for-byte; Node may native-normalize only
   `configPath`, which it uses for local file I/O and module import.
   Before sending a batch, Go coalesces verified native case aliases to one
   stable candidate ID and representative directory across all frontiers. The
   check requires both lexical case equivalence and matching resolved file and
   directory paths; arbitrary symlink owners are not deduplicated by physical
   target.
   After ownership is resolved, `activateConfigs` names only effective IDs;
   Node rechecks fingerprints, prepares plugin state only for that set, then
   rechecks the same effective sources before publishing the activation. This
   prevents a worker re-import during preparation from observing different
   config bytes than the normalized entries Go is about to commit.
5. A successful child config resets Git scope at its own directory: parent Git
   rules stop at that ownership handoff, then the child's local `.gitignore` is
   read. A failed child does not form a boundary and inherits the parent cursor.
   Config evaluation always precedes reading that directory's local Git source,
   and a Git-inaccessible branch's nested `.gitignore` is never read.
6. A failed candidate is recorded and discovery continues with the last
   reachable successful owner. If candidates existed but none loaded,
   discovery returns `ErrAllConfigsFailed` and does not activate Node state.
   Partial failures remain in `catalog.Failures`: CLI and native API emit
   warnings to stderr, LSP logs them after a successful commit, and all three
   continue with the effective fallback catalog. The LSP first-start recovery described below handles
   `ErrAllConfigsFailed` outside this shared coordinator.

The transport and target phase differ by surface:

- CLI sends `loadConfigs` / `activateConfigs` as reverse framed-IPC requests
  during initialization. The resulting catalog and the later Go lint-target
  walker are separate traversals, but config discovery already freezes the Git
  sources observed on its reachable frontier. There is no second per-owner
  directory sweep. Literal-only and mixed literal targets contribute only their
  exact owner-to-target source chains to the same source-keyed projection.
- `Rslint.lintFiles()` still expands target globs with `tinyglobby`, preserves
  literal provenance and canonical identities, and sends the resulting files
  plus their static config-scan roots in one API lint request. Go bounds catalog
  discovery to the supplied files' ancestor trie, owns file-to-config
  assignment, then applies selectors and ignores once to that exact target set;
  configs in unrelated descendants are not evaluated. The bidirectional API
  advertises `reverseConfigLoadV1`; low-level pre-resolved `config` requests and
  WASM do not use staged module discovery. If the resulting catalog contains
  object-form community plugins, Go additionally requires the peer's
  `reversePluginLint` capability at the request boundary; plugin-free catalogs
  do not require that handler. Every long-lived API call uses a
  fresh entry-module load so rewritten config bytes cannot be paired with stale
  normalized exports or a newer plugin-worker topology. API `overrideConfig`
  entries are structurally validated before that load and attached at the
  loader boundary as the final suffix of every successful config. Their global
  ignores and negations therefore participate in staged reachability and are
  published exactly once; an empty catalog uses the same override directly.
- The extension owns shared UI, commands, and output channels once. Its
  `WorkspaceRslintCoordinator` keys desired and active roots by workspace-folder
  URI (not display name), subscribes to folder changes before awaiting any root,
  and independently starts or closes one `Rslint` runtime per URI generation.
  One slow or failed root therefore cannot serialize healthy roots; terminal
  shutdown still attempts every per-root close before releasing shared
  resources. If an old generation does not confirm that it closed, its URI slot
  remains quarantined: no replacement starts, and the retained error is
  included in terminal shutdown. Each runtime owns its native server children,
  config watcher, transaction adapter, plugin pool, request handlers, and
  workspace logger. A process owner covers automatic LanguageClient restarts,
  terminates any still-live prior child before a restart spawn, blocks new
  spawns once closing begins, and awaits stdio close after bounded forced
  termination of any child that survives protocol shutdown. A closing-aware
  client error handler forbids restart, and per-root close waits for the pending
  initialize/state tail before extension-wide channels are released.
- Every runtime keeps a workspace-relative document selector, while
  `WorkspaceDocumentRouter` is the single authority for overlapping selectors.
  Among ready roots it assigns an open supported document to the longest
  matching URI root. A root activation or removal performs an ordered
  `didClose`/diagnostic-clear/`didOpen` handoff using the document's current
  in-memory text, without requiring the editor to close. Middleware admits
  changes, saves, diagnostics, and code actions only for the exact active
  runtime that currently owns the server-open document. Exact runtime identity
  rejects diagnostics from a replaced same-URI client, while a document epoch
  rejects code actions that finish after an ownership change. When the
  LanguageClient automatically restarts a native server, the router invalidates
  every recorded server-open session for that runtime—including documents that
  closed during the feature-listener gap—before LanguageClient replays
  `didOpen`, so the replacement process receives every still-open document
  exactly once and a later same-URI reopen cannot inherit stale ownership. The
  reset is queued as soon as a running transport exits and repeated at the next
  `Running` transition; root removal also drains any exact-runtime session that
  disappeared from VS Code's open-document list during that gap.
- Each root starts `rslint/configRefresh`, and Go scans that process cwd with a
  transaction-scoped cached VFS. Go sends
  `rslint/loadConfigs` and
  `rslint/activateConfigs`, then commits or aborts the matching plugin-host state
  through `rslint/commitConfigs` / `rslint/abortConfigs`. `fresh` loads cache-bust the config entry
  module; static transitive imports retain Node's normal module cache. If the
  first plugin preparation detects a source change between its two fingerprint
  checks, the extension keeps the language client alive and retries one
  serialized refresh from the new bytes.
- If `vscode-languageclient` automatically restarts the native process, its
  later `Running` transition first aborts any extension-side orphaned
  transaction, then requests a new initial catalog through the same serialized
  refresh chain. The previously committed plugin host remains available until
  the replacement Go process commits its own catalog.
- Only a fully committed Go/Node snapshot replaces a usable last-good
  snapshot. All-candidate failure, or a partial failure at an existing
  committed boundary, aborts and preserves that snapshot; a newly broken child
  can still use the core parent fallback. On first startup with every JS config
  broken, Go instead commits empty Node plugin-host state plus unavailable ownership
  boundaries, keeping the LSP alive without allowing JSON fallback through the
  broken subtrees. A Node commit retains one rollback predecessor: if the commit
  response is lost, Go's abort restores it; the next successful commit confirms
  the prior host state and begins normal grace retirement. Open documents remain
  separate per-file targets resolved against the committed catalog.

The LSP config wire exposes one identity, `transactionId`. The extension reuses
that value internally as the `PluginLintPool` host generation so an in-flight
plugin request is routed to the exact worker state paired with Go's committed
catalog. This is a concurrency/lifecycle identity, not a second config-discovery
model. The independent numeric document generation only rejects stale async
diagnostics after edits, closes, or config commits.

An explicit JS/TS `--config` or API `overrideConfigFile` bypasses automatic
candidate selection and loads the exact module. The invocation cwd remains its
matching directory. The exact config path is never gated by `.gitignore`;
lint targets are filtered afterward with the existing invocation-scoped
collector. Automatic candidates instead use the Git directory reachability
rules above.

No-candidate behavior is surface-specific. CLI performs no Node activation and
continues through its normal JSON fallback. Native API discovery performs no
reverse config call and uses `overrideConfig`, or an empty syntax-only config;
it never searches disk for JSON fallback. LSP explicitly stages and commits
an empty plugin-host state (an empty load batch followed by zero-ID activation),
while loading any JSON fallback in Go as part of the new snapshot. That empty
catalog is not a usable JavaScript last-good boundary: if a newly created JS
config is broken, LSP commits an unavailable boundary for it rather than
silently retaining JSON fallback below it.

**JSON config**:

1. Go searches for `rslint.json` / `rslint.jsonc`
2. JSONC parsing is applied
3. `normalizeJSONConfig()` injects core and plugin rules as defaults

JSON remains on the existing Go `ConfigLoader` path, not the JS staged module
coordinator. CLI loads it directly (including explicit non-JS `--config`), and
LSP keeps it as the Go-loaded fallback for files with no JS owner. The native
API discovery path has no disk JSON fallback; low-level API callers may instead
send an already-resolved serialized `config`.

### Configuration Merging

Config merging follows flat-config-style semantics in `GetConfigForFile()`:

1. entries containing only `ignores` and an optional `name` form the global-ignore set
2. the implicit default extension baseline plus effective explicit `files` selectors defines the config selector union; an entry's `ignores` prevents its selector from extending that union, top-level selectors are ORed, nested patterns are ANDed, and an explicit `files: []` is invalid
3. entries without `files` cascade across that selector union, while entry-level `ignores` prevent only that entry from contributing configuration to an otherwise selected file
4. later rule values override earlier values; a severity-only override retains earlier rule options
5. settings and language options recursively merge ordinary objects, while later arrays and scalar values replace earlier values

The staged coordinator builds the effective catalog used by
`ConfigOwnerResolver.Resolve()` before `GetConfigForFile()` merges the selected
entries. Staged CLI, native API, and transactional LSP paths therefore reuse the
same Go ownership rules instead of independently reconstructing hierarchy on
the Node side.

Ownership lookup never compares depth across lexical and physical path spaces:
the nearest exact lexical config wins, a native case alias is accepted only
after filesystem identity verification, and realpath ancestry is consulted only
when no lexical owner exists. Directory-walk handoff boundaries are likewise
built only from the lexical config hierarchy. Canonical paths remain file and
Program identities, not a second config inheritance tree. Before activation,
two distinct lexical config directories resolving to one physical directory are
rejected; verified alternate native casing of the same directory is the only
allowed alias.

Additional current behaviors:

- `.gitignore` is injected as a global-ignore entry through the shared
  `ConfigWithCollectedGitignore`/`ConfigWithGitignore` policy. The governing
  config directory is a hard upper boundary: its own and nested `.gitignore`
  files apply, while parent `.gitignore` files do not. In automatic catalogs,
  the staged walk records sources by owner and case-aware source identity,
  orders them parent-before-child, and materializes the synthetic Git entry
  before publishing. Direct automatically reachable child config directories
  are downward ownership handoff boundaries.
  Configs loaded only for explicit targets bound only their literal target
  chains, so adding a literal cannot truncate an ancestor-owned automatic
  target's `.gitignore` sources. This preserves ESLint v10's per-target global
  ignore behavior: adding another literal target cannot change whether an
  existing target is ignored. File-only CLI/API requests read only target
  directory chains within each governing config. The synthetic Git entry is
  ordered before authored entries, so a later config `!` may re-include a target
- when the client supports dynamic file-watch registration, Go watches
  workspace-descendant `.gitignore` files plus exact `.gitignore` paths in
  strict workspace ancestors that may contain an automatically selected config.
  Extension watchers are the sole refresh owner for
  workspace/descendant JS configs, JSON fallback, and dependency lockfiles;
  Go additionally watches only strict-ancestor JS configs and `.gitignore`.
  ts-go project watchers may still forward the same workspace events into the
  session, but those forwarded JS/JSON events do not start a second fresh config
  transaction. Create/change/delete events rebuild the frozen config/ignore snapshot and
  refresh open-document diagnostics
- the VS Code extension preserves last-good JS configs during reloads; a newly
  unavailable config with no usable JS ancestor contributes an empty boundary,
  preventing JSON fallback only in that authored config subtree. A normal
  transactional refresh receives successful entries with their Git projection
  already frozen, adds unavailable boundaries, then freezes and commits the Go
  catalog and Node plugin host under one transaction ID. Failures preserve
  a usable last-good catalog and ignore view together; the first-start all-broken
  recovery instead commits unavailable boundaries.
  If the first valid catalog cannot initialize its optional community-plugin
  worker, LSP commits the ordinary native config with an empty no-host plugin
  state and retries on later refreshes; once a usable snapshot exists, the same
  worker failure aborts and preserves that last-good snapshot. A successful
  no-candidate transaction removes the previous JS catalog and
  exposes the Go-loaded JSON fallback
- native and third-party plugin rules are gated by their normalized prefixes for JS/TS configs; third-party rules use process-wide Go registry placeholders, but LSP additionally filters those placeholders through the exact rule-name set committed for the current Node generation so metadata retained from an older generation cannot be dispatched to a newer worker
- CLI/API lint target selection is independent from TypeScript `Program` membership and considers only rslint-supported script extensions. The `.js`, `.mjs`, `.cjs`, `.jsx`, `.ts`, `.tsx`, `.mts`, and `.cts` default baseline is always selected; explicit config `files` contributes candidates only within the supported set. Global ignores and `.gitignore` remove targets, while an entry-level ignore prevents only its own selector/config contribution
- selected CLI/API targets can still appear as 0-rule lint results when no config entry contributes rules; this applies to default-baseline directory discovery and explicit supported files, and syntax diagnostics remain available in that state
- under automatic discovery, each selected file is governed by its nearest loadable config; an explicitly selected config is used directly. In either case, a target can bind only to a tsconfig declared by its governing config, and the first declared project containing the file wins
- `files`/`ignores` matching uses the stable target path in the governing config's path space; a Program source alias is used only to locate the AST and type information, so moving a target into or out of a tsconfig cannot change its rule configuration
- within each Program-registry build, normalized declared tsconfig paths are deduplicated across config associations; CLI fix passes create a new registry build. File-symlink declarations remain distinct because TypeScript resolves relative paths from the declared location. Selected files outside the governing config's Programs receive a non-project-backed fallback Program, and targets whose names collide under a case-insensitive ts-go path key are partitioned across fallback Programs so distinct physical files remain distinct
- `--type-check` and `--type-check-only` build every real tsconfig declared by the effective loaded config catalog. Git reachability may change which automatic configs enter that catalog; once it is established, program-wide checking is not filtered by lint targets, config `files`/`ignores`, `.gitignore`, or CLI file/directory arguments. Synthetic fallback Programs never participate, and `--type-check-only` skips the separate lint-target walk.
- for LSP, an open supported script is a per-file target independent of Program membership. Global config ignores, `.gitignore`, default-excluded paths, and unavailable config boundaries suppress native rules, plugin rules, and fixes; an available zero-rule config still parses the target and can report syntax diagnostics

### Inline Directives

Rslint supports inline directives with both `rslint-` and `eslint-` prefixes:

- `// rslint-disable-next-line @typescript-eslint/no-unused-vars`
- `/* rslint-disable @typescript-eslint/no-unsafe-assignment */`
- `// eslint-disable-next-line`

The `DisableManager` in `internal/rule/disable_manager.go` applies these
directives before diagnostics are emitted. It defers parsing until the first
disable check and uses an exact directive-text candidate check, so files without a
supported directive retain an empty manager without scanning comments.

## 9. CLI Flow

### Command Line Interface

```bash
rslint [options] [files...]
```

### CLI Processing Flow

The CLI has a two-layer architecture: a Node.js wrapper (`packages/rslint/src/cli/cli.ts`) and the Go binary (`cmd/rslint/`).

1. **Node.js Wrapper**: parses args, starts the Go engine, and hosts JS/TS module evaluation plus live third-party plugin objects
2. **Config Catalog**: for automatic JS/TS discovery or an explicitly selected JS/TS config, Go builds the staged catalog and batches exact module-evaluation requests to Node. If automatic discovery finds no candidate, or a non-JS config was explicitly selected, the existing Go JSON loader path remains in control
3. **Mode Selection**:
   - `--lsp`: starts the LSP server
   - `--api`: starts the IPC API server
   - default: runs direct CLI linting
4. **Lint Target Plan**: Go resolves a stable target set from CLI/API scope, the implicit default baseline, explicit config `files`, global ignores, and `.gitignore`
5. **Program Registry**: plain lint builds each normalized tsconfig path declared by an active governing config once; `--type-check` and `--type-check-only` instead retain every project declared by the effective loaded config catalog. Shared declared paths preserve each active config association and declaration order.
6. **Program Binding**: each target is bound by exact lexical or canonical filesystem identity to the first containing Program declared by its governing config; unbound targets, including projects with no tsconfig, are parsed through a non-project-backed fallback Program
7. **Rule Resolution**: `getRulesForFile` resolves enabled rules from the stable lint-target path, never the Program source alias, and filters type-aware rules off no-type-info gap files
8. **Rule Execution**: `RunLinter()` schedules per-Program work over the exact target plan; the unexported `runLintRulesInProgram()` does the actual per-file traversal. When `--type-check` is enabled, a separate program-wide pass over real tsconfig Programs aggregates `tsc --noEmit`-aligned diagnostics through `collectNoEmitDiagnostics()`
9. **Result Aggregation**: diagnostics are sent through one run-scoped diagnostics channel and collected at the CLI layer
10. **Fix Passes**: CLI multi-pass `--fix` applies fixes, rebuilds real Programs, and rebinds the unchanged target plan after each pass; a file may move between a real Program and fallback as its import graph changes
11. **Report Assembly**: the CLI builds one output report from the final post-fix diagnostics plus run metadata. Diagnostics carry an explicit lint or TypeScript origin, and the report computes error/warning/type-error counts once so the summary and exit policy use the same values; `--quiet` filters rendering only.
12. **Output Formatting**: the CLI-private output subsystem renders `default`, JSON line, GitHub workflow command, or GitLab Code Quality formats. Only `default` emits a summary; machine-readable formats never emit ANSI styling or a summary.
13. **Exit Code**: depends on the report counts, `--max-warnings`, and fix outcomes

### Concurrency Model

The main Go workload work groups and pools below honor `--singleThreaded`.
The flag serializes these workload stages, but IPC transport, diagnostic
collection, and plugin dispatch may still use infrastructure goroutines.

1. **Linter work group** (`RunLinter()` via `core.NewWorkGroup`)
   - Schedules per-Program lint work; runs rules in parallel within a Program.
   - `--singleThreaded` collapses the work group to serial execution.

2. **Type-check work group** (`runTypeCheckAcrossPrograms`)
   - Schedules diagnostics for real tsconfig Programs and merges results in
     stable Program order.
   - `--singleThreaded` computes Program diagnostics serially.

3. **Staged catalog discovery and Program creation**
   - A bounded Go worker pool scans one reachable sibling frontier at a time.
     Config-boundary nodes are suspended, batched after that frontier is
     processed, and resumed only after the Node result is merged.
   - Native API roots carry an exact target-ancestor trie from the tinyglobby
     result, so only sibling branches leading to supplied files enter those
     frontiers. CLI/LSP directory roots, including CLI mixed file+directory
     invocations, remain recursively unbounded within ignore boundaries.
   - Directory-root ancestry is loaded outer-to-inner before the root frontier.
     Each successful owner's authored global ignores and Git cursor control
     continuation below that boundary; local Git sources are observed only after
     the directory's config decision. Literal files use the separate
     nearest-first ownership exception described above.
   - Discovery commits only the config catalog in stable order. Plain lint then
     runs the separate lint-target walker and constructs Programs only for
     represented configs. `--type-check` and `--type-check-only` construct
     every Program in the effective reachable catalog; the latter skips the
     lint-target walk.
   - Configured Programs remain serial in stable config/project order because
     typescript-go's API is invoked one Program at a time.
   - `--singleThreaded` executes the same state machine with one Go discovery
     worker and serializes module evaluation within each Node frontier batch.
     Coordinator batches and results remain ordered in either mode.

4. **Lint-target directory walker** (`internal/config/file_discovery.go`)
   - `DiscoverLintTargets` uses a fixed-size worker pool (`walkPool`) that
     walks the directory tree. `DiscoverLintFiles` is the path-only
     compatibility wrapper. Live goroutine count is capped at `workers`, not
     the number of directories.
   - Default `workers = max(2, GOMAXPROCS)`; `--singleThreaded` forces
     `workers = 1`, which degenerates into a fully serial DFS-style traversal.
   - The walker is built on a `vfsAdapter` with `followSymlinks = false`:
     nested symlinked subdirectories are skipped. An explicitly requested
     directory alias is resolved once and used as a bounded scan root. This
     matches ESLint v10's
     flat-config file walker, which uses `@humanfs/node` and recurses only
     when `Dirent.isDirectory()` is true. Node's
     `readdir({ withFileTypes: true })` reports the dirent type without
     following symlinks, so
     `Dirent.isDirectory()` is false for symlinks. The skip also eliminates
     scheduling-dependent non-determinism that a parallel walker would
     otherwise introduce.

5. **Program source identity index** (`bindLintTargetPlan`)
   - Direct lexical Program lookups remain synchronous. If one misses, the
     binder resolves each unique Program source path once and builds a
     binding-pass canonical identity index. CLI fix passes rebuild it when they
     rebind the target plan.
   - Independent realpath lookups use `core.NewWorkGroup`; `--singleThreaded`
     runs the same work serially.

Other invariants:

- Target discovery returns both the caller-visible lexical path and a canonical
  identity hint. A regular directory walk derives the canonical path from the
  canonical config root without a per-file realpath call. Explicit directory
  aliases are resolved once and their descendants inherit the corresponding
  physical path; explicit files and file symlinks are resolved individually.
  Canonical identities use exact comparison, and aliases governed by different
  configs are rejected instead of choosing an owner by scan order.
- Multi-config target discovery processes config roots in stable order. It uses
  catalog-provided scopes for explicit-file ownership and invokes the bounded
  lint-target walker for each config. Its automatic ownership index omits
  explicit-only configs, while their scoped literal files are still processed
  by the corresponding config.
- Explicit targets stay in the caller's lexical path space for config
  ownership. Go consults physical ancestry only when lexical discovery finds no
  candidate. Literal files try nearer candidates before ancestors; directory
  roots evaluate their complete candidate ancestry outer-to-inner so an
  ancestor global ignore can prevent a nested module from executing.
- Go applies the same strict lexical-first ordering to the already-loaded config
  catalog. A physically deeper config loaded for another target cannot replace
  an existing lexical owner; physical ancestry is only a fallback for paths
  with no lexical config.
- `Rslint.lintFiles()` applies the same rule before creating a third-party
  plugin host or sending API requests. Its realpath memo is bounded to one API
  call and is not a persistent cache. Canonical target paths resolved during
  that plan are sent with the request so Go does not repeat the same realpath
  work.
- LSP uses a different orchestration model and keeps session access on its main
  dispatch loop. Its Program-source lookup follows the same exact lexical and
  canonical filesystem identity rules as CLI/API binding, including
  file-symlink aliases and rejection of case-folded nonidentical paths.

#### `--singleThreaded` semantics

`--singleThreaded` is honored in every parallelism point above:

| Point                         | Effect when set                                                                                               |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------- |
| Linter work group             | Collapsed to serial via `core.NewWorkGroup(true)`.                                                            |
| Type-check work group         | Program diagnostics are computed serially via `core.NewWorkGroup(true)`.                                      |
| Staged catalog discovery      | Sibling-frontier workers and Node module evaluation are serialized; batches and catalog merge remain ordered. |
| Lint-target walker workers    | Forced to 1 (single goroutine, no concurrency).                                                               |
| Program source identity index | Canonical source paths are resolved serially through `core.NewWorkGroup(true)`.                               |

These workload stages run serially with `--singleThreaded`; infrastructure
goroutines remain outside that guarantee.

## 10. Performance & Memory Considerations

### Design Principles

- **Direct ts-go Data Model**: rslint operates on ts-go `Program`, AST, and `TypeChecker` objects directly instead of converting through a second AST representation
- **Program-Level Parallelism**: `RunLinter` queues work per `Program` through `core.NewWorkGroup`; `--singleThreaded` forces the same flow to run serially
- **Single-Walk Rule Dispatch**: each file is traversed once, with rules registering listeners up front and sharing the same AST walk
- **Early Filtering**: exact lint target plans, skip paths, global-ignore filters, and gap-file type filtering reduce work before listeners run

### Performance Optimizations

- **Native Go Implementation**: Eliminates JavaScript runtime overhead
- **Direct TypeScript AST**: No AST conversion between parsers
- **Shared Type Checker**: `runLintRulesInProgram` acquires one checker for the lint phase and reuses it across files and rules in the same `Program`. The subsequent type-check phase (when `--type-check` is enabled) lets `GetSemanticDiagnostics` reacquire its own checker so the lint-phase checker can be released first
- **Checker Phase Separation**: the checker is released before TypeScript semantic diagnostics run, so `GetSemanticDiagnostics` can reacquire its own checker cleanly
- **File Filtering**: Skip node_modules and bundled files automatically
- **Gap-File Degradation**: fallback gap-file Programs skip type-aware rules and semantic diagnostics instead of paying unreliable semantic costs
- **Buffered Diagnostic Collection**: CLI mode funnels diagnostics through a buffered channel before formatting, which reduces contention between lint workers and output handling
- **On-Demand AST Encoding**: API/WASM responses only include encoded source files when `IncludeEncodedSourceFiles` is requested
- **Lazy Shared Comments**: each file owns one `CommentStore`; directive consumers and comment-aware rules materialize its canonical comment list only when needed and reuse the result. Rule-specific text checks avoid scanner work when their comment syntax cannot occur

### Caching Strategy

- **LSP Session Reuse**: `internal/lsp` builds a shared ts-go `project.Session`, so configured projects, inferred projects, and overlay document state are reused across requests
- **Parse Cache in LSP**: the LSP server passes a shared `project.ParseCache` into the session to avoid re-parsing from scratch on every change
- **Debounced Re-linting**: `refreshCh` and `debounceCh` collapse bursts of file changes and session refreshes onto the main dispatch loop
- **CLI/API Are Mostly Fresh Runs**: CLI and one-shot API requests generally rebuild `Program` state per run; there is no repository-local rule-result cache or persistent incremental lint cache in the main CLI path today. JavaScript API path canonicalization is also scoped to one `lintFiles()` call.
- **Run/Request-Scoped Source Snapshots**: CLI runs and individual API requests share immutable source text/hash snapshots across their Programs. Keys are the exact compiler-host source names, never real paths, so lexical, overlay, and symlink aliases remain distinct. Failed reads are not cached. The source layer in one cache binds to one filesystem view across its generations; compiler hosts using another view bypass this layer while retaining content-keyed AST reuse.
- **Generation-Based Fix Invalidation**: after every CLI fix write attempt, the compiler host atomically installs an empty source generation before any Program rebuild. Swapping generations rather than clearing a live map prevents an older in-flight read from repopulating the new generation. The API fix path only returns output and does not mutate or rebuild its overlay, while LSP remains version/didChange-driven.
- **Run-Scoped Parse Reuse**: CLI Program rebuilds within one invocation and Programs within one API request share the existing content-keyed AST parse cache. Source-generation invalidation does not clear AST entries, so unchanged bytes can reuse their `SourceFile`. The cache is discarded with its run/request and is never repository-persistent or shared across lint requests.
- **Bounded Multi-Pass Fixing**: `--fix` and LSP `fixAll` intentionally rerun lint after applying edits, but cap the cascade at `maxFixPasses = 10`

### Memory Management

- **ts-go Owns the Heavy Graphs**: AST nodes, checker state, `Program` graphs, and session state are primarily owned by ts-go; rslint adds listener maps, diagnostics, and config-derived rule lists on top
- **Short-Lived Per-File Structures**: comment stores, disable managers, and registered listener maps are allocated per file and dropped after traversal. A comment slice is allocated only if requested; `clear(registeredListeners)` helps release references promptly
- **Source Snapshot Ownership**: snapshot entries hold an immutable source string plus its 128-bit hash without explicitly copying source bytes; on an AST miss, that string is passed directly to the parser. After generation replacement, a retained unchanged AST may still hold the prior equal string while the fresh snapshot owns the new read. Replaced generations are reclaimed after any in-flight lookup releases them. AST retention and source-generation retention remain deliberately separate lifecycles.
- **Fix Application Uses Linear Rebuilds**: `ApplyRuleFixes` sorts fixes, skips overlapping edits, and rebuilds the output with `strings.Builder` rather than mutating source buffers in place
- **Bounded Queues**: CLI diagnostics use a buffered channel of 4096 items; LSP request/outgoing queues are buffered to 100, and debounce/refresh signals are single-slot channels
- **No Repo-Local Pooling Layer Today**: there is no explicit `sync.Pool`-based object pooling strategy in the main lint path at the moment
- **Fresh ESM Entry Lifetime**: fresh JS/TS config loads use a unique entry-module URL so rewritten bytes and module side effects are evaluated per transaction. Node retains those ESM module namespaces for the process lifetime, so a long-lived native API process can grow this cache slowly across repeated lint requests; static transitive imports continue to use Node's ordinary cache. Bounding this without weakening freshness requires a disposable evaluator realm or worker and remains a future optimization.
- **Garbage Collection Handles Cycles**: the repository does not implement custom cycle breaking for AST/checker graphs; lifecycle cleanup relies on Go GC and on dropping references after each run

## 11. Extensibility & Future Directions

### Plugin Architecture

Plugin execution has two paths:

- native plugins are compiled into the Go binary and execute through the shared listener traversal
- third-party ESLint plugin objects are loaded from JS/TS config on the Node side; Go registers routing placeholders for their rules and sends per-file batches back to the Node plugin worker over reverse IPC

JSON config supports only native plugin names because it cannot represent live JavaScript plugin objects. The repository currently ships native implementations for TypeScript ESLint, Import, Jest, JSX accessibility, Promise, React, React Hooks, and Unicorn rule namespaces.

### Rule Extension Points

- **Core Rules**: add a package under `internal/rules/<rule_name>/` and append the rule var to `internal/rules/all.go`'s `GetAllRules()` slice
- **Native Plugin Rules**: add a package under `internal/plugins/<plugin>/rules/<rule_name>/` and append the rule var to `internal/plugins/<plugin>/all.go`'s `GetAllRules()` slice
- **Third-Party Plugin Rules**: import a plugin object in JS/TS config and mount it under an object-form `plugins` prefix; no Go rule registration is required
- **Rule Options**: each rule receives parsed options through `Run(ctx, options)`
- **Custom Listener Shapes**: rules can listen on standard kinds and synthetic pattern/exit kinds

### Integration Points

- **Language Server**: `internal/lsp` exposes diagnostics and code actions
- **JavaScript API**: `packages/rslint` talks to `cmd/rslint --api` through the versioned `2.0.0` protocol; the handshake negotiates reverse `pluginLint` support before third-party rules run
- **WASM Playground**: `packages/rslint-wasm` runs the API server in a browser worker
- **Rust Client**: `crates/tsgo-client` consumes `cmd/tsgo`

### Future Enhancements

The current architecture already leaves room for:

- broader rule coverage
- richer editor features on top of the existing LSP/session foundation
- more shared tooling between CLI, API, LSP, and Playground

## 12. Testing Strategy

### Test Organization

- **Go Unit Tests**: colocated `*_test.go` files under `internal/...`
- **Rule Engine Tests**: `internal/linter` and `internal/config` have focused behavior tests
- **Go Rule Testing Helpers**: `internal/rule_tester`
- **JS/TS Integration Tests**: `packages/rslint/tests` and `packages/rslint-test-tools/tests`
- **VS Code Extension Tests**: `packages/vscode-extension/__tests__`
- **Rust / tsgo Tests**: `crates/tsgo-client/tests` and `cmd/tsgo/semantic_test.go`

### Rule Testing

Rules are tested in more than one style depending on where they live:

- direct Go unit tests for core engine behavior
- rule-focused tests under rule directories
- cross-ecosystem compatibility tests through `packages/rslint-test-tools`
- `@typescript-eslint/rule-tester`-based tests through `packages/rule-tester`

### Test Data Management

- **Fixtures**: live across Go, JS, and Rust test directories
- **Snapshots**: used in several JS and Rust integration tests
- **Virtual Configs / VFS Inputs**: used heavily for API, CLI, and type-checking scenarios

### Continuous Integration

- **Go Tests**: `pnpm run test:go`
- **TypeScript / JS Tests**: `pnpm run test`
- **Linting**: `pnpm run lint` and `pnpm run lint:go`
- **Build Verification**: `pnpm run build`

## 13. Adding a New Rule (Checklist)

The maintained rule-porting workflow now lives under [`.agents/skills/port-rule`](./.agents/skills/port-rule/).

Use these entry points instead of duplicating a separate checklist here:

- [`.agents/skills/port-rule/SKILL.md`](./.agents/skills/port-rule/SKILL.md): primary skill entry and workflow
- [`.agents/skills/port-rule/references/PORT_RULE.md`](./.agents/skills/port-rule/references/PORT_RULE.md): detailed end-to-end porting guide
- [`.agents/skills/port-rule/references/QUICK_REFERENCE.md`](./.agents/skills/port-rule/references/QUICK_REFERENCE.md): commands, naming conventions, and condensed checklist

If the rule-porting workflow changes, update the material under `.agents/skills/port-rule` rather than reintroducing a second checklist in this document.

## 14. Dependency Layering & Boundaries

### Layer Architecture

```
┌───────────────────────────────────────────────────────────┐
│ CLI / API / LSP / Website / WASM / tsgo                   │  ← cmd/, packages/, website/, crates/
├───────────────────────────────────────────────────────────┤
│ Configuration / Program Assembly / IPC / Inspector        │  ← internal/config/, cmd/rslint/programs.go,
│                                                           │     internal/api/, internal/inspector/,
│                                                           │     internal/utils/
├───────────────────────────────────────────────────────────┤
│ Linter Core                                               │  ← internal/linter/
├───────────────────────────────────────────────────────────┤
│ Rule Framework                                            │  ← internal/rule/
├───────────────────────────────────────────────────────────┤
│ Individual Rules and Plugins                              │  ← internal/rules/, internal/plugins/
├───────────────────────────────────────────────────────────┤
│ TypeScript-Go Bridge                                      │  ← shim/, tools/
├───────────────────────────────────────────────────────────┤
│ typescript-go                                             │  ← typescript-go/
└───────────────────────────────────────────────────────────┘
```

### Dependency Rules

- **Upward Dependencies**: Lower layers never depend on upper layers
- **Rule Isolation**: Individual rules only depend on rule framework
- **TypeScript Boundary**: All TypeScript integration goes through typescript-go
- **No Circular Dependencies**: Enforced by Go module system

### Key Interfaces

- **Config → Registry**: map merged config into enabled `ConfiguredRule` values
- **Programs → Linter**: ts-go `Program` instances define the compilation contexts the linter runs against
- **Rules → RuleContext**: rules interact only through the framework-provided context/reporting API
- **Integrations → Linter / Inspector**:
  - CLI/API/LSP use the linter
  - Playground inspection uses the inspector

## 15. Data Flow (Textual Diagram)

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                             CLI / API PATH                                   │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  JS/TS Configs or JSON Configs                                               │
│            │                                                                 │
│            ▼                                                                 │
│  Config Load / Normalize / Catalog                                           │
│            │                                                                 │
│            ├───────────────► CLI --type-check-only                           │
│            │                          │                                       │
│            │                          ▼                                       │
│            │                 Effective-catalog Program Registry              │
│            │                          │                                       │
│            │                          ▼                                       │
│            │                 Program-wide Type Check                         │
│            │                 (real tsconfigs)                                │
│            │                          │                                       │
│            │                          ▼                                       │
│            │                 CLI formatter / exit code                       │
│            │                                                                 │
│            └───────────────► Lint path (CLI / API)                           │
│                                       │                                      │
│                                       ▼                                      │
│  Stable Lint Target Plan (scope + selectors + ignores)                       │
│            │                                                                 │
│            ▼                                                                 │
│  Run-scoped Program Registry                                                 │
│  (active governing configs; effective catalog for CLI --type-check)          │
│            │                                                                 │
│            ▼                                                                 │
│  Bind by Governing Config / Project Order ─────► Non-project Fallback        │
│                           │                                                  │
│                           ▼                                                  │
│  Resolve / Merge Config and Enabled Rules Per File                           │
│            │                                                                 │
│            ▼                                                                 │
│  Run Rule Initializers -> Register Listeners                                 │
│            │                                                                 │
│            ▼                                                                 │
│  Single DFS AST Traversal -> Listener Dispatch                               │
│            │                                                                 │
│            ▼                                                                 │
│  RuleDiagnostic / Fix / Suggestion Collection                                │
│            │                                                                 │
│            ├───────────────► CLI --type-check: Program-wide Type Check       │
│            │                    (real tsconfigs)                              │
│            ├───────────────► CLI formatter / exit code                       │
│            └───────────────► API structured response                         │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────────┐
│                                LSP PATH                                      │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  VS Code Extension (shared UI / channels)                                   │
│     ├── WorkspaceRslintCoordinator (URI identity + generations)             │
│     ├── WorkspaceDocumentRouter (longest-ready-root ownership)               │
│     └── Rslint runtime per active root                                       │
│             └──────── rslint/configRefresh ────────────────┐                 │
│             ◄──────── load / activate / commit / abort     │                 │
│                                                            ▼                 │
│                                             cmd/rslint --lsp per root         │
│     │                                                                        │
│     ▼                                                                        │
│  internal/lsp + ts-go project.Session                                        │
│     │                                                                        │
│     ▼                                                                        │
│  LintSingleFile on session Program (per-file LSP path)                       │
│     │                                                                        │
│     ▼                                                                        │
│  LSP Diagnostics / Quick Fix / Fix All                                       │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────────┐
│                           PLAYGROUND / WASM PATH                             │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  website Playground                                                          │
│     │                                                                        │
│     ▼                                                                        │
│  packages/rslint-wasm worker                                                 │
│     │                                                                        │
│     ▼                                                                        │
│  wasm cmd/rslint --api                                                       │
│     │                                                                        │
│     ├───────────────► internal/linter     -> diagnostics / encoded sources   │
│     └───────────────► internal/inspector  -> node/type/symbol/flow info      │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

## 16. Glossary

- **AST**: Abstract Syntax Tree produced directly by ts-go
- **Code Action**: LSP action derived from diagnostics, suggestions, or bulk-fix operations such as quick fix and fix all
- **Comment Store**: short-lived per-file provider that lazily computes and shares the canonical source comment list
- **Config Entry**: One flat-config object whose `files`, `ignores`, `settings`, and `rules` participate in per-file config merging
- **ConfiguredRule**: Rule implementation plus resolved severity, settings, options, and type-info requirement
- **Diagnostic**: A lint finding reported by a rule or by TypeScript semantic diagnostics
- **Fallback Program**: Extra non-project-backed `Program` created from selected lint targets that are not covered by a tsconfig Program associated with their governing config; fallback files do not enable type-aware rules or participate in program-wide type-check
- **Flat Config**: ESLint-style array-based configuration model used by rslint to merge rule settings per file
- **Gap File**: A selected lint target that is not present in any tsconfig Program declared by its governing config
- **Inspector**: Auxiliary backend path that returns node, type, symbol, signature, and flow information for Playground inspection
- **IPC API**: Length-prefixed JSON message protocol exposed by `cmd/rslint --api` for Node and WASM clients; config path resolution and third-party plugin routing use separate keys when API overrides rebase relative patterns
- **Listener**: Callback registered by a rule for an AST kind or synthetic listener kind
- **Nearest Config**: In multi-config mode, the governing config selected by lexical-first ownership resolution
- **Node Kind**: Enumerated AST kind value used by ts-go and by the listener dispatcher to identify node types
- **Program**: ts-go compilation context, usually tied to a tsconfig or fallback root-file set
- **Program Registry**: Run-scoped set of Programs keyed by normalized declared tsconfig path, plus the governing configs and project declaration order associated with each Program
- **project.Session**: ts-go project manager used by LSP for inferred/configured projects and overlays
- **Rule Context**: Runtime environment through which a rule reads file/program/checker state and reports findings
- **RuleFix**: Text edit represented as a range plus replacement text; fixes are merged and applied after diagnostics are collected
- **Rule Registry**: Shared registry of rule implementations and config-to-rule resolution logic; the registry is implemented in `internal/config/rule_registry.go` and populated by `RegisterAllRules()` in `internal/config/config.go`
- **RuleSuggestion**: Suggested edit attached to a diagnostic that is surfaced to the user but not treated as a default autofix
- **Severity**: Effective diagnostic level for a configured rule, such as `off`, `warn`, or `error`
- **Source Code Fixer**: Fix-application layer that merges non-overlapping `RuleFix` edits and rewrites file contents
- **Synthetic Listener Kind**: rslint-defined pseudo-kind such as `OnExit`, `OnAllowPattern`, or `OnNotAllowPattern` used to distinguish traversal contexts beyond raw AST kinds
- **TypeChecker**: ts-go semantic engine acquired from a `Program` and used by type-aware rules for symbol and type queries
- **Type-aware Rule**: Rule that requires the TypeChecker and semantic information
- **TypeScript-Go**: Go port of the TypeScript compiler that supplies AST, checker, Program, project/session, scanner, and VFS
- **Overlay VFS**: In-memory filesystem layer used by API, LSP, and browser scenarios
- **Workspace**: Set of related files, config roots, and projects considered together by CLI, LSP, or editor integrations

## 17. TODO / Open Questions

### Implementation Details Needed

- [ ] Document specific concurrency patterns and worker pool implementation
- [ ] Detail memory management strategy and object pooling
- [ ] Explain caching mechanisms for TypeScript compilation and rule results
- [ ] Document error recovery strategy in parser
- [ ] Clarify node ID system and interning strategy if present

### Feature Documentation

- [ ] Document inline directive support (rslint-disable / eslint-disable comments)
- [x] Explain configuration merging and inheritance rules
- [ ] Detail source map support for transformed files
- [ ] Document plugin architecture and extensibility model
- [ ] Clarify incremental linting capabilities

### Performance Benchmarks

- [ ] Provide concrete performance comparison data vs ESLint
- [ ] Document memory usage patterns and limits
- [ ] Explain zero-copy optimization details
- [ ] Benchmark different concurrency strategies

### Integration Points

- [ ] Document JavaScript API surface and usage patterns
- [ ] Explain LSP server capabilities and VS Code integration
- [ ] Detail CI/CD integration best practices
- [ ] Document output format specifications

### Testing Strategy

- [ ] Explain golden test implementation and maintenance
- [ ] Document test data management and fixture organization
- [ ] Detail integration test coverage and automation
- [ ] Clarify performance regression testing approach

---

This architecture document should be updated as the project evolves. For questions or clarifications, please refer to the source code or open an issue on the [GitHub repository](https://github.com/web-infra-dev/rslint).
