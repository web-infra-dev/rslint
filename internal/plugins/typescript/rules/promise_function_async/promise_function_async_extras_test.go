package promise_function_async

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Expected diagnostics and outputs also verified with typescript-eslint v8.29.1.
func TestPromiseFunctionAsyncExportAndTokenBoundaries(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PromiseFunctionAsyncRule, nil, []rule_tester.InvalidTestCase{
		{
			Code:   "export function f() { return Promise.resolve(1); }",
			Output: []string{"export async function f() { return Promise.resolve(1); }"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingAsync", Message: "Functions that return promises must be async.", Line: 1, Column: 8, EndLine: 1, EndColumn: 18}},
		},
		{
			Code:   "export default function () { return Promise.resolve(1); }",
			Output: []string{"export default async function () { return Promise.resolve(1); }"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingAsync", Message: "Functions that return promises must be async.", Line: 1, Column: 16, EndLine: 1, EndColumn: 25}},
		},
		{
			Code:   "declare const key: unique symbol; class C { static/*x*/[key]() { return Promise.resolve(1); } }",
			Output: []string{"declare const key: unique symbol; class C { static/*x*/ async [key]() { return Promise.resolve(1); } }"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingAsync", Message: "Functions that return promises must be async.", Line: 1, Column: 45, EndLine: 1, EndColumn: 61}},
		},
		{
			Code:   "class C { static/* x */m() { return Promise.resolve(1); } }",
			Output: []string{"class C { static/* x */ async m() { return Promise.resolve(1); } }"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingAsync", Message: "Functions that return promises must be async.", Line: 1, Column: 11, EndLine: 1, EndColumn: 25}},
		},
		{
			Code:   "const o = {m() { return Promise.resolve(1); }};",
			Output: []string{"const o = { async m() { return Promise.resolve(1); }};"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingAsync", Message: "Functions that return promises must be async.", Line: 1, Column: 12, EndLine: 1, EndColumn: 13}},
		},
		{
			Code:   "class C { static m() { return Promise.resolve(1); } }",
			Output: []string{"class C { static async m() { return Promise.resolve(1); } }"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingAsync", Message: "Functions that return promises must be async.", Line: 1, Column: 11, EndLine: 1, EndColumn: 19}},
		},
	})
}
