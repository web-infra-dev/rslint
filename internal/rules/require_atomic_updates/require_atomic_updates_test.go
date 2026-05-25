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

			// ---- Property access name in closure is NOT a variable reference ----
			// `(item) => item.foo` only references `item`, not outer `foo`.
			// So outer `foo` is still local-without-escape.
			{Code: `
				async function x(list, api) {
					list.map((item) => item.foo);
					let foo = 1;
					if (list.length && foo !== 2) {
						await api.update({ foo });
						foo = 2;
					}
				}`},

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

			// ---- ESLint quirk: for-of / for-in / for-await-of bodies don't inherit
			//      pre-loop read state, so an outer read isn't outdated by body awaits. ----
			{Code: `let foo; function cb() { return foo; }
				async function x(gen) { foo; for await (const item of gen) { foo = item; } }`},
			{Code: `let foo; function cb() { return foo; }
				async function x(arr, hook) { foo; for (const c of arr) { await hook(); foo = 1; } }`},
			{Code: `let foo; function cb() { return foo; }
				async function x(arr, hook) { foo; for (const k in arr) { await hook(); foo = 1; } }`},

			// ---- for-of assignment form: bare-identifier target is pure-write ----
			// ESLint's getWriteExpr breaks at the for-of boundary only WHEN traversing
			// through a member access. A bare identifier is treated as a pure write,
			// so `foo` doesn't get marked as read → body await can't outdate it.
			{Code: `let foo; function cb() { return foo; }
				async function f(gen, x) { for (foo of gen) { await x; foo = 1; } }`},

			// ---- for-of assignment form: destructure target without pre-read ----
			{Code: `let foo; function cb() { return foo; }
				async function f(gen) { for ({foo} of gen) {} }`},

			// ---- for-of with plain const binding: no reads in initializer ----
			{Code: `let foo; function cb() { return foo; }
				async function f(gen, x) { for (const it of gen) { await x; foo = it; } }`},

			// ---- for-loop's update runs in a fresh segment: a simple `foo = await x`
			//      in the update doesn't see pre-loop reads, so no race is reported. ----
			{Code: `let foo; function cb() { return foo; }
				async function f(x) { foo; for(;; foo = await x) {} }`},

			// ---- switch where every case exits via return/throw: post-switch
			//      doesn't inherit any case's outdated state. ----
			{Code: `let foo; function cb() { return foo; }
				async function f(x, n) { foo; switch (n) { case 1: await x; return; case 2: return; } foo = 1; }`},

			// ---- Catch binding shadows outer: inner write doesn't see outer state ----
			// Also covers: optional catch (no binding) leaves the outer variable
			// visible, see invalid counterpart below.
			{Code: `let foo; function cb() { return foo; }
				async function f(x) { try { foo; await x; } catch (foo) { foo = 1; } }`},

			// ---- Catch binding destructured: same shadow rule applies ----
			{Code: `let foo; function cb() { return foo; }
				async function f(x) { try { foo; await x; } catch ({ foo }) { foo = 1; } }`},

			// ---- let/const in a bare block shadow outer for the block duration ----
			{Code: `let foo; function cb() { return foo; }
				async function f(x) { foo; await x; { let foo = 1; foo = 2; } }`},

			// ---- Block-scoped function / class declarations shadow outer ----
			{Code: `let foo; function cb() { return foo; }
				async function f(x) { foo; await x; { function foo() {} foo(); } }`},
			{Code: `let foo; function cb() { return foo; }
				async function f(x) { foo; await x; { class foo {} new foo(); } }`},

			// ---- for-loop let shadow covers init / cond / body / update ----
			{Code: `let foo; function cb() { return foo; }
				async function f(x) { foo; await x; for (let foo = 0; foo < 1; foo++) {} }`},

			// ---- Nested async function is a separate scope ----
			{Code: `let foo; function cb() { return foo; }
				async function f() { async function inner() { foo; } }`},

			// ---- namespace with a function export + function RHS: no report
			//      (function-like RHS skip aligns with ESLint's :expression:exit quirk). ----
			{Code: `namespace NS { export function foo() {} }
				function cb() { return NS.foo; }
				async function f(x: any) { NS.foo; await x; NS.foo = () => {}; }`},

			// ---- RHS is a function/arrow expression: ESLint silently skips the
			//      outdated check (the :expression:exit fires inside the inner
			//      function's CodePath whose referenceMap is null). Match it. ----
			{Code: `let foo: any; function cb() { return foo; }
				async function f(x: any) { foo; await x; foo = () => {}; }`},
			{Code: `let foo: any; function cb() { return foo; }
				async function f(x: any) { foo; await x; foo = function () {}; }`},
			{Code: `let foo: any; function cb() { return foo; }
				async function f(x: any) { foo; await x; foo = async () => {}; }`},
			// Class / object RHS still reports (they don't start a function CodePath).
			{Code: `let foo = {}; function cb() { return foo; }
				async function f(x: any) { await x; foo.bar = 1; }`},

			// ---- try that ALWAYS returns: catch entry = pre-try state ----
			// Outdates produced late in the try don't leak into catch's state.
			{Code: `let foo; function cb() { return foo; }
				async function f(x) { foo; try { await x; return; } catch {} foo = 1; }`},
			{Code: `let foo; function cb() { return foo; }
				async function f(x) { try { foo; await x; return; } catch { foo = 1; } }`},

			// ---- Both try and catch always exit: post-try unreachable ----
			{Code: `let foo; function cb() { return foo; }
				async function f(x) { try { await x; return; } catch { throw 0; } foo = 1; }`},

			// ---- Member write target where base is NOT read-before-await ----
			// `await x; foo.bar = 1` has no race because `foo` was never read before await.
			{Code: `let foo = {}; function cb() { return foo; }
				async function f(x) { await x; foo.bar = 1; foo = 2; }`},

			// ---- `foo[k] = 1` — the computed key `k` read marks k, clearing its outdated ----
			{Code: `let foo, k; function cb() { return foo + k; }
				async function f(x) { k; await x; foo[k] = 1; }`},

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

			// ---- Loop body awaits don't propagate to post-loop code ----
			// ESLint per-segment initialization: post-loop sees entry state, not body state.
			{Code: `let foo; function cb() { return foo; }
				async function f(x) { foo; for (let i = 0; i < 10; i++) { await x; } foo = 1; }`},
			{Code: `let foo; function cb() { return foo; }
				async function f(x) { foo; while (x) { await x; } foo = 1; }`},
			{Code: `let foo; function cb() { return foo; }
				async function f(x) { foo; do { await x; } while (x); foo = 1; }`},
			{Code: `let foo; function cb() { return foo; }
				async function f(gen) { foo; for (const it of gen) { await it; } foo = 1; }`},
			{Code: `let foo; function cb() { return foo; }
				async function f(gen) { foo; for await (const it of gen) {} foo = 1; }`},

			// ---- Update expressions (foo++, --foo) are not tracked as assignments ----
			{Code: `let foo; function cb() { return foo; } async function f(x) { foo; await x; foo++; }`},
			{Code: `let foo; function cb() { return foo; } async function f(x) { await x; --foo; }`},

			// ---- Await in LHS base doesn't register the member target ----
			// `(await ptr).foo = x` — the chain breaks at AwaitExpression, foo isn't tracked.
			{Code: `let foo; function cb() { return foo; } async function f(ptr, x) { foo; (await ptr).foo = x; }`},

			// ---- Labeled break before await: no race ----
			{Code: `let foo; function cb() { return foo; }
				async function f(x) { foo; outer: for (;;) { if (x) break outer; await x; } foo = 1; }`},

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
			// The type wrapper keeps foo as the write target but ESLint reports
			// as nonAtomicObjectUpdate with the wrapper text as the value
			// (since AssignmentExpression.left !== identifier).
			{
				Code: `let foo; async function x() { (foo as any) += await bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate"},
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
			// Same as type-assertion: ESLint reports the wrapper form as
			// nonAtomicObjectUpdate since the LHS is not the bare identifier.
			{
				Code: `let foo: number | null; async function x() { foo! += await bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate"},
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

			// ---- Element access: computed index on LHS of simple `=` is read ----
			// `foo[x] = await bar` reads `x`; `x = 1` after must report.
			{
				Code: `let x; let foo = {}; function cb() { return x; }
					async function f(bar) { foo[x] = await bar; x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Destructuring assignment: object pattern ----
			{
				Code: `let foo; function cb() { return foo; } async function f(src, x) { foo; await x; ({foo} = src); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate"},
				},
			},
			// ---- Destructuring assignment: array pattern ----
			{
				Code: `let foo; function cb() { return foo; } async function f(src, x) { foo; await x; [foo] = src; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate"},
				},
			},
			// ---- Destructuring: aliased target `{a: foo}` ----
			{
				Code: `let foo; function cb() { return foo; } async function f(src, x) { foo; await x; ({a: foo} = src); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate"},
				},
			},
			// ---- Destructuring: multiple variables ----
			{
				Code: `let foo, bar; function cb() { return foo + bar; } async function f(src, x) { foo; bar; await x; ({foo, bar} = src); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate"},
					{MessageId: "nonAtomicObjectUpdate"},
				},
			},
			// ---- Destructuring: default with await outdates target ----
			{
				Code: `let foo; function cb() { return foo; } async function f(src, x) { foo; ({foo = await x} = src); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate"},
				},
			},
			// ---- Outer variable still outdated after shadowed inner block ----
			// Inner `let foo`'s writes don't affect outer foo; the outer write
			// after the block still sees the outdated state from the earlier await.
			{
				Code: `let foo; function cb() { return foo; }
					async function f(x) { foo; await x; { let foo = 1; foo = 2; } foo = 3; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},
			// Same with if-block containing the shadowing let.
			{
				Code: `let foo; function cb() { return foo; }
					async function f(x, cond) { foo; await x; if (cond) { let foo = 1; foo = 2; } foo = 3; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},
			// Block-scoped function declaration shadow: outer write after block reports.
			{
				Code: `let foo; function cb() { return foo; }
					async function f(x) { foo; await x; { function foo() {} foo(); } foo = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},
			// Block-scoped class declaration shadow.
			{
				Code: `let foo; function cb() { return foo; }
					async function f(x) { foo; await x; { class foo {} new foo(); } foo = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},
			// for-loop let shadow: outer write after for reports.
			{
				Code: `let foo; function cb() { return foo; }
					async function f(x) { foo; await x; for (let foo = 0; foo < 1; foo++) {} foo = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},
			// Catch binding shadow: outer write after try/catch reports.
			{
				Code: `let foo; function cb() { return foo; }
					async function f(x) { try { foo; await x; } catch (foo) { foo = 1; } foo = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},
			// TypeScript namespace / module at outer scope creates a value binding
			// that is tracked as declared.
			{
				Code: `namespace NS { export var foo = 1; }
					function cb() { return NS.foo; }
					async function f(x: any) { NS.foo; await x; NS.foo = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate"},
				},
			},
			// Namespace with a function export, value RHS still reports
			{
				Code: `namespace NS { export function foo() {} }
					function cb() { return NS.foo; }
					async function f(x: any) { NS.foo; await x; NS.foo = 1 as any; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate"},
				},
			},
			// Nested namespace: outer member read + await + inner member write
			{
				Code: `namespace A { export namespace B { export var foo = 1; } }
					function cb() { return A.B.foo; }
					async function f(x: any) { A.B.foo; await x; A.B.foo = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate"},
				},
			},
			// declare namespace at outer scope is tracked too
			{
				Code: `declare namespace NS { var foo: number; }
					async function f(x: any) { NS.foo; await x; NS.foo = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate"},
				},
			},

			// TS type wrappers on the LHS are transparent for the write-target chain:
			// `(foo as any) = 1` treats foo as write target (no extra markRead), and
			// ESLint reports the write itself as nonAtomicObjectUpdate (wrapper text).
			{
				Code: `let foo: any; function cb() { return foo; }
					async function f(x: any) { foo; await x; (foo as any) = 1; foo = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate"},
					{MessageId: "nonAtomicUpdate"},
				},
			},
			// Non-null assertion directly on simple LHS
			{
				Code: `let foo: any; function cb() { return foo; }
					async function f(x: any) { foo; await x; foo! = 1; foo = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate"},
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// Optional catch (no binding): `foo` inside catch IS the outer variable.
			{
				Code: `let foo; function cb() { return foo; }
					async function f(x) { try { foo; await x; } catch { foo = 1; } foo = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- for-of declaration with await default in destructure ----
			// The default `foo + await gen.y` reads foo then awaits, outdating foo
			// within the loop's body state; the subsequent write reports.
			{
				Code: `let foo = 1; function cb() { return foo; }
					async function f(gen) { for (const { x = foo + await gen.y } of [{}]) { foo = x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},
			// ---- for-of assignment-form destructure with await default ----
			{
				Code: `let foo = 1; function cb() { return foo; }
					async function f(gen) { for ({ x = foo + await gen.y } of [{}]) { foo = x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},
			// ---- Compound `foo += await x` in a for-loop update still reports ----
			// The LHS is a read+write, so `foo` is marked fresh inside the update
			// segment; the await then outdates it and the implicit write reports.
			{
				Code: `let foo; function cb() { return foo; }
					async function f(x) { for(;; foo += await x) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Destructuring default with re-read doesn't clear outer outdated ----
			// Defaults are conditional branches: an outdate from one default
			// propagates by union, and a re-read inside another default does NOT
			// clear the outer's outdated flag for that name.
			{
				Code: `let foo, bar; function cb() { return foo + bar; }
					async function f(src, x) { foo; bar; const {a = await x, b = foo + bar} = src; foo = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Labeled break from nested loop carries state to the labeled
			//      loop's post-loop point. After `break outer`, post-outer-loop
			//      sees the state at the break (foo outdated). ----
			{
				Code: `let foo; function cb() { return foo; }
					async function f(x) { outer: for (;;) { for (;;) { foo; await x; break outer; } } foo = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- for-of assignment form: member target base IS read ----
			// `foo.bar of gen` reads foo (object base). await then outdates foo;
			// subsequent `foo = 1` reports.
			{
				Code: `let foo = {}; function cb() { return foo; }
					async function f(gen, x) { for (foo.bar of gen) { await x; foo = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Destructuring: rest element ----
			{
				Code: `let foo; function cb() { return foo; } async function f(src, x) { foo; await x; ({...foo} = src); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate"},
				},
			},

			// ---- Chained assignment `a = b = 1` reports each target ----
			{
				Code: `let a, b; function cb() { return a + b; } async function f(x) { a; b; await x; a = b = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
					{MessageId: "nonAtomicUpdate"},
				},
			},

			// ---- Non-null assertion on LHS breaks write-target chain ----
			// ESLint only walks MemberExpression.object; type assertions break the chain,
			// so `foo` in `foo!.bar = X` is a read (not a write-target base). Thus `foo = 1`
			// after an await is reported. The `foo!.bar =` line itself is NOT reported
			// (the chain breaks before reaching the assignment for registration).
			{
				Code: `let foo: any; function cb() { return foo; }
					async function f(bar) { foo!.bar = await bar; foo = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},
		},
	)
}
