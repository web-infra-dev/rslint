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
┌─────────────────────────────────────────────────────────────────┐
│                         RSLINT SYSTEM                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐  ┌──────────────┐  ┌─────────────────────────┐ │
│  │    CLI      │  │     API      │  │        LSP Server       │ │
│  │  (cmd/rslint)│  │  (internal/  │  │    (internal/lsp)       │ │
│  │             │  │    api)      │  │                         │ │
│  └─────┬───────┘  └──────┬───────┘  └──────────┬──────────────┘ │
│        │                 │                     │                │
│        └─────────────────┼─────────────────────┘                │
│                          │                                      │
│  ┌─────────────────────┬─▼──────────────────────────────────────┐ │
│  │                     │            LINTER CORE                │ │
│  │  ┌─────────────────┐│      (internal/linter)               │ │
│  │  │   CONFIG        ││                                      │ │
│  │  │  LOADER         ││  ┌──────────────┐ ┌─────────────────┐│ │
│  │  │(internal/config)││  │   PROJECT    │ │   RULE ENGINE   ││ │
│  │  │                 ││  │  DISCOVERY   │ │  (internal/rule) ││ │
│  │  └─────────────────┘│  │              │ │                 ││ │
│  │                     │  └──────────────┘ └─────────────────┘│ │
│  └─────────────────────┴───────────────────────────────────────┘ │
│                          │                                      │
│  ┌─────────────────────┬─▼──────────────────────────────────────┐ │
│  │                     │         TYPESCRIPT-GO                │ │
│  │  ┌─────────────────┐│     (typescript-go submodule)        │ │
│  │  │     RULES       ││                                      │ │
│  │  │ (internal/rules)││  ┌──────────┐ ┌──────────┐ ┌─────────┐│ │
│  │  │                 ││  │  LEXER   │ │  PARSER  │ │ CHECKER ││ │
│  │  │ - no_unused_vars││  │          │ │          │ │         ││ │
│  │  │ - array_type    ││  │          │ │          │ │         ││ │
│  │  │ - await_thenable││  │          │ │          │ │         ││ │
│  │  │ - ...           ││  └──────────┘ └──────────┘ └─────────┘│ │
│  │  └─────────────────┘│                                      │ │
│  └─────────────────────┴───────────────────────────────────────┘ │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │                   NODE.JS PACKAGES                         │ │
│  │                                                             │ │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────┐ │ │
│  │  │   @rslint   │ │ VS Code     │ │    Testing Tools        │ │ │
│  │  │   /core     │ │ Extension   │ │                         │ │ │
│  │  │             │ │             │ │ - rslint-test-tools     │ │ │
│  │  │             │ │             │ │ - rule-tester           │ │ │
│  │  └─────────────┘ └─────────────┘ └─────────────────────────┘ │ │
│  └─────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## 3. Directory / Crate Structure

| Path                          | Purpose                                                     |
| ----------------------------- | ----------------------------------------------------------- |
| `cmd/rslint/`                 | Main Go binary entry point with CLI, API, and LSP modes     |
| `internal/api/`               | API layer for JavaScript integration                        |
| `internal/config/`            | Configuration loading, parsing, and rule registry           |
| `internal/linter/`            | Core linting engine with file processing and rule execution |
| `internal/lsp/`               | Language Server Protocol implementation                     |
| `internal/rule/`              | Rule framework, context, diagnostics, and disable manager   |
| `internal/rule_tester/`       | Rule testing utilities and frameworks                       |
| `internal/rules/`             | Individual lint rule implementations (50+ rules)            |
| `internal/utils/`             | Shared utilities for string manipulation, file operations   |
| `packages/rslint/`            | Main npm package with JavaScript API and CLI wrapper        |
| `packages/rslint-test-tools/` | Testing utilities for rule development                      |
| `packages/rule-tester/`       | Rule testing framework for JavaScript                       |
| `packages/utils/`             | Shared JavaScript utilities                                 |
| `packages/vscode-extension/`  | VS Code extension for IDE integration                       |
| `typescript-go/`              | Git submodule containing TypeScript compiler Go port        |
| `shim/`                       | **TODO**: Go bindings and shims for TypeScript compiler API |

## 4. Parsing Pipeline

The parsing pipeline follows this flow:

```
Source Text → Lexer → Tokens → Parser → AST → Semantic Analysis → Rule Execution → Diagnostics
```

### Detailed Pipeline Steps

1. **Source Text Loading**: Files are discovered and loaded based on TypeScript project configuration
2. **Lexical Analysis**: typescript-go lexer tokenizes source text into tokens
3. **Syntax Parsing**: typescript-go parser constructs AST from tokens
4. **Semantic Analysis**: TypeScript type checker builds semantic model with type information
5. **Rule Registration**: Rules register AST node listeners for specific node kinds
6. **AST Traversal**: Linter traverses AST, invoking registered listeners
7. **Rule Execution**: Each rule analyzes nodes and reports diagnostics
8. **Diagnostic Collection**: Diagnostics are collected with severity, location, and fix suggestions
9. **Output Generation**: Results are formatted and reported to CLI, API, or LSP client

### Error Recovery Strategy

The TypeScript parser includes robust error recovery that:

- Continues parsing after syntax errors
- Produces partial AST for incomplete files
- Maintains accurate source locations even with errors
- **TODO**: Document specific error recovery patterns used by typescript-go

## 5. Abstract Syntax Tree (AST)

### AST Representation

The AST uses the TypeScript compiler's native node representation through typescript-go:

- **Node Types**: All TypeScript AST node kinds (ast.Kind enumeration)
- **Node Structure**: Parent-child relationships with bidirectional navigation
- **Source Locations**: Precise start/end positions and spans
- **Node IDs**: **TODO**: Document node ID system if present
- **Memory Layout**: **TODO**: Document memory representation and interning strategy

### Key AST Properties

```go
type Node struct {
    Kind     ast.Kind      // Node type (e.g., ast.KindFunctionDeclaration)
    Parent   *ast.Node     // Parent node reference
    Children []*ast.Node   // Child nodes
    Pos()    int          // Start position in source
    End()    int          // End position in source
    Text()   string       // Source text for this node
}
```

### Span and Location Handling

- **Positions**: 0-based character offsets from file start
- **Ranges**: Start and end positions defining node boundaries
- **Line/Column**: **TODO**: Document line/column calculation utilities
- **Source Maps**: **TODO**: Document source map support if available

## 6. Lint Rule Framework

### Rule Interface

Rules implement a standardized interface defined in `internal/rule/rule.go`:

```go
type Rule struct {
    Name string
    Run  func(ctx RuleContext, options any) RuleListeners
}

type RuleListeners map[ast.Kind](func(node *ast.Node))
```

### Rule Context

The `RuleContext` provides rules with access to:

```go
type RuleContext struct {
    SourceFile     *ast.SourceFile
    Program        *compiler.Program
    TypeChecker    *checker.TypeChecker
    DisableManager *DisableManager
    ReportRange    func(textRange core.TextRange, msg RuleMessage)
    ReportNode     func(node *ast.Node, msg RuleMessage)
    // ... additional reporting functions
}
```

### Listener Registration

Rules register listeners for specific AST node types:

```go
func (rule *MyRule) Run(ctx RuleContext, options any) RuleListeners {
    return RuleListeners{
        ast.KindFunctionDeclaration: func(node *ast.Node) {
            // Analyze function declaration
        },
        ast.KindVariableDeclaration: func(node *ast.Node) {
            // Analyze variable declaration
        },
    }
}
```

### Listener Types

- **OnEnter**: Called when entering AST node during traversal
- **OnExit**: Called when exiting AST node (using `ListenerOnExit(kind)`)
- **OnAllowPattern**: **TODO**: Document pattern-based listeners
- **OnNotAllowPattern**: **TODO**: Document negative pattern listeners

## 7. Diagnostics & Autofixes

### Diagnostic Structure

```go
type RuleDiagnostic struct {
    RuleName    string
    Range       core.TextRange
    Message     RuleMessage
    Suggestions *[]RuleSuggestion
    SourceFile  *ast.SourceFile
    Severity    DiagnosticSeverity
}

type RuleSuggestion struct {
    Description string
    Fix         *SourceCodeFixer
}
```

### Severity Levels

- `SeverityError`: Blocks compilation/CI
- `SeverityWarning`: Reports issues but doesn't block
- `SeverityOff`: Rule is disabled

### Autofix System

The `SourceCodeFixer` enables rules to provide automated fixes:

```go
type SourceCodeFixer struct {
    // TODO: Document SourceCodeFixer implementation
    // Likely includes text replacement operations
}
```

**TODO**: Document detailed autofix capabilities and safety guarantees

## 8. Configuration & Directives

### Configuration Format

Rslint uses a JSON array format in `rslint.json`:

```json
[
  {
    "ignores": ["./files-not-want-lint.ts", "./tests/**/fixtures/**.ts"],
    "languageOptions": {
      "parserOptions": {
        "project": ["./tsconfig.json", "packages/app1/tsconfig.json"]
      }
    },
    "rules": {
      "@typescript-eslint/no-unused-vars": "error",
      "@typescript-eslint/array-type": ["warn", { "default": "array" }]
    }
  }
]
```

### Configuration Loading

1. **Discovery**: Search for `rslint.json` in current directory and parents
2. **Parsing**: Parse JSON configuration with validation
3. **Merging**: **TODO**: Document configuration merging strategy
4. **Rule Registry**: Map rule names to implementations in `internal/config/rule_registry.go`

### Inline Directives

**TODO**: Document support for inline rule directives like:

- `// eslint-disable-next-line @typescript-eslint/no-unused-vars`
- `/* eslint-disable @typescript-eslint/no-unsafe-assignment */`

The `DisableManager` in `internal/rule/disable_manager.go` handles directive parsing and rule disabling.

## 9. CLI Flow

### Command Line Interface

```bash
rslint [options] [files...]
```

### CLI Processing Flow

1. **Argument Parsing**: Parse command line options and file patterns
2. **Mode Selection**:
   - `--lsp`: Start Language Server Protocol server
   - `--api`: Start API server for JavaScript integration
   - Default: Direct CLI linting
3. **Configuration Loading**: Load and validate `rslint.json`
4. **Project Discovery**: Find TypeScript projects and source files
5. **File Scheduling**: Determine which files to lint based on patterns and ignores
6. **Worker Model**: **TODO**: Document concurrency model and worker pools
7. **Rule Execution**: Run configured rules on each file
8. **Result Aggregation**: Collect diagnostics from all files
9. **Output Formatting**: Format and display results
10. **Exit Code**: Return appropriate exit code based on errors/warnings

### Concurrency Model

**TODO**: Document the specific concurrency patterns:

- File-level parallelism using Go routines
- Worker pool implementation in `core.NewWorkGroup()`
- Thread safety considerations for shared state

## 10. Performance & Memory Considerations

### Design Principles

- **Zero-Copy Operations**: **TODO**: Document zero-copy string slicing where possible
- **Memory Pooling**: **TODO**: Document object pooling for frequently allocated structures
- **Parallel Execution**: Rules run in parallel across files
- **Efficient AST Traversal**: Single traversal with multiple rule listeners

### Performance Optimizations

- **Native Go Implementation**: Eliminates JavaScript runtime overhead
- **Direct TypeScript AST**: No AST conversion between parsers
- **Shared Type Checker**: Reuse TypeScript semantic analysis across rules
- **File Filtering**: Skip node_modules and bundled files automatically

### Caching Strategy

**TODO**: Document caching mechanisms:

- TypeScript compilation caching
- Rule result caching
- Incremental linting support

### Memory Management

- **Reference Cycles**: **TODO**: Document how reference cycles in AST are handled
- **Memory Limits**: **TODO**: Document memory usage patterns and limits
- **Garbage Collection**: Relies on Go GC for memory management

## 11. Extensibility & Future Directions

### Plugin Architecture

**TODO**: Document plugin system if planned:

- Custom rule loading
- Third-party rule packages
- Dynamic rule registration

### Rule Extension Points

- **Custom Rules**: New rules can be added to `internal/rules/`
- **Rule Options**: All rules support configuration through options parameter
- **Custom Visitors**: Rules can register multiple AST node listeners

### Integration Points

- **Language Server**: LSP server for editor integration
- **JavaScript API**: npm package for programmatic usage
- **CI/CD Integration**: Exit codes and output formats for automation

### Future Enhancements

**TODO**: Document planned features:

- Incremental linting
- Configuration file inheritance
- Custom formatter support
- More ESLint rule coverage

## 12. Testing Strategy

### Test Organization

- **Unit Tests**: Individual rule testing in `*_test.go` files
- **Integration Tests**: CLI and API testing in `packages/`
- **Golden Tests**: **TODO**: Document diagnostic output comparison testing
- **Rule Test Tools**: Dedicated testing framework in `packages/rslint-test-tools/`

### Rule Testing

Each rule includes comprehensive tests:

```go
func TestRuleName(t *testing.T) {
    tester := rule_tester.NewRuleTester()
    tester.Run(t, rule.Rule{...}, []rule_tester.TestCase{
        {
            Code: "valid code",
            Valid: true,
        },
        {
            Code: "invalid code",
            Errors: []string{"expected error message"},
        },
    })
}
```

### Test Data Management

- **Fixtures**: Test files in `internal/rules/fixtures/`
- **Test Cases**: Embedded test code in Go test files
- **Expected Outputs**: **TODO**: Document expected diagnostic format

### Continuous Integration

- **Go Tests**: `pnpm run test:go` runs all Go unit tests
- **TypeScript Tests**: `pnpm run test` runs JavaScript/TypeScript tests
- **Linting**: `pnpm run lint` and `pnpm run lint:go` ensure code quality
- **Build Verification**: `pnpm run build` verifies compilation

## 13. Adding a New Rule (Checklist)

### Step-by-Step Process

1. **Create Rule Directory**

   ```bash
   mkdir internal/rules/my_new_rule
   ```

2. **Implement Rule**

   ```bash
   # Create internal/rules/my_new_rule/my_new_rule.go
   # Follow existing rule patterns in other rule directories
   ```

3. **Add Rule Tests**

   ```bash
   # Create internal/rules/my_new_rule/my_new_rule_test.go
   # Include valid and invalid test cases
   ```

4. **Register Rule**

   ```bash
   # Add import and registration in internal/config/rule_registry.go
   ```

5. **Update Configuration Schema**

   ```bash
   # Add rule to rslint-schema.json if needed
   ```

6. **Test Implementation**

   ```bash
   go test ./internal/rules/my_new_rule/
   pnpm run test:go
   ```

7. **Add Documentation**
   ```bash
   # Document rule behavior and options
   # Add to rule documentation
   ```

### Rule Implementation Template

```go
package my_new_rule

import (
    "github.com/microsoft/typescript-go/shim/ast"
    "github.com/web-infra-dev/rslint/internal/rule"
)

type Config struct {
    // Rule-specific configuration options
}

func Run(ctx rule.RuleContext, options any) rule.RuleListeners {
    config := parseOptions(options)

    return rule.RuleListeners{
        ast.KindTargetNode: func(node *ast.Node) {
            // Rule logic here
            if violatesRule(node) {
                ctx.ReportNode(node, rule.RuleMessage{
                    Id: "violation",
                    Description: "Rule violation description",
                })
            }
        },
    }
}

func parseOptions(options any) Config {
    // Parse and validate options
    return Config{}
}
```

## 14. Dependency Layering & Boundaries

### Layer Architecture

```
┌─────────────────────────────────────────┐
│               CLI / API / LSP           │  ← cmd/
├─────────────────────────────────────────┤
│              Configuration              │  ← internal/config/
├─────────────────────────────────────────┤
│               Linter Core               │  ← internal/linter/
├─────────────────────────────────────────┤
│              Rule Framework             │  ← internal/rule/
├─────────────────────────────────────────┤
│            Individual Rules             │  ← internal/rules/
├─────────────────────────────────────────┤
│              Shared Utils               │  ← internal/utils/
├─────────────────────────────────────────┤
│             TypeScript-Go               │  ← typescript-go/
└─────────────────────────────────────────┘
```

### Dependency Rules

- **Upward Dependencies**: Lower layers never depend on upper layers
- **Rule Isolation**: Individual rules only depend on rule framework
- **TypeScript Boundary**: All TypeScript integration goes through typescript-go
- **No Circular Dependencies**: Enforced by Go module system

### Key Interfaces

- **CLI → Linter**: Command line interface calls linter core
- **Linter → Rules**: Linter loads and executes rules through framework
- **Rules → TypeScript**: Rules access AST and type information through typescript-go
- **Config → Registry**: Configuration maps rule names to implementations

## 15. Data Flow (Textual Diagram)

```
Configuration Files (rslint.json)
    ↓
Config Loader (internal/config/)
    ↓
Rule Registry & Project Discovery
    ↓
TypeScript Project Loading (typescript-go)
    ↓
Source File Discovery & Filtering
    ↓
┌─────────────────────────────────────────┐
│           For Each File:                │
│                                         │
│  Source Text                            │
│      ↓                                  │
│  TypeScript Lexer (typescript-go)       │
│      ↓                                  │
│  TypeScript Parser (typescript-go)      │
│      ↓                                  │
│  AST + Semantic Model                   │
│      ↓                                  │
│  Rule Listener Registration             │
│      ↓                                  │
│  AST Traversal + Rule Execution         │
│      ↓                                  │
│  Diagnostic Collection                  │
└─────────────────────────────────────────┘
    ↓
Result Aggregation & Filtering
    ↓
Output Formatting (CLI/API/LSP)
    ↓
User Interface / CI System
```

## 16. Glossary

- **AST**: Abstract Syntax Tree representing parsed source code structure
- **Diagnostic**: A linting issue reported by a rule (error, warning, or info)
- **Listener**: Function registered by a rule to handle specific AST node types
- **Node Kind**: Enumerated type identifying specific AST node types
- **Rule Context**: Object providing rules access to source files, type checker, and reporting functions
- **Rule Registry**: Mapping from rule names to rule implementations
- **Severity**: Level of importance for a diagnostic (error, warning, off)
- **Source Code Fixer**: Automated code modification system for providing fixes
- **TypeScript-Go**: Go port of the TypeScript compiler providing native performance
- **Workspace**: Collection of related projects and files for linting

## 17. TODO / Open Questions

### Implementation Details Needed

- [ ] Document specific concurrency patterns and worker pool implementation
- [ ] Detail memory management strategy and object pooling
- [ ] Explain caching mechanisms for TypeScript compilation and rule results
- [ ] Document error recovery strategy in parser
- [ ] Clarify node ID system and interning strategy if present

### Feature Documentation

- [ ] Document inline directive support (eslint-disable comments)
- [ ] Explain configuration merging and inheritance rules
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
