//! Maps oxc's parser-driven token stream (`ParserReturn.tokens`, enabled via
//! `TokensParserConfig`) to the espree / ts-estree ESLint token contract.
//!
//! The token stream is produced by the real parse, so `/` regex-vs-division, template
//! continuation, JSX and TS `<` disambiguation are already correct -- this module only
//! has to translate each token's `Kind` to an ESLint token *type*. The rules below are
//! empirically pinned against espree@11 (.js/.jsx) and @typescript-eslint/typescript-estree
//! @8.59 (.ts). For `.tsx`, JSX tokens follow espree-style classification (the cover is
//! oxc's single AST, which labels JSX names the espree way) -- this diverges from ts-estree
//! on member-expressions-in-JSX (`{i.name}`: Identifier, not JSXIdentifier) and namespaced
//! names (`<a:b/>`: JSXIdentifier, not Identifier).
//!
//! Both oracles classify by AST role: a word sitting in an Identifier-leaf position
//! (member / property key / binding / type reference) is an `Identifier` token even when
//! it is a reserved word. That AST-dependent part is resolved by `Cover` (a span->type map
//! built from the ESTree AST). Pure-lexical cases (literals, punctuators, most keywords)
//! map straight from `Kind` + its classification flags.

use std::collections::HashMap;

use oxc_ast_visit::utf8_to_utf16::Utf8ToUtf16;
use oxc_parser::{Kind, Token};
use serde_json::Value;

/// ESLint token types (espree / ts-estree contract). `#[repr(u8)]` so it serializes
/// straight into the columnar `types` array.
#[derive(Clone, Copy, PartialEq, Eq, Debug)]
#[repr(u8)]
pub enum TokenType {
    Identifier = 0,
    Keyword = 1,
    Punctuator = 2,
    String = 3,
    Numeric = 4,
    RegularExpression = 5,
    Template = 6,
    Boolean = 7,
    Null = 8,
    PrivateIdentifier = 9,
    JsxIdentifier = 10,
    JsxText = 11,
    /// JSX attribute string value (`b="x"`). espree tokenizes it as `JSXText` too, but its
    /// value is the RAW slice (incl. quotes, no entity decode) -- unlike JSX text, which is
    /// entity-decoded + CRLF-folded. A distinct code lets the JS rebuild pick the right value
    /// path; `as_str` is still "JSXText" so the ESLint token type is unchanged.
    JsxTextAttr = 12,
}

impl TokenType {
    /// Inverse of `self as u8` -- recover the type from its columnar code. Test-only: the
    /// production token rebuild (and the ESLint type-name strings) live JS-side in
    /// token-builder.ts; this exists only for the Rust unit tests that assert mapped types.
    #[cfg(test)]
    pub fn from_u8(code: u8) -> Option<Self> {
        Some(match code {
            0 => TokenType::Identifier,
            1 => TokenType::Keyword,
            2 => TokenType::Punctuator,
            3 => TokenType::String,
            4 => TokenType::Numeric,
            5 => TokenType::RegularExpression,
            6 => TokenType::Template,
            7 => TokenType::Boolean,
            8 => TokenType::Null,
            9 => TokenType::PrivateIdentifier,
            10 => TokenType::JsxIdentifier,
            11 => TokenType::JsxText,
            12 => TokenType::JsxTextAttr,
            _ => return None,
        })
    }
}

/// AST-derived classification cover: token start (UTF-16) -> (node end, forced type).
/// Resolves the cases bare `Kind` can't express -- a reserved word used as a name, JSX
/// identifiers/text (oxc lexes these as plain `Ident`/`Str`), and `typeof this`.
pub type Cover = HashMap<u32, (u32, TokenType)>;

fn node_u32(map: &serde_json::Map<String, Value>, key: &str) -> Option<u32> {
    map.get(key).and_then(Value::as_u64).map(|n| n as u32)
}

/// Walk the ESTree AST collecting the span->type cover. `in_type_query` tracks whether we
/// are inside a `typeof` type query (where ts-estree tokenizes `this` as an Identifier).
fn walk_cover(node: &Value, in_type_query: bool, cover: &mut Cover) {
    match node {
        Value::Array(arr) => {
            for child in arr {
                walk_cover(child, in_type_query, cover);
            }
        }
        Value::Object(map) => {
            let ty = map.get("type").and_then(Value::as_str);
            if let (Some(t), Some(s), Some(e)) = (ty, node_u32(map, "start"), node_u32(map, "end"))
            {
                match t {
                    "JSXIdentifier" => {
                        cover.insert(s, (e, TokenType::JsxIdentifier));
                    }
                    "JSXText" => {
                        cover.insert(s, (e, TokenType::JsxText));
                    }
                    "Identifier" => {
                        cover.insert(s, (e, TokenType::Identifier));
                    }
                    // ts-estree tokenizes `this` inside a `typeof` type query as an Identifier.
                    "ThisExpression" if in_type_query => {
                        cover.insert(s, (e, TokenType::Identifier));
                    }
                    // espree quirk: a JSX attribute's string value tokenizes as JSXText.
                    "JSXAttribute" => {
                        if let Some(val) = map.get("value") {
                            if val.get("type").and_then(Value::as_str) == Some("Literal")
                                && val.get("value").map(Value::is_string).unwrap_or(false)
                            {
                                if let (Some(vs), Some(ve)) = (
                                    val.get("start").and_then(Value::as_u64),
                                    val.get("end").and_then(Value::as_u64),
                                ) {
                                    cover.insert(vs as u32, (ve as u32, TokenType::JsxTextAttr));
                                }
                            }
                        }
                    }
                    // MetaProperty's `meta` (new/import) stays a Keyword token even though it is
                    // an Identifier node -- don't cover it. The `property` (target/meta) is an
                    // Identifier token, so walk only that.
                    "MetaProperty" => {
                        if let Some(prop) = map.get("property") {
                            walk_cover(prop, in_type_query, cover);
                        }
                        return;
                    }
                    _ => {}
                }
            }
            let child_tq = in_type_query || ty == Some("TSTypeQuery");
            for (k, v) in map {
                if matches!(k.as_str(), "type" | "start" | "end" | "range" | "loc") {
                    continue;
                }
                walk_cover(v, child_tq, cover);
            }
        }
        _ => {}
    }
}

/// Build the span->type cover from the (UTF-16-offset) ESTree AST JSON value.
pub fn collect_cover(ast: &Value) -> Cover {
    let mut cover = Cover::new();
    walk_cover(ast, false, &mut cover);
    cover
}

/// Translate one oxc token to its ESLint token type. `start`/`end` are UTF-16 offsets
/// (so they line up with `cover`). `is_ts` selects the .ts/.tsx vs .js/.jsx rules.
pub fn map_token(kind: Kind, start: u32, end: u32, cover: &Cover, is_ts: bool) -> TokenType {
    use TokenType as T;

    // AST-assisted: a token whose span starts a named leaf node and lies within it takes
    // that node's type. `<= end` (not `==`) because oxc folds a binding's TS type annotation
    // into the Identifier node's span (`this: Window` -> Identifier[..]) while the token is
    // just the name; the name is always the node's prefix.
    if let Some(&(cover_end, cover_type)) = cover.get(&start) {
        if end <= cover_end {
            // espree exception: static/let/yield stay Keyword even in an Identifier-node
            // position (member/property name). ts-estree makes no such exception.
            let js_kw_exception = !is_ts
                && cover_type == T::Identifier
                && matches!(kind, Kind::Static | Kind::Let | Kind::Yield);
            if !js_kw_exception {
                return cover_type;
            }
        }
    }

    if kind.is_number() {
        return T::Numeric;
    }
    match kind {
        Kind::Str => T::String,
        Kind::RegExp => T::RegularExpression,
        Kind::NoSubstitutionTemplate
        | Kind::TemplateHead
        | Kind::TemplateMiddle
        | Kind::TemplateTail => T::Template,
        Kind::True | Kind::False => T::Boolean,
        Kind::Null => T::Null,
        Kind::PrivateIdentifier => T::PrivateIdentifier,
        Kind::Ident => T::Identifier,
        Kind::JSXText => T::JsxText,
        // espree's token layer treats `await` as a name (it is not in acorn's keyword set).
        Kind::Await => T::Identifier,
        // `enum` is a future-reserved word: Identifier in .js, Keyword in .ts.
        Kind::Enum => {
            if is_ts {
                T::Keyword
            } else {
                T::Identifier
            }
        }
        _ => {
            if kind.is_contextual_keyword() {
                // async/of/as/from/get/set/type/number/... -- a name token in both oracles.
                T::Identifier
            } else if kind.is_future_reserved_keyword() && kind != Kind::Static {
                // interface/private/public/... -- Identifier in .js, Keyword in .ts.
                if is_ts {
                    T::Keyword
                } else {
                    T::Identifier
                }
            } else if kind.is_any_keyword() {
                // true reserved words + let/yield/static.
                T::Keyword
            } else {
                T::Punctuator
            }
        }
    }
}

/// Columnar token output: parallel arrays, UTF-16 offsets. `value`/`loc` are
/// recomputed JS-side (slice + lazy loc), so they are not carried here.
pub struct TokenColumns {
    pub types: Vec<u8>,
    pub starts: Vec<u32>,
    pub ends: Vec<u32>,
}

/// Map the whole token stream to columnar ESLint tokens. `ast` is the UTF-16-offset ESTree
/// JSON value (same source/converter as the program, so offsets are bit-identical).
pub fn build_columns(
    tokens: &[Token],
    ast: &Value,
    is_ts: bool,
    conv: &Utf8ToUtf16,
) -> TokenColumns {
    let cover = collect_cover(ast);
    let mut converter = conv.converter();
    let n = tokens.len();
    let mut types = Vec::with_capacity(n);
    let mut starts = Vec::with_capacity(n);
    let mut ends = Vec::with_capacity(n);
    for token in tokens {
        let (mut start, mut end) = (token.start(), token.end());
        if let Some(c) = converter.as_mut() {
            // Tokens are in source order, so offsets ascend -> fast path.
            c.convert_offset(&mut start);
            c.convert_offset(&mut end);
        }
        types.push(map_token(token.kind(), start, end, &cover, is_ts) as u8);
        starts.push(start);
        ends.push(end);
    }
    TokenColumns {
        types,
        starts,
        ends,
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use oxc_allocator::Allocator;
    use oxc_parser::{config::TokensParserConfig, Parser};
    use oxc_span::SourceType;

    /// Parse with tokens, then return `(value, eslintType)` pairs for every token -- the
    /// shape the differential harness pins against espree/ts-estree.
    fn lex(source: &str, st: SourceType) -> Vec<(String, TokenType)> {
        let alloc = Allocator::default();
        let mut ret = Parser::new(&alloc, source, st)
            .with_config(TokensParserConfig)
            .parse();
        let conv = Utf8ToUtf16::new(source);
        conv.convert_program(&mut ret.program);
        let ast_json = if st.is_typescript() {
            ret.program.to_estree_ts_json(false)
        } else {
            ret.program.to_estree_js_json(false)
        };
        let ast: Value = serde_json::from_str(&ast_json).unwrap();
        let cols = build_columns(&ret.tokens, &ast, st.is_typescript(), &conv);
        // Recover value by slicing the source on the (UTF-16) ranges -- ASCII test inputs,
        // so byte == UTF-16 here; fine for asserting types.
        cols.types
            .iter()
            .zip(&cols.starts)
            .zip(&cols.ends)
            .map(|((&ty, &s), &e)| {
                // ASCII test inputs, so UTF-16 offsets index the source like bytes.
                let v = source.get(s as usize..e as usize).unwrap_or("").to_string();
                (v, TokenType::from_u8(ty).expect("valid type code"))
            })
            .collect()
    }

    fn ty_of(toks: &[(String, TokenType)], value: &str) -> Vec<TokenType> {
        toks.iter()
            .filter(|(v, _)| v == value)
            .map(|(_, t)| *t)
            .collect()
    }

    #[test]
    fn regex_vs_division_is_parser_driven() {
        use TokenType::*;
        // regex after control-paren / block / fn-decl body / return; division after call /
        // object / postfix -- all decided by the parser, no heuristics.
        for (src, want_regex) in [
            ("if (x) /re/.test(y)", true),
            ("function w(){ return /re/g }", true),
            ("function f(){} /re/", true),
            ("f(x) / 2", false),
            ("const k = ({a:1}) / 2", false),
            ("const x = i++ / 2", false),
        ] {
            let toks = lex(src, SourceType::mjs());
            let has_regex = toks.iter().any(|(_, t)| *t == RegularExpression);
            assert_eq!(has_regex, want_regex, "regex detection wrong for {src:?}");
        }
    }

    #[test]
    fn literals_and_templates() {
        use TokenType::*;
        let toks = lex(
            "const a = `x${b}y`, c = true, d = null, e = 0xFF, f = /re/g",
            SourceType::mjs(),
        );
        assert_eq!(ty_of(&toks, "`x${"), vec![Template]);
        assert_eq!(ty_of(&toks, "}y`"), vec![Template]);
        assert_eq!(ty_of(&toks, "true"), vec![Boolean]);
        assert_eq!(ty_of(&toks, "null"), vec![Null]);
        assert_eq!(ty_of(&toks, "0xFF"), vec![Numeric]);
        assert_eq!(ty_of(&toks, "/re/g"), vec![RegularExpression]);
    }

    #[test]
    fn keyword_classification_js() {
        use TokenType::*;
        // future-reserved + contextual as property/member names -> Identifier; static/let/
        // yield stay Keyword; real reserved (if) stays Keyword in statement position.
        let toks = lex(
            "if (x) {} const o = { enum: 1, async: 2 }; o.private; o.static",
            SourceType::cjs(),
        );
        assert_eq!(ty_of(&toks, "if"), vec![Keyword]);
        assert_eq!(ty_of(&toks, "enum"), vec![Identifier]);
        assert_eq!(ty_of(&toks, "async"), vec![Identifier]);
        assert_eq!(ty_of(&toks, "private"), vec![Identifier]);
        assert_eq!(ty_of(&toks, "static"), vec![Keyword]);
    }

    #[test]
    fn keyword_classification_ts() {
        use TokenType::*;
        // ts-estree: future-reserved in declaration -> Keyword; type/as contextual -> Identifier.
        let toks = lex(
            "enum E {} interface I {} type T = X; const y = z as const",
            SourceType::ts(),
        );
        assert_eq!(ty_of(&toks, "enum"), vec![Keyword]);
        assert_eq!(ty_of(&toks, "interface"), vec![Keyword]);
        assert_eq!(ty_of(&toks, "type"), vec![Identifier]);
        assert_eq!(ty_of(&toks, "as"), vec![Identifier]);
        // `as const`: the second `const` is an Identifier node (TSTypeReference) -> Identifier.
        assert_eq!(ty_of(&toks, "const"), vec![Keyword, Identifier]);
    }

    #[test]
    fn jsx_names_and_text() {
        use TokenType::*;
        let toks = lex(
            r#"const e = <div className="x">hi {y}</div>"#,
            SourceType::jsx(),
        );
        assert_eq!(ty_of(&toks, "div"), vec![JsxIdentifier, JsxIdentifier]);
        assert_eq!(ty_of(&toks, "className"), vec![JsxIdentifier]);
        assert_eq!(ty_of(&toks, "\"x\""), vec![JsxTextAttr]); // attr string: distinct code, "JSXText" type
        assert_eq!(ty_of(&toks, "hi "), vec![JsxText]);
        assert_eq!(ty_of(&toks, "y"), vec![Identifier]); // inside expression container
    }

    #[test]
    fn this_and_meta() {
        use TokenType::*;
        // ThisExpression -> Keyword; MetaProperty.meta (new) -> Keyword, property -> Identifier.
        let toks = lex(
            "class C { m() { return this.x } } function g(){ return new.target }",
            SourceType::mjs(),
        );
        assert_eq!(ty_of(&toks, "this"), vec![Keyword]);
        assert_eq!(ty_of(&toks, "new"), vec![Keyword]);
        assert_eq!(ty_of(&toks, "target"), vec![Identifier]);
        // ts `this` parameter is an Identifier node -> Identifier token.
        let toks = lex("function f(this: Window) {}", SourceType::ts());
        assert_eq!(ty_of(&toks, "this"), vec![Identifier]);
    }

    #[test]
    fn typeof_this_is_identifier() {
        use TokenType::*;
        // ts-estree: `this` inside a `typeof` type query is an Identifier; a value `this` is a
        // Keyword. Both appear here, so this also proves `in_type_query` is scoped to the query.
        let toks = lex(
            "class C { ch = 1; f = (a: typeof this.ch) => this.ch }",
            SourceType::ts(),
        );
        assert_eq!(ty_of(&toks, "this"), vec![Identifier, Keyword]);
    }

    #[test]
    fn let_yield_stay_keyword() {
        use TokenType::*;
        // espree keeps let/yield as Keyword even as member names (the static/let/yield exception),
        // and in declaration / generator position.
        let toks = lex("o.let; o.yield; let a = 1", SourceType::cjs());
        assert_eq!(ty_of(&toks, "let"), vec![Keyword, Keyword]); // member name + declaration
        assert_eq!(ty_of(&toks, "yield"), vec![Keyword]);
        let toks = lex("function* g(){ yield 1 }", SourceType::mjs());
        assert_eq!(ty_of(&toks, "yield"), vec![Keyword]);
    }

    #[test]
    fn await_is_identifier() {
        use TokenType::*;
        // espree's token layer treats `await` as a name, even inside async.
        let toks = lex("async function f(){ await x }", SourceType::mjs());
        assert_eq!(ty_of(&toks, "await"), vec![Identifier]);
    }

    #[test]
    fn string_and_private_identifier() {
        use TokenType::*;
        let toks = lex(
            "class C { #x = 1; m(){ return this.#x + \"s\" } }",
            SourceType::mjs(),
        );
        assert_eq!(ty_of(&toks, "\"s\""), vec![String]);
        assert_eq!(
            ty_of(&toks, "#x"),
            vec![PrivateIdentifier, PrivateIdentifier]
        );
    }

    #[test]
    fn template_grain() {
        use TokenType::*;
        // NoSubstitutionTemplate + Head/Middle/Tail all -> Template; grain matches espree.
        let toks = lex("const a = `plain`, b = `x${m}y${n}z`", SourceType::mjs());
        assert_eq!(ty_of(&toks, "`plain`"), vec![Template]); // NoSubstitutionTemplate
        assert_eq!(ty_of(&toks, "`x${"), vec![Template]); // TemplateHead
        assert_eq!(ty_of(&toks, "}y${"), vec![Template]); // TemplateMiddle
        assert_eq!(ty_of(&toks, "}z`"), vec![Template]); // TemplateTail
    }

    #[test]
    fn utf16_offsets_not_bytes() {
        // Offsets are UTF-16 code units, not bytes. `中` is 3 UTF-8 bytes but 1 UTF-16 unit;
        // `🚀` is 4 bytes but 2 units. Assert on the columns directly (the ASCII-only `lex`
        // value-recovery helper can't be used for non-ASCII).
        let alloc = Allocator::default();
        let source = "const 中 = '🚀'";
        let mut ret = Parser::new(&alloc, source, SourceType::mjs())
            .with_config(TokensParserConfig)
            .parse();
        let conv = Utf8ToUtf16::new(source);
        conv.convert_program(&mut ret.program);
        let ast: Value = serde_json::from_str(&ret.program.to_estree_js_json(false)).unwrap();
        let cols = build_columns(&ret.tokens, &ast, false, &conv);
        // tokens: const, 中 (Identifier), =, '🚀' (String)
        assert_eq!(cols.types.len(), 4);
        assert_eq!(
            TokenType::from_u8(cols.types[1]),
            Some(TokenType::Identifier)
        );
        assert_eq!((cols.starts[1], cols.ends[1]), (6, 7)); // 中: UTF-16 [6,7), not byte [6,9)
        assert_eq!(TokenType::from_u8(cols.types[3]), Some(TokenType::String));
        assert_eq!((cols.starts[3], cols.ends[3]), (10, 14)); // '🚀': UTF-16 [10,14), not byte [12,18)
    }
}
