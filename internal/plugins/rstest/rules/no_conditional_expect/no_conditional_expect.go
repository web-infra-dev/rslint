package no_conditional_expect

import (
	"github.com/microsoft/typescript-go/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_conditional_expect"
)

func sourceMayContainConditionalRstestExpect(sourceFile *ast.SourceFile) bool {
	if sourceFile == nil || sourceFile.Identifiers == nil {
		return true
	}
	_, ok := sourceFile.Identifiers["expect"]
	return ok
}

var NoConditionalExpectRule = shared.NewRule(shared.Config{
	Name: "rstest/no-conditional-expect",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		if !sourceMayContainConditionalRstestExpect(ctx.SourceFile) {
			return shared.Runtime{Skip: true}
		}
		callbacks := rstestUtils.CollectRstestTestCallbacks(ctx)
		return shared.Runtime{
			TestCallbackFunctions: callbacks.Functions,
			ClassifyCall: func(node *ast.Node, checkExpect bool) (bool, bool) {
				return callbacks.TestCalls[node],
					checkExpect &&
						callbacks.IsExpectCandidate(node) &&
						rstestUtils.IsRstestExpectCall(node, ctx, callbacks)
			},
		}
	},
})
