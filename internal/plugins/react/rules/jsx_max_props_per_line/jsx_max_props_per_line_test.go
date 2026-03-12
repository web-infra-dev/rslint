package jsx_max_props_per_line

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxMaxPropsPerLineRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxMaxPropsPerLineRule, []rule_tester.ValidTestCase{
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
			Code:    `var x = <div foo="bar" baz="qux" />`,
			Tsx:     true,
			Options: map[string]interface{}{"maximum": float64(2)},
		},
		{
			Code:    `var x = <div foo="bar" baz="qux" />`,
			Tsx:     true,
			Options: map[string]interface{}{"when": "multiline"},
		},
		{
			// maximum as object: single=2 allows 2 props on single-line
			Code: `var x = <div foo="bar" baz="qux" />`,
			Tsx:  true,
			Options: map[string]interface{}{
				"maximum": map[string]interface{}{"single": float64(2), "multi": float64(1)},
			},
		},
		{
			// maximum as object + when="multiline": single limit from object still applies (maximumIsObject=true, when doesn't clear singleLimit)
			Code: `var x = <div foo="bar" baz="qux" />`,
			Tsx:  true,
			Options: map[string]interface{}{
				"maximum": map[string]interface{}{"single": float64(2), "multi": float64(1)},
				"when":    "multiline",
			},
		},
		{
			// No props at all
			Code: `var x = <App />`,
			Tsx:  true,
		},
		{
			// Single prop is valid with default max=1
			Code: `var x = <App foo />`,
			Tsx:  true,
		},
		{
			// Spread prop on multiline is valid with when=multiline
			Code: `var x = <App
  foo
  {...this.props}
/>`,
			Tsx:     true,
			Options: map[string]interface{}{"when": "multiline"},
		},
		{
			// Three props with when=multiline on single line
			Code:    `var x = <App foo bar baz />`,
			Tsx:     true,
			Options: map[string]interface{}{"when": "multiline"},
		},
		{
			// Multiline: each prop on its own line with max=1
			Code: `var x = <App
  foo
  bar
  baz
/>`,
			Tsx: true,
		},
		{
			// Multi-line prop: foo spans multiple lines, bar on its own line — valid with max=1
			Code: `var x = <App
  foo={{
    a: 1
  }}
  bar
/>`,
			Tsx: true,
		},
	}, []rule_tester.InvalidTestCase{
		{
			// when="multiline": multiline tag with too many props per line
			Code: `var x = <div
  foo="bar" baz="qux"
/>`,
			Tsx:     true,
			Options: map[string]interface{}{"when": "multiline"},
			Output:  []string{"var x = <div\n  foo=\"bar\"\nbaz=\"qux\"\n/>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "newLine",
					Line:      2,
					Column:    13,
				},
			},
		},
		{
			Code: `var x = <div foo="bar" baz="qux" />`,
			Tsx:  true,
			Output: []string{"var x = <div foo=\"bar\"\nbaz=\"qux\" />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "newLine",
					Line:      1,
					Column:    24,
				},
			},
		},
		{
			Code:    `var x = <div foo="bar" baz="qux" abc="def" />`,
			Tsx:     true,
			Options: map[string]interface{}{"maximum": float64(2)},
			Output:  []string{"var x = <div foo=\"bar\" baz=\"qux\"\nabc=\"def\" />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "newLine",
					Line:      1,
					Column:    34,
				},
			},
		},
		{
			Code: `var x = <div
  foo="bar" baz="qux"
/>`,
			Tsx:    true,
			Output: []string{"var x = <div\n  foo=\"bar\"\nbaz=\"qux\"\n/>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "newLine",
					Line:      2,
					Column:    13,
				},
			},
		},
		{
			// maximum as object: multi=1, multi-line tag with 2 props on same line should error
			Code: `var x = <div
  foo="bar" baz="qux"
/>`,
			Tsx: true,
			Options: map[string]interface{}{
				"maximum": map[string]interface{}{"single": float64(3), "multi": float64(1)},
			},
			Output: []string{"var x = <div\n  foo=\"bar\"\nbaz=\"qux\"\n/>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "newLine",
					Line:      2,
					Column:    13,
				},
			},
		},
		{
			// maximum as object: single=1, single-line tag with 2 props should error
			Code: `var x = <div foo="bar" baz="qux" />`,
			Tsx:  true,
			Options: map[string]interface{}{
				"maximum": map[string]interface{}{"single": float64(1), "multi": float64(3)},
			},
			Output: []string{"var x = <div foo=\"bar\"\nbaz=\"qux\" />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "newLine",
					Line:      1,
					Column:    24,
				},
			},
		},
		{
			// Three props with default max=1: one error at first excess prop
			Code: `var x = <App foo bar baz />`,
			Tsx:  true,
			Output: []string{"var x = <App foo\nbar baz />", "var x = <App foo\nbar\nbaz />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "newLine",
					Line:      1,
					Column:    18,
				},
			},
		},
		{
			// Spread attribute with max=1: first excess prop is bar
			Code: `var x = <App {...this.props} bar />`,
			Tsx:  true,
			Output: []string{"var x = <App {...this.props}\nbar />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "newLine",
					Line:      1,
					Column:    30,
				},
			},
		},
		{
			// Multi-line prop: foo ends on line 4, bar and baz on same line = 3 props grouped, error at baz
			Code: `var x = <App
  foo={{
    a: 1
  }} bar baz
/>`,
			Tsx:     true,
			Options: map[string]interface{}{"maximum": float64(2)},
			Output: []string{"var x = <App\n  foo={{\n    a: 1\n  }} bar\nbaz\n/>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "newLine",
					Line:      4,
					Column:    10,
				},
			},
		},
		{
			// Multiline with too many props per line: max=2
			Code: `var x = <App
  foo bar baz
/>`,
			Tsx:     true,
			Options: map[string]interface{}{"maximum": float64(2)},
			Output: []string{"var x = <App\n  foo bar\nbaz\n/>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "newLine",
					Line:      2,
					Column:    11,
				},
			},
		},
	})
}
