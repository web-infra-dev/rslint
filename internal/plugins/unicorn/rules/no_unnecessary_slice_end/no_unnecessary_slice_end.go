// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
package no_unnecessary_slice_end

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var NoUnnecessarySliceEndRule = rule.Rule{
	Name: "unicorn/no-unnecessary-slice-end", Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		argumentCount := 2
		return rule.RuleListeners{ast.KindCallExpression: func(node *ast.Node) {
			call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{Methods: []string{"slice"}, ArgumentsLength: &argumentCount, RejectSpreadElement: true, AllowOptionalMember: true})
			if !ok {
				return
			}
			if ctx.TypeChecker != nil &&
				unicornutil.IsKnownNonStringType(ctx, call.Object) &&
				unicornutil.IsKnownNonIndexedCollection(ctx, call.Object) {
				return
			}
			argumentName := "end"
			unicornutil.ReportUnnecessaryLengthArgument(ctx, call, "no-unnecessary-slice-end", argumentName)
		}}
	},
}
