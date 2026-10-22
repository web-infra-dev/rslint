//! Assignment checks ported from frozen rslint e0a29f751d4def86d27931759098f27748fa3003.
//! Bindings are compared by compiler symbol/declaration identity, never by an
//! identifier spelling alone. No rule in this batch offers fixes or suggestions.
use super::{
    rsbuild_core_checks::write_reference,
    rsbuild_declaration_checks::{
        access, in_destructuring_assignment, static_expression, static_property,
    },
};
use crate::Violation;
use std::collections::HashMap;
use tsrs_ast::{
    self as ast, Kind, Node, NodeFlags, OuterExpressionKinds, SourceFile, Symbol, SymbolFlags,
};
use tsrs_checker::Checker;
use tsrs_core::P;

pub(crate) const NAMES: &[&str] = &[
    "no-class-assign",
    "no-const-assign",
    "no-ex-assign",
    "no-func-assign",
    "no-import-assign",
    "no-self-assign",
];
pub(crate) fn check(
    name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
) -> Vec<Violation> {
    match (name, node.kind()) {
        ("no-self-assign", Kind::BinaryExpression) => self_assign(node, file),
        ("no-const-assign", Kind::Identifier) if write_reference(node) => {
            let Some(symbol) = resolve_reference(node, checker) else {
                return Vec::new();
            };
            let constant = symbol.declarations().iter().any(|&declaration| {
                ast::get_source_file_of_node(declaration) == Some(file)
                    && declaration_list(declaration)
                        .is_some_and(|list| list.flags().intersects(NodeFlags::Constant))
            });
            if constant {
                vec![Violation::new(
                    node,
                    "const",
                    format!("'{}' is constant.", node.text()),
                )]
            } else {
                Vec::new()
            }
        }
        (
            "no-class-assign" | "no-func-assign" | "no-ex-assign" | "no-import-assign",
            Kind::SourceFile,
        ) => binding_writes(name, node, file, checker),
        _ => Vec::new(),
    }
}
fn walk(node: P<Node>, visitor: &mut impl FnMut(P<Node>)) {
    if node.flags().intersects(NodeFlags::Reparsed) {
        return;
    }
    visitor(node);
    node.for_each_child(&mut |child| {
        walk(child, visitor);
        false
    });
}
fn declaration_list(mut node: P<Node>) -> Option<P<Node>> {
    loop {
        match node.kind() {
            Kind::VariableDeclarationList => return Some(node),
            Kind::VariableDeclaration
            | Kind::BindingElement
            | Kind::ObjectBindingPattern
            | Kind::ArrayBindingPattern => node = node.parent()?,
            _ => return None,
        }
    }
}
fn binding_symbol(name: P<Node>) -> Option<P<Symbol>> {
    name.parent()
        .filter(|parent| parent.name() == Some(name))
        .and_then(|parent| parent.symbol())
}
fn same_symbol(left: P<Symbol>, right: P<Symbol>) -> bool {
    left == right
        || left
            .declarations()
            .iter()
            .any(|declaration| right.declarations().contains(declaration))
}
fn authored(node: P<Node>) -> bool {
    let mut current = Some(node);
    while let Some(node) = current {
        if node.flags().intersects(NodeFlags::Reparsed) {
            return false;
        }
        current = node.parent();
    }
    true
}
fn symbol_has_runtime_binding(symbol: P<Symbol>) -> bool {
    symbol.flags().intersects(
        SymbolFlags::Value | SymbolFlags::Namespace | SymbolFlags::Alias | SymbolFlags::ExportValue,
    ) && symbol.declarations().iter().any(|&node| {
        authored(node) && !matches!(node.kind(), Kind::ExportSpecifier | Kind::NamespaceExport)
    })
}
// Traverse the binder's actual scope tables before invoking the checker. This
// avoids synthesized JSDoc bindings and function-body declarations leaking into
// parameter initializers. Those are explicit authored-scope corrections in the
// frozen RefStore, particularly for no-import-assign on JavaScript files.
pub(crate) fn resolve_reference(node: P<Node>, checker: &mut Checker) -> Option<P<Symbol>> {
    let name = node.text();
    let mut child = node;
    while let Some(scope) = child.parent() {
        let parameter_scope =
            ast::is_function_like_declaration(scope) && scope.parameters().contains(&child);
        if parameter_scope {
            for &parameter in scope.parameters() {
                let mut names = Vec::new();
                if let Some(binding) = parameter.name() {
                    collect_binding(binding, &mut names);
                }
                if let Some(&name_node) = names.iter().find(|candidate| candidate.text() == name) {
                    return binding_symbol(name_node);
                }
            }
        }
        if let Some(symbol) = scope.locals().and_then(|locals| locals.get().get(name)) {
            let body_only = parameter_scope
                && scope.body().is_some_and(|body| {
                    !symbol.declarations().is_empty()
                        && symbol.declarations().iter().all(|declaration| {
                            declaration.pos() >= body.pos() && declaration.end() <= body.end()
                        })
                });
            if !body_only && symbol_has_runtime_binding(symbol) {
                return Some(symbol);
            }
        }
        if matches!(
            scope.kind(),
            Kind::FunctionExpression | Kind::ClassExpression
        ) && scope.name().is_some_and(|binding| binding.text() == name)
            && !(scope.kind() == Kind::ClassExpression && child.kind() == Kind::Decorator)
        {
            return scope.symbol();
        }
        // A JSDoc @import may occupy the table slot of a real source import.
        // The authored import's binder symbol remains on its declaration.
        if scope.kind() == Kind::SourceFile {
            for &statement in scope.statements() {
                if statement.kind() == Kind::ImportDeclaration {
                    for (binding, _) in import_bindings(statement) {
                        if binding.text() == name {
                            return binding_symbol(binding);
                        }
                    }
                }
            }
        }
        child = scope;
    }
    checker.resolve_name_exported(
        name,
        node,
        SymbolFlags::Value | SymbolFlags::Namespace | SymbolFlags::Alias,
        false,
    )
}
struct Binding {
    name: String,
    symbol: P<Symbol>,
    namespace: bool,
}
fn collect_binding(node: P<Node>, out: &mut Vec<P<Node>>) {
    match node.kind() {
        Kind::Identifier => out.push(node),
        Kind::ObjectBindingPattern | Kind::ArrayBindingPattern => {
            for &element in node.elements() {
                if element.kind() == Kind::BindingElement
                    && let Some(name) = element.as_binding_element().name()
                {
                    collect_binding(name, out);
                }
            }
        }
        _ => {}
    }
}
fn import_bindings(node: P<Node>) -> Vec<(P<Node>, bool)> {
    let mut out = Vec::new();
    let Some(clause) = node.as_import_declaration().import_clause() else {
        return out;
    };
    if let Some(name) = clause.name() {
        out.push((name, false));
    }
    if let Some(named) = clause.as_import_clause().named_bindings() {
        match named.kind() {
            Kind::NamespaceImport => {
                if let Some(name) = named.name() {
                    out.push((name, true));
                }
            }
            Kind::NamedImports => {
                for &specifier in named.as_named_imports().elements().nodes() {
                    if let Some(name) = specifier.name() {
                        out.push((name, false));
                    }
                }
            }
            _ => {}
        }
    }
    out
}
fn binding_writes(
    rule: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
) -> Vec<Violation> {
    let mut bindings = HashMap::<String, Vec<Binding>>::new();
    walk(node, &mut |declaration| {
        let mut names = Vec::new();
        match (rule, declaration.kind()) {
            ("no-class-assign", Kind::ClassDeclaration | Kind::ClassExpression)
            | ("no-func-assign", Kind::FunctionDeclaration | Kind::FunctionExpression) => {
                if let Some(name) = declaration.name() {
                    names.push((name, false));
                }
            }
            ("no-ex-assign", Kind::CatchClause) => {
                if let Some(variable) = declaration.as_catch_clause().variable_declaration() {
                    let mut binding_names = Vec::new();
                    collect_binding(
                        variable.as_variable_declaration().name(),
                        &mut binding_names,
                    );
                    names.extend(binding_names.into_iter().map(|name| (name, false)));
                }
            }
            ("no-import-assign", Kind::ImportDeclaration) => names = import_bindings(declaration),
            _ => {}
        }
        for (name, namespace) in names {
            if let Some(symbol) = binding_symbol(name) {
                bindings
                    .entry(name.text().to_owned())
                    .or_default()
                    .push(Binding {
                        name: name.text().to_owned(),
                        symbol,
                        namespace,
                    });
            }
        }
    });
    if bindings.is_empty() {
        return Vec::new();
    }
    let mut reports = Vec::new();
    walk(node, &mut |reference| {
        if reference.kind() != Kind::Identifier {
            return;
        }
        let Some(candidates) = bindings.get(reference.text()) else {
            return;
        };
        let direct = write_reference(reference);
        let member = rule == "no-import-assign"
            && !direct
            && candidates.iter().any(|binding| binding.namespace)
            && member_write(reference, file, checker);
        if !direct && !member {
            return;
        }
        let Some(symbol) = resolve_reference(reference, checker) else {
            return;
        };
        for binding in candidates {
            if !same_symbol(symbol, binding.symbol) || (member && !binding.namespace) {
                continue;
            }
            let (id, message) = match rule {
                "no-class-assign" => (
                    "classReassignment",
                    format!("'{}' is a class.", binding.name),
                ),
                "no-func-assign" => ("isAFunction", format!("'{}' is a function.", binding.name)),
                "no-ex-assign" => (
                    "unexpected",
                    "Do not assign to the exception parameter.".to_owned(),
                ),
                "no-import-assign" if direct => {
                    ("readonly", format!("'{}' is read-only.", binding.name))
                }
                "no-import-assign" => (
                    "readonlyMember",
                    format!("The members of '{}' are read-only.", binding.name),
                ),
                _ => unreachable!(),
            };
            reports.push(Violation::new(reference, id, message));
        }
    });
    reports
}
fn outer_parentheses(mut node: P<Node>) -> P<Node> {
    while let Some(parent) = node
        .parent()
        .filter(|parent| parent.kind() == Kind::ParenthesizedExpression)
    {
        node = parent;
    }
    node
}
fn wrapped_assertion(mut node: P<Node>) -> bool {
    while let Some(parent) = node.parent() {
        match parent.kind() {
            Kind::ParenthesizedExpression => node = parent,
            Kind::AsExpression | Kind::TypeAssertionExpression | Kind::NonNullExpression => {
                return true;
            }
            _ => return false,
        }
    }
    false
}
fn member_write(node: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> bool {
    let reference = outer_parentheses(node);
    let Some(parent) = reference.parent() else {
        return false;
    };
    if matches!(
        parent.kind(),
        Kind::PropertyAccessExpression | Kind::ElementAccessExpression
    ) && parent.expression() == Some(reference)
    {
        if wrapped_assertion(parent) {
            return false;
        }
        return write_reference(parent)
            || outer_parentheses(parent)
                .parent()
                .is_some_and(|parent| parent.kind() == Kind::DeleteExpression);
    }
    if parent.kind() == Kind::SpreadAssignment {
        return in_destructuring_assignment(parent);
    }
    mutation_argument(reference, file, checker)
}
fn mutation_argument(argument: P<Node>, _file: P<SourceFile>, checker: &mut Checker) -> bool {
    let Some(call) = argument
        .parent()
        .filter(|parent| parent.kind() == Kind::CallExpression)
    else {
        return false;
    };
    if call.arguments().first() != Some(&argument) {
        return false;
    }
    let callee = ast::skip_parentheses(call.as_call_expression().expression());
    let (object, method) = match callee.kind() {
        Kind::PropertyAccessExpression => {
            let value = callee.as_property_access_expression();
            (value.expression(), value.name().text().to_owned())
        }
        Kind::ElementAccessExpression => {
            let value = callee.as_element_access_expression();
            let method = ast::skip_parentheses(value.argument_expression());
            if !matches!(
                method.kind(),
                Kind::StringLiteral | Kind::NoSubstitutionTemplateLiteral
            ) {
                return false;
            }
            (value.expression(), method.text().to_owned())
        }
        _ => return false,
    };
    let object = ast::skip_parentheses(object);
    if object.kind() != Kind::Identifier {
        return false;
    }
    let known = match object.text() {
        "Object" => matches!(
            method.as_str(),
            "assign" | "defineProperty" | "defineProperties" | "freeze" | "setPrototypeOf"
        ),
        "Reflect" => matches!(
            method.as_str(),
            "defineProperty" | "deleteProperty" | "set" | "setPrototypeOf"
        ),
        _ => false,
    };
    if !known {
        return false;
    }
    let source = ast::get_source_file_of_node(object);
    !checker
        .get_symbol_at_location_exported(object)
        .is_some_and(|symbol| {
            symbol
                .declarations()
                .iter()
                .any(|&declaration| ast::get_source_file_of_node(declaration) == source)
        })
}

const SELF_OUTER: OuterExpressionKinds =
    OuterExpressionKinds::Parentheses.union(OuterExpressionKinds::TypeAssertions);
fn self_assign(node: P<Node>, file: P<SourceFile>) -> Vec<Violation> {
    let binary = node.as_binary_expression();
    if !matches!(
        binary.operator_token().kind(),
        Kind::EqualsToken
            | Kind::AmpersandAmpersandEqualsToken
            | Kind::BarBarEqualsToken
            | Kind::QuestionQuestionEqualsToken
    ) {
        return Vec::new();
    }
    let mut reports = Vec::new();
    each_self_assignment(binary.left(), binary.right(), file, &mut |right| {
        let begin =
            tsrs_scanner::skip_trivia(file.text(), right.pos().expect_source("parsed node start"))
                as usize;
        let end = right.end().expect_source("parsed node end") as usize;
        let name: String = file.text()[begin..end]
            .chars()
            .filter(|&character| !js_whitespace(character))
            .collect();
        reports.push(
            Violation::new(
                right,
                "selfAssignment",
                format!("'{name}' is assigned to itself."),
            )
            .with_byte_range([begin, end]),
        );
    });
    reports
}
fn each_self_assignment(
    left: P<Node>,
    right: P<Node>,
    file: P<SourceFile>,
    report: &mut impl FnMut(P<Node>),
) {
    let left = ast::skip_outer_expressions(left, SELF_OUTER);
    let right = ast::skip_outer_expressions(right, SELF_OUTER);
    match (left.kind(), right.kind()) {
        (Kind::Identifier, Kind::Identifier) if left.text() == right.text() => report(right),
        (Kind::ArrayLiteralExpression, Kind::ArrayLiteralExpression) => {
            let left = left.as_array_literal_expression().elements().nodes();
            let right = right.as_array_literal_expression().elements().nodes();
            for (index, (&l, &r)) in left.iter().zip(right).enumerate() {
                if l.kind() == Kind::SpreadElement {
                    if index == right.len() - 1 && r.kind() == Kind::SpreadElement {
                        each_self_assignment(
                            l.as_spread_element().expression(),
                            r.as_spread_element().expression(),
                            file,
                            report,
                        );
                    }
                    break;
                }
                if r.kind() == Kind::SpreadElement {
                    break;
                }
                each_self_assignment(l, r, file, report);
            }
        }
        (Kind::ObjectLiteralExpression, Kind::ObjectLiteralExpression) => {
            let l = left.as_object_literal_expression().properties().nodes();
            let r = right.as_object_literal_expression().properties().nodes();
            let first = r
                .iter()
                .rposition(|node| node.kind() == Kind::SpreadAssignment)
                .map_or(0, |index| index + 1);
            let mut index = HashMap::<String, Vec<P<Node>>>::new();
            for &property in &r[first..] {
                if let Some((name, value)) = property_pair(property, file) {
                    index.entry(name).or_default().push(value);
                }
            }
            for &property in l {
                if let Some((name, value)) = property_pair(property, file)
                    && let Some(candidates) = index.get(&name)
                {
                    for &right in candidates {
                        each_self_assignment(value, right, file, report);
                    }
                }
            }
        }
        _ if ast::is_access_expression(left)
            && ast::is_access_expression(right)
            && same_reference(left, right, file) =>
        {
            report(right)
        }
        _ => {}
    }
}
fn property_pair(node: P<Node>, file: P<SourceFile>) -> Option<(String, P<Node>)> {
    match node.kind() {
        Kind::PropertyAssignment => Some((
            static_property(node.name()?, file)?,
            node.as_property_assignment().initializer(),
        )),
        Kind::ShorthandPropertyAssignment
            if node
                .as_shorthand_property_assignment()
                .object_assignment_initializer()
                .is_none() =>
        {
            let name = node.name()?;
            Some((name.text().to_owned(), name))
        }
        _ => None,
    }
}
fn same_reference(left: P<Node>, right: P<Node>, file: P<SourceFile>) -> bool {
    let outer = OuterExpressionKinds::Parentheses | OuterExpressionKinds::Assertions;
    let left = ast::skip_outer_expressions(left, outer);
    let right = ast::skip_outer_expressions(right, outer);
    if left.kind() == right.kind() {
        match left.kind() {
            Kind::Identifier | Kind::PrivateIdentifier | Kind::StringLiteral => {
                return left.text() == right.text();
            }
            Kind::ThisKeyword
            | Kind::SuperKeyword
            | Kind::NullKeyword
            | Kind::TrueKeyword
            | Kind::FalseKeyword => return true,
            Kind::NumericLiteral | Kind::BigIntLiteral => {
                return static_expression(left, file) == static_expression(right, file);
            }
            Kind::RegularExpressionLiteral => return left.text() == right.text(),
            _ => {}
        }
    }
    if !ast::is_access_expression(left) || !ast::is_access_expression(right) {
        return false;
    }
    if let Some((object, _, name)) = access(left, file) {
        return access(right, file).is_some_and(|(other, _, other_name)| {
            name == other_name && same_reference(object, other, file)
        });
    }
    if left.kind() != right.kind()
        || !same_reference(
            left.expression().unwrap(),
            right.expression().unwrap(),
            file,
        )
    {
        return false;
    }
    match left.kind() {
        Kind::PropertyAccessExpression => {
            same_reference(left.name().unwrap(), right.name().unwrap(), file)
        }
        Kind::ElementAccessExpression => same_reference(
            left.as_element_access_expression().argument_expression(),
            right.as_element_access_expression().argument_expression(),
            file,
        ),
        _ => false,
    }
}
fn js_whitespace(character: char) -> bool {
    matches!(character,'\u{0009}'..='\u{000d}'|'\u{0020}'|'\u{00a0}'|'\u{1680}'|'\u{2000}'..='\u{200a}'|'\u{2028}'|'\u{2029}'|'\u{202f}'|'\u{205f}'|'\u{3000}'|'\u{feff}')
}
