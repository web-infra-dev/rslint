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
			{Code: `const { props, state } = this;`},
			{Code: `const { length } = this;`},
			{Code: `const { length, toString } = this;`},
			{Code: `const [foo] = this;`},
			{Code: `const [foo, bar] = this;`},
			{Code: `const self = this;`, Options: map[string]interface{}{"allowedNames": []interface{}{"self"}}},
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
				Code: `let that; that = this;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "thisAssignment"},
				},
			},
			{
				Code: `const { props, state } = this;`,
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
				Code: `const [foo] = this;`,
				Options: map[string]interface{}{"allowDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "thisDestructure"},
				},
			},
		},
	)
}
