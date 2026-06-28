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
			// ---- Local/source value declarations satisfy a value reference ----
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
			{Code: `namespace Promise { export const value = 1 }`},
			{Code: `Promise: for (;;) { break Promise; }`},

			// ---- Type references resolved against a local type binding ----
			{Code: `type Promise = string; let x: Promise;`},
			{Code: `interface Promise { then(): void } let x: Promise;`},
			{Code: `import { Promise } from "bluebird"; let x: Promise;`},
			{Code: `function f<Promise>(x: Promise) { return x; }`},
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

			// ---- typeof operand references the value, so the lib Promise reports ----
			{
				Code:   `type T = typeof Promise;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: msgPromiseNotDefined, Line: 1, Column: 17, EndLine: 1, EndColumn: 24}},
			},

			// ---- A type-only declaration does not satisfy a value reference ----
			{
				Code:   `interface Promise {} Promise.resolve(1);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: msgPromiseNotDefined, Line: 1, Column: 22, EndLine: 1, EndColumn: 29}},
			},

			// ---- A value declaration does not shadow a type reference ----
			{
				Code:   `var Promise = 1; let y: Promise;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: msgPromiseNotDefined, Line: 1, Column: 25, EndLine: 1, EndColumn: 32}},
			},

			// ---- Type references with no local type binding report ----
			{
				Code:   `let x: Promise;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: msgPromiseNotDefined, Line: 1, Column: 8, EndLine: 1, EndColumn: 15}},
			},
			{
				Code:   `class C extends Promise {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: msgPromiseNotDefined, Line: 1, Column: 17, EndLine: 1, EndColumn: 24}},
			},
			{
				Code:   `class C implements Promise<number> {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: msgPromiseNotDefined, Line: 1, Column: 20, EndLine: 1, EndColumn: 27}},
			},
		},
	)
}
