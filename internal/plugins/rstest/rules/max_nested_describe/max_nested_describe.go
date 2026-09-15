package max_nested_describe

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/max_nested_describe"
)

var MaxNestedDescribeRule = shared.NewRule(shared.Config{
	Name: "rstest/max-nested-describe",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		depth := rstestUtils.GetRstestDescribeDepth(ctx, analysis)
		return shared.Runtime{
			IsDescribeCall: func(node *ast.Node) bool {
				parsed := analysis.ParseFnCall(node)
				return parsed != nil && parsed.Kind == testFramework.FnKindDescribe
			},
			DescribeDepth: depth.Depth,
		}
	},
})
