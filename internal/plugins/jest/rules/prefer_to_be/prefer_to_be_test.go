package prefer_to_be_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_to_be"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferToBeRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_to_be.PreferToBeRule,
		[]rule_tester.ValidTestCase{
			// prefer-to-be
			{Code: `expect(null).toBeNull();`},
			{Code: `expect(null).not.toBeNull();`},
			{Code: `expect(null).toBe(1);`},
			{Code: `expect(null).toBe(-1);`},
			{Code: `expect(null).toBe(...1);`},
			{Code: `expect(obj).toStrictEqual([ x, 1 ]);`},
			{Code: `expect(obj).toStrictEqual({ x: 1 });`},
			{Code: `expect(obj).not.toStrictEqual({ x: 1 });`},
			{Code: `expect(value).toMatchSnapshot();`},
			{Code: `expect(catchError()).toStrictEqual({ message: 'oh noes!' })`},
			{Code: `expect("something");`},
			{Code: `expect(token).toStrictEqual(/[abc]+/g);`},
			{Code: `expect(token).toStrictEqual(new RegExp('[abc]+', 'g'));`},
			{Code: "expect(value).toEqual(dedent`my string`);"},
			// prefer-to-be: null
			{Code: `expect(null).not.toEqual();`},
			{Code: `expect(null).toBe();`},
			{Code: `expect(null).toMatchSnapshot();`},
			{Code: `expect("a string").toMatchSnapshot(null);`},
			{Code: `expect("a string").not.toMatchSnapshot();`},
			{Code: `expect(null).toBe`},
			// prefer-to-be: undefined
			{Code: `expect(undefined).toBeUndefined();`},
			{Code: `expect(true).toBeDefined();`},
			{Code: `expect({}).toEqual({});`},
			{Code: `expect(something).toBe()`},
			{Code: `expect(something).toBe(somethingElse)`},
			{Code: `expect(something).toEqual(somethingElse)`},
			{Code: `expect(something).not.toBe(somethingElse)`},
			{Code: `expect(something).not.toEqual(somethingElse)`},
			{Code: `expect(undefined).toBe`},
			// prefer-to-be: NaN
			{Code: `expect(NaN).toBeNaN();`},
			{Code: `expect(true).not.toBeNaN();`},
			{Code: `expect(value).toEqual(null!);`},
			{Code: `expect(value).toEqual(1 satisfies number);`},
			// prefer-to-be: typescript edition
			{Code: `(expect('Model must be bound to an array if the multiple property is true') as any).toHaveBeenTipped()`},
		},
		[]rule_tester.InvalidTestCase{
			// prefer-to-be
			{
				Code:   `expect(value).toEqual("my string");`,
				Output: []string{`expect(value).toBe("my string");`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBe", Line: 1, Column: 15},
				},
			},
			{
				Code:   `expect(value).toStrictEqual("my string");`,
				Output: []string{`expect(value).toBe("my string");`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBe", Line: 1, Column: 15},
				},
			},
			{
				Code:   `expect(value).toStrictEqual(1);`,
				Output: []string{`expect(value).toBe(1);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBe", Line: 1, Column: 15},
				},
			},
			{
				Code:   `expect(value).toStrictEqual(1,);`,
				Output: []string{`expect(value).toBe(1,);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBe", Line: 1, Column: 15},
				},
			},
			{
				Code:   `expect(value).toEqual((1));`,
				Output: []string{`expect(value).toBe((1));`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBe", Line: 1, Column: 15},
				},
			},
			{
				Code:   `expect(value).toStrictEqual(-1);`,
				Output: []string{`expect(value).toBe(-1);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBe", Line: 1, Column: 15},
				},
			},
			{
				Code:   "expect(value).toEqual(`my string`);",
				Output: []string{"expect(value).toBe(`my string`);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBe", Line: 1, Column: 15},
				},
			},
			{
				Code:   "expect(value)[\"toEqual\"](`my string`);",
				Output: []string{"expect(value)['toBe'](`my string`);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBe", Line: 1, Column: 15},
				},
			},
			{
				Code:   "expect(value).toStrictEqual(`my ${string}`);",
				Output: []string{"expect(value).toBe(`my ${string}`);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBe", Line: 1, Column: 15},
				},
			},
			{
				Code:   `expect(loadMessage()).resolves.toStrictEqual("hello world");`,
				Output: []string{`expect(loadMessage()).resolves.toBe("hello world");`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBe", Line: 1, Column: 32},
				},
			},
			{
				Code:   `expect(loadMessage()).resolves["toStrictEqual"]("hello world");`,
				Output: []string{`expect(loadMessage()).resolves['toBe']("hello world");`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBe", Line: 1, Column: 32},
				},
			},
			{
				Code:   `expect(loadMessage())["resolves"].toStrictEqual("hello world");`,
				Output: []string{`expect(loadMessage())["resolves"].toBe("hello world");`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBe", Line: 1, Column: 35},
				},
			},
			{
				Code:   `expect(loadMessage()).resolves.toStrictEqual(false);`,
				Output: []string{`expect(loadMessage()).resolves.toBe(false);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBe", Line: 1, Column: 32},
				},
			},

			// prefer-to-be: null
			{
				Code:   `expect(null).toBe(null);`,
				Output: []string{`expect(null).toBeNull();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeNull", Line: 1, Column: 14},
				},
			},
			{
				Code:   `expect(null).toEqual(null);`,
				Output: []string{`expect(null).toBeNull();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeNull", Line: 1, Column: 14},
				},
			},
			{
				Code:   `expect(null).toEqual(null,);`,
				Output: []string{`expect(null).toBeNull();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeNull", Line: 1, Column: 14},
				},
			},
			{
				Code:   `expect(null).toStrictEqual(null);`,
				Output: []string{`expect(null).toBeNull();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeNull", Line: 1, Column: 14},
				},
			},
			{
				Code:   `expect("a string").not.toBe(null);`,
				Output: []string{`expect("a string").not.toBeNull();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeNull", Line: 1, Column: 24},
				},
			},
			{
				Code:   `expect("a string").not["toBe"](null);`,
				Output: []string{`expect("a string").not['toBeNull']();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeNull", Line: 1, Column: 24},
				},
			},
			{
				Code:   `expect("a string")["not"]["toBe"](null);`,
				Output: []string{`expect("a string")["not"]['toBeNull']();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeNull", Line: 1, Column: 27},
				},
			},
			{
				Code:   `expect("a string").not.toEqual(null);`,
				Output: []string{`expect("a string").not.toBeNull();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeNull", Line: 1, Column: 24},
				},
			},
			{
				Code:   `expect("a string").not.toStrictEqual(null);`,
				Output: []string{`expect("a string").not.toBeNull();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeNull", Line: 1, Column: 24},
				},
			},

			// prefer-to-be: undefined
			{
				Code:   `expect(undefined).toBe(undefined);`,
				Output: []string{`expect(undefined).toBeUndefined();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeUndefined", Line: 1, Column: 19},
				},
			},
			{
				Code:   `expect(undefined).toEqual(undefined);`,
				Output: []string{`expect(undefined).toBeUndefined();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeUndefined", Line: 1, Column: 19},
				},
			},
			{
				Code:   `expect(undefined).toStrictEqual(undefined);`,
				Output: []string{`expect(undefined).toBeUndefined();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeUndefined", Line: 1, Column: 19},
				},
			},
			{
				Code:   `expect("a string").not.toBe(undefined);`,
				Output: []string{`expect("a string").toBeDefined();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeDefined", Line: 1, Column: 24},
				},
			},
			{
				Code:   `expect("a string").rejects.not.toBe(undefined);`,
				Output: []string{`expect("a string").rejects.toBeDefined();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeDefined", Line: 1, Column: 32},
				},
			},
			{
				Code:   `expect("a string").rejects.not["toBe"](undefined);`,
				Output: []string{`expect("a string").rejects['toBeDefined']();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeDefined", Line: 1, Column: 32},
				},
			},
			{
				Code:   `expect("a string")["not"]["toBe"](undefined);`,
				Output: []string{`expect("a string")['toBeDefined']();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeDefined", Line: 1, Column: 27},
				},
			},
			{
				Code:   `expect("a string").not.toEqual(undefined);`,
				Output: []string{`expect("a string").toBeDefined();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeDefined", Line: 1, Column: 24},
				},
			},
			{
				Code:   `expect("a string").not.toStrictEqual(undefined);`,
				Output: []string{`expect("a string").toBeDefined();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeDefined", Line: 1, Column: 24},
				},
			},

			// prefer-to-be: NaN
			{
				Code:   `expect(NaN).toBe(NaN);`,
				Output: []string{`expect(NaN).toBeNaN();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeNaN", Line: 1, Column: 13},
				},
			},
			{
				Code:   `expect(NaN).toEqual(NaN);`,
				Output: []string{`expect(NaN).toBeNaN();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeNaN", Line: 1, Column: 13},
				},
			},
			{
				Code:   `expect(NaN).toStrictEqual(NaN);`,
				Output: []string{`expect(NaN).toBeNaN();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeNaN", Line: 1, Column: 13},
				},
			},
			{
				Code:   `expect("a string").not.toBe(NaN);`,
				Output: []string{`expect("a string").not.toBeNaN();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeNaN", Line: 1, Column: 24},
				},
			},
			{
				Code:   `expect("a string").rejects.not.toBe(NaN);`,
				Output: []string{`expect("a string").rejects.not.toBeNaN();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeNaN", Line: 1, Column: 32},
				},
			},
			{
				Code:   `expect("a string")["rejects"].not.toBe(NaN);`,
				Output: []string{`expect("a string")["rejects"].not.toBeNaN();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeNaN", Line: 1, Column: 35},
				},
			},
			{
				Code:   `expect("a string").not.toEqual(NaN);`,
				Output: []string{`expect("a string").not.toBeNaN();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeNaN", Line: 1, Column: 24},
				},
			},
			{
				Code:   `expect("a string").not.toStrictEqual(NaN);`,
				Output: []string{`expect("a string").not.toBeNaN();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeNaN", Line: 1, Column: 24},
				},
			},

			// prefer-to-be: undefined vs defined
			{
				Code:   `expect(undefined).not.toBeDefined();`,
				Output: []string{`expect(undefined).toBeUndefined();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeUndefined", Line: 1, Column: 23},
				},
			},
			{
				Code:   `expect(undefined).resolves.not.toBeDefined();`,
				Output: []string{`expect(undefined).resolves.toBeUndefined();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeUndefined", Line: 1, Column: 32},
				},
			},
			{
				Code:   `expect(undefined).resolves.toBe(undefined);`,
				Output: []string{`expect(undefined).resolves.toBeUndefined();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeUndefined", Line: 1, Column: 28},
				},
			},
			{
				Code:   `expect("a string").not.toBeUndefined();`,
				Output: []string{`expect("a string").toBeDefined();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeDefined", Line: 1, Column: 24},
				},
			},
			{
				Code:   `expect("a string").rejects.not.toBeUndefined();`,
				Output: []string{`expect("a string").rejects.toBeDefined();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeDefined", Line: 1, Column: 32},
				},
			},

			// prefer-to-be: typescript edition
			{
				Code:   `expect(null).toEqual(1 as unknown as string as unknown as any);`,
				Output: []string{`expect(null).toBe(1 as unknown as string as unknown as any);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBe", Line: 1, Column: 14},
				},
			},
			{
				Code:   `expect(null).toEqual(-1 as unknown as string as unknown as any);`,
				Output: []string{`expect(null).toBe(-1 as unknown as string as unknown as any);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBe", Line: 1, Column: 14},
				},
			},
			{
				Code:   `expect("a string").not.toStrictEqual("string" as number);`,
				Output: []string{`expect("a string").not.toBe("string" as number);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBe", Line: 1, Column: 24},
				},
			},
			{
				Code:   `expect(null).toBe(null as unknown as string as unknown as any);`,
				Output: []string{`expect(null).toBeNull();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeNull", Line: 1, Column: 14},
				},
			},
			{
				Code:   `expect("a string").not.toEqual(null as number);`,
				Output: []string{`expect("a string").not.toBeNull();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeNull", Line: 1, Column: 24},
				},
			},
			{
				Code:   `expect(undefined).toBe(undefined as unknown as string as any);`,
				Output: []string{`expect(undefined).toBeUndefined();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeUndefined", Line: 1, Column: 19},
				},
			},
			{
				Code:   `expect("a string").toEqual(undefined as number);`,
				Output: []string{`expect("a string").toBeUndefined();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeUndefined", Line: 1, Column: 20},
				},
			},
			{
				Code:   `expect(value).toEqual((null));`,
				Output: []string{`expect(value).toBeNull();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeNull", Line: 1, Column: 15},
				},
			},
			{
				Code:   `expect(value).toEqual((NaN));`,
				Output: []string{`expect(value).toBeNaN();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeNaN", Line: 1, Column: 15},
				},
			},
			{
				Code:   `expect(value).toEqual((undefined));`,
				Output: []string{`expect(value).toBeUndefined();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBeUndefined", Line: 1, Column: 15},
				},
			},
		},
	)
}
