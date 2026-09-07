# Rslint contribution guide

Thank you for your interest in contributing to Rslint! Before you start your contribution, please take a moment to read the following guidelines.

## Setup the environment

Install [Node.js](https://nodejs.org/) and [Go](https://go.dev/) first.

## Build locally

For a targeted change, start with [Verify a change](#verify-a-change) and prepare only the dependencies and artifacts it needs. For a complete local build:

```bash
# Initialize the TypeScript repository (kept at the typescript-go/ path).
git submodule sync -- typescript-go
git submodule update --init --depth 1
pnpm install
go run ./tools/dump_rule_schemas > packages/rslint/rule-schemas.json
pnpm build
```

## Verify a change

Inspect `git status --short --branch` and the branch diff, including staged, unstaged and untracked work. Identify the changed packages and the callers affected by shared API changes, then select the relevant existing commands. The examples below are a menu; run only those needed for the change.

| Check                         | Existing command with explicit scope                                                                   |
| ----------------------------- | ------------------------------------------------------------------------------------------------------ |
| One Go rule package           | `go test ./internal/rules/max_params`                                                                  |
| Related Go packages           | `go test <changed-package-dir> <affected-consumer-dirs>`                                               |
| Catalog registration only     | `go test ./internal/rules -run '^TestAllContainsEveryGoRuleExactlyOnce$'`                              |
| One JS integration file       | `CI=true pnpm --dir packages/rslint-test-tools exec rs test run tests/eslint/rules/max-params.test.ts` |
| One Rust crate                | `cargo test -p tsgo-client`                                                                            |
| Go lint for a changed package | `golangci-lint run --new-from-merge-base=origin/main ./internal/rules/max_params`                      |
| Format changed JS/TS/docs     | `pnpm exec rs fmt <changed-files>`                                                                     |
| Format changed Go files       | `gofmt -w <changed-go-files>`                                                                          |
| Spell-check changed text      | `pnpm run check-spell <changed-text-files>`                                                            |

An aggregate package can run more than its name suggests: unfiltered `go test ./internal/rules` also runs all-rule compiler compatibility and heritage suites. For registration-only changes, keep the named test selection above; select other tests only when their behavior is affected.

Explicit `rs fmt` paths still obey `rstack.config.mts` exclusions, including rule Markdown under `internal/**/rules/**/*.md`. Select supported, non-ignored changed files and skip the command when none remain; an ignored-only selection fails instead of formatting those files.

Run Rstest verification with `CI=true` so missing snapshots fail instead of being created automatically. The example uses POSIX shell syntax; set the equivalent environment variable in other shells. Use `-u` only for intentional snapshot generation and review the result against the expected behavior.

Before JS integration tests exercise changed Go code, run `pnpm --filter @rslint/core build:bin`. Build affected JS artifacts with the workspace's existing build command when its source changed or the required output is missing. Go test-only changes need the owning package's tests. Shared helpers need their affected consumers; package imports are a starting point for tracing the changed API, not a reason to run every rule in a plugin.

Root `test:go` and `lint:go` already include `./cmd/... ./internal/...`; appending another directory adds to that scope. Use `go test` and `golangci-lint run` directly for selected packages. Root `pnpm test` also runs multiple workspaces. Full-suite verification follows an explicit user/reviewer request; CI's command list is not a local checklist.

Keep the passing commands/results in the task's existing progress record. Reuse them while their relevant inputs remain unchanged. Report platform, generated-content or cross-language gaps with the additional targeted checks they need. Documentation-only changes require no language tests unless executable examples, generated content, builds or runtime behavior are affected.

Branch naming and test organization follow [AGENTS.md](./AGENTS.md). Before delivery, check the current branch with `git branch --show-current`, review the upstream/extras split and verify new JS integration files are included in `packages/rslint-test-tools/rstack.config.mts`. Before committing, ensure `pnpm run check-spell <changed-text-files>` and `pnpm run format:check` have passed after the final relevant edit, reusing valid results. Spell-check file arguments also cover changed hidden paths such as `.agents/`; the default globs omit hidden directories.

The existing pre-commit hook runs `rs staged`. Install it through `pnpm run prepare`; `git config --get core.hooksPath` should point to `.rstack/hooks/_`. If a checkout still points to old Husky hooks, inspect them before migrating once with `pnpm exec rs hooks --force`. Keep ordinary `prepare` non-forcing.

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

## Sync release information

After bumping to a new stable version and before publishing, run the release sync
command separately:

```bash
git fetch origin --tags
pnpm sync:version-info
```

This replaces `pnpm sync:rule-releases`. It updates `website/releases.json` with
new rules and the pinned TypeScript commit. The `main` record is refreshed on each
run; when the package version is a new stable release, that release is recorded
as well. Stage any submodule revision change first. An exact upstream release tag supplies
`typescript.releaseVersion`; otherwise the value is `null`. A failed upstream
lookup stops the command without writing the file.

Commit the generated JSON with the release changes. The website uses it for rule
version badges and the TypeScript compiler version table. The table shows `main`
and new releases; versions through `0.9.1` retain only their existing rule history.

`pnpm sync:version-info full` rebuilds rule history from stable tags while preserving
recorded compiler bindings. Fetch all tags before running it.

The scripts live in `scripts/sync-version-info/`: `index.js` is the entry point,
`release.js` maintains rule release history, and `version.js` resolves the
TypeScript commit and release version.

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
