//! oxc 0.133 parse -> ESTree JSON + ESLint-shape comments, offsets in UTF-16 code units.
//! lang derivation lives here (Rust-side) as the single source of truth; the JS side
//! only forwards the `jsx` flag and `sourceType`.

use napi::bindgen_prelude::{Uint32Array, Uint8Array};
use napi_derive::napi;
use oxc_allocator::Allocator;
use oxc_ast_visit::utf8_to_utf16::Utf8ToUtf16;
use oxc_parser::config::TokensParserConfig;
use oxc_parser::{ParseOptions, Parser};
use oxc_span::SourceType;
use serde_json::Value;

use crate::token_map;

/// ESLint-shape comment (`{ type, value, start, end }`). start/end are UTF-16 offsets.
#[napi(object)]
pub struct CommentObj {
    /// "Line" | "Block"
    pub r#type: String,
    /// Comment body with the delimiters (`//` or `/* */`) stripped.
    pub value: String,
    pub start: u32,
    pub end: u32,
}

#[napi(object)]
pub struct ParseResult {
    /// ESTree JSON (UTF-16 offsets, no `range` -- normalize-ast derives it from start/end).
    /// The JS side runs `JSON.parse`.
    pub program: String,
    pub comments: Vec<CommentObj>,
    /// Parser-driven token stream in columnar form, all UTF-16 offsets.
    /// TypedArrays (not `Vec`/JS `number[]`) so the whole column transfers as one
    /// ArrayBuffer instead of N per-element `napi_set_element` crossings:
    /// - `token_types[i]`: ESLint token type as `token_map::TokenType as u8`
    /// - `token_starts[i]` / `token_ends[i]`: range
    /// `value`/`loc` are recomputed JS-side (source slice + lazy loc), so they are not sent.
    /// Empty only on a fatal parse abort (oxc sets `panicked`); recoverable
    /// syntax errors still produce a full token stream alongside the recovery AST.
    pub token_types: Uint8Array,
    pub token_starts: Uint32Array,
    pub token_ends: Uint32Array,
}

/// Parse `source` into ESTree.
///
/// Recovering semantics: this does NOT read `panicked`/`errors`; it always returns the
/// program (including oxc's recovery AST), matching the current "only a thrown parseSync
/// sets parseError" behavior. A real failure only surfaces when the napi layer catches an
/// actual panic (catch_unwind) or rejects the input up front (see lib.rs size guard).
pub fn parse_estree(filename: &str, source: &str, source_type_raw: &str, jsx: bool) -> ParseResult {
    let source_type = derive_source_type(filename, jsx, source_type_raw);
    let allocator = Allocator::default();
    let mut ret = Parser::new(&allocator, source, source_type)
        .with_options(ParseOptions {
            preserve_parens: false,
            ..ParseOptions::default()
        })
        // Collect the parser-driven token stream in the same parse: `/` regex-vs-
        // division, template, JSX and TS `<` are disambiguated by real parser state.
        .with_config(TokensParserConfig)
        .parse();

    // Extract each comment's value now, while still in the UTF-8 phase: value is the
    // content_span slice (delimiters stripped). This MUST happen before the conversion
    // below -- afterwards spans are UTF-16 and can no longer index the UTF-8 source.
    // `convert_comments` only rewrites spans (never length/order/kind), so zipping the
    // values back after conversion stays aligned, and `kind` is re-read post-conversion.
    let values: Vec<String> = ret
        .program
        .comments
        .iter()
        .map(|c| {
            let cs = c.content_span();
            debug_assert!(
                source.is_char_boundary(cs.start as usize)
                    && source.is_char_boundary(cs.end as usize),
                "comment content_span must lie on char boundaries"
            );
            source
                .get(cs.start as usize..cs.end as usize)
                .unwrap_or_default()
                .to_string()
        })
        .collect();

    // UTF-8 byte spans -> UTF-16 code units (program + comments). Same source and same
    // converter algorithm as the token side, so offsets are bit-identical across both.
    let conv = Utf8ToUtf16::new(source);
    conv.convert_program(&mut ret.program);
    conv.convert_comments(&mut ret.program.comments);

    let comments: Vec<CommentObj> = ret
        .program
        .comments
        .iter()
        .zip(values)
        .map(|(c, value)| CommentObj {
            r#type: if c.is_line() { "Line" } else { "Block" }.to_string(),
            value,
            start: c.span.start,
            end: c.span.end,
        })
        .collect();

    // ranges:false -- emit only start/end (matches npm oxc-parser's default; normalize-ast
    // lazily derives `range` from start/end, gated on `range == null`). Avoids ~+20% JSON.
    let is_ts = source_type.is_typescript();
    let program = if is_ts {
        ret.program.to_estree_ts_json(false)
    } else {
        ret.program.to_estree_js_json(false)
    };

    // Token columnar. The Kind->ESLint-type mapping needs an AST "cover"
    // (which spans are Identifier/JSX leaves). The cover is built from the ESTree JSON the
    // program serialized to above (so spans already match). NOTE(perf): this
    // re-parses that JSON into a `serde_json::Value` once per parse. The reviewed alternative
    // -- walk the oxc AST directly via `oxc_ast_visit::Visit` -- avoids the round-trip and the
    // `serde_json` runtime dep, but is deferred: the ESTree value gives ONE uniform
    // `Identifier`/`JSXIdentifier` node shape, whereas the native AST spreads the same token
    // type across IdentifierName/Reference/Binding/Label + TSThisParameter + the as-const
    // TSTypeReference + typeof-this ThisExpression, each needing its own visit hook and span
    // handling (oxc folds TS type annotations into the ESTree Identifier span, which the
    // `<= end` prefix match relies on). That rewrite warrants its own differential re-run.
    let ast: Value = serde_json::from_str(&program).unwrap_or(Value::Null);
    let cols = token_map::build_columns(&ret.tokens, &ast, is_ts, &conv);

    ParseResult {
        program,
        comments,
        token_types: Uint8Array::new(cols.types),
        token_starts: Uint32Array::new(cols.starts),
        token_ends: Uint32Array::new(cols.ends),
    }
}

/// Extension + jsx-flag lang derivation (this crate is the single source of truth):
/// - lang inferred from the file extension by default (oxc `from_path`).
/// - when `ecmaFeatures.jsx === true`, promote `.ts/.mts/.cts -> tsx` and
///   `.js/.mjs/.cjs -> jsx` (`.jsx`/`.tsx` already imply their own lang). `.d.ts`
///   ends with `.ts`, so it is promoted to `tsx` as well -- bit-for-bit with the
///   pre-migration JS regex `/\.[mc]?ts$/i`. Harmless: ambient declarations parse
///   identically under tsx (JSX ambiguity only hits expression-position arrows,
///   which declaration files don't contain).
/// - sourceType: `commonjs -> script` (mirrors the JS side), `module`/`script` pass through.
/// - preserveParens: false (set in ParseOptions; aligns the AST with v10/espree).
fn derive_source_type(filename: &str, jsx: bool, source_type_raw: &str) -> SourceType {
    // from_path is case-sensitive, mirroring npm oxc-parser's extension inference exactly
    // (so `.TS` etc. infer the same lang they do today). The jsx-promotion branch below is
    // case-insensitive, mirroring the JS side's `/\.[mc]?ts$/i`. This asymmetry is
    // intentional -- it reproduces the current ecma-language-plugin behavior bit-for-bit.
    let mut st = SourceType::from_path(filename).unwrap_or_else(|_| SourceType::mjs());
    if jsx {
        let lower = filename.to_ascii_lowercase();
        if lower.ends_with(".ts") || lower.ends_with(".mts") || lower.ends_with(".cts") {
            st = SourceType::tsx();
        } else if lower.ends_with(".js") || lower.ends_with(".mjs") || lower.ends_with(".cjs") {
            st = SourceType::jsx();
        }
        // .jsx / .tsx leave as-is. `.d.ts` is deliberately NOT special-cased: it ends
        // with `.ts`, so the branch above already promoted it to `tsx` (see doc above).
    }
    match source_type_raw {
        "script" | "commonjs" => st.with_script(true),
        _ => st.with_module(true), // "module" (default)
    }
}
