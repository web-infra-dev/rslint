package jsx_props_no_multi_spaces

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxPropsNoMultiSpacesRuleExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxPropsNoMultiSpacesRule, []rule_tester.ValidTestCase{
		{
			Code: `<X<T /* > */> a/>`,
			Tsx:  true,
		},
		{
			Code: `<X<Map<string, Array<number /* > */>>> a/>`,
			Tsx:  true,
		},
		{
			Code: `<X a /*

*/
 b/>`,
			Tsx: true,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code:   `<X<T /* > */>  a/>`,
			Tsx:    true,
			Output: []string{`<X<T /* > */> a/>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "onlyOneSpace",
					Message:   `Expected only one space between "X" and "a"`,
					Line:      1,
					Column:    16,
				},
			},
		},
		{
			Code:   `<X<T>  a/>`,
			Tsx:    true,
			Output: []string{`<X<T> a/>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "onlyOneSpace",
					Message:   `Expected only one space between "X" and "a"`,
					Line:      1,
					Column:    8,
				},
			},
		},
		{
			Code: `<X a /*

*/

 b/>`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noLineGap",
					Message:   `Expected no line gap between "a" and "b"`,
					Line:      5,
					Column:    2,
				},
			},
		},
		{
			Code: "<X a\n\u00a0\n b/>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noLineGap",
					Message:   `Expected no line gap between "a" and "b"`,
					Line:      3,
					Column:    2,
				},
			},
		},
		{
			Code:   `<X a  {...props}/>`,
			Tsx:    true,
			Output: []string{`<X a {...props}/>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "onlyOneSpace",
					Message:   `Expected only one space between "a" and "props"`,
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `<X {...props}  a/>`,
			Tsx:    true,
			Output: []string{`<X {...props} a/>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "onlyOneSpace",
					Message:   `Expected only one space between "props" and "a"`,
					Line:      1,
					Column:    16,
				},
			},
		},
	})
}
