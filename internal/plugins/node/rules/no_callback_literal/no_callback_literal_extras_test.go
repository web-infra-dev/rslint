// TestNoCallbackLiteralExtras covers branches and edge shapes beyond the upstream suite.
// The original cases live in no_callback_literal_upstream_test.go.
package no_callback_literal

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoCallbackLiteralExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoCallbackLiteralRule,
		[]rule_tester.ValidTestCase{
			// Locks in upstream couldBeError() known expressions: call, member, tagged template, yield.
			{Code: "cb(makeError()); cb(error.cause); cb(error[0]); cb(tag`error`); function* task() { cb(yield 0); }"},
			// Locks in upstream couldBeError() default: unfamiliar expressions remain possible errors.
			{Code: "cb(-1); cb(!error); cb(1 + 2); cb(void 0); cb(this); cb(() => {}); cb(function () {}); cb(class {});"},
			{Code: "async function task() { cb(await 0); }"},
			// Locks in upstream AssignmentExpression: all assignment operators inspect the right operand.
			{Code: "cb(error = null); cb(error += other); cb(error ||= null); cb(error ??= other);"},
			// Locks in upstream LogicalExpression: either arm can be an error, including the left of &&.
			{Code: "cb(error && false); cb(false || error); cb(false ?? error); cb(null && false);"},
			// Locks in upstream ConditionalExpression: null and either possible-error branch are allowed.
			{Code: "cb(test ? null : 1); cb(test ? 1 : null);"},
			// Locks in upstream SequenceExpression: only the final expression determines the result.
			{Code: "cb((1, false, null)); cb(([], {}, error));"},
			// ---- Dimension 4: parentheses around allowed arguments ----
			{Code: "((cb))((((null)))); cb(((error)));"},
			// ---- Dimension 4: optional chains are possible errors ----
			{Code: "cb(error?.cause); cb(error?.()); cb((error?.cause));"},
			// ---- Dimension 4: only identifier callees match; member and computed keys do not ----
			{Code: "object.cb(false); object.callback(false); object[\"cb\"](false); object[`callback`](false); object[0](false); object[Symbol.iterator](false); object?.cb(false); (object?.cb)(false);"},
			{Code: "class Task { #cb() {} run() { this.#cb(false); } }"},
			// Locks in upstream CallExpression(): other identifiers, constructor calls and indirect calls are ignored.
			{Code: "next(false); Callback(false); new cb(false); (0, cb)(false); cb.call(null, false); cb`false`;"},
			// ---- Dimension 4: authored TypeScript argument wrappers stay unknown ----
			{Code: "cb(0 as any); cb(0 satisfies number); cb(0!); cb(<any>0);", FileName: "input.ts"},
			// ---- Dimension 4: authored TypeScript callee wrappers are not identifiers ----
			{Code: "cb!(false); (cb as Function)(false); (cb satisfies Function)(false);", FileName: "input.ts"},
			// ---- Dimension 4: JSX values fall through the unknown-expression arm ----
			{Code: "cb(<Error />); cb(<>error</>);", FileName: "input.tsx"},
			// ---- Dimension 4: empty arguments and unknown spread arguments ----
			{Code: "cb(); callback(); cb(...errors); cb(...[]);"},
			// ---- Real-user: #162 error assertions in a rendering callback ----
			{Code: "try { cb(null, renderToString(component)); } catch (error) { cb(error as Error); }", FileName: "input.ts"},
			// ---- Real-user: #552 literals belong in the result position ----
			{Code: "cb(null, { message: value }); callback(null, [value]); cb(undefined, `result ${value}`);"},
			// N/A: declaration kinds and body-absent forms do not change a call-only listener.
			{Code: "declare function cb(error: Error): void; abstract class Task { abstract callback(): void; }", FileName: "input.ts"},
			// Shadowing does not change the syntactic handling of null or undefined.
			{Code: "function task(undefined) { cb(undefined); }"},
			// Instantiation and non-null expressions around a callee remain distinct from a generic call.
			{Code: "(cb<Error>)(false); cb!?.(false); (cb!)?.(false);", FileName: "input.ts"},
			// Newer regex syntax remains valid in the result position.
			{Code: "cb(null, /(?i:error)/); callback(null, /(?<a>a)|(?<a>b)/);"},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: every non-null Literal kind ----
			{Code: "cb(0); cb(0x10); cb(1e2); cb(1n); cb(true); cb(/error/u); cb(\"\");", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 6), unexpectedLiteralAt(1, 8, 1, 16), unexpectedLiteralAt(1, 18, 1, 25), unexpectedLiteralAt(1, 27, 1, 33), unexpectedLiteralAt(1, 35, 1, 43), unexpectedLiteralAt(1, 45, 1, 57), unexpectedLiteralAt(1, 59, 1, 65)}},
			// ---- Dimension 4: both template literal forms ----
			{Code: "cb(``); cb(`error`); cb(`${new Error()}`);", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 7), unexpectedLiteralAt(1, 9, 1, 20), unexpectedLiteralAt(1, 22, 1, 42)}},
			// ---- Real-user: #552 objects, arrays and templates in the error position ----
			{Code: "load((error, value) => { if (error) cb({ message: error.message }); else callback([value]); });", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 37, 1, 67), unexpectedLiteralAt(1, 74, 1, 91)}},
			// ---- Dimension 4: spread properties and empty arrays remain literal errors ----
			{Code: "cb({ ...error }); cb([...errors]); cb({});", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 17), unexpectedLiteralAt(1, 19, 1, 34), unexpectedLiteralAt(1, 36, 1, 42)}},
			// ---- Dimension 4: parentheses around callee, argument and complete call ----
			{Code: "((cb))(((false))); (callback((0)));", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 18), unexpectedLiteralAt(1, 21, 1, 34)}},
			// ---- Dimension 4: optional calls still visit the identifier callee ----
			{Code: "cb?.(false); (callback)?.(0);", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 12), unexpectedLiteralAt(1, 14, 1, 29)}},
			// ---- Dimension 4: generic calls still have an identifier callee ----
			{Code: "cb<Error>(false);", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 17)}},
			// ---- Dimension 4: JavaScript JSDoc casts remain transparent ----
			{Code: "(/** @type {Function} */ (cb))(/** @type {Error} */ (false));", FileName: "input.js", TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 61)}},
			// JavaScript satisfies casts also remain transparent.
			{Code: "cb(/** @satisfies {boolean} */ (false));", FileName: "input.js", TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 40)}},
			// Locks in upstream AssignmentExpression: even logical assignments inspect only the right side.
			{Code: "cb(error = false); cb(error += false); cb(error &&= false); cb(error ||= false); cb(error ??= false);", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 18), unexpectedLiteralAt(1, 20, 1, 38), unexpectedLiteralAt(1, 40, 1, 59), unexpectedLiteralAt(1, 61, 1, 80), unexpectedLiteralAt(1, 82, 1, 101)}},
			// Locks in upstream SequenceExpression: earlier possible errors are ignored.
			{Code: "cb((error, null, false));", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 25)}},
			// Locks in upstream LogicalExpression: both operands are literal errors.
			{Code: "cb(false && 0); cb(false || 0); cb(false ?? 0);", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 15), unexpectedLiteralAt(1, 17, 1, 31), unexpectedLiteralAt(1, 33, 1, 47)}},
			// Locks in upstream ConditionalExpression: the condition is irrelevant when both branches are literals.
			{Code: "cb(error ? 0 : false);", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 22)}},
			// Nested assignments, sequences, logical and conditional expressions recurse consistently.
			{Code: "cb((value = test ? (error, false) : (0 || [])));", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 48)}},
			// ---- Dimension 4: nested callbacks and shadowed names are still inspected independently ----
			{Code: "function task(cb) { cb(false); return function (callback) { callback(0); }; }", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 21, 1, 30), unexpectedLiteralAt(1, 61, 1, 72)}},
			// Report both nested calls, including calls inside a literal argument.
			{Code: "cb({ cause: callback(false) });", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 31), unexpectedLiteralAt(1, 13, 1, 28)}},
			// ---- Dimension 4: diagnostics cover the complete multiline call ----
			{Code: "callback(\n  false,\n  result\n);", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 4, 2)}},
			// UTF-16 ranges after an astral character and within the argument.
			{Code: "\"😀\"; cb(\"错误😀\");", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 7, 1, 17)}},
			// ---- Dimension 4: JSX containers still visit nested callback calls ----
			{Code: "const view = <Result value={cb(false)} />;", FileName: "input.tsx", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 29, 1, 38)}},
			// A rest binding does not hide a nested callback call.
			{Code: "const { ...rest } = cb(false);", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 21, 1, 30)}},
			// Escaped callback names use the decoded identifier, with ranges covering the source spelling.
			{Code: `c\u0062(false); call\u0062ack(0);`, Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 15), unexpectedLiteralAt(1, 17, 1, 33)}},
			// An optional generic call still has an identifier callee.
			{Code: "cb?.<Error>(false);", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 19)}},
			// Chained access or invocation does not report the same inner callback twice.
			{Code: "cb?.(false).x; (cb?.(0))(false); ((cb?.(null)))(false);", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 12), unexpectedLiteralAt(1, 17, 1, 24)}},
			// Nested JSDoc casts are transparent in JavaScript.
			{Code: "cb(/** @type {Error} */ (/** @satisfies {unknown} */ (0)));", FileName: "input.js", TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 59)}},
			// Overflowing numeric literals remain literals; unary expressions remain unknown.
			{Code: "cb(1e999); cb(-1e999); cb(0n);", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 10), unexpectedLiteralAt(1, 24, 1, 30)}},
			// Nested and destructuring assignments follow their final right operand.
			{Code: "cb((a = b = c = 0)); cb(([a] = [])); cb(({a} = {}));", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 20), unexpectedLiteralAt(1, 22, 1, 36), unexpectedLiteralAt(1, 38, 1, 52)}},
			// Documented difference: regex literals are rejected even when older Node.js versions cannot construct them.
			{Code: "cb(/(?i:error)/); cb(/(?<a>a)|(?<a>b)/);", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 17), unexpectedLiteralAt(1, 19, 1, 40)}},
			// Unicode set notation remains a non-error regex literal.
			{Code: `cb(/[\q{ab|cd}]/v);`, Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 19)}},
		},
	)
}
