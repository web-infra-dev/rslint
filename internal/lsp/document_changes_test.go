package lsp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rules"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func partialDocumentChange(
	startLine uint32,
	startCharacter uint32,
	endLine uint32,
	endCharacter uint32,
	text string,
) lsproto.TextDocumentContentChangePartialOrWholeDocument {
	return lsproto.TextDocumentContentChangePartialOrWholeDocument{
		Partial: &lsproto.TextDocumentContentChangePartial{
			Range: lsproto.Range{
				Start: lsproto.Position{Line: startLine, Character: startCharacter},
				End:   lsproto.Position{Line: endLine, Character: endCharacter},
			},
			Text: text,
		},
	}
}

func wholeDocumentChange(text string) lsproto.TextDocumentContentChangePartialOrWholeDocument {
	return lsproto.TextDocumentContentChangePartialOrWholeDocument{
		WholeDocument: &lsproto.TextDocumentContentChangeWholeDocument{Text: text},
	}
}

func TestApplyDocumentChanges(t *testing.T) {
	t.Parallel()

	rangeLength := uint32(999)
	tests := []struct {
		name    string
		content string
		changes []lsproto.TextDocumentContentChangePartialOrWholeDocument
		want    string
		wantErr bool
	}{
		{
			name:    "empty change list",
			content: "unchanged",
			want:    "unchanged",
		},
		{
			name:    "whole document fallback",
			content: "old",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{wholeDocumentChange("new")},
			want:    "new",
		},
		{
			name:    "ASCII insertion",
			content: "abc",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{partialDocumentChange(0, 1, 0, 1, "X")},
			want:    "aXbc",
		},
		{
			name:    "multiline deletion with CRLF",
			content: "a\r\nb",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{partialDocumentChange(0, 1, 1, 0, "")},
			want:    "ab",
		},
		{
			name:    "changes use preceding result",
			content: "abc",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{
				partialDocumentChange(0, 1, 0, 2, "XY"),
				partialDocumentChange(0, 3, 0, 4, "!"),
			},
			want: "aXY!",
		},
		{
			name:    "mixed whole and partial changes",
			content: "ignored",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{
				wholeDocumentChange("abc"),
				partialDocumentChange(0, 3, 0, 3, "d"),
			},
			want: "abcd",
		},
		{
			name:    "UTF-16 astral character",
			content: "a😀b",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{partialDocumentChange(0, 3, 0, 4, "c")},
			want:    "a😀c",
		},
		{
			name:    "UTF-16 astral insertion with sequential CRLF edit",
			content: "a😀b\r\nc",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{
				partialDocumentChange(0, 3, 0, 3, "X\r\nY"),
				partialDocumentChange(1, 1, 1, 2, "B"),
			},
			want: "a😀X\r\nYB\r\nc",
		},
		{
			name:    "UTF-16 CJK character",
			content: "甲乙c",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{partialDocumentChange(0, 2, 0, 3, "d")},
			want:    "甲乙d",
		},
		{
			name:    "UTF-16 offset inside surrogate pair clamps before rune",
			content: "a😀b",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{partialDocumentChange(0, 2, 0, 2, "X")},
			want:    "aX😀b",
		},
		{
			name:    "LSP lines exclude Unicode line separator",
			content: "a\u2028b",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{partialDocumentChange(0, 2, 0, 3, "c")},
			want:    "a\u2028c",
		},
		{
			name:    "out of bounds position clamps to document end",
			content: "abc",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{partialDocumentChange(99, 99, 99, 99, "d")},
			want:    "abcd",
		},
		{
			name:    "deprecated range length is ignored",
			content: "abc",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{{
				Partial: &lsproto.TextDocumentContentChangePartial{
					Range:       lsproto.Range{Start: lsproto.Position{Character: 1}, End: lsproto.Position{Character: 2}},
					RangeLength: &rangeLength,
					Text:        "X",
				},
			}},
			want: "aXc",
		},
		{
			name:    "reversed range is transactional",
			content: "a\nb",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{
				partialDocumentChange(0, 0, 0, 1, "A"),
				partialDocumentChange(1, 0, 0, 0, "invalid"),
			},
			want:    "a\nb",
			wantErr: true,
		},
		{
			name:    "missing change kind",
			content: "abc",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{{}},
			want:    "abc",
			wantErr: true,
		},
		{
			name:    "ambiguous change kind",
			content: "abc",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{{
				Partial:       partialDocumentChange(0, 0, 0, 1, "A").Partial,
				WholeDocument: wholeDocumentChange("whole").WholeDocument,
			}},
			want:    "abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := applyDocumentChanges(tt.content, tt.changes)
			if (err != nil) != tt.wantErr {
				t.Fatalf("applyDocumentChanges() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("applyDocumentChanges() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestComputeLSPLineStarts(t *testing.T) {
	t.Parallel()

	content := "a\r\nb\rc\nd\u2028e"
	want := []int{0, 3, 5, 7}
	got := computeLSPLineStarts(content)
	if len(got) != len(want) {
		t.Fatalf("computeLSPLineStarts() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("computeLSPLineStarts() = %v, want %v", got, want)
		}
	}
}

func FuzzApplyDocumentChanges(f *testing.F) {
	f.Add("a😀b\r\nc", "X", uint32(0), uint32(1), uint32(0), uint32(3))
	f.Add("甲乙\nc", "", uint32(0), uint32(1), uint32(1), uint32(0))
	f.Add("", "content", uint32(99), uint32(99), uint32(99), uint32(99))

	f.Fuzz(func(
		t *testing.T,
		content string,
		replacement string,
		startLine uint32,
		startCharacter uint32,
		endLine uint32,
		endCharacter uint32,
	) {
		if !utf8.ValidString(content) || !utf8.ValidString(replacement) {
			t.Skip()
		}

		updated, err := applyDocumentChanges(content, []lsproto.TextDocumentContentChangePartialOrWholeDocument{
			partialDocumentChange(startLine, startCharacter, endLine, endCharacter, replacement),
		})
		if err != nil {
			if updated != content {
				t.Fatalf("failed change returned partial content %q, want original %q", updated, content)
			}
			return
		}
		if !utf8.ValidString(updated) {
			t.Fatalf("valid inputs produced invalid UTF-8: %q", updated)
		}
	})
}

var benchmarkDocumentContent string

func BenchmarkApplyDocumentChanges(b *testing.B) {
	for _, tt := range []struct {
		name string
		size int
	}{
		{name: "4KiB", size: 4 << 10},
		{name: "256KiB", size: 256 << 10},
		{name: "1MiB", size: 1 << 20},
	} {
		content := strings.Repeat("a", tt.size)
		change := []lsproto.TextDocumentContentChangePartialOrWholeDocument{
			partialDocumentChange(0, uint32(tt.size/2), 0, uint32(tt.size/2), "x"),
		}
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(tt.size))
			for range b.N {
				updated, err := applyDocumentChanges(content, change)
				if err != nil {
					b.Fatal(err)
				}
				benchmarkDocumentContent = updated
			}
		})
	}
}

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
		Program: lintprogram.NewFromCompiler(program),
		File:    file,
		GetRulesForFile: func(f *ast.SourceFile) []linter.ConfiguredRule {
			rules, _ := resolver.EnabledRulesForFile(f.FileName())
			return rules
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
			resolver := config.NewFileConfigResolver(cfg, dir, rules.All(), false)

			target := lspConfigTarget(file, dir, fs)
			served := lintSingleFile(
				program, sourceFile, target, dir, true, resolver.ResolveTarget(target).EnabledRules, rule.EditDemandAll, context.Background(),
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
	resolver := config.NewFileConfigResolver(cfg, dir, rules.All(), false)

	target := lspConfigTarget(file, dir, fs)
	served := lintSingleFile(
		program, sourceFile, target, dir, true, resolver.ResolveTarget(target).EnabledRules, rule.EditDemandAll, context.Background(),
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
