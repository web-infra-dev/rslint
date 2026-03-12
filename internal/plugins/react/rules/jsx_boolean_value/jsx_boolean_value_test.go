package jsx_boolean_value

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxBooleanValueRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxBooleanValueRule, []rule_tester.ValidTestCase{
		{
			Code: `var x = <div disabled />`,
			Tsx:  true,
		},
		{
			Code:    `var x = <div disabled={true} />`,
			Tsx:     true,
			Options: "always",
		},
		{
			Code:    `var x = <div disabled />`,
			Tsx:     true,
			Options: "never",
		},
		{
			Code: `var x = <div disabled={someVar} />`,
			Tsx:  true,
		},
		{
			// "always" mode with "never" exceptions: disabled uses "never" behavior
			Code:    `var x = <div disabled />`,
			Tsx:     true,
			Options: []interface{}{"always", map[string]interface{}{"never": []interface{}{"disabled"}}},
		},
		{
			// "never" mode with "always" exceptions: autoFocus uses "always" behavior
			Code:    `var x = <div autoFocus={true} />`,
			Tsx:     true,
			Options: []interface{}{"never", map[string]interface{}{"always": []interface{}{"autoFocus"}}},
		},
		{
			Code: `var x = <div disabled={false} />`,
			Tsx:  true,
		},
		{
			// "always" with "never" exceptions: mixed valid
			Code:    `var x = <div foo bar={true} />`,
			Tsx:     true,
			Options: []interface{}{"always", map[string]interface{}{"never": []interface{}{"foo"}}},
		},
		{
			// "never" with "always" exceptions: mixed valid
			Code:    `var x = <div foo={true} bar />`,
			Tsx:     true,
			Options: []interface{}{"never", map[string]interface{}{"always": []interface{}{"foo"}}},
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `var x = <div disabled={true} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "omitBoolean",
					Line:      1,
					Column:    14,
				},
			},
			Output: []string{`var x = <div disabled />`},
		},
		{
			Code:    `var x = <div disabled />`,
			Tsx:     true,
			Options: "always",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "setBoolean",
					Line:      1,
					Column:    14,
				},
			},
			Output: []string{`var x = <div disabled={true} />`},
		},
		{
			// "always" mode with "never" exceptions: non-exception prop without value should error
			Code:    `var x = <div autoFocus />`,
			Tsx:     true,
			Options: []interface{}{"always", map[string]interface{}{"never": []interface{}{"disabled"}}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "setBoolean",
					Line:      1,
					Column:    14,
				},
			},
			Output: []string{`var x = <div autoFocus={true} />`},
		},
		{
			// "never" mode with "always" exceptions: exception prop without value should error
			Code:    `var x = <div autoFocus />`,
			Tsx:     true,
			Options: []interface{}{"never", map[string]interface{}{"always": []interface{}{"autoFocus"}}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "setBoolean",
					Line:      1,
					Column:    14,
				},
			},
			Output: []string{`var x = <div autoFocus={true} />`},
		},
		{
			Code:    `var x = <div disabled={false} />`,
			Tsx:     true,
			Options: []interface{}{"never", map[string]interface{}{"assumeUndefinedIsFalse": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "omitPropAndBoolean",
					Line:      1,
					Column:    14,
				},
			},
			Output: []string{`var x = <div />`},
		},
		{
			// Multiple false values with assumeUndefinedIsFalse: both removed
			Code:    `var x = <div foo={false} bar={false} />`,
			Tsx:     true,
			Options: []interface{}{"never", map[string]interface{}{"assumeUndefinedIsFalse": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "omitPropAndBoolean",
					Line:      1,
					Column:    14,
				},
				{
					MessageId: "omitPropAndBoolean",
					Line:      1,
					Column:    26,
				},
			},
			Output: []string{`var x = <div />`},
		},
		{
			// Mixed true and false with assumeUndefinedIsFalse
			Code:    `var x = <div foo={true} bar={false} />`,
			Tsx:     true,
			Options: []interface{}{"never", map[string]interface{}{"assumeUndefinedIsFalse": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "omitBoolean",
					Line:      1,
					Column:    14,
				},
				{
					MessageId: "omitPropAndBoolean",
					Line:      1,
					Column:    25,
				},
			},
			Output: []string{`var x = <div foo />`},
		},
		{
			// "always" with "never" exceptions: multiple non-exception props
			Code:    `var x = <div foo={true} bar={true} baz={true} />`,
			Tsx:     true,
			Options: []interface{}{"always", map[string]interface{}{"never": []interface{}{"foo", "bar"}}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "omitBoolean",
					Line:      1,
					Column:    14,
				},
				{
					MessageId: "omitBoolean",
					Line:      1,
					Column:    25,
				},
			},
			Output: []string{`var x = <div foo bar baz={true} />`},
		},
		{
			// "never" with "always" exceptions: multiple exception props without value
			Code:    `var x = <div foo bar baz />`,
			Tsx:     true,
			Options: []interface{}{"never", map[string]interface{}{"always": []interface{}{"foo", "bar"}}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "setBoolean",
					Line:      1,
					Column:    14,
				},
				{
					MessageId: "setBoolean",
					Line:      1,
					Column:    18,
				},
			},
			Output: []string{`var x = <div foo={true} bar={true} baz />`},
		},
	})
}
