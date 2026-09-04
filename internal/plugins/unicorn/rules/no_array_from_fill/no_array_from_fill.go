package no_array_from_fill

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "no-array-from-fill"

var noArrayFromFillMessage = rule.RuleMessage{
	Id:          messageID,
	Description: "Use the `Array.from(…, mapFunction)` argument instead of chaining `.fill()`.",
}

// NoArrayFromFillRule disallows calling .fill() directly on an array created
// by Array.from({length: ...}).
//
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/no-array-from-fill.js
var NoArrayFromFillRule = rule.Rule{
	Name:   "unicorn/no-array-from-fill",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		oneArgument := 1
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				fillCall, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
					Method:              "fill",
					MaximumArguments:    &oneArgument,
					RejectSpreadElement: true,
				})
				if !ok || !isArrayFromLengthCall(ctx, fillCall.Object, &oneArgument) {
					return
				}

				ctx.ReportNode(fillCall.Property, noArrayFromFillMessage)
			},
		}
	},
}

func isArrayFromLengthCall(ctx rule.RuleContext, node *ast.Node, oneArgument *int) bool {
	node = utils.ESTreeRuntimeExpression(node)
	fromCall, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
		Method:              "from",
		ArgumentsLength:     oneArgument,
		RejectSpreadElement: true,
	})
	if !ok {
		return false
	}

	arrayIdentifier := utils.ESTreeRuntimeExpression(fromCall.Object)
	if arrayIdentifier == nil || !ast.IsIdentifier(arrayIdentifier) ||
		arrayIdentifier.AsIdentifier().Text != "Array" ||
		!ctx.Globals.Access("Array").IsDeclared() ||
		!unicornutil.IsGlobalReference(ctx, arrayIdentifier) {
		return false
	}

	argument := utils.ESTreeRuntimeExpression(fromCall.Call.Arguments()[0])
	if argument == nil || !ast.IsObjectLiteralExpression(argument) {
		return false
	}

	properties := argument.AsObjectLiteralExpression().Properties
	return properties != nil && len(properties.Nodes) == 1 && isLengthProperty(properties.Nodes[0])
}

func isLengthProperty(property *ast.Node) bool {
	name, _, ok := unicornutil.ObjectDataProperty(property)
	if !ok {
		return false
	}
	switch name.Kind {
	case ast.KindIdentifier:
		return name.AsIdentifier().Text == "length"
	case ast.KindStringLiteral:
		return name.AsStringLiteral().Text == "length"
	default:
		return false
	}
}
