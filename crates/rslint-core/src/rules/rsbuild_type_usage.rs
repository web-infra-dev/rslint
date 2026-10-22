//! Frozen rslint e0a29f751d4def86d27931759098f27748fa3003 default typed rules.
//! Copyright (c) 2025 July-present Bytedance Inc and its affiliates.
//! Copyright (c) 2025 July typescript-eslint and other contributors.
//! MIT licensed; see LICENSE. The builtin declaration checks and method
//! provenance operate on the current tsrs Program, including imported symbols.

use tsrs_ast::{self as ast, Kind, Node, SourceFile, Symbol, SymbolFlags};
use tsrs_checker::{Checker, SignatureKind, Type, TypeFlags};
use tsrs_core::P;

use crate::Violation;

pub(crate) const NAMES: &[&str] = &[
    "@typescript-eslint/no-base-to-string",
    "@typescript-eslint/no-implied-eval",
    "@typescript-eslint/unbound-method",
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
        violations: Vec::new(),
    };
    match name {
        "@typescript-eslint/no-base-to-string" => context.base_to_string(node),
        "@typescript-eslint/no-implied-eval" => context.implied_eval(node),
        "@typescript-eslint/unbound-method" => context.unbound(node),
        _ => {}
    }
    context.violations
}

fn union(value: P<Type>) -> Vec<P<Type>> {
    if value.flags().intersects(TypeFlags::Union) {
        value.types().to_vec()
    } else {
        vec![value]
    }
}
fn intersection(value: P<Type>) -> Vec<P<Type>> {
    if value.flags().intersects(TypeFlags::Intersection) {
        value.types().to_vec()
    } else {
        vec![value]
    }
}

/// `None` matches any active builtin; `Some` matches the listed builtin names.
fn builtin(
    value: P<Type>,
    names: Option<&[&str]>,
    checker: &mut Checker,
    default_library: &dyn Fn(P<SourceFile>) -> bool,
) -> bool {
    if value.flags().intersects(TypeFlags::Intersection) {
        return value
            .types()
            .iter()
            .any(|&part| builtin(part, names, checker, default_library));
    }
    if value.flags().intersects(TypeFlags::Union) {
        return value
            .types()
            .iter()
            .all(|&part| builtin(part, names, checker, default_library));
    }
    if value.flags().intersects(TypeFlags::TypeParameter) {
        return checker
            .get_base_constraint_of_type(value)
            .is_some_and(|constraint| builtin(constraint, names, checker, default_library));
    }
    let Some(symbol) = value.symbol() else {
        return false;
    };
    if names.is_none_or(|names| names.contains(&symbol.name()))
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
            .any(|&base| builtin(base, names, checker, default_library));
    }
    false
}

fn type_name(value: P<Type>, checker: &mut Checker) -> String {
    if value.flags().intersects(TypeFlags::StringLike) {
        return "string".into();
    }
    if value.flags().intersects(TypeFlags::TypeParameter)
        && let Some(declaration) = value
            .symbol()
            .and_then(|symbol| symbol.declarations().first().copied())
        && declaration.kind() == Kind::TypeParameter
        && let Some(constraint) = declaration.as_type_parameter_declaration().constraint()
    {
        let constraint = checker.get_type_from_type_node(constraint);
        return type_name(constraint, checker);
    }
    if value.flags().intersects(TypeFlags::Union)
        && value
            .types()
            .iter()
            .all(|&part| type_name(part, checker) == "string")
        || value.flags().intersects(TypeFlags::Intersection)
            && value
                .types()
                .iter()
                .any(|&part| type_name(part, checker) == "string")
    {
        return "string".into();
    }
    checker.type_to_string_exported(value)
}

#[derive(Clone, Copy, PartialEq, Eq)]
enum Usefulness {
    Always,
    Never,
    Sometimes,
}

struct Context<'a> {
    file: P<SourceFile>,
    checker: &'a mut Checker,
    default_library: &'a dyn Fn(P<SourceFile>) -> bool,
    violations: Vec<Violation>,
}

impl Context<'_> {
    fn certainty(&mut self, value: P<Type>, visited: &[P<Type>], join: bool) -> Usefulness {
        if !join {
            if visited.contains(&value) {
                return Usefulness::Always;
            }
            if value.flags().intersects(TypeFlags::TypeParameter) {
                return self
                    .checker
                    .get_base_constraint_of_type(value)
                    .map(|constraint| self.certainty(constraint, visited, false))
                    .unwrap_or(Usefulness::Always);
            }
            if value.flags().intersects(TypeFlags::BooleanLike)
                || ["Error", "RegExp", "URL", "URLSearchParams"]
                    .contains(&type_name(value, self.checker).as_str())
            {
                return Usefulness::Always;
            }
        }
        if value.flags().intersects(TypeFlags::Intersection) {
            return if value
                .types()
                .iter()
                .any(|&part| self.certainty(part, visited, join) == Usefulness::Always)
            {
                Usefulness::Always
            } else {
                Usefulness::Never
            };
        }
        if value.flags().intersects(TypeFlags::Union) {
            let all = value
                .types()
                .iter()
                .map(|&part| self.certainty(part, visited, join))
                .collect::<Vec<_>>();
            return if all.iter().all(|&part| part == Usefulness::Never) {
                Usefulness::Never
            } else if all.iter().all(|&part| part == Usefulness::Always) {
                Usefulness::Always
            } else {
                Usefulness::Sometimes
            };
        }
        if tsrs_checker::is_tuple_type_exported(value) {
            let mut next = visited.to_vec();
            if !join {
                next.push(value);
            }
            let all = self
                .checker
                .get_type_arguments(value)
                .iter()
                .map(|&part| self.certainty(part, &next, false))
                .collect::<Vec<_>>();
            return if all.contains(&Usefulness::Never) {
                Usefulness::Never
            } else if all.contains(&Usefulness::Sometimes) {
                Usefulness::Sometimes
            } else {
                Usefulness::Always
            };
        }
        if self.checker.is_array_type(value) {
            let mut next = visited.to_vec();
            if !join {
                next.push(value);
            }
            let number = self.checker.get_number_type();
            let element = self
                .checker
                .get_index_type_of_type(value, number)
                .expect("array has number index type");
            return self.certainty(element, &next, false);
        }
        if join {
            return Usefulness::Always;
        }
        let symbol = self
            .checker
            .get_property_of_type(value, "toString")
            .or_else(|| self.checker.get_property_of_type(value, "toLocaleString"));
        let Some(symbol) = symbol else {
            return Usefulness::Always;
        };
        let declarations = symbol.declarations();
        if declarations.len() != 1 {
            return Usefulness::Always;
        }
        if declarations[0].parent().is_some_and(|parent| {
            parent.kind() == Kind::InterfaceDeclaration
                && parent.name().is_some_and(|name| name.text() == "Object")
        }) {
            Usefulness::Never
        } else {
            Usefulness::Always
        }
    }

    fn stringify(&mut self, node: P<Node>, value: Option<P<Type>>, join: bool) {
        if !join && ast::is_literal_expression(node) {
            return;
        }
        let value = value.unwrap_or_else(|| self.checker.get_type_at_location(node));
        if self.certainty(value, &[], join) == Usefulness::Always {
            return;
        }
        // Preserve the frozen message's full node text (including leading trivia)
        // and its certaintyToString behavior, which returns "always" for every
        // enum variant. Reporting spans are independently trimmed by the engine.
        let text = &self.file.text()[node.pos().expect_source("node start") as usize
            ..node.end().expect_source("node end") as usize];
        let (id, message) = if join {
            (
                "baseArrayJoin",
                format!(
                    "Using `join()` for {text} always use Object's default stringification format ('[object Object]') when stringified."
                ),
            )
        } else {
            (
                "baseToString",
                format!(
                    "'{text}' always use Object's default stringification format ('[object Object]') when stringified."
                ),
            )
        };
        self.violations.push(Violation::new(node, id, message));
    }

    fn base_to_string(&mut self, node: P<Node>) {
        match node.kind() {
            Kind::BinaryExpression => {
                let binary = node.as_binary_expression();
                if !matches!(
                    binary.operator_token.kind(),
                    Kind::PlusToken | Kind::PlusEqualsToken
                ) {
                    return;
                }
                let left = self.checker.get_type_at_location(binary.left);
                let right = self.checker.get_type_at_location(binary.right.get());
                if type_name(left, self.checker) == "string" {
                    self.stringify(binary.right.get(), Some(right), false);
                } else if type_name(right, self.checker) == "string"
                    && binary.left.kind() != Kind::PrivateIdentifier
                {
                    self.stringify(binary.left, Some(left), false);
                }
            }
            Kind::CallExpression => {
                let call = node.as_call_expression();
                if call.expression.kind() == Kind::Identifier
                    && call.expression.text() == "String"
                    && let Some(&argument) = call.arguments.nodes().first()
                {
                    let value = self.checker.get_type_at_location(call.expression);
                    if builtin(
                        value,
                        Some(&["String", "StringConstructor"]),
                        self.checker,
                        self.default_library,
                    ) && argument.kind() != Kind::SpreadElement
                    {
                        self.stringify(argument, None, false);
                        return;
                    }
                }
                if call.expression.kind() == Kind::PropertyAccessExpression {
                    let member = call.expression.as_property_access_expression();
                    match member.name().text() {
                        "join" => {
                            let value = self.checker.get_type_at_location(member.expression);
                            let value = self
                                .checker
                                .get_base_constraint_of_type(value)
                                .unwrap_or(value);
                            self.stringify(member.expression, Some(value), true);
                        }
                        "toString" | "toLocaleString" => {
                            self.stringify(member.expression, None, false)
                        }
                        _ => {}
                    }
                }
            }
            Kind::TemplateExpression => {
                if node
                    .parent()
                    .is_some_and(|parent| parent.kind() == Kind::TaggedTemplateExpression)
                {
                    return;
                }
                for &span in node.as_template_expression().template_spans.nodes() {
                    self.stringify(span.as_template_span().expression, None, false);
                }
            }
            _ => {}
        }
    }

    fn function(&mut self, node: P<Node>) -> bool {
        if ast::is_function_like(node) {
            return true;
        }
        if ast::is_literal_expression(node) {
            return false;
        }
        if node.kind() == Kind::CallExpression {
            let mut callee = node.as_call_expression().expression;
            if callee.kind() == Kind::PropertyAccessExpression {
                callee = callee.name().unwrap();
            }
            if callee.kind() == Kind::Identifier && callee.text() == "bind" {
                return true;
            }
        }
        let value = self.checker.get_type_at_location(node);
        value.symbol().is_some_and(|symbol| {
            symbol
                .flags()
                .intersects(SymbolFlags::Function | SymbolFlags::Method)
        }) || builtin(
            value,
            Some(&["Function"]),
            self.checker,
            self.default_library,
        ) || !self
            .checker
            .get_signatures_of_type(value, SignatureKind::Call)
            .is_empty()
    }

    fn implied_eval(&mut self, node: P<Node>) {
        if !matches!(node.kind(), Kind::CallExpression | Kind::NewExpression) {
            return;
        }
        let callee = node.expression().unwrap();
        let name = if callee.kind() == Kind::Identifier {
            callee.text()
        } else if ast::is_access_expression(callee)
            && callee.expression().is_some_and(|object| {
                object.kind() == Kind::Identifier
                    && ["global", "globalThis", "window"].contains(&object.text())
            })
        {
            if callee.kind() == Kind::PropertyAccessExpression {
                let name = callee.name().unwrap();
                if name.kind() != Kind::Identifier {
                    return;
                }
                name.text()
            } else {
                let argument = callee.as_element_access_expression().argument_expression;
                if argument.kind() != Kind::StringLiteral {
                    return;
                }
                argument.text()
            }
        } else {
            return;
        };
        if name == "Function" {
            let value = self.checker.get_type_at_location(callee);
            if value.symbol().is_none()
                || builtin(
                    value,
                    Some(&["FunctionConstructor"]),
                    self.checker,
                    self.default_library,
                )
            {
                self.violations.push(Violation::new(
                    node,
                    "noFunctionConstructor",
                    "Implied eval. Do not use the Function constructor to create functions.",
                ));
                return;
            }
        }
        let Some(&handler) = node.arguments().first() else {
            return;
        };
        if ["execScript", "setImmediate", "setInterval", "setTimeout"].contains(&name)
            && !self.function(handler)
        {
            let symbol = self.checker.get_symbol_at_location_exported(callee);
            if symbol.is_none_or(|symbol| {
                !symbol.declarations().iter().any(|&declaration| {
                    ast::get_source_file_of_node(declaration) == Some(self.file)
                })
            }) {
                self.violations.push(Violation::new(
                    handler,
                    "noImpliedEvalError",
                    "Implied eval. Consider passing a function.",
                ));
            }
        }
    }

    fn native_bound(&mut self, object: P<Node>, property: P<Node>) -> bool {
        if object.kind() == Kind::Identifier && property.kind() == Kind::Identifier {
            let symbol = if property.parent().is_some_and(heritage_qualified) {
                self.heritage_symbol(object)
            } else {
                self.checker.get_symbol_at_location_exported(object)
            };
            let nonlocal = symbol
                .and_then(|symbol| symbol.value_declaration())
                .is_some_and(|declaration| {
                    ast::get_source_file_of_node(declaration) != Some(self.file)
                });
            if nonlocal && natively_bound_member(object.text(), property.text()) {
                return true;
            }
        }
        let object_type = self.checker.get_type_at_location(object);
        let property_type = self.checker.get_type_at_location(property);
        builtin(
            object_type,
            Some(&[
                "NumberConstructor",
                "ObjectConstructor",
                "StringConstructor",
                "SymbolConstructor",
                "ArrayConstructor",
                "Array",
                "ProxyConstructor",
                "Console",
                "DateConstructor",
                "Atomics",
                "Math",
                "JSON",
            ]),
            self.checker,
            self.default_library,
        ) && builtin(property_type, None, self.checker, self.default_library)
    }

    fn heritage_symbol(&mut self, node: P<Node>) -> Option<P<Symbol>> {
        if node.kind() == Kind::Identifier {
            return self.checker.resolve_name_exported(
                node.text(),
                Some(node),
                SymbolFlags::Value,
                false,
            );
        }
        if node.kind() != Kind::QualifiedName {
            return None;
        }
        let name = node.as_qualified_name();
        let object = self.heritage_symbol(name.left)?;
        let value = self.checker.get_type_of_symbol(object);
        self.checker.get_property_of_type(value, name.right.text())
    }

    fn report_method(&mut self, node: P<Node>, symbol: Option<P<Symbol>>) -> bool {
        let Some(declaration) = symbol.and_then(|symbol| symbol.value_declaration()) else {
            return false;
        };
        let (dangerous, first_this) = match declaration.kind() {
            Kind::PropertyDeclaration => (
                declaration
                    .initializer()
                    .is_some_and(|value| value.kind() == Kind::FunctionExpression),
                true,
            ),
            Kind::PropertyAssignment => {
                let initializer = declaration.initializer().unwrap();
                if initializer.kind() != Kind::FunctionExpression {
                    return false;
                }
                method(initializer)
            }
            Kind::MethodDeclaration | Kind::MethodSignature => method(declaration),
            _ => (false, false),
        };
        if !dangerous {
            return false;
        }
        let (id, message) = if first_this {
            (
                "unbound",
                "Avoid referencing unbound methods which may cause unintentional scoping of `this`.",
            )
        } else {
            (
                "unboundWithoutThisAnnotation",
                "Avoid referencing unbound methods which may cause unintentional scoping of `this`.\nIf your function does not access `this`, you can annotate it with `this: void`, or consider using an arrow function instead.",
            )
        };
        self.violations.push(Violation::new(node, id, message));
        true
    }

    fn binding_type(&mut self, property: P<Node>, value: P<Type>) -> bool {
        if value
            .flags()
            .intersects(TypeFlags::Union | TypeFlags::Intersection)
        {
            return value
                .types()
                .iter()
                .any(|&part| self.binding_type(property, part));
        }
        let symbol = self.checker.get_property_of_type(value, property.text());
        self.report_method(property, symbol)
    }

    fn binding(
        &mut self,
        pattern: P<Node>,
        initializer: Option<P<Node>>,
        property: P<Node>,
        assignment: bool,
    ) {
        if let Some(initializer) = initializer {
            if !self.native_bound(initializer, property) {
                let value = self.checker.get_type_at_location(initializer);
                let symbol = self.checker.get_property_of_type(value, property.text());
                if self.report_method(property, symbol) {
                    return;
                }
            } else if !assignment {
                return;
            }
        }
        let value = self.checker.get_type_at_location(pattern);
        self.binding_type(property, value);
    }

    fn member_property(&mut self, node: P<Node>, object_type: P<Type>, name: &str) -> bool {
        for union_part in union(object_type) {
            for intersection_part in intersection(union_part) {
                let symbol = self.checker.get_property_of_type(intersection_part, name);
                if self.report_method(node, symbol) {
                    return true;
                }
            }
        }
        false
    }

    fn member(&mut self, node: P<Node>) {
        let (object, property) = member_parts(node);
        if safe_use(node) || self.native_bound(object, property) {
            return;
        }
        let object_type = self.checker.get_type_at_location(object);
        if node.kind() == Kind::ElementAccessExpression {
            let key = self.checker.get_type_at_location(property);
            for key in union(key) {
                if key
                    .flags()
                    .intersects(TypeFlags::StringLiteral | TypeFlags::NumberLiteral)
                {
                    let name = tsrs_checker::get_property_name_from_type_exported(key);
                    if self.member_property(node, object_type, &name) {
                        break;
                    }
                }
            }
        } else if property.kind() == Kind::Identifier {
            self.member_property(node, object_type, property.text());
        } else {
            let symbol = self.checker.get_symbol_at_location_exported(node);
            self.report_method(node, symbol);
        }
    }

    fn unbound(&mut self, node: P<Node>) {
        match node.kind() {
            Kind::PropertyAccessExpression | Kind::ElementAccessExpression => self.member(node),
            Kind::QualifiedName if heritage_qualified(node) => {
                let (object, property) = member_parts(node);
                if safe_use(node) || self.native_bound(object, property) {
                    return;
                }
                let symbol = self.heritage_symbol(node);
                self.report_method(node, symbol);
            }
            Kind::ObjectLiteralExpression => {
                let Some(parent) = node
                    .parent()
                    .filter(|&parent| ast::is_assignment_expression(parent, true))
                else {
                    return;
                };
                let initializer = parent.as_binary_expression().right.get();
                for &property in node.as_object_literal_expression().properties.nodes() {
                    if matches!(
                        property.kind(),
                        Kind::PropertyAssignment | Kind::ShorthandPropertyAssignment
                    ) && let Some(name) = property
                        .name()
                        .filter(|name| name.kind() == Kind::Identifier)
                    {
                        self.binding(node, Some(initializer), name, true);
                    }
                }
            }
            Kind::ObjectBindingPattern => {
                if inside_type_declaration(node) {
                    return;
                }
                let Some(parent) = node.parent() else { return };
                let assignment = matches!(parent.kind(), Kind::BindingElement | Kind::Parameter);
                let initializer = if parent.kind() == Kind::VariableDeclaration || assignment {
                    parent.initializer()
                } else {
                    None
                };
                for &property in node.as_binding_pattern().elements.nodes() {
                    if property.kind() != Kind::BindingElement {
                        continue;
                    }
                    let element = property.as_binding_element();
                    let name = element.property_name().or_else(|| element.name()).unwrap();
                    if element.dot_dot_dot_token().is_some() || name.kind() != Kind::Identifier {
                        continue;
                    }
                    self.binding(node, initializer, name, assignment);
                }
            }
            _ => {}
        }
    }
}

fn method(node: P<Node>) -> (bool, bool) {
    let this = ast::get_this_parameter(node);
    let void = this.is_some_and(|parameter| {
        ast::is_this_parameter(parameter)
            && parameter
                .type_node()
                .is_some_and(|value| value.kind() == Kind::VoidKeyword)
    });
    (!void, this.is_some())
}

fn modifier(node: P<Node>, wanted: Kind) -> bool {
    node.modifiers().is_some_and(|modifiers| {
        modifiers
            .nodes()
            .iter()
            .any(|modifier| modifier.kind() == wanted)
    })
}

fn inside_type_declaration(node: P<Node>) -> bool {
    let mut parent = node.parent();
    while let Some(node) = parent {
        match node.kind() {
            Kind::ClassDeclaration | Kind::VariableStatement
                if modifier(node, Kind::DeclareKeyword) =>
            {
                return true;
            }
            Kind::MethodDeclaration if modifier(node, Kind::AbstractKeyword) => return true,
            Kind::FunctionDeclaration if node.body().is_none() => return true,
            Kind::FunctionType | Kind::InterfaceDeclaration | Kind::TypeAliasDeclaration => {
                return true;
            }
            _ => {}
        }
        parent = node.parent();
    }
    false
}

fn heritage_qualified(node: P<Node>) -> bool {
    if node.kind() != Kind::QualifiedName {
        return false;
    }
    let mut entity = node;
    while let Some(parent) = entity.parent().filter(|parent| {
        parent.kind() == Kind::QualifiedName && parent.as_qualified_name().left == entity
    }) {
        entity = parent;
    }
    entity.parent().is_some_and(|reference| {
        reference.kind() == Kind::TypeReference
            && reference.as_type_reference_node().type_name == entity
            && reference
                .parent()
                .is_some_and(|parent| parent.kind() == Kind::HeritageClause)
    })
}

fn member_parts(node: P<Node>) -> (P<Node>, P<Node>) {
    match node.kind() {
        Kind::PropertyAccessExpression => (node.expression().unwrap(), node.name().unwrap()),
        Kind::ElementAccessExpression => (
            node.expression().unwrap(),
            node.as_element_access_expression().argument_expression,
        ),
        Kind::QualifiedName => {
            let name = node.as_qualified_name();
            (name.left, name.right)
        }
        _ => unreachable!("member expression"),
    }
}

fn safe_use(mut node: P<Node>) -> bool {
    while let Some(parent) = node.parent() {
        match parent.kind() {
            Kind::ParenthesizedExpression
            | Kind::NonNullExpression
            | Kind::AsExpression
            | Kind::TypeAssertionExpression => {}
            Kind::QualifiedName => return heritage_qualified(parent),
            Kind::IfStatement
            | Kind::ForStatement
            | Kind::PropertyAccessExpression
            | Kind::ElementAccessExpression
            | Kind::SwitchStatement
            | Kind::WhileStatement => return true,
            Kind::PostfixUnaryExpression => {
                return matches!(
                    parent.as_postfix_unary_expression().operator,
                    Kind::PlusPlusToken | Kind::MinusMinusToken
                );
            }
            Kind::PrefixUnaryExpression => {
                return matches!(
                    parent.as_prefix_unary_expression().operator,
                    Kind::PlusPlusToken | Kind::MinusMinusToken | Kind::ExclamationToken
                );
            }
            Kind::CallExpression => return parent.expression() == Some(node),
            Kind::ConditionalExpression => {
                return parent.as_conditional_expression().condition == node;
            }
            Kind::TaggedTemplateExpression => {
                return parent.as_tagged_template_expression().tag == node;
            }
            Kind::DeleteExpression | Kind::TypeOfExpression | Kind::VoidExpression => return true,
            Kind::BinaryExpression => {
                let binary = parent.as_binary_expression();
                let operator = binary.operator_token.kind();
                if operator == Kind::AmpersandAmpersandToken {
                    if binary.left == node {
                        return true;
                    }
                } else if matches!(
                    operator,
                    Kind::ExclamationEqualsToken
                        | Kind::ExclamationEqualsEqualsToken
                        | Kind::EqualsEqualsToken
                        | Kind::EqualsEqualsEqualsToken
                        | Kind::InstanceOfKeyword
                ) {
                    return true;
                } else if ast::is_logical_binary_operator(operator) {
                } else if ast::is_assignment_expression(parent, true) {
                    return binary.left == node
                        || ast::is_access_expression(node)
                            && node
                                .expression()
                                .is_some_and(|object| object.kind() == Kind::SuperKeyword)
                            && ast::is_access_expression(binary.left)
                            && binary
                                .left
                                .expression()
                                .is_some_and(|object| object.kind() == Kind::ThisKeyword);
                } else {
                    return false;
                }
            }
            _ => return false,
        }
        node = parent;
    }
    false
}

// Frozen generated native member table; kept independent of the running Node
// version so changing host globals cannot change lint diagnostics.
fn natively_bound_member(object: &str, property: &str) -> bool {
    let members: &[&str] = match object {
        "Number" => &[
            "isFinite",
            "isInteger",
            "isNaN",
            "isSafeInteger",
            "parseFloat",
            "parseInt",
        ],
        "Object" => &[
            "assign",
            "getOwnPropertyDescriptor",
            "getOwnPropertyDescriptors",
            "getOwnPropertyNames",
            "getOwnPropertySymbols",
            "hasOwn",
            "is",
            "preventExtensions",
            "seal",
            "create",
            "defineProperties",
            "defineProperty",
            "freeze",
            "getPrototypeOf",
            "setPrototypeOf",
            "isExtensible",
            "isFrozen",
            "isSealed",
            "keys",
            "entries",
            "fromEntries",
            "values",
            "groupBy",
        ],
        "String" => &["fromCharCode", "fromCodePoint", "raw"],
        "Symbol" => &["for", "keyFor"],
        "Array" => &["isArray", "from", "of", "fromAsync"],
        "Proxy" => &["revocable"],
        "Date" => &["now", "parse", "UTC"],
        "Atomics" => &[
            "load",
            "store",
            "add",
            "sub",
            "and",
            "or",
            "xor",
            "exchange",
            "compareExchange",
            "isLockFree",
            "wait",
            "waitAsync",
            "notify",
        ],
        "Reflect" => &[
            "defineProperty",
            "deleteProperty",
            "apply",
            "construct",
            "get",
            "getOwnPropertyDescriptor",
            "getPrototypeOf",
            "has",
            "isExtensible",
            "ownKeys",
            "preventExtensions",
            "set",
            "setPrototypeOf",
        ],
        "console" => &[
            "log",
            "info",
            "debug",
            "warn",
            "error",
            "dir",
            "time",
            "timeEnd",
            "timeLog",
            "trace",
            "assert",
            "clear",
            "count",
            "countReset",
            "group",
            "groupEnd",
            "table",
            "dirxml",
            "groupCollapsed",
            "Console",
            "profile",
            "profileEnd",
            "timeStamp",
            "context",
            "createTask",
        ],
        "Math" => &[
            "abs", "acos", "acosh", "asin", "asinh", "atan", "atanh", "atan2", "ceil", "cbrt",
            "expm1", "clz32", "cos", "cosh", "exp", "floor", "fround", "hypot", "imul", "log",
            "log1p", "log2", "log10", "max", "min", "pow", "random", "round", "sign", "sin",
            "sinh", "sqrt", "tan", "tanh", "trunc",
        ],
        "JSON" => &["parse", "stringify", "rawJSON", "isRawJSON"],
        "Intl" => &[
            "getCanonicalLocales",
            "supportedValuesOf",
            "DateTimeFormat",
            "NumberFormat",
            "Collator",
            "PluralRules",
            "RelativeTimeFormat",
            "ListFormat",
            "Locale",
            "DisplayNames",
            "Segmenter",
        ],
        _ => &[],
    };
    members.contains(&property)
}
