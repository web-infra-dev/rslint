package consistent_indexed_object_style

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentIndexedObjectStyleRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConsistentIndexedObjectStyleRule, []rule_tester.ValidTestCase{
		// Default mode (record style)
		{Code: "type Foo = Record<string, any>;"},
		{Code: "type Foo = Record<string, unknown>;"},
		{Code: "type Foo = Record<number, any>;"},
		{Code: "type Foo = Record<symbol, any>;"},
		{Code: "type Foo = Readonly<Record<string, any>>;"},
		{Code: "type Foo = Partial<Record<string, any>>;"},
		{Code: "type Foo = { [key: string]: any; bar: string };"},
		{Code: "type Foo = { bar: string; [key: string]: any };"},
		{Code: "type Foo = {};"},
		{Code: "interface Foo {}"},
		{Code: "interface Foo { bar: string; }"},
		{Code: "interface Foo { [key: string]: any; bar: string; }"},
		{Code: "interface Foo { bar: string; [key: string]: any; }"},
		{Code: "interface Foo extends Bar { [key: string]: any; }"},

		// Empty interfaces and types
		{Code: "type Empty = {};"},
		{Code: "interface Empty {}"},

		// Mixed properties with index signatures (valid in record mode)
		{Code: "type Foo = { a: string; [key: string]: any };"},
		{Code: "interface Foo { a: string; [key: string]: any; }"},

		// Function signatures with index signature parameters
		{Code: "function foo(arg: { [key: string]: any; bar: string }) {}"},
		{Code: "const foo = (arg: { [key: string]: any; bar: string }) => {};"},

		// Generic interfaces with index signatures and other properties
		{Code: "interface Foo<T> { [key: string]: T; bar: T; }"},

		// Circular type references (should not convert)
		{Code: "interface Foo { [key: string]: Foo; }"},
		{Code: "interface Foo { [key: string]: Foo | string; }"},
		{Code: "interface Foo { [key: string]: Foo[] | string; }"},
		{Code: "type Foo = { [key: string]: Foo; }"},

		// Mapped types that reference the key
		{Code: "type Foo<T extends string> = { [K in T]: K };"},

		// Index-signature mode
		{Code: "type Foo = { [key: string]: any };", Options: "index-signature"},
		{Code: "type Foo = { [key: number]: any };", Options: "index-signature"},
		{Code: "interface Foo { [key: string]: any; }", Options: "index-signature"},
		{Code: "type Foo = { readonly [key: string]: any };", Options: "index-signature"},
		{Code: "type Foo<T> = { [K in string]: T };", Options: "index-signature"},
		{Code: "type Foo = {};", Options: "index-signature"},
		{Code: "interface Foo {}", Options: "index-signature"},

		// Non-Record types in index-signature mode
		{Code: "type Foo = Map<string, any>;", Options: "index-signature"},
		{Code: "type Foo = Array<string>;", Options: "index-signature"},

		// Generic types other than Record
		{Code: "type Foo = Partial<Bar>;"},
		{Code: "type Foo = Required<Bar>;"},
		{Code: "type Foo = Pick<Bar, 'a' | 'b'>;"},
	}, []rule_tester.InvalidTestCase{
		// Default mode (prefer record) - interface with only index signature
		{
			Code: "interface Foo { [key: string]: any; }",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord"},
			},
		},
		{
			Code: "interface Foo { [key: string]: unknown; }",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord"},
			},
		},
		{
			Code: "interface Foo { [key: number]: any; }",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord"},
			},
		},
		{
			Code: "interface Foo { readonly [key: string]: any; }",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord"},
			},
		},

		// Type literals with only index signature
		{
			Code: "type Foo = { [key: string]: any };",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord"},
			},
		},
		{
			Code: "type Foo = { [key: string]: unknown };",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord"},
			},
		},
		{
			Code: "type Foo = { [key: number]: any };",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord"},
			},
		},
		{
			Code: "type Foo = { readonly [key: string]: any };",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord"},
			},
		},

		// Mapped types that can be converted
		{
			Code: "type Foo = { [K in string]: any };",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord"},
			},
		},
		{
			Code: "type Foo = { [K in number]: any };",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord"},
			},
		},
		{
			Code: "type Foo = { readonly [K in string]: any };",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord"},
			},
		},
		{
			Code: "type Foo = { [K in string]?: any };",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord"},
			},
		},

		// Generic interfaces with only index signature
		{
			Code: "interface Foo<T> { [key: string]: T; }",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord"},
			},
		},
		{
			Code: "interface Foo<T, K> { [key: string]: T | K; }",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord"},
			},
		},

		// Index-signature mode (prefer index-signature)
		{
			Code:    "type Foo = Record<string, any>;",
			Options: "index-signature",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIndexSignature"},
			},
		},
		{
			Code:    "type Foo = Record<string, unknown>;",
			Options: "index-signature",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIndexSignature"},
			},
		},
		{
			Code:    "type Foo = Record<number, any>;",
			Options: "index-signature",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIndexSignature"},
			},
		},
		{
			Code:    "type Foo = Record<symbol, any>;",
			Options: "index-signature",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIndexSignature"},
			},
		},
		{
			Code:    "type Foo = Record<'a' | 'b', any>;",
			Options: "index-signature",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIndexSignature"},
			},
		},
		{
			Code:    "type Foo<T> = Record<string, T>;",
			Options: "index-signature",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIndexSignature"},
			},
		},

		// Nested in other types
		{
			Code: "type Foo = Array<{ [key: string]: any }>;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord"},
			},
		},
		{
			Code: "type Foo = { prop: { [key: string]: any } };",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord"},
			},
		},

		// In function signatures
		{
			Code: "function foo(arg: { [key: string]: any }) {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord"},
			},
		},
		{
			Code: "const foo = (arg: { [key: string]: any }) => {};",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord"},
			},
		},
		{
			Code: "function foo(): { [key: string]: any } { return {}; }",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord"},
			},
		},

		// With comments
		{
			Code: "interface Foo { /* comment */ [key: string]: any; }",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord"},
			},
		},
		{
			Code: "type Foo = { /* comment */ [key: string]: any };",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord"},
			},
		},

		// Complex value types
		{
			Code: "type Foo = { [key: string]: string | number };",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord"},
			},
		},
		{
			Code: "type Foo = { [key: string]: { nested: string } };",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord"},
			},
		},
		{
			Code: "interface Foo { [key: string]: Array<string> }",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord"},
			},
		},
	})
}
