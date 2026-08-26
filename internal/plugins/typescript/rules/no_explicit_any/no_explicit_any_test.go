package no_explicit_any

import (
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoExplicitAnyRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoExplicitAnyRule, []rule_tester.ValidTestCase{
		{Code: `const number: number = 1;`},
		{
			Code:    `function foo(...args: any[]) {}`,
			Options: []interface{}{map[string]interface{}{"ignoreRestArgs": true}},
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `const number: any = 1;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 18,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    `const number: unknown = 1;`,
						},
						{
							MessageId: "suggestNever",
							Output:    `const number: never = 1;`,
						},
					},
				},
			},
		},
		{
			Code: `type T = keyof any;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    16,
					EndLine:   1,
					EndColumn: 19,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestPropertyKey",
							Output:    `type T = PropertyKey;`,
						},
					},
				},
			},
		},
		{
			Code: `function foo(...args: any[]) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    23,
					EndLine:   1,
					EndColumn: 26,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    `function foo(...args: unknown[]) {}`,
						},
						{
							MessageId: "suggestNever",
							Output:    `function foo(...args: never[]) {}`,
						},
					},
				},
			},
		},
		{
			Code:    `const number: any = 1;`,
			Options: []interface{}{map[string]interface{}{"fixToUnknown": true}},
			Output:  []string{`const number: unknown = 1;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 18,
				},
			},
		},
		{
			Code:    `type T = keyof any;`,
			Options: []interface{}{map[string]interface{}{"fixToUnknown": true}},
			Output:  []string{`type T = PropertyKey;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    16,
					EndLine:   1,
					EndColumn: 19,
				},
			},
		},
	})
}

func parseNoExplicitAnySource(t *testing.T, code string) *ast.SourceFile {
	t.Helper()
	return parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/source.ts",
		Path:     "/source.ts",
	}, code, core.ScriptKindTS)
}

func runNoExplicitAnyWithDemand(
	t *testing.T,
	code string,
	options []any,
	demand rule.EditDemand,
) []rule.RuleDiagnostic {
	t.Helper()

	sourceFile := parseNoExplicitAnySource(t, code)
	comments := rule.NewCommentStore(sourceFile)
	diagnostics := make([]rule.RuleDiagnostic, 0, 2)
	ctx := rule.RuleContext{
		SourceFile:     sourceFile,
		Comments:       comments,
		DisableManager: rule.NewDisableManager(sourceFile, comments),
	}.WithDiagnosticConsumer(
		NoExplicitAnyRule.Name,
		rule.SeverityError,
		rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		},
	)

	listener := NoExplicitAnyRule.Run(ctx, options)[ast.KindAnyKeyword]
	if listener == nil {
		t.Fatal("expected any-keyword listener")
	}
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindAnyKeyword {
			listener(node)
		}
		return node.ForEachChild(visit)
	}
	sourceFile.AsNode().ForEachChild(visit)
	return diagnostics
}

func TestNoExplicitAnyEditDemand(t *testing.T) {
	const code = `type Value = any; type Key = keyof any;`

	diagnosticsOnly := runNoExplicitAnyWithDemand(t, code, nil, rule.EditDemandNone)
	suggestions := runNoExplicitAnyWithDemand(t, code, nil, rule.EditDemandSuggestion)
	if len(diagnosticsOnly) != 2 || len(suggestions) != 2 {
		t.Fatalf("diagnostic counts = %d and %d, want 2", len(diagnosticsOnly), len(suggestions))
	}
	for index := range diagnosticsOnly {
		if diagnosticsOnly[index].FixesPtr != nil || diagnosticsOnly[index].Suggestions != nil {
			t.Fatalf("diagnostic %d unexpectedly materialized edits", index)
		}
		if diagnosticsOnly[index].Range != suggestions[index].Range ||
			diagnosticsOnly[index].Message.Id != suggestions[index].Message.Id ||
			diagnosticsOnly[index].Message.Description != suggestions[index].Message.Description {
			t.Fatalf("diagnostic %d changed when suggestions were requested", index)
		}
	}
	if suggestions[0].Suggestions == nil || len(*suggestions[0].Suggestions) != 2 {
		t.Fatalf("plain any suggestions = %#v, want 2", suggestions[0].Suggestions)
	}
	if suggestions[1].Suggestions == nil || len(*suggestions[1].Suggestions) != 1 {
		t.Fatalf("keyof any suggestions = %#v, want 1", suggestions[1].Suggestions)
	}
	if (*suggestions[0].Suggestions)[0].FixesArr[0].Text != "unknown" ||
		(*suggestions[0].Suggestions)[1].FixesArr[0].Text != "never" ||
		(*suggestions[1].Suggestions)[0].FixesArr[0].Text != "PropertyKey" {
		t.Fatalf("unexpected suggestion replacements: %#v, %#v", suggestions[0].Suggestions, suggestions[1].Suggestions)
	}

	const triviaCode = `type Value = /* leading trivia */ any;`
	withTrivia := runNoExplicitAnyWithDemand(t, triviaCode, nil, rule.EditDemandSuggestion)
	if len(withTrivia) != 1 || withTrivia[0].Suggestions == nil || len(*withTrivia[0].Suggestions) != 2 {
		t.Fatalf("trivia diagnostics = %#v, want one diagnostic with two suggestions", withTrivia)
	}
	wantStart := strings.Index(triviaCode, "any")
	wantRange := core.NewTextRange(wantStart, wantStart+len("any"))
	if withTrivia[0].Range != wantRange {
		t.Fatalf("trivia diagnostic range = %#v, want %#v", withTrivia[0].Range, wantRange)
	}
	for index, suggestion := range *withTrivia[0].Suggestions {
		if len(suggestion.FixesArr) != 1 || suggestion.FixesArr[0].Range != wantRange {
			t.Fatalf("trivia suggestion %d fixes = %#v, want range %#v", index, suggestion.FixesArr, wantRange)
		}
	}

	fixOptions := []any{map[string]interface{}{"fixToUnknown": true}}
	withoutFixes := runNoExplicitAnyWithDemand(t, code, fixOptions, rule.EditDemandNone)
	withFixes := runNoExplicitAnyWithDemand(t, code, fixOptions, rule.EditDemandAutofix)
	withSuggestionDemand := runNoExplicitAnyWithDemand(t, code, fixOptions, rule.EditDemandSuggestion)
	if len(withoutFixes) != 2 || len(withFixes) != 2 || len(withSuggestionDemand) != 2 {
		t.Fatalf(
			"fix-option diagnostic counts = %d, %d, and %d, want 2",
			len(withoutFixes),
			len(withFixes),
			len(withSuggestionDemand),
		)
	}
	for index, replacement := range []string{"unknown", "PropertyKey"} {
		if withoutFixes[index].FixesPtr != nil || withoutFixes[index].Suggestions != nil {
			t.Fatalf("diagnostic %d unexpectedly materialized fixes", index)
		}
		if withSuggestionDemand[index].FixesPtr != nil || withSuggestionDemand[index].Suggestions != nil {
			t.Fatalf("diagnostic %d materialized edits for the wrong demand", index)
		}
		if withFixes[index].FixesPtr == nil || len(*withFixes[index].FixesPtr) != 1 ||
			(*withFixes[index].FixesPtr)[0].Text != replacement {
			t.Fatalf("diagnostic %d fixes = %#v, want %q", index, withFixes[index].FixesPtr, replacement)
		}
	}

	suppressed := runNoExplicitAnyWithDemand(
		t,
		"// eslint-disable-next-line @typescript-eslint/no-explicit-any\ntype Value = any;",
		nil,
		rule.EditDemandSuggestion,
	)
	if len(suppressed) != 0 {
		t.Fatalf("suppressed diagnostics = %d, want 0", len(suppressed))
	}
}
