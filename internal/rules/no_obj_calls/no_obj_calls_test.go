package no_obj_calls

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoObjCallsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoObjCallsRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			{Code: `var x = Math.random();`},
			{Code: `var x = JSON.parse(foo);`},
			{Code: `Reflect.get(foo, 'x');`},
			{Code: `new Intl.Segmenter();`},
			{Code: `var x = Math;`},
			{Code: `var x = Math.PI;`},
			{Code: `var x = foo.Math();`},
			{Code: `var x = new foo.Math();`},
			{Code: `JSON.parse(foo)`},
			{Code: `new JSON.parse`},
			// globalThis property access (not calling the global itself)
			{Code: `var x = new globalThis.Math.foo;`},
			{Code: `new globalThis.Object()`},
			// Shadowed variable — should not be flagged
			{Code: `function f() { var Math = 1; Math(); }`},
			{Code: `function f(JSON: any) { JSON(); }`},
			{Code: `function f() { var globalThis = { Math: () => {} }; globalThis.Math(); }`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			{
				Code: `Math();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: `var x = JSON();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 9},
				},
			},
			{
				Code: `var x = Reflect();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 9},
				},
			},
			{
				Code: `Atomics();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: `Intl();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: `new Math();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: `new JSON();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: `new Reflect();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: `new Atomics();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: `new Intl();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			// globalThis access
			{
				Code: `var x = globalThis.Math();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 9},
				},
			},
			{
				Code: `var x = new globalThis.Math();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 9},
				},
			},
			{
				Code: `var x = globalThis.JSON();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 9},
				},
			},
			{
				Code: `var x = globalThis.Reflect();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 9},
				},
			},
			{
				Code: `globalThis.Atomics();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: `globalThis.Intl();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			// globalThis with optional chaining
			{
				Code: `var x = globalThis?.Reflect();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 9},
				},
			},
			{
				Code: `var x = (globalThis?.Reflect)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 9},
				},
			},
			// multiple errors in one expression
			{
				Code: `Math( JSON() );`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
					{MessageId: "unexpectedCall", Line: 1, Column: 7},
				},
			},
			{
				Code: `globalThis.Math( globalThis.JSON() );`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
					{MessageId: "unexpectedCall", Line: 1, Column: 18},
				},
			},
			// indirect references via variable assignment
			{
				Code: `var foo = JSON; foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 17},
				},
			},
			{
				Code: `var foo = Math; new foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 17},
				},
			},
			{
				Code: `var foo = bar ? baz : JSON; foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 29},
				},
			},
			{
				Code: `var foo = globalThis.JSON; foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 28},
				},
			},
			// indirect via logical operators
			{
				Code: `var foo = undefined || JSON; foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 30},
				},
			},
			{
				Code: `var foo = undefined ?? Reflect; foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 33},
				},
			},
			// TS type assertions as pass-through in initializer
			{
				Code: `var foo = JSON as any; foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 24},
				},
			},
			{
				Code: `var foo = JSON satisfies any; foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 31},
				},
			},
			{
				Code: `var foo = <any>JSON; foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 22},
				},
			},
			{
				Code: `var foo = JSON!; foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 18},
				},
			},
			// comma operator
			{
				Code: `var foo = (0, JSON); foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 22},
				},
			},
			// multi-hop indirect references
			{
				Code: `var a = JSON; var b = a; b();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 26},
				},
			},
			// direct call with TS assertion as callee
			{
				Code: `(JSON as any)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 1},
				},
			},
		},
	)
}
