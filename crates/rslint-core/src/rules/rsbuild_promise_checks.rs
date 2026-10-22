//! Promise rules from rslint e0a29f751d4def86d27931759098f27748fa3003.
//! Copyright (c) 2025 July-present Bytedance Inc and its affiliates.
//! Copyright (c) 2025 July typescript-eslint and other contributors.
//! MIT licensed; see the repository LICENSE. These implement the default options
//! enabled by the frozen Rsbuild configuration. All type queries use the owning
//! tsrs Program's checker; no declaration, callback, or Promise is text-matched.

use std::collections::HashSet;

use tsrs_ast::{self as ast, Kind, Node, SourceFile, Symbol, SymbolFlags};
use tsrs_checker::{Checker, ContextFlags, ObjectFlags, SignatureKind, Type, TypeFlags};
use tsrs_core::P;
use tsrs_scanner::skip_trivia;

use crate::{Violation, edits::Edit};

pub(crate) const NAMES: &[&str] = &[
    "@typescript-eslint/no-floating-promises",
    "@typescript-eslint/no-misused-promises",
];

pub(crate) fn check(
    name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    default_library: &dyn Fn(P<SourceFile>) -> bool,
) -> Vec<Violation> {
    let mut context = Context {
        file,
        checker,
        default_library,
        checked_conditionals: HashSet::new(),
        violations: Vec::new(),
    };
    match (name, node.kind()) {
        ("@typescript-eslint/no-floating-promises", Kind::ExpressionStatement) => {
            context.floating(node);
        }
        ("@typescript-eslint/no-misused-promises", Kind::SourceFile) => {
            context.walk_misused(node);
        }
        _ => {}
    }
    context.violations
}

fn parts(value: P<Type>) -> Vec<P<Type>> {
    if value.flags().intersects(TypeFlags::Union) {
        value.types().to_vec()
    } else {
        vec![value]
    }
}

fn start(node: P<Node>, file: P<SourceFile>) -> usize {
    skip_trivia(
        file.text(),
        node.pos().expect_source("parsed node start") as u32,
    ) as usize
}

fn end(node: P<Node>) -> usize {
    node.end().expect_source("parsed node end") as usize
}

fn higher_than_unary(node: P<Node>) -> bool {
    let operator = if node.kind() == Kind::BinaryExpression {
        node.as_binary_expression().operator_token.kind()
    } else {
        Kind::Unknown
    };
    ast::get_operator_precedence(node.kind(), operator, ast::OperatorPrecedenceFlags::None) as i32
        > ast::OperatorPrecedence::Unary as i32
}

/// IsPromiseLike checks instances of the actual active library Promise,
/// including base types and constrained parameters, but excludes constructors.
fn promise_like(
    value: P<Type>,
    checker: &mut Checker,
    default_library: &dyn Fn(P<SourceFile>) -> bool,
) -> bool {
    if value.flags().intersects(TypeFlags::Intersection) {
        return value
            .types()
            .iter()
            .any(|&part| promise_like(part, checker, default_library));
    }
    if value.flags().intersects(TypeFlags::Union) {
        return value
            .types()
            .iter()
            .all(|&part| promise_like(part, checker, default_library));
    }
    if value.flags().intersects(TypeFlags::TypeParameter) {
        return checker
            .get_base_constraint_of_type(value)
            .is_some_and(|constraint| promise_like(constraint, checker, default_library));
    }
    let Some(symbol) = value.symbol() else {
        return false;
    };
    if value.flags().intersects(TypeFlags::Object)
        && value.object_flags().intersects(ObjectFlags::Anonymous)
        && symbol.flags().intersects(SymbolFlags::Class)
    {
        return false;
    }
    if symbol.name() == "Promise"
        && symbol.declarations().iter().any(|&declaration| {
            ast::get_source_file_of_node(declaration).is_some_and(default_library)
        })
    {
        return true;
    }
    if symbol
        .flags()
        .intersects(SymbolFlags::Class | SymbolFlags::Interface)
    {
        let declared = checker.get_declared_type_of_symbol(symbol);
        return checker
            .get_base_types(declared)
            .iter()
            .any(|&base| promise_like(base, checker, default_library));
    }
    false
}

struct Context<'a> {
    file: P<SourceFile>,
    checker: &'a mut Checker,
    default_library: &'a dyn Fn(P<SourceFile>) -> bool,
    checked_conditionals: HashSet<P<Node>>,
    violations: Vec<Violation>,
}

impl Context<'_> {
    fn report(&mut self, node: P<Node>, message_id: &'static str, message: impl Into<String>) {
        self.violations
            .push(Violation::new(node, message_id, message));
    }

    fn callback(&mut self, parameter: P<Symbol>, node: P<Node>, unwrap_rest: bool) -> bool {
        let Some(value) = self
            .checker
            .get_type_of_symbol_at_location(parameter, Some(node))
        else {
            return false;
        };
        let mut value = self.checker.get_apparent_type(value);
        if unwrap_rest
            && parameter.value_declaration().is_some_and(|declaration| {
                declaration.kind() == Kind::Parameter
                    && declaration
                        .as_parameter_declaration()
                        .dot_dot_dot_token()
                        .is_some()
            })
        {
            let number = self.checker.get_number_type();
            let Some(element) = self.checker.get_index_type_of_type(value, number) else {
                return false;
            };
            value = element;
        }
        parts(value).into_iter().any(|part| {
            !self
                .checker
                .get_signatures_of_type(part, SignatureKind::Call)
                .is_empty()
        })
    }

    fn thenable_part(&mut self, node: P<Node>, value: P<Type>, unwrap_rest: bool) -> bool {
        let Some(then) = self.checker.get_property_of_type(value, "then") else {
            return false;
        };
        let Some(value) = self
            .checker
            .get_type_of_symbol_at_location(then, Some(node))
        else {
            return false;
        };
        for part in parts(value) {
            let signatures = self
                .checker
                .get_signatures_of_type(part, SignatureKind::Call);
            for &signature in signatures {
                if let Some(&parameter) = signature.parameters().first()
                    && self.callback(parameter, node, unwrap_rest)
                {
                    return true;
                }
            }
        }
        false
    }

    fn thenable(&mut self, node: P<Node>, value: P<Type>) -> bool {
        let apparent = self.checker.get_apparent_type(value);
        parts(apparent)
            .into_iter()
            .any(|part| self.thenable_part(node, part, true))
    }

    fn always_thenable(&mut self, node: P<Node>) -> bool {
        let value = self.checker.get_type_at_location(node);
        let apparent = self.checker.get_apparent_type(value);
        // The conditional variant deliberately does not unwrap rest callback
        // parameters; this is the frozen rule's isFunctionParam, not IsCallback.
        parts(apparent)
            .into_iter()
            .all(|part| self.thenable_part(node, part, false))
    }

    fn return_thenable_type(&mut self, node: P<Node>, value: P<Type>) -> bool {
        parts(value).into_iter().any(|part| {
            let signatures = self
                .checker
                .get_signatures_of_type(part, SignatureKind::Call);
            signatures.iter().any(|&signature| {
                let result = self.checker.get_return_type_of_signature(signature);
                self.thenable(node, result)
            })
        })
    }

    fn returns_thenable(&mut self, node: P<Node>) -> bool {
        let value = self.checker.get_type_at_location(node);
        let value = self.checker.get_apparent_type(value);
        self.return_thenable_type(node, value)
    }

    fn void_returning(&mut self, node: P<Node>, value: P<Type>) -> bool {
        let mut found_void = false;
        for part in parts(value) {
            let signatures = self
                .checker
                .get_signatures_of_type(part, SignatureKind::Call);
            for &signature in signatures {
                let result = self.checker.get_return_type_of_signature(signature);
                if self.thenable(node, result) {
                    return false;
                }
                found_void |= result.flags().intersects(TypeFlags::Void);
            }
        }
        found_void
    }

    fn floating_promise(&mut self, value: P<Type>) -> bool {
        let value = self.checker.get_apparent_type(value);
        parts(value)
            .into_iter()
            .any(|part| promise_like(part, self.checker, self.default_library))
    }

    fn promise_array(&mut self, value: P<Type>) -> bool {
        for part in parts(value) {
            let part = self.checker.get_apparent_type(part);
            if self.checker.is_array_type(part) {
                if let Some(&element) = self.checker.get_type_arguments(part).first()
                    && self.floating_promise(element)
                {
                    return true;
                }
            } else if tsrs_checker::is_tuple_type_exported(part) {
                for &element in self.checker.get_type_arguments(part) {
                    if self.floating_promise(element) {
                        return true;
                    }
                }
            }
        }
        false
    }

    /// (unhandled, non-function rejection handler, Promise array).
    fn unhandled(&mut self, node: P<Node>) -> (bool, bool, bool) {
        if ast::is_assignment_expression(node, false) {
            return (false, false, false);
        }
        if ast::is_comma_expression(node) {
            let binary = node.as_binary_expression();
            let left = self.unhandled(binary.left);
            return if left.0 {
                left
            } else {
                self.unhandled(binary.right.get())
            };
        }
        let value = self.checker.get_type_at_location(node);
        if self.promise_array(value) {
            return (true, false, true);
        }
        if node.kind() == Kind::AwaitExpression || !self.floating_promise(value) {
            return (false, false, false);
        }
        if node.kind() == Kind::CallExpression {
            let call = node.as_call_expression();
            let callee = call.expression;
            if ast::is_access_expression(callee) {
                let (method, _) = self.checker.get_accessed_property_name(callee);
                let arguments = call.arguments.nodes();
                let handler = match method.as_ref() {
                    "catch" => arguments.first(),
                    "then" => arguments.get(1),
                    "finally" => return self.unhandled(callee.expression().unwrap()),
                    _ => None,
                };
                if let Some(&handler) = handler {
                    let value = self.checker.get_type_at_location(handler);
                    let callable = !self
                        .checker
                        .get_signatures_of_type(value, SignatureKind::Call)
                        .is_empty();
                    return (!callable, !callable, false);
                }
            }
            return (true, false, false);
        }
        if node.kind() == Kind::ConditionalExpression {
            let conditional = node.as_conditional_expression();
            let when_false = self.unhandled(conditional.when_false);
            return if when_false.0 {
                when_false
            } else {
                self.unhandled(conditional.when_true)
            };
        }
        if ast::is_logical_or_coalescing_binary_expression(node) {
            let binary = node.as_binary_expression();
            let left = self.unhandled(binary.left);
            return if left.0 {
                left
            } else {
                self.unhandled(binary.right.get())
            };
        }
        (true, false, false)
    }

    fn add_unary(&self, statement: P<Node>, expression: P<Node>, operator: &str) -> Vec<Edit> {
        let beginning = start(statement, self.file);
        if higher_than_unary(statement.expression().unwrap()) {
            return vec![Edit::replace(
                [beginning, beginning],
                format!("{operator} "),
            )];
        }
        vec![
            Edit::replace([beginning, beginning], format!("{operator} (")),
            Edit::replace([end(expression), end(expression)], ")"),
        ]
    }

    fn floating(&mut self, statement: P<Node>) {
        let expression = ast::skip_parentheses(statement.expression().unwrap());
        let (unhandled, invalid_handler, array) = self.unhandled(expression);
        if !unhandled {
            return;
        }
        if array {
            self.report(statement, "floatingPromiseArrayVoid", "An array of Promises may be unintentional. Consider handling the promises' fulfillment or rejection with Promise.all or similar, or explicitly marking the expression as ignored with the `void` operator.");
            return;
        }
        let mut message = "Promises must be awaited, end with a call to .catch, end with a call to .then with a rejection handler or be explicitly marked as ignored with the `void` operator.".to_string();
        let id = if invalid_handler {
            message.push_str(" A rejection handler that is not a function will be ignored.");
            "floatingUselessRejectionHandlerVoid"
        } else {
            "floatingVoid"
        };
        self.violations.push(
            Violation::new(statement, id, message)
                .with_suggestion(
                    "floatingFixVoid",
                    "Add void operator to ignore.",
                    self.add_unary(statement, expression, "void"),
                )
                .with_suggestion(
                    "floatingFixAwait",
                    "Add await operator.",
                    self.add_unary(statement, expression, "await"),
                ),
        );
    }

    fn conditional(&mut self, node: Option<P<Node>>, test: bool) {
        let Some(node) = node else { return };
        if ast::is_assignment_expression(node, false) || !self.checked_conditionals.insert(node) {
            return;
        }
        let node = ast::skip_parentheses(node);
        if node.kind() == Kind::BinaryExpression && ast::is_logical_expression(node) {
            let binary = node.as_binary_expression();
            if binary.operator_token.kind() != Kind::QuestionQuestionToken || test {
                self.conditional(Some(binary.left), test);
            }
            if test {
                self.conditional(Some(binary.right.get()), test);
            }
        } else if self.always_thenable(node) {
            self.report(
                node,
                "conditional",
                "Expected non-Promise value in a boolean conditional.",
            );
        }
    }

    fn array_or_tuple(&mut self, value: P<Type>) -> bool {
        if value
            .flags()
            .intersects(TypeFlags::Union | TypeFlags::Intersection)
        {
            return value.types().iter().any(|&part| self.array_or_tuple(part));
        }
        self.checker.is_array_type(value) || tsrs_checker::is_tuple_type_exported(value)
    }

    fn array_predicate(&mut self, node: P<Node>) {
        let Some(parent) = node
            .parent()
            .filter(|parent| parent.kind() == Kind::CallExpression)
        else {
            return;
        };
        let call = parent.as_call_expression();
        if call.expression != node {
            return;
        }
        let Some(&callback) = call.arguments.nodes().first() else {
            return;
        };
        let (name, known) = self.checker.get_accessed_property_name(node);
        if !known
            || !matches!(
                name.as_ref(),
                "every" | "filter" | "find" | "findIndex" | "findLast" | "findLastIndex" | "some"
            )
        {
            return;
        }
        let value = self
            .checker
            .get_type_at_location(node.expression().unwrap());
        let value = self
            .checker
            .get_base_constraint_of_type(value)
            .unwrap_or(value);
        if self.array_or_tuple(value) && self.returns_thenable(callback) {
            self.report(
                callback,
                "predicate",
                "Expected a non-Promise value to be returned.",
            );
        }
    }

    fn classify_argument(
        &mut self,
        node: P<Node>,
        value: P<Type>,
        index: usize,
        thenable: &mut Vec<usize>,
        void: &mut Vec<usize>,
    ) {
        let callee = node.expression().unwrap();
        if self.return_thenable_type(callee, value) {
            thenable.push(index);
        } else if self.void_returning(callee, value) && !thenable.contains(&index) {
            void.push(index);
        }
        let contextual = self
            .checker
            .get_contextual_type_for_argument_at_index(node, index);
        if contextual != value {
            self.classify_argument(node, contextual, index, thenable, void);
        }
    }

    fn arguments(&mut self, node: P<Node>) {
        let callee = node.expression().unwrap();
        if node.kind() == Kind::CallExpression && ast::is_access_expression(callee) {
            let (name, ok) = self.checker.get_accessed_property_name(callee);
            if ok && name == "finally" {
                let value = self
                    .checker
                    .get_type_at_location(callee.expression().unwrap());
                let value = self
                    .checker
                    .get_base_constraint_of_type(value)
                    .unwrap_or(value);
                if promise_like(value, self.checker, self.default_library) {
                    return;
                }
            }
        }
        if node.argument_list().is_none() {
            return;
        }
        let arguments = node.arguments();
        let value = self.checker.get_type_at_location(callee);
        let mut thenable = Vec::new();
        let mut void = Vec::new();
        let kind = if node.kind() == Kind::CallExpression {
            SignatureKind::Call
        } else {
            SignatureKind::Construct
        };
        // All overloads matter; the resolved signature prefers early void
        // overloads and would reject legal Promise-returning callbacks.
        for part in parts(value) {
            for &signature in self.checker.get_signatures_of_type(part, kind) {
                for (index, &parameter) in signature.parameters().iter().enumerate() {
                    let Some(value) = self
                        .checker
                        .get_type_of_symbol_at_location(parameter, Some(callee))
                    else {
                        continue;
                    };
                    let rest = parameter.value_declaration().is_some_and(|declaration| {
                        declaration.kind() == Kind::Parameter
                            && declaration
                                .as_parameter_declaration()
                                .dot_dot_dot_token()
                                .is_some()
                    });
                    if !rest {
                        self.classify_argument(node, value, index, &mut thenable, &mut void);
                    } else if self.checker.is_array_type(value) {
                        if let Some(&element) = self.checker.get_type_arguments(value).first() {
                            for index in index..arguments.len() {
                                self.classify_argument(
                                    node,
                                    element,
                                    index,
                                    &mut thenable,
                                    &mut void,
                                );
                            }
                        }
                    } else if tsrs_checker::is_tuple_type_exported(value) {
                        for (offset, &element) in self
                            .checker
                            .get_type_arguments(value)
                            .iter()
                            .enumerate()
                            .take(arguments.len().saturating_sub(index))
                        {
                            self.classify_argument(
                                node,
                                element,
                                index + offset,
                                &mut thenable,
                                &mut void,
                            );
                        }
                    }
                }
            }
        }
        // Mirror the frozen first-occurrence removal rather than silently
        // changing its overload bookkeeping to a set.
        for index in thenable {
            if let Some(at) = void.iter().position(|&candidate| candidate == index) {
                void.remove(at);
            }
        }
        for (index, &argument) in arguments.iter().enumerate() {
            if void.contains(&index) && self.returns_thenable(argument) {
                self.report(
                    argument,
                    "voidReturnArgument",
                    "Promise returned in function argument where a void return was expected.",
                );
            }
        }
    }

    fn jsx_attribute(&mut self, node: P<Node>) {
        let Some(initializer) = node
            .as_jsx_attribute()
            .initializer()
            .filter(|value| value.kind() == Kind::JsxExpression)
        else {
            return;
        };
        let Some(expression) = initializer.as_jsx_expression().expression() else {
            return;
        };
        if let Some(contextual) = self
            .checker
            .get_contextual_type(initializer, ContextFlags::None)
            && self.void_returning(initializer, contextual)
            && self.returns_thenable(expression)
        {
            self.report(initializer, "voidReturnAttribute", "Promise-returning function provided to attribute where a void return was expected.");
        }
    }

    fn inherited(&mut self, node: P<Node>) {
        let clauses = match node.kind() {
            Kind::ClassDeclaration => node.as_class_declaration().heritage_clauses(),
            Kind::ClassExpression => node.as_class_expression().heritage_clauses(),
            Kind::InterfaceDeclaration => node.as_interface_declaration().heritage_clauses(),
            _ => None,
        };
        let Some(clauses) = clauses else { return };
        let mut heritage = Vec::new();
        for &clause in clauses.nodes() {
            for &value in clause.as_heritage_clause().types.get().nodes() {
                heritage.push(self.checker.get_type_at_location(value));
            }
        }
        for &member in node.members() {
            let Some(name) = member.name() else { continue };
            if !matches!(
                name.kind(),
                Kind::Identifier
                    | Kind::PrivateIdentifier
                    | Kind::StringLiteral
                    | Kind::NumericLiteral
                    | Kind::BigIntLiteral
            ) || ast::is_static(member)
                || !self.returns_thenable(member)
            {
                continue;
            }
            for &base in &heritage {
                let inherited = base
                    .symbol()
                    .and_then(|symbol| symbol.members())
                    .and_then(|members| members.lookup(name.text()))
                    .or_else(|| self.checker.get_property_of_type(base, name.text()));
                let Some(inherited) = inherited else { continue };
                let Some(value) = self
                    .checker
                    .get_type_of_symbol_at_location(inherited, Some(member))
                else {
                    continue;
                };
                if self.void_returning(member, value) {
                    let name = self.checker.type_to_string_exported(base);
                    self.report(member, "voidReturnInheritedMethod", format!("Promise-returning method provided where a void return was expected by extended/implemented type '{name}'."));
                }
            }
        }
    }

    fn report_property(&mut self, node: P<Node>, function: bool) {
        let message =
            "Promise-returning function provided to property where a void return was expected.";
        if function {
            if let Some(type_node) = node.type_node() {
                self.report(type_node, "voidReturnProperty", message);
            } else {
                self.violations.push(
                    Violation::new(node, "voidReturnProperty", message).with_byte_range(
                        super::rsbuild_await_checks::function_head_range(node, self.file),
                    ),
                );
            }
        } else {
            self.report(node, "voidReturnProperty", message);
        }
    }

    fn property(&mut self, node: P<Node>) {
        match node.kind() {
            Kind::PropertyAssignment => {
                let initializer = node.as_property_assignment().initializer();
                if let Some(contextual) = self
                    .checker
                    .get_contextual_type(initializer, ContextFlags::None)
                    && self.void_returning(initializer, contextual)
                    && self.returns_thenable(initializer)
                {
                    self.report_property(initializer, ast::is_function_like(initializer));
                }
            }
            Kind::ShorthandPropertyAssignment => {
                let name = node.name().unwrap();
                if let Some(contextual) = self.checker.get_contextual_type(name, ContextFlags::None)
                    && self.void_returning(name, contextual)
                    && self.returns_thenable(name)
                {
                    self.report_property(name, false);
                }
            }
            Kind::MethodDeclaration => {
                let name = node.name().unwrap();
                if name.kind() == Kind::ComputedPropertyName {
                    return;
                }
                let Some(object) = node
                    .parent()
                    .filter(|value| value.kind() == Kind::ObjectLiteralExpression)
                else {
                    return;
                };
                if !self.returns_thenable(node) {
                    return;
                }
                let Some(contextual) = self.checker.get_contextual_type(object, ContextFlags::None)
                else {
                    return;
                };
                let Some(symbol) = parts(contextual)
                    .into_iter()
                    .find_map(|part| self.checker.get_property_of_type(part, name.text()))
                else {
                    return;
                };
                if let Some(value) = self
                    .checker
                    .get_type_of_symbol_at_location(symbol, Some(name))
                    && self.void_returning(name, value)
                {
                    self.report_property(node, true);
                }
            }
            _ => {}
        }
    }

    fn returned(&mut self, node: P<Node>) {
        let Some(expression) = node.expression() else {
            return;
        };
        let mut function = node.parent();
        while let Some(current) = function {
            if ast::is_function_like(current) {
                break;
            }
            function = current.parent();
        }
        let Some(function) = function else { return };
        if function
            .type_node()
            .is_some_and(|value| !possibly_function(value))
        {
            return;
        }
        if let Some(contextual) = self
            .checker
            .get_contextual_type(expression, ContextFlags::None)
            && self.void_returning(expression, contextual)
            && self.returns_thenable(expression)
        {
            self.report(expression, "voidReturnReturnValue", "Promise-returning function provided to return value where a void return was expected.");
        }
    }

    fn assignment(&mut self, node: P<Node>) {
        let expression = node.as_binary_expression();
        let value = self.checker.get_type_at_location(expression.left);
        if self.void_returning(expression.left, value)
            && self.returns_thenable(expression.right.get())
        {
            self.report(
                expression.right.get(),
                "voidReturnVariable",
                "Promise-returning function provided to variable where a void return was expected.",
            );
        }
    }

    fn variable(&mut self, node: P<Node>) {
        let Some(initializer) = node.initializer() else {
            return;
        };
        let Some(annotation) = node.type_node() else {
            return;
        };
        if !possibly_function(annotation) {
            return;
        }
        let value = self.checker.get_type_at_location(node.name().unwrap());
        if self.void_returning(initializer, value) && self.returns_thenable(initializer) {
            self.report(
                initializer,
                "voidReturnVariable",
                "Promise-returning function provided to variable where a void return was expected.",
            );
        }
    }

    fn walk_misused(&mut self, node: P<Node>) {
        match node.kind() {
            Kind::BinaryExpression => {
                self.conditional(Some(node), false);
                if ast::is_assignment_expression(node, false) {
                    self.assignment(node);
                }
            }
            Kind::PropertyAccessExpression | Kind::ElementAccessExpression => {
                self.array_predicate(node)
            }
            Kind::PrefixUnaryExpression => {
                let expression = node.as_prefix_unary_expression();
                if expression.operator == Kind::ExclamationToken {
                    self.conditional(Some(expression.operand), true);
                }
            }
            Kind::ConditionalExpression => {
                self.conditional(Some(node.as_conditional_expression().condition), true)
            }
            Kind::ForStatement => self.conditional(node.as_for_statement().condition(), true),
            Kind::DoStatement | Kind::WhileStatement | Kind::IfStatement => {
                self.conditional(node.expression(), true)
            }
            Kind::CallExpression | Kind::NewExpression => self.arguments(node),
            Kind::JsxAttribute => self.jsx_attribute(node),
            Kind::ClassDeclaration | Kind::ClassExpression | Kind::InterfaceDeclaration => {
                self.inherited(node)
            }
            Kind::PropertyAssignment
            | Kind::MethodDeclaration
            | Kind::ShorthandPropertyAssignment => self.property(node),
            Kind::ReturnStatement => self.returned(node),
            Kind::VariableDeclaration => self.variable(node),
            Kind::SpreadElement | Kind::SpreadAssignment => {
                let expression = node.expression().unwrap();
                let value = self.checker.get_type_at_location(expression);
                if self.thenable(expression, value) {
                    self.report(
                        expression,
                        "spread",
                        "Expected a non-Promise value to be spread in an object.",
                    );
                }
            }
            _ => {}
        }
        node.for_each_child(&mut |child| {
            self.walk_misused(child);
            false
        });
    }
}

fn possibly_function(node: P<Node>) -> bool {
    match node.kind() {
        Kind::ConditionalType
        | Kind::ConstructorType
        | Kind::FunctionType
        | Kind::ImportType
        | Kind::IndexedAccessType
        | Kind::InferType
        | Kind::IntersectionType
        | Kind::QualifiedName
        | Kind::ThisType
        | Kind::TypeOperator
        | Kind::TypeQuery
        | Kind::TypeReference
        | Kind::UnionType => true,
        Kind::TypeLiteral => node.members().iter().any(|member| {
            matches!(
                member.kind(),
                Kind::CallSignature | Kind::ConstructSignature
            )
        }),
        _ => false,
    }
}
