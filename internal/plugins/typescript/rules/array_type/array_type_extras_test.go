package array_type

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestArrayTypeEditDemand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		code        string
		options     []any
		wantFixText []string
	}{
		{
			name: "generic to array",
			code: `type Simple = Array<Value>;
type Union = Array<string | number>;
type ReadonlySimple = ReadonlyArray<Value>;
type ReadonlyWrappedArray = Readonly<string[]>;
type NestedReadonlyArray = ReadonlyArray<Value>[];
type NestedReadonlyUnionArray = ReadonlyArray<string | number>[];
type NestedReadonlyWrappedArray = Readonly<string[]>[];`,
			wantFixText: []string{
				"Value[]",
				"(string | number)[]",
				"readonly Value[]",
				"readonly string[]",
				"(readonly Value[])",
				"(readonly (string | number)[])",
				"(readonly string[])",
			},
		},
		{
			name: "disable directives",
			code: `/* eslint-disable @typescript-eslint/array-type */
type Disabled = Array<string>;
/* eslint-enable @typescript-eslint/array-type */
type Enabled = Array<number>;
// eslint-disable-next-line @typescript-eslint/array-type
type NextLine = ReadonlyArray<string>;
type SameLine = Array<string>; // eslint-disable-line @typescript-eslint/array-type`,
			wantFixText: []string{"number[]"},
		},
		{
			name: "array to generic",
			code: `type Simple = Value[];
type Union = (string | number)[];
type ReadonlySimple = readonly Value[];`,
			options: []any{map[string]any{"default": "generic"}},
			wantFixText: []string{
				"Array<Value>",
				"Array<string | number>",
				"ReadonlyArray<Value>",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			program, sourceFile, err := helper.CreateTestProgram(
				test.code,
				"array-type-edit-demand.ts",
				"tsconfig.json",
			)
			if err != nil {
				t.Fatal(err)
			}

			run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
				t.Helper()

				var diagnostics []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program:     lintprogram.NewFromCompiler(program),
					File:        sourceFile.FileName(),
					HasTypeInfo: true,
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{
							Name:     "@typescript-eslint/array-type",
							Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return ArrayTypeRule.Run(ctx, test.options)
							},
						}}
					},
					Consumer: rule.DiagnosticConsumer{
						Demand: demand,
						Report: func(diagnostic rule.RuleDiagnostic) {
							diagnostics = append(diagnostics, diagnostic)
						},
					},
				})
				if len(diagnostics) != len(test.wantFixText) {
					t.Fatalf(
						"demand %d: diagnostics = %d, want %d",
						demand,
						len(diagnostics),
						len(test.wantFixText),
					)
				}
				return diagnostics
			}

			diagnosticsOnly := run(rule.EditDemandNone)
			autofixOnly := run(rule.EditDemandAutofix)
			suggestionOnly := run(rule.EditDemandSuggestion)
			allEdits := run(rule.EditDemandAll)

			withoutEdits := func(diagnostic rule.RuleDiagnostic) rule.RuleDiagnostic {
				diagnostic.FixesPtr = nil
				diagnostic.Suggestions = nil
				return diagnostic
			}

			for index, wantText := range test.wantFixText {
				wantIdentity := withoutEdits(allEdits[index])
				for demand, diagnostics := range map[rule.EditDemand][]rule.RuleDiagnostic{
					rule.EditDemandNone:       diagnosticsOnly,
					rule.EditDemandAutofix:    autofixOnly,
					rule.EditDemandSuggestion: suggestionOnly,
				} {
					if got := withoutEdits(diagnostics[index]); !reflect.DeepEqual(got, wantIdentity) {
						t.Errorf(
							"demand %d changed diagnostic %d:\ngot  %#v\nwant %#v",
							demand,
							index,
							got,
							wantIdentity,
						)
					}
				}

				if diagnosticsOnly[index].FixesPtr != nil || suggestionOnly[index].FixesPtr != nil {
					t.Fatalf("diagnostic %d: non-autofix demand materialized fixes", index)
				}
				for _, diagnostics := range [][]rule.RuleDiagnostic{
					diagnosticsOnly,
					autofixOnly,
					suggestionOnly,
					allEdits,
				} {
					if diagnostics[index].Suggestions != nil {
						t.Fatalf("diagnostic %d: autofix-only rule materialized suggestions", index)
					}
				}

				for demand, diagnostics := range map[rule.EditDemand][]rule.RuleDiagnostic{
					rule.EditDemandAutofix: autofixOnly,
					rule.EditDemandAll:     allEdits,
				} {
					fixes := diagnostics[index].FixesPtr
					if fixes == nil || len(*fixes) != 1 || (*fixes)[0].Text != wantText {
						t.Fatalf(
							"demand %d diagnostic %d: fixes = %#v, want one fix with text %q",
							demand,
							index,
							fixes,
							wantText,
						)
					}
				}

				if !reflect.DeepEqual(autofixOnly[index].FixesPtr, allEdits[index].FixesPtr) {
					t.Fatalf("diagnostic %d: autofix and all-edits demands produced different fixes", index)
				}
			}
		})
	}
}

func TestArrayTypeHeritage(t *testing.T) {
	var valid []rule_tester.ValidTestCase
	for _, declaration := range []string{"interface I extends", "class C implements", "class C extends"} {
		for _, target := range []string{"Array<string>", "ReadonlyArray<string>", "Readonly<string[]>"} {
			for _, options := range []any{nil, map[string]any{"default": "array-simple"}, map[string]any{"default": "generic", "readonly": "array"}} {
				// Under generic style the nested string[] still needs a fix.
				if target == "Readonly<string[]>" && options != nil {
					if options.(map[string]any)["default"] == "generic" {
						continue
					}
				}
				valid = append(valid, rule_tester.ValidTestCase{Code: fmt.Sprintf("%s %s {}", declaration, target), Options: options})
			}
		}
	}
	var invalid []rule_tester.InvalidTestCase
	for _, declaration := range []string{"interface I extends", "class C implements", "class C extends"} {
		for _, target := range []string{"Base", "Array"} {
			invalid = append(invalid, rule_tester.InvalidTestCase{
				Code:   fmt.Sprintf("%s %s<Array<string>> {}", declaration, target),
				Output: []string{fmt.Sprintf("%s %s<string[]> {}", declaration, target)},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "errorStringArray"}},
			})
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ArrayTypeRule, valid, invalid)
}

// Expected messages, ranges and fixes were checked against typescript-eslint v8.70.1.
func TestArrayTypeSyntaxAndScopes(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ArrayTypeRule, []rule_tester.ValidTestCase{
		{Code: "type T = (( null ))[];", Options: []any{map[string]any{"default": "array-simple"}}},
		{Code: "type T = (( string ))[];", Options: []any{map[string]any{"default": "array-simple"}}},
		{Code: "type T = readonly ((string)[]);", Options: []any{map[string]any{"default": "generic", "readonly": "array"}}},
		{Code: "type T = readonly ((string)[]);", Options: []any{map[string]any{"default": "generic", "readonly": "array-simple"}}},
		{Code: "type X = ReadonlyArray;", Options: []any{map[string]any{}}},
		{Code: "type Array<T> = T; type X = Array<string[]>;", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "interface Array<T> {} type X = Array<string[]>;", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "import type {Array} from \"x\"; type X = Array<string[]>;", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const Array = 0; type X = Array<string[]>;", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "namespace N { type Array<T> = T; type X = Array<string[]>; }", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "function f<Array>() { type X = Array<string[]>; }", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class C<Array> { static x: Array<string>; x: Array<string>; }", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "type X<Array> = Array<string>;", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "type X = < Array >() => Array<string>;", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "{ type X = Array<string[]>; let Array; }", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "function f(Array: unknown) { type X = Array<string[]>; }", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "function f(x: Array<string>) { const Array = 0; }", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "function f(x = 0 as Array<string>) { var Array; }", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "try {} catch (Array) { type X = Array<string[]>; }", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "for (let Array of []) { type X = Array<string[]>; }", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const f = function Array() { type X = Array<string[]>; };", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const C = class Array { field: Array<string> };", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "type X<T> = T extends infer Array ? Array<string> : Array<number>;", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "type X = { [ Array in \"x\" ]: Array<string> };", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "interface C<Array> { x: Array<string> }", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class C { [Array<string>] (Array: any) { type X = Array<string[]>; } }", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class C { m(@dec(null as Array<string>) Array: any) {} }", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class C { m(@dec(null as ReadonlyArray<string>) ReadonlyArray: any) {} }", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class C { m(@dec(null as Readonly<string>) Readonly: any) {} }", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class C { m(@dec(() => null as Readonly<string>) Readonly: any) {} }", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "namespace N { type ReadonlyArray<T> = T; type X = ReadonlyArray<string[]>; }", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "namespace N { type Readonly<T> = T; type X = Readonly<string[]>; }", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
	}, []rule_tester.InvalidTestCase{
		{Code: "type T = Array<((null))>;", Options: []any{map[string]any{"default": "array"}}, Output: []string{"type T = null[];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringArray", Message: "Array type using 'Array<null>' is forbidden. Use 'null[]' instead.", Line: 1, Column: 10, EndLine: 1, EndColumn: 25},
		}},
		{Code: "type T = Array<((string))>;", Options: []any{map[string]any{"default": "array"}}, Output: []string{"type T = string[];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringArray", Message: "Array type using 'Array<string>' is forbidden. Use 'string[]' instead.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "type T = (( Array<string> ))[];", Options: []any{map[string]any{"default": "array-simple"}}, Output: []string{"type T = (( string[] ))[];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringArraySimple", Message: "Array type using 'Array<string>' is forbidden for simple types. Use 'string[]' instead.", Line: 1, Column: 13, EndLine: 1, EndColumn: 26},
		}},
		{Code: "type T = Array<((Array<string>))>;", Options: []any{map[string]any{"default": "array"}}, Output: []string{"type T = Array<string>[];", "type T = string[][];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringArray", Message: "Array type using 'Array<Array<string>>' is forbidden. Use 'Array<string>[]' instead.", Line: 1, Column: 10, EndLine: 1, EndColumn: 34},
			{MessageId: "errorStringArray", Message: "Array type using 'Array<string>' is forbidden. Use 'string[]' instead.", Line: 1, Column: 18, EndLine: 1, EndColumn: 31},
		}},
		{Code: "type T = (( readonly string[] ))[];", Options: []any{map[string]any{"default": "array-simple"}}, Output: []string{"type T = Array<readonly string[]>;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringGenericSimple", Message: "Array type using 'T[]' is forbidden for non-simple types. Use 'Array<T>' instead.", Line: 1, Column: 10, EndLine: 1, EndColumn: 35},
		}},
		{Code: "type T = Array<((readonly string[]))>;", Options: []any{map[string]any{"default": "array"}}, Output: []string{"type T = (readonly string[])[];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringArray", Message: "Array type using 'Array<T>' is forbidden. Use 'T[]' instead.", Line: 1, Column: 10, EndLine: 1, EndColumn: 38},
		}},
		{Code: "type T = (( string | number ))[];", Options: []any{map[string]any{"default": "array-simple"}}, Output: []string{"type T = Array<string | number>;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringGenericSimple", Message: "Array type using 'T[]' is forbidden for non-simple types. Use 'Array<T>' instead.", Line: 1, Column: 10, EndLine: 1, EndColumn: 33},
		}},
		{Code: "type T = Array<((string | number))>;", Options: []any{map[string]any{"default": "array"}}, Output: []string{"type T = (string | number)[];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringArray", Message: "Array type using 'Array<T>' is forbidden. Use 'T[]' instead.", Line: 1, Column: 10, EndLine: 1, EndColumn: 36},
		}},
		{Code: "type T = (( () => void ))[];", Options: []any{map[string]any{"default": "array-simple"}}, Output: []string{"type T = Array<() => void>;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringGenericSimple", Message: "Array type using 'T[]' is forbidden for non-simple types. Use 'Array<T>' instead.", Line: 1, Column: 10, EndLine: 1, EndColumn: 28},
		}},
		{Code: "type T = Array<((() => void))>;", Options: []any{map[string]any{"default": "array"}}, Output: []string{"type T = (() => void)[];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringArray", Message: "Array type using 'Array<T>' is forbidden. Use 'T[]' instead.", Line: 1, Column: 10, EndLine: 1, EndColumn: 31},
		}},
		{Code: "type T = (( T extends U ? X : Y ))[];", Options: []any{map[string]any{"default": "array-simple"}}, Output: []string{"type T = Array<T extends U ? X : Y>;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringGenericSimple", Message: "Array type using 'T[]' is forbidden for non-simple types. Use 'Array<T>' instead.", Line: 1, Column: 10, EndLine: 1, EndColumn: 37},
		}},
		{Code: "type T = Array<((T extends U ? X : Y))>;", Options: []any{map[string]any{"default": "array"}}, Output: []string{"type T = (T extends U ? X : Y)[];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringArray", Message: "Array type using 'Array<T>' is forbidden. Use 'T[]' instead.", Line: 1, Column: 10, EndLine: 1, EndColumn: 40},
		}},
		{Code: "type T = readonly ((string)[]);", Options: []any{map[string]any{"default": "array", "readonly": "generic"}}, Output: []string{"type T = ReadonlyArray<string>;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringGeneric", Message: "Array type using 'readonly string[]' is forbidden. Use 'ReadonlyArray<string>' instead.", Line: 1, Column: 10, EndLine: 1, EndColumn: 31},
		}},
		{Code: "type T = Readonly<((string[]))>;", Options: []any{map[string]any{}}, Output: []string{"type T = readonly string[];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringArrayReadonly", Message: "Array type using 'Readonly<string[]>' is forbidden. Use 'readonly string[]' instead.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "type T = ReadonlyArray<((string | number))>[];", Options: []any{map[string]any{}}, Output: []string{"type T = (readonly (string | number)[])[];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringArray", Message: "Array type using 'ReadonlyArray<T>' is forbidden. Use 'readonly T[]' instead.", Line: 1, Column: 10, EndLine: 1, EndColumn: 44},
		}},
		{Code: "type T = Array</* before */ (string) /* after */>;", Options: []any{map[string]any{}}, Output: []string{"type T = string[];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringArray", Message: "Array type using 'Array<string>' is forbidden. Use 'string[]' instead.", Line: 1, Column: 10, EndLine: 1, EndColumn: 50},
		}},
		// cspell:ignore rray
		{Code: "type X = \\u0041rray<string>;", Options: []any{map[string]any{}}, Output: []string{"type X = string[];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringArray", Message: "Array type using 'Array<string>' is forbidden. Use 'string[]' instead.", Line: 1, Column: 10, EndLine: 1, EndColumn: 28},
		}},
		{Code: "type X = Array</*😀*/ string>;", Options: []any{map[string]any{}}, Output: []string{"type X = string[];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringArray", Message: "Array type using 'Array<string>' is forbidden. Use 'string[]' instead.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "function f(x = 0 as Array<string>) { { let Array; } }", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"function f(x = 0 as string[]) { { let Array; } }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringArray", Message: "Array type using 'Array<string>' is forbidden. Use 'string[]' instead.", Line: 1, Column: 21, EndLine: 1, EndColumn: 34},
		}},
		{Code: "declare global { interface Array<T> {} } type X = Array<string[]>;", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"declare global { interface Array<T> {} } type X = string[][];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringArray", Message: "Array type using 'Array<string[]>' is forbidden. Use 'string[][]' instead.", Line: 1, Column: 51, EndLine: 1, EndColumn: 66},
		}},
		{Code: "class C { @dec(null as Array<string>) m<Array>() {} }", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class C { @dec(null as string[]) m<Array>() {} }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringArray", Message: "Array type using 'Array<string>' is forbidden. Use 'string[]' instead.", Line: 1, Column: 24, EndLine: 1, EndColumn: 37},
		}},
		{Code: "class C { m(@dec(() => null as Array<string>) Array: any) {} }", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class C { m(@dec(() => null as string[]) Array: any) {} }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringArray", Message: "Array type using 'Array<string>' is forbidden. Use 'string[]' instead.", Line: 1, Column: 32, EndLine: 1, EndColumn: 45},
		}},
		{Code: "class C { m(@dec(() => null as ReadonlyArray<string>) ReadonlyArray: any) {} }", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class C { m(@dec(() => null as readonly string[]) ReadonlyArray: any) {} }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringArray", Message: "Array type using 'ReadonlyArray<string>' is forbidden. Use 'readonly string[]' instead.", Line: 1, Column: 32, EndLine: 1, EndColumn: 53},
		}},
		{Code: "namespace N { export type Array<T> = T; } type X = Array<string[]>;", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"namespace N { export type Array<T> = T; } type X = string[][];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringArray", Message: "Array type using 'Array<string[]>' is forbidden. Use 'string[][]' instead.", Line: 1, Column: 52, EndLine: 1, EndColumn: 67},
		}},
		{Code: "type Array<T> = T; type X = Array<string[]>;", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Output: []string{"type Array<T> = T; type X = string[][];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringArray", Message: "Array type using 'Array<string[]>' is forbidden. Use 'string[][]' instead.", Line: 1, Column: 29, EndLine: 1, EndColumn: 44},
		}},
		{Code: "type Array<T> = T; type X = Array<string[]>;", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Output: []string{"type Array<T> = T; type X = string[][];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringArray", Message: "Array type using 'Array<string[]>' is forbidden. Use 'string[][]' instead.", Line: 1, Column: 29, EndLine: 1, EndColumn: 44},
		}},
		{Code: "type ReadonlyArray<T> = T; type X = ReadonlyArray<string[]>;", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Output: []string{"type ReadonlyArray<T> = T; type X = readonly string[][];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringArray", Message: "Array type using 'ReadonlyArray<string[]>' is forbidden. Use 'readonly string[][]' instead.", Line: 1, Column: 37, EndLine: 1, EndColumn: 60},
		}},
		{Code: "type ReadonlyArray<T> = T; type X = ReadonlyArray<string[]>;", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Output: []string{"type ReadonlyArray<T> = T; type X = readonly string[][];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringArray", Message: "Array type using 'ReadonlyArray<string[]>' is forbidden. Use 'readonly string[][]' instead.", Line: 1, Column: 37, EndLine: 1, EndColumn: 60},
		}},
		{Code: "type Readonly<T> = T; type X = Readonly<string[]>;", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Output: []string{"type Readonly<T> = T; type X = readonly string[];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringArrayReadonly", Message: "Array type using 'Readonly<string[]>' is forbidden. Use 'readonly string[]' instead.", Line: 1, Column: 32, EndLine: 1, EndColumn: 50},
		}},
		{Code: "type Readonly<T> = T; type X = Readonly<string[]>;", Options: []any{map[string]any{}}, LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Output: []string{"type Readonly<T> = T; type X = readonly string[];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "errorStringArrayReadonly", Message: "Array type using 'Readonly<string[]>' is forbidden. Use 'readonly string[]' instead.", Line: 1, Column: 32, EndLine: 1, EndColumn: 50},
		}},
	})
}

func TestArrayTypeOptions(t *testing.T) {
	for _, option := range []string{"array", "array-simple", "generic"} {
		for _, readonly := range []string{"array", "array-simple", "generic"} {
			if err := ArrayTypeRule.Schema.Validate([]any{map[string]any{"default": option, "readonly": readonly}}); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, options := range [][]any{nil, {map[string]any{}}, {map[string]any{"readonly": "generic"}}} {
		if err := ArrayTypeRule.Schema.Validate(options); err != nil {
			t.Fatal(err)
		}
	}
	for _, options := range [][]any{{map[string]any{"default": nil}}, {map[string]any{"readonly": "other"}}, {map[string]any{"extra": true}}, {map[string]any{}, map[string]any{}}} {
		if err := ArrayTypeRule.Schema.Validate(options); err == nil {
			t.Errorf("expected invalid options: %#v", options)
		}
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ArrayTypeRule, []rule_tester.ValidTestCase{
		{Code: "type X = ReadonlyArray<string>;", Options: map[string]any{"readonly": "generic"}},
	}, []rule_tester.InvalidTestCase{
		{Code: "type X = Array<string>; type Y = ReadonlyArray<string>;", Options: map[string]any{"readonly": "generic"}, Output: []string{"type X = string[]; type Y = ReadonlyArray<string>;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "errorStringArray"}}},
	})
}
