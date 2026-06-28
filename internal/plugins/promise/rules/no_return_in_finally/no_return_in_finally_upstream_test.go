// TestNoReturnInFinallyUpstream migrates the full valid/invalid suite from
// upstream __tests__/no-return-in-finally.js 1:1. Position assertions cover
// line/column for every invalid case. rslint-specific lock-in cases live in
// the no_return_in_finally_extras_test.go file.
package no_return_in_finally_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/no_return_in_finally"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const noReturnMsg = "No return in finally"

func TestNoReturnInFinallyUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_return_in_finally.NoReturnInFinallyRule,
		[]rule_tester.ValidTestCase{
			// ---- valid: no return in finally callback ----
			{Code: `Promise.resolve(1).finally(() => { console.log(2) })`},
			{Code: `Promise.reject(4).finally(() => { console.log(2) })`},
			{Code: `Promise.reject(4).finally(() => {})`},
			{Code: `myPromise.finally(() => {});`},
			{Code: `Promise.resolve(1).finally(function () { })`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- invalid: return inside finally callback ----
			{
				Code:   `Promise.resolve(1).finally(() => { return 2 })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noReturnInFinally", Message: noReturnMsg, Line: 1, Column: 20}},
			},
			{
				Code:   `Promise.reject(0).finally(() => { return 2 })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noReturnInFinally", Message: noReturnMsg, Line: 1, Column: 19}},
			},
			{
				Code:   `myPromise.finally(() => { return 2 });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noReturnInFinally", Message: noReturnMsg, Line: 1, Column: 11}},
			},
			{
				Code:   `Promise.resolve(1).finally(function () { return 2 })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noReturnInFinally", Message: noReturnMsg, Line: 1, Column: 20}},
			},
		},
	)
}
