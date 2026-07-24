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
		{name: "node fixes", report: func() { ctx.ReportNodeWithDeferredFixes(nil, message, nil) }},
		{name: "node suggestions", report: func() { ctx.ReportNodeWithDeferredSuggestions(nil, message, nil) }},
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
