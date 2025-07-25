# ✨ Rslint ✨

Rslint aims to be a drop-in replacement for ESLint and TypeScript-ESLint, just as Rspack serves as a drop-in replacement for webpack and may integrated into Rspack in the future.

> [!NOTE]
> Rslint is a fork of [tsgolint](https://github.com/typescript-eslint/tsgolint), an experimental proof-of-concept typescript-go powered JS/TS linter. We would like to express our heartfelt gratitude to the tsgolint team, especially to [@auvred](https://github.com/auvred) for their pioneering work and innovative approach to TypeScript linting. We decided to fork tsgolint because it serves as a proof-of-concept with no current plans for continued development on this direction in the near term ([reference](https://x.com/bradzacher/status/1943475629376282998)).

## Goals and Non-Goals

> [!NOTE]
> Rslint is more like a TypeScript extension than an ESLint plugin — think of it as Rust Clippy for JavaScript.
>
> Rslint is currently in an experimental phase but is under active development.

### Goals

- Typed Linting First: We believe typed linting is essential for advanced semantic analysis. Rslint enables typed linting by default and aims to make it effortless to adopt — no complex setup required.
- TypeScript Semantics as the Single Source of Truth:
  While it’s possible to reimplement TypeScript’s semantics in another language, achieving 100% consistency is extremely difficult and often leads to subtle edge-case bugs—such as discrepancies in resolver behavior, symbol resolution, and project reference support. Rslint avoids these pitfalls by adopting tsgo’s native TypeScript semantics for static analysis, ensuring full alignment and eliminating inconsistencies.
- Project-Level Analysis First: Unlike ESLint, which defaults to file-level analysis, Rslint performs project-level analysis by default (similar to Clippy). This enables more powerful cross-module analysis and better support for incremental linting.
- First-Class Monorepo Support: TypeScript already offers strong monorepo support. Rslint builds on this by following TypeScript’s best practices and the excellent design of typescript-eslint's [project service](https://typescript-eslint.io/blog/project-service) to provide robust support for large-scale monorepos.

- Batteries Included: Rslint will include all existing typescript-eslint rules as well as widely used ESLint rules out of the box.

- Custom Rule Support: Rslint will expose both the AST, typed information, and global checker data to rule authors, making it easy to write complex cross-module analysis and custom rules.

### Non-Goals

- Language Agnostic: We don’t plan to support non-TypeScript/JavaScript languages (e.g., CSS or HTML) in the near term — though we remain open to exploring this in the long term.

## 🤝 Contribution

> [!NOTE]
> We highly value any contributions to Rslint!

Please read the [Contributing Guide](https://github.com/web-infra-dev/rslint/blob/main/CONTRIBUTING.md).

### Code of conduct

This repo has adopted the ByteDance open source code of conduct. Please check [Code of conduct](./CODE_OF_CONDUCT.md) for more details.
