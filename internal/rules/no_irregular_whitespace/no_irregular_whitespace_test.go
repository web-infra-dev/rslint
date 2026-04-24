package no_irregular_whitespace

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoIrregularWhitespace(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoIrregularWhitespaceRule,
		[]rule_tester.ValidTestCase{
			// ---- Escaped Unicode in strings (no actual irregular chars) ----
			{Code: `'\u000B';`},
			{Code: `'\u000C';`},
			{Code: `'\u0085';`},
			{Code: `'\u00A0';`},
			{Code: `'\u1680';`},
			{Code: `'\u180E';`},
			{Code: `'\ufeff';`},
			{Code: `'\u2000';`},
			{Code: `'\u2001';`},
			{Code: `'\u2002';`},
			{Code: `'\u2003';`},
			{Code: `'\u2004';`},
			{Code: `'\u2005';`},
			{Code: `'\u2006';`},
			{Code: `'\u2007';`},
			{Code: `'\u2008';`},
			{Code: `'\u2009';`},
			{Code: `'\u200A';`},
			{Code: `'\u200B';`},
			{Code: `'\u2028';`},
			{Code: `'\u2029';`},
			{Code: `'\u202F';`},
			{Code: `'\u205f';`},
			{Code: `'\u3000';`},

			// ---- Actual irregular whitespace inside strings (skipStrings default true) ----
			{Code: "'\u000B';"},
			{Code: "'\u000C';"},
			{Code: "'\u0085';"},
			{Code: "'\u00A0';"},
			{Code: "'\u1680';"},
			{Code: "'\u180E';"},
			{Code: "'\uFEFF';"},
			{Code: "'\u2000';"},
			{Code: "'\u2001';"},
			{Code: "'\u2002';"},
			{Code: "'\u2003';"},
			{Code: "'\u2004';"},
			{Code: "'\u2005';"},
			{Code: "'\u2006';"},
			{Code: "'\u2007';"},
			{Code: "'\u2008';"},
			{Code: "'\u2009';"},
			{Code: "'\u200A';"},
			{Code: "'\u200B';"},
			// Multiline strings with escaped backslash + line separator
			{Code: "'\\\u2028';"},
			{Code: "'\\\u2029';"},
			{Code: "'\u202F';"},
			{Code: "'\u205F';"},
			{Code: "'\u3000';"},

			// ---- skipComments: true (single-line) ----
			{Code: "// \u000B", Options: map[string]interface{}{"skipComments": true}},
			{Code: "// \u000C", Options: map[string]interface{}{"skipComments": true}},
			{Code: "// \u0085", Options: map[string]interface{}{"skipComments": true}},
			{Code: "// \u00A0", Options: map[string]interface{}{"skipComments": true}},
			{Code: "// \u1680", Options: map[string]interface{}{"skipComments": true}},
			{Code: "// \u180E", Options: map[string]interface{}{"skipComments": true}},
			{Code: "// \uFEFF", Options: map[string]interface{}{"skipComments": true}},
			{Code: "// \u2000", Options: map[string]interface{}{"skipComments": true}},
			{Code: "// \u2001", Options: map[string]interface{}{"skipComments": true}},
			{Code: "// \u2002", Options: map[string]interface{}{"skipComments": true}},
			{Code: "// \u2003", Options: map[string]interface{}{"skipComments": true}},
			{Code: "// \u2004", Options: map[string]interface{}{"skipComments": true}},
			{Code: "// \u2005", Options: map[string]interface{}{"skipComments": true}},
			{Code: "// \u2006", Options: map[string]interface{}{"skipComments": true}},
			{Code: "// \u2007", Options: map[string]interface{}{"skipComments": true}},
			{Code: "// \u2008", Options: map[string]interface{}{"skipComments": true}},
			{Code: "// \u2009", Options: map[string]interface{}{"skipComments": true}},
			{Code: "// \u200A", Options: map[string]interface{}{"skipComments": true}},
			{Code: "// \u200B", Options: map[string]interface{}{"skipComments": true}},
			{Code: "// \u202F", Options: map[string]interface{}{"skipComments": true}},
			{Code: "// \u205F", Options: map[string]interface{}{"skipComments": true}},
			{Code: "// \u3000", Options: map[string]interface{}{"skipComments": true}},

			// ---- skipComments: true (block) ----
			{Code: "/* \u000B */", Options: map[string]interface{}{"skipComments": true}},
			{Code: "/* \u000C */", Options: map[string]interface{}{"skipComments": true}},
			{Code: "/* \u0085 */", Options: map[string]interface{}{"skipComments": true}},
			{Code: "/* \u00A0 */", Options: map[string]interface{}{"skipComments": true}},
			{Code: "/* \u1680 */", Options: map[string]interface{}{"skipComments": true}},
			{Code: "/* \u180E */", Options: map[string]interface{}{"skipComments": true}},
			{Code: "/* \uFEFF */", Options: map[string]interface{}{"skipComments": true}},
			{Code: "/* \u2000 */", Options: map[string]interface{}{"skipComments": true}},
			{Code: "/* \u2001 */", Options: map[string]interface{}{"skipComments": true}},
			{Code: "/* \u2002 */", Options: map[string]interface{}{"skipComments": true}},
			{Code: "/* \u2003 */", Options: map[string]interface{}{"skipComments": true}},
			{Code: "/* \u2004 */", Options: map[string]interface{}{"skipComments": true}},
			{Code: "/* \u2005 */", Options: map[string]interface{}{"skipComments": true}},
			{Code: "/* \u2006 */", Options: map[string]interface{}{"skipComments": true}},
			{Code: "/* \u2007 */", Options: map[string]interface{}{"skipComments": true}},
			{Code: "/* \u2008 */", Options: map[string]interface{}{"skipComments": true}},
			{Code: "/* \u2009 */", Options: map[string]interface{}{"skipComments": true}},
			{Code: "/* \u200A */", Options: map[string]interface{}{"skipComments": true}},
			{Code: "/* \u200B */", Options: map[string]interface{}{"skipComments": true}},
			{Code: "/* \u2028 */", Options: map[string]interface{}{"skipComments": true}},
			{Code: "/* \u2029 */", Options: map[string]interface{}{"skipComments": true}},
			{Code: "/* \u202F */", Options: map[string]interface{}{"skipComments": true}},
			{Code: "/* \u205F */", Options: map[string]interface{}{"skipComments": true}},
			{Code: "/* \u3000 */", Options: map[string]interface{}{"skipComments": true}},

			// ---- skipRegExps: true ----
			{Code: "/\u000B/", Options: map[string]interface{}{"skipRegExps": true}},
			{Code: "/\u000C/", Options: map[string]interface{}{"skipRegExps": true}},
			{Code: "/\u0085/", Options: map[string]interface{}{"skipRegExps": true}},
			{Code: "/\u00A0/", Options: map[string]interface{}{"skipRegExps": true}},
			{Code: "/\u1680/", Options: map[string]interface{}{"skipRegExps": true}},
			{Code: "/\u180E/", Options: map[string]interface{}{"skipRegExps": true}},
			{Code: "/\uFEFF/", Options: map[string]interface{}{"skipRegExps": true}},
			{Code: "/\u2000/", Options: map[string]interface{}{"skipRegExps": true}},
			{Code: "/\u2001/", Options: map[string]interface{}{"skipRegExps": true}},
			{Code: "/\u2002/", Options: map[string]interface{}{"skipRegExps": true}},
			{Code: "/\u2003/", Options: map[string]interface{}{"skipRegExps": true}},
			{Code: "/\u2004/", Options: map[string]interface{}{"skipRegExps": true}},
			{Code: "/\u2005/", Options: map[string]interface{}{"skipRegExps": true}},
			{Code: "/\u2006/", Options: map[string]interface{}{"skipRegExps": true}},
			{Code: "/\u2007/", Options: map[string]interface{}{"skipRegExps": true}},
			{Code: "/\u2008/", Options: map[string]interface{}{"skipRegExps": true}},
			{Code: "/\u2009/", Options: map[string]interface{}{"skipRegExps": true}},
			{Code: "/\u200A/", Options: map[string]interface{}{"skipRegExps": true}},
			{Code: "/\u200B/", Options: map[string]interface{}{"skipRegExps": true}},
			{Code: "/\u202F/", Options: map[string]interface{}{"skipRegExps": true}},
			{Code: "/\u205F/", Options: map[string]interface{}{"skipRegExps": true}},
			{Code: "/\u3000/", Options: map[string]interface{}{"skipRegExps": true}},

			// ---- skipTemplates: true ----
			{Code: "`\u000B`", Options: map[string]interface{}{"skipTemplates": true}},
			{Code: "`\u000C`", Options: map[string]interface{}{"skipTemplates": true}},
			{Code: "`\u0085`", Options: map[string]interface{}{"skipTemplates": true}},
			{Code: "`\u00A0`", Options: map[string]interface{}{"skipTemplates": true}},
			{Code: "`\u1680`", Options: map[string]interface{}{"skipTemplates": true}},
			{Code: "`\u180E`", Options: map[string]interface{}{"skipTemplates": true}},
			{Code: "`\uFEFF`", Options: map[string]interface{}{"skipTemplates": true}},
			{Code: "`\u2000`", Options: map[string]interface{}{"skipTemplates": true}},
			{Code: "`\u2001`", Options: map[string]interface{}{"skipTemplates": true}},
			{Code: "`\u2002`", Options: map[string]interface{}{"skipTemplates": true}},
			{Code: "`\u2003`", Options: map[string]interface{}{"skipTemplates": true}},
			{Code: "`\u2004`", Options: map[string]interface{}{"skipTemplates": true}},
			{Code: "`\u2005`", Options: map[string]interface{}{"skipTemplates": true}},
			{Code: "`\u2006`", Options: map[string]interface{}{"skipTemplates": true}},
			{Code: "`\u2007`", Options: map[string]interface{}{"skipTemplates": true}},
			{Code: "`\u2008`", Options: map[string]interface{}{"skipTemplates": true}},
			{Code: "`\u2009`", Options: map[string]interface{}{"skipTemplates": true}},
			{Code: "`\u200A`", Options: map[string]interface{}{"skipTemplates": true}},
			{Code: "`\u200B`", Options: map[string]interface{}{"skipTemplates": true}},
			{Code: "`\u202F`", Options: map[string]interface{}{"skipTemplates": true}},
			{Code: "`\u205F`", Options: map[string]interface{}{"skipTemplates": true}},
			{Code: "`\u3000`", Options: map[string]interface{}{"skipTemplates": true}},
			// Template with expression
			{Code: "`\u3000${foo}\u3000`", Options: map[string]interface{}{"skipTemplates": true}},
			// Template in assignment
			{Code: "const error = ` \u3000 `;", Options: map[string]interface{}{"skipTemplates": true}},
			// Template with newlines
			{Code: "const error = `\n\u3000`;", Options: map[string]interface{}{"skipTemplates": true}},
			{Code: "const error = `\u3000\n`;", Options: map[string]interface{}{"skipTemplates": true}},
			{Code: "const error = `\n\u3000\n`;", Options: map[string]interface{}{"skipTemplates": true}},
			{Code: "const error = `foo\u3000bar\nfoo\u3000bar`;", Options: map[string]interface{}{"skipTemplates": true}},

			// ---- skipJSXText: true ----
			{Code: "<div>\u000B</div>;", Tsx: true, Options: map[string]interface{}{"skipJSXText": true}},
			{Code: "<div>\u000C</div>;", Tsx: true, Options: map[string]interface{}{"skipJSXText": true}},
			{Code: "<div>\u0085</div>;", Tsx: true, Options: map[string]interface{}{"skipJSXText": true}},
			{Code: "<div>\u00A0</div>;", Tsx: true, Options: map[string]interface{}{"skipJSXText": true}},
			{Code: "<div>\u1680</div>;", Tsx: true, Options: map[string]interface{}{"skipJSXText": true}},
			{Code: "<div>\u180E</div>;", Tsx: true, Options: map[string]interface{}{"skipJSXText": true}},
			{Code: "<div>\uFEFF</div>;", Tsx: true, Options: map[string]interface{}{"skipJSXText": true}},
			{Code: "<div>\u2000</div>;", Tsx: true, Options: map[string]interface{}{"skipJSXText": true}},
			{Code: "<div>\u2001</div>;", Tsx: true, Options: map[string]interface{}{"skipJSXText": true}},
			{Code: "<div>\u2002</div>;", Tsx: true, Options: map[string]interface{}{"skipJSXText": true}},
			{Code: "<div>\u2003</div>;", Tsx: true, Options: map[string]interface{}{"skipJSXText": true}},
			{Code: "<div>\u2004</div>;", Tsx: true, Options: map[string]interface{}{"skipJSXText": true}},
			{Code: "<div>\u2005</div>;", Tsx: true, Options: map[string]interface{}{"skipJSXText": true}},
			{Code: "<div>\u2006</div>;", Tsx: true, Options: map[string]interface{}{"skipJSXText": true}},
			{Code: "<div>\u2007</div>;", Tsx: true, Options: map[string]interface{}{"skipJSXText": true}},
			{Code: "<div>\u2008</div>;", Tsx: true, Options: map[string]interface{}{"skipJSXText": true}},
			{Code: "<div>\u2009</div>;", Tsx: true, Options: map[string]interface{}{"skipJSXText": true}},
			{Code: "<div>\u200A</div>;", Tsx: true, Options: map[string]interface{}{"skipJSXText": true}},
			{Code: "<div>\u200B</div>;", Tsx: true, Options: map[string]interface{}{"skipJSXText": true}},
			{Code: "<div>\u202F</div>;", Tsx: true, Options: map[string]interface{}{"skipJSXText": true}},
			{Code: "<div>\u205F</div>;", Tsx: true, Options: map[string]interface{}{"skipJSXText": true}},
			{Code: "<div>\u3000</div>;", Tsx: true, Options: map[string]interface{}{"skipJSXText": true}},

			// ---- Unicode BOM at start of file ----
			{Code: "\uFEFFconsole.log('hello BOM');"},

			// ---- No irregular whitespace ----
			{Code: "var a = 1;"},
			{Code: ""},

			// ---- Options via array format (tests JSON round-trip) ----
			{Code: "// \u00A0", Options: []interface{}{map[string]interface{}{"skipComments": true}}},

			// ---- Extra: nested template with skipTemplates ----
			{Code: "`outer ${ `inner\u3000` } outer`", Options: map[string]interface{}{"skipTemplates": true}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Irregular whitespace in code (all chars from upstream) ----
			{
				Code: "var any \u000B = 'thing';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Message: "Irregular whitespace not allowed.", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code: "var any \u000C = 'thing';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code: "var any \u00A0 = 'thing';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},
			// NOTE: upstream intentionally comments out \u180E (Mongolian Vowel Separator)
			// because it was removed from General_Category=Zs in Unicode 6.3.0.
			// We still detect it per the rule's regex, but don't add the upstream's
			// commented-out var test.
			{
				Code: "var any \uFEFF = 'thing';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code: "var any \u2000 = 'thing';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code: "var any \u2001 = 'thing';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code: "var any \u2002 = 'thing';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code: "var any \u2003 = 'thing';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code: "var any \u2004 = 'thing';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code: "var any \u2005 = 'thing';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code: "var any \u2006 = 'thing';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code: "var any \u2007 = 'thing';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code: "var any \u2008 = 'thing';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code: "var any \u2009 = 'thing';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code: "var any \u200A = 'thing';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code: "var any \u2028 = 'thing';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 9, EndLine: 2, EndColumn: 1},
				},
			},
			{
				Code: "var any \u2029 = 'thing';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 9, EndLine: 2, EndColumn: 1},
				},
			},
			{
				Code: "var any \u202F = 'thing';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code: "var any \u205F = 'thing';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code: "var any \u3000 = 'thing';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},

			// ---- Multi-line with \u2028 line separators (from upstream) ----
			{
				Code: "var a = 'b',\u2028c = 'd',\ne = 'f'\u2028",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 13, EndLine: 2, EndColumn: 1},
					{MessageId: "noIrregularWhitespace", Line: 3, Column: 8, EndLine: 4, EndColumn: 1},
				},
			},

			// ---- Multiple errors on same/multiple lines (from upstream) ----
			{
				Code: "var any \u3000 = 'thing', other \u3000 = 'thing';\nvar third \u3000 = 'thing';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 28, EndLine: 1, EndColumn: 29},
					{MessageId: "noIrregularWhitespace", Line: 2, Column: 11, EndLine: 2, EndColumn: 12},
				},
			},

			// ---- Single-line comments (all 21 chars from upstream) ----
			{Code: "// \u000B", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "// \u000C", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "// \u0085", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "// \u00A0", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "// \u180E", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "// \uFEFF", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "// \u2000", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "// \u2001", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "// \u2002", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "// \u2003", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "// \u2004", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "// \u2005", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "// \u2006", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "// \u2007", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "// \u2008", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "// \u2009", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "// \u200A", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "// \u200B", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "// \u202F", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "// \u205F", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "// \u3000", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},

			// ---- Block comments (all 23 chars from upstream, including \u2028/\u2029) ----
			{Code: "/* \u000B */", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "/* \u000C */", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "/* \u0085 */", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "/* \u00A0 */", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "/* \u180E */", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "/* \uFEFF */", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "/* \u2000 */", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "/* \u2001 */", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "/* \u2002 */", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "/* \u2003 */", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "/* \u2004 */", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "/* \u2005 */", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "/* \u2006 */", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "/* \u2007 */", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "/* \u2008 */", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "/* \u2009 */", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "/* \u200A */", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "/* \u200B */", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "/* \u2028 */", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 2, EndColumn: 1}}},
			{Code: "/* \u2029 */", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 2, EndColumn: 1}}},
			{Code: "/* \u202F */", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "/* \u205F */", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "/* \u3000 */", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},

			// ---- Regex (from upstream) ----
			{
				Code: "var any = /\u3000/, other = /\u000B/;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 25, EndLine: 1, EndColumn: 26},
				},
			},

			// ---- skipStrings: false (from upstream) ----
			{
				Code:    "var any = '\u3000', other = '\u000B';",
				Options: map[string]interface{}{"skipStrings": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 25, EndLine: 1, EndColumn: 26},
				},
			},

			// ---- Template literals (from upstream) ----
			{
				Code: "var any = `\u3000`, other = `\u000B`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 25, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code:    "var any = `\u3000`, other = `\u000B`;",
				Options: map[string]interface{}{"skipTemplates": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 25, EndLine: 1, EndColumn: 26},
				},
			},

			// ---- skipTemplates: true but irregular in expression part (from upstream) ----
			{
				Code:    "`something ${\u3000 10} another thing`",
				Options: map[string]interface{}{"skipTemplates": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    "`something ${10\u3000} another thing`",
				Options: map[string]interface{}{"skipTemplates": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 16, EndLine: 1, EndColumn: 17},
				},
			},

			// ---- skipTemplates: true but irregular outside template (from upstream) ----
			{
				Code:    "\u3000\n`\u3000template`",
				Options: map[string]interface{}{"skipTemplates": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			{
				Code:    "\u3000\n`\u3000multiline\ntemplate`",
				Options: map[string]interface{}{"skipTemplates": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			{
				Code:    "\u3000`\u3000template`",
				Options: map[string]interface{}{"skipTemplates": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			{
				Code:    "\u3000`\u3000multiline\ntemplate`",
				Options: map[string]interface{}{"skipTemplates": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			{
				Code:    "`\u3000template`\u3000",
				Options: map[string]interface{}{"skipTemplates": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    "`\u3000multiline\ntemplate`\u3000",
				Options: map[string]interface{}{"skipTemplates": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 2, Column: 10, EndLine: 2, EndColumn: 11},
				},
			},
			{
				Code:    "`\u3000template`\n\u3000",
				Options: map[string]interface{}{"skipTemplates": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 2, Column: 1, EndLine: 2, EndColumn: 2},
				},
			},
			{
				Code:    "`\u3000multiline\ntemplate`\n\u3000",
				Options: map[string]interface{}{"skipTemplates": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 3, Column: 1, EndLine: 3, EndColumn: 2},
				},
			},

			// ---- Full location tests (from upstream) ----
			{
				Code: "var foo = \u000B bar;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
				},
			},
			{
				Code: "var foo =\u000Bbar;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code: "var foo = \u000B\u000B bar;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 11, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "var foo = \u000B\u000C bar;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 11, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "var foo = \u000B \u000B bar;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: "var foo = \u000Bbar\u000B;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code: "\u000B",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			{
				Code: "\u00A0\u2002\u2003",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 1, EndLine: 1, EndColumn: 4},
				},
			},
			{
				Code: "var foo = \u000B\nbar;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
				},
			},
			{
				Code: "var foo =\u000B\n\u000Bbar;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
					{MessageId: "noIrregularWhitespace", Line: 2, Column: 1, EndLine: 2, EndColumn: 2},
				},
			},
			{
				Code: "var foo = \u000C\u000B\n\u000C\u000B\u000Cbar\n;\u000B\u000C\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 11, EndLine: 1, EndColumn: 13},
					{MessageId: "noIrregularWhitespace", Line: 2, Column: 1, EndLine: 2, EndColumn: 4},
					{MessageId: "noIrregularWhitespace", Line: 3, Column: 2, EndLine: 3, EndColumn: 4},
				},
			},
			// Line separator full location
			{
				Code: "var foo = \u2028bar;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 11, EndLine: 2, EndColumn: 1},
				},
			},
			{
				Code: "var foo =\u2029 bar;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 10, EndLine: 2, EndColumn: 1},
				},
			},
			{
				Code: "var foo = bar;\u2028",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 15, EndLine: 2, EndColumn: 1},
				},
			},
			{
				Code: "\u2029",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 1, EndLine: 2, EndColumn: 1},
				},
			},
			{
				Code: "foo\u2028\u2028",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 2, EndColumn: 1},
					{MessageId: "noIrregularWhitespace", Line: 2, Column: 1, EndLine: 3, EndColumn: 1},
				},
			},
			{
				Code: "foo\u2029\u2028",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 2, EndColumn: 1},
					{MessageId: "noIrregularWhitespace", Line: 2, Column: 1, EndLine: 3, EndColumn: 1},
				},
			},
			{
				Code: "foo\u2028\n\u2028",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 2, EndColumn: 1},
					{MessageId: "noIrregularWhitespace", Line: 3, Column: 1, EndLine: 4, EndColumn: 1},
				},
			},
			{
				Code: "foo\u000B\u2028\u000B",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5},
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 5, EndLine: 2, EndColumn: 1},
					{MessageId: "noIrregularWhitespace", Line: 2, Column: 1, EndLine: 2, EndColumn: 2},
				},
			},

			// ---- JSX text (all 21 chars from upstream, skipJSXText default false) ----
			{Code: "<div>\u000B</div>;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 6, EndLine: 1, EndColumn: 7}}},
			{Code: "<div>\u000C</div>;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 6, EndLine: 1, EndColumn: 7}}},
			{Code: "<div>\u0085</div>;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 6, EndLine: 1, EndColumn: 7}}},
			{Code: "<div>\u00A0</div>;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 6, EndLine: 1, EndColumn: 7}}},
			{Code: "<div>\u180E</div>;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 6, EndLine: 1, EndColumn: 7}}},
			{Code: "<div>\uFEFF</div>;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 6, EndLine: 1, EndColumn: 7}}},
			{Code: "<div>\u2000</div>;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 6, EndLine: 1, EndColumn: 7}}},
			{Code: "<div>\u2001</div>;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 6, EndLine: 1, EndColumn: 7}}},
			{Code: "<div>\u2002</div>;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 6, EndLine: 1, EndColumn: 7}}},
			{Code: "<div>\u2003</div>;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 6, EndLine: 1, EndColumn: 7}}},
			{Code: "<div>\u2004</div>;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 6, EndLine: 1, EndColumn: 7}}},
			{Code: "<div>\u2005</div>;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 6, EndLine: 1, EndColumn: 7}}},
			{Code: "<div>\u2006</div>;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 6, EndLine: 1, EndColumn: 7}}},
			{Code: "<div>\u2007</div>;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 6, EndLine: 1, EndColumn: 7}}},
			{Code: "<div>\u2008</div>;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 6, EndLine: 1, EndColumn: 7}}},
			{Code: "<div>\u2009</div>;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 6, EndLine: 1, EndColumn: 7}}},
			{Code: "<div>\u200A</div>;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 6, EndLine: 1, EndColumn: 7}}},
			{Code: "<div>\u200B</div>;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 6, EndLine: 1, EndColumn: 7}}},
			{Code: "<div>\u202F</div>;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 6, EndLine: 1, EndColumn: 7}}},
			{Code: "<div>\u205F</div>;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 6, EndLine: 1, EndColumn: 7}}},
			{Code: "<div>\u3000</div>;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 6, EndLine: 1, EndColumn: 7}}},

			// ---- Options via array format (tests JSON round-trip) ----
			{
				Code:    "var any = '\u3000';",
				Options: []interface{}{map[string]interface{}{"skipStrings": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
				},
			},

			// ---- Extra: \u1680 (Ogham Space Mark) — not tested upstream but in the rule's char set ----
			{
				Code: "var any \u1680 = 'thing';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},
			{Code: "// \u1680", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "/* \u1680 */", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 4, EndLine: 1, EndColumn: 5}}},
			{Code: "<div>\u1680</div>;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noIrregularWhitespace", Line: 1, Column: 6, EndLine: 1, EndColumn: 7}}},

			// ---- Extra: nested template with skipTemplates — irregular in outer expression ----
			{
				Code:    "`outer ${\u3000 `inner` } outer`",
				Options: map[string]interface{}{"skipTemplates": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIrregularWhitespace", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
				},
			},
		},
	)
}
