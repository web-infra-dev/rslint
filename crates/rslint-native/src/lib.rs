//! `@rslint/native`: oxc 0.133 parse exposed via napi to `@rslint/core`'s
//! ESLint-plugin runtime, replacing npm `oxc-parser` and the JS tokenizer.
//!
//! One parse produces both the ESTree AST and a parser-driven token stream (oxc's
//! `TokensParserConfig`); `token_map` translates those tokens to the espree/ts-estree
//! ESLint token contract. There is no hand-written lexer -- token disambiguation
//! (`/` regex-vs-division, templates, JSX, TS `<`) comes from real parser state.

mod parse;
mod token_map;

use napi_derive::napi;

pub use parse::{CommentObj, ParseResult};

/// Reject sources whose serialized ESTree JSON would exceed V8's ~512MB single-string
/// cap (the JSON is ~9-26x the source size). This is the JSON-transfer ceiling
/// (raw transfer is a future optimization). Surfacing a clear parseError here beats
/// the cryptic "Failed to convert rust String into napi string" that napi would throw.
const MAX_SOURCE_BYTES: usize = 16 * 1024 * 1024;

/// Parse JS/TS/JSX source -> ESTree JSON + ESLint-shape comments (UTF-16 offsets).
/// Replaces npm `oxc-parser`'s `parseSync`.
///
/// - `filename`: used for lang inference (extension).
/// - `source_type`: `"module"` | `"script"` | `"commonjs"` (commonjs is treated as script).
/// - `jsx`: `languageOptions.parserOptions.ecmaFeatures.jsx`; true promotes `.ts->tsx`/`.js->jsx`.
///
/// Returns `Err` only for the size guard above; the JS side maps that to a `parseError`
/// (matching the current "parseSync throw -> parseError" contract). `catch_unwind` turns a
/// Rust panic into a JS exception (the worker survives). Note: a stack overflow (deep
/// nesting) is a SIGSEGV that catch_unwind does NOT catch.
#[napi(catch_unwind)]
pub fn parse(
    filename: String,
    source: String,
    source_type: String,
    jsx: bool,
) -> napi::Result<ParseResult> {
    if source.len() > MAX_SOURCE_BYTES {
        return Err(napi::Error::from_reason(format!(
            "source too large ({} bytes > {}-byte JSON-transfer limit)",
            source.len(),
            MAX_SOURCE_BYTES
        )));
    }
    Ok(parse::parse_estree(&filename, &source, &source_type, jsx))
}
