package no_restricted_matchers

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_restricted_matchers"
)

var NoRestrictedMatchersRule = shared.NewRule(shared.Config{
	Name: "rstest/no-restricted-matchers",
	Modifiers: map[string]bool{
		"not":      true,
		"rejects":  true,
		"resolves": true,
	},
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return shared.Runtime{
			Parse: func(node *ast.Node) *shared.ExpectChain {
				parsed := analysis.ParseExpectCall(node)
				if parsed == nil || len(parsed.Matchers) == 0 {
					return nil
				}

				firstMatcher := &parsed.Matchers[0].Entry
				for i := range parsed.MemberEntries {
					entry := &parsed.MemberEntries[i]
					if entry.Node != firstMatcher.Node {
						continue
					}
					return &shared.ExpectChain{
						Names:   parsed.Members[:i+1],
						Entries: parsed.MemberEntries[:i+1],
					}
				}
				return nil
			},
		}
	},
})
