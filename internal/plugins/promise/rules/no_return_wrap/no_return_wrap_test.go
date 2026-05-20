package no_return_wrap_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/no_return_wrap"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const rejectMessage = "Expected throw instead of Promise.reject"
const resolveMessage = "Avoid wrapping return values in Promise.resolve"

func TestNoReturnWrap(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_return_wrap.NoReturnWrapRule,
		[]rule_tester.ValidTestCase{
			// ---- ESLint upstream valid cases ----
			{Code: `Promise.resolve(4).then(function(x) { return x })`},
			{Code: `Promise.reject(4).then(function(x) { return x })`},
			{Code: `Promise.resolve(4).then(function() {})`},
			{Code: `Promise.reject(4).then(function() {})`},
			{Code: `doThing().then(function() { return 4 })`},
			{Code: `doThing().then(function() { throw 4 })`},
			{Code: `doThing().then(null, function() { return 4 })`},
			{Code: `doThing().then(null, function() { throw 4 })`},
			{Code: `doThing().catch(null, function() { return 4 })`},
			{Code: `doThing().catch(null, function() { throw 4 })`},
			{Code: `doThing().then(function() { return Promise.all([a,b,c]) })`},
			{Code: `doThing().then(() => 4)`},
			{Code: `doThing().then(() => { throw 4 })`},
			{Code: `doThing().then(()=>{}, () => 4)`},
			{Code: `doThing().then(()=>{}, () => { throw 4 })`},
			{Code: `doThing().catch(() => 4)`},
			{Code: `doThing().catch(() => { throw 4 })`},
			{Code: `var x = function() { return Promise.resolve(4) }`},
			{Code: `function y() { return Promise.resolve(4) }`},
			{Code: `function then() { return Promise.reject() }`},
			{Code: `doThing(function(x) { return Promise.reject(x) })`},
			{Code: `doThing().then(function() { return })`},
			{Code: `doThing().then(function() { return Promise.reject(4) })`, Options: map[string]interface{}{"allowReject": true}},
			{Code: `doThing().then((function() { return Promise.resolve(4) }).toString())`},
			{Code: `doThing().then(() => Promise.reject(4))`, Options: map[string]interface{}{"allowReject": true}},
			{Code: `doThing().then(function() { return a() })`},
			{Code: `doThing().then(function() { return Promise.a() })`},
			{Code: `doThing().then(() => { return a() })`},
			{Code: `doThing().then(() => { return Promise.a() })`},
			{Code: `doThing().then(() => a())`},
			{Code: `doThing().then(() => Promise.a())`},

			// ---- Upstream semantic branches and tsgo shape locks ----
			// Locks in upstream checkCallExpression(): computed Promise['resolve'] is ignored.
			{Code: `doThing().then(function() { return Promise['resolve'](4) })`},
			// Locks in upstream isPromise(): computed then/catch access is ignored.
			{Code: `doThing()['then'](function() { return Promise.resolve(4) })`},
			// Promise.withResolvers itself is intentionally not considered a promise root upstream.
			{Code: `Promise.withResolvers(function() { return Promise.resolve(4) })`},
			// Nested non-promise callback must not be reported even inside a promise callback.
			{Code: `doThing().then(function() { fn(function() { return Promise.resolve(4) }); return 1 })`},
			// Optional chain: Promise?.resolve is not a reportable wrap.
			{Code: `doThing().then(function() { return Promise?.resolve(4) })`},

			// ---- Edge cases: optional chains ----
			// Optional call (?.()): QuestionDotToken is on the CallExpression, not the PropertyAccessExpression.
			{Code: `doThing().then(function() { return Promise.resolve?.() })`},
			// Optional-chain reject in arrow expression body.
			{Code: `doThing().then(() => Promise?.reject(4))`},

			// ---- Edge cases: parens + non-bind member ----
			// .toString() wrapped in parens: not bind, must not be treated as a promise argument.
			{Code: `doThing().then(((function() { return Promise.resolve(4) }).toString()))`},

			// ---- Edge cases: allowReject + parens ----
			// Parenthesized function with allowReject:true -> valid.
			{Code: `doThing().then((function() { return Promise.reject(4) }))`, Options: map[string]interface{}{"allowReject": true}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- ESLint upstream invalid cases ----
			{Code: `doThing().then(function() { return Promise.resolve(4) })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "resolve", Message: resolveMessage}}},
			{Code: `doThing().then(null, function() { return Promise.resolve(4) })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "resolve", Message: resolveMessage}}},
			{Code: `doThing().catch(function() { return Promise.resolve(4) })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "resolve", Message: resolveMessage}}},
			{Code: `doThing().then(function() { return Promise.reject(4) })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "reject", Message: rejectMessage}}},
			{Code: `doThing().then(null, function() { return Promise.reject(4) })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "reject", Message: rejectMessage}}},
			{Code: `doThing().catch(function() { return Promise.reject(4) })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "reject", Message: rejectMessage}}},
			{Code: `doThing().then(function(x) { if (x>1) { return Promise.resolve(4) } else { throw "bad" } })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "resolve", Message: resolveMessage}}},
			{Code: `doThing().then(function(x) { if (x>1) { return Promise.reject(4) } })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "reject", Message: rejectMessage}}},
			{Code: `doThing().then(null, function() { if (true && false) { return Promise.resolve() } })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "resolve", Message: resolveMessage}}},
			{Code: `doThing().catch(function(x) {if (x) { return Promise.resolve(4) } else { return Promise.reject() } })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "resolve", Message: resolveMessage}, {MessageId: "reject", Message: rejectMessage}}},
			{Code: `
      fn(function() {
        doThing().then(function() {
          return Promise.resolve(4)
        })
        return
      })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "resolve", Message: resolveMessage, Line: 4}}},
			{Code: `
      fn(function() {
        doThing().then(function nm() {
          return Promise.resolve(4)
        })
        return
      })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "resolve", Message: resolveMessage, Line: 4}}},
			{Code: `
      fn(function() {
        fn2(function() {
          doThing().then(function() {
            return Promise.resolve(4)
          })
        })
      })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "resolve", Message: resolveMessage, Line: 5}}},
			{Code: `
      fn(function() {
        fn2(function() {
          doThing().then(function() {
            fn3(function() {
              return Promise.resolve(4)
            })
            return Promise.resolve(4)
          })
        })
      })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "resolve", Message: resolveMessage, Line: 8}}},
			{Code: `
      const o = {
        fn: function() {
          return doThing().then(function() {
            return Promise.resolve(5);
          });
        },
      }
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "resolve", Message: resolveMessage, Line: 5}}},
			{Code: `
      fn(
        doThing().then(function() {
          return Promise.resolve(5);
        })
      );
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "resolve", Message: resolveMessage, Line: 4}}},
			{Code: `doThing().then((function() { return Promise.resolve(4) }).bind(this))`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "resolve", Message: resolveMessage}}},
			{Code: `doThing().then((function() { return Promise.resolve(4) }).bind(this).bind(this))`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "resolve", Message: resolveMessage}}},
			{Code: `doThing().then(() => { return Promise.resolve(4) })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "resolve", Message: resolveMessage}}},
			{Code: `
      function a () {
        return p.then(function(val) {
          return Promise.resolve(val * 4)
        })
      }
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "resolve", Message: resolveMessage}}},
			{Code: `doThing().then(() => Promise.resolve(4))`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "resolve", Message: resolveMessage}}},
			{Code: `doThing().then(() => Promise.reject(4))`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "reject", Message: rejectMessage}}},

			// Promise static roots other than resolve/reject are promise-like upstream.
			{Code: `Promise.all(xs).then(function() { return Promise.resolve(4) })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "resolve", Message: resolveMessage}}},
			// Locks in upstream isPromise(): .then() is promise-like regardless of the receiver.
			{Code: `Promise.withResolvers().then(function() { return Promise.resolve(4) })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "resolve", Message: resolveMessage}}},

			// ---- Parenthesized-callback false-negative fixes ----
			// Single parens around function expression.
			{Code: `doThing().then((function() { return Promise.resolve(4) }))`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "resolve", Message: resolveMessage}}},
			// Double parens around function expression.
			{Code: `doThing().then(((function() { return Promise.resolve(4) })))`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "resolve", Message: resolveMessage}}},
			// Parens around arrow function.
			{Code: `doThing().then((() => Promise.resolve(4)))`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "resolve", Message: resolveMessage}}},
			// Parens around the whole bind() call.
			{Code: `doThing().then(((function() { return Promise.resolve(4) }).bind(this)))`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "resolve", Message: resolveMessage}}},
			// Parens around the .bind member access itself.
			{Code: `doThing().then(((function() { return Promise.resolve(4) }).bind)(this))`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "resolve", Message: resolveMessage}}},

			// ---- Edge cases: parenthesized callbacks (invalid) ----
			// Parenthesized arrow + reject.
			{Code: `doThing().then((() => Promise.reject(4)))`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "reject", Message: rejectMessage}}},
			// Parenthesized callback in finally.
			{Code: `doThing().finally((function() { return Promise.resolve(4) }))`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "resolve", Message: resolveMessage}}},
			// Parenthesized callback in catch + reject.
			{Code: `doThing().catch((function() { return Promise.reject(4) }))`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "reject", Message: rejectMessage}}},
			// Chained bind + outer parens.
			{Code: `doThing().then(((function() { return Promise.resolve(4) }).bind(this).bind(this)))`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "resolve", Message: resolveMessage}}},
		},
	)
}
