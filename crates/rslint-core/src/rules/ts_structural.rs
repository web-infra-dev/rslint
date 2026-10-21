//! Optionless TypeScript rules, preserving the baseline's literal and binder
//! semantics. These rules do not offer fixes or suggestions.

use std::collections::HashMap;

use tsrs_ast::{
    InternalSymbolNameDefault, Kind, Node, Symbol, TokenFlags, get_locals,
    is_function_like_or_class_static_block_declaration, is_locals_container,
    is_string_literal_like, skip_parentheses,
};
use tsrs_checker::Checker;
use tsrs_core::{P, jsnum};

use crate::Violation;

pub(crate) const NAMES: &[&str] = &[
    "@typescript-eslint/no-duplicate-enum-values",
    "@typescript-eslint/no-misused-new",
    "@typescript-eslint/no-unsafe-declaration-merging",
];

pub(crate) fn check(rule: &str, node: P<Node>, _checker: &mut Checker) -> Vec<Violation> {
    match (rule, node.kind()) {
        ("@typescript-eslint/no-duplicate-enum-values", Kind::EnumDeclaration) => {
            duplicate_enum_values(node)
        }
        ("@typescript-eslint/no-misused-new", _) => misused_new(node).into_iter().collect(),
        ("@typescript-eslint/no-unsafe-declaration-merging", Kind::SourceFile) => {
            // One invocation per file keeps the baseline's large merged-symbol
            // cache local to a single traversal. No arena handles cross snapshots.
            let mut state = MergeState::default();
            let mut reports = Vec::new();
            state.visit(node, &mut reports);
            reports
        }
        _ => Vec::new(),
    }
}

const STRING: u8 = 1;
const NUMBER: u8 = 2;
const NEGATIVE_NUMBER: u8 = 4;
const NEGATIVE_ZERO: u8 = 8;
const INLINE_ENUM_VALUES: usize = 8;

#[derive(Clone)]
struct EnumValue {
    text: String,
    kind: u8,
}

#[derive(Default)]
struct SeenEnumValues {
    inline: [Option<EnumValue>; INLINE_ENUM_VALUES],
    inline_count: usize,
    value_count: usize,
    overflow: Option<HashMap<String, u8>>,
}

impl SeenEnumValues {
    fn add(&mut self, value: &EnumValue, capacity: usize) -> bool {
        self.value_count += 1;
        if let Some(overflow) = self.overflow.as_mut() {
            let kinds = overflow.entry(value.text.clone()).or_default();
            let duplicate = *kinds & value.kind != 0;
            *kinds |= value.kind;
            return duplicate;
        }
        for entry in self.inline[..self.inline_count].iter_mut().flatten() {
            if entry.text == value.text {
                let duplicate = entry.kind & value.kind != 0;
                entry.kind |= value.kind;
                return duplicate;
            }
        }
        if self.inline_count < INLINE_ENUM_VALUES {
            self.inline[self.inline_count] = Some(value.clone());
            self.inline_count += 1;
            return false;
        }
        let unique_count = self.inline_count + 1;
        let mut overflow =
            HashMap::with_capacity((capacity * unique_count / self.value_count).max(unique_count));
        for entry in self.inline.iter_mut().filter_map(Option::take) {
            overflow.insert(entry.text, entry.kind);
        }
        overflow.insert(value.text.clone(), value.kind);
        self.overflow = Some(overflow);
        false
    }
}

fn enum_value(initializer: P<Node>) -> Option<EnumValue> {
    let initializer = skip_parentheses(initializer);
    match initializer.kind() {
        Kind::NumericLiteral => Some(EnumValue {
            text: initializer.text().into(),
            kind: NUMBER,
        }),
        Kind::StringLiteral => Some(EnumValue {
            text: initializer.text().into(),
            kind: STRING,
        }),
        Kind::NoSubstitutionTemplateLiteral => {
            let literal = initializer.as_no_substitution_template_literal();
            (!literal
                .template_flags()
                .intersects(TokenFlags::ContainsInvalidEscape))
            .then(|| EnumValue {
                text: literal.text().into(),
                kind: STRING,
            })
        }
        Kind::PrefixUnaryExpression => {
            let expression = initializer.as_prefix_unary_expression();
            if !matches!(expression.operator(), Kind::PlusToken | Kind::MinusToken) {
                return None;
            }
            let mut value = enum_value(expression.operand())?;
            if value.kind == STRING {
                let mut number = jsnum::from_string(&value.text);
                if number.is_nan() {
                    return None;
                }
                if expression.operator() == Kind::MinusToken {
                    number = -number;
                }
                return Some(number_enum_value(number));
            }
            if expression.operator() == Kind::PlusToken {
                return Some(value);
            }
            value.kind = match value.kind {
                NEGATIVE_ZERO | NEGATIVE_NUMBER => NUMBER,
                _ if value.text == "0" => NEGATIVE_ZERO,
                _ => NEGATIVE_NUMBER,
            };
            Some(value)
        }
        _ => None,
    }
}

fn number_enum_value(number: jsnum::Number) -> EnumValue {
    if number.0 == 0.0 && number.0.is_sign_negative() {
        EnumValue {
            text: "0".into(),
            kind: NEGATIVE_ZERO,
        }
    } else if number.0 < 0.0 {
        EnumValue {
            text: (-number).string(),
            kind: NEGATIVE_NUMBER,
        }
    } else {
        EnumValue {
            text: number.string(),
            kind: NUMBER,
        }
    }
}

fn duplicate_enum_values(node: P<Node>) -> Vec<Violation> {
    let mut seen = SeenEnumValues::default();
    let mut reports = Vec::new();
    for &member in node.members() {
        let Some(value) = member.as_enum_member().initializer().and_then(enum_value) else {
            continue;
        };
        if seen.add(&value, node.members().len()) {
            let message = format!(
                "Duplicate enum member value {}{}.",
                if value.kind == NEGATIVE_NUMBER {
                    "-"
                } else {
                    ""
                },
                value.text,
            );
            reports.push(Violation::new(member, "duplicateValue", message));
        }
    }
    reports
}

fn member_named(name: Option<P<Node>>, expected: &str) -> bool {
    let Some(name) = name else { return false };
    match name.kind() {
        Kind::Identifier => name.text() == expected,
        Kind::ComputedPropertyName => {
            let expression = name.as_computed_property_name().expression();
            is_string_literal_like(expression) && expression.text() == expected
        }
        _ => false,
    }
}

fn returns_parent_type(type_node: Option<P<Node>>, parent: P<Node>) -> bool {
    let Some(type_node) = type_node.filter(|node| node.kind() == Kind::TypeReference) else {
        return false;
    };
    let name = type_node.as_type_reference_node().type_name();
    name.kind() == Kind::Identifier
        && parent
            .name()
            .is_some_and(|parent_name| name.text() == parent_name.text())
}

fn misused_new(node: P<Node>) -> Option<Violation> {
    match node.kind() {
        Kind::MethodDeclaration => {
            let method = node.as_method_declaration();
            if method.body().is_some() {
                return None;
            }
            let parent = node.parent()?;
            if matches!(
                parent.kind(),
                Kind::ClassDeclaration | Kind::ClassExpression
            ) && member_named(node.name(), "new")
                && returns_parent_type(node.type_node(), parent)
            {
                return Some(Violation::new(
                    node,
                    "errorMessageClass",
                    "Class cannot have method named `new`.",
                ));
            }
        }
        Kind::ConstructSignature => {
            let parent = node.parent()?;
            if parent.kind() == Kind::InterfaceDeclaration
                && returns_parent_type(node.type_node(), parent)
            {
                return Some(Violation::new(
                    node,
                    "errorMessageInterface",
                    "interfaces cannot be constructed, only classes.",
                ));
            }
        }
        Kind::MethodSignature if member_named(node.name(), "constructor") => {
            return Some(Violation::new(
                node,
                "errorMessageInterface",
                "interfaces cannot be constructed, only classes.",
            ));
        }
        _ => {}
    }
    None
}

const CLASS: u8 = 1;
const INTERFACE: u8 = 2;
const DECLARATION_CACHE_THRESHOLD: usize = 8;

fn declaration_kind(node: P<Node>) -> u8 {
    match node.kind() {
        Kind::ClassDeclaration => CLASS,
        Kind::InterfaceDeclaration => INTERFACE,
        _ => 0,
    }
}

fn enclosing_scope(node: P<Node>) -> Option<P<Node>> {
    let mut current = node.parent();
    while let Some(ancestor) = current {
        if is_function_like_or_class_static_block_declaration(ancestor)
            || matches!(ancestor.kind(), Kind::SourceFile | Kind::ModuleBlock)
        {
            return Some(ancestor);
        }
        current = ancestor.parent();
    }
    None
}

fn nearest_locals_container(node: P<Node>) -> Option<P<Node>> {
    let mut current = node.parent();
    while let Some(ancestor) = current {
        if is_locals_container(ancestor) {
            return Some(ancestor);
        }
        current = ancestor.parent();
    }
    None
}

#[derive(Default)]
struct MergeState {
    scope_kinds_by_symbol: HashMap<P<Symbol>, HashMap<Option<P<Node>>, u8>>,
}

impl MergeState {
    fn visit(&mut self, node: P<Node>, reports: &mut Vec<Violation>) {
        if let Some(report) = self.check(node) {
            reports.push(report);
        }
        node.for_each_child(&mut |child| {
            self.visit(child, reports);
            false
        });
    }

    fn check(&mut self, node: P<Node>) -> Option<Violation> {
        let unsafe_kind = match node.kind() {
            Kind::ClassDeclaration => INTERFACE,
            // Preserve the interface listener's upstream type-parameter scope.
            Kind::InterfaceDeclaration if node.type_parameters().is_empty() => CLASS,
            _ => return None,
        };
        let name = node.name()?;
        let mut symbol = node.symbol()?;
        if symbol.name() == InternalSymbolNameDefault {
            if let Some(local) = node.local_symbol() {
                symbol = local;
            } else if let Some(local) = nearest_locals_container(node)
                .and_then(|container| get_locals(container).lookup(name.text()))
            {
                symbol = local;
            }
        }
        if symbol.declarations().len() <= 1 {
            return None;
        }
        let scope = enclosing_scope(node);
        let has_unsafe_declaration = if symbol.declarations().len() <= DECLARATION_CACHE_THRESHOLD {
            symbol.declarations().iter().any(|&declaration| {
                declaration_kind(declaration) == unsafe_kind
                    && enclosing_scope(declaration) == scope
            })
        } else {
            let kinds = self.scope_kinds_by_symbol.entry(symbol).or_insert_with(|| {
                let mut kinds = HashMap::new();
                for &declaration in symbol.declarations() {
                    let kind = declaration_kind(declaration);
                    if kind != 0 {
                        *kinds.entry(enclosing_scope(declaration)).or_default() |= kind;
                    }
                }
                kinds
            });
            kinds.get(&scope).copied().unwrap_or_default() & unsafe_kind != 0
        };
        has_unsafe_declaration.then(|| {
            Violation::new(
                name,
                "unsafeMerging",
                "Unsafe declaration merging between classes and interfaces.",
            )
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use tsrs_ast::SourceFileParseOptions;
    use tsrs_core::{ScriptKind, arena::Region, tspath::Path};

    #[test]
    fn duplicate_enum_rule_ignores_templates_with_invalid_escapes() {
        // The baseline rule tester observes the rule even when the input has
        // syntax errors. Keep that rule-level contract separate from the public
        // Rust entry point, which currently rejects syntax diagnostics.
        let source = r"enum E { A = `\unicode`, B = `\unicode` }";
        let region = Region::new(1 << 20);
        let _scope = region.enter();
        let path = "/rslint-ts-structural-tests/invalid-template.ts";
        let file = tsrs_parser::parse_source_file(
            SourceFileParseOptions {
                file_name: path.into(),
                path: Path::new(path),
                ..Default::default()
            },
            source,
            ScriptKind::TS,
        );
        let statements = file.as_node().statements();
        assert_eq!(statements.len(), 1);
        let declaration = statements[0];
        assert_eq!(declaration.kind(), Kind::EnumDeclaration);
        assert_eq!(declaration.members().len(), 2);
        for member in declaration.members() {
            let initializer = member.as_enum_member().initializer().unwrap();
            assert_eq!(initializer.kind(), Kind::NoSubstitutionTemplateLiteral);
            assert!(
                initializer
                    .as_no_substitution_template_literal()
                    .template_flags()
                    .intersects(TokenFlags::ContainsInvalidEscape)
            );
            assert!(enum_value(initializer).is_none());
        }
        assert!(duplicate_enum_values(declaration).is_empty());
    }
}
