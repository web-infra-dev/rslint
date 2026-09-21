package url_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/prefer_global/url"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func urlError(messageID string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	message := "Unexpected use of 'require(\"url\").URL'. Use the global variable 'URL' instead."
	if messageID == "preferModule" {
		message = "Unexpected use of the global variable 'URL'. Use 'require(\"url\").URL' instead."
	}
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID, Message: message,
		Line: line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}

func runURLTests(t *testing.T, valid []rule_tester.ValidTestCase, invalid []rule_tester.InvalidTestCase) {
	t.Helper()
	globals := func(extra map[string]any) map[string]any {
		result := map[string]any{"URL": "readonly", "require": "readonly", "process": "readonly", "global": "readonly"}
		for name, access := range extra {
			result[name] = access
		}
		return result
	}
	for i := range valid {
		valid[i].Globals = globals(valid[i].Globals)
		if valid[i].LanguageOptions.SourceType == "" {
			valid[i].LanguageOptions.SourceType = "commonjs"
		}
	}
	for i := range invalid {
		invalid[i].Globals = globals(invalid[i].Globals)
		if invalid[i].LanguageOptions.SourceType == "" {
			invalid[i].LanguageOptions.SourceType = "commonjs"
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &url.PreferGlobalURLRule, valid, invalid)
}

// Every case from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/prefer-global/url.js
func TestPreferGlobalURLUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: "var b = new URL(s)"},
		{Code: "var b = new URL(s)", Options: []any{"always"}},
	}
	var invalid []rule_tester.InvalidTestCase
	for _, method := range []string{"require", "process.getBuiltinModule"} {
		for _, module := range []string{"url", "node:url"} {
			code := "var { URL } = " + method + "('" + module + "'); var b = new URL(s)"
			valid = append(valid, rule_tester.ValidTestCase{Code: code, Options: []any{"never"}})
			for _, options := range [][]any{nil, {"always"}} {
				invalid = append(invalid, rule_tester.InvalidTestCase{
					Code: code, Options: options,
					Errors: []rule_tester.InvalidTestCaseError{urlError("preferGlobal", 1, 7, 1, 10)},
				})
			}
		}
	}
	invalid = append(invalid, rule_tester.InvalidTestCase{
		Code: "var b = new URL(s)", Options: []any{"never"},
		Errors: []rule_tester.InvalidTestCaseError{urlError("preferModule", 1, 13, 1, 16)},
	})
	runURLTests(t, valid, invalid)
}

// Every JavaScript example from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-global/url.md
// Rule-enabling comments are omitted; the tester enables the rule.
func TestPreferGlobalURLDocumentation(t *testing.T) {
	runURLTests(t,
		[]rule_tester.ValidTestCase{
			{Code: "const u = new URL(s)"},
			{Code: "const { URL } = require(\"url\")\nconst u = new URL(s)", Options: []any{"never"}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   "console.log(URL === require(\"url\").URL) //→ true",
				Errors: []rule_tester.InvalidTestCaseError{urlError("preferGlobal", 1, 21, 1, 39)},
			},
			{
				Code:   "const { URL } = require(\"url\")\nconst u = new URL(s)",
				Errors: []rule_tester.InvalidTestCaseError{urlError("preferGlobal", 1, 9, 1, 12)},
			},
			{
				Code: "const u = new URL(s)", Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{urlError("preferModule", 1, 15, 1, 18)},
			},
		},
	)
}
