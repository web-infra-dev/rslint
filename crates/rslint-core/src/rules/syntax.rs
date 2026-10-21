//! Syntax-only rules migrated from the matching rslint rule implementations.
//!
//! These rules have no options or fixes. Keep TypeScript assertion nodes opaque:
//! only parentheses disappear in the equivalent ESTree expressions.

use tsrs_ast::{Kind, Node, SourceFile, skip_parentheses};
use tsrs_core::P;

use crate::Violation;

pub(crate) const NAMES: &[&str] = &[
    "no-caller",
    "no-compare-neg-zero",
    "no-debugger",
    "no-nested-ternary",
    "no-template-curly-in-string",
    "no-ternary",
    "no-with",
];

pub(crate) fn check(rule: &str, node: P<Node>, _file: P<SourceFile>) -> Vec<Violation> {
    let report = match (rule, node.kind()) {
        ("no-debugger", Kind::DebuggerStatement) => Some(violation(
            node,
            "no-debugger",
            "Unexpected 'debugger' statement.",
        )),
        ("no-with", Kind::WithStatement) => Some(violation(
            node,
            "unexpectedWith",
            "Unexpected use of 'with' statement.",
        )),
        ("no-ternary", Kind::ConditionalExpression) => Some(violation(
            node,
            "noTernaryOperator",
            "Ternary operator used.",
        )),
        ("no-nested-ternary", Kind::ConditionalExpression) => {
            let conditional = node.as_conditional_expression();
            (skip_parentheses(conditional.when_true()).kind() == Kind::ConditionalExpression
                || skip_parentheses(conditional.when_false()).kind() == Kind::ConditionalExpression)
                .then(|| violation(node, "noNestedTernary", "Do not nest ternary expressions."))
        }
        ("no-template-curly-in-string", Kind::StringLiteral) => {
            has_template_placeholder(node.text()).then(|| {
                violation(
                    node,
                    "unexpectedTemplateExpression",
                    "Template literal placeholder syntax in regular strings is not allowed.",
                )
            })
        }
        ("no-caller", Kind::PropertyAccessExpression | Kind::QualifiedName) => check_caller(node),
        ("no-compare-neg-zero", Kind::PrefixUnaryExpression) => check_negative_zero(node),
        _ => None,
    };
    report.into_iter().collect()
}

fn violation(node: P<Node>, message_id: &'static str, message: &str) -> Violation {
    Violation::new(node, message_id, message)
}

fn check_caller(node: P<Node>) -> Option<Violation> {
    let (object, property) = match node.kind() {
        Kind::PropertyAccessExpression => {
            let access = node.as_property_access_expression();
            (access.expression(), access.name())
        }
        Kind::QualifiedName if is_heritage_qualified_name(node) => {
            let qualified = node.as_qualified_name();
            (qualified.left(), qualified.right())
        }
        _ => return None,
    };
    let object = skip_parentheses(object);
    if object.kind() != Kind::Identifier || object.text() != "arguments" {
        return None;
    }
    match property.text() {
        "callee" => Some(violation(node, "unexpected", "Avoid arguments.callee.")),
        "caller" => Some(violation(node, "unexpected", "Avoid arguments.caller.")),
        _ => None,
    }
}

/// Dotted names are member expressions only in the heritage target itself.
/// Names inside ordinary type references, queries, or type arguments stay types.
fn is_heritage_qualified_name(node: P<Node>) -> bool {
    let mut entity = node;
    while let Some(parent) = entity.parent() {
        if parent.kind() != Kind::QualifiedName || parent.as_qualified_name().left() != entity {
            break;
        }
        entity = parent;
    }
    let Some(reference) = entity.parent() else {
        return false;
    };
    reference.kind() == Kind::TypeReference
        && reference.as_type_reference_node().type_name() == entity
        && reference
            .parent()
            .is_some_and(|parent| parent.kind() == Kind::HeritageClause)
}

/// Equivalent to the existing rule's `\$\{[^}]+\}` over decoded string text.
fn has_template_placeholder(text: &str) -> bool {
    text.match_indices("${")
        .any(|(start, _)| text[start + 2..].find('}').is_some_and(|end| end > 0))
}

fn zero_numeric_literal(node: P<Node>) -> bool {
    let node = skip_parentheses(node);
    // tsrs normalizes all zero numeric spellings, including 0x0 and 0e10.
    node.kind() == Kind::NumericLiteral && node.text() == "0"
}

fn negative_zero(node: P<Node>) -> bool {
    let node = skip_parentheses(node);
    node.kind() == Kind::PrefixUnaryExpression
        && node.as_prefix_unary_expression().operator() == Kind::MinusToken
        && zero_numeric_literal(node.as_prefix_unary_expression().operand())
}

fn check_negative_zero(node: P<Node>) -> Option<Violation> {
    let prefix = node.as_prefix_unary_expression();
    if prefix.operator() != Kind::MinusToken || !zero_numeric_literal(prefix.operand()) {
        return None;
    }
    let mut operand = node;
    while let Some(parent) = operand.parent() {
        if parent.kind() != Kind::ParenthesizedExpression
            || parent.as_parenthesized_expression().expression() != operand
        {
            break;
        }
        operand = parent;
    }
    let comparison_node = operand.parent()?;
    if comparison_node.kind() != Kind::BinaryExpression {
        return None;
    }
    let comparison = comparison_node.as_binary_expression();
    if comparison.left() != operand && comparison.right() != operand {
        return None;
    }
    let operator = match comparison.operator_token().kind() {
        Kind::GreaterThanToken => ">",
        Kind::GreaterThanEqualsToken => ">=",
        Kind::LessThanToken => "<",
        Kind::LessThanEqualsToken => "<=",
        Kind::EqualsEqualsToken => "==",
        Kind::EqualsEqualsEqualsToken => "===",
        Kind::ExclamationEqualsToken => "!=",
        Kind::ExclamationEqualsEqualsToken => "!==",
        _ => return None,
    };
    // The left negative-zero listener owns a comparison with two -0 operands.
    if comparison.right() == operand && negative_zero(comparison.left()) {
        return None;
    }
    Some(violation(
        comparison_node,
        "unexpected",
        &format!("Do not use the '{operator}' operator to compare against -0."),
    ))
}
