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
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		return shared.Runtime{
			ClassifyTest: func(node *ast.Node) shared.TestClassification {
				jestFn := utils.ParseJestFnCall(node, ctx)
				if jestFn == nil || jestFn.Kind != utils.JestFnTypeTest {
					return shared.TestClassification{}
				}
				return shared.TestClassification{IsTest: true, IsTodo: isTodoTestCall(jestFn)}
			},
			ResolveNamedCallback: func(callNode *ast.Node) shared.NamedCallback {
				info := utils.ResolveTestCallbackFunction(ctx, callNode.AsCallExpression())
				return shared.NamedCallback{
					DeclarationNode: info.FunctionNode,
					Name:            info.Name,
				}
			},
		}
	},
})
