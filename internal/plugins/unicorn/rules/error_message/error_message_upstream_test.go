// TestErrorMessageUpstream migrates the full valid/invalid suite from upstream
// test/error-message.js 1:1. Position assertions cover line/column for every
// invalid case. rslint-specific lock-in cases live in the
// error_message_extras_test.go file.
package error_message_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/error_message"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	messageIDMissing   = "missing-message"
	messageIDEmpty     = "message-is-empty-string"
	messageIDNotString = "message-is-not-a-string"

	msgMissingSuffix = "` constructor."
	msgEmpty         = "Error message should not be an empty string."
	msgNotString     = "Error message should be a string."
)

func msgMissing(name string) string {
	return "Pass a message to the `" + name + msgMissingSuffix
}

func TestErrorMessageUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&error_message.ErrorMessageRule,
		[]rule_tester.ValidTestCase{
			// ---- Core Builtin Errors ----
			jsValid("throw new Error('error')"),
			jsValid("throw new TypeError('error')"),
			jsValid("throw new MyCustomError('error')"),
			jsValid("throw new MyCustomError()"),
			jsValid("throw generateError()"),
			jsValid("throw foo()"),
			jsValid("throw err"),
			jsValid("throw 1"),
			jsValid("const err = TypeError('error');\nthrow err;"),
			jsValid("new Error(\"message\", 0, 0)"),
			jsValid("new Error(foo)"),
			jsValid("const errors = [];\nif (condition) {\n\terrors.push('hello');\n}\nif (errors.length) {\n\tthrow new Error(errors.join('\\n'));\n}"),
			jsValid("new Error(...foo)"),
			jsValid("/* global x */\nconst a = x;\nthrow x;"),
			jsValid("const Error = function () {};\nconst err = new Error({\n\tname: 'Unauthorized',\n});"),

			// ---- AggregateError ----
			jsValid("new AggregateError(errors, \"message\")"),
			jsValid("new NotAggregateError(errors)"),
			jsValid("new AggregateError(...foo)"),
			jsValid("new AggregateError(...foo, \"\")"),
			jsValid("new AggregateError(errors, ...foo)"),
			jsValid("new AggregateError(errors, message, \"\")"),
			jsValid("new AggregateError(\"\", message, \"\")"),

			// ---- SuppressedError ----
			jsValid("new SuppressedError(error, suppressed, \"message\")"),
			jsValid("new NotSuppressedError(error, suppressed)"),
			jsValid("new SuppressedError(...foo)"),
			jsValid("new SuppressedError(...foo, \"\")"),
			jsValid("new SuppressedError(error, suppressed, ...foo)"),
			jsValid("new SuppressedError(error, suppressed, message, \"\")"),
			jsValid("new SuppressedError(\"\", \"\", message, \"\")"),
		},
		[]rule_tester.InvalidTestCase{
			// ---- Core Builtin Errors ----
			invalid("throw new Error()", "new Error()", messageIDMissing, msgMissing("Error")),
			invalid("throw Error()", "Error()", messageIDMissing, msgMissing("Error")),
			invalid("throw new Error('')", "''", messageIDEmpty, msgEmpty),
			invalid("throw new Error(``)", "``", messageIDEmpty, msgEmpty),
			invalid("const err = new Error();\nthrow err;", "new Error()", messageIDMissing, msgMissing("Error")),
			invalid("let err = 1;\nerr = new Error();\nthrow err;", "new Error()", messageIDMissing, msgMissing("Error")),
			invalid("let err = new Error();\nerr = 1;\nthrow err;", "new Error()", messageIDMissing, msgMissing("Error")),
			invalid("const foo = new TypeError()", "new TypeError()", messageIDMissing, msgMissing("TypeError")),
			invalid("const foo = new SyntaxError()", "new SyntaxError()", messageIDMissing, msgMissing("SyntaxError")),
			invalid("const errorMessage = Object.freeze({errorMessage: 1}).errorMessage;\nthrow new Error(errorMessage)", "errorMessage", messageIDNotString, msgNotString, 4),
			invalid("throw new Error([])", "[]", messageIDNotString, msgNotString),
			invalid("throw new Error([foo])", "[foo]", messageIDNotString, msgNotString),
			invalid("throw new Error([0][0])", "[0][0]", messageIDNotString, msgNotString),
			invalid("throw new Error({})", "{}", messageIDNotString, msgNotString),
			invalid("throw new Error({foo})", "{foo}", messageIDNotString, msgNotString),
			invalid("throw new Error({foo: 0}.foo)", "{foo: 0}.foo", messageIDNotString, msgNotString),
			invalid("throw new Error(lineNumber=2)", "lineNumber=2", messageIDNotString, msgNotString),
			invalid("throw new Error(false)", "false", messageIDNotString, msgNotString),
			invalid("const error = new RangeError;", "new RangeError", messageIDMissing, msgMissing("RangeError")),
			invalid("throw Object.assign(new Error(), {foo})", "new Error()", messageIDMissing, msgMissing("Error")),

			// ---- AggregateError ----
			invalid("new AggregateError(errors)", "new AggregateError(errors)", messageIDMissing, msgMissing("AggregateError")),
			invalid("AggregateError(errors)", "AggregateError(errors)", messageIDMissing, msgMissing("AggregateError")),
			invalid("new AggregateError(errors, \"\")", "\"\"", messageIDEmpty, msgEmpty),
			invalid("new AggregateError(errors, ``)", "``", messageIDEmpty, msgEmpty),
			invalid("new AggregateError(errors, \"\", extraArgument)", "\"\"", messageIDEmpty, msgEmpty),
			invalid("const errorMessage = Object.freeze({errorMessage: 1}).errorMessage;\nthrow new AggregateError(errors, errorMessage)", "errorMessage", messageIDNotString, msgNotString, 4),
			invalid("new AggregateError(errors, [])", "[]", messageIDNotString, msgNotString),
			invalid("new AggregateError(errors, [foo])", "[foo]", messageIDNotString, msgNotString),
			invalid("new AggregateError(errors, [0][0])", "[0][0]", messageIDNotString, msgNotString),
			invalid("new AggregateError(errors, {})", "{}", messageIDNotString, msgNotString),
			invalid("new AggregateError(errors, {foo})", "{foo}", messageIDNotString, msgNotString),
			invalid("new AggregateError(errors, {foo: 0}.foo)", "{foo: 0}.foo", messageIDNotString, msgNotString),
			invalid("new AggregateError(errors, lineNumber=2)", "lineNumber=2", messageIDNotString, msgNotString),
			invalid("const error = new AggregateError;", "new AggregateError", messageIDMissing, msgMissing("AggregateError")),

			// ---- SuppressedError ----
			invalid("new SuppressedError(error, suppressed,)", "new SuppressedError(error, suppressed,)", messageIDMissing, msgMissing("SuppressedError")),
			invalid("new SuppressedError(error,)", "new SuppressedError(error,)", messageIDMissing, msgMissing("SuppressedError")),
			invalid("new SuppressedError()", "new SuppressedError()", messageIDMissing, msgMissing("SuppressedError")),
			invalid("SuppressedError(error, suppressed,)", "SuppressedError(error, suppressed,)", messageIDMissing, msgMissing("SuppressedError")),
			invalid("SuppressedError(error,)", "SuppressedError(error,)", messageIDMissing, msgMissing("SuppressedError")),
			invalid("SuppressedError()", "SuppressedError()", messageIDMissing, msgMissing("SuppressedError")),
			invalid("new SuppressedError(error, suppressed, \"\")", "\"\"", messageIDEmpty, msgEmpty),
			invalid("new SuppressedError(error, suppressed, ``)", "``", messageIDEmpty, msgEmpty),
			invalid("new SuppressedError(error, suppressed, \"\", options)", "\"\"", messageIDEmpty, msgEmpty),
			invalid("const errorMessage = Object.freeze({errorMessage: 1}).errorMessage;\nthrow new SuppressedError(error, suppressed, errorMessage)", "errorMessage", messageIDNotString, msgNotString, 4),
			invalid("new SuppressedError(error, suppressed, [])", "[]", messageIDNotString, msgNotString),
			invalid("new SuppressedError(error, suppressed, [foo])", "[foo]", messageIDNotString, msgNotString),
			invalid("new SuppressedError(error, suppressed, [0][0])", "[0][0]", messageIDNotString, msgNotString),
			invalid("new SuppressedError(error, suppressed, {})", "{}", messageIDNotString, msgNotString),
			invalid("new SuppressedError(error, suppressed, {foo})", "{foo}", messageIDNotString, msgNotString),
			invalid("new SuppressedError(error, suppressed, {foo: 0}.foo)", "{foo: 0}.foo", messageIDNotString, msgNotString),
			invalid("new SuppressedError(error, suppressed, lineNumber=2)", "lineNumber=2", messageIDNotString, msgNotString),
			invalid("const error = new SuppressedError;", "new SuppressedError", messageIDMissing, msgMissing("SuppressedError")),
		},
	)
}

func jsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{
		Code:     code,
		FileName: "file.js",
	}
}

func invalid(code string, target string, messageID string, message string, occurrence ...int) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.js",
		Errors: []rule_tester.InvalidTestCaseError{
			expectedError(code, target, messageID, message, occurrence...),
		},
	}
}

func expectedError(code string, target string, messageID string, message string, occurrence ...int) rule_tester.InvalidTestCaseError {
	nth := 1
	if len(occurrence) > 0 {
		nth = occurrence[0]
	}

	offset := nthIndex(code, target, nth)
	if offset < 0 {
		panic("target not found in error-message test: " + target + " in " + code)
	}

	line, column := lineColumnForOffset(code, offset)
	endLine, endColumn := lineColumnForOffset(code, offset+len(target))
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID,
		Message:   message,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func nthIndex(s string, target string, nth int) int {
	searchStart := 0
	for i := range nth {
		offset := strings.Index(s[searchStart:], target)
		if offset < 0 {
			return -1
		}
		searchStart += offset
		if i == nth-1 {
			return searchStart
		}
		searchStart += len(target)
	}
	return -1
}

func lineColumnForOffset(code string, offset int) (int, int) {
	line := 1
	column := 1
	for i := range offset {
		if code[i] == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return line, column
}
