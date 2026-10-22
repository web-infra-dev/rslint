//! Import graph rules from frozen rslint e0a29f751d4def86d27931759098f27748fa3003.
//! These syntax rules inspect the selected tsrs Program, without requiring a
//! type-aware rule gate. Authored export maps deliberately differ from checker
//! exports: synthetic defaults, missing re-exports and namespace aliases retain
//! the eslint-plugin-import compatibility decisions of the frozen implementation.
use crate::Violation;
use std::collections::{HashMap, HashSet};
use std::path::Path;
use tsrs_ast::{self as ast, Kind, ModifierFlags, Node, NodeFlags, SourceFile};
use tsrs_compiler::Program;
use tsrs_core::{Context, ModuleKind, ModuleResolutionKind, P, Tristate};

pub(crate) const NAMES: &[&str] = &[
    "import/default",
    "import/named",
    "import/namespace",
    "import/no-named-as-default",
];

pub(crate) fn check(
    name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    program: &Program,
) -> Vec<Violation> {
    if node.kind() != Kind::SourceFile {
        return vec![];
    }
    let mut graph = Graph::new(program);
    match name {
        "import/default" => default_imports(file, &mut graph),
        "import/named" => named_imports(file, &mut graph),
        "import/namespace" => namespaces(file, &mut graph),
        "import/no-named-as-default" => named_as_default(file, &mut graph),
        _ => vec![],
    }
}

#[derive(Clone, Copy, PartialEq)]
enum View {
    Authored,
    DefaultOnly,
}
#[derive(Clone)]
struct Link {
    file: P<SourceFile>,
    path: String,
    view: View,
    node_default: bool,
}
#[derive(Clone, Default)]
struct Meta {
    namespace: Option<usize>,
    unresolved: bool,
}
#[derive(Default)]
struct ExportMap {
    exports: HashMap<String, Meta>,
    reexports: HashMap<String, String>,
    unknown: bool,
    implicit_default: bool,
}
impl ExportMap {
    fn has(&self, name: &str) -> bool {
        self.unknown || self.exports.contains_key(name)
    }
    fn has_default(&self) -> bool {
        self.implicit_default || self.exports.get("default").is_some_and(|m| !m.unresolved)
    }
    fn empty(&self) -> bool {
        self.exports.is_empty() && !self.unknown
    }
}
#[derive(Clone)]
enum Step {
    Names(Vec<String>),
    Default(String),
    Star(Option<Link>),
    Named {
        from: bool,
        link: Option<Link>,
        names: Vec<(String, String, String)>,
    },
}
#[derive(Clone, Default)]
struct Local {
    steps: Vec<Step>,
    imports: Vec<(String, Option<Link>)>,
    implicit_default: bool,
}
struct Graph<'p> {
    program: &'p Program,
    locals: HashMap<P<SourceFile>, Local>,
    building: HashMap<P<SourceFile>, usize>,
    maps: Vec<ExportMap>,
}
impl<'p> Graph<'p> {
    fn new(program: &'p Program) -> Self {
        Self {
            program,
            locals: HashMap::new(),
            building: HashMap::new(),
            maps: vec![],
        }
    }
    fn link(&self, origin: P<SourceFile>, source: P<Node>, require_module: bool) -> Option<Link> {
        if !ast::is_string_literal_like(source) {
            return None;
        }
        let view = module_view(source)?;
        let resolved = self
            .program
            .get_resolved_module_from_module_specifier(origin, source);
        let resolved_path = resolved
            .filter(|module| module.is_resolved())
            .map(|module| module.resolved_file_name.to_owned());
        let (path, file) = if let Some(path) = resolved_path {
            let file = self
                .program
                .get_source_file_for_resolved_module(&path)
                .or_else(|| (view == View::DefaultOnly).then_some(origin))?;
            (path, file)
        } else {
            let file = relative_loaded_file(self.program, origin, source.text())?;
            (file.file_name().to_owned(), file)
        };
        if require_module && view == View::Authored && !ast::is_external_module(file) {
            return None;
        }
        let module = self.program.options().get_emit_module_kind();
        let node_default = (ModuleKind::Node16..=ModuleKind::NodeNext).contains(&module)
            && source.parent().is_some_and(|p| {
                matches!(p.kind(), Kind::ImportDeclaration | Kind::ExportDeclaration)
            })
            && self.program.get_mode_for_usage_location(origin, source) == ModuleKind::ESNext
            && self
                .program
                .get_source_file_meta_data(file.path())
                .implied_node_format
                == ModuleKind::CommonJS;
        Some(Link {
            file,
            path,
            view,
            node_default,
        })
    }
    fn interop(&self, explicit: bool) -> bool {
        let options = self.program.options();
        if explicit || options.es_module_interop != Tristate::Unknown {
            return options.es_module_interop == Tristate::True;
        }
        matches!(
            options.module,
            ModuleKind::Node16 | ModuleKind::NodeNext | ModuleKind::Preserve
        )
    }
    fn synthetic(&self, explicit: bool) -> bool {
        let options = self.program.options();
        if options.allow_synthetic_default_imports != Tristate::Unknown {
            return options.allow_synthetic_default_imports == Tristate::True;
        }
        self.interop(explicit)
            || options.module == ModuleKind::System
            || options.module_resolution == ModuleResolutionKind::Bundler
            || options.module_resolution == ModuleResolutionKind::Unknown
                && options.module == ModuleKind::Preserve
    }
    fn commonjs_declaration(&self, file: P<SourceFile>) -> bool {
        file.is_declaration_file()
            && ast::get_implied_node_format_for_emit_worker(
                file.file_name(),
                self.program.options().get_emit_module_kind(),
                self.program.get_source_file_meta_data(file.path()),
            ) != ModuleKind::ESNext
    }
    fn assignment_default(&self, file: P<SourceFile>, stmt: P<Node>, explicit: bool) -> bool {
        let assignment = stmt.as_export_assignment();
        if !assignment.is_export_equals {
            return true;
        }
        let expression = ast::skip_parentheses(assignment.expression.get());
        let name = match expression.kind() {
            Kind::Identifier => Some(expression.text()),
            Kind::FunctionExpression | Kind::ClassExpression => expression.name().map(|n| n.text()),
            _ => None,
        };
        let Some(name) = name else {
            return true;
        };
        for &stmt in file.statements().nodes() {
            let declares = if stmt.kind() == Kind::VariableStatement {
                declaration_names(stmt).iter().any(|n| n == name)
            } else {
                matches!(
                    stmt.kind(),
                    Kind::ClassDeclaration
                        | Kind::InterfaceDeclaration
                        | Kind::TypeAliasDeclaration
                        | Kind::EnumDeclaration
                        | Kind::ModuleDeclaration
                ) || stmt.kind() == Kind::FunctionDeclaration
                    && ast::has_syntactic_modifier(stmt, ModifierFlags::Ambient)
            };
            if declares
                && (stmt.kind() == Kind::VariableStatement
                    || stmt.name().is_some_and(|n| n.text() == name))
            {
                return stmt.kind() != Kind::ModuleDeclaration || self.synthetic(explicit);
            }
        }
        true
    }
    fn local(&mut self, file: P<SourceFile>) -> Local {
        if let Some(local) = self.locals.get(&file) {
            return local.clone();
        }
        let mut local = Local::default();
        for &stmt in file.statements().nodes() {
            if stmt.kind() == Kind::ImportDeclaration {
                let import = stmt.as_import_declaration();
                if let Some(clause) = import.import_clause {
                    let name = clause
                        .as_import_clause()
                        .named_bindings()
                        .filter(|n| n.kind() == Kind::NamespaceImport)
                        .and_then(|n| n.name())
                        .map_or("", |n| n.text());
                    local.imports.push((
                        name.to_owned(),
                        self.link(file, import.module_specifier, true),
                    ));
                }
            }
            if exported_declaration(stmt) {
                local.steps.push(Step::Names(declaration_names(stmt)));
                continue;
            }
            match stmt.kind() {
                Kind::ExportAssignment => {
                    let assignment = stmt.as_export_assignment();
                    if assignment.is_export_equals {
                        local.implicit_default |= self.assignment_default(file, stmt, false);
                        append_namespace(&mut local, file, assignment.expression.get());
                    } else {
                        local.steps.push(Step::Default(
                            identifier(assignment.expression.get()).to_owned(),
                        ));
                    }
                }
                Kind::NamespaceExportDeclaration if self.interop(false) => {
                    if let Some(name) = stmt.name() {
                        append_namespace(&mut local, file, name);
                    }
                }
                Kind::ExportDeclaration => {
                    let declaration = stmt.as_export_declaration();
                    let link = declaration
                        .module_specifier
                        .and_then(|source| self.link(file, source, true));
                    match declaration.export_clause {
                        None if declaration.module_specifier.is_some() => {
                            local.steps.push(Step::Star(link))
                        }
                        Some(clause) if clause.kind() == Kind::NamespaceExport => {
                            if let Some(name) = clause.name() {
                                local.steps.push(Step::Names(vec![name.text().to_owned()]));
                            }
                        }
                        Some(clause) if clause.kind() == Kind::NamedExports => {
                            let names = clause
                                .elements()
                                .iter()
                                .filter_map(|specifier| {
                                    let exported = specifier.name()?;
                                    let source = specifier.property_name_or_name()?;
                                    Some((
                                        exported.text().to_owned(),
                                        source.text().to_owned(),
                                        identifier(source).to_owned(),
                                    ))
                                })
                                .collect();
                            local.steps.push(Step::Named {
                                from: declaration.module_specifier.is_some(),
                                link,
                                names,
                            });
                        }
                        _ => {}
                    }
                }
                _ => {}
            }
        }
        if !local.implicit_default && self.commonjs_declaration(file) && self.synthetic(false) {
            local.implicit_default = !local.steps.iter().any(|step| match step {
                Step::Default(_) => true,
                Step::Names(names) => names
                    .iter()
                    .any(|s| matches!(s.as_str(), "default" | "__esModule")),
                Step::Named { names, .. } => names
                    .iter()
                    .any(|(s, _, _)| matches!(s.as_str(), "default" | "__esModule")),
                _ => false,
            });
        }
        self.locals.insert(file, local.clone());
        local
    }
    fn map_for_link(&mut self, link: &Link) -> usize {
        if link.view == View::DefaultOnly {
            let id = self.maps.len();
            let mut map = ExportMap::default();
            map.exports.insert("default".into(), Meta::default());
            self.maps.push(map);
            return id;
        }
        self.map(link.file)
    }
    fn namespace_meta(&mut self, local: &Local, name: &str) -> Option<Meta> {
        if name.is_empty() {
            return None;
        }
        for (binding, link) in &local.imports {
            if let Some(link) = link {
                let namespace = self.map_for_link(link);
                if binding == name {
                    return Some(Meta {
                        namespace: Some(namespace),
                        unresolved: false,
                    });
                }
            }
        }
        None
    }
    fn map(&mut self, file: P<SourceFile>) -> usize {
        if let Some(&id) = self.building.get(&file) {
            return id;
        }
        let id = self.maps.len();
        self.maps.push(ExportMap::default());
        self.building.insert(file, id);
        let local = self.local(file);
        self.maps[id].implicit_default = local.implicit_default;
        for step in &local.steps {
            match step {
                Step::Names(names) => {
                    for name in names {
                        self.maps[id].exports.insert(name.clone(), Meta::default());
                    }
                }
                Step::Default(name) => {
                    let meta = self.namespace_meta(&local, name).unwrap_or_default();
                    self.maps[id].exports.insert("default".into(), meta);
                }
                Step::Star(None) => self.maps[id].unknown = true,
                Step::Star(Some(link)) => {
                    let other = self.map_for_link(link);
                    let entries = self.maps[other].exports.clone();
                    self.maps[id].unknown |= self.maps[other].unknown;
                    for (name, meta) in entries {
                        if name != "default" {
                            self.maps[id].exports.insert(name, meta);
                        }
                    }
                }
                Step::Named { from, link, names } => {
                    let dependency = if *from {
                        link.as_ref().map(|link| self.map_for_link(link))
                    } else {
                        None
                    };
                    for (exported, source, identifier) in names {
                        let meta = if !*from {
                            self.namespace_meta(&local, identifier).unwrap_or_default()
                        } else if let Some(dependency) = dependency {
                            let link = link.as_ref().expect("resolved dependency");
                            self.maps[id]
                                .reexports
                                .insert(exported.clone(), link.path.clone());
                            let meta = self.maps[dependency].exports.get(source).cloned();
                            if source == "default"
                                && (self.maps[dependency].implicit_default || link.node_default)
                                && meta.as_ref().is_none_or(|meta| meta.unresolved)
                            {
                                Meta::default()
                            } else {
                                meta.unwrap_or(Meta {
                                    namespace: None,
                                    unresolved: true,
                                })
                            }
                        } else {
                            Meta {
                                namespace: None,
                                unresolved: true,
                            }
                        };
                        self.maps[id].exports.insert(exported.clone(), meta);
                    }
                }
            }
        }
        id
    }
    fn find(
        &mut self,
        link: &Link,
        name: &str,
        seen: &mut HashSet<(P<SourceFile>, String)>,
    ) -> (bool, Vec<String>) {
        let path = vec![link.path.clone()];
        if link.view == View::DefaultOnly {
            return (name == "default", path);
        }
        let file = link.file;
        if !file
            .external_module_indicator()
            .is_some_and(ast::is_external_module_indicator)
            || !self
                .program
                .get_syntactic_diagnostics(&Context::background(), Some(file))
                .is_empty()
        {
            return (true, vec![]);
        }
        if name == "default" && link.node_default {
            return (true, path);
        }
        if !seen.insert((file, name.to_owned())) {
            return (false, path);
        }
        let local = self.local(file);
        if name == "default" && local.implicit_default {
            return (true, path);
        }
        let mut reexport = None;
        for step in &local.steps {
            match step {
                Step::Names(names) if names.iter().any(|n| n == name) => return (true, path),
                Step::Default(_) if name == "default" => return (true, path),
                Step::Named { from, link, names } => {
                    for (exported, source, _) in names {
                        if exported == name {
                            if !*from {
                                return (true, path);
                            }
                            reexport = Some((link, source));
                        }
                    }
                }
                _ => {}
            }
        }
        if let Some((link, source)) = reexport {
            let Some(link) = link else {
                return (true, path);
            };
            if link.file == file && source == name {
                return (false, path);
            }
            let (found, dependency) = self.find(link, source, seen);
            return (found, path.into_iter().chain(dependency).collect());
        }
        if name != "default" {
            for step in &local.steps {
                if let Step::Star(link) = step {
                    let Some(link) = link else {
                        return (true, path);
                    };
                    let (found, dependency) = self.find(link, name, seen);
                    if found {
                        return (true, path.into_iter().chain(dependency).collect());
                    }
                }
            }
        }
        (false, path)
    }
    fn has_export(
        &mut self,
        origin: P<SourceFile>,
        source: P<Node>,
        name: &str,
        default_import: bool,
        seen: &mut HashSet<(P<SourceFile>, String)>,
    ) -> Option<bool> {
        let link = self.link(origin, source, false)?;
        if link.view == View::DefaultOnly {
            return Some(name == "default");
        }
        let file = link.file;
        if name == "default" && link.node_default && !default_import {
            return Some(true);
        }
        if default_import {
            if !unambiguous(file.text())
                || !module_declarations(file) && !runtime_import(file.as_node())
            {
                return None;
            }
            let interop = self.interop(true);
            let synthetic = self.synthetic(true);
            let may_fallback = interop || synthetic && file.is_declaration_file();
            let opted_out = !synthetic
                && (self.commonjs_declaration(file)
                    || file.statements().nodes().iter().any(|n| {
                        n.kind() == Kind::ExportAssignment
                            && n.as_export_assignment().is_export_equals
                    }));
            // Interop supplies only a default value. A re-export lookup for
            // an absent named binding must remain absent even when the target
            // has other local exports and an interop default.
            if name == "default" && may_fallback && !opted_out {
                let local = self.local(file);
                if synthetic && file.is_declaration_file() && local.implicit_default {
                    return Some(true);
                }
                if interop
                    && local.steps.iter().any(|s| match s {
                        Step::Default(_) => true,
                        Step::Names(n) => !n.is_empty(),
                        Step::Named {
                            from: false, names, ..
                        } => !names.is_empty(),
                        _ => false,
                    })
                {
                    return Some(true);
                }
            }
        }
        if !file
            .external_module_indicator()
            .is_some_and(ast::is_external_module_indicator)
            && !runtime_import(file.as_node())
        {
            return None;
        }
        if !default_import
            && name == "default"
            && file.is_declaration_file()
            && self.local(file).implicit_default
        {
            return Some(true);
        }
        let key = (file, name.to_owned());
        if !seen.insert(key.clone()) {
            return Some(false);
        }
        let result = self.has_export_statements(file, name, default_import, seen);
        seen.remove(&key);
        Some(result)
    }
    fn has_export_statements(
        &mut self,
        file: P<SourceFile>,
        name: &str,
        default_import: bool,
        seen: &mut HashSet<(P<SourceFile>, String)>,
    ) -> bool {
        for &stmt in file.statements().nodes() {
            if exported_declaration(stmt) && declaration_names(stmt).iter().any(|s| s == name) {
                return true;
            }
            match stmt.kind() {
                Kind::ExportAssignment
                    if name == "default" && self.assignment_default(file, stmt, default_import) =>
                {
                    return true;
                }
                Kind::ExportDeclaration => {
                    let decl = stmt.as_export_declaration();
                    match decl.export_clause {
                        None if name != "default" => {
                            if let Some(source) = decl.module_specifier
                                && self.has_export(file, source, name, default_import, seen)
                                    != Some(false)
                            {
                                return true;
                            }
                        }
                        Some(clause) if clause.kind() == Kind::NamespaceExport => {
                            if clause.name().is_some_and(|n| n.text() == name) {
                                return true;
                            }
                        }
                        Some(clause) if clause.kind() == Kind::NamedExports => {
                            for &specifier in clause.elements() {
                                if specifier.name().is_some_and(|n| n.text() == name) {
                                    let Some(source) = decl.module_specifier else {
                                        return true;
                                    };
                                    let imported =
                                        specifier.property_name_or_name().expect("export source");
                                    if self.has_export(
                                        file,
                                        source,
                                        imported.text(),
                                        default_import,
                                        seen,
                                    ) != Some(false)
                                    {
                                        return true;
                                    }
                                }
                            }
                        }
                        _ => {}
                    }
                }
                _ => {}
            }
        }
        false
    }
}

fn module_view(source: P<Node>) -> Option<View> {
    let attributes = source.parent().and_then(|parent| match parent.kind() {
        Kind::ImportDeclaration | Kind::JSImportDeclaration => {
            parent.as_import_declaration().attributes()
        }
        Kind::ExportDeclaration => parent.as_export_declaration().attributes,
        _ => None,
    });
    let Some(attributes) = attributes else {
        return Some(View::Authored);
    };
    if attributes.as_import_attributes().token != Kind::WithKeyword {
        return None;
    }
    let mut entries = HashMap::new();
    for &attribute in attributes.as_import_attributes().attributes.nodes() {
        let key = attribute.name()?;
        let value = attribute.as_import_attribute().value;
        if !matches!(key.kind(), Kind::Identifier | Kind::StringLiteral)
            || !ast::is_string_literal_like(value)
            || entries.insert(key.text(), value.text()).is_some()
        {
            return None;
        }
    }
    if entries.is_empty() || entries.keys().all(|&key| key == "resolution-mode") {
        return Some(View::Authored);
    }
    if matches!(entries.get("type"), Some(&"json" | &"text"))
        && entries
            .keys()
            .all(|&key| matches!(key, "type" | "resolution-mode"))
    {
        Some(View::DefaultOnly)
    } else {
        None
    }
}
fn identifier(node: P<Node>) -> &'static str {
    let node = ast::skip_parentheses(node);
    if node.kind() == Kind::Identifier {
        node.text()
    } else {
        ""
    }
}
fn walk(node: P<Node>, visit: &mut impl FnMut(P<Node>)) {
    if node
        .flags()
        .intersects(NodeFlags::JSDoc | NodeFlags::Reparsed)
        || ast::is_jsdoc_node(node)
    {
        return;
    }
    visit(node);
    node.for_each_child(&mut |child| {
        walk(child, visit);
        false
    });
}
fn runtime_import(node: P<Node>) -> bool {
    let mut found = false;
    walk(node, &mut |n| {
        if n.kind() == Kind::CallExpression
            && n.as_call_expression().expression().kind() == Kind::ImportKeyword
        {
            found = true;
        }
    });
    found
}
fn module_declarations(file: P<SourceFile>) -> bool {
    file.statements().nodes().iter().any(|&n| {
        matches!(
            n.kind(),
            Kind::ImportDeclaration | Kind::ExportDeclaration | Kind::ExportAssignment
        ) || ast::has_syntactic_modifier(n, ModifierFlags::Export)
    })
}
fn js_space(c: char) -> bool {
    matches!(
        c,
        '\t' | '\u{b}'
            | '\u{c}'
            | ' '
            | '\u{a0}'
            | '\u{feff}'
            | '\n'
            | '\r'
            | '\u{2028}'
            | '\u{2029}'
            | '\u{1680}'
            | '\u{2000}'..='\u{200a}' | '\u{202f}' | '\u{205f}' | '\u{3000}'
    )
}
fn unambiguous(text: &str) -> bool {
    if text.contains("import(") {
        return true;
    }
    for (offset, _) in text.char_indices() {
        if offset != 0
            && !text[..offset].chars().next_back().is_some_and(|c| {
                matches!(c, ';' | '}' | ')' | '\n' | '\r' | '\u{2028}' | '\u{2029}')
            })
        {
            continue;
        }
        let rest = text[offset..].trim_start_matches(js_space);
        let Some(tail) = rest
            .strip_prefix("export")
            .or_else(|| rest.strip_prefix("import"))
        else {
            continue;
        };
        let trimmed = tail.trim_start_matches(js_space);
        if trimmed.starts_with(['{', '*', '='])
            || trimmed.len() != tail.len()
                && trimmed
                    .chars()
                    .next()
                    .is_some_and(|c| c.is_ascii_alphanumeric() || c == '_')
        {
            return true;
        }
    }
    false
}
fn exported_declaration(node: P<Node>) -> bool {
    matches!(
        node.kind(),
        Kind::VariableStatement
            | Kind::FunctionDeclaration
            | Kind::ClassDeclaration
            | Kind::InterfaceDeclaration
            | Kind::TypeAliasDeclaration
            | Kind::EnumDeclaration
            | Kind::ModuleDeclaration
    ) && ast::has_syntactic_modifier(node, ModifierFlags::Export)
}
fn binding_names(node: P<Node>, names: &mut Vec<String>) {
    match node.kind() {
        Kind::Identifier => names.push(node.text().to_owned()),
        Kind::ObjectBindingPattern | Kind::ArrayBindingPattern => {
            for &element in node.elements() {
                if element.kind() == Kind::BindingElement
                    && let Some(name) = element.name()
                {
                    binding_names(name, names);
                }
            }
        }
        _ => {}
    }
}
fn declaration_names(node: P<Node>) -> Vec<String> {
    if ast::has_syntactic_modifier(node, ModifierFlags::Default) {
        return vec!["default".into()];
    }
    match node.kind() {
        Kind::VariableStatement => {
            let mut names = vec![];
            for &decl in node
                .as_variable_statement()
                .declaration_list
                .as_variable_declaration_list()
                .declarations
                .nodes()
            {
                if let Some(name) = decl.name() {
                    binding_names(name, &mut names);
                }
            }
            names
        }
        Kind::FunctionDeclaration
        | Kind::ClassDeclaration
        | Kind::InterfaceDeclaration
        | Kind::TypeAliasDeclaration
        | Kind::EnumDeclaration
        | Kind::ModuleDeclaration => node
            .name()
            .map(|n| vec![n.text().to_owned()])
            .unwrap_or_default(),
        _ => vec![],
    }
}
fn append_namespace(local: &mut Local, file: P<SourceFile>, expression: P<Node>) {
    let name = identifier(expression);
    if name.is_empty() {
        return;
    }
    for &stmt in file.statements().nodes() {
        if stmt.kind() != Kind::ModuleDeclaration
            || stmt.name().is_none_or(|n| n.text() != name)
            || ast::has_syntactic_modifier(stmt, ModifierFlags::Export)
        {
            continue;
        }
        let Some(body) = stmt.body().filter(|b| b.kind() == Kind::ModuleBlock) else {
            continue;
        };
        let mut names = vec![];
        for &member in body.statements() {
            if member.kind() == Kind::ImportEqualsDeclaration {
                if let Some(n) = member.name() {
                    names.push(n.text().to_owned());
                }
            } else {
                names.extend(declaration_names(member));
            }
        }
        if !names.is_empty() {
            local.steps.push(Step::Names(names));
        }
    }
}
fn default_imports(file: P<SourceFile>, graph: &mut Graph) -> Vec<Violation> {
    let mut result = vec![];
    walk(file.as_node(), &mut |node| {
        if !ast::is_import_declaration_or_js_import_declaration(node) {
            return;
        }
        let decl = node.as_import_declaration();
        let Some(name) = decl.import_clause.and_then(|n| n.name()) else {
            return;
        };
        if graph.has_export(
            file,
            decl.module_specifier,
            "default",
            true,
            &mut HashSet::new(),
        ) == Some(false)
        {
            result.push(Violation::new(
                name,
                "noDefault",
                format!(
                    "No default export found in imported module {}.",
                    super::go_quote::quote(decl.module_specifier.text())
                ),
            ));
        }
    });
    result
}
fn relative(from: &str, to: &str) -> String {
    let base = Path::new(from)
        .parent()
        .expect("source parent")
        .components()
        .collect::<Vec<_>>();
    let target = Path::new(to).components().collect::<Vec<_>>();
    let common = base.iter().zip(&target).take_while(|(a, b)| a == b).count();
    let mut parts = vec!["..".to_owned(); base.len() - common];
    parts.extend(
        target[common..]
            .iter()
            .map(|p| p.as_os_str().to_string_lossy().into_owned()),
    );
    if parts.is_empty() {
        ".".to_owned()
    } else {
        parts.join(std::path::MAIN_SEPARATOR_STR)
    }
}
fn named_imports(file: P<SourceFile>, graph: &mut Graph) -> Vec<Violation> {
    let mut result = vec![];
    walk(file.as_node(), &mut |node| {
        let (source, specifiers, imports) =
            if ast::is_import_declaration_or_js_import_declaration(node) {
                let decl = node.as_import_declaration();
                let Some(clause) = decl.import_clause else {
                    return;
                };
                if clause.is_type_only() {
                    return;
                }
                let Some(bindings) = clause
                    .as_import_clause()
                    .named_bindings()
                    .filter(|n| n.kind() == Kind::NamedImports)
                else {
                    return;
                };
                (decl.module_specifier, bindings.elements(), true)
            } else if node.kind() == Kind::ExportDeclaration {
                let decl = node.as_export_declaration();
                if decl.is_type_only {
                    return;
                }
                let (Some(source), Some(bindings)) = (
                    decl.module_specifier,
                    decl.export_clause
                        .filter(|n| n.kind() == Kind::NamedExports),
                ) else {
                    return;
                };
                (source, bindings.elements(), false)
            } else {
                return;
            };
        for &specifier in specifiers {
            if imports && specifier.is_type_only() {
                continue;
            }
            let Some(name) = specifier.property_name_or_name() else {
                continue;
            };
            let Some(link) = graph.link(file, source, true) else {
                continue;
            };
            let (found, path) = graph.find(&link, name.text(), &mut HashSet::new());
            if found || path.is_empty() {
                continue;
            }
            let message = if path.len() > 1 {
                format!(
                    "{} not found via {}",
                    name.text(),
                    path.iter()
                        .map(|p| relative(file.file_name(), p))
                        .collect::<Vec<_>>()
                        .join(" -> ")
                )
            } else {
                format!("{} not found in '{}'", name.text(), source.text())
            };
            result.push(Violation::new(name, "", message));
        }
    });
    result
}
fn named_as_default(file: P<SourceFile>, graph: &mut Graph) -> Vec<Violation> {
    let mut result = vec![];
    walk(file.as_node(), &mut |node| {
        if !ast::is_import_declaration_or_js_import_declaration(node) {
            return;
        }
        let decl = node.as_import_declaration();
        let Some(name) = decl.import_clause.and_then(|n| n.name()) else {
            return;
        };
        if name.text() == "default" {
            return;
        }
        let Some(link) = graph.link(file, decl.module_specifier, true) else {
            return;
        };
        let id = graph.map_for_link(&link);
        let map = &graph.maps[id];
        if (!map.has_default() && !link.node_default) || !map.exports.contains_key(name.text()) {
            return;
        }
        if map
            .reexports
            .get(name.text())
            .is_some_and(|s| !s.is_empty() && Some(s) == map.reexports.get("default"))
        {
            return;
        }
        result.push(Violation::new(
            name,
            "noNamedAsDefault",
            format!(
                "Using exported name '{}' as identifier for default import.",
                name.text()
            ),
        ));
    });
    result
}

fn namespaces(file: P<SourceFile>, graph: &mut Graph) -> Vec<Violation> {
    let mut result = vec![];
    let mut namespaces = HashMap::<String, usize>::new();
    for &stmt in file.statements().nodes() {
        if stmt.kind() != Kind::ImportDeclaration {
            continue;
        }
        let decl = stmt.as_import_declaration();
        let Some(clause) = decl.import_clause else {
            continue;
        };
        let Some(link) = graph.link(file, decl.module_specifier, true) else {
            continue;
        };
        let imports = graph.map_for_link(&link);
        if let Some(name) = clause.name()
            && let Some(namespace) = graph.maps[imports]
                .exports
                .get("default")
                .and_then(|meta| meta.namespace)
        {
            namespaces.insert(name.text().to_owned(), namespace);
        }
        let Some(bindings) = clause.as_import_clause().named_bindings() else {
            continue;
        };
        if bindings.kind() == Kind::NamespaceImport {
            if let Some(name) = bindings.name() {
                if graph.maps[imports].empty() {
                    result.push(no_exports(name, decl.module_specifier.text()));
                }
                namespaces.insert(name.text().to_owned(), imports);
            }
        } else if bindings.kind() == Kind::NamedImports {
            for &specifier in bindings.elements() {
                if let (Some(imported), Some(local)) =
                    (specifier.property_name_or_name(), specifier.name())
                    && let Some(namespace) = graph.maps[imports]
                        .exports
                        .get(imported.text())
                        .and_then(|meta| meta.namespace)
                {
                    namespaces.insert(local.text().to_owned(), namespace);
                }
            }
        }
    }
    walk(file.as_node(), &mut |node| {
        if let Some((object, _)) = member(node) {
            let object = ast::skip_parentheses(object);
            if object.kind() != Kind::Identifier {
                return;
            }
            let name = object.text();
            let Some(&namespace) = namespaces.get(name) else {
                return;
            };
            if shadowed(object, name) {
                return;
            }
            if let Some(parent) = node.parent().filter(|p| {
                p.kind() == Kind::BinaryExpression
                    && ast::is_assignment_operator(p.as_binary_expression().operator_token().kind())
                    && ast::skip_parentheses(p.as_binary_expression().left()) == node
            }) {
                result.push(Violation::new(
                    parent,
                    "assignment",
                    format!("Assignment to member of namespace '{name}'."),
                ));
            }
            namespace_access(node, namespace, name, graph, &mut result);
        } else if node.kind() == Kind::VariableDeclaration {
            let decl = node.as_variable_declaration();
            let Some(init) = decl
                .initializer()
                .map(ast::skip_parentheses)
                .filter(|n| n.kind() == Kind::Identifier)
            else {
                return;
            };
            let name = init.text();
            let Some(&namespace) = namespaces.get(name) else {
                return;
            };
            if shadowed(init, name) {
                return;
            }
            binding_pattern(
                decl.name(),
                namespace,
                &[name.to_owned()],
                graph,
                &mut result,
            );
        } else if node.kind() == Kind::NamespaceExport {
            let mut ancestor = node.parent();
            while let Some(parent) = ancestor {
                if parent.kind() == Kind::ExportDeclaration {
                    if let Some(source) = parent.as_export_declaration().module_specifier
                        && let Some(link) = graph.link(file, source, true)
                    {
                        let id = graph.map_for_link(&link);
                        if graph.maps[id].empty() {
                            result.push(no_exports(node, source.text()));
                        }
                    }
                    break;
                }
                ancestor = parent.parent();
            }
        }
    });
    result
}
fn no_exports(node: P<Node>, source: &str) -> Violation {
    Violation::new(
        node,
        "noExports",
        format!("No exported names found in module '{source}'."),
    )
}
fn not_found(node: P<Node>, name: &str, path: &[String]) -> Violation {
    Violation::new(
        node,
        "notFound",
        format!(
            "'{name}' not found in {} namespace '{}'.",
            if path.len() > 1 {
                "deeply imported"
            } else {
                "imported"
            },
            path.join(".")
        ),
    )
}
fn namespace_access(
    mut node: P<Node>,
    mut namespace: usize,
    root: &str,
    graph: &Graph,
    result: &mut Vec<Violation>,
) {
    let mut path = vec![root.to_owned()];
    loop {
        let Some((_, property)) = member(node) else {
            return;
        };
        if node.kind() == Kind::ElementAccessExpression {
            result.push(Violation::new(
                property,
                "computed",
                format!("Unable to validate computed reference to imported namespace '{root}'."),
            ));
            return;
        }
        let name = property.text();
        if !graph.maps[namespace].has(name) {
            result.push(not_found(property, name, &path));
            return;
        }
        let Some(next) = graph.maps[namespace]
            .exports
            .get(name)
            .and_then(|m| m.namespace)
        else {
            return;
        };
        namespace = next;
        path.push(name.to_owned());
        let mut parent = node.parent();
        while parent.is_some_and(|p| p.kind() == Kind::ParenthesizedExpression) {
            parent = parent.and_then(|p| p.parent());
        }
        let Some(parent) = parent.filter(|&p| {
            member(p).is_some_and(|(object, _)| ast::skip_parentheses(object) == node)
        }) else {
            return;
        };
        node = parent;
    }
}
fn binding_pattern(
    pattern: P<Node>,
    namespace: usize,
    path: &[String],
    graph: &Graph,
    result: &mut Vec<Violation>,
) {
    if pattern.kind() != Kind::ObjectBindingPattern {
        return;
    }
    for &element in pattern.elements() {
        if element.kind() != Kind::BindingElement
            || element.as_binding_element().dot_dot_dot_token().is_some()
        {
            continue;
        }
        let Some(key) = element.property_name_or_name() else {
            continue;
        };
        if key.kind() != Kind::Identifier {
            result.push(Violation::new(
                element,
                "onlyTopLevel",
                "Only destructure top-level names.",
            ));
            continue;
        }
        if !graph.maps[namespace].has(key.text()) {
            result.push(not_found(key, key.text(), path));
            continue;
        }
        if let Some(nested) = graph.maps[namespace]
            .exports
            .get(key.text())
            .and_then(|m| m.namespace)
        {
            let mut path = path.to_vec();
            path.push(key.text().to_owned());
            if let Some(name) = element.name() {
                binding_pattern(name, nested, &path, graph, result);
            }
        }
    }
}
fn member(node: P<Node>) -> Option<(P<Node>, P<Node>)> {
    match node.kind() {
        Kind::PropertyAccessExpression => Some((
            node.as_property_access_expression().expression(),
            node.name()?,
        )),
        Kind::ElementAccessExpression => Some((
            node.as_element_access_expression().expression(),
            node.as_element_access_expression().argument_expression(),
        )),
        Kind::QualifiedName => {
            let mut entity = node;
            while let Some(parent) = entity
                .parent()
                .filter(|p| p.kind() == Kind::QualifiedName && p.as_qualified_name().left == entity)
            {
                entity = parent;
            }
            let reference = entity.parent()?;
            if reference.kind() != Kind::TypeReference
                || reference
                    .parent()
                    .is_none_or(|p| p.kind() != Kind::HeritageClause)
            {
                return None;
            }
            Some((
                node.as_qualified_name().left,
                node.as_qualified_name().right,
            ))
        }
        _ => None,
    }
}
fn binding_has(node: P<Node>, name: &str) -> bool {
    let mut names = vec![];
    binding_names(node, &mut names);
    names.iter().any(|n| n == name)
}
fn list_has(node: P<Node>, name: &str) -> bool {
    node.kind() == Kind::VariableDeclarationList
        && node
            .as_variable_declaration_list()
            .declarations
            .nodes()
            .iter()
            .any(|n| n.name().is_some_and(|n| binding_has(n, name)))
}
fn hoisted(node: P<Node>, name: &str) -> bool {
    if node.kind() == Kind::ModuleDeclaration
        || ast::is_function_like_or_class_static_block_declaration(node)
    {
        return false;
    }
    if node.kind() == Kind::VariableDeclarationList
        && !node.flags().intersects(NodeFlags::BlockScoped)
    {
        return list_has(node, name);
    }
    let mut found = false;
    node.for_each_child(&mut |child| {
        found |= hoisted(child, name);
        found
    });
    found
}
fn signature_namespace(node: P<Node>, name: &str) -> bool {
    fn in_type(node: P<Node>, name: &str) -> bool {
        let mut found = false;
        walk(node, &mut |mut n| {
            if n.kind() == Kind::QualifiedName {
                while n.kind() == Kind::QualifiedName {
                    n = n.as_qualified_name().left;
                }
                if n.kind() == Kind::Identifier && n.text() == name {
                    found = true;
                }
            }
        });
        found
    }
    node.parameters()
        .iter()
        .any(|p| p.type_node().is_some_and(|t| in_type(t, name)))
        || node.type_node().is_some_and(|t| in_type(t, name))
}
fn shadowed(node: P<Node>, name: &str) -> bool {
    let mut current = node.parent();
    let mut nested_function = false;
    let mut nested_signature = false;
    while let Some(scope) = current {
        if scope.kind() == Kind::SourceFile {
            break;
        }
        if ast::is_function_like_declaration(scope) {
            let signature = signature_namespace(scope, name);
            if scope
                .parameters()
                .iter()
                .any(|p| p.name().is_some_and(|n| binding_has(n, name)))
                && !(nested_signature || !nested_function && signature)
            {
                return true;
            }
            if matches!(
                scope.kind(),
                Kind::FunctionDeclaration | Kind::FunctionExpression
            ) && scope
                .name()
                .is_some_and(|n| n.kind() == Kind::Identifier && n.text() == name)
            {
                return true;
            }
            if scope.body().is_some_and(|body| hoisted(body, name)) {
                return true;
            }
            nested_signature |= signature;
            nested_function = true;
        }
        match scope.kind() {
            Kind::Block => {
                for &stmt in scope.statements() {
                    if stmt.kind() == Kind::VariableStatement {
                        if list_has(stmt.as_variable_statement().declaration_list, name) {
                            return true;
                        }
                    } else if matches!(
                        stmt.kind(),
                        Kind::FunctionDeclaration
                            | Kind::ClassDeclaration
                            | Kind::EnumDeclaration
                            | Kind::ModuleDeclaration
                    ) && stmt
                        .name()
                        .is_some_and(|n| n.kind() == Kind::Identifier && n.text() == name)
                        && !(stmt.kind() == Kind::ModuleDeclaration
                            && stmt
                                .body()
                                .is_some_and(|b| b.kind() == Kind::ModuleDeclaration))
                    {
                        return true;
                    }
                }
            }
            Kind::CatchClause => {
                if scope
                    .as_catch_clause()
                    .variable_declaration
                    .and_then(|n| n.name())
                    .is_some_and(|n| binding_has(n, name))
                {
                    return true;
                }
            }
            Kind::ForStatement => {
                if scope
                    .as_for_statement()
                    .initializer()
                    .is_some_and(|n| list_has(n, name))
                {
                    return true;
                }
            }
            Kind::ForInStatement | Kind::ForOfStatement => {
                if list_has(scope.as_for_in_or_of_statement().initializer(), name) {
                    return true;
                }
            }
            Kind::ClassDeclaration | Kind::ClassExpression | Kind::EnumDeclaration
                if scope
                    .name()
                    .is_some_and(|n| n.kind() == Kind::Identifier && n.text() == name) =>
            {
                return true;
            }
            _ => {}
        }
        current = scope.parent();
    }
    false
}

// The frozen host probes already-loaded files only after TypeScript failed to
// resolve a relative request. Do not open arbitrary files or widen the Program.
fn relative_loaded_file(
    program: &Program,
    origin: P<SourceFile>,
    specifier: &str,
) -> Option<P<SourceFile>> {
    if !specifier.starts_with('.') && !Path::new(specifier).is_absolute() {
        return None;
    }
    let path = Path::new(origin.file_name()).parent()?.join(specifier);
    let mut normalized = std::path::PathBuf::new();
    for part in path.components() {
        match part {
            std::path::Component::CurDir => {}
            std::path::Component::ParentDir => {
                normalized.pop();
            }
            part => normalized.push(part.as_os_str()),
        }
    }
    let base = normalized.to_str()?;
    let known = [
        ".d.mts", ".d.cts", ".d.ts", ".tsx", ".jsx", ".mts", ".cts", ".mjs", ".cjs", ".ts", ".js",
        ".json",
    ];
    let extension = known.iter().copied().find(|ext| base.ends_with(ext));
    let mut candidates = vec![base.to_owned()];
    if let Some(extension) = extension {
        let stem = &base[..base.len() - extension.len()];
        let substitutions: &[&str] = match extension {
            ".js" => &[".ts", ".tsx", ".d.ts"],
            ".jsx" => &[".tsx", ".ts", ".d.ts"],
            ".mjs" => &[".mts", ".d.mts", ".ts"],
            ".cjs" => &[".cts", ".d.cts", ".ts"],
            _ => &[],
        };
        candidates.extend(substitutions.iter().map(|ext| format!("{stem}{ext}")));
    } else {
        let extensions = [
            ".ts", ".tsx", ".mts", ".cts", ".js", ".jsx", ".mjs", ".cjs", ".d.ts", ".d.mts",
            ".d.cts",
        ];
        candidates.extend(extensions.iter().map(|ext| format!("{base}{ext}")));
        candidates.extend(extensions.iter().map(|ext| format!("{base}/index{ext}")));
    }
    candidates
        .iter()
        .find_map(|candidate| program.get_source_file(candidate))
}
