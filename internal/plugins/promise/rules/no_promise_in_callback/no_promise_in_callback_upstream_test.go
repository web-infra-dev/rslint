// TestNoPromiseInCallbackUpstream migrates the full valid/invalid suite from
// eslint-plugin-promise __tests__/no-promise-in-callback.js 1:1. rslint-specific
// edge-shape and branch lock-in cases belong in no_promise_in_callback_extras_test.go.
package no_promise_in_callback_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/no_promise_in_callback"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const avoidPromiseInCallbackMessage = "Avoid using promises inside of callbacks."

func TestNoPromiseInCallbackUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_promise_in_callback.NoPromiseInCallbackRule,
		[]rule_tester.ValidTestCase{
			{Code: `go(function() { return Promise.resolve(4) })`},
			{Code: `go(function() { return a.then(b) })`},
			{Code: `go(function() { b.catch(c) })`},
			{Code: `go(function() { b.then(c, d) })`},

			// ---- arrow functions and other things ----
			{Code: `go(() => Promise.resolve(4))`},
			{Code: `go((errrr) => a.then(b))`},
			{Code: `go((helpers) => { b.catch(c) })`},
			{Code: `go((e) => { b.then(c, d) })`},

			// ---- within promises it won't complain ----
			{Code: `a.catch((err) => { b.then(c, d) })`},

			// ---- random unrelated things ----
			{Code: `var x = function() { return Promise.resolve(4) }`},
			{Code: `function y() { return Promise.resolve(4) }`},
			{Code: `function then() { return Promise.reject() }`},
			{Code: `doThing(function(x) { return Promise.reject(x) })`},
			{Code: `doThing().then(function() { return Promise.all([a,b,c]) })`},
			{Code: `doThing().then(function() { return Promise.resolve(4) })`},
			{Code: `doThing().then(() => Promise.resolve(4))`},
			{Code: `doThing().then(() => Promise.all([a]))`},

			// ---- weird case, upstream assumes it's okay if you return ----
			{Code: `a(function(err) { return doThing().then(a) })`},

			{
				Code: `
        function fn(err) {
          return { promise: Promise.resolve(err) };
        }
      `,
				Options: map[string]interface{}{"exemptDeclarations": true},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `a(function(err) { doThing().then(a) })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},
			{
				Code:   `a(function(error, zup, supa) { doThing().then(a) })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},
			{
				Code:   `a(function(error) { doThing().then(a) })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},

			// ---- arrow function ----
			{
				Code:   `a((error) => { doThing().then(a) })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},
			{
				Code:   `a((error) => doThing().then(a))`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},
			{
				Code:   `a((err, data) => { doThing().then(a) })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},
			{
				Code:   `a((err, data) => doThing().then(a))`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},

			// ---- function declarations and similar ----
			{
				Code:   `function x(err) { Promise.all() }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},
			{
				Code:   `function x(err) { Promise.allSettled() }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},
			{
				Code:   `function x(err) { Promise.any() }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},
			{
				Code:   `let x = (err) => doThingWith(err).then(a)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidPromiseInCallback", Message: avoidPromiseInCallbackMessage, Line: 1}},
			},
		},
	)
}
