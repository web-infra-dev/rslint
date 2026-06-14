package no_callback_in_promise_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/no_callback_in_promise"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoCallbackInPromiseExtras covers Layers 2 (edge-shape & real-user) and 3 (branch lock-ins).
func TestNoCallbackInPromiseExtras(t *testing.T) {
	e := func() rule_tester.InvalidTestCaseError {
		return rule_tester.InvalidTestCaseError{MessageId: callbackMsgId, Message: callbackMessage}
	}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_callback_in_promise.NoCallbackInPromiseRule,
		[]rule_tester.ValidTestCase{

			// ---- Dimension 4: Receiver / expression wrappers ----

			// Parenthesized .then receiver — tsgo preserves parens as a node, upstream flattens them.
			// isInsidePromise checks parent of the function, which is the .then() call. ✓
			{Code: `(a).then(function() { return 1 })`},

			// Computed member access — a['then'](cb) is NOT matched as a promise callback
			// (callee is ElementAccessExpression, not PropertyAccessExpression).
			// N/A for isInsidePromise: same structural reason.
			{Code: `a['then'](cb)`},

			// ---- Dimension 4: Parenthesized function arg ----

			// tsgo represents (function(){}) as a ParenthesizedExpression wrapping FunctionExpression.
			// isInsidePromise skips parens to reach the .then() call, so this IS flagged.
			// (tested in invalid section)

			// ---- Dimension 4: Optional chains ----

			// Optional call on .then: a.then?.(cb) — QuestionDotToken on the CallExpression.
			// getMemberCallName still returns "then" since Property name is still "then".
			// Upstream ESTree would see the same property name.
			{Code: `a.then?.(function() { return 1 })`},

			// Optional chain on receiver: a?.then(function() {...}) still matches .then —
			// property name is still "then". Valid because no callback is passed.
			{Code: `a?.then(function() { return 1 })`},

			// ---- Dimension 4: Callback not at first argument position ----

			// cb is the second arg to .then — only first arg is checked for direct passing.
			// Locks in upstream: name = node.arguments?.[0]?.name only looks at first arg.
			{Code: `a.then(null, cb)`},

			// ---- Dimension 4: Non-blacklist name passed to .then ----

			// myHandler is not in cbBlacklist → valid.
			// Locks in upstream Scenario A: CB_BLACKLIST.includes(name) === false path.
			{Code: `a.then(myHandler)`},
			{Code: `a.then(handler)`},
			{Code: `a.then(resolve)`},

			// ---- Dimension 4: Function declaration / class method boundaries ----

			// Method declaration is not FunctionExpression/ArrowFunction — isInsidePromise
			// correctly ignores it. But the outer function IS still a promise boundary:
			// cb() inside a method inside a class inside a promise WILL be flagged because
			// isInsidePromise checks ALL ancestors (not just the nearest function boundary).
			// This is upstream-identical behavior. Lock in with a method that is NOT inside a
			// promise handler to confirm non-flagging:
			{Code: `class Foo { method() { cb() } }`},
			{Code: `const o = { m() { cb() } }`},

			// ---- Dimension 4: requestAnimationFrame in timeout whitelist ----

			// requestAnimationFrame is in the whitelist — not flagged when timeoutsErr=false.
			{Code: `whatever.then((err) => { requestAnimationFrame(() => cb()) })`},
			{Code: `whatever.then((err) => requestAnimationFrame(cb))`},

			// ---- Dimension 4: Graceful degradation — empty call, no args ----

			// .then() with no args — no first argument, no report for Scenario A.
			{Code: `a.then()`},

			// ---- Dimension 4: Non-identifier first arg to .then ----

			// Literal as first arg: not a callback name.
			{Code: `a.then(42)`},
			// Expression as first arg (not a plain identifier).
			{Code: `a.then(arr[0])`},
			// Arrow function as first arg — handled by isInsidePromise path, no direct-arg report.
			{Code: `a.then(() => 1)`},

			// ---- Dimension 4: timeoutsErr=false (default), non-named first arg ----

			// Locks in upstream branch: !timeoutsErr → early return when NOT isPromiseMemberCall.
			// Any call with non-callback callee inside a promise is OK without timeoutsErr.
			{Code: `a.then(() => { foo(bar) })`},

			// ---- Dimension 4: timeoutsErr=true, first arg without identifier ----

			// Locks in upstream branch: !name → "Will be handled elsewhere" early return.
			// setTimeout with a literal/expression arg inside promise should NOT be flagged.
			{Code: `a.then(() => { foo(42) })`, Options: map[string]interface{}{"timeoutsErr": true}},
			{Code: `a.then(() => { foo() })`, Options: map[string]interface{}{"timeoutsErr": true}},

			// ---- Branch lock-ins (Layer 3) ----

			// Locks in upstream isInsidePromise: function whose parent is NOT then/catch.
			// cb() inside .map() callback — not inside a promise handler.
			{Code: `a.map(function() { cb() })`},
			{Code: `a.map(() => cb())`},
			{Code: `a.finally(function() { cb() })`},

			// Locks in upstream ancestor check FALSE: cb() at top level, no promise ancestor.
			{Code: `cb()`},
			{Code: `done()`},

			// Locks in upstream: callback NOT in blacklist after applying exceptions.
			{Code: `a.then(() => next())`, Options: map[string]interface{}{"exceptions": []interface{}{"next"}}},
			{Code: `a.then(() => done())`, Options: map[string]interface{}{"exceptions": []interface{}{"done"}}},

			// Locks in upstream valid: isCallback() = false AND isPromiseMemberCall() = false
			// AND timeoutsErr = true AND name is defined BUT not inside any promise ancestor.
			{Code: `setTimeout(callback)`, Options: map[string]interface{}{"timeoutsErr": true}},

			// ---- Real-user: #572 — non-promise context with callback-shaped arg ----

			// iterator.call(next, ...) outside a promise: not flagged.
			{Code: `call(next, iterator)`},
			// Same pattern inside a while loop that is NOT in a promise.
			{Code: `while (cond) { call(next, iter) }`},

			// Top-level then should not be flagged (not a member expression)
			{Code: `then(function() { cb() })`},
		},
		[]rule_tester.InvalidTestCase{

			// ---- Dimension 4: Parenthesized function arg ----

			// tsgo wraps (function(){}) as a ParenthesizedExpression; isInsidePromise
			// must skip it to reach the .then() call.
			{Code: `a.then((function(err) { cb(err) }))`,
				Errors: []rule_tester.InvalidTestCaseError{e()}},

			// ---- Dimension 4: Parenthesized callback callee ----

			// (cb)() inside a promise handler — isCallbackCall must skip parens on callee.
			{Code: `a.then(() => { (cb)() })`,
				Errors: []rule_tester.InvalidTestCaseError{e()}},

			// ---- Dimension 4: Parenthesized first arg to .then ----

			// a.then((cb)) — in tsgo, Arguments.Nodes[0] is a ParenthesizedExpression.
			// We skip outer expressions to reach the Identifier for the name check, but
			// report the raw argument node (the ParenthesizedExpression).
			{Code: `a.then((cb))`,
				Errors: []rule_tester.InvalidTestCaseError{e()}},

			// ---- Dimension 4: "done" and "next" as callback names ----

			// Locks in cbBlacklist entries other than "cb"/"callback".
			{Code: `a.then(done)`, Errors: []rule_tester.InvalidTestCaseError{e()}},
			{Code: `a.then(next)`, Errors: []rule_tester.InvalidTestCaseError{e()}},
			{Code: `a.then(() => done())`, Errors: []rule_tester.InvalidTestCaseError{e()}},
			{Code: `a.then(() => next())`, Errors: []rule_tester.InvalidTestCaseError{e()}},

			// ---- Dimension 4: .catch direct-arg case ----

			{Code: `a.catch(cb)`, Errors: []rule_tester.InvalidTestCaseError{e()}},
			{Code: `a.catch(done)`, Errors: []rule_tester.InvalidTestCaseError{e()}},

			// ---- Dimension 4: Optional chain receiver, callback inside ----

			// a?.then(() => cb()) — isInsidePromise still matches the arrow's parent (optional
			// .then call), so cb() inside is still flagged.
			{Code: `a?.then(() => cb())`,
				Errors: []rule_tester.InvalidTestCaseError{e()}},

			// ---- Dimension 4: requestAnimationFrame with timeoutsErr=true ----

			// Locks in requestAnimationFrame as a whitelist entry for isInsideTimeout.
			{Code: `whatever.then((err) => { requestAnimationFrame(() => cb()) })`,
				Errors:  []rule_tester.InvalidTestCaseError{e()},
				Options: map[string]interface{}{"timeoutsErr": true}},
			{Code: `whatever.then((err) => requestAnimationFrame(cb))`,
				Errors:  []rule_tester.InvalidTestCaseError{e()},
				Options: map[string]interface{}{"timeoutsErr": true}},

			// ---- Dimension 4: Method/class nested inside promise handler ----

			// cb() inside a method declaration inside a class inside a promise handler.
			// All ancestors are walked; the outer function expression (in .then()) triggers
			// isInsidePromise, so cb() is flagged. (Upstream-identical behavior.)
			{Code: `a.then(function() { class Foo { method() { cb() } } })`,
				Errors: []rule_tester.InvalidTestCaseError{e()}},

			// ---- Branch lock-ins (Layer 3) ----

			// Locks in: isCallbackCall = false AND isPromiseMemberCall = false AND
			// timeoutsErr = true AND name defined AND inside promise — ancestor check fires.
			{Code: `a.then(() => { foo(callback) })`,
				Errors:  []rule_tester.InvalidTestCaseError{e()},
				Options: map[string]interface{}{"timeoutsErr": true}},

			// Locks in: two different promise handlers in a chain both containing callbacks.
			{Code: `a.then(() => cb()).catch(() => done())`,
				Errors: []rule_tester.InvalidTestCaseError{
					e(),
					e(),
				}},

			// ---- Real-user: chained promises with callback ----

			// Deeper chain: a.then(...).then(() => cb()) — cb() is inside inner .then.
			{Code: `a.then(() => 1).then(() => cb())`,
				Errors: []rule_tester.InvalidTestCaseError{e()}},

			// Real-user: passing cb as arg to inner .then while outer .then has no callback.
			{Code: `outer.then(() => inner.then(cb))`,
				Errors: []rule_tester.InvalidTestCaseError{e()}},
		},
	)
}
