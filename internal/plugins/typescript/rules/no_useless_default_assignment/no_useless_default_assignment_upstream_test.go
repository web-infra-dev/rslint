// TestNoUselessDefaultAssignmentUpstream migrates the full valid/invalid
// suite from upstream typescript-eslint's
//
//	packages/eslint-plugin/tests/rules/no-useless-default-assignment.test.ts
//
// 1:1. Code blocks and indentation are preserved verbatim from upstream so
// the line/column assertions on every invalid case can be copied unchanged.
// rslint-specific lock-in cases (Dimension 4 edge shapes, branch lock-ins,
// real-user issue shapes) live in
// no_useless_default_assignment_extras_test.go.
package no_useless_default_assignment

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUselessDefaultAssignmentUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUselessDefaultAssignmentRule,
		[]rule_tester.ValidTestCase{
			{Code: `
        function Bar({ foo = '' }: { foo?: string }) {
          return foo;
        }
      `},
			{Code: `
        const { foo } = { foo: 'bar' };
      `},
			{Code: `
        [1, 2, 3, undefined].map((a = 42) => a + 1);
      `},
			{Code: `
        function test(a?: number) {
          return a;
        }
      `},
			{Code: `
        const obj: { a?: string } = {};
        const { a = 'default' } = obj;
      `},
			{Code: `
        function test(a: string | undefined = 'default') {
          return a;
        }
      `},
			{Code: `
        (a: string = 'default') => a;
      `},
			{Code: `
        function test(a: string = 'default') {
          return a;
        }
      `},
			{Code: `
        class C {
          public test(a: string = 'default') {
            return a;
          }
        }
      `},
			{Code: `
        const obj: { a: string | undefined } = { a: undefined };
        const { a = 'default' } = obj;
      `},
			{Code: `
        function test(arr: number[] | undefined = []) {
          return arr;
        }
      `},
			{Code: `
        function Bar({ nested: { foo = '' } = {} }: { nested?: { foo?: string } }) {
          return foo;
        }
      `},
			{Code: `
        function test(a: any = 'default') {
          return a;
        }
      `},
			{Code: `
        function test(a: unknown = 'default') {
          return a;
        }
      `},
			{Code: `
        function test(a = 5) {
          return a;
        }
      `},
			{Code: `
        function createValidator(): () => void {
          return (param = 5) => {};
        }
      `},
			{Code: `
        function Bar({ foo = '' }: { foo: any }) {
          return foo;
        }
      `},
			{Code: `
        function Bar({ foo = '' }: { foo: unknown }) {
          return foo;
        }
      `},
			{Code: `
        function getValue(): undefined;
        function getValue(box: { value: string }): string;
        function getValue({ value = '' }: { value?: string } = {}): string | undefined {
          return value;
        }
      `},
			{Code: `
        function getValueObject({ value = '' }: Partial<{ value: string }>) {
          return value;
        }
      `},
			{Code: `
        const { value = 'default' } = someUnknownFunction();
      `},
			{Code: `
        const [value = 'default'] = someUnknownFunction();
      `},
			{Code: `
        for (const { value = 'default' } of []) {
        }
      `},
			{Code: `
        for (const [value = 'default'] of []) {
        }
      `},
			{Code: `
        declare const x: [[number | undefined]];
        const [[a = 1]] = x;
      `},
			{Code: `
        function foo(x: string = '') {}
      `},
			{Code: `
        class C {
          method(x: string = '') {}
        }
      `},
			{Code: `
        const foo = (x: string = '') => {};
      `},
			{Code: `
        const obj = { ab: { x: 1 } };
        const {
          ['a' + 'b']: { x = 1 },
        } = obj;
      `},
			{Code: `
        const obj = { ab: 1 };
        const { ['a' + 'b']: x = 1 } = obj;
      `},
			{Code: `
        for ([[a = 1]] of []) {
        }
      `},
			{
				Code: `
        declare const g: Array<string>;
        const [foo = ''] = g;
      `,
				TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
			},
			{
				Code: `
        declare const g: Record<string, string>;
        const { foo = '' } = g;
      `,
				TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
			},
			{
				Code: `
        declare const h: { [key: string]: string };
        const { bar = '' } = h;
      `,
				TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
			},
			{Code: `
        declare const g: Array<string>;
        const [foo = ''] = g;
      `},
			{Code: `
        declare const g: Record<string, string>;
        const { foo = '' } = g;
      `},
			{Code: `
        declare const h: { [key: string]: string };
        const { bar = '' } = h;
      `},
			// https://github.com/typescript-eslint/typescript-eslint/issues/11849
			{Code: `
        type Merge = boolean | ((incoming: string[]) => void);

        const policy: { merge: Merge } = {
          merge: (incoming: string[] = []) => {
            incoming;
          },
        };
      `},
			// https://github.com/typescript-eslint/typescript-eslint/issues/11846
			{Code: `
        const [a, b = ''] = 'hello.world'.split('.');
      `},
			// https://github.com/typescript-eslint/typescript-eslint/issues/11846
			{Code: `
        declare const params: string[];
        const [c = '123'] = params;
      `},
			// https://github.com/typescript-eslint/typescript-eslint/issues/11846
			{Code: `
        declare function useCallback<T>(callback: T);
        useCallback((value: number[] = []) => {});
      `},
			{Code: `
        declare const tuple: [string];
        const [a, b = 'default'] = tuple;
      `},
			// https://github.com/typescript-eslint/typescript-eslint/issues/11911
			{Code: `
        const run = (cb: (...args: unknown[]) => void) => cb();
        const cb = (p: boolean = true) => null;
        run(cb);
        run((p: boolean = true) => null);
      `},
			// https://github.com/typescript-eslint/typescript-eslint/issues/11850
			{Code: `
        const { a = 'default' } = Math.random() > 0.5 ? { a: 'Hello' } : {};
      `},
			// https://github.com/typescript-eslint/typescript-eslint/issues/11850
			{Code: `
        const { a = 'default' } =
          Math.random() > 0.5 ? (Math.random() > 0.5 ? { a: 'Hello' } : {}) : {};
      `},
			// Optional parameter with meaningful default value
			{Code: `
        function findPosts({
          category,
          maxResults = 100,
        }: {
          category: string;
          maxResults?: number;
        }): Promise<string[]> {
          return Promise.resolve([category, String(maxResults)]);
        }
      `},
			// https://github.com/typescript-eslint/typescript-eslint/issues/11980
			{Code: `
        const { a = 'baz' } = cond ? {} : { a: 'bar' };
      `},
			{Code: `
        const { a = 'baz' } = cond ? foo : { a: 'bar' };
      `},
			{Code: `
        const { a = 'baz' } = foo && { a: 'bar' };
      `},
			{Code: `
        const { a = 'baz' } = cond ? { a: 'foo', ...extra } : { a: 'bar' };
      `},
			{Code: `
        const { a = 'baz' } = cond ? { ...foo } : { a: 'bar' };
      `},
			{Code: `
        const key = Math.random() > 0.5 ? 'a' : 'b';
        const { a = 'baz' } = cond ? { [key]: 'foo' } : { [key]: 'bar' };
      `},
			{Code: `
        const { a = 'baz' } = cond ? foo && { a: 'bar' } : { a: 'baz' };
      `},
			{Code: `
        const obj: unknown = { a: 'bar' };
        const { a = 'baz' } = cond ? obj : { a: 'bar' };
      `},
			{Code: `
        const sym = Symbol('a');
        const { a = 'baz' } = cond ? { [sym]: 'foo' } : { [sym]: 'bar' };
      `},
			{Code: "\n        const { a = 'baz' } = cond ? { [`a${1}`]: 'foo' } : { a: 'bar' };\n      "},
			// https://github.com/typescript-eslint/typescript-eslint/issues/11948
			{Code: `
        class AbstractEntity {
          public a: string | undefined;
          public static fromJson<T extends { a: string }>(
            this: new () => T,
            { inner = { a: 'test' } }: { inner?: { a: string } },
          ): T {
            const entity = new this();
            entity.a = inner?.a;
            return entity;
          }
        }
      `},
			{Code: `
        type FetchFn<TParams> =
          Partial<TParams> extends TParams
            ? (params?: TParams) => void
            : (params: TParams) => void;

        function createFetcher<TParams>() {
          type Params = TParams;

          const fn: FetchFn<TParams> = (
            params: Partial<Params> = {} as Partial<Params>,
          ) => {
            console.log(params);
          };

          return fn;
        }
      `},
			{Code: `
        interface Foos {
          bar?: number;
        }
        const foos: Foos[] = [];
        foos.flatMap(({ bar = 42 }) => bar);
      `},
			{Code: `
        function f(this: void, { bar = 42 }: { bar?: number }) {
          return bar;
        }
      `},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
        function Bar({ foo = '' }: { foo: string }) {
          return foo;
        }
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      2,
						Column:    30,
						EndColumn: 32,
					},
				},
				Output: []string{`
        function Bar({ foo }: { foo: string }) {
          return foo;
        }
      `},
			},
			{
				Code: `
        class C {
          public method({ foo = '' }: { foo: string }) {
            return foo;
          }
        }
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      3,
						Column:    33,
						EndColumn: 35,
					},
				},
				Output: []string{`
        class C {
          public method({ foo }: { foo: string }) {
            return foo;
          }
        }
      `},
			},
			{
				Code: `
        const { 'literal-key': literalKey = 'default' } = { 'literal-key': 'value' };
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      2,
						Column:    45,
						EndColumn: 54,
					},
				},
				Output: []string{`
        const { 'literal-key': literalKey } = { 'literal-key': 'value' };
      `},
			},
			{
				Code: `
        [1, 2, 3].map((a = 42) => a + 1);
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      2,
						Column:    28,
						EndColumn: 30,
					},
				},
				Output: []string{`
        [1, 2, 3].map((a) => a + 1);
      `},
			},
			{
				Code: `
        function getValue(): undefined;
        function getValue(box: { value: string }): string;
        function getValue({ value = '' }: { value: string } = {}): string | undefined {
          return value;
        }
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      4,
						Column:    37,
						EndColumn: 39,
					},
				},
				Output: []string{`
        function getValue(): undefined;
        function getValue(box: { value: string }): string;
        function getValue({ value }: { value: string } = {}): string | undefined {
          return value;
        }
      `},
			},
			{
				Code: `
        function getValue([value = '']: [string]) {
          return value;
        }
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      2,
						Column:    36,
						EndColumn: 38,
					},
				},
				Output: []string{`
        function getValue([value]: [string]) {
          return value;
        }
      `},
			},
			{
				Code: `
        declare const x: { hello: { world: string } };

        const {
          hello: { world = '' },
        } = x;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      5,
						Column:    28,
						EndColumn: 30,
					},
				},
				Output: []string{`
        declare const x: { hello: { world: string } };

        const {
          hello: { world },
        } = x;
      `},
			},
			{
				Code: `
        declare const x: { hello: Array<{ world: string }> };

        const {
          hello: [{ world = '' }],
        } = x;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      5,
						Column:    29,
						EndColumn: 31,
					},
				},
				Output: []string{`
        declare const x: { hello: Array<{ world: string }> };

        const {
          hello: [{ world }],
        } = x;
      `},
			},
			{
				Code: `
        interface B {
          foo: (b: boolean | string) => void;
        }

        const h: B = {
          foo: (b = false) => {},
        };
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      7,
						Column:    21,
						EndColumn: 26,
					},
				},
				Output: []string{`
        interface B {
          foo: (b: boolean | string) => void;
        }

        const h: B = {
          foo: (b) => {},
        };
      `},
			},
			{
				Code: `
        function foo(a = undefined) {}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessUndefined",
						Line:      2,
						Column:    26,
						EndColumn: 35,
					},
				},
				Output: []string{`
        function foo(a) {}
      `},
			},
			{
				Code: `
        const { a = undefined } = {};
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessUndefined",
						Line:      2,
						Column:    21,
						EndColumn: 30,
					},
				},
				Output: []string{`
        const { a } = {};
      `},
			},
			{
				Code: `
        const [a = undefined] = [];
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessUndefined",
						Line:      2,
						Column:    20,
						EndColumn: 29,
					},
				},
				Output: []string{`
        const [a] = [];
      `},
			},
			{
				Code: `
        function foo({ a = undefined }) {}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessUndefined",
						Line:      2,
						Column:    28,
						EndColumn: 37,
					},
				},
				Output: []string{`
        function foo({ a }) {}
      `},
			},
			// https://github.com/typescript-eslint/typescript-eslint/issues/11847
			{
				Code: `
        function myFunction(p1: string, p2: number | undefined = undefined) {
          console.log(p1, p2);
        }
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferOptionalSyntax",
						Line:      2,
						Column:    66,
						EndColumn: 75,
					},
				},
				Output: []string{`
        function myFunction(p1: string, p2?: number | undefined) {
          console.log(p1, p2);
        }
      `},
			},
			{
				Code: `
        type SomeType = number | undefined;
        function f(
          /* comment */ x /* comment 2 */ : /* comment 3 */ SomeType /* comment 4 */ = /* comment 5 */ undefined,
        ) {}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferOptionalSyntax",
						Line:      4,
						Column:    104,
						EndColumn: 113,
					},
				},
				Output: []string{`
        type SomeType = number | undefined;
        function f(
          /* comment */ x? /* comment 2 */ : /* comment 3 */ SomeType,
        ) {}
      `},
			},
			// noStrictNullCheck tests — upstream sets tsconfigRootDir to the 'unstrict'
			// fixture dir, which we mirror via TSConfig: "tsconfig.unstrict.json".
			{
				Code: `
        function Bar({ foo = '' }: { foo: string }) {
          return foo;
        }
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noStrictNullCheck",
						Line:      0,
						Column:    1,
					},
					{
						MessageId: "uselessDefaultAssignment",
						Line:      2,
						Column:    30,
						EndColumn: 32,
					},
				},
				TSConfig: "tsconfig.unstrict.json",
				Output: []string{`
        function Bar({ foo }: { foo: string }) {
          return foo;
        }
      `},
			},
			{
				Code: `
        function foo(a = undefined) {}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noStrictNullCheck",
						Line:      0,
						Column:    1,
					},
					{
						MessageId: "uselessUndefined",
						Line:      2,
						Column:    26,
						EndColumn: 35,
					},
				},
				TSConfig: "tsconfig.unstrict.json",
				Output: []string{`
        function foo(a) {}
      `},
			},
			{
				Code: `
        function Bar({ foo = '' }: { foo: string }) {
          return foo;
        }
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      2,
						Column:    30,
						EndColumn: 32,
					},
				},
				TSConfig: "tsconfig.unstrict.json",
				Options: map[string]interface{}{
					"allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing": true,
				},
				Output: []string{`
        function Bar({ foo }: { foo: string }) {
          return foo;
        }
      `},
			},
			// https://github.com/typescript-eslint/typescript-eslint/issues/11980
			{
				Code: `
        const { a = 'baz' } = Math.random() < 0.5 ? { a: 'foo' } : { a: 'bar' };
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      2,
						Column:    21,
						EndColumn: 26,
					},
				},
				Output: []string{`
        const { a } = Math.random() < 0.5 ? { a: 'foo' } : { a: 'bar' };
      `},
			},
			{
				Code: `
        const { a = 'baz' } =
          Math.random() < 0.5
            ? { a: 'foo' }
            : Math.random() > 0.2
              ? { a: 'bar' }
              : { a: 'qux' };
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      2,
						Column:    21,
						EndColumn: 26,
					},
				},
				Output: []string{`
        const { a } =
          Math.random() < 0.5
            ? { a: 'foo' }
            : Math.random() > 0.2
              ? { a: 'bar' }
              : { a: 'qux' };
      `},
			},
			{
				Code: `
        const { a = 'baz' } = cond ? { ['a']: 'foo' } : { ['a']: 'bar' };
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      2,
						Column:    21,
						EndColumn: 26,
					},
				},
				Output: []string{`
        const { a } = cond ? { ['a']: 'foo' } : { ['a']: 'bar' };
      `},
			},
			{
				Code: `
        const { a = 'baz' } = cond ? { a() {} } : { a: 'bar' };
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      2,
						Column:    21,
						EndColumn: 26,
					},
				},
				Output: []string{`
        const { a } = cond ? { a() {} } : { a: 'bar' };
      `},
			},
			{
				Code: "\n        const { a = 'b' } = Math.random() < 0.5 ? { [`a`]: 'a' } : { a: 'b' };\n      ",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      2,
						Column:    21,
						EndColumn: 24,
					},
				},
				Output: []string{"\n        const { a } = Math.random() < 0.5 ? { [`a`]: 'a' } : { a: 'b' };\n      "},
			},
		},
	)
}
