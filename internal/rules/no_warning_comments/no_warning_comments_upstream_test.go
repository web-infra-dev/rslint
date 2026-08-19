package no_warning_comments

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoWarningCommentsUpstream migrates the full valid/invalid suite from
// upstream eslint/tests/lib/rules/no-warning-comments.js 1:1. Position
// assertions cover line/column for every invalid case. rslint-specific
// lock-in cases live in the no_warning_comments_extras_test.go file.
func TestNoWarningCommentsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoWarningCommentsRule,
		[]rule_tester.ValidTestCase{
			{Code: "// any comment", Options: map[string]any{"terms": []interface{}{"fixme"}}},
			{Code: "// any comment", Options: map[string]any{"terms": []interface{}{"fixme", "todo"}}},
			{Code: "// any comment"},
			{Code: "// any comment", Options: map[string]any{"location": "anywhere"}},
			{Code: "// any comment with TODO, FIXME or XXX", Options: map[string]any{"location": "start"}},
			{Code: "// any comment with TODO, FIXME or XXX"},
			{Code: "/* any block comment */", Options: map[string]any{"terms": []interface{}{"fixme"}}},
			{Code: "/* any block comment */", Options: map[string]any{"terms": []interface{}{"fixme", "todo"}}},
			{Code: "/* any block comment */"},
			{Code: "/* any block comment */", Options: map[string]any{"location": "anywhere"}},
			{Code: "/* any block comment with TODO, FIXME or XXX */", Options: map[string]any{"location": "start"}},
			{Code: "/* any block comment with TODO, FIXME or XXX */"},
			{Code: "/* any block comment with (TODO, FIXME's or XXX!) */"},
			{
				Code:    "// comments containing terms as substrings like TodoMVC",
				Options: map[string]any{"terms": []interface{}{"todo"}, "location": "anywhere"},
			},
			{
				Code:    "// special regex characters don't cause a problem",
				Options: map[string]any{"terms": []interface{}{"[aeiou]"}, "location": "anywhere"},
			},
			{Code: "/*eslint no-warning-comments: [2, { \"terms\": [\"todo\", \"fixme\", \"any other term\"], \"location\": \"anywhere\" }]*/\n\nvar x = 10;\n"},
			{
				Code:    "/*eslint no-warning-comments: [2, { \"terms\": [\"todo\", \"fixme\", \"any other term\"], \"location\": \"anywhere\" }]*/\n\nvar x = 10;\n",
				Options: map[string]any{"location": "anywhere"},
			},
			{Code: "// foo", Options: map[string]any{"terms": []interface{}{"foo-bar"}}},
			{Code: "/** multi-line block comment with lines starting with\nTODO\nFIXME or\nXXX\n*/"},
			{Code: "//!TODO ", Options: map[string]any{"decoration": []interface{}{"*"}}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "// fixme",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				},
			},
			{
				Code:    "// any fixme",
				Options: map[string]any{"location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    "// any fixme",
				Options: map[string]any{"terms": []interface{}{"fixme"}, "location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    "// any FIXME",
				Options: map[string]any{"terms": []interface{}{"fixme"}, "location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    "// any fIxMe",
				Options: map[string]any{"terms": []interface{}{"fixme"}, "location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    "/* any fixme */",
				Options: map[string]any{"terms": []interface{}{"FIXME"}, "location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    "/* any FIXME */",
				Options: map[string]any{"terms": []interface{}{"FIXME"}, "location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    "/* any fIxMe */",
				Options: map[string]any{"terms": []interface{}{"FIXME"}, "location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    "// any fixme or todo",
				Options: map[string]any{"terms": []interface{}{"fixme", "todo"}, "location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    "/* any fixme or todo */",
				Options: map[string]any{"terms": []interface{}{"fixme", "todo"}, "location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "/* any fixme or todo */",
				Options: map[string]any{"location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code: "/* fixme and todo */",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    "/* fixme and todo */",
				Options: map[string]any{"location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    "/* any fixme */",
				Options: map[string]any{"location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    "/* fixme! */",
				Options: map[string]any{"terms": []interface{}{"fixme"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    "// regex [litera|$]",
				Options: map[string]any{"terms": []interface{}{"[litera|$]"}, "location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "/* eslint one-var: 2 */",
				Options: map[string]any{"terms": []interface{}{"eslint"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "/* eslint one-var: 2 */",
				Options: map[string]any{"terms": []interface{}{"one"}, "location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "/* any block comment with TODO, FIXME or XXX */",
				Options: map[string]any{"location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 48},
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 48},
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 48},
				},
			},
			{
				Code:    "/* any block comment with (TODO, FIXME's or XXX!) */",
				Options: map[string]any{"location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 53},
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 53},
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 53},
				},
			},
			{
				Code:    "/** \n *any block comment \n*with (TODO, FIXME's or XXX!) **/",
				Options: map[string]any{"location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 3, EndColumn: 34},
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 3, EndColumn: 34},
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 3, EndColumn: 34},
				},
			},
			{
				Code:    "// any comment with TODO, FIXME or XXX",
				Options: map[string]any{"location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 39},
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 39},
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 39},
				},
			},
			{
				Code:    "// TODO: something small",
				Options: map[string]any{"location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    "// TODO: something really longer than 40 characters",
				Options: map[string]any{"location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 52},
				},
			},
			{
				Code:    "/* TODO: something \n really longer than 40 characters \n and also a new line */",
				Options: map[string]any{"location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 3, EndColumn: 24},
				},
			},
			{
				Code:    "// TODO: small",
				Options: map[string]any{"location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    "// https://github.com/eslint/eslint/pull/13522#discussion_r470293411 TODO",
				Options: map[string]any{"location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 74},
				},
			},
			{
				Code:    "// Comment ending with term followed by punctuation TODO!",
				Options: map[string]any{"terms": []interface{}{"todo"}, "location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 58},
				},
			},
			{
				Code:    "// Comment ending with term including punctuation TODO!",
				Options: map[string]any{"terms": []interface{}{"todo!"}, "location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 56},
				},
			},
			{
				Code:    "// Comment ending with term including punctuation followed by more TODO!!!",
				Options: map[string]any{"terms": []interface{}{"todo!"}, "location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 75},
				},
			},
			{
				Code:    "// !TODO comment starting with term preceded by punctuation",
				Options: map[string]any{"terms": []interface{}{"todo"}, "location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 60},
				},
			},
			{
				Code:    "// !TODO comment starting with term including punctuation",
				Options: map[string]any{"terms": []interface{}{"!todo"}, "location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 58},
				},
			},
			{
				Code:    "// !!!TODO comment starting with term including punctuation preceded by more",
				Options: map[string]any{"terms": []interface{}{"!todo"}, "location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 77},
				},
			},
			{
				Code:    "// FIX!term ending with punctuation followed word character",
				Options: map[string]any{"terms": []interface{}{"FIX!"}, "location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 60},
				},
			},
			{
				Code:    "// Term starting with punctuation preceded word character!FIX",
				Options: map[string]any{"terms": []interface{}{"!FIX"}, "location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 62},
				},
			},
			{
				Code:    "//!XXX comment starting with no spaces (anywhere)",
				Options: map[string]any{"terms": []interface{}{"!xxx"}, "location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 50},
				},
			},
			{
				Code:    "//!XXX comment starting with no spaces (start)",
				Options: map[string]any{"terms": []interface{}{"!xxx"}, "location": "start"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 47},
				},
			},
			{
				Code:    "/*\nTODO undecorated multi-line block comment (start)\n*/",
				Options: map[string]any{"terms": []interface{}{"todo"}, "location": "start"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 3, EndColumn: 3},
				},
			},
			{
				Code: "///// TODO decorated single-line comment with decoration array \n /////",
				Options: map[string]any{
					"terms": []interface{}{"todo"}, "location": "start",
					"decoration": []interface{}{"*", "/"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 64},
				},
			},
			{
				Code: "///*/*/ TODO decorated single-line comment with multiple decoration characters (start) \n /////",
				Options: map[string]any{
					"terms": []interface{}{"todo"}, "location": "start",
					"decoration": []interface{}{"*", "/"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 88},
				},
			},
			{
				Code: "//**TODO term starts with a decoration character",
				Options: map[string]any{
					"terms": []interface{}{"*todo"}, "location": "start",
					"decoration": []interface{}{"*"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 49},
				},
			},
		},
	)
}
