//! Binding checks ported from frozen rslint e0a29f751d4def86d27931759098f27748fa3003.
//! Effective nondefault globals/sourceType and auxiliary global directives are
//! rejected by the shared admission layer; these rules implement the frozen
//! latest-ECMAScript/module defaults and the filename-selected CommonJS defaults.
use super::{
    rsbuild_assignment_checks::resolve_reference, rsbuild_core_checks::write_reference,
    rsbuild_declaration_checks::access, rsbuild_ts_syntax::lexically_shadowed,
};
use crate::{Violation, edits::Edit};
use std::collections::{HashMap, HashSet};
use tsrs_ast::{
    self as ast, Kind, ModifierFlags, Node, NodeFlags, OuterExpressionKinds, SourceFile, Symbol,
    SymbolFlags,
};
use tsrs_checker::Checker;
use tsrs_core::P;
use tsrs_scanner::skip_trivia;

pub(crate) const NAMES: &[&str] = &[
    "no-var",
    "no-obj-calls",
    "prefer-const",
    "no-global-assign",
    "no-new-native-nonconstructor",
];
pub(crate) fn check(
    name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
) -> Vec<Violation> {
    match (name, node.kind()) {
        ("no-obj-calls", Kind::SourceFile) => ObjectCalls::new(file, checker).report(),
        ("no-new-native-nonconstructor", Kind::NewExpression) => {
            new_native(node, file, checker).into_iter().collect()
        }
        ("no-global-assign", Kind::Identifier) => {
            global_assign(node, file, checker).into_iter().collect()
        }
        ("prefer-const" | "no-var", Kind::SourceFile) => {
            let index = ReferenceIndex::new(file, checker);
            if name == "prefer-const" {
                prefer_const(file, &index)
            } else {
                no_var(file, &index)
            }
        }
        _ => Vec::new(),
    }
}
fn start(node: P<Node>, file: P<SourceFile>) -> usize {
    skip_trivia(file.text(), node.pos().expect_source("parsed node")) as usize
}
fn end(node: P<Node>) -> usize {
    node.end().expect_source("parsed node") as usize
}
fn pos(node: P<Node>) -> usize {
    node.pos().expect_source("parsed node") as usize
}
fn walk(node: P<Node>, callback: &mut impl FnMut(P<Node>)) {
    if node.flags().intersects(NodeFlags::Reparsed) {
        return;
    }
    callback(node);
    node.for_each_child(&mut |child| {
        walk(child, callback);
        false
    });
}
fn ancestor(mut node: P<Node>, predicate: impl Fn(P<Node>) -> bool) -> Option<P<Node>> {
    while let Some(parent) = node.parent() {
        if predicate(parent) {
            return Some(parent);
        }
        node = parent;
    }
    None
}
fn bindings(node: P<Node>, out: &mut Vec<P<Node>>) {
    match node.kind() {
        Kind::Identifier => out.push(node),
        Kind::ObjectBindingPattern | Kind::ArrayBindingPattern => {
            for &element in node.elements() {
                if element.kind() == Kind::BindingElement
                    && let Some(name) = element.name()
                {
                    bindings(name, out);
                }
            }
        }
        _ => {}
    }
}
fn binding_symbol(node: P<Node>) -> Option<P<Symbol>> {
    node.parent()
        .filter(|parent| parent.name() == Some(node))
        .and_then(|parent| parent.symbol())
}
fn is_reference(node: P<Node>) -> bool {
    let Some(parent) = node.parent() else {
        return false;
    };
    match parent.kind() {
        Kind::PropertyAccessExpression if parent.name() == Some(node) => return false,
        Kind::QualifiedName if parent.as_qualified_name().right() == node => return false,
        Kind::ImportAttribute if parent.name() == Some(node) => return false,
        Kind::MetaProperty
        | Kind::LabeledStatement
        | Kind::BreakStatement
        | Kind::ContinueStatement => return false,
        Kind::ExportSpecifier => {
            if ast::is_type_only_import_or_export_declaration(parent) {
                return false;
            }
            if ancestor(parent, |node| node.kind() == Kind::ExportDeclaration)
                .is_some_and(|export| export.as_export_declaration().module_specifier().is_some())
            {
                return false;
            }
            return parent
                .as_export_specifier()
                .property_name()
                .unwrap_or(parent.name().unwrap())
                == node;
        }
        Kind::BindingElement if parent.as_binding_element().property_name() == Some(node) => {
            return false;
        }
        Kind::ImportSpecifier if parent.as_import_specifier().property_name() == Some(node) => {
            return false;
        }
        _ => {}
    }
    parent.kind() == Kind::ShorthandPropertyAssignment || !ast::is_declaration_name(node)
}
// RefStore indexes type-only exports when the selected symbol has a type or
// namespace binding. The runtime-root classifier above intentionally excludes
// them; they are still reads for prefer-const's read-before-assignment check.
fn binding_reference(node: P<Node>) -> bool {
    if node
        .parent()
        .is_some_and(|parent| parent.kind() == Kind::ExportSpecifier)
    {
        let parent = node.parent().unwrap();
        return !ancestor(parent, |node| node.kind() == Kind::ExportDeclaration)
            .is_some_and(|export| export.as_export_declaration().module_specifier().is_some())
            && parent
                .as_export_specifier()
                .property_name()
                .unwrap_or(parent.name().unwrap())
                == node;
    }
    is_reference(node)
}
fn resolve_binding_reference(node: P<Node>, checker: &mut Checker) -> Option<P<Symbol>> {
    let parent = node.parent()?;
    if parent.kind() == Kind::ExportSpecifier
        && ast::is_type_only_import_or_export_declaration(parent)
    {
        return checker.resolve_name_exported(
            node.text(),
            node,
            SymbolFlags::Type | SymbolFlags::Namespace | SymbolFlags::Alias,
            false,
        );
    }
    resolve_reference(node, checker)
}
struct ReferenceIndex {
    nodes: Vec<P<Node>>,
    resolved: HashMap<P<Node>, P<Symbol>>,
    references: HashMap<P<Symbol>, Vec<P<Node>>>,
    canonical: HashMap<P<Node>, P<Symbol>>,
}
impl ReferenceIndex {
    fn new(file: P<SourceFile>, checker: &mut Checker) -> Self {
        let mut nodes = Vec::new();
        walk(file.as_node(), &mut |node| nodes.push(node));
        let mut canonical = HashMap::new();
        for &node in &nodes {
            if let Some(symbol) = node.symbol() {
                for &declaration in symbol.declarations() {
                    canonical.entry(declaration).or_insert(symbol);
                }
            }
        }
        let mut result = Self {
            nodes,
            resolved: HashMap::new(),
            references: HashMap::new(),
            canonical,
        };
        for &node in &result.nodes {
            if node.kind() != Kind::Identifier || !binding_reference(node) {
                continue;
            }
            if let Some(symbol) = resolve_binding_reference(node, checker) {
                let symbol = result.canonical_symbol(symbol);
                result.resolved.insert(node, symbol);
                result.references.entry(symbol).or_default().push(node);
            }
        }
        result
    }
    fn canonical_symbol(&self, symbol: P<Symbol>) -> P<Symbol> {
        symbol
            .declarations()
            .iter()
            .find_map(|declaration| self.canonical.get(declaration).copied())
            .unwrap_or(symbol)
    }
    fn symbol(&self, node: P<Node>) -> Option<P<Symbol>> {
        self.resolved
            .get(&node)
            .copied()
            .or_else(|| binding_symbol(node).map(|symbol| self.canonical_symbol(symbol)))
    }
    fn refs(&self, symbol: P<Symbol>) -> &[P<Node>] {
        self.references
            .get(&self.canonical_symbol(symbol))
            .map(Vec::as_slice)
            .unwrap_or(&[])
    }
    fn writes(&self, symbol: P<Symbol>) -> Vec<P<Node>> {
        self.refs(symbol)
            .iter()
            .copied()
            .filter(|&node| write_reference(node))
            .collect()
    }
}
fn containing_block(node: P<Node>) -> Option<P<Node>> {
    ancestor(node, |node| {
        matches!(
            node.kind(),
            Kind::Block
                | Kind::SourceFile
                | Kind::ModuleBlock
                | Kind::CaseClause
                | Kind::DefaultClause
        )
    })
}
fn loop_assignee(node: P<Node>) -> bool {
    node.parent()
        .is_some_and(|parent| matches!(parent.kind(), Kind::ForInStatement | Kind::ForOfStatement))
}
fn direct_child_of_block(mut node: P<Node>, block: P<Node>) -> bool {
    while let Some(parent) = node.parent() {
        if parent == block {
            return true;
        }
        if matches!(
            parent.kind(),
            Kind::Block
                | Kind::SourceFile
                | Kind::ModuleBlock
                | Kind::CaseClause
                | Kind::DefaultClause
                | Kind::IfStatement
                | Kind::WhileStatement
                | Kind::DoStatement
                | Kind::ForStatement
                | Kind::ForInStatement
                | Kind::ForOfStatement
                | Kind::WithStatement
                | Kind::LabeledStatement
                | Kind::SwitchStatement
        ) {
            return false;
        }
        node = parent;
    }
    false
}
fn default_pattern(node: P<Node>) -> bool {
    if node.kind() != Kind::BinaryExpression
        || node.as_binary_expression().operator_token().kind() != Kind::EqualsToken
    {
        return false;
    }
    let mut expression = node;
    while let Some(parent) = expression
        .parent()
        .filter(|parent| parent.kind() == Kind::ParenthesizedExpression)
    {
        expression = parent;
    }
    let Some(target) = ast::get_assignment_target(expression) else {
        return false;
    };
    ast::is_destructuring_assignment(target)
        || (matches!(target.kind(), Kind::ForInStatement | Kind::ForOfStatement)
            && matches!(
                ast::skip_parentheses(target.as_for_in_or_of_statement().initializer()).kind(),
                Kind::ArrayLiteralExpression | Kind::ObjectLiteralExpression
            ))
}
fn standalone_assignment(node: P<Node>) -> bool {
    let Some(mut target) = ast::get_assignment_target(node) else {
        return false;
    };
    if matches!(
        target.kind(),
        Kind::PrefixUnaryExpression
            | Kind::PostfixUnaryExpression
            | Kind::ForInStatement
            | Kind::ForOfStatement
    ) {
        return false;
    }
    if default_pattern(target) {
        let Some(outer) = ancestor(target, ast::is_destructuring_assignment) else {
            return false;
        };
        target = outer;
    }
    ancestor(target, |node| node.kind() != Kind::ParenthesizedExpression)
        .is_some_and(|parent| parent.kind() == Kind::ExpressionStatement)
}
fn nonreportable_target(node: P<Node>, block: P<Node>, index: &ReferenceIndex) -> bool {
    let outside = |node| {
        index
            .symbol(node)
            .and_then(|symbol| symbol.declarations().first().copied())
            .is_none_or(|decl| containing_block(decl) != Some(block))
    };
    if node.kind() == Kind::Identifier {
        return outside(node);
    }
    node.for_each_child(&mut |child| match child.kind() {
        Kind::PropertyAccessExpression | Kind::ElementAccessExpression => true,
        Kind::Identifier => outside(child),
        Kind::ShorthandPropertyAssignment => child.name().is_some_and(outside),
        Kind::PropertyAssignment => {
            nonreportable_target(child.as_property_assignment().initializer(), block, index)
        }
        Kind::ArrayLiteralExpression
        | Kind::ObjectLiteralExpression
        | Kind::SpreadElement
        | Kind::SpreadAssignment => nonreportable_target(child, block, index),
        Kind::BinaryExpression => {
            nonreportable_target(child.as_binary_expression().left(), block, index)
        }
        _ => false,
    })
}
fn write_in_same_block(write: P<Node>, declaration: P<Node>, index: &ReferenceIndex) -> bool {
    let Some(block) = containing_block(declaration) else {
        return false;
    };
    if !direct_child_of_block(write, block) || !standalone_assignment(write) {
        return false;
    }
    ancestor(write, ast::is_destructuring_assignment).is_none_or(|assignment| {
        !nonreportable_target(assignment.as_binary_expression().left(), block, index)
    })
}
fn prefer_const(file: P<SourceFile>, index: &ReferenceIndex) -> Vec<Violation> {
    let mut out = Vec::new();
    for &list in &index.nodes {
        if list.kind() != Kind::VariableDeclarationList
            || !list.flags().intersects(NodeFlags::Let)
            || list
                .parent()
                .is_some_and(|parent| parent.kind() == Kind::ForStatement)
        {
            continue;
        }
        let mut candidates = Vec::new();
        let mut total = 0;
        let mut all_initialized = true;
        for &declaration in list.as_variable_declaration_list().declarations().nodes() {
            let initialized = declaration.initializer().is_some() || loop_assignee(list);
            all_initialized &= initialized;
            let mut names = Vec::new();
            bindings(declaration.as_variable_declaration().name(), &mut names);
            total += names.len();
            for name in names {
                let Some(symbol) = index.symbol(name) else {
                    continue;
                };
                let writes = index.writes(symbol);
                let report = if initialized {
                    if !writes.is_empty() {
                        continue;
                    }
                    name
                } else {
                    if writes.len() != 1 || !write_in_same_block(writes[0], declaration, index) {
                        continue;
                    }
                    if index.refs(symbol).iter().any(|&reference| {
                        pos(reference) >= pos(declaration)
                            && pos(reference) < pos(writes[0])
                            && !write_reference(reference)
                    }) {
                        name
                    } else {
                        writes[0]
                    }
                };
                candidates.push((name, report));
            }
        }
        let fix = all_initialized && total > 0 && candidates.len() == total;
        for (name, report) in candidates {
            let violation = Violation::new(
                report,
                "useConst",
                format!(
                    "'{}' is never reassigned. Use 'const' instead.",
                    name.text()
                ),
            );
            out.push(if fix {
                let begin = start(list, file);
                violation.with_fix(vec![Edit::replace([begin, begin + 3], "const")])
            } else {
                violation
            });
        }
    }
    out
}
pub(crate) fn runtime_shadowed(node: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> bool {
    resolve_reference(node, checker).map_or_else(
        || lexically_shadowed(node, node.text()),
        |symbol| local_value(symbol, file),
    )
}
fn new_native(node: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> Option<Violation> {
    let callee = ast::skip_outer_expressions(
        node.as_new_expression().expression(),
        OuterExpressionKinds::Parentheses | OuterExpressionKinds::Assertions,
    );
    if callee.kind() != Kind::Identifier
        || !matches!(callee.text(), "Symbol" | "BigInt")
        || runtime_shadowed(callee, file, checker)
    {
        return None;
    }
    Some(Violation::new(
        callee,
        "noNewNonconstructor",
        format!("`{}` cannot be called as a constructor.", callee.text()),
    ))
}
fn bad_direct_wrapper(node: P<Node>) -> Option<P<Node>> {
    let mut current = node;
    let mut count = 0;
    let mut satisfies = false;
    while let Some(parent) = current.parent() {
        match parent.kind() {
            Kind::AsExpression | Kind::TypeAssertionExpression | Kind::NonNullExpression => {
                count += 1
            }
            Kind::SatisfiesExpression => satisfies = true,
            Kind::ParenthesizedExpression => {}
            Kind::BinaryExpression => {
                return ((count > 1 || satisfies)
                    && parent.as_binary_expression().left() == current
                    && ast::is_assignment_operator(
                        parent.as_binary_expression().operator_token().kind(),
                    ))
                .then_some(parent);
            }
            Kind::PrefixUnaryExpression => {
                return ((count > 1 || satisfies)
                    && parent.as_prefix_unary_expression().operand() == current
                    && matches!(
                        parent.as_prefix_unary_expression().operator(),
                        Kind::PlusPlusToken | Kind::MinusMinusToken
                    ))
                .then_some(parent);
            }
            Kind::PostfixUnaryExpression => {
                return ((count > 1 || satisfies)
                    && parent.as_postfix_unary_expression().operand() == current
                    && matches!(
                        parent.as_postfix_unary_expression().operator(),
                        Kind::PlusPlusToken | Kind::MinusMinusToken
                    ))
                .then_some(parent);
            }
            _ => return None,
        }
        current = parent;
    }
    None
}
fn enclosing_write_pattern(mut node: P<Node>) -> bool {
    while let Some(parent) = node.parent() {
        match parent.kind() {
            Kind::BinaryExpression => {
                let binary = parent.as_binary_expression();
                if binary.operator_token().kind() == Kind::EqualsToken
                    && binary.left() == node
                    && matches!(
                        node.kind(),
                        Kind::ArrayLiteralExpression | Kind::ObjectLiteralExpression
                    )
                {
                    return true;
                }
                if ast::is_assignment_operator(binary.operator_token().kind())
                    && binary.left() != node
                {
                    return false;
                }
            }
            Kind::ForInStatement | Kind::ForOfStatement => {
                return parent.as_for_in_or_of_statement().initializer() == node
                    && matches!(
                        node.kind(),
                        Kind::ArrayLiteralExpression | Kind::ObjectLiteralExpression
                    );
            }
            Kind::CallExpression if parent.expression() != Some(node) => return false,
            Kind::PropertyAssignment if parent.initializer() != Some(node) => return false,
            Kind::ShorthandPropertyAssignment if parent.name() != Some(node) => return false,
            Kind::PropertyAccessExpression
            | Kind::ElementAccessExpression
            | Kind::ComputedPropertyName
            | Kind::Decorator => return false,
            Kind::ParenthesizedExpression
            | Kind::AsExpression
            | Kind::TypeAssertionExpression
            | Kind::NonNullExpression
            | Kind::SatisfiesExpression
                if parent.expression() != Some(node) =>
            {
                return false;
            }
            _ => {}
        }
        node = parent;
    }
    false
}
fn commonjs(file: P<SourceFile>) -> bool {
    file.file_name().ends_with(".cjs")
}
fn readonly_global(name: &str, file: P<SourceFile>) -> bool {
    ECMASCRIPT_GLOBALS.contains(&name)
        || (commonjs(file) && matches!(name, "global" | "module" | "require"))
}
fn global_assign(node: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> Option<Violation> {
    if !write_reference(node) || !readonly_global(node.text(), file) {
        return None;
    }
    if bad_direct_wrapper(node).is_some_and(|write| !enclosing_write_pattern(write)) {
        return None;
    }
    let resolved = resolve_reference(node, checker);
    if resolved.is_some_and(|symbol| local_value(symbol, file)) {
        return None;
    }
    if let Some(global) = checker.get_global_symbol_exported(node.text(), SymbolFlags::Value, None)
    {
        if resolved != Some(global) {
            return None;
        }
    } else if lexically_shadowed(node, node.text()) {
        return None;
    }
    Some(Violation::new(
        node,
        "globalShouldNotBeModified",
        format!("Read-only global '{}' should not be modified.", node.text()),
    ))
}

// Frozen ESLint 10.8 latest (ECMAScript 2026), rule/globals.go.
pub(crate) const ECMASCRIPT_GLOBALS: &[&str] = &[
    "Array",
    "Boolean",
    "constructor",
    "Date",
    "decodeURI",
    "decodeURIComponent",
    "encodeURI",
    "encodeURIComponent",
    "Error",
    "escape",
    "eval",
    "EvalError",
    "Function",
    "hasOwnProperty",
    "Infinity",
    "isFinite",
    "isNaN",
    "isPrototypeOf",
    "Math",
    "NaN",
    "Number",
    "Object",
    "parseFloat",
    "parseInt",
    "propertyIsEnumerable",
    "RangeError",
    "ReferenceError",
    "RegExp",
    "String",
    "SyntaxError",
    "toLocaleString",
    "toString",
    "TypeError",
    "undefined",
    "unescape",
    "URIError",
    "valueOf",
    "JSON",
    "ArrayBuffer",
    "DataView",
    "Float32Array",
    "Float64Array",
    "Int16Array",
    "Int32Array",
    "Int8Array",
    "Intl",
    "Map",
    "Promise",
    "Proxy",
    "Reflect",
    "Set",
    "Symbol",
    "Uint16Array",
    "Uint32Array",
    "Uint8Array",
    "Uint8ClampedArray",
    "WeakMap",
    "WeakSet",
    "Atomics",
    "SharedArrayBuffer",
    "BigInt",
    "BigInt64Array",
    "BigUint64Array",
    "globalThis",
    "AggregateError",
    "FinalizationRegistry",
    "WeakRef",
    "Float16Array",
    "Iterator",
    "AsyncDisposableStack",
    "DisposableStack",
    "SuppressedError",
    "Temporal",
];
fn variable_scope(node: P<Node>) -> Option<P<Node>> {
    ancestor(node, |node| {
        ast::is_function_like_or_class_static_block_declaration(node)
            || matches!(node.kind(), Kind::SourceFile | Kind::ModuleBlock)
    })
}
fn block_scope(node: P<Node>) -> Option<P<Node>> {
    ancestor(node, |node| {
        matches!(
            node.kind(),
            Kind::SourceFile
                | Kind::Block
                | Kind::ModuleBlock
                | Kind::ForStatement
                | Kind::ForInStatement
                | Kind::ForOfStatement
                | Kind::SwitchStatement
        )
    })
}
fn execution_scope(mut node: P<Node>) -> Option<P<Node>> {
    while node.parent().is_some() {
        let container = ast::get_this_container(node, true, false);
        if container.kind() == Kind::SourceFile
            || ast::is_function_like_declaration(container)
            || (container.kind() == Kind::PropertyDeclaration && !ast::is_static(container))
        {
            return Some(container);
        }
        node = container;
    }
    None
}
fn type_reference(node: P<Node>) -> bool {
    ast::is_part_of_type_node(node) || ast::is_part_of_type_query(node)
}
fn inside_own_initializer(name: P<Node>, location: usize) -> bool {
    let contains = |initializer: Option<P<Node>>| {
        initializer.is_some_and(|node| pos(node) <= location && location <= end(node))
    };
    let mut current = name.parent();
    while let Some(node) = current {
        match node.kind() {
            Kind::VariableDeclaration => {
                if contains(node.initializer()) {
                    return true;
                }
                return node
                    .parent()
                    .and_then(|list| list.parent())
                    .is_some_and(|parent| {
                        matches!(parent.kind(), Kind::ForInStatement | Kind::ForOfStatement)
                            && contains(Some(parent.as_for_in_or_of_statement().expression()))
                    });
            }
            Kind::BindingElement | Kind::Parameter => {
                if contains(node.initializer()) {
                    return true;
                }
            }
            Kind::FunctionDeclaration
            | Kind::FunctionExpression
            | Kind::ClassDeclaration
            | Kind::ClassExpression
            | Kind::ArrowFunction
            | Kind::CatchClause
            | Kind::ImportDeclaration
            | Kind::ExportDeclaration
            | Kind::MethodDeclaration
            | Kind::Constructor
            | Kind::GetAccessor
            | Kind::SetAccessor => return false,
            _ => {}
        }
        current = node.parent();
    }
    false
}
fn reopened_namespace(declaration: P<Node>, reference: P<Node>) -> bool {
    let mut current = declaration.parent();
    while let Some(outer) = current {
        if outer.kind() == Kind::ModuleDeclaration
            && let Some(symbol) = outer.symbol()
            && ancestor(reference, |inner| {
                inner.kind() == Kind::ModuleDeclaration
                    && inner != outer
                    && inner.symbol() == Some(symbol)
            })
            .is_some()
        {
            return true;
        }
        current = outer.parent();
    }
    false
}
fn invocation_result(mut node: P<Node>) -> P<Node> {
    while let Some(parent) = node.parent() {
        if ast::is_outer_expression(
            parent,
            OuterExpressionKinds::Parentheses | OuterExpressionKinds::Assertions,
        ) {
            node = parent;
            continue;
        }
        if parent.kind() == Kind::BinaryExpression {
            let binary = parent.as_binary_expression();
            let operator = binary.operator_token().kind();
            if (operator == Kind::CommaToken && binary.right() == node)
                || matches!(
                    operator,
                    Kind::BarBarToken | Kind::AmpersandAmpersandToken | Kind::QuestionQuestionToken
                )
            {
                node = parent;
                continue;
            }
        }
        if parent.kind() == Kind::ConditionalExpression {
            let conditional = parent.as_conditional_expression();
            if conditional.when_true() == node || conditional.when_false() == node {
                node = parent;
                continue;
            }
        }
        if parent.kind() == Kind::ReturnStatement
            || (parent.kind() == Kind::ArrowFunction && parent.body() == Some(node))
        {
            let owner = if parent.kind() == Kind::ReturnStatement {
                execution_scope(parent)
            } else {
                Some(parent)
            };
            if let Some(result) = owner.and_then(immediate_invocation) {
                node = result;
                continue;
            }
        }
        break;
    }
    node
}
fn is_callee(node: P<Node>) -> bool {
    node.parent().is_some_and(|parent| {
        matches!(parent.kind(), Kind::CallExpression | Kind::NewExpression)
            && parent.expression() == Some(node)
    })
}
fn immediate_invocation(node: P<Node>) -> Option<P<Node>> {
    let current = invocation_result(node);
    let parent = current.parent()?;
    if is_callee(current)
        || (parent.kind() == Kind::TaggedTemplateExpression
            && parent.as_tagged_template_expression().tag() == current)
    {
        return Some(parent);
    }
    if !ast::is_access_expression(parent)
        || parent.expression() != Some(current)
        || !is_callee(parent)
    {
        return None;
    }
    let file = ast::get_source_file_of_node(parent)?;
    let (_, _, name) = access(parent, file)?;
    match name.as_str() {
        "call" | "apply" => parent.parent(),
        "bind" => immediate_invocation(parent.parent()?),
        _ => None,
    }
}
fn has_initializer_call(
    name: P<Node>,
    function: P<Node>,
    binding: P<Node>,
    members: bool,
    index: &ReferenceIndex,
    active: &mut HashSet<P<Node>>,
) -> bool {
    if !active.insert(function) {
        return false;
    }
    let mut found = false;
    if let Some(symbol) = index.symbol(binding) {
        for &reference in index.refs(symbol) {
            if !inside_own_initializer(name, end(reference)) {
                continue;
            }
            let expression = invocation_result(reference);
            let parent = expression.parent();
            let invokes = immediate_invocation(reference).is_some()
                || parent.is_some_and(|parent| {
                    matches!(parent.kind(), Kind::CallExpression | Kind::NewExpression)
                })
                || (members
                    && parent.is_some_and(|parent| {
                        ast::is_access_expression(parent) && parent.expression() == Some(expression)
                    }));
            if invokes && executed_in_initializer(name, reference, index, active) {
                found = true;
                break;
            }
        }
    }
    active.remove(&function);
    found
}
fn executed_in_initializer(
    name: P<Node>,
    reference: P<Node>,
    index: &ReferenceIndex,
    active: &mut HashSet<P<Node>>,
) -> bool {
    let boundary = execution_scope(name);
    let mut scope = execution_scope(reference);
    while let Some(current) = scope.filter(|scope| Some(*scope) != boundary) {
        if immediate_invocation(current).is_some() {
            scope = execution_scope(current);
            continue;
        }
        let expression = invocation_result(current);
        if expression.parent().is_some_and(|parent| {
            matches!(parent.kind(), Kind::CallExpression | Kind::NewExpression)
        }) {
            scope = execution_scope(current);
            continue;
        }
        let binding = if current.kind() == Kind::FunctionDeclaration {
            current.name()
        } else {
            expression
                .parent()
                .filter(|parent| {
                    parent.kind() == Kind::VariableDeclaration
                        && parent.initializer() == Some(expression)
                })
                .and_then(|parent| parent.name())
        };
        if let Some(binding) = binding.filter(|binding| binding.kind() == Kind::Identifier) {
            return has_initializer_call(name, current, binding, false, index, active);
        }
        let mut owner = expression.parent();
        if owner.is_some_and(|owner| owner.kind() == Kind::PropertyAssignment) {
            owner = owner.and_then(|owner| owner.parent());
        }
        if let Some(owner) = owner.filter(|owner| {
            owner.kind() == Kind::ObjectLiteralExpression || ast::is_class_like(*owner)
        }) {
            if ast::is_class_like(owner)
                && owner.name().is_some_and(|binding| {
                    has_initializer_call(name, current, binding, true, index, active)
                })
            {
                return true;
            }
            let value = invocation_result(owner);
            if let Some(binding) = value
                .parent()
                .filter(|parent| parent.kind() == Kind::VariableDeclaration)
                .and_then(|parent| parent.name())
                .filter(|name| name.kind() == Kind::Identifier)
            {
                return has_initializer_call(name, current, binding, true, index, active);
            }
            if immediate_invocation(value).is_some()
                || value.parent().is_some_and(|parent| {
                    (ast::is_access_expression(parent) && parent.expression() == Some(value))
                        || parent.kind() == Kind::SpreadAssignment
                })
            {
                scope = execution_scope(owner);
                continue;
            }
        }
        return false;
    }
    true
}
fn unsafe_hoisted_function(name: P<Node>, refs: &[P<Node>], index: &ReferenceIndex) -> bool {
    let boundary = execution_scope(name);
    let mut checked = None;
    for &reference in refs {
        if pos(reference) < pos(name) || type_reference(reference) {
            continue;
        }
        let mut scope = execution_scope(reference);
        while let Some(current) = scope.filter(|scope| Some(*scope) != boundary) {
            if current.kind() == Kind::FunctionDeclaration
                && current.name().is_some()
                && Some(current) != checked
            {
                checked = Some(current);
                if let Some(symbol) = index.symbol(current.name().unwrap()) {
                    for &invocation in index.refs(symbol) {
                        if (pos(invocation) < pos(name)
                            || (inside_own_initializer(name, end(invocation))
                                && executed_in_initializer(
                                    name,
                                    invocation,
                                    index,
                                    &mut HashSet::new(),
                                )))
                            && immediate_invocation(invocation).is_some()
                        {
                            return true;
                        }
                    }
                }
            }
            scope = execution_scope(current);
        }
    }
    false
}
fn in_loop(mut node: P<Node>) -> bool {
    while let Some(parent) = node.parent() {
        if matches!(
            parent.kind(),
            Kind::ForStatement
                | Kind::ForInStatement
                | Kind::ForOfStatement
                | Kind::WhileStatement
                | Kind::DoStatement
        ) {
            return true;
        }
        if ast::is_function_like_or_class_static_block_declaration(parent)
            || parent.kind() == Kind::SourceFile
        {
            return false;
        }
        node = parent;
    }
    false
}
fn redeclared_by_function(name: P<Node>, own: P<Symbol>, boundary: P<Node>) -> bool {
    let last = if boundary.kind() == Kind::ModuleBlock {
        boundary
            .parent()
            .filter(|parent| parent.kind() == Kind::ModuleDeclaration)
            .unwrap_or(boundary)
    } else {
        boundary
    };
    let conflicts = |symbol: Option<P<Symbol>>| {
        symbol.is_some_and(|symbol| {
            symbol != own
                && symbol
                    .declarations()
                    .iter()
                    .any(|node| node.kind() == Kind::FunctionDeclaration)
        })
    };
    let mut current = name.parent();
    while let Some(node) = current {
        if conflicts(
            node.locals()
                .and_then(|locals| locals.get().get(name.text())),
        ) || conflicts(
            node.symbol()
                .and_then(|symbol| symbol.exports())
                .and_then(|exports| exports.get().get(name.text())),
        ) {
            return true;
        }
        if node == last {
            break;
        }
        current = node.parent();
    }
    false
}
fn var_fixable(list: P<Node>, index: &ReferenceIndex) -> bool {
    let Some(statement) = list.parent() else {
        return false;
    };
    if statement.kind() == Kind::VariableStatement {
        if ast::has_syntactic_modifier(statement, ModifierFlags::Export) {
            return false;
        }
        let Some(parent) = statement.parent() else {
            return false;
        };
        if !matches!(
            parent.kind(),
            Kind::SourceFile | Kind::Block | Kind::ModuleBlock
        ) {
            return false;
        }
    } else if !matches!(
        statement.kind(),
        Kind::ForStatement | Kind::ForInStatement | Kind::ForOfStatement
    ) {
        return false;
    }
    let mut names = Vec::new();
    for &declaration in list.as_variable_declaration_list().declarations().nodes() {
        bindings(declaration.as_variable_declaration().name(), &mut names);
    }
    for &name in &names {
        let Some(symbol) = binding_symbol(name) else {
            return false;
        };
        let Some(scope) = variable_scope(name) else {
            return false;
        };
        if symbol.declarations().len() >= 2
            || redeclared_by_function(name, symbol, scope)
            || name.text() == "let"
        {
            return false;
        }
        let mut current = name.parent();
        while let Some(node) = current {
            if node.kind() == Kind::CatchClause
                && let Some(declaration) = node.as_catch_clause().variable_declaration()
            {
                let mut caught = Vec::new();
                bindings(declaration.as_variable_declaration().name(), &mut caught);
                if caught.iter().any(|binding| binding.text() == name.text()) {
                    return false;
                }
            }
            if node == scope {
                break;
            }
            current = node.parent();
        }
    }
    let looping = in_loop(list);
    for name in names {
        let symbol = binding_symbol(name).unwrap();
        let refs: Vec<_> = index
            .refs(symbol)
            .iter()
            .copied()
            .filter(|reference| {
                !reference
                    .parent()
                    .is_some_and(|parent| parent.kind() == Kind::NamespaceExportDeclaration)
                    && !(name.flags().intersects(NodeFlags::Ambient)
                        && type_reference(*reference)
                        && reopened_namespace(name, *reference))
            })
            .collect();
        let root = ast::get_root_declaration(name.parent().unwrap());
        let direct_function = root.kind() == Kind::VariableDeclaration
            && root.initializer().is_some_and(|initializer| {
                matches!(
                    ast::skip_outer_expressions(
                        initializer,
                        OuterExpressionKinds::Parentheses | OuterExpressionKinds::Assertions
                    )
                    .kind(),
                    Kind::FunctionExpression | Kind::ArrowFunction
                )
            });
        for &reference in &refs {
            if !type_reference(reference) && inside_own_initializer(name, end(reference)) {
                let deferred = direct_function
                    && root.initializer().is_some_and(|initializer| {
                        pos(reference) >= pos(initializer) && end(reference) <= end(initializer)
                    });
                if !deferred {
                    return false;
                }
            }
            if block_scope(list)
                .is_some_and(|scope| pos(reference) < pos(scope) || pos(reference) >= end(scope))
            {
                return false;
            }
            if pos(reference) < pos(name) && reference != name && !type_reference(reference) {
                return false;
            }
            if looping
                && !type_reference(reference)
                && execution_scope(reference) != execution_scope(list)
            {
                return false;
            }
        }
        if unsafe_hoisted_function(name, &refs, index) {
            return false;
        }
    }
    if looping
        && !loop_assignee(list)
        && list
            .as_variable_declaration_list()
            .declarations()
            .nodes()
            .iter()
            .any(|declaration| declaration.initializer().is_none())
    {
        return false;
    }
    true
}
fn no_var(file: P<SourceFile>, index: &ReferenceIndex) -> Vec<Violation> {
    let mut reports = Vec::new();
    for &list in &index.nodes {
        if list.kind() != Kind::VariableDeclarationList
            || list.flags().intersects(NodeFlags::BlockScoped)
        {
            continue;
        }
        let statement = list.parent();
        if statement.is_some_and(|statement| {
            statement.kind() == Kind::VariableStatement
                && statement.parent().is_some_and(|parent| {
                    parent.kind() == Kind::ModuleBlock
                        && parent
                            .parent()
                            .is_some_and(ast::is_global_scope_augmentation)
                })
        }) {
            continue;
        }
        let mut span = [start(list, file), end(list)];
        if let Some(statement) =
            statement.filter(|statement| statement.kind() == Kind::VariableStatement)
        {
            span = [start(statement, file), end(statement)];
            if ast::has_syntactic_modifier(statement, ModifierFlags::Export) {
                span[0] = start(list, file);
                if let Some(modifiers) = statement.modifiers()
                    && let Some(&modifier) = modifiers
                        .nodes()
                        .iter()
                        .find(|node| node.kind() != Kind::ExportKeyword)
                {
                    span[0] = start(modifier, file);
                }
            }
        }
        let violation = Violation::new(
            list,
            "unexpectedVar",
            "Unexpected var, use let or const instead.",
        )
        .with_byte_range(span);
        reports.push(if var_fixable(list, index) {
            let begin = start(list, file);
            violation.with_fix(vec![Edit::replace([begin, begin + 3], "let")])
        } else {
            violation
        });
    }
    reports
}

const NONCALLABLE: &[&str] = &["Atomics", "JSON", "Math", "Reflect", "Intl", "Temporal"];
const GLOBAL_OBJECTS: &[&str] = &["global", "globalThis", "self", "window"];
#[derive(Clone, Copy)]
enum TrackedValue {
    GlobalObject,
    NonCallable(&'static str),
}
#[derive(Default)]
struct TraceWalk {
    variables: HashSet<P<Symbol>>,
    globals: HashSet<String>,
}
struct ObjectCalls<'a> {
    file: P<SourceFile>,
    checker: &'a mut Checker,
    index: ReferenceIndex,
    local_roots: HashSet<String>,
    identifiers: HashMap<String, Vec<P<Node>>>,
    calls: Vec<(P<Node>, &'static str)>,
}
fn local_value(symbol: P<Symbol>, file: P<SourceFile>) -> bool {
    symbol.declarations().iter().any(|&decl| {
        if ast::get_source_file_of_node(decl) != Some(file)
            || matches!(
                decl.kind(),
                Kind::InterfaceDeclaration | Kind::TypeAliasDeclaration
            )
        {
            return false;
        }
        ancestor(decl, |parent| {
            parent.kind() == Kind::ModuleDeclaration
                && (ast::is_global_scope_augmentation(parent)
                    || ast::is_external_module_augmentation(parent))
        })
        .is_none()
    })
}
fn trackable_binding(node: P<Node>) -> Option<P<Symbol>> {
    binding_symbol(node).filter(|symbol| {
        symbol.flags().intersects(
            SymbolFlags::Variable
                | SymbolFlags::Function
                | SymbolFlags::Class
                | SymbolFlags::Enum
                | SymbolFlags::Module
                | SymbolFlags::Alias,
        )
    })
}
fn pass_through(node: P<Node>, parent: P<Node>) -> bool {
    match parent.kind() {
        Kind::ParenthesizedExpression
        | Kind::AsExpression
        | Kind::SatisfiesExpression
        | Kind::TypeAssertionExpression
        | Kind::NonNullExpression
        | Kind::PartiallyEmittedExpression
        | Kind::ExpressionWithTypeArguments => parent.expression() == Some(node),
        Kind::ConditionalExpression => {
            let conditional = parent.as_conditional_expression();
            conditional.when_true() == node || conditional.when_false() == node
        }
        Kind::BinaryExpression => {
            let binary = parent.as_binary_expression();
            match binary.operator_token().kind() {
                Kind::BarBarToken | Kind::AmpersandAmpersandToken | Kind::QuestionQuestionToken => {
                    binary.left() == node || binary.right() == node
                }
                Kind::CommaToken => binary.right() == node,
                _ => false,
            }
        }
        _ => false,
    }
}
impl<'a> ObjectCalls<'a> {
    fn new(file: P<SourceFile>, checker: &'a mut Checker) -> Self {
        let index = ReferenceIndex::new(file, checker);
        let mut local_roots = HashSet::new();
        let mut identifiers: HashMap<String, Vec<P<Node>>> = HashMap::new();
        for &node in &index.nodes {
            if node.kind() != Kind::Identifier {
                continue;
            }
            if is_reference(node) {
                identifiers
                    .entry(node.text().into())
                    .or_default()
                    .push(node);
            } else if (NONCALLABLE.contains(&node.text()) || GLOBAL_OBJECTS.contains(&node.text()))
                && trackable_binding(node).is_some()
            {
                local_roots.insert(node.text().into());
            }
        }
        Self {
            file,
            checker,
            index,
            local_roots,
            identifiers,
            calls: Vec::new(),
        }
    }
    fn known_global(&self, name: &str) -> bool {
        readonly_global(name, self.file) || (commonjs(self.file) && name == "exports")
    }
    fn local_symbol(&self, node: P<Node>) -> Option<P<Symbol>> {
        self.index
            .symbol(node)
            .filter(|&symbol| local_value(symbol, self.file))
    }
    fn global_reference(&self, node: P<Node>) -> bool {
        if node.kind() != Kind::Identifier || !is_reference(node) {
            return false;
        }
        let root = NONCALLABLE.contains(&node.text()) || GLOBAL_OBJECTS.contains(&node.text());
        if root && !self.local_roots.contains(node.text()) {
            return true;
        }
        if let Some(symbol) = self.index.symbol(node) {
            return !local_value(symbol, self.file);
        }
        !lexically_shadowed(node, node.text())
    }
    fn collect(&mut self) {
        for &name in NONCALLABLE {
            self.root(name, TrackedValue::NonCallable(name));
        }
        for &name in GLOBAL_OBJECTS {
            self.root(name, TrackedValue::GlobalObject);
        }
    }
    fn root(&mut self, name: &str, value: TrackedValue) {
        if !self.known_global(name) {
            return;
        }
        let references: Vec<_> = self
            .identifiers
            .get(name)
            .into_iter()
            .flatten()
            .copied()
            .filter(|&node| self.global_reference(node))
            .collect();
        if references.iter().any(|&node| write_reference(node)) {
            return;
        }
        for node in references {
            self.expression(node, value, &mut TraceWalk::default());
        }
    }
    fn property(&mut self, node: P<Node>) -> Option<String> {
        if node.kind() == Kind::ComputedPropertyName {
            super::rsbuild_await_checks::static_string_no_bindings(
                node.expression()?,
                self.file,
                self.checker,
            )
        } else {
            super::rsbuild_declaration_checks::static_property(node, self.file)
        }
    }
    fn access_name(&mut self, node: P<Node>) -> Option<String> {
        if node.kind() == Kind::ElementAccessExpression {
            super::rsbuild_await_checks::static_string_no_bindings(
                node.as_element_access_expression().argument_expression(),
                self.file,
                self.checker,
            )
        } else {
            node.name().map(|name| name.text().into())
        }
    }
    fn expression(&mut self, mut node: P<Node>, value: TrackedValue, walk: &mut TraceWalk) {
        while let Some(parent) = node.parent().filter(|&parent| pass_through(node, parent)) {
            node = parent;
        }
        let Some(parent) = node.parent() else {
            return;
        };
        if ast::is_access_expression(parent) && parent.expression() == Some(node) {
            if matches!(value, TrackedValue::GlobalObject)
                && let Some(name) = self
                    .access_name(parent)
                    .and_then(|name| NONCALLABLE.iter().copied().find(|&root| root == name))
            {
                self.expression(parent, TrackedValue::NonCallable(name), walk);
            }
            return;
        }
        match parent.kind() {
            Kind::CallExpression | Kind::NewExpression if parent.expression() == Some(node) => {
                if let TrackedValue::NonCallable(name) = value {
                    self.calls.push((parent, name));
                }
            }
            Kind::BinaryExpression => {
                let binary = parent.as_binary_expression();
                if binary.right() == node
                    && ast::is_assignment_operator(binary.operator_token().kind())
                {
                    self.target(binary.left(), value, walk);
                    self.expression(parent, value, walk);
                }
            }
            Kind::VariableDeclaration | Kind::Parameter | Kind::BindingElement
                if parent.initializer() == Some(node) =>
            {
                if let Some(name) = parent.name() {
                    self.target(name, value, walk);
                }
            }
            Kind::ShorthandPropertyAssignment
                if parent
                    .as_shorthand_property_assignment()
                    .object_assignment_initializer()
                    == Some(node) =>
            {
                if let Some(name) = parent.name() {
                    self.target(name, value, walk);
                }
            }
            _ => {}
        }
    }
    fn target(&mut self, node: P<Node>, value: TrackedValue, walk: &mut TraceWalk) {
        let node = ast::skip_parentheses(node);
        match node.kind() {
            Kind::Identifier => {
                if let Some(symbol) = trackable_binding(node).or_else(|| self.local_symbol(node)) {
                    self.variable(symbol, value, walk);
                } else if self.known_global(node.text()) && self.global_reference(node) {
                    self.global_variable(node.text(), value, walk);
                }
            }
            Kind::ObjectBindingPattern if matches!(value, TrackedValue::GlobalObject) => {
                for &element in node.elements() {
                    if element.kind() != Kind::BindingElement {
                        continue;
                    }
                    let binding = element.as_binding_element();
                    if binding.dot_dot_dot_token().is_some() {
                        continue;
                    }
                    let Some(target) = binding.name() else {
                        continue;
                    };
                    let property = binding.property_name().unwrap_or(target);
                    if let Some(name) = self
                        .property(property)
                        .and_then(|name| NONCALLABLE.iter().copied().find(|&root| root == name))
                    {
                        self.target(target, TrackedValue::NonCallable(name), walk);
                    }
                }
            }
            Kind::ObjectLiteralExpression if matches!(value, TrackedValue::GlobalObject) => {
                for &property in node.properties() {
                    let target = match property.kind() {
                        Kind::PropertyAssignment => {
                            Some(property.as_property_assignment().initializer())
                        }
                        Kind::ShorthandPropertyAssignment => property.name(),
                        _ => None,
                    };
                    let Some(target) = target else {
                        continue;
                    };
                    if let Some(name) = property
                        .name()
                        .and_then(|name| self.property(name))
                        .and_then(|name| NONCALLABLE.iter().copied().find(|&root| root == name))
                    {
                        self.target(target, TrackedValue::NonCallable(name), walk);
                    }
                }
            }
            Kind::BinaryExpression
                if node.as_binary_expression().operator_token().kind() == Kind::EqualsToken =>
            {
                self.target(node.as_binary_expression().left(), value, walk)
            }
            _ => {}
        }
    }
    fn variable(&mut self, symbol: P<Symbol>, value: TrackedValue, walk: &mut TraceWalk) {
        let symbol = self.index.canonical_symbol(symbol);
        if !walk.variables.insert(symbol) {
            return;
        }
        let references = self.index.refs(symbol).to_vec();
        for reference in references {
            if !ast::is_write_only_access(reference) {
                self.expression(reference, value, walk);
            }
        }
        walk.variables.remove(&symbol);
    }
    fn global_variable(&mut self, name: &str, value: TrackedValue, walk: &mut TraceWalk) {
        if !walk.globals.insert(name.into()) {
            return;
        }
        let references = self.identifiers.get(name).cloned().unwrap_or_default();
        for reference in references {
            if !ast::is_write_only_access(reference) && self.global_reference(reference) {
                self.expression(reference, value, walk);
            }
        }
        walk.globals.remove(name);
    }
    fn report(mut self) -> Vec<Violation> {
        self.collect();
        self.calls.sort_by_key(|(node, _)| pos(*node));
        self.calls
            .into_iter()
            .map(|(call, global)| {
                let callee = ast::skip_parentheses(call.expression().expect("call expression"));
                let name = match callee.kind() {
                    Kind::Identifier => callee.text().into(),
                    Kind::PropertyAccessExpression => callee
                        .name()
                        .map(|name| name.text().into())
                        .unwrap_or("null".into()),
                    Kind::ElementAccessExpression => {
                        super::rsbuild_declaration_checks::static_expression(
                            ast::skip_parentheses(
                                callee.as_element_access_expression().argument_expression(),
                            ),
                            self.file,
                        )
                        .unwrap_or("null".into())
                    }
                    _ => "undefined".into(),
                };
                let (id, message) = if name == global {
                    ("unexpectedCall", format!("'{global}' is not a function."))
                } else {
                    (
                        "unexpectedRefCall",
                        format!("'{name}' is reference to '{global}', which is not a function."),
                    )
                };
                Violation::new(call, id, message)
            })
            .collect()
    }
}
