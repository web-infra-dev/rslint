package lsp

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// markedProgram writes a file whose bytes begin with a byte order mark and
// returns a Program over it, plus the normalized path. The mark is written as
// an escape: a literal U+FEFF in a source file is invisible, and Go rejects one
// in its own source outright.
func markedProgram(t *testing.T, source string) (string, string, *compiler.Program, vfs.FS) {
	t.Helper()

	dir := t.TempDir()
	file := tspath.NormalizePath(filepath.Join(dir, "marked.ts"))
	if err := os.WriteFile(file, []byte("\uFEFF"+source), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(dir, fs)
	program, err := utils.CreateProgramFromOptionsLenient(true, &core.CompilerOptions{
		Target: core.ScriptTargetESNext,
		Module: core.ModuleKindESNext,
	}, []string{file}, host)
	if err != nil {
		t.Fatalf("CreateProgramFromOptionsLenient: %v", err)
	}
	return dir, file, program, fs
}

// The mark is real and the rule finds it, so the editor reports it — a user
// who has `unicode-bom` on should see it flagged.
//
// What the editor cannot do is remove it. The fix is a range over [-1, 0], one
// position ahead of the text, because the mark is not in the text; an editor's
// document holds decoded text, with the mark already turned into the
// document's encoding. So the diagnostic arrives without its fix, and no
// quick fix or fixAll edit can be built from it.
func TestUnicodeBomIsReportedWithoutItsFix(t *testing.T) {
	config.RegisterAllRules()

	dir, file, program, fs := markedProgram(t, "let a = 1;\n")
	sourceFile := sourceFileForPath(program, file, fs)
	if sourceFile == nil {
		t.Fatal("fixture is not in the program")
	}

	cfg := config.RslintConfig{{Rules: config.Rules{"unicode-bom": "error"}}}
	resolver := config.NewFileConfigResolver(cfg, dir, false)

	served := lintSingleFile(
		program, sourceFile, file, true, resolver, rule.EditDemandAll, context.Background(),
	).Diagnostics

	if len(served) != 1 || served[0].RuleName != "unicode-bom" {
		t.Fatalf("the editor should report the mark, got %+v", served)
	}
	if fixes := served[0].Fixes(); len(fixes) != 0 {
		t.Errorf("the editor must not carry a fix it cannot apply, got %+v", fixes)
	}
	if action := createCodeActionFromRuleDiagnostic(served[0], "file:///marked.ts"); action != nil {
		t.Errorf("a diagnostic with no fix must yield no quick fix, got %q", action.Title)
	}
}

// Off the editor path the same rule on the same file keeps its fix, so the
// withholding above is the language server's doing and not a broken rule.
func TestUnicodeBomKeepsItsFixOffTheEditorPath(t *testing.T) {
	config.RegisterAllRules()

	dir, file, program, _ := markedProgram(t, "let a = 1;\n")
	cfg := config.RslintConfig{{Rules: config.Rules{"unicode-bom": "error"}}}
	resolver := config.NewFileConfigResolver(cfg, dir, false)

	var direct []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program: program,
		File:    file,
		GetRulesForFile: func(f *ast.SourceFile) []linter.ConfiguredRule {
			return resolver.ActiveRulesForFileHasTypeInfo(f.FileName(), true)
		},
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandAll,
			Report: func(d rule.RuleDiagnostic) { direct = append(direct, d) },
		},
	})

	if len(direct) != 1 || direct[0].RuleName != "unicode-bom" {
		t.Fatalf("expected one unicode-bom diagnostic, got %+v", direct)
	}
	fixes := direct[0].Fixes()
	if len(fixes) != 1 {
		t.Fatalf("expected the removal fix, got %+v", fixes)
	}
	// ESLint's [-1, 0]: one position ahead of the text, which is exactly why
	// an editor cannot express it.
	if fixes[0].Range.Pos() != -1 || fixes[0].Range.End() != 0 || fixes[0].Text != "" {
		t.Errorf("expected a removal over [-1, 0], got %+v", fixes[0])
	}
}

// Only unicode-bom is affected: another rule reporting on the same file in the
// same pass keeps its fix.
func TestOtherRulesKeepTheirFixesInTheEditor(t *testing.T) {
	config.RegisterAllRules()

	// `export` makes this a module and the function gives the `var` a local
	// scope: no-var declines to fix a global in script mode, and the point here
	// is a second rule that really does carry a fix.
	dir, file, program, fs := markedProgram(t, "export function f() {\n  var a = 1;\n  return a;\n}\n")
	sourceFile := sourceFileForPath(program, file, fs)
	if sourceFile == nil {
		t.Fatal("fixture is not in the program")
	}

	cfg := config.RslintConfig{{
		Rules: config.Rules{
			"unicode-bom": "error",
			"no-var":      "error",
		},
	}}
	resolver := config.NewFileConfigResolver(cfg, dir, false)

	served := lintSingleFile(
		program, sourceFile, file, true, resolver, rule.EditDemandAll, context.Background(),
	).Diagnostics

	byRule := make(map[string][]rule.RuleFix, len(served))
	for _, d := range served {
		byRule[d.RuleName] = d.Fixes()
	}
	if _, found := byRule["unicode-bom"]; !found {
		t.Fatalf("expected a unicode-bom diagnostic, got %v", byRule)
	}
	if len(byRule["unicode-bom"]) != 0 {
		t.Errorf("unicode-bom should have lost its fix, got %+v", byRule["unicode-bom"])
	}
	fixes, found := byRule["no-var"]
	if !found {
		t.Fatalf("expected a no-var diagnostic, got %v", byRule)
	}
	if len(fixes) == 0 {
		t.Error("no-var must keep its fix")
	}
}
