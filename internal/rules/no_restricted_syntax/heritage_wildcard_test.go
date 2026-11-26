package no_restricted_syntax

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestHeritageWildcardTraversal(t *testing.T) {
	// Wildcard and negative selectors visit the full TSESTree (Program,
	// declarations, type references, qualified names, heritage, and bodies),
	// matching ESLint + @typescript-eslint/parser.
	for _, tc := range []struct {
		code     string
		selector string
		ranges   [][2]int
	}{
		{`type Good = Foo.Bar;`, "*", [][2]int{{1, 21}, {1, 21}, {6, 10}, {13, 20}, {13, 20}, {13, 16}, {17, 20}}},
		{`type Good = typeof Foo.Bar;`, "*", [][2]int{{1, 28}, {1, 28}, {6, 10}, {13, 27}, {20, 27}, {20, 23}, {24, 27}}},
		{`interface Good extends Box<Foo.Bar> {}`, "*", [][2]int{{1, 39}, {1, 39}, {11, 15}, {24, 36}, {24, 27}, {27, 36}, {28, 35}, {28, 35}, {28, 31}, {32, 35}, {37, 39}}},
		{`interface Good extends Foo.Bar {}`, "*", [][2]int{{1, 34}, {1, 34}, {11, 15}, {24, 31}, {24, 31}, {24, 27}, {28, 31}, {32, 34}}},
		{`type Good = Foo.Bar;`, ":not(Identifier)", [][2]int{{1, 21}, {1, 21}, {13, 20}, {13, 20}}},
		{`interface Good extends Foo.Bar {}`, ":not(Identifier)", [][2]int{{1, 34}, {1, 34}, {24, 31}, {24, 31}, {32, 34}}},
	} {
		t.Run(tc.code+tc.selector, func(t *testing.T) {
			var errors []rule_tester.InvalidTestCaseError
			for _, span := range tc.ranges {
				errors = append(errors, rule_tester.InvalidTestCaseError{
					MessageId: "restrictedSyntax", Message: "Using '" + tc.selector + "' is not allowed.",
					Line: 1, Column: span[0], EndLine: 1, EndColumn: span[1],
				})
			}
			rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRestrictedSyntaxRule, nil,
				[]rule_tester.InvalidTestCase{{Code: tc.code, Options: []any{tc.selector}, Errors: errors}})
		})
	}
}
