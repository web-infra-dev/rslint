package prefer_snapshot_hint

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_snapshot_hint"
)

var PreferSnapshotHintRule = shared.NewRule(shared.Config{
	Name:                   "rstest/prefer-snapshot-hint",
	AllowInterpolatedHints: true,
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := utils.GetRstestCallAnalysis(ctx)
		return shared.Runtime{
			RegistrationCallbacks: analysis.RegistrationCallbacks(),
			IsRegistration: func(node *ast.Node) bool {
				parsed := analysis.ParseFnCall(node)
				return parsed != nil && (parsed.Kind == utils.RstestFnTypeTest || parsed.Kind == utils.RstestFnTypeDescribe)
			},
			Snapshots: func(node *ast.Node) []shared.Snapshot {
				parsed := analysis.ParseExpectCall(node)
				if parsed == nil || parsed.Head == nil ||
					parsed.Entry == utils.RstestExpectEntryPoll || parsed.Entry == utils.RstestExpectEntryElement {
					return nil
				}
				var snapshots []shared.Snapshot
				for _, matcher := range parsed.Matchers {
					if matcher.Negated || matcher.Entry.Call == nil ||
						(matcher.Name != "toMatchSnapshot" && matcher.Name != "matchSnapshot" && matcher.Name != "toThrowErrorMatchingSnapshot") {
						continue
					}
					snapshots = append(snapshots, shared.Snapshot{
						Matcher:    matcher.Entry.Node,
						Args:       matcher.Entry.Call.Arguments(),
						Properties: matcher.Name != "toThrowErrorMatchingSnapshot",
					})
				}
				return snapshots
			},
		}
	},
})
