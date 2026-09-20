package process_test

import (
	"fmt"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/prefer_global/process"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func processError(messageID string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	message := "Unexpected use of 'require(\"process\")'. Use the global variable 'process' instead."
	if messageID == "preferModule" {
		message = "Unexpected use of the global variable 'process'. Use 'require(\"process\")' instead."
	}
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID, Message: message,
		Line: line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}

// Every case from eslint-plugin-n v18.3.0, including both module-loading methods:
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/prefer-global/process.js
func TestProcessUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: `process.exit(0)`},
		{Code: `process.exit(0)`, Options: []any{"always"}},
		{Code: `var process = require('process'); process.exit(0)`, Options: []any{"never"}},
		{Code: `var process = require('node:process'); process.exit(0)`, Options: []any{"never"}},
		{Code: `process.getBuiltinModule('buffer')`, Options: []any{"always"}},
	}
	var invalid []rule_tester.InvalidTestCase
	for _, method := range []string{"require", "process.getBuiltinModule"} {
		for _, options := range [][]any{nil, {"always"}} {
			for _, module := range []string{"process", "node:process"} {
				call := fmt.Sprintf("%s('%s')", method, module)
				invalid = append(invalid, rule_tester.InvalidTestCase{
					Code: "var process_ = " + call + "; process_.exit(0)", Options: options,
					Errors: []rule_tester.InvalidTestCaseError{processError("preferGlobal", 1, 16, 1, 16+len(call))},
				})
			}
		}
	}
	invalid = append(invalid, rule_tester.InvalidTestCase{
		Code: `process.exit(0)`, Options: []any{"never"},
		Errors: []rule_tester.InvalidTestCaseError{processError("preferModule", 1, 1, 1, 8)},
	})
	runProcessTests(t, valid, invalid)
}

// All JavaScript examples from the documentation at the same tag.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-global/process.md
func TestProcessDocumentation(t *testing.T) {
	runProcessTests(t,
		[]rule_tester.ValidTestCase{
			{Code: `process.exit(0)`},
			{Code: "const process = require(\"process\")\nprocess.exit(0)", Options: []any{"never"}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `process.log(process === require("process"))`,
				Errors: []rule_tester.InvalidTestCaseError{processError("preferGlobal", 1, 25, 1, 43)},
			},
			{
				Code:   "const process = require(\"process\")\nprocess.exit(0)",
				Errors: []rule_tester.InvalidTestCaseError{processError("preferGlobal", 1, 17, 1, 35)},
			},
			{
				Code: `process.exit(0)`, Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{processError("preferModule", 1, 1, 1, 8)},
			},
		},
	)
}

func runProcessTests(t *testing.T, valid []rule_tester.ValidTestCase, invalid []rule_tester.InvalidTestCase) {
	t.Helper()
	configure := func(globals *map[string]any, language *rule.LanguageOptions) {
		if *globals == nil {
			*globals = map[string]any{}
		}
		for _, name := range []string{"require", "process", "global"} {
			if _, exists := (*globals)[name]; !exists {
				(*globals)[name] = "readonly"
			}
		}
		if language.SourceType == "" {
			language.SourceType = "commonjs"
		}
	}
	for i := range valid {
		configure(&valid[i].Globals, &valid[i].LanguageOptions)
	}
	for i := range invalid {
		configure(&invalid[i].Globals, &invalid[i].LanguageOptions)
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &process.ProcessRule, valid, invalid)
}
