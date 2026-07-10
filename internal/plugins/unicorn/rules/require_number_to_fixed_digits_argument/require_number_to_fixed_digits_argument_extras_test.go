// TestRequireNumberToFixedDigitsArgumentExtras locks in branches and edge
// shapes that the upstream test suite doesn't exercise. Each case carries an
// inline comment pointing at the specific branch, Dimension 4 row, or real-user
// scenario it covers.
package require_number_to_fixed_digits_argument_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	require_number_to_fixed_digits_argument "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/require_number_to_fixed_digits_argument"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRequireNumberToFixedDigitsArgumentExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&require_number_to_fixed_digits_argument.RequireNumberToFixedDigitsArgumentRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: element-access key forms are excluded ----
			{Code: `number['toFixed']()`, FileName: "file.ts"},
			{Code: "number[`toFixed`]()", FileName: "file.ts"},
			{Code: `number[0]()`, FileName: "file.ts"},
			{Code: `number[Symbol.toFixed]()`, FileName: "file.ts"},

			// ---- Dimension 4: optional call is excluded ----
			{Code: `(number.toFixed)?.()`, FileName: "file.ts"},

			// Locks in upstream create() arm 2: a direct NewExpression receiver is excluded.
			{Code: `(new Number(1)).toFixed()`, FileName: "file.ts"},
			{Code: `((new Number(1))).toFixed()`, FileName: "file.ts"},

			// ---- Graceful degradation: spread is an argument ----
			{Code: `number.toFixed(...digits)`, FileName: "file.ts"},

			// N/A: private member access is invalid outside a declaring class, and #toFixed cannot name Number.prototype.toFixed.
			// N/A: declaration/container forms; the rule only targets call expressions.
			// N/A: declaration key forms (string, numeric, private, computed); no declarations are inspected.
			// N/A: same-kind nesting and ancestor walks; the rule performs no ancestor traversal.
			// N/A: body-absent declarations and empty containers; the rule only inspects call arguments.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized receiver wrappers ----
			{
				Code:     `(number).toFixed()`,
				FileName: "file.ts",
				Output:   []string{`(number).toFixed(0)`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1,
					Column:    17,
					EndLine:   1,
					EndColumn: 19,
				}},
			},
			{
				Code:     `((number)).toFixed()`,
				FileName: "file.ts",
				Output:   []string{`((number)).toFixed(0)`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1,
					Column:    19,
					EndLine:   1,
					EndColumn: 21,
				}},
			},

			// ---- Dimension 4: TypeScript receiver wrappers ----
			{
				Code:     `number!.toFixed()`,
				FileName: "file.ts",
				Output:   []string{`number!.toFixed(0)`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1,
					Column:    16,
					EndLine:   1,
					EndColumn: 18,
				}},
			},
			{
				Code:     `(number as number).toFixed()`,
				FileName: "file.ts",
				Output:   []string{`(number as number).toFixed(0)`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1,
					Column:    27,
					EndLine:   1,
					EndColumn: 29,
				}},
			},
			{
				Code:     `(number satisfies number).toFixed()`,
				FileName: "file.ts",
				Output:   []string{`(number satisfies number).toFixed(0)`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1,
					Column:    34,
					EndLine:   1,
					EndColumn: 36,
				}},
			},

			// ---- Dimension 4: optional member access remains reportable ----
			{
				Code:     `number?.toFixed()`,
				FileName: "file.ts",
				Output:   []string{`number?.toFixed(0)`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1,
					Column:    16,
					EndLine:   1,
					EndColumn: 18,
				}},
			},

			// Locks in upstream isMethodCall() argumentsLength === 0 with type arguments.
			{
				Code:     `number.toFixed<number>()`,
				FileName: "file.ts",
				Output:   []string{`number.toFixed<number>(0)`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1,
					Column:    23,
					EndLine:   1,
					EndColumn: 25,
				}},
			},

			// ---- Real-user: #1463 Number-like factory call remains reportable ----
			{
				Code:     `BigNumber(1).toFixed()`,
				FileName: "file.ts",
				Output:   []string{`BigNumber(1).toFixed(0)`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1,
					Column:    21,
					EndLine:   1,
					EndColumn: 23,
				}},
			},

			// ---- Real-user: #1601 only the immediate new receiver is exempt ----
			// Locks in upstream create() arm 2: a non-NewExpression receiver remains reportable.
			{
				Code:     `new BigNumber(1).plus(2).toFixed()`,
				FileName: "file.ts",
				Output:   []string{`new BigNumber(1).plus(2).toFixed(0)`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1,
					Column:    33,
					EndLine:   1,
					EndColumn: 35,
				}},
			},

			// Locks in report range and appendArgument behavior across lines/comments.
			{
				Code:     "number.toFixed(\n  /* keep */\n)",
				FileName: "file.ts",
				Output:   []string{"number.toFixed(\n  /* keep */\n0)"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1,
					Column:    15,
					EndLine:   3,
					EndColumn: 2,
				}},
			},
		},
	)
}
