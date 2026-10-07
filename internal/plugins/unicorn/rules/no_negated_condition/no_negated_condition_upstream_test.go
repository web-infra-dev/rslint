// Ported from eslint-plugin-unicorn v77.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/test/no-negated-condition.js
package no_negated_condition_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_negated_condition"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNegatedConditionUpstreamComments(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_negated_condition.NoNegatedConditionRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{
			Code: "!x ? /* one */ 1 : /* two */ 2", FileName: "case.js",
			Output: []string{"x ? /* two */ 2 : /* one */ 1"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			},
		},
		{
			Code: "!x ? 1 /* one */ : 2", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			},
		},
		{
			Code: "!x ? 1 : 2 /* two */", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			},
		},
		{
			Code: "!x ? (1) /* one */ : (2)", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			},
		},
		{
			Code: "!x ? 1 : (2) /* two */", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			},
		},
		{
			Code: "!x ? 1 // one\n : 2", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			},
		},
		{
			Code: "x != y ? 1 /* one */ : 2", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
			},
		},
		{
			Code: "x !== y ? 1 : 2 /* two */", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
			},
		},
		{
			Code: "if (!x) {one();} /* one */ else {two();}", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 7},
			},
		},
		{
			Code: "if (!x) {one();} else {two();} /* two */", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 7},
			},
		},
		{
			Code: "if (!x) one(); /* one */ else two();", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 7},
			},
		},
		{
			Code: "if (!x) one(); else two(); // two", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 7},
			},
		},
	})
}

func TestNoNegatedConditionUpstreamSnapshots(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_negated_condition.NoNegatedConditionRule, []rule_tester.ValidTestCase{
		{Code: "if (a) {}", FileName: "case.js"},
		{Code: "if (a) {} else {}", FileName: "case.js"},
		{Code: "if (!a) {}", FileName: "case.js"},
		{Code: "if (!a) {} else if (b) {}", FileName: "case.js"},
		{Code: "if (!a) {} else if (b) {} else {}", FileName: "case.js"},
		{Code: "if (a == b) {}", FileName: "case.js"},
		{Code: "if (a == b) {} else {}", FileName: "case.js"},
		{Code: "if (a != b) {}", FileName: "case.js"},
		{Code: "if (a != b) {} else if (b) {}", FileName: "case.js"},
		{Code: "if (a != b) {} else if (b) {} else {}", FileName: "case.js"},
		{Code: "if (a !== b) {}", FileName: "case.js"},
		{Code: "if (a === b) {} else {}", FileName: "case.js"},
		{Code: "a ? b : c", FileName: "case.js"},
	}, []rule_tester.InvalidTestCase{
		{
			Code: "if (!a) {;} else {;}", FileName: "case.js",
			Output: []string{"if (a) {;} else {;}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 7},
			},
		},
		{
			Code: "if (a != b) {;} else {;}", FileName: "case.js",
			Output: []string{"if (a == b) {;} else {;}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 11},
			},
		},
		{
			Code: "if (a !== b) {;} else {;}", FileName: "case.js",
			Output: []string{"if (a === b) {;} else {;}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 12},
			},
		},
		{
			Code: "!a ? b : c", FileName: "case.js",
			Output: []string{"a ? c : b"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			},
		},
		{
			Code: "a != b ? c : d", FileName: "case.js",
			Output: []string{"a == b ? d : c"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
			},
		},
		{
			Code: "a !== b ? c : d", FileName: "case.js",
			Output: []string{"a === b ? d : c"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
			},
		},
		{
			Code: "(( !a )) ? b : c", FileName: "case.js",
			Output: []string{"(( a )) ? c : b"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 4, EndLine: 1, EndColumn: 6},
			},
		},
		{
			Code: "!(( a )) ? b : c", FileName: "case.js",
			Output: []string{"(( a )) ? c : b"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
			},
		},
		{
			Code: "if(!(( a ))) b(); else c();", FileName: "case.js",
			Output: []string{"if( a ) {c();} else {b();}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 4, EndLine: 1, EndColumn: 12},
			},
		},
		{
			Code: "if((( !a ))) b(); else c();", FileName: "case.js",
			Output: []string{"if((( a ))) {c();} else {b();}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 7, EndLine: 1, EndColumn: 9},
			},
		},
		{
			Code: "function a() {return!a ? b : c}", FileName: "case.js",
			Output: []string{"function a() {return a ? c : b}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 21, EndLine: 1, EndColumn: 23},
			},
		},
		{
			Code: "function a() {return!(( a )) ? b : c}", FileName: "case.js",
			Output: []string{"function a() {return (( a )) ? c : b}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 21, EndLine: 1, EndColumn: 29},
			},
		},
		{
			Code: "function a() {\n\treturn ! // comment\n\t\ta ? b : c;\n}", FileName: "case.js",
			Output: []string{"function a() {\n\treturn (  // comment\n\t\ta ? c : b);\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 2, Column: 9, EndLine: 3, EndColumn: 4},
			},
		},
		{
			Code: "function a() {\n\treturn (! // ReturnStatement argument is parenthesized\n\t\ta ? b : c);\n}", FileName: "case.js",
			Output: []string{"function a() {\n\treturn ( // ReturnStatement argument is parenthesized\n\t\ta ? c : b);\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 2, Column: 10, EndLine: 3, EndColumn: 4},
			},
		},
		{
			Code: "function a() {\n\treturn (\n\t\t! // UnaryExpression argument is parenthesized\n\t\ta) ? b : c;\n}", FileName: "case.js",
			Output: []string{"function a() {\n\treturn (\n\t\t // UnaryExpression argument is parenthesized\n\t\ta) ? c : b;\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 3, Column: 3, EndLine: 4, EndColumn: 4},
			},
		},
		{
			Code: "function a() {\n\tthrow ! // comment\n\t\ta ? b : c;\n}", FileName: "case.js",
			Output: []string{"function a() {\n\tthrow (  // comment\n\t\ta ? c : b);\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 2, Column: 8, EndLine: 3, EndColumn: 4},
			},
		},
		{
			Code: "!a ? b : c ? d : e", FileName: "case.js",
			Output: []string{"a ? c ? d : e : b"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			},
		},
		{
			Code: "!a ? b : (( c ? d : e ))", FileName: "case.js",
			Output: []string{"a ? (( c ? d : e )) : b"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			},
		},
		{
			Code: "a\n![] ? b : c", FileName: "case.js",
			Output: []string{"a\n;[] ? c : b"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 2, Column: 1, EndLine: 2, EndColumn: 4},
			},
		},
		{
			Code: "a\n!+b ? c : d", FileName: "case.js",
			Output: []string{"a\n;+b ? d : c"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 2, Column: 1, EndLine: 2, EndColumn: 4},
			},
		},
		{
			Code: "a\n!(b) ? c : d", FileName: "case.js",
			Output: []string{"a\n;(b) ? d : c"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 2, Column: 1, EndLine: 2, EndColumn: 5},
			},
		},
		{
			Code: "a\n!b ? c : d", FileName: "case.js",
			Output: []string{"a\nb ? d : c"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 2, Column: 1, EndLine: 2, EndColumn: 3},
			},
		},
		{
			Code: "if (!a)\n\tb()\nelse\n\tc()", FileName: "case.js",
			Output: []string{"if (a)\n\t{c()}\nelse\n\t{b()}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 7},
			},
		},
		{
			Code: "if(!a) b(); else c()", FileName: "case.js",
			Output: []string{"if(a) {c()} else {b();}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 4, EndLine: 1, EndColumn: 6},
			},
		},
		{
			Code: "function fn() {\n\tif(!a) b(); else return\n}", FileName: "case.js",
			Output: []string{"function fn() {\n\tif(a) {return} else {b();}\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 2, Column: 5, EndLine: 2, EndColumn: 7},
			},
		},
		{
			Code: "if(!a) {b()} else {c()}", FileName: "case.js",
			Output: []string{"if(a) {c()} else {b()}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 4, EndLine: 1, EndColumn: 6},
			},
		},
		{
			Code: "if(!!a) b(); else c();", FileName: "case.js",
			Output: []string{"if(!a) {c();} else {b();}", "if(a) {b();} else {c();}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 4, EndLine: 1, EndColumn: 7},
			},
		},
		{
			Code: "(!!a) ? b() : c();", FileName: "case.js",
			Output: []string{"(!a) ? c() : b();", "(a) ? b() : c();"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 2, EndLine: 1, EndColumn: 5},
			},
		},
		{
			Code: "function fn() {\n\treturn!a !== b ? c : d\n\treturn((!((a)) != b)) ? c : d\n}", FileName: "case.js",
			Output: []string{"function fn() {\n\treturn!a === b ? d : c\n\treturn((!((a)) == b)) ? d : c\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 2, Column: 8, EndLine: 2, EndColumn: 16},
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 3, Column: 10, EndLine: 3, EndColumn: 21},
			},
		},
		{
			Code: "if (!a) {\n\tb();\n} else if (!c) {\n\td();\n} else {\n\te();\n}", FileName: "case.js",
			Output: []string{"if (!a) {\n\tb();\n} else if (c) {\n\te();\n} else {\n\td();\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 3, Column: 12, EndLine: 3, EndColumn: 14},
			},
		},
		{
			Code: "!x ? /* one */ 1 : 2", FileName: "case.js",
			Output: []string{"x ? 2 : /* one */ 1"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			},
		},
		{
			Code: "!x ? 1 : /* two */ 2", FileName: "case.js",
			Output: []string{"x ? /* two */ 2 : 1"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			},
		},
		{
			Code: "!x ? /* one */ /* first */ 1 : /* two */ /* second */ 2", FileName: "case.js",
			Output: []string{"x ? /* two */ /* second */ 2 : /* one */ /* first */ 1"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			},
		},
		{
			Code: "!x ? /* one */ 1 : /* two */ 1", FileName: "case.js",
			Output: []string{"x ? /* two */ 1 : /* one */ 1"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			},
		},
		{
			Code: "!x ? /* one */ ((1)) : /* two */ ((2))", FileName: "case.js",
			Output: []string{"x ? /* two */ ((2)) : /* one */ ((1))"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			},
		},
		{
			Code: "!x ? (/* one */ 1) : (/* two */ 2)", FileName: "case.js",
			Output: []string{"x ? (/* two */ 2) : (/* one */ 1)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			},
		},
		{
			Code: "!x ? /* one */ (1 /* inside */) : /* two */ 2", FileName: "case.js",
			Output: []string{"x ? /* two */ 2 : /* one */ (1 /* inside */)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			},
		},
		{
			Code: "!x ? /* one */ first(/* inside */ 1) : /* two */ second(2)", FileName: "case.js",
			Output: []string{"x ? /* two */ second(2) : /* one */ first(/* inside */ 1)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			},
		},
		{
			Code: "x != y ? /* one */ 1 : /* two */ 2", FileName: "case.js",
			Output: []string{"x == y ? /* two */ 2 : /* one */ 1"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
			},
		},
		{
			Code: "x !== y ? /* one */ 1 : /* two */ 2", FileName: "case.js",
			Output: []string{"x === y ? /* two */ 2 : /* one */ 1"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
			},
		},
		{
			Code: "!x ? // one\n 1 : // two\n 2", FileName: "case.js",
			Output: []string{"x ? // two\n 2 : // one\n 1"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			},
		},
		{
			Code: "!x ? /* one */ 1 : // two\n 2", FileName: "case.js",
			Output: []string{"x ? // two\n 2 : /* one */ 1"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			},
		},
		{
			Code: "!x ? // one\r\n 1 : 2", FileName: "case.js",
			Output: []string{"x ? 2 : // one\r\n 1"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			},
		},
		{
			Code: "!x ? /* outer */ (!y ? /* one */ 1 : /* two */ 2) : /* three */ 3", FileName: "case.js",
			Output: []string{"x ? /* three */ 3 : /* outer */ (!y ? /* one */ 1 : /* two */ 2)", "x ? /* three */ 3 : /* outer */ (y ? /* two */ 2 : /* one */ 1)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21},
			},
		},
		{
			Code: "if (!x) /* one */ {one();} else /* two */ {two();}", FileName: "case.js",
			Output: []string{"if (x) /* two */ {two();} else /* one */ {one();}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 7},
			},
		},
		{
			Code: "if (!x) /* one */ one(); else /* two */ two();", FileName: "case.js",
			Output: []string{"if (x) {/* two */ two();} else {/* one */ one();}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 7},
			},
		},
		{
			Code: "if (!x) // one\n one(); else // two\n two();", FileName: "case.js",
			Output: []string{"if (x) {// two\n two();} else {// one\n one();}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 7},
			},
		},
		{
			Code: "if (!x) { /* one */ one(); } else { /* two */ two(); }", FileName: "case.js",
			Output: []string{"if (x) { /* two */ two(); } else { /* one */ one(); }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 7},
			},
		},
		{
			Code: "if (!x) /* one */ one(); else {two();}", FileName: "case.js",
			Output: []string{"if (x) {two();} else {/* one */ one();}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 7},
			},
		},
		{
			Code: "if (!x) {one();} else /* two */ two();", FileName: "case.js",
			Output: []string{"if (x) {/* two */ two();} else {one();}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 7},
			},
		},
		{
			Code: "!x ? /* one */ (one as number) : /* two */ two!", FileName: "case.ts",
			Output: []string{"x ? /* two */ two! : /* one */ (one as number)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			},
		},
	})
}

func TestNoNegatedConditionUpstreamDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_negated_condition.NoNegatedConditionRule, []rule_tester.ValidTestCase{
		{Code: "if (a) {\n\tdoSomethingB();\n} else {\n\tdoSomethingC();\n}", FileName: "case.js"},
		{Code: "if (a === b) {\n\tdoSomethingB();\n} else {\n\tdoSomethingC();\n}", FileName: "case.js"},
		{Code: "a ? b : c", FileName: "case.js"},
		{Code: "if (a == b) {\n\tdoSomethingB();\n} else {\n\tdoSomethingC();\n}", FileName: "case.js"},
		{Code: "if (!a) {\n\tdoSomething();\n}", FileName: "case.js"},
		{Code: "if (!a) {\n\tdoSomething();\n} else if (b) {\n\tdoSomethingElse();\n}", FileName: "case.js"},
		{Code: "if (a != b) {\n\tdoSomething();\n}", FileName: "case.js"},
	}, []rule_tester.InvalidTestCase{
		{
			Code: "if (!a) {\n\tdoSomethingC();\n} else {\n\tdoSomethingB();\n}", FileName: "case.js",
			Output: []string{"if (a) {\n\tdoSomethingB();\n} else {\n\tdoSomethingC();\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 7},
			},
		},
		{
			Code: "if (a !== b) {\n\tdoSomethingC();\n} else {\n\tdoSomethingB();\n}", FileName: "case.js",
			Output: []string{"if (a === b) {\n\tdoSomethingB();\n} else {\n\tdoSomethingC();\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 12},
			},
		},
		{
			Code: "!a ? c : b", FileName: "case.js",
			Output: []string{"a ? b : c"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			},
		},
		{
			Code: "if (a != b) {\n\tdoSomethingC();\n} else {\n\tdoSomethingB();\n}", FileName: "case.js",
			Output: []string{"if (a == b) {\n\tdoSomethingB();\n} else {\n\tdoSomethingC();\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 11},
			},
		},
	})
}
