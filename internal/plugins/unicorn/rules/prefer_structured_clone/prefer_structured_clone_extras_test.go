package prefer_structured_clone_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_structured_clone"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferStructuredCloneBoundaries(t *testing.T) {
	parenthesized := invalidFunction(
		"((_.cloneDeep))(foo)",
		"_.cloneDeep",
		"((structuredClone))(foo)",
	)
	spacedConfig := invalidFunction(
		"my.cloneDeep(foo)",
		"my.cloneDeep",
		"structuredClone(foo)",
		"  my.cloneDeep  ",
	)

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_structured_clone.PreferStructuredCloneRule,
		[]rule_tester.ValidTestCase{
			{
				Code:            "JSON.parse((JSON.stringify as any)(foo))",
				FileName:        "case.ts",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			},
			valid("JSON.parse(JSON[\"stringify\"](foo))"),
			valid("JSON[\"parse\"](JSON.stringify(foo))"),
		},
		[]rule_tester.InvalidTestCase{
			parenthesized,
			spacedConfig,
		},
	)
}
