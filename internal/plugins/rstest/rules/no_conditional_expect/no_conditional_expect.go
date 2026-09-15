package no_conditional_expect

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_conditional_expect"
)

func sourceMayContainConditionalRstestExpect(sourceFile *ast.SourceFile) bool {
	if sourceFile == nil || sourceFile.AsNode().Kind != ast.KindSourceFile {
		return true
	}
	ok := sourceFile.HasIdentifier("expect")
	return ok
}

var NoConditionalExpectRule = shared.NewRule(shared.Config{
	Name: "rstest/no-conditional-expect",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		if !sourceMayContainConditionalRstestExpect(ctx.SourceFile) {
			return shared.Runtime{Skip: true}
		}
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return shared.Runtime{
			TestCallbackFunctions: analysis.Callbacks().Functions,
			IsTestCall: func(node *ast.Node) bool {
				return analysis.ParseTestCall(node) != nil
			},
			IsExpectCall: func(node *ast.Node) bool {
				return analysis.IsExpectCall(node)
			},
		}
	},
})
