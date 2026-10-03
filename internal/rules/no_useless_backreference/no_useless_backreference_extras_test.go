package no_useless_backreference

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUselessBackreferenceExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUselessBackreferenceRule, []rule_tester.ValidTestCase{
		// Evaluating the computed method first must not hide the array mutation.
		{Code: `const method = ["reverse"];
const patterns = ["\\1(a)", "(a)"];
RegExp(method[0]);
patterns[method[0]]();
RegExp(patterns[0]);`},
	}, nil)
}
