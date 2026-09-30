package require_hook

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/require_hook"
)

func isJestFnCall(node *ast.Node, analysis *utils.JestCallAnalysis) bool {
	if analysis.ParseFnCall(node) != nil {
		return true
	}
	return strings.HasPrefix(testFramework.CalleeChainName(node), "jest.")
}

var RequireHookRule = shared.NewRule(shared.Config{
	Name: "jest/require-hook",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := utils.GetJestCallAnalysis(ctx)
		return shared.Runtime{
			IsFrameworkCall: func(call *ast.Node) bool {
				return isJestFnCall(call, analysis)
			},
			IsDescribe: func(call *ast.Node) bool {
				parsed := analysis.ParseFnCall(call)
				return parsed != nil && parsed.Kind == utils.JestFnTypeDescribe
			},
		}
	},
})
