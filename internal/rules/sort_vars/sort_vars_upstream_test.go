// TestSortVarsUpstream migrates the full valid/invalid suite from upstream
// eslint/tests/lib/rules/sort-vars.js 1:1. Position assertions cover line and
// column for every invalid case. rslint-specific lock-in cases live in
// sort_vars_extras_test.go.
package sort_vars

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const expectedMessage = "Variables within the same declaration block should be sorted alphabetically."

func sortVarsError(line, column int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "sortVars",
		Message:   expectedMessage,
		Line:      line,
		Column:    column,
	}
}

func TestSortVarsUpstream(t *testing.T) {
	ignoreCase := map[string]any{"ignoreCase": true}
	es2015 := rule.LanguageOptions{ECMAVersion: 2015}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&SortVarsRule,
		[]rule_tester.ValidTestCase{
			{Code: "var a=10, b=4, c='abc'"},
			{Code: "var a, b, c, d"},
			{Code: "var b; var a; var d;"},
			{Code: "var _a, a"},
			{Code: "var A, a"},
			{Code: "var A, b"},
			{Code: "var a, A;", Options: ignoreCase},
			{Code: "var A, a;", Options: ignoreCase},
			{Code: "var a, B, c;", Options: ignoreCase},
			{Code: "var A, b, C;", Options: ignoreCase},
			{Code: "var {a, b, c} = x;", Options: ignoreCase, LanguageOptions: es2015},
			{Code: "var {A, b, C} = x;", Options: ignoreCase, LanguageOptions: es2015},
			{Code: "var test = [1,2,3];", LanguageOptions: es2015},
			{Code: "var {a,b} = [1,2];", LanguageOptions: es2015},
			{Code: "var [a, B, c] = [1, 2, 3];", Options: ignoreCase, LanguageOptions: es2015},
			{Code: "var [A, B, c] = [1, 2, 3];", Options: ignoreCase, LanguageOptions: es2015},
			{Code: "var [A, b, C] = [1, 2, 3];", Options: ignoreCase, LanguageOptions: es2015},
			{Code: "let {a, b, c} = x;", LanguageOptions: es2015},
			{Code: "let [a, b, c] = [1, 2, 3];", LanguageOptions: es2015},
			{Code: `const {a, b, c} = {a: 1, b: true, c: "Moo"};`, Options: ignoreCase, LanguageOptions: es2015},
			{Code: `const [a, b, c] = [1, true, "Moo"];`, Options: ignoreCase, LanguageOptions: es2015},
			{Code: `const [c, a, b] = [1, true, "Moo"];`, Options: ignoreCase, LanguageOptions: es2015},
			{Code: "var {a, x: {b, c}} = {};", LanguageOptions: es2015},
			{Code: "var {c, x: {a, c}} = {};", LanguageOptions: es2015},
			{Code: "var {a, x: [b, c]} = {};", LanguageOptions: es2015},
			{Code: "var [a, {b, c}] = {};", LanguageOptions: es2015},
			{Code: "var [a, {x: {b, c}}] = {};", LanguageOptions: es2015},
			{Code: "var a = 42, {b, c } = {};", LanguageOptions: es2015},
			{Code: "var b = 42, {a, c } = {};", LanguageOptions: es2015},
			{Code: "var [b, {x: {a, c}}] = {};", LanguageOptions: es2015},
			{Code: "var [b, d, a, c] = {};", LanguageOptions: es2015},
			{Code: "var e, [a, c, d] = {};", LanguageOptions: es2015},
			{Code: "var a, [E, c, D] = [];", Options: ignoreCase, LanguageOptions: es2015},
			{Code: "var a, f, [e, c, d] = [1,2,3];", LanguageOptions: es2015},
			{
				Code:            "export default class {\n    render () {\n        let {\n            b\n        } = this,\n            a,\n            c;\n    }\n}",
				LanguageOptions: es2015,
			},
			{Code: "var {} = 1, a", Options: ignoreCase, LanguageOptions: es2015},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "var b, a", Output: []string{"var a, b"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortVars", Message: expectedMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 9}}},
			{Code: "var b , a", Output: []string{"var a , b"}, Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 9)}},
			{Code: "var b,\n    a;", Output: []string{"var a,\n    b;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortVars", Message: expectedMessage, Line: 2, Column: 5, EndLine: 2, EndColumn: 6}}},
			{Code: "var b=10, a=20;", Output: []string{"var a=20, b=10;"}, Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 11)}},
			{Code: "var b=10, a=20, c=30;", Output: []string{"var a=20, b=10, c=30;"}, Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 11)}},
			{Code: "var all=10, a = 1", Output: []string{"var a = 1, all=10"}, Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 13)}},
			{Code: "var b, c, a, d", Output: []string{"var a, b, c, d"}, Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 11)}},
			{Code: "var c, d, a, b", Output: []string{"var a, b, c, d"}, Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 11), sortVarsError(1, 14)}},
			{Code: "var a, A;", Output: []string{"var A, a;"}, Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 8)}},
			{Code: "var a, B;", Output: []string{"var B, a;"}, Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 8)}},
			{Code: "var a, B, c;", Output: []string{"var B, a, c;"}, Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 8)}},
			{Code: "var B, a;", Output: []string{"var a, B;"}, Options: ignoreCase, Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 8)}},
			{Code: "var B, A, c;", Output: []string{"var A, B, c;"}, Options: ignoreCase, Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 8)}},
			{Code: "var d, a, [b, c] = {};", Output: []string{"var a, d, [b, c] = {};"}, Options: ignoreCase, LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 8)}},
			{Code: "var d, a, [b, {x: {c, e}}] = {};", Output: []string{"var a, d, [b, {x: {c, e}}] = {};"}, Options: ignoreCase, LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 8)}},
			{Code: "var {} = 1, b, a", Output: []string{"var {} = 1, a, b"}, Options: ignoreCase, LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 16)}},
			{Code: "var b=10, a=f();", Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 11)}},
			{Code: "var b=10, a=b;", Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 11)}},
			{Code: "var b = 0, a = `${b}`;", LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 12)}},
			{Code: "var b = 0, a = `${f()}`", LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 12)}},
			{Code: "var b = 0, c = b, a;", Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 19)}},
			{Code: "var b = 0, c = 0, a = b + c;", Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 19)}},
			{Code: "var b = f(), c, d, a;", Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 20)}},
			{Code: "var b = `${f()}`, c, d, a;", LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 25)}},
			{Code: "var c, a = b = 0", Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 8)}},
		},
	)
}
