package no_global_assign

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// builtinGlobals contains names of known read-only built-in globals.
var builtinGlobals = map[string]bool{
	"AggregateError": true, "Array": true, "ArrayBuffer": true, "AsyncDisposableStack": true,
	"AsyncIterator": true, "Atomics": true,
	"BigInt": true, "BigInt64Array": true, "BigUint64Array": true,
	"Boolean": true, "DataView": true, "Date": true,
	"decodeURI": true, "decodeURIComponent": true, "DisposableStack": true,
	"encodeURI": true, "encodeURIComponent": true,
	"Error": true, "escape": true, "EvalError": true,
	"FinalizationRegistry": true, "Float32Array": true, "Float64Array": true, "Function": true,
	"globalThis": true, "Infinity": true, "Int8Array": true,
	"Int16Array": true, "Int32Array": true, "Intl": true, "isFinite": true,
	"isNaN": true, "Iterator": true, "JSON": true, "Map": true, "Math": true,
	"NaN": true, "Number": true, "Object": true, "parseFloat": true,
	"parseInt": true, "Promise": true, "Proxy": true, "RangeError": true,
	"ReferenceError": true, "Reflect": true, "RegExp": true,
	"Set": true, "SharedArrayBuffer": true, "String": true, "SuppressedError": true,
	"Symbol": true, "SyntaxError": true, "TypeError": true,
	"Uint8Array": true, "Uint8ClampedArray": true, "Uint16Array": true,
	"Uint32Array": true, "unescape": true, "URIError": true, "undefined": true,
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

// isWriteThroughTypeAssertion checks if the identifier reaches its assignment target
// through an AsExpression or TypeAssertionExpression. ESLint's scope analysis does not
// track writes through these TS-specific wrappers, so we skip them to match ESLint.
func isWriteThroughTypeAssertion(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		switch current.Kind {
		case ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindSatisfiesExpression:
			return true
		case ast.KindParenthesizedExpression, ast.KindNonNullExpression:
			current = current.Parent
			continue
		default:
			return false
		}
	}
	return false
}

// NoGlobalAssignRule disallows assignments to native objects or read-only global variables
var NoGlobalAssignRule = rule.Rule{
	Name: "no-global-assign",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				name := node.Text()
				if !builtinGlobals[name] || opts.exceptions[name] {
					return
				}

				if !utils.IsWriteReference(node) {
					return
				}

				if isWriteThroughTypeAssertion(node) {
					return
				}

				if utils.IsShadowed(node, name) {
					return
				}

				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "globalShouldNotBeModified",
					Description: fmt.Sprintf("Read-only global '%s' should not be modified.", name),
				})
			},
		}
	},
}
