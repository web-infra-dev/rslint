package no_func_assign

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "isAFunction",
		Description: "'" + name + "' is a function.",
	}
}

// checkReassignments reports every write-reference to declNode's own symbol.
// RefStore resolution is scope-correct by construction, so a local binding
// that shadows the function name is never returned as a reference here.
func checkReassignments(declNode *ast.Node, ctx *rule.RuleContext) {
	sym := declNode.Symbol()
	if sym == nil {
		return
	}
	for _, ref := range ctx.Refs.References(sym) {
		if utils.IsWriteReference(ref) {
			ctx.ReportNode(ref, buildMessage(sym.Name))
		}
	}
}

// NoFuncAssignRule disallows reassigning function declarations.
var NoFuncAssignRule = rule.Rule{
	Name: "no-func-assign",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				nameNode := node.Name()
				if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
					return
				}
				checkReassignments(node, &ctx)
			},

			// Named function expressions: the name is only visible inside the
			// body, which RefStore's scope-aware resolution enforces on its
			// own.
			ast.KindFunctionExpression: func(node *ast.Node) {
				nameNode := node.Name()
				if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
					return
				}
				checkReassignments(node, &ctx)
			},
		}
	},
}
