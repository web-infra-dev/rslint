//! Control-flow rules ported from rslint e0a29f751d4def86d27931759098f27748fa3003.
//! The reachability builder below preserves the frozen `internal/utils/cfg`
//! statement, branch, loop, exception and finally routing. Compiler nodes and
//! binder flags come directly from the pinned tsrs Program; no source-text CFG.
use super::rsbuild_cfg::Cfg;
use crate::Violation;
use std::collections::HashSet;
use tsrs_ast::{self as ast, FunctionFlags, Kind, ModifierFlags, Node, NodeFlags, SourceFile};
use tsrs_checker::Checker;
use tsrs_core::P;
use tsrs_scanner::skip_trivia;

pub(crate) const NAMES: &[&str] = &[
    "constructor-super",
    "no-this-before-super",
    "getter-return",
    "no-unreachable",
];
pub(crate) fn check(
    name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
) -> Vec<Violation> {
    match (name, node.kind()) {
        ("constructor-super", Kind::Constructor) => constructor_super(node),
        ("no-this-before-super", Kind::Constructor) => before_super(node),
        ("getter-return", Kind::GetAccessor) => getter_return(node, file),
        ("getter-return", Kind::ObjectLiteralExpression) => descriptor_getters(node, file, checker),
        ("no-unreachable", Kind::SourceFile) => unreachable(file),
        _ => Vec::new(),
    }
}
fn start(node: P<Node>, file: P<SourceFile>) -> usize {
    skip_trivia(file.text(), node.pos().expect_source("parsed node")) as usize
}
fn end(node: P<Node>) -> usize {
    node.end().expect_source("parsed node") as usize
}
fn children(node: P<Node>) -> Vec<P<Node>> {
    let mut result = Vec::new();
    node.for_each_child(&mut |child| {
        result.push(child);
        false
    });
    result
}
fn walk(node: P<Node>, f: &mut impl FnMut(P<Node>)) {
    if node.flags().intersects(NodeFlags::Reparsed) {
        return;
    }
    f(node);
    for child in children(node) {
        walk(child, f);
    }
}
fn scope(node: P<Node>) -> bool {
    ast::is_function_like_declaration(node) || ast::is_class_like(node)
}
fn root(node: P<Node>) -> bool {
    ast::is_function_like_declaration(node)
        || matches!(
            node.kind(),
            Kind::SourceFile | Kind::ClassStaticBlockDeclaration
        )
}
fn is_super(node: P<Node>) -> bool {
    node.kind() == Kind::CallExpression
        && node.as_call_expression().expression().kind() == Kind::SuperKeyword
}
fn clauses(node: P<Node>) -> Vec<P<Node>> {
    children(node.as_switch_statement().case_block())
        .into_iter()
        .filter(|n| matches!(n.kind(), Kind::CaseClause | Kind::DefaultClause))
        .collect()
}
fn body_statements(node: P<Node>) -> Vec<P<Node>> {
    if node.kind() == Kind::Block {
        node.statements().to_vec()
    } else {
        vec![node]
    }
}

fn find_super(node: P<Node>, calls: &mut Vec<P<Node>>) {
    if is_super(node) {
        calls.push(node);
    }
    // Constructor-super's frozen search has this narrower boundary, distinct
    // from no-this-before-super's all-function-like boundary below.
    if matches!(
        node.kind(),
        Kind::FunctionDeclaration
            | Kind::FunctionExpression
            | Kind::ArrowFunction
            | Kind::MethodDeclaration
            | Kind::ClassDeclaration
            | Kind::ClassExpression
    ) {
        return;
    }
    for child in children(node) {
        find_super(child, calls);
    }
}
fn possible_constructor(node: P<Node>) -> bool {
    match node.kind() {
        Kind::ParenthesizedExpression => {
            possible_constructor(node.as_parenthesized_expression().expression())
        }
        Kind::BinaryExpression => {
            let binary = node.as_binary_expression();
            matches!(
                binary.operator_token().kind(),
                Kind::EqualsToken | Kind::AmpersandAmpersandToken | Kind::BarBarToken
            ) && possible_constructor(binary.right())
        }
        Kind::ConditionalExpression => {
            let value = node.as_conditional_expression();
            possible_constructor(value.when_true()) && possible_constructor(value.when_false())
        }
        Kind::NumericLiteral
        | Kind::StringLiteral
        | Kind::TrueKeyword
        | Kind::FalseKeyword
        | Kind::NullKeyword => false,
        _ => true,
    }
}
fn invalid_extends(node: P<Node>) -> bool {
    match node.kind() {
        Kind::NullKeyword
        | Kind::NumericLiteral
        | Kind::StringLiteral
        | Kind::TrueKeyword
        | Kind::FalseKeyword => true,
        Kind::ParenthesizedExpression => {
            invalid_extends(node.as_parenthesized_expression().expression())
        }
        Kind::BinaryExpression => !possible_constructor(node),
        _ => false,
    }
}
fn direct_super(node: P<Node>) -> bool {
    node.kind() == Kind::ExpressionStatement && node.expression().is_some_and(is_super)
}
fn expression_super(node: P<Node>) -> bool {
    if is_super(node) {
        return true;
    }
    if node.kind() == Kind::ConditionalExpression {
        let value = node.as_conditional_expression();
        return expression_super(value.when_true()) && expression_super(value.when_false());
    }
    false
}
fn constructor_try_super(node: P<Node>) -> bool {
    let statement = node.as_try_statement();
    if statement
        .finally_block()
        .is_some_and(|body| constructor_paths(body.statements()))
    {
        return true;
    }
    constructor_paths(statement.try_block().statements())
        && statement
            .catch_clause()
            .is_none_or(|clause| constructor_paths(clause.as_catch_clause().block().statements()))
}
fn constructor_switch_super(node: P<Node>) -> bool {
    let clauses = clauses(node);
    clauses
        .iter()
        .any(|node| node.kind() == Kind::DefaultClause)
        && clauses
            .iter()
            .all(|&node| constructor_paths(&children(node)))
}
fn loop_guarantees_super(statements: &[P<Node>]) -> bool {
    for &statement in statements {
        if direct_super(statement) {
            return true;
        }
        match statement.kind() {
            Kind::ReturnStatement | Kind::ThrowStatement => return true,
            Kind::ExpressionStatement | Kind::VariableStatement | Kind::EmptyStatement => {}
            _ => return false,
        }
    }
    false
}
fn constructor_paths(statements: &[P<Node>]) -> bool {
    for &statement in statements {
        if direct_super(statement) {
            return true;
        }
        match statement.kind() {
            Kind::IfStatement => {
                let value = statement.as_if_statement();
                if constructor_paths(&[value.then_statement()])
                    && value
                        .else_statement()
                        .is_some_and(|n| constructor_paths(&[n]))
                {
                    return true;
                }
            }
            Kind::SwitchStatement if constructor_switch_super(statement) => return true,
            Kind::ReturnStatement | Kind::ThrowStatement => return true,
            Kind::ExpressionStatement if statement.expression().is_some_and(expression_super) => {
                return true;
            }
            Kind::Block if constructor_paths(statement.statements()) => return true,
            Kind::LabeledStatement
                if constructor_paths(&[statement.as_labeled_statement().statement()]) =>
            {
                return true;
            }
            Kind::TryStatement if constructor_try_super(statement) => return true,
            Kind::WhileStatement => {
                let value = statement.as_while_statement();
                if value.expression().kind() == Kind::TrueKeyword
                    && loop_guarantees_super(&body_statements(value.statement()))
                {
                    return true;
                }
            }
            Kind::ForStatement => {
                let value = statement.as_for_statement();
                if value
                    .condition()
                    .is_none_or(|n| n.kind() == Kind::TrueKeyword)
                    && loop_guarantees_super(&body_statements(value.statement()))
                {
                    return true;
                }
            }
            Kind::DoStatement
                if loop_guarantees_super(&body_statements(
                    statement.as_do_statement().statement(),
                )) =>
            {
                return true;
            }
            _ => {}
        }
    }
    false
}
fn shallow_terminates(node: P<Node>) -> bool {
    matches!(node.kind(), Kind::ReturnStatement | Kind::ThrowStatement)
        || node.kind() == Kind::Block
            && node
                .statements()
                .iter()
                .any(|n| matches!(n.kind(), Kind::ReturnStatement | Kind::ThrowStatement))
}
fn partial_return(body: P<Node>) -> bool {
    body.statements()
        .iter()
        .any(|&statement| match statement.kind() {
            Kind::SwitchStatement => true,
            Kind::IfStatement => {
                let value = statement.as_if_statement();
                shallow_terminates(value.then_statement())
                    != value.else_statement().is_some_and(shallow_terminates)
            }
            _ => false,
        })
}
fn contains_range(container: P<Node>, child: P<Node>) -> bool {
    child.pos() >= container.pos() && child.end() <= container.end()
}
fn loop_exits_after_super(body: P<Node>) -> bool {
    let mut seen = false;
    for statement in body_statements(body) {
        if !seen {
            if direct_super(statement) {
                seen = true;
                continue;
            }
            if matches!(
                statement.kind(),
                Kind::ExpressionStatement | Kind::VariableStatement | Kind::EmptyStatement
            ) {
                continue;
            }
            return false;
        }
        match statement.kind() {
            Kind::BreakStatement | Kind::ReturnStatement | Kind::ThrowStatement => return true,
            Kind::ExpressionStatement | Kind::VariableStatement | Kind::EmptyStatement => {}
            _ => return false,
        }
    }
    false
}
fn duplicate_in_loop(body: P<Node>, call: P<Node>) -> bool {
    let mut current = call.parent();
    while let Some(node) = current.filter(|&node| node != body) {
        match node.kind() {
            Kind::WhileStatement => {
                return !loop_exits_after_super(node.as_while_statement().statement());
            }
            Kind::DoStatement => {
                return !loop_exits_after_super(node.as_do_statement().statement());
            }
            Kind::ForStatement => {
                return !loop_exits_after_super(node.as_for_statement().statement());
            }
            Kind::ForInStatement | Kind::ForOfStatement => return false,
            _ => {}
        }
        current = node.parent();
    }
    false
}
fn duplicate_super(body: P<Node>, calls: &[P<Node>]) -> Vec<P<Node>> {
    let mut duplicates: Vec<_> = calls
        .iter()
        .copied()
        .filter(|&call| duplicate_in_loop(body, call))
        .collect();
    let mut first = false;
    for &statement in body.statements() {
        let statement_calls: Vec<_> = calls
            .iter()
            .copied()
            .filter(|&call| contains_range(statement, call))
            .collect();
        if direct_super(statement) {
            if first {
                for &call in &statement_calls {
                    if !duplicates.contains(&call) {
                        duplicates.push(call);
                    }
                }
            } else {
                first = true;
            }
        } else if first {
            for &call in &statement_calls {
                if !duplicates.contains(&call) {
                    duplicates.push(call);
                }
            }
        }
        match statement.kind() {
            Kind::IfStatement => {
                let value = statement.as_if_statement();
                if constructor_paths(&[value.then_statement()])
                    && value
                        .else_statement()
                        .is_some_and(|node| constructor_paths(&[node]))
                {
                    first = true;
                }
            }
            Kind::SwitchStatement if constructor_switch_super(statement) => first = true,
            Kind::ExpressionStatement => {
                if statement.expression().is_some_and(|expr| {
                    expr.kind() == Kind::ConditionalExpression && expression_super(expr)
                }) {
                    first = true;
                }
            }
            Kind::TryStatement => {
                let value = statement.as_try_statement();
                if constructor_paths(value.try_block().statements())
                    && let Some(finally) = value
                        .finally_block()
                        .filter(|n| constructor_paths(n.statements()))
                {
                    for &call in &statement_calls {
                        if contains_range(finally, call) && !duplicates.contains(&call) {
                            duplicates.push(call);
                        }
                    }
                }
                if constructor_try_super(statement) {
                    first = true;
                }
            }
            _ => {}
        }
    }
    duplicates
}
fn constructor_super(node: P<Node>) -> Vec<Violation> {
    let Some(class) = ast::get_containing_class(node) else {
        return Vec::new();
    };
    let valid = ast::get_class_extends_heritage_element(class).is_some_and(|node| {
        !invalid_extends(node.as_expression_with_type_arguments().expression())
    });
    let Some(body) = node.body() else {
        return Vec::new();
    };
    let mut calls = Vec::new();
    find_super(body, &mut calls);
    if !valid {
        return calls.into_iter().map(|call| Violation::new(call.as_call_expression().expression(), "badSuper", "Unexpected 'super()' call because this class does not extend a valid constructor.")).collect();
    }
    if calls.is_empty() {
        if !node.flags().intersects(NodeFlags::HasImplicitReturn) {
            return Vec::new();
        }
        return vec![if partial_return(body) {
            Violation::new(
                node,
                "missingSome",
                "Expected to call 'super()' in some paths.",
            )
        } else {
            Violation::new(
                node,
                "missingAll",
                "Expected to call 'super()' in all paths.",
            )
        }];
    }
    if !constructor_paths(body.statements()) {
        return vec![Violation::new(
            node,
            "missingSome",
            "Expected to call 'super()' in some paths.",
        )];
    }
    duplicate_super(body, &calls)
        .into_iter()
        .map(|call| {
            Violation::new(
                call.as_call_expression().expression(),
                "duplicate",
                "Unexpected duplicate 'super()' call.",
            )
        })
        .collect()
}

#[derive(Clone, Copy, PartialEq, Eq)]
enum Super {
    None,
    Some,
    All,
}
fn merge(left: Super, right: Super) -> Super {
    if left == Super::All && right == Super::All {
        Super::All
    } else if left != Super::None || right != Super::None {
        Super::Some
    } else {
        Super::None
    }
}
fn contains_super(node: P<Node>) -> bool {
    if scope(node) {
        return false;
    }
    is_super(node) || children(node).into_iter().any(contains_super)
}
fn always_exits(node: P<Node>) -> bool {
    match node.kind() {
        Kind::BreakStatement
        | Kind::ReturnStatement
        | Kind::ThrowStatement
        | Kind::ContinueStatement => true,
        Kind::Block => node.statements().iter().copied().any(always_exits),
        Kind::IfStatement => {
            let value = node.as_if_statement();
            always_exits(value.then_statement()) && value.else_statement().is_some_and(always_exits)
        }
        _ => false,
    }
}
struct BeforeSuper {
    reports: Vec<Violation>,
}
impl BeforeSuper {
    fn violations(&mut self, node: P<Node>, status: Super) {
        if status == Super::All || scope(node) {
            return;
        }
        match node.kind() {
            Kind::ThisKeyword => {
                self.reports.push(Violation::new(
                    node,
                    "noBeforeSuper",
                    "'this' is not allowed before 'super()'.",
                ));
                return;
            }
            Kind::SuperKeyword => {
                if !node
                    .parent()
                    .is_some_and(|p| p.kind() == Kind::CallExpression)
                {
                    self.reports.push(Violation::new(
                        node,
                        "noBeforeSuper",
                        "'super' is not allowed before 'super()'.",
                    ));
                }
                return;
            }
            Kind::CallExpression if is_super(node) => {
                self.arguments(node, status);
                return;
            }
            _ => {}
        }
        for child in children(node) {
            self.violations(child, status);
        }
    }
    fn arguments(&mut self, node: P<Node>, status: Super) {
        for child in children(node) {
            if child.kind() != Kind::SuperKeyword {
                self.violations(child, status);
            }
        }
    }
    fn expressions(&mut self, node: P<Node>, mut status: Super) -> Super {
        for child in children(node) {
            status = self.expression(child, status);
        }
        status
    }
    fn expression(&mut self, node: P<Node>, status: Super) -> Super {
        let node = ast::skip_parentheses(node);
        if is_super(node) {
            self.arguments(node, status);
            return Super::All;
        }
        match node.kind() {
            Kind::ConditionalExpression => {
                let value = node.as_conditional_expression();
                let condition = self.expression(value.condition(), status);
                let yes = self.expression(value.when_true(), condition);
                let no = self.expression(value.when_false(), condition);
                merge(yes, no)
            }
            Kind::BinaryExpression => {
                let value = node.as_binary_expression();
                match value.operator_token().kind() {
                    Kind::AmpersandAmpersandToken
                    | Kind::BarBarToken
                    | Kind::QuestionQuestionToken
                    | Kind::AmpersandAmpersandEqualsToken
                    | Kind::BarBarEqualsToken
                    | Kind::QuestionQuestionEqualsToken => {
                        let left = self.expression(value.left(), status);
                        if contains_super(value.right()) {
                            if left == Super::All {
                                Super::All
                            } else {
                                Super::Some
                            }
                        } else {
                            self.violations(value.right(), left);
                            left
                        }
                    }
                    Kind::EqualsToken => {
                        self.violations(value.left(), status);
                        self.expression(value.right(), status)
                    }
                    _ => {
                        let left = self.expression(value.left(), status);
                        self.expression(value.right(), left)
                    }
                }
            }
            Kind::VariableDeclarationList => self.variables(node, status),
            Kind::PrefixUnaryExpression
            | Kind::PostfixUnaryExpression
            | Kind::VoidExpression
            | Kind::TypeOfExpression
            | Kind::DeleteExpression => self.expressions(node, status),
            Kind::CallExpression | Kind::NewExpression => {
                let mut current = node
                    .expression()
                    .map_or(status, |callee| self.expression(callee, status));
                for &argument in node.arguments() {
                    current = self.expression(argument, current);
                }
                current
            }
            Kind::PropertyAccessExpression
            | Kind::ElementAccessExpression
            | Kind::SpreadElement => node
                .expression()
                .map_or(status, |expr| self.expression(expr, status)),
            Kind::ArrayLiteralExpression => {
                let mut current = status;
                for &element in node.as_array_literal_expression().elements().nodes() {
                    current = self.expression(element, current);
                }
                current
            }
            Kind::ObjectLiteralExpression => {
                let mut current = status;
                for &property in node.as_object_literal_expression().properties().nodes() {
                    if !scope(property) {
                        current = self.expressions(property, current);
                    }
                }
                current
            }
            Kind::TemplateExpression => {
                let mut current = status;
                for &span in node.as_template_expression().template_spans().nodes() {
                    current = self.expression(span.as_template_span().expression(), current);
                }
                current
            }
            Kind::TaggedTemplateExpression => {
                let value = node.as_tagged_template_expression();
                let tag = self.expression(value.tag(), status);
                self.expressions(value.template(), tag)
            }
            _ => {
                self.violations(node, status);
                status
            }
        }
    }
    fn variables(&mut self, node: P<Node>, mut status: Super) -> Super {
        for &declaration in node.as_variable_declaration_list().declarations().nodes() {
            if let Some(name) = declaration.name() {
                self.violations(name, status);
            }
            if let Some(initializer) = declaration.initializer() {
                status = self.expression(initializer, status);
            }
        }
        status
    }
    fn sequence(&mut self, statements: &[P<Node>], mut status: Super) -> Super {
        for &statement in statements {
            if status == Super::All {
                break;
            }
            status = self.statement(statement, status);
        }
        status
    }
    fn statement(&mut self, node: P<Node>, status: Super) -> Super {
        match node.kind() {
            Kind::ExpressionStatement => node
                .expression()
                .map_or(status, |expr| self.expression(expr, status)),
            Kind::IfStatement => {
                let value = node.as_if_statement();
                let condition = self.expression(value.expression(), status);
                let yes = self.statement(value.then_statement(), condition);
                if let Some(other) = value.else_statement() {
                    let no = self.statement(other, condition);
                    merge(yes, no)
                } else if yes != Super::None {
                    Super::Some
                } else {
                    status
                }
            }
            Kind::Block => self.sequence(node.statements(), status),
            Kind::LabeledStatement => {
                self.statement(node.as_labeled_statement().statement(), status)
            }
            Kind::ReturnStatement | Kind::ThrowStatement => {
                if let Some(expression) = node.expression() {
                    self.expression(expression, status);
                }
                Super::All
            }
            Kind::VariableStatement => {
                self.variables(node.as_variable_statement().declaration_list(), status)
            }
            Kind::DoStatement => {
                let value = node.as_do_statement();
                let body = self.statement(value.statement(), status);
                self.expression(value.expression(), body)
            }
            Kind::ForStatement => {
                let value = node.as_for_statement();
                let init = value
                    .initializer()
                    .map_or(status, |init| self.expression(init, status));
                let condition = value
                    .condition()
                    .map_or(init, |test| self.expression(test, init));
                if value.condition().is_none() {
                    return self.statement(value.statement(), condition);
                }
                if condition != Super::All {
                    let body = self.statement(value.statement(), condition);
                    if let Some(increment) = value
                        .incrementor()
                        .filter(|_| !always_exits(value.statement()))
                    {
                        self.expression(increment, body);
                    }
                }
                if condition != Super::All
                    && (contains_super(value.statement())
                        || value.incrementor().is_some_and(contains_super))
                {
                    Super::Some
                } else {
                    condition
                }
            }
            Kind::WhileStatement => {
                let value = node.as_while_statement();
                let condition = self.expression(value.expression(), status);
                if value.expression().kind() == Kind::TrueKeyword {
                    return self.statement(value.statement(), condition);
                }
                if condition != Super::All {
                    self.statement(value.statement(), condition);
                }
                if condition != Super::All && contains_super(value.statement()) {
                    Super::Some
                } else {
                    condition
                }
            }
            Kind::ForInStatement | Kind::ForOfStatement => {
                let value = node.as_for_in_or_of_statement();
                let condition = self.expression(value.expression(), status);
                if condition != Super::All {
                    self.violations(value.initializer(), condition);
                    self.statement(value.statement(), condition);
                }
                if condition != Super::All && contains_super(value.statement()) {
                    Super::Some
                } else {
                    condition
                }
            }
            Kind::SwitchStatement => {
                let expression = self.expression(node.as_switch_statement().expression(), status);
                let clauses = clauses(node);
                // The discriminant executes even without clauses. Do not
                // downgrade its guaranteed super() call to a possible call.
                if clauses.is_empty() {
                    return expression;
                }
                let mut default = false;
                let mut all = true;
                for (index, &clause) in clauses.iter().enumerate() {
                    default |= clause.kind() == Kind::DefaultClause;
                    let mut clause_status = self.sequence(clause.statements(), expression);
                    if clause_status == Super::All {
                        continue;
                    }
                    if !clause.statements().iter().copied().any(always_exits) {
                        for &next in &clauses[index + 1..] {
                            clause_status = self.sequence(next.statements(), clause_status);
                            if clause_status == Super::All {
                                break;
                            }
                        }
                    }
                    if clause_status != Super::All {
                        all = false;
                    }
                }
                if default && all {
                    Super::All
                } else if contains_super(node) {
                    Super::Some
                } else {
                    expression
                }
            }
            Kind::TryStatement => {
                let value = node.as_try_statement();
                let attempted = self.sequence(value.try_block().statements(), status);
                if let Some(finally) = value.finally_block() {
                    let mut combined = status;
                    if attempted == Super::All && status == Super::All {
                        combined = Super::All;
                    } else if attempted != Super::None {
                        combined = Super::Some;
                    }
                    if let Some(catch) = value.catch_clause() {
                        let caught =
                            self.sequence(catch.as_catch_clause().block().statements(), status);
                        combined = merge(combined, caught);
                    }
                    self.sequence(finally.statements(), combined)
                } else if let Some(catch) = value.catch_clause() {
                    let caught =
                        self.sequence(catch.as_catch_clause().block().statements(), status);
                    merge(attempted, caught)
                } else {
                    attempted
                }
            }
            _ => {
                self.violations(node, status);
                status
            }
        }
    }
}
fn before_super(node: P<Node>) -> Vec<Violation> {
    let Some(class) = ast::get_containing_class(node) else {
        return Vec::new();
    };
    let Some(extends) = ast::get_class_extends_heritage_element(class) else {
        return Vec::new();
    };
    if extends
        .as_expression_with_type_arguments()
        .expression()
        .kind()
        == Kind::NullKeyword
    {
        return Vec::new();
    }
    let Some(body) = node.body().filter(|n| n.kind() == Kind::Block) else {
        return Vec::new();
    };
    let mut analyzer = BeforeSuper {
        reports: Vec::new(),
    };
    for &parameter in node.parameters() {
        if let Some(init) = parameter.initializer() {
            analyzer.violations(init, Super::None);
        }
    }
    analyzer.sequence(body.statements(), Super::None);
    analyzer.reports
}

fn getter_name(node: P<Node>, file: P<SourceFile>) -> String {
    let parent = ast::walk_up_parenthesized_expressions(node.parent());
    let owner = parent
        .filter(|n| n.kind() == Kind::PropertyAssignment)
        .unwrap_or(node);
    let mut words = Vec::new();
    if owner.parent().is_some_and(ast::is_class_like) {
        if ast::has_syntactic_modifier(owner, ModifierFlags::Static) {
            words.push("static".to_owned());
        }
        if owner
            .name()
            .is_some_and(|n| n.kind() == Kind::PrivateIdentifier)
        {
            words.push("private".to_owned());
        }
    }
    let flags = ast::get_function_flags(Some(node));
    if flags.intersects(FunctionFlags::Async) {
        words.push("async".to_owned());
    }
    if flags.intersects(FunctionFlags::Generator) {
        words.push("generator".to_owned());
    }
    words.push(
        if node.kind() == Kind::GetAccessor {
            "getter"
        } else if node.kind() == Kind::MethodDeclaration || owner != node {
            "method"
        } else if node.kind() == Kind::ArrowFunction {
            "arrow function"
        } else {
            "function"
        }
        .to_owned(),
    );
    if let Some(name) = owner.name() {
        if name.kind() == Kind::PrivateIdentifier {
            words.push(name.text().to_owned());
        } else if let Some(text) = super::rsbuild_declaration_checks::static_property(name, file) {
            words.push(format!("'{text}'"));
        } else if node.kind() == Kind::FunctionExpression
            && let Some(name) = node.name()
        {
            words.push(format!("'{}'", name.text()));
        }
    }
    words.join(" ")
}
fn head_range(node: P<Node>, file: P<SourceFile>) -> [usize; 2] {
    let parent = ast::walk_up_parenthesized_expressions(node.parent());
    let owner = parent
        .filter(|n| n.kind() == Kind::PropertyAssignment)
        .unwrap_or(node);
    let mut from = start(owner, file);
    for child in children(owner) {
        if child.kind() == Kind::Decorator {
            from = skip_trivia(file.text(), end(child) as u32) as usize;
        }
    }
    let search = if matches!(node.kind(), Kind::GetAccessor | Kind::MethodDeclaration) {
        node.name().map_or(from, end)
    } else {
        start(node, file)
    };
    let mut paren = None;
    crate::comments::for_each_token_range(file, |kind, full_pos, _, _| {
        let pos = skip_trivia(file.text(), full_pos);
        if kind == Kind::OpenParenToken
            && (pos as usize) >= search
            && (pos as usize) < end(node)
            && paren.is_none()
        {
            paren = Some(pos as usize);
        }
        Ok(())
    })
    .expect("parsed token walk");
    if node.kind() == Kind::ArrowFunction && owner == node {
        let token = node
            .as_arrow_function()
            .equals_greater_than_token()
            .expect("parsed arrow");
        return [start(token, file), end(token)];
    }
    if node.kind() == Kind::ArrowFunction
        && node.parameters().len() == 1
        && (paren.is_none() || paren.is_some_and(|paren| paren > start(node.parameters()[0], file)))
    {
        return [from, start(node.parameters()[0], file)];
    }
    [
        from,
        paren.unwrap_or_else(|| {
            node.body().map_or(end(node), |body| {
                body.pos().expect_source("parsed body") as usize
            })
        }),
    ]
}
fn getter_return(node: P<Node>, file: P<SourceFile>) -> Vec<Violation> {
    let Some(body) = node.body().filter(|n| n.kind() == Kind::Block) else {
        return Vec::new();
    };
    let name = getter_name(node, file);
    let mut returns = Vec::new();
    ast::for_each_return_statement(body, |node| {
        returns.push(node);
        false
    });
    let mut reports: Vec<_> = returns
        .iter()
        .copied()
        .filter(|n| n.expression().is_none())
        .map(|node| {
            Violation::new(
                node,
                "expected",
                format!("Expected to return a value in {name}."),
            )
        })
        .collect();
    let reachable = if body
        .statements()
        .last()
        .is_some_and(|n| matches!(n.kind(), Kind::ReturnStatement | Kind::ThrowStatement))
    {
        false
    } else {
        let mut cfg = Cfg::new();
        cfg.statement(body);
        cfg.live()
    };
    if reachable {
        let (id, message) = if returns.is_empty() {
            ("expected", format!("Expected to return a value in {name}."))
        } else {
            (
                "expectedAlways",
                format!("Expected {name} to always return a value."),
            )
        };
        reports.push(Violation::new(node, id, message).with_byte_range(head_range(node, file)));
    }
    reports
}
fn global_method(
    call: P<Node>,
    object_name: &str,
    method: &str,
    file: P<SourceFile>,
    checker: &mut Checker,
) -> bool {
    let callee = ast::skip_parentheses(call.as_call_expression().expression());
    let Some((object, _, property)) = super::rsbuild_declaration_checks::access(callee, file)
    else {
        return false;
    };
    let object = ast::skip_parentheses(object);
    if property != method || object.kind() != Kind::Identifier || object.text() != object_name {
        return false;
    }
    if crate::directives::global_setting(file, object_name) == Some(false) {
        return false;
    }
    // RefStore in the frozen backend resolves authored lexical bindings
    // before library symbols. A conflicting script-level `let Object` may
    // otherwise be hidden by the checker's merged global Object symbol.
    if let Some(symbol) = super::rsbuild_assignment_checks::resolve_reference(object, checker) {
        return !symbol.declarations().iter().any(|&declaration| {
            if matches!(
                declaration.kind(),
                Kind::InterfaceDeclaration | Kind::TypeAliasDeclaration
            ) || ast::get_source_file_of_node(declaration) != Some(file)
            {
                return false;
            }
            let mut parent = declaration.parent();
            while let Some(node) = parent {
                if node.kind() == Kind::ModuleDeclaration
                    && (ast::is_global_scope_augmentation(node)
                        || ast::is_external_module_augmentation(node))
                {
                    return false;
                }
                parent = node.parent();
            }
            true
        });
    }
    true
}
fn is_descriptor(node: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> bool {
    let Some(parent) = ast::walk_up_parenthesized_expressions(node.parent()) else {
        return false;
    };
    if parent.kind() == Kind::CallExpression
        && parent
            .arguments()
            .get(2)
            .is_some_and(|&arg| ast::skip_parentheses(arg) == node)
        && (global_method(parent, "Object", "defineProperty", file, checker)
            || global_method(parent, "Reflect", "defineProperty", file, checker))
    {
        return true;
    }
    if parent.kind() != Kind::PropertyAssignment
        || !parent
            .initializer()
            .is_some_and(|init| ast::skip_parentheses(init) == node)
    {
        return false;
    }
    let Some(properties) = ast::walk_up_parenthesized_expressions(parent.parent())
        .filter(|n| n.kind() == Kind::ObjectLiteralExpression)
    else {
        return false;
    };
    let Some(call) = ast::walk_up_parenthesized_expressions(properties.parent())
        .filter(|n| n.kind() == Kind::CallExpression)
    else {
        return false;
    };
    call.arguments()
        .get(1)
        .is_some_and(|&arg| ast::skip_parentheses(arg) == properties)
        && (global_method(call, "Object", "create", file, checker)
            || global_method(call, "Object", "defineProperties", file, checker))
}
fn descriptor_getters(node: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> Vec<Violation> {
    if !is_descriptor(node, file, checker) {
        return Vec::new();
    }
    let mut reports = Vec::new();
    for &property in node.as_object_literal_expression().properties().nodes() {
        if property
            .name()
            .and_then(|name| super::rsbuild_declaration_checks::static_property(name, file))
            .as_deref()
            != Some("get")
        {
            continue;
        }
        let getter = match property.kind() {
            Kind::PropertyAssignment => property
                .initializer()
                .map(ast::skip_parentheses)
                .filter(|n| matches!(n.kind(), Kind::FunctionExpression | Kind::ArrowFunction)),
            Kind::MethodDeclaration => Some(property),
            _ => None,
        };
        if let Some(getter) = getter {
            reports.extend(getter_return(getter, file));
        }
    }
    reports
}

fn binder_unreachable(node: P<Node>) -> bool {
    if node.flow_node().is_some() {
        return false;
    }
    (node.kind() >= Kind::FirstStatement && node.kind() <= Kind::LastStatement)
        || node.flags().intersects(NodeFlags::Unreachable)
}
fn hoisted(node: P<Node>) -> bool {
    match node.kind() {
        Kind::FunctionDeclaration
        | Kind::EmptyStatement
        | Kind::TypeAliasDeclaration
        | Kind::InterfaceDeclaration => true,
        Kind::VariableStatement => {
            let list = node.as_variable_statement().declaration_list();
            !list
                .flags()
                .intersects(NodeFlags::Let | NodeFlags::Const | NodeFlags::Using)
                && list
                    .as_variable_declaration_list()
                    .declarations()
                    .nodes()
                    .iter()
                    .all(|node| node.initializer().is_none())
        }
        _ => false,
    }
}
fn terminal_block(node: P<Node>) -> bool {
    node.statements()
        .last()
        .is_some_and(|&last| match last.kind() {
            Kind::BreakStatement
            | Kind::ContinueStatement
            | Kind::ReturnStatement
            | Kind::ThrowStatement => true,
            Kind::Block => terminal_block(last),
            _ => false,
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
                    .is_some_and(terminal_block);
            }
            _ => return true,
        }
    }
    true
}
fn has_boolean_if(node: P<Node>, own: P<Node>) -> bool {
    if node != own && root(node) {
        return false;
    }
    if node.kind() == Kind::IfStatement
        && matches!(
            node.as_if_statement().expression().kind(),
            Kind::TrueKeyword | Kind::FalseKeyword
        )
    {
        return true;
    }
    children(node)
        .into_iter()
        .any(|child| has_boolean_if(child, own))
}
fn unreachable(file: P<SourceFile>) -> Vec<Violation> {
    let mut roots = Vec::new();
    walk(file.as_node(), &mut |node| {
        if root(node) {
            roots.push(node);
        }
    });
    let mut reachable = HashSet::new();
    for node in roots {
        if !has_boolean_if(node, node) {
            continue;
        }
        let mut cfg = Cfg::new();
        if node.kind() == Kind::SourceFile {
            cfg.sequence(file.statements().nodes());
        } else if node.kind() == Kind::ClassStaticBlockDeclaration {
            // tsrs stores a static block's body on its concrete declaration,
            // outside the generic function-like body data.
            cfg.statement(node.as_class_static_block_declaration().body());
        } else if let Some(body) = node.body() {
            cfg.statement(body);
        }
        reachable.extend(cfg.reached);
    }
    let dead = |node| binder_unreachable(node) && !reachable.contains(&node);
    let mut reports = Vec::new();
    walk(file.as_node(), &mut |node| {
        if matches!(
            node.kind(),
            Kind::SourceFile
                | Kind::Block
                | Kind::CaseClause
                | Kind::DefaultClause
                | Kind::ModuleBlock
        ) {
            let statements = node.statements();
            if statements.first().is_none_or(|&first| dead(first)) {
                return;
            }
            let mut group: Option<[usize; 2]> = None;
            for &statement in statements {
                if dead(statement) && !hoisted(statement) {
                    if let Some(range) = &mut group {
                        range[1] = end(statement);
                    } else {
                        group = Some([start(statement, file), end(statement)]);
                    }
                } else if let Some(range) = group.take() {
                    reports.push(
                        Violation::new(node, "unreachableCode", "Unreachable code.")
                            .with_byte_range(range),
                    );
                }
            }
            if let Some(range) = group {
                reports.push(
                    Violation::new(node, "unreachableCode", "Unreachable code.")
                        .with_byte_range(range),
                );
            }
        }
        if node.kind() == Kind::TryStatement && !dead(node) {
            let statement = node.as_try_statement();
            if let Some(clause) = statement
                .catch_clause()
                .filter(|_| !can_block_throw(statement.try_block()))
            {
                reports.push(Violation::new(
                    clause,
                    "unreachableCode",
                    "Unreachable code.",
                ));
            }
        }
    });
    reports
}
