# Rslint Project Copilot Instructions

Rslint is a high-performance TypeScript/JavaScript linter written in Go, designed as a drop-in replacement for ESLint and TypeScript-ESLint. It leverages [typescript-go](https://github.com/microsoft/typescript-go) to achieve 20-40x speedup over traditional ESLint setups through native parsing, direct TypeScript AST usage, and parallel processing.

## Code Standards

### Required Before Each Commit

- Run `pnpm install` to setup install
- Run `pnpm run build` to build package
- Run `pnpm run format` to ensure consistent code formatting
- Run `pnpm run typecheck` to verify TypeScript types
- Run `pnpm run lint` to check for linting errors
- Run `pnpm run test` to execute all tests

### Development Flow

- Install: `pnpm install`
- Build: `pnpm run build`
- Format: `pnpm run format`
- Lint: `pnpm run lint`
- Type Check: `pnpm run typecheck`
- Test: `pnpm run test`

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
