package object_shorthand

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestObjectShorthandRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ObjectShorthandRule,
		[]rule_tester.ValidTestCase{
			// default — always
			{Code: `var x = {y() {}}`},
			{Code: `var x = {y}`},
			{Code: `var x = {a: b}`},
			{Code: `var x = {a: 'a'}`},
			{Code: `var x = {'a': 'a'}`},
			{Code: `var x = {'a': b}`},
			{Code: `var x = {y(x) {}}`},
			{Code: `var {x,y,z} = x`},
			{Code: `var {x: {y}} = z`},
			{Code: `var x = {*x() {}}`},
			{Code: `var x = {x: y}`},
			{Code: `var x = {x: y, y: z}`},
			{Code: `var x = {x() {}, y: z, l(){}}`},
			{Code: `var x = {[y]: y}`},
			{Code: `doSomething({x: y})`},
			{Code: `!{ a: function a(){} };`},
			// arrow functions allowed by default
			{Code: `var x = {y: (x)=>x}`},
			{Code: `doSomething({y: (x)=>x})`},
			// getters/setters allowed
			{Code: `var x = {get y() {}}`},
			{Code: `var x = {set y(z) {}}`},
			{Code: `var x = {get y() {}, set y(z) {}}`},

			// options: properties
			{Code: `var x = {[y]: y}`, Options: []any{"properties"}},
			{Code: `var x = {['y']: 'y'}`, Options: []any{"properties"}},
			{Code: `var x = {['y']: y}`, Options: []any{"properties"}},

			// options: methods
			{Code: `var x = {[y]() {}}`, Options: []any{"methods"}},
			{Code: `var x = {[y]: function x() {}}`, Options: []any{"methods"}},
			{Code: `var x = {[y]: y}`, Options: []any{"methods"}},
			{Code: `var x = {y() {}}`, Options: []any{"methods"}},
			{Code: `var x = {x, y() {}, a:b}`, Options: []any{"methods"}},

			// options: properties disables method shorthand enforcement
			{Code: `var x = {y}`, Options: []any{"properties"}},
			{Code: `var x = {y: {b}}`, Options: []any{"properties"}},

			// options: never
			{Code: `var x = {a: n, c: d, f: g}`, Options: []any{"never"}},
			{Code: `var x = {a: function(){}, b: {c: d}}`, Options: []any{"never"}},

			// ignoreConstructors
			{Code: `var x = {ConstructorFunction: function(){}, a: b}`, Options: []any{"always", map[string]any{"ignoreConstructors": true}}},
			{Code: `var x = {_ConstructorFunction: function(){}, a: b}`, Options: []any{"always", map[string]any{"ignoreConstructors": true}}},
			{Code: `var x = {$ConstructorFunction: function(){}, a: b}`, Options: []any{"always", map[string]any{"ignoreConstructors": true}}},
			{Code: `var x = {__ConstructorFunction: function(){}, a: b}`, Options: []any{"always", map[string]any{"ignoreConstructors": true}}},
			{Code: `var x = {_0ConstructorFunction: function(){}, a: b}`, Options: []any{"always", map[string]any{"ignoreConstructors": true}}},
			{Code: `var x = {notConstructorFunction(){}, b: c}`, Options: []any{"always", map[string]any{"ignoreConstructors": true}}},

			// methodsIgnorePattern
			{Code: `var x = { foo: function() {}  }`, Options: []any{"always", map[string]any{"methodsIgnorePattern": "^foo$"}}},
			{Code: `var x = { 'foo': function() {}  }`, Options: []any{"always", map[string]any{"methodsIgnorePattern": "^foo$"}}},
			{Code: `var x = { ['foo']: function() {}  }`, Options: []any{"always", map[string]any{"methodsIgnorePattern": "^foo$"}}},
			{Code: `var x = { 123: function() {}  }`, Options: []any{"always", map[string]any{"methodsIgnorePattern": "^123$"}}},
			{Code: `var x = { afoob: function() {}  }`, Options: []any{"always", map[string]any{"methodsIgnorePattern": "foo"}}},
			{Code: `var x = { afoob: function() {}  }`, Options: []any{"always", map[string]any{"methodsIgnorePattern": "^.foo.$"}}},

			// avoidQuotes
			{Code: `var x = {'a': function(){}}`, Options: []any{"always", map[string]any{"avoidQuotes": true}}},
			{Code: `var x = {['a']: function(){}}`, Options: []any{"methods", map[string]any{"avoidQuotes": true}}},
			{Code: `var x = {'y': y}`, Options: []any{"properties", map[string]any{"avoidQuotes": true}}},

			// consistent
			{Code: `var x = {a: a, b: b}`, Options: []any{"consistent"}},
			{Code: `var x = {a: b, c: d, f: g}`, Options: []any{"consistent"}},
			{Code: `var x = {a, b}`, Options: []any{"consistent"}},
			{Code: `var x = {a, b, get test() { return 1; }}`, Options: []any{"consistent"}},

			// consistent-as-needed
			{Code: `var x = {a, b}`, Options: []any{"consistent-as-needed"}},
			{Code: `var x = {0: 'foo'}`, Options: []any{"consistent-as-needed"}},
			{Code: `var x = {'key': 'baz'}`, Options: []any{"consistent-as-needed"}},
			{Code: `var x = {foo: 'foo'}`, Options: []any{"consistent-as-needed"}},
			{Code: `var x = {[foo]: foo}`, Options: []any{"consistent-as-needed"}},
			{Code: `var x = {foo: function foo() {}}`, Options: []any{"consistent-as-needed"}},

			// avoidExplicitReturnArrows
			{Code: `({ x: () => foo })`, Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": false}}},
			{Code: `({ x: () => foo })`, Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}}},
			{Code: `({ x() { return; } })`, Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}}},
			{Code: `({ x: () => { this; } })`, Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}}},
			{Code: `function foo() { ({ x: () => { arguments; } }) }`, Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `var x = {x: x}`,
				Output: []string{`var x = {x}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyShorthand", Line: 1, Column: 10}},
			},
			{
				Code:   `var x = {'x': x}`,
				Output: []string{`var x = {x}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyShorthand", Line: 1, Column: 10}},
			},
			{
				Code:   `var x = {y: y, x: x}`,
				Output: []string{`var x = {y, x}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedPropertyShorthand", Line: 1, Column: 10},
					{MessageId: "expectedPropertyShorthand", Line: 1, Column: 16},
				},
			},
			{
				Code:   `var x = {y: function() {}}`,
				Output: []string{`var x = {y() {}}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 10}},
			},
			{
				Code:   `var x = {y: function*() {}}`,
				Output: []string{`var x = {*y() {}}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 10}},
			},
			{
				Code:   `doSomething({x: x})`,
				Output: []string{`doSomething({x})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyShorthand", Line: 1, Column: 14}},
			},
			{
				Code:   `doSomething({y: function() {}})`,
				Output: []string{`doSomething({y() {}})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 14}},
			},
			{
				Code:   `doSomething({[y]: function() {}})`,
				Output: []string{`doSomething({[y]() {}})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 14}},
			},
			// `options: ["never"]`
			{
				Code:    `var x = {y() {}}`,
				Output:  []string{`var x = {y: function() {}}`},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodLongform", Line: 1, Column: 10}},
			},
			{
				Code:    `var x = {*y() {}}`,
				Output:  []string{`var x = {y: function*() {}}`},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodLongform", Line: 1, Column: 10}},
			},
			{
				Code:    `var x = {y}`,
				Output:  []string{`var x = {y: y}`},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyLongform", Line: 1, Column: 10}},
			},
			// properties option
			{
				Code:    `var x = {x: x}`,
				Output:  []string{`var x = {x}`},
				Options: []any{"properties"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyShorthand", Line: 1, Column: 10}},
			},
			// methods option
			{
				Code:    `var x = {y: function() {}}`,
				Output:  []string{`var x = {y() {}}`},
				Options: []any{"methods"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 10}},
			},
			// avoidQuotes
			{
				Code:    `var x = {a: function(){}}`,
				Output:  []string{`var x = {a(){}}`},
				Options: []any{"methods", map[string]any{"avoidQuotes": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 10}},
			},
			{
				Code:    `var x = {'a'(){}}`,
				Output:  []string{`var x = {'a': function(){}}`},
				Options: []any{"always", map[string]any{"avoidQuotes": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedLiteralMethodLongform", Line: 1, Column: 10}},
			},
			// consistent
			{
				Code:    `var x = {a: a, b}`,
				Options: []any{"consistent"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedMix", Line: 1, Column: 9}},
			},
			// consistent-as-needed
			{
				Code:    `var x = {a: a, b: b}`,
				Options: []any{"consistent-as-needed"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedAllPropertiesShorthanded", Line: 1, Column: 9}},
			},
			{
				Code:    `var x = {a, z: function z(){}}`,
				Options: []any{"consistent-as-needed"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedMix", Line: 1, Column: 9}},
			},
			// avoidExplicitReturnArrows
			{
				Code:    `({ x: () => { return; } })`,
				Output:  []string{`({ x() { return; } })`},
				Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 4}},
			},
			{
				Code:    `({ x: foo => { return; } })`,
				Output:  []string{`({ x(foo) { return; } })`},
				Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 4}},
			},
		},
	)
}

func TestObjectShorthandEditDemand(t *testing.T) {
	t.Parallel()

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		`const first = {method: function() { return 1; }};
const second = {arrow: () => { return 2; }};`,
		"edit-demand.ts",
		"tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}

	options := []any{"always", map[string]any{"avoidExplicitReturnArrows": true}}
	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()

		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:      program,
			File:         sourceFile.FileName(),
			ExcludePaths: []string{},
			GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
				return []linter.ConfiguredRule{{
					Name:     ObjectShorthandRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return ObjectShorthandRule.Run(ctx, options)
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
		if len(diagnostics) != 2 {
			t.Fatalf("demand %d: diagnostics = %d, want 2", demand, len(diagnostics))
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
		want := withoutEdits(allEdits[index])
		for demand, diagnostics := range map[rule.EditDemand][]rule.RuleDiagnostic{
			rule.EditDemandNone:       diagnosticsOnly,
			rule.EditDemandAutofix:    autofixOnly,
			rule.EditDemandSuggestion: suggestionOnly,
		} {
			if got := withoutEdits(diagnostics[index]); !reflect.DeepEqual(got, want) {
				t.Errorf("demand %d diagnostic %d changed:\ngot  %#v\nwant %#v", demand, index, got, want)
			}
		}
	}

	wantFixes := []bool{true, true}
	for index, wantFix := range wantFixes {
		if got := autofixOnly[index].FixesPtr != nil; got != wantFix {
			t.Errorf("autofix diagnostic %d fix presence = %t, want %t", index, got, wantFix)
		}
		if !reflect.DeepEqual(autofixOnly[index].FixesPtr, allEdits[index].FixesPtr) {
			t.Errorf("autofix and all-edits diagnostic %d produced different fixes", index)
		}
	}
	for _, diagnostics := range [][]rule.RuleDiagnostic{diagnosticsOnly, suggestionOnly} {
		for index, diagnostic := range diagnostics {
			if diagnostic.FixesPtr != nil {
				t.Errorf("non-autofix diagnostic %d materialized fixes", index)
			}
		}
	}
	for _, diagnostics := range [][]rule.RuleDiagnostic{diagnosticsOnly, autofixOnly, suggestionOnly, allEdits} {
		for index, diagnostic := range diagnostics {
			if diagnostic.Suggestions != nil {
				t.Errorf("autofix-only diagnostic %d materialized suggestions", index)
			}
		}
	}
}
