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
			// resolves / rejects change what the assertion says about the
			// target, so the two statements are not two halves of one claim.
			{Code: `expect(x).resolves.toHaveBeenCalledOnce(); expect(x).toHaveBeenCalledWith('hoge');`},
			{Code: `expect(x).toHaveBeenCalledOnce(); expect(x).rejects.toHaveBeenCalledWith('hoge');`},
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
		},
		[]rule_tester.InvalidTestCase{
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
			{
				Code: `expect(x).toHaveBeenCalledOnce();
const hoge = 'foo';
expect(x).toHaveBeenCalledWith('hoge', 123);`,
				Output: []string{`const hoge = 'foo';
expect(x).toHaveBeenCalledExactlyOnceWith('hoge', 123);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledExactlyOnceWith"}},
			},
			{
				Code:   `expect(x).toHaveBeenCalledOnce(); y.mockClear(); expect(x).toHaveBeenCalledWith('hoge');`,
				Output: []string{` y.mockClear(); expect(x).toHaveBeenCalledExactlyOnceWith('hoge');`},
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
