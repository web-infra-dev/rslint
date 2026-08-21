// TestNoDivRegexUpstream migrates the full valid/invalid suite from upstream
// eslint/tests/lib/rules/no-div-regex.js 1:1, plus the incorrect/correct
// examples shown on the rule's documentation page. Position assertions cover
// line/column for every invalid case. rslint-specific lock-in cases live in
// the no_div_regex_extras_test.go file.
package no_div_regex

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDivRegexUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDivRegexRule,
		[]rule_tester.ValidTestCase{
			// ---- valid ----
			{Code: `var f = function() { return /foo/ig.test('bar'); };`},
			{Code: `var f = function() { return /\=foo/; };`},
			// ---- Doc examples: correct ----
			{Code: `function bar() { return /[=]foo/; }`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- invalid ----
			{
				Code:   `var f = function() { return /=foo/; };`,
				Output: []string{`var f = function() { return /[=]foo/; };`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Message: "A regular expression literal can be confused with '/='.", Line: 1, Column: 29, EndLine: 1, EndColumn: 35},
				},
			},
			// ---- Doc examples: incorrect ----
			{
				Code:   `function bar() { return /=foo/; }`,
				Output: []string{`function bar() { return /[=]foo/; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 25, EndLine: 1, EndColumn: 31},
				},
			},
		},
	)
}
