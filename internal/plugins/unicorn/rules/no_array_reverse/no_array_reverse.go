// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
package no_array_reverse

import (
	_ "embed"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed no_array_reverse.schema.json
var schemaJSON []byte

var NoArrayReverseRule = rule.Rule{
	Name:   "unicorn/no-array-reverse",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		allowStatement := unicornutil.AllowArrayMutationStatement(options)
		maximumArguments := 0
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
					Method:              "reverse",
					MaximumArguments:    &maximumArguments,
					RejectSpreadElement: true,
					AllowOptionalMember: true,
				})
				if !ok {
					return
				}

				unicornutil.ReportArrayMutation(ctx, call, "toReversed", allowStatement)
			},
		}
	},
}
