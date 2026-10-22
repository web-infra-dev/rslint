//! Default TypeScript policy rules enabled by the frozen Rsbuild workload.
//! Comment matching operates on compiler-owned trivia, never string contents.

use crate::{Violation, comments::for_each_comment, edits::Edit};
use tsrs_ast::{
    CommentRange, Kind, Node, OEK, SourceFile, SymbolFlags, get_source_file_of_node,
    is_assignment_operator, is_external_module_augmentation,
    is_function_like_or_class_static_block_declaration, is_global_scope_augmentation,
    is_prologue_directive, skip_outer_expressions,
};
use tsrs_checker::Checker;
use tsrs_core::P;
use tsrs_scanner::{get_leading_comment_ranges, get_token_pos_of_node};
use unicode_segmentation_v15::UnicodeSegmentation;

pub(crate) const NAMES: &[&str] = &[
    "@typescript-eslint/ban-ts-comment",
    "@typescript-eslint/no-require-imports",
    "@typescript-eslint/no-unused-expressions",
    "@typescript-eslint/triple-slash-reference",
];

pub(crate) fn check(
    name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
) -> Vec<Violation> {
    match name {
        "@typescript-eslint/ban-ts-comment" if node.kind() == Kind::SourceFile => {
            ban_comments(node, file)
        }
        "@typescript-eslint/triple-slash-reference" if node.kind() == Kind::SourceFile => {
            triple_slash_references(node, file)
        }
        "@typescript-eslint/no-require-imports" => {
            require_import(node, file, checker).into_iter().collect()
        }
        "@typescript-eslint/no-unused-expressions" => unused_expression(node).into_iter().collect(),
        _ => Vec::new(),
    }
}

// ECMAScript TrimString and regexp whitespace, including BOM but not U+0085.
fn js_whitespace(character: char) -> bool {
    matches!(character, '\u{0009}'..='\u{000d}' | ' ' | '\u{00a0}' | '\u{1680}'
        | '\u{2000}'..='\u{200a}' | '\u{2028}' | '\u{2029}' | '\u{202f}'
        | '\u{205f}' | '\u{3000}' | '\u{feff}')
}

fn skip_directive_whitespace(text: &str) -> &str {
    text.trim_start_matches([' ', '\t', '\n', '\x0c', '\r'])
}

#[derive(Clone, Copy, PartialEq, Eq)]
enum Directive {
    ExpectError,
    Ignore,
    Nocheck,
    Check,
}

fn match_directive(text: &str) -> Option<(Directive, &str)> {
    let name = text.strip_prefix("@ts-")?;
    for (spelling, kind) in [
        ("expect-error", Directive::ExpectError),
        ("ignore", Directive::Ignore),
        ("nocheck", Directive::Nocheck),
        ("check", Directive::Check),
    ] {
        if let Some(rest) = name.strip_prefix(spelling)
            && !rest
                .as_bytes()
                .first()
                .is_some_and(|byte| byte.is_ascii_alphanumeric() || *byte == b'_')
        {
            return Some((kind, rest));
        }
    }
    None
}

fn directive_comment(comment: CommentRange, text: &str) -> Option<(Directive, &str)> {
    if comment.kind == Kind::SingleLineCommentTrivia {
        let rest = text.trim_start_matches('/');
        let slashes = text.len() - rest.len();
        if slashes < 2 {
            return None;
        }
        let (directive, description) = match_directive(skip_directive_whitespace(rest))?;
        if matches!(directive, Directive::Check | Directive::Nocheck) && slashes > 3 {
            return None;
        }
        return Some((directive, description));
    }
    if comment.kind != Kind::MultiLineCommentTrivia || text.len() < 4 {
        return None;
    }
    let content = &text[2..text.len() - 2];
    let last_line = content.rsplit('\n').next()?;
    let rest = skip_directive_whitespace(last_line).trim_start_matches(['/', '*']);
    let (directive, description) = match_directive(skip_directive_whitespace(rest))?;
    matches!(directive, Directive::ExpectError | Directive::Ignore)
        .then_some((directive, description))
}

fn ban_comments(node: P<Node>, file: P<SourceFile>) -> Vec<Violation> {
    let mut reports = Vec::new();
    if !file.text().contains("@ts-") {
        return reports;
    }
    let first_statement = file
        .statements()
        .nodes()
        .first()
        .map(|&statement| get_token_pos_of_node(statement, file, false));
    for_each_comment(file, |comment| {
        let start = comment.pos() as usize;
        let end = comment.end() as usize;
        let text = &file.text()[start..end];
        let Some((directive, description)) = directive_comment(comment, text) else { return Ok(()); };
        let report = match directive {
            Directive::Check => return Ok(()),
            Directive::Nocheck => {
                if first_statement.is_some_and(|first| comment.pos() >= first) {
                    return Ok(());
                }
                Violation::new(node, "tsDirectiveComment", "Do not use '@ts-nocheck' because it alters compilation errors.")
            }
            Directive::Ignore => {
                // The original rule replaces the final occurrence, even if a
                // directive-looking marker is repeated in its description.
                let marker = text.rfind("@ts-ignore").expect("matched ignore directive");
                Violation::new(node, "tsIgnoreInsteadOfExpectError", "Prefer '@ts-expect-error' over '@ts-ignore' as it requires error to be present in next line.").with_suggestion(
                    "replaceTsIgnoreWithTsExpectError", "Replace '@ts-ignore' with '@ts-expect-error'.",
                    vec![Edit::replace([start + marker, start + marker + "@ts-ignore".len()], "@ts-expect-error")],
                )
            }
            Directive::ExpectError => {
                // The frozen Go oracle uses uniseg 0.4.7 / Unicode 15.0;
                // pin the corresponding Unicode segmentation tables, including
                // combining sequences, flag pairs and emoji ZWJ clusters.
                if description.trim_matches(js_whitespace).graphemes(true).take(3).count() >= 3 {
                    return Ok(());
                }
                Violation::new(node, "tsDirectiveCommentRequiresDescription", "Include a description after the '@ts-expect-error' directive to explain why the '@ts-expect-error' is necessary. The description must be 3 characters or longer.")
            }
        };
        reports.push(report.with_byte_range([start, end]));
        Ok(())
    }).expect("comment traversal callback is infallible");
    reports
}

fn in_augmentation(mut node: P<Node>) -> bool {
    while let Some(parent) = node.parent() {
        if parent.kind() == Kind::ModuleDeclaration
            && (is_global_scope_augmentation(parent) || is_external_module_augmentation(parent))
        {
            return true;
        }
        node = parent;
    }
    false
}

fn require_import(node: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> Option<Violation> {
    match node.kind() {
        Kind::CallExpression => {
            let callee = node.as_call_expression().expression();
            // Frozen rule deliberately does not unwrap a parenthesized callee.
            if callee.kind() != Kind::Identifier || callee.text() != "require" {
                return None;
            }
            let symbol = checker.resolve_name_exported(
                "require",
                callee,
                SymbolFlags::Value | SymbolFlags::Alias,
                false,
            );
            if symbol.is_some_and(|symbol| {
                symbol.declarations().iter().any(|&declaration| {
                    !matches!(
                        declaration.kind(),
                        Kind::InterfaceDeclaration | Kind::TypeAliasDeclaration
                    ) && get_source_file_of_node(declaration) == Some(file)
                        && !in_augmentation(declaration)
                })
            }) {
                return None;
            }
        }
        Kind::ExternalModuleReference => {}
        _ => return None,
    }
    Some(Violation::new(
        node,
        "noRequireImports",
        "A `require()` style import is forbidden.",
    ))
}

fn unused_expression(node: P<Node>) -> Option<Violation> {
    if node.kind() != Kind::ExpressionStatement
        || !disallowed(node.as_expression_statement().expression())
        || directive_prologue(node)
    {
        return None;
    }
    Some(Violation::new(
        node,
        "unusedExpression",
        "Expected an assignment or function call and instead saw an expression.",
    ))
}

fn disallowed(node: P<Node>) -> bool {
    let node = skip_outer_expressions(
        node,
        OEK::Parentheses | OEK::TypeAssertions | OEK::NonNullAssertions,
    );
    match node.kind() {
        Kind::ArrayLiteralExpression
        | Kind::ArrowFunction
        | Kind::ClassExpression
        | Kind::FalseKeyword
        | Kind::FunctionExpression
        | Kind::Identifier
        | Kind::MetaProperty
        | Kind::NoSubstitutionTemplateLiteral
        | Kind::NullKeyword
        | Kind::NumericLiteral
        | Kind::BigIntLiteral
        | Kind::ObjectLiteralExpression
        | Kind::PropertyAccessExpression
        | Kind::ElementAccessExpression
        | Kind::RegularExpressionLiteral
        | Kind::StringLiteral
        | Kind::SuperKeyword
        | Kind::TemplateExpression
        | Kind::ThisKeyword
        | Kind::TrueKeyword
        | Kind::TypeOfExpression
        | Kind::TaggedTemplateExpression
        | Kind::ConditionalExpression => true,
        Kind::BinaryExpression => {
            !is_assignment_operator(node.as_binary_expression().operator_token().kind())
        }
        Kind::PrefixUnaryExpression => !matches!(
            node.as_prefix_unary_expression().operator(),
            Kind::PlusPlusToken | Kind::MinusMinusToken
        ),
        Kind::ExpressionWithTypeArguments => node.expression().is_some_and(disallowed),
        // SatisfiesExpression intentionally remains in the allowed/unknown
        // bucket, unlike assertions. JSX is allowed by the frozen defaults.
        _ => false,
    }
}

fn directive_prologue(node: P<Node>) -> bool {
    if !is_prologue_directive(node) {
        return false;
    }
    let Some(parent) = node.parent() else {
        return false;
    };
    let container = matches!(parent.kind(), Kind::SourceFile | Kind::ModuleBlock)
        || (parent.kind() == Kind::Block
            && parent
                .parent()
                .is_some_and(is_function_like_or_class_static_block_declaration));
    if !container {
        return false;
    }
    for &statement in parent.statements() {
        if statement == node {
            return true;
        }
        if !is_prologue_directive(statement) {
            return false;
        }
    }
    false
}

// Match the frozen /^\/\s*<reference\s*(types|path|lib)\s*=\s*["|'](.*)["|']/
// without changing its deliberately permissive quote or greedy capture rules.
fn reference(value: &str) -> Option<(&str, &str)> {
    let mut text = value
        .strip_prefix('/')?
        .trim_start_matches(js_whitespace)
        .strip_prefix("<reference")?
        .trim_start_matches(js_whitespace);
    let kind = ["types", "path", "lib"]
        .into_iter()
        .find(|kind| text.starts_with(kind))?;
    text = text[kind.len()..]
        .trim_start_matches(js_whitespace)
        .strip_prefix('=')?
        .trim_start_matches(js_whitespace);
    let quote = text.as_bytes().first()?;
    if !matches!(quote, b'\'' | b'"' | b'|') {
        return None;
    }
    let rest = &text[1..];
    let end = rest.rfind(['\'', '"', '|'])?;
    Some((kind, &rest[..end]))
}

fn imported_modules(node: P<Node>, imports: &mut Vec<String>) {
    let expression = match node.kind() {
        Kind::ImportDeclaration => Some(node.as_import_declaration().module_specifier()),
        Kind::ImportEqualsDeclaration => {
            let reference = node.as_import_equals_declaration().module_reference();
            (reference.kind() == Kind::ExternalModuleReference)
                .then(|| reference.as_external_module_reference().expression())
        }
        _ => None,
    };
    if let Some(expression) = expression
        && matches!(
            expression.kind(),
            Kind::StringLiteral | Kind::NoSubstitutionTemplateLiteral
        )
    {
        imports.push(expression.text().to_owned());
    }
    node.for_each_child(&mut |child| {
        imported_modules(child, imports);
        false
    });
}

fn triple_slash_references(node: P<Node>, file: P<SourceFile>) -> Vec<Violation> {
    let mut references = Vec::new();
    for comment in get_leading_comment_ranges(file.text(), 0) {
        if comment.kind != Kind::SingleLineCommentTrivia {
            continue;
        }
        let start = comment.pos() as usize;
        let end = comment.end() as usize;
        let Some((kind, module)) = reference(&file.text()[start + 2..end]) else {
            continue;
        };
        if kind != "lib" {
            references.push((kind, module, start, end));
        }
    }
    let mut imports = Vec::new();
    if references.iter().any(|(kind, _, _, _)| *kind == "types") {
        imported_modules(node, &mut imports);
    }
    let mut reports = Vec::new();
    for (kind, module, start, end) in references {
        let occurrences = match kind {
            "path" => 1,
            "types" => imports
                .iter()
                .filter(|import| import.as_str() == module)
                .count(),
            _ => 0,
        };
        for _ in 0..occurrences {
            reports.push(Violation::new(node, "tripleSlashReference", format!("Do not use a triple slash reference for {module}, use `import` style instead.")).with_byte_range([start, end]));
        }
    }
    reports
}

#[cfg(test)]
mod tests {
    use super::{Directive, js_whitespace, match_directive, reference};

    #[test]
    fn reference_matcher_preserves_frozen_regexp_permissiveness() {
        for source in [
            "/ <reference path=\"foo\" />",
            "/<referencepath='foo'",
            "/\u{00a0}<reference\u{000b}path\u{000c}=\u{feff}|foo|",
        ] {
            assert_eq!(reference(source), Some(("path", "foo")));
        }
        assert_eq!(
            reference("/ <reference path=\"foo\" types=\"bar\" /> trailing"),
            Some(("path", "foo\" types=\"bar"))
        );
        assert_eq!(
            reference("/ <reference types='foo\""),
            Some(("types", "foo"))
        );
        for source in [
            "// <reference path='foo'",
            "/ <Reference path='foo'",
            "/ <reference PATH='foo'",
            "/ <reference other='first' path='foo'",
            "/\u{0085}<reference path='foo'",
            "/ <reference path=foo",
        ] {
            assert_eq!(reference(source), None, "{source}");
        }
    }

    #[test]
    fn directive_boundary_is_ascii_word_and_trim_is_ecmascript() {
        assert!(matches!(
            match_directive("@ts-ignoreλ"),
            Some((Directive::Ignore, "λ"))
        ));
        for source in ["@ts-ignored", "@ts-ignore1", "@ts-ignore_", "@ts-ignoreFoo"] {
            assert!(match_directive(source).is_none(), "{source}");
        }
        assert!(js_whitespace('\u{feff}'));
        assert!(js_whitespace('\u{2028}'));
        assert!(!js_whitespace('\u{0085}'));
    }
}
