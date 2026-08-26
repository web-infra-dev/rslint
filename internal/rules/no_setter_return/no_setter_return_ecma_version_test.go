package no_setter_return

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const reflectDescriptorSetter = `Reflect.defineProperty(foo, "bar", { set: function(value) { return value; } })`

func TestNoSetterReturnECMAVersion(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoSetterReturnRule,
		[]rule_tester.ValidTestCase{
			{Code: reflectDescriptorSetter, LanguageOptions: rule.LanguageOptions{ECMAVersion: 5}},
			{Code: reflectDescriptorSetter, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}, Globals: map[string]any{"Reflect": "off"}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:            reflectDescriptorSetter,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 5},
				Globals:         map[string]any{"Reflect": "readonly"},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "returnsValue"}},
			},
			{
				Code:            reflectDescriptorSetter,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "returnsValue"}},
			},
		},
	)
}
