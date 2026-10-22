# Rust and JavaScript migration (in progress)

The migration's missing parse/bind cache, repeated filesystem probes, per-node
rule dispatch, repeated JavaScript path calculations and duplicate CLI file
admission have been fixed. A fresh full-workload proof and 2 warmups + 7
interleaved measurements per backend show full lint falling from **5.056 s to
3.299 s (34.75% less time)**, with sampled process-tree peak RSS falling from
**887.16 MiB to 596.33 MiB (32.78% less)**. The original optimized Go baseline
keeps its default cache. These are end-to-end results for the fixed single-threaded
Rsbuild workload, not an isolated compiler-speed claim.

The acceptance scope is the **129 rules actually enabled by frozen rsbuild**, with all 2394 files, effective configuration, options, severities, type semantics and fixes preserved. The historical 813-rule catalog is provenance, not the current acceptance gate. Previously implemented out-of-scope rules remain; unsupported rules fail explicitly.

The [generated matrix](rsbuild-coverage.json) records 129 source-declared rules in scope, 0 missing, and 10 retained extras. Source declarations are not behavior acceptance. The direct Rust core uses official tsrs `280b426b89cd4ecd6153cf48228826bd60a99c61` and exchanges owned requests/results with JavaScript through N-API. It does not launch a Go backend.

The final Rust/JS candidate passed the complete Rsbuild CLI/API workload proof,
actual rule scheduling, 129 actual-options sentinel pairs (128 genuine rule hits;
the baseline parser guard for no-octal is disclosed), and exact Program source
graph comparison. The current JS build passed 97 consumer tests and 8 native
tests. The unchanged Rust implementation retains its hash-bound 44 focused
regressions and full affected-rule oracle results.
See the [Chinese report](REPORT.zh-CN.md), [current compact samples](rsbuild-performance-fixed.json)
and [raw sample table](rsbuild-performance-fixed.csv). The [old prototype samples](rsbuild-performance.json)
remain historical evidence and do not describe the fixed candidate. This does
not establish default multicore, cross-platform or production-wide completion.

## Reproduce the implemented path

Use the repository's pinned Rust and pnpm versions, Node 24, and the usual
native linker. The new core/addon checks below do not build or execute Go.
Use [the Rust build instructions](build.md) to prepare the pinned compatibility
graph before invoking Cargo directly. Cross-platform and default multicore validation remain outside this result.

```sh
pnpm install --frozen-lockfile --ignore-scripts
CARGO_BUILD_JOBS=2 node scripts/tsrs/build-rust.mjs --jobs 2
pnpm --filter @rslint/api build
pnpm --filter @rslint/core build:js
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

The oracle declares the original native plugin explicitly, checks rule-positive sentinels with the documented parser exception, retains silent files and compares all
message fields, including fix and suggestion payloads for the two non-null
assertion rules. A separate four-rule run compares automatic-fix output and
final diagnostics, including untouched files and suggestion-only diagnostics.
It does not substitute for every original rule's fix behavior or full public
CLI/API fix compatibility. The Go baseline is exclusively a test oracle, outside the
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

The final root CLI/API workload proof and supplemental current-artifact review
passed before timing. The fixed candidate shows **lower wall time and sampled
process-tree RSS**, detailed in the [Chinese report](REPORT.zh-CN.md). Older
two-rule, Go/Rust bridge and initial migration results do not describe it.

To repeat the comparison, use the same files, configuration, dependencies,
rules, fixes, compiler pins and thread/cache conditions. Measure at least two
warmups and seven interleaved A/B repetitions for both `rs lint` and its
rslint stage. Include startup, N-API, shutdown and the whole process tree in
wall time and sampled RSS; retain all samples, including thermal/sleep
contamination. Report medians, spread, absolute differences, percentages and
ratios. Do not time a reduced subset and call it the full result.
