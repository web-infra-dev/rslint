# Rslint contribution guide

This branch migrates the actual Rsbuild lint workload to Rust/tsrs and a direct
JavaScript binding. The acceptance set is its 129 enabled rules, not the historical
813-rule catalog. See [the migration report](docs/tsrs-migration/REPORT.zh-CN.md)
for passed checks, compatibility failures and unsupported entrypoints.

## Setup the environment

Use Node.js 24, the pnpm version in `packageManager`, Rust 1.99.1 and a native
linker. Python 3.9+ is used by the optional comparison and machine-lock tools.
The product build does not require a Go SDK, compiler checkout or Go runtime.

```sh
pnpm install --frozen-lockfile --ignore-scripts
rustup toolchain install 1.99.1
pnpm build
```

The default build produces the Rust addon, API JavaScript and core JavaScript.
Declaration generation uses pinned JavaScript TypeScript, explicitly disabling
the native TypeScript executable. Option schemas for the 129-rule scope are
checked-in data with original-source provenance; missing or changed schema data
fails instead of requesting a source generator.

## Verify a change

Inspect `git status --short --branch` and the branch diff, including staged,
unstaged and untracked changes. Identify owning packages and affected callers
before selecting checks. The commands below are a menu, not a full-suite request.

Prepare the exact Rust dependency graph before running Cargo directly:

```sh
node scripts/tsrs/prepare-rust-compiler.mjs \
  --config-out node_modules/.cache/rslint-rust-compiler/cargo-config.toml
```

| Scope                               | Command                                                                                                                                   |
| ----------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| One core rule suite                 | `cargo --config node_modules/.cache/rslint-rust-compiler/cargo-config.toml test --locked -p rslint-core --test rsbuild_flow_checks_rules` |
| One core module                     | `cargo --config node_modules/.cache/rslint-rust-compiler/cargo-config.toml test --locked -p rslint-core --lib directives::tests`          |
| Rust boundary                       | `cargo --config node_modules/.cache/rslint-rust-compiler/cargo-config.toml test --locked -p rslint-native`                                |
| Build current native code           | `pnpm --filter @rslint/core build:bin`                                                                                                    |
| One JS integration file             | `CI=true pnpm --dir packages/rslint exec rs test run tests/rust-cli.test.ts`                                                              |
| Build JS after entry/config changes | `pnpm --filter @rslint/api build && pnpm --filter @rslint/core build:js`                                                                  |
| Addon build/packaging helpers       | `node --test scripts/tsrs/build-rust.test.mjs scripts/ci-delivery.test.mjs`                                                               |
| Format changed JS/TS/docs           | `pnpm exec rs fmt <changed-files>`                                                                                                        |
| Spell-check changed text            | `pnpm run check-spell <changed-text-files>`                                                                                               |

Rebuild the addon before JS tests exercise changed Rust code. Select the host
`.node` path when running `crates/rslint-native/tests/rust-lint.test.mjs` and set
`RSLINT_NATIVE_BINDING` to its absolute path. Tests can disable subprocess APIs
using `--test-isolation=none --import ./scripts/tsrs/rust-no-child-process.mjs`.
Use `CI=true` with Rstest so missing snapshots fail. Review deliberate snapshot
updates; do not weaken oracle assertions to accept a port difference.

Reuse passing results while their relevant inputs remain unchanged. Shared
helpers require their affected consumers, not every rule in every plugin.
Full repository and whole-plugin tests need an explicit user/reviewer request;
a CI command list is not authorization for unrelated local suites.

Branch and test organization rules are in [AGENTS.md](AGENTS.md). Continue the
existing task branch, preserve unrelated work, and use Conventional Commits.
Before committing, pass `pnpm run check-spell <changed-text-files>` and
`pnpm run format:check`; explicit spell paths include changed hidden files.
The `rs staged` hook can be installed with `pnpm run prepare`.

## Compiler and source ownership

`scripts/tsrs/prepare-rust-compiler.mjs` fetches committed objects at the exact
pin, applies tracked Rust patches and verifies the resulting source cache outside
this repository. An optional `--mirror` supplies committed objects only. It never
reads uncommitted compiler changes or overwrites `.cargo/config.toml`.

The tracked declaration resources, ordered library aliases and target behavior
are aligned to the frozen original engine. The compiler remains Rust; version
and semantic differences must be disclosed. See
[compiler compatibility](docs/tsrs-migration/compiler-compatibility.json).

`scripts/tsrs/build-rust.mjs` always passes `--locked`, validates target/addon
format, and atomically replaces only a successful artifact. Cross compilation
needs the selected Rust target and linker. A failed build does not reuse an old
artifact as a successful result.

## Correctness and performance comparisons

Keep the original engine in an immutable checkout outside the candidate tree.
Use the fixed commits and real configuration in
[the workload manifest](docs/tsrs-migration/rsbuild-workload.json). Preserve all
files, options, severities, type capability and complete diagnostic/fix payloads.
A source registry, project planner or zero-diagnostic project alone is not a
complete semantic oracle; rules also need positive and negative inputs.

The current drivers are `scripts/tsrs/rust-rule-oracle.mjs`,
`rust-cli-oracle.py`, `rust-program-universe-oracle.mjs` and
`rust-rsbuild-benchmark.py`. The benchmark requires a successful complete-workload
proof plus hash-bound build and correctness evidence before measurement. It
records the full command and rslint stage separately, whole-process-tree RSS,
startup/shutdown, all repetitions and sleep/thermal contamination.

On the shared machine, run heavy builds and complete measurement batches through
`scripts/tsrs/with-machine-lock.py --lock "$RSLINT_MACHINE_LOCK" --log LOG -- COMMAND`,
using the same cooperating lock. The benchmark driver takes `--lock` directly;
do not wrap it in a second lock. Keep build parallelism bounded.

## Delivery boundaries

CI verifies Rust/JS on its declared targets; local static workflow checks do not
prove remote CI passed. Current ecosystem acceptance is Rsbuild. Browser WASM,
LSP, external JS rule plugins and legacy compiler-service subpaths are explicitly
unsupported in this scoped migration. Historical release-info generation is not
a valid release gate. Do not claim those paths were migrated or published.
