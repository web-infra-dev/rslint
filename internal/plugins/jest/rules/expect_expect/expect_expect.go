package expect_expect

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/expect_expect"
)

func isTodoTestCall(jestFn *utils.ParsedJestFnCall) bool {
	if jestFn == nil || jestFn.Kind != utils.JestFnTypeTest {
		return false
	}
	return len(jestFn.Members) > 0 && jestFn.Members[len(jestFn.Members)-1] == "todo"
}

var ExpectExpectRule = shared.NewRule(shared.Config{
	Name:                       "jest/expect-expect",
	DefaultAssertFunctionNames: []string{"expect"},
	ClassifyTest: func(node *ast.Node, ctx rule.RuleContext) shared.TestClassification {
		jestFn := utils.ParseJestFnCall(node, ctx)
		if jestFn == nil || jestFn.Kind != utils.JestFnTypeTest {
			return shared.TestClassification{}
		}
		return shared.TestClassification{IsTest: true, IsTodo: isTodoTestCall(jestFn)}
	},
	ResolveNamedCallback: func(callExpr *ast.CallExpression, ctx rule.RuleContext) shared.NamedCallback {
		declNode, name := utils.ResolveNamedFunctionCallback(ctx, callExpr)
		return shared.NamedCallback{DeclarationNode: declNode, Name: name}
	},
})
