package no_unused_vars

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnusedVarsAssignments(t *testing.T) {
	validTestCases := []rule_tester.ValidTestCase{
		// --- write-only destructuring assignment: used after reassignment ---
		{Code: `let a: any, b: any; [a, b] = [1, 2]; console.log(a, b);`},
		{Code: `let a: any; ({ a } = { a: 1 }); console.log(a);`},
		// spread in array destructuring assignment: used
		{Code: `let rest: any; [, ...rest] = [1, 2, 3]; console.log(rest);`},
		// spread in object destructuring assignment: used
		{Code: `let rest: any; ({ ...rest } = { a: 1, b: 2 } as any); console.log(rest);`},
		// nested: object in array assignment, used
		{Code: `let a: any; [{ a }] = [{ a: 1 }] as any; console.log(a);`},
		// nested: array in object assignment, used
		{Code: `let b: any; ({ x: [b] } = { x: [1] } as any); console.log(b);`},
		// renamed property assignment: used
		{Code: `let b: any; ({ a: b } = { a: 1 } as any); console.log(b);`},
		// for-of: variable used in body
		{Code: `let x: any; for (x of [1, 2]) { console.log(x); }`},
		// for-of destructuring: used
		{Code: `let a: any, b: any; for ([a, b] of [[1, 2]] as any) { console.log(a, b); }`},
		// for-in: variable used in body
		{Code: `let k: any; for (k in { a: 1 }) { console.log(k); }`},
		// for-of object destructuring: used
		{Code: `let a: any; for ({ a } of [{ a: 1 }] as any) { console.log(a); }`},
		// deeply nested array assignment: used
		{Code: `let a: any; [[[[a]]]] = [[[[1]]]] as any; console.log(a);`},
		// mixed deeply nested: used
		{Code: `let a: any; [{ x: [{ a }] }] = [{ x: [{ a: 1 }] }] as any; console.log(a);`},
		// default value in destructuring assignment: used
		{Code: `let a: any; [a = 5] = [] as any; console.log(a);`},
		// computed property name in destructuring assignment: key IS a read, both used
		{Code: `const key = "x"; let val: any; ({ [key]: val } = { x: 1 } as any); console.log(val);`},
		// property access assignment: object IS a read (not a write target)
		{Code: `const obj = { b: 0 }; obj.b = 1; console.log(obj);`},
		// element access assignment: array IS a read
		{Code: `const arr = [0]; arr[0] = 1; console.log(arr);`},
		// chain assignment: both a and b written, c read
		{Code: `let a: any, b: any; const c = 1; a = b = c; console.log(a, b);`},
		// skip holes in array destructuring: used
		{Code: `let a: any; [,, a] = [1, 2, 3] as any; console.log(a);`},
		// renamed property with default value in assignment: used
		{Code: `let b: any; ({ a: b = 5 } = {} as any); console.log(b);`},
		// parenthesized assignment target: used
		{Code: `let a: any; [(a)] = [1] as any; console.log(a);`},
		// nested parenthesized: used
		{Code: `let b: any; ({a: (b)} = {a: 1} as any); console.log(b);`},

		// --- self-assignment: used (result consumed) ---
		{Code: `let a = 0; a = a + 1; console.log(a);`},
		{Code: `let a = 0; a++; console.log(a);`},
		{Code: `let a = 0; const b = (a = a + 1); console.log(b);`},

		// --- for-in with computed property access: loop var used ---
		{Code: `
const box: any = { a: 2 };
for (const prop in box) {
  box[prop] = parseInt(box[prop]);
}
`},
		// for-in with var
		{Code: `
const box: any = { a: 2 };
for (var prop in box) {
  box[prop] = parseInt(box[prop]);
}
`},
		// for-of with const
		{Code: `for (const item of [1, 2, 3]) { console.log(item); }`},
		// for-of with let
		{Code: `for (let item of [1, 2, 3]) { console.log(item); }`},
		// for-in with const
		{Code: `for (const key in { a: 1 }) { console.log(key); }`},

		// --- sequence expression: self-modification result consumed ---
		{Code: `let x = 0; x++, console.log(x);`},
		// result of sequence consumed by assignment
		{Code: `let x = 0; const y = (x++, x); console.log(y);`},

		// --- logical assignment: used ---
		{Code: `let a: any = null; a ??= 1; console.log(a);`},
		{Code: `let a: any = ""; a ||= "default"; console.log(a);`},
		{Code: `let a: any = true; a &&= false; console.log(a);`},
	}

	invalidTestCases := []rule_tester.InvalidTestCase{
		// --- write-only destructuring assignment: declared via let, assigned via array pattern ---
		// destructuredArrayIgnorePattern applies via assignment-side array destructuring write
		{
			Code: `
let _a: any, b: any;
[_a, b] = [1, 2];
`,
			Options: map[string]interface{}{"destructuredArrayIgnorePattern": "^_"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 6},
			},
		},
		// simple object destructuring assignment (write-only)
		{
			Code: `
let a: any, b: any;
({ a, b } = { a: 1, b: 2 });
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 2, Column: 5},
				{MessageId: "unusedVar", Line: 2, Column: 13},
			},
		},
		// nested destructuring assignment (write-only)
		{
			Code: `
let a: any;
[[a]] = [[1]];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 3},
			},
		},
		// spread in destructuring assignment (write-only)
		{
			Code: `
let rest: any;
[, ...rest] = [1, 2, 3];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 7},
			},
		},
		// mixed: some destructuring-assigned vars used, some not
		{
			Code: `
let a: any, b: any;
[a, b] = [1, 2];
console.log(a);
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 5},
			},
		},
		// object spread assignment (write-only)
		{
			Code: `
let rest: any;
({ ...rest } = { a: 1 } as any);
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 7},
			},
		},
		// renamed property assignment (write-only)
		{
			Code: `
let b: any;
({ a: b } = { a: 1 } as any);
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 7},
			},
		},
		// nested: object in array assignment (write-only)
		{
			Code: `
let a: any;
[{ a }] = [{ a: 1 }] as any;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 2, Column: 5},
			},
		},
		// nested: array in object assignment (write-only)
		{
			Code: `
let b: any;
({ x: [b] } = { x: [1] } as any);
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 8},
			},
		},
		// for-of: write-only (never used in body)
		{
			Code: `
let x: any;
for (x of [1, 2]) {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 6},
			},
		},
		// for-of destructuring: write-only
		{
			Code: `
let a: any, b: any;
for ([a, b] of [[1, 2]] as any) {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 7},
				{MessageId: "unusedVar", Line: 3, Column: 10},
			},
		},
		// for-in: write-only
		{
			Code: `
let k: any;
for (k in { a: 1 }) {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 6},
			},
		},
		// for-of object destructuring: write-only
		{
			Code: `
let a: any;
for ({ a } of [{ a: 1 }] as any) {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 2, Column: 5},
			},
		},
		// deeply nested array assignment: write-only
		{
			Code: `
let a: any;
[[[[a]]]] = [[[[1]]]] as any;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 5},
			},
		},
		// mixed deeply nested: write-only
		{
			Code: `
let a: any;
[{ x: [{ a }] }] = [{ x: [{ a: 1 }] }] as any;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 2, Column: 5},
			},
		},
		// default value in destructuring assignment: write-only
		{
			Code: `
let a: any;
[a = 5] = [] as any;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 2},
			},
		},
		// parenthesized assignment target: write-only
		{
			Code: `
let a: any;
[(a)] = [1] as any;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 3},
			},
		},
		// parenthesized in object destructuring assignment: write-only
		{
			Code: `
let b: any;
({a: (b)} = {a: 1} as any);
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 7},
			},
		},
		// renamed property with default value: write-only
		{
			Code: `
let b: any;
({ a: b = 5 } = {} as any);
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 7},
			},
		},
		// skip holes: write-only
		{
			Code: `
let a: any;
[,, a] = [1, 2, 3] as any;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 5},
			},
		},
		// chain assignment: write-only (b never read)
		{
			Code: `
let a: any, b: any;
a = b = 1;
console.log(a);
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 5},
			},
		},

		// --- self-assignment: variable only used to modify itself ---
		{
			Code:   `var a = 0; a = a + 1;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 12}},
		},
		{
			Code:   `var a = 0; a = a + a;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 12}},
		},
		{
			Code:   `var a = 0; a += a + 1;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 12}},
		},
		{
			Code:   `var a = 0; a++;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 12}},
		},
		{
			Code:   `function foo(a: number) { a = a + 1; } foo(1);`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 27}},
		},
		{
			Code:   `function foo(a: number) { a++; } foo(1);`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 27}},
		},
		{
			Code:   `var a = 3; a = a * 5 + 6;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 12}},
		},
		// prefix decrement
		{
			Code:   `var a = 0; --a;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 14}},
		},
		// reassignment-only x = foo(x)
		{
			Code: `
let x = null;
x = foo(x);
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 3, Column: 1}},
		},

		// --- sequence expression: self-modification (comma operator) ---
		{
			Code: `
let x = 0;
x++, (x = 0);
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 3, Column: 7}},
		},
		{
			Code: `
let b = 0;
0, b++;
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 3, Column: 4}},
		},
		{
			Code: `
let c = 0;
(c += 1), 0;
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 3, Column: 2}},
		},
		// nested sequence
		{
			Code: `
let x = 0;
0, (1, x++);
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 3, Column: 8}},
		},
		// self-assignment in sequence, followed by direct reassignment
		{
			Code: `
let z = 0;
(z = z + 1), (z = 2);
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 3, Column: 15}},
		},
		// compound assignment in sequence
		{
			Code: `
let x = 0;
0, (x += 1);
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 3, Column: 5}},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnusedVarsRule, validTestCases, invalidTestCases)
}
