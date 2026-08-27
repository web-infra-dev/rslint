package consistent_type_definitions

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentTypeDefinitionsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConsistentTypeDefinitionsRule, []rule_tester.ValidTestCase{
		// Default options (style: 'interface')
		{Code: `var foo = {};`},
		{Code: `interface A {}`},
		{Code: `interface A { x: number; }`},
		{Code: `interface A extends B { x: number; }`},
		{Code: `type U = string;`},
		{Code: `type V = { x: number } | { y: string };`},
		{Code: `type V = { x: number } & { y: string };`},
		{Code: `type Record<T, U> = { [K in T]: U };`},
		{Code: `type T = string | number;`},
		{Code: `type T = () => void;`},
		{Code: `type T = new () => void;`},
		{Code: `type T = [number, string];`},
		{Code: `type T = number[];`},
		{Code: `type T = readonly number[];`},
		{Code: `type T<U> = U & { x: number };`},

		// style: 'type'
		{Code: `type T = { x: number; };`, Options: []interface{}{"type"}},
		{Code: `type T = { x: number };`, Options: []interface{}{"type"}},
		{Code: `type T = { x: number; y: string; };`, Options: []interface{}{"type"}},
		{Code: `type A = { x: number } & B & C;`, Options: []interface{}{"type"}},
		{Code: `type A = { x: number } & B<T1> & C<T2>;`, Options: []interface{}{"type"}},
		{Code: `export type W<T> = { x: T };`, Options: []interface{}{"type"}},
		{Code: `export type W<T> = { x: T; y: U; };`, Options: []interface{}{"type"}},
		{Code: `type U = string;`, Options: []interface{}{"type"}},
		{Code: `type V = { x: number } | { y: string };`, Options: []interface{}{"type"}},
		{Code: `type Record<T, U> = { [K in T]: U };`, Options: []interface{}{"type"}},
	}, []rule_tester.InvalidTestCase{
		// Default options (style: 'interface') - expect type to be interface
		{
			Code:   `type T = { [K: string]: number };`,
			Output: []string{`interface T { [K: string]: number }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "interfaceOverType",
					Message:   "Use an `interface` instead of a `type`.",
					Line:      1,
					Column:    6,
					EndLine:   1,
					EndColumn: 7,
				},
			},
		},
		{
			Code:   `type T = { x: number; };`,
			Output: []string{`interface T { x: number; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code:   `type T={ x: number; };`,
			Output: []string{`interface T { x: number; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code:   `type T= { x: number; };`,
			Output: []string{`interface T { x: number; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code:   `type T = { x: number };`,
			Output: []string{`interface T { x: number }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code:   `type T = { x: number; y: string; };`,
			Output: []string{`interface T { x: number; y: string; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code:   `type T = { x: number; y: { z: string; }; };`,
			Output: []string{`interface T { x: number; y: { z: string; }; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code:   `export type W<T> = { x: T; };`,
			Output: []string{`export interface W<T> { x: T; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code:   `type T<U> = { x: U; };`,
			Output: []string{`interface T<U> { x: U; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code:   `type Foo = { a: string; };`,
			Output: []string{`interface Foo { a: string; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code:   `type Foo = ({ a: string; });`,
			Output: []string{`interface Foo { a: string; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code:   `type Foo = (  { a: string; });`,
			Output: []string{`interface Foo { a: string; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		// type → interface with comment
		{
			Code:   `type T /* comment */={ x: number; };`,
			Output: []string{`interface T /* comment */ { x: number; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code:   `type /* before-name */ T = { x: number };`,
			Output: []string{`interface /* before-name */ T { x: number }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code:   `export /* before-type */ type T = { x: number };`,
			Output: []string{`export /* before-type */ interface T { x: number }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code:   `type T /* first */ /* second */ = { x: number };`,
			Output: []string{`interface T /* first */ /* second */ { x: number }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code:   `type T = /* after-equals */ { x: number };`,
			Output: []string{`interface T { x: number }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		{
			Code:   `type T</* parameter */ U> /* before-equals */ = (/* after-equals */ { x: U });`,
			Output: []string{`interface T</* parameter */ U> /* before-equals */ { x: U }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		// type → interface with excessive whitespace
		{
			Code:   `type T=                         { x: number; };`,
			Output: []string{`interface T { x: number; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		// no closing semicolon
		{
			Code:   "type Foo = {\n  a: string;\n}",
			Output: []string{"interface Foo {\n  a: string;\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		// no closing semicolon; ensure we don't erase subsequent code.
		{
			Code:   "type Foo = {\n  a: string;\n}\ntype Bar = string;",
			Output: []string{"interface Foo {\n  a: string;\n}\ntype Bar = string;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		// Parenthesized type - multiple layers
		{
			Code:   `type Foo = ((((((((({ a: string; })))))))));`,
			Output: []string{`interface Foo { a: string; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		// no closing semicolon with parenthesized type
		{
			Code:   "type Foo = ((({ a: string; })))\n\nconst bar = 1;",
			Output: []string{"interface Foo { a: string; }\n\nconst bar = 1;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},
		// export declare type
		{
			Code:   "export declare type Test = {\n  foo: string;\n  bar: string;\n};",
			Output: []string{"export declare interface Test {\n  foo: string;\n  bar: string;\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "interfaceOverType"},
			},
		},

		// style: 'type' - expect interface to be type
		{
			Code:    `interface T { x: number; }`,
			Options: []interface{}{"type"},
			Output:  []string{`type T = { x: number; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "typeOverInterface",
					Message:   "Use a `type` instead of an `interface`.",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 12,
				},
			},
		},
		{
			Code:    `interface T { x: number }`,
			Options: []interface{}{"type"},
			Output:  []string{`type T = { x: number }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `interface T { x: number; y: string; }`,
			Options: []interface{}{"type"},
			Output:  []string{`type T = { x: number; y: string; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `interface A extends B, C { x: number; };`,
			Options: []interface{}{"type"},
			Output:  []string{`type A = { x: number; } & B & C;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `interface A extends B<T1>, C<T2> { x: number; };`,
			Options: []interface{}{"type"},
			Output:  []string{`type A = { x: number; } & B<T1> & C<T2>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `export interface W<T> { x: T; };`,
			Options: []interface{}{"type"},
			Output:  []string{`export type W<T> = { x: T; };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `interface T<U> { x: U; };`,
			Options: []interface{}{"type"},
			Output:  []string{`type T<U> = { x: U; };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `interface Foo { a: string; }`,
			Options: []interface{}{"type"},
			Output:  []string{`type Foo = { a: string; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `namespace Foo { export interface Bar {} }`,
			Options: []interface{}{"type"},
			Output:  []string{`namespace Foo { export type Bar = {} }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `interface /* before-name */ T { x: number }`,
			Options: []interface{}{"type"},
			Output:  []string{`type /* before-name */ T = { x: number }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `export /* before-interface */ interface T { x: number }`,
			Options: []interface{}{"type"},
			Output:  []string{`export /* before-interface */ type T = { x: number }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		// interface → type with excessive whitespace
		{
			Code:    `interface T                          { x: number; }`,
			Options: []interface{}{"type"},
			Output:  []string{`type T = { x: number; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		// interface → type, no space before brace
		{
			Code:    `interface T{ x: number; }`,
			Options: []interface{}{"type"},
			Output:  []string{`type T = { x: number; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		// namespace JSX
		{
			Code:    "namespace JSX {\n  interface Array<T> {\n    foo(x: (x: number) => T): T[];\n  }\n}",
			Options: []interface{}{"type"},
			Output:  []string{"namespace JSX {\n  type Array<T> = {\n    foo(x: (x: number) => T): T[];\n  }\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		// global without declare (should be fixable)
		{
			Code:    "global {\n  interface Array<T> {\n    foo(x: (x: number) => T): T[];\n  }\n}",
			Options: []interface{}{"type"},
			Output:  []string{"global {\n  type Array<T> = {\n    foo(x: (x: number) => T): T[];\n  }\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `declare namespace global { interface A {} }`,
			Options: []interface{}{"type"},
			Output:  []string{`declare namespace global { type A = {} }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `declare module global { interface A {} }`,
			Options: []interface{}{"type"},
			Output:  []string{`declare module global { type A = {} }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `declare namespace Outer { global { interface A {} } }`,
			Options: []interface{}{"type"},
			Output:  []string{`declare namespace Outer { global { type A = {} } }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		// export default interface
		{
			Code:    "export default interface Test {\n  bar(): string;\n  foo(): number;\n}",
			Options: []interface{}{"type"},
			Output:  []string{"type Test = {\n  bar(): string;\n  foo(): number;\n}\nexport default Test"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `export default interface Test extends Base<T>, Extra { x: T };`,
			Options: []interface{}{"type"},
			Output:  []string{"type Test = { x: T } & Base<T> & Extra\nexport default Test;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		// export declare interface
		{
			Code:    "export declare interface Test {\n  foo: string;\n  bar: string;\n}",
			Options: []interface{}{"type"},
			Output:  []string{"export declare type Test = {\n  foo: string;\n  bar: string;\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},

		// Global module cases - declare global: report but no fix
		{
			Code:    `declare global { interface Array<T> { foo(): void; } }`,
			Options: []interface{}{"type"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
		{
			Code:    `declare global { namespace Foo { interface Bar {} } }`,
			Options: []interface{}{"type"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverInterface"},
			},
		},
	})
}

func TestConsistentTypeDefinitionsEditDemand(t *testing.T) {
	testCases := []struct {
		name      string
		code      string
		options   []any
		wantFixes []bool
	}{
		{
			name:      "type to interface",
			code:      `export type Shape<T> = { value: T; nested: { id: number } };`,
			wantFixes: []bool{true},
		},
		{
			name: "interface to type and declare global",
			code: `export {};
interface Shape<T> extends Base<T>, Extra { value: T; };
declare global { interface Array<T> { item: T; } }`,
			options:   []any{"type"},
			wantFixes: []bool{true, false},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			program, sourceFile, err := helper.CreateTestProgram(
				testCase.code,
				"edit-demand.ts",
				"tsconfig.json",
			)
			if err != nil {
				t.Fatal(err)
			}

			run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
				t.Helper()

				var diagnostics []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program),
					File:    sourceFile.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{
							Name:     ConsistentTypeDefinitionsRule.Name,
							Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return ConsistentTypeDefinitionsRule.Run(ctx, testCase.options)
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
				if len(diagnostics) != len(testCase.wantFixes) {
					t.Fatalf(
						"demand %d: diagnostics = %d, want %d",
						demand,
						len(diagnostics),
						len(testCase.wantFixes),
					)
				}
				return diagnostics
			}

			diagnostics := map[rule.EditDemand][]rule.RuleDiagnostic{
				rule.EditDemandNone:       run(rule.EditDemandNone),
				rule.EditDemandAutofix:    run(rule.EditDemandAutofix),
				rule.EditDemandSuggestion: run(rule.EditDemandSuggestion),
				rule.EditDemandAll:        run(rule.EditDemandAll),
			}

			withoutEdits := func(diagnostic rule.RuleDiagnostic) rule.RuleDiagnostic {
				diagnostic.FixesPtr = nil
				diagnostic.Suggestions = nil
				return diagnostic
			}
			for index, allEdits := range diagnostics[rule.EditDemandAll] {
				for demand, demandDiagnostics := range diagnostics {
					if got, want := withoutEdits(demandDiagnostics[index]), withoutEdits(allEdits); !reflect.DeepEqual(got, want) {
						t.Errorf(
							"diagnostic %d demand %d changed identity:\ngot  %#v\nwant %#v",
							index,
							demand,
							got,
							want,
						)
					}
					if demandDiagnostics[index].Suggestions != nil {
						t.Errorf("diagnostic %d demand %d materialized suggestions", index, demand)
					}
				}

				none := diagnostics[rule.EditDemandNone][index]
				autofix := diagnostics[rule.EditDemandAutofix][index]
				suggestion := diagnostics[rule.EditDemandSuggestion][index]
				if none.FixesPtr != nil || suggestion.FixesPtr != nil {
					t.Errorf("diagnostic %d: non-autofix demand materialized fixes", index)
				}
				if !reflect.DeepEqual(autofix.FixesPtr, allEdits.FixesPtr) {
					t.Errorf("diagnostic %d: autofix and all-edits demands produced different fixes", index)
				}
				if testCase.wantFixes[index] {
					if allEdits.FixesPtr == nil || len(*allEdits.FixesPtr) == 0 {
						t.Errorf("diagnostic %d: all-edits demand produced no fixes", index)
					}
				} else if allEdits.FixesPtr != nil {
					t.Errorf("diagnostic %d: non-fixable report materialized fixes", index)
				}
			}
		})
	}
}

func TestConsistentTypeDefinitionsFixEdits(t *testing.T) {
	testCases := []struct {
		name             string
		code             string
		options          []any
		wantMergedSource string
		wantMergedText   string
		wantOutput       string
	}{
		{
			name: "reported DoActionV2Resp export",
			code: `export type DoActionV2Resp = {
  code?: number;
  msg?: string;
  data?: {
    resp?: string;
    success?: boolean;
  };
};`,
			wantMergedSource: `type DoActionV2Resp = {
  code?: number;
  msg?: string;
  data?: {
    resp?: string;
    success?: boolean;
  };
};`,
			wantMergedText: `interface DoActionV2Resp {
  code?: number;
  msg?: string;
  data?: {
    resp?: string;
    success?: boolean;
  };
}`,
			wantOutput: `export interface DoActionV2Resp {
  code?: number;
  msg?: string;
  data?: {
    resp?: string;
    success?: boolean;
  };
}`,
		},
		{
			name: "reported GetShareDeepResearchPodcastResp export",
			code: `export type GetShareDeepResearchPodcastResp = {
  podcast_gen_status?: PodcastGenStatus;
  audit_status?: AuditStatus;
  episode?: Episode;
};`,
			wantMergedSource: `type GetShareDeepResearchPodcastResp = {
  podcast_gen_status?: PodcastGenStatus;
  audit_status?: AuditStatus;
  episode?: Episode;
};`,
			wantMergedText: `interface GetShareDeepResearchPodcastResp {
  podcast_gen_status?: PodcastGenStatus;
  audit_status?: AuditStatus;
  episode?: Episode;
}`,
			wantOutput: `export interface GetShareDeepResearchPodcastResp {
  podcast_gen_status?: PodcastGenStatus;
  audit_status?: AuditStatus;
  episode?: Episode;
}`,
		},
		{
			name:             "type alias parameters preserve trivia before close",
			code:             `export type T<U /* before close */,> /* before equals */ = ({ x: U });`,
			wantMergedSource: `type T<U /* before close */,> /* before equals */ = ({ x: U });`,
			wantMergedText:   `interface T<U /* before close */,> /* before equals */ { x: U }`,
			wantOutput:       `export interface T<U /* before close */,> /* before equals */ { x: U }`,
		},
		{
			name:             "export interface keeps modifier outside fix",
			code:             `export interface T { x: number }`,
			options:          []any{"type"},
			wantMergedSource: `interface T `,
			wantMergedText:   `type T = `,
			wantOutput:       `export type T = { x: number }`,
		},
		{
			name:             "type parameters comments and heritage",
			code:             `export declare interface T<U /* before close */> extends A, B<U> { x: U }`,
			options:          []any{"type"},
			wantMergedSource: `interface T<U /* before close */> extends A, B<U> { x: U }`,
			wantMergedText:   `type T<U /* before close */> = { x: U } & A & B<U>`,
			wantOutput:       `export declare type T<U /* before close */> = { x: U } & A & B<U>`,
		},
		{
			name:             "default export spans wrapper",
			code:             `export default interface T extends A, B { x: number }`,
			options:          []any{"type"},
			wantMergedSource: `export default interface T extends A, B { x: number }`,
			wantMergedText:   "type T = { x: number } & A & B\nexport default T",
			wantOutput:       "type T = { x: number } & A & B\nexport default T",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			diagnostics := lintConsistentTypeDefinitions(t, testCase.code, testCase.options, rule.EditDemandAutofix)
			if len(diagnostics) != 1 {
				t.Fatalf("diagnostics = %d, want 1", len(diagnostics))
			}
			if diagnostics[0].FixesPtr == nil {
				t.Fatal("diagnostic has no fixes")
			}

			merged := mergeFixesForTest(testCase.code, *diagnostics[0].FixesPtr)
			if got := testCase.code[merged.Range.Pos():merged.Range.End()]; got != testCase.wantMergedSource {
				t.Errorf("merged fix source = %q, want %q", got, testCase.wantMergedSource)
			}
			if merged.Text != testCase.wantMergedText {
				t.Errorf("merged fix text = %q, want %q", merged.Text, testCase.wantMergedText)
			}

			output, unapplied, fixed := linter.ApplyRuleFixes(testCase.code, diagnostics)
			if !fixed || len(unapplied) != 0 {
				t.Fatalf("ApplyRuleFixes fixed = %v, unapplied = %d", fixed, len(unapplied))
			}
			if output != testCase.wantOutput {
				t.Errorf("fixed output = %q, want %q", output, testCase.wantOutput)
			}
		})
	}
}

func lintConsistentTypeDefinitions(
	t *testing.T,
	code string,
	options []any,
	demand rule.EditDemand,
) []rule.RuleDiagnostic {
	t.Helper()

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(code, "fix-edits.ts", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}

	var diagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program: lintprogram.NewFromCompiler(program),
		File:    sourceFile.FileName(),
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     ConsistentTypeDefinitionsRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return ConsistentTypeDefinitionsRule.Run(ctx, options)
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
	return diagnostics
}

func mergeFixesForTest(source string, fixes []rule.RuleFix) rule.RuleFix {
	fixes = slices.Clone(fixes)
	slices.SortFunc(fixes, func(a rule.RuleFix, b rule.RuleFix) int {
		if byStart := a.Range.Pos() - b.Range.Pos(); byStart != 0 {
			return byStart
		}
		return a.Range.End() - b.Range.End()
	})

	start := fixes[0].Range.Pos()
	end := fixes[len(fixes)-1].Range.End()
	lastEnd := start
	var text strings.Builder
	for _, fix := range fixes {
		text.WriteString(source[lastEnd:fix.Range.Pos()])
		text.WriteString(fix.Text)
		lastEnd = fix.Range.End()
	}
	return rule.RuleFixReplaceRange(fixes[0].Range.WithEnd(end), text.String())
}
