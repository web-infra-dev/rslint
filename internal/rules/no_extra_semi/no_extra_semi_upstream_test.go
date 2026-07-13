// TestNoExtraSemiUpstream migrates the full valid/invalid suite from upstream
// eslint/tests/lib/rules/no-extra-semi.js 1:1. Position assertions cover
// line/column for every invalid case.
// rslint-specific lock-in cases live in the no_extra_semi_extras_test.go file.
package no_extra_semi

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoExtraSemiUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoExtraSemiRule,
		[]rule_tester.ValidTestCase{
			{Code: "var x = 5;"},
			{Code: "function foo(){}"},
			{Code: "for(;;);"},
			{Code: "while(0);"},
			{Code: "do;while(0);"},
			{Code: "for(a in b);"},
			{Code: "for(a of b);"},
			{Code: "if(true);"},
			{Code: "if(true); else;"},
			{Code: "foo: ;"},
			{Code: "with(foo);"},

			// Class body
			{Code: "class A { }"},
			{Code: "var A = class { };"},
			{Code: "class A { a() { this; } }"},
			{Code: "var A = class { a() { this; } };"},
			{Code: "class A { } a;"},
			{Code: "class A { field; }"},
			{Code: "class A { field = 0; }"},
			{Code: "class A { static { foo; } }"},

			// modules
			{Code: "export const x = 42;"},
			{Code: "export default 42;"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   "var x = 5;;",
				Output: []string{"var x = 5;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11},
				},
			},
			{
				Code:   "function foo(){};",
				Output: []string{"function foo(){}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 17},
				},
			},
			{
				Code:   "for(;;);;",
				Output: []string{"for(;;);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 9},
				},
			},
			{
				Code:   "while(0);;",
				Output: []string{"while(0);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 10},
				},
			},
			{
				Code:   "do;while(0);;",
				Output: []string{"do;while(0);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 13},
				},
			},
			{
				Code:   "for(a in b);;",
				Output: []string{"for(a in b);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 13},
				},
			},
			{
				Code:   "for(a of b);;",
				Output: []string{"for(a of b);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 13},
				},
			},
			{
				Code:   "if(true);;",
				Output: []string{"if(true);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 10},
				},
			},
			{
				Code:   "if(true){} else;;",
				Output: []string{"if(true){} else;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 17},
				},
			},
			{
				Code:   "if(true){;} else {;}",
				Output: []string{"if(true){} else {}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 10},
					{MessageId: "unexpected", Line: 1, Column: 19},
				},
			},
			{
				Code:   "foo:;;",
				Output: []string{"foo:;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 6},
				},
			},
			{
				Code:   "with(foo);;",
				Output: []string{"with(foo);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11},
				},
			},
			{
				Code:   "with(foo){;}",
				Output: []string{"with(foo){}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11},
				},
			},
			{
				Code:   "class A { static { ; } }",
				Output: []string{"class A { static {  } }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 20},
				},
			},
			{
				Code:   "class A { static { a;; } }",
				Output: []string{"class A { static { a; } }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 22},
				},
			},

			// Class body
			{
				Code:   "class A { ; }",
				Output: []string{"class A {  }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11},
				},
			},
			{
				Code:   "class A { /*a*/; }",
				Output: []string{"class A { /*a*/ }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 16},
				},
			},
			{
				Code:   "class A { ; a() {} }",
				Output: []string{"class A {  a() {} }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11},
				},
			},
			{
				Code:   "class A { a() {}; }",
				Output: []string{"class A { a() {} }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 17},
				},
			},
			{
				Code:   "class A { a() {}; b() {} }",
				Output: []string{"class A { a() {} b() {} }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 17},
				},
			},
			{
				Code:   "class A {; a() {}; b() {}; }",
				Output: []string{"class A { a() {} b() {} }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 10},
					{MessageId: "unexpected", Line: 1, Column: 18},
					{MessageId: "unexpected", Line: 1, Column: 26},
				},
			},
			{
				Code:   "class A { a() {}; get b() {} }",
				Output: []string{"class A { a() {} get b() {} }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 17},
				},
			},
			{
				Code:   "class A { field;; }",
				Output: []string{"class A { field; }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 17},
				},
			},
			{
				Code:   "class A { static {}; }",
				Output: []string{"class A { static {} }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 20},
				},
			},
			{
				Code:   "class A { static { a; }; foo(){} }",
				Output: []string{"class A { static { a; } foo(){} }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 24},
				},
			},

			// Directive hazards
			{
				Code: "; 'use strict'",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:   "; ; 'use strict'",
				Output: []string{" ; 'use strict'"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
					{MessageId: "unexpected", Line: 1, Column: 3},
				},
			},
			{
				Code: "debugger;\n;\n'use strict'",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 2, Column: 1},
				},
			},
			{
				Code: "function foo() { ; 'bar'; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 18},
				},
			},
			{
				Code:   "{ ; 'foo'; }",
				Output: []string{"{  'foo'; }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 3},
				},
			},
			{
				Code:   "; ('use strict');",
				Output: []string{" ('use strict');"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:   "; 1;",
				Output: []string{" 1;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
		},
	)
}
