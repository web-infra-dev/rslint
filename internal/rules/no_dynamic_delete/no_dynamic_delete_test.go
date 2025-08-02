package no_dynamic_delete_test

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rule_tester"
	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
	"github.com/typescript-eslint/rslint/internal/rules/no_dynamic_delete"
)

func TestNoDynamicDelete(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_dynamic_delete.NoDynamicDeleteRule, []rule_tester.ValidTestCase{
		{
			Code: `
const container: { [i: string]: 0 } = {};
delete container.aaa;
`,
		},
		{
			Code: `
const container: { [i: string]: 0 } = {};
delete container.delete;
`,
		},
		{
			Code: `
const container: { [i: string]: 0 } = {};
delete container[7];
`,
		},
		{
			Code: `
const container: { [i: string]: 0 } = {};
delete container[-7];
`,
		},
		{
			Code: `
const container: { [i: string]: 0 } = {};
delete container['-Infinity'];
`,
		},
		{
			Code: `
const container: { [i: string]: 0 } = {};
delete container['+Infinity'];
`,
		},
		{
			Code: `
const value = 1;
delete value;
`,
		},
		{
			Code: `
const container: { [i: string]: 0 } = {};
delete container['aaa'];
`,
		},
		{
			Code: `
const container: { [i: string]: 0 } = {};
delete container['delete'];
`,
		},
		{
			Code: `
const container: { [i: string]: 0 } = {};
delete container['NaN'];
`,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
const container: { [i: string]: 0 } = {};
delete container['aa' + 'b'];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "dynamicDelete",
				},
			},
		},
		{
			Code: `
const container: { [i: string]: 0 } = {};
delete container[+7];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "dynamicDelete",
				},
			},
		},
		{
			Code: `
const container: { [i: string]: 0 } = {};
delete container[-Infinity];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "dynamicDelete",
				},
			},
		},
		{
			Code: `
const container: { [i: string]: 0 } = {};
delete container[+Infinity];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "dynamicDelete",
				},
			},
		},
		{
			Code: `
const container: { [i: string]: 0 } = {};
delete container[NaN];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "dynamicDelete",
				},
			},
		},
		{
			Code: `
const container: { [i: string]: 0 } = {};
const name = 'name';
delete container[name];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "dynamicDelete",
				},
			},
		},
		{
			Code: `
const container: { [i: string]: 0 } = {};
const getName = () => 'aaa';
delete container[getName()];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "dynamicDelete",
				},
			},
		},
		{
			Code: `
const container: { [i: string]: 0 } = {};
const name = { foo: { bar: 'bar' } };
delete container[name.foo.bar];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "dynamicDelete",
				},
			},
		},
		{
			Code: `
const container: { [i: string]: 0 } = {};
delete container[+'Infinity'];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "dynamicDelete",
				},
			},
		},
		{
			Code: `
const container: { [i: string]: 0 } = {};
delete container[typeof 1];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "dynamicDelete",
				},
			},
		},
	})
}