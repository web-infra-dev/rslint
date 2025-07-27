package no_unnecessary_type_parameters

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rule_tester"
	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
)

func TestNoUnnecessaryTypeParametersRule(t *testing.T) {
	validTestCases := []rule_tester.ValidTestCase{
		{
			Code: `class ClassyArray<T> {
					arr: T[];
				}`,
		},
		{
			Code: `class ClassyArray<T> {
					value1: T;
					value2: T;
				}`,
		},
		{
			Code: `function identity<T>(arg: T): T {
					return arg;
				}`,
		},
		{
			Code: `function getProperty<T, K extends keyof T>(obj: T, key: K) {
					return obj[key];
				}`,
		},
		{
			Code: `type Fn = <T>(input: T) => T;`,
		},
		{
			Code: `type Fn = <T>(input: T) => Partial<T>;`,
		},
		{
			Code: `function both<Args extends unknown[]>(
				fn1: (...args: Args) => void,
				fn2: (...args: Args) => void,
			): (...args: Args) => void {
				return function (...args: Args) {
					fn1(...args);
					fn2(...args);
				};
				}`,
		},
		{
			Code: `declare function makeMap<K, V>(): Map<K, V>;`,
		},
		{
			Code: `declare function compare<T>(param1: T, param2: T): boolean;`,
		},
	}

	invalidTestCases := []rule_tester.InvalidTestCase{
		{
			Code: `const func = <T,>(param: T) => null;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "sole",
					Line: 1,
					Column: 15,
				},
			},
		},
		{
			Code: `const f1 = <T,>(): T => {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "sole",
					Line: 1,
					Column: 13,
				},
			},
		},
		{
			Code: `function third<A, B, C>(a: A, b: B, c: C): C {
					return c;
				}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "sole",
					Line: 1,
					Column: 16,
				},
				{
					MessageId: "sole",
					Line: 1,
					Column: 18,
				},
			},
		},
		{
			Code: `class Joiner<T extends string | number> {
					join(el: T, other: string) {
						return [el, other].join(',');
					}
				}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "sole",
					Line: 1,
					Column: 14,
				},
			},
		},
		{
			Code: `declare function get<T>(): T;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "sole",
					Line: 1,
					Column: 22,
				},
			},
		},
		{
			Code: `declare function take<T>(param: T): void;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "sole",
					Line: 1,
					Column: 23,
				},
			},
		},
		{
			Code: `type Fn = <U>(param: U) => void;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "sole",
					Line: 1,
					Column: 12,
				},
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryTypeParametersRule, validTestCases, invalidTestCases)
}