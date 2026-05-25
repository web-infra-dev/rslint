// TestJsxClosingTagLocationUpstream migrates the full valid/invalid suite from
// upstream packages/eslint-plugin/rules/jsx-closing-tag-location/
// jsx-closing-tag-location.test.ts 1:1. Upstream asserts only messageId on its
// invalid cases; line/column/endLine/endColumn here are computed from the exact
// source the case carries. rslint-specific lock-in cases live in
// jsx_closing_tag_location_extras_test.go.
package jsx_closing_tag_location

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxClosingTagLocationUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxClosingTagLocationRule, []rule_tester.ValidTestCase{
		// ---- default tag-aligned ----
		{Code: "\n    <App>\n      foo\n    </App>\n  ", Tsx: true},
		{Code: "\n    <App>foo</App>\n  ", Tsx: true},
		// ---- fragment (features: ['fragment']) ----
		{Code: "\n    <>\n      foo\n    </>\n  ", Tsx: true},
		{Code: "\n    <>foo</>\n  ", Tsx: true},
		// ---- line-aligned ----
		{Code: "\n    const foo = () => {\n      return <App>\n   bar</App>\n    }\n  ", Tsx: true, Options: []interface{}{"line-aligned"}},
		{Code: "\n    const foo = () => {\n      return <App>\n          bar</App>\n    }\n  ", Tsx: true},
		{Code: "\n    const foo = () => {\n      return <App>\n          bar\n      </App>\n    }\n  ", Tsx: true, Options: []interface{}{"line-aligned"}},
		{Code: "\n    const foo = <App>\n          bar\n    </App>\n  ", Tsx: true, Options: []interface{}{"line-aligned"}},
		{Code: "\n    const x = <App>\n          foo\n              </App>\n  ", Tsx: true},
		{Code: "\n    const foo =\n      <App>\n          bar\n      </App>\n  ", Tsx: true, Options: []interface{}{"line-aligned"}},
	}, []rule_tester.InvalidTestCase{
		// ---- default tag-aligned: closing tag alone on its line, wrong indent ----
		{
			Code:   "\n    <App>\n      foo\n      </App>\n  ",
			Tsx:    true,
			Output: []string{"\n    <App>\n      foo\n    </App>\n  "},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "matchIndent",
				Message:   "Expected closing tag to match indentation of opening.",
				Line:      4,
				Column:    7,
				EndLine:   4,
				EndColumn: 13,
			}},
		},
		// ---- default tag-aligned: closing tag shares line with content ----
		{
			Code:   "\n    <App>\n      foo</App>\n  ",
			Tsx:    true,
			Output: []string{"\n    <App>\n      foo\n    </App>\n  "},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "onOwnLine",
				Message:   "Closing tag of a multiline JSX expression must be on its own line.",
				Line:      3,
				Column:    10,
				EndLine:   3,
				EndColumn: 16,
			}},
		},
		// ---- fragment, alone on its line, wrong indent (features: ['fragment', 'no-ts-old']) ----
		{
			Code:   "\n    <>\n      foo\n      </>\n  ",
			Tsx:    true,
			Output: []string{"\n    <>\n      foo\n    </>\n  "},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "matchIndent",
				Line:      4,
				Column:    7,
				EndLine:   4,
				EndColumn: 10,
			}},
		},
		// ---- fragment, shares line with content (features: ['fragment', 'no-ts-old']) ----
		{
			Code:   "\n    <>\n      foo</>\n  ",
			Tsx:    true,
			Output: []string{"\n    <>\n      foo\n    </>\n  "},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "onOwnLine",
				Line:      3,
				Column:    10,
				EndLine:   3,
				EndColumn: 13,
			}},
		},
		// ---- line-aligned: closing tag shares line with content ----
		{
			Code:    "\n    const x = () => {\n      return <App>\n          foo</App>\n    }\n  ",
			Tsx:     true,
			Options: []interface{}{"line-aligned"},
			Output:  []string{"\n    const x = () => {\n      return <App>\n          foo\n      </App>\n    }\n  "},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "onOwnLine",
				Line:      4,
				Column:    14,
				EndLine:   4,
				EndColumn: 20,
			}},
		},
		// ---- line-aligned: closing tag alone on its line, over-indented ----
		{
			Code:    "\n    const x = <App>\n          foo\n              </App>\n  ",
			Tsx:     true,
			Options: []interface{}{"line-aligned"},
			Output:  []string{"\n    const x = <App>\n          foo\n    </App>\n  "},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "alignWithOpening",
				Message:   "Expected closing tag to be aligned with the line containing the opening tag",
				Line:      4,
				Column:    15,
				EndLine:   4,
				EndColumn: 21,
			}},
		},
	})
}
