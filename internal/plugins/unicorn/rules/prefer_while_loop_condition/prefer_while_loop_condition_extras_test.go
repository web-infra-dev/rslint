package prefer_while_loop_condition_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_while_loop_condition"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Expectations were checked against eslint-plugin-unicorn v79.0.0.
func TestPreferWhileLoopConditionExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&prefer_while_loop_condition.PreferWhileLoopConditionRule,
		[]rule_tester.ValidTestCase{
			// using binding.
			{Code: "while (true) { if (resource) break; using resource = acquire(); work(); }", FileName: "case.ts"},
			// await using binding.
			{Code: "async function run() { while (true) { if (resource) break; await using resource = acquire(); work(); } }", FileName: "case.ts"},
			// catch binding inside guard.
			{Code: "while (true) { if ((() => { try { work(); } catch (done) { return done; } })()) break; work(); }", FileName: "case.js"},
			// nested lexical destructuring.
			{Code: "while (true) { if (done) break; const {state: {done}} = input; work(); }", FileName: "case.js"},
			// ambient function binding.
			{Code: "while (true) { if (ready()) break; declare function ready(): boolean; work(); }", FileName: "case.ts"},
			// script block function.
			{Code: "while (true) { if (done) break; function done() {} work(); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
			// empty body.
			{Code: "while (true) {}", FileName: "case.js"},
			// bare body.
			{Code: "while (true) if (done) break;", FileName: "case.js"},
			// empty consequent.
			{Code: "while (true) { if (done) {} work(); }", FileName: "case.js"},
			// nested break block.
			{Code: "while (true) { if (done) {{break;}} work(); }", FileName: "case.js"},
			// non-break consequent.
			{Code: "while (true) { if (done) work(); }", FileName: "case.js"},
			// empty first statement.
			{Code: "while (true) { ; if (done) break; }", FileName: "case.js"},
			// numeric for test.
			{Code: "for (;1;) { if (done) break; }", FileName: "case.js"},
			// numeric do test.
			{Code: "do { if (done) break; } while (1);", FileName: "case.js"},
			// typed infinite test.
			{Code: "while (true as true) { if (done) break; }", FileName: "case.ts"},
			// non-null infinite test.
			{Code: "while (true!) { if (done) break; }", FileName: "case.ts"},
			// break in finally.
			{Code: "while (true) { if (done) break; try { work(); } finally { if (stop) break; } }", FileName: "case.js"},
			// break in labeled block.
			{Code: "while (true) { if (done) break; block: { if (stop) break; } }", FileName: "case.js"},
			// destructured lexical binding.
			{Code: "while (true) { if (done) break; const {done} = state; }", FileName: "case.js"},
			// array lexical binding.
			{Code: "while (true) { if (done) break; let [done] = state; }", FileName: "case.js"},
			// computed lexical reference.
			{Code: "while (true) { if (obj?.[done]) break; let done; }", FileName: "case.js"},
			// shorthand lexical reference.
			{Code: "while (true) { if ({done}) break; let done; }", FileName: "case.js"},
			// closure reference.
			{Code: "while (true) { if (check(() => done)) break; let done; }", FileName: "case.js"},
			// guard-local parameter.
			{Code: "while (true) { if (((done) => done)(state)) break; work(); }", FileName: "case.js"},
			// guard-local function name.
			{Code: "while (true) { if ((function done() { return done; })()) break; work(); }", FileName: "case.js"},
			// type reference.
			{Code: "while (true) { if (value as Local) break; type Local = boolean; }", FileName: "case.ts"},
			// JSX lexical component.
			{Code: "while (true) { if (<Widget />) break; const Widget = component; }", FileName: "case.tsx"},
		},
		[]rule_tester.InvalidTestCase{
			// nested var destructuring.
			{
				Code: "while (true) { if (done) break; var {state: {done}} = input; work(); }", FileName: "case.js",
				Output: []string{"while (!done) {  var {state: {done}} = input; work(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 32},
				},
			},
			// instantiation expression.
			{
				Code: "while (true) { if (predicate<boolean>) break; work(); }", FileName: "case.ts",
				Output: []string{"while (!predicate<boolean>) {  work(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 46},
				},
			},
			// negated optional access.
			{
				Code: "while (true) { if (!obj?.value) break; work(); }", FileName: "case.js",
				Output: []string{"while (obj?.value) {  work(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 39},
				},
			},
			// JSDoc cast on infinite test.
			{
				Code: "while (/** @type {true} */ (true)) { if (done) break; work(); }", FileName: "case.js",
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 38, EndLine: 1, EndColumn: 54},
				},
			},
			// regular expression text.
			{
				Code: "while (true) { if (/\\/\\//.test(value)) break; work(); }", FileName: "case.js",
				Output: []string{"while (!/\\/\\//.test(value)) {  work(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 46},
				},
			},
			// retained same-line comment.
			{
				Code: "while (true) { if (done) break; work(/* keep */ value); }", FileName: "case.js",
				Output: []string{"while (!done) {  work(/* keep */ value); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 32},
				},
			},
			// script hoisted nested function.
			{
				Code: "while (true) { if (done) break; { function done() {} } work(); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Output: []string{"while (!done) {  { function done() {} } work(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 32},
				},
			},
			// Unicode separators delimit the guard's standalone line.
			{
				Code: "while (true) {\u2028\tif (done) break;\u2029\twork();\u2028}", FileName: "case.js",
				Output: []string{"while (!done) {\u2029\twork();\u2028}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 2, EndColumn: 18},
				},
			},
			// destructured var binding.
			{
				Code: "while (true) { if (done) break; var {done} = state; }", FileName: "case.js",
				Output: []string{"while (!done) {  var {done} = state; }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 32},
				},
			},
			// nested lexical binding.
			{
				Code: "while (true) { if (done) break; { let done = true; work(done); } }", FileName: "case.js",
				Output: []string{"while (!done) {  { let done = true; work(done); } }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 32},
				},
			},
			// property name is not a reference.
			{
				Code: "while (true) { if (obj.done) break; let done; }", FileName: "case.js",
				Output: []string{"while (!obj.done) {  let done; }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 36},
				},
			},
			// object key is not a reference.
			{
				Code: "while (true) { if ({done: false}) break; let done; }", FileName: "case.js",
				Output: []string{"while (!{done: false}) {  let done; }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 41},
				},
			},
			// nested loop break.
			{
				Code: "while (true) { if (done) break; while (active) { break; } do { break; } while (active); for (;active;) { break; } for (const key in obj) { break; } }", FileName: "case.js",
				Output: []string{"while (!done) {  while (active) { break; } do { break; } while (active); for (;active;) { break; } for (const key in obj) { break; } }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 32},
				},
			},
			// nested functions.
			{
				Code: "while (true) { if (done) break; const f = () => { while (active) { break; } }; class Worker { run() { while (active) { break; } } } }", FileName: "case.js",
				Output: []string{"while (!done) {  const f = () => { while (active) { break; } }; class Worker { run() { while (active) { break; } } } }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 32},
				},
			},
			// labeled later break.
			{
				Code: "outer: { while (true) { if (done) break; if (stop) break outer; } }", FileName: "case.js",
				Output: []string{"outer: { while (!done) {  if (stop) break outer; } }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 25, EndLine: 1, EndColumn: 41},
				},
			},
			// parenthesized negation.
			{
				Code: "while (((true))) { if (((!((ready))))) break; work(); }", FileName: "case.js",
				Output: []string{"while (((ready))) {  work(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 20, EndLine: 1, EndColumn: 46},
				},
			},
			// double negation.
			{
				Code: "for (;;) { if (!!done) break; work(); }", FileName: "case.js",
				Output: []string{"while (!done) {  work(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 12, EndLine: 1, EndColumn: 30},
				},
			},
			// negated sequence.
			{
				Code: "do { if (!(a(), b())) break; work(); } while (true);", FileName: "case.js",
				Output: []string{"while (a(), b()) {  work(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 6, EndLine: 1, EndColumn: 29},
				},
			},
			// sequence.
			{
				Code: "while (true) { if ((a(), b())) break; work(); }", FileName: "case.js",
				Output: []string{"while (!(a(), b())) {  work(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 38},
				},
			},
			// comparison.
			{
				Code: "while (true) { if (a === b) break; work(); }", FileName: "case.js",
				Output: []string{"while (!(a === b)) {  work(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 35},
				},
			},
			// exponentiation.
			{
				Code: "while (true) { if (a ** b) break; work(); }", FileName: "case.js",
				Output: []string{"while (!(a ** b)) {  work(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 34},
				},
			},
			// assignment.
			{
				Code: "while (true) { if (a = b()) break; work(); }", FileName: "case.js",
				Output: []string{"while (!(a = b())) {  work(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 35},
				},
			},
			// ternary.
			{
				Code: "while (true) { if (a ? b : c) break; work(); }", FileName: "case.js",
				Output: []string{"while (!(a ? b : c)) {  work(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 37},
				},
			},
			// prefix increment.
			{
				Code: "while (true) { if (++index) break; work(); }", FileName: "case.js",
				Output: []string{"while (!(++index)) {  work(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 35},
				},
			},
			// postfix increment.
			{
				Code: "while (true) { if (index++) break; work(); }", FileName: "case.js",
				Output: []string{"while (!(index++)) {  work(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 35},
				},
			},
			// unary typeof.
			{
				Code: "while (true) { if (typeof value) break; work(); }", FileName: "case.js",
				Output: []string{"while (!typeof value) {  work(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 40},
				},
			},
			// await condition.
			{
				Code: "async function run() { while (true) { if (await done()) break; work(); } }", FileName: "case.js",
				Output: []string{"async function run() { while (!await done()) {  work(); } }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 39, EndLine: 1, EndColumn: 63},
				},
			},
			// yield condition.
			{
				Code: "function* run() { while (true) { if (yield done()) break; work(); } }", FileName: "case.js",
				Output: []string{"function* run() { while (!(yield done())) {  work(); } }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 34, EndLine: 1, EndColumn: 58},
				},
			},
			// as condition.
			{
				Code: "while (true) { if (value as boolean) break; work(); }", FileName: "case.ts",
				Output: []string{"while (!(value as boolean)) {  work(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 44},
				},
			},
			// satisfies condition.
			{
				Code: "while (true) { if (value satisfies boolean) break; work(); }", FileName: "case.ts",
				Output: []string{"while (!(value satisfies boolean)) {  work(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 51},
				},
			},
			// assertion condition.
			{
				Code: "while (true) { if (<boolean>value) break; work(); }", FileName: "case.ts",
				Output: []string{"while (!(<boolean>value)) {  work(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 42},
				},
			},
			// optional condition.
			{
				Code: "while (true) { if (obj?.[key]?.()) break; work(); }", FileName: "case.js",
				Output: []string{"while (!obj?.[key]?.()) {  work(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 42},
				},
			},
			// private condition.
			{
				Code: "class Worker { #done; run() { while (true) { if (this.#done) break; work(); } } }", FileName: "case.js",
				Output: []string{"class Worker { #done; run() { while (!this.#done) {  work(); } } }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 46, EndLine: 1, EndColumn: 68},
				},
			},
			// JSX intrinsic.
			{
				Code: "while (true) { if (<div />) break; let div; }", FileName: "case.tsx",
				Output: []string{"while (!<div />) {  let div; }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 35},
				},
			},
			// non-ASCII position.
			{
				Code: "const emoji = \"😀\"; while (true) { if (done) break; work(); }", FileName: "case.js",
				Output: []string{"const emoji = \"😀\"; while (!done) {  work(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 36, EndLine: 1, EndColumn: 52},
				},
			},
			// body comment-like strings.
			{
				Code: "while (true) { if (\"/* not a comment */\") break; work(\"//\"); }", FileName: "case.js",
				Output: []string{"while (!\"/* not a comment */\") {  work(\"//\"); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 49},
				},
			},
			// parenthesized comment.
			{
				Code: "while (true) { if ((/* keep */ done)) break; work(); }", FileName: "case.js",
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 45},
				},
			},
			// same-line trailing multiline comment.
			{
				Code: "while (true) { if (done) break; /* keep\ncomment */ work(); }", FileName: "case.js",
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 32},
				},
			},
			// retained body comment.
			{
				Code: "while (true) {\n if (done) break;\n /* keep */ work();\n}", FileName: "case.js",
				Output: []string{"while (!done) {\n /* keep */ work();\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 2, EndColumn: 18},
				},
			},
			// while trailing comment.
			{
				Code: "while (true) { if (done) break; work(); } // keep", FileName: "case.js",
				Output: []string{"while (!done) {  work(); } // keep"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 32},
				},
			},
			// do tail multiline comment.
			{
				Code: "do { if (done) break; work(); } while (true); /* keep\ncomment */", FileName: "case.js",
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 6, EndLine: 1, EndColumn: 22},
				},
			},
			// next-line do comment.
			{
				Code: "do { if (done) break; work(); } while (true);\n// keep", FileName: "case.js",
				Output: []string{"while (!done) {  work(); }\n// keep"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 6, EndLine: 1, EndColumn: 22},
				},
			},
			// do without space or semicolon.
			{
				Code: "do{if(done)break;work()}while(true)", FileName: "case.js",
				Output: []string{"while (!done){work()}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 4, EndLine: 1, EndColumn: 18},
				},
			},
			// nested reports.
			{
				Code: "while (true) { if (done) break; do { if (stopped) break; work(); } while (true); }", FileName: "case.js",
				Output: []string{"while (!done) {  while (!stopped) {  work(); } }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 32},
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 38, EndLine: 1, EndColumn: 57},
				},
			},
			// CR line removal.
			{
				Code: "for (;;) {\r  if (done) break;\r  work();\r}", FileName: "case.js",
				Output: []string{"while (!done) {\r  work();\r}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 3, EndLine: 2, EndColumn: 19},
				},
			},
			// JavaScript JSDoc cast.
			{
				Code: "while (true) { if (/** @type {boolean} */ (done)) break; work(); }", FileName: "case.js",
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 1, Column: 16, EndLine: 1, EndColumn: 57},
				},
			},
		},
	)
}

func TestPreferWhileLoopConditionEditDemand(t *testing.T) {
	const ruleName = "unicorn/prefer-while-loop-condition"
	for _, test := range []struct {
		name, code, output string
		suppressed         bool
	}{
		{"while", "while (true) { if (done) break; work(); }", "while (!done) {  work(); }", false},
		{"for", "for (;;) { if (done) break; work(); }", "while (!done) {  work(); }", false},
		{"do", "do { if (done) break; work(); } while (true);", "while (!done) {  work(); }", false},
		{"comment", "while (true) { if (/* keep */ done) break; work(); }", "", false},
		{"disabled", "// eslint-disable-next-line unicorn/prefer-while-loop-condition\nwhile (true) { if (done) break; work(); }", "", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			program, sourceFile, err := helper.CreateTestProgram(test.code, "edit-demand.js", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			var baseline *rule.RuleDiagnostic
			var autofixes []rule.RuleFix
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				var diagnostics []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: sourceFile.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{Name: ruleName, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return prefer_while_loop_condition.PreferWhileLoopConditionRule.Run(ctx, nil)
							},
						}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }},
				})
				count := 1
				if test.suppressed {
					count = 0
				}
				if len(diagnostics) != count {
					t.Fatalf("demand %d: got %d diagnostics, want %d", demand, len(diagnostics), count)
				}
				wantFix := test.output != "" && demand&rule.EditDemandAutofix != 0
				if count != 0 {
					diagnostic := diagnostics[0]
					if (diagnostic.FixesPtr != nil) != wantFix || diagnostic.Suggestions != nil {
						t.Fatalf("demand %d: unexpected edits: %+v", demand, diagnostic)
					}
					if diagnostic.Range.Pos() != strings.Index(test.code, "if (") ||
						diagnostic.Range.End() != strings.Index(test.code, "break;")+len("break;") ||
						diagnostic.Message.Id != "prefer-while-loop-condition" ||
						diagnostic.Message.Description != "Prefer putting the condition in the `while` statement." {
						t.Fatalf("demand %d: unexpected diagnostic: %+v", demand, diagnostic)
					}
					if wantFix {
						if autofixes != nil && !reflect.DeepEqual(autofixes, *diagnostic.FixesPtr) {
							t.Fatalf("demand %d changed fixes", demand)
						}
						autofixes = append([]rule.RuleFix(nil), (*diagnostic.FixesPtr)...)
					}
					diagnostic.FixesPtr = nil
					if baseline == nil {
						baseline = &diagnostic
					} else if !reflect.DeepEqual(*baseline, diagnostic) {
						t.Fatalf("demand %d changed diagnostic", demand)
					}
				}
				wantOutput := test.code
				if wantFix {
					wantOutput = test.output
				}
				output, _, fixed := linter.ApplyRuleFixes(test.code, diagnostics)
				if output != wantOutput || fixed != wantFix {
					t.Fatalf("demand %d: got %q (%t), want %q (%t)", demand, output, fixed, wantOutput, wantFix)
				}
			}
		})
	}
}

func TestPreferWhileLoopConditionWithTypeInformation(t *testing.T) {
	// Another enabled rule can make a checker available. Library declarations
	// must not be mistaken for lexical declarations inside this loop.
	body := strings.Repeat("  work();\n", 8000)
	source := "while (true) {\n  if (Boolean(done)) break;\n" + body + "}"
	typedRule := prefer_while_loop_condition.PreferWhileLoopConditionRule
	typedRule.Run = func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		if ctx.TypeChecker == nil {
			t.Fatal("expected a checker for the typed-program regression")
		}
		return prefer_while_loop_condition.PreferWhileLoopConditionRule.Run(ctx, options)
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &typedRule, nil,
		[]rule_tester.InvalidTestCase{{
			Code: source, FileName: "case.ts",
			Output: []string{"while (!Boolean(done)) {\n" + body + "}"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 3, EndLine: 2, EndColumn: 28}},
		}},
	)
}
