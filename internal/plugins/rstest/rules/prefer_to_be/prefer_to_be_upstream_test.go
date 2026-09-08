// TestPreferToBeUpstream migrates the full valid/invalid suite from upstream
// @vitest/eslint-plugin@v1.6.27 tests/prefer-to-be.test.ts 1:1. Position
// assertions cover line/column for every invalid case. Rstest-specific API
// sources, semantic corrections and edge-shape lock-ins live in
// prefer_to_be_extras_test.go.
package prefer_to_be

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferToBeUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferToBeRule,
		[]rule_tester.ValidTestCase{
			{Code: `expect(null).toBeNull();`},
			{Code: `expect(null).not.toBeNull();`},
			{Code: `expect(null).toBe(-1);`},
			{Code: `expect(null).toBe(1);`},
			{Code: `expect(obj).toStrictEqual([ x, 1 ]);`},
			{Code: `expect(obj).toStrictEqual({ x: 1 });`},
			{Code: `expect(obj).not.toStrictEqual({ x: 1 });`},
			{Code: `expect(value).toMatchSnapshot();`},
			{Code: `expect(catchError()).toStrictEqual({ message: 'oh noes!' })`},
			{Code: `expect("something");`},
			{Code: `expect("hey").to.be.a("string");`},
			{Code: `expect(token).toStrictEqual(/[abc]+/g);`},
			{Code: `expect(token).toStrictEqual(new RegExp('[abc]+', 'g'));`},
			{Code: `expect(0.1 + 0.2).toEqual(0.3);`},

			{Code: `expect(NaN).toBeNaN();`},
			{Code: `expect(true).not.toBeNaN();`},
			{Code: `expect({}).toEqual({});`},
			{Code: `expect(something).toBe()`},
			{Code: `expect(something).toBe(somethingElse)`},
			{Code: `expect(something).toEqual(somethingElse)`},
			{Code: `expect(something).not.toBe(somethingElse)`},
			{Code: `expect(something).not.toEqual(somethingElse)`},
			{Code: `expect(undefined).toBe`},
			{Code: `expect("something");`},

			{Code: `expect(null).toBeNull();`},
			{Code: `expect(null).not.toBeNull();`},
			{Code: `expect(null).toBe(1);`},
			{Code: `expect(obj).toStrictEqual([ x, 1 ]);`},
			{Code: `expect(obj).toStrictEqual({ x: 1 });`},
			{Code: `expect(obj).not.toStrictEqual({ x: 1 });`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `expect(value).toEqual("my string");`,
				Output: []string{`expect(value).toBe("my string");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 1, Column: 15}},
			},
			{
				Code:   `expect("a string").not.toEqual(null);`,
				Output: []string{`expect("a string").not.toBeNull();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeNull", Line: 1, Column: 24}},
			},
			{
				Code:   `expect("a string").not.toStrictEqual(null);`,
				Output: []string{`expect("a string").not.toBeNull();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeNull", Line: 1, Column: 24}},
			},

			{
				Code:   `expect(NaN).toBe(NaN);`,
				Output: []string{`expect(NaN).toBeNaN();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeNaN", Line: 1, Column: 13}},
			},
			{
				Code:   `expect("a string").not.toBe(NaN);`,
				Output: []string{`expect("a string").not.toBeNaN();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeNaN", Line: 1, Column: 24}},
			},
			{
				Code:   `expect("a string").not.toStrictEqual(NaN);`,
				Output: []string{`expect("a string").not.toBeNaN();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeNaN", Line: 1, Column: 24}},
			},

			{
				Code:   `expect(null).toBe(null);`,
				Output: []string{`expect(null).toBeNull();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeNull", Line: 1, Column: 14}},
			},
			{
				Code:   `expect(null).toEqual(null);`,
				Output: []string{`expect(null).toBeNull();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeNull", Line: 1, Column: 14}},
			},
			{
				Code:   `expect("a string").not.toEqual(null as number);`,
				Output: []string{`expect("a string").not.toBeNull();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeNull", Line: 1, Column: 24}},
			},
			{
				Code:   `expect(undefined).toBe(undefined as unknown as string as any);`,
				Output: []string{`expect(undefined).toBeUndefined();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeUndefined", Line: 1, Column: 19}},
			},
			{
				Code:   `expect("a string").toEqual(undefined as number);`,
				Output: []string{`expect("a string").toBeUndefined();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeUndefined", Line: 1, Column: 20}},
			},
		},
	)
}
