//! `@rslint/native`: shared byte storage and Oxc 0.133 parsing via N-API.
//! The transport owns memory capabilities; the parser is a byte consumer for
//! `@rslint/core`'s ESLint-plugin runtime.
//!
//! One parse produces both the ESTree AST and a parser-driven token stream (oxc's
//! `TokensParserConfig`); `token_map` translates those tokens to the espree/ts-estree
//! ESLint token contract. There is no hand-written lexer -- token disambiguation
//! (`/` regex-vs-division, templates, JSX, TS `<`) comes from real parser state.

mod memory_transport;
mod parse;
#[cfg(feature = "test-worker-termination")]
mod parser_worker_test;
mod token_map;

use napi::{Env, JsString};
use napi_derive::napi;

use memory_transport::SharedBytes;
pub use parse::{CommentObj, ParseResult};

/// Reject sources whose serialized ESTree JSON would exceed V8's ~512MB single-string
/// cap (the JSON is ~9-26x the source size). This is the JSON-transfer ceiling
/// (raw transfer is a future optimization). Surfacing a clear parseError here beats
/// the cryptic "Failed to convert rust String into napi string" that napi would throw.
const MAX_SOURCE_BYTES: usize = 16 * 1024 * 1024;

/// A terminated worker may reject N-API conversion with PendingException even
/// though no ordinary JavaScript exception exists. Return directly to Node in
/// that case: trying to manufacture and throw another error can abort a debug
/// addon. Application callers remain responsible for ignoring cancelled results.
pub struct NativeResult<T>(napi::Result<T>);

impl<T: napi::bindgen_prelude::ToNapiValue> napi::bindgen_prelude::ToNapiValue for NativeResult<T> {
    unsafe fn to_napi_value(
        env: napi::sys::napi_env,
        value: Self,
    ) -> napi::Result<napi::sys::napi_value> {
        match value.0.and_then(|value| T::to_napi_value(env, value)) {
            // Node-API callbacks may return NULL with an exception pending.
            // Preserve the original exception/termination; make no N-API calls.
            Err(error) if error.status == napi::Status::PendingException => {
                Ok(std::ptr::null_mut())
            }
            result => result,
        }
    }
}

/// Parse JS/TS/JSX source -> ESTree JSON + ESLint-shape comments (UTF-16 offsets).
/// Replaces npm `oxc-parser`'s `parseSync`.
///
/// - `filename`: used for lang inference (extension).
/// - `source_type`: `"module"` | `"script"` | `"commonjs"` (commonjs is treated as script).
/// - `jsx`: `languageOptions.parserOptions.ecmaFeatures.jsx`; true promotes `.ts->tsx`/`.js->jsx`.
///
/// The size guard becomes a JS exception, which the JS side maps to a `parseError`
/// (matching the current "parseSync throw -> parseError" contract). `catch_unwind`
/// also turns a Rust panic into a JS exception (the worker survives). Note: a stack
/// overflow (deep nesting) is a SIGSEGV that catch_unwind does NOT catch.
#[napi(catch_unwind, ts_return_type = "ParseResult")]
pub fn parse(
    filename: String,
    source: String,
    source_type: String,
    jsx: bool,
) -> NativeResult<ParseResult> {
    NativeResult(
        check_source_size(source.len())
            .map(|()| parse::parse_estree(&filename, &source, &source_type, jsx)),
    )
}

#[napi(object)]
pub struct SourceParseResult<'env> {
    pub parsed: ParseResult,
    pub source_text: JsString<'env>,
    pub had_bom: bool,
}

/// Parse a source snapshot while its transport lease pins the mapped bytes.
/// Source decoding and BOM handling belong to this parser boundary, not the
/// shared-memory owner. The ESTree JSON result is identical to inline parsing.
#[napi(catch_unwind, ts_return_type = "SourceParseResult")]
pub fn parse_shared_bytes(
    env: &Env,
    filename: String,
    source: SharedBytes,
    source_type: String,
    jsx: bool,
) -> NativeResult<SourceParseResult<'_>> {
    #[cfg(feature = "test-worker-termination")]
    let lease = source.lease;
    NativeResult(memory_transport::with_bytes(source, |bytes| {
        let text = std::str::from_utf8(bytes)
            .map_err(|_| napi::Error::from_reason("invalid shared source UTF-8"))?;
        let had_bom = text.starts_with('\u{feff}');
        let text = text.strip_prefix('\u{feff}').unwrap_or(text);
        check_source_size(text.len())?;
        // The feature-only barrier stops inside this native call, after the
        // reader has pinned and decoded the bytes, but before Oxc reads them.
        #[cfg(feature = "test-worker-termination")]
        parser_worker_test::before_parse(lease);
        let parsed = parse::parse_estree(&filename, text, &source_type, jsx);
        #[cfg(feature = "test-worker-termination")]
        parser_worker_test::after_parse(lease, &parsed.program);
        // N-API copies borrowed UTF-8 directly into the required JS SourceCode
        // string, with no temporary Rust String or JS -> Rust round trip.
        let source_text = env.create_string(text)?;
        Ok(SourceParseResult {
            parsed,
            source_text,
            had_bom,
        })
    }))
}

fn check_source_size(size: usize) -> napi::Result<()> {
    if size > MAX_SOURCE_BYTES {
        return Err(napi::Error::from_reason(format!(
            "source too large ({} bytes > {}-byte JSON-transfer limit)",
            size, MAX_SOURCE_BYTES
        )));
    }
    Ok(())
}
