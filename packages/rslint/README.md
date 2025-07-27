# Rslint

🚀 Rocket Speed Linter - A high-performance TypeScript/JavaScript linter written in Go.

Rslint is designed as a drop-in replacement for ESLint and TypeScript-ESLint, leveraging [typescript-go](https://github.com/microsoft/typescript-go) to achieve 20-40x speedup over traditional ESLint setups through native parsing, direct TypeScript AST usage, and parallel processing.

## Features

- ⚡ **Ultra-fast**: 20-40x faster than ESLint through Go-powered parallel processing
- 🎯 **Typed linting first**: Enables typed linting by default for advanced semantic analysis
- 🔧 **Drop-in replacement**: Compatible with ESLint and TypeScript-ESLint configurations
- 🏗️ **Project-level analysis**: Performs cross-module analysis for better linting results
- 📦 **Monorepo support**: First-class support for large-scale TypeScript monorepos
- 🔋 **Batteries included**: Ships with all TypeScript-ESLint rules out of the box

## Installation

```bash
npm install -D @rslint/core
```

## Quick Start

```bash
# Create a rslint.json config file
npx rslint --init

# Lint your project
npx rslint

# See available options
npx rslint --help
```

## Documentation

See the [main repository](https://github.com/web-infra-dev/rslint) for complete documentation and examples.
