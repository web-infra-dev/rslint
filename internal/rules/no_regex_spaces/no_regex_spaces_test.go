package no_regex_spaces

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoRegexSpacesRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoRegexSpacesRule,
		[]rule_tester.ValidTestCase{
			// ---- Baseline: no consecutive spaces ----
			{Code: `var foo = /foo/;`},
			{Code: `var foo = RegExp('foo')`},
			{Code: `var foo = / /;`},
			{Code: `var foo = RegExp(' ')`},
			{Code: `var foo = / a b c d /;`},

			// ---- Single space followed by explicit quantifier ----
			{Code: `var foo = /bar {3}baz/g;`},
			{Code: `var foo = RegExp('bar {3}baz', 'g')`},
			{Code: `var foo = new RegExp('bar {3}baz')`},
			{Code: `var foo = /  +/;`},
			{Code: `var foo = /  ?/;`},
			{Code: `var foo = /  */;`},
			{Code: `var foo = /  {2}/;`},

			// ---- Tabs / non-space whitespace don't count ----
			{Code: "var foo = /bar\t\t\tbaz/;"},
			{Code: "var foo = RegExp('bar\t\t\tbaz');"},
			{Code: "var foo = new RegExp('bar\t\t\tbaz');"},

			// ---- RegExp is shadowed in the enclosing scope ----
			{Code: `var RegExp = function() {}; var foo = new RegExp('bar   baz');`},
			{Code: `var RegExp = function() {}; var foo = RegExp('bar   baz');`},

			// ---- No consecutive spaces in the source code ----
			{Code: `var foo = /bar \ baz/;`},
			{Code: `var foo = /bar\ \ baz/;`},
			{Code: `var foo = /bar \u0020 baz/;`},
			{Code: `var foo = /bar\u0020\u0020baz/;`},
			{Code: `var foo = new RegExp('bar \ baz')`},
			{Code: `var foo = new RegExp('bar\ \ baz')`},
			{Code: `var foo = new RegExp('bar \\ baz')`},
			{Code: `var foo = new RegExp('bar \u0020 baz')`},
			{Code: `var foo = new RegExp('bar\u0020\u0020baz')`},
			{Code: `var foo = new RegExp('bar \\u0020 baz')`},

			// ---- Spaces inside character classes ----
			{Code: `var foo = /[  ]/;`},
			{Code: `var foo = /[   ]/;`},
			{Code: `var foo = / [  ] /;`},
			{Code: `var foo = / [  ] [  ] /;`},
			{Code: `var foo = new RegExp('[  ]');`},
			{Code: `var foo = new RegExp('[   ]');`},
			{Code: `var foo = new RegExp(' [  ] ');`},
			{Code: `var foo = RegExp(' [  ] [  ] ');`},
			{Code: `var foo = new RegExp(' \[   ');`},
			{Code: `var foo = new RegExp(' \[   \] ');`},

			// ---- ES2024 (v flag) ----
			{Code: `var foo = /  {2}/v;`},
			{Code: `var foo = /[\q{    }]/v;`},

			// ---- Syntactically invalid patterns — ESLint skips via parse error ----
			{Code: `var foo = new RegExp('[  ');`},
			{Code: `var foo = new RegExp('{  ', 'u');`},
			{Code: `var foo = new RegExp('{  ', 'v');`},

			// ---- Flags cannot be determined ----
			{Code: `new RegExp('  ', flags)`},
			{Code: `new RegExp('[[abc]  ]', flags + 'v')`},
			{Code: `new RegExp('[[abc]\q{  }]', flags + 'v')`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Regex literals ----
			{
				Code:   `var foo = /bar  baz/;`,
				Output: []string{`var foo = /bar {2}baz/;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},
			{
				Code:   `var foo = /bar    baz/;`,
				Output: []string{`var foo = /bar {4}baz/;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},
			{
				Code:   `var foo = / a b  c d /;`,
				Output: []string{`var foo = / a b {2}c d /;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},

			// ---- RegExp constructor ----
			{
				Code:   `var foo = RegExp(' a b c d  ');`,
				Output: []string{`var foo = RegExp(' a b c d {2}');`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},
			{
				Code:   `var foo = RegExp('bar    baz');`,
				Output: []string{`var foo = RegExp('bar {4}baz');`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},
			{
				Code:   `var foo = new RegExp('bar    baz');`,
				Output: []string{`var foo = new RegExp('bar {4}baz');`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},

			// ---- RegExp not shadowed where it's called ----
			{
				Code:   `{ let RegExp = function() {}; } var foo = RegExp('bar    baz');`,
				Output: []string{`{ let RegExp = function() {}; } var foo = RegExp('bar {4}baz');`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 43},
				},
			},

			// ---- Space runs followed by a quantifier — trailing space is quantified ----
			{
				Code:   `var foo = /bar   {3}baz/;`,
				Output: []string{`var foo = /bar {2} {3}baz/;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},
			{
				Code:   `var foo = /bar    ?baz/;`,
				Output: []string{`var foo = /bar {3} ?baz/;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},
			{
				Code:   `var foo = new RegExp('bar   *baz')`,
				Output: []string{`var foo = new RegExp('bar {2} *baz')`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},
			{
				Code:   `var foo = RegExp('bar   +baz')`,
				Output: []string{`var foo = RegExp('bar {2} +baz')`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},
			{
				Code:   `var foo = new RegExp('bar    ');`,
				Output: []string{`var foo = new RegExp('bar {4}');`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},

			// ---- Escaped backslash + spaces in regex literal ----
			{
				Code:   `var foo = /bar\  baz/;`,
				Output: []string{`var foo = /bar\ {2}baz/;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},

			// ---- Spaces outside character classes ----
			{
				Code:   `var foo = /[   ]  /;`,
				Output: []string{`var foo = /[   ] {2}/;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},
			{
				Code:   `var foo = /  [   ] /;`,
				Output: []string{`var foo = / {2}[   ] /;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},
			{
				Code:   `var foo = new RegExp('[   ]  ');`,
				Output: []string{`var foo = new RegExp('[   ] {2}');`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},
			{
				Code:   `var foo = RegExp('  [ ]');`,
				Output: []string{`var foo = RegExp(' {2}[ ]');`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},

			// ---- Escaped brackets don't open character classes ----
			{
				Code:   `var foo = /\[  /;`,
				Output: []string{`var foo = /\[ {2}/;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},
			{
				Code:   `var foo = /\[  \]/;`,
				Output: []string{`var foo = /\[ {2}\]/;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},

			// ---- Non-capturing groups and assertions ----
			{
				Code:   `var foo = /(?:  )/;`,
				Output: []string{`var foo = /(?: {2})/;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},
			{
				Code:   `var foo = RegExp('^foo(?=   )');`,
				Output: []string{`var foo = RegExp('^foo(?= {3})');`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},

			// ---- Escape of the space character ----
			{
				Code:   `var foo = /\  /`,
				Output: []string{`var foo = /\ {2}/`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},
			{
				Code:   `var foo = / \  /`,
				Output: []string{`var foo = / \ {2}/`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},

			// ---- Report only the first occurrence per pass; the rule tester
			// re-runs the fix until stable, so a second pass catches the later
			// run that the first pass left intact. Errors asserts against the
			// FIRST-pass diagnostics only.
			{
				Code: `var foo = /  foo   /;`,
				Output: []string{
					`var foo = / {2}foo   /;`,
					`var foo = / {2}foo {3}/;`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},

			// ---- Strings containing escape sequences — report but no fix ----
			{
				Code: `var foo = new RegExp('\\d  ')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},
			{
				Code: `var foo = RegExp('\u0041   ')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},
			{
				Code: `var foo = new RegExp('\\[  \\]');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},

			// ---- ES2024 v-flag: nested character classes ----
			{
				Code:   `var foo = /[[    ]    ]    /v;`,
				Output: []string{`var foo = /[[    ]    ] {4}/v;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},
			{
				Code:   `var foo = new RegExp('[[    ]    ]    ', 'v');`,
				Output: []string{`var foo = new RegExp('[[    ]    ] {4}', 'v');`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSpaces", Line: 1, Column: 11},
				},
			},
		},
	)
}
