//! Direct JavaScript boundary for the Rust lint engine.
//!
//! This experimental entry point is deliberately separate from the existing
//! public API until the engine implements its full compatibility contract. It
//! never calls the legacy executable or transfers compiler state through Go.

use napi_derive::napi;

/// Lint one JSON request using the in-process tsrs compiler and Rust rules.
///
/// Owned strings are the only values crossing the Node-API boundary. AST nodes,
/// type-checker state, and any compiler thread remain inside `rslint-core`.
/// Engine validation errors become JavaScript exceptions; an unsupported rule
/// must fail explicitly instead of selecting a different backend.
#[napi(catch_unwind, js_name = "lintTsrsSync")]
pub fn lint_tsrs_sync(request: String) -> napi::Result<String> {
    rslint_core::lint_json(&request).map_err(napi::Error::from_reason)
}
