//! Declaration rules with the defaults enabled by the frozen Rsbuild workload.
//! Diagnostic and edit semantics follow rslint e0a29f7, including declaration
//! merging and source-preserving rewrites; non-default options are rejected by
//! the request validator rather than silently treated as defaults.

use std::collections::BTreeMap;

use crate::{Suggestion, Violation, edits::Edit};
use tsrs_ast::{
    Kind, ModifierFlags, Node, SourceFile, Symbol, has_syntactic_modifier,
    is_assignment_expression, is_global_scope_augmentation, skip_parentheses,
};
use tsrs_checker::Checker;
use tsrs_core::P;
use tsrs_scanner::{get_range_of_token_at_position, skip_trivia};

pub(crate) const NAMES: &[&str] = &[
    "@typescript-eslint/no-empty-object-type",
    "@typescript-eslint/no-namespace",
    "@typescript-eslint/no-this-alias",
    "@typescript-eslint/no-array-constructor",
];

pub(crate) fn check(
    name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    has_type_information: bool,
) -> Vec<Violation> {
    let report = match name {
        "@typescript-eslint/no-empty-object-type" => {
            empty_object_type(node, file, checker, has_type_information)
        }
        "@typescript-eslint/no-namespace" => namespace(node, file),
        "@typescript-eslint/no-this-alias" => this_alias(node, file),
        "@typescript-eslint/no-array-constructor" => array_constructor(node, file),
        _ => None,
    };
    report.into_iter().collect()
}

fn end(node: P<Node>) -> usize {
    node.end().expect_source("parsed node has source end") as usize
}

fn range(node: P<Node>, file: P<SourceFile>) -> [usize; 2] {
    [
        skip_trivia(
            file.text(),
            node.pos().expect_source("parsed node has source start"),
        ) as usize,
        end(node),
    ]
}

fn text(node: P<Node>, file: P<SourceFile>) -> &'static str {
    let [start, end] = range(node, file);
    &file.text()[start..end]
}

fn other_declaration(symbol: Option<P<Symbol>>, node: P<Node>) -> bool {
    symbol.is_some_and(|symbol| {
        symbol.declarations().iter().any(|&declaration| {
            declaration != node
                && matches!(
                    declaration.kind(),
                    Kind::InterfaceDeclaration | Kind::ClassDeclaration
                )
        })
    })
}

fn may_replace_interface(node: P<Node>, checker: &mut Checker, has_type_information: bool) -> bool {
    if has_syntactic_modifier(node, ModifierFlags::Export)
        && has_syntactic_modifier(node, ModifierFlags::Default)
    {
        return false;
    }
    let declaration = node.as_interface_declaration();
    !other_declaration(declaration.symbol(), node)
        && (!has_type_information
            || !other_declaration(
                checker.get_symbol_at_location_exported(declaration.name()),
                node,
            ))
}

fn empty_message(interface: bool) -> String {
    let (intro, option) = if interface {
        ("An empty interface declaration", "allowInterfaces")
    } else {
        ("The `{}` (\"empty object\") type", "allowObjectTypes")
    };
    format!(
        "{intro} allows any non-nullish value, including literals like `0` and `\"\"`.\n\
- If that's what you want, disable this lint rule with an inline comment or configure the '{option}' rule option.\n\
- If you want a type meaning \"any object\", you probably want `object` instead.\n\
- If you want a type meaning \"any value\", you probably want `unknown` instead."
    )
}

fn empty_object_type(
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    has_type_information: bool,
) -> Option<Violation> {
    if node.kind() == Kind::TypeLiteral {
        if !node.members().is_empty()
            || node
                .parent()
                .is_some_and(|parent| parent.kind() == Kind::IntersectionType)
        {
            return None;
        }
        let mut report = Violation::new(node, "noEmptyObject", empty_message(false));
        for replacement in ["object", "unknown"] {
            report.suggestions.push(Suggestion {
                message_id: "replaceEmptyObjectType",
                message: format!("Replace `{{}}` with `{replacement}`."),
                data: Some(BTreeMap::from([(
                    "replacement".to_owned(),
                    replacement.to_owned(),
                )])),
                edits: vec![Edit::replace(range(node, file), replacement)],
            });
        }
        return Some(report);
    }
    if node.kind() != Kind::InterfaceDeclaration || !node.members().is_empty() {
        return None;
    }
    let declaration = node.as_interface_declaration();
    let extended = declaration
        .heritage_clauses()
        .and_then(|clauses| {
            clauses.nodes().iter().find_map(|clause| {
                let heritage = clause.as_heritage_clause();
                (heritage.token() == Kind::ExtendsKeyword).then(|| heritage.types().nodes())
            })
        })
        .unwrap_or_default();
    if extended.len() > 1 {
        return None;
    }
    let name = declaration.name();
    let mut report = if extended.is_empty() {
        Violation::new(name, "noEmptyInterface", empty_message(true))
    } else {
        Violation::new(
            name,
            "noEmptyInterfaceWithSuper",
            "An interface declaring no members is equivalent to its supertype.",
        )
    };
    if !may_replace_interface(node, checker, has_type_information) {
        return Some(report);
    }
    let parameters = declaration
        .type_parameters()
        .filter(|list| !list.nodes().is_empty())
        .map(|list| {
            let left = get_range_of_token_at_position(file, end(name) as u32);
            let right = get_range_of_token_at_position(
                file,
                list.end()
                    .expect_source("parsed type parameters have source end"),
            );
            &file.text()[left.pos().expect_source("token has source start") as usize
                ..right.end().expect_source("token has source end") as usize]
        })
        .unwrap_or("");
    let start = declaration
        .modifiers()
        .and_then(|modifiers| modifiers.nodes().last().copied())
        .map_or_else(
            || range(node, file)[0],
            |modifier| {
                get_range_of_token_at_position(file, end(modifier) as u32)
                    .pos()
                    .expect_source("interface token has source start") as usize
            },
        );
    let fix_range = [start, end(node)];
    if let Some(&base) = extended.first() {
        report = report.with_suggestion(
            "replaceEmptyInterfaceWithSuper",
            "Replace empty interface with a type alias.",
            vec![Edit::replace(
                fix_range,
                format!(
                    "type {}{parameters} = {}",
                    text(name, file),
                    text(base, file)
                ),
            )],
        );
    } else {
        for replacement in ["object", "unknown"] {
            report.suggestions.push(Suggestion {
                message_id: "replaceEmptyInterface",
                message: format!("Replace empty interface with `{replacement}`."),
                data: Some(BTreeMap::from([(
                    "replacement".to_owned(),
                    replacement.to_owned(),
                )])),
                edits: vec![Edit::replace(
                    fix_range,
                    format!("type {}{parameters} = {replacement}", text(name, file)),
                )],
            });
        }
    }
    Some(report)
}

fn namespace(node: P<Node>, file: P<SourceFile>) -> Option<Violation> {
    if node.kind() != Kind::ModuleDeclaration || file.is_declaration_file() {
        return None;
    }
    let declaration = node.as_module_declaration();
    if declaration.name().kind() == Kind::StringLiteral
        || is_global_scope_augmentation(node)
        || node
            .parent()
            .is_some_and(|parent| parent.kind() == Kind::ModuleDeclaration)
    {
        return None;
    }
    Some(Violation::new(
        node,
        "moduleSyntaxIsPreferred",
        "ES2015 module syntax is preferred over namespaces.",
    ))
}

fn this_alias(node: P<Node>, file: P<SourceFile>) -> Option<Violation> {
    if node.kind() != Kind::ThisKeyword {
        return None;
    }
    let mut value = node;
    let mut parent = value.parent()?;
    while parent.kind() == Kind::ParenthesizedExpression {
        value = parent;
        parent = value.parent()?;
    }
    let target = match parent.kind() {
        Kind::VariableDeclaration
            if parent.as_variable_declaration().initializer() == Some(value) =>
        {
            parent.as_variable_declaration().name()
        }
        Kind::BinaryExpression
            if is_assignment_expression(parent, false)
                && parent.as_binary_expression().right() == value =>
        {
            skip_parentheses(parent.as_binary_expression().left())
        }
        _ => return None,
    };
    // Default allowDestructuring is true; property targets are also left alone.
    if target.kind() != Kind::Identifier {
        return None;
    }
    let mut report_range = range(target, file);
    if let Some(declaration) = target
        .parent()
        .filter(|node| node.kind() == Kind::VariableDeclaration)
    {
        let declaration = declaration.as_variable_declaration();
        if declaration.name() == target {
            for suffix in [declaration.exclamation_token(), declaration.type_()]
                .into_iter()
                .flatten()
            {
                report_range[1] = report_range[1].max(end(suffix));
            }
        }
    }
    Some(
        Violation::new(
            target,
            "thisAssignment",
            "Unexpected aliasing of 'this' to local variable.",
        )
        .with_byte_range(report_range),
    )
}

fn array_constructor(node: P<Node>, file: P<SourceFile>) -> Option<Violation> {
    let (callee, arguments) = match node.kind() {
        Kind::CallExpression => {
            let call = node.as_call_expression();
            (call.expression(), Some(call.arguments()))
        }
        Kind::NewExpression => {
            let call = node.as_new_expression();
            (call.expression(), call.arguments())
        }
        _ => return None,
    };
    if arguments.is_some_and(|args| args.nodes().len() == 1) || !node.type_arguments().is_empty() {
        return None;
    }
    let callee = skip_parentheses(callee);
    // The frozen rule is intentionally syntax-only: a local Array binding is
    // still reported, but member calls and TypeScript wrappers are not unwrapped.
    if callee.kind() != Kind::Identifier || callee.text() != "Array" {
        return None;
    }
    let report_range = range(node, file);
    let edits = arguments
        .and_then(|args| {
            let start = args.pos().source_offset()? as usize;
            let close = report_range[1].checked_sub(1)?;
            (start >= report_range[0]
                && close >= start
                && file.text().as_bytes().get(close) == Some(&b')'))
            .then(|| {
                vec![
                    Edit::replace([report_range[0], start], "["),
                    Edit::replace([close, report_range[1]], "]"),
                ]
            })
        })
        .unwrap_or_else(|| vec![Edit::replace(report_range, "[]")]);
    Some(
        Violation::new(
            node,
            "useLiteral",
            "The array literal notation [] is preferable.",
        )
        .with_byte_range(report_range)
        .with_fix(edits),
    )
}
