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
//
// name comes from the declaration's name node rather than the symbol: a
// default-exported class is bound to an export symbol named "default",
// while the reported name must be the one written in the source.
func checkClassReassignments(classNode *ast.Node, name string, ctx *rule.RuleContext) {
	sym := classNode.Symbol()
	if sym == nil {
		return
	}
	for _, ref := range ctx.Refs.References(sym) {
		if utils.IsWriteReference(ref) {
			ctx.ReportNode(ref, buildClassReassignmentMessage(name))
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
				nameNode := node.AsClassDeclaration().Name()
				if nameNode == nil {
					return
				}
				checkClassReassignments(node, nameNode.Text(), &ctx)
			},

			// Check named class expressions. The name is only visible inside
			// the class body, which RefStore's scope-aware resolution
			// enforces on its own.
			ast.KindClassExpression: func(node *ast.Node) {
				nameNode := node.AsClassExpression().Name()
				if nameNode == nil {
					return
				}
				checkClassReassignments(node, nameNode.Text(), &ctx)
			},
		}
	},
}
