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
			{
				Code:            `export {}; this; function f(){ this; }`,
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			{Code: `/* @thisX */ function g(){ this; }`},
			{Code: "/** \u00a0@this */ function g(){ this; }"},
			{Code: "/** \ufeff@this */ function g(){ this; }"},
			{Code: `/** @this */ const x = [function(){ this; }];`},
			{Code: `/** @this */ const x = !function(){ this; };`},
			{Code: `Uint8Array.from([], function () { this; }, obj);`},
			{Code: `BigInt64Array['from']([], function () { this; }, obj);`},
			{Code: `class C { accessor x = function(){ this; } }`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:            `this; function f(){ this; }`,
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedThis", Line: 1, Column: 1},
					{MessageId: "unexpectedThis", Line: 1, Column: 21},
				},
			},
			{
				Code:   `/* @this */ const x = [function(){ this; }];`,
				Errors: unexpected(1, 36),
			},
			{
				Code: `/** @this */

const x = [function(){ this; }];`,
				Errors: unexpected(3, 24),
			},
			{
				Code:   `function outer() { class C { @dec((@d(this) class I {})) m(){} } }`,
				Errors: unexpected(1, 39),
			},
		},
	)
}
