//! Unsafe argument/return semantics from rslint e0a29f751d4def86d27931759098f27748fa3003.
//! Types, generic targets, resolved signatures and contextual types come from tsrs.
use tsrs_ast::{self as ast, Kind, ModifierFlags, Node, SourceFile, Symbol, skip_parentheses};
use tsrs_checker::{
    CheckMode, Checker, ContextFlags, ElementFlags, ObjectFlags, Signature, SignatureKind, Type,
    TypeFlags, is_tuple_type_exported,
};
use tsrs_core::P;

use crate::Violation;

pub(crate) const NAMES: &[&str] = &[
    "@typescript-eslint/no-unsafe-argument",
    "@typescript-eslint/no-unsafe-return",
];
pub(crate) fn check(
    name: &str,
    node: P<Node>,
    _file: P<SourceFile>,
    checker: &mut Checker,
    _default_library: &dyn Fn(P<SourceFile>) -> bool,
) -> Vec<Violation> {
    match name {
        "@typescript-eslint/no-unsafe-argument" => argument(node, checker),
        "@typescript-eslint/no-unsafe-return" => unsafe_return(node, checker),
        _ => Vec::new(),
    }
}
fn any(t: P<Type>) -> bool {
    t.flags().intersects(TypeFlags::Any)
}
fn unknown(t: P<Type>) -> bool {
    t.flags().intersects(TypeFlags::Unknown)
}
fn error(t: P<Type>) -> bool {
    t.flags().intersects(TypeFlags::Intrinsic) && t.as_intrinsic_type().intrinsic_name() == "error"
}
fn reference(t: P<Type>) -> bool {
    t.flags().intersects(TypeFlags::Object) && t.object_flags().intersects(ObjectFlags::Reference)
}
fn array_of(t: P<Type>, flag: TypeFlags, checker: &mut Checker) -> bool {
    checker.is_array_type(t)
        && checker
            .get_type_arguments(t)
            .first()
            .is_some_and(|element| element.flags().intersects(flag))
}
fn constrained(node: P<Node>, checker: &mut Checker) -> P<Type> {
    let t = checker.get_type_at_location(node);
    checker.get_base_constraint_of_type(t).unwrap_or(t)
}
fn describe(t: P<Type>, checker: &mut Checker) -> String {
    if error(t) {
        "error typed".into()
    } else {
        format!("`{}`", checker.type_to_string_exported(t))
    }
}
fn unsafe_assignment(
    sender: P<Type>,
    receiver: P<Type>,
    node: Option<P<Node>>,
    checker: &mut Checker,
    visited: &mut Vec<(P<Type>, P<Type>)>,
) -> bool {
    if any(sender) {
        if unknown(receiver) {
            return false;
        }
        if !any(receiver) {
            return true;
        }
    }
    if visited.contains(&(sender, receiver)) {
        return false;
    }
    visited.push((sender, receiver));
    if !reference(sender) || !reference(receiver) || sender.target() != receiver.target() {
        return false;
    }
    // Frozen baseline exception for the default Map<any, any> constructor.
    if node.is_some_and(|node| {
        node.kind() == Kind::NewExpression
            && node
                .expression()
                .is_some_and(|callee| callee.kind() == Kind::Identifier && callee.text() == "Map")
            && node.arguments().is_empty()
            && node.type_argument_list().is_none()
    }) {
        return false;
    }
    let send = checker.get_type_arguments(sender).to_vec();
    let receive = checker.get_type_arguments(receiver).to_vec();
    send.into_iter()
        .zip(receive)
        .any(|(s, r)| unsafe_assignment(s, r, node, checker, visited))
}

enum Rest {
    Tuple { index: usize, types: Vec<P<Type>> },
    Other(P<Type>),
}
struct Parameters {
    types: Vec<P<Type>>,
    rest: Option<Rest>,
    index: usize,
    consumed: bool,
}
impl Parameters {
    fn new(node: P<Node>, checker: &mut Checker) -> Self {
        let resolved = checker.get_resolved_signature(node, None, CheckMode::Normal);
        let mut types = Vec::new();
        let mut rest = None;
        for (index, &parameter) in resolved.parameters().iter().enumerate() {
            let t = checker
                .get_type_of_symbol_at_location(parameter, Some(node))
                .expect("resolved parameter type");
            let constrained = checker.get_base_constraint_of_type(t).unwrap_or(t);
            if parameter.declarations().first().is_some_and(|decl| {
                decl.kind() == Kind::Parameter
                    && decl
                        .as_parameter_declaration()
                        .dot_dot_dot_token()
                        .is_some()
            }) {
                rest = Some(if is_tuple_type_exported(constrained) {
                    Rest::Tuple {
                        index,
                        types: checker.get_type_arguments(constrained).to_vec(),
                    }
                } else {
                    let number = checker.get_number_type();
                    Rest::Other(
                        checker
                            .get_index_type_of_type(constrained, number)
                            .unwrap_or(constrained),
                    )
                });
                break;
            }
            types.push(t);
        }
        Self {
            types,
            rest,
            index: 0,
            consumed: false,
        }
    }
    fn next(&mut self) -> Option<P<Type>> {
        let index = self.index;
        self.index += 1;
        if !self.consumed && index < self.types.len() {
            return Some(self.types[index]);
        }
        match &self.rest {
            None => None,
            Some(Rest::Other(t)) => Some(*t),
            Some(Rest::Tuple {
                index: start,
                types,
            }) => {
                if self.consumed {
                    types.last().copied()
                } else {
                    types.get(index - start).or_else(|| types.last()).copied()
                }
            }
        }
    }
}
fn argument(node: P<Node>, checker: &mut Checker) -> Vec<Violation> {
    let (callee, arguments) = match node.kind() {
        Kind::CallExpression | Kind::NewExpression => (
            node.expression().expect("callee"),
            node.arguments().to_vec(),
        ),
        Kind::TaggedTemplateExpression => {
            let tag = node.as_tagged_template_expression();
            if tag.template.kind() != Kind::TemplateExpression {
                return Vec::new();
            }
            (
                tag.tag,
                tag.template
                    .as_template_expression()
                    .template_spans
                    .nodes()
                    .iter()
                    .map(|span| span.as_template_span().expression)
                    .collect(),
            )
        }
        _ => return Vec::new(),
    };
    if arguments.is_empty() || any(checker.get_type_at_location(callee)) {
        return Vec::new();
    }
    let mut signature = Parameters::new(node, checker);
    if node.kind() == Kind::TaggedTemplateExpression {
        signature.next();
    }
    let mut result = Vec::new();
    for argument in arguments {
        if argument.kind() == Kind::SpreadElement {
            let spread =
                checker.get_type_at_location(argument.expression().expect("spread expression"));
            if any(spread) {
                result.push(Violation::new(
                    argument,
                    "unsafeSpread",
                    format!("Unsafe spread of an {} type.", describe(spread, checker)),
                ));
            } else if array_of(spread, TypeFlags::Any, checker) {
                let element = checker.get_type_arguments(spread)[0];
                let name = if error(element) {
                    "error".into()
                } else {
                    describe(spread, checker)
                };
                result.push(Violation::new(
                    argument,
                    "unsafeArraySpread",
                    format!("Unsafe spread of an {name} array type."),
                ));
            } else if is_tuple_type_exported(spread) {
                let elements = checker.get_type_arguments(spread).to_vec();
                for element in elements {
                    let Some(parameter) = signature.next() else {
                        continue;
                    };
                    if unsafe_assignment(element, parameter, None, checker, &mut Vec::new()) {
                        let sender = if error(element) {
                            "error typed".into()
                        } else {
                            format!("of type {}", describe(element, checker))
                        };
                        result.push(Violation::new(argument, "unsafeTupleSpread", format!("Unsafe spread of a tuple type. The argument is {sender} and is assigned to a parameter of type {}.", describe(parameter, checker))));
                    }
                }
                if spread
                    .target_tuple_type()
                    .combined_flags
                    .get()
                    .intersects(ElementFlags::Variable)
                {
                    signature.consumed = true;
                }
            }
            // Other iterable/array-union spreads do not advance the frozen cursor.
        } else if let Some(parameter) = signature.next() {
            let reporting = skip_parentheses(argument);
            let value = checker.get_type_at_location(reporting);
            if unsafe_assignment(value, parameter, Some(reporting), checker, &mut Vec::new()) {
                result.push(Violation::new(
                    reporting,
                    "unsafeArgument",
                    format!(
                        "Unsafe argument of type {} assigned to a parameter of type {}.",
                        describe(value, checker),
                        describe(parameter, checker)
                    ),
                ));
            }
        }
    }
    result
}
fn union_parts(t: P<Type>) -> Vec<P<Type>> {
    if t.flags().intersects(TypeFlags::Union) {
        t.types().to_vec()
    } else {
        vec![t]
    }
}
fn callback(parameter: P<Symbol>, node: P<Node>, checker: &mut Checker) -> bool {
    let Some(t) = checker.get_type_of_symbol_at_location(parameter, Some(node)) else {
        return false;
    };
    let mut t = checker.get_apparent_type(t);
    if parameter.value_declaration().is_some_and(|decl| {
        decl.kind() == Kind::Parameter
            && decl
                .as_parameter_declaration()
                .dot_dot_dot_token()
                .is_some()
    }) {
        let number = checker.get_number_type();
        let Some(element) = checker.get_index_type_of_type(t, number) else {
            return false;
        };
        t = element;
    }
    union_parts(t).into_iter().any(|part| {
        !checker
            .get_signatures_of_type(part, SignatureKind::Call)
            .is_empty()
    })
}
fn thenable(node: P<Node>, t: P<Type>, checker: &mut Checker) -> bool {
    let apparent = checker.get_apparent_type(t);
    for part in union_parts(apparent) {
        let Some(then) = checker.get_property_of_type(part, "then") else {
            continue;
        };
        let Some(t) = checker.get_type_of_symbol_at_location(then, Some(node)) else {
            continue;
        };
        for part in union_parts(t) {
            let signatures = checker
                .get_signatures_of_type(part, SignatureKind::Call)
                .to_vec();
            for signature in signatures {
                if signature
                    .parameters()
                    .first()
                    .is_some_and(|&param| callback(param, node, checker))
                {
                    return true;
                }
            }
        }
    }
    false
}
#[derive(Clone, Copy, PartialEq)]
enum AnyKind {
    Any,
    Array,
    Promise,
    Safe,
}
fn discriminate(
    t: P<Type>,
    node: P<Node>,
    checker: &mut Checker,
    visited: &mut Vec<P<Type>>,
) -> AnyKind {
    if visited.contains(&t) {
        return AnyKind::Safe;
    }
    visited.push(t);
    if any(t) {
        return AnyKind::Any;
    }
    if array_of(t, TypeFlags::Any, checker) {
        return AnyKind::Array;
    }
    fn promise(
        t: P<Type>,
        node: P<Node>,
        checker: &mut Checker,
        visited: &mut Vec<P<Type>>,
    ) -> bool {
        if t.flags().intersects(TypeFlags::UnionOrIntersection) {
            return t
                .types()
                .iter()
                .any(|&part| promise(part, node, checker, visited));
        }
        thenable(node, t, checker)
            && checker.get_awaited_type(t).is_some_and(|awaited| {
                discriminate(awaited, node, checker, visited) == AnyKind::Any
            })
    }
    if promise(t, node, checker, visited) {
        AnyKind::Promise
    } else {
        AnyKind::Safe
    }
}
fn contextual(node: P<Node>, checker: &mut Checker) -> Option<P<Type>> {
    let parent = node.parent()?;
    match parent.kind() {
        Kind::CallExpression | Kind::NewExpression => {
            if parent.expression() == Some(node) {
                return None;
            }
        }
        Kind::VariableDeclaration | Kind::PropertyDeclaration | Kind::Parameter => {
            return parent
                .type_node()
                .map(|ty| checker.get_type_from_type_node_exported(ty));
        }
        Kind::JsxExpression | Kind::TemplateSpan => {}
        Kind::PropertyAssignment | Kind::ShorthandPropertyAssignment
            if node.kind() == Kind::Identifier => {}
        Kind::BinaryExpression
            if parent.as_binary_expression().operator_token.kind() == Kind::EqualsToken
                && parent.as_binary_expression().right() == node =>
        {
            return Some(checker.get_type_at_location(parent.as_binary_expression().left));
        }
        _ => return None,
    }
    checker.get_contextual_type(node, ContextFlags::None)
}
fn collect_signatures(t: P<Type>, checker: &mut Checker) -> Vec<P<Signature>> {
    if t.flags().intersects(TypeFlags::Union) {
        return t
            .types()
            .iter()
            .flat_map(|&part| {
                checker
                    .get_signatures_of_type(part, SignatureKind::Call)
                    .to_vec()
            })
            .collect();
    }
    if t.flags().intersects(TypeFlags::Intersection) {
        let mut signatures = Vec::new();
        for &part in t.types() {
            let next = checker
                .get_signatures_of_type(part, SignatureKind::Call)
                .to_vec();
            if !next.is_empty() {
                if !signatures.is_empty() {
                    return Vec::new();
                }
                signatures = next;
            }
        }
        return signatures;
    }
    checker
        .get_signatures_of_type(t, SignatureKind::Call)
        .to_vec()
}
fn this_expression(mut node: P<Node>) -> Option<P<Node>> {
    loop {
        node = skip_parentheses(node);
        match node.kind() {
            Kind::ThisKeyword => return Some(node),
            Kind::CallExpression
            | Kind::PropertyAccessExpression
            | Kind::ElementAccessExpression => node = node.expression()?,
            _ => return None,
        }
    }
}
fn unsafe_return(node: P<Node>, checker: &mut Checker) -> Vec<Violation> {
    let (value_node, reporting) = match node.kind() {
        Kind::ArrowFunction => {
            let Some(body) = node.body() else {
                return Vec::new();
            };
            if body.kind() == Kind::Block {
                return Vec::new();
            }
            let body = skip_parentheses(body);
            (body, body)
        }
        Kind::ReturnStatement => {
            let Some(value) = node.expression() else {
                return Vec::new();
            };
            (value, node)
        }
        _ => return Vec::new(),
    };
    let mut parent = value_node.parent();
    let function = loop {
        let Some(candidate) = parent else {
            return Vec::new();
        };
        if ast::is_function_like_declaration(candidate) {
            break candidate;
        }
        parent = candidate.parent();
    };
    let value = checker.get_type_at_location(value_node);
    let primitive = TypeFlags::Primitive | TypeFlags::Unknown | TypeFlags::Never;
    let kind = if any(value) {
        AnyKind::Any
    } else if value.flags().intersects(primitive) {
        AnyKind::Safe
    } else {
        discriminate(value, value_node, checker, &mut Vec::new())
    };
    if kind == AnyKind::Safe && value.flags().intersects(primitive) {
        return Vec::new();
    }
    let function_type = if matches!(
        function.kind(),
        Kind::FunctionExpression | Kind::ArrowFunction
    ) {
        contextual(function, checker)
    } else {
        None
    }
    .unwrap_or_else(|| checker.get_type_at_location(function));
    let mut signatures = Vec::new();
    let mut loaded = false;
    let is_async = ast::has_syntactic_modifier(function, ModifierFlags::Async);
    if function.type_node().is_some() {
        signatures = collect_signatures(function_type, checker);
        loaded = true;
        for &signature in &signatures {
            let returned = checker.get_return_type_of_signature(signature);
            if returned == value
                || returned
                    .flags()
                    .intersects(TypeFlags::Any | TypeFlags::Unknown)
            {
                return Vec::new();
            }
            if is_async {
                let a = checker.get_awaited_type(returned);
                let b = checker.get_awaited_type(value);
                if a == b
                    || a.is_some_and(|t| t.flags().intersects(TypeFlags::Any | TypeFlags::Unknown))
                {
                    return Vec::new();
                }
            }
        }
    }
    if kind != AnyKind::Safe {
        if !loaded {
            signatures = collect_signatures(function_type, checker);
        }
        for signature in signatures {
            let returned = checker.get_return_type_of_signature(signature);
            if (kind == AnyKind::Any && unknown(returned))
                || (kind == AnyKind::Array && array_of(returned, TypeFlags::Unknown, checker))
                || (kind == AnyKind::Promise
                    && checker.get_awaited_type(returned).is_some_and(unknown))
            {
                return Vec::new();
            }
        }
        if kind == AnyKind::Promise && !is_async {
            return Vec::new();
        }
        let name = match kind {
            AnyKind::Any if error(constrained(value_node, checker)) => "error",
            AnyKind::Any => "`any`",
            AnyKind::Array => "`any[]`",
            AnyKind::Promise => "`Promise<any>`",
            AnyKind::Safe => unreachable!(),
        };
        if !checker.no_implicit_this
            && this_expression(value_node).is_some_and(|this| any(constrained(this, checker)))
        {
            return vec![Violation::new(
                reporting,
                "unsafeReturnThis",
                format!(
                    "Unsafe return of a value of type `{name}`. `this` is typed as `any`.\nYou can try to fix this by turning on the `noImplicitThis` compiler option, or adding a `this` parameter to the function."
                ),
            )];
        }
        return vec![Violation::new(
            reporting,
            "unsafeReturn",
            format!("Unsafe return of a value of type {name}."),
        )];
    }
    if !loaded
        || function_type
            .flags()
            .intersects(TypeFlags::UnionOrIntersection)
    {
        signatures = checker
            .get_signatures_of_type(function_type, SignatureKind::Call)
            .to_vec();
    }
    let Some(&signature) = signatures.first() else {
        return Vec::new();
    };
    let returned = checker.get_return_type_of_signature(signature);
    if unsafe_assignment(value, returned, Some(value_node), checker, &mut Vec::new()) {
        return vec![Violation::new(
            reporting,
            "unsafeReturnAssignment",
            format!(
                "Unsafe return of type `{}` from function with return type `{}`.",
                checker.type_to_string_exported(value),
                checker.type_to_string_exported(returned)
            ),
        )];
    }
    Vec::new()
}
