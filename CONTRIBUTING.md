# Rslint contribution guide

Thank you for your interest in contributing to Rslint! Before you start your contribution, please take a moment to read the following guidelines.

## Setup the environment

Install [Node.js](https://nodejs.org/) and [Go](https://go.dev/) first.

## Build locally

Build the project:

```bash
# init typescript-go submodule
git submodule update --init --recursive
pnpm install
pnpm build
```

Test the setup:

```bash
# Run all tests
pnpm test

# Run Go tests only
pnpm run test:go

# Run linting
pnpm run lint

# Run type checking
pnpm run typecheck

# Check code formatting
pnpm run format:check
```

## Test the CLI

After building, you can test the rslint CLI:

```bash
# Test the binary
./packages/rslint/bin/rslint --help


# Lint the project itself
./packages/rslint/bin/rslint --config rslint.json
```

## Known Limitations

### Test Infrastructure

The test runner in `/packages/rule-tester` has a limitation where it cannot pass TypeScript compiler options from individual test cases to the Go implementation. This affects rules that need to behave differently based on compiler options like `noPropertyAccessFromIndexSignature`.

**Issue**: The test runner always uses the same `rslint.json` config file for all test cases, which references a fixed `tsconfig.json`. Individual test cases can specify different tsconfig files via `languageOptions.parserOptions.project`, but these are ignored by the test runner.

**Impact**: Tests that expect different behavior based on TypeScript compiler options may fail or require workarounds.

**Location**: See `/packages/rule-tester/src/index.ts` line 273 - the `lint()` call doesn't use per-test configuration.

## Debugging VSCode Extension

To Debug the VSCode Extension:

1. **Setup launch configuration**

```bash
cp .vscode/launch.template.json .vscode/launch.json
```

2. **Start debugging**

- Open the Command Palette (`Cmd+Shift+P`)
- Run `Debug: Start Debugging` or press `F5`
- Alternatively, go to the `Run and Debug` sidebar and select `Run Extension`
