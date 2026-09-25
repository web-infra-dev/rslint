// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/prefer-structured-clone.js
package prefer_structured_clone_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_structured_clone"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func valid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}}
}

func invalidJSON(code, output string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code: code, FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
		Output: []string{},
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "prefer-structured-clone/error",
			Message:   "Prefer `structuredClone(…)` over `JSON.parse(JSON.stringify(…))` to create a deep clone.",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
				MessageId: "prefer-structured-clone/suggestion",
				Output:    output,
			}},
		}},
	}
}

func invalidFunction(code, path, output string, functions ...string) rule_tester.InvalidTestCase {
	testCase := rule_tester.InvalidTestCase{
		Code: code, FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
		Output: []string{},
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "prefer-structured-clone/error",
			Message:   "Prefer `structuredClone(…)` over `" + path + "(…)` to create a deep clone.",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
				MessageId: "prefer-structured-clone/suggestion",
				Output:    output,
			}},
		}},
	}
	if len(functions) > 0 {
		values := make([]any, len(functions))
		for i, value := range functions {
			values[i] = value
		}
		testCase.Options = []any{map[string]any{"functions": values}}
	}
	return testCase
}

func TestPreferStructuredCloneUpstream(t *testing.T) {
	validCases := []rule_tester.ValidTestCase{
		valid("structuredClone(foo)"),
		valid("JSON.parse(new JSON.stringify(foo))"),
		valid("new JSON.parse(JSON.stringify(foo))"),
		valid("JSON.parse(JSON.stringify())"),
		valid("JSON.parse(JSON.stringify(...foo))"),
		valid("JSON.parse(JSON.stringify(foo, extraArgument))"),
		valid("JSON.parse(...JSON.stringify(foo))"),
		valid("JSON.parse(JSON.stringify(foo), extraArgument)"),
		valid("JSON.parse(JSON.stringify?.(foo))"),
		valid("JSON.parse(JSON?.stringify(foo))"),
		valid("JSON.parse?.(JSON.stringify(foo))"),
		valid("JSON?.parse(JSON.stringify(foo))"),
		valid("JSON.parse(JSON.not_stringify(foo))"),
		valid("JSON.parse(not_JSON.stringify(foo))"),
		valid("JSON.not_parse(JSON.stringify(foo))"),
		valid("not_JSON.parse(JSON.stringify(foo))"),
		valid("JSON.stringify(JSON.parse(foo))"),
		valid("JSON.parse(JSON.stringify(foo, undefined, 2))"),
		valid("new _.cloneDeep(foo)"),
		valid("notMatchedFunction(foo)"),
		valid("_.cloneDeep()"),
		valid("_.cloneDeep(...foo)"),
		valid("_.cloneDeep(foo, extraArgument)"),
		valid("_.cloneDeep?.(foo)"),
		valid("_?.cloneDeep(foo)"),
	}

	invalidCases := []rule_tester.InvalidTestCase{
		invalidJSON("JSON.parse(JSON.stringify(foo))", "structuredClone(foo)"),
		invalidJSON("JSON.parse(JSON.stringify(foo),)", "structuredClone(foo,)"),
		invalidJSON("JSON.parse(JSON.stringify(foo,))", "structuredClone(foo)"),
		invalidJSON("JSON.parse(JSON.stringify(foo,),)", "structuredClone(foo,)"),
		invalidJSON("JSON.parse( ((JSON.stringify)) (foo))", "structuredClone(  foo)"),
		invalidJSON("(( JSON.parse)) (JSON.stringify(foo))", "(( structuredClone)) (foo)"),
		invalidJSON("JSON.parse(JSON.stringify( ((foo)) ))", "structuredClone( ((foo)) )"),
		invalidJSON("function foo() {\n\treturn JSON\n\t\t.parse(\n\t\t\tJSON.\n\t\t\t\tstringify(\n\t\t\t\t\tbar,\n\t\t\t\t),\n\t\t);\n}", "function foo() {\n\treturn structuredClone(\n\t\t\t\n\t\t\t\t\tbar\n\t\t\t\t,\n\t\t);\n}"),
		invalidFunction("_.cloneDeep(foo)", "_.cloneDeep", "structuredClone(foo)"),
		invalidFunction("lodash.cloneDeep(foo)", "lodash.cloneDeep", "structuredClone(foo)"),
		invalidFunction("lodash.cloneDeep(foo,)", "lodash.cloneDeep", "structuredClone(foo,)"),
		invalidFunction("myCustomDeepCloneFunction(foo,)", "myCustomDeepCloneFunction", "structuredClone(foo,)", "myCustomDeepCloneFunction"),
		invalidFunction("my.cloneDeep(foo,)", "my.cloneDeep", "structuredClone(foo,)", "my.cloneDeep"),
		invalidFunction("class A {\n\tconstructor() {\n\t\tthis.a = new.target.cloneDeep(foo);\n\t\tthis.b = import.meta.cloneDeep(foo);\n\t}\n}", "new.target.cloneDeep", "class A {\n\tconstructor() {\n\t\tthis.a = structuredClone(foo);\n\t\tthis.b = import.meta.cloneDeep(foo);\n\t}\n}", "new.target.cloneDeep"),
		invalidFunction("class A {\n\tconstructor() {\n\t\tthis.a = super.cloneDeep(foo);\n\t}\n}", "super.cloneDeep", "class A {\n\tconstructor() {\n\t\tthis.a = structuredClone(foo);\n\t}\n}", "super.cloneDeep"),
	}

	if len(validCases) != 25 || len(invalidCases) != 15 {
		t.Fatalf("upstream coverage accounting changed: valid=%d invalid=%d", len(validCases), len(invalidCases))
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_structured_clone.PreferStructuredCloneRule, validCases, invalidCases)
}
