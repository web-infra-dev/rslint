// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
package prefer_dom_node_text_content

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var PreferDomNodeTextContentRule = rule.Rule{
	Name:   "unicorn/prefer-dom-node-text-content",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		isKnownNonDOMNode := unicornutil.NewNonDOMNodeCheck(ctx)
		report := func(key *ast.Node, shorthand bool) {
			ctx.ReportNodeWithDeferredSuggestions(key, rule.RuleMessage{
				Id: "error", Description: "Prefer `.textContent` over `.innerText`.",
			}, func() []rule.RuleSuggestion {
				replacement := "textContent"
				if shorthand {
					replacement = "textContent: innerText"
				}
				return []rule.RuleSuggestion{{
					Message:  rule.RuleMessage{Id: "suggestion", Description: "Switch to `.textContent`."},
					FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, key, replacement)},
				}}
			})
		}

		checkAssignmentProperty := func(node *ast.Node) {
			key, pattern := node.Name(), node.Parent
			if !isInnerText(key) || pattern == nil || !ast.IsObjectLiteralExpression(pattern) {
				return
			}
			target := ast.GetAssignmentTarget(pattern)
			if target == nil || (!ast.IsAssignmentExpression(target, true) && !ast.IsForInOrOfStatement(target)) {
				return
			}
			var source *ast.Node
			if ast.IsAssignmentExpression(target, true) && utils.ESTreeRuntimeExpression(target.AsBinaryExpression().Left) == pattern &&
				!utils.IsDefaultValueInDestructuringAssignment(target) {
				source = target.AsBinaryExpression().Right
			}
			if !isKnownNonDOMNode(pattern) && !isKnownNonDOMNode(source) {
				report(key, node.Kind == ast.KindShorthandPropertyAssignment)
			}
		}

		checkMember := func(node *ast.Node) {
			object, key := utils.MemberExpressionParts(node)
			if isInnerText(key) && !utils.IsInJsxTagName(node) && !utils.IsInJSDocSyntax(node) &&
				!isKnownNonDOMNode(object) {
				report(key, false)
			}
		}

		return rule.RuleListeners{
			ast.KindPropertyAccessExpression: checkMember,
			ast.KindQualifiedName:            checkMember,
			ast.KindBindingElement: func(node *ast.Node) {
				binding, pattern := node.AsBindingElement(), node.Parent
				if binding.DotDotDotToken != nil || pattern == nil || pattern.Kind != ast.KindObjectBindingPattern {
					return
				}
				key := node.PropertyNameOrName()
				if !isInnerText(key) || isKnownNonDOMNode(pattern) {
					return
				}
				if parent := pattern.Parent; parent != nil && ast.IsVariableDeclaration(parent) &&
					isKnownNonDOMNode(parent.AsVariableDeclaration().Initializer) {
					return
				}
				report(key, binding.PropertyName == nil)
			},
			ast.KindPropertyAssignment:          checkAssignmentProperty,
			ast.KindShorthandPropertyAssignment: checkAssignmentProperty,
		}
	},
}

func isInnerText(node *ast.Node) bool {
	return node != nil && ast.IsIdentifier(node) && node.Text() == "innerText"
}
