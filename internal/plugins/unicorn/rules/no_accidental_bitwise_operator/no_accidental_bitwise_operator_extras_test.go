package no_accidental_bitwise_operator_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_accidental_bitwise_operator"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoAccidentalBitwiseOperatorExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t,
		&no_accidental_bitwise_operator.NoAccidentalBitwiseOperatorRule,
		[]rule_tester.ValidTestCase{
			valid(`obj & other.obj;`, "file.js"),
			valid(`options | (value as number);`, "file.ts"),
		},
		[]rule_tester.InvalidTestCase{
			invalid(`options | ({});`, `|`, `||`, "file.js"),
			invalid(`obj & (obj.value);`, `&`, `&&`, "file.js"),
		},
	)
}
