package spec_only_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/spec_only"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// All cases from eslint-plugin-promise v7.3.0, plus its documentation examples.
// https://github.com/eslint-community/eslint-plugin-promise/blob/v7.3.0/__tests__/spec-only.js
// https://github.com/eslint-community/eslint-plugin-promise/blob/v7.3.0/docs/rules/spec-only.md
func TestSpecOnlyUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &spec_only.SpecOnlyRule,
		[]rule_tester.ValidTestCase{
			{Code: `Promise.resolve()`},
			{Code: `Promise.reject()`},
			{Code: `Promise.all()`},
			{Code: `Promise["all"]`},
			{Code: `Promise[method];`},
			{Code: `Promise.prototype;`},
			{Code: `Promise.prototype[method];`},
			{Code: `Promise.prototype["then"];`},
			{Code: `Promise.race()`},
			// cspell:disable-next-line
			{Code: `var ctch = Promise.prototype.catch`},
			{Code: `Promise.withResolvers()`},
			{Code: `new Promise(function (resolve, reject) {})`},
			{Code: `SomeClass.resolve()`},
			{Code: `doSomething(Promise.all)`},
			{Code: `Promise.permittedMethod()`,
				Options: []any{map[string]any{"allowedMethods": []any{"permittedMethod"}}}},
			{Code: `Promise.prototype.permittedInstanceMethod`,
				Options: []any{map[string]any{"allowedMethods": []any{"permittedInstanceMethod"}}}},
			// Pinned documentation: valid example.
			{Code: `const x = Promise.resolve('good')`},
		},
		[]rule_tester.InvalidTestCase{
			{Code: `Promise.done()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.done'", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				}},
			{Code: `Promise.something()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.something'", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				}},
			{Code: `new Promise.done()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.done'", Line: 1, Column: 5, EndLine: 1, EndColumn: 17},
				}},
			{Code: `
        function foo() {
          var a = getA()
          return Promise.done(a)
        }
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.done'", Line: 4, Column: 18, EndLine: 4, EndColumn: 30},
				}},
			{Code: `
        function foo() {
          getA(Promise.done)
        }
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.done'", Line: 3, Column: 16, EndLine: 3, EndColumn: 28},
				}},
			{Code: `var done = Promise.prototype.done`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.prototype'", Line: 1, Column: 12, EndLine: 1, EndColumn: 29},
				}},
			{Code: `Promise["done"];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.done'", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				}},
			// Pinned documentation: invalid example.
			{Code: `const x = Promise.done('bad')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.done'", Line: 1, Column: 11, EndLine: 1, EndColumn: 23},
				}},
		},
	)
}
