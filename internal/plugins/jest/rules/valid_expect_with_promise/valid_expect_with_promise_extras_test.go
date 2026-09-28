package valid_expect_with_promise

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestValidExpectWithPromiseExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers,
// so future refactors can't silently regress them without breaking a named
// lock-in.
func TestValidExpectWithPromiseExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ValidExpectWithPromiseRule,
		[]rule_tester.ValidTestCase{
			// Locks in upstream parseJestFnCall() arm 1: calls that are not expect chains are ignored
			{Code: `foo(Promise.resolve()).toBe(1)`},
			{Code: `import { expect } from 'other'; expect(Promise.resolve()).toBe(1)`},
			{Code: `const expect = (x: unknown) => ({ toBe(_: unknown) {} }); expect(Promise.resolve()).toBe(1)`},
			{Code: `import * as g from '@jest/globals'; g.expect(Promise.resolve()).toBe(1)`},
			// Locks in upstream parseJestFnCall() arm 2: chains without a called matcher are ignored
			{Code: `expect(Promise.resolve())`},
			{Code: `expect(Promise.resolve()).resolves`},
			{Code: `expect(Promise.resolve()).not`},
			{Code: `expect(1).not.resolves.toBe(1)`},
			{Code: `expect(1).resolves.resolves.toBe(1)`},
			// Locks in upstream create() arm 1: the expect head must itself be called
			{Code: `expect.assertions(1)`},
			{Code: `expect.resolves.toBe(1)`},
			// Locks in upstream create() arm 2: a promise with a modifier is valid
			{Code: `expect(Promise.resolve()).rejects.not.toBe(1)`},
			{Code: `class P extends Promise<number> {}; expect(P).resolves.toBe(1)`},
			{Code: `declare const p: Promise<1> & { x: 1 }; expect(p).resolves.toBe(1)`},
			{Code: `expect(Promise.resolve()).resolves.toBe(1).catch(() => {})`},
			// Locks in upstream create() arm 3: a non-promise without a modifier is valid
			{Code: `expect(1).not.toBe(1)`},
			{Code: `declare const p: any; expect(p).toBe(1)`},
			{Code: `declare const p: unknown; expect(p).toBe(1)`},
			{Code: `declare const p: never; expect(p).toBe(1)`},
			{Code: `expect(Promise).toBe(1)`},
			{Code: `expect([Promise.resolve(1)]).toEqual([1])`},
			{Code: `expect(expect(Promise.resolve()).resolves).toBe(1)`},
			// Locks in isBuiltinSymbolLike(): only the default-library Promise counts
			{Code: `declare class Promise2<T> {}; declare const p: Promise2<number>; expect(p).toBe(1)`},
			{Code: `declare const p: Readonly<Promise<number>>; expect(p).toBe(1)`},
			{Code: `declare const p: PromiseLike<number>; expect(p).toBe(1)`},
			{Code: `expect(new Promise<void>(() => {}) as PromiseLike<void>).toBe(1)`},
			// Locks in isBuiltinSymbolLikeRecurser() union arm: every member must be a promise
			{Code: `declare const p: Promise<1> | undefined; expect(p).toBe(1)`},
			{Code: `declare const p: PromiseLike<number> | Promise<number>; expect(p).toBe(1)`},
			{Code: `declare const o: { p?: Promise<number> }; expect(o.p).toBe(1)`},
			{Code: `declare const o: { p: Promise<number> } | undefined; expect(o?.p).toBe(1)`},
			// Locks in isThenableType(): the option accepts both-callback thenables
			{Code: `declare const p: PromiseLike<number>; expect(p).resolves.toBe(1)`, Options: checkThenables},
			// Locks in isThenableType() isFunctionParam(): each callback parameter must be callable
			{Code: `declare const p: { then(...a: Array<() => void>): void }; expect(p).toBe(1)`, Options: checkThenables},
			{Code: `declare const p: { then(a: () => void, b: string): void }; expect(p).toBe(1)`, Options: checkThenables},
			{Code: `declare const p: { then(a: Function, b: Function): void }; expect(p).toBe(1)`, Options: checkThenables},
			// ---- Dimension 4: a non-null assertion inside the chain is not a Jest call ----
			{Code: `expect(Promise.resolve())!.toBe(1)`},
			// ---- Dimension 4: computed promise modifier on a promise ----
			{Code: `expect(Promise.resolve())['resolves'].toBe(1)`},
			// ---- Real-user: awaited values and awaited resolves assertions ----
			{Code: `it('x', async () => { await expect(fetch('x')).resolves.toBeDefined() })`},
			{Code: `it('x', async () => { const r = await fetch('x'); await expect(r.json()).resolves.toEqual({}) })`},
			{Code: `it('x', async () => { expect(await Promise.all([Promise.resolve(1)])).toEqual([1]) })`},
			{Code: `expect(Promise.all([Promise.resolve(1)])).resolves.toEqual([1])`},
		},
		[]rule_tester.InvalidTestCase{
			// Locks in upstream create(): a missing argument is not a promise
			{Code: `expect().resolves.toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{unneeded("resolves", 1, 10, 1, 18)}},
			{Code: `expect().rejects.not.toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{unneeded("rejects", 1, 10, 1, 17)}},
			// Locks in upstream modifiers.find(): `not` is skipped when choosing the modifier
			{Code: `expect(Promise.resolve()).not.toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 1, 1, 38)}},
			{Code: `expect(1).rejects.not.toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{unneeded("rejects", 1, 11, 1, 18)}},
			// Locks in isBuiltinSymbolLike() on any/unknown-like subjects with a modifier
			{Code: `declare const p: any; expect(p).resolves.toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{unneeded("resolves", 1, 33, 1, 41)}},
			{Code: `expect(Promise.resolve() as unknown).resolves.toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{unneeded("resolves", 1, 38, 1, 46)}},
			{Code: `declare const x: never; expect(x).resolves.toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{unneeded("resolves", 1, 35, 1, 43)}},
			// Locks in isBuiltinSymbolLikeRecurser() base-type arm: Promise subclasses and their constructors
			{Code: `class P extends Promise<number> {}; expect(P).toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 37, 1, 54)}},
			{Code: `interface MyP extends Promise<number> {}; declare const p: MyP; expect(p).toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 65, 1, 82)}},
			{Code: `type MyP = Promise<number>; declare const p: MyP; expect(p).toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 51, 1, 68)}},
			{Code: `expect(Promise).resolves.toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{unneeded("resolves", 1, 17, 1, 25)}},
			// Locks in isBuiltinSymbolLikeRecurser() intersection and union arms
			{Code: `declare const p: Promise<1> & { x: 1 }; expect(p).toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 41, 1, 58)}},
			{Code: `declare const p: Promise<1> | undefined; expect(p).resolves.toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{unneeded("resolves", 1, 52, 1, 60)}},
			{Code: `declare const p: PromiseLike<number> | Promise<number>; expect(p).resolves.toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{unneeded("resolves", 1, 67, 1, 75)}},
			{Code: `declare function f(): Promise<void> | void; expect(f()).resolves.toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{unneeded("resolves", 1, 57, 1, 65)}},
			// Locks in isBuiltinSymbolLikeRecurser() type-parameter arm
			{Code: `function f<U extends Promise<number>, T extends U>(v: T) { expect(v).toBe(1) }`, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 60, 1, 77)}},
			{Code: `function f<T>(v: T) { expect(v).resolves.toBe(1) }`, Errors: []rule_tester.InvalidTestCaseError{unneeded("resolves", 1, 33, 1, 41)}},
			{Code: `function f<T>(v: T) { expect(v).resolves.toBe(1) }`, Options: checkThenables, Errors: []rule_tester.InvalidTestCaseError{unneeded("resolves", 1, 33, 1, 41)}},
			// Locks in the checkThenables option: thenables are only promises when enabled
			{Code: `declare const p: PromiseLike<number>; expect(p).resolves.toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{unneeded("resolves", 1, 49, 1, 57)}},
			{Code: `declare const p: PromiseLike<number>; expect(p).toBe(1)`, Options: checkThenables, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 39, 1, 56)}},
			{Code: `declare const p: PromiseLike<number> | Promise<number>; expect(p).toBe(1)`, Options: checkThenables, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 57, 1, 74)}},
			{Code: `declare const p: PromiseLike<number> & { x: 1 }; expect(p).toBe(1)`, Options: checkThenables, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 50, 1, 67)}},
			{Code: `function f<T extends PromiseLike<number>>(v: T) { expect(v).toBe(1) }`, Options: checkThenables, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 51, 1, 68)}},
			{Code: `function f<U extends PromiseLike<number>, T extends U>(v: T) { expect(v).toBe(1) }`, Options: checkThenables, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 64, 1, 81)}},
			// Locks in isThenableType(): object literals, optional `then`, overloaded `then`, and nullable callbacks
			{Code: `expect({ then(a: () => void, b: () => void) {} }).toBe(1)`, Options: checkThenables, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 1, 1, 58)}},
			{Code: `declare const p: { then?(a: () => void, b: () => void): void }; expect(p).toBe(1)`, Options: checkThenables, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 65, 1, 82)}},
			{Code: `declare const p: { then: ((a: () => void, b: () => void) => void) | ((a: string) => void) }; expect(p).toBe(1)`, Options: checkThenables, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 94, 1, 111)}},
			{Code: `declare const p: { then(a: (() => void) | string, b: (() => void) | undefined): void }; expect(p).toBe(1)`, Options: checkThenables, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 89, 1, 106)}},
			// Functions are subjects, not promises, even when they return one
			{Code: `expect(async () => 1).rejects.toThrow()`, Errors: []rule_tester.InvalidTestCaseError{unneeded("rejects", 1, 23, 1, 30)}},
			{Code: `expect(() => Promise.resolve(1)).resolves.toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{unneeded("resolves", 1, 34, 1, 42)}},
			{Code: `declare const p: Awaited<Promise<number>>; expect(p).resolves.toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{unneeded("resolves", 1, 54, 1, 62)}},
			// ---- Dimension 4: parenthesized callee, subject, and chain ----
			{Code: `(expect)(Promise.resolve()).toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 1, 1, 36)}},
			{Code: `expect((Promise.resolve())).toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 1, 1, 36)}},
			{Code: `(expect(Promise.resolve())).toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 1, 1, 36)}},
			{Code: `(expect(1).resolves).toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{unneeded("resolves", 1, 12, 1, 20)}},
			// ---- Dimension 4: computed modifier keys report the key expression ----
			{Code: `expect(1)['resolves'].toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{unneeded("resolves", 1, 11, 1, 21)}},
			{Code: "expect(1)[`rejects`].not.toBe(1)", Errors: []rule_tester.InvalidTestCaseError{unneeded("rejects", 1, 11, 1, 20)}},
			{Code: `expect(1).resolves['not'].toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{unneeded("resolves", 1, 11, 1, 19)}},
			// ---- Dimension 4: type wrappers on the subject ----
			{Code: `expect(Promise.resolve() as any).resolves.toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{unneeded("resolves", 1, 34, 1, 42)}},
			{Code: `expect(1 as unknown as Promise<number>).toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 1, 1, 48)}},
			{Code: `expect(Promise.resolve() satisfies Promise<void>).toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 1, 1, 58)}},
			{Code: `expect(<Promise<number>>(<unknown>1)).toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 1, 1, 46)}},
			{Code: `expect(Promise.resolve()!).toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 1, 1, 35)}},
			{Code: `declare const o: { p: Promise<number> }; expect(o?.p).toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 42, 1, 62)}},
			// ---- Dimension 4: spread and extra arguments use the first argument ----
			{Code: `expect(...[Promise.resolve(1)]).toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 1, 1, 40)}},
			{Code: `expect(...[1]).resolves.toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{unneeded("resolves", 1, 16, 1, 24)}},
			{Code: `expect(Promise.resolve(), 'message').toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 1, 1, 45)}},
			// ---- Dimension 4: optional chains and type arguments ----
			{Code: `expect(Promise.resolve())?.toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 1, 1, 35)}},
			{Code: `expect(Promise.resolve()).toBe?.(1)`, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 1, 1, 36)}},
			{Code: `expect(1)?.resolves.toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{unneeded("resolves", 1, 12, 1, 20)}},
			{Code: `expect(1).resolves?.toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{unneeded("resolves", 1, 11, 1, 19)}},
			{Code: `expect<Promise<number>>(Promise.resolve(1)).toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 1, 1, 52)}},
			// ---- Dimension 4: calls chained after the matcher are parsed as their own expect calls ----
			{Code: `expect(Promise.resolve()).toBe(1).then(() => {})`, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 1, 1, 49), poorly(1, 1, 1, 34)}},
			{Code: `expect(1).resolves.toBe(1).then(() => {})`, Errors: []rule_tester.InvalidTestCaseError{unneeded("resolves", 1, 11, 1, 19), unneeded("resolves", 1, 11, 1, 19)}},
			{Code: `expect(Promise.resolve()).toBe(1)()`, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 1, 1, 36), poorly(1, 1, 1, 34)}},
			{Code: `expect(Promise.resolve()).toEqual(expect(1).resolves.toBe(1))`, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 1, 1, 62), unneeded("resolves", 1, 45, 1, 53)}},
			// ---- Dimension 4: renamed and aliased expect bindings ----
			{Code: `import { expect as check } from '@jest/globals'; check(Promise.resolve()).toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 50, 1, 82)}},
			{Code: `import { expect as check } from '@jest/globals'; check(1).resolves.toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{unneeded("resolves", 1, 59, 1, 67)}},
			{
				Code:     `check(Promise.resolve()).toBe(1)`,
				Settings: map[string]any{"jest": map[string]any{"globalAliases": map[string]any{"expect": []any{"check"}}}},
				Errors:   []rule_tester.InvalidTestCaseError{poorly(1, 1, 1, 33)},
			},
			// ---- Real-user: a request promise asserted directly inside a test ----
			{Code: `it('x', async () => { expect(fetch('x')).toBeDefined() })`, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 23, 1, 55)}},
			{Code: `declare const p: Promise<number>; describe('x', () => { it('y', () => { expect(p).toBe(1) }) })`, Errors: []rule_tester.InvalidTestCaseError{poorly(1, 73, 1, 90)}},
		},
	)
}
