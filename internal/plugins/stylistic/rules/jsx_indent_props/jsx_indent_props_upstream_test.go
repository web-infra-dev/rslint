// TestJsxIndentPropsUpstream migrates the full valid/invalid suite from upstream
// packages/eslint-plugin/rules/jsx-indent-props/jsx-indent-props.test.ts 1:1.
// Upstream asserts messageId + data on its invalid cases; the Message here
// reproduces upstream's exact rendered text (from the needed/type/gotten data),
// and Line/Column point at the reported attribute, computed from the exact
// source each case carries. rslint-specific edge-shape and branch lock-in cases
// live in jsx_indent_props_extras_test.go.
package jsx_indent_props

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxIndentPropsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxIndentPropsRule, []rule_tester.ValidTestCase{
		// ---- Default 4-space indent (no options) ----
		{Code: "\n        <App foo\n        />\n      ", Tsx: true},

		// ---- 2-space indent ----
		{Code: "\n        <App\n          foo\n        />\n      ", Tsx: true, Options: []interface{}{float64(2)}},

		// ---- Complex multi-element array literal (options: [2]) ----
		{Code: "\n        const Test = () => ([\n          (x\n            ? <div key=\"1\" />\n            : <div key=\"2\" />),\n          <div\n            key=\"3\"\n            align=\"left\"\n          />,\n          <div\n            key=\"4\"\n            align=\"left\"\n          />,\n        ]);\n      ", Tsx: true, Options: []interface{}{float64(2)}},

		// ---- 0-indent (props flush to column 0) ----
		{Code: "\n        <App\n        foo\n        />\n      ", Tsx: true, Options: []interface{}{float64(0)}},

		// ---- Negative indent (props two cols LEFT of `<App`) ----
		{Code: "\n          <App\n        foo\n          />\n      ", Tsx: true, Options: []interface{}{float64(-2)}},

		// ---- Tab indentation ----
		{Code: "\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t/>\n\t\t\t", Tsx: true, Options: []interface{}{"tab"}},

		// ---- 'first' option, no props ----
		{Code: "\n        <App/>\n      ", Tsx: true, Options: []interface{}{"first"}},

		// ---- 'first' option, props aligned with first prop's column ----
		{Code: "\n        <App aaa\n             b\n             cc\n        />\n      ", Tsx: true, Options: []interface{}{"first"}},
		{Code: "\n        <App   aaa\n               b\n               cc\n        />\n      ", Tsx: true, Options: []interface{}{"first"}},
		{Code: "\n        const test = <App aaa\n                          b\n                          cc\n                     />\n      ", Tsx: true, Options: []interface{}{"first"}},
		{Code: "\n        <App aaa x\n             b y\n             cc\n        />\n      ", Tsx: true, Options: []interface{}{"first"}},
		{Code: "\n        const test = <App aaa x\n                          b y\n                          cc\n                     />\n      ", Tsx: true, Options: []interface{}{"first"}},
		{Code: "\n        <App aaa\n             b\n        >\n            <Child c\n                   d/>\n        </App>\n      ", Tsx: true, Options: []interface{}{"first"}},
		{Code: "\n        <Fragment>\n          <App aaa\n               b\n               cc\n          />\n          <OtherApp a\n                    bbb\n                    c\n          />\n        </Fragment>\n      ", Tsx: true, Options: []interface{}{"first"}},

		// ---- 'first' with all props on new lines: each prop must match the
		// first prop's column. ----
		{Code: "\n        <App\n          a\n          b\n        />\n      ", Tsx: true, Options: []interface{}{"first"}},

		// ---- ignoreTernaryOperator: false (default), input already bumped. ----
		{Code: "\n        {this.props.ignoreTernaryOperatorFalse\n          ? <span\n              className=\"value\"\n              some={{aaa}}\n            />\n          : null}\n      ", Tsx: true, Options: []interface{}{map[string]interface{}{"indentMode": float64(2), "ignoreTernaryOperator": false}}},

		// ---- Function returning JSX in conditional, default indentMode=2. ----
		{Code: "\n        const F = () => {\n          const foo = true\n            ? <div id=\"id\">test</div>\n            : false;\n\n          return <div\n            id=\"id\"\n          >\n            test\n          </div>\n        }\n      ", Tsx: true, Options: []interface{}{map[string]interface{}{"indentMode": float64(2), "ignoreTernaryOperator": false}}},
		{Code: "\n        const F = () => {\n          const foo = true\n            ? <div id=\"id\">test</div>\n            : false;\n\n          return <div\n            id=\"id\"\n          >\n            test\n          </div>\n        }\n      ", Tsx: true, Options: []interface{}{map[string]interface{}{"indentMode": float64(2), "ignoreTernaryOperator": true}}},

		// ---- Tab indent with conditional + return JSX ----
		{Code: "\n\t\t\t\tconst F = () => {\n\t\t\t\t\tconst foo = true\n\t\t\t\t\t\t? <div id=\"id\">test</div>\n\t\t\t\t\t\t: false;\n\n\t\t\t\t\treturn <div\n\t\t\t\t\t\tid=\"id\"\n\t\t\t\t\t>\n\t\t\t\t\t\ttest\n\t\t\t\t\t</div>\n\t\t\t\t}\n", Tsx: true, Options: []interface{}{map[string]interface{}{"indentMode": "tab", "ignoreTernaryOperator": false}}},
		{Code: "\n\t\t\t\tconst F = () => {\n\t\t\t\t\tconst foo = true\n\t\t\t\t\t\t? <div id=\"id\">test</div>\n\t\t\t\t\t\t: false;\n\n\t\t\t\t\treturn <div\n\t\t\t\t\t\tid=\"id\"\n\t\t\t\t\t>\n\t\t\t\t\t\ttest\n\t\t\t\t\t</div>\n\t\t\t\t}\n", Tsx: true, Options: []interface{}{map[string]interface{}{"indentMode": "tab", "ignoreTernaryOperator": true}}},

		// ---- ignoreTernaryOperator: true — props inside ternary side do NOT
		// receive the extra bump. ----
		{Code: "\n        {this.props.ignoreTernaryOperatorTrue\n          ? <span\n            className=\"value\"\n            some={{aaa}}\n            />\n          : null}\n      ", Tsx: true, Options: []interface{}{map[string]interface{}{"indentMode": float64(2), "ignoreTernaryOperator": true}}},

		// ---- Realistic anchor element (object form with only indentMode). ----
		{Code: "\n        <a\n          role={'button'}\n          className={`navbar-burger ${open ? 'is-active' : ''}`}\n          href={'#'}\n          aria-label={'menu'}\n          aria-expanded={false}\n          onClick={openMenu}>\n          <span aria-hidden={'true'}/>\n          <span aria-hidden={'true'}/>\n          <span aria-hidden={'true'}/>\n        </a>\n      ", Tsx: true, Options: []interface{}{map[string]interface{}{"indentMode": float64(2)}}},

		// ---- Same realistic element under 'first' alignment ----
		{Code: "\n        <a role={'button'}\n           className={`navbar-burger ${open ? 'is-active' : ''}`}\n           href={'#'}\n           aria-label={'menu'}\n           aria-expanded={false}\n           onClick={openMenu}>\n          <span aria-hidden={'true'}/>\n          <span aria-hidden={'true'}/>\n          <span aria-hidden={'true'}/>\n        </a>\n      ", Tsx: true, Options: []interface{}{"first"}},
	}, []rule_tester.InvalidTestCase{
		// ---- Default 4-space — child at 10 must be 12. ----
		{
			Code:   "\n        <App\n          foo\n        />\n      ",
			Output: []string{"\n        <App\n            foo\n        />\n      "},
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 10.", Line: 3, Column: 11},
			},
		},

		// ---- 2-space — child at 12 must be 10. ----
		{
			Code:    "\n        <App\n            foo\n        />\n      ",
			Output:  []string{"\n        <App\n          foo\n        />\n      "},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 10 space characters but found 12.", Line: 3, Column: 13},
			},
		},

		// ---- Conditional with JSX in both branches — bump applies on each
		// ternary side. ----
		{
			Code:    "\n        const test = true\n          ? <span\n            attr=\"value\"\n            />\n          : <span\n            attr=\"otherValue\"\n            />\n      ",
			Output:  []string{"\n        const test = true\n          ? <span\n              attr=\"value\"\n            />\n          : <span\n              attr=\"otherValue\"\n            />\n      "},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 14 space characters but found 12.", Line: 4, Column: 13},
				{MessageId: "wrongIndent", Message: "Expected indentation of 14 space characters but found 12.", Line: 7, Column: 13},
			},
		},

		// ---- Alternate wrapped in its own parens — no ternary bump. ----
		{
			Code:    "\n        const test = true\n          ? <span attr=\"value\" />\n          : (\n            <span\n                attr=\"otherValue\"\n            />\n          )\n      ",
			Output:  []string{"\n        const test = true\n          ? <span attr=\"value\" />\n          : (\n            <span\n              attr=\"otherValue\"\n            />\n          )\n      "},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 14 space characters but found 16.", Line: 6, Column: 17},
			},
		},

		// ---- Ternary alternate: one prop. ----
		{
			Code:    "\n        {test.isLoading\n          ? <Value/>\n          : <OtherValue\n            some={aaa}/>\n        }\n      ",
			Output:  []string{"\n        {test.isLoading\n          ? <Value/>\n          : <OtherValue\n              some={aaa}/>\n        }\n      "},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 14 space characters but found 12.", Line: 5, Column: 13},
			},
		},

		// ---- Ternary alternate: two props, both bumped. ----
		{
			Code:    "\n        {test.isLoading\n          ? <Value/>\n          : <OtherValue\n            some={aaa}\n            other={bbb}/>\n        }\n      ",
			Output:  []string{"\n        {test.isLoading\n          ? <Value/>\n          : <OtherValue\n              some={aaa}\n              other={bbb}/>\n        }\n      "},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 14 space characters but found 12.", Line: 5, Column: 13},
				{MessageId: "wrongIndent", Message: "Expected indentation of 14 space characters but found 12.", Line: 6, Column: 13},
			},
		},

		// ---- Ternary consequent inside JSXExpressionContainer: bump applies. ----
		{
			Code:    "\n        {this.props.test\n          ? <span\n            className=\"value\"\n            some={{aaa}}\n            />\n          : null}\n      ",
			Output:  []string{"\n        {this.props.test\n          ? <span\n              className=\"value\"\n              some={{aaa}}\n            />\n          : null}\n      "},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 14 space characters but found 12.", Line: 4, Column: 13},
				{MessageId: "wrongIndent", Message: "Expected indentation of 14 space characters but found 12.", Line: 5, Column: 13},
			},
		},

		// ---- Tab option, prop has no leading tab — needed=1 (singular). ----
		{
			Code:    "\n        <App1\n            foo\n        />\n      ",
			Output:  []string{"\n        <App1\n\tfoo\n        />\n      "},
			Tsx:     true,
			Options: []interface{}{"tab"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 1 tab character but found 0.", Line: 3, Column: 13},
			},
		},

		// ---- Tab option, too many tabs — needed=5. ----
		{
			Code:    "\n\t\t\t\t<App\n\t\t\t\t\t\t\tfoo\n\t\t\t\t/>\n\t\t\t",
			Output:  []string{"\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t/>\n\t\t\t"},
			Tsx:     true,
			Options: []interface{}{"tab"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 5 tab characters but found 7.", Line: 3, Column: 8},
			},
		},

		// ---- 'first': prop at col 10 must match firstPropCol=13. ----
		{
			Code:    "\n        <App a\n          b\n        />\n      ",
			Output:  []string{"\n        <App a\n             b\n        />\n      "},
			Tsx:     true,
			Options: []interface{}{"first"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 13 space characters but found 10.", Line: 3, Column: 11},
			},
		},

		// ---- 'first': prop at col 11 must match firstPropCol=14. ----
		{
			Code:    "\n        <App  a\n           b\n        />\n      ",
			Output:  []string{"\n        <App  a\n              b\n        />\n      "},
			Tsx:     true,
			Options: []interface{}{"first"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 14 space characters but found 11.", Line: 3, Column: 12},
			},
		},

		// ---- 'first', first prop on new line: subsequent props match the
		// FIRST prop's column. ----
		{
			Code:    "\n        <App\n              a\n           b\n        />\n      ",
			Output:  []string{"\n        <App\n              a\n              b\n        />\n      "},
			Tsx:     true,
			Options: []interface{}{"first"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 14 space characters but found 11.", Line: 4, Column: 12},
			},
		},

		// ---- 'first', three props with two mis-aligned. ----
		{
			Code:    "\n        <App\n          a\n         b\n           c\n        />\n      ",
			Output:  []string{"\n        <App\n          a\n          b\n          c\n        />\n      "},
			Tsx:     true,
			Options: []interface{}{"first"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 10 space characters but found 9.", Line: 4, Column: 10},
				{MessageId: "wrongIndent", Message: "Expected indentation of 10 space characters but found 11.", Line: 5, Column: 12},
			},
		},

		// ---- ignoreTernaryOperator:false, return JSX one indent off. ----
		{
			Code:    "\n        const F = () => {\n          const foo = true\n            ? <div id=\"id\">test</div>\n            : false;\n\n          return <div\n              id=\"id\"\n          >\n            test\n          </div>\n        }\n      ",
			Output:  []string{"\n        const F = () => {\n          const foo = true\n            ? <div id=\"id\">test</div>\n            : false;\n\n          return <div\n            id=\"id\"\n          >\n            test\n          </div>\n        }\n      "},
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"indentMode": float64(2), "ignoreTernaryOperator": false}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 14.", Line: 8, Column: 15},
			},
		},

		// ---- Same shape, ignoreTernaryOperator:true (return JSX not in a
		// ternary, so behaviour is identical). ----
		{
			Code:    "\n        const F = () => {\n          const foo = true\n            ? <div id=\"id\">test</div>\n            : false;\n\n          return <div\n              id=\"id\"\n          >\n            test\n          </div>\n        }\n      ",
			Output:  []string{"\n        const F = () => {\n          const foo = true\n            ? <div id=\"id\">test</div>\n            : false;\n\n          return <div\n            id=\"id\"\n          >\n            test\n          </div>\n        }\n      "},
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"indentMode": float64(2), "ignoreTernaryOperator": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 14.", Line: 8, Column: 15},
			},
		},

		// ---- Tab indent + return JSX off by one tab. ----
		{
			Code:    "\n\t\t\t\tconst F = () => {\n\t\t\t\t\tconst foo = true\n\t\t\t\t\t\t? <div id=\"id\">test</div>\n\t\t\t\t\t\t: false;\n\n\t\t\t\t\treturn <div\n\t\t\t\t\t\t\tid=\"id\"\n\t\t\t\t\t>\n\t\t\t\t\t\ttest\n\t\t\t\t\t</div>\n\t\t\t\t}\n",
			Output:  []string{"\n\t\t\t\tconst F = () => {\n\t\t\t\t\tconst foo = true\n\t\t\t\t\t\t? <div id=\"id\">test</div>\n\t\t\t\t\t\t: false;\n\n\t\t\t\t\treturn <div\n\t\t\t\t\t\tid=\"id\"\n\t\t\t\t\t>\n\t\t\t\t\t\ttest\n\t\t\t\t\t</div>\n\t\t\t\t}\n"},
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"indentMode": "tab", "ignoreTernaryOperator": false}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 6 tab characters but found 7.", Line: 8, Column: 8},
			},
		},
		{
			Code:    "\n\t\t\t\tconst F = () => {\n\t\t\t\t\tconst foo = true\n\t\t\t\t\t\t? <div id=\"id\">test</div>\n\t\t\t\t\t\t: false;\n\n\t\t\t\t\treturn <div\n\t\t\t\t\t\t\tid=\"id\"\n\t\t\t\t\t>\n\t\t\t\t\t\ttest\n\t\t\t\t\t</div>\n\t\t\t\t}\n",
			Output:  []string{"\n\t\t\t\tconst F = () => {\n\t\t\t\t\tconst foo = true\n\t\t\t\t\t\t? <div id=\"id\">test</div>\n\t\t\t\t\t\t: false;\n\n\t\t\t\t\treturn <div\n\t\t\t\t\t\tid=\"id\"\n\t\t\t\t\t>\n\t\t\t\t\t\ttest\n\t\t\t\t\t</div>\n\t\t\t\t}\n"},
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"indentMode": "tab", "ignoreTernaryOperator": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 6 tab characters but found 7.", Line: 8, Column: 8},
			},
		},
	})
}
