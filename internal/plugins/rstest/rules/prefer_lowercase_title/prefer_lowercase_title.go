package prefer_lowercase_title

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_lowercase_title"
)

var PreferLowercaseTitleRule = shared.NewRule(shared.Config{
	Name: "rstest/prefer-lowercase-title",
	DescribeAliases: []string{
		"describe",
	},
	TestAliases: []string{
		"test",
	},
	ItAliases: []string{
		"it",
	},
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		depth := rstestUtils.GetRstestDescribeDepth(ctx, analysis)
		return shared.Runtime{
			Parse: func(node *ast.Node) *testFramework.ParsedCall {
				parsed := analysis.ParseFnCall(node)
				if parsed == nil {
					return nil
				}
				return &parsed.ParsedCall
			},
			IsTodo: func(node *ast.Node) bool {
				parsed := analysis.ParseFnCall(node)
				return parsed != nil && parsed.Todo
			},
			DescribeDepth: depth.Depth,
		}
	},
})
