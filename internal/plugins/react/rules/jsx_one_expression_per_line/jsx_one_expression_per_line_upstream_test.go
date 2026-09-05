// TestJsxOneExpressionPerLineUpstream is a case-for-case port of the full
// eslint-plugin-react v7.37.5 test suite. The source code, expected output,
// message data, and upstream case metadata are preserved verbatim; Tsx is
// the local RuleTester file-mode setting required for these JSX cases.
package jsx_one_expression_per_line

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxOneExpressionPerLineUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxOneExpressionPerLineRule, []rule_tester.ValidTestCase{
		{
			Code: "<App />",
			Tsx:  true,
		},
		{
			Code: "\n\t\t\t\t<AllTabs>\n\t\t\t\t\tFail\n\t\t\t\t</AllTabs>\n      ",
			Tsx:  true,
		},
		{
			Code: "\n\t\t\t\t<TagsWithTabs>\n          Fail\n\t\t\t\t</TagsWithTabs>\n      ",
			Tsx:  true,
		},
		{
			Code: "\n        <ClosedTagWithTabs>\n          Fail\n\t\t\t\t</ClosedTagWithTabs>\n      ",
			Tsx:  true,
		},
		{
			Code: "\n\t\t\t\t<OpenTagWithTabs>\n          OK\n        </OpenTagWithTabs>\n      ",
			Tsx:  true,
		},
		{
			Code: "\n        <TextWithTabs>\n\t\t\t\t\t\tOK\n        </TextWithTabs>\n      ",
			Tsx:  true,
		},
		{
			Code: "\n        <AllSpaces>\n          OK\n        </AllSpaces>\n      ",
			Tsx:  true,
		},
		{
			Code: "<App></App>",
			Tsx:  true,
		},
		{
			Code: "<App foo=\"bar\" />",
			Tsx:  true,
		},
		{
			Code: "\n        <App>\n          <Foo />\n        </App>\n      ",
			Tsx:  true,
		},
		{
			Code: "\n        <App>\n          <Foo />\n          <Bar />\n        </App>\n      ",
			Tsx:  true,
		},
		{
			Code: "\n        <App>\n          <Foo></Foo>\n        </App>\n      ",
			Tsx:  true,
		},
		{
			Code: "\n        <App>\n          foo bar baz  whatever\n        </App>\n      ",
			Tsx:  true,
		},
		{
			Code: "\n        <App>\n          <Foo>\n          </Foo>\n        </App>\n      ",
			Tsx:  true,
		},
		{
			Code: "\n        <App\n          foo=\"bar\"\n        >\n        <Foo />\n        </App>\n      ",
			Tsx:  true,
		},
		{
			Code: "\n        <\n        App\n        >\n          <\n            Foo\n          />\n        </\n        App\n        >\n      ",
			Tsx:  true,
		},
		{
			Code:    "<App>foo</App>",
			Tsx:     true,
			Options: map[string]any{"allow": "literal"},
		},
		{
			Code:    "<App>123</App>",
			Tsx:     true,
			Options: map[string]any{"allow": "literal"},
		},
		{
			Code:    "<App>foo</App>",
			Tsx:     true,
			Options: map[string]any{"allow": "single-child"},
		},
		{
			Code:    "<App>{\"foo\"}</App>",
			Tsx:     true,
			Options: map[string]any{"allow": "single-child"},
		},
		{
			Code:    "<App>123</App>",
			Tsx:     true,
			Options: map[string]any{"allow": "non-jsx"},
		},
		{
			Code:    "<App>foo</App>",
			Tsx:     true,
			Options: map[string]any{"allow": "non-jsx"},
		},
		{
			Code:    "<App>{\"foo\"}</App>",
			Tsx:     true,
			Options: map[string]any{"allow": "non-jsx"},
		},
		{
			Code:    "<App>{<Bar />}</App>",
			Tsx:     true,
			Options: map[string]any{"allow": "non-jsx"},
		},
		{
			Code:    "<App>{foo && <Bar />}</App>",
			Tsx:     true,
			Options: map[string]any{"allow": "single-child"},
		},
		{
			Code:    "<App><Foo /></App>",
			Tsx:     true,
			Options: map[string]any{"allow": "single-child"},
		},
		{
			Code: "<></>",
			Tsx:  true,
		},
		{
			Code: "\n        <>\n          <Foo />\n        </>\n      ",
			Tsx:  true,
		},
		{
			Code: "\n        <>\n          <Foo />\n          <Bar />\n        </>\n      ",
			Tsx:  true,
		},
		{
			Code:    "<App>Hello {name}</App>",
			Tsx:     true,
			Options: map[string]any{"allow": "non-jsx"},
		},
		{
			Code:    "\n        <App>\n          Hello {name} there!\n        </App>",
			Tsx:     true,
			Options: map[string]any{"allow": "non-jsx"},
		},
		{
			Code:    "\n        <App>\n          Hello {<Bar />} there!\n        </App>",
			Tsx:     true,
			Options: map[string]any{"allow": "non-jsx"},
		},
		{
			Code:    "\n        <App>\n          Hello {(<Bar />)} there!\n        </App>",
			Tsx:     true,
			Options: map[string]any{"allow": "non-jsx"},
		},
		{
			Code:    "\n        <App>\n          Hello {(() => <Bar />)()} there!\n        </App>",
			Tsx:     true,
			Options: map[string]any{"allow": "non-jsx"},
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code:   "\n        <App>{\"foo\"}</App>\n      ",
			Tsx:    true,
			Output: []string{"\n        <App>\n{\"foo\"}\n</App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`{\"foo\"}` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <App>foo</App>\n      ",
			Tsx:    true,
			Output: []string{"\n        <App>\nfoo\n</App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`foo` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <div>\n          foo {\"bar\"}\n        </div>\n      ",
			Tsx:    true,
			Output: []string{"\n        <div>\n          foo \n{' '}\n{\"bar\"}\n        </div>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`{\"bar\"}` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <div>\n          {\"foo\"} bar\n        </div>\n      ",
			Tsx:    true,
			Output: []string{"\n        <div>\n          {\"foo\"}\n{' '}\nbar\n</div>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "` bar        ` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <App>\n          <Foo /><Bar />\n        </App>\n      ",
			Tsx:    true,
			Output: []string{"\n        <App>\n          <Foo />\n<Bar />\n        </App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`Bar` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <div>\n          <span />foo\n        </div>\n      ",
			Tsx:    true,
			Output: []string{"\n        <div>\n          <span />\nfoo\n</div>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`foo        ` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <div>\n          <span />{\"foo\"}\n        </div>\n      ",
			Tsx:    true,
			Output: []string{"\n        <div>\n          <span />\n{\"foo\"}\n        </div>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`{\"foo\"}` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <div>\n          {\"foo\"} { I18n.t('baz') }\n        </div>\n      ",
			Tsx:    true,
			Output: []string{"\n        <div>\n          {\"foo\"} \n{' '}\n{ I18n.t('baz') }\n        </div>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`{ I18n.t('baz') }` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <Text style={styles.foo}>{ bar } <Text/> { I18n.t('baz') }</Text>\n      ",
			Tsx:    true,
			Output: []string{"\n        <Text style={styles.foo}>\n{ bar } \n{' '}\n<Text/> \n{' '}\n{ I18n.t('baz') }\n</Text>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`{ bar }` must be placed on a new line"},
				{MessageId: "moveToNewLine", Message: "`Text` must be placed on a new line"},
				{MessageId: "moveToNewLine", Message: "`{ I18n.t('baz') }` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <Text style={styles.foo}> <Bar/> <Baz/></Text>\n      ",
			Tsx:    true,
			Output: []string{"\n        <Text style={styles.foo}> \n{' '}\n<Bar/> \n{' '}\n<Baz/>\n</Text>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`Bar` must be placed on a new line"},
				{MessageId: "moveToNewLine", Message: "`Baz` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <Text style={styles.foo}> <Bar/> <Baz/> <Bunk/> <Bruno/> </Text>\n      ",
			Tsx:    true,
			Output: []string{"\n        <Text style={styles.foo}> \n{' '}\n<Bar/> \n{' '}\n<Baz/> \n{' '}\n<Bunk/> \n{' '}\n<Bruno/>\n{' '}\n </Text>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`Bar` must be placed on a new line"},
				{MessageId: "moveToNewLine", Message: "`Baz` must be placed on a new line"},
				{MessageId: "moveToNewLine", Message: "`Bunk` must be placed on a new line"},
				{MessageId: "moveToNewLine", Message: "`Bruno` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <Text style={styles.foo}> <Bar /></Text>\n      ",
			Tsx:    true,
			Output: []string{"\n        <Text style={styles.foo}> \n{' '}\n<Bar />\n</Text>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`Bar` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <Text style={styles.foo}> <Bar />\n        </Text>\n      ",
			Tsx:    true,
			Output: []string{"\n        <Text style={styles.foo}> \n{' '}\n<Bar />\n        </Text>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`Bar` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <Text style={styles.foo}>\n          <Bar /> <Baz />\n        </Text>\n      ",
			Tsx:    true,
			Output: []string{"\n        <Text style={styles.foo}>\n          <Bar /> \n{' '}\n<Baz />\n        </Text>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`Baz` must be placed on a new line"},
			},
		},
		{
			Code:    "\n        <Text style={styles.foo}>\n          <Bar /> <Baz />\n        </Text>\n      ",
			Tsx:     true,
			Options: map[string]any{"allow": "non-jsx"},
			Output:  []string{"\n        <Text style={styles.foo}>\n          <Bar /> \n{' '}\n<Baz />\n        </Text>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`Baz` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <Text style={styles.foo}>\n          { bar } { I18n.t('baz') }\n        </Text>\n      ",
			Tsx:    true,
			Output: []string{"\n        <Text style={styles.foo}>\n          { bar } \n{' '}\n{ I18n.t('baz') }\n        </Text>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`{ I18n.t('baz') }` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <div>\n          foo<input />\n        </div>\n      ",
			Tsx:    true,
			Output: []string{"\n        <div>\n          foo\n<input />\n        </div>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`input` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <div>\n          {\"foo\"}<span />\n        </div>\n      ",
			Tsx:    true,
			Output: []string{"\n        <div>\n          {\"foo\"}\n<span />\n        </div>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`span` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <div>\n          foo <input />\n        </div>\n      ",
			Tsx:    true,
			Output: []string{"\n        <div>\n          foo \n{' '}\n<input />\n        </div>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`input` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <div>\n          <input /> foo\n        </div>\n      ",
			Tsx:    true,
			Output: []string{"\n        <div>\n          <input />\n{' '}\nfoo\n</div>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "` foo        ` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <div>\n          <span /> <input />\n        </div>\n      ",
			Tsx:    true,
			Output: []string{"\n        <div>\n          <span /> \n{' '}\n<input />\n        </div>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`input` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <div>\n          <span />\n        {' '}<input />\n        </div>\n      ",
			Tsx:    true,
			Output: []string{"\n        <div>\n          <span />\n        {' '}\n<input />\n        </div>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`input` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <div>\n          {\"foo\"} <input />\n        </div>\n      ",
			Tsx:    true,
			Output: []string{"\n        <div>\n          {\"foo\"} \n{' '}\n<input />\n        </div>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`input` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <div>\n          <input /> {\"foo\"}\n        </div>\n      ",
			Tsx:    true,
			Output: []string{"\n        <div>\n          <input /> \n{' '}\n{\"foo\"}\n        </div>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`{\"foo\"}` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <App>\n          <Foo></Foo><Bar></Bar>\n        </App>\n      ",
			Tsx:    true,
			Output: []string{"\n        <App>\n          <Foo></Foo>\n<Bar></Bar>\n        </App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`Bar` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <App>\n        <Foo></Foo></App>\n      ",
			Tsx:    true,
			Output: []string{"\n        <App>\n        <Foo></Foo>\n</App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`Foo` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <App><Foo />\n        </App>\n      ",
			Tsx:    true,
			Output: []string{"\n        <App>\n<Foo />\n        </App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`Foo` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <App>\n        <Foo/></App>\n      ",
			Tsx:    true,
			Output: []string{"\n        <App>\n        <Foo/>\n</App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`Foo` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <App><Foo\n        />\n        </App>\n      ",
			Tsx:    true,
			Output: []string{"\n        <App>\n<Foo\n        />\n        </App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`Foo` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <App\n        >\n        <Foo /></App>\n      ",
			Tsx:    true,
			Output: []string{"\n        <App\n        >\n        <Foo />\n</App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`Foo` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <App\n        >\n        <Foo\n        /></App>\n      ",
			Tsx:    true,
			Output: []string{"\n        <App\n        >\n        <Foo\n        />\n</App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`Foo` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <App\n        ><Foo />\n        </App>\n      ",
			Tsx:    true,
			Output: []string{"\n        <App\n        >\n<Foo />\n        </App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`Foo` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <App>\n          <Foo></Foo\n        ></App>\n      ",
			Tsx:    true,
			Output: []string{"\n        <App>\n          <Foo></Foo\n        >\n</App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`Foo` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <App>\n          <Foo></\n        Foo></App>\n      ",
			Tsx:    true,
			Output: []string{"\n        <App>\n          <Foo></\n        Foo>\n</App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`Foo` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <App>\n          <Foo></\n        Foo><Bar />\n        </App>\n      ",
			Tsx:    true,
			Output: []string{"\n        <App>\n          <Foo></\n        Foo>\n<Bar />\n        </App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`Bar` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <App>\n          <Foo>\n            <Bar /></Foo>\n        </App>\n      ",
			Tsx:    true,
			Output: []string{"\n        <App>\n          <Foo>\n            <Bar />\n</Foo>\n        </App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`Bar` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <App>\n          <Foo>\n            <Bar> baz </Bar>\n          </Foo>\n        </App>\n      ",
			Tsx:    true,
			Output: []string{"\n        <App>\n          <Foo>\n            <Bar>\n{' '}\nbaz\n{' '}\n</Bar>\n          </Foo>\n        </App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "` baz ` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <App>\n          foo {\"bar\"} baz\n        </App>\n      ",
			Tsx:    true,
			Skip:   true, // ESLint's single-pass output differs from rslint's multipass fixer.
			Output: []string{"\n        <App>\n          foo \n{' '}\n{\"bar\"} baz\n        </App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`{\"bar\"}` must be placed on a new line"},
				{MessageId: "moveToNewLine", Message: "` baz        ` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <App>\n          foo {\"bar\"}\n        </App>\n      ",
			Tsx:    true,
			Output: []string{"\n        <App>\n          foo \n{' '}\n{\"bar\"}\n        </App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`{\"bar\"}` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <App>\n          foo\n        {' '}\n        {\"bar\"} baz\n        </App>\n      ",
			Tsx:    true,
			Output: []string{"\n        <App>\n          foo\n        {' '}\n        {\"bar\"}\n{' '}\nbaz\n</App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "` baz        ` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <App>\n\n          foo {\"bar\"} baz\n\n        </App>\n      ",
			Tsx:    true,
			Skip:   true, // ESLint's single-pass output differs from rslint's multipass fixer.
			Output: []string{"\n        <App>\n\n          foo \n{' '}\n{\"bar\"} baz\n\n        </App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`{\"bar\"}` must be placed on a new line"},
				{MessageId: "moveToNewLine", Message: "` baz        ` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <App>\n\n          foo\n        {' '}\n        {\"bar\"} baz\n\n        </App>\n      ",
			Tsx:    true,
			Output: []string{"\n        <App>\n\n          foo\n        {' '}\n        {\"bar\"}\n{' '}\nbaz\n\n</App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "` baz        ` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <App>{\n          foo\n        }</App>\n      ",
			Tsx:    true,
			Output: []string{"\n        <App>\n{\n          foo\n        }\n</App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`{          foo        }` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <App> {\n          foo\n        } </App>\n      ",
			Tsx:    true,
			Output: []string{"\n        <App> \n{' '}\n{\n          foo\n        }\n{' '}\n </App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`{          foo        }` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <App>\n        {' '}\n        {\n          foo\n        } </App>\n      ",
			Tsx:    true,
			Output: []string{"\n        <App>\n        {' '}\n        {\n          foo\n        }\n{' '}\n </App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`{          foo        }` must be placed on a new line"},
			},
		},
		{
			Code:    "\n        <App><Foo /></App>\n      ",
			Tsx:     true,
			Options: map[string]any{"allow": "none"},
			Output:  []string{"\n        <App>\n<Foo />\n</App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`Foo` must be placed on a new line"},
			},
		},
		{
			Code:    "\n        <App>foo</App>\n      ",
			Tsx:     true,
			Options: map[string]any{"allow": "none"},
			Output:  []string{"\n        <App>\nfoo\n</App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`foo` must be placed on a new line"},
			},
		},
		{
			Code:    "\n        <App>{\"foo\"}</App>\n      ",
			Tsx:     true,
			Options: map[string]any{"allow": "none"},
			Output:  []string{"\n        <App>\n{\"foo\"}\n</App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`{\"foo\"}` must be placed on a new line"},
			},
		},
		{
			Code:    "\n        <App>foo\n        </App>\n      ",
			Tsx:     true,
			Options: map[string]any{"allow": "literal"},
			Output:  []string{"\n        <App>\nfoo\n</App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`foo        ` must be placed on a new line"},
			},
		},
		{
			Code:    "\n        <App><Foo /></App>\n      ",
			Tsx:     true,
			Options: map[string]any{"allow": "literal"},
			Output:  []string{"\n        <App>\n<Foo />\n</App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`Foo` must be placed on a new line"},
			},
		},
		{
			Code:    "\n        <App><Foo /></App>\n      ",
			Tsx:     true,
			Options: map[string]any{"allow": "non-jsx"},
			Output:  []string{"\n        <App>\n<Foo />\n</App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`Foo` must be placed on a new line"},
			},
		},
		{
			Code:    "\n        <App\n          foo=\"1\"\n          bar=\"2\"\n        >baz</App>\n      ",
			Tsx:     true,
			Options: map[string]any{"allow": "literal"},
			Output:  []string{"\n        <App\n          foo=\"1\"\n          bar=\"2\"\n        >\nbaz\n</App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`baz` must be placed on a new line"},
			},
		},
		{
			Code:    "\n        <App>foo\n        bar\n        </App>\n      ",
			Tsx:     true,
			Options: map[string]any{"allow": "literal"},
			Output:  []string{"\n        <App>\nfoo\n        bar\n</App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`foo        bar        ` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <>{\"foo\"}</>\n      ",
			Tsx:    true,
			Output: []string{"\n        <>\n{\"foo\"}\n</>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`{\"foo\"}` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <App>\n          <Foo /><></>\n        </App>\n      ",
			Tsx:    true,
			Output: []string{"\n        <App>\n          <Foo />\n<></>\n        </App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`<></>` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <\n        ><Foo />\n        </>\n      ",
			Tsx:    true,
			Output: []string{"\n        <\n        >\n<Foo />\n        </>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`Foo` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <div>\n        <MyComponent>a</MyComponent>\n        <MyOther>{a}</MyOther>\n        </div>\n      ",
			Tsx:    true,
			Output: []string{"\n        <div>\n        <MyComponent>\na\n</MyComponent>\n        <MyOther>\n{a}\n</MyOther>\n        </div>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`a` must be placed on a new line"},
				{MessageId: "moveToNewLine", Message: "`{a}` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        const IndexPage = () => (\n          <h1>{\"Hi people\"}<button/></h1>\n        );\n      ",
			Tsx:    true,
			Skip:   true, // ESLint's single-pass output differs from rslint's multipass fixer.
			Output: []string{"\n        const IndexPage = () => (\n          <h1>\n{\"Hi people\"}<button/></h1>\n        );\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`{\"Hi people\"}` must be placed on a new line"},
				{MessageId: "moveToNewLine", Message: "`button` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        const IndexPage = () => (\n          <h1>\n{\"Hi people\"}<button/></h1>\n        );\n      ",
			Tsx:    true,
			Output: []string{"\n        const IndexPage = () => (\n          <h1>\n{\"Hi people\"}\n<button/>\n</h1>\n        );\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`button` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <Layout>\n        <p>Welcome to your new Gatsby site.</p>\n        <p>Now go build something great.</p>\n        <h1>Hi people<button/></h1>\n        </Layout>\n      ",
			Tsx:    true,
			Skip:   true, // ESLint's single-pass output differs from rslint's multipass fixer.
			Output: []string{"\n        <Layout>\n        <p>\nWelcome to your new Gatsby site.\n</p>\n        <p>\nNow go build something great.\n</p>\n        <h1>\nHi people<button/></h1>\n        </Layout>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`Welcome to your new Gatsby site.` must be placed on a new line"},
				{MessageId: "moveToNewLine", Message: "`Now go build something great.` must be placed on a new line"},
				{MessageId: "moveToNewLine", Message: "`Hi people` must be placed on a new line"},
				{MessageId: "moveToNewLine", Message: "`button` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <Layout>\n        <p>\nWelcome to your new Gatsby site.\n</p>\n        <p>\nNow go build something great.\n</p>\n        <h1>\nHi people<button/></h1>\n        </Layout>\n      ",
			Tsx:    true,
			Output: []string{"\n        <Layout>\n        <p>\nWelcome to your new Gatsby site.\n</p>\n        <p>\nNow go build something great.\n</p>\n        <h1>\nHi people\n<button/>\n</h1>\n        </Layout>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`button` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <Layout>\n          <div style={{ maxWidth: `300px`, marginBottom: `1.45rem` }}>\n            <Image />\n          </div><Link to=\"/page-2/\">Go to page 2</Link>\n        </Layout>\n      ",
			Tsx:    true,
			Skip:   true, // ESLint's single-pass output differs from rslint's multipass fixer.
			Output: []string{"\n        <Layout>\n          <div style={{ maxWidth: `300px`, marginBottom: `1.45rem` }}>\n            <Image />\n          </div>\n<Link to=\"/page-2/\">Go to page 2</Link>\n        </Layout>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`Link` must be placed on a new line"},
				{MessageId: "moveToNewLine", Message: "`Go to page 2` must be placed on a new line"},
			},
		},
		{
			Code:   "\n        <Layout>\n          <div style={{ maxWidth: `300px`, marginBottom: `1.45rem` }}>\n            <Image />\n          </div>\n<Link to=\"/page-2/\">Go to page 2</Link>\n        </Layout>\n      ",
			Tsx:    true,
			Output: []string{"\n        <Layout>\n          <div style={{ maxWidth: `300px`, marginBottom: `1.45rem` }}>\n            <Image />\n          </div>\n<Link to=\"/page-2/\">\nGo to page 2\n</Link>\n        </Layout>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "moveToNewLine", Message: "`Go to page 2` must be placed on a new line"},
			},
		},
	})
}
