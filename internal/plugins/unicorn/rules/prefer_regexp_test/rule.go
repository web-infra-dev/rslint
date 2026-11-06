package prefer_regexp_test

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/jsnum"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var messages = map[string]string{
	"regexp-exec":   "Prefer `.test(…)` over `.exec(…)`.",
	"string-match":  "Prefer `RegExp#test(…)` over `String#match(…)`.",
	"string-search": "Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.",
	"suggestion":    "Switch to `RegExp#test(…)`.",
}

func message(id string) rule.RuleMessage {
	return rule.RuleMessage{Id: id, Description: messages[id]}
}

// PreferRegexpTestRule follows eslint-plugin-unicorn v79.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-regexp-test.js
var PreferRegexpTestRule = rule.Rule{
	Name:   "unicorn/prefer-regexp-test",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		evaluator := utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
		evaluator.GlobalAccess = ctx.Globals.Access
		var types *expressionTypes
		return rule.RuleListeners{ast.KindCallExpression: func(node *ast.Node) {
			one := 1
			call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
				Methods: []string{"exec", "match", "search"}, ArgumentsLength: &one, RejectSpreadElement: true,
			})
			if !ok {
				return
			}
			method := call.Property.Text()
			search, negate := searchCheck(node)
			if (method == "search") != (search != nil) {
				return
			}
			length := lengthCheck(ctx, node)
			if length == nil && search == nil && (ast.IsOptionalChain(node) || !isBooleanContext(ctx, node)) {
				return
			}
			stringNode, regexpNode := utils.ESTreeRuntimeExpression(call.Object), utils.ESTreeRuntimeExpression(node.Arguments()[0])
			id := "string-" + method
			report := node
			if method == "exec" {
				stringNode, regexpNode = regexpNode, stringNode
				id, report = "regexp-exec", call.Property
			} else if method == "match" && hasUnparenthesizedOptionalChain(call.Object) {
				return
			}
			if types == nil {
				types = &expressionTypes{ctx: ctx, evaluator: evaluator, visiting: make(map[*ast.Symbol]bool)}
			}
			stringType, regexpType := types.expression(stringNode), types.expression(regexpNode)
			if stringType == typeOther || stringType == typeRegExp || stringType == typeRegExpOrOther ||
				regexpType == typeOther || regexpType == typeString || regexpType == typeStringOrOther {
				return
			}
			buildFix := func() []rule.RuleFix {
				var fixes []rule.RuleFix
				if method == "exec" {
					fixes = []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, call.Property, "test")}
				} else {
					fixes = swapReceiverFixes(ctx, call.Property, stringNode, regexpNode, negate)
				}
				if length != nil {
					for _, span := range length.removals(ctx) {
						fixes = append(fixes, rule.RuleFixRemoveRange(span))
					}
				} else if search != nil {
					if negate {
						start := utils.TrimNodeTextRange(ctx.SourceFile, search).Pos()
						fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(start, start), "!"))
					}
					fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(parenthesizedEnd(node), search.End())))
				}
				return fixes
			}
			// Edit policy belongs to the rule. The evaluator only supplies a
			// safe static RegExp value, without exposing its internal objects.
			ctx.ReportNodeWithDeferredFixesAndSuggestions(report, message(id), func() []rule.RuleFix {
				if length != nil || !canRemoveSearchTail(ctx, node, search) || !canAutofix(evaluator, regexpNode, search != nil) {
					return nil
				}
				return buildFix()
			}, func() []rule.RuleSuggestion {
				if length != nil {
					for _, span := range length.removals(ctx) {
						if utils.HasCommentInSpan(ctx.Comments.All(), span.Pos(), span.End()) {
							return nil
						}
					}
				} else if !canRemoveSearchTail(ctx, node, search) || canAutofix(evaluator, regexpNode, search != nil) {
					return nil
				}
				return []rule.RuleSuggestion{{Message: message("suggestion"), FixesArr: buildFix()}}
			})
		}}
	},
}

func isBooleanContext(ctx rule.RuleContext, node *ast.Node) bool {
	return unicornutil.IsBooleanExpression(ctx, node) || unicornutil.IsControlFlowTest(node)
}

func searchCheck(node *ast.Node) (*ast.Node, bool) {
	if ast.IsOptionalChain(node) {
		return nil, false
	}
	parent := utils.ESTreeParent(node)
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return nil, false
	}
	bin := parent.AsBinaryExpression()
	if utils.ESTreeRuntimeExpression(bin.Left) != node {
		return nil, false
	}
	right := utils.ESTreeRuntimeExpression(bin.Right)
	negativeOne := right.Kind == ast.KindPrefixUnaryExpression && right.AsPrefixUnaryExpression().Operator == ast.KindMinusToken && isNumber(right.AsPrefixUnaryExpression().Operand, 1)
	switch bin.OperatorToken.Kind {
	case ast.KindExclamationEqualsToken, ast.KindExclamationEqualsEqualsToken, ast.KindGreaterThanToken:
		if negativeOne {
			return parent, false
		}
	case ast.KindEqualsEqualsToken, ast.KindEqualsEqualsEqualsToken:
		if negativeOne {
			return parent, true
		}
	case ast.KindGreaterThanEqualsToken:
		if isNumber(right, 0) {
			return parent, false
		}
	case ast.KindLessThanToken:
		if isNumber(right, 0) {
			return parent, true
		}
	}
	return nil, false
}

func isNumber(node *ast.Node, value jsnum.Number) bool {
	node = utils.ESTreeRuntimeExpression(node)
	return node != nil && node.Kind == ast.KindNumericLiteral && jsnum.FromString(node.Text()) == value
}

type lengthWrapper struct{ member, comparison *ast.Node }

func lengthCheck(ctx rule.RuleContext, node *ast.Node) *lengthWrapper {
	member := utils.ESTreeParent(node)
	if member == nil || member.Kind != ast.KindPropertyAccessExpression ||
		member.Name().Kind != ast.KindIdentifier || member.Name().Text() != "length" ||
		utils.ESTreeRuntimeExpression(member.Expression()) != node || isNegatedBooleanValue(ctx, member) {
		return nil
	}
	if isBooleanContext(ctx, member) {
		return &lengthWrapper{member: member}
	}
	parent := utils.ESTreeParent(member)
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return nil
	}
	bin := parent.AsBinaryExpression()
	if utils.ESTreeRuntimeExpression(bin.Left) == member && bin.OperatorToken.Kind == ast.KindGreaterThanToken && isNumber(bin.Right, 0) &&
		!isNegatedBooleanValue(ctx, parent) && isBooleanContext(ctx, parent) {
		return &lengthWrapper{member: member, comparison: parent}
	}
	return nil
}

func isNegatedBooleanValue(ctx rule.RuleContext, node *ast.Node) bool {
	for {
		parent := utils.ESTreeParent(node)
		if unicornutil.IsLogicalExpression(parent) || (unicornutil.IsGlobalBooleanCall(ctx, parent) && utils.ESTreeRuntimeExpression(parent.Arguments()[0]) == node) {
			node = parent
			continue
		}
		return unicornutil.IsLogicalNot(parent)
	}
}

func (length *lengthWrapper) removals(ctx rule.RuleContext) []core.TextRange {
	ranges := []core.TextRange{core.NewTextRange(parenthesizedEnd(length.member.Expression()), length.member.End())}
	if length.comparison != nil {
		ranges = append(ranges, core.NewTextRange(parenthesizedEnd(length.member), length.comparison.End()))
	}
	return ranges
}

func parenthesizedEnd(node *ast.Node) int {
	return utils.OutermostParenthesizedExpression(node).End()
}

func canRemoveSearchTail(ctx rule.RuleContext, node, search *ast.Node) bool {
	return search == nil || !utils.HasCommentInSpan(ctx.Comments.All(), parenthesizedEnd(node), search.End())
}

func canAutofix(evaluator *utils.StaticStringEvaluator, node *ast.Node, search bool) bool {
	if node.Kind == ast.KindRegularExpressionLiteral {
		_, flags := utils.ExtractRegexPatternAndFlags(node.Text())
		if !strings.ContainsAny(flags, "gy") {
			return true
		}
	}
	_, flags, known := evaluator.EvalRegExp(node)
	return known && !strings.ContainsRune(flags, 'g') && (!search || !strings.ContainsRune(flags, 'y'))
}

// Parentheses end the optional chain. TS assertions do not, so follow only
// erased wrappers and access/call receivers inside the original chain.
func hasUnparenthesizedOptionalChain(node *ast.Node) bool {
	for node != nil && !ast.IsParenthesizedExpression(node) {
		if ast.IsOptionalChainRoot(node) {
			return true
		}
		if !ast.IsAccessExpression(node) && !ast.IsCallExpression(node) &&
			node.Kind != ast.KindNonNullExpression && !utils.IsJSDocTypeCastWrapper(node) {
			return false
		}
		node = node.Expression()
	}
	return false
}

func swapReceiverFixes(ctx rule.RuleContext, method, stringNode, regexpNode *ast.Node, negate bool) []rule.RuleFix {
	stringText := utils.TrimmedNodeText(ctx.SourceFile, stringNode)
	if utils.OutermostParenthesizedExpression(regexpNode) == regexpNode && stringNode.Kind == ast.KindBinaryExpression && stringNode.AsBinaryExpression().OperatorToken.Kind == ast.KindCommaToken {
		stringText = "(" + stringText + ")"
	}
	regexpText := utils.TrimmedNodeText(ctx.SourceFile, regexpNode)
	position := stringNode
	if negate {
		position = regexpNode
	}
	if utils.OutermostParenthesizedExpression(stringNode) == stringNode && unicornutil.ShouldAddParenthesesToMemberExpressionObjectAt(ctx.SourceFile, regexpNode, position) {
		regexpText = "(" + regexpText + ")"
	}
	var fixes []rule.RuleFix
	if unicornutil.NeedsSemicolonBefore(ctx.SourceFile, stringNode, regexpText) {
		start := utils.TrimNodeTextRange(ctx.SourceFile, stringNode).Pos()
		fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(start, start), ";"))
	}
	fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, method, "test"), rule.RuleFixReplace(ctx.SourceFile, regexpNode, stringText), rule.RuleFixReplace(ctx.SourceFile, stringNode, regexpText))
	return append(fixes, unicornutil.SpaceAroundKeywordFixes(ctx.SourceFile, stringNode)...)
}
