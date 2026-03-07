package self_closing_comp

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestSelfClosingCompRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &SelfClosingCompRule, []rule_tester.ValidTestCase{
		{
			Code: `var x = <div>Hello</div>`,
			Tsx:  true,
		},
		{
			Code: `var x = <Component>children</Component>`,
			Tsx:  true,
		},
		{
			Code: `var x = <div />`,
			Tsx:  true,
		},
		{
			Code: `var x = <Component />`,
			Tsx:  true,
		},
		{
			Code:    `var x = <div></div>`,
			Tsx:     true,
			Options: map[string]interface{}{"html": false},
		},
		{
			// Spaces without newline are NOT treated as empty (matches ESLint)
			Code: `var x = <Component>  </Component>`,
			Tsx:  true,
		},
		{
			// Compound component self-closing is valid
			Code: `var x = <Hello.Compound name="John" />`,
			Tsx:  true,
		},
		{
			// Component with entity content is not empty
			Code: `var x = <Hello name="John">&nbsp;</Hello>`,
			Tsx:  true,
		},
		{
			// component: false allows non-self-closing components
			Code:    `var x = <Hello name="John"></Hello>`,
			Tsx:     true,
			Options: map[string]interface{}{"component": false},
		},
		{
			// html: false allows non-self-closing HTML elements
			Code:    `var x = <div className="content"></div>`,
			Tsx:     true,
			Options: map[string]interface{}{"component": true, "html": false},
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `var x = <Component></Component>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "notSelfClosing",
					Line:      1,
					Column:    9,
				},
			},
			Output: []string{`var x = <Component />`},
		},
		{
			Code: `var x = <div></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "notSelfClosing",
					Line:      1,
					Column:    9,
				},
			},
			Output: []string{`var x = <div />`},
		},
		{
			// Whitespace with newline IS treated as empty
			Code: "var x = <Component>\n</Component>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "notSelfClosing",
					Line:      1,
					Column:    9,
				},
			},
			Output: []string{`var x = <Component />`},
		},
		{
			// Compound component without children should self-close
			Code: `var x = <Hello.Compound name="John"></Hello.Compound>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "notSelfClosing",
					Line:      1,
					Column:    9,
				},
			},
			Output: []string{`var x = <Hello.Compound name="John" />`},
		},
		{
			// html: true explicitly — empty div should self-close
			Code:    `var x = <div className="content"></div>`,
			Tsx:     true,
			Options: map[string]interface{}{"html": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "notSelfClosing",
					Line:      1,
					Column:    9,
				},
			},
			Output: []string{`var x = <div className="content" />`},
		},
	})
}
