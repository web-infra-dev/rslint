//! Preserve the two optionless TypeScript rules' fix/suggestion distinction.

use crate::{Violation, edits::Edit};
use tsrs_ast::{
    Kind, Node, is_optional_chain, is_optional_chain_root, is_outermost_optional_chain,
    skip_parentheses,
};
use tsrs_core::P;

pub(crate) const NAMES: &[&str] = &[
    "@typescript-eslint/no-extra-non-null-assertion",
    "@typescript-eslint/no-non-null-asserted-optional-chain",
];

pub(crate) fn check(rule: &str, node: P<Node>) -> Vec<Violation> {
    if node.kind() != Kind::NonNullExpression {
        return Vec::new();
    }
    let mut wrapped = node;
    while let Some(parent) = wrapped.parent() {
        if parent.kind() != Kind::ParenthesizedExpression
            || parent.as_parenthesized_expression().expression() != wrapped
        {
            break;
        }
        wrapped = parent;
    }
    let end = node
        .end()
        .expect_source("parsed assertion has source position") as usize;
    // A parsed non-null expression ends immediately after its own `!`.
    // Remove just that token, keeping comments and surrounding parentheses.
    let edit = Edit::remove([end - 1, end]);
    if rule == NAMES[0] {
        let Some(parent) = wrapped.parent() else {
            return Vec::new();
        };
        if parent.kind() == Kind::NonNullExpression
            || (matches!(
                parent.kind(),
                Kind::PropertyAccessExpression
                    | Kind::ElementAccessExpression
                    | Kind::CallExpression
            ) && is_optional_chain_root(parent)
                && parent.expression() == Some(wrapped))
        {
            return vec![
                Violation::new(
                    node,
                    "noExtraNonNullAssertion",
                    "Forbidden extra non-null assertion.",
                )
                .with_fix(vec![edit]),
            ];
        }
    } else if (is_optional_chain(node) && is_outermost_optional_chain(node))
        || (!is_optional_chain(node)
            && is_optional_chain(skip_parentheses(node.as_non_null_expression().expression())))
    {
        return vec![Violation::new(node, "noNonNullOptionalChain", "Optional chain expressions can return undefined by design - using a non-null assertion is unsafe and wrong.")
            .with_suggestion("suggestRemovingNonNull", "You should remove the non-null assertion.", vec![edit])];
    }
    Vec::new()
}
