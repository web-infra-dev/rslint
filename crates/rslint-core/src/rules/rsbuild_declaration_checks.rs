//! Declaration and property checks from frozen rslint e0a29f751d4def86d27931759098f27748fa3003.
//! Only the default options used by Rsbuild are admitted by the request layer.

use crate::{Violation, edits::Edit};
use std::collections::HashMap;
use tsrs_ast::{
    self as ast, Kind, Node, OuterExpressionKinds, SourceFile, Symbol, SymbolFlags, TokenFlags,
};
use tsrs_checker::Checker;
use tsrs_core::{P, jsnum};
use tsrs_scanner::skip_trivia;

pub(crate) const NAMES: &[&str] = &[
    "no-setter-return",
    "no-prototype-builtins",
    "no-shadow-restricted-names",
    "no-cond-assign",
    "no-dupe-keys",
    "no-dupe-class-members",
];

pub(crate) fn check(
    name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
) -> Vec<Violation> {
    match name {
        "no-dupe-keys" if node.kind() == Kind::ObjectLiteralExpression => {
            duplicate_keys(node, file)
        }
        "no-dupe-class-members" if ast::is_class_like(node) => duplicate_members(node, file),
        "no-shadow-restricted-names" => shadow_restricted(node, file, checker),
        "no-setter-return" => setter_return(node, file, checker).into_iter().collect(),
        "no-prototype-builtins" if node.kind() == Kind::CallExpression => {
            prototype_builtins(node, file).into_iter().collect()
        }
        "no-cond-assign" => conditional_assignment(node).into_iter().collect(),
        _ => Vec::new(),
    }
}
fn start(node: P<Node>, file: P<SourceFile>) -> usize {
    skip_trivia(file.text(), node.pos().expect_source("parsed start")) as usize
}
fn end(node: P<Node>) -> usize {
    node.end().expect_source("parsed end") as usize
}

fn conditional_assignment(node: P<Node>) -> Option<Violation> {
    let (mut test, required) = match node.kind() {
        Kind::IfStatement => (node.as_if_statement().expression(), 1),
        Kind::WhileStatement => (node.as_while_statement().expression(), 1),
        Kind::DoStatement => (node.as_do_statement().expression(), 1),
        Kind::ForStatement => (node.as_for_statement().condition()?, 1),
        Kind::ConditionalExpression => (node.as_conditional_expression().condition(), 2),
        _ => return None,
    };
    let mut depth = 0;
    while test.kind() == Kind::ParenthesizedExpression {
        test = test.as_parenthesized_expression().expression();
        depth += 1;
    }
    if depth >= required
        || test.kind() != Kind::BinaryExpression
        || !ast::is_assignment_operator(test.as_binary_expression().operator_token().kind())
    {
        return None;
    }
    Some(Violation::new(
        test,
        "missing",
        "Expected a conditional expression and instead saw an assignment.",
    ))
}

pub(crate) fn static_expression(node: P<Node>, file: P<SourceFile>) -> Option<String> {
    match node.kind() {
        Kind::StringLiteral | Kind::NoSubstitutionTemplateLiteral => Some(node.text().to_owned()),
        Kind::NumericLiteral => {
            let radix = node.as_numeric_literal().token_flags().intersects(
                TokenFlags::BinarySpecifier | TokenFlags::OctalSpecifier | TokenFlags::HexSpecifier,
            );
            if radix {
                let raw = &file.text()[start(node, file)..end(node)];
                let base = match raw.as_bytes().get(1) {
                    Some(b'b' | b'B') => 2,
                    Some(b'o' | b'O') => 8,
                    Some(b'x' | b'X') => 16,
                    _ => 0,
                };
                if base != 0 {
                    let mut value = 0.0;
                    for character in raw[2..].chars().filter(|&character| character != '_') {
                        let digit = character.to_digit(base)?;
                        // Acorn rounds the multiply and add separately. Avoid
                        // combining these with mul_add for large radix keys.
                        value *= base as f64;
                        value += digit as f64;
                    }
                    return Some(jsnum::Number(value).string());
                }
            }
            Some(jsnum::from_string(node.text()).string())
        }
        Kind::BigIntLiteral => Some(jsnum::parse_pseudo_big_int(node.text())),
        Kind::TrueKeyword => Some("true".to_owned()),
        Kind::FalseKeyword => Some("false".to_owned()),
        Kind::NullKeyword => Some("null".to_owned()),
        Kind::RegularExpressionLiteral => {
            let (pattern, flags) = node.text().strip_prefix('/')?.rsplit_once('/')?;
            let mut flags: Vec<_> = flags.chars().collect();
            flags.sort_unstable();
            Some(format!(
                "/{pattern}/{}",
                flags.into_iter().collect::<String>()
            ))
        }
        _ => None,
    }
}
pub(crate) fn static_property(node: P<Node>, file: P<SourceFile>) -> Option<String> {
    match node.kind() {
        Kind::Identifier => Some(node.text().to_owned()),
        Kind::StringLiteral | Kind::NumericLiteral | Kind::BigIntLiteral => {
            static_expression(node, file)
        }
        Kind::ComputedPropertyName => static_expression(
            ast::skip_parentheses(node.as_computed_property_name().expression()),
            file,
        ),
        _ => None,
    }
}
pub(crate) fn access(node: P<Node>, file: P<SourceFile>) -> Option<(P<Node>, P<Node>, String)> {
    match node.kind() {
        Kind::PropertyAccessExpression => {
            let value = node.as_property_access_expression();
            let name = value.name();
            if name.kind() == Kind::PrivateIdentifier {
                return None;
            }
            Some((value.expression(), name, name.text().to_owned()))
        }
        Kind::ElementAccessExpression => {
            let value = node.as_element_access_expression();
            let name = value.argument_expression();
            let unwrapped = ast::skip_outer_expressions(
                name,
                OuterExpressionKinds::Parentheses | OuterExpressionKinds::Assertions,
            );
            Some((
                value.expression(),
                name,
                static_expression(unwrapped, file)?,
            ))
        }
        _ => None,
    }
}
fn prototype_builtins(node: P<Node>, file: P<SourceFile>) -> Option<Violation> {
    let call = node.as_call_expression();
    let callee = ast::skip_parentheses(call.expression());
    let (object, report_node, name) = access(callee, file)?;
    if !matches!(
        name.as_str(),
        "hasOwnProperty" | "isPrototypeOf" | "propertyIsEnumerable"
    ) {
        return None;
    }
    let report = Violation::new(
        report_node,
        "prototypeBuildIn",
        format!("Do not access Object.prototype method '{name}' from target object."),
    );
    if after_optional(node) || super::rsbuild_ts_syntax::lexically_shadowed(node, "Object") {
        return Some(report);
    }
    let object = ast::skip_parentheses(object);
    let mut text = file.text()[start(object, file)..end(object)].to_owned();
    if object.kind() == Kind::BinaryExpression
        && object.as_binary_expression().operator_token().kind() == Kind::CommaToken
    {
        text = format!("({text})");
    }
    let open = skip_trivia(file.text(), end(call.expression()) as u32) as usize;
    if file.text().as_bytes().get(open) != Some(&b'(') {
        return Some(report);
    }
    if !call.arguments().nodes().is_empty() {
        text.push_str(", ");
    }
    Some(report.with_suggestion(
        "callObjectPrototype",
        format!("Call Object.prototype.{name} explicitly."),
        vec![
            Edit::replace(
                [start(callee, file), end(callee)],
                format!("Object.prototype.{name}.call"),
            ),
            Edit::replace([open + 1; 2], text),
        ],
    ))
}
fn after_optional(mut node: P<Node>) -> bool {
    loop {
        node = ast::skip_parentheses(node);
        if ast::is_optional_chain_root(node) {
            return true;
        }
        node = match node.kind() {
            Kind::CallExpression => node.as_call_expression().expression(),
            Kind::PropertyAccessExpression => node.as_property_access_expression().expression(),
            Kind::ElementAccessExpression => node.as_element_access_expression().expression(),
            _ => return false,
        };
    }
}

fn setter_return(node: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> Option<Violation> {
    let (function, report) = match node.kind() {
        Kind::ArrowFunction => {
            let body = node.body()?;
            if body.kind() == Kind::Block {
                return None;
            }
            (node, body)
        }
        Kind::ReturnStatement => {
            node.as_return_statement().expression()?;
            let mut ancestor = node.parent()?;
            while !ast::is_function_like_declaration(ancestor) {
                ancestor = ancestor.parent()?;
            }
            (ancestor, node)
        }
        _ => return None,
    };
    is_setter(function, file, checker)
        .then(|| Violation::new(report, "returnsValue", "Setter cannot return a value."))
}
fn is_setter(node: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> bool {
    if node.kind() == Kind::SetAccessor {
        return true;
    }
    let Some(parent) = node.parent() else {
        return false;
    };
    if parent.kind() == Kind::PropertyAssignment && parent.initializer() == Some(node) {
        return parent
            .name()
            .and_then(|name| static_property(name, file))
            .as_deref()
            == Some("set")
            && parent
                .parent()
                .is_some_and(|object| is_descriptor(object, file, checker));
    }
    node.kind() == Kind::MethodDeclaration
        && node
            .name()
            .and_then(|name| static_property(name, file))
            .as_deref()
            == Some("set")
        && is_descriptor(parent, file, checker)
}
fn is_descriptor(node: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> bool {
    if node.kind() != Kind::ObjectLiteralExpression {
        return false;
    }
    let Some(parent) = node.parent() else {
        return false;
    };
    if parent.kind() == Kind::CallExpression && parent.arguments().get(2) == Some(&node) {
        return global_method(parent, "Object", "defineProperty", file, checker)
            || global_method(parent, "Reflect", "defineProperty", file, checker);
    }
    if parent.kind() != Kind::PropertyAssignment || parent.initializer() != Some(node) {
        return false;
    }
    let Some(object) = parent
        .parent()
        .filter(|node| node.kind() == Kind::ObjectLiteralExpression)
    else {
        return false;
    };
    let Some(call) = object
        .parent()
        .filter(|node| node.kind() == Kind::CallExpression)
    else {
        return false;
    };
    call.arguments().get(1) == Some(&object)
        && (global_method(call, "Object", "create", file, checker)
            || global_method(call, "Object", "defineProperties", file, checker))
}
fn global_method(
    call: P<Node>,
    object_name: &str,
    method: &str,
    file: P<SourceFile>,
    checker: &mut Checker,
) -> bool {
    let callee = ast::skip_parentheses(call.as_call_expression().expression());
    // The frozen setter rule unwraps callee parentheses only; a parenthesized
    // receiver or computed key is deliberately a different AST boundary.
    let (object, name) = match callee.kind() {
        Kind::PropertyAccessExpression => {
            let v = callee.as_property_access_expression();
            (v.expression(), Some(v.name().text().to_owned()))
        }
        Kind::ElementAccessExpression => {
            let v = callee.as_element_access_expression();
            (
                v.expression(),
                static_expression(v.argument_expression(), file),
            )
        }
        _ => return false,
    };
    if name.as_deref() != Some(method)
        || object.kind() != Kind::Identifier
        || object.text() != object_name
    {
        return false;
    }
    if let Some(symbol) = checker.resolve_name_exported(
        object_name,
        object,
        SymbolFlags::Value | SymbolFlags::Namespace | SymbolFlags::Alias,
        false,
    ) {
        return !symbol.declarations().iter().any(|&declaration| {
            !matches!(
                declaration.kind(),
                Kind::InterfaceDeclaration | Kind::TypeAliasDeclaration
            ) && ast::get_source_file_of_node(declaration) == Some(file)
                && !ambient_augmentation(declaration)
        });
    }
    !super::rsbuild_ts_syntax::lexically_shadowed(object, object_name)
}
fn ambient_augmentation(mut node: P<Node>) -> bool {
    while let Some(parent) = node.parent() {
        if parent.kind() == Kind::ModuleDeclaration
            && (ast::is_global_scope_augmentation(parent)
                || ast::is_external_module_augmentation(parent))
        {
            return true;
        }
        node = parent;
    }
    false
}

fn register(state: &mut u8, kind: Kind) -> bool {
    let (flag, conflicts) = match kind {
        Kind::GetAccessor => (2, 3),
        Kind::SetAccessor => (4, 5),
        _ => (1, 7),
    };
    let duplicate = *state & conflicts != 0;
    *state |= flag;
    duplicate
}
fn duplicate_keys(node: P<Node>, file: P<SourceFile>) -> Vec<Violation> {
    if ast::is_assignment_target(node) && in_destructuring_assignment(node) {
        return Vec::new();
    }
    let mut states = HashMap::<String, u8>::new();
    let mut reports = Vec::new();
    for &property in node.as_object_literal_expression().properties().nodes() {
        if property.kind() == Kind::SpreadAssignment {
            continue;
        }
        let Some(name_node) = property.name() else {
            continue;
        };
        let Some(name) = static_property(name_node, file) else {
            continue;
        };
        if name == "__proto__"
            && property.kind() == Kind::PropertyAssignment
            && name_node.kind() != Kind::ComputedPropertyName
        {
            continue;
        }
        if register(states.entry(name.clone()).or_default(), property.kind()) {
            let report = if name_node.kind() == Kind::ComputedPropertyName {
                ast::skip_parentheses(name_node.as_computed_property_name().expression())
            } else {
                name_node
            };
            reports.push(Violation::new(
                report,
                "unexpected",
                format!("Duplicate key '{name}'."),
            ));
        }
    }
    reports
}
fn duplicate_members(node: P<Node>, file: P<SourceFile>) -> Vec<Violation> {
    let mut states = HashMap::<(String, bool), u8>::new();
    let mut reports = Vec::new();
    for &member in node.members() {
        if (member.kind() == Kind::Constructor && !ast::is_static(member))
            || (member.kind() != Kind::PropertyDeclaration && member.body().is_none())
        {
            continue;
        }
        if !matches!(
            member.kind(),
            Kind::GetAccessor
                | Kind::SetAccessor
                | Kind::MethodDeclaration
                | Kind::PropertyDeclaration
                | Kind::Constructor
        ) {
            continue;
        }
        let name_node = member.name();
        let name = if let Some(name) = name_node {
            let Some(name) = static_property(name, file) else {
                continue;
            };
            name
        } else if member.kind() == Kind::Constructor {
            "constructor".to_owned()
        } else {
            continue;
        };
        if register(
            states
                .entry((name.clone(), ast::is_static(member)))
                .or_default(),
            member.kind(),
        ) {
            let report = name_node.map_or(member, |name| {
                if name.kind() == Kind::ComputedPropertyName {
                    name.as_computed_property_name().expression()
                } else {
                    name
                }
            });
            reports.push(Violation::new(
                report,
                "unexpected",
                format!("Duplicate name '{name}'."),
            ));
        }
    }
    reports
}
// Mirror the baseline utility's plain object/array boundary. In particular,
// assertion wrappers around an entire object are not silently treated as a
// destructuring pattern, and default-value/computed-key expressions remain reads.
pub(crate) fn in_destructuring_assignment(mut current: P<Node>) -> bool {
    loop {
        let Some(mut parent) = current.parent() else {
            return false;
        };
        while parent.kind() == Kind::ParenthesizedExpression {
            let Some(next) = parent.parent() else {
                return false;
            };
            parent = next;
        }
        if matches!(
            current.kind(),
            Kind::ObjectLiteralExpression | Kind::ArrayLiteralExpression
        ) {
            if parent.kind() == Kind::BinaryExpression {
                let binary = parent.as_binary_expression();
                if binary.operator_token().kind() == Kind::EqualsToken
                    && ast::skip_parentheses(binary.left()) == current
                {
                    return true;
                }
            }
            if matches!(parent.kind(), Kind::ForInStatement | Kind::ForOfStatement)
                && parent.initializer() == Some(current)
            {
                return true;
            }
        }
        if (parent.kind() == Kind::BinaryExpression
            && parent.as_binary_expression().operator_token().kind() == Kind::EqualsToken
            && parent.as_binary_expression().right() == current)
            || (parent.kind() == Kind::ShorthandPropertyAssignment
                && parent
                    .as_shorthand_property_assignment()
                    .object_assignment_initializer()
                    == Some(current))
            || parent.kind() == Kind::ComputedPropertyName
        {
            return false;
        }
        let Some(next) = current.parent() else {
            return false;
        };
        current = next;
    }
}

fn shadow_restricted(node: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> Vec<Violation> {
    let mut bindings = Vec::new();
    match node.kind() {
        Kind::VariableDeclaration => {
            // CatchClause owns its binding; checking its child as an ordinary
            // variable would duplicate the report or admit safe undefined.
            if node
                .parent()
                .is_some_and(|parent| parent.kind() == Kind::CatchClause)
            {
                return Vec::new();
            }
            let variable = node.as_variable_declaration();
            let safe = variable.initializer().is_none() && !loop_variable(node);
            collect_binding(variable.name(), safe, &mut bindings);
        }
        Kind::Parameter if node.parent().is_some_and(ast::is_function_like_declaration) => {
            if let Some(name) = node.name() {
                collect_binding(name, false, &mut bindings);
            }
        }
        Kind::CatchClause => {
            if let Some(variable) = node.as_catch_clause().variable_declaration() {
                collect_binding(
                    variable.as_variable_declaration().name(),
                    false,
                    &mut bindings,
                );
            }
        }
        Kind::FunctionDeclaration
        | Kind::FunctionExpression
        | Kind::ClassDeclaration
        | Kind::ClassExpression => {
            if let Some(name) = node.name().filter(|node| node.kind() == Kind::Identifier) {
                bindings.push((name, false));
            }
        }
        Kind::ImportDeclaration => {
            if let Some(clause) = node.as_import_declaration().import_clause() {
                if let Some(name) = clause.name() {
                    bindings.push((name, false));
                }
                if let Some(named) = clause.as_import_clause().named_bindings() {
                    match named.kind() {
                        Kind::NamespaceImport => {
                            if let Some(name) = named.name() {
                                bindings.push((name, false));
                            }
                        }
                        Kind::NamedImports => {
                            for &specifier in named.as_named_imports().elements().nodes() {
                                if let Some(name) = specifier.name() {
                                    bindings.push((name, false));
                                }
                            }
                        }
                        _ => {}
                    }
                }
            }
        }
        _ => {}
    }
    bindings
        .into_iter()
        .filter_map(|(identifier, safe)| {
            let name = identifier.text();
            if !matches!(
                name,
                "undefined" | "NaN" | "Infinity" | "arguments" | "eval" | "globalThis"
            ) {
                return None;
            }
            if name == "undefined" && safe && safely_shadowed_undefined(identifier, file, checker) {
                return None;
            }
            Some(Violation::new(
                identifier,
                "shadowingRestrictedName",
                format!("Shadowing of global property '{name}'."),
            ))
        })
        .collect()
}
fn collect_binding(node: P<Node>, safe: bool, out: &mut Vec<(P<Node>, bool)>) {
    match node.kind() {
        Kind::Identifier => out.push((node, safe)),
        Kind::ObjectBindingPattern | Kind::ArrayBindingPattern => {
            for &element in node.elements() {
                if element.kind() == Kind::BindingElement
                    && let Some(name) = element.as_binding_element().name()
                {
                    collect_binding(name, safe, out);
                }
            }
        }
        _ => {}
    }
}
fn loop_variable(node: P<Node>) -> bool {
    node.parent()
        .filter(|list| list.kind() == Kind::VariableDeclarationList)
        .is_some_and(|list| {
            list.parent().is_some_and(|parent| {
                matches!(parent.kind(), Kind::ForInStatement | Kind::ForOfStatement)
                    && parent.initializer() == Some(list)
            })
        })
}
fn enclosing_scope(mut node: P<Node>) -> Option<P<Node>> {
    while let Some(parent) = node.parent() {
        if ast::is_function_like_or_class_static_block_declaration(parent)
            || matches!(parent.kind(), Kind::SourceFile | Kind::ModuleBlock)
        {
            return Some(parent);
        }
        node = parent;
    }
    None
}
fn safely_shadowed_undefined(
    identifier: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
) -> bool {
    let symbol = identifier
        .parent()
        .filter(|declaration| declaration.name() == Some(identifier))
        .and_then(|declaration| declaration.symbol());
    if let Some(symbol) = symbol
        && (symbol.declarations().iter().any(|&declaration| {
            declaration.kind() != Kind::VariableDeclaration
                || declaration.initializer().is_some()
                || loop_variable(declaration)
        }) || has_undefined_write(file.as_node(), symbol, checker))
    {
        return false;
    }
    let Some(scope) = enclosing_scope(identifier) else {
        return true;
    };
    !same_scope_nonvar(file.as_node(), scope)
}
fn same_scope_nonvar(node: P<Node>, scope: P<Node>) -> bool {
    if matches!(
        node.kind(),
        Kind::FunctionDeclaration | Kind::ClassDeclaration
    ) && node
        .name()
        .is_some_and(|name| name.kind() == Kind::Identifier && name.text() == "undefined")
        && enclosing_scope(node) == Some(scope)
    {
        return true;
    }
    node.for_each_child(&mut |child| same_scope_nonvar(child, scope))
}
fn has_undefined_write(node: P<Node>, symbol: P<Symbol>, checker: &mut Checker) -> bool {
    if node.kind() == Kind::Identifier
        && node.text() == "undefined"
        && write_reference(node)
        && checker.resolve_name_exported(
            "undefined",
            node,
            SymbolFlags::Value | SymbolFlags::Namespace | SymbolFlags::Alias,
            false,
        ) == Some(symbol)
    {
        return true;
    }
    node.for_each_child(&mut |child| has_undefined_write(child, symbol, checker))
}
fn write_reference(node: P<Node>) -> bool {
    let Some(parent) = node.parent() else {
        return false;
    };
    match parent.kind() {
        Kind::AsExpression
        | Kind::TypeAssertionExpression
        | Kind::SatisfiesExpression
        | Kind::NonNullExpression
        | Kind::ParenthesizedExpression => write_reference(parent),
        _ => ast::get_assignment_target(node).is_some(),
    }
}
