package capitalized_comments

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestCapitalizedCommentsUpstream migrates the full valid/invalid suite from
// upstream eslint/tests/lib/rules/capitalized-comments.js 1:1. Position
// assertions cover line/column for every invalid case. rslint-specific
// lock-in cases live in the capitalized_comments_extras_test.go file.
func TestCapitalizedCommentsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&CapitalizedCommentsRule,
		[]rule_tester.ValidTestCase{
			// ---- No options: capitalization required ----
			{Code: "//Uppercase"},
			{Code: "// Uppercase"},
			{Code: "/*Uppercase */"},
			{Code: "/* Uppercase */"},
			{Code: "/*\nUppercase */"},
			{Code: "/** Uppercase */"},
			{Code: "/**\nUppercase */"},
			{Code: "//Über"},
			{Code: "//Π"},
			{Code: "/* Uppercase\nsecond line need not be uppercase */"},

			// ---- No options: skips comments that only contain whitespace ----
			{Code: "// "},
			{Code: "//\t"},
			{Code: "/* */"},
			{Code: "/*\t*/"},
			{Code: "/*\n*/"},
			{Code: "/*\r*/"},
			{Code: "/*\r\n*/"},
			{Code: "/* */"},
			{Code: "/* */"},

			// ---- No options: non-alphabetical is okay ----
			{Code: "//123"},
			{Code: "// 123"},
			{Code: "/*123*/"},
			{Code: "/* 123 */"},
			{Code: "/**123 */"},
			{Code: "/** 123 */"},
			{Code: "/**\n123 */"},
			{Code: "/*\n123 */"},
			{Code: "/*123\nsecond line need not be uppercase */"},
			{Code: "/**\n * @fileoverview This is a file */"},

			// ---- No options: eslint/istanbul/jshint/jscs/globals?/exported are okay ----
			{Code: "// jscs: enable"},
			{Code: "// jscs:disable"},
			{Code: "// eslint-disable-line"},
			{Code: "// eslint-disable-next-line"},
			{Code: "/* eslint semi:off */"},
			{Code: "/* eslint-enable */"},
			{Code: "/* istanbul ignore next */"},
			{Code: "/* jshint asi:true */"},
			{Code: "/* jscs: enable */"},
			{Code: "/* global var1, var2 */"},
			{Code: "/* global var1:true, var2 */"},
			{Code: "/* globals var1, var2 */"},
			{Code: "/* globals var1:true, var2 */"},
			{Code: "/* exported myVar */"},

			// ---- Ignores shebangs ----
			{Code: "#!foo"},
			{Code: "#!foo", Options: []any{"always"}},
			{Code: "#!Foo", Options: []any{"never"}},
			{Code: "#!/usr/bin/env node"},
			{Code: "#!/usr/bin/env node", Options: []any{"always"}},
			{Code: "#!/usr/bin/env node", Options: []any{"never"}},

			// ---- Using "always" string option ----
			{Code: "//Uppercase", Options: []any{"always"}},
			{Code: "// Uppercase", Options: []any{"always"}},
			{Code: "/*Uppercase */", Options: []any{"always"}},
			{Code: "/* Uppercase */", Options: []any{"always"}},
			{Code: "/*\nUppercase */", Options: []any{"always"}},
			{Code: "/** Uppercase */", Options: []any{"always"}},
			{Code: "/**\nUppercase */", Options: []any{"always"}},
			{Code: "//Über", Options: []any{"always"}},
			{Code: "//Π", Options: []any{"always"}},
			{Code: "/* Uppercase\nsecond line need not be uppercase */", Options: []any{"always"}},

			// ---- Using "always" string option: non-alphabetical is okay ----
			{Code: "//123", Options: []any{"always"}},
			{Code: "// 123", Options: []any{"always"}},
			{Code: "/*123*/", Options: []any{"always"}},
			{Code: "/**123*/", Options: []any{"always"}},
			{Code: "/* 123 */", Options: []any{"always"}},
			{Code: "/** 123*/", Options: []any{"always"}},
			{Code: "/**\n123*/", Options: []any{"always"}},
			{Code: "/*\n123 */", Options: []any{"always"}},
			{Code: "/*123\nsecond line need not be uppercase */", Options: []any{"always"}},
			{Code: "/**\n @todo: foobar\n */", Options: []any{"always"}},
			{Code: "/**\n * @fileoverview This is a file */", Options: []any{"always"}},

			// ---- Using "always" string option: eslint/istanbul/jshint/jscs/globals?/exported are okay ----
			{Code: "// jscs: enable", Options: []any{"always"}},
			{Code: "// jscs:disable", Options: []any{"always"}},
			{Code: "// eslint-disable-line", Options: []any{"always"}},
			{Code: "// eslint-disable-next-line", Options: []any{"always"}},
			{Code: "/* eslint semi:off */", Options: []any{"always"}},
			{Code: "/* eslint-enable */", Options: []any{"always"}},
			{Code: "/* istanbul ignore next */", Options: []any{"always"}},
			{Code: "/* jshint asi:true */", Options: []any{"always"}},
			{Code: "/* jscs: enable */", Options: []any{"always"}},
			{Code: "/* global var1, var2 */", Options: []any{"always"}},
			{Code: "/* global var1:true, var2 */", Options: []any{"always"}},
			{Code: "/* globals var1, var2 */", Options: []any{"always"}},
			{Code: "/* globals var1:true, var2 */", Options: []any{"always"}},
			{Code: "/* exported myVar */", Options: []any{"always"}},

			// ---- Using "never" string option ----
			{Code: "//lowercase", Options: []any{"never"}},
			{Code: "// lowercase", Options: []any{"never"}},
			{Code: "/*lowercase */", Options: []any{"never"}},
			{Code: "/* lowercase */", Options: []any{"never"}},
			{Code: "/*\nlowercase */", Options: []any{"never"}},
			{Code: "//über", Options: []any{"never"}},
			{Code: "//π", Options: []any{"never"}},
			{Code: "/* lowercase\nSecond line need not be lowercase */", Options: []any{"never"}},

			// ---- Using "never" string option: non-alphabetical is okay ----
			{Code: "//123", Options: []any{"never"}},
			{Code: "// 123", Options: []any{"never"}},
			{Code: "/*123*/", Options: []any{"never"}},
			{Code: "/* 123 */", Options: []any{"never"}},
			{Code: "/*\n123 */", Options: []any{"never"}},
			{Code: "/*123\nsecond line need not be uppercase */", Options: []any{"never"}},
			{Code: "/**\n @TODO: foobar\n */", Options: []any{"never"}},
			{Code: "/**\n * @Fileoverview This is a file */", Options: []any{"never"}},

			// ---- If first word in comment matches ignorePattern, don't warn ----
			{Code: "// matching", Options: []any{"always", map[string]any{"ignorePattern": "match"}}},
			{Code: "// Matching", Options: []any{"never", map[string]any{"ignorePattern": "Match"}}},
			{Code: "// bar", Options: []any{"always", map[string]any{"ignorePattern": "foo|bar"}}},
			{Code: "// Bar", Options: []any{"never", map[string]any{"ignorePattern": "Foo|Bar"}}},

			// ---- Inline comments are not warned if ignoreInlineComments: true ----
			{Code: "foo(/* ignored */ a);", Options: []any{"always", map[string]any{"ignoreInlineComments": true}}},
			{Code: "foo(/* Ignored */ a);", Options: []any{"never", map[string]any{"ignoreInlineComments": true}}},

			// ---- Inline comments can span multiple lines ----
			{Code: "foo(/*\nignored */ a);", Options: []any{"always", map[string]any{"ignoreInlineComments": true}}},
			{Code: "foo(/*\nIgnored */ a);", Options: []any{"never", map[string]any{"ignoreInlineComments": true}}},

			// ---- Tolerating consecutive comments ----
			{
				Code: "// This comment is valid since it is capitalized,\n" +
					"// and this one is valid since it follows a valid one,\n" +
					"// and same with this one.",
				Options: []any{"always", map[string]any{"ignoreConsecutiveComments": true}},
			},
			{
				Code: "/* This comment is valid since it is capitalized, */\n" +
					"/* and this one is valid since it follows a valid one, */\n" +
					"/* and same with this one. */",
				Options: []any{"always", map[string]any{"ignoreConsecutiveComments": true}},
			},
			{
				Code: "/*\n" +
					" * This comment is valid since it is capitalized,\n" +
					" */\n" +
					"/* and this one is valid since it follows a valid one, */\n" +
					"/*\n" +
					" * and same with this one.\n" +
					" */",
				Options: []any{"always", map[string]any{"ignoreConsecutiveComments": true}},
			},
			{
				Code: "// This comment is valid since it is capitalized,\n" +
					"// and this one is valid since it follows a valid one,\n" +
					"foo();\n" +
					"// This comment now has to be capitalized.",
				Options: []any{"always", map[string]any{"ignoreConsecutiveComments": true}},
			},

			// ---- Comments which start with URLs should always be valid ----
			{Code: "// https://github.com", Options: []any{"always"}},
			{Code: "// HTTPS://GITHUB.COM", Options: []any{"never"}},

			// ---- Using different options for line/block comments ----
			{
				Code: "// Valid capitalized line comment\n" +
					"/* Valid capitalized block comment */\n" +
					"// lineCommentIgnorePattern\n" +
					"/* blockCommentIgnorePattern */",
				Options: []any{
					"always",
					map[string]any{
						"line":  map[string]any{"ignorePattern": "lineCommentIgnorePattern"},
						"block": map[string]any{"ignorePattern": "blockCommentIgnorePattern"},
					},
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- No options: capitalization required ----
			{
				Code:   "//lowercase",
				Output: []string{"//Lowercase"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:   "// lowercase",
				Output: []string{"// Lowercase"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:   "/*lowercase */",
				Output: []string{"/*Lowercase */"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:   "/* lowercase */",
				Output: []string{"/* Lowercase */"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:   "/** lowercase */",
				Output: []string{"/** Lowercase */"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:   "/*\nlowercase */",
				Output: []string{"/*\nLowercase */"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:   "/**\nlowercase */",
				Output: []string{"/**\nLowercase */"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:   "//über",
				Output: []string{"//Über"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:   "//π",
				Output: []string{"//Π"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:   "/* lowercase\nSecond line need not be lowercase */",
				Output: []string{"/* Lowercase\nSecond line need not be lowercase */"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:   "// ꮳꮃꭹ",
				Output: []string{"// Ꮳꮃꭹ"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				// right-to-left-text
				Code:   "/* 𐳡𐳡𐳡 */",
				Output: []string{"/* 𐲡𐳡𐳡 */"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},

			// ---- Using "always" string option ----
			{
				Code:    "//lowercase",
				Output:  []string{"//Lowercase"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "// lowercase",
				Output:  []string{"// Lowercase"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "/*lowercase */",
				Output:  []string{"/*Lowercase */"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "/* lowercase */",
				Output:  []string{"/* Lowercase */"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "/** lowercase */",
				Output:  []string{"/** Lowercase */"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "/**\nlowercase */",
				Output:  []string{"/**\nLowercase */"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "//über",
				Output:  []string{"//Über"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "//π",
				Output:  []string{"//Π"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "/* lowercase\nsecond line need not be uppercase */",
				Output:  []string{"/* Lowercase\nsecond line need not be uppercase */"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},

			// ---- Using "never" string option ----
			{
				Code:    "//Uppercase",
				Output:  []string{"//uppercase"},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUppercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "// Uppercase",
				Output:  []string{"// uppercase"},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUppercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "/*Uppercase */",
				Output:  []string{"/*uppercase */"},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUppercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "/* Uppercase */",
				Output:  []string{"/* uppercase */"},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUppercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "/*\nUppercase */",
				Output:  []string{"/*\nuppercase */"},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUppercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "//Über",
				Output:  []string{"//über"},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUppercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "//Π",
				Output:  []string{"//π"},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUppercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "/* Uppercase\nsecond line need not be uppercase */",
				Output:  []string{"/* uppercase\nsecond line need not be uppercase */"},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUppercaseComment", Line: 1, Column: 1}},
			},
			{
				// Georgian Mtavruli Capital Letter Gan (U+1C92) -> Georgian Letter Gan (U+10D2)
				Code:    "// Გ",
				Output:  []string{"// გ"},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUppercaseComment", Line: 1, Column: 1}},
			},
			{
				// Warang Citi Capital Letter Wi (U+118A2) -> Warang Citi Small Letter Wi (U+118C2)
				Code:    "// 𑢢",
				Output:  []string{"// 𑣂"},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUppercaseComment", Line: 1, Column: 1}},
			},

			// ---- Default ignore words should be warned if there are non-whitespace characters in the way ----
			{
				Code:    "//* jscs: enable",
				Output:  []string{"//* Jscs: enable"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "//* jscs:disable",
				Output:  []string{"//* Jscs:disable"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "//* eslint-disable-line",
				Output:  []string{"//* Eslint-disable-line"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "//* eslint-disable-next-line",
				Output:  []string{"//* Eslint-disable-next-line"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "/*\n * eslint semi:off */",
				Output:  []string{"/*\n * Eslint semi:off */"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "/*\n * eslint-env node */",
				Output:  []string{"/*\n * Eslint-env node */"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "/*\n *  istanbul ignore next */",
				Output:  []string{"/*\n *  Istanbul ignore next */"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "/*\n *  jshint asi:true */",
				Output:  []string{"/*\n *  Jshint asi:true */"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "/*\n *  jscs: enable */",
				Output:  []string{"/*\n *  Jscs: enable */"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "/*\n *  global var1, var2 */",
				Output:  []string{"/*\n *  Global var1, var2 */"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "/*\n *  global var1:true, var2 */",
				Output:  []string{"/*\n *  Global var1:true, var2 */"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "/*\n *  globals var1, var2 */",
				Output:  []string{"/*\n *  Globals var1, var2 */"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "/*\n *  globals var1:true, var2 */",
				Output:  []string{"/*\n *  Globals var1:true, var2 */"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "/*\n *  exported myVar */",
				Output:  []string{"/*\n *  Exported myVar */"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},

			// ---- Inline comments should be warned if ignoreInlineComments is omitted or false ----
			{
				Code:    "foo(/* invalid */a);",
				Output:  []string{"foo(/* Invalid */a);"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 5}},
			},
			{
				Code:    "foo(/* invalid */a);",
				Output:  []string{"foo(/* Invalid */a);"},
				Options: []any{"always", map[string]any{"ignoreInlineComments": false}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 5}},
			},

			// ---- ignoreInlineComments should only allow inline comments to pass ----
			{
				Code:    "foo(a, // not an inline comment\nb);",
				Output:  []string{"foo(a, // Not an inline comment\nb);"},
				Options: []any{"always", map[string]any{"ignoreInlineComments": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 8}},
			},
			{
				Code:    "foo(a, /* not an inline comment */\nb);",
				Output:  []string{"foo(a, /* Not an inline comment */\nb);"},
				Options: []any{"always", map[string]any{"ignoreInlineComments": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 8}},
			},
			{
				Code:    "foo(a,\n/* not an inline comment */b);",
				Output:  []string{"foo(a,\n/* Not an inline comment */b);"},
				Options: []any{"always", map[string]any{"ignoreInlineComments": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 2, Column: 1}},
			},
			{
				Code:    "foo(a,\n/* not an inline comment */\nb);",
				Output:  []string{"foo(a,\n/* Not an inline comment */\nb);"},
				Options: []any{"always", map[string]any{"ignoreInlineComments": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 2, Column: 1}},
			},
			{
				Code:    "foo(a, // Not an inline comment\nb);",
				Output:  []string{"foo(a, // not an inline comment\nb);"},
				Options: []any{"never", map[string]any{"ignoreInlineComments": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUppercaseComment", Line: 1, Column: 8}},
			},
			{
				Code:    "foo(a, /* Not an inline comment */\nb);",
				Output:  []string{"foo(a, /* not an inline comment */\nb);"},
				Options: []any{"never", map[string]any{"ignoreInlineComments": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUppercaseComment", Line: 1, Column: 8}},
			},
			{
				Code:    "foo(a,\n/* Not an inline comment */b);",
				Output:  []string{"foo(a,\n/* not an inline comment */b);"},
				Options: []any{"never", map[string]any{"ignoreInlineComments": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUppercaseComment", Line: 2, Column: 1}},
			},
			{
				Code:    "foo(a,\n/* Not an inline comment */\nb);",
				Output:  []string{"foo(a,\n/* not an inline comment */\nb);"},
				Options: []any{"never", map[string]any{"ignoreInlineComments": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUppercaseComment", Line: 2, Column: 1}},
			},

			// ---- Comments which do not match ignorePattern are still warned ----
			{
				Code:    "// not matching",
				Output:  []string{"// Not matching"},
				Options: []any{"always", map[string]any{"ignorePattern": "ignored?"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "// Not matching",
				Output:  []string{"// not matching"},
				Options: []any{"never", map[string]any{"ignorePattern": "ignored?"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUppercaseComment", Line: 1, Column: 1}},
			},

			// ---- ignoreConsecutiveComments only applies to comments with no tokens between them ----
			{
				Code: "// This comment is valid since it is capitalized,\n" +
					"// and this one is valid since it follows a valid one,\n" +
					"foo();\n" +
					"// this comment is now invalid.",
				Output: []string{
					"// This comment is valid since it is capitalized,\n" +
						"// and this one is valid since it follows a valid one,\n" +
						"foo();\n" +
						"// This comment is now invalid.",
				},
				Options: []any{"always", map[string]any{"ignoreConsecutiveComments": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 4, Column: 1}},
			},

			// ---- Only the initial comment should warn if ignoreConsecutiveComments:true ----
			{
				Code: "// this comment is invalid since it is not capitalized,\n" +
					"// but this one is ignored since it is consecutive.",
				Output: []string{
					"// This comment is invalid since it is not capitalized,\n" +
						"// but this one is ignored since it is consecutive.",
				},
				Options: []any{"always", map[string]any{"ignoreConsecutiveComments": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code: "// This comment is invalid since it is not capitalized,\n" +
					"// But this one is ignored since it is consecutive.",
				Output: []string{
					"// this comment is invalid since it is not capitalized,\n" +
						"// But this one is ignored since it is consecutive.",
				},
				Options: []any{"never", map[string]any{"ignoreConsecutiveComments": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUppercaseComment", Line: 1, Column: 1}},
			},

			// ---- Consecutive comments should warn if ignoreConsecutiveComments:false ----
			{
				Code: "// This comment is valid since it is capitalized,\n" +
					"// but this one is invalid even if it follows a valid one.",
				Output: []string{
					"// This comment is valid since it is capitalized,\n" +
						"// But this one is invalid even if it follows a valid one.",
				},
				Options: []any{"always", map[string]any{"ignoreConsecutiveComments": false}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 2, Column: 1}},
			},

			// ---- Comments are warned if URL is not at the start of the comment ----
			{
				Code:    "// should fail. https://github.com",
				Output:  []string{"// Should fail. https://github.com"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "// Should fail. https://github.com",
				Output:  []string{"// should fail. https://github.com"},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUppercaseComment", Line: 1, Column: 1}},
			},
		},
	)
}
