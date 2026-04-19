package require_yield

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildMissingYieldMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingYield",
		Description: "This generator function does not have 'yield'.",
	}
}

func isGenerator(node *ast.Node) bool {
	return ast.GetFunctionFlags(node)&ast.FunctionFlagsGenerator != 0
}

func hasNonEmptyBody(node *ast.Node) bool {
	body := node.Body()
	if body == nil || body.Kind != ast.KindBlock {
		return false
	}
	block := body.AsBlock()
	if block == nil || block.Statements == nil {
		return false
	}
	return len(block.Statements.Nodes) > 0
}

// https://eslint.org/docs/latest/rules/require-yield
var RequireYieldRule = rule.Rule{
	Name: "require-yield",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		stack := make([]int, 0, 8)

		beginChecking := func(node *ast.Node) {
			if isGenerator(node) {
				stack = append(stack, 0)
			}
		}

		endChecking := func(node *ast.Node) {
			if !isGenerator(node) {
				return
			}
			n := len(stack)
			if n == 0 {
				return
			}
			count := stack[n-1]
			stack = stack[:n-1]

			if count == 0 && hasNonEmptyBody(node) {
				ctx.ReportRange(
					utils.GetFunctionHeadLoc(ctx.SourceFile, node),
					buildMissingYieldMessage(),
				)
			}
		}

		incYield := func(node *ast.Node) {
			if n := len(stack); n > 0 {
				stack[n-1]++
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration:                      beginChecking,
			rule.ListenerOnExit(ast.KindFunctionDeclaration): endChecking,
			ast.KindFunctionExpression:                       beginChecking,
			rule.ListenerOnExit(ast.KindFunctionExpression):  endChecking,
			ast.KindMethodDeclaration:                        beginChecking,
			rule.ListenerOnExit(ast.KindMethodDeclaration):   endChecking,

			ast.KindYieldExpression: incYield,
		}
	},
}
