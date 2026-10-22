//! Frozen core variable rules. Compiler symbols preserve scopes; the shared CFG
//! supplies execution order and edges for backward assignment liveness.
use super::{
    rsbuild_cfg as cfg, rsbuild_core_checks::write_reference, rsbuild_unused_vars as bindings,
};
use crate::{Suggestion, Violation, edits::Edit};
use std::collections::{BTreeMap, HashMap, HashSet};
use tsrs_ast::{self as ast, Kind, Node, NodeFlags, SourceFile, Symbol, SymbolFlags};
use tsrs_checker::Checker;
use tsrs_core::P;
use tsrs_scanner::skip_trivia;
pub(crate) const NAMES: &[&str] = &[
    "no-redeclare",
    "no-unassigned-vars",
    "no-unused-private-class-members",
    "no-useless-assignment",
];
pub(crate) fn check(
    rule: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
) -> Vec<Violation> {
    if node.kind() != Kind::SourceFile {
        return Vec::new();
    }
    match rule {
        "no-redeclare" => redeclare(node, file),
        "no-unassigned-vars" => unassigned(node, checker),
        "no-unused-private-class-members" => private_members(node, file),
        "no-useless-assignment" => useless(node, file, checker),
        _ => Vec::new(),
    }
}
fn walk(node: P<Node>, f: &mut impl FnMut(P<Node>)) {
    if node.flags().intersects(NodeFlags::Reparsed) {
        return;
    }
    f(node);
    node.for_each_child(&mut |child| {
        walk(child, f);
        false
    });
}
fn pos(node: P<Node>) -> usize {
    node.pos().expect_source("authored variable position") as usize
}
fn end(node: P<Node>) -> usize {
    node.end().expect_source("authored variable end") as usize
}
fn trim(node: P<Node>, file: P<SourceFile>) -> usize {
    skip_trivia(file.text(), pos(node) as u32) as usize
}
fn variable_scope(mut node: P<Node>) -> Option<P<Node>> {
    while let Some(parent) = node.parent() {
        if ast::is_function_like_or_class_static_block_declaration(parent)
            || matches!(parent.kind(), Kind::SourceFile | Kind::ModuleBlock)
        {
            return Some(parent);
        }
        node = parent;
    }
    None
}
fn binding_names(node: P<Node>, names: &mut Vec<P<Node>>) {
    match node.kind() {
        Kind::Identifier => names.push(node),
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
fn binding_range(name: P<Node>, file: P<SourceFile>) -> [usize; 2] {
    let mut finish = end(name);
    if let Some(parent) = name.parent() {
        match parent.kind() {
            Kind::VariableDeclaration => {
                if let Some(token) = parent.as_variable_declaration().exclamation_token() {
                    finish = finish.max(end(token));
                }
                if let Some(ty) = parent.type_node() {
                    finish = finish.max(end(ty));
                }
            }
            Kind::Parameter
                if parent
                    .as_parameter_declaration()
                    .dot_dot_dot_token()
                    .is_none() =>
            {
                if let Some(token) = parent.as_parameter_declaration().question_token() {
                    finish = finish.max(end(token));
                }
                if let Some(ty) = parent.type_node() {
                    finish = finish.max(end(ty));
                }
            }
            _ => {}
        }
    }
    [trim(name, file), finish]
}
fn redeclare(root: P<Node>, file: P<SourceFile>) -> Vec<Violation> {
    // Default flat-config modules and implicit CommonJS wrappers have a distinct
    // file scope from directive globals. Admission rejects other sourceType and
    // languageOptions.globals configurations; built-ins are not lexical bindings.
    let mut scopes = HashMap::<P<Node>, HashMap<String, Vec<P<Node>>>>::new();
    scopes.insert(root, HashMap::new());
    walk(root, &mut |node| {
        let mut additions = Vec::<(P<Node>, P<Node>)>::new();
        match node.kind() {
            Kind::FunctionDeclaration => {
                if let (Some(owner), Some(name)) =
                    (ast::get_enclosing_block_scope_container(node), node.name())
                {
                    additions.push((owner, name));
                }
                if node.body().is_some() {
                    scopes.entry(node).or_default();
                    for &p in node.parameters() {
                        if let Some(name) = p.name() {
                            let mut names = Vec::new();
                            binding_names(name, &mut names);
                            additions.extend(names.into_iter().map(|name| (node, name)));
                        }
                    }
                    for &p in node.type_parameters() {
                        if let Some(name) = p.name() {
                            additions.push((node, name));
                        }
                    }
                }
            }
            Kind::FunctionExpression
            | Kind::ArrowFunction
            | Kind::MethodDeclaration
            | Kind::Constructor
            | Kind::GetAccessor
            | Kind::SetAccessor => {
                if node.body().is_some() {
                    scopes.entry(node).or_default();
                    for &p in node.parameters() {
                        if let Some(name) = p.name() {
                            let mut names = Vec::new();
                            binding_names(name, &mut names);
                            additions.extend(names.into_iter().map(|name| (node, name)));
                        }
                    }
                    for &p in node.type_parameters() {
                        if let Some(name) = p.name() {
                            additions.push((node, name));
                        }
                    }
                }
            }
            Kind::Block
                if node.parent().is_some_and(|parent| {
                    !ast::is_function_like_or_class_static_block_declaration(parent)
                }) =>
            {
                scopes.entry(node).or_default();
            }
            Kind::ClassStaticBlockDeclaration => {
                scopes.entry(node).or_default();
            }
            Kind::ForStatement | Kind::ForInStatement | Kind::ForOfStatement => {
                if node.initializer().is_some_and(|list| {
                    list.kind() == Kind::VariableDeclarationList
                        && list.flags().intersects(NodeFlags::BlockScoped)
                }) {
                    scopes.entry(node).or_default();
                }
            }
            Kind::SwitchStatement => {
                scopes
                    .entry(node.as_switch_statement().case_block())
                    .or_default();
            }
            Kind::VariableDeclarationList => {
                let owner = if node.flags().intersects(NodeFlags::BlockScoped) {
                    ast::get_enclosing_block_scope_container(node)
                } else {
                    variable_scope(node)
                };
                if let Some(owner) = owner {
                    for &decl in node.as_variable_declaration_list().declarations().nodes() {
                        if let Some(name) = decl.name() {
                            let mut names = Vec::new();
                            binding_names(name, &mut names);
                            additions.extend(names.into_iter().map(|name| (owner, name)));
                        }
                    }
                }
            }
            Kind::ClassDeclaration
            | Kind::InterfaceDeclaration
            | Kind::TypeAliasDeclaration
            | Kind::EnumDeclaration
            | Kind::ModuleDeclaration => {
                let skip = node.kind() == Kind::ModuleDeclaration
                    && (ast::is_global_scope_augmentation(node)
                        || node
                            .body()
                            .is_some_and(|body| body.kind() == Kind::ModuleDeclaration));
                if !skip
                    && let (Some(owner), Some(name)) =
                        (ast::get_enclosing_block_scope_container(node), node.name())
                    && name.kind() == Kind::Identifier
                {
                    additions.push((owner, name));
                }
            }
            Kind::ImportDeclaration | Kind::ImportEqualsDeclaration => {
                if let Some(owner) = ast::get_enclosing_block_scope_container(node) {
                    walk(node, &mut |child| {
                        if matches!(
                            child.kind(),
                            Kind::ImportClause
                                | Kind::ImportSpecifier
                                | Kind::NamespaceImport
                                | Kind::ImportEqualsDeclaration
                        ) && let Some(name) = child.name()
                            && name.kind() == Kind::Identifier
                        {
                            additions.push((owner, name));
                        }
                    });
                }
            }
            _ => {}
        }
        for (owner, name) in additions {
            if let Some(scope) = scopes.get_mut(&owner) {
                scope.entry(name.text().to_owned()).or_default().push(name);
            }
        }
    });
    let mut reports = Vec::new();
    for scope in scopes.values() {
        for (name, decls) in scope {
            for &node in &decls[1..] {
                reports.push(
                    Violation::new(node, "redeclared", format!("'{name}' is already defined."))
                        .with_byte_range(binding_range(node, file)),
                );
            }
        }
    }
    // Directives remain in the outer global scope, independently of module
    // declarations. Every comment contributes one declaration; duplicates
    // within the same comment have already collapsed to their first range.
    let declarations = crate::directives::global_declarations(file).expect("validated directives");
    for declaration in &declarations {
        if !declaration.access.is_declared() {
            continue;
        }
        let name = declaration.name.as_str();
        let builtin = super::rsbuild_binding_checks::ECMASCRIPT_GLOBALS.contains(&name)
            || file.file_name().ends_with(".cjs")
                && matches!(name, "exports" | "global" | "module" | "require");
        let (message_id, message) = if builtin {
            (
                "redeclaredAsBuiltin",
                format!("'{name}' is already defined as a built-in global variable."),
            )
        } else {
            ("redeclared", format!("'{name}' is already defined."))
        };
        for &range in declaration.name_ranges.iter().skip(usize::from(!builtin)) {
            reports.push(Violation::new(root, message_id, message.clone()).with_byte_range(range));
        }
    }
    reports.sort_by_key(|report| {
        report
            .byte_range
            .unwrap_or([trim(report.node, file), end(report.node)])
    });
    reports
}
fn is_reference(node: P<Node>) -> bool {
    if node.kind() != Kind::Identifier {
        return false;
    }
    let Some(parent) = node.parent() else {
        return false;
    };
    if parent.kind() == Kind::ExportSpecifier
        && ast::is_type_only_import_or_export_declaration(parent)
    {
        return false;
    }
    !bindings::property_name(node)
        && !bindings::jsx_non_reference(node)
        && !bindings::import_type_syntax(node)
        && (matches!(
            parent.kind(),
            Kind::ShorthandPropertyAssignment
                | Kind::ExportSpecifier
                | Kind::NamespaceExportDeclaration
        ) || !ast::is_declaration_name(node))
}
fn symbol(node: P<Node>, checker: &mut Checker) -> Option<P<Symbol>> {
    if node.parent().is_some_and(|p| {
        p.name() == Some(node)
            && matches!(
                p.kind(),
                Kind::VariableDeclaration | Kind::BindingElement | Kind::Parameter
            )
    }) {
        return node
            .parent()
            .and_then(|p| p.symbol())
            .map(|s| checker.get_merged_symbol(s));
    }
    if node
        .parent()
        .is_some_and(|p| p.kind() == Kind::ExportSpecifier)
    {
        let parent = node.parent().unwrap();
        if let Some(decl) = find_parent(parent, Kind::ExportDeclaration)
            && decl.as_export_declaration().module_specifier().is_some()
        {
            return None;
        }
        let local = parent
            .as_export_specifier()
            .property_name()
            .unwrap_or(parent.name().unwrap());
        if local != node {
            return None;
        }
        let meaning = if ast::is_type_only_import_or_export_declaration(parent) {
            SymbolFlags::Type | SymbolFlags::Namespace | SymbolFlags::Alias
        } else {
            SymbolFlags::Value | SymbolFlags::Type | SymbolFlags::Namespace | SymbolFlags::Alias
        };
        return bindings::export_reference(node, meaning, checker)
            .map(|s| checker.get_merged_symbol(s));
    }
    bindings::reference_symbol(node, checker).map(|s| checker.get_merged_symbol(s))
}
fn find_parent(mut node: P<Node>, kind: Kind) -> Option<P<Node>> {
    while let Some(parent) = node.parent() {
        if parent.kind() == kind {
            return Some(parent);
        }
        node = parent;
    }
    None
}
fn loop_declaration(decl: P<Node>) -> bool {
    decl.parent().is_some_and(|list| {
        list.kind() == Kind::VariableDeclarationList
            && list.parent().is_some_and(|owner| {
                matches!(owner.kind(), Kind::ForInStatement | Kind::ForOfStatement)
            })
    })
}
fn unassigned(root: P<Node>, checker: &mut Checker) -> Vec<Violation> {
    let mut refs = HashMap::<P<Symbol>, Vec<P<Node>>>::new();
    let mut declarations = Vec::new();
    walk(root, &mut |node| {
        if node.kind() == Kind::VariableDeclaration {
            declarations.push(node);
        }
        if is_reference(node)
            && let Some(s) = symbol(node, checker)
        {
            refs.entry(s).or_default().push(node);
        }
    });
    let mut out = Vec::new();
    for decl in declarations {
        if decl.initializer().is_some()
            || loop_declaration(decl)
            || decl.flags().intersects(NodeFlags::Ambient)
            || decl.parent().is_none_or(|list| {
                list.kind() != Kind::VariableDeclarationList
                    || list
                        .flags()
                        .intersects(NodeFlags::Constant | NodeFlags::Using | NodeFlags::AwaitUsing)
            })
        {
            continue;
        }
        let Some(name) = decl.name().filter(|name| name.kind() == Kind::Identifier) else {
            continue;
        };
        let mut symbols = Vec::new();
        for sym in [decl.symbol(), decl.local_symbol()].into_iter().flatten() {
            for s in [Some(sym), sym.export_symbol()].into_iter().flatten() {
                let s = checker.get_merged_symbol(s);
                if !symbols.contains(&s) {
                    symbols.push(s);
                }
            }
        }
        if symbols.iter().any(|s| {
            s.declarations().iter().any(|d| {
                d.kind() == Kind::VariableDeclaration
                    && (d.initializer().is_some() || loop_declaration(*d))
            })
        }) {
            continue;
        }
        let uses: Vec<_> = symbols
            .iter()
            .flat_map(|s| refs.get(s).into_iter().flatten())
            .copied()
            .collect();
        if uses.iter().any(|n| write_reference(*n)) {
            continue;
        }
        if uses
            .iter()
            .any(|n| !ast::is_part_of_type_node(*n) && !ast::is_part_of_type_query(*n))
        {
            out.push(Violation::new(
                decl,
                "unassigned",
                format!(
                    "'{}' is always 'undefined' because it's never assigned.",
                    name.text()
                ),
            ));
        }
    }
    out
}
#[derive(Clone)]
struct Item {
    start: usize,
    end: usize,
    text: String,
    comment: bool,
}
fn items(file: P<SourceFile>) -> Vec<Item> {
    let mut out = Vec::new();
    crate::comments::for_each_token_range(file, |_, start, end, _| {
        let start = skip_trivia(file.text(), start) as usize;
        let end = end as usize;
        if start < end {
            out.push(Item {
                start,
                end,
                text: file.text()[start..end].to_owned(),
                comment: false,
            });
        }
        Ok(())
    })
    .expect("parsed token ranges");
    crate::comments::for_each_comment(file, |comment| {
        let start = comment.pos() as usize;
        let end = comment.end() as usize;
        out.push(Item {
            start,
            end,
            text: file.text()[start..end].to_owned(),
            comment: true,
        });
        Ok(())
    })
    .expect("parsed comment ranges");
    out.sort_by_key(|i| (i.start, i.end, i.comment));
    out.dedup_by_key(|i| (i.start, i.end, i.comment));
    out
}
fn before(items: &[Item], position: usize, comments: bool) -> Option<&Item> {
    items
        .iter()
        .take_while(|i| i.start < position)
        .filter(|i| i.end <= position && (comments || !i.comment))
        .last()
}
fn after(items: &[Item], position: usize, comments: bool) -> Option<&Item> {
    items
        .iter()
        .find(|i| i.start >= position && (comments || !i.comment))
}
fn line(text: &str, position: usize) -> usize {
    let mut n = 0;
    let mut prev = '\0';
    for c in text[..position].chars() {
        if c == '\r' || c == '\u{2028}' || c == '\u{2029}' || c == '\n' && prev != '\r' {
            n += 1;
        }
        prev = c;
    }
    n
}
fn line_start(text: &str, mut p: usize) -> usize {
    let bytes = text.as_bytes();
    while p > 0 && !matches!(bytes[p - 1], b'\n' | b'\r') {
        p -= 1;
    }
    p
}
fn shares(text: &str, member: &Item, items: &[Item]) -> bool {
    before(items, member.start, false)
        .is_some_and(|i| line(text, i.end) == line(text, member.start))
        || after(items, member.end, false)
            .is_some_and(|i| line(text, i.start) == line(text, member.end))
}
fn own_line(text: &str, item: &Item, items: &[Item]) -> bool {
    before(items, item.start, true).is_none_or(|i| line(text, i.end) != line(text, item.start))
}
fn removal(file: P<SourceFile>, node: P<Node>, items: &[Item]) -> [usize; 2] {
    let text = file.text();
    let member = Item {
        start: trim(node, file),
        end: end(node),
        text: String::new(),
        comment: false,
    };
    let prev = before(items, member.start, false);
    let start_after = prev.map_or(0, |i| i.end);
    let comments: Vec<_> = items
        .iter()
        .filter(|i| i.comment && i.start >= start_after && i.end <= member.start)
        .collect();
    let mut first = 0;
    for (index, comment) in comments.iter().enumerate() {
        let next = comments.get(index + 1).copied().unwrap_or(&member);
        if !own_line(text, comment, items)
            || line(text, next.start).saturating_sub(line(text, comment.end)) > 1
        {
            first = index + 1;
        }
    }
    let leading = &comments[first..];
    let remove_leading = !leading.is_empty() && !shares(text, &member, items);
    let next = after(items, member.end, false);
    let end_before = next.map_or(text.len(), |i| i.start);
    let trailing: Vec<_> = if shares(text, &member, items) {
        Vec::new()
    } else {
        items
            .iter()
            .filter(|i| {
                i.comment
                    && i.start >= member.end
                    && i.end <= end_before
                    && line(text, i.start) == line(text, member.end)
            })
            .collect()
    };
    let last = trailing.last().copied().unwrap_or(&member);
    let next = after(items, last.end, true);
    let newline = next.is_some_and(|i| line(text, i.start) > line(text, last.end));
    let mut start = member.start;
    let mut end = last.end;
    if remove_leading {
        if newline {
            start = line_start(text, leading[0].start);
            end = line_start(text, next.unwrap().start);
        } else {
            start = leading[0].start;
            end = next.map_or(last.end, |i| i.start);
        }
    } else if own_line(text, &member, items) && newline {
        start = line_start(text, member.start);
        end = line_start(text, next.unwrap().start);
    } else if let Some(previous) = prev.filter(|i| line(text, i.end) == line(text, member.start)) {
        start = previous.end;
    } else if let Some(next) = next.filter(|i| line(text, i.start) == line(text, last.end)) {
        end = next.start;
    }
    [start, end]
}
fn insertion(file: P<SourceFile>, class: P<Node>, node: P<Node>, items: &[Item]) -> Option<usize> {
    let next = after(items, end(node), false)?;
    if !matches!(next.text.as_str(), "[" | "*" | "in" | "instanceof") {
        return None;
    }
    let prev = before(items, trim(node, file), false)?;
    if matches!(prev.text.as_str(), ":" | ";" | "{" | "=>" | "++" | "--") {
        return None;
    }
    let index = class.members().iter().position(|&member| member == node)?;
    let previous = *class.members().get(index.checked_sub(1)?)?;
    if previous.kind() != Kind::PropertyDeclaration {
        return None;
    }
    let init = previous.initializer()?;
    if file.text()[pos(previous)..end(previous)]
        .trim_end_matches([' ', '\t', '\n', '\r'])
        .ends_with(';')
        || init.kind() == Kind::PostfixUnaryExpression
        || init.kind() == Kind::ArrowFunction
            && init.body().is_some_and(|body| body.kind() == Kind::Block)
    {
        return None;
    }
    Some(prev.end)
}
struct Member {
    class: P<Node>,
    decl: P<Node>,
    name: P<Node>,
    accessor: bool,
    referenced: bool,
    used: bool,
}
fn private_members(root: P<Node>, file: P<SourceFile>) -> Vec<Violation> {
    let mut members = Vec::<Member>::new();
    let mut classes = HashMap::<P<Node>, HashMap<String, usize>>::new();
    walk(root, &mut |node| {
        if ast::is_class_like(node) {
            let mut map = HashMap::new();
            for &decl in node.members() {
                if matches!(
                    decl.kind(),
                    Kind::PropertyDeclaration
                        | Kind::MethodDeclaration
                        | Kind::GetAccessor
                        | Kind::SetAccessor
                ) && let Some(name) = decl
                    .name()
                    .filter(|name| name.kind() == Kind::PrivateIdentifier)
                {
                    let index = map.get(name.text()).copied().unwrap_or(members.len());
                    let m = Member {
                        class: node,
                        decl,
                        name,
                        accessor: matches!(decl.kind(), Kind::GetAccessor | Kind::SetAccessor),
                        referenced: false,
                        used: false,
                    };
                    if index == members.len() {
                        members.push(m);
                    } else {
                        members[index] = m;
                    }
                    map.insert(name.text().to_owned(), index);
                }
            }
            classes.insert(node, map);
        }
    });
    walk(root, &mut |node| {
        if node.kind() != Kind::PrivateIdentifier
            || node.parent().is_some_and(|parent| {
                matches!(
                    parent.kind(),
                    Kind::PropertyDeclaration
                        | Kind::MethodDeclaration
                        | Kind::GetAccessor
                        | Kind::SetAccessor
                ) && parent.name() == Some(node)
            })
        {
            return;
        }
        let mut owner = node.parent();
        while let Some(class) = owner {
            if let Some(&index) = classes.get(&class).and_then(|map| map.get(node.text())) {
                let member = &mut members[index];
                member.referenced = true;
                if member.accessor
                    || node.parent().is_none_or(|parent| {
                        parent.kind() != Kind::PropertyAccessExpression
                            || !private_write_only(parent)
                    })
                {
                    member.used = true;
                }
                break;
            }
            owner = class.parent();
        }
    });
    let source_items = items(file);
    let mut out = Vec::new();
    for member in members {
        if member.used {
            continue;
        }
        let name = member.name.text();
        let mut report = Violation::new(
            member.name,
            "unusedPrivateClassMember",
            format!("'{name}' is defined but never used."),
        );
        if !member.referenced {
            let mut edits = vec![Edit::replace(removal(file, member.decl, &source_items), "")];
            if let Some(position) = insertion(file, member.class, member.decl, &source_items) {
                edits.push(Edit::replace([position, position], ";"));
            }
            report.suggestions.push(Suggestion {
                message_id: "removeUnusedPrivateClassMember",
                message: format!("Remove unused private class member '{name}'."),
                data: Some(BTreeMap::from([(
                    "classMemberName".to_owned(),
                    name.to_owned(),
                )])),
                edits,
            });
        }
        out.push(report);
    }
    out
}
fn private_write_only(node: P<Node>) -> bool {
    let Some(target) = ast::get_assignment_target(node) else {
        return false;
    };
    match target.kind() {
        Kind::BinaryExpression
            if target.as_binary_expression().operator_token().kind() == Kind::EqualsToken =>
        {
            true
        }
        Kind::BinaryExpression | Kind::PrefixUnaryExpression | Kind::PostfixUnaryExpression => {
            let mut parent = target.parent();
            while parent.is_some_and(|p| p.kind() == Kind::ParenthesizedExpression) {
                parent = parent.and_then(|p| p.parent());
            }
            parent.is_some_and(|p| p.kind() == Kind::ExpressionStatement)
        }
        Kind::ForInStatement | Kind::ForOfStatement => true,
        _ => false,
    }
}

fn assignment_names(mut node: P<Node>, out: &mut Vec<P<Node>>) {
    while node.kind() == Kind::ParenthesizedExpression {
        node = node.expression().unwrap();
    }
    match node.kind() {
        Kind::Identifier => out.push(node),
        Kind::ObjectBindingPattern | Kind::ArrayBindingPattern => binding_names(node, out),
        Kind::ObjectLiteralExpression => {
            for &property in node.as_object_literal_expression().properties().nodes() {
                match property.kind() {
                    Kind::PropertyAssignment => {
                        assignment_names(property.initializer().unwrap(), out)
                    }
                    Kind::ShorthandPropertyAssignment => {
                        assignment_names(property.name().unwrap(), out)
                    }
                    Kind::SpreadAssignment => assignment_names(property.expression().unwrap(), out),
                    _ => {}
                }
            }
        }
        Kind::ArrayLiteralExpression => {
            for &element in node.elements() {
                assignment_names(element, out);
            }
        }
        Kind::SpreadElement => assignment_names(node.expression().unwrap(), out),
        Kind::BinaryExpression
            if node.as_binary_expression().operator_token().kind() == Kind::EqualsToken =>
        {
            assignment_names(node.as_binary_expression().left(), out)
        }
        // Assertions intentionally are not patterns in the frozen rule.
        _ => {}
    }
}
fn read_write(mut node: P<Node>) -> bool {
    while let Some(parent) = node.parent() {
        if matches!(
            parent.kind(),
            Kind::ParenthesizedExpression
                | Kind::NonNullExpression
                | Kind::AsExpression
                | Kind::TypeAssertionExpression
                | Kind::SatisfiesExpression
        ) {
            node = parent;
        } else {
            break;
        }
    }
    let Some(parent) = node.parent() else {
        return false;
    };
    match parent.kind() {
        Kind::BinaryExpression => {
            let binary = parent.as_binary_expression();
            binary.left() == node
                && binary.operator_token().kind() != Kind::EqualsToken
                && ast::is_assignment_operator(binary.operator_token().kind())
        }
        Kind::PrefixUnaryExpression => {
            let unary = parent.as_prefix_unary_expression();
            unary.operand() == node
                && matches!(
                    unary.operator(),
                    Kind::PlusPlusToken | Kind::MinusMinusToken
                )
        }
        Kind::PostfixUnaryExpression => parent.as_postfix_unary_expression().operand() == node,
        _ => false,
    }
}
fn exported(sym: P<Symbol>, root: P<Node>, aliases: &HashSet<P<Symbol>>) -> bool {
    aliases.contains(&sym)
        || sym.declarations().iter().any(|&decl| {
            let mut target = Some(decl);
            while let Some(node) = target {
                if matches!(
                    node.kind(),
                    Kind::BindingElement
                        | Kind::ObjectBindingPattern
                        | Kind::ArrayBindingPattern
                        | Kind::VariableDeclaration
                        | Kind::VariableDeclarationList
                ) {
                    target = node.parent();
                } else {
                    break;
                }
            }
            target.is_some_and(|node| {
                node.parent() == Some(root)
                    && ast::has_syntactic_modifier(node, ast::ModifierFlags::Export)
            })
        })
}
fn in_try(mut node: P<Node>, root: P<Node>) -> bool {
    while let Some(parent) = node.parent() {
        if parent == root {
            return false;
        }
        if parent.kind() == Kind::TryStatement && parent.as_try_statement().try_block() == node {
            return true;
        }
        node = parent;
    }
    false
}
#[derive(Clone, Copy)]
struct Assignment {
    name: P<Node>,
    symbol: P<Symbol>,
}
fn useless(root: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> Vec<Violation> {
    let mut references = HashMap::<P<Symbol>, Vec<P<Node>>>::new();
    let mut by_root = HashMap::<P<Node>, Vec<Assignment>>::new();
    let mut roots = Vec::new();
    let mut aliases = HashSet::new();
    walk(root, &mut |node| {
        if is_reference(node)
            && let Some(sym) = symbol(node, checker)
        {
            references.entry(sym).or_default().push(node);
            if node
                .parent()
                .is_some_and(|parent| parent.kind() == Kind::ExportSpecifier)
                && find_parent(node, Kind::ExportDeclaration).is_some_and(|decl| {
                    decl.parent() == Some(root)
                        && decl.as_export_declaration().module_specifier().is_none()
                })
            {
                aliases.insert(sym);
            }
        }
        let target = match node.kind() {
            Kind::VariableDeclaration if node.initializer().is_some() => node.name(),
            Kind::BinaryExpression
                if ast::is_assignment_operator(
                    node.as_binary_expression().operator_token().kind(),
                ) =>
            {
                Some(node.as_binary_expression().left())
            }
            Kind::PrefixUnaryExpression
                if matches!(
                    node.as_prefix_unary_expression().operator(),
                    Kind::PlusPlusToken | Kind::MinusMinusToken
                ) =>
            {
                Some(node.as_prefix_unary_expression().operand())
            }
            Kind::PostfixUnaryExpression => Some(node.as_postfix_unary_expression().operand()),
            _ => None,
        };
        if let Some(target) = target {
            let mut names = Vec::new();
            assignment_names(target, &mut names);
            for name in names {
                if let (Some(owner), Some(sym)) = (cfg::root_of(name), symbol(name, checker)) {
                    if !by_root.contains_key(&owner) {
                        roots.push(owner);
                    }
                    by_root
                        .entry(owner)
                        .or_default()
                        .push(Assignment { name, symbol: sym });
                }
            }
        }
    });
    let mut reports = Vec::new();
    for owner in roots {
        let mut variables = HashMap::<P<Symbol>, Option<usize>>::new();
        let mut count = 0;
        let mut reads = HashMap::new();
        let mut writes = Vec::new();
        for assignment in &by_root[&owner] {
            let variable = *variables.entry(assignment.symbol).or_insert_with(|| {
                let sym = assignment.symbol;
                let declared_here = sym.declarations().iter().any(|decl| {
                    decl.name()
                        .is_some_and(|name| cfg::root_of(name) == Some(owner))
                });
                if !declared_here
                    || ast::is_external_module(file)
                        && owner == root
                        && exported(sym, root, &aliases)
                {
                    return None;
                }
                let references: Vec<_> = references
                    .get(&sym)
                    .into_iter()
                    .flatten()
                    .copied()
                    .filter(|&node| !write_reference(node) || read_write(node))
                    .collect();
                if references.is_empty()
                    || references
                        .iter()
                        .any(|&node| cfg::root_of(node) != Some(owner))
                {
                    return None;
                }
                let variable = count;
                count += 1;
                for read in references {
                    reads.insert(read, variable);
                }
                Some(variable)
            });
            if let Some(variable) = variable {
                writes.push((assignment.name, variable));
            }
        }
        if writes.is_empty() {
            continue;
        }
        reports.extend(
            dead_writes(owner, &reads, &writes, count)
                .into_iter()
                .filter(|&node| !in_try(node, owner)),
        );
    }
    reports.sort_by_key(|&node| pos(node));
    reports
        .into_iter()
        .map(|node| {
            Violation::new(
                node,
                "unnecessaryAssignment",
                format!(
                    "The value assigned to '{}' is not used in subsequent statements.",
                    node.text()
                ),
            )
        })
        .collect()
}
#[derive(Clone, Copy)]
struct VariableEvent {
    variable: usize,
    write: Option<usize>,
}
fn dead_writes(
    root: P<Node>,
    reads: &HashMap<P<Node>, usize>,
    writes: &[(P<Node>, usize)],
    variables: usize,
) -> Vec<P<Node>> {
    let graph = cfg::promise_flow(root);
    let indices: HashMap<_, _> = writes
        .iter()
        .enumerate()
        .map(|(index, &(node, _))| (node, index))
        .collect();
    let mut dead = vec![false; writes.len()];
    let mut events = vec![Vec::<VariableEvent>::new(); graph.blocks.len()];
    for (index, block) in graph.blocks.iter().enumerate() {
        if !block.reachable {
            continue;
        }
        for event in &block.events {
            match event {
                cfg::PromiseEvent::Read(node) => {
                    if let Some(&variable) = reads.get(node) {
                        events[index].push(VariableEvent {
                            variable,
                            write: None,
                        });
                    }
                }
                cfg::PromiseEvent::Write(node) => {
                    if let Some(&write) = indices.get(node) {
                        dead[write] = true;
                        events[index].push(VariableEvent {
                            variable: writes[write].1,
                            write: Some(write),
                        });
                    }
                }
                _ => {}
            }
        }
    }
    let words = variables.div_ceil(64);
    let mut definitions = vec![vec![0u64; words]; graph.blocks.len()];
    let mut uses = definitions.clone();
    for (block, events) in events.iter().enumerate() {
        for event in events.iter().rev() {
            let (word, bit) = (event.variable / 64, 1u64 << (event.variable % 64));
            if event.write.is_some() {
                definitions[block][word] |= bit;
                uses[block][word] &= !bit;
            } else {
                uses[block][word] |= bit;
            }
        }
    }
    let mut live_in = vec![vec![0u64; words]; graph.blocks.len()];
    let mut live_out = live_in.clone();
    loop {
        let mut changed = false;
        for (index, block) in graph.blocks.iter().enumerate().rev() {
            if !block.reachable {
                continue;
            }
            live_out[index].fill(0);
            for &successor in &block.successors {
                if graph.blocks[successor].reachable {
                    for word in 0..words {
                        live_out[index][word] |= live_in[successor][word];
                    }
                }
            }
            for word in 0..words {
                let incoming =
                    uses[index][word] | (live_out[index][word] & !definitions[index][word]);
                changed |= live_in[index][word] != incoming;
                live_in[index][word] = incoming;
            }
        }
        if !changed {
            break;
        }
    }
    for (index, events) in events.iter().enumerate() {
        let live = &mut live_out[index];
        for event in events.iter().rev() {
            let (word, bit) = (event.variable / 64, 1u64 << (event.variable % 64));
            if let Some(write) = event.write {
                if live[word] & bit != 0 {
                    dead[write] = false;
                }
                live[word] &= !bit;
            } else {
                live[word] |= bit;
            }
        }
    }
    writes
        .iter()
        .zip(dead)
        .filter_map(|(&(node, _), dead)| dead.then_some(node))
        .collect()
}
