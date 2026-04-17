package guard_for_in

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/guard-for-in
var GuardForInRule = rule.Rule{
	Name: "guard-for-in",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindForInStatement: func(node *ast.Node) {
				if isGuarded(node.AsForInOrOfStatement().Statement) {
					return
				}
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "wrap",
					Description: "The body of a for-in should be wrapped in an if statement to filter unwanted properties from the prototype.",
				})
			},
		}
	},
}

// isGuarded mirrors ESLint's accepted shapes for a for-in body: an empty/if
// statement, an empty block, a block whose first statement is an `if`, and the
// `if (...) continue;` short-circuit (direct or wrapped in a single-statement
// block) that filters unwanted keys before the rest of the body runs.
func isGuarded(body *ast.Node) bool {
	if body == nil {
		return false
	}
	switch body.Kind {
	case ast.KindEmptyStatement, ast.KindIfStatement:
		return true
	case ast.KindBlock:
		statements := body.AsBlock().Statements.Nodes
		if len(statements) == 0 {
			return true
		}
		first := statements[0]
		if first.Kind != ast.KindIfStatement {
			return false
		}
		if len(statements) == 1 {
			return true
		}
		return isContinueGuard(first.AsIfStatement().ThenStatement)
	}
	return false
}

// isContinueGuard reports whether `consequent` is a `continue` (or a block
// containing only a `continue`), which means later body statements run only for
// keys that pass the `if` check.
func isContinueGuard(consequent *ast.Node) bool {
	if consequent == nil {
		return false
	}
	if consequent.Kind == ast.KindContinueStatement {
		return true
	}
	if consequent.Kind == ast.KindBlock {
		statements := consequent.AsBlock().Statements.Nodes
		return len(statements) == 1 && statements[0].Kind == ast.KindContinueStatement
	}
	return false
}
