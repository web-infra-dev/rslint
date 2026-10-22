pub(crate) mod control_flow;
pub(crate) mod go_quote;
pub(crate) mod non_null;
pub(crate) mod rsbuild_assignment_checks;
pub(crate) mod rsbuild_await_checks;
pub(crate) mod rsbuild_binding_checks;
pub(crate) mod rsbuild_core_checks;
pub(crate) mod rsbuild_cfg;
pub(crate) mod rsbuild_regex_checks;
pub(crate) mod rsbuild_regexp;
pub(crate) mod rsbuild_error_array;
pub(crate) mod rsbuild_variable_usage;
pub(crate) mod rsbuild_safety_checks;
pub(crate) mod rsbuild_rstest_promise;
pub(crate) mod rsbuild_declaration_checks;
pub(crate) mod rsbuild_expression_checks;
pub(crate) mod rsbuild_flow_checks;
pub(crate) mod rsbuild_flow_conditions;
pub(crate) mod rsbuild_promise_checks;
pub(crate) mod rsbuild_type_usage;
pub(crate) mod rsbuild_imports;
pub(crate) mod rsbuild_literals;
pub(crate) mod rsbuild_lexical_checks;
pub(crate) mod rsbuild_rstest;
pub(crate) mod rsbuild_rstest_called_once;
pub(crate) mod rsbuild_rstest_completion;
pub(crate) mod rsbuild_rstest_expect;
pub(crate) mod rsbuild_rstest_mock;
pub(crate) mod rsbuild_syntax;
pub(crate) mod rsbuild_ts_declarations;
pub(crate) mod rsbuild_ts_policy;
pub(crate) mod rsbuild_ts_syntax;
pub(crate) mod rsbuild_type_assertion;
pub(crate) mod rsbuild_type_relations;
pub(crate) mod rsbuild_type_values;
pub(crate) mod rsbuild_typed;
pub(crate) mod rsbuild_unsafe_access;
pub(crate) mod rsbuild_unsafe_values;
pub(crate) mod rsbuild_unused_vars;
pub(crate) mod structural;
pub(crate) mod syntax;
pub(crate) mod ts_structural;
pub(crate) mod typed;

use crate::{RuleConfig, Violation};
use tsrs_ast::{Kind, Node, SourceFile};
use tsrs_checker::{Checker, TypeFlags};
use tsrs_core::P;

const UNSAFE_MINUS: &str = "@typescript-eslint/no-unsafe-unary-minus";

pub(crate) fn supported(name: &str) -> bool {
    name == UNSAFE_MINUS
        || rsbuild_syntax::NAMES.contains(&name)
        || rsbuild_ts_syntax::NAMES.contains(&name)
        || rsbuild_ts_policy::NAMES.contains(&name)
        || rsbuild_ts_declarations::NAMES.contains(&name)
        || rsbuild_unsafe_access::NAMES.contains(&name)
        || rsbuild_unsafe_values::NAMES.contains(&name)
        || rsbuild_typed::NAMES.contains(&name)
        || rsbuild_type_values::NAMES.contains(&name)
        || rsbuild_literals::NAMES.contains(&name)
        || rsbuild_lexical_checks::NAMES.contains(&name)
        || rsbuild_rstest::NAMES.contains(&name)
        || rsbuild_rstest_completion::NAMES.contains(&name)
        || rsbuild_rstest_called_once::NAMES.contains(&name)
        || rsbuild_rstest_expect::NAMES.contains(&name)
        || rsbuild_rstest_mock::NAMES.contains(&name)
        || rsbuild_assignment_checks::NAMES.contains(&name)
        || rsbuild_type_relations::NAMES.contains(&name)
        || rsbuild_type_assertion::NAMES.contains(&name)
        || rsbuild_regex_checks::NAMES.contains(&name)
        || rsbuild_error_array::NAMES.contains(&name)
        || rsbuild_variable_usage::NAMES.contains(&name)
        || rsbuild_safety_checks::NAMES.contains(&name)
        || rsbuild_rstest_promise::NAMES.contains(&name)
        || rsbuild_core_checks::NAMES.contains(&name)
        || rsbuild_unused_vars::NAMES.contains(&name)
        || rsbuild_await_checks::NAMES.contains(&name)
        || rsbuild_binding_checks::NAMES.contains(&name)
        || rsbuild_expression_checks::NAMES.contains(&name)
        || rsbuild_flow_checks::NAMES.contains(&name)
        || rsbuild_flow_conditions::NAMES.contains(&name)
        || rsbuild_promise_checks::NAMES.contains(&name)
        || rsbuild_type_usage::NAMES.contains(&name)
        || rsbuild_imports::NAMES.contains(&name)
        || rsbuild_declaration_checks::NAMES.contains(&name)
        || syntax::NAMES.contains(&name)
        || control_flow::NAMES.contains(&name)
        || typed::NAMES.contains(&name)
        || structural::NAMES.contains(&name)
        || non_null::NAMES.contains(&name)
        || ts_structural::NAMES.contains(&name)
}

/// Resolve implementation selection once for an admitted file, rather than
/// rescanning every rule group's string names for every AST node. The borrowed
/// configuration preserves the caller's exact rule order, options and severity.
#[derive(Clone, Copy)]
pub(crate) struct PreparedRule<'a> {
    pub(crate) config: &'a RuleConfig,
    handler: Handler,
}

/// Listener indices are appended in configuration order, including conservative
/// all-kind listeners. Selecting a bucket does not reorder rules or AST visits.
pub(crate) struct PreparedRules<'a> {
    all: Vec<PreparedRule<'a>>,
    by_kind: Vec<Vec<usize>>,
}

impl<'a> PreparedRules<'a> {
    pub(crate) fn for_kind(&self, kind: Kind) -> impl Iterator<Item = &PreparedRule<'a>> {
        self.by_kind[kind as usize]
            .iter()
            .map(|&index| &self.all[index])
    }
}

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
enum Handler {
    RsbuildFlowChecks,
    RsbuildFlowConditions,
    RsbuildPromiseChecks,
    RsbuildTypeUsage,
    RsbuildImports,
    RsbuildSyntax,
    RsbuildTsSyntax,
    RsbuildTsPolicy,
    RsbuildTsDeclarations,
    RsbuildUnsafeAccess,
    RsbuildUnsafeValues,
    RsbuildTyped,
    RsbuildTypeValues,
    RsbuildRstestMock,
    RsbuildRstest,
    RsbuildRstestExpect,
    RsbuildRstestCompletion,
    RsbuildRstestCalledOnce,
    RsbuildLiterals,
    RsbuildLexicalChecks,
    RsbuildDeclarationChecks,
    RsbuildAssignmentChecks,
    RsbuildTypeAssertion,
    RsbuildTypeRelations,
    RsbuildExpressionChecks,
    RsbuildUnusedVars,
    RsbuildAwaitChecks,
    RsbuildBindingChecks,
    RsbuildRegexChecks,
    RsbuildErrorArray,
    RsbuildVariableUsage,
    RsbuildSafetyChecks,
    RsbuildRstestPromise,
    RsbuildCoreChecks,
    TsStructural,
    Structural,
    NonNull,
    ControlFlow,
    Typed,
    Syntax,
    UnsafeMinus,
    NoOp,
}

/// The original no-useless-backreference rule registers no listeners when its
/// source has no backslash. Every reportable literal or folded constructor
/// pattern needs one in the source. Evaluate that file-wide condition once,
/// not on each AST node; keep the configured rule in the result either way.
pub(crate) fn prepare<'a>(
    configs: &'a [RuleConfig],
    text: &str,
) -> Result<PreparedRules<'a>, String> {
    let can_have_backreferences = configs
        .iter()
        .any(|rule| rule.name == "no-useless-backreference")
        && text.contains('\\');
    let all = configs
        .iter()
        .map(|config| PreparedRule::new(config, can_have_backreferences))
        .collect::<Result<Vec<_>, _>>()?;
    let mut by_kind = vec![Vec::new(); Kind::Count as usize + 1];
    for (index, rule) in all.iter().enumerate() {
        match rule.listener_kinds() {
            Some(kinds) => {
                for &kind in kinds {
                    by_kind[kind as usize].push(index);
                }
            }
            None => {
                for listeners in &mut by_kind {
                    listeners.push(index);
                }
            }
        }
    }
    Ok(PreparedRules { all, by_kind })
}

impl<'a> PreparedRule<'a> {
    fn new(config: &'a RuleConfig, can_have_backreferences: bool) -> Result<Self, String> {
        let handler = Handler::for_name(&config.name)
            .ok_or_else(|| format!("rule has not been migrated to Rust: {}", config.name))?;
        let handler = if config.name == "no-useless-backreference" && !can_have_backreferences {
            Handler::NoOp
        } else {
            handler
        };
        Ok(Self { config, handler })
    }

    /// These are the entry guards of each implementation, not guesses based on
    /// which nodes happened to report in Rsbuild. A file-level rule still runs
    /// its complete internal traversal. New handlers default to every kind
    /// until their guards have been audited, so omitted metadata cannot disable
    /// a rule. Keep this table aligned with the corresponding `check` guards.
    fn listener_kinds(&self) -> Option<&'static [Kind]> {
        use Kind::*;
        match self.handler {
            Handler::NoOp => return Some(&[]),
            Handler::RsbuildImports
            | Handler::RsbuildRstest
            | Handler::RsbuildRstestExpect
            | Handler::RsbuildRstestCompletion
            | Handler::RsbuildRstestCalledOnce
            | Handler::RsbuildRstestMock
            | Handler::RsbuildUnusedVars
            | Handler::RsbuildVariableUsage
            | Handler::RsbuildRstestPromise => return Some(&[SourceFile]),
            Handler::NonNull => return Some(&[NonNullExpression]),
            Handler::RsbuildTypeRelations => return Some(&[UnionType, IntersectionType]),
            Handler::RsbuildTypeAssertion => {
                return Some(&[AsExpression, TypeAssertionExpression, NonNullExpression]);
            }
            _ => {}
        }
        Some(match self.config.name.as_str() {
            "no-unreachable"
            | "no-obj-calls"
            | "prefer-const"
            | "no-var"
            | "no-class-assign"
            | "no-func-assign"
            | "no-ex-assign"
            | "no-import-assign"
            | "no-extra-boolean-cast"
            | "use-isnan"
            | "no-irregular-whitespace"
            | "no-unused-labels"
            | "@typescript-eslint/no-unsafe-declaration-merging"
            | "@typescript-eslint/ban-ts-comment"
            | "@typescript-eslint/triple-slash-reference"
            | "@typescript-eslint/require-await"
            | "@typescript-eslint/no-misused-promises" => &[SourceFile],
            "no-debugger" => &[DebuggerStatement],
            "no-with" => &[WithStatement],
            "no-ternary" | "no-nested-ternary" => &[ConditionalExpression],
            "no-template-curly-in-string" | "no-nonoctal-decimal-escape" => &[StringLiteral],
            "no-caller" => &[PropertyAccessExpression, QualifiedName],
            "no-compare-neg-zero" | "@typescript-eslint/no-unsafe-unary-minus" => {
                &[PrefixUnaryExpression]
            }
            "no-case-declarations" | "no-duplicate-case" => &[CaseBlock],
            "no-empty-static-block" => &[ClassStaticBlockDeclaration],
            "no-empty-pattern" => &[ObjectBindingPattern, ArrayBindingPattern],
            "require-yield" => &[FunctionDeclaration, FunctionExpression, MethodDeclaration],
            "constructor-super" | "no-this-before-super" => &[Constructor],
            "getter-return" => &[GetAccessor, ObjectLiteralExpression],
            "no-fallthrough" => &[SwitchStatement],
            "no-dupe-else-if" => &[IfStatement],
            "no-constant-binary-expression"
            | "no-self-assign"
            | "no-unsafe-negation"
            | "@typescript-eslint/restrict-plus-operands" => &[BinaryExpression],
            "no-constant-condition" | "no-cond-assign" => &[
                IfStatement,
                WhileStatement,
                DoStatement,
                ForStatement,
                ConditionalExpression,
            ],
            "@typescript-eslint/no-floating-promises"
            | "@typescript-eslint/no-unused-expressions" => &[ExpressionStatement],
            "@typescript-eslint/no-unnecessary-type-constraint" => &[TypeParameter],
            "@typescript-eslint/prefer-as-const" => &[
                VariableDeclaration,
                PropertyDeclaration,
                AsExpression,
                TypeAssertionExpression,
            ],
            "@typescript-eslint/prefer-namespace-keyword" | "@typescript-eslint/no-namespace" => {
                &[ModuleDeclaration]
            }
            "@typescript-eslint/no-wrapper-object-types"
            | "@typescript-eslint/no-unsafe-function-type" => &[TypeReference],
            "@typescript-eslint/no-require-imports" => &[CallExpression, ExternalModuleReference],
            "@typescript-eslint/no-empty-object-type" => &[TypeLiteral, InterfaceDeclaration],
            "@typescript-eslint/no-this-alias" => &[ThisKeyword],
            "@typescript-eslint/no-array-constructor" | "@typescript-eslint/no-implied-eval" => {
                &[CallExpression, NewExpression]
            }
            "@typescript-eslint/no-unsafe-call" | "@typescript-eslint/no-unsafe-argument" => {
                &[CallExpression, NewExpression, TaggedTemplateExpression]
            }
            "@typescript-eslint/no-unsafe-member-access" => {
                &[PropertyAccessExpression, ElementAccessExpression]
            }
            "@typescript-eslint/no-unsafe-return" | "no-setter-return" => {
                &[ArrowFunction, ReturnStatement]
            }
            "@typescript-eslint/no-for-in-array" => &[ForInStatement],
            "@typescript-eslint/no-array-delete" | "no-delete-var" => &[DeleteExpression],
            "@typescript-eslint/only-throw-error" | "preserve-caught-error" => &[ThrowStatement],
            "@typescript-eslint/restrict-template-expressions" => &[TemplateExpression],
            "@typescript-eslint/prefer-promise-reject-errors"
            | "no-prototype-builtins"
            | "prefer-spread" => &[CallExpression],
            "no-octal" | "no-loss-of-precision" => &[NumericLiteral],
            "no-empty-character-class" => &[RegularExpressionLiteral],
            "no-unexpected-multiline" => &[
                ElementAccessExpression,
                CallExpression,
                TaggedTemplateExpression,
                BinaryExpression,
            ],
            "no-dupe-keys" => &[ObjectLiteralExpression],
            "no-dupe-class-members" => &[ClassDeclaration, ClassExpression],
            "no-shadow-restricted-names" => &[
                VariableDeclaration,
                Parameter,
                CatchClause,
                FunctionDeclaration,
                FunctionExpression,
                ClassDeclaration,
                ClassExpression,
                ImportDeclaration,
            ],
            "no-global-assign" | "no-const-assign" | "prefer-rest-params" => &[Identifier],
            "no-new-native-nonconstructor" | "no-async-promise-executor" => &[NewExpression],
            "no-empty" => &[Block, SwitchStatement],
            "valid-typeof" => &[TypeOfExpression],
            "for-direction" => &[ForStatement],
            "no-await-in-loop" => &[AwaitExpression, ForOfStatement],
            "no-constructor-return" => &[ReturnStatement],
            "no-continue" => &[ContinueStatement],
            "no-negated-condition" => &[IfStatement, ConditionalExpression],
            "no-unsafe-finally" => &[
                ReturnStatement,
                ThrowStatement,
                BreakStatement,
                ContinueStatement,
            ],
            "no-useless-catch" => &[CatchClause],
            "no-dupe-args" => &[FunctionDeclaration, FunctionExpression],
            "no-sparse-arrays" => &[ArrayLiteralExpression],
            "@typescript-eslint/no-mixed-enums" | "@typescript-eslint/no-duplicate-enum-values" => {
                &[EnumDeclaration]
            }
            "@typescript-eslint/no-unsafe-enum-comparison" => &[BinaryExpression, CaseClause],
            "@typescript-eslint/related-getter-setter-pairs" => &[
                ClassDeclaration,
                ClassExpression,
                InterfaceDeclaration,
                TypeLiteral,
            ],
            "@typescript-eslint/no-misused-new" => {
                &[MethodDeclaration, ConstructSignature, MethodSignature]
            }
            "@typescript-eslint/await-thenable" => &[
                AwaitExpression,
                ForOfStatement,
                VariableDeclarationList,
                CallExpression,
            ],
            "unicorn/prefer-array-some" => &[CallExpression, BinaryExpression],
            "no-invalid-regexp" => &[CallExpression, NewExpression],
            "no-control-regex"
            | "no-regex-spaces"
            | "no-misleading-character-class"
            | "no-useless-backreference" => {
                &[RegularExpressionLiteral, CallExpression, NewExpression]
            }
            "no-useless-escape" => &[
                StringLiteral,
                NoSubstitutionTemplateLiteral,
                TemplateHead,
                TemplateMiddle,
                TemplateTail,
                RegularExpressionLiteral,
            ],
            "@typescript-eslint/no-base-to-string" => {
                &[BinaryExpression, CallExpression, TemplateExpression]
            }
            "@typescript-eslint/unbound-method" => &[
                PropertyAccessExpression,
                ElementAccessExpression,
                QualifiedName,
                ObjectLiteralExpression,
                ObjectBindingPattern,
            ],
            "no-unsafe-optional-chaining" => &[
                CallExpression,
                NewExpression,
                TaggedTemplateExpression,
                PropertyAccessExpression,
                ElementAccessExpression,
                BinaryExpression,
                SpreadElement,
                ForOfStatement,
                VariableDeclaration,
                BindingElement,
                WithStatement,
                ClassDeclaration,
                ClassExpression,
            ],
            _ => return None,
        })
    }

    pub(crate) fn check(
        &self,
        node: P<Node>,
        file: P<SourceFile>,
        checker: &mut Checker,
        default_library: &dyn Fn(P<SourceFile>) -> bool,
        has_type_information: bool,
        program: &tsrs_compiler::Program,
    ) -> Vec<Violation> {
        let name = self.config.name.as_str();
        let options = &self.config.options;
        match self.handler {
            Handler::RsbuildFlowChecks => rsbuild_flow_checks::check(name, node, file, checker),
            Handler::RsbuildFlowConditions => rsbuild_flow_conditions::check(
                name,
                node,
                file,
                checker,
                default_library,
                has_type_information,
            ),
            Handler::RsbuildPromiseChecks => {
                rsbuild_promise_checks::check(name, node, file, checker, default_library)
            }
            Handler::RsbuildTypeUsage => {
                rsbuild_type_usage::check(name, node, file, checker, default_library)
            }
            Handler::RsbuildImports => rsbuild_imports::check(name, node, file, program),
            Handler::RsbuildSyntax => rsbuild_syntax::check(name, node, file, checker),
            Handler::RsbuildTsSyntax => rsbuild_ts_syntax::check(name, node, file, checker),
            Handler::RsbuildTsPolicy => rsbuild_ts_policy::check(name, node, file, checker),
            Handler::RsbuildTsDeclarations => {
                rsbuild_ts_declarations::check(name, node, file, checker, has_type_information)
            }
            Handler::RsbuildUnsafeAccess => {
                rsbuild_unsafe_access::check(name, node, file, checker, default_library)
            }
            Handler::RsbuildUnsafeValues => {
                rsbuild_unsafe_values::check(name, node, file, checker, default_library)
            }
            Handler::RsbuildTyped => {
                rsbuild_typed::check(name, node, file, checker, default_library)
            }
            Handler::RsbuildTypeValues => {
                rsbuild_type_values::check(name, node, file, checker, default_library)
            }
            Handler::RsbuildRstestMock => rsbuild_rstest_mock::check(
                name,
                node,
                file,
                checker,
                default_library,
                has_type_information,
            ),
            Handler::RsbuildRstest => {
                rsbuild_rstest::check(name, node, file, checker, has_type_information)
            }
            Handler::RsbuildRstestExpect => {
                rsbuild_rstest_expect::check(name, node, file, checker, has_type_information)
            }
            Handler::RsbuildRstestCompletion => {
                rsbuild_rstest_completion::check(name, node, file, checker, has_type_information)
            }
            Handler::RsbuildRstestCalledOnce => rsbuild_rstest_called_once::check(
                name,
                node,
                file,
                checker,
                default_library,
                has_type_information,
            ),
            Handler::RsbuildLiterals => rsbuild_literals::check(name, node, file, checker),
            Handler::RsbuildLexicalChecks => {
                rsbuild_lexical_checks::check(name, node, file, checker)
            }
            Handler::RsbuildDeclarationChecks => {
                rsbuild_declaration_checks::check(name, node, file, checker)
            }
            Handler::RsbuildAssignmentChecks => {
                rsbuild_assignment_checks::check(name, node, file, checker)
            }
            Handler::RsbuildTypeAssertion => {
                rsbuild_type_assertion::check(name, node, file, checker)
            }
            Handler::RsbuildTypeRelations => {
                rsbuild_type_relations::check(name, node, file, checker)
            }
            Handler::RsbuildExpressionChecks => {
                rsbuild_expression_checks::check(name, node, file, checker)
            }
            Handler::RsbuildUnusedVars => {
                rsbuild_unused_vars::check(name, node, file, checker, options)
            }
            Handler::RsbuildAwaitChecks => {
                rsbuild_await_checks::check(name, node, file, checker, default_library)
            }
            Handler::RsbuildBindingChecks => {
                rsbuild_binding_checks::check(name, node, file, checker)
            }
            Handler::RsbuildRegexChecks => rsbuild_regex_checks::check(
                name,
                node,
                file,
                checker,
                default_library,
                has_type_information,
            ),
            Handler::RsbuildErrorArray => {
                rsbuild_error_array::check(name, node, file, checker, has_type_information)
            }
            Handler::RsbuildVariableUsage => {
                rsbuild_variable_usage::check(name, node, file, checker)
            }
            Handler::RsbuildSafetyChecks => rsbuild_safety_checks::check(name, node, file),
            Handler::RsbuildRstestPromise => {
                rsbuild_rstest_promise::check(name, node, file, checker)
            }
            Handler::RsbuildCoreChecks => rsbuild_core_checks::check(name, node, file, checker),
            Handler::TsStructural => ts_structural::check(name, node, checker),
            Handler::Structural => structural::check(name, node, file),
            Handler::NonNull => non_null::check(name, node),
            Handler::ControlFlow => control_flow::check(name, node, file),
            Handler::Typed => typed::check(name, node, checker),
            Handler::Syntax => syntax::check(name, node, file),
            Handler::UnsafeMinus => check_unsafe_minus(node, checker),
            Handler::NoOp => Vec::new(),
        }
    }
}

impl Handler {
    fn for_name(name: &str) -> Option<Self> {
        if rsbuild_flow_checks::NAMES.contains(&name) {
            return Some(Self::RsbuildFlowChecks);
        }
        if rsbuild_flow_conditions::NAMES.contains(&name) {
            return Some(Self::RsbuildFlowConditions);
        }
        if rsbuild_promise_checks::NAMES.contains(&name) {
            return Some(Self::RsbuildPromiseChecks);
        }
        if rsbuild_type_usage::NAMES.contains(&name) {
            return Some(Self::RsbuildTypeUsage);
        }
        if rsbuild_imports::NAMES.contains(&name) {
            return Some(Self::RsbuildImports);
        }
        if rsbuild_syntax::NAMES.contains(&name) {
            return Some(Self::RsbuildSyntax);
        }
        if rsbuild_ts_syntax::NAMES.contains(&name) {
            return Some(Self::RsbuildTsSyntax);
        }
        if rsbuild_ts_policy::NAMES.contains(&name) {
            return Some(Self::RsbuildTsPolicy);
        }
        if rsbuild_ts_declarations::NAMES.contains(&name) {
            return Some(Self::RsbuildTsDeclarations);
        }
        if rsbuild_unsafe_access::NAMES.contains(&name) {
            return Some(Self::RsbuildUnsafeAccess);
        }
        if rsbuild_unsafe_values::NAMES.contains(&name) {
            return Some(Self::RsbuildUnsafeValues);
        }
        if rsbuild_typed::NAMES.contains(&name) {
            return Some(Self::RsbuildTyped);
        }
        if rsbuild_type_values::NAMES.contains(&name) {
            return Some(Self::RsbuildTypeValues);
        }
        if rsbuild_rstest_mock::NAMES.contains(&name) {
            return Some(Self::RsbuildRstestMock);
        }
        if rsbuild_rstest::NAMES.contains(&name) {
            return Some(Self::RsbuildRstest);
        }
        if rsbuild_rstest_expect::NAMES.contains(&name) {
            return Some(Self::RsbuildRstestExpect);
        }
        if rsbuild_rstest_completion::NAMES.contains(&name) {
            return Some(Self::RsbuildRstestCompletion);
        }
        if rsbuild_rstest_called_once::NAMES.contains(&name) {
            return Some(Self::RsbuildRstestCalledOnce);
        }
        if rsbuild_literals::NAMES.contains(&name) {
            return Some(Self::RsbuildLiterals);
        }
        if rsbuild_lexical_checks::NAMES.contains(&name) {
            return Some(Self::RsbuildLexicalChecks);
        }
        if rsbuild_declaration_checks::NAMES.contains(&name) {
            return Some(Self::RsbuildDeclarationChecks);
        }
        if rsbuild_assignment_checks::NAMES.contains(&name) {
            return Some(Self::RsbuildAssignmentChecks);
        }
        if rsbuild_type_assertion::NAMES.contains(&name) {
            return Some(Self::RsbuildTypeAssertion);
        }
        if rsbuild_type_relations::NAMES.contains(&name) {
            return Some(Self::RsbuildTypeRelations);
        }
        if rsbuild_expression_checks::NAMES.contains(&name) {
            return Some(Self::RsbuildExpressionChecks);
        }
        if rsbuild_unused_vars::NAMES.contains(&name) {
            return Some(Self::RsbuildUnusedVars);
        }
        if rsbuild_await_checks::NAMES.contains(&name) {
            return Some(Self::RsbuildAwaitChecks);
        }
        if rsbuild_binding_checks::NAMES.contains(&name) {
            return Some(Self::RsbuildBindingChecks);
        }
        if rsbuild_regex_checks::NAMES.contains(&name) {
            return Some(Self::RsbuildRegexChecks);
        }
        if rsbuild_error_array::NAMES.contains(&name) {
            return Some(Self::RsbuildErrorArray);
        }
        if rsbuild_variable_usage::NAMES.contains(&name) {
            return Some(Self::RsbuildVariableUsage);
        }
        if rsbuild_safety_checks::NAMES.contains(&name) {
            return Some(Self::RsbuildSafetyChecks);
        }
        if rsbuild_rstest_promise::NAMES.contains(&name) {
            return Some(Self::RsbuildRstestPromise);
        }
        if rsbuild_core_checks::NAMES.contains(&name) {
            return Some(Self::RsbuildCoreChecks);
        }
        if ts_structural::NAMES.contains(&name) {
            return Some(Self::TsStructural);
        }
        if structural::NAMES.contains(&name) {
            return Some(Self::Structural);
        }
        if non_null::NAMES.contains(&name) {
            return Some(Self::NonNull);
        }
        if control_flow::NAMES.contains(&name) {
            return Some(Self::ControlFlow);
        }
        if typed::NAMES.contains(&name) {
            return Some(Self::Typed);
        }
        if name == UNSAFE_MINUS {
            Some(Self::UnsafeMinus)
        } else if syntax::NAMES.contains(&name) {
            Some(Self::Syntax)
        } else {
            None
        }
    }
}

fn check_unsafe_minus(node: P<Node>, checker: &mut Checker) -> Vec<Violation> {
    if node.kind() != Kind::PrefixUnaryExpression {
        return Vec::new();
    }
    let expression = node.as_prefix_unary_expression();
    if expression.operator() != Kind::MinusToken {
        return Vec::new();
    }
    let value = checker.get_type_at_location(expression.operand());
    let value = checker.get_base_constraint_of_type(value).unwrap_or(value);
    let parts = if value.flags().intersects(TypeFlags::Union) {
        value.types().to_vec()
    } else {
        vec![value]
    };
    for part in parts {
        if !part.flags().intersects(
            TypeFlags::Any | TypeFlags::Never | TypeFlags::BigIntLike | TypeFlags::NumberLike,
        ) {
            return vec![Violation::new(
                node,
                "unaryMinus",
                format!(
                    "Argument of unary negation should be assignable to number | bigint but is {} instead.",
                    checker.type_to_string_exported(part),
                ),
            )];
        }
    }
    Vec::new()
}

/// Same type capability declared by each rule in the frozen original engine.
/// This is rule metadata; project root selection determines the capability.
pub(crate) fn requires_type_information(name: &str) -> bool {
    typed::NAMES.contains(&name)
        || matches!(
            name,
            "@typescript-eslint/await-thenable"
                | "@typescript-eslint/no-array-delete"
                | "@typescript-eslint/no-base-to-string"
                | "@typescript-eslint/no-duplicate-type-constituents"
                | "@typescript-eslint/no-floating-promises"
                | "@typescript-eslint/no-for-in-array"
                | "@typescript-eslint/no-implied-eval"
                | "@typescript-eslint/no-misused-promises"
                | "@typescript-eslint/no-redundant-type-constituents"
                | "@typescript-eslint/no-unnecessary-type-assertion"
                | "@typescript-eslint/no-unsafe-argument"
                | "@typescript-eslint/no-unsafe-call"
                | "@typescript-eslint/no-unsafe-enum-comparison"
                | "@typescript-eslint/no-unsafe-function-type"
                | "@typescript-eslint/no-unsafe-member-access"
                | "@typescript-eslint/no-unsafe-return"
                | "@typescript-eslint/no-unsafe-unary-minus"
                | "@typescript-eslint/no-unused-vars"
                | "@typescript-eslint/no-wrapper-object-types"
                | "@typescript-eslint/only-throw-error"
                | "@typescript-eslint/prefer-promise-reject-errors"
                | "@typescript-eslint/require-await"
                | "@typescript-eslint/restrict-plus-operands"
                | "@typescript-eslint/restrict-template-expressions"
                | "@typescript-eslint/unbound-method"
                | "rstest/require-local-test-context-for-concurrent-snapshots"
        )
}

pub(crate) fn validate_options(name: &str, options: &[serde_json::Value]) -> Result<(), String> {
    if rsbuild_unused_vars::NAMES.contains(&name) {
        return rsbuild_unused_vars::validate_options(options);
    }
    if options.is_empty() {
        Ok(())
    } else {
        Err(format!("{name} does not accept options"))
    }
}

#[cfg(test)]
#[test]
fn prepared_dispatch_covers_every_declared_rule_without_ambiguous_groups() {
    let groups: &[(Handler, &[&str])] = &[
        (Handler::RsbuildFlowChecks, rsbuild_flow_checks::NAMES),
        (
            Handler::RsbuildFlowConditions,
            rsbuild_flow_conditions::NAMES,
        ),
        (Handler::RsbuildPromiseChecks, rsbuild_promise_checks::NAMES),
        (Handler::RsbuildTypeUsage, rsbuild_type_usage::NAMES),
        (Handler::RsbuildImports, rsbuild_imports::NAMES),
        (Handler::RsbuildSyntax, rsbuild_syntax::NAMES),
        (Handler::RsbuildTsSyntax, rsbuild_ts_syntax::NAMES),
        (Handler::RsbuildTsPolicy, rsbuild_ts_policy::NAMES),
        (
            Handler::RsbuildTsDeclarations,
            rsbuild_ts_declarations::NAMES,
        ),
        (Handler::RsbuildUnsafeAccess, rsbuild_unsafe_access::NAMES),
        (Handler::RsbuildUnsafeValues, rsbuild_unsafe_values::NAMES),
        (Handler::RsbuildTyped, rsbuild_typed::NAMES),
        (Handler::RsbuildTypeValues, rsbuild_type_values::NAMES),
        (Handler::RsbuildRstestMock, rsbuild_rstest_mock::NAMES),
        (Handler::RsbuildRstest, rsbuild_rstest::NAMES),
        (Handler::RsbuildRstestExpect, rsbuild_rstest_expect::NAMES),
        (
            Handler::RsbuildRstestCompletion,
            rsbuild_rstest_completion::NAMES,
        ),
        (
            Handler::RsbuildRstestCalledOnce,
            rsbuild_rstest_called_once::NAMES,
        ),
        (Handler::RsbuildLiterals, rsbuild_literals::NAMES),
        (Handler::RsbuildLexicalChecks, rsbuild_lexical_checks::NAMES),
        (
            Handler::RsbuildDeclarationChecks,
            rsbuild_declaration_checks::NAMES,
        ),
        (
            Handler::RsbuildAssignmentChecks,
            rsbuild_assignment_checks::NAMES,
        ),
        (Handler::RsbuildTypeAssertion, rsbuild_type_assertion::NAMES),
        (Handler::RsbuildTypeRelations, rsbuild_type_relations::NAMES),
        (
            Handler::RsbuildExpressionChecks,
            rsbuild_expression_checks::NAMES,
        ),
        (Handler::RsbuildUnusedVars, rsbuild_unused_vars::NAMES),
        (Handler::RsbuildAwaitChecks, rsbuild_await_checks::NAMES),
        (Handler::RsbuildBindingChecks, rsbuild_binding_checks::NAMES),
        (Handler::RsbuildRegexChecks, rsbuild_regex_checks::NAMES),
        (Handler::RsbuildErrorArray, rsbuild_error_array::NAMES),
        (Handler::RsbuildVariableUsage, rsbuild_variable_usage::NAMES),
        (Handler::RsbuildSafetyChecks, rsbuild_safety_checks::NAMES),
        (Handler::RsbuildRstestPromise, rsbuild_rstest_promise::NAMES),
        (Handler::RsbuildCoreChecks, rsbuild_core_checks::NAMES),
        (Handler::TsStructural, ts_structural::NAMES),
        (Handler::Structural, structural::NAMES),
        (Handler::NonNull, non_null::NAMES),
        (Handler::ControlFlow, control_flow::NAMES),
        (Handler::Typed, typed::NAMES),
        (Handler::Syntax, syntax::NAMES),
        (Handler::UnsafeMinus, &[UNSAFE_MINUS]),
    ];
    let mut names = std::collections::BTreeSet::new();
    for &(expected, group) in groups {
        assert!(!group.is_empty());
        for &name in group {
            assert!(
                names.insert(name),
                "rule has ambiguous implementations: {name}"
            );
            assert!(
                supported(name),
                "prepared rule missing from registry: {name}"
            );
            let config = RuleConfig {
                name: name.into(),
                severity: 1,
                options: vec![],
            };
            let prepared = PreparedRule::new(&config, true).unwrap();
            assert_eq!(
                prepared.handler, expected,
                "wrong implementation for {name}"
            );
            assert!(std::ptr::eq(prepared.config, &config));
        }
    }
    assert_eq!(
        names.len(),
        139,
        "update this bound when the registry changes"
    );
    let unknown = RuleConfig {
        name: "not-a-migrated-rule".into(),
        severity: 2,
        options: vec![],
    };
    assert!(!supported(&unknown.name));
    assert!(PreparedRule::new(&unknown, true).is_err());
}

#[cfg(test)]
#[test]
fn prepared_backreference_guard_keeps_configuration_and_other_rule_handlers() {
    let configs = [
        RuleConfig {
            name: "no-useless-backreference".into(),
            severity: 1,
            options: vec![],
        },
        RuleConfig {
            name: "no-debugger".into(),
            severity: 2,
            options: vec![],
        },
    ];
    let clean = prepare(&configs, "const value = /(value)/; debugger;").unwrap();
    assert_eq!(clean.all.len(), configs.len());
    assert_eq!(clean.all[0].handler, Handler::NoOp);
    assert_eq!(clean.all[1].handler, Handler::Syntax);
    assert!(std::ptr::eq(clean.all[0].config, &configs[0]));
    assert!(std::ptr::eq(clean.all[1].config, &configs[1]));
    assert!(
        clean
            .for_kind(Kind::RegularExpressionLiteral)
            .next()
            .is_none()
    );
    let backreference = prepare(&configs, r"const value = /(value)\1/; debugger;").unwrap();
    assert_eq!(backreference.all[0].handler, Handler::RsbuildRegexChecks);
    assert_eq!(backreference.all[1].handler, Handler::Syntax);
    assert_eq!(
        backreference
            .for_kind(Kind::RegularExpressionLiteral)
            .map(|rule| rule.config.name.as_str())
            .collect::<Vec<_>>(),
        ["no-useless-backreference"]
    );
}

#[cfg(test)]
#[test]
fn prepared_listener_buckets_preserve_rule_order_and_borrowed_options() {
    let configs: Vec<RuleConfig> = serde_json::from_value(serde_json::json!([
        {"name":"no-var", "severity":1},
        {"name":"no-debugger", "severity":2},
        {"name":"rstest/no-focused-tests", "severity":1},
        {"name":"@typescript-eslint/no-unused-vars", "severity":2,
         "options":[{"argsIgnorePattern":"^_", "varsIgnorePattern":"^_"}]},
        {"name":"no-extra-boolean-cast", "severity":2},
        {"name":"no-nested-ternary", "severity":1},
        {"name":"no-ternary", "severity":2}
    ]))
    .unwrap();
    let prepared = prepare(&configs, "").unwrap();
    let names = |kind| {
        prepared
            .for_kind(kind)
            .map(|rule| rule.config.name.as_str())
            .collect::<Vec<_>>()
    };
    assert_eq!(
        names(Kind::SourceFile),
        [
            "no-var",
            "rstest/no-focused-tests",
            "@typescript-eslint/no-unused-vars",
            "no-extra-boolean-cast"
        ]
    );
    assert_eq!(names(Kind::DebuggerStatement), ["no-debugger"]);
    assert_eq!(
        names(Kind::ConditionalExpression),
        ["no-nested-ternary", "no-ternary"]
    );
    assert!(names(Kind::Identifier).is_empty());
    let unused = prepared.for_kind(Kind::SourceFile).nth(2).unwrap();
    assert!(std::ptr::eq(unused.config, &configs[3]));
    assert_eq!(
        unused.config.options[0]["argsIgnorePattern"],
        serde_json::json!("^_")
    );
}

#[cfg(test)]
#[test]
fn prepared_dispatch_preserves_public_diagnostics_for_file_and_node_rules() {
    let request = serde_json::json!({
        "schemaVersion":1,
        "cwd":"/rslint-prepared-dispatch",
        "files":[
            {"path":"hit.ts", "text":"var value = cond ? left ? 1 : 2 : 3; debugger;"},
            {"path":"clean.ts", "text":"export const value = 1;"}
        ],
        "rules":[
            {"name":"no-var", "severity":1},
            {"name":"no-ternary", "severity":2},
            {"name":"no-nested-ternary", "severity":1},
            {"name":"no-debugger", "severity":2}
        ]
    });
    let result: serde_json::Value =
        serde_json::from_str(&crate::lint_json(&request.to_string()).unwrap()).unwrap();
    let messages = result["files"][0]["messages"].as_array().unwrap();
    assert_eq!(
        messages
            .iter()
            .map(|message| message["ruleId"].as_str().unwrap())
            .collect::<Vec<_>>(),
        [
            "no-var",
            "no-ternary",
            "no-nested-ternary",
            "no-ternary",
            "no-debugger"
        ]
    );
    assert_eq!(messages[0]["severity"], 1);
    assert_eq!(messages[1]["severity"], 2);
    assert_eq!(messages[2]["severity"], 1);
    assert_eq!(messages[1]["range"], messages[2]["range"]);
    assert!(messages[0].get("fix").is_some());
    assert_eq!(result["files"][1]["messages"], serde_json::json!([]));
    assert_eq!(
        result["files"][1]["appliedRules"],
        serde_json::json!(["no-var", "no-ternary", "no-nested-ternary", "no-debugger"])
    );
}

#[cfg(test)]
#[test]
fn prepared_dispatch_checks_promise_aggregators_without_await_nodes() {
    let request = serde_json::json!({
        "schemaVersion":1,
        "cwd":"/rslint-prepared-dispatch-aggregators",
        "files":[
            {"path":"hit.ts", "text":concat!(
                "Promise.all([Promise.resolve(1), 42]);\n",
                "Promise.allSettled([Promise.resolve(1), 42]);\n",
                "Promise.any([Promise.resolve(1), 42]);\n",
                "Promise.race([Promise.resolve(1), 42]);\nexport {};"
            )},
            {"path":"clean.ts", "text":concat!(
                "Promise.all([Promise.resolve(1), Promise.resolve(42)]);\n",
                "Promise.allSettled([Promise.resolve(1), Promise.resolve(42)]);\n",
                "Promise.any([Promise.resolve(1), Promise.resolve(42)]);\n",
                "Promise.race([Promise.resolve(1), Promise.resolve(42)]);\nexport {};"
            )}
        ],
        "rules":[{"name":"@typescript-eslint/await-thenable"}]
    });
    let result: serde_json::Value =
        serde_json::from_str(&crate::lint_json(&request.to_string()).unwrap()).unwrap();
    let messages = result["files"][0]["messages"].as_array().unwrap();
    assert_eq!(messages.len(), 4, "{result}");
    for (index, message) in messages.iter().enumerate() {
        assert_eq!(message["ruleId"], "@typescript-eslint/await-thenable");
        assert_eq!(message["messageId"], "invalidPromiseAggregatorInput");
        assert_eq!(message["line"], index + 1);
        assert_eq!(
            message["endColumn"].as_u64().unwrap() - message["column"].as_u64().unwrap(),
            2
        );
        assert!(message.get("fix").is_none());
    }
    assert_eq!(result["files"][1]["messages"], serde_json::json!([]));
    for file in result["files"].as_array().unwrap() {
        assert_eq!(
            file["appliedRules"],
            serde_json::json!(["@typescript-eslint/await-thenable"])
        );
    }
}

#[cfg(test)]
#[test]
fn prepared_dispatch_checks_division_that_looks_like_a_newline_regexp() {
    let request = serde_json::json!({
        "schemaVersion":1,
        "cwd":"/rslint-prepared-dispatch-multiline",
        "files":[
            {"path":"hit.ts", "text":"foo\n/ bar /g;\nexport {};"},
            {"path":"clean.ts", "text":"foo / bar /g;\nexport {};"}
        ],
        "rules":[{"name":"no-unexpected-multiline"}]
    });
    let result: serde_json::Value =
        serde_json::from_str(&crate::lint_json(&request.to_string()).unwrap()).unwrap();
    let messages = result["files"][0]["messages"].as_array().unwrap();
    assert_eq!(messages.len(), 1, "{result}");
    assert_eq!(messages[0]["ruleId"], "no-unexpected-multiline");
    assert_eq!(messages[0]["messageId"], "division");
    assert_eq!(messages[0]["line"], 2);
    assert_eq!(messages[0]["column"], 1);
    assert_eq!(messages[0]["range"], serde_json::json!([4, 5]));
    assert_eq!(result["files"][1]["messages"], serde_json::json!([]));
}
