package error_message_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/error_message"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestErrorMessageECMAVersion(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&error_message.ErrorMessageRule,
		[]rule_tester.ValidTestCase{
			{Code: `new AggregateError([])`, FileName: "file.js", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020}},
			{Code: `new SuppressedError(error, suppressed)`, FileName: "file.js", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2025}},
			{Code: `new SuppressedError(error, suppressed)`, FileName: "file.js", Globals: map[string]any{"SuppressedError": "off"}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:            `new AggregateError([])`,
				FileName:        "file.js",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
				Globals:         map[string]any{"AggregateError": "readonly"},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "missing-message"}},
			},
			{
				Code:            `new SuppressedError(error, suppressed)`,
				FileName:        "file.js",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2026},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "missing-message"}},
			},
		},
	)
}
