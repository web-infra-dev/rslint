package no_process_env_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/no_process_env"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func noProcessEnvAt(line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "unexpectedProcessEnv",
		Message:   "Unexpected use of process.env.",
		Line:      line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}

// Every case from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-process-env.js
func TestNoProcessEnvUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_process_env.NoProcessEnvRule,
		[]rule_tester.ValidTestCase{
			{
				Code: `Process.env`,
			},
			{
				Code: `process[env]`,
			},
			{
				Code: `process.nextTick`,
			},
			{
				Code: `process.execArgv`,
			},
			// allowedVariables
			{
				Code:    `process.env.NODE_ENV`,
				Options: []any{map[string]any{"allowedVariables": []any{"NODE_ENV"}}},
			},
			{
				Code:    `process.env['NODE_ENV']`,
				Options: []any{map[string]any{"allowedVariables": []any{"NODE_ENV"}}},
			},
			{
				Code:    `process['env'].NODE_ENV`,
				Options: []any{map[string]any{"allowedVariables": []any{"NODE_ENV"}}},
			},
			{
				Code:    `process['env']['NODE_ENV']`,
				Options: []any{map[string]any{"allowedVariables": []any{"NODE_ENV"}}},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `process.env`,
				Errors: []rule_tester.InvalidTestCaseError{noProcessEnvAt(1, 1, 1, 12)},
			},
			{
				Code:   `process['env']`,
				Errors: []rule_tester.InvalidTestCaseError{noProcessEnvAt(1, 1, 1, 15)},
			},
			{
				Code:   `process.env.ENV`,
				Errors: []rule_tester.InvalidTestCaseError{noProcessEnvAt(1, 1, 1, 12)},
			},
			{
				Code:   `f(process.env)`,
				Errors: []rule_tester.InvalidTestCaseError{noProcessEnvAt(1, 3, 1, 14)},
			},
			// allowedVariables
			{
				Code:    `process.env['OTHER_VARIABLE']`,
				Options: []any{map[string]any{"allowedVariables": []any{"NODE_ENV"}}},
				Errors:  []rule_tester.InvalidTestCaseError{noProcessEnvAt(1, 1, 1, 12)},
			},
			{
				Code:    `process.env.OTHER_VARIABLE`,
				Options: []any{map[string]any{"allowedVariables": []any{"NODE_ENV"}}},
				Errors:  []rule_tester.InvalidTestCaseError{noProcessEnvAt(1, 1, 1, 12)},
			},
			{
				Code:    `process['env']['OTHER_VARIABLE']`,
				Options: []any{map[string]any{"allowedVariables": []any{"NODE_ENV"}}},
				Errors:  []rule_tester.InvalidTestCaseError{noProcessEnvAt(1, 1, 1, 15)},
			},
			{
				Code:    `process['env'].OTHER_VARIABLE`,
				Options: []any{map[string]any{"allowedVariables": []any{"NODE_ENV"}}},
				Errors:  []rule_tester.InvalidTestCaseError{noProcessEnvAt(1, 1, 1, 15)},
			},
			{
				Code:    `process.env[NODE_ENV]`,
				Options: []any{map[string]any{"allowedVariables": []any{"NODE_ENV"}}},
				Errors:  []rule_tester.InvalidTestCaseError{noProcessEnvAt(1, 1, 1, 12)},
			},
			{
				Code:    `process['env'][NODE_ENV]`,
				Options: []any{map[string]any{"allowedVariables": []any{"NODE_ENV"}}},
				Errors:  []rule_tester.InvalidTestCaseError{noProcessEnvAt(1, 1, 1, 15)},
			},
		},
	)
}

// Every JavaScript example from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-process-env.md
// Rule-enabling comments are omitted; the tester enables the rule.
func TestNoProcessEnvDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_process_env.NoProcessEnvRule,
		[]rule_tester.ValidTestCase{
			// Use a configuration module.
			{
				Code: `var config = require("./config");

if(config.env === "development") {
    //...
}`,
			},
		},
		[]rule_tester.InvalidTestCase{
			// Read an environment variable directly.
			{
				Code: `if(process.env.NODE_ENV === "development") {
    //...
}`,
				Errors: []rule_tester.InvalidTestCaseError{noProcessEnvAt(1, 4, 1, 15)},
			},
		},
	)
}
