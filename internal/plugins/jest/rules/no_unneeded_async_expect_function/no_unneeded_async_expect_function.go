package no_unneeded_async_expect_function

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_unneeded_async_expect_function"
)

func isAsyncNonGenerator(node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil || !ast.IsFunctionLike(node) || !ast.IsAsyncFunction(node) {
		return false
	}
	return ast.GetFunctionFlags(node)&ast.FunctionFlagsGenerator == 0 &&
		len(node.Parameters()) == 0
}

func symbolIsLocalConstAsyncArrow(ctx rule.RuleContext, symbol *ast.Symbol, use *ast.Node) bool {
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	if declaration == nil || ast.GetSourceFileOfNode(declaration) != ctx.SourceFile ||
		use == nil || declaration.Pos() >= use.Pos() {
		return false
	}

	if declaration.Kind != ast.KindVariableDeclaration || !ast.IsVarConst(declaration) {
		return false
	}
	initializer := utils.SkipAssertionsAndParens(declaration.AsVariableDeclaration().Initializer)
	return initializer != nil && initializer.Kind == ast.KindArrowFunction &&
		isAsyncNonGenerator(initializer)
}

func isKnownAsyncCall(ctx rule.RuleContext, call *shared.ExpectCall, awaited *ast.Node) bool {
	if call == nil || call.Head == nil || len(call.Head.Arguments()) == 0 ||
		!isAsyncNonGenerator(call.Head.Arguments()[0]) ||
		awaited == nil || awaited.Kind != ast.KindCallExpression ||
		ast.IsOptionalChainRoot(awaited) {
		return false
	}
	callExpression := awaited.AsCallExpression()
	if callExpression == nil || len(callExpression.Arguments.Nodes) != 0 ||
		callExpression.TypeArguments != nil {
		return false
	}
	callee := utils.SkipAssertionsAndParens(callExpression.Expression)
	if callee == nil || callee.Kind != ast.KindIdentifier || ctx.Refs == nil {
		return false
	}
	if !symbolIsLocalConstAsyncArrow(ctx, ctx.Refs.ResolveInFile(callee), callee) {
		return false
	}
	// A block wrapper discards the fulfilled value: async () => { await f() }
	// resolves to undefined even when f() resolves to another value. That also
	// changes Jest's failure output for rejects when f() unexpectedly resolves.
	// Only a concise arrow returns the inner result unchanged.
	wrapper := utils.SkipAssertionsAndParens(call.Head.Arguments()[0])
	if wrapper == nil || wrapper.Kind != ast.KindArrowFunction {
		return false
	}
	body := ast.SkipParentheses(wrapper.AsArrowFunction().Body)
	return body != nil && body.Kind == ast.KindAwaitExpression
}

var NoUnneededAsyncExpectFunctionRule = shared.NewRule(shared.Config{
	Name: "jest/no-unneeded-async-expect-function",
	ReportModifiers: map[string]bool{
		"resolves": true,
		"rejects":  true,
	},
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		return shared.Runtime{ParseExpectCall: func(node *ast.Node) *shared.ExpectCall {
			parsed := jestUtils.ParseJestFnCall(node, ctx)
			if parsed == nil || parsed.Kind != jestUtils.JestFnTypeExpect {
				return nil
			}
			head := parsed.Head.Local.Node.Parent
			if head == nil || head.Kind != ast.KindCallExpression {
				return nil
			}
			return &shared.ExpectCall{Head: head, Modifiers: parsed.Modifiers}
		}}
	},
	ShouldReportAwaitedCall: isKnownAsyncCall,
})
