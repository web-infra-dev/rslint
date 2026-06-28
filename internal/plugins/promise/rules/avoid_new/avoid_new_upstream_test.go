// TestAvoidNewUpstream migrates the full valid/invalid suite from upstream
// __tests__/avoid-new.js 1:1. Position assertions cover line/column for every
// invalid case. rslint-specific lock-in cases live in avoid_new_extras_test.go.

package avoid_new_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/avoid_new"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const avoidNewMessage = "Avoid creating new promises."

func TestAvoidNewUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&avoid_new.AvoidNewRule,
		[]rule_tester.ValidTestCase{
			{Code: `Promise.resolve()`},
			{Code: `Promise.reject()`},
			{Code: `Promise.all()`},
			{Code: `new Horse()`},
			{Code: `new PromiseLikeThing()`},
			// callee is a member expression — not a plain identifier named "Promise"
			{Code: `new Promise.resolve()`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `var x = new Promise(function (x, y) {})`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNew", Message: avoidNewMessage, Line: 1, Column: 9}},
			},
			{
				Code:   `new Promise()`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNew", Message: avoidNewMessage, Line: 1, Column: 1}},
			},
			{
				Code:   `Thing(new Promise(() => {}))`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNew", Message: avoidNewMessage, Line: 1, Column: 7}},
			},
		},
	)
}
