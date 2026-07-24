package no_class_assign

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Message builder
func buildClassReassignmentMessage(className string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "classReassignment",
		Description: "'" + className + "' is a class.",
	}
}

// checkClassReassignments reports every write-reference to classNode's own
// symbol. RefStore resolution is scope-correct by construction, so a local
// binding that shadows the class name is never returned as a reference here.
func checkClassReassignments(classNode *ast.Node, ctx *rule.RuleContext) {
	sym := classNode.Symbol()
	if sym == nil {
		return
	}
	for _, ref := range ctx.Refs.References(sym) {
		if utils.IsWriteReference(ref) {
			ctx.ReportNode(ref, buildClassReassignmentMessage(sym.Name))
		}
	}
}

// NoClassAssignRule disallows reassigning class declarations
var NoClassAssignRule = rule.Rule{
	Name: "no-class-assign",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			// Check class declarations
			ast.KindClassDeclaration: func(node *ast.Node) {
				if node.AsClassDeclaration().Name() == nil {
					return
				}
				checkClassReassignments(node, &ctx)
			},

			// Check named class expressions. The name is only visible inside
			// the class body, which RefStore's scope-aware resolution
			// enforces on its own.
			ast.KindClassExpression: func(node *ast.Node) {
				if node.AsClassExpression().Name() == nil {
					return
				}
				checkClassReassignments(node, &ctx)
			},
		}
	},
}
