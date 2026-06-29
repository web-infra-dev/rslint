package no_callback_in_promise_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/no_callback_in_promise"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const callbackMessage = "Avoid calling back inside of a promise."
const callbackMsgId = "callback"

func cbErr(col int) rule_tester.InvalidTestCaseError {
	e := rule_tester.InvalidTestCaseError{MessageId: callbackMsgId, Message: callbackMessage}
	if col > 0 {
		e.Column = col
	}
	return e
}

// TestNoCallbackInPromiseUpstream migrates every valid/invalid case from the upstream
// eslint-plugin-promise test file for no-callback-in-promise.
func TestNoCallbackInPromiseUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_callback_in_promise.NoCallbackInPromiseRule,
		[]rule_tester.ValidTestCase{
			// ---- ESLint upstream valid cases ----
			{Code: `function thing(cb) { cb() }`},
			{Code: `doSomething(function(err) { cb(err) })`},
			{Code: `function thing(callback) { callback() }`},
			{Code: `doSomething(function(err) { callback(err) })`},

			// Support safe callbacks (#220) — timeoutsErr defaults to false
			{Code: `whatever.then((err) => { process.nextTick(() => cb()) })`},
			{Code: `whatever.then((err) => { setImmediate(() => cb()) })`},
			{Code: `whatever.then((err) => setImmediate(() => cb()))`},
			{Code: `whatever.then((err) => process.nextTick(() => cb()))`},
			{Code: `whatever.then((err) => process.nextTick(cb))`},
			{Code: `whatever.then((err) => setImmediate(cb))`},

			// Arrow functions and other things
			{Code: `let thing = (cb) => cb()`},
			{Code: `doSomething(err => cb(err))`},

			// Exceptions option
			{Code: `a.then(() => next())`, Options: map[string]interface{}{"exceptions": []interface{}{"next"}}},
			{Code: `a.then(() => next()).catch((err) => next(err))`, Options: map[string]interface{}{"exceptions": []interface{}{"next"}}},
			{Code: `a.then(next)`, Options: map[string]interface{}{"exceptions": []interface{}{"next"}}},
			{Code: `a.then(next).catch(next)`, Options: map[string]interface{}{"exceptions": []interface{}{"next"}}},

			// #572 — while-loop with next inside non-promise iterator call
			{Code: `while (!(step = call(next, iterator)).done) {
       if (result !== undefined) break;
     }`},
			// #572 comment — function with callback parameter used only as a regular argument
			{Code: `function hasCallbackArg(callback) {
       console.log(callback);
     }`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- ESLint upstream invalid cases ----

			// cb directly passed to .then / .catch
			{Code: `a.then(cb)`, Errors: []rule_tester.InvalidTestCaseError{cbErr(8)}},
			{Code: `a.then(() => cb())`, Errors: []rule_tester.InvalidTestCaseError{cbErr(0)}},
			{Code: `a.then(function(err) { cb(err) })`, Errors: []rule_tester.InvalidTestCaseError{cbErr(24)}},
			{Code: `a.then(function(data) { cb(data) }, function(err) { cb(err) })`,
				Errors: []rule_tester.InvalidTestCaseError{cbErr(25), cbErr(53)}},
			{Code: `a.catch(function(err) { cb(err) })`, Errors: []rule_tester.InvalidTestCaseError{cbErr(0)}},

			// "callback" name also flagged
			{Code: `a.then(callback)`, Errors: []rule_tester.InvalidTestCaseError{cbErr(8)}},
			{Code: `a.then(() => callback())`, Errors: []rule_tester.InvalidTestCaseError{cbErr(0)}},
			{Code: `a.then(function(err) { callback(err) })`, Errors: []rule_tester.InvalidTestCaseError{cbErr(24)}},
			{Code: `a.then(function(data) { callback(data) }, function(err) { callback(err) })`,
				Errors: []rule_tester.InvalidTestCaseError{cbErr(0), cbErr(59)}},
			{Code: `a.catch(function(err) { callback(err) })`, Errors: []rule_tester.InvalidTestCaseError{cbErr(0)}},

			// #167 — timeoutsErr: true cases
			{
				Code: `
        function wait (callback) {
          return Promise.resolve()
            .then(() => {
              setTimeout(callback);
            });
        }
      `,
				Errors:  []rule_tester.InvalidTestCaseError{cbErr(0)},
				Options: map[string]interface{}{"timeoutsErr": true},
			},
			{
				Code: `
        function wait (callback) {
          return Promise.resolve()
            .then(() => {
              setTimeout(() => callback());
            });
        }
      `,
				Errors:  []rule_tester.InvalidTestCaseError{cbErr(0)},
				Options: map[string]interface{}{"timeoutsErr": true},
			},

			// timeoutsErr: true — timeout functions inside promise handlers
			{Code: `whatever.then((err) => { process.nextTick(() => cb()) })`,
				Errors:  []rule_tester.InvalidTestCaseError{cbErr(0)},
				Options: map[string]interface{}{"timeoutsErr": true}},
			{Code: `whatever.then((err) => { setImmediate(() => cb()) })`,
				Errors:  []rule_tester.InvalidTestCaseError{cbErr(0)},
				Options: map[string]interface{}{"timeoutsErr": true}},
			{Code: `whatever.then((err) => setImmediate(() => cb()))`,
				Errors:  []rule_tester.InvalidTestCaseError{cbErr(0)},
				Options: map[string]interface{}{"timeoutsErr": true}},
			{Code: `whatever.then((err) => process.nextTick(() => cb()))`,
				Errors:  []rule_tester.InvalidTestCaseError{cbErr(0)},
				Options: map[string]interface{}{"timeoutsErr": true}},
			{Code: `whatever.then((err) => process.nextTick(cb))`,
				Errors:  []rule_tester.InvalidTestCaseError{cbErr(0)},
				Options: map[string]interface{}{"timeoutsErr": true}},
			{Code: `whatever.then((err) => setImmediate(cb))`,
				Errors:  []rule_tester.InvalidTestCaseError{cbErr(0)},
				Options: map[string]interface{}{"timeoutsErr": true}},
		},
	)
}
