package max_expects

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/max_expects"
)

// isCountedExpectCall excludes calls through a static `expect` member, such as
// `expect.hasAssertions()`.
func isCountedExpectCall(jestFnCall *utils.ParsedJestFnCall) bool {
	if jestFnCall == nil || jestFnCall.Kind != utils.JestFnTypeExpect {
		return false
	}

	headNode := jestFnCall.Head.Local.Node
	return headNode == nil || !utils.IsMemberAccessNode(headNode.Parent)
}

var MaxExpectsRule = shared.NewRule(shared.Config{
	Name: "jest/max-expects",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := utils.GetJestCallAnalysis(ctx)
		return shared.Runtime{
			Classify: func(node *ast.Node) shared.CallKind {
				parsed := analysis.ParseFnCall(node)
				if parsed == nil {
					return shared.CallNone
				}
				switch parsed.Kind {
				case utils.JestFnTypeTest:
					return shared.CallTest
				case utils.JestFnTypeHook:
					return shared.CallHook
				}
				if isCountedExpectCall(parsed) {
					return shared.CallAssertion
				}
				return shared.CallNone
			},
			Callbacks: func() map[*ast.Node]bool {
				return analysis.Callbacks().Functions
			},
		}
	},
})
