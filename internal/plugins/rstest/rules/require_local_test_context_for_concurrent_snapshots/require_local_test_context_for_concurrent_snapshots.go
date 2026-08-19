package require_local_test_context_for_concurrent_snapshots

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func isSnapshotMatcher(matcher rstestUtils.ParsedRstestExpectMatcher) bool {
	return rstestUtils.RSTEST_SNAPSHOT_MATCHERS[matcher.Name]
}

func requireLocalTestContextMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "requireLocalTestContext",
		Description: "Use local Test Context instead",
	}
}

var RequireLocalTestContextForConcurrentSnapshotsRule = rule.Rule{
	Name: "rstest/require-local-test-context-for-concurrent-snapshots",
	// The local Test Context form this rule asks for is recognized by symbol
	// identity, so without a checker every `({ expect })` callback looks like a
	// global expect and the correct code gets reported. Declaring the
	// requirement keeps the rule off source-only programs instead.
	RequiresTypeInfo: true,
	Schema:           rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		// The ownership index behind this context is built on the first query,
		// so obtaining it here costs nothing on files without snapshots.
		concurrentContext := rstestUtils.GetRstestConcurrentContext(ctx, analysis)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				// Only the outermost call of a chain is parsed, so a chain
				// carrying two snapshot assertions is answered once.
				if rstestUtils.FindTopMostCallExpression(node) != node {
					return
				}
				parsed := analysis.ParseExpectCall(node)
				if parsed == nil || parsed.FromTestContext ||
					!concurrentContext.IsInConcurrentTest(node) {
					return
				}
				// Chai permits several assertions in one chain, so every matcher
				// is considered rather than just the first one: the snapshot
				// matcher of `expect(x).to.be.a("string").and.matchSnapshot()`
				// is not the one Matcher mirrors.
				if !slices.ContainsFunc(parsed.Matchers, isSnapshotMatcher) {
					return
				}
				ctx.ReportNode(parsed.Expression, requireLocalTestContextMessage())
			},
		}
	},
}
