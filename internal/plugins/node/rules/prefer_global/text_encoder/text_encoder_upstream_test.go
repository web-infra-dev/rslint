package text_encoder_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/prefer_global/text_encoder"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func textEncoderAt(messageID string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	message := "Unexpected use of 'require(\"util\").TextEncoder'. Use the global variable 'TextEncoder' instead."
	if messageID == "preferModule" {
		message = "Unexpected use of the global variable 'TextEncoder'. Use 'require(\"util\").TextEncoder' instead."
	}
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID, Message: message,
		Line: line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}

func runTextEncoderTests(t *testing.T, valid []rule_tester.ValidTestCase, invalid []rule_tester.InvalidTestCase) {
	t.Helper()
	globals := func(overrides map[string]any) map[string]any {
		result := map[string]any{"TextEncoder": "readonly", "require": "readonly", "process": "readonly", "global": "readonly"}
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
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &text_encoder.TextEncoderRule, valid, invalid)
}

// Every case from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/prefer-global/text-encoder.js
func TestTextEncoderUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: "var b = new TextEncoder(s)"},
		{Code: "var b = new TextEncoder(s)", Options: []any{"always"}},
	}
	var invalid []rule_tester.InvalidTestCase
	for _, method := range []string{"require", "process.getBuiltinModule"} {
		for _, module := range []string{"util", "node:util"} {
			code := "var { TextEncoder } = " + method + "('" + module + "'); var b = new TextEncoder(s)"
			valid = append(valid, rule_tester.ValidTestCase{Code: code, Options: []any{"never"}})
			for _, options := range [][]any{nil, {"always"}} {
				invalid = append(invalid, rule_tester.InvalidTestCase{
					Code: code, Options: options,
					Errors: []rule_tester.InvalidTestCaseError{textEncoderAt("preferGlobal", 1, 7, 1, 18)},
				})
			}
		}
	}
	invalid = append(invalid, rule_tester.InvalidTestCase{
		Code: "var b = new TextEncoder(s)", Options: []any{"never"},
		Errors: []rule_tester.InvalidTestCaseError{textEncoderAt("preferModule", 1, 13, 1, 24)},
	})
	runTextEncoderTests(t, valid, invalid)
}

// Every JavaScript example from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-global/text-encoder.md
// Rule-enabling comments are omitted; the tester enables the rule.
func TestTextEncoderDocumentation(t *testing.T) {
	runTextEncoderTests(t,
		[]rule_tester.ValidTestCase{
			{Code: "const u = new TextEncoder(s)"},
			{Code: "const { TextEncoder } = require(\"util\")\nconst u = new TextEncoder(s)", Options: []any{"never"}},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "console.log(TextEncoder === require(\"util\").TextEncoder) //→ true",
				Errors: []rule_tester.InvalidTestCaseError{textEncoderAt("preferGlobal", 1, 29, 1, 56)}},
			{Code: "const { TextEncoder } = require(\"util\")\nconst u = new TextEncoder(s)",
				Errors: []rule_tester.InvalidTestCaseError{textEncoderAt("preferGlobal", 1, 9, 1, 20)}},
			{Code: "const u = new TextEncoder(s)", Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{textEncoderAt("preferModule", 1, 15, 1, 26)}},
		},
	)
}
