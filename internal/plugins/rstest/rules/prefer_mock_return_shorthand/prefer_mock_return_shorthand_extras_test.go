// TestPreferMockReturnShorthandExtras covers what the upstream suites cannot:
// Rstest mock receivers, tsgo node shapes that ESTree does not distinguish, the
// branches of the shared engine that upstream never exercises, and the two places
// this rule deliberately decides differently from both reference plugins.
package prefer_mock_return_shorthand

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferMockReturnShorthandExtras(t *testing.T) {
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
		&PreferMockReturnShorthandRule,
		[]rule_tester.ValidTestCase{
			// --- the method has to be the one this rule rewrites ---
			{Code: `mockImplementation(() => 1);`},
			{Code: `aVariable.mockImplementationTwice(() => 1);`},
			{Code: `aVariable.mockImplementationOnceMore(() => 1);`},
			{Code: `new aVariable.mockImplementation(() => 1);`},
			{Code: `aVariable.mockImplementation?.length;`},
			// A computed identifier key names a variable, so the call site does
			// not say which method is reached.
			{Code: `const name = 'mockImplementation';
aVariable[name](() => 1);`},

			// --- the callback has to be collapsible ---
			{Code: `aVariable.mockImplementation(function (this: Thingy) { return 1; });`},
			{Code: `aVariable.mockImplementation(<T>(): T => null as T);`},
			{Code: `aVariable.mockImplementation(() => { return; });`},
			{Code: `aVariable.mockImplementation(() => {});`},
			{Code: `aVariable.mockImplementation(async () => 1);`},
			{Code: `aVariable.mockImplementation(aFunctionDefinedElsewhere);`},

			// --- a write would move from every call to the one configuration ---
			// Upstream guards only a top-level update expression. Each of these
			// runs its side effect once per call before the rewrite and exactly
			// once after it.
			{Code: `let count = 0;
aVariable.mockImplementation(() => (count += 1));`},
			{Code: `const counter = { value: 0 };
aVariable.mockImplementation(() => (counter.value = 1));`},
			{Code: `const counter = { value: 0 };
aVariable.mockImplementation(() => ({ next: (counter.value += 1) }));`},
			{Code: `const counter = { value: 0 };
aVariable.mockImplementation(() => delete counter.value);`},
			{Code: `const counter = { value: 0 };
aVariable.mockImplementation(() => [counter.value++]);`},
			// A class or method is evaluated each time the expression is, and so
			// are its static initializers, static blocks, computed names and
			// `extends` expression.
			{Code: `const state = { count: 0 };
aVariable.mockImplementation(() => class { static value = ++state.count; });`},
			{Code: `const state = { count: 0 };
aVariable.mockImplementation(() => class { static { state.count++; } });`},
			{Code: `const state = { count: 0 };
aVariable.mockImplementation(() => class { [++state.count]() {} });`},
			{Code: `const state = { count: 0 };
aVariable.mockImplementation(() => class extends (state.count++, Base) {});`},
			{Code: `const state = { count: 0 };
aVariable.mockImplementation(() => ({ [++state.count]() {} }));`},

			// --- a mutable binding would be frozen at its configuration value ---
			// Every shape below is a false positive in @vitest/eslint-plugin,
			// which only inspects the outermost node of the returned expression;
			// the ones using a template, a tag, a class or a TypeScript wrapper
			// are false positives in eslint-plugin-jest too.
			{Code: `let value = 1;
aVariable.mockImplementation(() => ({ value }));`},
			{Code: `let value = 1;
aVariable.mockImplementation(() => [value]);`},
			{Code: `let value = 1;
aVariable.mockImplementation(() => value + 1);`},
			{Code: "let value = 1;\naVariable.mockImplementation(() => `x${value}`);"},
			{Code: "let value = 1;\naVariable.mockImplementation(() => tag`x${value}`);"},
			{Code: `let value = 1;
aVariable.mockImplementation(() => value as number);`},
			{Code: `let value: any = 1;
aVariable.mockImplementation(() => value!);`},
			{Code: `let value = 1;
aVariable.mockImplementation(() => value satisfies number);`},
			{Code: `let value = 1;
aVariable.mockImplementation(() => class { x = value });`},
			{Code: `let value = 1;
aVariable.mockImplementation(() => new Thingy(() => value));`},
			{Code: `let value = 1;
aVariable.mockImplementation(() => ({ deep: { deeper: [1, ...[value]] } }));`},
			{Code: `let key = 'a';
aVariable.mockImplementation(() => ({ [key]: 1 }));`},
			{Code: `var value = 1;
aVariable.mockImplementation(() => value);`},
			{Code: `let { value } = source;
aVariable.mockImplementation(() => value);`},
			{Code: `let [value] = source;
aVariable.mockImplementation(() => value);`},
			// A closure the expression hands back still observes the binding after
			// the mock is configured, so a mutable one blocks the rewrite wherever
			// in the expression it is read.
			{Code: `let count = 0;
aVariable.mockImplementation(() => () => { count += 1; });`},
			// TypeScript treats the operand of a type query as a value reference,
			// so this is left alone even though nothing reads `value` at run time.
			{Code: `let value = 1;
aVariable.mockImplementation(() => 1 as typeof value);`},

			// --- a `function` callback's `this` and `arguments` come from the call ---
			// Both reference plugins lift these out of the callback, where they
			// name whatever encloses the configuration site instead.
			{Code: `aVariable.mockImplementation(function () { return this.id; });`},
			{Code: `aVariable.mockImplementation(function () { return arguments; });`},
			{Code: `aVariable.mockImplementation(function () { return () => this.id; });`},
			// `new.target` is bound the same way: a mock can be called with `new`,
			// and lifting the read out of the callback either fails to parse or
			// starts reporting on the enclosing function instead.
			{Code: `aVariable.mockImplementation(function () { return new.target; });`},
			{Code: `function outer() {
  aVariable.mockImplementation(function () { return new.target; });
}`},
			{Code: `aVariable.mockImplementation(function () { return () => new.target; });`},
			// A nested method or class rebinds `this` only inside its bodies. Its
			// computed names, decorators and `extends` expression are evaluated
			// with the callback's bindings.
			{Code: `aVariable.mockImplementation(function () { return { [this.key]() {} }; });`},
			{Code: `aVariable.mockImplementation(function () { return { get [this.key]() { return 1; } }; });`},
			{Code: `aVariable.mockImplementation(function () { return class extends this.Base {}; });`},
			{Code: `aVariable.mockImplementation(function () { return class { [this.key]() {} }; });`},
			{Code: `aVariable.mockImplementation(function () { return class { static [arguments[0]] = 1; }; });`},
			{Code: `aVariable.mockImplementation(function () { return class { @this.decorate method() {} }; });`},

			// --- a rejected promise is guarded in every accessor spelling ---
			{Code: `aVariable.mockImplementation(() => Promise['reject'](13));`},
			{Code: "aVariable.mockImplementation(() => Promise[`reject`](13));"},
			{Code: `aVariable.mockImplementation(() => Promise?.reject(13));`},
			// Type assertions are erased at run time, so they do not hide the
			// rejected promise from the guard.
			{Code: `aVariable.mockImplementation(() => Promise.reject(new Error('nope')) as Promise<never>);`},
			{Code: `aVariable.mockImplementation(() => Promise.reject(new Error('nope'))!);`},
			{Code: `aVariable.mockImplementation(() => Promise.reject(new Error('nope')) satisfies Promise<never>);`},
			{Code: `aVariable.mockImplementation(() => <Promise<never>>Promise.reject(new Error('nope')));`},
			{Code: `aVariable.mockImplementation(() => (Promise as any).reject(13));`},
			{Code: `aVariable.mockImplementation(() => Promise!.reject(13));`},
			{Code: `aVariable.mockImplementation(() => (Promise.reject as any)(13));`},
			{Code: `aVariable.mockImplementation(() => (Promise.reject<never>)(new Error('nope')));`},
			{Code: `aVariable.mockImplementation(() => (Promise<never>).reject(new Error('nope')));`},

			// --- a generator returns an iterator, not the value it names ---
			// Both reference plugins rewrite this to `mockReturnValue(1)`, which
			// changes what every caller of the mock gets back.
			{Code: `aVariable.mockImplementation(function* () { return 1; });`},
			{Code: `aVariable.mockImplementationOnce(function* named() { return 1; });`},

			// --- a rejected promise must stay where it is built ---
			{Code: `rs.fn().mockImplementationOnce(() => Promise.reject(new Error('nope')));`},
			{Code: `rs.spyOn(fs.promises, 'readFile').mockImplementation(() => {
  return Promise.reject(new Error('nope'));
});`},
		},
		[]rule_tester.InvalidTestCase{
			// --- Rstest mock receivers ---
			{
				Code:   `rs.fn().mockImplementation(() => 1);`,
				Output: []string{`rs.fn().mockReturnValue(1);`},
				Errors: report("mockReturnValue", 1, 9),
			},
			{
				Code:   `rstest.fn().mockImplementationOnce(() => 1);`,
				Output: []string{`rstest.fn().mockReturnValueOnce(1);`},
				Errors: report("mockReturnValueOnce", 1, 13),
			},
			{
				Code:   `import.meta.rstest.fn().mockImplementation(() => 1);`,
				Output: []string{`import.meta.rstest.fn().mockReturnValue(1);`},
				Errors: report("mockReturnValue", 1, 25),
			},
			{
				Code: `import { fn as makeMock } from '@rstest/core';
makeMock().mockImplementation(() => 1);`,
				Output: []string{`import { fn as makeMock } from '@rstest/core';
makeMock().mockReturnValue(1);`},
				Errors: report("mockReturnValue", 2, 12),
			},
			{
				Code:   `rs.spyOn(thingy, 'method').mockImplementation(() => 1);`,
				Output: []string{`rs.spyOn(thingy, 'method').mockReturnValue(1);`},
				Errors: report("mockReturnValue", 1, 28),
			},
			{
				Code:   `rs.mocked(thingy.method).mockImplementation(() => 1);`,
				Output: []string{`rs.mocked(thingy.method).mockReturnValue(1);`},
				Errors: report("mockReturnValue", 1, 26),
			},

			// --- accessor spellings ---
			// Both reference plugins replace the whole string literal here,
			// producing `aVariable[mockReturnValue](1)` — a reference to an
			// undeclared variable. Only the text between the delimiters is
			// rewritten.
			{
				Code:   `aVariable['mockImplementation'](() => 1);`,
				Output: []string{`aVariable['mockReturnValue'](1);`},
				Errors: report("mockReturnValue", 1, 11),
			},
			{
				Code:   "aVariable[`mockImplementationOnce`](() => 1);",
				Output: []string{"aVariable[`mockReturnValueOnce`](1);"},
				Errors: report("mockReturnValueOnce", 1, 11),
			},
			{
				Code:   `aVariable?.mockImplementation(() => 1);`,
				Output: []string{`aVariable?.mockReturnValue(1);`},
				Errors: report("mockReturnValue", 1, 12),
			},
			{
				Code:   `aVariable?.['mockImplementation'](() => 1);`,
				Output: []string{`aVariable?.['mockReturnValue'](1);`},
				Errors: report("mockReturnValue", 1, 13),
			},
			{
				Code:   `(aVariable.mockImplementation)(() => 1);`,
				Output: []string{`(aVariable.mockReturnValue)(1);`},
				Errors: report("mockReturnValue", 1, 12),
			},

			// --- tsgo node shapes around the callback and its result ---
			{
				Code:   `aVariable.mockImplementation((() => 1));`,
				Output: []string{`aVariable.mockReturnValue((1));`},
				Errors: report("mockReturnValue", 1, 11),
			},
			{
				Code:   `aVariable.mockImplementation(() => (1));`,
				Output: []string{`aVariable.mockReturnValue(1);`},
				Errors: report("mockReturnValue", 1, 11),
			},
			{
				Code:   `aVariable.mockImplementation((): number => 1);`,
				Output: []string{`aVariable.mockReturnValue(1);`},
				Errors: report("mockReturnValue", 1, 11),
			},
			{
				Code:   `aVariable.mockImplementation(() => 1 as number);`,
				Output: []string{`aVariable.mockReturnValue(1 as number);`},
				Errors: report("mockReturnValue", 1, 11),
			},
			{
				Code:   `aVariable.mockImplementation(() => thingy!);`,
				Output: []string{`aVariable.mockReturnValue(thingy!);`},
				Errors: report("mockReturnValue", 1, 11),
			},
			{
				Code:   `aVariable.mockImplementation<() => number>(() => 1);`,
				Output: []string{`aVariable.mockReturnValue<() => number>(1);`},
				Errors: report("mockReturnValue", 1, 11),
			},
			// A comma expression keeps the parentheses the rewrite would
			// otherwise drop. Both reference plugins emit `mockReturnValue(0, 1)`
			// here, which returns 0 rather than 1.
			{
				Code:   `aVariable.mockImplementation(() => (0, 1));`,
				Output: []string{`aVariable.mockReturnValue((0, 1));`},
				Errors: report("mockReturnValue", 1, 11),
			},
			{
				Code: `aVariable.mockImplementation(() => {
  return (0, 1);
});`,
				Output: []string{`aVariable.mockReturnValue((0, 1));`},
				Errors: report("mockReturnValue", 1, 11),
			},

			// --- bindings that stay fixable ---
			{
				Code: `const { value } = source;
aVariable.mockImplementation(() => value);`,
				Output: []string{`const { value } = source;
aVariable.mockReturnValue(value);`},
				Errors: report("mockReturnValue", 2, 11),
			},
			{
				Code: `function makeMock(value) {
  return rs.fn().mockImplementation(() => value);
}`,
				Output: []string{`function makeMock(value) {
  return rs.fn().mockReturnValue(value);
}`},
				Errors: report("mockReturnValue", 2, 18),
			},
			// A write inside a function the expression only hands back does not
			// run while the value is built, so only the binding it reads decides.
			{
				Code: `const counter = { value: 0 };
aVariable.mockImplementation(() => () => { counter.value += 1; });`,
				Output: []string{`const counter = { value: 0 };
aVariable.mockReturnValue(() => { counter.value += 1; });`},
				Errors: report("mockReturnValue", 2, 11),
			},

			// An arrow has no `this` or `arguments` of its own, so both already name
			// the enclosing scope and the rewrite does not move them.
			{
				Code:   `aVariable.mockImplementation(() => this.id);`,
				Output: []string{`aVariable.mockReturnValue(this.id);`},
				Errors: report("mockReturnValue", 1, 11),
			},
			// A nested `function` rebinds `this` and `new.target`, so lifting them
			// is safe.
			{
				Code:   `aVariable.mockImplementation(function () { return function () { return this.id; }; });`,
				Output: []string{`aVariable.mockReturnValue(function () { return this.id; });`},
				Errors: report("mockReturnValue", 1, 11),
			},
			{
				Code:   `aVariable.mockImplementation(function () { return function () { return new.target; }; });`,
				Output: []string{`aVariable.mockReturnValue(function () { return new.target; });`},
				Errors: report("mockReturnValue", 1, 11),
			},
			// `import.meta` is the other meta property and names the module, not
			// the call, so it travels with the expression.
			{
				Code:   `aVariable.mockImplementation(function () { return import.meta.url; });`,
				Output: []string{`aVariable.mockReturnValue(import.meta.url);`},
				Errors: report("mockReturnValue", 1, 11),
			},
			// A comma expression written without parentheses gains them, so the
			// mock keeps returning the last operand.
			{
				Code: `aVariable.mockImplementation(() => {
  return 0, 1;
});`,
				Output: []string{`aVariable.mockReturnValue((0, 1));`},
				Errors: report("mockReturnValue", 1, 11),
			},

			// --- the rejected-promise guard is about the global Promise ---
			{
				Code: `const Promise = FakePromise;
aVariable.mockImplementation(() => Promise.reject(1));`,
				Output: []string{`const Promise = FakePromise;
aVariable.mockReturnValue(Promise.reject(1));`},
				Errors: report("mockReturnValue", 2, 11),
			},
			{
				Code:   `aVariable.mockImplementation(() => Promise.reject);`,
				Output: []string{`aVariable.mockReturnValue(Promise.reject);`},
				Errors: report("mockReturnValue", 1, 11),
			},
			{
				Code:   `aVariable.mockImplementation(() => thingy.reject(1));`,
				Output: []string{`aVariable.mockReturnValue(thingy.reject(1));`},
				Errors: report("mockReturnValue", 1, 11),
			},
			{
				Code:   `aVariable.mockImplementation(() => Promise['resolve'](1));`,
				Output: []string{`aVariable.mockReturnValue(Promise['resolve'](1));`},
				Errors: report("mockReturnValue", 1, 11),
			},
			{
				Code: `const Promise = FakePromise;
aVariable.mockImplementation(() => Promise['reject'](1));`,
				Output: []string{`const Promise = FakePromise;
aVariable.mockReturnValue(Promise['reject'](1));`},
				Errors: report("mockReturnValue", 2, 11),
			},

			// --- a comment the rewrite would delete withholds the fix ---
			{
				Code:   `aVariable.mockImplementation(() => /* why */ 1);`,
				Output: []string{},
				Errors: report("mockReturnValue", 1, 11),
			},
			{
				Code: `aVariable.mockImplementation(() => {
  // why
  return 1;
});`,
				Output: []string{},
				Errors: report("mockReturnValue", 1, 11),
			},
			{
				Code: `aVariable.mockImplementation(() => {
  return 1; // why
});`,
				Output: []string{},
				Errors: report("mockReturnValue", 1, 11),
			},
			// --- a declaration the rewrite would delete withholds the fix ---
			{
				Code:   `aVariable.mockImplementation(function impl() { return impl; });`,
				Output: []string{},
				Errors: report("mockReturnValue", 1, 11),
			},
			{
				Code:   `aVariable.mockImplementation(function impl() { return () => impl; });`,
				Output: []string{},
				Errors: report("mockReturnValue", 1, 11),
			},
			{
				Code: `aVariable.mockImplementation(() => {
  return getValue();
  function getValue() { return 42; }
});`,
				Output: []string{},
				Errors: report("mockReturnValue", 1, 11),
			},
			// A name that only matches the callback's, but resolves elsewhere,
			// keeps the fix.
			{
				Code:   `aVariable.mockImplementation(function impl() { return function impl() { return impl; }; });`,
				Output: []string{`aVariable.mockReturnValue(function impl() { return impl; });`},
				Errors: report("mockReturnValue", 1, 11),
			},
			{
				Code:   `aVariable.mockImplementation(function impl() { return 1; });`,
				Output: []string{`aVariable.mockReturnValue(1);`},
				Errors: report("mockReturnValue", 1, 11),
			},
			// A method or class body runs later, with its own `this`, so only the
			// parts evaluated with the class keep the rewrite from happening.
			{
				Code:   `aVariable.mockImplementation(function () { return class extends Base { static self = this; method() { return this; } }; });`,
				Output: []string{`aVariable.mockReturnValue(class extends Base { static self = this; method() { return this; } });`},
				Errors: report("mockReturnValue", 1, 11),
			},
			{
				Code: `const state = { count: 0 };
aVariable.mockImplementation(() => class { value = ++state.count; method() { state.count++; } });`,
				Output: []string{`const state = { count: 0 };
aVariable.mockReturnValue(class { value = ++state.count; method() { state.count++; } });`},
				Errors: report("mockReturnValue", 2, 11),
			},
			// A comment outside the callback is not touched by the rewrite.
			{
				Code:   `aVariable.mockImplementation(/* why */ () => 1);`,
				Output: []string{`aVariable.mockReturnValue(/* why */ 1);`},
				Errors: report("mockReturnValue", 1, 11),
			},
		},
	)
}
