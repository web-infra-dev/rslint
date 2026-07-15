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

| Path                           | Purpose                                                                                                                                                   | Key Relationships                                                                                                                                                                                                                                                                                   |
| ------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `website/`                     | Documentation site and Playground UI                                                                                                                      | Uses `packages/rslint-wasm` to run browser linting and `packages/rslint-api` to decode encoded source files; Playground lint requests ultimately reach `internal/linter`, and inspect requests reach `internal/inspector` through `internal/api`                                                    |
| `cmd/rslint/`                  | Main Go binary entry point with CLI, API, and LSP modes                                                                                                   | JS/TS config CLI path is `internal/config/discovery -> internal/config -> cmd/rslint/programs.go -> internal/linter`; JSON config starts at `internal/config`. `--api` is consumed by `packages/rslint` and `packages/rslint-wasm`; `--lsp` is consumed by `packages/vscode-extension`              |
| `cmd/rslint/internal/output/`  | CLI report model, summary, colors, and stdout formatters                                                                                                  | Consumes final sorted `internal/rule` diagnostics from the CLI pipeline and renders `default`, `jsonline`, `github`, or `gitlab`; it is intentionally private to `cmd/rslint` and is not shared by the structured API or LSP adapters                                                               |
| `cmd/tsgo/`                    | ts-go semantic inspection/export tool                                                                                                                     | Talks directly to `typescript-go` and bypasses the lint framework; consumed by `packages/tsgo` and `crates/tsgo-client`                                                                                                                                                                             |
| `internal/api/`                | stdio IPC protocol and service types for JS/WASM integration                                                                                              | Shared protocol layer for `cmd/rslint --api`; used by `packages/rslint`, `packages/rslint-wasm`, `internal/linter`, and `internal/inspector`                                                                                                                                                        |
| `internal/config/`             | Configuration models, JSON loading, ConfigArray evaluation/merging, runtime ownership resolution, lint-target planning, and centralized rule registration | Owns string/function selector semantics, ordered ignores, exact decision caches, and the lazy product-level `.gitignore` layer. `RegisterAllRules()` orchestrates rule registration; `rule_registry.go` implements registry/query logic used by `cmd/rslint/programs.go` and `internal/linter`      |
| `internal/config/discovery/`   | Go-owned ESLint-style input planning, config-aware target search, JS/TS candidate selection, and immutable catalog construction                           | Imports the parent config evaluator, issues unary module loads plus batched live-predicate calls to a host-supplied Node loader, and returns lexical targets with their sole merged config result. CLI, API, and LSP call its one-shot `Build`; the parent package never imports this child package |
| `internal/config/gitignore/`   | Config-scoped `.gitignore` source discovery and pattern conversion                                                                                        | Supplies the evaluator's product-level first ignore layer. Sources are read only along exact target/directory ancestry, cached per owner/source directory, and never require a second target post-filter or an LSP owner-subtree sweep                                                              |
| `internal/hostpath/`           | Native host-path syntax and lexical identity                                                                                                              | Preserves legal POSIX backslash bytes while also supporting synthetic drive/UNC roots in cross-platform tests; configuration discovery and diagnostics use this path space                                                                                                                          |
| `internal/hostfs/`             | Native-path adapter for the ts-go VFS interface                                                                                                           | Keeps host filesystem operations from applying ts-go's platform-independent backslash normalization; used by discovery, CLI/API, and LSP snapshots                                                                                                                                                  |
| `internal/compilerpath/`       | Compiler-only aliases for host paths ts-go cannot represent                                                                                               | Maps POSIX backslash filenames to unique temporary compiler paths while retaining the original lexical path for configuration, diagnostics, API results, and fix output                                                                                                                             |
| `internal/inspector/`          | AST/type/symbol/signature/flow inspection for Playground                                                                                                  | Auxiliary backend used mainly by website Playground inspect panels; builds rich semantic data from `typescript-go` programs                                                                                                                                                                         |
| `internal/linter/`             | Core lint engine, traversal, and fix application                                                                                                          | Consumes rules from `internal/rule`, file config from `internal/config`, and `Program` / `TypeChecker` data from `typescript-go`; also serves `internal/api` and `internal/lsp`                                                                                                                     |
| `internal/lsp/`                | Language Server Protocol implementation                                                                                                                   | Wraps `typescript-go project.Session`, owns transactional config discovery and last-good commit state with `packages/vscode-extension` as the module/plugin host, and invokes `internal/linter` on session-backed programs                                                                          |
| `internal/rule/`               | Rule framework, context, diagnostics, fixes, and disable manager                                                                                          | Shared foundation for core rules and plugin rules; called by `internal/linter` through listeners and reporting APIs                                                                                                                                                                                 |
| `internal/rule_tester/`        | Go-side rule testing helpers                                                                                                                              | Supports rule development and complements JS-side testers in `packages/rule-tester` and `packages/rslint-test-tools`                                                                                                                                                                                |
| `internal/rules/`              | Core lint rule implementations without plugin namespace; `all.go` aggregates them into the `GetAllRules()` slice                                          | `internal/config/config.go`'s `RegisterAllRules()` consumes the slice and registers each rule; then executed by `internal/linter` like plugin rules                                                                                                                                                 |
| `internal/plugins/typescript/` | `@typescript-eslint`-style rules                                                                                                                          | Registered into the shared rule registry by `RegisterAllRules()` and often rely on `TypeChecker` from `typescript-go`                                                                                                                                                                               |
| `internal/plugins/react/`      | React rule implementations                                                                                                                                | Registered into the shared rule registry by `RegisterAllRules()` and executed through the same listener pipeline in `internal/linter`                                                                                                                                                               |
| `internal/plugins/jest/`       | Jest rule implementations                                                                                                                                 | Registered into the shared rule registry by `RegisterAllRules()` and executed through the same listener pipeline in `internal/linter`                                                                                                                                                               |
| `internal/plugins/import/`     | Import plugin registration and rules                                                                                                                      | Contributes plugin rules through `RegisterAllRules()` and participates in normal config-driven linting                                                                                                                                                                                              |
| `internal/utils/`              | Shared utilities for JSONC, compiler hosts, overlay VFS, and helpers                                                                                      | Supports `cmd/rslint/programs.go`, config loading, and various linter entry points                                                                                                                                                                                                                  |
| `packages/rslint/`             | Main npm package with JavaScript API and CLI wrapper                                                                                                      | Spawns `cmd/rslint --api` in JavaScript runtime environments and uses `internal/api` message shapes                                                                                                                                                                                                 |
| `packages/rslint-api/`         | Frontend-facing encoded source file / AST decoding helpers                                                                                                | Used mainly by website Playground to decode AST/source data returned from the Go API                                                                                                                                                                                                                |
| `packages/rslint-test-tools/`  | Testing utilities and cross-ecosystem rule tests                                                                                                          | Supports package-side and integration-style tests around the linter and rule ecosystem                                                                                                                                                                                                              |
| `packages/rslint-wasm/`        | Browser/WASM package for running `rslint --api` in a worker                                                                                               | Starts the browser worker, hosts the wasm runtime, and bridges website Playground requests to `internal/api`, `internal/linter`, and `internal/inspector`                                                                                                                                           |
| `packages/rule-tester/`        | Forked `@typescript-eslint/rule-tester` package used in tests                                                                                             | JS-side rule testing support that complements Go-side helpers                                                                                                                                                                                                                                       |
| `packages/utils/`              | Shared JavaScript utilities                                                                                                                               | Shared support package for the JS/website tooling layer                                                                                                                                                                                                                                             |
| `packages/vscode-extension/`   | VS Code extension for IDE integration                                                                                                                     | Launches `cmd/rslint --lsp`, starts `rslint/configRefresh`, serves reverse load/activate/commit/abort and plugin-lint requests, and consumes diagnostics/code actions from `internal/lsp`                                                                                                           |
| `packages/tsgo/`               | JS wrapper package for the `tsgo` tool                                                                                                                    | JavaScript-facing wrapper around `cmd/tsgo` output                                                                                                                                                                                                                                                  |
| `typescript-go/`               | Git submodule containing TypeScript compiler Go port                                                                                                      | Provides parser, AST, checker, `Program`, `project.Session`, diagnostics, scanner, and VFS primitives used throughout the backend                                                                                                                                                                   |
| `shim/`                        | Generated bridge packages exposing ts-go internals                                                                                                        | Bridge layer between repository Go code and `typescript-go` internals; generated and updated by `tools/`                                                                                                                                                                                            |
| `tools/`                       | Shim generator and ts-go update scripts                                                                                                                   | Generates `shim/` code and maintains the pinned `typescript-go` integration                                                                                                                                                                                                                         |
| `crates/tsgo-client/`          | Rust client for communicating with `cmd/tsgo`                                                                                                             | Spawns `cmd/tsgo` and consumes its semantic/project output from Rust                                                                                                                                                                                                                                |

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

1. **Source Text Loading**: Files come from the real filesystem, an overlay VFS, or LSP document overlays. Initial CLI builds and individual API requests keep a compiler-host source snapshot keyed by the exact normalized source path. The snapshot stores the first successful text read and its xxh3 hash for the current generation, allowing later Programs in that view to skip both work items. Each CLI lexical fix target rebuilds against a private overlay and fresh parse cache; after all completed outputs are written, the initial cache advances its source generation once. API snapshots end with the request. LSP does not use this one-shot layer because its `project.Session` already follows document versions and `didChange` updates.
2. **Program Construction**: Plain CLI/API lint resolves its effective targets first and builds `Program` objects only for governing configs that own a selected target. Program-wide type-check modes build every project declared by the effective loaded config catalog. LSP reuses matching projects from ts-go `project.Session`; a declared custom tsconfig that project service has not loaded is built on demand against an isolated editor overlay.
3. **Lexical + Syntax Parsing**: ts-go tokenizes and parses source files into TypeScript-native AST nodes.
4. **Semantic Analysis**: When needed, the linter acquires a `TypeChecker` from the `Program`.
5. **Rule Registration**: Enabled rules register listeners keyed by AST kind.
6. **AST Traversal**: The linter traverses each file once using a DFS walk.
7. **Rule Execution**: Listener callbacks inspect nodes and may use syntax only or syntax plus type information.
8. **Diagnostic Collection**: Findings are reported as `RuleDiagnostic` values, with optional fixes or suggestions.
9. **Output Generation**: CLI builds one report from the final post-fix diagnostics and passes it to `cmd/rslint/internal/output`; API returns structured data, and LSP converts diagnostics to LSP diagnostics/code actions.

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
- **Comments**: collected separately and used for directives and comment-based rules

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

The linter parses `/* global */` comments once per file before rules run.
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
| `basePath`        | `string`                                   | Per-entry matching root, resolved from the effective config base                     |
| `files`           | `(string \| function \| matcher[])[]`      | Non-empty selector list; top-level selectors are ORed and nested selectors are ANDed |
| `ignores`         | `(string \| function)[]`                   | Glob or live-predicate matchers excluded by this entry                               |
| `languageOptions` | `object`                                   | Parser options such as `project` and `projectService`                                |
| `rules`           | `Record<string, …>`                        | Rule level or `[level, options]`                                                     |
| `plugins`         | `string[] \| Record<string, ESLintPlugin>` | Native plugin declarations or third-party plugin instances                           |
| `settings`        | `Record<string, …>`                        | Shared settings available in `RuleContext`                                           |

### Configuration Loading

The loading flow differs by config type:

**JS/TS staged catalog discovery**:

CLI, the native JavaScript API path, and transactional LSP refreshes reuse the
one-shot `internal/config/discovery.Build` operation. Go receives the raw lint
inputs, performs stat-first file/directory/glob classification, groups glob
patterns by their ESLint-compatible glob parent, walks the filesystem, selects
the nearest config for each visited location, applies configuration matching,
and returns an immutable catalog containing exact lint targets and their
owners. Node does not walk the filesystem. It only evaluates and normalizes the
exact JS/TS config modules requested by Go and retains live third-party plugin
objects for the final activation.

The package boundary is deliberate: `internal/config/discovery` owns search
planning, target-pattern matching, nearest-config selection, and immutable
catalog construction. It imports the parent `internal/config` model and the
ordered global-ignore matcher; the parent package never imports discovery.
The discovery evaluator applies product-level `.gitignore`, defaults, authored
entries, and live functions in one ordered selection and stores that merged
result on every target. CLI/API construct Programs directly from those exact
target/owner/result triples. The LSP adapter retains the catalog evaluator and
Node predicate session for files opened after commit, and owns transactional
commit/abort plus last-good lifecycle. A process nonce plus an atomic sequence
allocates transaction IDs that cannot collide with stale Node host state after
a native-process restart.

Automatic discovery follows ESLint v10's `findFiles()` and `ConfigLoader`
shape (apart from rslint's config filename suffix set):

1. Existing inputs are classified with `stat` before glob syntax is considered.
   Existing files become direct targets. Existing directories become `/**`
   searches. Nonexistent glob-shaped inputs are grouped by `glob-parent`; equal
   lexical bases share one search and nested bases remain independent. The cwd
   search is pre-seeded so search and unmatched-error order matches ESLint.
2. Each search walks only directories that can still match at least one input
   pattern. A directly named directory is already admitted, so an ancestor
   config cannot block entry into that root. A dynamic glob must pass the
   current parent config's directory-ignore gate before traversal. Nested
   directory symlinks are not followed.
3. For each visited directory or file, Go searches lexically upward for the
   nearest automatic config candidate. Filename priority is `.js` → `.mjs` →
   `.ts` → `.mts`. Results are cached per directory and config module, so files
   sharing an owner neither repeat the upward search nor reload the module.
   Failure of the nearest candidate is fatal; discovery never skips a broken
   candidate to use an ancestor config. No realpath/canonical fallback
   participates in config selection.
4. Rslint's product-level `.gitignore` layer is reduced first, followed by
   ESLint's default `.git` and `node_modules` entry, module config, and
   `overrideConfig`. Later authored negations may reopen a file or directory
   ignored by either earlier layer, including traversal into `.git` or
   `node_modules`. Git's own parent-prune rule still prevents a descendant
   `.gitignore` source from being read below a Git-ignored directory; reopening
   that directory does not retroactively activate the unread source within the
   same governing owner. A nested config ownership handoff starts a new hard
   `.gitignore` root at that owner's config directory. Sources are loaded lazily
   only along the exact file/directory ancestry being evaluated;
   nearest-config lookup for an exact location is completed before that
   location's file status is tested.
5. Every visited file is tested exactly once against the effective flat config rooted at its
   selected config directory (or at invocation cwd for an explicit/inline
   config). Direct files retain `configured`, `ignored`, `unconfigured`, or
   `external` provenance so CLI/API can produce ESLint-style warning results.
   The complete ConfigArray decision is cached by exact path, and configured
   targets carry the resulting merged config. No later target walker or
   native/plugin/fix consumer may widen the set, recompute ownership, or invoke
   a live matcher again.
6. Independent explicit files and glob-parent searches run concurrently unless
   `--singleThreaded` is set. Each individual filesystem search remains a stable
   lexical DFS. A newly reached config candidate starts one unary load; Go
   shares the per-path in-flight state, while Node shares both candidate
   promises and raw module source by physical identity. Independent unary loads
   may therefore evaluate concurrently without a discovery-wide ordering
   barrier. Go validates each transaction ID, candidate ID, status, and
   normalized config shape before publishing the candidate. The first fatal
   outer-member error (lookup, explicit input, or the combined glob-search
   member) returns fail-fast, cancels sibling members, and transfers eventual
   worker cleanup to a deferred waiter when an in-flight host promise cannot be
   interrupted immediately. Independent glob searches all-settle inside their
   combined member so insertion-order error selection remains deterministic.
7. `activateConfigs` names only effective config IDs. Node rechecks source
   fingerprints, prepares plugin state for precisely that set, and rechecks the
   same sources before publication, preventing normalized entries from being
   paired with different config bytes or plugin topology.

`.gitignore` is an intentional rslint product extension, but it is not a second
discovery phase. `ConfigEvaluator` requests the sources on an exact target or
directory chain, compiles them as the first ordered global-ignore layer, and
caches both sources and complete decisions. This cannot change the already
selected owner. Because authored entries follow the synthetic layer, their
negations can re-include a gitignored path. LSP retains the same evaluator and
its snapshot filesystem, so a later document in an already committed owner is
evaluated on demand without pre-scanning that owner's subtree.

The surfaces differ only in orchestration:

- CLI sends raw positionals to Go and hosts reverse `loadConfigs`,
  `evaluateConfigPredicates`, and `activateConfigs` requests. If no JS/TS config
  participates, the existing JSON/JSONC loader and target flow remain the
  intentional legacy fallback.
- `Rslint.lintFiles()` and `lintText()` send raw inputs to Go. The native API
  handles reverse config loads directly; low-level pre-resolved `config`
  requests and WASM do not use host-filesystem discovery. A catalog with
  object-form plugins additionally requires `reversePluginLint`. API config
  imports use Node's ordinary cached-module semantics; their Go catalog and
  Program state remain request-scoped.
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
- Each runtime starts `rslint/configRefresh`. Go searches that workspace root's
  cwd with `CollectTargets=false`, an explicit cwd-owner probe, and exact lookup
  paths for open documents. The probe lets an empty workspace commit the config
  that governs future open documents; it is an editor-only lifecycle behavior,
  not part of CLI/API `findFiles()` semantics. The extension stages Node state
  through `rslint/loadConfigs` and `rslint/activateConfigs`, services live
  matchers through `rslint/evaluateConfigPredicates`, then commits or aborts it
  through `rslint/commitConfigs` / `rslint/abortConfigs`. `fresh` loads
  cache-bust the config entry module; static transitive imports retain Node's
  normal module cache. If initial plugin preparation detects a source change
  between its fingerprint checks, the extension keeps the language client alive
  and retries one serialized refresh from the new bytes.
- If `vscode-languageclient` automatically restarts the native process, its
  later `Running` transition first aborts any extension-side orphaned
  transaction, then requests a new initial catalog through the same serialized
  refresh chain. The previously committed plugin host remains available until
  the replacement Go process commits its own catalog.
- Only a fully committed Go/Node snapshot replaces a usable last-good snapshot.
  A newly failing boundary aborts and preserves that snapshot; failures already
  covered by committed unavailable boundaries may be recommitted as unavailable
  while independent successful owners refresh. On initial startup, successful
  siblings commit and failed nearest-config boundaries become unavailable; if
  every JS config fails, Go commits empty Node plugin state plus those boundaries
  so JSON fallback cannot leak into the broken subtrees. A committed predicate
  session stays live, and Node retains one rollback predecessor: if the commit
  response is lost, Go's abort restores it; the next successful commit retires
  the obsolete predecessor. Open documents remain separate exact-file targets
  resolved by the committed catalog evaluator.

The LSP config wire exposes one identity, `transactionId`. The extension reuses
that value internally as the `PluginLintPool` host generation so an in-flight
plugin request is routed to the exact worker state paired with Go's committed
catalog. This is a concurrency/lifecycle identity, not a second config-discovery
model. The independent numeric document generation only rejects stale async
diagnostics after edits, closes, or config commits.

An explicit JS/TS `--config` or API `overrideConfigFile` bypasses automatic
candidate selection and loads the exact module. The invocation cwd remains its
matching directory. Exact module loading itself is not filtered by
`.gitignore`. Automatic traversal, however, is admitted by the current owner's
full evaluator, including `.gitignore`, so a nested candidate inside a pruned
subtree is never reached. Direct files and LSP lookup paths resolve their exact
owner without widening a workspace traversal.

No-candidate behavior is surface-specific. CLI performs no Node activation and
continues through its normal JSON fallback. Native high-level API discovery
treats a missing automatic/explicit config as fatal and never searches disk for
JSON fallback; callers that intentionally disable discovery use inline
`overrideConfig` or an empty syntax-only config. LSP explicitly stages and
commits an empty plugin-host state (an empty load batch followed by zero-ID
activation), while loading any JSON fallback in Go as part of the new snapshot.
That empty catalog is not a usable JavaScript last-good boundary: if a newly
created JS config is broken, LSP commits an unavailable boundary for it rather
than silently retaining JSON fallback below it.

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

1. entries containing only `ignores` plus optional `name` / `basePath` metadata form the global-ignore set; each entry resolves its patterns from its own matching root
2. the implicit default extension baseline plus effective explicit `files` selectors defines the config selector union; an entry's `ignores` prevents its selector from extending that union, top-level selectors are ORed, nested patterns are ANDed, and an explicit `files: []` is invalid
3. entries without `files` cascade across that selector union, while entry-level `ignores` prevent only that entry from contributing configuration to an otherwise selected file
4. later rule values override earlier values; a severity-only override retains earlier rule options
5. settings and language options recursively merge ordinary objects, while later arrays and scalar values replace earlier values

Discovery resolves lexical ownership and performs the sole ConfigArray merge
before committing each target. `ConfigOwnerResolver` remains the indexed
runtime/on-demand adapter for retained catalogs and the JSON path; it is not a
second ownership pass over discovery targets. Staged CLI, native API, and
transactional LSP paths therefore reuse the same Go ownership rules instead of
reconstructing hierarchy on the Node side.

Discovery ownership and directory handoff boundaries are exclusively lexical.
The runtime `ConfigOwnerResolver` first tries that lexical catalog and consults
canonical ancestry only when no lexical owner exists, so it never compares
depth across the two spaces. Canonical identity may also recover a target's
relative location when a TypeScript Program reports a physical source name;
matching is then projected back under the authored config directory. This keeps
absolute `basePath` values, diagnostics, and lexical aliases in one path space.
Distinct lexical config directories remain distinct owners even when they
resolve to the same physical directory.

Additional current behaviors:

- `.gitignore` is injected lazily as the first global-ignore layer through the
  shared `ConfigEvaluator` policy. The governing config directory is a hard
  upper boundary: sources below that owner apply, while parent sources do not.
  Each file or directory query reads only its exact ancestry and caches the
  compiled layer by source directory; sibling searches cannot change an
  existing decision. The synthetic Git layer precedes defaults and authored
  entries, so a later config `!` may re-include a target
- when the client supports dynamic file-watch registration, Go watches
  workspace-descendant `.gitignore` files plus exact `.gitignore` paths in
  strict workspace ancestors that may contain an automatically selected config.
  Extension watchers are the sole refresh owner for
  workspace/descendant JS configs, JSON fallback, and dependency lockfiles;
  Go additionally watches only strict-ancestor JS configs and `.gitignore`.
  ts-go project watchers may still forward the same workspace events into the
  session, but those forwarded JS/JSON events do not start a second fresh config
  transaction. Create/change/delete events rebuild the config/ignore generation and
  refresh open-document diagnostics
- the VS Code extension preserves last-good JS configs during reloads; a newly
  unavailable config with no usable JS ancestor contributes an empty boundary,
  preventing JSON fallback only in that authored config subtree. A normal
  transactional refresh prepares all successful and unavailable boundaries,
  then commits the Go catalog/evaluators and Node plugin host under one
  transaction ID. Failures preserve a usable last-good catalog and ignore view
  together; the first-start all-broken
  recovery instead commits unavailable boundaries.
  If the first valid catalog cannot initialize its optional community-plugin
  worker, LSP commits the ordinary native config with an empty no-host plugin
  state and retries on later refreshes; once a usable snapshot exists, the same
  worker failure aborts and preserves that last-good snapshot. A successful
  no-candidate transaction removes the previous JS catalog and
  exposes the Go-loaded JSON fallback
- native and third-party plugin rules are gated by their normalized prefixes for JS/TS configs; third-party rules use process-wide Go registry placeholders, but LSP additionally filters those placeholders through the exact rule-name set committed for the current Node generation so metadata retained from an older generation cannot be dispatched to a newer worker
- CLI/API lint target selection is independent from TypeScript `Program` membership. The `.js`, `.mjs`, `.cjs`, `.jsx`, `.ts`, `.tsx`, `.mts`, and `.cts` suffixes form the implicit default baseline; a specific authored `files` selector may additionally select a custom suffix, including an explicitly named custom file. Global ignores and `.gitignore` remove targets, while an entry-level ignore prevents only its own selector/config contribution
- selected CLI/API targets can still appear as 0-rule lint results when no config entry contributes rules; this applies to default-baseline files, authored custom-suffix selectors, and explicit configured files, and syntax diagnostics remain available in that state
- under automatic discovery, each selected file is governed by its nearest config candidate; a broken nearest candidate is fatal rather than skipped for an ancestor. Explicit config modes use the selected config directly. In either mode, a target can bind only to a tsconfig declared by its governing config, and the first declared project containing the file wins
- `files`/`ignores` matching uses the stable target path in the governing config's path space; a Program source alias is used only to locate the AST and type information, so moving a target into or out of a tsconfig cannot change its rule configuration
- within each Program-registry build, normalized declared tsconfig paths are deduplicated across config associations; CLI fix passes create a new registry build. File-symlink declarations remain distinct because TypeScript resolves relative paths from the declared location. Selected files outside the governing config's Programs receive a non-project-backed fallback Program, and targets whose names collide under a case-insensitive ts-go path key are partitioned across fallback Programs so distinct physical files remain distinct
- `--type-check` and `--type-check-only` build every real tsconfig declared by the effective loaded config catalog. Once that catalog is established, program-wide checking is not filtered by lint targets, config `files`/`ignores`, `.gitignore`, or CLI file/directory arguments; synthetic fallback Programs never participate. `--type-check-only` skips target retention and per-file lint selection, but its catalog traversal still evaluates directory admission and therefore lazily reads relevant `.gitignore` ancestry
- for LSP, an open supported script is a per-file target independent of Program membership. Global config ignores, `.gitignore`, default-excluded paths, and unavailable config boundaries suppress native rules, plugin rules, and fixes; an available zero-rule config still parses the target and can report syntax diagnostics

### Inline Directives

Rslint supports inline directives with both `rslint-` and `eslint-` prefixes:

- `// rslint-disable-next-line @typescript-eslint/no-unused-vars`
- `/* rslint-disable @typescript-eslint/no-unsafe-assignment */`
- `// eslint-disable-next-line`

The `DisableManager` in `internal/rule/disable_manager.go` parses and applies these directives before diagnostics are emitted.

## 9. CLI Flow

### Command Line Interface

```bash
rslint [options] [files...]
```

### CLI Processing Flow

The CLI has a two-layer architecture: a Node.js wrapper (`packages/rslint/src/cli/cli.ts`) and the Go binary (`cmd/rslint/`).

1. **Node.js Wrapper**: parses args, starts the Go engine, and hosts JS/TS module evaluation plus live third-party plugin objects
2. **Config Catalog**: for automatic or explicit JS/TS config mode, Go builds the staged catalog and sends an independently awaitable unary request for each newly reached module candidate to Node; duplicate lexical/physical work is coalesced on the respective side. If automatic discovery finds no candidate, or a non-JS config was explicitly selected, the existing Go JSON loader path remains in control
3. **Mode Selection**:
   - `--lsp`: starts the LSP server
   - `--api`: starts the IPC API server
   - default: runs direct CLI linting
4. **Lint Target Plan**: Go resolves a stable target set from CLI/API scope, the implicit default baseline, explicit config `files`, global ignores, and `.gitignore`
5. **Program Registry**: plain lint builds each normalized tsconfig path declared by an active governing config once; `--type-check` and `--type-check-only` instead retain every project declared by the effective loaded config catalog. Shared declared paths preserve each active config association and declaration order.
6. **Program Binding**: each target is bound by exact lexical or canonical filesystem identity to the first containing Program declared by its governing config; unbound targets, including projects with no tsconfig, are parsed through a non-project-backed fallback Program
7. **Rule Resolution**: `getRulesForFile` resolves enabled rules from the stable lint-target path, never the Program source alias, and filters type-aware rules off no-type-info gap files
8. **Rule Execution**: `RunLinter()` schedules exact `LintProgramView` work. Multiple lexical config views may share one project-backed Program without sharing target/config identity. Within each view, files are partitioned by checker-pool ownership and each shard is traversed serially by the worker that exclusively owns that checker. When `--type-check` is enabled, a separate program-wide pass over unique real tsconfig Programs aggregates `tsc --noEmit`-aligned diagnostics through `collectNoEmitDiagnostics()`
9. **Result Aggregation**: diagnostics are sent through one run-scoped diagnostics channel and collected at the CLI layer
10. **Fix Passes**: CLI `--fix` verifies each lexical target independently for at most ten passes against an isolated in-memory overlay, rebuilding and rebinding only that target with its exact merged config after each applied pass. All target results settle before output writes begin, so two aliases never combine configs in memory. Default writes remain concurrent like ESLint's `outputFixes`; `--singleThreaded` writes them serially. The API intentionally returns one in-band fix pass per lexical result and never writes it
11. **Report Assembly**: the CLI builds one output report from the final post-fix diagnostics plus run metadata. Diagnostics carry an explicit lint or TypeScript origin, and the report computes error/warning/type-error counts once so the summary and exit policy use the same values; `--quiet` filters rendering only.
12. **Output Formatting**: the CLI-private output subsystem renders `default`, JSON line, GitHub workflow command, or GitLab Code Quality formats. Only `default` emits a summary; machine-readable formats never emit ANSI styling or a summary.
13. **Exit Code**: depends on the report counts, `--max-warnings`, and fix outcomes

### Concurrency Model

The main Go workload work groups and pools below honor `--singleThreaded`.
The flag serializes these workload stages, but IPC transport, diagnostic
collection, and plugin dispatch may still use infrastructure goroutines.

1. **Linter work group** (`RunLinter()` via `core.NewWorkGroup`)
   - Schedules lexical `LintProgramView` work; multiple views may reference one
     Program. Each view partitions its files into checker-owned shards so a
     checker is never accessed concurrently.
   - `--singleThreaded` serializes views and file shards.

2. **Type-check work group** (`runTypeCheckAcrossPrograms`)
   - Schedules diagnostics for real tsconfig Programs and merges results in
     stable Program order.
   - `--singleThreaded` computes Program diagnostics serially.

3. **Staged catalog discovery and Program creation**
   - Stat calls for independent raw inputs fan out concurrently. Existing files
     become direct searches; directory/glob inputs are grouped by lexical glob
     parent exactly once.
   - Direct files and independent glob-parent searches run concurrently. Each
     individual search is a stable lexical DFS whose directory admission and
     file selection use the nearest config for that location. `.gitignore`
     sources are read lazily only when the evaluator reaches an exact
     directory/file chain.
   - Duplicate nearest-config lookups share one in-flight owner/config state.
     Each distinct candidate uses an independent unary Node request; Go
     coalesces duplicate candidate paths, Node coalesces candidate promises and
     physical module source, and each DFS waits only at its own dependency.
   - Discovery commits the exact target/owner/merged-config plan together with
     the config catalog. Plain lint constructs Programs only for owners represented by
     retained targets. `--type-check` and `--type-check-only` construct every
     Program in the effective reachable catalog; the latter skips target
     retention but uses the same config-aware traversal and directory ignore
     gates.
   - Configured Programs remain serial in stable config/project order because
     typescript-go's API is invoked one Program at a time.
   - `--singleThreaded` disables input stat fan-out and serializes direct
     file/search execution and unary Node module evaluation. Catalog/result
     reduction remains deterministic in either mode.

4. **Legacy JSON/JSONC lint-target directory walker** (`internal/config/file_discovery.go`)
   - `DiscoverLintTargets` uses a fixed-size worker pool (`walkPool`) that
     walks the directory tree for the retained JSON/JSONC path. JS/TS config
     flows do not call this walker; they consume discovery's exact targets.
     `DiscoverLintFiles` is the path-only compatibility wrapper. Live goroutine
     count is capped at `workers`, not the number of directories.
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

5. **Fix verification and output** (`runIndependentFixes`)
   - Lexical targets verify in parallel against isolated overlay/Program state;
     outputs begin only after all verification settles. Default output writes
     are independent and concurrent, matching ESLint `outputFixes` even for two
     aliases of one inode.
   - `--singleThreaded` serializes both verification and writes.

6. **Program source identity index** (`bindLintTargetPlan`)
   - Direct lexical Program lookups remain synchronous. If one misses, the
     binder resolves each unique Program source path once and builds a
     binding-pass canonical identity index. CLI fix passes rebuild it when they
     rebind the target plan.
   - Independent realpath lookups use `core.NewWorkGroup`; `--singleThreaded`
     runs the same work serially.

Other invariants:

- Discovery returns caller-visible lexical targets with their already-merged
  configs; it does not return canonical identities. After the catalog and target
  plan are fixed, the CLI/API target-plan projection resolves a physical hint
  for Program binding. The lexical path remains the result/config identity,
  including when multiple spellings resolve to one physical file; aliases are
  neither coalesced nor assigned an owner by scan order.
- JS/TS targets stay in the caller's lexical path space for pattern matching
  and nearest-config selection. Config discovery never consults physical
  ancestry. Canonical identity is resolved only after the exact target/owner
  plan is fixed, for TypeScript Program source binding.
- Direct directory roots bypass the parent directory-admission gate, while
  dynamic glob roots do not. Once inside a search, the current nearest config's
  global ignores control traversal into each child directory.
- `Rslint.lintFiles()` performs no Node-side stat, glob, realpath, or ownership
  planning; it sends the raw patterns and hosts only module/plugin execution.
- LSP uses a different orchestration model and keeps session access on its main
  dispatch loop. Its Program-source lookup follows the same exact lexical and
  canonical filesystem identity rules as CLI/API binding, including
  file-symlink aliases and rejection of case-folded nonidentical paths.

#### `--singleThreaded` semantics

`--singleThreaded` is honored in every parallelism point above:

| Point                         | Effect when set                                                                                                             |
| ----------------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| Linter work group             | Collapsed to serial via `core.NewWorkGroup(true)`.                                                                          |
| Type-check work group         | Program diagnostics are computed serially via `core.NewWorkGroup(true)`.                                                    |
| Staged catalog discovery      | Input stat, independent searches, and unary Node module evaluation are serialized; catalog reduction remains deterministic. |
| Lint-target walker workers    | Forced to 1 (single goroutine, no concurrency).                                                                             |
| Fix verification and writes   | Lexical targets are verified and their completed outputs are written serially.                                              |
| Program source identity index | Canonical source paths are resolved serially through `core.NewWorkGroup(true)`.                                             |

These workload stages run serially with `--singleThreaded`; infrastructure
goroutines remain outside that guarantee.

## 10. Performance & Memory Considerations

### Design Principles

- **Direct ts-go Data Model**: rslint operates on ts-go `Program`, AST, and `TypeChecker` objects directly instead of converting through a second AST representation
- **View/Checker Parallelism**: `RunLinter` queues lexical `LintProgramView` work through `core.NewWorkGroup`, and each view partitions files into checker-owned shards; `--singleThreaded` forces both levels to run serially
- **Single-Walk Rule Dispatch**: each file is traversed once, with rules registering listeners up front and sharing the same AST walk
- **Early Filtering**: exact lint target plans, skip paths, global-ignore filters, and gap-file type filtering reduce work before listeners run

### Performance Optimizations

- **Native Go Implementation**: Eliminates JavaScript runtime overhead
- **Direct TypeScript AST**: No AST conversion between parsers
- **Checker-Pool Sharding**: each `LintProgramView` groups files by checker-pool ownership; one worker exclusively holds each checker while processing its shard, so files and lexical views can run concurrently without concurrent access to one checker
- **Checker Phase Separation**: lint workers release their exclusive checkers before TypeScript semantic diagnostics run, so `GetSemanticDiagnostics` can acquire its checker state cleanly
- **File Filtering**: bundled compiler files are excluded. `.git` and `node_modules` are ordered default ignore entries, not unconditional skips; a later authored negation may deliberately reopen them
- **Gap-File Degradation**: fallback gap-file Programs skip type-aware rules and semantic diagnostics instead of paying unreliable semantic costs
- **Buffered Diagnostic Collection**: CLI mode funnels diagnostics through a buffered channel before formatting, which reduces contention between lint workers and output handling
- **On-Demand AST Encoding**: API/WASM responses only include encoded source files when `IncludeEncodedSourceFiles` is requested

### Caching Strategy

- **LSP Session Reuse**: `internal/lsp` builds a shared ts-go `project.Session`, so configured projects, inferred projects, and overlay document state are reused across requests
- **Parse Cache in LSP**: the LSP server passes a shared `project.ParseCache` into the session to avoid re-parsing from scratch on every change
- **Debounced Re-linting**: `refreshCh` and `debounceCh` collapse bursts of file changes and session refreshes onto the main dispatch loop
- **CLI/API Are Mostly Fresh Runs**: CLI and one-shot API requests generally rebuild `Program` state per run; there is no repository-local rule-result cache or persistent incremental lint cache in the main CLI path today. JavaScript API path canonicalization is also scoped to one `lintFiles()` call.
- **Run/Request-Scoped Source Snapshots**: CLI runs and individual API requests share immutable source text/hash snapshots across their Programs. Keys are the exact compiler-host source names, never real paths, so lexical, overlay, and symlink aliases remain distinct. Failed reads are not cached. The source layer in one cache binds to one filesystem view across its generations; compiler hosts using another view bypass this layer while retaining content-keyed AST reuse.
- **Isolated Fix Overlays and Final Invalidation**: each CLI lexical target performs its bounded fix/re-lint cascade in a private overlay, so aliases never observe each other's intermediate bytes or configs. After every target settles and completed outputs are written, the run-scoped parse cache installs one fresh source-snapshot generation. The API fix path only returns one-pass output and does not mutate disk, while LSP remains version/didChange-driven.
- **Run-Scoped Parse Reuse**: CLI Program rebuilds within one invocation and Programs within one API request share the existing content-keyed AST parse cache. Source-generation invalidation does not clear AST entries, so unchanged bytes can reuse their `SourceFile`. The cache is discarded with its run/request and is never repository-persistent or shared across lint requests.
- **Bounded Multi-Pass Fixing**: `--fix` and LSP `fixAll` intentionally rerun lint after applying edits, but cap the cascade at `maxFixPasses = 10`

### Memory Management

- **ts-go Owns the Heavy Graphs**: AST nodes, checker state, `Program` graphs, and session state are primarily owned by ts-go; rslint adds listener maps, diagnostics, and config-derived rule lists on top
- **Short-Lived Per-File Structures**: comment slices, disable managers, and registered listener maps are allocated per file and dropped after traversal; `clear(registeredListeners)` helps release references promptly
- **Source Snapshot Ownership**: snapshot entries hold an immutable source string plus its 128-bit hash without explicitly copying source bytes; on an AST miss, that string is passed directly to the parser. After generation replacement, a retained unchanged AST may still hold the prior equal string while the fresh snapshot owns the new read. Replaced generations are reclaimed after any in-flight lookup releases them. AST retention and source-generation retention remain deliberately separate lifecycles.
- **Fix Application Uses Linear Rebuilds**: `ApplyRuleFixes` sorts fixes, skips overlapping edits, and rebuilds the output with `strings.Builder` rather than mutating source buffers in place
- **Bounded Queues**: CLI diagnostics use a buffered channel of 4096 items; LSP request/outgoing queues are buffered to 100, and debounce/refresh signals are single-slot channels
- **No Repo-Local Pooling Layer Today**: there is no explicit `sync.Pool`-based object pooling strategy in the main lint path at the moment
- **Config Module Lifetime**: LSP fresh transactions use a unique entry-module URL so rewritten bytes and side effects are re-evaluated; Node retains those ESM namespaces for the process lifetime. Native API requests use cached entry imports, while their community-plugin workers are request-scoped and re-import config modules in a separate realm. Atomic cross-request reuse of the Go catalog, entry value, and plugin worker is intentionally left as follow-up; callers must not rely on config rewrites becoming visible (or staying cached) uniformly within one long-lived API instance.
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
- **JavaScript API**: `packages/rslint` talks to `cmd/rslint --api` through the versioned `3.0.0` protocol; the handshake negotiates reverse `pluginLint` support before third-party rules run
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
