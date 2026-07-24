package rule

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
)

func TestRuleContextReportWithoutReporterPanics(t *testing.T) {
	textRange := core.NewTextRange(0, 0)
	message := RuleMessage{Description: "must not be dropped silently"}
	tests := []struct {
		name   string
		report func(*RuleContext)
	}{
		{name: "range", report: func(ctx *RuleContext) { ctx.ReportRange(textRange, message) }},
		{name: "range fixes", report: func(ctx *RuleContext) { ctx.ReportRangeWithFixes(textRange, message) }},
		{name: "range suggestions", report: func(ctx *RuleContext) { ctx.ReportRangeWithSuggestions(textRange, message) }},
		{name: "range combined", report: func(ctx *RuleContext) { ctx.ReportRangeWithFixesAndSuggestions(textRange, message, nil, nil) }},
		{name: "range deferred fixes", report: func(ctx *RuleContext) {
			ctx.ReportRangeWithDeferredFixes(textRange, message, func() []RuleFix { return nil })
		}},
		{name: "range deferred suggestions", report: func(ctx *RuleContext) {
			ctx.ReportRangeWithDeferredSuggestions(textRange, message, func() []RuleSuggestion { return nil })
		}},
		{name: "range deferred combined", report: func(ctx *RuleContext) {
			ctx.ReportRangeWithDeferredFixesAndSuggestions(
				textRange,
				message,
				func() []RuleFix { return nil },
				func() []RuleSuggestion { return nil },
			)
		}},
		{name: "node", report: func(ctx *RuleContext) { ctx.ReportNode(nil, message) }},
		{name: "node fixes", report: func(ctx *RuleContext) { ctx.ReportNodeWithFixes(nil, message) }},
		{name: "node suggestions", report: func(ctx *RuleContext) { ctx.ReportNodeWithSuggestions(nil, message) }},
		{name: "node combined", report: func(ctx *RuleContext) { ctx.ReportNodeWithFixesAndSuggestions(nil, message, nil, nil) }},
		{name: "node deferred fixes", report: func(ctx *RuleContext) {
			ctx.ReportNodeWithDeferredFixes(nil, message, func() []RuleFix { return nil })
		}},
		{name: "node deferred suggestions", report: func(ctx *RuleContext) {
			ctx.ReportNodeWithDeferredSuggestions(nil, message, func() []RuleSuggestion { return nil })
		}},
		{name: "node deferred combined", report: func(ctx *RuleContext) {
			ctx.ReportNodeWithDeferredFixesAndSuggestions(
				nil,
				message,
				func() []RuleFix { return nil },
				func() []RuleSuggestion { return nil },
			)
		}},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			defer func() {
				got := recover()
				if got != "rule: uninitialized RuleContext reporter" {
					t.Fatalf("panic = %v, want uninitialized reporter failure", got)
				}
			}()

			var ctx RuleContext
			testCase.report(&ctx)
		})
	}
}

func TestEditDemandValidation(t *testing.T) {
	if !EditDemand(0).IsValid() {
		t.Fatal("zero demand should be valid")
	}
	if !EditDemandAll.IsValid() {
		t.Fatal("all supported edit kinds should be valid")
	}
	if EditDemand(1 << 7).IsValid() {
		t.Fatal("unknown edit kind should be invalid")
	}
}

func TestRuleContextRejectsInvalidDiagnosticConsumer(t *testing.T) {
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, "const value = 1;", core.ScriptKindTS)

	tests := []struct {
		name     string
		consumer DiagnosticConsumer
	}{
		{name: "missing report callback"},
		{
			name: "unknown edit kind",
			consumer: DiagnosticConsumer{
				Demand: EditDemand(1 << 7),
				Report: func(RuleDiagnostic) {},
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected invalid diagnostic consumer to panic")
				}
			}()
			RuleContext{SourceFile: sourceFile}.WithDiagnosticConsumer(
				"test",
				SeverityWarning,
				testCase.consumer,
			)
		})
	}
}

func TestDeferredReportsRejectNilBuilder(t *testing.T) {
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, "const value = 1;", core.ScriptKindTS)
	ctx := RuleContext{SourceFile: sourceFile}.WithDiagnosticConsumer(
		"test",
		SeverityWarning,
		DiagnosticConsumer{Report: func(RuleDiagnostic) {}},
	)
	textRange := core.NewTextRange(0, 0)
	message := RuleMessage{Description: "report"}

	tests := []struct {
		name   string
		report func()
	}{
		{name: "range fixes", report: func() { ctx.ReportRangeWithDeferredFixes(textRange, message, nil) }},
		{name: "range suggestions", report: func() { ctx.ReportRangeWithDeferredSuggestions(textRange, message, nil) }},
		{name: "range combined fixes", report: func() {
			ctx.ReportRangeWithDeferredFixesAndSuggestions(textRange, message, nil, func() []RuleSuggestion { return nil })
		}},
		{name: "range combined suggestions", report: func() {
			ctx.ReportRangeWithDeferredFixesAndSuggestions(textRange, message, func() []RuleFix { return nil }, nil)
		}},
		{name: "node fixes", report: func() { ctx.ReportNodeWithDeferredFixes(nil, message, nil) }},
		{name: "node suggestions", report: func() { ctx.ReportNodeWithDeferredSuggestions(nil, message, nil) }},
		{name: "node combined fixes", report: func() {
			ctx.ReportNodeWithDeferredFixesAndSuggestions(nil, message, nil, func() []RuleSuggestion { return nil })
		}},
		{name: "node combined suggestions", report: func() {
			ctx.ReportNodeWithDeferredFixesAndSuggestions(nil, message, func() []RuleFix { return nil }, nil)
		}},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected nil deferred builder to panic")
				}
			}()
			testCase.report()
		})
	}
}

func TestDeferredFixesAndSuggestionsDemand(t *testing.T) {
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, "const value = 1;", core.ScriptKindTS)
	textRange := core.NewTextRange(0, 5)
	message := RuleMessage{Id: "combined", Description: "combined report"}
	fix := RuleFixReplaceRange(textRange, "let")
	suggestion := RuleSuggestion{
		Message:  RuleMessage{Id: "suggestion", Description: "use let"},
		FixesArr: []RuleFix{fix},
	}

	tests := []struct {
		name               string
		demand             EditDemand
		wantFixBuilder     int
		wantSuggestBuilder int
		wantFixes          bool
		wantSuggestions    bool
	}{
		{name: "diagnostics only", demand: EditDemandNone},
		{name: "autofix", demand: EditDemandAutofix, wantFixBuilder: 1, wantFixes: true},
		{name: "suggestions", demand: EditDemandSuggestion, wantSuggestBuilder: 1, wantSuggestions: true},
		{
			name:               "all edits",
			demand:             EditDemandAll,
			wantFixBuilder:     1,
			wantSuggestBuilder: 1,
			wantFixes:          true,
			wantSuggestions:    true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			fixBuilderCalls := 0
			suggestionBuilderCalls := 0
			var got []RuleDiagnostic
			ctx := RuleContext{
				SourceFile:     sourceFile,
				DisableManager: NewDisableManager(sourceFile, NewCommentStore(sourceFile)),
			}.WithDiagnosticConsumer("test", SeverityWarning, DiagnosticConsumer{
				Demand: testCase.demand,
				Report: func(diagnostic RuleDiagnostic) {
					got = append(got, diagnostic)
				},
			})

			ctx.ReportRangeWithDeferredFixesAndSuggestions(
				textRange,
				message,
				func() []RuleFix {
					fixBuilderCalls++
					return []RuleFix{fix}
				},
				func() []RuleSuggestion {
					suggestionBuilderCalls++
					return []RuleSuggestion{suggestion}
				},
			)

			if fixBuilderCalls != testCase.wantFixBuilder {
				t.Errorf("fix builder calls = %d, want %d", fixBuilderCalls, testCase.wantFixBuilder)
			}
			if suggestionBuilderCalls != testCase.wantSuggestBuilder {
				t.Errorf("suggestion builder calls = %d, want %d", suggestionBuilderCalls, testCase.wantSuggestBuilder)
			}
			if len(got) != 1 {
				t.Fatalf("diagnostics = %d, want 1", len(got))
			}
			if got[0].Range != textRange || !reflect.DeepEqual(got[0].Message, message) {
				t.Errorf("diagnostic identity changed: %#v", got[0])
			}
			if (got[0].FixesPtr != nil) != testCase.wantFixes {
				t.Errorf("fix presence = %v, want %v", got[0].FixesPtr != nil, testCase.wantFixes)
			} else if testCase.wantFixes && !reflect.DeepEqual(*got[0].FixesPtr, []RuleFix{fix}) {
				t.Errorf("fixes = %#v, want %#v", *got[0].FixesPtr, []RuleFix{fix})
			}
			if (got[0].Suggestions != nil) != testCase.wantSuggestions {
				t.Errorf("suggestion presence = %v, want %v", got[0].Suggestions != nil, testCase.wantSuggestions)
			} else if testCase.wantSuggestions && !reflect.DeepEqual(*got[0].Suggestions, []RuleSuggestion{suggestion}) {
				t.Errorf("suggestions = %#v, want %#v", *got[0].Suggestions, []RuleSuggestion{suggestion})
			}
		})
	}
}

func TestNodeDeferredFixesAndSuggestionsSkipsSuppressedBuilders(t *testing.T) {
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, "const visible = 1;\n// rslint-disable-next-line test\nconst blocked = 2;\n", core.ScriptKindTS)
	if sourceFile.Statements == nil || len(sourceFile.Statements.Nodes) != 2 {
		t.Fatal("suppression fixture did not parse into two statements")
	}

	fixBuilderCalls := 0
	suggestionBuilderCalls := 0
	var got []RuleDiagnostic
	ctx := RuleContext{
		SourceFile:     sourceFile,
		DisableManager: NewDisableManager(sourceFile, NewCommentStore(sourceFile)),
	}.WithDiagnosticConsumer("test", SeverityWarning, DiagnosticConsumer{
		Demand: EditDemandAll,
		Report: func(diagnostic RuleDiagnostic) {
			got = append(got, diagnostic)
		},
	})
	report := func(node *ast.Node) {
		ctx.ReportNodeWithDeferredFixesAndSuggestions(
			node,
			RuleMessage{Description: "report"},
			func() []RuleFix {
				fixBuilderCalls++
				return nil
			},
			func() []RuleSuggestion {
				suggestionBuilderCalls++
				return nil
			},
		)
	}

	report(sourceFile.Statements.Nodes[0])
	report(sourceFile.Statements.Nodes[1])

	if len(got) != 1 {
		t.Fatalf("diagnostics = %d, want only the unsuppressed diagnostic", len(got))
	}
	visibleNode := sourceFile.Statements.Nodes[0]
	wantRange := core.NewTextRange(visibleNode.Pos(), visibleNode.End())
	if got[0].Range != wantRange {
		t.Fatalf("reported range = %v, want visible node range %v", got[0].Range, wantRange)
	}
	if fixBuilderCalls != 1 || suggestionBuilderCalls != 1 {
		t.Fatalf(
			"builder calls = fixes %d, suggestions %d; want one call each for only the unsuppressed diagnostic",
			fixBuilderCalls,
			suggestionBuilderCalls,
		)
	}
	if got[0].FixesPtr != nil || got[0].Suggestions != nil {
		t.Fatalf("empty builders should attach no artifacts: %#v", got[0])
	}
}

func TestWithReporterPreservesAllEditArtifacts(t *testing.T) {
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, "const value = 1;", core.ScriptKindTS)
	textRange := core.NewTextRange(0, 5)
	fix := RuleFixReplaceRange(textRange, "let")
	builderCalls := 0
	var got []RuleDiagnostic

	ctx := RuleContext{
		SourceFile:     sourceFile,
		DisableManager: NewDisableManager(sourceFile, NewCommentStore(sourceFile)),
	}.WithReporter("test", SeverityWarning, func(d RuleDiagnostic) {
		got = append(got, d)
	})
	ctx.ReportRangeWithDeferredFixes(textRange, RuleMessage{Description: "report"}, func() []RuleFix {
		builderCalls++
		return []RuleFix{fix}
	})

	if builderCalls != 1 {
		t.Fatalf("builder calls = %d, want 1", builderCalls)
	}
	if len(got) != 1 || got[0].FixesPtr == nil || !reflect.DeepEqual(*got[0].FixesPtr, []RuleFix{fix}) {
		t.Fatalf("WithReporter should preserve all edit artifacts, got %#v", got)
	}
}
