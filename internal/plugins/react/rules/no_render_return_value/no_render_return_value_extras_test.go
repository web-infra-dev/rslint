package no_render_return_value

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoRenderReturnValueExtras(t *testing.T) {
	defaultReact013 := map[string]interface{}{
		"react": map[string]interface{}{"defaultVersion": "0.13.0"},
	}
	explicitReact15 := map[string]interface{}{
		"react": map[string]interface{}{
			"version":        "15.0.0",
			"defaultVersion": "0.13.0",
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRenderReturnValueRule, []rule_tester.ValidTestCase{
		{
			Code:     `var instance = ReactDOM.render(<div />, document.body);`,
			Tsx:      true,
			Settings: defaultReact013,
		},
		{
			Code:     `var instance = React.render(<div />, document.body);`,
			Tsx:      true,
			Settings: explicitReact15,
		},
		{
			Code: `class Example {
				[ReactDOM.render(<div />, document.body)]() {}
				get [ReactDOM.render(<div />, document.body)]() { return true; }
				set [ReactDOM.render(<div />, document.body)](value) {}
				[ReactDOM.render(<div />, document.body)] = true;
			}`,
			Tsx: true,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code:     `var instance = React.render(<div />, document.body);`,
			Tsx:      true,
			Settings: defaultReact013,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue", Message: "Do not depend on the return value from React.render"},
			},
		},
		{
			Code:     `var instance = ReactDOM.render(<div />, document.body);`,
			Tsx:      true,
			Settings: explicitReact15,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue", Message: "Do not depend on the return value from ReactDOM.render"},
			},
		},
		{
			Code: `var value = {
				[ReactDOM.render(<div />, document.body)]: true,
			};`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue", Line: 2, Column: 6},
			},
		},
		{
			Code: `var value = {
				[ReactDOM.render(<div />, document.body)]() {},
			};`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue", Line: 2, Column: 6},
			},
		},
		{
			Code: `var value = {
				get [ReactDOM.render(<div />, document.body)]() { return true; },
				set [ReactDOM.render(<div />, document.body)](next) {},
			};`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue"},
				{MessageId: "noReturnValue"},
			},
		},
		{
			Code: `var value;
			({ [ReactDOM.render(<div />, document.body)]: value } = source);`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue", Line: 2, Column: 8},
			},
		},
		{
			Code: `function read({
				[ReactDOM.render(<div />, document.body)]: value,
			}) {}`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue", Line: 2, Column: 6},
			},
		},

		// Intentional differences from upstream: both forms consume the call's
		// result even though ESTree inserts a ChainExpression or classifies the
		// inner assignment as an AssignmentPattern.
		{
			Code: `var instance = ReactDOM?.render(<div />, document.body);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue"},
			},
		},
		{
			Code: `var value; [value = ReactDOM.render(<div />, document.body)] = source;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue"},
			},
		},
	})
}
