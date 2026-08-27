// TestRequireAwaitedExpectPollExtras locks in branches and edge shapes that
// the upstream test suite doesn't exercise: the Rstest expect source matrix,
// the handled-position set this port widens beyond `await` / `return`, the
// assertion-factory boundary, the comma-expression walk, and the Dimension 4
// receiver and accessor forms. Each case carries an inline comment pointing at
// the specific branch, matrix row or tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in.
package require_awaited_expect_poll

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRequireAwaitedExpectPollExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RequireAwaitedExpectPollRule,
		[]rule_tester.ValidTestCase{
			// ---- A. The two positions upstream accepts ----
			// Locks in upstream isHandled arm 1: AwaitExpression.
			{Code: `async function run() { await expect.poll(() => el).toBeVisible(); }`},
			{Code: `async function run() { await expect.element(el).toBeVisible(); }`},
			// Locks in upstream isHandled arm 2: ReturnStatement.
			{Code: `function run() { return expect.poll(() => el).toBeVisible(); }`},
			{Code: `function run() { return expect.element(el).toBeVisible(); }`},

			// ---- B. Positions this port adds to the handled set ----
			// The promise is bound or passed on rather than dropped, so
			// reporting would be a false positive. Upstream reports every one
			// of these.
			{Code: `const assertion = expect.poll(() => el).toBeVisible();`},
			{Code: `let assertion; assertion = expect.poll(() => el).toBeVisible();`},
			{Code: `let assertions = []; assertions[0] = expect.poll(() => el).toBeVisible();`},
			{Code: `const assertVisible = () => expect.element(el).toBeVisible();`},
			{Code: `async function run() { await Promise.all([expect.poll(() => el).toBeVisible()]); }`},
			{Code: `async function run() { await Promise.allSettled([expect.poll(() => el).toBeVisible(), expect.element(el).toBeVisible()]); }`},
			{Code: `const assertions = [expect.poll(() => el).toBeVisible()];`},
			{Code: `function* pending() { yield expect.poll(() => el).toBeVisible(); }`},
			{Code: `const pending = { visible: expect.element(el).toBeVisible() };`},
			{Code: `collect(expect.poll(() => el).toBeVisible());`},

			// ---- B2. Destructuring and parameter defaults bind the promise too ----
			// Locks in a review finding on PR #1920: BindingElement,
			// ParameterDeclaration and ShorthandPropertyAssignment initializers
			// are all handled positions, on the same theory as B — the default
			// only runs when it binds the promise to a name.
			{Code: `const [p = expect.poll(() => el).toBeVisible()] = values;`},
			{Code: `const {p = expect.poll(() => el).toBeVisible()} = values;`},
			// Nested pattern: the default sits two BindingElement levels deep.
			{Code: `const [[p = expect.poll(() => el).toBeVisible()]] = values;`},
			{Code: `const {a: {p = expect.poll(() => el).toBeVisible()}} = values;`},
			{Code: `for (const [p = expect.poll(() => el).toBeVisible()] of list) {}`},
			{Code: `function run(p = expect.poll(() => el).toBeVisible()) {}`},
			{Code: `function run([p = expect.poll(() => el).toBeVisible()]) {}`},
			{Code: `function run({p = expect.poll(() => el).toBeVisible()} = {}) {}`},
			// Assignment-pattern defaults (no declaration): an array element
			// default is an ordinary BinaryExpression already covered by the
			// KindBinaryExpression case above; an object shorthand default is
			// its own ShorthandPropertyAssignment node with no BinaryExpression
			// anywhere in it, so it needed its own case to match.
			{Code: `let p; ([p = expect.poll(() => el).toBeVisible()] = values);`},
			{Code: `let p; ({p = expect.poll(() => el).toBeVisible()} = values);`},

			// ---- C. Wrappers that do not consume the promise ----
			// tsgo keeps parentheses and TypeScript assertions as real nodes
			// where ESTree has none, so the walk out of the chain has to see
			// through them to find the `await` on the other side.
			{Code: `async function run() { await (expect.poll(() => el).toBeVisible()); }`},
			{Code: `async function run() { await ((expect.poll(() => el).toBeVisible())); }`},
			{Code: `async function run() { await (expect.poll(() => el).toBeVisible() as Promise<void>); }`},
			{Code: `async function run() { await (expect.poll(() => el).toBeVisible() satisfies Promise<void>); }`},
			{Code: `async function run() { await (expect.poll(() => el).toBeVisible()!); }`},
			{Code: `const assertion = (expect.poll(() => el).toBeVisible() as Promise<void>);`},

			// ---- D. Assertion factories that settle synchronously ----
			// Locks in upstream memberRequiresAwait(): only `poll` and
			// `element` are in awaitedMembers.
			{Code: `expect(el).toBeVisible();`},
			{Code: `expect.soft(el).toBeVisible();`},
			{Code: `expect.assertions(1);`},
			{Code: `expect.hasAssertions();`},
			{Code: `expect.unreachable();`},
			{Code: `expect.syncElement(el).toBeVisible();`},
			{Code: `expect.not.stringContaining('a');`},
			// `poll` and `element` only name a factory in first position; the
			// parser resolves these as ordinary members of an expect(x) chain.
			{Code: `expect(el).poll.toBeVisible();`},
			{Code: `expect(el).element.toBeVisible();`},
			{Code: `expect.soft.poll(() => el).toBeVisible();`},

			// ---- E. Chains with no matcher to await ----
			// The factory returns an assertion object; without a matcher call
			// nothing asynchronous has started. rstest/valid-expect owns these.
			{Code: `expect.poll(() => el);`},
			{Code: `expect.element(el);`},
			{Code: `expect.poll(() => el).toBeVisible;`},
			{Code: `expect.poll(() => el).resolves.resolves.toBeVisible();`},
			{Code: `expect.poll(() => el).unknownMember.toBeVisible();`},

			// ---- F. Reverse sources: foreign and local expects ----
			// §5 source matrix, negative rows.
			{Code: `import { expect } from 'vitest';
expect.poll(() => el).toBeVisible();`},
			{Code: `import { expect } from '@jest/globals';
expect.poll(() => el).toBeVisible();`},
			{Code: `import { expect } from '@playwright/test';
expect.poll(() => el).toBeVisible();`},
			{Code: `import { expect } from 'chai';
expect.poll(() => el).toBeVisible();`},
			{Code: `const expect = createAssertionLibrary();
expect.poll(() => el).toBeVisible();`},
			{Code: `custom.expect.poll(() => el).toBeVisible();`},
			// A same-file alias of expect is not followed by the parser.
			{Code: `import { expect } from '@rstest/core';
const check = expect;
check.poll(() => el).toBeVisible();`},
			// Shadowing.
			{Code: `import { expect } from '@rstest/core';
function run(expect: any) { expect.poll(() => el).toBeVisible(); }`},
			{Code: `import { expect as check } from '@rstest/core';
function run(check: any) { check.poll(() => el).toBeVisible(); }`},
			{Code: `import * as core from '@rstest/core';
function run() { const core = helper(); core.expect.poll(() => el).toBeVisible(); }`},

			// ---- G. Dimension 4: receiver wrappers on the expect root ----
			// A TypeScript assertion or a non-null assertion around the root
			// puts the chain outside what the shared parser resolves, so the
			// assertion is silently accepted rather than guessed at.
			{Code: `expect!.poll(() => el).toBeVisible();`},
			{Code: `(expect as any).poll(() => el).toBeVisible();`},
			{Code: `(expect satisfies any).poll(() => el).toBeVisible();`},

			// ---- G. Dimension 4: accessor forms ----
			// A computed identifier key names the factory at runtime, and a
			// numeric key names no factory at all.
			{Code: `expect[factoryName](() => el).toBeVisible();`},
			{Code: `expect[factories.poll](() => el).toBeVisible();`},
			{Code: `expect[0](() => el).toBeVisible();`},
			// N/A: private identifiers (`#poll`) cannot appear on the expect
			// object, and declaration/container forms — class vs function,
			// async vs generator, overload signatures — are unrelated to a
			// rule whose single listener runs over resolved expect chains.

			// ---- H. Comma expressions: last operand inherits the position ----
			// Locks in upstream skipSequenceExpressions() arm 1 crossed with
			// each handled position, including the ones this port adds.
			{Code: `const assertion = (sideEffect(), expect.poll(() => el).toBeVisible());`},
			{Code: `async function run() { await Promise.all([(sideEffect(), expect.poll(() => el).toBeVisible())]); }`},
			{Code: `const run = () => (sideEffect(), expect.element(el).toBeVisible());`},
			// Comma expressions are left-associative, so the assertion is the
			// right operand of the outermost one here rather than a nested one.
			{Code: `async function run() { await (sideEffect(), sideEffect(), expect.poll(() => el).toBeVisible()); }`},

			// ---- I. Real-user: vitest#496, the browser-mode request this rule answers ----
			{Code: `test('shows the dialog', async () => {
  await openDialog();

  await expect.element(page.getByRole('dialog')).toBeVisible();
});`},
			// ---- I. Real-user: assertion helpers that hand the promise back ----
			{Code: `const expectVisible = (locator: unknown) =>
  expect.element(locator).toBeVisible();

test('shows the dialog', async () => {
  await expectVisible(page.getByRole('dialog'));
});`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- J. Expect source matrix ----
			// §5, positive rows. The message always renders `expect.poll`
			// regardless of how the root is spelled at the call site.
			{
				Code: `import { expect } from '@rstest/core';
expect.poll(() => el).toBeVisible();`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "notAwaited",
					Message:   "`expect.poll` calls should be awaited",
					Line:      2,
					Column:    1,
					EndLine:   2,
					EndColumn: 12,
				}},
			},
			{
				Code: `import { expect as check } from '@rstest/core';
check.element(el).toBeVisible();`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "notAwaited",
					Message:   "`expect.element` calls should be awaited",
					Line:      2,
					Column:    1,
					EndLine:   2,
					EndColumn: 14,
				}},
			},
			{
				Code: `import * as core from '@rstest/core';
core.expect.poll(() => el).toBeVisible();`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "notAwaited",
					Message:   "`expect.poll` calls should be awaited",
					Line:      2,
					Column:    1,
					EndLine:   2,
					EndColumn: 17,
				}},
			},
			{
				Code: `const { expect } = require('@rstest/core');
expect.poll(() => el).toBeVisible();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notAwaited", Line: 2, Column: 1, EndLine: 2, EndColumn: 12}},
			},
			{
				Code: `const core = require('@rstest/core');
core.expect.poll(() => el).toBeVisible();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notAwaited", Line: 2, Column: 1, EndLine: 2, EndColumn: 17}},
			},
			{
				// The head is `import.meta.rstest.expect.poll`, but the message
				// stays on upstream's fixed `expect.` prefix.
				Code: `import.meta.rstest.expect.poll(() => el).toBeVisible();`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "notAwaited",
					Message:   "`expect.poll` calls should be awaited",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 31,
				}},
			},
			{
				Code: `const { expect } = import.meta.rstest;
expect.element(el).toBeVisible();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notAwaited", Line: 2, Column: 1, EndLine: 2, EndColumn: 15}},
			},
			{
				Code: `import { expect } from '@rstest/playwright';
expect.poll(() => el).toBeVisible();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notAwaited", Line: 2, Column: 1, EndLine: 2, EndColumn: 12}},
			},
			{
				Code:   `test('x', ({ expect }) => { expect.poll(() => el).toBeVisible(); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notAwaited", Line: 1, Column: 29, EndLine: 1, EndColumn: 40}},
			},
			{
				Code:   `test('x', ctx => { ctx.expect.element(el).toBeVisible(); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notAwaited", Line: 1, Column: 20, EndLine: 1, EndColumn: 38}},
			},

			// ---- K. Dimension 4: accessor and optional-chain forms ----
			// The reported range is the whole factory access, so a bracketed
			// accessor's closing bracket is included and the quoting style
			// does not shift the range.
			{
				Code:   "expect[`poll`](() => el).toBeVisible();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notAwaited", Line: 1, Column: 1, EndLine: 1, EndColumn: 15}},
			},
			{
				// tsgo records the optional chain as a flag on the access
				// rather than a wrapper node, so the range is unaffected.
				Code:   `expect?.poll(() => el).toBeVisible();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notAwaited", Line: 1, Column: 1, EndLine: 1, EndColumn: 13}},
			},
			{
				Code:   `expect.poll?.(() => el).toBeVisible();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notAwaited", Line: 1, Column: 1, EndLine: 1, EndColumn: 12}},
			},
			{
				Code:   `expect.poll(() => el)?.toBeVisible();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notAwaited", Line: 1, Column: 1, EndLine: 1, EndColumn: 12}},
			},
			{
				// A parenthesized root is transparent to the parser, and the
				// report still covers only the factory access.
				Code:   `(expect).poll(() => el).toBeVisible();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notAwaited", Line: 1, Column: 1, EndLine: 1, EndColumn: 14}},
			},
			{
				Code:   `((expect)).poll(() => el).toBeVisible();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notAwaited", Line: 1, Column: 1, EndLine: 1, EndColumn: 16}},
			},
			{
				// An empty argument list still starts a polling assertion.
				Code:   `expect.poll().toBeVisible();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notAwaited", Line: 1, Column: 1, EndLine: 1, EndColumn: 12}},
			},
			{
				// A member chain broken across lines reports the accessor, not
				// the statement.
				Code: `expect
  .poll(() => el)
  .toBeVisible();`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "notAwaited",
					Line:      1,
					Column:    1,
					EndLine:   2,
					EndColumn: 8,
				}},
			},

			// ---- L. Modifier and matcher chains still resolve one report ----
			{
				Code:   `expect.poll(() => el).not.toBeVisible();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notAwaited", Line: 1, Column: 1, EndLine: 1, EndColumn: 12}},
			},
			{
				// A Chai property assertion ends the chain on a member access
				// rather than a call; the parser still resolves one chain, so
				// the rule still reports exactly once.
				Code:   `expect.poll(() => el).to.be.ok;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notAwaited", Line: 1, Column: 1, EndLine: 1, EndColumn: 12}},
			},
			{
				// Two factories in one statement are two chains, each reported
				// on its own accessor.
				Code: `expect.poll(() => el).toBeVisible(); expect.element(el).toBeVisible();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notAwaited", Line: 1, Column: 1, EndLine: 1, EndColumn: 12},
					{MessageId: "notAwaited", Line: 1, Column: 38, EndLine: 1, EndColumn: 52},
				},
			},
			{
				// A polling assertion nested inside another one's callback is
				// a separate chain with its own handled position.
				Code: `async function run() { await expect.poll(() => { expect.element(el).toBeVisible(); return el; }).toBeVisible(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notAwaited", Line: 1, Column: 50, EndLine: 1, EndColumn: 64},
				},
			},

			// ---- M. Positions that drop the promise ----
			// Locks in upstream isHandled falling through both arms, and the
			// negative side of each position this port adds.
			{
				// A block-bodied arrow evaluates the statement and discards it.
				Code:   `const run = () => { expect.poll(() => el).toBeVisible(); };`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notAwaited", Line: 1, Column: 21, EndLine: 1, EndColumn: 32}},
			},
			{
				// The chain is the callee, not an argument.
				Code:   `expect.poll(() => el).toBeVisible()();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notAwaited", Line: 1, Column: 1, EndLine: 1, EndColumn: 12}},
			},
			{
				// `void` and a non-assignment operator both discard the value.
				Code:   `void expect.poll(() => el).toBeVisible();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notAwaited", Line: 1, Column: 6, EndLine: 1, EndColumn: 17}},
			},
			{
				Code:   `const ok = expect.poll(() => el).toBeVisible() && other;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notAwaited", Line: 1, Column: 12, EndLine: 1, EndColumn: 23}},
			},
			{
				// A variable declaration's *name* is not its initializer.
				Code:   `if (ready) { expect.element(el).toBeVisible(); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notAwaited", Line: 1, Column: 14, EndLine: 1, EndColumn: 28}},
			},

			// ---- N. Comma expressions: every position but the last ----
			// Locks in upstream skipSequenceExpressions() arm 2: the walk stops
			// when the assertion is not the last operand.
			{
				Code:   `async function run() { await (expect.poll(() => el).toBeVisible(), sideEffect()); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notAwaited", Line: 1, Column: 31, EndLine: 1, EndColumn: 42}},
			},
			{
				Code:   `const assertion = (expect.element(el).toBeVisible(), sideEffect());`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notAwaited", Line: 1, Column: 20, EndLine: 1, EndColumn: 34}},
			},
			{
				// Last operand of an inner comma expression, but that inner one
				// is not the last operand of the outer one.
				Code:   `async function run() { await ((sideEffect(), expect.poll(() => el).toBeVisible()), sideEffect()); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notAwaited", Line: 1, Column: 46, EndLine: 1, EndColumn: 57}},
			},
		},
	)
}
