package expect_expect

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
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
		analysis := utils.GetJestCallAnalysis(ctx)
		return shared.Runtime{
			ClassifyTest: func(node *ast.Node) shared.TestClassification {
				jestFn := analysis.ParseTestCall(node)
				if jestFn == nil {
					return shared.TestClassification{}
				}
				return shared.TestClassification{IsTest: true, IsTodo: isTodoTestCall(jestFn)}
			},
			ResolveNamedCallback: func(callNode *ast.Node) shared.NamedCallback {
				function, name := analysis.TestCallback(callNode)
				return shared.NamedCallback{
					DeclarationNode: function,
					Name:            name,
				}
			},
		}
	},
})
