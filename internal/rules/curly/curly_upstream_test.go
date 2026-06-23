// TestCurlyUpstream migrates the full valid/invalid suite from upstream
// eslint/tests/lib/rules/curly.js 1:1. Position assertions cover
// line/column/endLine/endColumn for every invalid case (values upstream pins
// down are mirrored verbatim; the rest are the rule's actual report ranges).
// rslint-specific lock-in cases live in the curly_extras_test.go file.
package curly

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestCurlyUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&CurlyRule,
		[]rule_tester.ValidTestCase{
			{Code: "if (foo) { bar() }"},
			{Code: "if (foo) { bar() } else if (foo2) { baz() }"},
			{Code: "while (foo) { bar() }"},
			{Code: "do { bar(); } while (foo)"},
			{Code: "for (;foo;) { bar() }"},
			{Code: "for (var foo in bar) { console.log(foo) }"},
			{Code: "for (var foo of bar) { console.log(foo) }"},
			{Code: "for (;foo;) bar()", Options: "multi"},
			{Code: "if (foo) bar()", Options: "multi"},
			{Code: "if (a) { b; c; }", Options: "multi"},
			{Code: "for (var foo in bar) console.log(foo)", Options: "multi"},
			{Code: "for (var foo in bar) { console.log(1); console.log(2) }", Options: "multi"},
			{Code: "for (var foo of bar) console.log(foo)", Options: "multi"},
			{Code: "for (var foo of bar) { console.log(1); console.log(2) }", Options: "multi"},
			{Code: "if (foo) bar()", Options: "multi-line"},
			{Code: "if (foo) bar() \n", Options: "multi-line"},
			{Code: "if (foo) bar(); else baz()", Options: "multi-line"},
			{Code: "if (foo) bar(); \n else baz()", Options: "multi-line"},
			{Code: "if (foo) bar() \n else if (foo) bar() \n else baz()", Options: "multi-line"},
			{Code: "do baz(); while (foo)", Options: "multi-line"},
			{Code: "if (foo) { bar() }", Options: "multi-line"},
			{Code: "for (var foo in bar) console.log(foo)", Options: "multi-line"},
			{Code: "for (var foo in bar) { \n console.log(1); \n console.log(2); \n }", Options: "multi-line"},
			{Code: "for (var foo of bar) console.log(foo)", Options: "multi-line"},
			{Code: "for (var foo of bar) { \n console.log(1); \n console.log(2); \n }", Options: "multi-line"},
			{Code: "if (foo) { \n bar(); \n baz(); \n }", Options: "multi-line"},
			{Code: "do bar() \n while (foo)", Options: "multi-line"},
			{Code: "if (foo) { \n quz = { \n bar: baz, \n qux: foo \n }; \n }", Options: "multi-or-nest"},
			{Code: "while (true) { \n if (foo) \n doSomething(); \n else \n doSomethingElse(); \n }", Options: "multi-or-nest"},
			{Code: "if (foo) \n quz = true;", Options: "multi-or-nest"},
			{Code: "if (foo) { \n // line of comment \n quz = true; \n }", Options: "multi-or-nest"},
			{Code: "// line of comment \n if (foo) \n quz = true; \n", Options: "multi-or-nest"},
			{Code: "while (true) \n doSomething();", Options: "multi-or-nest"},
			{Code: "for (var i = 0; foo; i++) \n doSomething();", Options: "multi-or-nest"},
			{Code: "if (foo) { \n if(bar) \n doSomething(); \n } else \n doSomethingElse();", Options: "multi-or-nest"},
			{Code: "for (var foo in bar) \n console.log(foo)", Options: "multi-or-nest"},
			{Code: "for (var foo in bar) { \n if (foo) console.log(1); \n else console.log(2) \n }", Options: "multi-or-nest"},
			{Code: "for (var foo of bar) \n console.log(foo)", Options: "multi-or-nest"},
			{Code: "for (var foo of bar) { \n if (foo) console.log(1); \n else console.log(2) \n }", Options: "multi-or-nest"},
			{Code: "if (foo) { const bar = 'baz'; }", Options: "multi"},
			{Code: "if (foo) { using bar = getThing(); }", Options: "multi"},
			{Code: "if (foo) { await using bar = getThing(); }", Options: "multi"},
			{Code: "while (foo) { let bar = 'baz'; }", Options: "multi"},
			{Code: "for(;;) { function foo() {} }", Options: "multi"},
			{Code: "for (foo in bar) { class Baz {} }", Options: "multi"},
			{Code: "if (foo) { let bar; } else { baz(); }", Options: []interface{}{"multi", "consistent"}},
			{Code: "if (foo) { bar(); } else { const baz = 'quux'; }", Options: []interface{}{"multi", "consistent"}},
			{Code: "if (foo) { \n const bar = 'baz'; \n }", Options: "multi-or-nest"},
			{Code: "if (foo) { \n let bar = 'baz'; \n }", Options: "multi-or-nest"},
			{Code: "if (foo) { \n function bar() {} \n }", Options: "multi-or-nest"},
			{Code: "if (foo) { \n class bar {} \n }", Options: "multi-or-nest"},

			// https://github.com/eslint/eslint/issues/12370
			{Code: "if (foo) doSomething() \n ;", Options: "multi-or-nest"},
			{Code: "if (foo) doSomething(); \n else if (bar) doSomethingElse() \n ;", Options: "multi-or-nest"},
			{Code: "if (foo) doSomething(); \n else doSomethingElse() \n ;", Options: "multi-or-nest"},
			{Code: "if (foo) doSomething(); \n else if (bar) doSomethingElse(); \n else doAnotherThing() \n ;", Options: "multi-or-nest"},
			{Code: "for (var i = 0; foo; i++) doSomething() \n ;", Options: "multi-or-nest"},
			{Code: "for (var foo in bar) console.log(foo) \n ;", Options: "multi-or-nest"},
			{Code: "for (var foo of bar) console.log(foo) \n ;", Options: "multi-or-nest"},
			{Code: "while (foo) doSomething() \n ;", Options: "multi-or-nest"},
			{Code: "do doSomething() \n ;while (foo)", Options: "multi-or-nest"},
			{Code: "if (foo)\n;", Options: "multi-or-nest"},
			{Code: "if (foo) doSomething(); \n else if (bar)\n;", Options: "multi-or-nest"},
			{Code: "if (foo) doSomething(); \n else\n;", Options: "multi-or-nest"},
			{Code: "if (foo) doSomething(); \n else if (bar) doSomethingElse(); \n else\n;", Options: "multi-or-nest"},
			{Code: "for (var i = 0; foo; i++)\n;", Options: "multi-or-nest"},
			{Code: "for (var foo in bar)\n;", Options: "multi-or-nest"},
			{Code: "for (var foo of bar)\n;", Options: "multi-or-nest"},
			{Code: "while (foo)\n;", Options: "multi-or-nest"},
			{Code: "do\n;while (foo)", Options: "multi-or-nest"},

			// https://github.com/eslint/eslint/issues/3856
			{Code: "if (true) { if (false) console.log(1) } else console.log(2)", Options: "multi"},
			{Code: "if (a) { if (b) console.log(1); else if (c) console.log(2) } else console.log(3)", Options: "multi"},
			{Code: "if (true) { while(false) if (true); } else;", Options: "multi"},
			{Code: "if (true) { label: if (false); } else;", Options: "multi"},
			{Code: "if (true) { with(0) if (false); } else;", Options: "multi"},
			{Code: "if (true) { while(a) if(b) while(c) if (d); else; } else;", Options: "multi"},
			{Code: "if (true) foo(); else { bar(); baz(); }", Options: "multi"},
			{Code: "if (true) { foo(); } else { bar(); baz(); }", Options: []interface{}{"multi", "consistent"}},
			{Code: "if (true) { foo(); } else if (true) { faa(); } else { bar(); baz(); }", Options: []interface{}{"multi", "consistent"}},
			{Code: "if (true) { foo(); faa(); } else { bar(); }", Options: []interface{}{"multi", "consistent"}},

			// https://github.com/feross/standard/issues/664
			{Code: "if (true) foo()\n;[1, 2, 3].bar()", Options: "multi-line"},

			// https://github.com/eslint/eslint/issues/12928 (also in invalid[])
			{Code: "if (x) for (var i in x) { if (i > 0) console.log(i); } else console.log('whoops');", Options: "multi"},
			{Code: "if (a) { if (b) foo(); } else bar();", Options: "multi"},
			{Code: "if (a) { if (b) foo(); } else bar();", Options: "multi-or-nest"},
			{Code: "if (a) { if (b) foo(); } else { bar(); }", Options: []interface{}{"multi", "consistent"}},
			{Code: "if (a) { if (b) foo(); } else { bar(); }", Options: []interface{}{"multi-or-nest", "consistent"}},
			{Code: "if (a) { if (b) { foo(); bar(); } } else baz();", Options: "multi"},
			{Code: "if (a) foo(); else if (b) { if (c) bar(); } else baz();", Options: "multi"},
			{Code: "if (a) { if (b) foo(); else if (c) bar(); } else baz();", Options: "multi"},
			{Code: "if (a) if (b) foo(); else { if (c) bar(); } else baz();", Options: "multi"},
			{Code: "if (a) { lbl:if (b) foo(); } else bar();", Options: "multi"},
			{Code: "if (a) { lbl1:lbl2:if (b) foo(); } else bar();", Options: "multi"},
			{Code: "if (a) { for (;;) if (b) foo(); } else bar();", Options: "multi"},
			{Code: "if (a) { for (key in obj) if (b) foo(); } else bar();", Options: "multi"},
			{Code: "if (a) { for (elem of arr) if (b) foo(); } else bar();", Options: "multi"},
			{Code: "if (a) { with (obj) if (b) foo(); } else bar();", Options: "multi"},
			{Code: "if (a) { while (cond) if (b) foo(); } else bar();", Options: "multi"},
			{Code: "if (a) { while (cond) for (;;) for (key in obj) if (b) foo(); } else bar();", Options: "multi"},
			{Code: "if (a) while (cond) { for (;;) for (key in obj) if (b) foo(); } else bar();", Options: "multi"},
			{Code: "if (a) while (cond) for (;;) { for (key in obj) if (b) foo(); } else bar();", Options: "multi"},
			{Code: "if (a) while (cond) for (;;) for (key in obj) { if (b) foo(); } else bar();", Options: "multi"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   "if (foo) bar()",
				Output: []string{"if (foo) {bar()}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 10, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:   "if (foo) \n bar()",
				Output: []string{"if (foo) \n {bar()}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 2, Column: 2, EndLine: 2, EndColumn: 7},
				},
			},
			{
				Code:   "if (foo) { bar() } else baz()",
				Output: []string{"if (foo) { bar() } else {baz()}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter", Line: 1, Column: 25, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code:   "if (foo) { bar() } else if (faa) baz()",
				Output: []string{"if (foo) { bar() } else if (faa) {baz()}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 34, EndLine: 1, EndColumn: 39},
				},
			},
			{
				Code:   "while (foo) bar()",
				Output: []string{"while (foo) {bar()}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 13, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:   "while (foo) \n bar()",
				Output: []string{"while (foo) \n {bar()}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 2, Column: 2, EndLine: 2, EndColumn: 7},
				},
			},
			{
				Code:   "do bar(); while (foo)",
				Output: []string{"do {bar();} while (foo)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter", Line: 1, Column: 4, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code:   "do \n bar(); while (foo)",
				Output: []string{"do \n {bar();} while (foo)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter", Line: 2, Column: 2, EndLine: 2, EndColumn: 8},
				},
			},
			{
				Code:   "for (;foo;) bar()",
				Output: []string{"for (;foo;) {bar()}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 13, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:   "for (var foo in bar) console.log(foo)",
				Output: []string{"for (var foo in bar) {console.log(foo)}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter", Line: 1, Column: 22, EndLine: 1, EndColumn: 38},
				},
			},
			{
				Code:   "for (var foo of bar) console.log(foo)",
				Output: []string{"for (var foo of bar) {console.log(foo)}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter", Line: 1, Column: 22, EndLine: 1, EndColumn: 38},
				},
			},
			{
				Code:   "for (var foo of bar) \n console.log(foo)",
				Output: []string{"for (var foo of bar) \n {console.log(foo)}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter", Line: 2, Column: 2, EndLine: 2, EndColumn: 18},
				},
			},
			{
				Code:   "for (a;;) console.log(foo)",
				Output: []string{"for (a;;) {console.log(foo)}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 11, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code:   "for (a;;) \n console.log(foo)",
				Output: []string{"for (a;;) \n {console.log(foo)}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 2, Column: 2, EndLine: 2, EndColumn: 18},
				},
			},
			{
				Code:    "for (var foo of bar) {console.log(foo)}",
				Output:  []string{"for (var foo of bar) console.log(foo)"},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Line: 1, Column: 22, EndLine: 1, EndColumn: 40},
				},
			},
			{
				Code:    "do{foo();} while(bar);",
				Output:  []string{"do foo(); while(bar);"},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Line: 1, Column: 3, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code:    "for (;foo;) { bar() }",
				Output:  []string{"for (;foo;)  bar() "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 13, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code:   "for (;foo;) \n bar()",
				Output: []string{"for (;foo;) \n {bar()}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 2, Column: 2, EndLine: 2, EndColumn: 7},
				},
			},
			{
				Code:    "if (foo) { bar() }",
				Output:  []string{"if (foo)  bar() "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 10, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "if (foo) if (bar) { baz() }",
				Output:  []string{"if (foo) if (bar)  baz() "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 19, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code:    "if (foo) if (bar) baz(); else if (quux) { quuux(); }",
				Output:  []string{"if (foo) if (bar) baz(); else if (quux)  quuux(); "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 41, EndLine: 1, EndColumn: 53},
				},
			},
			{
				Code:    "while (foo) { bar() }",
				Output:  []string{"while (foo)  bar() "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 13, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code:    "if (foo) baz(); else { bar() }",
				Output:  []string{"if (foo) baz(); else  bar() "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Line: 1, Column: 22, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "if (foo) if (bar); else { baz() }",
				Output:  []string{"if (foo) if (bar); else  baz() "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Line: 1, Column: 25, EndLine: 1, EndColumn: 34},
				},
			},
			{
				Code:    "if (true) { if (false) console.log(1) }",
				Output:  []string{"if (true)  if (false) console.log(1) "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 11, EndLine: 1, EndColumn: 40},
				},
			},
			{
				Code:    "if (a) { if (b) console.log(1); else console.log(2) } else console.log(3)",
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 54},
				},
			},
			{
				Code: "if (0)\n    console.log(0)\nelse if (1) {\n    console.log(1)\n    console.log(1)\n} else {\n    if (2)\n        console.log(2)\n    else\n        console.log(3)\n}",
				Output: []string{
					"if (0)\n    console.log(0)\nelse if (1) {\n    console.log(1)\n    console.log(1)\n} else \n    if (2)\n        console.log(2)\n    else\n        console.log(3)\n",
				},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Line: 6, Column: 8, EndLine: 11, EndColumn: 2},
				},
			},
			{
				Code:    "for (var foo in bar) { console.log(foo) }",
				Output:  []string{"for (var foo in bar)  console.log(foo) "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Line: 1, Column: 22, EndLine: 1, EndColumn: 42},
				},
			},
			{
				Code:    "for (var foo of bar) { console.log(foo) }",
				Output:  []string{"for (var foo of bar)  console.log(foo) "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Line: 1, Column: 22, EndLine: 1, EndColumn: 42},
				},
			},
			{
				Code:    "if (foo) \n baz()",
				Output:  []string{"if (foo) \n {baz()}"},
				Options: "multi-line",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 2, Column: 2, EndLine: 2, EndColumn: 7},
				},
			},
			{
				Code:   "if (foo) baz()",
				Output: []string{"if (foo) {baz()}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 10, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    "while (foo) \n baz()",
				Output:  []string{"while (foo) \n {baz()}"},
				Options: "multi-line",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 2, Column: 2, EndLine: 2, EndColumn: 7},
				},
			},
			{
				Code:    "for (;foo;) \n bar()",
				Output:  []string{"for (;foo;) \n {bar()}"},
				Options: "multi-line",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 2, Column: 2, EndLine: 2, EndColumn: 7},
				},
			},
			{
				Code:    "while (bar && \n baz) \n foo()",
				Output:  []string{"while (bar && \n baz) \n {foo()}"},
				Options: "multi-line",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 3, Column: 2, EndLine: 3, EndColumn: 7},
				},
			},
			{
				Code:    "if (foo) bar(baz, \n baz)",
				Output:  []string{"if (foo) {bar(baz, \n baz)}"},
				Options: "multi-line",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 10, EndLine: 2, EndColumn: 6},
				},
			},
			{
				Code:    "do foo(); while (bar)",
				Output:  []string{"do {foo();} while (bar)"},
				Options: "all",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter", Line: 1, Column: 4, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code:    "do \n foo(); \n while (bar)",
				Output:  []string{"do \n {foo();} \n while (bar)"},
				Options: "multi-line",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter", Line: 2, Column: 2, EndLine: 2, EndColumn: 8},
				},
			},
			{
				Code:    "for (var foo in bar) {console.log(foo)}",
				Output:  []string{"for (var foo in bar) console.log(foo)"},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Line: 1, Column: 22, EndLine: 1, EndColumn: 40},
				},
			},
			{
				Code:    "for (var foo in bar) \n console.log(foo)",
				Output:  []string{"for (var foo in bar) \n {console.log(foo)}"},
				Options: "multi-line",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter", Line: 2, Column: 2, EndLine: 2, EndColumn: 18},
				},
			},
			{
				Code:    "for (var foo in bar) \n console.log(1); \n console.log(2)",
				Output:  []string{"for (var foo in bar) \n {console.log(1);} \n console.log(2)"},
				Options: "multi-line",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter", Line: 2, Column: 2, EndLine: 2, EndColumn: 17},
				},
			},
			{
				Code:    "for (var foo of bar) \n console.log(foo)",
				Output:  []string{"for (var foo of bar) \n {console.log(foo)}"},
				Options: "multi-line",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter", Line: 2, Column: 2, EndLine: 2, EndColumn: 18},
				},
			},
			{
				Code:    "for (var foo of bar) \n console.log(1); \n console.log(2)",
				Output:  []string{"for (var foo of bar) \n {console.log(1);} \n console.log(2)"},
				Options: "multi-line",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter", Line: 2, Column: 2, EndLine: 2, EndColumn: 17},
				},
			},
			{
				Code:    "if (foo) \n quz = { \n bar: baz, \n qux: foo \n };",
				Output:  []string{"if (foo) \n {quz = { \n bar: baz, \n qux: foo \n };}"},
				Options: "multi-or-nest",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 2, Column: 2, EndLine: 5, EndColumn: 4},
				},
			},
			{
				Code:    "while (true) \n if (foo) \n doSomething(); \n else \n doSomethingElse(); \n",
				Output:  []string{"while (true) \n {if (foo) \n doSomething(); \n else \n doSomethingElse();} \n"},
				Options: "multi-or-nest",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 2, Column: 2, EndLine: 5, EndColumn: 20},
				},
			},
			{
				Code:    "if (foo) { \n quz = true; \n }",
				Output:  []string{"if (foo)  \n quz = true; \n "},
				Options: "multi-or-nest",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 10, EndLine: 3, EndColumn: 3},
				},
			},
			{
				Code:    "if (foo) { var bar = 'baz'; }",
				Output:  []string{"if (foo)  var bar = 'baz'; "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code:    "if (foo) { let bar; } else baz();",
				Output:  []string{"if (foo) { let bar; } else {baz();}"},
				Options: []interface{}{"multi", "consistent"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter", Line: 1, Column: 28, EndLine: 1, EndColumn: 34},
				},
			},
			{
				Code:    "if (foo) bar(); else { const baz = 'quux' }",
				Output:  []string{"if (foo) {bar();} else { const baz = 'quux' }"},
				Options: []interface{}{"multi", "consistent"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 10, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    "if (foo) { \n var bar = 'baz'; \n }",
				Output:  []string{"if (foo)  \n var bar = 'baz'; \n "},
				Options: "multi-or-nest",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 10, EndLine: 3, EndColumn: 3},
				},
			},
			{
				Code:    "while (true) { \n doSomething(); \n }",
				Output:  []string{"while (true)  \n doSomething(); \n "},
				Options: "multi-or-nest",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 14, EndLine: 3, EndColumn: 3},
				},
			},
			{
				Code:    "for (var i = 0; foo; i++) { \n doSomething(); \n }",
				Output:  []string{"for (var i = 0; foo; i++)  \n doSomething(); \n "},
				Options: "multi-or-nest",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 27, EndLine: 3, EndColumn: 3},
				},
			},
			{
				Code:    "for (var foo in bar) if (foo) console.log(1); else console.log(2);",
				Output:  []string{"for (var foo in bar) {if (foo) console.log(1); else console.log(2);}", "for (var foo in bar) {if (foo) {console.log(1);} else {console.log(2);}}"},
				Options: "all",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter", Line: 1, Column: 22, EndLine: 1, EndColumn: 67},
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 31, EndLine: 1, EndColumn: 46},
					{MessageId: "missingCurlyAfter", Line: 1, Column: 52, EndLine: 1, EndColumn: 67},
				},
			},
			{
				Code:    "for (var foo in bar) \n if (foo) console.log(1); \n else console.log(2);",
				Output:  []string{"for (var foo in bar) \n {if (foo) console.log(1); \n else console.log(2);}"},
				Options: "multi-or-nest",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter", Line: 2, Column: 2, EndLine: 3, EndColumn: 22},
				},
			},
			{
				Code:    "for (var foo in bar) { if (foo) console.log(1) }",
				Output:  []string{"for (var foo in bar)  if (foo) console.log(1) "},
				Options: "multi-or-nest",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Line: 1, Column: 22, EndLine: 1, EndColumn: 49},
				},
			},
			{
				Code:    "for (var foo of bar) \n if (foo) console.log(1); \n else console.log(2);",
				Output:  []string{"for (var foo of bar) \n {if (foo) console.log(1); \n else console.log(2);}"},
				Options: "multi-or-nest",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter", Line: 2, Column: 2, EndLine: 3, EndColumn: 22},
				},
			},
			{
				Code:    "for (var foo of bar) { if (foo) console.log(1) }",
				Output:  []string{"for (var foo of bar)  if (foo) console.log(1) "},
				Options: "multi-or-nest",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Line: 1, Column: 22, EndLine: 1, EndColumn: 49},
				},
			},
			{
				Code:    "if (true) foo(); \n else { \n bar(); \n baz(); \n }",
				Output:  []string{"if (true) {foo();} \n else { \n bar(); \n baz(); \n }"},
				Options: []interface{}{"multi", "consistent"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 11, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code:    "if (true) { foo(); faa(); }\n else bar();",
				Output:  []string{"if (true) { foo(); faa(); }\n else {bar();}"},
				Options: []interface{}{"multi", "consistent"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter", Line: 2, Column: 7, EndLine: 2, EndColumn: 13},
				},
			},
			{
				Code:    "if (true) foo(); else { baz(); }",
				Output:  []string{"if (true) foo(); else  baz(); "},
				Options: []interface{}{"multi", "consistent"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Line: 1, Column: 23, EndLine: 1, EndColumn: 33},
				},
			},
			{
				Code:    "if (true) foo(); else if (true) faa(); else { bar(); baz(); }",
				Output:  []string{"if (true) {foo();} else if (true) {faa();} else { bar(); baz(); }"},
				Options: []interface{}{"multi", "consistent"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 11, EndLine: 1, EndColumn: 17},
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 33, EndLine: 1, EndColumn: 39},
				},
			},
			{
				Code:    "if (true) if (true) foo(); else { bar(); baz(); }",
				Output:  []string{"if (true) if (true) {foo();} else { bar(); baz(); }"},
				Options: []interface{}{"multi", "consistent"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 21, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code:    "do{foo();} while (bar)",
				Output:  []string{"do foo(); while (bar)"},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Line: 1, Column: 3, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code:    "do\n{foo();} while (bar)",
				Output:  []string{"do\nfoo(); while (bar)"},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Line: 2, Column: 1, EndLine: 2, EndColumn: 9},
				},
			},
			{
				Code:    "while (bar) { foo(); }",
				Output:  []string{"while (bar)  foo(); "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 13, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "while (bar) \n{\n foo(); }",
				Output:  []string{"while (bar) \n\n foo(); "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 2, Column: 1, EndLine: 3, EndColumn: 10},
				},
			},
			{
				Code:    "for (;;) { foo(); }",
				Output:  []string{"for (;;)  foo(); "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 10, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "do{[1, 2, 3].map(bar);} while (bar)",
				Output:  []string{"do[1, 2, 3].map(bar); while (bar)"},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Line: 1, Column: 3, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "if (foo) {bar()} baz()",
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 10, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code:    "do {foo();} while (bar)",
				Output:  []string{"do foo(); while (bar)"},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Line: 1, Column: 4, EndLine: 1, EndColumn: 12},
				},
			},

			// Don't remove curly braces if it would cause issues due to ASI.
			{
				Code:    "if (foo) { bar }\n++baz;",
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 10, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code:    "if (foo) { bar; }\n++baz;",
				Output:  []string{"if (foo)  bar; \n++baz;"},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 10, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    "if (foo) { bar++ }\nbaz;",
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 10, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "if (foo) { bar }\n[1, 2, 3].map(foo);",
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 10, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code:    "if (foo) { bar }\n(1).toString();",
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 10, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code:    "if (foo) { bar }\n/regex/.test('foo');",
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 10, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code:    "if (foo) { bar }\nBaz();",
				Output:  []string{"if (foo)  bar \nBaz();"},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 10, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code:    "if (a) {\n  while (b) {\n    c();\n    d();\n  }\n} else e();",
				Output:  []string{"if (a) \n  while (b) {\n    c();\n    d();\n  }\n else e();"},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 6, EndColumn: 2},
				},
			},
			{
				Code:    "if (foo) { while (bar) {} } else {}",
				Output:  []string{"if (foo)  while (bar) {}  else {}"},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 10, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code:    "if (foo) { var foo = () => {} } else {}",
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
				},
			},
			{
				Code:    "if (foo) { var foo = function() {} } else {}",
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 10, EndLine: 1, EndColumn: 37},
				},
			},
			{
				Code:    "if (foo) { var foo = function*() {} } else {}",
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 10, EndLine: 1, EndColumn: 38},
				},
			},
			{
				Code:    "if (true)\nfoo()\n;[1, 2, 3].bar()",
				Output:  []string{"if (true)\n{foo()\n;}[1, 2, 3].bar()"},
				Options: "multi-line",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 2, Column: 1, EndLine: 3, EndColumn: 2},
				},
			},

			// https://github.com/eslint/eslint/issues/12370
			{
				Code:    "if (foo) {\ndoSomething()\n;\n}",
				Output:  []string{"if (foo) \ndoSomething()\n;\n"},
				Options: "multi-or-nest",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 10, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code:    "if (foo) doSomething();\nelse if (bar) {\ndoSomethingElse()\n;\n}",
				Output:  []string{"if (foo) doSomething();\nelse if (bar) \ndoSomethingElse()\n;\n"},
				Options: "multi-or-nest",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 2, Column: 15, EndLine: 5, EndColumn: 2},
				},
			},
			{
				Code:    "if (foo) doSomething();\nelse {\ndoSomethingElse()\n;\n}",
				Output:  []string{"if (foo) doSomething();\nelse \ndoSomethingElse()\n;\n"},
				Options: "multi-or-nest",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Line: 2, Column: 6, EndLine: 5, EndColumn: 2},
				},
			},
			{
				Code:    "for (var i = 0; foo; i++) {\ndoSomething()\n;\n}",
				Output:  []string{"for (var i = 0; foo; i++) \ndoSomething()\n;\n"},
				Options: "multi-or-nest",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 27, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code:    "for (var foo in bar) {\ndoSomething()\n;\n}",
				Output:  []string{"for (var foo in bar) \ndoSomething()\n;\n"},
				Options: "multi-or-nest",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Line: 1, Column: 22, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code:    "for (var foo of bar) {\ndoSomething()\n;\n}",
				Output:  []string{"for (var foo of bar) \ndoSomething()\n;\n"},
				Options: "multi-or-nest",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Line: 1, Column: 22, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code:    "while (foo) {\ndoSomething()\n;\n}",
				Output:  []string{"while (foo) \ndoSomething()\n;\n"},
				Options: "multi-or-nest",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 13, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code:    "do {\ndoSomething()\n;\n} while (foo)",
				Output:  []string{"do \ndoSomething()\n;\n while (foo)"},
				Options: "multi-or-nest",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Line: 1, Column: 4, EndLine: 4, EndColumn: 2},
				},
			},

			// https://github.com/eslint/eslint/issues/12928 (also in valid[])
			{
				Code:    "if (a) { if (b) foo(); }",
				Output:  []string{"if (a)  if (b) foo(); "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    "if (a) { if (b) foo(); else bar(); }",
				Output:  []string{"if (a)  if (b) foo(); else bar(); "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 37},
				},
			},
			{
				Code:    "if (a) { if (b) foo(); else bar(); } baz();",
				Output:  []string{"if (a)  if (b) foo(); else bar();  baz();"},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 37},
				},
			},
			{
				Code:    "if (a) { while (cond) if (b) foo(); }",
				Output:  []string{"if (a)  while (cond) if (b) foo(); "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 38},
				},
			},
			{
				Code:    "if (a) while (cond) { if (b) foo(); }",
				Output:  []string{"if (a) while (cond)  if (b) foo(); "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 21, EndLine: 1, EndColumn: 38},
				},
			},
			{
				Code:    "if (a) while (cond) { if (b) foo(); else bar(); }",
				Output:  []string{"if (a) while (cond)  if (b) foo(); else bar(); "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 21, EndLine: 1, EndColumn: 50},
				},
			},
			{
				Code:    "if (a) { while (cond) { if (b) foo(); } bar(); baz() } else quux();",
				Output:  []string{"if (a) { while (cond)  if (b) foo();  bar(); baz() } else quux();"},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 23, EndLine: 1, EndColumn: 40},
				},
			},
			{
				Code:    "if (a) { if (b) foo(); } bar();",
				Output:  []string{"if (a)  if (b) foo();  bar();"},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    "if(a) { if (b) foo(); } if (c) bar(); else baz();",
				Output:  []string{"if(a)  if (b) foo();  if (c) bar(); else baz();"},
				Options: "multi-or-nest",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 7, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "if (a) { do if (b) foo(); while (cond); } else bar();",
				Output:  []string{"if (a)  do if (b) foo(); while (cond);  else bar();"},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 42},
				},
			},
			{
				Code:    "if (a) do { if (b) foo(); } while (cond); else bar();",
				Output:  []string{"if (a) do  if (b) foo();  while (cond); else bar();"},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Line: 1, Column: 11, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code:    "if (a) { if (b) foo(); else bar(); } else baz();",
				Output:  []string{"if (a)  if (b) foo(); else bar();  else baz();"},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 37},
				},
			},
			{
				Code:    "if (a) while (cond) { bar(); } else baz();",
				Output:  []string{"if (a) while (cond)  bar();  else baz();"},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 21, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "if (a) { for (;;); } else bar();",
				Output:  []string{"if (a)  for (;;);  else bar();"},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    "if (a) { while (cond) if (b) foo() } else bar();",
				Output:  []string{"if (a) { while (cond) if (b) foo() } else {bar();}"},
				Options: []interface{}{"multi", "consistent"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter", Line: 1, Column: 43, EndLine: 1, EndColumn: 49},
				},
			},
			{
				Code:    "if (a)  while (cond) if (b) foo()  \nelse\n {bar();}",
				Output:  []string{"if (a)  while (cond) if (b) foo()  \nelse\n bar();"},
				Options: []interface{}{"multi", "consistent"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Line: 3, Column: 2, EndLine: 3, EndColumn: 10},
				},
			},
			{
				Code:   "if (a) foo() \nelse\n bar();",
				Output: []string{"if (a) {foo()} \nelse\n {bar();}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 13},
					{MessageId: "missingCurlyAfter", Line: 3, Column: 2, EndLine: 3, EndColumn: 8},
				},
			},
			{
				Code:    "if (a) { while (cond) if (b) foo() } ",
				Output:  []string{"if (a)  while (cond) if (b) foo()  "},
				Options: []interface{}{"multi", "consistent"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 37},
				},
			},
			{
				Code:    "if(a) { if (b) foo(); } if (c) bar(); else if(foo){bar();}",
				Output:  []string{"if(a)  if (b) foo();  if (c) bar(); else if(foo)bar();"},
				Options: "multi-or-nest",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 7, EndLine: 1, EndColumn: 24},
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 51, EndLine: 1, EndColumn: 59},
				},
			},
			{
				Code:    "if (true) [1, 2, 3]\n.bar()",
				Output:  []string{"if (true) {[1, 2, 3]\n.bar()}"},
				Options: "multi-line",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 11, EndLine: 2, EndColumn: 7},
				},
			},
			{
				Code:    "for(\n;\n;\n) {foo()}",
				Output:  []string{"for(\n;\n;\n) foo()"},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 4, Column: 3, EndLine: 4, EndColumn: 10},
				},
			},
			{
				Code:    "for(\n;\n;\n) \nfoo()\n",
				Output:  []string{"for(\n;\n;\n) \n{foo()}\n"},
				Options: "multi-line",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 5, Column: 1, EndLine: 5, EndColumn: 6},
				},
			},
			{
				// Reports 2 errors, but one pair of braces is necessary if the other pair gets removed.
				Code:    "if (a) { while (cond) { if (b) foo(); } } else bar();",
				Output:  []string{"if (a)  while (cond) { if (b) foo(); }  else bar();"},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 42},
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 23, EndLine: 1, EndColumn: 40},
				},
			},
			{
				Code:   "for(;;)foo()\n",
				Output: []string{"for(;;){foo()}\n"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:   "for(var \ni \n in \n z)foo()\n",
				Output: []string{"for(var \ni \n in \n z){foo()}\n"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter", Line: 4, Column: 4, EndLine: 4, EndColumn: 9},
				},
			},
			{
				Code:   "for(var i of \n z)\nfoo()\n",
				Output: []string{"for(var i of \n z)\n{foo()}\n"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter", Line: 3, Column: 1, EndLine: 3, EndColumn: 6},
				},
			},
		},
	)
}
