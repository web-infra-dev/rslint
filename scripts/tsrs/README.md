# Rust + JavaScript migration tools

The active backend is a native Rust addon called directly by the Node CLI and
API. Normal builds do not launch a compiler process, semantic server, or WASM
bridge. The obsolete bridge packages and their build/test scripts have been
removed; historical evidence remains outside the product repository.

Build and verification instructions are in
[the migration build guide](../../docs/tsrs-migration/build.md). The current
rsbuild workload is frozen in
[rsbuild-workload.json](../../docs/tsrs-migration/rsbuild-workload.json).

- `build-rust.mjs` prepares the pinned Rust compiler and builds the native addon.
- `prepare-rust-compiler.mjs` applies the reviewed library compatibility patch.
- `rust-rsbuild-benchmark.py` verifies the frozen workload and gates paired
  whole-command measurements on correctness evidence.
- `rust-rsbuild-observe.mjs` observes actual admission and native rule execution.
- `benchmark.py` provides the shared process-tree RSS sampler and clock guards.
- `rust-rule-oracle.mjs`, `rust-cli-oracle.py`, and the other `rust-*` oracle
  scripts compare candidate behavior with an explicit external baseline.
- `rust-migration-inventory.py` reads frozen metadata from the checked-in catalog
  without retaining the former Go implementation in the candidate tree.
- `extract-js-ast.py` and `generate-printable-ranges.mjs` reproduce checked-in
  compatibility data without executing the old compiler or requiring its SDK.

The external tsgo baseline is an oracle and benchmark input, never a candidate
fallback. Performance results apply only to the exact admitted files, enabled
rules, configuration, and optimized artifacts recorded by the measurement gate.
