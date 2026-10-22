//! Regex rules used by the fixed Rsbuild configuration, ported from rslint
//! e0a29f751d4def86d27931759098f27748fa3003 (MIT). Pattern decisions use the
//! ECMAScript AST in rsbuild_regexp, retaining rule-specific constructor scope.
use super::{rsbuild_assignment_checks::resolve_reference, rsbuild_regexp as rx};
use crate::{Violation, directives, edits::Edit};
use oxc_regular_expression::ast::CharacterKind;
use tsrs_ast::{self as ast, Kind, Node, SourceFile};
use tsrs_checker::{Checker, TypeFlags};
use tsrs_core::P;

pub(crate) const NAMES: &[&str] = &[
    "no-control-regex",
    "no-invalid-regexp",
    "no-misleading-character-class",
    "no-regex-spaces",
    "no-useless-backreference",
    "no-useless-escape",
];
fn start(node: P<Node>, file: P<SourceFile>) -> usize {
    tsrs_scanner::skip_trivia(file.text(), node.pos().expect_source("parsed node")) as usize
}
fn end(node: P<Node>) -> usize {
    node.end().expect_source("parsed node") as usize
}
fn raw<'a>(node: P<Node>, file: P<SourceFile>) -> &'a str {
    &file.text()[start(node, file)..end(node)]
}
fn skip(mut node: P<Node>) -> P<Node> {
    while node.kind() == Kind::ParenthesizedExpression {
        node = node.as_parenthesized_expression().expression();
    }
    node
}
fn assertions(mut node: P<Node>) -> P<Node> {
    loop {
        node = skip(node);
        node = match node.kind() {
            Kind::AsExpression => node.as_as_expression().expression(),
            Kind::TypeAssertionExpression => node.as_type_assertion().expression(),
            Kind::NonNullExpression => node.as_non_null_expression().expression(),
            Kind::SatisfiesExpression => node.as_satisfies_expression().expression(),
            _ => return node,
        };
    }
}
fn call(node: P<Node>) -> Option<(P<Node>, Vec<P<Node>>)> {
    match node.kind() {
        Kind::CallExpression => Some((
            skip(node.as_call_expression().expression()),
            node.arguments().to_vec(),
        )),
        Kind::NewExpression => Some((
            skip(node.as_new_expression().expression()),
            node.arguments().to_vec(),
        )),
        _ => None,
    }
}
fn literal<'a>(node: P<Node>, file: P<SourceFile>) -> Option<(&'a str, &'a str)> {
    raw(node, file).strip_prefix('/')?.rsplit_once('/')
}
fn walk(node: P<Node>, f: &mut impl FnMut(P<Node>)) {
    f(node);
    node.for_each_child(&mut |child| {
        walk(child, f);
        false
    });
}
// Resolve the authored runtime binding. Type-only declarations do not shadow
// the global value, and declarations inside `declare global` augment that
// value rather than introducing a lexical binding in this file.
fn global(node: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> bool {
    if node.kind() != Kind::Identifier {
        return false;
    }
    !resolve_reference(node, checker).is_some_and(|symbol| {
        symbol.declarations().iter().any(|&declaration| {
            if ast::get_source_file_of_node(declaration) != Some(file)
                || matches!(
                    declaration.kind(),
                    Kind::TypeAliasDeclaration | Kind::InterfaceDeclaration
                )
            {
                return false;
            }
            let mut ancestor = declaration.parent();
            while let Some(parent) = ancestor {
                if ast::is_global_scope_augmentation(parent) {
                    return false;
                }
                ancestor = parent.parent();
            }
            if declaration.kind() == Kind::ModuleDeclaration {
                // The TypeScript checker can merge a block namespace with
                // lib.d.ts's RegExp symbol, even for an outside reference.
                // The frozen reference tracker gives that runtime namespace
                // the scope of its containing block; the symbol's aggregate
                // declaration list cannot establish lexical visibility.
                let mut scope = declaration.parent();
                while let Some(container) = scope {
                    if matches!(
                        container.kind(),
                        Kind::SourceFile | Kind::Block | Kind::ModuleBlock | Kind::CaseBlock
                    ) {
                        let mut reference = Some(node);
                        while let Some(reference_node) = reference {
                            if reference_node == container {
                                return true;
                            }
                            reference = reference_node.parent();
                        }
                        return false;
                    }
                    scope = container.parent();
                }
            }
            true
        })
    })
}
fn global_available(node: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> bool {
    global(node, file, checker)
        && directives::global_setting(file, node.text())
            .unwrap_or(matches!(node.text(), "RegExp" | "globalThis" | "String"))
}
fn root_unwritten(node: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> bool {
    if !global_available(node, file, checker) {
        return false;
    }
    let mut modified = false;
    walk(file.as_node(), &mut |candidate| {
        let target = match candidate.kind() {
            Kind::BinaryExpression => {
                let b = candidate.as_binary_expression();
                if ast::is_assignment_operator(b.operator_token().kind()) {
                    Some(b.left())
                } else {
                    None
                }
            }
            Kind::PrefixUnaryExpression => {
                let p = candidate.as_prefix_unary_expression();
                matches!(p.operator(), Kind::PlusPlusToken | Kind::MinusMinusToken)
                    .then_some(p.operand())
            }
            Kind::PostfixUnaryExpression => Some(candidate.as_postfix_unary_expression().operand()),
            _ => None,
        };
        if let Some(target) = target {
            let target = assertions(target);
            if target.kind() == Kind::Identifier
                && target.text() == node.text()
                && global(target, file, checker)
            {
                modified = true;
            }
        }
    });
    !modified
}
fn access(node: P<Node>) -> Option<(P<Node>, String)> {
    match node.kind() {
        Kind::PropertyAccessExpression => Some((
            node.as_property_access_expression().expression(),
            node.as_property_access_expression().name().text().into(),
        )),
        Kind::ElementAccessExpression => {
            let a = node.as_element_access_expression();
            let key = skip(a.argument_expression());
            matches!(
                key.kind(),
                Kind::StringLiteral | Kind::NoSubstitutionTemplateLiteral
            )
            .then(|| (a.expression(), key.text().into()))
        }
        _ => None,
    }
}
// A resolved symbol can denote a function, class or import without an
// initializer field. The frozen Go tracker follows known initializer bindings;
// this reverse lookup must guard kinds before calling tsrs Node::initializer.
fn declaration_initializer(node: P<Node>) -> Option<P<Node>> {
    match node.kind() {
        Kind::VariableDeclaration
        | Kind::Parameter
        | Kind::BindingElement
        | Kind::PropertyDeclaration
        | Kind::PropertySignature
        | Kind::PropertyAssignment
        | Kind::EnumMember => node.initializer(),
        _ => None,
    }
}
fn static_property_name(
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    depth: usize,
) -> Option<String> {
    match node.kind() {
        Kind::Identifier => Some(node.text().into()),
        Kind::ComputedPropertyName => static_string(
            node.as_computed_property_name().expression(),
            file,
            checker,
            depth + 1,
        ),
        _ => static_string(node, file, checker, depth + 1),
    }
}
fn same_binding(left: P<Node>, right: P<Node>, checker: &mut Checker) -> bool {
    left.kind() == Kind::Identifier
        && right.kind() == Kind::Identifier
        && resolve_reference(left, checker).is_some_and(|left| {
            resolve_reference(right, checker).is_some_and(|right| left == right)
        })
}
fn has_independent_value(node: P<Node>) -> bool {
    let node = assertions(node);
    match node.kind() {
        Kind::ConditionalExpression => {
            let condition = node.as_conditional_expression();
            has_independent_value(condition.when_true())
                || has_independent_value(condition.when_false())
        }
        Kind::BinaryExpression => {
            let expression = node.as_binary_expression();
            has_independent_value(expression.left()) || has_independent_value(expression.right())
        }
        Kind::ArrowFunction
        | Kind::FunctionExpression
        | Kind::ClassExpression
        | Kind::ObjectLiteralExpression
        | Kind::NumericLiteral
        | Kind::NullKeyword => true,
        _ => false,
    }
}
fn constructor_type(
    node: P<Node>,
    checker: &mut Checker,
    default_library: &dyn Fn(P<SourceFile>) -> bool,
) -> Option<bool> {
    let value = checker.get_type_at_location(node);
    let value = checker.get_base_constraint_of_type(value).unwrap_or(value);
    if value
        .flags()
        .intersects(TypeFlags::Any | TypeFlags::Unknown)
    {
        return None;
    }
    Some(value.symbol().is_some_and(|symbol| {
        symbol.name() == "RegExpConstructor"
            && symbol.declarations().iter().any(|&declaration| {
                ast::get_source_file_of_node(declaration).is_some_and(default_library)
            })
    }))
}
fn builtin(
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    default_library: &dyn Fn(P<SourceFile>) -> bool,
    typed: bool,
    depth: usize,
) -> bool {
    if depth > 32 {
        return false;
    }
    let node = assertions(node);
    if node.kind() == Kind::Identifier {
        if node.text() == "RegExp" && global(node, file, checker) {
            return root_unwritten(node, file, checker);
        }
        if let Some(symbol) = resolve_reference(node, checker) {
            for &declaration in symbol.declarations() {
                if let Some(initializer) = declaration_initializer(declaration) {
                    // A root-reachable alias with independent value sources
                    // follows the checker's flow type at this particular read.
                    if typed && has_independent_value(initializer) {
                        if let Some(result) = constructor_type(node, checker, default_library) {
                            return result;
                        }
                    }
                    if builtin(
                        initializer,
                        file,
                        checker,
                        default_library,
                        typed,
                        depth + 1,
                    ) {
                        return true;
                    }
                }
                if declaration.kind() == Kind::BindingElement {
                    let binding = declaration.as_binding_element();
                    let name = binding.property_name().or_else(|| declaration.name());
                    if name
                        .and_then(|name| static_property_name(name, file, checker, depth + 1))
                        .as_deref()
                        == Some("RegExp")
                    {
                        if let Some(init) = declaration
                            .parent()
                            .and_then(|p| p.parent())
                            .and_then(declaration_initializer)
                        {
                            if global_object(init, file, checker, depth + 1) {
                                return true;
                            }
                        }
                    }
                }
            }
        }
        // ReferenceTracker also propagates writes to an existing alias and
        // destructuring assignments, independently of declaration initializers.
        let mut traced = false;
        if ["RegExp", "globalThis", "window", "self", "global"]
            .iter()
            .any(|name| file.has_identifier(name))
        {
            walk(file.as_node(), &mut |candidate| {
                if traced || candidate.kind() != Kind::BinaryExpression {
                    return;
                }
                let expression = candidate.as_binary_expression();
                if expression.operator_token().kind() != Kind::EqualsToken {
                    return;
                }
                let target = assertions(expression.left());
                if same_binding(node, target, checker) {
                    traced = builtin(
                        expression.right(),
                        file,
                        checker,
                        default_library,
                        typed,
                        depth + 1,
                    );
                } else if target.kind() == Kind::ObjectLiteralExpression
                    && global_object(expression.right(), file, checker, depth + 1)
                {
                    for &property in target.as_object_literal_expression().properties().nodes() {
                        if property.kind() == Kind::PropertyAssignment {
                            let property = property.as_property_assignment();
                            if static_property_name(property.name(), file, checker, depth + 1)
                                .as_deref()
                                == Some("RegExp")
                                && same_binding(node, assertions(property.initializer()), checker)
                            {
                                traced = true;
                            }
                        }
                    }
                }
            });
        }
        if traced {
            return true;
        }
        // A known alias initializer is authoritative, including a disabled
        // effective global root such as window. Library types must not revive it.
        let initialized = resolve_reference(node, checker).is_some_and(|symbol| {
            symbol
                .declarations()
                .iter()
                .any(|&declaration| declaration_initializer(declaration).is_some())
        });
        if typed && node.text() != "RegExp" && !initialized {
            return constructor_type(node, checker, default_library).unwrap_or(false);
        }
        return false;
    }
    if let Some((object, name)) = access(node) {
        return name == "RegExp" && global_object(object, file, checker, depth + 1);
    }
    match node.kind() {
        Kind::BinaryExpression => {
            let b = node.as_binary_expression();
            match b.operator_token().kind() {
                Kind::CommaToken | Kind::EqualsToken => {
                    builtin(b.right(), file, checker, default_library, typed, depth + 1)
                }
                Kind::BarBarToken | Kind::AmpersandAmpersandToken | Kind::QuestionQuestionToken => {
                    builtin(b.left(), file, checker, default_library, typed, depth + 1)
                        || builtin(b.right(), file, checker, default_library, typed, depth + 1)
                }
                _ => false,
            }
        }
        Kind::ConditionalExpression => {
            let c = node.as_conditional_expression();
            builtin(
                c.when_true(),
                file,
                checker,
                default_library,
                typed,
                depth + 1,
            ) || builtin(
                c.when_false(),
                file,
                checker,
                default_library,
                typed,
                depth + 1,
            )
        }
        _ => false,
    }
}
fn global_object(node: P<Node>, file: P<SourceFile>, checker: &mut Checker, depth: usize) -> bool {
    if depth > 32 {
        return false;
    }
    let node = assertions(node);
    if node.kind() != Kind::Identifier {
        return false;
    }
    if matches!(node.text(), "globalThis" | "window" | "self" | "global")
        && global_available(node, file, checker)
    {
        return root_unwritten(node, file, checker);
    }
    resolve_reference(node, checker).is_some_and(|symbol| {
        symbol.declarations().iter().any(|declaration| {
            declaration_initializer(*declaration)
                .is_some_and(|init| global_object(init, file, checker, depth + 1))
        })
    })
}
fn stable_initializer(
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
) -> Option<P<Node>> {
    let symbol = resolve_reference(node, checker)?;
    let declaration = symbol
        .value_declaration()
        .or_else(|| symbol.declarations().first().copied())?;
    let initializer = declaration_initializer(declaration)?;
    let mut written = false;
    walk(file.as_node(), &mut |candidate| {
        let target = match candidate.kind() {
            Kind::BinaryExpression
                if ast::is_assignment_operator(
                    candidate.as_binary_expression().operator_token().kind(),
                ) =>
            {
                Some(candidate.as_binary_expression().left())
            }
            Kind::PrefixUnaryExpression
                if matches!(
                    candidate.as_prefix_unary_expression().operator(),
                    Kind::PlusPlusToken | Kind::MinusMinusToken
                ) =>
            {
                Some(candidate.as_prefix_unary_expression().operand())
            }
            Kind::PostfixUnaryExpression => Some(candidate.as_postfix_unary_expression().operand()),
            _ => None,
        };
        if target.is_some_and(|target| same_binding(node, assertions(target), checker)) {
            written = true;
        }
    });
    (!written).then_some(initializer)
}
fn string_constructor(
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    depth: usize,
) -> bool {
    if depth > 32 {
        return false;
    }
    let node = assertions(node);
    if node.kind() != Kind::Identifier {
        return false;
    }
    if node.text() == "String" && global_available(node, file, checker) {
        return true;
    }
    stable_initializer(node, file, checker)
        .is_some_and(|initializer| string_constructor(initializer, file, checker, depth + 1))
}
fn static_string(
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    depth: usize,
) -> Option<String> {
    if depth > 32 {
        return None;
    }
    let node = assertions(node);
    match node.kind() {
        Kind::StringLiteral | Kind::NoSubstitutionTemplateLiteral => Some(node.text().into()),
        Kind::Identifier => {
            let initializer = stable_initializer(node, file, checker)?;
            if initializer == node {
                return None;
            }
            static_string(initializer, file, checker, depth + 1)
        }
        Kind::ConditionalExpression => {
            let expression = node.as_conditional_expression();
            match assertions(expression.condition()).kind() {
                Kind::TrueKeyword => {
                    static_string(expression.when_true(), file, checker, depth + 1)
                }
                Kind::FalseKeyword => {
                    static_string(expression.when_false(), file, checker, depth + 1)
                }
                _ => None,
            }
        }
        Kind::CallExpression => {
            let expression = node.as_call_expression();
            if !string_constructor(expression.expression(), file, checker, depth + 1) {
                return None;
            }
            match expression.arguments().nodes().first() {
                Some(&argument) => {
                    static_string(argument, file, checker, depth + 1).or_else(|| {
                        super::rsbuild_declaration_checks::static_expression(
                            assertions(argument),
                            file,
                        )
                    })
                }
                None => Some(String::new()),
            }
        }
        Kind::BinaryExpression => {
            let b = node.as_binary_expression();
            if b.operator_token().kind() != Kind::PlusToken {
                return None;
            }
            Some(
                static_string(b.left(), file, checker, depth + 1)?
                    + &static_string(b.right(), file, checker, depth + 1)?,
            )
        }
        Kind::TemplateExpression => {
            let template = node.as_template_expression();
            let mut result = template.head().text().to_owned();
            for &span in template.template_spans().nodes() {
                let span = span.as_template_span();
                let expr = span.expression();
                result += &static_string(expr, file, checker, depth + 1)
                    .or_else(|| super::rsbuild_declaration_checks::static_expression(expr, file))?;
                result += span.literal().text();
            }
            Some(result)
        }
        Kind::TaggedTemplateExpression => {
            let tagged = node.as_tagged_template_expression();
            let (object, name) = access(tagged.tag())?;
            if name != "raw" || !string_constructor(object, file, checker, depth + 1) {
                return None;
            }
            let template = tagged.template();
            if template.kind() == Kind::NoSubstitutionTemplateLiteral {
                let text = raw(template, file);
                return Some(text[1..text.len() - 1].into());
            }
            if template.kind() != Kind::TemplateExpression {
                return None;
            }
            let template = template.as_template_expression();
            let head = raw(template.head(), file);
            let mut result = head[1..head.len() - 2].to_owned();
            for &span in template.template_spans().nodes() {
                let span = span.as_template_span();
                let expression = span.expression();
                result += &static_string(expression, file, checker, depth + 1).or_else(|| {
                    super::rsbuild_declaration_checks::static_expression(
                        assertions(expression),
                        file,
                    )
                })?;
                let tail = raw(span.literal(), file);
                let suffix = if span.literal().kind() == Kind::TemplateTail {
                    1
                } else {
                    2
                };
                result += &tail[1..tail.len() - suffix];
            }
            Some(result)
        }
        _ => None,
    }
}
fn is_handled_literal(
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    default_library: &dyn Fn(P<SourceFile>) -> bool,
    typed: bool,
    any_flags: bool,
) -> bool {
    let mut outer = node;
    while let Some(parent) = outer
        .parent()
        .filter(|p| p.kind() == Kind::ParenthesizedExpression)
    {
        outer = parent;
    }
    let Some(parent) = outer.parent() else {
        return false;
    };
    let Some((callee, args)) = call(parent) else {
        return false;
    };
    !args.is_empty()
        && skip(args[0]) == node
        && (any_flags || args.len() > 1)
        && builtin(callee, file, checker, default_library, typed, 0)
}

pub(crate) fn check(
    name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    default_library: &dyn Fn(P<SourceFile>) -> bool,
    has_type_information: bool,
) -> Vec<Violation> {
    if name == "no-useless-escape" {
        return useless_escape(node, file);
    }
    if name == "no-invalid-regexp" {
        return invalid_regexp(node, file);
    }
    let literal_node = node.kind() == Kind::RegularExpressionLiteral;
    if literal_node {
        if matches!(
            name,
            "no-useless-backreference" | "no-misleading-character-class"
        ) && is_handled_literal(
            node,
            file,
            checker,
            default_library,
            has_type_information,
            name == "no-useless-backreference",
        ) {
            return Vec::new();
        }
        let Some((pattern, flags)) = literal(node, file) else {
            return Vec::new();
        };
        return analyze(
            name,
            node,
            node,
            file,
            pattern,
            flags,
            Some((pattern, start(node, file) + 1)),
            None,
        );
    }
    let Some((callee, args)) = call(node) else {
        return Vec::new();
    };
    if args.is_empty() {
        return Vec::new();
    }
    if matches!(name, "no-control-regex" | "no-regex-spaces") {
        if callee.kind() != Kind::Identifier
            || callee.text() != "RegExp"
            || !global(callee, file, checker)
        {
            return Vec::new();
        }
        if name == "no-control-regex" && !global_available(callee, file, checker) {
            return Vec::new();
        }
    } else if !builtin(
        callee,
        file,
        checker,
        default_library,
        has_type_information && name == "no-useless-backreference",
        0,
    ) {
        return Vec::new();
    }
    let mut argument = if name == "no-regex-spaces" {
        args[0]
    } else {
        skip(args[0])
    };
    if matches!(name, "no-control-regex" | "no-regex-spaces")
        && argument.kind() != Kind::StringLiteral
    {
        return Vec::new();
    }
    if name == "no-misleading-character-class"
        && has_type_information
        && argument.kind() == Kind::Identifier
    {
        if let Some(initializer) = stable_initializer(argument, file, checker) {
            let initializer = assertions(initializer);
            if matches!(
                initializer.kind(),
                Kind::StringLiteral | Kind::NoSubstitutionTemplateLiteral
            ) {
                argument = initializer;
            }
        }
    }
    let flags = if let Some(&arg) = args.get(1) {
        let flag_node = if name == "no-regex-spaces" {
            arg
        } else {
            skip(arg)
        };
        if matches!(name, "no-control-regex" | "no-regex-spaces") {
            if flag_node.kind() == Kind::StringLiteral {
                Some(flag_node.text().to_owned())
            } else if name == "no-control-regex" {
                Some(String::new())
            } else {
                None
            }
        } else {
            static_string(flag_node, file, checker, 0)
                .or_else(|| (name == "no-useless-backreference").then(String::new))
        }
    } else {
        Some(String::new())
    };
    let Some(mut flags) = flags else {
        return Vec::new();
    };
    let (pattern, raw_pattern) = if argument.kind() == Kind::RegularExpressionLiteral {
        let Some((pattern, literal_flags)) = literal(argument, file) else {
            return Vec::new();
        };
        if args.len() < 2 {
            if name == "no-misleading-character-class" {
                return Vec::new();
            }
            flags = literal_flags.into();
        }
        (pattern.into(), Some((pattern, start(argument, file) + 1)))
    } else {
        let Some(value) = static_string(argument, file, checker, 0) else {
            return Vec::new();
        };
        let raw_pattern = matches!(
            argument.kind(),
            Kind::StringLiteral | Kind::NoSubstitutionTemplateLiteral
        )
        .then(|| {
            let text = raw(argument, file);
            (&text[1..text.len() - 1], start(argument, file) + 1)
        });
        (value, raw_pattern)
    };
    let report_node = if name == "no-control-regex" {
        argument
    } else {
        node
    };
    analyze(
        name,
        report_node,
        argument,
        file,
        &pattern,
        &flags,
        raw_pattern,
        Some(&args),
    )
}
fn analyze(
    name: &str,
    node: P<Node>,
    argument: P<Node>,
    file: P<SourceFile>,
    pattern: &str,
    flags: &str,
    raw_pattern: Option<(&str, usize)>,
    args: Option<&[P<Node>]>,
) -> Vec<Violation> {
    if name == "no-control-regex" {
        return control_regex(node, pattern, flags);
    }
    if name == "no-regex-spaces" {
        return regex_spaces(node, pattern, flags, raw_pattern);
    }
    if name == "no-useless-backreference" {
        let Ok(parsed) = rx::parse(pattern, flags) else {
            return Vec::new();
        };
        return parsed.reference_problems().into_iter().map(|(r,g,id,count)|{
            let reference=&parsed.nodes[r];let group=&parsed.nodes[g];let reference=&pattern[reference.start..reference.end];let group=&pattern[group.start..group.end];
            let other=match count{1=>String::new(),2=>" and another group".into(),_=>format!(" and other {} groups",count-1)};
            let ending=match id{"nested"=>"from within that group","forward"=>"which appears later in the pattern","backward"=>"which appears before in the same lookbehind","disjunctive"=>"which is in another alternative",_=>"which is in a negative lookaround"};
            Violation::new(node,id,format!("Backreference '{reference}' will be ignored. It references group '{group}'{other} {ending}."))
        }).collect();
    }
    misleading(node, argument, file, pattern, flags, raw_pattern, args)
}
fn control_regex(node: P<Node>, pattern: &str, flags: &str) -> Vec<Violation> {
    let parsed = match rx::parse(pattern, flags) {
        Ok(parsed) => parsed,
        Err(error) => {
            // regexpp emits character events before a later grammar error.
            // The prefix can still contain a complete control escape even
            // though no full RegExp AST can be returned by Oxc.
            let prefix = &pattern[..error.offset.min(pattern.len())];
            let chars = partial_controls(prefix, flags.contains('u') || flags.contains('v'));
            return if chars.is_empty() {
                Vec::new()
            } else {
                vec![Violation::new(
                    node,
                    "unexpected",
                    format!(
                        "Unexpected control character(s) in regular expression: {}.",
                        chars.join(", ")
                    ),
                )]
            };
        }
    };
    let chars: Vec<_> = parsed
        .chars
        .iter()
        .filter(|c| {
            c.value <= 0x1f
                && matches!(
                    c.kind,
                    CharacterKind::Symbol
                        | CharacterKind::Identifier
                        | CharacterKind::HexadecimalEscape
                        | CharacterKind::UnicodeEscape
                )
        })
        .map(|c| format!("\\x{:02x}", c.value))
        .collect();
    if chars.is_empty() {
        Vec::new()
    } else {
        vec![Violation::new(
            node,
            "unexpected",
            format!(
                "Unexpected control character(s) in regular expression: {}.",
                chars.join(", ")
            ),
        )]
    }
}
fn partial_controls(pattern: &str, unicode: bool) -> Vec<String> {
    let bytes = pattern.as_bytes();
    let mut out = Vec::new();
    let mut index = 0;
    while index < bytes.len() {
        if bytes[index] == b'\\' && index + 1 < bytes.len() {
            let begin = index + 2;
            let count = match bytes[index + 1] {
                b'x' => 2,
                b'u' => 4,
                _ => 0,
            };
            let escaped = if unicode && bytes[index + 1] == b'u' && bytes.get(begin) == Some(&b'{')
            {
                pattern[begin + 1..]
                    .find('}')
                    .map(|end| (begin + 1, begin + 1 + end, begin + 2 + end))
            } else if count > 0 && begin + count <= bytes.len() {
                Some((begin, begin + count, begin + count))
            } else {
                None
            };
            if let Some((begin, end, next)) = escaped {
                if begin < end && bytes[begin..end].iter().all(u8::is_ascii_hexdigit) {
                    if let Ok(value) = u32::from_str_radix(&pattern[begin..end], 16) {
                        if value <= 0x1f {
                            out.push(format!("\\x{value:02x}"));
                        }
                    }
                    index = next;
                    continue;
                }
            }
            index += 2;
        } else {
            if bytes[index] <= 0x1f {
                out.push(format!("\\x{:02x}", bytes[index]));
            }
            index += 1;
        }
    }
    out
}
// The frozen scanner's Unicode property table is the grammar authority for
// this rule; Oxc accepts newer property aliases which that baseline rejects.
fn compiler_regex_valid(pattern: &str, flags: &str) -> bool {
    if !flags.contains('u') && !flags.contains('v') {
        return true;
    }
    let bytes = pattern.as_bytes();
    let mut index = 0;
    while index + 2 < bytes.len() {
        if bytes[index] != b'\\' {
            index += 1;
            continue;
        }
        if matches!(bytes[index + 1], b'p' | b'P') && bytes[index + 2] == b'{' {
            let Some(close) = pattern[index + 3..].find('}') else {
                return false;
            };
            let end = index + 4 + close;
            // Validate the property itself, not the surrounding pattern: the
            // baseline adapter separately corrects scanner gaps in otherwise
            // valid set intersections, named groups and escaped ampersands.
            let text = format!(
                "/{}/{}",
                &pattern[index..end],
                if flags.contains('v') { "v" } else { "u" }
            );
            let mut scanner = tsrs_scanner::Scanner::new();
            scanner.set_on_error(true);
            scanner.set_text(tsrs_core::alloc_str(&text));
            scanner.scan();
            scanner.re_scan_slash_token(true);
            if scanner.has_errors() {
                return false;
            }
            index = end;
        } else {
            index += 2;
        }
    }
    true
}
fn regex_spaces(
    node: P<Node>,
    pattern: &str,
    flags: &str,
    raw_pattern: Option<(&str, usize)>,
) -> Vec<Violation> {
    let Some((raw, offset)) = raw_pattern else {
        return Vec::new();
    };
    if !raw.contains("  ") || !compiler_regex_valid(pattern, flags) {
        return Vec::new();
    }
    let Ok(parsed) = rx::parse(pattern, flags) else {
        return Vec::new();
    };
    let bytes = pattern.as_bytes();
    let mut i = 0;
    while i < bytes.len() {
        if bytes[i] != b' ' {
            i += 1;
            continue;
        }
        let start = i;
        while i < bytes.len() && bytes[i] == b' ' {
            i += 1;
        }
        let mut count = i - start;
        if bytes.get(i).is_some_and(|b| b"+*{?".contains(b)) {
            count = count.saturating_sub(1);
        }
        if count < 2
            || parsed
                .classes
                .iter()
                .any(|c| c.start <= start && start < c.end)
        {
            continue;
        }
        let mut report = Violation::new(
            node,
            "multipleSpaces",
            format!("Spaces are hard to count. Use {{{count}}}."),
        );
        if raw == pattern {
            report = report.with_fix(vec![Edit::replace(
                [offset + start, offset + start + count],
                format!(" {{{count}}}"),
            )]);
        }
        return vec![report];
    }
    Vec::new()
}
fn escape_report(
    node: P<Node>,
    offset: usize,
    character: &str,
    directive: bool,
    disable_backslash: bool,
) -> Violation {
    let mut report = Violation::new(
        node,
        "unnecessaryEscape",
        format!("Unnecessary escape character: \\{character}."),
    )
    .with_byte_range([offset, offset + 1])
    .with_suggestion(
        if directive {
            "removeEscapeDoNotKeepSemantics"
        } else {
            "removeEscape"
        },
        if directive {
            "Remove the `\\` if it was inserted by mistake."
        } else {
            "Remove the `\\`. This maintains the current functionality."
        },
        vec![Edit::remove([offset, offset + 1])],
    );
    if !disable_backslash {
        report = report.with_suggestion(
            "escapeBackslash",
            "Replace the `\\` with `\\\\` to include the actual backslash character.",
            vec![Edit::replace([offset, offset], "\\")],
        );
    }
    report
}
fn directive(node: P<Node>, file: P<SourceFile>) -> bool {
    let Some(statement) = node
        .parent()
        .filter(|p| p.kind() == Kind::ExpressionStatement)
    else {
        return false;
    };
    let Some(container) = statement.parent() else {
        return false;
    };
    let statements = match container.kind() {
        Kind::SourceFile => file.statements().nodes().to_vec(),
        Kind::ModuleBlock => container.statements().to_vec(),
        Kind::Block => {
            if !container.parent().is_some_and(|p| {
                ast::is_function_like(p) || p.kind() == Kind::ClassStaticBlockDeclaration
            }) {
                return false;
            }
            container.statements().to_vec()
        }
        _ => return false,
    };
    for candidate in statements {
        if candidate == statement {
            return true;
        }
        if !ast::is_prologue_directive(candidate) {
            return false;
        }
    }
    false
}
fn useless_escape(node: P<Node>, file: P<SourceFile>) -> Vec<Violation> {
    let kind = node.kind();
    let is_template = matches!(
        kind,
        Kind::NoSubstitutionTemplateLiteral
            | Kind::TemplateHead
            | Kind::TemplateMiddle
            | Kind::TemplateTail
    );
    if kind == Kind::StringLiteral || is_template {
        if node.parent().is_some_and(|p| {
            matches!(
                p.kind(),
                Kind::JsxAttribute | Kind::JsxElement | Kind::JsxFragment
            )
        }) {
            return Vec::new();
        }
        if is_template {
            let mut parent = node.parent();
            for _ in 0..3 {
                let Some(p) = parent else {
                    break;
                };
                if p.kind() == Kind::TaggedTemplateExpression {
                    return Vec::new();
                }
                parent = p.parent();
            }
        }
        let text = raw(node, file);
        let bytes = text.as_bytes();
        let mut i = 0;
        let mut out = Vec::new();
        while i < bytes.len() {
            if bytes[i] != b'\\' {
                i += 1;
                continue;
            }
            let Some(character) = text[i + 1..].chars().next() else {
                break;
            };
            let size = character.len_utf8();
            let allowed = character.is_ascii_digit()
                || matches!(
                    character,
                    '\\' | 'n'
                        | 'r'
                        | 'v'
                        | 't'
                        | 'b'
                        | 'f'
                        | 'u'
                        | 'x'
                        | '\n'
                        | '\r'
                        | '\u{2028}'
                        | '\u{2029}'
                )
                || (!is_template && bytes.first() == Some(&(character as u8)));
            let necessary = if is_template {
                match character {
                    '`' => true,
                    '$' => bytes.get(i + 1 + size) == Some(&b'{'),
                    '{' => i > 0 && bytes[i - 1] == b'$',
                    _ => allowed,
                }
            } else {
                allowed
            };
            if !necessary {
                out.push(escape_report(
                    node,
                    start(node, file) + i,
                    &text[i + 1..i + 1 + size],
                    kind == Kind::StringLiteral && directive(node, file),
                    false,
                ));
            }
            i += 1 + size;
        }
        return out;
    }
    if kind != Kind::RegularExpressionLiteral {
        return Vec::new();
    }
    let Some((pattern, flags)) = literal(node, file) else {
        return Vec::new();
    };
    let Ok(parsed) = rx::parse(pattern, flags) else {
        return Vec::new();
    };
    let mut out = Vec::new();
    for c in &parsed.chars {
        let Some(raw) = pattern.get(c.start..c.end) else {
            continue;
        };
        let Some(escaped) = raw.strip_prefix('\\') else {
            continue;
        };
        let Some(character) = escaped.chars().next() else {
            continue;
        };
        if escaped.len() != character.len_utf8() {
            continue;
        }
        let mut allowed = "\\bcdDfn pPrsStvwWx u0123456789]"
            .replace(' ', "")
            .contains(character);
        let class = c.class.map(|id| &parsed.classes[id]);
        if class.is_none() {
            allowed |= "^/.$*+?[{}|()Bk".contains(character);
        } else if flags.contains('v') {
            allowed |= "q/[{}|()-".contains(character);
        }
        let mut disable = false;
        if let Some(class) = class {
            if character == '^' && c.start == class.start + 1 {
                allowed = true;
            }
            if !flags.contains('v')
                && character == '-'
                && c.start != class.start + 1
                && c.end != class.end - 1
            {
                allowed = true;
            }
            if flags.contains('v') {
                if "!#$%&*+,.:;<=>?@^`~".contains(character) {
                    if pattern[c.end..].starts_with(character) {
                        allowed = true;
                    }
                    if pattern[..c.start].ends_with(character)
                        && (character != '^' || !class.negative || class.start + 1 < c.start - 1)
                    {
                        allowed = true;
                    }
                }
                disable = class.set_operation;
            }
        }
        if !allowed {
            out.push(escape_report(
                node,
                start(node, file) + 1 + c.start,
                escaped,
                false,
                disable,
            ));
        }
    }
    out
}
fn misleading_pattern(pattern: &str, flags: &str) -> Result<rx::Parsed, rx::ParseError> {
    match rx::parse(pattern, flags) {
        Ok(parsed) => Ok(parsed),
        Err(error) if flags.contains('v') => {
            // The frozen detector visits class characters even in a malformed
            // set union before a later set operator. Keep source byte offsets
            // while separating those operands for the AST-backed detectors.
            let mut recovered = pattern.as_bytes().to_vec();
            let mut depth = 0usize;
            let mut index = 0;
            while index + 1 < recovered.len() {
                if recovered[index] == b'\\' {
                    index += 2;
                    continue;
                }
                match recovered[index] {
                    b'[' => depth += 1,
                    b']' => depth = depth.saturating_sub(1),
                    b'-' | b'&' if depth > 0 && recovered[index + 1] == recovered[index] => {
                        recovered[index] = b']';
                        recovered[index + 1] = b'[';
                        index += 1;
                    }
                    _ => {}
                }
                index += 1;
            }
            let recovered = String::from_utf8(recovered).expect("only ASCII set operators changed");
            rx::parse(&recovered, flags).map_err(|_| error)
        }
        Err(error) => Err(error),
    }
}
fn misleading(
    node: P<Node>,
    argument: P<Node>,
    file: P<SourceFile>,
    pattern: &str,
    flags: &str,
    raw_pattern: Option<(&str, usize)>,
    args: Option<&[P<Node>]>,
) -> Vec<Violation> {
    let (parsed, offset, precise) = if matches!(
        argument.kind(),
        Kind::StringLiteral | Kind::NoSubstitutionTemplateLiteral
    ) {
        let Ok(parsed) = rx::parse_constructor(raw(argument, file), flags) else {
            return Vec::new();
        };
        (parsed, start(argument, file), true)
    } else {
        let Ok(parsed) = misleading_pattern(pattern, flags) else {
            return Vec::new();
        };
        (
            parsed,
            raw_pattern.map_or(0, |(_, offset)| offset),
            raw_pattern.is_some(),
        )
    };
    let uv = flags.contains('u') || flags.contains('v');
    let mut found = Vec::new();
    for class in &parsed.classes {
        for chars in &class.sequences {
            let mut sequences = Vec::new();
            for c in chars {
                sequences.push(c.clone());
            }
            let chars = &sequences;
            for (kind, predicate) in [
                ("surrogatePairWithoutUFlag", 0),
                ("surrogatePair", 1),
                ("combiningClass", 2),
                ("emojiModifier", 3),
                ("regionalIndicatorSymbol", 4),
            ] {
                for pair in chars.windows(2) {
                    let a = &pair[0];
                    let b = &pair[1];
                    let surrogate = (0xd800..=0xdbff).contains(&a.value)
                        && (0xdc00..=0xdfff).contains(&b.value);
                    let brace = |c: &rx::Char| {
                        let source = if precise
                            && matches!(
                                argument.kind(),
                                Kind::StringLiteral | Kind::NoSubstitutionTemplateLiteral
                            ) {
                            raw(argument, file)
                        } else {
                            pattern
                        };
                        source
                            .get(c.start..c.end)
                            .is_some_and(|raw| raw.starts_with("\\u{"))
                    };
                    let matched = match predicate {
                        0 => !uv && surrogate && !brace(a) && !brace(b),
                        1 => uv && surrogate && (brace(a) || brace(b)),
                        2 => rx::is_mark(b.value) && !rx::is_mark(a.value),
                        3 => {
                            (0x1f3fb..=0x1f3ff).contains(&b.value)
                                && !(0x1f3fb..=0x1f3ff).contains(&a.value)
                        }
                        _ => {
                            (0x1f1e6..=0x1f1ff).contains(&a.value)
                                && (0x1f1e6..=0x1f1ff).contains(&b.value)
                        }
                    };
                    if matched {
                        found.push((kind, a.start, b.end));
                    }
                }
            }
            let mut joined: Option<(usize, usize)> = None;
            for window in chars.windows(3) {
                let a = &window[0];
                let b = &window[1];
                let c = &window[2];
                if b.value == 0x200d && a.value != 0x200d && c.value != 0x200d {
                    if let Some((_, end)) = &mut joined {
                        if *end == a.end {
                            *end = c.end;
                            continue;
                        }
                    }
                    if let Some((start, end)) = joined.take() {
                        found.push(("zwj", start, end));
                    }
                    joined = Some((a.start, c.end));
                }
            }
            if let Some((start, end)) = joined {
                found.push(("zwj", start, end));
            }
        }
    }
    let mut seen = std::collections::BTreeSet::new();
    found
        .into_iter()
        .filter(|(kind, _, _)| precise || seen.insert(*kind))
        .map(|(id, begin, finish)| {
            let description = match id {
                "surrogatePairWithoutUFlag" => {
                    "Unexpected surrogate pair in character class. Use 'u' flag."
                }
                "surrogatePair" => "Unexpected surrogate pair in character class.",
                "combiningClass" => "Unexpected combined character in character class.",
                "emojiModifier" => "Unexpected modified Emoji in character class.",
                "regionalIndicatorSymbol" => "Unexpected national flag in character class.",
                _ => "Unexpected joined character sequence in character class.",
            };
            let range = if precise {
                [offset + begin, offset + finish]
            } else {
                [start(argument, file), end(argument)]
            };
            let mut report = Violation::new(node, id, description).with_byte_range(range);
            if id == "surrogatePairWithoutUFlag" && rx::parse(pattern, "u").is_ok() {
                let edits = if let Some(args) = args {
                    if args.len() < 2 {
                        Some(vec![Edit::replace(
                            [end(*args.last().unwrap()), end(*args.last().unwrap())],
                            ", \"u\"",
                        )])
                    } else {
                        let flags = skip(args[1]);
                        matches!(
                            flags.kind(),
                            Kind::StringLiteral | Kind::NoSubstitutionTemplateLiteral
                        )
                        .then(|| vec![Edit::replace([end(flags) - 1, end(flags) - 1], "u")])
                    }
                } else {
                    Some(vec![Edit::replace([end(node), end(node)], "u")])
                };
                if let Some(edits) = edits {
                    report = report.with_suggestion(
                        "suggestUnicodeFlag",
                        "Add unicode 'u' flag to regex.",
                        edits,
                    );
                }
            }
            report
        })
        .collect()
}
fn invalid_regexp(node: P<Node>, _file: P<SourceFile>) -> Vec<Violation> {
    let Some((callee, args)) = call(node) else {
        return Vec::new();
    };
    if callee.kind() != Kind::Identifier || callee.text() != "RegExp" || args.is_empty() {
        return Vec::new();
    }
    let flags = if let Some(&flags) = args.get(1) {
        (flags.kind() == Kind::StringLiteral).then(|| flags.text())
    } else {
        Some("")
    };
    let message = flags.and_then(invalid_flags).or_else(|| {
        if args[0].kind() != Kind::StringLiteral {
            return None;
        }
        let pattern = args[0].text();
        match flags {
            Some(flags) => rx::legacy_error(
                pattern,
                if flags.contains('u') || flags.contains('v') {
                    "u"
                } else {
                    ""
                },
            ),
            None => {
                if rx::legacy_error(pattern, "u").is_none() {
                    None
                } else {
                    rx::legacy_error(pattern, "")
                }
            }
        }
        .map(|error| format!("Invalid regular expression: /{pattern}/: {error}"))
    });
    message
        .map(|message| vec![Violation::new(node, "regexMessage", message + ".")])
        .unwrap_or_default()
}
fn invalid_flags(flags: &str) -> Option<String> {
    if flags.contains('u') && flags.contains('v') {
        return Some("Regex 'u' and 'v' flags cannot be used together".into());
    }
    let mut rest = flags.to_owned();
    for flag in "dgimsuvy".chars() {
        if let Some(index) = rest.find(flag) {
            rest.remove(index);
        }
    }
    let duplicates: String = rest.chars().filter(|&c| "dgimsuvy".contains(c)).collect();
    if !duplicates.is_empty() {
        return Some(format!(
            "Duplicate flags ('{duplicates}') supplied to RegExp constructor"
        ));
    }
    (!rest.is_empty()).then(|| format!("Invalid flags supplied to RegExp constructor '{rest}'"))
}
