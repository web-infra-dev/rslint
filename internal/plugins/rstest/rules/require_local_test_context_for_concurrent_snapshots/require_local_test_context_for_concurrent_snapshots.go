package require_local_test_context_for_concurrent_snapshots

import (
	"github.com/microsoft/typescript-go/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func requireLocalTestContextMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "requireLocalTestContext",
		Description: "Use local Test Context instead",
	}
}

var RequireLocalTestContextForConcurrentSnapshotsRule = rule.Rule{
	Name:   "rstest/require-local-test-context-for-concurrent-snapshots",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		// The ownership index behind this context is built on the first query,
		// so obtaining it here costs nothing on files without snapshots.
		concurrentContext := rstestUtils.GetRstestConcurrentContext(ctx, analysis)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := analysis.ParseExpectCall(node)
				if parsed == nil || parsed.FromTestContext ||
					!rstestUtils.RSTEST_SNAPSHOT_MATCHERS[parsed.Matcher] ||
					!concurrentContext.IsInConcurrentTest(node) {
					return
				}
				ctx.ReportNode(parsed.Expression, requireLocalTestContextMessage())
			},
		}
	},
}
