package no_invalid_this

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestReviewComments(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoInvalidThisRule,
		[]rule_tester.ValidTestCase{
			// ES3 treats directive-looking string literals as inert.
			{
				Code:            `function foo() { "use strict"; this; }`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 3},
			},
			// ESLint's @this matcher intentionally accepts tag-name prefixes.
			{Code: `export {}; function f(){ /* @thisX */ function g(){ this; } }`},
			// Array.from recognition includes TypedArray constructor names.
			{Code: `export {}; Uint8Array.from([], function () { this; }, obj);`},
			{Code: `export {}; BigInt64Array['from']([], function () { this; }, obj);`},
		},
		nil,
	)
}
