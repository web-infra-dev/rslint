package require_await

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRequireAwaitSuggestionBoundaries(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RequireAwaitRule,
		nil,
		[]rule_tester.InvalidTestCase{
			{
				Code: `async function commented(): Promise /* before-open */ < /* after-open */ Array<number> /* before-close */ > {
  return [];
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "removeAsync",
						Output: `function commented():  /* after-open */ Array<number> /* before-close */  {
  return [];
}`,
					}},
				}},
			},
			{
				Code: `async function qualified(): globalThis.Promise<number> {
  return 1;
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "removeAsync",
						Output: `function qualified(): globalThis.Promise<number> {
  return 1;
}`,
					}},
				}},
			},
			{
				Code: `async function noArguments(): Promise {
  return 1;
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "removeAsync",
						Output: `function noArguments(): Promise {
  return 1;
}`,
					}},
				}},
			},
			{
				Code: `async function parenthesized(): ((Promise<number>)) {
  return 1;
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "removeAsync",
						Output: `function parenthesized(): ((number)) {
  return 1;
}`,
					}},
				}},
			},
			{
				Code: `async function* qualifiedGenerator(): globalThis.AsyncGenerator<number> {
  yield 1;
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "removeAsync",
						Output: `function* qualifiedGenerator(): globalThis.AsyncGenerator<number> {
  yield 1;
}`,
					}},
				}},
			},
			{
				Code: `class Example {
  typed: Value
  async [computed]() {
    return 0;
  }
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "removeAsync",
						Output: `class Example {
  typed: Value
  ;[computed]() {
    return 0;
  }
}`,
					}},
				}},
			},
			{
				Code: `class Example {
  plain
  async [computed]() {
    return 0;
  }
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "removeAsync",
						Output: `class Example {
  plain
  ;[computed]() {
    return 0;
  }
}`,
					}},
				}},
			},
			{
				Code: "async\u00a0/* keep-comment */ function comments() {\r\n  return 0;\r\n}",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "removeAsync",
						Output:    "/* keep-comment */ function comments() {\r\n  return 0;\r\n}",
					}},
				}},
			},
		},
	)
}

func TestRequireAwaitSuggestionDemand(t *testing.T) {
	t.Parallel()

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	compilerProgram, sourceFile, err := helper.CreateTestProgram(
		"async function value(): Promise<number> { return 1; }",
		"require-await-suggestion-demand.ts",
		"tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}
	program := lintprogram.NewFromCompiler(compilerProgram)
	configuredRules := []rule.ConfiguredRule{{
		Name:             RequireAwaitRule.Name,
		Severity:         rule.SeverityError,
		RequiresTypeInfo: true,
		Run: func(ctx rule.RuleContext) rule.RuleListeners {
			return RequireAwaitRule.Run(ctx, nil)
		},
	}}
	getRules := func(*ast.SourceFile) []rule.ConfiguredRule { return configuredRules }

	diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
	for _, demand := range []rule.EditDemand{
		rule.EditDemandNone,
		rule.EditDemandAutofix,
		rule.EditDemandSuggestion,
		rule.EditDemandAll,
	} {
		var got []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:         program,
			File:            sourceFile.FileName(),
			HasTypeInfo:     true,
			GetRulesForFile: getRules,
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					got = append(got, diagnostic)
				},
			},
		})
		if len(got) != 1 {
			t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(got))
		}
		diagnostics[demand] = got[0]
	}

	diagnosticsOnly := diagnostics[rule.EditDemandNone]
	for demand, diagnostic := range diagnostics {
		want, got := diagnosticsOnly, diagnostic
		want.FixesPtr, want.Suggestions = nil, nil
		got.FixesPtr, got.Suggestions = nil, nil
		if !reflect.DeepEqual(got, want) {
			t.Errorf("demand %d changed diagnostic metadata:\ngot:  %#v\nwant: %#v", demand, got, want)
		}
		if len(diagnostic.Fixes()) != 0 {
			t.Errorf("demand %d unexpectedly materialized autofixes", demand)
		}
	}

	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix} {
		if diagnostics[demand].Suggestions != nil {
			t.Errorf("demand %d unexpectedly materialized suggestions", demand)
		}
	}
	for _, demand := range []rule.EditDemand{rule.EditDemandSuggestion, rule.EditDemandAll} {
		suggestions := diagnostics[demand].Suggestions
		if suggestions == nil || len(*suggestions) != 1 {
			t.Fatalf("demand %d: suggestions = %#v, want one", demand, suggestions)
		}
		if (*suggestions)[0].Message.Id != "removeAsync" {
			t.Errorf("demand %d: suggestion message = %q, want removeAsync", demand, (*suggestions)[0].Message.Id)
		}
	}
}
