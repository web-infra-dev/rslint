//! Rstest title, expectation and commented registration rules at the frozen baseline.
//! Only the default options enabled by the frozen Rsbuild configuration are exposed.
use super::rsbuild_rstest::{self as calls, Part};
use crate::{Violation, comments, edits::Edit};
use std::collections::HashSet;
use tsrs_ast::{self as ast, CommentRange, Kind, Node, SourceFile, SourceFileParseOptions, Symbol};
use tsrs_checker::Checker;
use tsrs_core::P;
use tsrs_scanner::skip_trivia;

pub(crate) const NAMES: &[&str] = &[
    "rstest/valid-title",
    "rstest/no-identical-title",
    "rstest/no-commented-out-tests",
    "rstest/no-unneeded-async-expect-function",
    "rstest/no-standalone-expect",
];

#[derive(Clone)]
pub(crate) struct ExpectMatcher {
    pub(crate) entry: Part,
    pub(crate) negated: bool,
    pub(crate) is_getter: bool,
}
pub(crate) struct ParsedExpect {
    pub(crate) expression: P<Node>,
    pub(crate) head: Option<P<Node>>,
    pub(crate) entry: &'static str,
    pub(crate) members: Vec<Part>,
    pub(crate) modifiers: Vec<Part>,
    pub(crate) matchers: Vec<ExpectMatcher>,
    pub(crate) valid: bool,
    pub(crate) root_invoked: bool,
    pub(crate) trailing_members: usize,
    pub(crate) reason: &'static str,
    pub(crate) from_test_context: bool,
}

fn expect_chain_root(mut node: P<Node>) -> Option<P<Node>> {
    loop {
        node = ast::skip_parentheses(node);
        node = match node.kind() {
            Kind::Identifier | Kind::MetaProperty => return Some(node),
            Kind::CallExpression => node.as_call_expression().expression(),
            Kind::PropertyAccessExpression => node.as_property_access_expression().expression(),
            Kind::ElementAccessExpression => node.as_element_access_expression().expression(),
            Kind::TaggedTemplateExpression => node.as_tagged_template_expression().tag(),
            // Assertions are boundaries in the frozen expect grammar. A lost
            // receiver must not turn a later member called `expect` into a root.
            _ => return None,
        };
    }
}

fn dynamic_member(mut node: P<Node>) -> bool {
    if node.kind() != Kind::Identifier {
        return false;
    }
    while let Some(parent) = node.parent() {
        if parent.kind() == Kind::ParenthesizedExpression {
            node = parent;
            continue;
        }
        return parent.kind() == Kind::ElementAccessExpression
            && parent.as_element_access_expression().argument_expression() == node;
    }
    false
}

pub(crate) fn parse_expect(
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    contexts: &[(P<Symbol>, bool)],
    nodes: &[P<Node>],
) -> Option<ParsedExpect> {
    if node.kind() != Kind::CallExpression || calls::top_call(node) != node {
        return None;
    }
    expect_chain_root(node.as_call_expression().expression())?;
    let mut expression = node;
    let mut current = node;
    while let Some(parent) = current.parent() {
        let extends = match parent.kind() {
            Kind::ParenthesizedExpression => true,
            Kind::PropertyAccessExpression => {
                parent.as_property_access_expression().expression() == current
            }
            Kind::ElementAccessExpression => {
                parent.as_element_access_expression().expression() == current
            }
            _ => false,
        };
        if !extends {
            break;
        }
        if parent.kind() != Kind::ParenthesizedExpression {
            expression = parent;
        }
        current = parent;
    }
    let mut entries = vec![];
    calls::expect_entries(expression, &mut entries);
    let root = calls::expect_root(&entries, file, checker, contexts, nodes)?;
    let from_test_context = checker
        .get_symbol_at_location_exported(entries[0].node)
        .is_some_and(|symbol| contexts.iter().any(|(binding, _)| *binding == symbol));
    let root_invoked = entries[root].call.is_some();
    let members = entries[root + 1..].to_vec();
    let mut result = ParsedExpect {
        expression,
        head: entries[root].call,
        entry: "call",
        trailing_members: members.len(),
        members,
        modifiers: vec![],
        matchers: vec![],
        valid: false,
        root_invoked,
        reason: "matcherNotFound",
        from_test_context,
    };
    let mut start = 0;
    if !root_invoked {
        let first = result.members.first()?;
        if first.call.is_some() && matches!(first.name.as_str(), "soft" | "poll" | "element") {
            result.entry = match first.name.as_str() {
                "soft" => "soft",
                "poll" => "poll",
                _ => "element",
            };
            result.head = first.call;
            start = 1;
        } else {
            result.entry = "static";
        }
    }
    let chain = &result.members[start..];
    let kinds: Vec<_> = chain
        .iter()
        .enumerate()
        .map(|(index, part)| calls::expect_kind(part, result.entry == "static" && index == 0))
        .collect();
    let first_matcher = kinds.iter().position(|&kind| kind == 1);
    let mut not_count = 0;
    let mut promise_count = 0;
    for (index, (part, &kind)) in chain
        .iter()
        .zip(&kinds)
        .take(first_matcher.unwrap_or(chain.len()))
        .enumerate()
    {
        match kind {
            0 if first_matcher.is_none() && index + 1 == chain.len() => continue,
            2 if part.call.is_none() => continue,
            3 => {
                if part.name == "not" {
                    not_count += 1;
                } else {
                    promise_count += 1;
                }
                if not_count > 1 || promise_count > 1 {
                    result.reason = "modifierUnknown";
                    result.modifiers.clear();
                    return Some(result);
                }
                result.modifiers.push(part.clone());
            }
            _ => {
                result.modifiers.clear();
                result.reason = "modifierUnknown";
                return Some(result);
            }
        }
    }
    let Some(first_matcher) = first_matcher else {
        result.modifiers.clear();
        if !chain.is_empty()
            && matches!(
                expression.kind(),
                Kind::PropertyAccessExpression | Kind::ElementAccessExpression
            )
            && !chain.iter().any(|part| dynamic_member(part.node))
        {
            result.reason = "matcherNotCalled";
        }
        return Some(result);
    };
    let mut negated = not_count > 0;
    for (part, &kind) in chain[first_matcher..].iter().zip(&kinds[first_matcher..]) {
        match kind {
            0 => break,
            1 => result.matchers.push(ExpectMatcher {
                entry: part.clone(),
                negated,
                is_getter: part.call.is_none(),
            }),
            2 if part.call.is_some() => {
                result.modifiers.clear();
                result.matchers.clear();
                result.reason = "modifierUnknown";
                return Some(result);
            }
            3 if part.name == "not" => negated = true,
            _ => {}
        }
    }
    result.valid = true;
    result.reason = "";
    Some(result)
}

fn pos(node: P<Node>) -> u32 {
    node.pos().expect_source("parsed source position")
}
fn end(node: P<Node>) -> u32 {
    node.end().expect_source("parsed source end")
}
fn range(node: P<Node>, file: P<SourceFile>) -> [usize; 2] {
    [
        skip_trivia(file.text(), pos(node)) as usize,
        end(node) as usize,
    ]
}
fn raw(node: P<Node>, file: P<SourceFile>) -> &'static str {
    let [a, b] = range(node, file);
    &file.text()[a..b]
}
fn space(c: char) -> bool {
    matches!(c, '\u{0009}'..='\u{000d}' | '\u{0020}' | '\u{00a0}' | '\u{1680}' | '\u{2000}'..='\u{200a}' | '\u{2028}' | '\u{2029}' | '\u{202f}' | '\u{205f}' | '\u{3000}' | '\u{feff}')
}
fn binary_string(node: P<Node>) -> bool {
    if node.kind() != Kind::BinaryExpression {
        return false;
    }
    let b = node.as_binary_expression();
    let kind = b.operator_token().kind();
    if ast::is_logical_or_coalescing_binary_operator(kind)
        || ast::is_assignment_operator(kind)
        || kind == Kind::CommaToken
    {
        return false;
    }
    ast::is_string_literal_like(b.left())
        || ast::is_string_literal_like(b.right())
        || binary_string(b.left())
}
fn valid_title(node: P<Node>, file: P<SourceFile>, parsed: &calls::Registration) -> Vec<Violation> {
    let Some(&title_node) = node.arguments().first() else {
        return vec![];
    };
    let title = match title_node.kind() {
        Kind::StringLiteral => title_node.text(),
        Kind::NoSubstitutionTemplateLiteral => {
            let raw = raw(title_node, file);
            &raw[1..raw.len() - 1]
        }
        _ => {
            return if title_node.kind() == Kind::TemplateExpression || binary_string(title_node) {
                vec![]
            } else {
                vec![Violation::new(
                    title_node,
                    "titleMustBeString",
                    "Title must be a string",
                )]
            };
        }
    };
    let semantic = if parsed.is_test() {
        parsed.name.as_str()
    } else {
        "describe"
    };
    if title.is_empty() {
        return vec![Violation::new(
            node,
            "emptyTitle",
            format!(
                "{} should not have an empty title",
                if parsed.is_test() { "test" } else { "describe" }
            ),
        )];
    }
    let mut reports = vec![];
    if parsed.parameterized != 0
        && ast::skip_parentheses(node.as_call_expression().expression()).kind()
            == Kind::CallExpression
    {
        let unescaped = title.replace("%%", "");
        let mut chars = unescaped.chars();
        while let Some(c) = chars.next() {
            if c == '%'
                && let Some(spec) = chars.next()
                && !"sdjifoOc#$%".contains(spec)
            {
                reports.push(Violation::new(
                    title_node,
                    "invalidEachSpecifier",
                    format!(
                        "{} is not a valid format specifier",
                        super::go_quote::quote(&format!("%{spec}"))
                    ),
                ));
                break;
            }
        }
    }
    let literal = raw(title_node, file);
    if title.trim_matches(space) != title {
        let inner = &literal[1..literal.len() - 1];
        let replacement = format!(
            "{}{}{}",
            &literal[..1],
            inner.trim_matches(' '),
            &literal[literal.len() - 1..]
        );
        let mut report = Violation::new(
            title_node,
            "accidentalSpace",
            "should not have leading or trailing spaces",
        );
        if replacement != literal {
            report = report.with_fix(vec![Edit::replace(range(title_node, file), replacement)]);
        }
        reports.push(report);
    }
    if title
        .split(' ')
        .next()
        .is_some_and(|first| first.to_lowercase() == semantic)
    {
        let mut report = Violation::new(
            title_node,
            "duplicatePrefix",
            "should not have duplicate prefix",
        );
        let end = 1 + semantic.len();
        if literal.len() >= semantic.len() + 3
            && literal.as_bytes().get(end) == Some(&b' ')
            && literal
                .get(1..end)
                .is_some_and(|name| name.to_lowercase() == semantic)
        {
            report = report.with_fix(vec![Edit::replace(
                range(title_node, file),
                format!("{}{}", &literal[..1], &literal[end + 1..]),
            )]);
        }
        reports.push(report);
    }
    reports
}

fn any(node: P<Node>, predicate: &impl Fn(P<Node>) -> bool) -> bool {
    if predicate(node) {
        return true;
    }
    let mut found = false;
    node.for_each_child(&mut |child| {
        found = any(child, predicate);
        found
    });
    found
}
fn nested_await(node: P<Node>) -> bool {
    if node.kind() == Kind::AwaitExpression {
        return true;
    }
    if ast::is_function_like(node) {
        return node
            .name()
            .is_some_and(|name| name.kind() == Kind::ComputedPropertyName && nested_await(name));
    }
    let mut found = false;
    node.for_each_child(&mut |child| {
        found = nested_await(child);
        found
    });
    found
}
fn async_wrapper(parsed: &ParsedExpect, file: P<SourceFile>) -> Option<Violation> {
    if !parsed.valid
        || !matches!(parsed.entry, "call" | "soft")
        || !parsed
            .modifiers
            .iter()
            .any(|m| matches!(m.name.as_str(), "resolves" | "rejects"))
    {
        return None;
    }
    let head = parsed.head?;
    let wrapper = *head.arguments().first()?;
    let function = ast::skip_parentheses(wrapper);
    if !matches!(
        function.kind(),
        Kind::ArrowFunction | Kind::FunctionExpression
    ) || !ast::is_async_function(function)
    {
        return None;
    }
    let mut expression = function.body()?;
    if expression.kind() == Kind::Block {
        let statements = expression.as_block().statements().nodes();
        if statements.len() != 1 || statements[0].kind() != Kind::ExpressionStatement {
            return None;
        }
        expression = statements[0].as_expression_statement().expression();
    }
    expression = ast::skip_parentheses(expression);
    if expression.kind() != Kind::AwaitExpression {
        return None;
    }
    let awaited = ast::skip_parentheses(expression.as_await_expression().expression());
    if awaited.kind() != Kind::CallExpression {
        return None;
    }
    let mut report = Violation::new(
        wrapper,
        "noAsyncWrapperForExpectedPromise",
        "Avoid wrapping asynchronous expectations in an unnecessary async function.",
    );
    let own_bindings = function.kind() == Kind::FunctionExpression
        && (function.name().is_some()
            || any(awaited, &|child| {
                matches!(child.kind(), Kind::ThisKeyword)
                    || child.kind() == Kind::Identifier && child.text() == "arguments"
                    || child.kind() == Kind::MetaProperty
                        && child.as_meta_property().keyword_token() == Kind::NewKeyword
            }));
    if head.as_call_expression().type_arguments().is_some()
        || !function.parameters().is_empty()
        || !function.type_parameters().is_empty()
        || own_bindings
        || nested_await(awaited)
    {
        return Some(report);
    }
    let [wa, wb] = range(wrapper, file);
    let [aa, ab] = range(awaited, file);
    let mut drops_comment = false;
    let ready = comments::for_each_comment(file, |comment| {
        let a = comment.pos() as usize;
        let b = comment.end() as usize;
        drops_comment |= a < aa && b > wa || a < wb && b > ab;
        Ok(())
    });
    if ready.is_ok() && !drops_comment {
        report = report.with_suggestion(
            "suggestRemoveAsyncWrapper",
            "Pass the awaited call to `expect()` directly",
            vec![Edit::replace([wa, wb], raw(awaited, file))],
        );
    }
    Some(report)
}

fn static_expect(parsed: &ParsedExpect) -> bool {
    let first = parsed
        .matchers
        .first()
        .map(|m| m.entry.name.as_str())
        .unwrap_or("");
    if !parsed.root_invoked
        && parsed.trailing_members == 1
        && !matches!(first, "assertions" | "hasAssertions")
    {
        return true;
    }
    parsed.entry == "static"
        && parsed.modifiers.len() == 1
        && parsed.modifiers[0].name == "not"
        && matches!(
            first,
            "any"
                | "anything"
                | "arrayContaining"
                | "objectContaining"
                | "stringContaining"
                | "stringMatching"
                | "closeTo"
                | "schemaMatching"
                | "toSatisfy"
                | "toBeOneOf"
        )
}
#[derive(Clone, Copy)]
struct Scope {
    kind: u8,
    start: u32,
    end: u32,
}
fn block_kind(node: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> u8 {
    let Some(function) = node.parent() else {
        return 0;
    };
    if !ast::is_function_like(function) {
        return 0;
    }
    if matches!(
        function.kind(),
        Kind::FunctionDeclaration
            | Kind::MethodDeclaration
            | Kind::Constructor
            | Kind::GetAccessor
            | Kind::SetAccessor
    ) || function
        .parent()
        .is_some_and(|p| p.kind() == Kind::VariableDeclaration)
    {
        return 2;
    }
    if let Some(parent) = function.parent()
        && parent.kind() == Kind::CallExpression
        && calls::registration(parent, file, checker).is_some_and(|r| !r.is_test())
    {
        return 3;
    }
    0
}
fn standalone(
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    contexts: &[(P<Symbol>, bool)],
    nodes: &[P<Node>],
    scopes: &mut Vec<Scope>,
    reports: &mut Vec<Violation>,
) {
    let checkpoint = scopes.len();
    if node.kind() == Kind::CallExpression {
        if let Some(parsed) = parse_expect(node, file, checker, contexts, nodes) {
            if !static_expect(&parsed)
                && scopes
                    .iter()
                    .rev()
                    .find(|s| s.start <= pos(node) && pos(node) < s.end)
                    .is_none_or(|s| s.kind == 3)
            {
                reports.push(Violation::new(
                    node,
                    "unexpectedExpect",
                    "Expect must be inside of a test block",
                ));
            }
        } else {
            if calls::registration(node, file, checker).is_some_and(|r| r.is_test()) {
                let body = node.arguments().iter().rev().find_map(|arg| arg.body());
                scopes.push(Scope {
                    kind: 1,
                    start: body.map_or(pos(node), |b| pos(b)),
                    end: body.map_or(pos(node), |b| end(b)),
                });
            }
            if node.as_call_expression().expression().kind() == Kind::TaggedTemplateExpression {
                scopes.push(Scope {
                    kind: 5,
                    start: pos(node),
                    end: end(node),
                });
            }
        }
    } else if node.kind() == Kind::Block {
        let kind = block_kind(node, file, checker);
        if kind != 0 {
            scopes.push(Scope {
                kind,
                start: pos(node),
                end: end(node),
            });
        }
    } else if node.kind() == Kind::ArrowFunction
        && node
            .parent()
            .is_some_and(|p| p.kind() != Kind::CallExpression)
    {
        let body = node.body().unwrap_or(node);
        scopes.push(Scope {
            kind: 4,
            start: pos(body),
            end: end(body),
        });
    }
    node.for_each_child(&mut |child| {
        standalone(child, file, checker, contexts, nodes, scopes, reports);
        false
    });
    scopes.truncate(checkpoint);
}
fn identical(
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    layers: &mut Vec<(HashSet<String>, HashSet<String>)>,
    reports: &mut Vec<Violation>,
) {
    let checkpoint = layers.len();
    if node.kind() == Kind::CallExpression
        && let Some(parsed) = calls::registration(node, file, checker)
    {
        let current = layers.len() - 1;
        if !parsed.is_test() {
            layers.push((HashSet::new(), HashSet::new()));
        }
        if parsed.parameterized == 0
            && let Some(&arg) = node.arguments().first()
            && ast::is_string_literal_like(arg)
        {
            let titles = if parsed.is_test() {
                &mut layers[current].1
            } else {
                &mut layers[current].0
            };
            if !titles.insert(arg.text().to_owned()) {
                reports.push(if parsed.is_test() {
                    Violation::new(
                        arg,
                        "multipleTestTitle",
                        "Test title is used multiple times in the same describe block",
                    )
                } else {
                    Violation::new(
                        arg,
                        "multipleDescribeTitle",
                        "Describe block title is used multiple times in the same describe block",
                    )
                });
            }
        }
    }
    node.for_each_child(&mut |child| {
        identical(child, file, checker, layers, reports);
        false
    });
    layers.truncate(checkpoint);
}

pub(crate) fn check(
    name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    _has_type_information: bool,
) -> Vec<Violation> {
    if node.kind() != Kind::SourceFile {
        return vec![];
    }
    let mut nodes = vec![];
    calls::collect(node, &mut nodes);
    let mut reports = vec![];
    match name {
        "rstest/no-commented-out-tests" => return commented(file),
        "rstest/no-identical-title" => identical(
            node,
            file,
            checker,
            &mut vec![(HashSet::new(), HashSet::new())],
            &mut reports,
        ),
        "rstest/no-standalone-expect" => {
            let contexts = calls::context_bindings(&nodes, file, checker);
            standalone(
                node,
                file,
                checker,
                &contexts,
                &nodes,
                &mut vec![],
                &mut reports,
            );
        }
        "rstest/valid-title" => {
            for &call in &nodes {
                if call.kind() == Kind::CallExpression
                    && let Some(parsed) = calls::registration(call, file, checker)
                {
                    reports.extend(valid_title(call, file, &parsed));
                }
            }
        }
        "rstest/no-unneeded-async-expect-function" => {
            let contexts = calls::context_bindings(&nodes, file, checker);
            for &call in &nodes {
                if let Some(parsed) = parse_expect(call, file, checker, &contexts, &nodes)
                    && let Some(report) = async_wrapper(&parsed, file)
                {
                    reports.push(report);
                }
            }
        }
        _ => {}
    }
    reports
}

fn line_break(text: &str, offset: usize) -> usize {
    let rest = &text[offset..];
    if rest.starts_with("\r\n") {
        2
    } else if rest.starts_with(['\r', '\n']) {
        1
    } else if rest.starts_with(['\u{2028}', '\u{2029}']) {
        3
    } else {
        0
    }
}
fn lines(text: &str) -> Vec<(&str, usize)> {
    let mut result = vec![];
    let mut start = 0;
    let mut offset = 0;
    while offset < text.len() {
        let n = line_break(text, offset);
        if n > 0 {
            result.push((&text[start..offset], n));
            offset += n;
            start = offset;
        } else {
            offset += text[offset..].chars().next().unwrap().len_utf8();
        }
    }
    if start < text.len() {
        result.push((&text[start..], 0));
    }
    result
}
fn comment_body(text: &str, comment: CommentRange) -> Option<String> {
    let a = comment.pos() as usize;
    let b = comment.end() as usize;
    let raw = text.get(a..b)?;
    if comment.kind == Kind::SingleLineCommentTrivia {
        return raw.strip_prefix("//").map(str::to_owned);
    }
    if comment.kind != Kind::MultiLineCommentTrivia || raw.starts_with("/**") {
        return None;
    }
    let body = raw.strip_prefix("/*")?.strip_suffix("*/")?;
    let mut normalized = String::new();
    for (line, newline) in lines(body) {
        let prefix = line.len() - line.trim_start_matches([' ', '\t']).len();
        let mut content = line;
        if line[prefix..].starts_with('*') {
            let mut start = prefix + 1;
            if line[start..].starts_with([' ', '\t']) {
                start += 1;
            }
            content = &line[start..];
        }
        normalized.push_str(content);
        if newline > 0 {
            normalized.push('\n');
        }
    }
    Some(normalized)
}
struct CommentBlock {
    text: String,
    segments: Vec<(usize, CommentRange)>,
}
fn comment_blocks(file: P<SourceFile>) -> Vec<CommentBlock> {
    let mut all = vec![];
    if comments::for_each_comment(file, |comment| {
        all.push(comment);
        Ok(())
    })
    .is_err()
    {
        return vec![];
    }
    all.sort_by_key(|c| (c.pos(), c.end()));
    all.dedup_by_key(|c| (c.pos(), c.end()));
    let text = file.text();
    let mut result = vec![];
    let mut i = 0;
    while i < all.len() {
        let comment = all[i];
        i += 1;
        let Some(body) = comment_body(text, comment) else {
            continue;
        };
        let mut block = CommentBlock {
            text: body,
            segments: vec![(0, comment)],
        };
        if comment.kind == Kind::SingleLineCommentTrivia {
            let mut previous = comment;
            while i < all.len() {
                let next = all[i];
                if next.kind != Kind::SingleLineCommentTrivia || previous.end() > next.pos() {
                    break;
                }
                let gap = &text[previous.end() as usize..next.pos() as usize];
                if !gap.chars().all(space) {
                    break;
                }
                let breaks = lines(gap).iter().filter(|(_, n)| *n > 0).count();
                if breaks == 0 {
                    break;
                }
                let Some(body) = comment_body(text, next) else {
                    break;
                };
                block.text.push_str(&"\n".repeat(breaks));
                block.segments.push((block.text.len(), next));
                block.text.push_str(&body);
                previous = next;
                i += 1;
            }
        }
        result.push(block);
    }
    result
}
#[derive(Clone, Copy)]
struct CommentExpression {
    test: bool,
    phase: u8,
    root: u32,
    extendable: bool,
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
fn commented_expression(node: P<Node>) -> Option<CommentExpression> {
    let node = unwrap(node);
    match node.kind() {
        Kind::Identifier => match node.text() {
            "test" | "it" => Some(CommentExpression {
                test: true,
                phase: 1,
                root: pos(node),
                extendable: true,
            }),
            "describe" => Some(CommentExpression {
                test: false,
                phase: 1,
                root: pos(node),
                extendable: false,
            }),
            _ => None,
        },
        Kind::PropertyAccessExpression | Kind::ElementAccessExpression => {
            let (receiver, name) = if node.kind() == Kind::PropertyAccessExpression {
                let a = node.as_property_access_expression();
                (a.expression(), a.name().text())
            } else {
                let a = node.as_element_access_expression();
                let key = unwrap(a.argument_expression());
                if !ast::is_string_literal_like(key) {
                    return None;
                }
                (a.expression(), key.text())
            };
            let mut result = commented_expression(receiver)?;
            if result.phase != 1 {
                return None;
            }
            match name {
                "only" | "skip" | "todo" | "concurrent" | "sequential" => result.extendable = false,
                "fails" if result.test => result.extendable = false,
                "runIf" | "skipIf" => {
                    result.phase = 2;
                    result.extendable = false;
                }
                "each" | "for" => {
                    result.phase = 3;
                    result.extendable = false;
                }
                "extend" if result.test && result.extendable => result.phase = 4,
                _ => return None,
            }
            Some(result)
        }
        Kind::CallExpression => {
            let call = node.as_call_expression();
            let mut result = commented_expression(call.expression())?;
            match result.phase {
                1 | 5 => {
                    if call.type_arguments().is_some() {
                        return None;
                    }
                    result.phase = 6;
                    result.extendable = false;
                }
                2 => {
                    if call.type_arguments().is_some() {
                        return None;
                    }
                    result.phase = 1;
                    result.extendable = false;
                }
                3 => result.phase = 5,
                4 => {
                    result.phase = 1;
                    result.extendable = true;
                }
                _ => return None,
            }
            Some(result)
        }
        Kind::TaggedTemplateExpression => {
            let mut result = commented_expression(node.as_tagged_template_expression().tag())?;
            if result.phase != 3 {
                return None;
            }
            result.phase = 5;
            Some(result)
        }
        _ => None,
    }
}
fn registration_statement(mut node: P<Node>) -> bool {
    while let Some(parent) = node.parent() {
        match parent.kind() {
            Kind::ParenthesizedExpression
            | Kind::AsExpression
            | Kind::TypeAssertionExpression
            | Kind::NonNullExpression
            | Kind::SatisfiesExpression => node = parent,
            Kind::ExpressionStatement => return true,
            _ => {
                if pos(parent) != pos(node) {
                    return false;
                }
                node = parent;
            }
        }
    }
    false
}
fn comment_window(text: &str, root: usize, file: P<SourceFile>, complete: bool) -> bool {
    let parsed = tsrs_parser::parse_source_file(
        SourceFileParseOptions {
            file_name: "/commented-test".into(),
            path: "/commented-test".into(),
            ..Default::default()
        },
        text,
        file.script_kind(),
    );
    if !complete && !parsed.diagnostics().is_empty() {
        return false;
    }
    let mut nodes = vec![];
    calls::collect(parsed.as_node(), &mut nodes);
    nodes.into_iter().any(|node| {
        if node.kind() != Kind::CallExpression {
            return false;
        }
        let Some(result) = commented_expression(node) else {
            return false;
        };
        if result.phase != 6
            || result.root as usize > root
            || text[result.root as usize..root].contains('\n')
            || !registration_statement(node)
        {
            return false;
        }
        if parsed.diagnostics().iter().any(|d| {
            d.pos().is_some_and(|start| start <= end(node))
                && d.end().is_some_and(|finish| finish >= pos(node))
        }) {
            return false;
        }
        let tail = &text[end(node) as usize..];
        let tail = tail.split('\n').next().unwrap_or("").trim_matches(space);
        tail.is_empty() || tail.starts_with(';') || tail.starts_with("//") || tail.starts_with("/*")
    })
}
fn candidate_root(line: &str) -> Option<usize> {
    let bytes = line.as_bytes();
    let mut i = 0;
    while i < bytes.len() && matches!(bytes[i], b' ' | b'\t') {
        i += 1;
    }
    while i < bytes.len() && matches!(bytes[i], b';' | b'(') {
        i += 1;
    }
    while i < bytes.len() && matches!(bytes[i], b' ' | b'\t') {
        i += 1;
    }
    for name in ["describe", "it", "test"] {
        if line[i..].starts_with(name) {
            let end = i + name.len();
            if bytes
                .get(end)
                .is_none_or(|c| !c.is_ascii_alphanumeric() && *c != b'_')
            {
                return Some(i);
            }
        }
    }
    None
}
fn commented(file: P<SourceFile>) -> Vec<Violation> {
    let mut reports = vec![];
    for block in comment_blocks(file) {
        let mut offset = 0;
        let mut fenced = false;
        let mut reported = HashSet::new();
        for line in block.text.split_inclusive('\n') {
            let trimmed = line.trim_matches(space);
            if trimmed.starts_with("```") || trimmed.starts_with("~~~") {
                fenced = !fenced;
                offset += line.len();
                continue;
            }
            if !fenced && let Some(root) = candidate_root(line) {
                let candidate = &block.text[offset..];
                let mut count = 1;
                let mut found = false;
                loop {
                    let end = candidate
                        .match_indices('\n')
                        .nth(count - 1)
                        .map_or(candidate.len(), |(i, _)| i + 1);
                    let complete = end >= candidate.len();
                    if end >= root && comment_window(&candidate[..end], root, file, complete) {
                        found = true;
                        break;
                    }
                    if complete {
                        break;
                    }
                    count *= 2;
                }
                if found {
                    let absolute = offset + root;
                    if let Some(&(_, comment)) = block
                        .segments
                        .iter()
                        .rev()
                        .find(|(start, _)| *start <= absolute)
                        && reported.insert((comment.pos(), comment.end()))
                    {
                        reports.push(
                            Violation::new(
                                file.as_node(),
                                "commentedTests",
                                "Do not comment out tests",
                            )
                            .with_byte_range([comment.pos() as usize, comment.end() as usize]),
                        );
                    }
                }
            }
            offset += line.len();
        }
    }
    reports
}
