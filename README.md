# ✨ Rslint ✨

Rslint aims to be a drop-in replacement for ESLint and TypeScript-ESLint, just as Rspack serves as a drop-in replacement for webpack and may integrated into Rspack in the future.

<p>
  <a href="https://discord.gg/YtTedhuq7N"><img src="https://img.shields.io/badge/chat-discord-blue?style=flat-square&logo=discord&colorA=564341&colorB=EDED91" alt="discord channel" /></a>
  <a href="https://npmjs.com/package/@rslint/core?activeTab=readme"><img src="https://img.shields.io/npm/v/@rslint/core?style=flat-square&colorA=564341&colorB=EDED91" alt="npm version" /></a>
  <a href="https://npmcharts.com/compare/@rslint/core?minimal=true"><img src="https://img.shields.io/npm/dm/@rslint/core.svg?style=flat-square&colorA=564341&colorB=EDED91" alt="downloads" /></a>
  <a href="https://github.com/web-infra-dev/rslint/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square&colorA=564341&colorB=EDED91" alt="license" /></a>
</p>

> [!NOTE]
> Rslint is a fork of [tsgolint](https://github.com/typescript-eslint/tsgolint), an experimental proof-of-concept typescript-go powered JS/TS linter. We would like to express our heartfelt gratitude to the tsgolint team, especially to [@auvred](https://github.com/auvred) for their pioneering work and innovative approach to TypeScript linting. We decided to fork tsgolint because it serves as a proof-of-concept with no current plans for continued development on this direction in the near term ([reference](https://x.com/bradzacher/status/1943475629376282998)).

## Goals and Non-Goals

> [!NOTE]
> Rslint is more like a TypeScript extension than an ESLint plugin — think of it as Rust Clippy for JavaScript.
>
> Rslint is currently in an experimental phase but is under active development.

### Goals

- **Minimal Migration Cost**: Rslint aims to be highly compatible with ESLint and TypeScript-ESLint configurations, significantly reducing the cost of migration.
- **Typed Linting First**: We believe typed linting is essential for advanced semantic analysis. Rslint enables typed linting by default and aims to make it effortless to adopt — no complex setup required.
- **TypeScript Semantics as the Single Source of Truth**: While it’s possible to reimplement TypeScript’s semantics in another language, achieving 100% consistency is extremely difficult and often leads to subtle edge-case bugs—such as discrepancies in resolver behavior, symbol resolution, and project reference support. Rslint avoids these pitfalls by adopting tsgo's native TypeScript semantics for static analysis, ensuring full alignment and eliminating inconsistencies.
- **Project-Level Analysis First**: Unlike ESLint, which defaults to file-level analysis, Rslint performs project-level analysis by default (similar to Clippy). This enables more powerful cross-module analysis and better support for incremental linting.
- **First-Class Monorepo Support**: TypeScript already offers strong monorepo support. Rslint builds on this by following TypeScript’s best practices and the excellent design of typescript-eslint's [project service](https://typescript-eslint.io/blog/project-service) to provide robust support for large-scale monorepos.
- **Batteries Included**: Rslint will include all existing typescript-eslint rules as well as widely used ESLint rules out of the box.
- **Custom Rule Support**: Rslint will expose both the AST, typed information, and global checker data to rule authors, making it easy to write complex cross-module analysis and custom rules.

### Non-Goals

- **Language Agnostic**: We don't plan to support non-TypeScript/JavaScript languages (e.g., CSS or HTML) in the near term — though we remain open to exploring this in the long term.

## 🦀 Rstack

Rstack is a unified JavaScript toolchain built around Rspack, with high performance and consistent architecture.

| Name                                                  | Description              |
| ----------------------------------------------------- | ------------------------ |
| [Rspack](https://github.com/web-infra-dev/rspack)     | Bundler                  |
| [Rsbuild](https://github.com/web-infra-dev/rsbuild)   | Build tool               |
| [Rslib](https://github.com/web-infra-dev/rslib)       | Library development tool |
| [Rspress](https://github.com/web-infra-dev/rspress)   | Static site generator    |
| [Rsdoctor](https://github.com/web-infra-dev/rsdoctor) | Build analyzer           |
| [Rstest](https://github.com/web-infra-dev/rstest)     | Testing framework        |
| [Rslint](https://github.com/web-infra-dev/rslint)     | Linter                   |

## 🤝 Contribution

> [!NOTE]
> We highly value any contributions to Rslint!

Please read the [Contributing Guide](https://github.com/web-infra-dev/rslint/blob/main/CONTRIBUTING.md).

### Code of conduct

This repo has adopted the ByteDance open source code of conduct. Please check [Code of conduct](./CODE_OF_CONDUCT.md) for more details.

## 🙏 Credits

Rslint has been inspired by several outstanding projects in the community. We would like to acknowledge and express our sincere gratitude to the following developers, teams and projects:

- [@auvred](https://github.com/auvred) - The original author of [tsgolint](https://github.com/typescript-eslint/tsgolint), from which Rslint is forked. We are deeply grateful for his pioneering work and innovative approach to TypeScript linting.
- [@JamesHenry](https://github.com/JamesHenry) - The creator of [typescript-eslint](https://github.com/typescript-eslint/typescript-eslint), who has provided valuable guidance and suggestions for Rslint's development.
- The [typescript-eslint](https://github.com/typescript-eslint) team - Rslint's configuration design and test cases have been significantly influenced by and adapted from typescript-eslint's excellent implementation.

## 📖 License

Rslint is licensed under the [MIT License](https://github.com/web-infra-dev/rslint/blob/main/LICENSE).
