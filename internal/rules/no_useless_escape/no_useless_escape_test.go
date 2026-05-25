package no_useless_escape

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoUselessEscapeRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUselessEscapeRule,
		[]rule_tester.ValidTestCase{
			// ---- Regex: meaningful escapes outside character classes ----
			{Code: `var foo = /\./`},
			{Code: `var foo = /\//g`},
			{Code: `var foo = /""/`},
			{Code: `var foo = /''/`},
			{Code: `var foo = /([A-Z])\t+/g`},
			{Code: `var foo = /([A-Z])\n+/g`},
			{Code: `var foo = /([A-Z])\v+/g`},
			{Code: `var foo = /\D/`},
			{Code: `var foo = /\W/`},
			{Code: `var foo = /\w/`},
			{Code: `var foo = /\\/g`},
			{Code: `var foo = /\w\$\*\./`},
			{Code: `var foo = /\^\+\./`},
			{Code: `var foo = /\|\}\{\./`},
			{Code: `var foo = /]\[\(\)\//`},

			// ---- String literals: valid escape sequences ----
			{Code: `var foo = "\x12"`},
			{Code: `var foo = "©"`},
			{Code: `var foo = "\""`},
			{Code: `var foo = "xsℑ"`},
			{Code: `var foo = "foo \\ bar";`},
			{Code: `var foo = "\t";`},
			{Code: `var foo = "foo \b bar";`},
			{Code: `var foo = '\n';`},
			{Code: `var foo = 'foo \r bar';`},
			{Code: `var foo = '\v';`},
			{Code: `var foo = '\f';`},
			// Line continuations: \<LF>, \<CRLF>
			{Code: "var foo = '\\\n';"},
			{Code: "var foo = '\\\r\n';"},
			// LineSeparator / ParagraphSeparator continuations.
			{Code: "var foo = 'foo \\  bar'"},
			{Code: "var foo = 'foo \\  bar'"},

			// ---- Template literals: valid escape sequences ----
			{Code: "var foo = `\\x12`"},
			{Code: "var foo = `\\u00a9`"},
			{Code: "var foo = `xs\\u2111`"},
			{Code: "var foo = `foo \\\\ bar`;"},
			{Code: "var foo = `\\t`;"},
			{Code: "var foo = `foo \\b bar`;"},
			{Code: "var foo = `\\n`;"},
			{Code: "var foo = `foo \\r bar`;"},
			{Code: "var foo = `\\v`;"},
			{Code: "var foo = `\\f`;"},
			{Code: "var foo = `\\\n`;"},
			{Code: "var foo = `\\\r\n`;"},
			{Code: "var foo = `${foo} \\x12`"},
			{Code: "var foo = `${foo} \\u00a9`"},
			{Code: "var foo = `${foo} xs\\u2111`"},
			{Code: "var foo = `${foo} \\\\ ${bar}`;"},
			{Code: "var foo = `${foo} \\b ${bar}`;"},
			{Code: "var foo = `${foo}\\t`;"},
			{Code: "var foo = `${foo}\\n`;"},
			{Code: "var foo = `${foo}\\r`;"},
			{Code: "var foo = `${foo}\\v`;"},
			{Code: "var foo = `${foo}\\f`;"},
			{Code: "var foo = `${foo}\\\n`;"},
			{Code: "var foo = `${foo}\\\r\n`;"},
			// Quote escape inside template: `\``.
			{Code: "var foo = `\\``"},
			{Code: "var foo = `\\`${foo}\\``"},
			// `\$` followed by `{` and `\{` preceded by `$` are necessary.
			{Code: "var foo = `\\${{${foo}`;"},
			{Code: "var foo = `$\\{{${foo}`;"},
			// Tagged template — escapes are exposed via the `raw` array.
			{Code: "var foo = String.raw`\\.`"},
			{Code: "var foo = myFunc`\\.`"},

			// ---- Regex character classes ----
			{Code: `var foo = /[\d]/`},
			{Code: `var foo = /[a\-b]/`},
			{Code: `var foo = /foo\?/`},
			{Code: `var foo = /example\.com/`},
			{Code: `var foo = /foo\|bar/`},
			{Code: `var foo = /\^bar/`},
			{Code: `var foo = /[\^bar]/`},
			{Code: `var foo = /\(bar\)/`},
			{Code: `var foo = /[[\]]/`}, // class containing '[' and ']'
			{Code: `var foo = /[[]\./`},
			{Code: `var foo = /[\]\]]/`},
			{Code: `var foo = /\[abc]/`},
			{Code: `var foo = /\[foo\.bar]/`},
			{Code: `var foo = /vi/m`},
			{Code: `var foo = /\B/`},

			// ---- Special regex escapes (issue #7472) ----
			{Code: `var foo = /\0/`},
			{Code: `var foo = /\1/`},
			{Code: `var foo = /(a)\1/`},
			{Code: `var foo = /(a)\12/`},
			{Code: `var foo = /[\0]/`},

			// ---- Unicode-property and named-backreference escapes (uvMode) ----
			{Code: `var foo = /(?<a>)\k<a>/`},
			{Code: `var foo = /(\\?<a>)/`},
			{Code: `var foo = /\p{ASCII}/u`},
			{Code: `var foo = /\P{ASCII}/u`},
			{Code: `var foo = /[\p{ASCII}]/u`},
			{Code: `var foo = /[\P{ASCII}]/u`},

			// ---- Carets ----
			{Code: `/[^^]/`},
			{Code: `/[^^]/u`},

			// ---- ES2024 v-flag character class escapes ----
			{Code: "/[\\q{abc}]/v"},
			{Code: `/[\(]/v`},
			{Code: `/[\)]/v`},
			{Code: `/[\{]/v`},
			{Code: `/[\]]/v`},
			{Code: `/[\}]/v`},
			{Code: `/[\/]/v`},
			{Code: `/[\-]/v`},
			{Code: `/[\|]/v`},
			{Code: `/[\$$]/v`},
			{Code: `/[\&&]/v`},
			{Code: `/[\!!]/v`},
			{Code: `/[\##]/v`},
			{Code: `/[\%%]/v`},
			{Code: `/[\**]/v`},
			{Code: `/[\++]/v`},
			{Code: `/[\,,]/v`},
			{Code: `/[\..]/v`},
			{Code: `/[\::]/v`},
			{Code: `/[\;;]/v`},
			{Code: `/[\<<]/v`},
			{Code: `/[\==]/v`},
			{Code: `/[\>>]/v`},
			{Code: `/[\??]/v`},
			{Code: `/[\@@]/v`},
			{Code: "/[\\``]/v"},
			{Code: `/[\~~]/v`},
			{Code: `/[^\^^]/v`},
			{Code: `/[_\^^]/v`},
			{Code: `/[$\$]/v`},
			{Code: `/[&\&]/v`},
			{Code: `/[!\!]/v`},
			{Code: `/[#\#]/v`},
			{Code: `/[%\%]/v`},
			{Code: `/[*\*]/v`},
			{Code: `/[+\+]/v`},
			{Code: `/[,\,]/v`},
			{Code: `/[.\.]/v`},
			{Code: `/[:\:]/v`},
			{Code: `/[;\;]/v`},
			{Code: `/[<\<]/v`},
			{Code: `/[=\=]/v`},
			{Code: `/[>\>]/v`},
			{Code: `/[?\?]/v`},
			{Code: `/[@\@]/v`},
			{Code: "/[`\\`]/v"},
			{Code: `/[~\~]/v`},
			{Code: `/[^^\^]/v`},
			{Code: `/[_^\^]/v`},
			{Code: `/[\&&&\&]/v`},
			{Code: `/[[\-]\-]/v`},
			{Code: `/[\^]/v`},

			// ---- Option allowRegexCharacters ----
			{Code: `var foo = /\#/;`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"#"}}},
			{Code: `var foo = /\;/;`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{";"}}},
			{Code: `var foo = /\#\;/;`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"#", ";"}}},
			{Code: `var foo = /[ab\-]/`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"-"}}},
			{Code: `var foo = /[\-ab]/`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"-"}}},
			{Code: `var foo = /[ab\?]/`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"?"}}},
			{Code: `var foo = /[ab\.]/`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"."}}},
			{Code: `var foo = /[a\|b]/`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"|"}}},
			{Code: `var foo = /\-/`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"-"}}},
			{Code: `var foo = /[\-]/`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"-"}}},
			{Code: `var foo = /[ab\$]/`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"$"}}},
			{Code: `var foo = /[\(paren]/`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"("}}},
			{Code: `var foo = /[\[]/`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"["}}},
			{Code: `var foo = /[\/]/`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"/"}}},
			{Code: `var foo = /[\B]/`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"B"}}},
			{Code: `var foo = /[a][\-b]/`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"-"}}},
			{Code: `var foo = /\-[]/`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"-"}}},
			{Code: `var foo = /[a\^]/`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"^"}}},
			{Code: `/[^\^]/`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"^"}}},
			{Code: `/[^\^]/u`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"^"}}},
			{Code: `/[\$]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"$"}}},
			{Code: `/[\&\&]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"&"}}},
			{Code: `/[\!!]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"!"}}},
			{Code: `/[\##]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"#"}}},
			{Code: `/[\%%]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"%"}}},
			{Code: `/[\*\*]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"*"}}},
			{Code: `/[\+\+]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"+"}}},
			{Code: `/[\,,]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{","}}},
			{Code: `/[\..]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"."}}},
			{Code: `/[\:\:]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{":"}}},
			{Code: `/[\;\;]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{";"}}},
			{Code: `/[\<\<]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"<"}}},
			{Code: `/[\=\=]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"="}}},
			{Code: `/[\>\>]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{">"}}},
			{Code: `/[\?\?]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"?"}}},
			{Code: `/[\@\@]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"@"}}},
			{Code: "/[\\``]/v", Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"`"}}},
			{Code: `/[\~\~]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"~"}}},
			{Code: `/[^\^\^]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"^"}}},
			{Code: `/[_\^\^]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"^"}}},
			{Code: `/[\&\&&\&]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"&"}}},
			{Code: `/[\p{ASCII}--\.]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"."}}},
			{Code: `/[\p{ASCII}&&\.]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"."}}},
			{Code: `/[\.--[.&]]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"."}}},
			{Code: `/[\.&&[.&]]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"."}}},
			{Code: `/[\.--\.--\.]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"."}}},
			{Code: `/[\.&&\.&&\.]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"."}}},
			{Code: `/[[\.&]--[\.&]]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"."}}},
			{Code: `/[[\.&]&&[\.&]]/v`, Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"."}}},

			// ---- JSX attribute strings — the only JSX shape where StringLiteral
			// is a direct AST child of a JSX-related node. These MUST not fire;
			// other JSX-child shapes wrap StringLiteral in JsxText / JsxExpression
			// and never reach KindJsxElement / KindJsxFragment as direct parent.
			{Code: `var x = <foo attr="\d"/>`, Tsx: true},
			{Code: `var x = <foo attr='\d'></foo>`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Regex outside character class ----
			{
				Code: `var foo = /\#/;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \#.`,
					Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = /#/;`},
						{MessageId: "escapeBackslash", Output: `var foo = /\\#/;`},
					},
				}},
			},
			{
				Code: `var foo = /\;/;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \;.`,
					Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = /;/;`},
						{MessageId: "escapeBackslash", Output: `var foo = /\\;/;`},
					},
				}},
			},

			// ---- String literals: identity escapes that aren't valid ----
			{
				Code: `var foo = "\'";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \'.`,
					Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = "'";`},
						{MessageId: "escapeBackslash", Output: `var foo = "\\'";`},
					},
				}},
			},
			{
				Code: `var foo = "\#/";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \#.`,
					Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = "#/";`},
						{MessageId: "escapeBackslash", Output: `var foo = "\\#/";`},
					},
				}},
			},
			{
				Code: `var foo = "\a"`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \a.`,
					Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = "a"`},
						{MessageId: "escapeBackslash", Output: `var foo = "\\a"`},
					},
				}},
			},
			{
				Code: `var foo = "\B";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \B.`,
					Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = "B";`},
						{MessageId: "escapeBackslash", Output: `var foo = "\\B";`},
					},
				}},
			},
			{
				Code: `var foo = "\@";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \@.`,
					Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = "@";`},
						{MessageId: "escapeBackslash", Output: `var foo = "\\@";`},
					},
				}},
			},
			{
				Code: `var foo = "foo \a bar";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \a.`,
					Line:      1, Column: 16, EndLine: 1, EndColumn: 17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = "foo a bar";`},
						{MessageId: "escapeBackslash", Output: `var foo = "foo \\a bar";`},
					},
				}},
			},
			{
				Code: `var foo = '\"';`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \".`,
					Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = '"';`},
						{MessageId: "escapeBackslash", Output: `var foo = '\\"';`},
					},
				}},
			},
			{
				Code: `var foo = '\#';`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \#.`,
					Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = '#';`},
						{MessageId: "escapeBackslash", Output: `var foo = '\\#';`},
					},
				}},
			},
			{
				Code: `var foo = '\$';`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \$.`,
					Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = '$';`},
						{MessageId: "escapeBackslash", Output: `var foo = '\\$';`},
					},
				}},
			},
			{
				Code: `var foo = '\p';`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \p.`,
					Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = 'p';`},
						{MessageId: "escapeBackslash", Output: `var foo = '\\p';`},
					},
				}},
			},
			{
				Code: `var foo = '\p\a\@';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryEscape",
						Message:   `Unnecessary escape character: \p.`,
						Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: `var foo = 'p\a\@';`},
							{MessageId: "escapeBackslash", Output: `var foo = '\\p\a\@';`},
						},
					},
					{
						MessageId: "unnecessaryEscape",
						Message:   `Unnecessary escape character: \a.`,
						Line:      1, Column: 14, EndLine: 1, EndColumn: 15,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: `var foo = '\pa\@';`},
							{MessageId: "escapeBackslash", Output: `var foo = '\p\\a\@';`},
						},
					},
					{
						MessageId: "unnecessaryEscape",
						Message:   `Unnecessary escape character: \@.`,
						Line:      1, Column: 16, EndLine: 1, EndColumn: 17,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: `var foo = '\p\a@';`},
							{MessageId: "escapeBackslash", Output: `var foo = '\p\a\\@';`},
						},
					},
				},
			},
			{
				Code: `var foo = '\` + "`" + `';`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   "Unnecessary escape character: \\`.",
					Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: "var foo = '`';"},
						{MessageId: "escapeBackslash", Output: "var foo = '\\\\`';"},
					},
				}},
			},

			// ---- Template literals: identity escapes that aren't valid ----
			{
				Code: "var foo = `\\\"`;",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \".`,
					Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: "var foo = `\"`;"},
						{MessageId: "escapeBackslash", Output: "var foo = `\\\\\"`;"},
					},
				}},
			},
			{
				Code: "var foo = `\\'`;",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \'.`,
					Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: "var foo = `'`;"},
						{MessageId: "escapeBackslash", Output: "var foo = `\\\\'`;"},
					},
				}},
			},
			{
				Code: "var foo = `\\#`;",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \#.`,
					Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: "var foo = `#`;"},
						{MessageId: "escapeBackslash", Output: "var foo = `\\\\#`;"},
					},
				}},
			},
			{
				Code: "var foo = '\\`foo\\`';",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryEscape",
						Message:   "Unnecessary escape character: \\`.",
						Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: "var foo = '`foo\\`';"},
							{MessageId: "escapeBackslash", Output: "var foo = '\\\\`foo\\`';"},
						},
					},
					{
						MessageId: "unnecessaryEscape",
						Message:   "Unnecessary escape character: \\`.",
						Line:      1, Column: 17, EndLine: 1, EndColumn: 18,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: "var foo = '\\`foo`';"},
							{MessageId: "escapeBackslash", Output: "var foo = '\\`foo\\\\`';"},
						},
					},
				},
			},
			{
				Code: "var foo = `\\\"${foo}\\\"`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryEscape",
						Message:   `Unnecessary escape character: \".`,
						Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: "var foo = `\"${foo}\\\"`;"},
							{MessageId: "escapeBackslash", Output: "var foo = `\\\\\"${foo}\\\"`;"},
						},
					},
					{
						MessageId: "unnecessaryEscape",
						Message:   `Unnecessary escape character: \".`,
						Line:      1, Column: 20, EndLine: 1, EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: "var foo = `\\\"${foo}\"`;"},
							{MessageId: "escapeBackslash", Output: "var foo = `\\\"${foo}\\\\\"`;"},
						},
					},
				},
			},
			{
				Code: "var foo = `\\'${foo}\\'`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryEscape",
						Message:   `Unnecessary escape character: \'.`,
						Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: "var foo = `'${foo}\\'`;"},
							{MessageId: "escapeBackslash", Output: "var foo = `\\\\'${foo}\\'`;"},
						},
					},
					{
						MessageId: "unnecessaryEscape",
						Message:   `Unnecessary escape character: \'.`,
						Line:      1, Column: 20, EndLine: 1, EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: "var foo = `\\'${foo}'`;"},
							{MessageId: "escapeBackslash", Output: "var foo = `\\'${foo}\\\\'`;"},
						},
					},
				},
			},
			{
				Code: "var foo = `\\#${foo}`;",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \#.`,
					Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: "var foo = `#${foo}`;"},
						{MessageId: "escapeBackslash", Output: "var foo = `\\\\#${foo}`;"},
					},
				}},
			},
			{
				Code: `let foo = '\ ';`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   "Unnecessary escape character: \\ .",
					Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `let foo = ' ';`},
						{MessageId: "escapeBackslash", Output: `let foo = '\\ ';`},
					},
				}},
			},
			{
				Code: `let foo = /\ /;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   "Unnecessary escape character: \\ .",
					Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `let foo = / /;`},
						{MessageId: "escapeBackslash", Output: `let foo = /\\ /;`},
					},
				}},
			},
			{
				Code: "var foo = `\\$\\{{${foo}`;",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \$.`,
					Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: "var foo = `$\\{{${foo}`;"},
						{MessageId: "escapeBackslash", Output: "var foo = `\\\\$\\{{${foo}`;"},
					},
				}},
			},
			{
				Code: "var foo = `\\$a${foo}`;",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \$.`,
					Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: "var foo = `$a${foo}`;"},
						{MessageId: "escapeBackslash", Output: "var foo = `\\\\$a${foo}`;"},
					},
				}},
			},
			{
				Code: "var foo = `a\\{{${foo}`;",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \{.`,
					Line:      1, Column: 13, EndLine: 1, EndColumn: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: "var foo = `a{{${foo}`;"},
						{MessageId: "escapeBackslash", Output: "var foo = `a\\\\{{${foo}`;"},
					},
				}},
			},

			// ---- Multi-line template literal: line/column track newlines ----
			{
				Code: "`multiline template\nliteral with useless \\escape`",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \e.`,
					Line:      2, Column: 22, EndLine: 2, EndColumn: 23,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: "`multiline template\nliteral with useless escape`"},
						{MessageId: "escapeBackslash", Output: "`multiline template\nliteral with useless \\\\escape`"},
					},
				}},
			},
			{
				Code: "`multiline template\r\nliteral with useless \\escape`",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \e.`,
					Line:      2, Column: 22, EndLine: 2, EndColumn: 23,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: "`multiline template\r\nliteral with useless escape`"},
						{MessageId: "escapeBackslash", Output: "`multiline template\r\nliteral with useless \\\\escape`"},
					},
				}},
			},
			{
				Code: "`template literal with line continuation \\\nand useless \\escape`",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \e.`,
					Line:      2, Column: 13, EndLine: 2, EndColumn: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: "`template literal with line continuation \\\nand useless escape`"},
						{MessageId: "escapeBackslash", Output: "`template literal with line continuation \\\nand useless \\\\escape`"},
					},
				}},
			},

			// ---- Regex character class ----
			{
				Code: `var foo = /[ab\-]/`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \-.`,
					Line:      1, Column: 15, EndLine: 1, EndColumn: 16,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = /[ab-]/`},
						{MessageId: "escapeBackslash", Output: `var foo = /[ab\\-]/`},
					},
				}},
			},
			{
				Code: `var foo = /[\-ab]/`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \-.`,
					Line:      1, Column: 13, EndLine: 1, EndColumn: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = /[-ab]/`},
						{MessageId: "escapeBackslash", Output: `var foo = /[\\-ab]/`},
					},
				}},
			},
			{
				Code: `var foo = /[ab\?]/`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 15, EndLine: 1, EndColumn: 16,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = /[ab?]/`},
						{MessageId: "escapeBackslash", Output: `var foo = /[ab\\?]/`},
					},
				}},
			},
			{
				Code: `var foo = /[ab\.]/`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 15, EndLine: 1, EndColumn: 16,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = /[ab.]/`},
						{MessageId: "escapeBackslash", Output: `var foo = /[ab\\.]/`},
					},
				}},
			},
			{
				Code: `var foo = /[a\|b]/`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 14, EndLine: 1, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = /[a|b]/`},
						{MessageId: "escapeBackslash", Output: `var foo = /[a\\|b]/`},
					},
				}},
			},
			{
				Code: `var foo = /\-/`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = /-/`},
						{MessageId: "escapeBackslash", Output: `var foo = /\\-/`},
					},
				}},
			},
			{
				Code: `var foo = /[\-]/`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 13, EndLine: 1, EndColumn: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = /[-]/`},
						{MessageId: "escapeBackslash", Output: `var foo = /[\\-]/`},
					},
				}},
			},
			{
				Code: `var foo = /[ab\$]/`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 15, EndLine: 1, EndColumn: 16,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = /[ab$]/`},
						{MessageId: "escapeBackslash", Output: `var foo = /[ab\\$]/`},
					},
				}},
			},
			{
				Code: `var foo = /[\(paren]/`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 13, EndLine: 1, EndColumn: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = /[(paren]/`},
						{MessageId: "escapeBackslash", Output: `var foo = /[\\(paren]/`},
					},
				}},
			},
			{
				Code: `var foo = /[\[]/`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 13, EndLine: 1, EndColumn: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = /[[]/`},
						{MessageId: "escapeBackslash", Output: `var foo = /[\\[]/`},
					},
				}},
			},
			{
				Code: `var foo = /[\/]/`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 13, EndLine: 1, EndColumn: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = /[/]/`},
						{MessageId: "escapeBackslash", Output: `var foo = /[\\/]/`},
					},
				}},
			},
			{
				Code: `var foo = /[\B]/`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 13, EndLine: 1, EndColumn: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = /[B]/`},
						{MessageId: "escapeBackslash", Output: `var foo = /[\\B]/`},
					},
				}},
			},
			{
				Code: `var foo = /[a][\-b]/`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 16, EndLine: 1, EndColumn: 17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = /[a][-b]/`},
						{MessageId: "escapeBackslash", Output: `var foo = /[a][\\-b]/`},
					},
				}},
			},
			{
				Code: `var foo = /\-[]/`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = /-[]/`},
						{MessageId: "escapeBackslash", Output: `var foo = /\\-[]/`},
					},
				}},
			},
			{
				Code: `var foo = /[a\^]/`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 14, EndLine: 1, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = /[a^]/`},
						{MessageId: "escapeBackslash", Output: `var foo = /[a\\^]/`},
					},
				}},
			},

			// ---- Caret in negated class ----
			{
				Code: `/[^\^]/`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: "/[^^]/"},
						{MessageId: "escapeBackslash", Output: `/[^\\^]/`},
					},
				}},
			},
			{
				Code: `/[^\^]/u`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: "/[^^]/u"},
						{MessageId: "escapeBackslash", Output: `/[^\\^]/u`},
					},
				}},
			},

			// ---- Directive prologue: removeEscapeDoNotKeepSemantics suggestion ----
			{
				Code: `"use\ strict";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   "Unnecessary escape character: \\ .",
					Line:      1, Column: 5, EndLine: 1, EndColumn: 6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscapeDoNotKeepSemantics", Output: `"use strict";`},
						{MessageId: "escapeBackslash", Output: `"use\\ strict";`},
					},
				}},
			},
			{
				Code: `({ foo() { "foo"; "bar"; "ba\z" } })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \z.`,
					Line:      1, Column: 29, EndLine: 1, EndColumn: 30,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscapeDoNotKeepSemantics", Output: `({ foo() { "foo"; "bar"; "baz" } })`},
						{MessageId: "escapeBackslash", Output: `({ foo() { "foo"; "bar"; "ba\\z" } })`},
					},
				}},
			},

			// ---- ES2024 v-mode reserved-double-punctuator escapes ----
			// Each leading `\X` is at the start of the class so neither neighbour
			// nor double-punctuator exemption applies — flagged with both
			// suggestions.
			{
				Code: `/[\$]/v`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 3, EndLine: 1, EndColumn: 4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: "/[$]/v"},
						{MessageId: "escapeBackslash", Output: `/[\\$]/v`},
					},
				}},
			},
			{
				Code: `/[\&\&]/v`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \&.`,
					Line:      1, Column: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `/[&\&]/v`},
						{MessageId: "escapeBackslash", Output: `/[\\&\&]/v`},
					},
				}},
			},
			{
				Code: `/[\!\!]/v`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `/[!\!]/v`},
						{MessageId: "escapeBackslash", Output: `/[\\!\!]/v`},
					},
				}},
			},
			{
				Code: `/[\#\#]/v`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `/[#\#]/v`},
						{MessageId: "escapeBackslash", Output: `/[\\#\#]/v`},
					},
				}},
			},
			{
				Code: `/[\%\%]/v`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `/[%\%]/v`},
						{MessageId: "escapeBackslash", Output: `/[\\%\%]/v`},
					},
				}},
			},
			{
				Code: `/[\*\*]/v`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `/[*\*]/v`},
						{MessageId: "escapeBackslash", Output: `/[\\*\*]/v`},
					},
				}},
			},
			{
				Code: `/[\+\+]/v`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `/[+\+]/v`},
						{MessageId: "escapeBackslash", Output: `/[\\+\+]/v`},
					},
				}},
			},

			// ---- v-mode set-operation escapes: only `removeEscape`, no `escapeBackslash` ----
			{
				Code: `/[\p{ASCII}--\.]/v`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \..`,
					Line:      1, Column: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `/[\p{ASCII}--.]/v`},
					},
				}},
			},
			{
				Code: `/[\p{ASCII}&&\.]/v`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `/[\p{ASCII}&&.]/v`},
					},
				}},
			},
			{
				Code: `/[\.--[.&]]/v`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `/[.--[.&]]/v`},
					},
				}},
			},
			{
				Code: `/[\.&&[.&]]/v`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `/[.&&[.&]]/v`},
					},
				}},
			},
			{
				Code: `/[\.--\.--\.]/v`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryEscape",
						Line:      1, Column: 3,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: `/[.--\.--\.]/v`},
						},
					},
					{
						MessageId: "unnecessaryEscape",
						Line:      1, Column: 7,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: `/[\.--.--\.]/v`},
						},
					},
					{
						MessageId: "unnecessaryEscape",
						Line:      1, Column: 11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: `/[\.--\.--.]/v`},
						},
					},
				},
			},
			{
				Code: `/[\.&&\.&&\.]/v`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryEscape",
						Line:      1, Column: 3,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: `/[.&&\.&&\.]/v`},
						},
					},
					{
						MessageId: "unnecessaryEscape",
						Line:      1, Column: 7,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: `/[\.&&.&&\.]/v`},
						},
					},
					{
						MessageId: "unnecessaryEscape",
						Line:      1, Column: 11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: `/[\.&&\.&&.]/v`},
						},
					},
				},
			},
			// Nested classes — escapes are inside [.&], not directly in the
			// outer set-operation class, so escapeBackslash IS suggested.
			{
				Code: `/[[\.&]--[\.&]]/v`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryEscape",
						Line:      1, Column: 4,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: `/[[.&]--[\.&]]/v`},
							{MessageId: "escapeBackslash", Output: `/[[\\.&]--[\.&]]/v`},
						},
					},
					{
						MessageId: "unnecessaryEscape",
						Line:      1, Column: 11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: `/[[\.&]--[.&]]/v`},
							{MessageId: "escapeBackslash", Output: `/[[\.&]--[\\.&]]/v`},
						},
					},
				},
			},
			{
				Code: `/[[\.&]&&[\.&]]/v`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryEscape",
						Line:      1, Column: 4,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: `/[[.&]&&[\.&]]/v`},
							{MessageId: "escapeBackslash", Output: `/[[\\.&]&&[\.&]]/v`},
						},
					},
					{
						MessageId: "unnecessaryEscape",
						Line:      1, Column: 11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: `/[[\.&]&&[.&]]/v`},
							{MessageId: "escapeBackslash", Output: `/[[\.&]&&[\\.&]]/v`},
						},
					},
				},
			},

			// ---- Option allowRegexCharacters: only allows specific chars ----
			{
				Code:    `var foo = "\#/";`,
				Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"#"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \#.`,
					Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = "#/";`},
						{MessageId: "escapeBackslash", Output: `var foo = "\\#/";`},
					},
				}},
			},
			{
				Code:    `var foo = /\#\@/;`,
				Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"#"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \@.`,
					Line:      1, Column: 14, EndLine: 1, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = /\#@/;`},
						{MessageId: "escapeBackslash", Output: `var foo = /\#\\@/;`},
					},
				}},
			},
			{
				Code:    `var foo = /[a\@b]/`,
				Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"#"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \@.`,
					Line:      1, Column: 14, EndLine: 1, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = /[a@b]/`},
						{MessageId: "escapeBackslash", Output: `var foo = /[a\\@b]/`},
					},
				}},
			},
			{
				Code:    `/[\@\@]/v`,
				Options: map[string]interface{}{"allowRegexCharacters": []interface{}{"#"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `/[@\@]/v`},
						{MessageId: "escapeBackslash", Output: `/[\\@\@]/v`},
					},
				}},
			},

			// =================================================================
			// rslint-specific extensions beyond the upstream test suite. The
			// goal is to lock in tsgo / Go-port edge cases that the ESLint
			// suite never exercises.
			// =================================================================

			// ---- Multi-span template literals: each span reports independently ----
			{
				Code: "var foo = `\\#${a}\\@${b}\\!`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryEscape",
						Message:   `Unnecessary escape character: \#.`,
						Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: "var foo = `#${a}\\@${b}\\!`;"},
							{MessageId: "escapeBackslash", Output: "var foo = `\\\\#${a}\\@${b}\\!`;"},
						},
					},
					{
						MessageId: "unnecessaryEscape",
						Message:   `Unnecessary escape character: \@.`,
						Line:      1, Column: 18, EndLine: 1, EndColumn: 19,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: "var foo = `\\#${a}@${b}\\!`;"},
							{MessageId: "escapeBackslash", Output: "var foo = `\\#${a}\\\\@${b}\\!`;"},
						},
					},
					{
						MessageId: "unnecessaryEscape",
						Message:   `Unnecessary escape character: \!.`,
						Line:      1, Column: 24, EndLine: 1, EndColumn: 25,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: "var foo = `\\#${a}\\@${b}!`;"},
							{MessageId: "escapeBackslash", Output: "var foo = `\\#${a}\\@${b}\\\\!`;"},
						},
					},
				},
			},

			// ---- TS surface around string literals — non-null, parens, as ----
			{
				Code: `var foo = ("\#")!;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \#.`,
					Line:      1, Column: 13, EndLine: 1, EndColumn: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = ("#")!;`},
						{MessageId: "escapeBackslash", Output: `var foo = ("\\#")!;`},
					},
				}},
			},
			{
				Code: `var foo = "\@" as string;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \@.`,
					Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = "@" as string;`},
						{MessageId: "escapeBackslash", Output: `var foo = "\\@" as string;`},
					},
				}},
			},
			{
				Code: `var foo = "\@" satisfies string;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = "@" satisfies string;`},
						{MessageId: "escapeBackslash", Output: `var foo = "\\@" satisfies string;`},
					},
				}},
			},

			// ---- String literals in real TS contexts ----
			{
				Code: `import foo from "./a\#";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 21, EndLine: 1, EndColumn: 22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `import foo from "./a#";`},
						{MessageId: "escapeBackslash", Output: `import foo from "./a\\#";`},
					},
				}},
			},
			{
				Code: `enum E { A = "\#" }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 15, EndLine: 1, EndColumn: 16,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `enum E { A = "#" }`},
						{MessageId: "escapeBackslash", Output: `enum E { A = "\\#" }`},
					},
				}},
			},
			{
				Code: `type T = "\#";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 11, EndLine: 1, EndColumn: 12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `type T = "#";`},
						{MessageId: "escapeBackslash", Output: `type T = "\\#";`},
					},
				}},
			},
			{
				Code: `const obj = { ["\#"]: 1 };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 17, EndLine: 1, EndColumn: 18,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `const obj = { ["#"]: 1 };`},
						{MessageId: "escapeBackslash", Output: `const obj = { ["\\#"]: 1 };`},
					},
				}},
			},

			// ---- Function-body directive prologue (use removeEscapeDoNotKeepSemantics) ----
			{
				Code: `function f() { "use\ strict"; return 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 20, EndLine: 1, EndColumn: 21,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscapeDoNotKeepSemantics", Output: `function f() { "use strict"; return 1; }`},
						{MessageId: "escapeBackslash", Output: `function f() { "use\\ strict"; return 1; }`},
					},
				}},
			},
			// String literal in expression position (NOT a directive) → standard removeEscape.
			{
				Code: `var x = "\@"; "ba\z";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryEscape",
						Line:      1, Column: 10, EndLine: 1, EndColumn: 11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: `var x = "@"; "ba\z";`},
							{MessageId: "escapeBackslash", Output: `var x = "\\@"; "ba\z";`},
						},
					},
					// `"ba\z";` is the second statement, NOT a prologue — uses removeEscape, not the DoNotKeepSemantics variant.
					{
						MessageId: "unnecessaryEscape",
						Line:      1, Column: 18, EndLine: 1, EndColumn: 19,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: `var x = "\@"; "baz";`},
							{MessageId: "escapeBackslash", Output: `var x = "\@"; "ba\\z";`},
						},
					},
				},
			},

			// ---- Surrogate-pair / multi-byte preceding the escape (column counted in UTF-16) ----
			// 👍 occupies 2 UTF-16 code units (cols 12-13); the `\#` lands at column 14.
			{
				Code: "var foo = '👍\\#';",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 14, EndLine: 1, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: "var foo = '👍#';"},
						{MessageId: "escapeBackslash", Output: "var foo = '👍\\\\#';"},
					},
				}},
			},

			// ---- Multi-line regex-equivalent (multi-line content via continuations / line 2 reports) ----
			{
				Code: "var foo = '\\\nbar\\@';",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      2, Column: 4, EndLine: 2, EndColumn: 5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: "var foo = '\\\nbar@';"},
						{MessageId: "escapeBackslash", Output: "var foo = '\\\nbar\\\\@';"},
					},
				}},
			},

			// ---- v-mode triple-nested classes — escape directly under outer set-op ----
			// `[\.&&[a&&[b]]]` — \. is at top of outer class which has &&; no escapeBackslash.
			{
				Code: `/[\.&&[a&&[b]]]/v`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 3, EndLine: 1, EndColumn: 4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `/[.&&[a&&[b]]]/v`},
					},
				}},
			},
			// Escape three levels deep inside a class without set ops — has escapeBackslash.
			{
				Code: `/[a&&[b&&[\.]]]/v`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 11, EndLine: 1, EndColumn: 12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `/[a&&[b&&[.]]]/v`},
						{MessageId: "escapeBackslash", Output: `/[a&&[b&&[\\.]]]/v`},
					},
				}},
			},

			// ---- v-mode `\^` adjacent-to-negate-caret edge ----
			// Before-caret: `\^` IS the negate; second `^` is literal — only the
			// inner `\^` should fire if it's after the negate but doubled with
			// another `^` after.
			// Already covered by /[^\^^]/v (valid) and /[^\^]/v (invalid in v).
			// Add: nested negate class with `\^` not at start.
			{
				Code: `/[^a\^]/v`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 5, EndLine: 1, EndColumn: 6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `/[^a^]/v`},
						{MessageId: "escapeBackslash", Output: `/[^a\\^]/v`},
					},
				}},
			},

			// ---- Multi-byte UTF-8 escapes inside string ----
			// `\é` (multi-byte X) — flagged with the multi-byte text in the message.
			{
				Code: `var foo = "\é";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   "Unnecessary escape character: \\é.",
					Line:      1, Column: 12, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = "é";`},
						{MessageId: "escapeBackslash", Output: `var foo = "\\é";`},
					},
				}},
			},

			// ---- Multiple escapes interleaved with valid escapes in regex ----
			{
				Code: `var foo = /\d\@\w\#\s/;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryEscape",
						Message:   `Unnecessary escape character: \@.`,
						Line:      1, Column: 14, EndLine: 1, EndColumn: 15,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: `var foo = /\d@\w\#\s/;`},
							{MessageId: "escapeBackslash", Output: `var foo = /\d\\@\w\#\s/;`},
						},
					},
					{
						MessageId: "unnecessaryEscape",
						Message:   `Unnecessary escape character: \#.`,
						Line:      1, Column: 18, EndLine: 1, EndColumn: 19,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: `var foo = /\d\@\w#\s/;`},
							{MessageId: "escapeBackslash", Output: `var foo = /\d\@\w\\#\s/;`},
						},
					},
				},
			},

			// ---- Regex with leading/trailing valid escapes guarding an invalid one ----
			{
				Code: `var foo = /\b\@\b/;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 14, EndLine: 1, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var foo = /\b@\b/;`},
						{MessageId: "escapeBackslash", Output: `var foo = /\b\\@\b/;`},
					},
				}},
			},

			// ---- Directive-container precision: only function-body Blocks count ----
			// Inside `if (true) { "ba\z"; }` — `{ }` is a plain Block, NOT a directive
			// container. The string is a regular expression statement, so the
			// remove-escape suggestion should be the standard `removeEscape`,
			// NOT `removeEscapeDoNotKeepSemantics`.
			{
				Code: `if (true) { "ba\z"; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 16, EndLine: 1, EndColumn: 17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `if (true) { "baz"; }`},
						{MessageId: "escapeBackslash", Output: `if (true) { "ba\\z"; }`},
					},
				}},
			},
			{
				Code: `for (;;) { "ba\z"; break; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 15, EndLine: 1, EndColumn: 16,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `for (;;) { "baz"; break; }`},
						{MessageId: "escapeBackslash", Output: `for (;;) { "ba\\z"; break; }`},
					},
				}},
			},
			// Plain Block as a statement — also NOT a directive container.
			{
				Code: `{ "ba\z"; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 6, EndLine: 1, EndColumn: 7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `{ "baz"; }`},
						{MessageId: "escapeBackslash", Output: `{ "ba\\z"; }`},
					},
				}},
			},
			// Class static block IS a directive container.
			{
				Code: `class C { static { "use\ strict"; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 24, EndLine: 1, EndColumn: 25,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscapeDoNotKeepSemantics", Output: `class C { static { "use strict"; } }`},
						{MessageId: "escapeBackslash", Output: `class C { static { "use\\ strict"; } }`},
					},
				}},
			},
			// Arrow-function body IS a function body → directive container.
			{
				Code: `var f = () => { "use\ strict"; return 1; };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 21, EndLine: 1, EndColumn: 22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscapeDoNotKeepSemantics", Output: `var f = () => { "use strict"; return 1; };`},
						{MessageId: "escapeBackslash", Output: `var f = () => { "use\\ strict"; return 1; };`},
					},
				}},
			},

			// ---- TSX: probe whether StringLiteral CAN appear under JsxElement /
			// JsxFragment as direct children. If it can, our defensive skip
			// branches matter; if it can't, they're dead code (matching upstream's
			// defensive structural check). The cases below assert the actually
			// reachable JSX shapes.

			// JsxAttribute value (StringLiteral child of JsxAttribute) — SKIPPED.
			// (No error expected; this case lives in the valid section above as
			// `<div attr="\d" />`. Repeated here as documentation.)

			// StringLiteral inside `{…}` — parent is JsxExpression, NOT
			// JsxAttribute / JsxElement / JsxFragment — should be FLAGGED.
			{
				Code: `<foo attr={"\#"}/>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 13, EndLine: 1, EndColumn: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `<foo attr={"#"}/>`},
						{MessageId: "escapeBackslash", Output: `<foo attr={"\\#"}/>`},
					},
				}},
			},
			// StringLiteral inside JsxExpression in element body — same: parent
			// is JsxExpression. Confirms the JsxElement-direct-child case is
			// unreachable in tsgo (StringLiteral is wrapped in JsxExpression).
			{
				Code: `var x = <div>{"\@"}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Line:      1, Column: 16, EndLine: 1, EndColumn: 17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `var x = <div>{"@"}</div>`},
						{MessageId: "escapeBackslash", Output: `var x = <div>{"\\@"}</div>`},
					},
				}},
			},

			// ---- Non-v mode: literal `[` inside a class is NOT a nested
			// class start; the next `]` still closes the class. Caught from a
			// real-world rsbuild/css-loader bundled regex that we initially
			// missed. ----
			{
				Code: "var p = /[ -,.\\/:-@[\\]\\^`{-~]/;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryEscape",
						Message:   `Unnecessary escape character: \/.`,
						Line:      1, Column: 15, EndLine: 1, EndColumn: 16,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: "var p = /[ -,./:-@[\\]\\^`{-~]/;"},
							{MessageId: "escapeBackslash", Output: "var p = /[ -,.\\\\/:-@[\\]\\^`{-~]/;"},
						},
					},
					{
						MessageId: "unnecessaryEscape",
						Message:   `Unnecessary escape character: \^.`,
						Line:      1, Column: 23, EndLine: 1, EndColumn: 24,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: "var p = /[ -,.\\/:-@[\\]^`{-~]/;"},
							{MessageId: "escapeBackslash", Output: "var p = /[ -,.\\/:-@[\\]\\\\^`{-~]/;"},
						},
					},
				},
			},

			// ---- v-mode `\q{…}` body: nested escapes are flagged ----
			// `\.` inside `\q{a\.b}` is a Character node in regexpp's AST, so
			// upstream flags it. Our walker must too.
			{
				Code: `/[\q{a\.b}]/v`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unnecessaryEscape",
					Message:   `Unnecessary escape character: \..`,
					Line:      1, Column: 7, EndLine: 1, EndColumn: 8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "removeEscape", Output: `/[\q{a.b}]/v`},
						{MessageId: "escapeBackslash", Output: `/[\q{a\\.b}]/v`},
					},
				}},
			},
			// Multiple alternatives, multiple flags.
			{
				Code: `/[\q{\@a|\#b}]/v`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryEscape",
						Message:   `Unnecessary escape character: \@.`,
						Line:      1, Column: 6, EndLine: 1, EndColumn: 7,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: `/[\q{@a|\#b}]/v`},
							{MessageId: "escapeBackslash", Output: `/[\q{\\@a|\#b}]/v`},
						},
					},
					{
						MessageId: "unnecessaryEscape",
						Message:   `Unnecessary escape character: \#.`,
						Line:      1, Column: 10, EndLine: 1, EndColumn: 11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeEscape", Output: `/[\q{\@a|#b}]/v`},
							{MessageId: "escapeBackslash", Output: `/[\q{\@a|\\#b}]/v`},
						},
					},
				},
			},
		},
	)
}

// =============================================================================
// Hardening tests for the regex pattern parser. These cover malformed-input
// skip behavior (matching ESLint's regexpp try/catch) and boundary conditions
// that aren't surfaced through the rule_tester DSL because TS rejects the
// regex at parse time. We exercise the parser as a pure function below.
// =============================================================================

func TestPatternParses(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		flagStr string
		want    bool
	}{
		// Well-formed patterns — should parse.
		{name: "empty", pattern: "", flagStr: "", want: true},
		{name: "literal", pattern: "abc", flagStr: "", want: true},
		{name: "simple class", pattern: "[abc]", flagStr: "", want: true},
		{name: "negated class", pattern: "[^abc]", flagStr: "", want: true},
		{name: "v-mode nested classes", pattern: "[[a]&&[b]]", flagStr: "v", want: true},
		{name: "u-mode property escape", pattern: "\\p{ASCII}", flagStr: "u", want: true},
		{name: "v-mode q-string", pattern: "[\\q{abc}]", flagStr: "v", want: true},
		{name: "v-mode q-string with literal ]", pattern: "[\\q{a]b}c]", flagStr: "v", want: true},
		{name: "named backreference", pattern: "(?<a>x)\\k<a>", flagStr: "", want: true},
		{name: "literal-]-outside-class", pattern: "abc]", flagStr: "", want: true},

		// Malformed — should skip.
		{name: "unterminated class", pattern: "[abc", flagStr: "", want: false},
		{name: "trailing backslash", pattern: "abc\\", flagStr: "", want: false},
		{name: "unterminated u-brace under uvMode", pattern: "\\u{1F600", flagStr: "u", want: false},
		{name: "unterminated p-brace", pattern: "\\p{ASCI", flagStr: "u", want: false},
		{name: "unterminated q-brace", pattern: "[\\q{abc", flagStr: "v", want: false},
		{name: "unterminated k-name", pattern: "\\k<name", flagStr: "", want: false},
		// Note: `\u{` in non-uvMode is identity-escape `u`, NOT an error.
		{name: "u-brace in non-uvMode is fine", pattern: "\\u{1F600}", flagStr: "", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := parseFlagsFromStr(tt.flagStr)
			got := patternParses(tt.pattern, flags)
			if got != tt.want {
				t.Errorf("patternParses(%q, %q) = %v, want %v", tt.pattern, tt.flagStr, got, tt.want)
			}
		})
	}
}

func TestPreScanClass(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		start       int
		flagStr     string
		wantNegate  bool
		wantSetOp   bool
		wantEnd     int
	}{
		{name: "simple", pattern: "[abc]", start: 0, flagStr: "", wantEnd: 5},
		{name: "negated", pattern: "[^abc]", start: 0, flagStr: "", wantNegate: true, wantEnd: 6},
		{name: "v-mode no set op", pattern: "[abc]", start: 0, flagStr: "v", wantEnd: 5},
		{name: "v-mode subtract", pattern: "[\\.--\\.]", start: 0, flagStr: "v", wantSetOp: true, wantEnd: 8},
		{name: "v-mode intersect", pattern: "[a&&b]", start: 0, flagStr: "v", wantSetOp: true, wantEnd: 6},
		// Nested set-op only counts at the OUTER class's level.
		{name: "v-mode set-op nested only", pattern: "[a[b&&c]]", start: 0, flagStr: "v", wantEnd: 9},
		{name: "v-mode set-op outer with nested", pattern: "[a&&[b]]", start: 0, flagStr: "v", wantSetOp: true, wantEnd: 8},
		// Escaped `]` should not close the class early.
		{name: "escaped ] in class", pattern: "[\\]a]", start: 0, flagStr: "", wantEnd: 5},
		// `\q{...}` in v-mode contains a `]` literal; class shouldn't close inside.
		{name: "q-string inside class", pattern: "[\\q{a]b}c]", start: 0, flagStr: "v", wantEnd: 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := parseFlagsFromStr(tt.flagStr)
			frame := preScanClass(tt.pattern, tt.start, flags)
			if frame.negate != tt.wantNegate {
				t.Errorf("negate = %v, want %v", frame.negate, tt.wantNegate)
			}
			if frame.hasSetOp != tt.wantSetOp {
				t.Errorf("hasSetOp = %v, want %v", frame.hasSetOp, tt.wantSetOp)
			}
			if frame.end != tt.wantEnd {
				t.Errorf("end = %d, want %d", frame.end, tt.wantEnd)
			}
		})
	}
}

func parseFlagsFromStr(s string) utils.RegexFlags {
	return utils.ParseRegexFlags(s)
}
