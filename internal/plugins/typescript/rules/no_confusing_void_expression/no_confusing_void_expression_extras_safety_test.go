package no_confusing_void_expression

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// These safety cases intentionally differ from upstream v8.67.0's unsafe fixes.
// Diagnostic messages and ranges are still pinned to the upstream input.
func TestNoConfusingVoidExpressionSafety(t *testing.T) {
	cases := []struct {
		name string
		test rule_tester.InvalidTestCase
	}{
		{name: "return-iife", test: rule_tester.InvalidTestCase{
			Code:   "function f() { return (function () {}()); }",
			Output: []string{"function f() { ;(function () {}()); }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 24, EndLine: 1, EndColumn: 40},
			},
		}},
		{name: "arrow-iife", test: rule_tester.InvalidTestCase{
			Code:   "const f = () => (function () {}());",
			Output: []string{"const f = () => { (function () {}()); };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 1, Column: 18, EndLine: 1, EndColumn: 34},
			},
		}},
		{name: "return-object", test: rule_tester.InvalidTestCase{
			Code:   "function f() { return ({ m() {} }.m()); }",
			Output: []string{"function f() { ;({ m() {} }.m()); }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 24, EndLine: 1, EndColumn: 38},
			},
		}},
		{name: "arrow-object", test: rule_tester.InvalidTestCase{
			Code:   "const f = () => ({ m() {} }.m());",
			Output: []string{"const f = () => { ({ m() {} }.m()); };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 1, Column: 18, EndLine: 1, EndColumn: 32},
			},
		}},
		{name: "return-class", test: rule_tester.InvalidTestCase{
			Code:   "function f() { return (class { static m() {} }.m()); }",
			Output: []string{"function f() { ;(class { static m() {} }.m()); }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 24, EndLine: 1, EndColumn: 51},
			},
		}},
		{name: "arrow-class", test: rule_tester.InvalidTestCase{
			Code:   "const f = () => (class { static m() {} }.m());",
			Output: []string{"const f = () => { (class { static m() {} }.m()); };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 1, Column: 18, EndLine: 1, EndColumn: 45},
			},
		}},
		{name: "return-objectConditional", test: rule_tester.InvalidTestCase{
			Code:   "function f() { return ({} ? console.log(\"yes\") : console.log(\"no\")); }",
			Output: []string{"function f() { ;({} ? console.log(\"yes\") : console.log(\"no\")); }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 29, EndLine: 1, EndColumn: 47},
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 50, EndLine: 1, EndColumn: 67},
			},
		}},
		{name: "arrow-objectConditional", test: rule_tester.InvalidTestCase{
			Code:   "const f = () => ({} ? console.log(\"yes\") : console.log(\"no\"));",
			Output: []string{"const f = () => { ({} ? console.log(\"yes\") : console.log(\"no\")); };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 1, Column: 23, EndLine: 1, EndColumn: 41},
				{MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 1, Column: 44, EndLine: 1, EndColumn: 61},
			},
		}},
		{name: "asi-plus", test: rule_tester.InvalidTestCase{
			Code: `function f() {
  let n = 1
  return (+0 ? console.log('yes') : console.log('no'));
}`,
			Output: []string{`function f() {
  let n = 1
  ;+0 ? console.log('yes') : console.log('no');
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 16, EndLine: 3, EndColumn: 34},
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 37, EndLine: 3, EndColumn: 54},
			},
		}},
		{name: "asi-minus", test: rule_tester.InvalidTestCase{
			Code: `function f() {
  let n = 1
  return (-0 ? console.log('yes') : console.log('no'));
}`,
			Output: []string{`function f() {
  let n = 1
  ;-0 ? console.log('yes') : console.log('no');
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 16, EndLine: 3, EndColumn: 34},
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 37, EndLine: 3, EndColumn: 54},
			},
		}},
		{name: "asi-regex", test: rule_tester.InvalidTestCase{
			Code: `function f() {
  let n = 1
  return (/x/.test("y") ? console.log('yes') : console.log('no'));
}`,
			Output: []string{`function f() {
  let n = 1
  ;/x/.test("y") ? console.log('yes') : console.log('no');
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 27, EndLine: 3, EndColumn: 45},
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 48, EndLine: 3, EndColumn: 65},
			},
		}},
		{name: "asi-angle", test: rule_tester.InvalidTestCase{
			Code: `function f() {
  let n = 1
  return (<boolean>false ? console.log('yes') : console.log('no'));
}`,
			Output: []string{`function f() {
  let n = 1
  ;<boolean>false ? console.log('yes') : console.log('no');
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 28, EndLine: 3, EndColumn: 46},
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 49, EndLine: 3, EndColumn: 66},
			},
		}},
		{name: "asi-bracket", test: rule_tester.InvalidTestCase{
			Code: `function f() {
  let n = 1
  return ([0][0] ? console.log('yes') : console.log('no'));
}`,
			Output: []string{`function f() {
  let n = 1
  ;[0][0] ? console.log('yes') : console.log('no');
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 20, EndLine: 3, EndColumn: 38},
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 41, EndLine: 3, EndColumn: 58},
			},
		}},
		{name: "asi-template", test: rule_tester.InvalidTestCase{
			Code:   "function f() {\n  let n = 1\n  return (`x` === \"y\" ? console.log('yes') : console.log('no'));\n}",
			Output: []string{"function f() {\n  let n = 1\n  ;`x` === \"y\" ? console.log('yes') : console.log('no');\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 25, EndLine: 3, EndColumn: 43},
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 46, EndLine: 3, EndColumn: 63},
			},
		}},
		{name: "asi-object", test: rule_tester.InvalidTestCase{
			Code: `function f() {
  let n = 1
  return ({} ? console.log('yes') : console.log('no'));
}`,
			Output: []string{`function f() {
  let n = 1
  ;({} ? console.log('yes') : console.log('no'));
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 16, EndLine: 3, EndColumn: 34},
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 37, EndLine: 3, EndColumn: 54},
			},
		}},
		{name: "asi-asyncFn", test: rule_tester.InvalidTestCase{
			Code: `function f() {
  let n = 1
  return (async function () {} ? console.log('yes') : console.log('no'));
}`,
			Output: []string{`function f() {
  let n = 1
  ;(async function () {} ? console.log('yes') : console.log('no'));
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 34, EndLine: 3, EndColumn: 52},
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 55, EndLine: 3, EndColumn: 72},
			},
		}},
		{name: "bare-asi-plus", test: rule_tester.InvalidTestCase{
			Code: `function f() {
  let n = 1
  return +0 ? console.log('yes') : console.log('no');
}`,
			Output: []string{`function f() {
  let n = 1
  ;+0 ? console.log('yes') : console.log('no');
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 15, EndLine: 3, EndColumn: 33},
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 36, EndLine: 3, EndColumn: 53},
			},
		}},
		{name: "bare-asi-minus", test: rule_tester.InvalidTestCase{
			Code: `function f() {
  let n = 1
  return -0 ? console.log('yes') : console.log('no');
}`,
			Output: []string{`function f() {
  let n = 1
  ;-0 ? console.log('yes') : console.log('no');
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 15, EndLine: 3, EndColumn: 33},
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 36, EndLine: 3, EndColumn: 53},
			},
		}},
		{name: "bare-asi-regex", test: rule_tester.InvalidTestCase{
			Code: `function f() {
  let n = 1
  return /x/.test("y") ? console.log('yes') : console.log('no');
}`,
			Output: []string{`function f() {
  let n = 1
  ;/x/.test("y") ? console.log('yes') : console.log('no');
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 26, EndLine: 3, EndColumn: 44},
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 47, EndLine: 3, EndColumn: 64},
			},
		}},
		{name: "bare-asi-template", test: rule_tester.InvalidTestCase{
			Code:   "function f() {\n  let n = 1\n  return `x` === \"y\" ? console.log('yes') : console.log('no');\n}",
			Output: []string{"function f() {\n  let n = 1\n  ;`x` === \"y\" ? console.log('yes') : console.log('no');\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 24, EndLine: 3, EndColumn: 42},
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 45, EndLine: 3, EndColumn: 62},
			},
		}},
		{name: "bare-asi-templateHead", test: rule_tester.InvalidTestCase{
			Code:   "function f() {\n  let n = 1\n  return `x${0}` === \"y\" ? console.log('yes') : console.log('no');\n}",
			Output: []string{"function f() {\n  let n = 1\n  ;`x${0}` === \"y\" ? console.log('yes') : console.log('no');\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 28, EndLine: 3, EndColumn: 46},
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 49, EndLine: 3, EndColumn: 66},
			},
		}},
		{name: "bare-asi-angle", test: rule_tester.InvalidTestCase{
			Code: `function f() {
  let n = 1
  return <boolean>false ? console.log('yes') : console.log('no');
}`,
			Output: []string{`function f() {
  let n = 1
  ;<boolean>false ? console.log('yes') : console.log('no');
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 27, EndLine: 3, EndColumn: 45},
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 48, EndLine: 3, EndColumn: 65},
			},
		}},
		{name: "bare-asi-not", test: rule_tester.InvalidTestCase{
			Code: `function f() {
  let n = 1
  return !true ? console.log('yes') : console.log('no');
}`,
			Output: []string{`function f() {
  let n = 1
  !true ? console.log('yes') : console.log('no');
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 18, EndLine: 3, EndColumn: 36},
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 39, EndLine: 3, EndColumn: 56},
			},
		}},
		{name: "bare-asi-tilde", test: rule_tester.InvalidTestCase{
			Code: `function f() {
  let n = 1
  return ~(-1) ? console.log('yes') : console.log('no');
}`,
			Output: []string{`function f() {
  let n = 1
  ~(-1) ? console.log('yes') : console.log('no');
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 18, EndLine: 3, EndColumn: 36},
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 39, EndLine: 3, EndColumn: 56},
			},
		}},
		{name: "bare-asi-typeOf", test: rule_tester.InvalidTestCase{
			Code: `function f() {
  let n = 1
  return typeof 0 === "string" ? console.log('yes') : console.log('no');
}`,
			Output: []string{`function f() {
  let n = 1
  typeof 0 === "string" ? console.log('yes') : console.log('no');
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 34, EndLine: 3, EndColumn: 52},
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 55, EndLine: 3, EndColumn: 72},
			},
		}},
		{name: "bare-asi-voidOp", test: rule_tester.InvalidTestCase{
			Code: `function f() {
  let n = 1
  return void 0 ? console.log('yes') : console.log('no');
}`,
			Output: []string{`function f() {
  let n = 1
  void 0 ? console.log('yes') : console.log('no');
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 19, EndLine: 3, EndColumn: 37},
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 40, EndLine: 3, EndColumn: 57},
			},
		}},
		{name: "bare-asi-identifier", test: rule_tester.InvalidTestCase{
			Code: `function f() {
  let n = 1
  return Boolean(0) ? console.log('yes') : console.log('no');
}`,
			Output: []string{`function f() {
  let n = 1
  Boolean(0) ? console.log('yes') : console.log('no');
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 23, EndLine: 3, EndColumn: 41},
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 44, EndLine: 3, EndColumn: 61},
			},
		}},
		{name: "arrow-bare-iife", test: rule_tester.InvalidTestCase{
			Code:   "const f = () => function () {}();",
			Output: []string{"const f = () => { (function () {}()); };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 1, Column: 17, EndLine: 1, EndColumn: 33},
			},
		}},
		{name: "arrow-bare-class", test: rule_tester.InvalidTestCase{
			Code:   "const f = () => class { static m() {} }.m();",
			Output: []string{"const f = () => { (class { static m() {} }.m()); };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 1, Column: 17, EndLine: 1, EndColumn: 44},
			},
		}},
		{name: "comment-directive", test: rule_tester.InvalidTestCase{
			Code:   "function f() { return (/* istanbul ignore next */ console.log()); }",
			Output: []string{"function f() { ;(/* istanbul ignore next */ console.log()); }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 51, EndLine: 1, EndColumn: 64},
			},
		}},
		{name: "comment-return-34", test: rule_tester.InvalidTestCase{
			Code:   "function f() { return /*KEEP*/ console.log(); }",
			Output: []string{"function f() { /*KEEP*/ console.log(); }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 32, EndLine: 1, EndColumn: 45},
			},
		}},
		{name: "comment-arrow-35", test: rule_tester.InvalidTestCase{
			Code:   "const f = () => /*KEEP*/ console.log();",
			Output: []string{"const f = () => /*KEEP*/ { console.log(); };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 1, Column: 26, EndLine: 1, EndColumn: 39},
			},
		}},
		{name: "comment-return-36", test: rule_tester.InvalidTestCase{
			Code:   "function f() { return ( /*KEEP*/ console.log()); }",
			Output: []string{"function f() { ;( /*KEEP*/ console.log()); }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 34, EndLine: 1, EndColumn: 47},
			},
		}},
		{name: "comment-arrow-37", test: rule_tester.InvalidTestCase{
			Code:   "const f = () => ( /*KEEP*/ console.log());",
			Output: []string{"const f = () => { ( /*KEEP*/ console.log()); };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 1, Column: 28, EndLine: 1, EndColumn: 41},
			},
		}},
		{name: "comment-return-38", test: rule_tester.InvalidTestCase{
			Code:   "function f() { return (console.log() /*KEEP*/); }",
			Output: []string{"function f() { ;(console.log() /*KEEP*/); }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 24, EndLine: 1, EndColumn: 37},
			},
		}},
		{name: "comment-arrow-39", test: rule_tester.InvalidTestCase{
			Code:   "const f = () => (console.log() /*KEEP*/);",
			Output: []string{"const f = () => { (console.log() /*KEEP*/); };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 1, Column: 18, EndLine: 1, EndColumn: 31},
			},
		}},
		{name: "comment-return-40", test: rule_tester.InvalidTestCase{
			Code: `function f() { return (console.log() //KEEP
); }`,
			Output: []string{`function f() { ;(console.log() //KEEP
); }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 24, EndLine: 1, EndColumn: 37},
			},
		}},
		{name: "comment-arrow-41", test: rule_tester.InvalidTestCase{
			Code: `const f = () => (console.log() //KEEP
);`,
			Output: []string{`const f = () => { (console.log() //KEEP
); };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 1, Column: 18, EndLine: 1, EndColumn: 31},
			},
		}},
		{name: "comment-return-42", test: rule_tester.InvalidTestCase{
			Code: `function f() { return (
//KEEP
console.log()); }`,
			Output: []string{`function f() { ;(
//KEEP
console.log()); }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 1, EndLine: 3, EndColumn: 14},
			},
		}},
		{name: "comment-arrow-43", test: rule_tester.InvalidTestCase{
			Code: `const f = () => (
//KEEP
console.log());`,
			Output: []string{`const f = () => { (
//KEEP
console.log()); };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 3, Column: 1, EndLine: 3, EndColumn: 14},
			},
		}},
		{name: "mixed-non-final", test: rule_tester.InvalidTestCase{
			Code: "function f(flag: boolean) { return flag ? console.log('yes') : 42; console.log('unreachable'); }",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturn", Message: "Returning a void expression from a function is forbidden. Please move it before the `return` statement.", Line: 1, Column: 43, EndLine: 1, EndColumn: 61},
			},
		}},
		{name: "mixed-if", test: rule_tester.InvalidTestCase{
			Code: "function f(flag: boolean) { if (true) return flag ? console.log('yes') : 42; return 7; }",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturn", Message: "Returning a void expression from a function is forbidden. Please move it before the `return` statement.", Line: 1, Column: 53, EndLine: 1, EndColumn: 71},
			},
		}},
		{name: "mixed-final-control", test: rule_tester.InvalidTestCase{
			Code: "function f(flag: boolean) { return flag ? console.log('yes') : 42; }",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 43, EndLine: 1, EndColumn: 61},
			},
		}},
		{name: "mixed-arrow-control", test: rule_tester.InvalidTestCase{
			Code: "const f = (flag: boolean) => flag ? console.log('yes') : 42;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 1, Column: 37, EndLine: 1, EndColumn: 55},
			},
		}},
		{name: "arrow-let-bracket", test: rule_tester.InvalidTestCase{
			Code:   "declare const let: any; const f = () => (let[0] ? console.log() : console.log());",
			Output: []string{"declare const let: any; const f = () => { (let[0] ? console.log() : console.log()); };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 1, Column: 51, EndLine: 1, EndColumn: 64},
				{MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 1, Column: 67, EndLine: 1, EndColumn: 80},
			},
		}},
		{name: "return-let-bracket", test: rule_tester.InvalidTestCase{
			Code:   "declare const let: any; function f() { return (let[0] ? console.log() : console.log()); }",
			Output: []string{"declare const let: any; function f() { ;(let[0] ? console.log() : console.log()); }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 57, EndLine: 1, EndColumn: 70},
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 73, EndLine: 1, EndColumn: 86},
			},
		}},
		{name: "crlf-plus", test: rule_tester.InvalidTestCase{
			Code:   "function f() {\r\n let n = 1\r\n return (+0 ? console.log('yes') : console.log('no'));\r\n}",
			Output: []string{"function f() {\r\n let n = 1\r\n ;+0 ? console.log('yes') : console.log('no');\r\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 15, EndLine: 3, EndColumn: 33},
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 3, Column: 36, EndLine: 3, EndColumn: 53},
			},
		}},
		{name: "unicode-comment", test: rule_tester.InvalidTestCase{
			Code:   "function f() { return (/* 😀 KEEP */ console.log()); }",
			Output: []string{"function f() { ;(/* 😀 KEEP */ console.log()); }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 38, EndLine: 1, EndColumn: 51},
			},
		}}, {name: "bare-return-iife", test: rule_tester.InvalidTestCase{
			Code:   "function f() { return function () {}(); }",
			Output: []string{"function f() { ;(function () {}()); }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 23, EndLine: 1, EndColumn: 39},
			},
		}},
		{name: "bare-return-object", test: rule_tester.InvalidTestCase{
			Code:   "function f() { return { m() {} }.m(); }",
			Output: []string{"function f() { ;({ m() {} }.m()); }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 23, EndLine: 1, EndColumn: 37},
			},
		}},
		{name: "bare-return-class", test: rule_tester.InvalidTestCase{
			Code:   "function f() { return class { static m() {} }.m(); }",
			Output: []string{"function f() { ;(class { static m() {} }.m()); }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 23, EndLine: 1, EndColumn: 50},
			},
		}},
		{name: "callee-already-grouped", test: rule_tester.InvalidTestCase{
			Code:   "const f = () => (function () {})();",
			Output: []string{"const f = () => { (function () {})(); };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 1, Column: 17, EndLine: 1, EndColumn: 35},
			},
		}},
		{name: "object-already-grouped", test: rule_tester.InvalidTestCase{
			Code:   "const f = () => ({ m() {} }).m();",
			Output: []string{"const f = () => { ({ m() {} }).m(); };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 1, Column: 17, EndLine: 1, EndColumn: 33},
			},
		}},
		{name: "comments-before-and-after", test: rule_tester.InvalidTestCase{
			Code:   "function f() { return /*before*/ (console.log() /*inside*/) /*after*/; }",
			Output: []string{"function f() { ;/*before*/ (console.log() /*inside*/) /*after*/; }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 35, EndLine: 1, EndColumn: 48},
			},
		}},
		{name: "comment-before-semicolon", test: rule_tester.InvalidTestCase{
			Code: `function f() { return console.log() //KEEP
; }`,
			Output: []string{`function f() { console.log() //KEEP
; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 23, EndLine: 1, EndColumn: 36},
			},
		}},
		{name: "early-return-comment", test: rule_tester.InvalidTestCase{
			Code:   "function f() { if (true) return /*KEEP*/ console.log(); }",
			Output: []string{"function f() { if (true) { /*KEEP*/ console.log(); return; } }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturn", Message: "Returning a void expression from a function is forbidden. Please move it before the `return` statement.", Line: 1, Column: 42, EndLine: 1, EndColumn: 55},
			},
		}},
		{name: "early-return-trailing-comment", test: rule_tester.InvalidTestCase{
			Code: `function f() { if (true) return console.log() //KEEP
; }`,
			Output: []string{`function f() { if (true) { console.log() //KEEP
; return; } }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturn", Message: "Returning a void expression from a function is forbidden. Please move it before the `return` statement.", Line: 1, Column: 33, EndLine: 1, EndColumn: 46},
			},
		}},
		{name: "comment-only-no-semicolon", test: rule_tester.InvalidTestCase{
			Code: `function f() { return console.log() /*KEEP*/
 }`,
			Output: []string{`function f() { console.log(); /*KEEP*/
 }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 23, EndLine: 1, EndColumn: 36},
			},
		}},
		{name: "callee-comment", test: rule_tester.InvalidTestCase{
			Code:   "function f() { return console /*KEEP*/ .log(); }",
			Output: []string{"function f() { console /*KEEP*/ .log(); }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 23, EndLine: 1, EndColumn: 46},
			},
		}},
		{name: "arrow-line-comment-before-body", test: rule_tester.InvalidTestCase{
			Code: `const f = () => //KEEP
console.log();`,
			Output: []string{`const f = () => //KEEP
{ console.log(); };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 2, Column: 1, EndLine: 2, EndColumn: 14},
			},
		}},
		{name: "arrow-crlf-comment", test: rule_tester.InvalidTestCase{
			Code:   "const f = () => (\r\n//KEEP\r\nconsole.log());",
			Output: []string{"const f = () => { (\r\n//KEEP\r\nconsole.log()); };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 3, Column: 1, EndLine: 3, EndColumn: 14},
			},
		}},
		{name: "regexp-slash-equals", test: rule_tester.InvalidTestCase{
			Code: `function f() { let n = 1
return /=/.test("x") ? console.log(1) : console.log(2); }`,
			Output: []string{`function f() { let n = 1
;/=/.test("x") ? console.log(1) : console.log(2); }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 2, Column: 24, EndLine: 2, EndColumn: 38},
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 2, Column: 41, EndLine: 2, EndColumn: 55},
			},
		}},
		{name: "return-jsx", test: rule_tester.InvalidTestCase{
			Code: `function f() { let n = 1
return <button /> ? console.log(1) : console.log(2); }`,
			Output: []string{`function f() { let n = 1
;<button /> ? console.log(1) : console.log(2); }`},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 2, Column: 21, EndLine: 2, EndColumn: 35},
				{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 2, Column: 38, EndLine: 2, EndColumn: 52},
			},
		}},
		{name: "labelled-return", test: rule_tester.InvalidTestCase{
			Code:   "function f() { label: return console.log(); }",
			Output: []string{"function f() { label: { console.log(); return; } }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturn", Message: "Returning a void expression from a function is forbidden. Please move it before the `return` statement.", Line: 1, Column: 30, EndLine: 1, EndColumn: 43},
			},
		}},
		{name: "switch-return", test: rule_tester.InvalidTestCase{
			Code:   "function f() { switch (1) { case 1: return console.log(); } }",
			Output: []string{"function f() { switch (1) { case 1: { console.log(); return; } } }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturn", Message: "Returning a void expression from a function is forbidden. Please move it before the `return` statement.", Line: 1, Column: 44, EndLine: 1, EndColumn: 57},
			},
		}},
		{name: "mixed-logical-return", test: rule_tester.InvalidTestCase{
			Code: "function f(flag: boolean) { if (true) return flag && console.log(); }",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturn", Message: "Returning a void expression from a function is forbidden. Please move it before the `return` statement.", Line: 1, Column: 54, EndLine: 1, EndColumn: 67},
			},
		}},
		{name: "mixed-void-operator", test: rule_tester.InvalidTestCase{
			Code:    "function f(flag: boolean) { if (true) return flag ? console.log() : 42; }",
			Output:  []string{"function f(flag: boolean) { if (true) return flag ? void console.log() : 42; }"},
			Options: map[string]any{"ignoreVoidOperator": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturnWrapVoid", Message: "Void expressions returned from a function must be marked explicitly with the `void` operator.", Line: 1, Column: 53, EndLine: 1, EndColumn: 66},
			},
		}},
		{name: "mixed-context-option", test: rule_tester.InvalidTestCase{
			Code:    "function f(flag: boolean) { if (true) return flag ? console.log() : 42; }",
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprReturn", Message: "Returning a void expression from a function is forbidden. Please move it before the `return` statement.", Line: 1, Column: 53, EndLine: 1, EndColumn: 66},
			},
		}},
	}
	filename := tspath.ResolvePath(fixtures.GetRootDir().Dir, "void-safety.ts")
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			kind := core.ScriptKindTS
			if item.test.Tsx {
				kind = core.ScriptKindTSX
			}
			for _, code := range append([]string{item.test.Code}, item.test.Output...) {
				file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: filename, Path: tspath.Path(filename)}, code, kind)
				if len(file.Diagnostics()) != 0 {
					t.Fatalf("safety fixture must parse without errors: %v\n%s", file.Diagnostics(), code)
				}
			}
			rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoConfusingVoidExpressionRule, nil, []rule_tester.InvalidTestCase{item.test})
		})
	}
}
