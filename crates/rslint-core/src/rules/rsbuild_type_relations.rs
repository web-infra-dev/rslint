//! Type constituent relations, preserving the frozen rslint rule algorithms.
//! All type handles stay within one checker/snapshot traversal.

use super::go_quote::quote as quote_literal;
use std::collections::HashMap;

use crate::{Violation, edits::Edit};
use tsrs_ast::{Kind, Node, SourceFile, is_function_like, skip_parentheses, skip_type_parentheses};
use tsrs_checker::{Checker, Type, TypeFlags};
use tsrs_core::{P, TextRange};
use tsrs_scanner::{get_scanner_for_source_file, skip_trivia};

pub(crate) const NAMES: &[&str] = &[
    "@typescript-eslint/no-duplicate-type-constituents",
    "@typescript-eslint/no-redundant-type-constituents",
];

pub(crate) fn check(
    name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
) -> Vec<Violation> {
    if !matches!(node.kind(), Kind::UnionType | Kind::IntersectionType) {
        return Vec::new();
    }
    match name {
        "@typescript-eslint/no-duplicate-type-constituents" => duplicate_types(node, file, checker),
        "@typescript-eslint/no-redundant-type-constituents"
            if node.kind() == Kind::IntersectionType =>
        {
            redundant_intersection(node, checker)
        }
        "@typescript-eslint/no-redundant-type-constituents" => redundant_union(node, checker),
        _ => Vec::new(),
    }
}

fn pos(node: P<Node>) -> u32 {
    node.pos().expect_source("parsed type node position")
}
fn end(node: P<Node>) -> u32 {
    node.end().expect_source("parsed type node end")
}
fn trim_range(node: P<Node>, file: P<SourceFile>) -> [usize; 2] {
    [
        skip_trivia(file.text(), pos(node)) as usize,
        end(node) as usize,
    ]
}
fn remove_token(range: TextRange) -> Edit {
    Edit::replace(
        [
            range.pos().expect_source("source token position") as usize,
            range.end().expect_source("source token end") as usize,
        ],
        "",
    )
}
fn is_error(value: P<Type>) -> bool {
    value.flags().intersects(TypeFlags::Intrinsic)
        && value.as_intrinsic_type().intrinsic_name() == "error"
}
fn constituents(node: P<Node>) -> &'static [P<Node>] {
    match node.kind() {
        Kind::UnionType => node.as_union_type_node().types().nodes(),
        Kind::IntersectionType => node.as_intersection_type_node().types().nodes(),
        _ => &[],
    }
}
fn same_parent(mut node: P<Node>, kind: Kind) -> Option<P<Node>> {
    loop {
        node = node.parent()?;
        if node.kind() == kind {
            return Some(node);
        }
        if node.kind() != Kind::ParenthesizedType {
            return None;
        }
    }
}

fn duplicate_edits(node: P<Node>, kind: Kind, file: P<SourceFile>) -> Vec<Edit> {
    let parent = same_parent(node, kind).expect("constituent belongs to its union/intersection");
    let mut scanner = get_scanner_for_source_file(file, pos(parent));
    let mut separator = None;
    let mut opening = Vec::new();
    while scanner.token_start() < pos(node) {
        match scanner.token() {
            Kind::AmpersandToken | Kind::BarToken => {
                separator = Some(scanner.token_start());
                opening.clear();
            }
            Kind::OpenParenToken => opening.push(scanner.token_range()),
            _ => {}
        }
        scanner.scan();
    }
    let mut edits = vec![Edit::replace(trim_range(node, file), "")];
    if let Some(separator) = separator {
        edits.push(Edit::replace(
            [separator as usize, separator as usize + 1],
            "",
        ));
        edits.extend(opening.iter().copied().map(remove_token));
        scanner.reset_pos(end(node));
        for _ in &opening {
            scanner.scan();
            assert_eq!(
                scanner.token(),
                Kind::CloseParenToken,
                "duplicate constituent parentheses must balance"
            );
            edits.push(remove_token(scanner.token_range()));
        }
    } else {
        scanner.reset_pos(end(node));
        let mut closing_count = 0;
        loop {
            scanner.scan();
            assert!(
                scanner.token_start() < end(parent),
                "constituent must have a following separator"
            );
            if matches!(scanner.token(), Kind::AmpersandToken | Kind::BarToken) {
                edits.push(remove_token(scanner.token_range()));
                break;
            }
            assert_eq!(
                scanner.token(),
                Kind::CloseParenToken,
                "duplicate constituent parentheses must balance"
            );
            closing_count += 1;
            edits.push(remove_token(scanner.token_range()));
        }
        // Preserve the frozen scanner's removal ordering for nested optional
        // undefined constituents, including its trivia-preserving boundaries.
        let mut opening = Vec::new();
        scanner.reset_pos(pos(parent));
        for _ in 0..closing_count {
            scanner.scan();
            if scanner.token() == Kind::OpenParenToken {
                if opening.len() < closing_count {
                    opening.push(scanner.token_range());
                }
            } else {
                opening.clear();
            }
            if scanner.token_start() == pos(node) {
                assert_eq!(
                    opening.len(),
                    closing_count,
                    "constituent opening parentheses must balance"
                );
                break;
            }
            edits.extend(opening.iter().copied().map(remove_token));
        }
    }
    edits
}

struct DuplicateState<'a> {
    checker: &'a mut Checker,
    file: P<SourceFile>,
    kind: Kind,
    optional_parameter: bool,
    seen: HashMap<P<Type>, P<Node>>,
    reports: Vec<Violation>,
}

impl DuplicateState<'_> {
    fn visit(&mut self, node: P<Node>, with_fix: bool) -> bool {
        let value = self.checker.get_type_at_location(node);
        if is_error(value) {
            return false;
        }
        if let Some(&previous) = self.seen.get(&value) {
            // The frozen message deliberately uses the prior node's full
            // source range, preserving its leading trivia.
            let previous = &self.file.text()[pos(previous) as usize..end(previous) as usize];
            let container = if self.kind == Kind::UnionType {
                "Union"
            } else {
                "Intersection"
            };
            let mut report = Violation::new(
                node,
                "duplicate",
                format!("{container} type constituent is duplicated with {previous}."),
            );
            if with_fix {
                report = report.with_fix(duplicate_edits(node, self.kind, self.file));
            }
            self.reports.push(report);
            return true;
        }
        if self.optional_parameter && value.flags().intersects(TypeFlags::Undefined) {
            self.reports.push(
                Violation::new(
                    node,
                    "unnecessary",
                    "Explicit undefined is unnecessary on an optional parameter.",
                )
                .with_fix(duplicate_edits(node, self.kind, self.file)),
            );
        }
        self.seen.insert(value, node);
        let unwrapped = skip_type_parentheses(node);
        if unwrapped.kind() != self.kind {
            return false;
        }
        let types = constituents(unwrapped);
        let mut all_removed = true;
        for (index, &part) in types.iter().enumerate() {
            if !self.visit(part, index < types.len() - 1 || !all_removed) {
                all_removed = false;
            }
        }
        false
    }
}

fn duplicate_types(node: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> Vec<Violation> {
    if same_parent(node, node.kind()).is_some() {
        return Vec::new();
    }
    let optional_parameter = node.kind() == Kind::UnionType
        && node.parent().is_some_and(|parent| {
            parent.kind() == Kind::Parameter
                && parent.as_parameter_declaration().question_token().is_some()
        });
    let mut state = DuplicateState {
        checker,
        file,
        kind: node.kind(),
        optional_parameter,
        seen: HashMap::new(),
        reports: Vec::new(),
    };
    for &part in constituents(node) {
        state.visit(part, true);
    }
    state.reports
}

#[derive(Clone, Copy)]
struct Part {
    flags: TypeFlags,
    node: Option<P<Node>>,
    value: Option<P<Type>>,
}
impl Part {
    fn display(self, checker: &mut Checker) -> String {
        let Some(node) = self.node else {
            return checker.type_to_string_exported(self.value.expect("type-backed constituent"));
        };
        let spelling = match node.kind() {
            Kind::AnyKeyword => "any",
            Kind::BooleanKeyword => "boolean",
            Kind::NeverKeyword => "never",
            Kind::NumberKeyword => "number",
            Kind::StringKeyword => "string",
            Kind::UnknownKeyword => "unknown",
            Kind::LiteralType => {
                let literal = node.as_literal_type_node().literal();
                return match literal.kind() {
                    Kind::TemplateLiteralType | Kind::NoSubstitutionTemplateLiteral => {
                        "template literal type".into()
                    }
                    Kind::StringLiteral => quote_literal(literal.text()),
                    Kind::NumericLiteral | Kind::BigIntLiteral => literal.text().into(),
                    Kind::TrueKeyword => "true".into(),
                    Kind::FalseKeyword => "false".into(),
                    _ => "literal type".into(),
                };
            }
            _ => "literal type",
        };
        spelling.into()
    }
}

fn type_parts(node: P<Node>, checker: &mut Checker) -> Vec<Part> {
    let node = skip_parentheses(node);
    let direct = match node.kind() {
        Kind::AnyKeyword => TypeFlags::Any,
        Kind::BigIntKeyword => TypeFlags::BigInt,
        Kind::BooleanKeyword => TypeFlags::Boolean,
        Kind::NeverKeyword => TypeFlags::Never,
        Kind::NumberKeyword => TypeFlags::Number,
        Kind::StringKeyword => TypeFlags::String,
        Kind::UnknownKeyword => TypeFlags::Unknown,
        Kind::LiteralType => match node.as_literal_type_node().literal().kind() {
            Kind::BigIntLiteral => TypeFlags::BigIntLiteral,
            Kind::TrueKeyword | Kind::FalseKeyword => TypeFlags::BooleanLiteral,
            Kind::NumericLiteral => TypeFlags::NumberLiteral,
            Kind::StringLiteral => TypeFlags::StringLiteral,
            _ => TypeFlags::empty(),
        },
        _ => TypeFlags::empty(),
    };
    if !direct.is_empty() {
        return vec![Part {
            flags: direct,
            node: Some(node),
            value: None,
        }];
    }
    if node.kind() == Kind::UnionType {
        return constituents(node)
            .iter()
            .flat_map(|&part| type_parts(part, checker))
            .collect();
    }
    let value = checker.get_type_at_location(node);
    let values = if value == checker.get_boolean_type() || !value.is_union() {
        vec![value]
    } else {
        value.types().to_vec()
    };
    values
        .into_iter()
        .map(|value| Part {
            flags: value.flags(),
            node: None,
            value: Some(value),
        })
        .collect()
}

fn primitive(flags: TypeFlags) -> Option<usize> {
    match flags {
        TypeFlags::BigInt => Some(0),
        TypeFlags::Boolean => Some(1),
        TypeFlags::Number => Some(2),
        TypeFlags::String => Some(3),
        _ => None,
    }
}
fn literal(flags: TypeFlags) -> Option<usize> {
    match flags {
        TypeFlags::BigIntLiteral => Some(0),
        TypeFlags::BooleanLiteral => Some(1),
        TypeFlags::NumberLiteral => Some(2),
        TypeFlags::StringLiteral | TypeFlags::TemplateLiteral => Some(3),
        _ => None,
    }
}
const PRIMITIVES: [&str; 4] = ["bigint", "boolean", "number", "string"];
fn join_parts(parts: &[Part], checker: &mut Checker) -> String {
    parts
        .iter()
        .map(|part| part.display(checker))
        .collect::<Vec<_>>()
        .join(" | ")
}
fn top_bottom(
    part: Part,
    node: P<Node>,
    intersection: bool,
    return_type: bool,
    checker: &mut Checker,
) -> Option<Violation> {
    let container = if intersection {
        "intersection"
    } else {
        "union"
    };
    let (id, message) = match part.flags {
        TypeFlags::Any => {
            if let Some(alias) = part
                .value
                .filter(|&value| is_error(value))
                .and_then(|value| value.alias())
                && let Some(symbol) = alias.symbol()
            {
                return Some(Violation::new(
                    node,
                    "errorTypeOverrides",
                    format!(
                        "'{}' is an 'error' type that acts as 'any' and overrides all other types in this {container} type.",
                        symbol.name()
                    ),
                ));
            }
            (
                "overrides",
                format!("'any' overrides all other types in this {container} type."),
            )
        }
        TypeFlags::Unknown if intersection => (
            "overridden",
            format!(
                "'{}' is overridden by other types in this intersection type.",
                part.display(checker)
            ),
        ),
        TypeFlags::Unknown => (
            "overrides",
            format!(
                "'{}' overrides all other types in this union type.",
                part.display(checker)
            ),
        ),
        TypeFlags::Never if intersection => (
            "overrides",
            format!(
                "'{}' overrides all other types in this intersection type.",
                part.display(checker)
            ),
        ),
        TypeFlags::Never if !return_type => (
            "overridden",
            "'never' is overridden by other types in this union type.".into(),
        ),
        _ => return None,
    };
    Some(Violation::new(node, id, message))
}

fn redundant_union(node: P<Node>, checker: &mut Checker) -> Vec<Violation> {
    let mut reports = Vec::new();
    let mut seen = [false; 4];
    let mut literals: [Vec<(P<Node>, Vec<Part>)>; 4] = std::array::from_fn(|_| Vec::new());
    let return_type = node.parent().is_some_and(is_function_like);
    for &node in constituents(node) {
        let node = skip_type_parentheses(node);
        for part in type_parts(node, checker) {
            if let Some(report) = top_bottom(part, node, false, return_type, checker) {
                reports.push(report);
                continue;
            }
            if let Some(index) = literal(part.flags) {
                if let Some((_, parts)) = literals[index]
                    .iter_mut()
                    .find(|(previous, _)| *previous == node)
                {
                    parts.push(part);
                } else {
                    literals[index].push((node, vec![part]));
                }
            }
            // Unlike the intersection rule's exact switch, the original union
            // rule ORs these bits. Canonical boolean has Boolean | Union flags.
            for (index, flag) in [
                TypeFlags::BigInt,
                TypeFlags::Boolean,
                TypeFlags::Number,
                TypeFlags::String,
            ]
            .into_iter()
            .enumerate()
            {
                seen[index] |= part.flags.intersects(flag);
            }
        }
    }
    for index in 0..4 {
        if seen[index] {
            for (node, parts) in &literals[index] {
                reports.push(Violation::new(
                    *node,
                    "literalOverridden",
                    format!(
                        "{} is overridden by {} in this union type.",
                        join_parts(parts, checker),
                        PRIMITIVES[index]
                    ),
                ));
            }
        }
    }
    reports
}

fn redundant_intersection(node: P<Node>, checker: &mut Checker) -> Vec<Violation> {
    let mut reports = Vec::new();
    let mut primitive_nodes: [Vec<P<Node>>; 4] = std::array::from_fn(|_| Vec::new());
    let mut literals: [Vec<Part>; 4] = std::array::from_fn(|_| Vec::new());
    let mut unions = Vec::new();
    for &node in constituents(node) {
        let node = skip_type_parentheses(node);
        let parts = type_parts(node, checker);
        if parts.len() >= 2 {
            unions.push((node, parts.clone()));
        }
        for part in parts {
            if let Some(report) = top_bottom(part, node, true, false, checker) {
                reports.push(report);
                continue;
            }
            if unions.is_empty()
                && let Some(index) = literal(part.flags)
            {
                literals[index].push(part);
            }
            if let Some(index) = primitive(part.flags) {
                primitive_nodes[index].push(node);
            }
        }
    }
    if !unions.is_empty() {
        // Preserve the frozen rule's message orientation and its reused first
        // union-literal text. These are observed compatibility behavior rather
        // than an attempt to redesign the original algorithm.
        let mut values = String::new();
        for (node, parts) in unions {
            let mut last = None;
            for part in &parts {
                last = literal(part.flags).filter(|&index| !primitive_nodes[index].is_empty());
                if last.is_none() {
                    break;
                }
            }
            let Some(index) = last else {
                continue;
            };
            if values.is_empty() {
                values = join_parts(&parts, checker);
            }
            reports.push(Violation::new(
                node,
                "primitiveOverridden",
                format!(
                    "{values} is overridden by the {} in this intersection type.",
                    PRIMITIVES[index]
                ),
            ));
        }
        return reports;
    }
    for index in 0..4 {
        if literals[index].is_empty() {
            continue;
        }
        let values = join_parts(&literals[index], checker);
        for &node in &primitive_nodes[index] {
            reports.push(Violation::new(
                node,
                "primitiveOverridden",
                format!(
                    "{} is overridden by the {values} in this intersection type.",
                    PRIMITIVES[index]
                ),
            ));
        }
    }
    reports
}
