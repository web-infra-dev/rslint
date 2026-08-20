package max_statements

import (
	"fmt"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestMaxStatementsUpstream migrates the full valid/invalid suite from upstream
// eslint/tests/lib/rules/max-statements.js 1:1. Position assertions cover
// line/column for every invalid case. rslint-specific lock-in cases live in
// the max_statements_extras_test.go file.
func TestMaxStatementsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&MaxStatementsRule,
		[]rule_tester.ValidTestCase{
			{Code: "function foo() { var bar = 1; function qux () { var noCount = 2; } return 3; }", Options: option(3)},
			{Code: "function foo() { var bar = 1; if (true) { for (;;) { var qux = null; } } else { quxx(); } return 3; }", Options: option(6)},
			{Code: "function foo() { var x = 5; function bar() { var y = 6; } bar(); z = 10; baz(); }", Options: option(5)},
			{Code: "function foo() { var a; var b; var c; var x; var y; var z; bar(); baz(); qux(); quxx(); }"},
			{Code: "(function() { var bar = 1; return function () { return 42; }; })()", Options: optionsWithTopLevel(1, true)},
			{Code: "function foo() { var bar = 1; var baz = 2; }", Options: optionsWithTopLevel(1, true)},
			{Code: "define(['foo', 'qux'], function(foo, qux) { var bar = 1; var baz = 2; })", Options: optionsWithTopLevel(1, true)},

			// ---- object property options ----
			{Code: "var foo = { thing: function() { var bar = 1; var baz = 2; } }", Options: option(2)},
			{Code: "var foo = { thing() { var bar = 1; var baz = 2; } }", Options: option(2)},
			{Code: "var foo = { ['thing']() { var bar = 1; var baz = 2; } }", Options: option(2)},
			{Code: "var foo = { thing: () => { var bar = 1; var baz = 2; } }", Options: option(2)},
			{Code: "var foo = { thing: function() { var bar = 1; var baz = 2; } }", Options: option(map[string]interface{}{"max": 2})},

			// ---- this rule does not apply to class static blocks, and statements in them should not count as statements in the enclosing function ----
			{Code: "class C { static { one; two; three; { four; five; six; } } }", Options: option(2)},
			{Code: "function foo() { class C { static { one; two; three; { four; five; six; } } } }", Options: option(2)},
			{Code: "class C { static { one; two; three; function foo() { 1; 2; } four; five; six; } }", Options: option(2)},
			{Code: "class C { static { { one; two; three; function foo() { 1; 2; } four; five; six; } } }", Options: option(2)},
			{Code: "function top_level() { 1; /* 2 */ class C { static { one; two; three; { four; five; six; } } } 3;}", Options: optionsWithTopLevel(2, true)},
			{Code: "function top_level() { 1; 2; } class C { static { one; two; three; { four; five; six; } } }", Options: optionsWithTopLevel(1, true)},
			{Code: "class C { static { one; two; three; { four; five; six; } } } function top_level() { 1; 2; } ", Options: optionsWithTopLevel(1, true)},
			{Code: "function foo() { let one; let two = class { static { let three; let four; let five; if (six) { let seven; let eight; let nine; } } }; }", Options: option(2)},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    "function foo() { var bar = 1; var baz = 2; var qux = 3; }",
				Options: option(2),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'foo'", 3, 2), Line: 1, Column: 1, EndLine: 1, EndColumn: 13}},
			},
			{
				Code:    "var foo = () => { var bar = 1; var baz = 2; var qux = 3; };",
				Options: option(2),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Arrow function", 3, 2), Line: 1, Column: 14, EndLine: 1, EndColumn: 16}},
			},
			{
				Code:    "var foo = function() { var bar = 1; var baz = 2; var qux = 3; };",
				Options: option(2),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function", 3, 2), Line: 1, Column: 11, EndLine: 1, EndColumn: 19}},
			},
			{
				Code:    "function foo() { var bar = 1; if (true) { while (false) { var qux = null; } } return 3; }",
				Options: option(4),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'foo'", 5, 4), Line: 1, Column: 1, EndLine: 1, EndColumn: 13}},
			},
			{
				Code:    "function foo() { var bar = 1; if (true) { for (;;) { var qux = null; } } return 3; }",
				Options: option(4),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'foo'", 5, 4), Line: 1, Column: 1, EndLine: 1, EndColumn: 13}},
			},
			{
				Code:    "function foo() { var bar = 1; if (true) { for (;;) { var qux = null; } } else { quxx(); } return 3; }",
				Options: option(5),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'foo'", 6, 5), Line: 1, Column: 1, EndLine: 1, EndColumn: 13}},
			},
			{
				Code:    "function foo() { var x = 5; function bar() { var y = 6; } bar(); z = 10; baz(); }",
				Options: option(3),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'foo'", 5, 3), Line: 1, Column: 1, EndLine: 1, EndColumn: 13}},
			},
			{
				Code:    "function foo() { var x = 5; function bar() { var y = 6; } bar(); z = 10; baz(); }",
				Options: option(4),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'foo'", 5, 4), Line: 1, Column: 1, EndLine: 1, EndColumn: 13}},
			},
			{
				Code:    ";(function() { var bar = 1; return function () { var z; return 42; }; })()",
				Options: optionsWithTopLevel(1, true),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function", 2, 1), Line: 1, Column: 36, EndLine: 1, EndColumn: 45}},
			},
			{
				Code:    ";(function() { var bar = 1; var baz = 2; })(); (function() { var bar = 1; var baz = 2; })()",
				Options: optionsWithTopLevel(1, true),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Message: exceedMessage("Function", 2, 1), Line: 1, Column: 3, EndLine: 1, EndColumn: 11},
					{MessageId: "exceed", Message: exceedMessage("Function", 2, 1), Line: 1, Column: 49, EndLine: 1, EndColumn: 57},
				},
			},
			{
				Code:    "define(['foo', 'qux'], function(foo, qux) { var bar = 1; var baz = 2; return function () { var z; return 42; }; })",
				Options: optionsWithTopLevel(1, true),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function", 2, 1), Line: 1, Column: 78, EndLine: 1, EndColumn: 87}},
			},
			{
				Code:   "function foo() { var a; var b; var c; var x; var y; var z; bar(); baz(); qux(); quxx(); foo(); }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'foo'", 11, 10), Line: 1, Column: 1, EndLine: 1, EndColumn: 13}},
			},

			// ---- object property options ----
			{
				Code:    "var foo = { thing: function() { var bar = 1; var baz = 2; var baz2; } }",
				Options: option(2),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Method 'thing'", 3, 2), Line: 1, Column: 13, EndLine: 1, EndColumn: 28}},
			},
			{
				Code:    "var foo = { thing() { var bar = 1; var baz = 2; var baz2; } }",
				Options: option(2),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Method 'thing'", 3, 2), Line: 1, Column: 13, EndLine: 1, EndColumn: 18}},
			},
			{
				Code:    "var foo = { thing: () => { var bar = 1; var baz = 2; var baz2; } }",
				Options: option(2),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Method 'thing'", 3, 2), Line: 1, Column: 13, EndLine: 1, EndColumn: 20}},
			},
			{
				Code:    "var foo = { thing: function() { var bar = 1; var baz = 2; var baz2; } }",
				Options: option(map[string]interface{}{"max": 2}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Method 'thing'", 3, 2), Line: 1, Column: 13, EndLine: 1, EndColumn: 28}},
			},
			{
				Code:    "function foo() { 1; 2; 3; 4; 5; 6; 7; 8; 9; 10; 11; }",
				Options: option(map[string]interface{}{}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'foo'", 11, 10), Line: 1, Column: 1, EndLine: 1, EndColumn: 13}},
			},
			{
				Code:    "function foo() { 1; }",
				Options: option(map[string]interface{}{"max": 0}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'foo'", 1, 0), Line: 1, Column: 1, EndLine: 1, EndColumn: 13}},
			},
			{
				Code:    "function foo() { foo_1; /* foo_ 2 */ class C { static { one; two; three; four; { five; six; seven; eight; } } } foo_3 }",
				Options: option(2),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'foo'", 3, 2), Line: 1, Column: 1, EndLine: 1, EndColumn: 13}},
			},
			{
				Code:    "class C { static { one; two; three; four; function not_top_level() { 1; 2; 3; } five; six; seven; eight; } }",
				Options: optionsWithTopLevel(2, true),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'not_top_level'", 3, 2), Line: 1, Column: 43, EndLine: 1, EndColumn: 65}},
			},
			{
				Code:    "class C { static { { one; two; three; four; function not_top_level() { 1; 2; 3; } five; six; seven; eight; } } }",
				Options: optionsWithTopLevel(2, true),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'not_top_level'", 3, 2), Line: 1, Column: 45, EndLine: 1, EndColumn: 67}},
			},
			{
				Code:    "class C { static { { one; two; three; four; } function not_top_level() { 1; 2; 3; } { five; six; seven; eight; } } }",
				Options: optionsWithTopLevel(2, true),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'not_top_level'", 3, 2), Line: 1, Column: 47, EndLine: 1, EndColumn: 69}},
			},
		},
	)
}

func option(v any) []interface{} {
	return []interface{}{v}
}

func optionsWithTopLevel(maxStatements int, ignoreTopLevelFunctions bool) []interface{} {
	return []interface{}{maxStatements, map[string]interface{}{"ignoreTopLevelFunctions": ignoreTopLevelFunctions}}
}

func exceedMessage(name string, count, maxAllowed int) string {
	return fmt.Sprintf("%s has too many statements (%d). Maximum allowed is %d.", name, count, maxAllowed)
}
