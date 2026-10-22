//! Module mock factories from frozen rslint e0a29f751d4def86d27931759098f27748fa3003.
//! Match the Rstest build's syntax, then classify syntax, local declarations,
//! and (only for typed files) native Promise return types, in that order.

use std::{cell::OnceCell, collections::HashMap};
use tsrs_ast::{self as ast, FunctionFlags, Kind, Node, NodeFlags, SourceFile, Symbol};
use tsrs_checker::{Checker, ObjectFlags, SignatureKind, Type, TypeFlags};
use tsrs_core::P;
use tsrs_scanner::skip_trivia;

use super::rsbuild_ts_syntax::lexically_shadowed;
use crate::{Violation, edits::Edit};

pub(crate) const NAMES: &[&str] = &["rstest/no-async-mock-factory"];

pub(crate) fn check(
    name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    default_library: &dyn Fn(P<SourceFile>) -> bool,
    has_type_information: bool,
) -> Vec<Violation> {
    if name != NAMES[0] || node.kind() != Kind::SourceFile {
        return vec![];
    }
    if !["rs", "rstest"]
        .iter()
        .any(|name| file.has_identifier(name))
        || !["mock", "doMock", "mockRequire", "doMockRequire"]
            .iter()
            .any(|name| file.has_identifier(name))
    {
        return vec![];
    }
    let mut nodes = vec![];
    collect(node, &mut nodes);
    let mut scan = Scan {
        file,
        nodes: &nodes,
        functions: HashMap::new(),
        declarations: HashMap::new(),
        written: HashMap::new(),
        comments: OnceCell::new(),
    };
    let mut reports = vec![];
    for &node in &nodes {
        let Some((argument, member)) = mock_factory(node) else {
            continue;
        };
        let factory = skip(argument);
        let syntax = scan.syntactic(factory);
        let declaration = if syntax == Verdict::Unknown {
            scan.declaration(factory)
        } else {
            syntax
        };
        let promise = declaration == Verdict::Promise
            || (declaration == Verdict::Unknown && has_type_information && {
                let value = checker.get_type_at_location(factory);
                always_returns_promise(value, checker, default_library)
            });
        if promise {
            let report = Violation::new(
                argument,
                "asyncMockFactory",
                format!(
                    "Mock factory must return the module object synchronously; '{member}()' throws when the factory returns a Promise. To keep part of the original module, import it with `with {{ rstest: 'importActual' }}` and spread it in."
                ),
            );
            reports.push(scan.suggestions(factory, report));
        }
    }
    reports
}

fn collect(node: P<Node>, out: &mut Vec<P<Node>>) {
    out.push(node);
    node.for_each_child(&mut |child| {
        collect(child, out);
        false
    });
}
fn skip(mut node: P<Node>) -> P<Node> {
    loop {
        node = match node.kind() {
            Kind::ParenthesizedExpression => node.as_parenthesized_expression().expression(),
            Kind::AsExpression => node.as_as_expression().expression(),
            Kind::SatisfiesExpression => node.as_satisfies_expression().expression(),
            Kind::TypeAssertionExpression => node.as_type_assertion().expression(),
            Kind::NonNullExpression => node.as_non_null_expression().expression(),
            _ => return node,
        };
    }
}
fn expression_function(node: P<Node>) -> bool {
    matches!(node.kind(), Kind::ArrowFunction | Kind::FunctionExpression)
}
fn range(node: P<Node>, file: P<SourceFile>) -> [usize; 2] {
    [
        skip_trivia(file.text(), node.pos().expect_source("parsed node")) as usize,
        node.end().expect_source("parsed node") as usize,
    ]
}

// This table is intentionally distinct from Rstest's import/alias resolver:
// mock APIs are transformed even when a local binding shadows `rs`/`rstest`.
fn plugin_call(node: P<Node>) -> Option<(P<Node>, String)> {
    if node.kind() != Kind::CallExpression {
        return None;
    }
    let call = node.as_call_expression();
    let callee = skip(call.expression());
    let (receiver, member, computed) = match callee.kind() {
        Kind::PropertyAccessExpression => {
            let access = callee.as_property_access_expression();
            if access.question_dot_token().is_some() || access.name().kind() != Kind::Identifier {
                return None;
            }
            (access.expression(), access.name().text().to_owned(), false)
        }
        Kind::ElementAccessExpression => {
            let access = callee.as_element_access_expression();
            let key = skip(access.argument_expression());
            if access.question_dot_token().is_some() || key.kind() != Kind::StringLiteral {
                return None;
            }
            (access.expression(), key.text().to_owned(), true)
        }
        _ => return None,
    };
    let actual = matches!(member.as_str(), "importActual" | "requireActual");
    if !matches!(
        member.as_str(),
        "mock"
            | "doMock"
            | "mockRequire"
            | "doMockRequire"
            | "unmock"
            | "doUnmock"
            | "importMock"
            | "requireMock"
            | "importActual"
            | "requireActual"
            | "hoisted"
    ) || ((computed || call.question_dot_token().is_some()) && !actual)
    {
        return None;
    }
    let receiver = skip(receiver);
    (receiver.kind() == Kind::Identifier && matches!(receiver.text(), "rs" | "rstest"))
        .then_some((receiver, member))
}
fn mock_factory(node: P<Node>) -> Option<(P<Node>, String)> {
    let (_, member) = plugin_call(node)?;
    if !matches!(
        member.as_str(),
        "mock" | "doMock" | "mockRequire" | "doMockRequire"
    ) {
        return None;
    }
    let mut outer = node;
    loop {
        let parent = outer.parent()?;
        if skip(parent) != outer {
            if parent.kind() != Kind::ExpressionStatement {
                return None;
            }
            break;
        }
        outer = parent;
    }
    let arguments = node.arguments();
    if arguments.len() != 2
        || arguments
            .iter()
            .any(|arg| arg.kind() == Kind::SpreadElement)
    {
        return None;
    }
    Some((arguments[1], member))
}

// Value references in the source-only baseline use the binder's lexical symbol
// tables, not a checker. Do not follow import aliases into a different file.
fn resolve(identifier: P<Node>) -> Option<P<Symbol>> {
    let name = identifier.text();
    let mut cursor = Some(identifier);
    while let Some(node) = cursor {
        if let Some(symbol) = node.locals().and_then(|table| table.get().get(name)) {
            return Some(symbol);
        }
        if node.kind() == Kind::FunctionExpression
            && node
                .name()
                .is_some_and(|name_node| name_node.text() == name)
        {
            return node.symbol();
        }
        cursor = node.parent();
    }
    None
}

#[derive(Clone, Copy, PartialEq, Eq)]
enum Verdict {
    Unknown,
    Promise,
    Sync,
}
struct Scan<'a> {
    file: P<SourceFile>,
    nodes: &'a [P<Node>],
    functions: HashMap<P<Node>, Verdict>,
    declarations: HashMap<P<Symbol>, Verdict>,
    written: HashMap<P<Symbol>, bool>,
    comments: OnceCell<Option<Vec<[usize; 2]>>>,
}
impl Scan<'_> {
    fn syntactic(&mut self, factory: P<Node>) -> Verdict {
        if factory.kind() == Kind::ObjectLiteralExpression {
            return Verdict::Sync;
        }
        if expression_function(factory) {
            self.function(factory)
        } else {
            Verdict::Unknown
        }
    }
    fn function(&mut self, function: P<Node>) -> Verdict {
        if let Some(&value) = self.functions.get(&function) {
            return value;
        }
        let value = self.compute_function(function);
        self.functions.insert(function, value);
        value
    }
    fn compute_function(&self, function: P<Node>) -> Verdict {
        let flags = ast::get_function_flags(Some(function));
        if flags.intersects(FunctionFlags::Generator) {
            return Verdict::Sync;
        }
        if flags.intersects(FunctionFlags::Async) {
            return Verdict::Promise;
        }
        if function
            .body()
            .is_some_and(|body| body.kind() == Kind::Block)
            && function.flags().intersects(NodeFlags::HasImplicitReturn)
        {
            return Verdict::Unknown;
        }
        let Some(returned) = returned_expressions(function) else {
            return Verdict::Unknown;
        };
        if !returned.is_empty()
            && returned
                .into_iter()
                .all(|node| self.promise_expression(node))
        {
            Verdict::Promise
        } else {
            Verdict::Unknown
        }
    }
    fn declaration(&mut self, factory: P<Node>) -> Verdict {
        if factory.kind() != Kind::Identifier {
            return Verdict::Unknown;
        }
        let Some(symbol) = resolve(factory) else {
            return Verdict::Unknown;
        };
        if let Some(&value) = self.declarations.get(&symbol) {
            return value;
        }
        let value = self.compute_declaration(symbol);
        self.declarations.insert(symbol, value);
        value
    }
    fn compute_declaration(&mut self, symbol: P<Symbol>) -> Verdict {
        let declarations = symbol.declarations();
        if declarations.len() != 1 {
            return Verdict::Unknown;
        }
        let declaration = declarations[0];
        if ast::get_source_file_of_node(declaration) != Some(self.file) {
            return Verdict::Unknown;
        }
        match declaration.kind() {
            Kind::FunctionDeclaration => {
                let written = *self.written.entry(symbol).or_insert_with(|| {
                    self.nodes.iter().any(|&node| {
                        node.kind() == Kind::Identifier
                            && is_write(node)
                            && resolve(node) == Some(symbol)
                    })
                });
                if written {
                    Verdict::Unknown
                } else {
                    self.function(declaration)
                }
            }
            Kind::VariableDeclaration if ast::is_var_const(declaration) => {
                if let Some(initializer) = declaration
                    .initializer()
                    .map(skip)
                    .filter(|&node| expression_function(node))
                {
                    self.function(initializer)
                } else {
                    Verdict::Unknown
                }
            }
            _ => Verdict::Unknown,
        }
    }
    fn promise_identifier(&self, node: P<Node>) -> bool {
        node.kind() == Kind::Identifier
            && node.text() == "Promise"
            && !lexically_shadowed(node, "Promise")
    }
    fn promise_expression(&self, node: P<Node>) -> bool {
        let node = skip(node);
        if node.kind() == Kind::NewExpression {
            return self.promise_identifier(skip(node.as_new_expression().expression()));
        }
        if node.kind() != Kind::CallExpression {
            return false;
        }
        let call = node.as_call_expression();
        if call.expression().kind() == Kind::ImportKeyword {
            return true;
        }
        if call.question_dot_token().is_some() {
            return false;
        }
        if let Some((receiver, member)) = plugin_call(node) {
            if member == "importMock" {
                return true;
            }
            if member != "importActual" {
                return false;
            }
            return !resolve(receiver).is_some_and(|symbol| {
                symbol.declarations().iter().any(|&decl| {
                    ast::get_source_file_of_node(decl) == Some(self.file)
                        && !matches!(
                            decl.kind(),
                            Kind::ImportSpecifier
                                | Kind::ImportClause
                                | Kind::NamespaceImport
                                | Kind::ImportEqualsDeclaration
                        )
                })
            });
        }
        let callee = skip(call.expression());
        if callee.kind() != Kind::PropertyAccessExpression {
            return false;
        }
        let access = callee.as_property_access_expression();
        access.question_dot_token().is_none()
            && access.name().kind() == Kind::Identifier
            && matches!(
                access.name().text(),
                "resolve" | "reject" | "all" | "allSettled" | "race" | "any"
            )
            && self.promise_identifier(skip(access.expression()))
    }
    fn definitely_not_promise(&self, node: P<Node>) -> bool {
        let node = skip(node);
        match node.kind() {
            Kind::ObjectLiteralExpression
            | Kind::ArrayLiteralExpression
            | Kind::StringLiteral
            | Kind::NoSubstitutionTemplateLiteral
            | Kind::TemplateExpression
            | Kind::NumericLiteral
            | Kind::BigIntLiteral
            | Kind::RegularExpressionLiteral
            | Kind::TrueKeyword
            | Kind::FalseKeyword
            | Kind::NullKeyword
            | Kind::ArrowFunction
            | Kind::FunctionExpression
            | Kind::ClassExpression => true,
            Kind::Identifier => {
                node.text() == "undefined" && !lexically_shadowed(node, "undefined")
            }
            _ => false,
        }
    }
    fn has_comment(&self, start: usize, end: usize) -> bool {
        if start >= end {
            return false;
        }
        let comments = self.comments.get_or_init(|| {
            let mut ranges = vec![];
            crate::comments::for_each_comment(self.file, |comment| {
                ranges.push([comment.pos() as usize, comment.end() as usize]);
                Ok(())
            })
            .ok()
            .map(|()| ranges)
        });
        comments.as_ref().is_none_or(|comments| {
            comments
                .iter()
                .any(|range| range[0] < end && range[1] > start)
        })
    }
    fn suggestions(&self, factory: P<Node>, report: Violation) -> Violation {
        if !expression_function(factory) || factory.type_node().is_some() {
            return report;
        }
        let Some(returned) = returned_expressions(factory) else {
            return report;
        };
        if ast::get_function_flags(Some(factory)).intersects(FunctionFlags::Async) {
            if returned
                .iter()
                .any(|&node| !self.definitely_not_promise(node))
                || body_awaits(factory)
            {
                return report;
            }
            let Some(modifiers) = factory.modifiers() else {
                return report;
            };
            for &modifier in modifiers.nodes() {
                if modifier.kind() != Kind::AsyncKeyword {
                    continue;
                }
                let mut removal = range(modifier, self.file);
                for ch in self.file.text()[removal[1]..].chars() {
                    if !js_whitespace(ch) {
                        break;
                    }
                    removal[1] += ch.len_utf8();
                }
                if self.has_comment(removal[0], removal[1]) {
                    return report;
                }
                return report.with_suggestion(
                    "suggestRemoveAsync",
                    "Remove 'async' so the factory returns the module object directly.",
                    vec![Edit::replace(removal, "")],
                );
            }
            return report;
        }
        if returned.len() != 1 {
            return report;
        }
        let wrapper = skip(returned[0]);
        if wrapper.kind() != Kind::CallExpression {
            return report;
        }
        let callee = skip(wrapper.as_call_expression().expression());
        if callee.kind() != Kind::PropertyAccessExpression {
            return report;
        }
        let access = callee.as_property_access_expression();
        if access.name().kind() != Kind::Identifier
            || access.name().text() != "resolve"
            || !self.promise_identifier(skip(access.expression()))
            || wrapper.arguments().len() != 1
        {
            return report;
        }
        let resolved = wrapper.arguments()[0];
        if resolved.kind() == Kind::SpreadElement || !self.definitely_not_promise(resolved) {
            return report;
        }
        let outer = range(wrapper, self.file);
        let inner = range(resolved, self.file);
        if self.has_comment(outer[0], inner[0]) || self.has_comment(inner[1], outer[1]) {
            return report;
        }
        let mut replacement = self.file.text()[inner[0]..inner[1]].to_string();
        if factory.body() == Some(returned[0]) && replacement.starts_with('{') {
            replacement = format!("({replacement})");
        }
        report.with_suggestion(
            "suggestUnwrapPromiseResolve",
            "Return the module object directly instead of wrapping it in 'Promise.resolve()'.",
            vec![Edit::replace(outer, replacement)],
        )
    }
}

fn returned_expressions(function: P<Node>) -> Option<Vec<P<Node>>> {
    let body = function.body()?;
    if body.kind() != Kind::Block {
        return Some(vec![body]);
    }
    let mut returned = vec![];
    let mut complete = true;
    fn walk(node: P<Node>, returned: &mut Vec<P<Node>>, complete: &mut bool) {
        if ast::is_function_like(node) {
            return;
        }
        if node.kind() == Kind::ReturnStatement {
            if let Some(expression) = node.as_return_statement().expression() {
                returned.push(expression);
            } else {
                *complete = false;
            }
            return;
        }
        node.for_each_child(&mut |child| {
            walk(child, returned, complete);
            false
        });
    }
    body.for_each_child(&mut |child| {
        walk(child, &mut returned, &mut complete);
        false
    });
    complete.then_some(returned)
}
fn body_awaits(function: P<Node>) -> bool {
    fn walk(node: P<Node>) -> bool {
        if ast::is_function_like(node) {
            let body = node.body();
            return node.for_each_child(&mut |child| Some(child) != body && walk(child));
        }
        if node.kind() == Kind::AwaitExpression
            || (node.kind() == Kind::ForOfStatement
                && node.as_for_in_or_of_statement().await_modifier.is_some())
            || (node.kind() == Kind::VariableDeclarationList && ast::is_var_await_using(node))
        {
            return true;
        }
        node.for_each_child(&mut walk)
    }
    function
        .body()
        .is_some_and(|body| body.for_each_child(&mut walk))
}
fn js_whitespace(ch: char) -> bool {
    matches!(ch, '\u{0009}'..='\u{000d}' | '\u{0020}' | '\u{00a0}' | '\u{1680}'
        | '\u{2000}'..='\u{200a}' | '\u{2028}' | '\u{2029}' | '\u{202f}' | '\u{205f}' | '\u{3000}' | '\u{feff}')
}
fn is_write(mut node: P<Node>) -> bool {
    while let Some(parent) = node.parent() {
        match parent.kind() {
            Kind::BinaryExpression => {
                let binary = parent.as_binary_expression();
                return binary.left == node
                    && ast::is_assignment_operator(binary.operator_token.kind());
            }
            Kind::PrefixUnaryExpression => {
                return matches!(
                    parent.as_prefix_unary_expression().operator,
                    Kind::PlusPlusToken | Kind::MinusMinusToken
                );
            }
            Kind::PostfixUnaryExpression => {
                return matches!(
                    parent.as_postfix_unary_expression().operator,
                    Kind::PlusPlusToken | Kind::MinusMinusToken
                );
            }
            Kind::ForInStatement | Kind::ForOfStatement => {
                return parent.initializer() == Some(node);
            }
            Kind::ShorthandPropertyAssignment => {
                if parent.name() != Some(node) {
                    return false;
                }
            }
            Kind::PropertyAssignment => {
                if parent.as_property_assignment().initializer() != node {
                    return false;
                }
            }
            Kind::ParenthesizedExpression
            | Kind::AsExpression
            | Kind::SatisfiesExpression
            | Kind::TypeAssertionExpression
            | Kind::NonNullExpression
            | Kind::ArrayLiteralExpression
            | Kind::ObjectLiteralExpression
            | Kind::SpreadElement
            | Kind::SpreadAssignment
            | Kind::ObjectBindingPattern
            | Kind::ArrayBindingPattern
            | Kind::BindingElement => {}
            _ => return false,
        }
        node = parent;
    }
    false
}

fn always_returns_promise(
    value: P<Type>,
    checker: &mut Checker,
    default_library: &dyn Fn(P<SourceFile>) -> bool,
) -> bool {
    if value.flags().intersects(TypeFlags::Union) {
        return !value.types().is_empty()
            && value
                .types()
                .iter()
                .all(|&part| always_returns_promise(part, checker, default_library));
    }
    let signatures = checker.get_signatures_of_type(value, SignatureKind::Call);
    !signatures.is_empty()
        && signatures.iter().all(|&signature| {
            let returned = checker.get_return_type_of_signature(signature);
            promise_instance(returned, checker, default_library, 0)
        })
}
fn promise_instance(
    value: P<Type>,
    checker: &mut Checker,
    default_library: &dyn Fn(P<SourceFile>) -> bool,
    depth: usize,
) -> bool {
    if depth > 8
        || value
            .flags()
            .intersects(TypeFlags::Any | TypeFlags::Unknown)
    {
        return false;
    }
    if value.flags().intersects(TypeFlags::Union) {
        return !value.types().is_empty()
            && value
                .types()
                .iter()
                .all(|&part| promise_instance(part, checker, default_library, depth + 1));
    }
    if value.flags().intersects(TypeFlags::Intersection) {
        return value
            .types()
            .iter()
            .any(|&part| promise_instance(part, checker, default_library, depth + 1));
    }
    let resolved = if value.flags().intersects(TypeFlags::Object)
        && value.object_flags().intersects(ObjectFlags::Reference)
    {
        value.target().unwrap_or(value)
    } else {
        value
    };
    if resolved.symbol().is_some_and(|symbol| {
        symbol.name() == "Promise"
            && symbol
                .declarations()
                .iter()
                .any(|&decl| ast::get_source_file_of_node(decl).is_some_and(default_library))
    }) {
        return true;
    }
    resolved
        .object_flags()
        .intersects(ObjectFlags::ClassOrInterface)
        && checker
            .get_base_types(resolved)
            .iter()
            .any(|&base| promise_instance(base, checker, default_library, depth + 1))
}
