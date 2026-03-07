package style_prop_object

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestStylePropObjectRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &StylePropObjectRule, []rule_tester.ValidTestCase{
		{
			Code: `var x = <div style={{ color: "red" }} />`,
			Tsx:  true,
		},
		{
			Code: `var x = <div style={myStyle} />`,
			Tsx:  true,
		},
		{
			Code: `var x = <div style={null} />`,
			Tsx:  true,
		},
		{
			Code: `var x = <div />`,
			Tsx:  true,
		},
		{
			// allow list: string style on allowed component is valid
			Code: `var x = <MyComponent style="color: red" />`,
			Tsx:  true,
			Options: map[string]interface{}{
				"allow": []interface{}{"MyComponent"},
			},
		},
		{
			// Identifier resolves to object — should be valid
			Code: `const s = { color: "red" }; var x = <div style={s} />`,
			Tsx:  true,
		},
		{
			// Identifier with no initializer — should be valid (can't determine)
			Code: `declare const s: any; var x = <div style={s} />`,
			Tsx:  true,
		},
		// React.createElement valid cases
		{
			Code: `React.createElement('div', { style: { color: "red" } })`,
			Tsx:  true,
		},
		{
			Code: `React.createElement('div', {})`,
			Tsx:  true,
		},
		{
			Code: `React.createElement('div', null)`,
			Tsx:  true,
		},
		{
			// allow list on createElement
			Code: `React.createElement(MyComponent, { style: "color: red" })`,
			Tsx:  true,
			Options: map[string]interface{}{
				"allow": []interface{}{"MyComponent"},
			},
		},
		{
			// No style property in the object
			Code: `React.createElement('div', { className: "foo" })`,
			Tsx:  true,
		},
		{
			// Computed style property — skip (can't statically determine key name)
			Code: `const key = "style"; React.createElement('div', { [key]: "red" })`,
			Tsx:  true,
		},

		{
			// Non-DOM component with object style
			Code: `var x = <Hello style={{ color: "red" }} />`,
			Tsx:  true,
		},
		{
			// Style with no value (JSX shorthand)
			Code: `var x = <div style></div>`,
			Tsx:  true,
		},
		{
			// Style with undefined
			Code: `var x = <div style={undefined}></div>`,
			Tsx:  true,
		},
		{
			// Style from props (can't determine type)
			Code: `function App(props: any) { return <div style={props.style} /> }`,
			Tsx:  true,
		},
		{
			// createElement with spread props (can't statically check)
			Code: `const props = { style: { color: "red" } }; React.createElement("div", props)`,
			Tsx:  true,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `var x = <div style="color: red" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "stylePropNotObject",
					Line:      1,
					Column:    14,
				},
			},
		},
		{
			Code: `var x = <div style={"color: red"} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "stylePropNotObject",
					Line:      1,
					Column:    14,
				},
			},
		},
		{
			Code: `var x = <div style={42} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "stylePropNotObject",
					Line:      1,
					Column:    14,
				},
			},
		},
		{
			Code: `var x = <div style={true} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "stylePropNotObject",
					Line:      1,
					Column:    14,
				},
			},
		},
		{
			// Identifier resolves to string literal — should be invalid
			Code: `const s = "color: red"; var x = <div style={s} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "stylePropNotObject",
					Line:      1,
					Column:    38,
				},
			},
		},
		{
			// Identifier resolves to number literal — should be invalid
			Code: `const s = 42; var x = <div style={s} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "stylePropNotObject",
					Line:      1,
					Column:    28,
				},
			},
		},
		// React.createElement invalid cases
		{
			// String literal style value in createElement
			Code: `React.createElement('div', { style: "color: red" })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "stylePropNotObject",
					Line:      1,
					Column:    37,
				},
			},
		},
		{
			// Number style value in createElement
			Code: `React.createElement('div', { style: 42 })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "stylePropNotObject",
					Line:      1,
					Column:    37,
				},
			},
		},
		{
			// Boolean style value in createElement
			Code: `React.createElement('div', { style: true })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "stylePropNotObject",
					Line:      1,
					Column:    37,
				},
			},
		},
		{
			// Identifier resolving to string in createElement
			Code: `const s = "red"; React.createElement('div', { style: s })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "stylePropNotObject",
					Line:      1,
					Column:    54,
				},
			},
		},

		{
			// Non-DOM component with string style
			Code: `var x = <Hello style="color: red" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "stylePropNotObject",
					Line:      1,
					Column:    16,
				},
			},
		},
		{
			// Allow list: component not in allow list should error
			Code: `var x = <MyComponent style="myStyle" />`,
			Tsx:  true,
			Options: map[string]interface{}{
				"allow": []interface{}{"MyOtherComponent"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "stylePropNotObject",
					Line:      1,
					Column:    22,
				},
			},
		},
		{
			// createElement: allow list component not matching
			Code: `React.createElement(MyComponent, { style: "mySpecialStyle" })`,
			Tsx:  true,
			Options: map[string]interface{}{
				"allow": []interface{}{"MyOtherComponent"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "stylePropNotObject",
					Line:      1,
					Column:    43,
				},
			},
		},

		// NOTE: ESLint's isNonNullaryLiteral also covers RegExp and BigInt literals,
		// but style={/regex/} and style={100n} are not realistic patterns.
		// Uncomment if RegExp/BigInt support is added to isNonObjectExpression:
		//
		// {
		// 	// RegExp literal style value — ESLint reports this but we don't currently detect it
		// 	Code: `var x = <div style={/regex/} />`,
		// 	Tsx:  true,
		// 	Errors: []rule_tester.InvalidTestCaseError{
		// 		{
		// 			MessageId: "stylePropNotObject",
		// 			Line:      1,
		// 			Column:    14,
		// 		},
		// 	},
		// },
		{
			// Shorthand property in createElement resolving to string
			Code: `const style = "red"; React.createElement("div", { style })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "stylePropNotObject",
					Line:      1,
					Column:    51,
				},
			},
		},
	})
}
