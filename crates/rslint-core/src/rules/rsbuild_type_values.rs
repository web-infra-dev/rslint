//! Frozen rsbuild default options for template interpolation and promise rejection.
//! Ported from rslint e0a29f751d4def86d27931759098f27748fa3003. The
//! declaration-origin checks deliberately distinguish builtins from lookalikes.

use tsrs_ast::{
    Kind, Node, SourceFile, Symbol, SymbolFlags, get_source_file_of_node, skip_parentheses,
};
use tsrs_checker::{Checker, ObjectFlags, Type, TypeFlags};
use tsrs_core::P;

use crate::Violation;

pub(crate) const NAMES: &[&str] = &[
    "@typescript-eslint/restrict-template-expressions",
    "@typescript-eslint/prefer-promise-reject-errors",
];

pub(crate) fn check(
    rule: &str,
    node: P<Node>,
    _file: P<SourceFile>,
    checker: &mut Checker,
    default_library: &dyn Fn(P<SourceFile>) -> bool,
) -> Vec<Violation> {
    match (rule, node.kind()) {
        ("@typescript-eslint/restrict-template-expressions", Kind::TemplateExpression) => {
            if node
                .parent()
                .is_some_and(|parent| parent.kind() == Kind::TaggedTemplateExpression)
            {
                return Vec::new();
            }
            node.as_template_expression()
                .template_spans
                .nodes()
                .iter()
                .filter_map(|span| {
                    let expression = span.as_template_span().expression;
                    let value = checker.get_type_at_location(expression);
                    let value = checker.get_base_constraint_of_type(value).unwrap_or(value);
                    (!template_allowed(value, checker, default_library)).then(|| {
                        Violation::new(
                            expression,
                            "invalidType",
                            format!(
                                "Invalid type \"{}\" of template literal expression.",
                                checker.type_to_string_exported(value)
                            ),
                        )
                    })
                })
                .collect()
        }
        ("@typescript-eslint/prefer-promise-reject-errors", Kind::CallExpression) => {
            if !reject_call(node, checker, default_library) {
                return Vec::new();
            }
            if let Some(&argument) = node.as_call_expression().arguments.nodes().first() {
                let value = checker.get_type_at_location(argument);
                if builtin_like(value, Builtin::Name("Error"), checker, default_library)
                    || builtin_like(value, Builtin::ReadonlyError, checker, default_library)
                {
                    return Vec::new();
                }
            }
            vec![Violation::new(
                node,
                "rejectAnError",
                "Expected the Promise rejection reason to be an Error.",
            )]
        }
        _ => Vec::new(),
    }
}

fn from_default_library(
    symbol: P<Symbol>,
    default_library: &dyn Fn(P<SourceFile>) -> bool,
) -> bool {
    symbol
        .declarations()
        .iter()
        .any(|&declaration| get_source_file_of_node(declaration).is_some_and(default_library))
}

#[derive(Clone, Copy)]
enum Builtin {
    Name(&'static str),
    Promise,
    PromiseConstructor,
    ReadonlyError,
}

// Like the frozen helper: all union parts / any intersection part must match.
// Each caller keeps its own predicate, including Error versus Readonly<Error>.
fn builtin_like(
    value: P<Type>,
    wanted: Builtin,
    checker: &mut Checker,
    default_library: &dyn Fn(P<SourceFile>) -> bool,
) -> bool {
    if value.flags().intersects(TypeFlags::Union) {
        return value
            .types()
            .iter()
            .all(|&part| builtin_like(part, wanted, checker, default_library));
    }
    if value.flags().intersects(TypeFlags::Intersection) {
        return value
            .types()
            .iter()
            .any(|&part| builtin_like(part, wanted, checker, default_library));
    }
    if value.flags().intersects(TypeFlags::TypeParameter) {
        return checker
            .get_base_constraint_of_type(value)
            .is_some_and(|constraint| builtin_like(constraint, wanted, checker, default_library));
    }
    match wanted {
        Builtin::Promise => {
            return !class_constructor(value)
                && builtin_like(value, Builtin::Name("Promise"), checker, default_library);
        }
        Builtin::PromiseConstructor => {
            return builtin_like(
                value,
                Builtin::Name("PromiseConstructor"),
                checker,
                default_library,
            ) || class_constructor(value)
                && builtin_like(value, Builtin::Name("Promise"), checker, default_library);
        }
        Builtin::ReadonlyError => {
            let Some(alias) = value.alias() else {
                return false;
            };
            let Some(&argument) = alias.type_arguments().first() else {
                return false;
            };
            if let Some(symbol) = alias.symbol()
                && from_default_library(symbol, default_library)
                && symbol.name() == "Readonly"
                && (builtin_like(argument, Builtin::Name("Error"), checker, default_library)
                    || builtin_like(argument, Builtin::ReadonlyError, checker, default_library))
            {
                return true;
            }
        }
        Builtin::Name(name) => {
            let Some(symbol) = value.symbol() else {
                return false;
            };
            if symbol.name() == name && from_default_library(symbol, default_library) {
                return true;
            }
        }
    }
    if let Some(symbol) = value.symbol()
        && symbol
            .flags()
            .intersects(SymbolFlags::Class | SymbolFlags::Interface)
    {
        let declared = checker.get_declared_type_of_symbol(symbol);
        return checker
            .get_base_types(declared)
            .iter()
            .any(|&base| builtin_like(base, wanted, checker, default_library));
    }
    false
}

fn class_constructor(value: P<Type>) -> bool {
    value.flags().intersects(TypeFlags::Object)
        && value.object_flags().intersects(ObjectFlags::Anonymous)
        && value
            .symbol()
            .is_some_and(|symbol| symbol.flags().intersects(SymbolFlags::Class))
}

fn reject_call(
    node: P<Node>,
    checker: &mut Checker,
    default_library: &dyn Fn(P<SourceFile>) -> bool,
) -> bool {
    let callee = skip_parentheses(node.expression().expect("call expression"));
    if matches!(
        callee.kind(),
        Kind::PropertyAccessExpression | Kind::ElementAccessExpression
    ) {
        if checker.get_accessed_property_name(callee).0 != "reject" {
            return false;
        }
        let value = checker.get_type_at_location(callee.expression().expect("member receiver"));
        return builtin_like(value, Builtin::PromiseConstructor, checker, default_library)
            || builtin_like(value, Builtin::Promise, checker, default_library);
    }
    if callee.kind() != Kind::Identifier {
        return false;
    }
    let Some(symbol) = checker.get_symbol_at_location_exported(callee) else {
        return false;
    };
    let Some(parameter) = symbol.value_declaration() else {
        return false;
    };
    if parameter.kind() != Kind::Parameter {
        return false;
    }
    let Some(function) = parameter.parent() else {
        return false;
    };
    if !matches!(
        function.kind(),
        Kind::FunctionExpression | Kind::ArrowFunction
    ) || function.parameters().get(1).copied() != Some(parameter)
    {
        return false;
    }
    let mut parent = function.parent();
    while let Some(node) = parent {
        if node.kind() == Kind::ParenthesizedExpression {
            parent = node.parent();
            continue;
        }
        if node.kind() != Kind::NewExpression {
            return false;
        }
        let value =
            checker.get_type_at_location(node.expression().expect("constructor expression"));
        return builtin_like(value, Builtin::PromiseConstructor, checker, default_library);
    }
    false
}

fn template_allowed(
    value: P<Type>,
    checker: &mut Checker,
    default_library: &dyn Fn(P<SourceFile>) -> bool,
) -> bool {
    if value.flags().intersects(TypeFlags::Union) {
        return value
            .types()
            .iter()
            .all(|&part| template_allowed(part, checker, default_library));
    }
    if value.flags().intersects(TypeFlags::Intersection) {
        return value
            .types()
            .iter()
            .any(|&part| template_allowed(part, checker, default_library));
    }
    if value.flags().intersects(
        TypeFlags::StringLike
            | TypeFlags::Any
            | TypeFlags::BooleanLike
            | TypeFlags::Null
            | TypeFlags::Undefined
            | TypeFlags::NumberLike
            | TypeFlags::BigIntLike,
    ) {
        return true;
    }
    if allowed_lib_type_or_base(value, checker, default_library, &mut Vec::new()) {
        return true;
    }
    // The baseline's allowRegExp option uses the printed type name, rather than
    // the declaration identity used by the Error/URL/URLSearchParams allow list.
    type_name(value, checker) == "RegExp"
}

fn allowed_lib_type_or_base(
    value: P<Type>,
    checker: &mut Checker,
    default_library: &dyn Fn(P<SourceFile>) -> bool,
    seen: &mut Vec<P<Type>>,
) -> bool {
    if seen.contains(&value) {
        return false;
    }
    seen.push(value);
    if allowed_lib_specifier(value, default_library) {
        return true;
    }
    if !value.flags().intersects(TypeFlags::Object) {
        return false;
    }
    let target = if value.object_flags().intersects(ObjectFlags::Reference) {
        value.target().unwrap_or(value)
    } else {
        value
    };
    if !target
        .object_flags()
        .intersects(ObjectFlags::ClassOrInterface)
    {
        return false;
    }
    checker
        .get_base_types(target)
        .iter()
        .any(|&base| allowed_lib_type_or_base(base, checker, default_library, seen))
}

fn allowed_lib_specifier(value: P<Type>, default_library: &dyn Fn(P<SourceFile>) -> bool) -> bool {
    if value.flags().intersects(TypeFlags::Union) {
        return value
            .types()
            .iter()
            .all(|&part| allowed_lib_specifier(part, default_library));
    }
    // Specifier names prefer the alias, but declaration origins prefer the
    // actual symbol. This preserves the baseline's type alias behavior.
    let alias = value.alias().and_then(|alias| alias.symbol());
    let name_symbol = alias.or_else(|| value.symbol());
    if name_symbol
        .is_some_and(|symbol| matches!(symbol.name(), "Error" | "URL" | "URLSearchParams"))
        && let Some(symbol) = value.symbol().or(alias)
        && (symbol.declarations().is_empty() || from_default_library(symbol, default_library))
    {
        return true;
    }
    value.flags().intersects(TypeFlags::Intersection)
        && value
            .types()
            .iter()
            .any(|&part| allowed_lib_specifier(part, default_library))
}

fn type_name(value: P<Type>, checker: &mut Checker) -> String {
    if value.flags().intersects(TypeFlags::StringLike) {
        return "string".to_string();
    }
    if value.flags().intersects(TypeFlags::TypeParameter)
        && let Some(symbol) = value.symbol()
        && let Some(&declaration) = symbol.declarations().first()
        && declaration.kind() == Kind::TypeParameter
        && let Some(constraint) = declaration.as_type_parameter_declaration().constraint
    {
        let value = checker.get_type_from_type_node_exported(constraint);
        return type_name(value, checker);
    }
    if value.flags().intersects(TypeFlags::Union)
        && value
            .types()
            .iter()
            .all(|&part| type_name(part, checker) == "string")
    {
        return "string".to_string();
    }
    if value.flags().intersects(TypeFlags::Intersection)
        && value
            .types()
            .iter()
            .any(|&part| type_name(part, checker) == "string")
    {
        return "string".to_string();
    }
    checker.type_to_string_exported(value)
}
