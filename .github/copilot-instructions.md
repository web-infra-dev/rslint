# Rslint Project Copilot Instructions

Always follow these instructions first and fallback to additional search and context gathering only if the information in the instructions is incomplete or found to be in error.

Rslint is a high-performance TypeScript/JavaScript linter written in Go, designed as a drop-in replacement for ESLint and TypeScript-ESLint. It leverages [typescript-go](https://github.com/microsoft/typescript-go) to achieve 20-40x speedup over traditional ESLint setups through native parsing, direct TypeScript AST usage, and parallel processing.

## Working Effectively

### Bootstrap, Build, and Test the Repository

**CRITICAL: NEVER CANCEL builds or tests. Set timeouts appropriately and wait for completion.**

**Prerequisites:**

- Node.js 24+ (verified: v24.5.0 works)
- Go 1.24+ (verified: go1.24.1 works)
- pnpm 10+ (verified: v10.13.1 works)

**Setup Commands (Execute in Order):**

1. `git submodule update --init --recursive` -- typically takes about 3 minutes, but duration may vary significantly depending on network speed and repository size. NEVER CANCEL. Set timeout to 10+ minutes.
2. `pnpm install` -- takes <1 second (dependencies cached). Set timeout to 5+ minutes for safety.
3. `pnpm build` -- takes 5 minutes. NEVER CANCEL. Set timeout to 15+ minutes.

**Validation Commands:**

- `pnpm run typecheck` -- takes 4 seconds. Set timeout to 2+ minutes.
- `pnpm run format:check` -- takes 3 seconds. Set timeout to 2+ minutes.
- `pnpm run lint` -- takes 1 second. Set timeout to 2+ minutes.
- `pnpm run test:go` -- takes 6 minutes. NEVER CANCEL. Set timeout to 15+ minutes.
- `pnpm test` -- may fail due to VS Code extension network issues (see Limitations). Use `pnpm run test:go` instead.

### Run and Test CLI

- Test CLI: `./packages/rslint/bin/rslint --help`
- Lint with CLI: `./packages/rslint/bin/rslint --config rslint.json`
- CLI works correctly and produces formatted diagnostic output with TypeScript analysis

### Development Commands

- Install: `pnpm install`
- Build: `pnpm run build` -- 5 minutes, NEVER CANCEL
- Format code: `pnpm run format`
- Check format: `pnpm run format:check`
- Lint: `pnpm run lint`
- Type check: `pnpm run typecheck`
- Test: `pnpm run test:go` (recommended) or `pnpm test` (may fail due to network issues)

## Validation Scenarios

**ALWAYS run through at least one complete end-to-end scenario after making changes:**

1. **Basic Linting Validation:**

   - Create a test TypeScript file with intentional errors (e.g., `const x: any = 'test'; const y = x + 1;`)
   - Run `./packages/rslint/bin/rslint` on the file
   - Verify that TypeScript-ESLint rules correctly identify issues
   - Expected: Errors for unsafe assignment, unused variables, etc.

2. **Build and CLI Integration:**

   - Run full build: `pnpm build`
   - Test CLI functionality: `./packages/rslint/bin/rslint --help`
   - Lint the repository itself: `./packages/rslint/bin/rslint --config rslint.json`
   - Expected: Clean build, working CLI, proper diagnostic output

3. **VS Code Extension Debug Setup (Optional):**
   - Copy launch configuration: `cp .vscode/launch.template.json .vscode/launch.json`
   - Open in VS Code and use F5 to debug extension
   - Note: May require manual VS Code installation or network access

## Limitations and Workarounds

**Known Issues:**

- `pnpm test` fails due to VS Code extension tests requiring network access to download VS Code
  - **Workaround:** Use `pnpm run test:go` for Go tests only
- `pnpm run lint:go` fails because `golangci-lint` is not installed by default
  - **Workaround:** Go linting is handled by CI; focus on TypeScript/Node.js linting locally
- One failing Go test in `no_unnecessary_type_assertion` rule (existing codebase issue)
  - **Workaround:** This is a known failing test; do not attempt to fix unless directly related to your changes

## Code Standards & Required Validation Before Commits

**ALWAYS run before committing changes:**

1. `pnpm run format` -- Format all code consistently
2. `pnpm run typecheck` -- Verify TypeScript types
3. `pnpm run lint` -- Check for linting errors (expect some warnings)
4. `pnpm run test:go` -- Run Go tests (expect 1 known failure)
5. Manual CLI validation with a test file

**CI Pipeline Requirements:**

- All TypeScript packages must pass type checking
- Code must be properly formatted (Prettier)
- Go tests must pass (except known failing test)
- Linting must not introduce new errors

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

## Common Tasks

The following are outputs from frequently run commands. Reference them instead of viewing, searching, or running bash commands to save time.

### Repository Root Structure

```
.git                    # Git repository
.github/                # GitHub workflows and actions
.gitignore             # Git ignore rules
.gitmodules            # Git submodule configuration
.golangci.yml          # Go linting configuration
.vscode/               # VS Code configuration templates
cmd/                   # Go CLI entry point
internal/              # Go internal packages (rules, linter, API)
node_modules/          # Node.js dependencies (after pnpm install)
packages/              # pnpm workspace packages
pnpm-lock.yaml         # pnpm lockfile
pnpm-workspace.yaml    # pnpm workspace configuration
rslint.json            # Rslint configuration for self-linting
shim/                  # TypeScript-Go shim bindings
tools/                 # Development tools
typescript-go/         # TypeScript compiler Go port (submodule)
```

### Package Structure (`packages/`)

```
rslint/                # Main npm package and CLI
rslint-test-tools/     # Testing utilities and frameworks
rule-tester/           # Rule testing infrastructure
utils/                 # Shared utilities
vscode-extension/      # VS Code editor integration
```

### Internal Structure (`internal/`)

```
api/                   # Public API layer
config/                # Configuration handling
linter/                # Core linting engine
lsp/                   # Language Server Protocol support
rule/                  # Rule definition framework
rule_tester/           # Rule testing utilities
rules/                 # Individual lint rule implementations (50+ rules)
utils/                 # Internal utilities
```

### Key Configuration Files

- `rslint.json` -- Project linting configuration (TypeScript projects, rules)
- `go.mod` -- Go module dependencies with typescript-go shims
- `package.json` -- Root package with workspace scripts
- `pnpm-workspace.yaml` -- Workspace package definitions
- `.github/workflows/ci.yml` -- CI pipeline with Go and Node.js testing

## Common Patterns

### Adding a New Rule

```go
// internal/rules/my_rule/my_rule.go
package my_rule

import (
    "github.com/web-infra-dev/rslint/internal/rule"
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

## Technologies & Languages

### Primary Stack

- **Go 1.24+**: Core linter implementation
- **TypeScript/JavaScript**: Node.js API, tooling, and VS Code extension
- **typescript-go**: TypeScript compiler bindings for Go

### Build Tools

- **pnpm**: Package management (workspace setup)
- **Go modules**: Go dependency management
- **TypeScript**: Compilation and type checking

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

## Compatibility & Migration

- Maintain ESLint rule compatibility where possible
- Provide migration guides for ESLint users
- Support TypeScript-ESLint configuration formats
- Ensure backward compatibility in JavaScript APIs
