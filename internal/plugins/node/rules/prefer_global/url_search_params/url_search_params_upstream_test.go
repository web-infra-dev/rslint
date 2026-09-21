package url_search_params_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/prefer_global/url_search_params"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func urlSearchParamsAt(messageID string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	message := "Unexpected use of 'require(\"url\").URLSearchParams'. Use the global variable 'URLSearchParams' instead."
	if messageID == "preferModule" {
		message = "Unexpected use of the global variable 'URLSearchParams'. Use 'require(\"url\").URLSearchParams' instead."
	}
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID, Message: message,
		Line: line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}

func runURLSearchParamsTests(t *testing.T, valid []rule_tester.ValidTestCase, invalid []rule_tester.InvalidTestCase) {
	t.Helper()
	globals := func(overrides map[string]any) map[string]any {
		result := map[string]any{"URLSearchParams": "readonly", "require": "readonly", "process": "readonly", "global": "readonly"}
		for name, access := range overrides {
			result[name] = access
		}
		return result
	}
	for i := range valid {
		valid[i].Globals = globals(valid[i].Globals)
		if valid[i].FileName == "" {
			valid[i].FileName, valid[i].TSConfig = "input.js", "tsconfig.allowJs.json"
		}
	}
	for i := range invalid {
		invalid[i].Globals = globals(invalid[i].Globals)
		if invalid[i].FileName == "" {
			invalid[i].FileName, invalid[i].TSConfig = "input.js", "tsconfig.allowJs.json"
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &url_search_params.URLSearchParamsRule, valid, invalid)
}

// Every case from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/prefer-global/url-search-params.js
func TestURLSearchParamsUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: "var b = new URLSearchParams(s)"},
		{Code: "var b = new URLSearchParams(s)", Options: []any{"always"}},
	}
	var invalid []rule_tester.InvalidTestCase
	for _, method := range []string{"require", "process.getBuiltinModule"} {
		for _, module := range []string{"url", "node:url"} {
			code := "var { URLSearchParams } = " + method + "('" + module + "'); var b = new URLSearchParams(s)"
			valid = append(valid, rule_tester.ValidTestCase{Code: code, Options: []any{"never"}})
			for _, options := range [][]any{nil, {"always"}} {
				invalid = append(invalid, rule_tester.InvalidTestCase{
					Code: code, Options: options,
					Errors: []rule_tester.InvalidTestCaseError{urlSearchParamsAt("preferGlobal", 1, 7, 1, 22)},
				})
			}
		}
	}
	invalid = append(invalid, rule_tester.InvalidTestCase{
		Code: "var b = new URLSearchParams(s)", Options: []any{"never"},
		Errors: []rule_tester.InvalidTestCaseError{urlSearchParamsAt("preferModule", 1, 13, 1, 28)},
	})
	runURLSearchParamsTests(t, valid, invalid)
}

// Every JavaScript example from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-global/url-search-params.md
// Rule-enabling comments are omitted; the tester enables the rule.
func TestURLSearchParamsDocumentation(t *testing.T) {
	runURLSearchParamsTests(t,
		[]rule_tester.ValidTestCase{
			{Code: "const u = new URLSearchParams(s)"},
			{Code: "const { URLSearchParams } = require(\"url\")\nconst u = new URLSearchParams(s)", Options: []any{"never"}},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "console.log(URLSearchParams === require(\"url\").URLSearchParams) //→ true",
				Errors: []rule_tester.InvalidTestCaseError{urlSearchParamsAt("preferGlobal", 1, 33, 1, 63)}},
			{Code: "const { URLSearchParams } = require(\"url\")\nconst u = new URLSearchParams(s)",
				Errors: []rule_tester.InvalidTestCaseError{urlSearchParamsAt("preferGlobal", 1, 9, 1, 24)}},
			{Code: "const u = new URLSearchParams(s)", Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{urlSearchParamsAt("preferModule", 1, 15, 1, 30)}},
		},
	)
}
