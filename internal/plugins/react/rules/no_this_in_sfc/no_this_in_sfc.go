package no_this_in_sfc

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// getParentStatelessComponent mirrors eslint-plugin-react's
// `Components.detect`-paired helper of the same name: walk up enclosing
// FunctionLike scopes from `node` and return the first one classified as a
// stateless functional component. A non-component function does NOT stop the
// walk — the next outer FunctionLike still gets a chance, matching upstream's
// `scope.upper` traversal.
//
// ES6 class scopes / class field initializers / module scope are non-FunctionLike
// nodes that are simply skipped during the walk; `this` inside a class field
// thus resolves to the nearest enclosing function-like (matching upstream's
// `getScope` + `scope.upper` chain, which also walks past class scope without
// classifying it as a component).
func getParentStatelessComponent(node *ast.Node, pragma string) *ast.Node {
	for p := node.Parent; p != nil; p = p.Parent {
		if !ast.IsFunctionLike(p) {
			continue
		}
		if reactutil.IsStatelessReactComponent(p, pragma) {
			return p
		}
	}
	return nil
}

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
func isPropertyOwnedSFC(component *ast.Node) bool {
	parent := component.Parent
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
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		pragma := reactutil.GetReactPragma(ctx.Settings)

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
			component := getParentStatelessComponent(node, pragma)
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
