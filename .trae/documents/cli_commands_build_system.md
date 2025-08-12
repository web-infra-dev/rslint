# RSLint CLI Commands & Build System Documentation

## 1. Overview

RSLint provides a comprehensive command-line interface with multiple execution modes, extensive build automation, and cross-platform distribution support. The system is built using Go for the core linting engine and Node.js/pnpm for package management and distribution.

## 2. CLI Architecture

### 2.1 Main Entry Point

```go
// cmd/rslint/main.go
func main() {
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    os.Exit(runMain())
}

func runMain() int {
    waitForDebugSignal(10000 * time.Millisecond)
    args := os.Args[1:]
    if len(args) > 0 {
        switch args[0] {
        case "--lsp":
            return runLSP()    // Language Server Protocol mode
        case "--api":
            return runAPI()    // IPC API mode
        }
    }
    return runCMD()            // Standard CLI mode
}
```

### 2.2 Execution Modes

#### Standard CLI Mode (`runCMD()`)

- **Purpose**: Direct command-line linting with human-readable output
- **Features**: File processing, rule configuration, output formatting
- **Target Users**: Developers, CI/CD systems

#### LSP Mode (`runLSP()`)

- **Purpose**: Language Server Protocol integration
- **Features**: Real-time diagnostics, IDE integration
- **Target Users**: Code editors, IDEs

#### API Mode (`runAPI()`)

- **Purpose**: IPC communication for programmatic access
- **Features**: Structured JSON communication, embedding support
- **Target Users**: Build tools, custom integrations

## 3. CLI Command Structure

### 3.1 Command Usage

```
🚀 Rslint - Rocket Speed Linter

Usage:
  rslint [OPTIONS]

Options:
  --config PATH         Which rslint config file to use. Defaults to rslint.json.
  --format FORMAT       Output format: default | jsonline
  --fix                 Automatically fix problems
  --no-color            Disable colored output
  --force-color         Force colored output
  --quiet               Report errors only
  --max-warnings Int    Number of warnings to trigger nonzero exit code
  -h, --help            Show help
```

### 3.2 Command-Line Flags

#### Core Functionality Flags

```go
var (
    config      string  // Configuration file path
    fix         bool    // Enable automatic fixing
    format      string  // Output format (default, jsonline)
    help        bool    // Show help message
)
```

#### Output Control Flags

```go
var (
    noColor     bool    // Disable colored output
    forceColor  bool    // Force colored output
    quiet       bool    // Report errors only
    maxWarnings int     // Warning threshold for exit code
)
```

#### Performance and Debugging Flags

```go
var (
    traceOut       string  // Trace output file
    cpuprofOut     string  // CPU profiling output file
    singleThreaded bool    // Single-threaded execution
)
```

### 3.3 Flag Processing

```go
flag.StringVar(&format, "format", "default", "output format")
flag.StringVar(&config, "config", "", "which rslint config to use")
flag.BoolVar(&fix, "fix", false, "automatically fix problems")
flag.BoolVar(&help, "help", false, "show help")
flag.BoolVar(&help, "h", false, "show help")
flag.BoolVar(&noColor, "no-color", false, "disable colored output")
flag.BoolVar(&forceColor, "force-color", false, "force colored output")
flag.BoolVar(&quiet, "quiet", false, "report errors only")
flag.IntVar(&maxWarnings, "max-warnings", -1, "Number of warnings to trigger nonzero exit code")
```

## 4. Output Formatting System

### 4.1 Color Scheme Management

```go
type ColorScheme struct {
    RuleName    func(format string, a ...interface{}) string
    FileName    func(format string, a ...interface{}) string
    ErrorText   func(format string, a ...interface{}) string
    SuccessText func(format string, a ...interface{}) string
    DimText     func(format string, a ...interface{}) string
    BoldText    func(format string, a ...interface{}) string
    BorderText  func(format string, a ...interface{}) string
    WarnText    func(format string, a ...interface{}) string
}
```

#### Color Environment Handling

```go
func setupColors() *ColorScheme {
    // Handle environment variables
    if os.Getenv("NO_COLOR") != "" {
        color.NoColor = true
    }
    if os.Getenv("FORCE_COLOR") != "" {
        color.NoColor = false
    }

    // GitHub Actions specific handling
    if os.Getenv("GITHUB_ACTIONS") != "" {
        color.NoColor = false
    }

    // Create color functions
    return &ColorScheme{
        RuleName:    color.New(color.FgHiGreen).SprintfFunc(),
        FileName:    color.New(color.FgCyan, color.Italic).SprintfFunc(),
        ErrorText:   color.New(color.FgRed, color.Bold).SprintfFunc(),
        SuccessText: color.New(color.FgGreen, color.Bold).SprintfFunc(),
        DimText:     color.New(color.Faint).SprintfFunc(),
        BoldText:    color.New(color.Bold).SprintfFunc(),
        BorderText:  color.New(color.Faint).SprintfFunc(),
        WarnText:    color.New(color.FgYellow).SprintfFunc(),
    }
}
```

### 4.2 Output Formats

#### Default Format

- **Visual Design**: Rich terminal output with colors, borders, and code context
- **Code Display**: Shows source code with syntax highlighting and error underlining
- **Summary**: Comprehensive statistics with execution time and thread count

#### JSON Line Format

- **Structure**: One JSON object per line for machine processing
- **Compatibility**: LSP and IDE integration friendly
- **Fields**: Structured diagnostic information with precise positioning

```go
type Diagnostic struct {
    RuleName string `json:"ruleName"`
    Message  string `json:"message"`
    FilePath string `json:"filePath"`
    Range    Range  `json:"range"`
    Severity string `json:"severity"`
}
```

### 4.3 Diagnostic Display

#### Default Format Example

```
 @typescript-eslint/no-unused-vars  — [error] 'unused' is assigned a value but never used.
  ╭─┴──────────( src/index.ts:1:7 )─────
  │ 1 │  const unused = 42;
  │   │        ^^^^^^
  ╰────────────────────────────────
```

#### JSON Line Format Example

```json
{
  "ruleName": "@typescript-eslint/no-unused-vars",
  "message": "'unused' is assigned a value but never used.",
  "filePath": "src/index.ts",
  "range": {
    "start": { "line": 1, "column": 7 },
    "end": { "line": 1, "column": 13 }
  },
  "severity": "error"
}
```

## 5. Build System Architecture

### 5.1 Monorepo Structure

```yaml
# pnpm-workspace.yaml
packages:
  - 'packages/*'
  - 'npm/*'
```

#### Package Organization

- **packages/**: TypeScript/JavaScript packages
  - `rslint/`: Main Node.js package
  - `rslint-test-tools/`: Testing utilities
  - `rule-tester/`: Rule testing framework
  - `utils/`: Shared utilities
  - `vscode-extension/`: VS Code extension
- **npm/**: Platform-specific binaries
  - `darwin-arm64/`, `darwin-x64/`
  - `linux-arm64/`, `linux-x64/`
  - `win32-arm64/`, `win32-x64/`

### 5.2 Go Module Configuration

```go
// go.mod
module github.com/web-infra-dev/rslint

go 1.24.1

// Local shim replacements
replace (
    github.com/microsoft/typescript-go/shim/ast => ./shim/ast
    github.com/microsoft/typescript-go/shim/bundled => ./shim/bundled
    github.com/microsoft/typescript-go/shim/checker => ./shim/checker
    // ... additional shim replacements
)
```

#### Key Dependencies

- **TypeScript Integration**: `github.com/microsoft/typescript-go` shims
- **File Processing**: `github.com/bmatcuk/doublestar/v4` for glob patterns
- **Terminal Output**: `github.com/fatih/color` for colored output
- **JSON Processing**: `github.com/tailscale/hujson` for JSONC support
- **LSP Support**: `github.com/sourcegraph/jsonrpc2`

### 5.3 Cross-Platform Build System

#### Build Script (`scripts/build-npm.mjs`)

```javascript
const platforms = [
  { os: 'darwin', arch: 'amd64', 'node-arch': 'x64' },
  { os: 'darwin', arch: 'arm64', 'node-arch': 'arm64' },
  { os: 'linux', arch: 'amd64', 'node-arch': 'x64' },
  { os: 'linux', arch: 'arm64', 'node-arch': 'arm64' },
  { os: 'windows', arch: 'amd64', 'node-arch': 'x64', 'node-os': 'win32' },
  { os: 'windows', arch: 'arm64', 'node-arch': 'arm64', 'node-os': 'win32' },
];

for (const platform of platforms) {
  await $`GOOS=${platform.os} GOARCH=${platform.arch} go build -o npm/${platform['node-os'] || platform.os}-${platform['node-arch']}/rslint ./cmd/rslint`;
}
```

#### Supported Platforms

- **macOS**: Intel (x64) and Apple Silicon (arm64)
- **Linux**: x64 and arm64
- **Windows**: x64 and arm64

### 5.4 Package Scripts

```json
{
  "scripts": {
    "build": "pnpm -r build",
    "build:npm": "zx scripts/build-npm.mjs",
    "test": "pnpm -r test",
    "test:go": "go test ./internal/...",
    "typecheck": "pnpm tsc -b tsconfig.json",
    "lint": "rslint",
    "lint:go": "golangci-lint run ./cmd/... ./internal/...",
    "format:go": "golangci-lint fmt ./cmd/... ./internal/...",
    "version": "zx scripts/version.mjs",
    "release": "pnpm publish -r --no-git-checks",
    "publish:vsce": "zx scripts/publish-marketplace.mjs",
    "publish:ovsx": "zx scripts/publish-marketplace.mjs --marketplace=ovsx"
  }
}
```

## 6. Quality Assurance and Linting

### 6.1 Go Linting Configuration (`.golangci.yml`)

#### Enabled Linters

```yaml
linters:
  enable:
    - asasalint # Slice assignment analysis
    - bidichk # Unicode bidirectional text detection
    - bodyclose # HTTP response body closing
    - canonicalheader # HTTP header canonicalization
    - copyloopvar # Loop variable copying
    - durationcheck # Duration usage validation
    - errcheck # Error checking
    - errchkjson # JSON error checking
    - errname # Error naming conventions
    - errorlint # Error wrapping
    - fatcontext # Context usage
    - ineffassign # Ineffectual assignments
    - misspell # Spelling errors
    - staticcheck # Static analysis
    - unused # Unused code detection
```

#### Custom Rules

```yaml
settings:
  depguard:
    rules:
      main:
        deny:
          - pkg: 'encoding/json$'
            desc: 'Use "github.com/microsoft/typescript-go/internal/json" instead.'
```

### 6.2 TypeScript Configuration

#### Base Configuration (`tsconfig.base.json`)

- **Target**: ES2022
- **Module**: ESNext
- **Strict Mode**: Enabled
- **Path Mapping**: Workspace package resolution

#### Build Configuration (`tsconfig.build.json`)

- **Composite**: True for project references
- **Declaration**: True for type generation
- **Source Maps**: Enabled for debugging

## 7. Performance Optimization

### 7.1 Concurrent Processing

```go
// Multi-threaded execution by default
threadsCount := runtime.GOMAXPROCS(0)
if singleThreaded {
    threadsCount = 1
}

// Concurrent diagnostic processing
diagnosticsChan := make(chan rule.RuleDiagnostic, 4096)
var wg sync.WaitGroup

wg.Add(1)
go func() {
    defer wg.Done()
    w := bufio.NewWriterSize(os.Stdout, 4096*100)
    defer w.Flush()
    for d := range diagnosticsChan {
        printDiagnostic(d, w, comparePathOptions, format)
    }
}()
```

### 7.2 Memory Management

- **Buffered I/O**: Large buffers for output operations
- **Channel Buffering**: 4096-element diagnostic channel
- **VFS Caching**: Cached virtual file system for repeated access
- **Program Reuse**: TypeScript programs cached across files

### 7.3 Profiling Support

```go
// CPU Profiling
if cpuprofOut != "" {
    f, _ := os.Create(cpuprofOut)
    defer f.Close()
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()
}

// Execution Tracing
if traceOut != "" {
    f, _ := os.Create(traceOut)
    defer f.Close()
    trace.Start(f)
    defer trace.Stop()
}
```

## 8. Automatic Fixing System

### 8.1 Fix Application Process

```go
if fix && len(diagnosticsByFile) > 0 {
    for fileName, fileDiagnostics := range diagnosticsByFile {
        diagnosticsWithFixes := make([]rule.RuleDiagnostic, 0)
        for _, d := range fileDiagnostics {
            if len(d.Fixes()) > 0 {
                diagnosticsWithFixes = append(diagnosticsWithFixes, d)
            }
        }

        if len(diagnosticsWithFixes) > 0 {
            originalContent := diagnosticsWithFixes[0].SourceFile.Text()
            fixedContent, unapplied, wasFixed := linter.ApplyRuleFixes(originalContent, diagnosticsWithFixes)

            if wasFixed {
                err := os.WriteFile(fileName, []byte(fixedContent), 0644)
                if err == nil {
                    fixedCount += len(diagnosticsWithFixes) - len(unapplied)
                }
            }
        }
    }
}
```

### 8.2 Fix Safety

- **Non-destructive**: Original content preserved until successful fix application
- **Atomic**: All fixes for a file applied together or not at all
- **Validation**: Fixed content validated before writing
- **Reporting**: Clear feedback on fix success/failure

## 9. Integration Points

### 9.1 IDE Integration

#### VS Code Extension

- **Location**: `packages/vscode-extension/`
- **Features**: Real-time diagnostics, quick fixes, configuration
- **Communication**: LSP protocol for editor integration

#### Language Server Protocol

- **Mode**: `rslint --lsp`
- **Features**: Hover information, diagnostics, code actions
- **Transport**: JSON-RPC over stdio

### 9.2 CI/CD Integration

#### Exit Codes

- **0**: No errors found
- **1**: Errors found or execution failure
- **Configurable**: `--max-warnings` threshold

#### Output Formats

- **Human-readable**: Default format for developer feedback
- **Machine-readable**: JSON line format for automated processing
- **Quiet mode**: Error-only output for CI environments

### 9.3 Build Tool Integration

#### Package Manager Integration

```json
{
  "scripts": {
    "lint": "rslint",
    "lint:fix": "rslint --fix",
    "lint:ci": "rslint --format=jsonline --quiet"
  }
}
```

#### Pre-commit Hooks

```yaml
# .husky/pre-commit
lint-staged:
  '*.{js,jsx,ts,tsx}': ['rslint --fix', 'prettier --write']
```

## 10. Distribution and Deployment

### 10.1 NPM Package Distribution

#### Platform-specific Packages

- **Structure**: Separate packages per platform in `npm/` directory
- **Naming**: `@rslint/{platform}-{arch}` (e.g., `@rslint/darwin-arm64`)
- **Content**: Platform-specific binary + package.json

#### Main Package

- **Location**: `packages/rslint/`
- **Dependencies**: Platform-specific packages as optional dependencies
- **Binary Selection**: Runtime platform detection and binary selection

### 10.2 Release Process

#### Version Management

```javascript
// scripts/version.mjs
// Synchronizes versions across all packages
// Updates package.json files
// Creates git tags
```

#### Publishing Pipeline

```javascript
// scripts/publish-marketplace.mjs
// Publishes VS Code extension to marketplace
// Supports both Visual Studio Marketplace and Open VSX
```

### 10.3 Binary Distribution

#### Build Artifacts

- **Go Binaries**: Cross-compiled for all supported platforms
- **Size Optimization**: Stripped binaries for production
- **Compression**: Optional binary compression for distribution

#### Installation Methods

- **NPM**: `npm install -g rslint`
- **Direct Download**: Platform-specific binaries from releases
- **Package Managers**: Homebrew, Chocolatey, etc. (future)

This CLI and build system provides a robust, scalable foundation for RSLint distribution and usage across multiple platforms and integration scenarios while maintaining high performance and developer experience standards.
