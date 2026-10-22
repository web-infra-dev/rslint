//! Frozen default-option error preservation and array-existence checks.
//! Ported from rslint e0a29f751d4def86d27931759098f27748fa3003. Decisions
//! use parsed nodes, binder identities and the compiler's actual types.
use std::collections::{BTreeMap, HashSet};

use tsrs_ast::{
    self as ast, Kind, ModifierFlags, Node, NodeFlags, SourceFile, Symbol, SymbolFlags,
};
use tsrs_checker::{Checker, Type, TypeFlags};
use tsrs_core::{P, jsnum};
use tsrs_scanner::skip_trivia;

use super::rsbuild_assignment_checks::resolve_reference;
use super::rsbuild_ts_syntax::lexically_shadowed;
use crate::{Violation, edits::Edit};

pub(crate) const NAMES: &[&str] = &["preserve-caught-error", "unicorn/prefer-array-some"];

pub(crate) fn check(
    name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    has_type_information: bool,
) -> Vec<Violation> {
    let result = match (name, node.kind()) {
        ("preserve-caught-error", Kind::ThrowStatement) => {
            preserve_caught_error(node, file, checker)
        }
        ("unicorn/prefer-array-some", Kind::CallExpression) => ArraySome {
            file,
            checker,
            has_type_information,
        }
        .find_call(node),
        ("unicorn/prefer-array-some", Kind::BinaryExpression) => {
            let mut context = ArraySome {
                file,
                checker,
                has_type_information,
            };
            context
                .find_index(node)
                .or_else(|| context.filter_length(node))
        }
        _ => None,
    };
    result.into_iter().collect()
}
fn start(node: P<Node>, file: P<SourceFile>) -> usize {
    skip_trivia(file.text(), node.pos().expect_source("parsed node start")) as usize
}
fn end(node: P<Node>) -> usize {
    node.end().expect_source("parsed node end") as usize
}
fn range(node: P<Node>, file: P<SourceFile>) -> [usize; 2] {
    [start(node, file), end(node)]
}
fn insert(pos: usize, text: impl Into<String>) -> Edit {
    Edit::replace([pos, pos], text)
}
fn outer_parentheses(mut node: P<Node>) -> P<Node> {
    while let Some(parent) = node
        .parent()
        .filter(|p| p.kind() == Kind::ParenthesizedExpression)
    {
        node = parent;
    }
    node
}
fn effective_parent(node: P<Node>) -> Option<P<Node>> {
    outer_parentheses(node).parent()
}
fn global(node: P<Node>) -> bool {
    node.kind() == Kind::Identifier && !lexically_shadowed(node, node.text())
}

fn global_error_constructor(node: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> bool {
    if global(node) {
        return true;
    }
    if node.kind() != Kind::Identifier {
        return false;
    }
    // The shared lexical helper also models type-space bindings. This rule's
    // frozen IsShadowed check only un-declares a built-in for value bindings:
    // an authored `interface Error {}` may merge with the library interface
    // without replacing the Error constructor. Resolve the value-space symbol
    // before deciding whether an otherwise-shadowed name has a local binding.
    resolve_reference(node, checker).is_none_or(|symbol| {
        symbol.declarations().iter().all(|&declaration| {
            ast::get_source_file_of_node(declaration) != Some(file)
                || matches!(
                    declaration.kind(),
                    Kind::InterfaceDeclaration | Kind::TypeAliasDeclaration
                )
        })
    })
}

// Only a direct rethrow from the catch participates. Function and static-block
// boundaries are intentionally not crossed, even when they capture the binding.
fn parent_catch(mut node: P<Node>) -> Option<P<Node>> {
    loop {
        if node.kind() == Kind::CatchClause {
            return Some(node);
        }
        if ast::is_function_like_or_class_static_block_declaration(node)
            || node.kind() == Kind::SourceFile
        {
            return None;
        }
        node = node.parent()?;
    }
}
fn preserve_caught_error(
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
) -> Option<Violation> {
    let catch = parent_catch(node)?;
    let expression = ast::skip_parentheses(node.as_throw_statement().expression());
    if !matches!(
        expression.kind(),
        Kind::CallExpression | Kind::NewExpression
    ) || ast::is_optional_chain(expression)
    {
        return None;
    }
    let callee = ast::skip_parentheses(expression.expression()?);
    if ast::is_optional_chain(callee)
        || !global_error_constructor(callee, file, checker)
        || !matches!(
            callee.text(),
            "Error"
                | "EvalError"
                | "RangeError"
                | "ReferenceError"
                | "SyntaxError"
                | "TypeError"
                | "URIError"
                | "AggregateError"
        )
    {
        return None;
    }
    let parameter = catch.as_catch_clause().variable_declaration()?;
    let parameter_name = parameter.name()?;
    if parameter_name.kind() != Kind::Identifier {
        return Some(Violation::new(
            catch,
            "partiallyLostError",
            "Re-throws cannot preserve the caught error as a part of it is being lost due to destructuring.",
        ));
    }
    let index = if callee.text() == "AggregateError" {
        2
    } else {
        1
    };
    let args = expression.arguments();
    if args
        .iter()
        .enumerate()
        .any(|(position, arg)| position <= index && arg.kind() == Kind::SpreadElement)
    {
        return None;
    }
    let mut cause = None;
    let mut count = 0;
    if let Some(&options) = args.get(index) {
        let options = ast::skip_parentheses(options);
        if options.kind() != Kind::ObjectLiteralExpression {
            return None;
        }
        for &member in options.as_object_literal_expression().properties().nodes() {
            if member.kind() == Kind::SpreadAssignment {
                return None;
            }
            if member
                .name()
                .and_then(|name| super::rsbuild_declaration_checks::static_property(name, file))
                .as_deref()
                == Some("cause")
            {
                cause = Some(member);
                count += 1;
            }
        }
    }
    let Some(member) = cause else {
        let options = format!("{{ cause: {} }}", parameter_name.text());
        let edits = if let Some(&options_node) = args.get(index) {
            let properties = ast::skip_parentheses(options_node)
                .as_object_literal_expression()
                .properties();
            if let Some(&last) = properties.nodes().last() {
                vec![insert(
                    end(last),
                    format!(", cause: {}", parameter_name.text()),
                )]
            } else {
                vec![insert(
                    properties.pos().expect_source("object members") as usize,
                    format!("cause: {}", parameter_name.text()),
                )]
            }
        } else {
            let suffix = match (index, args.len()) {
                (2, 0) => format!("[], \"\", {options}"),
                (2, 1) => format!(", \"\", {options}"),
                (1, 0) => format!("\"\", {options}"),
                _ => format!(", {options}"),
            };
            if let Some(&last) = args.last() {
                vec![insert(end(last), suffix)]
            } else {
                let list = if expression.kind() == Kind::NewExpression {
                    expression.as_new_expression().arguments()
                } else {
                    Some(expression.as_call_expression().arguments())
                };
                if let Some(list) = list {
                    vec![insert(
                        list.pos().expect_source("call arguments") as usize,
                        suffix,
                    )]
                } else {
                    vec![insert(end(expression), format!("({suffix})"))]
                }
            }
        };
        return Some(
            Violation::new(
                node,
                "missingCause",
                "There is no `cause` attached to the symptom error being thrown.",
            )
            .with_suggestion(
                "includeCause",
                "Include the original caught error as the `cause` of the symptom error.",
                edits,
            ),
        );
    };
    let value = match member.kind() {
        Kind::PropertyAssignment => Some(ast::skip_parentheses(
            member.as_property_assignment().initializer(),
        )),
        Kind::ShorthandPropertyAssignment => member.name(),
        _ => None,
    };
    if value.is_none_or(|value| {
        value.kind() != Kind::Identifier || value.text() != parameter_name.text()
    }) {
        let report_range = if let Some(value) = value {
            range(value, file)
        } else {
            [
                member.name().map_or(end(member), |name| {
                    skip_trivia(file.text(), name.end().expect_source("member name end")) as usize
                }),
                end(member),
            ]
        };
        let mut report = Violation::new(
            member,
            "incorrectCause",
            "The symptom error is being thrown with an incorrect `cause`.",
        )
        .with_byte_range(report_range);
        if count == 1 {
            let edit = if matches!(
                member.kind(),
                Kind::ShorthandPropertyAssignment
                    | Kind::MethodDeclaration
                    | Kind::GetAccessor
                    | Kind::SetAccessor
            ) {
                Edit::replace(
                    range(member, file),
                    format!("cause: {}", parameter_name.text()),
                )
            } else {
                Edit::replace(range(value?, file), parameter_name.text())
            };
            report = report.with_suggestion(
                "includeCause",
                "Include the original caught error as the `cause` of the symptom error.",
                vec![edit],
            );
        }
        return Some(report);
    }
    if let (Some(caught), Some(target)) = (parameter.symbol(), resolve_reference(value?, checker))
        && caught != target
    {
        return Some(Violation::new(
            node,
            "caughtErrorShadowed",
            "The caught error is being attached as `cause`, but is shadowed by a closer scoped redeclaration.",
        ));
    }
    None
}

// ESTree discards parentheses and synthesized JSDoc casts, but not authored
// TypeScript assertions. Those wrappers remain significant to method matching.
fn jsdoc_wrapper(node: P<Node>) -> bool {
    ast::is_jsdoc_type_assertion(node)
        || (matches!(node.kind(), Kind::AsExpression | Kind::SatisfiesExpression)
            && node.parent().is_some_and(ast::is_jsdoc_type_assertion))
}
fn runtime(mut node: P<Node>) -> P<Node> {
    loop {
        if node.kind() == Kind::ParenthesizedExpression {
            node = node.as_parenthesized_expression().expression();
        } else if matches!(node.kind(), Kind::AsExpression | Kind::SatisfiesExpression)
            && jsdoc_wrapper(node)
        {
            node = node.expression().expect("assertion expression");
        } else {
            return node;
        }
    }
}
fn runtime_parent(node: P<Node>) -> Option<P<Node>> {
    let mut parent = node.parent()?;
    while parent.kind() == Kind::ParenthesizedExpression
        || (matches!(
            parent.kind(),
            Kind::AsExpression | Kind::SatisfiesExpression
        ) && jsdoc_wrapper(parent))
    {
        parent = parent.parent()?;
    }
    Some(parent)
}
#[derive(Clone, Copy)]
struct DotCall {
    call: P<Node>,
    object: P<Node>,
    property: P<Node>,
}
fn method_call(
    node: P<Node>,
    names: &[&str],
    minimum: usize,
    maximum: Option<usize>,
    reject_spread: bool,
) -> Option<DotCall> {
    if node.kind() != Kind::CallExpression || ast::is_optional_chain_root(node) {
        return None;
    }
    let args = node.arguments();
    if args.len() < minimum
        || maximum.is_some_and(|max| args.len() > max)
        || (reject_spread
            && maximum.is_some_and(|max| {
                args.iter()
                    .take(max)
                    .any(|arg| arg.kind() == Kind::SpreadElement)
            }))
    {
        return None;
    }
    let raw = node.as_call_expression().expression();
    let callee = runtime(raw);
    if (raw != callee && ast::is_optional_chain(callee))
        || callee.kind() != Kind::PropertyAccessExpression
        || ast::is_optional_chain_root(callee)
    {
        return None;
    }
    let access = callee.as_property_access_expression();
    let property = access.name();
    if property.kind() != Kind::Identifier || !names.contains(&property.text()) {
        return None;
    }
    Some(DotCall {
        call: node,
        object: access.expression(),
        property,
    })
}
fn logical(node: P<Node>) -> bool {
    node.kind() == Kind::BinaryExpression
        && matches!(
            node.as_binary_expression().operator_token().kind(),
            Kind::AmpersandAmpersandToken | Kind::BarBarToken
        )
}
fn not(node: P<Node>) -> bool {
    node.kind() == Kind::PrefixUnaryExpression
        && node.as_prefix_unary_expression().operator() == Kind::ExclamationToken
}
fn boolean_call(node: P<Node>) -> bool {
    if node.kind() != Kind::CallExpression
        || node.as_call_expression().question_dot_token().is_some()
    {
        return false;
    }
    let raw = node.as_call_expression().expression();
    let callee = runtime(raw);
    let args = node.arguments();
    !(raw != callee && ast::is_optional_chain(callee))
        && callee.kind() == Kind::Identifier
        && callee.text() == "Boolean"
        && global(callee)
        && args.len() == 1
        && args[0].kind() != Kind::SpreadElement
}
fn boolean_expression(mut node: P<Node>) -> bool {
    loop {
        let parent = runtime_parent(node);
        if not(node)
            || boolean_call(node)
            || parent.is_some_and(|parent| {
                not(parent) || (boolean_call(parent) && runtime(parent.arguments()[0]) == node)
            })
        {
            return true;
        }
        let Some(parent) = parent.filter(|parent| logical(*parent)) else {
            return false;
        };
        node = parent;
    }
}
fn control_flow(mut node: P<Node>) -> bool {
    loop {
        let Some(parent) = runtime_parent(node) else {
            return false;
        };
        let test = match parent.kind() {
            Kind::IfStatement => Some(parent.as_if_statement().expression()),
            Kind::WhileStatement => Some(parent.as_while_statement().expression()),
            Kind::DoStatement => Some(parent.as_do_statement().expression()),
            Kind::ForStatement => parent.as_for_statement().condition(),
            Kind::ConditionalExpression => Some(parent.as_conditional_expression().condition()),
            _ => None,
        };
        if test.is_some_and(|test| runtime(test) == node) {
            return true;
        }
        if !logical(parent) {
            return false;
        }
        node = parent;
    }
}
fn undefined(node: P<Node>) -> bool {
    node.kind() == Kind::Identifier && node.text() == "undefined"
}
fn comparison_parent(node: P<Node>) -> Option<P<Node>> {
    let parent = effective_parent(node)?;
    if parent.kind() != Kind::BinaryExpression {
        return None;
    }
    let binary = parent.as_binary_expression();
    if ast::skip_parentheses(binary.left()) != node {
        return None;
    }
    let matches = match binary.operator_token().kind() {
        Kind::EqualsEqualsToken | Kind::ExclamationEqualsToken => {
            undefined(binary.right())
                || ast::skip_parentheses(binary.right()).kind() == Kind::NullKeyword
        }
        Kind::EqualsEqualsEqualsToken | Kind::ExclamationEqualsEqualsToken => {
            undefined(binary.right())
        }
        _ => false,
    };
    matches.then_some(parent)
}
fn has_comment(file: P<SourceFile>, from: usize, to: usize) -> bool {
    // These spans contain only member/operator tokens and numeric/null/undefined
    // operands, never a string/template/regexp which could imitate a comment.
    let tail = &file.text()[from..to];
    tail.contains("//") || tail.contains("/*")
}
fn some_report(property: P<Node>) -> Violation {
    Violation::new(
        property,
        "some",
        format!("Prefer `.some(…)` over `.{}(…)`.", property.text()),
    )
}
fn numeric(node: P<Node>, number: f64) -> bool {
    let node = ast::skip_parentheses(node);
    node.kind() == Kind::NumericLiteral && jsnum::from_string(node.text()).0 == number
}
fn negative_one(node: P<Node>) -> bool {
    let node = ast::skip_parentheses(node);
    node.kind() == Kind::PrefixUnaryExpression
        && node.as_prefix_unary_expression().operator() == Kind::MinusToken
        && numeric(node.as_prefix_unary_expression().operand(), 1.0)
}
fn not_function(node: P<Node>) -> bool {
    match node.kind() {
        Kind::ArrayLiteralExpression
        | Kind::ObjectLiteralExpression
        | Kind::ClassExpression
        | Kind::TemplateExpression
        | Kind::NoSubstitutionTemplateLiteral
        | Kind::PrefixUnaryExpression
        | Kind::PostfixUnaryExpression
        | Kind::TypeOfExpression
        | Kind::VoidExpression
        | Kind::DeleteExpression
        | Kind::StringLiteral
        | Kind::NumericLiteral
        | Kind::BigIntLiteral
        | Kind::RegularExpressionLiteral
        | Kind::TrueKeyword
        | Kind::FalseKeyword
        | Kind::NullKeyword
        | Kind::AwaitExpression
        | Kind::NewExpression
        | Kind::TaggedTemplateExpression
        | Kind::ThisKeyword => true,
        Kind::BinaryExpression => !matches!(
            node.as_binary_expression().operator_token().kind(),
            Kind::AmpersandAmpersandToken
                | Kind::BarBarToken
                | Kind::QuestionQuestionToken
                | Kind::CommaToken
        ),
        Kind::CallExpression
            if !ast::is_optional_chain(node)
                && node.as_call_expression().expression().kind() != Kind::ImportKeyword =>
        {
            method_call(node, &["bind"], 0, None, false).is_none()
        }
        _ => undefined(node),
    }
}
struct ArraySome<'a> {
    file: P<SourceFile>,
    checker: &'a mut Checker,
    has_type_information: bool,
}
impl ArraySome<'_> {
    fn find_call(&mut self, node: P<Node>) -> Option<Violation> {
        let callee = ast::skip_parentheses(node.as_call_expression().expression());
        if callee.kind() != Kind::PropertyAccessExpression
            || !matches!(
                callee.as_property_access_expression().name().text(),
                "find" | "findLast"
            )
        {
            return None;
        }
        let comparison = comparison_parent(node);
        let boolean_use = comparison.is_some() || boolean_expression(node) || control_flow(node);
        if !boolean_use
            && !node.parent().is_some_and(|parent| {
                parent.kind() == Kind::VariableDeclaration
                    && parent.as_variable_declaration().initializer() == Some(node)
            })
        {
            return None;
        }
        let call = method_call(node, &["find", "findLast"], 1, Some(2), true)?;
        if !node.type_arguments().is_empty() {
            return None;
        }
        if boolean_use {
            if self.receiver(
                call.object,
                true,
                self.has_type_information,
                &mut HashSet::new(),
                0,
            ) == Class::NonTarget
            {
                return None;
            }
        } else if !self.variable_boolean(call) {
            return None;
        }
        let mut report = some_report(call.property);
        let call_end = end(outer_parentheses(node));
        if comparison.is_some_and(|comparison| has_comment(self.file, call_end, end(comparison))) {
            return Some(report);
        }
        let mut edits = vec![Edit::replace(range(call.property, self.file), "some")];
        if let Some(comparison) = comparison {
            edits.push(Edit::replace([call_end, end(comparison)], ""));
            if !matches!(
                comparison.as_binary_expression().operator_token().kind(),
                Kind::ExclamationEqualsToken | Kind::ExclamationEqualsEqualsToken
            ) {
                edits.push(insert(start(outer_parentheses(node), self.file), "!"));
            }
        }
        report = report.with_suggestion(
            "some-suggestion",
            format!("Replace `.{}(…)` with `.some(…)`.", call.property.text()),
            edits,
        );
        report.suggestions[0].data = Some(BTreeMap::from([(
            "method".to_owned(),
            call.property.text().to_owned(),
        )]));
        Some(report)
    }
    fn variable_boolean(&mut self, call: DotCall) -> bool {
        let Some(declaration) = call
            .call
            .parent()
            .filter(|parent| parent.kind() == Kind::VariableDeclaration)
        else {
            return false;
        };
        let variable = declaration.as_variable_declaration();
        if variable.initializer() != Some(call.call)
            || variable.name().kind() != Kind::Identifier
            || variable.type_().is_some()
        {
            return false;
        }
        let Some(list) = declaration.parent().filter(|parent| {
            parent.kind() == Kind::VariableDeclarationList
                && parent.flags().intersects(NodeFlags::Const)
        }) else {
            return false;
        };
        let Some(statement) = list
            .parent()
            .filter(|parent| parent.kind() == Kind::VariableStatement)
        else {
            return false;
        };
        if statement.modifier_flags().intersects(ModifierFlags::Export)
            || statement
                .parent()
                .is_some_and(|parent| parent.kind() == Kind::ExportDeclaration)
        {
            return false;
        }
        let typed_source = matches!(
            std::path::Path::new(self.file.file_name())
                .extension()
                .and_then(|ext| ext.to_str()),
            Some("ts" | "tsx" | "mts" | "cts")
        );
        if self.receiver(
            call.object,
            false,
            typed_source && self.has_type_information,
            &mut HashSet::new(),
            0,
        ) != Class::Target
        {
            return false;
        }
        let Some(symbol) = declaration.symbol() else {
            return false;
        };
        let mut stack = vec![self.file.as_node()];
        let mut count = 0;
        while let Some(node) = stack.pop() {
            node.for_each_child(&mut |child| {
                stack.push(child);
                false
            });
            if node.kind() != Kind::Identifier
                || node.text() != variable.name().text()
                || !is_reference(node)
                || resolve_reference(node, self.checker) != Some(symbol)
            {
                continue;
            }
            count += 1;
            if !boolean_expression(node) && !control_flow(node) {
                return false;
            }
        }
        count > 0
    }
    fn find_index(&mut self, node: P<Node>) -> Option<Violation> {
        let binary = node.as_binary_expression();
        let operator = binary.operator_token().kind();
        let matches = match operator {
            Kind::ExclamationEqualsEqualsToken
            | Kind::ExclamationEqualsToken
            | Kind::GreaterThanToken
            | Kind::EqualsEqualsEqualsToken
            | Kind::EqualsEqualsToken => negative_one(binary.right()),
            Kind::GreaterThanEqualsToken | Kind::LessThanToken => numeric(binary.right(), 0.0),
            _ => false,
        };
        if !matches {
            return None;
        }
        let call = method_call(
            binary.left(),
            &["findIndex", "findLastIndex"],
            1,
            Some(1),
            true,
        )?;
        if self.receiver(
            call.object,
            true,
            self.has_type_information,
            &mut HashSet::new(),
            0,
        ) == Class::NonTarget
        {
            return None;
        }
        let report = some_report(call.property);
        let operator_start = start(binary.operator_token(), self.file);
        if has_comment(self.file, operator_start, end(node)) {
            return Some(report);
        }
        let mut edits = Vec::new();
        if matches!(
            operator,
            Kind::EqualsEqualsEqualsToken | Kind::EqualsEqualsToken | Kind::LessThanToken
        ) {
            edits.push(insert(start(node, self.file), "!"));
        }
        edits.push(Edit::replace(range(call.property, self.file), "some"));
        edits.push(Edit::replace([operator_start, end(node)], ""));
        Some(report.with_fix(edits))
    }
    fn filter_length(&mut self, node: P<Node>) -> Option<Violation> {
        let binary = node.as_binary_expression();
        if !matches!(
            binary.operator_token().kind(),
            Kind::GreaterThanToken | Kind::ExclamationEqualsEqualsToken
        ) {
            return None;
        }
        let right = ast::skip_parentheses(binary.right());
        if right.kind() != Kind::NumericLiteral
            || &self.file.text()[start(right, self.file)..end(right)] != "0"
        {
            return None;
        }
        let length = ast::skip_parentheses(binary.left());
        if length.kind() != Kind::PropertyAccessExpression || ast::is_optional_chain(length) {
            return None;
        }
        let access = length.as_property_access_expression();
        if access.name().kind() != Kind::Identifier || access.name().text() != "length" {
            return None;
        }
        let call = method_call(
            ast::skip_parentheses(access.expression()),
            &["filter"],
            0,
            None,
            false,
        )?;
        let object = ast::skip_parentheses(call.object);
        if (object.kind() == Kind::Identifier && object.text().starts_with('$'))
            || self.receiver(
                call.object,
                true,
                self.has_type_information,
                &mut HashSet::new(),
                0,
            ) == Class::NonTarget
        {
            return None;
        }
        if call
            .call
            .arguments()
            .first()
            .is_none_or(|&argument| not_function(argument))
        {
            return None;
        }
        let report = Violation::new(
            call.property,
            "filter",
            "Prefer `.some(…)` over non-zero length check from `.filter(…)`.",
        );
        let call_end = end(outer_parentheses(call.call));
        if has_comment(self.file, call_end, end(node)) {
            return Some(report);
        }
        Some(report.with_fix(vec![
            Edit::replace(range(call.property, self.file), "some"),
            Edit::replace([call_end, end(length)], ""),
            Edit::replace([end(outer_parentheses(length)), end(node)], ""),
        ]))
    }
}
fn is_reference(node: P<Node>) -> bool {
    let Some(parent) = node.parent() else {
        return false;
    };
    if (parent.kind() == Kind::PropertyAccessExpression && parent.name() == Some(node))
        || (parent.kind() == Kind::QualifiedName && parent.as_qualified_name().right() == node)
        || matches!(
            parent.kind(),
            Kind::LabeledStatement
                | Kind::BreakStatement
                | Kind::ContinueStatement
                | Kind::ImportSpecifier
                | Kind::ImportClause
                | Kind::NamespaceImport
        )
    {
        return false;
    }
    if parent.kind() == Kind::ExportSpecifier {
        return parent
            .as_export_specifier()
            .property_name()
            .unwrap_or(parent.name().expect("export name"))
            == node;
    }
    parent.kind() == Kind::ShorthandPropertyAssignment || !ast::is_declaration_name(node)
}

#[derive(Clone, Copy, PartialEq, Eq)]
enum Class {
    Unknown,
    Target,
    NonTarget,
}
fn combine(classes: &[Class], intersection: bool) -> Class {
    if intersection && classes.contains(&Class::Target) {
        return Class::Target;
    }
    if !classes.is_empty() && classes.iter().all(|class| *class == Class::Target) {
        Class::Target
    } else if classes.iter().all(|class| *class == Class::NonTarget) {
        Class::NonTarget
    } else {
        Class::Unknown
    }
}
fn typed_array(name: &str) -> bool {
    matches!(
        name,
        "Int8Array"
            | "Uint8Array"
            | "Uint8ClampedArray"
            | "Int16Array"
            | "Uint16Array"
            | "Int32Array"
            | "Uint32Array"
            | "Float16Array"
            | "Float32Array"
            | "Float64Array"
            | "BigInt64Array"
            | "BigUint64Array"
    )
}
fn known_name(name: &str, indexed: bool) -> Class {
    if matches!(name, "Array" | "ReadonlyArray") || (indexed && typed_array(name)) {
        Class::Target
    } else if typed_array(name)
        || matches!(
            name,
            "Map" | "ReadonlyMap" | "WeakMap" | "Set" | "ReadonlySet" | "WeakSet"
        )
    {
        Class::NonTarget
    } else {
        Class::Unknown
    }
}
impl ArraySome<'_> {
    fn receiver(
        &mut self,
        node: P<Node>,
        indexed: bool,
        typed: bool,
        visiting: &mut HashSet<P<Symbol>>,
        depth: usize,
    ) -> Class {
        if depth > 1024 {
            return Class::Unknown;
        }
        let node = ast::skip_parentheses(node);
        if matches!(
            node.kind(),
            Kind::SatisfiesExpression | Kind::NonNullExpression
        ) {
            return self.receiver(
                node.expression().expect("wrapper expression"),
                indexed,
                typed,
                visiting,
                depth + 1,
            );
        }
        if matches!(
            node.kind(),
            Kind::AsExpression | Kind::TypeAssertionExpression
        ) {
            if typed && let Some(annotation) = node.type_node() {
                let class = self.annotation(annotation, indexed, &mut HashSet::new());
                if class != Class::Unknown {
                    return class;
                }
            }
            return self.receiver(
                node.expression().expect("assertion expression"),
                indexed,
                typed,
                visiting,
                depth + 1,
            );
        }
        if node.kind() == Kind::Identifier
            && let Some(symbol) = resolve_reference(node, self.checker)
            && symbol.declarations().len() == 1
            && !visiting.contains(&symbol)
        {
            let declaration = symbol.declarations()[0];
            if ast::get_source_file_of_node(declaration) == Some(self.file) {
                visiting.insert(symbol);
                let mut class = Class::Unknown;
                if typed
                    && matches!(
                        declaration.kind(),
                        Kind::VariableDeclaration | Kind::Parameter
                    )
                    && let Some(annotation) = declaration.type_node()
                {
                    class = self.annotation(annotation, indexed, &mut HashSet::new());
                }
                if class == Class::Unknown
                    && declaration.kind() == Kind::VariableDeclaration
                    && declaration
                        .name()
                        .is_some_and(|name| name.kind() == Kind::Identifier)
                    && declaration.parent().is_some_and(|list| {
                        list.kind() == Kind::VariableDeclarationList
                            && list.flags().intersects(NodeFlags::Const)
                            && (typed || !list.flags().intersects(NodeFlags::Using))
                    })
                    && let Some(initializer) = declaration.as_variable_declaration().initializer()
                {
                    class = self.receiver(initializer, indexed, typed, visiting, depth + 1);
                }
                visiting.remove(&symbol);
                if class != Class::Unknown {
                    return class;
                }
            }
        }
        if node.kind() == Kind::ConditionalExpression {
            let expression = node.as_conditional_expression();
            let left = self.receiver(expression.when_true(), indexed, typed, visiting, depth + 1);
            let right = self.receiver(expression.when_false(), indexed, typed, visiting, depth + 1);
            let class = combine(&[left, right], false);
            if class != Class::Unknown || typed {
                return class;
            }
        }
        if node.kind() == Kind::BinaryExpression
            && node.as_binary_expression().operator_token().kind() == Kind::CommaToken
        {
            return self.receiver(
                node.as_binary_expression().right(),
                indexed,
                typed,
                visiting,
                depth + 1,
            );
        }
        // Literal values precede type queries in the baseline static evaluator.
        if matches!(
            node.kind(),
            Kind::StringLiteral
                | Kind::NumericLiteral
                | Kind::BigIntLiteral
                | Kind::RegularExpressionLiteral
                | Kind::TrueKeyword
                | Kind::FalseKeyword
                | Kind::NullKeyword
                | Kind::NoSubstitutionTemplateLiteral
                | Kind::PrefixUnaryExpression
                | Kind::TypeOfExpression
                | Kind::VoidExpression
        ) {
            return Class::NonTarget;
        }
        if node.kind() == Kind::ArrayLiteralExpression {
            return Class::Target;
        }
        if matches!(node.kind(), Kind::CallExpression | Kind::NewExpression)
            && let Some(callee) = node.expression().map(ast::skip_parentheses)
        {
            if callee.kind() == Kind::Identifier
                && (callee.text() == "Array"
                    || (node.kind() == Kind::NewExpression
                        && indexed
                        && typed_array(callee.text())))
            {
                return Class::Target;
            }
            if node.kind() == Kind::CallExpression
                && callee.kind() == Kind::PropertyAccessExpression
                && !ast::is_optional_chain(callee)
            {
                let access = callee.as_property_access_expression();
                let object = ast::skip_parentheses(access.expression());
                if object.kind() == Kind::Identifier
                    && object.text() == "Array"
                    && matches!(access.name().text(), "from" | "of")
                {
                    return Class::Target;
                }
            }
        }
        if typed {
            let value = self.checker.get_type_at_location(node);
            let class = self.type_class(value, indexed, 0);
            if class != Class::Unknown {
                return class;
            }
        }
        if matches!(
            node.kind(),
            Kind::NewExpression
                | Kind::ObjectLiteralExpression
                | Kind::FunctionExpression
                | Kind::ArrowFunction
                | Kind::ClassExpression
                | Kind::TemplateExpression
                | Kind::NoSubstitutionTemplateLiteral
        ) {
            return Class::NonTarget;
        }
        Class::Unknown
    }
    fn type_class(&mut self, value: P<Type>, indexed: bool, depth: usize) -> Class {
        if depth > 256
            || value
                .flags()
                .intersects(TypeFlags::Any | TypeFlags::Unknown)
        {
            return Class::Unknown;
        }
        if value
            .flags()
            .intersects(TypeFlags::Null | TypeFlags::Undefined)
        {
            return Class::NonTarget;
        }
        if value.flags().intersects(TypeFlags::TypeParameter) {
            return self
                .checker
                .get_base_constraint_of_type(value)
                .filter(|constraint| *constraint != value)
                .map_or(Class::Unknown, |constraint| {
                    self.type_class(constraint, indexed, depth + 1)
                });
        }
        if value
            .flags()
            .intersects(TypeFlags::Union | TypeFlags::Intersection)
        {
            let classes: Vec<_> = value
                .types()
                .iter()
                .map(|&part| self.type_class(part, indexed, depth + 1))
                .collect();
            return combine(&classes, value.flags().intersects(TypeFlags::Intersection));
        }
        if self.checker.is_array_or_tuple_type(value) {
            return Class::Target;
        }
        if let Some(constraint) = self
            .checker
            .get_base_constraint_of_type(value)
            .filter(|constraint| *constraint != value)
        {
            return self.type_class(constraint, indexed, depth + 1);
        }
        let symbol = value
            .symbol()
            .or_else(|| value.alias().and_then(|alias| alias.symbol()));
        if let Some(symbol) = symbol {
            return if known_name(symbol.name(), indexed) == Class::Target {
                Class::Target
            } else {
                Class::NonTarget
            };
        }
        if value
            .flags()
            .intersects(TypeFlags::Primitive | TypeFlags::Intrinsic)
        {
            Class::NonTarget
        } else {
            Class::Unknown
        }
    }
    fn annotation(
        &mut self,
        node: P<Node>,
        indexed: bool,
        visiting: &mut HashSet<P<Symbol>>,
    ) -> Class {
        match node.kind() {
            Kind::ParenthesizedType => {
                self.annotation(node.as_parenthesized_type_node().type_(), indexed, visiting)
            }
            Kind::TypeOperator
                if node.as_type_operator_node().operator() == Kind::ReadonlyKeyword =>
            {
                self.annotation(node.as_type_operator_node().type_(), indexed, visiting)
            }
            Kind::ArrayType | Kind::TupleType => Class::Target,
            Kind::NullKeyword
            | Kind::UndefinedKeyword
            | Kind::LiteralType
            | Kind::BigIntKeyword
            | Kind::BooleanKeyword
            | Kind::NeverKeyword
            | Kind::NumberKeyword
            | Kind::StringKeyword
            | Kind::SymbolKeyword
            | Kind::VoidKeyword
            | Kind::TypeLiteral
            | Kind::FunctionType
            | Kind::ConstructorType => Class::NonTarget,
            Kind::UnionType | Kind::IntersectionType => {
                let parts = if node.kind() == Kind::UnionType {
                    node.as_union_type_node().types()
                } else {
                    node.as_intersection_type_node().types()
                };
                let classes: Vec<_> = parts
                    .nodes()
                    .iter()
                    .map(|&part| self.annotation(part, indexed, visiting))
                    .collect();
                combine(&classes, node.kind() == Kind::IntersectionType)
            }
            Kind::TypeReference => {
                self.type_name(node.as_type_reference_node().type_name(), indexed, visiting)
            }
            _ => Class::Unknown,
        }
    }
    fn type_name(
        &mut self,
        name: P<Node>,
        indexed: bool,
        visiting: &mut HashSet<P<Symbol>>,
    ) -> Class {
        if name.kind() != Kind::Identifier {
            return Class::Unknown;
        }
        let symbol = self.checker.resolve_name_exported(
            name.text(),
            name,
            SymbolFlags::Type | SymbolFlags::Namespace | SymbolFlags::Alias,
            false,
        );
        let Some(symbol) = symbol.filter(|symbol| {
            symbol
                .declarations()
                .iter()
                .any(|&declaration| ast::get_source_file_of_node(declaration) == Some(self.file))
        }) else {
            return known_name(name.text(), indexed);
        };
        if !visiting.insert(symbol) {
            return Class::Unknown;
        }
        let mut class = Class::Unknown;
        for &declaration in symbol.declarations() {
            class = match declaration.kind() {
                Kind::TypeAliasDeclaration => declaration
                    .as_type_alias_declaration()
                    .type_()
                    .map_or(Class::Unknown, |annotation| {
                        self.annotation(annotation, indexed, visiting)
                    }),
                Kind::TypeParameter => declaration
                    .as_type_parameter_declaration()
                    .constraint()
                    .map_or(Class::Unknown, |constraint| {
                        self.annotation(constraint, indexed, visiting)
                    }),
                Kind::InterfaceDeclaration => {
                    let mut classes = Vec::new();
                    if let Some(heritage) =
                        declaration.as_interface_declaration().heritage_clauses()
                    {
                        for &clause in heritage.nodes() {
                            if clause.as_heritage_clause().token() != Kind::ExtendsKeyword {
                                continue;
                            }
                            for &base in clause.as_heritage_clause().types().nodes() {
                                // Interface heritage is parsed as TypeReference;
                                // class heritage uses ExpressionWithTypeArguments.
                                // Never cast between their different AST layouts.
                                let base_name = match base.kind() {
                                    Kind::TypeReference => {
                                        Some(base.as_type_reference_node().type_name())
                                    }
                                    Kind::ExpressionWithTypeArguments => {
                                        Some(base.as_expression_with_type_arguments().expression())
                                    }
                                    _ => None,
                                };
                                classes.push(base_name.map_or(Class::Unknown, |name| {
                                    self.type_name(name, indexed, visiting)
                                }));
                            }
                        }
                    }
                    combine(&classes, true)
                }
                Kind::ClassDeclaration | Kind::ClassExpression => Class::NonTarget,
                Kind::ImportClause | Kind::ImportSpecifier | Kind::NamespaceImport => {
                    let imported = if declaration.kind() == Kind::ImportSpecifier {
                        declaration
                            .as_import_specifier()
                            .property_name()
                            .map_or(name.text(), |property| property.text())
                    } else {
                        name.text()
                    };
                    if known_name(name.text(), indexed) != Class::Unknown
                        || known_name(imported, indexed) != Class::Unknown
                    {
                        Class::NonTarget
                    } else {
                        Class::Unknown
                    }
                }
                _ => continue,
            };
            break;
        }
        visiting.remove(&symbol);
        class
    }
}
