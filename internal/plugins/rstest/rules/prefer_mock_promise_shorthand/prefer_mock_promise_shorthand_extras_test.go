// TestPreferMockPromiseShorthandExtras covers what the upstream suites cannot:
// Rstest mock receivers, tsgo node shapes that ESTree does not distinguish, the
// branches of the shared engine that upstream never exercises, and the places
// this rule deliberately decides differently from both reference plugins.
package prefer_mock_promise_shorthand

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferMockPromiseShorthandExtras(t *testing.T) {
	report := func(replacement string, line, column int) []rule_tester.InvalidTestCaseError {
		return []rule_tester.InvalidTestCaseError{{
			MessageId: "useMockShorthand", Message: "Prefer " + replacement,
			Line: line, Column: column,
		}}
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferMockPromiseShorthandRule,
		[]rule_tester.ValidTestCase{
			// --- the method has to be one this rule rewrites ---
			{Code: `mockReturnValue(Promise.resolve(1));`},
			{Code: `aVariable.mockReturnValueTwice(Promise.resolve(1));`},
			{Code: `aVariable.mockImplementationOnceMore(() => Promise.resolve(1));`},
			{Code: `new aVariable.mockReturnValue(Promise.resolve(1));`},
			// A computed identifier key names a variable, so the call site does
			// not say which method is reached.
			{Code: `const name = 'mockReturnValue';
aVariable[name](Promise.resolve(1));`},

			// --- the value has to be a promise built by the global Promise ---
			{Code: `aVariable.mockReturnValue(Promise.all([1]));`},
			{Code: `aVariable.mockReturnValue(Promise.resolve);`},
			{Code: `aVariable.mockReturnValue(thingy.resolve(1));`},
			{Code: `aVariable.mockReturnValue(globalThis.Promise.resolve(1));`},
			{Code: `const method = 'resolve';
aVariable.mockReturnValue(Promise[method](1));`},
			// A locally declared `Promise` is some other object.
			{Code: `const Promise = FakePromise;
aVariable.mockReturnValue(Promise.resolve(1));`},
			{Code: `class Promise {}
aVariable.mockImplementation(() => Promise.reject(1));`},

			// --- the callback has to be collapsible ---
			{Code: `aVariable.mockImplementation(function (this: Thingy) { return Promise.resolve(1); });`},
			{Code: `aVariable.mockImplementation(<T,>() => Promise.resolve(null as T));`},
			{Code: `aVariable.mockImplementation(aFunctionDefinedElsewhere);`},
			// Both reference plugins rewrite this to `mockResolvedValue(1)`, but a
			// generator returns an iterator, not the promise its `return` names.
			{Code: `aVariable.mockImplementation(function* () { return Promise.resolve(1); });`},

			// --- the value would move from every call to the one configuration ---
			// Both reference plugins rewrite every shape below. The mock then
			// settles with the value the expression had when it was configured,
			// or runs its write once instead of once per call.
			{Code: `let value = 1;
aVariable.mockImplementation(() => Promise.resolve(value));`},
			{Code: `var value = 1;
aVariable.mockImplementationOnce(() => Promise.resolve(value));`},
			{Code: `let attempt = 0;
aVariable.mockImplementation(() => Promise.reject(new Error(` + "`attempt ${attempt}`" + `)));`},
			{Code: `let value = 1;
aVariable.mockImplementation(() => Promise.resolve(() => value));`},
			{Code: `let count = 0;
aVariable.mockImplementation(() => Promise.resolve(count++));`},
			{Code: `const state = { count: 0 };
aVariable.mockImplementation(() => Promise.resolve(state.count += 1));`},
			{Code: `const state = { count: 0 };
aVariable.mockImplementation(() => { return Promise.resolve(delete state.count); });`},
			// A spread iterates its operand on every call. An iterator the first
			// call exhausts hands the next one `undefined`.
			{Code: `const values = makeIterator();
aVariable.mockImplementation(() => Promise.resolve(...values));`},

			// --- a `function` callback's `this` and `arguments` come from the call ---
			{Code: `aVariable.mockImplementation(function () { return Promise.resolve(this.id); });`},
			{Code: `aVariable.mockImplementation(function () { return Promise.resolve(arguments[0]); });`},
			{Code: `aVariable.mockImplementation(function () { return Promise.resolve(new.target); });`},
			{Code: `aVariable.mockImplementation(async function () { return Promise.resolve(this.id); });`},

			// --- an `await` belongs to the async callback it is written in ---
			// Both reference plugins lift it out, which fails to parse inside a
			// function or turns into a top-level await that runs once.
			{Code: `aVariable.mockImplementation(async () => Promise.resolve(await load()));`},
			{Code: `aVariable.mockImplementation(async () => { return Promise.reject(await load()); });`},
		},
		[]rule_tester.InvalidTestCase{
			// --- Rstest mock receivers ---
			{
				Code:   `rs.fn().mockReturnValue(Promise.resolve(1));`,
				Output: []string{`rs.fn().mockResolvedValue(1);`},
				Errors: report("mockResolvedValue", 1, 9),
			},
			{
				Code:   `rstest.fn().mockImplementationOnce(() => Promise.reject(error));`,
				Output: []string{`rstest.fn().mockRejectedValueOnce(error);`},
				Errors: report("mockRejectedValueOnce", 1, 13),
			},
			{
				Code:   `import.meta.rstest.fn().mockReturnValueOnce(Promise.resolve(1));`,
				Output: []string{`import.meta.rstest.fn().mockResolvedValueOnce(1);`},
				Errors: report("mockResolvedValueOnce", 1, 25),
			},
			{
				Code: `import { fn as makeMock } from '@rstest/core';
makeMock().mockImplementation(() => Promise.resolve(1));`,
				Output: []string{`import { fn as makeMock } from '@rstest/core';
makeMock().mockResolvedValue(1);`},
				Errors: report("mockResolvedValue", 2, 12),
			},
			{
				Code:   `rs.spyOn(fs.promises, 'readFile').mockReturnValue(Promise.reject(new Error('nope')));`,
				Output: []string{`rs.spyOn(fs.promises, 'readFile').mockRejectedValue(new Error('nope'));`},
				Errors: report("mockRejectedValue", 1, 35),
			},
			{
				Code:   `rs.mocked(thingy.load).mockImplementation(() => Promise.resolve(1));`,
				Output: []string{`rs.mocked(thingy.load).mockResolvedValue(1);`},
				Errors: report("mockResolvedValue", 1, 24),
			},

			// --- accessor spellings ---
			// Both reference plugins replace the whole string literal here,
			// producing `aVariable[mockResolvedValue](1)` — a reference to an
			// undeclared variable. Only the text between the delimiters is
			// rewritten.
			{
				Code:   `aVariable['mockReturnValue'](Promise.resolve(1));`,
				Output: []string{`aVariable['mockResolvedValue'](1);`},
				Errors: report("mockResolvedValue", 1, 11),
			},
			{
				Code:   "aVariable[`mockImplementationOnce`](() => Promise.reject(1));",
				Output: []string{"aVariable[`mockRejectedValueOnce`](1);"},
				Errors: report("mockRejectedValueOnce", 1, 11),
			},
			{
				Code:   `aVariable?.mockReturnValue(Promise.resolve(1));`,
				Output: []string{`aVariable?.mockResolvedValue(1);`},
				Errors: report("mockResolvedValue", 1, 12),
			},
			{
				Code:   `(aVariable.mockImplementation)(() => Promise.resolve(1));`,
				Output: []string{`(aVariable.mockResolvedValue)(1);`},
				Errors: report("mockResolvedValue", 1, 12),
			},

			// --- Promise spellings ---
			{
				Code:   `aVariable.mockReturnValue(Promise['resolve'](1));`,
				Output: []string{`aVariable.mockResolvedValue(1);`},
				Errors: report("mockResolvedValue", 1, 11),
			},
			{
				Code:   "aVariable.mockImplementation(() => Promise[`reject`](1));",
				Output: []string{`aVariable.mockRejectedValue(1);`},
				Errors: report("mockRejectedValue", 1, 11),
			},
			{
				Code:   `aVariable.mockReturnValue(Promise?.resolve(1));`,
				Output: []string{`aVariable.mockResolvedValue(1);`},
				Errors: report("mockResolvedValue", 1, 11),
			},
			{
				Code:   `aVariable.mockReturnValue((Promise as any).resolve(1));`,
				Output: []string{`aVariable.mockResolvedValue(1);`},
				Errors: report("mockResolvedValue", 1, 11),
			},
			{
				Code:   `aVariable.mockReturnValue(Promise.resolve<number>(1));`,
				Output: []string{`aVariable.mockResolvedValue(1);`},
				Errors: report("mockResolvedValue", 1, 11),
			},
			{
				Code:   `aVariable.mockImplementation(() => (Promise.reject<never>)(error));`,
				Output: []string{`aVariable.mockRejectedValue(error);`},
				Errors: report("mockRejectedValue", 1, 11),
			},

			// --- a type assertion around the promise ---
			// A rejected promise's error is untyped, so the assertion carries
			// nothing the shorthand needs.
			{
				Code:   `aVariable.mockReturnValue(Promise.reject(error) as Promise<never>);`,
				Output: []string{`aVariable.mockRejectedValue(error);`},
				Errors: report("mockRejectedValue", 1, 11),
			},
			{
				Code:   `aVariable.mockImplementation(() => Promise.reject(error)!);`,
				Output: []string{`aVariable.mockRejectedValue(error);`},
				Errors: report("mockRejectedValue", 1, 11),
			},
			{
				Code:   `aVariable.mockImplementation(() => <Promise<never>>Promise.reject(error));`,
				Output: []string{`aVariable.mockRejectedValue(error);`},
				Errors: report("mockRejectedValue", 1, 11),
			},
			// A resolved promise's assertion decides the type the value is
			// checked against, so the fix is withheld rather than dropping it.
			{
				Code:   `aVariable.mockReturnValue(Promise.resolve(value) as Promise<Thing>);`,
				Output: []string{},
				Errors: report("mockResolvedValue", 1, 11),
			},
			{
				Code:   `aVariable.mockImplementation(() => Promise.resolve(value) satisfies Promise<Thing>);`,
				Output: []string{},
				Errors: report("mockResolvedValue", 1, 11),
			},

			// --- tsgo node shapes around the value ---
			{
				Code:   `aVariable.mockReturnValue((Promise.resolve(1)));`,
				Output: []string{`aVariable.mockResolvedValue((1));`},
				Errors: report("mockResolvedValue", 1, 11),
			},
			{
				Code:   `aVariable.mockImplementation((() => Promise.resolve(1)));`,
				Output: []string{`aVariable.mockResolvedValue((1));`},
				Errors: report("mockResolvedValue", 1, 11),
			},
			{
				Code:   `aVariable.mockImplementation(() => (Promise.resolve(1)));`,
				Output: []string{`aVariable.mockResolvedValue(1);`},
				Errors: report("mockResolvedValue", 1, 11),
			},
			{
				Code:   `aVariable.mockImplementation((): Promise<number> => Promise.resolve(1));`,
				Output: []string{`aVariable.mockResolvedValue(1);`},
				Errors: report("mockResolvedValue", 1, 11),
			},
			{
				Code:   `aVariable.mockReturnValue(Promise.resolve(({ a: 1 })));`,
				Output: []string{`aVariable.mockResolvedValue({ a: 1 });`},
				Errors: report("mockResolvedValue", 1, 11),
			},
			// A comma expression keeps the parentheses the rewrite would
			// otherwise drop. Both reference plugins emit
			// `mockResolvedValue(0, 1)` here, which resolves to 0 rather than 1.
			{
				Code:   `aVariable.mockReturnValue(Promise.resolve((0, 1)));`,
				Output: []string{`aVariable.mockResolvedValue((0, 1));`},
				Errors: report("mockResolvedValue", 1, 11),
			},
			// `(...values)` is not an expression, so the parentheses around a
			// spread go with it.
			{
				Code:   `aVariable.mockReturnValue((Promise.resolve(...values)));`,
				Output: []string{`aVariable.mockResolvedValue(...values);`},
				Errors: report("mockResolvedValue", 1, 11),
			},
			{
				Code:   `aVariable.mockReturnValue((/* why */ Promise.resolve(...values)));`,
				Output: []string{},
				Errors: report("mockResolvedValue", 1, 11),
			},
			{
				Code:   `aVariable.mockImplementation(() => Promise.reject());`,
				Output: []string{`aVariable.mockRejectedValue(undefined);`},
				Errors: report("mockRejectedValue", 1, 11),
			},
			// A local binding named `undefined` would be read by the identifier,
			// so `void 0` stands in for it.
			{
				Code: `function setup(undefined: number) {
  return rs.fn().mockReturnValue(Promise.resolve());
}`,
				Output: []string{`function setup(undefined: number) {
  return rs.fn().mockResolvedValue(void 0);
}`},
				Errors: report("mockResolvedValue", 2, 18),
			},
			{
				Code: `function setup() {
  const undefined = 'set';
  return rs.fn().mockImplementation(() => Promise.reject());
}`,
				Output: []string{`function setup() {
  const undefined = 'set';
  return rs.fn().mockRejectedValue(void 0);
}`},
				Errors: report("mockRejectedValue", 3, 18),
			},
			// A binding in a scope the call is not in does not shadow it.
			{
				Code: `function other(undefined: number) {}
aVariable.mockReturnValue(Promise.resolve());`,
				Output: []string{`function other(undefined: number) {}
aVariable.mockResolvedValue(undefined);`},
				Errors: report("mockResolvedValue", 2, 11),
			},
			{
				Code:   `aVariable.mockImplementation(function () { return Promise.resolve(1); });`,
				Output: []string{`aVariable.mockResolvedValue(1);`},
				Errors: report("mockResolvedValue", 1, 11),
			},

			// --- values that stay fixable ---
			// `mockReturnValue` already evaluates its argument once, so the
			// binding it reads is frozen before and after the rewrite alike.
			{
				Code: `let value = 1;
aVariable.mockReturnValue(Promise.resolve(value));`,
				Output: []string{`let value = 1;
aVariable.mockResolvedValue(value);`},
				Errors: report("mockResolvedValue", 2, 11),
			},
			// For the same reason a spread is iterated once either way.
			{
				Code:   `aVariable.mockReturnValue(Promise.resolve(...values));`,
				Output: []string{`aVariable.mockResolvedValue(...values);`},
				Errors: report("mockResolvedValue", 1, 11),
			},
			{
				Code: `const value = 1;
aVariable.mockImplementation(() => Promise.resolve(value));`,
				Output: []string{`const value = 1;
aVariable.mockResolvedValue(value);`},
				Errors: report("mockResolvedValue", 2, 11),
			},
			// An arrow has no `this` of its own, so it already names the
			// enclosing scope.
			{
				Code:   `aVariable.mockImplementation(() => Promise.resolve(this.id));`,
				Output: []string{`aVariable.mockResolvedValue(this.id);`},
				Errors: report("mockResolvedValue", 1, 11),
			},
			// An async function returning a promise settles the way that promise
			// does.
			{
				Code:   `aVariable.mockImplementation(async () => Promise.resolve(1));`,
				Output: []string{`aVariable.mockResolvedValue(1);`},
				Errors: report("mockResolvedValue", 1, 11),
			},
			{
				Code:   `aVariable.mockImplementationOnce(async function () { return Promise.reject(error); });`,
				Output: []string{`aVariable.mockRejectedValueOnce(error);`},
				Errors: report("mockRejectedValueOnce", 1, 11),
			},
			// An `await` inside a function the value hands back runs later, in
			// that function.
			{
				Code:   `aVariable.mockImplementation(async () => Promise.resolve(async () => await load()));`,
				Output: []string{`aVariable.mockResolvedValue(async () => await load());`},
				Errors: report("mockResolvedValue", 1, 11),
			},

			// --- type arguments on the mock call withhold the fix ---
			// They describe the value the old method took, not the settled value
			// the shorthand takes.
			{
				Code:   `aVariable.mockReturnValue<Promise<number>>(Promise.resolve(1));`,
				Output: []string{},
				Errors: report("mockResolvedValue", 1, 11),
			},
			{
				Code:   `aVariable.mockImplementation<() => Promise<number>>(() => Promise.resolve(1));`,
				Output: []string{},
				Errors: report("mockResolvedValue", 1, 11),
			},

			// --- a comment the rewrite would delete withholds the fix ---
			{
				Code:   `aVariable.mockReturnValue(Promise.resolve(/* why */ 1));`,
				Output: []string{},
				Errors: report("mockResolvedValue", 1, 11),
			},
			{
				Code:   `aVariable.mockReturnValue(Promise.resolve(/* nothing */));`,
				Output: []string{},
				Errors: report("mockResolvedValue", 1, 11),
			},
			{
				Code:   `aVariable.mockImplementation(() => /* why */ Promise.reject(1));`,
				Output: []string{},
				Errors: report("mockRejectedValue", 1, 11),
			},
			{
				Code: `aVariable.mockImplementation(() => {
  // why
  return Promise.resolve(1);
});`,
				Output: []string{},
				Errors: report("mockResolvedValue", 1, 11),
			},
			// A comment outside the replaced argument is not touched.
			{
				Code:   `aVariable.mockReturnValue(/* why */ Promise.resolve(1));`,
				Output: []string{`aVariable.mockResolvedValue(/* why */ 1);`},
				Errors: report("mockResolvedValue", 1, 11),
			},

			// --- a resolved value that may be a promise withholds the fix ---
			// The shorthand adopts it at run time just as `Promise.resolve` does,
			// but its parameter is typed as the settled value, so
			// `mockResolvedValue(p)` would stop type-checking.
			{
				Code: `const p = Promise.resolve(1);
const mock = rs.fn<() => Promise<number>>();
mock.mockReturnValue(Promise.resolve(p));`,
				Output: []string{},
				Errors: report("mockResolvedValue", 3, 6),
			},
			{
				Code: `const p = Promise.resolve(1);
aVariable.mockImplementation(() => Promise.resolve(p));`,
				Output: []string{},
				Errors: report("mockResolvedValue", 2, 11),
			},
			{
				Code: `declare const value: number | Promise<number>;
aVariable.mockReturnValue(Promise.resolve(value));`,
				Output: []string{},
				Errors: report("mockResolvedValue", 2, 11),
			},
			{
				Code: `declare const value: PromiseLike<number>;
aVariable.mockReturnValue(Promise.resolve(value));`,
				Output: []string{},
				Errors: report("mockResolvedValue", 2, 11),
			},
			{
				Code: `function setup<T extends Promise<number>>(value: T) {
  aVariable.mockReturnValue(Promise.resolve(value));
}`,
				Output: []string{},
				Errors: report("mockResolvedValue", 2, 13),
			},
			{
				Code: `declare const values: [Promise<number>];
aVariable.mockReturnValue(Promise.resolve(...values));`,
				Output: []string{},
				Errors: report("mockResolvedValue", 2, 11),
			},
			// A rejection reason is not adopted, and the shorthand takes any value.
			{
				Code: `const p = Promise.resolve(1);
aVariable.mockReturnValue(Promise.reject(p));`,
				Output: []string{`const p = Promise.resolve(1);
aVariable.mockRejectedValue(p);`},
				Errors: report("mockRejectedValue", 2, 11),
			},
			{
				Code: `declare const values: [number];
aVariable.mockReturnValue(Promise.resolve(...values));`,
				Output: []string{`declare const values: [number];
aVariable.mockResolvedValue(...values);`},
				Errors: report("mockResolvedValue", 2, 11),
			},
			{
				Code: `declare const value: any;
aVariable.mockReturnValue(Promise.resolve(value));`,
				Output: []string{`declare const value: any;
aVariable.mockResolvedValue(value);`},
				Errors: report("mockResolvedValue", 2, 11),
			},

			// --- a declaration the rewrite would delete withholds the fix ---
			{
				Code:   `aVariable.mockImplementation(function impl() { return Promise.resolve(impl); });`,
				Output: []string{},
				Errors: report("mockResolvedValue", 1, 11),
			},
			{
				Code: `aVariable.mockImplementation(() => {
  return Promise.resolve(getValue());
  function getValue() { return 42; }
});`,
				Output: []string{},
				Errors: report("mockResolvedValue", 1, 11),
			},
		},
	)
}

// TestPreferMockPromiseShorthandWithoutTypeInfo locks the source-only path: with
// no way to tell whether a resolved value is a promise, the fix is kept.
func TestPreferMockPromiseShorthandWithoutTypeInfo(t *testing.T) {
	r := PreferMockPromiseShorthandRule
	r.Run = func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		ctx.TypeChecker = nil
		return PreferMockPromiseShorthandRule.Run(ctx, options)
	}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&r,
		[]rule_tester.ValidTestCase{
			{Code: `let value = 1;
aVariable.mockImplementation(() => Promise.resolve(value));`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `const p = Promise.resolve(1);
aVariable.mockReturnValue(Promise.resolve(p));`,
				Output: []string{`const p = Promise.resolve(1);
aVariable.mockResolvedValue(p);`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "useMockShorthand", Message: "Prefer mockResolvedValue",
					Line: 2, Column: 11,
				}},
			},
			{
				Code: `function setup(undefined: number) {
  return rs.fn().mockReturnValue(Promise.resolve());
}`,
				Output: []string{`function setup(undefined: number) {
  return rs.fn().mockResolvedValue(void 0);
}`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "useMockShorthand", Message: "Prefer mockResolvedValue",
					Line: 2, Column: 18,
				}},
			},
		},
	)
}
