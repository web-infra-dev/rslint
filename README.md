<h1 align="left">✨ tsgolint ✨</h1>

**tsgolint** is an experimental proof-of-concept [typescript-go](https://github.com/microsoft/typescript-go) powered JS/TS linter written in Go.

## What's done so far

- Primitive linter engine
- Lint rules tester
- Source code fixer
- All recommended type-aware rules from typescript-eslint's [`recommended-type-checked-only`](https://typescript-eslint.io/rules/?=recommended-typeInformation) ruleset
- Basic `tsgolint` CLI

## Speedup over ESLint

TODO

### Implemented rules

| Name | Status |
|---|---|
|await-thenable|✅|
|no-array-delete|✅|
|no-base-to-string|✅|
|no-duplicate-type-constituents|✅|
|no-floating-promises|✅|
|no-for-in-array|✅|
|no-implied-eval|✅|
|no-misused-promises|✅|
|no-redundant-type-constituents|✅|
|no-unnecessary-type-assertion|✅|
|no-unsafe-argument|✅|
|no-unsafe-assignment|✅|
|no-unsafe-call|✅|
|no-unsafe-enum-comparison|✅|
|no-unsafe-member-access|✅|
|no-unsafe-return|✅|
|no-unsafe-unary-minus|✅|
|only-throw-error|✅|
|prefer-promise-reject-errors|✅|
|require-await|✅|
|restrict-plus-operands|✅|
|restrict-template-expressions|✅|
|unbound-method|✅|
