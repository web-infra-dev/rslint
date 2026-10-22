# Rust compiler compatibility patches

Apply these patches only to oxc-project/tsrs commit
`280b426b89cd4ecd6153cf48228826bd60a99c61` with
`scripts/tsrs/prepare-rust-compiler.mjs`. The script verifies every patch and
resource fingerprint and produces a separate Cargo override. This directory is
independent of the superseded Go bridge's compiler preparation.

`0001-baseline-library-map.patch` restores the exact ordered library-name map and
target acceptance of the frozen TypeScript library pin. The Node preparer copies
the licensed declarations from `vendor/typescript-libs` and regenerates the two
Rust embedding tables. It removes the newer ES2026 resources from its own freshly
created compiler checkout, so they cannot remain reachable by a direct library
filename. No compiler work queue, source text, project configuration, checker
algorithm, or tracing behavior is changed.
