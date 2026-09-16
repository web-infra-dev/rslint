package handle_callback_err_test

// cspell:ignore errerr

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/handle_callback_err"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestHandleCallbackErrExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &handle_callback_err.HandleCallbackErrRule,
		[]rule_tester.ValidTestCase{
			// A default on an enclosing destructuring pattern initializes its bindings.
			{Code: `function f({outer: {err} = {}}) {}`},
			// Type predicates count as parameter references.
			{Code: `function f(err: unknown): err is Error { return true; }`},
			// A type and value parameter may share one scope variable upstream.
			{Code: `function f<err>(err: unknown) { let value: err; }`},
			// JSX member receivers read the parameter, unlike a plain intrinsic tag.
			{Code: `const f = (err) => <err.Member />;`, FileName: "input.tsx"},
			// A differently named catch binding leaves the parameter's write visible.
			{Code: `function f(err) { try {} catch (error) { var err = 1; } }`},
			// Constructor parameter properties also have a local parameter binding.
			{Code: `class C { constructor(public err: Error) { use(err); } }`},
			// No parameters or a later error parameter.
			{Code: `function f() {} const g = (data, err) => {};`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// Explicit default, and empty option falls back to err.
			{Code: `function f(error) {}`, Options: []any{``}, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// Custom literal replaces the default.
			{Code: `function f(err) {}`, Options: []any{`error`}, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// Only a leading caret enables regexp syntax.
			{Code: `function f(err) {}`, Options: []any{`err|error`}, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// Parameter defaults and nested pattern defaults are writes.
			{Code: `function f(err = null) {} function g({err = null}) {} function h([{err}] = []) {}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// Redeclaration initializers and iteration targets are writes.
			{Code: `function f(err) { var err = null; } function g(err) { for (var err of errors) {} } function h(err) { var {err} = result; } function i(err) { for (var err in errors) {} }`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// References in later defaults and computed binding keys.
			{Code: `function f(err, data = err) {} function g(err, {[err]: value}) {}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// Assignments count even without a read.
			{Code: `function f(err) { err = null; } function g(err) { ({err} = value); } function h(err) { err ||= null; }`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// Parentheses, optional access, computed keys and shorthand.
			{Code: `function f(err) { return ((err))?.[key]; } function g(err) { return {[err]: true, err}; }`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// Nested closures resolve the original parameter.
			{Code: `const f = (err) => class { [err]() {} field = err; static { void err; } };`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// Destructuring checks its first bound name.
			{Code: `function f({data, err}) {} function g([, data, err]) {} function h({err: error}) {}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// Pattern mismatch.
			{Code: `function f(err) {}`, Options: []any{`^error$`}, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// Getter and class static block are not callbacks.
			{Code: `class C { get err() { return value; } static { const err = 1; } }`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// TS signatures do not introduce runtime functions.
			{Code: `declare function f(err: Error): void; type F = (err: Error) => void; interface C { method(err: Error): void; } abstract class A { abstract method(err: Error): void; }`, FileName: `input.ts`},
			// Upstream includes the TS this parameter before err.
			{Code: `function f(this: void, err?: Error): void {}`, FileName: `input.ts`},
			// TS assertions and typeof references count.
			{Code: `function f(err: unknown) { return (err as Error)!; } function g(err: Error) { type T = typeof err; }`, FileName: `input.ts`},
			// JSX expression references count.
			{Code: `const f = (err) => <div>{err}</div>;`, FileName: `input.tsx`},
			// Parameter used only by a JSX component.
			{Code: `const f = (Err) => <Err/>;`, Options: []any{`Err`}, FileName: `input.tsx`},
			// Parameter and function redeclarations share references.
			{Code: `function f(err) { function err() {} err(); }`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
		},
		[]rule_tester.InvalidTestCase{
			// Report ranges preserve constructor modifiers and intervening comments.
			{Code: `class C { public constructor /* ( */ (err: Error) {} }`, Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 38, 1, 53)}},
			// Generator and async methods named constructor remain ordinary methods.
			{Code: `class C { static *constructor<T>(err: T) {} }`, Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 30, 1, 44)}},
			{Code: `class C { static async ["constructor"](err) {} }`, Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 39, 1, 47)}},
			// The Go tester enables the rule as "test"; directives still suppress reports.
			{Code: `/* eslint-disable test */
function f(err) {}
/* eslint-enable test */
function g(err) {}`, Errors: []rule_tester.InvalidTestCaseError{expectedAt(4, 1, 4, 19)}},
			// Parameter decorators are evaluated outside the callback's scope.
			{Code: `class C { method(@decorate(err) err: Error) {} constructor(@decorate(err) err: Error) {} }`, Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 17, 1, 47), expectedAt(1, 59, 1, 89)}},
			// tsgo represents a static method named constructor as a constructor node.
			{Code: `class C { static constructor<T>(err: T) {} }`, Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 29, 1, 43)}},
			// A var declaration is hoisted, but its write resolves to the catch binding.
			{Code: `function f(err) { try {} catch (err) { var err = 1; } }`, Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 56)}},
			{Code: `function f(err) { try {} catch (err) { for (var err of xs) {} } }`, Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 66)}},
			// Empty leading patterns do not hide a later binding.
			{Code: `function f({}, [,,], err) {}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 29)}},
			// Renamed and nested bindings preserve binding order.
			{Code: `function f({value: [err], later}) {}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 37)}},
			// Rest parameters bind the first name.
			{Code: `const f = (...err) => {};`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 11, 1, 25)}},
			// Bare var declarations are not references.
			{Code: `function f(err) { var err; }`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 29)}},
			// A nested binding shadows the parameter.
			{Code: `function f(err) { { let err = 1; use(err); } try {} catch (err) { use(err); } }`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 80)}},
			// Property and label names do not reference parameters.
			{Code: `function f(err) { obj.err; ({err: true}); err: while (false) { break err; } }`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 78)}},
			// Defaults of other destructured bindings do not use err.
			{Code: `function f({err, data = 1}) {}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 31)}},
			// Function names and same-name parameters are distinct.
			{Code: `const f = function err(err) {};`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 11, 1, 31)}},
			// Explicit default option.
			{Code: `function f(err) {}`, Options: []any{`err`}, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 19)}},
			// Empty option uses default matching.
			{Code: `function f(err) {}`, Options: []any{``}, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 19)}},
			// Unicode regexp matches astral identifier characters.
			{Code: `const f = (𐐀Error) => {};`, Options: []any{`^.$|^.{6}$`}, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 11, 1, 26)}},
			// ECMAScript lookahead and backreferences.
			{Code: `function f(errerr) {}`, Options: []any{`^(?=err)(err)\1$`}, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 22)}},
			// Unicode category escapes use the u flag.
			{Code: `function f(错误) {}`, Options: []any{`^\p{L}+$`}, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 18)}},
			// Unicode escaped names are decoded.
			{Code: `function f(e\u0072r) {}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 24)}},
			// Function declarations exclude export modifiers.
			{Code: `export default async function f(err) {}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 16, 1, 40)}},
			// Anonymous default export.
			{Code: `export default function (err) {}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 16, 1, 33)}},
			// Methods, constructors and setters report their function values.
			{Code: `class C { constructor(err) {} async *method(err) {} set value(err) {} #private(err) {} [key()](err) {} }`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 22, 1, 30), expectedAt(1, 44, 1, 52), expectedAt(1, 62, 1, 70), expectedAt(1, 79, 1, 87), expectedAt(1, 95, 1, 103)}},
			// Object methods and anonymous properties.
			{Code: `const o = { method(err) {}, [key](err) {}, set value(err) {}, fn: function(err) {} };`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 19, 1, 27), expectedAt(1, 34, 1, 42), expectedAt(1, 53, 1, 61), expectedAt(1, 67, 1, 83)}},
			// Multiline function expression with non-ASCII prefix.
			{Code: `const prefix = "😀"; use(
  function (err) {
    log("错误");
  }
);`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(2, 3, 4, 4)}},
			// Arrow ranges exclude wrapping parentheses.
			{Code: `const f = (((err) => {})); const g = async err => 1;`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 13, 1, 24), expectedAt(1, 38, 1, 52)}},
			// An explicitly matched TS this parameter has no identifier references.
			{Code: `function f(this: void) { this; }`, Options: []any{`this`}, FileName: `input.ts`, Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 33)}},
			// A pure type reference does not read a value-only parameter.
			{Code: `function f(err: unknown) { type T = err; }`, FileName: `input.ts`, Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 43)}},
			// TS generic and optional methods.
			{Code: `class C { method?<T>(err: Error): void {} }`, FileName: `input.ts`, Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 18, 1, 42)}},
			// TS parameter properties are still callback parameters.
			{Code: `class C { constructor(public err: Error) {} }`, FileName: `input.ts`, Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 22, 1, 44)}},
			// JSX property names and intrinsic tags are not references.
			{Code: `const f = (err) => <err err="value" />;`, FileName: `input.tsx`, Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 11, 1, 39)}},
			// JSDoc comments do not reference err.
			{Code: `function f(err) { /** @type {typeof err} */ let value; }`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 57)}},
			// An unused function redeclaration is not a reference.
			{Code: `function f(err) { function err() {} }`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 38)}},
		},
	)
}

func TestHandleCallbackErrOptions(t *testing.T) {
	for _, tc := range []struct {
		name    string
		options []any
		wantErr bool
	}{
		{name: "default"},
		{name: "empty name", options: []any{""}},
		{name: "literal name", options: []any{"error"}},
		{name: "literal regexp syntax", options: []any{"["}},
		{name: "non-leading caret", options: []any{"err^["}},
		{name: "escaped leading caret", options: []any{`\^[`}},
		{name: "caret after newline", options: []any{"\n^["}},
		{name: "name pattern", options: []any{`^(err|error)$`}},
		{name: "lookahead and backreference", options: []any{`^(?=err)(err)\1$`}},
		{name: "Unicode category", options: []any{`^\p{L}+$`}},
		{name: "unclosed class", options: []any{`^[`}, wantErr: true},
		{name: "unclosed group", options: []any{`^(`}, wantErr: true},
		{name: "invalid Unicode escape", options: []any{`^\a`}, wantErr: true},
		// Long property names remain unsupported; reject the configuration
		// instead of silently disabling the rule. See Differences from upstream.
		{name: "unsupported Unicode category alias", options: []any{`^\p{Letter}+$`}, wantErr: true},
		{name: "non-string option", options: []any{true}, wantErr: true},
		{name: "extra option", options: []any{"err", "error"}, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := handle_callback_err.HandleCallbackErrRule.Schema.Validate(tc.options)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Validate(%#v) = %v; want error: %v", tc.options, err, tc.wantErr)
			}
		})
	}
}
