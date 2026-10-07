// Additional AST, comment and statement-boundary cases checked against Unicorn v77.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/test/no-negated-condition.js
package no_negated_condition_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_negated_condition"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNegatedConditionEditDemand(t *testing.T) {
	for _, testCase := range []struct{ code, output string }{
		{"if (!x) one(); else two();", "if (x) {two();} else {one();}"},
		{"x != y ? one : two", "x == y ? two : one"},
		{"function f(){return!\nready ? one : two}", "function f(){return  (\nready ? two : one)}"},
		{"!x ? one /* ambiguous */ : two", "!x ? one /* ambiguous */ : two"},
		{"!{} ? one : two", "({}) ? two : one"},
	} {
		t.Run(testCase.code, func(t *testing.T) {
			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			program, sourceFile, err := helper.CreateTestProgram(testCase.code, "edit-demand.js", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic)
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				var got []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: sourceFile.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{Name: no_negated_condition.NoNegatedConditionRule.Name, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return no_negated_condition.NoNegatedConditionRule.Run(ctx, nil)
							},
						}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { got = append(got, d) }},
				})
				if len(got) != 1 {
					t.Fatalf("demand %d: got %d diagnostics", demand, len(got))
				}
				diagnostics[demand] = got[0]
			}
			baseline, all := diagnostics[rule.EditDemandNone], diagnostics[rule.EditDemandAll]
			for demand, diagnostic := range diagnostics {
				if diagnostic.Range != baseline.Range || !reflect.DeepEqual(diagnostic.Message, baseline.Message) || diagnostic.Severity != baseline.Severity {
					t.Errorf("demand %d changed diagnostic identity", demand)
				}
				wantFix := testCase.code != testCase.output && demand&rule.EditDemandAutofix != 0
				if (diagnostic.FixesPtr != nil) != wantFix {
					t.Errorf("demand %d: unexpected autofix artifacts", demand)
				}
				if wantFix && !reflect.DeepEqual(diagnostic.FixesPtr, all.FixesPtr) {
					t.Errorf("demand %d changed autofix artifacts", demand)
				}
				if diagnostic.Suggestions != nil {
					t.Errorf("demand %d: unexpected suggestions", demand)
				}
			}
			for demand, diagnostic := range diagnostics {
				wantFix := testCase.code != testCase.output && demand&rule.EditDemandAutofix != 0
				expected := testCase.code
				if wantFix {
					expected = testCase.output
				}
				output, _, fixed := linter.ApplyRuleFixes(testCase.code, []rule.RuleDiagnostic{diagnostic})
				if output != expected || fixed != wantFix {
					t.Errorf("demand %d: got fix %q, want %q", demand, output, expected)
				}
			}
		})
	}
}

func TestNoNegatedConditionScript(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_negated_condition.NoNegatedConditionRule, nil, []rule_tester.InvalidTestCase{
		{Code: "var let=[]; !let[0] ? a : b", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			Output: []string{"var let=[]; (let[0]) ? b : a"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 13, EndLine: 1, EndColumn: 20}},
		},
		{Code: "var let=[]; (()=> !let[0] ? a : b)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			Output: []string{"var let=[]; (()=> let[0] ? b : a)"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 19, EndLine: 1, EndColumn: 26}},
		},
		{Code: "var let=[]; !let.value ? a : b", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			Output: []string{"var let=[]; let.value ? b : a"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 13, EndLine: 1, EndColumn: 23}},
		},
	})
}

func TestNoNegatedConditionExtrasSyntaxAndFixes(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_negated_condition.NoNegatedConditionRule, []rule_tester.ValidTestCase{
		{Code: "async function f(){return await!ready ? left : right}", FileName: "case.js"},
		{Code: "if (~x) one(); else two();", FileName: "case.js"},
		{Code: "if (x <= y) one(); else two();", FileName: "case.js"},
		{Code: "if (!x && y) one(); else two();", FileName: "case.js"},
		{Code: "!x || y ? one() : two()", FileName: "case.js"},
		{Code: "(!x as boolean) ? one() : two()", FileName: "case.ts"},
		{Code: "if ((!x satisfies boolean)) one(); else two();", FileName: "case.ts"},
		{Code: "(!x)! ? one() : two()", FileName: "case.ts"},
	}, []rule_tester.InvalidTestCase{
		{
			Code: "if (!((a && b))) one(); else two();", FileName: "case.js",
			Output: []string{"if (a && b) {two();} else {one();}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 16},
			},
		},
		{
			Code: "if (!/** @type {boolean} */ ((x))) one(); else two();", FileName: "case.js",
			Output: []string{"if (/** @type {boolean} */ x) {two();} else {one();}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 34},
			},
		},
		{
			Code: "if (/** @type {boolean} */ (!x)) one(); else two();", FileName: "case.js",
			Output: []string{"if (/** @type {boolean} */ (x)) {two();} else {one();}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 29, EndLine: 1, EndColumn: 31},
			},
		},
		{
			Code: "/** @type {boolean} */ (!x) ? one() : two()", FileName: "case.js",
			Output: []string{"/** @type {boolean} */ (x) ? two() : one()"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 25, EndLine: 1, EndColumn: 27},
			},
		},
		{
			Code: "if (x /* left */ != /* right */ y) one(); else two();", FileName: "case.js",
			Output: []string{"if (x /* left */ == /* right */ y) {two();} else {one();}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 34},
			},
		},
		{
			Code: "if (!x) if (y) one(); else two(); else three();", FileName: "case.js",
			Output: []string{"if (x) {three();} else {if (y) one(); else two();}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 7},
			},
		},
		{
			Code: "if (!x) one(); else one();", FileName: "case.js",
			Output: []string{"if (x) one(); else one();"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 7},
			},
		},
		{
			Code: "!x ? /* same */ one() : /* same */ one()", FileName: "case.js",
			Output: []string{"x ? /* same */ one() : /* same */ one()"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			},
		},
		{
			Code: "!x ? one() /* ambiguous */ : two()", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			},
		},
		{
			Code: "!x ? /a\\/\\/b/.test(s) : \"/* not a comment */\"", FileName: "case.js",
			Output: []string{"x ? \"/* not a comment */\" : /a\\/\\/b/.test(s)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			},
		},
		{
			Code: "!obj?.[key] ? left : right", FileName: "case.js",
			Output: []string{"obj?.[key] ? right : left"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 12},
			},
		},
		{
			Code: "class C { #ready; m() { return !this.#ready ? left : right; } }", FileName: "case.js",
			Output: []string{"class C { #ready; m() { return this.#ready ? right : left; } }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 32, EndLine: 1, EndColumn: 44},
			},
		},
		{
			Code: "const view = !ready ? <Pending/> : <Ready/>;", FileName: "case.js",
			Output: []string{"const view = ready ? <Ready/> : <Pending/>;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 14, EndLine: 1, EndColumn: 20},
			},
		},
		{
			Code: "const view = !ready ? <Pending/> : <Ready/>;", FileName: "case.tsx",
			Output: []string{"const view = ready ? <Ready/> : <Pending/>;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 14, EndLine: 1, EndColumn: 20},
			},
		},
		{
			Code: "!foo! ? (left as number) : (right satisfies number)", FileName: "case.ts",
			Output: []string{"foo! ? (right satisfies number) : (left as number)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
			},
		},
		{
			Code: "!<boolean>foo ? left : right", FileName: "case.ts",
			Output: []string{"<boolean>foo ? right : left"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
			},
		},
		{
			Code: "\"😀\"; !ready ? left : right;", FileName: "case.js",
			Output: []string{"\"😀\"; ready ? right : left;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
			},
		},
		{
			Code: "if (\n a !==\n b\n) one(); else two();", FileName: "case.js",
			Output: []string{"if (\n a ===\n b\n) {two();} else {one();}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 2, Column: 2, EndLine: 3, EndColumn: 3},
			},
		},
		{
			Code: "function f(){return!/* keep\n */ready ? left : right;}", FileName: "case.js",
			Output: []string{"function f(){return  (/* keep\n */ready ? right : left);}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 20, EndLine: 2, EndColumn: 9},
			},
		},
		{
			Code: "function f(){throw!\nready ? left : right}", FileName: "case.js",
			Output: []string{"function f(){throw  (\nready ? right : left)}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 19, EndLine: 2, EndColumn: 6},
			},
		},
		{
			Code: "function f(){return! ready ? left : right}", FileName: "case.js",
			Output: []string{"function f(){return  ( ready ? right : left)}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 20, EndLine: 2, EndColumn: 6},
			},
		},
		{
			Code: "const f = function() {}\n!(ready) ? left : right", FileName: "case.js",
			Output: []string{"const f = function() {}\n;(ready) ? right : left"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 2, Column: 1, EndLine: 2, EndColumn: 9},
			},
		},
		{
			Code: "const C = class {}\n![] ? left : right", FileName: "case.js",
			Output: []string{"const C = class {}\n;[] ? right : left"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 2, Column: 1, EndLine: 2, EndColumn: 4},
			},
		},
		{
			Code: "if (ready) ![x] ? left : right", FileName: "case.js",
			Output: []string{"if (ready) [x] ? right : left"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 12, EndLine: 1, EndColumn: 16},
			},
		},
		{
			Code: "while (ready) ![x] ? left : right", FileName: "case.js",
			Output: []string{"while (ready) [x] ? right : left"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 15, EndLine: 1, EndColumn: 19},
			},
		},
		{
			Code: "const f = () => ![x] ? left : right", FileName: "case.js",
			Output: []string{"const f = () => [x] ? right : left"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 17, EndLine: 1, EndColumn: 21},
			},
		},
		{
			Code: "x;\n![] ? left : right", FileName: "case.js",
			Output: []string{"x;\n[] ? right : left"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 2, Column: 1, EndLine: 2, EndColumn: 4},
			},
		},
		{
			Code: "x\n!`template` ? left : right", FileName: "case.js",
			Output: []string{"x\n;`template` ? right : left"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 2, Column: 1, EndLine: 2, EndColumn: 12},
			},
		},
		{
			Code: "x\n!/a/.test(s) ? left : right", FileName: "case.js",
			Output: []string{"x\n;/a/.test(s) ? right : left"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 2, Column: 1, EndLine: 2, EndColumn: 13},
			},
		},
		{
			Code: "x\n!-ready ? left : right", FileName: "case.js",
			Output: []string{"x\n;-ready ? right : left"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 2, Column: 1, EndLine: 2, EndColumn: 8},
			},
		},
		{
			Code: "for(const x of!ready ? left : right){}", FileName: "case.js",
			Output: []string{"for(const x of ready ? right : left){}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 15, EndLine: 1, EndColumn: 21},
			},
		},
	})
}

// These fixes remain parseable where Unicorn v77.0.0 exposes a block or declaration.
func TestNoNegatedConditionStatementExpressionFixes(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_negated_condition.NoNegatedConditionRule, nil, []rule_tester.InvalidTestCase{
		{Code: "!{} ? a : b", FileName: "case.js",
			Output: []string{"({}) ? b : a"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
		},
		{Code: "!function(){} ? a : b", FileName: "case.js",
			Output: []string{"(function(){}) ? b : a"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14}},
		},
		{Code: "!class{} ? a : b", FileName: "case.js",
			Output: []string{"(class{}) ? b : a"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9}},
		},
		{Code: "!async function(){} ? a : b", FileName: "case.js",
			Output: []string{"(async function(){}) ? b : a"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}},
		},
		{Code: "!function*(){} ? a : b", FileName: "case.js",
			Output: []string{"(function*(){}) ? b : a"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 15}},
		},
		{Code: "!{}.value ? a : b", FileName: "case.js",
			Output: []string{"({}.value) ? b : a"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 10}},
		},
		{Code: "!class{}.name ? a : b", FileName: "case.js",
			Output: []string{"(class{}.name) ? b : a"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14}},
		},
		{Code: "!async function(){}() ? a : b", FileName: "case.js",
			Output: []string{"(async function(){}()) ? b : a"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22}},
		},
		{Code: "if (ready) !{} ? a : b", FileName: "case.js",
			Output: []string{"if (ready) ({}) ? b : a"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 12, EndLine: 1, EndColumn: 15}},
		},
		{Code: "if (ready) x();else!{} ? a : b", FileName: "case.js",
			Output: []string{"if (ready) x();else ({}) ? b : a"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 20, EndLine: 1, EndColumn: 23}},
		},
		{Code: "while (ready) !{} ? a : b", FileName: "case.js",
			Output: []string{"while (ready) ({}) ? b : a"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 15, EndLine: 1, EndColumn: 18}},
		},
		{Code: "label: !{} ? a : b", FileName: "case.js",
			Output: []string{"label: ({}) ? b : a"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 8, EndLine: 1, EndColumn: 11}},
		},
		{Code: "previous()\n!{} ? a : b", FileName: "case.js",
			Output: []string{"previous()\n;({}) ? b : a"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 2, Column: 1, EndLine: 2, EndColumn: 4}},
		},
		{Code: "const x = function(){}\n!class{} ? a : b", FileName: "case.js",
			Output: []string{"const x = function(){}\n;(class{}) ? b : a"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 2, Column: 1, EndLine: 2, EndColumn: 9}},
		},
		{Code: "! /* keep */ {} ? a : b", FileName: "case.js",
			Output: []string{"( /* keep */ {}) ? b : a"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 16}},
		},
		{Code: "!\nclass{} ? a : b", FileName: "case.js",
			Output: []string{"(\nclass{}) ? b : a"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 2, EndColumn: 8}},
		},
		{Code: "!!{} ? a : b", FileName: "case.js",
			Output: []string{"!{} ? b : a", "({}) ? a : b"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 5}},
		},
		{Code: "!function<T>(){}<string> ? a : b", FileName: "case.ts",
			Output: []string{"(function<T>(){}<string>) ? b : a"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25}},
		},
		{Code: "(!{}) ? a : b", FileName: "case.js",
			Output: []string{"({}) ? b : a"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 2, EndLine: 1, EndColumn: 5}},
		},
		{Code: "!({}) ? a : b", FileName: "case.js",
			Output: []string{"({}) ? b : a"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6}},
		},
		{Code: "const x = !{} ? a : b", FileName: "case.js",
			Output: []string{"const x = {} ? b : a"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 11, EndLine: 1, EndColumn: 14}},
		},
		{Code: "function f(){return !{} ? a : b}", FileName: "case.js",
			Output: []string{"function f(){return {} ? b : a}"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 21, EndLine: 1, EndColumn: 24}},
		},
		{Code: "const f = () => !{} ? a : b", FileName: "case.js",
			Output: []string{"const f = () => ({}) ? b : a"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 17, EndLine: 1, EndColumn: 20}},
		},
		{Code: "if (!{}) one(); else two();", FileName: "case.js",
			Output: []string{"if ({}) {two();} else {one();}"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 8}},
		},
		{Code: "!async ? a : b", FileName: "case.js",
			Output: []string{"async ? b : a"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7}},
		},
		{Code: "!async.value ? a : b", FileName: "case.js",
			Output: []string{"async.value ? b : a"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-negated-condition", Message: "Unexpected negated condition.", Line: 1, Column: 1, EndLine: 1, EndColumn: 13}},
		},
	})
}
