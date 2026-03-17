package prefer_const

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/prefer-const
var PreferConstRule = rule.Rule{
	Name: "prefer-const",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		if ctx.TypeChecker == nil {
			return rule.RuleListeners{}
		}

		return rule.RuleListeners{
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				declList := node.AsVariableDeclarationList()
				if declList == nil || node.Flags&ast.NodeFlagsLet == 0 || declList.Declarations == nil {
					return
				}

				isForInOrOf := isInForInOrOf(node)

				for _, decl := range declList.Declarations.Nodes {
					varDecl := decl.AsVariableDeclaration()
					if varDecl == nil || varDecl.Name() == nil {
						continue
					}

					// Must have an initializer, OR be in a for-in/for-of loop
					if varDecl.Initializer == nil && !isForInOrOf {
						continue
					}

					checkBindingNames(varDecl.Name(), decl, &ctx)
				}
			},
		}
	},
}

// isInForInOrOf checks if a VariableDeclarationList is the initializer of a for-in or for-of statement.
func isInForInOrOf(node *ast.Node) bool {
	if node.Parent == nil {
		return false
	}
	return node.Parent.Kind == ast.KindForInStatement || node.Parent.Kind == ast.KindForOfStatement
}

// checkBindingNames recursively checks all identifier nodes from a binding pattern.
func checkBindingNames(nameNode *ast.Node, declNode *ast.Node, ctx *rule.RuleContext) {
	switch nameNode.Kind {
	case ast.KindIdentifier:
		checkIdentifier(nameNode, declNode, ctx)

	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		nameNode.ForEachChild(func(child *ast.Node) bool {
			if child.Kind == ast.KindBindingElement {
				bindingName := child.Name()
				if bindingName != nil {
					checkBindingNames(bindingName, declNode, ctx)
				}
			}
			return false
		})
	}
}

// checkIdentifier checks a single identifier to see if it should be const.
func checkIdentifier(nameNode *ast.Node, declNode *ast.Node, ctx *rule.RuleContext) {
	sym := ctx.TypeChecker.GetSymbolAtLocation(nameNode)
	if sym == nil {
		return
	}

	if !isReassigned(sym, nameNode.Text(), declNode, ctx) {
		name := nameNode.Text()
		ctx.ReportNode(nameNode, rule.RuleMessage{
			Id:          "useConst",
			Description: "'" + name + "' is never reassigned. Use 'const' instead.",
		})
	}
}

// isReassigned checks if a symbol is ever assigned to after its declaration.
// Uses a single-pass walk from the enclosing scope rather than the entire source file.
func isReassigned(sym *ast.Symbol, declName string, declNode *ast.Node, ctx *rule.RuleContext) bool {
	// Find enclosing scope to limit the walk
	scope := findEnclosingScope(declNode)
	if scope == nil {
		scope = ctx.SourceFile.AsNode()
	}

	found := false
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n == nil || found {
			return
		}

		if n.Kind == ast.KindIdentifier && !isPartOfDeclaration(n, declNode) {
			refSym := ctx.TypeChecker.GetSymbolAtLocation(n)
			if refSym == sym && utils.IsWriteReference(n) {
				found = true
				return
			}
		}

		// Also check ShorthandPropertyAssignment — in ({x} = {x: 2}), the TypeChecker
		// resolves the shorthand name to the property symbol, not the variable symbol.
		// Use name-based matching combined with scope check for this case.
		if n.Kind == ast.KindShorthandPropertyAssignment && !isPartOfDeclaration(n, declNode) {
			shorthand := n.AsShorthandPropertyAssignment()
			if shorthand != nil && shorthand.Name() != nil && utils.IsInDestructuringAssignment(n) {
				name := shorthand.Name().Text()
				if name == declName && isInSameScope(n, declNode) {
					found = true
					return
				}
			}
		}

		n.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return found
		})
	}
	walk(scope)
	return found
}

// findEnclosingScope finds the nearest function/module/source file scope.
func findEnclosingScope(node *ast.Node) *ast.Node {
	current := node.Parent
	for current != nil {
		switch current.Kind {
		case ast.KindSourceFile, ast.KindModuleBlock:
			return current
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
			return current
		}
		current = current.Parent
	}
	return nil
}

// isPartOfDeclaration checks if an identifier node is part of the variable declaration itself.
func isPartOfDeclaration(identNode *ast.Node, declNode *ast.Node) bool {
	current := identNode
	for current != nil {
		if current == declNode {
			return true
		}
		if current.Kind == ast.KindVariableDeclaration {
			return false
		}
		current = current.Parent
	}
	return false
}

// isInSameScope checks if two nodes share the same enclosing function/module/source scope.
func isInSameScope(a *ast.Node, b *ast.Node) bool {
	return findEnclosingScope(a) == findEnclosingScope(b)
}
