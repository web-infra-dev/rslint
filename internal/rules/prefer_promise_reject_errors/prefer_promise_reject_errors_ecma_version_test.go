package prefer_promise_reject_errors

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferPromiseRejectErrorsECMAVersion(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferPromiseRejectErrorsRule,
		[]rule_tester.ValidTestCase{
			{Code: `Promise.reject(5)`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 5}},
			{Code: `new Promise(function(resolve, reject) { reject(5) })`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}, Globals: map[string]any{"Promise": "off"}},
			{Code: `function f(Promise) { Promise.reject(5); }`},
			{Code: `function f(undefined) { Promise.reject(undefined); }`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:            `Promise.reject(5)`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 5},
				Globals:         map[string]any{"Promise": "readonly"},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "rejectAnError"}},
			},
			{
				Code:            `new Promise(function(resolve, reject) { reject(5) })`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "rejectAnError"}},
			},
			{
				Code:   `Promise.reject(undefined)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "rejectAnError"}},
			},
		},
	)
}
