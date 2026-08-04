package no_conditional_expect

import (
	"github.com/microsoft/typescript-go/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_conditional_expect"
)

var NoConditionalExpectRule = shared.NewRule(shared.Config{
	Name: "rstest/no-conditional-expect",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		callbacks := rstestUtils.CollectRstestTestCallbacks(ctx)
		return shared.Runtime{
			TestCallbackFunctions: callbacks.Functions,
			ClassifyCall: func(node *ast.Node) (bool, bool) {
				parsed := rstestUtils.ParseRstestFnCallWithOfficialExtensions(node, ctx)
				return parsed != nil && parsed.Kind == rstestUtils.RstestFnTypeTest,
					rstestUtils.IsRstestExpectCall(node, ctx, callbacks)
			},
		}
	},
})
