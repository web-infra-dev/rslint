package no_global_assign

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoGlobalAssignFileLanguageDefaults(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.allow-js.json",
		t,
		&NoGlobalAssignRule,
		[]rule_tester.ValidTestCase{
			{Code: `exports = {};`, FileName: "exports.cjs"},
			{
				Code:     `require = replacement;`,
				FileName: "writable-require.cjs",
				Globals:  map[string]any{"require": "writable"},
			},
			{
				Code:     `/* global module:writable */ module = replacement;`,
				FileName: "writable-module.cjs",
			},
			{
				Code:     `arguments = [];`,
				FileName: "wrapper-arguments.cjs",
				Globals:  map[string]any{"arguments": "readonly"},
			},
			{
				Code:     `function f() { arguments = []; }`,
				FileName: "function-arguments.cjs",
				Globals:  map[string]any{"arguments": "readonly"},
			},
			{
				Code:     `let require; require = replacement;`,
				FileName: "local-require.cjs",
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     `global = replacement;`,
				FileName: "global.cjs",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1}},
			},
			{
				Code:     `module = replacement;`,
				FileName: "module.cjs",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1}},
			},
			{
				Code:     `require = replacement;`,
				FileName: "require.cjs",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1}},
			},
			{
				Code:     `exports = replacement;`,
				FileName: "readonly-exports.cjs",
				Globals:  map[string]any{"exports": "readonly"},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1}},
			},
		},
	)
}
