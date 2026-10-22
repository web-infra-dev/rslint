<picture>
  <img alt="Rslint Banner" src="https://assets.rspack.rs/rslint/rslint-banner.png">
</picture>

# Rslint

<p>
  <a href="https://discord.gg/YtTedhuq7N"><img src="https://img.shields.io/badge/chat-discord-blue?style=flat-square&logo=discord&colorA=564341&colorB=EDED91" alt="discord channel" /></a>
  <a href="https://npmjs.com/package/@rslint/core?activeTab=readme"><img src="https://img.shields.io/npm/v/@rslint/core?style=flat-square&colorA=564341&colorB=EDED91" alt="npm version" /></a>
  <a href="https://npmcharts.com/compare/@rslint/core?minimal=true"><img src="https://img.shields.io/npm/dm/@rslint/core.svg?style=flat-square&colorA=564341&colorB=EDED91" alt="downloads" /></a>
  <a href="https://github.com/web-infra-dev/rslint/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square&colorA=564341&colorB=EDED91" alt="license" /></a>
  <a href="https://codspeed.io/web-infra-dev/rslint?utm_source=badge"><img src="https://img.shields.io/endpoint?url=https://codspeed.io/badge.json" alt="CodSpeed"/></a>
</p>

Rslint is an ESLint-compatible linter for JavaScript and TypeScript.

Powered by the pinned [tsrs compiler](https://github.com/oxc-project/tsrs), the current migration executes lint rules, parsing and type analysis in Rust through a direct Node.js binding. Its acceptance scope is the 129 rules used by the frozen Rsbuild workload; see the [migration report](docs/tsrs-migration/REPORT.zh-CN.md) for validation and limitations.

## ✨ Goals

- 🚀 **Performance**: Measure complete lint time and memory use, including Node/native startup, binding costs and shutdown; report regressions as well as improvements.
- ⚡ **Minimal Configuration**: Typed linting enabled by default with minimal setup required — no complex configuration needed.
- 📦 **Best Effort ESLint Compatible**: Compatible with most ESLint and TypeScript-ESLint configurations, significantly reducing migration costs.
- 🎯 **TypeScript First**: Uses compiler semantics for type-aware rules, with compatibility checked through rule and integration tests.
- 🛠️ **Project-Level Analysis**: Performs cross-module analysis by default, enabling more powerful semantic analysis than file-level linting.
- 🏢 **Monorepo Ready**: First-class support for large-scale monorepos with TypeScript project references and workspace configurations.
- 📋 **Batteries Included**: The experimental Rust path covers the 129 rules enabled by the frozen Rsbuild workload.
- 🔧 **Extensible**: Provides direct JavaScript CLI/API access to the Rust engine; external plugin and service APIs remain unsupported.

## ✅ Current Status

This branch is an experimental Rust/JavaScript migration scoped to Rsbuild’s 129 enabled rules. The fixed whole-repository workload and CLI/API results match the original backend in the recorded checks. Its measured single-threaded full-lint median is 13.11 s versus 4.89 s for the original, with higher sampled process-tree RSS. Unsupported rules and plugin/service/LSP/WASM entrypoints fail explicitly. See the [report](docs/tsrs-migration/REPORT.zh-CN.md) for exact evidence and remaining limitations.

## 🚀 Getting Started

See [Guide](https://rslint.rs/guide/).

## 📖 Architecture Overview

- [Architecture Overview](./architecture.md) - Comprehensive system architecture and implementation details

## 🦀 Rstack

Rslint is part of Rstack, the fast, unified JavaScript toolchain for developers and agents.

| Name                                                  | Description              | Version                                                                                                                                                                          |
| ----------------------------------------------------- | ------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [Rspack](https://github.com/web-infra-dev/rspack)     | Bundler                  | <a href="https://npmjs.com/package/@rspack/core"><img src="https://img.shields.io/npm/v/@rspack/core?style=flat-square&colorA=564341&colorB=EDED91" alt="npm version" /></a>     |
| [Rsbuild](https://github.com/web-infra-dev/rsbuild)   | Build tool               | <a href="https://npmjs.com/package/@rsbuild/core"><img src="https://img.shields.io/npm/v/@rsbuild/core?style=flat-square&colorA=564341&colorB=EDED91" alt="npm version" /></a>   |
| [Rslib](https://github.com/web-infra-dev/rslib)       | Library development tool | <a href="https://npmjs.com/package/@rslib/core"><img src="https://img.shields.io/npm/v/@rslib/core?style=flat-square&colorA=564341&colorB=EDED91" alt="npm version" /></a>       |
| [Rspress](https://github.com/web-infra-dev/rspress)   | Static site generator    | <a href="https://npmjs.com/package/@rspress/core"><img src="https://img.shields.io/npm/v/@rspress/core?style=flat-square&colorA=564341&colorB=EDED91" alt="npm version" /></a>   |
| [Rsdoctor](https://github.com/web-infra-dev/rsdoctor) | Build analyzer           | <a href="https://npmjs.com/package/@rsdoctor/core"><img src="https://img.shields.io/npm/v/@rsdoctor/core?style=flat-square&colorA=564341&colorB=EDED91" alt="npm version" /></a> |
| [Rstest](https://github.com/web-infra-dev/rstest)     | Testing framework        | <a href="https://npmjs.com/package/@rstest/core"><img src="https://img.shields.io/npm/v/@rstest/core?style=flat-square&colorA=564341&colorB=EDED91" alt="npm version" /></a>     |
| [Rslint](https://github.com/web-infra-dev/rslint)     | Linter                   | <a href="https://npmjs.com/package/@rslint/core"><img src="https://img.shields.io/npm/v/@rslint/core?style=flat-square&colorA=564341&colorB=EDED91" alt="npm version" /></a>     |

## 🤝 Contribution

Please read the [Contributing Guide](https://github.com/web-infra-dev/rslint/blob/main/CONTRIBUTING.md) and let's build Rslint together.

### Code of Conduct

This repo has adopted the ByteDance Open Source Code of Conduct. Please check [Code of conduct](./CODE_OF_CONDUCT.md) for more details.

## 💬 Community

Come chat with us on [Discord](https://discord.gg/uPSudkun2b)! Rslint team and users are active there, and we're always looking for contributions.

## 🙏 Credits

Rslint has been inspired by several outstanding projects in the community:

- [@auvred](https://github.com/auvred) - The original author of [tsgolint](https://github.com/typescript-eslint/tsgolint), from which Rslint is forked. We are deeply grateful for his pioneering work and innovative approach to TypeScript linting.
- [@JamesHenry](https://github.com/JamesHenry) - The creator of [typescript-eslint](https://github.com/typescript-eslint/typescript-eslint), who has provided valuable guidance and suggestions for Rslint's development.
- [@JoshuaKGoldberg](https://github.com/JoshuaKGoldberg) - For his insightful blog series ["If I Wrote a Linter"](https://www.joshuakgoldberg.com/blog/if-i-wrote-a-linter-part-1-architecture/) which provided valuable architectural insights for modern linter design.
- The [typescript-eslint](https://github.com/typescript-eslint) team - Rslint's configuration design and test cases have been significantly influenced by and adapted from typescript-eslint's excellent implementation.
- The [ESLint](https://github.com/eslint/eslint) team - Rslint builds upon the foundational work of ESLint, the pioneering JavaScript linter that established the standards and patterns for static code analysis in the JavaScript ecosystem.
- The [Rust Clippy](https://github.com/rust-lang/rust-clippy) team - Rslint draws inspiration from Clippy's approach to compiler-integrated linting, bringing similar TypeScript-native analysis to the JavaScript ecosystem.
- The [TypeScript](https://github.com/microsoft/TypeScript) and [tsrs](https://github.com/oxc-project/tsrs) projects - Provide the compiler semantics, Rust implementation, and upstream data models used by Rslint.

## 📖 License

Rslint is [MIT licensed](https://github.com/web-infra-dev/rslint/blob/main/LICENSE).
