package no_watch

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/effector/effectorutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var restrictedMessage = rule.RuleMessage{
	Id:          "restricted",
	Description: "Using `.watch` method leads to imperative code. Replace it with an operator `sample` or use the `target` parameter of `sample` operator.",
}

var NoWatchRule = rule.Rule{
	Name:             "effector/no-watch",
	RequiresTypeInfo: true,
	Schema:           rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				object := watchCallObject(node)
				if object == nil {
					return
				}
				if !effectorutil.IsUnit(ctx.TypeChecker.GetTypeAtLocation(object), ctx.Program()) {
					return
				}
				ctx.ReportNode(node, restrictedMessage)
			},
		}
	},
}

// watchCallObject mirrors the upstream selector
// `CallExpression[callee.type="MemberExpression"][callee.property.name="watch"]`
// and returns the member object, or nil when the call does not match.
//
// ESTree drops parentheses, so `(unit.watch)(fn)` still has a MemberExpression
// callee, while `(unit?.watch)(fn)` has a ChainExpression callee and does not
// match. `property.name` exists on Identifier and PrivateIdentifier, so a
// computed identifier key (`unit[watch]`) matches and a string key
// (`unit["watch"]`) does not.
func watchCallObject(node *ast.Node) *ast.Node {
	callee := utils.ESTreeCallCallee(node.AsCallExpression().Expression)
	object, property := utils.MemberExpressionParts(callee)
	if object == nil || property == nil {
		return nil
	}
	property = utils.ESTreeRuntimeExpression(property)
	switch property.Kind {
	case ast.KindIdentifier:
		if property.Text() != "watch" {
			return nil
		}
	case ast.KindPrivateIdentifier:
		if property.Text() != "#watch" {
			return nil
		}
	default:
		return nil
	}
	return utils.ESTreeRuntimeExpression(object)
}
