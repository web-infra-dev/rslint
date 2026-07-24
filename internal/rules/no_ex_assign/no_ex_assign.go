package no_ex_assign

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Message builder
func buildExAssignMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: "Do not assign to the exception parameter.",
	}
}

// checkBindingReassignments reports every write-reference to ident's own
// declared symbol. RefStore resolution is scope-correct by construction, so a
// local binding that shadows the catch parameter is never returned as a
// reference here.
func checkBindingReassignments(ident *ast.Node, ctx *rule.RuleContext) {
	decl := ident.Parent
	if decl == nil || decl.Name() != ident {
		return
	}
	sym := decl.Symbol()
	if sym == nil {
		return
	}
	for _, ref := range ctx.Refs.References(sym) {
		if utils.IsWriteReference(ref) {
			ctx.ReportNode(ref, buildExAssignMessage())
		}
	}
}

var NoExAssignRule = rule.Rule{
	Name: "no-ex-assign",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCatchClause: func(node *ast.Node) {
				varDeclNode := node.AsCatchClause().VariableDeclaration
				if varDeclNode == nil {
					return
				}
				varDecl := varDeclNode.AsVariableDeclaration()
				if varDecl == nil || varDecl.Name() == nil {
					return
				}

				utils.CollectBindingNames(varDecl.Name(), func(ident *ast.Node, _ string) {
					checkBindingReassignments(ident, &ctx)
				})
			},
		}
	},
}
