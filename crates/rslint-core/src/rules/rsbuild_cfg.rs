//! Shared Rust control-flow graph, ported from frozen rslint
//! e0a29f751d4def86d27931759098f27748fa3003 `internal/utils/cfg`.
//! Nodes are borrowed from the active tsrs Program. All graph storage is owned
//! and drops before the Program; no AST allocation is leaked or copied.
use std::collections::HashSet;
use tsrs_ast::{self as ast, Kind, Node};
use tsrs_core::{P, jsnum};

#[derive(Clone, Copy, Debug)]
pub(crate) enum PromiseEvent {
    Expression(P<Node>),
    Read(P<Node>),
    Write(P<Node>),
    Statement(P<Node>),
}
#[derive(Default)]
pub(crate) struct PromiseBlock {
    pub(crate) reachable: bool,
    pub(crate) successors: Vec<usize>,
    pub(crate) events: Vec<PromiseEvent>,
    incoming: bool,
}
pub(crate) struct PromiseGraph {
    pub(crate) blocks: Vec<PromiseBlock>,
}

pub(crate) fn promise_flow(root: P<Node>) -> PromiseGraph {
    let mut builder = Cfg::new();
    builder.record_events = true;
    builder.build_root(root);
    PromiseGraph {
        blocks: builder.blocks,
    }
}
fn children(node: P<Node>) -> Vec<P<Node>> {
    let mut result = Vec::new();
    node.for_each_child(&mut |child| {
        result.push(child);
        false
    });
    result
}
pub(crate) fn is_root(node: P<Node>) -> bool {
    ast::is_function_like_declaration(node)
        || matches!(
            node.kind(),
            Kind::SourceFile | Kind::ClassStaticBlockDeclaration
        )
        || node.kind() == Kind::PropertyDeclaration && node.initializer().is_some()
}
// Whether a syntax reference executes inside a root, rather than in the
// surrounding declaration (decorators and computed names execute outside).
pub(crate) fn root_of(node: P<Node>) -> Option<P<Node>> {
    let mut previous = node;
    let mut current = node.parent();
    let mut decorator = None;
    while let Some(parent) = current {
        if parent.kind() == Kind::Decorator {
            decorator = Some(parent);
        }
        if parent.kind() == Kind::SourceFile {
            return Some(parent);
        }
        if is_root(parent) {
            let inside = match parent.kind() {
                Kind::PropertyDeclaration => parent.initializer() == Some(previous),
                Kind::ClassStaticBlockDeclaration => {
                    parent.as_class_static_block_declaration().body() == previous
                }
                _ => {
                    parent.name() != Some(previous)
                        && !decorator.is_some_and(|decorator: P<Node>| {
                            decorator.parent() == Some(parent)
                                || decorator.parent() == Some(previous)
                        })
                }
            };
            if inside {
                return Some(parent);
            }
        }
        previous = parent;
        current = parent.parent();
    }
    None
}
fn clauses(node: P<Node>) -> Vec<P<Node>> {
    node.as_switch_statement()
        .case_block()
        .as_case_block()
        .clauses()
        .nodes()
        .to_vec()
}

struct Jump {
    labels: Vec<String>,
    break_to: usize,
    continue_to: Option<usize>,
    breakable: bool,
}
#[derive(Clone)]
struct TryFrame {
    in_catch: bool,
    catch: Option<usize>,
    finally: Option<usize>,
    forked: bool,
    returned: bool,
    thrown: bool,
    implicit: Vec<usize>,
}
pub(crate) struct Cfg {
    blocks: Vec<PromiseBlock>,
    current: usize,
    jumps: Vec<Jump>,
    tries: Vec<TryFrame>,
    pub(crate) reached: HashSet<P<Node>>,
    record_events: bool,
    chain_joins: Vec<usize>,
}
impl Cfg {
    pub(crate) fn new() -> Self {
        Self {
            blocks: vec![PromiseBlock {
                incoming: true,
                reachable: true,
                successors: Vec::new(),
                events: Vec::new(),
            }],
            current: 0,
            jumps: Vec::new(),
            tries: Vec::new(),
            reached: HashSet::new(),
            record_events: false,
            chain_joins: Vec::new(),
        }
    }
    fn block(&mut self) -> usize {
        let index = self.blocks.len();
        self.blocks.push(PromiseBlock::default());
        index
    }
    pub(crate) fn live(&self) -> bool {
        self.blocks[self.current].reachable
    }
    fn link(&mut self, from: usize, to: Option<usize>) {
        if let Some(to) = to {
            if self.record_events {
                self.blocks[from].successors.push(to);
            }
            self.blocks[to].incoming |= self.blocks[from].reachable;
        }
    }
    fn edge(&mut self, to: usize) {
        self.link(self.current, Some(to));
    }
    fn enter(&mut self, block: usize) {
        self.blocks[block].reachable = self.blocks[block].incoming;
        self.current = block;
    }
    fn dead(&mut self) {
        let next = self.block();
        if self.record_events {
            self.blocks[self.current].successors.push(next);
        }
        self.current = next;
    }
    fn throw_frame(&self) -> Option<usize> {
        self.tries
            .iter()
            .rposition(|frame| !frame.in_catch || frame.finally.is_some())
    }
    fn return_frame(&self) -> Option<usize> {
        self.tries.iter().rposition(|frame| frame.finally.is_some())
    }
    fn throw_target(frame: &TryFrame) -> Option<usize> {
        if !frame.in_catch && frame.catch.is_some() {
            frame.catch
        } else {
            frame.finally
        }
    }
    fn throwable(&mut self) {
        if !self.live() {
            return;
        }
        let Some(index) = self.throw_frame() else {
            return;
        };
        if self.tries[index].forked {
            return;
        }
        self.tries[index].forked = true;
        if self.tries[index].in_catch && self.tries[index].finally.is_some() {
            self.tries[index].implicit.push(self.current);
        } else {
            self.tries[index].thrown = true;
            self.link(self.current, Self::throw_target(&self.tries[index]));
        }
        let next = self.block();
        self.edge(next);
        self.enter(next);
    }
    fn abrupt_return(&mut self) {
        if self.live() {
            if let Some(index) = self.return_frame() {
                self.tries[index].returned = true;
                self.link(self.current, self.tries[index].finally);
            }
            self.dead();
        }
    }
    fn abrupt_throw(&mut self) {
        if self.live() {
            if let Some(index) = self.throw_frame() {
                self.tries[index].thrown = true;
                self.link(self.current, Self::throw_target(&self.tries[index]));
            }
            self.dead();
        }
    }
    pub(crate) fn sequence(&mut self, statements: &[P<Node>]) {
        for &statement in statements {
            self.statement(statement);
        }
    }
    fn condition(&mut self, node: Option<P<Node>>, yes: Option<usize>, no: Option<usize>) {
        let Some(node) = node else {
            self.link(self.current, yes);
            self.link(self.current, no);
            return;
        };
        match node.kind() {
            Kind::ParenthesizedExpression => {
                self.emit(PromiseEvent::Expression(node));
                self.condition(node.expression(), yes, no);
            }
            Kind::PrefixUnaryExpression
                if node.as_prefix_unary_expression().operator() == Kind::ExclamationToken =>
            {
                self.emit(PromiseEvent::Expression(node));
                self.condition(Some(node.as_prefix_unary_expression().operand()), no, yes)
            }
            Kind::BinaryExpression
                if matches!(
                    node.as_binary_expression().operator_token().kind(),
                    Kind::AmpersandAmpersandToken | Kind::BarBarToken | Kind::QuestionQuestionToken
                ) =>
            {
                self.emit(PromiseEvent::Expression(node));
                let binary = node.as_binary_expression();
                let right = self.block();
                match binary.operator_token().kind() {
                    Kind::AmpersandAmpersandToken => {
                        self.condition(Some(binary.left()), Some(right), no)
                    }
                    Kind::BarBarToken => self.condition(Some(binary.left()), yes, Some(right)),
                    _ => {
                        self.expr(binary.left());
                        self.link(self.current, yes);
                        self.link(self.current, no);
                        self.edge(right);
                    }
                }
                self.enter(right);
                self.condition(Some(binary.right()), yes, no);
            }
            Kind::ConditionalExpression => {
                self.emit(PromiseEvent::Expression(node));
                let conditional = node.as_conditional_expression();
                let left = self.block();
                let right = self.block();
                self.condition(Some(conditional.condition()), Some(left), Some(right));
                self.enter(left);
                self.condition(Some(conditional.when_true()), yes, no);
                self.enter(right);
                self.condition(Some(conditional.when_false()), yes, no);
            }
            _ => {
                self.expr(node);
                self.link(self.current, yes);
                self.link(self.current, no);
            }
        }
    }
    fn labels(node: P<Node>) -> Vec<String> {
        let mut labels = Vec::new();
        let mut current = node;
        while let Some(parent) = current
            .parent()
            .filter(|n| n.kind() == Kind::LabeledStatement)
        {
            let labeled = parent.as_labeled_statement();
            if labeled.statement() != current {
                break;
            }
            labels.push(labeled.label().text().to_owned());
            current = parent;
        }
        labels
    }
    fn jump(&mut self, node: P<Node>, after: usize, next: Option<usize>) {
        self.jumps.push(Jump {
            labels: Self::labels(node),
            break_to: after,
            continue_to: next,
            breakable: true,
        });
    }
    pub(crate) fn statement(&mut self, node: P<Node>) {
        self.emit(PromiseEvent::Statement(node));
        if self.live() {
            self.reached.insert(node);
        }
        match node.kind() {
            Kind::SourceFile | Kind::Block | Kind::ModuleBlock => self.sequence(node.statements()),
            Kind::VariableStatement => {
                self.variable_list(node.as_variable_statement().declaration_list())
            }
            Kind::ExpressionStatement | Kind::ExportAssignment => {
                if let Some(expression) = node.expression() {
                    self.expr(expression);
                }
            }
            Kind::IfStatement => {
                let stmt = node.as_if_statement();
                let after = self.block();
                let yes = self.block();
                let no = if stmt.else_statement().is_some() {
                    self.block()
                } else {
                    after
                };
                self.condition(Some(stmt.expression()), Some(yes), Some(no));
                self.enter(yes);
                self.statement(stmt.then_statement());
                self.edge(after);
                if let Some(other) = stmt.else_statement() {
                    self.enter(no);
                    self.statement(other);
                    self.edge(after);
                }
                self.enter(after);
            }
            Kind::WhileStatement => {
                let stmt = node.as_while_statement();
                let test = self.block();
                self.edge(test);
                self.enter(test);
                let after = self.block();
                let body = self.block();
                self.condition(
                    Some(stmt.expression()),
                    Some(body),
                    (!always_truthy(stmt.expression())).then_some(after),
                );
                self.jump(node, after, Some(test));
                self.enter(body);
                self.statement(stmt.statement());
                self.edge(test);
                self.jumps.pop();
                self.enter(after);
            }
            Kind::DoStatement => {
                let stmt = node.as_do_statement();
                let body = self.block();
                self.edge(body);
                let test = self.block();
                let after = self.block();
                self.jump(node, after, Some(test));
                self.enter(body);
                self.statement(stmt.statement());
                self.edge(test);
                self.jumps.pop();
                self.enter(test);
                self.expr(stmt.expression());
                self.edge(body);
                if !always_truthy(stmt.expression()) {
                    self.edge(after);
                }
                self.enter(after);
            }
            Kind::ForStatement => {
                let stmt = node.as_for_statement();
                if let Some(init) = stmt.initializer() {
                    if init.kind() == Kind::VariableDeclarationList {
                        self.variable_list(init);
                    } else {
                        self.expr(init);
                    }
                }
                if stmt.condition().is_none() && stmt.incrementor().is_none() {
                    let body = self.block();
                    self.edge(body);
                    let after = self.block();
                    self.jump(node, after, Some(body));
                    self.enter(body);
                    self.statement(stmt.statement());
                    self.edge(body);
                    self.jumps.pop();
                    self.enter(after);
                    return;
                }
                let test = self.block();
                self.edge(test);
                self.enter(test);
                let after = self.block();
                let body = self.block();
                self.condition(
                    stmt.condition(),
                    Some(body),
                    stmt.condition()
                        .filter(|&n| !always_truthy(n))
                        .map(|_| after),
                );
                let update = if let Some(increment) = stmt.incrementor() {
                    let update = self.block();
                    self.blocks[update].incoming |= self.live();
                    self.enter(update);
                    self.expr(increment);
                    self.edge(test);
                    update
                } else {
                    test
                };
                self.jump(node, after, Some(update));
                self.enter(body);
                self.statement(stmt.statement());
                self.edge(update);
                self.jumps.pop();
                self.enter(after);
            }
            Kind::ForInStatement | Kind::ForOfStatement => {
                let stmt = node.as_for_in_or_of_statement();
                self.expr(stmt.expression());
                let head = self.block();
                self.edge(head);
                self.enter(head);
                let after = self.block();
                self.edge(after);
                let body = self.block();
                self.edge(body);
                self.jump(node, after, Some(head));
                self.enter(body);
                let init = stmt.initializer();
                if init.kind() == Kind::VariableDeclarationList {
                    for &declaration in init.as_variable_declaration_list().declarations().nodes() {
                        if let Some(name) = declaration.name() {
                            self.pattern_bind(name);
                        }
                    }
                } else {
                    self.pattern_bind(init);
                }
                self.statement(stmt.statement());
                self.edge(head);
                self.jumps.pop();
                self.enter(after);
            }
            Kind::SwitchStatement => {
                self.expr(node.as_switch_statement().expression());
                let after = self.block();
                let clauses = clauses(node);
                let bodies: Vec<_> = clauses.iter().map(|_| self.block()).collect();
                let mut default = None;
                for (i, &clause) in clauses.iter().enumerate() {
                    if clause.kind() == Kind::DefaultClause {
                        default = Some(i);
                    } else {
                        if let Some(expression) = clause.expression() {
                            self.expr(expression);
                        }
                        self.edge(bodies[i]);
                        let next = self.block();
                        self.edge(next);
                        self.enter(next);
                    }
                }
                self.edge(default.map_or(after, |index| bodies[index]));
                self.jump(node, after, None);
                let mut fallthrough = None;
                for (i, &clause) in clauses.iter().enumerate() {
                    if let Some(previous) = fallthrough {
                        self.link(previous, Some(bodies[i]));
                    }
                    self.enter(bodies[i]);
                    self.sequence(clause.statements());
                    fallthrough = Some(self.current);
                }
                if let Some(previous) = fallthrough {
                    self.link(previous, Some(after));
                }
                self.jumps.pop();
                self.enter(after);
            }
            Kind::TryStatement => self.try_statement(node),
            Kind::LabeledStatement => {
                let stmt = node.as_labeled_statement();
                if breakable(stmt.statement()) {
                    self.statement(stmt.statement());
                } else {
                    let after = self.block();
                    self.jumps.push(Jump {
                        labels: vec![stmt.label().text().to_owned()],
                        break_to: after,
                        continue_to: None,
                        breakable: false,
                    });
                    self.statement(stmt.statement());
                    self.jumps.pop();
                    self.edge(after);
                    self.enter(after);
                }
            }
            Kind::ReturnStatement => {
                if let Some(expression) = node.expression() {
                    self.expr(expression);
                }
                self.abrupt_return();
            }
            Kind::ThrowStatement => {
                if let Some(expression) = node.expression() {
                    self.expr(expression);
                }
                self.abrupt_throw();
            }
            Kind::BreakStatement | Kind::ContinueStatement => {
                if self.live() {
                    let label = if node.kind() == Kind::BreakStatement {
                        node.as_break_statement().label()
                    } else {
                        node.as_continue_statement().label()
                    };
                    let is_continue = node.kind() == Kind::ContinueStatement;
                    let target = self
                        .jumps
                        .iter()
                        .rev()
                        .find(|jump| {
                            if let Some(label) = label {
                                jump.labels.iter().any(|name| name == label.text())
                                    && (!is_continue || jump.continue_to.is_some())
                            } else if is_continue {
                                jump.continue_to.is_some()
                            } else {
                                jump.breakable
                            }
                        })
                        .and_then(|jump| {
                            if is_continue {
                                jump.continue_to
                            } else {
                                Some(jump.break_to)
                            }
                        });
                    self.link(self.current, target);
                    self.dead();
                }
            }
            Kind::WithStatement => {
                let stmt = node.as_with_statement();
                self.expr(stmt.expression());
                self.statement(stmt.statement());
            }
            Kind::EmptyStatement
            | Kind::DebuggerStatement
            | Kind::ImportDeclaration
            | Kind::ImportEqualsDeclaration
            | Kind::FunctionDeclaration
            | Kind::ExportDeclaration => {}
            Kind::ClassDeclaration => self.class_like(node),
            _ => {
                for child in children(node) {
                    self.unknown(child);
                }
            }
        }
    }
    fn try_statement(&mut self, node: P<Node>) {
        let stmt = node.as_try_statement();
        let catch = stmt.catch_clause().map(|_| self.block());
        let finally = stmt.finally_block().map(|_| self.block());
        self.tries.push(TryFrame {
            in_catch: false,
            catch,
            finally,
            forked: false,
            returned: false,
            thrown: false,
            implicit: Vec::new(),
        });
        self.statement(stmt.try_block());
        let mut normal = Vec::new();
        if let (Some(catch), Some(clause)) = (catch, stmt.catch_clause()) {
            self.edge(catch);
            normal.push(self.current);
            let frame = self.tries.last_mut().unwrap();
            frame.in_catch = true;
            frame.forked = false;
            self.enter(catch);
            let clause = clause.as_catch_clause();
            if let Some(variable) = clause.variable_declaration() {
                if let Some(name) = variable.name() {
                    self.pattern_bind(name);
                }
            }
            self.statement(clause.block());
        }
        normal.push(self.current);
        let frame = self.tries.pop().unwrap();
        let after = self.block();
        let Some(finally_node) = stmt.finally_block() else {
            for previous in normal {
                self.link(previous, Some(after));
            }
            self.enter(after);
            return;
        };
        let snapshot: Vec<_> = self.tries.iter().map(|frame| frame.forked).collect();
        let normal_entry = self.block();
        for previous in normal.iter().chain(&frame.implicit) {
            self.link(*previous, Some(normal_entry));
        }
        self.enter(normal_entry);
        self.statement(finally_node);
        self.edge(after);
        if let Some(finally) = frame.finally.filter(|&index| self.blocks[index].incoming) {
            for (outer, forked) in self.tries.iter_mut().zip(snapshot) {
                outer.forked = forked;
            }
            self.enter(finally);
            self.statement(finally_node);
            if self.live() {
                if frame.returned
                    && let Some(index) = self.return_frame()
                {
                    self.tries[index].returned = true;
                    self.link(self.current, self.tries[index].finally);
                }
                if frame.thrown
                    && let Some(index) = self.throw_frame()
                {
                    self.tries[index].thrown = true;
                    self.link(self.current, Self::throw_target(&self.tries[index]));
                }
            }
        }
        self.enter(after);
    }
    fn emit(&mut self, event: PromiseEvent) {
        if self.record_events {
            self.blocks[self.current].events.push(event);
        }
    }
    fn build_root(&mut self, node: P<Node>) {
        match node.kind() {
            Kind::SourceFile => self.sequence(node.statements()),
            Kind::ClassStaticBlockDeclaration => {
                self.statement(node.as_class_static_block_declaration().body());
            }
            Kind::PropertyDeclaration => {
                if let Some(initializer) = node.initializer() {
                    self.expr(initializer);
                }
            }
            _ => {
                for &parameter in node.parameters() {
                    if let Some(ty) = parameter.type_node() {
                        self.expr(ty);
                    }
                    if let Some(name) = parameter.name() {
                        self.bind_default(name, parameter.initializer());
                    }
                }
                if let Some(ty) = node.type_node() {
                    self.expr(ty);
                }
                if let Some(body) = node.body() {
                    if body.kind() == Kind::Block {
                        self.statement(body);
                    } else {
                        self.expr(body);
                    }
                }
            }
        }
    }
    fn unknown(&mut self, node: P<Node>) {
        if ast::is_statement(node) {
            self.statement(node);
        } else {
            self.expr(node);
        }
    }
    fn decorators(&mut self, node: P<Node>) {
        for decorator in node.decorators() {
            self.expr(decorator.as_decorator().expression());
        }
    }
    fn member_header(&mut self, node: P<Node>) {
        self.decorators(node);
        if ast::is_function_like_declaration(node) {
            for &parameter in node.parameters() {
                self.decorators(parameter);
            }
        }
        if let Some(name) = node
            .name()
            .filter(|node| node.kind() == Kind::ComputedPropertyName)
        {
            self.expr(name.as_computed_property_name().expression());
        }
        match node.kind() {
            Kind::PropertyDeclaration => {
                if let Some(ty) = node.type_node() {
                    self.expr(ty);
                }
            }
            Kind::IndexSignature => {
                for &parameter in node.parameters() {
                    if let Some(ty) = parameter.type_node() {
                        self.expr(ty);
                    }
                }
                if let Some(ty) = node.type_node() {
                    self.expr(ty);
                }
            }
            _ => {}
        }
    }
    fn class_like(&mut self, node: P<Node>) {
        self.decorators(node);
        for &parameter in node.type_parameters() {
            let parameter = parameter.as_type_parameter_declaration();
            if let Some(constraint) = parameter.constraint() {
                self.expr(constraint);
            }
            if let Some(default) = parameter.default_type() {
                self.expr(default);
            }
        }
        for child in children(node) {
            if child.kind() == Kind::HeritageClause {
                for &ty in child.as_heritage_clause().types().nodes() {
                    self.expr(ty);
                }
            }
        }
        for &member in node.members() {
            self.member_header(member);
        }
    }
    fn variable_list(&mut self, list: P<Node>) {
        for &declaration in list.as_variable_declaration_list().declarations().nodes() {
            if let Some(ty) = declaration.type_node() {
                self.expr(ty);
            }
            let name = declaration.as_variable_declaration().name();
            if matches!(
                name.kind(),
                Kind::ObjectBindingPattern | Kind::ArrayBindingPattern
            ) {
                if let Some(initializer) = declaration.initializer() {
                    self.expr(initializer);
                }
                self.pattern_bind(name);
            } else {
                self.pattern_reads(name);
                if let Some(initializer) = declaration.initializer() {
                    self.expr(initializer);
                    self.pattern_writes(name);
                }
            }
        }
    }
    fn pattern_reads(&mut self, node: P<Node>) {
        match node.kind() {
            Kind::Identifier => {
                self.emit(PromiseEvent::Read(node));
                if throwable_identifier(node) {
                    self.throwable();
                }
            }
            Kind::ParenthesizedExpression
            | Kind::NonNullExpression
            | Kind::AsExpression
            | Kind::SatisfiesExpression
            | Kind::TypeAssertionExpression => {
                if let Some(inner) = node.expression() {
                    self.pattern_reads(inner);
                }
            }
            _ => self.expr(node),
        }
    }
    fn pattern_writes(&mut self, node: P<Node>) {
        match node.kind() {
            Kind::Identifier => self.emit(PromiseEvent::Write(node)),
            Kind::ParenthesizedExpression
            | Kind::NonNullExpression
            | Kind::AsExpression
            | Kind::SatisfiesExpression
            | Kind::TypeAssertionExpression => {
                if let Some(inner) = node.expression() {
                    self.pattern_writes(inner);
                }
            }
            _ => {}
        }
    }
    fn pattern_bind(&mut self, node: P<Node>) {
        match node.kind() {
            Kind::Identifier => {
                self.emit(PromiseEvent::Read(node));
                if throwable_identifier(node) {
                    self.throwable();
                }
                self.emit(PromiseEvent::Write(node));
            }
            Kind::OmittedExpression => {}
            Kind::ParenthesizedExpression
            | Kind::NonNullExpression
            | Kind::AsExpression
            | Kind::SatisfiesExpression
            | Kind::TypeAssertionExpression => {
                if let Some(inner) = node.expression() {
                    self.pattern_bind(inner);
                }
            }
            Kind::ObjectBindingPattern | Kind::ArrayBindingPattern => {
                for &element in node.elements() {
                    if element.kind() != Kind::BindingElement {
                        continue;
                    }
                    let binding = element.as_binding_element();
                    if let Some(property) = binding
                        .property_name()
                        .filter(|property| property.kind() == Kind::ComputedPropertyName)
                    {
                        self.expr(property.as_computed_property_name().expression());
                    }
                    if let Some(name) = binding.name() {
                        self.bind_default(name, binding.initializer());
                    }
                }
            }
            Kind::ObjectLiteralExpression => {
                for &property in node.properties() {
                    match property.kind() {
                        Kind::PropertyAssignment => {
                            if let Some(name) = property
                                .name()
                                .filter(|name| name.kind() == Kind::ComputedPropertyName)
                            {
                                self.expr(name.as_computed_property_name().expression());
                            }
                            self.pattern_bind(property.as_property_assignment().initializer());
                        }
                        Kind::ShorthandPropertyAssignment => {
                            let shorthand = property.as_shorthand_property_assignment();
                            self.bind_default(
                                shorthand.name(),
                                shorthand.object_assignment_initializer(),
                            );
                        }
                        Kind::SpreadAssignment => {
                            self.pattern_bind(property.as_spread_assignment().expression())
                        }
                        _ => {}
                    }
                }
            }
            Kind::ArrayLiteralExpression => {
                for &element in node.elements() {
                    self.pattern_bind(element);
                }
            }
            Kind::SpreadElement => self.pattern_bind(node.as_spread_element().expression()),
            Kind::BinaryExpression
                if node.as_binary_expression().operator_token().kind() == Kind::EqualsToken =>
            {
                let binary = node.as_binary_expression();
                self.bind_default(binary.left(), Some(binary.right()));
            }
            _ => self.expr(node),
        }
    }
    fn bind_default(&mut self, target: P<Node>, fallback: Option<P<Node>>) {
        if let Some(fallback) = fallback {
            let join = self.block();
            self.edge(join);
            let branch = self.block();
            self.edge(branch);
            self.enter(branch);
            self.expr(fallback);
            self.edge(join);
            self.enter(join);
        }
        self.pattern_bind(target);
    }
    fn binary_expression(&mut self, node: P<Node>) {
        let binary = node.as_binary_expression();
        let op = binary.operator_token().kind();
        if matches!(
            op,
            Kind::AmpersandAmpersandToken | Kind::BarBarToken | Kind::QuestionQuestionToken
        ) {
            self.expr(binary.left());
            let join = self.block();
            self.edge(join);
            let right = self.block();
            self.edge(right);
            self.enter(right);
            self.expr(binary.right());
            self.edge(join);
            self.enter(join);
        } else if matches!(
            op,
            Kind::AmpersandAmpersandEqualsToken
                | Kind::BarBarEqualsToken
                | Kind::QuestionQuestionEqualsToken
        ) {
            self.pattern_reads(binary.left());
            let join = self.block();
            self.edge(join);
            let right = self.block();
            self.edge(right);
            self.enter(right);
            self.expr(binary.right());
            self.pattern_writes(binary.left());
            self.edge(join);
            self.enter(join);
        } else if op == Kind::EqualsToken {
            if matches!(
                ast::skip_parentheses(binary.left()).kind(),
                Kind::ObjectLiteralExpression | Kind::ArrayLiteralExpression
            ) {
                self.expr(binary.right());
                self.pattern_bind(binary.left());
            } else {
                self.pattern_reads(binary.left());
                self.expr(binary.right());
                self.pattern_writes(binary.left());
            }
        } else if ast::is_assignment_operator(op) {
            self.pattern_reads(binary.left());
            self.expr(binary.right());
            self.pattern_writes(binary.left());
        } else {
            self.expr(binary.left());
            self.expr(binary.right());
        }
    }
    fn fork_optional(&mut self, node: P<Node>) {
        if ast::is_optional_chain_root(node)
            && let Some(&join) = self.chain_joins.last()
        {
            self.edge(join);
            let next = self.block();
            self.edge(next);
            self.enter(next);
        }
    }
    fn type_arguments(&mut self, node: P<Node>) {
        for &argument in node.type_arguments() {
            self.expr(argument);
        }
    }
    fn access_or_call(&mut self, node: P<Node>) {
        let outermost = ast::is_optional_chain(node) && ast::is_outermost_optional_chain(node);
        if outermost {
            let join = self.block();
            self.chain_joins.push(join);
        }
        match node.kind() {
            Kind::PropertyAccessExpression => {
                self.expr(node.as_property_access_expression().expression());
                self.fork_optional(node);
            }
            Kind::ElementAccessExpression => {
                let access = node.as_element_access_expression();
                self.expr(access.expression());
                self.fork_optional(node);
                self.expr(access.argument_expression());
            }
            Kind::CallExpression => {
                self.expr(node.as_call_expression().expression());
                self.fork_optional(node);
                self.type_arguments(node);
                for &argument in node.arguments() {
                    self.expr(argument);
                }
            }
            Kind::NewExpression => {
                self.expr(node.as_new_expression().expression());
                self.type_arguments(node);
                for &argument in node.arguments() {
                    self.expr(argument);
                }
            }
            Kind::TaggedTemplateExpression => {
                let tagged = node.as_tagged_template_expression();
                self.expr(tagged.tag());
                self.type_arguments(node);
                self.expr(tagged.template());
            }
            _ => unreachable!(),
        }
        self.throwable();
        if outermost {
            let join = self.chain_joins.pop().unwrap();
            self.edge(join);
            self.enter(join);
        }
    }
    fn expr(&mut self, node: P<Node>) {
        self.emit(PromiseEvent::Expression(node));
        match node.kind() {
            Kind::Identifier => {
                self.emit(PromiseEvent::Read(node));
                if throwable_identifier(node) {
                    self.throwable();
                }
            }
            Kind::ParenthesizedExpression => {
                self.expr(node.as_parenthesized_expression().expression())
            }
            Kind::BinaryExpression => self.binary_expression(node),
            Kind::ConditionalExpression => {
                let conditional = node.as_conditional_expression();
                self.expr(conditional.condition());
                let branch = self.current;
                let join = self.block();
                let yes = self.block();
                self.link(branch, Some(yes));
                self.enter(yes);
                self.expr(conditional.when_true());
                self.edge(join);
                let no = self.block();
                self.link(branch, Some(no));
                self.enter(no);
                self.expr(conditional.when_false());
                self.edge(join);
                self.enter(join);
            }
            Kind::PrefixUnaryExpression => {
                let unary = node.as_prefix_unary_expression();
                if matches!(
                    unary.operator(),
                    Kind::PlusPlusToken | Kind::MinusMinusToken
                ) {
                    self.pattern_reads(unary.operand());
                    self.pattern_writes(unary.operand());
                } else {
                    self.expr(unary.operand());
                }
            }
            Kind::PostfixUnaryExpression => {
                let operand = node.as_postfix_unary_expression().operand();
                self.pattern_reads(operand);
                self.pattern_writes(operand);
            }
            Kind::PropertyAccessExpression
            | Kind::ElementAccessExpression
            | Kind::CallExpression
            | Kind::NewExpression
            | Kind::TaggedTemplateExpression => self.access_or_call(node),
            Kind::YieldExpression => {
                if let Some(expression) = node.expression() {
                    self.expr(expression);
                }
                if self.live() {
                    if let Some(index) = self.return_frame() {
                        self.tries[index].returned = true;
                        self.link(self.current, self.tries[index].finally);
                    }
                    if let Some(index) = self.throw_frame() {
                        self.tries[index].thrown = true;
                        self.link(self.current, Self::throw_target(&self.tries[index]));
                    }
                    let next = self.block();
                    self.edge(next);
                    self.enter(next);
                }
            }
            Kind::ClassExpression => self.class_like(node),
            _ if is_root(node) => self.member_header(node),
            _ => {
                for child in children(node) {
                    self.unknown(child);
                }
            }
        }
    }
}
fn breakable(node: P<Node>) -> bool {
    matches!(
        node.kind(),
        Kind::SwitchStatement
            | Kind::WhileStatement
            | Kind::DoStatement
            | Kind::ForStatement
            | Kind::ForInStatement
            | Kind::ForOfStatement
    )
}
fn always_truthy(node: P<Node>) -> bool {
    let node = ast::skip_parentheses(node);
    match node.kind() {
        Kind::TrueKeyword | Kind::RegularExpressionLiteral => true,
        Kind::NumericLiteral => jsnum::from_string(node.text()).0 != 0.0,
        Kind::BigIntLiteral => jsnum::parse_pseudo_big_int(node.text()) != "0",
        Kind::StringLiteral => !node.text().is_empty(),
        _ => false,
    }
}
fn throwable_identifier(node: P<Node>) -> bool {
    let Some(parent) = node.parent() else {
        return false;
    };
    match parent.kind() {
        Kind::LabeledStatement
        | Kind::BreakStatement
        | Kind::ContinueStatement
        | Kind::ArrayBindingPattern
        | Kind::ImportSpecifier
        | Kind::ExportSpecifier
        | Kind::ImportClause
        | Kind::NamespaceImport
        | Kind::NamespaceExport
        | Kind::JsxAttribute
        | Kind::JsxNamespacedName
        | Kind::CatchClause => false,
        Kind::BindingElement => {
            let binding = parent.as_binding_element();
            binding.dot_dot_dot_token().is_none()
                && binding.property_name() != Some(node)
                && parent
                    .parent()
                    .is_some_and(|n| n.kind() == Kind::ObjectBindingPattern)
        }
        Kind::FunctionDeclaration
        | Kind::FunctionExpression
        | Kind::ArrowFunction
        | Kind::ClassDeclaration
        | Kind::ClassExpression
        | Kind::VariableDeclaration
        | Kind::PropertyAssignment
        | Kind::PropertyDeclaration
        | Kind::MethodDeclaration
        | Kind::GetAccessor
        | Kind::SetAccessor
        | Kind::EnumDeclaration
        | Kind::ModuleDeclaration
        | Kind::InterfaceDeclaration
        | Kind::TypeAliasDeclaration => parent.name() != Some(node),
        _ => !ast::is_jsx_tag_name(node),
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use tsrs_ast::{SourceFile, SourceFileParseOptions};
    use tsrs_core::{ScriptKind, arena::Region, tspath::Path};

    fn with_file(source: &str, check: impl FnOnce(P<SourceFile>)) {
        let region = Region::new(1 << 20);
        let _scope = region.enter();
        let name = "/rslint-cfg-tests/input.ts";
        let file = tsrs_parser::parse_source_file(
            SourceFileParseOptions {
                file_name: name.into(),
                path: Path::new(name),
                ..Default::default()
            },
            source,
            ScriptKind::TS,
        );
        check(file);
    }
    fn references(graph: &PromiseGraph) -> Vec<String> {
        graph
            .blocks
            .iter()
            .flat_map(|block| block.events.iter())
            .filter_map(|event| match event {
                PromiseEvent::Read(node) => Some(format!("read:{}", node.text())),
                PromiseEvent::Write(node) => Some(format!("write:{}", node.text())),
                _ => None,
            })
            .collect()
    }
    fn read_blocks(graph: &PromiseGraph, name: &str) -> Vec<usize> {
        graph
            .blocks
            .iter()
            .enumerate()
            .filter(|(_, block)| {
                block
                    .events
                    .iter()
                    .any(|event| matches!(event, PromiseEvent::Read(node) if node.text() == name))
            })
            .map(|(index, _)| index)
            .collect()
    }
    fn reachable_without(graph: &PromiseGraph, from: usize, to: usize, omitted: &[usize]) -> bool {
        let mut seen = HashSet::new();
        let mut pending = vec![from];
        while let Some(index) = pending.pop() {
            if omitted.contains(&index) || !graph.blocks[index].reachable || !seen.insert(index) {
                continue;
            }
            if index == to {
                return true;
            }
            pending.extend(&graph.blocks[index].successors);
        }
        false
    }
    #[test]
    fn assignment_reads_receiver_then_rhs_before_identifier_write() {
        with_file(
            "function f() { let target = source(); target = next; obj.key = target; }",
            |file| {
                let graph = promise_flow(file.as_node().statements()[0]);
                assert_eq!(
                    references(&graph),
                    [
                        "read:target",
                        "read:source",
                        "write:target",
                        "read:target",
                        "read:next",
                        "write:target",
                        "read:obj",
                        "read:target"
                    ]
                );
                assert_eq!(graph.blocks.len(), 1);
                assert!(graph.blocks[0].successors.is_empty());
            },
        );
    }
    #[test]
    fn destructuring_evaluates_rhs_key_default_before_target_write() {
        with_file(
            "function f() { const { [key()]: target = fallback() } = source(); consume(target); }",
            |file| {
                let graph = promise_flow(file.as_node().statements()[0]);
                assert_eq!(
                    references(&graph),
                    [
                        "read:source",
                        "read:key",
                        "read:target",
                        "write:target",
                        "read:consume",
                        "read:target",
                        "read:fallback"
                    ]
                );
                let fallback = read_blocks(&graph, "fallback");
                let consume = read_blocks(&graph, "consume")[0];
                assert_eq!(fallback.len(), 1);
                assert!(reachable_without(&graph, 0, consume, &fallback));
                assert!(reachable_without(&graph, fallback[0], consume, &[]));
            },
        );
    }
    #[test]
    fn optional_chain_skips_computed_key_and_call_arguments() {
        with_file(
            "function f() { receiver?.[key()](argument()); after(); }",
            |file| {
                let graph = promise_flow(file.as_node().statements()[0]);
                let key = read_blocks(&graph, "key");
                let argument = read_blocks(&graph, "argument");
                let after = read_blocks(&graph, "after")[0];
                assert_eq!(key.len(), 1);
                assert_eq!(argument.len(), 1);
                assert!(reachable_without(&graph, 0, after, &key));
                assert!(reachable_without(&graph, 0, after, &argument));
                assert!(reachable_without(&graph, key[0], after, &[]));
            },
        );
    }
    #[test]
    fn finally_is_emitted_for_normal_and_abrupt_paths() {
        with_file(
            "function f() { try { if (condition) return value; normal(); } finally { cleanup(); } after(); }",
            |file| {
                let graph = promise_flow(file.as_node().statements()[0]);
                let cleanup = read_blocks(&graph, "cleanup");
                let after = read_blocks(&graph, "after")[0];
                assert_eq!(cleanup.len(), 2);
                assert_eq!(
                    cleanup
                        .iter()
                        .filter(|&&index| reachable_without(&graph, index, after, &[]))
                        .count(),
                    1
                );
                assert!(cleanup.iter().all(|&index| graph.blocks[index].reachable));
            },
        );
    }
    #[test]
    fn loop_back_edges_and_unreachable_tail_preserve_events() {
        with_file(
            "function f() { for (;;) { repeat(); if (condition) break; } return; dead(); }",
            |file| {
                let graph = promise_flow(file.as_node().statements()[0]);
                let repeat = read_blocks(&graph, "repeat")[0];
                let dead = read_blocks(&graph, "dead")[0];
                assert!(graph.blocks[repeat].reachable);
                assert!(!graph.blocks[dead].reachable);
                assert!(
                    graph
                        .blocks
                        .iter()
                        .enumerate()
                        .any(|(index, block)| index != repeat
                            && block.reachable
                            && block.successors.contains(&repeat))
                );
                assert!(!reachable_without(&graph, 0, dead, &[]));
            },
        );
    }
    #[test]
    fn static_block_roots_include_body_events_and_exclude_siblings() {
        with_file(
            "class C { static { before(); throw error; after(); } field = sibling(); }",
            |file| {
                let class = file.as_node().statements()[0];
                let block = class.members()[0];
                let body = block.as_class_static_block_declaration().body();
                let before = body.statements()[0]
                    .as_expression_statement()
                    .expression()
                    .as_call_expression()
                    .expression();
                assert_eq!(root_of(before), Some(block));
                let graph = promise_flow(block);
                let before_block = read_blocks(&graph, "before");
                let after_block = read_blocks(&graph, "after");
                assert_eq!(before_block.len(), 1);
                assert_eq!(after_block.len(), 1);
                assert!(graph.blocks[before_block[0]].reachable);
                assert!(!graph.blocks[after_block[0]].reachable);
                assert!(read_blocks(&graph, "sibling").is_empty());
            },
        );
    }
    #[test]
    fn root_of_excludes_computed_keys_and_parameter_decorators() {
        with_file(
            "class C { field = source(); [key()](@decorator argument = fallback()) { body(); } }",
            |file| {
                let class = file.as_node().statements()[0];
                let field = class.members()[0];
                let method = class.members()[1];
                fn find(node: P<Node>, name: &str) -> Option<P<Node>> {
                    if node.kind() == Kind::Identifier && node.text() == name {
                        return Some(node);
                    }
                    children(node)
                        .into_iter()
                        .find_map(|child| find(child, name))
                }
                assert_eq!(root_of(find(class, "source").unwrap()), Some(field));
                assert_eq!(root_of(find(class, "key").unwrap()), Some(file.as_node()));
                assert_eq!(
                    root_of(find(class, "decorator").unwrap()),
                    Some(file.as_node())
                );
                assert_eq!(root_of(find(class, "fallback").unwrap()), Some(method));
                assert_eq!(root_of(find(class, "body").unwrap()), Some(method));
            },
        );
    }
}
