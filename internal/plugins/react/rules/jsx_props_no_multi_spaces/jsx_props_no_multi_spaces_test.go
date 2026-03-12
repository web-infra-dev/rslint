package jsx_props_no_multi_spaces

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxPropsNoMultiSpacesRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxPropsNoMultiSpacesRule, []rule_tester.ValidTestCase{
		{
			Code: `var x = <div foo="bar" baz="qux" />`,
			Tsx:  true,
		},
		{
			Code: `var x = <div foo="bar" />`,
			Tsx:  true,
		},
		{
			Code: `var x = <div
  foo="bar"
  baz="qux"
/>`,
			Tsx: true,
		},
		{
			Code: `var x = <div />`,
			Tsx:  true,
		},
		{
			// Namespaced attribute: single space is valid
			Code: `var x = <div foo:bar="baz" qux="quux" />`,
			Tsx:  true,
		},
		{
			// Spread attribute: single space is valid
			Code: `var x = <div {...props} foo="bar" />`,
			Tsx:  true,
		},
		{
			// No props at all
			Code: `var x = <App />`,
			Tsx:  true,
		},
		{
			// Single prop
			Code: `var x = <App foo />`,
			Tsx:  true,
		},
		{
			// Compound component name with single space
			Code: `var x = <Foo.Bar baz="quux" />`,
			Tsx:  true,
		},
		{
			// Long compound component name
			Code: `var x = <Foo.Bar.Baz.Qux quux="corge" />`,
			Tsx:  true,
		},
		{
			// Spaces inside attribute value are OK
			Code: `var x = <App foo="with  spaces   " bar />`,
			Tsx:  true,
		},
		{
			// Comment between props bridges the gap (no blank line)
			Code: `var x = <button
  title="Some button"
  // this is a comment
  onClick={() => {}}
  type="button"
/>`,
			Tsx: true,
		},
		{
			// Two comments between props bridge the gap
			Code: `var x = <button
  title="Some button"
  // first comment
  // second comment
  onClick={() => {}}
/>`,
			Tsx: true,
		},
		{
			// Inline comment + line comment
			Code: `var x = <App
  foo="bar" // comment
  // comment
  bar=""
/>`,
			Tsx: true,
		},
		{
			// TypeScript generic tag name: single space
			Code: `var x = <App<T> foo bar />`,
			Tsx: true,
		},
		{
			// Multi-line prop value: next prop starts on end line of previous prop — treated as same line
			Code: `var x = <App foo={{
  a: 1
}} bar />`,
			Tsx: true,
		},
		{
			// TypeScript generic with proper spacing
			Code: `var x = <App<T> foo="bar" baz="qux" />`,
			Tsx:  true,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code:   `var x = <div foo="bar"  baz="qux" />`,
			Tsx:    true,
			Output: []string{`var x = <div foo="bar" baz="qux" />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "onlyOneSpace",
					Line:      1,
					Column:    25,
				},
			},
		},
		{
			Code:   `var x = <div  foo="bar" />`,
			Tsx:    true,
			Output: []string{`var x = <div foo="bar" />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "onlyOneSpace",
					Line:      1,
					Column:    15,
				},
			},
		},
		{
			Code: `var x = <div
  foo="bar"

  baz="qux"
/>`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noLineGap",
					Line:      4,
					Column:    3,
				},
			},
		},
		{
			// Namespaced attribute: multiple spaces should be reported
			Code:   `var x = <div foo:bar="baz"  qux="quux" />`,
			Tsx:    true,
			Output: []string{`var x = <div foo:bar="baz" qux="quux" />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "onlyOneSpace",
					Line:      1,
					Column:    29,
				},
			},
		},
		{
			// Spread attribute: multiple spaces should be reported
			Code:   `var x = <div {...props}  foo="bar" />`,
			Tsx:    true,
			Output: []string{`var x = <div {...props} foo="bar" />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "onlyOneSpace",
					Line:      1,
					Column:    26,
				},
			},
		},
		{
			// Multiple violations: tag-to-prop and prop-to-prop
			Code:   `var x = <App  foo   bar />`,
			Tsx:    true,
			Output: []string{`var x = <App foo bar />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "onlyOneSpace",
					Line:      1,
					Column:    15,
				},
				{
					MessageId: "onlyOneSpace",
					Line:      1,
					Column:    21,
				},
			},
		},
		{
			// Spread with multiple spaces on both sides
			Code:   `var x = <App foo  {...test}  bar />`,
			Tsx:    true,
			Output: []string{`var x = <App foo {...test} bar />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "onlyOneSpace",
					Line:      1,
					Column:    19,
				},
				{
					MessageId: "onlyOneSpace",
					Line:      1,
					Column:    30,
				},
			},
		},
		{
			// Compound component name with multiple spaces
			Code:   `var x = <Foo.Bar  baz="quux" />`,
			Tsx:    true,
			Output: []string{`var x = <Foo.Bar baz="quux" />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "onlyOneSpace",
					Line:      1,
					Column:    19,
				},
			},
		},
		{
			// Multiple blank lines between props
			Code: `var x = <div
  foo="bar"

  baz="qux"

  quux="corge"
/>`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noLineGap",
					Line:      4,
					Column:    3,
				},
				{
					MessageId: "noLineGap",
					Line:      6,
					Column:    3,
				},
			},
		},
		{
			// TypeScript generic with multiple spaces
			Code:   `var x = <App<T>  foo="bar" />`,
			Tsx:    true,
			Output: []string{`var x = <App<T> foo="bar" />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "onlyOneSpace",
					Line:      1,
					Column:    18,
				},
			},
		},
		{
			// Comment then blank line: only the blank line gap is reported
			Code: `var x = <button
  title="Some button"
  // this is a comment
  onClick={() => {}}

  type="button"
/>`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noLineGap",
					Line:      6,
					Column:    3,
				},
			},
		},
	})
}
