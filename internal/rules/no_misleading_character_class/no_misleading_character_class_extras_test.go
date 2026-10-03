package no_misleading_character_class

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoMisleadingCharacterClassExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMisleadingCharacterClassRule, []rule_tester.ValidTestCase{
		// Evaluating the computed method first must not hide the array mutation.
		{Code: `const method = ["reverse"];
const patterns = ["[👍]", "a"];
RegExp(method[0]);
patterns[method[0]]();
RegExp(patterns[0]);`},
	}, nil)
}
