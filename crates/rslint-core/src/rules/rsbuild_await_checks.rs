//! Await rules from frozen rslint e0a29f751d4def86d27931759098f27748fa3003.
//! Thenability uses actual callback signatures and well-known symbol identity.

use crate::{Violation, edits::Edit};
use num_bigint::BigInt;
use num_traits::{FromPrimitive, One, Signed, ToPrimitive, Zero};
use std::collections::HashMap;
use tsrs_ast::{
    self as ast, FunctionFlags, Kind, ModifierFlags, Node, NodeFlags, SourceFile, Symbol,
    SymbolFlags,
};
use tsrs_checker::{Checker, ObjectFlags, SignatureKind, Type, TypeFlags, is_tuple_type_exported};
use tsrs_core::{P, jsnum, stringutil};
use tsrs_scanner::{get_scanner_for_source_file, skip_trivia};

pub(crate) const NAMES: &[&str] = &[
    "@typescript-eslint/await-thenable",
    "@typescript-eslint/require-await",
];

pub(crate) fn check(
    name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    default_library: &dyn Fn(P<SourceFile>) -> bool,
) -> Vec<Violation> {
    match name {
        "@typescript-eslint/await-thenable" => await_thenable(node, file, checker, default_library),
        "@typescript-eslint/require-await" if node.kind() == Kind::SourceFile => {
            require_await(file, checker)
        }
        _ => Vec::new(),
    }
}

fn pos(node: P<Node>) -> usize {
    node.pos().expect_source("parsed node start") as usize
}
fn end(node: P<Node>) -> usize {
    node.end().expect_source("parsed node end") as usize
}
fn start(node: P<Node>, file: P<SourceFile>) -> usize {
    skip_trivia(file.text(), pos(node) as u32) as usize
}
fn range(node: P<Node>, file: P<SourceFile>) -> [usize; 2] {
    [start(node, file), end(node)]
}
fn parts(value: P<Type>) -> Vec<P<Type>> {
    if value.flags().intersects(TypeFlags::Union) {
        value.types().to_vec()
    } else {
        vec![value]
    }
}
fn constrained(node: P<Node>, checker: &mut Checker) -> P<Type> {
    let value = checker.get_type_at_location(node);
    checker.get_base_constraint_of_type(value).unwrap_or(value)
}
fn callback(parameter: P<Symbol>, node: P<Node>, checker: &mut Checker) -> bool {
    let Some(value) = checker.get_type_of_symbol_at_location(parameter, Some(node)) else {
        return false;
    };
    let mut value = checker.get_apparent_type(value);
    if parameter.value_declaration().is_some_and(|declaration| {
        declaration.kind() == Kind::Parameter
            && declaration
                .as_parameter_declaration()
                .dot_dot_dot_token()
                .is_some()
    }) {
        let number = checker.get_number_type();
        let Some(element) = checker.get_index_type_of_type(value, number) else {
            return false;
        };
        value = element;
    }
    parts(value).into_iter().any(|part| {
        !checker
            .get_signatures_of_type(part, SignatureKind::Call)
            .is_empty()
    })
}
fn thenable(node: P<Node>, value: P<Type>, checker: &mut Checker) -> bool {
    let value = checker.get_apparent_type(value);
    for part in parts(value) {
        let Some(then) = checker.get_property_of_type(part, "then") else {
            continue;
        };
        let Some(value) = checker.get_type_of_symbol_at_location(then, Some(node)) else {
            continue;
        };
        for part in parts(value) {
            let signatures = checker
                .get_signatures_of_type(part, SignatureKind::Call)
                .to_vec();
            for signature in signatures {
                if let Some(&parameter) = signature.parameters().first()
                    && callback(parameter, node, checker)
                {
                    return true;
                }
            }
        }
    }
    false
}
fn never_awaitable(node: P<Node>, value: P<Type>, checker: &mut Checker) -> bool {
    let value = if value.flags().intersects(TypeFlags::TypeParameter) {
        let Some(constraint) = checker.get_base_constraint_of_type(value) else {
            return false;
        };
        constraint
    } else {
        value
    };
    !value
        .flags()
        .intersects(TypeFlags::Any | TypeFlags::Unknown)
        && !thenable(node, value, checker)
}
fn well_known(value: P<Type>, name: &str, checker: &mut Checker) -> bool {
    let name = checker.get_property_name_for_known_symbol_name(name);
    checker.get_property_of_type(value, &name).is_some()
}
fn async_iterator(value: P<Type>, checker: &mut Checker) -> bool {
    if value
        .flags()
        .intersects(TypeFlags::Union | TypeFlags::Intersection)
    {
        return value
            .types()
            .iter()
            .any(|&part| async_iterator(part, checker));
    }
    well_known(value, "asyncIterator", checker)
}
fn promise_constructor(
    value: P<Type>,
    checker: &mut Checker,
    default_library: &dyn Fn(P<SourceFile>) -> bool,
) -> bool {
    if value.flags().intersects(TypeFlags::Intersection) {
        return value
            .types()
            .iter()
            .any(|&part| promise_constructor(part, checker, default_library));
    }
    if value.flags().intersects(TypeFlags::Union) {
        return value
            .types()
            .iter()
            .all(|&part| promise_constructor(part, checker, default_library));
    }
    if value.flags().intersects(TypeFlags::TypeParameter) {
        return checker
            .get_base_constraint_of_type(value)
            .is_some_and(|constraint| promise_constructor(constraint, checker, default_library));
    }
    let Some(symbol) = value.symbol() else {
        return false;
    };
    if symbol.name() == "PromiseConstructor"
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
            .any(|&base| promise_constructor(base, checker, default_library));
    }
    false
}
fn first_token(node: P<Node>, file: P<SourceFile>) -> [usize; 2] {
    let scanner = get_scanner_for_source_file(file, pos(node) as u32);
    [scanner.token_start() as usize, scanner.token_end() as usize]
}
fn await_thenable(
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    default_library: &dyn Fn(P<SourceFile>) -> bool,
) -> Vec<Violation> {
    match node.kind() {
        Kind::AwaitExpression => {
            let expression = node.as_await_expression().expression();
            let value = checker.get_type_at_location(expression);
            if never_awaitable(expression, value, checker) {
                return vec![
                    Violation::new(
                        node,
                        "await",
                        "Unexpected `await` of a non-Promise (non-\"Thenable\") value.",
                    )
                    .with_suggestion(
                        "removeAwait",
                        "Remove unnecessary `await`.",
                        vec![Edit::replace(first_token(node, file), "")],
                    ),
                ];
            }
        }
        Kind::ForOfStatement => {
            let statement = node.as_for_in_or_of_statement();
            let Some(await_modifier) = statement.await_modifier() else {
                return Vec::new();
            };
            let value = checker.get_type_at_location(statement.expression());
            if !value.flags().intersects(TypeFlags::Any)
                && !parts(value)
                    .into_iter()
                    .any(|part| well_known(part, "asyncIterator", checker))
            {
                return vec![
                    Violation::new(
                        node,
                        "forAwaitOfNonAsyncIterable",
                        "Unexpected `for await...of` of a value that is not async iterable.",
                    )
                    .with_byte_range([start(node, file), pos(statement.statement())])
                    .with_suggestion(
                        "convertToOrdinaryFor",
                        "Convert to an ordinary `for...of` loop.",
                        vec![Edit::replace(range(await_modifier, file), "")],
                    ),
                ];
            }
        }
        Kind::VariableDeclarationList if ast::is_var_await_using(node) => {
            let declarations = node.as_variable_declaration_list().declarations().nodes();
            let mut reports = Vec::new();
            for &declaration in declarations {
                let Some(initializer) = declaration.initializer() else {
                    continue;
                };
                let value = checker.get_type_at_location(initializer);
                if value.flags().intersects(TypeFlags::Any)
                    || parts(value)
                        .into_iter()
                        .any(|part| well_known(part, "asyncDispose", checker))
                {
                    continue;
                }
                let mut report = Violation::new(
                    initializer,
                    "awaitUsingOfNonAsyncDisposable",
                    "Unexpected `await using` of a value that is not async disposable.",
                );
                if declarations.len() == 1 {
                    report = report.with_suggestion(
                        "removeAwait",
                        "Remove unnecessary `await`.",
                        vec![Edit::replace(first_token(node, file), "")],
                    );
                }
                reports.push(report);
            }
            return reports;
        }
        Kind::CallExpression => {
            let call = node.as_call_expression();
            let callee = ast::skip_parentheses(call.expression());
            let Some(&argument) = call.arguments().nodes().first() else {
                return Vec::new();
            };
            let (object, method) = match callee.kind() {
                Kind::PropertyAccessExpression => {
                    let access = callee.as_property_access_expression();
                    if access.name().kind() == Kind::PrivateIdentifier {
                        return Vec::new();
                    }
                    (access.expression(), Some(access.name().text().to_owned()))
                }
                Kind::ElementAccessExpression => {
                    let access = callee.as_element_access_expression();
                    (
                        access.expression(),
                        static_value(
                            access.argument_expression(),
                            file,
                            checker,
                            EvalScope::Bindings,
                            &mut Evaluation::default(),
                        )
                        .and_then(|v| v.string()),
                    )
                }
                _ => return Vec::new(),
            };
            if !method
                .as_deref()
                .is_some_and(|name| matches!(name, "all" | "allSettled" | "any" | "race"))
            {
                return Vec::new();
            }
            let value = constrained(object, checker);
            if !promise_constructor(value, checker, default_library) {
                return Vec::new();
            }
            let argument = ast::skip_parentheses(argument);
            let report = |node| {
                Violation::new(
                    node,
                    "invalidPromiseAggregatorInput",
                    "Unexpected iterable of non-Promise (non-\"Thenable\") values passed to promise aggregator.",
                )
            };
            if argument.kind() == Kind::ArrayLiteralExpression {
                let mut reports = Vec::new();
                for &element in argument.as_array_literal_expression().elements().nodes() {
                    if element.kind() == Kind::OmittedExpression {
                        continue;
                    }
                    let element = ast::skip_parentheses(element);
                    let value = constrained(element, checker);
                    if parts(value)
                        .into_iter()
                        .all(|part| never_awaitable(element, part, checker))
                    {
                        reports.push(report(element));
                    }
                }
                return reports;
            }
            let value = constrained(argument, checker);
            let values = parts(value);
            if !values
                .iter()
                .all(|&part| well_known(part, "iterator", checker))
            {
                return Vec::new();
            }
            for value in values {
                let arguments = if is_tuple_type_exported(value) {
                    checker.get_type_arguments(value).to_vec()
                } else if checker.is_array_like_type(value) {
                    let number = checker.get_number_type();
                    checker
                        .get_index_type_of_type(value, number)
                        .into_iter()
                        .collect()
                } else if value.object_flags().intersects(ObjectFlags::Reference) {
                    checker
                        .get_type_arguments(value)
                        .first()
                        .copied()
                        .into_iter()
                        .collect()
                } else {
                    Vec::new()
                };
                for value in arguments {
                    if parts(value)
                        .into_iter()
                        .any(|part| never_awaitable(argument, part, checker))
                    {
                        return vec![report(argument)];
                    }
                }
            }
        }
        _ => {}
    }
    Vec::new()
}

fn function(node: P<Node>) -> bool {
    matches!(
        node.kind(),
        Kind::FunctionDeclaration
            | Kind::FunctionExpression
            | Kind::ArrowFunction
            | Kind::MethodDeclaration
            | Kind::Constructor
            | Kind::GetAccessor
            | Kind::SetAccessor
    )
}
fn nonempty(node: P<Node>) -> bool {
    node.body().is_some_and(|body| {
        body.kind() != Kind::Block || !body.as_block().statements().nodes().is_empty()
    })
}
fn has_await(function_node: P<Node>, checker: &mut Checker) -> bool {
    let flags = ast::get_function_flags(Some(function_node));
    let mut nodes = Vec::new();
    function_node.for_each_child(&mut |child| {
        nodes.push(child);
        false
    });
    while let Some(node) = nodes.pop() {
        // The original listener pushes a scope before all function children,
        // including parameters, computed names and decorators.
        if function(node) {
            continue;
        }
        if node.kind() == Kind::AwaitExpression
            || (node.kind() == Kind::ForOfStatement
                && node.as_for_in_or_of_statement().await_modifier().is_some())
            || (node.kind() == Kind::VariableDeclarationList && ast::is_var_await_using(node))
        {
            return true;
        }
        if node.kind() == Kind::ReturnStatement {
            if let Some(expression) = node.expression() {
                let value = checker.get_type_at_location(expression);
                if thenable(expression, value, checker) {
                    return true;
                }
            }
        } else if node.kind() == Kind::YieldExpression
            && flags.contains(FunctionFlags::AsyncGenerator)
            && let Some(expression) = node.expression()
            && !ast::is_literal_expression(expression)
        {
            let value = checker.get_type_at_location(expression);
            if if node.as_yield_expression().asterisk_token().is_none() {
                thenable(expression, value, checker)
            } else {
                async_iterator(value, checker)
            } {
                return true;
            }
        }
        node.for_each_child(&mut |child| {
            nodes.push(child);
            false
        });
    }
    if function_node.kind() == Kind::ArrowFunction
        && let Some(body) = function_node.body()
    {
        let body = ast::skip_parentheses(body);
        if !matches!(body.kind(), Kind::Block | Kind::AwaitExpression) {
            let value = checker.get_type_at_location(body);
            return thenable(body, value, checker);
        }
    }
    false
}

#[derive(Clone, Copy)]
struct Token {
    kind: Kind,
    start: usize,
    end: usize,
}
fn tokens(file: P<SourceFile>) -> Vec<Token> {
    let mut tokens = Vec::new();
    crate::comments::for_each_token_range(file, |kind, pos, end, _parent| {
        let start = skip_trivia(file.text(), pos) as usize;
        if start < end as usize {
            tokens.push(Token {
                kind,
                start,
                end: end as usize,
            });
        }
        Ok(())
    })
    .expect("source token traversal");
    tokens
}
fn token_before(tokens: &[Token], position: usize) -> Option<Token> {
    let index = tokens.partition_point(|token| token.end <= position);
    index.checked_sub(1).map(|i| tokens[i])
}
fn token_after(tokens: &[Token], position: usize) -> Option<Token> {
    tokens
        .get(tokens.partition_point(|token| token.start < position))
        .copied()
}
fn first_paren(node: P<Node>, file: P<SourceFile>, from: usize) -> Option<usize> {
    // This deliberately matches the frozen helper's scanner behavior, rather
    // than using the parameter-list start for generic constraints with `(`.
    let mut scanner = get_scanner_for_source_file(file, from as u32);
    while (scanner.token_start() as usize) < end(node) && scanner.token() != Kind::EndOfFile {
        if scanner.token() == Kind::OpenParenToken {
            return Some(scanner.token_start() as usize);
        }
        scanner.scan();
    }
    None
}
fn start_without_decorators(node: P<Node>, file: P<SourceFile>) -> usize {
    node.modifiers()
        .and_then(|modifiers| {
            modifiers
                .nodes()
                .iter()
                .filter(|modifier| modifier.kind() == Kind::Decorator)
                .map(|&modifier| end(modifier))
                .max()
        })
        .map_or_else(
            || start(node, file),
            |position| skip_trivia(file.text(), position as u32) as usize,
        )
}
fn estree_parent(node: P<Node>) -> Option<P<Node>> {
    let mut parent = node.parent();
    while let Some(value) = parent {
        if value.kind() == Kind::ParenthesizedExpression
            || (matches!(value.kind(), Kind::AsExpression | Kind::SatisfiesExpression)
                && jsdoc_cast(value))
        {
            parent = value.parent();
        } else {
            break;
        }
    }
    parent
}
fn jsdoc_syntax(node: P<Node>) -> bool {
    node.flags()
        .intersects(NodeFlags::JSDoc | NodeFlags::Reparsed)
        || ast::is_jsdoc_node(node)
}
fn jsdoc_cast(node: P<Node>) -> bool {
    ast::is_jsdoc_type_assertion(node)
        || (ast::is_in_js_file(node)
            && matches!(node.kind(), Kind::AsExpression | Kind::SatisfiesExpression)
            && node.type_node().is_some_and(jsdoc_syntax))
        || (node.kind() == Kind::AsExpression
            && node.parent().is_some_and(ast::is_jsdoc_type_assertion))
}
fn holds_function(node: P<Node>) -> bool {
    node.kind() == Kind::PropertyAssignment
        || (node.kind() == Kind::PropertyDeclaration
            && !ast::has_syntactic_modifier(node, ModifierFlags::Accessor))
}
pub(crate) fn function_head_range(node: P<Node>, file: P<SourceFile>) -> [usize; 2] {
    function_head(node, file, &tokens(file))
}

fn function_head(node: P<Node>, file: P<SourceFile>, tokens: &[Token]) -> [usize; 2] {
    match node.kind() {
        Kind::MethodDeclaration | Kind::GetAccessor | Kind::SetAccessor | Kind::Constructor => {
            let start = start_without_decorators(node, file);
            let search = node.name().map_or(start, end);
            [
                start,
                first_paren(node, file, search)
                    .or_else(|| node.body().map(pos))
                    .unwrap_or(end(node)),
            ]
        }
        Kind::ArrowFunction => {
            let arrow = node.as_arrow_function();
            if let Some(parent) = estree_parent(node).filter(|&parent| holds_function(parent)) {
                let parameters = node.parameters();
                let boundary = if parameters.len() == 1 {
                    let parameter_start = start(parameters[0], file);
                    token_before(tokens, parameter_start)
                        .filter(|t| t.kind == Kind::OpenParenToken)
                        .map_or(parameter_start, |t| t.start)
                } else {
                    first_paren(node, file, pos(node)).unwrap_or(pos(arrow
                        .equals_greater_than_token()
                        .expect("parsed arrow token")))
                };
                [start_without_decorators(parent, file), boundary]
            } else {
                range(
                    arrow
                        .equals_greater_than_token()
                        .expect("parsed arrow token"),
                    file,
                )
            }
        }
        Kind::FunctionExpression => {
            let start = estree_parent(node)
                .filter(|&parent| holds_function(parent))
                .map_or_else(
                    || start(node, file),
                    |parent| start_without_decorators(parent, file),
                );
            [
                start,
                first_paren(node, file, pos(node))
                    .or_else(|| node.body().map(pos))
                    .unwrap_or(end(node)),
            ]
        }
        Kind::FunctionDeclaration => {
            // Frozen FindFunctionKeywordPos retains async/declare; only
            // export/default prefixes are outside the function's ESTree span.
            let mut from = pos(node);
            if let Some(modifiers) = node.modifiers() {
                for &modifier in modifiers.nodes() {
                    if matches!(modifier.kind(), Kind::ExportKeyword | Kind::DefaultKeyword) {
                        from = end(modifier);
                    }
                }
            }
            let start = skip_trivia(file.text(), from as u32) as usize;
            [
                start,
                first_paren(node, file, pos(node))
                    .or_else(|| node.body().map(pos))
                    .unwrap_or(end(node)),
            ]
        }
        _ => range(node, file),
    }
}
fn function_owner(node: P<Node>, parent: P<Node>) -> Option<P<Node>> {
    if matches!(
        node.kind(),
        Kind::MethodDeclaration | Kind::GetAccessor | Kind::SetAccessor
    ) {
        return Some(node);
    }
    holds_function(parent).then_some(parent).filter(|parent| {
        parent
            .initializer()
            .is_some_and(|initializer| ast::skip_parentheses(initializer) == node)
    })
}
fn outer_name(node: P<Node>) -> Option<String> {
    if node.name().is_some() {
        return None;
    }
    if ast::has_syntactic_modifier(node, ModifierFlags::ExportDefault) {
        return Some("default".to_owned());
    }
    let parent = ast::walk_up_parenthesized_expressions(node.parent())?;
    let binding = match parent.kind() {
        Kind::VariableDeclaration | Kind::Parameter | Kind::BindingElement => {
            parent
                .initializer()
                .filter(|&initializer| ast::skip_parentheses(initializer) == node)?;
            parent.name()?
        }
        Kind::ShorthandPropertyAssignment => {
            parent
                .as_shorthand_property_assignment()
                .object_assignment_initializer()
                .filter(|&initializer| ast::skip_parentheses(initializer) == node)?;
            parent.name()?
        }
        Kind::BinaryExpression => {
            let binary = parent.as_binary_expression();
            if !ast::is_assignment_operator(binary.operator_token().kind())
                || ast::skip_parentheses(binary.right()) != node
            {
                return None;
            }
            ast::skip_parentheses(binary.left())
        }
        Kind::ExportAssignment => {
            let export = parent.as_export_assignment();
            return (!export.is_export_equals()
                && ast::skip_parentheses(export.expression()) == node)
                .then(|| "default".to_owned());
        }
        _ => return None,
    };
    (binding.kind() == Kind::Identifier).then(|| binding.text().to_owned())
}
fn function_name(node: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> String {
    if node.kind() == Kind::Constructor {
        return "Constructor".to_owned();
    }
    let Some(parent) = ast::walk_up_parenthesized_expressions(node.parent()) else {
        return String::new();
    };
    let owner = function_owner(node, parent);
    let mut words = Vec::<String>::new();
    if let Some(owner) = owner
        && owner.parent().is_some_and(ast::is_class_like)
    {
        if ast::has_syntactic_modifier(owner, ModifierFlags::Static) {
            words.push("static".into());
        }
        if owner
            .name()
            .is_some_and(|name| name.kind() == Kind::PrivateIdentifier)
        {
            words.push("private".into());
        }
    }
    let flags = ast::get_function_flags(Some(node));
    if flags.intersects(FunctionFlags::Async) {
        words.push("async".into());
    }
    if flags.intersects(FunctionFlags::Generator) {
        words.push("generator".into());
    }
    words.push(
        match node.kind() {
            Kind::GetAccessor => "getter",
            Kind::SetAccessor => "setter",
            _ if owner.is_some() => "method",
            Kind::ArrowFunction => "arrow function",
            _ => "function",
        }
        .to_owned(),
    );
    if let Some(owner) = owner {
        if let Some(name) = owner.name() {
            if name.kind() == Kind::PrivateIdentifier {
                words.push(name.text().to_owned());
            } else {
                let value =
                    super::rsbuild_declaration_checks::static_property(name, file).or_else(|| {
                        if name.kind() != Kind::ComputedPropertyName {
                            return None;
                        }
                        let expression =
                            ast::skip_parentheses(name.as_computed_property_name().expression());
                        if expression.kind() == Kind::Identifier {
                            return None;
                        }
                        static_value(
                            expression,
                            file,
                            checker,
                            EvalScope::None,
                            &mut Evaluation::default(),
                        )
                        .and_then(|value| value.string())
                    });
                if let Some(name) = value.filter(|name| !name.is_empty()) {
                    words.push(format!("'{name}'"));
                }
            }
        }
    } else {
        let own = matches!(
            node.kind(),
            Kind::FunctionDeclaration | Kind::FunctionExpression
        )
        .then(|| node.name())
        .flatten()
        .filter(|name| name.kind() == Kind::Identifier)
        .map(|name| name.text().to_owned());
        if let Some(name) = own
            .or_else(|| outer_name(node))
            .filter(|name| !name.is_empty())
        {
            words.push(format!("'{name}'"));
        }
    }
    let mut name = words.join(" ");
    if let Some(first) = name.get_mut(..1) {
        first.make_ascii_uppercase();
    }
    name
}

fn js_space(character: char) -> bool {
    matches!(
        character,
        '\t' | '\n' | '\u{b}' | '\u{c}' | '\r' | ' ' | '\u{a0}' | '\u{1680}' | '\u{2000}'
            ..='\u{200a}'
                | '\u{2028}'
                | '\u{2029}'
                | '\u{202f}'
                | '\u{205f}'
                | '\u{3000}'
                | '\u{feff}'
    )
}
fn skip_space(text: &str, from: usize, to: usize) -> usize {
    text[from..to]
        .char_indices()
        .find(|&(_, character)| !js_space(character))
        .map_or(to, |(offset, _)| from + offset)
}
fn class_semicolon(node: P<Node>, file: P<SourceFile>, next: Token) -> bool {
    if !matches!(
        next.kind,
        Kind::OpenBracketToken | Kind::AsteriskToken | Kind::InKeyword | Kind::InstanceOfKeyword
    ) && !(next.kind == Kind::Identifier
        && matches!(&file.text()[next.start..next.end], "in" | "instanceof"))
    {
        return false;
    }
    let Some(class) = node.parent().filter(|&parent| ast::is_class_like(parent)) else {
        return false;
    };
    let members = class.members();
    let Some(index) = members
        .iter()
        .position(|&member| member == node)
        .filter(|&index| index > 0)
    else {
        return false;
    };
    let previous = members[index - 1];
    if previous.kind() != Kind::PropertyDeclaration
        || file.text()[pos(previous)..end(previous)]
            .trim_end_matches(js_space)
            .ends_with(';')
    {
        return false;
    }
    let Some(initializer) = previous.initializer() else {
        return true;
    };
    if initializer.kind() == Kind::PostfixUnaryExpression {
        return next.kind != Kind::OpenBracketToken;
    }
    !(initializer.kind() == Kind::ArrowFunction
        && initializer
            .body()
            .is_some_and(|body| body.kind() == Kind::Block))
}
fn expression_statement_start(node: P<Node>, file: P<SourceFile>) -> bool {
    let first = start(node, file);
    let mut parent = node.parent();
    while let Some(current) = parent {
        if current.kind() == Kind::ParenthesizedExpression {
            return false;
        }
        if current.kind() == Kind::ExpressionStatement {
            return first == start(current, file);
        }
        if !ast::is_expression(current) {
            return false;
        }
        parent = current.parent();
    }
    false
}
fn node_or_parent(node: P<Node>, kinds: &[Kind]) -> bool {
    kinds.contains(&node.kind())
        || node
            .parent()
            .is_some_and(|parent| kinds.contains(&parent.kind()))
}
fn preceding_semicolon(node: P<Node>, file: P<SourceFile>, tokens: &[Token]) -> bool {
    let first = start(node, file);
    let Some(previous) = token_before(tokens, first) else {
        return false;
    };
    if !file.text()[previous.end..first].contains(['\n', '\r', '\u{2028}', '\u{2029}']) {
        return false;
    }
    if matches!(
        previous.kind,
        Kind::ColonToken
            | Kind::SemicolonToken
            | Kind::OpenBraceToken
            | Kind::EqualsGreaterThanToken
            | Kind::PlusPlusToken
            | Kind::MinusMinusToken
    ) {
        return false;
    }
    let previous_node = ast::get_node_at_position(file, previous.start as u32, false);
    match previous.kind {
        Kind::CloseParenToken => !matches!(
            previous_node.kind(),
            Kind::DoStatement
                | Kind::ForInStatement
                | Kind::ForOfStatement
                | Kind::ForStatement
                | Kind::IfStatement
                | Kind::WhileStatement
                | Kind::WithStatement
        ),
        Kind::CloseBraceToken => match previous_node.kind() {
            Kind::Block => previous_node
                .parent()
                .is_some_and(|parent| parent.kind() == Kind::FunctionExpression),
            Kind::ClassExpression | Kind::ObjectLiteralExpression => true,
            _ => false,
        },
        Kind::StringLiteral => !node_or_parent(
            previous_node,
            &[Kind::ExportDeclaration, Kind::ImportDeclaration],
        ),
        Kind::Identifier
        | Kind::BreakKeyword
        | Kind::ContinueKeyword
        | Kind::DebuggerKeyword
        | Kind::DoKeyword
        | Kind::ElseKeyword
        | Kind::ReturnKeyword
        | Kind::YieldKeyword => {
            let declaration = if previous_node.kind() == Kind::VariableDeclaration {
                previous_node
            } else {
                previous_node.parent().unwrap_or(previous_node)
            };
            if declaration.kind() == Kind::VariableDeclaration
                && declaration.initializer().is_none()
            {
                return false;
            }
            if node_or_parent(
                previous_node,
                &[Kind::BreakStatement, Kind::ContinueStatement],
            ) {
                return false;
            }
            let expected = match previous.kind {
                Kind::BreakKeyword => Kind::BreakStatement,
                Kind::ContinueKeyword => Kind::ContinueStatement,
                Kind::DebuggerKeyword => Kind::DebuggerStatement,
                Kind::DoKeyword => Kind::DoStatement,
                Kind::ElseKeyword => Kind::IfStatement,
                Kind::ReturnKeyword => Kind::ReturnStatement,
                Kind::YieldKeyword => Kind::YieldExpression,
                _ => return true,
            };
            previous_node.kind() != expected
        }
        _ => true,
    }
}
fn remove_async(node: P<Node>, file: P<SourceFile>, tokens: &[Token]) -> Option<Vec<Edit>> {
    let modifiers = node.modifiers()?;
    let mut prefix = false;
    let mut async_token = None;
    for &modifier in modifiers.nodes() {
        if modifier.kind() == Kind::AsyncKeyword {
            async_token = Some(modifier);
            break;
        }
        if !jsdoc_syntax(modifier) {
            prefix = true;
        }
    }
    let async_token = async_token?;
    let async_range = range(async_token, file);
    let remove_end = skip_space(file.text(), async_range[1], file.text().len());
    let next = token_after(tokens, async_range[1]);
    if let Some(next) = next {
        let mut jsdoc = false;
        if file.text()[async_range[1]..next.start].contains("/**") {
            crate::comments::for_each_comment(file, |comment| {
                let begin = comment.pos() as usize;
                let finish = comment.end() as usize;
                if comment.kind == Kind::MultiLineCommentTrivia
                    && begin >= async_range[1]
                    && finish <= next.start
                    && begin + 3 < finish
                    && file.text()[begin..finish].starts_with("/**")
                    && file.text().as_bytes()[begin + 3] != b'/'
                {
                    jsdoc = true;
                }
                Ok(())
            })
            .expect("source comment traversal");
        }
        if jsdoc {
            return None;
        }
    }
    let semicolon = next.is_some_and(|next| {
        if node.kind() == Kind::MethodDeclaration {
            !prefix && class_semicolon(node, file, next)
        } else {
            matches!(next.kind, Kind::OpenBracketToken | Kind::OpenParenToken)
                && expression_statement_start(node, file)
                && preceding_semicolon(node, file, tokens)
        }
    });
    let mut edits = Vec::new();
    if semicolon
        && node.kind() == Kind::MethodDeclaration
        && pos(node) < async_range[0]
        && skip_space(file.text(), pos(node), async_range[0]) != async_range[0]
    {
        edits.push(Edit::replace([pos(node), pos(node)], ";"));
        edits.push(Edit::replace([async_range[0], remove_end], ""));
    } else {
        edits.push(Edit::replace(
            [async_range[0], remove_end],
            if semicolon { ";" } else { "" },
        ));
    }
    if let Some(return_type) = node.type_node() {
        let return_type = ast::skip_type_parentheses(return_type);
        if return_type.kind() == Kind::TypeReference {
            let reference = return_type.as_type_reference_node();
            let type_name = reference.type_name();
            if type_name.kind() == Kind::Identifier {
                if ast::get_function_flags(Some(node)).intersects(FunctionFlags::Generator) {
                    if type_name.text() == "AsyncGenerator" {
                        edits.push(Edit::replace(range(type_name, file), "Generator"));
                    }
                } else if type_name.text() == "Promise"
                    && let Some(arguments) = reference.type_arguments()
                {
                    let open_end = arguments.pos().expect_source("type argument start") as usize;
                    let close_start = end(return_type) - 1;
                    let text = file.text().as_bytes();
                    if open_end > end(type_name)
                        && open_end <= text.len()
                        && text[open_end - 1] == b'<'
                        && close_start >= open_end
                        && text.get(close_start) == Some(&b'>')
                    {
                        edits.push(Edit::replace([close_start, close_start + 1], ""));
                        edits.push(Edit::replace([start(type_name, file), open_end], ""));
                    }
                }
            }
        }
    }
    Some(edits)
}
fn require_await(file: P<SourceFile>, checker: &mut Checker) -> Vec<Violation> {
    let mut queue = vec![file.as_node()];
    let mut missing = Vec::new();
    while let Some(node) = queue.pop() {
        if function(node)
            && nonempty(node)
            && ast::get_function_flags(Some(node)).intersects(FunctionFlags::Async)
            && !has_await(node, checker)
        {
            missing.push(node);
        }
        node.for_each_child(&mut |child| {
            queue.push(child);
            false
        });
    }
    if missing.is_empty() {
        return Vec::new();
    }
    let tokens = tokens(file);
    missing
        .into_iter()
        .map(|node| {
            let mut violation = Violation::new(
                node,
                "missingAwait",
                format!(
                    "{} has no 'await' expression.",
                    function_name(node, file, checker)
                ),
            )
            .with_byte_range(function_head(node, file, &tokens));
            if let Some(edits) = remove_async(node, file, &tokens) {
                violation = violation.with_suggestion("removeAsync", "Remove 'async'.", edits);
            }
            violation
        })
        .collect()
}

#[derive(Clone, Copy, PartialEq, Eq)]
enum EvalScope {
    None,
    Globals,
    Bindings,
}

/// The original no-obj-calls evaluator has a source file but no checker:
/// globals can be folded, local initializers cannot, and Eval accepts strings
/// only (unlike EvalToString used for property names).
pub(crate) fn static_string_no_bindings(
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
) -> Option<String> {
    match static_value(
        node,
        file,
        checker,
        EvalScope::Globals,
        &mut Evaluation::default(),
    )? {
        StaticValue::String(value) => Some(value),
        _ => None,
    }
}

// Mutation facts are scoped to one static evaluation. Initializers may alias
// aggregate values even when the alias binding is later reassigned.
#[derive(Default)]
struct Evaluation {
    symbols: Vec<P<Symbol>>,
    mutations: Option<MutationFacts>,
}
#[derive(Default)]
struct MutationFacts {
    flags: HashMap<P<Symbol>, u8>,
    array_prototype: bool,
    string_prototype: bool,
}
const PROPERTY_MUTATION: u8 = 1;
const NESTED_MUTATION: u8 = 2;

fn property_mutated(
    symbol: P<Symbol>,
    file: P<SourceFile>,
    checker: &mut Checker,
    scope: EvalScope,
    state: &mut Evaluation,
) -> bool {
    ensure_mutations(file, checker, scope, state);
    state.mutations.as_ref().is_some_and(|facts| {
        facts
            .flags
            .get(&symbol)
            .is_some_and(|flags| flags & PROPERTY_MUTATION != 0)
    })
}
fn direct_builtin(
    node: P<Node>,
    name: &str,
    file: P<SourceFile>,
    checker: &mut Checker,
    scope: EvalScope,
) -> bool {
    let node = unwrap_expression(node);
    node.kind() == Kind::Identifier
        && node.text() == name
        && builtin_identifier(node, name, file, checker, scope)
}
fn ensure_mutations(
    file: P<SourceFile>,
    checker: &mut Checker,
    scope: EvalScope,
    state: &mut Evaluation,
) {
    if state.mutations.is_some() {
        return;
    }
    // The independent stack is essential when a computed mutator key itself
    // refers to the aggregate currently being evaluated by the caller.
    let mut analysis = Evaluation {
        mutations: Some(MutationFacts::default()),
        ..Evaluation::default()
    };
    let mut queue = vec![file.as_node()];
    let mut pending = Vec::new();
    while let Some(node) = queue.pop() {
        if node.kind() == Kind::CallExpression {
            let callee = unwrap_expression(node.as_call_expression().expression());
            if let Some((_, _, name)) = super::rsbuild_declaration_checks::access(callee, file) {
                mark_call_mutation(node, callee, &name, file, checker, scope, &mut analysis);
            } else if matches!(
                callee.kind(),
                Kind::PropertyAccessExpression | Kind::ElementAccessExpression
            ) {
                pending.push((node, callee));
            }
        }
        if matches!(
            node.kind(),
            Kind::PropertyAccessExpression | Kind::ElementAccessExpression
        ) {
            let mut outer = node;
            while let Some(parent) = outer.parent() {
                if unwrap_expression(parent) == node {
                    outer = parent;
                } else {
                    break;
                }
            }
            if super::rsbuild_core_checks::write_reference(outer)
                || outer
                    .parent()
                    .is_some_and(|parent| parent.kind() == Kind::DeleteExpression)
            {
                if let Some(object) = node.expression() {
                    mark_reference(
                        object,
                        PROPERTY_MUTATION,
                        file,
                        checker,
                        scope,
                        &mut analysis,
                    );
                }
            }
        }
        let mut children = Vec::new();
        node.for_each_child(&mut |child| {
            children.push(child);
            false
        });
        queue.extend(children.into_iter().rev());
    }
    for (call, callee) in pending {
        if let Some((_, name)) = static_access(callee, file, checker, scope, &mut analysis) {
            mark_call_mutation(call, callee, &name, file, checker, scope, &mut analysis);
        }
    }
    state.mutations = analysis.mutations;
}
fn mark_call_mutation(
    call: P<Node>,
    callee: P<Node>,
    name: &str,
    file: P<SourceFile>,
    checker: &mut Checker,
    scope: EvalScope,
    state: &mut Evaluation,
) {
    let Some(object) = callee.expression().map(unwrap_expression) else {
        return;
    };
    if matches!(
        name,
        "copyWithin"
            | "fill"
            | "pop"
            | "push"
            | "reverse"
            | "shift"
            | "sort"
            | "splice"
            | "unshift"
    ) {
        mark_reference(object, PROPERTY_MUTATION, file, checker, scope, state);
    }
    let object_mutation = matches!(
        name,
        "defineProperty" | "setPrototypeOf" | "defineProperties" | "assign"
    );
    let reflect_mutation = matches!(
        name,
        "defineProperty" | "setPrototypeOf" | "deleteProperty" | "set"
    );
    let arguments = call.as_call_expression().arguments().nodes();
    if !arguments.is_empty()
        && ((object_mutation && direct_builtin(object, "Object", file, checker, scope))
            || (reflect_mutation && direct_builtin(object, "Reflect", file, checker, scope)))
    {
        mark_mutation_argument(arguments, 0, file, checker, scope, state);
        if name == "set" {
            mark_mutation_argument(arguments, 3, file, checker, scope, state);
        }
    }
}
fn mark_mutation_argument(
    arguments: &[P<Node>],
    mut position: usize,
    file: P<SourceFile>,
    checker: &mut Checker,
    scope: EvalScope,
    state: &mut Evaluation,
) {
    let mut spread_seen = false;
    for &argument in arguments {
        if argument.kind() == Kind::SpreadElement {
            mark_reference(
                argument.as_spread_element().expression(),
                PROPERTY_MUTATION | NESTED_MUTATION,
                file,
                checker,
                scope,
                state,
            );
            spread_seen = true;
            continue;
        }
        if position == 0 || spread_seen {
            mark_reference(argument, PROPERTY_MUTATION, file, checker, scope, state);
        }
        if position == 0 {
            return;
        }
        position -= 1;
    }
}
fn mark_reference(
    node: P<Node>,
    mut flag: u8,
    file: P<SourceFile>,
    checker: &mut Checker,
    scope: EvalScope,
    state: &mut Evaluation,
) {
    let mut node = unwrap_expression(node);
    while node.kind() == Kind::Identifier {
        if flag & NESTED_MUTATION != 0 {
            let array = direct_builtin(node, "Array", file, checker, scope);
            let string = direct_builtin(node, "String", file, checker, scope);
            let facts = state.mutations.as_mut().expect("mutation scan initialized");
            facts.array_prototype |= array;
            facts.string_prototype |= string;
        }
        let Some(symbol) = super::rsbuild_assignment_checks::resolve_reference(node, checker)
        else {
            return;
        };
        let flags = state
            .mutations
            .as_mut()
            .expect("mutation scan initialized")
            .flags
            .entry(symbol)
            .or_default();
        if *flags & flag == flag {
            return;
        }
        *flags |= flag;
        let declarations = symbol.declarations();
        if declarations.len() != 1 {
            return;
        }
        let mut declaration = declarations[0];
        if declaration.kind() == Kind::BindingElement {
            while declaration.kind() == Kind::BindingElement {
                let element = declaration.as_binding_element();
                if element.dot_dot_dot_token().is_some() && flag & NESTED_MUTATION == 0 {
                    return;
                }
                if let Some(initializer) = element.initializer() {
                    mark_reference(initializer, flag, file, checker, scope, state);
                }
                flag |= NESTED_MUTATION;
                let Some(parent) = declaration.parent().and_then(|pattern| pattern.parent()) else {
                    return;
                };
                declaration = parent;
            }
        }
        if declaration.kind() != Kind::VariableDeclaration {
            return;
        }
        let Some(initializer) = declaration.initializer() else {
            return;
        };
        node = unwrap_expression(initializer);
    }
    match node.kind() {
        Kind::ConditionalExpression => {
            let conditional = node.as_conditional_expression();
            mark_reference(conditional.when_true(), flag, file, checker, scope, state);
            mark_reference(conditional.when_false(), flag, file, checker, scope, state);
        }
        Kind::BinaryExpression => {
            let binary = node.as_binary_expression();
            match binary.operator_token().kind() {
                Kind::AmpersandAmpersandToken
                | Kind::BarBarToken
                | Kind::QuestionQuestionToken
                | Kind::AmpersandAmpersandEqualsToken
                | Kind::BarBarEqualsToken
                | Kind::QuestionQuestionEqualsToken => {
                    mark_reference(binary.left(), flag, file, checker, scope, state);
                    mark_reference(binary.right(), flag, file, checker, scope, state);
                }
                Kind::CommaToken | Kind::EqualsToken => {
                    mark_reference(binary.right(), flag, file, checker, scope, state)
                }
                _ => {}
            }
        }
        Kind::PropertyAccessExpression | Kind::ElementAccessExpression => {
            if let Some(object) = node.expression() {
                mark_reference(object, flag | NESTED_MUTATION, file, checker, scope, state);
            }
        }
        Kind::ObjectLiteralExpression if flag & NESTED_MUTATION != 0 => {
            for &property in node.as_object_literal_expression().properties().nodes() {
                let value = match property.kind() {
                    Kind::PropertyAssignment => property.initializer(),
                    Kind::ShorthandPropertyAssignment => property.name(),
                    Kind::SpreadAssignment => property.expression(),
                    _ => None,
                };
                if let Some(value) = value {
                    mark_reference(value, flag, file, checker, scope, state);
                }
            }
        }
        Kind::ArrayLiteralExpression if flag & NESTED_MUTATION != 0 => {
            for &element in node.as_array_literal_expression().elements().nodes() {
                let value = if element.kind() == Kind::SpreadElement {
                    element.as_spread_element().expression()
                } else {
                    element
                };
                mark_reference(value, flag, file, checker, scope, state);
            }
        }
        Kind::CallExpression => {
            let call = node.as_call_expression();
            let callee = unwrap_expression(call.expression());
            let Some((object, name)) = static_access(callee, file, checker, scope, state) else {
                return;
            };
            let arguments = call.arguments().nodes();
            if matches!(name.as_str(), "freeze" | "seal" | "preventExtensions")
                && direct_builtin(object, "Object", file, checker, scope)
                && let Some(&argument) = arguments
                    .first()
                    .filter(|argument| argument.kind() != Kind::SpreadElement)
            {
                mark_reference(argument, flag, file, checker, scope, state);
            } else if name == "assign" && builtin_identifier(object, "Object", file, checker, scope)
            {
                for &argument in arguments {
                    let spread = argument.kind() == Kind::SpreadElement;
                    let (value, flag) = if spread {
                        (
                            argument.as_spread_element().expression(),
                            flag | NESTED_MUTATION,
                        )
                    } else {
                        (argument, flag)
                    };
                    mark_reference(value, flag, file, checker, scope, state);
                    if !spread && flag & NESTED_MUTATION == 0 {
                        break;
                    }
                }
            } else if flag & NESTED_MUTATION != 0
                && name == "of"
                && builtin_identifier(object, "Array", file, checker, scope)
            {
                for &argument in arguments {
                    let value = if argument.kind() == Kind::SpreadElement {
                        argument.as_spread_element().expression()
                    } else {
                        argument
                    };
                    mark_reference(value, flag, file, checker, scope, state);
                }
            }
        }
        _ => {}
    }
}
fn iterator_values(
    value: StaticValue,
    file: P<SourceFile>,
    checker: &mut Checker,
    scope: EvalScope,
    state: &mut Evaluation,
) -> Option<Vec<StaticValue>> {
    if scope != EvalScope::None {
        ensure_mutations(file, checker, scope, state);
        let facts = state.mutations.as_ref().expect("mutation scan completed");
        if matches!(value, StaticValue::Array(_)) && facts.array_prototype
            || matches!(value, StaticValue::String(_)) && facts.string_prototype
        {
            return None;
        }
    }
    match value {
        StaticValue::Array(values) => Some(values),
        StaticValue::String(value) => {
            let units = string_units(&value);
            let mut values = Vec::new();
            let mut index = 0;
            while index < units.len() {
                let width = if (0xd800..=0xdbff).contains(&units[index])
                    && units
                        .get(index + 1)
                        .is_some_and(|unit| (0xdc00..=0xdfff).contains(unit))
                {
                    2
                } else {
                    1
                };
                values.push(StaticValue::String(units_string(
                    &units[index..index + width],
                )));
                index += width;
            }
            Some(values)
        }
        _ => None,
    }
}

#[derive(Clone, PartialEq)]
enum StaticValue {
    String(String),
    Number(f64),
    BigInt(BigInt),
    Boolean(bool),
    Null,
    Undefined,
    Array(Vec<StaticValue>),
    Object(Vec<(String, StaticValue)>, Option<Box<StaticValue>>),
    RegExp {
        source: String,
        flags: String,
        node: P<Node>,
    },
    Symbol {
        key: String,
        registered: bool,
    },
}
fn string_units(text: &str) -> Vec<u16> {
    let mut units = Vec::new();
    let mut bytes = text.as_bytes();
    while !bytes.is_empty() {
        let (character, size) = stringutil::decode_js_string_rune_bytes(bytes);
        bytes = &bytes[size..];
        if character > 0xffff {
            let (high, low) = stringutil::code_point_to_surrogate_pair(character);
            units.extend([high as u16, low as u16]);
        } else {
            units.push(character as u16);
        }
    }
    units
}
fn units_string(units: &[u16]) -> String {
    let encoded = units
        .iter()
        .map(|&unit| stringutil::encode_js_string_rune(u32::from(unit)))
        .collect::<String>();
    stringutil::combine_surrogate_pairs(&encoded).into_owned()
}
impl StaticValue {
    fn string(&self) -> Option<String> {
        Some(match self {
            Self::String(value) => value.clone(),
            Self::Number(value) => jsnum::Number(*value).string(),
            Self::BigInt(value) => value.to_string(),
            Self::Boolean(value) => value.to_string(),
            Self::Null => "null".into(),
            Self::Undefined => "undefined".into(),
            Self::Array(values) => values
                .iter()
                .map(|value| {
                    if matches!(value, Self::Null | Self::Undefined) {
                        Some(String::new())
                    } else {
                        value.string()
                    }
                })
                .collect::<Option<Vec<_>>>()?
                .join(","),
            Self::RegExp { source, flags, .. } => format!("/{source}/{flags}"),
            Self::Symbol { .. } => return None,
            Self::Object(values, prototype) => {
                if prototype
                    .as_deref()
                    .is_some_and(|value| matches!(value, Self::Null | Self::RegExp { .. }))
                {
                    return None;
                }
                if values
                    .iter()
                    .any(|(key, _)| matches!(key.as_str(), "toString" | "valueOf"))
                {
                    return None;
                }
                "[object Object]".into()
            }
        })
    }
    fn number(&self) -> Option<f64> {
        Some(match self {
            Self::Number(value) => *value,
            Self::BigInt(_) => return None,
            Self::Boolean(value) => f64::from(u8::from(*value)),
            Self::Null => 0.0,
            Self::Undefined => f64::NAN,
            _ => {
                let text = self.string()?;
                if std::str::from_utf8(text.as_bytes()).is_err() {
                    f64::NAN
                } else {
                    jsnum::from_string(&text).0
                }
            }
        })
    }
    fn truthy(&self) -> bool {
        match self {
            Self::String(value) => !value.is_empty(),
            Self::Number(value) => *value != 0.0 && !value.is_nan(),
            Self::BigInt(value) => !value.is_zero(),
            Self::Boolean(value) => *value,
            Self::Null | Self::Undefined => false,
            _ => true,
        }
    }
    fn aggregate(&self) -> bool {
        matches!(
            self,
            Self::Array(_) | Self::Object(_, _) | Self::RegExp { .. }
        )
    }
}
fn unwrap_expression(node: P<Node>) -> P<Node> {
    ast::skip_outer_expressions(
        node,
        ast::OuterExpressionKinds::Parentheses | ast::OuterExpressionKinds::Assertions,
    )
}
fn static_key(
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    scope: EvalScope,
    resolving: &mut Evaluation,
) -> Option<String> {
    if node.kind() == Kind::Identifier {
        return Some(node.text().into());
    }
    let expression = if node.kind() == Kind::ComputedPropertyName {
        node.as_computed_property_name().expression()
    } else {
        node
    };
    static_value(expression, file, checker, scope, resolving)?.string()
}
fn builtin_identifier(
    node: P<Node>,
    name: &str,
    file: P<SourceFile>,
    checker: &mut Checker,
    scope: EvalScope,
) -> bool {
    builtin_alias(node, name, file, checker, scope, &mut Vec::new())
}
fn builtin_alias(
    mut node: P<Node>,
    name: &str,
    file: P<SourceFile>,
    checker: &mut Checker,
    scope: EvalScope,
    resolving: &mut Vec<P<Symbol>>,
) -> bool {
    if scope == EvalScope::None {
        return false;
    }
    node = unwrap_expression(node);
    if node.kind() != Kind::Identifier {
        return false;
    }
    if node.text() == name {
        if scope == EvalScope::Globals {
            return !super::rsbuild_ts_syntax::lexically_shadowed(node, name);
        }
        let global = checker.resolve_name_exported(name, node, SymbolFlags::All, false);
        if global.is_none_or(|symbol| {
            symbol
                .declarations()
                .iter()
                .all(|&decl| ast::get_source_file_of_node(decl) != Some(file))
        }) {
            return true;
        }
    }
    if scope != EvalScope::Bindings {
        return false;
    }
    let Some((symbol, initializer)) = stable_initializer(node, file, checker) else {
        return false;
    };
    if resolving.contains(&symbol) {
        return false;
    }
    resolving.push(symbol);
    let result = builtin_alias(initializer, name, file, checker, scope, resolving);
    resolving.pop();
    result
}
fn stable_initializer(
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
) -> Option<(P<Symbol>, P<Node>)> {
    let symbol = super::rsbuild_assignment_checks::resolve_reference(node, checker)?;
    let declarations = symbol.declarations();
    if declarations.len() != 1 {
        return None;
    }
    let declaration = declarations[0];
    if declaration.kind() != Kind::VariableDeclaration
        || declaration
            .name()
            .is_none_or(|name| name.kind() != Kind::Identifier || name.text() != node.text())
    {
        return None;
    }
    let initializer = declaration.initializer()?;
    let list = declaration.parent()?;
    if list.kind() != Kind::VariableDeclarationList
        || ast::is_var_using(list)
        || ast::is_var_await_using(list)
    {
        return None;
    }
    if !ast::is_var_const(list) {
        let mut queue = vec![file.as_node()];
        while let Some(reference) = queue.pop() {
            if reference.kind() == Kind::Identifier
                && reference != declaration.name().unwrap()
                && reference.text() == node.text()
                && super::rsbuild_core_checks::write_reference(reference)
                && super::rsbuild_assignment_checks::resolve_reference(reference, checker)
                    == Some(symbol)
            {
                return None;
            }
            reference.for_each_child(&mut |child| {
                queue.push(child);
                false
            });
        }
    }
    Some((symbol, initializer))
}
fn static_value(
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    scope: EvalScope,
    resolving: &mut Evaluation,
) -> Option<StaticValue> {
    if optional_short_circuit(node, file, checker, scope, resolving) {
        return Some(StaticValue::Undefined);
    }
    let node = unwrap_expression(node);
    match node.kind() {
        Kind::StringLiteral | Kind::NoSubstitutionTemplateLiteral => {
            Some(StaticValue::String(node.text().into()))
        }
        Kind::BigIntLiteral => {
            let text = node.text().trim_end_matches('n').replace('_', "");
            parse_bigint(&text).map(StaticValue::BigInt)
        }
        Kind::RegularExpressionLiteral => {
            let text = node.text();
            let end = text.rfind('/')?;
            let source = text.get(1..end)?.to_owned();
            let flags = "dgimsuvy"
                .chars()
                .filter(|&flag| text[end + 1..].contains(flag))
                .collect();
            Some(StaticValue::RegExp {
                source,
                flags,
                node,
            })
        }
        Kind::NumericLiteral => Some(StaticValue::Number(
            jsnum::from_string(&super::rsbuild_declaration_checks::static_expression(
                node, file,
            )?)
            .0,
        )),
        Kind::TrueKeyword => Some(StaticValue::Boolean(true)),
        Kind::FalseKeyword => Some(StaticValue::Boolean(false)),
        Kind::NullKeyword => Some(StaticValue::Null),
        Kind::VoidExpression => Some(StaticValue::Undefined),
        Kind::TypeOfExpression => {
            let value = static_value(node.expression()?, file, checker, scope, resolving)?;
            Some(StaticValue::String(
                match value {
                    StaticValue::String(_) => "string",
                    StaticValue::Number(_) => "number",
                    StaticValue::BigInt(_) => "bigint",
                    StaticValue::Boolean(_) => "boolean",
                    StaticValue::Undefined => "undefined",
                    StaticValue::Symbol { .. } => return None,
                    _ => "object",
                }
                .into(),
            ))
        }
        Kind::Identifier if scope != EvalScope::None => {
            for name in ["undefined", "Infinity", "NaN"] {
                if builtin_identifier(node, name, file, checker, scope) {
                    return Some(match name {
                        "undefined" => StaticValue::Undefined,
                        "Infinity" => StaticValue::Number(f64::INFINITY),
                        _ => StaticValue::Number(f64::NAN),
                    });
                }
            }
            if scope == EvalScope::Globals {
                return None;
            }
            let (symbol, initializer) = stable_initializer(node, file, checker)?;
            if resolving.symbols.contains(&symbol) {
                return None;
            }
            let literal = matches!(
                unwrap_expression(initializer).kind(),
                Kind::ObjectLiteralExpression
                    | Kind::ArrayLiteralExpression
                    | Kind::RegularExpressionLiteral
            );
            if literal && property_mutated(symbol, file, checker, scope, resolving) {
                return None;
            }
            resolving.symbols.push(symbol);
            let value = static_value(initializer, file, checker, scope, resolving);
            resolving.symbols.pop();
            if !literal
                && value.as_ref().is_some_and(StaticValue::aggregate)
                && property_mutated(symbol, file, checker, scope, resolving)
            {
                return None;
            }
            value
        }
        Kind::TemplateExpression => {
            let template = node.as_template_expression();
            let mut result = template.head().text().to_owned();
            for &span in template.template_spans().nodes() {
                let span = span.as_template_span();
                result.push_str(
                    &static_value(span.expression(), file, checker, scope, resolving)?.string()?,
                );
                result.push_str(span.literal().text());
            }
            Some(StaticValue::String(result))
        }
        Kind::ConditionalExpression => {
            let conditional = node.as_conditional_expression();
            let condition = static_value(conditional.condition(), file, checker, scope, resolving)?;
            static_value(
                if condition.truthy() {
                    conditional.when_true()
                } else {
                    conditional.when_false()
                },
                file,
                checker,
                scope,
                resolving,
            )
        }
        Kind::PrefixUnaryExpression => {
            let unary = node.as_prefix_unary_expression();
            let value = static_value(unary.operand(), file, checker, scope, resolving)?;
            if let StaticValue::BigInt(value) = value {
                return Some(match unary.operator() {
                    Kind::ExclamationToken => StaticValue::Boolean(value.is_zero()),
                    Kind::MinusToken => StaticValue::BigInt(-value),
                    Kind::TildeToken => StaticValue::BigInt(!value),
                    _ => return None,
                });
            }
            Some(match unary.operator() {
                Kind::ExclamationToken => StaticValue::Boolean(!value.truthy()),
                Kind::PlusToken => StaticValue::Number(value.number()?),
                Kind::MinusToken => StaticValue::Number(-value.number()?),
                Kind::TildeToken => {
                    StaticValue::Number(jsnum::Number(value.number()?).bitwise_not().0)
                }
                _ => return None,
            })
        }
        Kind::BinaryExpression => {
            let binary = node.as_binary_expression();
            let operator = binary.operator_token().kind();
            if matches!(operator, Kind::CommaToken | Kind::EqualsToken) {
                let value = static_value(binary.right(), file, checker, scope, resolving)?;
                return (operator != Kind::EqualsToken || !value.aggregate()).then_some(value);
            }
            let left = static_value(binary.left(), file, checker, scope, resolving)?;
            if (operator == Kind::BarBarToken && left.truthy())
                || (operator == Kind::AmpersandAmpersandToken && !left.truthy())
                || (operator == Kind::QuestionQuestionToken
                    && !matches!(left, StaticValue::Null | StaticValue::Undefined))
            {
                return Some(left);
            }
            let right = static_value(binary.right(), file, checker, scope, resolving)?;
            if (matches!(left, StaticValue::BigInt(_)) || matches!(right, StaticValue::BigInt(_)))
                && !matches!(
                    operator,
                    Kind::BarBarToken
                        | Kind::AmpersandAmpersandToken
                        | Kind::QuestionQuestionToken
                        | Kind::EqualsEqualsEqualsToken
                        | Kind::ExclamationEqualsEqualsToken
                )
                && !(operator == Kind::PlusToken
                    && (matches!(left, StaticValue::String(_))
                        || matches!(right, StaticValue::String(_))
                        || left.aggregate()
                        || right.aggregate()))
            {
                return bigint_binary(operator, &left, &right);
            }
            Some(match operator {
                Kind::BarBarToken | Kind::AmpersandAmpersandToken | Kind::QuestionQuestionToken => {
                    right
                }
                Kind::PlusToken
                    if matches!(
                        left,
                        StaticValue::String(_) | StaticValue::Array(_) | StaticValue::Object(_, _)
                    ) || matches!(
                        right,
                        StaticValue::String(_) | StaticValue::Array(_) | StaticValue::Object(_, _)
                    ) =>
                {
                    StaticValue::String(left.string()? + &right.string()?)
                }
                Kind::PlusToken => StaticValue::Number(left.number()? + right.number()?),
                Kind::MinusToken => StaticValue::Number(left.number()? - right.number()?),
                Kind::AsteriskToken => StaticValue::Number(left.number()? * right.number()?),
                Kind::SlashToken => StaticValue::Number(left.number()? / right.number()?),
                Kind::PercentToken => StaticValue::Number(
                    jsnum::Number(left.number()?)
                        .remainder(jsnum::Number(right.number()?))
                        .0,
                ),
                Kind::BarToken => StaticValue::Number(
                    jsnum::Number(left.number()?)
                        .bitwise_or(jsnum::Number(right.number()?))
                        .0,
                ),
                Kind::AmpersandToken => StaticValue::Number(
                    jsnum::Number(left.number()?)
                        .bitwise_and(jsnum::Number(right.number()?))
                        .0,
                ),
                Kind::CaretToken => StaticValue::Number(
                    jsnum::Number(left.number()?)
                        .bitwise_xor(jsnum::Number(right.number()?))
                        .0,
                ),
                Kind::LessThanLessThanToken => StaticValue::Number(
                    jsnum::Number(left.number()?)
                        .left_shift(jsnum::Number(right.number()?))
                        .0,
                ),
                Kind::GreaterThanGreaterThanToken => StaticValue::Number(
                    jsnum::Number(left.number()?)
                        .signed_right_shift(jsnum::Number(right.number()?))
                        .0,
                ),
                Kind::GreaterThanGreaterThanGreaterThanToken => StaticValue::Number(
                    jsnum::Number(left.number()?)
                        .unsigned_right_shift(jsnum::Number(right.number()?))
                        .0,
                ),
                Kind::EqualsEqualsToken | Kind::ExclamationEqualsToken => StaticValue::Boolean(
                    loose_equal(&left, &right)? == (operator == Kind::EqualsEqualsToken),
                ),
                Kind::LessThanToken
                | Kind::LessThanEqualsToken
                | Kind::GreaterThanToken
                | Kind::GreaterThanEqualsToken => {
                    let left = if left.aggregate() {
                        StaticValue::String(left.string()?)
                    } else {
                        left
                    };
                    let right = if right.aggregate() {
                        StaticValue::String(right.string()?)
                    } else {
                        right
                    };
                    let order = match (&left, &right) {
                        (StaticValue::String(a), StaticValue::String(b)) => {
                            Some(string_units(a).cmp(&string_units(b)))
                        }
                        _ => left.number()?.partial_cmp(&right.number()?),
                    };
                    StaticValue::Boolean(order.is_some_and(|order| match operator {
                        Kind::LessThanToken => order.is_lt(),
                        Kind::LessThanEqualsToken => order.is_le(),
                        Kind::GreaterThanToken => order.is_gt(),
                        _ => order.is_ge(),
                    }))
                }
                Kind::AsteriskAsteriskToken => StaticValue::Number(
                    jsnum::Number(left.number()?)
                        .exponentiate(jsnum::Number(right.number()?))
                        .0,
                ),
                Kind::EqualsEqualsEqualsToken | Kind::ExclamationEqualsEqualsToken
                    if !left.aggregate() && !right.aggregate() =>
                {
                    StaticValue::Boolean(
                        (left == right) == (operator == Kind::EqualsEqualsEqualsToken),
                    )
                }
                _ => return None,
            })
        }
        Kind::ObjectLiteralExpression => {
            let mut values = Vec::new();
            let mut prototype = None;
            for &member in node.as_object_literal_expression().properties().nodes() {
                match member.kind() {
                    Kind::PropertyAssignment => {
                        let property = member.as_property_assignment();
                        let key = static_key(property.name(), file, checker, scope, resolving)?;
                        let value =
                            static_value(property.initializer(), file, checker, scope, resolving)?;
                        if key == "__proto__"
                            && property.name().kind() != Kind::ComputedPropertyName
                        {
                            if value.aggregate() || matches!(value, StaticValue::Null) {
                                prototype = Some(Box::new(value));
                            }
                        } else {
                            values.push((key, value));
                        }
                    }
                    Kind::ShorthandPropertyAssignment => {
                        let name = member.name()?;
                        values.push((
                            name.text().to_owned(),
                            static_value(name, file, checker, scope, resolving)?,
                        ));
                    }
                    Kind::SpreadAssignment => {
                        let StaticValue::Object(members, _) = static_value(
                            member.as_spread_assignment().expression(),
                            file,
                            checker,
                            scope,
                            resolving,
                        )?
                        else {
                            return None;
                        };
                        values.extend(members);
                    }
                    _ => return None,
                }
            }
            Some(StaticValue::Object(values, prototype))
        }
        Kind::ArrayLiteralExpression => {
            let mut values = Vec::new();
            for &element in node.as_array_literal_expression().elements().nodes() {
                if element.kind() == Kind::OmittedExpression {
                    values.push(StaticValue::Undefined);
                } else if element.kind() == Kind::SpreadElement {
                    let value = static_value(
                        element.as_spread_element().expression(),
                        file,
                        checker,
                        scope,
                        resolving,
                    )?;
                    let elements = iterator_values(value, file, checker, scope, resolving)?;
                    values.extend(elements);
                } else {
                    values.push(static_value(element, file, checker, scope, resolving)?);
                }
            }
            Some(StaticValue::Array(values))
        }
        Kind::PropertyAccessExpression | Kind::ElementAccessExpression => {
            let (object, key) = static_access(node, file, checker, scope, resolving)?;
            let object = unwrap_expression(object);
            if builtin_identifier(object, "Symbol", file, checker, scope)
                && WELL_KNOWN_SYMBOLS.contains(&key.as_str())
            {
                return Some(StaticValue::Symbol {
                    key,
                    registered: false,
                });
            }
            for global in ["Math", "Number"] {
                if builtin_identifier(object, global, file, checker, scope)
                    && let Some(value) = global_number(global, &key)
                {
                    return Some(StaticValue::Number(value));
                }
            }
            let value = static_value(object, file, checker, scope, resolving)?;
            match value {
                StaticValue::Object(values, prototype) => {
                    object_member(&values, prototype.as_deref(), &key)
                }
                StaticValue::RegExp { source, flags, .. } => regex_member(&source, &flags, &key),
                StaticValue::Array(values) => {
                    if key == "length" {
                        return Some(StaticValue::Number(values.len() as f64));
                    }
                    let index = key.parse::<usize>().ok()?;
                    (index.to_string() == key)
                        .then(|| values.get(index).cloned().unwrap_or(StaticValue::Undefined))
                }
                StaticValue::String(value) if key == "length" => {
                    Some(StaticValue::Number(string_units(&value).len() as f64))
                }
                StaticValue::String(value) => {
                    let index = key.parse::<usize>().ok()?;
                    if index.to_string() != key {
                        return None;
                    }
                    Some(
                        string_units(&value)
                            .get(index)
                            .map_or(StaticValue::Undefined, |unit| {
                                StaticValue::String(units_string(&[*unit]))
                            }),
                    )
                }
                _ => None,
            }
        }
        Kind::TaggedTemplateExpression => {
            let tagged = node.as_tagged_template_expression();
            let (object, key) = static_access(
                unwrap_expression(tagged.tag()),
                file,
                checker,
                scope,
                resolving,
            )?;
            if key != "raw" || !builtin_identifier(object, "String", file, checker, scope) {
                return None;
            }
            let template = tagged.template();
            if template.kind() == Kind::NoSubstitutionTemplateLiteral {
                return Some(StaticValue::String(raw_template(template, file)?));
            }
            if template.kind() != Kind::TemplateExpression {
                return None;
            }
            let template = template.as_template_expression();
            let mut result = raw_template(template.head(), file)?;
            for &span in template.template_spans().nodes() {
                let span = span.as_template_span();
                result.push_str(
                    &static_value(span.expression(), file, checker, scope, resolving)?.string()?,
                );
                result.push_str(&raw_template(span.literal(), file)?);
            }
            Some(StaticValue::String(result))
        }
        Kind::CallExpression => static_call(node, file, checker, scope, resolving),
        _ => None,
    }
}
fn static_access(
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    scope: EvalScope,
    resolving: &mut Evaluation,
) -> Option<(P<Node>, String)> {
    match node.kind() {
        Kind::PropertyAccessExpression => {
            let access = node.as_property_access_expression();
            (access.name().kind() != Kind::PrivateIdentifier)
                .then(|| (access.expression(), access.name().text().to_owned()))
        }
        Kind::ElementAccessExpression => {
            let access = node.as_element_access_expression();
            Some((
                access.expression(),
                static_value(
                    access.argument_expression(),
                    file,
                    checker,
                    scope,
                    resolving,
                )?
                .string()?,
            ))
        }
        _ => None,
    }
}
fn static_call(
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    scope: EvalScope,
    resolving: &mut Evaluation,
) -> Option<StaticValue> {
    let call = node.as_call_expression();
    let mut callee = unwrap_expression(call.expression());
    if scope == EvalScope::Bindings {
        let mut aliases = Vec::new();
        while callee.kind() == Kind::Identifier {
            let Some((symbol, initializer)) = stable_initializer(callee, file, checker) else {
                break;
            };
            if aliases.contains(&symbol) {
                return None;
            }
            aliases.push(symbol);
            callee = unwrap_expression(initializer);
        }
    }
    let args = call.arguments().nodes();
    let mut values = Vec::new();
    for &argument in args {
        if argument.kind() == Kind::SpreadElement {
            let value = static_value(
                argument.as_spread_element().expression(),
                file,
                checker,
                scope,
                resolving,
            )?;
            let elements = iterator_values(value, file, checker, scope, resolving)?;
            values.extend(elements);
        } else {
            values.push(static_value(argument, file, checker, scope, resolving)?);
        }
    }
    if builtin_identifier(callee, "String", file, checker, scope) {
        return Some(StaticValue::String(values.first().map_or_else(
            || Some(String::new()),
            |value| match value {
                StaticValue::Symbol { key, registered } => Some(format!(
                    "Symbol({}{key})",
                    if *registered { "" } else { "Symbol." }
                )),
                _ => value.string(),
            },
        )?));
    }
    if builtin_identifier(callee, "parseInt", file, checker, scope) {
        return parse_int(&values).map(StaticValue::Number);
    }
    let (object, method) = static_access(callee, file, checker, scope, resolving)?;
    let object = unwrap_expression(object);
    if method == "parseInt" && builtin_identifier(object, "Number", file, checker, scope) {
        return parse_int(&values).map(StaticValue::Number);
    }
    if builtin_identifier(object, "Symbol", file, checker, scope) {
        return match method.as_str() {
            "for" => Some(StaticValue::Symbol {
                key: values
                    .first()
                    .map_or_else(|| Some("undefined".into()), StaticValue::string)?,
                registered: true,
            }),
            "keyFor" => match values.first()? {
                StaticValue::Symbol {
                    key,
                    registered: true,
                } => Some(StaticValue::String(key.clone())),
                StaticValue::Symbol {
                    registered: false, ..
                } => Some(StaticValue::Undefined),
                _ => None,
            },
            _ => None,
        };
    }
    if builtin_identifier(object, "Math", file, checker, scope) {
        return math_call(&method, &values).map(StaticValue::Number);
    }
    if builtin_identifier(object, "Object", file, checker, scope)
        && matches!(method.as_str(), "freeze" | "seal" | "preventExtensions")
    {
        return values.into_iter().next();
    }
    if builtin_identifier(object, "Array", file, checker, scope) && method == "of" {
        return Some(StaticValue::Array(values));
    }
    if builtin_identifier(object, "String", file, checker, scope) && method == "fromCharCode" {
        let units = values
            .iter()
            .map(|value| {
                value.number().map(|v| {
                    jsnum::Number(v).unsigned_right_shift(jsnum::Number(0.0)).0 as u32 as u16
                })
            })
            .collect::<Option<Vec<_>>>()?;
        return Some(StaticValue::String(units_string(&units)));
    }
    let receiver = static_value(object, file, checker, scope, resolving)?;
    if method == "join"
        && let StaticValue::Array(elements) = &receiver
    {
        let separator = values
            .first()
            .filter(|v| !matches!(v, StaticValue::Undefined))
            .map_or_else(|| Some(",".into()), StaticValue::string)?;
        let values = elements
            .iter()
            .map(|value| {
                if matches!(value, StaticValue::Null | StaticValue::Undefined) {
                    Some(String::new())
                } else {
                    value.string()
                }
            })
            .collect::<Option<Vec<_>>>()?;
        return Some(StaticValue::String(values.join(&separator)));
    }
    let StaticValue::String(text) = receiver else {
        return None;
    };
    match method.as_str() {
        "toString" => Some(StaticValue::String(text)),
        "trim" => {
            let units = string_units(&text);
            let whitespace = |unit: u16| char::from_u32(u32::from(unit)).is_some_and(js_space);
            let first = units
                .iter()
                .position(|&unit| !whitespace(unit))
                .unwrap_or(units.len());
            let last = units
                .iter()
                .rposition(|&unit| !whitespace(unit))
                .map_or(first, |index| index + 1);
            Some(StaticValue::String(units_string(&units[first..last])))
        }
        "toUpperCase" => Some(StaticValue::String(stringutil::to_upper_js(&text))),
        "toLowerCase" => Some(StaticValue::String(stringutil::to_lower_js(&text))),
        "concat" => Some(StaticValue::String(
            text + &values
                .iter()
                .map(StaticValue::string)
                .collect::<Option<Vec<_>>>()?
                .concat(),
        )),
        "slice" | "substring" | "substr" | "charAt" | "indexOf" => {
            let units = string_units(&text);
            let len = units.len() as f64;
            let integer = |index: usize, default: f64| -> Option<f64> {
                let Some(value) = values.get(index) else {
                    return Some(default);
                };
                if matches!(value, StaticValue::Undefined) {
                    return Some(default);
                }
                let value = value.number()?;
                Some(if value.is_nan() { 0.0 } else { value.trunc() })
            };
            if method == "indexOf" {
                let needle = string_units(
                    &values
                        .first()
                        .map_or_else(|| Some("undefined".into()), StaticValue::string)?,
                );
                let from = integer(1, 0.0)?.clamp(0.0, len) as usize;
                let index = if needle.is_empty() {
                    Some(from)
                } else {
                    units[from..]
                        .windows(needle.len())
                        .position(|part| part == needle)
                        .map(|i| i + from)
                };
                return Some(StaticValue::Number(index.map_or(-1.0, |i| i as f64)));
            }
            let first = integer(0, 0.0)?;
            let (mut from, mut to) = match method.as_str() {
                "charAt" => {
                    if first < 0.0 || first >= len {
                        return Some(StaticValue::String(String::new()));
                    }
                    (first, first + 1.0)
                }
                "substring" => (first.clamp(0.0, len), integer(1, len)?.clamp(0.0, len)),
                "substr" => {
                    let from = if first < 0.0 {
                        (len + first).max(0.0)
                    } else {
                        first.min(len)
                    };
                    (from, (from + integer(1, len)?.max(0.0)).min(len))
                }
                _ => {
                    let last = integer(1, len)?;
                    (
                        if first < 0.0 {
                            (len + first).max(0.0)
                        } else {
                            first.min(len)
                        },
                        if last < 0.0 {
                            (len + last).max(0.0)
                        } else {
                            last.min(len)
                        },
                    )
                }
            };
            if method == "substring" && from > to {
                std::mem::swap(&mut from, &mut to);
            }
            to = to.max(from);
            Some(StaticValue::String(units_string(
                &units[from as usize..to as usize],
            )))
        }
        _ => None,
    }
}

const WELL_KNOWN_SYMBOLS: &[&str] = &[
    "asyncIterator",
    "hasInstance",
    "isConcatSpreadable",
    "iterator",
    "match",
    "matchAll",
    "replace",
    "search",
    "species",
    "split",
    "toPrimitive",
    "toStringTag",
    "unscopables",
    "dispose",
    "asyncDispose",
];
fn object_member(
    values: &[(String, StaticValue)],
    prototype: Option<&StaticValue>,
    key: &str,
) -> Option<StaticValue> {
    if let Some((_, value)) = values.iter().rev().find(|(name, _)| name == key) {
        return Some(value.clone());
    }
    if let Some(prototype) = prototype {
        match prototype {
            StaticValue::Object(values, prototype) => {
                return object_member(values, prototype.as_deref(), key);
            }
            StaticValue::Null => return Some(StaticValue::Undefined),
            StaticValue::RegExp { .. } => {
                return (key == "lastIndex").then_some(StaticValue::Number(0.0));
            }
            _ => return None,
        }
    }
    // Prototype methods carry callable identity and are deliberately unknown.
    if [
        "constructor",
        "toString",
        "toLocaleString",
        "valueOf",
        "hasOwnProperty",
        "isPrototypeOf",
        "propertyIsEnumerable",
        "__proto__",
        "__defineGetter__",
        "__defineSetter__",
        "__lookupGetter__",
        "__lookupSetter__",
    ]
    .contains(&key)
    {
        None
    } else {
        Some(StaticValue::Undefined)
    }
}
fn optional_short_circuit(
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    scope: EvalScope,
    resolving: &mut Evaluation,
) -> bool {
    if !ast::is_optional_chain(node) {
        return false;
    }
    if !matches!(
        node.kind(),
        Kind::PropertyAccessExpression
            | Kind::ElementAccessExpression
            | Kind::CallExpression
            | Kind::NonNullExpression
    ) {
        return false;
    }
    let Some(input) = node.expression() else {
        return false;
    };
    optional_short_circuit(input, file, checker, scope, resolving)
        || (ast::is_optional_chain_root(node)
            && static_value(input, file, checker, scope, resolving)
                .is_some_and(|value| matches!(value, StaticValue::Null | StaticValue::Undefined)))
}
fn global_number(global: &str, key: &str) -> Option<f64> {
    use std::f64::consts;
    Some(match (global, key) {
        ("Math", "E") => consts::E,
        ("Math", "LN2") => consts::LN_2,
        ("Math", "LN10") => consts::LN_10,
        ("Math", "LOG2E") => consts::LOG2_E,
        ("Math", "LOG10E") => consts::LOG10_E,
        ("Math", "PI") => consts::PI,
        ("Math", "SQRT1_2") => consts::FRAC_1_SQRT_2,
        ("Math", "SQRT2") => consts::SQRT_2,
        ("Number", "EPSILON") => f64::EPSILON,
        ("Number", "MAX_SAFE_INTEGER") => 9007199254740991.0,
        ("Number", "MIN_SAFE_INTEGER") => -9007199254740991.0,
        ("Number", "MAX_VALUE") => f64::MAX,
        ("Number", "MIN_VALUE") => f64::from_bits(1),
        ("Number", "NaN") => f64::NAN,
        ("Number", "NEGATIVE_INFINITY") => f64::NEG_INFINITY,
        ("Number", "POSITIVE_INFINITY") => f64::INFINITY,
        _ => return None,
    })
}
fn math_call(method: &str, values: &[StaticValue]) -> Option<f64> {
    let number = |index: usize| {
        values
            .get(index)
            .map_or(Some(f64::NAN), StaticValue::number)
    };
    let uint = |value: f64| {
        jsnum::Number(value)
            .unsigned_right_shift(jsnum::Number(0.0))
            .0 as u32
    };
    if matches!(method, "min" | "max" | "hypot") {
        let numbers = values
            .iter()
            .map(StaticValue::number)
            .collect::<Option<Vec<_>>>()?;
        if method == "hypot" && numbers.iter().any(|value| value.is_infinite()) {
            return Some(f64::INFINITY);
        }
        if numbers.iter().any(|value| value.is_nan()) {
            return Some(f64::NAN);
        }
        let mut result = match method {
            "min" => f64::INFINITY,
            "max" => f64::NEG_INFINITY,
            _ => 0.0,
        };
        for value in numbers {
            result = match method {
                "min" if value == 0.0 && result == 0.0 => {
                    if value.is_sign_negative() || result.is_sign_negative() {
                        -0.0
                    } else {
                        0.0
                    }
                }
                "max" if value == 0.0 && result == 0.0 => {
                    if !value.is_sign_negative() || !result.is_sign_negative() {
                        0.0
                    } else {
                        -0.0
                    }
                }
                "min" => result.min(value),
                "max" => result.max(value),
                _ => result.hypot(value),
            };
        }
        return Some(result);
    }
    let value = number(0)?;
    Some(match method {
        "abs" => value.abs(),
        "acos" => value.acos(),
        "acosh" => value.acosh(),
        "asin" => value.asin(),
        "asinh" => value.asinh(),
        "atan" => value.atan(),
        "atanh" => value.atanh(),
        "atan2" => value.atan2(number(1)?),
        "cbrt" => value.cbrt(),
        "ceil" => value.ceil(),
        "clz32" => uint(value).leading_zeros() as f64,
        "cos" => value.cos(),
        "cosh" => value.cosh(),
        "exp" => value.exp(),
        "expm1" => value.exp_m1(),
        "floor" => value.floor(),
        "fround" => value as f32 as f64,
        "imul" => uint(value).wrapping_mul(uint(number(1)?)) as i32 as f64,
        "log" => value.ln(),
        "log10" => value.log10(),
        "log1p" => value.ln_1p(),
        "log2" => value.log2(),
        "pow" => {
            jsnum::Number(value)
                .exponentiate(jsnum::Number(number(1)?))
                .0
        }
        "round" => {
            if !value.is_finite() || value == value.trunc() {
                value
            } else if (-0.5..0.0).contains(&value) {
                -0.0
            } else {
                let lower = value.floor();
                if value - lower < 0.5 {
                    lower
                } else {
                    lower + 1.0
                }
            }
        }
        "sign" => {
            if value.is_nan() || value == 0.0 {
                value
            } else {
                1.0_f64.copysign(value)
            }
        }
        "sin" => value.sin(),
        "sinh" => value.sinh(),
        "sqrt" => value.sqrt(),
        "tan" => value.tan(),
        "tanh" => value.tanh(),
        "trunc" => value.trunc(),
        "f16round" => {
            if !value.is_finite() || value == 0.0 {
                value
            } else if value.abs() >= 65520.0 {
                f64::INFINITY.copysign(value)
            } else {
                let magnitude = value.abs();
                let spacing = 2.0_f64.powi(((magnitude.log2().floor() as i32) + 1 - 11).max(-24));
                ((magnitude / spacing).round_ties_even() * spacing).copysign(value)
            }
        }
        _ => return None,
    })
}
fn loose_equal(left: &StaticValue, right: &StaticValue) -> Option<bool> {
    if std::mem::discriminant(left) == std::mem::discriminant(right) {
        return (!left.aggregate()).then_some(left == right);
    }
    let nullish = |value: &StaticValue| matches!(value, StaticValue::Null | StaticValue::Undefined);
    if nullish(left) || nullish(right) {
        return Some(nullish(left) && nullish(right));
    }
    if left.aggregate() {
        return loose_equal(&StaticValue::String(left.string()?), right);
    }
    if right.aggregate() {
        return loose_equal(left, &StaticValue::String(right.string()?));
    }
    Some(left.number()? == right.number()?)
}
fn regex_member(source: &str, flags: &str, key: &str) -> Option<StaticValue> {
    Some(match key {
        "source" => StaticValue::String(source.into()),
        "flags" => StaticValue::String(flags.into()),
        "lastIndex" => StaticValue::Number(0.0),
        "dotAll" => StaticValue::Boolean(flags.contains('s')),
        "global" => StaticValue::Boolean(flags.contains('g')),
        "hasIndices" => StaticValue::Boolean(flags.contains('d')),
        "ignoreCase" => StaticValue::Boolean(flags.contains('i')),
        "multiline" => StaticValue::Boolean(flags.contains('m')),
        "sticky" => StaticValue::Boolean(flags.contains('y')),
        "unicode" => StaticValue::Boolean(flags.contains('u')),
        _ => return None,
    })
}

const MAX_BIGINT_BITS: u64 = 1 << 20;
fn parse_bigint(text: &str) -> Option<BigInt> {
    let text = text.trim_matches(js_space);
    if text.len() > MAX_BIGINT_BITS as usize {
        return None;
    }
    if text.is_empty() {
        return Some(BigInt::zero());
    }
    let (radix, digits) = if text.starts_with("0x") || text.starts_with("0X") {
        (16, &text[2..])
    } else if text.starts_with("0o") || text.starts_with("0O") {
        (8, &text[2..])
    } else if text.starts_with("0b") || text.starts_with("0B") {
        (2, &text[2..])
    } else {
        (10, text)
    };
    let valid = if radix == 10 {
        digits
            .strip_prefix('+')
            .or_else(|| digits.strip_prefix('-'))
            .unwrap_or(digits)
    } else {
        digits
    };
    if valid.is_empty()
        || !valid
            .bytes()
            .all(|ch| ch.is_ascii() && (ch as char).is_digit(radix))
    {
        return None;
    }
    let value = BigInt::parse_bytes(digits.as_bytes(), radix)?;
    (value.bits() <= MAX_BIGINT_BITS).then_some(value)
}
fn bigint_compare(
    left: &BigInt,
    right: &StaticValue,
    equality: bool,
) -> Option<Option<std::cmp::Ordering>> {
    use std::cmp::Ordering;
    match right {
        StaticValue::BigInt(right) => return Some(Some(left.cmp(right))),
        StaticValue::Null | StaticValue::Undefined if equality => return Some(None),
        StaticValue::String(text) => return Some(parse_bigint(text).map(|right| left.cmp(&right))),
        _ if right.aggregate() => {
            return bigint_compare(left, &StaticValue::String(right.string()?), equality);
        }
        _ => {}
    }
    let right = right.number()?;
    if right.is_nan() {
        return Some(None);
    }
    if right == f64::INFINITY {
        return Some(Some(Ordering::Less));
    }
    if right == f64::NEG_INFINITY {
        return Some(Some(Ordering::Greater));
    }
    let integer = BigInt::from_f64(right.trunc())?;
    let mut order = left.cmp(&integer);
    if order == Ordering::Equal && right != right.trunc() {
        order = if right.is_sign_positive() {
            Ordering::Less
        } else {
            Ordering::Greater
        };
    }
    Some(Some(order))
}
fn bigint_binary(operator: Kind, left: &StaticValue, right: &StaticValue) -> Option<StaticValue> {
    if matches!(
        operator,
        Kind::EqualsEqualsToken
            | Kind::ExclamationEqualsToken
            | Kind::LessThanToken
            | Kind::LessThanEqualsToken
            | Kind::GreaterThanToken
            | Kind::GreaterThanEqualsToken
    ) {
        let equality = matches!(
            operator,
            Kind::EqualsEqualsToken | Kind::ExclamationEqualsToken
        );
        let order = if let StaticValue::BigInt(left) = left {
            bigint_compare(left, right, equality)?
        } else if let StaticValue::BigInt(right) = right {
            bigint_compare(right, left, equality)?.map(std::cmp::Ordering::reverse)
        } else {
            return None;
        };
        return Some(StaticValue::Boolean(match operator {
            Kind::EqualsEqualsToken => order.is_some_and(|order| order.is_eq()),
            Kind::ExclamationEqualsToken => order.is_none_or(|order| !order.is_eq()),
            Kind::LessThanToken => order.is_some_and(|order| order.is_lt()),
            Kind::LessThanEqualsToken => order.is_some_and(|order| order.is_le()),
            Kind::GreaterThanToken => order.is_some_and(|order| order.is_gt()),
            _ => order.is_some_and(|order| order.is_ge()),
        }));
    }
    let (StaticValue::BigInt(left), StaticValue::BigInt(right)) = (left, right) else {
        return None;
    };
    let value = match operator {
        Kind::PlusToken => left + right,
        Kind::MinusToken => left - right,
        Kind::AsteriskToken => {
            if left.bits() + right.bits() > MAX_BIGINT_BITS {
                return None;
            }
            left * right
        }
        Kind::SlashToken | Kind::PercentToken => {
            if right.is_zero() {
                return None;
            }
            if operator == Kind::SlashToken {
                left / right
            } else {
                left % right
            }
        }
        Kind::BarToken => left | right,
        Kind::AmpersandToken => left & right,
        Kind::CaretToken => left ^ right,
        Kind::AsteriskAsteriskToken => {
            if right.is_negative() {
                return None;
            }
            if right.is_zero() {
                BigInt::one()
            } else if left.is_zero() {
                BigInt::zero()
            } else if left.is_one() {
                BigInt::one()
            } else if left == &-BigInt::one() {
                if (right & BigInt::one()).is_zero() {
                    BigInt::one()
                } else {
                    -BigInt::one()
                }
            } else {
                let power = right.to_u32()?;
                if left.bits().saturating_mul(u64::from(power)) > MAX_BIGINT_BITS {
                    return None;
                }
                left.pow(power)
            }
        }
        Kind::LessThanLessThanToken | Kind::GreaterThanGreaterThanToken => {
            let left_shift = (operator == Kind::LessThanLessThanToken) != right.is_negative();
            let shift = right.abs().to_u64();
            if shift.is_none_or(|shift| shift > MAX_BIGINT_BITS) {
                if left_shift && !left.is_zero() {
                    return None;
                }
                if !left_shift && left.is_negative() {
                    -BigInt::one()
                } else {
                    BigInt::zero()
                }
            } else if left_shift {
                let shift = shift.unwrap();
                if left.bits() + shift > MAX_BIGINT_BITS {
                    return None;
                }
                left << (shift as usize)
            } else {
                left >> (shift.unwrap() as usize)
            }
        }
        _ => return None,
    };
    (value.bits() <= MAX_BIGINT_BITS).then_some(StaticValue::BigInt(value))
}
fn raw_template(node: P<Node>, file: P<SourceFile>) -> Option<String> {
    let raw = node.raw_text();
    let text = if raw.is_empty() {
        let source = file.text().get(start(node, file)..end(node))?;
        let closing = if matches!(
            node.kind(),
            Kind::NoSubstitutionTemplateLiteral | Kind::TemplateTail
        ) {
            1
        } else {
            2
        };
        source.get(1..source.len().checked_sub(closing)?)?
    } else {
        raw
    };
    Some(text.replace("\r\n", "\n").replace('\r', "\n"))
}
fn parse_int(values: &[StaticValue]) -> Option<f64> {
    let text = values
        .first()
        .map_or_else(|| Some("undefined".into()), StaticValue::string)?;
    if text.len() > 1 << 20 {
        return None;
    }
    let mut text = text.trim_start_matches(js_space);
    let sign = if text.starts_with('-') { -1.0 } else { 1.0 };
    if text.starts_with(['+', '-']) {
        text = &text[1..];
    }
    let mut radix = values
        .get(1)
        .map_or(Some(0.0), StaticValue::number)
        .map(|number| {
            jsnum::Number(number)
                .signed_right_shift(jsnum::Number(0.0))
                .0 as i32
        })?;
    if radix != 0 && !(2..=36).contains(&radix) {
        return Some(f64::NAN);
    }
    if (radix == 0 || radix == 16) && (text.starts_with("0x") || text.starts_with("0X")) {
        radix = 16;
        text = &text[2..];
    }
    if radix == 0 {
        radix = 10;
    }
    let digits = text
        .bytes()
        .take_while(|&byte| byte.is_ascii() && (byte as char).is_digit(radix as u32))
        .count();
    if digits == 0 {
        return Some(f64::NAN);
    }
    let value = BigInt::parse_bytes(&text.as_bytes()[..digits], radix as u32)?;
    Some(value.to_f64().unwrap_or(f64::INFINITY).copysign(sign))
}
