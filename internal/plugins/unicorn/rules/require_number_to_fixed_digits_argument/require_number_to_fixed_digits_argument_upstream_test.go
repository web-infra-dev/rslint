// TestRequireNumberToFixedDigitsArgumentUpstream migrates the full valid/invalid
// suite from upstream test/require-number-to-fixed-digits-argument.js 1:1.
// Position assertions cover line/column for every invalid case. rslint-specific
// lock-in cases live in require_number_to_fixed_digits_argument_extras_test.go.
package require_number_to_fixed_digits_argument_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	require_number_to_fixed_digits_argument "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/require_number_to_fixed_digits_argument"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	messageID = "require-number-to-fixed-digits-argument"
	message   = "Missing the digits argument."
)

func TestRequireNumberToFixedDigitsArgumentUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&require_number_to_fixed_digits_argument.RequireNumberToFixedDigitsArgumentRule,
		[]rule_tester.ValidTestCase{
			{Code: `number.toFixed(0)`, FileName: "file.js"},
			{Code: `number.toFixed(...[])`, FileName: "file.js"},
			{Code: `number.toFixed(2)`, FileName: "file.js"},
			{Code: `number.toFixed(1,2,3)`, FileName: "file.js"},
			{Code: `number[toFixed]()`, FileName: "file.js"},
			{Code: `number["toFixed"]()`, FileName: "file.js"},
			{Code: `number.toFixed?.()`, FileName: "file.js"},
			{Code: `number.notToFixed();`, FileName: "file.js"},

			// `callee` object is a NewExpression.
			{Code: `new BigNumber(1).toFixed()`, FileName: "file.js"},
			{Code: `new Number(1).toFixed()`, FileName: "file.js"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     `const string = number.toFixed();`,
				FileName: "file.js",
				Output:   []string{`const string = number.toFixed(0);`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   message,
					Line:      1,
					Column:    30,
					EndLine:   1,
					EndColumn: 32,
				}},
			},
			{
				Code:     `const string = number?.toFixed() ?? "";`,
				FileName: "file.js",
				Output:   []string{`const string = number?.toFixed(0) ?? "";`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1,
					Column:    31,
					EndLine:   1,
					EndColumn: 33,
				}},
			},
			{
				Code:     `const string = number.toFixed( /* comment */ );`,
				FileName: "file.js",
				Output:   []string{`const string = number.toFixed( /* comment */ 0);`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1,
					Column:    30,
					EndLine:   1,
					EndColumn: 47,
				}},
			},
			{
				Code:     `Number(1).toFixed()`,
				FileName: "file.js",
				Output:   []string{`Number(1).toFixed(0)`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 20,
				}},
			},
			{
				Code:     `const bigNumber = new BigNumber(1); const string = bigNumber.toFixed();`,
				FileName: "file.js",
				Output:   []string{`const bigNumber = new BigNumber(1); const string = bigNumber.toFixed(0);`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1,
					Column:    69,
					EndLine:   1,
					EndColumn: 71,
				}},
			},
		},
	)
}
