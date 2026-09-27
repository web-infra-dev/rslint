//! `@rslint/native`: oxc 0.133 parse exposed via napi to `@rslint/core`'s
//! ESLint-plugin runtime, replacing npm `oxc-parser` and the JS tokenizer.
//!
//! One parse produces both the ESTree AST and a parser-driven token stream (oxc's
//! `TokensParserConfig`); `token_map` translates those tokens to the espree/ts-estree
//! ESLint token contract. There is no hand-written lexer -- token disambiguation
//! (`/` regex-vs-division, templates, JSX, TS `<`) comes from real parser state.

mod parse;
mod source_transport;
mod token_map;

use napi::{Env, JsString};
use napi_derive::napi;

pub use parse::{CommentObj, ParseResult};
use source_transport::SharedSource;

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
    check_source_size(source.len())?;
    Ok(parse::parse_estree(&filename, &source, &source_type, jsx))
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
#[napi(catch_unwind)]
pub fn parse_shared_source(
    env: &Env,
    filename: String,
    source: SharedSource,
    source_type: String,
    jsx: bool,
) -> napi::Result<SourceParseResult<'_>> {
    source_transport::with_bytes(source, |bytes| {
        let text = std::str::from_utf8(bytes)
            .map_err(|_| napi::Error::from_reason("invalid shared source UTF-8"))?;
        let had_bom = text.starts_with('\u{feff}');
        let text = text.strip_prefix('\u{feff}').unwrap_or(text);
        check_source_size(text.len())?;
        let parsed = parse::parse_estree(&filename, text, &source_type, jsx);
        // N-API copies borrowed UTF-8 directly into the required JS SourceCode
        // string, with no temporary Rust String or JS -> Rust round trip.
        let source_text = env.create_string(text)?;
        Ok(SourceParseResult {
            parsed,
            source_text,
            had_bom,
        })
    })
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
