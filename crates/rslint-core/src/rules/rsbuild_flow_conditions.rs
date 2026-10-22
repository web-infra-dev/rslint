//! Default condition and switch rules from frozen rslint e0a29f751d4def86d27931759098f27748fa3003.
//! The two constant rules deliberately retain their distinct frozen predicates,
//! including global-binding checks; neither substitutes arbitrary type folding.
use crate::Violation;
use tsrs_ast::{self as ast, Kind, Node, SourceFile};
use tsrs_checker::Checker;
use tsrs_core::{P, jsnum};
use tsrs_scanner::skip_trivia;

pub(crate) const NAMES: &[&str] = &[
    "no-constant-binary-expression",
    "no-constant-condition",
    "no-dupe-else-if",
    "no-fallthrough",
];
pub(crate) fn check(
    name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    default_library: &dyn Fn(P<SourceFile>) -> bool,
    has_type_information: bool,
) -> Vec<Violation> {
    match name {
        "no-fallthrough" if node.kind() == Kind::SwitchStatement => fallthrough(node, file),
        "no-dupe-else-if" if node.kind() == Kind::IfStatement => {
            duplicate_else_if(node, file).into_iter().collect()
        }
        "no-constant-binary-expression" if node.kind() == Kind::BinaryExpression => Constants {
            file,
            checker,
            default_library,
            has_type_information,
            condition_rule: false,
        }
        .binary(node)
        .into_iter()
        .collect(),
        "no-constant-condition" => Constants {
            file,
            checker,
            default_library,
            has_type_information,
            condition_rule: true,
        }
        .condition(node)
        .into_iter()
        .collect(),
        _ => Vec::new(),
    }
}
fn children(node: P<Node>) -> Vec<P<Node>> {
    let mut children = Vec::new();
    node.for_each_child(&mut |child| {
        children.push(child);
        false
    });
    children
}
fn start(node: P<Node>, file: P<SourceFile>) -> usize {
    skip_trivia(file.text(), node.pos().expect_source("parsed node")) as usize
}
fn end(node: P<Node>) -> usize {
    node.end().expect_source("parsed node") as usize
}
fn numeric_operator(op: Kind) -> bool {
    matches!(
        op,
        Kind::PlusToken
            | Kind::MinusToken
            | Kind::AsteriskToken
            | Kind::SlashToken
            | Kind::PercentToken
            | Kind::AsteriskAsteriskToken
            | Kind::LessThanLessThanToken
            | Kind::GreaterThanGreaterThanToken
            | Kind::GreaterThanGreaterThanGreaterThanToken
            | Kind::BarToken
            | Kind::AmpersandToken
            | Kind::CaretToken
    )
}
fn logical_assignment(op: Kind) -> bool {
    matches!(
        op,
        Kind::BarBarEqualsToken
            | Kind::AmpersandAmpersandEqualsToken
            | Kind::QuestionQuestionEqualsToken
    )
}
fn compound_assignment(op: Kind) -> bool {
    matches!(
        op,
        Kind::PlusEqualsToken
            | Kind::MinusEqualsToken
            | Kind::AsteriskEqualsToken
            | Kind::SlashEqualsToken
            | Kind::PercentEqualsToken
            | Kind::AsteriskAsteriskEqualsToken
            | Kind::LessThanLessThanEqualsToken
            | Kind::GreaterThanGreaterThanEqualsToken
            | Kind::GreaterThanGreaterThanGreaterThanEqualsToken
            | Kind::BarEqualsToken
            | Kind::AmpersandEqualsToken
            | Kind::CaretEqualsToken
    )
}
fn comparison(op: Kind) -> bool {
    matches!(
        op,
        Kind::LessThanToken
            | Kind::LessThanEqualsToken
            | Kind::GreaterThanToken
            | Kind::GreaterThanEqualsToken
            | Kind::EqualsEqualsToken
            | Kind::ExclamationEqualsToken
            | Kind::EqualsEqualsEqualsToken
            | Kind::ExclamationEqualsEqualsToken
            | Kind::InstanceOfKeyword
    )
}
struct Constants<'a> {
    file: P<SourceFile>,
    checker: &'a mut Checker,
    default_library: &'a dyn Fn(P<SourceFile>) -> bool,
    has_type_information: bool,
    condition_rule: bool,
}
impl Constants<'_> {
    fn global(&mut self, node: P<Node>) -> bool {
        if node.kind() != Kind::Identifier
            || !super::rsbuild_binding_checks::ECMASCRIPT_GLOBALS.contains(&node.text())
        {
            return false;
        }
        if let Some(symbol) =
            super::rsbuild_assignment_checks::resolve_reference(node, self.checker)
        {
            return !symbol.declarations().iter().any(|&declaration| {
                if matches!(
                    declaration.kind(),
                    Kind::InterfaceDeclaration | Kind::TypeAliasDeclaration
                ) || ast::get_source_file_of_node(declaration) != Some(self.file)
                {
                    return false;
                }
                let mut current = declaration.parent();
                while let Some(parent) = current {
                    if parent.kind() == Kind::ModuleDeclaration
                        && (ast::is_global_scope_augmentation(parent)
                            || ast::is_external_module_augmentation(parent))
                    {
                        return false;
                    }
                    current = parent.parent();
                }
                true
            });
        }
        !super::rsbuild_ts_syntax::lexically_shadowed(node, node.text())
    }
    fn nullish(&mut self, node: P<Node>) -> bool {
        match node.kind() {
            Kind::NullKeyword | Kind::VoidExpression => true,
            Kind::Identifier => node.text() == "undefined" && self.global(node),
            Kind::ParenthesizedExpression => {
                self.nullish(node.as_parenthesized_expression().expression())
            }
            _ => false,
        }
    }
    fn boolean_literal(&self, node: P<Node>) -> Option<bool> {
        match node.kind() {
            Kind::TrueKeyword => Some(true),
            Kind::FalseKeyword | Kind::NullKeyword => Some(false),
            Kind::NumericLiteral => Some(!matches!(node.text(), "0" | "0.0" | "-0")),
            Kind::BigIntLiteral if self.condition_rule => Some(!matches!(
                node.text(),
                "0n" | "0x0n" | "0X0n" | "0b0n" | "0B0n" | "0o0n" | "0O0n"
            )),
            Kind::StringLiteral | Kind::NoSubstitutionTemplateLiteral => {
                let len = node.text().len();
                Some(if self.condition_rule && len >= 2 {
                    len - 2 > 0
                } else {
                    len > 0
                })
            }
            Kind::Identifier if self.condition_rule && node.text() == "undefined" => Some(false),
            _ => None,
        }
    }
    fn identity(&self, node: P<Node>, op: Kind) -> bool {
        if self.boolean_literal(node).is_some_and(|value| {
            op == Kind::BarBarToken && value || op == Kind::AmpersandAmpersandToken && !value
        }) {
            return true;
        }
        if op == Kind::BarBarToken
            && matches!(
                node.kind(),
                Kind::RegularExpressionLiteral
                    | Kind::ArrowFunction
                    | Kind::FunctionExpression
                    | Kind::ClassExpression
                    | Kind::ObjectLiteralExpression
                    | Kind::ArrayLiteralExpression
            )
        {
            return true;
        }
        if node.kind() == Kind::VoidExpression {
            return op == Kind::AmpersandAmpersandToken;
        }
        if node.kind() == Kind::ParenthesizedExpression {
            return self.identity(node.as_parenthesized_expression().expression(), op);
        }
        if node.kind() == Kind::BinaryExpression {
            let binary = node.as_binary_expression();
            let own = binary.operator_token().kind();
            if own == op && matches!(op, Kind::BarBarToken | Kind::AmpersandAmpersandToken) {
                return self.identity(binary.left(), op) || self.identity(binary.right(), op);
            }
            if matches!(
                own,
                Kind::BarBarEqualsToken | Kind::AmpersandAmpersandEqualsToken
            ) {
                return op
                    == (if own == Kind::BarBarEqualsToken {
                        Kind::BarBarToken
                    } else {
                        Kind::AmpersandAmpersandToken
                    })
                    && self.identity(binary.right(), op);
            }
        }
        false
    }
    fn constant(&mut self, node: P<Node>, boolean: bool) -> bool {
        match node.kind() {
            Kind::NumericLiteral
            | Kind::StringLiteral
            | Kind::NoSubstitutionTemplateLiteral
            | Kind::TrueKeyword
            | Kind::FalseKeyword
            | Kind::NullKeyword
            | Kind::RegularExpressionLiteral
            | Kind::BigIntLiteral
            | Kind::ArrowFunction
            | Kind::FunctionExpression
            | Kind::ClassExpression
            | Kind::ObjectLiteralExpression
            | Kind::VoidExpression => true,
            Kind::Identifier => {
                if self.condition_rule {
                    matches!(node.text(), "undefined" | "Infinity")
                } else {
                    node.text() == "undefined" && self.global(node)
                }
            }
            Kind::ArrayLiteralExpression => {
                boolean
                    || node.elements().iter().all(|&element| {
                        element.kind() == Kind::OmittedExpression
                            || self.constant(
                                if element.kind() == Kind::SpreadElement {
                                    element.as_spread_element().expression()
                                } else {
                                    element
                                },
                                false,
                            )
                    })
            }
            Kind::TemplateExpression => {
                let template = node.as_template_expression();
                boolean
                    && (!template.head().text().is_empty()
                        || template
                            .template_spans()
                            .nodes()
                            .iter()
                            .any(|span| !span.as_template_span().literal().text().is_empty()))
                    || template
                        .template_spans()
                        .nodes()
                        .iter()
                        .all(|span| self.constant(span.as_template_span().expression(), false))
            }
            Kind::ParenthesizedExpression => {
                self.constant(node.as_parenthesized_expression().expression(), boolean)
            }
            Kind::TypeOfExpression => {
                boolean || self.constant(node.as_type_of_expression().expression(), false)
            }
            Kind::DeleteExpression if !self.condition_rule => {
                self.constant(node.as_delete_expression().expression(), false)
            }
            Kind::PrefixUnaryExpression => {
                let prefix = node.as_prefix_unary_expression();
                match prefix.operator() {
                    Kind::ExclamationToken => self.constant(prefix.operand(), true),
                    Kind::PlusToken | Kind::MinusToken | Kind::TildeToken => {
                        self.constant(prefix.operand(), false)
                    }
                    _ => false,
                }
            }
            Kind::BinaryExpression => {
                let binary = node.as_binary_expression();
                let op = binary.operator_token().kind();
                if matches!(op, Kind::CommaToken | Kind::EqualsToken) {
                    return self.constant(binary.right(), boolean);
                }
                if matches!(
                    op,
                    Kind::BarBarEqualsToken | Kind::AmpersandAmpersandEqualsToken
                ) {
                    return boolean
                        && self.identity(
                            binary.right(),
                            if op == Kind::BarBarEqualsToken {
                                Kind::BarBarToken
                            } else {
                                Kind::AmpersandAmpersandToken
                            },
                        );
                }
                if logical_assignment(op) || compound_assignment(op) {
                    return false;
                }
                if matches!(
                    op,
                    Kind::AmpersandAmpersandToken | Kind::BarBarToken | Kind::QuestionQuestionToken
                ) {
                    let left = self.constant(binary.left(), boolean);
                    let right = self.constant(binary.right(), boolean);
                    return left && (right || self.identity(binary.left(), op))
                        || boolean && right && self.identity(binary.right(), op);
                }
                if numeric_operator(op) || comparison(op) {
                    return self.constant(binary.left(), false)
                        && self.constant(binary.right(), false);
                }
                if self.condition_rule
                    && op == Kind::InKeyword
                    && !matches!(
                        binary.right().kind(),
                        Kind::ObjectLiteralExpression | Kind::ArrayLiteralExpression
                    )
                {
                    return self.constant(binary.left(), false)
                        && self.constant(binary.right(), false);
                }
                false
            }
            Kind::ConditionalExpression if self.condition_rule => {
                self.constant(node.as_conditional_expression().condition(), false)
            }
            Kind::NewExpression => {
                if boolean {
                    return true;
                }
                if !self.condition_rule {
                    return false;
                }
                let call = node.as_new_expression();
                call.expression().kind() == Kind::Identifier
                    && matches!(call.expression().text(), "Boolean" | "String" | "Number")
                    && node
                        .arguments()
                        .iter()
                        .all(|&argument| self.constant(argument, false))
            }
            Kind::SpreadElement if !self.condition_rule => {
                self.constant(node.as_spread_element().expression(), boolean)
            }
            Kind::CallExpression => {
                let call = node.as_call_expression();
                let callee = call.expression();
                if callee.kind() != Kind::Identifier {
                    return false;
                }
                if !self.condition_rule {
                    return callee.text() == "Boolean"
                        && self.global(callee)
                        && node
                            .arguments()
                            .first()
                            .is_none_or(|&argument| self.constant(argument, true));
                }
                if !boolean
                    || !self.has_type_information
                    || !matches!(callee.text(), "Boolean" | "String" | "Number")
                {
                    return false;
                }
                let Some(symbol) = self.checker.get_symbol_at_location_exported(callee) else {
                    return false;
                };
                if symbol.declarations().iter().any(|&declaration| {
                    ast::get_source_file_of_node(declaration) == Some(self.file)
                }) || !symbol.declarations().iter().any(|&declaration| {
                    ast::get_source_file_of_node(declaration).is_some_and(self.default_library)
                }) {
                    return false;
                }
                node.arguments().iter().all(|&argument| {
                    argument.kind() != Kind::SpreadElement
                        && self.constant(argument, callee.text() == "Boolean")
                })
            }
            _ => false,
        }
    }
    fn static_boolean(&mut self, node: P<Node>) -> bool {
        match node.kind() {
            Kind::TrueKeyword | Kind::FalseKeyword => true,
            Kind::ParenthesizedExpression => {
                self.static_boolean(node.as_parenthesized_expression().expression())
            }
            Kind::PrefixUnaryExpression => {
                let prefix = node.as_prefix_unary_expression();
                prefix.operator() == Kind::ExclamationToken && self.constant(prefix.operand(), true)
            }
            Kind::CallExpression => {
                let call = node.as_call_expression();
                call.expression().kind() == Kind::Identifier
                    && call.expression().text() == "Boolean"
                    && self.global(call.expression())
                    && node
                        .arguments()
                        .first()
                        .is_none_or(|&argument| self.constant(argument, true))
            }
            _ => false,
        }
    }
    fn always_new(&mut self, node: P<Node>) -> bool {
        match node.kind() {
            Kind::ObjectLiteralExpression
            | Kind::ArrayLiteralExpression
            | Kind::ArrowFunction
            | Kind::FunctionExpression
            | Kind::ClassExpression
            | Kind::RegularExpressionLiteral => true,
            Kind::NewExpression => self.global(node.as_new_expression().expression()),
            Kind::ParenthesizedExpression => {
                self.always_new(node.as_parenthesized_expression().expression())
            }
            Kind::ConditionalExpression => {
                let value = node.as_conditional_expression();
                self.always_new(value.when_true()) && self.always_new(value.when_false())
            }
            Kind::BinaryExpression => {
                let binary = node.as_binary_expression();
                matches!(
                    binary.operator_token().kind(),
                    Kind::CommaToken | Kind::EqualsToken
                ) && self.always_new(binary.right())
            }
            _ => false,
        }
    }
    fn constant_nullishness(&mut self, node: P<Node>, non_nullish: bool) -> bool {
        if non_nullish && self.nullish(node) {
            return false;
        }
        match node.kind() {
            Kind::ObjectLiteralExpression
            | Kind::ArrayLiteralExpression
            | Kind::ArrowFunction
            | Kind::FunctionExpression
            | Kind::ClassExpression
            | Kind::NewExpression
            | Kind::RegularExpressionLiteral
            | Kind::TrueKeyword
            | Kind::FalseKeyword
            | Kind::NullKeyword
            | Kind::NumericLiteral
            | Kind::StringLiteral
            | Kind::NoSubstitutionTemplateLiteral
            | Kind::BigIntLiteral
            | Kind::TemplateExpression
            | Kind::VoidExpression
            | Kind::TypeOfExpression
            | Kind::DeleteExpression
            | Kind::PrefixUnaryExpression
            | Kind::PostfixUnaryExpression => true,
            Kind::Identifier => node.text() == "undefined" && self.global(node),
            Kind::ParenthesizedExpression => self
                .constant_nullishness(node.as_parenthesized_expression().expression(), non_nullish),
            Kind::BinaryExpression => {
                let binary = node.as_binary_expression();
                let op = binary.operator_token().kind();
                match op {
                    Kind::QuestionQuestionToken => self.constant_nullishness(binary.right(), true),
                    Kind::AmpersandAmpersandToken | Kind::BarBarToken => false,
                    Kind::CommaToken | Kind::EqualsToken => {
                        self.constant_nullishness(binary.right(), non_nullish)
                    }
                    _ => !logical_assignment(op),
                }
            }
            Kind::CallExpression => {
                let callee = node.as_call_expression().expression();
                callee.kind() == Kind::Identifier
                    && matches!(callee.text(), "Boolean" | "String" | "Number")
                    && self.global(callee)
            }
            _ => false,
        }
    }
    fn loose_boolean_comparison(&mut self, node: P<Node>) -> bool {
        match node.kind() {
            Kind::ObjectLiteralExpression
            | Kind::ClassExpression
            | Kind::ArrowFunction
            | Kind::FunctionExpression
            | Kind::VoidExpression
            | Kind::TypeOfExpression
            | Kind::NumericLiteral
            | Kind::StringLiteral
            | Kind::NoSubstitutionTemplateLiteral
            | Kind::TrueKeyword
            | Kind::FalseKeyword
            | Kind::NullKeyword
            | Kind::RegularExpressionLiteral
            | Kind::BigIntLiteral => true,
            Kind::ArrayLiteralExpression => {
                node.elements().is_empty()
                    || node
                        .elements()
                        .iter()
                        .filter(|element| {
                            !matches!(
                                element.kind(),
                                Kind::OmittedExpression | Kind::SpreadElement
                            )
                        })
                        .count()
                        > 1
            }
            Kind::PrefixUnaryExpression => {
                let prefix = node.as_prefix_unary_expression();
                prefix.operator() == Kind::ExclamationToken && self.constant(prefix.operand(), true)
            }
            Kind::Identifier => node.text() == "undefined" && self.global(node),
            Kind::ParenthesizedExpression => {
                self.loose_boolean_comparison(node.as_parenthesized_expression().expression())
            }
            Kind::CallExpression => self.static_boolean(node),
            Kind::BinaryExpression => {
                let binary = node.as_binary_expression();
                matches!(
                    binary.operator_token().kind(),
                    Kind::CommaToken | Kind::EqualsToken
                ) && self.loose_boolean_comparison(binary.right())
            }
            _ => false,
        }
    }
    fn strict_boolean_comparison(&mut self, node: P<Node>) -> bool {
        match node.kind() {
            Kind::ObjectLiteralExpression
            | Kind::ArrayLiteralExpression
            | Kind::ArrowFunction
            | Kind::FunctionExpression
            | Kind::ClassExpression
            | Kind::NewExpression
            | Kind::RegularExpressionLiteral
            | Kind::NumericLiteral
            | Kind::StringLiteral
            | Kind::NoSubstitutionTemplateLiteral
            | Kind::BigIntLiteral
            | Kind::NullKeyword
            | Kind::TrueKeyword
            | Kind::FalseKeyword
            | Kind::TemplateExpression
            | Kind::VoidExpression
            | Kind::TypeOfExpression
            | Kind::PostfixUnaryExpression => true,
            Kind::Identifier => node.text() == "undefined",
            Kind::PrefixUnaryExpression => {
                let prefix = node.as_prefix_unary_expression();
                prefix.operator() != Kind::ExclamationToken || self.constant(prefix.operand(), true)
            }
            Kind::ParenthesizedExpression => {
                self.strict_boolean_comparison(node.as_parenthesized_expression().expression())
            }
            Kind::BinaryExpression => {
                let binary = node.as_binary_expression();
                let op = binary.operator_token().kind();
                numeric_operator(op)
                    || compound_assignment(op)
                    || matches!(op, Kind::CommaToken | Kind::EqualsToken)
                        && self.strict_boolean_comparison(binary.right())
            }
            Kind::CallExpression => {
                let callee = node.as_call_expression().expression();
                if callee.kind() != Kind::Identifier {
                    return false;
                }
                match callee.text() {
                    "String" | "Number" => self.global(callee),
                    "Boolean" => self.static_boolean(node),
                    _ => false,
                }
            }
            _ => false,
        }
    }
    fn constant_comparison(&mut self, left: P<Node>, right: P<Node>, op: Kind) -> bool {
        self.nullish(left) && self.constant_nullishness(right, false)
            || self.static_boolean(left)
                && if matches!(op, Kind::EqualsEqualsToken | Kind::ExclamationEqualsToken) {
                    self.loose_boolean_comparison(right)
                } else {
                    self.strict_boolean_comparison(right)
                }
    }
    fn binary(&mut self, node: P<Node>) -> Option<Violation> {
        let binary = node.as_binary_expression();
        let op = binary.operator_token().kind();
        let property = match op {
            Kind::AmpersandAmpersandToken => Some("&&"),
            Kind::BarBarToken => Some("||"),
            Kind::QuestionQuestionToken => Some("??"),
            _ => None,
        };
        if let Some(property) = property {
            let constant = if op == Kind::QuestionQuestionToken {
                self.constant_nullishness(binary.left(), false)
            } else {
                self.constant(binary.left(), true)
            };
            return constant.then(|| Violation::new(node, "constantShortCircuit", format!("Unexpected constant {property} on the left-hand side of a `{property}` expression.")));
        }
        if !matches!(
            op,
            Kind::EqualsEqualsToken
                | Kind::ExclamationEqualsToken
                | Kind::EqualsEqualsEqualsToken
                | Kind::ExclamationEqualsEqualsToken
        ) {
            return None;
        }
        if self.constant_comparison(binary.left(), binary.right(), op)
            || self.constant_comparison(binary.right(), binary.left(), op)
        {
            return Some(Violation::new(
                node,
                "constantBinaryOperand",
                "Unexpected constant binary expression. Comparisons will always evaluate the same.",
            ));
        }
        if matches!(
            op,
            Kind::EqualsEqualsEqualsToken | Kind::ExclamationEqualsEqualsToken
        ) {
            if self.always_new(binary.left()) || self.always_new(binary.right()) {
                return Some(Violation::new(
                    node,
                    "alwaysNew",
                    "Unexpected comparison to newly constructed object. These two values can never be equal.",
                ));
            }
        } else if self.always_new(binary.left()) && self.always_new(binary.right()) {
            return Some(Violation::new(
                node,
                "bothAlwaysNew",
                "Unexpected comparison of two newly constructed objects. These two values can never be equal.",
            ));
        }
        None
    }
    fn condition(&mut self, node: P<Node>) -> Option<Violation> {
        let test = match node.kind() {
            Kind::IfStatement => node.as_if_statement().expression(),
            Kind::ConditionalExpression => node.as_conditional_expression().condition(),
            Kind::WhileStatement => {
                let value = node.as_while_statement();
                if ast::skip_parentheses(value.expression()).kind() == Kind::TrueKeyword
                    || contains_yield(value.statement())
                {
                    return None;
                }
                value.expression()
            }
            Kind::DoStatement => {
                let value = node.as_do_statement();
                if contains_yield(value.statement()) {
                    return None;
                }
                value.expression()
            }
            Kind::ForStatement => {
                let value = node.as_for_statement();
                if contains_yield(value.statement()) {
                    return None;
                }
                value.condition()?
            }
            _ => return None,
        };
        self.constant(test, true)
            .then(|| Violation::new(test, "unexpected", "Unexpected constant condition."))
    }
}
fn contains_yield(node: P<Node>) -> bool {
    if node.kind() == Kind::YieldExpression {
        return true;
    }
    if matches!(
        node.kind(),
        Kind::FunctionDeclaration | Kind::FunctionExpression | Kind::ArrowFunction
    ) {
        return false;
    }
    children(node).into_iter().any(contains_yield)
}

fn logical(node: P<Node>, op: Kind) -> bool {
    let node = ast::skip_parentheses(node);
    node.kind() == Kind::BinaryExpression
        && node.as_binary_expression().operator_token().kind() == op
}
fn split_logical(node: P<Node>, op: Kind) -> Vec<P<Node>> {
    let inner = ast::skip_parentheses(node);
    if logical(inner, op) {
        let binary = inner.as_binary_expression();
        let mut result = split_logical(binary.left(), op);
        result.extend(split_logical(binary.right(), op));
        result
    } else {
        vec![node]
    }
}
fn matrix(node: P<Node>) -> Vec<Vec<P<Node>>> {
    split_logical(node, Kind::BarBarToken)
        .into_iter()
        .map(|node| split_logical(node, Kind::AmpersandAmpersandToken))
        .collect()
}
fn duplicate_else_if(node: P<Node>, file: P<SourceFile>) -> Option<Violation> {
    if !node.parent().is_some_and(|parent| {
        parent.kind() == Kind::IfStatement
            && parent.as_if_statement().else_statement() == Some(node)
    }) {
        return None;
    }
    let test = node.as_if_statement().expression();
    let mut candidates = vec![test];
    if logical(test, Kind::AmpersandAmpersandToken) {
        candidates.extend(split_logical(test, Kind::AmpersandAmpersandToken));
    }
    let mut matrices: Vec<_> = candidates.into_iter().map(matrix).collect();
    let mut tokens = Vec::new();
    crate::comments::for_each_token_range(file, |_, full_start, end, _| {
        // This traversal yields token *full* starts so comment consumers see
        // trivia. Equality uses the trimmed token span, like frozen ForEachToken.
        let start = skip_trivia(file.text(), full_start);
        if start < end {
            tokens.push((start as usize, end as usize));
        }
        Ok(())
    })
    .expect("parsed token walk");
    fn equal(
        left: P<Node>,
        right: P<Node>,
        tokens: &[(usize, usize)],
        file: P<SourceFile>,
    ) -> bool {
        let left = ast::skip_parentheses(left);
        let right = ast::skip_parentheses(right);
        if left.kind() != right.kind() {
            return false;
        }
        if left.kind() == Kind::BinaryExpression {
            let a = left.as_binary_expression();
            let b = right.as_binary_expression();
            let op = a.operator_token().kind();
            if op == b.operator_token().kind()
                && matches!(op, Kind::BarBarToken | Kind::AmpersandAmpersandToken)
            {
                return equal(a.left(), b.left(), tokens, file)
                    && equal(a.right(), b.right(), tokens, file)
                    || equal(a.left(), b.right(), tokens, file)
                        && equal(a.right(), b.left(), tokens, file);
            }
        }
        let signature = |node| {
            let from = start(node, file);
            let to = end(node);
            tokens
                .iter()
                .filter(move |&&(a, b)| a >= from && b <= to)
                .map(|&(a, b)| &file.text()[a..b])
                .collect::<Vec<_>>()
        };
        signature(left) == signature(right)
    }
    let mut current = node;
    while let Some(parent) = current.parent().filter(|parent| {
        parent.kind() == Kind::IfStatement
            && parent.as_if_statement().else_statement() == Some(current)
    }) {
        let prior = matrix(parent.as_if_statement().expression());
        current = parent;
        for candidates in &mut matrices {
            candidates.retain(|and| {
                !prior.iter().any(|subset| {
                    subset
                        .iter()
                        .all(|&left| and.iter().any(|&right| equal(left, right, &tokens, file)))
                })
            });
            if candidates.is_empty() {
                return Some(Violation::new(
                    ast::skip_parentheses(test),
                    "unexpected",
                    "This branch can never execute. Its condition is a duplicate or covered by previous conditions in the if-else-if chain.",
                ));
            }
        }
    }
    None
}
fn block_terminal(node: P<Node>) -> bool {
    node.statements().last().is_some_and(|&node| {
        matches!(
            node.kind(),
            Kind::BreakStatement
                | Kind::ContinueStatement
                | Kind::ReturnStatement
                | Kind::ThrowStatement
        ) || node.kind() == Kind::Block && block_terminal(node)
    })
}
fn can_block_throw(node: P<Node>) -> bool {
    if node.statements().is_empty() {
        return false;
    }
    for &statement in node.statements() {
        match statement.kind() {
            Kind::BreakStatement | Kind::ContinueStatement => return false,
            Kind::ReturnStatement => return statement.expression().is_some(),
            Kind::ThrowStatement => return true,
            Kind::EmptyStatement => {}
            Kind::Block => return can_block_throw(statement),
            Kind::TryStatement => {
                return !statement
                    .as_try_statement()
                    .finally_block()
                    .is_some_and(block_terminal);
            }
            _ => return true,
        }
    }
    true
}
fn literal_truthy(node: P<Node>) -> bool {
    match node.kind() {
        Kind::TrueKeyword => true,
        Kind::NumericLiteral => {
            let number = jsnum::from_string(&node.text().replace('_', "")).0;
            number != 0.0 && !number.is_nan()
        }
        Kind::StringLiteral => !node.text().is_empty(),
        _ => false,
    }
}
fn loop_has_break(node: P<Node>) -> bool {
    let mut labels = Vec::new();
    let mut parent = node.parent();
    while let Some(value) = parent.filter(|node| node.kind() == Kind::LabeledStatement) {
        labels.push(value.as_labeled_statement().label().text());
        parent = value.parent();
    }
    fn visit(node: P<Node>, labels: &[&str]) -> bool {
        if matches!(
            node.kind(),
            Kind::SwitchStatement
                | Kind::ForStatement
                | Kind::ForOfStatement
                | Kind::ForInStatement
                | Kind::DoStatement
                | Kind::WhileStatement
        ) || ast::is_function_like(node)
        {
            return false;
        }
        if node.kind() == Kind::BreakStatement {
            return node
                .as_break_statement()
                .label()
                .is_none_or(|label| labels.contains(&label.text()));
        }
        children(node).into_iter().any(|child| visit(child, labels))
    }
    children(node)
        .into_iter()
        .any(|child| visit(child, &labels))
}
fn ast_terminal(node: P<Node>) -> bool {
    match node.kind() {
        Kind::Block => node.statements().iter().copied().any(ast_terminal),
        Kind::IfStatement => {
            let statement = node.as_if_statement();
            ast_terminal(statement.then_statement())
                && statement.else_statement().is_some_and(ast_terminal)
        }
        Kind::LabeledStatement => ast_terminal(node.as_labeled_statement().statement()),
        Kind::WhileStatement => {
            literal_truthy(node.as_while_statement().expression()) && !loop_has_break(node)
        }
        Kind::DoStatement => {
            literal_truthy(node.as_do_statement().expression()) && !loop_has_break(node)
        }
        Kind::ForStatement => {
            node.as_for_statement()
                .condition()
                .is_some_and(literal_truthy)
                && !loop_has_break(node)
        }
        Kind::TryStatement => {
            let statement = node.as_try_statement();
            statement.finally_block().is_some_and(block_terminal)
                || !can_block_throw(statement.try_block())
        }
        _ => false,
    }
}
fn js_whitespace(character: char) -> bool {
    matches!(character, '\u{9}'..='\u{d}' | '\u{20}' | '\u{a0}' | '\u{1680}' | '\u{2000}'..='\u{200a}' | '\u{2028}' | '\u{2029}' | '\u{202f}' | '\u{205f}' | '\u{3000}' | '\u{feff}')
}
fn fallthrough_comment(text: &str) -> bool {
    // The default pattern is the fixed ECMAScript /falls?\s?through/iu.
    // Only long-s has a non-ASCII simple fold into one of its ASCII letters.
    let folded: String = text
        .chars()
        .map(|c| {
            if c == '\u{17f}' {
                's'
            } else {
                c.to_ascii_lowercase()
            }
        })
        .collect();
    folded.match_indices("fall").any(|(index, _)| {
        let tail = &folded[index + 4..];
        let tail = tail.strip_prefix('s').unwrap_or(tail);
        let tail = if let Some(c) = tail.chars().next().filter(|&c| js_whitespace(c)) {
            &tail[c.len_utf8()..]
        } else {
            tail
        };
        tail.starts_with("through")
    })
}
fn fallthrough(node: P<Node>, file: P<SourceFile>) -> Vec<Violation> {
    let clauses = node
        .as_switch_statement()
        .case_block()
        .as_case_block()
        .clauses();
    let mut reports = Vec::new();
    for pair in clauses.nodes().windows(2) {
        let statements = pair[0].statements();
        let Some(last) = statements.last() else {
            continue;
        };
        if pair[0]
            .as_case_or_default_clause()
            .fallthrough_flow_node()
            .is_none()
            || statements.iter().copied().any(ast_terminal)
        {
            continue;
        }
        let next = pair[1];
        let comment_start = end(*last);
        let next_start = start(next, file);
        if comment_start < next_start
            && fallthrough_comment(&file.text()[comment_start..next_start])
        {
            continue;
        }
        let label = if next.kind() == Kind::DefaultClause {
            "default"
        } else {
            "case"
        };
        reports.push(Violation::new(
            next,
            label,
            format!("Expected a 'break' statement before '{label}'."),
        ));
    }
    reports
}
