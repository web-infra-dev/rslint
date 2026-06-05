package no_nesting

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/promiseutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const skipTransparent = ast.OEKParentheses

// isThenOrCatchCall reports whether node is a call whose callee is a non-computed
// .then or .catch member access. Mirrors eslint-plugin-promise's has-promise-callback.
func isThenOrCatchCall(node *ast.Node) bool {
	return promiseutil.IsMemberCall(node, "then") || promiseutil.IsMemberCall(node, "catch")
}

// isPromiseCallback reports whether node is a FunctionExpression or ArrowFunction
// directly passed as argument to a .then() or .catch() call.
// Mirrors eslint-plugin-promise's lib/is-inside-promise helper.
func isPromiseCallback(node *ast.Node) bool {
	if node.Kind != ast.KindFunctionExpression && node.Kind != ast.KindArrowFunction {
		return false
	}
	// In tsgo, an argument node's parent is the containing CallExpression.
	parent := node.Parent
	for parent != nil && ast.IsOuterExpression(parent, skipTransparent) {
		parent = parent.Parent
	}
	return isThenOrCatchCall(parent)
}

// collectScopeBindings collects into names all binding identifiers introduced by fn:
// parameters and locally declared variables at any nesting depth (stopping at nested
// function boundaries). This mirrors ESLint's scope.variables for the callback's scope.
func collectScopeBindings(fn *ast.Node, names map[string]bool) {
	var params *ast.NodeList
	switch fn.Kind {
	case ast.KindFunctionExpression:
		if fe := fn.AsFunctionExpression(); fe != nil {
			params = fe.Parameters
		}
	case ast.KindArrowFunction:
		if af := fn.AsArrowFunction(); af != nil {
			params = af.Parameters
		}
	}
	if params != nil {
		for _, param := range params.Nodes {
			if param.Kind == ast.KindParameter {
				if pd := param.AsParameterDeclaration(); pd != nil {
					utils.CollectBindingNames(pd.Name(), func(_ *ast.Node, name string) {
						names[name] = true
					})
				}
			}
		}
	}

	body := fn.Body()
	if body != nil && body.Kind == ast.KindBlock {
		collectDeclsInBlock(body, names)
	}
}

func collectDeclsInBlock(block *ast.Node, names map[string]bool) {
	stmts := block.AsBlock().Statements
	if stmts == nil {
		return
	}
	for _, stmt := range stmts.Nodes {
		collectDeclsInStmt(stmt, names)
	}
}

func collectDeclsInStmt(node *ast.Node, names map[string]bool) {
	if node == nil || ast.IsFunctionLike(node) {
		return
	}
	switch node.Kind {
	case ast.KindVariableStatement:
		collectVarDeclListNames(node.AsVariableStatement().DeclarationList, names)
	case ast.KindVariableDeclarationList:
		// Reached when a for/for-in/for-of initializer is a declaration list.
		collectVarDeclListNames(node, names)
	default:
		// Recurse into all other nodes (if/for/while/try/block/etc.).
		node.ForEachChild(func(child *ast.Node) bool {
			collectDeclsInStmt(child, names)
			return false
		})
	}
}

func collectVarDeclListNames(listNode *ast.Node, names map[string]bool) {
	if listNode == nil {
		return
	}
	dl := listNode.AsVariableDeclarationList()
	if dl == nil || dl.Declarations == nil {
		return
	}
	for _, decl := range dl.Declarations.Nodes {
		if vd := decl.AsVariableDeclaration(); vd != nil {
			utils.CollectBindingNames(vd.Name(), func(_ *ast.Node, name string) {
				names[name] = true
			})
		}
	}
}

// argsContainRef reports whether any identifier anywhere in the call's argument list
// has a name present in names. This mirrors the upstream's position-based reference
// check (ESLint scope references are checked against argument ranges).
func argsContainRef(callNode *ast.Node, names map[string]bool) bool {
	args := callNode.AsCallExpression().Arguments
	if args == nil {
		return false
	}
	for _, arg := range args.Nodes {
		if identInNames(arg, names) {
			return true
		}
	}
	return false
}

func identInNames(node *ast.Node, names map[string]bool) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindIdentifier {
		return names[node.AsIdentifier().Text]
	}
	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		if identInNames(child, names) {
			found = true
			return true
		}
		return false
	})
	return found
}

func buildAvoidNestingMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "avoidNesting",
		Description: "Avoid nesting promises.",
	}
}

var NoNestingRule = rule.Rule{
	Name: "promise/no-nesting",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Stack of promise-callback function nodes, closest last.
		callbackStack := []*ast.Node{}

		onEnter := func(node *ast.Node) {
			if isPromiseCallback(node) {
				callbackStack = append(callbackStack, node)
			}
		}
		onExit := func(node *ast.Node) {
			if isPromiseCallback(node) && len(callbackStack) > 0 {
				callbackStack = callbackStack[:len(callbackStack)-1]
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionExpression:                      onEnter,
			ast.KindArrowFunction:                           onEnter,
			rule.ListenerOnExit(ast.KindFunctionExpression): onExit,
			rule.ListenerOnExit(ast.KindArrowFunction):      onExit,

			ast.KindCallExpression: func(node *ast.Node) {
				if !isThenOrCatchCall(node) || len(callbackStack) == 0 {
					return
				}
				closestCallback := callbackStack[len(callbackStack)-1]
				names := make(map[string]bool)
				collectScopeBindings(closestCallback, names)
				if argsContainRef(node, names) {
					return
				}
				callee := ast.SkipOuterExpressions(node.AsCallExpression().Expression, skipTransparent)
				if callee == nil || !ast.IsPropertyAccessExpression(callee) {
					return
				}
				nameNode := callee.AsPropertyAccessExpression().Name()
				if nameNode != nil {
					ctx.ReportNode(nameNode, buildAvoidNestingMessage())
				}
			},
		}
	},
}
