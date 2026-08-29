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
│  │  │  internal/program/loader             │  │  internal/api/server     │  │  │
│  │  │  internal/program                    │  │  internal/lsp            │  │  │
│  │  │  internal/linter (pipeline + engine) │  │  internal/inspector      │  │  │
│  │  │  internal/rule / utils               │  │                          │  │  │
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

| Path                           | Purpose                                                                                                                | Key Relationships                                                                                                                                                                                                                                                                                                                                                                                                          |
| ------------------------------ | ---------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `website/`                     | Documentation site and Playground UI                                                                                   | Uses `packages/rslint-wasm` to run browser linting and `packages/rslint-api` to decode encoded source files; Playground requests pass through `internal/api` and `internal/api/server` before reaching `internal/linter` or `internal/inspector`                                                                                                                                                                           |
| `cmd/rslint/`                  | Main Go binary entry point with CLI, API, and LSP modes                                                                | Owns mode selection and process stdio/exit-code composition. The default CLI prepares config, targets, Program-generation and final disk-projection adapters for `internal/linter.RunPipeline`; `--api` delegates concrete requests to `internal/api/server`; `--lsp` delegates to `internal/lsp`                                                                                                                          |
| `internal/output/`             | Report model, summary, colors, and stdout formatters                                                                   | Consumes final sorted `internal/rule` diagnostics and renders `default`, `jsonline`, `github`, or `gitlab`; the CLI is its current consumer, while the package remains available to other repository integrations that need the same output behavior                                                                                                                                                                       |
| `cmd/tsgo/`                    | ts-go semantic inspection/export tool                                                                                  | Talks directly to `typescript-go` and bypasses the lint framework; consumed by `packages/tsgo` and `crates/tsgo-client`                                                                                                                                                                                                                                                                                                    |
| `internal/api/`                | stdio IPC protocol, wire types, and generic bidirectional service for JS/WASM integration                              | Defines the stable request/response boundary used by `packages/rslint`, `packages/rslint-wasm`, and `internal/api/server`; it does not import concrete lint, config, Program-loader, or command implementations                                                                                                                                                                                                            |
| `internal/api/server/`         | Concrete rslint API request handlers                                                                                   | Implements lint and AST-inspection requests over `internal/api`, owns API-specific paths, request-overlay generation adapters, reverse config/plugin adapters, and structured response projection; it delegates lint planning, execution, and in-memory fix application to `internal/linter`                                                                                                                               |
| `internal/config/`             | Configuration models, JSON loading, authored path-space snapshots, matching/merging, and configured-rule evaluation    | Owns the shared extension/default-exclude policy, the `GlobalIgnoreMatcher` consumed by config-candidate discovery, and the `TargetMatcher` consumed by lint-target planning; both matchers share the private config-target resolver. File resolvers receive an immutable `internal/rule.Catalog` explicitly and convert merged rule settings into execution descriptors without owning target walking or config ownership |
| `internal/config/lint/`        | Effective config and rule resolution for already-selected lint targets                                                 | Joins the immutable target/source binding returned by `internal/program/loader` to the owning config's `FileConfigResolver`. CLI and API use the same resolver; it cannot discover targets, infer a different owner, construct Programs, or execute rules                                                                                                                                                                  |
| `internal/config/target/`      | Lint-target planning, explicit-file outcomes, directory walking, canonical deduplication, and config-owner routing     | Imports the parent config model and matching policy through `PathIdentity`, `PathSpaceSnapshot`, and `TargetMatcher`. It also owns exact-first source-path lookup in immutable source-to-target maps. CLI, API, LSP, `internal/config/lint`, and `internal/program/loader` consume `File`, `Plan`, `OwnerIndex`, or that lookup; the parent config package never imports this child package                                |
| `internal/config/discovery/`   | Go-owned JS/TS config candidate discovery and immutable catalog construction                                           | Imports the parent config model/matching policy and `target.OwnerScope` as a narrow provenance handoff, batches exact candidates to a host-supplied Node loader, and returns configs/scopes/failures/effective IDs. CLI, API, and LSP call `DiscoverAutomatic` or `LoadExplicitConfig`; neither the parent config package nor target planning imports discovery                                                            |
| `internal/config/gitignore/`   | Config-scoped `.gitignore` parsing, directory reachability, and pattern projection                                     | Staged JS/TS catalogs carry a filesystem-independent cursor through their existing walk, pruning Git-inaccessible subtrees and freezing observed patterns for lint-target admission; JSON/JSONC and low-level fallback paths reuse the direct collector                                                                                                                                                                    |
| `internal/inspector/`          | AST/type/symbol/signature/flow inspection for Playground                                                               | Auxiliary backend used mainly by website Playground inspect panels; builds rich semantic data from `typescript-go` programs                                                                                                                                                                                                                                                                                                |
| `internal/program/`            | Immutable rslint Program facade and source-generation indexes                                                          | Privately adapts project/compiler and root-parser hosts into one source, filesystem, syntax, optional checker, module-resolution, and module-reference contract. `Program.ModuleGraph()` and generation-scoped caches derive from that same authority; adapter identity is not observable by linter or rules                                                                                                               |
| `internal/program/loader/`     | Run-scoped Program construction and lint-target binding for CLI/API                                                    | Consumes target-owned lint plans, limits plain-lint project construction to configs that own selected targets, binds targets by governing-config order and physical identity, and returns one backend-agnostic Program sequence plus its execution projection. Construction choices and ts-go compatibility details remain private; LSP wraps session-owned Programs directly and does not use this request-scoped loader  |
| `internal/linter/`             | Unified lint pipeline, immutable lint plans, core engine, plugin dispatch, traversal, and bounded autofix lifecycle    | Owns `RunPipeline`: generation acquisition/release, exact `LintPlan` preparation, native/plugin scheduling, diagnostic aggregation, pure fix planning, all in-memory fix rounds, re-observation, and optional one-shot terminal change commit. It imports no config discovery, Program loader, persistence medium, output, API, or LSP protocol                                                                            |
| `internal/lsp/`                | Language Server Protocol implementation                                                                                | Wraps `typescript-go project.Session`, owns transactional config discovery and last-good config state, and supplies document/speculative-generation plus progressive-presentation adapters to `internal/linter`; fix-all maps the core's final in-memory delta to one `TextEdit` and never mutates editor/session state while computing it                                                                                 |
| `internal/rule/`               | Rule framework, immutable rule catalogs, configured-rule descriptors, context, diagnostics, fixes, and disable manager | Defines catalog snapshots and exact object-form ESLint-plugin catalog derivation shared by entry points; `internal/linter` consumes its immutable rule environments, listeners, reporting APIs, and Program-derived cache helper                                                                                                                                                                                           |
| `internal/rule_tester/`        | Go-side rule testing helpers                                                                                           | Supports rule development and complements JS-side testers in `packages/rule-tester` and `packages/rslint-test-tools`                                                                                                                                                                                                                                                                                                       |
| `internal/rules/`              | Core lint rule implementations and final aggregation of every Go rule                                                  | `all.go` consumes the existing core and plugin aggregators once and exposes their shared immutable catalog through `rules.All()`; entry points derive a catalog for the exact request-scoped object-form ESLint-plugin set before config resolution                                                                                                                                                                        |
| `internal/plugins/typescript/` | `@typescript-eslint`-style rules                                                                                       | Its `all.go` contributes rules to `internal/rules`, often relying on `TypeChecker` from `typescript-go`                                                                                                                                                                                                                                                                                                                    |
| `internal/plugins/react/`      | React rule implementations                                                                                             | Its `all.go` contributes rules to `internal/rules` and they execute through the same listener pipeline in `internal/linter`                                                                                                                                                                                                                                                                                                |
| `internal/plugins/jest/`       | Jest rule implementations                                                                                              | Its `all.go` contributes rules to `internal/rules` and they execute through the same listener pipeline in `internal/linter`                                                                                                                                                                                                                                                                                                |
| `internal/plugins/import/`     | Import plugin rule implementations                                                                                     | Its `all.go` contributes rules to `internal/rules` and they participate in normal config-driven linting                                                                                                                                                                                                                                                                                                                    |
| `internal/utils/`              | Shared utilities for JSONC, compiler hosts, invocation-scoped compiler construction, overlay VFS, and AST/type helpers | Provides the low-level compiler-construction and snapshot services privately composed by `internal/program/loader`; command entry points do not coordinate compiler hosts or source generations directly. LSP, rule tests, and auxiliary entry points reuse lower-level compiler helpers. Source-generation module resolution belongs to `internal/program`                                                                |
| `packages/rslint/`             | Main npm package with JavaScript API and CLI wrapper                                                                   | Spawns `cmd/rslint --api` in JavaScript runtime environments and uses `internal/api` message shapes                                                                                                                                                                                                                                                                                                                        |
| `packages/rslint-api/`         | Frontend-facing encoded source file / AST decoding helpers                                                             | Used mainly by website Playground to decode AST/source data returned from the Go API                                                                                                                                                                                                                                                                                                                                       |
| `packages/rslint-test-tools/`  | Testing utilities and cross-ecosystem rule tests                                                                       | Supports package-side and integration-style tests around the linter and rule ecosystem                                                                                                                                                                                                                                                                                                                                     |
| `packages/rslint-wasm/`        | Browser/WASM package for running `rslint --api` in a worker                                                            | Starts the browser worker, hosts the wasm runtime, and bridges website Playground requests through `internal/api` and `internal/api/server` to `internal/linter` and `internal/inspector`                                                                                                                                                                                                                                  |
| `packages/rule-tester/`        | Forked `@typescript-eslint/rule-tester` package used in tests                                                          | JS-side rule testing support that complements Go-side helpers                                                                                                                                                                                                                                                                                                                                                              |
| `packages/utils/`              | Shared JavaScript utilities                                                                                            | Shared support package for the JS/website tooling layer                                                                                                                                                                                                                                                                                                                                                                    |
| `packages/vscode-extension/`   | VS Code extension for IDE integration                                                                                  | Resolves the nearest project-local `@rslint/core` per open document, launches that installation's `cmd/rslint --lsp`, serves reverse config/plugin requests, and routes diagnostics/code actions to the document's selected runtime                                                                                                                                                                                        |
| `packages/tsgo/`               | `@rslint/tsgo-server` JS wrapper package for the `tsgo` tool                                                           | JavaScript-facing wrapper around `cmd/tsgo` output; resolves the matching `@rslint/tsgo-server-<platform>-<arch>` binary package                                                                                                                                                                                                                                                                                           |
| `typescript-go/`               | Git submodule containing TypeScript compiler Go port                                                                   | Provides parser, AST, checker, `Program`, `project.Session`, diagnostics, scanner, and VFS primitives used throughout the backend                                                                                                                                                                                                                                                                                          |
| `shim/`                        | Generated bridge packages exposing ts-go internals                                                                     | Bridge layer between repository Go code and `typescript-go` internals; generated and updated by `tools/`                                                                                                                                                                                                                                                                                                                   |
| `tools/`                       | Shim generator and ts-go update scripts                                                                                | Generates `shim/` code and maintains the pinned `typescript-go` integration                                                                                                                                                                                                                                                                                                                                                |
| `crates/tsgo-client/`          | Rust client for communicating with `cmd/tsgo`                                                                          | Spawns `cmd/tsgo` and consumes its semantic/project output from Rust                                                                                                                                                                                                                                                                                                                                                       |

## 4. Parsing Pipeline

The parsing and linting pipeline uses ts-go's native AST data model directly.
At the rslint boundary, every source universe is represented by one immutable
`internal/program.Program`. Its constructors privately adapt the available
source host; linter, rules, caches, and module analysis consume only the facade
and cannot observe which construction path supplied it.

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
│ rslint Program        │
│ - source universe     │
│ - source services     │
│ - optional checker    │
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

1. **Source and Metadata Loading**: Files come from the real filesystem, an overlay VFS, or LSP document overlays. Each CLI run or API request creates a `program/loader.Session` around its initial immutable VFS view. The session's private compiler hosts keep source snapshots keyed by the exact normalized source path, storing the first successful text read and its xxh3 hash for that generation. The same request scope snapshots successful `package.json` and explicitly registered tsconfig reads by the exact requested path; other JSON, source, ignore, and config-discovery reads remain uncached. When multiple Programs are constructed concurrently, a derived context view coalesces their concurrent cold realpath queries for the same exact path while sharing all parent caches. During autofix, `internal/linter` owns the evolving text in memory and passes only the net changed-file snapshot to the integration's generation provider; CLI/API rebuild request-local Programs over that overlay without changing the base medium. LSP follows document events and its own session/versioned cache lifecycle instead.
2. **Program Loading**: `internal/config/target` first resolves a stable lint-target plan from the loaded config policy, while `internal/config` retains authored matching and ordered tsconfig declarations; project ownership remains private to `internal/program/loader`. Focused file/subdirectory lint parses TypeScript configs in declaration order and uses the authoritative `ParsedCommandLine.FileNames` root set to select the first direct project for each target without constructing the full declared project set. Each parsed config is retained by that selection execution and reused if its Program is built. Selected direct projects build through the bounded worker pool. Only targets absent from every declared root enter the compatibility fallback, which lazily constructs projects in declaration order until the first import-containing Program is found; projects whose compiler options cannot admit the target's extension are skipped. Full-CWD/no-argument lint retains the original eager parallel construction path, then derives direct-root ownership in a separate batched exact/canonical binding step. Program-wide type-check modes still build every project declared by the effective loaded config catalog. Targets outside all projects are loaded into source-only rslint Programs with parser, binder, package/module metadata, and direct-import resolution but without recursively loading a project graph. If a root is unsupported by the direct parser path, the loader may use a compatibility ts-go host internally; callers still receive the same Program facade and query capabilities instead of construction kind. A request-local implementation of ts-go's `ExtendedConfigCache` reuses common `extends` parses without modifying ts-go, while root and project-reference configs are explicitly registered for raw-read snapshots. LSP bypasses the request loader but applies the same root-first, extension-filtered import-fallback policy through one selector shared by normal diagnostics and speculative fixes. Session-owned projects expose their authored command-line roots through a generation cache; unloaded custom projects expose watcher-protected parsed metadata, and Program construction reuses that exact parse. Lexical tsconfig declaration paths remain distinct so symlink-relative includes keep TypeScript's declared-path meaning. For a declared custom tsconfig that Session has not loaded, rslint can retain the existing session-external ts-go Program by exact config path and advance source-only edits through ts-go's `Program.UpdateProgram`; fallback Programs that do not contain the target remain transient. Project-shape events discard resident metadata and Programs, and dependency reads are registered through ts-go's watcher mapper before either becomes resident. If watcher coverage cannot be established, rslint transparently keeps fresh request-local construction. Speculative fix Programs remain isolated and request-local.
3. **Lexical + Syntax Parsing**: ts-go tokenizes and parses source files into TypeScript-native AST nodes. Source-only roots additionally run the ts-go binder so syntax-only rules retain symbols and lexical scopes before their rslint Program is published.
4. **Semantic Analysis**: The lint plan reads each Program's per-file checker capability once, freezes that result with the file's configured rules, and acquires a checker only for eligible files. LSP can additionally narrow rule eligibility for a request before planning. Program capability, configured-rule eligibility, actual checker delivery, and program-wide diagnostics remain separate decisions.
5. **Rule Registration**: Enabled rules register listeners keyed by AST kind.
6. **AST Traversal**: The linter traverses each file once using a DFS walk. It prunes syntax that TypeScript-Go synthesized from JSDoc comments; ESLint-compatible parsers expose that text through comment APIs rather than rule AST listeners.
7. **Rule Execution**: Listener callbacks inspect nodes and may use syntax only or syntax plus type information.
8. **Diagnostic Collection**: Findings are reported as `RuleDiagnostic` values, with optional fixes or suggestions.
9. **Output Generation**: `RunPipeline` returns the last successful observation, the final in-memory delta, and final text for every target to which a fix was applied. CLI optionally commits the net delta to disk once and builds a report from final diagnostics, API returns final structured data plus per-fixed-file `output` (including a file restored to its input), and LSP maps fix-all's net delta to a whole-document edit.

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
- Source-only Programs cannot provide a `TypeChecker` or
  program-wide diagnostics, so the plan excludes type-aware rules and Phase 2
  observes an empty result through the same facade without testing their source
  kind

## 5. Abstract Syntax Tree (AST)

### AST Representation

The AST comes directly from ts-go. Rslint does not build a second custom AST layer for linting.

Important characteristics:

- **Node Types**: ts-go `ast.Kind` values
- **Node Objects**: `*ast.Node` and `*ast.SourceFile`
- **Traversal Style**: `ForEachChild(...)` with depth-first recursion
- **JSDoc Boundary**: synthetic syntax reparsed from JSDoc is excluded from
  rule-listener traversal; comment-aware rules use the shared comment store or
  TypeScript-Go's explicit JSDoc APIs. Hosted tags can also populate fields on
  ordinary source nodes, so syntax-sensitive rules use authored-only AST views
  without removing JSDoc semantics from the type checker.
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

`RequiresTypeInfo` is admitted only when the Program can provide a checker for
the file. Integrations with a narrower request-level guarantee, such as LSP's
per-document `HasTypeInfo`, filter the configured rules before planning; rules
never infer eligibility from a private Program construction strategy.

Within `internal/rule`, ownership is split by concern: `rule.go` owns rule metadata and listeners,
`diagnostic.go` owns diagnostic/edit value types, and `context.go` owns the
runtime context and reporting pipeline. `DiagnosticConsumer` and `EditDemand`
are canonical rule-framework types; `internal/linter` consumes them directly
instead of re-exporting aliases.

### Rule Context

`RuleContext` is the runtime environment passed to each rule. It includes:

```go
type RuleContext struct {
    SourceFile     *ast.SourceFile
    Settings       map[string]interface{}
    LanguageOptions LanguageOptions
    Globals        Globals
    Comments       *CommentStore
    Refs           *RefStore
    BOM            *SourceBOM
    TypeChecker    *checker.Checker
    DisableManager *DisableManager

    program *program.Program
}

func (*RuleContext) Program() *program.Program
func (RuleContext) WithProgram(*program.Program) RuleContext // one-time assembly hook
func (*RuleContext) ReportRange(...)
func (*RuleContext) ReportRangeWithFixes(...)
func (*RuleContext) ReportRangeWithSuggestions(...)
func (*RuleContext) ReportNode(...)
func (*RuleContext) ReportNodeWithFixes(...)
func (*RuleContext) ReportNodeWithSuggestions(...)
func (*RuleContext) ReportNodeWithFixesAndSuggestions(...)
func (*RuleContext) ReportRangeWithFixesAndSuggestions(...)
func (*RuleContext) ReportNodeWithDeferredFixes(..., func() []RuleFix)
func (*RuleContext) ReportRangeWithDeferredFixes(..., func() []RuleFix)
func (*RuleContext) ReportNodeWithDeferredSuggestions(..., func() []RuleSuggestion)
func (*RuleContext) ReportRangeWithDeferredSuggestions(..., func() []RuleSuggestion)
func (*RuleContext) ReportNodeWithDeferredFixesAndSuggestions(
    ..., func() []RuleFix, func() []RuleSuggestion,
)
func (*RuleContext) ReportRangeWithDeferredFixesAndSuggestions(
    ..., func() []RuleFix, func() []RuleSuggestion,
)
```

`Program()` is the only source authority in a linter-created context. Rules do
not receive a raw compiler Program, an adapter discriminator, or a second module
runtime. Rebinding an assembled context to a different Program is rejected
because `SourceFile`, references, checker state, and other file-derived values
would belong to the old generation. `TypeChecker` remains the actual per-file
semantic capability delivered by the frozen lint plan; Program construction
alone never admits a `RequiresTypeInfo` rule. Process cwd is stored once in the
file-shared cache and exposed through `ProcessCurrentDirectory()` rather than
copied into every per-rule context.

Configuration is resolved once per file shape into one immutable
`RuleEnvironment` shared by that file's `ConfiguredRule` entries. During
planning that environment is frozen beside the file's rules and checker grant;
execution constructs settings and globals once per file and copies the resulting
base context for each rule. Module references and whole-Program indexes are not
context fields: generic source references come from `Program().ModuleGraph()`,
while rule-specific derived indexes use `CachedByProgram`. Both remain keyed by
the same Program generation and can never become a second source identity.

The linter creates one short-lived `CommentStore` per file. `Comments.All()`
materializes the scanner-backed, source-ordered, deduplicated comment list only
for the first consumer; later consumers share that list. A source without `//`
or `/*` takes a cheap byte-scan fast path. Inline-global parsing first checks
for an exact raw-text directive candidate, so ordinary files do not force
comment collection.
The linter also creates one shared `RefStore` handle per file. Its candidate
identifier walk is deferred until the first reference query, and binder name
resolution is then performed once per queried symbol name. Rules query it with
binder declaration symbols instead of repeating AST walks or TypeChecker
lookups; files whose rules never request references do not materialize the
index. `Resolve` (identifier → symbol) and `References` (symbol → identifiers)
try the binder scope walk first, which answers most queries without ever
touching the TypeChecker; when the binder can't place an identifier — a
symbol declared outside the file (cross-file, `.d.ts`, standard-library
globals) — `Resolve` falls back to the TypeChecker, at the cost of a round
trip for that identifier, and `References` picks up that same fallback
automatically when queried with a symbol `Resolve` obtained that way. Without
a TypeChecker (`NewRefStore`'s third argument is `nil`), that fallback is a
no-op and both methods only ever see symbols declared in the current file.
`ResolveInFile` is the explicit binder-only forward lookup: it never takes the
TypeChecker fallback even when one is available. ESLint scope rules such as
`no-undef` use this path so DOM/lib, ambient `.d.ts`, and cross-file TypeScript
symbols cannot make their result depend on whether the file happened to receive
a checker. `IsDefinedInFile` extends that answer with declaration-less bindings
from the resolved language defaults, while `IsNameDefinedInFile`
supports scope rules whose query location is not itself an identifier
reference. `HasNonGlobalTopLevelScope` exposes the corresponding scope fact
without exposing a language mode or requiring rules to parse paths.
Config resolution normalizes the per-file `ecmaVersion` and authored top-level
`languageOptions.sourceType` (`module`, `script`, or `commonjs`) into
`LanguageOptions`, which is exposed as a whole on each native `RuleContext`.
`ecmaVersion`'s zero value means the moving `latest` edition. The legacy
`parserOptions.sourceType` location is not read. The linter resolves
an omitted source type from the filename (`.cjs` to `commonjs`; `.js`/`.jsx`/
`.mjs` and TypeScript-flavoured `.ts`/`.tsx`/`.cts`/`.mts` to `module`).
The remaining zero value has module semantics through `EffectiveSourceType`.
Source type does not change TypeScript parsing or compiler module resolution. The linter uses the normalized edition to build
one `Globals` value for each native rule context. Rules read
`LanguageOptions` when upstream behavior depends on language configuration;
they use `Globals` for variable-availability decisions. `Globals` owns the
ESLint-versioned language-global set, resolved language defaults, the authored
`languageOptions.globals` source, inline `/* global */` settings and ranges,
and the effective access after applying their precedence. Rules use
`Globals.Access` for standard language-global decisions and its narrower source
accessors only when upstream behavior depends on provenance, instead of
rebuilding the merge. A rule whose upstream semantics add another source, such
as TypeScript library globals, applies this view last so `ecmaVersion` and
authored overrides remain authoritative. Non-global wrapper bindings remain a
`RefStore` initialization concern.

Before constructing rule contexts, the linter calls `ResolveLanguageDefaults`
once and passes its concrete `GlobalsInit`, `RefStoreInit`, and effective
`LanguageOptions` results to their respective consumers. An omitted source
type is filled from the filename (`.cjs` → `commonjs`; `.js`/`.jsx`/`.mjs`
and TypeScript-flavoured extensions `.ts`/`.tsx`/`.cts`/`.mts` → `module`),
matching espree and typescript-eslint. The resolver then selects inits from
that effective source type: `commonjs` contributes writable `exports`,
read-only `global`, `module`, and `require` on every extension, plus — on
espree-parsed extensions (`.js`/`.jsx`/`.mjs`/`.cjs`) — non-global wrapper
scope and the wrapper-local `arguments` binding; `module` contributes a
non-global top-level scope; `script` forces a global program scope even when
module syntax is present; TypeScript-flavoured `commonjs` keeps that same
global program scope.
Authored `sourceType` therefore applies on every extension, including
`.ts`/`.tsx`. The resolver does not inspect `package.json`. A rule reads
`RuleContext.LanguageOptions` when its upstream behavior depends on them.
Every authored alias is normalized to one of ESLint's three access levels —
`utils.GlobalAccess`, whose zero value means no source mentioned the name.
Booleans follow the `globals` package: `true` is writable, `false` is read-only.
The Node plugin scope seeds the same levels into its scope manager for
ESLint-compatible scope APIs.
The linter binds immutable rule name, severity, and diagnostic-sink metadata to
each context once. The reporting methods use that state directly rather than
allocating bound callback closures for every reporting variant.

Native lint entry points provide one `DiagnosticConsumer` containing both the
report callback and an `EditDemand` bit mask. Autofixes and suggestions are
independent demand bits; zero requests diagnostics without optional edits.
This is a reporting-pass capability, not a `Program` property:
`RunLinter` normalizes it once and passes the same immutable consumer separately
to every per-Program lint task and then into each rule reporter.

The deferred reporting methods apply inline-disable suppression first, inspect
the matching demand bit, and only then invoke each category-specific builder
synchronously. Builders are never retained and must contain only work needed
to construct their optional artifact; diagnostic detection, message, range,
and severity are decided before the builder. A diagnostic that can carry both
autofixes and suggestions supplies two independent builders, so a consumer
requesting one category does not materialize the other. Keeping builders
category-specific avoids exposing a general callback protocol or a demand
query that rules could accidentally use to change diagnostic semantics. Adding
another native optional artifact is additive: define one demand bit and one
category-specific builder path. The pass/Program scheduling boundary and
existing rule APIs do not need to change.

Existing eager `Report*WithFixes` and `Report*WithSuggestions` methods remain
available for gradual rule migration. Their already-built artifacts are
filtered by the same consumer demand, but only the deferred methods can avoid
the construction cost. Direct `RuleContext.WithReporter` compatibility callers
request all edits; production lint entry points bind `DiagnosticConsumer`
explicitly. The demand neither changes TypeChecker acquisition nor the
independent serialized eslint-plugin reverse-dispatch protocol.

### Listener Registration

Rules do not walk the AST themselves. Instead:

1. `rules.All()` from `internal/rules` supplies the shared immutable Go rule catalog; an entry point derives a catalog for the exact object-form ESLint-plugin set required by its CLI run, API request, or committed LSP config generation
2. config merge resolves enabled rule names against that explicitly supplied catalog
3. each enabled rule runs `Run(ctx)`
4. `Run(ctx)` returns listeners keyed by `ast.Kind`
5. the linter appends them, in rule order, to the checker-shard task's sparse
   dispatch registry
6. after traversing a file, the task clears every listener slot and reuses the
   registry's map and per-kind backing slices for its next serial file

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
    Range        core.TextRange
    RuleName     string
    Message      RuleMessage
    FixesPtr     *[]RuleFix
    Suggestions  *[]RuleSuggestion
    SourceFile   ast.SourceFileLike
    FilePath     string
    Severity     DiagnosticSeverity
    Origin       DiagnosticOrigin
    PreFormatted bool
}

type RuleFix struct {
    Text  string
    Range core.TextRange
}
```

`internal/linter` owns the shared producer semantics around this model. Native
rules, TypeScript syntax/program diagnostics, and reconstructed third-party
plugin diagnostics all enter the surfaces as `RuleDiagnostic` values. The
single-file TypeScript syntax projection is shared by CLI/API target
collection and LSP document linting, while TypeScript program diagnostics keep
their richer message-chain and related-information formatting. Completed CLI
and API diagnostic sets are stably ordered by caller-visible file path and
start byte offset only after each surface has projected paths into its own
identity space; equal keys retain producer emission order. LSP deliberately
does not use that completed-set ordering because it publishes native results
first and merges generation-stamped plugin results later.

After that semantic boundary, representation remains integration-owned. CLI
builds an `internal/output.Report`, API projects to its 1-based structured wire
model and flat UTF-16 edit offsets, and LSP projects to 0-based LSP positions
while retaining its stale-generation and code-action lifecycle. Counts, path
bases, stderr notices, and protocol empty-array rules stay integration-owned;
lint/fix observation order and fix-round state belong to the core pipeline.

### Severity Levels

- `SeverityError`: lint error
- `SeverityWarning`: lint warning
- `SeverityOff`: rule disabled

### Autofix System

Autofix is implemented as text edits:

- insert = replace an empty range with text
- replace = replace a non-empty range with text
- remove = replace a range with the empty string

Rules with cheap, unconditional edits can attach fixes through
`ReportRangeWithFixes(...)` or `ReportNodeWithFixes(...)`. Rules whose edit-only
analysis is expensive should use `ReportRangeWithDeferredFixes(...)` or
`ReportNodeWithDeferredFixes(...)`, allowing the framework to skip that
analysis when the current native consumer does not need autofixes. The
suggestion counterparts provide the same migration path for expensive
suggestion construction. A single diagnostic with independently expensive
autofixes and suggestions should use
`Report*WithDeferredFixesAndSuggestions(...)`; its two builders are gated
separately while suppression and diagnostic emission still happen once.

Fix application happens in `internal/linter/source_code_fixer.go`:

1. sort fixes within each diagnostic
2. sort fixable diagnostics by position
3. skip overlapping or conflicting adjacent edits
4. rebuild the source text

`internal/linter.RunPipeline` owns the complete product-level fix lifecycle.
Its request constructors seal valid combinations of observation policy,
autofix policy, and semantic ports; CLI, API, and LSP do not call planning,
native execution, plugin dispatch, or the source-code fixer as independent
stages. The lifecycle is:

1. Ask a `GenerationProvider` to materialize one immutable Program generation
   for the core-owned `SourceSnapshot`.
2. Build exactly one `LintPlan` from that generation's Programs, exact target
   projection, and rule resolver. The plan is the only file/rule authority for
   both native and third-party plugin work. Rare syntax failures are frozen as
   sparse plan-level diagnostic groups instead of enlarging every file's hot
   execution record. Parallel preparation publishes only after every worker
   joins; cancellation returns no partial plan, and a worker panic is re-raised
   on the caller goroutine so the generation lease is still released.
3. Validate the generation's stable target projection once, retaining its
   target/SourceFile result only when the observation's artifact demand asks
   for it (the API does; CLI and LSP do not).
4. Join the observation's required producers against one generation text view,
   freeze the exact fix text snapshot, and compute deterministic whole-file changes. An
   initial disk-backed CLI plugin pass may use the host filesystem fast path;
   detached work and every later memory generation carry complete inline
   plugin inputs so one observation cannot mix memory and stale disk.
5. Validate and apply the entire change set atomically to the pipeline's private
   memory state. No integration mutation or persistence callback runs between
   fix rounds; the generation provider only projects the immutable snapshot
   needed for the next observation.
6. Re-materialize and re-observe the new snapshot until stable, restored to the
   initial state, or bounded by the linter-owned ten-round product limit; an
   adapter that returns a stale generation is rejected before execution.
7. After successful completion only, optionally call `FinalChangeCommitter`
   once with the net initial-to-final delta. Intermediate changes are never
   exposed as persistence commands.

These are semantic ports rather than generic pre/post hooks:

- `GenerationProvider` is the pre-observation boundary. It may map a snapshot
  to an overlay VFS, an editor overlay, or another source medium and build the
  corresponding immutable generation, but it cannot advance fix state.
- `TargetProjection` maps Program source paths to stable target identities and
  exposes the exact complete text of that generation. Core validation, frozen
  fix text, and inline plugin input all consume this same view.
- `FinalChangeCommitter` is the optional terminal persistence boundary. It
  is independent from generation construction, confirms complete file
  replacements, and is never called after an unsuccessful observation or
  between rounds.
- `ProgressiveDiagnostics` owns LSP-specific baseline presentation and detached
  enrichment admission without becoming a second lint orchestrator.

The linter files mirror those responsibilities: `pipeline_contract.go` and
`pipeline_result.go` define the public boundary; `pipeline.go` is the sealed
entry; `pipeline_generation.go`, `pipeline_observation.go`, and
`pipeline_native.go` own one immutable observation; `pipeline_plugin.go` owns
plugin task materialization; and `pipeline_autofix.go`,
`pipeline_autofix_state.go`, `pipeline_fix_text.go`, and `pipeline_fix.go` own
the in-memory autofix state machine and its pure text transformations.

Important behavior differences by integration:

- **CLI lint-only**: requests diagnostics only, so migrated native rules do not
  construct autofixes or suggestions
- **CLI `--fix`**: rebuilds each Program generation over the core-owned memory
  snapshot, uses diagnostics-only mode for the final no-more-rounds verification,
  then validates and writes each net-changed disk file once
- **LSP quick fix**: returns direct text edits for one diagnostic
- **LSP fix-all**: materializes isolated speculative generations for repeated
  core-owned memory rounds, then returns one whole-document replacement edit
- **LSP normal diagnostics and API**: request all native and third-party plugin
  edits because fixes and suggestions are response metadata even when they are
  not immediately applied
- **LSP speculative fix-all passes**: request native autofixes only
- **API**: `lint({ fix: true })` selects the same linter-owned ten-round autofix
  lifecycle as CLI and requests a complete final observation. Diagnostics,
  counts, encoded sources, and remaining edit metadata therefore describe the
  final in-memory source; `output` includes every file to which a fix was
  applied, even if later rounds restored its initial text. The JS side alone
  persists that output through `Rslint.outputFixes`.

## 8. Configuration & Directives

### Configuration Formats

Rslint supports two configuration formats following ESLint flat config semantics (array of config entries):

#### JS/TS Configuration (Recommended)

Rslint automatically discovers `rslint.config.js`, `rslint.config.mjs`, `rslint.config.ts`, and `rslint.config.mts`. Explicit configuration paths also support `.cjs` and `.cts` files through CLI `--config` and API `overrideConfigFile`. JS/TS config files support preset composition via `defineConfig()`. The package root exports the complete catalog from its pinned build-time `globals` dependency, so consumers use the same flat-config composition shape without installing another package. During the library-surface Rspack compilation, an asset plugin mechanically splits every upstream top-level set into `dist/globals/<name>.json`. The root contains only the ordered upstream set-name table and synchronous accessors: importing it or calling `Object.keys(globals)` parses no environment data; the first `globals.browser` access loads and caches only `browser.json`. Rslint-specific sets that do not exist upstream are small ordinary objects registered directly in the runtime catalog; they neither pass through the asset plugin nor issue a runtime `require`. Operations that read every upstream value, such as `{ ...globals }`, intentionally load every upstream set. These assets share the existing worker distribution contract: direct package use works, while a redistributor must preserve the complete `dist` layout or leave `@rslint/core` external instead of treating one entry file as a self-contained bundle. The pinned package's exact readonly/literal declarations are shipped in the private `dist/globals/index.d.ts` module; the root declaration uses an `import()` type query, so upstream declaration names stay isolated and consumers have no dependency on `globals`.

A selected map enters the existing `languageOptions.globals` path as an explicit declaration: config matching, flat merge, the Go rule runtime, and the ESLint-plugin worker all consume the same effective globals. This does not enable any runtime environment by default or change the parser edition. The data assets are emitted only with the root library surface, not duplicated into the service, internal, or worker outputs. The worker keeps a small edition-aware ECMAScript table aligned with the Go catalog; for TypeScript ASTs it reconciles scope-manager's value bindings to that table while retaining its type-only lib bindings.

```typescript
import { defineConfig, globals, js, ts } from '@rslint/core';

export default defineConfig([
  {
    ignores: ['**/dist/**', '**/fixtures/**'],
  },
  js.configs.recommended,
  ts.configs.recommended,
  {
    files: ['**/*.js'],
    languageOptions: { globals: globals.node },
  },
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

**Key difference**: JSON configs are normalized against the explicitly supplied Go rule catalog, which auto-enables core rules and rules from declared bundled plugins unless explicitly overridden. JS/TS configs only enable what the normalized config entries specify, usually via presets.

### Config Entry Structure

Each entry in the config array supports:

| Field             | Type                                       | Description                                                                          |
| ----------------- | ------------------------------------------ | ------------------------------------------------------------------------------------ |
| `files`           | `(string \| string[])[]`                   | Non-empty selector list; top-level selectors are ORed and nested selectors are ANDed |
| `ignores`         | `string[]`                                 | Glob patterns excluded by this entry                                                 |
| `languageOptions` | `object`                                   | `ecmaVersion`, globals, and parser options such as project settings                  |
| `rules`           | `Record<string, …>`                        | Rule level or `[level, options]`                                                     |
| `plugins`         | `string[] \| Record<string, ESLintPlugin>` | Bundled plugin declarations or third-party plugin instances                          |
| `settings`        | `Record<string, …>`                        | Shared settings available in `RuleContext`                                           |

### Configuration Loading

The loading flow differs by config type:

**JS/TS staged catalog discovery**:

CLI, the native JavaScript API path, and transactional LSP refreshes reuse
the one-shot `internal/config/discovery.DiscoverAutomatic` operation (or
`LoadExplicitConfig` for an exact path). Automatic discovery builds an immutable
config/ownership catalog and observes `.gitignore` sources during the same
directory walk. Explicit loading first selects the exact module unconditionally,
then freezes that invocation-wide owner's Git projection with the same frontier
without probing nested config candidates. Neither path collects lint targets.
For LSP, the client's first `rslint/configRefresh` may include one absolute
`configPath`. The server fixes that optional path for its lifetime: later client
refreshes must repeat the same choice, while Go-owned `.gitignore` refreshes
reuse it internally. Changing between automatic discovery and an explicit path,
or changing the path itself, requires a new server process. Explicit LSP mode
uses only the selected JS/TS module and does not load JSON fallback config.
Go owns candidate discovery, default exclusions, config hierarchy, authored and
Git directory reachability, the frozen Git projection for each owner, and final
effective IDs. Node only
executes the exact JS/TS modules requested by Go, normalizes their entries,
retains live third-party plugin objects, and returns serializable entries. Source
fingerprints stay in the Node transaction session; after Go selects the final
effective IDs, activation revalidates those fingerprints and returns only the
effective plugin metadata.

The package boundary is deliberate: `internal/config/discovery` imports the
parent `internal/config` model and its narrow authored-global-ignore matcher,
plus `internal/config/target.OwnerScope` as the typed provenance handed to
target planning. The root config package imports neither child; target planning
imports only the root config package and never discovery. Runtime file routing
is owned by `internal/config/target.OwnerIndex`, while CLI/API/LSP adapters own
transport, commit/abort, and last-good lifecycle. Discovery has no cross-transaction session,
synchronization, or generation state because every production request is one
transaction; request-local coordination only freezes concurrent observations.
A process-random nonce plus atomic sequence allocates IDs that cannot collide
with a stale host session after a native-process restart. The returned catalog
publishes final configs, scopes, failures, effective IDs, plugin metadata, and
whether the invocation used an explicit config. Candidate fingerprints and
plugin-aggregation scratch remain private to the Node transaction session;
source-selection scratch remains private to the Go discovery transaction.

Within `internal/config/discovery`, one request-scoped coordinator owns only
phase order and final statistics. Ownership resolution and directory walking
find candidate boundaries; the module-load coordinator owns candidate identity,
Node batches, failures, and final activation; the Git projection owns frozen
`.gitignore` reads and materialization; and the catalog draft is the sole writer
of selected configs and owner scopes. Filesystem workers return frontier results
without mutating those shared transaction results. The coordinator merges each
sorted frontier, loads its candidate batch, and adopts the result before the next
frontier starts; Node activation happens only after catalog finalization.

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
  walker are separate traversals, but the staged catalog already freezes the
  Git sources observed on its reachable frontier. There is no second per-owner
  directory sweep. Automatic literal scopes and explicit file-only invocations
  contribute only their exact owner-to-target source chains to the same
  source-keyed projection.
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
- The extension owns shared UI, commands, and output channels once. For each
  open supported file, `CoreResolver` follows normal `node_modules` ancestry or
  the resource-scoped `rslint.corePath` override to an exact
  `@rslint/core` package. There is no bundled fallback and no alternate PnP
  execution path. That package supplies the Go binary, config-module host,
  protocol version, and lazily loaded ESLint-plugin worker host as one coherent
  runtime boundary.
- `RuntimeManager` keys a runtime by VS Code workspace-folder URI plus the
  package directory's normalized physical realpath. Documents resolving through
  symlinks to the same installation share its process and worker state; distinct
  physical copies remain isolated even when their version strings match. Only
  installations selected by currently open documents start. Reference release
  closes an unused runtime, while a document switch starts and validates the
  replacement before withdrawing its last-good owner. An unconfirmed close
  quarantines that exact runtime key so a replacement cannot overlap a possibly
  live process. Terminal shutdown still attempts every runtime close before
  releasing shared resources.
- Each runtime owns its native server children, config watcher, transaction
  adapter, plugin pool, request handlers, and workspace logger. A process owner
  covers automatic LanguageClient restarts, terminates any still-live prior
  child before a restart spawn, blocks new spawns once closing begins, and
  awaits stdio close after bounded forced termination of any child that survives
  protocol shutdown. A closing-aware client error handler forbids restart.
- Every runtime retains a workspace-relative document selector, while
  `WorkspaceDocumentRouter` is the single authority for the selectors' overlap.
  The manager assigns each document explicitly to the runtime selected by local
  package resolution. An assignment change performs an ordered
  `didClose`/diagnostic-clear/`didOpen` handoff using the document's current
  in-memory text, without requiring the editor to close. Middleware admits
  changes, saves, diagnostics, and code actions only for the exact active
  runtime that currently owns the server-open document. Exact runtime identity
  rejects diagnostics from a replaced same-key client, while a document epoch
  rejects code actions that finish after an ownership change. When the
  LanguageClient automatically restarts a native server, the router invalidates
  every recorded server-open session for that runtime before LanguageClient
  replays `didOpen`, so the replacement process receives each still-open owned
  document exactly once.
- Each selected runtime starts `rslint/configRefresh`. With no `configPath`, Go
  scans that process's workspace-folder cwd with a transaction-scoped cached
  VFS. A client may instead repeat one fixed absolute JS/TS `configPath` on
  every refresh; Go loads exactly that module, uses the module's directory as
  its authored config base, and retains the process cwd as the invocation-wide
  target scope. The client owns change notifications
  for the explicit path, while Go's existing `.gitignore` watcher refreshes
  the already-fixed source. Go sends
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

LSP keeps config evaluation and project selection as two orthogonal
generations. A config transaction atomically publishes entries, owner indexes,
path-space snapshots, per-owner file resolvers, the rule catalog, and the
matching Node plugin-host generation. Watched tsconfig changes update only the
project-selection view derived from those committed declarations and invalidate
resident project metadata/Programs; they do not re-evaluate a JS/TS config
module or replace its rule/path-space generation. Each document lint snapshot
samples one committed config generation together with the current project view
on the serialized server dispatch loop.

Within `internal/lsp`, `server.go` remains the single transport, dispatch-loop,
and mutable-state owner. `initialization.go` constructs its session;
`config_discovery.go` owns JS/TS transaction prepare/commit/abort, while
`config_watch.go` owns watcher registration and event handling together with
JSON fallback and tsconfig-derived state refresh. `document_sync.go` owns the
open-buffer mirror and debounce state.
`document_lint_snapshot.go` then freezes target identity, config ownership,
rule catalog, and declared projects before `lint_generation.go` materializes a
leased diagnostics generation or `lint_fix_all_generation.go` materializes an
isolated fix-all generation. `diagnostics.go` and `code_action.go`
construct core requests from that same snapshot, and `eslint_plugin.go` returns
through the server-owned generation checks. These are LSP adapters around the
single `internal/linter` pipeline, not independently stateful services.

An explicit JS/TS `--config` or API `overrideConfigFile` bypasses automatic
candidate selection and loads the exact module. Its directory remains the
authored base for relative config content, while the invocation cwd remains the
implicit scan and response-path root. The exact config path is never gated by `.gitignore`;
only after it loads does a fixed-owner frontier freeze the invocation-scoped
Git projection used to filter lint targets. That frontier never probes or
activates nested config candidates. Automatic candidates instead use Git
directory reachability while selecting ownership.

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
LSP keeps it as the Go-loaded fallback for files with no JS owner. The
API discovery path has no disk JSON fallback; low-level API callers may instead
send an already-resolved serialized `config`.

### Configuration Merging

Config merging follows flat-config-style semantics in `GetConfigForFile()`:

1. entries containing only `ignores` and an optional `name` form the global-ignore set
2. the implicit default extension baseline plus effective explicit `files` selectors defines the config selector union; an entry's `ignores` prevents its selector from extending that union, top-level selectors are ORed, nested patterns are ANDed, and an explicit `files: []` is invalid
3. entries without `files` cascade across that selector union, while entry-level `ignores` prevent only that entry from contributing configuration to an otherwise selected file
4. later rule values override earlier values; a severity-only override retains earlier rule options
5. settings and language options recursively merge ordinary objects, while later arrays and scalar values replace earlier values

`FileConfigResolver` applies that policy in two phases: it first matches a file
to the exact ordered set of contributing entries, then compiles the set into an
immutable effective config/rule plan. Files with the same set share one plan for
that resolver's lifetime. Shape identity is collision-free (`uint64` for the
first 64 entries plus the complete remaining bitset), resolver-local, and
independent of path identity. File-cache keys remain the exact caller strings;
filesystem identity may select the matching space, but it never changes the
cache identity or merges distinct lexical targets.
The direct `GetConfigForFile()` compatibility path retains its allocation-light
single-origin matcher and delegates composed multi-origin arrays to the target
resolver; both paths feed the same ordered merge policy without retaining a
plan cache.

The staged coordinator builds the effective catalog used by
`target.OwnerIndex` before config's target matcher merges the selected entries.
CLI/API target plans call ownership once during discovery and carry that owner
through Program binding. Each target records four separate facts: the caller's
lexical `Path`, the file's `CanonicalPath`, the lexical parent's physical
`CanonicalParentPath`, and the governing `ConfigDirectory`. The parent identity
distinguishes a leaf file symlink from a directory alias; it is not treated as
proof of the target's complete ancestor chain. Native-case owner aliases may be
verified once while the target's owner is selected. LSP then freezes one
document lint snapshot containing the same complete target plus its selected
owner, resolved config, matching rule catalog, and project paths; Go rules, plugin rules,
diagnostics, resident Programs, and every fix pass consume that snapshot. No
later stage resolves the target file or selects its owner again. Staged CLI,
native API, and transactional LSP paths therefore reuse the same Go ownership
rules instead of independently reconstructing hierarchy on the Node side.

Configuration meaning and invocation scope are separate inputs. For one
invocation-wide config, `ConfigDirectory` is the config file's directory and is
the authored base used by `files`, `ignores`, and `parserOptions.project`;
`ScanRoot` is the invocation working directory used by implicit/no-argument
target discovery and the default `.gitignore` collection scope. Supplying an external config therefore does not scan
the config directory unless the user also targets it, and does not rebase the
config's relative patterns onto the invocation cwd. A composed config entry may
retain a different authored base (the native API's inline `overrideConfig` is
the current case), so matching resolves every entry from its own origin before
merging the matched shape, and project declarations resolve from that same
origin before they reach the Program loader. Automatic config catalogs already
encode one owner directory per config and do not use the invocation-wide
`ScanRoot` to infer ownership.

`target.Request` is the lint-target boundary shared by CLI/API target
selection. Explicit files and recursive directories form one union; omitted
CLI targets become the invocation cwd, and multiple files/directories do not
change one another's config matching. Each planned target retains its
caller-visible path, file and parent filesystem identities, and established
config owner. For each literal file request, the plan also carries the
existence and ignore outcome produced by that same discovery decision; CLI
warning rendering never re-reads the filesystem or re-resolves ownership. The
plan exposes the read-only `config.PathSpaceSnapshot` observed during discovery,
but does not own a rule catalog or execution config. CLI `--rule` and API
overrides are layered after target selection and config resolvers evaluate them
in those same frozen path spaces.
`internal/program/loader` consumes the plan but cannot add targets, change
ownership, or reinterpret config-relative paths.

After Program binding, `internal/config/lint` maps each Program source path
through its `Resolver` back to that immutable selected target and resolves rules
only through the target's recorded config owner. In invocation-wide single-config mode an
unbound source still resolves against that one config; in an automatic
multi-config catalog an unbound source receives no config. CLI and API share
this decision and do not reconstruct owner selection independently.

Git collection scope is independent from config-entry origin. Targets
representable beneath `ScanRoot` share its scope; every requested directory
outside that tree creates an independent scope, while an exact outside file
does not invent a directory scope. The synthetic Git entry retains every active
scope, including a scope that produced no patterns. Matching first assigns a
target to exactly one scope by lexical containment, falling back to the same
ordered scope list by canonical-to-physical containment only when no lexical
scope applies. Final ignore decisions, directory pruning, and negation-based
reopening all consume that one assignment, so aliases and sibling directory
arguments cannot borrow or override one another's `.gitignore` policy. A
requested physical ancestor of an aliased `ScanRoot` is outside, not beneath,
that root and therefore receives its own scope; this keeps every directory the
target walker can scan covered by the same Git-source frontier.

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
  `ConfigWithCollectedGitignore`/`ConfigWithGitignore` policy. Automatic config
  owners use their config directory as the hard upper collection boundary; an
  explicit invocation-wide config instead uses the invocation Git scopes
  described above, even when the config file is external. Sources below each
  scope boundary apply, while sources in its parents do not. In staged JS/TS
  catalogs, the walk records sources by owner, scope, and case-aware source
  identity, orders them parent-before-child within that scope, and
  materializes the synthetic Git entry before publishing. Direct automatically
  reachable child config directories are
  downward ownership handoff boundaries; an explicitly selected config remains
  the fixed invocation-wide owner and never creates nested handoffs.
  Configs loaded only for explicit targets bound only their literal target
  chains, so adding a literal cannot truncate an ancestor-owned automatic
  target's `.gitignore` sources. This preserves ESLint v10's per-target global
  ignore behavior: adding another literal target cannot change whether an
  existing target is ignored. File-only CLI/API requests read only target
  directory chains within each governing config. Explicit JS/TS and JSON CLI
  directory requests read the target ancestry and then recurse only below the
  requested directories; mixed requests add the exact-file chains. Automatic
  config discovery keeps its existing ownership walk. The synthetic Git entry
  is ordered before authored entries, so a later config `!` may re-include a
  target
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
  worker, LSP commits the ordinary Go config with an empty no-host plugin
  state and retries on later refreshes; once a usable snapshot exists, the same
  worker failure aborts and preserves that last-good snapshot. A successful
  no-candidate transaction removes the previous JS catalog and
  exposes the Go-loaded JSON fallback
- bundled Go and third-party object-form plugin rules are gated by their normalized prefixes for JS/TS configs; each CLI run and API request derives its own catalog, while LSP commits a JS-owner catalog with the matching Node generation and retains the pure Go catalog for JSON-owned files. Replacing or removing plugins therefore replaces the JS catalog instead of retaining process-wide placeholders
- CLI/API lint target selection is independent from TypeScript `Program` membership and considers only rslint-supported script extensions. The `.js`, `.mjs`, `.cjs`, `.jsx`, `.ts`, `.tsx`, `.mts`, and `.cts` default baseline is always selected; explicit config `files` contributes candidates only within the supported set. Global ignores and `.gitignore` remove targets, while an entry-level ignore prevents only its own selector/config contribution
- selected CLI/API targets can still appear as 0-rule lint results when no config entry contributes rules; this applies to default-baseline directory discovery and explicit supported files, and syntax diagnostics remain available in that state
- under automatic discovery, each selected file is governed by its nearest loadable config; an explicitly selected config is used directly. In either case, a target can bind only to a tsconfig declared by its governing config. The first declared project whose parsed root set contains the file wins. Only when no declared root contains it does the first declaration-order Program containing it through imports win
- omitting `parserOptions.project` enables the governing config directory's default `tsconfig.json` fallback; an explicit empty project list disables that fallback
- `files`/`ignores` matching uses the stable target path in the governing config's path space; a ts-go Program source alias is used only to locate the AST and type information, so moving a target into or out of a tsconfig cannot change its rule configuration
- within each Program-registry build, normalized declared tsconfig paths are deduplicated across config associations; each CLI autofix observation creates a new Program generation over the core-owned memory overlay and reruns focused project selection without writing the base filesystem. Import-only fallback membership is therefore recomputed after source edits. File-symlink declarations remain distinct because TypeScript resolves relative paths from the declared location. Selected CLI files outside the governing config's ts-go Programs are parsed and bound as independent rslint Programs. Targets whose names collide under a case-insensitive ts-go path key are partitioned across independent Programs so distinct physical files and package scopes remain distinct
- `--type-check` and `--type-check-only` build every real tsconfig declared by the effective loaded config catalog. Git reachability may change which automatic configs enter that catalog; once it is established, program-wide checking is not filtered by lint targets, config `files`/`ignores`, `.gitignore`, or CLI file/directory arguments. Only those project-backed source generations expose the program-diagnostics capability consumed by Phase 2; `--type-check-only` skips the separate lint-target walk.
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

The default Go CLI has no command-local lint pipeline facade. `lint_command.go`
prepares the request and calls `internal/linter.RunPipeline` once;
`lint_generation.go` maps core-owned memory snapshots to request-local overlay
VFS generations; and `lint_commit.go` implements only the optional one-shot terminal
disk commit. Other supporting files retain bounded command concerns:
`lint_options.go` parses and validates flags, `lint_configuration.go` owns
catalog warnings, `lint_diagnostics.go` owns CLI path/warning normalization,
and `eslint_plugin.go` owns CLI plugin routing and stderr policy. Target
selection remains in `internal/config/target`, source-to-config rule resolution
in `internal/config/lint`, Program construction in `internal/program/loader`,
the complete lint/fix lifecycle in `internal/linter`, and report rendering in
`internal/output`.

The Go API mode similarly prepares one request in
`internal/api/server/lint.go` and calls `RunPipeline` once. Its generation provider
maps any core snapshot onto the request's existing file-content/canonical-path
overlay; the core-owned bounded autofix lifecycle returns a verified final
observation plus per-fixed-file source for the response instead of requiring a
separate API fix path. API-specific path
bases, opaque plugin routing keys, reverse transports, structured responses,
and the inspector cache remain in `internal/api/server`; planning, execution,
and mutation sequencing do not.

1. **Node.js Wrapper**: parses args, starts the Go engine, and hosts JS/TS module evaluation plus live third-party plugin objects
2. **Config Catalog**: for automatic JS/TS discovery or an explicitly selected JS/TS config, Go builds the staged catalog and batches exact module-evaluation requests to Node. If automatic discovery finds no candidate, or a non-JS config was explicitly selected, the existing Go JSON loader path remains in control
3. **Mode Selection**:
   - `--lsp`: starts the LSP server
   - `--api`: starts the IPC API server
   - default: runs direct CLI linting
4. **Lint Target Plan**: Go resolves a stable target set from CLI/API scope, the implicit default baseline, explicit config `files`, global ignores, and `.gitignore`
5. **Project Loading**: one request-scoped `program/loader.Session` resolves normalized tsconfig identities once per load. Plain lint passes the complete resolved config catalog and target plan to the loader; the loader admits project declarations only from config owners represented in that plan. A focused selection execution retains parsed root metadata while choosing direct winners, then builds the deduplicated winner set in parallel; targets with no direct winner advance an ordered fallback frontier. Full-CWD/no-argument lint eagerly builds every declaration from those active owners, while program-wide type-check modes build every project from the complete effective catalog. Shared declared paths preserve each active config association and declaration order.
6. **Target Binding**: the loader binds each target by exact lexical or canonical filesystem identity in two tiers: first declared direct root, then first declaration-order import-containing Program. It creates source-only Programs for supported unbound CLI targets, including projects with no tsconfig, and returns one ordered Program sequence plus parallel target projections. CLI/API never coordinate compiler construction or implement ownership themselves
7. **Core Request**: CLI/API/LSP choose a sealed `RunPipeline` request kind and provide semantic adapters. They do not call `PrepareLintPlan`, `RunLinter`, plugin dispatch, or fix application as separate stages.
8. **Rule Plan Preparation**: for each acquired generation, the core calls `PrepareLintPlan()` with one ordered Program sequence, its exact parallel target projection, and the generation's rule resolver. Planning never discovers, excludes, or scans for fallback targets; a target absent from its bound Program is an invariant error. The immutable result owns that Program sequence, freezes rare syntax diagnostics in a sparse plan-level projection, and keeps each hot per-file execution unit limited to its source, resolved rules, shared rule environment, and checker capability. Syntax-error and zero-rule files remain selected, and the same non-empty file/rule projection feeds third-party plugin dispatch. Worker results are joined in stable Program/file order; cancellation cannot publish a partial plan, and callback panics return through the caller so the enclosing generation lifecycle remains exact-once.
9. **Rule Execution**: the core invokes `RunLinter()` only with the prepared plan plus scheduling, diagnostic, timing, and optional program-wide type-check concerns. A nil plan explicitly skips lint for `--type-check-only`. Execution never recollects files, re-resolves rules, or accepts a second Program authority alongside a plan. Plans bind exact Program pointers, so every autofix re-observation constructs a fresh Program and plan over the current in-memory snapshot. When `--type-check` is enabled, Phase 2 schedules only Programs that expose complete program diagnostics.
10. **Fix Rounds and Aggregation**: `RunPipeline` joins native/plugin results, projects stable target paths, computes whole-file changes, applies them to private memory, and reacquires a generation until stable or bounded. A file may move between Program generations when its import graph changes, but the target plan remains stable. Once memory differs from the initial generation, every plugin target is frozen inline from that generation even when the initial CLI host could read disk. CLI independently supplies a terminal committer and therefore touches disk only once after the last successful observation; API and LSP consume the returned memory delta without a committer. Integrations retain only presentation policy for structured plugin notices, path conversion, and output/protocol mapping.
11. **Report Assembly**: the CLI builds one output report from the final post-fix diagnostics plus run metadata. Diagnostics carry an explicit lint or TypeScript origin, and the report computes error/warning/type-error counts once so the summary and exit policy use the same values; `--quiet` filters rendering only.
12. **Output Formatting**: the CLI-private output subsystem renders `default`, JSON line, GitHub workflow command, or GitLab Code Quality formats. Only `default` emits a summary; machine-readable formats never emit ANSI styling or a summary.
13. **Output Synchronization**: in the default IPC CLI, Go drains all report text into ordered `output` notifications and then sends its terminal `shutdown` request. Node seals forwarding and waits for every real-stdout write callback before acknowledging; only after Go receives that acknowledgement does it emit deferred stderr (currently the `--timing` table), so merged `2>&1` output cannot place the table before or inside the report.
14. **Exit Code**: depends on the report counts, `--max-warnings`, and fix outcomes

### Concurrency Model

The main Go workload work groups and pools below honor `--singleThreaded`.
The flag serializes these workload stages, but IPC transport, diagnostic
collection, and plugin dispatch may still use infrastructure goroutines.

1. **Lint-plan work group** (`PrepareLintPlan()` via `core.NewWorkGroup`)
   - Resolves per-file rules into stable rslint Program slots with at most
     `min(GOMAXPROCS, selected file count)` workers.
   - `--singleThreaded` resolves the same slots serially in Program/file order.

2. **Linter work group** (`RunLinter()` via `core.NewWorkGroup`)
   - Schedules one task per rslint Program; checker-capable work retains
     checker-exclusive shards, while sufficiently large checker-free Programs
     use bounded file chunks.
   - `--singleThreaded` collapses the work group to serial execution.

3. **Type-check work group** (`runTypeCheckAcrossPrograms`)
   - Schedules diagnostics for real tsconfig Programs and merges results in
     stable Program order.
   - `--singleThreaded` computes Program diagnostics serially.

4. **Staged catalog discovery and Program creation**
   - A bounded Go worker pool scans one reachable sibling frontier at a time.
     Config-boundary nodes are suspended, batched after that frontier is
     processed, and resumed only after the Node result is merged.
   - Native API roots carry an exact target-ancestor trie from the tinyglobby
     result, so only sibling branches leading to supplied files enter those
     frontiers. CLI/LSP directory roots, including CLI mixed file+directory
     invocations, remain recursively unbounded during automatic config
     discovery. After an explicit JS/TS config loads, its fixed-owner Git
     projection instead carries the CLI target trie: exact-file branches stop
     at their parent, and requested directory branches recurse from that root.
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
   - Configured Program identities, config associations, and result slots are
     planned serially in stable config/project order. Construction then uses at
     most `min(GOMAXPROCS, Program count)` workers and merges results and
     errors by the planned order.
   - Focused lint parses each governing config's tsconfig frontier in declaration
     order. Independent config frontiers and confirmed direct winner Programs
     may run concurrently; import-only fallback remains serial within one
     governing config so a later completion cannot overtake an earlier project.
   - `--singleThreaded` executes the same state machine with one Go discovery
     worker and serializes module evaluation within each Node frontier batch.
     Coordinator batches and results remain ordered in either mode.

5. **Lint-target directory walker** (`internal/config/target`)
   - `target.Resolve` drives the private fixed-size worker pool (`walkPool`)
     that walks the directory tree and publishes only the resulting `Plan`.
     Live goroutine count is capped at `workers`, not the number of directories.
   - Default `workers = max(2, GOMAXPROCS)`; `--singleThreaded` forces
     `workers = 1`, which degenerates into a fully serial DFS-style traversal.
   - The walker reads sorted VFS directory entries without following nested
     symlinked subdirectories. An explicitly requested
     directory alias is resolved once and used as a bounded scan root. This
     matches ESLint v10's
     flat-config file walker, which uses `@humanfs/node` and recurses only
     when `Dirent.isDirectory()` is true. Node's
     `readdir({ withFileTypes: true })` reports the dirent type without
     following symlinks, so
     `Dirent.isDirectory()` is false for symlinks. The skip also eliminates
     scheduling-dependent non-determinism that a parallel walker would
     otherwise introduce.

6. **Program source identity index** (`internal/program/loader`)
   - Direct lexical Program lookups remain synchronous. The target identity
     maps are not allocated until one misses. The binder then builds the
     governing config's Program indexes in one canonicalization batch, reusing
     target identities established by discovery before resolving unknown source
     paths. Programs associated only with other configs remain untouched.
   - Unknown source identities are cached once across the Programs actually
     inspected during that binding pass. With complete VFS symlink metadata,
     regular files in one lexical directory share its canonical directory
     lookup; file symlinks, ambiguous casing, missing entries, and VFSes without
     that metadata fall back to per-file realpath. Per-Program indexes retain
     only canonical identities present in the target plan. CLI autofix
     observations create a fresh index when they rebind the target plan over
     the current memory overlay.
   - Independent realpath lookups use `core.NewWorkGroup`; `--singleThreaded`
     runs the same work serially.

Other invariants:

- Target discovery returns both the caller-visible lexical path and a canonical
  identity hint. A regular directory walk derives the canonical path from the
  canonical config root without a per-file realpath call. Explicit directory
  aliases are resolved once and their descendants inherit the corresponding
  physical path; explicit files and file symlinks are resolved individually.
  A custom VFS whose `Entries.Symlinks` is nil has not established file
  identity metadata, so the target walker conservatively resolves each selected
  file instead of treating it as regular.
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
| Linter work group             | Program work, including checker-free file chunks, collapses to serial execution.                              |
| Type-check work group         | Program diagnostics are computed serially via `core.NewWorkGroup(true)`.                                      |
| Staged catalog discovery      | Sibling-frontier workers and Node module evaluation are serialized; batches and catalog merge remain ordered. |
| Configured Program builds     | Planned and constructed serially in stable config/project order, retaining fail-fast error behavior.          |
| Lint-target walker workers    | Forced to 1 (single goroutine, no concurrency).                                                               |
| Program source identity index | Canonical source paths are resolved serially through `core.NewWorkGroup(true)`.                               |

These workload stages run serially with `--singleThreaded`; infrastructure
goroutines remain outside that guarantee.

## 10. Performance & Memory Considerations

### Design Principles

- **Direct ts-go Data Model**: rslint Programs preserve ts-go AST objects directly and provide optional checker access without converting through a second AST representation or exposing their private adapter
- **Unified Program Scheduling**: configured ts-go Program construction uses a bounded, stable-slot worker pool, and `RunLinter` queues subsequent lint work over one ordered rslint Program sequence through `core.NewWorkGroup`; checker-free Programs are additionally split into bounded chunks. `--singleThreaded` forces all of these flows to run serially
- **Single-Walk Rule Dispatch**: each file is traversed once, with rules registering listeners up front and sharing the same AST walk
- **Early Filtering**: exact lint target plans, skip paths, global-ignore filters, and per-file capability filtering reduce work before listeners run
- **Shared Prepared Lint Plan**: CLI and native API resolve each selected
  file's complete rule set once into immutable rslint Program slots. Native
  execution and optional third-party plugin dispatch consume projections of the
  same plan instead of repeating target collection or rule resolution

### Performance Optimizations

- **Native Go Implementation**: Eliminates JavaScript runtime overhead
- **Direct TypeScript AST**: No AST conversion between parsers
- **Shared Type Checker**: `runLintRulesInProgram` acquires one checker for the lint phase and reuses it across files and rules in the same `Program`. The subsequent type-check phase (when `--type-check` is enabled) lets `GetSemanticDiagnostics` reacquire its own checker so the lint-phase checker can be released first
- **Checker Phase Separation**: the checker is released before TypeScript semantic diagnostics run, so `GetSemanticDiagnostics` can reacquire its own checker cleanly
- **File Filtering**: Skip node_modules and bundled files automatically
- **Source-Only Programs**: project-unbound CLI roots use the ts-go parser, binder, package/module metadata, and direct-import resolution to construct the same rslint Program facade without a synthetic project graph or checker; capability gating excludes type-aware rules
- **Shared Cross-File Rule Structures**: module edges, export maps, and whole-Program indexes are immutable derived values cached by Program generation. Project adapters preserve the underlying ts-go generation's weak lifetime; root-parsed Programs own the same cache directly. Rules do not select either strategy
- **Buffered Diagnostic Collection**: CLI mode funnels diagnostics through a buffered channel before formatting, which reduces contention between lint tasks and output handling
- **On-Demand AST Encoding**: API/WASM responses only include encoded source files when `IncludeEncodedSourceFiles` is requested
- **Lazy Shared Comments**: each file owns one `CommentStore`; directive consumers and comment-aware rules materialize its canonical comment list only when needed and reuse the result. Rule-specific text checks avoid scanner work when their comment syntax cannot occur
- **Lazy Shared References**: each file exposes one `RefStore`; its candidate identifier walk is deferred until first use, and each queried symbol name is binder-resolved once for reuse across native rules
- **Task-Local Listener Reuse**: each checker-shard task owns one sparse listener registry and reuses its map and per-kind slice capacity across that task's serial files. Registries are never shared across tasks or requests
- **Direct Rule Reporting Methods**: each rule context stores one compact immutable reporter state; its reporting methods replace the former family of per-rule bound closures
- **Demand-Driven Native Edits**: native consumers explicitly request optional
  edit kinds. Deferred edit builders run only after suppression and only when
  their matching kind is requested, so lint-only runs preserve diagnostics while
  avoiding edit-only analysis. Eager reporting methods remain compatible while
  rules migrate incrementally.
- **Exact Config-Shape Interning**: a `FileConfigResolver` parses global and
  entry-local ignores once, maps each exact file path to its complete matched
  flat-config entry bitset, and merges/prepares enabled rules once per unique
  bitset. The prepared lint plan references these published immutable rule
  slices without moving per-file runtime state into the config cache. LSP keeps
  its existing per-document resolver lifetime.

### Caching Strategy

- **LSP Session Reuse**: `internal/lsp` builds a shared ts-go `project.Session`, so configured projects, inferred projects, and overlay document state are reused across requests.
- **LSP Configured-Project Selection Cache**: normal diagnostics and speculative fixes share one root-first selector while keeping separate Program owners. Session projects cache exact/canonical root membership by authored `ParsedCommandLine` generation. Session-external projects retain lightweight parsed metadata only after all config and directory reads have watcher coverage, then reuse that same parse when a Program is needed. Only a Program that contains the requested target becomes resident; non-containing import probes are transient. Config, project-shape, and covered filesystem changes discard affected metadata and Programs. Registration failure and conflicting editor aliases use fresh request-local construction. Speculative fix Programs always remain isolated. This path is independent from the CLI/API request loader; no second LSP Session or rslint ParseCache is introduced.
- **Parse Cache in LSP**: the LSP server passes a shared ts-go `project.ParseCache` into the session to avoid re-parsing from scratch on every change.
- **SourceFile-Owned Module Syntax Reuse**: any `Program.ModuleGraph()` query may attach the syntax-only module-specifier projection to an immutable ts-go `SourceFile`, independently of project ownership or watcher availability. Programs reusing that exact file object share collection, while resolution and target ASTs remain local to each Program generation. Attached values may contain only scalars and nodes owned by that SourceFile—never a Program, resolved target, checker state, or another SourceFile. Replaced files carry the attached data out with their own AST lifetime; no LSP server map, project-membership sweep, or explicit reset owns this cache. Source-only Programs retain it only for their generation's SourceFile lifetime.
- **Incremental LSP Document Sync**: editor changes reach the server immediately and are applied in order to both rslint's document mirror and the ts-go Session overlay. rslint selects the mandatory LSP UTF-16 encoding so incoming changes, native diagnostics, plugin diagnostics, and edits share VS Code's coordinate model. Whole-document changes remain a supported protocol fallback.
- **Server-Owned Debounced Re-linting**: `refreshCh` and `debounceCh` collapse bursts of file changes and session refreshes onto the main dispatch loop. The server remains the single owner of the 200 ms typing debounce; open and save diagnostics stay immediate, while save, fix-all, and close discard redundant pending work for their target document.
- **CLI/API Are Mostly Fresh Runs**: CLI and one-shot API requests generally rebuild `Program` state per run; there is no repository-local rule-result cache or persistent incremental lint cache in the main CLI path today. JavaScript API path canonicalization is also scoped to one `lintFiles()` call.
- **Parallel Program Realpath Queries**: when a CLI or lint API request builds multiple Programs concurrently, the loader session derives a Program-only VFS view that coalesces same-path `Realpath` cold queries across those compiler hosts. Completed realpath and empty-result values remain request-local. Exact path strings are used as keys without cleaning, separator rewriting, case folding, or realpath-key merging. Existence checks and all other operations continue through the existing VFS stack. Serial Program builds, config discovery, target binding, LSP, and `cmd/tsgo` retain their existing filesystem paths and cache lifecycles.
- **Resolver-Scoped Effective Config Plans**: each `FileConfigResolver` owns
  exact-path file entries and exact matched-entry shape entries. Both caches use
  publish-once values so concurrent readers of the same resolver wait for
  complete initialization. They are discarded with the resolver; CLI fix
  generations, API requests, LSP lint passes, and config reloads never share a
  plan cache. Merged config maps and configured-rule slices returned by the
  resolver are immutable shared state.
- **Program Loader Generation Boundary**: the initial CLI/API observation and each changed autofix observation own a `program/loader.Session`, created only after that generation's overlay/canonical VFS wrappers are complete. A session privately composes compiler construction, source snapshots, metadata caches, project loading, and target binding for one immutable filesystem view. It is never global, never shared across API requests, and is not used by LSP, whose project session has a different invalidation model.
- **Run/Request-Scoped Program Metadata**: successful `package.json` reads and explicitly registered root, project-reference, and extended tsconfig reads are snapshotted for the context lifetime. Keys remain exact caller paths—no cleaning, case folding, resolving real paths, or symlink merging—and failed reads are retried. Per-key read single-flight avoids duplicate concurrent I/O; generation swaps make future VFS writes safe without clearing a live map. Arbitrary JSON and non-metadata reads bypass this layer.
- **Extended Config Parse Reuse**: the context implements ts-go's `ExtendedConfigCache` shim contract and shares common `extends` parse results across Programs. Parsing occurs outside map locks and publishes with `LoadOrStore`, avoiding recursive cross-cycle lock ordering; rare concurrent misses may duplicate parsing but share the winning immutable result and still single-flight raw bytes. Root `ParsedCommandLine` values and parsed `package.json` objects are not cached.
- **Generation-Scoped Source Snapshots**: Programs built for one CLI/API observation share immutable source text/hash snapshots across project and transient root-parser hosts. Keys are the exact compiler-host source names, never real paths, so lexical, overlay, and symlink aliases remain distinct. Concurrent misses for one key share the successful read/hash operation; failed reads are shared only by the overlapping callers and are not retained. The next changed autofix observation receives a fresh session bound to its new overlay, so no source snapshot crosses generation boundaries and no bound AST is published across rslint Programs.
- **Core-Owned Fix Generations**: after each successful fix round, `internal/linter` advances only its private source memory. CLI/API generation providers project that snapshot onto a fresh request-local overlay and rebuild Programs without mutating the base medium; CLI alone commits the final net delta to disk once. LSP speculative generations remain isolated from its document/session state, which still advances only through versioned document events.
- **Generation-Scoped Parse Reuse**: all Programs within one CLI/API observation share that loader session's content-keyed AST parse cache. Concurrent misses for the same full parse key are single-flight, so multi-Program construction does not duplicate parsing. A changed autofix observation starts a fresh cache for its new source generation; caches are never repository-persistent or shared across lint requests.
- **Bounded Multi-Pass Fixing**: CLI `--fix`, API `fix: true`, and LSP `fixAll` share the linter-owned limit of ten writable rounds. CLI and API request a final observation after the tenth write so their diagnostics describe the returned source; LSP consumes only the net text delta and does not pay for that extra observation.

### Memory Management

- **ts-go Owns the Heavy Graphs**: AST nodes, checker state, compiler project graphs, and session state are primarily owned by ts-go; an rslint Program adds a small immutable facade and generation cache, while listeners, diagnostics, lint selection, and configuration stay outside that object
- **Short-Lived Per-File Structures**: comment stores, disable managers, and rule contexts are allocated per file and dropped after traversal. A comment slice is allocated only if requested
- **Bounded Listener Retention**: a listener registry lives only for one checker-shard task. After each file it clears every function slot before shortening the slices, so backing capacity can be reused without retaining closures, source files, checker state, or rule contexts. The registry is dropped when that task completes and is never pooled across runs or LSP requests
- **Source Snapshot Ownership**: snapshot entries hold an immutable source string plus its 128-bit hash without explicitly copying source bytes; on an AST miss, that string is passed directly to the parser. After generation replacement, a retained unchanged AST may still hold the prior equal string while the fresh snapshot owns the new read. Replaced generations are reclaimed after any in-flight lookup releases them. AST retention and source-generation retention remain deliberately separate lifecycles.
- **Metadata Snapshot Ownership**: metadata strings and extended-config parse entries live only for one loader session. The cache stores successful reads only, and its scope bounds growth to metadata touched by one CLI invocation or API request; no metadata entry survives into another request or the LSP session.
- **Fix Application Uses Linear Rebuilds**: `ApplyRuleFixes` sorts fixes, skips overlapping edits, and rebuilds the output with `strings.Builder` rather than mutating source buffers in place
- **Bounded Queues**: CLI diagnostics use a buffered channel of 4096 items; LSP request/outgoing queues are buffered to 100, and debounce/refresh signals are single-slot channels
- **No Repo-Local Pooling Layer Today**: there is no explicit `sync.Pool`-based object pooling strategy in the main lint path at the moment
- **Fresh ESM Entry Lifetime**: fresh JS/TS config loads use a unique entry-module URL so rewritten bytes and module side effects are evaluated per transaction. Node retains those ESM module namespaces for the process lifetime, so a long-lived native API process can grow this cache slowly across repeated lint requests; static transitive imports continue to use Node's ordinary cache. Bounding this without weakening freshness requires a disposable evaluator realm or worker and remains a future optimization.
- **Garbage Collection Handles Cycles**: the repository does not implement custom cycle breaking for AST/checker graphs; lifecycle cleanup relies on Go GC and on dropping references after each run

## 11. Extensibility & Future Directions

### Plugin Architecture

Plugin execution has two paths:

- bundled Go plugins execute through the shared listener traversal
- third-party ESLint plugin objects are loaded from JS/TS config on the Node side; Go derives catalog placeholders for their rules and sends per-file batches back to the Node plugin worker over reverse IPC

JSON config supports only bundled plugin names because it cannot represent live JavaScript plugin objects. The repository currently ships Go implementations for TypeScript ESLint, Import, Jest, JSX accessibility, Promise, React, React Hooks, Rstest, and Unicorn rule namespaces.

### Rule Extension Points

- **Core Rules**: add a package under `internal/rules/<rule_name>/` and append the rule var to `internal/rules/all.go`'s core rule slice
- **Go Plugin Rules**: add a package under `internal/plugins/<plugin>/rules/<rule_name>/` and append the rule var to that plugin's `all.go`; a new bundled plugin also needs one aggregation entry in `internal/rules/all.go` and, when JSON config may name it, its declaration aliases in `internal/config/plugin_declarations.go`
- **Third-Party Plugin Rules**: import a plugin object in JS/TS config and mount it under an object-form `plugins` prefix; no Go rule implementation or bundled aggregation entry is required
- **Rule Options**: each rule receives parsed options through `Run(ctx, options)`
- **Custom Listener Shapes**: rules can listen on standard kinds and synthetic pattern/exit kinds

### Integration Points

- **Language Server**: `internal/lsp` exposes diagnostics and code actions
- **JavaScript API**: `packages/rslint` talks to the `internal/api/server` handler composed by `cmd/rslint --api` through the versioned `3.0.0` protocol; the handshake negotiates reverse `pluginLint` support before third-party rules run
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
- **Shared Go Test Infrastructure**: `internal/testutil`; package-specific helpers remain beside the owning tests
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

- **Small Go Inputs**: remain inline or table-driven in the owning `*_test.go` file
- **Filesystem Fixtures**: live under the nearest package's `testdata/` directory; related portable text trees may be grouped in `.txtar` archives
- **Txtar Materialization**: `internal/testutil/txtarfs` validates portable relative paths and extracts each selected tree into a fresh `t.TempDir`; it intentionally does not model permissions, symlinks, or other OS metadata
- **Imperative OS Scenarios**: permissions, symlinks, concurrency, and platform-specific behavior stay in Go so their setup and assertions remain explicit
- **Snapshots**: used in several JS and Rust integration tests
- **Virtual Configs / VFS Inputs**: used heavily for API, CLI, and type-checking scenarios

Txtar fixtures are ordinary committed test inputs: they require no generator,
golden-update mode, or separate build step, and run through the owning package's
normal `go test` invocation.

### Continuous Integration

- **Go Tests**: `pnpm run test:go`
- **TypeScript / JS Tests**: `pnpm run test`
- **Linting**: `pnpm run lint` and `pnpm run lint:go`
- **Build Verification**: `pnpm run build`

### Release Lines

- **rslint packages**: `scripts/version.mjs` and `.github/workflows/release.yml` own the unified rslint npm version. They explicitly exclude the tsgo server packages.
- **tsgo distribution**: `pnpm version:crates` updates `tsgo-client`, `Cargo.lock`, `@rslint/tsgo-server`, and all six platform packages to one version. `.github/workflows/release-crates.yml` assembles the platform packages and publishes npm and crates.io artifacts together.

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
│ Configuration / Target Planning / Program Loading / IPC   │  ← internal/config/, internal/config/target/,
│ / API Server / Inspector                                  │     internal/program/loader/, internal/api/,
│                                                           │     internal/api/server/, internal/inspector/
├───────────────────────────────────────────────────────────┤
│ Linter Core / Rule Implementations / Catalog Assembly     │  ← internal/linter/, internal/rules/,
│                                                           │     internal/plugins/
├───────────────────────────────────────────────────────────┤
│ Rule Framework                                            │  ← internal/rule/
├───────────────────────────────────────────────────────────┤
│ Unified Source Program / TS Utilities                     │  ← internal/program/, internal/utils/
├───────────────────────────────────────────────────────────┤
│ TypeScript-Go Bridge                                      │  ← shim/, tools/
├───────────────────────────────────────────────────────────┤
│ typescript-go                                             │  ← typescript-go/
└───────────────────────────────────────────────────────────┘
```

### Dependency Rules

- **Upward Dependencies**: Lower layers never depend on upper layers
- **Rule Isolation**: Individual rules depend on the rule framework and unified Program facade, never command/config assembly or private Program adapters
- **TypeScript Boundary**: All TypeScript integration goes through typescript-go
- **No Circular Dependencies**: Enforced by Go module system

### Lint Path Ownership

The production lint path is a one-way handoff:

```
Config / discovery
  → frozen targets and config ownership
  → Program generation and exact target binding
  → prepared LintPlan
  → RunPipeline
  → integration projection and optional terminal commit
```

Downstream stages may validate or project an upstream decision, but do not
repeat it under a second source of truth.

- **Targets and config**: target selection and config ownership are frozen before Program binding (`target.Plan` for CLI/API and a document snapshot for LSP). Later stages do not add lint targets, rediscover configs, or reassign owners
- **Program generation**: each published `program.Program` is one logically immutable source, module-resolution, filesystem, and optional-checker generation. CLI/API build it through `internal/program/loader`; LSP adapts its session or isolated overlay without exposing the private backend
- **Lint plan**: `PrepareLintPlan` accepts only files already bound to those Programs and freezes each file's rules, shared environment, and checker eligibility. Execution does not scan Program roots or resolve config and rules again
- **Pipeline**: `RunPipeline` is the production orchestration boundary. CLI, API, and LSP choose a complete request and provide generation, plugin transport, presentation, or commit adapters; raw preparation, native lint, plugin dispatch, and fix stages are not product integration APIs
- **Autofix**: fix rounds advance only pipeline-owned in-memory snapshots. Integrations receive the final in-memory delta for the operation, and optional persistence is one terminal commit rather than a series of intermediate writes
- **Rules**: `RuleContext` exposes the bound Program and only the checker granted to that file. Shared structures such as module graphs derive from the Program generation rather than becoming a second authority

Playground inspection is a separate read-only path through `internal/inspector`.

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
│            │                 Loader: Effective-catalog Projects              │
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
│  Run-scoped Program Loader Session                                           │
│  (active projects + governing-config binding + source-only roots)            │
│            │                                                                 │
│            ▼                                                                 │
│  Unified Program Sequence + Target / Path Projections                        │
│            │                                                                 │
│            ▼                                                                 │
│  Match Config Shape -> Reuse/Merge Immutable Config and Enabled Rules        │
│            │                                                                 │
│            ▼                                                                 │
│  RunPipeline: Prepared Lint Plan (Exact Targets + Rules + Checker Grant)     │
│            │                                                                 │
│            ▼                                                                 │
│  Run Rule Initializers -> Register Listeners                                 │
│            │                                                                 │
│            ▼                                                                 │
│  Single DFS AST Traversal -> Listener Dispatch                               │
│            │                                                                 │
│            ▼                                                                 │
│  Native Edit Demand -> Report / Suppress -> Deferred Edit Materialization    │
│            │                                                                 │
│            ▼                                                                 │
│  RuleDiagnostic / Requested Fix / Requested Suggestion Collection            │
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
│     ├── CoreResolver (nearest local core per open document)                  │
│     ├── RuntimeManager (workspace URI + physical core identity)              │
│     ├── WorkspaceDocumentRouter (explicit document ownership)                │
│     └── Rslint runtime per selected installation                             │
│             └──────── rslint/configRefresh ────────────────┐                 │
│             ◄──────── load / activate / commit / abort     │                 │
│                                                            ▼                 │
│                                  local core's cmd/rslint --lsp per runtime   │
│     │                                                                        │
│     ▼                                                                        │
│  internal/lsp + ts-go project.Session                                        │
│     │                                                                        │
│     ▼                                                                        │
│  Frozen Document Snapshot + Session / Overlay Program Generation             │
│     │                                                                        │
│     ▼                                                                        │
│  RunPipeline (Progressive Diagnostics or In-Memory Fix All)                  │
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
│  wasm cmd/rslint --api -> internal/api/server                                │
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
- **ConfiguredRule**: Rule execution descriptor containing implementation, severity, options, type-info requirement, and a shared immutable `RuleEnvironment` reference
- **Diagnostic**: A lint finding reported by a rule or by TypeScript semantic diagnostics
- **Flat Config**: ESLint-style array-based configuration model used by rslint to merge rule settings per file
- **Lint Target Plan**: Immutable `internal/config/target.Plan` containing the selected files, literal-file outcomes, active config owners, and the read-only path-space generation used for later config evaluation
- **Lint Plan**: Immutable `internal/linter.LintPlan` bound to an exact ordered Program/target projection, with each eligible file's resolved rules and checker policy frozen for native execution and optional plugin dispatch
- **Inspector**: Auxiliary backend path that returns node, type, symbol, signature, and flow information for Playground inspection
- **IPC API**: Length-prefixed JSON message protocol exposed by `cmd/rslint --api` for Node and WASM clients; config path resolution and third-party plugin routing have separate identities
- **Listener**: Callback registered by a rule for an AST kind or synthetic listener kind
- **Nearest Config**: In multi-config mode, the governing config selected by lexical-first ownership resolution
- **Owner Index**: Read-only `internal/config/target.OwnerIndex` that routes a frozen file identity to one already-loaded config owner without discovering or parsing config files
- **Node Kind**: Enumerated AST kind value used by ts-go and by the listener dispatcher to identify node types
- **Module Graph**: Program-derived index of module references and their resolved target files; it is cached by source generation and never stored as an independent `RuleContext` authority
- **Program (rslint)**: Immutable generation-scoped facade in `internal/program`; the sole source, filesystem, syntax, module-resolution, cache, and optional checker authority visible to linter and rules
- **Program (ts-go)**: Compiler/project object used by configured or inferred projects and compatibility assembly; it may be privately adapted by an rslint Program but is never exposed through `RuleContext`
- **Program Loader Session**: Request-scoped `internal/program/loader` owner that builds configured projects, binds target-plan files, creates source-only generations when needed, and returns one unified Program sequence to CLI/API
- **Path-Space Snapshot**: Immutable `internal/config.PathSpaceSnapshot` of the lexical and physical authored bases used by one target/config evaluation generation
- **Project Set**: Loader-private, stable set of configured ts-go project generations keyed by normalized declared tsconfig path together with governing-config declaration order
- **project.Session**: ts-go project manager used by LSP for inferred/configured projects and overlays
- **Rule Context**: Runtime environment through which a rule reads file/program/checker state and reports findings
- **Rule Environment**: Immutable settings, language options, and configured globals shared by every rule descriptor from one resolved file-config shape
- **RuleFix**: Text edit represented as a range plus replacement text; fixes are merged and applied after diagnostics are collected
- **Rule Catalog**: Immutable name-indexed snapshot of rule implementations. `rules.All()` from `internal/rules` supplies the shared Go implementation base, and `rule.Catalog.ForESLintPlugins()` derives a snapshot for one exact object-form ESLint-plugin set; config resolvers receive the appropriate snapshot explicitly
- **RuleSuggestion**: Suggested edit attached to a diagnostic that is surfaced to the user but not treated as a default autofix
- **Severity**: Effective diagnostic level for a configured rule, such as `off`, `warn`, or `error`
- **Source Code Fixer**: Fix-application layer that merges non-overlapping `RuleFix` edits and rewrites file contents
- **Synthetic Listener Kind**: rslint-defined pseudo-kind such as `OnExit`, `OnAllowPattern`, or `OnNotAllowPattern` used to distinguish traversal contexts beyond raw AST kinds
- **TypeChecker**: ts-go semantic engine delivered only when Program capability and per-file lint policy both permit semantic rule execution
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
- [ ] Detail integration test coverage and automation
- [ ] Clarify performance regression testing approach

---

This architecture document should be updated as the project evolves. For questions or clarifications, please refer to the source code or open an issue on the [GitHub repository](https://github.com/web-infra-dev/rslint).
