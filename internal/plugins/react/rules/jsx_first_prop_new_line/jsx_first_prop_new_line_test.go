package jsx_first_prop_new_line

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxFirstPropNewLineRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxFirstPropNewLineRule, []rule_tester.ValidTestCase{
		{
			Code: `var x = <div foo="bar" />`,
			Tsx:  true,
		},
		{
			Code: `var x = <div foo="bar" baz="qux" />`,
			Tsx:  true,
		},
		{
			Code: `var x = <div
  foo="bar"
/>`,
			Tsx:     true,
			Options: "always",
		},
		{
			Code:    `var x = <div foo="bar" />`,
			Tsx:     true,
			Options: "never",
		},
		{
			Code: `var x = <div
  foo="bar"
  baz="qux"
/>`,
			Tsx: true,
		},
		{
			// "never": no props is valid
			Code:    `var x = <Foo />`,
			Tsx:     true,
			Options: "never",
		},
		{
			// "multiline-multiprop": single prop on same line is valid
			Code:    `var x = <Foo bar />`,
			Tsx:     true,
			Options: "multiline-multiprop",
		},
		{
			// "multiline-multiprop": multiple props on same single line is valid
			Code:    `var x = <Foo bar baz />`,
			Tsx:     true,
			Options: "multiline-multiprop",
		},
		{
			// "multiline-multiprop": multiline with first prop on new line is valid
			Code: `var x = <Foo
  foo={{}}
  bar
/>`,
			Tsx:     true,
			Options: "multiline-multiprop",
		},
		{
			// "multiprop": single prop is valid on same line
			Code:    `var x = <Foo bar />`,
			Tsx:     true,
			Options: "multiprop",
		},
		{
			// "multiprop": no props is valid
			Code:    `var x = <Foo />`,
			Tsx:     true,
			Options: "multiprop",
		},
		{
			// "always": no props is valid
			Code:    `var x = <Foo />`,
			Tsx:     true,
			Options: "always",
		},
		{
			// "multiline": single-line single prop is valid
			Code:    `var x = <Foo bar />`,
			Tsx:     true,
			Options: "multiline",
		},
		{
			// "multiline": multiline with first prop on new line is valid
			Code: `var x = <Foo
  propOne="one"
  propTwo="two"
/>`,
			Tsx:     true,
			Options: "multiline",
		},
	}, []rule_tester.InvalidTestCase{
		{
			// Default mode (multiline-multiprop): multiline with multiple props, first on same line
			Code: `var x = <Foo bar baz
/>`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "propOnNewLine",
					Line:      1,
					Column:    14,
				},
			},
			Output: []string{`var x = <Foo
bar baz
/>`},
		},
		{
			Code:    `var x = <div foo="bar" />`,
			Tsx:     true,
			Options: "always",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "propOnNewLine",
					Line:      1,
					Column:    14,
				},
			},
			Output: []string{"var x = <div\nfoo=\"bar\" />"},
		},
		{
			Code: `var x = <div
  foo="bar"
/>`,
			Tsx:     true,
			Options: "never",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "propOnSameLine",
					Line:      2,
					Column:    3,
				},
			},
			Output: []string{`var x = <div foo="bar"
/>`},
		},
		{
			Code: `var x = <div foo="bar"
  baz="qux"
/>`,
			Tsx:     true,
			Options: "multiprop",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "propOnNewLine",
					Line:      1,
					Column:    14,
				},
			},
			Output: []string{`var x = <div
foo="bar"
  baz="qux"
/>`},
		},
		{
			// "always": multiline first prop on same line
			Code:    `var x = <Foo propOne="one" propTwo="two" />`,
			Tsx:     true,
			Options: "always",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "propOnNewLine",
					Line:      1,
					Column:    14,
				},
			},
			Output: []string{"var x = <Foo\npropOne=\"one\" propTwo=\"two\" />"},
		},
		{
			// "multiline": prop on same line when multiline
			Code: `var x = <Foo prop={{}}
/>`,
			Tsx:     true,
			Options: "multiline",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "propOnNewLine",
					Line:      1,
					Column:    14,
				},
			},
			Output: []string{`var x = <Foo
prop={{}}
/>`},
		},
		{
			// "multiline-multiprop": multiline with first prop on same line
			Code: `var x = <Foo bar={{}} baz
/>`,
			Tsx:     true,
			Options: "multiline-multiprop",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "propOnNewLine",
					Line:      1,
					Column:    14,
				},
			},
			Output: []string{`var x = <Foo
bar={{}} baz
/>`},
		},
		{
			// "multiprop": multiple props on same line should error
			Code:    `var x = <Foo propOne="one" propTwo="two" />`,
			Tsx:     true,
			Options: "multiprop",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "propOnNewLine",
					Line:      1,
					Column:    14,
				},
			},
			Output: []string{"var x = <Foo\npropOne=\"one\" propTwo=\"two\" />"},
		},
		{
			// "multiprop": single prop on new line in multiline element should be on same line
			Code: `var x = <Foo
  bar
/>`,
			Tsx:     true,
			Options: "multiprop",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "propOnSameLine",
					Line:      2,
					Column:    3,
				},
			},
			Output: []string{`var x = <Foo bar
/>`},
		},
	})
}
