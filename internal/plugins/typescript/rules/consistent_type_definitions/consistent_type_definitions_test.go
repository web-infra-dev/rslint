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
		{Code: `type T = { [K: string]: number };`},
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
			Code: `type T = { x: number; };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code: `type T={ x: number; };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code: `type T= { x: number; };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code: `type T = { x: number };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code: `type T = { x: number; y: string; };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code: `type T = { x: number; y: { z: string; }; };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code: `export type W<T> = { x: T; };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code: `type T<U> = { x: U; };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code: `type Foo = { a: string; };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code: `type Foo = ({ a: string; });`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code: `type Foo = (  { a: string; });`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},

		// style: 'type' - expect interface to be type
		{
			Code:    `interface T { x: number; }`,
			Options: []interface{}{"type"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `interface T { x: number }`,
			Options: []interface{}{"type"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `interface T { x: number; y: string; }`,
			Options: []interface{}{"type"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `interface A extends B, C { x: number; };`,
			Options: []interface{}{"type"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `interface A extends B<T1>, C<T2> { x: number; };`,
			Options: []interface{}{"type"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `export interface W<T> { x: T; };`,
			Options: []interface{}{"type"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `interface T<U> { x: U; };`,
			Options: []interface{}{"type"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `interface Foo { a: string; }`,
			Options: []interface{}{"type"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `namespace Foo { export interface Bar {} }`,
			Options: []interface{}{"type"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},

		// Global module cases
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
