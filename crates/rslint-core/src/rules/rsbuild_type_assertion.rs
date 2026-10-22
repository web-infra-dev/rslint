//! Default no-unnecessary-type-assertion semantics from rslint
//! e0a29f751d4def86d27931759098f27748fa3003. Inference, recursive-signature,
//! readonly-property and source-edit guards deliberately preserve the baseline.
use crate::{Violation, edits::Edit};
use std::collections::HashSet;
use tsrs_ast::{
    self as ast, CheckFlags, Kind, ModifierFlags, Node, NodeFlags, SourceFile, SymbolFlags,
};
use tsrs_checker::{
    Checker, ContextFlags, MappedType, ObjectFlags, Signature, SignatureKind, Type, TypeFlags,
};
use tsrs_core::P;
use tsrs_scanner::{Scanner, skip_trivia};

pub(crate) const NAMES: &[&str] = &["@typescript-eslint/no-unnecessary-type-assertion"];
const UNNECESSARY: &str =
    "This assertion is unnecessary since it does not change the type of the expression.";
const CONTEXTUAL: &str =
    "This assertion is unnecessary since the receiver accepts the original type of the expression.";

pub(crate) fn check(
    _name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
) -> Vec<Violation> {
    match node.kind() {
        Kind::AsExpression | Kind::TypeAssertionExpression => {
            assertion(node, file, checker).into_iter().collect()
        }
        Kind::NonNullExpression => non_null(node, file, checker).into_iter().collect(),
        _ => vec![],
    }
}
fn end(node: P<Node>) -> usize {
    node.end().expect_source("parsed node end") as usize
}
fn pos(node: P<Node>) -> usize {
    node.pos().expect_source("parsed node position") as usize
}
fn start(node: P<Node>, file: P<SourceFile>) -> usize {
    skip_trivia(file.text(), pos(node) as u32) as usize
}
fn expr(node: P<Node>) -> P<Node> {
    node.expression().expect("assertion/expression operand")
}
fn walk_parens(mut node: P<Node>) -> P<Node> {
    while let Some(parent) = node.parent() {
        if parent.kind() != Kind::ParenthesizedExpression || expr(parent) != node {
            break;
        }
        node = parent;
    }
    node
}
fn original(node: P<Node>) -> P<Node> {
    let mut expression = ast::skip_parentheses(expr(node));
    while matches!(
        expression.kind(),
        Kind::AsExpression | Kind::TypeAssertionExpression
    ) {
        expression = ast::skip_parentheses(expr(expression));
    }
    expression
}
fn parts(ty: P<Type>) -> Vec<P<Type>> {
    if ty.flags().intersects(TypeFlags::Union) {
        ty.types().to_vec()
    } else {
        vec![ty]
    }
}
fn union_flags(ty: P<Type>) -> TypeFlags {
    parts(ty)
        .iter()
        .fold(TypeFlags::None, |flags, part| flags | part.flags())
}
fn nullable(ty: P<Type>) -> bool {
    union_flags(ty).intersects(
        TypeFlags::Any
            | TypeFlags::Unknown
            | TypeFlags::Null
            | TypeFlags::Undefined
            | TypeFlags::Void,
    )
}
fn literal(ty: P<Type>) -> bool {
    ty.flags()
        .intersects(TypeFlags::Literal | TypeFlags::BooleanLiteral)
}
fn arguments(checker: &mut Checker, ty: P<Type>) -> Vec<P<Type>> {
    if let Some(alias) = ty
        .alias()
        .filter(|alias| !alias.type_arguments().is_empty())
    {
        return alias.type_arguments().to_vec();
    }
    if ty.flags().intersects(TypeFlags::Object)
        && ty.object_flags().intersects(ObjectFlags::Reference)
    {
        checker.get_type_arguments(ty).to_vec()
    } else {
        vec![]
    }
}
fn signatures(checker: &mut Checker, ty: P<Type>) -> Vec<P<Signature>> {
    checker
        .get_signatures_of_type(ty, SignatureKind::Call)
        .to_vec()
}
fn contains_type_variable(checker: &mut Checker, root: P<Type>) -> bool {
    let mut seen = HashSet::new();
    let mut stack = vec![root];
    while let Some(ty) = stack.pop() {
        if !seen.insert(ty) {
            continue;
        }
        if ty
            .flags()
            .intersects(TypeFlags::TypeVariable | TypeFlags::Index)
        {
            return true;
        }
        if ty
            .flags()
            .intersects(TypeFlags::Union | TypeFlags::Intersection)
        {
            stack.extend_from_slice(ty.types());
            continue;
        }
        stack.extend(arguments(checker, ty));
        for signature in signatures(checker, ty) {
            stack.push(checker.get_return_type_of_signature(signature));
            for &parameter in signature.parameters() {
                stack.push(checker.get_type_of_symbol(parameter));
            }
        }
    }
    false
}
#[derive(Clone, Copy, PartialEq, Eq, Hash)]
enum SignatureIdentity {
    Declaration(P<Node>),
    Signature(P<Signature>),
}
fn identity(mut signature: P<Signature>) -> SignatureIdentity {
    loop {
        if let Some(declaration) = signature.declaration() {
            return SignatureIdentity::Declaration(declaration);
        }
        let Some(target) = signature.target().filter(|&target| target != signature) else {
            return SignatureIdentity::Signature(signature);
        };
        signature = target;
    }
}
fn contains_any(checker: &mut Checker, root: P<Type>) -> bool {
    enum Frame {
        Type(P<Type>),
        Signature(P<Signature>),
        Exit(SignatureIdentity),
    }
    let mut seen = HashSet::new();
    let mut active = HashSet::new();
    let mut expanded = 0;
    let mut stack = vec![Frame::Type(root)];
    while let Some(frame) = stack.pop() {
        match frame {
            Frame::Exit(identity) => {
                active.remove(&identity);
            }
            Frame::Signature(signature) => {
                let identity = identity(signature);
                expanded += 1;
                if active.contains(&identity) || expanded > 10_000 {
                    return true;
                }
                active.insert(identity);
                stack.push(Frame::Exit(identity));
                for &parameter in signature.parameters().iter().rev() {
                    stack.push(Frame::Type(checker.get_type_of_symbol(parameter)));
                }
                stack.push(Frame::Type(checker.get_return_type_of_signature(signature)));
            }
            Frame::Type(ty) => {
                if !seen.insert(ty) {
                    continue;
                }
                if ty.flags().intersects(TypeFlags::Any) {
                    return true;
                }
                if ty
                    .flags()
                    .intersects(TypeFlags::Union | TypeFlags::Intersection)
                {
                    stack.extend(ty.types().iter().rev().map(|&ty| Frame::Type(ty)));
                    continue;
                }
                stack.extend(
                    signatures(checker, ty)
                        .into_iter()
                        .rev()
                        .map(Frame::Signature),
                );
                stack.extend(arguments(checker, ty).into_iter().rev().map(Frame::Type));
            }
        }
    }
    false
}
fn has_index(checker: &mut Checker, ty: P<Type>) -> bool {
    parts(ty)
        .iter()
        .any(|&ty| !checker.get_index_infos_of_type(ty).is_empty())
}
fn readonly_property(checker: &mut Checker, ty: P<Type>, name: &str) -> bool {
    if ty.flags().intersects(TypeFlags::Intersection) {
        return ty.types().iter().any(|&part| {
            checker.get_property_of_type(part, name).is_some()
                && readonly_property(checker, part, name)
        });
    }
    if ty.object_flags().intersects(ObjectFlags::Mapped)
        && let Some(symbol) = ty.symbol()
        && let Some(&declaration) = symbol
            .declarations()
            .first()
            .filter(|declaration| declaration.kind() == Kind::MappedType)
    {
        if let Some(token) = declaration.as_mapped_type_node().readonly_token() {
            return token.kind() != Kind::MinusToken;
        }
        MappedType::resolve_components(checker, ty);
        let mapped = ty.as_mapped_type();
        if let Some(name_type) = mapped.name_type()
            && mapped
                .type_parameter()
                .is_none_or(|parameter| !checker.is_type_assignable_to(name_type, parameter))
        {
            return false;
        }
    }
    let Some(property) = checker.get_property_of_type(ty, name) else {
        return false;
    };
    if property.check_flags().intersects(CheckFlags::Readonly)
        || property
            .flags()
            .intersects(SymbolFlags::ValueModule | SymbolFlags::EnumMember)
        || (property.flags().intersects(SymbolFlags::GetAccessor)
            && !property.flags().intersects(SymbolFlags::SetAccessor))
    {
        return true;
    }
    for &declaration in property.declarations() {
        if declaration
            .modifier_flags()
            .intersects(ModifierFlags::Readonly)
            || (declaration.kind() == Kind::VariableDeclaration
                && declaration
                    .parent()
                    .is_some_and(|parent| parent.flags().intersects(NodeFlags::Constant)))
            || readonly_assignment(checker, declaration)
        {
            return true;
        }
    }
    tsrs_checker::get_declaration_modifier_flags_from_symbol_exported(property)
        .intersects(ModifierFlags::Readonly)
}
fn readonly_assignment(checker: &mut Checker, node: P<Node>) -> bool {
    if node.kind() != Kind::CallExpression || !ast::is_bindable_object_define_property_call(node) {
        return false;
    }
    let descriptor = checker.get_type_at_location(node.arguments()[2]);
    if checker.get_property_of_type(descriptor, "value").is_some() {
        let Some(writable) = checker.get_property_of_type(descriptor, "writable") else {
            return false;
        };
        let ty = if let Some(declaration) = writable
            .value_declaration()
            .filter(|node| node.kind() == Kind::PropertyAssignment)
        {
            checker.get_type_at_location(
                declaration
                    .initializer()
                    .expect("property assignment value"),
            )
        } else {
            checker.get_type_of_symbol(writable)
        };
        return ty.flags().intersects(TypeFlags::BooleanLiteral)
            && ty.as_literal_type().value() == Some(tsrs_checker::LiteralValue::Boolean(false));
    }
    checker.get_property_of_type(descriptor, "set").is_none()
}
fn same_properties(checker: &mut Checker, left: P<Type>, right: P<Type>) -> bool {
    let left_properties = checker.get_properties_of_type(left).to_vec();
    let right_properties = checker.get_properties_of_type(right).to_vec();
    left_properties.len() == right_properties.len()
        && left_properties.iter().all(|property| {
            right_properties
                .iter()
                .any(|other| other.name() == property.name())
                && readonly_property(checker, left, property.name())
                    == readonly_property(checker, right, property.name())
        })
}
fn generic_call(checker: &mut Checker, ty: P<Type>) -> bool {
    signatures(checker, ty)
        .iter()
        .any(|signature| !signature.type_parameters().is_empty())
}
fn argument_index(call: P<Node>, assertion: P<Node>) -> Option<usize> {
    if !matches!(call.kind(), Kind::CallExpression | Kind::NewExpression) {
        return None;
    }
    call.arguments()
        .iter()
        .position(|&argument| ast::skip_parentheses(argument) == assertion)
}
fn in_generic_context(node: P<Node>, checker: &mut Checker) -> bool {
    let mut child = walk_parens(node);
    let mut seen_function = false;
    while let Some(parent) = child.parent() {
        match parent.kind() {
            Kind::FunctionDeclaration => return false,
            Kind::MethodDeclaration | Kind::GetAccessor | Kind::SetAccessor | Kind::Constructor
                if parent.body() == Some(child) || child.kind() == Kind::Parameter =>
            {
                return false;
            }
            Kind::FunctionExpression | Kind::ArrowFunction => {
                if parent.body().is_some_and(|body| body.kind() == Kind::Block) || seen_function {
                    return false;
                }
                seen_function = true;
            }
            Kind::CallExpression | Kind::NewExpression => {
                if parent.type_argument_list().is_some() {
                    child = parent;
                    continue;
                }
                if parent.kind() == Kind::CallExpression
                    && matches!(
                        ast::skip_parentheses(expr(parent)).kind(),
                        Kind::PropertyAccessExpression | Kind::ElementAccessExpression
                    )
                    && argument_index(parent, node).is_some()
                {
                    child = parent;
                    continue;
                }
                let callee = checker.get_type_at_location(expr(parent));
                if generic_call(checker, callee) {
                    return true;
                }
            }
            _ => {}
        }
        child = parent;
    }
    false
}
fn conceptual_literal(node: P<Node>) -> bool {
    matches!(
        ast::skip_parentheses(node).kind(),
        Kind::NumericLiteral
            | Kind::StringLiteral
            | Kind::BigIntLiteral
            | Kind::RegularExpressionLiteral
            | Kind::TrueKeyword
            | Kind::FalseKeyword
            | Kind::NullKeyword
            | Kind::ArrayLiteralExpression
            | Kind::ObjectLiteralExpression
            | Kind::TemplateExpression
            | Kind::NoSubstitutionTemplateLiteral
            | Kind::ClassExpression
            | Kind::FunctionExpression
            | Kind::ArrowFunction
            | Kind::JsxElement
            | Kind::JsxSelfClosingElement
            | Kind::JsxFragment
    )
}
fn destructuring(node: P<Node>) -> bool {
    let semantic = walk_parens(node);
    semantic.parent().is_some_and(|parent| {
        parent.kind() == Kind::VariableDeclaration
            && parent.initializer() == Some(semantic)
            && parent.name().is_some_and(|name| {
                matches!(
                    name.kind(),
                    Kind::ObjectBindingPattern | Kind::ArrayBindingPattern
                )
            })
    })
}
fn problematic_property(node: P<Node>, checker: &mut Checker) -> bool {
    let semantic = walk_parens(node);
    let Some(property) = semantic.parent().filter(|parent| {
        parent.kind() == Kind::PropertyAssignment && parent.initializer() == Some(semantic)
    }) else {
        return false;
    };
    let Some(object) = property
        .parent()
        .filter(|node| node.kind() == Kind::ObjectLiteralExpression)
    else {
        return false;
    };
    if checker
        .get_contextual_type(object, ContextFlags::None)
        .is_some_and(|ty| ty.flags().intersects(TypeFlags::Union))
    {
        let Some(context) = checker.get_contextual_type(node, ContextFlags::None) else {
            return true;
        };
        let context = checker.get_non_nullable_type(context);
        if context.flags().intersects(TypeFlags::Union) {
            return true;
        }
        let uncast = checker.get_type_at_location(expr(node));
        return !checker.is_type_assignable_to(uncast, context);
    }
    let Some(parent) = walk_parens(object).parent() else {
        return false;
    };
    parent.kind() == Kind::SatisfiesExpression
        || (parent.kind() == Kind::CallExpression
            && walk_parens(parent)
                .parent()
                .is_some_and(|parent| parent.kind() == Kind::SatisfiesExpression))
}
fn contextual_from_parent(node: P<Node>, checker: &mut Checker) -> Option<P<Type>> {
    let parent = node.parent()?;
    match parent.kind() {
        Kind::CallExpression | Kind::NewExpression if expr(parent) == node => return None,
        Kind::VariableDeclaration | Kind::PropertyDeclaration | Kind::Parameter => {
            return parent
                .type_node()
                .map(|node| checker.get_type_from_type_node(node));
        }
        Kind::BinaryExpression
            if parent.as_binary_expression().operator_token().kind() == Kind::EqualsToken
                && parent.as_binary_expression().right() == node =>
        {
            return Some(checker.get_type_at_location(parent.as_binary_expression().left()));
        }
        Kind::JsxExpression | Kind::TemplateSpan | Kind::CallExpression | Kind::NewExpression => {}
        Kind::PropertyAssignment | Kind::ShorthandPropertyAssignment
            if node.kind() == Kind::Identifier => {}
        _ => return None,
    }
    checker.get_contextual_type(node, ContextFlags::None)
}
fn contextual(node: P<Node>, checker: &mut Checker) -> Option<P<Type>> {
    checker
        .get_contextual_type(node, ContextFlags::None)
        .or_else(|| {
            if walk_parens(node).parent().is_some_and(|parent| {
                matches!(parent.kind(), Kind::Parameter | Kind::PropertyDeclaration)
            }) {
                contextual_from_parent(node, checker)
            } else {
                None
            }
        })
}
fn overloaded_argument(node: P<Node>, checker: &mut Checker) -> bool {
    let Some(parent) = walk_parens(node).parent() else {
        return false;
    };
    let Some(index) = argument_index(parent, node) else {
        return false;
    };
    let callee = checker.get_type_at_location(expr(parent));
    let callee = checker.get_non_nullable_type(callee);
    let signatures = signatures(checker, callee);
    if signatures.len() <= 1 {
        return false;
    }
    let mut types = vec![];
    for signature in signatures {
        let Some(&parameter) = signature.parameters().get(index) else {
            return true;
        };
        let mut ty = checker.get_type_of_symbol(parameter);
        if parameter.value_declaration().is_some_and(|node| {
            node.kind() == Kind::Parameter
                && node
                    .as_parameter_declaration()
                    .dot_dot_dot_token()
                    .is_some()
        }) && let Some(&argument) = arguments(checker, ty).first()
        {
            ty = argument;
        }
        types.push(ty);
    }
    if types.iter().all(|ty| *ty == types[0]) {
        return false;
    }
    let uncast = checker.get_type_at_location(expr(node));
    types
        .into_iter()
        .any(|ty| !checker.is_type_assignable_to(uncast, ty))
}
fn generics_mismatch(checker: &mut Checker, uncast: P<Type>, context: P<Type>) -> bool {
    for &property in checker.get_properties_of_type(context) {
        let ty = checker.get_type_of_symbol(property);
        if !generic_call(checker, ty) {
            continue;
        }
        let Some(original) = checker.get_property_of_type(uncast, property.name()) else {
            return true;
        };
        let original = checker.get_type_of_symbol(original);
        if !generic_call(checker, original) {
            return true;
        }
    }
    false
}
fn empty_object(checker: &mut Checker, ty: P<Type>) -> bool {
    ty.flags().intersects(TypeFlags::NonPrimitive)
        || (checker.get_properties_of_type(ty).is_empty()
            && signatures(checker, ty).is_empty()
            && checker
                .get_signatures_of_type(ty, SignatureKind::Construct)
                .is_empty()
            && checker.get_string_index_type(ty).is_none()
            && checker.get_number_index_type(ty).is_none())
}
fn phantom(checker: &mut Checker, ty: P<Type>) -> bool {
    empty_object(checker, ty) && !arguments(checker, ty).is_empty()
}
fn phantom_mismatch(
    node: P<Node>,
    checker: &mut Checker,
    uncast: P<Type>,
    context: P<Type>,
) -> bool {
    in_generic_context(node, checker)
        && (phantom(checker, uncast) || phantom(checker, context))
        && arguments(checker, uncast) != arguments(checker, context)
}
fn logical(kind: Kind) -> bool {
    matches!(
        kind,
        Kind::BarBarToken | Kind::AmpersandAmpersandToken | Kind::QuestionQuestionToken
    )
}
fn skip_contextual(node: P<Node>, checker: &mut Checker, cast_any: bool) -> bool {
    let semantic = walk_parens(node);
    let Some(parent) = semantic.parent() else {
        return true;
    };
    if cast_any {
        return (parent.kind() == Kind::BinaryExpression
            && logical(parent.as_binary_expression().operator_token().kind()))
            || in_generic_context(node, checker);
    }
    if ast::skip_parentheses(expr(node)).kind() == Kind::TemplateExpression {
        return true;
    }
    if matches!(
        parent.kind(),
        Kind::AsExpression
            | Kind::TypeAssertionExpression
            | Kind::SpreadElement
            | Kind::SpreadAssignment
            | Kind::SatisfiesExpression
    ) {
        return true;
    }
    if parent.kind() == Kind::BinaryExpression {
        let binary = parent.as_binary_expression();
        if binary.right() == semantic
            && (matches!(
                binary.operator_token().kind(),
                Kind::BarBarEqualsToken
                    | Kind::AmpersandAmpersandEqualsToken
                    | Kind::QuestionQuestionEqualsToken
            ) || (binary.operator_token().kind() == Kind::EqualsToken
                && walk_parens(parent)
                    .parent()
                    .is_none_or(|parent| parent.kind() != Kind::ExpressionStatement)))
        {
            return true;
        }
    }
    if ast::skip_parentheses(expr(node)).kind() == Kind::ArrayLiteralExpression
        || destructuring(node)
        || problematic_property(node, checker)
        || overloaded_argument(node, checker)
    {
        return true;
    }
    in_generic_context(node, checker)
        && !conceptual_literal(original(node))
        && parent.kind() != Kind::PropertyAssignment
}
fn unchanged(
    node: P<Node>,
    expression: P<Node>,
    uncast: P<Type>,
    cast: P<Type>,
    checker: &mut Checker,
) -> bool {
    if uncast == cast {
        return true;
    }
    if node
        .type_node()
        .is_some_and(|node| node.kind() == Kind::IntersectionType)
        && contains_type_variable(checker, cast)
    {
        return false;
    }
    if union_flags(uncast).intersects(TypeFlags::Undefined)
        && union_flags(cast).intersects(TypeFlags::Undefined)
        && checker.exact_optional_property_types
    {
        let uncast: HashSet<_> = parts(uncast)
            .into_iter()
            .filter(|ty| !ty.flags().intersects(TypeFlags::Undefined))
            .collect();
        let cast = parts(cast);
        let cast: Vec<_> = cast
            .into_iter()
            .filter(|ty| !ty.flags().intersects(TypeFlags::Undefined))
            .collect();
        return uncast.len() == cast.len() && cast.iter().all(|ty| uncast.contains(ty));
    }
    if (union_flags(uncast).intersects(TypeFlags::NonPrimitive)
        && !union_flags(cast).intersects(TypeFlags::NonPrimitive))
        || has_index(checker, uncast) != has_index(checker, cast)
        || contains_any(checker, uncast)
        || contains_any(checker, cast)
        || (contains_type_variable(checker, cast) && !contains_type_variable(checker, uncast))
    {
        return false;
    }
    let expression = ast::skip_parentheses(expression);
    if conceptual_literal(expression) {
        if expression.kind() != Kind::ObjectLiteralExpression || expression.properties().is_empty()
        {
            return false;
        }
        for &property in checker.get_properties_of_type(cast) {
            if literal(checker.get_type_of_symbol(property)) {
                return false;
            }
        }
    }
    if cast.flags().intersects(TypeFlags::Intersection)
        && !uncast.flags().intersects(TypeFlags::Intersection)
    {
        if uncast.flags().intersects(TypeFlags::TypeParameter) && cast.types().len() == 2 {
            let parts = cast.types();
            let other = if parts[0] == uncast {
                Some(parts[1])
            } else if parts[1] == uncast {
                Some(parts[0])
            } else {
                None
            };
            if let Some(other) = other
                && empty_object(checker, other)
                && !contains_type_variable(checker, other)
                && checker
                    .get_base_constraint_of_type(uncast)
                    .is_some_and(|constraint| !nullable(constraint))
            {
                return true;
            }
        }
        return false;
    }
    same_properties(checker, uncast, cast)
        && arguments(checker, uncast) == arguments(checker, cast)
        && checker.is_type_assignable_to(uncast, cast)
        && checker.is_type_assignable_to(cast, uncast)
}
fn uncast_type(node: P<Node>, checker: &mut Checker) -> P<Type> {
    let expression = expr(node);
    if expression.kind() == Kind::CallExpression {
        let callee = ast::skip_parentheses(expr(expression));
        if matches!(
            callee.kind(),
            Kind::ArrowFunction | Kind::FunctionExpression
        ) {
            let ty = checker.get_type_at_location(callee);
            if let Some(signature) = signatures(checker, ty).first() {
                let returned = checker.get_return_type_of_signature(*signature);
                return if callee.type_node().is_none()
                    && returned.flags().intersects(TypeFlags::Undefined)
                {
                    checker.get_void_type()
                } else {
                    returned
                };
            }
        }
    }
    checker.get_type_at_location(expression)
}
fn implicitly_narrowed(node: P<Node>) -> bool {
    if expr(node).kind() == Kind::TemplateExpression {
        return false;
    }
    walk_parens(node).parent().is_some_and(|parent| {
        (parent.kind() == Kind::VariableDeclaration
            && parent.parent().is_some_and(|parent| {
                parent.kind() == Kind::VariableDeclarationList
                    && parent.flags().intersects(NodeFlags::Const)
            }))
            || (parent.kind() == Kind::PropertyDeclaration
                && parent.modifier_flags().intersects(ModifierFlags::Readonly))
    })
}
fn token(file: P<SourceFile>, position: usize) -> [usize; 2] {
    let mut scanner = Scanner::new();
    scanner.set_text(file.text());
    scanner.set_language_variant(file.language_variant());
    scanner.reset_pos(position as u32);
    scanner.scan();
    [scanner.token_start() as usize, scanner.token_end() as usize]
}
fn report(node: P<Node>, file: P<SourceFile>, contextual: bool) -> Violation {
    Violation::new(
        node,
        if contextual {
            "contextuallyUnnecessary"
        } else {
            "unnecessaryAssertion"
        },
        if contextual { CONTEXTUAL } else { UNNECESSARY },
    )
    .with_byte_range([token(file, pos(node))[0], end(node)])
}
fn starts_expression_statement(node: P<Node>, file: P<SourceFile>) -> bool {
    let begin = start(node, file);
    let mut current = node;
    while let Some(parent) = current.parent() {
        if parent.kind() == Kind::ParenthesizedExpression {
            return false;
        }
        if parent.kind() == Kind::ExpressionStatement {
            return start(parent, file) == begin;
        }
        if !ast::is_expression(parent) {
            return false;
        }
        current = parent;
    }
    false
}
fn starts_arrow_body(node: P<Node>, file: P<SourceFile>) -> bool {
    let begin = start(node, file);
    let mut current = node;
    while let Some(parent) = current.parent() {
        if parent.kind() == Kind::ParenthesizedExpression {
            return false;
        }
        if parent.kind() == Kind::ArrowFunction {
            return parent.body() == Some(current);
        }
        if !ast::is_expression(parent) || start(parent, file) != begin {
            return false;
        }
        current = parent;
    }
    false
}
fn js_whitespace(ch: char) -> bool {
    matches!(ch, '\u{0009}'..='\u{000d}' | '\u{0020}' | '\u{00a0}' | '\u{1680}' | '\u{2000}'..='\u{200a}' | '\u{2028}' | '\u{2029}' | '\u{202f}' | '\u{205f}' | '\u{3000}' | '\u{feff}')
}
fn assertion_report(node: P<Node>, file: P<SourceFile>, contextual: bool) -> Violation {
    let expression = expr(node);
    let ty = node.type_node().expect("assertion type");
    let mut edits = vec![];
    if node.kind() == Kind::AsExpression {
        let keyword = token(file, end(expression));
        let mut begin = keyword[0];
        for (offset, ch) in file.text()[end(expression)..keyword[0]]
            .char_indices()
            .rev()
        {
            if !js_whitespace(ch) {
                break;
            }
            begin = end(expression) + offset;
        }
        edits.push(Edit::remove([begin, end(ty)]));
    } else {
        edits.push(Edit::remove([token(file, pos(node))[0], pos(expression)]));
        let first = token(file, pos(expression));
        let text = &file.text()[first[0]..first[1]];
        if (["{", "function", "class"].contains(&text) && starts_expression_statement(node, file))
            || (text == "{" && starts_arrow_body(node, file))
        {
            edits.push(Edit::replace([start(node, file); 2], "("));
            edits.push(Edit::replace([end(node); 2], ")"));
        }
    }
    report(node, file, contextual).with_fix(edits)
}
fn assertion(node: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> Option<Violation> {
    let type_node = node.type_node()?;
    let is_const = type_node.kind() == Kind::TypeReference
        && type_node.as_type_reference_node().type_name().kind() == Kind::Identifier
        && type_node.as_type_reference_node().type_name().text() == "const";
    let cast = checker.get_type_at_location(node);
    let cast_literal = literal(cast);
    if cast_literal && is_const {
        return None;
    }
    let expression = expr(node);
    let uncast = uncast_type(node, checker);
    let same_inferred = if cast_literal {
        implicitly_narrowed(node)
    } else {
        !is_const
    };
    if unchanged(node, expression, uncast, cast, checker) && same_inferred {
        return Some(assertion_report(node, file, false));
    }
    let parent = walk_parens(node).parent();
    let parent_skips_any = parent.is_some_and(|parent| {
        matches!(
            parent.kind(),
            Kind::AsExpression
                | Kind::TypeAssertionExpression
                | Kind::SpreadElement
                | Kind::SpreadAssignment
                | Kind::SatisfiesExpression
        )
    });
    let cast_any = cast.flags().intersects(TypeFlags::Any) && !parent_skips_any;
    let context = if skip_contextual(node, checker, cast_any) {
        None
    } else {
        contextual(node, checker)
    };
    if let Some(context) = context {
        let context_any = context.flags().intersects(TypeFlags::Any);
        let call_argument = parent.is_some_and(|parent| argument_index(parent, node).is_some());
        let allowed_any = if context_any {
            call_argument && !contains_any(checker, cast)
        } else {
            !contains_any(checker, context)
        };
        let original = ast::skip_parentheses(expression);
        let nullish_to_union = cast.flags().intersects(TypeFlags::Union)
            && (original.kind() == Kind::NullKeyword
                || (original.kind() == Kind::Identifier && original.text() == "undefined"));
        if !is_const
            && !contains_any(checker, uncast)
            && allowed_any
            && !phantom_mismatch(node, checker, uncast, context)
            && (cast_any || !generics_mismatch(checker, uncast, context))
            && (context_any || checker.is_type_assignable_to(uncast, context))
            && !nullish_to_union
        {
            return Some(assertion_report(node, file, true));
        }
    }
    let inner = ast::skip_parentheses(expression);
    if !matches!(
        inner.kind(),
        Kind::AsExpression | Kind::TypeAssertionExpression
    ) {
        return None;
    }
    let original = original(node);
    let original_type = checker.get_type_at_location(original);
    let contextual = if unchanged(node, inner, original_type, cast, checker)
        && !cast.flags().intersects(TypeFlags::Any)
    {
        false
    } else {
        let context = context?;
        let intermediate = checker.get_type_at_location(inner);
        if !intermediate
            .flags()
            .intersects(TypeFlags::Any | TypeFlags::Unknown)
            || !checker.is_type_assignable_to(original_type, context)
        {
            return None;
        }
        true
    };
    let mut replacement = file.text()[start(original, file)..end(original)].to_string();
    let semantic = walk_parens(node);
    if original.kind() == Kind::ObjectLiteralExpression
        && semantic.parent().is_some_and(|parent| {
            parent.kind() == Kind::ArrowFunction && parent.body() == Some(semantic)
        })
    {
        replacement = format!("({replacement})");
    }
    Some(report(node, file, contextual).with_fix(vec![Edit::replace(
        [start(node, file), end(node)],
        replacement,
    )]))
}
fn possibly_unassigned(
    node: P<Node>,
    declaration: Option<P<Node>>,
    constrained: Option<P<Type>>,
    checker: &mut Checker,
) -> bool {
    let Some(declaration) = declaration else {
        return true;
    };
    if !checker.strict_null_checks || declaration.kind() != Kind::VariableDeclaration {
        return false;
    }
    if declaration.parent().is_some_and(|parent| {
        parent.kind() == Kind::VariableDeclarationList && parent.flags() == NodeFlags::None
    }) {
        let declaration_scope = ast::get_enclosing_block_scope_container(declaration);
        let scope = ast::get_enclosing_block_scope_container(node);
        let mut parent = declaration_scope.and_then(ast::get_enclosing_block_scope_container);
        while let Some(current) = parent {
            if Some(current) == scope {
                return true;
            }
            parent = ast::get_enclosing_block_scope_container(current);
        }
    }
    let variable = declaration.as_variable_declaration();
    if variable.initializer().is_none()
        && variable.exclamation_token().is_none()
        && let Some(type_node) = variable.type_()
    {
        let declared = checker.get_type_from_type_node(type_node);
        let constrained = constrained.unwrap_or_else(|| {
            let ty = checker.get_type_at_location(node);
            checker.get_base_constraint_of_type(ty).unwrap_or(ty)
        });
        let declared_keyword = declaration
            .parent()
            .filter(|parent| parent.kind() == Kind::VariableDeclarationList)
            .and_then(|parent| parent.parent())
            .is_some_and(|parent| {
                parent.kind() == Kind::VariableStatement
                    && parent.modifier_flags().intersects(ModifierFlags::Ambient)
            });
        if declared == constrained && !declared_keyword {
            return true;
        }
    }
    false
}
fn non_null(node: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> Option<Violation> {
    let expression = expr(node);
    let finish = |contextual| {
        report(node, file, contextual).with_fix(vec![Edit::remove(token(file, end(expression)))])
    };
    if let Some(parent) = node
        .parent()
        .filter(|&parent| ast::is_assignment_expression(parent, true))
    {
        return (parent.as_binary_expression().left() == node).then(|| finish(true));
    }
    let symbol = if expression.kind() == Kind::Identifier {
        checker.get_symbol_at_location_exported(expression)
    } else {
        None
    };
    let declaration = symbol.and_then(|symbol| symbol.declarations().first().copied());
    if let Some(symbol) = symbol
        && let Some(declared) = checker.get_type_of_symbol_at_location(symbol, None)
    {
        let constrained = checker
            .get_base_constraint_of_type(declared)
            .unwrap_or(declared);
        if declaration.is_some() && !nullable(constrained) {
            return (!possibly_unassigned(expression, declaration, None, checker))
                .then(|| finish(false));
        }
    }
    let actual = checker.get_type_at_location(expression);
    let constrained = checker
        .get_base_constraint_of_type(actual)
        .unwrap_or(actual);
    if !nullable(constrained) && !nullable(actual) {
        if expression.kind() == Kind::Identifier
            && possibly_unassigned(expression, declaration, Some(constrained), checker)
        {
            return None;
        }
        return Some(finish(false));
    }
    if constrained != actual {
        return None;
    }
    let context = contextual_from_parent(node, checker)?;
    let flags = union_flags(constrained);
    let context_flags = union_flags(context);
    if flags.intersects(TypeFlags::Unknown) && !context_flags.intersects(TypeFlags::Unknown) {
        return None;
    }
    for flag in [TypeFlags::Undefined, TypeFlags::Null, TypeFlags::Void] {
        if flags.intersects(flag) && !context_flags.intersects(flag) {
            return None;
        }
    }
    Some(finish(true))
}
