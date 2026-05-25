package param_names_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/param_names"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const msgForResolve = `Promise constructor parameters must be named to match "^_?resolve$"`
const msgForReject = `Promise constructor parameters must be named to match "^_?reject$"`

func TestParamNames(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&param_names.ParamNamesRule,
		[]rule_tester.ValidTestCase{
			// ---- ESLint upstream valid cases ----
			{Code: `new Promise(function(resolve, reject) {})`},
			{Code: `new Promise(function(resolve, _reject) {})`},
			{Code: `new Promise(function(_resolve, reject) {})`},
			{Code: `new Promise(function(_resolve, _reject) {})`},
			{Code: `new Promise(function(resolve) {})`},
			{Code: `new Promise(function(_resolve) {})`},
			{Code: `new Promise(resolve => {})`},
			{Code: `new Promise((resolve, reject) => {})`},
			{Code: `new Promise(() => {})`},
			{Code: `new NonPromise()`},
			{
				Code:    `new Promise((yes, no) => {})`,
				Options: map[string]interface{}{"resolvePattern": "^yes$", "rejectPattern": "^no$"},
			},

			// ---- Patterns/defaults/rest skip (mirrors ESLint `.name===undefined`) ----
			{Code: `new Promise(function({ resolve, reject }) {})`},
			{Code: `new Promise(function([resolve, reject]) {})`},
			{Code: `new Promise(function(resolve = () => {}, reject = () => {}) {})`},
			{Code: `new Promise(function(...args) {})`},
			{Code: `new Promise(function({ a }, { b }) {})`},
			// Mixed: first is identifier (checked), second is destructuring (skipped)
			{Code: `new Promise(function(resolve, { foo }) {})`},
			// Mixed: first is destructuring (skipped), second is identifier (checked)
			{Code: `new Promise(function({ foo }, reject) {})`},

			// ---- Not a Promise constructor invocation ----
			{Code: `new Promise(handler)`},
			{Code: `new Promise(function(reject, resolve) {}, extraArg)`},
			{Code: `Promise(function(reject, resolve) {})`},
			{Code: `new Foo.Promise(function(reject, resolve) {})`},
			{Code: `new globalThis.Promise(function(reject, resolve) {})`},
			{Code: `new Promise()`},
			{Code: `new Promise(123)`},
			{Code: `new Promise("abc")`},

			// ---- Async executor (FunctionExpression / ArrowFunction) ----
			{Code: `new Promise(async function(resolve, reject) {})`},
			{Code: `new Promise(async (resolve, reject) => {})`},

			// ---- TS type assertion on callee / executor: rule silently skips,
			//      mirroring ESLint's `callee.type !== 'Identifier'` short-circuit.
			//      Bad names here are NOT reported (aligns with eslint-plugin-promise
			//      under @typescript-eslint/parser).
			{Code: `new (Promise as any)(function(resolve, reject) {})`},
			{Code: `new (Promise as any)(function(ok, fail) {})`},
			{Code: `new (<any>Promise)(function(ok, fail) {})`},
			{Code: `new Promise((function(ok, fail) {}) as any)`},

			// ---- Partial options: other field falls back to default ----
			{
				Code:    `new Promise((yes, reject) => {})`,
				Options: map[string]interface{}{"resolvePattern": "^yes$"},
			},
			{
				Code:    `new Promise((resolve, no) => {})`,
				Options: map[string]interface{}{"rejectPattern": "^no$"},
			},

			// ---- regexp2 / ECMAScript features (RE2 can't handle these) ----
			// Lookbehind
			{
				Code:    `new Promise((resolve) => {})`,
				Options: map[string]interface{}{"resolvePattern": "(?<!_)resolve"},
			},
			// Unicode property escape
			{
				Code:    `new Promise((resolve) => {})`,
				Options: map[string]interface{}{"resolvePattern": `^\p{Ll}+$`},
			},

			// ---- Invalid pattern: rule silently no-ops (can't match) ----
			{
				Code:    `new Promise((reject, resolve) => {})`,
				Options: map[string]interface{}{"resolvePattern": "[unclosed"},
			},

			// ---- TS generic type argument on Promise ----
			{Code: `new Promise<number>((resolve, reject) => {})`},
			{Code: `new Promise<void>(function(resolve, reject) {})`},

			// ---- 3+ executor params: ESLint only checks first two ----
			{Code: `new Promise((resolve, reject, extra) => {})`},
			{Code: `new Promise((resolve, reject, a, b, c) => {})`},

			// ---- Named function expression ----
			{Code: `new Promise(function named(resolve, reject) {})`},

			// ---- Empty params on FunctionExpression (not just arrow) ----
			{Code: `new Promise(function() {})`},

			// ---- Spread executor argument is not a FunctionExpression ----
			{Code: `new Promise(...args)`},

			// ---- TS `this` parameter is stripped before resolve/reject indexing ----
			{Code: `new Promise<void>(function(this: any, resolve, reject) {})`},
			{Code: `new Promise<void>(function(this: unknown, _resolve) {})`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- ESLint upstream invalid cases ----
			{
				Code: `new Promise(function(reject, resolve) {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "resolveParamNames",
						Message:   msgForResolve,
						Line:      1,
						Column:    22,
						EndLine:   1,
						EndColumn: 28,
					},
					{
						MessageId: "rejectParamNames",
						Message:   msgForReject,
						Line:      1,
						Column:    30,
						EndLine:   1,
						EndColumn: 37,
					},
				},
			},
			{
				Code: `new Promise(function(resolve, rej) {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "rejectParamNames",
						Message:   msgForReject,
						Line:      1,
						Column:    31,
						EndLine:   1,
						EndColumn: 34,
					},
				},
			},
			{
				Code: `new Promise(yes => {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "resolveParamNames",
						Message:   msgForResolve,
						Line:      1,
						Column:    13,
						EndLine:   1,
						EndColumn: 16,
					},
				},
			},
			{
				Code: `new Promise((yes, no) => {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "resolveParamNames",
						Message:   msgForResolve,
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 17,
					},
					{
						MessageId: "rejectParamNames",
						Message:   msgForReject,
						Line:      1,
						Column:    19,
						EndLine:   1,
						EndColumn: 21,
					},
				},
			},
			{
				Code:    `new Promise(function(resolve, reject) {})`,
				Options: map[string]interface{}{"resolvePattern": "^yes$", "rejectPattern": "^no$"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "resolveParamNames",
						Message:   `Promise constructor parameters must be named to match "^yes$"`,
						Line:      1,
						Column:    22,
						EndLine:   1,
						EndColumn: 29,
					},
					{
						MessageId: "rejectParamNames",
						Message:   `Promise constructor parameters must be named to match "^no$"`,
						Line:      1,
						Column:    31,
						EndLine:   1,
						EndColumn: 37,
					},
				},
			},

			// ---- Single underscore is NOT acceptable under default pattern ----
			{
				Code: `new Promise(function(_, reject) {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "resolveParamNames",
						Message:   msgForResolve,
						Line:      1,
						Column:    22,
						EndLine:   1,
						EndColumn: 23,
					},
				},
			},

			// ---- Parenthesized Promise identifier is still recognized ----
			{
				Code: `new (Promise)(function(ok, fail) {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "resolveParamNames",
						Message:   msgForResolve,
						Line:      1,
						Column:    24,
						EndLine:   1,
						EndColumn: 26,
					},
					{
						MessageId: "rejectParamNames",
						Message:   msgForReject,
						Line:      1,
						Column:    28,
						EndLine:   1,
						EndColumn: 32,
					},
				},
			},

			// ---- Async function executor ----
			{
				Code: `new Promise(async function(ok, fail) {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "resolveParamNames",
						Message:   msgForResolve,
						Line:      1,
						Column:    28,
						EndLine:   1,
						EndColumn: 30,
					},
					{
						MessageId: "rejectParamNames",
						Message:   msgForReject,
						Line:      1,
						Column:    32,
						EndLine:   1,
						EndColumn: 36,
					},
				},
			},
			{
				Code: `new Promise(async (ok, fail) => {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "resolveParamNames",
						Message:   msgForResolve,
						Line:      1,
						Column:    20,
						EndLine:   1,
						EndColumn: 22,
					},
					{
						MessageId: "rejectParamNames",
						Message:   msgForReject,
						Line:      1,
						Column:    24,
						EndLine:   1,
						EndColumn: 28,
					},
				},
			},

			// ---- Partial options: only resolvePattern changed ----
			{
				Code:    `new Promise((no, rejectX) => {})`,
				Options: map[string]interface{}{"resolvePattern": "^yes$"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "resolveParamNames",
						Message:   `Promise constructor parameters must be named to match "^yes$"`,
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 16,
					},
					{
						MessageId: "rejectParamNames",
						Message:   msgForReject,
						Line:      1,
						Column:    18,
						EndLine:   1,
						EndColumn: 25,
					},
				},
			},

			// ---- Partial options: only rejectPattern changed ----
			{
				Code:    `new Promise((resolveX, yes) => {})`,
				Options: map[string]interface{}{"rejectPattern": "^no$"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "resolveParamNames",
						Message:   msgForResolve,
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 22,
					},
					{
						MessageId: "rejectParamNames",
						Message:   `Promise constructor parameters must be named to match "^no$"`,
						Line:      1,
						Column:    24,
						EndLine:   1,
						EndColumn: 27,
					},
				},
			},

			// ---- Pattern with backslashes: message must preserve the raw source
			//      (ESLint uses `regex.source`; our formatter must NOT re-escape `\`).
			{
				Code:    `new Promise((bad) => {})`,
				Options: map[string]interface{}{"resolvePattern": `^\w+resolve$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "resolveParamNames",
						Message:   `Promise constructor parameters must be named to match "^\w+resolve$"`,
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 17,
					},
				},
			},

			// ---- ECMAScript-only lookbehind pattern compiles under regexp2 ----
			{
				Code:    `new Promise((_resolve) => {})`,
				Options: map[string]interface{}{"resolvePattern": "(?<!_)resolve"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "resolveParamNames",
						Message:   `Promise constructor parameters must be named to match "(?<!_)resolve"`,
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 22,
					},
				},
			},

			// ---- Two-line executor body: column still 1-based on correct line ----
			{
				Code: "new Promise(function(ok, fail) {\n  return 0;\n})",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "resolveParamNames",
						Message:   msgForResolve,
						Line:      1,
						Column:    22,
						EndLine:   1,
						EndColumn: 24,
					},
					{
						MessageId: "rejectParamNames",
						Message:   msgForReject,
						Line:      1,
						Column:    26,
						EndLine:   1,
						EndColumn: 30,
					},
				},
			},

			// ---- TS generic type argument: still reports bad names ----
			{
				Code: `new Promise<number>((ok, fail) => {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "resolveParamNames",
						Message:   msgForResolve,
						Line:      1,
						Column:    22,
						EndLine:   1,
						EndColumn: 24,
					},
					{
						MessageId: "rejectParamNames",
						Message:   msgForReject,
						Line:      1,
						Column:    26,
						EndLine:   1,
						EndColumn: 30,
					},
				},
			},

			// ---- 3+ params: only first two are checked, third/fourth ignored ----
			{
				Code: `new Promise((ok, fail, also_wrong) => {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "resolveParamNames",
						Message:   msgForResolve,
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 16,
					},
					{
						MessageId: "rejectParamNames",
						Message:   msgForReject,
						Line:      1,
						Column:    18,
						EndLine:   1,
						EndColumn: 22,
					},
				},
			},

			// ---- TS `this` parameter is stripped; bad names in real params still reported ----
			{
				Code: `new Promise<void>(function(this: any, ok, fail) {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "resolveParamNames",
						Message:   msgForResolve,
						Line:      1,
						Column:    39,
						EndLine:   1,
						EndColumn: 41,
					},
					{
						MessageId: "rejectParamNames",
						Message:   msgForReject,
						Line:      1,
						Column:    43,
						EndLine:   1,
						EndColumn: 47,
					},
				},
			},

			// ---- Named function expression ----
			{
				Code: `new Promise(function named(ok, fail) {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "resolveParamNames",
						Message:   msgForResolve,
						Line:      1,
						Column:    28,
						EndLine:   1,
						EndColumn: 30,
					},
					{
						MessageId: "rejectParamNames",
						Message:   msgForReject,
						Line:      1,
						Column:    32,
						EndLine:   1,
						EndColumn: 36,
					},
				},
			},

			// ---- Params split across lines (multi-line position) ----
			{
				Code: "new Promise(function(\n  ok,\n  fail\n) {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "resolveParamNames",
						Message:   msgForResolve,
						Line:      2,
						Column:    3,
						EndLine:   2,
						EndColumn: 5,
					},
					{
						MessageId: "rejectParamNames",
						Message:   msgForReject,
						Line:      3,
						Column:    3,
						EndLine:   3,
						EndColumn: 7,
					},
				},
			},
		},
	)
}
