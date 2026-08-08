package prefer_number_properties_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_number_properties"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferNumberPropertiesECMAVersion(t *testing.T) {
	windowCase := fixed(`window.NaN`, `Number.NaN`, "NaN", "NaN", 1, 1, 1, 11)
	windowCase.Globals = map[string]any{"window": "readonly"}
	globalThisCase := fixed(`globalThis.NaN`, `Number.NaN`, "NaN", "NaN", 1, 1, 1, 15)
	globalThisCase.Globals = map[string]any{"NaN": "off"}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_number_properties.PreferNumberPropertiesRule,
		[]rule_tester.ValidTestCase{
			{Code: `globalThis.NaN`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2019}},
			{Code: `window.NaN`},
			{Code: `NaN`, Globals: map[string]any{"NaN": "off"}},
			{Code: `globalThis.NaN`, Globals: map[string]any{"globalThis": "off"}},
		},
		[]rule_tester.InvalidTestCase{windowCase, globalThisCase},
	)
}
