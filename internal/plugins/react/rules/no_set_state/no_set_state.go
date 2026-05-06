package no_set_state

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoSetStateRule mirrors eslint-plugin-react's `no-set-state`: report any
// `this.setState(...)` call that lexically sits inside a detected React
// component (ES6 class extending Component / PureComponent, an
// `createReactClass` object literal, or a stateless functional component).
//
// Upstream uses `Components.detect` + a deferred `Program:exit` pass that
// walks `components.list()` and reports each accumulated `setStateUsage`.
// We collapse that into one synchronous report per CallExpression — the
// observable contract is identical because:
//
//   - upstream's `components.list()` only yields nodes its detection logic
//     classified as components;
//   - `GetEnclosingReactComponentOrStateless` mirrors that priority
//     (getParentES6Component | getParentES5Component | getParentStatelessComponent);
//   - upstream's `components.set(node, ...)` walks `node.parent` until it
//     finds a node already in the registry, then merges. That's the same
//     enclosing-walk we do here — no later observation can flip a
//     component into "not a component", so there's no need to defer.
var NoSetStateRule = rule.Rule{
	Name: "react/no-set-state",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		pragma := reactutil.GetReactPragma(ctx.Settings)
		createClass := reactutil.GetReactCreateClass(ctx.Settings)
		wrappers := reactutil.GetComponentWrapperFunctions(ctx.Settings, pragma)

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				// Upstream: `callee.type !== 'MemberExpression'` short-circuits.
				// SkipParentheses unwraps `(this.setState)({})` so we still see
				// the underlying PropertyAccessExpression (tsgo preserves parens
				// that ESTree flattens).
				callee := ast.SkipParentheses(call.Expression)
				if callee.Kind != ast.KindPropertyAccessExpression {
					return
				}
				prop := callee.AsPropertyAccessExpression()
				// Upstream: `callee.object.type !== 'ThisExpression'`. Wrappers
				// like `(this as any)`, `this!`, `this satisfies T`, and
				// `cond ? this : x` all break this check — receivers other
				// than a bare ThisKeyword (modulo parens) never match.
				if ast.SkipParentheses(prop.Expression).Kind != ast.KindThisKeyword {
					return
				}
				// Upstream: `callee.property.name !== 'setState'`.
				// `EsTreeName` returns "" for non-name shapes (computed
				// keys, etc.) — the equality check naturally rejects those
				// without extra guards. PrivateIdentifier's spec-defined
				// `.name` strips the leading `#`, so a private
				// `this.#setState(...)` matches just like upstream.
				if reactutil.EsTreeName(prop.Name()) != "setState" {
					return
				}

				// Upstream: `components.get(utils.getParentComponent(node))`.
				// The result is non-null only when the enclosing scope is a
				// detected component.
				if reactutil.GetEnclosingReactComponentOrStateless(node, pragma, createClass, wrappers) == nil {
					return
				}

				// Upstream reports on `setStateUsage`, which is the callee
				// MemberExpression (`this.setState`), not the full
				// CallExpression. Reporting on the unwrapped callee mirrors
				// that — when the callee itself is parenthesized (e.g.
				// `(this.setState)({})`), the column lands at `this`, after
				// the stripped `(`.
				ctx.ReportNode(callee, rule.RuleMessage{
					Id:          "noSetState",
					Description: "Do not use setState",
				})
			},
		}
	},
}
