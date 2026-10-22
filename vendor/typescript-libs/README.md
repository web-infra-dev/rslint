# Frozen TypeScript library declarations

These 108 declaration files are copied byte for byte from
[microsoft/TypeScript at `1f70213d4922b434345f639b441681e470c7cfc1`](https://github.com/microsoft/TypeScript/tree/1f70213d4922b434345f639b441681e470c7cfc1/tsc/internal/bundled/libs),
the compiler-library revision used by the frozen rslint comparison baseline.
They total 3,785,075 bytes. The original copyright headers, `LICENSE.txt`,
`NOTICE.txt`, and bundled `CopyrightNotice.txt` are preserved. No Go source or
executable is included or required.

`manifest.json` records every upstream path, byte count, and SHA256.
`lib-map.json` preserves the baseline's 107 library-name mappings in insertion
order. That order also determines the compiler's library priority; it is not
alphabetized. The manifest records the source-file hash for this mapping.

The pinned Rust compiler follows a newer TypeScript revision. Its ESNext
libraries introduce `Math.sumPrecise`, `JSON.rawJSON`, `JSON.isRawJSON`,
`Iterator.concat`, and additional JSON types/overloads. Those declarations change
type-aware lint results: the baseline reports four unsafe calls for the first
four examples while the unmodified newer libraries report none. The compatibility
layer uses these frozen resources without changing the user's `lib` or `target`
options or the source being linted.

Prepare the Rust compiler and an explicit Cargo override:

```sh
node scripts/tsrs/prepare-rust-compiler.mjs
cargo --config node_modules/.cache/rslint-rust-compiler/cargo-config.toml test -p rslint-core
```

Preparation downloads the exact public tsrs commit into an isolated source cache
outside the consuming repository. It patches only the library/target maps,
copies these declarations, and generates Rust `include_str!` tables. All tsrs
workspace packages are redirected to that same checkout so compiler pointers and
arena types cannot come from mixed graphs. Existing workspace Cargo configuration
and external checkouts are left intact. The consumer's lockfile must record this
prepared graph before using `cargo --locked`.

The small Rust patch restores the old names for `esnext.array`,
`esnext.collection`, `esnext.error`, and `esnext.typedarrays`, and maps
`esnext.iterator` back to `lib.es2025.iterator.d.ts`. ES2026 library and target
options, which the baseline does not support, are removed. `ESNext` continues to
select `lib.esnext.full.d.ts`. The compiler implementation and its reported
version remain at the pinned tsrs revision; this is a library compatibility
layer, not a claim that every compiler semantic is identical to the baseline.

Verify the vendored data against the public upstream commit:

```sh
node scripts/tsrs/prepare-rust-compiler.mjs --validate-inputs --verify-upstream
```

Optional `--mirror` and `--typescript-mirror` paths read committed Git objects
only, allowing offline preparation and provenance checks without using local
uncommitted modifications. `--check` validates an existing prepared checkout,
resource set, patch fingerprint, and Cargo override without modifying them.
Unexpected modifications fail verification instead of being reset or overwritten.
