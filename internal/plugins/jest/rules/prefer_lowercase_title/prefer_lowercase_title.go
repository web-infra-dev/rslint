package prefer_lowercase_title

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_lowercase_title"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

var PreferLowercaseTitleRule = shared.NewRule(shared.Config{
	Name: "jest/prefer-lowercase-title",
	DescribeAliases: []string{
		"describe", "fdescribe", "xdescribe", "context",
	},
	TestAliases: []string{
		"test", "xtest",
	},
	ItAliases: []string{
		"it", "xit", "fit",
	},
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		return shared.Runtime{
			Parse: func(node *ast.Node) *testFramework.ParsedCall {
				parsed := jestUtils.ParseJestFnCall(node, ctx)
				if parsed == nil {
					return nil
				}
				return &parsed.ParsedCall
			},
		}
	},
})
