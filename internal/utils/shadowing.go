package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
)

// HasShadowingParameter checks if a function-like node has a parameter
// whose binding name matches the given name (supports destructuring patterns).
func HasShadowingParameter(node *ast.Node, name string) bool {
	for _, param := range node.Parameters() {
		if param == nil {
			continue
		}
		nameNode := param.Name()
		if nameNode == nil {
			continue
		}
		found := false
		CollectBindingNames(nameNode, func(_ *ast.Node, n string) {
			if n == name {
				found = true
			}
		})
		if found {
			return true
		}
	}
	return false
}

// HasShadowingDeclaration checks if a block contains a variable, function,
// or class declaration whose name matches the given name.
func HasShadowingDeclaration(node *ast.Node, name string) bool {
	if node.Kind != ast.KindBlock {
		return false
	}

	block := node.AsBlock()
	if block == nil || block.Statements == nil {
		return false
	}

	for _, stmt := range block.Statements.Nodes {
		if stmt == nil {
			continue
		}
		switch stmt.Kind {
		case ast.KindVariableStatement:
			varStmt := stmt.AsVariableStatement()
			if varStmt == nil || varStmt.DeclarationList == nil {
				continue
			}
			declList := varStmt.DeclarationList.AsVariableDeclarationList()
			if declList == nil || declList.Declarations == nil {
				continue
			}
			for _, decl := range declList.Declarations.Nodes {
				if decl == nil || decl.Kind != ast.KindVariableDeclaration {
					continue
				}
				varDecl := decl.AsVariableDeclaration()
				if varDecl == nil || varDecl.Name() == nil {
					continue
				}
				found := false
				CollectBindingNames(varDecl.Name(), func(_ *ast.Node, n string) {
					if n == name {
						found = true
					}
				})
				if found {
					return true
				}
			}
		case ast.KindFunctionDeclaration:
			if n := stmt.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
				return true
			}
		case ast.KindClassDeclaration:
			if n := stmt.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
				return true
			}
		}
	}

	return false
}
