//! The frozen rsbuild configuration enables these rules with default options.
//! Numeric comparison uses the exact binary rational and ECMAScript's half-up
//! decimal rounding; it must not use Rust's ties-to-even decimal formatting.

use crate::Violation;
use num_bigint::BigUint;
use num_traits::{One, Zero};
use tsrs_ast::{Kind, Node, SourceFile, is_optional_chain_root, walk_up_parenthesized_expressions};
use tsrs_checker::Checker;
use tsrs_core::P;

pub(crate) const NAMES: &[&str] = &[
    "no-irregular-whitespace",
    "no-unexpected-multiline",
    "no-loss-of-precision",
];

pub(crate) fn check(
    name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    _checker: &mut Checker,
) -> Vec<Violation> {
    check_lexical(name, node, file)
}

fn check_lexical(name: &str, node: P<Node>, file: P<SourceFile>) -> Vec<Violation> {
    match name {
        "no-irregular-whitespace" if node.kind() == Kind::SourceFile => {
            irregular_whitespace(node, file)
        }
        "no-unexpected-multiline" => unexpected_multiline(node, file).into_iter().collect(),
        "no-loss-of-precision" if node.kind() == Kind::NumericLiteral => {
            let start = token_start(node, file);
            let end = node_end(node);
            if loses_precision(&file.text()[start..end]) {
                vec![
                    Violation::new(
                        node,
                        "noLossOfPrecision",
                        "This number literal will lose precision at runtime.",
                    )
                    .with_byte_range([start, end]),
                ]
            } else {
                Vec::new()
            }
        }
        _ => Vec::new(),
    }
}

fn token_start(node: P<Node>, file: P<SourceFile>) -> usize {
    tsrs_scanner::skip_trivia(file.text(), node.pos().expect_source("parsed lexical node")) as usize
}

fn node_end(node: P<Node>) -> usize {
    node.end().expect_source("parsed lexical node") as usize
}

fn irregular(ch: char) -> bool {
    matches!(
        ch,
        '\u{000b}' | '\u{000c}' | '\u{0085}' | '\u{00a0}' | '\u{1680}' | '\u{180e}' | '\u{2000}'
            ..='\u{200b}'
                | '\u{2028}'
                | '\u{2029}'
                | '\u{202f}'
                | '\u{205f}'
                | '\u{3000}'
                | '\u{feff}'
    )
}

fn irregular_ranges(source: &str) -> Vec<[usize; 2]> {
    let mut result = Vec::new();
    let mut chars = source.char_indices().peekable();
    while let Some((start, ch)) = chars.next() {
        if !irregular(ch) || (start == 0 && ch == '\u{feff}') {
            continue;
        }
        let mut end = start + ch.len_utf8();
        if !matches!(ch, '\u{2028}' | '\u{2029}') {
            while let Some(&(pos, next)) = chars.peek() {
                if !irregular(next) || matches!(next, '\u{2028}' | '\u{2029}') {
                    break;
                }
                end = pos + next.len_utf8();
                chars.next();
            }
        }
        result.push([start, end]);
    }
    result
}

fn irregular_whitespace(node: P<Node>, file: P<SourceFile>) -> Vec<Violation> {
    let ranges = irregular_ranges(file.text());
    if ranges.is_empty() {
        return Vec::new();
    }
    // Default skipStrings=true. Templates, regexes, JSX text and comments are
    // deliberately not exempt. AST token starts exclude leading trivia.
    fn strings(node: P<Node>, file: P<SourceFile>, ranges: &mut Vec<[usize; 2]>) {
        if node.kind() == Kind::StringLiteral {
            ranges.push([token_start(node, file), node_end(node)]);
        }
        node.for_each_child(&mut |child| {
            strings(child, file, ranges);
            false
        });
    }
    let mut exempt = Vec::new();
    strings(node, file, &mut exempt);
    exempt.sort_unstable();
    let mut cursor = 0;
    ranges
        .into_iter()
        .filter_map(|range| {
            while cursor < exempt.len() && exempt[cursor][1] <= range[0] {
                cursor += 1;
            }
            if exempt
                .get(cursor)
                .is_some_and(|e| e[0] <= range[0] && e[1] >= range[1])
            {
                return None;
            }
            Some(
                Violation::new(
                    node,
                    "noIrregularWhitespace",
                    "Irregular whitespace not allowed.",
                )
                .with_byte_range(range),
            )
        })
        .collect()
}

fn line_terminator(text: &str, start: usize, end: usize) -> bool {
    text.get(start..end)
        .is_some_and(|gap| gap.contains(['\r', '\n', '\u{2028}', '\u{2029}']))
}

fn newline_after(
    expr: P<Node>,
    node: P<Node>,
    file: P<SourceFile>,
    id: &'static str,
    message: &'static str,
) -> Option<Violation> {
    let end = node_end(expr);
    let start = tsrs_scanner::skip_trivia(file.text(), end as u32) as usize;
    (start < file.text().len() && line_terminator(file.text(), end, start))
        .then(|| Violation::new(node, id, message).with_byte_range([start, start + 1]))
}

fn unexpected_multiline(node: P<Node>, file: P<SourceFile>) -> Option<Violation> {
    match node.kind() {
        Kind::ElementAccessExpression if !is_optional_chain_root(node) => newline_after(
            node.as_element_access_expression().expression(),
            node,
            file,
            "property",
            "Unexpected newline between object and [ of property access.",
        ),
        Kind::CallExpression if !is_optional_chain_root(node) && !node.arguments().is_empty() => {
            newline_after(
                node.as_call_expression().expression(),
                node,
                file,
                "function",
                "Unexpected newline between function and ( of function call.",
            )
        }
        Kind::TaggedTemplateExpression => {
            let template = node.as_tagged_template_expression().template();
            let start = token_start(template, file);
            let trivia_start = template.pos().expect_source("parsed template") as usize;
            (start < file.text().len()
                && trivia_start > 0
                && line_terminator(file.text(), trivia_start, start))
            .then(|| {
                Violation::new(
                    node,
                    "taggedTemplate",
                    "Unexpected newline between template tag and template literal.",
                )
                .with_byte_range([start, start + 1])
            })
        }
        Kind::BinaryExpression => {
            let binary = node.as_binary_expression();
            if binary.operator_token().kind() != Kind::SlashToken {
                return None;
            }
            let slash_end = node_end(binary.operator_token());
            if !line_terminator(file.text(), node_end(binary.left()), slash_end - 1) {
                return None;
            }
            let child = walk_up_parenthesized_expressions(node)?;
            let parent = child.parent()?;
            if parent.kind() != Kind::BinaryExpression {
                return None;
            }
            let outer = parent.as_binary_expression();
            if outer.operator_token().kind() != Kind::SlashToken
                || outer.left() != child
                || !regex_flag_identifier(file.text(), node_end(outer.operator_token()))
            {
                return None;
            }
            Some(
                Violation::new(
                    node,
                    "division",
                    "Unexpected newline between numerator and division operator.",
                )
                .with_byte_range([slash_end - 1, slash_end]),
            )
        }
        _ => None,
    }
}

fn regex_flag(ch: char) -> bool {
    matches!(ch, 'g' | 'i' | 'm' | 's' | 'u' | 'y')
}

fn unicode_escape(text: &str, pos: usize) -> Option<(u32, usize)> {
    let remaining = text.get(pos..)?.strip_prefix("\\u")?;
    if let Some(braced) = remaining.strip_prefix('{') {
        let digits = braced.bytes().take_while(u8::is_ascii_hexdigit).count();
        if digits == 0 || braced.as_bytes().get(digits) != Some(&b'}') {
            return None;
        }
        // Leading zeroes are permitted even when their count exceeds eight.
        let mut value = 0_u32;
        for digit in braced[..digits].chars() {
            value = value.checked_mul(16)?.checked_add(digit.to_digit(16)?)?;
            if value > 0x10ffff {
                return None;
            }
        }
        Some((value, pos + 4 + digits))
    } else {
        let digits = remaining.as_bytes().get(..4)?;
        if !digits.iter().all(u8::is_ascii_hexdigit) {
            return None;
        }
        Some((u32::from_str_radix(&remaining[..4], 16).ok()?, pos + 6))
    }
}

fn regex_flag_identifier(text: &str, mut pos: usize) -> bool {
    let mut matched = false;
    while let Some(ch) = text.get(pos..).and_then(|tail| tail.chars().next()) {
        if regex_flag(ch) {
            matched = true;
            pos += 1;
        } else if ch == '\\' {
            let Some((decoded, end)) = unicode_escape(text, pos) else {
                return matched;
            };
            if matched {
                if !tsrs_scanner::is_identifier_part(decoded) {
                    return true;
                }
            } else if !tsrs_scanner::is_identifier_start(decoded) {
                return false;
            }
            if !char::from_u32(decoded).is_some_and(regex_flag) {
                return false;
            }
            matched = true;
            pos = end;
        } else {
            return matched && !tsrs_scanner::is_identifier_part(ch as u32);
        }
    }
    matched
}

fn loses_precision(raw: &str) -> bool {
    let normalized = raw.replace('_', "");
    let bytes = normalized.as_bytes();
    if bytes.first() == Some(&b'0') && bytes.len() >= 2 {
        let base = match bytes[1] {
            b'b' | b'B' => 2,
            b'o' | b'O' => 8,
            b'x' | b'X' => 16,
            _ => 0,
        };
        if base != 0 {
            return non_decimal_loses_precision(&normalized[2..], base);
        }
        if bytes[1..].iter().all(|b| matches!(b, b'0'..=b'7')) {
            return non_decimal_loses_precision(&normalized[1..], 8);
        }
    }
    const EXACT_INTEGER: &str = "9007199254740992";
    if !bytes.is_empty() && bytes.iter().all(u8::is_ascii_digit) {
        let significant = normalized.trim_start_matches('0');
        if significant.len() < EXACT_INTEGER.len()
            || (significant.len() == EXACT_INTEGER.len() && significant <= EXACT_INTEGER)
        {
            return false;
        }
    }
    let Ok(value) = normalized.parse::<f64>() else {
        return false;
    };
    if value == 0.0 || value.is_nan() {
        return false;
    }
    if value.is_infinite() {
        return true;
    }
    let original = scientific(&normalized);
    let precision = original.0.len();
    if precision > 100 {
        return true;
    }
    original != precision_scientific(value.abs(), precision.max(1))
}

fn non_decimal_loses_precision(digits: &str, base: u32) -> bool {
    if digits.is_empty() || digits.starts_with(['+', '-']) {
        return false;
    }
    let Some(integer) = BigUint::parse_bytes(digits.as_bytes(), base) else {
        return false;
    };
    !integer.is_zero()
        && (integer.bits() > 1024 || integer.bits() - integer.trailing_zeros().unwrap() > 53)
}

fn trim_leading_zeros(value: &str) -> &str {
    let trimmed = value.trim_start_matches('0');
    if trimmed.is_empty() { value } else { trimmed }
}

fn scientific(raw: &str) -> (String, i32) {
    let (coefficient, exponent) = raw
        .split_once(['e', 'E'])
        .map_or((raw, 0), |(coefficient, exponent)| {
            (coefficient, exponent.parse::<i32>().unwrap_or(0))
        });
    let trimmed = trim_leading_zeros(coefficient);
    let (digits, magnitude) = if raw.contains('.') {
        match trimmed.find('.') {
            Some(0) => {
                let digits = trim_leading_zeros(&trimmed[1..]);
                (
                    digits.to_owned(),
                    digits.len() as i32 - trimmed.len() as i32,
                )
            }
            Some(dot) => (trimmed.replace('.', ""), dot as i32 - 1),
            None => (trimmed.to_owned(), trimmed.len() as i32 - 1),
        }
    } else {
        let digits = trimmed.trim_end_matches('0');
        (
            if digits.is_empty() { trimmed } else { digits }.to_owned(),
            trimmed.len() as i32 - 1,
        )
    };
    (digits, magnitude + exponent)
}

fn decimal_power(exponent: u32) -> BigUint {
    BigUint::from(10_u32).pow(exponent)
}

fn compare_power(numerator: &BigUint, denominator: &BigUint, exponent: i32) -> std::cmp::Ordering {
    if exponent >= 0 {
        numerator.cmp(&(denominator * decimal_power(exponent as u32)))
    } else {
        (numerator * decimal_power((-exponent) as u32)).cmp(denominator)
    }
}

fn precision_scientific(value: f64, precision: usize) -> (String, i32) {
    let bits = value.to_bits();
    let exponent_bits = ((bits >> 52) & 0x7ff) as i32;
    let fraction = bits & ((1_u64 << 52) - 1);
    let (significand, exponent) = if exponent_bits == 0 {
        (fraction, -1074)
    } else {
        (fraction | (1_u64 << 52), exponent_bits - 1023 - 52)
    };
    let mut numerator = BigUint::from(significand);
    let mut denominator = BigUint::one();
    if exponent >= 0 {
        numerator <<= exponent as usize;
    } else {
        denominator <<= (-exponent) as usize;
    }
    let mut magnitude = value.log10().floor() as i32;
    while compare_power(&numerator, &denominator, magnitude).is_lt() {
        magnitude -= 1;
    }
    while !compare_power(&numerator, &denominator, magnitude + 1).is_lt() {
        magnitude += 1;
    }
    let scale = precision as i32 - 1 - magnitude;
    if scale >= 0 {
        numerator *= decimal_power(scale as u32);
    } else {
        denominator *= decimal_power((-scale) as u32);
    }
    let mut quotient = &numerator / &denominator;
    let remainder = numerator % &denominator;
    if (remainder << 1_usize) >= denominator {
        quotient += 1_u32;
    }
    let mut coefficient = quotient.to_str_radix(10);
    if coefficient.len() > precision {
        magnitude += (coefficient.len() - precision) as i32;
        coefficient.truncate(precision);
    }
    if coefficient.len() < precision {
        coefficient = format!(
            "{}{}",
            "0".repeat(precision - coefficient.len()),
            coefficient
        );
    }
    (coefficient, magnitude)
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::{Value, json};
    use tsrs_ast::SourceFileParseOptions;
    use tsrs_core::{ScriptKind, arena::Region, tspath::Path};

    fn with_reports(source: &str, rule: &str, tsx: bool, inspect: impl FnOnce(Vec<Violation>)) {
        let region = Region::new(1 << 20);
        let _scope = region.enter();
        let path = if tsx {
            "/lexical/input.tsx"
        } else {
            "/lexical/input.ts"
        };
        let file = tsrs_parser::parse_source_file(
            SourceFileParseOptions {
                file_name: path.into(),
                path: Path::new(path),
                ..Default::default()
            },
            source,
            if tsx { ScriptKind::TSX } else { ScriptKind::TS },
        );
        fn visit(node: P<Node>, file: P<SourceFile>, rule: &str, reports: &mut Vec<Violation>) {
            reports.extend(check_lexical(rule, node, file));
            node.for_each_child(&mut |child| {
                visit(child, file, rule, reports);
                false
            });
        }
        let mut reports = Vec::new();
        visit(file.as_node(), file, rule, &mut reports);
        reports.sort_by_key(|report| report.byte_range.unwrap());
        inspect(reports);
    }

    fn point(source: &str, position: usize) -> (usize, usize) {
        let mut line = 1;
        let mut column = 1;
        let mut was_cr = false;
        for ch in source[..position].chars() {
            match ch {
                '\n' if was_cr => {}
                '\n' | '\r' | '\u{2028}' | '\u{2029}' => {
                    line += 1;
                    column = 1;
                }
                _ => {
                    column += ch.len_utf16();
                }
            }
            was_cr = ch == '\r';
        }
        (line, column)
    }

    #[test]
    fn frozen_listener_corpus_preserves_every_default_diagnostic_range() {
        let groups: Value = serde_json::from_str(include_str!(
            "../../../../scripts/tsrs/oracle-cases/rsbuild-lexical-checks.json"
        ))
        .unwrap();
        for group in groups.as_array().unwrap() {
            let rule = group["rule"].as_str().unwrap();
            for category in ["valid", "invalid", "knownBaselineSkipped"] {
                for case in group[category].as_array().unwrap() {
                    let source = case["text"].as_str().unwrap();
                    let tsx = case["extension"] == ".tsx";
                    with_reports(source, rule, tsx, |reports| {
                        let expected = case
                            .get("expectedRuleDiagnostics")
                            .and_then(Value::as_array)
                            .cloned()
                            .unwrap_or_default();
                        assert_eq!(reports.len(), expected.len(), "{rule}: {source:?}");
                        for (report, want) in reports.iter().zip(expected) {
                            let [start, end] = report.byte_range.unwrap();
                            let (line, column) = point(source, start);
                            let (end_line, end_column) = point(source, end);
                            let got = json!({"messageId":report.message_id, "message":report.message, "line":line,"column":column,"endLine":end_line,"endColumn":end_column});
                            for (key, value) in want.as_object().unwrap() {
                                assert_eq!(&got[key], value, "{rule}: {source:?}: {key}");
                            }
                            assert!(report.edits.is_empty(), "rule never offers a fix");
                            assert!(
                                report.suggestions.is_empty(),
                                "rule never offers a suggestion"
                            );
                        }
                    });
                }
            }
        }
    }

    #[test]
    fn decimal_rounding_uses_exact_binary_value_and_half_up() {
        for raw in [
            "255.10000610351562",
            "555.9771118164062",
            "9726.622680664062",
            "4e-324",
            "1.7976931348623159e308",
            "9999999999999999",
            "99.99999999999998",
            "0009007199254740993",
            "0400000000000000001",
        ] {
            assert!(loses_precision(raw), "{raw}");
        }
        for raw in [
            "123.456",
            "5e-324",
            "1.7976931348623157e308",
            "9999999999999998",
            "9007199254740992",
            "019.5",
            "00195",
            "0008",
            "0377777777777777777",
            "0x40000000000000",
            "1e-9999",
            "0.000000000000000000000000000000000000000000000000000000000000000000000000000000",
        ] {
            assert!(!loses_precision(raw), "{raw}");
        }
        assert_eq!(
            precision_scientific(255.10000610351562, 17),
            ("25510000610351563".into(), 2)
        );
        assert_eq!(precision_scientific(1.25, 2), ("13".into(), 0));
        assert_eq!(precision_scientific(9.99, 2), ("10".into(), 1));
    }

    #[test]
    fn finite_binary_integer_boundary_keeps_shifted_exact_numbers() {
        for bits in 1..=1050 {
            for significant in [1, 2, 52, 53, 54, 60] {
                if significant > bits {
                    continue;
                }
                let mut number: BigUint = BigUint::one() << (significant - 1);
                if significant > 1 {
                    number += 1_u32;
                }
                number <<= bits - significant;
                for base in [2, 8, 16] {
                    assert_eq!(
                        non_decimal_loses_precision(&number.to_str_radix(base), base),
                        bits > 1024 || significant > 53,
                        "{bits}/{significant}/{base}"
                    );
                }
            }
        }
        for invalid in ["", "+1", "-1", "xyz"] {
            assert!(!non_decimal_loses_precision(invalid, 16));
        }
    }

    #[test]
    fn unicode_whitespace_set_and_group_boundaries_are_exact() {
        for scalar in 0..=0x10ffff {
            if let Some(ch) = char::from_u32(scalar) {
                let want = [
                    0x0b, 0x0c, 0x85, 0xa0, 0x1680, 0x180e, 0x2028, 0x2029, 0x202f, 0x205f, 0x3000,
                    0xfeff,
                ]
                .contains(&scalar)
                    || (0x2000..=0x200b).contains(&scalar);
                assert_eq!(irregular(ch), want, "U+{scalar:04X}");
            }
        }
        assert_eq!(
            irregular_ranges("\u{feff}x\u{000b}\u{00a0}\u{2028}\u{2029}\u{3000}x"),
            vec![[4, 7], [7, 10], [10, 13], [13, 16]]
        );
    }

    #[test]
    fn escaped_regex_flag_identifier_uses_decoded_value_and_full_token() {
        for text in [
            "g",
            "gimsuy",
            "g\\u0069",
            "\\u{00000000000067}i",
            "g;",
            "g\\u{20}",
        ] {
            assert!(regex_flag_identifier(text, 0), "{text}");
        }
        for text in [
            "d",
            "v",
            "gym2",
            "g_$",
            "g中文",
            "\\u0061",
            "\\u{110000}",
            "GYM",
            "2",
            " g",
        ] {
            assert!(!regex_flag_identifier(text, 0), "{text}");
        }
    }
}
