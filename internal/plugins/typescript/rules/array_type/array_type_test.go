package array_type

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestArrayTypeRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ArrayTypeRule, []rule_tester.ValidTestCase{
		// Base cases - array option
		{Code: "let a: number[] = [];", Options: map[string]interface{}{"default": "array"}},
		{Code: "let a: (string | number)[] = [];", Options: map[string]interface{}{"default": "array"}},
		{Code: "let a: readonly number[] = [];", Options: map[string]interface{}{"default": "array"}},
		{Code: "let a: readonly (string | number)[] = [];", Options: map[string]interface{}{"default": "array"}},
		{Code: "let a: number[] = [];", Options: map[string]interface{}{"default": "array", "readonly": "array"}},
		{Code: "let a: (string | number)[] = [];", Options: map[string]interface{}{"default": "array", "readonly": "array"}},
		{Code: "let a: readonly number[] = [];", Options: map[string]interface{}{"default": "array", "readonly": "array"}},
		{Code: "let a: readonly (string | number)[] = [];", Options: map[string]interface{}{"default": "array", "readonly": "array"}},
		{Code: "let a: number[] = [];", Options: map[string]interface{}{"default": "array", "readonly": "array-simple"}},
		{Code: "let a: (string | number)[] = [];", Options: map[string]interface{}{"default": "array", "readonly": "array-simple"}},
		{Code: "let a: readonly number[] = [];", Options: map[string]interface{}{"default": "array", "readonly": "array-simple"}},
		{Code: "let a: ReadonlyArray<string | number> = [];", Options: map[string]interface{}{"default": "array", "readonly": "array-simple"}},
		{Code: "let a: number[] = [];", Options: map[string]interface{}{"default": "array", "readonly": "generic"}},
		{Code: "let a: (string | number)[] = [];", Options: map[string]interface{}{"default": "array", "readonly": "generic"}},
		{Code: "let a: ReadonlyArray<number> = [];", Options: map[string]interface{}{"default": "array", "readonly": "generic"}},
		{Code: "let a: ReadonlyArray<string | number> = [];", Options: map[string]interface{}{"default": "array", "readonly": "generic"}},

		// array-simple option
		{Code: "let a: number[] = [];", Options: map[string]interface{}{"default": "array-simple"}},
		{Code: "let a: Array<string | number> = [];", Options: map[string]interface{}{"default": "array-simple"}},
		{Code: "let a: readonly number[] = [];", Options: map[string]interface{}{"default": "array-simple"}},
		{Code: "let a: ReadonlyArray<string | number> = [];", Options: map[string]interface{}{"default": "array-simple"}},
		{Code: "let a: number[] = [];", Options: map[string]interface{}{"default": "array-simple", "readonly": "array"}},
		{Code: "let a: Array<string | number> = [];", Options: map[string]interface{}{"default": "array-simple", "readonly": "array"}},
		{Code: "let a: readonly number[] = [];", Options: map[string]interface{}{"default": "array-simple", "readonly": "array"}},
		{Code: "let a: readonly (string | number)[] = [];", Options: map[string]interface{}{"default": "array-simple", "readonly": "array"}},
		{Code: "let a: number[] = [];", Options: map[string]interface{}{"default": "array-simple", "readonly": "array-simple"}},
		{Code: "let a: Array<string | number> = [];", Options: map[string]interface{}{"default": "array-simple", "readonly": "array-simple"}},
		{Code: "let a: readonly number[] = [];", Options: map[string]interface{}{"default": "array-simple", "readonly": "array-simple"}},
		{Code: "let a: ReadonlyArray<string | number> = [];", Options: map[string]interface{}{"default": "array-simple", "readonly": "array-simple"}},
		{Code: "let a: number[] = [];", Options: map[string]interface{}{"default": "array-simple", "readonly": "generic"}},
		{Code: "let a: Array<string | number> = [];", Options: map[string]interface{}{"default": "array-simple", "readonly": "generic"}},
		{Code: "let a: ReadonlyArray<number> = [];", Options: map[string]interface{}{"default": "array-simple", "readonly": "generic"}},
		{Code: "let a: ReadonlyArray<string | number> = [];", Options: map[string]interface{}{"default": "array-simple", "readonly": "generic"}},

		// generic option
		{Code: "let a: Array<number> = [];", Options: map[string]interface{}{"default": "generic"}},
		{Code: "let a: Array<string | number> = [];", Options: map[string]interface{}{"default": "generic"}},
		{Code: "let a: ReadonlyArray<number> = [];", Options: map[string]interface{}{"default": "generic"}},
		{Code: "let a: ReadonlyArray<string | number> = [];", Options: map[string]interface{}{"default": "generic"}},
		{Code: "let a: Array<number> = [];", Options: map[string]interface{}{"default": "generic", "readonly": "generic"}},
		{Code: "let a: Array<string | number> = [];", Options: map[string]interface{}{"default": "generic", "readonly": "generic"}},
		{Code: "let a: ReadonlyArray<number> = [];", Options: map[string]interface{}{"default": "generic", "readonly": "generic"}},
		{Code: "let a: ReadonlyArray<string | number> = [];", Options: map[string]interface{}{"default": "generic", "readonly": "generic"}},
		{Code: "let a: Array<number> = [];", Options: map[string]interface{}{"default": "generic", "readonly": "array"}},
		{Code: "let a: Array<string | number> = [];", Options: map[string]interface{}{"default": "generic", "readonly": "array"}},
		{Code: "let a: readonly number[] = [];", Options: map[string]interface{}{"default": "generic", "readonly": "array"}},
		{Code: "let a: readonly (string | number)[] = [];", Options: map[string]interface{}{"default": "generic", "readonly": "array"}},
		{Code: "let a: Array<number> = [];", Options: map[string]interface{}{"default": "generic", "readonly": "array-simple"}},
		{Code: "let a: Array<string | number> = [];", Options: map[string]interface{}{"default": "generic", "readonly": "array-simple"}},
		{Code: "let a: readonly number[] = [];", Options: map[string]interface{}{"default": "generic", "readonly": "array-simple"}},
		{Code: "let a: ReadonlyArray<string | number> = [];", Options: map[string]interface{}{"default": "generic", "readonly": "array-simple"}},

		// BigInt support
		{Code: "let a: Array<bigint> = [];", Options: map[string]interface{}{"default": "generic", "readonly": "array"}},
		{Code: "let a: readonly bigint[] = [];", Options: map[string]interface{}{"default": "generic", "readonly": "array"}},
		{Code: "let a: readonly (string | bigint)[] = [];", Options: map[string]interface{}{"default": "generic", "readonly": "array"}},
		{Code: "let a: Array<bigint> = [];", Options: map[string]interface{}{"default": "generic", "readonly": "array-simple"}},
		{Code: "let a: Array<string | bigint> = [];", Options: map[string]interface{}{"default": "generic", "readonly": "array-simple"}},
		{Code: "let a: readonly bigint[] = [];", Options: map[string]interface{}{"default": "generic", "readonly": "array-simple"}},
		{Code: "let a: ReadonlyArray<string | bigint> = [];", Options: map[string]interface{}{"default": "generic", "readonly": "array-simple"}},

		// Other valid cases
		{Code: "let a = new Array();", Options: map[string]interface{}{"default": "array"}},
		{Code: "let a: { foo: Bar[] }[] = [];", Options: map[string]interface{}{"default": "array"}},
		{Code: "function foo(a: Array<Bar>): Array<Bar> {}", Options: map[string]interface{}{"default": "generic"}},
		{Code: "let yy: number[][] = [[4, 5], [6]];", Options: map[string]interface{}{"default": "array-simple"}},
		{Code: `
function fooFunction(foo: Array<ArrayClass<string>>) {
  return foo.map(e => e.foo);
}
		`, Options: map[string]interface{}{"default": "array-simple"}},
		{Code: `
function bazFunction(baz: Arr<ArrayClass<String>>) {
  return baz.map(e => e.baz);
}
		`, Options: map[string]interface{}{"default": "array-simple"}},
		{Code: "let fooVar: Array<(c: number) => number>;", Options: map[string]interface{}{"default": "array-simple"}},
		{Code: "type fooUnion = Array<string | number | boolean>;", Options: map[string]interface{}{"default": "array-simple"}},
		{Code: "type fooIntersection = Array<string & number>;", Options: map[string]interface{}{"default": "array-simple"}},
		{Code: `
namespace fooName {
  type BarType = { bar: string };
  type BazType<T> = Arr<T>;
}
		`, Options: map[string]interface{}{"default": "array-simple"}},
		{Code: `
interface FooInterface {
  '.bar': { baz: string[] };
}
		`, Options: map[string]interface{}{"default": "array-simple"}},

		// nested readonly
		{Code: "let a: ReadonlyArray<number[]> = [[]];", Options: map[string]interface{}{"default": "array", "readonly": "generic"}},
		{Code: "let a: readonly Array<number>[] = [[]];", Options: map[string]interface{}{"default": "generic", "readonly": "array"}},
		{Code: "let a: Readonly = [];", Options: map[string]interface{}{"default": "generic", "readonly": "array"}},
		{Code: "const x: Readonly<string> = 'a';", Options: map[string]interface{}{"default": "array"}},
	}, []rule_tester.InvalidTestCase{
		// Base cases - errors with array option
		{
			Code:    "let a: Array<number> = [];",
			Options: map[string]interface{}{"default": "array"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorStringArray",
					Line:      1,
					Column:    8,
				},
			},
			Output: []string{"let a: number[] = [];"},
		},
		{
			Code:    "let a: Array<string | number> = [];",
			Options: map[string]interface{}{"default": "array"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorStringArray",
					Line:      1,
					Column:    8,
				},
			},
			Output: []string{"let a: (string | number)[] = [];"},
		},
		{
			Code:    "let a: ReadonlyArray<number> = [];",
			Options: map[string]interface{}{"default": "array"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorStringArray",
					Line:      1,
					Column:    8,
				},
			},
			Output: []string{"let a: readonly number[] = [];"},
		},
		{
			Code:    "let a: ReadonlyArray<string | number> = [];",
			Options: map[string]interface{}{"default": "array"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorStringArray",
					Line:      1,
					Column:    8,
				},
			},
			Output: []string{"let a: readonly (string | number)[] = [];"},
		},

		// array-simple option errors
		{
			Code:    "let a: Array<number> = [];",
			Options: map[string]interface{}{"default": "array-simple"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorStringArraySimple",
					Line:      1,
					Column:    8,
				},
			},
			Output: []string{"let a: number[] = [];"},
		},
		{
			Code:    "let a: (string | number)[] = [];",
			Options: map[string]interface{}{"default": "array-simple"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorStringGenericSimple",
					Line:      1,
					Column:    8,
				},
			},
			Output: []string{"let a: Array<string | number> = [];"},
		},
		{
			Code:    "let a: ReadonlyArray<number> = [];",
			Options: map[string]interface{}{"default": "array-simple"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorStringArraySimple",
					Line:      1,
					Column:    8,
				},
			},
			Output: []string{"let a: readonly number[] = [];"},
		},
		{
			Code:    "let a: readonly (string | number)[] = [];",
			Options: map[string]interface{}{"default": "array-simple"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorStringGenericSimple",
					Line:      1,
					Column:    8,
				},
			},
			Output: []string{"let a: ReadonlyArray<string | number> = [];"},
		},

		// generic option errors
		{
			Code:    "let a: number[] = [];",
			Options: map[string]interface{}{"default": "generic"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorStringGeneric",
					Line:      1,
					Column:    8,
				},
			},
			Output: []string{"let a: Array<number> = [];"},
		},
		{
			Code:    "let a: (string | number)[] = [];",
			Options: map[string]interface{}{"default": "generic"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorStringGeneric",
					Line:      1,
					Column:    8,
				},
			},
			Output: []string{"let a: Array<string | number> = [];"},
		},
		{
			Code:    "let a: readonly number[] = [];",
			Options: map[string]interface{}{"default": "generic"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorStringGeneric",
					Line:      1,
					Column:    8,
				},
			},
			Output: []string{"let a: ReadonlyArray<number> = [];"},
		},
		{
			Code:    "let a: readonly (string | number)[] = [];",
			Options: map[string]interface{}{"default": "generic"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorStringGeneric",
					Line:      1,
					Column:    8,
				},
			},
			Output: []string{"let a: ReadonlyArray<string | number> = [];"},
		},

		// Complex cases
		{
			Code:    "let a: { foo: Array<Bar> }[] = [];",
			Options: map[string]interface{}{"default": "array"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorStringArray",
					Line:      1,
					Column:    15,
				},
			},
			Output: []string{"let a: { foo: Bar[] }[] = [];"},
		},
		{
			Code:    "let a: Array<{ foo: Bar[] }> = [];",
			Options: map[string]interface{}{"default": "generic"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorStringGeneric",
					Line:      1,
					Column:    21,
				},
			},
			Output: []string{"let a: Array<{ foo: Array<Bar> }> = [];"},
		},
		{
			Code:    "function foo(a: Array<Bar>): Array<Bar> {}",
			Options: map[string]interface{}{"default": "array"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorStringArray",
					Line:      1,
					Column:    17,
				},
				{
					MessageId: "errorStringArray",
					Line:      1,
					Column:    30,
				},
			},
			Output: []string{"function foo(a: Bar[]): Bar[] {}"},
		},

		// Empty arrays
		{
			Code:    "let z: Array = [3, '4'];",
			Options: map[string]interface{}{"default": "array"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorStringArray",
					Line:      1,
					Column:    8,
				},
			},
			Output: []string{"let z: any[] = [3, '4'];"},
		},
		{
			Code:    "let z: Array<> = [3, '4'];",
			Options: map[string]interface{}{"default": "array"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorStringArray",
					Line:      1,
					Column:    8,
				},
			},
			Output: []string{"let z: any[] = [3, '4'];"},
		},

		// BigInt cases
		{
			Code:    "let a: bigint[] = [];",
			Options: map[string]interface{}{"default": "generic", "readonly": "array-simple"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorStringGeneric",
					Line:      1,
					Column:    8,
				},
			},
			Output: []string{"let a: Array<bigint> = [];"},
		},
		{
			Code:    "let a: (string | bigint)[] = [];",
			Options: map[string]interface{}{"default": "generic", "readonly": "array-simple"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorStringGeneric",
					Line:      1,
					Column:    8,
				},
			},
			Output: []string{"let a: Array<string | bigint> = [];"},
		},
		{
			Code:    "let a: ReadonlyArray<bigint> = [];",
			Options: map[string]interface{}{"default": "generic", "readonly": "array-simple"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorStringArraySimple",
					Line:      1,
					Column:    8,
				},
			},
			Output: []string{"let a: readonly bigint[] = [];"},
		},

		// Special readonly cases
		{
			Code:    "const x: Readonly<string[]> = ['a', 'b'];",
			Options: map[string]interface{}{"default": "array"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorStringArrayReadonly",
					Line:      1,
					Column:    10,
				},
			},
			Output: []string{"const x: readonly string[] = ['a', 'b'];"},
		},
		{
			Code:    "declare function foo<E extends Readonly<string[]>>(extra: E): E;",
			Options: map[string]interface{}{"default": "array-simple"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorStringArraySimpleReadonly",
					Line:      1,
					Column:    32,
				},
			},
			Output: []string{"declare function foo<E extends readonly string[]>(extra: E): E;"},
		},

		// Complex template and conditional types
		{
			Code:    "type Conditional<T> = Array<T extends string ? string : number>;",
			Options: map[string]interface{}{"default": "array"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorStringArray",
					Line:      1,
					Column:    23,
				},
			},
			Output: []string{"type Conditional<T> = (T extends string ? string : number)[];"},
		},
		{
			Code:    "type Conditional<T> = (T extends string ? string : number)[];",
			Options: map[string]interface{}{"default": "array-simple"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorStringGenericSimple",
					Line:      1,
					Column:    23,
				},
			},
			Output: []string{"type Conditional<T> = Array<T extends string ? string : number>;"},
		},
		{
			Code:    "type Conditional<T> = (T extends string ? string : number)[];",
			Options: map[string]interface{}{"default": "generic"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorStringGeneric",
					Line:      1,
					Column:    23,
				},
			},
			Output: []string{"type Conditional<T> = Array<T extends string ? string : number>;"},
		},
	})
}

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
type MissingTypeArgument = Array;
type ReadonlyWrappedArray = Readonly<string[]>;
type NestedReadonlyArray = ReadonlyArray<Value>[];
type NestedReadonlyUnionArray = ReadonlyArray<string | number>[];
type NestedReadonlyWrappedArray = Readonly<string[]>[];`,
			wantFixText: []string{
				"Value[]",
				"(string | number)[]",
				"readonly Value[]",
				"any[]",
				"readonly string[]",
				"(readonly Value[])",
				"(readonly (string | number)[])",
				"(readonly string[])",
			},
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
					Program:      program,
					File:         sourceFile.FileName(),
					HasTypeInfo:  true,
					ExcludePaths: []string{},
					GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
						return []linter.ConfiguredRule{{
							Name:     ArrayTypeRule.Name,
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
