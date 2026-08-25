package no_invalid_this

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestReviewComments(t *testing.T) {
	unexpected := func(line, column int) []rule_tester.InvalidTestCaseError {
		return []rule_tester.InvalidTestCaseError{
			{MessageId: "unexpectedThis", Line: line, Column: column},
		}
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoInvalidThisRule,
		[]rule_tester.ValidTestCase{
			// Configured script semantics win over accepted module syntax.
			{
				Code:            `export {}; this; function f(){ this; }`,
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			// Class decorators are evaluated in the enclosing, sloppy script scope.
			{
				Code:            `@dec(function(){ this; }) class C {}`,
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			// ES3 treats directive-looking string literals as inert.
			{
				Code:            `function foo() { "use strict"; this; }`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 3, SourceType: "script"},
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
			// Auto-accessor initializers receive the same field frame as regular fields.
			{Code: `class C { accessor x = function(){ this; } }`},
		},
		[]rule_tester.InvalidTestCase{
			// Configured module semantics apply even without import/export syntax.
			{
				Code:            `this; function f(){ this; }`,
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedThis", Line: 1, Column: 1},
					{MessageId: "unexpectedThis", Line: 1, Column: 21},
				},
			},
			// Ancestor lookup accepts only adjacent JSDoc block comments.
			{
				Code:   `export {}; /* @this */ const x = [function(){ this; }];`,
				Errors: unexpected(1, 47),
			},
			{
				Code: `export {}; /** @this */

const x = [function(){ this; }];`,
				Errors: unexpected(3, 24),
			},
			// A class decorator adds no frame while walking out to a member decorator.
			{
				Code:   `export {}; function outer() { class C { @dec((@d(this) class I {})) m(){} } }`,
				Errors: unexpected(1, 50),
			},
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
