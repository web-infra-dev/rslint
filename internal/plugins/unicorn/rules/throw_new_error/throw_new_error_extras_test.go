// TestThrowNewErrorExtras covers cases the upstream JavaScript suite cannot
// express: decorator syntax, which needs a TypeScript file, and the rslint-side
// lock-in for callees tsgo represents but ESTree does not.
package throw_new_error_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/throw_new_error"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestThrowNewErrorExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&throw_new_error.ThrowNewErrorRule,
		[]rule_tester.ValidTestCase{
			// A decorator is a call expression whose name ends in `Error`, so
			// only the decorator check keeps these quiet.
			validIn("file.ts", `@RegisterServiceError()
export class SomeError extends Error {}`),
			validIn("file.ts", `@decorators.RegisterServiceError()
export class SomeError extends Error {}`),
			validIn("file.ts", `class Service {
	@OnQueueError()
	handle() {}
}`),
			validIn("file.ts", `function RegisterServiceError() {
	return function <T extends new (...arguments_: any[]) => Error>(constructor: T) {
		return constructor;
	};
}

@RegisterServiceError()
export class SomeError extends Error {}`),

			// The parenthesis ends tsgo's optional chain, but upstream keeps
			// walking past it and skips the call.
			validIn("file.js", `throw (lib?.mod).Error()`),

			// tsgo keeps assertion nodes the way a TypeScript parser sees them,
			// so the callee is not an identifier. Upstream reads the same
			// expression the same way.
			validIn("file.ts", `throw (Error as any)()`),
			validIn("file.ts", `throw lib.#Error()`),

			// A bare error class is not a call.
			validIn("file.js", `class SomeError extends Error {}`),
			// `new` already applied on every spelling upstream accepts.
			validIn("file.js", `throw new lib.mod.CustomError('foo')`),
		},
		[]rule_tester.InvalidTestCase{
			// The decorator itself is exempt, but an argument inside it is an
			// ordinary call expression.
			invalidIn("file.ts", `@Decorator(Error())
export class SomeError extends Error {}`, `Error()`),

			// Nested inside another call.
			invalidIn("file.js", `foo(Error())`, `Error()`),
			// The rule watches calls, not only `throw` arguments.
			invalidIn("file.js", `const error = lib.mod.CustomError()`, `lib.mod.CustomError()`),
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
