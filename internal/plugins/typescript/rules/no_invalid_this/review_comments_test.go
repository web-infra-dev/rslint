package no_invalid_this

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestReviewComments(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoInvalidThisRule,
		[]rule_tester.ValidTestCase{
			{Code: `/* @thisX */ function g(){ this; }`},
			{Code: `Uint8Array.from([], function () { this; }, obj);`},
			{Code: `BigInt64Array['from']([], function () { this; }, obj);`},
		},
		nil,
	)
}
