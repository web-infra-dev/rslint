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
			// JavaScript's regexp whitespace includes non-ASCII ECMAScript whitespace.
			{Code: "export {}; function f(){ /** \u00a0@this */ function g(){ this; } }"},
			{Code: "export {}; function f(){ /** \ufeff@this */ function g(){ this; } }"},
			// Declaration-level JSDoc applies through expression containers.
			{Code: `export {}; /** @this */ const x = [function(){ this; }];`},
			{Code: `export {}; /** @this */ const x = !function(){ this; };`},
			// Array.from recognition includes TypedArray constructor names.
			{Code: `export {}; Uint8Array.from([], function () { this; }, obj);`},
			{Code: `export {}; BigInt64Array['from']([], function () { this; }, obj);`},
		},
		[]rule_tester.InvalidTestCase{
			// ES3 disables directive strictness, not strictness inherited from modules or classes.
			{
				Code:            `export {}; function foo(){ this; }`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 3},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 28}},
			},
			{
				Code:            `class C { m(){ function foo(){ this; } } }`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 3},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 32}},
			},
		},
	)
}
