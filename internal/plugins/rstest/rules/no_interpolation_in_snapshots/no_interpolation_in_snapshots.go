package no_interpolation_in_snapshots

import (
	"github.com/microsoft/typescript-go/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildNoInterpolationMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noInterpolation",
		Description: "Do not use string interpolation inside of snapshots",
	}
}

var NoInterpolationInSnapshotsRule = rule.Rule{
	Name:   "rstest/no-interpolation-in-snapshots",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		// expect bound to a test context parameter — test("x", ({ expect }) => ...)
		// — can only be recognized through the collected callbacks.
		callbacks := rstestUtils.CollectRstestTestCallbacks(ctx)

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := rstestUtils.ParseRstestExpectCall(node, ctx, callbacks)
				if parsed == nil {
					return
				}

				// Chai permits several assertions in one chain, so every matcher is
				// checked rather than just the first one.
				for _, matcher := range parsed.Matchers {
					// RSTEST_INLINE_SNAPSHOT_MATCHERS, deliberately not
					// RSTEST_SNAPSHOT_MATCHERS: toMatchSnapshot, matchSnapshot and
					// Rstest's own toMatchFileSnapshot keep their expected value
					// outside the source file, where interpolation is legitimate —
					// toMatchFileSnapshot even takes a path as its first argument.
					if !rstestUtils.RSTEST_INLINE_SNAPSHOT_MATCHERS[matcher.Name] {
						continue
					}
					// Property-style Chai assertions carry no call and hence no
					// arguments to interpolate into.
					call := matcher.Entry.Call
					if call == nil {
						continue
					}
					// Both overloads of the inline snapshot matchers may hold the
					// snapshot in arguments[0] or arguments[1], so all are checked.
					for _, arg := range call.Arguments() {
						if arg := ast.SkipParentheses(arg); arg != nil && arg.Kind == ast.KindTemplateExpression {
							ctx.ReportNode(arg, buildNoInterpolationMessage())
						}
					}
				}
			},
		}
	},
}
