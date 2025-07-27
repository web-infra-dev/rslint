package no_this_alias

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rule_tester"
	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
)

func TestNoThisAlias(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoThisAliasRule,
		[]rule_tester.ValidTestCase{
		// Valid cases
		{Code: "const self = foo(this);"},
		{
			Code: `
const { props, state } = this;
const { length } = this;
const { length, toString } = this;
const [foo] = this;
const [foo, bar] = this;
`,
			Options: map[string]interface{}{
				"allowDestructuring": true,
			},
		},
		{
			Code: "const self = this;",
			Options: map[string]interface{}{
				"allowedNames": []interface{}{"self"},
			},
		},
		{
			Code: `
declare module 'foo' {
  declare const aVar: string;
}
`,
		},
	}, []rule_tester.InvalidTestCase{
		// Invalid cases
		{
			Code: "const self = this;",
			Options: map[string]interface{}{
				"allowDestructuring": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "thisAssignment",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code: "const self = this;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "thisAssignment",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code: `
let that;
that = this;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "thisAssignment",
					Line:      3,
					Column:    1,
				},
			},
		},
		{
			Code: "const { props, state } = this;",
			Options: map[string]interface{}{
				"allowDestructuring": false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "thisDestructure",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code: `
var unscoped = this;

function testFunction() {
  let inFunction = this;
}
const testLambda = () => {
  const inLambda = this;
};
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "thisAssignment",
					Line:      2,
					Column:    5,
				},
				{
					MessageId: "thisAssignment",
					Line:      5,
					Column:    7,
				},
				{
					MessageId: "thisAssignment",
					Line:      8,
					Column:    9,
				},
			},
		},
		{
			Code: `
class TestClass {
  constructor() {
    const inConstructor = this;
    const asThis: this = this;

    const asString = 'this';
    const asArray = [this];
    const asArrayString = ['this'];
  }

  public act(scope: this = this) {
    const inMemberFunction = this;
    const { act } = this;
    const { act, constructor } = this;
    const [foo] = this;
    const [foo, bar] = this;
  }
}
`,
			Options: map[string]interface{}{
				"allowDestructuring": false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "thisAssignment",
					Line:      4,
					Column:    11,
				},
				{
					MessageId: "thisAssignment",
					Line:      5,
					Column:    11,
				},
				{
					MessageId: "thisAssignment",
					Line:      13,
					Column:    11,
				},
				{
					MessageId: "thisDestructure",
					Line:      14,
					Column:    11,
				},
				{
					MessageId: "thisDestructure",
					Line:      15,
					Column:    11,
				},
				{
					MessageId: "thisDestructure",
					Line:      16,
					Column:    11,
				},
				{
					MessageId: "thisDestructure",
					Line:      17,
					Column:    11,
				},
			},
		},
	})
}