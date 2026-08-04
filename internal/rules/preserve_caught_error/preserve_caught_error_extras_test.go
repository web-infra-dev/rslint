package preserve_caught_error

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// TestPreserveCaughtErrorExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers,
// so future refactors can't silently regress them without breaking a named
// lock-in.
//
// N/A Dimension 4 rows:
//   - PrivateIdentifier keys: an object literal cannot declare `#cause`, and the
//     rule only reads keys of the error-options object literal.
//   - Class declaration vs class expression, function declaration vs expression
//     vs arrow: the rule targets `throw` statements, so these shapes matter only
//     as traversal boundaries — each is covered by a boundary case below.
//   - Overload signatures / `abstract` / `declare` members: a body-less member
//     holds no `throw` statement to report on.
func TestPreserveCaughtErrorExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreserveCaughtErrorRule,
		[]rule_tester.ValidTestCase{
			// Locks in upstream isBuiltInGlobalError arm 2: a global turned off
			// through the config un-declares the built-in error constructor.
			{Code: `try {} catch (err) { throw new Error("m"); }`, Globals: map[string]bool{"Error": false}},
			// ---- Dimension 4: parenthesized options argument already carrying the cause ----
			{Code: `try {} catch (err) { throw new Error("m", ({ cause: err })); }`},
			// ---- Dimension 4: parenthesized cause value matching the caught error ----
			{Code: `try {} catch (err) { throw new Error("m", { cause: (err) }); }`},
			// ---- Dimension 4: non-null assertion on the callee ----
			{Code: `try {} catch (err) { throw new Error!("m"); }`},
			// ---- Dimension 4: non-null assertion on the thrown expression ----
			{Code: `try {} catch (err) { throw (new Error("m"))!; }`},
			// ---- Dimension 4: type assertion on the thrown expression ----
			{Code: `try {} catch (err) { throw (new Error("m") as Error); }`},
			// ---- Dimension 4: satisfies wrapper on the options argument ----
			{Code: `try {} catch (err) { throw new Error("m", {} satisfies object); }`},
			// ---- Dimension 4: optional call of a global error constructor ----
			{Code: `try {} catch (err) { throw Error?.("m"); }`},
			// ---- Dimension 4: optional chain reaching a configured error class ----
			{Code: `try {} catch (err) { throw a?.b.AppError("m"); }`, Options: map[string]interface{}{"errorClassNames": []interface{}{"AppError"}}},
			// ---- Dimension 4: optional call of a configured error class ----
			{Code: `try {} catch (err) { throw obj.AppError?.("m"); }`, Options: map[string]interface{}{"errorClassNames": []interface{}{"AppError"}}},
			// ---- Dimension 4: parenthesized optional chain as the call callee ----
			{Code: `try {} catch (err) { throw (ns?.AppError)("m"); }`, Options: map[string]interface{}{"errorClassNames": []interface{}{"AppError"}}},
			// ---- Dimension 4: parenthesized optional chain as the new callee ----
			{Code: `try {} catch (err) { throw new (ns?.AppError)("m"); }`, Options: map[string]interface{}{"errorClassNames": []interface{}{"AppError"}}},
			// ---- Dimension 4: the chain root itself is parenthesized, the property is not ----
			{Code: `try {} catch (err) { throw (ns?.errors.AppError)("m"); }`, Options: map[string]interface{}{"errorClassNames": []interface{}{"AppError"}}},
			// ---- Dimension 4: element access callee is never a configured error class ----
			{Code: `try {} catch (err) { throw new lib["AppError"]("m"); }`, Options: map[string]interface{}{"errorClassNames": []interface{}{"AppError"}}},
			// ---- Dimension 4: function declaration boundary ----
			{Code: `try {} catch (err) { function f() { throw new Error("m"); } }`},
			// ---- Dimension 4: function expression boundary ----
			{Code: `try {} catch (err) { const f = function () { throw new Error("m"); }; }`},
			// ---- Dimension 4: arrow function boundary ----
			{Code: `try {} catch (err) { const f = () => { throw new Error("m"); }; }`},
			// ---- Dimension 4: async generator boundary ----
			{Code: `try {} catch (err) { async function* f() { throw new Error("m"); } }`},
			// ---- Dimension 4: class method boundary ----
			{Code: `try {} catch (err) { class A { m() { throw new Error("m"); } } }`},
			// ---- Dimension 4: class constructor boundary ----
			{Code: `try {} catch (err) { class A { constructor() { throw new Error("m"); } } }`},
			// ---- Dimension 4: class getter boundary ----
			{Code: `try {} catch (err) { class A { get p() { throw new Error("m"); } } }`},
			// ---- Dimension 4: class setter boundary ----
			{Code: `try {} catch (err) { class A { set p(v) { throw new Error("m"); } } }`},
			// ---- Dimension 4: class static block boundary ----
			{Code: `try {} catch (err) { class A { static { throw new Error("m"); } } }`},
			// ---- Dimension 4: class field arrow boundary ----
			{Code: `try {} catch (err) { class A { f = () => { throw new Error("m"); }; } }`},
			// ---- Dimension 4: object accessor boundary ----
			{Code: `try {} catch (err) { o = { get p() { throw new Error("m"); } }; }`},
			// ---- Dimension 4: nesting boundary — the innermost catch owns the throw ----
			{Code: `try {} catch (outer) { try {} catch (inner) { throw new Error("m", { cause: inner }); } }`},
			// ---- Dimension 4: spread member in the options bag ----
			{Code: `try {} catch (err) { throw new Error("m", { ...o, cause: err }); }`},
			// ---- Dimension 4: spread member after the cause in the options bag ----
			{Code: `try {} catch (err) { throw new Error("m", { cause: err, ...o }); }`},
			// ---- Dimension 4: spread argument before the options position ----
			{Code: `try {} catch (err) { throw new Error(...a); }`},
			// ---- Dimension 4: throwing a value that is not a constructed error ----
			{Code: `try {} catch (err) { throw err; }`},
			// ---- Dimension 4: options argument that is not an object literal ----
			{Code: `try {} catch (err) { throw new Error("m", opts); }`},
			// Locks in upstream ThrowStatement listener arm 1: a throw outside any catch is never checked.
			{Code: `throw new Error("m");`},
			// Locks in upstream isBuiltInGlobalError arm 1: a name outside the built-in set is only checked when configured.
			{Code: `try {} catch (err) { throw new Foo("m"); }`},
			// Locks in upstream isBuiltInGlobalError arm 2: a local class declaration un-declares the global.
			{Code: `class Error {}
try {} catch (err) { throw new Error("m"); }`},
			// Locks in upstream isBuiltInGlobalError arm 2: a local variable un-declares the global.
			{Code: `let TypeError = Custom;
try {} catch (err) { throw new TypeError("m"); }`},
			// Locks in upstream isBuiltInGlobalError arm 2: a namespace declaration un-declares the global.
			{Code: `namespace RangeError {}
try {} catch (err) { throw new RangeError("m"); }`},
			// Locks in upstream optionsIndex arm 1: AggregateError takes its options third.
			{Code: `try {} catch (err) { throw new AggregateError([e], "m", { cause: err }); }`},
			// Locks in upstream optionsIndex arm 2: other built-in errors take their options second.
			{Code: `try {} catch (err) { throw new EvalError("m", { cause: err }); }`},
			// Locks in upstream optionsIndex arm 3: a string entry means the options come second.
			{Code: `try {} catch (err) { throw new AppError("m", { cause: err }); }`, Options: map[string]interface{}{"errorClassNames": []interface{}{"AppError"}}},
			// Locks in upstream optionsIndex arm 3: argumentPosition 1 puts the options first.
			{Code: `try {} catch (err) { throw new AppError({ cause: err }); }`, Options: map[string]interface{}{"errorClassNames": []interface{}{map[string]interface{}{"name": "AppError", "argumentPosition": 1}}}},
			// Locks in upstream getErrorCause arm 1: a spread before the options position blocks the check.
			{Code: `try {} catch (err) { throw new AggregateError(...a, "m", {}); }`},
			// Locks in upstream getErrorCause arm 1: a spread exactly at the options position blocks the check.
			{Code: `try {} catch (err) { throw new Error("m", ...a); }`},
			// Locks in upstream getErrorCause arm 2: a non-object options argument blocks the check.
			{Code: `try {} catch (err) { throw new Error("m", cond ? {} : {}); }`},
			// Locks in upstream getErrorCause arm 3: a spread member in the options bag blocks the check.
			{Code: `try {} catch (err) { throw new Error("m", { ...defaults }); }`},
			// Locks in upstream getErrorCause arm 4: the last of several cause definitions decides the verdict.
			{Code: `try {} catch (err) { throw new Error("m", { cause: other, cause: err }); }`},
			// Locks in upstream catch-parameter arm 2: a missing parameter is accepted by default.
			{Code: `try {} catch { throw new Error("m"); }`},
			// Locks in upstream catch-parameter arm 2: requireCatchParameter only fires on a checked error class.
			{Code: `try {} catch { throw "m"; }`, Options: map[string]interface{}{"requireCatchParameter": true}},
			// Locks in upstream catch-parameter arm 3: a type-annotated parameter is still an identifier.
			{Code: `try {} catch (err: unknown) { throw new Error("m", { cause: err }); }`},
			// Locks in upstream cause-value arm 2: a shorthand naming the caught error is accepted.
			{Code: `try {} catch (cause) { throw new Error("m", { cause }); }`},
			// Locks in upstream shadow arm 1: an unshadowed caught error passes.
			{Code: `try {} catch (err) { if (x) { throw new Error("m", { cause: err }); } }`},
			// ---- Dimension 4: a parenthesized computed key names the static `cause` ----
			{Code: `try {} catch (err) { throw new Error("m", { [('cause')]: err }); }`},
			// ---- Dimension 4: `var` hoists out of the catch clause, so it is not a closer binding ----
			{Code: `try {} catch (err) { var err; throw new Error("m", { cause: err }); }`},
			// ---- Dimension 4: `var` in a nested block still hoists past the catch clause ----
			{Code: `try {} catch (err) { { var err; } throw new Error("m", { cause: err }); }`},
			// ---- Dimension 4: a `var` for-of binding is not scoped to the loop ----
			{Code: `try {} catch (err) { for (var err of list) {} throw new Error("m", { cause: err }); }`},
			// ---- Dimension 4: a `var` for binding is not scoped to the loop ----
			{Code: `try {} catch (err) { for (var err = 0; err < 1; err++) {} throw new Error("m", { cause: err }); }`},
			// ---- Dimension 4: a `var` in another switch case is not a closer binding ----
			{Code: `try {} catch (err) { switch (x) { case 1: var err = 1; break; case 2: throw new Error("m", { cause: err }); } }`},
			// ---- Real-user: extra diagnostic data alongside the cause is accepted in any order ----
			{Code: `try {} catch (error) { throw new Error("m", { retryable: true, cause: error }); }`},
			// ---- Options: array-wrapped errorClassNames ----
			{Code: `try {} catch (err) { throw new AppError("m", { cause: err }); }`, Options: []interface{}{map[string]interface{}{"errorClassNames": []interface{}{"AppError"}}}},
			// ---- Options: an empty options object keeps every default ----
			{Code: `try {} catch { throw new Error("m"); }`, Options: map[string]interface{}{}},
			// ---- Options: an unlisted custom error class is never checked ----
			{Code: `try {} catch (err) { throw new AppError("m"); }`, Options: map[string]interface{}{"errorClassNames": []interface{}{"OtherError"}}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized computed key holding the wrong value ----
			{
				Code:   `try {} catch (err) { throw new Error("m", { [('cause')]: other }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 1, Column: 58, EndLine: 1, EndColumn: 63, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new Error("m", { [('cause')]: err }); }`}}}},
			},
			// ---- Dimension 4: a chain broken by parentheses before the property ----
			{
				Code:    `try {} catch (err) { throw (ns?.errors).AppError("m"); }`,
				Options: map[string]interface{}{"errorClassNames": []interface{}{"AppError"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 22, EndLine: 1, EndColumn: 55, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw (ns?.errors).AppError("m", { cause: err }); }`}}}},
			},
			// ---- Dimension 4: type arguments with no argument list ----
			// The suggestion appends the argument list after the type arguments;
			// ESLint puts it before them. See the rule doc's "Differences from
			// ESLint".
			{
				Code:    `try {} catch (err) { throw new AppError<string>; }`,
				Options: map[string]interface{}{"errorClassNames": []interface{}{map[string]interface{}{"name": "AppError", "argumentPosition": 1}}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 22, EndLine: 1, EndColumn: 49, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new AppError<string>({ cause: err }); }`}}}},
			},
			// ---- Dimension 4: parenthesized callee with no argument list ----
			// The suggestion appends the argument list after the parentheses;
			// ESLint puts it inside them. See the rule doc's "Differences from
			// ESLint".
			{
				Code:    `try {} catch (err) { throw new (AppError); }`,
				Options: map[string]interface{}{"errorClassNames": []interface{}{map[string]interface{}{"name": "AppError", "argumentPosition": 1}}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 22, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new (AppError)({ cause: err }); }`}}}},
			},
			// ---- Dimension 4: parenthesized thrown expression ----
			{
				Code:   `try {} catch (err) { throw (new Error("m")); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 22, EndLine: 1, EndColumn: 45, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw (new Error("m", { cause: err })); }`}}}},
			},
			// ---- Dimension 4: multi-level parenthesized thrown expression ----
			{
				Code:   `try {} catch (err) { throw ((new Error("m"))); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 22, EndLine: 1, EndColumn: 47, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw ((new Error("m", { cause: err }))); }`}}}},
			},
			// ---- Dimension 4: multi-level parenthesized callee ----
			{
				Code:   `try {} catch (err) { throw new ((Error))("m"); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 22, EndLine: 1, EndColumn: 47, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new ((Error))("m", { cause: err }); }`}}}},
			},
			// ---- Dimension 4: parenthesized options argument ----
			{
				Code:   `try {} catch (err) { throw new Error("m", ({})); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 22, EndLine: 1, EndColumn: 49, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new Error("m", ({cause: err})); }`}}}},
			},
			// ---- Dimension 4: parenthesized cause value not matching the caught error ----
			{
				Code:   `try {} catch (err) { throw new Error("m", { cause: (other) }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 1, Column: 53, EndLine: 1, EndColumn: 58, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new Error("m", { cause: (err) }); }`}}}},
			},
			// ---- Dimension 4: non-null assertion on the cause value ----
			{
				Code:   `try {} catch (err) { throw new Error("m", { cause: err! }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 1, Column: 52, EndLine: 1, EndColumn: 56, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new Error("m", { cause: err }); }`}}}},
			},
			// ---- Dimension 4: type assertion on the cause value ----
			{
				Code:   `try {} catch (err) { throw new Error("m", { cause: err as Error }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 1, Column: 52, EndLine: 1, EndColumn: 64, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new Error("m", { cause: err }); }`}}}},
			},
			// ---- Dimension 4: parentheses end the optional chain, so the call is checked ----
			{
				Code:    `try {} catch (err) { throw (a?.b).AppError("m"); }`,
				Options: map[string]interface{}{"errorClassNames": []interface{}{"AppError"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 22, EndLine: 1, EndColumn: 49, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw (a?.b).AppError("m", { cause: err }); }`}}}},
			},
			// ---- Dimension 4: single-quoted cause key with a wrong value ----
			{
				Code:   `try {} catch (err) { throw new Error("m", { 'cause': other }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 1, Column: 54, EndLine: 1, EndColumn: 59, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new Error("m", { 'cause': err }); }`}}}},
			},
			// ---- Dimension 4: computed string cause key with a wrong value ----
			{
				Code:   `try {} catch (err) { throw new Error("m", { ["cause"]: other }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 1, Column: 56, EndLine: 1, EndColumn: 61, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new Error("m", { ["cause"]: err }); }`}}}},
			},
			// ---- Dimension 4: computed template cause key with a wrong value ----
			{
				Code:   "try {} catch (err) { throw new Error(\"m\", { [`cause`]: other }); }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 1, Column: 56, EndLine: 1, EndColumn: 61, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: "try {} catch (err) { throw new Error(\"m\", { [`cause`]: err }); }"}}}},
			},
			// ---- Dimension 4: numeric key is not a cause key ----
			{
				Code:   `try {} catch (err) { throw new Error("m", { 0: err }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 22, EndLine: 1, EndColumn: 55, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new Error("m", { 0: err, cause: err }); }`}}}},
			},
			// ---- Dimension 4: computed numeric key is not a cause key ----
			{
				Code:   `try {} catch (err) { throw new Error("m", { [0]: err }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 22, EndLine: 1, EndColumn: 57, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new Error("m", { [0]: err, cause: err }); }`}}}},
			},
			// ---- Dimension 4: nested member callee matches on its property name ----
			{
				Code:    `try {} catch (err) { throw new a.b.AppError("m"); }`,
				Options: map[string]interface{}{"errorClassNames": []interface{}{"AppError"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 22, EndLine: 1, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new a.b.AppError("m", { cause: err }); }`}}}},
			},
			// ---- Dimension 4: member callee behind an element access still matches ----
			{
				Code:    `try {} catch (err) { throw new a["b"].AppError("m"); }`,
				Options: map[string]interface{}{"errorClassNames": []interface{}{"AppError"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 22, EndLine: 1, EndColumn: 53, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new a["b"].AppError("m", { cause: err }); }`}}}},
			},
			// ---- Dimension 4: nesting boundary — a nested try block is still inside the catch ----
			{
				Code:   `try {} catch (err) { try { throw new Error("m"); } catch {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 28, EndLine: 1, EndColumn: 49, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { try { throw new Error("m", { cause: err }); } catch {} }`}}}},
			},
			// ---- Dimension 4: nesting boundary — a nested finally block is still inside the catch ----
			{
				Code:   `try {} catch (err) { try {} finally { throw new Error("m"); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 39, EndLine: 1, EndColumn: 60, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { try {} finally { throw new Error("m", { cause: err }); } }`}}}},
			},
			// ---- Dimension 4: nesting boundary — a catch inside a nested function owns its own throw ----
			{
				Code:   `try {} catch (outer) { function g() { try {} catch (inner) { throw new Error("m"); } } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 62, EndLine: 1, EndColumn: 83, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (outer) { function g() { try {} catch (inner) { throw new Error("m", { cause: inner }); } } }`}}}},
			},
			// ---- Dimension 4: spread argument after the options position ----
			{
				Code:   `try {} catch (err) { throw new Error("m", {}, ...a); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 22, EndLine: 1, EndColumn: 53, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new Error("m", {cause: err}, ...a); }`}}}},
			},
			// ---- Dimension 4: empty argument list holding only a comment ----
			{
				Code:   `try {} catch (err) { throw new Error(/* why */); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 22, EndLine: 1, EndColumn: 49, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new Error("", { cause: err }/* why */); }`}}}},
			},
			// ---- Dimension 4: trailing comma in the options bag ----
			{
				Code:   `try {} catch (err) { throw new Error("m", { a: 1, }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 22, EndLine: 1, EndColumn: 54, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new Error("m", { a: 1, cause: err, }); }`}}}},
			},
			// ---- Dimension 4: multi-line options bag ----
			{
				Code: `try {} catch (err) {
  throw new Error("m", {
    a: 1
  });
}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 2, Column: 3, EndLine: 4, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) {
  throw new Error("m", {
    a: 1, cause: err
  });
}`}}}},
			},
			// Locks in upstream isBuiltInGlobalError arm 3: a type-only declaration leaves the global value in place.
			{
				Code: `interface Error {}
try {} catch (err) { throw new Error("m"); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 2, Column: 22, EndLine: 2, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `interface Error {}
try {} catch (err) { throw new Error("m", { cause: err }); }`}}}},
			},
			// Locks in upstream isBuiltInGlobalError arm 3: a type alias leaves the global value in place.
			{
				Code: `type SyntaxError = never;
try {} catch (err) { throw new SyntaxError("m"); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 2, Column: 22, EndLine: 2, EndColumn: 49, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `type SyntaxError = never;
try {} catch (err) { throw new SyntaxError("m", { cause: err }); }`}}}},
			},
			// Locks in upstream isThrowingNewError arm 1: every built-in error constructor is covered.
			{
				Code:   `try {} catch (err) { throw new URIError("m"); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 22, EndLine: 1, EndColumn: 46, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new URIError("m", { cause: err }); }`}}}},
			},
			// Locks in upstream isThrowingNewError arm 2: a shadowed built-in name still matches through errorClassNames.
			{
				Code: `class Error {}
try {} catch (err) { throw new Error("m"); }`,
				Options: map[string]interface{}{"errorClassNames": []interface{}{"Error"}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 2, Column: 22, EndLine: 2, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `class Error {}
try {} catch (err) { throw new Error("m", { cause: err }); }`}}}},
			},
			// Locks in upstream optionsIndex arm 3: a global error name configured through errorClassNames keeps its built-in position.
			{
				Code:    `try {} catch (err) { throw new AggregateError([e], "m"); }`,
				Options: map[string]interface{}{"errorClassNames": []interface{}{map[string]interface{}{"name": "AggregateError", "argumentPosition": 1}}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 22, EndLine: 1, EndColumn: 57, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new AggregateError([e], "m", { cause: err }); }`}}}},
			},
			// Locks in upstream getErrorCause arm 4: several cause definitions suppress the suggestion.
			{
				Code:   `try {} catch (err) { throw new Error("m", { cause: err, cause: other }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 1, Column: 64, EndLine: 1, EndColumn: 69}},
			},
			// Locks in upstream catch-parameter arm 1: an array pattern parameter loses part of the caught error.
			{
				Code:   `try {} catch ([first]) { throw new Error("m"); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "partiallyLostError", Line: 1, Column: 8, EndLine: 1, EndColumn: 49}},
			},
			// Locks in upstream catch-parameter arm 1: the destructuring report wins over requireCatchParameter.
			{
				Code:    `try {} catch ({ message }) { throw new Error("m"); }`,
				Options: map[string]interface{}{"requireCatchParameter": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "partiallyLostError", Line: 1, Column: 8, EndLine: 1, EndColumn: 53}},
			},
			// ---- Dimension 4: empty destructuring pattern parameter ----
			{
				Code:   `try {} catch ({}) { throw new Error("m"); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "partiallyLostError", Line: 1, Column: 8, EndLine: 1, EndColumn: 44}},
			},
			// Locks in upstream cause-value arm 1: a setter cannot carry the caught error.
			{
				Code:   `try {} catch (err) { throw new Error("m", { set cause(v) {} }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 1, Column: 54, EndLine: 1, EndColumn: 60, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new Error("m", { cause: err }); }`}}}},
			},
			// Locks in upstream cause-value arm 1: an async method cannot carry the caught error.
			{
				Code:   `try {} catch (err) { throw new Error("m", { async cause() {} }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 1, Column: 56, EndLine: 1, EndColumn: 61, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new Error("m", { cause: err }); }`}}}},
			},
			// Locks in upstream cause-value arm 1: a generator method cannot carry the caught error.
			{
				Code:   `try {} catch (err) { throw new Error("m", { *cause() {} }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 1, Column: 51, EndLine: 1, EndColumn: 56, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new Error("m", { cause: err }); }`}}}},
			},
			// Locks in upstream cause-value arm 1: a computed accessor key is still recognized as cause.
			{
				Code:   `try {} catch (err) { throw new Error("m", { get ["cause"]() {} }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 1, Column: 58, EndLine: 1, EndColumn: 63, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new Error("m", { cause: err }); }`}}}},
			},
			// Locks in upstream cause-value arm 2: the caught error name is used verbatim in the suggestion.
			{
				Code:   `try {} catch (cause) { throw new Error("m", { cause: other }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 1, Column: 54, EndLine: 1, EndColumn: 59, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (cause) { throw new Error("m", { cause: cause }); }`}}}},
			},
			// Locks in upstream cause-value arm 3: null is not the caught error.
			{
				Code:   `try {} catch (err) { throw new Error("m", { cause: null }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 1, Column: 52, EndLine: 1, EndColumn: 56, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new Error("m", { cause: err }); }`}}}},
			},
			// Locks in upstream shadow arm 2: a const in a nested block shadows the caught error.
			{
				Code:   `try {} catch (err) { { const err = other; throw new Error("m", { cause: err }); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "caughtErrorShadowed", Line: 1, Column: 43, EndLine: 1, EndColumn: 80}},
			},
			// Locks in upstream shadow arm 2: a for-loop binding shadows the caught error.
			{
				Code:   `try {} catch (err) { for (let err of list) { throw new Error("m", { cause: err }); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "caughtErrorShadowed", Line: 1, Column: 46, EndLine: 1, EndColumn: 83}},
			},
			// Locks in upstream shadow arm 2: a class declaration shadows the caught error.
			{
				Code:   `try {} catch (err) { { class err {} throw new Error("m", { cause: err }); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "caughtErrorShadowed", Line: 1, Column: 37, EndLine: 1, EndColumn: 74}},
			},
			// Locks in upstream includeCause fix arm 1: AggregateError with only the errors argument.
			{
				Code:   `try {} catch (err) { throw new AggregateError([e]); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 22, EndLine: 1, EndColumn: 52, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new AggregateError([e], "", { cause: err }); }`}}}},
			},
			// Locks in upstream includeCause fix arm 2: AggregateError with an existing options bag.
			{
				Code:   `try {} catch (err) { throw new AggregateError([e], "m", { a: 1 }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 22, EndLine: 1, EndColumn: 67, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new AggregateError([e], "m", { a: 1, cause: err }); }`}}}},
			},
			// Locks in upstream includeCause fix arm 3: a custom class with fewer arguments than the options position gets no suggestion.
			{
				Code:    `try {} catch (err) { throw new AppError("m"); }`,
				Options: map[string]interface{}{"errorClassNames": []interface{}{map[string]interface{}{"name": "AppError", "argumentPosition": 3}}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 22, EndLine: 1, EndColumn: 46}},
			},
			// Locks in upstream includeCause fix arm 4: a custom class whose options slot is the next argument gets an appended options bag.
			{
				Code:    `try {} catch (err) { throw new AppError("m", ctx); }`,
				Options: map[string]interface{}{"errorClassNames": []interface{}{map[string]interface{}{"name": "AppError", "argumentPosition": 3}}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 22, EndLine: 1, EndColumn: 51, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new AppError("m", ctx, { cause: err }); }`}}}},
			},
			// Locks in upstream includeCause fix arm 5: a custom class taking options first and called without arguments.
			{
				Code:    `try {} catch (err) { throw new AppError(); }`,
				Options: map[string]interface{}{"errorClassNames": []interface{}{map[string]interface{}{"name": "AppError", "argumentPosition": 1}}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 22, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new AppError({ cause: err }); }`}}}},
			},
			// Locks in upstream includeCause fix arm 6: a custom class with an existing options bag.
			{
				Code:    `try {} catch (err) { throw new AppError("m", { a: 1 }); }`,
				Options: map[string]interface{}{"errorClassNames": []interface{}{"AppError"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 22, EndLine: 1, EndColumn: 56, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new AppError("m", { a: 1, cause: err }); }`}}}},
			},
			// ---- Real-user: an AggregateError built from the caught error still needs it as the cause ----
			{
				Code:   `try {} catch (error) { throw new AggregateError([error], "Multiple failures"); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 24, EndLine: 1, EndColumn: 79, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (error) { throw new AggregateError([error], "Multiple failures", { cause: error }); }`}}}},
			},
			// ---- Real-user: a namespaced custom error class is matched by its property name ----
			{
				Code: `const errors = { AppError: class AppError extends Error {} };
try {} catch (error) { throw new errors.AppError("m", { cause: error.message }); }`,
				Options: map[string]interface{}{"errorClassNames": []interface{}{"AppError"}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 2, Column: 64, EndLine: 2, EndColumn: 77, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `const errors = { AppError: class AppError extends Error {} };
try {} catch (error) { throw new errors.AppError("m", { cause: error }); }`}}}},
			},
			// ---- Real-user: a getter returning the caught error is still not the caught error ----
			{
				Code:   `try {} catch (error) { throw new Error("m", { get cause() { return error; } }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 1, Column: 56, EndLine: 1, EndColumn: 76, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (error) { throw new Error("m", { cause: error }); }`}}}},
			},
			// ---- Options: array-wrapped requireCatchParameter ----
			{
				Code:    `try {} catch { throw new Error("m"); }`,
				Options: []interface{}{map[string]interface{}{"requireCatchParameter": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingCatchErrorParam", Line: 1, Column: 16, EndLine: 1, EndColumn: 37}},
			},
			// ---- Message text: missingCause ----
			{
				Code:   `try {} catch (err) { throw new ReferenceError("m"); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Message: "There is no `cause` attached to the symptom error being thrown.", Line: 1, Column: 22, EndLine: 1, EndColumn: 52, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new ReferenceError("m", { cause: err }); }`}}}},
			},
			// ---- Message text: incorrectCause ----
			{
				Code:   `try {} catch (err) { throw new ReferenceError("m", { cause: other }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Message: "The symptom error is being thrown with an incorrect `cause`.", Line: 1, Column: 61, EndLine: 1, EndColumn: 66, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new ReferenceError("m", { cause: err }); }`}}}},
			},
			// ---- Message text: missingCatchErrorParam ----
			{
				Code:    `try {} catch { throw new ReferenceError("m"); }`,
				Options: map[string]interface{}{"requireCatchParameter": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingCatchErrorParam", Message: "The caught error is not accessible because the catch clause lacks the error parameter. Start referencing the caught error using the catch parameter.", Line: 1, Column: 16, EndLine: 1, EndColumn: 46}},
			},
			// ---- Message text: partiallyLostError ----
			{
				Code:   `try {} catch ({ message }) { throw new ReferenceError(message); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "partiallyLostError", Message: "Re-throws cannot preserve the caught error as a part of it is being lost due to destructuring.", Line: 1, Column: 8, EndLine: 1, EndColumn: 66}},
			},
			// ---- Message text: caughtErrorShadowed ----
			{
				Code:   `try {} catch (err) { { const err = other; throw new ReferenceError("m", { cause: err }); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "caughtErrorShadowed", Message: "The caught error is being attached as `cause`, but is shadowed by a closer scoped redeclaration.", Line: 1, Column: 43, EndLine: 1, EndColumn: 89}},
			},
		},
	)
}

// TestPreserveCaughtErrorEditDemand verifies that the suggestion builder does
// not change what the rule reports: diagnostic count, message, and range stay
// identical across every edit demand, and the suggestion is materialized only
// when it was requested.
func TestPreserveCaughtErrorEditDemand(t *testing.T) {
	t.Parallel()

	const source = "try {\n} catch (err) {\n  throw new Error(\"m\");\n}\n"

	program, sourceFile := createPreserveCaughtErrorProgram(t, "edit-demand.ts", source)
	options := rule_tester.ResolveTestCaseOptions(t, &PreserveCaughtErrorRule, nil)

	diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
	for _, demand := range []rule.EditDemand{
		rule.EditDemandNone,
		rule.EditDemandAutofix,
		rule.EditDemandSuggestion,
		rule.EditDemandAll,
	} {
		got := lintPreserveCaughtErrorWithDemand(program, sourceFile, options, demand)
		if len(got) != 1 {
			t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(got))
		}
		if got[0].Message.Id != "missingCause" {
			t.Errorf("demand %d: unexpected message id %q", demand, got[0].Message.Id)
		}
		if got[0].Message.Description != "There is no `cause` attached to the symptom error being thrown." {
			t.Errorf("demand %d: unexpected message %q", demand, got[0].Message.Description)
		}
		diagnostics[demand] = got[0]
	}

	diagnosticsOnly := diagnostics[rule.EditDemandNone]
	for demand, diagnostic := range diagnostics {
		want, got := diagnosticsOnly, diagnostic
		want.FixesPtr, want.Suggestions = nil, nil
		got.FixesPtr, got.Suggestions = nil, nil
		if !reflect.DeepEqual(got, want) {
			t.Errorf("demand %d changed diagnostic metadata:\ngot:  %#v\nwant: %#v", demand, got, want)
		}
	}

	// The rule offers suggestions only, so neither the diagnostics-only nor the
	// autofix demand may materialize anything.
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix} {
		diagnostic := diagnostics[demand]
		if diagnostic.FixesPtr != nil || diagnostic.Suggestions != nil {
			t.Errorf(
				"demand %d unexpectedly materialized edits: fixes=%#v suggestions=%#v",
				demand,
				diagnostic.FixesPtr,
				diagnostic.Suggestions,
			)
		}
	}

	for _, demand := range []rule.EditDemand{rule.EditDemandSuggestion, rule.EditDemandAll} {
		diagnostic := diagnostics[demand]
		if diagnostic.FixesPtr != nil {
			t.Errorf("demand %d unexpectedly materialized autofixes: %#v", demand, diagnostic.FixesPtr)
		}
		if diagnostic.Suggestions == nil || len(*diagnostic.Suggestions) != 1 {
			t.Fatalf("demand %d: suggestions = %#v, want exactly one", demand, diagnostic.Suggestions)
		}
		suggestion := (*diagnostic.Suggestions)[0]
		if suggestion.Message.Id != "includeCause" {
			t.Errorf("demand %d: unexpected suggestion id %q", demand, suggestion.Message.Id)
		}
		fixes := suggestion.Fixes()
		if len(fixes) != 1 || fixes[0].Text != ", { cause: err }" {
			t.Errorf("demand %d: unexpected suggestion fixes %#v", demand, fixes)
		}
	}
}

func lintPreserveCaughtErrorWithDemand(
	program *compiler.Program,
	sourceFile *ast.SourceFile,
	options []any,
	demand rule.EditDemand,
) []rule.RuleDiagnostic {
	var diagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program:         program,
		File:            sourceFile.FileName(),
		HasTypeInfo:     true,
		GetRulesForFile: preserveCaughtErrorConfiguredRules(options),
		ExcludePaths:    []string{},
		Consumer: rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		},
	})
	return diagnostics
}

func preserveCaughtErrorConfiguredRules(options []any) func(*ast.SourceFile) []linter.ConfiguredRule {
	return func(*ast.SourceFile) []linter.ConfiguredRule {
		return []linter.ConfiguredRule{{
			Name:     PreserveCaughtErrorRule.Name,
			Severity: rule.SeverityError,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return PreserveCaughtErrorRule.Run(ctx, options)
			},
		}}
	}
}

func createPreserveCaughtErrorProgram(t testing.TB, fileName string, code string) (*compiler.Program, *ast.SourceFile) {
	t.Helper()

	rootDir := fixtures.GetRootDir()
	fs := utils.NewOverlayVFSForFile(tspath.ResolvePath(rootDir, fileName), code)
	host := utils.CreateCompilerHost(rootDir, fs)
	program, err := utils.CreateProgram(true, fs, rootDir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("failed to create program: %v", err)
	}
	sourceFile := program.GetSourceFile(fileName)
	if sourceFile == nil {
		t.Fatalf("source file %q not found", fileName)
	}
	return program, sourceFile
}
