package no_useless_length_check_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_useless_length_check"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUselessLengthCheckEditDemand(t *testing.T) {
	for _, source := range []string{
		"array.length > 0 && array.some(Boolean)",
		"array.some(value => value) && array.length > 0",
		"array.length === 0 || array.every(createPredicate())",
		"array.some(predicate) && array.length > 0",
		"if (array.length > 0) for (const element of array) use(element)",
		"if (array.length > 0) { /* keep */ for (const element of array) use(element) }",
		"array.length > 0 && /* keep */ array.some(createPredicate())",
	} {
		t.Run(source, func(t *testing.T) {
			program, file, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(source, "demand.js", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic)
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				var got []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: file.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{Name: no_useless_length_check.NoUselessLengthCheckRule.Name, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return no_useless_length_check.NoUselessLengthCheckRule.Run(ctx, nil)
							},
						}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { got = append(got, d) }},
				})
				if len(got) != 1 {
					t.Fatalf("demand %d: expected one diagnostic, got %d", demand, len(got))
				}
				diagnostics[demand] = got[0]
			}
			all := diagnostics[rule.EditDemandAll]
			for demand, diagnostic := range diagnostics {
				if diagnostic.Range != all.Range || !reflect.DeepEqual(diagnostic.Message, all.Message) || diagnostic.Severity != all.Severity {
					t.Errorf("demand %d changed diagnostic identity", demand)
				}
				if demand&rule.EditDemandAutofix != 0 {
					if !reflect.DeepEqual(diagnostic.FixesPtr, all.FixesPtr) {
						t.Errorf("demand %d changed autofix artifacts", demand)
					}
				} else if diagnostic.FixesPtr != nil {
					t.Errorf("demand %d produced an autofix without demand", demand)
				}
				if demand&rule.EditDemandSuggestion != 0 {
					if !reflect.DeepEqual(diagnostic.Suggestions, all.Suggestions) {
						t.Errorf("demand %d changed suggestion artifacts", demand)
					}
				} else if diagnostic.Suggestions != nil {
					t.Errorf("demand %d produced suggestions without demand", demand)
				}
			}
			if all.Suggestions != nil && (*all.Suggestions)[0].Message.Description != "Remove the length check." {
				t.Errorf("unexpected suggestion message: %q", (*all.Suggestions)[0].Message.Description)
			}
		})
	}
}

// Additional AST and edit cases checked against eslint-plugin-unicorn v79.0.0.

func TestNoUselessLengthCheckExtraSyntax(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_length_check.NoUselessLengthCheckRule, []rule_tester.ValidTestCase{
		{Code: "object.array.length > 0 && object?.array.some(Boolean)", FileName: "case.js"},
		{Code: "object.array.length > 0 && (object?.array.some)(Boolean)", FileName: "case.js"},
		{Code: "array.length > 0 && (array.some)?.(Boolean)", FileName: "case.js"},
		{Code: "array.length > 0 && (array.some!)(Boolean)", FileName: "case.ts"},
		{Code: "class C { #length = 0; f() { return this.#length === 0 || this.every(Boolean); } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array[\"length\"] === 0 || array.every(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length === 0 || array[\"every\"](Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "(array.length as number) > 0 && array.some(Boolean)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length > 0 && (array.some as any)(Boolean)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "(array.length > 0 as boolean) && array.some(Boolean)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length > 0 && (array?.some)(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "items?.array.length > 0 && items.array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "((foo && array.length > 0) as boolean) && array.some(Boolean)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "\"array\".length > 0 && `array`.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
	}, []rule_tester.InvalidTestCase{
		// Parentheses end the optional chain before the ordinary method call.
		{Code: "object.array.length > 0 && (object?.array).some(Boolean)", FileName: "case.js", Output: []string{"(object?.array).some(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 14, EndLine: 1, EndColumn: 24},
		}},
		{Code: "if (object.array.length > 0) for (const x of object?.array) {}", FileName: "case.js", Output: []string{"for (const x of object?.array) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 18, EndLine: 1, EndColumn: 28},
		}},
		{Code: "const item = <div/>; array.length > 0 && array.some(Boolean)", FileName: "case.tsx", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const item = <div/>; array.some(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 28, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class C { #items = []; f() { return this.#items.length > 0 && this.#items.some(Boolean); } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class C { #items = []; f() { return this.#items.some(Boolean); } }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 49, EndLine: 1, EndColumn: 59, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "object[\"arr\" + \"ay\"].length > 0 && object.array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"object.array.some(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 22, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "object[key].length === 0 || object[key].every(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"object[key].every(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 1, Column: 13, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(/** @type {Array<number>} */ (array)).length > 0 && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 40, EndLine: 1, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const astral = \"𐐀\"; if (array.length > 0) { for (const x of array) use(x); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const astral = \"𐐀\"; for (const x of array) use(x);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 32, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "array.length > 0\n&& array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"array.some(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 7, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}

func TestNoUselessLengthCheckExtraReceivers(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_length_check.NoUselessLengthCheckRule, []rule_tester.ValidTestCase{
		{Code: "const array = new Set(); array.length > 0 && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const array = {}; if (array.length > 0) for (const x of array) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
	}, []rule_tester.InvalidTestCase{
		{Code: "const array = new Int8Array(); array.length > 0 && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const array = new Int8Array(); array.some(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 38, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function f(array: number[]) { return array.length > 0 && array.some(Boolean) }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"function f(array: number[]) { return array.some(Boolean) }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 44, EndLine: 1, EndColumn: 54, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function f(array: Set<number> | number[]) { return array.length > 0 && array.some(Boolean) }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"function f(array: Set<number> | number[]) { return array.some(Boolean) }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 58, EndLine: 1, EndColumn: 68, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}

func TestNoUselessLengthCheckExtraLoops(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_length_check.NoUselessLengthCheckRule, []rule_tester.ValidTestCase{
		{Code: "if (array.length > 0) { { for (const x of array) {} } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (array.length > 0) { ; for (const x of array) {} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (array.length > 0) for (let array of array) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (array.length > 0) for (let [array] of array) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (object.array.length > 0) for (const {value: object = other} of object.array) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (array.length > 0) for (const {array} of array) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
	}, []rule_tester.InvalidTestCase{
		{Code: "if (array.length > 0) for (const {array: x} of array) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"for (const {array: x} of array) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (array.length > 0) for (var array of array) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"for (var array of array) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (array.length > 0) for (array of array) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"for (array of array) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (array.length > 0) { for (const x of array) { use(\"a\\\nb\"); } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (array.length > 0) { for (const x of array) use(x); /* keep */ }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (array.length > 0) { for (const x of array) {\n /* keep\n */\n use(x);\n} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (array.length > 0) { \tfor (const x of array) { \t\tuse(x); \t} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"for (const x of array) { \tuse(x); }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}

func TestNoUselessLengthCheckExtraEdits(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_length_check.NoUselessLengthCheckRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{Code: "array.length > 0 && array.some()", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"array.some()"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 7, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "array.some() && array.length > 0", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 23, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "remove-length-check", Output: "array.some()"}}},
		}},
		{Code: "array.length > 0 && array.some(...callbacks)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"array.some(...callbacks)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 7, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "array.some(value => value, ...args) && array.length > 0", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"array.some(value => value, ...args)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 46, EndLine: 1, EndColumn: 56, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "array.length > 0 && array.some(predicate, makeThis())", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 7, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "remove-length-check", Output: "array.some(predicate, makeThis())"}}},
		}},
		{Code: "array.length > 0 && array.some(++predicate)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 7, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "remove-length-check", Output: "array.some(++predicate)"}}},
		}},
		// A tag executes even for an empty array after removing the guard.
		// Keep the shared evaluator's conservative suggestion (documented difference).
		{Code: "array.length > 0 && array.some(tag`text`)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 7, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "remove-length-check", Output: "array.some(tag`text`)"}}},
		}},
		{Code: "array.some(function (value) { return value > 0; }) && array.length > 0", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"array.some(function (value) { return value > 0; })"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 61, EndLine: 1, EndColumn: 71, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "array.some(({value = update()}) => value) && array.length > 0", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 52, EndLine: 1, EndColumn: 62, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "remove-length-check", Output: "array.some(({value = update()}) => value)"}}},
		}},
		{Code: "array.some(value => [...value]) && array.length > 0", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 42, EndLine: 1, EndColumn: 52, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "remove-length-check", Output: "array.some(value => [...value])"}}},
		}},
		{Code: "array.some(value => ({...value})) && array.length > 0", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 44, EndLine: 1, EndColumn: 54, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "remove-length-check", Output: "array.some(value => ({...value}))"}}},
		}},
		{Code: "array.some(value => tag`text`) && array.length > 0", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 41, EndLine: 1, EndColumn: 51, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "remove-length-check", Output: "array.some(value => tag`text`)"}}},
		}},
		{Code: "array.some(value => () => [...value]) && array.length > 0", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 48, EndLine: 1, EndColumn: 58, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "remove-length-check", Output: "array.some(value => () => [...value])"}}},
		}},
		{Code: "array.some((value => value) as any) && array.length > 0", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 46, EndLine: 1, EndColumn: 56, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "remove-length-check", Output: "array.some((value => value) as any)"}}},
		}},
		{Code: "array.some(value => value++) && array.length > 0", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 39, EndLine: 1, EndColumn: 49, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "remove-length-check", Output: "array.some(value => value++)"}}},
		}},
		{Code: "array.some(async value => await value) && array.length > 0", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 49, EndLine: 1, EndColumn: 59, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "remove-length-check", Output: "array.some(async value => await value)"}}},
		}},
		{Code: "array.some(value => { function nested() { use(value); } return value; }) && array.length > 0", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 83, EndLine: 1, EndColumn: 93, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "remove-length-check", Output: "array.some(value => { function nested() { use(value); } return value; })"}}},
		}},
		{Code: "const previous = function () {}\narray.length > 0 && (array).some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const previous = function () {}\n;(array).some(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 2, Column: 7, EndLine: 2, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "array.length > 0 /* keep */ && array.some(create())", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 7, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "array.every(value => value) || /* keep */ array.length === 0", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 1, Column: 49, EndLine: 1, EndColumn: 61, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "array.length > 0 && array.some(value => value) && array.length > 0", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"array.some(value => value)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 7, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 57, EndLine: 1, EndColumn: 67, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "array.length === 0 || array.every(value => value) || array.length === 0", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"array.every(value => value)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 1, Column: 7, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 1, Column: 60, EndLine: 1, EndColumn: 72, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}

// Additional representation and source-text cases checked against Unicorn v79.0.0.
func TestNoUselessLengthCheckRepresentationAndSourceText(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_length_check.NoUselessLengthCheckRule, []rule_tester.ValidTestCase{
		{Code: "object.array.length > 0 && object?.array!.some(Boolean)", FileName: "case.ts"},
		{Code: "if (array.length > 0) for (const [{value: array}] of array) {}", FileName: "case.js"},
	}, []rule_tester.InvalidTestCase{
		{Code: "if (array.length > 0) {\n  for (const x of array) {\n    render(<div>\n      {x}\n    </div>);\n  }\n}", FileName: "case.tsx", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (array.length > 0) {\n  for (const x of array) {\n    render(<div>\n      text\n      more text\n    </div>);\n  }\n}", FileName: "case.tsx", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		// Runtime generic instantiations must remain visible inside nested callbacks.
		{Code: "array.some(value => () => (tag`text`)<T>) && array.length > 0", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 52, EndLine: 1, EndColumn: 62, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "remove-length-check", Output: "array.some(value => () => (tag`text`)<T>)"}}},
		}},
		{Code: "array.some(value => () => (fn)<T>) && array.length > 0", FileName: "case.ts", Output: []string{"array.some(value => () => (fn)<T>)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 45, EndLine: 1, EndColumn: 55, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "array.some(value => <div {...value}/>) && array.length > 0", FileName: "case.tsx", Output: []string{"array.some(value => <div {...value}/>)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 49, EndLine: 1, EndColumn: 59, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "array.length > 0 && array.some(<div {...value}/>)", FileName: "case.tsx", Output: []string{"array.some(<div {...value}/>)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 7, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = <A.B/>; A.length > 0 && A.some(Boolean)", FileName: "case.tsx", Output: []string{"const x = <A.B/>; A.some(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 21, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(/** @type {boolean} */ (array.length > 0)) && array.some(Boolean)", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 32, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "array.some(/** @type {Function} */ (value => value)) && array.length > 0", FileName: "case.js", Output: []string{"array.some(/** @type {Function} */ (value => value))"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 63, EndLine: 1, EndColumn: 73, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (array.length > 0) {\n  for (const x of array) {\n    use(x);\n  }\n}", FileName: "case.js", Output: []string{"for (const x of array) {\n  use(x);\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (array.length > 0) {\r  for (const x of array) {\r    use(x);\r  }\r}", FileName: "case.js", Output: []string{"for (const x of array) {\r  use(x);\r}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const 𐐀 = 1; array.length === 0 || array.every(Boolean)", FileName: "case.js", Output: []string{"const 𐐀 = 1; array.every(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 1, Column: 21, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class C { #array = []; f() { if (this.#array.length > 0) for (const x of this.#array) use(x); } }", FileName: "case.js", Output: []string{"class C { #array = []; f() { for (const x of this.#array) use(x); } }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 46, EndLine: 1, EndColumn: 56, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "array.length > 0 && array.some(value => value) // keep\n&& /* stop */ array.length > 0", FileName: "case.js", Output: []string{"array.some(value => value) // keep\n&& /* stop */ array.length > 0"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 7, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 2, Column: 21, EndLine: 2, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}
