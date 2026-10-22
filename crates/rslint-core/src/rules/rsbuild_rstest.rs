//! Rstest rules and registration grammar from rslint e0a29f751d4def86d27931759098f27748fa3003.
//! Imports are resolved through tsrs symbols; modifiers retain their source accessor
//! through const aliases so focus suggestions edit the declaration that introduced it.

use crate::{Violation, edits::Edit};
use std::collections::HashSet;
use tsrs_ast::{self as ast, Kind, Node, NodeFlags, SourceFile, Symbol};
use tsrs_checker::{Checker, TypeFlags};
use tsrs_core::P;
use tsrs_scanner::{Scanner, skip_trivia};

pub(crate) const NAMES: &[&str] = &[
    "rstest/no-import-node-test",
    "rstest/no-mocks-import",
    "rstest/no-disabled-tests",
    "rstest/no-focused-tests",
    "rstest/no-interpolation-in-snapshots",
];

pub(crate) fn check(
    name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    has_type_information: bool,
) -> Vec<Violation> {
    // A file-local pass owns focus suggestion deduplication and callback bindings.
    if node.kind() != Kind::SourceFile {
        return vec![];
    }
    let mut nodes = vec![];
    collect(node, &mut nodes);
    let mut reports = vec![];
    let mut suggested = HashSet::new();
    let contexts = if name == "rstest/no-interpolation-in-snapshots" {
        context_bindings(&nodes, file, checker)
    } else {
        vec![]
    };
    for node in nodes.iter().copied() {
        match name {
            "rstest/no-import-node-test" => {
                if let Some(report) = node_test(node, file, &nodes) {
                    reports.push(report);
                }
            }
            "rstest/no-mocks-import" => {
                if let Some(report) = mocks_import(node) {
                    reports.push(report);
                }
            }
            "rstest/no-disabled-tests" | "rstest/no-focused-tests"
                if node.kind() == Kind::CallExpression =>
            {
                let Some(parsed) = registration(node, file, checker) else {
                    continue;
                };
                if name == "rstest/no-disabled-tests" {
                    if parsed.skip {
                        reports.push(Violation::new(
                            node,
                            "skippedTest",
                            "Tests should not be skipped",
                        ));
                    }
                    let args = node.arguments();
                    if parsed.is_test()
                        && !parsed.todo
                        && (args.len() < 2
                            || (args.len() == 2
                                && ast::skip_parentheses(args[1]).kind()
                                    == Kind::ObjectLiteralExpression))
                    {
                        reports.push(Violation::new(
                            node,
                            "missingFunction",
                            "Test is missing function argument",
                        ));
                    }
                } else if !parsed.focus.is_empty() {
                    let anchor = parsed
                        .focus
                        .iter()
                        .find(|part| parsed.site.iter().any(|site| site.node == part.node))
                        .map_or(parsed.root, |part| part.node);
                    let mut report =
                        Violation::new(anchor, "focusedTest", "Unexpected focused test");
                    if parsed
                        .focus
                        .iter()
                        .any(|part| !suggested.contains(&part.node))
                        && let Some(edits) = focus_edits(&parsed.focus, file)
                    {
                        suggested.extend(parsed.focus.iter().map(|part| part.node));
                        report = report.with_suggestion(
                            "suggestRemoveFocus",
                            "Suggest removing focus from test",
                            edits,
                        );
                    }
                    reports.push(report);
                }
            }
            "rstest/no-interpolation-in-snapshots"
                if node.kind() == Kind::CallExpression && top_call(node) == node =>
            {
                reports.extend(interpolations(
                    node,
                    file,
                    checker,
                    &contexts,
                    &nodes,
                    has_type_information,
                ));
            }
            _ => {}
        }
    }
    reports
}
pub(crate) fn collect(node: P<Node>, out: &mut Vec<P<Node>>) {
    out.push(node);
    node.for_each_child(&mut |child| {
        collect(child, out);
        false
    });
}
fn end(node: P<Node>) -> usize {
    node.end().expect_source("parsed source end") as usize
}
fn start(node: P<Node>, file: P<SourceFile>) -> usize {
    skip_trivia(
        file.text(),
        node.pos().expect_source("parsed source position"),
    ) as usize
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
fn string(node: P<Node>) -> Option<String> {
    matches!(
        node.kind(),
        Kind::StringLiteral | Kind::NoSubstitutionTemplateLiteral
    )
    .then(|| node.text().to_string())
}
fn property(node: P<Node>) -> Option<String> {
    if node.kind() == Kind::Identifier {
        Some(node.text().to_string())
    } else if node.kind() == Kind::ComputedPropertyName {
        string(skip(node.as_computed_property_name().expression()))
    } else {
        string(node)
    }
}
fn require_module(node: P<Node>, assertions: bool) -> Option<String> {
    let unwrap = |node| {
        if assertions {
            skip(node)
        } else {
            ast::skip_parentheses(node)
        }
    };
    let node = unwrap(node);
    if node.kind() != Kind::CallExpression {
        return None;
    }
    let call = node.as_call_expression();
    let receiver = unwrap(call.expression());
    if receiver.kind() != Kind::Identifier
        || receiver.text() != "require"
        || node.arguments().len() != 1
    {
        return None;
    }
    string(unwrap(node.arguments()[0]))
}
fn referenced_module(node: P<Node>) -> Option<String> {
    match node.kind() {
        Kind::ImportDeclaration | Kind::JSImportDeclaration => {
            string(node.as_import_declaration().module_specifier())
        }
        Kind::ImportEqualsDeclaration => {
            let reference = node.as_import_equals_declaration().module_reference();
            (reference.kind() == Kind::ExternalModuleReference)
                .then(|| {
                    string(ast::skip_parentheses(
                        reference.as_external_module_reference().expression(),
                    ))
                })
                .flatten()
        }
        Kind::CallExpression => require_module(node, false),
        _ => None,
    }
}
fn node_test(node: P<Node>, file: P<SourceFile>, nodes: &[P<Node>]) -> Option<Violation> {
    let (module, anchor) = match node.kind() {
        Kind::ImportDeclaration | Kind::ImportEqualsDeclaration => (referenced_module(node)?, node),
        Kind::CallExpression => {
            let callee = ast::skip_parentheses(node.as_call_expression().expression());
            if callee.kind() != Kind::Identifier || callee.text() != "require" {
                return None;
            }
            let arg = ast::skip_parentheses(*node.arguments().first()?);
            (string(arg)?, arg)
        }
        _ => return None,
    };
    if module != "node:test" && !module.starts_with("node:test/") {
        return None;
    }
    let mut report = Violation::new(
        anchor,
        "noImportNodeTest",
        "Do not import the Node test runner in Rstest test files",
    );
    if module != "node:test" || node.kind() != Kind::ImportDeclaration {
        return Some(report);
    }
    let declaration = node.as_import_declaration();
    let Some(clause) = declaration.import_clause() else {
        return Some(report);
    };
    let clause = clause.as_import_clause();
    if clause.name().is_some() {
        return Some(report);
    }
    let Some(bindings) = clause
        .named_bindings()
        .filter(|bindings| bindings.kind() == Kind::NamedImports)
    else {
        return Some(report);
    };
    let names = bindings.as_named_imports().elements().nodes();
    if names.is_empty()
        || !names.iter().all(|item| {
            let spec = item.as_import_specifier();
            ["describe", "it", "test"].contains(&spec.property_name().unwrap_or(spec.name()).text())
        })
    {
        return Some(report);
    }
    let mut rstack = false;
    let mut rstest = false;
    for reference in file.type_reference_directives() {
        let name = reference.file_name.as_str();
        rstest |= name == "@rstest/core" || name.starts_with("@rstest/core/");
        rstack |= name == "rstack/test" || name.starts_with("rstack/test/");
    }
    for &node in nodes {
        if let Some(module) = referenced_module(node) {
            rstest |= module == "@rstest/core";
            rstack |= module == "rstack/test";
        }
    }
    let module = declaration.module_specifier();
    let range = [start(module, file), end(module)];
    let raw = &file.text()[range[0]..range[1]];
    let quote = raw.as_bytes().first().copied().unwrap_or_default();
    if b"'\"".contains(&quote) && raw.as_bytes().last() == Some(&quote) {
        let replacement = if rstack && !rstest {
            "rstack/test"
        } else {
            "@rstest/core"
        };
        report = report.with_fix(vec![Edit::replace(
            range,
            format!("{}{replacement}{}", quote as char, quote as char),
        )]);
    }
    Some(report)
}
fn mocks_import(node: P<Node>) -> Option<Violation> {
    let (path, anchor) = match node.kind() {
        Kind::ImportDeclaration => (
            string(node.as_import_declaration().module_specifier())?,
            node,
        ),
        Kind::CallExpression => {
            let callee = node.as_call_expression().expression();
            if callee.kind() != Kind::Identifier || callee.text() != "require" {
                return None;
            }
            let argument = *node.arguments().first()?;
            (string(argument)?, argument)
        }
        _ => return None,
    };
    path.split('/').any(|component| component == "__mocks__").then(|| Violation::new(anchor, "noManualImport", "Mocks should not be manually imported from a __mocks__ directory. Instead use `rs.mock` and import from the original module path"))
}

#[derive(Clone)]
pub(crate) struct Part {
    pub(crate) name: String,
    pub(crate) node: P<Node>,
    pub(crate) invoked: u8,
    pub(crate) call: Option<P<Node>>,
}
struct Chain {
    root: P<Node>,
    parts: Vec<Part>,
    invoked: bool,
}
fn is_import_meta(node: P<Node>) -> bool {
    let node = ast::skip_parentheses(node);
    node.kind() == Kind::MetaProperty
        && node.as_meta_property().keyword_token() == Kind::ImportKeyword
        && node.as_meta_property().name().text() == "meta"
}
pub(crate) fn import_meta(node: P<Node>) -> bool {
    let node = ast::skip_parentheses(node);
    match node.kind() {
        Kind::PropertyAccessExpression => {
            let access = node.as_property_access_expression();
            access.name().text() == "rstest" && is_import_meta(access.expression())
        }
        Kind::ElementAccessExpression => {
            let access = node.as_element_access_expression();
            string(ast::skip_parentheses(access.argument_expression())).as_deref() == Some("rstest")
                && is_import_meta(access.expression())
        }
        _ => false,
    }
}
fn chain(node: P<Node>) -> Option<Chain> {
    let node = skip(node);
    if node.kind() == Kind::Identifier || import_meta(node) {
        return Some(Chain {
            root: node,
            parts: vec![],
            invoked: false,
        });
    }
    match node.kind() {
        Kind::PropertyAccessExpression | Kind::ElementAccessExpression => {
            let (receiver, name) = if node.kind() == Kind::PropertyAccessExpression {
                let access = node.as_property_access_expression();
                (access.expression(), access.name())
            } else {
                let access = node.as_element_access_expression();
                (
                    access.expression(),
                    ast::skip_parentheses(access.argument_expression()),
                )
            };
            let mut chain = chain(receiver)?;
            let value = if node.kind() == Kind::ElementAccessExpression {
                string(name)?
            } else {
                property(name)?
            };
            chain.parts.push(Part {
                name: value,
                node: name,
                invoked: 0,
                call: None,
            });
            Some(chain)
        }
        Kind::CallExpression | Kind::TaggedTemplateExpression => {
            let receiver = if node.kind() == Kind::CallExpression {
                node.as_call_expression().expression()
            } else {
                node.as_tagged_template_expression().tag()
            };
            let mut chain = chain(receiver)?;
            if let Some(last) = chain.parts.last_mut() {
                if last.invoked != 0 {
                    return None;
                }
                last.invoked = if node.kind() == Kind::CallExpression {
                    1
                } else {
                    2
                };
                last.call = (node.kind() == Kind::CallExpression).then_some(node);
            } else if chain.invoked || node.kind() == Kind::TaggedTemplateExpression {
                return None;
            } else {
                chain.invoked = true;
            }
            Some(chain)
        }
        Kind::BinaryExpression
            if node.as_binary_expression().operator_token().kind() == Kind::CommaToken =>
        {
            chain(node.as_binary_expression().right())
        }
        _ => None,
    }
}
fn import_declaration(mut node: P<Node>) -> Option<P<Node>> {
    loop {
        if matches!(
            node.kind(),
            Kind::ImportDeclaration | Kind::JSImportDeclaration
        ) {
            return Some(node);
        }
        node = node.parent()?;
    }
}
fn enclosing_variable(mut node: P<Node>) -> Option<P<Node>> {
    loop {
        if node.kind() == Kind::VariableDeclaration {
            return Some(node);
        }
        node = node.parent()?;
    }
}
fn profile(module: &str) -> Option<bool> {
    match module {
        "@rstest/core" | "rstack/test" => Some(false),
        "@rstest/playwright" => Some(true),
        _ => None,
    }
}
// Returns the direct export (or an empty export for a namespace) and profile.
pub(crate) fn imported(declaration: P<Node>) -> Option<(String, bool)> {
    if ast::is_type_only_import_declaration(declaration) {
        return None;
    }
    match declaration.kind() {
        Kind::ImportSpecifier | Kind::NamespaceImport => {
            let module = referenced_module(import_declaration(declaration)?)?;
            let profile = profile(&module)?;
            let name = if declaration.kind() == Kind::NamespaceImport {
                String::new()
            } else {
                let specifier = declaration.as_import_specifier();
                specifier
                    .property_name()
                    .unwrap_or(specifier.name())
                    .text()
                    .to_string()
            };
            Some((name, profile))
        }
        Kind::ImportEqualsDeclaration => {
            Some((String::new(), profile(&referenced_module(declaration)?)?))
        }
        Kind::VariableDeclaration => Some((
            String::new(),
            profile(&require_module(declaration.initializer()?, true)?)?,
        )),
        Kind::BindingElement => {
            let variable = enclosing_variable(declaration)?;
            let initializer = variable.initializer()?;
            let binding = declaration.as_binding_element();
            if import_meta(initializer) {
                return Some((
                    property(binding.property_name().or(binding.name())?)?,
                    false,
                ));
            }
            let profile = profile(&require_module(initializer, true)?)?;
            // The frozen require-binding resolver accepts identifier property
            // names only, then falls back to the local binding name. In
            // particular, { ["test"]: run } does not resolve to `test`, while
            // import.meta.rstest intentionally uses static property names.
            let name = binding
                .property_name()
                .filter(|node| matches!(node.kind(), Kind::Identifier | Kind::PrivateIdentifier))
                .or_else(|| {
                    binding.name().filter(|node| {
                        matches!(node.kind(), Kind::Identifier | Kind::PrivateIdentifier)
                    })
                })?;
            Some((name.text().to_string(), profile))
        }
        _ => None,
    }
}
#[derive(Clone)]
pub(crate) struct Registration {
    pub(crate) name: String,
    pub(crate) state: u8,
    playwright: bool,
    skip: bool,
    pub(crate) todo: bool,
    pub(crate) execution_mode: u8,
    pub(crate) parameterized: u8,
    focus: Vec<Part>,
    root: P<Node>,
    site: Vec<Part>,
}
impl Registration {
    pub(crate) fn is_test(&self) -> bool {
        matches!(self.state, 1 | 2 | 4)
    }
    fn apply(&mut self, part: &Part) -> bool {
        self.state = match self.state {
            1 | 2 => match (part.name.as_str(), part.invoked) {
                ("only" | "skip" | "todo" | "fails" | "concurrent" | "sequential", 0) => 1,
                ("fail", 0) if self.playwright => 1,
                ("runIf" | "skipIf", 1) => 1,
                ("each" | "for", 1 | 2) => 4,
                ("extend", 1) if self.state == 2 => 2,
                ("describe", 0) if self.playwright => 3,
                _ => 0,
            },
            3 => match (part.name.as_str(), part.invoked) {
                ("only" | "skip" | "todo" | "concurrent" | "sequential", 0) => 3,
                ("runIf" | "skipIf", 1) => 3,
                ("each" | "for", 1 | 2) => 5,
                _ => 0,
            },
            _ => 0,
        };
        if self.state == 0 {
            return false;
        }
        match part.name.as_str() {
            "skip" => self.skip = true,
            "todo" => self.todo = true,
            "concurrent" => self.execution_mode = 1,
            "sequential" => self.execution_mode = 2,
            "only" => self.focus.push(part.clone()),
            "each" => self.parameterized = 1,
            "for" => self.parameterized = 2,
            _ => {}
        }
        true
    }
}
fn direct(name: &str, playwright: bool, root: P<Node>) -> Option<Registration> {
    let state = match name {
        "test" => 2,
        "it" if !playwright => 2,
        "describe" => 3,
        _ => return None,
    };
    Some(Registration {
        name: name.to_string(),
        state,
        playwright,
        skip: false,
        todo: false,
        execution_mode: 0,
        parameterized: 0,
        focus: vec![],
        root,
        site: vec![],
    })
}
fn resolved(
    chain: &Chain,
    file: P<SourceFile>,
    checker: &mut Checker,
    depth: usize,
) -> Option<Registration> {
    if depth > 16 || chain.invoked {
        return None;
    }
    let root = chain.root;
    let mut consumed = 0;
    let mut resolved = None;
    if import_meta(root) {
        let first = chain.parts.first()?;
        if first.invoked != 0 {
            return None;
        }
        consumed = 1;
        resolved = direct(&first.name, false, root);
    } else {
        let symbol = checker.get_symbol_at_location_exported(root);
        let declarations = symbol
            .map(|symbol| symbol.declarations().to_vec())
            .unwrap_or_default();
        let mut local = false;
        for &declaration in &declarations {
            if ast::is_type_only_import_declaration(declaration) {
                continue;
            }
            local |= ast::get_source_file_of_node(declaration) == Some(file);
            // The frozen registration path does not recognize import-equals
            // namespace roots (confirmed through the configured public API).
            // Keep this registration boundary separate from expect's namespace
            // resolver, which also uses `imported` below.
            if declaration.kind() == Kind::ImportEqualsDeclaration {
                continue;
            }
            if let Some((name, playwright)) = imported(declaration) {
                if name.is_empty() {
                    let first = chain.parts.first()?;
                    if first.invoked != 0 {
                        return None;
                    }
                    consumed = 1;
                    resolved = direct(&first.name, playwright, root);
                } else {
                    resolved = direct(&name, playwright, root);
                }
                if resolved.is_some() {
                    break;
                }
            }
        }
        if resolved.is_none() && !local {
            resolved = direct(root.text(), false, root);
        }
        if resolved.is_none() {
            for declaration in declarations {
                if declaration.kind() != Kind::VariableDeclaration
                    || !declaration.parent().is_some_and(|parent| {
                        parent.kind() == Kind::VariableDeclarationList
                            && parent.flags().intersects(NodeFlags::Const)
                    })
                {
                    continue;
                }
                let Some(initializer) = declaration.initializer() else {
                    continue;
                };
                if import_meta(initializer) {
                    let first = chain.parts.first()?;
                    if first.invoked != 0 {
                        return None;
                    }
                    resolved = direct(&first.name, false, root);
                    consumed = 1;
                } else if let Some(alias) = self::chain(initializer) {
                    resolved = self::resolved(&alias, file, checker, depth + 1);
                }
                if resolved.is_some() {
                    break;
                }
            }
        }
    }
    let mut result = resolved?;
    for part in &chain.parts[consumed..] {
        if !result.apply(part) {
            return None;
        }
    }
    result.root = root;
    result.site = chain.parts[consumed..].to_vec();
    Some(result)
}
pub(crate) fn registration(
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
) -> Option<Registration> {
    if node.kind() != Kind::CallExpression {
        return None;
    }
    resolved(
        &chain(node.as_call_expression().expression())?,
        file,
        checker,
        0,
    )
}
fn token_ranges(file: P<SourceFile>, begin: usize, finish: usize) -> (bool, Vec<[usize; 2]>) {
    let mut scanner = Scanner::new();
    scanner.set_text(file.text());
    scanner.set_skip_trivia(false);
    scanner.reset_pos(begin as u32);
    let mut comment = false;
    let mut ranges = vec![];
    loop {
        let token = scanner.scan();
        let start = scanner.token_start() as usize;
        let end = scanner.token_end() as usize;
        if token == Kind::EndOfFile || start >= finish {
            break;
        }
        if matches!(
            token,
            Kind::SingleLineCommentTrivia | Kind::MultiLineCommentTrivia
        ) {
            comment = true;
        } else if token > Kind::LastTriviaToken && end <= finish {
            ranges.push([start, end]);
        }
    }
    (comment, ranges)
}
fn syntax_ranges(file: P<SourceFile>, begin: usize, finish: usize) -> Vec<[usize; 2]> {
    let (comment, tokens) = token_ranges(file, begin, finish);
    if comment {
        tokens
    } else {
        vec![[begin, finish]]
    }
}
fn focus_edits(parts: &[Part], file: P<SourceFile>) -> Option<Vec<Edit>> {
    let mut ranges = vec![];
    for part in parts {
        let mut accessor = part.node.parent()?;
        while accessor.kind() == Kind::ParenthesizedExpression {
            accessor = accessor.parent()?;
        }
        let (receiver, question) = match accessor.kind() {
            Kind::PropertyAccessExpression => {
                let access = accessor.as_property_access_expression();
                (access.expression(), access.question_dot_token())
            }
            Kind::ElementAccessExpression => {
                let access = accessor.as_element_access_expression();
                (access.expression(), access.question_dot_token())
            }
            _ => return None,
        };
        ranges.extend(syntax_ranges(
            file,
            question.map_or(end(receiver), end),
            end(accessor),
        ));
        let Some(question) = question else {
            continue;
        };
        let mut next = accessor.parent()?;
        while next.kind() == Kind::ParenthesizedExpression {
            next = next.parent()?;
        }
        match next.kind() {
            Kind::PropertyAccessExpression => ranges.extend(syntax_ranges(
                file,
                end(accessor),
                start(next.as_property_access_expression().name(), file),
            )),
            Kind::ElementAccessExpression => {
                if let Some(question) = next.as_element_access_expression().question_dot_token() {
                    ranges.extend(syntax_ranges(file, end(accessor), end(question)));
                }
            }
            Kind::CallExpression => {
                if let Some(question) = next.as_call_expression().question_dot_token() {
                    ranges.extend(syntax_ranges(file, end(accessor), end(question)));
                }
            }
            Kind::VariableDeclaration => ranges.push([start(question, file), end(question)]),
            _ => return None,
        }
    }
    ranges.sort_unstable();
    if ranges.is_empty() || ranges.windows(2).any(|pair| pair[0][1] > pair[1][0]) {
        return None;
    }
    Some(ranges.into_iter().map(Edit::remove).collect())
}

pub(crate) fn top_call(node: P<Node>) -> P<Node> {
    let mut current = node;
    let mut top = node;
    while let Some(parent) = current.parent() {
        match parent.kind() {
            Kind::ParenthesizedExpression => {}
            Kind::CallExpression if parent.as_call_expression().expression() == current => {
                top = parent
            }
            Kind::PropertyAccessExpression
                if parent.as_property_access_expression().expression() == current => {}
            Kind::ElementAccessExpression
                if parent.as_element_access_expression().expression() == current => {}
            _ => break,
        }
        current = parent;
    }
    top
}
fn written(node: P<Node>) -> bool {
    let mut current = node;
    while let Some(parent) = current.parent() {
        if matches!(
            parent.kind(),
            Kind::ParenthesizedExpression
                | Kind::NonNullExpression
                | Kind::AsExpression
                | Kind::SatisfiesExpression
                | Kind::TypeAssertionExpression
        ) {
            current = parent;
            continue;
        }
        return match parent.kind() {
            Kind::BinaryExpression => {
                let binary = parent.as_binary_expression();
                binary.left() == current
                    && binary.operator_token().kind() >= Kind::FirstAssignment
                    && binary.operator_token().kind() <= Kind::LastAssignment
            }
            Kind::PrefixUnaryExpression => matches!(
                parent.as_prefix_unary_expression().operator(),
                Kind::PlusPlusToken | Kind::MinusMinusToken
            ),
            Kind::PostfixUnaryExpression => true,
            Kind::ForInStatement | Kind::ForOfStatement => {
                parent.as_for_in_or_of_statement().initializer() == current
            }
            Kind::ArrayLiteralExpression
            | Kind::ObjectLiteralExpression
            | Kind::PropertyAssignment
            | Kind::ShorthandPropertyAssignment
            | Kind::SpreadAssignment
            | Kind::SpreadElement => {
                current = parent;
                continue;
            }
            _ => false,
        };
    }
    false
}
pub(crate) fn symbol_written(
    symbol: Option<P<Symbol>>,
    name: &str,
    nodes: &[P<Node>],
    checker: &mut Checker,
) -> bool {
    nodes.iter().any(|&node| {
        node.kind() == Kind::Identifier
            && node.text() == name
            && written(node)
            && checker.get_symbol_at_location_exported(node) == symbol
    })
}
pub(crate) fn callback(
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    nodes: &[P<Node>],
) -> Option<P<Node>> {
    let node = skip(node);
    if matches!(node.kind(), Kind::FunctionExpression | Kind::ArrowFunction) {
        return Some(node);
    }
    if node.kind() != Kind::Identifier {
        return None;
    }
    let symbol = checker.get_symbol_at_location_exported(node)?;
    if symbol_written(Some(symbol), node.text(), nodes, checker) {
        return None;
    }
    let mut found = None;
    for &declaration in symbol.declarations() {
        if ast::get_source_file_of_node(declaration) != Some(file) {
            return None;
        }
        let candidate = match declaration.kind() {
            Kind::FunctionDeclaration if declaration.body().is_none() => continue,
            Kind::FunctionDeclaration => declaration,
            Kind::VariableDeclaration => {
                let initializer = skip(declaration.initializer()?);
                if !matches!(
                    initializer.kind(),
                    Kind::FunctionExpression | Kind::ArrowFunction
                ) {
                    return None;
                }
                initializer
            }
            _ => return None,
        };
        if found.replace(candidate).is_some() {
            return None;
        }
    }
    found
}
pub(crate) fn context_bindings(
    nodes: &[P<Node>],
    file: P<SourceFile>,
    checker: &mut Checker,
) -> Vec<(P<Symbol>, bool)> {
    let mut result = vec![];
    for &call in nodes {
        if call.kind() != Kind::CallExpression {
            continue;
        }
        let Some(parsed) = registration(call, file, checker) else {
            continue;
        };
        if !parsed.is_test() || parsed.parameterized == 1 {
            continue;
        }
        let args = call.arguments();
        if args.len() < 2 {
            continue;
        }
        let function = args
            .get(2)
            .and_then(|&arg| callback(arg, file, checker, nodes))
            .or_else(|| callback(args[1], file, checker, nodes));
        let Some(function) = function else {
            continue;
        };
        let index = usize::from(parsed.parameterized == 2);
        let Some(parameter) = function.parameters().get(index) else {
            continue;
        };
        let name = parameter.as_parameter_declaration().name();
        if name.kind() == Kind::Identifier {
            if let Some(symbol) = checker.get_symbol_at_location_exported(name) {
                result.push((symbol, true));
            }
        } else if name.kind() == Kind::ObjectBindingPattern {
            for &element in name.as_binding_pattern().elements().nodes() {
                if element.kind() != Kind::BindingElement {
                    continue;
                }
                let element = element.as_binding_element();
                let Some(name) = element.name() else {
                    continue;
                };
                if name.kind() == Kind::Identifier
                    && property(element.property_name().unwrap_or(name)).as_deref()
                        == Some("expect")
                    && let Some(symbol) = checker.get_symbol_at_location_exported(name)
                {
                    result.push((symbol, false));
                }
            }
        }
    }
    result
}
// Unlike registration parsing, expectation chains intentionally stop at TS
// assertions. Dynamic identifier keys remain entries so a later unknown member
// can truncate a chain without losing an earlier snapshot matcher.
pub(crate) fn expect_entries(node: P<Node>, entries: &mut Vec<Part>) {
    let node = ast::skip_parentheses(node);
    match node.kind() {
        Kind::Identifier => entries.push(Part {
            name: node.text().to_string(),
            node,
            invoked: 0,
            call: None,
        }),
        Kind::PropertyAccessExpression | Kind::ElementAccessExpression => {
            let (receiver, name) = if node.kind() == Kind::PropertyAccessExpression {
                let access = node.as_property_access_expression();
                (access.expression(), access.name())
            } else {
                let access = node.as_element_access_expression();
                (
                    access.expression(),
                    ast::skip_parentheses(access.argument_expression()),
                )
            };
            expect_entries(receiver, entries);
            if let Some(name_text) = property(name) {
                entries.push(Part {
                    name: name_text,
                    node: name,
                    invoked: 0,
                    call: None,
                });
            } else if node.kind() == Kind::ElementAccessExpression {
                entries.clear();
            }
        }
        Kind::CallExpression => {
            expect_entries(node.as_call_expression().expression(), entries);
            if let Some(last) = entries.last_mut() {
                last.call = Some(node);
                last.invoked = 1;
            }
        }
        Kind::TaggedTemplateExpression => {
            expect_entries(node.as_tagged_template_expression().tag(), entries)
        }
        _ => {}
    }
}
pub(crate) fn expect_root(
    entries: &[Part],
    file: P<SourceFile>,
    checker: &mut Checker,
    contexts: &[(P<Symbol>, bool)],
    nodes: &[P<Node>],
) -> Option<usize> {
    let first = entries.first()?;
    let root = first.node;
    if root.kind() != Kind::Identifier {
        return None;
    }
    if first.name == "rstest" && root.parent().is_some_and(import_meta) {
        return entries
            .get(1)
            .filter(|part| part.name == "expect")
            .map(|_| 1);
    }
    let symbol = checker.get_symbol_at_location_exported(root);
    if symbol_written(symbol, root.text(), nodes, checker) {
        return None;
    }
    let declarations = symbol
        .map(|symbol| symbol.declarations().to_vec())
        .unwrap_or_default();
    let initialized = declarations
        .iter()
        .filter(|declaration| match declaration.kind() {
            Kind::VariableDeclaration | Kind::BindingElement => declaration.initializer().is_some(),
            Kind::Parameter
            | Kind::ImportSpecifier
            | Kind::ImportClause
            | Kind::NamespaceImport
            | Kind::FunctionDeclaration => true,
            _ => false,
        })
        .count();
    if initialized > 1 {
        return None;
    }
    if let Some(symbol) = symbol
        && let Some((_, receiver)) = contexts.iter().find(|(binding, _)| *binding == symbol)
    {
        return if *receiver {
            entries
                .get(1)
                .filter(|part| part.name == "expect")
                .map(|_| 1)
        } else {
            Some(0)
        };
    }
    let mut local = false;
    for declaration in declarations {
        if ast::is_type_only_import_declaration(declaration) {
            continue;
        }
        local |= ast::get_source_file_of_node(declaration) == Some(file);
        if declaration.kind() == Kind::VariableDeclaration
            && declaration.initializer().is_some_and(import_meta)
        {
            return entries
                .get(1)
                .filter(|part| part.name == "expect")
                .map(|_| 1);
        }
        if let Some((name, _)) = imported(declaration) {
            if name == "expect" {
                return Some(0);
            }
            if name.is_empty() {
                return entries
                    .get(1)
                    .filter(|part| part.name == "expect")
                    .map(|_| 1);
            }
        }
    }
    (!local && first.name == "expect").then_some(0)
}
fn dynamic_member(node: P<Node>) -> bool {
    if node.kind() != Kind::Identifier {
        return false;
    }
    let mut current = node;
    while let Some(parent) = current.parent() {
        if parent.kind() == Kind::ParenthesizedExpression {
            current = parent;
            continue;
        }
        return parent.kind() == Kind::ElementAccessExpression
            && parent.as_element_access_expression().argument_expression() == current;
    }
    false
}
fn language(name: &str) -> bool {
    [
        "also", "and", "at", "be", "been", "but", "does", "has", "have", "is", "itself", "of",
        "same", "still", "that", "to", "which", "with", "all", "any", "deep", "nested", "ordered",
        "own",
    ]
    .contains(&name)
}
fn dual(name: &str) -> bool {
    [
        "a",
        "an",
        "contain",
        "contains",
        "include",
        "includes",
        "length",
        "lengthOf",
        "respondTo",
    ]
    .contains(&name)
}
fn getter(name: &str) -> bool {
    [
        "Arguments",
        "NaN",
        "arguments",
        "callable",
        "called",
        "calledOnce",
        "calledThrice",
        "calledTwice",
        "empty",
        "exist",
        "exists",
        "extensible",
        "false",
        "finite",
        "frozen",
        "iterable",
        "null",
        "numeric",
        "ok",
        "sealed",
        "true",
        "undefined",
    ]
    .contains(&name)
}
fn modifier(name: &str) -> bool {
    ["not", "rejects", "resolves"].contains(&name)
}
// 0 unknown, 1 matcher, 2 language, 3 modifier.
pub(crate) fn expect_kind(part: &Part, first_static: bool) -> u8 {
    if dynamic_member(part.node) {
        return 0;
    }
    if part.call.is_some() {
        if first_static && part.name == "any" {
            return 1;
        }
        if getter(&part.name) || modifier(&part.name) || language(&part.name) {
            return 2;
        }
        return 1;
    }
    if getter(&part.name) {
        1
    } else if language(&part.name) || dual(&part.name) {
        2
    } else if modifier(&part.name) {
        3
    } else {
        0
    }
}
fn interpolations(
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    contexts: &[(P<Symbol>, bool)],
    nodes: &[P<Node>],
    has_type_information: bool,
) -> Vec<Violation> {
    let mut entries = vec![];
    expect_entries(node, &mut entries);
    let Some(root) = expect_root(&entries, file, checker, contexts, nodes) else {
        return vec![];
    };
    let mut first = root + 1;
    let root_invoked = entries[root].call.is_some();
    let mut is_static = !root_invoked;
    if !root_invoked
        && entries.get(first).is_some_and(|part| {
            part.call.is_some() && ["soft", "poll", "element"].contains(&part.name.as_str())
        })
    {
        first += 1;
        is_static = false;
    }
    let rest = &entries[first..];
    let kinds: Vec<_> = rest
        .iter()
        .enumerate()
        .map(|(index, part)| expect_kind(part, is_static && index == 0))
        .collect();
    let Some(matcher) = kinds.iter().position(|&kind| kind == 1) else {
        return vec![];
    };
    let mut not = 0;
    let mut promise = 0;
    for (part, kind) in rest[..matcher].iter().zip(&kinds[..matcher]) {
        match kind {
            2 if part.call.is_none() => {}
            3 => {
                if part.name == "not" {
                    not += 1;
                } else {
                    promise += 1;
                }
                if not > 1 || promise > 1 {
                    return vec![];
                }
            }
            _ => return vec![],
        }
    }
    let mut reports = vec![];
    for (part, kind) in rest[matcher..].iter().zip(&kinds[matcher..]) {
        if *kind == 0 {
            break;
        }
        if *kind == 2 && part.call.is_some() {
            return vec![];
        }
        if *kind != 1
            || ![
                "toMatchInlineSnapshot",
                "toThrowErrorMatchingInlineSnapshot",
            ]
            .contains(&part.name.as_str())
        {
            continue;
        }
        let Some(call) = part.call else {
            continue;
        };
        let args = call.arguments();
        let Some(&first) = args.first() else {
            continue;
        };
        let selected = if part.name == "toThrowErrorMatchingInlineSnapshot"
            || args.len() == 1
            || is_string_snapshot(first, checker, has_type_information)
        {
            first
        } else {
            args[1]
        };
        let selected = ast::skip_parentheses(selected);
        if selected.kind() == Kind::TemplateExpression {
            reports.push(Violation::new(
                selected,
                "noInterpolation",
                "Do not use string interpolation inside of snapshots",
            ));
        }
    }
    reports
}
fn is_string_snapshot(node: P<Node>, checker: &mut Checker, has_type_information: bool) -> bool {
    let node = ast::skip_parentheses(node);
    if matches!(
        node.kind(),
        Kind::StringLiteral | Kind::NoSubstitutionTemplateLiteral | Kind::TemplateExpression
    ) {
        return true;
    }
    if !has_type_information {
        return false;
    }
    let ty = checker.get_type_at_location(node);
    let ty = checker.get_base_constraint_of_type(ty).unwrap_or(ty);
    if ty.flags().intersects(TypeFlags::Union) {
        ty.types()
            .iter()
            .all(|part| part.flags().intersects(TypeFlags::StringLike))
    } else {
        ty.flags().intersects(TypeFlags::StringLike)
    }
}
