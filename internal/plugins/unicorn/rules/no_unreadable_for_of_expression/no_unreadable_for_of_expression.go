package no_unreadable_for_of_expression

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const ruleName = "no-unreadable-for-of-expression"

var forOfMessage = rule.RuleMessage{
	Id:          ruleName,
	Description: "Move the complex iterable expression out of the `for…of` loop header.",
	Data:        map[string]string{"loopKind": "for…of"},
}

var forAwaitOfMessage = rule.RuleMessage{
	Id:          ruleName,
	Description: "Move the complex iterable expression out of the `for await…of` loop header.",
	Data:        map[string]string{"loopKind": "for await…of"},
}

var NoUnreadableForOfExpressionRule = rule.Rule{
	Name:   "unicorn/" + ruleName,
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindForOfStatement: func(node *ast.Node) {
				loop := node.AsForInOrOfStatement()
				right := unicornutil.UnwrapIteratorExpression(loop.Expression)
				if right == nil {
					return
				}

				// Eager map and filter calls are reported by no-duplicate-loops.
				if call, ok := unicornutil.MatchMapOrFilterCall(right); ok {
					if !unicornutil.IsIteratorExpression(&ctx, call.Object) {
						return
					}
				} else if !unicornutil.IsLazyIteratorHelperCall(&ctx, right) && isSimpleIterableExpression(right) {
					return
				}

				message := forOfMessage
				if loop.AwaitModifier != nil {
					message = forAwaitOfMessage
				}
				ctx.ReportNode(utils.ESTreeRuntimeExpression(loop.Expression), message)
			},
		}
	},
}

func isSimpleIterableExpression(node *ast.Node) bool {
	node = unicornutil.UnwrapIteratorExpression(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindArrayLiteralExpression:
		return true
	case ast.KindPropertyAccessExpression:
		return isSimpleMemberExpression(node)
	case ast.KindCallExpression:
		return isSimpleIterableCall(node)
	default:
		return isSimpleValue(node)
	}
}

func isSimpleIterableCall(node *ast.Node) bool {
	if !hasSimpleCallee(node) {
		return false
	}
	if _, ok := unicornutil.MatchMapOrFilterCall(node); ok {
		return false
	}
	for _, argument := range node.Arguments() {
		argument = unicornutil.UnwrapIteratorExpression(argument)
		if isSimpleArgument(argument) {
			continue
		}
		if argument == nil || !ast.IsArrayLiteralExpression(argument) {
			return false
		}
		for _, element := range argument.Elements() {
			if !isSimpleArgument(unicornutil.UnwrapIteratorExpression(element)) {
				return false
			}
		}
	}
	return true
}

func isSimpleArgument(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindTemplateExpression:
		return true
	case ast.KindPropertyAccessExpression:
		return isSimpleMemberExpression(node)
	case ast.KindCallExpression:
		return len(node.Arguments()) == 0 && hasSimpleCallee(node)
	default:
		return isSimpleValue(node)
	}
}

func hasSimpleCallee(call *ast.Node) bool {
	if ast.IsOptionalChainRoot(call) {
		return false
	}
	callee := utils.ESTreeCallCallee(call.Expression())
	return callee != nil && (ast.IsIdentifier(callee) ||
		(ast.IsPropertyAccessExpression(callee) && isSimpleMemberExpression(callee)))
}

func isSimpleMemberExpression(node *ast.Node) bool {
	if ast.IsOptionalChainRoot(node) {
		return false
	}
	name := node.Name()
	return name != nil && (ast.IsIdentifier(name) || ast.IsPrivateIdentifier(name)) &&
		isSimpleIterableExpression(node.Expression())
}

func isSimpleValue(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindIdentifier, ast.KindThisKeyword, ast.KindSuperKeyword,
		ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral, ast.KindNoSubstitutionTemplateLiteral,
		ast.KindNullKeyword, ast.KindTrueKeyword, ast.KindFalseKeyword:
		return true
	}
	return false
}
