package no_unassigned_vars

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoUnassignedVarsUpstream migrates the full valid/invalid suite from
// upstream tests/lib/rules/no-unassigned-vars.js 1:1. Position assertions cover
// line/column for every invalid case. rslint-specific lock-in cases live in the
// no_unassigned_vars_extras_test.go file.
func TestNoUnassignedVarsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnassignedVarsRule,
		[]rule_tester.ValidTestCase{
			// ---- upstream docs valid: JavaScript ----
			{Code: `let message = "hello"; console.log(message);`},
			{Code: `let user; user = getUser(); console.log(user.name);`},
			{Code: `let count; count = 1; count++;`},
			{Code: `let temp;`},
			{Code: `let error; if (somethingWentWrong) { error = "Something went wrong"; } console.log(error);`},
			{Code: `let item; for (item of items) { process(item); }`},
			{Code: `let config; function setup() { config = { debug: true }; } setup(); console.log(config);`},
			{Code: `let one = undefined; if (one === two) { }`},

			// ---- upstream valid: JavaScript ----
			{Code: `let x;`},
			{Code: `var x;`},
			{Code: `const x = undefined; log(x);`},
			{Code: `let y = undefined; log(y);`},
			{Code: `var y = undefined; log(y);`},
			{Code: `let a = x, b = y; log(a, b);`},
			{Code: `var a = x, b = y; log(a, b);`},
			{Code: `const foo = (two) => { let one; if (one !== two) one = two; }`},

			// ---- upstream valid: TypeScript ----
			{Code: `let z: number | undefined = undefined; log(z);`},
			{Code: `declare let c: string | undefined; log(c);`},
			{Code: `
const foo = (two: string): void => {
	let one: string | undefined;
	if (one !== two) {
		one = two;
	}
}`},
			{Code: `
declare module 'module' {
	import type { T } from 'module';
	let x: T;
	export = x;
}`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- upstream docs invalid: JavaScript ----
			{
				Code: `let status; if (status === "ready") { console.log("Ready!"); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("status", 1, 5, 1, 11),
				},
			},
			// ---- upstream docs invalid: TypeScript ----
			{
				Code: `let value: number | undefined; console.log(value);`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("value", 1, 5, 1, 30),
				},
			},

			// ---- upstream invalid: JavaScript ----
			{
				Code: `let x; let a = x, b; log(x, a, b);`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("x", 1, 5, 1, 6),
					unassignedError("b", 1, 19, 1, 20),
				},
			},
			{
				Code: `const foo = (two) => { let one; if (one === two) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("one", 1, 28, 1, 31),
				},
			},
			{
				Code: `let user; greet(user);`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("user", 1, 5, 1, 9),
				},
			},
			{
				Code: `function test() { let error; return error || 'Unknown error'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("error", 1, 23, 1, 28),
				},
			},
			{
				Code: `let options; const { debug } = options || {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("options", 1, 5, 1, 12),
				},
			},
			{
				Code: `let flag; while (!flag) { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("flag", 1, 5, 1, 9),
				},
			},
			{
				Code: `let config; function init() { return config?.enabled; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("config", 1, 5, 1, 11),
				},
			},

			// ---- upstream invalid: TypeScript ----
			{
				Code: `let x: number; log(x);`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("x", 1, 5, 1, 14),
				},
			},
			{
				Code: `let x: number | undefined; log(x);`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("x", 1, 5, 1, 26),
				},
			},
			{
				Code: `const foo = (two: string): void => { let one: string | undefined; if (one === two) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("one", 1, 42, 1, 65),
				},
			},
			{
				Code: `declare module 'module' {
	let x: string;
}
let y: string;
console.log(y);`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("y", 4, 5, 4, 14),
				},
			},
		},
	)
}

func unassignedError(name string, line int, column int, endLine int, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "unassigned",
		Message:   messageUnassigned(name).Description,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}
