package no_useless_recursion

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var recursionMessage = rule.RuleMessage{
	Id:          "no-useless-recursion",
	Description: "Use a loop instead of this recursive function call.",
}

var NoUselessRecursionRule = rule.Rule{
	Name:   "unicorn/no-useless-recursion",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindReturnStatement: func(node *ast.Node) {
				checkReturn(ctx, node)
			},
		}
	},
}

func checkReturn(ctx rule.RuleContext, node *ast.Node) {
	expression := utils.ESTreeRuntimeExpression(node.AsReturnStatement().Expression)
	if expression == nil || expression.Kind != ast.KindCallExpression || ast.IsOptionalChain(expression) {
		return
	}
	callee := utils.ESTreeCallCallee(expression.AsCallExpression().Expression)
	if callee == nil || callee.Kind != ast.KindIdentifier {
		return
	}

	function := ast.GetContainingFunction(node)
	if function == nil || (function.Kind != ast.KindFunctionDeclaration && function.Kind != ast.KindFunctionExpression) {
		return
	}
	name := function.Name()
	if name == nil || name.Text() != callee.Text() || function.BodyData().AsteriskToken != nil {
		return
	}
	// Upstream's scope lookup also stops at a same-named type declaration.
	symbol := ctx.Refs.ResolveInFileWithMeaning(callee,
		ast.SymbolFlagsValue|ast.SymbolFlagsType|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias)
	if symbol == nil || !slices.Contains(symbol.Declarations, function) || hasTryOrActiveUsing(node, function) {
		return
	}

	ctx.ReportNode(callee, recursionMessage)
}

func hasTryOrActiveUsing(node, function *ast.Node) bool {
	for ; node != function; node = node.Parent {
		switch node.Kind {
		case ast.KindTryStatement:
			return true
		case ast.KindForStatement, ast.KindForOfStatement:
			if isUsingList(node.Initializer()) {
				return true
			}
		}

		// Only declarations before this branch in an enclosing block are active.
		if parent := node.Parent; parent.Kind == ast.KindBlock {
			for _, statement := range parent.Statements() {
				if statement == node {
					break
				}
				if statement.Kind == ast.KindVariableStatement && isUsingList(statement.AsVariableStatement().DeclarationList) {
					return true
				}
			}
		}
	}
	return false
}

func isUsingList(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindVariableDeclarationList &&
		(ast.IsVarUsing(node) || ast.IsVarAwaitUsing(node))
}
