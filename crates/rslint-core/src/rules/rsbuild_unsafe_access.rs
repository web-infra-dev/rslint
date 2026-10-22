//! Default-option unsafe access rules enabled by the frozen rsbuild workload.
//! Behavior follows rslint e0a29f751d4def86d27931759098f27748fa3003,
//! including its error-type messages, `this` handling, and heritage exclusions.

use tsrs_ast::{
    Kind, Node, SourceFile, SymbolFlags, get_source_file_of_node, is_literal_expression,
    is_optional_chain, skip_parentheses,
};
use tsrs_checker::{Checker, SignatureKind, Type, TypeFlags};
use tsrs_core::P;

use crate::Violation;

pub(crate) const NAMES: &[&str] = &[
    "@typescript-eslint/no-unsafe-call",
    "@typescript-eslint/no-unsafe-member-access",
];

pub(crate) fn check(
    rule: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    default_library: &dyn Fn(P<SourceFile>) -> bool,
) -> Vec<Violation> {
    match rule {
        "@typescript-eslint/no-unsafe-call" => unsafe_call(node, checker, default_library),
        "@typescript-eslint/no-unsafe-member-access" => unsafe_member(node, file, checker),
        _ => Vec::new(),
    }
}

fn any(value: P<Type>) -> bool {
    value.flags().intersects(TypeFlags::Any)
}

fn error(value: P<Type>) -> bool {
    value.flags().intersects(TypeFlags::Intrinsic)
        && value.as_intrinsic_type().intrinsic_name() == "error"
}

fn constrained(node: P<Node>, checker: &mut Checker) -> P<Type> {
    let value = checker.get_type_at_location(node);
    checker.get_base_constraint_of_type(value).unwrap_or(value)
}

fn access(node: P<Node>) -> bool {
    matches!(
        node.kind(),
        Kind::PropertyAccessExpression | Kind::ElementAccessExpression
    )
}

fn this_expression(mut node: P<Node>) -> Option<P<Node>> {
    loop {
        node = skip_parentheses(node);
        match node.kind() {
            Kind::ThisKeyword => return Some(node),
            Kind::CallExpression
            | Kind::PropertyAccessExpression
            | Kind::ElementAccessExpression => {
                node = node.expression().expect("call/access expression");
            }
            _ => return None,
        }
    }
}

fn function_like(
    value: P<Type>,
    checker: &mut Checker,
    default_library: &dyn Fn(P<SourceFile>) -> bool,
) -> bool {
    if value.flags().intersects(TypeFlags::Intersection) {
        return value
            .types()
            .iter()
            .any(|&part| function_like(part, checker, default_library));
    }
    if value.flags().intersects(TypeFlags::Union) {
        return value
            .types()
            .iter()
            .all(|&part| function_like(part, checker, default_library));
    }
    if value.flags().intersects(TypeFlags::TypeParameter) {
        return checker
            .get_base_constraint_of_type(value)
            .is_some_and(|constraint| function_like(constraint, checker, default_library));
    }
    let Some(symbol) = value.symbol() else {
        return false;
    };
    if symbol.name() == "Function"
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
            .any(|&base| function_like(base, checker, default_library));
    }
    false
}

fn unsafe_call(
    node: P<Node>,
    checker: &mut Checker,
    default_library: &dyn Fn(P<SourceFile>) -> bool,
) -> Vec<Violation> {
    let (callee, reporting, message_id, action, new_call) = match node.kind() {
        Kind::CallExpression => {
            let callee = node.expression().expect("call expression");
            if callee.kind() == Kind::ImportKeyword {
                return Vec::new();
            }
            (callee, callee, "unsafeCall", "call", false)
        }
        Kind::NewExpression => (
            node.expression().expect("new expression"),
            node,
            "unsafeNew",
            "construction",
            true,
        ),
        Kind::TaggedTemplateExpression => {
            let tag = node.as_tagged_template_expression().tag;
            (tag, tag, "unsafeTemplateTag", "use", false)
        }
        _ => return Vec::new(),
    };

    // The frozen rslint baseline checks `this` before the callee, even with
    // noImplicitThis enabled and for construction/template expressions.
    if let Some(this) = this_expression(callee) {
        let value = constrained(this, checker);
        if any(value) {
            let name = if error(value) {
                "`error` type"
            } else {
                "`any`"
            };
            return vec![Violation::new(
                reporting,
                "unsafeCallThis",
                format!(
                    "Unsafe call of a(n) {name} typed value. `this` is typed as {name}.\nYou can try to fix this by turning on the `noImplicitThis` compiler option, or adding a `this` parameter to the function."
                ),
            )];
        }
    }

    let value = constrained(callee, checker);
    let name = if any(value) {
        if error(value) {
            "`error` type"
        } else {
            "`any`"
        }
    } else if function_like(value, checker, default_library) {
        if !checker
            .get_signatures_of_type(value, SignatureKind::Construct)
            .is_empty()
        {
            return Vec::new();
        }
        let signatures = checker
            .get_signatures_of_type(value, SignatureKind::Call)
            .to_vec();
        if (!new_call && !signatures.is_empty())
            || (new_call
                && signatures.iter().any(|&signature| {
                    !checker
                        .get_return_type_of_signature_exported(signature)
                        .flags()
                        .intersects(TypeFlags::Void)
                }))
        {
            return Vec::new();
        }
        "`Function`"
    } else {
        return Vec::new();
    };
    let message = if message_id == "unsafeTemplateTag" {
        format!("Unsafe use of a(n) {name} typed template tag.")
    } else {
        format!("Unsafe {action} of a(n) {name} typed value.")
    };
    vec![Violation::new(reporting, message_id, message)]
}

fn ignored_heritage(node: P<Node>) -> bool {
    let mut ancestor = node.parent();
    while let Some(parent) = ancestor {
        if parent.kind() == Kind::HeritageClause {
            return parent.as_heritage_clause().token == Kind::ImplementsKeyword
                || parent
                    .parent()
                    .is_some_and(|declaration| declaration.kind() == Kind::InterfaceDeclaration);
        }
        // A function/class expression inside class extends can have its own
        // heritage; only an actual ancestor clause owns the ignored range.
        ancestor = parent.parent();
    }
    false
}

fn jsx_tag_member(mut node: P<Node>) -> bool {
    while let Some(parent) = node.parent() {
        if parent.kind() == Kind::PropertyAccessExpression && parent.expression() == Some(node) {
            node = parent;
        } else {
            return matches!(
                parent.kind(),
                Kind::JsxOpeningElement | Kind::JsxClosingElement | Kind::JsxSelfClosingElement
            ) && parent.tag_name() == node;
        }
    }
    false
}

// The default option does not allow optional access. A member chain is unsafe
// once its first unsafe receiver is found. Parenthesized optional chains and
// calls end suppression, as they do in the baseline's cached visitor.
fn member_state_unsafe(node: P<Node>, checker: &mut Checker) -> bool {
    if ignored_heritage(node) || jsx_tag_member(node) {
        return false;
    }
    let expression = node.expression().expect("member receiver");
    let unwrapped = skip_parentheses(expression);
    if access(unwrapped)
        && (expression == unwrapped || !is_optional_chain(unwrapped))
        && member_state_unsafe(unwrapped, checker)
    {
        return true;
    }
    any(checker.get_type_at_location(expression))
}

fn text(node: P<Node>, file: P<SourceFile>) -> &'static str {
    let start = tsrs_scanner::skip_trivia(file.text(), node.pos().expect_source("source position"))
        as usize;
    let end = node.end().expect_source("source end") as usize;
    &file.text()[start..end]
}

fn unsafe_member(node: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> Vec<Violation> {
    if !access(node) || jsx_tag_member(node) {
        return Vec::new();
    }
    let (property, name) = if node.kind() == Kind::PropertyAccessExpression {
        let property = node.name().expect("property name");
        (property, format!(".{}", text(property, file)))
    } else {
        let property = skip_parentheses(node.as_element_access_expression().argument_expression);
        (property, format!("[{}]", text(property, file)))
    };
    let expression = node.expression().expect("member receiver");
    let unwrapped = skip_parentheses(expression);
    let suppressed = access(unwrapped)
        && (expression == unwrapped || !is_optional_chain(unwrapped))
        && member_state_unsafe(unwrapped, checker);
    let mut violations = Vec::new();
    if !ignored_heritage(node) && !suppressed {
        let value = checker.get_type_at_location(expression);
        if any(value) {
            let (mut id, mut message) = if error(value) {
                (
                    "errorMemberExpression",
                    format!("Unsafe member access {name} on a type that cannot be resolved."),
                )
            } else {
                (
                    "unsafeMemberExpression",
                    format!("Unsafe member access {name} on an `any` value."),
                )
            };
            if !checker.no_implicit_this
                && let Some(this) = this_expression(node)
            {
                let this_type = constrained(this, checker);
                if any(this_type) {
                    (id, message) = if error(this_type) {
                        (
                            "errorThisMemberExpression",
                            format!(
                                "Unsafe member access {name}. The type of `this` cannot be resolved."
                            ),
                        )
                    } else {
                        (
                            "unsafeThisMemberExpression",
                            format!(
                                "Unsafe member access {name} on an `any` value. `this` is typed as `any`."
                            ),
                        )
                    };
                    message.push_str("\nYou can try to fix this by turning on the `noImplicitThis` compiler option, or adding a `this` parameter to the function.");
                }
            }
            violations.push(Violation::new(property, id, message));
        }
    }
    if node.kind() == Kind::ElementAccessExpression
        && !is_literal_expression(property)
        && !matches!(
            property.kind(),
            Kind::TrueKeyword | Kind::FalseKeyword | Kind::NullKeyword
        )
    {
        let operator = match property.kind() {
            Kind::PrefixUnaryExpression => property.as_prefix_unary_expression().operator(),
            Kind::PostfixUnaryExpression => property.as_postfix_unary_expression().operator,
            _ => Kind::Unknown,
        };
        if !matches!(operator, Kind::PlusPlusToken | Kind::MinusMinusToken) {
            let value = checker.get_type_at_location(property);
            if any(value) {
                let (id, message) = if error(value) {
                    (
                        "errorComputedMemberAccess",
                        format!("The type of computed name {name} cannot be resolved."),
                    )
                } else {
                    (
                        "unsafeComputedMemberAccess",
                        format!("Computed name {name} resolves to an `any` value."),
                    )
                };
                violations.push(Violation::new(property, id, message));
            }
        }
    }
    violations
}
