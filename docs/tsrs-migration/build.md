# Build the Rust and JavaScript candidate

The active Cargo workspace contains `rslint-core` and `rslint-native`. The core
owns tsrs compiler state and lint rules; the native crate exposes the core through
N-API. The JavaScript CLI and API call the addon directly.

Install the Node.js and pnpm versions declared by `package.json`, and the Rust
toolchain pinned by `rust-toolchain.toml`. Git is required to fetch the pinned
compiler source. The product build below does not invoke Go or build a compiler
service, Go executable, or WebAssembly bridge.

```bash
pnpm install --frozen-lockfile
node scripts/tsrs/build-rust.mjs --plan
pnpm --filter @rslint/core build:bin
pnpm --filter @rslint/api build
pnpm --filter @rslint/core build:js
```

`build:bin` prepares the compiler graph, runs an optimized Cargo build with
`--locked`, and copies the successful addon to its host native npm package. It
defaults to two Cargo jobs. `--platform` and `--target` select supported cross
compilation targets; their Rust targets and linkers must be installed separately.
The script records tool versions and source, lockfile, and artifact hashes under
`target/rust-js/`.

The compiler pin is `280b426b89cd4ecd6153cf48228826bd60a99c61`. Preparation applies
the checked-in Rust compatibility patches and verifies the vendored TypeScript
libraries from `1f70213d4922b434345f639b441681e470c7cfc1`. The source cache is
isolated from mutable development checkouts. An optional `--mirror` reads only
committed objects from a local tsrs clone.

## Direct Cargo commands

Preparation writes an ignored configuration file. Pass it explicitly to Cargo;
the repository's `.cargo/config.toml` must not contain machine-specific compiler
patch paths.

```bash
node scripts/tsrs/prepare-rust-compiler.mjs \
  --config-out node_modules/.cache/rslint-rust-compiler/cargo-config.toml
cargo --config node_modules/.cache/rslint-rust-compiler/cargo-config.toml \
  test --locked -p rslint-core --jobs 2
```

Choose affected test targets when verifying a narrower change. The complete core
command above covers the candidate lint engine. `node scripts/tsrs/build-rust.mjs
--test` runs the core and native crate tests with the same prepared graph and
locked dependencies.

On a shared benchmark machine, wrap builds and complete measurement batches in
the existing machine lock. Set `RSLINT_MACHINE_LOCK` to the shared lock file;
Python is an auxiliary lock runner, not a product dependency.

```bash
python3 scripts/tsrs/with-machine-lock.py \
  --lock "$RSLINT_MACHINE_LOCK" --lock-timeout 120 --timeout 1800 \
  --log target/rust-build.log -- \
  node scripts/tsrs/build-rust.mjs --jobs 2
```

## Current validation boundary

The active workspace has only the Rust core and addon. Historical Go sources,
compiler-service and WASM bridge crates, generators and build actions have been
removed after external archival. The original baseline belongs in an independent
checkout outside this repository and is used only as a test oracle and benchmark.
Unsupported LSP, browser WASM and plugin/service APIs fail explicitly.

The root optimized addon and both JavaScript packages have built successfully.
The complete single-threaded Rsbuild Program source universe matches the baseline,
including declaration content, library flags and metadata. These results do not
by themselves prove clean-checkout installation or repeated performance. See
[the migration status](README.md) for the exact remaining gates and test scope.
