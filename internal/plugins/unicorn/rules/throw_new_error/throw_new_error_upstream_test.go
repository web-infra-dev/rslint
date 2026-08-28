// TestThrowNewErrorUpstream migrates the valid/invalid suite from upstream
// test/throw-new-error.js 1:1. Position assertions cover line/column for every
// invalid case. rslint-specific lock-in cases live in
// throw_new_error_extras_test.go.
package throw_new_error_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/throw_new_error"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const messageID = "throw-new-error"

const messageDescription = "Use `new` when creating an error."

func TestThrowNewErrorUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&throw_new_error.ThrowNewErrorRule,
		[]rule_tester.ValidTestCase{
			{Code: `throw new Error()`},
			{Code: `new Error()`},
			{Code: `throw new TypeError()`},
			{Code: `throw new EvalError()`},
			{Code: `throw new RangeError()`},
			{Code: `throw new ReferenceError()`},
			{Code: `throw new SyntaxError()`},
			{Code: `throw new URIError()`},
			{Code: `throw new CustomError()`},
			{Code: `throw new FooBarBazError()`},
			{Code: `throw new ABCError()`},

			// Not `FooError` like.
			{Code: `throw getError()`},
			// Not a call expression.
			{Code: `throw CustomError`},
			// Callee is neither an identifier nor a static member.
			{Code: `throw getErrorConstructor()()`},
			// `MemberExpression.computed`.
			{Code: `throw lib[Error]()`},
			{Code: `throw lib["Error"]()`},
			// `new` cannot be applied to an optional-chained call.
			{Code: `throw Error?.()`},
			{Code: `throw lib?.Error()`},
			{Code: `throw lib?.foo.Error()`},
			// Optional chain in the callee even when the call is not the
			// chain's outermost element.
			{Code: `throw lib?.foo.Error().message`},
			// Not `FooError` like.
			{Code: `throw lib.getError()`},
			// https://github.com/sindresorhus/eslint-plugin-unicorn/issues/2654
			// `Data.TaggedError` is a factory, not a constructor.
			{Code: `class QueryError extends Data.TaggedError('QueryError') {}`},
			validIn("file.ts", `function RegisterServiceError() {
	return function <T extends new (...arguments_: any[]) => Error>(constructor: T) {
		return constructor;
	};
}

@RegisterServiceError()
export class SomeError extends Error {}`),
			validIn("file.ts", `@decorators.RegisterServiceError()
export class SomeError extends Error {}`),
			validIn("file.ts", `class Service {
	@OnQueueError()
	handle() {}
}`),
		},
		[]rule_tester.InvalidTestCase{
			invalid(`throw Error()`, `Error()`),
			invalid(`throw (Error)()`, `(Error)()`),
			invalid(`throw lib.Error()`, `lib.Error()`),
			invalid(`throw lib.mod.Error()`, `lib.mod.Error()`),
			invalid(`throw lib[mod].Error()`, `lib[mod].Error()`),
			invalid(`throw (lib.mod).Error()`, `(lib.mod).Error()`),
			invalid(`throw Error('foo')`, `Error('foo')`),
			invalid(`throw CustomError('foo')`, `CustomError('foo')`),
			invalid(`throw FooBarBazError('foo')`, `FooBarBazError('foo')`),
			invalid(`throw ABCError('foo')`, `ABCError('foo')`),
			invalid(`throw Abc3Error('foo')`, `Abc3Error('foo')`),
			invalid(`throw TypeError()`, `TypeError()`),
			invalid(`throw EvalError()`, `EvalError()`),
			invalid(`throw RangeError()`, `RangeError()`),
			invalid(`throw ReferenceError()`, `ReferenceError()`),
			invalid(`throw SyntaxError()`, `SyntaxError()`),
			invalid(`throw URIError()`, `URIError()`),
			invalid(`throw (( URIError() ))`, `URIError()`),
			invalid(`throw (( URIError ))()`, `(( URIError ))()`),
			invalid(`throw getGlobalThis().Error()`, `getGlobalThis().Error()`),
			invalid(`throw utils.getGlobalThis().Error()`, `utils.getGlobalThis().Error()`),
			invalid(`throw (( getGlobalThis().Error ))()`, `(( getGlobalThis().Error ))()`),
			// The rule watches every call expression, not only `throw`.
			invalid(`const error = Error()`, `Error()`),
			invalid(`throw Object.assign(Error(), {foo})`, `Error()`),
			invalid(
				"new Promise((resolve, reject) => {\n\treject(Error('message'));\n});",
				`Error('message')`,
			),
			invalid(
				"function foo() {\n\treturn[globalThis][0].Error('message');\n}",
				`[globalThis][0].Error('message')`,
			),
			invalidIn("file.ts", `@Decorator(Error())
export class SomeError extends Error {}`, `Error()`),
		},
	)
}

func validIn(fileName string, code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: fileName}
}

func invalidIn(fileName string, code string, target string) rule_tester.InvalidTestCase {
	testCase := invalid(code, target)
	testCase.FileName = fileName
	return testCase
}

// invalid reports the position of the single call expression named by target.
func invalid(code string, target string) rule_tester.InvalidTestCase {
	offset := strings.Index(code, target)
	if offset < 0 {
		panic("target not found in throw-new-error test: " + target)
	}

	line, column := lineColumn(code, offset)
	endLine, endColumn := lineColumn(code, offset+len(target))

	return rule_tester.InvalidTestCase{
		Code: code,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: messageID,
				Message:   messageDescription,
				Line:      line,
				Column:    column,
				EndLine:   endLine,
				EndColumn: endColumn,
			},
		},
	}
}

func lineColumn(code string, offset int) (int, int) {
	line := 1
	column := 1
	for index, character := range code {
		if index >= offset {
			break
		}
		if character == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return line, column
}
