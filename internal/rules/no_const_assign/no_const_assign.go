package no_const_assign

import (
	"sort"

	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildConstMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "const",
		Description: "'" + name + "' is constant.",
	}
}

// NoConstAssignRule disallows reassigning const variables
var NoConstAssignRule = rule.Rule{
	Name:             "no-const-assign",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		// Violations are collected per declaration (as the traversal visits each
		// VariableDeclarationList) but reported only once the whole file has
		// been walked, sorted back into source order: two separate `const`
		// declarations of the same name in different scopes are visited in
		// declaration order, not in the order their writes actually appear.
		var violations []*ast.Node

		flush := func() {
			sort.Slice(violations, func(i, j int) bool {
				return violations[i].Pos() < violations[j].Pos()
			})
			for _, ref := range violations {
				ctx.ReportNode(ref, buildConstMessage(ref.Text()))
			}
		}

		listeners := rule.RuleListeners{
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				if ctx.Refs == nil || node.Flags&ast.NodeFlagsConst == 0 {
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
						sym := bindingDecl.Symbol()
						if sym == nil {
							return
						}
						for _, ref := range ctx.Refs.References(sym) {
							if utils.IsWriteReference(ref) {
								violations = append(violations, ref)
							}
						}
					})
				}
			},
		}

		// Flush once the whole file has been walked. The traversal never visits
		// the SourceFile node itself (only its children), so the hook is the
		// exit of the last top-level statement, mirroring no_unused_vars.
		statements := ctx.SourceFile.Statements
		if statements == nil || len(statements.Nodes) == 0 {
			flush()
		} else {
			lastTopLevelNode := statements.Nodes[len(statements.Nodes)-1]
			listeners[rule.ListenerOnExit(lastTopLevelNode.Kind)] = func(node *ast.Node) {
				if node == lastTopLevelNode {
					flush()
				}
			}
		}

		return listeners
	},
}
