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

type stackFrame struct {
	node  *ast.Node
	count int
}

// https://eslint.org/docs/latest/rules/require-yield
var RequireYieldRule = rule.Rule{
	Name: "require-yield",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		stack := make([]stackFrame, 0, 8)

		enter := func(node *ast.Node) {
			stack = append(stack, stackFrame{node: node})
		}

		exit := func(node *ast.Node) {
			n := len(stack)
			if n == 0 {
				return
			}
			top := stack[n-1]
			stack = stack[:n-1]
			// PropertyDeclaration initializer and ClassStaticBlockDeclaration
			// both represent implicit-constructor-like scopes that are never
			// themselves generators.
			if top.node.Kind == ast.KindPropertyDeclaration ||
				top.node.Kind == ast.KindClassStaticBlockDeclaration {
				return
			}
			if isGenerator(top.node) && top.count == 0 && utils.HasNonEmptyFunctionBody(top.node) {
				ctx.ReportRange(
					utils.GetFunctionHeadLoc(ctx.SourceFile, top.node),
					buildMissingYieldMessage(),
				)
			}
		}

		countYield := func(node *ast.Node) {
			for i := len(stack) - 1; i >= 0; i-- {
				bp, be, ok := utils.BodyLikeRange(stack[i].node)
				if !ok {
					continue
				}
				if node.Pos() >= bp && node.End() <= be {
					stack[i].count++
					return
				}
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration:                              enter,
			rule.ListenerOnExit(ast.KindFunctionDeclaration):         exit,
			ast.KindFunctionExpression:                               enter,
			rule.ListenerOnExit(ast.KindFunctionExpression):          exit,
			ast.KindMethodDeclaration:                                enter,
			rule.ListenerOnExit(ast.KindMethodDeclaration):           exit,
			ast.KindArrowFunction:                                    enter,
			rule.ListenerOnExit(ast.KindArrowFunction):               exit,
			ast.KindGetAccessor:                                      enter,
			rule.ListenerOnExit(ast.KindGetAccessor):                 exit,
			ast.KindSetAccessor:                                      enter,
			rule.ListenerOnExit(ast.KindSetAccessor):                 exit,
			ast.KindConstructor:                                      enter,
			rule.ListenerOnExit(ast.KindConstructor):                 exit,
			ast.KindPropertyDeclaration:                              enter,
			rule.ListenerOnExit(ast.KindPropertyDeclaration):         exit,
			ast.KindClassStaticBlockDeclaration:                      enter,
			rule.ListenerOnExit(ast.KindClassStaticBlockDeclaration): exit,

			ast.KindYieldExpression: countYield,
		}
	},
}
