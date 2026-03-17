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

// isConstBinding checks if a variable declaration is a const binding
func isConstBinding(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindVariableDeclarationList {
		return false
	}

	varDeclList := node.AsVariableDeclarationList()
	if varDeclList == nil {
		return false
	}

	// Check if the declaration is const (or using/await using in the future)
	// In TypeScript AST, const declarations have flags
	return (varDeclList.Flags & ast.NodeFlagsConst) != 0
}

// getIdentifierName gets the name of an identifier node
func getIdentifierName(node *ast.Node) string {
	if node == nil || node.Kind != ast.KindIdentifier {
		return ""
	}

	return node.Text()
}

// checkIdentifierWrite checks if an identifier is a write reference to a const variable
func checkIdentifierWrite(node *ast.Node, ctx *rule.RuleContext, constSymbols map[*ast.Symbol]bool) {
	// Check if this is a write reference (assignment, increment, etc.)
	if !utils.IsWriteReference(node) {
		return
	}

	// Get the symbol for this identifier
	if ctx.TypeChecker == nil {
		return
	}

	symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
	if symbol == nil {
		return
	}

	// Check if this symbol refers to a const variable
	if !constSymbols[symbol] {
		return
	}

	// Report the violation
	identName := getIdentifierName(node)
	ctx.ReportNode(node, buildConstMessage(identName))
}

// NoConstAssignRule disallows reassigning const variables
var NoConstAssignRule = rule.CreateRule(rule.Rule{
	Name: "no-const-assign",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Track const declarations by their symbol
		constSymbols := make(map[*ast.Symbol]bool)

		return rule.RuleListeners{
			// Track const variable declarations
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				if !isConstBinding(node) {
					return
				}

				varDeclList := node.AsVariableDeclarationList()
				if varDeclList == nil || varDeclList.Declarations == nil {
					return
				}

				// Track all identifiers declared as const using their symbols
				for _, decl := range varDeclList.Declarations.Nodes {
					if decl.Kind != ast.KindVariableDeclaration {
						continue
					}

					varDecl := decl.AsVariableDeclaration()
					if varDecl == nil || varDecl.Name() == nil {
						continue
					}

					// Collect symbols for all identifiers in the binding name
					collectSymbols(varDecl.Name(), &ctx, constSymbols)
				}
			},

			// Check for reassignments to const variables
			ast.KindIdentifier: func(node *ast.Node) {
				checkIdentifierWrite(node, &ctx, constSymbols)
			},

			// Check shorthand property assignments in destructuring (e.g., {x} = obj)
			ast.KindShorthandPropertyAssignment: func(node *ast.Node) {
				shorthand := node.AsShorthandPropertyAssignment()
				if shorthand == nil || shorthand.Name() == nil {
					return
				}

				// Check if this shorthand is in a destructuring assignment
				if !utils.IsInDestructuringAssignment(node) {
					return
				}

				// This is a write reference, check if it refers to a const variable
				if ctx.TypeChecker == nil {
					return
				}

				symbol := ctx.TypeChecker.GetSymbolAtLocation(shorthand.Name())
				if symbol == nil {
					return
				}

				// Check if this symbol refers to a const variable
				if !constSymbols[symbol] {
					return
				}

				// Report the violation
				identName := getIdentifierName(shorthand.Name())
				ctx.ReportNode(shorthand.Name(), buildConstMessage(identName))
			},
		}
	},
})

// collectSymbols recursively collects symbols for all identifiers from a binding pattern
func collectSymbols(bindingName *ast.Node, ctx *rule.RuleContext, constSymbols map[*ast.Symbol]bool) {
	if bindingName == nil || ctx.TypeChecker == nil {
		return
	}

	switch bindingName.Kind {
	case ast.KindIdentifier:
		symbol := ctx.TypeChecker.GetSymbolAtLocation(bindingName)
		if symbol != nil {
			constSymbols[symbol] = true
		}

	case ast.KindObjectBindingPattern:
		// Walk through child nodes to find binding elements
		bindingName.ForEachChild(func(child *ast.Node) bool {
			if child.Kind == ast.KindBindingElement {
				bindingElem := child.AsBindingElement()
				if bindingElem != nil && bindingElem.Name() != nil {
					collectSymbols(bindingElem.Name(), ctx, constSymbols)
				}
			}
			return false
		})

	case ast.KindArrayBindingPattern:
		// Walk through child nodes to find binding elements
		bindingName.ForEachChild(func(child *ast.Node) bool {
			if child.Kind == ast.KindBindingElement {
				bindingElem := child.AsBindingElement()
				if bindingElem != nil && bindingElem.Name() != nil {
					collectSymbols(bindingElem.Name(), ctx, constSymbols)
				}
			}
			return false
		})
	}
}
