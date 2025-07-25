# RSLint Project Reference

This document provides comprehensive technical information for RSLint, a high-performance TypeScript/JavaScript linter written in Go.

## Overview

RSLint is a drop-in replacement for ESLint and TypeScript-ESLint, built on top of `typescript-go` (a Go port of the TypeScript compiler). It provides typed linting with project-level analysis and offers multiple integration modes.

## Architecture

### Project Structure
```
/cmd/rslint/              # Main Go CLI binary
/internal/                # Core Go packages
  ├── api/                # IPC communication protocol
  ├── linter/             # Core linting engine
  ├── rule/               # Rule system interfaces
  ├── rules/              # Built-in rules (48 rules)
  ├── config/             # Configuration handling
  └── utils/              # Utility functions
/packages/                # JavaScript/TypeScript packages
  ├── rslint/             # Main Node.js package
  ├── rslint-test-tools/  # Testing framework
  └── vscode-extension/   # VS Code extension
/typescript-go/           # TypeScript compiler Go port
/eslint-to-go-porter/     # Tool for porting ESLint rules
```

## Command-Line Interface

### Basic Usage
```bash
rslint [OPTIONS] [FILES...]
```

### Available Options

| Flag | Description | Default |
|------|-------------|---------|
| `--tsconfig PATH` | TypeScript configuration file | `tsconfig.json` |
| `--config PATH` | RSLint configuration file | `rslint.jsonc` |
| `--list-files` | List matched files | |
| `--format FORMAT` | Output format: `default` \| `jsonline` | `default` |
| `--lsp` | Run in Language Server Protocol mode | |
| `--api` | Run in IPC mode for JavaScript integration | |
| `--no-color` | Disable colored output | |
| `--force-color` | Force colored output | |
| `--trace FILE` | File for trace output | |
| `--cpuprof FILE` | File for CPU profiling | |
| `--singleThreaded` | Run in single-threaded mode | |
| `-h, --help` | Show help | |

### Output Formats

#### Default Format
Rich, colored terminal output with code context and error highlighting.

#### JSON Line Format
Machine-readable JSON output for CI/CD integration:
```json
{
  "ruleName": "no-array-delete",
  "message": "Using the `delete` operator with an array expression is unsafe.",
  "filePath": "src/example.ts",
  "range": {
    "start": {"line": 5, "column": 1},
    "end": {"line": 5, "column": 15}
  },
  "severity": "error"
}
```

## IPC (Inter-Process Communication) API

### Protocol Overview
- **Transport**: Binary message format over stdio
- **Encoding**: 4-byte length prefix (uint32 little endian) + JSON content
- **Communication**: Request/response pattern with async support

### Message Types
```go
type MessageKind string

const (
    KindHandshake MessageKind = "handshake"  // Initial connection verification
    KindLint      MessageKind = "lint"       // Lint request
    KindResponse  MessageKind = "response"   // Successful response
    KindError     MessageKind = "error"      // Error response
    KindExit      MessageKind = "exit"       // Termination request
)
```

### Request Structure
```go
type LintRequest struct {
    Files            []string          `json:"files,omitempty"`
    TSConfig         string            `json:"tsconfig,omitempty"`
    Format           string            `json:"format,omitempty"`
    WorkingDirectory string            `json:"workingDirectory,omitempty"`
    RuleOptions      map[string]string `json:"ruleOptions,omitempty"`
    FileContents     map[string]string `json:"fileContents,omitempty"`
}
```

### Response Structure
```go
type LintResponse struct {
    Diagnostics []Diagnostic `json:"diagnostics"`
    ErrorCount  int          `json:"errorCount"`
    FileCount   int          `json:"fileCount"`
    RuleCount   int          `json:"ruleCount"`
}

type Diagnostic struct {
    RuleName string `json:"ruleName"`
    Message  string `json:"message"`
    FilePath string `json:"filePath"`
    Range    Range  `json:"range"`
    Severity string `json:"severity,omitempty"`
}

type Range struct {
    Start Position `json:"start"`
    End   Position `json:"end"`
}

type Position struct {
    Line      int `json:"line"`      // 0-based
    Character int `json:"character"` // 0-based
}
```

### Error Handling
```go
type ErrorResponse struct {
    Message string `json:"message"`
}
```

## Language Server Protocol (LSP)

### Server Capabilities
```go
type ServerCapabilities struct {
    TextDocumentSync   int  `json:"textDocumentSync"`    // 1 = Full document sync
    DiagnosticProvider bool `json:"diagnosticProvider"`   // true
}
```

### Supported Methods

| Method | Description | Request Type | Response Type |
|--------|-------------|--------------|---------------|
| `initialize` | Server initialization | `InitializeParams` | `InitializeResult` |
| `initialized` | Post-initialization notification | - | - |
| `textDocument/didOpen` | Document opened | `DidOpenTextDocumentParams` | - |
| `textDocument/didChange` | Document changed | `DidChangeTextDocumentParams` | - |
| `textDocument/didSave` | Document saved | - | - |
| `textDocument/diagnostic` | Diagnostic request | - | - |
| `shutdown` | Server shutdown | - | - |
| `exit` | Server termination | - | - |

### LSP Data Structures

#### Initialize Parameters
```go
type InitializeParams struct {
    ProcessID    *int    `json:"processId"`
    RootPath     *string `json:"rootPath"`
    RootURI      *string `json:"rootUri"`
    Capabilities ClientCapabilities `json:"capabilities"`
}
```

#### Text Document Synchronization
```go
type DidOpenTextDocumentParams struct {
    TextDocument TextDocumentItem `json:"textDocument"`
}

type TextDocumentItem struct {
    URI        string `json:"uri"`        // file:// URI
    LanguageID string `json:"languageId"` // "typescript", "javascript"
    Version    int    `json:"version"`
    Text       string `json:"text"`
}
```

#### Diagnostics
```go
type LspDiagnostic struct {
    Range    Range  `json:"range"`
    Severity int    `json:"severity"`     // 1 = Error, 2 = Warning
    Source   string `json:"source"`       // "rslint"
    Message  string `json:"message"`
}

type PublishDiagnosticsParams struct {
    URI         string          `json:"uri"`
    Diagnostics []LspDiagnostic `json:"diagnostics"`
}
```

## JavaScript/Node.js API

### Installation
```bash
npm install @rslint/core
```

### Core Service Class

#### Constructor
```typescript
export class RSLintService {
  constructor(options: RSlintOptions = {})
}

interface RSlintOptions {
  rslintPath?: string;                       // Path to rslint binary
  workingDirectory?: string;                 // Working directory
}
```

#### Lint Method
```typescript
async lint(options: LintOptions): Promise<LintResponse>

interface LintOptions {
  files?: string[];                          // Specific files to lint
  tsconfig?: string;                         // TypeScript config path
  workingDirectory?: string;                 // Working directory
  ruleOptions?: Record<string, string>;      // Rule-specific options
  fileContents?: Record<string, string>;     // Virtual file system
}
```

#### Cleanup
```typescript
close(): Promise<void>
```

### Convenience Function
```typescript
export async function lint(options: LintOptions): Promise<LintResponse>
```

### Response Types
```typescript
export interface LintResponse {
  diagnostics: Diagnostic[];
  errorCount: number;
  fileCount: number;
  ruleCount: number;
  duration: string;
}

export interface Diagnostic {
  ruleName: string;
  message: string;
  filePath: string;
  range: Range;
  severity?: string;
}

export interface Range {
  start: Position;
  end: Position;
}

export interface Position {
  line: number;      // 0-based
  character: number; // 0-based
}
```

### Usage Examples

#### One-shot Linting
```javascript
import { lint } from '@rslint/core';

const result = await lint({
  tsconfig: './tsconfig.json',
  files: ['src/**/*.ts']
});

console.log(`Found ${result.errorCount} errors`);
```

#### Service-based Usage
```javascript
import { RSLintService } from '@rslint/core';

const service = new RSLintService();
const result = await service.lint({ 
  files: ['src/index.ts'] 
});
await service.close();
```

#### Virtual File System
```javascript
const result = await lint({
  fileContents: {
    '/path/to/file.ts': 'let x: any = 10; x.foo = 5;'
  }
});
```

## Configuration

### Configuration File Format
RSLint uses an array-based configuration format (JSON/JSONC):

```jsonc
[
  {
    "language": "typescript",
    "files": ["src/**/*.ts", "src/**/*.tsx"],
    "languageOptions": {
      "parserOptions": {
        "project": ["./tsconfig.json"],
        "projectService": false
      }
    },
    "rules": {
      "no-array-delete": "error",
      "no-unsafe-assignment": "warn",
      "prefer-as-const": "error"
    }
  }
]
```

### Configuration Schema
```go
type RslintConfig []ConfigEntry

type ConfigEntry struct {
    Language        string           `json:"language"`
    Files           []string         `json:"files"`
    LanguageOptions *LanguageOptions `json:"languageOptions,omitempty"`
    Rules           Rules            `json:"rules"`
}

type LanguageOptions struct {
    ParserOptions *ParserOptions `json:"parserOptions,omitempty"`
}

type ParserOptions struct {
    ProjectService bool     `json:"projectService"`
    Project        []string `json:"project,omitempty"`
}
```

### Rule Configuration
Rules can be configured with the following severity levels:
- `"off"` or `0` - Rule is disabled
- `"warn"` or `1` - Rule produces warnings
- `"error"` or `2` - Rule produces errors

## Rule Development API

### Rule Structure
```go
type Rule struct {
    Name string
    Run  func(ctx RuleContext, options any) RuleListeners
}

type RuleListeners map[ast.Kind](func(node *ast.Node))
```

### Rule Context
```go
type RuleContext struct {
    SourceFile                 *ast.SourceFile
    Program                    *compiler.Program
    TypeChecker                *checker.Checker
    ReportRange                func(textRange core.TextRange, msg RuleMessage)
    ReportRangeWithSuggestions func(textRange core.TextRange, msg RuleMessage, suggestions ...RuleSuggestion)
    ReportNode                 func(node *ast.Node, msg RuleMessage)
    ReportNodeWithFixes        func(node *ast.Node, msg RuleMessage, fixes ...RuleFix)
    ReportNodeWithSuggestions  func(node *ast.Node, msg RuleMessage, suggestions ...RuleSuggestion)
}
```

### Reporting Functions

#### Basic Reporting
```go
// Report diagnostic at specific text range
ctx.ReportRange(textRange, RuleMessage{Description: "Error message"})

// Report diagnostic at AST node
ctx.ReportNode(node, RuleMessage{Description: "Error message"})
```

#### Advanced Reporting
```go
// Report with automatic fixes
ctx.ReportNodeWithFixes(node, message, 
    RuleFix{Text: "replacement", Range: range})

// Report with suggestions
ctx.ReportNodeWithSuggestions(node, message,
    RuleSuggestion{
        Message: RuleMessage{Description: "Suggestion"},
        FixesArr: []RuleFix{{Text: "fix", Range: range}},
    })
```

### Rule Development Pattern
```go
var MyRule = rule.Rule{
    Name: "my-rule",
    Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
        return rule.RuleListeners{
            ast.KindBinaryExpression: func(node *ast.Node) {
                // Rule logic here
                if violatesRule(node) {
                    ctx.ReportNode(node, rule.RuleMessage{
                        Description: "This violates my rule",
                    })
                }
            },
        }
    },
}
```

### AST Node Kinds
Rules can listen to specific AST node types:
- `ast.KindBinaryExpression`
- `ast.KindCallExpression`
- `ast.KindVariableDeclaration`
- `ast.KindFunctionDeclaration`
- `ast.KindDeleteExpression`
- And many more...

## Built-in Rules (48 total)

### Type Safety Rules
- `await-thenable` - Disallows awaiting non-thenable values
- `no-unsafe-argument` - Disallows calling with unsafe arguments
- `no-unsafe-assignment` - Disallows unsafe value assignments
- `no-unsafe-call` - Disallows calling unsafe values
- `no-unsafe-enum-comparison` - Disallows unsafe enum comparisons
- `no-unsafe-member-access` - Disallows unsafe member access
- `no-unsafe-return` - Disallows unsafe return values
- `no-unsafe-type-assertion` - Disallows unsafe type assertions
- `no-unsafe-unary-minus` - Disallows unsafe unary minus operations

### Promise Handling Rules
- `no-floating-promises` - Requires proper handling of promises
- `no-misused-promises` - Prevents misuse of promises
- `prefer-promise-reject-errors` - Requires rejecting with Error objects
- `promise-function-async` - Requires promise-returning functions to be async
- `return-await` - Enforces consistent await usage in return statements

### Code Quality Rules
- `no-array-delete` - Disallows delete operator on arrays
- `no-base-to-string` - Disallows toString() on non-string types
- `no-for-in-array` - Disallows for-in loops over arrays
- `no-implied-eval` - Disallows implied eval()
- `no-meaningless-void-operator` - Disallows meaningless void operators
- `no-unnecessary-boolean-literal-compare` - Disallows unnecessary boolean comparisons
- `no-unnecessary-template-expression` - Disallows unnecessary template expressions
- `no-unnecessary-type-arguments` - Disallows unnecessary type arguments
- `no-unnecessary-type-assertion` - Disallows unnecessary type assertions
- `only-throw-error` - Requires throwing Error objects
- `prefer-as-const` - Prefers const assertions
- `prefer-reduce-type-parameter` - Prefers explicit reduce type parameters
- `prefer-return-this-type` - Prefers return this type annotations
- `require-array-sort-compare` - Requires compare function for Array.sort()
- `require-await` - Requires await in async functions
- `restrict-plus-operands` - Restricts + operator operands
- `restrict-template-expressions` - Restricts template expression types
- `switch-exhaustiveness-check` - Requires exhaustive switch statements
- `unbound-method` - Prevents unbound method calls

## Integration Examples

### Command Line
```bash
# Basic linting
rslint

# Custom config and tsconfig
rslint --config ./rslint.jsonc --tsconfig ./tsconfig.json

# JSON output for CI
rslint --format jsonline > lint-results.json

# LSP mode for editors
rslint --lsp

# IPC mode for JavaScript
rslint --api
```

### VS Code Integration
The VS Code extension automatically uses the LSP server for real-time linting.

### Build System Integration
```javascript
// webpack.config.js
const { lint } = require('@rslint/core');

const rslintPlugin = {
  apply(compiler) {
    compiler.hooks.compilation.tap('RSLintPlugin', async () => {
      const result = await lint({
        files: ['src/**/*.ts'],
        tsconfig: './tsconfig.json'
      });
      
      if (result.errorCount > 0) {
        console.error(`RSLint found ${result.errorCount} errors`);
      }
    });
  }
};

module.exports = {
  plugins: [rslintPlugin]
};
```

### CI/CD Integration
```yaml
# .github/workflows/lint.yml
name: Lint
on: [push, pull_request]
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-node@v2
      - run: npm install
      - run: npx rslint --format jsonline > lint-results.json
      - run: cat lint-results.json
```

## Performance Characteristics

- **Multi-threaded**: Parallel rule execution for optimal performance
- **Project-level caching**: Reuses TypeScript compilation across files
- **Memory efficient**: Streaming diagnostics collection
- **Type-aware**: Full TypeScript type checker integration

## File Support

- TypeScript: `.ts`, `.tsx`
- JavaScript: `.js`, `.jsx`
- Configuration: JSON, JSONC
- Virtual files: In-memory content via API

## Error Codes and Exit Status

- `0` - Success, no errors found
- `1` - Linting errors found
- `2` - Configuration or runtime error

## Common Debugging Patterns & Gotchas

### Rule Registration
- Rules must be registered with BOTH namespaced and non-namespaced names:
  ```go
  GlobalRuleRegistry.Register("@typescript-eslint/array-type", array_type.ArrayTypeRule)
  GlobalRuleRegistry.Register("array-type", array_type.ArrayTypeRule)  // Also needed for tests!
  ```
- Tests often use the non-namespaced version (e.g., "array-type" not "@typescript-eslint/array-type")

### AST Navigation
- Use `utils.GetNameFromMember()` for robust property name extraction instead of custom implementations
- Class members are accessed via `node.Members()` which returns `[]*ast.Node` directly (not a NodeList)
- Accessor properties (`accessor a = ...`) are `PropertyDeclaration` nodes with the accessor modifier

### Position/Range Reporting
- RSLint uses 1-based line and column numbers for compatibility with TypeScript-ESLint
- The IPC API uses 0-based positions internally but converts to 1-based for display
- When reporting on nodes, consider which part should be highlighted (e.g., identifier vs entire declaration)

### Common Issues & Solutions
1. **Infinite loops/timeouts**: Check for recursive function calls without proper base cases
2. **Rule not executing**: Verify rule is registered in `config.go` with both naming variants
3. **Message ID mismatches**: Use camelCase message IDs or ensure `messageId` field is populated
4. **Test snapshot mismatches**: Update snapshots when rule count changes

### Testing Best Practices
- Run specific tests with: `node --import=tsx/esm --test tests/typescript-eslint/rules/RULE_NAME.test.ts`
- Update API test snapshots after rule changes: `cd packages/rslint && npm test -- --update-snapshots`
- Debug output can be added with `fmt.Printf()` but remember to remove it before committing

# Rslint Project Copilot Instructions

## Project Overview

Rslint is a high-performance TypeScript/JavaScript linter written in Go, designed as a drop-in replacement for ESLint and TypeScript-ESLint. It leverages [typescript-go](https://github.com/microsoft/typescript-go) to achieve 20-40x speedup over traditional ESLint setups through native parsing, direct TypeScript AST usage, and parallel processing.

## Project Structure

### Core Components
- **Go Backend** (`cmd/`, `internal/`): Core linter engine written in Go
- **Node.js Package** (`packages/rslint/`): JavaScript API and CLI wrapper
- **VS Code Extension** (`packages/vscode-extension/`): Editor integration
- **Test Tools** (`packages/rslint-test-tools/`): Testing utilities
- **TypeScript Shims** (`shim/`): Go bindings for TypeScript compiler

### Key Directories
```
cmd/rslint/          # Main Go binary entry point
internal/
├── api/               # API layer
├── linter/            # Core linting engine
├── rule/              # Rule definition framework
├── rule_tester/       # Rule testing utilities
├── rules/             # Individual lint rule implementations
└── utils/             # Shared utilities
packages/
├── rslint/            # Main npm package
├── rslint-test-tools/ # Testing framework
└── vscode-extension/  # VS Code integration
typescript-go/         # TypeScript compiler Go port
```

## Technologies & Languages

### Primary Stack
- **Go 1.24+**: Core linter implementation
- **TypeScript/JavaScript**: Node.js API, tooling, and VS Code extension
- **typescript-go**: TypeScript compiler bindings for Go

### Build Tools
- **pnpm**: Package management (workspace setup)
- **Go modules**: Go dependency management
- **TypeScript**: Compilation and type checking

## Development Guidelines

### Code Style & Patterns

#### Go Code
- Follow standard Go conventions and `gofmt` formatting
- Use structured error handling with context
- Implement rules as separate packages in `internal/rules/`
- Each rule should have corresponding tests in its directory
- Use the rule framework defined in `internal/rule/rule.go`

#### TypeScript/JavaScript Code
- Use TypeScript for all new code
- Follow the existing ESM module structure
- Maintain compatibility with Node.js APIs
- Use proper type definitions for Go binary interactions

### Rule Implementation
When implementing new lint rules:

1. **Create rule directory**: `internal/rules/rule_name/`
2. **Implement rule logic**: Follow the `Rule` interface
3. **Add tests**: Include test cases with expected diagnostics
4. **Register rule**: Add to the rule registry
5. **Update documentation**: Include rule description and examples

### Testing Strategy
- **Go tests**: Use Go's built-in testing framework
- **Rule tests**: Utilize the `rule_tester` package
- **Node.js tests**: Use Node.js test runner for JavaScript API
- **Integration tests**: Test the complete CLI workflow

### Build Process
```bash
# Run Install
pnpm install

# Run build
pnpm build

# Run format  
pnpm format:check

# Run lint  
pnpm lint

# Run tests
pnpm test

```

## API Guidelines

### Go API
- Rules implement the `Rule` interface with `Check()` method
- Use `Diagnostic` structs for reporting issues
- Leverage `SourceCodeFixer` for auto-fixes
- Access TypeScript type information through the checker

### JavaScript API
- Provide ESLint-compatible configuration format
- Support async operations for file processing
- Maintain compatibility with existing ESLint tooling
- Export both CLI and programmatic interfaces

## Performance Considerations

### Optimization Principles
- **Parallel processing**: Utilize all CPU cores for file processing
- **Memory efficiency**: Minimize allocations in hot paths
- **Caching**: Cache TypeScript compiler results when possible
- **Direct AST usage**: Avoid AST transformations/conversions

### Profiling & Benchmarks
- Use Go's built-in profiling tools (`go tool pprof`)
- Maintain benchmarks in `benchmarks/` directory
- Compare performance against ESLint baselines
- Monitor memory usage and GC pressure

## Integration Points

### VS Code Extension
- Language Server Protocol (LSP) support via `--lsp` flag
- Real-time diagnostics and quick fixes
- Configuration integration with workspace settings

### Node.js Ecosystem
- ESLint configuration compatibility
- npm package distribution
- CI/CD integration support

## Common Patterns

### Adding a New Rule
```go
// internal/rules/my_rule/my_rule.go
package my_rule

import (
    "github.com/typescript-eslint/rslint/internal/rule"
    // other imports
)

type Rule struct{}

func (r *Rule) Check(ctx *rule.Context) {
    // Rule implementation
}

func NewRule() rule.Rule {
    return &Rule{}
}
```

### TypeScript Integration
- Use `checker` package for type information
- Access AST nodes through `ast` package
- Utilize `scanner` for source location details

### Error Handling
- Return structured diagnostics with precise locations
- Include fix suggestions when possible
- Provide clear, actionable error messages

## File Naming Conventions
- Go files: `snake_case.go`
- TypeScript files: `kebab-case.ts` or `camelCase.ts`
- Test files: `*_test.go` for Go, `*.test.ts` for TypeScript
- Rule directories: `rule_name/` (snake_case)

## Documentation Requirements
- Document all public APIs
- Include usage examples for rules
- Maintain README files for major components
- Update benchmarks when adding performance-critical features

## Compatibility & Migration
- Maintain ESLint rule compatibility where possible
- Provide migration guides for ESLint users
- Support TypeScript-ESLint configuration formats
- Ensure backward compatibility in JavaScript APIs
