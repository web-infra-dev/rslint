package no_new_func

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Mirrors eslint/eslint v10.9.1 tests/lib/rules/no-new-func.js.
func TestNoNewFuncRuleUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoNewFuncRule,
		[]rule_tester.ValidTestCase{
			{Code: `var a = new _function("b", "c", "return b+c");`},
			{Code: `var a = _function("b", "c", "return b+c");`},
			{Code: `class Function {}; new Function()`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `const fn = () => { class Function {}; new Function() }`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},
			{Code: `function Function() {}; Function()`},
			{Code: `var fn = function () { function Function() {}; Function() }`},
			{Code: `var x = function Function() { Function(); }`},
			{Code: `call(Function)`},
			{Code: `new Class(Function)`},
			{Code: `foo[Function]()`},
			{Code: `foo(Function.bind)`},
			{Code: `Function.toString()`},
			{Code: `Function[call]()`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `var a = new Function("b", "c", "return b+c");`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor"}},
			},
			{
				Code:   `var a = Function("b", "c", "return b+c");`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor"}},
			},
			{
				Code:   `var a = Function.call(null, "b", "c", "return b+c");`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor"}},
			},
			{
				Code:   `var a = Function.apply(null, ["b", "c", "return b+c"]);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor"}},
			},
			{
				Code:   `var a = Function.bind(null, "b", "c", "return b+c")();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor"}},
			},
			{
				Code:   `var a = Function.bind(null, "b", "c", "return b+c");`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor"}},
			},
			{
				Code:   `var a = Function["call"](null, "b", "c", "return b+c");`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor"}},
			},
			{
				Code:            `var a = (Function?.call)(null, "b", "c", "return b+c");`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2021},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor"}},
			},
			{
				Code:            `const fn = () => { class Function {} }; new Function('', '')`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor"}},
			},
			{
				Code:   `var fn = function () { function Function() {} }; Function('', '')`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor"}},
			},
		},
	)
}
