package prefer_object_has_own

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildUseHasOwnMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useHasOwn",
		Description: "Use 'Object.hasOwn()' instead of 'Object.prototype.hasOwnProperty.call()'.",
	}
}

// hasOwnReplacement is the text the autofix puts in place of the callee.
const hasOwnReplacement = "Object.hasOwn"

// staticMemberName returns the property a member access names, when that name
// is known statically. A private name (`obj.#call`) never names a static
// property, so it matches nothing this rule looks for.
//
// A computed key is read through parentheses only. `utils.AccessExpressionStaticName`
// also reads through a TypeScript assertion, but `Object[("prototype" as any)]`
// is an assertion expression rather than a string, so upstream reads no static
// name from it at all.
func staticMemberName(node *ast.Node) (string, bool) {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		name := node.AsPropertyAccessExpression().Name()
		if name == nil || name.Kind == ast.KindPrivateIdentifier {
			return "", false
		}
		return name.Text(), true
	case ast.KindElementAccessExpression:
		key := node.AsElementAccessExpression().ArgumentExpression
		if key == nil {
			return "", false
		}
		return utils.GetStaticExpressionValue(ast.SkipParentheses(key))
	}
	return "", false
}

// isMemberAccess reports whether node is one of the two forms ESTree calls a
// `MemberExpression`.
func isMemberAccess(node *ast.Node) bool {
	return node != nil &&
		(node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression)
}

// unwrapParens removes the parentheses around an expression, and returns nil
// when those parentheses end an optional chain.
//
// Only parentheses come off: a TypeScript wrapper such as `(Object as any)` or
// `Object!` is an expression of its own, and the receiver it wraps no longer
// counts as `Object` itself.
//
// Parentheses end an optional chain, and ESTree marks that with a
// `ChainExpression` node wrapping the chain — a node kind this rule matches
// nowhere, so `(a?.b).c` reads as a different shape from `a?.b.c` and goes
// unreported. tsgo has no such node: it flags the chain on each access, which
// is why the flag is read only where a parenthesis actually came off.
func unwrapParens(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	inner := ast.SkipParentheses(node)
	if inner != node && ast.IsOptionalChain(inner) {
		return nil
	}
	return inner
}

// memberObject returns the receiver of a member access, read through
// parentheses by [unwrapParens].
func memberObject(node *ast.Node) *ast.Node {
	return unwrapParens(utils.AccessExpressionObject(node))
}

// hasLeftHandObject reports whether the receiver of `.hasOwnProperty` is an
// access to a property of `Object.prototype` — that is, `Object`,
// `Object.prototype`, or an empty object literal. An object literal with
// properties in it is a different object, so `({ foo }).hasOwnProperty` is not
// one of these.
func hasLeftHandObject(node *ast.Node) bool {
	object := memberObject(node)
	if object == nil {
		return false
	}

	if object.Kind == ast.KindObjectLiteralExpression {
		literal := object.AsObjectLiteralExpression()
		return literal.Properties == nil || len(literal.Properties.Nodes) == 0
	}

	objectToCheck := object
	if isMemberAccess(object) {
		if name, ok := staticMemberName(object); ok && name == "prototype" {
			objectToCheck = memberObject(object)
		}
	}

	return objectToCheck != nil && objectToCheck.Kind == ast.KindIdentifier &&
		objectToCheck.AsIdentifier().Text == "Object"
}

// https://eslint.org/docs/latest/rules/prefer-object-has-own
var PreferObjectHasOwnRule = rule.Rule{
	Name:   "prefer-object-has-own",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				// ESTree drops parentheses, so a callee wrapped in them is the
				// member access itself; tsgo keeps them as nodes of their own.
				callee := unwrapParens(node.AsCallExpression().Expression)
				if !isMemberAccess(callee) {
					return
				}
				calleeObject := memberObject(callee)
				if !isMemberAccess(calleeObject) {
					return
				}

				if name, ok := staticMemberName(callee); !ok || name != "call" {
					return
				}
				if name, ok := staticMemberName(calleeObject); !ok || name != "hasOwnProperty" {
					return
				}
				if !hasLeftHandObject(calleeObject) {
					return
				}

				// `Object` has to be the global one for `Object.hasOwn` to be
				// the same call: a local binding of that name, or a config
				// `/* global Object: off */` entry that un-declares it, means
				// the name resolves elsewhere and the rule stays silent. A call
				// in a parameter default sees the body's bindings as well, one
				// scope wider than the runtime lookup.
				if utils.IsShadowed(node, "Object") ||
					utils.IsShadowedFromParameterInitializer(node, "Object") ||
					!ctx.Globals.Access("Object").IsDeclared() {
					return
				}

				ctx.ReportNodeWithDeferredFixes(node, buildUseHasOwnMessage(), func() []rule.RuleFix {
					calleeRange := utils.TrimNodeTextRange(ctx.SourceFile, callee)
					// A comment anywhere inside the callee would be dropped by
					// the replacement, so the diagnostic carries no fix.
					if utils.HasCommentInSpan(ctx.Comments.All(), calleeRange.Pos(), calleeRange.End()) {
						return nil
					}
					// The replacement opens with an identifier, so it needs a
					// space when the character in front of it closes one —
					// otherwise `return{}.hasOwnProperty.call(a, b)` fuses into
					// `returnObject.hasOwn(a, b)`. Nothing that can sit against
					// the callee ends in an identifier character except an
					// identifier or a keyword, so that character decides it on
					// its own.
					replacement := hasOwnReplacement
					if utils.NeedsLeadingSpaceForReplacement(ctx.SourceFile.Text(), calleeRange.Pos(), replacement) {
						replacement = " " + replacement
					}
					return []rule.RuleFix{rule.RuleFixReplaceRange(calleeRange, replacement)}
				})
			},
		}
	},
}
