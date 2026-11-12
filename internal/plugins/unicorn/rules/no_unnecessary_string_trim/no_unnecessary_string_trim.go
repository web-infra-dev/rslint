package no_unnecessary_string_trim

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

const messageID = "no-unnecessary-string-trim"

var (
	trimStartMessage = rule.RuleMessage{
		Id:          messageID,
		Description: "Prefer `String#trimStart()` before `String#startsWith()`.",
		Data:        map[string]string{"method": "startsWith", "replacement": "trimStart"},
	}
	trimEndMessage = rule.RuleMessage{
		Id:          messageID,
		Description: "Prefer `String#trimEnd()` before `String#endsWith()`.",
		Data:        map[string]string{"method": "endsWith", "replacement": "trimEnd"},
	}
)

// NoUnnecessaryStringTrimRule prefers trimming only the end that a prefix or
// suffix check needs.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-unnecessary-string-trim.js
var NoUnnecessaryStringTrimRule = rule.Rule{
	Name:   "unicorn/" + messageID,
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		zero, one := 0, 1
		searchOptions := unicornutil.DotMethodCallOptions{
			Methods: []string{"startsWith", "endsWith"}, MaximumArguments: &one, RejectSpreadElement: true,
		}
		trimOptions := unicornutil.DotMethodCallOptions{Method: "trim", ArgumentsLength: &zero}
		var evaluator *utils.StaticStringEvaluator
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				search, ok := unicornutil.MatchDotMethodCall(node, searchOptions)
				if !ok {
					return
				}
				trim, ok := unicornutil.MatchDotMethodCall(utils.ESTreeRuntimeExpression(search.Object), trimOptions)
				if !ok || unicornutil.HasOptionalChainElementThroughInstantiation(node) {
					return
				}
				receiver := ast.SkipOuterExpressions(trim.Object, ast.OEKAll)
				if isZodStringCall(ctx, receiver) {
					return
				}

				getEvaluator := func() *utils.StaticStringEvaluator {
					if evaluator == nil {
						evaluator = utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
						evaluator.GlobalAccess = ctx.Globals.Access
					}
					return evaluator
				}
				replacement, message, trimOpposite := "trimStart", trimStartMessage, strings.TrimRightFunc
				if search.Property.Text() == "endsWith" {
					replacement, message, trimOpposite = "trimEnd", trimEndMessage, strings.TrimLeftFunc
				}
				if args := node.Arguments(); len(args) > 0 {
					argument := ast.SkipOuterExpressions(args[0], ast.OEKAll)
					text, known := utils.GetStaticStringLiteralValue(argument)
					if !known {
						text, known = staticSearchString(getEvaluator(), argument)
					}
					if !known || text != trimOpposite(text, ecmascript.IsWhiteSpaceOrLineTerminator) {
						return
					}
				}
				if unicornutil.IsKnownNonStringWithAnnotations(ctx, trim.Object) ||
					isStaticNonString(ctx.SourceFile, getEvaluator(), receiver) ||
					hasNonStringJSDocReceiver(ctx, receiver) {
					return
				}
				ctx.ReportNodeWithDeferredFixes(trim.Property, message, func() []rule.RuleFix {
					return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, trim.Property, replacement)}
				})
			},
		}
	},
}

func staticSearchString(evaluator *utils.StaticStringEvaluator, node *ast.Node) (string, bool) {
	value, known := evaluator.EvalValueIfNoSideEffects(node)
	if !known || !utils.IsPrimitiveStaticValue(value) {
		return "", false
	}
	if text, ok := value.(string); ok {
		return text, true
	}
	return evaluator.EvalToString(node)
}

func isStaticNonString(file *ast.SourceFile, evaluator *utils.StaticStringEvaluator, node *ast.Node) bool {
	// A direct const initializer can prove the receiver is an object even
	// when evaluating the initializer would have side effects.
	if initializer, ok := evaluator.ResolveIdentifierInitializer(node); ok &&
		ast.IsVarConst(initializer.Parent) && ast.GetSourceFileOfNode(initializer) == file {
		node = ast.SkipOuterExpressions(initializer, ast.OEKAll)
	}
	switch node.Kind {
	case ast.KindArrayLiteralExpression, ast.KindArrowFunction, ast.KindClassExpression,
		ast.KindFunctionExpression, ast.KindObjectLiteralExpression:
		return true
	}
	value, known := evaluator.EvalValueIfNoSideEffects(node)
	_, isString := value.(string)
	return known && !isString
}

func isZodStringCall(ctx rule.RuleContext, node *ast.Node) bool {
	if !ast.IsCallExpression(node) {
		return false
	}
	callee := utils.ESTreeCallCallee(node.Expression())
	if callee == nil {
		return false
	}
	callee = ast.SkipOuterExpressions(callee, ast.OEKAll)
	reference := callee
	direct := ast.IsIdentifier(callee)
	if !direct {
		if name, ok := utils.AccessExpressionStaticName(callee); !ok || name != "string" {
			return false
		}
		reference = ast.SkipOuterExpressions(utils.AccessExpressionObject(callee), ast.OEKAll)
	}
	// Resolve the runtime binding so a same-named TS type cannot hide a Zod import.
	symbol := ctx.Refs.ResolveInFile(reference)
	if symbol == nil {
		return false
	}
	for _, binding := range symbol.Declarations {
		if direct {
			if binding.Kind != ast.KindImportSpecifier || binding.PropertyNameOrName().Text() != "string" {
				continue
			}
		} else {
			switch binding.Kind {
			case ast.KindImportClause, ast.KindNamespaceImport:
			case ast.KindImportSpecifier:
				if name := binding.PropertyNameOrName().Text(); name != "z" && name != "default" {
					continue
				}
			default:
				continue
			}
		}
		if ast.IsTypeOnlyImportDeclaration(binding) {
			continue
		}
		if declaration := ast.FindAncestorKind(binding, ast.KindImportDeclaration); declaration != nil {
			if source := ast.GetExternalModuleName(declaration); source != nil && ast.IsStringLiteral(source) && source.Text() == "zod" {
				return true
			}
		}
	}
	return false
}
