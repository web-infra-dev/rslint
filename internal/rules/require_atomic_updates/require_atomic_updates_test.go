package require_atomic_updates

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRequireAtomicUpdatesRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RequireAtomicUpdatesRule,
		// ================================================================
		// VALID CASES
		// ================================================================
		[]rule_tester.ValidTestCase{
			// ---- Basic: no await/yield ----
			{Code: `let foo; async function x() { foo += bar; }`},
			{Code: `let foo; async function x() { foo = foo + bar; }`},
			{Code: `let foo; function* x() { foo = bar + foo; }`},

			// ---- Await BEFORE read of target → safe ----
			{Code: `let foo; async function x() { foo = await bar + foo; }`},
			{Code: `let foo; async function x() { foo = (await result)(foo); }`},
			{Code: `let foo; async function x() { foo = bar(await something, foo) }`},

			// ---- Local variable (declared inside function) → safe ----
			{Code: `async function x() { let foo; foo += await bar; }`},
			{Code: `function* x() { let foo; foo += yield bar; }`},
			{Code: `async function x() { const foo = 0; const bar = foo + await baz; }`},

			// ---- var hoisting: var declared after usage is still local ----
			{Code: `async function x() { foo += await bar; var foo = 0; }`},

			// ---- Parameter destructuring → local ----
			{Code: `async function f({foo}) { foo += await bar; }`},
			{Code: `async function f([foo]) { foo += await bar; }`},
			{Code: `async function f({a: {b: foo}}) { foo += await bar; }`},

			// ---- Object property write without member read before await ----
			{Code: `const foo = {}; async function x() { foo.bar = await baz; }`},
			{Code: `const foo = []; async function x() { foo[x] += 1; }`},

			// ---- Ternary: read and await in different branches ----
			{Code: `let foo; async function x() { foo = condition ? foo : await bar; }`},

			// ---- Closure: inner function doesn't reference outer foo ----
			{Code: `async function x() { let foo; bar(() => baz += 1); foo += await amount; }`},

			// ---- Shadowed variable in inner closure ----
			{Code: `async function x() { let foo; bar(() => { let foo; blah(foo); }); foo += await result; }`},

			// ---- Read and write before await (separate statements) ----
			{Code: `let foo; async function x() { foo = foo + 1; await bar; }`},

			// ---- Undeclared global → not tracked ----
			{Code: `async function x() { foo += await bar; }`},

			// ---- Re-read after await (separate statement) → current value ----
			{Code: `
				let count = 0
				let queue = []
				async function A(...args) {
					count += 1
					await new Promise(resolve=>resolve())
					count -= 1
					return
				}`},

			// ---- Await in unrelated statement, write to undeclared ----
			{Code: `
				async function run() {
					await a;
					b = 1;
				}`},

			// ---- Property: local variable declared with let → safe ----
			{Code: `
				async function f() {
					let foo = {}
					let bar = await get(foo.id);
					foo.prop = bar.prop;
				}`},

			// ---- Property: different variable on LHS and RHS after await ----
			{Code: `
				async function f(foo) {
					let bar = await get(foo.id);
					bar.prop = foo.prop;
				}`},

			// ---- Property: parameter reassigned after await ----
			{Code: `
				async function f(foo) {
					let bar = await get(foo.id);
					foo = bar.prop;
				}`},

			// ---- try/catch: reads in different branches ----
			{Code: `
				async function f() {
					try {
						this.foo = doSomething();
					} catch (e) {
						this.foo = null;
						await doElse();
					}
				}`},

			// ---- records assigned, then used in closure after ----
			{Code: `
				async function f() {
					let records
					records = await a.records
					g(() => { records })
				}`},

			// ---- Arrow function: local variable safe ----
			{Code: `const f = async () => { let foo; foo += await bar; };`},

			// ---- Arrow function expression body: await before read ----
			{Code: `let foo; const f = async () => foo = await bar + foo;`},

			// ---- Nested ternary: await before read (left of +) ----
			{Code: `let foo; async function x() { foo = (await bar) ? foo : baz; }`},

			// ---- Template literal: await before read ----
			{Code: "let foo; async function x() { foo = `${await bar}${foo}`; }"},

			// ---- Logical OR: no assignment to foo ----
			{Code: `let foo; async function x() { const r = foo || await bar; }`},

			// ---- for-of: loop variable is local ----
			{Code: `async function x() { for (let foo of items) { foo += await bar; } }`},

			// ---- for-in: loop variable is local ----
			{Code: `async function x() { for (let foo in obj) { foo += await bar; } }`},

			// ---- Nested async: inner function is separate scope ----
			{Code: `let foo; async function x() { async function y() { let foo; foo += await bar; } }`},

			// ---- Block-scoped let in inner block doesn't affect outer ----
			// The outer foo is declared, inner let foo is block-scoped. No race on outer
			// because the outer foo is only written, not read-before-await.
			{Code: `let foo; async function x() { if (true) { let foo = 1; } foo = 1; }`},

			// ---- var in inner block hoists to function scope → local, safe ----
			{Code: `async function x() { if (true) { var foo = 0; } foo += await bar; }`},

			// ---- TS type assertion: type name not treated as variable read ----
			{Code: `let MyType = 0; async function x() { let v = (a as MyType); await bar; MyType = 1; }`},

			// ---- NonNull assertion: safe when local ----
			{Code: `async function x() { let foo: number | null = 1; foo! += await bar; }`},

			// ---- Multiple awaits, but target only read after the last one ----
			{Code: `let foo; async function x() { await a; await b; foo = foo + 1; }`},

			// ---- Labeled statement wrapping async logic ----
			{Code: `let foo; async function x() { label: { await bar; foo = 1; } }`},

			// ---- Comma expression: await before read ----
			{Code: `let foo; async function x() { foo = (await bar, foo); }`},

			// ---- this.foo: `this` is not a tracked variable ----
			{Code: `async function x() { this.foo += await bar; }`},

			// ---- Switch: break prevents fallthrough ----
			{Code: `
				let foo;
				async function x() {
					switch (n) {
						case 1: foo; break;
						case 2: await bar; foo = 1; break;
					}
				}`},

			// ---- Switch: return prevents fallthrough ----
			{Code: `
				let foo;
				async function x() {
					switch (n) {
						case 1: foo; return;
						case 2: await bar; foo = 1;
					}
				}`},

			// ---- Switch: break in middle stops fallthrough to later cases ----
			{Code: `
				let foo;
				async function x() {
					switch (n) {
						case 1: foo; break;
						case 2: await bar;
						case 3: foo = 1;
					}
				}`},

			// ---- Property shorthand: await before shorthand ----
			{Code: `let foo; async function x() { foo = {bar: await baz, foo}; }`},

			// ---- Catch variable shadows outer: catch-local e is safe ----
			{Code: `let e; async function x() { try { await bar; } catch (e) { e += 1; } }`},

			// ---- for-await-of: loop variable is local ----
			{Code: `async function x() { for await (const item of gen) { item.foo; } }`},

			// ---- for-await-of: no read of outer var before loop ----
			{Code: `let foo; async function x() { for await (const item of gen) { foo = item; } }`},

			// ---- yield*: still a yield, but local var is safe ----
			{Code: `function* x() { let foo; foo += yield* gen; }`},

			// ---- Async IIFE: local variable safe ----
			{Code: `(async () => { let foo; foo += await bar; })();`},

			// ---- Computed property name: await in key, read in value (safe order) ----
			{Code: `let foo; async function x() { foo = { [await bar]: foo }; }`},

			// ---- Destructuring default without await: safe ----
			{Code: `async function x() { let foo; const {a = foo} = obj; foo += await bar; }`},

			// ---- Parameter default without await: safe ----
			{Code: `let foo; async function x(a = foo) { foo = 1; }`},

			// ---- Read+await inside terminating branch: continuation is safe ----
			{Code: `
				let foo;
				async function x() {
					if (cond) {
						foo;
						await bar;
						return;
					}
					foo = 1;
				}`},

			// ---- Nested if: both inner branches terminate → outer terminates ----
			{Code: `
				let foo;
				async function x() {
					if (cond) {
						foo;
						await bar;
						if (inner) return;
						else throw new Error();
					}
					foo = 1;
				}`},

			// ---- Block with early return (dead code after) still terminates ----
			{Code: `
				let foo;
				async function x() {
					if (cond) {
						foo;
						await bar;
						return;
						console.log("dead code");
					}
					foo = 1;
				}`},

			// ---- Else terminates, then does not read foo ----
			{Code: `
				let foo;
				async function x() {
					if (cond) {
						doStuff();
					} else {
						foo;
						await bar;
						return;
					}
					foo = 1;
				}`},

			// ---- allowProperties ----
			{
				Code: `
					async function a(foo) {
						if (foo.bar) {
							foo.bar = await something;
						}
					}`,
				Options: map[string]interface{}{"allowProperties": true},
			},
			{
				Code: `
					function* g(foo) {
						baz = foo.bar;
						yield something;
						foo.bar = 1;
					}`,
				Options: map[string]interface{}{"allowProperties": true},
			},

			// ---- Exponential time regression test (many branches) ----
			{Code: `
				async function foo() {
					if (1); if (2); if (3); if (4); if (5);
					if (6); if (7); if (8); if (9); if (10);
					if (11); if (12); if (13); if (14); if (15);
					if (16); if (17); if (18); if (19); if (20);
				}`},
		},
		// ================================================================
		// INVALID CASES
		// ================================================================
		[]rule_tester.InvalidTestCase{
			// ---- Compound assignment + await in RHS ----
			{
				Code: `let foo; async function x() { foo += await amount; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate", Line: 1, Column: 31},
				},
			},
			{
				Code: `if (1); let foo; async function x() { foo += await amount; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate", Line: 1, Column: 39},
				},
			},

			// ---- Inside while loop ----
			{
				Code: `let foo; async function x() { while (condition) { foo += await amount; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate", Line: 1, Column: 51},
				},
			},

			// ---- Simple assignment: read before await ----
			{
				Code: `let foo; async function x() { foo = foo + await amount; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate", Line: 1, Column: 31},
				},
			},

			// ---- Ternary in RHS ----
			{
				Code: `let foo; async function x() { foo = foo + (bar ? baz : await amount); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate", Line: 1, Column: 31},
				},
			},
			{
				Code: `let foo; async function x() { foo = foo + (bar ? await amount : baz); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate", Line: 1, Column: 31},
				},
			},
			{
				Code: `let foo; async function x() { foo = condition ? foo + await amount : somethingElse; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate", Line: 1, Column: 31},
				},
			},
			{
				Code: `let foo; async function x() { foo = (condition ? foo : await bar) + await bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate", Line: 1, Column: 31},
				},
			},

			// ---- Compound with extra term before await ----
			{
				Code: `let foo; async function x() { foo += bar + await amount; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate", Line: 1, Column: 31},
				},
			},

			// ---- Variable escapes to closure ----
			{
				Code: `async function x() { let foo; bar(() => foo); foo += await amount; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate", Line: 1, Column: 47},
				},
			},

			// ---- Generator with yield ----
			{
				Code: `let foo; function* x() { foo += yield baz }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate", Line: 1, Column: 26},
				},
			},

			// ---- foo as argument before await argument ----
			{
				Code: `let foo; async function x() { foo = bar(foo, await something) }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate", Line: 1, Column: 31},
				},
			},

			// ---- Member expression: static property ----
			{
				Code: `const foo = {}; async function x() { foo.bar += await baz }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate", Line: 1, Column: 38},
				},
			},

			// ---- Member expression: computed property ----
			{
				Code: `const foo = []; async function x() { foo[bar].baz += await result;  }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate", Line: 1, Column: 38},
				},
			},

			// ---- Member expression: private property ----
			{
				Code: `const foo = {}; class C { #bar; async wrap() { foo.#bar += await baz } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate", Line: 1, Column: 48},
				},
			},

			// ---- Async generator: yield then await ----
			{
				Code: `let foo; async function* x() { foo = (yield foo) + await bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate", Line: 1, Column: 32},
				},
			},

			// ---- Await call with target in args ----
			{
				Code: `let foo; async function x() { foo = foo + await result(foo); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate", Line: 1, Column: 31},
				},
			},

			// ---- Nested await with target in args ----
			{
				Code: `let foo; async function x() { foo = await result(foo, await somethingElse); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate", Line: 1, Column: 31},
				},
			},

			// ---- Inner async function in generator ----
			{
				Code: `function* x() { let foo; yield async function y() { foo += await bar; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate", Line: 1, Column: 53},
				},
			},

			// ---- Async generator: await then yield ----
			{
				Code: `let foo; async function* x() { foo = await foo + (yield bar); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate", Line: 1, Column: 32},
				},
			},

			// ---- Read inside await operand ----
			{
				Code: `let foo; async function x() { foo = bar + await foo; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate", Line: 1, Column: 31},
				},
			},

			// ---- Multiple errors in one expression ----
			{
				Code: `let foo = {}; async function x() { foo[bar].baz = await (foo.bar += await foo[bar].baz) }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate", Line: 1, Column: 58},
					{MessageId: "nonAtomicObjectUpdate", Line: 1, Column: 36},
				},
			},

			// ---- String initialization ----
			{
				Code: `let foo = ''; async function x() { foo += await bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate", Line: 1, Column: 36},
				},
			},

			// ---- Nested ternary with foo in deep branch ----
			{
				Code: `let foo = 0; async function x() { foo = (a ? b : foo) + await bar; if (baz); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate", Line: 1, Column: 35},
				},
			},
			{
				Code: `let foo = 0; async function x() { foo = (a ? b ? c ? d ? foo : e : f : g : h) + await bar; if (baz); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate", Line: 1, Column: 35},
				},
			},

			// ---- Cross-statement: property read → await → property write ----
			{
				Code: `
					async function f(foo) {
						let buz = await get(foo.id);
						foo.bar = buz.bar;
					}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate"},
				},
			},

			// ---- Property: default (no option) ----
			{
				Code: `
					async function a(foo) {
						if (foo.bar) {
							foo.bar = await something;
						}
					}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate"},
				},
			},
			{
				Code: `
					function* g(foo) {
						baz = foo.bar;
						yield something;
						foo.bar = 1;
					}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate"},
				},
			},

			// ---- Property: allowProperties=false ----
			{
				Code: `
					async function a(foo) {
						if (foo.bar) {
							foo.bar = await something;
						}
					}`,
				Options: map[string]interface{}{"allowProperties": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate"},
				},
			},
			{
				Code: `
					function* g(foo) {
						baz = foo.bar;
						yield something;
						foo.bar = 1;
					}`,
				Options: map[string]interface{}{"allowProperties": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate"},
				},
			},

			// ---- allowProperties=true still reports variable issues ----
			{
				Code: `
					let foo;
					async function a() {
						if (foo) {
							foo = await something;
						}
					}`,
				Options: map[string]interface{}{"allowProperties": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},
			{
				Code: `
					let foo;
					function* g() {
						baz = foo;
						yield something;
						foo = 1;
					}`,
				Options: map[string]interface{}{"allowProperties": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ================================================================
			// Additional edge cases
			// ================================================================

			// ---- Arrow function (block body) ----
			{
				Code: `let foo; const f = async () => { foo += await bar; };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Arrow function (expression body) ----
			{
				Code: `let foo; const f = async () => foo += await bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Async method in class ----
			{
				Code: `let foo; class C { async m() { foo += await bar; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Logical assignment operators ----
			{
				Code: `let foo; async function x() { foo ||= await bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},
			{
				Code: `let foo; async function x() { foo &&= await bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},
			{
				Code: `let foo; async function x() { foo ??= await bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Template literal: read before await ----
			{
				Code: "let foo; async function x() { foo = `${foo}${await bar}`; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Deeply nested member access ----
			{
				Code: `let foo; async function x() { foo.a.b.c += await bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate"},
				},
			},

			// ---- Mixed bracket and dot access ----
			{
				Code: `let foo; async function x() { foo[0].bar += await baz; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate"},
				},
			},

			// ---- for loop body ----
			{
				Code: `let foo; async function x() { for (let i = 0; i < 10; i++) { foo += await bar; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- do-while loop body ----
			{
				Code: `let foo; async function x() { do { foo += await bar; } while (cond); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- for-in loop body ----
			{
				Code: `let foo; async function x() { for (const key in obj) { foo += await bar; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- for-of loop body ----
			{
				Code: `let foo; async function x() { for (const item of items) { foo += await bar; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- switch/case ----
			{
				Code: `let foo; async function x() { switch(x) { case 1: foo += await bar; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- try/catch: read in try, await in try, write in try ----
			{
				Code: `
					let foo;
					async function x() {
						try {
							foo += await bar;
						} catch (e) {}
					}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Cross-statement with if: read in condition, await+write in body ----
			{
				Code: `
					let foo;
					async function x() {
						if (foo) {
							foo = await bar;
						}
					}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Spread before await in array literal ----
			{
				Code: `let foo; async function x() { foo = [...foo, await bar]; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Await wrapped in parentheses ----
			{
				Code: `let foo; async function x() { foo = foo + (await bar); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Exponentiation assignment ----
			{
				Code: `let foo; async function x() { foo **= await bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Bitwise assignment ----
			{
				Code: `let foo; async function x() { foo &= await bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},
			{
				Code: `let foo; async function x() { foo |= await bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},
			{
				Code: `let foo; async function x() { foo ^= await bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Shift assignment ----
			{
				Code: `let foo; async function x() { foo <<= await bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},
			{
				Code: `let foo; async function x() { foo >>= await bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},
			{
				Code: `let foo; async function x() { foo >>>= await bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- TypeScript: type assertion wrapping LHS ----
			{
				Code: `let foo; async function x() { (foo as any) += await bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Nested property: compound on deep member ----
			{
				Code: `const obj = {}; async function x() { obj.a.b.c.d += await bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate"},
				},
			},

			// ---- Variable in finally block: read before try/await, write in finally ----
			{
				Code: `
					let foo;
					async function x() {
						if (foo > 0) {
							try {
								await doSomething();
							} finally {
								foo = 0;
							}
						}
					}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Multiple variables: only the stale one reported ----
			{
				Code: `let foo; let bar; async function x() { foo = foo + await bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Await in condition of if, then compound assignment ----
			{
				Code: `
					let foo;
					async function x() {
						foo;
						if (await cond) {
							foo = 1;
						}
					}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Block-scoped let NOT treated as function-local ----
			// Inner `let foo` in if-block is block-scoped; outer `let foo` is non-local.
			{
				Code: `
					let foo;
					async function x() {
						if (true) { let foo = 1; }
						foo += await bar;
					}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Comma expression: read before await ----
			{
				Code: `let foo; async function x() { foo = (foo, await bar); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Tagged template: read before await ----
			{
				Code: "let foo; async function x() { foo = tag`${foo}${await bar}`; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Object literal: read before await in values ----
			{
				Code: `let foo; async function x() { foo = {a: foo, b: await bar}; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- new expression: read before await in args ----
			{
				Code: `let foo; async function x() { foo = new Cls(foo, await bar); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Nested async arrow inside async function ----
			{
				Code: `let foo; const outer = async () => { const inner = async () => { foo += await bar; }; };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Async method in object literal ----
			{
				Code: `
					let foo;
					const obj = {
						async method() { foo += await bar; }
					};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Multiple awaits: read → first await → outdated stays through second await ----
			{
				Code: `
					let foo;
					async function x() {
						foo;
						await a;
						await b;
						foo = 1;
					}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- NonNull assertion on LHS ----
			{
				Code: `let foo: number | null; async function x() { foo! += await bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Deeply nested control flow ----
			{
				Code: `
					let foo;
					async function x() {
						for (let i = 0; i < 10; i++) {
							if (cond) {
								try {
									foo += await bar;
								} catch (e) {}
							}
						}
					}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- void expression wrapping assignment ----
			{
				Code: `let foo; async function x() { void (foo += await bar); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- for-await-of: read before loop, write inside ----
			{
				Code: `
					let foo;
					async function x() {
						foo;
						for await (const item of gen) {
							foo = item;
						}
					}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Property shorthand: read before await ----
			{
				Code: `let foo; async function x() { foo = {foo, bar: await baz}; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Computed property name: read in value before await in key ----
			{
				Code: `let foo; async function x() { foo = { a: foo, [await bar]: baz }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Logical OR: read-await pattern in assignment ----
			{
				Code: `let foo; async function x() { foo = foo || await bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Logical AND: read-await pattern in assignment ----
			{
				Code: `let foo; async function x() { foo = foo && await bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Nullish coalescing: read-await pattern in assignment ----
			{
				Code: `let foo; async function x() { foo = foo ?? await bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Await inside parens of a call ----
			{
				Code: `let foo; async function x() { foo = foo.method(await bar); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Cross-statement: read in while condition, await+write in body ----
			{
				Code: `
					let foo;
					async function x() {
						while (foo > 0) {
							foo = await decrement();
						}
					}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- yield* with outer var ----
			{
				Code: `let foo; function* x() { foo += yield* gen; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Subtraction assignment ----
			{
				Code: `let foo; async function x() { foo -= await bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Multiplication assignment ----
			{
				Code: `let foo; async function x() { foo *= await bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Division assignment ----
			{
				Code: `let foo; async function x() { foo /= await bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Remainder assignment ----
			{
				Code: `let foo; async function x() { foo %= await bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Parameter default with await: read before pause ----
			{
				Code: `let foo; async function x(a = foo + await bar) { foo = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Destructuring default with await ----
			{
				Code: `
					let foo;
					async function x() {
						foo;
						const {a = await bar} = obj;
						foo = 1;
					}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Nested destructuring default with await ----
			{
				Code: `
					let foo;
					async function x() {
						foo;
						const {a: {b = await bar}} = obj;
						foo = 1;
					}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- If-branch reads+awaits but does NOT exit → report ----
			{
				Code: `
					let foo;
					async function x() {
						if (cond) {
							foo;
							await bar;
						}
						foo = 1;
					}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Else-branch terminates but then-branch doesn't → then state used ----
			{
				Code: `
					let foo;
					async function x() {
						if (cond) {
							foo;
							await bar;
						} else {
							return;
						}
						foo = 1;
					}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Array destructuring default with await ----
			{
				Code: `
					let foo;
					async function x() {
						foo;
						const [a = await bar] = arr;
						foo = 1;
					}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Global variable from ambient declaration (ESLint #15076) ----
			// Mirrors ESLint's process.exitCode test: TypeChecker resolves the
			// ambient `process` declaration, so it's tracked as a known variable.
			{
				Code: `
					declare var process: { stdin: any; exitCode: number };
					declare var opts: { spec: any };
					declare function run(o: any): Promise<{ exit_code: number }>;
					async () => {
						opts.spec = process.stdin;
						try {
							const { exit_code } = await run(opts);
							process.exitCode = exit_code;
						} catch (e) {
							process.exitCode = 1;
						}
					};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate"},
					{MessageId: "nonAtomicObjectUpdate"},
				},
			},

			// ---- Fix #1: if without else — re-read in then shouldn't clear outdated for the non-taken path ----
			{
				Code: `
					let foo;
					async function x() {
						foo;
						await bar;
						if (cond) { foo; }
						foo = 1;
					}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Fix #6: block with return in middle (dead code after) still terminates ----
			// Then-branch terminates despite dead code after return → no report for code after if.
			// Instead we test: block without early terminator in then-branch DOES report.
			{
				Code: `
					let foo;
					async function x() {
						if (cond) {
							foo;
							await bar;
							doStuff();
						}
						foo = 1;
					}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Fix #7: switch fallthrough — read in case 1 carries to case 2 ----
			{
				Code: `
					let foo;
					async function x() {
						switch (n) {
							case 1:
								foo;
							case 2:
								await bar;
								foo = 1;
						}
					}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Switch: fallthrough across 3 cases ----
			{
				Code: `
					let foo;
					async function x() {
						switch (n) {
							case 1: foo;
							case 2: await bar;
							case 3: foo = 1;
						}
					}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Switch: default case with fallthrough ----
			{
				Code: `
					let foo;
					async function x() {
						switch (n) {
							default: foo;
							case 1: await bar; foo = 1;
						}
					}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},
		},
	)
}
