// TestErrorMessageExtras locks in branches and edge shapes that the upstream test suite doesn't exercise.
// Each case carries an inline comment pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers,
// so future refactors can't silently regress them without breaking a named lock-in.
package error_message_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/error_message"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestErrorMessageExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&error_message.ErrorMessageRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: Optional chain calls are skipped ----
			jsValid("Error?.('msg')"),
			jsValid("AggregateError?.(errors, 'msg')"),

			// ---- Dimension 4: Shadowed global Error ----
			jsValid("function test(Error) { new Error(); }"),
			jsValid("const TypeError = 123; new TypeError();"),
			jsValid("import { Error } from 'my-module'; new Error();"),

			// ---- Dimension 4: Receiver wrappers / non-null assertion / satisfies ----
			// A wrapped callee is not an identifier, so the call is not checked
			// at all — asserted with no message argument, which is what would be
			// reported if the wrapper were skipped.
			tsValid("Error!()"),
			tsValid("(Error as any)()"),
			tsValid("(Error satisfies any)()"),

			// ---- Dimension 4: Spread element before message index is skipped ----
			jsValid("new AggregateError(...errors, 'msg')"),
			jsValid("new SuppressedError(error, ...suppressed, 'msg')"),

			// ---- Locks in upstream branch: Callee is not an identifier ----
			jsValid("globalThis.Error()"),
			jsValid("obj.Error()"),

			// ---- Locks in upstream branch: Callee identifier does not match builtin errors ----
			jsValid("new MyCustomError()"),
			jsValid("new CustomTypeError()"),

			// ---- Locks in upstream branch: Unknown dynamic values are allowed ----
			jsValid("new Error(getMessage())"),
			jsValid("new Error(arg1 + arg2)"),
			jsValid("new Error(errors.join('\\n'))"),

			// ---- Locks in upstream branch: Node kinds the static evaluator gives up on ----
			jsValid("new Error(() => {})"),
			jsValid("new Error(function () {})"),
			jsValid("new Error(class {})"),
			jsValid("new Error(Symbol('msg'))"),

			// ---- Locks in upstream branch: A reassigned binding has no static value ----
			jsValid("let msg = 1;\nmsg = 2;\nnew Error(msg);"),

			// ---- Locks in upstream branch: A known type is not a known value ----
			tsValid("declare const msg: number;\nnew Error(msg);"),
			tsValid("function run(msg: number) { throw new Error(msg); }"),
			tsValid("declare const msg: {a: 1};\nnew Error(msg);"),
			tsValid("declare const msg: Date;\nnew Error(msg);"),
			tsValid("enum Code { A = 1 }\nnew Error(Code.A);"),
			tsValid("function run(msg: unknown) { if (typeof msg === 'number') throw new Error(msg); }"),

			// ---- Locks in upstream branch: Shapes that stop static folding ----
			jsValid("new Error({get msg() { return ''; }}.msg)"),
			jsValid("new Error({...foo}.msg)"),
			jsValid("new Error(msg += '')"),
			jsValid("new Error(['a', 'b'].join(''))"),
			jsValid("let Object = {freeze: x => x};\nnew Error(Object.freeze({msg: ''}).msg)"),
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: Parenthesized receiver ----
			invalid("throw (Error)('')", "''", messageIDEmpty, msgEmpty),
			invalid("throw ((TypeError))('')", "''", messageIDEmpty, msgEmpty),

			// ---- Dimension 4: Nesting under class/functions ----
			invalid("function run() { throw new Error(); }", "new Error()", messageIDMissing, msgMissing("Error")),
			invalid("class A { constructor() { throw new Error(); } }", "new Error()", messageIDMissing, msgMissing("Error")),

			// ---- Locks in upstream branch: Static evaluator returns non-string (BigInt, null) ----
			invalid("throw new Error(123n)", "123n", messageIDNotString, msgNotString),
			invalid("throw new Error(null)", "null", messageIDNotString, msgNotString),
			invalid("throw new Error(/msg/)", "/msg/", messageIDNotString, msgNotString),
			invalid("throw new Error(undefined)", "undefined", messageIDNotString, msgNotString),

			// ---- Interfaces and type aliases carry no value, so they do not shadow ----
			tsInvalid("interface Error { code: string }\nnew Error();", "new Error()", messageIDMissing, msgMissing("Error")),
			tsInvalid("type Error = string;\nnew Error();", "new Error()", messageIDMissing, msgMissing("Error")),

			// ---- Static values folded through member access and pass-through calls ----
			invalid("throw new Error(Object.freeze({msg: ''}).msg)", "Object.freeze({msg: ''}).msg", messageIDEmpty, msgEmpty),
			invalid("const msg = {a: {b: 1}};\nthrow new Error(msg.a.b)", "msg.a.b", messageIDNotString, msgNotString),
			invalid("throw new Error(['', 'msg'][0])", "['', 'msg'][0]", messageIDEmpty, msgEmpty),
			invalid("throw new Error(msg = 1)", "msg = 1", messageIDNotString, msgNotString),
			invalid("throw new Error({['msg']: ''}.msg)", "{['msg']: ''}.msg", messageIDEmpty, msgEmpty),
			invalid("throw new Error({0: ''}[0])", "{0: ''}[0]", messageIDEmpty, msgEmpty),
			invalid("throw new Error({1e3: ''}[1000])", "{1e3: ''}[1000]", messageIDEmpty, msgEmpty),
			invalid("throw new Error(Object.seal({msg: ''}).msg)", "Object.seal({msg: ''}).msg", messageIDEmpty, msgEmpty),
			invalid("throw new Error(Object.preventExtensions({msg: 1}).msg)", "Object.preventExtensions({msg: 1}).msg", messageIDNotString, msgNotString),

			// ---- Reading a key that is not there folds to `undefined` ----
			invalid("throw new Error({}.msg)", "{}.msg", messageIDNotString, msgNotString),
			invalid("throw new Error([][0])", "[][0]", messageIDNotString, msgNotString),
			invalid("throw new Error(['msg'][1])", "['msg'][1]", messageIDNotString, msgNotString),
			invalid("throw new Error([, 'msg'][0])", "[, 'msg'][0]", messageIDNotString, msgNotString),
		},
	)
}

func tsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{
		Code:     code,
		FileName: "file.ts",
	}
}

func tsInvalid(code string, target string, messageID string, message string, occurrence ...int) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.ts",
		Errors: []rule_tester.InvalidTestCaseError{
			expectedError(code, target, messageID, message, occurrence...),
		},
	}
}
