package no_will_update_set_state

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// methodName is the single lifecycle hook this rule targets. Upstream's
// makeNoMethodSetStateRule factory is parameterized; since
// no-will-update-set-state always passes 'componentWillUpdate', we hardcode it.
// Upstream additionally accepts the `UNSAFE_componentWillUpdate` alias once
// `testReactVersion(ctx, '>= 16.3.0')` passes — see checkUnsafe below.
const methodName = "componentWillUpdate"

// unsafeMethodName is the React 16.3+ renamed alias. Upstream matches it via
// `shouldCheckUnsafeCb` which is wired to `testReactVersion(ctx, '>= 16.3.0')`
// in `no-will-update-set-state.js`.
const unsafeMethodName = "UNSAFE_componentWillUpdate"

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

// esTreeNameOf returns the string an ESTree consumer would read from
// `node.name` for an Identifier / PrivateIdentifier. PrivateIdentifier
// strips the leading `#` per spec (tsgo retains it, so we trim). Any other
// node kind — ComputedPropertyName, StringLiteral, NumericLiteral, etc. —
// has no `.name` in ESTree and yields "". Callers that compare against a
// known identifier string therefore get the correct "never equal" answer
// without extra guards.
//
// Used in both contexts the rule cares about: the member-expression
// property (`callee.property.name`) and the container key
// (`ancestor.key.name`). Neither is a true "static property name" in the
// wider sense — e.g. `this['setState']()` must NOT match (upstream:
// `'name' in callee.property` is false), so reusing
// `utils.GetStaticPropertyName` here would be incorrect (it resolves
// string literals as static names).
func esTreeNameOf(nameNode *ast.Node) string {
	if nameNode == nil {
		return ""
	}
	switch nameNode.Kind {
	case ast.KindIdentifier:
		return nameNode.AsIdentifier().Text
	case ast.KindPrivateIdentifier:
		return strings.TrimPrefix(nameNode.AsPrivateIdentifier().Text, "#")
	}
	return ""
}

// stopperName returns the ESTree-visible name of a stopper node's key.
// See [esTreeNameOf] for the aliasing rationale.
func stopperName(node *ast.Node) string {
	return esTreeNameOf(node.Name())
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

// shouldCheckUnsafe mirrors upstream's `testReactVersion(context, '>= 16.3.0')`
// callback wired as `shouldCheckUnsafeCb`. When it returns true, upstream's
// `nameMatches` also accepts `UNSAFE_componentWillUpdate`.
//
// Note: `no-will-update-set-state` is NOT in upstream's `methodNoopsAsOf` map,
// so `shouldBeNoop` always returns false — the rule stays active regardless of
// React version. The version check is used ONLY to decide whether to also
// match the UNSAFE_ alias, not to disable the rule.
func shouldCheckUnsafe(settings map[string]interface{}) bool {
	return !reactutil.ReactVersionLessThan(settings, 16, 3, 0)
}

var NoWillUpdateSetStateRule = rule.Rule{
	Name: "react/no-will-update-set-state",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		disallowInFunc := parseDisallowInFunc(options)
		checkUnsafe := shouldCheckUnsafe(ctx.Settings)

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
				// === 'setState'`. esTreeNameOf yields "" for shapes ESTree's
				// `.name` lookup would miss (string/numeric/computed keys),
				// so the equality check below correctly rejects them.
				if esTreeNameOf(prop.Name()) != "setState" {
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
				// host a runtime `this.setState()` under `componentWillUpdate`.
				depth := 0
				for p := node.Parent; p != nil; p = p.Parent {
					if ast.IsFunctionLikeDeclaration(p) {
						depth++
					}
					if !isStopper(p) {
						continue
					}
					sn := stopperName(p)
					// nameMatches: methodName, or (when checkUnsafe) the UNSAFE_ alias.
					if sn != methodName && (!checkUnsafe || sn != unsafeMethodName) {
						continue
					}
					if !disallowInFunc && depth > 1 {
						continue
					}
					ctx.ReportNode(callee, rule.RuleMessage{
						Id:          "noSetState",
						Description: "Do not use setState in " + sn,
					})
					return
				}
			},
		}
	},
}
