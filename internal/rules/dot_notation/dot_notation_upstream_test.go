// TestDotNotationUpstream migrates the full valid/invalid suite from
// upstream eslint/tests/lib/rules/dot-notation.js 1:1. Position assertions
// cover line/column for every invalid case. rslint-specific lock-in cases
// live in the dot_notation_extras_test.go file.
package dot_notation

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestDotNotationUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &DotNotationRule, []rule_tester.ValidTestCase{
		{Code: "a.b;"},
		{Code: "a.b.c;"},
		{Code: "a['12'];"},
		{Code: "a[b];"},
		{Code: "a[0];"},
		{Code: "a.b.c;", Options: map[string]interface{}{"allowKeywords": false}},
		{Code: "a.arguments;", Options: map[string]interface{}{"allowKeywords": false}},
		{Code: "a.let;", Options: map[string]interface{}{"allowKeywords": false}},
		{Code: "a.yield;", Options: map[string]interface{}{"allowKeywords": false}},
		{Code: "a.eval;", Options: map[string]interface{}{"allowKeywords": false}},
		{Code: "a[0];", Options: map[string]interface{}{"allowKeywords": false}},
		{Code: "a['while'];", Options: map[string]interface{}{"allowKeywords": false}},
		{Code: "a['true'];", Options: map[string]interface{}{"allowKeywords": false}},
		{Code: "a['null'];", Options: map[string]interface{}{"allowKeywords": false}},
		{Code: "a[true];", Options: map[string]interface{}{"allowKeywords": false}},
		{Code: "a[null];", Options: map[string]interface{}{"allowKeywords": false}},
		{Code: "a.true;", Options: map[string]interface{}{"allowKeywords": true}},
		{Code: "a.null;", Options: map[string]interface{}{"allowKeywords": true}},
		{Code: "a['snake_case'];", Options: map[string]interface{}{"allowPattern": "^[a-z]+(_[a-z]+)+$"}},
		{Code: "a['lots_of_snake_case'];", Options: map[string]interface{}{"allowPattern": "^[a-z]+(_[a-z]+)+$"}},
		// Template literal WITH substitution is not a static key.
		{Code: "a[`time${range}`];"},
		// Static template literal matching a keyword: NOT converted to dot
		// when allowKeywords is false (bracket notation is required there).
		{Code: "a[`while`];", Options: map[string]interface{}{"allowKeywords": false}},
		// Static template literal whose value isn't a valid identifier.
		{Code: "a[`time range`];"},
		{Code: "a.true;"},
		{Code: "a.null;"},
		{Code: "a[undefined];"},
		{Code: "a[void 0];"},
		{Code: "a[b()];"},
		{Code: "a[/(?<zero>0)/];"},
		{Code: "class C { foo() { this['#a'] } }"},
		{Code: "class C { #in; foo() { this.#in; } }", Options: map[string]interface{}{"allowKeywords": false}},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "a.true;",
			Output:  []string{`a["true"];`},
			Options: map[string]interface{}{"allowKeywords": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "useBrackets", Message: `.true is a syntax error.`, Line: 1, Column: 3}},
		},
		{
			Code:   "a['true'];",
			Output: []string{"a.true;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Message: `["true"] is better written in dot notation.`, Line: 1, Column: 3}},
		},
		{
			Code:   "a[`time`];",
			Output: []string{"a.time;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Message: "[`time`] is better written in dot notation.", Line: 1, Column: 3}},
		},
		{
			Code:   "a[null];",
			Output: []string{"a.null;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Message: "[null] is better written in dot notation.", Line: 1, Column: 3}},
		},
		{
			Code:   "a[true];",
			Output: []string{"a.true;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Message: "[true] is better written in dot notation.", Line: 1, Column: 3}},
		},
		{
			Code:   "a[false];",
			Output: []string{"a.false;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Message: "[false] is better written in dot notation.", Line: 1, Column: 3}},
		},
		{
			Code:   "a['b'];",
			Output: []string{"a.b;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Message: `["b"] is better written in dot notation.`, Line: 1, Column: 3, EndLine: 1, EndColumn: 6}},
		},
		{
			Code:   "a.b['c'];",
			Output: []string{"a.b.c;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 5}},
		},
		{
			Code:    "a['_dangle'];",
			Output:  []string{"a._dangle;"},
			Options: map[string]interface{}{"allowPattern": "^[a-z]+(_[a-z]+)+$"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 3}},
		},
		{
			Code:    "a['SHOUT_CASE'];",
			Output:  []string{"a.SHOUT_CASE;"},
			Options: map[string]interface{}{"allowPattern": "^[a-z]+(_[a-z]+)+$"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 3}},
		},
		{
			Code:   "a\n  ['SHOUT_CASE'];",
			Output: []string{"a\n  .SHOUT_CASE;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 2, Column: 4, EndLine: 2, EndColumn: 16}},
		},
		{
			Code: "getResource()\n" +
				"    .then(function(){})\n" +
				"    [\"catch\"](function(){})\n" +
				"    .then(function(){})\n" +
				"    [\"catch\"](function(){});",
			Output: []string{
				"getResource()\n" +
					"    .then(function(){})\n" +
					"    .catch(function(){})\n" +
					"    .then(function(){})\n" +
					"    .catch(function(){});",
			},
			// Diagnostics are reported in AST-traversal order (outer node
			// before the nested one it contains), not sorted by source
			// position - the line-5 access is the outer node of the chain,
			// so it's visited (and reported) before the line-3 access nested
			// inside it. Each fix replaces only its own bracket part, so the
			// two ranges don't overlap and both apply in a single pass.
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot", Line: 5, Column: 6},
				{MessageId: "useDot", Line: 3, Column: 6},
			},
		},
		{
			Code:    "foo\n  .while;",
			Output:  []string{"foo\n  [\"while\"];"},
			Options: map[string]interface{}{"allowKeywords": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "useBrackets", Line: 2, Column: 4}},
		},
		{
			// Not fixed: a comment lives inside the brackets.
			Code:   "foo[ /* comment */ 'bar' ]",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 20}},
		},
		{
			// Not fixed: a comment lives inside the brackets.
			Code:   "foo[ 'bar' /* comment */ ]",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 6}},
		},
		{
			Code:   "foo[    'bar'    ];",
			Output: []string{"foo.bar;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 9}},
		},
		{
			// Not fixed: a comment lives between the dot and the property name.
			Code:    "foo. /* comment */ while",
			Options: map[string]interface{}{"allowKeywords": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "useBrackets", Line: 1, Column: 20}},
		},
		{
			Code:   "foo[('bar')]",
			Output: []string{"foo.bar"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 6}},
		},
		{
			Code:   "foo[(null)]",
			Output: []string{"foo.null"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 6}},
		},
		{
			Code:   "(foo)['bar']",
			Output: []string{"(foo).bar"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 7}},
		},
		{
			// ---- Dimension 4 / numeric literal autofix boundary ----
			// Decimal integer object: a space must separate it from the
			// inserted `.` so the fixed text doesn't read as a fractional
			// number literal.
			Code:   "1['toString']",
			Output: []string{"1 .toString"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 3}},
		},
		{
			// Identifier-fusion guard: the fixed `.bar` must not run directly
			// into the following `instanceof` keyword.
			Code:   "foo['bar']instanceof baz",
			Output: []string{"foo.bar instanceof baz"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 5}},
		},
		{
			// `let[` is parsed as a destructuring declaration, so the fixer
			// must not turn this into `let["if"]()`.
			Code:    "let.if()",
			Options: map[string]interface{}{"allowKeywords": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "useBrackets", Line: 1, Column: 5}},
		},
		{
			Code:   "5['prop']",
			Output: []string{"5 .prop"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 3}},
		},
		{
			Code:   "-5['prop']",
			Output: []string{"-5 .prop"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 4}},
		},
		{
			// SKIP: rslint does not support legacy octal-style numeric
			// literals (a leading zero followed by octal digits, e.g. `0123`).
			// tsgo's parser rejects this as a hard syntax error (TS1121 "Octal
			// literals are not allowed. Use the syntax '0o1'.") even in
			// non-strict scripts, so the file can't even be parsed - this is a
			// framework gap, not a rule-semantic one.
			Code:   "01['prop']",
			Output: []string{"01.prop"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 4}},
			Skip:   true,
		},
		{
			// SKIP: same framework gap as above.
			Code:   "01234567['prop']",
			Output: []string{"01234567.prop"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 10}},
			Skip:   true,
		},
		{
			// SKIP: a leading-zero decimal containing an 8/9 digit (not a
			// valid legacy octal) is TS1489 "Decimals with leading zeros are
			// not allowed." under tsgo - also a hard parse error.
			Code:   "08['prop']",
			Output: []string{"08 .prop"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 4}},
			Skip:   true,
		},
		{
			// SKIP: same framework gap as above.
			Code:   "090['prop']",
			Output: []string{"090 .prop"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 5}},
			Skip:   true,
		},
		{
			// SKIP: same framework gap as above.
			Code:   "018['prop']",
			Output: []string{"018 .prop"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 5}},
			Skip:   true,
		},
		{
			Code:   "5_000['prop']",
			Output: []string{"5_000 .prop"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 7}},
		},
		{
			Code:   "5_000_00['prop']",
			Output: []string{"5_000_00 .prop"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 10}},
		},
		{
			// Not a decimal-integer literal (contains a `.`): no space.
			Code:   "5.000_000['prop']",
			Output: []string{"5.000_000.prop"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 11}},
		},
		{
			// Binary literal: no ambiguity, no space.
			Code:   "0b1010_1010['prop']",
			Output: []string{"0b1010_1010.prop"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 13}},
		},
		{
			// Optional chaining bracket access.
			Code:   "obj?.['prop']",
			Output: []string{"obj?.prop"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 7}},
		},
		{
			Code:   "0?.['prop']",
			Output: []string{"0?.prop"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 5}},
		},
		{
			Code:    "obj?.true",
			Output:  []string{`obj?.["true"]`},
			Options: map[string]interface{}{"allowKeywords": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "useBrackets", Line: 1, Column: 6}},
		},
		{
			// Optional access on `let` is NOT the destructuring-declaration
			// shape, so the fix IS applied here (unlike `let.if()` above).
			Code:    "let?.true",
			Output:  []string{`let?.["true"]`},
			Options: map[string]interface{}{"allowKeywords": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "useBrackets", Line: 1, Column: 6}},
		},
	})
}
