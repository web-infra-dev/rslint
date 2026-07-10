package no_this_in_sfc

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// isPropertyOwnedSFC mirrors upstream's
// `component.node.parent.type === 'Property'` filter. ESTree wraps every
// object-literal entry in a `Property` node, so any FunctionExpression /
// ArrowFunction / method serving as a property value has parent `Property`.
//
// In tsgo there is no `Property` wrapper:
//   - `{ Foo() {} }`         → MethodDeclaration directly inside ObjectLiteralExpression
//   - `{ Foo: function() }`  → FunctionExpression as the initializer of a PropertyAssignment
//   - `{ Foo: () => {} }`    → ArrowFunction as the initializer of a PropertyAssignment
//
// All three correspond to upstream's "Property" filter and must skip reporting.
//
// ParenthesizedExpression wrappers between a FunctionExpression / ArrowFunction
// and its PropertyAssignment owner are transparent — tsgo preserves them while
// ESTree flattens them — so `{ Foo: (function() {...}) }` still hits the
// carve-out. MethodDeclaration cannot be paren-wrapped (illegal syntax), so
// the walk only matters for the FE/Arrow path.
func isPropertyOwnedSFC(component *ast.Node) bool {
	parent := component.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent == nil {
		return false
	}
	switch component.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		return parent.Kind == ast.KindObjectLiteralExpression
	}
	return parent.Kind == ast.KindPropertyAssignment
}

var NoThisInSfcRule = rule.Rule{
	Name: "react/no-this-in-sfc",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		pragma := reactutil.GetReactPragma(ctx.Settings)
		wrappers := reactutil.GetComponentWrapperFunctions(ctx.Settings, pragma)

		report := func(node *ast.Node) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "noThisInSFC",
				Description: "Stateless functional components should not use `this`",
			})
		}

		// check fires on a member-access node whose object position must be
		// inspected for `this`. Both PropertyAccessExpression (`this.x`) and
		// ElementAccessExpression (`this['x']`) reach here, matching ESTree's
		// unified `MemberExpression` listener. ParenthesizedExpression wrappers
		// around the object are transparent (tsgo preserves; ESTree flattens).
		check := func(node, expr *ast.Node) {
			if ast.SkipParentheses(expr).Kind != ast.KindThisKeyword {
				return
			}
			component := reactutil.GetParentStatelessComponent(node, pragma, wrappers)
			if component == nil {
				return
			}
			if isPropertyOwnedSFC(component) {
				return
			}
			report(node)
		}

		return rule.RuleListeners{
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				check(node, node.AsPropertyAccessExpression().Expression)
			},
			ast.KindElementAccessExpression: func(node *ast.Node) {
				check(node, node.AsElementAccessExpression().Expression)
			},
		}
	},
}
