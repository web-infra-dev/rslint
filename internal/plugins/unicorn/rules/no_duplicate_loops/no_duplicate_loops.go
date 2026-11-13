package no_duplicate_loops

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const ruleName = "no-duplicate-loops"

var NoDuplicateLoopsRule = rule.Rule{
	Name:   "unicorn/" + ruleName,
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindForOfStatement: func(node *ast.Node) {
				loop := node.AsForInOrOfStatement()
				call, ok := unicornutil.MatchMapOrFilterCall(unicornutil.UnwrapIteratorExpression(loop.Expression))
				if !ok || unicornutil.IsIteratorExpression(&ctx, call.Object) {
					return
				}

				loopKind := "for…of"
				if loop.AwaitModifier != nil {
					loopKind = "for await…of"
				}
				method := call.Property.Text()
				ctx.ReportNode(call.Property, rule.RuleMessage{
					Id:          ruleName,
					Description: "Do not use `." + method + "()` directly in a `" + loopKind + "` loop header. It creates an intermediate array before the loop.",
					Data:        map[string]string{"method": method, "loopKind": loopKind},
				})
			},
		}
	},
}
