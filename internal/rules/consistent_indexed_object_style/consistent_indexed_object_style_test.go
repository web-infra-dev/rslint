package consistent_indexed_object_style

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestConsistentIndexedObjectStyleRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConsistentIndexedObjectStyleRule,
		[]rule_tester.ValidTestCase{
			// Basic valid cases
			{Code: "type Foo = Record<string, any>;"},
			{Code: "interface Foo {}"},
			{Code: `interface Foo {
  bar: string;
}`},
			{Code: `interface Foo {
  bar: string;
  [key: string]: any;
}`},

			// Circular references that should be allowed (blocked from conversion)
			{Code: "type Foo = { [key: string]: string | Foo };"},
			{Code: "type Foo = { [key: string]: Foo };"},
			{Code: "type Foo = { [key: string]: Foo } | Foo;"},
			{Code: `interface Foo {
  [key: string]: Foo;
}`},
			{Code: `interface Foo<T> {
  [key: string]: Foo<T>;
}`},
			// Wrapped self-references should also be blocked (matching TypeScript-ESLint)
			{Code: `interface Foo {
  [key: string]: Foo[];
}`},
			{Code: `interface Foo {
  [key: string]: () => Foo;
}`},
			{Code: `interface Foo {
  [s: string]: [Foo];
}`},
			{Code: `interface Foo {
  [key: string]: { foo: Foo };
}`},

			// More complex circular reference patterns
			{Code: `interface Foo {
  [s: string]: Foo & {};
}`},
			{Code: `interface Foo {
  [s: string]: Foo | string;
}`},
			{Code: `interface Foo<T> {
  [s: string]: Foo extends T ? string : number;
}`},
			{Code: `interface Foo<T> {
  [s: string]: T extends Foo ? string : number;
}`},
			{Code: `interface Foo<T> {
  [s: string]: T extends true ? Foo : number;
}`},
			{Code: `interface Foo<T> {
  [s: string]: T extends true ? string : Foo;
}`},
			{Code: `interface Foo {
  [s: string]: Foo[number];
}`},
			{Code: `interface Foo {
  [s: string]: {}[Foo];
}`},

			// Indirect circular references
			{Code: `interface Foo1 {
  [key: string]: Foo2;
}

interface Foo2 {
  [key: string]: Foo1;
}`},

			// Mapped types that cannot be converted to Record - these use 'in' keyword which is different from index signatures
			// Note: The current implementation only handles index signatures (with ':'), not mapped types (with 'in')
			// These are kept as comments to document what's not supported:
			// {Code: "type T = { [key in Foo]: key | number };"},
			// {Code: `function foo(e: { readonly [key in PropertyKey]-?: key }) {}`},
			// {Code: `function f(): { [k in keyof ParseResult]: unknown; } { return {}; }`},

			// index-signature mode valid cases
			{Code: "type Foo = { [key: string]: any };", Options: []interface{}{"index-signature"}},
			{Code: "type Foo = Record;", Options: []interface{}{"index-signature"}},
			{Code: "type Foo = Record<string>;", Options: []interface{}{"index-signature"}},
			{Code: "type Foo = Record<string, number, unknown>;", Options: []interface{}{"index-signature"}},
			{Code: "type T = A.B;", Options: []interface{}{"index-signature"}},
		},
		[]rule_tester.InvalidTestCase{
			// Basic interface conversion
			{
				Code: `interface Foo {
  [key: string]: any;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRecord", Line: 1, Column: 1},
				},
				Output: []string{`type Foo = Record<string, any>;`},
			},

			// Readonly interface
			{
				Code: `interface Foo {
  readonly [key: string]: any;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRecord", Line: 1, Column: 1},
				},
				Output: []string{`type Foo = Readonly<Record<string, any>>;`},
			},

			// Interface with generic parameter
			{
				Code: `interface Foo<A> {
  [key: string]: A;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRecord", Line: 1, Column: 1},
				},
				Output: []string{`type Foo<A> = Record<string, A>;`},
			},

			// Interface with default generic parameter
			{
				Code: `interface Foo<A = any> {
  [key: string]: A;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRecord", Line: 1, Column: 1},
				},
				Output: []string{`type Foo<A = any> = Record<string, A>;`},
			},

			// Interface with extends (no fix available)
			{
				Code: `interface B extends A {
  [index: number]: unknown;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRecord", Line: 1, Column: 1},
				},
				Output: []string{}, // No fix available
			},

			// Interface with multiple generic parameters
			{
				Code: `interface Foo<A, B> {
  [key: A]: B;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRecord", Line: 1, Column: 1},
				},
				Output: []string{`type Foo<A, B> = Record<A, B>;`},
			},

			// Readonly interface with multiple generic parameters
			{
				Code: `interface Foo<A, B> {
  readonly [key: A]: B;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRecord", Line: 1, Column: 1},
				},
				Output: []string{`type Foo<A, B> = Readonly<Record<A, B>>;`},
			},

			// Type literal conversion
			{
				Code: "type Foo = { [key: string]: any };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRecord", Line: 1, Column: 12},
				},
				Output: []string{"type Foo = Record<string, any>;"},
			},

			// Readonly type literal
			{
				Code: "type Foo = { readonly [key: string]: any };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRecord", Line: 1, Column: 12},
				},
				Output: []string{"type Foo = Readonly<Record<string, any>>;"},
			},

			// Generic type literal
			{
				Code: "type Foo = Generic<{ [key: string]: any }>;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRecord", Line: 1, Column: 20},
				},
				Output: []string{"type Foo = Generic<Record<string, any>>;"},
			},

			// Function parameter
			{
				Code: "function foo(arg: { [key: string]: any }) {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRecord", Line: 1, Column: 19},
				},
				Output: []string{"function foo(arg: Record<string, any>) {}"},
			},

			// Function return type
			{
				Code: "function foo(): { [key: string]: any } {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRecord", Line: 1, Column: 17},
				},
				Output: []string{"function foo(): Record<string, any> {}"},
			},

			// The critical nested case - inner type literal should be converted
			{
				Code: "type Foo = { [key: string]: { [key: string]: Foo } };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRecord", Line: 1, Column: 29},
				},
				Output: []string{"type Foo = { [key: string]: Record<string, Foo> };"},
			},

			// Union with type literal
			{
				Code: "type Foo = { [key: string]: string } | Foo;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRecord", Line: 1, Column: 12},
				},
				Output: []string{"type Foo = Record<string, string> | Foo;"},
			},

			// index-signature mode tests
			{
				Code:    "type Foo = Record<string, any>;",
				Options: []interface{}{"index-signature"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferIndexSignature", Line: 1, Column: 12},
				},
				Output: []string{"type Foo = { [key: string]: any };"},
			},

			{
				Code:    "type Foo<T> = Record<string, T>;",
				Options: []interface{}{"index-signature"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferIndexSignature", Line: 1, Column: 15},
				},
				Output: []string{"type Foo<T> = { [key: string]: T };"},
			},

			// Note: Mapped types (with 'in' keyword) are not supported by the current implementation
			// The rule only handles index signatures (with ':' syntax)

			// Missing type annotation (edge case)
			{
				Code: `interface Foo {
  [key: string]: Bar;
}

interface Bar {
  [key: string];
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRecord", Line: 1, Column: 1},
				},
				Output: []string{`type Foo = Record<string, Bar>;

interface Bar {
  [key: string];
}`},
			},

			// Record with complex key type (should use suggestion)
			{
				Code:    "type Foo = Record<string | number, any>;",
				Options: []interface{}{"index-signature"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferIndexSignature", Line: 1, Column: 12},
				},
				// Note: Suggestions not supported in this test framework version
			},

			// Record with number key
			{
				Code:    "type Foo = Record<number, any>;",
				Options: []interface{}{"index-signature"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferIndexSignature", Line: 1, Column: 12},
				},
				Output: []string{"type Foo = { [key: number]: any };"},
			},

			// Record with symbol key
			{
				Code:    "type Foo = Record<symbol, any>;",
				Options: []interface{}{"index-signature"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferIndexSignature", Line: 1, Column: 12},
				},
				Output: []string{"type Foo = { [key: symbol]: any };"},
			},
		})
}
