// TestPreferExpectTypeOfUpstream migrates the complete
// @vitest/eslint-plugin@v1.6.27 prefer-expect-type-of suite. Rstest-only edge
// cases, source matrices, fix-trivia coverage and edit-demand parity live in
// prefer_expect_type_of_extras_test.go.
package prefer_expect_type_of

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferExpectTypeOfUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferExpectTypeOfRule,
		[]rule_tester.ValidTestCase{
			{Code: `expect("name").toBeTypeOf("string")`},
			{Code: `expect("name").not.toBeTypeOf("string")`},
			{Code: `expect(12).toBeTypeOf("number")`},
			{Code: `expect(true).toBeTypeOf("boolean")`},
			{Code: `expect({a: 1}).toBeTypeOf("object")`},
			{Code: `expect(() => {}).toBeTypeOf("function")`},
			{Code: `expect(sym).toBeTypeOf("symbol")`},
			{Code: `expect(BigInt(123)).toBeTypeOf("bigint")`},
			{Code: `expect(undefined).toBeTypeOf("undefined")`},
			{Code: `expect(value).not.toBe(42)`},
			{Code: `expect(value).not.toEqual(42)`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `expect(typeof 12).toBe("number")`,
				Output: []string{`expect(12).toBeTypeOf("number")`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferExpectTypeOf",
					Message:   "Use `expect(12).toBeTypeOf(\"number\")` instead of `expect(typeof 12).toBe(\"number\")`",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 33,
				}},
			},
			{
				Code:   `expect(typeof "name").toBe("string")`,
				Output: []string{`expect("name").toBeTypeOf("string")`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferExpectTypeOf",
					Message:   "Use `expect(\"name\").toBeTypeOf(\"string\")` instead of `expect(typeof \"name\").toBe(\"string\")`",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 37,
				}},
			},
			{
				Code:   `expect(typeof true).toBe("boolean")`,
				Output: []string{`expect(true).toBeTypeOf("boolean")`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferExpectTypeOf",
					Message:   "Use `expect(true).toBeTypeOf(\"boolean\")` instead of `expect(typeof true).toBe(\"boolean\")`",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 36,
				}},
			},
			{
				Code:   `expect(typeof variable).toBe("object")`,
				Output: []string{`expect(variable).toBeTypeOf("object")`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferExpectTypeOf",
					Message:   "Use `expect(variable).toBeTypeOf(\"object\")` instead of `expect(typeof variable).toBe(\"object\")`",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 39,
				}},
			},
			{
				Code:   `expect(typeof fn).toBe("function")`,
				Output: []string{`expect(fn).toBeTypeOf("function")`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferExpectTypeOf",
					Message:   "Use `expect(fn).toBeTypeOf(\"function\")` instead of `expect(typeof fn).toBe(\"function\")`",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 35,
				}},
			},
			{
				Code:   `expect(typeof sym).toBe("symbol")`,
				Output: []string{`expect(sym).toBeTypeOf("symbol")`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferExpectTypeOf",
					Message:   "Use `expect(sym).toBeTypeOf(\"symbol\")` instead of `expect(typeof sym).toBe(\"symbol\")`",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 34,
				}},
			},
			{
				Code:   `expect(typeof big).toBe("bigint")`,
				Output: []string{`expect(big).toBeTypeOf("bigint")`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferExpectTypeOf",
					Message:   "Use `expect(big).toBeTypeOf(\"bigint\")` instead of `expect(typeof big).toBe(\"bigint\")`",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 34,
				}},
			},
			{
				Code:   `expect(typeof value).toBe("undefined")`,
				Output: []string{`expect(value).toBeTypeOf("undefined")`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferExpectTypeOf",
					Message:   "Use `expect(value).toBeTypeOf(\"undefined\")` instead of `expect(typeof value).toBe(\"undefined\")`",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 39,
				}},
			},
			{
				Code:   `expect(typeof value).toEqual("string")`,
				Output: []string{`expect(value).toBeTypeOf("string")`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferExpectTypeOf",
					Message:   "Use `expect(value).toBeTypeOf(\"string\")` instead of `expect(typeof value).toBe(\"string\")`",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 39,
				}},
			},
			{
				Code:   `expect(typeof value).not.toBe("string")`,
				Output: []string{`expect(value).not.toBeTypeOf("string")`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferExpectTypeOf",
					Message:   "Use `expect(value).toBeTypeOf(\"string\")` instead of `expect(typeof value).toBe(\"string\")`",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 40,
				}},
			},
			{
				Code:   `expect(typeof value).toBe("unknown")`,
				Output: []string{`expect(value).toBeTypeOf("unknown")`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferExpectTypeOf",
					Message:   "Use `expect(value).toBeTypeOf(\"unknown\")` instead of `expect(typeof value).toBe(\"unknown\")`",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 37,
				}},
			},
			{
				Code:   `expect(typeof value).toBe(typeName)`,
				Output: []string{`expect(value).toBeTypeOf(typeName)`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferExpectTypeOf",
					Message:   "Use `expect(value).toBeTypeOf(typeName)` instead of `expect(typeof value).toBe(typeName)`",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 36,
				}},
			},
		},
	)
}
