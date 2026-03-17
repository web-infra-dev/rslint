package no_var

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-var
var NoVarRule = rule.Rule{
	Name: "no-var",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				// BlockScoped = Let | Const | Using | AwaitUsing
				// If none of those flags are set, it's a var declaration.
				if node.Flags&ast.NodeFlagsBlockScoped != 0 {
					return
				}

				// Skip var inside `declare global { var ... }` (TypeScript ambient context)
				if isInDeclareGlobal(node) {
					return
				}

				// Report on the VariableStatement parent if it exists (for standalone var),
				// otherwise report on the VariableDeclarationList itself (for-loop initializer).
				reportNode := node
				if node.Parent != nil && node.Parent.Kind == ast.KindVariableStatement {
					reportNode = node.Parent
				}

				ctx.ReportNode(reportNode, rule.RuleMessage{
					Id:          "unexpectedVar",
					Description: "Unexpected var, use let or const instead.",
				})
			},
		}
	},
}

// isInDeclareGlobal checks if a node is inside a `declare global { }` block.
func isInDeclareGlobal(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		if current.Kind == ast.KindModuleBlock {
			// Check if the parent ModuleDeclaration is `declare global`
			parent := current.Parent
			if parent != nil && parent.Kind == ast.KindModuleDeclaration {
				modDecl := parent.AsModuleDeclaration()
				if modDecl != nil && modDecl.Name() != nil &&
					modDecl.Name().Kind == ast.KindGlobalKeyword {
					return true
				}
			}
		}
		current = current.Parent
	}
	return false
}
