//! Optionless structural rules from rslint e0a29f751d4def86d27931759098f27748fa3003.
//! Preserve the baseline's binding-pattern and parentheses boundaries; none of
//! these rules offers a fix or suggestion.

use std::collections::HashSet;

use tsrs_ast::{Kind, Node, SourceFile, is_assignment_target, skip_parentheses};
use tsrs_core::P;

use crate::Violation;

pub(crate) const NAMES: &[&str] = &["no-delete-var", "no-dupe-args", "no-sparse-arrays"];

pub(crate) fn check(rule: &str, node: P<Node>, file: P<SourceFile>) -> Vec<Violation> {
    match (rule, node.kind()) {
        ("no-delete-var", Kind::DeleteExpression)
            if skip_parentheses(node.as_delete_expression().expression()).kind()
                == Kind::Identifier =>
        {
            vec![Violation::new(
                node,
                "unexpected",
                "Variables should not be deleted.",
            )]
        }
        ("no-dupe-args", Kind::FunctionDeclaration | Kind::FunctionExpression) => {
            duplicate_arguments(node)
        }
        ("no-sparse-arrays", Kind::ArrayLiteralExpression) => sparse_array(node, file),
        _ => Vec::new(),
    }
}

fn sparse_array(node: P<Node>, file: P<SourceFile>) -> Vec<Violation> {
    let mut checked_assignment_target = false;
    let mut reports = Vec::new();
    for &element in node.as_array_literal_expression().elements().nodes() {
        if element.kind() != Kind::OmittedExpression {
            continue;
        }
        if !checked_assignment_target {
            // The compiler represents destructuring assignment targets as array
            // literals too; their elisions skip values instead of making holes.
            if is_assignment_target(node) {
                return Vec::new();
            }
            checked_assignment_target = true;
        }
        let position = element
            .pos()
            .expect_source("parsed omitted expressions have source positions");
        let range = tsrs_scanner::get_range_of_token_at_position(file, position);
        let start = range
            .pos()
            .expect_source("scanner ranges have source positions") as usize;
        let end = range
            .end()
            .expect_source("scanner ranges have source positions") as usize;
        reports.push(
            Violation::new(
                element,
                "unexpectedSparseArray",
                "Unexpected comma in middle of array.",
            )
            .with_byte_range([start, end]),
        );
    }
    reports
}

fn duplicate_arguments(node: P<Node>) -> Vec<Violation> {
    let mut seen = HashSet::new();
    let mut reports = Vec::new();
    for &parameter in node.parameters() {
        if parameter.kind() != Kind::Parameter {
            continue;
        }
        collect_binding_names(parameter.as_parameter_declaration().name(), &mut |name| {
            let text = name.text();
            if !seen.insert(text) {
                // rslint reports each duplicate identifier, rather than one
                // report on the function for every duplicated name.
                reports.push(Violation::new(
                    name,
                    "unexpected",
                    format!("Duplicate param '{text}'."),
                ));
            }
        });
    }
    reports
}

fn collect_binding_names(node: P<Node>, callback: &mut impl FnMut(P<Node>)) {
    match node.kind() {
        Kind::Identifier if !node.text().is_empty() => callback(node),
        Kind::ObjectBindingPattern | Kind::ArrayBindingPattern => {
            node.for_each_child(&mut |child| {
                if child.kind() == Kind::BindingElement
                    && let Some(name) = child.as_binding_element().name()
                {
                    collect_binding_names(name, callback);
                }
                false
            });
        }
        _ => {}
    }
}
