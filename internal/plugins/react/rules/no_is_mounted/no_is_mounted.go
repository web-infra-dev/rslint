package no_is_mounted

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// isPropertyOrMethodDefinition mirrors ESLint's check for an ancestor whose
// AST type is `Property` (object-literal member) or `MethodDefinition` (class
// method/accessor/constructor). SpreadAssignment (`{...x}`) and class-body
// fields / static-blocks are deliberately excluded — they are neither
// Property nor MethodDefinition in ESTree.
func isPropertyOrMethodDefinition(node *ast.Node) bool {
	if ast.IsMethodOrAccessor(node) {
		return true
	}
	switch node.Kind {
	case ast.KindPropertyAssignment,
		ast.KindShorthandPropertyAssignment,
		ast.KindConstructor:
		return true
	}
	return false
}

var NoIsMountedRule = rule.Rule{
	Name: "react/no-is-mounted",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				callee := ast.SkipParentheses(call.Expression)
				if callee.Kind != ast.KindPropertyAccessExpression {
					return
				}
				prop := callee.AsPropertyAccessExpression()
				if ast.SkipParentheses(prop.Expression).Kind != ast.KindThisKeyword {
					return
				}
				// ESLint's check is `'name' in property && property.name === 'isMounted'`,
				// which is satisfied by both Identifier (`this.isMounted`) and
				// PrivateIdentifier (`this.#isMounted`). In tsgo the
				// PrivateIdentifier's Text retains the leading `#`.
				nameNode := prop.Name()
				if nameNode == nil {
					return
				}
				switch nameNode.Kind {
				case ast.KindIdentifier:
					if nameNode.AsIdentifier().Text != "isMounted" {
						return
					}
				case ast.KindPrivateIdentifier:
					if nameNode.AsPrivateIdentifier().Text != "#isMounted" {
						return
					}
				default:
					return
				}
				if ast.FindAncestor(node.Parent, isPropertyOrMethodDefinition) == nil {
					return
				}
				ctx.ReportNode(callee, rule.RuleMessage{
					Id:          "noIsMounted",
					Description: "Do not use isMounted",
				})
			},
		}
	},
}
