package no_case_declarations

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func unexpectedError(line int, column int, suggestionOutput string) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "unexpected",
		Message:   "Unexpected lexical declaration in case block.",
		Line:      line,
		Column:    column,
		Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
			MessageId: "addBrackets",
			Output:    suggestionOutput,
		}},
	}
}

func TestNoCaseDeclarationsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoCaseDeclarationsRule,
		[]rule_tester.ValidTestCase{
			{Code: `switch (a) { case 1: { let x = 1; break; } }`},
			{Code: `switch (a) { case 1: { const x = 1; break; } }`},
			{Code: `switch (a) { case 1: { function f() {} break; } }`},
			{Code: `switch (a) { case 1: { class C {} break; } }`},
			{Code: `switch (a) { case 1: var x = 1; break; }`},
			{Code: `switch (a) { default: var x = 1; break; }`},
			{Code: `switch (a) { case 1: break; }`},
			{Code: `switch (a) {}`},
			{Code: `switch (a) { case 1: case 2: {} }`},
			{Code: `switch (a) { case 1: if (a) { const x = 1; } break; }`},
			{Code: `switch (a) { case 1: type T = string; interface I { value: T } break; }`},
			{Code: `switch (a) { case 1: enum E { A } namespace N {} break; }`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `switch (a) { case 1: let x = 1; break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedError(1, 22, `switch (a) { case 1: { let x = 1; break; } }`),
				},
			},
			{
				Code: `switch (a) { case 1: const x = 1; break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedError(1, 22, `switch (a) { case 1: { const x = 1; break; } }`),
				},
			},
			{
				Code: `switch (a) { case 1: function f() {} break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedError(1, 22, `switch (a) { case 1: { function f() {} break; } }`),
				},
			},
			{
				Code: `switch (a) { case 1: class C {} break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedError(1, 22, `switch (a) { case 1: { class C {} break; } }`),
				},
			},
			{
				Code: `switch (a) { default: let x = 1; break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedError(1, 23, `switch (a) { default: { let x = 1; break; } }`),
				},
			},
			{
				Code: `switch (a) {
  case 1:
    let x = 1;
    const y = 2;
    break;
  case 2:
  default:
    class C {}
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedError(3, 5, `switch (a) {
  case 1:
    { let x = 1;
    const y = 2;
    break; }
  case 2:
  default:
    class C {}
}`),
					unexpectedError(4, 5, `switch (a) {
  case 1:
    { let x = 1;
    const y = 2;
    break; }
  case 2:
  default:
    class C {}
}`),
					unexpectedError(8, 5, `switch (a) {
  case 1:
    let x = 1;
    const y = 2;
    break;
  case 2:
  default:
    { class C {} }
}`),
				},
			},
			{
				Code: `switch (a) { case 1: using resource = acquire(); break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedError(1, 0, `switch (a) { case 1: { using resource = acquire(); break; } }`),
				},
			},
			{
				Code: `async function f() { switch (a) { case 1: await using resource = acquire(); break; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedError(1, 0, `async function f() { switch (a) { case 1: { await using resource = acquire(); break; } } }`),
				},
			},
			{
				Code: `switch (a) {
  case 1:
    {}
    function f() {}
    break;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedError(4, 5, `switch (a) {
  case 1:
    { {}
    function f() {}
    break; }
}`),
				},
			},
			{
				Code: `switch (a) {
  case 1:
  case 2:
    let x;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedError(4, 5, `switch (a) {
  case 1:
  case 2:
    { let x; }
}`),
				},
			},
		},
	)
}

func lintNoCaseDeclarationsSource(t *testing.T, source string, demand rule.EditDemand) []rule.RuleDiagnostic {
	t.Helper()

	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/no-case-declarations.ts",
		Path:     "/no-case-declarations.ts",
	}, source, core.ScriptKindTS)
	comments := rule.NewCommentStore(sourceFile)
	var diagnostics []rule.RuleDiagnostic
	ctx := rule.RuleContext{
		SourceFile:     sourceFile,
		Comments:       comments,
		DisableManager: rule.NewDisableManager(sourceFile, comments),
	}.WithDiagnosticConsumer(NoCaseDeclarationsRule.Name, rule.SeverityError, rule.DiagnosticConsumer{
		Demand: demand,
		Report: func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
	})
	listeners := NoCaseDeclarationsRule.Run(ctx, nil)
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if listener := listeners[node.Kind]; listener != nil {
			listener(node)
		}
		return node.ForEachChild(visit)
	}
	sourceFile.AsNode().ForEachChild(visit)
	return diagnostics
}

func TestNoCaseDeclarationsEditDemand(t *testing.T) {
	t.Parallel()

	const source = `switch (value) {
  case 0:
    // keep before first statement
    let first = 1;
    const second = 2;
    break; // keep after last statement
}`
	const suggested = `switch (value) {
  case 0:
    // keep before first statement
    { let first = 1;
    const second = 2;
    break; } // keep after last statement
}`

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()

		diagnostics := lintNoCaseDeclarationsSource(t, source, demand)
		if len(diagnostics) != 2 {
			t.Fatalf("demand %d: diagnostics = %d, want 2", demand, len(diagnostics))
		}
		return diagnostics
	}

	diagnostics := map[rule.EditDemand][]rule.RuleDiagnostic{}
	for _, demand := range []rule.EditDemand{
		rule.EditDemandNone,
		rule.EditDemandAutofix,
		rule.EditDemandSuggestion,
		rule.EditDemandAll,
	} {
		diagnostics[demand] = run(demand)
	}

	wantRangeText := []string{"let first = 1;", "const second = 2;"}
	for index, text := range wantRangeText {
		baseline := diagnostics[rule.EditDemandNone][index]
		if got := source[baseline.Range.Pos():baseline.Range.End()]; got != text {
			t.Errorf("diagnostic %d range text = %q, want %q", index, got, text)
		}
		for demand, got := range diagnostics {
			actual := got[index]
			if actual.Range != baseline.Range || actual.Message.Id != baseline.Message.Id ||
				actual.Message.Description != baseline.Message.Description || actual.RuleName != baseline.RuleName {
				t.Errorf("demand %d changed diagnostic %d identity", demand, index)
			}
			if actual.FixesPtr != nil {
				t.Errorf("demand %d diagnostic %d unexpectedly materialized autofixes", demand, index)
			}
		}

		for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix} {
			if diagnostics[demand][index].Suggestions != nil {
				t.Errorf("demand %d diagnostic %d unexpectedly materialized suggestions", demand, index)
			}
		}
		for _, demand := range []rule.EditDemand{rule.EditDemandSuggestion, rule.EditDemandAll} {
			attached := diagnostics[demand][index].Suggestions
			if attached == nil || len(*attached) != 1 {
				t.Fatalf("demand %d diagnostic %d suggestions = %#v, want one", demand, index, attached)
			}
			suggestion := (*attached)[0]
			if suggestion.Message.Id != "addBrackets" || suggestion.Message.Description != "Add {} brackets around the case block." {
				t.Errorf("demand %d diagnostic %d has unexpected suggestion message %#v", demand, index, suggestion.Message)
			}
			output, _, _ := linter.ApplyRuleFixes(source, []rule.RuleSuggestion{suggestion})
			if output != suggested {
				t.Errorf("demand %d diagnostic %d suggestion output:\n%s\nwant:\n%s", demand, index, output, suggested)
			}
		}
	}
}

func TestNoCaseDeclarationsDisableDirectives(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name           string
		source         string
		wantRangeText  string
		wantSuggestion string
	}{
		{
			name: "next-line",
			source: `switch (a) {
  case 1:
    /* eslint-disable-next-line no-case-declarations */
    let suppressed = 1;
    const reported = 2;
    break;
}`,
			wantRangeText: "const reported = 2;",
			wantSuggestion: `switch (a) {
  case 1:
    /* eslint-disable-next-line no-case-declarations */
    { let suppressed = 1;
    const reported = 2;
    break; }
}`,
		},
		{
			name: "block-disable-enable",
			source: `/* eslint-disable no-case-declarations */
switch (a) {
  case 1:
    let suppressed = 1;
}
/* eslint-enable no-case-declarations */
switch (a) {
  default:
    const reported = 2;
    break;
}`,
			wantRangeText: "const reported = 2;",
			wantSuggestion: `/* eslint-disable no-case-declarations */
switch (a) {
  case 1:
    let suppressed = 1;
}
/* eslint-enable no-case-declarations */
switch (a) {
  default:
    { const reported = 2;
    break; }
}`,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			diagnostics := lintNoCaseDeclarationsSource(t, testCase.source, rule.EditDemandAll)
			if len(diagnostics) != 1 {
				t.Fatalf("diagnostics = %d, want 1", len(diagnostics))
			}
			diagnostic := diagnostics[0]
			if got := testCase.source[diagnostic.Range.Pos():diagnostic.Range.End()]; got != testCase.wantRangeText {
				t.Errorf("range text = %q, want %q", got, testCase.wantRangeText)
			}
			if diagnostic.Suggestions == nil || len(*diagnostic.Suggestions) != 1 {
				t.Fatalf("suggestions = %#v, want one", diagnostic.Suggestions)
			}
			output, _, _ := linter.ApplyRuleFixes(testCase.source, []rule.RuleSuggestion{(*diagnostic.Suggestions)[0]})
			if output != testCase.wantSuggestion {
				t.Errorf("suggestion output:\n%s\nwant:\n%s", output, testCase.wantSuggestion)
			}
		})
	}
}
