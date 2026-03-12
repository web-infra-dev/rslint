package jsx_equals_spacing

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxEqualsSpacingRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxEqualsSpacingRule, []rule_tester.ValidTestCase{
		{
			Code: `var x = <div foo="bar" />`,
			Tsx:  true,
		},
		{
			Code: `var x = <div foo={bar} />`,
			Tsx:  true,
		},
		{
			Code: `var x = <div foo />`,
			Tsx:  true,
		},
		{
			Code: `var x = <div foo="bar" baz={qux} />`,
			Tsx:  true,
		},
		{
			// "always" mode: spaces around = are valid
			Code:    `var x = <div foo = "bar" />`,
			Tsx:     true,
			Options: "always",
		},
		{
			// Spread attribute (default: no issue)
			Code: `var x = <App {...props} />`,
			Tsx:  true,
		},
		{
			// Boolean attribute with "always"
			Code:    `var x = <App foo />`,
			Tsx:     true,
			Options: "always",
		},
		{
			// Expression attribute with "always"
			Code:    `var x = <App foo = {e => bar(e)} />`,
			Tsx:     true,
			Options: "always",
		},
		{
			// Multiple valid attributes (default)
			Code: `var x = <App foo="bar" baz={qux} quux />`,
			Tsx:  true,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `var x = <div foo ="bar" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noSpaceBefore",
					Line:      1,
					Column:    17,
				},
			},
			Output: []string{`var x = <div foo="bar" />`},
		},
		{
			// "never" mode: space after = should error
			Code: `var x = <div foo= "bar" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noSpaceAfter",
					Line:      1,
					Column:    18,
				},
			},
			Output: []string{`var x = <div foo="bar" />`},
		},
		{
			// "always" mode: no spaces around = should error
			Code:    `var x = <div foo="bar" />`,
			Tsx:     true,
			Options: "always",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "needSpaceBefore",
					Line:      1,
					Column:    17,
				},
				{
					MessageId: "needSpaceAfter",
					Line:      1,
					Column:    18,
				},
			},
			Output: []string{`var x = <div foo = "bar" />`},
		},
		{
			// Both spaces in "never" mode: two errors
			Code: `var x = <div foo = {bar} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noSpaceBefore",
					Line:      1,
					Column:    17,
				},
				{
					MessageId: "noSpaceAfter",
					Line:      1,
					Column:    19,
				},
			},
			Output: []string{`var x = <div foo={bar} />`},
		},
		{
			// Multiple attrs in "never" mode: errors on each
			Code: `var x = <div foo= {bar} baz = {qux} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noSpaceAfter",
					Line:      1,
					Column:    18,
				},
				{
					MessageId: "noSpaceBefore",
					Line:      1,
					Column:    28,
				},
				{
					MessageId: "noSpaceAfter",
					Line:      1,
					Column:    30,
				},
			},
			Output: []string{`var x = <div foo={bar} baz={qux} />`},
		},
		{
			// "always" mode: only space after missing
			Code:    `var x = <div foo ={bar} />`,
			Tsx:     true,
			Options: "always",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "needSpaceAfter",
					Line:      1,
					Column:    19,
				},
			},
			Output: []string{`var x = <div foo = {bar} />`},
		},
		{
			// "always" mode: only space before missing
			Code:    `var x = <div foo= {bar} />`,
			Tsx:     true,
			Options: "always",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "needSpaceBefore",
					Line:      1,
					Column:    17,
				},
			},
			Output: []string{`var x = <div foo = {bar} />`},
		},
		{
			// "always" mode: multiple attrs with errors
			Code:    `var x = <div foo={bar} baz ={qux} />`,
			Tsx:     true,
			Options: "always",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "needSpaceBefore",
					Line:      1,
					Column:    17,
				},
				{
					MessageId: "needSpaceAfter",
					Line:      1,
					Column:    18,
				},
				{
					MessageId: "needSpaceAfter",
					Line:      1,
					Column:    29,
				},
			},
			Output: []string{`var x = <div foo = {bar} baz = {qux} />`},
		},
	})
}
