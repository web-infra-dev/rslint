package no_interpolation_in_snapshots

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

func buildNoInterpolationMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noInterpolation",
		Description: "Do not use string interpolation inside of snapshots",
	}
}

func isStringSnapshotArgument(ctx rule.RuleContext, node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	if internalUtils.IsStringLiteralOrTemplate(node) {
		return true
	}
	if ctx.TypeChecker == nil {
		return false
	}

	t := internalUtils.GetConstrainedTypeAtLocation(ctx.TypeChecker, node)
	return internalUtils.Every(internalUtils.UnionTypeParts(t), func(part *checker.Type) bool {
		return internalUtils.IsTypeFlagSet(part, checker.TypeFlagsStringLike)
	})
}

func getInlineSnapshotArgument(ctx rule.RuleContext, matcherName string, call *ast.Node) *ast.Node {
	if call == nil {
		return nil
	}
	arguments := call.Arguments()
	if len(arguments) == 0 {
		return nil
	}

	switch matcherName {
	case "toThrowErrorMatchingInlineSnapshot":
		return arguments[0]
	case "toMatchInlineSnapshot":
		// Rstest decides between (snapshot, message) and
		// (properties, snapshot, message) from the first argument's runtime
		// type.
		if len(arguments) == 1 || isStringSnapshotArgument(ctx, arguments[0]) {
			return arguments[0]
		}
		return arguments[1]
	default:
		return nil
	}
}

var NoInterpolationInSnapshotsRule = rule.Rule{
	Name:   "rstest/no-interpolation-in-snapshots",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		analysis := rstestUtils.NewRstestCallAnalysis(ctx)

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := analysis.ParseExpectCall(node)
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
					// Custom messages follow the snapshot argument and are never
					// rewritten into the source file, so only inspect the argument
					// selected by the matcher's overload.
					arg := getInlineSnapshotArgument(ctx, matcher.Name, call)
					if arg == nil {
						continue
					}
					arg = ast.SkipParentheses(arg)
					if arg.Kind == ast.KindTemplateExpression {
						ctx.ReportNode(arg, buildNoInterpolationMessage())
					}
				}
			},
		}
	},
}
