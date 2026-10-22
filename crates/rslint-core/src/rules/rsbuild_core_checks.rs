//! Core checks enabled by the frozen Rsbuild configuration. The decisions,
//! spans and suggestions follow rslint e0a29f751d4def86d27931759098f27748fa3003.
//! These checks use the frozen configuration's default rule options. The
//! request boundary rejects non-default options instead of dropping them.

use std::collections::{BTreeMap, HashSet};

use tsrs_ast::{
    self as ast, Kind, Node, NodeFlags, OuterExpressionKinds, SourceFile, Symbol, SymbolFlags,
};
use tsrs_checker::Checker;
use tsrs_core::P;
use tsrs_scanner::skip_trivia;

use crate::{Violation, edits::Edit};

pub(crate) const NAMES: &[&str] = &[
    "no-async-promise-executor",
    "no-empty",
    "no-unsafe-negation",
    "valid-typeof",
    "for-direction",
];

pub(crate) fn check(
    name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
) -> Vec<Violation> {
    let report = match (name, node.kind()) {
        ("no-async-promise-executor", Kind::NewExpression) => async_executor(node, file, checker),
        ("no-empty", Kind::Block | Kind::SwitchStatement) => empty_statement(node, file),
        ("no-unsafe-negation", Kind::BinaryExpression) => unsafe_negation(node, file),
        ("valid-typeof", Kind::TypeOfExpression) => valid_typeof(node, file),
        ("for-direction", Kind::ForStatement) => for_direction(node, file, checker),
        _ => None,
    };
    report.into_iter().collect()
}

fn end(node: P<Node>) -> usize {
    node.end().expect_source("parsed node end") as usize
}

fn start(node: P<Node>, file: P<SourceFile>) -> usize {
    skip_trivia(file.text(), node.pos().expect_source("parsed node start")) as usize
}

fn async_executor(node: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> Option<Violation> {
    let expression = node.as_new_expression();
    let callee = ast::skip_parentheses(expression.expression());
    if callee.kind() != Kind::Identifier || callee.text() != "Promise" {
        return None;
    }
    let executor = ast::skip_parentheses(*expression.arguments()?.nodes().first()?);
    if !matches!(
        executor.kind(),
        Kind::FunctionExpression | Kind::ArrowFunction
    ) {
        return None;
    }
    let modifiers = executor.modifiers()?;
    let modifier = *modifiers
        .nodes()
        .iter()
        .find(|modifier| modifier.kind() == Kind::AsyncKeyword)?;
    // The frozen rule uses binder resolution in the value/namespace/alias
    // spaces. A type-only binding in a module does not hide Promise's value;
    // a script-level declaration of any kind hides the configured global.
    if ast::is_global_source_file(file.as_node())
        && file
            .as_node()
            .locals()
            .is_some_and(|locals| locals.get().get("Promise").is_some())
    {
        return None;
    }
    if checker
        .resolve_name_exported(
            "Promise",
            callee,
            SymbolFlags::Value | SymbolFlags::Namespace | SymbolFlags::Alias,
            true,
        )
        .is_some()
    {
        return None;
    }
    Some(
        Violation::new(
            modifier,
            "async",
            "Promise executor functions should not be async.",
        )
        .with_byte_range([start(modifier, file), end(modifier)]),
    )
}

fn empty_statement(node: P<Node>, file: P<SourceFile>) -> Option<Violation> {
    let (braces, label) = if node.kind() == Kind::Block {
        if !node.statements().is_empty()
            || node
                .parent()
                .is_some_and(ast::is_function_like_or_class_static_block_declaration)
        {
            return None;
        }
        (node, "block")
    } else {
        let block = node.as_switch_statement().case_block();
        if !block.as_case_block().clauses().nodes().is_empty() {
            return None;
        }
        (block, "switch")
    };
    let begin = start(braces, file);
    let finish = end(braces);
    let inner = if finish <= begin + 1 {
        [begin, begin]
    } else {
        [begin + 1, finish - 1]
    };
    // A parsed empty braced node contains only trivia, so comment openers here
    // cannot be string, regexp, template or JSX contents.
    let text = &file.text()[inner[0]..inner[1]];
    if text.contains("//") || text.contains("/*") {
        return None;
    }
    let mut report = Violation::new(braces, "unexpected", format!("Empty {label} statement."))
        .with_byte_range([begin, finish])
        .with_suggestion(
            "suggestComment",
            format!("Add comment inside empty {label} statement."),
            vec![Edit::replace(inner, " /* empty */ ")],
        );
    report.suggestions[0].data = Some(BTreeMap::from([("type".to_owned(), label.to_owned())]));
    Some(report)
}

fn unsafe_negation(node: P<Node>, file: P<SourceFile>) -> Option<Violation> {
    let binary = node.as_binary_expression();
    let operator = match binary.operator_token().kind() {
        Kind::InKeyword => "in",
        Kind::InstanceOfKeyword => "instanceof",
        _ => return None,
    };
    let left = binary.left();
    if left.kind() != Kind::PrefixUnaryExpression
        || left.as_prefix_unary_expression().operator() != Kind::ExclamationToken
    {
        return None;
    }
    let begin = start(left, file);
    let negation_end = begin + 1;
    Some(Violation::new(left, "unexpected", format!("Unexpected negating the left operand of '{operator}' operator."))
        .with_suggestion("suggestNegatedExpression", format!("Negate '{operator}' expression instead of its left operand. This changes the current behavior."),
            vec![Edit::replace([negation_end, end(node)], format!("({})", &file.text()[negation_end..end(node)]))])
        .with_suggestion("suggestParenthesisedNegation", "Wrap negation in '()' to make the intention explicit. This preserves the current behavior.",
            vec![Edit::replace([begin, end(left)], format!("({})", &file.text()[begin..end(left)]))]))
}

fn valid_typeof(node: P<Node>, file: P<SourceFile>) -> Option<Violation> {
    let mut parent = node.parent()?;
    while parent.kind() == Kind::ParenthesizedExpression {
        parent = parent.parent()?;
    }
    if parent.kind() != Kind::BinaryExpression {
        return None;
    }
    let binary = parent.as_binary_expression();
    if !matches!(
        binary.operator_token().kind(),
        Kind::EqualsEqualsToken
            | Kind::EqualsEqualsEqualsToken
            | Kind::ExclamationEqualsToken
            | Kind::ExclamationEqualsEqualsToken
    ) {
        return None;
    }
    let sibling = ast::skip_parentheses(if ast::skip_parentheses(binary.left()) == node {
        binary.right()
    } else {
        binary.left()
    });
    let report = || Violation::new(sibling, "invalidValue", "Invalid typeof comparison value.");
    match sibling.kind() {
        Kind::StringLiteral | Kind::NoSubstitutionTemplateLiteral => (!matches!(
            sibling.text(),
            "undefined"
                | "object"
                | "boolean"
                | "number"
                | "string"
                | "function"
                | "symbol"
                | "bigint"
        ))
        .then(report),
        Kind::NumericLiteral
        | Kind::BigIntLiteral
        | Kind::RegularExpressionLiteral
        | Kind::TrueKeyword
        | Kind::FalseKeyword
        | Kind::NullKeyword => Some(report()),
        Kind::Identifier
            if sibling.text() == "undefined"
                && !super::rsbuild_ts_syntax::lexically_shadowed(sibling, "undefined") =>
        {
            Some(report().with_suggestion(
                "suggestString",
                "Use \"undefined\" instead of undefined.",
                vec![Edit::replace(
                    [start(sibling, file), end(sibling)],
                    "\"undefined\"",
                )],
            ))
        }
        _ => None,
    }
}

fn for_direction(node: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> Option<Violation> {
    let statement = node.as_for_statement();
    let condition = statement.condition()?;
    if condition.kind() != Kind::BinaryExpression {
        return None;
    }
    let binary = condition.as_binary_expression();
    let less = match binary.operator_token().kind() {
        Kind::LessThanToken | Kind::LessThanEqualsToken => true,
        Kind::GreaterThanToken | Kind::GreaterThanEqualsToken => false,
        _ => return None,
    };
    let incrementor = statement.incrementor()?;
    let update = ast::skip_parentheses(incrementor);
    if !matches!(
        update.kind(),
        Kind::PostfixUnaryExpression | Kind::PrefixUnaryExpression | Kind::BinaryExpression
    ) {
        return None;
    }
    let mut evaluator = NumericEvaluator {
        checker,
        file,
        resolving: HashSet::new(),
    };
    for (operand, left) in [(binary.left(), true), (binary.right(), false)] {
        let operand = ast::skip_parentheses(operand);
        if operand.kind() != Kind::Identifier {
            continue;
        }
        let (direction, count) = evaluator.modification(update, operand.text());
        if count == 1 && direction == if less == left { -1 } else { 1 } {
            let close = skip_trivia(file.text(), end(incrementor) as u32) as usize;
            let finish = if file.text().as_bytes().get(close) == Some(&b')') {
                close + 1
            } else {
                end(incrementor)
            };
            return Some(
                Violation::new(
                    node,
                    "incorrectDirection",
                    "The update clause in this loop moves the variable in the wrong direction.",
                )
                .with_byte_range([start(node, file), finish]),
            );
        }
    }
    None
}

#[derive(Clone, Copy, PartialEq, Eq)]
enum NumericKind {
    Number,
    BigInt,
    Boolean,
}
#[derive(Clone, Copy)]
struct Numeric {
    value: f64,
    kind: NumericKind,
}
impl Numeric {
    fn number(value: f64) -> Self {
        Self {
            value,
            kind: NumericKind::Number,
        }
    }
    fn boolean(value: bool) -> Self {
        Self {
            value: if value { 1.0 } else { 0.0 },
            kind: NumericKind::Boolean,
        }
    }
    fn truthy(self) -> bool {
        self.value != 0.0 && !self.value.is_nan()
    }
}
struct NumericEvaluator<'a> {
    checker: &'a mut Checker,
    file: P<SourceFile>,
    resolving: HashSet<P<Symbol>>,
}
impl NumericEvaluator<'_> {
    fn modification(&mut self, node: P<Node>, counter: &str) -> (i8, usize) {
        let node = ast::skip_parentheses(node);
        let (operand, operator) = match node.kind() {
            Kind::PrefixUnaryExpression => {
                let value = node.as_prefix_unary_expression();
                (value.operand(), value.operator())
            }
            Kind::PostfixUnaryExpression => {
                let value = node.as_postfix_unary_expression();
                (value.operand(), value.operator())
            }
            Kind::BinaryExpression => {
                let value = node.as_binary_expression();
                let operator = value.operator_token().kind();
                if operator == Kind::CommaToken {
                    let left = self.modification(value.left(), counter);
                    let right = self.modification(value.right(), counter);
                    let count = left.1 + right.1;
                    return (
                        if count != 1 {
                            0
                        } else if left.1 == 1 {
                            left.0
                        } else {
                            right.0
                        },
                        count,
                    );
                }
                if !ast::is_assignment_operator(operator)
                    || !identifier_named(value.left(), counter)
                {
                    return (0, 0);
                }
                let mut direction = 0;
                if matches!(operator, Kind::PlusEqualsToken | Kind::MinusEqualsToken) {
                    if let Some(value) = self.value(value.right()) {
                        direction = if value.value == 0.0 || value.value.is_nan() {
                            0
                        } else if value.value < 0.0 {
                            -1
                        } else {
                            1
                        };
                    }
                    if operator == Kind::MinusEqualsToken {
                        direction = -direction;
                    }
                }
                return (direction, 1);
            }
            _ => return (0, 0),
        };
        if !identifier_named(operand, counter) {
            return (0, 0);
        }
        match operator {
            Kind::PlusPlusToken => (1, 1),
            Kind::MinusMinusToken => (-1, 1),
            _ => (0, 0),
        }
    }
    fn value(&mut self, node: P<Node>) -> Option<Numeric> {
        let node = ast::skip_outer_expressions(node, OuterExpressionKinds::All);
        match node.kind() {
            Kind::NumericLiteral => Some(Numeric::number(parse_numeric(node.text())?)),
            Kind::BigIntLiteral => Some(Numeric {
                value: parse_numeric(node.text().strip_suffix('n')?)?,
                kind: NumericKind::BigInt,
            }),
            Kind::TrueKeyword => Some(Numeric::boolean(true)),
            Kind::FalseKeyword => Some(Numeric::boolean(false)),
            Kind::Identifier => self.identifier(node),
            Kind::ConditionalExpression => {
                let expression = node.as_conditional_expression();
                let condition = self.value(expression.condition())?;
                self.value(if condition.truthy() {
                    expression.when_true()
                } else {
                    expression.when_false()
                })
            }
            Kind::PrefixUnaryExpression => {
                let expression = node.as_prefix_unary_expression();
                let mut value = self.value(expression.operand())?;
                match expression.operator() {
                    Kind::PlusToken => {
                        if value.kind == NumericKind::BigInt {
                            return None;
                        }
                        value.kind = NumericKind::Number;
                    }
                    Kind::MinusToken => {
                        value.value = -value.value;
                        if value.kind != NumericKind::BigInt {
                            value.kind = NumericKind::Number;
                        }
                    }
                    Kind::TildeToken => {
                        if value.kind == NumericKind::BigInt {
                            value.value = -value.value - 1.0;
                        } else {
                            value = Numeric::number((!to_uint32(value.value) as i32) as f64);
                        }
                    }
                    Kind::ExclamationToken => value = Numeric::boolean(!value.truthy()),
                    _ => return None,
                }
                Some(value)
            }
            Kind::BinaryExpression => self.binary(node),
            _ => None,
        }
    }
    fn binary(&mut self, node: P<Node>) -> Option<Numeric> {
        let binary = node.as_binary_expression();
        let op = binary.operator_token().kind();
        if matches!(op, Kind::EqualsToken | Kind::CommaToken) {
            return self.value(binary.right());
        }
        let left = self.value(binary.left())?;
        match op {
            Kind::AmpersandAmpersandToken => {
                return if left.truthy() {
                    self.value(binary.right())
                } else {
                    Some(left)
                };
            }
            Kind::BarBarToken => {
                return if left.truthy() {
                    Some(left)
                } else {
                    self.value(binary.right())
                };
            }
            Kind::QuestionQuestionToken => return Some(left),
            _ => {}
        }
        let right = self.value(binary.right())?;
        let boolean = |value| Some(Numeric::boolean(value));
        if [left, right].iter().any(|value| {
            value.kind == NumericKind::BigInt && value.value.abs() > 9_007_199_254_740_991.0
        }) {
            return match op {
                Kind::EqualsEqualsEqualsToken if left.kind != right.kind => boolean(false),
                Kind::ExclamationEqualsEqualsToken if left.kind != right.kind => boolean(true),
                _ => None,
            };
        }
        match op {
            Kind::EqualsEqualsEqualsToken => {
                return boolean(left.kind == right.kind && left.value == right.value);
            }
            Kind::ExclamationEqualsEqualsToken => {
                return boolean(left.kind != right.kind || left.value != right.value);
            }
            Kind::EqualsEqualsToken => return boolean(left.value == right.value),
            Kind::ExclamationEqualsToken => return boolean(left.value != right.value),
            _ => {}
        }
        let bigint = left.kind == NumericKind::BigInt;
        if bigint != (right.kind == NumericKind::BigInt) {
            return None;
        }
        let (l, r) = (left.value, right.value);
        let value = match op {
            Kind::PlusToken => l + r,
            Kind::MinusToken => l - r,
            Kind::AsteriskToken => l * r,
            Kind::SlashToken => {
                if bigint {
                    if r == 0.0 {
                        return None;
                    }
                    (l / r).trunc()
                } else {
                    l / r
                }
            }
            Kind::PercentToken => {
                if bigint && r == 0.0 {
                    return None;
                }
                l % r
            }
            Kind::AsteriskAsteriskToken => {
                if bigint && (r < 0.0 || r != r.trunc() || r.is_infinite()) {
                    return None;
                }
                l.powf(r)
            }
            Kind::LessThanToken => return boolean(l < r),
            Kind::LessThanEqualsToken => return boolean(l <= r),
            Kind::GreaterThanToken => return boolean(l > r),
            Kind::GreaterThanEqualsToken => return boolean(l >= r),
            _ => {
                if bigint || !l.is_finite() || !r.is_finite() {
                    return None;
                }
                let (l, r) = (to_uint32(l), to_uint32(r));
                let value = match op {
                    Kind::AmpersandToken => (l & r) as i32 as f64,
                    Kind::BarToken => (l | r) as i32 as f64,
                    Kind::CaretToken => (l ^ r) as i32 as f64,
                    Kind::LessThanLessThanToken => l.wrapping_shl(r & 31) as i32 as f64,
                    Kind::GreaterThanGreaterThanToken => (l as i32).wrapping_shr(r & 31) as f64,
                    Kind::GreaterThanGreaterThanGreaterThanToken => l.wrapping_shr(r & 31) as f64,
                    _ => return None,
                };
                return Some(Numeric::number(value));
            }
        };
        Some(Numeric {
            value,
            kind: if bigint {
                NumericKind::BigInt
            } else {
                NumericKind::Number
            },
        })
    }
    fn identifier(&mut self, node: P<Node>) -> Option<Numeric> {
        let symbol = self.checker.get_symbol_at_location_exported(node)?;
        if !self.resolving.insert(symbol) {
            return None;
        }
        let result = self.identifier_value(symbol);
        self.resolving.remove(&symbol);
        result
    }
    fn identifier_value(&mut self, symbol: P<Symbol>) -> Option<Numeric> {
        let mut declarations = symbol.declarations().iter().copied().filter(|declaration| {
            declaration.kind() == Kind::VariableDeclaration
                && declaration
                    .name()
                    .is_some_and(|name| name.kind() == Kind::Identifier)
        });
        let declaration = declarations.next()?;
        if declarations.next().is_some() {
            return None;
        }
        let list = declaration.parent()?;
        if list.kind() != Kind::VariableDeclarationList {
            return None;
        }
        if !list.flags().intersects(NodeFlags::Const)
            && has_write(self.file.as_node(), symbol, self.checker)
        {
            return None;
        }
        self.value(declaration.as_variable_declaration().initializer()?)
    }
}

fn identifier_named(node: P<Node>, name: &str) -> bool {
    let node = ast::skip_parentheses(node);
    node.kind() == Kind::Identifier && node.text() == name
}

// RefStore's effectively-constant check sees writes anywhere in the current
// source, including closures, destructuring and transparent TS wrappers.
fn has_write(node: P<Node>, symbol: P<Symbol>, checker: &mut Checker) -> bool {
    if node.kind() == Kind::Identifier
        && write_reference(node)
        && checker.resolve_name_exported(
            node.text(),
            node,
            SymbolFlags::Value | SymbolFlags::Alias,
            false,
        ) == Some(symbol)
    {
        return true;
    }
    node.for_each_child(&mut |child| has_write(child, symbol, checker))
}
pub(crate) fn write_reference(node: P<Node>) -> bool {
    let Some(parent) = node.parent() else {
        return false;
    };
    match parent.kind() {
        Kind::AsExpression
        | Kind::TypeAssertionExpression
        | Kind::SatisfiesExpression
        | Kind::NonNullExpression
        | Kind::ParenthesizedExpression => write_reference(parent),
        _ => ast::get_assignment_target(node).is_some(),
    }
}

fn parse_numeric(text: &str) -> Option<f64> {
    if text.len() > 2 && text.starts_with('0') {
        let radix = match text.as_bytes()[1] {
            b'b' | b'B' => 2,
            b'o' | b'O' => 8,
            b'x' | b'X' => 16,
            _ => 0,
        };
        if radix != 0 {
            return match u64::from_str_radix(&text[2..], radix) {
                Ok(value) => Some(value as f64),
                Err(error) if matches!(error.kind(), std::num::IntErrorKind::PosOverflow) => {
                    Some(f64::INFINITY)
                }
                Err(_) => None,
            };
        }
    }
    text.parse().ok()
}
fn to_uint32(value: f64) -> u32 {
    if value == 0.0 || !value.is_finite() {
        return 0;
    }
    let value = value.trunc() % 4_294_967_296.0;
    (if value < 0.0 {
        value + 4_294_967_296.0
    } else {
        value
    }) as u32
}
