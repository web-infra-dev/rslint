package no_non_null_assertion

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

func TestNoNonNullAssertionRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNonNullAssertionRule, []rule_tester.ValidTestCase{
		{Code: `x;`},
		{Code: `x.y;`},
		{Code: `x.y.z;`},
		{Code: `x?.y.z;`},
		{Code: `x?.y?.z;`},
		{Code: `!x;`},
	}, []rule_tester.InvalidTestCase{
		// Simple non-null assertion — no suggestion
		{
			Code: `x!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 3,
				},
			},
		},
		// Non-null before property access — suggest optional chain
		{
			Code: `x!.y;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestOptionalChain",
							Output:    `x?.y;`,
						},
					},
				},
			},
		},
		// Non-null at end — no suggestion
		{
			Code: `x.y!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 5,
				},
			},
		},
		// Prefix ! with non-null before property access
		{
			Code: `!x!.y;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    2,
					EndLine:   1,
					EndColumn: 4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestOptionalChain",
							Output:    `!x?.y;`,
						},
					},
				},
			},
		},
		// Non-null before optional chain
		{
			Code: `x!.y?.z;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestOptionalChain",
							Output:    `x?.y?.z;`,
						},
					},
				},
			},
		},
		// Non-null before element access
		{
			Code: `x![y];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestOptionalChain",
							Output:    `x?.[y];`,
						},
					},
				},
			},
		},
		// Non-null before element access with optional chain
		{
			Code: `x![y]?.z;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestOptionalChain",
							Output:    `x?.[y]?.z;`,
						},
					},
				},
			},
		},
		// Non-null before call
		{
			Code: `x.y.z!();`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestOptionalChain",
							Output:    `x.y.z?.();`,
						},
					},
				},
			},
		},
		// Non-null before optional call
		{
			Code: `x.y?.z!();`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestOptionalChain",
							Output:    `x.y?.z?.();`,
						},
					},
				},
			},
		},
		// Triple non-null — no suggestions
		{
			Code: `x!!!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 5,
				},
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 4,
				},
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 3,
				},
			},
		},
		// Double non-null before property access
		{
			Code: `x!!.y;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestOptionalChain",
							Output:    `x!?.y;`,
						},
					},
				},
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 3,
				},
			},
		},
		// Double non-null at end — no suggestions
		{
			Code: `x.y!!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 6,
				},
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 5,
				},
			},
		},
		// Double non-null before call
		{
			Code: `x.y.z!!();`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestOptionalChain",
							Output:    `x.y.z!?.();`,
						},
					},
				},
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 7,
				},
			},
		},
		// Already optional element access with non-null
		{
			Code: `x!?.[y].z;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestOptionalChain",
							Output:    `x?.[y].z;`,
						},
					},
				},
			},
		},
		// Already optional property access with non-null
		{
			Code: `x!?.y.z;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestOptionalChain",
							Output:    `x?.y.z;`,
						},
					},
				},
			},
		},
		// Already optional call with non-null
		{
			Code: `x.y.z!?.();`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestOptionalChain",
							Output:    `x.y.z?.();`,
						},
					},
				},
			},
		},
		// Assignment targets are diagnosed without an unsafe optional-chain suggestion.
		{
			Code: `x!.y.z = value;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 3,
				},
			},
		},
		// Trivia between the assertion and property token stays in place.
		{
			Code: "x!\n// comment\n.y;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestOptionalChain",
							Output:    "x\n// comment\n?.y;",
						},
					},
				},
			},
		},
		// Computed access keeps intervening comments after replacing the assertion.
		{
			Code: "x!\n/* comment */ [y];",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestOptionalChain",
							Output:    "x?.\n/* comment */ [y];",
						},
					},
				},
			},
		},
	})
}

func parseNoNonNullAssertionSource(code string) *ast.SourceFile {
	return parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/source.ts",
		Path:     "/source.ts",
	}, code, core.ScriptKindTS)
}

func runNoNonNullAssertionWithDemand(
	t *testing.T,
	code string,
	demand rule.EditDemand,
) []rule.RuleDiagnostic {
	t.Helper()

	sourceFile := parseNoNonNullAssertionSource(code)
	comments := rule.NewCommentStore(sourceFile)
	diagnostics := make([]rule.RuleDiagnostic, 0, 6)
	ctx := rule.RuleContext{
		SourceFile:     sourceFile,
		Comments:       comments,
		DisableManager: rule.NewDisableManager(sourceFile, comments),
	}.WithDiagnosticConsumer(
		NoNonNullAssertionRule.Name,
		rule.SeverityError,
		rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		},
	)

	listener := NoNonNullAssertionRule.Run(ctx, nil)[ast.KindNonNullExpression]
	if listener == nil {
		t.Fatal("expected non-null-expression listener")
	}
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindNonNullExpression {
			listener(node)
		}
		return node.ForEachChild(visit)
	}
	sourceFile.AsNode().ForEachChild(visit)
	return diagnostics
}

func TestNoNonNullAssertionEditDemand(t *testing.T) {
	const code = `plain!;
property!.value;
optional!?.value;
element![key];
callback!();
target!.value.deep = replacement;`

	diagnosticsOnly := runNoNonNullAssertionWithDemand(t, code, rule.EditDemandNone)
	autofixDemand := runNoNonNullAssertionWithDemand(t, code, rule.EditDemandAutofix)
	suggestionDemand := runNoNonNullAssertionWithDemand(t, code, rule.EditDemandSuggestion)
	if len(diagnosticsOnly) != 6 || len(autofixDemand) != 6 || len(suggestionDemand) != 6 {
		t.Fatalf(
			"diagnostic counts = %d, %d, and %d, want 6",
			len(diagnosticsOnly),
			len(autofixDemand),
			len(suggestionDemand),
		)
	}

	for index := range diagnosticsOnly {
		if diagnosticsOnly[index].FixesPtr != nil || diagnosticsOnly[index].Suggestions != nil {
			t.Fatalf("diagnostics-only result %d unexpectedly materialized edits", index)
		}
		if autofixDemand[index].FixesPtr != nil || autofixDemand[index].Suggestions != nil {
			t.Fatalf("autofix-demand result %d unexpectedly materialized edits", index)
		}
		if diagnosticsOnly[index].Range != suggestionDemand[index].Range ||
			diagnosticsOnly[index].Message.Id != suggestionDemand[index].Message.Id ||
			diagnosticsOnly[index].Message.Description != suggestionDemand[index].Message.Description {
			t.Fatalf("diagnostic %d changed when suggestions were requested", index)
		}
	}

	if suggestionDemand[0].Suggestions != nil || suggestionDemand[5].Suggestions != nil {
		t.Fatalf("plain and assignment diagnostics should not suggest: %#v, %#v", suggestionDemand[0], suggestionDemand[5])
	}
	wantFixes := []struct {
		index        int
		sourceTexts  []string
		replacements []string
	}{
		{index: 1, sourceTexts: []string{"!", "."}, replacements: []string{"", "?."}},
		{index: 2, sourceTexts: []string{"!"}, replacements: []string{""}},
		{index: 3, sourceTexts: []string{"!"}, replacements: []string{"?."}},
		{index: 4, sourceTexts: []string{"!"}, replacements: []string{"?."}},
	}
	for _, want := range wantFixes {
		diagnostic := suggestionDemand[want.index]
		if diagnostic.Suggestions == nil || len(*diagnostic.Suggestions) != 1 {
			t.Fatalf("diagnostic %d suggestions = %#v, want one", want.index, diagnostic.Suggestions)
		}
		suggestion := (*diagnostic.Suggestions)[0]
		if suggestion.Message.Id != "suggestOptionalChain" || len(suggestion.FixesArr) != len(want.sourceTexts) {
			t.Fatalf("diagnostic %d suggestion = %#v", want.index, suggestion)
		}
		for fixIndex, fix := range suggestion.FixesArr {
			sourceText := diagnostic.SourceFile.Text()[fix.Range.Pos():fix.Range.End()]
			if sourceText != want.sourceTexts[fixIndex] || fix.Text != want.replacements[fixIndex] {
				t.Fatalf(
					"diagnostic %d fix %d = replace %q with %q, want %q with %q",
					want.index,
					fixIndex,
					sourceText,
					fix.Text,
					want.sourceTexts[fixIndex],
					want.replacements[fixIndex],
				)
			}
		}
	}
}

func TestNoNonNullAssertionDisableDirectives(t *testing.T) {
	const code = `// eslint-disable-next-line @typescript-eslint/no-non-null-assertion
line!.value;
/* eslint-disable @typescript-eslint/no-non-null-assertion */
scoped![key];
/* eslint-enable @typescript-eslint/no-non-null-assertion */
reported!;`

	diagnostics := runNoNonNullAssertionWithDemand(t, code, rule.EditDemandSuggestion)
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %d, want only the enabled assertion", len(diagnostics))
	}
	if got := diagnostics[0].SourceFile.Text()[diagnostics[0].Range.Pos():diagnostics[0].Range.End()]; got != "reported!" {
		t.Fatalf("reported range text = %q, want %q", got, "reported!")
	}
}

func TestNonNullAssertionTokenRangeFallback(t *testing.T) {
	const code = `value!.property;`
	sourceFile := parseNoNonNullAssertionSource(code)
	var assertion *ast.Node
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindNonNullExpression {
			assertion = node
			return true
		}
		return node.ForEachChild(visit)
	}
	sourceFile.AsNode().ForEachChild(visit)
	if assertion == nil {
		t.Fatal("test setup did not find a non-null assertion")
	}

	expression := assertion.AsNonNullExpression().Expression
	wantStart := strings.Index(code, "!")
	want := core.NewTextRange(wantStart, wantStart+1)
	if got := nonNullAssertionTokenRange(sourceFile, assertion, expression); got != want {
		t.Fatalf("fast-path range = %#v, want %#v", got, want)
	}

	original := assertion.Loc
	for _, malformedEnd := range []int{0, len(code)} {
		assertion.Loc = core.NewTextRange(original.Pos(), malformedEnd)
		if got := nonNullAssertionTokenRange(sourceFile, assertion, expression); got != want {
			t.Fatalf("fallback range with end %d = %#v, want %#v", malformedEnd, got, want)
		}
	}
}
