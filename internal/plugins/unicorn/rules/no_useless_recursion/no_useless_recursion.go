package no_useless_recursion

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
	"github.com/web-infra-dev/rslint/internal/utils/scopeanalysis"
)

var recursionMessage = rule.RuleMessage{
	Id:          "no-useless-recursion",
	Description: "Use a loop instead of this recursive function call.",
}

var NoUselessRecursionRule = rule.Rule{
	Name:   "unicorn/no-useless-recursion",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		var reassigned map[*ast.Node]struct{}
		isReassigned := func(function *ast.Node) bool {
			if reassigned == nil {
				reassigned = reassignedFunctions(ctx)
			}
			_, ok := reassigned[function]
			return ok
		}
		return rule.RuleListeners{
			ast.KindReturnStatement: func(node *ast.Node) {
				checkReturn(ctx, node, isReassigned)
			},
		}
	},
}

func checkReturn(ctx rule.RuleContext, node *ast.Node, isReassigned func(*ast.Node) bool) {
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
	if symbol == nil || !slices.Contains(symbol.Declarations, function) || hasTryWithOrActiveUsing(node, function) {
		return
	}
	// A named function expression has a private, immutable name.
	if function.Kind == ast.KindFunctionDeclaration && isReassigned(function) {
		return
	}

	ctx.ReportNode(callee, recursionMessage)
}

func reassignedFunctions(ctx rule.RuleContext) map[*ast.Node]struct{} {
	// The JavaScript scope graph groups var redeclarations with function names.
	manager := scopeanalysis.References(ctx, nil)
	reassigned := make(map[*ast.Node]struct{})
	for _, reference := range manager.References {
		if reference.IsValueReference() && utils.IsWriteReference(reference.Identifier) {
			for _, binding := range reference.Declarations {
				if binding.Kind == scope.DefFunctionName {
					reassigned[binding.DefNode] = struct{}{}
				}
			}
		}
	}
	for _, lexicalScope := range manager.Scopes {
		for _, binding := range lexicalScope.Vars {
			if binding.Kind == scope.DefFunctionName &&
				slices.ContainsFunc(lexicalScope.Declarations(binding.Name), func(declaration *scope.Variable) bool {
					return declaration.Kind == scope.DefVariable && utils.IsVariableWriteReference(declaration.ID)
				}) {
				reassigned[binding.DefNode] = struct{}{}
			}
		}
	}
	return reassigned
}

func hasTryWithOrActiveUsing(node, function *ast.Node) bool {
	for ; node != function; node = node.Parent {
		switch node.Kind {
		case ast.KindTryStatement, ast.KindWithStatement:
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
