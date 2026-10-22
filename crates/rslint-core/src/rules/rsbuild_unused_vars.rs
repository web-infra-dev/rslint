//! Symbol-based unused binding analysis ported from frozen rslint e0a29f7.
//! The supported options are the defaults and Rsbuild's two `^_` patterns.
//! References are indexed by compiler binder identity, not identifier frequency.
use super::rsbuild_assignment_checks::resolve_reference;
use crate::{Violation, edits::Edit};
use serde_json::Value;
use std::collections::{HashMap, HashSet};
use tsrs_ast::{
    self as ast, Kind, ModifierFlags, Node, NodeFlags, SourceFile, Symbol, SymbolFlags,
};
use tsrs_checker::Checker;
use tsrs_core::P;

pub(crate) const NAMES: &[&str] = &["@typescript-eslint/no-unused-vars"];

pub(crate) fn validate_options(options: &[Value]) -> Result<(), String> {
    if options.is_empty() {
        return Ok(());
    }
    if options.len() != 1 {
        return Err("no-unused-vars accepts one options object".into());
    }
    let object = options[0]
        .as_object()
        .ok_or("no-unused-vars requires an options object")?;
    for (key, value) in object {
        let accepted = match key.as_str() {
            "argsIgnorePattern" | "varsIgnorePattern" => matches!(value.as_str(), Some("" | "^_")),
            "vars" => value == "all",
            "args" => value == "after-used",
            "caughtErrors" => value == "all",
            "caughtErrorsIgnorePattern" | "destructuredArrayIgnorePattern" => value == "",
            "ignoreRestSiblings"
            | "ignoreClassWithStaticInitBlock"
            | "ignoreUsingDeclarations"
            | "reportUsedIgnorePattern" => value == false,
            "enableAutofixRemoval" => value.as_object().is_some_and(|object| {
                object.is_empty()
                    || object.len() == 1 && object.get("imports") == Some(&Value::Bool(false))
            }),
            _ => false,
        };
        if !accepted {
            return Err(format!(
                "no-unused-vars option {key} is not supported by the Rsbuild migration"
            ));
        }
    }
    Ok(())
}

#[derive(Clone, Default)]
struct References {
    reads: Vec<P<Node>>,
    writes: Vec<P<Node>>,
}
#[derive(Clone, Copy)]
struct Binding {
    name: P<Node>,
    declaration: P<Node>,
    symbol: Option<P<Symbol>>,
}
#[derive(Clone, Copy, PartialEq)]
enum Assignment {
    None,
    Write,
    ReadWrite,
}

pub(crate) fn check(
    name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    options: &[Value],
) -> Vec<Violation> {
    if name != NAMES[0] || node.kind() != Kind::SourceFile {
        return Vec::new();
    }
    let mut nodes = Vec::new();
    walk(node, &mut |node| nodes.push(node));
    let mut bindings = Vec::new();
    let mut type_parameters = Vec::new();
    for &node in &nodes {
        if candidate(node, file, checker)
            && let Some(name) = node.name()
        {
            if node.kind() == Kind::TypeParameter {
                collect_binding(name, node, checker, &mut type_parameters);
            } else {
                collect_binding(name, node, checker, &mut bindings);
            }
        }
    }
    // The baseline defers type parameters until all signature parameters exist.
    let signature_used = if type_parameters.is_empty() {
        HashSet::new()
    } else {
        signature_type_usage(node)
    };
    bindings.extend(type_parameters);
    // Parameter properties are not diagnostics candidates, but real reads of
    // their parameter/property symbols still define the after-used boundary.
    let mut property_parameters = HashMap::<P<Symbol>, P<Node>>::new();
    for &parameter in &nodes {
        if parameter.kind() == Kind::Parameter
            && let Some(owner) = parameter.parent()
            && ast::is_parameter_property_declaration(parameter, owner)
            && let Some(name) = parameter
                .name()
                .filter(|name| name.kind() == Kind::Identifier)
        {
            let (local, property) =
                checker.get_symbols_of_parameter_property_declaration(parameter, name.text());
            property_parameters.insert(checker.get_merged_symbol(local), parameter);
            property_parameters.insert(checker.get_merged_symbol(property), parameter);
        }
    }
    let mut used_property_parameters = HashSet::new();
    let mut references = HashMap::<P<Symbol>, References>::new();
    let mut exports = HashSet::new();
    for &identifier in &nodes {
        if identifier.kind() != Kind::Identifier {
            continue;
        }
        let Some(parent) = identifier.parent() else {
            continue;
        };
        if parent.kind() == Kind::ExportSpecifier {
            let Some(export) = ancestor(parent, Kind::ExportDeclaration) else {
                continue;
            };
            if export.as_export_declaration().module_specifier().is_some() {
                continue;
            }
            let local = parent
                .as_export_specifier()
                .property_name()
                .unwrap_or(parent.name().unwrap());
            if identifier != local {
                continue;
            }
            let meaning = if ast::is_type_only_import_or_export_declaration(parent) {
                SymbolFlags::Type | SymbolFlags::Namespace | SymbolFlags::Alias
            } else {
                SymbolFlags::Value | SymbolFlags::Type | SymbolFlags::Namespace | SymbolFlags::Alias
            };
            if let Some(symbol) = export_reference(local, meaning, checker) {
                let symbol = checker.get_merged_symbol(symbol);
                exports.insert(symbol);
                references.entry(symbol).or_default().reads.push(identifier);
            }
            continue;
        }
        if property_name(identifier)
            && !property_parameters.is_empty()
            && let Some(symbol) = checker.get_symbol_at_location_exported(identifier)
            && let Some(parameter) = property_parameters.get(&checker.get_merged_symbol(symbol))
        {
            used_property_parameters.insert(*parameter);
        }
        if import_type_syntax(identifier)
            || jsx_non_reference(identifier)
            || property_name(identifier)
            || (!matches!(
                parent.kind(),
                Kind::ShorthandPropertyAssignment | Kind::NamespaceExportDeclaration
            ) && ast::is_declaration_name(identifier))
        {
            continue;
        }
        let resolved = reference_symbol(identifier, checker);
        if let Some(symbol) = resolved {
            let symbol = checker.get_merged_symbol(symbol);
            let info = references.entry(symbol).or_default();
            let assignment = assignment(identifier);
            if assignment != Assignment::Write
                && let Some(parameter) = property_parameters.get(&symbol)
            {
                used_property_parameters.insert(*parameter);
            }
            if assignment != Assignment::Write {
                info.reads.push(identifier);
            }
            if assignment != Assignment::None
                && !(assignment == Assignment::Write
                    && parent.kind() == Kind::ShorthandPropertyAssignment)
            {
                info.writes.push(identifier);
            }
        }
    }
    let first_jsx = nodes.iter().copied().find(|node| {
        matches!(
            node.kind(),
            Kind::JsxElement | Kind::JsxSelfClosingElement | Kind::JsxFragment
        )
    });
    let first_fragment = nodes
        .iter()
        .copied()
        .find(|node| node.kind() == Kind::JsxFragment);
    // Match the frozen rule's compiler-option lookup. Source @jsx pragmas are
    // intentionally not treated as unused-vars factory configuration.
    let factory = first_jsx.map(|_| {
        let configured = &checker.compiler_options.jsx_factory;
        if configured.is_empty() {
            "React".to_owned()
        } else {
            configured.split('.').next().unwrap_or("").to_owned()
        }
    });
    let fragment_factory = first_fragment.and_then(|_| {
        let configured = &checker.compiler_options.jsx_fragment_factory;
        (!configured.is_empty()).then(|| configured.split('.').next().unwrap_or("").to_owned())
    });
    let mut seen = HashSet::new();
    let mut unused_imports = HashSet::new();
    let mut out = Vec::new();
    for binding in &bindings {
        let Binding {
            name,
            declaration,
            symbol,
        } = *binding;
        if directly_exported(declaration) {
            continue;
        }
        if let Some(symbol) = symbol {
            if !seen.insert(symbol) {
                continue;
            }
            if exports.contains(&symbol)
                || symbol.declarations().iter().copied().any(directly_exported)
            {
                continue;
            }
        }
        let info = symbol
            .and_then(|symbol| references.get(&symbol))
            .cloned()
            .unwrap_or_default();
        if declaration.kind() == Kind::Parameter
            && before_last_used(
                declaration,
                &bindings,
                &references,
                &used_property_parameters,
            )
        {
            continue;
        }
        let declarations = symbol.map(|symbol| symbol.declarations()).unwrap_or(&[]);
        let mut only_type = false;
        let mut used = false;
        for &reference in &info.reads {
            if reference == name
                || inside_own_declaration(reference, declaration, declarations)
                || symbol.is_some_and(|symbol| {
                    !info.writes.is_empty() && self_modification(reference, symbol, name, checker)
                })
            {
                continue;
            }
            if type_only(reference) && !signature_parameter(reference) {
                only_type = true;
            } else {
                used = true;
                break;
            }
        }
        if is_import(declaration)
            && (factory.as_deref() == Some(name.text())
                || fragment_factory.as_deref() == Some(name.text()))
        {
            used = true;
        }
        if only_type && type_or_import(declaration) {
            used = true;
        }
        if !used
            && declaration.kind() == Kind::TypeParameter
            && (signature_used.contains(&declaration)
                || declarations
                    .iter()
                    .any(|declaration| signature_used.contains(declaration)))
        {
            used = true;
        }
        if used {
            continue;
        }
        let category = if ast::is_catch_clause_variable_declaration_or_binding_element(declaration)
        {
            "caught errors"
        } else if ast::is_part_of_parameter_declaration(declaration) {
            "args"
        } else {
            "vars"
        };
        let ignored = options
            .first()
            .and_then(|options| {
                options.get(if category == "args" {
                    "argsIgnorePattern"
                } else if category == "vars" {
                    "varsIgnorePattern"
                } else {
                    "caughtErrorsIgnorePattern"
                })
            })
            .and_then(Value::as_str)
            == Some("^_");
        if ignored && name.text().starts_with('_') {
            continue;
        }
        let additional = if ignored {
            format!(". Allowed unused {category} must match /^_/u")
        } else {
            String::new()
        };
        let assigned = has_assignment(declaration, &info.writes);
        let action = if assigned {
            "assigned a value"
        } else {
            "defined"
        };
        let report_node = info
            .writes
            .iter()
            .rev()
            .copied()
            .find(|&reference| enclosing_scope(reference) == enclosing_scope(declaration))
            .unwrap_or(name);
        let mut report = Violation::new(
            report_node,
            if only_type {
                "usedOnlyAsType"
            } else {
                "unusedVar"
            },
            format!(
                "'{}' is {action} but {}{additional}.",
                name.text(),
                if only_type {
                    "only used as a type"
                } else {
                    "never used"
                }
            ),
        );
        if !only_type && is_import(declaration) {
            unused_imports.insert(declaration);
            if let Some((id, message, edit)) = import_suggestion(declaration, file, &unused_imports)
            {
                report = report.with_suggestion(id, message, vec![edit]);
            }
        }
        out.push(report);
    }
    out
}

fn walk(node: P<Node>, visit: &mut impl FnMut(P<Node>)) {
    if node.flags().intersects(NodeFlags::Reparsed) {
        return;
    }
    visit(node);
    node.for_each_child(&mut |child| {
        walk(child, visit);
        false
    });
}
fn ancestor(mut node: P<Node>, kind: Kind) -> Option<P<Node>> {
    while let Some(parent) = node.parent() {
        if parent.kind() == kind {
            return Some(parent);
        }
        node = parent;
    }
    None
}
fn contains(parent: P<Node>, mut node: P<Node>) -> bool {
    loop {
        if parent == node {
            return true;
        }
        let Some(next) = node.parent() else {
            return false;
        };
        node = next;
    }
}
fn collect_binding(
    name: P<Node>,
    declaration: P<Node>,
    checker: &mut Checker,
    bindings: &mut Vec<Binding>,
) {
    if name.kind() == Kind::Identifier {
        let symbol = declaration
            .symbol()
            .or_else(|| checker.get_symbol_at_location_exported(name))
            .map(|symbol| checker.get_merged_symbol(symbol));
        bindings.push(Binding {
            name,
            declaration,
            symbol,
        });
    } else if matches!(
        name.kind(),
        Kind::ObjectBindingPattern | Kind::ArrayBindingPattern
    ) {
        for &element in name.elements() {
            if element.kind() == Kind::BindingElement
                && let Some(name) = element.name()
            {
                collect_binding(name, element, checker, bindings);
            }
        }
    }
}
fn explicit_exports(container: P<Node>) -> bool {
    container.statements().iter().any(|node| {
        matches!(
            node.kind(),
            Kind::ExportDeclaration | Kind::ExportAssignment
        )
    })
}
fn ambient(node: P<Node>, file: P<SourceFile>) -> bool {
    let module = ancestor(node, Kind::ModuleBlock);
    if file.is_declaration_file() && module.is_none() {
        return !explicit_exports(file.as_node());
    }
    if !node.flags().intersects(NodeFlags::Ambient) {
        return false;
    }
    let Some(block) = module else {
        return false;
    };
    let mut owner = block.parent();
    let mut is_ambient = file.is_declaration_file();
    while let Some(node) = owner {
        if node.kind() == Kind::ModuleDeclaration
            && (ast::get_combined_modifier_flags(node).intersects(ModifierFlags::Ambient)
                || ast::is_global_scope_augmentation(node))
        {
            is_ambient = true;
            break;
        }
        owner = node.parent();
    }
    is_ambient && !explicit_exports(block)
}
fn candidate(node: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> bool {
    match node.kind() {
        Kind::VariableDeclaration => {
            if ambient(node, file) {
                return false;
            }
            if node
                .name()
                .is_some_and(|name| name.kind() == Kind::Identifier)
                && node.parent().is_some_and(|parent| {
                    parent.kind() == Kind::VariableDeclarationList
                        && parent.parent().is_some_and(|parent| {
                            matches!(parent.kind(), Kind::ForInStatement | Kind::ForOfStatement)
                                && return_first(parent.as_for_in_or_of_statement().statement())
                        })
                })
            {
                return false;
            }
            true
        }
        Kind::Parameter => {
            node.parent().is_some_and(|owner| {
                !bodyless_parameter(node) && !ast::is_parameter_property_declaration(node, owner)
            }) && node.name().is_some_and(|name| {
                name.kind() != Kind::ThisKeyword
                    && (name.kind() != Kind::Identifier || name.text() != "this")
            })
        }
        Kind::FunctionDeclaration => {
            node.name().is_some() && (node.body().is_some() || !ambient(node, file))
        }
        Kind::ClassDeclaration
        | Kind::InterfaceDeclaration
        | Kind::TypeAliasDeclaration
        | Kind::EnumDeclaration => !ambient(node, file),
        Kind::ModuleDeclaration => {
            if ambient(node, file)
                || node
                    .name()
                    .is_none_or(|name| name.kind() != Kind::Identifier)
                || node
                    .body()
                    .is_some_and(|body| body.kind() == Kind::ModuleDeclaration)
            {
                return false;
            }
            let mut current = Some(node);
            while let Some(owner) = current {
                if ast::is_global_scope_augmentation(owner) {
                    return false;
                }
                current = owner.parent();
            }
            !checker
                .get_symbol_at_location_exported(node.name().unwrap())
                .is_some_and(|symbol| {
                    symbol
                        .declarations()
                        .iter()
                        .any(|&declaration| declaration != node)
                })
        }
        Kind::ImportSpecifier
        | Kind::ImportClause
        | Kind::NamespaceImport
        | Kind::ImportEqualsDeclaration => true,
        Kind::TypeParameter => {
            !ambient(node, file)
                && node.parent().is_none_or(|parent| {
                    !matches!(parent.kind(), Kind::MappedType | Kind::JSDocTemplateTag)
                })
        }
        _ => false,
    }
}
fn bodyless_parameter(node: P<Node>) -> bool {
    let Some(owner) = node.parent() else {
        return false;
    };
    match owner.kind() {
        Kind::SetAccessor
        | Kind::FunctionType
        | Kind::ConstructorType
        | Kind::CallSignature
        | Kind::ConstructSignature
        | Kind::IndexSignature
        | Kind::MethodSignature => true,
        Kind::FunctionDeclaration | Kind::MethodDeclaration | Kind::Constructor => {
            owner.body().is_none()
        }
        _ => false,
    }
}
fn signature_parameter(mut node: P<Node>) -> bool {
    loop {
        if node.kind() == Kind::Parameter
            && let Some(owner) = node.parent()
        {
            match owner.kind() {
                Kind::SetAccessor
                | Kind::FunctionType
                | Kind::ConstructorType
                | Kind::CallSignature
                | Kind::ConstructSignature
                | Kind::MethodSignature => return true,
                Kind::FunctionDeclaration
                | Kind::FunctionExpression
                | Kind::MethodDeclaration
                | Kind::Constructor
                    if owner.body().is_none() =>
                {
                    return true;
                }
                _ => {}
            }
        }
        let Some(parent) = node.parent() else {
            return false;
        };
        node = parent;
    }
}
// UnusedVarsVisitor deliberately visits *all* identifiers in a signature
// parameter, including property keys, tuple labels and infer declarations.
// They are not checker references. Reconstruct only the lexical type scopes
// needed for this additional marking, using binder tables for ordinary names.
// Conditional inferences are visible in check/extends/true but not false;
// nested function/mapped scopes delegate infer declarations to that conditional.
struct SignatureScope {
    node: P<Node>,
    parent: Option<usize>,
    function_type: bool,
    declarations: HashMap<String, Vec<P<Node>>>,
}
struct SignatureScopes {
    scopes: Vec<SignatureScope>,
    identifiers: Vec<(P<Node>, usize)>,
}
impl SignatureScopes {
    fn push(&mut self, node: P<Node>, parent: Option<usize>, function_type: bool) -> usize {
        let index = self.scopes.len();
        self.scopes.push(SignatureScope {
            node,
            parent,
            function_type,
            declarations: HashMap::new(),
        });
        index
    }
    fn add(&mut self, scope: usize, declaration: P<Node>) {
        if let Some(name) = declaration.name() {
            self.add_name(scope, name, declaration);
        }
    }
    fn add_name(&mut self, scope: usize, name: P<Node>, declaration: P<Node>) {
        match name.kind() {
            Kind::Identifier if name.text() != "this" => self.scopes[scope]
                .declarations
                .entry(name.text().to_owned())
                .or_default()
                .push(declaration),
            Kind::ObjectBindingPattern | Kind::ArrayBindingPattern => {
                for &element in name.elements() {
                    if element.kind() == Kind::BindingElement {
                        self.add(scope, element);
                    }
                }
            }
            _ => {}
        }
    }
    fn visit(&mut self, node: P<Node>, outer: usize) {
        if node.flags().intersects(NodeFlags::Reparsed) {
            return;
        }
        if node.kind() == Kind::ConditionalType {
            let scope = self.push(node, Some(outer), false);
            let conditional = node.as_conditional_type_node();
            self.visit(conditional.check_type(), scope);
            self.visit(conditional.extends_type(), scope);
            self.visit(conditional.true_type(), scope);
            self.visit(conditional.false_type(), outer);
            return;
        }
        if node.kind() == Kind::InferType {
            let mut target = outer;
            while self.scopes[target].function_type
                || self.scopes[target].node.kind() == Kind::MappedType
            {
                let Some(parent) = self.scopes[target].parent else {
                    break;
                };
                target = parent;
            }
            if self.scopes[target].node.kind() != Kind::ConditionalType {
                target = outer;
            }
            let parameter = node.as_infer_type_node().type_parameter();
            self.add(target, parameter);
            // The declaration identifier is visited in its actual lexical
            // scope, allowing a nearer generic or parameter to shadow it.
            self.visit(parameter, outer);
            return;
        }
        let function_type = matches!(
            node.kind(),
            Kind::FunctionType
                | Kind::ConstructorType
                | Kind::CallSignature
                | Kind::ConstructSignature
                | Kind::MethodSignature
        );
        let runtime_function = matches!(
            node.kind(),
            Kind::FunctionDeclaration
                | Kind::FunctionExpression
                | Kind::ArrowFunction
                | Kind::MethodDeclaration
                | Kind::Constructor
                | Kind::GetAccessor
                | Kind::SetAccessor
        );
        let creates_scope = function_type
            || runtime_function
            || matches!(
                node.kind(),
                Kind::ClassDeclaration
                    | Kind::ClassExpression
                    | Kind::MappedType
                    | Kind::ModuleBlock
                    | Kind::Block
                    | Kind::CatchClause
                    | Kind::ClassStaticBlockDeclaration
            )
            || matches!(
                node.kind(),
                Kind::TypeAliasDeclaration | Kind::InterfaceDeclaration
            ) && !node.type_parameters().is_empty();
        let scope = if creates_scope {
            self.push(node, Some(outer), function_type)
        } else {
            outer
        };
        if creates_scope {
            // Compiler accessors are partial: blocks, catch scopes and mapped
            // types do not have a generic TypeParameterList field.
            if function_type
                || runtime_function
                || matches!(
                    node.kind(),
                    Kind::ClassDeclaration
                        | Kind::ClassExpression
                        | Kind::TypeAliasDeclaration
                        | Kind::InterfaceDeclaration
                )
            {
                for &parameter in node.type_parameters() {
                    self.add(scope, parameter);
                }
            }
            if function_type || runtime_function {
                for &parameter in node.parameters() {
                    self.add(scope, parameter);
                }
            }
            if node.kind() == Kind::MappedType {
                self.add(scope, node.as_mapped_type_node().type_parameter());
            }
            if matches!(
                node.kind(),
                Kind::FunctionExpression | Kind::ClassExpression
            ) {
                self.add(scope, node);
            }
        }
        if node.kind() == Kind::Identifier && signature_parameter(node) {
            self.identifiers.push((node, scope));
        }
        // Computed method names evaluate outside their signature's scope.
        let computed = if function_type || runtime_function {
            node.name()
                .filter(|name| name.kind() == Kind::ComputedPropertyName)
        } else {
            None
        };
        node.for_each_child(&mut |child| {
            self.visit(
                child,
                if computed == Some(child) {
                    outer
                } else {
                    scope
                },
            );
            false
        });
    }
    fn mark(&self) -> HashSet<P<Node>> {
        let mut used = HashSet::new();
        for &(identifier, initial) in &self.identifiers {
            let mut current = Some(initial);
            while let Some(scope) = current {
                let scope = &self.scopes[scope];
                if let Some(declarations) = scope.declarations.get(identifier.text()) {
                    used.extend(
                        declarations
                            .iter()
                            .copied()
                            .filter(|declaration| declaration.kind() == Kind::TypeParameter),
                    );
                    break;
                }
                if let Some(symbol) = scope
                    .node
                    .locals()
                    .and_then(|locals| locals.get().get(identifier.text()))
                {
                    used.extend(
                        symbol
                            .declarations()
                            .iter()
                            .copied()
                            .filter(|declaration| declaration.kind() == Kind::TypeParameter),
                    );
                    break;
                }
                current = scope.parent;
            }
        }
        used
    }
}
fn signature_type_usage(root: P<Node>) -> HashSet<P<Node>> {
    let mut scopes = SignatureScopes {
        scopes: Vec::new(),
        identifiers: Vec::new(),
    };
    let root_scope = scopes.push(root, None, false);
    scopes.visit(root, root_scope);
    scopes.mark()
}
pub(crate) fn directly_exported(mut declaration: P<Node>) -> bool {
    loop {
        match declaration.kind() {
            Kind::FunctionDeclaration
            | Kind::ClassDeclaration
            | Kind::InterfaceDeclaration
            | Kind::TypeAliasDeclaration
            | Kind::EnumDeclaration
            | Kind::ModuleDeclaration
            | Kind::ImportEqualsDeclaration
            | Kind::VariableStatement => {
                return ast::get_combined_modifier_flags(declaration)
                    .intersects(ModifierFlags::Export);
            }
            Kind::VariableDeclaration
            | Kind::VariableDeclarationList
            | Kind::BindingElement
            | Kind::ObjectBindingPattern
            | Kind::ArrayBindingPattern => {
                let Some(parent) = declaration.parent() else {
                    return false;
                };
                declaration = parent;
            }
            _ => return false,
        }
    }
}
pub(crate) fn export_reference(
    node: P<Node>,
    meaning: SymbolFlags,
    checker: &mut Checker,
) -> Option<P<Symbol>> {
    let mut location = node;
    loop {
        let symbol = checker.resolve_name_exported(node.text(), location, meaning, false)?;
        let declarations = symbol.declarations();
        let alias_only = !declarations.is_empty()
            && declarations.iter().all(|declaration| {
                matches!(
                    declaration.kind(),
                    Kind::ExportSpecifier | Kind::NamespaceExport
                )
            });
        if !alias_only
            && !declarations
                .iter()
                .any(|&declaration| Some(declaration) == node.parent())
        {
            return Some(symbol);
        }
        // Namespace exports populate Exports without declaring a local. Retry
        // outside that namespace, retaining the requested type/value meaning.
        let declaration = declarations.first().copied().unwrap_or(node);
        let module = ancestor(declaration, Kind::ModuleDeclaration)?;
        let outer = module.parent()?;
        if outer == location {
            return None;
        }
        location = outer;
    }
}
pub(crate) fn property_name(node: P<Node>) -> bool {
    let Some(parent) = node.parent() else {
        return false;
    };
    match parent.kind() {
        Kind::ShorthandPropertyAssignment | Kind::ExportSpecifier => false,
        Kind::PropertyAccessExpression
        | Kind::PropertyAssignment
        | Kind::JsxAttribute
        | Kind::ImportAttribute
        | Kind::MetaProperty => parent.name() == Some(node),
        Kind::QualifiedName => parent.as_qualified_name().right() == node,
        Kind::BindingElement => parent.as_binding_element().property_name() == Some(node),
        Kind::ImportSpecifier => parent.as_import_specifier().property_name() == Some(node),
        Kind::JsxNamespacedName => ast::is_in_js_file(node) || !ast::is_jsx_tag_name(parent),
        Kind::BreakStatement | Kind::ContinueStatement | Kind::NamespaceExport => true,
        Kind::LabeledStatement => parent.as_labeled_statement().label() == node,
        _ => false,
    }
}
pub(crate) fn import_type_syntax(node: P<Node>) -> bool {
    let Some(parent) = node.parent() else {
        return false;
    };
    if !(parent.kind() >= Kind::FirstTypeNode && parent.kind() <= Kind::LastTypeNode)
        && !matches!(
            parent.kind(),
            Kind::QualifiedName | Kind::ImportAttribute | Kind::ImportAttributes
        )
    {
        return false;
    }
    let mut current = node;
    while let Some(parent) = current.parent() {
        if parent.kind() == Kind::ImportType {
            return parent.as_import_type_node().qualifier() == Some(current)
                || parent.as_import_type_node().argument() == current
                || parent.as_import_type_node().attributes() == Some(current);
        }
        current = parent;
    }
    false
}
pub(crate) fn jsx_non_reference(node: P<Node>) -> bool {
    let Some(parent) = node.parent() else {
        return false;
    };
    if ast::is_in_js_file(node) {
        let mut current = node;
        while let Some(parent) = current.parent() {
            match parent.kind() {
                Kind::PropertyAccessExpression | Kind::JsxNamespacedName => current = parent,
                Kind::JsxClosingElement => return true,
                _ => break,
            }
        }
    }
    if matches!(
        parent.kind(),
        Kind::JsxOpeningElement | Kind::JsxSelfClosingElement | Kind::JsxClosingElement
    ) {
        let name = node.text();
        if name == "this" {
            return ast::is_in_js_file(node);
        }
        let Some(first) = name.chars().next() else {
            return true;
        };
        // JavaScript indexes the first UTF-16 code unit. An astral character's
        // high surrogate has no case mapping; use the compiler's JS casing for
        // BMP code points, including multi-code-point uppercase expansions.
        return (first as u32) <= 0xffff
            && tsrs_core::stringutil::to_upper_js(&name[..first.len_utf8()])
                != name[..first.len_utf8()];
    }
    false
}
fn heritage_root(mut node: P<Node>) -> bool {
    if node.parent().is_none_or(|parent| {
        parent.kind() != Kind::QualifiedName || parent.as_qualified_name().left() != node
    }) {
        return false;
    }
    while let Some(parent) = node.parent() {
        if parent.kind() == Kind::QualifiedName && parent.as_qualified_name().left() == node {
            node = parent;
        } else {
            return parent.kind() == Kind::TypeReference
                && parent
                    .parent()
                    .is_some_and(|owner| owner.kind() == Kind::HeritageClause);
        }
    }
    false
}
pub(crate) fn reference_symbol(node: P<Node>, checker: &mut Checker) -> Option<P<Symbol>> {
    let parent = node.parent()?;
    if parent.kind() == Kind::ShorthandPropertyAssignment {
        return resolve_reference(node, checker);
    }
    if (ast::is_part_of_type_node(node) || heritage_root(node)) && !type_only(node) {
        // Qualified heritage names use a type-capable lexical lookup, as in the
        // frozen rule's withHeritageReferences adjustment.
        return checker.resolve_name_exported(
            node.text(),
            node,
            SymbolFlags::Type | SymbolFlags::Namespace | SymbolFlags::Alias,
            false,
        );
    }
    resolve_reference(node, checker).or_else(|| checker.get_symbol_at_location_exported(node))
}
fn type_only(mut node: P<Node>) -> bool {
    if ast::is_part_of_type_query(node) {
        return true;
    }
    while matches!(node.kind(), Kind::Identifier | Kind::QualifiedName) {
        let Some(parent) = node.parent() else {
            return false;
        };
        node = parent;
    }
    node.kind() == Kind::TypePredicate
}
fn is_import(node: P<Node>) -> bool {
    matches!(
        node.kind(),
        Kind::ImportSpecifier
            | Kind::ImportClause
            | Kind::NamespaceImport
            | Kind::ImportEqualsDeclaration
    )
}
fn type_or_import(node: P<Node>) -> bool {
    is_import(node)
        || matches!(
            node.kind(),
            Kind::InterfaceDeclaration
                | Kind::TypeAliasDeclaration
                | Kind::EnumDeclaration
                | Kind::TypeParameter
        )
}
fn return_first(statement: P<Node>) -> bool {
    statement.kind() == Kind::ReturnStatement
        || statement.kind() == Kind::Block
            && statement
                .statements()
                .first()
                .is_some_and(|node| node.kind() == Kind::ReturnStatement)
}
fn assignment(node: P<Node>) -> Assignment {
    let Some(parent) = node.parent() else {
        return Assignment::None;
    };
    let target = match parent.kind() {
        Kind::BinaryExpression => {
            let binary = parent.as_binary_expression();
            if binary.left() != node || !ast::is_assignment_operator(binary.operator_token().kind())
            {
                return Assignment::None;
            }
            Some(parent)
        }
        Kind::PrefixUnaryExpression => matches!(
            parent.as_prefix_unary_expression().operator(),
            Kind::PlusPlusToken | Kind::MinusMinusToken
        )
        .then_some(parent),
        Kind::PostfixUnaryExpression => Some(parent),
        Kind::ForInStatement | Kind::ForOfStatement => {
            (parent.initializer() == Some(node)).then_some(parent)
        }
        Kind::ParenthesizedExpression
        | Kind::ArrayLiteralExpression
        | Kind::SpreadElement
        | Kind::NonNullExpression
        | Kind::SpreadAssignment
        | Kind::ShorthandPropertyAssignment
        | Kind::PropertyAssignment => ast::get_assignment_target(node),
        _ => None,
    };
    let Some(target) = target else {
        return Assignment::None;
    };
    match target.kind() {
        Kind::BinaryExpression => {
            if target.as_binary_expression().operator_token().kind() == Kind::EqualsToken {
                Assignment::Write
            } else {
                Assignment::ReadWrite
            }
        }
        Kind::PrefixUnaryExpression | Kind::PostfixUnaryExpression => Assignment::ReadWrite,
        Kind::ForInStatement | Kind::ForOfStatement => {
            if target.initializer().map(ast::skip_parentheses) == Some(node)
                && return_first(target.as_for_in_or_of_statement().statement())
            {
                Assignment::None
            } else {
                Assignment::Write
            }
        }
        _ => Assignment::None,
    }
}
fn has_assignment(mut declaration: P<Node>, writes: &[P<Node>]) -> bool {
    if !writes.is_empty() {
        return true;
    }
    if matches!(
        declaration.kind(),
        Kind::VariableDeclaration | Kind::Parameter
    ) {
        return declaration.initializer().is_some();
    }
    if declaration.kind() != Kind::BindingElement {
        return false;
    }
    let definition_only = ast::is_part_of_parameter_declaration(declaration)
        || ast::is_catch_clause_variable_declaration_or_binding_element(declaration);
    loop {
        if matches!(declaration.kind(), Kind::BindingElement | Kind::Parameter)
            && declaration.initializer().is_some()
        {
            return true;
        }
        if declaration.kind() == Kind::VariableDeclaration {
            return !definition_only;
        }
        if matches!(declaration.kind(), Kind::Parameter | Kind::CatchClause) {
            return false;
        }
        let Some(parent) = declaration.parent() else {
            return false;
        };
        declaration = parent;
    }
}
fn enclosing_scope(mut node: P<Node>) -> P<Node> {
    while let Some(parent) = node.parent() {
        if ast::is_function_like_or_class_static_block_declaration(parent)
            || matches!(parent.kind(), Kind::SourceFile | Kind::ModuleBlock)
        {
            return parent;
        }
        node = parent;
    }
    node
}
fn before_last_used(
    parameter: P<Node>,
    bindings: &[Binding],
    references: &HashMap<P<Symbol>, References>,
    used_property_parameters: &HashSet<P<Node>>,
) -> bool {
    let Some(owner) = parameter.parent() else {
        return false;
    };
    owner
        .parameters()
        .iter()
        .rev()
        .take_while(|&&later| later != parameter)
        .any(|&later| {
            later.initializer().is_some()
                || used_property_parameters.contains(&later)
                || bindings.iter().any(|binding| {
                    ast::get_root_declaration(binding.declaration) == later
                        && binding
                            .symbol
                            .and_then(|symbol| references.get(&symbol))
                            .is_some_and(|info| {
                                info.reads
                                    .iter()
                                    .any(|&reference| reference != binding.name)
                            })
                })
        })
}
fn inside_own_declaration(usage: P<Node>, declaration: P<Node>, declarations: &[P<Node>]) -> bool {
    let own = |declaration: P<Node>| {
        let body = match declaration.kind() {
            Kind::ModuleDeclaration | Kind::FunctionDeclaration => declaration.body(),
            Kind::InterfaceDeclaration | Kind::TypeAliasDeclaration | Kind::EnumDeclaration => {
                Some(declaration)
            }
            Kind::VariableDeclaration => declaration
                .initializer()
                .and_then(|initializer| initializer.body()),
            _ => None,
        };
        body.is_some_and(|body| contains(body, usage))
    };
    own(declaration) || declarations.iter().copied().any(own)
}
fn unused_expression(mut node: P<Node>) -> bool {
    while let Some(parent) = node.parent() {
        match parent.kind() {
            Kind::ExpressionStatement => return true,
            Kind::ParenthesizedExpression => {}
            Kind::BinaryExpression
                if parent.as_binary_expression().operator_token().kind() == Kind::CommaToken =>
            {
                if parent.as_binary_expression().left() == node {
                    return true;
                }
            }
            _ => return false,
        }
        node = parent;
    }
    false
}
fn inside_loop(mut node: P<Node>) -> bool {
    loop {
        if ast::is_function_like(node) {
            return false;
        }
        if matches!(
            node.kind(),
            Kind::ForStatement
                | Kind::ForInStatement
                | Kind::ForOfStatement
                | Kind::WhileStatement
                | Kind::DoStatement
        ) {
            return true;
        }
        let Some(parent) = node.parent() else {
            return false;
        };
        node = parent;
    }
}
fn same_reference(node: P<Node>, symbol: P<Symbol>, checker: &mut Checker) -> bool {
    checker
        .get_symbol_at_location_exported(node)
        .is_some_and(|resolved| checker.get_merged_symbol(resolved) == symbol)
}
fn inside_storable_function(mut node: P<Node>, boundary: P<Node>) -> bool {
    while let Some(parent) = node.parent() {
        if parent == boundary {
            return false;
        }
        if ast::is_function_like(parent)
            && ast::get_immediately_invoked_function_expression(parent).is_none()
        {
            return true;
        }
        node = parent;
    }
    false
}
fn inside_function_assigned_to_self(
    mut node: P<Node>,
    symbol: P<Symbol>,
    checker: &mut Checker,
) -> bool {
    loop {
        if matches!(node.kind(), Kind::FunctionExpression | Kind::ArrowFunction) {
            let mut current = node;
            while let Some(parent) = current.parent() {
                match parent.kind() {
                    Kind::ParenthesizedExpression => {}
                    Kind::BinaryExpression => {
                        let binary = parent.as_binary_expression();
                        if binary.operator_token().kind() != Kind::CommaToken {
                            if ast::is_assignment_operator(binary.operator_token().kind())
                                && same_reference(binary.left(), symbol, checker)
                            {
                                return unused_expression(parent);
                            }
                            break;
                        }
                    }
                    Kind::CallExpression
                        if ast::skip_parentheses(parent.as_call_expression().expression())
                            == node => {}
                    _ => break,
                }
                current = parent;
            }
        }
        let Some(parent) = node.parent() else {
            return false;
        };
        node = parent;
    }
}
fn self_modification(
    node: P<Node>,
    symbol: P<Symbol>,
    declaration: P<Node>,
    checker: &mut Checker,
) -> bool {
    if let Some(target) = ast::get_assignment_target(node) {
        match target.kind() {
            Kind::PrefixUnaryExpression | Kind::PostfixUnaryExpression => {
                return unused_expression(target);
            }
            Kind::BinaryExpression => {
                let operator = target.as_binary_expression().operator_token().kind();
                if ast::is_compound_assignment(operator) {
                    if matches!(
                        operator,
                        Kind::BarBarEqualsToken
                            | Kind::AmpersandAmpersandEqualsToken
                            | Kind::QuestionQuestionEqualsToken
                    ) {
                        return false;
                    }
                    return unused_expression(target);
                }
            }
            _ => {}
        }
    }
    if inside_function_assigned_to_self(node, symbol, checker) {
        return true;
    }
    let mut current = node;
    while let Some(parent) = current.parent() {
        match parent.kind() {
            Kind::BinaryExpression => {
                let binary = parent.as_binary_expression();
                if ast::is_assignment_operator(binary.operator_token().kind()) {
                    if same_reference(binary.left(), symbol, checker) {
                        if ast::get_containing_function(parent)
                            != ast::get_containing_function(declaration)
                            || inside_loop(parent)
                        {
                            return false;
                        }
                        return unused_expression(parent);
                    }
                    break;
                }
            }
            Kind::ParenthesizedExpression => {}
            Kind::PropertyAccessExpression
                if parent.as_property_access_expression().expression() == current => {}
            Kind::CallExpression => {
                let callee = parent.as_call_expression().expression();
                let same_method = callee.kind() == Kind::PropertyAccessExpression
                    && same_reference(
                        callee.as_property_access_expression().expression(),
                        symbol,
                        checker,
                    );
                if callee != current && !same_method && inside_storable_function(node, parent) {
                    break;
                }
            }
            _ => break,
        }
        current = parent;
    }
    false
}
fn pos(node: P<Node>) -> usize {
    node.pos().expect_source("authored unused binding position") as usize
}
fn end(node: P<Node>) -> usize {
    node.end().expect_source("authored unused binding end") as usize
}
fn comma_after(text: &[u8], mut index: usize) -> Option<usize> {
    while let Some(&byte) = text.get(index) {
        if byte == b',' {
            return Some(index + 1);
        }
        if !matches!(byte, b' ' | b'\t') {
            break;
        }
        index += 1;
    }
    None
}
fn remove_line(node: P<Node>, file: P<SourceFile>) -> Edit {
    let text = file.text().as_bytes();
    let mut end = end(node);
    if text.get(end) == Some(&b'\n') {
        end += 1;
    } else if text.get(end..end + 2) == Some(b"\r\n") {
        end += 2;
    }
    Edit::replace([pos(node), end], "")
}
fn all_imports_unused(clause: P<Node>, unused: &HashSet<P<Node>>) -> bool {
    if clause.name().is_some() && !unused.contains(&clause) {
        return false;
    }
    match clause.as_import_clause().named_bindings() {
        Some(bindings) if bindings.kind() == Kind::NamespaceImport => unused.contains(&bindings),
        Some(bindings) => bindings
            .as_named_imports()
            .elements()
            .nodes()
            .iter()
            .all(|node| unused.contains(node)),
        None => true,
    }
}
fn import_suggestion(
    node: P<Node>,
    file: P<SourceFile>,
    unused: &HashSet<P<Node>>,
) -> Option<(&'static str, String, Edit)> {
    let whole = |node| {
        (
            "removeUnusedImportDeclaration",
            "Remove unused import declaration.".to_string(),
            remove_line(node, file),
        )
    };
    let single = |edit| {
        (
            "removeUnusedVar",
            format!(
                "Remove unused variable \"{}\".",
                node.name().unwrap().text()
            ),
            edit,
        )
    };
    if node.kind() == Kind::ImportEqualsDeclaration {
        return Some(whole(node));
    }
    let import = ancestor(node, Kind::ImportDeclaration)?;
    if node.kind() == Kind::NamespaceImport {
        return Some(whole(import));
    }
    let clause = import.as_import_declaration().import_clause()?;
    if all_imports_unused(clause, unused) {
        return Some(whole(import));
    }
    let text = file.text().as_bytes();
    if node.kind() == Kind::ImportClause {
        let name = node.name()?;
        return comma_after(text, end(name)).map(|end| single(Edit::replace([pos(name), end], "")));
    }
    let start = pos(node);
    let mut text_start = start;
    while text_start < end(node) && matches!(text[text_start], b' ' | b'\t' | b'\n' | b'\r') {
        text_start += 1;
    }
    let mut preceding = start;
    while preceding > 0 {
        preceding -= 1;
        if text[preceding] == b',' {
            return Some(single(Edit::replace([preceding, end(node)], "")));
        }
        if !matches!(text[preceding], b' ' | b'\t' | b'\n' | b'\r') {
            break;
        }
    }
    let mut finish = end(node);
    if let Some(after) = comma_after(text, finish) {
        finish = after;
        while matches!(text.get(finish), Some(b' ' | b'\t')) {
            finish += 1;
        }
    }
    Some(single(Edit::replace([text_start, finish], "")))
}
