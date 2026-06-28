// TestNoPromiseInCallbackExtras locks in branches and edge shapes that the upstream
// suite does not exercise: callback identification, promise-callback exemptions,
// return exemptions, and tsgo AST wrapper shapes.
package no_promise_in_callback_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/no_promise_in_callback"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoPromiseInCallbackExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_promise_in_callback.NoPromiseInCallbackRule,
		[]rule_tester.ValidTestCase{
			// ---- Branch lock-in: only err/error as the first parameter are callbacks ----
			{Code: `a(function(e) { doThing().then(a) })`},
			{Code: `a(function(errrr) { doThing().then(a) })`},
			{Code: `a(function({ err }) { doThing().then(a) })`},
			{Code: `a(function(...err) { doThing().then(a) })`},
			{Code: `a(function(err = null) { doThing().then(a) })`},

			// ---- Branch lock-in: non-promise calls and computed members are ignored ----
			{Code: `a(function(err) { doThing() })`},
			{Code: `a(function(err) { doThing()["then"](a) })`},
			{Code: `a(function(err) { Promise["resolve"](value) })`},
			{Code: `a(function(err) { Promise.withResolvers() })`},

			// ---- Branch lock-in: returned promises are allowed ----
			{Code: `a(function(err) { return (doThing().then(a)) })`},
			{Code: `a(function(err) { return Promise.resolve(err) })`},

			// ---- Branch lock-in: promise callbacks for .then/.catch are exempt ----
			{Code: `a.then(function(err) { doThing().then(a) })`},
			{Code: `a.then(ok, function(err) { doThing().then(a) })`},
			{Code: `a.catch(function(error) { Promise.resolve(error) })`},
			{Code: `a.catch((error) => { Promise.resolve(error) })`},

			// ---- Branch lock-in: exemptDeclarations only exempts declarations ----
			{
				Code:    `function handler(err) { Promise.resolve(err) }`,
				Options: map[string]interface{}{"exemptDeclarations": true},
			},

			// ---- Dimension 4: parenthesized function and call wrappers ----
			{Code: `a((function(err) { return (Promise.resolve(err)) }))`},

			// ---- Real-user: Node-style callback that returns the promise to the caller ----
			{Code: `fs.readFile(file, function(err, data) { if (err) return cb(err); return transform(data).then(cb) })`},

			// ---- Real-user: legacy helper callback with unrelated error-shaped name ----
			{Code: `legacy(function(errorMessage) { fetchData().then(useData) })`},

			// ---- Method/accessor containers whose first param isn't err/error ----
			{Code: `const o = { onError(e) { Promise.resolve(e) } }`},
			{Code: `class X { onError(reason) { audit().catch(log) } }`},
			// A getter takes no parameters, so it is never an err/error callback.
			{Code: `class X { get value() { return p.then(a) } }`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Branch lock-in: err/error as first parameter are callbacks ----
			{
				Code:   `a(function(err) { doThing().then(a) })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},
			{
				Code:   `a(function(error) { Promise.resolve(error) })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},

			// ---- Branch lock-in: function expressions are not exempted by exemptDeclarations ----
			{
				Code:    `a(function(err) { Promise.resolve(err) })`,
				Options: map[string]interface{}{"exemptDeclarations": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},
			{
				Code:    `a((err) => Promise.resolve(err))`,
				Options: map[string]interface{}{"exemptDeclarations": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},

			// ---- Branch lock-in: function declarations report unless explicitly exempted ----
			{
				Code:   `function handler(err) { Promise.resolve(err) }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},

			// ---- Branch lock-in: .finally callbacks are not promise-callback exemptions upstream ----
			{
				Code:   `a.finally(function(err) { doThing().then(a) })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},

			// ---- Dimension 4: parenthesized promise call still reports ----
			{
				Code:   `a(function(err) { (doThing().then(a)) })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},

			// ---- Branch lock-in: upstream only exempts the promise call itself being returned ----
			// A promise nested inside a returned object still reports because the call is
			// not the direct return expression.
			{
				Code:   `a(function(err) { return { promise: Promise.resolve(err) } })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},

			// ---- Dimension 4: optional chain promise-like call ----
			{
				Code:   `a(function(err) { doThing()?.then(a) })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},

			// ---- Dimension 4: TS assertion wrapper around the receiver ----
			// The .then() call remains promise-like because the outer callee is still a
			// dotted property access named then.
			{
				Code:   `a(function(err) { (doThingWith(err) as any).then(a) })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},

			// ---- Branch lock-in: nested ordinary callbacks preserve the outer callback context ----
			{
				Code:   `a(function(err) { helper(function() { Promise.resolve(err) }) })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},

			// ---- Real-user: Node-style callback mixing async APIs ----
			{
				Code:   `fs.readFile(file, function(err, data) { if (err) return cb(err); save(data).then(done) })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},

			// ---- Real-user: Express-style middleware callback ----
			{
				Code:   `middleware(function(error, req, res, next) { audit(req).catch(next) })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},

			// ---- Method-like containers: ESTree models these bodies as FunctionExpressions,
			// so upstream flags methods/constructors/setters whose first param is err/error.
			{
				Code:   `const o = { onError(err) { Promise.resolve(err) } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},
			{
				Code:   `class X { onError(error) { audit().catch(log) } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},
			{
				Code:   `class X { constructor(err) { Promise.resolve(err) } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},
			{
				Code:   `class X { set err(error) { audit().catch(log) } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},
		},
	)
}
