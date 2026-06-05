package no_new_statics_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/no_new_statics"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoNewStaticsUpstream migrates every valid/invalid case from the
// eslint-plugin-promise upstream test suite for `no-new-statics`.
func TestNoNewStaticsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_new_statics.NoNewStaticsRule,
		[]rule_tester.ValidTestCase{
			// ---- ESLint upstream valid cases ----
			{Code: `Promise.resolve()`},
			{Code: `Promise.reject()`},
			{Code: `Promise.all()`},
			{Code: `Promise.race()`},
			{Code: `Promise.withResolvers()`},
			{Code: `new Promise(function (resolve, reject) {})`},
			{Code: `new SomeClass()`},
			{Code: `SomeClass.resolve()`},
			{Code: `new SomeClass.resolve()`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- ESLint upstream invalid cases ----
			{
				Code:   `new Promise.resolve()`,
				Output: []string{`Promise.resolve()`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "avoidNewStatic",
						Message:   "Avoid calling 'new' on 'Promise.resolve()'",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 22,
					},
				},
			},
			{
				Code:   `new Promise.reject()`,
				Output: []string{`Promise.reject()`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "avoidNewStatic",
						Message:   "Avoid calling 'new' on 'Promise.reject()'",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 21,
					},
				},
			},
			{
				Code:   `new Promise.all()`,
				Output: []string{`Promise.all()`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "avoidNewStatic",
						Message:   "Avoid calling 'new' on 'Promise.all()'",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 18,
					},
				},
			},
			{
				Code:   `new Promise.allSettled()`,
				Output: []string{`Promise.allSettled()`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "avoidNewStatic",
						Message:   "Avoid calling 'new' on 'Promise.allSettled()'",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 25,
					},
				},
			},
			{
				Code:   `new Promise.any()`,
				Output: []string{`Promise.any()`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "avoidNewStatic",
						Message:   "Avoid calling 'new' on 'Promise.any()'",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 18,
					},
				},
			},
			{
				Code:   `new Promise.race()`,
				Output: []string{`Promise.race()`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "avoidNewStatic",
						Message:   "Avoid calling 'new' on 'Promise.race()'",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 19,
					},
				},
			},
			{
				Code:   `new Promise.withResolvers()`,
				Output: []string{`Promise.withResolvers()`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "avoidNewStatic",
						Message:   "Avoid calling 'new' on 'Promise.withResolvers()'",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 28,
					},
				},
			},
			// multi-line / nested context
			{
				Code:   "function a() { return new Promise.resolve(a) }",
				Output: []string{"function a() { return Promise.resolve(a) }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "avoidNewStatic",
						Message:   "Avoid calling 'new' on 'Promise.resolve()'",
						Line:      1,
						Column:    23,
						EndLine:   1,
						EndColumn: 45,
					},
				},
			},
		},
	)
}
