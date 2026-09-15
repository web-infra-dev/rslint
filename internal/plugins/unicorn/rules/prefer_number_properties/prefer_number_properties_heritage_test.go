package prefer_number_properties_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_number_properties"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferNumberPropertiesHeritage(t *testing.T) {
	// Checked against eslint-plugin-unicorn: heritage names are member reads,
	// while qualified names inside ordinary types and type arguments are not.
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&prefer_number_properties.PreferNumberPropertiesRule,
		[]rule_tester.ValidTestCase{
			{Code: `type I = globalThis.parseInt;`},
			{Code: `type I = typeof globalThis.parseInt;`},
			{Code: `interface I extends Box<globalThis.parseInt> {}`},
			{Code: `class C implements Box<globalThis.parseInt> {}`},
		},
		[]rule_tester.InvalidTestCase{
			fixed(`interface I extends globalThis.parseInt {}`, `interface I extends Number.parseInt {}`, "parseInt", "parseInt", 1, 21, 1, 40),
			fixed(`class C implements globalThis.parseInt {}`, `class C implements Number.parseInt {}`, "parseInt", "parseInt", 1, 20, 1, 39),
			fixed(`class C extends globalThis.parseInt {}`, `class C extends Number.parseInt {}`, "parseInt", "parseInt", 1, 17, 1, 36),
		},
	)
}
