package no_callback_in_promise

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/promiseutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const skipTransparent = ast.OEKParentheses

var cbBlacklist = []string{"callback", "cb", "next", "done"}
var timeoutWhitelist = []string{"setImmediate", "setTimeout", "requestAnimationFrame", "nextTick"}

type Options struct {
	Exceptions  []string
	TimeoutsErr bool
}

func parseOptions(options any) Options {
	opts := Options{}
	optsMap := utils.GetOptionsMap(options)
	if optsMap == nil {
		return opts
	}
	if v, ok := optsMap["timeoutsErr"].(bool); ok {
		opts.TimeoutsErr = v
	}
	if v, ok := optsMap["exceptions"].([]interface{}); ok {
		for _, e := range v {
			if s, ok := e.(string); ok {
				opts.Exceptions = append(opts.Exceptions, s)
			}
		}
	}
	return opts
}

func buildCallbackMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "callback",
		Description: "Avoid calling back inside of a promise.",
	}
}

// isCallbackName returns true if name is in cbBlacklist and not in exceptions.
func isCallbackName(name string, exceptions []string) bool {
	for _, e := range exceptions {
		if name == e {
			return false
		}
	}
	for _, cb := range cbBlacklist {
		if name == cb {
			return true
		}
	}
	return false
}

// isCallbackCall returns true if node is a call expression whose callee (skipping parens) is an
// identifier with a callback name.
func isCallbackCall(node *ast.Node, exceptions []string) bool {
	if !ast.IsCallExpression(node) {
		return false
	}
	callee := ast.SkipOuterExpressions(node.AsCallExpression().Expression, skipTransparent)
	return callee != nil && ast.IsIdentifier(callee) && isCallbackName(callee.AsIdentifier().Text, exceptions)
}

// getMemberCallName returns the member-property name (for a.foo()) or the direct identifier name
// (for foo()) from a call-expression callee. Returns "" for anything else.
func getMemberCallName(node *ast.Node) string {
	if !ast.IsCallExpression(node) {
		return ""
	}
	callee := ast.SkipOuterExpressions(node.AsCallExpression().Expression, skipTransparent)
	if ast.IsPropertyAccessExpression(callee) {
		propName := callee.AsPropertyAccessExpression().Name()
		if propName != nil && ast.IsIdentifier(propName) {
			return propName.AsIdentifier().Text
		}
		return ""
	}
	if ast.IsIdentifier(callee) {
		return callee.AsIdentifier().Text
	}
	return ""
}

func isPromiseMemberCall(node *ast.Node) bool {
	return promiseutil.IsMemberCall(node, "then") || promiseutil.IsMemberCall(node, "catch")
}

// isFunctionLike returns true for FunctionExpression and ArrowFunction nodes.
func isFunctionLike(node *ast.Node) bool {
	return node.Kind == ast.KindFunctionExpression || node.Kind == ast.KindArrowFunction
}

// isInsidePromise mirrors eslint-plugin-promise's lib/is-inside-promise:
// a FunctionExpression/ArrowFunction whose parent (skipping parens) is a .then()/.catch() call.
func isInsidePromise(node *ast.Node) bool {
	if !isFunctionLike(node) {
		return false
	}
	parent := node.Parent
	for parent != nil && ast.IsOuterExpression(parent, skipTransparent) {
		parent = parent.Parent
	}
	return parent != nil && isPromiseMemberCall(parent)
}

// isInsideTimeout mirrors eslint-plugin-promise's isInsideTimeout:
// a FunctionExpression/ArrowFunction whose parent (skipping parens) is a timeout whitelist call.
func isInsideTimeout(node *ast.Node) bool {
	if !isFunctionLike(node) {
		return false
	}
	parent := node.Parent
	for parent != nil && ast.IsOuterExpression(parent, skipTransparent) {
		parent = parent.Parent
	}
	if parent == nil {
		return false
	}
	name := getMemberCallName(parent)
	for _, t := range timeoutWhitelist {
		if name == t {
			return true
		}
	}
	return false
}

// ancestorSome walks up the parent chain and returns true if pred matches any ancestor.
func ancestorSome(node *ast.Node, pred func(*ast.Node) bool) bool {
	for cur := node.Parent; cur != nil; cur = cur.Parent {
		if pred(cur) {
			return true
		}
	}
	return false
}

var NoCallbackInPromiseRule = rule.Rule{
	Name: "promise/no-callback-in-promise",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				exceptions := opts.Exceptions

				if !isCallbackCall(node, exceptions) {
					// node is not itself a callback call — check if it is a promise callback
					// call (.then/.catch) whose first arg is a callback name, or handle
					// the timeoutsErr path.
					callExpr := node.AsCallExpression()
					var firstArg *ast.Node
					var firstArgName string
					if callExpr.Arguments != nil && len(callExpr.Arguments.Nodes) > 0 {
						firstArg = ast.SkipOuterExpressions(callExpr.Arguments.Nodes[0], skipTransparent)
						if firstArg != nil && ast.IsIdentifier(firstArg) {
							firstArgName = firstArg.AsIdentifier().Text
						}
					}

					if isPromiseMemberCall(node) {
						// Scenario A: a.then(cb) — callback passed directly as argument. Report the
						// unwrapped identifier (firstArg) to match upstream's node.arguments[0]; the raw
						// Nodes[0] would span the surrounding parens, e.g. `(cb)` instead of `cb`.
						if firstArgName != "" && isCallbackName(firstArgName, exceptions) {
							ctx.ReportNode(firstArg, buildCallbackMessage())
						}
						return
					}

					if !opts.TimeoutsErr {
						return
					}
					if firstArgName == "" {
						return
					}
				}

				// Scenario B/C/D: callback call (or named-arg call in timeoutsErr mode)
				// inside a promise handler.
				if ancestorSome(node, isInsidePromise) &&
					(opts.TimeoutsErr || !ancestorSome(node, isInsideTimeout)) {
					ctx.ReportNode(node, buildCallbackMessage())
				}
			},
		}
	},
}
