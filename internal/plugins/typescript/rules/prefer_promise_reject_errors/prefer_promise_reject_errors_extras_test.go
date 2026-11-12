package prefer_promise_reject_errors

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Binding identity must include duplicate parameters and var redeclarations,
// while excluding shadowed names and functions that are not Promise executors.
// Expected diagnostics were independently verified with typescript-eslint v8.28.0.
func TestPreferPromiseRejectErrorsBindings(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferPromiseRejectErrorsRule, []rule_tester.ValidTestCase{
		{Code: "new Promise(function(reject, reject) { reject(new Error()); });"},
		{Code: "new Promise(function(reject, reject) { function inner(reject) { reject(5); } });"},
		{Code: "new Promise(function(resolve, reject) { { const reject = x => x; reject(5); } });"},
		{Code: "new Promise((resolve, reject) => resolve(5));"},
		{Code: "new Promise((resolve, reject = () => {}) => reject(5));"},
		{Code: "new Promise((resolve, ...reject) => reject(5));"},
		{Code: "new Promise(undefined, (resolve, reject) => reject(5));"},
		{Code: "new Promise((resolve, {reject}) => reject(5));"},
	}, []rule_tester.InvalidTestCase{
		{Code: "new Promise(function(reject, reject) { (() => reject(5))(); });",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "rejectAnError",
				Message:   "Expected the Promise rejection reason to be an Error.",
				Line:      1,
				Column:    47,
				EndLine:   1,
				EndColumn: 56,
			}},
		},
		{Code: "new Promise(function(resolve, reject, reject) { reject(5); });",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "rejectAnError",
				Message:   "Expected the Promise rejection reason to be an Error.",
				Line:      1,
				Column:    49,
				EndLine:   1,
				EndColumn: 58,
			}},
		},
		{Code: "new Promise(function(resolve, reject) { var reject; reject(5); });",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "rejectAnError",
				Message:   "Expected the Promise rejection reason to be an Error.",
				Line:      1,
				Column:    53,
				EndLine:   1,
				EndColumn: 62,
			}},
		},
		{Code: "new Promise((resolve, reject) => ((reject))(5));",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "rejectAnError",
				Message:   "Expected the Promise rejection reason to be an Error.",
				Line:      1,
				Column:    34,
				EndLine:   1,
				EndColumn: 47,
			}},
		},
		{Code: "new Promise((resolve, reject) => { /* 😀 */ reject(5); });",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "rejectAnError",
				Message:   "Expected the Promise rejection reason to be an Error.",
				Line:      1,
				Column:    45,
				EndLine:   1,
				EndColumn: 54,
			}},
		},
	})
}
