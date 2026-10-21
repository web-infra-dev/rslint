pub(crate) mod control_flow;
pub(crate) mod syntax;
pub(crate) mod typed;

use crate::Violation;
use tsrs_ast::{Kind, Node, SourceFile};
use tsrs_checker::{Checker, TypeFlags};
use tsrs_core::P;

const UNSAFE_MINUS: &str = "@typescript-eslint/no-unsafe-unary-minus";

pub(crate) fn supported(name: &str) -> bool {
    name == UNSAFE_MINUS
        || syntax::NAMES.contains(&name)
        || control_flow::NAMES.contains(&name)
        || typed::NAMES.contains(&name)
}

pub(crate) fn check(
    name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
) -> Vec<Violation> {
    if control_flow::NAMES.contains(&name) {
        return control_flow::check(name, node, file);
    }
    if typed::NAMES.contains(&name) {
        return typed::check(name, node, checker);
    }
    if name != UNSAFE_MINUS {
        return syntax::check(name, node, file);
    }
    if node.kind() != Kind::PrefixUnaryExpression {
        return Vec::new();
    }
    let expression = node.as_prefix_unary_expression();
    if expression.operator() != Kind::MinusToken {
        return Vec::new();
    }
    let value = checker.get_type_at_location(expression.operand());
    let value = checker.get_base_constraint_of_type(value).unwrap_or(value);
    let parts = if value.flags().intersects(TypeFlags::Union) {
        value.types().to_vec()
    } else {
        vec![value]
    };
    for part in parts {
        if !part.flags().intersects(
            TypeFlags::Any | TypeFlags::Never | TypeFlags::BigIntLike | TypeFlags::NumberLike,
        ) {
            return vec![Violation {
                node,
                message_id: "unaryMinus",
                message: format!(
                    "Argument of unary negation should be assignable to number | bigint but is {} instead.",
                    checker.type_to_string_exported(part),
                ),
            }];
        }
    }
    Vec::new()
}
