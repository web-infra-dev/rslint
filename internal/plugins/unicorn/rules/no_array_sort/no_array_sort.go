// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
package no_array_sort

import (
	_ "embed"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_array_sort.schema.json
var schemaJSON []byte

var NoArraySortRule = rule.Rule{
	Name:   "unicorn/no-array-sort",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		allowStatement := unicornutil.AllowArrayMutationStatement(options)
		maximumArguments := 1
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
					Method:              "sort",
					MaximumArguments:    &maximumArguments,
					RejectSpreadElement: true,
					AllowOptionalMember: true,
				})
				if !ok {
					return
				}
				if args := node.Arguments(); len(args) > 0 && unicornutil.IsNodeValueNotFunction(utils.ESTreeRuntimeExpression(args[0])) {
					return
				}
				unicornutil.ReportArrayMutation(ctx, call, "toSorted", allowStatement)
			},
		}
	},
}
