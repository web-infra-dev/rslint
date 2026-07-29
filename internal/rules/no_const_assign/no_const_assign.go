package no_const_assign

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Message builder
func buildConstMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "const",
		Description: "'" + name + "' is constant.",
	}
}

// NoConstAssignRule disallows reassigning const variables
var NoConstAssignRule = rule.Rule{
	Name: "no-const-assign",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		// Track const declarations by their symbol
		constSymbols := make(map[*ast.Symbol]bool)

		return rule.RuleListeners{
			// Track const variable declarations
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				if node.Flags&ast.NodeFlagsConst == 0 {
					return
				}

				declList := node.AsVariableDeclarationList()
				if declList == nil || declList.Declarations == nil {
					return
				}

				for _, decl := range declList.Declarations.Nodes {
					if decl.Kind != ast.KindVariableDeclaration {
						continue
					}
					varDecl := decl.AsVariableDeclaration()
					if varDecl == nil || varDecl.Name() == nil {
						continue
					}
					utils.CollectBindingNames(varDecl.Name(), func(ident *ast.Node, _ string) {
						// The binder attaches the symbol to the enclosing
						// declaration (VariableDeclaration or BindingElement).
						bindingDecl := ident.Parent
						if bindingDecl == nil || bindingDecl.Name() != ident {
							return
						}
						if sym := bindingDecl.Symbol(); sym != nil {
							constSymbols[sym] = true
						}
					})
				}
			},

			// Check every write reference in the file — including shorthand
			// destructuring writes like `({x} = {x: 1})`, since ctx.Refs.Resolve
			// resolves the shorthand name's own binding symbol rather than the
			// property symbol a TypeChecker lookup would otherwise return.
			ast.KindIdentifier: func(node *ast.Node) {
				if len(constSymbols) == 0 || !utils.IsWriteReference(node) {
					return
				}
				sym := ctx.Refs.Resolve(node)
				if sym == nil || !constSymbols[sym] {
					return
				}
				ctx.ReportNode(node, buildConstMessage(node.Text()))
			},
		}
	},
}
