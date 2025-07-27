package consistent_indexed_object_style

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rule_tester"
	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
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
			
			// Indirect circular references
			{Code: `interface Foo1 {
  [key: string]: Foo2;
}

interface Foo2 {
  [key: string]: Foo1;
}`},
			
			// index-signature mode valid cases
			{Code: "type Foo = { [key: string]: any };", Options: []interface{}{"index-signature"}},
			{Code: "type Foo = Record;", Options: []interface{}{"index-signature"}},
			{Code: "type Foo = Record<string>;", Options: []interface{}{"index-signature"}},
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
			
			// Interface with wrapped self-reference (should convert)
			{
				Code: `interface Foo {
  [key: string]: { foo: Foo };
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRecord", Line: 1, Column: 1},
				},
				Output: []string{`type Foo = Record<string, { foo: Foo }>;`},
			},
			
			// Interface with array self-reference (should convert)
			{
				Code: `interface Foo {
  [key: string]: Foo[];
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRecord", Line: 1, Column: 1},
				},
				Output: []string{`type Foo = Record<string, Foo[]>;`},
			},
			
			// Interface with function self-reference (should convert)
			{
				Code: `interface Foo {
  [key: string]: () => Foo;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRecord", Line: 1, Column: 1},
				},
				Output: []string{`type Foo = Record<string, () => Foo>;`},
			},
			
			// Interface with tuple self-reference (should convert)
			{
				Code: `interface Foo {
  [s: string]: [Foo];
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRecord", Line: 1, Column: 1},
				},
				Output: []string{`type Foo = Record<string, [Foo]>;`},
			},
			
			// index-signature mode tests
			{
				Code: "type Foo = Record<string, any>;",
				Options: []interface{}{"index-signature"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferIndexSignature", Line: 1, Column: 12},
				},
				Output: []string{"type Foo = { [key: string]: any };"},
			},
			
			{
				Code: "type Foo<T> = Record<string, T>;",
				Options: []interface{}{"index-signature"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferIndexSignature", Line: 1, Column: 15},
				},
				Output: []string{"type Foo<T> = { [key: string]: T };"},
			},
			
			// Record with complex key type (should use suggestion)
			{
				Code: "type Foo = Record<string | number, any>;",
				Options: []interface{}{"index-signature"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferIndexSignature", Line: 1, Column: 12},
				},
				// Note: Suggestions not supported in this test framework version
			},
		})
}