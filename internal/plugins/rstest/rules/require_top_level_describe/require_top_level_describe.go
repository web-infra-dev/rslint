package require_top_level_describe

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/require_top_level_describe"
)

var RequireTopLevelDescribeRule = shared.NewRule(shared.Config{
	Name: "rstest/require-top-level-describe",
	Messages: shared.Messages{
		UnexpectedTestCase: "All test cases must be wrapped in a describe block",
		UnexpectedHook:     "All hooks must be wrapped in a describe block",
	},
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		depth := rstestUtils.GetRstestDescribeDepth(ctx, analysis)
		return shared.Runtime{
			Parse: func(node *ast.Node) *shared.ParsedCall {
				parsed := analysis.ParseFnCall(node)
				if parsed == nil {
					return nil
				}
				return &parsed.ParsedCall
			},
			DescribeDepth:  depth.Depth,
			InsideDescribe: depth.InsideDescribe,
		}
	},
})
