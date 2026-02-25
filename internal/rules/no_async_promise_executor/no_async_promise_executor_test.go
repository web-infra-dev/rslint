package no_async_promise_executor

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoAsyncPromiseExecutorRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoAsyncPromiseExecutorRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Regular Promise executors (not async)
			{Code: `new Promise((resolve, reject) => {})`},
			{Code: `new Promise((resolve, reject) => {}, async function unrelated() {})`},
			{Code: `new Foo(async (resolve, reject) => {})`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Async named function
			{
				Code: `new Promise(async function foo(resolve, reject) {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "async",
						Line:      1,
						Column:    13,
					},
				},
			},

			// Async arrow function
			{
				Code: `new Promise(async (resolve, reject) => {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "async",
						Line:      1,
						Column:    13,
					},
				},
			},

			// Wrapped async arrow function
			{
				Code: `new Promise(((((async () => {})))))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "async",
						Line:      1,
						Column:    17,
					},
				},
			},
		},
	)
}
