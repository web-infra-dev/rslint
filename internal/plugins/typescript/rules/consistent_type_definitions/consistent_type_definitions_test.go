package consistent_type_definitions

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentTypeDefinitionsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConsistentTypeDefinitionsRule, []rule_tester.ValidTestCase{
		// Default options (style: 'interface')
		{Code: `var foo = {};`},
		{Code: `interface A {}`},
		{Code: `interface A { x: number; }`},
		{Code: `interface A extends B { x: number; }`},
		{Code: `type U = string;`},
		{Code: `type V = { x: number } | { y: string };`},
		{Code: `type V = { x: number } & { y: string };`},
		{Code: `type Record<T, U> = { [K in T]: U };`},
		{Code: `type T = string | number;`},
		{Code: `type T = () => void;`},
		{Code: `type T = new () => void;`},
		{Code: `type T = [number, string];`},
		{Code: `type T = number[];`},
		{Code: `type T = readonly number[];`},
		{Code: `type T<U> = U & { x: number };`},

		// style: 'type'
		{Code: `type T = { x: number; };`, Options: []interface{}{"type"}},
		{Code: `type T = { x: number };`, Options: []interface{}{"type"}},
		{Code: `type T = { x: number; y: string; };`, Options: []interface{}{"type"}},
		{Code: `type A = { x: number } & B & C;`, Options: []interface{}{"type"}},
		{Code: `type A = { x: number } & B<T1> & C<T2>;`, Options: []interface{}{"type"}},
		{Code: `export type W<T> = { x: T };`, Options: []interface{}{"type"}},
		{Code: `export type W<T> = { x: T; y: U; };`, Options: []interface{}{"type"}},
		{Code: `type U = string;`, Options: []interface{}{"type"}},
		{Code: `type V = { x: number } | { y: string };`, Options: []interface{}{"type"}},
		{Code: `type Record<T, U> = { [K in T]: U };`, Options: []interface{}{"type"}},
	}, []rule_tester.InvalidTestCase{
		// Default options (style: 'interface') - expect type to be interface
		{
			Code:   `type T = { [K: string]: number };`,
			Output: []string{`interface T { [K: string]: number }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code:   `type T = { x: number; };`,
			Output: []string{`interface T { x: number; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code:   `type T={ x: number; };`,
			Output: []string{`interface T { x: number; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code:   `type T= { x: number; };`,
			Output: []string{`interface T { x: number; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code:   `type T = { x: number };`,
			Output: []string{`interface T { x: number }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code:   `type T = { x: number; y: string; };`,
			Output: []string{`interface T { x: number; y: string; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code:   `type T = { x: number; y: { z: string; }; };`,
			Output: []string{`interface T { x: number; y: { z: string; }; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code:   `export type W<T> = { x: T; };`,
			Output: []string{`export interface W<T> { x: T; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code:   `type T<U> = { x: U; };`,
			Output: []string{`interface T<U> { x: U; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code:   `type Foo = { a: string; };`,
			Output: []string{`interface Foo { a: string; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code:   `type Foo = ({ a: string; });`,
			Output: []string{`interface Foo { a: string; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code:   `type Foo = (  { a: string; });`,
			Output: []string{`interface Foo { a: string; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		// type → interface with comment
		{
			Code:   `type T /* comment */={ x: number; };`,
			Output: []string{`interface T /* comment */ { x: number; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		// type → interface with excessive whitespace
		{
			Code:   `type T=                         { x: number; };`,
			Output: []string{`interface T { x: number; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		// no closing semicolon
		{
			Code:   "type Foo = {\n  a: string;\n}",
			Output: []string{"interface Foo {\n  a: string;\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		// no closing semicolon; ensure we don't erase subsequent code.
		{
			Code:   "type Foo = {\n  a: string;\n}\ntype Bar = string;",
			Output: []string{"interface Foo {\n  a: string;\n}\ntype Bar = string;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		// Parenthesized type - multiple layers
		{
			Code:   `type Foo = ((((((((({ a: string; })))))))));`,
			Output: []string{`interface Foo { a: string; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		// no closing semicolon with parenthesized type
		{
			Code:   "type Foo = ((({ a: string; })))\n\nconst bar = 1;",
			Output: []string{"interface Foo { a: string; }\n\nconst bar = 1;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		// export declare type
		{
			Code:   "export declare type Test = {\n  foo: string;\n  bar: string;\n};",
			Output: []string{"export declare interface Test {\n  foo: string;\n  bar: string;\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},

		// style: 'type' - expect interface to be type
		{
			Code:    `interface T { x: number; }`,
			Options: []interface{}{"type"},
			Output:  []string{`type T = { x: number; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `interface T { x: number }`,
			Options: []interface{}{"type"},
			Output:  []string{`type T = { x: number }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `interface T { x: number; y: string; }`,
			Options: []interface{}{"type"},
			Output:  []string{`type T = { x: number; y: string; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `interface A extends B, C { x: number; };`,
			Options: []interface{}{"type"},
			Output:  []string{`type A = { x: number; } & B & C;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `interface A extends B<T1>, C<T2> { x: number; };`,
			Options: []interface{}{"type"},
			Output:  []string{`type A = { x: number; } & B<T1> & C<T2>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `export interface W<T> { x: T; };`,
			Options: []interface{}{"type"},
			Output:  []string{`export type W<T> = { x: T; };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `interface T<U> { x: U; };`,
			Options: []interface{}{"type"},
			Output:  []string{`type T<U> = { x: U; };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `interface Foo { a: string; }`,
			Options: []interface{}{"type"},
			Output:  []string{`type Foo = { a: string; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `namespace Foo { export interface Bar {} }`,
			Options: []interface{}{"type"},
			Output:  []string{`namespace Foo { export type Bar = {} }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		// interface → type with excessive whitespace
		{
			Code:    `interface T                          { x: number; }`,
			Options: []interface{}{"type"},
			Output:  []string{`type T = { x: number; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		// interface → type, no space before brace
		{
			Code:    `interface T{ x: number; }`,
			Options: []interface{}{"type"},
			Output:  []string{`type T = { x: number; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		// namespace JSX
		{
			Code:    "namespace JSX {\n  interface Array<T> {\n    foo(x: (x: number) => T): T[];\n  }\n}",
			Options: []interface{}{"type"},
			Output:  []string{"namespace JSX {\n  type Array<T> = {\n    foo(x: (x: number) => T): T[];\n  }\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		// global without declare (should be fixable)
		{
			Code:    "global {\n  interface Array<T> {\n    foo(x: (x: number) => T): T[];\n  }\n}",
			Options: []interface{}{"type"},
			Output:  []string{"global {\n  type Array<T> = {\n    foo(x: (x: number) => T): T[];\n  }\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		// export default interface
		{
			Code:    "export default interface Test {\n  bar(): string;\n  foo(): number;\n}",
			Options: []interface{}{"type"},
			Output:  []string{"type Test = {\n  bar(): string;\n  foo(): number;\n}\nexport default Test"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		// export declare interface
		{
			Code:    "export declare interface Test {\n  foo: string;\n  bar: string;\n}",
			Options: []interface{}{"type"},
			Output:  []string{"export declare type Test = {\n  foo: string;\n  bar: string;\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},

		// Global module cases - declare global: report but no fix
		{
			Code:    `declare global { interface Array<T> { foo(): void; } }`,
			Options: []interface{}{"type"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `declare global { namespace Foo { interface Bar {} } }`,
			Options: []interface{}{"type"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
	})
}
