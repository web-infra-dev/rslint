package no_this_alias

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoThisAliasRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoThisAliasRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			{Code: `const self = foo(this);`},
			{Code: `const value = this as unknown;`},
			{Code: `const value = this!;`},
			{Code: `const value = this satisfies unknown;`},
			{Code: `const value = <unknown>this;`},
			{Code: `function fn(value = this) {}`},
			{Code: `class Example { value = this; }`},
			{Code: `object.value = this;`},
			{Code: `let value; value = foo(this); value = this as unknown; value = this!;`},
			{Code: `let value; ({ value } = this); [value] = this;`},
			{Code: `let value; (value as unknown) = this;`},
			{Code: `const { props, state } = this;`},
			{Code: `const { length } = this;`},
			{Code: `const { length, toString } = this;`},
			{Code: `const [foo] = this;`},
			{Code: `const [foo, bar] = this;`},
			{Code: `const self = this;`, Options: map[string]interface{}{"allowedNames": []interface{}{"self"}}},
			{Code: `let self = 1; self ||= this;`, Options: map[string]interface{}{"allowedNames": []interface{}{"self"}}},
			{Code: `setTimeout(() => { this.doWork(); });`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			{
				Code: `const self = this;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "thisAssignment"},
				},
			},
			{
				Code: `const self = (((this)));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "thisAssignment", Line: 1, Column: 7, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code: `let that; that = this;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "thisAssignment"},
				},
			},
			{
				Code: `let that; ((that)) = this;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "thisAssignment", Line: 1, Column: 13, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code: `let that = 1; that += this;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "thisAssignment"},
				},
			},
			{
				Code: `let value: unknown;
value &&= this;
value ||= (this);
value ??= this;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "thisAssignment"},
					{MessageId: "thisAssignment"},
					{MessageId: "thisAssignment"},
				},
			},
			{
				Code:    `const { props, state } = this;`,
				Options: map[string]interface{}{"allowDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "thisDestructure"},
				},
			},
			{
				Code: `var unscoped = this;
function testFunction() {
  let inFunction = this;
}
const testLambda = () => {
  const inLambda = this;
};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "thisAssignment"},
					{MessageId: "thisAssignment"},
					{MessageId: "thisAssignment"},
				},
			},
			{
				Code: `class TestClass {
  constructor() {
    const inConstructor = this;
  }
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "thisAssignment"},
				},
			},
			{
				Code:    `const [foo] = this;`,
				Options: map[string]interface{}{"allowDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "thisDestructure"},
				},
			},
			{
				Code:    `object.value = this;`,
				Options: map[string]interface{}{"allowDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "thisDestructure"},
				},
			},
			{
				Code:    `let value; (value as unknown) = this;`,
				Options: map[string]interface{}{"allowDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "thisDestructure"},
				},
			},
			{
				Code:    `({ value } = (this));`,
				Options: map[string]interface{}{"allowDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "thisDestructure"},
				},
			},
		},
	)
}
