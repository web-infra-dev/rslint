//! Default Rstest assertion completion rules in the frozen Rsbuild workload.
//! Callback ownership and syntactic await handling intentionally follow the
//! frozen implementation, including its observable attribution boundaries.
use super::rsbuild_rstest::{self as calls, Registration};
use super::rsbuild_rstest_expect::{ParsedExpect, parse_expect};
use crate::{Violation, edits::Edit};
use std::collections::{HashMap, HashSet};
use tsrs_ast::{self as ast, Kind, Node, SourceFile, Symbol};
use tsrs_checker::Checker;
use tsrs_core::P;
use tsrs_scanner::skip_trivia;

pub(crate) const NAMES: &[&str] = &[
    "rstest/expect-expect",
    "rstest/no-conditional-expect",
    "rstest/require-local-test-context-for-concurrent-snapshots",
    "rstest/valid-expect",
];

fn start(node: P<Node>, file: P<SourceFile>) -> usize {
    skip_trivia(
        file.text(),
        node.pos().expect_source("parsed source position"),
    ) as usize
}
fn end(node: P<Node>) -> usize {
    node.end().expect_source("parsed source end") as usize
}
fn outer(mut node: P<Node>) -> P<Node> {
    while let Some(parent) = node.parent() {
        if parent.kind() != Kind::ParenthesizedExpression {
            break;
        }
        node = parent;
    }
    node
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

/// Select the `(name, callback[, timeout])` or `(name, options, callback)`
/// function. The production baseline has a RefStore, so unresolved names do
/// not use the scope-blind fallback reserved for parser-only tests.
pub(crate) fn test_callback(
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    nodes: &[P<Node>],
) -> Option<P<Node>> {
    let arguments = node.arguments();
    if arguments.len() < 2 {
        return None;
    }
    if let Some(&third) = arguments.get(2)
        && let Some(callback) = calls::callback(third, file, checker, nodes)
    {
        return Some(callback);
    }
    calls::callback(arguments[1], file, checker, nodes)
}

pub(crate) fn callee_name(node: P<Node>) -> String {
    let node = ast::skip_parentheses(node);
    match node.kind() {
        Kind::Identifier => node.text().to_string(),
        Kind::PropertyAccessExpression => {
            let access = node.as_property_access_expression();
            let left = callee_name(access.expression());
            let key = access.name();
            if left.is_empty() || !matches!(key.kind(), Kind::Identifier | Kind::PrivateIdentifier)
            {
                return left;
            }
            format!("{left}.{}", key.text())
        }
        Kind::ElementAccessExpression => {
            let access = node.as_element_access_expression();
            let left = callee_name(access.expression());
            let key = ast::skip_parentheses(access.argument_expression());
            if left.is_empty()
                || !matches!(
                    key.kind(),
                    Kind::Identifier | Kind::StringLiteral | Kind::NoSubstitutionTemplateLiteral
                )
            {
                return String::new();
            }
            format!("{left}.{}", key.text())
        }
        Kind::CallExpression => callee_name(node.as_call_expression().expression()),
        Kind::NewExpression => callee_name(node.as_new_expression().expression()),
        Kind::TaggedTemplateExpression => callee_name(node.as_tagged_template_expression().tag()),
        _ => String::new(),
    }
}

fn assertion_pattern(node: P<Node>) -> bool {
    let name = callee_name(node);
    let root = name.split('.').next().unwrap_or_default();
    // The frozen default patterns use JavaScript /iu. Long s is the sole
    // non-ASCII simple-fold partner appearing in these two ASCII words.
    let folded: String = root
        .chars()
        .map(|ch| {
            if ch == '\u{017f}' {
                's'
            } else {
                ch.to_ascii_lowercase()
            }
        })
        .collect();
    matches!(folded.as_str(), "expect" | "assert")
}

fn assert_call(
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    nodes: &[P<Node>],
) -> bool {
    let mut entries = vec![];
    calls::expect_entries(node.as_call_expression().expression(), &mut entries);
    let Some(first) = entries.first() else {
        return false;
    };
    let root = first.node;
    if root.kind() != Kind::Identifier {
        return false;
    }
    if first.name == "rstest" && root.parent().is_some_and(calls::import_meta) {
        return entries.get(1).is_some_and(|part| part.name == "assert");
    }
    let symbol = checker.get_symbol_at_location_exported(root);
    if calls::symbol_written(symbol, root.text(), nodes, checker) {
        return false;
    }
    let Some(symbol) = symbol else {
        return root.text() == "assert";
    };
    let mut local = false;
    for &declaration in symbol.declarations() {
        if ast::is_type_only_import_declaration(declaration) {
            continue;
        }
        local |= ast::get_source_file_of_node(declaration) == Some(file);
        if declaration.kind() == Kind::VariableDeclaration
            && declaration.initializer().is_some_and(calls::import_meta)
        {
            return entries.get(1).is_some_and(|part| part.name == "assert");
        }
        if let Some((name, _)) = calls::imported(declaration) {
            if name == "assert" {
                return true;
            }
            if name.is_empty() {
                return entries.get(1).is_some_and(|part| part.name == "assert");
            }
        }
    }
    !local && root.text() == "assert"
}

fn expect_expect(
    nodes: &[P<Node>],
    file: P<SourceFile>,
    checker: &mut Checker,
    contexts: &[(P<Symbol>, bool)],
) -> Vec<Violation> {
    let mut unchecked = vec![];
    let mut by_declaration: HashMap<P<Node>, Vec<P<Node>>> = HashMap::new();
    let mut asserted = HashSet::new();
    let resolve_assertions = file.has_identifier("expect") || file.has_identifier("assert");
    for &node in nodes {
        if node.kind() != Kind::CallExpression {
            continue;
        }
        if let Some(parsed) = calls::registration(node, file, checker).filter(Registration::is_test)
        {
            if parsed.todo {
                continue;
            }
            if let Some(callback) = test_callback(node, file, checker, nodes) {
                if asserted.contains(&callback) {
                    continue;
                }
                by_declaration.entry(callback).or_default().push(node);
            }
            unchecked.push(node);
            continue;
        }
        if !assertion_pattern(node.as_call_expression().expression())
            && !(resolve_assertions
                && (parse_expect(node, file, checker, contexts, nodes).is_some()
                    || assert_call(node, file, checker, nodes)))
        {
            continue;
        }
        let mut ancestors = vec![];
        let mut parent = node.parent();
        while let Some(ancestor) = parent {
            ancestors.push(ancestor);
            parent = ancestor.parent();
        }
        for ancestor in ancestors.into_iter().rev() {
            let declaration = match ancestor.kind() {
                Kind::FunctionDeclaration if ancestor.name().is_some() => Some(ancestor),
                Kind::VariableDeclaration
                    if ancestor
                        .name()
                        .is_some_and(|n| n.kind() == Kind::Identifier) =>
                {
                    ancestor
                        .initializer()
                        .map(ast::skip_parentheses)
                        .filter(|n| {
                            matches!(n.kind(), Kind::FunctionExpression | Kind::ArrowFunction)
                        })
                }
                _ => None,
            };
            if let Some(declaration) = declaration {
                asserted.insert(declaration);
                if let Some(calls) = by_declaration.remove(&declaration) {
                    unchecked.retain(|call| !calls.contains(call));
                }
            }
            if ancestor.kind() == Kind::CallExpression
                && let Some(index) = unchecked.iter().position(|&call| call == ancestor)
            {
                unchecked.remove(index);
                break;
            }
        }
    }
    unchecked
        .into_iter()
        .map(|call| {
            Violation::new(
                call.as_call_expression().expression(),
                "noAssertions",
                "Test has no assertions",
            )
        })
        .collect()
}

fn trailing_name(node: P<Node>) -> &'static str {
    let node = ast::skip_parentheses(node);
    match node.kind() {
        Kind::Identifier => node.text(),
        Kind::PropertyAccessExpression => {
            let name = node.as_property_access_expression().name();
            if name.kind() == Kind::Identifier {
                name.text()
            } else {
                ""
            }
        }
        Kind::ElementAccessExpression => {
            let name =
                ast::skip_parentheses(node.as_element_access_expression().argument_expression());
            if matches!(
                name.kind(),
                Kind::Identifier | Kind::StringLiteral | Kind::NoSubstitutionTemplateLiteral
            ) {
                name.text()
            } else {
                ""
            }
        }
        Kind::CallExpression => trailing_name(node.as_call_expression().expression()),
        Kind::TaggedTemplateExpression => trailing_name(node.as_tagged_template_expression().tag()),
        _ => "",
    }
}
fn conditional(node: P<Node>) -> bool {
    matches!(
        node.kind(),
        Kind::CatchClause | Kind::IfStatement | Kind::SwitchStatement | Kind::ConditionalExpression
    ) || (node.kind() == Kind::BinaryExpression
        && matches!(
            node.as_binary_expression().operator_token().kind(),
            Kind::AmpersandAmpersandToken | Kind::BarBarToken | Kind::QuestionQuestionToken
        ))
}
fn conditional_expect(
    nodes: &[P<Node>],
    file: P<SourceFile>,
    checker: &mut Checker,
    contexts: &[(P<Symbol>, bool)],
) -> Vec<Violation> {
    if !file.has_identifier("expect") {
        return vec![];
    }
    let tests: HashSet<_> = nodes
        .iter()
        .copied()
        .filter(|node| {
            node.kind() == Kind::CallExpression
                && calls::registration(*node, file, checker).is_some_and(|parsed| parsed.is_test())
        })
        .collect();
    let callbacks: HashSet<_> = tests
        .iter()
        .filter_map(|&node| test_callback(node, file, checker, nodes))
        .collect();
    let expects: HashSet<_> = nodes
        .iter()
        .copied()
        .filter(|&node| parse_expect(node, file, checker, contexts, nodes).is_some())
        .collect();
    fn walk(
        node: P<Node>,
        tests: &HashSet<P<Node>>,
        callbacks: &HashSet<P<Node>>,
        expects: &HashSet<P<Node>>,
        mut test: usize,
        mut condition: usize,
        mut catch: usize,
        reports: &mut Vec<Violation>,
    ) {
        test += usize::from(tests.contains(&node)) + usize::from(callbacks.contains(&node));
        if test > 0 && conditional(node) {
            condition += 1;
        }
        if node.kind() == Kind::CallExpression
            && trailing_name(node.as_call_expression().expression()) == "catch"
        {
            catch += 1;
        }
        if expects.contains(&node) {
            for _ in 0..(usize::from(test > 0 && condition > 0) + usize::from(catch > 0)) {
                reports.push(Violation::new(
                    node,
                    "conditionalExpect",
                    "Avoid calling `expect` conditionally",
                ));
            }
        }
        node.for_each_child(&mut |child| {
            walk(
                child, tests, callbacks, expects, test, condition, catch, reports,
            );
            false
        });
    }
    let mut reports = vec![];
    walk(
        file.as_node(),
        &tests,
        &callbacks,
        &expects,
        0,
        0,
        0,
        &mut reports,
    );
    reports
}

type Ownership = HashMap<P<Node>, Vec<(P<Node>, u8)>>;
fn nearest_owned(node: P<Node>, ownership: &Ownership) -> Option<P<Node>> {
    let mut parent = node.parent();
    while let Some(node) = parent {
        if function(node) && ownership.contains_key(&node) {
            return Some(node);
        }
        parent = node.parent();
    }
    None
}
fn concurrent(
    callback: Option<P<Node>>,
    ownership: &Ownership,
    modes: &mut HashMap<P<Node>, u8>,
) -> (bool, bool) {
    let Some(callback) = callback else {
        return (false, true);
    };
    let Some(registrations) = ownership.get(&callback) else {
        return (false, true);
    };
    match modes.get(&callback).copied().unwrap_or_default() {
        1 => return (false, false),
        2 => return (false, true),
        3 => return (true, true),
        _ => (),
    }
    modes.insert(callback, 1);
    let mut resolved = true;
    for &(call, execution_mode) in registrations {
        if execution_mode == 1 {
            modes.insert(callback, 3);
            return (true, true);
        }
        if execution_mode == 2 {
            continue;
        }
        let (outer_concurrent, outer_resolved) =
            concurrent(nearest_owned(call, ownership), ownership, modes);
        if outer_concurrent {
            modes.insert(callback, 3);
            return (true, true);
        }
        resolved &= outer_resolved;
    }
    modes.insert(callback, if resolved { 2 } else { 0 });
    (false, resolved)
}
fn concurrent_snapshots(
    nodes: &[P<Node>],
    file: P<SourceFile>,
    checker: &mut Checker,
    contexts: &[(P<Symbol>, bool)],
) -> Vec<Violation> {
    let mut ownership = Ownership::new();
    for &node in nodes {
        if node.kind() != Kind::CallExpression {
            continue;
        }
        if let Some(parsed) = calls::registration(node, file, checker)
            && let Some(callback) = test_callback(node, file, checker, nodes)
        {
            ownership
                .entry(callback)
                .or_default()
                .push((node, parsed.execution_mode));
        }
    }
    let mut modes = HashMap::new();
    let mut reports = vec![];
    for &node in nodes {
        let Some(parsed) = parse_expect(node, file, checker, contexts, nodes) else {
            continue;
        };
        if parsed.from_test_context
            || !parsed.matchers.iter().any(|matcher| {
                matches!(
                    matcher.entry.name.as_str(),
                    "matchSnapshot"
                        | "toMatchSnapshot"
                        | "toMatchInlineSnapshot"
                        | "toThrowErrorMatchingSnapshot"
                        | "toThrowErrorMatchingInlineSnapshot"
                        | "toMatchFileSnapshot"
                )
            })
        {
            continue;
        }
        if concurrent(nearest_owned(node, &ownership), &ownership, &mut modes).0 {
            reports.push(Violation::new(
                parsed.expression,
                "requireLocalTestContext",
                "Use local Test Context instead",
            ));
        }
    }
    reports
}

fn broken_expect(parsed: &ParsedExpect) -> Option<Violation> {
    let (node, reason, message) = match parsed.reason {
        "matcherNotFound" => (
            parsed.expression,
            "matcherNotFound",
            "Expect must have a corresponding matcher call",
        ),
        "matcherNotCalled" => {
            let last = parsed.members.last();
            let node = last.map_or(parsed.expression, |part| part.node);
            if last.is_some_and(|part| matches!(part.name.as_str(), "not" | "resolves" | "rejects"))
            {
                (
                    node,
                    "matcherNotFound",
                    "Expect must have a corresponding matcher call",
                )
            } else {
                (
                    node,
                    "matcherNotCalled",
                    "Matchers must be called to assert",
                )
            }
        }
        "modifierUnknown" => (
            parsed.expression,
            "modifierUnknown",
            "Expect has an unknown modifier",
        ),
        _ => return None,
    };
    Some(Violation::new(node, reason, message))
}
fn acceptable_return(node: P<Node>) -> bool {
    match node.kind() {
        Kind::ReturnStatement | Kind::ArrowFunction | Kind::AwaitExpression => true,
        Kind::ConditionalExpression => outer(node).parent().is_some_and(acceptable_return),
        _ => false,
    }
}
fn promise_method(node: P<Node>) -> bool {
    if node.kind() != Kind::CallExpression {
        return false;
    }
    let callee = ast::skip_parentheses(node.as_call_expression().expression());
    let receiver = match callee.kind() {
        Kind::PropertyAccessExpression => callee.as_property_access_expression().expression(),
        Kind::ElementAccessExpression => callee.as_element_access_expression().expression(),
        _ => return false,
    };
    callee_name(receiver) == "Promise"
}
fn async_report(mut node: P<Node>) -> (P<Node>, bool, bool, bool) {
    loop {
        let Some(grandparent) = outer(node).parent().and_then(|parent| parent.parent()) else {
            break;
        };
        if grandparent.kind() != Kind::CallExpression {
            break;
        }
        let member = grandparent.as_call_expression().expression();
        if !matches!(
            member.kind(),
            Kind::PropertyAccessExpression | Kind::ElementAccessExpression
        ) {
            break;
        }
        let mut entries = vec![];
        calls::expect_entries(member, &mut entries);
        if !entries
            .last()
            .is_some_and(|last| matches!(last.name.as_str(), "then" | "catch"))
        {
            break;
        }
        node = grandparent;
    }
    let outermost = outer(node);
    let in_array = outermost
        .parent()
        .is_some_and(|parent| parent.kind() == Kind::ArrayLiteralExpression);
    let mut wrapped = false;
    if let Some(mut parent) = outermost.parent().filter(|parent| {
        parent.parent().is_some()
            && matches!(
                parent.kind(),
                Kind::CallExpression | Kind::ArrayLiteralExpression
            )
    }) {
        if parent.kind() == Kind::ArrayLiteralExpression
            && let Some(call) = outer(parent)
                .parent()
                .filter(|call| call.kind() == Kind::CallExpression)
        {
            parent = call;
        }
        if promise_method(parent) {
            node = parent;
            wrapped = true;
        }
    }
    let report = outer(node)
        .parent()
        .is_some_and(|parent| !acceptable_return(parent));
    (node, wrapped, in_array, report)
}
fn async_insert_at(function: P<Node>, file: P<SourceFile>) -> usize {
    if matches!(
        function.kind(),
        Kind::MethodDeclaration | Kind::Constructor | Kind::GetAccessor | Kind::SetAccessor
    ) {
        // The function head starts after decorators but before modifiers. This
        // includes static/get/set exactly as the frozen deferred fixer does.
        if let Some(modifiers) = function.modifiers() {
            let mut last = None;
            for &modifier in modifiers.nodes() {
                if modifier.kind() == Kind::Decorator {
                    last = Some(end(modifier));
                }
            }
            if let Some(last) = last {
                return skip_trivia(file.text(), last as u32) as usize;
            }
        }
    }
    start(function, file)
}
fn valid_expect(
    nodes: &[P<Node>],
    file: P<SourceFile>,
    checker: &mut Checker,
    contexts: &[(P<Symbol>, bool)],
) -> Vec<Violation> {
    let mut reports = vec![];
    let mut descriptors = vec![];
    let mut arrays = HashSet::new();
    for &node in nodes {
        let Some(parsed) = parse_expect(node, file, checker, contexts, nodes) else {
            continue;
        };
        if let Some(report) = broken_expect(&parsed) {
            reports.push(report);
        }
        if !parsed.reason.is_empty() {
            continue;
        }
        let Some(head) = parsed.head else {
            continue;
        };
        let args = head.arguments();
        if args.is_empty() {
            let begin = end(head.as_call_expression().expression());
            let range = file
                .text()
                .as_bytes()
                .get(begin..end(head))
                .and_then(|bytes| bytes.iter().position(|&b| b == b'('))
                .map_or([start(head, file), begin], |offset| {
                    [begin + offset, begin + offset + 1]
                });
            reports.push(
                Violation::new(head, "notEnoughArgs", "Expect requires at least 1 argument")
                    .with_byte_range(range),
            );
        }
        if args.len() > 1 {
            let allowed = args.len() == 2
                && (matches!(parsed.entry, "poll" | "element")
                    || matches!(
                        ast::skip_parentheses(args[1]).kind(),
                        Kind::StringLiteral
                            | Kind::NoSubstitutionTemplateLiteral
                            | Kind::TemplateExpression
                    ));
            if !allowed {
                let begin = start(args[1], file);
                let finish = end(*args.last().expect("nonempty arguments"));
                reports.push(
                    Violation::new(head, "tooManyArgs", "Expect takes at most 1 argument")
                        .with_byte_range([
                            begin,
                            finish.saturating_sub(usize::from(finish > begin)),
                        ]),
                );
            }
        }
        if parsed.matchers.is_empty()
            || !(parsed.modifiers.iter().any(|part| part.name != "not")
                || parsed.matchers.first().is_some_and(|matcher| {
                    matches!(matcher.entry.name.as_str(), "toReject" | "toResolve")
                }))
        {
            continue;
        }
        let (report, wrapped, in_array, required) = async_report(parsed.expression);
        let key = (start(report, file), end(report));
        if required && !arrays.contains(&key) {
            descriptors.push((report, wrapped));
        }
        if in_array {
            arrays.insert(key);
        }
    }
    let mut async_inserted = HashSet::new();
    for (node, wrapped) in descriptors {
        let mut report = if wrapped {
            Violation::new(
                node,
                "promisesWithAsyncAssertionsMustBeAwaited",
                "Promises which return async assertions must be awaited or returned",
            )
        } else {
            Violation::new(
                node,
                "asyncMustBeAwaited",
                "Async assertions must be awaited or returned",
            )
        };
        if let Some(function) = ast::get_containing_function(node) {
            let mut fixes = vec![];
            if !ast::is_async_function(function) && async_inserted.insert(function) {
                let at = async_insert_at(function, file);
                fixes.push(Edit::replace([at, at], "async "));
            }
            let at = start(node, file);
            fixes.push(Edit::replace([at, at], "await "));
            report = report.with_fix(fixes);
        }
        reports.push(report);
    }
    reports
}

pub(crate) fn check(
    name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    has_type_information: bool,
) -> Vec<Violation> {
    if node.kind() != Kind::SourceFile {
        return vec![];
    }
    if name == "rstest/require-local-test-context-for-concurrent-snapshots" && !has_type_information
    {
        return vec![];
    }
    let mut nodes = vec![];
    calls::collect(node, &mut nodes);
    let contexts = calls::context_bindings(&nodes, file, checker);
    match name {
        "rstest/expect-expect" => expect_expect(&nodes, file, checker, &contexts),
        "rstest/no-conditional-expect" => conditional_expect(&nodes, file, checker, &contexts),
        "rstest/require-local-test-context-for-concurrent-snapshots" => {
            concurrent_snapshots(&nodes, file, checker, &contexts)
        }
        "rstest/valid-expect" => valid_expect(&nodes, file, checker, &contexts),
        _ => vec![],
    }
}
