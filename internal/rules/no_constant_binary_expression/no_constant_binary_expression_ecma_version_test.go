package no_constant_binary_expression

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoConstantBinaryExpressionECMAVersion(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoConstantBinaryExpressionRule,
		[]rule_tester.ValidTestCase{
			{Code: `new Promise(function() {}) === value`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 5}},
			{Code: `new Promise(function() {}) === value`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}, Globals: map[string]any{"Promise": "off"}},
			{Code: `new Image() === value`, Globals: map[string]any{"Image": "readonly"}},
			{Code: `function f(undefined) { return undefined === "x"; }`},
			{Code: `undefined ?? value`, Globals: map[string]any{"undefined": "off"}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:            `new Promise(function() {}) === value`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 5},
				Globals:         map[string]any{"Promise": "readonly"},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "alwaysNew"}},
			},
			{
				Code:            `new Promise(function() {}) === value`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "alwaysNew"}},
			},
			{
				Code:   `undefined === "x"`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "constantBinaryOperand"}},
			},
		},
	)
}
