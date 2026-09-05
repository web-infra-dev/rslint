package utils

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func parseAnalysisCacheFixture() *ast.SourceFile {
	return parser.ParseSourceFile(
		ast.SourceFileParseOptions{
			FileName: "/shared-analysis.test.ts",
			Path:     "/shared-analysis.test.ts",
		},
		`test("case", () => {});`,
		core.ScriptKindTS,
	)
}

func TestGetRstestCallAnalysisSharesWithinFileCache(t *testing.T) {
	sourceFile := parseAnalysisCacheFixture()
	cache := rule.NewFileCache()
	first := GetRstestCallAnalysis(rule.RuleContext{
		SourceFile: sourceFile,
		Settings:   map[string]interface{}{"owner": "first"},
	}.WithFileCache(cache))
	second := GetRstestCallAnalysis(rule.RuleContext{
		SourceFile: sourceFile,
		Settings:   map[string]interface{}{"owner": "second"},
	}.WithFileCache(cache))
	if first != second {
		t.Fatal("contexts for one file did not share their analysis")
	}
	if first.ctx.Settings != nil {
		t.Fatal("shared analysis retained rule-specific context fields")
	}
}

func TestGetRstestCallAnalysisSeparatesFileCaches(t *testing.T) {
	sourceFile := parseAnalysisCacheFixture()
	first := GetRstestCallAnalysis(
		rule.RuleContext{SourceFile: sourceFile}.WithFileCache(rule.NewFileCache()),
	)
	second := GetRstestCallAnalysis(
		rule.RuleContext{SourceFile: sourceFile}.WithFileCache(rule.NewFileCache()),
	)
	if first == second {
		t.Fatal("different file caches shared an analysis")
	}
}

func TestGetRstestCallAnalysisWithoutFileCache(t *testing.T) {
	ctx := rule.RuleContext{SourceFile: parseAnalysisCacheFixture()}
	first := GetRstestCallAnalysis(ctx)
	second := GetRstestCallAnalysis(ctx)
	if first == second {
		t.Fatal("context without a file cache unexpectedly promised sharing")
	}
}

func TestGetRstestCallAnalysisDropsReporter(t *testing.T) {
	sourceFile := parseAnalysisCacheFixture()
	ctx := rule.RuleContext{
		SourceFile: sourceFile,
	}.WithFileCache(rule.NewFileCache()).WithReporter(
		"owner-rule",
		rule.SeverityWarning,
		func(rule.RuleDiagnostic) {},
	)
	analysis := GetRstestCallAnalysis(ctx)

	defer func() {
		if got := recover(); got != "rule: uninitialized RuleContext reporter" {
			t.Fatalf("analysis reporter panic = %v", got)
		}
	}()
	analysis.ctx.ReportRange(core.NewTextRange(0, 0), rule.RuleMessage{})
}
