// TestNoErrorEqualExtras covers subject shapes and type forms beyond the
// upstream suite. Every expectation matches eslint-plugin-jest@v29.16.6 run
// with @typescript-eslint/parser on the same input and compiler options.
package no_error_equal_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_error_equal"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoErrorEqualExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_error_equal.NoErrorEqualRule,
		[]rule_tester.ValidTestCase{
			// A missing expect argument has no type to inspect.
			{Code: `expect().toEqual(new Error());`},
			// The head must be called directly, not through a member.
			{Code: `declare const e: Error;
expect.extend(e).toEqual(1);`},
			// A matcher that is never called is not an assertion.
			{Code: `declare const e: Error;
expect(e).toEqual;`},
			// An invalid modifier order is not a parsed expect call.
			{Code: `declare const e: Error;
expect(e).not.resolves.toEqual(1);`},
			// Only toEqual and toStrictEqual are checked.
			{Code: `declare const e: Error;
expect(e).toMatchObject({});`},
			{Code: `declare const e: Error;
expect(e).toThrowError(e);`},
			{Code: `declare const e: Error;
expect(e).toBeError(1);`},
			// The subject is a Promise, not an Error, even with resolves.
			{Code: `declare const p: Promise<Error>;
expect(p).resolves.toEqual(new Error());`},
			// Containers of errors are not errors.
			{Code: `declare const es: Error[];
expect(es).toEqual([]);`},
			{Code: `declare const e: ErrorConstructor;
expect(e).toEqual(1);`},
			// Types without an Error symbol: unknown, any, never, catch bindings.
			{Code: `declare const e: unknown;
expect(e).toEqual(1);`},
			{Code: `declare const e: any;
expect(e).toEqual(1);`},
			{Code: `declare const e: never;
expect(e).toEqual(1);`},
			{Code: `try {} catch (e) { expect(e).toEqual(1); }`},
			// Every union member must be an Error.
			{Code: `declare const e: Error | undefined;
expect(e).toEqual(1);`},
			// An unconstrained type parameter is not an Error.
			{Code: `function f<T>(v: T) { expect(v).toEqual(1); }`},
			// implements does not make a class inherit from Error.
			{Code: `class Error2 implements Error { name = ''; message = ''; }
expect(new Error2()).toEqual(1);`},
			// A module-local class named Error is not the built-in Error.
			{Code: `export {};
class Error {}
expect(new Error()).toEqual(1);`},
			// A local expect is not Jest's expect.
			{Code: `const expect = (x: unknown) => ({ toEqual(y: unknown) {} });
expect(new Error()).toEqual(1);`},
		},
		[]rule_tester.InvalidTestCase{
			// Parentheses around the head, the argument or the head call.
			{
				Code: `declare const e: Error;
(expect)(e).toEqual(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 2, Column: 1, EndLine: 2, EndColumn: 23},
				},
			},
			{
				Code: `declare const e: Error;
expect((e)).toEqual(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 2, Column: 1, EndLine: 2, EndColumn: 23},
				},
			},
			{
				Code: `declare const e: Error;
(expect(e)).toEqual(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 2, Column: 1, EndLine: 2, EndColumn: 23},
				},
			},
			// Computed and optional matcher access.
			{
				Code: `declare const e: Error;
expect(e)['toEqual'](1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 2, Column: 1, EndLine: 2, EndColumn: 24},
				},
			},
			{
				Code: "declare const e: Error;\nexpect(e)[`toStrictEqual`](1);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 2, Column: 1, EndLine: 2, EndColumn: 30},
				},
			},
			{
				Code: `declare const e: Error;
expect(e).toEqual?.(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 2, Column: 1, EndLine: 2, EndColumn: 23},
				},
			},
			{
				Code: `declare const e: Error;
expect(e)?.toEqual(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 2, Column: 1, EndLine: 2, EndColumn: 22},
				},
			},
			// Modifiers do not change the inspected subject type.
			{
				Code: `declare const e: Error;
expect(e).resolves.not.toEqual(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 2, Column: 1, EndLine: 2, EndColumn: 34},
				},
			},
			// A spread argument is typed as its element.
			{
				Code: `declare const errors: Error[];
expect(...errors).toEqual(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 2, Column: 1, EndLine: 2, EndColumn: 29},
				},
			},
			// TypeScript expression wrappers and type arguments.
			{
				Code: `declare const e: Error;
expect(e!).toEqual(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 2, Column: 1, EndLine: 2, EndColumn: 22},
				},
			},
			{
				Code: `declare const e: Error;
expect(e satisfies Error).toEqual(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 2, Column: 1, EndLine: 2, EndColumn: 37},
				},
			},
			{
				Code: `declare const e: Error;
expect<Error>(e).toEqual(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 2, Column: 1, EndLine: 2, EndColumn: 28},
				},
			},
			{
				Code: `declare const e: Error;
expect(e, 'message').toStrictEqual(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 2, Column: 1, EndLine: 2, EndColumn: 38},
				},
			},
			// Unions of errors, intersections and aliases.
			{
				Code: `declare const e: Error | TypeError;
expect(e).toEqual(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 2, Column: 1, EndLine: 2, EndColumn: 21},
				},
			},
			{
				Code: `declare const e: TypeError | RangeError | AggregateError;
expect(e).toEqual(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 2, Column: 1, EndLine: 2, EndColumn: 21},
				},
			},
			{
				Code: `declare const e: Error & { code: string };
expect(e).toEqual(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 2, Column: 1, EndLine: 2, EndColumn: 21},
				},
			},
			{
				Code: `type E = Error;
declare const e: E;
expect(e).toEqual(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 3, Column: 1, EndLine: 3, EndColumn: 21},
				},
			},
			{
				Code: `declare const e: readonly Error[];
expect(e[0]).toEqual(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 2, Column: 1, EndLine: 2, EndColumn: 24},
				},
			},
			// A type parameter constrained through another type parameter.
			{
				Code: `function f<U extends Error, T extends U>(v: T) { expect(v).toEqual(1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 1, Column: 50, EndLine: 1, EndColumn: 70},
				},
			},
			// Interfaces extending Error, including DOMException from lib.dom.
			{
				Code: `interface E2 extends Error {}
declare const e: E2;
expect(e).toEqual(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 3, Column: 1, EndLine: 3, EndColumn: 21},
				},
			},
			{
				Code: `declare const e: DOMException;
expect(e).toEqual(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 2, Column: 1, EndLine: 2, EndColumn: 21},
				},
			},
			{
				Code: `export {};
class Error extends globalThis.Error {}
expect(new Error()).toEqual(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 3, Column: 1, EndLine: 3, EndColumn: 31},
				},
			},
			// In a script, class Error merges with the global Error declaration.
			{
				Code: `class Error {}
expect(new Error()).toEqual(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 2, Column: 1, EndLine: 2, EndColumn: 31},
				},
			},
			// Global augmentation keeps the default-library declaration.
			{
				Code: `declare global { interface Error { code?: string } }
declare const e: Error;
expect(e).toEqual(1);
export {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 3, Column: 1, EndLine: 3, EndColumn: 21},
				},
			},
			// Narrowing is respected.
			{
				Code: `try {} catch (e) { if (e instanceof Error) { expect(e).toEqual(1); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 1, Column: 46, EndLine: 1, EndColumn: 66},
				},
			},
			{
				Code: `declare const e: Error;
it('x', () => expect(e).toStrictEqual(1));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 2, Column: 15, EndLine: 2, EndColumn: 41},
				},
			},
			// Aliased import from @jest/globals.
			{
				Code: `import { expect as e } from '@jest/globals';
e(new Error()).toEqual(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 2, Column: 1, EndLine: 2, EndColumn: 26},
				},
			},
			// A multiline report spans the whole matcher call.
			{
				Code: `declare const e: Error;
expect(
  e
).toEqual(
  1
);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 2, Column: 1, EndLine: 6, EndColumn: 2},
				},
			},
			{
				// A matcher chained on another matcher's result parses twice.
				Code: "declare const e: Error;\nexpect(e).toEqual(1).toEqual(2);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 2, Column: 1, EndLine: 2, EndColumn: 32},
					{MessageId: "equalError", Message: msg, Line: 2, Column: 1, EndLine: 2, EndColumn: 21},
				},
			},
		},
	)
}
