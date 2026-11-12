// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
package no_array_sort_for_min_max

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	messageIDError      = "no-array-sort-for-min-max/error"
	messageIDSuggestion = "no-array-sort-for-min-max/suggestion"
	typeScriptWrappers  = ast.OEKAssertions | ast.OEKExpressionsWithTypeArguments
)

var NoArraySortForMinMaxRule = rule.Rule{
	Name:   "unicorn/no-array-sort-for-min-max",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		oneArgument := 1
		atOptions := unicornutil.DotMethodCallOptions{Method: "at", ArgumentsLength: &oneArgument, RejectSpreadElement: true}
		sortOptions := unicornutil.DotMethodCallOptions{Methods: []string{"sort", "toSorted"}, ArgumentsLength: &oneArgument, RejectSpreadElement: true}
		receiverOptions := unicornutil.ArrayReceiverOptions{CheckArrayProducers: true, ReadJSDoc: true}
		check := func(node, sorted *ast.Node, last bool) {
			call, ok := unicornutil.MatchDotMethodCall(utils.ESTreeRuntimeExpression(sorted), sortOptions)
			if !ok || unicornutil.IsReferenceTarget(node) || unicornutil.HasOptionalChainElementThroughInstantiation(call.Object) {
				return
			}
			direction := comparatorDirection(call.Call.Arguments()[0])
			if direction == unordered ||
				unicornutil.IsKnownNonIndexedCollectionWithOptions(ctx, call.Object, receiverOptions) ||
				unicornutil.IsKnownBigIntTypedArray(ctx, call.Object) {
				return
			}
			method, extreme := "min", "minimum"
			if direction == ascending && last || direction == descending && !last {
				method, extreme = "max", "maximum"
			}
			reportExtreme(ctx, node, utils.ESTreeRuntimeExpression(call.Object), method, extreme)
		}
		return rule.RuleListeners{
			ast.KindElementAccessExpression: func(node *ast.Node) {
				if ast.IsOptionalChainRoot(node) {
					return
				}
				index := utils.ESTreeRuntimeExpression(node.AsElementAccessExpression().ArgumentExpression)
				if isNumericLiteral(index, "0") {
					check(node, node.Expression(), false)
				}
			},
			ast.KindCallExpression: func(node *ast.Node) {
				call, ok := unicornutil.MatchDotMethodCall(node, atOptions)
				if !ok {
					return
				}
				index := utils.ESTreeRuntimeExpression(node.Arguments()[0])
				if isNumericLiteral(index, "0") {
					check(node, call.Object, false)
				} else if ast.IsPrefixUnaryExpression(index) {
					unary := index.AsPrefixUnaryExpression()
					if unary.Operator == ast.KindMinusToken && isNumericLiteral(utils.ESTreeRuntimeExpression(unary.Operand), "1") {
						check(node, call.Object, true)
					}
				}
			},
		}
	},
}

func isNumericLiteral(node *ast.Node, value string) bool {
	// The tsgo scanner stores the canonical Number text, including for hex,
	// binary, octal, and scientific notation.
	return node != nil && node.Kind == ast.KindNumericLiteral && node.Text() == value
}

type sortDirection uint8

const (
	unordered sortDirection = iota
	ascending
	descending
)

func comparatorDirection(node *ast.Node) sortDirection {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || (!ast.IsArrowFunction(node) && !ast.IsFunctionExpression(node)) ||
		ast.GetFunctionFlags(node) != ast.FunctionFlagsNormal {
		return unordered
	}
	parameters := node.Parameters()
	if len(parameters) != 2 {
		return unordered
	}
	first := unicornutil.PlainParameterIdentifier(parameters[0])
	second := unicornutil.PlainParameterIdentifier(parameters[1])
	if first == nil || second == nil || first.Text() == second.Text() {
		return unordered
	}

	body := node.Body()
	if ast.IsBlock(body) {
		statements := body.AsBlock().Statements.Nodes
		if len(statements) != 1 || statements[0].Kind != ast.KindReturnStatement {
			return unordered
		}
		body = statements[0].AsReturnStatement().Expression
	}
	if body == nil {
		return unordered
	}
	body = ast.SkipOuterExpressions(body, ast.OEKParentheses|typeScriptWrappers)
	if !ast.IsBinaryExpression(body) || body.AsBinaryExpression().OperatorToken.Kind != ast.KindMinusToken {
		return unordered
	}
	binary := body.AsBinaryExpression()
	left := ast.SkipOuterExpressions(binary.Left, ast.OEKParentheses|typeScriptWrappers)
	right := ast.SkipOuterExpressions(binary.Right, ast.OEKParentheses|typeScriptWrappers)
	if unicornutil.IsSameIdentifier(left, first) && unicornutil.IsSameIdentifier(right, second) {
		return ascending
	}
	if unicornutil.IsSameIdentifier(left, second) && unicornutil.IsSameIdentifier(right, first) {
		return descending
	}
	return unordered
}

func reportExtreme(ctx rule.RuleContext, node, source *ast.Node, method, extreme string) {
	span := utils.TrimNodeTextRange(ctx.SourceFile, node)
	ctx.ReportRangeWithDeferredSuggestions(span, rule.RuleMessage{
		Id:          messageIDError,
		Description: "Use `Math." + method + "()` instead of sorting to get the " + extreme + " value.",
		Data:        map[string]string{"method": method, "extreme": extreme},
	}, func() []rule.RuleSuggestion {
		if source.Kind == ast.KindSuperKeyword || !ctx.Globals.Access("Math").IsDeclared() ||
			!ctx.Refs.IsGlobalNameReference(node, "Math", ast.SymbolFlagsValue|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias) {
			return nil
		}
		sourceSpan := utils.TrimNodeTextRange(ctx.SourceFile, source)
		comments := ctx.Comments.All()
		if utils.HasCommentInSpan(comments, span.Pos(), sourceSpan.Pos()) ||
			utils.HasCommentInSpan(comments, sourceSpan.End(), span.End()) {
			return nil
		}
		sourceText := ctx.SourceFile.Text()[sourceSpan.Pos():sourceSpan.End()]
		if !isStrongPrecedenceReceiver(source) {
			sourceText = "(" + sourceText + ")"
		}
		replacement := "Math." + method + "(..." + sourceText + ")"
		fixedText := replacement
		if unicornutil.IsInUnparenthesizedConstructorCallee(node) {
			fixedText = "(" + fixedText + ")"
		}
		fixes := []rule.RuleFix{rule.RuleFixReplaceRange(span, fixedText)}
		fixes = append(fixes, unicornutil.SpaceAroundKeywordFixes(ctx.SourceFile, node)...)
		return []rule.RuleSuggestion{{
			Message: rule.RuleMessage{
				Id:          messageIDSuggestion,
				Description: "Replace with `" + replacement + "`.",
				Data:        map[string]string{"replacement": replacement},
			},
			FixesArr: fixes,
		}}
	})
}

func isStrongPrecedenceReceiver(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindObjectLiteralExpression, ast.KindExpressionWithTypeArguments:
		return false
	case ast.KindThisKeyword, ast.KindNullKeyword, ast.KindTemplateExpression, ast.KindNonNullExpression:
		return true
	default:
		return utils.IsStrongPrecedenceNode(node)
	}
}
