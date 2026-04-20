package prefer_promise_reject_errors

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferPromiseRejectErrorsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferPromiseRejectErrorsRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			{Code: `Promise.resolve(5)`},
			{Code: `Foo.reject(5)`},
			{Code: `Promise.reject(foo)`},
			{Code: `Promise.reject(foo.bar)`},
			{Code: `Promise.reject(foo.bar())`},
			{Code: `Promise.reject(new Error())`},
			{Code: `Promise.reject(new TypeError)`},
			{Code: `Promise.reject(new Error('foo'))`},
			{Code: `Promise.reject(foo || 5)`},
			{Code: `Promise.reject(5 && foo)`},
			{Code: `new Foo((resolve, reject) => reject(5))`},
			{Code: `new Promise(function(resolve, reject) { return function(reject) { reject(5) } })`},
			{Code: `new Promise(function(resolve, reject) { if (foo) { const reject = somethingElse; reject(5) } })`},
			{Code: `new Promise(function(resolve, {apply}) { apply(5) })`},
			{Code: `new Promise(function(resolve, reject) { resolve(5, reject) })`},
			{Code: `async function foo() { Promise.reject(await foo); }`},
			{
				Code:    `Promise.reject()`,
				Options: map[string]interface{}{"allowEmptyReject": true},
			},
			{
				Code:    `new Promise(function(resolve, reject) { reject() })`,
				Options: map[string]interface{}{"allowEmptyReject": true},
			},

			// ---- Optional chaining ----
			{Code: `Promise.reject(obj?.foo)`},
			{Code: `Promise.reject(obj?.foo())`},

			// ---- Assignments ----
			{Code: `Promise.reject(foo = new Error())`},
			{Code: `Promise.reject(foo ||= 5)`},
			{Code: `Promise.reject(foo.bar ??= 5)`},
			{Code: `Promise.reject(foo[bar] ??= 5)`},

			// ---- Private fields ----
			{Code: `class C { #reject; foo() { Promise.#reject(5); } }`},
			{Code: `class C { #error; foo() { Promise.reject(this.#error); } }`},

			// ---- ESLint requires params[1].type === "Identifier"; non-plain
			// second-parameter shapes are not analyzed.
			{Code: `new Promise((resolve, reject = foo) => reject(5))`},
			{Code: `new Promise((resolve, ...reject) => reject[0](5))`},
			{Code: `new Promise(function(resolve, [reject]) { reject(5) })`},

			// ---- Identifiers that resolve to globals such as NaN / Infinity
			// are Identifier nodes, so couldBeError treats them as possibly Error.
			{Code: `Promise.reject(NaN)`},
			{Code: `Promise.reject(Infinity)`},

			// ---- Calling Promise without `new` is not an executor pattern. ----
			{Code: `Promise((resolve, reject) => reject(5))`},

			// ---- TaggedTemplateExpression result could be an Error. ----
			{Code: "Promise.reject(tag`msg`)"},

			// ---- Reject as argument, not callee, does not invoke it. ----
			{Code: `new Promise((resolve, reject) => arr.push(reject))`},

			// ---- Reject method-like access (.call / .bind / .apply) is treated
			// as a different operation by ESLint and is not flagged.
			{Code: `new Promise((resolve, reject) => reject.call(null, new Error()))`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// ---- TS assertion wrappers are NOT transparent in upstream ESLint ----
			// Verified: ESLint core run on a `.ts` file via `@typescript-eslint/parser`
			// reports each of these — TSAsExpression / TSTypeAssertion /
			// TSNonNullExpression / TSSatisfiesExpression are absent from
			// `astUtils.couldBeError` and fall through to its default branch.
			{
				Code: `Promise.reject(foo as Error)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code: `Promise.reject(<Error>foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code: `Promise.reject(foo!)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code: `Promise.reject(foo satisfies Error)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code: `Promise.reject(5)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code: `Promise.reject('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code: "Promise.reject(`foo`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code: `Promise.reject(!foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code: `Promise.reject(void foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code: `Promise.reject()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "rejectAnError",
						Message:   "Expected the Promise rejection reason to be an Error.",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: `Promise.reject(undefined)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code: `Promise.reject({ foo: 1 })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code: `Promise.reject([1, 2, 3])`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code:    `Promise.reject()`,
				Options: map[string]interface{}{"allowEmptyReject": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code:    `new Promise(function(resolve, reject) { reject() })`,
				Options: map[string]interface{}{"allowEmptyReject": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 41},
				},
			},
			{
				Code:    `Promise.reject(undefined)`,
				Options: map[string]interface{}{"allowEmptyReject": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code: `Promise.reject('foo', somethingElse)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code: `new Promise(function(resolve, reject) { reject(5) })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 41},
				},
			},
			{
				Code: `new Promise((resolve, reject) => { reject(5) })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 36},
				},
			},
			{
				Code: `new Promise((resolve, reject) => reject(5))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 34},
				},
			},
			{
				Code: `new Promise((resolve, reject) => reject())`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 34},
				},
			},
			{
				Code: `new Promise(function(yes, no) { no(5) })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 33},
				},
			},
			{
				Code: `
          new Promise((resolve, reject) => {
            fs.readFile('foo.txt', (err, file) => {
              if (err) reject('File not found')
              else resolve(file)
            })
          })
        `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 4, Column: 24},
				},
			},
			{
				Code: `new Promise(({foo, bar, baz}, reject) => reject(5))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 42},
				},
			},
			{
				Code: `new Promise(function(reject, reject) { reject(5) })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 40},
				},
			},
			{
				Code: `new Promise(function(foo, arguments) { arguments(5) })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 40},
				},
			},
			{
				Code: `new Promise((foo, arguments) => arguments(5))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 33},
				},
			},
			{
				Code: `new Promise(function({}, reject) { reject(5) })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 36},
				},
			},
			{
				Code: `new Promise(({}, reject) => reject(5))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 29},
				},
			},
			{
				Code: `new Promise((resolve, reject, somethingElse = reject(5)) => {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 47},
				},
			},

			// ---- `var` re-declaration in a non-arrow function body merges with
			// the parameter symbol (function-scope), so the call still reports.
			{
				Code: `new Promise(function(resolve, reject) { var reject = somethingElse; reject(5) })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 69},
				},
			},

			// ---- reject called from a nested callback inside the executor still
			// resolves to the executor's parameter binding.
			{
				Code: `new Promise((resolve, reject) => setTimeout(() => reject(5), 0))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 51},
				},
			},
			{
				Code: `new Promise((resolve, reject) => arr.forEach(function () { reject('bad') }))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 60},
				},
			},

			// ---- Spread cannot be statically classified as an Error candidate;
			// ESLint's couldBeError lacks a SpreadElement case so it falls through
			// to the default and reports.
			{
				Code: `Promise.reject(...args)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},

			// ---- Plain literals that are not Errors ----
			{
				Code: `Promise.reject(null)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code: `Promise.reject(0)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code: `Promise.reject(true)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},

			// ---- Computed bracket access with static string ----
			{
				Code: "Promise['reject'](5)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code: "Promise[`reject`](5)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},

			// ---- Named function expression as executor ----
			{
				Code: `new Promise(function fn(resolve, reject) { reject(5) })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 44},
				},
			},

			// ---- Async / generator executors are still functions, so reject
			// calls inside them should be checked.
			{
				Code: `new Promise(async (resolve, reject) => reject(5))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 40},
				},
			},
			{
				Code: `new Promise(function *(resolve, reject) { reject(5) })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 43},
				},
			},

			// ---- Multiple reject calls in a single executor ----
			{
				Code: `new Promise((resolve, reject) => { reject(1); reject('x'); reject(); })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 36},
					{MessageId: "rejectAnError", Line: 1, Column: 47},
					{MessageId: "rejectAnError", Line: 1, Column: 60},
				},
			},

			// ---- Parenthesized reject callee inside the executor body ----
			{
				Code: `new Promise((resolve, reject) => (reject)(5))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 34},
				},
			},

			// ---- Optional chaining on the executor reject callback ----
			{
				Code: `new Promise((resolve, reject) => reject?.(5))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 34},
				},
			},

			// ---- Parenthesized / TS-asserted Promise constructor must be
			// transparent (ESTree strips parens, so ESLint already accepts
			// `new (Promise)(executor)`).
			{
				Code: `new (Promise)((resolve, reject) => reject(5))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 36},
				},
			},
			{
				Code: `new Promise(((resolve, reject) => reject(5)))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 35},
				},
			},

			// ---- Optional chaining ----
			{
				Code: `Promise.reject?.(5)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code: `Promise?.reject(5)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code: `Promise?.reject?.(5)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code: `(Promise?.reject)(5)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code: `(Promise?.reject)?.(5)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},

			// ---- Mathematical / bitwise assignments evaluate to primitives or throw ----
			{
				Code: `Promise.reject(foo += new Error())`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code: `Promise.reject(foo -= new Error())`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code: `Promise.reject(foo **= new Error())`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code: `Promise.reject(foo <<= new Error())`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code: `Promise.reject(foo |= new Error())`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code: `Promise.reject(foo &= new Error())`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},

			// ---- && short-circuit yields the falsy left or the right operand ----
			{
				Code: `Promise.reject(foo && 5)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
			{
				Code: `Promise.reject(foo &&= 5)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "rejectAnError", Line: 1, Column: 1},
				},
			},
		},
	)
}
