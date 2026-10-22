//! Literal rules ported from the frozen rslint baseline.
//!
//! RegExp validation and constructor shadowing are separate prerequisites for
//! `no-regex-spaces`; a literal-only subset is deliberately not registered here.

use crate::{Violation, edits::Edit};
use tsrs_ast::{Kind, Node, SourceFile, TokenFlags};
use tsrs_checker::Checker;
use tsrs_core::P;

pub(crate) const NAMES: &[&str] = &[
    "no-octal",
    "no-nonoctal-decimal-escape",
    "no-empty-character-class",
];

pub(crate) fn check(
    name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    _checker: &mut Checker,
) -> Vec<Violation> {
    check_literal(name, node, file)
}

fn check_literal(name: &str, node: P<Node>, file: P<SourceFile>) -> Vec<Violation> {
    match (name, node.kind()) {
        ("no-octal", Kind::NumericLiteral)
            if node
                .as_numeric_literal()
                .token_flags()
                .intersects(TokenFlags::Octal | TokenFlags::ContainsLeadingZero) =>
        {
            vec![Violation::new(
                node,
                "noOctal",
                "Octal literals should not be used.",
            )]
        }
        ("no-nonoctal-decimal-escape", Kind::StringLiteral)
            if node
                .as_string_literal()
                .token_flags()
                .intersects(TokenFlags::ContainsInvalidEscape)
                || node
                    .parent()
                    .is_some_and(|p| p.kind() == Kind::JsxAttribute) =>
        {
            decimal_escapes(node, file)
        }
        ("no-empty-character-class", Kind::RegularExpressionLiteral) => {
            let text = node.text();
            let Some(body) = text.strip_prefix('/') else {
                return Vec::new();
            };
            let (pattern, flags) = body.rsplit_once('/').unwrap_or((body, ""));
            if has_empty_class(pattern, flags.contains('u'), flags.contains('v')) {
                vec![Violation::new(node, "unexpected", "Empty class.")]
            } else {
                Vec::new()
            }
        }
        _ => Vec::new(),
    }
}

fn decimal_escapes(node: P<Node>, file: P<SourceFile>) -> Vec<Violation> {
    let source = file.text();
    let raw_start = tsrs_scanner::skip_trivia(
        source,
        node.pos()
            .expect_source("parsed string has a source position"),
    ) as usize;
    let raw_end = node.end().expect_source("parsed string has an end") as usize;
    let Some(raw) = source.get(raw_start..raw_end).filter(|raw| !raw.is_empty()) else {
        return Vec::new();
    };
    let bytes = raw.as_bytes();
    let mut offset = 0;
    let mut previous_escape = None;
    let mut reports = Vec::new();
    while offset < bytes.len() {
        if bytes[offset] != b'\\' {
            offset += 1;
            previous_escape = None;
            continue;
        }
        let Some(&next) = bytes.get(offset + 1) else {
            break;
        };
        if !matches!(next, b'8' | b'9') {
            previous_escape = Some(offset);
            offset += 2;
            continue;
        }
        let digit = char::from(next);
        let range = [raw_start + offset, raw_start + offset + 2];
        let escaped = format!("\\{digit}");
        let unicode = format!("\\u003{digit}");
        let mut report = Violation::new(
            node,
            "decimalEscape",
            format!("Don't use '{escaped}' escape sequence."),
        )
        .with_byte_range(range);
        if let Some(previous) = previous_escape
            && bytes.get(previous..previous + 2) == Some(b"\\0")
        {
            let replacement = format!("\\u0000{digit}");
            report = report
                .with_suggestion(
                    "refactor",
                    format!(
                        "Replace '\\0{escaped}' with '{replacement}'. This maintains the current functionality."
                    ),
                    vec![Edit::replace([raw_start + previous, range[1]], replacement)],
                )
                .with_suggestion(
                    "refactor",
                    format!(
                        "Replace '{escaped}' with '{unicode}'. This maintains the current functionality."
                    ),
                    vec![Edit::replace(range, unicode)],
                );
        } else {
            report = report.with_suggestion(
                "refactor",
                format!(
                    "Replace '{escaped}' with '{digit}'. This maintains the current functionality."
                ),
                vec![Edit::replace(range, digit.to_string())],
            );
        }
        let replacement = format!("\\\\{digit}");
        reports.push(report.with_suggestion(
            "escapeBackslash",
            format!(
                "Replace '{escaped}' with '{replacement}' to include the actual backslash character."
            ),
            vec![Edit::replace(range, replacement)],
        ));
        offset += 2;
        previous_escape = None;
    }
    reports
}

// Match the baseline class-boundary scanner, including its recovery behavior.
// It is not a regex validator: the rule inspects complete classes even when a
// later portion of the pattern is malformed, and allows the negated empty [^].
fn has_empty_class(pattern: &str, unicode: bool, sets: bool) -> bool {
    let bytes = pattern.as_bytes();
    let mut offset = 0;
    while offset < bytes.len() {
        match bytes[offset] {
            b'\\' => {
                let Some(step) = escape_length(pattern, offset, unicode || sets) else {
                    return false;
                };
                offset += step;
            }
            b'[' => {
                let Some((end, empty)) = class_end(pattern, offset, unicode || sets, sets) else {
                    return false;
                };
                if empty {
                    return true;
                }
                offset = end;
            }
            _ => offset += 1,
        }
    }
    false
}

fn class_end(pattern: &str, start: usize, unicode: bool, sets: bool) -> Option<(usize, bool)> {
    let bytes = pattern.as_bytes();
    let mut stack = vec![start];
    let mut offset = start + 1;
    let mut empty = false;
    while offset < bytes.len() {
        match bytes[offset] {
            b'\\' => offset += escape_length(pattern, offset, unicode)?,
            b'[' if sets => {
                stack.push(offset);
                offset += 1;
            }
            b']' => {
                let open = stack.pop()?;
                empty |= offset == open + 1;
                offset += 1;
                if stack.is_empty() {
                    return Some((offset, empty));
                }
            }
            _ => offset += 1,
        }
    }
    None
}

fn escape_length(pattern: &str, offset: usize, unicode: bool) -> Option<usize> {
    let bytes = pattern.as_bytes();
    let next = *bytes.get(offset + 1)?;
    match next {
        b'x' => {
            if bytes
                .get(offset + 2..offset + 4)
                .is_some_and(|v| v.iter().all(u8::is_ascii_hexdigit))
            {
                return Some(4);
            }
            return Some(2);
        }
        b'u' => {
            if bytes.get(offset + 2) == Some(&b'{') {
                if unicode && let Some(close) = bytes[offset + 3..].iter().position(|&b| b == b'}')
                {
                    return Some(close + 4);
                }
                return Some(2);
            }
            if bytes
                .get(offset + 2..offset + 6)
                .is_some_and(|v| v.iter().all(u8::is_ascii_hexdigit))
            {
                return Some(6);
            }
            return Some(2);
        }
        b'c' => {
            if bytes.get(offset + 2).is_some_and(|&b| {
                b.is_ascii_alphabetic() || (!unicode && (b.is_ascii_digit() || b == b'_'))
            }) {
                return Some(3);
            }
            return (!unicode).then_some(1);
        }
        b'p' | b'P' | b'q' => {
            if unicode
                && bytes.get(offset + 2) == Some(&b'{')
                && let Some(close) = bytes[offset + 3..].iter().position(|&b| b == b'}')
            {
                return Some(close + 4);
            }
            return Some(2);
        }
        _ => {}
    }
    Some(1 + pattern[offset + 1..].chars().next()?.len_utf8())
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::edits::merge_edits;
    use tsrs_ast::SourceFileParseOptions;
    use tsrs_core::{ScriptKind, arena::Region, tspath::Path};

    fn with_reports(source: &str, rule: &str, inspect: impl FnOnce(Vec<Violation>)) {
        let region = Region::new(1 << 20);
        let _scope = region.enter();
        let path = "/rslint-literal-tests/case.js";
        let file = tsrs_parser::parse_source_file(
            SourceFileParseOptions {
                file_name: path.into(),
                path: Path::new(path),
                ..Default::default()
            },
            source,
            ScriptKind::JS,
        );
        fn visit(node: P<Node>, file: P<SourceFile>, rule: &str, reports: &mut Vec<Violation>) {
            reports.extend(check_literal(rule, node, file));
            node.for_each_child(&mut |child| {
                visit(child, file, rule, reports);
                false
            });
        }
        let mut reports = Vec::new();
        visit(file.as_node(), file, rule, &mut reports);
        inspect(reports);
    }

    #[test]
    fn octal_flags_preserve_legacy_and_leading_zero_tokens() {
        // The baseline also tests these against the parsed AST because its
        // tsconfig-bound tester rejects TS1121/TS1489. The Rust public API's
        // syntax policy is covered separately; no parse guard is bypassed there.
        for token in [
            "01234", "07", "00", "08", "09.1", "09e1", "09.1e1", "018", "019.1", "019e1", "019.1e1",
        ] {
            let source = format!("/* trivia */ -{token};");
            with_reports(&source, "no-octal", |reports| {
                assert_eq!(reports.len(), 1, "{source}");
                let report = &reports[0];
                assert_eq!(report.message_id, "noOctal");
                assert_eq!(report.message, "Octal literals should not be used.");
                assert_eq!(report.node.text(), report.node.as_numeric_literal().text());
                let start = tsrs_scanner::skip_trivia(
                    &source,
                    report.node.pos().expect_source("parsed token"),
                ) as usize;
                let end = report.node.end().expect_source("parsed token") as usize;
                assert_eq!(&source[start..end], token);
                assert!(report.edits.is_empty());
                assert!(report.suggestions.is_empty());
            });
        }
        for token in [
            "0", "0.1", "0.5e1", "0x1234", "0X5", "0o17", "0O17", "0b101", "0B101", "0n", "1",
            "123", "1.5", "1e5", ".5",
        ] {
            with_reports(token, "no-octal", |reports| {
                assert!(reports.is_empty(), "{token}")
            });
        }
        for (source, token) in [("01_2", "01"), ("08_9", "08")] {
            with_reports(source, "no-octal", |reports| {
                assert_eq!(reports.len(), 1);
                assert_eq!(
                    &source[..reports[0].node.end().expect_source("parsed token") as usize],
                    token
                );
            });
        }
    }

    #[test]
    fn decimal_escape_pairs_reset_and_suggestions_keep_exact_source_ranges() {
        for (source, count, suggestion_counts) in [
            (r"'\8'", 1, vec![2]),
            (r"'\9'", 1, vec![2]),
            (r"'\0\8\9'", 2, vec![3, 2]),
            (r"'\8\0\9'", 2, vec![2, 3]),
            (r"'\0 \8'", 1, vec![2]),
            (r"'\0\1\8'", 1, vec![2]),
            (r"'\01\8'", 1, vec![2]),
            (r"'\\\8'", 1, vec![2]),
            (r"'\\0\8'", 1, vec![2]),
            ("'\\😀\\8'", 1, vec![2]),
            ("'\\\r\n\\9'", 1, vec![2]),
        ] {
            with_reports(source, "no-nonoctal-decimal-escape", |reports| {
                assert_eq!(reports.len(), count, "{source}");
                assert_eq!(
                    reports
                        .iter()
                        .map(|r| r.suggestions.len())
                        .collect::<Vec<_>>(),
                    suggestion_counts
                );
                for report in reports {
                    let [start, end] = report.byte_range.unwrap();
                    assert!(matches!(&source[start..end], r"\8" | r"\9"));
                    assert_eq!(report.message_id, "decimalEscape");
                    assert!(
                        report.edits.is_empty(),
                        "suggestions must never become auto-fixes"
                    );
                    assert_eq!(
                        report.suggestions.last().unwrap().message_id,
                        "escapeBackslash"
                    );
                    for suggestion in report.suggestions {
                        assert!(merge_edits(source, &suggestion.edits).unwrap().is_some());
                    }
                }
            });
        }
        let source = "\u{feff}'😀\\0\\8';";
        with_reports(source, "no-nonoctal-decimal-escape", |reports| {
            assert_eq!(reports.len(), 1);
            let report = &reports[0];
            assert_eq!(report.message, "Don't use '\\8' escape sequence.");
            let suggestions = &report.suggestions;
            assert_eq!(
                suggestions[0].message,
                "Replace '\\0\\8' with '\\u00008'. This maintains the current functionality."
            );
            assert_eq!(
                suggestions[1].message,
                "Replace '\\8' with '\\u0038'. This maintains the current functionality."
            );
            assert_eq!(
                suggestions[2].message,
                "Replace '\\8' with '\\\\8' to include the actual backslash character."
            );
            for (suggestion, range, text) in [
                (&suggestions[0], [3, 7], r"\u00008"),
                (&suggestions[1], [5, 7], r"\u0038"),
                (&suggestions[2], [5, 7], r"\\8"),
            ] {
                let fix = merge_edits(source, &suggestion.edits).unwrap().unwrap();
                assert_eq!(fix.range, range);
                assert_eq!(fix.text, text);
            }
        });
        for source in [
            r"'foo'",
            r"'8'",
            r"'9'",
            r"'\1'",
            r"'\0'",
            r"'\\8'",
            r"'\\9'",
            r"'\\\\\\8'",
            r"String.raw`\8`",
        ] {
            with_reports(source, "no-nonoctal-decimal-escape", |reports| {
                assert!(reports.is_empty(), "{source}")
            });
        }
    }

    #[test]
    fn character_classes_respect_escapes_nested_sets_and_malformed_boundaries() {
        for (pattern, sets, expected) in [
            ("[]", false, true),
            ("[^]", false, false),
            ("[[]", false, false),
            (r"\[]", false, false),
            (r"\\[]", false, true),
            (r"[\]]", false, false),
            (r"[\[]", false, false),
            (r"\[[]", false, true),
            ("[][]", false, true),
            ("[[]]", true, true),
            ("[[^]]", true, false),
            (r"[[\]]]", true, false),
            ("[a--[]]", true, true),
            ("[[]&&b]", true, true),
            ("[a[b[]]]", true, true),
            (r"[\q{ab|c]d}]", true, false),
            (r"[\q{[]}]", true, false),
            (r"\p{Letter}[a]", false, false),
            ("[abc", false, false),
            ("[[]", true, false),
            ("[]\\", true, true),
            (r"\c[]", false, true),
            (r"\c[]", true, false),
            (r"\u{[]}[]", true, true),
        ] {
            assert_eq!(
                has_empty_class(pattern, false, sets),
                expected,
                "{pattern}, v={sets}"
            );
        }
        // Iterative nesting preserves the baseline's stack-safe class-end scan.
        let pattern = format!("{}{}", "[".repeat(10_000), "]".repeat(10_000));
        assert!(has_empty_class(&pattern, false, true));
    }
}
