package return_await

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Expected diagnostics and outputs also verified with typescript-eslint v8.67.0.
func TestReturnAwaitWhitespaceAndComments(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ReturnAwaitRule, nil, []rule_tester.InvalidTestCase{
		{
			Code:   "async function f() { return await /*keep*/ 1; }",
			Output: []string{"async function f() { return /*keep*/ 1; }"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nonPromiseAwait", Message: "Returning an awaited value that is not a promise is not allowed.", Line: 1, Column: 29, EndLine: 1, EndColumn: 45}},
		},
		{
			Code:   "async function f() { return await\r\n 1; }",
			Output: []string{"async function f() { return 1; }"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nonPromiseAwait", Message: "Returning an awaited value that is not a promise is not allowed.", Line: 1, Column: 29, EndLine: 2, EndColumn: 3}},
		},
		{
			Code:   "async function f() { return await(1); }",
			Output: []string{"async function f() { return (1); }"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nonPromiseAwait", Message: "Returning an awaited value that is not a promise is not allowed.", Line: 1, Column: 29, EndLine: 1, EndColumn: 37}},
		},
		{
			Code:   "async function f() { return await /*outer*/ (await 1); }",
			Output: []string{"async function f() { return /*outer*/ (await 1); }", "async function f() { return /*outer*/ (1); }"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nonPromiseAwait", Message: "Returning an awaited value that is not a promise is not allowed.", Line: 1, Column: 29, EndLine: 1, EndColumn: 54}},
		},
	})
}
