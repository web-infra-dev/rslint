package expect_expect

import (
	"github.com/microsoft/typescript-go/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/expect_expect"
)

var ExpectExpectRule = shared.NewRule(shared.Config{
	Name: "rstest/expect-expect",
	// Rstest ships both `expect` and `assert` (chai) as globals
	// (packages/core/src/utils/constants.ts globalApiList), so the default
	// asserting functions match eslint-plugin-vitest rather than jest.
	DefaultAssertFunctionNames: []string{"expect", "assert"},
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return shared.Runtime{
			ClassifyTest: func(node *ast.Node) shared.TestClassification {
				parsed := analysis.ParseFnCall(node)
				if parsed == nil || parsed.Kind != rstestUtils.RstestFnTypeTest {
					return shared.TestClassification{}
				}
				// Todo is a semantic field that survives const aliases; the
				// call-site Members can be empty for `const t = test.todo;
				// t('x')`, so never derive the exemption from Members here.
				return shared.TestClassification{IsTest: true, IsTodo: parsed.Todo}
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
