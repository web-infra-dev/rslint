// TestJsxOneExpressionPerLineExtras locks in branches and edge shapes that
// the upstream test suite does not fully exercise. The upstream migration
// cases live in the sibling jsx_one_expression_per_line_upstream_test.go file.
package jsx_one_expression_per_line

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxOneExpressionPerLineExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxOneExpressionPerLineRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: parenthesized JSX expression ----
		{Code: "const value = (<App>\n  <Foo />\n</App>);", Tsx: true},
		// ---- Dimension 4: nested element and fragment boundaries ----
		{Code: "const value = <App>\n  <Foo>\n    <Bar />\n  </Foo>\n</App>;", Tsx: true},
		{Code: "const value = <>\n  <Foo />\n  <Bar />\n</>;", Tsx: true},
		// ---- Dimension 4: self-closing child ----
		{Code: "<App>\n  <Foo />\n</App>", Tsx: true},
		// ---- Dimension 4: empty children and whitespace-only text ----
		{Code: `<App></App>`, Tsx: true},
		{Code: "<App>\n  \t\n</App>", Tsx: true},
		// ---- Dimension 4: member and namespaced tag names ----
		{Code: "<App>\n  <Foo.Bar />\n</App>", Tsx: true},
		{Code: "<App>\n  <svg:path />\n</App>", Tsx: true},
		// ---- Dimension 4: multiline attributes do not become children ----
		{Code: "<App\n  foo=\"bar\"\n>\n  <Foo />\n</App>", Tsx: true},
		// ---- Dimension 4: async / TypeScript-containing expression child ----
		{Code: `<App>{(value as string)}</App>`, Tsx: true, Options: map[string]any{"allow": "single-child"}},
		// ---- Real-user: issue #1835, text after a component ----
		{Code: "<div>\n  <MyComponent>\n    a\n  </MyComponent>\n  <MyOther>\n    {a}\n  </MyOther>\n</div>", Tsx: true},
		// ---- Real-user: issue #1893, CRLF line endings ----
		{Code: "<div>\r\n  <Foo />\r\n</div>", Tsx: true},
		// ---- Real-user: issue #2318, Gatsby-style text children ----
		{Code: "<Layout>\n  <p>\n    Welcome to your new Gatsby site.\n  </p>\n  <p>\n    Now go build something great.\n  </p>\n</Layout>", Tsx: true},
		// Locks in upstream handleJSX() arm 2: non-jsx returns when no direct JSX child exists.
		{Code: `<App>text {value}</App>`, Tsx: true, Options: map[string]any{"allow": "non-jsx"}},
	}, []rule_tester.InvalidTestCase{
		{Code: `<App><Foo.Bar /></App>`, Tsx: true, Output: []string{"<App>\n<Foo.Bar />\n</App>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine", Message: "`Foo.Bar` must be placed on a new line"},
		}},
		{Code: `<App><svg:path /></App>`, Tsx: true, Output: []string{"<App>\n<svg:path />\n</App>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine", Message: "`svg:path` must be placed on a new line"},
		}},
		// Locks in upstream handleJSX() arm 1: empty children return without a report.
		{Code: `<App><Foo />text</App>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine", Message: "`Foo` must be placed on a new line", Line: 1, Column: 6, EndLine: 1, EndColumn: 13},
			{MessageId: "moveToNewLine", Message: "`text` must be placed on a new line", Line: 1, Column: 13, EndLine: 1, EndColumn: 17},
		}, Output: []string{"<App>\n<Foo />\ntext\n</App>"}},
		// Locks in upstream handleJSX() arm 3: a direct JSX child keeps the rule active in non-jsx mode.
		{Code: `<App>text <Foo /></App>`, Tsx: true, Options: map[string]any{"allow": "non-jsx"}, Output: []string{"<App>\ntext\n{' '}\n<Foo />\n</App>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine", Message: "`text ` must be placed on a new line", Line: 1, Column: 6, EndLine: 1, EndColumn: 11},
			{MessageId: "moveToNewLine", Message: "`Foo` must be placed on a new line", Line: 1, Column: 11, EndLine: 1, EndColumn: 18},
		}},
		// Locks in upstream single-child allow arms: literal allows text, not an expression container.
		{Code: `<App>{value}</App>`, Tsx: true, Options: map[string]any{"allow": "literal"}, Output: []string{"<App>\n{value}\n</App>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine", Message: "`{value}` must be placed on a new line", Line: 1, Column: 6, EndLine: 1, EndColumn: 13},
		}},
		// Locks in upstream grouping arm: a multiline text child is represented on both endpoint lines.
		{Code: "<App>\n  foo <Bar />\n</App>", Tsx: true, Output: []string{"<App>\n  foo \n{' '}\n<Bar />\n</App>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine", Message: "`Bar` must be placed on a new line", Line: 2, Column: 7, EndLine: 2, EndColumn: 14},
		}},
		// ---- Dimension 4: a multiline JSX child gets one report per conflicting sibling relation ----
		{Code: "<App>\n  <Foo>\n    text\n  </Foo><Bar />\n</App>", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine", Line: 4, Column: 9, EndLine: 4, EndColumn: 16},
		}, Output: []string{"<App>\n  <Foo>\n    text\n  </Foo>\n<Bar />\n</App>"}},
	})
}
