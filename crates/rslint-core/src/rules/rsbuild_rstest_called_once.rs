//! Pair Rstest call-count and argument assertions without crossing a possible
//! change in call history. Ported from rslint e0a29f751d4def86d27931759098f27748fa3003.

use std::collections::{HashMap, HashSet};
use tsrs_ast::{self as ast, Kind, Node, NodeFlags, SourceFile, Symbol};
use tsrs_checker::{Checker, SignatureKind, TypeFlags};
use tsrs_core::P;
use tsrs_scanner::skip_trivia;

use super::rsbuild_rstest::{collect, context_bindings, expect_entries};
use super::rsbuild_rstest_expect::{ParsedExpect, parse_expect};
use crate::{Violation, edits::Edit};

pub(crate) const NAMES: &[&str] = &["rstest/prefer-called-exactly-once-with"];

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
    // The identifier cache includes escaped names. Parenthesized literal keys
    // need the fallback traversal because they are absent from that cache.
    let roles = ["calledOnce", "toHaveBeenCalledOnce"]
        .iter()
        .any(|name| file.has_identifier(name))
        && ["calledWith", "toHaveBeenCalledWith"]
            .iter()
            .any(|name| file.has_identifier(name));
    if !roles && !bracket_opens_on_parenthesis(file.text()) {
        return vec![];
    }
    let mut nodes = vec![];
    collect(node, &mut nodes);
    let contexts = context_bindings(&nodes, file, checker);
    let mut scan = Scan {
        file,
        checker,
        default_library,
        typed: has_type_information,
        nodes: &nodes,
        contexts: &contexts,
    };
    let mut reports = vec![];
    for &node in &nodes {
        if matches!(node.kind(), Kind::SourceFile | Kind::Block) {
            reports.extend(scan.block(node));
        }
    }
    reports
}

fn bracket_opens_on_parenthesis(text: &str) -> bool {
    text.match_indices('[').any(|(index, _)| {
        text.as_bytes()
            .get(skip_trivia(text, (index + 1) as u32) as usize)
            == Some(&b'(')
    })
}

fn unwrap(mut node: P<Node>) -> P<Node> {
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
fn stable(node: P<Node>) -> bool {
    let node = unwrap(node);
    match node.kind() {
        Kind::Identifier
        | Kind::ThisKeyword
        | Kind::SuperKeyword
        | Kind::StringLiteral
        | Kind::NumericLiteral
        | Kind::BigIntLiteral
        | Kind::NoSubstitutionTemplateLiteral
        | Kind::TrueKeyword
        | Kind::FalseKeyword
        | Kind::NullKeyword => true,
        Kind::PropertyAccessExpression => stable(node.as_property_access_expression().expression()),
        Kind::ElementAccessExpression => {
            let access = node.as_element_access_expression();
            stable(access.expression())
                && matches!(
                    unwrap(access.argument_expression()).kind(),
                    Kind::StringLiteral
                        | Kind::NumericLiteral
                        | Kind::NoSubstitutionTemplateLiteral
                )
        }
        _ => false,
    }
}
fn assertion_root(mut node: P<Node>) -> Option<(P<Node>, bool)> {
    let mut awaited = false;
    loop {
        node = ast::skip_parentheses(node);
        node = match node.kind() {
            Kind::CallExpression => return Some((node, awaited)),
            Kind::AwaitExpression => {
                awaited = true;
                node.as_await_expression().expression()
            }
            Kind::PropertyAccessExpression => node.as_property_access_expression().expression(),
            Kind::ElementAccessExpression => node.as_element_access_expression().expression(),
            _ => return None,
        };
    }
}
fn range(node: P<Node>, file: P<SourceFile>) -> [usize; 2] {
    [
        skip_trivia(file.text(), node.pos().expect_source("parsed node")) as usize,
        node.end().expect_source("parsed node") as usize,
    ]
}

#[derive(Clone)]
struct Hit {
    once: bool,
    matcher: String,
    combined: &'static str,
    node: P<Node>,
}
struct Candidate {
    hits: Vec<Hit>,
    statement: P<Node>,
    index: usize,
    arguments: Vec<P<Node>>,
    position: usize,
    key: String,
    fixable: bool,
}
fn message(first: &Hit, second: &Hit) -> String {
    let (once, with) = if first.once {
        (first, second)
    } else {
        (second, first)
    };
    format!(
        "Using `{}` and `{}` on the same target; prefer `{}` instead.",
        once.matcher, with.matcher, with.combined
    )
}
struct Scan<'a> {
    file: P<SourceFile>,
    checker: &'a mut Checker,
    default_library: &'a dyn Fn(P<SourceFile>) -> bool,
    typed: bool,
    nodes: &'a [P<Node>],
    contexts: &'a [(P<Symbol>, bool)],
}
impl Scan<'_> {
    fn parse(&mut self, node: P<Node>) -> Option<ParsedExpect> {
        parse_expect(node, self.file, self.checker, self.contexts, self.nodes)
    }
    fn candidate(&mut self, statement: P<Node>, index: usize) -> Option<Candidate> {
        if statement.kind() != Kind::ExpressionStatement {
            return None;
        }
        let (root, awaited) = assertion_root(statement.as_expression_statement().expression())?;
        let parsed = self.parse(root)?;
        if !matches!(parsed.entry, "call" | "soft")
            || parsed.members.iter().any(|part| part.name == "not")
        {
            return None;
        }
        let head = parsed.head?;
        let mut hits: Vec<Hit> = vec![];
        for matcher in &parsed.matchers {
            let (once, combined) = match matcher.entry.name.as_str() {
                "calledOnce" | "toHaveBeenCalledOnce" => (true, ""),
                "calledWith" => (false, "calledOnceWith"),
                "toHaveBeenCalledWith" => (false, "toHaveBeenCalledExactlyOnceWith"),
                _ => continue,
            };
            if hits.iter().any(|hit| hit.once == once) {
                return None;
            }
            hits.push(Hit {
                once,
                combined,
                matcher: matcher.entry.name.clone(),
                node: matcher.entry.node,
            });
        }
        if hits.is_empty() || head.arguments().is_empty() {
            return None;
        }
        let [start, end] = range(head, self.file);
        let head_text = self.file.text().get(start..end)?;
        if head_text.is_empty() {
            return None;
        }
        let modifiers = parsed
            .modifiers
            .iter()
            .map(|part| part.name.as_str())
            .collect::<Vec<_>>()
            .join(".");
        Some(Candidate {
            hits,
            statement,
            index,
            arguments: head.arguments().to_vec(),
            position: range(statement, self.file)[0],
            key: format!(
                "{head_text}\0{modifiers}{}",
                if awaited { "\0await" } else { "" }
            ),
            fixable: parsed.matchers.len() == 1,
        })
    }
    fn callable(&mut self, node: P<Node>) -> bool {
        if matches!(
            unwrap(node).kind(),
            Kind::FunctionExpression | Kind::ArrowFunction | Kind::ClassExpression
        ) || !self.typed
        {
            return true;
        }
        let ty = self
            .checker
            .get_type_at_location(ast::skip_parentheses(node));
        let parts = if ty.flags().intersects(TypeFlags::Union) {
            ty.types().to_vec()
        } else {
            vec![ty]
        };
        parts.iter().any(|&part| {
            part.flags().intersects(TypeFlags::Any | TypeFlags::Unknown)
                || !self
                    .checker
                    .get_signatures_of_type(part, SignatureKind::Call)
                    .is_empty()
                || !self
                    .checker
                    .get_signatures_of_type(part, SignatureKind::Construct)
                    .is_empty()
        })
    }
    fn library_call(&mut self, node: P<Node>) -> bool {
        if !self.typed || node.kind() != Kind::CallExpression {
            return false;
        }
        let callee = ast::skip_parentheses(node.as_call_expression().expression());
        let Some(symbol) = self.checker.get_symbol_at_location_exported(callee) else {
            return false;
        };
        if matches!(symbol.name(), "eval" | "call" | "apply")
            || !symbol
                .declarations()
                .iter()
                .any(|&decl| ast::get_source_file_of_node(decl).is_some_and(self.default_library))
        {
            return false;
        }
        !node
            .arguments()
            .iter()
            .any(|&argument| self.callable(argument))
    }
    fn inert_subtree(
        &mut self,
        node: P<Node>,
        framework_calls: &HashSet<P<Node>>,
        root: &str,
    ) -> bool {
        if !framework_calls.contains(&node)
            && !helper_call(node, root)
            && !self.library_call(node)
            && control_transfer(node)
        {
            return false;
        }
        !node.for_each_child(&mut |child| !self.inert_subtree(child, framework_calls, root))
    }
    fn inert_declaration(&mut self, statement: P<Node>, names: &HashSet<String>) -> bool {
        if statement.kind() != Kind::VariableStatement {
            return false;
        }
        let list = statement.as_variable_statement().declaration_list();
        let hoisted = !list.flags().intersects(NodeFlags::Const | NodeFlags::Let);
        for &declaration in list.as_variable_declaration_list().declarations().nodes() {
            let Some(name) = declaration.name() else {
                return false;
            };
            if hoisted && contains_name(name, names) || name.kind() != Kind::Identifier {
                return false;
            }
            if let Some(initializer) = declaration.initializer()
                && !self.inert_subtree(initializer, &HashSet::new(), "")
            {
                return false;
            }
        }
        true
    }
    fn inert_statement(&mut self, statement: P<Node>) -> bool {
        if statement.kind() != Kind::ExpressionStatement {
            return false;
        }
        let expression = statement.as_expression_statement().expression();
        let root = assertion_root(expression);
        if root.is_some_and(|(_, awaited)| awaited) {
            return false;
        }
        let parsed = root.and_then(|(root, _)| self.parse(root));
        let Some(parsed) = parsed.filter(|parsed| parsed.head.is_some()) else {
            return self.inert_subtree(expression, &HashSet::new(), "");
        };
        if parsed
            .modifiers
            .iter()
            .any(|part| matches!(part.name.as_str(), "resolves" | "rejects"))
        {
            return false;
        }
        let head = parsed.head.unwrap();
        let mut calls = HashSet::from([head]);
        for matcher in parsed.matchers {
            if invoking_matcher(&matcher.entry.name) {
                return false;
            }
            if let Some(call) = matcher.entry.call {
                calls.insert(call);
            }
        }
        let mut head_entries = vec![];
        expect_entries(
            ast::skip_parentheses(head.as_call_expression().expression()),
            &mut head_entries,
        );
        let root = head_entries.first().map_or("", |entry| entry.name.as_str());
        self.inert_subtree(expression, &calls, root)
    }
    fn inert_between(
        &mut self,
        statements: &[P<Node>],
        first: &Candidate,
        second: &Candidate,
    ) -> bool {
        if second.index - first.index == 1 {
            return true;
        }
        let mut names = HashSet::new();
        identifier_names(first.statement, &mut names);
        identifier_names(second.statement, &mut names);
        statements[first.index + 1..second.index]
            .iter()
            .all(|&statement| {
                self.inert_statement(statement) || self.inert_declaration(statement, &names)
            })
    }
    fn pair(
        &mut self,
        statements: &[P<Node>],
        first: &Candidate,
        second: &Candidate,
    ) -> Option<Violation> {
        let a = &first.hits[0];
        let b = &second.hits[0];
        if a.once == b.once || !self.inert_between(statements, first, second) {
            return None;
        }
        let mut report = Violation::new(b.node, "preferCalledExactlyOnceWith", message(a, b));
        if first.fixable
            && second.fixable
            && first
                .arguments
                .iter()
                .chain(&second.arguments)
                .all(|&argument| stable(argument))
        {
            let (once, with) = if a.once { (first, b) } else { (second, a) };
            if let Some(matcher_range) = replacement_range(with.node, self.file) {
                let remove_range = removal_range(once.statement, self.file);
                if remove_range[0] >= matcher_range[1] || matcher_range[0] >= remove_range[1] {
                    report.edits = vec![
                        Edit::replace(matcher_range, with.combined),
                        Edit::remove(remove_range),
                    ];
                }
            }
        }
        Some(report)
    }
    fn block(&mut self, node: P<Node>) -> Vec<Violation> {
        let statements = node.statements();
        let mut by_key: HashMap<String, Vec<Candidate>> = HashMap::new();
        let mut pending: Vec<(usize, Violation)> = vec![];
        for (index, &statement) in statements.iter().enumerate() {
            let Some(candidate) = self.candidate(statement, index) else {
                continue;
            };
            if candidate.hits.len() == 2 {
                pending.push((
                    candidate.position,
                    Violation::new(
                        candidate.hits[1].node,
                        "preferCalledExactlyOnceWith",
                        message(&candidate.hits[0], &candidate.hits[1]),
                    ),
                ));
            } else {
                by_key
                    .entry(candidate.key.clone())
                    .or_default()
                    .push(candidate);
            }
        }
        for candidates in by_key.values() {
            if candidates.len() == 2
                && let Some(report) = self.pair(statements, &candidates[0], &candidates[1])
            {
                pending.push((candidates[1].position, report));
            }
        }
        pending.sort_by_key(|(position, _)| *position);
        pending.into_iter().map(|(_, report)| report).collect()
    }
}

fn control_transfer(node: P<Node>) -> bool {
    match node.kind() {
        Kind::CallExpression
        | Kind::NewExpression
        | Kind::TaggedTemplateExpression
        | Kind::SpreadElement
        | Kind::AwaitExpression
        | Kind::YieldExpression
        | Kind::DeleteExpression => true,
        Kind::BinaryExpression => {
            ast::is_assignment_operator(node.as_binary_expression().operator_token().kind())
        }
        Kind::PrefixUnaryExpression => matches!(
            node.as_prefix_unary_expression().operator(),
            Kind::PlusPlusToken | Kind::MinusMinusToken
        ),
        Kind::PostfixUnaryExpression => matches!(
            node.as_postfix_unary_expression().operator(),
            Kind::PlusPlusToken | Kind::MinusMinusToken
        ),
        _ => false,
    }
}
fn helper_call(node: P<Node>, root: &str) -> bool {
    if root.is_empty() || node.kind() != Kind::CallExpression {
        return false;
    }
    let mut entries = vec![];
    expect_entries(
        ast::skip_parentheses(node.as_call_expression().expression()),
        &mut entries,
    );
    entries.len() == 2
        && entries[0].name == root
        && entries[0].call.is_none()
        && matches!(
            entries[1].name.as_str(),
            "anything"
                | "any"
                | "stringContaining"
                | "stringMatching"
                | "objectContaining"
                | "arrayContaining"
                | "closeTo"
                | "toBeOneOf"
        )
}
fn invoking_matcher(name: &str) -> bool {
    matches!(
        name,
        "toThrow"
            | "toThrowError"
            | "toThrowErrorMatchingSnapshot"
            | "toThrowErrorMatchingInlineSnapshot"
            | "throw"
            | "throws"
            | "Throw"
            | "change"
            | "changes"
            | "increase"
            | "increases"
            | "decrease"
            | "decreases"
            | "toSatisfy"
            | "satisfy"
            | "satisfies"
    )
}
fn identifier_names(node: P<Node>, into: &mut HashSet<String>) {
    if node.kind() == Kind::Identifier {
        into.insert(node.text().to_owned());
    }
    node.for_each_child(&mut |child| {
        identifier_names(child, into);
        false
    });
}
fn contains_name(node: P<Node>, names: &HashSet<String>) -> bool {
    (node.kind() == Kind::Identifier && names.contains(node.text()))
        || node.for_each_child(&mut |child| contains_name(child, names))
}
fn replacement_range(node: P<Node>, file: P<SourceFile>) -> Option<[usize; 2]> {
    let [start, end] = range(node, file);
    match node.kind() {
        Kind::Identifier => {
            let mut outer = node;
            while let Some(parent) = outer.parent() {
                if parent.kind() == Kind::ParenthesizedExpression {
                    outer = parent;
                    continue;
                }
                if parent.kind() == Kind::ElementAccessExpression
                    && parent.as_element_access_expression().argument_expression() == outer
                {
                    return None;
                }
                break;
            }
            Some([start, end])
        }
        Kind::StringLiteral | Kind::NoSubstitutionTemplateLiteral if end >= start + 2 => {
            Some([start + 1, end - 1])
        }
        _ => None,
    }
}
fn removal_range(statement: P<Node>, file: P<SourceFile>) -> [usize; 2] {
    let original @ [mut start, mut end] = range(statement, file);
    let text = file.text().as_bytes();
    while start > 0 && text[start - 1] != b'\n' {
        if !matches!(text[start - 1], b' ' | b'\t') {
            return original;
        }
        start -= 1;
    }
    while end < text.len() && text[end] != b'\n' {
        if !matches!(text[end], b' ' | b'\t' | b'\r') {
            return original;
        }
        end += 1;
    }
    if end < text.len() {
        end += 1;
    }
    [start, end]
}

#[cfg(test)]
mod tests {
    use super::bracket_opens_on_parenthesis;

    #[test]
    fn frozen_bracket_parenthesis_gate_cases() {
        for text in [
            "expect(spy)[(\"calledOnce\")]",
            "expect(spy)[\n  (\"calledOnce\")\n]",
            "expect(spy)[/* key */ (\"calledOnce\")]",
            "expect(spy)[// key\n(\"calledOnce\")]",
            "expect(spy)[\u{feff}(\"calledOnce\")]",
            "expect(spy)[// key\u{2028}(\"calledOnce\")]",
            "rows[index] && other[(index)]",
        ] {
            assert!(bracket_opens_on_parenthesis(text), "{text}");
        }
        for text in [
            "expect(spy)[\"calledOnce\"]",
            "rows[index]",
            "no brackets at all",
            "trailing bracket [",
            "unterminated block comment [/*",
            "unterminated line comment [//",
        ] {
            assert!(!bracket_opens_on_parenthesis(text), "{text}");
        }
    }
}
