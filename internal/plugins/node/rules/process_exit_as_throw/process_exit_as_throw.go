package process_exit_as_throw

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// This rule changes code-path completion; diagnostics belong to its consumers.
var ProcessExitAsThrowRule = rule.Rule{
	Name:       "node/process-exit-as-throw",
	Schema:     rule.EmptyArraySchema,
	CallThrows: isProcessExit,
	Run:        func(rule.RuleContext, []any) rule.RuleListeners { return nil },
}

func isProcessExit(node *ast.Node) bool {
	if node.Kind != ast.KindCallExpression {
		return false
	}
	callee := utils.ESTreeCallCallee(node.AsCallExpression().Expression)
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	member := callee.AsPropertyAccessExpression()
	object := utils.ESTreeRuntimeExpression(member.Expression)
	return object != nil && object.Kind == ast.KindIdentifier && object.Text() == "process" &&
		member.Name().Kind == ast.KindIdentifier && member.Name().Text() == "exit"
}
