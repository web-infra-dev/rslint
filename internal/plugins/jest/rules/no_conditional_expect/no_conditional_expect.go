package no_conditional_expect

import (
	"github.com/microsoft/typescript-go/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_conditional_expect"
)

func collectTestFunctionCallbacks(ctx rule.RuleContext) map[*ast.Node]bool {
	callbacks := map[*ast.Node]bool{}
	pendingNames := map[string]bool{}

	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}

		if node.Kind == ast.KindCallExpression {
			jestFnCall := jestUtils.ParseJestFnCall(node, ctx)
			if jestFnCall != nil && jestFnCall.Kind == jestUtils.JestFnTypeTest {
				info := jestUtils.ResolveTestCallbackFunction(ctx, node.AsCallExpression())
				if info.FunctionNode != nil {
					callbacks[info.FunctionNode] = true
				} else if info.Name != "" {
					pendingNames[info.Name] = true
				}
			}
		}

		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}

	if ctx.SourceFile != nil {
		visit(ctx.SourceFile.Node.AsNode())
	}

	if len(pendingNames) > 0 {
		resolvePendingTestCallbackNames(ctx, pendingNames, callbacks)
	}

	return callbacks
}

func resolvePendingTestCallbackNames(
	ctx rule.RuleContext,
	names map[string]bool,
	callbacks map[*ast.Node]bool,
) {
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}

		switch node.Kind {
		case ast.KindFunctionDeclaration:
			fn := node.AsFunctionDeclaration()
			if fn != nil && fn.Name() != nil && names[fn.Name().Text()] {
				callbacks[node] = true
				delete(names, fn.Name().Text())
			}
		case ast.KindVariableDeclaration:
			vd := node.AsVariableDeclaration()
			if vd == nil {
				break
			}
			id := vd.Name()
			if id == nil || id.Kind != ast.KindIdentifier {
				break
			}
			name := id.AsIdentifier().Text
			if !names[name] {
				break
			}
			if vd.Initializer == nil {
				break
			}
			init := ast.SkipParentheses(vd.Initializer)
			if ast.IsFunctionExpressionOrArrowFunction(init) {
				callbacks[init] = true
				delete(names, name)
			}
		}

		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}

	if ctx.SourceFile != nil {
		visit(ctx.SourceFile.Node.AsNode())
	}
}

var NoConditionalExpectRule = shared.NewRule(shared.Config{
	Name: "jest/no-conditional-expect",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		return shared.Runtime{
			TestCallbackFunctions: collectTestFunctionCallbacks(ctx),
			ClassifyCall: func(node *ast.Node) (bool, bool) {
				parsed := jestUtils.ParseJestFnCall(node, ctx)
				return parsed != nil && parsed.Kind == jestUtils.JestFnTypeTest,
					parsed != nil && parsed.Kind == jestUtils.JestFnTypeExpect
			},
		}
	},
})
