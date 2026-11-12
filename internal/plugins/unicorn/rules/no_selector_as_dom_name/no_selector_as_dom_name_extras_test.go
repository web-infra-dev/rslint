package no_selector_as_dom_name_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_selector_as_dom_name"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoSelectorAsDomNameExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_selector_as_dom_name.NoSelectorAsDomNameRule,
		[]rule_tester.ValidTestCase{
			{Code: `element["classList"].add(".foo");`, FileName: "file.js"},
			{Code: `(element?.classList).add(".foo");`, FileName: "file.js"},
			{Code: `document.getElementById("#app", extra);`, FileName: "file.js"},
			{Code: `document.getElementById(...ids);`, FileName: "file.js"},
			{Code: `element.classList.replace("old", ...tokens, ".extra");`, FileName: "file.js"},
		}, []rule_tester.InvalidTestCase{
			invalidCase(`element.classList.add(...tokens, ".tail");`, "file.js", `element.classList.add(...tokens, "tail");`, [4]int{1, 34, 1, 41}),
			invalidCase(`element.classList.replace(".old", ...tokens);`, "file.js", `element.classList.replace("old", ...tokens);`, [4]int{1, 27, 1, 33}),
			invalidCase(`document.getElementById(/* 😀 */ "#app");`, "file.js", `document.getElementById(/* 😀 */ "app");`, [4]int{1, 34, 1, 40}),
			invalidCase("element.classList.add(`.${first}${second}`);", "file.js", "", [4]int{1, 23, 1, 43}),
			invalidCase(`element.classList.add("._class-1", ".-name");`, "file.js", `element.classList.add("_class-1", "-name");`, [4]int{1, 23, 1, 34}, [4]int{1, 36, 1, 44}),
		})
}
