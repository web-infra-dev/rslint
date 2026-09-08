// TestNoAsyncMockFactoryExtras covers what the syntax layer cannot settle on
// the factory alone: which call shapes the Rstest build actually rewrites,
// where the call has to stand, what the argument list has to look like, and the
// boundaries of the two suggestions.
//
// `rs['mock'](…)`, `rs.mock?.(…)` and `rs?.mock(…)` were each run against
// @rstest/core 0.11.8 with an async factory: all three threw
// "mock() was not transformed by Rstest" rather than the async-factory error,
// so the factory never runs and there is nothing to report.
package no_async_mock_factory

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoAsyncMockFactoryExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoAsyncMockFactoryRule,
		[]rule_tester.ValidTestCase{
			// ---- shapes the build leaves as the throwing stub ----
			// None of these four members is matched through a computed
			// property, an optional call or an optional receiver, so the
			// factory never runs and there is nothing to report.
			{Code: `rs['mock']('./sum', async () => ({ sum: 0 }))`},
			{Code: `rs.mock?.('./sum', async () => ({ sum: 0 }))`},
			{Code: `rs?.mock('./sum', async () => ({ sum: 0 }))`},
			{Code: "rs[`mock`]('./sum', async () => ({ sum: 0 }))"},

			// The rewrite matches the receiver by the name written at the call
			// site, so a renamed binding is an ordinary call that throws where
			// it stands.
			{Code: `import { rs as mocker } from '@rstest/core';
mocker.mock('./sum', async () => ({ sum: 0 }))`},
			{Code: `import * as core from '@rstest/core';
core.rs.mock('./sum', async () => ({ sum: 0 }))`},

			// ---- positions the call cannot be lifted out of ----
			{Code: `const mocked = rs.mock('./sum', async () => ({ sum: 0 }))`},
			{Code: `await rs.mock('./sum', async () => ({ sum: 0 }))`},
			{Code: `consume(rs.mock('./sum', async () => ({ sum: 0 })))`},
			{Code: `void rs.mock('./sum', async () => ({ sum: 0 }))`},
			{Code: `rs.mock('./sum', async () => ({ sum: 0 })), other()`},

			// ---- argument lists the transform gives up on ----
			{Code: `rs.mock(...args)`},
			{Code: `rs.mock('./sum', ...factories)`},
			{Code: `rs.mock('./sum', async () => ({ sum: 0 }), extra)`},

			// ---- the local declaration decides where the build asks for one ----
			// `importActual` is one of the two members whose rewrite honours an
			// ordinary binding, so this file's own `rs` really runs and the
			// factory is synchronous.
			{Code: `const rs = { mock: (p: string, f: () => unknown) => f(), importActual: () => ({ sum: 0 }) };
rs.mock('./sum', () => rs.importActual('./sum'))`},
			// A file that declares its own `Promise` is left to the types,
			// which see a plain object here.
			{Code: `export {};
const Promise = { resolve: (value: unknown) => value };
rs.mock('./sum', () => Promise.resolve({ sum: 0 }))`},

			// ---- the types settle what the syntax could not ----
			{Code: `declare const factory: () => { sum: number };
rs.mock('./sum', factory)`},
			// A union that mixes an async factory with a synchronous one may
			// well work at run time.
			{Code: `declare const factory: (() => Promise<{ sum: number }>) | (() => { sum: number });
rs.mock('./sum', factory)`},
			{Code: `declare const factory: any;
rs.mock('./sum', factory)`},
			// A module that exports `then` gives the factory a thenable return
			// type, but `instanceof Promise` does not match it and the runtime
			// does not throw.
			{Code: `declare const factory: () => { then(onFulfilled: (value: unknown) => void): void };
rs.mock('./sum', factory)`},
			{Code: `import { syncModuleFactory } from './async-mock-factories';
rs.mock('./sum', syncModuleFactory)`},
			// A local class named `Promise` is not the one the runtime checks
			// against: `new Promise()` hands back an ordinary object here.
			{Code: `export {};
class Promise { sum = 0; }
rs.doMockRequire('./sum', () => new Promise())`},

			// ---- a path that returns nothing ----
			// A false `flag` reaches the end of the body and returns
			// `undefined`, which is not a promise, so the factory only
			// sometimes hands one back.
			{Code: `rs.doMockRequire('./sum', () => { if (flag) return Promise.resolve({ sum: 0 }); })`},
			{Code: `rs.mock('./sum', function () {
  if (flag) {
    return Promise.resolve({ sum: 0 });
  }
})`},
			// The declaration is not the factory that is installed: the name is
			// written to before the call, and the type layer sees only
			// `() => unknown`.
			{Code: `function factory(): unknown {
  return Promise.resolve({ sum: 0 });
}
factory = () => ({ sum: 0 });
rs.doMock('./sum', factory)`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- shapes the build does rewrite ----
			// `as`, `satisfies` and `!` are erased before the rewrite runs, and
			// it reads through parentheses.
			{
				Code: `(rs as any)!.mock('./sum', async () => ({ sum: 0 }))`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    28,
					EndLine:   1,
					EndColumn: 52,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestRemoveAsync",
						Output:    `(rs as any)!.mock('./sum', () => ({ sum: 0 }))`,
					}},
				}},
			},
			// The report covers the argument as written, so an assertion
			// wrapped around the factory is part of the range.
			{
				Code: `rs.mock('./sum', (async () => ({ sum: 0 })) as never)`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 53,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestRemoveAsync",
						Output:    `rs.mock('./sum', (() => ({ sum: 0 })) as never)`,
					}},
				}},
			},

			// A block is still a statement position; the build lifts the call
			// out of it.
			{
				Code: `
if (flag) {
  rs.mock('./sum', async () => ({ sum: 0 }));
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      3,
					Column:    20,
					EndLine:   3,
					EndColumn: 44,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestRemoveAsync",
						Output: `
if (flag) {
  rs.mock('./sum', () => ({ sum: 0 }));
}
`,
					}},
				}},
			},

			// ---- what the first argument is written as does not matter ----
			{
				Code: `rs.mock(import('./sum'), async () => ({ sum: 0 }))`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    26,
					EndLine:   1,
					EndColumn: 50,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestRemoveAsync",
						Output:    `rs.mock(import('./sum'), () => ({ sum: 0 }))`,
					}},
				}},
			},
			{
				Code: `rs.mock(modulePath, async () => ({ sum: 0 }))`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    21,
					EndLine:   1,
					EndColumn: 45,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestRemoveAsync",
						Output:    `rs.mock(modulePath, () => ({ sum: 0 }))`,
					}},
				}},
			},
			// An explicit type argument names the mocked module's shape. It
			// says nothing about when the factory hands it back.
			{
				Code: `rs.mock<{ sum: number }>('./sum', async () => ({ sum: 0 }))`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    35,
					EndLine:   1,
					EndColumn: 59,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestRemoveAsync",
						Output:    `rs.mock<{ sum: number }>('./sum', () => ({ sum: 0 }))`,
					}},
				}},
			},

			// ---- the types settle what the syntax could not ----
			{
				Code: `declare const factory: () => Promise<{ sum: number }>;
rs.mock('./sum', factory)`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      2,
					Column:    18,
					EndLine:   2,
					EndColumn: 25,
				}},
			},
			// Inline, and still undecidable from the syntax alone.
			{
				Code: `declare function buildMock(): Promise<{ sum: number }>;
rs.mock('./sum', () => buildMock())`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      2,
					Column:    18,
					EndLine:   2,
					EndColumn: 35,
				}},
			},
			{
				Code: `import { asyncModuleFactory } from './async-mock-factories';
rs.mock('./sum', asyncModuleFactory)`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      2,
					Column:    18,
					EndLine:   2,
					EndColumn: 36,
				}},
			},
			// A subclass of Promise is still `instanceof Promise`.
			{
				Code: `declare class Deferred<T> extends Promise<T> {}
declare const factory: () => Deferred<{ sum: number }>;
rs.mock('./sum', factory)`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      3,
					Column:    18,
					EndLine:   3,
					EndColumn: 25,
				}},
			},

			// ---- suggestion boundaries ----
			// Removing `async` would not make the body synchronous.
			{
				Code: `declare function buildMock(): Promise<{ sum: number }>;
rs.mock('./sum', async () => buildMock())`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      2,
					Column:    18,
					EndLine:   2,
					EndColumn: 41,
				}},
			},
			// A bare `return` leaves the returned set incomplete.
			{
				Code: `rs.mock('./sum', async function () { if (flag) { return; } return { sum: 0 }; })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 80,
				}},
			},
			// The body suspends, so dropping `async` would not compile.
			{
				Code: `rs.mock('./sum', async () => { await ready; return { sum: 0 }; })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 65,
				}},
			},
			{
				Code: `rs.mock('./sum', async () => { for await (const part of parts) { use(part); } return { sum: 0 }; })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 99,
				}},
			},
			{
				Code: `rs.mock('./sum', async () => { await using handle = open(); return { sum: 0 }; })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 81,
				}},
			},
			// A nested function's `await` belongs to that function.
			{
				Code: `rs.mock('./sum', async () => ({ load: async () => { await ready; } }))`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 70,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestRemoveAsync",
						Output:    `rs.mock('./sum', () => ({ load: async () => { await ready; } }))`,
					}},
				}},
			},
			// Unwrapping would leave a promise behind.
			{
				Code: `rs.mock('./sum', () => Promise.resolve(buildMock()))`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 52,
				}},
			},
			// A comment inside the replaced span would be deleted with it.
			{
				Code: `rs.mock('./sum', () => Promise.resolve(/* keep */ { sum: 0 }))`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 62,
				}},
			},
			// Removing `async` would leave the annotation promising a promise
			// the factory no longer returns, and unwrapping `Promise.resolve`
			// has the same problem, so neither is suggested.
			{
				Code: `rs.mock('./sum', async (): Promise<{ sum: number }> => ({ sum: 0 }))`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 68,
				}},
			},
			{
				Code: `rs.mock('./sum', (): Promise<{ sum: number }> => Promise.resolve({ sum: 0 }))`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 77,
				}},
			},
			// A nested function's computed name is evaluated in this body, so
			// the `await` in it is the factory's own and dropping `async` would
			// not parse.
			{
				Code: `rs.mock('./sum', async () => ({ [await ready]() { return 0; } }))`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 65,
				}},
			},
			// A `return` statement takes the object without parentheses; a
			// concise arrow body needs them.
			{
				Code: `rs.mock('./sum', () => Promise.resolve([1, 2]))`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncMockFactory",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 47,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestUnwrapPromiseResolve",
						Output:    `rs.mock('./sum', () => [1, 2])`,
					}},
				}},
			},
		},
	)
}
