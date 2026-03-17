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

// isNameShadowed checks whether the identifier at `node` refers to a different
// binding than `declNode`'s name, i.e. a local variable/parameter/catch binding
// that shadows the function name.
//
// When a TypeChecker is available, symbol identity is checked first (fast path).
// Because the TypeChecker sometimes resolves shorthand-property identifiers in
// destructuring assignments to a *property* symbol rather than the variable
// symbol, a positive result is confirmed by a scope-based walk before we
// declare the name shadowed.
func isNameShadowed(node *ast.Node, name string, declNode *ast.Node, ctx *rule.RuleContext) bool {
	if node == nil || ctx.TypeChecker == nil {
		return false
	}

	identSymbol := ctx.TypeChecker.GetSymbolAtLocation(node)
	if identSymbol == nil {
		return false
	}

	// Get the symbol at the declaration's name.
	declName := declNode.Name()
	if declName == nil {
		return false
	}
	declSymbol := ctx.TypeChecker.GetSymbolAtLocation(declName)

	if declSymbol != nil {
		if identSymbol == declSymbol {
			return false // same binding — not shadowed
		}
		// Different symbol — confirm with scope analysis before concluding,
		// because some AST positions (e.g. shorthand destructuring) resolve
		// to a property symbol rather than the variable symbol.
		return isInShadowingScope(node, name, declNode)
	}

	return isInShadowingScope(node, name, declNode)
}

// isInShadowingScope walks from `node` up to (but not including) `declNode`,
// returning true if any intermediate scope introduces a binding with the given name.
func isInShadowingScope(node *ast.Node, name string, declNode *ast.Node) bool {
	for current := node.Parent; current != nil && current != declNode; current = current.Parent {
		if ast.IsFunctionLikeDeclaration(current) {
			if utils.HasShadowingParameter(current, name) {
				return true
			}
		}

		if current.Kind == ast.KindBlock {
			if utils.HasShadowingDeclaration(current, name) {
				return true
			}
		}

		if current.Kind == ast.KindCatchClause {
			cc := current.AsCatchClause()
			if cc != nil && cc.VariableDeclaration != nil {
				vd := cc.VariableDeclaration.AsVariableDeclaration()
				if vd != nil && vd.Name() != nil && vd.Name().Kind == ast.KindIdentifier && vd.Name().Text() == name {
					return true
				}
			}
		}
	}
	// Also check declNode's own parameters (e.g. function foo(foo) { foo = bar; }).
	if declNode != nil && utils.HasShadowingParameter(declNode, name) {
		return true
	}
	return false
}

// checkReassignments walks `searchRoot` and reports every write-reference to `name`
// that targets the same binding as `declNode`.
func checkReassignments(searchRoot *ast.Node, name string, declNode *ast.Node, ctx *rule.RuleContext) {
	if name == "" {
		return
	}

	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}

		if node.Kind == ast.KindIdentifier && node.Text() == name {
			// Skip the declaration's own name node.
			if node.Parent == declNode {
				return
			}
			if utils.IsWriteReference(node) && !isNameShadowed(node, name, declNode, ctx) {
				ctx.ReportNode(node, buildMessage(name))
			}
		}

		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}

	walk(searchRoot)
}

// NoFuncAssignRule disallows reassigning function declarations.
var NoFuncAssignRule = rule.Rule{
	Name: "no-func-assign",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				nameNode := node.Name()
				if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
					return
				}
				funcName := nameNode.Text()
				if funcName == "" {
					return
				}

				searchRoot := ast.GetEnclosingBlockScopeContainer(node)
				if searchRoot == nil {
					return
				}

				checkReassignments(searchRoot, funcName, node, &ctx)
			},

			// Named function expressions: the name is only visible inside the body.
			ast.KindFunctionExpression: func(node *ast.Node) {
				nameNode := node.Name()
				if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
					return
				}
				funcName := nameNode.Text()
				if funcName == "" {
					return
				}

				checkReassignments(node, funcName, node, &ctx)
			},
		}
	},
}
