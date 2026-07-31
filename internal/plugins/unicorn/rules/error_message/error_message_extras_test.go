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
			tsValid("Error!('msg')"),
			tsValid("(Error as any)('msg')"),
			tsValid("(Error satisfies any)('msg')"),

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
		},
	)
}

func tsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{
		Code:     code,
		FileName: "file.ts",
	}
}
