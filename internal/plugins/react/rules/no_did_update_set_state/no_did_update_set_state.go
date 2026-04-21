package no_did_update_set_state

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// methodName is the single lifecycle hook this rule targets. Upstream's
// makeNoMethodSetStateRule factory is parameterized; since
// no-did-update-set-state always passes 'componentDidUpdate', we hardcode it.
const methodName = "componentDidUpdate"

// isStopper mirrors ESLint's ancestor.type check for containers whose `key`
// can identify the target method: Property, MethodDefinition, ClassProperty,
// PropertyDefinition. In tsgo these collapse into PropertyAssignment (object
// `name: fn`), MethodDeclaration (class methods + object shorthand),
// GetAccessor / SetAccessor / Constructor, and PropertyDeclaration (class
// field). ShorthandPropertyAssignment is omitted — it has no function body,
// so setState could never appear beneath it.
//
// Decomposed as `IsMethodOrAccessor || Kind in {Constructor, PropertyAssignment,
// PropertyDeclaration}`: IsMethodOrAccessor covers MethodDeclaration /
// GetAccessor / SetAccessor, and the remaining three are the non-function
// key-bearing containers.
func isStopper(node *ast.Node) bool {
	if ast.IsMethodOrAccessor(node) {
		return true
	}
	switch node.Kind {
	case ast.KindConstructor,
		ast.KindPropertyAssignment,
		ast.KindPropertyDeclaration:
		return true
	}
	return false
}

// stopperName returns the name-string of a stopper node's key as ESLint
// would expose it via `ancestor.key.name`. Identifier and PrivateIdentifier
// both populate `.name` in ESTree — the latter without the leading `#` per
// the spec — so `class { #componentDidUpdate() { ... } }` matches upstream
// just like the public form. In tsgo, PrivateIdentifier.Text retains the
// `#`, so we strip it to align with ESLint.
//
// Non-name keys (ComputedPropertyName, StringLiteral, NumericLiteral, etc.)
// yield "" — ESLint's `key.name` is undefined for those shapes and never
// equals "componentDidUpdate".
func stopperName(node *ast.Node) string {
	n := node.Name()
	if n == nil {
		return ""
	}
	switch n.Kind {
	case ast.KindIdentifier:
		return n.AsIdentifier().Text
	case ast.KindPrivateIdentifier:
		return strings.TrimPrefix(n.AsPrivateIdentifier().Text, "#")
	}
	return ""
}

// reactVersionNoop reports whether the rule should be a no-op for the
// configured React version. Mirrors upstream `shouldBeNoop('componentDidUpdate')`
// exactly:
//
//	methodName in noops && testReactVersion(ctx, '>= 16.3.0')
//	                    && !testReactVersion(ctx, '>= 999.999.999')
//
// i.e. noop iff the configured version is in the half-open range
// [16.3.0, 999.999.999). The upper bound keeps the rule active when
// eslint-plugin-react's "version absent" default (999.999.999) or a user
// who explicitly set "999.999.999" is in effect — upstream uses the same
// sentinel to express "user left version unpinned".
//
// Our `ParseReactVersion` returns (999,999,999) for an absent setting, so
// checking `settings.react.version` is a string key lets us distinguish
// "unset" from "explicitly 999.999.999" only structurally — both cases keep
// the rule active, matching upstream.
func reactVersionNoop(settings map[string]interface{}) bool {
	if settings == nil {
		return false
	}
	rs, ok := settings["react"].(map[string]interface{})
	if !ok {
		return false
	}
	if _, ok := rs["version"].(string); !ok {
		return false
	}
	return !reactutil.ReactVersionLessThan(settings, 16, 3, 0) &&
		reactutil.ReactVersionLessThan(settings, 999, 999, 999)
}

func parseDisallowInFunc(options any) bool {
	switch v := options.(type) {
	case string:
		return v == "disallow-in-func"
	case []interface{}:
		if len(v) > 0 {
			if s, ok := v[0].(string); ok {
				return s == "disallow-in-func"
			}
		}
	}
	return false
}

var NoDidUpdateSetStateRule = rule.Rule{
	Name: "react/no-did-update-set-state",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		if reactVersionNoop(ctx.Settings) {
			return rule.RuleListeners{}
		}
		disallowInFunc := parseDisallowInFunc(options)

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				callee := ast.SkipParentheses(call.Expression)
				if callee.Kind != ast.KindPropertyAccessExpression {
					return
				}
				prop := callee.AsPropertyAccessExpression()
				if ast.SkipParentheses(prop.Expression).Kind != ast.KindThisKeyword {
					return
				}
				// Upstream: `'name' in callee.property && callee.property.name
				// === 'setState'`. Both Identifier and PrivateIdentifier
				// populate `.name` in ESTree, and per-spec PrivateIdentifier
				// strips the leading `#` — so `this.#setState()` matches
				// upstream just like `this.setState()`. tsgo retains the `#`
				// on PrivateIdentifier.Text, so we strip it before comparing.
				nameNode := prop.Name()
				if nameNode == nil {
					return
				}
				var propName string
				switch nameNode.Kind {
				case ast.KindIdentifier:
					propName = nameNode.AsIdentifier().Text
				case ast.KindPrivateIdentifier:
					propName = strings.TrimPrefix(nameNode.AsPrivateIdentifier().Text, "#")
				default:
					return
				}
				if propName != "setState" {
					return
				}

				// Walk ancestors innermost-to-outermost, mirroring ESLint's
				// `findLast(ancestors, cb)`. Depth counts function-like
				// wrappers crossed before the first matching stopper; in
				// default mode a match at depth > 1 is skipped (setState in a
				// nested callback), while `disallow-in-func` reports at any
				// depth.
				//
				// `ast.IsFunctionLikeDeclaration` returns true for
				// FunctionDeclaration / FunctionExpression / ArrowFunction /
				// MethodDeclaration / Constructor / GetAccessor / SetAccessor —
				// exactly the set ESLint's /Function(Expression|Declaration)$/
				// regex matches, extended to the method/accessor/constructor
				// nodes that tsgo exposes directly (ESTree wraps those in a
				// FunctionExpression child that the regex also matches).
				// Excludes signature kinds (MethodSignature, CallSignature,
				// etc.) and ClassStaticBlockDeclaration — none of which can
				// host a runtime `this.setState()` under `componentDidUpdate`.
				depth := 0
				for p := node.Parent; p != nil; p = p.Parent {
					if ast.IsFunctionLikeDeclaration(p) {
						depth++
					}
					if !isStopper(p) {
						continue
					}
					if stopperName(p) != methodName {
						continue
					}
					if !disallowInFunc && depth > 1 {
						continue
					}
					ctx.ReportNode(callee, rule.RuleMessage{
						Id:          "noSetState",
						Description: "Do not use setState in componentDidUpdate",
					})
					return
				}
			},
		}
	},
}
