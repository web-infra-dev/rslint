//! Optionless control-flow rules ported from the frozen rslint baseline.
//! Parent walks deliberately retain each rule's own boundary semantics.

use tsrs_ast::{
    Kind, Node, SourceFile, is_class_like, is_function_like_declaration, is_iteration_statement,
    skip_parentheses,
};
use tsrs_core::P;

use crate::Violation;

pub(crate) const NAMES: &[&str] = &[
    "no-await-in-loop",
    "no-constructor-return",
    "no-continue",
    "no-negated-condition",
    "no-unsafe-finally",
    "no-useless-catch",
];

pub(crate) fn check(rule: &str, node: P<Node>, _file: P<SourceFile>) -> Vec<Violation> {
    let report = match (rule, node.kind()) {
        ("no-continue", Kind::ContinueStatement) => Some(violation(
            node,
            "unexpected",
            "Unexpected use of continue statement.",
        )),
        ("no-await-in-loop", Kind::AwaitExpression) if in_loop(node) => Some(violation(
            node,
            "unexpectedAwait",
            "Unexpected `await` inside a loop.",
        )),
        ("no-await-in-loop", Kind::ForOfStatement) if for_await(node) && in_loop(node) => node
            .as_for_in_or_of_statement()
            .await_modifier()
            .map(|modifier| {
                violation(
                    modifier,
                    "unexpectedAwait",
                    "Unexpected `await` inside a loop.",
                )
            }),
        ("no-constructor-return", Kind::ReturnStatement)
            if node.as_return_statement().expression().is_some() && in_constructor(node) =>
        {
            Some(violation(
                node,
                "unexpected",
                "Unexpected return statement in constructor.",
            ))
        }
        ("no-negated-condition", Kind::IfStatement) => {
            let statement = node.as_if_statement();
            (statement
                .else_statement()
                .is_some_and(|branch| branch.kind() != Kind::IfStatement)
                && negated_test(statement.expression()))
            .then(|| violation(node, "unexpectedNegated", "Unexpected negated condition."))
        }
        ("no-negated-condition", Kind::ConditionalExpression)
            if negated_test(node.as_conditional_expression().condition()) =>
        {
            Some(violation(
                node,
                "unexpectedNegated",
                "Unexpected negated condition.",
            ))
        }
        ("no-unsafe-finally", _) => unsafe_finally(node),
        ("no-useless-catch", Kind::CatchClause) => useless_catch(node),
        _ => None,
    };
    report.into_iter().collect()
}

fn violation(node: P<Node>, message_id: &'static str, message: &str) -> Violation {
    Violation {
        node,
        message_id,
        message: message.to_owned(),
    }
}

fn for_await(node: P<Node>) -> bool {
    node.kind() == Kind::ForOfStatement
        && node.as_for_in_or_of_statement().await_modifier().is_some()
}

fn in_loop(node: P<Node>) -> bool {
    let mut ancestor = node.parent();
    while let Some(current) = ancestor {
        if for_await(current) {
            return false;
        }
        if is_iteration_statement(current, false) {
            let initializer = match current.kind() {
                Kind::ForInStatement | Kind::ForOfStatement => {
                    Some(current.as_for_in_or_of_statement().expression())
                }
                Kind::ForStatement => current.as_for_statement().initializer(),
                _ => None,
            };
            return !initializer.is_some_and(|root| in_initializer_subtree(node, root));
        }
        if is_function_like_declaration(current) {
            return false;
        }
        ancestor = current.parent();
    }
    false
}

fn in_initializer_subtree(node: P<Node>, root: P<Node>) -> bool {
    let mut ancestor = Some(node);
    while let Some(current) = ancestor {
        if current == root {
            return true;
        }
        if is_iteration_statement(current, false) || is_function_like_declaration(current) {
            return false;
        }
        ancestor = current.parent();
    }
    false
}

fn in_constructor(node: P<Node>) -> bool {
    let mut ancestor = node.parent();
    while let Some(current) = ancestor {
        match current.kind() {
            Kind::Constructor => return true,
            // Match the baseline rule's explicit boundary set. Do not broaden
            // this to all function-like nodes during a backend migration.
            Kind::FunctionDeclaration
            | Kind::FunctionExpression
            | Kind::ArrowFunction
            | Kind::MethodDeclaration => return false,
            _ => ancestor = current.parent(),
        }
    }
    false
}

fn negated_test(node: P<Node>) -> bool {
    let node = skip_parentheses(node);
    match node.kind() {
        Kind::PrefixUnaryExpression => {
            node.as_prefix_unary_expression().operator() == Kind::ExclamationToken
        }
        Kind::BinaryExpression => matches!(
            node.as_binary_expression().operator_token().kind(),
            Kind::ExclamationEqualsToken | Kind::ExclamationEqualsEqualsToken
        ),
        _ => false,
    }
}

fn unsafe_finally(node: P<Node>) -> Option<Violation> {
    let (kind, label, stop_at_loop, stop_at_switch) = match node.kind() {
        Kind::ReturnStatement => ("ReturnStatement", None, false, false),
        Kind::ThrowStatement => ("ThrowStatement", None, false, false),
        Kind::BreakStatement => {
            let label = node.as_break_statement().label();
            ("BreakStatement", label, label.is_none(), label.is_none())
        }
        Kind::ContinueStatement => (
            "ContinueStatement",
            node.as_continue_statement().label(),
            true,
            false,
        ),
        _ => return None,
    };
    let mut ancestor = node.parent();
    while let Some(current) = ancestor {
        if is_function_like_declaration(current)
            || is_class_like(current)
            || (stop_at_loop && is_iteration_statement(current, false))
            || (stop_at_switch && current.kind() == Kind::SwitchStatement)
        {
            return None;
        }
        if current.kind() == Kind::LabeledStatement
            && label
                .is_some_and(|label| current.as_labeled_statement().label().text() == label.text())
        {
            return None;
        }
        if let Some(parent) = current.parent()
            && parent.kind() == Kind::TryStatement
            && parent.as_try_statement().finally_block() == Some(current)
        {
            return Some(violation(
                node,
                "unsafeUsage",
                &format!("Unsafe usage of {kind}."),
            ));
        }
        ancestor = current.parent();
    }
    None
}

fn useless_catch(node: P<Node>) -> Option<Violation> {
    let catch = node.as_catch_clause();
    let declaration = catch.variable_declaration()?;
    let parameter = declaration.as_variable_declaration().name();
    if parameter.kind() != Kind::Identifier {
        return None;
    }
    let block = catch.block().as_block();
    let statement = *block.statements().nodes().first()?;
    if statement.kind() != Kind::ThrowStatement {
        return None;
    }
    let thrown = skip_parentheses(statement.as_throw_statement().expression());
    if thrown.kind() != Kind::Identifier || thrown.text() != parameter.text() {
        return None;
    }
    let parent = node.parent()?;
    if parent.kind() != Kind::TryStatement {
        return None;
    }
    if parent.as_try_statement().finally_block().is_some() {
        Some(violation(
            node,
            "unnecessaryCatchClause",
            "Unnecessary catch clause.",
        ))
    } else {
        Some(violation(
            parent,
            "unnecessaryCatch",
            "Unnecessary try/catch wrapper.",
        ))
    }
}
