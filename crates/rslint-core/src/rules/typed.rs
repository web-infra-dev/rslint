//! Type-aware rules ported from the frozen rslint implementations.
//!
//! These rules have no configurable options, fixes, or suggestions. All type
//! identities and assignability queries belong to the current tsrs checker.

use std::collections::BTreeMap;

use tsrs_ast::{Kind, Node, SymbolFlags, get_source_file_of_node, is_literal_expression};
use tsrs_checker::{Checker, Type, TypeFlags};
use tsrs_core::P;

use crate::Violation;

pub(crate) const NAMES: &[&str] = &[
    "@typescript-eslint/no-mixed-enums",
    "@typescript-eslint/no-unsafe-enum-comparison",
    "@typescript-eslint/related-getter-setter-pairs",
];

pub(crate) fn check(rule: &str, node: P<Node>, checker: &mut Checker) -> Vec<Violation> {
    match (rule, node.kind()) {
        ("@typescript-eslint/no-mixed-enums", Kind::EnumDeclaration) => {
            check_mixed_enum(node, checker).into_iter().collect()
        }
        ("@typescript-eslint/no-unsafe-enum-comparison", Kind::BinaryExpression) => {
            let expression = node.as_binary_expression();
            if !matches!(
                expression.operator_token().kind(),
                Kind::LessThanToken
                    | Kind::LessThanEqualsToken
                    | Kind::GreaterThanToken
                    | Kind::GreaterThanEqualsToken
                    | Kind::EqualsEqualsToken
                    | Kind::EqualsEqualsEqualsToken
                    | Kind::ExclamationEqualsToken
                    | Kind::ExclamationEqualsEqualsToken
            ) {
                return Vec::new();
            }
            let left = checker.get_type_at_location(expression.left());
            let right = checker.get_type_at_location(expression.right());
            mismatched_enums(left, right, checker)
                .then(|| {
                    violation(
                        node,
                        "mismatchedCondition",
                        "The two values in this comparison do not have a shared enum type.",
                    )
                })
                .into_iter()
                .collect()
        }
        ("@typescript-eslint/no-unsafe-enum-comparison", Kind::CaseClause) => {
            let switch = node
                .parent()
                .and_then(|block| block.parent())
                .expect("case switch");
            let left =
                checker.get_type_at_location(switch.expression().expect("switch expression"));
            let right = checker.get_type_at_location(node.expression().expect("case expression"));
            mismatched_enums(left, right, checker)
                .then(|| {
                    violation(
                        node,
                        "mismatchedCase",
                        "The case statement does not have a shared enum type with the switch predicate.",
                    )
                })
                .into_iter()
                .collect()
        }
        (
            "@typescript-eslint/related-getter-setter-pairs",
            Kind::ClassDeclaration
            | Kind::ClassExpression
            | Kind::InterfaceDeclaration
            | Kind::TypeLiteral,
        ) => check_accessors(node, checker),
        _ => Vec::new(),
    }
}

fn violation(node: P<Node>, message_id: &'static str, message: &str) -> Violation {
    Violation {
        node,
        message_id,
        message: message.to_owned(),
    }
}

#[derive(Clone, Copy, PartialEq, Eq)]
enum EnumValueKind {
    Number,
    String,
    Unknown,
}

fn enum_member_kind(member: P<Node>, checker: &mut Checker) -> EnumValueKind {
    let Some(initializer) = member.as_enum_member().initializer() else {
        return EnumValueKind::Number;
    };
    match initializer.kind() {
        Kind::NumericLiteral => EnumValueKind::Number,
        Kind::StringLiteral => EnumValueKind::String,
        _ => {
            let value = checker.get_type_at_location(initializer);
            if value.flags().intersects(TypeFlags::StringLike) {
                EnumValueKind::String
            } else if value.flags().intersects(TypeFlags::NumberLike) {
                EnumValueKind::Number
            } else {
                EnumValueKind::Unknown
            }
        }
    }
}

fn check_mixed_enum(node: P<Node>, checker: &mut Checker) -> Option<Violation> {
    if node.members().is_empty() {
        return None;
    }
    // Declaration merging uses the first declaration's first member, including
    // declarations outside the currently admitted lint file.
    let symbol = checker.get_symbol_at_location_exported(node.as_enum_declaration().name())?;
    let declaration = *symbol.declarations().first()?;
    let first = *declaration.members().first()?;
    let desired = enum_member_kind(first, checker);
    if desired == EnumValueKind::Unknown {
        return None;
    }
    for &member in node.members() {
        let actual = enum_member_kind(member, checker);
        if actual == EnumValueKind::Unknown {
            return None;
        }
        if actual != desired {
            return Some(violation(
                member.as_enum_member().initializer().unwrap_or(member),
                "mixed",
                "Mixing number and string enums can be confusing.",
            ));
        }
    }
    None
}

fn parts(value: P<Type>, flag: TypeFlags) -> Vec<P<Type>> {
    if value.flags().intersects(flag) {
        value.types().to_vec()
    } else {
        vec![value]
    }
}

fn enum_types(value: P<Type>, checker: &mut Checker) -> Vec<P<Type>> {
    parts(value, TypeFlags::Union)
        .into_iter()
        .filter(|part| part.flags().intersects(TypeFlags::EnumLiteral))
        .map(|part| {
            let Some(symbol) = part.symbol() else {
                return part;
            };
            if !symbol.flags().intersects(SymbolFlags::EnumMember) {
                return part;
            }
            let declaration = symbol.value_declaration().expect("enum member declaration");
            checker.get_type_at_location(declaration.parent().expect("enum declaration"))
        })
        .collect()
}

fn is_type_like(value: P<Type>, flags: TypeFlags) -> bool {
    parts(value, TypeFlags::Union)
        .into_iter()
        .all(|union_part| {
            parts(union_part, TypeFlags::Intersection)
                .into_iter()
                .any(|intersection_part| intersection_part.flags().intersects(flags))
        })
}

fn type_violates(left: &[P<Type>], right: P<Type>) -> bool {
    let mut number_enum = false;
    let mut string_enum = false;
    for part in left {
        if part.flags().intersects(TypeFlags::EnumLike) {
            if part.flags().intersects(TypeFlags::NumberLiteral) {
                number_enum = true;
            } else {
                string_enum = true;
            }
        }
    }
    (number_enum && is_type_like(right, TypeFlags::NumberLike))
        || (string_enum && is_type_like(right, TypeFlags::StringLike))
}

fn mismatched_enums(left: P<Type>, right: P<Type>, checker: &mut Checker) -> bool {
    let left_enums = enum_types(left, checker);
    let right_enums = enum_types(right, checker);
    if left_enums.is_empty() && right_enums.is_empty() {
        return false;
    }
    if left_enums.iter().any(|value| right_enums.contains(value)) {
        return false;
    }
    let left_parts = parts(left, TypeFlags::Union);
    let right_parts = parts(right, TypeFlags::Union);
    if left_parts.iter().any(|value| right_parts.contains(value)) {
        return false;
    }
    type_violates(&left_parts, right) || type_violates(&right_parts, left)
}

// This preserves GetNameFromMember, including its source-text distinction for
// quoted names and its identifier-validity handling for computed literals.
fn member_name(node: P<Node>) -> String {
    match node.kind() {
        Kind::Identifier | Kind::PrivateIdentifier => return node.text().to_owned(),
        Kind::ComputedPropertyName => {
            let expression = node.as_computed_property_name().expression();
            if is_literal_expression(expression) {
                let text = expression.text();
                return if tsrs_scanner::is_valid_identifier(text) {
                    text.to_owned()
                } else {
                    format!("\"{text}\"")
                };
            }
        }
        _ => {}
    }
    let file = get_source_file_of_node(node).expect("accessor source file");
    let start = tsrs_scanner::skip_trivia(
        file.text(),
        node.pos().source_offset().expect("source position"),
    );
    let end = node.end().source_offset().expect("source position");
    file.text()[start as usize..end as usize].to_owned()
}

fn check_accessor_pair(
    getter: P<Node>,
    setter: P<Node>,
    checker: &mut Checker,
) -> Option<Violation> {
    let get_type = checker.get_type_at_location(getter);
    let set_type = checker.get_type_at_location(setter.parameters()[0]);
    (!checker.is_type_assignable_to(get_type, set_type)).then(|| {
        violation(
            getter.type_node().expect("explicit getter type"),
            "mismatch",
            "`get()` type should be assignable to its equivalent `set()` type.",
        )
    })
}

fn check_accessors(node: P<Node>, checker: &mut Checker) -> Vec<Violation> {
    let mut getters = BTreeMap::new();
    let mut setters = BTreeMap::new();
    let mut reports = Vec::new();
    for &member in node.members() {
        let report = match member.kind() {
            Kind::GetAccessor if member.type_node().is_some() => {
                let name = member_name(member.as_get_accessor_declaration().name());
                if let Some(&setter) = setters.get(&name) {
                    check_accessor_pair(member, setter, checker)
                } else {
                    getters.insert(name, member);
                    None
                }
            }
            Kind::SetAccessor if member.parameters().len() == 1 => {
                let name = member_name(member.as_set_accessor_declaration().name());
                if let Some(&getter) = getters.get(&name) {
                    check_accessor_pair(getter, member, checker)
                } else {
                    setters.insert(name, member);
                    None
                }
            }
            _ => None,
        };
        reports.extend(report);
    }
    reports
}
