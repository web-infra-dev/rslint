package max_expects

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/max_expects"
)

func isCountedExpectCall(parsed *rstestUtils.ParsedRstestExpectCall) bool {
	if parsed == nil ||
		parsed.Reason != rstestUtils.RstestExpectParseReasonNone ||
		parsed.MatcherEntry == nil ||
		parsed.Head == nil {
		return false
	}

	switch parsed.Entry {
	case rstestUtils.RstestExpectEntryCall,
		rstestUtils.RstestExpectEntrySoft,
		rstestUtils.RstestExpectEntryPoll,
		rstestUtils.RstestExpectEntryElement:
		return true
	default:
		return false
	}
}

var MaxExpectsRule = shared.NewRule(shared.Config{
	Name: "rstest/max-expects",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return shared.Runtime{
			Classify: func(node *ast.Node) shared.CallKind {
				if analysis.ParseTestCall(node) != nil {
					return shared.CallTest
				}
				if parsed := analysis.ParseFnCall(node); parsed != nil &&
					parsed.Kind == rstestUtils.RstestFnTypeHook {
					return shared.CallHook
				}
				if isCountedExpectCall(analysis.ParseExpectCall(node)) {
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
