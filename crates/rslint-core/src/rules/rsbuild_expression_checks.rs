//! Default expression-rule behavior from frozen rslint e0a29f751d4def86d27931759098f27748fa3003.
use super::{
    rsbuild_assignment_checks::resolve_reference, rsbuild_declaration_checks::static_expression,
    rsbuild_ts_syntax::lexically_shadowed,
};
use crate::{Violation, comments, edits::Edit};
use tsrs_ast::{
    self as ast, Kind, Node, NodeFlags, OperatorPrecedence, OuterExpressionKinds, SourceFile,
    Symbol,
};
use tsrs_checker::Checker;
use tsrs_core::{P, ScriptKind};
use tsrs_scanner::{self as scanner, Scanner};

pub(crate) const NAMES: &[&str] = &[
    "no-extra-boolean-cast",
    "prefer-rest-params",
    "prefer-spread",
    "use-isnan",
];
pub(crate) fn check(
    name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
) -> Vec<Violation> {
    match (name, node.kind()) {
        ("no-extra-boolean-cast", Kind::SourceFile) => boolean_casts(file),
        ("prefer-rest-params", Kind::Identifier) => rest_params(node).into_iter().collect(),
        ("prefer-spread", Kind::CallExpression) => prefer_spread(node, file).into_iter().collect(),
        ("use-isnan", Kind::SourceFile) => use_isnan(file, checker),
        _ => Vec::new(),
    }
}
fn start(node: P<Node>, file: P<SourceFile>) -> usize {
    scanner::skip_trivia(file.text(), node.pos().expect_source("parsed node")) as usize
}
fn end(node: P<Node>) -> usize {
    node.end().expect_source("parsed node") as usize
}
fn range(node: P<Node>, file: P<SourceFile>) -> [usize; 2] {
    [start(node, file), end(node)]
}
fn raw(node: P<Node>, file: P<SourceFile>) -> &'static str {
    &file.text()[start(node, file)..end(node)]
}
fn walk(node: P<Node>, callback: &mut impl FnMut(P<Node>)) {
    if node.flags().intersects(NodeFlags::Reparsed) {
        return;
    }
    callback(node);
    node.for_each_child(&mut |child| {
        walk(child, callback);
        false
    });
}
fn effective_parent(mut child: P<Node>) -> Option<(P<Node>, P<Node>)> {
    while let Some(parent) = child.parent() {
        if parent.kind() != Kind::ParenthesizedExpression {
            return Some((parent, child));
        }
        child = parent;
    }
    None
}
fn boolean_call(node: P<Node>) -> bool {
    matches!(node.kind(), Kind::CallExpression | Kind::NewExpression)
        && node
            .expression()
            .is_some_and(|callee| callee.kind() == Kind::Identifier && callee.text() == "Boolean")
}
fn boolean_context(node: P<Node>) -> bool {
    let Some((parent, child)) = effective_parent(node) else {
        return false;
    };
    match parent.kind() {
        Kind::IfStatement => parent.as_if_statement().expression() == child,
        Kind::WhileStatement => parent.as_while_statement().expression() == child,
        Kind::DoStatement => parent.as_do_statement().expression() == child,
        Kind::ForStatement => parent.as_for_statement().condition() == Some(child),
        Kind::ConditionalExpression => parent.as_conditional_expression().condition() == child,
        Kind::PrefixUnaryExpression => {
            parent.as_prefix_unary_expression().operator() == Kind::ExclamationToken
        }
        Kind::CallExpression | Kind::NewExpression => {
            boolean_call(parent) && parent.arguments().first() == Some(&child)
        }
        _ => false,
    }
}
fn comma(node: P<Node>) -> bool {
    node.kind() == Kind::BinaryExpression
        && node.as_binary_expression().operator_token().kind() == Kind::CommaToken
}
fn replacement_needs_parens(previous: P<Node>, replacement: P<Node>) -> bool {
    let Some(parent) = previous.parent() else {
        return false;
    };
    if parent.kind() == Kind::ParenthesizedExpression {
        return false;
    }
    match parent.kind() {
        Kind::CallExpression | Kind::NewExpression => comma(replacement),
        Kind::ConditionalExpression => {
            if parent.as_conditional_expression().condition() == previous {
                ast::get_expression_precedence(replacement)
                    <= ast::get_expression_precedence(parent)
            } else {
                ast::get_expression_precedence(replacement) < OperatorPrecedence::Assignment
            }
        }
        Kind::PrefixUnaryExpression => {
            ast::get_expression_precedence(replacement) < ast::get_expression_precedence(parent)
        }
        // Default options do not propagate boolean context through binary
        // operators; retaining this branch matches the shared frozen fixer.
        Kind::BinaryExpression => {
            let binary = parent.as_binary_expression();
            let op = binary.operator_token().kind();
            if op == Kind::CommaToken {
                return false;
            }
            if replacement.kind() == Kind::BinaryExpression {
                let replacement_op = replacement.as_binary_expression().operator_token().kind();
                let logical = |op| matches!(op, Kind::BarBarToken | Kind::AmpersandAmpersandToken);
                if (op == Kind::QuestionQuestionToken && logical(replacement_op))
                    || (replacement_op == Kind::QuestionQuestionToken && logical(op))
                {
                    return true;
                }
            }
            if binary.left() == previous {
                ast::get_expression_precedence(replacement) < ast::get_expression_precedence(parent)
            } else {
                ast::get_expression_precedence(replacement)
                    <= ast::get_expression_precedence(parent)
            }
        }
        _ => false,
    }
}
fn prefix_space(file: P<SourceFile>, begin: usize, replacement: &str) -> String {
    if replacement
        .chars()
        .next()
        .is_some_and(|first| scanner::is_identifier_start(first as u32))
        && file.text()[..begin]
            .chars()
            .next_back()
            .is_some_and(|previous| scanner::is_identifier_part(previous as u32))
    {
        format!(" {replacement}")
    } else {
        replacement.to_owned()
    }
}
fn boolean_casts(file: P<SourceFile>) -> Vec<Violation> {
    // Collect once per file, through AST-owned token gaps, so a comment-like
    // sequence in a string/template/regexp never suppresses a safe fix.
    let mut comment_ranges = Vec::new();
    let comments_ready = comments::for_each_comment(file, |comment| {
        comment_ranges.push([comment.pos() as usize, comment.end() as usize]);
        Ok(())
    })
    .is_ok();
    let has_comment = |[begin, end]: [usize; 2]| {
        !comments_ready || comment_ranges.iter().any(|&[a, b]| a < end && b > begin)
    };
    let mut out = Vec::new();
    walk(file.as_node(), &mut |node| {
        let negation = node.kind() == Kind::PrefixUnaryExpression
            && node.as_prefix_unary_expression().operator() == Kind::ExclamationToken
            && {
                let operand = node.as_prefix_unary_expression().operand();
                operand.kind() == Kind::PrefixUnaryExpression
                    && operand.as_prefix_unary_expression().operator() == Kind::ExclamationToken
            };
        let call = node.kind() == Kind::CallExpression && boolean_call(node);
        if (!negation && !call) || !boolean_context(node) {
            return;
        }
        let mut violation = if negation {
            Violation::new(node, "unexpectedNegation", "Redundant double negation.")
        } else {
            Violation::new(node, "unexpectedCall", "Redundant Boolean call.")
        };
        let (target, replacement) = if negation {
            (
                node,
                node.as_prefix_unary_expression()
                    .operand()
                    .as_prefix_unary_expression()
                    .operand(),
            )
        } else if node.arguments().is_empty() {
            let target = effective_parent(node)
                .map(|(parent, _)| parent)
                .filter(|parent| {
                    parent.kind() == Kind::PrefixUnaryExpression
                        && parent.as_prefix_unary_expression().operator() == Kind::ExclamationToken
                })
                .unwrap_or(node);
            let target_range = range(target, file);
            if !has_comment(target_range) {
                violation = violation.with_fix(vec![Edit::replace(
                    target_range,
                    prefix_space(
                        file,
                        target_range[0],
                        if target == node { "false" } else { "true" },
                    ),
                )]);
            }
            out.push(violation);
            return;
        } else if node.arguments().len() == 1 && node.arguments()[0].kind() != Kind::SpreadElement {
            (node, node.arguments()[0])
        } else {
            out.push(violation);
            return;
        };
        let target_range = range(target, file);
        if !has_comment(target_range) {
            let inner = ast::skip_parentheses(replacement);
            let text = if replacement_needs_parens(target, inner) {
                format!("({})", raw(inner, file))
            } else {
                prefix_space(file, target_range[0], raw(inner, file))
            };
            violation = violation.with_fix(vec![Edit::replace(target_range, text)]);
        }
        out.push(violation);
    });
    out
}
fn simple_arguments(node: P<Node>) -> bool {
    node.kind() == Kind::Identifier && node.text() == "arguments"
}
fn list_declares_arguments(list: P<Node>) -> bool {
    list.kind() == Kind::VariableDeclarationList
        && list
            .as_variable_declaration_list()
            .declarations()
            .nodes()
            .iter()
            .any(|decl| simple_arguments(decl.as_variable_declaration().name()))
}
fn hoisted_arguments(node: P<Node>) -> bool {
    if ast::is_function_like_declaration(node) {
        return false;
    }
    if node.kind() == Kind::VariableDeclaration
        && node.parent().is_some_and(|parent| {
            parent.kind() == Kind::VariableDeclarationList
                && !parent.flags().intersects(NodeFlags::BlockScoped)
        })
        && simple_arguments(node.as_variable_declaration().name())
    {
        return true;
    }
    node.for_each_child(&mut hoisted_arguments)
}
fn rest_params(node: P<Node>) -> Option<Violation> {
    if !simple_arguments(node) || ast::is_identifier_name(node) {
        return None;
    }
    let parent = node.parent()?;
    if matches!(
        parent.kind(),
        Kind::Parameter
            | Kind::VariableDeclaration
            | Kind::FunctionDeclaration
            | Kind::FunctionExpression
            | Kind::BindingElement
    ) && parent.name() == Some(node)
    {
        return None;
    }
    let mut function = parent;
    while !ast::is_function_like_declaration(function) || function.kind() == Kind::ArrowFunction {
        function = function.parent()?;
    }
    if function
        .parameters()
        .iter()
        .any(|parameter| parameter.name().is_some_and(simple_arguments))
        || function.body().is_some_and(hoisted_arguments)
    {
        return None;
    }
    let mut scope = parent;
    while scope != function {
        let shadowed = match scope.kind() {
            Kind::Block | Kind::ModuleBlock => scope.statements().iter().any(|statement| {
                if statement.kind() != Kind::VariableStatement {
                    return false;
                }
                let list = statement.as_variable_statement().declaration_list();
                list.flags().intersects(NodeFlags::BlockScoped) && list_declares_arguments(list)
            }),
            Kind::ForInStatement | Kind::ForOfStatement => {
                list_declares_arguments(scope.as_for_in_or_of_statement().initializer())
            }
            Kind::CatchClause => scope
                .as_catch_clause()
                .variable_declaration()
                .is_some_and(|decl| simple_arguments(decl.as_variable_declaration().name())),
            _ => false,
        };
        if shadowed {
            return None;
        }
        scope = scope.parent()?;
    }
    if parent.kind() == Kind::PropertyAccessExpression
        && parent.as_property_access_expression().expression() == node
    {
        return None;
    }
    Some(Violation::new(
        node,
        "preferRestParams",
        "Use the rest parameters instead of 'arguments'.",
    ))
}
fn prefer_spread(node: P<Node>, file: P<SourceFile>) -> Option<Violation> {
    let call = node.as_call_expression();
    let callee = ast::skip_parentheses(call.expression());
    let object = match callee.kind() {
        Kind::PropertyAccessExpression
            if callee
                .name()
                .is_some_and(|name| name.kind() == Kind::Identifier && name.text() == "apply") =>
        {
            callee.as_property_access_expression().expression()
        }
        Kind::ElementAccessExpression
            if static_expression(
                ast::skip_parentheses(callee.as_element_access_expression().argument_expression()),
                file,
            )
            .as_deref()
                == Some("apply") =>
        {
            callee.as_element_access_expression().expression()
        }
        _ => return None,
    };
    let args = node.arguments();
    if args.len() != 2
        || matches!(
            ast::skip_parentheses(args[1]).kind(),
            Kind::ArrayLiteralExpression | Kind::SpreadElement
        )
    {
        return None;
    }
    let applied = ast::skip_parentheses(object);
    let valid = if ast::is_access_expression(applied) {
        same_tokens(
            ast::skip_parentheses(applied.expression()?),
            ast::skip_parentheses(args[0]),
            file,
        )
    } else {
        let argument = ast::skip_parentheses(args[0]);
        matches!(argument.kind(), Kind::NullKeyword | Kind::VoidExpression)
            || (argument.kind() == Kind::Identifier && argument.text() == "undefined")
    };
    valid.then(|| {
        Violation::new(
            node,
            "preferSpread",
            "Use the spread operator instead of '.apply()'.",
        )
    })
}
fn local_value(symbol: P<Symbol>, file: P<SourceFile>) -> bool {
    symbol.declarations().iter().any(|&decl| {
        if matches!(
            decl.kind(),
            Kind::InterfaceDeclaration | Kind::TypeAliasDeclaration
        ) || ast::get_source_file_of_node(decl) != Some(file)
        {
            return false;
        }
        let mut parent = decl.parent();
        while let Some(node) = parent {
            if node.kind() == Kind::ModuleDeclaration
                && (ast::is_global_scope_augmentation(node)
                    || ast::is_external_module_augmentation(node))
            {
                return false;
            }
            parent = node.parent();
        }
        true
    })
}
fn global_reference(
    node: P<Node>,
    may_shadow: &[bool; 2],
    file: P<SourceFile>,
    checker: &mut Checker,
) -> bool {
    let mut parent = node.parent();
    while let Some(ancestor) = parent {
        if matches!(
            ancestor.kind(),
            Kind::FunctionExpression | Kind::ClassExpression
        ) && ancestor
            .name()
            .is_some_and(|name| name.kind() == Kind::Identifier && name.text() == node.text())
        {
            return false;
        }
        parent = ancestor.parent();
    }
    if !may_shadow[usize::from(node.text() == "Number")] {
        return true;
    }
    if let Some(symbol) = resolve_reference(node, checker) {
        return !local_value(symbol, file);
    }
    !lexically_shadowed(node, node.text())
}
fn is_nan(
    mut node: P<Node>,
    may_shadow: &[bool; 2],
    file: P<SourceFile>,
    checker: &mut Checker,
) -> bool {
    node = ast::skip_parentheses(node);
    if node.kind() == Kind::BinaryExpression
        && node.as_binary_expression().operator_token().kind() == Kind::CommaToken
    {
        node = ast::skip_parentheses(node.as_binary_expression().right());
    }
    if node.kind() == Kind::Identifier {
        return node.text() == "NaN" && global_reference(node, may_shadow, file, checker);
    }
    let object = match node.kind() {
        Kind::PropertyAccessExpression
            if node
                .name()
                .is_some_and(|name| name.kind() == Kind::Identifier && name.text() == "NaN") =>
        {
            node.as_property_access_expression().expression()
        }
        Kind::ElementAccessExpression
            if static_expression(
                ast::skip_outer_expressions(
                    node.as_element_access_expression().argument_expression(),
                    OuterExpressionKinds::Parentheses | OuterExpressionKinds::Assertions,
                ),
                file,
            )
            .as_deref()
                == Some("NaN") =>
        {
            node.as_element_access_expression().expression()
        }
        _ => return false,
    };
    let object = ast::skip_parentheses(object);
    object.kind() == Kind::Identifier
        && object.text() == "Number"
        && global_reference(object, may_shadow, file, checker)
}
fn use_isnan(file: P<SourceFile>, checker: &mut Checker) -> Vec<Violation> {
    let mut may_shadow = [false; 2];
    walk(file.as_node(), &mut |node| {
        if let Some(locals) = node.locals() {
            for (index, name) in ["NaN", "Number"].iter().enumerate() {
                may_shadow[index] |= locals
                    .get()
                    .get(name)
                    .is_some_and(|symbol| local_value(symbol, file));
            }
        }
    });
    let mut out = Vec::new();
    walk(file.as_node(), &mut |node| match node.kind() {
        Kind::BinaryExpression => {
            let binary = node.as_binary_expression();
            if matches!(
                binary.operator_token().kind(),
                Kind::EqualsEqualsToken
                    | Kind::EqualsEqualsEqualsToken
                    | Kind::ExclamationEqualsToken
                    | Kind::ExclamationEqualsEqualsToken
                    | Kind::LessThanToken
                    | Kind::LessThanEqualsToken
                    | Kind::GreaterThanToken
                    | Kind::GreaterThanEqualsToken
            ) && (is_nan(binary.left(), &may_shadow, file, checker)
                || is_nan(binary.right(), &may_shadow, file, checker))
            {
                out.push(Violation::new(
                    node,
                    "comparisonWithNaN",
                    "Use the isNaN function to compare with NaN.",
                ));
            }
        }
        Kind::SwitchStatement => {
            let switch = node.as_switch_statement();
            if is_nan(switch.expression(), &may_shadow, file, checker) {
                out.push(Violation::new(node,"switchNaN","'switch(NaN)' can never match a case clause. Use Number.isNaN instead of the switch."));
            }
            for &clause in switch.case_block().as_case_block().clauses().nodes() {
                if clause.kind() == Kind::CaseClause
                    && is_nan(
                        clause.as_case_or_default_clause().expression().unwrap(),
                        &may_shadow,
                        file,
                        checker,
                    )
                {
                    out.push(Violation::new(
                        clause,
                        "caseNaN",
                        "'case NaN' can never match. Use Number.isNaN before the switch.",
                    ));
                }
            }
        }
        _ => {}
    });
    out
}

// The frozen HasSameTokens compares AST-owned leaves and separately scans only
// the gaps. This preserves regexp/template/JSX lexical state and punctuation,
// including inner grouping parentheses and trailing commas.
fn same_tokens(left: P<Node>, right: P<Node>, file: P<SourceFile>) -> bool {
    let jsx_text = |kind| matches!(kind, Kind::JsxText | Kind::JsxTextAllWhiteSpaces);
    if jsx_text(left.kind()) || jsx_text(right.kind()) {
        if !jsx_text(left.kind()) || !jsx_text(right.kind()) {
            return false;
        }
        let a = &file.text()[left.pos().expect_source("parsed JSX") as usize..end(left)];
        let b = &file.text()[right.pos().expect_source("parsed JSX") as usize..end(right)];
        return match file.script_kind() {
            ScriptKind::JS | ScriptKind::JSX => jsx_text_units(a) == jsx_text_units(b),
            ScriptKind::TS | ScriptKind::TSX => a == b,
            _ => left.kind() == right.kind() && a == b,
        };
    }
    if left.kind() != right.kind() {
        return false;
    }
    let children = |node: P<Node>| {
        let mut children = Vec::new();
        node.for_each_child(&mut |child| {
            if !child.flags().intersects(NodeFlags::Reparsed) {
                children.push(child);
            }
            false
        });
        children
    };
    let a = children(left);
    let b = children(right);
    if a.is_empty() && b.is_empty() {
        if matches!(
            left.kind(),
            Kind::ArrayLiteralExpression | Kind::ObjectLiteralExpression
        ) {
            return gap_tokens(
                file,
                left.pos().expect_source("parsed node") as usize,
                end(left),
            ) == gap_tokens(
                file,
                right.pos().expect_source("parsed node") as usize,
                end(right),
            );
        }
        if matches!(left.kind(), Kind::Identifier | Kind::PrivateIdentifier) {
            return left.text() == right.text()
                && (!matches!(file.script_kind(), ScriptKind::TS | ScriptKind::TSX)
                    || raw(left, file) == raw(right, file));
        }
        return raw(left, file) == raw(right, file);
    }
    if a.len() != b.len() {
        return false;
    }
    let mut a_pos = left.pos().expect_source("parsed node") as usize;
    let mut b_pos = right.pos().expect_source("parsed node") as usize;
    for (&a, &b) in a.iter().zip(&b) {
        if gap_tokens(file, a_pos, a.pos().expect_source("parsed node") as usize)
            != gap_tokens(file, b_pos, b.pos().expect_source("parsed node") as usize)
            || !same_tokens(a, b, file)
        {
            return false;
        }
        a_pos = end(a);
        b_pos = end(b);
    }
    gap_tokens(file, a_pos, end(left)) == gap_tokens(file, b_pos, end(right))
}
fn gap_tokens(file: P<SourceFile>, begin: usize, end: usize) -> Vec<(Kind, &'static str)> {
    let mut result = Vec::new();
    if begin >= end {
        return result;
    }
    let mut scanner = Scanner::new();
    scanner.set_text(file.text());
    scanner.set_language_variant(file.language_variant());
    scanner.reset_pos(begin as u32);
    loop {
        let kind = scanner.scan();
        let start = scanner.token_start() as usize;
        let finish = scanner.token_end() as usize;
        if kind == Kind::EndOfFile || start >= end || finish > end {
            break;
        }
        result.push((kind, &file.text()[start..finish]));
    }
    result
}
// Espree/acorn-jsx reads named XHTML entities and numeric references as UTF-16
// code units, folding only literal CRLF. A semicolon must be within ten UTF-16
// units after '&'. TypeScript token equality keeps raw JSX source instead.
fn jsx_text_units(text: &str) -> Vec<u16> {
    let mut units = Vec::new();
    let mut index = 0;
    while index < text.len() {
        let rest = &text[index..];
        if let Some(entity) = rest.strip_prefix('&') {
            let mut width = 0;
            let mut decoded = None;
            for (pos, ch) in entity.char_indices() {
                width += ch.len_utf16();
                if width > 10 {
                    break;
                }
                if ch == ';' {
                    decoded = decode_jsx_entity(&entity[..pos]).map(|value| (value, pos + 2));
                    break;
                }
            }
            if let Some((value, consumed)) = decoded {
                if entity.starts_with('#') {
                    units.push(value as u16);
                } else if let Some(ch) = char::from_u32(value) {
                    let mut buffer = [0; 2];
                    units.extend_from_slice(ch.encode_utf16(&mut buffer));
                }
                index += consumed;
                continue;
            }
        }
        if rest.starts_with("\r\n") {
            units.push(b'\n' as u16);
            index += 2;
            continue;
        }
        let ch = rest.chars().next().unwrap();
        let mut buffer = [0; 2];
        units.extend_from_slice(ch.encode_utf16(&mut buffer));
        index += ch.len_utf8();
    }
    units
}
fn decode_jsx_entity(entity: &str) -> Option<u32> {
    if let Some(number) = entity.strip_prefix('#') {
        let (base, number) = number
            .strip_prefix('x')
            .map_or((10, number), |number| (16, number));
        if number.is_empty()
            || !number.bytes().all(|byte| {
                if base == 10 {
                    byte.is_ascii_digit()
                } else {
                    byte.is_ascii_hexdigit()
                }
            })
        {
            return None;
        }
        return u32::from_str_radix(number, base).ok();
    }
    jsx_named_entity(entity)
}

// XHTML entity mapping adapted from tsrs jsxtransforms/jsx.rs at
// 280b426b89cd4ecd6153cf48228826bd60a99c61 (MIT, TypeScript contributors).
// The upstream decoder is private; no additional runtime/compiler dependency.
fn jsx_named_entity(name: &str) -> Option<u32> {
    Some(match name {
        "quot" => 0x0022,
        "amp" => 0x0026,
        "apos" => 0x0027,
        "lt" => 0x003C,
        "gt" => 0x003E,
        "nbsp" => 0x00A0,
        "iexcl" => 0x00A1,
        "cent" => 0x00A2,
        "pound" => 0x00A3,
        "curren" => 0x00A4,
        "yen" => 0x00A5,
        "brvbar" => 0x00A6,
        "sect" => 0x00A7,
        "uml" => 0x00A8,
        "copy" => 0x00A9,
        "ordf" => 0x00AA,
        "laquo" => 0x00AB,
        "not" => 0x00AC,
        "shy" => 0x00AD,
        "reg" => 0x00AE,
        "macr" => 0x00AF,
        "deg" => 0x00B0,
        "plusmn" => 0x00B1,
        "sup2" => 0x00B2,
        "sup3" => 0x00B3,
        "acute" => 0x00B4,
        "micro" => 0x00B5,
        "para" => 0x00B6,
        "middot" => 0x00B7,
        "cedil" => 0x00B8,
        "sup1" => 0x00B9,
        "ordm" => 0x00BA,
        "raquo" => 0x00BB,
        "frac14" => 0x00BC,
        "frac12" => 0x00BD,
        "frac34" => 0x00BE,
        "iquest" => 0x00BF,
        "Agrave" => 0x00C0,
        "Aacute" => 0x00C1,
        "Acirc" => 0x00C2,
        "Atilde" => 0x00C3,
        "Auml" => 0x00C4,
        "Aring" => 0x00C5,
        "AElig" => 0x00C6,
        "Ccedil" => 0x00C7,
        "Egrave" => 0x00C8,
        "Eacute" => 0x00C9,
        "Ecirc" => 0x00CA,
        "Euml" => 0x00CB,
        "Igrave" => 0x00CC,
        "Iacute" => 0x00CD,
        "Icirc" => 0x00CE,
        "Iuml" => 0x00CF,
        "ETH" => 0x00D0,
        "Ntilde" => 0x00D1,
        "Ograve" => 0x00D2,
        "Oacute" => 0x00D3,
        "Ocirc" => 0x00D4,
        "Otilde" => 0x00D5,
        "Ouml" => 0x00D6,
        "times" => 0x00D7,
        "Oslash" => 0x00D8,
        "Ugrave" => 0x00D9,
        "Uacute" => 0x00DA,
        "Ucirc" => 0x00DB,
        "Uuml" => 0x00DC,
        "Yacute" => 0x00DD,
        "THORN" => 0x00DE,
        "szlig" => 0x00DF,
        "agrave" => 0x00E0,
        "aacute" => 0x00E1,
        "acirc" => 0x00E2,
        "atilde" => 0x00E3,
        "auml" => 0x00E4,
        "aring" => 0x00E5,
        "aelig" => 0x00E6,
        "ccedil" => 0x00E7,
        "egrave" => 0x00E8,
        "eacute" => 0x00E9,
        "ecirc" => 0x00EA,
        "euml" => 0x00EB,
        "igrave" => 0x00EC,
        "iacute" => 0x00ED,
        "icirc" => 0x00EE,
        "iuml" => 0x00EF,
        "eth" => 0x00F0,
        "ntilde" => 0x00F1,
        "ograve" => 0x00F2,
        "oacute" => 0x00F3,
        "ocirc" => 0x00F4,
        "otilde" => 0x00F5,
        "ouml" => 0x00F6,
        "divide" => 0x00F7,
        "oslash" => 0x00F8,
        "ugrave" => 0x00F9,
        "uacute" => 0x00FA,
        "ucirc" => 0x00FB,
        "uuml" => 0x00FC,
        "yacute" => 0x00FD,
        "thorn" => 0x00FE,
        "yuml" => 0x00FF,
        "OElig" => 0x0152,
        "oelig" => 0x0153,
        "Scaron" => 0x0160,
        "scaron" => 0x0161,
        "Yuml" => 0x0178,
        "fnof" => 0x0192,
        "circ" => 0x02C6,
        "tilde" => 0x02DC,
        "Alpha" => 0x0391,
        "Beta" => 0x0392,
        "Gamma" => 0x0393,
        "Delta" => 0x0394,
        "Epsilon" => 0x0395,
        "Zeta" => 0x0396,
        "Eta" => 0x0397,
        "Theta" => 0x0398,
        "Iota" => 0x0399,
        "Kappa" => 0x039A,
        "Lambda" => 0x039B,
        "Mu" => 0x039C,
        "Nu" => 0x039D,
        "Xi" => 0x039E,
        "Omicron" => 0x039F,
        "Pi" => 0x03A0,
        "Rho" => 0x03A1,
        "Sigma" => 0x03A3,
        "Tau" => 0x03A4,
        "Upsilon" => 0x03A5,
        "Phi" => 0x03A6,
        "Chi" => 0x03A7,
        "Psi" => 0x03A8,
        "Omega" => 0x03A9,
        "alpha" => 0x03B1,
        "beta" => 0x03B2,
        "gamma" => 0x03B3,
        "delta" => 0x03B4,
        "epsilon" => 0x03B5,
        "zeta" => 0x03B6,
        "eta" => 0x03B7,
        "theta" => 0x03B8,
        "iota" => 0x03B9,
        "kappa" => 0x03BA,
        "lambda" => 0x03BB,
        "mu" => 0x03BC,
        "nu" => 0x03BD,
        "xi" => 0x03BE,
        "omicron" => 0x03BF,
        "pi" => 0x03C0,
        "rho" => 0x03C1,
        "sigmaf" => 0x03C2,
        "sigma" => 0x03C3,
        "tau" => 0x03C4,
        "upsilon" => 0x03C5,
        "phi" => 0x03C6,
        "chi" => 0x03C7,
        "psi" => 0x03C8,
        "omega" => 0x03C9,
        "thetasym" => 0x03D1,
        "upsih" => 0x03D2,
        "piv" => 0x03D6,
        "ensp" => 0x2002,
        "emsp" => 0x2003,
        "thinsp" => 0x2009,
        "zwnj" => 0x200C,
        "zwj" => 0x200D,
        "lrm" => 0x200E,
        "rlm" => 0x200F,
        "ndash" => 0x2013,
        "mdash" => 0x2014,
        "lsquo" => 0x2018,
        "rsquo" => 0x2019,
        "sbquo" => 0x201A,
        "ldquo" => 0x201C,
        "rdquo" => 0x201D,
        "bdquo" => 0x201E,
        "dagger" => 0x2020,
        "Dagger" => 0x2021,
        "bull" => 0x2022,
        "hellip" => 0x2026,
        "permil" => 0x2030,
        "prime" => 0x2032,
        "Prime" => 0x2033,
        "lsaquo" => 0x2039,
        "rsaquo" => 0x203A,
        "oline" => 0x203E,
        "frasl" => 0x2044,
        "euro" => 0x20AC,
        "image" => 0x2111,
        "weierp" => 0x2118,
        "real" => 0x211C,
        "trade" => 0x2122,
        "alefsym" => 0x2135,
        "larr" => 0x2190,
        "uarr" => 0x2191,
        "rarr" => 0x2192,
        "darr" => 0x2193,
        "harr" => 0x2194,
        "crarr" => 0x21B5,
        "lArr" => 0x21D0,
        "uArr" => 0x21D1,
        "rArr" => 0x21D2,
        "dArr" => 0x21D3,
        "hArr" => 0x21D4,
        "forall" => 0x2200,
        "part" => 0x2202,
        "exist" => 0x2203,
        "empty" => 0x2205,
        "nabla" => 0x2207,
        "isin" => 0x2208,
        "notin" => 0x2209,
        "ni" => 0x220B,
        "prod" => 0x220F,
        "sum" => 0x2211,
        "minus" => 0x2212,
        "lowast" => 0x2217,
        "radic" => 0x221A,
        "prop" => 0x221D,
        "infin" => 0x221E,
        "ang" => 0x2220,
        "and" => 0x2227,
        "or" => 0x2228,
        "cap" => 0x2229,
        "cup" => 0x222A,
        "int" => 0x222B,
        "there4" => 0x2234,
        "sim" => 0x223C,
        "cong" => 0x2245,
        "asymp" => 0x2248,
        "ne" => 0x2260,
        "equiv" => 0x2261,
        "le" => 0x2264,
        "ge" => 0x2265,
        "sub" => 0x2282,
        "sup" => 0x2283,
        "nsub" => 0x2284,
        "sube" => 0x2286,
        "supe" => 0x2287,
        "oplus" => 0x2295,
        "otimes" => 0x2297,
        "perp" => 0x22A5,
        "sdot" => 0x22C5,
        "lceil" => 0x2308,
        "rceil" => 0x2309,
        "lfloor" => 0x230A,
        "rfloor" => 0x230B,
        "lang" => 0x2329,
        "rang" => 0x232A,
        "loz" => 0x25CA,
        "spades" => 0x2660,
        "clubs" => 0x2663,
        "hearts" => 0x2665,
        "diams" => 0x2666,
        _ => return None,
    })
}
