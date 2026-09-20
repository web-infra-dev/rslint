package buffer_test

import (
	"fmt"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/prefer_global/buffer"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func preferBufferAt(messageID string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	message := "Unexpected use of 'require(\"buffer\").Buffer'. Use the global variable 'Buffer' instead."
	if messageID == "preferModule" {
		message = "Unexpected use of the global variable 'Buffer'. Use 'require(\"buffer\").Buffer' instead."
	}
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID, Message: message,
		Line: line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}

// Match the Node globals and CommonJS source type supplied by the upstream tester.
func runBufferTests(t *testing.T, valid []rule_tester.ValidTestCase, invalid []rule_tester.InvalidTestCase) {
	t.Helper()
	globals := func(overrides map[string]any) map[string]any {
		result := map[string]any{"Buffer": "readonly", "require": "readonly", "process": "readonly", "global": "readonly"}
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
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &buffer.PreferGlobalBufferRule, valid, invalid)
}

// Every case from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/prefer-global/buffer.js
func TestPreferGlobalBufferUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: "var b = Buffer.alloc(10)"},
		{Code: "var b = Buffer.alloc(10)", Options: []any{"always"}},
	}
	var invalid []rule_tester.InvalidTestCase
	for _, method := range []string{"require", "process.getBuiltinModule"} {
		for _, module := range []string{"buffer", "node:buffer"} {
			code := fmt.Sprintf("var { Buffer } = %s('%s'); var b = Buffer.alloc(10)", method, module)
			valid = append(valid, rule_tester.ValidTestCase{Code: code, Options: []any{"never"}})
			for _, options := range [][]any{nil, {"always"}} {
				invalid = append(invalid, rule_tester.InvalidTestCase{
					Code: code, Options: options,
					Errors: []rule_tester.InvalidTestCaseError{preferBufferAt("preferGlobal", 1, 7, 1, 13)},
				})
			}
		}
	}
	invalid = append(invalid, rule_tester.InvalidTestCase{
		Code: "var b = Buffer.alloc(10)", Options: []any{"never"},
		Errors: []rule_tester.InvalidTestCaseError{preferBufferAt("preferModule", 1, 9, 1, 15)},
	})
	runBufferTests(t, valid, invalid)
}

// Every JavaScript example from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-global/buffer.md
// Rule-enabling comments are omitted because the tester supplies the configuration.
func TestPreferGlobalBufferDocumentation(t *testing.T) {
	runBufferTests(t, []rule_tester.ValidTestCase{
		{Code: "const b = Buffer.alloc(16)"},
		{Code: "const { Buffer } = require(\"buffer\")\nconst b = Buffer.alloc(16)", Options: []any{"never"}},
	}, []rule_tester.InvalidTestCase{
		{Code: "console.log(Buffer === require(\"buffer\").Buffer) //→ true",
			Errors: []rule_tester.InvalidTestCaseError{preferBufferAt("preferGlobal", 1, 24, 1, 48)}},
		{Code: "const { Buffer } = require(\"buffer\")\nconst b = Buffer.alloc(16)",
			Errors: []rule_tester.InvalidTestCaseError{preferBufferAt("preferGlobal", 1, 9, 1, 15)}},
		{Code: "const b = Buffer.alloc(16)", Options: []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{preferBufferAt("preferModule", 1, 11, 1, 17)}},
	})
}
