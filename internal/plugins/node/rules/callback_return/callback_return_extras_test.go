package callback_return_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/callback_return"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestCallbackReturnExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &callback_return.CallbackReturnRule,
		[]rule_tester.ValidTestCase{
			// TypeScript instantiation, assertion and satisfies expressions remain callee wrappers.
			{Code: `function f() { if (err) { (cb<string>)(); (<Function>cb)(); (cb satisfies Function)(); } }`},
			// JSDoc casts and Unicode whitespace preserve a terminal callback expression.
			{Code: "function f() {\n  /** @type {void} */ (\uFEFF(cb)());\n}", FileName: "review.js", TSConfig: "tsconfig.allowJs.json"},
			// The rule tester uses the name "test" for suppression comments.
			{Code: `function f() {
  if (err) {
    /* eslint-disable test */
    cb();
  }
  /* eslint-enable test */
  cb();
}`},
			// An empty name list disables matching; custom names replace the defaults.
			{Code: `function f() { if (err) { callback(); cb(); next(); } }`, Options: []any{[]any{}}},
			// Parentheses are transparent when comparing callback expressions.
			{Code: `function f() { (cb()); } function g() { ready && (cb()); }`},
			// Matching uses authored text, including comments, whitespace and escapes.
			{Code: `function f() { if (err) { obj . done(); obj./* gap */done(); c\u0062(); } }`, Options: []any{[]any{"obj.done", "cb"}}},
			// Calls, this, super and constructors do not have identifier receivers.
			{Code: `class C extends B { f() { if (err) { this.done(); super.done(); obj().done(); new cb(); } } }`, Options: []any{[]any{"this.done", "super.done", "obj().done", "cb"}}},
			// Parentheses terminating an optional chain prevent the receiver match.
			{Code: `function f() { if (err) { (obj?.done)(); (obj?.part).done(); } }`, Options: []any{[]any{"obj?.done", "(obj?.part).done"}}},
			{Code: `const f = () => cb?.(); function g() { return cb?.(); }`},
			// Binary and logical expressions allow only a direct callback on their right.
			{Code: `function f() { value + cb(); } function g() { value ?? cb(); } function h() { value || cb(); }`},
			// The immediately preceding statement must be a callback expression.
			{Code: `function f() { if (err) { ready && (cb()); return value; } }`},
			// Upstream permits a callback before a return even if that return calls it again.
			{Code: `function f() { cb(); return cb(); }`},
			{Code: `function f() { switch (x) { case 1: { cb(); return; } } }`},
			// Nested functions are checked independently; declarations after a callback still count as later statements.
			{Code: `function f() { if (err) { (function () { cb(); })(); } cb(); }`},
			{Code: `const f = (value = cb()) => {};`},
			// Method bodies include constructors and accessors.
			{Code: `class C { constructor() { cb(); } get x() { cb(); } set x(value) { cb(); } }`},
			// Computed method keys and class initializers are outside the method function.
			{Code: `const o = { [cb()]() {} }; class C { [cb()]() {} field = cb(); static { cb(); } }`},
			// A static block is not an ESTree BlockStatement and must not hide an enclosing return.
			{Code: `function f() { return class C { static { cb(); } }; }`},
			// Authored TypeScript callee wrappers remain visible.
			{Code: `function f() { if (err) { (cb as any)(); cb!(); (obj as any).done(); } }`, Options: []any{[]any{"cb", "(obj as any).done"}}, FileName: "input.ts"},
			// Body-absent TypeScript declarations are not runtime function expressions.
			{Code: `declare function f(cb: () => void): void; abstract class C { abstract f(cb: () => void): void; }`, FileName: "input.ts"},
			// Method decorators are evaluated outside the function value.
			{Code: `class C { @cb() method() {} }`, FileName: "input.ts"},
			// JSDoc casts in JavaScript do not introduce ESTree wrappers.
			{Code: `function f() { /** @type {void} */ (cb()); }`, FileName: "input.js", TSConfig: "tsconfig.allowJs.json"},
			// JSX expression containers do not change the enclosing function or return.
			{Code: `const f = () => <cb next={cb()} />; function g() { return <cb next={cb()} />; }`, FileName: "input.tsx"},
		},
		[]rule_tester.InvalidTestCase{
			// Logical assignments are assignments, not the allowed logical expression form.
			{Code: `function f() { value ||= cb(); } function g() { value ??= cb(); } function h() { value &&= cb(); }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 26, 1, 30), missingReturnAt(1, 59, 1, 63), missingReturnAt(1, 92, 1, 96)}},
			// An optional call does not qualify even immediately before a trailing return.
			{Code: `function f() { cb?.(); return; }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 16, 1, 22)}},
			// Constructing a callback is ignored, but a callback call used as a constructor is checked.
			{Code: `function f() { new (cb())(); }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 21, 1, 25)}},
			// Preserve leading trivia and the exact multiline spelling of member callback names.
			{Code: `function f() {
  if (err) {
    /* leading */  ((cb))(
      "😀"
    );
    obj /* comment
*/ .cb();
  }
}`, Options: []any{[]any{"cb", "obj /* comment\n*/ .cb"}}, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(3, 20, 5, 6), missingReturnAt(6, 5, 7, 9)}},
			// Line suppression preserves diagnostic ranges with CRLF input.
			{Code: "function f() {\r\n  if (err) {\r\n    // eslint-disable-next-line test\r\n    cb();\r\n    next();\r\n  }\r\n}", Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(5, 5, 5, 11)}},
			// Parameter decorators belong to the method function in the upstream TypeScript AST.
			{Code: `class C { constructor(@cb() public value) {} }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 24, 1, 28)}},
			// The omitted option also recognizes Node middleware's next callback.
			{Code: `function middleware(req, res, next) { if (err) next(err); }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 48, 1, 57)}},
			{Code: `function f() { if (err) { cb(); done(); } }`, Options: []any{[]any{"done", "done", ""}}, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 33, 1, 39)}},
			// Explicit defaults recognize all three names.
			{Code: `function f() { if (err) { callback(); cb(); next(); } }`, Options: []any{[]any{"callback", "cb", "next"}}, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 27, 1, 37), missingReturnAt(1, 39, 1, 43), missingReturnAt(1, 45, 1, 51)}},
			{Code: `function f() { if (err) { ((cb))(err); } }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 27, 1, 38)}},
			{Code: `function f() { if (err) { obj . done(); obj./* gap */done(); c\u0062(); } }`, Options: []any{[]any{"obj . done", "obj./* gap */done", "c\\u0062"}}, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 27, 1, 39), missingReturnAt(1, 41, 1, 60), missingReturnAt(1, 62, 1, 71)}},
			{Code: `function f() { if (err) { (obj).done(); (obj.part).done(); } }`, Options: []any{[]any{"(obj).done", "(obj.part).done"}}, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 27, 1, 39), missingReturnAt(1, 41, 1, 58)}},
			// Only the receiver chain is restricted; computed properties may be arbitrary expressions.
			{Code: `function f() { if (err) { obj["done"](); obj[key()](); obj["part"].done(); } }`, Options: []any{[]any{"obj[\"done\"]", "obj[key()]", "obj[\"part\"].done"}}, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 27, 1, 40), missingReturnAt(1, 42, 1, 54), missingReturnAt(1, 56, 1, 74)}},
			{Code: `class C { #done() {} f(obj) { if (err) { obj.#done(); } } }`, Options: []any{[]any{"obj.#done"}}, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 42, 1, 53)}},
			// Optional calls remain inside a ChainExpression, even at the end of a function.
			{Code: `function f() { cb?.(); } function g() { ready && cb?.(); }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 16, 1, 22), missingReturnAt(1, 50, 1, 56)}},
			{Code: `function f() { obj?.done(); obj.part?.(); }`, Options: []any{[]any{"obj?.done", "obj.part"}}, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 16, 1, 27), missingReturnAt(1, 29, 1, 41)}},
			{Code: `function f() { obj?.part.done(); }`, Options: []any{[]any{"obj?.part.done"}}, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 16, 1, 32)}},
			{Code: `function f() { cb() && cb(); } function g() { a && (b && cb()); }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 16, 1, 20), missingReturnAt(1, 58, 1, 62)}},
			// Assignments, sequences, conditionals, await and wrappers do not qualify as a callback statement.
			{Code: `async function f() { await cb(); } function g() { result = cb(); } function h() { a, cb(); }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 28, 1, 32), missingReturnAt(1, 60, 1, 64), missingReturnAt(1, 86, 1, 90)}},
			{Code: `function f() { flag ? cb() : cb(); } function g() { wrap(cb()); }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 23, 1, 27), missingReturnAt(1, 30, 1, 34), missingReturnAt(1, 58, 1, 62)}},
			{Code: `function f() { cb(); ; return; } function g() { cb(); throw err; }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 16, 1, 20), missingReturnAt(1, 49, 1, 53)}},
			// Blocks are checked structurally, without control-flow inference.
			{Code: `function f() { try { cb(); } catch (err) { cb(); return; } finally { cb(); } }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 22, 1, 26), missingReturnAt(1, 70, 1, 74)}},
			{Code: `function f() { for (const x of xs) { cb(); } do { cb(); } while (x); }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 38, 1, 42), missingReturnAt(1, 51, 1, 55)}},
			{Code: `function f() { switch (x) { case 1: cb(); return; } }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 37, 1, 41)}},
			{Code: `function f() { cb(); function g() { cb(); } }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 16, 1, 20)}},
			// Default parameters belong to functions, while an arrow is an implicit-return boundary.
			{Code: `function f(value = cb()) {} const g = function(value = cb()) {};`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 20, 1, 24), missingReturnAt(1, 56, 1, 60)}},
			{Code: `class C { constructor() { if (err) { cb(); } } get x() { if (err) { cb(); } } set x(value) { if (err) { cb(); } } }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 38, 1, 42), missingReturnAt(1, 69, 1, 73), missingReturnAt(1, 105, 1, 109)}},
			{Code: `function f() { const o = { [cb()]() {} }; class C { field = cb(); static { cb(); } } }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 29, 1, 33), missingReturnAt(1, 61, 1, 65), missingReturnAt(1, 76, 1, 80)}},
			{Code: `const o = { [cb()](value = cb()) {} };`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 28, 1, 32)}},
			// Default-parameter callbacks are checked in nested function declarations too.
			{Code: `function outer() { function inner(value = cb()) {} }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 43, 1, 47)}},
			{Code: `function f() { cb() as void; } function g() { cb()!; } function h() { cb() satisfies void; }`, FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 16, 1, 20), missingReturnAt(1, 47, 1, 51), missingReturnAt(1, 71, 1, 75)}},
			{Code: `function f() { if (err) { cb<void>(); } }`, FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 27, 1, 37)}},
			{Code: `function f() { class C { @cb() method() {} } }`, FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 27, 1, 31)}},
			{Code: `function f() { if (err) { (/** @type {Function} */ (cb))(); } }`, FileName: "input.js", TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 27, 1, 59)}},
			{Code: `function f() { const view = <cb next={cb()} />; }`, FileName: "input.tsx", Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 39, 1, 43)}},
			// Diagnostics cover the complete call using UTF-16 columns and multiline ranges.
			{Code: `function f() { "😀"; if (err) { (callback)(
  err,
); } }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 33, 3, 2)}},
		},
	)
}
