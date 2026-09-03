package no_array_front_mutation

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const messageID = "no-array-front-mutation"

var ignoredUnshiftCallees = []string{
	"stream.unshift",
	"this.unshift",
	"this.stream.unshift",
	"process.stdin.unshift",
	"process.stdout.unshift",
	"process.stderr.unshift",
}

func mutationMessage(method string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageID,
		Description: "Avoid front-of-array mutation with `Array#" + method + "()`.",
		Data: map[string]string{
			"method": method,
		},
	}
}

// NoArrayFrontMutationRule disallows Array#shift() and Array#unshift(), whose
// front-of-array mutation can be unexpectedly expensive for large arrays.
//
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/no-array-front-mutation.js
var NoArrayFrontMutationRule = rule.Rule{
	Name:   "unicorn/no-array-front-mutation",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
					Methods:             []string{"shift", "unshift"},
					AllowOptionalMember: true,
				})
				if !ok {
					return
				}

				method := call.Property.AsIdentifier().Text
				if method == "unshift" && isIgnoredUnshiftCallee(call.Callee) {
					return
				}
				if unicornutil.ShouldSkipKnownNonArrayReceiver(ctx, call.Object) {
					return
				}

				ctx.ReportNode(call.Property, mutationMessage(method))
			},
		}
	},
}

func isIgnoredUnshiftCallee(callee *ast.Node) bool {
	for _, path := range ignoredUnshiftCallees {
		if unicornutil.NodeMatchesPath(callee, path) {
			return true
		}
	}
	return false
}
