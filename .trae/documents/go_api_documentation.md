# RSLint Go API Documentation

This document provides a comprehensive map of all available APIs in the Go part of the RSLint project.

## 1. Overview

The RSLint Go codebase is organized into several key components:

- **Command Line Interface** (`cmd/rslint/`)
- **Core Linting Engine** (`internal/`)
- **TypeScript-Go Integration** (`typescript-go/`)
- **Language Server Protocol (LSP)** support

## 2. Command Line Interface APIs

### 2.1 Main Entry Points

#### Location: `cmd/rslint/main.go`

- **Purpose**: Main CLI entry point
- **Commands**:
  - `lint` - Run linting on files
  - `lsp` - Start LSP server
  - `version` - Show version information

#### Location: `cmd/rslint/cmd.go`

- **Functions**:
  - `printDiagnostics(diagnostics []rule.RuleDiagnostic, format string)` - Format and print diagnostics
  - Command-line argument parsing and execution

### 2.2 IPC Handler API

#### Location: `cmd/rslint/api.go`

- **Type**: `IPCHandler`
- **Purpose**: Handle Inter-Process Communication for lint requests
- **Methods**:
  - Processes lint requests via IPC protocol
  - Returns lint results and diagnostics

### 2.3 LSP Server API

#### Location: `cmd/rslint/lsp.go`

- **Type**: `LSPServer`
- **Purpose**: Language Server Protocol implementation
- **Methods**:
  - `handleInitialize(ctx context.Context, req *jsonrpc2.Request)` - Initialize LSP server
  - `handleDidOpen(ctx context.Context, req *jsonrpc2.Request)` - Handle document open events
  - `handleDidChange(ctx context.Context, req *jsonrpc2.Request)` - Handle document change events
  - `handleDidSave(ctx context.Context, req *jsonrpc2.Request)` - Handle document save events
  - `handleCodeAction(ctx context.Context, req *jsonrpc2.Request)` - Provide code actions and fixes
  - `runDiagnostics(ctx context.Context, uri DocumentUri, content string)` - Run linting diagnostics

**Supported LSP Methods**:

- `initialize` - Server initialization
- `textDocument/didOpen` - Document opened
- `textDocument/didChange` - Document content changed
- `textDocument/didSave` - Document saved
- `textDocument/codeAction` - Code actions and quick fixes
- `shutdown` - Server shutdown

## 3. Core Linting Engine APIs

### 3.1 Rule System API

#### Location: `internal/rule/rule.go`

**Core Types**:

```go
type Rule struct {
    Name string
    Run  func(ctx RuleContext, options any) RuleListeners
}

type RuleContext struct {
    SourceFile                 *ast.SourceFile
    Program                    *compiler.Program
    TypeChecker                *checker.Checker
    DisableManager             *DisableManager
    ReportRange                func(textRange core.TextRange, msg RuleMessage)
    ReportRangeWithSuggestions func(textRange core.TextRange, msg RuleMessage, suggestions ...RuleSuggestion)
    ReportNode                 func(node *ast.Node, msg RuleMessage)
    ReportNodeWithFixes        func(node *ast.Node, msg RuleMessage, fixes ...RuleFix)
    ReportNodeWithSuggestions  func(node *ast.Node, msg RuleMessage, suggestions ...RuleSuggestion)
}

type RuleDiagnostic struct {
    Range       core.TextRange
    RuleName    string
    Message     RuleMessage
    FixesPtr    *[]RuleFix
    Suggestions *[]RuleSuggestion
    SourceFile  *ast.SourceFile
    Severity    DiagnosticSeverity
}
```

**Severity Levels**:

- `SeverityError` - Error level diagnostic
- `SeverityWarning` - Warning level diagnostic
- `SeverityOff` - Disabled rule

**Fix and Suggestion APIs**:

- `RuleFixInsertBefore(file *ast.SourceFile, node *ast.Node, text string) RuleFix`
- `RuleFixInsertAfter(node *ast.Node, text string) RuleFix`
- `RuleFixReplace(file *ast.SourceFile, node *ast.Node, text string) RuleFix`
- `RuleFixRemove(file *ast.SourceFile, node *ast.Node) RuleFix`

### 3.2 Linter Engine API

#### Location: `internal/linter/linter.go`

**Main Function**:

```go
func RunLinter(
    programs []*compiler.Program,
    singleThreaded bool,
    allowFiles []string,
    getRulesForFile func(sourceFile *ast.SourceFile) []ConfiguredRule,
    onDiagnostic func(diagnostic rule.RuleDiagnostic)
) (int32, error)
```

**Types**:

```go
type ConfiguredRule struct {
    Name     string
    Severity rule.DiagnosticSeverity
    Run      func(ctx rule.RuleContext) rule.RuleListeners
}
```

**Features**:

- Multi-threaded linting support
- File filtering capabilities
- AST visitor pattern implementation
- Rule disable comment support

### 3.3 Configuration API

#### Location: `internal/config/config.go`

**Configuration Types**:

```go
type RslintConfig []ConfigEntry

type ConfigEntry struct {
    Language        string
    Files           []string
    Ignores         []string
    LanguageOptions *LanguageOptions
    Rules           Rules
    Plugins         []string
}

type RuleConfig struct {
    Level   string
    Options map[string]interface{}
}
```

**Rule Management**:

- `GetAllRulesForPlugin(plugin string) []rule.Rule` - Get all rules for a plugin
- `parseArrayRuleConfig(ruleArray []interface{}) *RuleConfig` - Parse ESLint-style rule config

### 3.4 IPC Protocol API

#### Location: `internal/api/api.go`

**Message Types**:

```go
type LintRequest struct {
    Files   []string
    Config  *config.RslintConfig
    Options *LintOptions
}

type LintResponse struct {
    Diagnostics []rule.RuleDiagnostic
    FileCount   int32
    Success     bool
    Error       string
}
```

**Service Interface**:

```go
type Service struct {
    // IPC communication management
    // Request/response handling
    // Process lifecycle management
}
```

## 4. TypeScript-Go Integration APIs

### 4.1 Core API Service

#### Location: `typescript-go/internal/api/api.go`

**Main API Type**:

```go
type API struct {
    host                APIHost
    options             APIOptions
    documentStore       *project.DocumentStore
    configFileRegistry  *project.ConfigFileRegistry
    projects            handleMap[project.Project]
    files               handleMap[ast.SourceFile]
    symbols             handleMap[ast.Symbol]
    types               handleMap[checker.Type]
}
```

**API Methods**:

- `HandleRequest(ctx context.Context, method string, payload []byte) ([]byte, error)`
- `GetSourceFile(project Handle[project.Project], fileName string) (*ast.SourceFile, error)`
- `ParseConfigFile(configFileName string) (*ConfigFileResponse, error)`
- `LoadProject(configFileName string) (Handle[project.Project], error)`
- `GetSymbolAtPosition(ctx context.Context, project Handle[project.Project], fileName string, position int)`
- `GetSymbolAtLocation(ctx context.Context, project Handle[project.Project], location Handle[ast.Node])`
- `GetTypeOfSymbol(ctx context.Context, project Handle[project.Project], symbol Handle[ast.Symbol])`

**Supported API Methods**:

- `MethodRelease` - Release handles
- `MethodGetSourceFile` - Get source file information
- `MethodParseConfigFile` - Parse TypeScript config files
- `MethodLoadProject` - Load TypeScript projects
- `MethodGetSymbolAtPosition` - Get symbol at specific position
- `MethodGetSymbolsAtPositions` - Get symbols at multiple positions
- `MethodGetSymbolAtLocation` - Get symbol at AST location
- `MethodGetTypeOfSymbol` - Get type information for symbols

### 4.2 LSP Server Integration

#### Location: `typescript-go/internal/lsp/server.go`

**Server Type**:

```go
type Server struct {
    // LSP protocol handling
    // Project management
    // File watching capabilities
    // TypeScript service integration
}
```

**Key Interfaces**:

- `project.ServiceHost` - TypeScript service host implementation
- `project.Client` - Client communication interface

**Features**:

- File system watching
- Project service management
- LSP protocol compliance
- TypeScript compiler integration

## 5. Available Rules

The system includes the following TypeScript ESLint rules:

### 5.1 Core Rules

- `adjacent-overload-signatures`
- `array-type`
- `await-thenable`
- `class-literal-property-style`
- `dot-notation`
- `explicit-member-accessibility`
- `max-params`
- `member-ordering`

### 5.2 Safety Rules

- `no-array-delete`
- `no-base-to-string`
- `no-confusing-void-expression`
- `no-duplicate-type-constituents`
- `no-empty-function`
- `no-empty-interface`
- `no-floating-promises`
- `no-for-in-array`
- `no-implied-eval`
- `no-meaningless-void-operator`
- `no-misused-promises`
- `no-misused-spread`
- `no-mixed-enums`
- `no-redundant-type-constituents`
- `no-require-imports`
- `no-unnecessary-boolean-literal-compare`
- `no-unnecessary-template-expression`
- `no-unnecessary-type-arguments`
- `no-unnecessary-type-assertion`
- `no-unsafe-argument`
- `no-unsafe-assignment`
- `no-unsafe-call`
- `no-unsafe-enum-comparison`
- `no-unsafe-member-access`
- `no-unsafe-return`
- `no-unsafe-type-assertion`
- `no-unsafe-unary-minus`
- `no-unused-vars`
- `no-useless-empty-export`
- `no-var-requires`

### 5.3 Style and Best Practice Rules

- `non-nullable-type-assertion-style`
- `only-throw-error`
- `prefer-as-const`
- `prefer-promise-reject-errors`
- `prefer-reduce-type-parameter`
- `prefer-return-this-type`
- `promise-function-async`
- `related-getter-setter-pairs`
- `require-array-sort-compare`
- `require-await`
- `restrict-plus-operands`
- `restrict-template-expressions`
- `return-await`
- `switch-exhaustiveness-check`
- `unbound-method`
- `use-unknown-in-catch-callback-variable`

## 6. Usage Examples

### 6.1 Running Linter Programmatically

```go
import (
    "github.com/web-infra-dev/rslint/internal/linter"
    "github.com/web-infra-dev/rslint/internal/config"
)

// Configure rules
rules := []linter.ConfiguredRule{
    {
        Name:     "@typescript-eslint/no-unused-vars",
        Severity: rule.SeverityError,
        Run:      no_unused_vars.Rule.Run,
    },
}

// Run linter
fileCount, err := linter.RunLinter(
    programs,
    false, // multi-threaded
    nil,   // all files
    func(file *ast.SourceFile) []linter.ConfiguredRule {
        return rules
    },
    func(diagnostic rule.RuleDiagnostic) {
        // Handle diagnostic
    },
)
```

### 6.2 Creating Custom Rules

```go
var MyCustomRule = rule.CreateRule(rule.Rule{
    Name: "my-custom-rule",
    Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
        return rule.RuleListeners{
            ast.KindVariableDeclaration: func(node *ast.Node) {
                // Rule logic here
                ctx.ReportNode(node, rule.RuleMessage{
                    Id:          "custom-error",
                    Description: "Custom rule violation",
                })
            },
        }
    },
})
```

### 6.3 LSP Integration

```go
// Start LSP server
server := NewLSPServer()
conn := jsonrpc2.NewConn(
    context.Background(),
    jsonrpc2.NewBufferedStream(os.Stdin, jsonrpc2.VSCodeObjectCodec{}),
    server,
)
<-conn.DisconnectNotify()
```

## 7. Error Handling

All APIs follow Go's standard error handling patterns:

- Functions return `(result, error)` tuples
- Errors are properly wrapped with context
- Diagnostic severity levels control error reporting
- LSP protocol errors are handled according to specification

## 8. Performance Considerations

- **Multi-threading**: Linter supports concurrent processing
- **Caching**: TypeScript programs and parsed files are cached
- **Incremental**: LSP server supports incremental updates
- **Memory Management**: Proper handle management for TypeScript objects

This documentation covers all major APIs available in the RSLint Go codebase. Each API is designed to be composable and follows Go best practices for error handling and concurrency.
