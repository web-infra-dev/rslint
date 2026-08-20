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
│  │  │  internal/program/projectselection   │  │  internal/lsp            │  │  │
│  │  │  internal/program/loader             │  │                          │  │  │
│  │  │  internal/program                    │  │  internal/inspector      │  │  │
│  │  │  internal/linter / rule / utils      │  │                          │  │  │
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

| Path                                 | Purpose                                                                                                                                                       | Key Relationships                                                                                                                                                                                                                                                                                                                                           |
| ------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `website/`                           | Documentation site and Playground UI                                                                                                                          | Uses `packages/rslint-wasm` to run browser linting and `packages/rslint-api` to decode encoded source files; Playground lint requests ultimately reach `internal/linter`, and inspect requests reach `internal/inspector` through `internal/api`                                                                                                            |
| `cmd/rslint/`                        | Main Go binary entry point with CLI, API, and LSP modes                                                                                                       | Orchestrates config loading, `internal/program/loader`, lint execution, fixes, and output without selecting a Program backend. JSON config starts at `internal/config`; `--api` is consumed by `packages/rslint` and `packages/rslint-wasm`; `--lsp` is consumed by `packages/vscode-extension`                                                             |
| `internal/output/`                   | Report model, summary, colors, and stdout formatters                                                                                                          | Consumes final sorted `internal/rule` diagnostics and renders `default`, `jsonline`, `github`, or `gitlab`; the CLI is its current consumer, while the package remains available to other repository integrations that need the same output behavior                                                                                                        |
| `cmd/tsgo/`                          | ts-go semantic inspection/export tool                                                                                                                         | Talks directly to `typescript-go` and bypasses the lint framework; consumed by `packages/tsgo` and `crates/tsgo-client`                                                                                                                                                                                                                                     |
| `internal/api/`                      | stdio IPC protocol and service types for JS/WASM integration                                                                                                  | Shared protocol layer for `cmd/rslint --api`; used by `packages/rslint`, `packages/rslint-wasm`, `internal/linter`, and `internal/inspector`                                                                                                                                                                                                                |
| `internal/config/`                   | Configuration models, JSON loading, matching/merging, runtime ownership resolution, lint-target/effective-project planning, and centralized rule registration | Owns target discovery and freezes one effective config plus one ordered lexical project-candidate list per target. Owner-wide project declarations are validated separately from target matching. `RegisterAllRules()` orchestrates rule registration; `rule_registry.go` resolves enabled `internal/rule` descriptors for every entry point                |
| `internal/config/discovery/`         | Go-owned JS/TS config candidate discovery and immutable catalog construction                                                                                  | Imports the parent config model/matching policy, batches exact candidates to a host-supplied Node loader, and returns configs/scopes/failures/effective IDs. CLI, API, and LSP call `DiscoverAutomatic` or `LoadExplicitConfig`; the parent package never imports this child package                                                                        |
| `internal/config/gitignore/`         | Config-scoped `.gitignore` parsing, directory reachability, and pattern projection                                                                            | Staged JS/TS catalogs carry a filesystem-independent cursor through their existing walk, pruning Git-inaccessible subtrees and freezing observed patterns for lint-target admission; JSON/JSONC and low-level fallback paths reuse the direct collector                                                                                                     |
| `internal/inspector/`                | AST/type/symbol/signature/flow inspection for Playground                                                                                                      | Auxiliary backend used mainly by website Playground inspect panels; builds rich semantic data from `typescript-go` programs                                                                                                                                                                                                                                 |
| `internal/program/`                  | Immutable rslint Program facade and source-generation indexes                                                                                                 | Privately adapts project/compiler and root-parser hosts into one source, filesystem, syntax, optional checker, module-resolution, and module-reference contract. `Program.ModuleGraph()` and generation-scoped caches derive from that same authority; adapter identity is not observable by linter or rules                                                |
| `internal/program/projectselection/` | Entry-neutral configured-project ownership policy                                                                                                             | Consumes target-specific ordered project IDs and provider callbacks. It alone implements all-direct-before-import selection, declaration order, extension eligibility, complete bindings, and the logical error frontier; it owns no filesystem, config, Program, watcher, or cache lifetime                                                                |
| `internal/program/loader/`           | Run-scoped project provider, Program construction, and final source projection for CLI/API                                                                    | Interns exact lexical tsconfig declarations, parses/builds each stage once per request generation, supplies evidence to the shared selector, and projects its complete binding into configured/source-only rslint Programs. Prefetch changes scheduling only; LSP keeps its resident provider instead of using this request-scoped lifetime                 |
| `internal/linter/`                   | Core lint engine, traversal, and fix application                                                                                                              | Consumes unified `internal/program.Program` inputs, rules from `internal/rule`, file config from `internal/config`, and optional ts-go `TypeChecker` data; also serves `internal/api` and `internal/lsp`                                                                                                                                                    |
| `internal/lsp/`                      | Language Server Protocol implementation                                                                                                                       | Wraps `typescript-go project.Session`, owns transactional config discovery, last-good state, watchers, overlays, and resident Programs, while adapting those Programs to the same effective-config planner and project selector used by CLI/API                                                                                                             |
| `internal/rule/`                     | Rule framework, configured-rule descriptors, context, diagnostics, fixes, and disable manager                                                                 | Shared foundation for config resolution, core rules, and plugin rules; `internal/linter` consumes its immutable rule environments, listeners, reporting APIs, and Program-derived cache helper                                                                                                                                                              |
| `internal/rule_tester/`              | Go-side rule testing helpers                                                                                                                                  | Supports rule development and complements JS-side testers in `packages/rule-tester` and `packages/rslint-test-tools`                                                                                                                                                                                                                                        |
| `internal/rules/`                    | Core lint rule implementations without plugin namespace; `all.go` aggregates them into the `GetAllRules()` slice                                              | `internal/config/config.go`'s `RegisterAllRules()` consumes the slice and registers each rule; then executed by `internal/linter` like plugin rules                                                                                                                                                                                                         |
| `internal/plugins/typescript/`       | `@typescript-eslint`-style rules                                                                                                                              | Registered into the shared rule registry by `RegisterAllRules()` and often rely on `TypeChecker` from `typescript-go`                                                                                                                                                                                                                                       |
| `internal/plugins/react/`            | React rule implementations                                                                                                                                    | Registered into the shared rule registry by `RegisterAllRules()` and executed through the same listener pipeline in `internal/linter`                                                                                                                                                                                                                       |
| `internal/plugins/jest/`             | Jest rule implementations                                                                                                                                     | Registered into the shared rule registry by `RegisterAllRules()` and executed through the same listener pipeline in `internal/linter`                                                                                                                                                                                                                       |
| `internal/plugins/import/`           | Import plugin registration and rules                                                                                                                          | Contributes plugin rules through `RegisterAllRules()` and participates in normal config-driven linting                                                                                                                                                                                                                                                      |
| `internal/utils/`                    | Shared utilities for JSONC, compiler hosts, invocation-scoped compiler construction, overlay VFS, and AST/type helpers                                        | Provides the low-level compiler-construction and snapshot services privately composed by `internal/program/loader`; command entry points do not coordinate compiler hosts or source generations directly. LSP, rule tests, and auxiliary entry points reuse lower-level compiler helpers. Source-generation module resolution belongs to `internal/program` |
| `packages/rslint/`                   | Main npm package with JavaScript API and CLI wrapper                                                                                                          | Spawns `cmd/rslint --api` in JavaScript runtime environments and uses `internal/api` message shapes                                                                                                                                                                                                                                                         |
| `packages/rslint-api/`               | Frontend-facing encoded source file / AST decoding helpers                                                                                                    | Used mainly by website Playground to decode AST/source data returned from the Go API                                                                                                                                                                                                                                                                        |
| `packages/rslint-test-tools/`        | Testing utilities and cross-ecosystem rule tests                                                                                                              | Supports package-side and integration-style tests around the linter and rule ecosystem                                                                                                                                                                                                                                                                      |
| `packages/rslint-wasm/`              | Browser/WASM package for running `rslint --api` in a worker                                                                                                   | Starts the browser worker, hosts the wasm runtime, and bridges website Playground requests to `internal/api`, `internal/linter`, and `internal/inspector`                                                                                                                                                                                                   |
| `packages/rule-tester/`              | Forked `@typescript-eslint/rule-tester` package used in tests                                                                                                 | JS-side rule testing support that complements Go-side helpers                                                                                                                                                                                                                                                                                               |
| `packages/utils/`                    | Shared JavaScript utilities                                                                                                                                   | Shared support package for the JS/website tooling layer                                                                                                                                                                                                                                                                                                     |
| `packages/vscode-extension/`         | VS Code extension for IDE integration                                                                                                                         | Resolves the nearest project-local `@rslint/core` per open document, launches that installation's `cmd/rslint --lsp`, serves reverse config/plugin requests, and routes diagnostics/code actions to the document's selected runtime                                                                                                                         |
| `packages/tsgo/`                     | `@rslint/tsgo-server` JS wrapper package for the `tsgo` tool                                                                                                  | JavaScript-facing wrapper around `cmd/tsgo` output; resolves the matching `@rslint/tsgo-server-<platform>-<arch>` binary package                                                                                                                                                                                                                            |
| `typescript-go/`                     | Git submodule containing TypeScript compiler Go port                                                                                                          | Provides parser, AST, checker, `Program`, `project.Session`, diagnostics, scanner, and VFS primitives used throughout the backend                                                                                                                                                                                                                           |
| `shim/`                              | Generated bridge packages exposing ts-go internals                                                                                                            | Bridge layer between repository Go code and `typescript-go` internals; generated and updated by `tools/`                                                                                                                                                                                                                                                    |
| `tools/`                             | Shim generator and ts-go update scripts                                                                                                                       | Generates `shim/` code and maintains the pinned `typescript-go` integration                                                                                                                                                                                                                                                                                 |
| `crates/tsgo-client/`                | Rust client for communicating with `cmd/tsgo`                                                                                                                 | Spawns `cmd/tsgo` and consumes its semantic/project output from Rust                                                                                                                                                                                                                                                                                        |

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

1. **Source and Metadata Loading**: Files come from the real filesystem, an overlay VFS, or LSP document overlays. Each CLI run or API request creates one `program/loader.Session` around its final VFS view. The session's private compiler hosts keep source snapshots keyed by the exact normalized source path, storing the first successful text read and its xxh3 hash for the current generation. The same request scope snapshots successful `package.json` and explicitly registered tsconfig reads by the exact requested path; other JSON, source, ignore, and config-discovery reads remain uncached. When multiple Programs are constructed concurrently, a derived context view coalesces their concurrent cold realpath queries for the same exact path while sharing all parent caches. CLI fix writes replace the source generation before rebuilding Programs; API snapshots end with the request. LSP follows document events and its own session/versioned cache lifecycle instead.
2. **Effective Config and Project Selection**: `internal/config` first freezes each discovered target's governing config, exact matched-entry shape, merged rules/language settings, and ordered lexical `parserOptions.project` candidates. The active owner's complete authored project catalog is path/glob-validated once, but an unmatched entry never becomes a target candidate. `internal/program/projectselection` then applies one request-wide policy for CLI, API, normal LSP diagnostics, and speculative LSP fixes: finish direct-root selection for every target, validate all selected direct Programs, and only then try declaration-order import membership for unresolved targets whose compiler options support the extension. It returns a complete direct/import/none binding. CLI/API's request provider and LSP's resident provider supply the metadata and Program facts; neither may choose an owner. A recursive CLI scope at or above its nearest candidate tsconfig directory, or spanning multiple nearest candidates, may prefetch the union of its targets' effective candidates for throughput. A directory strictly inside one common nearest candidate tsconfig directory remains demand-driven, as do file-only CLI requests, API requests, and LSP requests. Prefetch and demand hints cannot widen candidates, publish an otherwise-unreached error, or change a binding. `--type-check` and `--type-check-only` independently materialize the per-owner catalog through the same exact lexical project slots; catalog-only projects cannot bind lint targets, and lint-only candidates cannot enter program-wide diagnostics. Targets with no selected project become source-only rslint Programs with parser, binder, package/module metadata, and direct-import resolution but no recursively loaded project graph or type-check capability. Each CLI fix pass creates a new Program generation and reruns the immutable selection plan. LSP keeps Session-owned Programs, watcher coverage, overlay invalidation, and request-isolated fix Programs while sharing only the effective plan and selector. Lexical tsconfig declarations are never realpath-merged because their relative include/extends bases are authored semantics; exact/canonical identity is used only for target/root/source aliases.
3. **Lexical + Syntax Parsing**: ts-go tokenizes and parses source files into TypeScript-native AST nodes. Source-only roots additionally run the ts-go binder so syntax-only rules retain symbols and lexical scopes before their rslint Program is published.
4. **Semantic Analysis**: The lint plan reads each Program's per-file checker capability once, freezes that result with the file's configured rules, and acquires a checker only for eligible files. LSP can additionally narrow rule eligibility for a request before planning. Program capability, configured-rule eligibility, actual checker delivery, and program-wide diagnostics remain separate decisions.
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
Config resolution normalizes the per-file `ecmaVersion` into `LanguageOptions`;
its zero value means the moving `latest` edition. The linter exposes the
normalized value through `RuleContext.LanguageOptions` and also uses it to
build one `Globals` value for each native rule context. Rules read
`LanguageOptions` when upstream behavior depends on language configuration;
they use `Globals` for variable-availability decisions. `Globals` owns the
ESLint-versioned language-global set, resolved language defaults, the authored
`languageOptions.globals` source, inline `/* global */` settings and ranges,
and the effective access after applying their precedence. Rules use
`Globals.Access` for standard language-global decisions and its narrower source
accessors only when upstream behavior depends on provenance, instead of
rebuilding the merge. A rule whose upstream semantics add another source, such
as TypeScript library globals, applies this view last so `ecmaVersion` and
authored overrides remain authoritative. This keeps the language-global
composition point extensible for future language options such as `sourceType`;
non-global wrapper bindings remain a `RefStore` initialization concern.

Before constructing rule contexts, the linter calls `ResolveLanguageDefaults`
once and passes its concrete `GlobalsInit` and `RefStoreInit` results to their
respective consumers. Today the resolver selects defaults from the source
filename: `.js` and `.mjs` contribute a non-global top-level scope; `.cjs`
additionally contributes writable `exports`, read-only `global`, `module`, and
`require`, plus the wrapper-local `arguments` binding. Other extensions
contribute no defaults. The resolver does not inspect `package.json` or authored
`sourceType`; adding `sourceType` later changes the resolver input and this one
call site, while `Globals`, `RefStore`, and rules keep consuming the same
initialization types.
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

1. `RegisterAllRules()` in `internal/config/config.go` populates the global registry once
2. config merge selects enabled rules for a file
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

Important behavior differences by integration:

- **CLI lint-only**: requests diagnostics only, so migrated native rules do not
  construct autofixes or suggestions
- **CLI `--fix`**: requests native autofixes, can rerun lint and fix for multiple
  passes, and uses diagnostics-only mode for the final no-more-writes
  verification pass
- **LSP quick fix**: returns direct text edits for one diagnostic
- **LSP fix-all**: runs repeated lint-fix cycles, then returns one whole-document replacement edit
- **LSP normal diagnostics and API**: request all native edits because fixes and
  suggestions are response metadata even when they are not immediately applied
- **LSP speculative fix-all passes**: request native autofixes only
- **API**: `lint({ fix: true })` applies fixes in a single pass and returns the fixed source per file in `output` (the JS side persists it via `Rslint.outputFixes`). There is no separate `applyFixes`, and—unlike the CLI—it does not re-lint across passes.

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

**Key difference**: JSON configs are normalized by `normalizeJSONConfig()`, which auto-enables core rules and rules from declared plugins unless explicitly overridden. JS/TS configs only enable what the normalized config entries specify, usually via presets.

### Config Entry Structure

Each entry in the config array supports:

| Field             | Type                                       | Description                                                                          |
| ----------------- | ------------------------------------------ | ------------------------------------------------------------------------------------ |
| `files`           | `(string \| string[])[]`                   | Non-empty selector list; top-level selectors are ORed and nested selectors are ANDed |
| `ignores`         | `string[]`                                 | Glob patterns excluded by this entry                                                 |
| `languageOptions` | `object`                                   | `ecmaVersion`, globals, and parser options such as project settings                  |
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
  every refresh; Go loads exactly that module while retaining the process cwd
  as its invocation-wide matching root. The client owns change notifications
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

An explicit JS/TS `--config` or API `overrideConfigFile` bypasses automatic
candidate selection and loads the exact module. The invocation cwd remains its
matching directory. The exact config path is never gated by `.gitignore`;
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

`FileConfigResolver` applies that policy in two phases: it first matches a file
to the exact ordered set of contributing entries, then compiles the set into an
immutable `EffectiveFileConfig`. Rules/plugin maps and the effective
`parserOptions.project` declaration are read from that same object; integrations
must not match the file a second time after Program binding. Files with the same
set share one plan for that resolver's lifetime. Shape identity is collision-free (`uint64` for the
first 64 entries plus the complete remaining bitset), resolver-local, and
independent of path identity. File-cache keys remain the exact caller strings;
the resolver does not clean, case-fold, canonicalize, or merge symlink aliases.
The direct `GetConfigForFile()` compatibility path uses the same match/merge
primitives without retaining a cache.

`ProjectPathResolver` is the config-generation companion to that resolver. It
validates every authored `parserOptions.project` literal/glob for each active
owner in original entry order, then projects only the final declaration from a
target's `EffectiveFileConfig` into its candidate list. An unspecified project
and an explicitly empty project keep the current compatibility behavior of
probing that owner's `tsconfig.json`; an empty value still clears an earlier
non-empty declaration before the probe. A target with no contributing config
entry receives no project candidates. Type-check catalogs are derived
independently per owner and then combined, so one owner's explicit projects never
disable another owner's default probe.

LSP native lint, background third-party-plugin diagnostics, and every
speculative fix-all pass carry the same `PlannedLintTarget` produced for that
pass. Plugin dispatch reads rules, settings, language options, and the authored
config key from that plan; it never performs another owner lookup or files /
ignores match.

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
  files apply, while parent `.gitignore` files do not. In staged JS/TS catalogs,
  the walk records sources by owner and case-aware source identity, orders them
  parent-before-child, and materializes the synthetic Git entry before
  publishing. Direct automatically reachable child config directories are
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
  worker, LSP commits the ordinary native config with an empty no-host plugin
  state and retries on later refreshes; once a usable snapshot exists, the same
  worker failure aborts and preserves that last-good snapshot. A successful
  no-candidate transaction removes the previous JS catalog and
  exposes the Go-loaded JSON fallback
- native and third-party plugin rules are gated by their normalized prefixes for JS/TS configs; third-party rules use process-wide Go registry placeholders, but LSP additionally filters those placeholders through the exact rule-name set committed for the current Node generation so metadata retained from an older generation cannot be dispatched to a newer worker
- CLI/API lint target selection is independent from TypeScript `Program` membership and considers only rslint-supported script extensions. The `.js`, `.mjs`, `.cjs`, `.jsx`, `.ts`, `.tsx`, `.mts`, and `.cts` default baseline is always selected; explicit config `files` contributes candidates only within the supported set. Global ignores and `.gitignore` remove targets, while an entry-level ignore prevents only its own selector/config contribution
- selected CLI/API targets can still appear as 0-rule lint results when no config entry contributes rules; this applies to default-baseline directory discovery and explicit supported files, and syntax diagnostics remain available in that state
- under automatic discovery, each selected file is governed by its nearest loadable config; an explicitly selected config is invocation-wide. A target can bind only to projects from its own final matched config shape, not the union of every project mentioned by that owner. The first candidate whose parsed root set contains the file wins. Only after every target's direct tier is decided may the first declaration-order Program containing an unresolved target through imports win
- `files`/`ignores` matching uses the stable target path in the governing config's path space; a ts-go Program source alias is used only to locate the AST and type information. Rules, plugin settings, and project candidates therefore stay attached to the same effective config even when Program membership or source spelling changes
- exact lexical tsconfig declarations are deduplicated across targets and owners within one provider generation, while different symlink declarations remain distinct because TypeScript resolves relative paths from the authored location. CLI fix passes create a new source/Program generation and rerun the immutable target-specific plan, so import-only membership is recomputed after source edits. Selected CLI/API files with no configured owner are parsed and bound as independent source-only Programs. Targets whose names collide under a case-insensitive ts-go path key are partitioned across independent Programs so distinct physical files and package scopes remain distinct
- `--type-check` and `--type-check-only` build every real tsconfig in the per-owner catalog. Git reachability may change which automatic configs enter that catalog; once established, program-wide checking is not filtered by lint targets, config `files`/`ignores`, `.gitignore`, or CLI file/directory arguments. The catalog role and lint-candidate role share construction slots but not semantics: catalog-only Programs cannot claim lint targets, lint-only Programs do not produce program-wide diagnostics, and a Program carrying both roles is constructed once. `--type-check-only` skips the separate lint-target walk
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
4. **Lint and Effective-Config Plan**: Go resolves the stable target set from CLI/API scope, the default extension baseline, config `files`, global ignores, and `.gitignore`. The config generation then freezes the exact matched config, enabled-rule source, and ordered project candidates for every target. No later stage reruns `files`/`ignores` or infers config from a Program source path.
5. **Project Provider**: one request-scoped `program/loader.Session` interns exact lexical tsconfig IDs and owns separate parse-metadata and Program-construction stages. A CLI directory at or above its nearest candidate tsconfig directory, or spanning multiple nearest candidates, may ask it to prefetch the target-effective candidate union. A directory strictly below one common candidate tsconfig directory, file-only CLI requests, API requests, and LSP requests stay demand-driven. `--type-check` adds the per-owner catalog as a separate consumer role. These choices affect scheduling and materialization only.
6. **Shared Selection and Projection**: `internal/program/projectselection` binds every target in two request-wide phases: declaration-order direct roots first, then declaration-order import membership for unresolved targets. The loader consumes the complete direct/import/none vector exactly once, creates source-only Programs for none, and returns one ordered Program sequence plus target/config indexes. CLI/API do not reselect project ownership; they only locate the target inside its selected Program. LSP supplies the same selector from its resident project/session adapter.
7. **Rule Plan Preparation**: `PrepareLintPlan()` receives one ordered sequence of rslint Programs and resolves each selected file's rules directly from its frozen `EffectiveFileConfig`, never the ts-go source alias. The immutable result freezes the shared rule environment and per-file type eligibility alongside syntax-error and zero-rule files for native accounting while exposing the non-empty file/rule projection needed by third-party plugin dispatch. CLI and native API reuse the same plan for plugin and native execution; LSP keeps its single-document `LintSingleFile` adapter.
8. **Rule Execution**: `RunLinter()` schedules the unified Program sequence over the prepared plan. It asks only the Program facade to acquire checkers, retaining checker-exclusive grouping when available and sharding sufficiently large checker-free generations across at most `GOMAXPROCS` workers. Small checker-free Programs stay in one shard so scheduling and shared-consumer overhead cannot dominate their rule work. Each Program constructor freezes its complete source universe; target/config filters produce a separate execution projection, while cross-file rules always see the full universe. Prepared plans bind the exact Program pointer, so a reparse/fix generation must construct a new Program and plan. Module graphs and other derived structures are cached by Program generation rather than copied into `RuleContext`. The CLI supplies native edit demand independently of rule selection: diagnostics-only for plain lint, autofix-only for writable `--fix` passes, and diagnostics-only for the final verification pass. When `--type-check` is enabled, Phase 2 schedules only Programs that expose complete program diagnostics; source-only generations expose no such capability and require no parallel skip state.
9. **Result Aggregation**: diagnostics are sent through one run-scoped diagnostics channel and collected at the CLI layer
10. **Fix Passes**: CLI multi-pass `--fix` applies fixes, rebuilds source generations, and rebinds the unchanged target plan after each pass. A file may be owned by a different Program generation after its import graph changes; Program objects are immutable identities and are never refreshed in place
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
   - After target discovery, `ProjectPathResolver` freezes effective config and
     candidate slots with at most `min(GOMAXPROCS, target count)` workers.
     Results and errors are committed in original target order. With
     `--singleThreaded`, it resolves that order directly and stops at the first
     error.
   - The request provider interns exact lexical project identities in stable
     target/candidate order. Confirmed direct winners, type-check catalog
     projects, and optional recursive-directory candidate unions use one bounded
     Program-construction queue with at most `min(GOMAXPROCS, project count)`
     workers. Fully prefetched root and Program source indexes share one
     generation-local `program.PathIdentityResolver`; the selected ProjectSet
     carries that same resolver into final source projection. Physical identity
     work is authoritative, exact-first, and lazily batched across the completed
     candidate set on the first exact miss. A focused request may schedule one
     request-wide proximity hint early only after its metadata proves every
     target direct. Recursive scopes at or above a candidate tsconfig directory,
     or across distinct nearest candidates, instead use the bounded union path. Parse/build
     completion cannot publish an error or owner by itself.
   - The shared selector consumes metadata and Program results in request target
     order and candidate declaration order. It completes every direct decision
     and validation before import fallback, so concurrent completion cannot
     overtake the logical owner/error frontier. `--singleThreaded` uses the same
     selector with synchronous provider stages.
   - `--singleThreaded` executes the same state machine with one Go discovery
     worker and serializes module evaluation within each Node frontier batch.
     Coordinator batches and results remain ordered in either mode.

5. **Lint-target directory walker** (`internal/config/file_discovery.go`)
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
     only canonical identities present in the target plan. CLI fix passes
     create a fresh index when they rebind the target plan.
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
  one resolver for each committed config generation and replaces it when that
  generation changes.

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
  complete initialization. CLI fix passes reuse the invocation's frozen target
  and effective plan while rebuilding Program generations. Each API request
  owns a fresh resolver, while LSP lint passes share the resolver from their
  committed config generation until a config or project refresh replaces it. Merged config
  maps and configured-rule slices returned by a resolver are immutable shared
  state.
- **Program Loader Session Boundary**: one CLI invocation or API lint request owns one `program/loader.Session`, created only after the request's overlay/canonical VFS wrappers are complete. The session privately composes compiler construction, source snapshots, metadata caches, project loading, target binding, and fix-generation invalidation. It is never global, never shared across API requests, and is not used by LSP, whose project session has a different invalidation model.
- **Run/Request-Scoped Program Metadata**: successful `package.json` reads and explicitly registered root, project-reference, and extended tsconfig reads are snapshotted for the context lifetime. Keys remain exact caller paths—no cleaning, case folding, resolving real paths, or symlink merging—and failed reads are retried. Per-key read single-flight avoids duplicate concurrent I/O; generation swaps make future VFS writes safe without clearing a live map. Arbitrary JSON and non-metadata reads bypass this layer.
- **Extended Config Parse Reuse**: the context implements ts-go's `ExtendedConfigCache` shim contract and shares common `extends` parse results across Programs. Parsing occurs outside map locks and publishes with `LoadOrStore`, avoiding recursive cross-cycle lock ordering; rare concurrent misses may duplicate parsing but share the winning immutable result and still single-flight raw bytes. Root `ParsedCommandLine` values and parsed `package.json` objects are not cached.
- **Run/Request-Scoped Source Snapshots**: CLI runs and individual API requests share immutable source text/hash snapshots across project and transient root-parser hosts. Keys are the exact compiler-host source names, never real paths, so lexical, overlay, and symlink aliases remain distinct. Concurrent misses for one key share the successful read/hash operation; failed reads are shared only by the overlapping callers and are not retained. The source layer in one cache binds to one filesystem view across its generations; compiler hosts using another view bypass this layer while retaining content-keyed AST reuse. Snapshot sharing never publishes a bound AST across rslint Program generations.
- **Generation-Based Fix Invalidation**: after every CLI fix write attempt, the compiler host atomically installs an empty source generation before any Program rebuild. Swapping generations rather than clearing a live map prevents an older in-flight read from repopulating the new generation. The API fix path only returns output and does not mutate or rebuild its overlay, while LSP remains version/didChange-driven.
- **Run-Scoped Parse Reuse**: CLI Program rebuilds within one invocation and Programs within one API request share the existing content-keyed AST parse cache. Concurrent misses for the same full parse key are single-flight, so bounded Program construction does not duplicate parsing. Source-generation invalidation does not clear AST entries, so unchanged bytes can reuse their `SourceFile`. The cache is discarded with its run/request and is never repository-persistent or shared across lint requests.
- **Bounded Multi-Pass Fixing**: `--fix` and LSP `fixAll` intentionally rerun lint after applying edits, but cap the cascade at `maxFixPasses = 10`

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
│ Configuration / Program Loading / IPC / Inspector         │  ← internal/config/, internal/program/loader/,
│                                                           │     internal/api/, internal/inspector/
├───────────────────────────────────────────────────────────┤
│ Linter Core / Rule Implementations                        │  ← internal/linter/, internal/rules/,
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

### Key Interfaces

- **Config → Registry**: map each merged config shape into one immutable `RuleEnvironment` plus enabled `ConfiguredRule` descriptors that share it
- **Config → Project Selection**: config owns target discovery, governing ownership, the immutable effective config, and each target's ordered lexical project candidates. It never parses a tsconfig or owns a Program
- **Project Providers → Shared Selector**: CLI/API's request loader and LSP's resident adapter expose only metadata, Program availability, and source-membership facts. The selector alone owns direct/import/none policy, declaration order, and the logical error frontier
- **Program Loader → Integrations**: CLI/API receive one ordered rslint Program sequence plus complete target/config indexes and never reselect ownership. LSP retains its session/watcher lifetime but consumes the same selector binding
- **Programs → Linter**: immutable rslint `program.Program` instances are the only source-universe authority; private adapter identity is unobservable, while per-file checker and program-wide diagnostic availability are queried as capabilities rather than inferred from construction
- **Rules → RuleContext**: rules receive one rslint Program and the checker actually granted to that file. Module graphs and other shared structures derive from Program rather than adding parallel context authority
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
│  Effective Config Plan per Target                                            │
│  (merged rules/settings + ordered lexical project candidates)                │
│            │                                                                 │
│            ▼                                                                 │
│  Shared Project Selector + CLI/API Request Provider                          │
│  (all direct roots -> ordered import fallback -> complete binding)           │
│            │                                                                 │
│            ▼                                                                 │
│  Unified Program Sequence + Target Indexes                                   │
│  (configured owners + source-only gaps; no project reselection)              │
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
│  Committed Effective Config Plan for the document                           │
│     │                                                                        │
│     ▼                                                                        │
│  Shared Project Selector + resident Session/watcher provider                 │
│     │                                                                        │
│     ▼                                                                        │
│  LintSingleFile + plugin input from the same Effective Config                │
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
- **ConfiguredRule**: Rule execution descriptor containing implementation, severity, options, type-info requirement, and a shared immutable `RuleEnvironment` reference
- **Diagnostic**: A lint finding reported by a rule or by TypeScript semantic diagnostics
- **Effective File Config**: Immutable config-generation result for one exact matched-entry shape; rules, plugin settings, language options, and project declaration all come from this single source
- **Flat Config**: ESLint-style array-based configuration model used by rslint to merge rule settings per file
- **Inspector**: Auxiliary backend path that returns node, type, symbol, signature, and flow information for Playground inspection
- **IPC API**: Length-prefixed JSON message protocol exposed by `cmd/rslint --api` for Node and WASM clients; config path resolution and third-party plugin routing use separate keys when API overrides rebase relative patterns
- **Listener**: Callback registered by a rule for an AST kind or synthetic listener kind
- **Nearest Config**: In multi-config mode, the governing config selected by lexical-first ownership resolution
- **Node Kind**: Enumerated AST kind value used by ts-go and by the listener dispatcher to identify node types
- **Module Graph**: Program-derived index of module references and their resolved target files; it is cached by source generation and never stored as an independent `RuleContext` authority
- **Program (rslint)**: Immutable generation-scoped facade in `internal/program`; the sole source, filesystem, syntax, module-resolution, cache, and optional checker authority visible to linter and rules
- **Program (ts-go)**: Compiler/project object used by configured or inferred projects and compatibility assembly; it may be privately adapted by an rslint Program but is never exposed through `RuleContext`
- **Program Loader Session**: Request-scoped `internal/program/loader` provider that parses/builds exact lexical projects once, feeds evidence to the shared selector, projects its complete binding, creates source-only generations when needed, and returns one unified Program sequence to CLI/API
- **Lint Project Plan**: Config-owned request plan containing every discovered target, its effective config, and its own ordered exact-lexical project candidates
- **Project Selection**: Entry-neutral direct/import/none state machine shared by CLI, API, and LSP; providers supply evidence and lifetime but cannot choose ownership
- **Project Set**: Loader-private, stable set of configured ts-go project generations keyed by exact lexical tsconfig declaration path, with independent lint-candidate and type-check-catalog roles
- **project.Session**: ts-go project manager used by LSP for inferred/configured projects and overlays
- **Rule Context**: Runtime environment through which a rule reads file/program/checker state and reports findings
- **Rule Environment**: Immutable settings, language options, and configured globals shared by every rule descriptor from one resolved file-config shape
- **RuleFix**: Text edit represented as a range plus replacement text; fixes are merged and applied after diagnostics are collected
- **Rule Registry**: Shared registry of rule implementations and config-to-rule resolution logic; the registry is implemented in `internal/config/rule_registry.go` and populated by `RegisterAllRules()` in `internal/config/config.go`
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
