package no_unnecessary_type_constraint

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoUnnecessaryTypeConstraintRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryTypeConstraintRule, []rule_tester.ValidTestCase{
		{Code: `function data() {}`},
		{Code: `function data<T>() {}`},
		{Code: `function data<T, U>() {}`},
		{Code: `function data<T extends number>() {}`},
		{Code: `function data<T extends number | string>() {}`},
		{Code: `function data<T extends any | number>() {}`},
		{Code: `
type TODO = any;
function data<T extends TODO>() {}`},
		{Code: `const data = () => {};`},
		{Code: `const data = <T,>() => {};`},
		{Code: `const data = <T, U>() => {};`},
		{Code: `const data = <T extends number>() => {};`},
		{Code: `const data = <T extends number | string>() => {};`},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `function data<T extends any>() {}`,
			Output: []string{`function data<T>() {}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line: 1,
					Column: 15,
				},
			},
		},
		{
			Code: `function data<T extends any, U>() {}`,
			Output: []string{`function data<T, U>() {}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line: 1,
					Column: 15,
				},
			},
		},
		{
			Code: `function data<T, U extends any>() {}`,
			Output: []string{`function data<T, U>() {}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line: 1,
					Column: 18,
				},
			},
		},
		{
			Code: `function data<T extends any, U extends T>() {}`,
			Output: []string{`function data<T, U extends T>() {}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line: 1,
					Column: 15,
				},
			},
		},
		{
			Code: `const data = <T extends any>() => {};`,
			Output: []string{`const data = <T>() => {};`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line: 1,
					Column: 15,
				},
			},
		},
		{
			Code: `const data = <T extends any>() => {};`,
			Output: []string{`const data = <T,>() => {};`},
			Filename: "test.tsx",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line: 1,
					Column: 15,
				},
			},
		},
		{
			Code: `const data = <T extends any = unknown>() => {};`,
			Output: []string{`const data = <T = unknown>() => {};`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line: 1,
					Column: 15,
				},
			},
		},
		{
			Code: `const data = <T extends any, U extends any>() => {};`,
			Output: []string{`const data = <T, U>() => {};`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line: 1,
					Column: 15,
				},
				{
					MessageId: "unnecessaryConstraint",
					Line: 1,
					Column: 30,
				},
			},
		},
		{
			Code: `function data<T extends unknown>() {}`,
			Output: []string{`function data<T>() {}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line: 1,
					Column: 15,
				},
			},
		},
		{
			Code: `const data = <T extends unknown>() => {};`,
			Output: []string{`const data = <T>() => {};`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line: 1,
					Column: 15,
				},
			},
		},
		{
			Code: `class Data<T extends unknown> {}`,
			Output: []string{`class Data<T> {}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line: 1,
					Column: 12,
				},
			},
		},
		{
			Code: `const Data = class<T extends unknown> {};`,
			Output: []string{`const Data = class<T> {};`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line: 1,
					Column: 20,
				},
			},
		},
		{
			Code: `
class Data {
  member<T extends unknown>() {}
}`,
			Output: []string{`
class Data {
  member<T>() {}
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line: 3,
					Column: 10,
				},
			},
		},
		{
			Code: `
const Data = class {
  member<T extends unknown>() {}
};`,
			Output: []string{`
const Data = class {
  member<T>() {}
};`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line: 3,
					Column: 10,
				},
			},
		},
		{
			Code: `interface Data<T extends unknown> {}`,
			Output: []string{`interface Data<T> {}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line: 1,
					Column: 16,
				},
			},
		},
		{
			Code: `type Data<T extends unknown> = {};`,
			Output: []string{`type Data<T> = {};`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line: 1,
					Column: 11,
				},
			},
		},
	})
}