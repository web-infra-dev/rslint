package no_invalid_void_type

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoInvalidVoidTypeRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoInvalidVoidTypeRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			{Code: `function func(): void {}`},
			{Code: `type NormalType = () => void;`},
			{Code: `let voidPromise: Promise<void> = new Promise<void>(() => {});`},
			{Code: `type GenericVoid = Generic<void>;`},
			{Code: `function foo(): void | never { throw new Error('Test'); }`},
			{Code: `type voidNeverUnion = void | never;`},
			{Code: `type neverVoidUnion = never | void;`},
			{Code: `async function asyncFunc(): Promise<void> {}`},
			{Code: `type FunctionType = (x: number) => void;`},
			{Code: `interface Foo { method(): void; }`},
			{Code: `function f(this: void) {}`, Options: map[string]interface{}{"allowAsThisParameter": true}},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			{
				Code: `function takeVoid(thing: void) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturn"},
				},
			},
			{
				Code: `type UnionType2 = string | number | void;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturn"},
				},
			},
			{
				Code: `type IntersectionType = string & number & void;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturn"},
				},
			},
			{
				Code: `interface Interface {
  prop: void;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturn"},
				},
			},
			{
				Code: `class MyClass {
  private readonly propName: void;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturn"},
				},
			},
			{
				Code: `let voidPromise: Promise<void> = new Promise<void>(() => {});`,
				Options: map[string]interface{}{"allowInGenericTypeArguments": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturn"},
					{MessageId: "invalidVoidNotReturn"},
				},
			},
			{
				Code: `type GenericVoid = Generic<void>;`,
				Options: map[string]interface{}{"allowInGenericTypeArguments": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturn"},
				},
			},
		},
	)
}
