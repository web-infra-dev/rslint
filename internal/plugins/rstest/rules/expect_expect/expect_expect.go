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
	ClassifyTest: func(node *ast.Node, ctx rule.RuleContext) shared.TestClassification {
		parsed := rstestUtils.ParseRstestFnCallWithOfficialExtensions(node, ctx)
		if parsed == nil || parsed.Kind != rstestUtils.RstestFnTypeTest {
			return shared.TestClassification{}
		}
		// Todo is a semantic field that survives const aliases; the call-site
		// Members can be empty for `const t = test.todo; t('x')`, so never derive
		// the todo exemption from Members here.
		return shared.TestClassification{IsTest: true, IsTodo: parsed.Todo}
	},
	ResolveNamedCallback: func(callExpr *ast.CallExpression, ctx rule.RuleContext) shared.NamedCallback {
		declNode, name := rstestUtils.ResolveNamedRstestCallback(ctx, callExpr)
		return shared.NamedCallback{DeclarationNode: declNode, Name: name}
	},
})
