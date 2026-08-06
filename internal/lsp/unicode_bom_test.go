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

// bom is the Unicode byte order mark, U+FEFF, spelled as an escape: a literal
// one in a source file is invisible.
const bom = "\uFEFF"

// programOver writes a file with exactly the given bytes and returns a Program
// over it, plus the normalized path.
func programOver(t *testing.T, content string) (string, string, *compiler.Program, vfs.FS) {
	t.Helper()

	dir := t.TempDir()
	file := tspath.NormalizePath(filepath.Join(dir, "subject.ts"))
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
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

// lintOffTheEditorPath runs the configured rules the way everything that is not
// the language server does, with no filtering in between.
func lintOffTheEditorPath(program *compiler.Program, file string, resolver *config.FileConfigResolver) []rule.RuleDiagnostic {
	var reported []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program: program,
		File:    file,
		GetRulesForFile: func(f *ast.SourceFile) []linter.ConfiguredRule {
			return resolver.ActiveRulesForFileHasTypeInfo(f.FileName(), true)
		},
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandAll,
			Report: func(d rule.RuleDiagnostic) { reported = append(reported, d) },
		},
	})
	return reported
}

// The language server does not run unicode-bom, in either direction: a marked
// file under `never` and an unmarked one under `always` both come back silent.
// An editor's document holds decoded text, so the mark it would report on is
// not in the document, cannot be reached by a text edit, and — for an unsaved
// buffer — is only answerable from a file on disk the buffer may already
// disagree with.
//
// Each case is run off the editor path too, where it does report: the silence
// is the language server's doing, and not a rule that never had anything to say.
func TestUnicodeBomIsNotServedToEditors(t *testing.T) {
	config.RegisterAllRules()

	for _, test := range []struct {
		name    string
		content string
		option  any
		// The fix the rule reports off the editor path. Removal is ESLint's
		// [-1, 0], one position ahead of the text, because the mark is not in
		// the text; `rslint --fix` rewrites the file and can act on it.
		fixPos, fixEnd int
		fixText        string
	}{
		{
			name:    "never on a marked file",
			content: bom + "let a = 1;\n",
			option:  "error",
			fixPos:  -1, fixEnd: 0, fixText: "",
		},
		{
			name:    "always on an unmarked file",
			content: "let a = 1;\n",
			option:  []any{"error", "always"},
			fixPos:  0, fixEnd: 0, fixText: bom,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir, file, program, fs := programOver(t, test.content)
			sourceFile := sourceFileForPath(program, file, fs)
			if sourceFile == nil {
				t.Fatal("fixture is not in the program")
			}

			cfg := config.RslintConfig{{Rules: config.Rules{"unicode-bom": test.option}}}
			resolver := config.NewFileConfigResolver(cfg, dir, false)

			served := lintSingleFile(
				program, sourceFile, file, dir, true, resolver, rule.EditDemandAll, context.Background(),
			).Diagnostics

			if len(served) != 0 {
				t.Errorf("the editor should be served nothing, got %+v", served)
			}

			direct := lintOffTheEditorPath(program, file, resolver)
			if len(direct) != 1 || direct[0].RuleName != "unicode-bom" {
				t.Fatalf("expected one unicode-bom diagnostic off the editor path, got %+v", direct)
			}
			fixes := direct[0].Fixes()
			if len(fixes) != 1 {
				t.Fatalf("expected one fix off the editor path, got %+v", fixes)
			}
			if fixes[0].Range.Pos() != test.fixPos || fixes[0].Range.End() != test.fixEnd || fixes[0].Text != test.fixText {
				t.Errorf("expected a fix over [%d, %d] with %q, got %+v",
					test.fixPos, test.fixEnd, test.fixText, fixes[0])
			}
		})
	}
}

// Only unicode-bom is held back: another rule configured for the same file in
// the same pass reports and keeps its fix.
func TestOtherRulesStillRunInTheEditor(t *testing.T) {
	config.RegisterAllRules()

	// `export` makes this a module and the function gives the `var` a local
	// scope: no-var declines to fix a global in script mode, and the point here
	// is a second rule that really does carry a fix.
	dir, file, program, fs := programOver(t, bom+"export function f() {\n  var a = 1;\n  return a;\n}\n")
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
		program, sourceFile, file, dir, true, resolver, rule.EditDemandAll, context.Background(),
	).Diagnostics

	byRule := make(map[string][]rule.RuleFix, len(served))
	for _, d := range served {
		byRule[d.RuleName] = d.Fixes()
	}
	if fixes, found := byRule["unicode-bom"]; found {
		t.Errorf("unicode-bom should not have run, got %+v", fixes)
	}
	fixes, found := byRule["no-var"]
	if !found {
		t.Fatalf("expected a no-var diagnostic, got %v", byRule)
	}
	if len(fixes) == 0 {
		t.Error("no-var must keep its fix")
	}
}
