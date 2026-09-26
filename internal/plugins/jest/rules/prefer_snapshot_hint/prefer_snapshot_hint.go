package prefer_snapshot_hint

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_snapshot_hint"
)

var PreferSnapshotHintRule = shared.NewRule(shared.Config{
	Name: "jest/prefer-snapshot-hint",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := utils.GetJestCallAnalysis(ctx)
		return shared.Runtime{
			RegistrationCallbacks: analysis.RegistrationCallbacks(),
			IsRegistration: func(node *ast.Node) bool {
				parsed := analysis.ParseFnCall(node)
				return parsed != nil && (parsed.Kind == utils.JestFnTypeTest || parsed.Kind == utils.JestFnTypeDescribe)
			},
			Snapshots: func(node *ast.Node) []shared.Snapshot {
				parsed := analysis.ParseExpectCall(node)
				if parsed == nil || parsed.MatcherEntry == nil ||
					(parsed.Matcher != "toMatchSnapshot" && parsed.Matcher != "toThrowErrorMatchingSnapshot") {
					return nil
				}
				return []shared.Snapshot{{Matcher: parsed.MatcherEntry.Node, Args: node.Arguments(), Properties: parsed.Matcher == "toMatchSnapshot"}}
			},
		}
	},
})
