package no_ex_assign

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-ex-assign

func buildExAssignMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: "Do not assign to the exception parameter.",
	}
}

// catchBindingSymbols returns the binder symbol of every name bound by the
// catch clause's variable declaration: the lone identifier of `catch (e)` or
// each name nested in a destructuring pattern. The symbol is attached to the
// declaration node owning the name — the VariableDeclaration itself for a
// plain identifier, the enclosing BindingElement for pattern names.
func catchBindingSymbols(name *ast.Node) []*ast.Symbol {
	var symbols []*ast.Symbol
	utils.CollectBindingNames(name, func(ident *ast.Node, _ string) {
		if ident.Parent == nil {
			return
		}
		if sym := ident.Parent.Symbol(); sym != nil {
			symbols = append(symbols, sym)
		}
	})
	return symbols
}

var NoExAssignRule = rule.Rule{
	Name: "no-ex-assign",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCatchClause: func(node *ast.Node) {
				varDecl := node.AsCatchClause().VariableDeclaration
				if varDecl == nil || varDecl.Name() == nil {
					return
				}

				var writes []*ast.Node
				for _, sym := range catchBindingSymbols(varDecl.Name()) {
					for _, ref := range ctx.Refs.References(sym) {
						if utils.IsWriteReference(ref) {
							writes = append(writes, ref)
						}
					}
				}

				// References is source-ordered per symbol; a destructuring
				// pattern binds several symbols, so restore document order
				// across them before reporting.
				slices.SortFunc(writes, func(a, b *ast.Node) int { return a.Pos() - b.Pos() })
				for _, ref := range writes {
					ctx.ReportNode(ref, buildExAssignMessage())
				}
			},
		}
	},
}
