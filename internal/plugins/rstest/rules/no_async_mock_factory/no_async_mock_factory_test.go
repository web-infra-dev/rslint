// TestNoAsyncMockFactory covers the shapes the syntax settles on its own: the
// four module mock APIs, the ways a factory is written, and the expressions
// that can only produce a promise.
//
// The shapes here were run against @rstest/core 0.11.8 by mocking a real module
// and importing it. `async () => …`, `async function () {}`,
// `Promise.resolve/all/allSettled/race/any(…)`, `new Promise(…)`, `import(…)`
// and `rs.importActual(…)` each threw
// "[Rstest] An async mock factory is not supported."; `doMock` threw it with
// the same factory, and the runtime builds all four APIs from one
// `getMockImplementation`. The generator, async generator and synchronous
// factories completed without it.
package no_async_mock_factory

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoAsyncMockFactory(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoAsyncMockFactoryRule,
		[]rule_tester.ValidTestCase{
			// ---- no factory at all ----
			{Code: `rs.mock('./sum')`},
			{Code: `rs.doMock('./sum')`},
			{Code: `rs.mockRequire('./sum')`},
			{Code: `rs.doMockRequire('./sum')`},

			// ---- the second argument is the options bag ----
			{Code: `rs.mock('./sum', { spy: true })`},
			{Code: `rstest.doMock('./sum', { spy: true })`},

			// ---- synchronous factories ----
			{Code: `rs.mock('./sum', () => ({ sum: () => 0 }))`},
			{Code: `rs.mock('./sum', function () { return { sum: 0 }; })`},
			{Code: `rs.mock('./sum', function factory() { return { sum: 0 }; })`},
			{Code: `rs.mock('./sum', () => buildMock())`},

			// A generator hands back a Generator and an async generator an
			// AsyncGenerator. Neither is `instanceof Promise`, so the runtime
			// lets both through.
			{Code: `rs.mock('./sum', function* () { yield 1; })`},
			{Code: `rs.mock('./sum', async function* () { yield 1; })`},

			// `instanceof Promise` does not recognize a hand-written thenable,
			// so the runtime does not throw on one.
			{Code: `rs.mock('./sum', () => ({ then() {} }))`},

			// `Promise.withResolvers()` returns an object holding a promise.
			{Code: `rs.mock('./sum', () => Promise.withResolvers())`},

			// One branch is synchronous, so the call may well work.
			{Code: `rs.mock('./sum', () => (cond ? Promise.resolve({}) : { sum: 0 }))`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- the four APIs share one implementation ----
			{
				Code: `rs.mock('./sum', async () => ({ sum: () => 0 }))`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Message:   "Mock factory must return the module object synchronously; 'mock()' throws when the factory returns a Promise. To keep part of the original module, import it with `with { rstest: 'importActual' }` and spread it in.",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 48,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestRemoveAsync",
						Output:    `rs.mock('./sum', () => ({ sum: () => 0 }))`,
					}},
				}},
			},
			{
				Code: `rs.doMock('./sum', async () => ({ sum: () => 0 }))`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Message:   "Mock factory must return the module object synchronously; 'doMock()' throws when the factory returns a Promise. To keep part of the original module, import it with `with { rstest: 'importActual' }` and spread it in.",
					Line:      1,
					Column:    20,
					EndLine:   1,
					EndColumn: 50,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestRemoveAsync",
						Output:    `rs.doMock('./sum', () => ({ sum: () => 0 }))`,
					}},
				}},
			},
			{
				Code: `rs.mockRequire('./sum', async () => ({ sum: () => 0 }))`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    25,
					EndLine:   1,
					EndColumn: 55,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestRemoveAsync",
						Output:    `rs.mockRequire('./sum', () => ({ sum: () => 0 }))`,
					}},
				}},
			},
			{
				Code: `rstest.doMockRequire('./sum', async () => ({ sum: () => 0 }))`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    31,
					EndLine:   1,
					EndColumn: 61,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestRemoveAsync",
						Output:    `rstest.doMockRequire('./sum', () => ({ sum: () => 0 }))`,
					}},
				}},
			},

			// ---- every way of writing an async function ----
			{
				Code: `rs.mock('./sum', async function () { return { sum: 0 }; })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 58,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestRemoveAsync",
						Output:    `rs.mock('./sum', function () { return { sum: 0 }; })`,
					}},
				}},
			},
			{
				Code: `rs.mock('./sum', async function factory() { return { sum: 0 }; })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 65,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestRemoveAsync",
						Output:    `rs.mock('./sum', function factory() { return { sum: 0 }; })`,
					}},
				}},
			},

			// The shape the migration guide warns about: the factory awaits the
			// real module. Turning that into a top-level
			// `import … with { rstest: 'importActual' }` crosses statements, so
			// the diagnostic stands without a suggestion.
			{
				Code: `
rs.mock('./sum', async () => {
  const actual = await rs.importActual('./sum');
  return { ...actual, sum: () => 0 };
});
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      2,
					Column:    18,
					EndLine:   5,
					EndColumn: 2,
				}},
			},

			// ---- the expressions that can only produce a promise ----
			{
				Code: `rs.mock('./sum', () => Promise.resolve({ sum: 0 }))`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 51,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestUnwrapPromiseResolve",
						Output:    `rs.mock('./sum', () => ({ sum: 0 }))`,
					}},
				}},
			},
			{
				Code: `rs.mock('./sum', function () { return Promise.resolve({ sum: 0 }); })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 69,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestUnwrapPromiseResolve",
						Output:    `rs.mock('./sum', function () { return { sum: 0 }; })`,
					}},
				}},
			},
			{
				Code: `rs.mock('./sum', () => Promise.reject(new Error('boom')))`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 57,
				}},
			},
			{
				Code: `rs.mock('./sum', () => Promise.all([]))`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 39,
				}},
			},
			{
				Code: `rs.mock('./sum', () => Promise.allSettled([]))`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 46,
				}},
			},
			{
				Code: `rs.mock('./sum', () => Promise.race([]))`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 40,
				}},
			},
			{
				Code: `rs.mock('./sum', () => Promise.any([]))`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 39,
				}},
			},
			{
				Code: `rs.mock('./sum', () => new Promise((resolve) => resolve({ sum: 0 })))`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 69,
				}},
			},
			{
				Code: `rs.mock('./sum', () => import('./sum'))`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 39,
				}},
			},
			{
				Code: `rs.mock('./sum', () => rs.importActual('./sum'))`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 48,
				}},
			},
			{
				Code: `rs.mock('./sum', () => rs.importMock('./sum'))`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 46,
				}},
			},

			// Every branch of the body hands back a promise.
			{
				Code: `
rs.mock('./sum', function () {
  if (flag) {
    return Promise.resolve({ sum: 0 });
  }
  return import('./sum');
});
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      2,
					Column:    18,
					EndLine:   7,
					EndColumn: 2,
				}},
			},

			// ---- a declaration in this file settles the identifier ----
			{
				Code: `
const factory = async () => ({ sum: () => 0 });
rs.mock('./sum', factory);
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      3,
					Column:    18,
					EndLine:   3,
					EndColumn: 25,
				}},
			},
			{
				Code: `
async function factory() {
  return { sum: () => 0 };
}
rs.mock('./sum', factory);
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      5,
					Column:    18,
					EndLine:   5,
					EndColumn: 25,
				}},
			},
			{
				Code: `
const factory = () => Promise.resolve({ sum: () => 0 });
rs.mock('./sum', factory);
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      3,
					Column:    18,
					EndLine:   3,
					EndColumn: 25,
				}},
			},
		},
	)
}
