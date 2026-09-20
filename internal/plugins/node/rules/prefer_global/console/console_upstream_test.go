package console_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/prefer_global/console"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func consoleError(messageID string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	message := "Unexpected use of 'require(\"console\")'. Use the global variable 'console' instead."
	if messageID == "preferModule" {
		message = "Unexpected use of the global variable 'console'. Use 'require(\"console\")' instead."
	}
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID, Message: message,
		Line: line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}

// Every case from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/prefer-global/console.js
func TestPreferGlobalConsoleUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: "console.log(10)"},
		{Code: "console.log(10)", Options: []any{"always"}},
	}
	var invalid []rule_tester.InvalidTestCase
	for _, method := range []string{"require", "process.getBuiltinModule"} {
		for _, module := range []string{"console", "node:console"} {
			load := method + "('" + module + "')"
			code := "var console = " + load + "; console.log(10)"
			valid = append(valid, rule_tester.ValidTestCase{Code: code, Options: []any{"never"}})
			for _, options := range [][]any{nil, {"always"}} {
				invalid = append(invalid, rule_tester.InvalidTestCase{
					Code: code, Options: options,
					Errors: []rule_tester.InvalidTestCaseError{consoleError("preferGlobal", 1, 15, 1, 15+len(load))},
				})
			}
		}
	}
	invalid = append(invalid, rule_tester.InvalidTestCase{
		Code: "console.log(10)", Options: []any{"never"},
		Errors: []rule_tester.InvalidTestCaseError{consoleError("preferModule", 1, 1, 1, 8)},
	})
	runConsoleTests(t, valid, invalid)
}

// Every JavaScript example from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-global/console.md
// Rule-enabling comments are omitted; the tester supplies those options.
func TestPreferGlobalConsoleDocumentation(t *testing.T) {
	runConsoleTests(t, []rule_tester.ValidTestCase{
		{Code: `console.log("hello")`},
		{Code: "const console = require(\"console\")\nconsole.log(\"hello\")", Options: []any{"never"}},
	}, []rule_tester.InvalidTestCase{
		{Code: `console.log(console === require("console")) //→ true`,
			Errors: []rule_tester.InvalidTestCaseError{consoleError("preferGlobal", 1, 25, 1, 43)},
		},
		{Code: "const console = require(\"console\")\nconsole.log(\"hello\")",
			Errors: []rule_tester.InvalidTestCaseError{consoleError("preferGlobal", 1, 17, 1, 35)},
		},
		{Code: `console.log("hello")`, Options: []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{consoleError("preferModule", 1, 1, 1, 8)},
		},
	})
}

func runConsoleTests(t *testing.T, valid []rule_tester.ValidTestCase, invalid []rule_tester.InvalidTestCase) {
	t.Helper()
	globals := func(overrides map[string]any) map[string]any {
		result := map[string]any{"console": "readonly", "require": "readonly", "process": "readonly", "global": "readonly"}
		for name, access := range overrides {
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
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &console.PreferGlobalConsoleRule, valid, invalid)
}
