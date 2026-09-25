package jsx_tag_spacing

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Upstream: https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/tests/lib/rules/jsx-tag-spacing.js
// Documentation: https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/jsx-tag-spacing.md
// Parser matrix duplicates are represented once. Ranges and messages are checked
// against ESLint 8.57.1; start-only locations use a zero-width native range.

func TestJsxTagSpacingUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxTagSpacingRule, []rule_tester.ValidTestCase{
		// Upstream valid 1.
		{Code: "<App />", Tsx: true},
		// Upstream valid 2.
		{Code: "<App />", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "always", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Upstream valid 3.
		{Code: "<App foo />", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "always", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Upstream valid 4.
		{Code: "<App foo={bar} />", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "always", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Upstream valid 5.
		{Code: "<App {...props} />", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "always", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Upstream valid 6.
		{Code: "<App></App>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "always", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Upstream valid 7.
		{Code: "\n        <App\n          foo={bar}\n        />\n      ", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "always", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Upstream valid 8.
		{Code: "<App/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "never", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Upstream valid 9.
		{Code: "<App />", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "proportional-always", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Upstream valid 10.
		{Code: "<App foo />", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "proportional-always", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Upstream valid 11.
		{Code: "\n        <App\n          foo={bar}\n          blat\n        >\n          hello\n        </App>\n      ", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "proportional-always"}}},
		// Upstream valid 12.
		{Code: "\n        <App foo={bar}>\n          hello\n        </App>\n      ", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "proportional-always"}}},
		// Upstream valid 13.
		{Code: "\n        <App\n          foo={bar}\n        />\n      ", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "proportional-always", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Upstream valid 14.
		{Code: "<App foo/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "never", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Upstream valid 15.
		{Code: "<App foo={bar}/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "never", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Upstream valid 16.
		{Code: "<App {...props}/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "never", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Upstream valid 17.
		{Code: "<App></App>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "never", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Upstream valid 18.
		{Code: "\n        <App\n          foo={bar}\n        />\n      ", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "never", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Upstream valid 19.
		{Code: "<App/>;", Tsx: true, Options: []any{map[string]any{"closingSlash": "never", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Upstream valid 20.
		{Code: "<App />;", Tsx: true, Options: []any{map[string]any{"closingSlash": "never", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Upstream valid 21.
		{Code: "<div className=\"bar\"></div>;", Tsx: true, Options: []any{map[string]any{"closingSlash": "never", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Upstream valid 22.
		{Code: "<div className=\"bar\"></ div>;", Tsx: true, Options: []any{map[string]any{"closingSlash": "never", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Upstream valid 23.
		{Code: "<App prop=\"foo\">< /App>", Tsx: true, Options: []any{map[string]any{"closingSlash": "always", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "allow"}}, Skip: true}, // tsgo does not parse whitespace between `<` and `/` in closing tags.

		// Upstream valid 24.
		{Code: "<p/ >", Tsx: true, Options: []any{map[string]any{"closingSlash": "always", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Upstream valid 25.
		{Code: "<App/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "never", "beforeClosing": "allow"}}},
		// Upstream valid 26.
		{Code: "<App></App>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "never", "beforeClosing": "allow"}}},
		// Upstream valid 27.
		{Code: "< App></ App>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "always", "beforeClosing": "allow"}}},
		// Upstream valid 28.
		{Code: "< App/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "always", "beforeClosing": "allow"}}},
		// Upstream valid 29.
		{Code: "\n        <\n        App/>\n      ", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow-multiline", "beforeClosing": "allow"}}},
		// Upstream valid 30.
		{Code: "<App />", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "never"}}},
		// Upstream valid 31.
		{Code: "<App></App>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "never"}}},
		// Upstream valid 32.
		{Code: "\n        <App\n        foo=\"bar\"\n        >\n        </App>\n      ", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "never"}}},
		// Upstream valid 33.
		{Code: "\n        <App\n           foo=\"bar\"\n        >\n        </App>\n      ", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "never"}}},
		// Upstream valid 34.
		{Code: "<App ></App >", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "always"}}},
		// Upstream valid 35.
		{Code: "\n        <App\n        foo=\"bar\"\n        >\n        </App >\n      ", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "always"}}},
		// Upstream valid 36.
		{Code: "\n        <App\n            foo=\"bar\"\n        >\n        </App >\n      ", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "always"}}},
		// Upstream valid 37.
		{Code: "<App/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "never", "beforeSelfClosing": "never", "afterOpening": "never", "beforeClosing": "never"}}},
		// Upstream valid 38.
		{Code: "< App / >", Tsx: true, Options: []any{map[string]any{"closingSlash": "always", "beforeSelfClosing": "always", "afterOpening": "always", "beforeClosing": "always"}}},
	}, []rule_tester.InvalidTestCase{
		// Upstream invalid 1.
		{Code: "<App/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "always", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"<App />"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNeedSpace", Message: "A space is required before closing bracket", Line: 1, Column: 5, EndLine: 1, EndColumn: 5},
		}},
		// Upstream invalid 2.
		{Code: "<App foo/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "always", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"<App foo />"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNeedSpace", Message: "A space is required before closing bracket", Line: 1, Column: 9, EndLine: 1, EndColumn: 9},
		}},
		// Upstream invalid 3.
		{Code: "<App foo={bar}/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "always", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"<App foo={bar} />"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNeedSpace", Message: "A space is required before closing bracket", Line: 1, Column: 15, EndLine: 1, EndColumn: 15},
		}},
		// Upstream invalid 4.
		{Code: "<App {...props}/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "always", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"<App {...props} />"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNeedSpace", Message: "A space is required before closing bracket", Line: 1, Column: 16, EndLine: 1, EndColumn: 16},
		}},
		// Upstream invalid 5.
		{Code: "<App />", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "never", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"<App/>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNoSpace", Message: "A space is forbidden before closing bracket", Line: 1, Column: 6, EndLine: 1, EndColumn: 6},
		}},
		// Upstream invalid 6.
		{Code: "<App foo />", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "never", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"<App foo/>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNoSpace", Message: "A space is forbidden before closing bracket", Line: 1, Column: 10, EndLine: 1, EndColumn: 10},
		}},
		// Upstream invalid 7.
		{Code: "<App foo={bar} />", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "never", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"<App foo={bar}/>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNoSpace", Message: "A space is forbidden before closing bracket", Line: 1, Column: 16, EndLine: 1, EndColumn: 16},
		}},
		// Upstream invalid 8.
		{Code: "<App/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "proportional-always", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"<App />"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNeedSpace", Message: "A space is required before closing bracket", Line: 1, Column: 5, EndLine: 1, EndColumn: 5},
		}},
		// Upstream invalid 9.
		{Code: "<App foo/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "proportional-always", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"<App foo />"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNeedSpace", Message: "A space is required before closing bracket", Line: 1, Column: 9, EndLine: 1, EndColumn: 9},
		}},
		// Upstream invalid 10.
		{Code: "\n        <App\n          foo={bar}/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "proportional-always", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"\n        <App\n          foo={bar}\n/>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNeedNewline", Message: "A newline is required before closing bracket", Line: 3, Column: 20, EndLine: 3, EndColumn: 20},
		}},
		// Upstream invalid 11.
		{Code: "\n        <App\n          foo={bar} />", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "proportional-always", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"\n        <App\n          foo={bar} \n/>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNeedNewline", Message: "A newline is required before closing bracket", Line: 3, Column: 20, EndLine: 3, EndColumn: 20},
		}},
		// Upstream invalid 12.
		{Code: "\n        <App\n          foo={bar}\n          blat >\n          hello\n        </App>\n      ", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "proportional-always"}}, Output: []string{"\n        <App\n          foo={bar}\n          blat \n>\n          hello\n        </App>\n      "}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeCloseNeedNewline", Message: "A newline is required before closing bracket", Line: 4, Column: 15, EndLine: 4, EndColumn: 15},
		}},
		// Upstream invalid 13.
		{Code: "\n        <App\n          foo={bar}>\n          hello\n        </App>\n      ", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "proportional-always"}}, Output: []string{"\n        <App\n          foo={bar}\n>\n          hello\n        </App>\n      "}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeCloseNeedNewline", Message: "A newline is required before closing bracket", Line: 3, Column: 20, EndLine: 3, EndColumn: 20},
		}},
		// Upstream invalid 14.
		{Code: "<App {...props} />", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "never", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"<App {...props}/>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNoSpace", Message: "A space is forbidden before closing bracket", Line: 1, Column: 17, EndLine: 1, EndColumn: 17},
		}},
		// Upstream invalid 15.
		{Code: "<App/ >;", Tsx: true, Options: []any{map[string]any{"closingSlash": "never", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"<App/>;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "selfCloseSlashNoSpace", Message: "Whitespace is forbidden between `/` and `>`; write `/>`", Line: 1, Column: 5, EndLine: 1, EndColumn: 8},
		}},
		// Upstream invalid 16.
		{Code: "\n        <App_selfCloseSlashNoSpace/\n        >\n      ", Tsx: true, Options: []any{map[string]any{"closingSlash": "never", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"\n        <App_selfCloseSlashNoSpace/>\n      "}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "selfCloseSlashNoSpace", Message: "Whitespace is forbidden between `/` and `>`; write `/>`", Line: 2, Column: 35, EndLine: 3, EndColumn: 10},
		}},
		// Upstream invalid 17.
		{Code: "<div className=\"bar\">< /div>;", Tsx: true, Options: []any{map[string]any{"closingSlash": "never", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "allow"}}, Skip: true, // tsgo does not parse whitespace between `<` and `/` in closing tags.
			Output: []string{"<div className=\"bar\"></div>;"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "closeSlashNoSpace", Message: "Whitespace is forbidden between `<` and `/`; write `</`", Line: 1, Column: 22, EndLine: 1, EndColumn: 25},
			}},
		// Upstream invalid 18.
		{Code: "\n        <div className=\"bar\"><\n        /div>;\n      ", Tsx: true, Options: []any{map[string]any{"closingSlash": "never", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "allow"}}, Skip: true, // tsgo does not parse whitespace between `<` and `/` in closing tags.
			Output: []string{"\n        <div className=\"bar\"></div>;\n      "}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "closeSlashNoSpace", Message: "Whitespace is forbidden between `<` and `/`; write `</`", Line: 2, Column: 30, EndLine: 3, EndColumn: 10},
			}},
		// Upstream invalid 19.
		// The upstream fix produces `< /Tag>`, which tsgo cannot reparse.
		// The diagnostic and edit are covered by TestJsxTagSpacingClosingSlashFix.
		{Code: "<App prop=\"foo\"></App>", Tsx: true, Skip: true, Options: []any{map[string]any{"closingSlash": "always", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"<App prop=\"foo\">< /App>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "closeSlashNeedSpace", Message: "Whitespace is required between `<` and `/`; write `< /`", Line: 1, Column: 17, EndLine: 1, EndColumn: 19},
		}},
		// Upstream invalid 20.
		{Code: "<p/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "always", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"<p/ >"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "selfCloseSlashNeedSpace", Message: "Whitespace is required between `/` and `>`; write `/ >`", Line: 1, Column: 3, EndLine: 1, EndColumn: 5},
		}},
		// Upstream invalid 21.
		{Code: "< App/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "never", "beforeClosing": "allow"}}, Output: []string{"<App/>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "afterOpenNoSpace", Message: "A space is forbidden after opening bracket", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
		}},
		// Upstream invalid 22.
		{Code: "< App></App>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "never", "beforeClosing": "allow"}}, Output: []string{"<App></App>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "afterOpenNoSpace", Message: "A space is forbidden after opening bracket", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
		}},
		// Upstream invalid 23.
		{Code: "<App></ App>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "never", "beforeClosing": "allow"}}, Output: []string{"<App></App>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "afterOpenNoSpace", Message: "A space is forbidden after opening bracket", Line: 1, Column: 7, EndLine: 1, EndColumn: 9},
		}},
		// Upstream invalid 24.
		{Code: "< App></ App>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "never", "beforeClosing": "allow"}}, Output: []string{"<App></App>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "afterOpenNoSpace", Message: "A space is forbidden after opening bracket", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			{MessageId: "afterOpenNoSpace", Message: "A space is forbidden after opening bracket", Line: 1, Column: 8, EndLine: 1, EndColumn: 10},
		}},
		// Upstream invalid 25.
		{Code: "\n        <\n        App1/>\n      ", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "never", "beforeClosing": "allow"}}, Output: []string{"\n        <App1/>\n      "}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "afterOpenNoSpace", Message: "A space is forbidden after opening bracket", Line: 2, Column: 9, EndLine: 3, EndColumn: 9},
		}},
		// Upstream invalid 26.
		{Code: "<App2></ App2>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "always", "beforeClosing": "allow"}}, Output: []string{"< App2></ App2>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "afterOpenNeedSpace", Message: "A space is required after opening bracket", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
		}},
		// Upstream invalid 27.
		{Code: "< App3></App3>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "always", "beforeClosing": "allow"}}, Output: []string{"< App3></ App3>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "afterOpenNeedSpace", Message: "A space is required after opening bracket", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
		}},
		// Upstream invalid 28.
		{Code: "<App4></App4>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "always", "beforeClosing": "allow"}}, Output: []string{"< App4></ App4>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "afterOpenNeedSpace", Message: "A space is required after opening bracket", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
			{MessageId: "afterOpenNeedSpace", Message: "A space is required after opening bracket", Line: 1, Column: 8, EndLine: 1, EndColumn: 9},
		}},
		// Upstream invalid 29.
		{Code: "<App5/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "always", "beforeClosing": "allow"}}, Output: []string{"< App5/>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "afterOpenNeedSpace", Message: "A space is required after opening bracket", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
		}},
		// Upstream invalid 30.
		{Code: "< App6/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow-multiline", "beforeClosing": "allow"}}, Output: []string{"<App6/>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "afterOpenNoSpace", Message: "A space is forbidden after opening bracket", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
		}},
		// Upstream invalid 31.
		{Code: "<App7 ></App7>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "never"}}, Output: []string{"<App7></App7>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeCloseNoSpace", Message: "A space is forbidden before closing bracket", Line: 1, Column: 6, EndLine: 1, EndColumn: 7},
		}},
		// Upstream invalid 32.
		{Code: "<App8></App8 >", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "never"}}, Output: []string{"<App8></App8>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeCloseNoSpace", Message: "A space is forbidden before closing bracket", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
		}},
		// Upstream invalid 33.
		{Code: "\n        <App9\n        foo=\"bar\"\n        >\n        </App9 >\n      ", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "never"}}, Output: []string{"\n        <App9\n        foo=\"bar\"\n        >\n        </App9>\n      "}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeCloseNoSpace", Message: "A space is forbidden before closing bracket", Line: 5, Column: 15, EndLine: 5, EndColumn: 16},
		}},
		// Upstream invalid 34.
		{Code: "<App10></App10 >", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "always"}}, Output: []string{"<App10 ></App10 >"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeCloseNeedSpace", Message: "Whitespace is required before closing bracket", Line: 1, Column: 7, EndLine: 1, EndColumn: 7},
		}},
		// Upstream invalid 35.
		{Code: "<App11 ></App11>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "always"}}, Output: []string{"<App11 ></App11 >"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeCloseNeedSpace", Message: "Whitespace is required before closing bracket", Line: 1, Column: 16, EndLine: 1, EndColumn: 16},
		}},
		// Upstream invalid 36.
		{Code: "\n        <App12\n        foo=\"bar\"\n        >\n        </App12>\n      ", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "always"}}, Output: []string{"\n        <App12\n        foo=\"bar\"\n        >\n        </App12 >\n      "}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeCloseNeedSpace", Message: "Whitespace is required before closing bracket", Line: 5, Column: 16, EndLine: 5, EndColumn: 16},
		}},
	})
}

// Documentation examples isolate the described option, as upstream's test helpers do.
func TestJsxTagSpacingDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxTagSpacingRule, []rule_tester.ValidTestCase{
		// Documentation valid 1.
		{Code: "<App/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "never", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Documentation valid 2.
		{Code: "<input/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "never", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Documentation valid 3.
		{Code: "<Provider></Provider>", Tsx: true, Options: []any{map[string]any{"closingSlash": "never", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Documentation valid 4.
		{Code: "<Hello/ >", Tsx: true, Options: []any{map[string]any{"closingSlash": "always", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Documentation valid 5.
		{Code: "<Goodbye>< /Goodbye>", Tsx: true, Options: []any{map[string]any{"closingSlash": "always", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "allow"}}, Skip: true}, // tsgo does not parse whitespace between `<` and `/` in closing tags.

		// Documentation valid 6.
		{Code: "<Hello />", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "always", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Documentation valid 7.
		{Code: "<Hello firstName=\"John\" />", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "always", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Documentation valid 8.
		{Code: "<Hello\n  firstName=\"John\"\n  lastName=\"Smith\"\n/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "always", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Documentation valid 9.
		{Code: "<Hello/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "never", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Documentation valid 10.
		{Code: "<Hello firstname=\"John\"/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "never", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Documentation valid 11.
		{Code: "<Hello\n  firstName=\"John\"\n  lastName=\"Smith\"\n/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "never", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Documentation valid 12.
		{Code: "<Hello\n  firstName=\"John\"\n  lastName=\"Smith\"\n/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "proportional-always", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Documentation valid 13.
		{Code: "< Hello></ Hello>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "always", "beforeClosing": "allow"}}},
		// Documentation valid 14.
		{Code: "< Hello firstName=\"John\"/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "always", "beforeClosing": "allow"}}},
		// Documentation valid 15.
		{Code: "<\n  Hello\n  firstName=\"John\"\n  lastName=\"Smith\"\n/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "always", "beforeClosing": "allow"}}},
		// Documentation valid 16.
		{Code: "<Hello></Hello>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "never", "beforeClosing": "allow"}}},
		// Documentation valid 17.
		{Code: "<Hello firstname=\"John\"/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "never", "beforeClosing": "allow"}}},
		// Documentation valid 18.
		{Code: "<Hello\n  firstName=\"John\"\n  lastName=\"Smith\"\n/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "never", "beforeClosing": "allow"}}},
		// Documentation valid 19.
		{Code: "<Hello></Hello>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow-multiline", "beforeClosing": "allow"}}},
		// Documentation valid 20.
		{Code: "<Hello firstName=\"John\"/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow-multiline", "beforeClosing": "allow"}}},
		// Documentation valid 21.
		{Code: "<\n  Hello\n  firstName=\"John\"\n  lastName=\"Smith\"\n/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow-multiline", "beforeClosing": "allow"}}},
		// Documentation valid 22.
		{Code: "<Hello ></Hello >", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "always"}}},
		// Documentation valid 23.
		{Code: "<Hello\n  firstName=\"John\"\n>\n</Hello >", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "always"}}},
		// Documentation valid 24.
		{Code: "<Hello></Hello>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "never"}}},
		// Documentation valid 25.
		{Code: "<Hello\n  firstName=\"John\"\n>\n</Hello>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "never"}}},
		// Documentation valid 26.
		{Code: "<Hello\n  firstName=\"John\"\n  lastName=\"Smith\"\n>\n  Goodbye\n</Hello>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "proportional-always"}}},
	}, []rule_tester.InvalidTestCase{
		// Documentation invalid 1.
		{Code: "<App/ >", Tsx: true, Options: []any{map[string]any{"closingSlash": "never", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"<App/>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "selfCloseSlashNoSpace", Message: "Whitespace is forbidden between `/` and `>`; write `/>`", Line: 1, Column: 5, EndLine: 1, EndColumn: 8},
		}},
		// Documentation invalid 2.
		{Code: "<input/\n>", Tsx: true, Options: []any{map[string]any{"closingSlash": "never", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"<input/>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "selfCloseSlashNoSpace", Message: "Whitespace is forbidden between `/` and `>`; write `/>`", Line: 1, Column: 7, EndLine: 2, EndColumn: 2},
		}},
		// Documentation invalid 3.
		{Code: "<Provider>< /Provider>", Tsx: true, Options: []any{map[string]any{"closingSlash": "never", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "allow"}}, Skip: true, // tsgo does not parse whitespace between `<` and `/` in closing tags.
			Output: []string{"<Provider></Provider>"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "closeSlashNoSpace", Message: "Whitespace is forbidden between `<` and `/`; write `</`", Line: 1, Column: 11, EndLine: 1, EndColumn: 14},
			}},
		// Documentation invalid 4.
		{Code: "<Hello/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "always", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"<Hello/ >"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "selfCloseSlashNeedSpace", Message: "Whitespace is required between `/` and `>`; write `/ >`", Line: 1, Column: 7, EndLine: 1, EndColumn: 9},
		}},
		// Documentation invalid 5.
		// The upstream fix produces `< /Tag>`, which tsgo cannot reparse.
		// The diagnostic and edit are covered by TestJsxTagSpacingClosingSlashFix.
		{Code: "<Goodbye></Goodbye>", Tsx: true, Skip: true, Options: []any{map[string]any{"closingSlash": "always", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"<Goodbye>< /Goodbye>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "closeSlashNeedSpace", Message: "Whitespace is required between `<` and `/`; write `< /`", Line: 1, Column: 10, EndLine: 1, EndColumn: 12},
		}},
		// Documentation invalid 6.
		{Code: "<Hello/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "always", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"<Hello />"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNeedSpace", Message: "A space is required before closing bracket", Line: 1, Column: 7, EndLine: 1, EndColumn: 7},
		}},
		// Documentation invalid 7.
		{Code: "<Hello firstname=\"John\"/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "always", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"<Hello firstname=\"John\" />"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNeedSpace", Message: "A space is required before closing bracket", Line: 1, Column: 24, EndLine: 1, EndColumn: 24},
		}},
		// Documentation invalid 8.
		{Code: "<Hello />", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "never", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"<Hello/>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNoSpace", Message: "A space is forbidden before closing bracket", Line: 1, Column: 8, EndLine: 1, EndColumn: 8},
		}},
		// Documentation invalid 9.
		{Code: "<Hello firstName=\"John\" />", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "never", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"<Hello firstName=\"John\"/>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNoSpace", Message: "A space is forbidden before closing bracket", Line: 1, Column: 25, EndLine: 1, EndColumn: 25},
		}},
		// Documentation invalid 10.
		{Code: "<Hello\n  firstName=\"John\"\n  lastName=\"Smith\" />", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "proportional-always", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"<Hello\n  firstName=\"John\"\n  lastName=\"Smith\" \n/>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNeedNewline", Message: "A newline is required before closing bracket", Line: 3, Column: 19, EndLine: 3, EndColumn: 19},
		}},
		// Documentation invalid 11.
		{Code: "<Hello\n  firstName=\"John\"\n  lastName=\"Smith\"/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "proportional-always", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"<Hello\n  firstName=\"John\"\n  lastName=\"Smith\"\n/>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNeedNewline", Message: "A newline is required before closing bracket", Line: 3, Column: 19, EndLine: 3, EndColumn: 19},
		}},
		// Documentation invalid 12.
		{Code: "<Hello></Hello>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "always", "beforeClosing": "allow"}}, Output: []string{"< Hello></ Hello>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "afterOpenNeedSpace", Message: "A space is required after opening bracket", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
			{MessageId: "afterOpenNeedSpace", Message: "A space is required after opening bracket", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
		}},
		// Documentation invalid 13.
		{Code: "<Hello firstname=\"John\"/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "always", "beforeClosing": "allow"}}, Output: []string{"< Hello firstname=\"John\"/>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "afterOpenNeedSpace", Message: "A space is required after opening bracket", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
		}},
		// Documentation invalid 14.
		{Code: "<Hello\n  firstName=\"John\"\n  lastName=\"Smith\"\n/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "always", "beforeClosing": "allow"}}, Output: []string{"< Hello\n  firstName=\"John\"\n  lastName=\"Smith\"\n/>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "afterOpenNeedSpace", Message: "A space is required after opening bracket", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
		}},
		// Documentation invalid 15.
		{Code: "< Hello></ Hello>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "never", "beforeClosing": "allow"}}, Output: []string{"<Hello></Hello>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "afterOpenNoSpace", Message: "A space is forbidden after opening bracket", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			{MessageId: "afterOpenNoSpace", Message: "A space is forbidden after opening bracket", Line: 1, Column: 10, EndLine: 1, EndColumn: 12},
		}},
		// Documentation invalid 16.
		{Code: "< Hello firstName=\"John\"/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "never", "beforeClosing": "allow"}}, Output: []string{"<Hello firstName=\"John\"/>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "afterOpenNoSpace", Message: "A space is forbidden after opening bracket", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
		}},
		// Documentation invalid 17.
		{Code: "<\n  Hello\n  firstName=\"John\"\n  lastName=\"Smith\"\n/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "never", "beforeClosing": "allow"}}, Output: []string{"<Hello\n  firstName=\"John\"\n  lastName=\"Smith\"\n/>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "afterOpenNoSpace", Message: "A space is forbidden after opening bracket", Line: 1, Column: 1, EndLine: 2, EndColumn: 3},
		}},
		// Documentation invalid 18.
		{Code: "< Hello></ Hello>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow-multiline", "beforeClosing": "allow"}}, Output: []string{"<Hello></Hello>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "afterOpenNoSpace", Message: "A space is forbidden after opening bracket", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			{MessageId: "afterOpenNoSpace", Message: "A space is forbidden after opening bracket", Line: 1, Column: 10, EndLine: 1, EndColumn: 12},
		}},
		// Documentation invalid 19.
		{Code: "< Hello firstName=\"John\"/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow-multiline", "beforeClosing": "allow"}}, Output: []string{"<Hello firstName=\"John\"/>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "afterOpenNoSpace", Message: "A space is forbidden after opening bracket", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
		}},
		// Documentation invalid 20.
		{Code: "< Hello\n  firstName=\"John\"\n  lastName=\"Smith\"\n/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow-multiline", "beforeClosing": "allow"}}, Output: []string{"<Hello\n  firstName=\"John\"\n  lastName=\"Smith\"\n/>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "afterOpenNoSpace", Message: "A space is forbidden after opening bracket", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
		}},
		// Documentation invalid 21.
		{Code: "<Hello></Hello>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "always"}}, Output: []string{"<Hello ></Hello >"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeCloseNeedSpace", Message: "Whitespace is required before closing bracket", Line: 1, Column: 7, EndLine: 1, EndColumn: 7},
			{MessageId: "beforeCloseNeedSpace", Message: "Whitespace is required before closing bracket", Line: 1, Column: 15, EndLine: 1, EndColumn: 15},
		}},
		// Documentation invalid 22.
		{Code: "<Hello></Hello >", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "always"}}, Output: []string{"<Hello ></Hello >"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeCloseNeedSpace", Message: "Whitespace is required before closing bracket", Line: 1, Column: 7, EndLine: 1, EndColumn: 7},
		}},
		// Documentation invalid 23.
		{Code: "<Hello ></Hello>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "always"}}, Output: []string{"<Hello ></Hello >"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeCloseNeedSpace", Message: "Whitespace is required before closing bracket", Line: 1, Column: 16, EndLine: 1, EndColumn: 16},
		}},
		// Documentation invalid 24.
		{Code: "<Hello ></Hello>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "never"}}, Output: []string{"<Hello></Hello>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeCloseNoSpace", Message: "A space is forbidden before closing bracket", Line: 1, Column: 7, EndLine: 1, EndColumn: 8},
		}},
		// Documentation invalid 25.
		{Code: "<Hello></Hello >", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "never"}}, Output: []string{"<Hello></Hello>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeCloseNoSpace", Message: "A space is forbidden before closing bracket", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
		}},
		// Documentation invalid 26.
		{Code: "<Hello ></Hello >", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "never"}}, Output: []string{"<Hello></Hello>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeCloseNoSpace", Message: "A space is forbidden before closing bracket", Line: 1, Column: 7, EndLine: 1, EndColumn: 8},
			{MessageId: "beforeCloseNoSpace", Message: "A space is forbidden before closing bracket", Line: 1, Column: 16, EndLine: 1, EndColumn: 17},
		}},
		// Documentation invalid 27.
		{Code: "<Hello\n  firstName=\"John\"\n  lastName=\"Smith\">\n</Hello>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "proportional-always"}}, Output: []string{"<Hello\n  firstName=\"John\"\n  lastName=\"Smith\"\n>\n</Hello>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeCloseNeedNewline", Message: "A newline is required before closing bracket", Line: 3, Column: 19, EndLine: 3, EndColumn: 19},
		}},
		// Documentation invalid 28.
		{Code: "<Hello\n  firstName=\"John\"\n  lastName=\"Smith\" >\n  Goodbye\n</Hello>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "proportional-always"}}, Output: []string{"<Hello\n  firstName=\"John\"\n  lastName=\"Smith\" \n>\n  Goodbye\n</Hello>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeCloseNeedNewline", Message: "A newline is required before closing bracket", Line: 3, Column: 19, EndLine: 3, EndColumn: 19},
		}},
	})
}
