package no_redeclare

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoRedeclareFileLanguageDefaults(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.allow-js.json",
		t,
		&NoRedeclareRule,
		[]rule_tester.ValidTestCase{
			{Code: `var require;`, FileName: "var.cjs"},
			{
				Code:     `var require;`,
				FileName: "legacy-var.cjs",
				TSConfig: "tsconfig.allow-js-legacy-module-detection.json",
			},
			{
				Code:     `var Object;`,
				FileName: "legacy-module.js",
				TSConfig: "tsconfig.allow-js-legacy-module-detection.json",
			},
			{Code: `let require;`, FileName: "let.cjs"},
			{Code: `var Object;`, FileName: "builtin.cjs"},
			{
				Code:     `var custom;`,
				FileName: "configured.cjs",
				Globals:  map[string]any{"custom": "readonly"},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     `var require; var require;`,
				FileName: "duplicate.cjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclared", Line: 1, Column: 18},
				},
			},
			{
				Code:     `/* global require */`,
				FileName: "directive.cjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclaredAsBuiltin", Line: 1, Column: 11},
				},
			},
		},
	)
}
