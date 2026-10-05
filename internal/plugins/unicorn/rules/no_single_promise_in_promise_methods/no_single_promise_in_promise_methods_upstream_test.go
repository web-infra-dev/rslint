// Ported from eslint-plugin-unicorn v77.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/test/no-single-promise-in-promise-methods.js
// Documentation examples are kept in their own group below.
package no_single_promise_in_promise_methods_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_single_promise_in_promise_methods"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoSinglePromiseInPromiseMethodsUpstreamAwaited(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_single_promise_in_promise_methods.NoSinglePromiseInPromiseMethodsRule,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			{
				Code:     "await Promise.race([(0, promise)])",
				FileName: "case.mjs",
				Output:   []string{"await (0, promise)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 34,
					},
				},
			},
			{
				Code:     "async function * foo() {await Promise.race([yield promise])}",
				FileName: "case.mjs",
				Output:   []string{"async function * foo() {await (yield promise)}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 44, EndLine: 1, EndColumn: 59,
					},
				},
			},
			{
				Code:     "async function * foo() {await Promise.race([yield* promise])}",
				FileName: "case.mjs",
				Output:   []string{"async function * foo() {await (yield* promise)}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 44, EndLine: 1, EndColumn: 60,
					},
				},
			},
			{
				Code:     "await Promise.race([() => promise,],)",
				FileName: "case.mjs",
				Output:   []string{"await (() => promise)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 36,
					},
				},
			},
			{
				Code:     "await Promise.race([a ? b : c,],)",
				FileName: "case.mjs",
				Output:   []string{"await (a ? b : c)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 32,
					},
				},
			},
			{
				Code:     "await Promise.race([x ??= y,],)",
				FileName: "case.mjs",
				Output:   []string{"await (x ??= y)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 30,
					},
				},
			},
			{
				Code:     "await Promise.race([x ||= y,],)",
				FileName: "case.mjs",
				Output:   []string{"await (x ||= y)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 30,
					},
				},
			},
			{
				Code:     "await Promise.race([x &&= y,],)",
				FileName: "case.mjs",
				Output:   []string{"await (x &&= y)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 30,
					},
				},
			},
			{
				Code:     "await Promise.race([x |= y,],)",
				FileName: "case.mjs",
				Output:   []string{"await (x |= y)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 29,
					},
				},
			},
			{
				Code:     "await Promise.race([x ^= y,],)",
				FileName: "case.mjs",
				Output:   []string{"await (x ^= y)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 29,
					},
				},
			},
			{
				Code:     "await Promise.race([x | y,],)",
				FileName: "case.mjs",
				Output:   []string{"await (x | y)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 28,
					},
				},
			},
			{
				Code:     "await Promise.race([x ^ y,],)",
				FileName: "case.mjs",
				Output:   []string{"await (x ^ y)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 28,
					},
				},
			},
			{
				Code:     "await Promise.race([x & y,],)",
				FileName: "case.mjs",
				Output:   []string{"await (x & y)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 28,
					},
				},
			},
			{
				Code:     "await Promise.race([x !== y,],)",
				FileName: "case.mjs",
				Output:   []string{"await (x !== y)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 30,
					},
				},
			},
			{
				Code:     "await Promise.race([x == y,],)",
				FileName: "case.mjs",
				Output:   []string{"await (x == y)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 29,
					},
				},
			},
			{
				Code:     "await Promise.race([x in y,],)",
				FileName: "case.mjs",
				Output:   []string{"await (x in y)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 29,
					},
				},
			},
			{
				Code:     "await Promise.race([x >>> y,],)",
				FileName: "case.mjs",
				Output:   []string{"await (x >>> y)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 30,
					},
				},
			},
			{
				Code:     "await Promise.race([x + y,],)",
				FileName: "case.mjs",
				Output:   []string{"await (x + y)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 28,
					},
				},
			},
			{
				Code:     "await Promise.race([x / y,],)",
				FileName: "case.mjs",
				Output:   []string{"await (x / y)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 28,
					},
				},
			},
			{
				Code:     "await Promise.race([x ** y,],)",
				FileName: "case.mjs",
				Output:   []string{"await (x ** y)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 29,
					},
				},
			},
			{
				Code:     "await Promise.race([promise,],)",
				FileName: "case.mjs",
				Output:   []string{"await promise"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 30,
					},
				},
			},
			{
				Code:     "await Promise.race([getPromise(),],)",
				FileName: "case.mjs",
				Output:   []string{"await getPromise()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 35,
					},
				},
			},
			{
				Code:     "await Promise.race([promises[0],],)",
				FileName: "case.mjs",
				Output:   []string{"await promises[0]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 34,
					},
				},
			},
			{
				Code:     "await Promise.race([await promise])",
				FileName: "case.mjs",
				Output:   []string{"await await promise"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 35,
					},
				},
			},
			{
				Code:     "await Promise.any([promise])",
				FileName: "case.mjs",
				Output:   []string{"await promise"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.any()` is unnecessary.",
						Line: 1, Column: 19, EndLine: 1, EndColumn: 28,
					},
				},
			},
			{
				Code:     "await Promise.any([promise,],)",
				FileName: "case.mjs",
				Output:   []string{"await promise"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.any()` is unnecessary.",
						Line: 1, Column: 19, EndLine: 1, EndColumn: 29,
					},
				},
			},
			{
				Code:     "await Promise.race([promise])",
				FileName: "case.mjs",
				Output:   []string{"await promise"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 29,
					},
				},
			},
			{
				Code:     "await Promise.race([new Promise(() => {})])",
				FileName: "case.mjs",
				Output:   []string{"await new Promise(() => {})"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 43,
					},
				},
			},
			{
				Code:     "+await Promise.race([+1])",
				FileName: "case.mjs",
				Output:   []string{"+await +1"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 21, EndLine: 1, EndColumn: 25,
					},
				},
			},
			{
				Code:     "await Promise.race([(x,y)])\n[0].toString()",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 27,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "await (x,y)\n[0].toString()"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "await Promise.resolve((x,y))\n[0].toString()"},
						},
					},
				},
			},
		})
}

func TestNoSinglePromiseInPromiseMethodsUpstreamNotAwaited(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_single_promise_in_promise_methods.NoSinglePromiseInPromiseMethodsRule,
		[]rule_tester.ValidTestCase{
			{Code: "Promise.race([promise, anotherPromise])", FileName: "case.mjs"},
			{Code: "Promise.race(notArrayLiteral)", FileName: "case.mjs"},
			{Code: "Promise.race([...promises])", FileName: "case.mjs"},
			{Code: "Promise.any([promise, anotherPromise])", FileName: "case.mjs"},
			{Code: "Promise.notListedMethod([promise])", FileName: "case.mjs"},
			{Code: "Promise[race]([promise])", FileName: "case.mjs"},
			{Code: "Promise.race([,])", FileName: "case.mjs"},
			{Code: "NotPromise.race([promise])", FileName: "case.mjs"},
			{Code: "Promise?.race([promise])", FileName: "case.mjs"},
			{Code: "Promise.race?.([promise])", FileName: "case.mjs"},
			{Code: "Promise.race(...[promise])", FileName: "case.mjs"},
			{Code: "Promise.race([promise], extraArguments)", FileName: "case.mjs"},
			{Code: "Promise.race()", FileName: "case.mjs"},
			{Code: "new Promise.race([promise])", FileName: "case.mjs"},
			{Code: "globalThis.Promise.race([promise])", FileName: "case.mjs"},
			{Code: "Promise[\"race\"]([promise])", FileName: "case.mjs"},
			{Code: "Promise.allSettled([promise])", FileName: "case.mjs"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     "Promise.race([promise,],)",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 14, EndLine: 1, EndColumn: 24,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "promise"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "Promise.resolve(promise,)"},
						},
					},
				},
			},
			{
				Code:     "foo\nPromise.race([(0, promise),],)",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 2, Column: 14, EndLine: 2, EndColumn: 29,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "foo\n;(0, promise)"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "foo\nPromise.resolve((0, promise),)"},
						},
					},
				},
			},
			{
				Code:     "foo\nPromise.race([[array][0],],)",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 2, Column: 14, EndLine: 2, EndColumn: 27,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "foo\n;[array][0]"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "foo\nPromise.resolve([array][0],)"},
						},
					},
				},
			},
			{
				Code:     "Promise.race([promise]).then()",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 14, EndLine: 1, EndColumn: 23,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "promise.then()"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "Promise.resolve(promise).then()"},
						},
					},
				},
			},
			{
				Code:     "Promise.race([1]).then()",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 14, EndLine: 1, EndColumn: 17,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "(1).then()"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "Promise.resolve(1).then()"},
						},
					},
				},
			},
			{
				Code:     "Promise.race([1.]).then()",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 14, EndLine: 1, EndColumn: 18,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "(1.).then()"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "Promise.resolve(1.).then()"},
						},
					},
				},
			},
			{
				Code:     "Promise.race([.1]).then()",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 14, EndLine: 1, EndColumn: 18,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "(.1).then()"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "Promise.resolve(.1).then()"},
						},
					},
				},
			},
			{
				Code:     "Promise.race([(0, promise)]).then()",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 14, EndLine: 1, EndColumn: 28,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "(0, promise).then()"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "Promise.resolve((0, promise)).then()"},
						},
					},
				},
			},
			{
				Code:     "const _ = () => Promise.race([ a ?? b ,],)",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 30, EndLine: 1, EndColumn: 41,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "const _ = () => (a ?? b)"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "const _ = () => Promise.resolve( a ?? b ,)"},
						},
					},
				},
			},
			{
				Code:     "Promise.race([ {a} = 1 ,],)",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 14, EndLine: 1, EndColumn: 26,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "({a} = 1)"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "Promise.resolve( {a} = 1 ,)"},
						},
					},
				},
			},
			{
				Code:     "Promise.race([ function () {} ,],)",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 14, EndLine: 1, EndColumn: 33,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "(function () {})"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "Promise.resolve( function () {} ,)"},
						},
					},
				},
			},
			{
				Code:     "Promise.race([ class {} ,],)",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 14, EndLine: 1, EndColumn: 27,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "(class {})"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "Promise.resolve( class {} ,)"},
						},
					},
				},
			},
			{
				Code:     "Promise.race([ new Foo ,],).then()",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 14, EndLine: 1, EndColumn: 26,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "(new Foo).then()"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "Promise.resolve( new Foo ,).then()"},
						},
					},
				},
			},
			{
				Code:     "Promise.race([ new Foo ,],).toString",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 14, EndLine: 1, EndColumn: 26,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "(new Foo).toString"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "Promise.resolve( new Foo ,).toString"},
						},
					},
				},
			},
			{
				Code:     "foo(Promise.race([promise]))",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 18, EndLine: 1, EndColumn: 27,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "foo(promise)"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "foo(Promise.resolve(promise))"},
						},
					},
				},
			},
			{
				Code:     "Promise.any([promise])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.any()` is unnecessary.",
						Line: 1, Column: 13, EndLine: 1, EndColumn: 22,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "promise"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "Promise.resolve(promise)"},
						},
					},
				},
			},
			{
				Code:     "foo(Promise.any([promise]))",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.any()` is unnecessary.",
						Line: 1, Column: 17, EndLine: 1, EndColumn: 26,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "foo(promise)"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "foo(Promise.resolve(promise))"},
						},
					},
				},
			},
			{
				Code:     "const obj = {p: Promise.all([promise])}",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 29, EndLine: 1, EndColumn: 38,
					},
				},
			},
			{
				Code:     "Promise.race([promise]).foo = 1",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 14, EndLine: 1, EndColumn: 23,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "promise.foo = 1"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "Promise.resolve(promise).foo = 1"},
						},
					},
				},
			},
			{
				Code:     "Promise.race([promise])[0] ||= 1",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 14, EndLine: 1, EndColumn: 23,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "promise[0] ||= 1"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "Promise.resolve(promise)[0] ||= 1"},
						},
					},
				},
			},
			{
				Code:     "Promise.race([undefined]).then()",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 14, EndLine: 1, EndColumn: 25,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "undefined.then()"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "Promise.resolve(undefined).then()"},
						},
					},
				},
			},
			{
				Code:     "Promise.race([null]).then()",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 14, EndLine: 1, EndColumn: 20,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "(null).then()"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "Promise.resolve(null).then()"},
						},
					},
				},
			},
		})
}

func TestNoSinglePromiseInPromiseMethodsUpstreamPromiseAll(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_single_promise_in_promise_methods.NoSinglePromiseInPromiseMethodsRule,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			{
				Code:     "Promise.all([promise])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 13, EndLine: 1, EndColumn: 22,
					},
				},
			},
			{
				Code:     "await Promise.all([promise])",
				FileName: "case.mjs",
				Output:   []string{"await promise"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 19, EndLine: 1, EndColumn: 28,
					},
				},
			},
			{
				Code:     "const foo = () => Promise.all([promise])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 31, EndLine: 1, EndColumn: 40,
					},
				},
			},
			{
				Code:     "const foo = await Promise.all([promise])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 31, EndLine: 1, EndColumn: 40,
					},
				},
			},
			{
				Code:     "foo = await Promise.all([promise])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 25, EndLine: 1, EndColumn: 34,
					},
				},
			},
			{
				Code:     "const foo = await Promise.race([promise])",
				FileName: "case.mjs",
				Output:   []string{"const foo = await promise"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 32, EndLine: 1, EndColumn: 41,
					},
				},
			},
			{
				Code:     "const foo = () => Promise.race([promise])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 32, EndLine: 1, EndColumn: 41,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "const foo = () => promise"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "const foo = () => Promise.resolve(promise)"},
						},
					},
				},
			},
			{
				Code:     "foo = await Promise.race([promise])",
				FileName: "case.mjs",
				Output:   []string{"foo = await promise"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 26, EndLine: 1, EndColumn: 35,
					},
				},
			},
			{
				Code:     "const results = await Promise.any([promise])",
				FileName: "case.mjs",
				Output:   []string{"const results = await promise"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.any()` is unnecessary.",
						Line: 1, Column: 35, EndLine: 1, EndColumn: 44,
					},
				},
			},
			{
				Code:     "const results = await Promise.race([promise])",
				FileName: "case.mjs",
				Output:   []string{"const results = await promise"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 36, EndLine: 1, EndColumn: 45,
					},
				},
			},
			{
				Code:     "const [foo] = await Promise.all([promise])",
				FileName: "case.mjs",
				Output:   []string{"const foo = await promise"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 33, EndLine: 1, EndColumn: 42,
					},
				},
			},
			{
				Code:     "[foo] = await Promise.all([promise])",
				FileName: "case.mjs",
				Output:   []string{"foo = await promise"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 27, EndLine: 1, EndColumn: 36,
					},
				},
			},
			{
				Code:     "const foo = (await Promise.all([promise]))[0]",
				FileName: "case.mjs",
				Output:   []string{"const foo = await promise"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 32, EndLine: 1, EndColumn: 41,
					},
				},
			},
			{
				Code:     "const foo = (await Promise.all([promise]))[0.0]",
				FileName: "case.mjs",
				Output:   []string{"const foo = await promise"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 32, EndLine: 1, EndColumn: 41,
					},
				},
			},
			{
				Code:     "foo = (await Promise.all([promise]))[0]",
				FileName: "case.mjs",
				Output:   []string{"foo = await promise"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 26, EndLine: 1, EndColumn: 35,
					},
				},
			},
			{
				Code:     "const [foo] = await Promise.all([a ? b : c])",
				FileName: "case.mjs",
				Output:   []string{"const foo = await (a ? b : c)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 33, EndLine: 1, EndColumn: 44,
					},
				},
			},
			{
				Code:     "const [foo = bar] = await Promise.all([promise])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 39, EndLine: 1, EndColumn: 48,
					},
				},
			},
			{
				Code:     "const [...foo] = await Promise.all([promise])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 36, EndLine: 1, EndColumn: 45,
					},
				},
			},
			{
				Code:     "const [, foo] = await Promise.all([promise])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 35, EndLine: 1, EndColumn: 44,
					},
				},
			},
			{
				Code:     "const result = ([foo] = await Promise.all([promise]))",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 43, EndLine: 1, EndColumn: 52,
					},
				},
			},
			{
				Code:     "const foo = (await Promise.all([promise]))[1]",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 32, EndLine: 1, EndColumn: 41,
					},
				},
			},
			{
				Code:     "const foo = (await Promise.all([promise])) /* comment */ [0]",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 32, EndLine: 1, EndColumn: 41,
					},
				},
			},
			{
				Code:     "const [/* comment */ foo] = await Promise.all([promise])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 47, EndLine: 1, EndColumn: 56,
					},
				},
			},
			{
				Code:     "const [foo] = await Promise.all([/* comment */ promise])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 33, EndLine: 1, EndColumn: 56,
					},
				},
			},
			{
				Code:     "const [foo]: [Foo] = await Promise.all([promise])",
				FileName: "case.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 40, EndLine: 1, EndColumn: 49,
					},
				},
			},
			{
				Code:     "await Promise.race([/* comment */ promise])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 43,
					},
				},
			},
			{
				Code:     "await Promise.race([promise /* comment */])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 43,
					},
				},
			},
			{
				Code:     "await Promise.race([promise, /* comment */])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 44,
					},
				},
			},
			{
				Code:     "await Promise.race([\n\t// comment\n\tpromise,\n])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 4, EndColumn: 2,
					},
				},
			},
			{
				Code:     "await Promise.any([/* comment */ promise])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.any()` is unnecessary.",
						Line: 1, Column: 19, EndLine: 1, EndColumn: 42,
					},
				},
			},
			{
				Code:     "Promise.race([/* comment */ promise])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 14, EndLine: 1, EndColumn: 37,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "Promise.resolve(/* comment */ promise)"},
						},
					},
				},
			},
		})
}

func TestNoSinglePromiseInPromiseMethodsUpstreamDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_single_promise_in_promise_methods.NoSinglePromiseInPromiseMethodsRule,
		[]rule_tester.ValidTestCase{
			{Code: "const foo = await promise;", FileName: "case.mjs"},
			{Code: "const promise = Promise.resolve(nonPromise);", FileName: "case.mjs"},
			{Code: "const foo = await Promise.all(promises);", FileName: "case.mjs"},
			{Code: "const foo = await Promise.any([promise, anotherPromise]);", FileName: "case.mjs"},
			{Code: "const [{value: foo, reason: error}] = await Promise.allSettled([promise]);", FileName: "case.mjs"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     "const foo = await Promise.all([promise]);",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 31, EndLine: 1, EndColumn: 40,
					},
				},
			},
			{
				Code:     "const foo = await Promise.any([promise]);",
				FileName: "case.mjs",
				Output:   []string{"const foo = await promise;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.any()` is unnecessary.",
						Line: 1, Column: 31, EndLine: 1, EndColumn: 40,
					},
				},
			},
			{
				Code:     "const foo = await Promise.race([promise]);",
				FileName: "case.mjs",
				Output:   []string{"const foo = await promise;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 32, EndLine: 1, EndColumn: 41,
					},
				},
			},
			{
				Code:     "const promise = Promise.all([nonPromise]);",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 29, EndLine: 1, EndColumn: 41,
					},
				},
			},
		})
}
