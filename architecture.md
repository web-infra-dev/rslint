# Architecture Overview

Rslint is a high-performance JavaScript and TypeScript linter, designed as a drop-in replacement for ESLint and TypeScript-ESLint. It uses the [TypeScript Go compiler](https://github.com/microsoft/TypeScript) to achieve 20-40x speedup over traditional ESLint setups through native parsing, direct TypeScript AST usage, and parallel processing.

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
- **Batteries Included**: Ships with all existing typescript-eslint rules and ESLint rules

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

| Path                                 | Purpose                                                                                                                        | Key Relationships                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| ------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `website/`                           | Documentation site and Playground UI                                                                                           | Uses `packages/rslint-wasm` to run browser linting and `packages/rslint-api` to decode encoded source files; Playground requests pass through `internal/api` and `internal/api/server` before reaching `internal/linter` or `internal/inspector`                                                                                                                                                                                                                |
| `cmd/rslint/`                        | Main Go binary entry point with CLI, API, and LSP modes                                                                        | Owns mode selection and process stdio/exit-code composition. The default CLI prepares config, targets, Program-generation and final disk-projection adapters for `internal/linter.RunPipeline`; `--api` delegates concrete requests to `internal/api/server`; `--lsp` delegates to `internal/lsp`                                                                                                                                                               |
| `internal/output/`                   | Report model, lifecycle/status text, colors, and stdout formatters                                                             | Consumes a completed presentation-only `Report` assembled by the CLI and renders `default`, `jsonline`, `github`, or `gitlab`. Its diagnostics contain only message, location, severity, and optional immutable source-text projections; origin has already been reduced into report counts. The package imports no linter, rule, config, Program, AST, VFS, or compiler-core contracts                                                                         |
| `cmd/tsgo/`                          | ts-go semantic inspection/export tool                                                                                          | Talks directly to `typescript-go` and bypasses the lint framework; consumed by `packages/tsgo` and `crates/tsgo-client`                                                                                                                                                                                                                                                                                                                                         |
| `internal/ipc/`                      | Generic framed requests and text/binary attachment transport                                                                   | Owns request matching, shared-memory publication, acknowledgements and private platform backends. Application message structures and decoding belong to their consumers; IPC imports no linter or plugin implementation. Node uses `packages/rslint/src/ipc`, and Rust provides byte access through `memory_transport`                                                                                                                                          |
| `internal/api/`                      | stdio IPC protocol, wire types, and generic bidirectional service for JS/WASM integration                                      | Defines the stable request/response boundary used by `packages/rslint`, `packages/rslint-wasm`, and `internal/api/server`; it does not import concrete lint, config, Program-loader, or command implementations                                                                                                                                                                                                                                                 |
| `internal/api/server/`               | Concrete rslint API request handlers                                                                                           | Implements lint and AST-inspection requests over `internal/api`, owns API-specific paths, request-overlay generation adapters, reverse config adapters, and structured response projection; it delegates lint planning, execution, and in-memory fix application to `internal/linter`                                                                                                                                                                           |
| `internal/api/plugin/`               | Reverse plugin RPC adapter shared by CLI and API hosts                                                                         | Connects the linter batch interface to API requests and channel-owned output bytes. CLI and API handlers compose it directly; language projection and generic IPC import none of its plugin policy.                                                                                                                                                                                                                                                             |
| `internal/config/`                   | Configuration models, authored path-space snapshots, matching/merging, project-path resolution, and configured-rule evaluation | Owns the shared extension/default-exclude policy, the `GlobalIgnoreMatcher` consumed by config-candidate discovery, and the `TargetMatcher` consumed by lint-target planning; both matchers share the private config-target resolver. File resolvers receive an immutable `internal/rule.Catalog` explicitly and convert merged rule settings into execution descriptors without owning target walking, module loading, or config ownership                     |
| `internal/config/lint/`              | Effective config and rule resolution for already-selected lint targets                                                         | Joins the immutable target/source binding returned by `internal/program/loader` to the owning config's `FileConfigResolver`. CLI and API use the same resolver; it cannot discover targets, infer a different owner, construct Programs, or execute rules                                                                                                                                                                                                       |
| `internal/config/target/`            | Lint-target planning, explicit-file outcomes, directory walking, canonical deduplication, and config-owner routing             | Imports the parent config model and matching policy through `PathIdentity`, `PathSpaceSnapshot`, and `TargetMatcher`. It also owns exact-first source-path lookup in immutable source-to-target maps. CLI, API, LSP, `internal/config/lint`, and `internal/program/loader` consume `File`, `Plan`, `OwnerIndex`, or that lookup; the parent config package never imports this child package                                                                     |
| `internal/config/discovery/`         | Go-owned JS/TS config candidate discovery and immutable catalog construction                                                   | Imports the parent config model/matching policy and `target.OwnerScope` as a narrow provenance handoff, batches exact candidates to a host-supplied Node loader, and returns configs/scopes/failures/effective IDs. CLI, API, and LSP call `DiscoverAutomatic` or `LoadExplicitConfig`; neither the parent config package nor target planning imports discovery                                                                                                 |
| `internal/config/gitignore/`         | Git ignore pattern matching, repository boundaries, and directory reachability                                                 | Staged JS/TS catalogs carry a filesystem-independent cursor through their existing walk, pruning Git-inaccessible subtrees and freezing observed patterns for lint-target admission; low-level pre-resolved configs and LSP bootstrap reuse the same reader through the direct collector; committed LSP fallbacks share the catalog policy. Pure pattern projection is provided by `internal/utils/gitignore`; collection and target-admission policy stay here |
| `internal/inspector/`                | AST/type/symbol/signature/flow inspection for Playground                                                                       | Auxiliary backend used mainly by website Playground inspect panels; builds rich semantic data from `typescript-go` programs                                                                                                                                                                                                                                                                                                                                     |
| `internal/program/`                  | Immutable rslint Program facade and source-generation indexes                                                                  | Privately adapts project/compiler and root-parser hosts into one source, filesystem, syntax, optional checker, module-resolution, and module-reference contract. `Program.ModuleGraph()` and generation-scoped caches derive from that same authority; adapter identity is not observable by linter or rules                                                                                                                                                    |
| `internal/program/loader/`           | Run-scoped Program construction and lint-target binding for CLI/API                                                            | Consumes target-owned lint plans, selects configured projects for matched targets in every CLI/API mode, binds targets by governing-config order and physical identity, and returns one backend-agnostic Program sequence plus its execution projection. Construction choices and ts-go compatibility details remain private; LSP owns resident Programs separately and does not use this request-scoped loader                                                 |
| `internal/languages/js/estree/`      | JS-plugin-only ESTree projection and binary frame codec                                                                        | Projects a selected JS plugin target's existing Program SourceFile into independently owned bytes. Native traversal and Go rules keep their original AST and utilities. It imports syntax and ECMAScript helpers, and owns no Program generation, parser invocation, filesystem, configuration, scheduler, transport, checker, or rule policy.                                                                                                                  |
| `internal/linter/`                   | Unified lint pipeline, immutable lint plans, core engine, plugin dispatch, traversal, and bounded autofix lifecycle            | Owns `RunPipeline`: generation acquisition/release, exact `LintPlan` preparation, native/plugin scheduling, diagnostic aggregation, pure fix planning, all in-memory fix rounds, re-observation, and optional one-shot terminal change commit. It imports no config discovery, Program loader, persistence medium, output, API, or LSP protocol                                                                                                                 |
| `internal/lsp/`                      | Language Server Protocol implementation                                                                                        | Owns selected compiler Programs, owns transactional config discovery and last-good config state, and supplies document/speculative-generation plus progressive-presentation adapters to `internal/linter`; fix-all maps the core's final in-memory delta to one `TextEdit` and never mutates editor/session state while computing it                                                                                                                            |
| `internal/rule/`                     | Rule framework, immutable rule catalogs, configured-rule descriptors, context, diagnostics, fixes, and disable manager         | Defines catalog snapshots and exact object-form ESLint-plugin catalog derivation shared by entry points; `internal/linter` consumes its immutable rule environments, listeners, reporting APIs, and Program-derived cache helper                                                                                                                                                                                                                                |
| `internal/rule_tester/`              | Go-side rule testing helpers                                                                                                   | Supports rule development and complements JS-side testers in `packages/rule-tester` and `packages/rslint-test-tools`                                                                                                                                                                                                                                                                                                                                            |
| `internal/rules/`                    | Core lint rule implementations and final aggregation of every Go rule                                                          | `all.go` consumes the existing core and plugin aggregators once and exposes their shared immutable catalog through `rules.All()`; entry points derive a catalog for the exact request-scoped object-form ESLint-plugin set before config resolution                                                                                                                                                                                                             |
| `internal/plugins/typescript/`       | `@typescript-eslint`-style rules                                                                                               | Its `all.go` contributes rules to `internal/rules`, often relying on `TypeChecker` from `typescript-go`                                                                                                                                                                                                                                                                                                                                                         |
| `internal/plugins/react/`            | React rule implementations                                                                                                     | Its `all.go` contributes rules to `internal/rules` and they execute through the same listener pipeline in `internal/linter`                                                                                                                                                                                                                                                                                                                                     |
| `internal/plugins/jest/`             | Jest rule implementations                                                                                                      | Its `all.go` contributes rules to `internal/rules` and they execute through the same listener pipeline in `internal/linter`                                                                                                                                                                                                                                                                                                                                     |
| `internal/plugins/import/`           | Import plugin rule implementations                                                                                             | Its `all.go` contributes rules to `internal/rules` and they participate in normal config-driven linting                                                                                                                                                                                                                                                                                                                                                         |
| `internal/utils/`                    | Shared utilities for JSONC, compiler hosts, invocation-scoped compiler construction, overlay VFS, and AST/type helpers         | Provides the low-level compiler-construction and snapshot services privately composed by `internal/program/loader`; command entry points do not coordinate compiler hosts or source generations directly. LSP, rule tests, and auxiliary entry points reuse lower-level compiler helpers. Source-generation module resolution belongs to `internal/program`                                                                                                     |
| `packages/rslint/`                   | Main npm package with JavaScript API and CLI wrapper                                                                           | Spawns `cmd/rslint --api` in JavaScript runtime environments and uses `internal/api` message shapes                                                                                                                                                                                                                                                                                                                                                             |
| `packages/rslint/src/languages/js/`  | JavaScript/TypeScript ESTree operations for JS plugins                                                                         | Owns ESTree types, frame decoding/schema, positions, visitor keys, token reconstruction, JSX token decoding, built-in globals and scope analysis. It imports no plugin host, RuleContext, SourceCode, configuration loader, native binding or IPC implementation.                                                                                                                                                                                               |
| `packages/rslint/src/eslint-plugin/` | ESLint plugin SDK, rule execution and worker coordination                                                                      | Adapts final language data to SourceCode and RuleContext, diagnostics, directives and listeners. Host, protocol and WorkerPool coordinate requests without decoding ASTs. It depends on `languages/js`; language modules never depend on it.                                                                                                                                                                                                                    |
| `packages/rslint-api/`               | Frontend-facing encoded source file / AST decoding helpers                                                                     | Used mainly by website Playground to decode AST/source data returned from the Go API                                                                                                                                                                                                                                                                                                                                                                            |
| `packages/rslint-test-tools/`        | Testing utilities and cross-ecosystem rule tests                                                                               | Supports package-side and integration-style tests around the linter and rule ecosystem                                                                                                                                                                                                                                                                                                                                                                          |
| `packages/rslint-wasm/`              | Browser/WASM package for running `rslint --api` in a worker                                                                    | Starts the browser worker, hosts the wasm runtime, and bridges website Playground requests through `internal/api` and `internal/api/server` to `internal/linter` and `internal/inspector`                                                                                                                                                                                                                                                                       |
| `packages/rule-tester/`              | Forked `@typescript-eslint/rule-tester` package used in tests                                                                  | JS-side rule testing support that complements Go-side helpers                                                                                                                                                                                                                                                                                                                                                                                                   |
| `packages/utils/`                    | Shared JavaScript utilities                                                                                                    | Shared support package for the JS/website tooling layer                                                                                                                                                                                                                                                                                                                                                                                                         |
| `packages/vscode-extension/`         | VS Code extension for IDE integration                                                                                          | Activation first checks whether an enabled `rstack.rstack` takes over; if so, this extension stands down to a status bar item and an uninstall prompt. Otherwise it resolves the nearest project-local `@rslint/core` per open document, rejects cores outside `^0.9.0`, launches that installation's `cmd/rslint --lsp`, serves reverse config/plugin requests, and routes diagnostics/code actions to the document's selected runtime                         |
| `packages/tsgo/`                     | `@rslint/tsgo-server` JS wrapper package for the `tsgo` tool                                                                   | JavaScript-facing wrapper around `cmd/tsgo` output; resolves the matching `@rslint/tsgo-server-<platform>-<arch>` binary package                                                                                                                                                                                                                                                                                                                                |
| `typescript-go/`                     | Git submodule containing TypeScript compiler Go port                                                                           | Provides parser, AST, checker, `Program`, `project.Session`, diagnostics, scanner, and VFS primitives used throughout the backend                                                                                                                                                                                                                                                                                                                               |
| `shim/`                              | Generated bridge packages exposing ts-go internals                                                                             | Bridge layer between repository Go code and `typescript-go` internals; generated and updated by `tools/`                                                                                                                                                                                                                                                                                                                                                        |
| `tools/`                             | Shim generator and ts-go update scripts                                                                                        | Generates `shim/` code and maintains the pinned `typescript-go` integration                                                                                                                                                                                                                                                                                                                                                                                     |
| `crates/tsgo-client/`                | Rust client for communicating with `cmd/tsgo`                                                                                  | Spawns `cmd/tsgo` and consumes its semantic/project output from Rust                                                                                                                                                                                                                                                                                                                                                                                            |

The `typescript-go/` directory is the Git submodule for `microsoft/TypeScript`.
`go.work` selects its `tsc/` compiler module, and the compiler and shim Go
module/import/linkname paths use `github.com/microsoft/TypeScript/tsc`.
`packages/rslint-api` imports the JS AST API from the same checkout's
`packages/typescript` sources. The tsgo release workflow copies that compiler's
bundled standard libraries through the upstream `lib` build task.
See [the compiler dependency guide](./CONTRIBUTING.md#typescript-compiler-dependency)
for the pinned revision, migration rationale, and update workflow.

Rule-facing string and matching helpers belong under `internal/utils/`:
`ecmascript` owns JavaScript string values and UTF-16 conversion,
`ecmascript/regexp` owns RegExp rewriting, capture numbering and replacement
substitutions, and `minimatch3` owns the pinned glob grammar and path matching.
`gitignore` shares pure pattern parsing: `ParseText` preserves the configuration
collector's existing glob projection, while `NewMatcher` and
`NewMatcherFromText` adapt explicit pattern lists and ignore-file text to the
JavaScript matching engine. Explicit matching keeps array entries intact,
including embedded newlines; configuration collection retains its existing Git
character-class and case behavior. No matcher discovers files or imports config.
RegExp compilation reuses a process-wide, bounded cache keyed by the exact
source and flags. The regexp package owns rewriting, validation and backend
reuse; rule definitions and per-file initialization do not participate in the
cache lifecycle. Each match borrows an independent backend from the pattern's
pool, avoiding a shared backend's runner lock under concurrent rule execution.
`Unwrap` permanently assigns a private backend to that handle, so mutations stay
local and later operations on that handle keep using the exposed backend.
Shared package metadata queries live in `internal/utils/packagejson`. `Read`
decodes one package object; `FindNearest` preserves the nearest package boundary,
while `FindNearestValid` explicitly skips invalid objects for Node's policy.
Node and Jest share this Program-scoped metadata cache but keep their own field
validation, configuration priority and invalid-package behavior.
Node `bin` aliases, path conversion, workspace dependency checks and npm
publication policy remain in `internal/plugins/node/nodeutil`. Import and require
rules share publication checks and select their own references and resolution
mode. Publication paths and metadata exemptions are relative to the containing
package; compiled publication data is cached per package and Program generation.
Configured pattern caches include the complete ordered list in their key.

`internal/utils/modules` collects import/export, dynamic import, CommonJS and AMD
source expressions without evaluating them or resolving targets. It reuses the
parser's module list when complete and otherwise walks the AST, including empty
sources. A single syntax cache belongs to each exact immutable SourceFile.
`Program.ModuleGraph()` adds compiler resolution and Program-scoped targets;
Node's import adapter applies its own literal coercion, loader-parameter removal,
builtin filtering and type-only policy to the same collection. The shared
collector does not depend on Program, rule configuration or a TypeChecker.

`internal/utils/moduleresolver` provides runtime lookup through the Program's
filesystem and generation-scoped cache, reusing tsgo's resolver for package
traversal and export-path validation. It selects runtime files independently of
compiler declaration lookup and does not load resolved files into the Program.
Explicit options select aliases, extensions, export conditions, package entry
fields and directory entry names. Fallback aliases run only after ordinary
resolution fails; `fullySpecified` limits file/directory guessing for the request
while package entry points and alias redirects retain their own lookup rules.
Package alias fields reuse shared package metadata; all aliases reuse the same
file probes before symlink resolution.
Results retain failure details and separate resource suffixes from filesystem
paths. `internal/utils/tsconfig` shares explicit and nearest config queries using
tsgo's parser, including `extends`, without creating another Program.
Node's option precedence, TypeScript alias policy and extension mapping remain
in `nodeutil`. Restrictions retain lexical paths for unresolved local sources
and compare host filesystem spellings; normalized paths stay at the resolver/VFS
boundary. The import and require restriction rules still share ordered matching
and message construction there. Rules select exemptions, diagnostic nodes and edits.

`internal/utils/npmsemver` shares npm range parsing, minimum-version queries and
subset checks. It reuses tsgo's range expansion and prerelease comparison while
adapting JavaScript whitespace, safe integers and npm's generated bounds. Ranges
retain npm's whitespace-normalized input for diagnostic messages.
`nodeutil` retains the priority of rule options, plugin settings and package
engine constraints, plus Node's default version range.
API reference tracking lives in `internal/utils/referencetracker`. Node rules
and Unicorn's document-cookie rule share static property, alias and
destructuring traversal, reusing tsgo's binding helpers, `RuleContext.Refs`
and a name index cached by `RuleContext` per file. Symbols are resolved within
the linted file; shorthand assignment properties do not introduce bindings.
Each tracker owns its callbacks and active recursion stacks, so independent
paths retain upstream duplicate reports without sharing rule state. Node
module entry points, builtin aliases, and strict versus legacy ESM behavior
remain in `nodeutil`; explicit module traces take precedence over builtin aliases
so rules can distinguish the availability of `fs` and `node:fs`.
Require targets also share constant argument evaluation
and loader-parameter removal there.
The shared constant evaluator reuses compiler number operations through
`shim/jsnum`; JavaScript coercion and bounded BigInt evaluation stay in `internal/utils`.
Its property-name adapters extend literal-name queries with computed-expression
evaluation. Reference tracking uses the no-scope evaluator, so a constant key
identifier is not resolved even when a TypeChecker is available.
Rules own API metadata, ignore options and diagnostic messages. Node version configuration
and range comparisons use package metadata here and the compiler semver parser
through `shim/semver`; rules own their feature availability thresholds.
Relative lookup settings use the process directory from `RuleContext`,
independently of the owning tsconfig's directory.
The glob matcher uses code-unit input for regexp2; rule callers do not select a
character encoding or translate backend capture numbers. These packages do not
decide which files to lint or discover ignore files; those policies belong to
configuration and target selection. Match timeouts remain at the engine boundary.

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

1. **Source and Metadata Loading**: Files come from the real filesystem, an overlay VFS, or LSP document overlays. Each CLI run or API request creates a `program/loader.Session` around its initial immutable VFS view. The session's private compiler hosts keep source snapshots keyed by the exact normalized source path, storing the first successful text read and its xxh3 hash for that generation. The same request scope snapshots successful `package.json` and explicitly registered tsconfig reads by the exact requested path; other JSON, source, ignore, and config-discovery reads remain uncached. When multiple Programs are constructed concurrently, a derived context view coalesces their concurrent cold realpath queries for the same exact path while sharing all parent caches. During autofix, `internal/linter` owns the evolving text in memory and passes only the net changed-file snapshot to the integration's generation provider; CLI/API rebuild request-local Programs over that overlay without changing the base medium. LSP follows document events, retaining Programs selected by open lint documents and reading dependency text from current editor overlays.
2. **Program Loading**: `internal/config/target` freezes targets, their governing config owners, and path identities without consulting Programs. `config.FileConfigResolver` validates and reuses each target's discovery match when compatible, then merges its final configuration once. `config.ResolveProjectPolicy` projects the effective explicit project declaration and service/root/reset options from that same immutable result; it performs no matching, filesystem lookup, or project construction. CLI/API use `config/lint.Resolver` both before construction and after source binding, sharing its file and matched-shape caches. Before construction, file configuration resolves in parallel with at most `GOMAXPROCS` workers; policy projection and failures retain target input order. `--singleThreaded` keeps this preparation serial. Workers finish before the request proceeds, including when a file fails. LSP consumes its existing document snapshot's resolved config. Node only evaluates and serializes values; presets remain ordinary reusable values.

   Ordinary lint takes `project` only from each target's final matched configuration. Merging retains the winning declaration's patterns and literal authored base together; later project values replace or clear both. `config.ResolveProjectPaths` expands that effective declaration; an effective absolute `tsconfigRootDir` overrides its base, while omission or a null reset retains the authored base. No effective project or service means source-only lint, without an implicit root tsconfig lookup. The same config projection resolves the service boundary from invocation cwd for CLI and request `WorkingDirectory` for API, independently of config owners. LSP projects each containing workspace onto the document's lexical parent chain and selects the nearest resulting boundary. A closer directory alias can therefore win over an outer lexical workspace. Each document snapshot freezes this equivalent lexical boundary for diagnostics and speculative fixes, so later workspace-link changes cannot reinterpret it. LSP falls back to initialization `rootUri`/`rootPath`, then server cwd when no workspace context was supplied. Documents outside those roots use their own directory. An explicit `tsconfigRootDir` overrides the default; null restores it. The resolved service root remains separate from the explicit override used by ordinary project paths. Config location, imports, preset composition and `basePath` do not redefine the service default.

   `utils.DirectorySearchBoundary` owns directory ancestry and boundary projection for both LSP workspace selection and project discovery. It preserves exact lexical ancestry, including directory links below that root. For other spellings, lexical ancestors establish containment through frozen physical paths or native directory identity, without assuming equal physical and lexical depths. The outermost physically contained lexical ancestor preserves the root's search range when child links leave, return to, or skip the root. Global case policy and `Realpath` casing alone cannot establish directory identity. Only parent-directory identity participates; a leaf file symlink cannot change the boundary. The helper owns no workspace protocol state, config selection, project parsing or cache. Discovery then walks the target's lexical parents and stops at the exact projected boundary; external targets retain their own-ancestor search behavior.

   `program/loader.Session` assembles one project plan. Broad plain lint and focused CLI/API requests share construction slots and use target membership to select direct owners. Broad lint still validates every effective candidate and ranks direct roots by exact and physical identity before construction; focused requests retain their existing incremental root search and validation range. Targets absent from every applicable parsed root set become source-only gaps without constructing import candidates. The shared CLI/API binder enforces this root requirement even when other targets or program-wide checking retained a Program that imports the file. A listed root rejected by the compiler remains a separate compatibility boundary: broad ordinary lint retains parallel ordered source fallback for that target, while focused selection reports its existing missing-source error. Broad and AllDeclared binding retain the existing fallback only after a root matched; unmatched targets never enter it. Program-wide type checking uses `config.ResolveTsConfigPathsWithPolicy` to retain the complete raw declaration/default range. In combined mode, that full Program set remains available for checking while every lint target receives its own effective candidate list. A pure explicit-project `--type-check-only` run does not scan lint targets. Per-file service/root/reset settings require target discovery to determine their effective values. For type checking, every selected target's root context contributes the owner's complete explicit declaration list, even when its lint binding is disabled or uses service. Service and disabled targets do not trigger an implicit default project. An owner with no selected targets retains the original program-wide declaration/default behavior; there is no file against which to apply scoped overrides.

   Shared predicates do not choose project order or construction scope.
   `program.RootFileIndex` supplies focused explicit-project root membership
   to both loader and LSP. Extension admission remains in the selected project
   construction and does not substitute for root membership. Each selector owns the ordered
   candidates and selected result. Broad loader binding keeps its batch root
   ranking and source indexes. `utils.ExactPathID`, also exposed through
   `config.ExactPathID`, gives source lookup and target binding the same exact
   path key: absolute Windows drive spelling is normalized while directory
   and file casing is retained. Both entries use `utils.ProgramSourceLookup`
   to validate exact Program hits; the loader requests only lexical matching
   before its own batch physical fallback. Root membership's VFS case policy
   remains distinct from exact source identity. These helpers own no build
   queue, declaration ordering, Session, watcher, or cross-request cache.

   `utils.TypeScriptProjectDiscovery` searches ancestor and reference configs using parsed `files`/`include` roots, checking the boundary directory itself before stopping. Directory aliases into the service root use the equivalent lexical boundary, found by resolving lexical ancestor directories rather than counting physical parent levels. If a directory link skips the service root itself, the outermost lexical ancestor physically within that root bounds the search. Truly external targets retain TypeScript's behavior of searching their own ancestor tree up to their filesystem root or `node_modules`; references and extends may cross the search boundary. Discovery receives only a filesystem and config parser; it cannot construct or return a Program. Imported-only targets are not admitted. Metadata identifies source-reference overlap and reference order without loading source files. `disableReferencedProjectLoad` stops reference discovery independently of previous selections; no TypeScript server warm-project lifecycle is reproduced.

   The loader appends selected config metadata to its existing project plan before construction. Config parsing uses a metadata-only host; source snapshots are bound only when Program construction has chosen its serial or parallel filesystem view. Ordinary and service modes share scheduling, source snapshots and one target-binding pass. Target-driven construction consumes service discovery's selected owner directly, then validates its exact source; it does not independently reselect service ownership using ordinary batch matching. The AllDeclared construction path validates service sources before publishing its complete type-check set, including when lint binding will be skipped. Complete service Programs use source references and remain distinct from explicit Programs with different construction options. Every lint target has an explicit candidate list, including an empty list for gaps. Constructed projects retained for other targets or type checking cannot supply types outside that list. A sole explicit candidate still requires parsed root membership; import membership cannot select a lint project. Service candidates never leak into another target's list.

   LSP expands explicit project paths from each document snapshot's effective policy; the config catalog does not validate unmatched project declarations. After excluding ignored documents, generation providers distinguish effective project configuration errors from other lint failures. The diagnostics adapter presents those errors on the affected document and clears cached rule diagnostics and fixes; it does not reject configuration updates for unrelated documents. Speculative generation providers use the same error classification without publishing diagnostics. LSP runs the same metadata discovery against its current editor overlay before requesting a Program. Explicit selection stops after the ordered root scan; it never acquires configured Programs to probe import-only membership. Editor URI decoding uses tsgo's drive-letter spelling for overlays and file lifecycle events; document identity and protocol responses retain the original URI. Normal diagnostics and speculative fixes share the complete binding policy: disabled or unmatched targets use a source-only generation and cannot borrow an unselected project's module graph. Program acquisition receives the selected parsed config, rather than only its path. The resident store reuses a selected Program only while its construction inputs and referenced config snapshots match; a mismatch rebuilds it from the selected metadata. Both explicit and discovered LSP projects load referenced source files from editor overlays; project selection does not change editor construction semantics. A missing source in an otherwise compatible selected Program remains an error.

   Existing watcher coverage owns resident cache invalidation. Recursive watches use URI roots when supported so physical compiler paths remain observable through workspace aliases. Metadata already observed by a request also constrains resident Program reuse, including reference changes with an unchanged root config. Source lookup uses the same editor overlay as Program construction, including directory aliases beneath missing virtual-file ancestry. Document events resolve that ancestry before comparing physical identities, even when the dependency was never selected as a lint root. Construction hosts retain only immutable parsed reference snapshots; unknown references still use the compiler host's normal reads and failed-lookup tracking. Metadata observed before watcher coverage stabilized cannot become resident. Speculative fixes keep their private source overlays and do not mutate editor documents or resident Programs.

   Service construction clones the parsed command line and options before enabling `allowNonTsExtensions`, preserving config identity, complete roots and authored `allowJs`. This admits explicitly listed JS roots without changing config glob expansion or ordinary project options. Unmatched files retain source-only gap capabilities even if another selected Program imports them.

3. **Lexical + Syntax Parsing**: ts-go tokenizes and parses source files into TypeScript-native AST nodes. Source-only roots additionally run the ts-go binder so syntax-only rules retain symbols and lexical scopes before their rslint Program is published.

   Already-selected gap roots share `program.SourceOnlyCompilerOptions`: ESNext target and modules, preserved JSX, JavaScript admission, no default libraries, and no dependency expansion. Direct imports can still resolve to paths; resolution does not add their sources to the root universe. Each construction receives fresh options.

   Construction and capability boundaries remain separate. CLI parses supported gap roots with `program.NewFromRoots`, retaining the loader's grouping, parallelism and request-local source snapshots. Unsupported roots keep compiler compatibility admission. API retains that compiler construction for all gap roots and adapts its bound sources with `program.NewFromBoundSources`, which exposes no checker capability. Both compiler paths already reuse `utils.CreateProgramFromOptionsLenient` for parsing and binding. LSP uses that same helper for one fallback document over its editor or speculative overlay, keeps the compiler-backed facade, and filters type-aware rules when the document lacks a selected project. It does not borrow the CLI/API loader's snapshots or make gap Programs resident. The shared options do not select projects, create hosts, schedule work, cache across generations, or determine facade capabilities.

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

Rule IDs remain complete strings throughout configuration, diagnostics and
plugin dispatch. `rule.Namespace` owns their ESLint-compatible interpretation:
unscoped IDs split at the first slash (`node/prefer-global/url` belongs to
`node`), while IDs starting with `@` split at the last slash. Catalog assembly
and the Node plugin loader accept only rules whose parsed namespace matches
the declaring plugin, so alternative splits cannot shadow another plugin's
rule. Config resolution checks that namespace against the merged declarations.

For repository tooling, slashes in a plugin-local rule name correspond to
subdirectories: `prefer_global/url/url.go` and `url.md` represent
`prefer-global/url`. `scripts/rule-paths.js` owns this source-path convention
for manifest generation and release indexing; it does not resolve runtime IDs.

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
planning that environment belongs to the shared prepared rule set, while each
file retains its own checker grant. Execution constructs settings and globals once per file and copies the resulting
base context for each rule. Module references and whole-Program indexes are not
context fields: generic source references come from `Program().ModuleGraph()`,
while rule-specific derived indexes use `CachedByProgram`. Both remain keyed by
the same Program generation and can never become a second source identity.

Import rules resolve authored requests and alternative spellings through
`modules.Source`-aware APIs. `Program.ResolveModuleNameAt` retains the reference's
resolution mode when comparing candidate names. The shared
`moduleresolver.PathAliasResolver` probes only the selected compiler `paths`
substitutions under that request's mode and attributes, so missing aliases can
still fall back to package lookup. Low-level TypeScript resolver calls remain
inside those resolution services.

The linter creates one short-lived `CommentStore` per file. `Comments.All()`
materializes the scanner-backed, source-ordered, deduplicated comment list only
for the first consumer; later consumers share that list. A source without `//`
or `/*` takes a cheap byte-scan fast path. Inline-global parsing first checks
for an exact raw-text directive candidate, so ordinary files do not force
comment collection.
Comment traversal uses parser-owned token ranges and scans the gaps between
AST children with a scanner local to the walk. It does not populate the
SourceFile token cache; token APIs still return canonical cached nodes.
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
one `Globals` view per file, shared by its native rule contexts. `Globals` and
`Exported` are value handles to private, immutable per-file data; copying a
context copies these handles rather than their private storage. Their zero
values retain the default global set and no exported names,
respectively. Rule identity and diagnostic consumers remain independently bound
to each context. Rules read
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
validates and projects each semantic diagnostic into an
`internal/output.Diagnostic`, computes report counts and one outcome, and then
builds the completed `internal/output.Report`. API projects to its 1-based
structured wire model and flat UTF-16 edit offsets, and LSP projects to 0-based
LSP positions while retaining its stale-generation and code-action lifecycle.
Counts, path bases, stderr notices, and protocol empty-array rules stay
integration-owned; lint/fix observation order and fix-round state belong to the
core pipeline.

Observations that do not plan source changes project diagnostic AST references
to immutable text frames before aggregation. Observations that plan fixes keep
the original frames until source identity checks and fix text freezing finish.
Producer finalization waits for actual generation borrowers; detached plugin
work continues with frozen inputs. Frames are shared by source object identity,
not by file path, and compute ECMAScript line maps lazily. `LintedFiles` requests
selected target paths; `SourceFiles` additionally retains their original ASTs.
The API requests ASTs only for its encoded-source response. Diagnostic
projection does not invalidate those explicitly retained artifacts.

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

1. Ask a `GenerationProvider` to plan acquisition for the core-owned
   `SourceSnapshot`, then acquire one complete immutable binding for the fix observation.
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

`internal/linter/loader` is the shared CLI/API adapter from request-scoped
`program/loader.ProgramLoad` values to `linter.GenerationLoad` values. It maps
admission metadata and deferred acquisition once, while each entry point supplies
its own binding-to-generation projection. This adapter depends on both packages;
neither core package depends on the adapter or on the other core package. Config
activation, overlay rebuilding, and result demands stay with their entry points.
LSP supplies resident Program leases directly through the same pipeline contract.

The linter files mirror those responsibilities: `pipeline_contract.go` and
`pipeline_result.go` define the public boundary; `pipeline.go` is the sealed
entry; `pipeline_load.go` owns acquisition admission, leases, and result merging;
`pipeline_generation.go`, `pipeline_observation.go`, and
`pipeline_native.go` own one immutable observation; `pipeline_plugin.go` owns
plugin task materialization. `lint_plan.go` freezes immutable execution input;
`lint_execution.go` consumes private project tasks and joins their workers;
`linter.go` executes each project's files and rules. `pipeline_autofix.go`,
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

Rslint's runtime configuration follows ESLint flat config semantics (an array
of config entries) and is loaded from JavaScript or TypeScript modules:

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

#### Legacy JSON Migration

Runtime loading of `rslint.json` and `rslint.jsonc` is not supported. `--init`
remains the isolated migration entry point: when it finds a legacy JSON/JSONC
file, it parses that file, writes the recommended JS/TS module config, and
removes the migrated source after the new file is safely created. No CLI, API,
or LSP lint path reads those filenames.

Migration preserves entry order and runtime parser-option resets. An explicit
`tsconfigRootDir: null` reset selects JavaScript output (`.js` in ESM packages,
otherwise `.mjs`), because the public TypeScript option accepts only a string.
This output-format choice does not change the selected TypeScript presets.

### Config Entry Structure

Each entry in the config array supports:

| Field             | Type                                       | Description                                                                          |
| ----------------- | ------------------------------------------ | ------------------------------------------------------------------------------------ |
| `basePath`        | `string`                                   | Entry-local base for selectors, ignores, and explicit TypeScript projects            |
| `files`           | `(string \| string[])[]`                   | Non-empty selector list; top-level selectors are ORed and nested selectors are ANDed |
| `ignores`         | `string[]`                                 | Glob patterns excluded by this entry                                                 |
| `languageOptions` | `object`                                   | `ecmaVersion`, globals, and parser options such as project settings                  |
| `rules`           | `Record<string, …>`                        | Rule level or `[level, options]`                                                     |
| `plugins`         | `string[] \| Record<string, ESLintPlugin>` | Bundled plugin declarations or third-party plugin instances                          |
| `settings`        | `Record<string, …>`                        | Shared settings available in `RuleContext`                                           |

### Configuration Loading

Runtime configuration uses staged JS/TS module catalogs:

CLI, the native JavaScript API path, and transactional LSP refreshes reuse
the one-shot `internal/config/discovery.DiscoverAutomatic` operation (or
`LoadExplicitConfig` for an exact path). Automatic discovery builds an immutable
config/ownership catalog and observes `.gitignore` sources during the same
directory walk. Explicit loading selects the exact module unconditionally, or
accepts an already-resolved inline array without module loading,
then freezes target-tree Git sources with the same frontier
without probing nested config candidates. Neither path applies file selectors;
both preserve the reachable subset of an already-expanded candidate set.
For explicit-project `--type-check-only`, explicit loading omits target-tree
Git collection unless the loaded config has per-file service/root/reset options.
For LSP, the client's first `rslint/configRefresh` may include one absolute
`configPath`. The server fixes that optional path for its lifetime: later client
refreshes must repeat the same choice, while Go-owned `.gitignore` refreshes
reuse it internally. Changing between automatic discovery and an explicit path,
or changing the path itself, requires a new server process. Explicit LSP mode
uses only the selected JS/TS module.
Go owns candidate discovery, default exclusions, config hierarchy, authored and
Git directory reachability, the independent frozen Git policy, and final
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
Node batches, failures, and final activation; the request-local Git reader owns frozen `.git` marker and `.gitignore` observations; and the catalog draft is the sole writer of authored configs and owner scopes. Compiling the Git reader snapshot produces one immutable `config.GitPolicy`, separate from all config arrays. Filesystem workers return frontier results
without mutating those shared transaction results. The coordinator merges each
sorted frontier, loads its candidate batch, and adopts the result before the next
frontier starts; Node activation happens only after catalog finalization.
Physical-directory ambiguity is checked only among successful effective owners
during finalization. A failed candidate reached through another lexical alias
retains its failure metadata without preventing fallback to a valid ancestor.
Node identifies each load candidate by its module IO path together with its
opaque logical config directory, so aliases can report independent failures
before Go selects the effective owners.

Automatic discovery uses these rules:

1. `.git` and `node_modules` are default discovery boundaries. Within one
   directory, automatic filename priority is `.js` → `.mjs` → `.ts` → `.mts`.
   Git-inaccessible directory nodes are config-discovery boundaries regardless
   of whether a JS/TS owner has loaded. Git never filters a candidate filename, so a
   local or ancestor pattern naming `rslint.config.js` does not change priority
   or fall through to `.mjs`. A directly supplied directory/static-glob root
   allows ownership lookup at that root, but not hidden intermediate configs;
   overlapping supplied roots retain independent reachability. When a requested
   root is default-excluded, Go skips downward traversal but still resolves
   reachable ancestors outside the boundary. Supplying an ignored directory does
   not itself make its files or descendants reachable: traversal still requires
   a reachable raw Git cursor or an authored negation that may select descendants.
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
   fails. Lexical ancestry is authoritative; when it has no candidate, the
   frozen physical parent's ancestry is tried before a leaf symlink's destination.
   A chain with a candidate retains its ancestor fallback even if loading fails;
   it does not switch to the leaf destination. A default-excluded
   literal file is not lintable and cannot escape through canonical fallback.
   Git is not used to choose that literal's owner; `.gitignore` is applied after
   ownership is known. A config reached only
   through the literal exception is marked explicit-only: it owns its
   discovery-scoped literal files and is excluded from automatic lint-target
   ownership and handoff. Its presence never changes Git source inheritance.
   Files produced by a
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
   Directory enumeration and module requests use the captured physical
   directory; the lexical directory remains the owner and selector coordinate.
   The catalog passes its frozen authored path spaces to target planning and
   LSP activation, so neither reconstructs them through a moved directory alias.
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
5. Config ownership never resets Git state. Each target uses the nearest `.git`
   directory or file on its physical containing-directory chain, falling back
   to the filesystem, drive, or UNC share root when there is no marker. A
   directory entry is checked against the current repository's rules before an
   admitted nested marker resets the raw Git cursor. Git bounds source
   inheritance, not the lexical config ancestry search. The frozen policy keeps
   blocked directory entries separate from repository-internal matching:
   recursive target walks consult those admission decisions, while separately
   requested roots retain their independent entry points. Adding a literal
   target inside a nested repository therefore cannot reopen its siblings to
   the outer walk.
   A walk kept open only for an authored leaf negation retains its outer
   admission restriction across nested Git roots. Each file must still match
   an effective exception; discovering a marker does not authorize hidden
   configs or unrelated files. Target walks check this route restriction once
   per directory and apply the same ordered global-ignore matcher to its files.
   Route checks use caller-visible directory projections, as do directory
   pruning and file matching, including physical config-owner fallbacks.
   Authored negations may keep traversal open without making raw Git sources
   beneath an excluded parent readable. The supplied-root config ownership
   exception remains separate from this source reachability.
6. Candidate failures remain in `catalog.Failures`; even zero-candidate and
   all-failed builds return complete frozen inputs and Git policy. CLI/API
   derive `ErrAllConfigsFailed` from that catalog and stop. LSP may commit
   first-start unavailable boundaries, or retain its complete last-good
   generation. Only nonempty effective config sets activate during discovery;
   LSP's existing empty-generation activation handles recovery and absence.
   Cancellation, protocol and activation errors still abort the build.

The transport and target phase differ by surface:

- CLI sends `loadConfigs` and the CLI-only `prepareConfigs` as reverse
  framed-IPC requests during initialization. After final effective-ID selection
  and the first fingerprint check, Node starts the plugin-host warmup
  and returns provisional plugin metadata. Go may use that metadata for
  read-only target planning and Program construction. Before `RunPipeline`,
  `activateConfigs` joins the same activation, including its second fingerprint
  check and host publication. The CLI adapter validates that the completed
  metadata matches the prepared metadata. This barrier also runs with zero
  targets, disabled plugin rules, or `--type-check-only`; earlier command
  failures join the pending activation before returning. Initialization errors
  remain fatal and can now appear through the ordinary aborted-run output
  after the interactive start line, before any diagnostics or fixes execute.
  The start line uses stdout; initialization and other aborted-run errors use
  stderr in every format. If Go preparation and worker initialization both
  fail, both errors are reported. Explicit `--cpuprof` and `--trace` outputs
  record the Go preparation and execution reached by this invocation, including
  failure paths. Recording starts before the activation barrier and replaces
  the selected output file; an initialization failure still finalizes that
  recording. Cancellation retains the engine's existing termination fallback:
  a Go child blocked in I/O may require forced termination and report its
  corresponding signal exit code while the engine drains owned workers.
  Native-only configurations return fully verified metadata without a worker
  or a second request. `--singleThreaded` still uses one JS worker and serial
  Go work groups; that worker's initialization can overlap Go preparation.
  The engine owns one activation per invocation: identical requests share its
  completion, conflicting selections fail, and shutdown prevents late host
  publication while draining owned builds with the existing initialization
  timeout. API and LSP activation/commit behavior is unchanged.
  The resulting catalog and the later Go lint-target
  walker are separate traversals, but the staged catalog already freezes the
  Git sources observed on its reachable frontier. There is no second per-owner
  directory sweep. Automatic literal scopes and explicit file-only invocations
  contribute only their exact owner-to-target source chains to the same
  source-keyed projection.
- `Rslint.lintFiles()` sends raw file/directory/glob patterns in one API
  request. Go shares `target.SelectInputs` with the CLI, including literal
  priority, lexical scan entries, equal-root walk batching, and compiled-pattern
  pruning. File contents are an independent read overlay. Exact-file requests
  retain their own identities and are never parsed as patterns.
  Go bounds catalog discovery to each glob root's candidate ancestor trie while
  retaining recursive literal directories, then applies selectors, ignores,
  ownership, and canonical deduplication through the existing target planner.
  Configs in unrelated descendants are not evaluated. Explicit-module and
  inline-only arrays use the same input scope, independently of config location.
  The bidirectional API
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
  published exactly once. An empty catalog applies its fallback override array
  through the same bounded input traversal and Git reader without evaluating
  more modules. Its array base is then the invocation cwd; a discovered module
  retains its own array base. The fallback is not a config owner or candidate.
- Extension activation makes a one-time stand-down decision through
  `rstackEditorTakesOver()`: an enabled `rstack.rstack` alone takes over, and
  `Extension`, `RuntimeManager`, and LSP startup are skipped entirely. The
  migration status bar item and prompts in `src/migrationNotice.ts` never block
  activation; the prompts keep their "Don't show again" flags in `globalState`.
  Standing down prompts to uninstall this extension. Otherwise `RuntimeManager`
  reports each ready runtime; the first one prompts once to install Rstack
  unless an unsupported core was reported earlier. A live language client
  cannot be added or removed safely after activation, so later Rstack state
  changes only switch the status bar item to a reload prompt. Whichever
  extension activated the Rslint language server retains ownership of it.
- The extension owns shared UI, commands, and output channels once. For each
  open supported file, `CoreResolver` follows normal `node_modules` ancestry or
  the resource-scoped `rslint.corePath` override to an exact
  `@rslint/core` package. There is no bundled fallback and no alternate PnP
  execution path. The final standalone release reads that package's
  `package.json` version before loading any core module and rejects versions
  outside `^0.9.0` with an Install Rstack error; no runtime starts for them.
  That package supplies the Go binary, config-module host, protocol version,
  and lazily loaded ESLint-plugin worker host as one coherent runtime boundary.
- LSP tracing has one window-scoped policy and one extension-owned output
  channel across all runtimes. Each `LanguageClient` uses the same `rslint`
  client ID, so `vscode-languageclient` is the sole authority that reads
  `rslint.trace.server`, applies the initial level, and sends live `$/setTrace`
  updates without restarting a native server. The server accepts that standard
  notification as a no-op because the protocol trace is rendered client-side.
  Every `LanguageClient` borrows the same channel directly; only the extension
  disposes it after all runtimes close.
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
  boundaries, keeping the LSP alive without linting through broken subtrees. A
  Node commit retains one rollback predecessor: if the commit
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
and mutable-state owner. `initialization.go` enables the resident Program store;
`config_discovery.go` owns JS/TS transaction prepare/commit/abort, while
`config_watch.go` owns watcher registration and event handling together with
tsconfig-derived state refresh. `document_sync.go` owns the
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
config owner and the existing base for entries that omit `basePath`; the
invocation cwd anchors an explicitly authored `basePath` and remains the
implicit scan and response-path root. The exact config path is never gated by `.gitignore`;
a fixed-owner frontier freezes the target trees' Git policy independently of config success. That frontier never probes or
activates nested config candidates. Automatic candidates instead use Git
directory reachability while selecting ownership.

Config loading resolves each authored `basePath` into the entry's existing
internal path-origin model. The existing files/ignores matcher and explicit
project-path resolver consume that effective directory; target planning,
program loading, and LSP do not interpret `basePath`. Automatic configs resolve
it from the module directory, while an explicitly selected config resolves it
from the invocation cwd. Authored path bases never determine `.gitignore` collection roots.

No-candidate behavior is surface-specific. CLI reports that no rslint config
was found and points to `rslint --init`. Native API discovery performs no
reverse config call and uses `overrideConfig`, or an empty syntax-only config.
LSP explicitly stages and commits an empty plugin-host state (an empty load
batch followed by zero-ID activation) plus an empty-rule workspace fallback
whose only policy is the transaction's frozen `.gitignore` view. That empty
catalog is not a usable module-config last-good boundary: if a newly created
config is broken, LSP commits an unavailable boundary for it rather than
linting through the failed owner.

Low-level API and WASM callers may still send an already-resolved serialized
`config`; JSON there is a wire representation of the typed model, not a config
file lookup. The only disk JSONC parser for rslint configuration belongs to the
`--init` migration path.

### Configuration Merging

Config merging follows flat-config-style semantics in `GetConfigForFile()`:

1. entries containing `ignores` plus optional `name` or `basePath`, and no configuration fields, form the global-ignore set
2. the implicit default extension baseline plus effective explicit `files` selectors defines the config selector union; an entry's `ignores` prevents its selector from extending that union, top-level selectors are ORed, nested patterns are ANDed, and an explicit `files: []` is invalid
3. entries without `files` cascade across that selector union, while entry-level `ignores` prevent only that entry from contributing configuration to an otherwise selected file
4. later rule values override earlier values; a severity-only override retains earlier rule options
5. settings and language options recursively merge ordinary objects, while later arrays and scalar values replace earlier values

`TargetMatcher` produces an immutable `TargetMatch` containing the exact ordered
set of contributing entries and its complete target identity. Target planning
carries this opaque result with each selected `target.File`; it cannot read or
change the entry bitset, merge configuration, or prepare rules.
`FileConfigResolver` reuses the match only for the same complete lexical,
canonical-file and canonical-parent identity, config owner, path-space generation
and matching policy. It checks config compatibility once per originating matcher.
Changes to rules, settings or language options may retain the selection; appended
unconditional entries contribute their values without widening discovery. Changed
selectors, ordered ignores, authored bases or the Git policy generation require ordinary matching.
Targets without a discovery match, including LSP document snapshots and direct
config callers, also use ordinary matching. `PathSpaceSnapshot` remains a read-only
snapshot of authored path bases and stores no file decisions.

The file resolver compiles the matched set into an immutable effective config/rule
plan using its own final configuration and rule catalog. Files with the same set
share one plan for that resolver's lifetime. Shape identity is collision-free (`uint64` for the
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
proof of the target's complete ancestor chain. If parent directories have not
been created, identity capture resolves the nearest existing ancestor and
appends the missing suffix; an existing leaf symlink retains its own resolved
file identity and its containing directory's separate Git coordinate.
When the physical parent is inside the selected config tree, selectors use that
parent plus the lexical basename even if the leaf symlink points outside it.
Leaf resolution starts in the captured physical parent, so a directory alias
cannot switch the file into another parent generation. API canonical-file hints
override only the file identity at the request boundary.
Native-case owner aliases may be verified once while the target's owner is
selected. LSP then freezes one
document lint snapshot containing the same complete target plus its selected
owner, resolved config, matching rule catalog, and project paths; Go rules, plugin rules,
diagnostics, resident Programs, and every fix pass consume that snapshot. No
later stage resolves the target file or selects its owner again. Staged CLI,
native API, and transactional LSP paths therefore reuse the same Go ownership
rules instead of independently reconstructing hierarchy on the Node side.
Editor overlays use the same ancestor identity for source reads, directory
entries and metadata reached through another directory alias. Open virtual
dependencies share their physical keys, preserving the document's type
information in configured Programs and speculative fixes.

Configuration meaning and invocation scope are separate inputs. For one
invocation-wide config, `ConfigDirectory` is the config file's directory and is
the authored base used by entries without `basePath`; `ScanRoot` is the
invocation working directory used by implicit/no-argument target discovery and
no-argument target root. Git boundaries depend on each target instead. A composed config entry may retain a
different authored base, and an entry with `basePath` receives its resolved
effective base during config loading. Automatic config catalogs already encode
one owner directory per config and do not use `ScanRoot` to infer ownership.
`config/lint.ResolverOptions.DefaultRootDirectory` carries invocation cwd
separately for project-service policy, including multi-owner catalogs; config
ownership never replaces that default. API requests do not change process cwd.

`target.Request` is the lint-target boundary shared by CLI/API target
selection. Explicit files and recursive directories form one union; omitted
CLI targets become the invocation cwd, and multiple files/directories do not
change one another's config matching. Each planned target retains its
caller-visible path, file and parent filesystem identities, established
config owner, and opaque discovery match. For each literal file request, the plan
also carries the existence and ignore outcome produced by that same discovery decision; CLI
warning rendering never re-reads the filesystem or re-resolves ownership. The
plan exposes the read-only `config.PathSpaceSnapshot` observed during discovery,
but does not own a rule catalog or execution config. CLI `--rule` and API
overrides are layered after target selection and config resolvers evaluate them
in those same frozen path spaces.
`internal/program/loader` consumes the plan but cannot add targets, change
ownership, or reinterpret config-relative paths.
Its request-scoped `WithFrozenIO` view projects filesystem operations through
the recorded directory and file identities before source and metadata caches
are created. Program filenames retain their authored spelling and extension;
relative config paths are normalized in that path space before projection.
Canonical-file hints map individual files and never establish directory roots.
API request buffers and every fix generation place their memory overlay above
this same physical view, preserving exact caller-buffer precedence.
Program IO follows the filesystem's case policy while prioritizing exact frozen
identities; ambiguous case-equivalent anchors never collapse distinct roots.
Each absolute disk-path query normalizes once. Folded lookup keys omit an immutable
shared directory prefix, so its case folding is reused and ordinary lowercase
dependency suffixes allocate no strings. Differently cased prefixes use complete
path folding before removing the folded prefix; root, mixed-root and non-disk
views retain complete keys. Original and folded ancestors advance independently,
without retaining query-specific caches. An unchanged directory
anchor bypasses relative-path reconstruction; frozen file identities still take
precedence in `Realpath` and fix-overlay lookups.

After Program binding, `internal/config/lint` maps each Program source path
through its `Resolver` back to that immutable selected target and resolves rules
only through the target's recorded config owner. In invocation-wide single-config mode an
unbound source still resolves against that one config; in an automatic
multi-config catalog an unbound source receives no config. CLI and API share
this decision and do not reconstruct owner selection independently.

Git policy is independent of config-entry origin, config ownership and cwd.
`config/gitignore.Reader` caches marker results, parsed sources and misses for one
request. Each directory's observation initializes once; unrelated directories
perform IO concurrently, outside the reader's publication mutex. Snapshots join
already-published source observations outside that mutex and copy their patterns.
Its pure cursors keep raw source reachability separate from authored
config admission. `config.GitPolicy` compiles that snapshot once and retains
physical repository roots, source-directory groups and precise covered
containing directories; it owns no VFS, module loader or owner lookup.
A file's Git coordinate is its frozen canonical parent plus original basename,
so a leaf symlink cannot switch Git repositories. Directory aliases use their
frozen physical roots. Requests hand those typed identities from discovery to
target planning rather than resolving them again. Distinct explicit directory
roots remain independent even when one contains another; final selected files
are deduplicated as before.

`target.Request` receives the policy and `target.Plan.GitPolicy()` carries it to
`config/lint` and file config resolvers alongside `PathSpaces()`. Matching uses
one Git path state across the target's source chain, deepest source first;
all Git patterns share its repository-root coordinate. Authored global ignores
then apply in their existing order. Directory pruning combines Git coverage
with authored negation reach. Reusing a carried target match requires the same
Git policy pointer as well as the same identity, authored selection prefix and
path-space snapshot. No target, Program, rule or plugin phase rereads Git files.

Shared path identities and containment ignore ASCII Windows drive-letter case
while retaining directory and file spelling. Catalog keys and caller-visible
paths keep their original strings. These comparisons need no realpath proof
for a drive-only difference, including for new files that do not exist on disk.
Owner indexes, Program source validation and source mappings share the drive
primitive in `internal/utils`; config still owns path-space and ownership policy.
Editor overlays accept either drive spelling for files and directories without
folding names or retaining a separate alias index. LSP uses that drive identity
when assembling overlay content, so the selected document overrides other open
aliases of the same target. Before loading config modules,
discovery chooses one raw owner-directory spelling for drive-equivalent
candidates, so candidate priority and plugin routing share one owner.
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

- Config arrays contain authored entries only. Git rules are neither injected
  synthetic entries nor partitioned by config owner. File-only requests observe
  only ancestor chains; recursive directory requests add reachable descendants.
  No source collection scans unrelated repository subtrees. Nested markers
  remain authoritative even when they produce no patterns.
- LSP adds open-document identities as Git-only exact inputs in the same
  discovery build. They never become config discovery seeds. A committed policy
  covers exact physical containing directories, not every descendant of a
  known repository. New documents and uncovered documents with moved directory aliases trigger the existing
  serialized refresh after the config handshake; failed refreshes leave them
  unavailable for diagnostics, quick fixes and fix-all. Content edits do not
  repeat a failed refresh for the same physical directory. Config/watch events
  can retry it, and a successful commit clears the failed-directory set.
- Dynamic Git watches cover workspace lexical/physical trees and open-document
  ancestor chains for `.gitignore` and `.git` marker creation/deletion, never
  `.git` contents. Watch ranges are registered before source observation and
  replaced without dropping last-good or newly requested ranges after failure.
  Once the current range is registered, old-subscription cleanup failures do
  not block configuration refresh. The existing watcher set retains failed or
  uncertain registrations for cleanup on a later refresh; cancellation still
  aborts the transaction.
  Extension watchers still own descendant JS/TS config and lockfile refreshes;
  Go retains strict-ancestor JS config watchers.
- LSP prepares and atomically commits authored configs, unavailable boundaries,
  the Git policy and the matching Node generation. Workspace fallback remains
  a plain empty config using that same policy. Preparation and commit never
  recollect Git files. First-start all-failed recovery uses the complete catalog
  rather than fabricating one; a policy alone does not make it usable last-good.
  Before the handshake, bootstrap uses a request-local direct collector. A
  refresh failure preserves config and Git policy from the same previous
  generation, while a successful no-candidate transaction commits an empty
  plugin generation and the newly observed Git policy.
- bundled Go and third-party object-form plugin rules are always gated by their normalized declared prefixes; each CLI run and API request derives its own catalog, while LSP commits a module-owner catalog with the matching Node generation. Replacing or removing plugins therefore replaces that catalog instead of retaining process-wide placeholders
- CLI/API lint target selection is independent from TypeScript `Program` membership and considers only rslint-supported script extensions. The `.js`, `.mjs`, `.cjs`, `.jsx`, `.ts`, `.tsx`, `.mts`, and `.cts` default baseline is always selected; explicit config `files` contributes candidates only within the supported set. Global ignores and `.gitignore` remove targets, while an entry-level ignore prevents only its own selector/config contribution
- selected CLI/API targets can still appear as 0-rule lint results when no config entry contributes rules; this applies to default-baseline directory discovery and explicit supported files, and syntax diagnostics remain available in that state
- automatic Rslint config discovery selects the nearest loadable lint config; explicit config selection uses that config directly. Matching entries determine the effective explicit project, service, root and reset options. Service discovery starts at the lexical source location; ordinary selection uses the first matching parsed root in the effective project list. Import-only membership cannot select a lint project
- omitting both project settings leaves ordinary lint source-only; the historical config-directory default applies only to program-wide type checking. Effective `project: false`, `null`, or `[]` clears the target's ordinary explicit binding; `projectService: true` instead discovers configured ownership; unowned files use source-only gap linting. Built-in TypeScript presets do not set either project option, matching upstream preset composition. `tsconfigRootDir` is an absolute ancestor-search boundary, not a forced project; service discovery defaults to the CLI/API invocation directory or the LSP document workspace described above
- `files`/`ignores` matching uses the stable target path in the governing config's path space; a ts-go Program source alias is used only to locate the AST and type information, so moving a target into or out of a tsconfig cannot change its rule configuration
- within each Program-registry build, normalized declared tsconfig paths are deduplicated across config associations; each CLI autofix observation creates a new Program generation over the core-owned memory overlay and repeats the original construction scope without writing the base filesystem. Parsed-root lint ownership is therefore recomputed for each source generation; changing imports alone cannot promote an unmatched lint target to a typed project. File-symlink declarations remain distinct because TypeScript resolves relative paths from the declared location. Selected CLI files outside the governing config's parsed tsconfig root sets are parsed and bound as independent rslint Programs. Targets whose names collide under a case-insensitive ts-go path key are partitioned across independent Programs so distinct physical files and package scopes remain distinct
- `--type-check` and `--type-check-only` check complete explicit declarations plus service-selected Programs. Per-target root contexts affect path resolution, while disabled lint bindings do not remove declared projects from program-wide checking. Only configured Programs expose program diagnostics. Pure explicit-project type-check-only skips target discovery and file-config resolution; service/root/reset options require target discovery but never execute lint rules
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
`lint_generation.go` maps core-owned memory snapshots to deferred request-local
overlay VFS acquisition plans; and `lint_commit.go` implements only the optional one-shot terminal
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
plans acquisition over the request's existing file-content/canonical-path
overlay and any core snapshot; the core-owned bounded autofix lifecycle returns a verified final
observation plus per-fixed-file source for the response instead of requiring a
separate API fix path. API-specific path
bases, opaque plugin routing keys, reverse transports, structured responses,
and the inspector cache remain in `internal/api/server`; planning, execution,
and mutation sequencing do not.

1. **Node.js Wrapper**: parses args, starts the Go engine, and hosts JS/TS module evaluation plus live third-party plugin objects
   CLI positional inputs and native API `Rslint.lintFiles()` patterns use `target.SelectInputs` before config discovery. The flag parser retains raw positional inputs; the selector owns their filesystem classification. The selector accepts a `vfs.FS`, gives existing files/directories literal precedence, and reuses the compiled `minimatch3` grammar for static roots, full matching, and directory pruning. Equal lexical roots share a walk; parent/child roots and physical aliases remain independent. Candidates retain only the scan roots that actually matched them, while literal inputs take precedence over overlapping scans. Config discovery bounds glob roots to candidate ancestry and preserves recursive literal directories; target planning owns config admission and canonical deduplication. The exact-file IPC and LSP document paths never reinterpret filenames as patterns.

   API `patterns` and exact `files` are separate request contracts. `fileContents` is a read overlay: it can supply virtual exact targets and dependencies but does not add entries to the raw pattern search. Omitting both target fields retains the low-level overlay-target contract; an explicitly empty target array stays empty. The selector's enumerable VFS is independent from the Program's overlay view, so other filesystem implementations can reuse selection without importing native OS or Node glob logic. Protocol version 5 prevents older binaries from silently ignoring raw patterns.

2. **Config Catalog**: for automatic JS/TS discovery or an explicitly selected JS/TS config, Go builds the staged catalog and batches exact module-evaluation requests to Node. If automatic discovery finds no candidate, or a non-JS config was explicitly selected, the existing Go JSON loader path remains in control
3. **Mode Selection**:
   - `--lsp`: starts the LSP server
   - `--api`: starts the IPC API server
   - default: runs direct CLI linting
4. **Lint Target Plan**: Go resolves a stable target set from CLI/API scope, the implicit default baseline, explicit config `files`, global ignores, and `.gitignore`. Combined `--type-check` resolves lint targets and their effective policies before building the complete checking set. Pure explicit-project `--type-check-only` skips the scan; service/root/reset options require effective target config first
5. **Project Loading**: `program/loader.Session` accepts raw project declarations, frozen targets, and sparse service/root/reset policies projected from the shared file resolver. It assembles one project plan and adds service-selected config metadata before construction. Ordinary lint shares target-driven construction while preserving broad/focused validation and binding semantics; program-wide checking retains complete declarations. Explicit and service construction modes remain separate; per-target candidate lists carry scoped root choices and explicit gaps
6. **Target Binding**: the loader locates the selected project's AST by lexical or frozen canonical identity without changing project policy or target scope. Disabled and unmatched targets, including explicit import-only files and service misses, receive source-only Programs. They cannot borrow a Program selected by another target's parser policy. CLI/API receive one Program sequence and exact parallel target projections; they never construct compiler hosts or implement ownership

   CLI, API, and LSP provide ordered acquisition units in a `GenerationPlan` through the same `GenerationProvider` port. Planning does not construct Programs. The core invokes each unit's `Acquire` to construct a complete binding or borrow an editor generation, and owns its release through the common observation executor. For CLI/API, `loader.PlanLoads` can establish disjoint target ownership from project metadata; each unit still constructs a complete configured Program and keeps unmatched roots in their existing complete grouping. Broad and focused requests share the materialized loader's root-selection authority, including focused declaration-prefix validation. Root readability checks run concurrently unless the request is single-threaded, and all checks join before any acquisition units are published. The API preserves compiler-backed root representations for its encoded AST contract.

   The provider resolves the binding boundary before publishing executable work. Source aliases, compiler-rejected roots, and source remapping retain a complete binding. Fix requests retain their coherent source-identity boundary, and program-wide type checking retains the complete declared Program set. These requirements do not depend on whether rules run natively or in JavaScript. Every unit uses the same preparation, native/plugin execution, diagnostic projection, and lease handling, including API requests for edits or ASTs. The existing CPU/shard admission budget applies before construction. Single-Program units transfer their reservation to native shards; complete source-only root groups keep their existing checker-free scheduling. A plugin AST borrower retains capacity until projection and dispatch finish. All units share one observation-wide plugin dispatch capacity and buffer pool. Results merge in logical unit order, preserving global plugin config grouping and applying plugin failure policy once. A load failure joins acquired work and returns an error without publishing partial native results or retrying rules. Entry adapters only present acquisition errors. LSP release ends the request borrow; the editor owner retains and prunes its resident Programs according to open documents.

7. **Core Request**: CLI/API/LSP choose a sealed `RunPipeline` request kind and provide semantic adapters. They do not call `PrepareLintPlan`, `RunLinter`, plugin dispatch, or fix application as separate stages.
8. **Rule Plan Preparation**: for each acquired binding, the core calls `PrepareLintPlan()` with its ordered Program sequence, its exact parallel target projection, and the generation's rule resolver. Planning never discovers, excludes, or scans for fallback targets; a target absent from its bound Program is an invariant error. The immutable result owns that Program sequence, freezes rare syntax diagnostics in a sparse plan-level projection, and keeps each hot per-file execution unit limited to its source, a shared prepared rule set, and checker capability. Prepared sets own eligibility, ordered native/plugin projections and rule-derived environments; execution consumes these views without filtering rules again. Syntax-error and zero-rule files remain selected, and the same non-empty file/rule projection feeds third-party plugin dispatch. Worker results are joined in stable Program/file order; cancellation cannot publish a partial plan, and callback panics return through the caller so the enclosing generation lifecycle remains exact-once.
9. **Rule Execution**: the core invokes `RunLinter()` only with the prepared plan plus scheduling, diagnostic, timing, and optional program-wide type-check concerns. A nil plan explicitly skips lint for `--type-check-only`. Execution never recollects files, re-resolves rules, or accepts a second Program authority alongside a plan. Plans bind exact Program pointers, so every autofix re-observation constructs a fresh Program and plan over the current in-memory snapshot. When `--type-check` is enabled, Phase 2 schedules only Programs that expose complete program diagnostics.
10. **Fix Rounds and Aggregation**: `RunPipeline` joins native/plugin results, projects stable target paths, computes whole-file changes, applies them to private memory, and reacquires a generation until stable or bounded. A file may move between Program generations when its import graph changes, but the target plan remains stable. Every plugin frame uses that observation's exact Program syntax and source snapshot. Joined CLI/API work holds the generation lease through projection and plugin completion; detached LSP work freezes bytes and metadata before release. CLI independently supplies a terminal committer and therefore touches disk only once after the last successful observation; API and LSP consume the returned memory delta without a committer. The CLI committer accepts only selected caller paths and uses their frozen canonical destinations for both fresh stale-content preflight and writes. Errors and confirmed paths retain caller spelling. Integrations retain only presentation policy for structured plugin notices, path conversion, and output/protocol mapping.
11. **Report Assembly**: the CLI's concrete, side-effect-free report projection validates final post-fix diagnostics and converts them from rule-domain values to output-owned values. It computes error/warning/type-error counts and one outcome from the same snapshot plus `--max-warnings`; the completed `Report` supplies that outcome to both status rendering and exit policy. Only the default format requests a `Summary` and immutable source snapshots for code frames. During initial acquisition, before execution, the CLI asks the loader to freeze roots from the actual type-capable Programs selected for the run. It combines their identities with already-frozen lint-target identities and immediately reduces them to the canonical-union file count consumed after execution. Machine formats consume projected positions, do not request root identities, and do not construct fake zero-valued summaries. `--quiet` filters rendering only.
12. **Output Formatting**: after config/rule preflight and before Program/target work, the CLI asks `internal/output` to emit a default-format start line. A complete report ends with a mode-specific success/error status, while a post-start runtime failure uses a separate incomplete-run status. The package consumes only the completed presentation report and owns its text, format dispatch, and ANSI spans; the command owns lifecycle ordering and report/outcome assembly. JSON line, GitHub workflow command, and GitLab Code Quality formats remain diagnostics-only and never emit lifecycle text or ANSI styling.
13. **Output Synchronization**: in the default IPC CLI, Go sends the rendered lifecycle start as an acknowledged `output` request; Node responds only after the real-stdout write callback completes, so post-start work cannot emit inherited stderr ahead of the start line. Remaining stdout text uses ordered, bounded `output` notifications for large diagnostic sets. Go then sends its terminal `shutdown` request. Node seals forwarding and waits for every remaining real-stdout write callback before acknowledging; only after Go receives that acknowledgement does it emit deferred stderr (currently the `--timing` table), so merged `2>&1` output cannot place the table before or inside the report.
14. **Exit Code**: uses the same command-owned outcome rendered by the default status; diagnostic errors take precedence over an exceeded warning limit

### Concurrency Model

The workload work groups and pools below honor `--singleThreaded`.
The flag serializes these workload stages, but IPC transport, diagnostic
collection, and plugin dispatch may still use infrastructure goroutines.

1. **Lint-plan work group** (`PrepareLintPlan()` via `core.NewWorkGroup`)
   - Resolves per-file rules into stable rslint Program slots with at most
     `min(GOMAXPROCS, selected file count)` workers.
   - `--singleThreaded` resolves the same slots serially in Program/file order.

2. **Native lint scheduling** (`RunLinter()`)
   - Schedules one task per rslint Program; checker-capable work retains
     checker-exclusive shards, while sufficiently large checker-free Programs
     use bounded file chunks.
   - When the Program count exceeds `GOMAXPROCS`, one dispatcher reserves
     estimated unfinished shards before starting each Program, then corrects
     the reservation after checker grouping. Each completed shard returns
     capacity immediately, allowing another Program to start before its
     predecessor finishes. Smaller Program sets keep the direct work group.
   - Both the shard budget and active Program limit are `GOMAXPROCS`. A whole
     Program may cross the shard budget to preserve its existing checker
     parallelism. The budget controls scheduling, not bytes or checker count.
   - Only the dispatcher waits for capacity. Program completion joins every
     entered shard before returning unstarted reservations and releasing its
     slot, including on panic or abnormal worker exit. Execution consumes
     private task slots; shared plans, Programs and checker caches stay valid.
     Program-wide type checking starts after the native phase joins.
   - `--singleThreaded` collapses the work group to serial execution.

3. **Type-check work group** (`runTypeCheckAcrossPrograms`)
   - Schedules diagnostics for real tsconfig Programs and merges results in
     stable Program order.
   - `--singleThreaded` computes Program diagnostics serially.

4. **Staged catalog discovery and Program creation**
   - A bounded Go worker pool scans one reachable sibling frontier at a time.
     Config-boundary nodes are suspended, batched after that frontier is
     processed, and resumed only after the Node result is merged.
   - CLI and native API glob roots carry an exact target-ancestor trie from
     Go input selection, so only branches leading to matched files enter those
     frontiers. Literal directory roots remain recursively unbounded, including
     mixed file/directory/glob invocations. LSP retains its document-driven
     orchestration. Explicit-config discovery uses the same Git reader without
     searching for config candidates: exact-file inputs observe their containing
     directory chains, and requested directories recurse from their frozen roots.
   - Directory-root ancestry is loaded outer-to-inner before the root frontier.
     The current owner's authored global ignores and the independent raw Git
     cursor control continuation below that boundary; local Git sources are observed only after
     the directory's config decision. Literal files use the separate
     nearest-first ownership exception described above.
   - Discovery commits the authored config catalog, frozen target inputs, and
     independent Git policy in stable order. Lint then runs the separate target
     walker with that same policy. Plain explicit-project type-check-only skips
     target planning; per-file service/root/reset options require it. Focused
     construction selects parsed roots and leaves unmatched targets as gaps.
     Program-wide type checking retains complete declared projects and adds
     service-selected projects; lint binding does not narrow their diagnostics.
   - Configured Program identities, config associations, and result slots are
     planned serially in stable config/project order. Construction then uses at
     most `min(GOMAXPROCS, Program count)` workers and merges results and
     errors by the planned order.
   - Focused lint parses each governing config's tsconfig frontier in declaration
     order. Independent config frontiers and confirmed direct winner Programs
     may run concurrently; ownership results remain in declaration order so a
     later completion cannot overtake an earlier project. Unmatched files do
     not schedule import probes.
   - `--singleThreaded` executes the same state machine with one Go discovery
     worker and serializes module evaluation within each Node frontier batch.
     Coordinator batches and results remain ordered in either mode.

   - During independent request-owned loads, the parsed-AST index uses weak references: overlapping Programs can reuse an identical AST, but the cache cannot keep an otherwise completed Program's sources alive. Immutable source-text snapshots remain request-owned, so a later reparse observes the same bytes. Root-readability preflight and compiler parsing share these snapshots through the same Program filesystem view; a bounded file-reading pool avoids both serial project reads and duplicate source I/O. Whole-binding acquisition releases construction references before execution. Editor-owned caches remain governed by their session leases. No global GC policy or fixed Program-count limit is introduced.

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
   - Independent binding realpath lookups use `core.NewWorkGroup`;
     `--singleThreaded` runs the same work serially.
   - A separate operation-scoped projection freezes roots from the actual
     type-capable Programs selected for the run only when the default CLI
     summary needs them. Frozen lint-target identities seed the pass; every
     remaining root uses the VFS's authoritative per-file realpath operation
     through one bounded worker pool. It never enumerates a directory or
     infers a file identity from its parent, so work is bounded by the
     requested roots and remains valid for overlay VFS implementations. The
     result is deduplicated in declaration order. This complete projection is
     independent from the partial root set needed to decide target ownership;
     the command immediately reduces it to a scalar and retains neither
     identity set in loader or `Program` state.

7. **JavaScript plugin workers** (`packages/rslint/src/eslint-plugin`)
   CLI, native API, and LSP share `createPluginLintHost`. `WorkerPool` owns
   bounded, demand-driven scheduling and thread lifetime within each host;
   adapters own configuration activation and generation lifetimes. Plugin
   loading uses the config entry versions selected by `ConfigModuleHost`.
   The CLI and native API load the dedicated `dist/eslint-plugin/host.js` bundle
   for worker coordination and protocol conversion. The CLI uses a path relative
   to its built module in `dist/`. The API resolves the existing public
   `@rslint/core/eslint-plugin` entry without evaluating it, then loads the private
   sibling `host.js`. Anchoring this path to the installed package preserves
   consumers that bundle the API into a different directory. This entry excludes the per-file lint runtime
   and frame decoder, which load inside `lint-worker.js`. The existing
   `eslint-plugin/index.js` entry retains the full exported runtime for direct
   callers and LSP consumers. Both entries bundle the same host implementation
   and resolve the sibling `lint-worker.js` relative to their module URL. The host
   is a private CLI/API asset, not a new package export; `build:js` emits it in a
   separate library block to keep the worker independent of shared chunks.

`internal/ipc.Channel` is the common Go/Node data transport. Its request API
accepts an application payload plus optional `Text` or `Bytes` attachments.
Text has JSON-compatible Unicode semantics; bytes preserve arbitrary binary
data, including invalid UTF-8 and NUL. Storage, publication and lifetime do not
depend on file paths, source code, plugin rules or an application message kind.
Inline binary attachments use base64; negotiated shared memory avoids that
encoding and carries only attachment ranges in the control frame.
Shared-memory acceleration currently covers Go-to-Node request attachments;
Node-to-Go attachments use the same text/binary contract inline.

One logical request can use multiple memory slots. The channel packs complete
attachments into the available storage, including attachments that span slots,
without splitting the application request or merging application results.
The linter retains its rule/config grouping, splits lightweight inputs into
32-file work items, and uses its existing bounded concurrency for both
homogeneous and fragmented groups.
Only files whose prepared lint plan contains enabled JS plugin rules enter
`internal/languages/js/estree.Builder`. When plugins are absent, their rules
are disabled or their targets do not match, the pipeline produces no ESTree
frames or plugin attachments.
The builder reads the acquired Program's concrete SourceFile without modifying
its AST or invoking a second parser. Native Go rules continue to traverse the
original ts-go AST. Each window opens a dispatcher-owned output lease and holds
at most 32 files, targeting 64 MiB of frame bytes. A larger individual frame uses
its own window within the transport's capacity. `Builder.BuildInto` requests the exact
final size through a plain allocation function; the language encoder knows
nothing about storage, reuse or transport.

For CLI/API shared transport, `internal/ipc.RequestOutput` reserves unpublished
storage in the existing mapping. Projection writes its final frame directly
there, with no intermediate Go-heap output frame or publication copy. Allocation
order is preserved by the attachment ranges; a failed encoding rolls back its
last allocation. Pre-frozen inputs are copied into that window's storage rather
than extending a reservation into another window. If an additional frame would
exceed the window byte target, it remains an unconsumed input and is rebuilt in the next
window before allocating output.

Publication ends producer access to those bytes and transfers slot ownership to
the existing exact reader-acknowledgement protocol. Failure before publication
releases every reservation. Shutdown cannot unmap under an active encoder, but
published outputs do not pin the Go mapping while Node executes rules. Full,
fragmented or unavailable storage retains the existing owned/inline byte route;
no parser fallback or old plugin protocol is involved. The Node side continues
to read independent writable bytes and materialize ordinary mutable ASTs.

Inline/LSP dispatch uses the owned-frame adapter. Small frames share stable
output blocks whose combined retained capacity is bounded; growth never copies
completed frames or changes their views. Retained scratch is capped at 4 MiB per
window. Owned output adds at most 4 MiB per window; shared output does not allocate
those Go output blocks. Before dispatch the scheduler drops frame references
from result metadata and discards source references and oversized scratch.
Scratch/output caches never own Program references or reuse an in-flight frame.
Reset evicts the largest scratch columns until their combined capacity is at
most 4 MiB, preserving smaller reusable columns. A cleared string dictionary
is reused only while it has at most 4,096 entries; larger dictionaries are
cleared and dropped after the file. Projection reuses one scanner value and
zeros its complete state, including source-backed numeric caches, on every
return.
Already frozen LSP inputs retain their independent ownership.
Projection owns a reusable sparse UTF-16 offset column. It does not populate
the native SourceFile's position-map cache; the column is cleared after each
file and participates in the same 4 MiB scratch budget.
The offset index stores sparse 64-byte blocks with prefix deltas and rune-end
bitmaps. Population counts answer byte-offset queries, including offsets inside
a UTF-8 scalar, without a record per Unicode character or a retained source view.
String text is staged while frame sizes are counted, then encoded directly into the final frame
instead of a temporary UTF-16 byte column. Success, failure and cancellation clear every
staged string reference. The private frame references strings by sequential IDs.
An immutable source-backed string may use a checked UTF-8 source range instead
of a duplicate UTF-16 record. Other strings, including unpaired surrogates,
retain UTF-16 encoding. The matching Go producer owns record shapes, references
and ranges. Node checks that the header and column extent fit the received bytes
before allocating, then indexes strings in one pass and materializes nodes
directly. It uses no list-boundary bitmap, validation graph or per-file string
hash table. A private array caches independently decoded strings, retaining no
source substring or transport view. Empty wire lists share a single record,
while each Node-side read still creates an independent mutable array.
Token coverage uses a private, source-ordered scalar buffer; the scanner walks
it alongside tokens, preserving the last cover at repeated source positions.
Parsed compound operators retain their full spans while TypeScript generic
closers remain separate. JSX attribute strings use their parsed boundaries,
including literal backslashes and line breaks.
Projection follows Espree's JS/JSX data contract and TypeScript ESTree's TS/TSX
contract, selected by the SourceFile's script kind. This includes JSX entity
values, token classifications, optional fields and template raw values; it
does not change syntax admission, rule execution or diagnostic/fix policy.
Unparenthesized comma chains produce one final expression list, without
encoding intermediate prefixes. Parentheses retain nested sequences.
Optional-chain operands reclaim replaced wrappers before encoding; parentheses
preserve their separate short-circuit boundaries.
Directive prologues advance once per statement list. Token classification uses
the projection's JSX name context rather than walking each identifier's ancestors.
Long trivia runs share a per-file token-start index, avoiding repeated scans
at the common full-start position of deep expressions. Short spacing is not
indexed; the index is dropped on every success, failure or cancellation.
Shared publication carries ranges for the already encoded bytes; inline fallback
serializes owned bytes. Shared slots require the existing reader release
acknowledgement before reuse. No whole-group frame cache is retained. Private pipeline input slots transfer their syntax
references to the bounded projection windows and are cleared as consumed.
Diagnostic metadata shares the native source's independent text frame for
read-only observations, preserving exact source identity without retaining
syntax. Fix planning retains the original diagnostic source until its existing
text validation completes. Public dispatch helpers preserve caller-owned input
slices. The Node host applies `--singleThreaded` to its worker count; reverse
requests keep the existing bounded scheduler.
Projection checks cancellation and the encoding's integer representation while
generating nodes, strings, lists and tokens; it imposes no separate file-size
quota. Initial capacity estimates limit speculative reservations, while columns
can grow for larger frames. Unsupported syntax fails explicitly rather
than producing a generic node or an empty successful lint.
The temporary token-cover index stores 32-bit byte offsets within that encoding.
It contains no AST pointers and is discarded when the builder resets.

The frame contains the final ESTree schema, UTF-16 node/token/comment offsets,
UTF-16 literal strings, numeric values, the BOM-stripped UTF-8 source, and an
explicit original-BOM flag. If source bytes are not valid UTF-8, Node preserves
the compiler's code units and original byte offsets for diagnostics and edits;
the decoded text and position map own their storage. Go and Node ship together and use one private
format, without per-frame version negotiation or schema fingerprints. The schema
contract is checked by the existing conformance suites. Go preserves source
ranges, authored modifiers and binding contexts while excluding compiler-created
JSDoc syntax. SourceType, ecmaVersion,
syntax-error admission, rule selection, reporting, disable directives and fix
policy remain in their existing owners.

For source with invalid UTF-8, a compound plugin edit keeps its single atomic
range and carries replacement fragments: frozen-source byte ranges and new
text. The Go plugin-result adapter copies the original ranges from that same
generation's text, preserving untouched bytes through JSON transport. The Node
result adapter creates owned wire fixes from the plugin's range and text plus locally
recorded source fragments; plugin fix objects and their extra metadata remain
outside the private format. Node
normalizes new replacement text before returning results,
replacing unpaired surrogates with U+FFFD as Node does when encoding a string
as UTF-8. All result transports carry the same Unicode text, without a
compiler-internal WTF-8 representation or a separate UTF-16 payload.
Fixes that split a scalar expand to its byte boundaries and include the retained
UTF-16 halves in their replacement text before encoding. Autofixes and
suggestions share this reconstruction; neither path retains transport bytes
or native syntax.

The plugin SDK's SourceCode queries the owned token columns through private
views. Individual and bounded queries construct only visited token objects;
comments use their own projected data and token offsets. A per-file indexed
cache preserves object identity and observes mutations of returned tokens.
The merged token/comment order is cached as numeric indices until a complete
array is requested. Program.tokens and SourceCode.tokensAndComments still
expose ordinary arrays containing those same mutable objects; after exposure,
queries observe the public arrays. These views retain no transport capability
or original frame buffer. Token value/range/location construction remains in
languages/js, while caching and ESLint query behavior belong to SourceCode.

The listener dispatcher preserves one specificity-sorted order for simple and
complex selectors. Only complex ESQuery matches prepare an ancestor array,
shared within the current enter or exit phase. Simple listeners do not copy
ancestor paths, including in configurations that also contain complex selectors.

`WithFrameAttachments` associates each file with exactly one binary attachment.
CLI and native API use it unconditionally; LSP serializes the same owned bytes
inline as base64 through its existing JSON-RPC route. The host resolves bytes
and forwards configuration keys; workers own configuration lookup. Neither the
plugin adapter nor Node IPC performs a separate canonical base64 validation
pass. Native mapping bounds, publication generations and exact release
acknowledgements remain transport invariants. API protocol 5.0.0 requires
matching peers, with no old text-input negotiation or parser fallback. Detached
LSP continuations contain owned frames, diagnostic text and cloned JSON
metadata, with no SourceFile, checker, mutable observer or native rule graph.
The generation lease ends before worker execution. Joined CLI/API tasks finish
within the original generation lease. There is no new transport package, AST
cache, plugin scheduler or Program authority.

Mapping descriptors, published batches and release acknowledgements belong to
the transport envelope, never to plugin requests or results. The private pool
and platform mapping files live in the existing IPC package. Allocation failure
or unavailable capacity preserves complete inline attachments. Attachments larger
than the mapping use owned storage and the same inline path. The JSON envelope
length must fit its uint32 header and the host runtime's buffer representation;
there is no separate application size quota. Inline binary attachments use
base64, so the Node runtime's string representation remains a natural limit.
A runtime representation failure closes the channel and rejects pending requests.
The transport does not implement an unbounded streaming
protocol. Closing the channel rejects new work, wakes pending calls and waits
for a current memory writer before releasing the mapping; mapped slices never
escape the storage implementation.

On Node, CLI and the resident API share `spawnIpcProcess` and `IpcClient` for
process startup, framing, request matching and reverse-request dispatch. The
API adapter owns only process lifetime, activity-based `ref`/`unref` and the
projection from response envelopes to API payloads, retaining generic attachment
context for inbound handlers. Native API sessions negotiate the same generic memory transport as CLI
sessions. Reverse config/plugin capabilities describe application handlers;
frame attachments are mandatory for plugin requests in this API version.
API request serialization and configuration generations remain in their
application owners.

`IpcClient` has one terminal cleanup path for explicit close, input EOF/close,
read errors and output errors/close/finish. It detaches its listeners, rejects
pending requests, clears retained buffers and handlers, closes native storage
once, and publishes the first terminal error through `done` even if native
cleanup fails. Request rejection and transport termination remain distinguishable
when their events arrive together. A stream fault does
not wait for the child process to exit. The API adapter terminates its child
when the channel closes, so a pending reverse handler cannot keep a broken
session alive. The process factory owns pipe error guards until child `close`,
covering queued write errors after the client has detached its own listeners.
The CLI allows a disconnected child the existing process-exit grace period,
preserving its natural exit code. A child that outlives that grace is terminated
and reported as a host failure.

For shared attachments, `spawnIpcProcess` prepares an empty anonymous Unix
descriptor before spawning Go; Windows needs no mapping at this point. Before
sending the first application request, `IpcClient` requests the storage configuration through the
same channel's `transportConfig` request. Go supplies the layout, and Rust
uses it to size and map the arena. The configuration request is handled inside
IPC and does not consume the application request's mapping bootstrap. Concurrent
initial requests share this one configuration exchange. `IpcClient` then sends
the initialized mapping in the first application request envelope and owns its
lifetime. A failed setup closes native resources and keeps complete attachments inline
for the session. Neither the CLI engine nor the plugin host chooses a
platform mapping or manages a lease. The client validates attachment ranges,
registers one native capability for all of the request's batches, dispatches
the application handler and revokes the capability before returning either a
result or an error. It acknowledges reuse only after every native reader returns.
Go matches the entire ordered batch set to that request's published storage.
Published storage remains channel-owned after a caller cancels: a later exact
acknowledgement may release it, but cannot revive the cancelled request. An
ordinary result, cancellation or timeout never authorizes reuse by itself.
If native readers outlive a handler, the result returns immediately without an
acknowledgement. `IpcClient` retains only a release function and request ID and
polls pending leases with one unreferenced 25 ms timer. When a lease becomes
reusable, a payload-free `transportRelease` frame carries its original request
ID and exact ordered batch set. Go retains that publication until the ACK or
channel close; the control frame never resolves an application request. Both
queues are bounded by occupied slots. Closing the session cancels the timer and
discards pending reclamation; outstanding native readers still pin their mapping.
The shared plugin host resolves file attachment indices before worker dispatch.
Its wire request type carries an inline frame or `frameAttachment` index;
the resolved type contains owned inline bytes or a private native capability.
Task construction accepts only the resolved type. Each attachment must be
binary, referenced once in order, and accompanied by no surplus attachments.

Linux uses a sealed anonymous memory file, macOS an immediately unlinked POSIX
shared-memory object, and Windows an anonymous pagefile mapping whose handle
Go duplicates. These are private backend details: attachment transfer never uses
filesystem paths or process addresses. Windows reserves the arena and commits
its control page initially; each data slot is committed on its first write.
Allocation or commitment failures, exhausted capacity and attachments larger
than the available budget retain their complete inline representation.
The arena is bounded to sixteen 16 MiB slots plus a 4 KiB control region.

An aligned 32-bit publication word per slot supplies the release/acquire memory
fence between Go and Rust. The native reader holds a writable control view for
atomic references and a separate read-only data view. `MemoryArena` registers
all batches atomically. A `SharedBytes` capability addresses their logical byte
space using a lease, offset and length; offsets count actual published bytes,
not slot capacity. A reader pins both views for every batch. The generic `readBytes` entry pins those views and copies all requested
chunks directly into one independent, Node-owned Buffer. JavaScript writes
cannot mutate a published snapshot, and Buffer lifetime does not delay slot
reuse. Rust contains no parser, ESTree schema, filename, BOM or plugin policy.

JavaScript language data lives in `packages/rslint/src/languages/js`:
ESTree types, the frame schema/decoder, source positions, visitor keys, token
reconstruction, JSX token text decoding, built-in globals and scope analyzers.
Visitor keys use the public TypeScript-ESLint data for JS, JSX and TS, independently
of the frame's supported node set. Unknown synthetic nodes use ESLint's public
key lookup for traversal and scope analysis. Compiler-synthesized JSDoc remains
comment data; unsupported types
authored in source fail projection instead of becoming null AST fields.
These modules consume syntax data and explicit language options; they import
no ESLint plugin runtime, native binding, IPC, config loader or worker host.
The dependency direction is `eslint-plugin → languages/js`. SourceCode and
RuleContext adapt that data to the ESLint API; diagnostics use language position
functions directly, rather than importing SourceCode or obtaining language
helpers through RuleContext. The plugin host, attachment adapter and WorkerPool
retain request routing, byte access and execution coordination. These language
modules serve JS plugins; they do not replace the native Go AST model. Go's
counterpart is `internal/languages/js/estree`, called only by the linter
plugin dispatcher for selected plugin targets. It reads existing syntax helpers
from `internal/utils`; native traversal and Go rules retain those helpers in
their original module and never depend on `languages/js`. Projection owns only
its output bytes, not the input AST. No language registry or provider abstraction
is introduced.

The worker TypeScript project includes the language modules. Library projects
exclude them from file discovery, while public plugin entry points continue to
export their types from the modules that own them. There are no old internal
path forwarding modules.

The worker checks cancellation before capability reads and checks the frame
header and column extents before allocating AST objects. The matching Go
producer owns record shapes, references and ranges. Node eagerly
materializes ordinary ESTree objects, ranges, loc, canonical parent links and
owned token columns. Schema-local constructors allocate each node's own present
fields together, including non-root nodes' final parent slots. Parents are
assigned after all fields are complete, in the existing traversal order.
Constructors are cached by present-field shape, with at most eight variants per
schema and a generic constructor for additional shapes. They retain schema
constants only. Nodes and arrays use their ordinary JavaScript prototypes,
with mutable own fields and absent optional fields. Materialization does not
adapt to changes to global prototypes. String offsets are indexed once,
and decoding uses one temporary byte view for the string table instead of a
Buffer wrapper per string. Tagged words use signed reads to avoid boxed unsigned
magnitudes while preserving their tag and payload bits. No record-validation
pass or graph-validation scratch is
allocated. No input view survives materialization. BigInt and RegExp values are
created in JavaScript. No mapped Buffer, lazy AST proxy or capability escapes to
plugin code; plugins may retain any AST node or SourceCode after lint returns.
Existing scope, listener,
SourceCode, report and fix execution consumes this final tree.
SourceCode's private comment-aware token index reserves the known stream size
and keeps the actual merge length, avoiding repeated backing-array growth.
Node tables, AST lists and sparse token caches use ordinary arrays.
Token objects, identity, public arrays, range mutations and filter callbacks
retain their existing behavior.
Listener dispatch uses one callback and a private reversed-ancestor buffer per
traversal. The synchronous selector matcher borrows that buffer; rule callbacks
receive only the ordinary AST node. Nested traversals have independent buffers,
and no AST or traversal buffer enters the worker's selector compilation cache.
The SDK's private native-copy read and decode call ends before SourceCode
construction and rule execution, so its temporary Buffer is no longer a stack
root during rule callbacks. Caller-owned inline frames keep their ownership.
Lazy scope resources capture language metadata and the owned AST, never the
transport request. Retaining SourceCode therefore does not retain an inline
frame Buffer or a shared-memory capability after decoding.
The decoder uses wire node IDs directly for parent assignment. Production
transport supplies an owned Buffer; mapped bytes are copied before decoding.
SourceCode's generic visitor fallback for caller-created nodes and its shebang
comment API remain SDK behavior; framed production syntax accepts only the fixed
schema.

Revocation removes the capability from the reader registry immediately. Active
readers keep the revoked lease and all its slots occupied. Reuse is acknowledged
only after the last reader returns, and only once; old capabilities remain
invalid after reuse. Worker termination may wait for a synchronous native byte
copy. Native conversion preserves pending exception or termination state,
without attempting a second exception. Native lease tracking contains no
request IDs, timers, wire messages or scheduling policy.

`internal/ipc/protocol.go` owns the storage settings: slot count, slot size,
control-region size and publication stride. They travel to Node and Rust at
runtime; native mapping, address calculations and range checks use that received
configuration. Capacity is derived from the layout instead of transmitted as
another independent setting. Rust validates the supported publication version,
integer bounds, alignment and arithmetic before mapping, and permits only one
configuration attempt per arena. The four-byte little-endian frame prefix, JSON
envelope and 32-bit publication words remain fixed wire conventions. There is no
protocol code generator or generated binding build step. Each language validates
inputs at its own boundary. Platform handles, native reader leases and
application scheduling remain implementation details of their owners.

The native layer owns its N-API input and capability types without importing
IPC. These value types live in `native/types.ts`, which has no Node dependencies;
browser and WASM protocol consumers do not import the Node binding loader.
The loader and its Buffer-returning API stay in `native/binding.ts`.
IPC owns the wire envelope and converts nullable native mapping
fields at that boundary. Frame types and decoding do not depend on the
IPC storage layout or its build inputs.

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
- Multi-config target discovery processes config roots in stable order. The
  catalog records admitted directory entries separately from config locations:
  requested directories retain their original bounds, while normally reached
  child configs receive only their admitted subtree. Target planning consumes
  those entries and explicit-file scopes in the bounded walker for each config;
  the automatic ownership index uses the same directory entries for config
  handoffs. Loading an ancestor config never expands a requested subdirectory
  or takes files still reached through an enclosing owner. The index omits
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
  with no lexical config. Frozen physical parent ancestry takes precedence over
  the leaf file's destination, matching staged discovery without filesystem reads.
- `Rslint.lintFiles()` sends raw inputs to Go; it does not enumerate, exclude,
  canonicalize, or assign config owners to targets in JavaScript. Go freezes
  identities and rejects conflicting physical ownership before lint execution.
  The exact-file API still accepts independent canonical hints, including for
  `lintText()` virtual buffers. JavaScript resolves source aliases only when
  required to assemble multi-edit fixes or suggestions; that memo stays local
  to the call. Pattern inputs are combined as a union rather than interpreted as ordered
  exclusions; config `ignores` remains the exclusion mechanism.
- LSP uses a different orchestration model and keeps document and Program access on its main
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
goroutines remain outside that guarantee. CLI plugin initialization uses one
JS worker and can overlap the serial Go target planning and Program loading;
lint and fix execution still await full config activation.

## 10. Performance & Memory Considerations

### Design Principles

- **Direct ts-go Data Model**: rslint Programs preserve ts-go AST objects directly and provide optional checker access without converting through a second AST representation or exposing their private adapter
- **Unified Program Scheduling**: `RunPipeline` acquires complete Program bindings through `GenerationPlan`. Independent bindings enter the existing native shard budget before construction and retain that reservation through execution and release. Whole-binding consumers use the same executor and lease boundary; their configured Program construction keeps its bounded worker pool. Each Program preserves its checker parallelism and complete dependency graph. `--singleThreaded` keeps workload stages serial
- **Single-Walk Rule Dispatch**: each file is traversed once, with rules registering listeners up front and sharing the same AST walk
- **Early Filtering**: exact lint target plans, skip paths, global-ignore filters, and per-file capability filtering reduce work before listeners run
- **Shared Prepared Lint Plan**: CLI and native API resolve each selected
  file's complete rule set once into immutable rslint Program slots. Native
  execution and optional third-party plugin dispatch consume projections of the
  same plan instead of repeating target collection or rule resolution
- **Shared Prepared Rule Sets**: plan preparation owns rule eligibility and the
  ordered native/plugin projections together. Files with the same immutable
  input slice and checker capability share one completely prepared rule set;
  native execution and plugin input construction consume its views directly.
  A preparation-local lookup uses the input's exact start, length and capability,
  and is discarded after workers join. File contexts, checker instances, language
  defaults and plugin routing remain per-file. Each new plan, including autofix
  observations, prepares independent sets

### Performance Optimizations

- **Native Go Implementation**: Eliminates JavaScript runtime overhead
- **Direct TypeScript AST**: No AST conversion between parsers
- **Shared Type Checker**: native lint workers reuse a checker across files and rules within each checker shard of a `Program`
- **Checker Phase Separation**: all native workers for a Program join and release their checkers before its TypeScript diagnostics acquire checkers. Under `ProgramComplete`, those diagnostics may overlap native lint work for other Programs
- **File Filtering**: Skip node_modules and bundled files automatically
- **Source-Only Programs**: project-unbound CLI roots use the ts-go parser, binder, package/module metadata, and direct-import resolution to construct the same rslint Program facade without a synthetic project graph or checker; capability gating excludes type-aware rules
- **Shared Cross-File Rule Structures**: module edges, export maps, and whole-Program indexes are immutable derived values cached by Program generation. The CLI/API loader adapts each freshly built project set once; its Program facades own these caches directly, with duplicate raw inputs sharing within that set. Loading or binding the same set preserves its facade pointers and cached answers. Raw generations reused across temporary facade lifetimes, including LSP projects, retain their generation cache through the weak-key registry. Root-parsed Programs also own their cache directly. Cache ownership belongs to Program assembly; rules use the same cache contract and never select a lifetime or clear a live cache
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
  entry-local ignores once, accepts a compatible discovery match or evaluates
  the complete flat-config entry bitset, and merges/prepares enabled rules once
  per unique bitset. The prepared lint plan references these published immutable rule
  slices without moving per-file runtime state into the config cache. LSP keeps
  its existing per-document resolver lifetime.
- **Conservative Rule Queries**: import rules share a syntax-only receiver-name
  superset through the per-file cache. The Program/settings-owned module index
  can rule out absent export names or namespace metadata without building a
  merged map; possible names and unknown edges still use the complete export
  implementation. Query traversal tracks file/name pairs through re-export
  cycles. React DOM property lookups share four immutable version bands;
  Promise argument checks retain checker authority even for primitive literals,
  and rename-state decisions remain part of diagnostic evaluation.

### Caching Strategy

- **LSP Program Ownership**: `internal/lsp` retains only configured Programs selected by open, lint-admitted documents. The last selecting document closing or losing its selection releases the Program. The lint-only server does not create a general TypeScript language-service Session or its inferred projects, ancestor graphs, or auto-import index. All open texts remain available to selected Programs as dependency overlays, regardless of whether those documents are linted.
- **LSP Checker Ownership**: each resident Program uses one compiler checker for the server's serialized file requests. This execution policy is applied to a copied parsed configuration and is independent of parallel project loading; authored configuration metadata remains unchanged.
- **LSP Incremental Source Lookup**: when the compiler reports graph reuse, the existing canonical path index is rebound to the updated Program's source objects. A changed graph builds a new index. Text edits do not repeat the physical source path scan for every unrelated project.
- **LSP Configured-Project Selection Cache**: diagnostics and speculative fixes share the metadata-first project selector while keeping separate Program owners. Only a selected project containing the requested root is constructed. The resident store covers config, directory, dependency, and missing-config discovery reads with existing watchers before retaining state; new discovery coverage requests reselection across the registration interval. Config and filesystem changes invalidate metadata and Programs. Registration failure and conflicting editor aliases use fresh request-local construction. Saving a dependency and receiving a changed-open-file notification both refresh open-document diagnostics independently of Program invalidation. Speculative fix Programs always remain isolated.
- **Parse Cache in LSP**: resident Programs share the existing full-key, content-addressed `utils.ParseCache` through `WithASTCache`. Every source read uses the current editor overlay; LSP does not bind immutable source snapshots to its mutable host. Generation release and document close prune parsed ASTs to the current resident Programs, so old content versions and closed project graphs are not retained by the cache. Configured Program construction uses the compiler's parallel loader. Disk and projected-overlay query caches belong to one observation and are replaced on each request and watcher-registration retry; they never carry cached misses across editor or filesystem changes.
- **SourceFile-Owned Module Syntax Reuse**: any `Program.ModuleGraph()` query may attach the syntax-only module-specifier projection to an immutable ts-go `SourceFile`, independently of project ownership or watcher availability. Programs reusing that exact file object share collection, while resolution and target ASTs remain local to each Program generation. Attached values may contain only scalars and nodes owned by that SourceFile—never a Program, resolved target, checker state, or another SourceFile. Replaced files carry the attached data out with their own AST lifetime; no LSP server map, project-membership sweep, or explicit reset owns this cache. Source-only Programs retain it only for their generation's SourceFile lifetime.
- **SourceFile-Owned Export Syntax**: the import module index similarly shares
  immutable export-step templates and authored documentation nodes attached to
  the exact SourceFile. Only import/export statements occupy template slots.
  Each Program/settings index binds fresh resolution links into its own steps;
  TypeScript namespace assignments and interop policy stay Program-local.
  Shared template slices and maps are never mutated by consumers.
- **Module-Local Export Demand**: the import module index collects each file's
  declared names and possible namespace bindings once. Name queries consult
  that local summary without traversing re-export dependencies. Star exports
  and graph-dependent namespace metadata go through the existing full
  ExportMap, preserving its ordering, cycles, interop and documentation rules.
  An already cached full map bypasses the filter. Wide barrels and queries
  from different graph roots therefore do not repeat per-name graph walks.
- **SourceFile-Owned React Pragmas**: the first `@jsx` annotation, including
  its absence, is cached on the exact SourceFile. Hashbang and comment order
  retain their existing precedence. Only the annotation name and presence
  are shared; settings fallback and the final pragma remain context-local.
- **Bounded Frozen Path Projection**: a frozen IO view memoizes at most 4,096
  exact request spellings to projected paths and frozen-file membership. After
  capacity is reached, new spellings use the original computation. This cache
  owns copies of retained strings so path substrings cannot pin source buffers.
  It lives only with that view and contains no Programs or filesystem answers:
  existence, contents, metadata and non-frozen Realpath results still query the
  underlying filesystem. Runtime module resolution separately shares results
  by directory only when aliases, fallbacks, alias fields and main fields are
  absent; recursive options retain the original containing-file key so cycle
  diagnostics keep their observable identity.
- **Incremental LSP Document Sync**: editor changes reach the server immediately and are applied in order to rslint's document store; compiler hosts read frozen views of all open texts. rslint selects the mandatory LSP UTF-16 encoding so incoming changes, native diagnostics, plugin diagnostics, and edits share VS Code's coordinate model. Whole-document changes remain a supported protocol fallback.
- **Server-Owned Debounced Re-linting**: `refreshCh` and `debounceCh` collapse bursts of file changes and project refreshes onto the main dispatch loop. The server remains the single owner of the 200 ms typing debounce; open and save diagnostics stay immediate, while save, fix-all, and close discard redundant pending work for their target document.
- **CLI/API Are Mostly Fresh Runs**: CLI and one-shot API requests generally rebuild `Program` state per run; there is no repository-local rule-result cache or persistent incremental lint cache in the main CLI path today. JavaScript API source-alias lookup is also scoped to one call.
- **Report Root Identities Are Demand-Scoped**: only the default CLI format asks the loader to freeze roots from the actual type-capable Programs selected for the run, during initial acquisition and before native execution. The loader reuses target hints and resolves only the remaining root paths through one bounded per-file identity pass; the command immediately reduces the frozen result to a file count. Machine formats, autofix generations, `Program` caches, and report rendering retain none of this identity state. As with the rest of a loader Session generation, this assumes the filesystem namespace is stable for the run; the shared request-local VFS may be warmed by the projection, and concurrent external namespace mutation is not a supported consistency boundary.
- **Worker-Local JavaScript Selector Compilation**: each third-party-plugin worker keeps a 2,048-entry FIFO cache keyed by the exact non-bare selector string. Entries contain only the immutable canonical esquery parse result, specificity, conservative anchor types, or an invalid-selector message; per-file contexts, listeners, options, ASTs, source code, and match results remain outside the cache. Bare node-type selectors retain their direct dispatch path, and replacing the worker naturally discards the cache.
- **Parallel Program Realpath Queries**: when a CLI or lint API request builds multiple Programs concurrently, the loader session derives a Program-only VFS view that coalesces same-path `Realpath` cold queries across those compiler hosts. Completed realpath and empty-result values remain request-local. Exact path strings are used as keys without cleaning, separator rewriting, case folding, or realpath-key merging. Existence checks and all other operations continue through the existing VFS stack. Serial Program builds, config discovery, target binding, LSP, and `cmd/tsgo` retain their existing filesystem paths and cache lifecycles.
- **Resolver-Scoped Effective Config Plans**: each `FileConfigResolver` owns
  exact-path file entries and exact matched-entry shape entries. Both caches use
  publish-once values so concurrent readers of the same resolver wait for
  complete initialization. They are discarded with the resolver. CLI/API
  autofix generations within one request reuse immutable config results while
  replacing source mappings; separate requests get new resolvers. LSP reuses
  resolvers owned by its committed config catalog until a config refresh
  replaces them. Program and filesystem snapshots retain their separate
  generation lifetimes. Merged config maps and configured-rule slices returned
  by the resolver are immutable shared state.
- **Program Loader Generation Boundary**: the initial CLI/API observation and each changed autofix observation own a `program/loader.Session`, created only after that generation's overlay/canonical VFS wrappers are complete. A session privately composes compiler construction, source snapshots, metadata caches, project loading, and target binding for one immutable filesystem view. It is never global, never shared across API requests, and is not used by LSP, whose resident Program store has a different invalidation model.
- **Run/Request-Scoped Program Metadata**: successful `package.json` reads and explicitly registered root, project-reference, and extended tsconfig reads are snapshotted for the context lifetime. Keys remain exact caller paths—no cleaning, case folding, resolving real paths, or symlink merging—and failed reads are retried. Per-key read single-flight avoids duplicate concurrent I/O; generation swaps make future VFS writes safe without clearing a live map. Arbitrary JSON and non-metadata reads bypass this layer.
- **Extended Config Parse Reuse**: the context implements ts-go's `ExtendedConfigCache` shim contract and shares common `extends` parse results across Programs. Parsing occurs outside map locks and publishes with `LoadOrStore`, avoiding recursive cross-cycle lock ordering; rare concurrent misses may duplicate parsing but share the winning immutable result and still single-flight raw bytes. Root `ParsedCommandLine` values and parsed `package.json` objects are not cached.
- **Generation-Scoped Source Snapshots**: Programs built for one CLI/API observation share immutable source text/hash snapshots across project and transient root-parser hosts. Keys are the exact compiler-host source names, never real paths, so lexical, overlay, and symlink aliases remain distinct. Concurrent misses for one key share the successful read/hash operation; failed reads are shared only by the overlapping callers and are not retained. The next changed autofix observation receives a fresh session bound to its new overlay, so no source snapshot crosses generation boundaries and no bound AST is published across rslint Programs.
- **Core-Owned Fix Generations**: after each successful fix round, `internal/linter` advances only its private source memory. CLI/API generation providers project that snapshot onto a fresh request-local overlay and rebuild Programs without mutating the base medium; CLI alone commits the final net delta to disk once. LSP speculative generations remain isolated from its document and resident Program state, which still advances only through versioned document events.
- **Generation-Scoped Parse Reuse**: all Programs within one CLI/API observation share that loader session's content-keyed AST parse cache. Concurrent misses for the same full parse key are single-flight, so multi-Program construction does not duplicate parsing. A changed autofix observation starts a fresh cache for its new source generation; caches are never repository-persistent or shared across lint requests.
- **Construction-Cache Ownership**: CLI/API whole-binding acquisition calls `Session.ReleaseBuildCache()` after binding to drop construction AST and Program indexes. Published Programs and frozen source/metadata snapshots remain valid. Independent loads use weak AST references so the request cache cannot retain completed Programs’ sources. A later load may reparse the same immutable text. Fix observations use fresh Sessions, while LSP borrows its document-owned resident Programs through the same core lease boundary without evicting the editor cache.
- **Bounded Multi-Pass Fixing**: CLI `--fix`, API `fix: true`, and LSP `fixAll` share the linter-owned limit of ten writable rounds. CLI and API request a final observation after the tenth write so their diagnostics describe the returned source; LSP consumes only the net text delta and does not pay for that extra observation.

### Memory Management

- **ts-go Owns the Heavy Graphs**: AST nodes, checker state, compiler project graphs, and session state are primarily owned by ts-go; an rslint Program adds a small immutable facade and generation cache, while listeners, diagnostics, lint selection, and configuration stay outside that object
- **Short-Lived Per-File Structures**: comment stores, disable managers, and rule contexts are allocated per file and dropped after traversal. A comment slice is allocated only if requested
- **Bounded Listener Retention**: a listener registry lives only for one checker-shard task. After each file it clears every function slot before shortening the slices, so backing capacity can be reused without retaining closures, source files, checker state, or rule contexts. The registry is dropped when that task completes and is never pooled across runs or LSP requests
- **Source Snapshot Ownership**: snapshot entries hold an immutable source string plus its 128-bit hash without explicitly copying source bytes; on an AST miss, that string is passed directly to the parser. After generation replacement, a retained unchanged AST may still hold the prior equal string while the fresh snapshot owns the new read. Replaced generations are reclaimed after any in-flight lookup releases them. AST retention and source-generation retention remain deliberately separate lifecycles.
- **Execution Task Ownership**: CLI/API providers transfer their initial generation references during acquisition. After complete binding, syntax, rule, target, and plugin preparation, the pipeline retains private consumable project tasks rather than the producer generation or immutable `LintPlan`. Dispatch removes each pending task slot; completion drops only that task's Program/file/rule references. The reusable `RunLinter` API copies private task containers from its borrowed plan and uses the same executor. Program objects, shared slices, maps, and ASTs are never reset or invalidated.
- **Program Execution Boundaries**: non-mutating batch CLI lint explicitly opts into `ProgramComplete`: each Program runs its NoEmit diagnostics after its own native workers join, while other Programs may still be linting. Type diagnostics are buffered until all native reports finish, with cross-project deduplication following stable Program input order. `ProgramPhases` remains the compatibility default for direct `RunLinter`, API, autofix, and LSP requests: all native tasks join before Program diagnostics start. Complete-Program execution is rejected for mutating and progressive requests. Type-check-only keeps its independent scheduling path.
- **Last Consumer Barriers**: a complete-Program task retains admission capacity for pending diagnostics until the task finishes; phased execution retains a second Program borrow through the post-native type-check phase. Native project/file workers join before an abnormal exit is re-raised. Native failure cancels and joins concurrent plugin work before release; a handled plugin batch failure still lets other batches finish and preserves their diagnostics. Producer finalization remains exact-once and runs after its actual generation borrowers finish, including LSP watcher coverage checks. A completed finalizer no longer retains its release closure.
- **Diagnostic and Plugin Ownership**: observations that do not plan source changes project diagnostics to text before aggregation. One lightweight frame is attached to each exact SourceFile without a back-reference or global strong AST-key map, preserving shared-source identity across native and joined plugin producers. Plugin wire tasks retain only rule identity, options, severity and required source text; native Run closures and environments are excluded. Detached fix-source candidates exist only when the observation plans fixes and requests plugin autofixes. Fix planning preserves original source identity through text freezing; retained initial/final observations own diagnostic text and any explicitly requested AST artifacts. Source snapshots keep their generation lifetime independently of execution tasks; independent-load parse caches retain ASTs only weakly.
- **Metadata Snapshot Ownership**: metadata strings and extended-config parse entries live only for one loader session. The cache stores successful reads only, and its scope bounds growth to metadata touched by one CLI invocation or API request; no metadata entry survives into another request or the LSP process.
- **Fix Application Uses Linear Rebuilds**: `ApplyRuleFixes` sorts fixes, skips overlapping edits, and rebuilds the output with `strings.Builder` rather than mutating source buffers in place
- **Bounded Queues**: CLI diagnostics use a buffered channel of 4096 items; LSP request/outgoing queues are buffered to 100, and debounce/refresh signals are single-slot channels
- **Bounded RegExp Retention**: the regexp cache retains at most 256 compilation results, including errors, in FIFO order. Sources plus flags over 4 KiB or successful rewrites over 64 KiB bypass retention. Cached strings are copied so a small pattern cannot retain a source file's backing storage. Idle matching backends live in per-pattern `sync.Pool` instances and can be discarded by GC; concurrent demand or collection can require recompiling the validated backend pattern. Live handles remain valid after eviction. The cache retains no AST, rule context, options object or linted subject; pool size follows concurrent demand rather than a fixed heap bound. Backends exposed by `Unwrap` belong to their handle and never return to a shared pool.
- **Fresh ESM Entry Lifetime**: fresh JS/TS config loads use a unique entry-module URL so rewritten bytes and module side effects are evaluated per transaction. Node retains those ESM module namespaces for the process lifetime, so a long-lived native API process can grow this cache slowly across repeated lint requests; static transitive imports continue to use Node's ordinary cache. Bounding this without weakening freshness requires a disposable evaluator realm or worker and remains a future optimization.
- **Garbage Collection Handles Cycles**: the repository does not implement custom cycle breaking for AST/checker graphs; lifecycle cleanup relies on Go GC and on dropping references after each run

## 11. Extensibility & Future Directions

### Plugin Architecture

Plugin execution has two paths:

- bundled Go plugins execute through the shared listener traversal
- third-party ESLint plugin objects are loaded from JS/TS config on the Node side; Go derives catalog placeholders for their rules and sends per-file batches back to the Node plugin worker over reverse IPC

String plugin declarations select bundled Go plugin namespaces. Live third-party plugin objects require a JS/TS module config. The repository currently ships Go implementations for TypeScript ESLint, Effector, Import, Jest, JSX accessibility, Node.js (`node`), Promise, React, React Hooks, Rstest, and Unicorn rule namespaces.

### Rule Extension Points

- **Core Rules**: add a package under `internal/rules/<rule_name>/` and append the rule var to `internal/rules/all.go`'s core rule slice
- **Go Plugin Rules**: add a package under `internal/plugins/<plugin>/rules/<rule_name>/` and append the rule var to that plugin's `all.go`; a new bundled plugin also needs one aggregation entry in `internal/rules/all.go` and any accepted string declaration aliases in `internal/config/plugin_declarations.go`
- **Third-Party Plugin Rules**: import a plugin object in JS/TS config and mount it under an object-form `plugins` prefix; no Go rule implementation or bundled aggregation entry is required
- **Rule Options**: each rule receives parsed options through `Run(ctx, options)`
- **Custom Listener Shapes**: rules can listen on standard kinds and synthetic pattern/exit kinds

### Integration Points

- **Language Server**: `internal/lsp` exposes diagnostics and code actions
- **JavaScript API**: `packages/rslint` talks to the `internal/api/server` handler composed by `cmd/rslint --api` through the versioned `3.1.0` protocol; the handshake negotiates reverse `pluginLint` support before third-party rules run
- **WASM Playground**: `packages/rslint-wasm` runs the API server in a browser worker
- **Rust Client**: `crates/tsgo-client` consumes `cmd/tsgo`

`cmd/tsgo` dispatches two one-shot subcommands, each parsing its own options.
`project --api --config <path>` returns the existing binary project information;
the Rust client selects this command, and flag-only invocations remain compatible.
`config --config <path>` resolves tsconfig inheritance and prints
`FinalCompilerOptions` directly as JSON for JavaScript consumers to `JSON.parse`.
This command includes defaults and implied values for the exported option subset,
without creating a Program or checking source code. Parsing errors go to stderr
with a nonzero exit code. Both commands default to `tsconfig.json` in the working
directory and use the same pinned compiler without its JavaScript API.

### Future Enhancements

The current architecture already leaves room for:

- broader rule coverage
- richer editor features on top of the existing LSP foundation
- more shared tooling between CLI, API, LSP, and Playground

## 12. Testing Strategy

### Test Organization

- **Go Unit Tests**: colocated `*_test.go` files under `internal/...`
- **Rule Engine Tests**: `internal/linter` and `internal/config` have focused behavior tests
- **Go Rule Testing Helpers**: `internal/rule_tester`
- **Shared Go Test Infrastructure**: `internal/testutil`; package-specific helpers remain beside the owning tests
- **Core JS/TS and API Tests**: `packages/rslint/tests`; CLI engine integration
  selects its own built worker locally, without shared test setup.
- **Cross-Language and Ecosystem Tests**: `packages/rslint-test-tools/tests` owns
  the Go ESTree producer, upstream AST/token/SourceCode comparisons, and real
  plugin worker/package integration. Its `go-frame` project builds the producer
  only for source-to-frame tests. Reference parsers and ESLint are test-tools
  dependencies; core keeps the libraries bundled into its runtime.
- **VS Code Extension Tests**: `packages/vscode-extension/__tests__`
- **Rust / tsgo Tests**: `crates/tsgo-client/tests` verifies exported semantic data through the matching `cmd/tsgo` binary

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

### Local Verification and Continuous Integration

Local verification uses existing Go, pnpm/Rstack and Cargo commands with explicit package or test-file arguments. See [CONTRIBUTING.md](./CONTRIBUTING.md#verify-a-change) for examples and [AGENTS.md](./AGENTS.md#local-verification) for scope and result-reuse rules. Test layout and semantic coverage remain part of reviewing the change.

`.github/workflows/ci.yml` owns the full Go, JS, Rust, WASM and integration checks. Its command list does not expand local test scope.

### Release Lines

- **rslint packages**: `scripts/version.mjs` and `.github/workflows/release.yml` own the unified rslint npm version. They explicitly exclude the tsgo server packages.
- **Release assembly**: the release workflow resolves the requested branch once
  and gives the same commit SHA to every Go, WASM, native addon, JavaScript and
  extension job. Artifact assembly requires one nonempty input per expected Go
  target and native addon tuple, plus the rule schema dump, before writing any
  platform package output. Existing local binaries cannot conceal missing inputs.
  Linux Go release builds disable cgo so the same executable can serve both gnu
  and musl packages; native addons retain their separate libc targets.
- **tsgo distribution**: `pnpm version:crates` updates `tsgo-client`, `Cargo.lock`, `@rslint/tsgo-server`, and all six platform packages to one version. `.github/workflows/release-crates.yml` assembles the platform packages and publishes npm and crates.io artifacts together.

## 13. Adding a New Rule (Checklist)

The maintained rule-porting workflow now lives under [`.agents/skills/port-rule`](./.agents/skills/port-rule/).

Use these entry points instead of duplicating a separate checklist here:

- [`.agents/skills/port-rule/SKILL.md`](./.agents/skills/port-rule/SKILL.md): new-rule workflow and reference routing
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
  → generation, transport, presentation, and optional commit adapters
  → RunPipeline:
      plan and admit complete binding acquisition
      → acquire or borrow each binding
      → prepare LintPlan and native/plugin work
      → consume private execution tasks, join borrowers, and release the lease
      → finish the observation, in-memory fix rounds, and optional terminal commit
  → integration result projection
```

Downstream stages may validate or project an upstream decision, but do not
repeat it under a second source of truth.

- **Targets and config**: target selection and config ownership are frozen before Program binding (`target.Plan` for CLI/API and a document snapshot for LSP). Later stages do not add lint targets, rediscover configs, or reassign owners
- **Program generation**: each published `program.Program` is one logically immutable source, module-resolution, filesystem, and optional-checker generation. CLI/API build it through `internal/program/loader`; LSP adapts its selected resident Program or isolated overlay without exposing the private backend
- **Lint plan**: `PrepareLintPlan` accepts only files already bound to those Programs and freezes each file's rules, shared environment, and checker eligibility. Execution does not scan Program roots or resolve config and rules again
- **Pipeline**: `RunPipeline` is the production orchestration boundary. CLI, API, and LSP provide ordered loading units, plugin transport, presentation, and optional commit adapters. Each unit owns or borrows a complete binding; rule backends do not select different loading paths. The core admits loading, prepares native and plugin work from the same `LintPlan`, joins its borrowers, releases the lease, and combines unit results into one observation. Plugin projection and transport windows share one observation-wide capacity and buffer pool. Plugin failure policy applies once to the combined result, including notices and cancellation records. Execution does not retain the preparation plan or repeat upstream binding decisions
- **Binding scope**: the loader can expose independently bound complete Programs when root metadata establishes ownership. Source aliases, compiler-rejected roots, and source remapping retain coherent binding through the same executor, without retrying already executed rules. Fix generations retain one coherent source-identity boundary for validation; LSP borrows its owner's generation. These ownership and result requirements do not depend on whether rules run in Go or JavaScript
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
│            ├───────────────► Pure explicit --type-check-only                 │
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
│  Shared File Config Resolver: matched service/root/reset options             │
│            │                                                                 │
│            ▼                                                                 │
│  Program Loader: declared/service candidates, build, then source binding      │
│            │                                                                 │
│            ▼                                                                 │
│  Unified Program Sequence + Target / Path Projections                        │
│            │                                                                 │
│            ▼                                                                 │
│  Reuse the Same Immutable File Config and Enabled Rules                      │
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
│  internal/lsp + document-owned lintProgramStore                                        │
│     │                                                                        │
│     ▼                                                                        │
│  Frozen Document Snapshot + Selected / Overlay Program Generation             │
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
│     ├──── supported source import ────► esm.sh declaration graph             │
│     │                                      │                                 │
│     │                                      ├────► Monaco TypeScript libs     │
│     │                                      └────► lint fileContents + paths  │
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

Playground JavaScript configs are imported as modules in a Blob worker created inside an iframe with an opaque origin and `sandbox="allow-scripts"`. The iframe only relays messages; the worker inherits a Content Security Policy that limits scripts and connections to package CDNs and has no DOM or navigation APIs. The page validates the serialized config before linting and removes the iframe on completion or timeout.

Playground source dependency types are allowlisted by package and fetched from esm.sh only after a matching static import or export appears. The resolved declaration graph is cached for the page session, mounted into Monaco, and copied into the lint request's `fileContents`; request-local `compilerOptions.paths` entries resolve package specifiers to those virtual declarations without modifying the visible tsconfig. When declarations are present, `files` keeps the selected source file as the only lint target, so dependency declarations contribute types but are not linted themselves.

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
- **Project Set**: Loader-owned, stable set of configured ts-go generations with ordinary owner declaration order and sparse per-target candidates. Explicit and service construction modes remain isolated, even for the same tsconfig path
- **LSP Program Store**: LSP-owned cache of selected compiler Programs, retained by open documents and invalidated through editor events and watcher coverage
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
