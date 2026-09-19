// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
package prefer_dom_node_append

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var PreferDomNodeAppendRule = rule.Rule{
	Name: "unicorn/prefer-dom-node-append", Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		argumentCount := 1
		return rule.RuleListeners{ast.KindCallExpression: func(node *ast.Node) {
			call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{Method: "appendChild", ArgumentsLength: &argumentCount, RejectSpreadElement: true, AllowOptionalMember: true})
			if !ok || isDefinitelyNotDOMNode(ctx, call.Object) || isDefinitelyNotDOMNode(ctx, node.Arguments()[0]) {
				return
			}
			ctx.ReportNodeWithDeferredFixes(node, rule.RuleMessage{Id: "prefer-dom-node-append", Description: "Prefer `Element#append()` over `Node#appendChild()`."}, func() []rule.RuleFix {
				// appendChild returns its argument, whereas append returns undefined.
				if parent := utils.ESTreeParent(node); parent != nil && ast.IsExpressionStatement(parent) {
					return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, call.Property, "append")}
				}
				return nil
			})
		}}
	},
}

func isDefinitelyNotDOMNode(ctx rule.RuleContext, node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	if utils.IsCommaOperator(node) {
		return isDefinitelyNotDOMNode(ctx, node.AsBinaryExpression().Right)
	}
	if node.Kind == ast.KindVoidExpression {
		return true
	}
	switch node.Kind {
	case ast.KindArrayLiteralExpression, ast.KindArrowFunction, ast.KindClassExpression, ast.KindFunctionExpression,
		ast.KindObjectLiteralExpression, ast.KindTemplateExpression, ast.KindNoSubstitutionTemplateLiteral,
		ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindRegularExpressionLiteral,
		ast.KindNullKeyword, ast.KindTrueKeyword, ast.KindFalseKeyword:
		return true
	}
	if utils.IsUndefinedIdentifier(node) {
		return unicornutil.IsGlobalReference(ctx, node)
	}
	return false
}
