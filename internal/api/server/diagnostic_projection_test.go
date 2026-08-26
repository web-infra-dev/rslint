package server

import (
	"context"
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/api"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type diagnosticProjectionSource struct{ text string }

func (source diagnosticProjectionSource) Text() string { return source.text }
func (source diagnosticProjectionSource) ECMALineMap() []core.TextPos {
	return core.ComputeECMALineStarts(source.text)
}

func TestProjectLintDiagnosticsPreservesUTF16CoordinatesAndCounts(t *testing.T) {
	const text = "😀x\r\n中y"
	fixes := []rule.RuleFix{{Text: "z", Range: core.NewTextRange(4, 5)}}
	suggestions := []rule.RuleSuggestion{{
		Message:  rule.RuleMessage{Id: "replaceCJK", Description: "Replace CJK", Data: map[string]string{"value": "中"}},
		FixesArr: []rule.RuleFix{{Text: "z", Range: core.NewTextRange(7, 10)}},
	}}
	projection := projectLintDiagnostics([]rule.RuleDiagnostic{{
		RuleName:    "test/rule",
		Message:     rule.RuleMessage{Id: "message", Description: "message"},
		FilePath:    "source.ts",
		SourceFile:  diagnosticProjectionSource{text: text},
		Range:       core.NewTextRange(4, 5),
		Severity:    rule.SeverityWarning,
		FixesPtr:    &fixes,
		Suggestions: &suggestions,
	}})

	if projection.errorCount != 0 || projection.warningCount != 1 ||
		projection.fixableErrorCount != 0 || projection.fixableWarningCount != 1 {
		t.Fatalf("counts = errors:%d warnings:%d fixableErrors:%d fixableWarnings:%d",
			projection.errorCount, projection.warningCount, projection.fixableErrorCount, projection.fixableWarningCount)
	}
	if len(projection.diagnostics) != 1 {
		t.Fatalf("diagnostics = %d, want 1", len(projection.diagnostics))
	}
	diagnostic := projection.diagnostics[0]
	if diagnostic.Range.Start.Line != 1 || diagnostic.Range.Start.Column != 3 ||
		diagnostic.Range.End.Line != 1 || diagnostic.Range.End.Column != 4 {
		t.Fatalf("diagnostic range = %+v, want line 1 columns 3..4", diagnostic.Range)
	}
	if len(diagnostic.Fixes) != 1 || diagnostic.Fixes[0].StartPos != 2 || diagnostic.Fixes[0].EndPos != 3 {
		t.Fatalf("fixes = %+v, want flat UTF-16 range [2,3)", diagnostic.Fixes)
	}
	if len(diagnostic.Suggestions) != 1 || len(diagnostic.Suggestions[0].Fixes) != 1 ||
		diagnostic.Suggestions[0].Fixes[0].StartPos != 5 || diagnostic.Suggestions[0].Fixes[0].EndPos != 6 {
		t.Fatalf("suggestions = %+v, want flat UTF-16 range [5,6)", diagnostic.Suggestions)
	}
}

func TestHandleLintSortsDiagnosticsInResponsePathSpace(t *testing.T) {
	root := t.TempDir()
	workingDirectory := filepath.Join(root, "a-workspace")
	outsideDirectory := filepath.Join(root, "z-outside")
	writeProgramTestFiles(t, root, map[string]string{
		"a-workspace/inside.js": "debugger;\n",
		"z-outside/outside.js":  "debugger;\n",
	})
	insidePath := filepath.Join(workingDirectory, "inside.js")
	outsidePath := filepath.Join(outsideDirectory, "outside.js")

	response, err := (&Handler{}).HandleLintWithContext(context.Background(), api.LintRequest{
		Files:            []string{insidePath, outsidePath},
		WorkingDirectory: workingDirectory,
		ConfigDirectory:  workingDirectory,
		Config:           json.RawMessage(`[{"rules":{"no-debugger":"error"}}]`),
	}, nil)
	if err != nil {
		t.Fatalf("HandleLintWithContext: %v", err)
	}
	got := make([]string, 0, len(response.Diagnostics))
	for _, diagnostic := range response.Diagnostics {
		got = append(got, diagnostic.FilePath)
	}
	want := []string{
		tspath.NormalizePath(filepath.Join("..", "z-outside", "outside.js")),
		"inside.js",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("diagnostic order = %v, want response-path order %v", got, want)
	}
}
