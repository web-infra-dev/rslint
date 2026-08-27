package no_restricted_properties

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoRestrictedPropertiesJsxReview locks in the JSXMemberExpression parity
// requested in PR #1750's review. Dotted JSX tag names are not ordinary member
// expressions in ESLint, while member access inside JSX expressions still is.
func TestNoRestrictedPropertiesJsxReview(t *testing.T) {
	options := []any{
		map[string]any{"object": "foo", "property": "bar"},
		map[string]any{"property": "baz"},
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoRestrictedPropertiesRule,
		[]rule_tester.ValidTestCase{
			{Code: `const element = <foo.bar />;`, Options: options, Tsx: true},
			{Code: `const element = <foo.bar></foo.bar>;`, Options: options, Tsx: true},
			{Code: `const element = <foo.bar.baz />;`, Options: options, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    `const element = <div value={foo.bar} />;`,
				Options: options,
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedObjectProperty",
					Message:   "'foo.bar' is restricted from being used.",
					Line:      1,
					Column:    29,
					EndLine:   1,
					EndColumn: 36,
				}},
			},
			{
				Code:    `const element = <div>{foo.bar.baz}</div>;`,
				Options: options,
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedProperty",
						Message:   "'baz' is restricted from being used.",
						Line:      1,
						Column:    23,
						EndLine:   1,
						EndColumn: 34,
					},
					{
						MessageId: "restrictedObjectProperty",
						Message:   "'foo.bar' is restricted from being used.",
						Line:      1,
						Column:    23,
						EndLine:   1,
						EndColumn: 30,
					},
				},
			},
		},
	)
}
