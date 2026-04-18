package no_extend_native

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// nativeBuiltins lists ECMAScript globals whose first letter is uppercase.
// Mirrors ESLint's `Object.keys(astUtils.ECMASCRIPT_GLOBALS).filter(b => b[0].toUpperCase() === b[0])`,
// where `ECMASCRIPT_GLOBALS = globals[`es${LATEST_ECMA_VERSION}`]` from the
// `globals` package. Generated from `globals.es2027` to stay byte-for-byte
// in sync with upstream ESLint.
var nativeBuiltins = map[string]bool{
	"AggregateError": true, "Array": true, "ArrayBuffer": true, "Atomics": true,
	"BigInt": true, "BigInt64Array": true, "BigUint64Array": true, "Boolean": true,
	"DataView": true, "Date": true,
	"Error": true, "EvalError": true,
	"FinalizationRegistry": true, "Float16Array": true, "Float32Array": true, "Float64Array": true, "Function": true,
	"Infinity": true, "Int16Array": true, "Int32Array": true, "Int8Array": true, "Intl": true, "Iterator": true,
	"JSON": true,
	"Map": true, "Math": true,
	"NaN": true, "Number": true,
	"Object": true,
	"Promise": true, "Proxy": true,
	"RangeError": true, "ReferenceError": true, "Reflect": true, "RegExp": true,
	"Set": true, "SharedArrayBuffer": true, "String": true, "Symbol": true, "SyntaxError": true,
	"TypeError": true,
	"URIError": true, "Uint16Array": true, "Uint32Array": true, "Uint8Array": true, "Uint8ClampedArray": true,
	"WeakMap": true, "WeakRef": true, "WeakSet": true,
}

type options struct {
	exceptions map[string]bool
}

func parseOptions(opts any) options {
	result := options{exceptions: make(map[string]bool)}
	optsMap := utils.GetOptionsMap(opts)
	if optsMap != nil {
		if exceptions, ok := optsMap["exceptions"].([]interface{}); ok {
			for _, e := range exceptions {
				if s, ok := e.(string); ok {
					result.exceptions[s] = true
				}
			}
		}
	}
	return result
}

// isAssignmentOperator reports whether the given binary operator is a
// (compound) assignment, including logical assignments.
func isAssignmentOperator(kind ast.Kind) bool {
	switch kind {
	case ast.KindEqualsToken,
		ast.KindPlusEqualsToken,
		ast.KindMinusEqualsToken,
		ast.KindAsteriskEqualsToken,
		ast.KindAsteriskAsteriskEqualsToken,
		ast.KindSlashEqualsToken,
		ast.KindPercentEqualsToken,
		ast.KindLessThanLessThanEqualsToken,
		ast.KindGreaterThanGreaterThanEqualsToken,
		ast.KindGreaterThanGreaterThanGreaterThanEqualsToken,
		ast.KindAmpersandEqualsToken,
		ast.KindBarEqualsToken,
		ast.KindCaretEqualsToken,
		ast.KindAmpersandAmpersandEqualsToken,
		ast.KindBarBarEqualsToken,
		ast.KindQuestionQuestionEqualsToken:
		return true
	}
	return false
}

// memberObject returns the object expression of a property/element access node,
// or nil if the node is not a member access.
func memberObject(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		return node.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		return node.AsElementAccessExpression().Expression
	}
	return nil
}

// staticMemberName returns the static property name of a member access, or
// ("", false) if it cannot be determined statically.
func staticMemberName(node *ast.Node) (string, bool) {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		name := node.AsPropertyAccessExpression().Name()
		if name != nil && name.Kind == ast.KindIdentifier {
			return name.AsIdentifier().Text, true
		}
	case ast.KindElementAccessExpression:
		return utils.GetStaticExpressionValue(ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression))
	}
	return "", false
}

// skipParensUp walks up from `node` through ParenthesizedExpression parents.
func skipParensUp(node *ast.Node) *ast.Node {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	return node
}

// https://eslint.org/docs/latest/rules/no-extend-native
var NoExtendNativeRule = rule.Rule{
	Name: "no-extend-native",
	Run: func(ctx rule.RuleContext, opts any) rule.RuleListeners {
		o := parseOptions(opts)

		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				name := node.Text()
				if !nativeBuiltins[name] || o.exceptions[name] {
					return
				}

				// In tsgo, parentheses are explicit nodes — `(Object).prototype.p`
				// wraps the identifier in a ParenthesizedExpression. Walk up
				// through any wrapping parens before checking the member access.
				identExpr := skipParensUp(node)
				parent := identExpr.Parent
				if parent == nil {
					return
				}

				// `identExpr` must be the object of a member access whose property is `prototype`.
				obj := memberObject(parent)
				if obj != identExpr {
					return
				}
				propName, ok := staticMemberName(parent)
				if !ok || propName != "prototype" {
					return
				}

				if utils.IsShadowed(node, name) {
					return
				}

				// Walk up through any wrapping parentheses to find the next significant parent.
				prototypeAccess := skipParensUp(parent)
				next := prototypeAccess.Parent
				if next == nil {
					return
				}

				// Case 1: assignment to a property of the prototype.
				// e.g. `Object.prototype.p = 0`, `(Object?.prototype).p = 0`,
				//      `Array.prototype.p &&= 0`.
				if memberObject(next) == prototypeAccess {
					memberAccess := skipParensUp(next)
					assign := memberAccess.Parent
					if assign != nil && assign.Kind == ast.KindBinaryExpression {
						bin := assign.AsBinaryExpression()
						if bin.OperatorToken != nil &&
							isAssignmentOperator(bin.OperatorToken.Kind) &&
							bin.Left == memberAccess {
							ctx.ReportNode(assign, rule.RuleMessage{
								Id:          "unexpected",
								Description: name + " prototype is read only, properties should not be added.",
							})
							return
						}
					}
				}

				// Case 2: first argument of `Object.defineProperty` /
				// `Object.defineProperties`.
				if next.Kind == ast.KindCallExpression {
					call := next.AsCallExpression()
					if call.Arguments == nil || len(call.Arguments.Nodes) == 0 ||
						call.Arguments.Nodes[0] != prototypeAccess {
						return
					}
					if utils.IsSpecificMemberAccess(call.Expression, "Object", "defineProperty") ||
						utils.IsSpecificMemberAccess(call.Expression, "Object", "defineProperties") {
						ctx.ReportNode(next, rule.RuleMessage{
							Id:          "unexpected",
							Description: name + " prototype is read only, properties should not be added.",
						})
					}
				}
			},
		}
	},
}
