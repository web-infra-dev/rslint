package no_exports_assign

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoExportsAssignRule enforces that 'exports' is not assigned to directly,
// unless it's also assigning to 'module.exports'.
//
// See: https://github.com/eslint-community/eslint-plugin-n/blob/master/docs/rules/no-exports-assign.md
var NoExportsAssignRule = rule.Rule{
	Name: "n/no-exports-assign",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binaryExpr := node.AsBinaryExpression()
				// Only check assignment expressions
				if binaryExpr.OperatorToken.Kind != ast.KindEqualsToken {
					return
				}

				// Check if left side is 'exports'
				if !isExports(binaryExpr.Left) {
					return
				}

				// Check for allowed cases
				// 1. module.exports = exports = {} (Parent is assignment to module.exports)
				if isAssignmentToModuleExports(node.Parent) {
					return
				}

				// 2. exports = module.exports = {} (Right is assignment to module.exports)
				// Note: binaryExpr.Right is an Expression, which can be a BinaryExpression (assignment)
				if isAssignmentToModuleExports(binaryExpr.Right) {
					return
				}

				// 3. exports = module.exports (Right is module.exports)
				// Allow syncing exports with module.exports
				if isModuleExports(binaryExpr.Right) {
					return
				}

				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "noExportsAssign",
					Description: "Unexpected assignment to 'exports' variable. Use 'module.exports' instead.",
				})
			},
		}
	},
}

// isExports checks if the node is the identifier "exports"
func isExports(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindIdentifier {
		return false
	}
	return node.AsIdentifier().Text == "exports"
}

// isAssignmentToModuleExports checks if the node is an assignment to "module.exports"
func isAssignmentToModuleExports(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return false
	}
	binaryExpr := node.AsBinaryExpression()
	if binaryExpr.OperatorToken.Kind != ast.KindEqualsToken {
		return false
	}

	return isModuleExports(binaryExpr.Left)
}

// isModuleExports checks if the node is "module.exports"
func isModuleExports(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	propAccess := node.AsPropertyAccessExpression()

	// Check object is 'module'
	if propAccess.Expression == nil || propAccess.Expression.Kind != ast.KindIdentifier {
		return false
	}
	if propAccess.Expression.AsIdentifier().Text != "module" {
		return false
	}

	// Check property is 'exports'
	if propAccess.Name() == nil {
		return false
	}
	return propAccess.Name().Text() == "exports"
}
