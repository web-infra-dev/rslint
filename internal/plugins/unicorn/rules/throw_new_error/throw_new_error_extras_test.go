// TestThrowNewErrorExtras locks in branches and edge shapes that the upstream
// test suite does not exercise. Each group names the AST boundary it protects;
// the full upstream suite lives in throw_new_error_upstream_test.go.
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
			// ---- Dimension 4: parenthesized decorator calls ----
			// Parentheses are transparent between the call and Decorator, even
			// when tsgo preserves several explicit wrapper nodes.
			validIn("file.ts", `@(RegisterServiceError())
export class SomeError extends Error {}`),
			validIn("file.ts", `@(((RegisterServiceError())))
export class SomeError extends Error {}`),
			validIn("file.js", `@(/** @type {any} */ (RegisterServiceError()))
export class SomeError extends Error {}`),
			validIn("file.js", `@(/** @satisfies {Function} */ (RegisterServiceError()))
export class SomeError extends Error {}`),

			// ---- Dimension 4: optional chains behind transparent wrappers ----
			validIn("file.js", `throw (lib?.mod).Error()`),
			validIn("file.ts", `throw (lib?.mod as any).CustomError()`),
			validIn("file.ts", `throw (<any>lib?.mod).CustomError()`),
			validIn("file.ts", `throw (lib?.mod satisfies any).CustomError()`),
			validIn("file.ts", `throw (lib?.mod!).CustomError()`),
			validIn("file.ts", `throw ((((lib?.mod as any)!) satisfies any) as any).CustomError()`),
			validIn("file.ts", `throw (factory?.() as any).CustomError()`),
			validIn("file.ts", `throw ((lib?.mod as any)[key]).CustomError()`),
			validIn("file.ts", `throw (((lib?.mod))).CustomError()`),
			validIn("file.tsx", `<Component value={(lib?.mod as any).CustomError()} />`),

			// ---- Dimension 4: callee classification does not unwrap TS assertions ----
			validIn("file.ts", `throw (Error as any)()`),
			validIn("file.ts", `throw (lib.CustomError as any)()`),
			validIn("file.ts", `throw lib.#Error()`),

			// A bare error class is not a call.
			validIn("file.js", `class SomeError extends Error {}`),
			// `new` already applied on every spelling upstream accepts.
			validIn("file.js", `throw new lib.mod.CustomError('foo')`),
		},
		[]rule_tester.InvalidTestCase{
			// ESTree does not expose wrappers that tsgo synthesizes for JSDoc
			// casts, so they do not hide an otherwise reportable callee.
			invalidIn("file.js", `throw (/** @type {any} */ (Error))()`, `(/** @type {any} */ (Error))()`),
			invalidIn(
				"file.js",
				`throw (/** @satisfies {Function} */ (Error))()`,
				`(/** @satisfies {Function} */ (Error))()`,
			),
			invalidIn(
				"file.js",
				`throw (/** @type {any} */ (lib.CustomError))()`,
				`(/** @type {any} */ (lib.CustomError))()`,
			),

			// Locks in upstream create() decorator-parent branch: only
			// parentheses are transparent. Authored TS wrappers remain parents.
			invalidIn("file.ts", `@(RegisterServiceError() as any)
export class SomeError extends Error {}`, `RegisterServiceError()`),
			invalidIn("file.ts", `@(RegisterServiceError() satisfies any)
export class SomeError extends Error {}`, `RegisterServiceError()`),
			invalidIn("file.ts", `@(RegisterServiceError()!)
export class SomeError extends Error {}`, `RegisterServiceError()`),
			invalidIn("file.ts", `@(<any>RegisterServiceError())
export class SomeError extends Error {}`, `RegisterServiceError()`),

			// A nested call does not inherit the outer call's decorator exemption.
			invalidIn("file.ts", `@((OuterError(Error())))
export class SomeError extends Error {}`, `Error()`),

			// Locks in upstream hasOptionalChainElement() TS-wrapper branch:
			// wrappers without an optional link still report.
			invalidIn("file.ts", `throw (lib.mod as any).CustomError()`, `(lib.mod as any).CustomError()`),
			invalidIn("file.ts", `throw (<any>lib.mod).CustomError()`, `(<any>lib.mod).CustomError()`),
			invalidIn("file.ts", `throw (lib.mod satisfies any).CustomError()`, `(lib.mod satisfies any).CustomError()`),
			invalidIn("file.ts", `throw (lib.mod!).CustomError()`, `(lib.mod!).CustomError()`),
			invalidIn(
				"file.ts",
				`throw ((((lib.mod as any)!) satisfies any) as any).CustomError()`,
				`((((lib.mod as any)!) satisfies any) as any).CustomError()`,
			),
			invalidIn("file.ts", `throw (factory() as any).CustomError()`, `(factory() as any).CustomError()`),
			invalidIn("file.ts", `throw ((lib.mod as any)[key]).CustomError()`, `((lib.mod as any)[key]).CustomError()`),
			invalidIn("file.ts", `throw (((lib.mod))).CustomError()`, `(((lib.mod))).CustomError()`),
			invalidIn(
				"file.tsx",
				`<Component value={(lib.mod as any).CustomError()} />`,
				`(lib.mod as any).CustomError()`,
			),

			// ---- Adversarial: optional links outside the callee/object path ----
			invalidIn("file.ts", `throw maker(value?.member).CustomError()`, `maker(value?.member).CustomError()`),
			invalidIn("file.ts", `throw lib[value?.member].CustomError()`, `lib[value?.member].CustomError()`),
			invalidIn(
				"file.ts",
				`async function run() { throw (await lib?.mod).CustomError(); }`,
				`(await lib?.mod).CustomError()`,
			),
			invalidIn("file.ts", `throw (lib?.mod<T>).CustomError()`, `(lib?.mod<T>).CustomError()`),

			// Nested inside another call.
			invalidIn("file.js", `foo(Error())`, `Error()`),
			// The rule watches calls, not only `throw` arguments.
			invalidIn("file.js", `const error = lib.mod.CustomError()`, `lib.mod.CustomError()`),
		},
	)
}
