package consistent_type_definitions

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestConsistentTypeDefinitionsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConsistentTypeDefinitionsRule, []rule_tester.ValidTestCase{
		// Valid with 'interface' option (default)
		{Code: "var foo = {};"},
		{Code: "interface A {}"},
		{Code: `
interface A extends B {
  x: number;
}
		`},
		{Code: "type U = string;"},
		{Code: "type V = { x: number } | { y: string };"},
		{Code: `
type Record<T, U> = {
  [K in T]: U;
};
		`},
		
		// Valid with 'type' option
		{Code: "type T = { x: number };", Options: []interface{}{"type"}},
		{Code: "type A = { x: number } & B & C;", Options: []interface{}{"type"}},
		{Code: "type A = { x: number } & B<T1> & C<T2>;", Options: []interface{}{"type"}},
		{Code: `
export type W<T> = {
  x: T;
};
		`, Options: []interface{}{"type"}},
	}, []rule_tester.InvalidTestCase{
		// Interface option violations (type -> interface)
		{
			Code: "type T = { x: number; };",
			Options: []interface{}{"interface"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "interfaceOverType",
					Line:      1,
					Column:    6,
				},
			},
			Output: []string{"interface T { x: number; }"},
		},
		{
			Code: "type T={ x: number; };",
			Options: []interface{}{"interface"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "interfaceOverType",
					Line:      1,
					Column:    6,
				},
			},
			Output: []string{"interface T { x: number; }"},
		},
		{
			Code: "type T=                         { x: number; };",
			Options: []interface{}{"interface"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "interfaceOverType",
					Line:      1,
					Column:    6,
				},
			},
			Output: []string{"interface T { x: number; }"},
		},
		{
			Code: "type T /* comment */={ x: number; };",
			Options: []interface{}{"interface"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "interfaceOverType",
					Line:      1,
					Column:    6,
				},
			},
			Output: []string{"interface T /* comment */ { x: number; }"},
		},
		{
			Code: `
export type W<T> = {
  x: T;
};
			`,
			Options: []interface{}{"interface"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "interfaceOverType",
					Line:      2,
					Column:    13,
				},
			},
			Output: []string{`
export interface W<T> {
  x: T;
}
			`},
		},
		
		// Type option violations (interface -> type)
		{
			Code: "interface T { x: number; }",
			Options: []interface{}{"type"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "typeOverInterface",
					Line:      1,
					Column:    11,
				},
			},
			Output: []string{"type T = { x: number; }"},
		},
		{
			Code: "interface T{ x: number; }",
			Options: []interface{}{"type"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "typeOverInterface",
					Line:      1,
					Column:    11,
				},
			},
			Output: []string{"type T = { x: number; }"},
		},
		{
			Code: "interface T                          { x: number; }",
			Options: []interface{}{"type"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "typeOverInterface",
					Line:      1,
					Column:    11,
				},
			},
			Output: []string{"type T = { x: number; }"},
		},
		{
			Code: "interface A extends B, C { x: number; };",
			Options: []interface{}{"type"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "typeOverInterface",
					Line:      1,
					Column:    11,
				},
			},
			Output: []string{"type A = { x: number; } & B & C;"},
		},
		{
			Code: "interface A extends B<T1>, C<T2> { x: number; };",
			Options: []interface{}{"type"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "typeOverInterface",
					Line:      1,
					Column:    11,
				},
			},
			Output: []string{"type A = { x: number; } & B<T1> & C<T2>;"},
		},
		{
			Code: `
export interface W<T> {
  x: T;
}
			`,
			Options: []interface{}{"type"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "typeOverInterface",
					Line:      2,
					Column:    18,
				},
			},
			Output: []string{`
export type W<T> = {
  x: T;
}
			`},
		},
		{
			Code: `
namespace JSX {
  interface Array<T> {
    foo(x: (x: number) => T): T[];
  }
}
			`,
			Options: []interface{}{"type"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "typeOverInterface",
					Line:      3,
					Column:    13,
				},
			},
			Output: []string{`
namespace JSX {
  type Array<T> = {
    foo(x: (x: number) => T): T[];
  }
}
			`},
		},
		{
			Code: `
global {
  interface Array<T> {
    foo(x: (x: number) => T): T[];
  }
}
			`,
			Options: []interface{}{"type"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "typeOverInterface",
					Line:      3,
					Column:    13,
				},
			},
			Output: []string{`
global {
  type Array<T> = {
    foo(x: (x: number) => T): T[];
  }
}
			`},
		},
		{
			Code: `
declare global {
  interface Array<T> {
    foo(x: (x: number) => T): T[];
  }
}
			`,
			Options: []interface{}{"type"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "typeOverInterface",
					Line:      3,
					Column:    13,
				},
			},
			// No output expected - should not provide fixes for declare global
		},
		{
			Code: `
declare global {
  namespace Foo {
    interface Bar {}
  }
}
			`,
			Options: []interface{}{"type"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "typeOverInterface",
					Line:      4,
					Column:    15,
				},
			},
			// No output expected - should not provide fixes for declare global
		},
		{
			Code: `
export default interface Test {
  bar(): string;
  foo(): number;
}
			`,
			Options: []interface{}{"type"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "typeOverInterface",
					Line:      2,
					Column:    26,
				},
			},
			Output: []string{`
type Test = {
  bar(): string;
  foo(): number;
}
export default Test
			`},
		},
		{
			Code: `
export declare type Test = {
  foo: string;
  bar: string;
};
			`,
			Options: []interface{}{"interface"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "interfaceOverType",
					Line:      2,
					Column:    21,
				},
			},
			Output: []string{`
export declare interface Test {
  foo: string;
  bar: string;
}
			`},
		},
		{
			Code: `
export declare interface Test {
  foo: string;
  bar: string;
}
			`,
			Options: []interface{}{"type"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "typeOverInterface",
					Line:      2,
					Column:    26,
				},
			},
			Output: []string{`
export declare type Test = {
  foo: string;
  bar: string;
}
			`},
		},
		{
			Code: `
type Foo = ({
  a: string;
});
			`,
			Options: []interface{}{"interface"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "interfaceOverType",
					Line:      2,
					Column:    6,
				},
			},
			Output: []string{`
interface Foo {
  a: string;
}
			`},
		},
		{
			Code: `
type Foo = ((((((((({
  a: string;
})))))))));
			`,
			Options: []interface{}{"interface"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "interfaceOverType",
					Line:      2,
					Column:    6,
				},
			},
			Output: []string{`
interface Foo {
  a: string;
}
			`},
		},
		{
			Code: `
type Foo = {
  a: string;
}
			`,
			Options: []interface{}{"interface"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "interfaceOverType",
					Line:      2,
					Column:    6,
				},
			},
			Output: []string{`
interface Foo {
  a: string;
}
			`},
		},
		{
			Code: `
type Foo = {
  a: string;
}
type Bar = string;
			`,
			Options: []interface{}{"interface"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "interfaceOverType",
					Line:      2,
					Column:    6,
				},
			},
			Output: []string{`
interface Foo {
  a: string;
}
type Bar = string;
			`},
		},
		{
			Code: `
type Foo = ((({
  a: string;
})))

const bar = 1;
			`,
			Options: []interface{}{"interface"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "interfaceOverType",
					Line:      2,
					Column:    6,
				},
			},
			Output: []string{`
interface Foo {
  a: string;
}

const bar = 1;
			`},
		},
	})
}