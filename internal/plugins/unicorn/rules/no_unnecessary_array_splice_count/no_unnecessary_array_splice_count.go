// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
package no_unnecessary_array_splice_count

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var NoUnnecessaryArraySpliceCountRule = rule.Rule{
	Name: "unicorn/no-unnecessary-array-splice-count", Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		argumentCount := 2
		return rule.RuleListeners{ast.KindCallExpression: func(node *ast.Node) {
			call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{Methods: []string{"splice", "toSpliced"}, ArgumentsLength: &argumentCount, RejectSpreadElement: true, AllowOptionalMember: true})
			if !ok {
				return
			}
			if unicornutil.ShouldSkipKnownNonArrayReceiver(ctx, call.Object) {
				return
			}
			argumentName := "deleteCount"
			if call.Property.Text() == "toSpliced" {
				argumentName = "skipCount"
			}
			unicornutil.ReportUnnecessaryLengthArgument(ctx, call, "no-unnecessary-array-splice-count", argumentName)
		}}
	},
}
