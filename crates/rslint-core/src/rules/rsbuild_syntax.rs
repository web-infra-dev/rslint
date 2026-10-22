//! Structural rules enabled by the frozen Rsbuild configuration.
//! Ported from rslint e0a29f751d4def86d27931759098f27748fa3003.

use std::collections::HashSet;

use tsrs_ast::{self as ast, Kind, Node, NodeFlags, SourceFile};
use tsrs_checker::Checker;
use tsrs_core::P;
use tsrs_scanner::{Scanner, skip_trivia};

use crate::{Violation, edits::Edit};

pub(crate) const NAMES: &[&str] = &[
    "no-case-declarations",
    "no-empty-static-block",
    "no-empty-pattern",
    "no-duplicate-case",
    "require-yield",
];

pub(crate) fn check(
    name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    _checker: &mut Checker,
) -> Vec<Violation> {
    match (name, node.kind()) {
        ("no-case-declarations", Kind::CaseBlock) => case_declarations(node, file),
        ("no-empty-static-block", Kind::ClassStaticBlockDeclaration) => {
            empty_static_block(node, file)
        }
        ("no-empty-pattern", Kind::ObjectBindingPattern | Kind::ArrayBindingPattern)
            if node.as_binding_pattern().elements().nodes().is_empty() =>
        {
            vec![Violation::new(
                node,
                "unexpected",
                if node.kind() == Kind::ObjectBindingPattern {
                    "Unexpected empty object pattern."
                } else {
                    "Unexpected empty array pattern."
                },
            )]
        }
        ("no-duplicate-case", Kind::CaseBlock) => duplicate_cases(node, file),
        (
            "require-yield",
            Kind::FunctionDeclaration | Kind::FunctionExpression | Kind::MethodDeclaration,
        ) => require_yield(node, file),
        _ => Vec::new(),
    }
}

fn pos(node: P<Node>) -> usize {
    node.pos()
        .expect_source("parsed nodes have source positions") as usize
}

fn end(node: P<Node>) -> usize {
    node.end()
        .expect_source("parsed nodes have source positions") as usize
}

fn start(node: P<Node>, file: P<SourceFile>) -> usize {
    skip_trivia(file.text(), pos(node) as u32) as usize
}

fn case_declarations(node: P<Node>, file: P<SourceFile>) -> Vec<Violation> {
    let mut reports = Vec::new();
    for &clause in node.as_case_block().clauses().nodes() {
        let statements = clause.as_case_or_default_clause().statements().nodes();
        let (Some(&first), Some(&last)) = (statements.first(), statements.last()) else {
            continue;
        };
        for &statement in statements {
            let lexical = match statement.kind() {
                Kind::FunctionDeclaration | Kind::ClassDeclaration => true,
                Kind::VariableStatement => statement
                    .as_variable_statement()
                    .declaration_list()
                    .flags()
                    .intersects(NodeFlags::BlockScoped),
                _ => false,
            };
            if lexical {
                reports.push(
                    Violation::new(
                        statement,
                        "unexpected",
                        "Unexpected lexical declaration in case block.",
                    )
                    .with_suggestion(
                        "addBrackets",
                        "Add {} brackets around the case block.",
                        vec![
                            Edit::replace([start(first, file); 2], "{ "),
                            Edit::replace([end(last); 2], " }"),
                        ],
                    ),
                );
            }
        }
    }
    reports
}

fn empty_static_block(node: P<Node>, file: P<SourceFile>) -> Vec<Violation> {
    let body = node.as_class_static_block_declaration().body();
    if !body.statements().is_empty() {
        return Vec::new();
    }
    let body_start = start(body, file);
    let body_end = end(body);
    // The body is already known to have no statements. Scan its interior to
    // distinguish comments from whitespace; comments outside its braces do not
    // exempt this block.
    let mut scanner = Scanner::new();
    scanner.set_text(file.text());
    scanner.set_skip_trivia(false);
    scanner.reset_pos((body_start + 1) as u32);
    loop {
        let token = scanner.scan();
        if scanner.token_start() as usize >= body_end - 1 || token == Kind::EndOfFile {
            break;
        }
        if matches!(
            token,
            Kind::SingleLineCommentTrivia | Kind::MultiLineCommentTrivia
        ) {
            return Vec::new();
        }
    }
    vec![
        Violation::new(body, "unexpected", "Unexpected empty static block.").with_suggestion(
            "suggestComment",
            "Add comment inside empty static block.",
            vec![Edit::replace(
                [body_start + 1, body_end - 1],
                " /* empty */ ",
            )],
        ),
    ]
}

fn duplicate_cases(node: P<Node>, file: P<SourceFile>) -> Vec<Violation> {
    let mut seen = HashSet::new();
    let mut reports = Vec::new();
    for &clause in node.as_case_block().clauses().nodes() {
        if let Some(expression) = clause.as_case_or_default_clause().expression() {
            let mut tokens = Vec::new();
            expression_tokens(expression, file, &mut tokens);
            if !seen.insert(tokens.join(" ")) {
                reports.push(Violation::new(
                    clause,
                    "unexpected",
                    "Duplicate case label.",
                ));
            }
        }
    }
    reports
}

// The AST owns regex, string, template and JSX tokens. Only gaps between AST
// children are scanned, preserving literal contents and lexical state. This is
// the baseline ForEachToken algorithm, not whitespace/comment removal on text.
fn expression_tokens(node: P<Node>, file: P<SourceFile>, tokens: &mut Vec<String>) {
    if ast::is_token_kind(node.kind()) {
        let begin = start(node, file);
        if begin < end(node) {
            tokens.push(file.text()[begin..end(node)].to_owned());
        }
        return;
    }
    let mut position = pos(node);
    node.for_each_child(&mut |child| {
        if !child.flags().intersects(NodeFlags::Reparsed) {
            gap_tokens(file, position, pos(child), tokens);
            expression_tokens(child, file, tokens);
            position = end(child);
        }
        false
    });
    gap_tokens(file, position, end(node), tokens);
}

fn gap_tokens(file: P<SourceFile>, begin: usize, end: usize, tokens: &mut Vec<String>) {
    if begin >= end {
        return;
    }
    let mut scanner = Scanner::new();
    scanner.set_text(file.text());
    scanner.set_language_variant(file.language_variant());
    scanner.reset_pos(begin as u32);
    loop {
        let kind = scanner.scan();
        let start = scanner.token_start() as usize;
        let finish = scanner.token_end() as usize;
        if kind == Kind::EndOfFile || start >= end {
            break;
        }
        if start < finish {
            tokens.push(file.text()[start..finish].to_owned());
        }
        if finish >= end {
            break;
        }
    }
}

fn body_range(node: P<Node>) -> Option<[usize; 2]> {
    match node.kind() {
        Kind::FunctionDeclaration
        | Kind::FunctionExpression
        | Kind::MethodDeclaration
        | Kind::ArrowFunction
        | Kind::GetAccessor
        | Kind::SetAccessor
        | Kind::Constructor => {
            let body = node.body()?;
            let begin = node.parameter_list().map_or_else(
                || pos(body),
                |parameters| {
                    parameters
                        .loc
                        .get()
                        .pos()
                        .expect_source("parsed parameter positions") as usize
                },
            );
            Some([begin, end(body)])
        }
        Kind::PropertyDeclaration => node.initializer().map(|value| [pos(value), end(value)]),
        Kind::ClassStaticBlockDeclaration => {
            let body = node.as_class_static_block_declaration().body();
            Some([pos(body), end(body)])
        }
        _ => None,
    }
}

fn has_own_yield(node: P<Node>, owner: P<Node>) -> bool {
    if node.kind() == Kind::YieldExpression {
        let mut parent = node.parent();
        while let Some(current) = parent {
            if let Some([begin, finish]) = body_range(current)
                && pos(node) >= begin
                && end(node) <= finish
            {
                return current == owner;
            }
            parent = current.parent();
        }
    }
    node.for_each_child(&mut |child| has_own_yield(child, owner))
}

fn require_yield(node: P<Node>, file: P<SourceFile>) -> Vec<Violation> {
    if node
        .body_data()
        .is_none_or(|data| data.asterisk_token().is_none())
    {
        return Vec::new();
    }
    let Some(body) = node.body() else {
        return Vec::new();
    };
    if (body.kind() == Kind::Block && body.statements().is_empty()) || has_own_yield(node, node) {
        return Vec::new();
    }
    // NodeList starts immediately after the actual parameter opening paren,
    // including for generic constraints, computed names and decorator calls.
    let parameters = node
        .parameter_list()
        .expect("parsed function parameter list");
    let open = parameters
        .loc
        .get()
        .pos()
        .expect_source("parsed parameter positions") as usize
        - 1;
    assert_eq!(
        file.text().as_bytes()[open],
        b'(',
        "parameter-list source invariant"
    );
    let begin = match node.kind() {
        Kind::FunctionDeclaration => {
            let mut offset = pos(node);
            if let Some(modifiers) = node.modifiers() {
                for &modifier in modifiers.nodes() {
                    if matches!(modifier.kind(), Kind::ExportKeyword | Kind::DefaultKeyword) {
                        offset = end(modifier);
                    } else {
                        offset = pos(modifier);
                        break;
                    }
                }
            }
            skip_trivia(file.text(), offset as u32) as usize
        }
        Kind::MethodDeclaration => start_without_decorators(node, file),
        Kind::FunctionExpression => {
            if let Some(parent) = node.parent()
                && matches!(
                    parent.kind(),
                    Kind::PropertyAssignment | Kind::PropertyDeclaration
                )
            {
                start_without_decorators(parent, file)
            } else {
                start(node, file)
            }
        }
        _ => unreachable!(),
    };
    vec![
        Violation::new(
            node,
            "missingYield",
            "This generator function does not have 'yield'.",
        )
        .with_byte_range([begin, open]),
    ]
}

fn start_without_decorators(node: P<Node>, file: P<SourceFile>) -> usize {
    let mut offset = pos(node);
    if let Some(modifiers) = node.modifiers() {
        for &modifier in modifiers.nodes() {
            if modifier.kind() == Kind::Decorator {
                offset = offset.max(end(modifier));
            }
        }
    }
    skip_trivia(file.text(), offset as u32) as usize
}
