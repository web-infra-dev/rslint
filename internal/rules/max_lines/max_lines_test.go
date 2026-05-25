package max_lines

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestMaxLines exercises three overlapping test corpora in a single run:
//
//  1. ESLint parity — ports of cases from eslint/tests/lib/rules/max-lines.js,
//     asserting identical line / column / endLine / endColumn / messageId.
//  2. Additional edge cases (negative max, shebang, complex comment/code
//     interactions, multi-byte column precision, ECMA line terminators).
//  3. Differential-test corpus — 121 cases whose expected diagnostics were
//     captured verbatim from real ESLint v9. Any divergence surfaces here.
//
// The three groups overlap; that is intentional redundancy — a regression in
// any single dimension should fire in multiple subtests.
func TestMaxLines(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&MaxLinesRule,
		[]rule_tester.ValidTestCase{
			// ============================================================
			// 1. ESLint parity — mirrors tests/lib/rules/max-lines.js
			// ============================================================
			{Code: `var x;`},
			{Code: "var xy;\nvar xy;"},
			{Code: `A`, Options: 1},
			{Code: "A\n", Options: 1},
			{Code: "A\r", Options: 1},
			{Code: "A\r\n", Options: 1},
			{Code: "var xy;\nvar xy;", Options: 2},
			{Code: "var xy;\nvar xy;\n", Options: 2},
			{Code: "var xy;\nvar xy;", Options: map[string]interface{}{"max": 2}},
			{
				Code:    "// comment\n",
				Options: map[string]interface{}{"max": 0, "skipComments": true},
			},
			{
				Code:    "foo;\n /* comment */\n",
				Options: map[string]interface{}{"max": 1, "skipComments": true},
			},
			{
				Code: strings.Join([]string{
					"//a single line comment",
					"var xy;",
					"var xy;",
					" /* a multiline",
					" really really",
					" long comment*/ ",
				}, "\n"),
				Options: map[string]interface{}{"max": 2, "skipComments": true},
			},
			{
				Code: strings.Join([]string{
					"var x; /* inline comment",
					" spanning multiple lines */ var z;",
				}, "\n"),
				Options: map[string]interface{}{"max": 2, "skipComments": true},
			},
			{
				Code: strings.Join([]string{
					"var x; /* inline comment",
					" spanning multiple lines */",
					"var z;",
				}, "\n"),
				Options: map[string]interface{}{"max": 2, "skipComments": true},
			},
			{
				Code:    strings.Join([]string{"var x;", "", "\t", "\t  ", "var y;"}, "\n"),
				Options: map[string]interface{}{"max": 2, "skipBlankLines": true},
			},
			{
				Code: strings.Join([]string{
					"//a single line comment",
					"var xy;",
					" ",
					"var xy;",
					" ",
					" /* a multiline",
					" really really",
					" long comment*/",
				}, "\n"),
				Options: map[string]interface{}{"max": 2, "skipComments": true, "skipBlankLines": true},
			},

			// ============================================================
			// 2. Additional edge cases
			// ============================================================

			// --- ECMA line terminators ---
			{Code: "var a;\u2028var b;\u2028var c;", Options: 3},
			{Code: "var a;\u2029var b;\u2029var c;", Options: 3},
			{Code: "a\nb\r\nc\rd", Options: 4},
			{Code: "a\nb\u2028c\u2029d", Options: 4},

			// --- Multi-byte source ---
			{Code: "const 日本語 = 'テスト';\nconst 中文 = '测试';", Options: 2},
			{Code: "var a = '🎉';\nvar b = '🎊';", Options: 2},

			// --- Comment arrangements ---
			{Code: "foo; /* a */ /* b */ bar;", Options: 1},
			{
				Code:    "var x; // hint\nvar y;",
				Options: map[string]interface{}{"max": 2, "skipComments": true},
			},
			{
				Code:    "/* line1\nline2\nline3 */",
				Options: map[string]interface{}{"max": 0, "skipComments": true},
			},
			{
				Code:    "/**/\nvar x;",
				Options: map[string]interface{}{"max": 1, "skipComments": true},
			},
			{
				Code:    "/**\n * doc\n */\nfunction f() { return 1; }",
				Options: map[string]interface{}{"max": 1, "skipComments": true},
			},
			{
				Code:    "// 🎉 celebrate\nvar x;\nvar y;",
				Options: map[string]interface{}{"max": 2, "skipComments": true},
			},
			{
				Code:    "var x;\n\t\n\v\n\f\nvar y;",
				Options: map[string]interface{}{"max": 2, "skipBlankLines": true},
			},
			{
				Code:    "#!/usr/bin/env node\nvar x;\nvar y;",
				Options: map[string]interface{}{"max": 2, "skipComments": true},
			},
			{
				Code:    "#!/usr/bin/env node\n// regular\nvar x;",
				Options: map[string]interface{}{"max": 1, "skipComments": true},
			},
			// BOM / NBSP on a line is "blank" per JS trim, unlike Go's
			// strings.TrimSpace — uses utils.IsStrWhiteSpace for alignment.
			{
				Code:    "var x;\n\u00A0\u00A0\n\uFEFF\nvar y;",
				Options: map[string]interface{}{"max": 2, "skipBlankLines": true},
			},

			// --- Nested / realistic code shapes ---
			{
				Code:    "class Foo {\n  bar() {\n    return 1;\n  }\n}",
				Options: 5,
			},
			{
				Code:    "const fns = [\n  () => ({\n    run() { return 1; },\n  }),\n];",
				Options: 5,
			},
			{Code: "var s = `a\nb\nc`;", Options: 3},
			{
				Code: strings.Join([]string{
					"(function () {",
					"  const x = 1;",
					"  return x;",
					"})();",
				}, "\n"),
				Options: 4,
			},

			// --- Options shapes ---
			{Code: "var x;\nvar y;", Options: 10000},
			{
				Code:    strings.Repeat("var x;\n", 10),
				Options: map[string]interface{}{},
			},
			{Code: "var x;\nvar y;", Options: []interface{}{2}},
			{
				Code:    "var x;\nvar y;",
				Options: []interface{}{map[string]interface{}{"max": 2}},
			},
			{Code: "var x;\nvar y;", Options: []interface{}{}},

			// ============================================================
			// 3. Differential-test corpus (verified against ESLint v9)
			// ============================================================
			{Code: "", Options: []interface{}{map[string]interface{}{"max": 1}}}, // empty-max-1
			{Code: "A", Options: []interface{}{map[string]interface{}{"max": 1}}},
			{Code: "#!/usr/bin/env node", Options: []interface{}{map[string]interface{}{"max": 0, "skipComments": true}}},
			{Code: "#!/usr/bin/env node\nvar x;\nvar y;", Options: []interface{}{map[string]interface{}{"max": 2, "skipComments": true}}},
			{Code: "#!/usr/bin/env node\n", Options: []interface{}{map[string]interface{}{"max": 0, "skipComments": true}}},
			{Code: "// a", Options: []interface{}{map[string]interface{}{"max": 0, "skipComments": true}}},
			{Code: "/* a */", Options: []interface{}{map[string]interface{}{"max": 0, "skipComments": true}}},
			{Code: "/* line1\nline2\nline3 */", Options: []interface{}{map[string]interface{}{"max": 0, "skipComments": true}}},
			{Code: "/**/\nvar x;", Options: []interface{}{map[string]interface{}{"max": 1, "skipComments": true}}},
			{Code: "foo; /* a */ /* b */ bar;", Options: []interface{}{1}},
			{Code: "var x;\n// trailing", Options: []interface{}{map[string]interface{}{"max": 1, "skipComments": true}}},
			{Code: "// leading\nvar x;", Options: []interface{}{map[string]interface{}{"max": 1, "skipComments": true}}},
			{Code: "/* a\nb */", Options: []interface{}{map[string]interface{}{"max": 0, "skipComments": true}}},
			{Code: "var x;\n\t\n\v\n\f\nvar y;", Options: []interface{}{map[string]interface{}{"max": 2, "skipBlankLines": true}}},
			{Code: "var x;\n\u00a0\u00a0\nvar y;", Options: []interface{}{map[string]interface{}{"max": 2, "skipBlankLines": true}}},
			{Code: "\n\n\nvar x;", Options: []interface{}{map[string]interface{}{"max": 1, "skipBlankLines": true}}},
			{Code: "var x;\n\n\n", Options: []interface{}{map[string]interface{}{"max": 1, "skipBlankLines": true}}},
			{Code: "const 日本語 = 1;\nconst 中文 = 2;", Options: []interface{}{2}},
			{Code: "var a = '🎉';\nvar b = '🎊';", Options: []interface{}{2}},
			{Code: "// 🎉 celebrate\nvar x;\nvar y;", Options: []interface{}{map[string]interface{}{"max": 2, "skipComments": true}}},
			{Code: "var s = `a\nb\nc`;", Options: []interface{}{3}},
			{Code: "class Foo {\n  bar() {\n    return 1;\n  }\n}", Options: []interface{}{5}},
			{Code: "const fns = [\n  () => ({\n    run() { return 1; },\n  }),\n];", Options: []interface{}{5}},
			{Code: "(function () {\n  const x = 1;\n  return x;\n})();", Options: []interface{}{4}},
			{Code: "\n\n\n", Options: []interface{}{map[string]interface{}{"max": 0, "skipBlankLines": true}}},
			{Code: "var x;\nvar y;"},
			{Code: "var x;\nvar y;", Options: []interface{}{map[string]interface{}{}}},
			{Code: "// a\n// b\nvar x;", Options: []interface{}{map[string]interface{}{"skipComments": true}}},
			{Code: "\n\nvar x;", Options: []interface{}{map[string]interface{}{"skipBlankLines": true}}},
			{Code: "// a\n\n\n// b\nvar x;", Options: []interface{}{map[string]interface{}{"max": 1, "skipComments": true, "skipBlankLines": true}}},
			{Code: "var x;\n// comment", Options: []interface{}{map[string]interface{}{"max": 1, "skipComments": true}}},
			{Code: "var x;\n// comment\n", Options: []interface{}{map[string]interface{}{"max": 1, "skipComments": true}}},
			{Code: "var r = /abc/g;", Options: []interface{}{1}},
			{Code: "a\n\rb", Options: []interface{}{3}},
			{Code: "a\r\r\nb", Options: []interface{}{3}},
			{Code: "var x;\nvar y;", Options: []interface{}{2}},
			{Code: "var x;\n", Options: []interface{}{map[string]interface{}{"max": 1}}},
			{Code: "\"use strict\";\nvar x;\nvar y;", Options: []interface{}{3}},
			{Code: "var s = `${\n  a\n  + b\n}`;", Options: []interface{}{4}},
			{Code: "var aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa = 1;", Options: []interface{}{1}},
			{Code: "\nvar x;\n// comment\n\n", Options: []interface{}{map[string]interface{}{"max": 1, "skipComments": true, "skipBlankLines": true}}},
			{Code: "// a\n// b\n// c\nvar x;", Options: []interface{}{map[string]interface{}{"max": 1, "skipComments": true}}},
			{Code: "/* a */\n/* b */\n/* c */\nvar x;", Options: []interface{}{map[string]interface{}{"max": 1, "skipComments": true}}},
			{Code: "/**\n * desc\n */\nfunction f() {}", Options: []interface{}{map[string]interface{}{"max": 1, "skipComments": true}}},
			{Code: "var x;\n\t\t\t\nvar y;", Options: []interface{}{map[string]interface{}{"max": 2, "skipBlankLines": true}}},
			{Code: "var x;\n\n// c\n\nvar y;\n\n// d\n\nvar z;", Options: []interface{}{map[string]interface{}{"max": 3, "skipComments": true, "skipBlankLines": true}}},
			{Code: "\uFEFF" + "var x;", Options: []interface{}{1}},
			{Code: "\ufeff\nvar x;", Options: []interface{}{map[string]interface{}{"max": 1, "skipBlankLines": true}}},
			{Code: "var x;\r", Options: []interface{}{1}},
			{Code: "var x;\u2028", Options: []interface{}{1}},
			{Code: "var x;\u2029", Options: []interface{}{1}},
			{Code: "#!/usr/bin/env -S node --harmony\nvar x;", Options: []interface{}{map[string]interface{}{"max": 1, "skipComments": true}}},
			{Code: "//\n//\n//\nvar x;", Options: []interface{}{map[string]interface{}{"max": 1, "skipComments": true}}},
			{Code: "var x;\u2028\u2028\u2028var y;", Options: []interface{}{map[string]interface{}{"max": 2, "skipBlankLines": true}}},
			{Code: "/// <reference path=\"foo\" />\nvar x;", Options: []interface{}{map[string]interface{}{"max": 1, "skipComments": true}}},
			{Code: "a\rb", Options: []interface{}{2}},
			{Code: "var x;\n// comment   \n", Options: []interface{}{map[string]interface{}{"max": 1, "skipComments": true}}},
			{Code: "var x;\n/* a\n*/   ", Options: []interface{}{map[string]interface{}{"max": 1, "skipComments": true}}},
			{Code: "var x;\nvar y;\nvar z;", Options: []interface{}{map[string]interface{}{"max": 300}}},
			{Code: "a\r\n\rb", Options: []interface{}{3}},
			{Code: "\n/* comment */\n", Options: []interface{}{map[string]interface{}{"max": 0, "skipComments": true, "skipBlankLines": true}}},
			{Code: "//", Options: []interface{}{map[string]interface{}{"max": 0, "skipComments": true}}},
			{Code: "var x;\n\t\nvar y;", Options: []interface{}{map[string]interface{}{"max": 2, "skipBlankLines": true}}},
			{Code: "var x;\nvar y;\nvar z;\nvar a;\nvar b;\nvar c;\nvar d;\nvar e;\nvar f;\nvar g;", Options: []interface{}{10}},
			{Code: "// a\n// b\n// c", Options: []interface{}{map[string]interface{}{"max": 0, "skipComments": true}}},
			{Code: "/* a\u2028b */", Options: []interface{}{map[string]interface{}{"max": 0, "skipComments": true}}},
			{Code: "/* a *//* b */\nvar x;", Options: []interface{}{map[string]interface{}{"max": 1, "skipComments": true}}},
			{Code: "/* a */// b\nvar x;", Options: []interface{}{map[string]interface{}{"max": 1, "skipComments": true}}},
			{Code: "const x = {\n  a: {\n    b: [\n      {\n        c: () => ({\n          d: 1,\n        }),\n      },\n    ],\n  },\n};", Options: []interface{}{11}},
			{Code: "var x;"},
		},
		[]rule_tester.InvalidTestCase{
			// ============================================================
			// 1. ESLint parity — mirrors tests/lib/rules/max-lines.js
			// ============================================================
			{
				Code:    "var xyz;\nvar xyz;\nvar xyz;",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 9},
				},
			},
			{
				Code:    "/* a multiline comment\n that goes to many lines*/\nvar xy;\nvar xy;",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 3, Column: 1, EndLine: 4, EndColumn: 8},
				},
			},
			{
				Code:    "//a single line comment\nvar xy;\nvar xy;",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 8},
				},
			},
			{
				Code:    strings.Join([]string{"var x;", "", "", "", "var y;"}, "\n"),
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 3, Column: 1, EndLine: 5, EndColumn: 7},
				},
			},
			{
				Code: strings.Join([]string{
					"//a single line comment",
					"var xy;",
					" ",
					"var xy;",
					" ",
					" /* a multiline",
					" really really",
					" long comment*/",
				}, "\n"),
				Options: map[string]interface{}{"max": 2, "skipComments": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 4, Column: 1, EndLine: 8, EndColumn: 16},
				},
			},
			{
				Code:    strings.Join([]string{"var x; // inline comment", "var y;", "var z;"}, "\n"),
				Options: map[string]interface{}{"max": 2, "skipComments": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 7},
				},
			},
			{
				Code: strings.Join([]string{
					"var x; /* inline comment",
					" spanning multiple lines */",
					"var y;",
					"var z;",
				}, "\n"),
				Options: map[string]interface{}{"max": 2, "skipComments": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 4, Column: 1, EndLine: 4, EndColumn: 7},
				},
			},
			{
				Code: strings.Join([]string{
					"//a single line comment",
					"var xy;",
					" ",
					"var xy;",
					" ",
					" /* a multiline",
					" really really",
					" long comment*/",
				}, "\n"),
				Options: map[string]interface{}{"max": 2, "skipBlankLines": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 4, Column: 1, EndLine: 8, EndColumn: 16},
				},
			},
			{
				Code:    strings.TrimRight(strings.Repeat("AAAAAAAA\n", 301), " \t\n\r"),
				Options: map[string]interface{}{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 301, Column: 1, EndLine: 301, EndColumn: 9},
				},
			},
			{
				Code:    "",
				Options: map[string]interface{}{"max": 0},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 1, Column: 1, EndLine: 1, EndColumn: 1},
				},
			},
			{
				Code:    " ",
				Options: map[string]interface{}{"max": 0},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			{
				Code:    "\n",
				Options: map[string]interface{}{"max": 0},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 1, Column: 1, EndLine: 2, EndColumn: 1},
				},
			},
			{
				Code:    "A",
				Options: map[string]interface{}{"max": 0},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			{
				Code:    "A\n",
				Options: map[string]interface{}{"max": 0},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 1, Column: 1, EndLine: 2, EndColumn: 1},
				},
			},
			{
				Code:    "A\n ",
				Options: map[string]interface{}{"max": 0},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 1, Column: 1, EndLine: 2, EndColumn: 2},
				},
			},
			{
				Code:    "A\n ",
				Options: map[string]interface{}{"max": 1},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 2, Column: 1, EndLine: 2, EndColumn: 2},
				},
			},
			{
				Code:    "A\n\n",
				Options: map[string]interface{}{"max": 1},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 2, Column: 1, EndLine: 3, EndColumn: 1},
				},
			},
			{
				Code:    strings.Join([]string{"var a = 'a'; ", "var x", "var c;", "console.log"}, "\n"),
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 3, Column: 1, EndLine: 4, EndColumn: 12},
				},
			},
			{
				Code:    "var a = 'a',\nc,\nx;\r",
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 3, Column: 1, EndLine: 4, EndColumn: 1},
				},
			},
			{
				Code:    "var a = 'a',\nc,\nx;\n",
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 3, Column: 1, EndLine: 4, EndColumn: 1},
				},
			},
			{
				Code:    "\n\nvar a = 'a',\nc,\nx;\n",
				Options: map[string]interface{}{"max": 2, "skipBlankLines": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 5, Column: 1, EndLine: 6, EndColumn: 1},
				},
			},
			{
				Code: strings.Join([]string{
					"var a = 'a'; ",
					"var x",
					"var c;",
					"console.log",
					"// some block ",
					"// comments",
				}, "\n"),
				Options: map[string]interface{}{"max": 2, "skipComments": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 3, Column: 1, EndLine: 6, EndColumn: 12},
				},
			},
			{
				Code: strings.Join([]string{
					"var a = 'a'; ",
					"var x",
					"var c;",
					"console.log",
					"/* block comments */",
				}, "\n"),
				Options: map[string]interface{}{"max": 2, "skipComments": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 3, Column: 1, EndLine: 5, EndColumn: 21},
				},
			},
			{
				Code: strings.Join([]string{
					"var a = 'a'; ",
					"var x",
					"var c;",
					"console.log",
					"/* block comments */\n",
				}, "\n"),
				Options: map[string]interface{}{"max": 2, "skipComments": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 3, Column: 1, EndLine: 6, EndColumn: 1},
				},
			},
			{
				Code: strings.Join([]string{
					"var a = 'a'; ",
					"var x",
					"var c;",
					"console.log",
					"/** block \n\n comments */",
				}, "\n"),
				Options: map[string]interface{}{"max": 2, "skipComments": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 3, Column: 1, EndLine: 7, EndColumn: 13},
				},
			},
			{
				Code:    strings.Join([]string{"var a = 'a'; ", "", "", "// comment"}, "\n"),
				Options: map[string]interface{}{"max": 2, "skipComments": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 3, Column: 1, EndLine: 4, EndColumn: 11},
				},
			},
			{
				Code: strings.Join([]string{
					"var a = 'a'; ",
					"var x",
					"\n",
					"var c;",
					"console.log",
					"\n",
				}, "\n"),
				Options: map[string]interface{}{"max": 2, "skipBlankLines": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 5, Column: 1, EndLine: 8, EndColumn: 1},
				},
			},
			{
				Code: strings.Join([]string{
					"var a = 'a'; ",
					"\n",
					"var x",
					"var c;",
					"console.log",
					"\n",
				}, "\n"),
				Options: map[string]interface{}{"max": 2, "skipBlankLines": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 5, Column: 1, EndLine: 8, EndColumn: 1},
				},
			},
			{
				Code: strings.Join([]string{
					"var a = 'a'; ",
					"//",
					"var x",
					"var c;",
					"console.log",
					"//",
				}, "\n"),
				Options: map[string]interface{}{"max": 2, "skipComments": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 4, Column: 1, EndLine: 6, EndColumn: 3},
				},
			},
			{
				Code: strings.Join([]string{
					"// hello world",
					"/*hello",
					" world 2 */",
					"var a,",
					"b",
					"// hh",
					"c,",
					"e,",
					"f;",
				}, "\n"),
				Options: map[string]interface{}{"max": 2, "skipComments": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 7, Column: 1, EndLine: 9, EndColumn: 3},
				},
			},
			{
				Code: strings.Join([]string{
					"",
					"var x = '';",
					"",
					"// comment",
					"",
					"var b = '',",
					"c,",
					"d,",
					"e",
					"",
					"// comment",
				}, "\n"),
				Options: map[string]interface{}{"max": 2, "skipComments": true, "skipBlankLines": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 7, Column: 1, EndLine: 11, EndColumn: 11},
				},
			},

			// ============================================================
			// 2. Additional edge cases
			// ============================================================

			// --- ECMA line terminators over limit ---
			{
				Code:    "var a;\u2028var b;\u2028var c;",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 7},
				},
			},
			{
				Code:    "var a;\u2029var b;\u2029var c;",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 7},
				},
			},
			{
				Code:    "a\nb\r\nc\rd",
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 4, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code:    "a\nb\u2028c\u2029d",
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 4, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},

			// --- Multi-byte content: end column uses UTF-16 units ---
			{
				Code:    "const 日本語 = '1';\nconst b = '🎉';\nconst c = '3';",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 15},
				},
			},

			// --- Template literal counts physical lines ---
			{
				Code:    "var s = `a\nb\nc\nd`;",
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 4, Column: 1, EndLine: 4, EndColumn: 4},
				},
			},

			// --- Nested code shapes ---
			{
				Code:    "class Foo {\n  bar() {\n    return 1;\n  }\n}",
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 4, Column: 1, EndLine: 5, EndColumn: 2},
				},
			},
			{
				Code:    "const fns = [\n  () => 1,\n  () => 2,\n  () => 3,\n];",
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 4, Column: 1, EndLine: 5, EndColumn: 3},
				},
			},

			// --- Comment interactions ---
			{
				Code:    "foo; /* a */ /* b */ bar;\nbar;\nbaz;",
				Options: map[string]interface{}{"max": 2, "skipComments": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 5},
				},
			},
			{
				Code:    "/**/ var x;\nvar y;\nvar z;",
				Options: map[string]interface{}{"max": 2, "skipComments": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 7},
				},
			},
			{
				Code:    "/* line1\nline2\nline3 */",
				Options: map[string]interface{}{"max": 0},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 1, Column: 1, EndLine: 3, EndColumn: 9},
				},
			},

			// --- Blank line handling without skipBlankLines ---
			{
				Code:    "var a;\n\n\n\n\n\n\n\nvar b;\nvar c;",
				Options: map[string]interface{}{"max": 3},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 4, Column: 1, EndLine: 10, EndColumn: 7},
				},
			},

			// --- Option shapes for invalid cases ---
			{
				Code:    "var a;\nvar b;\nvar c;",
				Options: []interface{}{2},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 7},
				},
			},
			{
				Code:    "var a;\nvar b;\nvar c;",
				Options: []interface{}{map[string]interface{}{"max": 2}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 7},
				},
			},
			// Defensive: negative max is treated like 0 rather than crashing.
			{
				Code:    "var a;",
				Options: map[string]interface{}{"max": -1},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
				},
			},
			// Hashbang without skipComments counts toward the limit.
			{
				Code:    "#!/usr/bin/env node\nvar x;\nvar y;",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 7},
				},
			},
			// Hashbang + skipComments: line 1 filtered, remaining still exceeds.
			{
				Code:    "#!/usr/bin/env node\nvar x;\nvar y;",
				Options: map[string]interface{}{"max": 1, "skipComments": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 7},
				},
			},

			// ============================================================
			// 3. Differential-test corpus (verified against ESLint v9)
			// ============================================================
			{Code: "", Options: []interface{}{map[string]interface{}{"max": 0}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1, EndLine: 1, EndColumn: 1}}},
			{Code: " ", Options: []interface{}{map[string]interface{}{"max": 0}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1, EndLine: 1, EndColumn: 2}}},
			{Code: "\n", Options: []interface{}{map[string]interface{}{"max": 0}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1, EndLine: 2, EndColumn: 1}}},
			{Code: "\r", Options: []interface{}{map[string]interface{}{"max": 0}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1, EndLine: 2, EndColumn: 1}}},
			{Code: "\r\n", Options: []interface{}{map[string]interface{}{"max": 0}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1, EndLine: 2, EndColumn: 1}}},
			{Code: "\u2028", Options: []interface{}{map[string]interface{}{"max": 0}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1, EndLine: 2, EndColumn: 1}}},
			{Code: "\u2029", Options: []interface{}{map[string]interface{}{"max": 0}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1, EndLine: 2, EndColumn: 1}}},
			{Code: "\n\n", Options: []interface{}{map[string]interface{}{"max": 0}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1, EndLine: 3, EndColumn: 1}}},
			{Code: "A", Options: []interface{}{map[string]interface{}{"max": 0}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1, EndLine: 1, EndColumn: 2}}},
			{Code: "var a;\nvar b;\nvar c;", Options: []interface{}{2}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 7}}},
			{Code: "var a;\rvar b;\rvar c;", Options: []interface{}{2}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 7}}},
			{Code: "var a;\r\nvar b;\r\nvar c;", Options: []interface{}{2}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 7}}},
			{Code: "var a;\u2028var b;\u2028var c;", Options: []interface{}{2}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 7}}},
			{Code: "var a;\u2029var b;\u2029var c;", Options: []interface{}{2}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 7}}},
			{Code: "a\nb\r\nc\rd", Options: []interface{}{3}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 4, Column: 1, EndLine: 4, EndColumn: 2}}},
			{Code: "a\nb\u2028c\u2029d", Options: []interface{}{3}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 4, Column: 1, EndLine: 4, EndColumn: 2}}},
			{Code: "#!/usr/bin/env node", Options: []interface{}{map[string]interface{}{"max": 0}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}}},
			{Code: "#!/usr/bin/env node\nvar x;\nvar y;", Options: []interface{}{map[string]interface{}{"max": 1, "skipComments": true}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 7}}},
			{Code: "#!/usr/bin/env node\n// other\nvar x;", Options: []interface{}{map[string]interface{}{"max": 0, "skipComments": true}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 7}}},
			{Code: "// a", Options: []interface{}{map[string]interface{}{"max": 0}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1, EndLine: 1, EndColumn: 5}}},
			{Code: "/* line1\nline2\nline3 */", Options: []interface{}{map[string]interface{}{"max": 0}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1, EndLine: 3, EndColumn: 9}}},
			{Code: "/* a */ // b\nvar x;", Options: []interface{}{map[string]interface{}{"max": 0, "skipComments": true}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 2, Column: 1, EndLine: 2, EndColumn: 7}}},
			{Code: "foo; /* a */ /* b */ bar;\nbar;\nbaz;", Options: []interface{}{map[string]interface{}{"max": 2, "skipComments": true}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 5}}},
			{Code: "var x; // tail\nvar y;", Options: []interface{}{map[string]interface{}{"max": 1, "skipComments": true}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 2, Column: 1, EndLine: 2, EndColumn: 7}}},
			{Code: "a /* b\nc */", Options: []interface{}{map[string]interface{}{"max": 0, "skipComments": true}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1, EndLine: 2, EndColumn: 5}}},
			{Code: "/* a\nb */ c", Options: []interface{}{map[string]interface{}{"max": 0, "skipComments": true}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 2, Column: 1, EndLine: 2, EndColumn: 7}}},
			{Code: "a /* b\nc */ d", Options: []interface{}{map[string]interface{}{"max": 0, "skipComments": true}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1, EndLine: 2, EndColumn: 7}}},
			{Code: "const 日本語 = '1';\nconst b = '🎉';\nconst c = '3';", Options: []interface{}{2}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 15}}},
			{Code: "var s = `a\nb\nc\nd`;", Options: []interface{}{3}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 4, Column: 1, EndLine: 4, EndColumn: 4}}},
			{Code: "class Foo {\n  bar() {\n    return 1;\n  }\n}", Options: []interface{}{3}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 4, Column: 1, EndLine: 5, EndColumn: 2}}},
			{Code: "/* line1\nline2\nline3 */", Options: []interface{}{map[string]interface{}{"max": 0}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1, EndLine: 3, EndColumn: 9}}},
			{Code: "// a\n\n\n// b\nvar x;", Options: []interface{}{map[string]interface{}{"max": 0, "skipComments": true, "skipBlankLines": true}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 5, Column: 1, EndLine: 5, EndColumn: 7}}},
			{Code: "/* a */ var x;\nvar y;", Options: []interface{}{map[string]interface{}{"max": 1, "skipComments": true}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 2, Column: 1, EndLine: 2, EndColumn: 7}}},
			{Code: "var x; /* a */\nvar y;", Options: []interface{}{map[string]interface{}{"max": 1, "skipComments": true}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 2, Column: 1, EndLine: 2, EndColumn: 7}}},
			{Code: "a\n\rb", Options: []interface{}{2}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 2}}},
			{Code: "a\r\r\nb", Options: []interface{}{2}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 2}}},
			{Code: "var x;\nvar y;\nvar z;", Options: []interface{}{2}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 7}}},
			{Code: "\"use strict\";\nvar x;\nvar y;", Options: []interface{}{2}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 7}}},
			{Code: "var s = `${\n  a\n  + b\n}`;", Options: []interface{}{3}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 4, Column: 1, EndLine: 4, EndColumn: 4}}},
			{Code: "var aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa = 1;", Options: []interface{}{0}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1, EndLine: 1, EndColumn: 62}}},
			{Code: "a\rb", Options: []interface{}{1}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 2, Column: 1, EndLine: 2, EndColumn: 2}}},
			{Code: "a\r\n\rb", Options: []interface{}{2}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 2}}},
			{Code: "a\r\n\nb", Options: []interface{}{2}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 2}}},
			{Code: "\n/* comment */\n", Options: []interface{}{map[string]interface{}{"max": 0, "skipComments": true}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1, EndLine: 3, EndColumn: 1}}},
			{Code: "//", Options: []interface{}{map[string]interface{}{"max": 0}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1, EndLine: 1, EndColumn: 3}}},
			{Code: "var x;\nvar y;\nvar z;\nvar a;\nvar b;\nvar c;\nvar d;\nvar e;\nvar f;\nvar g;", Options: []interface{}{9}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 10, Column: 1, EndLine: 10, EndColumn: 7}}},
			{Code: "var x; // a\nvar y; // b\nvar z; // c", Options: []interface{}{map[string]interface{}{"max": 2, "skipComments": true}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 12}}},
			{Code: "var s = '\u2028';", Options: []interface{}{1}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 2, Column: 1, EndLine: 2, EndColumn: 3}}},
			{Code: "var s = '\u2028';", Options: []interface{}{map[string]interface{}{"max": 1, "skipBlankLines": true}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 2, Column: 1, EndLine: 2, EndColumn: 3}}},
			{Code: "var x;\nvar y;\nvar z = '日';", Options: []interface{}{2}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 3, Column: 1, EndLine: 3, EndColumn: 13}}},
			{Code: "const x = {\n  a: {\n    b: [\n      {\n        c: () => ({\n          d: 1,\n        }),\n      },\n    ],\n  },\n};", Options: []interface{}{5}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 6, Column: 1, EndLine: 11, EndColumn: 3}}},
		},
	)
}
