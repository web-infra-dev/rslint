// TestNoAsyncMockFactoryExtrasReuse covers what the rule reuses across the calls of
// one file: the shadowing answer for `Promise` and `undefined`, the verdict of
// a function node, the verdict a symbol's declaration gives, and whether a
// symbol is written to. Each case puts several mock calls in one file, so an
// answer reused where it should not have been shows up as a report on the wrong
// call or as a missing one.
//
// The file-level fast-fail is covered here too: every case that reports has to
// keep reporting now that the listener is registered only for files naming both
// a receiver and a mock member.
package no_async_mock_factory

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoAsyncMockFactoryExtrasReuse(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoAsyncMockFactoryRule,
		[]rule_tester.ValidTestCase{
			// Two factories that each declare their own `Promise`. The syntax
			// layer declines both and the local object's `resolve` gives the
			// type layer nothing, so neither is reported — and the two scopes
			// have to be asked separately to get there.
			{Code: `
rs.mock('./a', () => {
  const Promise = { resolve: (value: unknown) => value };
  return Promise.resolve({ a: 0 });
});
rs.mock('./b', () => {
  const Promise = { resolve: (value: unknown) => value };
  return Promise.resolve({ b: 0 });
});
`},

			// One synchronous factory installed by several calls: classifying
			// it once must not turn into a report on any of them.
			{Code: `
const factory = () => ({ sum: 0 });
rs.mock('./a', factory);
rs.doMock('./b', factory);
rs.mockRequire('./c', factory);
`},

			// A `Promise` shadowed by a parameter rather than by a local, so
			// the answer depends on the function scope the call site sits in.
			{Code: `
function build(Promise: { resolve: (value: unknown) => unknown }) {
  rs.mock('./a', () => Promise.resolve({ a: 0 }));
}
`},

			// The names the file-level fast-fail looks for. A file naming a
			// mock member but no receiver, or a receiver but no member, cannot
			// hold a call this rule reports.
			{Code: `
const mock = async () => ({ sum: 0 });
const doMock = mock;
export { mock, doMock };
`},
			{Code: `
import { rs } from '@rstest/core';
rs.fn(async () => ({ sum: 0 }));
`},
		},
		[]rule_tester.InvalidTestCase{
			// One shared factory installed by several calls. Each call is its
			// own report, so the declaration verdict is reused without the
			// reports collapsing into one.
			{
				Code: `
const factory = () => Promise.resolve({ sum: 0 });
rs.mock('./a', factory);
rs.doMock('./b', factory);
rs.mockRequire('./c', factory);
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMockFactory", Line: 3, Column: 16, EndLine: 3, EndColumn: 23},
					{MessageId: "asyncMockFactory", Line: 4, Column: 18, EndLine: 4, EndColumn: 25},
					{MessageId: "asyncMockFactory", Line: 5, Column: 23, EndLine: 5, EndColumn: 30},
				},
			},

			// `Promise` shadowed inside the first factory and global in the
			// second. An answer cached per file rather than per scope would
			// report both calls or neither.
			{
				Code: `
rs.mock('./a', () => {
  const Promise = { resolve: (value: unknown) => value };
  return Promise.resolve({ a: 0 });
});
rs.mock('./b', () => Promise.resolve({ b: 0 }));
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      6,
					Column:    16,
					EndLine:   6,
					EndColumn: 47,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestUnwrapPromiseResolve",
						Output: `
rs.mock('./a', () => {
  const Promise = { resolve: (value: unknown) => value };
  return Promise.resolve({ a: 0 });
});
rs.mock('./b', () => ({ b: 0 }));
`,
					}},
				}},
			},

			// The same file the other way round, so the cache is filled by the
			// unshadowed call first.
			{
				Code: `
rs.mock('./a', () => Promise.resolve({ a: 0 }));
rs.mock('./b', () => {
  const Promise = { resolve: (value: unknown) => value };
  return Promise.resolve({ b: 0 });
});
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      2,
					Column:    16,
					EndLine:   2,
					EndColumn: 47,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestUnwrapPromiseResolve",
						Output: `
rs.mock('./a', () => ({ a: 0 }));
rs.mock('./b', () => {
  const Promise = { resolve: (value: unknown) => value };
  return Promise.resolve({ b: 0 });
});
`,
					}},
				}},
			},

			// Two `Promise.resolve` factories in scopes that answer alike, so
			// the cached answer really is shared between them.
			{
				Code: `
rs.mock('./a', () => Promise.resolve({ a: 0 }));
rs.mock('./b', () => Promise.resolve({ b: 0 }));
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "asyncMockFactory",
						Line:      2,
						Column:    16,
						EndLine:   2,
						EndColumn: 47,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
							MessageId: "suggestUnwrapPromiseResolve",
							Output: `
rs.mock('./a', () => ({ a: 0 }));
rs.mock('./b', () => Promise.resolve({ b: 0 }));
`,
						}},
					},
					{
						MessageId: "asyncMockFactory",
						Line:      3,
						Column:    16,
						EndLine:   3,
						EndColumn: 47,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
							MessageId: "suggestUnwrapPromiseResolve",
							Output: `
rs.mock('./a', () => Promise.resolve({ a: 0 }));
rs.mock('./b', () => ({ b: 0 }));
`,
						}},
					},
				},
			},

			// `undefined` reaches the same cache, from the suggestion layer
			// rather than the classification layers. It is shadowed in the
			// first factory, so only the second may drop its `async`.
			{
				Code: `
rs.mock('./a', async () => {
  const undefined = 1;
  return undefined;
});
rs.mock('./b', async () => undefined);
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMockFactory", Line: 2, Column: 16, EndLine: 5, EndColumn: 2},
					{
						MessageId: "asyncMockFactory",
						Line:      6,
						Column:    16,
						EndLine:   6,
						EndColumn: 37,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
							MessageId: "suggestRemoveAsync",
							Output: `
rs.mock('./a', async () => {
  const undefined = 1;
  return undefined;
});
rs.mock('./b', () => undefined);
`,
						}},
					},
				},
			},

			// Two factories written identically must not share a verdict by
			// shape: the cache is keyed by node, and only the second one is
			// reportable.
			{
				Code: `
const sync = () => ({ sum: 0 });
const asyncFactory = () => Promise.resolve({ sum: 0 });
rs.mock('./a', sync);
rs.mock('./b', asyncFactory);
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      5,
					Column:    16,
					EndLine:   5,
					EndColumn: 28,
				}},
			},
		},
	)
}
