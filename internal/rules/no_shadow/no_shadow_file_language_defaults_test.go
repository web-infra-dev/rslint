package no_shadow

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoShadowFileLanguageDefaults(t *testing.T) {
	builtinGlobals := map[string]any{"builtinGlobals": true}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.allow-js.json",
		t,
		&NoShadowRule,
		[]rule_tester.ValidTestCase{
			{
				Code:     `function f(require) { return require; }`,
				FileName: "plain-script.js",
				Options:  builtinGlobals,
			},
			{
				Code:     `var require;`,
				FileName: "require-off.cjs",
				Options:  builtinGlobals,
				Globals:  map[string]any{"require": "off"},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     `function f(require) { return require; }`,
				FileName: "function.cjs",
				Options:  builtinGlobals,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadowGlobal", Line: 1, Column: 12},
				},
			},
			{
				Code:     `var exports;`,
				FileName: "exports.cjs",
				Options:  builtinGlobals,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noShadowGlobal", Line: 1, Column: 5},
				},
			},
		},
	)
}
