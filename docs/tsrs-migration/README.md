# Rust and JavaScript migration (draft)

The target is a repository with no Go source, SDK, generated Go, subprocess,
precompiled Go component, or Go WASM dependency. Rust owns the tsrs compiler
and lint engine; JavaScript retains the public API, configuration modules,
plugins and editor integration where appropriate. **This draft does not yet
meet that target.** The default CLI/API still use the original engine.

The new `rslint-core` crate directly holds a tsrs `Session`, snapshot, program,
AST and exclusive checker lease. Rules walk the compiler AST and query that
checker in Rust. The existing Node addon exports `lintTsrsSync`, which passes
owned request/diagnostic JSON across N-API. It does not export an AST graph,
launch a compiler process or fall back to another engine. Checker leases end
before their snapshot/session. A dedicated native thread supplies the compiler
stack; the synchronous entry blocks its calling JavaScript thread.

The compiler is the **unmodified** official `oxc-project/tsrs` commit
`280b426b89cd4ecd6153cf48228826bd60a99c61`, pinned in the manifests and lockfile.
This direct path does not require the API patches from the earlier process
bridge experiment. License and upstream notice copies are in
[`crates/rslint-core/third-party`](../../crates/rslint-core/third-party).

## Current implementation

Twenty-five rules have Rust implementations: twenty-one syntax/control-flow
or binder-based rules and four type-aware rules. Their list and targeted parity evidence are
in [`validation.json`](validation.json). The frozen
[`catalog`](../../scripts/tsrs/rust-migration-catalog.json) enumerates **813**
baseline rules, their full option schemas, source/test hashes and fourteen
public surfaces. The previous count of 812 missed
`import/no-webpack-loader-syntax`. Its inventory status is not a test result.

`packages/rslint/src/rust/index.ts` exposes experimental source-level
`lintRustFilesSync` and `lintRustFiles`. These are separate from the public
`Rslint` API and are not package exports. They support explicit files, uniform
rules and an optional explicit project/config module. Results retain silent
files and the applied rule list, with UTF-16 diagnostic ranges. Unsupported
rules/options, scoped configuration, plugins, ignores, discovery, automatic fixes,
directives and malformed syntax fail explicitly. Lone-surrogate request text
is currently rejected. Proposed fixes/suggestions are returned for implemented
rules; they are not applied by this entry. These limitations are compatibility gaps, not intended
changes to the finished public API.

The remaining rules, all default entry points, configuration/ignore/project
discovery, fix/suggestion execution, plugin contracts, compiler diagnostics,
LSP and browser WASM still require migration and oracle validation. Existing
Go code is retained until its behavior has a verified replacement. A draft
with twenty-five rules is not completion of the full migration.

## Reproduce the implemented path

Use the repository's pinned Rust and pnpm versions, Node 24, and the usual
native linker. The new core/addon checks below do not build or execute Go.
The rest of the repository has not yet reached that condition.

```sh
pnpm install --frozen-lockfile --ignore-scripts
CARGO_BUILD_JOBS=2 cargo test --locked -p rslint-core
CARGO_BUILD_JOBS=2 pnpm --filter @rslint/native build
RSLINT_NATIVE_BINDING="$PWD/npm/rslint/darwin-arm64/rslint.darwin-arm64.node" \
  node --test crates/rslint-native/tests/rust-lint.test.mjs
CI=true pnpm --dir packages/rslint exec rs test run tests/rust-engine.test.ts
```

Select the actual host addon path on other platforms. To verify that the
native tests never launch another process:

```sh
RSLINT_NATIVE_BINDING="$PWD/npm/rslint/darwin-arm64/rslint.darwin-arm64.node" \
  node --test-isolation=none --import ./scripts/tsrs/rust-no-child-process.mjs \
  --test crates/rslint-native/tests/rust-lint.test.mjs
```

Keep the original engine in an external checkout at
`e0a29f751d4def86d27931759098f27748fa3003`, build it according to that revision's
instructions, then run the targeted diagnostic oracle:

```sh
node scripts/tsrs/rust-oracle.mjs "$BASELINE" "$ADDON" target/rust-oracle.json
python3 scripts/tsrs/rust-migration-inventory.py --help
```

The oracle declares the original native plugin explicitly, requires every
rule to produce a positive sentinel, retains silent files and compares all
message fields, including fix and suggestion payloads for the two non-null
assertion rules. It does not substitute for every original test or iterative
autofix compatibility. The Go baseline is exclusively a test oracle, outside the
candidate repository.

On a shared machine, wrap heavy builds and measurements in
`scripts/tsrs/with-machine-lock.py --lock "$RSLINT_MACHINE_LOCK" --log LOG -- COMMAND`.
Keep the same lock across cooperating tasks and limit build concurrency.

## Rsbuild acceptance

[`rsbuild-workload.json`](rsbuild-workload.json) preserves the pinned real
workload, including files that produce no diagnostics. The candidate must
support that entire workload before A/B timing can be accepted. Run:

```sh
node scripts/tsrs/rust-rsbuild-gate.mjs --help
```

The current gate rejects the incomplete candidate. There is **no valid
rsbuild speedup, slowdown or peak-memory comparison for this architecture**.
The older two-rule and Go/Rust bridge measurements do not describe it.
See the [Chinese report](REPORT.zh-CN.md) for the current acceptance state.

Once equivalence passes, use the same files, configuration, dependencies,
rules, fixes, compiler pins and thread/cache conditions. Measure at least two
warmups and seven interleaved A/B repetitions for both `rs lint` and its
rslint stage. Include startup, N-API, shutdown and the whole process tree in
wall time and sampled RSS; retain all samples, including thermal/sleep
contamination. Report medians, spread, absolute differences, percentages and
ratios. Do not time a reduced subset and call it the full result.
