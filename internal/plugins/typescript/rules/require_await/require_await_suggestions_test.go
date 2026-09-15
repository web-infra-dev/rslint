package require_await

import (
	"context"
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
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
				Code: `class Example {
  plain
  /** @deprecated */
  static async [computed]() {
    return 0;
  }
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "removeAsync",
						Output: `class Example {
  plain
  /** @deprecated */
  static [computed]() {
    return 0;
  }
}`,
					}},
				}},
			},
			{
				Code: `class Example {
  plain
  /** @deprecated */
  @dec
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
  /** @deprecated */
  @dec
  [computed]() {
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
			{
				Code: `async /* text containing /** is still an ordinary comment */ function ordinary() {
  return 0;
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "removeAsync",
						Output: `/* text containing /** is still an ordinary comment */ function ordinary() {
  return 0;
}`,
					}},
				}},
			},
			{
				Code: `async /**/ function emptyComment() {
  return 0;
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "removeAsync",
						Output: `/**/ function emptyComment() {
  return 0;
}`,
					}},
				}},
			},
		},
	)
}

func TestRequireAwaitSkipsSuggestionsThatCouldCreateJSDoc(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RequireAwaitRule,
		nil,
		[]rule_tester.InvalidTestCase{
			{
				Code: `class Example {
  field = this.value++
  async /** @deprecated */ [computed]() { return 1; }
}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingAwait"}},
			},
			{
				Code: `async /* ordinary */ /** @deprecated */ function declaration() {
  return 1;
}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingAwait"}},
			},
			{
				Code: `const expression = async /** @deprecated */ function () {
  return 1;
};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingAwait"}},
			},
			{
				Code:   `const arrow = async /** @deprecated */ () => 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingAwait"}},
			},
			{
				Code: `const object = {
  async /** @deprecated */ method() { return 1; },
};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingAwait"}},
			},
			{
				Code: `async /** @deprecated */ function* generator() {
  yield 1;
}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingAwait"}},
			},
		},
	)
}

func TestRequireAwaitClassSuggestionSemantics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		fileName          string
		configName        string
		source            string
		want              string
		wantFixes         int
		wantModifierFlags ast.ModifierFlags
	}{
		{
			name: "postfix-generator",
			source: `class Example {
  field = this.value++
  async *method() { yield 1; }
}`,
			want: `class Example {
  field = this.value++
  ;*method() { yield 1; }
}`,
			wantFixes: 1,
		},
		{
			name: "postfix-in",
			source: `class Example {
  field = this.value--
  async in() { return 1; }
}`,
			want: `class Example {
  field = this.value--
  ;in() { return 1; }
}`,
			wantFixes: 1,
		},
		{
			name: "postfix-instanceof",
			source: `class Example {
  field = this.value++
  async instanceof() { return 1; }
}`,
			want: `class Example {
  field = this.value++
  ;instanceof() { return 1; }
}`,
			wantFixes: 1,
		},
		{
			name: "postfix-computed",
			source: `class Example {
  field = this.value++
  async [computed]() { return 1; }
}`,
			want: `class Example {
  field = this.value++
  [computed]() { return 1; }
}`,
			wantFixes: 1,
		},
		{
			name: "leading-jsdoc-and-return-type",
			source: `class Example {
  field = this.value // trailing
  /** @deprecated */
  async [computed](): Promise<number> { return 1; }
}`,
			want: `class Example {
  field = this.value; // trailing
  /** @deprecated */
  [computed](): number { return 1; }
}`,
			wantFixes: 4,
		},
		{
			name:       "jsdoc-synthesized-modifier",
			fileName:   "jsdoc-synthesized-modifier.js",
			configName: "tsconfig.allowJs.json",
			source: `class Example {
  field = this.value
  /** @public */
  async in() { return 1; }
}`,
			want: `class Example {
  field = this.value;
  /** @public */
  in() { return 1; }
}`,
			wantFixes:         2,
			wantModifierFlags: ast.ModifierFlagsPublic,
		},
		{
			name: "static-prefix",
			source: `class Example {
  field = this.value
  /** @deprecated */
  static async in(): Promise<number> { return 1; }
}`,
			want: `class Example {
  field = this.value
  /** @deprecated */
  static in(): number { return 1; }
}`,
			wantFixes: 3,
		},
		{
			name:      "accessibility-prefix-with-unicode-trivia",
			source:    "class Example {\r\n  field = this.value\r\n  /** @deprecated */\r\n\u00a0\u00a0protected async instanceof() { return 1; }\r\n}",
			want:      "class Example {\r\n  field = this.value\r\n  /** @deprecated */\r\n\u00a0\u00a0protected instanceof() { return 1; }\r\n}",
			wantFixes: 1,
		},
		{
			name: "decorator-prefix",
			source: `class Example {
  field = this.value
  /** @deprecated */
  @dec
  async *method(): AsyncGenerator<number> { yield 1; }
}`,
			want: `class Example {
  field = this.value
  /** @deprecated */
  @dec
  *method(): Generator<number> { yield 1; }
}`,
			wantFixes: 2,
		},
		{
			name: "mid-header-ordinary-comment",
			source: `class Example {
  field = this.value
  async /* keep */ [computed]() { return 1; }
}`,
			want: `class Example {
  field = this.value
  ;/* keep */ [computed]() { return 1; }
}`,
			wantFixes: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			fileName := test.fileName
			if fileName == "" {
				fileName = test.name + ".ts"
			}
			configName := test.configName
			if configName == "" {
				configName = "tsconfig.json"
			}
			program, sourceFile, err := helper.CreateTestProgram(test.source, fileName, configName)
			if err != nil {
				t.Fatal(err)
			}
			if diagnostics := program.GetSyntacticDiagnostics(context.Background(), sourceFile); len(diagnostics) != 0 {
				t.Fatalf("source has syntactic diagnostics: %#v", diagnostics)
			}

			classNode := sourceFile.Statements.Nodes[0]
			members := classNode.Members()
			method := members[len(members)-1]
			wasDeprecated := ast.GetJSDocDeprecatedTag(method) != nil
			modifierFlags := ast.GetCombinedModifierFlags(method) &^ ast.ModifierFlagsAsync
			if test.wantModifierFlags != 0 && modifierFlags&test.wantModifierFlags != test.wantModifierFlags {
				t.Fatalf("original modifier flags = %v, want at least %v", modifierFlags, test.wantModifierFlags)
			}
			isGenerator := ast.GetFunctionFlags(method)&ast.FunctionFlagsGenerator != 0
			suggestions := buildRemoveAsyncSuggestion(sourceFile, method, isGenerator)
			if len(suggestions) != 1 {
				t.Fatalf("suggestions = %#v, want one", suggestions)
			}
			rawFixes := append([]rule.RuleFix(nil), suggestions[0].FixesArr...)
			if len(rawFixes) != test.wantFixes {
				t.Fatalf("fixes = %#v, want %d", rawFixes, test.wantFixes)
			}
			if len(rawFixes) > 1 && rawFixes[0].Range.Pos() == rawFixes[1].Range.Pos() &&
				rawFixes[0].Range.Pos() != rawFixes[0].Range.End() {
				t.Fatalf("same-position LSP edits start with a replacement: %#v", rawFixes[:2])
			}

			got, unapplied, fixed := linter.ApplyRuleFixes(test.source, suggestions)
			if !fixed || len(unapplied) != 0 {
				t.Fatalf("suggestion application: fixed=%v, unapplied=%#v", fixed, unapplied)
			}
			if got != test.want {
				t.Fatalf("suggestion output:\ngot:\n%s\nwant:\n%s", got, test.want)
			}

			fixedProgram, fixedFile, err := helper.CreateTestProgram(got, "fixed-"+fileName, configName)
			if err != nil {
				t.Fatal(err)
			}
			if diagnostics := fixedProgram.GetSyntacticDiagnostics(context.Background(), fixedFile); len(diagnostics) != 0 {
				t.Fatalf("suggestion produced syntactic diagnostics: %#v", diagnostics)
			}
			fixedMembers := fixedFile.Statements.Nodes[0].Members()
			fixedMethod := fixedMembers[len(fixedMembers)-1]
			if gotDeprecated := ast.GetJSDocDeprecatedTag(fixedMethod) != nil; gotDeprecated != wasDeprecated {
				t.Errorf("deprecated metadata changed from %v to %v", wasDeprecated, gotDeprecated)
			}
			if gotFlags := ast.GetCombinedModifierFlags(fixedMethod); gotFlags != modifierFlags {
				t.Errorf("modifier flags = %v, want %v", gotFlags, modifierFlags)
			}
		})
	}
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
