// TestNoUnnecessaryTypeAssertionContextual covers contextual-assignability
// edge cases and false-positive guards beyond the original upstream migration.
package no_unnecessary_type_assertion

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnnecessaryTypeAssertionContextual(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnnecessaryTypeAssertionRule,
		[]rule_tester.ValidTestCase{
			{Code: `const value = 5 as any;`},
			{
				Code:    `declare const x: number; x as /* comment */ number;`,
				Options: []any{map[string]any{"typesToIgnore": []any{"number"}}},
			},
			{Code: `declare const value: unknown; const result: number = value as number;`},
			{Code: `declare function fn(x: string | undefined): void; fn(undefined as string | undefined);`},
			{Code: `declare function fn(x: number[]): void; fn([...(1 as any)]);`},
			{Code: `declare const source: unknown; const { value } = source as { value: string };`},
			{Code: `declare let x: number | undefined; x ??= 1 as number;`},
			{Code: `declare let x: number; const y = (x = 1 as number);`},
			{Code: `declare const value: string | undefined; (value as string | undefined)?.toLowerCase();`},
			{Code: `declare const value: string | undefined; const result = (value as string | undefined) || 'fallback';`},
			{Code: `
interface W { name?: string }
interface N { name: string }
declare const n: N;
declare function id<T>(value: T): T;
const result = (id({ value: n as W })) satisfies { value: W };
`},
			{Code: `
interface W { name?: string }
interface N { name: string }
declare const n: N;
declare function id<T>(value: T): T;
const result = (((id({ value: n as W })))) satisfies { value: W };
`},
			{Code: `type ValuePath = 'values' | ` + "`values.${string}`" + `; declare function apply(paths: ValuePath[]): void; declare const ids: string[]; apply(ids.map(id => ` + "`values.${id}`" + ` as ValuePath));`},
			{Code: `
declare function fn(x: string): void;
declare function fn(x: number): void;
declare const value: string | number;
fn(value as string);
`},
			{Code: `
declare function fn<T>(x: T): T;
declare const value: 'a';
fn(value as string);
`},
			{Code: `
type Infer<T> = T extends () => infer V ? V : never;
declare function fn<T>(o: { p: T }): { [K in keyof T]: Infer<T[K]> };
const result = fn({ p: { a: Object as () => string } });
result.a.toLowerCase();
`},
			{Code: `
type Test<T extends Record<string, unknown>> = {};
declare function inferred<T extends Test<never>[]>(input: { addons?: T }): void;
inferred({ addons: [{} as Test<{ parameters: { potato: boolean } }>] });
`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `declare const value: 'a'; const result = true ? (value as string) : 'b';`,
				Output: []string{`declare const value: 'a'; const result = true ? (value) : 'b';`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "contextuallyUnnecessary"}},
			},
			{
				Code:   `const x = ((1 as 1));`,
				Output: []string{`const x = ((1));`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
			},
			{
				Code:   `class C { readonly x = (1 as 1); }`,
				Output: []string{`class C { readonly x = (1); }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
			},
			{
				Code:   `declare namespace JSX { interface IntrinsicElements { div: { p: string } } } declare const x: 'a'; <div p={x as string} />;`,
				Output: []string{`declare namespace JSX { interface IntrinsicElements { div: { p: string } } } declare const x: 'a'; <div p={x} />;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "contextuallyUnnecessary"}},
				Tsx:    true,
			},
			{
				Code:   `declare namespace JSX { interface IntrinsicElements { div: { p?: string | null } } } declare const x: 'a'; <div p={true ? x as string : null} />;`,
				Output: []string{`declare namespace JSX { interface IntrinsicElements { div: { p?: string | null } } } declare const x: 'a'; <div p={true ? x : null} />;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "contextuallyUnnecessary"}},
				Tsx:    true,
			},
			{
				Code:   `declare namespace JSX { interface IntrinsicElements { div: { p?: string | null } } } declare const x: 'a'; <div p={(x as string) || null} />;`,
				Output: []string{`declare namespace JSX { interface IntrinsicElements { div: { p?: string | null } } } declare const x: 'a'; <div p={(x) || null} />;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "contextuallyUnnecessary"}},
				Tsx:    true,
			},
			{
				Code:   `declare namespace JSX { interface IntrinsicElements { div: { p?: string | null } } } declare const x: string | null; <div p={x!} />;`,
				Output: []string{`declare namespace JSX { interface IntrinsicElements { div: { p?: string | null } } } declare const x: string | null; <div p={x} />;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "contextuallyUnnecessary"}},
				Tsx:    true,
			},
			{
				Code:   `declare namespace JSX { interface IntrinsicElements { div: { p?: string | null } } } declare const x: 'a'; declare function f(value: string): string; <div p={f(x as string)} />;`,
				Output: []string{`declare namespace JSX { interface IntrinsicElements { div: { p?: string | null } } } declare const x: 'a'; declare function f(value: string): string; <div p={f(x)} />;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "contextuallyUnnecessary"}},
				Tsx:    true,
			},
			{
				Code:   `declare const x: number; x  as number;`,
				Output: []string{`declare const x: number; x;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
			},
			{
				Code:   "const v: number = 5;\nconst x = <number>(<any>v);",
				Output: []string{"const v: number = 5;\nconst x = v;"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
			},
			{
				Code:   "declare function fn(param: number): void;\nfn(42 as unknown as number);",
				Output: []string{"declare function fn(param: number): void;\nfn(42);"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "contextuallyUnnecessary",
					Line:      2,
					Column:    4,
					EndLine:   2,
					EndColumn: 27,
				}},
			},
			{
				Code:   `const fn = (): {} => <any>{};`,
				Output: []string{`const fn = (): {} => ({});`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "contextuallyUnnecessary",
				}},
			},
			{
				Code:   "function fn(x: number): void {}\nfn(5 as any);",
				Output: []string{"function fn(x: number): void {}\nfn(5);"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "contextuallyUnnecessary",
					Message:   "This assertion is unnecessary since the receiver accepts the original type of the expression.",
					Line:      2,
					Column:    4,
					EndLine:   2,
					EndColumn: 12,
				}},
			},
			{
				Code:   `type AB = 'a' | 'b'; declare const a: 'a'; declare function fn(x: AB): void; fn(a as AB);`,
				Output: []string{`type AB = 'a' | 'b'; declare const a: 'a'; declare function fn(x: AB): void; fn(a);`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "contextuallyUnnecessary",
					Line:      1,
					Column:    81,
					EndLine:   1,
					EndColumn: 88,
				}},
			},
			{
				Code:   `declare function fn(x: { value: number }): void; fn({ value: 42 as number });`,
				Output: []string{`declare function fn(x: { value: number }): void; fn({ value: 42 });`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "contextuallyUnnecessary",
					Line:      1,
					Column:    62,
					EndLine:   1,
					EndColumn: 74,
				}},
			},
			{
				Code:   `declare const a: 'a'; const value: string = a as string;`,
				Output: []string{`declare const a: 'a'; const value: string = a;`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "contextuallyUnnecessary",
					Line:      1,
					Column:    45,
					EndLine:   1,
					EndColumn: 56,
				}},
			},
			{
				Code:   `declare const a: 'a'; const value: string = <string>a;`,
				Output: []string{`declare const a: 'a'; const value: string = a;`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "contextuallyUnnecessary",
					Line:      1,
					Column:    45,
					EndLine:   1,
					EndColumn: 54,
				}},
			},
			{
				Code:   `declare function fn(x: any): void; declare const value: string | number; fn(value as number);`,
				Output: []string{`declare function fn(x: any): void; declare const value: string | number; fn(value);`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "contextuallyUnnecessary",
					Line:      1,
					Column:    77,
					EndLine:   1,
					EndColumn: 92,
				}},
			},
		},
	)
}
