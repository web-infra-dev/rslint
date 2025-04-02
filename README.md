# ✨ tsgolint ✨

**tsgolint** is an experimental proof-of-concept [typescript-go](https://github.com/microsoft/typescript-go) powered JS/TS linter written in Go.

![Running tsgolint on microsoft/typescript repo](./docs/record.gif)

## What's done so far

- Primitive linter engine
- Lint rules tester
- Source code fixer
- All recommended type-aware rules from typescript-eslint's [`recommended-type-checked-only`](https://typescript-eslint.io/rules/?=recommended-typeInformation) ruleset
- Basic `tsgolint` CLI

## Speedup over ESLint

tsgolint is **20-40 times faster** than ESLint + typescript-eslint.

See [benchmarks](./benchmarks/README.md) for more info.

## Implemented rules

| Name                                                                                                | Status |
| --------------------------------------------------------------------------------------------------- | ------ |
| [await-thenable](https://typescript-eslint.io/rules/await-thenable)                                 | ✅     |
| [no-array-delete](https://typescript-eslint.io/rules/no-array-delete)                               | ✅     |
| [no-base-to-string](https://typescript-eslint.io/rules/no-base-to-string)                           | ✅     |
| [no-duplicate-type-constituents](https://typescript-eslint.io/rules/no-duplicate-type-constituents) | ✅     |
| [no-floating-promises](https://typescript-eslint.io/rules/no-floating-promises)                     | ✅     |
| [no-for-in-array](https://typescript-eslint.io/rules/no-for-in-array)                               | ✅     |
| [no-implied-eval](https://typescript-eslint.io/rules/no-implied-eval)                               | ✅     |
| [no-misused-promises](https://typescript-eslint.io/rules/no-misused-promises)                       | ✅     |
| [no-redundant-type-constituents](https://typescript-eslint.io/rules/no-redundant-type-constituents) | ✅     |
| [no-unnecessary-type-assertion](https://typescript-eslint.io/rules/no-unnecessary-type-assertion)   | ✅     |
| [no-unsafe-argument](https://typescript-eslint.io/rules/no-unsafe-argument)                         | ✅     |
| [no-unsafe-assignment](https://typescript-eslint.io/rules/no-unsafe-assignment)                     | ✅     |
| [no-unsafe-call](https://typescript-eslint.io/rules/no-unsafe-call)                                 | ✅     |
| [no-unsafe-enum-comparison](https://typescript-eslint.io/rules/no-unsafe-enum-comparison)           | ✅     |
| [no-unsafe-member-access](https://typescript-eslint.io/rules/no-unsafe-member-access)               | ✅     |
| [no-unsafe-return](https://typescript-eslint.io/rules/no-unsafe-return)                             | ✅     |
| [no-unsafe-unary-minus](https://typescript-eslint.io/rules/no-unsafe-unary-minus)                   | ✅     |
| [only-throw-error](https://typescript-eslint.io/rules/only-throw-error)                             | ✅     |
| [prefer-promise-reject-errors](https://typescript-eslint.io/rules/prefer-promise-reject-errors)     | ✅     |
| [require-await](https://typescript-eslint.io/rules/require-await)                                   | ✅     |
| [restrict-plus-operands](https://typescript-eslint.io/rules/restrict-plus-operands)                 | ✅     |
| [restrict-template-expressions](https://typescript-eslint.io/rules/restrict-template-expressions)   | ✅     |
| [switch-exhaustiveness-check](https://typescript-eslint.io/rules/switch-exhaustiveness-check)       | ✅     |
| [unbound-method](https://typescript-eslint.io/rules/unbound-method)                                 | ✅     |

## Building `tsgolint`

```bash
git submodule update --init                       # init typescript-go submodule

cd typescript-go
git am --3way --no-gpg-sign ../patches/*.patch    # apply typescript-go patches
cd ..

go build -o tsgolint ./cmd/tsgolint                        # build tsgolint
```
