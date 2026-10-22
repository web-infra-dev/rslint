//! Type-aware rsbuild rules, using the frozen rslint rules' default options.
//! The dispatcher rejects non-default options before entering this module.

use tsrs_ast::{
    Kind, Node, SourceFile, Symbol, SymbolFlags, get_source_file_of_node, skip_parentheses,
};
use tsrs_checker::{Checker, SignatureKind, Type, TypeFlags};
use tsrs_core::P;

use crate::{Violation, edits::Edit};

pub(crate) const NAMES: &[&str] = &[
    "@typescript-eslint/no-for-in-array",
    "@typescript-eslint/no-array-delete",
    "@typescript-eslint/restrict-plus-operands",
    "@typescript-eslint/only-throw-error",
];

pub(crate) fn check(
    rule: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    default_library: &dyn Fn(P<SourceFile>) -> bool,
) -> Vec<Violation> {
    match (rule, node.kind()) {
        ("@typescript-eslint/no-for-in-array", Kind::ForInStatement) => {
            let value = constrained(node.as_for_in_or_of_statement().expression(), checker);
            if !array_like(value, checker) {
                return Vec::new();
            }
            let start = trimmed_range(node, file)[0];
            let end = node
                .statement()
                .pos()
                .expect_source("parsed statement position") as usize;
            vec![Violation::new(node, "forInViolation", "For-in loops over arrays skips holes, returns indices as strings, and may visit the prototype chain or other enumerable properties. Use a more robust iteration method such as for-of or array.forEach instead.")
                .with_byte_range([start, end])]
        }
        ("@typescript-eslint/no-array-delete", Kind::DeleteExpression) => {
            array_delete(node, file, checker)
        }
        ("@typescript-eslint/only-throw-error", Kind::ThrowStatement) => {
            only_throw_error(node, checker, default_library)
        }
        ("@typescript-eslint/restrict-plus-operands", Kind::BinaryExpression)
            if matches!(
                node.as_binary_expression().operator_token().kind(),
                Kind::PlusToken | Kind::PlusEqualsToken
            ) =>
        {
            plus_operands(node, checker)
        }
        _ => Vec::new(),
    }
}

fn constrained(node: P<Node>, checker: &mut Checker) -> P<Type> {
    let value = checker.get_type_at_location(node);
    checker.get_base_constraint_of_type(value).unwrap_or(value)
}

fn trimmed_range(node: P<Node>, file: P<SourceFile>) -> [usize; 2] {
    let start = node.pos().expect_source("parsed node position");
    let end = node.end().expect_source("parsed node end");
    [
        tsrs_scanner::skip_trivia(file.text(), start) as usize,
        end as usize,
    ]
}

fn token_range(file: P<SourceFile>, offset: usize) -> [usize; 2] {
    let range = tsrs_scanner::get_range_of_token_at_position(file, offset as u32);
    [
        range.pos().expect_source("token position") as usize,
        range.end().expect_source("token end") as usize,
    ]
}

// The baseline accepts a union/intersection if *any* recursively expanded part
// has a numeric index and numeric length; this differs from isArrayType.
fn array_like(value: P<Type>, checker: &mut Checker) -> bool {
    if value.flags().intersects(TypeFlags::UnionOrIntersection) {
        return value.types().iter().any(|&part| array_like(part, checker));
    }
    let number = checker.get_number_type();
    if checker.get_index_type_of_type(value, number).is_none() {
        return false;
    }
    let Some(length) = checker.get_property_of_type(value, "length") else {
        return false;
    };
    checker
        .get_type_of_symbol(length)
        .flags()
        .intersects(TypeFlags::NumberLike)
}

fn underlying_array(value: P<Type>, checker: &mut Checker) -> bool {
    if value.flags().intersects(TypeFlags::Union) {
        value
            .types()
            .iter()
            .all(|&part| checker.is_array_or_tuple_type(part))
    } else if value.flags().intersects(TypeFlags::Intersection) {
        value
            .types()
            .iter()
            .any(|&part| checker.is_array_or_tuple_type(part))
    } else {
        checker.is_array_or_tuple_type(value)
    }
}

fn array_delete(node: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> Vec<Violation> {
    let access = skip_parentheses(node.as_delete_expression().expression());
    if access.kind() != Kind::ElementAccessExpression {
        return Vec::new();
    }
    let access = access.as_element_access_expression();
    let value = constrained(access.expression(), checker);
    if !underlying_array(value, checker) {
        return Vec::new();
    }
    let expression = trimmed_range(access.expression(), file);
    let argument = trimmed_range(access.argument_expression(), file);
    let delete = token_range(
        file,
        node.pos().expect_source("parsed delete position") as usize,
    );
    let open = token_range(file, expression[1]);
    let close = token_range(file, argument[1]);
    vec![
        Violation::new(
            node,
            "noArrayDelete",
            "Using the `delete` operator with an array expression is unsafe.",
        )
        .with_suggestion(
            "useSplice",
            "Use `array.splice()` instead.",
            vec![
                Edit::remove(delete),
                Edit::replace(open, ".splice("),
                Edit::replace(close, ", 1)"),
            ],
        ),
    ]
}

const STRING_LIKE: &str =
    "string, allowing a string + any of: `any`, `boolean`, `null`, `RegExp`, `undefined`";

fn type_name(value: P<Type>, checker: &mut Checker) -> String {
    if value.flags().intersects(TypeFlags::StringLike) {
        return "string".into();
    }
    if value.flags().intersects(TypeFlags::TypeParameter)
        && let Some(declaration) = value
            .symbol()
            .and_then(|symbol| symbol.declarations().first().copied())
        && declaration.kind() == Kind::TypeParameter
        && let Some(constraint) = declaration.as_type_parameter_declaration().constraint()
    {
        let constraint = checker.get_type_from_type_node_exported(constraint);
        return type_name(constraint, checker);
    }
    if (value.flags().intersects(TypeFlags::Union)
        && value
            .types()
            .iter()
            .all(|&part| type_name(part, checker) == "string"))
        || (value.flags().intersects(TypeFlags::Intersection)
            && value
                .types()
                .iter()
                .any(|&part| type_name(part, checker) == "string"))
    {
        return "string".into();
    }
    checker.type_to_string_exported(value)
}

fn invalid_object(value: P<Type>, other: P<Type>, checker: &mut Checker) -> Option<String> {
    let parameter = value.flags().intersects(TypeFlags::TypeParameter);
    let object = if value.flags().intersects(TypeFlags::Intersection) {
        value
            .types()
            .iter()
            .all(|part| part.flags().intersects(TypeFlags::Object))
    } else {
        value.flags().intersects(TypeFlags::Object)
    };
    if !parameter && !object {
        return None;
    }
    let name = type_name(value, checker);
    let rendered = if parameter {
        checker.type_to_string_exported(value)
    } else {
        name.clone()
    };
    if name == "RegExp" {
        return other
            .flags()
            .intersects(TypeFlags::NumberLike)
            .then_some(rendered);
    }
    object.then_some(rendered)
}

fn invalid_operand(
    node: P<Node>,
    value: P<Type>,
    other: P<Type>,
    checker: &mut Checker,
    reports: &mut Vec<Violation>,
) -> (TypeFlags, bool) {
    let parts = if value.flags().intersects(TypeFlags::Union) {
        value.types().to_vec()
    } else {
        vec![value]
    };
    let flags = parts
        .iter()
        .fold(TypeFlags::None, |flags, part| flags | part.flags());
    let report = |name: String| {
        Violation::new(
            node,
            "invalid",
            format!(
                "Invalid operand for a '+' operation. Operands must each be a number or {STRING_LIKE}. Got `{name}`."
            ),
        )
    };
    if flags.intersects(TypeFlags::ESSymbolLike | TypeFlags::Never | TypeFlags::Unknown) {
        reports.push(report(checker.type_to_string_exported(value)));
        return (flags, true);
    }
    let mut invalid = false;
    for part in parts {
        if let Some(name) = invalid_object(part, other, checker) {
            reports.push(report(name));
            invalid = true;
        }
    }
    (flags, invalid)
}

fn plus_operands(node: P<Node>, checker: &mut Checker) -> Vec<Violation> {
    let binary = node.as_binary_expression();
    let left = constrained(binary.left(), checker);
    let left = checker.get_base_type_of_literal_type(left);
    let right = constrained(binary.right(), checker);
    let right = checker.get_base_type_of_literal_type(right);
    if left == right
        && left
            .flags()
            .intersects(TypeFlags::BigIntLike | TypeFlags::NumberLike | TypeFlags::StringLike)
    {
        return Vec::new();
    }
    let mut reports = Vec::new();
    let (left_flags, left_invalid) = invalid_operand(
        skip_parentheses(binary.left()),
        left,
        right,
        checker,
        &mut reports,
    );
    let (right_flags, right_invalid) = invalid_operand(
        skip_parentheses(binary.right()),
        right,
        left,
        checker,
        &mut reports,
    );
    if !left_invalid
        && !right_invalid
        && ((left_flags.intersects(TypeFlags::NumberLike)
            && right_flags.intersects(TypeFlags::BigIntLike))
            || (right_flags.intersects(TypeFlags::NumberLike)
                && left_flags.intersects(TypeFlags::BigIntLike)))
    {
        let left = checker.type_to_string_exported(left);
        let right = checker.type_to_string_exported(right);
        reports.push(Violation::new(node, "bigintAndNumber", format!("Numeric '+' operations must either be both bigints or both numbers. Got `{left}` + `{right}`.")));
    }
    reports
}

fn union_parts(value: P<Type>) -> Vec<P<Type>> {
    if value.flags().intersects(TypeFlags::Union) {
        value.types().to_vec()
    } else {
        vec![value]
    }
}

fn is_callback(parameter: P<Symbol>, node: P<Node>, checker: &mut Checker) -> bool {
    let Some(value) = checker.get_type_of_symbol_at_location(parameter, Some(node)) else {
        return false;
    };
    let mut value = checker.get_apparent_type(value);
    if parameter.value_declaration().is_some_and(|declaration| {
        declaration.kind() == Kind::Parameter
            && declaration
                .as_parameter_declaration()
                .dot_dot_dot_token()
                .is_some()
    }) {
        let number = checker.get_number_type();
        let Some(element) = checker.get_index_type_of_type(value, number) else {
            return false;
        };
        value = element;
    }
    union_parts(value).iter().any(|&part| {
        !checker
            .get_signatures_of_type(part, SignatureKind::Call)
            .is_empty()
    })
}

fn thenable(node: P<Node>, checker: &mut Checker) -> bool {
    let value = checker.get_type_at_location(node);
    let value = checker.get_apparent_type(value);
    for part in union_parts(value) {
        let Some(then) = checker.get_property_of_type(part, "then") else {
            continue;
        };
        let Some(then_type) = checker.get_type_of_symbol_at_location(then, Some(node)) else {
            continue;
        };
        for part in union_parts(then_type) {
            let signatures = checker
                .get_signatures_of_type(part, SignatureKind::Call)
                .to_vec();
            for signature in signatures {
                if let Some(&first) = signature.parameters().first()
                    && is_callback(first, node, checker)
                {
                    return true;
                }
            }
        }
    }
    false
}

fn rethrown(expr: P<Node>, checker: &mut Checker) -> bool {
    if expr.kind() != Kind::Identifier {
        return false;
    }
    let Some(declaration) = checker
        .get_symbol_at_location_exported(expr)
        .and_then(|symbol| symbol.value_declaration())
    else {
        return false;
    };
    if declaration
        .parent()
        .is_some_and(|parent| parent.kind() == Kind::CatchClause)
    {
        return true;
    }
    if declaration.kind() != Kind::Parameter
        || declaration
            .as_parameter_declaration()
            .dot_dot_dot_token()
            .is_some()
    {
        return false;
    }
    let Some(function) = declaration
        .parent()
        .filter(|node| node.kind() == Kind::ArrowFunction)
    else {
        return false;
    };
    if function.parameters().first() != Some(&declaration) {
        return false;
    }
    let Some(call) = function
        .parent()
        .filter(|node| node.kind() == Kind::CallExpression)
    else {
        return false;
    };
    let call = call.as_call_expression();
    let callee = call.expression();
    if !matches!(
        callee.kind(),
        Kind::PropertyAccessExpression | Kind::ElementAccessExpression
    ) {
        return false;
    }
    let (name, found) = checker.get_accessed_property_name(callee);
    if !found {
        return false;
    }
    let args = call.arguments().nodes();
    let rejected = match name.as_ref() {
        "catch" if !args.is_empty() && args[0].kind() != Kind::SpreadElement => args[0],
        "then"
            if args.len() >= 2
                && args[0].kind() != Kind::SpreadElement
                && args[1].kind() != Kind::SpreadElement =>
        {
            args[1]
        }
        _ => return false,
    };
    rejected == function
        && thenable(
            callee.expression().expect("access expression receiver"),
            checker,
        )
}

fn error_like(
    value: P<Type>,
    checker: &mut Checker,
    default_library: &dyn Fn(P<SourceFile>) -> bool,
) -> bool {
    if value.flags().intersects(TypeFlags::Intersection) {
        return value
            .types()
            .iter()
            .any(|&part| error_like(part, checker, default_library));
    }
    if value.flags().intersects(TypeFlags::Union) {
        return value
            .types()
            .iter()
            .all(|&part| error_like(part, checker, default_library));
    }
    if value.flags().intersects(TypeFlags::TypeParameter) {
        return checker
            .get_base_constraint_of_type(value)
            .is_some_and(|constraint| error_like(constraint, checker, default_library));
    }
    let Some(symbol) = value.symbol() else {
        return false;
    };
    if symbol.name() == "Error"
        && symbol
            .declarations()
            .iter()
            .any(|&declaration| get_source_file_of_node(declaration).is_some_and(default_library))
    {
        return true;
    }
    if symbol
        .flags()
        .intersects(SymbolFlags::Class | SymbolFlags::Interface)
    {
        let declared = checker.get_declared_type_of_symbol(symbol);
        return checker
            .get_base_types(declared)
            .iter()
            .any(|&base| error_like(base, checker, default_library));
    }
    false
}

fn only_throw_error(
    node: P<Node>,
    checker: &mut Checker,
    default_library: &dyn Fn(P<SourceFile>) -> bool,
) -> Vec<Violation> {
    let expr = node.expression().expect("throw statement expression");
    if rethrown(expr, checker) {
        return Vec::new();
    }
    let value = checker.get_type_at_location(expr);
    if value.flags().intersects(TypeFlags::Undefined) {
        return vec![Violation::new(expr, "undef", "Do not throw undefined.")];
    }
    if value
        .flags()
        .intersects(TypeFlags::Any | TypeFlags::Unknown)
        || error_like(value, checker, default_library)
    {
        return Vec::new();
    }
    vec![Violation::new(
        expr,
        "object",
        "Expected an error object to be thrown.",
    )]
}
