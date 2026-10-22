//! Default options of two syntax rules in the frozen Rsbuild configuration.
//! Ported from rslint e0a29f751d4def86d27931759098f27748fa3003 (MIT).
use crate::{Violation, edits::Edit};
use tsrs_ast::{self as ast, Kind, Node, SourceFile};
use tsrs_core::P;

pub(crate) const NAMES: &[&str] = &["no-unused-labels", "no-unsafe-optional-chaining"];

pub(crate) fn check(name: &str, node: P<Node>, file: P<SourceFile>) -> Vec<Violation> {
    match name {
        "no-unused-labels" if node.kind() == Kind::SourceFile => unused_labels(node, file),
        "no-unsafe-optional-chaining" => unsafe_optional(node),
        _ => Vec::new(),
    }
}

fn start(node: P<Node>, file: P<SourceFile>) -> usize {
    tsrs_scanner::skip_trivia(file.text(), node.pos().expect_source("parsed safety node")) as usize
}

fn unused_labels(root: P<Node>, file: P<SourceFile>) -> Vec<Violation> {
    struct Label {
        name: String,
        used: bool,
    }
    fn visit(
        node: P<Node>,
        file: P<SourceFile>,
        labels: &mut Vec<Label>,
        out: &mut Vec<Violation>,
    ) {
        if node.kind() == Kind::LabeledStatement {
            labels.push(Label {
                name: node.as_labeled_statement().label.text().into(),
                used: false,
            });
        }
        let reference = match node.kind() {
            Kind::BreakStatement => node.as_break_statement().label,
            Kind::ContinueStatement => node.as_continue_statement().label,
            _ => None,
        };
        if let Some(reference) = reference
            && let Some(label) = labels
                .iter_mut()
                .rev()
                .find(|label| label.name == reference.text())
        {
            label.used = true;
        }
        node.for_each_child(&mut |child| {
            visit(child, file, labels, out);
            false
        });
        if node.kind() != Kind::LabeledStatement {
            return;
        }
        let label = labels.pop().expect("label entered before descendants");
        if label.used {
            return;
        }
        let statement = node.as_labeled_statement();
        let mut issue = Violation::new(
            statement.label,
            "unused",
            format!("'{}:' is defined but never used.", label.name),
        );
        let label_end = statement.label.end().expect_source("label token") as usize;
        let body_start = start(statement.statement, file);
        let separator = &file.text()[label_end..body_start];
        // This span contains only the colon and trivia, never string tokens.
        let has_comment = separator.contains("//") || separator.contains("/*");
        let mut ancestor = node.parent();
        while ancestor.is_some_and(|p| p.kind() == Kind::LabeledStatement) {
            ancestor = ancestor.and_then(|p| p.parent());
        }
        let directive_container = ancestor.is_some_and(|p| {
            p.kind() == Kind::SourceFile
                || (p.kind() == Kind::Block && p.parent().is_some_and(ast::is_function_like))
        });
        let potential_directive = directive_container
            && statement.statement.kind() == Kind::ExpressionStatement
            && ast::is_string_literal_like(ast::skip_parentheses(
                statement.statement.as_expression_statement().expression,
            ));
        if !has_comment && !potential_directive {
            issue = issue.with_fix(vec![Edit::remove([start(node, file), body_start])]);
        }
        out.push(issue);
    }
    let mut out = Vec::new();
    visit(root, file, &mut Vec::new(), &mut out);
    out
}

fn undefined_short_circuit(node: P<Node>, out: &mut Vec<Violation>) {
    let node = ast::skip_parentheses(node);
    match node.kind() {
        Kind::PropertyAccessExpression | Kind::ElementAccessExpression | Kind::CallExpression
            if ast::is_optional_chain(node) => out.push(Violation::new(node, "unsafeOptionalChain",
                "Unsafe usage of optional chaining. If it short-circuits with 'undefined' the evaluation will throw TypeError.")),
        Kind::AwaitExpression => undefined_short_circuit(node.as_await_expression().expression, out),
        Kind::BinaryExpression => {
            let binary = node.as_binary_expression();
            match binary.operator_token.kind() {
                Kind::BarBarToken | Kind::QuestionQuestionToken | Kind::CommaToken => undefined_short_circuit(binary.right(), out),
                Kind::AmpersandAmpersandToken => {
                    undefined_short_circuit(binary.left(), out);
                    undefined_short_circuit(binary.right(), out);
                }
                _ => {},
            }
        }
        Kind::ConditionalExpression => {
            let conditional = node.as_conditional_expression();
            undefined_short_circuit(conditional.when_true, out);
            undefined_short_circuit(conditional.when_false, out);
        }
        _ => {},
    }
}

fn unsafe_optional(node: P<Node>) -> Vec<Violation> {
    let expression = match node.kind() {
        Kind::CallExpression if !ast::is_optional_chain(node) => {
            Some(node.as_call_expression().expression())
        }
        Kind::NewExpression => Some(node.as_new_expression().expression()),
        Kind::TaggedTemplateExpression => Some(node.as_tagged_template_expression().tag),
        Kind::PropertyAccessExpression if !ast::is_optional_chain(node) => {
            Some(node.as_property_access_expression().expression)
        }
        Kind::ElementAccessExpression if !ast::is_optional_chain(node) => {
            Some(node.as_element_access_expression().expression)
        }
        Kind::BinaryExpression
            if ast::is_destructuring_assignment(node)
                || matches!(
                    node.as_binary_expression().operator_token.kind(),
                    Kind::InKeyword | Kind::InstanceOfKeyword
                ) =>
        {
            Some(node.as_binary_expression().right())
        }
        Kind::SpreadElement
            if node
                .parent()
                .is_none_or(|p| p.kind() != Kind::ObjectLiteralExpression) =>
        {
            Some(node.as_spread_element().expression)
        }
        Kind::ForOfStatement => Some(node.as_for_in_or_of_statement().expression),
        Kind::VariableDeclaration if node.name().is_some_and(ast::is_binding_pattern) => {
            node.as_variable_declaration().initializer.get()
        }
        Kind::BindingElement if node.name().is_some_and(ast::is_binding_pattern) => {
            node.as_binding_element().initializer()
        }
        Kind::WithStatement => Some(node.as_with_statement().expression),
        Kind::ClassDeclaration | Kind::ClassExpression => {
            ast::get_class_extends_heritage_element(node)
                .map(|element| element.as_expression_with_type_arguments().expression)
        }
        _ => None,
    };
    let mut out = Vec::new();
    if let Some(expression) = expression {
        undefined_short_circuit(expression, &mut out);
    }
    out
}
