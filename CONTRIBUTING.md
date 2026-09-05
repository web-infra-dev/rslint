# Rslint contribution guide

Thank you for your interest in contributing to Rslint! Before you start your contribution, please take a moment to read the following guidelines.

## Setup the environment

Install [Node.js](https://nodejs.org/) and [Go](https://go.dev/) first.

## Build locally

Build the project:

```bash
# Initialize the TypeScript repository (kept at the typescript-go/ path).
git submodule sync -- typescript-go
git submodule update --init --depth 1
pnpm install
go run ./tools/dump_rule_schemas > packages/rslint/rule-schemas.json
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

## TypeScript compiler dependency

The `typescript-go/` submodule tracks [microsoft/TypeScript](https://github.com/microsoft/TypeScript). The Go compiler lives under `typescript-go/tsc/`, and its JS AST API lives under `typescript-go/packages/typescript/`. Existing checkouts must run `git submodule sync -- typescript-go` before updating the submodule to pick up the new remote URL.

The migration pins the latest main-branch commit checked on **2026-09-05**:

| Reference                       | Commit                                                                                                                                |
| ------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| Selected TypeScript main        | [`1f70213d4922b434345f639b441681e470c7cfc1`](https://github.com/microsoft/TypeScript/commit/1f70213d4922b434345f639b441681e470c7cfc1) |
| Latest stable release, `v7.0.2` | [`1e4744d68260a7cb91b62b12edc3f6a2187faaf1`](https://github.com/microsoft/TypeScript/releases/tag/v7.0.2)                             |
| Previous standalone compiler    | `01cbcdd8643cfa17cc8156b60849559c56324601`                                                                                            |
| Corresponding imported history  | [`3e5f89624cdd03a69d6c5bb92d3a352195ad3a41`](https://github.com/microsoft/TypeScript/commit/3e5f89624cdd03a69d6c5bb92d3a352195ad3a41) |

The old commit hash is absent from the new repository because its history was rewritten. Its root tree exactly matches the imported commit's `tsc/` tree (`23a5bf0c01e3f6dc1d4c0caf7aaef69ba9d9de64`). However, that historical commit retains `tsc/_submodules/TypeScript` without a root `.gitmodules` mapping. Recursive Git operations, including the authentication setup in our `actions/checkout` version, fail on it. The stable tag also predates the final repository layout. The selected main commit has the completed layout and removes this obsolete nested submodule.

The Go compiler module and all shim module/import/linkname paths use `github.com/microsoft/TypeScript/tsc`. `go.work` resolves the compiler to `typescript-go/tsc`. Both `@rslint/api` and the Go encoder use this same pinned checkout so that AST layouts and enum values stay aligned.

The JS declaration build uses TypeScript `6.0.3`, matching upstream's build dependency. TypeScript `5.9.3` rejects the new generator-backed API method types. The AST declarations now reference both the synchronous and asynchronous APIs, which are included in `packages/rslint-api/tsconfig.build.json`.

Compiler compatibility changes include the lazy `SourceFile.HasIdentifier` cache, type-reference nodes for interface `extends` and class `implements`, and new compiler-host, resolver, and LSP client interfaces. Rules retain their existing ESLint diagnostics and fix boundaries despite the new node kinds. On macOS, upstream's `Realpath` now preserves path casing; config discovery verifies filesystem identity before coalescing native case aliases. The shim generator also mirrors private generic checker stores and validates their memory layout.

To select a future upstream revision and refresh its integration:

```bash
git submodule sync -- typescript-go
git -C typescript-go fetch --depth 1 origin <commit-or-tag>
git -C typescript-go checkout --detach FETCH_HEAD
bash tools/update-typescript-go.sh
```

The helper resolves the exact submodule commit to a Go pseudo-version, updates the root and shim module requirements, regenerates shims from the local checkout, tidies those modules, builds both Go entrypoints, and checks the unsafe checker mirror's field layout. Review the resulting changes and run the Go and JS tests before committing the new submodule revision. Compiler API changes may require adapting the shim declarations and their consumers.

## Test the CLI

After building, you can test the rslint CLI:

```bash
# Test the binary
./packages/rslint/bin/rslint.js --help


# Lint the project itself
./packages/rslint/bin/rslint.js
```

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
