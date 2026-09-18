package no_unneeded_async_expect_function

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_unneeded_async_expect_function"
)

var NoUnneededAsyncExpectFunctionRule = shared.NewRule(shared.Config{
	Name: "rstest/no-unneeded-async-expect-function",
	ReportModifiers: map[string]bool{
		"resolves": true,
	},
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return shared.Runtime{ParseExpectCall: func(node *ast.Node) *shared.ExpectCall {
			parsed := analysis.ParseExpectCall(node)
			if parsed == nil ||
				parsed.Reason != rstestUtils.RstestExpectParseReasonNone ||
				parsed.Head == nil ||
				(parsed.Entry != rstestUtils.RstestExpectEntryCall &&
					parsed.Entry != rstestUtils.RstestExpectEntrySoft) {
				return nil
			}
			return &shared.ExpectCall{Head: parsed.Head, Modifiers: parsed.Modifiers}
		}}
	},
})
