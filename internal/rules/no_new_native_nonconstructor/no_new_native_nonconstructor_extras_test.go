package no_new_native_nonconstructor

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoNewNativeNonconstructorExtras locks in branches and edge shapes that the upstream test
// suite doesn't exercise. Each case carries an inline comment pointing at the specific
// branch / Dimension 4 row / tsgo AST quirk it covers, so future refactors can't silently
// regress them without breaking a named lock-in.
func TestNoNewNativeNonconstructorExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoNewNativeNonconstructorRule,
		[]rule_tester.ValidTestCase{
			// Locks in upstream Program:exit arm 1: bare calls are references but not NewExpression callees.
			{Code: `Symbol('foo'); BigInt(1);`},

			// Locks in upstream Program:exit arm 2: new-expression arguments are references but not callees.
			{Code: `new Foo(Symbol, BigInt);`},

			// Locks in upstream parent check: indirect references are not reported unless the
			// native identifier itself is the NewExpression callee.
			{Code: `const C = Symbol; new C(); const D = BigInt; new D();`},
			{Code: `new (condition ? Symbol : Foo)(); new (BigInt || Foo)(); new (0, Symbol)();`},

			// Locks in upstream variable.defs.length guard: local value declarations shadow native globals.
			{Code: `let Symbol = class {}; const BigInt = class {}; new Symbol(); new BigInt();`},

			// ---- Dimension 4: Shadowing declaration forms ----
			{Code: `class Symbol {} enum BigInt { A } new Symbol(); new BigInt();`},
			{Code: `class Symbol {} var BigInt = class {}; new Symbol(); new BigInt();`},
			{Code: `namespace Symbol { export const x = 1; } namespace BigInt { export const x = 1; } new Symbol(); new BigInt();`},
			{Code: `declare class Symbol {} declare function BigInt(): void; new Symbol(); new BigInt();`},
			{Code: `import Symbol, { BigInt } from "mod"; new Symbol(); new BigInt();`},
			{Code: `import * as Symbol from "mod"; import { BigInt } from "mod2"; new Symbol(); new BigInt();`},

			// ---- Dimension 4: Shadowing parameter and destructuring forms ----
			{Code: `function f(Symbol = class {}) { new Symbol(); } function g(...BigInt) { new BigInt(); }`},
			{Code: `function f({ a: { Symbol } }) { new Symbol(); } function g([, BigInt]) { new BigInt(); }`},

			// ---- Dimension 4: Shadowing loop/catch/class scopes ----
			{Code: `for (let Symbol of xs) { new Symbol(); } for (const BigInt in xs) { new BigInt(); }`},
			{Code: `try {} catch ({ Symbol }) { new Symbol(); } try {} catch ([BigInt]) { new BigInt(); }`},
			{Code: `class C { method(Symbol) { new Symbol(); } field = (BigInt) => new BigInt(); }`},
			{Code: `function f() { new Symbol(); var Symbol; } function g() { new BigInt(); function BigInt() {} }`},

			// ---- Dimension 4: Access / key forms ----
			{Code: `new foo.Symbol(); new foo["BigInt"](); new globalThis.Symbol(); new globalThis["BigInt"]();`},
			{Code: `new this.Symbol(); class C extends B { m() { new super.BigInt(); } }`},

			// ---- Dimension 4: Nesting / traversal boundaries ----
			{Code: `function f(Symbol) { new Symbol(); } function g(BigInt) { new BigInt(); }`},

			// ---- Dimension 4: Graceful degradation ----
			{Code: `const { ...Symbol } = obj; new Symbol(); const [...BigInt] = arr; new BigInt();`},
			{Code: `new class Symbol {}; new (class BigInt {})();`},

			// ---- Dimension 4: Optional chain forms ----
			{Code: `new (Symbol?.for("x"))(); new (BigInt?.(1))();`},

			// ---- Config `/* global Symbol: off */` / `BigInt: off` un-declares the builtin ----
			{Code: `new Symbol();`, Globals: map[string]bool{"Symbol": false}},
			{Code: `new BigInt(1);`, Globals: map[string]bool{"BigInt": false}},

			// N/A: Declaration/container rows for object/class members do not apply; this rule only inspects NewExpression callees.
			// N/A: Autofix rows do not apply; this rule does not provide fixes.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: Receiver / expression wrappers ----
			{
				Code: `new (Symbol)(); new ((BigInt))();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewNonconstructor", Line: 1, Column: 6},
					{MessageId: "noNewNonconstructor", Line: 1, Column: 23},
				},
			},
			{
				Code: `new (Symbol as any)(); new (BigInt satisfies any)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewNonconstructor", Line: 1, Column: 6},
					{MessageId: "noNewNonconstructor", Line: 1, Column: 29},
				},
			},
			{
				Code: `new Symbol!(); new (BigInt!)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewNonconstructor", Line: 1, Column: 5},
					{MessageId: "noNewNonconstructor", Line: 1, Column: 21},
				},
			},
			{
				Code: `new ((Symbol as any)!)(); new (<any>BigInt)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewNonconstructor", Line: 1, Column: 7},
					{MessageId: "noNewNonconstructor", Line: 1, Column: 37},
				},
			},
			{
				Code: `new Symbol<string>(); new BigInt<number>();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewNonconstructor", Line: 1, Column: 5},
					{MessageId: "noNewNonconstructor", Line: 1, Column: 27},
				},
			},

			// Locks in upstream nonConstructorGlobalFunctionNames arm 1: Symbol.
			{
				Code: `new Symbol();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewNonconstructor", Message: "`Symbol` cannot be called as a constructor.", Line: 1, Column: 5, EndLine: 1, EndColumn: 11},
				},
			},

			// Locks in upstream nonConstructorGlobalFunctionNames arm 2: BigInt.
			{
				Code: `new BigInt();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewNonconstructor", Message: "`BigInt` cannot be called as a constructor.", Line: 1, Column: 5, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code: `new Symbol(); new BigInt();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewNonconstructor", Message: "`Symbol` cannot be called as a constructor.", Line: 1, Column: 5, EndLine: 1, EndColumn: 11},
					{MessageId: "noNewNonconstructor", Message: "`BigInt` cannot be called as a constructor.", Line: 1, Column: 19, EndLine: 1, EndColumn: 25},
				},
			},

			// Locks in upstream variable.defs.length guard: nested function names do not shadow outer global references.
			{
				Code: `function f() { return function Symbol() {}; } new Symbol(); function g() { return function BigInt() {}; } new BigInt();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewNonconstructor", Line: 1, Column: 51},
					{MessageId: "noNewNonconstructor", Line: 1, Column: 111},
				},
			},

			// ---- Dimension 4: Type-only declarations do not shadow value-space globals ----
			{
				Code: `type Symbol = string; interface BigInt {} new Symbol(); new BigInt();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewNonconstructor", Line: 1, Column: 47},
					{MessageId: "noNewNonconstructor", Line: 1, Column: 61},
				},
			},

			// ---- Dimension 4: Shadowing boundaries do not leak out of nested scopes ----
			{
				Code: `{ let Symbol = class {}; new Symbol(); } new Symbol(); function f(BigInt) { new BigInt(); } new BigInt();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewNonconstructor", Line: 1, Column: 46},
					{MessageId: "noNewNonconstructor", Line: 1, Column: 97},
				},
			},

			// ---- Dimension 4: Multiple nested global-name reports ----
			{
				Code: `new Symbol(); function f() { if (ok) { class C { static { new BigInt(); } } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewNonconstructor", Line: 1, Column: 5},
					{MessageId: "noNewNonconstructor", Line: 1, Column: 63},
				},
			},

			// ---- Real-user: eslint/eslint#16322 BigInt constructor confusion ----
			{
				Code: `const b = new BigInt(9007199254740991);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewNonconstructor", Line: 1, Column: 15},
				},
			},

			// ---- Real-user: eslint/eslint#16513 overlap with deprecated no-new-symbol ----
			{
				Code: `const s = new Symbol();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewNonconstructor", Line: 1, Column: 15},
				},
			},

			// ---- Dimension 4: multiline position lock-in ----
			{
				Code: "const value =\n  new Symbol();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewNonconstructor", Line: 2, Column: 7, EndLine: 2, EndColumn: 13},
				},
			},

			// Config declares Symbol/BigInt as writable globals — still the builtins.
			{
				Code:    `new Symbol();`,
				Globals: map[string]bool{"Symbol": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewNonconstructor", Line: 1, Column: 5},
				},
			},
			{
				Code:    `new BigInt(1);`,
				Globals: map[string]bool{"BigInt": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewNonconstructor", Line: 1, Column: 5},
				},
			},
		},
	)
}
