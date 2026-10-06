package no_meaningless_void_operator

import (
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoMeaninglessVoidOperatorUnionTypes(t *testing.T) {
	for _, checkNever := range []bool{false, true} {
		name := "default"
		if checkNever {
			name = "checkNever"
		}
		t.Run(name, func(t *testing.T) {
			options := map[string]any{"checkNever": checkNever}
			valid := []rule_tester.ValidTestCase{
				{Code: "declare const audio: { pause(): number } | undefined; void audio?.pause();"},
				{Code: "declare const call: (() => Promise<void>) | undefined; void call?.();"},
				{Code: "declare function call(): void | number; void call();"},
				{Code: "declare function call(): void | undefined | number; void call();"},
				{Code: "declare function call(): void | null; void call();"},
				{Code: "declare function call(): void | unknown; void call();"},
				{Code: "declare function call(): void | any; void call();"},
				{Code: "declare function call(): void & { brand: true }; void call();"},
				{Code: "function test<T extends void | undefined>(call: () => T) { void call(); }"},
				{Code: "function test<T extends never>(call: () => T) { void call(); }"},
			}
			for i := range valid {
				valid[i].Options = options
			}
			invalid := []rule_tester.InvalidTestCase{
				{
					Code:   "declare const audio: { pause(): void } | undefined;\nvoid audio?.pause();",
					Output: []string{"declare const audio: { pause(): void } | undefined;\naudio?.pause();"},
					Errors: []rule_tester.InvalidTestCaseError{{
						MessageId: "meaninglessVoidOperator",
						Message:   "void operator shouldn't be used on void | undefined; it should convey that a return value is being ignored",
						Line:      2, Column: 1, EndLine: 2, EndColumn: 20,
					}},
				},
				{
					Code:   "declare const call: (() => void) | undefined;\nvoid call?.();",
					Output: []string{"declare const call: (() => void) | undefined;\ncall?.();"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOperator", Line: 2, Column: 1}},
				},
				{
					Code:   "declare const audio: { pause?: () => void } | null;\nvoid audio?.pause?.();",
					Output: []string{"declare const audio: { pause?: () => void } | null;\naudio?.pause?.();"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOperator"}},
				},
				{
					Code:   "declare function call(): void | undefined;\nvoid call();",
					Output: []string{"declare function call(): void | undefined;\ncall();"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOperator"}},
				},
				{
					Code:   "type Result = void | undefined; declare function call(): Result;\nvoid call();",
					Output: []string{"type Result = void | undefined; declare function call(): Result;\ncall();"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOperator"}},
				},
				{
					Code:   "declare function call(): void | undefined | never;\nvoid call();",
					Output: []string{"declare function call(): void | undefined | never;\ncall();"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOperator"}},
				},
				{
					Code:   "declare const call: (() => never) | undefined;\nvoid call?.();",
					Output: []string{"declare const call: (() => never) | undefined;\ncall?.();"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOperator"}},
				},
			}
			for i := range invalid {
				invalid[i].Options = options
			}
			rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMeaninglessVoidOperatorRule, valid, invalid)
		})
	}
}

func TestNoMeaninglessVoidOperatorEditDemand(t *testing.T) {
	const ruleName = "@typescript-eslint/no-meaningless-void-operator"
	for _, test := range []struct {
		name        string
		declaration string
		expression  string
		operand     string
		typeName    string
		suggestion  bool
	}{
		{"union", "declare const call: (() => void) | undefined;", "void /* remove */ (/* keep */ call?.())", "(/* keep */ call?.())", "void | undefined", false},
		{"never", "declare function fail(): never;", "void // remove\n\t(fail())", "(fail())", "never", true},
	} {
		for _, suppression := range []struct {
			name   string
			prefix string
		}{
			{"enabled", ""},
			{"line", "// eslint-disable-next-line " + ruleName + "\n"},
			{"block", "/* eslint-disable " + ruleName + " */\n"},
		} {
			t.Run(test.name+"/"+suppression.name, func(t *testing.T) {
				code := test.declaration + "\n" + suppression.prefix + test.expression + ";"
				helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
				program, file, err := helper.CreateTestProgram(code, "void-edit-demand.ts", "tsconfig.json")
				if err != nil {
					t.Fatal(err)
				}
				typeChecker, release := program.GetTypeChecker(t.Context())
				defer release()
				options := rule_tester.ResolveTestCaseOptions(t, &NoMeaninglessVoidOperatorRule, map[string]any{"checkNever": true})
				for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
					var diagnostics []rule.RuleDiagnostic
					comments := rule.NewCommentStore(file)
					ctx := (rule.RuleContext{
						SourceFile:     file,
						TypeChecker:    typeChecker,
						Comments:       comments,
						DisableManager: rule.NewDisableManager(file, comments),
					}).WithDiagnosticConsumer(ruleName, rule.SeverityWarning, rule.DiagnosticConsumer{
						Demand: demand,
						Report: func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) },
					})
					listener := NoMeaninglessVoidOperatorRule.Run(ctx, options)[ast.KindVoidExpression]
					var visit func(*ast.Node) bool
					visit = func(node *ast.Node) bool {
						if node.Kind == ast.KindVoidExpression {
							listener(node)
						}
						return node.ForEachChild(visit)
					}
					file.AsNode().ForEachChild(visit)
					if suppression.prefix != "" {
						if len(diagnostics) != 0 {
							t.Fatalf("demand %d: suppressed diagnostic was reported", demand)
						}
						continue
					}
					if len(diagnostics) != 1 {
						t.Fatalf("demand %d: got %d diagnostics, want 1", demand, len(diagnostics))
					}
					diagnostic := diagnostics[0]
					start := strings.Index(code, test.expression)
					if diagnostic.Range != core.NewTextRange(start, start+len(test.expression)) ||
						diagnostic.Message.Id != "meaninglessVoidOperator" ||
						diagnostic.Message.Description != "void operator shouldn't be used on "+test.typeName+"; it should convey that a return value is being ignored" ||
						diagnostic.RuleName != ruleName || diagnostic.Severity != rule.SeverityWarning {
						t.Fatalf("demand %d: unexpected diagnostic: %#v", demand, diagnostic)
					}
					wantFixes := []rule.RuleFix{{Range: core.NewTextRange(start, start+strings.Index(test.expression, test.operand))}}
					if !test.suggestion && demand&rule.EditDemandAutofix != 0 {
						if diagnostic.FixesPtr == nil || !reflect.DeepEqual(*diagnostic.FixesPtr, wantFixes) {
							t.Fatalf("demand %d: unexpected fixes: %#v", demand, diagnostic.FixesPtr)
						}
					} else if diagnostic.FixesPtr != nil {
						t.Fatalf("demand %d: unexpected autofix", demand)
					}
					if test.suggestion && demand&rule.EditDemandSuggestion != 0 {
						want := []rule.RuleSuggestion{{Message: rule.RuleMessage{Id: "removeVoid", Description: "Remove 'void'"}, FixesArr: wantFixes}}
						if diagnostic.Suggestions == nil || !reflect.DeepEqual(*diagnostic.Suggestions, want) {
							t.Fatalf("demand %d: unexpected suggestions: %#v", demand, diagnostic.Suggestions)
						}
					} else if diagnostic.Suggestions != nil {
						t.Fatalf("demand %d: unexpected suggestions", demand)
					}
				}
			})
		}
	}
}

func TestNoMeaninglessVoidOperatorFixes(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMeaninglessVoidOperatorRule, []rule_tester.ValidTestCase{
		{Code: "declare function fail(): never; void fail();"},
	}, []rule_tester.InvalidTestCase{
		{
			Code:   "declare function call(): void | undefined;\nvoid /* remove */ (/* keep */ call());",
			Output: []string{"declare function call(): void | undefined;\n(/* keep */ call());"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOperator"}},
		},
		{
			Code:   "declare function call(): void | undefined;\nvoid // remove\n\tcall();",
			Output: []string{"declare function call(): void | undefined;\ncall();"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOperator"}},
		},
		{
			Code:   "declare function call(): void | undefined;\nvoid\u00a0call();",
			Output: []string{"declare function call(): void | undefined;\ncall();"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOperator"}},
		},
		{
			Code:   "declare function call(): void | undefined;\nconst result = () => void (call());",
			Output: []string{"declare function call(): void | undefined;\nconst result = () => (call());"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOperator"}},
		},
		{
			Code:    "declare function fail(): never;\nvoid /* remove */ (fail());",
			Options: map[string]any{"checkNever": true},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "meaninglessVoidOperator",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeVoid",
					Output:    "declare function fail(): never;\n(fail());",
				}},
			}},
		},
	})
}
