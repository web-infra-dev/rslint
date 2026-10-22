//! Rstest Promise assertion consumption at rslint e0a29f751d4def86d27931759098f27748fa3003.
//! Copyright (c) 2025 July-present Bytedance Inc and its affiliates.
//! Copyright (c) 2025 July typescript-eslint and other contributors.
//! MIT licensed; see LICENSE. Handler assertions flow over a finite graph;
//! binding consumption is solved on the actual per-function control-flow graph.

use std::collections::{HashMap, HashSet, VecDeque};
use tsrs_ast::{self as ast, Kind, Node, SourceFile, Symbol};
use tsrs_checker::Checker;
use tsrs_core::{P, jsnum};

use super::{
    rsbuild_cfg::{PromiseEvent, promise_flow},
    rsbuild_rstest as calls, rsbuild_rstest_completion as completion,
    rsbuild_rstest_expect as expect,
};
use crate::Violation;

pub(crate) const NAMES: &[&str] = &["rstest/valid-expect-in-promise"];

pub(crate) fn check(
    name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
) -> Vec<Violation> {
    if name != NAMES[0] || node.kind() != Kind::SourceFile {
        return Vec::new();
    }
    let mut nodes = Vec::new();
    calls::collect(node, &mut nodes);
    // This exact AST gate also recognizes escaped and parenthesized literal keys.
    if !nodes.iter().any(|&node| promise_chain(node)) {
        return Vec::new();
    }
    let contexts = calls::context_bindings(&nodes, file, checker);
    let callbacks = nodes
        .iter()
        .filter_map(|&node| {
            calls::registration(node, file, checker)
                .filter(|registration| registration.is_test())?;
            completion::test_callback(node, file, checker, &nodes)
        })
        .collect::<HashSet<_>>();
    let mut analyzer = Analyzer {
        file,
        checker,
        nodes: &nodes,
        contexts: &contexts,
        functions: Vec::new(),
        function_ids: HashMap::new(),
        groups: Vec::new(),
        group_ids: HashMap::new(),
    };
    // Source order makes graph collection deterministic, while each callback is
    // still summarized only once when several test registrations share it.
    for &node in &nodes {
        if callbacks.contains(&node) {
            analyzer.collect_function(node, None);
        }
    }
    analyzer.resolve_consumption();
    analyzer.propagate_and_report()
}

struct Group {
    root: P<Node>,
    function: usize,
    assertion: bool,
    safe: bool,
    reachable: bool,
    binding: Option<Binding>,
    report: P<Node>,
}
struct Function {
    node: P<Node>,
    owners: HashSet<usize>,
    direct_assertion: bool,
    assertion: bool,
}
#[derive(Clone, Copy)]
struct Binding {
    target: P<Node>,
    symbol: P<Symbol>,
    extends: bool,
    aggregate: bool,
}
#[derive(Clone, Copy)]
enum Event {
    Bind(P<Symbol>, usize, bool),
    Overwrite(P<Symbol>),
    Consume(P<Symbol>, bool),
    Failure,
}
struct Analyzer<'a> {
    file: P<SourceFile>,
    checker: &'a mut Checker,
    nodes: &'a [P<Node>],
    contexts: &'a [(P<Symbol>, bool)],
    functions: Vec<Function>,
    function_ids: HashMap<P<Node>, usize>,
    groups: Vec<Group>,
    group_ids: HashMap<P<Node>, usize>,
}

impl Analyzer<'_> {
    fn collect_function(&mut self, node: P<Node>, owner: Option<usize>) {
        if let Some(&id) = self.function_ids.get(&node) {
            if let Some(owner) = owner {
                self.functions[id].owners.insert(owner);
            }
            return;
        }
        let id = self.functions.len();
        let mut owners = HashSet::new();
        if let Some(owner) = owner {
            owners.insert(owner);
        }
        // Publish before walking so recursive handlers add graph edges rather
        // than recursively collecting their bodies again.
        self.functions.push(Function {
            node,
            owners,
            direct_assertion: false,
            assertion: false,
        });
        self.function_ids.insert(node, id);
        if let Some(body) = node.body() {
            self.walk(body, id);
        }
    }

    fn walk(&mut self, node: P<Node>, function: usize) {
        if node != self.functions[function].node && is_function(node) {
            return;
        }
        if node.kind() == Kind::CallExpression {
            if promise_chain(node) {
                let group = self.group(node, function);
                let call = node.as_call_expression();
                self.walk(call.expression(), function);
                for &argument in call.arguments.nodes() {
                    if let Some(handler) = self.resolve_function(argument) {
                        self.collect_function(handler, Some(group));
                    } else {
                        self.walk(argument, function);
                    }
                }
                return;
            }
            if self.assertion_call(node) {
                self.functions[function].direct_assertion = true;
            }
        }
        node.for_each_child(&mut |child| {
            self.walk(child, function);
            false
        });
    }

    fn group(&mut self, node: P<Node>, function: usize) -> usize {
        let root = top_chain(node);
        if let Some(&id) = self.group_ids.get(&root) {
            return id;
        }
        let id = self.groups.len();
        self.groups.push(Group {
            root,
            function,
            assertion: false,
            safe: false,
            reachable: false,
            binding: None,
            report: report_node(root, self.functions[function].node),
        });
        self.group_ids.insert(root, id);
        id
    }

    fn resolve_function(&mut self, node: P<Node>) -> Option<P<Node>> {
        let node = ast::skip_parentheses(node);
        if matches!(node.kind(), Kind::FunctionExpression | Kind::ArrowFunction) {
            return Some(node);
        }
        if node.kind() != Kind::Identifier {
            return None;
        }
        let symbol = self.checker.get_symbol_at_location_exported(node)?;
        let declaration = *symbol.declarations().first()?;
        match declaration.kind() {
            Kind::FunctionDeclaration => Some(declaration),
            Kind::VariableDeclaration => declaration
                .initializer()
                .map(ast::skip_parentheses)
                .filter(|node| {
                    matches!(node.kind(), Kind::FunctionExpression | Kind::ArrowFunction)
                }),
            _ => None,
        }
    }

    fn assertion_call(&mut self, node: P<Node>) -> bool {
        expect::parse_expect(node, self.file, self.checker, self.contexts, self.nodes).is_some()
            || rstest_assert(node, self.file, self.checker)
    }

    fn async_sink(&mut self, call: P<Node>, value: P<Node>) -> bool {
        if call.kind() != Kind::CallExpression {
            return false;
        }
        let Some(parsed) = expect::parse_expect(
            calls::top_call(call),
            self.file,
            self.checker,
            self.contexts,
            self.nodes,
        ) else {
            return false;
        };
        if !parsed
            .modifiers
            .iter()
            .any(|part| matches!(part.name.as_str(), "resolves" | "rejects"))
        {
            return false;
        }
        parsed
            .head
            .and_then(|head| head.arguments().first().copied())
            .is_some_and(|argument| ast::skip_parentheses(argument) == ast::skip_parentheses(value))
    }

    fn symbol(&mut self, node: P<Node>) -> Option<P<Symbol>> {
        // A shorthand's name represents both an object property and a lexical
        // binding. The checker location API returns the property symbol; the
        // frozen reference store instead resolves the value binding. Use that
        // binding for both destructuring writes and their later consumption.
        if let Some(parent) = node.parent().filter(|parent| {
            parent.kind() == Kind::ShorthandPropertyAssignment && parent.name() == Some(node)
        }) {
            return self
                .checker
                .get_shorthand_assignment_value_symbol(Some(parent));
        }
        node.symbol()
            .or_else(|| self.checker.get_symbol_at_location_exported(node))
    }

    fn preserves_symbol(&mut self, value: P<Node>, symbol: P<Symbol>) -> bool {
        let value = ast::skip_parentheses(value);
        match value.kind() {
            Kind::Identifier => self.symbol(value) == Some(symbol),
            Kind::ConditionalExpression => {
                let conditional = value.as_conditional_expression();
                self.preserves_symbol(conditional.when_true(), symbol)
                    && self.preserves_symbol(conditional.when_false(), symbol)
            }
            Kind::BinaryExpression => {
                let binary = value.as_binary_expression();
                match binary.operator_token().kind() {
                    Kind::BarBarToken | Kind::QuestionQuestionToken => {
                        self.preserves_symbol(binary.left(), symbol)
                    }
                    Kind::CommaToken => self.preserves_symbol(binary.right(), symbol),
                    _ => false,
                }
            }
            _ => false,
        }
    }

    fn write_preserves(&mut self, node: P<Node>) -> bool {
        let mut current = node;
        while let Some(parent) = current.parent() {
            if parent.kind() == Kind::ParenthesizedExpression {
                current = parent;
                continue;
            }
            if !ast::is_assignment_expression(parent, false) {
                return false;
            }
            let binary = parent.as_binary_expression();
            if ast::skip_parentheses(binary.left()) != ast::skip_parentheses(current) {
                return false;
            }
            return match binary.operator_token().kind() {
                Kind::BarBarEqualsToken | Kind::QuestionQuestionEqualsToken => true,
                Kind::EqualsToken => self
                    .symbol(node)
                    .is_some_and(|symbol| self.preserves_symbol(binary.right(), symbol)),
                _ => false,
            };
        }
        false
    }

    fn binding(&mut self, root: P<Node>) -> Option<Binding> {
        if let Some(target) = static_binding(root) {
            return self.symbol(target).map(|symbol| Binding {
                target,
                symbol,
                extends: false,
                aggregate: false,
            });
        }
        let mut current = root;
        let mut aggregate = false;
        while let Some(parent) = current.parent() {
            match parent.kind() {
                Kind::ParenthesizedExpression => current = parent,
                Kind::VariableDeclaration => {
                    if parent.initializer().map(ast::skip_parentheses)
                        != Some(ast::skip_parentheses(current))
                    {
                        return None;
                    }
                    let target = ast::skip_parentheses(parent.name()?);
                    if target.kind() != Kind::Identifier {
                        return None;
                    }
                    return self.symbol(target).map(|symbol| Binding {
                        target,
                        symbol,
                        extends: false,
                        aggregate,
                    });
                }
                Kind::ArrayLiteralExpression => {
                    aggregate = true;
                    current = parent;
                }
                Kind::ConditionalExpression => {
                    let conditional = parent.as_conditional_expression();
                    if conditional.when_true() != current && conditional.when_false() != current {
                        return None;
                    }
                    current = parent;
                }
                _ if ast::is_assignment_expression(parent, false) => {
                    let binary = parent.as_binary_expression();
                    let operator = binary.operator_token().kind();
                    if operator != Kind::EqualsToken && !logical_assignment(operator)
                        || ast::skip_parentheses(binary.right()) != ast::skip_parentheses(current)
                    {
                        return None;
                    }
                    let target = ast::skip_parentheses(binary.left());
                    if target.kind() != Kind::Identifier {
                        return None;
                    }
                    let symbol = self.symbol(target)?;
                    let starts = first_identifier(root.as_call_expression().expression())
                        .is_some_and(|first| self.symbol(first) == Some(symbol));
                    let extends = matches!(
                        operator,
                        Kind::BarBarEqualsToken | Kind::QuestionQuestionEqualsToken
                    ) || starts
                        || self.preserves_symbol(binary.right(), symbol);
                    return Some(Binding {
                        target,
                        symbol,
                        extends,
                        aggregate,
                    });
                }
                _ if flows_binary(parent, current) => current = parent,
                _ => return None,
            }
        }
        None
    }

    fn directly_consumed(&mut self, root: P<Node>, function: P<Node>) -> bool {
        if function.kind() == Kind::ArrowFunction
            && function.body().map(ast::skip_parentheses) == Some(ast::skip_parentheses(root))
        {
            return true;
        }
        let mut current = root;
        while let Some(parent) = current.parent().filter(|&parent| parent != function) {
            match parent.kind() {
                Kind::ParenthesizedExpression => current = parent,
                Kind::AwaitExpression | Kind::ReturnStatement => {
                    return propagates_failure(parent, function);
                }
                Kind::ArrowFunction => return parent.body() == Some(current),
                Kind::ArrayLiteralExpression => {
                    return aggregate_literal(parent, function) || for_await(parent, function);
                }
                Kind::CallExpression => {
                    return self.async_sink(parent, root) || wrapper(parent, current, function);
                }
                Kind::ConditionalExpression => {
                    let conditional = parent.as_conditional_expression();
                    if conditional.when_true() != current && conditional.when_false() != current {
                        return false;
                    }
                    current = parent;
                }
                Kind::BinaryExpression if flows_binary(parent, current) => current = parent,
                _ => return false,
            }
        }
        function.kind() == Kind::ArrowFunction
            && function.body().map(ast::skip_parentheses) == Some(ast::skip_parentheses(current))
    }

    fn safe_identifier(&mut self, identifier: P<Node>, function: P<Node>) -> (bool, bool) {
        if identifier.kind() != Kind::Identifier {
            return (false, false);
        }
        let mut current = identifier;
        let mut element = false;
        while let Some(parent) = current.parent() {
            match parent.kind() {
                Kind::ParenthesizedExpression => current = parent,
                Kind::AwaitExpression | Kind::ReturnStatement => {
                    return (propagates_failure(parent, function), element);
                }
                Kind::ForOfStatement => return (for_await(current, function), true),
                Kind::ArrayLiteralExpression => {
                    return (
                        aggregate_literal(parent, function) || for_await(parent, function),
                        false,
                    );
                }
                Kind::SpreadElement => {
                    let Some(array) = parent
                        .parent()
                        .filter(|node| node.kind() == Kind::ArrayLiteralExpression)
                    else {
                        return (false, false);
                    };
                    return if aggregate_literal(array, function) || for_await(array, function) {
                        (true, true)
                    } else {
                        (false, false)
                    };
                }
                Kind::CallExpression => {
                    if self.async_sink(parent, identifier) || wrapper(parent, current, function) {
                        return (true, false);
                    }
                    return if aggregate_value(parent, current, function) {
                        (true, true)
                    } else {
                        (false, false)
                    };
                }
                Kind::ConditionalExpression => {
                    let conditional = parent.as_conditional_expression();
                    if conditional.when_true() != current && conditional.when_false() != current {
                        return (false, false);
                    }
                    current = parent;
                }
                Kind::BinaryExpression if flows_binary(parent, current) => current = parent,
                Kind::ElementAccessExpression => {
                    element |= parent.as_element_access_expression().expression() == current;
                    current = parent;
                }
                Kind::PropertyAccessExpression => current = parent,
                _ => return (false, false),
            }
        }
        (false, false)
    }

    fn resolve_consumption(&mut self) {
        for function in 0..self.functions.len() {
            let ids = (0..self.groups.len())
                .filter(|&id| self.groups[id].function == function)
                .collect::<Vec<_>>();
            if ids.is_empty() {
                continue;
            }
            self.function_consumption(function, &ids);
        }
    }

    fn function_consumption(&mut self, function_id: usize, ids: &[usize]) {
        let function = self.functions[function_id].node;
        let mut roots: HashMap<P<Node>, Vec<usize>> = HashMap::new();
        let mut targets: HashMap<P<Node>, Vec<usize>> = HashMap::new();
        for &id in ids {
            roots.entry(self.groups[id].root).or_default().push(id);
            self.groups[id].binding = self.binding(self.groups[id].root);
            if let Some(binding) = self.groups[id].binding {
                targets.entry(binding.target).or_default().push(id);
                self.groups[id].safe = true;
            }
        }
        let graph = promise_flow(function);
        let mut events = vec![Vec::new(); graph.blocks.len()];
        for (index, block) in graph.blocks.iter().enumerate() {
            for &event in &block.events {
                match event {
                    PromiseEvent::Expression(node) if block.reachable => {
                        if let Some(ids) = roots.get(&node) {
                            for &id in ids {
                                self.groups[id].reachable = true;
                            }
                        }
                    }
                    PromiseEvent::Read(node) => {
                        let (safe, aggregate) = self.safe_identifier(node, function);
                        if safe && let Some(symbol) = self.symbol(node) {
                            events[index].push(Event::Consume(symbol, aggregate));
                        }
                    }
                    PromiseEvent::Write(node) => {
                        let Some(symbol) = self.symbol(node) else {
                            continue;
                        };
                        if let Some(ids) = targets.get(&node) {
                            for (position, &id) in ids.iter().enumerate() {
                                events[index].push(Event::Bind(
                                    self.groups[id].binding.unwrap().symbol,
                                    id,
                                    self.groups[id].binding.unwrap().extends || position > 0,
                                ));
                            }
                        } else if !self.write_preserves(node) {
                            events[index].push(Event::Overwrite(symbol));
                        }
                    }
                    PromiseEvent::Statement(node)
                        if node.kind() == Kind::ThrowStatement
                            && propagates_failure(node, function) =>
                    {
                        events[index].push(Event::Failure)
                    }
                    _ => {}
                }
            }
        }
        type State = HashMap<P<Symbol>, HashSet<usize>>;
        let mut incoming: Vec<Option<State>> = vec![None; graph.blocks.len()];
        incoming[0] = Some(HashMap::new());
        let mut queue = VecDeque::from([0]);
        let mut queued = vec![false; graph.blocks.len()];
        queued[0] = true;
        while let Some(index) = queue.pop_front() {
            queued[index] = false;
            let block = &graph.blocks[index];
            if !block.reachable {
                continue;
            }
            let mut state = incoming[index].clone().unwrap_or_default();
            for &event in &events[index] {
                match event {
                    Event::Failure => state.clear(),
                    Event::Consume(symbol, aggregate) => {
                        if let Some(candidates) = state.get_mut(&symbol) {
                            candidates.retain(|&id| {
                                self.groups[id]
                                    .binding
                                    .is_some_and(|binding| binding.aggregate != aggregate)
                            });
                            if candidates.is_empty() {
                                state.remove(&symbol);
                            }
                        }
                    }
                    Event::Overwrite(symbol) => {
                        if let Some(candidates) = state.remove(&symbol) {
                            for id in candidates {
                                self.groups[id].safe = false;
                            }
                        }
                    }
                    Event::Bind(symbol, id, extends) => {
                        if !extends && let Some(candidates) = state.remove(&symbol) {
                            for id in candidates {
                                self.groups[id].safe = false;
                            }
                        }
                        state.entry(symbol).or_default().insert(id);
                    }
                }
            }
            let mut successor = false;
            for &next in &block.successors {
                if !graph.blocks[next].reachable {
                    continue;
                }
                successor = true;
                let first = incoming[next].is_none();
                let target = incoming[next].get_or_insert_with(HashMap::new);
                let mut changed = false;
                for (&symbol, candidates) in &state {
                    let target = target.entry(symbol).or_default();
                    for &id in candidates {
                        changed |= target.insert(id);
                    }
                }
                if (first || changed) && !queued[next] {
                    queue.push_back(next);
                    queued[next] = true;
                }
            }
            if !successor {
                for candidates in state.values() {
                    for &id in candidates {
                        self.groups[id].safe = false;
                    }
                }
            }
        }
        for &id in ids {
            if !self.groups[id].reachable {
                self.groups[id].safe = true;
            } else if !self.groups[id].safe {
                self.groups[id].safe = self.directly_consumed(self.groups[id].root, function);
            }
        }
    }

    fn propagate_and_report(mut self) -> Vec<Violation> {
        let mut functions = VecDeque::new();
        let mut groups = VecDeque::new();
        for (id, function) in self.functions.iter_mut().enumerate() {
            if function.direct_assertion {
                function.assertion = true;
                functions.push_back(id);
            }
        }
        while !functions.is_empty() || !groups.is_empty() {
            if let Some(id) = functions.pop_front() {
                for &owner in &self.functions[id].owners {
                    if !self.groups[owner].assertion {
                        self.groups[owner].assertion = true;
                        groups.push_back(owner);
                    }
                }
                continue;
            }
            let id = groups.pop_front().unwrap();
            let function = self.groups[id].function;
            if self.groups[id].safe && !self.functions[function].assertion {
                self.functions[function].assertion = true;
                functions.push_back(function);
            }
        }
        self.groups.into_iter().rev().filter(|group|group.assertion && !group.safe).map(|group|Violation::new(group.report,
            "expectInFloatingPromise", "This promise should either be returned or awaited to ensure the assertions in its chain are called")).collect()
    }
}

fn is_function(node: P<Node>) -> bool {
    matches!(
        node.kind(),
        Kind::ArrowFunction
            | Kind::FunctionExpression
            | Kind::FunctionDeclaration
            | Kind::MethodDeclaration
            | Kind::GetAccessor
            | Kind::SetAccessor
            | Kind::Constructor
    )
}
fn promise_chain(node: P<Node>) -> bool {
    if node.kind() != Kind::CallExpression {
        return false;
    }
    let call = node.as_call_expression();
    let callee = ast::skip_parentheses(call.expression());
    let name = match callee.kind() {
        Kind::PropertyAccessExpression => callee
            .name()
            .filter(|name| name.kind() == Kind::Identifier)
            .map(|name| name.text()),
        Kind::ElementAccessExpression => {
            let key =
                ast::skip_parentheses(callee.as_element_access_expression().argument_expression());
            matches!(
                key.kind(),
                Kind::StringLiteral | Kind::NoSubstitutionTemplateLiteral
            )
            .then(|| key.text())
        }
        _ => None,
    };
    let count = call.arguments.nodes().len();
    count > 0
        && match name {
            Some("then") => count < 3,
            Some("catch" | "finally") => count < 2,
            _ => false,
        }
}
fn top_chain(node: P<Node>) -> P<Node> {
    let mut top = node;
    let mut current = node;
    while let Some(parent) = current.parent() {
        if parent.kind() == Kind::ParenthesizedExpression {
            current = parent;
            continue;
        }
        if !ast::is_access_expression(parent) || parent.expression() != Some(current) {
            break;
        }
        let Some(call) = parent.parent().filter(|call| {
            call.kind() == Kind::CallExpression
                && call.expression() == Some(parent)
                && promise_chain(*call)
        }) else {
            break;
        };
        top = call;
        current = call;
    }
    top
}
fn report_node(root: P<Node>, function: P<Node>) -> P<Node> {
    let mut current = root;
    while let Some(parent) = current.parent().filter(|&parent| parent != function) {
        match parent.kind() {
            Kind::ParenthesizedExpression
            | Kind::CallExpression
            | Kind::PropertyAccessExpression
            | Kind::ElementAccessExpression => current = parent,
            Kind::ExpressionStatement | Kind::VariableDeclaration => return parent,
            _ => {
                if ast::is_assignment_expression(parent, false)
                    && (parent.as_binary_expression().operator_token().kind() == Kind::EqualsToken
                        || logical_assignment(
                            parent.as_binary_expression().operator_token().kind(),
                        ))
                {
                    return parent;
                }
                return current;
            }
        }
    }
    current
}
fn first_identifier(mut node: P<Node>) -> Option<P<Node>> {
    loop {
        node = ast::skip_parentheses(node);
        node = match node.kind() {
            Kind::Identifier => return Some(node),
            Kind::CallExpression
            | Kind::PropertyAccessExpression
            | Kind::ElementAccessExpression => node.expression()?,
            Kind::TaggedTemplateExpression => node.as_tagged_template_expression().tag(),
            Kind::BinaryExpression
                if node.as_binary_expression().operator_token().kind() == Kind::CommaToken =>
            {
                node.as_binary_expression().right()
            }
            _ => return None,
        };
    }
}
fn logical_assignment(kind: Kind) -> bool {
    matches!(
        kind,
        Kind::BarBarEqualsToken
            | Kind::AmpersandAmpersandEqualsToken
            | Kind::QuestionQuestionEqualsToken
    )
}
fn flows_binary(parent: P<Node>, current: P<Node>) -> bool {
    if parent.kind() != Kind::BinaryExpression {
        return false;
    }
    let binary = parent.as_binary_expression();
    match binary.operator_token().kind() {
        Kind::BarBarToken | Kind::QuestionQuestionToken => {
            binary.left() == current || binary.right() == current
        }
        Kind::AmpersandAmpersandToken | Kind::CommaToken => binary.right() == current,
        operator => logical_assignment(operator) && binary.right() == current,
    }
}
fn chain_name(node: P<Node>) -> String {
    let node = ast::skip_parentheses(node);
    match node.kind() {
        Kind::Identifier => node.text().into(),
        Kind::PropertyAccessExpression => {
            let left = chain_name(node.expression().unwrap());
            let key = static_name(node.name().unwrap()).unwrap_or_default();
            if left.is_empty() || key.is_empty() {
                left
            } else {
                format!("{left}.{key}")
            }
        }
        Kind::ElementAccessExpression => {
            let left = chain_name(node.expression().unwrap());
            let key = literal_name(node.as_element_access_expression().argument_expression())
                .unwrap_or_default();
            if left.is_empty() || key.is_empty() {
                String::new()
            } else {
                format!("{left}.{key}")
            }
        }
        Kind::CallExpression | Kind::NewExpression => chain_name(node.expression().unwrap()),
        Kind::TaggedTemplateExpression => chain_name(node.as_tagged_template_expression().tag()),
        _ => String::new(),
    }
}
fn awaited_or_returned(mut node: P<Node>, function: P<Node>) -> bool {
    while let Some(parent) = node.parent() {
        if parent.kind() == Kind::ParenthesizedExpression {
            node = parent;
            continue;
        }
        return matches!(parent.kind(), Kind::AwaitExpression | Kind::ReturnStatement)
            && propagates_failure(parent, function)
            || parent.kind() == Kind::ArrowFunction && parent.body() == Some(node);
    }
    false
}
fn call_consumes(call: P<Node>, value: P<Node>, function: P<Node>, name: &str) -> bool {
    call.kind() == Kind::CallExpression
        && call.arguments().len() == 1
        && ast::skip_parentheses(call.arguments()[0]) == ast::skip_parentheses(value)
        && chain_name(call.expression().unwrap()) == name
        && awaited_or_returned(call, function)
}
fn wrapper(call: P<Node>, value: P<Node>, function: P<Node>) -> bool {
    call_consumes(call, value, function, "Promise.resolve")
}
fn aggregate_value(call: P<Node>, value: P<Node>, function: P<Node>) -> bool {
    call_consumes(call, value, function, "Promise.all")
}
fn aggregate_literal(array: P<Node>, function: P<Node>) -> bool {
    array.kind() == Kind::ArrayLiteralExpression
        && array
            .parent()
            .is_some_and(|call| aggregate_value(call, array, function))
}
fn for_await(mut node: P<Node>, function: P<Node>) -> bool {
    while let Some(parent) = node.parent() {
        if parent.kind() == Kind::ParenthesizedExpression {
            node = parent;
            continue;
        }
        return parent.kind() == Kind::ForOfStatement
            && parent
                .as_for_in_or_of_statement()
                .await_modifier()
                .is_some()
            && parent.expression() == Some(node)
            && propagates_failure(parent, function);
    }
    false
}

fn propagates_failure(mut node: P<Node>, boundary: P<Node>) -> bool {
    let catch_suppresses = matches!(
        node.kind(),
        Kind::AwaitExpression | Kind::ThrowStatement | Kind::ForOfStatement
    );
    while let Some(parent) = node.parent().filter(|&parent| parent != boundary) {
        if parent.kind() == Kind::TryStatement {
            let statement = parent.as_try_statement();
            if let Some(finally) = statement.finally_block().filter(|&finally| finally != node)
                && may_suppress(finally)
            {
                return false;
            }
            if catch_suppresses
                && statement.try_block() == node
                && let Some(catch) = statement.catch_clause()
                && !always_throws(catch.as_catch_clause().block())
            {
                return false;
            }
        }
        node = parent;
    }
    true
}
fn always_throws(node: P<Node>) -> bool {
    match node.kind() {
        Kind::ThrowStatement => true,
        Kind::Block => {
            for &statement in node.statements() {
                if always_throws(statement) {
                    return true;
                }
                if matches!(
                    statement.kind(),
                    Kind::ReturnStatement | Kind::BreakStatement | Kind::ContinueStatement
                ) {
                    return false;
                }
            }
            false
        }
        Kind::IfStatement => {
            let statement = node.as_if_statement();
            statement.else_statement().is_some_and(|other| {
                always_throws(statement.then_statement()) && always_throws(other)
            })
        }
        Kind::TryStatement => {
            let statement = node.as_try_statement();
            statement.finally_block().is_some_and(always_throws)
                || always_throws(statement.try_block())
                    && statement
                        .catch_clause()
                        .is_none_or(|catch| always_throws(catch.as_catch_clause().block()))
        }
        _ => false,
    }
}
fn may_suppress(node: P<Node>) -> bool {
    if matches!(
        node.kind(),
        Kind::ReturnStatement | Kind::BreakStatement | Kind::ContinueStatement
    ) {
        return true;
    }
    !is_function(node) && node.for_each_child(&mut may_suppress)
}

fn rstest_assert(node: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> bool {
    if node.kind() != Kind::CallExpression {
        return false;
    }
    let mut entries = Vec::new();
    calls::expect_entries(node, &mut entries);
    let Some(first) = entries.first() else {
        return false;
    };
    if entries.len() >= 2
        && first.name == "rstest"
        && entries[1].name == "assert"
        && first_identifier(node.expression().unwrap()).is_none()
        && contains_import_meta(node.expression().unwrap())
    {
        return true;
    }
    if first.node.kind() != Kind::Identifier {
        return false;
    }
    let symbol = checker.get_symbol_at_location_exported(first.node);
    let declarations = symbol
        .map(|symbol| symbol.declarations())
        .unwrap_or_default();
    // The frozen assert adapter deliberately does not reject later writes.
    for &declaration in declarations {
        if declaration.kind() == Kind::BindingElement {
            let mut parent = declaration.parent();
            while let Some(current) = parent {
                if current.kind() == Kind::VariableDeclaration {
                    if current.initializer().is_some_and(calls::import_meta) {
                        let element = declaration.as_binding_element();
                        return element
                            .property_name()
                            .map(static_name)
                            .unwrap_or_else(|| {
                                element
                                    .name()
                                    .filter(|name| name.kind() == Kind::Identifier)
                                    .map(|name| name.text().into())
                            })
                            .as_deref()
                            == Some("assert");
                    }
                    break;
                }
                parent = current.parent();
            }
        }
        if declaration.kind() == Kind::VariableDeclaration
            && declaration.initializer().is_some_and(calls::import_meta)
        {
            return entries.get(1).is_some_and(|entry| entry.name == "assert");
        }
    }
    let mut local = false;
    for &declaration in declarations {
        if ast::is_type_only_import_declaration(declaration) {
            continue;
        }
        local |= ast::get_source_file_of_node(declaration) == Some(file);
        if let Some((name, playwright)) = calls::imported(declaration)
            && !playwright
        {
            return name == "assert"
                || name.is_empty() && entries.get(1).is_some_and(|entry| entry.name == "assert");
        }
    }
    !local && first.name == "assert"
}
fn contains_import_meta(node: P<Node>) -> bool {
    let node = ast::skip_parentheses(node);
    match node.kind() {
        Kind::MetaProperty => node.as_meta_property().name().text() == "meta",
        Kind::CallExpression | Kind::PropertyAccessExpression | Kind::ElementAccessExpression => {
            node.expression().is_some_and(contains_import_meta)
        }
        _ => false,
    }
}

// Static destructuring maps a chain value to the identifier that receives it.
// Spread, dynamic keys, duplicate object keys and rest bindings are ambiguous.
enum Step {
    Index(usize),
    Key(String),
}
fn static_binding(value: P<Node>) -> Option<P<Node>> {
    let mut current = ast::skip_parentheses(value);
    let mut steps = Vec::new();
    while let Some(parent) = current.parent() {
        match parent.kind() {
            Kind::ParenthesizedExpression => current = parent,
            Kind::ArrayLiteralExpression => {
                let elements = parent.as_array_literal_expression().elements.nodes();
                if elements
                    .iter()
                    .any(|element| element.kind() == Kind::SpreadElement)
                {
                    return None;
                }
                let index = elements
                    .iter()
                    .position(|&element| ast::skip_parentheses(element) == current)?;
                steps.push(Step::Index(index));
                current = parent;
            }
            Kind::PropertyAssignment => {
                if parent.initializer().map(ast::skip_parentheses) != Some(current) {
                    return None;
                }
                let object = parent
                    .parent()
                    .filter(|object| object.kind() == Kind::ObjectLiteralExpression)?;
                let key = static_name(parent.name()?)?;
                let mut count = 0;
                for &property in object.as_object_literal_expression().properties.nodes() {
                    match property.kind() {
                        Kind::PropertyAssignment => {
                            let name = static_name(property.name()?)?;
                            if name == key {
                                count += 1;
                                if property != parent {
                                    return None;
                                }
                            }
                        }
                        Kind::ShorthandPropertyAssignment => {
                            if property.name()?.text() == key {
                                return None;
                            }
                        }
                        _ => return None,
                    }
                }
                if count != 1 {
                    return None;
                }
                steps.push(Step::Key(key));
                current = object;
            }
            Kind::VariableDeclaration => {
                if parent.initializer().map(ast::skip_parentheses) != Some(current)
                    || steps.is_empty()
                {
                    return None;
                }
                return apply_steps(parent.name()?, &steps);
            }
            _ if ast::is_assignment_expression(parent, true) => {
                let binary = parent.as_binary_expression();
                if binary.operator_token().kind() != Kind::EqualsToken
                    || ast::skip_parentheses(binary.right()) != current
                    || steps.is_empty()
                {
                    return None;
                }
                return apply_steps(binary.left(), &steps);
            }
            _ => return None,
        }
    }
    None
}
fn apply_steps(mut target: P<Node>, steps: &[Step]) -> Option<P<Node>> {
    for step in steps.iter().rev() {
        target = match step {
            Step::Index(index) => array_target(target, *index)?,
            Step::Key(key) => object_target(target, key)?,
        };
    }
    target = unwrap_default(target);
    (target.kind() == Kind::Identifier).then_some(target)
}
fn array_target(pattern: P<Node>, index: usize) -> Option<P<Node>> {
    let pattern = ast::skip_parentheses(pattern);
    match pattern.kind() {
        Kind::ArrayBindingPattern => {
            let element = *pattern.as_binding_pattern().elements.nodes().get(index)?;
            if element.kind() != Kind::BindingElement
                || element.as_binding_element().dot_dot_dot_token().is_some()
            {
                return None;
            }
            element.name()
        }
        Kind::ArrayLiteralExpression => {
            let element = ast::skip_parentheses(
                *pattern
                    .as_array_literal_expression()
                    .elements
                    .nodes()
                    .get(index)?,
            );
            (!matches!(
                element.kind(),
                Kind::OmittedExpression | Kind::SpreadElement
            ))
            .then(|| unwrap_default(element))
        }
        _ => None,
    }
}
fn object_target(pattern: P<Node>, key: &str) -> Option<P<Node>> {
    let pattern = ast::skip_parentheses(pattern);
    match pattern.kind() {
        Kind::ObjectBindingPattern => {
            for &element in pattern.as_binding_pattern().elements.nodes() {
                if element.kind() != Kind::BindingElement {
                    continue;
                }
                let element = element.as_binding_element();
                if element.dot_dot_dot_token().is_some() {
                    continue;
                }
                if element
                    .property_name()
                    .or_else(|| element.name())
                    .and_then(static_name)
                    .as_deref()
                    == Some(key)
                {
                    return element.name();
                }
            }
        }
        Kind::ObjectLiteralExpression => {
            for &property in pattern.as_object_literal_expression().properties.nodes() {
                match property.kind() {
                    Kind::PropertyAssignment => {
                        if property.name().and_then(static_name).as_deref() == Some(key) {
                            return property.initializer().map(unwrap_default);
                        }
                    }
                    Kind::ShorthandPropertyAssignment => {
                        if property.name().is_some_and(|name| name.text() == key) {
                            return property.name();
                        }
                    }
                    Kind::SpreadAssignment => return None,
                    _ => {}
                }
            }
        }
        _ => {}
    }
    None
}
fn unwrap_default(node: P<Node>) -> P<Node> {
    let node = ast::skip_parentheses(node);
    if node.kind() == Kind::BinaryExpression
        && node.as_binary_expression().operator_token().kind() == Kind::EqualsToken
    {
        ast::skip_parentheses(node.as_binary_expression().left())
    } else {
        node
    }
}
fn literal_name(node: P<Node>) -> Option<String> {
    let node = ast::skip_parentheses(node);
    matches!(
        node.kind(),
        Kind::StringLiteral | Kind::NoSubstitutionTemplateLiteral
    )
    .then(|| node.text().to_string())
}
fn static_name(node: P<Node>) -> Option<String> {
    let mut node = ast::skip_parentheses(node);
    let computed = node.kind() == Kind::ComputedPropertyName;
    if computed {
        node = ast::skip_parentheses(node.as_computed_property_name().expression());
    }
    if let Some(value) = literal_name(node) {
        return Some(value);
    }
    if node.kind() == Kind::NumericLiteral {
        return Some(jsnum::from_string(node.text()).string());
    }
    (!computed && node.kind() == Kind::Identifier).then(|| node.text().to_string())
}
