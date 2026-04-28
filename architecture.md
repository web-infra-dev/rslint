# Architecture Overview

Rslint is a high-performance JavaScript and TypeScript linter written in Go, designed as a drop-in replacement for ESLint and TypeScript-ESLint. It leverages [typescript-go](https://github.com/microsoft/typescript-go) to achieve 20-40x speedup over traditional ESLint setups through native parsing, direct TypeScript AST usage, and parallel processing.

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

- **100% ESLint Plugin Compatibility**: Focus on core rules, not every community plugin
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

| Path                           | Purpose                                                                               | Key Relationships                                                                                                                                                                                                                                |
| ------------------------------ | ------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `website/`                     | Documentation site and Playground UI                                                  | Uses `packages/rslint-wasm` to run browser linting and `packages/rslint-api` to decode encoded source files; Playground lint requests ultimately reach `internal/linter`, and inspect requests reach `internal/inspector` through `internal/api` |
| `cmd/rslint/`                  | Main Go binary entry point with CLI, API, and LSP modes                               | Main CLI path is `internal/config -> cmd/rslint/programs.go -> internal/linter`; `--api` is consumed by `packages/rslint` and `packages/rslint-wasm`; `--lsp` is consumed by `packages/vscode-extension`                                         |
| `cmd/tsgo/`                    | ts-go semantic inspection/export tool                                                 | Talks directly to `typescript-go` and bypasses the lint framework; consumed by `packages/tsgo` and `crates/tsgo-client`                                                                                                                          |
| `internal/api/`                | stdio IPC protocol and service types for JS/WASM integration                          | Shared protocol layer for `cmd/rslint --api`; used by `packages/rslint`, `packages/rslint-wasm`, `internal/linter`, and `internal/inspector`                                                                                                     |
| `internal/config/`             | Configuration loading, parsing, merging, discovery, and centralized rule registration | `RegisterAllRules()` in `config.go` populates the shared registry, while `rule_registry.go` implements registry/query logic used by `cmd/rslint/programs.go` and `internal/linter`                                                               |
| `internal/inspector/`          | AST/type/symbol/signature/flow inspection for Playground                              | Auxiliary backend used mainly by website Playground inspect panels; builds rich semantic data from `typescript-go` programs                                                                                                                      |
| `internal/linter/`             | Core lint engine, traversal, and fix application                                      | Consumes rules from `internal/rule`, file config from `internal/config`, and `Program` / `TypeChecker` data from `typescript-go`; also serves `internal/api` and `internal/lsp`                                                                  |
| `internal/lsp/`                | Language Server Protocol implementation                                               | Wraps `typescript-go project.Session`, receives config updates from `packages/vscode-extension`, and invokes `internal/linter` on session-backed programs                                                                                        |
| `internal/rule/`               | Rule framework, context, diagnostics, fixes, and disable manager                      | Shared foundation for core rules and plugin rules; called by `internal/linter` through listeners and reporting APIs                                                                                                                              |
| `internal/rule_tester/`        | Go-side rule testing helpers                                                          | Supports rule development and complements JS-side testers in `packages/rule-tester` and `packages/rslint-test-tools`                                                                                                                             |
| `internal/rules/`              | Core lint rule implementations without plugin namespace                               | Registered through `internal/config/config.go` via `RegisterAllRules()` and then executed by `internal/linter` like plugin rules                                                                                                                 |
| `internal/plugins/typescript/` | `@typescript-eslint`-style rules                                                      | Registered into the shared rule registry by `RegisterAllRules()` and often rely on `TypeChecker` from `typescript-go`                                                                                                                            |
| `internal/plugins/react/`      | React rule implementations                                                            | Registered into the shared rule registry by `RegisterAllRules()` and executed through the same listener pipeline in `internal/linter`                                                                                                            |
| `internal/plugins/jest/`       | Jest rule implementations                                                             | Registered into the shared rule registry by `RegisterAllRules()` and executed through the same listener pipeline in `internal/linter`                                                                                                            |
| `internal/plugins/import/`     | Import plugin registration and rules                                                  | Contributes plugin rules through `RegisterAllRules()` and participates in normal config-driven linting                                                                                                                                           |
| `internal/utils/`              | Shared utilities for JSONC, compiler hosts, overlay VFS, and helpers                  | Supports `cmd/rslint/programs.go`, config loading, and various linter entry points                                                                                                                                                               |
| `packages/rslint/`             | Main npm package with JS API and CLI wrapper                                          | Spawns `cmd/rslint --api` in Node environments and uses `internal/api` message shapes                                                                                                                                                            |
| `packages/rslint-api/`         | Frontend-facing encoded source file / AST decoding helpers                            | Used mainly by website Playground to decode AST/source data returned from the Go API                                                                                                                                                             |
| `packages/rslint-test-tools/`  | Testing utilities and cross-ecosystem rule tests                                      | Supports package-side and integration-style tests around the linter and rule ecosystem                                                                                                                                                           |
| `packages/rslint-wasm/`        | Browser/WASM package for running `rslint --api` in a worker                           | Starts the browser worker, hosts the wasm runtime, and bridges website Playground requests to `internal/api`, `internal/linter`, and `internal/inspector`                                                                                        |
| `packages/rule-tester/`        | Forked `@typescript-eslint/rule-tester` package used in tests                         | JS-side rule testing support that complements Go-side helpers                                                                                                                                                                                    |
| `packages/utils/`              | Shared JavaScript utilities                                                           | Shared support package for the JS/website tooling layer                                                                                                                                                                                          |
| `packages/vscode-extension/`   | VS Code extension for IDE integration                                                 | Launches `cmd/rslint --lsp`, sends JS/TS config payloads through `rslint/configUpdate`, and consumes diagnostics/code actions from `internal/lsp`                                                                                                |
| `packages/tsgo/`               | JS wrapper package for the `tsgo` tool                                                | JavaScript-facing wrapper around `cmd/tsgo` output                                                                                                                                                                                               |
| `typescript-go/`               | Git submodule containing TypeScript compiler Go port                                  | Provides parser, AST, checker, `Program`, `project.Session`, diagnostics, scanner, and VFS primitives used throughout the backend                                                                                                                |
| `shim/`                        | Generated bridge packages exposing ts-go internals                                    | Bridge layer between repository Go code and `typescript-go` internals; generated and updated by `tools/`                                                                                                                                         |
| `tools/`                       | Shim generator and ts-go update scripts                                               | Generates `shim/` code and maintains the pinned `typescript-go` integration                                                                                                                                                                      |
| `crates/tsgo-client/`          | Rust client for communicating with `cmd/tsgo`                                         | Spawns `cmd/tsgo` and consumes its semantic/project output from Rust                                                                                                                                                                             |

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

1. **Source Text Loading**: Files come from the real filesystem, an overlay VFS, or LSP document overlays.
2. **Program Construction**: CLI/API build `Program` objects directly; LSP uses ts-go `project.Session`.
3. **Lexical + Syntax Parsing**: ts-go tokenizes and parses source files into TypeScript-native AST nodes.
4. **Semantic Analysis**: When needed, the linter acquires a `TypeChecker` from the `Program`.
5. **Rule Registration**: Enabled rules register listeners keyed by AST kind.
6. **AST Traversal**: The linter traverses each file once using a DFS walk.
7. **Rule Execution**: Listener callbacks inspect nodes and may use syntax only or syntax plus type information.
8. **Diagnostic Collection**: Findings are reported as `RuleDiagnostic` values, with optional fixes or suggestions.
9. **Output Generation**: CLI prints results, API returns structured data, and LSP converts them to LSP diagnostics/code actions.

### Error Recovery Strategy

The parser and program builder are tolerant enough to support editor and fallback scenarios:

- ts-go can continue producing ASTs after syntax errors
- LSP delays lint on rapid edits to avoid repeated work on broken intermediate text
- fallback Programs for gap files are created leniently so parse failures there do not fail the whole run
- CLI reports syntax diagnostics separately when program creation fails in the strict tsconfig-backed path

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
    Run              func(ctx RuleContext, options any) RuleListeners
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
- **API**: returns fix metadata to the caller and also exposes `applyFixes`

## 8. Configuration & Directives

### Configuration Formats

Rslint supports two configuration formats following ESLint flat config semantics (array of config entries):

#### JS/TS Configuration (Recommended)

JS/TS config files (`rslint.config.ts`, `rslint.config.js`, `rslint.config.mjs`, `rslint.config.mts`) are the recommended approach. They support preset composition via `defineConfig()`:

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

| Field             | Type                | Description                                           |
| ----------------- | ------------------- | ----------------------------------------------------- |
| `files`           | `string[]`          | Glob patterns this entry applies to                   |
| `ignores`         | `string[]`          | Glob patterns excluded by this entry                  |
| `languageOptions` | `object`            | Parser options such as `project` and `projectService` |
| `rules`           | `Record<string, …>` | Rule level or `[level, options]`                      |
| `plugins`         | `string[]`          | Plugin declarations used for plugin gating            |
| `settings`        | `Record<string, …>` | Shared settings available in `RuleContext`            |

### Configuration Loading

The loading flow differs by config type:

**JS/TS config**:

1. `packages/rslint/src/cli.ts` discovers one or more config files
2. each config is loaded and normalized on the Node side
3. the Node wrapper sends a stdin payload to the Go binary via `--config-stdin`
4. Go parses either a multi-config payload or a legacy single-config payload
5. nearest-config lookup is used later to decide file ownership and rule selection

**JSON config**:

1. Go searches for `rslint.json` / `rslint.jsonc`
2. JSONC parsing is applied
3. `normalizeJSONConfig()` injects core and plugin rules as defaults

### Configuration Merging

Config merging follows flat-config-style semantics in `GetConfigForFile()`:

1. global-ignore-only entries apply first
2. `files` patterns decide which entries match
3. `ignores` can remove files from an otherwise matching entry
4. later rule values override earlier rule values
5. later settings override earlier settings by key
6. `languageOptions.parserOptions` are merged field-by-field

In multi-config mode, `FindNearestConfig()` is used before `GetConfigForFile()` so that each file is merged against the nearest config directory.

Additional current behaviors:

- `.gitignore` is injected into CLI configs as a global-ignore entry
- plugin rules are gated by declared plugins for JS/TS configs
- files passed explicitly on the CLI can still be linted even if they do not match a config `files` pattern, if merged config still assigns them rules
- files outside all tsconfig-backed Programs can become gap files and receive an AST-only fallback Program

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

The CLI has a two-layer architecture: a Node.js wrapper (`packages/rslint/src/cli.ts`) and the Go binary (`cmd/rslint/`).

1. **Node.js Wrapper**: parses args, discovers JS/TS configs, and decides whether to use `--config-stdin`
2. **Config Path Selection**:
   - JS/TS configs are normalized in Node and piped to Go
   - JSON configs are loaded directly by Go
3. **Mode Selection**:
   - `--lsp`: starts the LSP server
   - `--api`: starts the IPC API server
   - default: runs direct CLI linting
4. **Program Creation**: Go builds one or more tsconfig-backed Programs, plus optional no-tsconfig or gap-file fallback Programs
5. **Ownership Filtering**: multi-config mode computes nearest-config ownership so files are linted once
6. **Rule Resolution**: `getRulesForFile` resolves enabled rules per file
7. **Rule Execution**: `RunLinter()` schedules per-Program work and `RunLinterInProgram()` does the actual file traversal
8. **Result Aggregation**: diagnostics are sent through one run-scoped diagnostics channel and collected at the CLI layer
9. **Fix Passes**: when enabled, fixes are applied and Programs can be rebuilt for another pass
10. **Output Formatting**: default, JSON line, and GitHub workflow formats are supported
11. **Exit Code**: depends on diagnostics, warnings, and fix outcomes

### Concurrency Model

The Go side has four parallelism points; each one honors `--singleThreaded`,
which is the user-facing escape hatch for serial / reproducible execution.

1. **Linter work group** (`RunLinter()` via `core.NewWorkGroup`)
   - Schedules per-Program lint work; runs rules in parallel within a Program.
   - `--singleThreaded` collapses the work group to serial execution.

2. **gitignore reading ‖ Program creation** (in `cmd/rslint/cmd.go`)
   - `ReadGitignoreAsGlobs` walks `.gitignore` for each config; it is independent
     of `createProgramsForConfig` (which only reads
     `languageOptions.parserOptions.project`, never `Ignores`). The two are
     dispatched as parallel goroutines per config.
   - In multi-config mode, gitignore reads also fan out across configs in
     parallel; `createProgramsForConfig` invocations run serially across configs
     (typescript-go's API is invoked one config at a time).
   - `--singleThreaded` runs both stages sequentially — no goroutines spawned.

3. **Gap-file directory walker** (`internal/config/file_discovery.go`)
   - `DiscoverGapFiles` uses a fixed-size worker pool (`walkPool`) that walks
     the directory tree. Live goroutine count is capped at `workers`, not the
     number of directories.
   - Default `workers = max(2, GOMAXPROCS)`; `--singleThreaded` forces
     `workers = 1`, which degenerates into a fully serial DFS-style traversal.
   - The walker is built on a `vfsAdapter` with `followSymlinks = false`:
     symlinked subdirectories are skipped. This matches ESLint v10's
     flat-config file walker, which uses `@humanfs/node` and recurses only
     when `Dirent.isDirectory()` is true (Node's `readdir({withFileTypes:
true})` reports the dirent type without following symlinks, so
     `Dirent.isDirectory()` is false for symlinks). The skip also eliminates
     scheduling-dependent non-determinism that a parallel walker would
     otherwise introduce.

4. **Multi-config gap discovery** (`DiscoverGapFilesMultiConfig`)
   - Iterates `configMap` serially, calling `DiscoverGapFiles` once per config.
   - Each call is itself bounded by its own worker pool, so total live
     goroutines remain bounded by `workers`, not `len(configMap) × workers`.

Other invariants:

- File-ownership filtering avoids duplicate work in multi-config mode.
- LSP uses a different orchestration model and keeps session access on its
  main dispatch loop.

#### `--singleThreaded` semantics

`--singleThreaded` is honored in every parallelism point above:

| Point                          | Effect when set                                               |
| ------------------------------ | ------------------------------------------------------------- |
| Linter work group              | Collapsed to serial via `core.NewWorkGroup(true)`.            |
| gitignore ‖ Program creation   | Both stages run sequentially in the main goroutine.           |
| Multi-config gitignore fan-out | Replaced by a sequential for-loop.                            |
| Gap-file walker workers        | Forced to 1 (single goroutine, no concurrency).               |
| Multi-config gap discovery     | Already serial across configs; inner walker also forced to 1. |

End result: with `--singleThreaded`, the Go side spawns no goroutines beyond
those typescript-go itself creates for syntactic / semantic work.

## 10. Performance & Memory Considerations

### Design Principles

- **Direct ts-go Data Model**: rslint operates on ts-go `Program`, AST, and `TypeChecker` objects directly instead of converting through a second AST representation
- **Program-Level Parallelism**: `RunLinter` queues work per `Program` through `core.NewWorkGroup`; `--singleThreaded` forces the same flow to run serially
- **Single-Walk Rule Dispatch**: each file is traversed once, with rules registering listeners up front and sharing the same AST walk
- **Early Filtering**: skip paths, allowFiles/allowDirs, nearest-config ownership filters, and gap-file type filtering reduce work before listeners run

### Performance Optimizations

- **Native Go Implementation**: Eliminates JavaScript runtime overhead
- **Direct TypeScript AST**: No AST conversion between parsers
- **Shared Type Checker**: `RunLinterInProgram` acquires one checker for the lint phase and reuses it across files and rules in the same `Program`
- **Checker Phase Separation**: the checker is released before TypeScript semantic diagnostics run, so `GetSemanticDiagnostics` can reacquire its own checker cleanly
- **File Filtering**: Skip node_modules and bundled files automatically
- **Gap-File Degradation**: fallback gap-file Programs skip type-aware rules and semantic diagnostics instead of paying unreliable semantic costs
- **Buffered Diagnostic Collection**: CLI mode funnels diagnostics through a buffered channel before formatting, which reduces contention between lint workers and output handling
- **On-Demand AST Encoding**: API/WASM responses only include encoded source files when `IncludeEncodedSourceFiles` is requested

### Caching Strategy

- **LSP Session Reuse**: `internal/lsp` builds a shared ts-go `project.Session`, so configured projects, inferred projects, and overlay document state are reused across requests
- **Parse Cache in LSP**: the LSP server passes a shared `project.ParseCache` into the session to avoid re-parsing from scratch on every change
- **Debounced Re-linting**: `refreshCh` and `debounceCh` collapse bursts of file changes and session refreshes onto the main dispatch loop
- **CLI/API Are Mostly Fresh Runs**: CLI and one-shot API requests generally rebuild `Program` state per run; there is no repository-local rule-result cache or persistent incremental lint cache in the main CLI path today
- **Bounded Multi-Pass Fixing**: `--fix` and LSP `fixAll` intentionally rerun lint after applying edits, but cap the cascade at `maxFixPasses = 10`

### Memory Management

- **ts-go Owns the Heavy Graphs**: AST nodes, checker state, `Program` graphs, and session state are primarily owned by ts-go; rslint adds listener maps, diagnostics, and config-derived rule lists on top
- **Short-Lived Per-File Structures**: comment slices, disable managers, and registered listener maps are allocated per file and dropped after traversal; `clear(registeredListeners)` helps release references promptly
- **Fix Application Uses Linear Rebuilds**: `ApplyRuleFixes` sorts fixes, skips overlapping edits, and rebuilds the output with `strings.Builder` rather than mutating source buffers in place
- **Bounded Queues**: CLI diagnostics use a buffered channel of 4096 items; LSP request/outgoing queues are buffered to 100, and debounce/refresh signals are single-slot channels
- **No Repo-Local Pooling Layer Today**: there is no explicit `sync.Pool`-based object pooling strategy in the main lint path at the moment
- **Garbage Collection Handles Cycles**: the repository does not implement custom cycle breaking for AST/checker graphs; lifecycle cleanup relies on Go GC and on dropping references after each run

## 11. Extensibility & Future Directions

### Plugin Architecture

Today, plugin support is built-in rather than dynamically loading arbitrary third-party ESLint plugins at runtime.

The repository currently ships internal support for:

- TypeScript plugin rules
- React rules
- Jest rules
- Import plugin rules

### Rule Extension Points

- **Core Rules**: add packages under `internal/rules/`
- **Plugin Rules**: add packages under `internal/plugins/<plugin>/rules/`
- **Rule Options**: each rule receives parsed options through `Run(ctx, options)`
- **Custom Listener Shapes**: rules can listen on standard kinds and synthetic pattern/exit kinds

### Integration Points

- **Language Server**: `internal/lsp` exposes diagnostics and code actions
- **JavaScript API**: `packages/rslint` talks to `cmd/rslint --api`
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

The maintained rule-porting workflow now lives under [`agents/port-rule`](./agents/port-rule/).

Use these entry points instead of duplicating a separate checklist here:

- [`agents/port-rule/SKILL.md`](./agents/port-rule/SKILL.md): primary skill entry and workflow
- [`agents/port-rule/references/PORT_RULE.md`](./agents/port-rule/references/PORT_RULE.md): detailed end-to-end porting guide
- [`agents/QUICK_REFERENCE.md`](./agents/QUICK_REFERENCE.md): commands, naming conventions, and condensed checklist

If the rule-porting workflow changes, update the material under `agents/port-rule` rather than reintroducing a second checklist in this document.

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
│  Config Load / Normalize / Merge                                             │
│            │                                                                 │
│            ▼                                                                 │
│  Create tsconfig-backed Programs                                             │
│            │                                                                 │
│            ├───────────────► Discover Gap Files ───────► Fallback Program    │
│            │                                                                 │
│            ▼                                                                 │
│  Resolve Enabled Rules Per File                                              │
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
│            ├───────────────► CLI formatter / exit code                       │
│            └───────────────► API structured response                         │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────────┐
│                                LSP PATH                                      │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  VS Code Extension                                                           │
│     │                                                                        │
│     ├───────────────► rslint/configUpdate (JS/TS config payloads)            │
│     ▼                                                                        │
│  cmd/rslint --lsp                                                            │
│     │                                                                        │
│     ▼                                                                        │
│  internal/lsp + ts-go project.Session                                        │
│     │                                                                        │
│     ▼                                                                        │
│  RunLinterInProgram on session Program                                       │
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
- **Fallback Program**: Extra `Program` created for no-tsconfig projects or for gap files; gap-file fallbacks are intentionally non-type-aware
- **Flat Config**: ESLint-style array-based configuration model used by rslint to merge rule settings per file
- **Gap File**: A file matched by config but not present in any tsconfig-backed Program
- **Inspector**: Auxiliary backend path that returns node, type, symbol, signature, and flow information for Playground inspection
- **IPC API**: Length-prefixed JSON message protocol exposed by `cmd/rslint --api` for Node and WASM clients
- **Listener**: Callback registered by a rule for an AST kind or synthetic listener kind
- **Nearest Config**: In multi-config mode, the deepest config directory that owns a file
- **Node Kind**: Enumerated AST kind value used by ts-go and by the listener dispatcher to identify node types
- **Program**: ts-go compilation context, usually tied to a tsconfig or fallback root-file set
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
