package no_native_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/no_native"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const msgPromiseNotDefined = `"Promise" is not defined.`

func TestNoNativeUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_native.NoNativeRule,
		[]rule_tester.ValidTestCase{
			{Code: `var Promise = null; function x() { return Promise.resolve("hi"); }`},
			{Code: `var Promise = window.Promise || require("bluebird"); var x = Promise.reject();`},
			{Code: `import Promise from "bluebird"; var x = Promise.reject();`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `new Promise(function(reject, resolve) { })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: msgPromiseNotDefined, Line: 1, Column: 5, EndLine: 1, EndColumn: 12}},
			},
			{
				Code:   `Promise.resolve()`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: msgPromiseNotDefined, Line: 1, Column: 1, EndLine: 1, EndColumn: 8}},
			},
		},
	)
}
