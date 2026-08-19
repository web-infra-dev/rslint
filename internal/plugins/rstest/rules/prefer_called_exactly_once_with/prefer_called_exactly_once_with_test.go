package prefer_called_exactly_once_with_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_called_exactly_once_with"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferCalledExactlyOnceWithRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t,
		&prefer_called_exactly_once_with.PreferCalledExactlyOnceWithRule,
		[]rule_tester.ValidTestCase{
			{Code: `expect(fn).toHaveBeenCalledExactlyOnceWith();`},
			{Code: `expect(x).toHaveBeenCalledOnce();`},
			{Code: `expect(x).toHaveBeenCalledWith('hoge');`},
			{Code: `expect(x).toHaveBeenCalledOnce(); expect(y).toHaveBeenCalledWith('hoge');`},
			{Code: `expect(x).toHaveBeenCalledWith('hoge'); expect(x).toHaveBeenCalledWith('foo');`},
			{Code: `expect(x).toHaveBeenCalledOnce(); expect(x).not.toHaveBeenCalledWith('hoge');`},
			{Code: `expect(x).not.toHaveBeenCalledOnce(); expect(x).toHaveBeenCalledWith('hoge');`},
			// Both negated is still not mergeable: `¬once ∧ ¬with` is not
			// `¬(once ∧ with)`.
			{Code: `expect(x).not.toHaveBeenCalledOnce(); expect(x).not.toHaveBeenCalledWith('hoge');`},
			// A modifier changes what the assertion says about the target, so
			// two assertions that disagree on modifiers are not two halves of
			// one claim.
			{Code: `expect(x).resolves.toHaveBeenCalledOnce(); expect(x).toHaveBeenCalledWith('hoge');`},
			{Code: `expect(x).toHaveBeenCalledOnce(); expect(x).rejects.toHaveBeenCalledWith('hoge');`},
			{Code: `expect(x).resolves.toHaveBeenCalledOnce(); expect(x).rejects.toHaveBeenCalledWith('hoge');`},
			// Awaiting one half and not the other is not a pair: dropping the
			// awaited statement would leave a floating promise whose failure
			// escapes as an unhandled rejection.
			{Code: `async function f() { await expect(x).resolves.toHaveBeenCalledOnce(); expect(x).resolves.toHaveBeenCalledWith('a'); }`},
			{Code: `async function f() { expect(x).resolves.toHaveBeenCalledOnce(); await expect(x).resolves.toHaveBeenCalledWith('a'); }`},
			{Code: `expect(x).toHaveBeenCalledOnce(); x.mockRestore(); expect(x).toHaveBeenCalledWith('hoge');`},
			{Code: `expect(x).toHaveBeenCalledOnce(); x.mockReset(); expect(x).toHaveBeenCalledWith('hoge');`},
			{Code: `expect(x).toHaveBeenCalledOnce(); x.mockClear(); expect(x).toHaveBeenCalledWith('hoge');`},
			{Code: `expect(obj.fn).toHaveBeenCalledOnce(); obj.fn.mockClear(); expect(obj.fn).toHaveBeenCalledWith('hoge');`},
			// A reset nested in a block still splits the two call histories.
			{Code: `expect(x).toHaveBeenCalledOnce(); if (c) { x.mockClear(); } expect(x).toHaveBeenCalledWith('hoge');`},
			{Code: `expect.soft(x).toHaveBeenCalledOnce(); expect(x).toHaveBeenCalledWith('hoge');`},
			{Code: `test('x', ctx => { ctx.expect(x).toHaveBeenCalledOnce(); expect(x).toHaveBeenCalledWith('hoge'); });`},
			{Code: `import { expect } from 'vitest'; expect(x).toHaveBeenCalledOnce(); expect(x).toHaveBeenCalledWith('a');`},
			{Code: `import { expect } from '@jest/globals'; expect(x).toHaveBeenCalledOnce(); expect(x).toHaveBeenCalledWith('a');`},
			{Code: `import { expect } from '@playwright/test'; expect(x).toHaveBeenCalledOnce(); expect(x).toHaveBeenCalledWith('a');`},
			{Code: `const expect = makeExpect(); expect(x).toHaveBeenCalledOnce(); expect(x).toHaveBeenCalledWith('a');`},
			{Code: `test('a', () => { expect(x).toHaveBeenCalledOnce(); }); test('b', () => { expect(x).toHaveBeenCalledWith('a'); });`},
			{Code: `expect(x).toHaveBeenCalledOnce(); expect(x).toHaveBeenCalledWith('a'); expect(x).toHaveBeenCalledWith('b');`},
			// Chai-style spellings, which Rstest ships through
			// @vitest/expect's ChaiStyleAssertions.
			{Code: `expect(x).to.have.been.calledOnceWith('a');`},
			{Code: `expect(x).to.have.been.calledOnce;`},
			{Code: `expect(x).to.have.been.calledWith('a');`},
			{Code: `expect(x).to.have.been.calledOnce; expect(y).to.have.been.calledWith('a');`},
			{Code: `expect(x).to.have.been.calledOnce; x.mockClear(); expect(x).to.have.been.calledWith('a');`},
			// toBeCalledWith is an alias of toHaveBeenCalledWith; canonicalizing
			// spellings is no-alias-methods' job, not this rule's.
			{Code: `expect(x).toHaveBeenCalledOnce(); expect(x).toBeCalledWith('a');`},
			// A reset receiver that differs from the target only in trivia or
			// parentheses still splits the two call histories.
			{Code: `expect(obj.fn).toHaveBeenCalledOnce(); obj /* keep */ .fn.mockClear(); expect(obj.fn).toHaveBeenCalledWith('a');`},
			{Code: `expect(obj.fn).toHaveBeenCalledOnce(); (obj.fn).mockClear(); expect(obj.fn).toHaveBeenCalledWith('a');`},
			{Code: `expect(obj.fn).toHaveBeenCalledOnce(); (obj.fn as Mock).mockClear(); expect(obj.fn).toHaveBeenCalledWith('a');`},
			{Code: `expect(mocks['a']).toHaveBeenCalledOnce(); mocks["a"].mockClear(); expect(mocks['a']).toHaveBeenCalledWith('x');`},
			// Chai allows a modifier between matchers, so `not` can follow the
			// first matcher. The merged matcher would assert the opposite.
			{Code: `expect(x).to.have.been.calledOnce.and.not.calledWith('a');`},
			// Nothing between the two assertions may run code that could reach
			// the target: a rebind and a further call both invalidate the merge,
			// and neither is decidable through an arbitrary call.
			{Code: `expect(x).toHaveBeenCalledOnce(); x = createMock(); expect(x).toHaveBeenCalledWith('a');`},
			{Code: `expect(x).toHaveBeenCalledOnce(); x('b'); expect(x).toHaveBeenCalledWith('a');`},
			{Code: `expect(x).toHaveBeenCalledOnce(); doMoreWork(); expect(x).toHaveBeenCalledWith('a');`},
			// A reset of some other mock cannot reach the target, but nothing
			// here proves `y.mockClear` is that reset, so the pair is left alone.
			{Code: `expect(x).toHaveBeenCalledOnce(); y.mockClear(); expect(x).toHaveBeenCalledWith('hoge');`},
			// A call the rule cannot resolve to the default library could reach
			// the target through anything it calls.
			{Code: `expect(x).toHaveBeenCalledOnce(); const y = compute(); expect(x).toHaveBeenCalledWith('a');`},
			{Code: `expect(x).toHaveBeenCalledOnce(); expect(y).toEqual(makeExpected()); expect(x).toHaveBeenCalledWith('a');`},
			// A default-library call handed something callable can invoke it,
			// and what it invokes may reach the target.
			{Code: `expect(x).toHaveBeenCalledOnce(); [1, 2].forEach(cb); expect(x).toHaveBeenCalledWith('a');`},
			// An argument the checker cannot pin down stays callable, so a
			// library call handed one is not inert either.
			{Code: `expect(x).toHaveBeenCalledOnce(); console.log(unresolved); expect(x).toHaveBeenCalledWith('a');`},
			{Code: `expect(x).toHaveBeenCalledOnce(); setTimeout(() => x('b'), 0); expect(x).toHaveBeenCalledWith('a');`},
			{Code: `expect(x).toHaveBeenCalledOnce(); console.log(x); expect(x).toHaveBeenCalledWith('a');`},
			// A `console` of the author's own resolves to their binding, not the
			// library's, so nothing here says what it runs.
			{Code: `function f() { const console = makeLogger(); expect(x).toHaveBeenCalledOnce(); console.log('checkpoint'); expect(x).toHaveBeenCalledWith('a'); }`},
			// `var` is hoisted, so unlike `const` and `let` it can rebind a name
			// the assertions read.
			{Code: `expect(x).toHaveBeenCalledOnce(); var x = makeMock; expect(x).toHaveBeenCalledWith('a');`},
			// An intervening assertion that executes what it asserts on, or
			// awaits it, hands control to code this rule cannot see.
			{Code: `expect(x).toHaveBeenCalledOnce(); expect(callX).toThrow(); expect(x).toHaveBeenCalledWith('a');`},
			{Code: `async function f() { expect(x).toHaveBeenCalledOnce(); await expect(p).resolves.toBe(1); expect(x).toHaveBeenCalledWith('a'); }`},
			{Code: `expect(x).toHaveBeenCalledOnce(); expect(y).toEqual(makeExpected()); expect(x).toHaveBeenCalledWith('a');`},
			{Code: `expect(x).toHaveBeenCalledOnce(); expect(callX).toThrowErrorMatchingSnapshot(); expect(x).toHaveBeenCalledWith('a');`},
			{Code: `expect(x).toHaveBeenCalledOnce(); expect(callX).toThrowErrorMatchingInlineSnapshot(); expect(x).toHaveBeenCalledWith('a');`},
			// Chai's own assertions run the author's code just as the
			// jest-style ones do: `satisfy` calls its argument with the
			// subject, and `change` calls the subject itself.
			{Code: `expect(x).toHaveBeenCalledOnce(); expect(v).to.satisfy(pred); expect(x).toHaveBeenCalledWith('a');`},
			{Code: `expect(x).toHaveBeenCalledOnce(); expect(v).to.satisfies(pred); expect(x).toHaveBeenCalledWith('a');`},
			{Code: `expect(x).toHaveBeenCalledOnce(); expect(callX).to.change(obj, 'value'); expect(x).toHaveBeenCalledWith('a');`},
			{Code: `expect(x).toHaveBeenCalledOnce(); expect(callX).to.increase(obj, 'value'); expect(x).toHaveBeenCalledWith('a');`},
			{Code: `expect(x).toHaveBeenCalledOnce(); expect(callX).to.decrease(obj, 'value'); expect(x).toHaveBeenCalledWith('a');`},
			// `eval` is a default-library call, but the source it runs arrives
			// as a string, so nothing callable has to be handed to it.
			{Code: `expect(x).toHaveBeenCalledOnce(); eval('x("b")'); expect(x).toHaveBeenCalledWith('a');`},
			{Code: `expect(x).toHaveBeenCalledOnce(); globalThis.eval('x("b")'); expect(x).toHaveBeenCalledWith('a');`},
			// A polled assertion retries until it passes, so the two halves
			// can settle against different call histories.
			{Code: `async function f() { await expect.poll(getSpy).toHaveBeenCalledOnce(); await expect.poll(getSpy).toHaveBeenCalledWith('a'); }`},
			{Code: `async function f() { await expect.element(locator).toHaveBeenCalledOnce(); await expect.element(locator).toHaveBeenCalledWith('a'); }`},
			// Only the value constructors Rstest ships are exempt under the
			// assertion's root name. `toSatisfy` calls its predicate when the
			// comparison runs, just as the instance matcher of that name does.
			{Code: `expect(x).toHaveBeenCalledOnce(); expect(y).toEqual(expect.toSatisfy(pred)); expect(x).toHaveBeenCalledWith('a');`},
			{Code: `test('t', ({ expect }) => { expect(x).toHaveBeenCalledOnce(); expect(y).toEqual(expect.myAsymmetric()); expect(x).toHaveBeenCalledWith('a'); });`},
		},
		[]rule_tester.InvalidTestCase{
			// An argument that is not stable under a second evaluation is still
			// worth reporting, but folding the pair would drop an evaluation.
			{
				Code:   `expect(getMock()).toHaveBeenCalledOnce(); expect(getMock()).toHaveBeenCalledWith('a');`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledExactlyOnceWith"}},
			},
			{
				Code:   `expect(mocks[next()]).toHaveBeenCalledOnce(); expect(mocks[next()]).toHaveBeenCalledWith('a');`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledExactlyOnceWith"}},
			},
			// A literal element access is stable, so the merge is still fixed.
			{
				Code:   `expect(mocks['a']).toHaveBeenCalledOnce(); expect(mocks['a']).toHaveBeenCalledWith('x');`,
				Output: []string{` expect(mocks['a']).toHaveBeenCalledExactlyOnceWith('x');`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledExactlyOnceWith"}},
			},
			{
				Code:   `expect(x).toHaveBeenCalledOnce(); expect(x).toHaveBeenCalledWith('hoge');`,
				Output: []string{` expect(x).toHaveBeenCalledExactlyOnceWith('hoge');`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledExactlyOnceWith", Column: 45}},
			},
			{
				Code:   `expect(x).toHaveBeenCalledWith('hoge', 123); expect(x).toHaveBeenCalledOnce();`,
				Output: []string{`expect(x).toHaveBeenCalledExactlyOnceWith('hoge', 123); `},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledExactlyOnceWith"}},
			},
			// A statement that runs nothing of the author's may sit between the
			// two halves: assertions on other mocks, which is how they are
			// commonly grouped, and calls into the default library, which have
			// nothing of the author's to invoke.
			{
				Code: `expect(x).toHaveBeenCalledOnce();
console.log('checkpoint');
expect(x).toHaveBeenCalledWith('a');`,
				Output: []string{`console.log('checkpoint');
expect(x).toHaveBeenCalledExactlyOnceWith('a');`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledExactlyOnceWith"}},
			},
			{
				Code: `expect(x).toHaveBeenCalledOnce();
const parsed = JSON.parse('{}');
expect(x).toHaveBeenCalledWith('a');`,
				Output: []string{`const parsed = JSON.parse('{}');
expect(x).toHaveBeenCalledExactlyOnceWith('a');`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledExactlyOnceWith"}},
			},
			// An unawaited poll builds its promise only in then, catch or
			// finally, so the factory is never called and the statement runs
			// nothing between the two halves.
			{
				Code: `expect(x).toHaveBeenCalledOnce();
expect.poll(getSpy).toBe(1);
expect(x).toHaveBeenCalledWith('a');`,
				Output: []string{`expect.poll(getSpy).toBe(1);
expect(x).toHaveBeenCalledExactlyOnceWith('a');`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledExactlyOnceWith"}},
			},
			// A hoisted declaration of a name neither assertion reads cannot
			// change what they assert on.
			{
				Code: `expect(x).toHaveBeenCalledOnce();
var hoge = 'foo';
expect(x).toHaveBeenCalledWith('a');`,
				Output: []string{`var hoge = 'foo';
expect(x).toHaveBeenCalledExactlyOnceWith('a');`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledExactlyOnceWith"}},
			},
			{
				Code: `expect(x).toHaveBeenCalledOnce();
expect(y).toEqual({ id: 1 });
expect(x).toHaveBeenCalledWith('a');`,
				Output: []string{`expect(y).toEqual({ id: 1 });
expect(x).toHaveBeenCalledExactlyOnceWith('a');`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledExactlyOnceWith"}},
			},
			{
				Code: `expect(x).toHaveBeenCalledOnce();
expect(y).toHaveBeenCalledWith(expect.any(String));
expect(x).toHaveBeenCalledWith('a');`,
				Output: []string{`expect(y).toHaveBeenCalledWith(expect.any(String));
expect(x).toHaveBeenCalledExactlyOnceWith('a');`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledExactlyOnceWith"}},
			},
			{
				Code: `expect(x).toHaveBeenCalledOnce();
const hoge = 'foo';
expect(x).toHaveBeenCalledWith('hoge', 123);`,
				Output: []string{`const hoge = 'foo';
expect(x).toHaveBeenCalledExactlyOnceWith('hoge', 123);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledExactlyOnceWith"}},
			},
			{
				Code: `expect(x).toHaveBeenCalledOnce();
expect(x).toHaveBeenCalledWith<
  [string, number]
>('hoge', /* keep */ 123);`,
				Output: []string{`expect(x).toHaveBeenCalledExactlyOnceWith<
  [string, number]
>('hoge', /* keep */ 123);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledExactlyOnceWith"}},
			},
			{
				Code: `test('x', ({ expect }) => {
  expect(x).toHaveBeenCalledOnce();
  expect(x).toHaveBeenCalledWith('a');
});`,
				Output: []string{`test('x', ({ expect }) => {
  expect(x).toHaveBeenCalledExactlyOnceWith('a');
});`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledExactlyOnceWith"}},
			},
			{
				Code: `import { expect as check } from '@rstest/core';
check(x).toHaveBeenCalledOnce();
check(x).toHaveBeenCalledWith('a');`,
				Output: []string{`import { expect as check } from '@rstest/core';
check(x).toHaveBeenCalledExactlyOnceWith('a');`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledExactlyOnceWith"}},
			},
			{
				Code: `import.meta.rstest.expect(x).toHaveBeenCalledOnce();
import.meta.rstest.expect(x).toHaveBeenCalledWith('a');`,
				Output: []string{`import.meta.rstest.expect(x).toHaveBeenCalledExactlyOnceWith('a');`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledExactlyOnceWith"}},
			},
			{
				Code: `expect(x).toHaveBeenCalledOnce();
expect(x)['toHaveBeenCalledWith']('a');`,
				Output: []string{`expect(x)['toHaveBeenCalledExactlyOnceWith']('a');`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledExactlyOnceWith"}},
			},
			// Two independent pairs in one block report in source order.
			{
				Code: `expect(aaa).toHaveBeenCalledOnce();
expect(aaa).toHaveBeenCalledWith('a');
expect(bbb).toHaveBeenCalledWith('b');
expect(bbb).toHaveBeenCalledOnce();`,
				Output: []string{`expect(aaa).toHaveBeenCalledExactlyOnceWith('a');
expect(bbb).toHaveBeenCalledExactlyOnceWith('b');
`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferCalledExactlyOnceWith", Line: 2},
					{MessageId: "preferCalledExactlyOnceWith", Line: 4},
				},
			},
			// A statement that shares its line keeps the conservative removal:
			// the fix must not reach into code it did not inspect.
			{
				Code:   `foo(); expect(x).toHaveBeenCalledOnce(); expect(x).toHaveBeenCalledWith('a');`,
				Output: []string{`foo();  expect(x).toHaveBeenCalledExactlyOnceWith('a');`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledExactlyOnceWith"}},
			},
			{
				Code: `expect(x).toHaveBeenCalledOnce(); // keep me
expect(x).toHaveBeenCalledWith('a');`,
				Output: []string{` // keep me
expect(x).toHaveBeenCalledExactlyOnceWith('a');`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledExactlyOnceWith"}},
			},
			// Matching promise modifiers on both halves do merge: the awaited
			// value and the conjunction are unchanged.
			{
				Code: `expect(x).resolves.toHaveBeenCalledOnce();
expect(x).resolves.toHaveBeenCalledWith('a');`,
				Output: []string{`expect(x).resolves.toHaveBeenCalledExactlyOnceWith('a');`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledExactlyOnceWith"}},
			},
			{
				Code: `expect(x).rejects.toHaveBeenCalledWith('a');
expect(x).rejects.toHaveBeenCalledOnce();`,
				Output: []string{`expect(x).rejects.toHaveBeenCalledExactlyOnceWith('a');
`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledExactlyOnceWith"}},
			},
			// Awaited on both halves: the merge keeps the await.
			{
				Code: `async function f() {
  await expect(x).resolves.toHaveBeenCalledOnce();
  await expect(x).resolves.toHaveBeenCalledWith('a');
}`,
				Output: []string{`async function f() {
  await expect(x).resolves.toHaveBeenCalledExactlyOnceWith('a');
}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledExactlyOnceWith"}},
			},
			// Chai-style spellings merge into calledOnceWith.
			{
				Code: `expect(x).to.have.been.calledOnce;
expect(x).to.have.been.calledWith('a');`,
				Output: []string{`expect(x).to.have.been.calledOnceWith('a');`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferCalledExactlyOnceWith",
					Message:   "Using `calledOnce` and `calledWith` on the same target; prefer `calledOnceWith` instead.",
				}},
			},
			// Mixed spellings still assert on the same spy, so they merge; the
			// surviving argument-side call decides the resulting spelling.
			{
				Code: `expect(x).to.have.been.calledOnce;
expect(x).toHaveBeenCalledWith('a');`,
				Output: []string{`expect(x).toHaveBeenCalledExactlyOnceWith('a');`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferCalledExactlyOnceWith",
					Message:   "Using `calledOnce` and `toHaveBeenCalledWith` on the same target; prefer `toHaveBeenCalledExactlyOnceWith` instead.",
				}},
			},
			{
				Code: `expect(x).toHaveBeenCalledOnce();
expect(x).to.have.been.calledWith('a');`,
				Output: []string{`expect(x).to.have.been.calledOnceWith('a');`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferCalledExactlyOnceWith",
					Message:   "Using `toHaveBeenCalledOnce` and `calledWith` on the same target; prefer `calledOnceWith` instead.",
				}},
			},
			{
				Code: `expect(x).to.have.been.calledWith('a');
expect(x).to.have.been.calledOnce;`,
				Output: []string{`expect(x).to.have.been.calledOnceWith('a');
`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledExactlyOnceWith"}},
			},
			// A chain that asserts more than once is reported without a fix:
			// folding the statement away would delete `to.be.ok` with it.
			{
				Code: `expect(x).to.have.been.calledOnce.and.to.be.ok;
expect(x).to.have.been.calledWith('a');`,
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferCalledExactlyOnceWith",
					Message:   "Using `calledOnce` and `calledWith` on the same target; prefer `calledOnceWith` instead.",
					Line:      2,
				}},
			},
			{
				Code: `expect(x).toHaveBeenCalledOnce();
expect(x).to.have.been.calledWith('a').and.to.be.ok;`,
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledExactlyOnceWith", Line: 2}},
			},
			// A Chai chain that already states both halves merges with itself;
			// rewriting the chain in place is left to the author.
			{
				Code:   `expect(x).to.have.been.calledOnce.and.calledWith('a');`,
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferCalledExactlyOnceWith",
					Message:   "Using `calledOnce` and `calledWith` on the same target; prefer `calledOnceWith` instead.",
				}},
			},
			{
				Code:   `expect(x).to.have.been.calledWith('a').and.calledOnce;`,
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledExactlyOnceWith"}},
			},
			// A self-contained chain is reported on its own and does not count
			// toward the pair tally for the same target.
			{
				Code: `expect(x).to.have.been.calledOnce.and.calledWith('a');
expect(y).toHaveBeenCalledOnce();
expect(y).toHaveBeenCalledWith('b');`,
				Output: []string{`expect(x).to.have.been.calledOnce.and.calledWith('a');
expect(y).toHaveBeenCalledExactlyOnceWith('b');`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferCalledExactlyOnceWith", Line: 1},
					{MessageId: "preferCalledExactlyOnceWith", Line: 3},
				},
			},
		},
	)
}
