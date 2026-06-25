// TestPreferCatchUpstream migrates the full valid/invalid suite from upstream
// __tests__/prefer-catch.js 1:1. Position assertions cover line/column for
// every invalid case. rslint-specific lock-in cases live in prefer_catch_extras_test.go.
package prefer_catch_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/prefer_catch"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const preferCatchMessage = "Prefer `catch` to `then(a, b)`/`then(null, b)`."

func TestPreferCatchUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_catch.PreferCatchRule,
		[]rule_tester.ValidTestCase{
			// ---- upstream valid cases ----
			{Code: `prom.then()`},
			{Code: `prom.then(fn)`},
			{Code: `prom.then(fn1).then(fn2)`},
			{Code: `prom.then(() => {})`},
			{Code: `prom.then(function () {})`},
			{Code: `prom.catch()`},
			{Code: `prom.catch(handleErr).then(handle)`},
			{Code: `prom.catch(handleErr)`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- upstream invalid cases ----
			{
				Code:   `hey.then(fn1, fn2)`,
				Output: []string{`hey.catch(fn2).then(fn1)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCatchToThen", Message: preferCatchMessage, Line: 1, Column: 5, EndLine: 1, EndColumn: 9}},
			},
			{
				Code:   `hey.then(fn1, (fn2))`,
				Output: []string{`hey.catch(fn2).then(fn1)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCatchToThen", Message: preferCatchMessage, Line: 1, Column: 5}},
			},
			{
				Code:   `hey.then(null, fn2)`,
				Output: []string{`hey.catch(fn2)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCatchToThen", Message: preferCatchMessage, Line: 1, Column: 5}},
			},
			{
				Code:   `hey.then(undefined, fn2)`,
				Output: []string{`hey.catch(fn2)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCatchToThen", Message: preferCatchMessage, Line: 1, Column: 5}},
			},
			{
				Code:   `function foo() { hey.then(x => {}, () => {}) }`,
				Output: []string{`function foo() { hey.catch(() => {}).then(x => {}) }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCatchToThen", Message: preferCatchMessage, Line: 1, Column: 22}},
			},
			{
				Code: `
function foo() {
  hey.then(function a() { }, function b() {}).then(fn1, fn2)
}
`,
				Output: []string{`
function foo() {
  hey.catch(function b() {}).then(function a() { }).catch(fn2).then(fn1)
}
`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferCatchToThen", Message: preferCatchMessage, Line: 3, Column: 47},
					{MessageId: "preferCatchToThen", Message: preferCatchMessage, Line: 3, Column: 7},
				},
			},
		},
	)
}
