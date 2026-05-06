package reactutil

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoMethodSetStateConfig configures the rule produced by [MakeNoMethodSetStateRule].
//
// The fields mirror upstream's `makeNoMethodSetStateRule(methodName, shouldCheckUnsafeCb)`
// factory in eslint-plugin-react/lib/util/makeNoMethodSetStateRule.js, with
// `ShouldBeNoop` extracted as a separate hook so the `methodNoopsAsOf` map
// upstream encodes inline can be expressed per-rule on our side.
type NoMethodSetStateConfig struct {
	// RuleName is the registered rule name, e.g. "react/no-did-mount-set-state".
	RuleName string

	// MethodName is the lifecycle hook the rule guards against, e.g.
	// "componentDidMount". Stopper containers (class methods, class fields,
	// object-literal properties, accessors, constructors) whose key resolves
	// to this name — or to "UNSAFE_<MethodName>" when [ShouldCheckUnsafe]
	// returns true — are reported.
	MethodName string

	// ShouldCheckUnsafe gates the `UNSAFE_<MethodName>` alias. When nil or
	// returning false, only [MethodName] matches. Mirrors upstream's
	// `shouldCheckUnsafeCb`.
	ShouldCheckUnsafe func(settings map[string]interface{}) bool

	// ShouldBeNoop, when non-nil and returning true, makes the rule emit no
	// listeners — it produces no diagnostics regardless of source. Mirrors
	// upstream's `shouldBeNoop(context, methodName)` gate (the
	// `methodNoopsAsOf` map plus the `>= 999.999.999` upper-bound carve-out).
	ShouldBeNoop func(settings map[string]interface{}) bool
}

// MakeNoMethodSetStateRule produces a rule that reports `this.setState(...)`
// calls whose enclosing class method, class field initializer, or
// object-literal property is keyed [NoMethodSetStateConfig.MethodName].
//
// Mirrors upstream's `makeNoMethodSetStateRule` 1:1 in semantics — including
// the innermost-stopper search (`findLast(ancestors, ...)`), the
// function-depth counter that gates `disallow-in-func`, and the ESTree-shaped
// `'name' in callee.property` test (which we model with [esTreeName]: any
// non-Identifier / non-PrivateIdentifier key yields "" and never matches).
//
// All three of `no-did-mount-set-state`, `no-did-update-set-state`, and
// `no-will-update-set-state` are thin wrappers around this factory; do not
// inline these helpers in any new rule — extend the factory instead.
func MakeNoMethodSetStateRule(cfg NoMethodSetStateConfig) rule.Rule {
	return rule.Rule{
		Name: cfg.RuleName,
		Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
			if cfg.ShouldBeNoop != nil && cfg.ShouldBeNoop(ctx.Settings) {
				return rule.RuleListeners{}
			}
			disallowInFunc := parseDisallowInFunc(options)
			checkUnsafe := cfg.ShouldCheckUnsafe != nil && cfg.ShouldCheckUnsafe(ctx.Settings)
			unsafeName := "UNSAFE_" + cfg.MethodName
			description := "Do not use setState in " + cfg.MethodName
			unsafeDescription := "Do not use setState in " + unsafeName

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
					// === 'setState'`. esTreeName yields "" for non-name shapes
					// (string/numeric/computed keys), so the equality check
					// below correctly rejects them.
					if esTreeName(prop.Name()) != "setState" {
						return
					}

					// Walk ancestors innermost-to-outermost, mirroring ESLint's
					// `findLast(ancestors, cb)`. Depth counts function-like
					// wrappers crossed before the first matching stopper; in
					// default mode a match at depth > 1 is skipped (setState
					// in a nested callback), while `disallow-in-func` reports
					// at any depth.
					depth := 0
					for p := node.Parent; p != nil; p = p.Parent {
						if ast.IsFunctionLikeDeclaration(p) {
							depth++
						}
						if !isMethodSetStateStopper(p) {
							continue
						}
						sn := stopperName(p)
						matched := sn == cfg.MethodName
						unsafe := false
						if !matched && checkUnsafe && sn == unsafeName {
							matched = true
							unsafe = true
						}
						if !matched {
							continue
						}
						if !disallowInFunc && depth > 1 {
							continue
						}
						desc := description
						if unsafe {
							desc = unsafeDescription
						}
						ctx.ReportNode(callee, rule.RuleMessage{
							Id:          "noSetState",
							Description: desc,
						})
						return
					}
				},
			}
		},
	}
}

// isMethodSetStateStopper mirrors upstream's stopper-type check
// (`Property | MethodDefinition | ClassProperty | PropertyDefinition`),
// translated to tsgo's collapsed AST.
//
// In tsgo these collapse into PropertyAssignment (object `name: fn`),
// MethodDeclaration (class methods + object shorthand), GetAccessor /
// SetAccessor / Constructor, and PropertyDeclaration (class field).
// ShorthandPropertyAssignment is omitted — it has no function body, so
// setState could never appear beneath it.
//
// `ast.IsMethodOrAccessor` covers MethodDeclaration / GetAccessor /
// SetAccessor; the remaining three kinds are listed explicitly.
func isMethodSetStateStopper(node *ast.Node) bool {
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

// stopperName returns the ESTree-visible name of a stopper node's key.
// See [esTreeName] for the aliasing rationale.
func stopperName(node *ast.Node) string {
	return esTreeName(node.Name())
}

// esTreeName returns the string an ESTree consumer would read from
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
func esTreeName(nameNode *ast.Node) string {
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

// parseDisallowInFunc recognizes the rule's single string option,
// `"disallow-in-func"`, in both the bare and array-wrapped shapes the
// rule_tester / config layers may produce.
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

// MethodNoopAtReactVersion returns a [NoMethodSetStateConfig.ShouldBeNoop]
// callback equivalent to upstream's `shouldBeNoop` for entries in
// `methodNoopsAsOf`: noop iff `settings.react.version` is explicitly set
// AND falls in the half-open range [major.minor.patch, 999.999.999).
//
// The upper-bound 999.999.999 carve-out matches upstream — it is the
// "version absent" sentinel both upstream and rslint use, so a user who
// pins that string is treated as "version unpinned" and the rule stays
// active.
func MethodNoopAtReactVersion(major, minor, patch int) func(map[string]interface{}) bool {
	return func(settings map[string]interface{}) bool {
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
		return !ReactVersionLessThan(settings, major, minor, patch) &&
			ReactVersionLessThan(settings, 999, 999, 999)
	}
}

// CheckUnsafeAtReactVersion returns a [NoMethodSetStateConfig.ShouldCheckUnsafe]
// callback equivalent to upstream's `testReactVersion(context, '>= X.Y.Z')`:
// true iff the configured (or defaulted) React version is at least
// major.minor.patch.
func CheckUnsafeAtReactVersion(major, minor, patch int) func(map[string]interface{}) bool {
	return func(settings map[string]interface{}) bool {
		return !ReactVersionLessThan(settings, major, minor, patch)
	}
}
