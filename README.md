# ✨ Rslint ✨

Rslint aims to be a drop-in replacement for ESLint and TypeScript-ESLint, just as Rspack serves as a drop-in replacement for webpack and may integrated into Rspack in the future.

> [!NOTE]
> Rslint is a fork of [tsgolint](https://github.com/typescript-eslint/tsgolint), an experimental proof-of-concept typescript-go powered JS/TS linter. We would like to express our heartfelt gratitude to the tsgolint team, especially to [@auvred](https://github.com/auvred) for their pioneering work and innovative approach to TypeScript linting. We decided to fork tsgolint because it serves as a proof-of-concept with no current plans for continued development on this direction in the near term ([reference](https://x.com/bradzacher/status/1943475629376282998)).

Based on [tsgolint](https://github.com/typescript-eslint/tsgolint) early exploration.

**tsgolint** is an experimental proof-of-concept [typescript-go](https://github.com/microsoft/typescript-go) powered JS/TS linter written in Go.

> [!IMPORTANT] > **tsgolint** is a prototype in the early stages of development.
> It is not actively being worked on, nor is it expected to be production ready.
> See **[Goals and Non-Goals](#goals-and-non-goals)**.

![Running tsgolint on microsoft/typescript repo](./docs/record.gif)

## What's been prototyped

- Primitive linter engine
- Lint rules tester
- Source code fixer
- 40 [type-aware](https://typescript-eslint.io/blog/typed-linting) typescript-eslint's rules
- Basic `tsgolint` CLI

### Speedup over ESLint

**tsgolint** is **20-40 times faster** than ESLint + typescript-eslint.

Most of the speedup is due to the following facts:

- Native speed parsing and type-checking (thanks to [typescript-go](https://github.com/microsoft/typescript-go))
- No more [TS AST -> ESTree AST](https://typescript-eslint.io/blog/asts-and-typescript-eslint/#ast-formats) conversions. TS AST is directly used in rules.
- Parallel parsing, type checking and linting. **tsgolint** uses all available CPU cores.

See [benchmarks](./benchmarks/README.md) for more info.

### Implemented rules

| Name                                                                                                                | Status |
| ------------------------------------------------------------------------------------------------------------------- | ------ |
| [await-thenable](https://typescript-eslint.io/rules/await-thenable)                                                 | ✅     |
| [no-array-delete](https://typescript-eslint.io/rules/no-array-delete)                                               | ✅     |
| [no-base-to-string](https://typescript-eslint.io/rules/no-base-to-string)                                           | ✅     |
| [no-confusing-void-expression](https://typescript-eslint.io/rules/no-confusing-void-expression)                     | ✅     |
| [no-duplicate-type-constituents](https://typescript-eslint.io/rules/no-duplicate-type-constituents)                 | ✅     |
| [no-floating-promises](https://typescript-eslint.io/rules/no-floating-promises)                                     | ✅     |
| [no-for-in-array](https://typescript-eslint.io/rules/no-for-in-array)                                               | ✅     |
| [no-implied-eval](https://typescript-eslint.io/rules/no-implied-eval)                                               | ✅     |
| [no-meaningless-void-operator](https://typescript-eslint.io/rules/no-meaningless-void-operator)                     | ✅     |
| [no-misused-promises](https://typescript-eslint.io/rules/no-misused-promises)                                       | ✅     |
| [no-misused-spread](https://typescript-eslint.io/rules/no-misused-spread)                                           | ✅     |
| [no-mixed-enums](https://typescript-eslint.io/rules/no-mixed-enums)                                                 | ✅     |
| [no-redundant-type-constituents](https://typescript-eslint.io/rules/no-redundant-type-constituents)                 | ✅     |
| [no-unnecessary-boolean-literal-compare](https://typescript-eslint.io/rules/no-unnecessary-boolean-literal-compare) | ✅     |
| [no-unnecessary-template-expression](https://typescript-eslint.io/rules/no-unnecessary-template-expression)         | ✅     |
| [no-unnecessary-type-arguments](https://typescript-eslint.io/rules/no-unnecessary-type-arguments)                   | ✅     |
| [no-unnecessary-type-assertion](https://typescript-eslint.io/rules/no-unnecessary-type-assertion)                   | ✅     |
| [no-unsafe-argument](https://typescript-eslint.io/rules/no-unsafe-argument)                                         | ✅     |
| [no-unsafe-assignment](https://typescript-eslint.io/rules/no-unsafe-assignment)                                     | ✅     |
| [no-unsafe-call](https://typescript-eslint.io/rules/no-unsafe-call)                                                 | ✅     |
| [no-unsafe-enum-comparison](https://typescript-eslint.io/rules/no-unsafe-enum-comparison)                           | ✅     |
| [no-unsafe-member-access](https://typescript-eslint.io/rules/no-unsafe-member-access)                               | ✅     |
| [no-unsafe-return](https://typescript-eslint.io/rules/no-unsafe-return)                                             | ✅     |
| [no-unsafe-type-assertion](https://typescript-eslint.io/rules/no-unsafe-type-assertion)                             | ✅     |
| [no-unsafe-unary-minus](https://typescript-eslint.io/rules/no-unsafe-unary-minus)                                   | ✅     |
| [non-nullable-type-assertion-style](https://typescript-eslint.io/rules/non-nullable-type-assertion-style)           | ✅     |
| [only-throw-error](https://typescript-eslint.io/rules/only-throw-error)                                             | ✅     |
| [prefer-promise-reject-errors](https://typescript-eslint.io/rules/prefer-promise-reject-errors)                     | ✅     |
| [prefer-reduce-type-parameter](https://typescript-eslint.io/rules/prefer-reduce-type-parameter)                     | ✅     |
| [prefer-return-this-type](https://typescript-eslint.io/rules/prefer-return-this-type)                               | ✅     |
| [promise-function-async](https://typescript-eslint.io/rules/promise-function-async)                                 | ✅     |
| [related-getter-setter-pairs](https://typescript-eslint.io/rules/related-getter-setter-pairs)                       | ✅     |
| [require-array-sort-compare](https://typescript-eslint.io/rules/require-array-sort-compare)                         | ✅     |
| [require-await](https://typescript-eslint.io/rules/require-await)                                                   | ✅     |
| [restrict-plus-operands](https://typescript-eslint.io/rules/restrict-plus-operands)                                 | ✅     |
| [restrict-template-expressions](https://typescript-eslint.io/rules/restrict-template-expressions)                   | ✅     |
| [return-await](https://typescript-eslint.io/rules/return-await)                                                     | ✅     |
| [switch-exhaustiveness-check](https://typescript-eslint.io/rules/switch-exhaustiveness-check)                       | ✅     |
| [unbound-method](https://typescript-eslint.io/rules/unbound-method)                                                 | ✅     |
| [use-unknown-in-catch-callback-variable](https://typescript-eslint.io/rules/use-unknown-in-catch-callback-variable) | ✅     |

## What hasn't been prototyped

- Non-type-aware rules
- Editor extension
- Rich CLI features
- Config file
- Plugin system

### What about JS plugins?

JS-based plugins are not currently supported.

- Experimental support is available on the [`experimental-eslint-compat`](https://github.com/typescript-eslint/tsgolint/tree/experimental-eslint-compat) branch using the [goja](https://github.com/dop251/goja) JavaScript engine.
- While functional, performance was significantly worse than ESLint running in Node.js, so this approach is currently on hold.
- If a faster, lower-allocation JS interpreter in Go becomes available in the future, we may revisit this idea.

## Goals and Non-Goals

**tsgolint** is an experiment.
It is not under active development.

**Goals**: to explore architectures and performance characteristics.
We want to investigate how much faster linting could be if we moved the linter to Go alongside typescript-go.

**Non-Goals**: we have no plans to take significant development budget away from typescript-eslint to work on tsgolint.
Our plan is to continue to work on typescript-eslint to supported typed linting with ESLint.
Experiments such as tsgolint should not be taken as indications of any project direction.

> If you want faster typed linting with ESLint, see [typescript-eslint/typescript-eslint#10940 Enhancement: Use TypeScript's Go port (tsgo / typescript-go) for type information](https://github.com/typescript-eslint/typescript-eslint/issues/10940).

## 🤝 Contribution

> [!NOTE]
> We highly value any contributions to Rslint!

Please read the [Contributing Guide](https://github.com/web-infra-dev/rslint/blob/main/CONTRIBUTING.md).

### Code of conduct

This repo has adopted the ByteDance open source code of conduct. Please check [Code of conduct](./CODE_OF_CONDUCT.md) for more details.

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
