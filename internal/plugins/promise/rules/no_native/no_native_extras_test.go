package no_native_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/no_native"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNativeExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_native.NoNativeRule,
		[]rule_tester.ValidTestCase{
			// ---- Local/source declarations are allowed ----
			{Code: `const Promise = require("bluebird"); Promise.resolve(1);`},
			{Code: `let Promise = getPromiseCtor(); new Promise(function(resolve) { resolve(1); });`},
			{Code: `function f(Promise) { return Promise.resolve(1); }`},
			{Code: `const f = (Promise) => Promise.resolve(1);`},
			{Code: `try { throw getPromiseCtor(); } catch (Promise) { Promise.resolve(1); }`},
			{Code: `const { Promise } = registry; Promise.resolve(1);`},
			{Code: `const { Promise: LocalPromise } = registry; LocalPromise.resolve(1);`},
			{Code: `class Promise {}; new Promise();`},

			// ---- Non-reference identifiers named Promise are ignored ----
			{Code: `const obj = { Promise: 1 }; obj.Promise;`},
			{Code: `const obj = { Promise() { return 1; } }; obj.Promise();`},
			{Code: `type Promise = string; let x: Promise;`},
			{Code: `type T = typeof Promise;`},
			{Code: `interface Promise { then(): void }`},
			{Code: `namespace Promise { export const value = 1 }`},
			{Code: `Promise: for (;;) { break Promise; }`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `Promise.all([]); Promise.resolve(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "name", Message: msgPromiseNotDefined, Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
					{MessageId: "name", Message: msgPromiseNotDefined, Line: 1, Column: 18, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:   `const obj = { Promise };`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: msgPromiseNotDefined, Line: 1, Column: 15, EndLine: 1, EndColumn: 22}},
			},
			{
				Code:   `function f() { return Promise.resolve(1); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: msgPromiseNotDefined, Line: 1, Column: 23, EndLine: 1, EndColumn: 30}},
			},
		},
	)
}
