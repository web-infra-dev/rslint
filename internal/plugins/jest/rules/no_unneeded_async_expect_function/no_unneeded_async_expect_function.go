package no_unneeded_async_expect_function

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslintUtils "github.com/web-infra-dev/rslint/internal/utils"
)

func buildNoAsyncWrapperForExpectedPromiseMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noAsyncWrapperForExpectedPromise",
		Description: "Avoid wrapping asynchronous expectations in an unnecessary async function.",
	}
}

func isAsyncFunction(node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	return node != nil &&
		ast.IsFunctionExpressionOrArrowFunction(node) &&
		ast.IsAsyncFunction(node)
}

func functionBody(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	node = ast.SkipParentheses(node)

	switch node.Kind {
	case ast.KindArrowFunction:
		return node.AsArrowFunction().Body
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().Body
	default:
		return nil
	}
}

func singleStatementExpression(body *ast.Node) *ast.Node {
	if body == nil || body.Kind != ast.KindBlock {
		return body
	}

	block := body.AsBlock()
	if block == nil || block.Statements == nil || len(block.Statements.Nodes) != 1 {
		return nil
	}

	stmt := block.Statements.Nodes[0]
	if stmt == nil || stmt.Kind != ast.KindExpressionStatement {
		return nil
	}

	return stmt.AsExpressionStatement().Expression
}

func getUnwrappedAwaitedExpression(fn *ast.Node) *ast.Node {
	expr := singleStatementExpression(functionBody(fn))
	if expr == nil {
		return nil
	}
	expr = ast.SkipParentheses(expr)
	if expr == nil || expr.Kind != ast.KindAwaitExpression {
		return nil
	}

	awaited := expr.AsAwaitExpression().Expression
	if awaited == nil {
		return nil
	}
	awaited = ast.SkipParentheses(awaited)
	if awaited == nil || awaited.Kind != ast.KindCallExpression {
		return nil
	}

	return awaited
}

func hasPromiseExpectModifier(jestFnCall *jestUtils.ParsedJestFnCall) bool {
	return slices.Contains(jestFnCall.Modifiers, "resolves") ||
		slices.Contains(jestFnCall.Modifiers, "rejects")
}

var NoUnneededAsyncExpectFunctionRule = rule.Rule{
	Name: "jest/no-unneeded-async-expect-function",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := jestUtils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil ||
					jestFnCall.Kind != jestUtils.JestFnTypeExpect ||
					!hasPromiseExpectModifier(jestFnCall) {
					return
				}

				expectCall := jestFnCall.Head.Local.Node.Parent
				if expectCall == nil || expectCall.Kind != ast.KindCallExpression {
					return
				}

				args := expectCall.Arguments()
				if len(args) == 0 || !isAsyncFunction(args[0]) {
					return
				}

				awaited := getUnwrappedAwaitedExpression(args[0])
				if awaited == nil {
					return
				}

				sourceFile := ctx.SourceFile
				replacement := rslintUtils.TrimmedNodeText(sourceFile, awaited)
				ctx.ReportNodeWithFixes(
					args[0],
					buildNoAsyncWrapperForExpectedPromiseMessage(),
					rule.RuleFixReplace(sourceFile, args[0], replacement),
				)
			},
		}
	},
}
