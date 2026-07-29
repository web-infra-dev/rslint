package no_inferrable_types

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoInferrableTypesRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoInferrableTypesRule, []rule_tester.ValidTestCase{
		// No type annotation - valid
		{Code: `const a = 10;`},
		{Code: `const a = true;`},
		{Code: `const a = 'str';`},
		{Code: `const a = null;`},
		{Code: `const a = undefined;`},
		{Code: `const a = /a/;`},
		{Code: `const a = 10n;`},
		{Code: `const a = Symbol('a');`},

		// Type annotation with different type - valid
		{Code: `const a: unknown = 10;`},
		{Code: `const a: any = true;`},

		// Function parameters with ignoreParameters: true
		{
			Code:    `function fn(a: number = 5) {}`,
			Options: []interface{}{map[string]interface{}{"ignoreParameters": true}},
		},

		// Class properties with ignoreProperties: true
		{
			Code:    `class Foo { prop: number = 5; }`,
			Options: []interface{}{map[string]interface{}{"ignoreProperties": true}},
		},

		// Readonly class properties should be ignored even without option
		{Code: `class Foo { readonly prop: number = 5; }`},

		// Optional properties with initializers are not flagged by this rule.
		{
			Code: `class Foo {
  a?: number = 5;
}`,
		},
	}, []rule_tester.InvalidTestCase{
		// bigint
		{
			Code:   `const a: bigint = 10n;`,
			Output: []string{`const a = 10n;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: bigint = -10n;`,
			Output: []string{`const a = -10n;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: bigint = BigInt(10);`,
			Output: []string{`const a = BigInt(10);`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},

		// boolean
		{
			Code:   `const a: boolean = true;`,
			Output: []string{`const a = true;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: boolean = false;`,
			Output: []string{`const a = false;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: boolean = Boolean(null);`,
			Output: []string{`const a = Boolean(null);`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: boolean = !0;`,
			Output: []string{`const a = !0;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},

		// number
		{
			Code:   `const a: number = 10;`,
			Output: []string{`const a = 10;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: number = +10;`,
			Output: []string{`const a = +10;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: number = -10;`,
			Output: []string{`const a = -10;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: number = Number('1');`,
			Output: []string{`const a = Number('1');`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: number = Infinity;`,
			Output: []string{`const a = Infinity;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: number = NaN;`,
			Output: []string{`const a = NaN;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},

		// null
		{
			Code:   `const a: null = null;`,
			Output: []string{`const a = null;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},

		// RegExp
		{
			Code:   `const a: RegExp = /a/;`,
			Output: []string{`const a = /a/;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: RegExp = RegExp('a');`,
			Output: []string{`const a = RegExp('a');`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: RegExp = new RegExp('a');`,
			Output: []string{`const a = new RegExp('a');`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},

		// string
		{
			Code:   `const a: string = 'str';`,
			Output: []string{`const a = 'str';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   "const a: string = `str`;",
			Output: []string{"const a = `str`;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: string = String(1);`,
			Output: []string{`const a = String(1);`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},

		// symbol
		{
			Code:   `const a: symbol = Symbol('a');`,
			Output: []string{`const a = Symbol('a');`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},

		// undefined
		{
			Code:   `const a: undefined = undefined;`,
			Output: []string{`const a = undefined;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: undefined = void 0;`,
			Output: []string{`const a = void 0;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},

		// Function parameters (without ignoreParameters option)
		{
			Code:   `function fn(a: number = 5) {}`,
			Output: []string{`function fn(a = 5) {}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    13,
				},
			},
		},
		{
			Code:   `const fn = (a: boolean = true) => {};`,
			Output: []string{`const fn = (a = true) => {};`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    13,
				},
			},
		},

		// Class properties (without ignoreProperties option)
		{
			Code:   `class Foo { prop: number = 5; }`,
			Output: []string{`class Foo { prop = 5; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    13,
				},
			},
		},

		// Class properties with definite assignment assertion (!)
		{
			Code: `class A {
  a!: number = 1;
}`,
			Output: []string{`class A {
  a = 1;
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      2,
					Column:    3,
				},
			},
		},

		// Auto-accessor properties
		{
			Code: `class Foo {
  accessor a: number = 5;
}`,
			Output: []string{`class Foo {
  accessor a = 5;
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      2,
					Column:    3,
				},
			},
		},

		// Fixes use AST token boundaries rather than searching through comments
		// that may themselves contain colons.
		{
			Code:   `const value /* name: keep */ : /* type: remove */ number = 1;`,
			Output: []string{`const value /* name: keep */  = 1;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `function fn(value /* name: keep */ ? /* optional: keep */ : /* type: remove */ number = 1) {}`,
			Output: []string{`function fn(value /* name: keep */  /* optional: keep */  = 1) {}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    13,
				},
			},
		},
		{
			Code: `class Foo {
  value /* name: keep */ ! /* definite: keep */ : /* type: remove */ string = 'value';
}`,
			Output: []string{`class Foo {
  value /* name: keep */  /* definite: keep */  = 'value';
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      2,
					Column:    3,
				},
			},
		},
	})
}

func TestNoInferrableTypesEditDemand(t *testing.T) {
	t.Parallel()

	const source = `const value: number = 1;
function fn(optional?: boolean = true) {}
class Example {
  property!: string = 'value';
}`

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		source,
		"no-inferrable-types-edit-demand.ts",
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
					Name:     NoInferrableTypesRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return NoInferrableTypesRule.Run(ctx, nil)
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
		if len(diagnostics) != 3 {
			t.Fatalf("demand %d: diagnostics = %d, want 3", demand, len(diagnostics))
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
	for index := range allEdits {
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
			wantFixes := 1
			if index > 0 {
				wantFixes = 2
			}
			if fixes == nil || len(*fixes) != wantFixes {
				t.Fatalf(
					"demand %d diagnostic %d: fixes = %#v, want %d",
					demand,
					index,
					fixes,
					wantFixes,
				)
			}
		}
	}

	fixedSource, _, fixed := linter.ApplyRuleFixes(source, allEdits)
	const wantFixed = `const value = 1;
function fn(optional = true) {}
class Example {
  property = 'value';
}`
	if !fixed || fixedSource != wantFixed {
		t.Fatalf("fixed source:\n%s\nwant:\n%s", fixedSource, wantFixed)
	}
}
